package main

import (
	"encoding/json"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"truerepublic/token"
	"truerepublic/x/dex"
	"truerepublic/x/truedemocracy"
)

func newGenesisTestApp(t *testing.T) *TrueRepublicApp {
	t.Helper()
	return NewTrueRepublicApp(log.NewNopLogger(), dbm.NewMemDB(), t.TempDir())
}

func defaultGenesisForApp(app *TrueRepublicApp) map[string]json.RawMessage {
	return ModuleBasics.DefaultGenesis(app.appCodec)
}

func setJSONGenesis(t *testing.T, state map[string]json.RawMessage, moduleName string, value any) {
	t.Helper()
	bz, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	state[moduleName] = bz
}

func setBankGenesis(t *testing.T, app *TrueRepublicApp, state map[string]json.RawMessage, balances []banktypes.Balance) {
	t.Helper()
	supply := sdk.NewCoins()
	for _, balance := range balances {
		supply = supply.Add(balance.Coins...)
	}
	genesis := banktypes.NewGenesisState(banktypes.DefaultParams(), balances, supply, nil, nil)
	bz, err := app.appCodec.MarshalJSON(genesis)
	if err != nil {
		t.Fatal(err)
	}
	state[banktypes.ModuleName] = bz
}

func initGenesisApp(app *TrueRepublicApp, state map[string]json.RawMessage) error {
	bz, err := json.Marshal(state)
	if err != nil {
		return err
	}
	_, err = app.InitChain(&abci.RequestInitChain{
		ChainId:       "",
		Time:          time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC),
		AppStateBytes: bz,
	})
	return err
}

func TestDefaultFullAppGenesisIsExactlyBackedAndValid(t *testing.T) {
	app := newGenesisTestApp(t)
	state := defaultGenesisForApp(app)
	if err := initGenesisApp(app, state); err != nil {
		t.Fatalf("default genesis failed: %v", err)
	}
	routes := make(map[string]bool)
	for _, route := range app.crisisKeeper.Routes() {
		routes[route.FullRoute()] = true
	}
	for _, route := range []string{"token/supply-cap", "truedemocracy/escrow-parity", "dex/reserve-custody", "dex/lp-conservation"} {
		if !routes[route] {
			t.Fatalf("runtime invariant %s is not registered: %v", route, routes)
		}
	}
	if app.MsgServiceRouter().Handler(&dex.MsgCreatePool{}) == nil ||
		app.MsgServiceRouter().Handler(&truedemocracy.MsgCreateDomain{}) == nil {
		t.Fatal("custom transaction message routes are not registered")
	}
}

func TestFullAppGenesisRejectsUnbackedCustomClaims(t *testing.T) {
	app := newGenesisTestApp(t)
	state := defaultGenesisForApp(app)
	admin := sdk.AccAddress("genesis-admin")
	setJSONGenesis(t, state, truedemocracy.ModuleName, truedemocracy.GenesisState{
		Domains: []truedemocracy.Domain{{
			Name:          "unbacked",
			Admin:         admin,
			Members:       []string{admin.String()},
			Treasury:      sdk.NewCoins(token.NewCoin(math.NewInt(1_000))),
			Issues:        []truedemocracy.Issue{},
			PermissionReg: []string{},
		}},
		Validators: []truedemocracy.GenesisValidator{},
	})
	if err := initGenesisApp(app, state); err == nil {
		t.Fatal("unbacked treasury genesis was accepted")
	}
}

func TestFullAppGenesisAcceptsExactlyBackedClaims(t *testing.T) {
	app := newGenesisTestApp(t)
	state := defaultGenesisForApp(app)
	provider := sdk.AccAddress("genesis-provider")
	const treasuryAmount int64 = 1_000
	democracyGenesis := truedemocracy.DefaultGenesisState()
	democracyGenesis.Domains[0].Treasury = sdk.NewCoins(token.NewCoin(math.NewInt(treasuryAmount)))
	setJSONGenesis(t, state, truedemocracy.ModuleName, democracyGenesis)
	dexGenesis := dex.DefaultGenesisState()
	dexGenesis.Pools = []dex.Pool{{
		PnyxReserve:     math.NewInt(2_000),
		AssetReserve:    math.NewInt(1_000),
		AssetDenom:      "atom",
		TotalShares:     math.NewInt(100),
		TotalBurned:     math.ZeroInt(),
		TotalVolumePnyx: math.ZeroInt(),
	}}
	dexGenesis.LPPositions = []dex.LPPosition{{AssetDenom: "atom", Provider: provider.String(), Shares: math.NewInt(100)}}
	setJSONGenesis(t, state, dex.ModuleName, dexGenesis)

	setBankGenesis(t, app, state, []banktypes.Balance{
		{Address: authtypes.NewModuleAddress(truedemocracy.ModuleName).String(), Coins: sdk.NewCoins(token.NewCoin(math.NewInt(treasuryAmount + democracyGenesis.Validators[0].Stake)))},
		{Address: authtypes.NewModuleAddress(dex.ModuleName).String(), Coins: sdk.NewCoins(sdk.NewInt64Coin("atom", 1_000), token.NewCoin(math.NewInt(2_000)))},
	})
	if err := initGenesisApp(app, state); err != nil {
		t.Fatalf("exactly backed genesis failed: %v", err)
	}
}

func TestFullAppGenesisExportImportPreservesSupplyAndCustody(t *testing.T) {
	app := newGenesisTestApp(t)
	if err := initGenesisApp(app, defaultGenesisForApp(app)); err != nil {
		t.Fatal(err)
	}
	blockTime := time.Date(2026, 7, 11, 0, 0, 1, 0, time.UTC)
	if _, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1, Time: blockTime}); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Commit(); err != nil {
		t.Fatal(err)
	}
	header := cmtproto.Header{Height: 1, Time: blockTime}
	ctx := app.NewUncachedContext(false, header)
	exported, err := app.mm.ExportGenesis(ctx, app.appCodec)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateLedgerGenesis(app.appCodec, exported); err != nil {
		t.Fatalf("exported ledger is not reconciled: %v", err)
	}
	supplyBefore := app.bankKeeper.GetSupply(ctx, token.BaseDenom).Amount

	restored := newGenesisTestApp(t)
	if err := initGenesisApp(restored, exported); err != nil {
		t.Fatalf("re-import failed: %v", err)
	}
	if _, err := restored.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1, Time: blockTime}); err != nil {
		t.Fatal(err)
	}
	if _, err := restored.Commit(); err != nil {
		t.Fatal(err)
	}
	restoredCtx := restored.NewUncachedContext(false, header)
	supplyAfter := restored.bankKeeper.GetSupply(restoredCtx, token.BaseDenom).Amount
	if !supplyAfter.Equal(supplyBefore) {
		t.Fatalf("canonical supply changed across export/import: before=%s after=%s", supplyBefore, supplyAfter)
	}
	restored.crisisKeeper.AssertInvariants(restoredCtx)
}

func TestRegisteredRuntimeInvariantHaltsOnEscrowDivergence(t *testing.T) {
	app := newGenesisTestApp(t)
	if err := initGenesisApp(app, defaultGenesisForApp(app)); err != nil {
		t.Fatal(err)
	}
	ctx := app.NewContext(false)
	admin := sdk.AccAddress("tamper-admin")
	app.tdKeeper.CreateDomain(ctx, "tampered", admin, sdk.NewCoins(token.NewCoin(math.OneInt())))
	defer func() {
		if recover() == nil {
			t.Fatal("registered crisis invariants did not halt on escrow divergence")
		}
	}()
	app.crisisKeeper.AssertInvariants(ctx)
}
