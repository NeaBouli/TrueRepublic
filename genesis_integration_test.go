package main

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	cryptoproto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

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

func exactlyBackedGenesisForApp(t *testing.T, app *TrueRepublicApp) map[string]json.RawMessage {
	t.Helper()
	state := defaultGenesisForApp(app)
	provider := sdk.AccAddress("genesis-provider")
	const treasuryAmount int64 = 1_000
	admin := sdk.AccAddress("genesis-admin")
	democracyGenesis := truedemocracy.GenesisState{
		Domains: []truedemocracy.Domain{{
			Name:          "Test",
			Admin:         admin,
			Members:       []string{admin.String()},
			Treasury:      sdk.NewCoins(token.NewCoin(math.NewInt(treasuryAmount))),
			Issues:        []truedemocracy.Issue{},
			Options:       truedemocracy.DomainOptions{AdminElectable: true},
			PermissionReg: []string{},
		}},
		Validators: []truedemocracy.GenesisValidator{{
			OperatorAddr: admin.String(),
			PubKey:       ed25519.GenPrivKeyFromSecret([]byte("full-app-genesis-test")).PubKey().Bytes(),
			Stake:        100_000 * truedemocracy.PNYXUnit,
			Domain:       "Test",
		}},
	}
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
	return state
}

func initGenesisApp(app *TrueRepublicApp, state map[string]json.RawMessage) error {
	var democracyGenesis truedemocracy.GenesisState
	if decodeErr := json.Unmarshal(state[truedemocracy.ModuleName], &democracyGenesis); decodeErr != nil {
		return decodeErr
	}
	validators := []abci.ValidatorUpdate{}
	if len(democracyGenesis.Domains) == 0 && len(democracyGenesis.Validators) == 0 {
		pubKey := ed25519.GenPrivKeyFromSecret([]byte("full-app-consensus-test")).PubKey().Bytes()
		democracyGenesis.BootstrapOperatorAddresses = []string{sdk.AccAddress(bytes.Repeat([]byte{0x39}, 20)).String()}
		democracyJSON, marshalErr := json.Marshal(democracyGenesis)
		if marshalErr != nil {
			return marshalErr
		}
		state[truedemocracy.ModuleName] = democracyJSON
		validators = append(validators, abci.ValidatorUpdate{
			PubKey: cryptoproto.PublicKey{Sum: &cryptoproto.PublicKey_Ed25519{Ed25519: pubKey}},
			Power:  1,
		})
	}
	bz, err := json.Marshal(state)
	if err != nil {
		return err
	}
	consensusParams := cmttypes.DefaultConsensusParams().ToProto()
	_, err = app.InitChain(&abci.RequestInitChain{
		ChainId:         "",
		Time:            time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC),
		AppStateBytes:   bz,
		Validators:      validators,
		ConsensusParams: &consensusParams,
	})
	return err
}

func TestEmptyAppGenesisRequiresConsensusValidatorKey(t *testing.T) {
	app := newGenesisTestApp(t)
	bz, err := json.Marshal(defaultGenesisForApp(app))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.InitChain(&abci.RequestInitChain{AppStateBytes: bz}); err == nil {
		t.Fatal("empty app genesis accepted without a real consensus validator key")
	}
}

func TestEnsureConsensusGenesisRejectsMismatchedExistingValidatorSet(t *testing.T) {
	app := newGenesisTestApp(t)
	appState, err := json.Marshal(defaultGenesisForApp(app))
	if err != nil {
		t.Fatal(err)
	}
	genesis := &genutiltypes.AppGenesis{
		ChainID: "validator-binding-test", AppState: appState, Consensus: &genutiltypes.ConsensusGenesis{},
	}
	pubKey := ed25519.GenPrivKeyFromSecret([]byte("validator-binding-key")).PubKey().Bytes()
	operator := sdk.AccAddress(bytes.Repeat([]byte{0x3a}, 20)).String()
	if err := configureGenesisValidatorSet(genesis, []genesisValidatorIdentity{{
		Name: "validator-1", PubKey: pubKey, OperatorAddr: operator,
	}}); err != nil {
		t.Fatal(err)
	}
	var state map[string]json.RawMessage
	if err := json.Unmarshal(genesis.AppState, &state); err != nil {
		t.Fatal(err)
	}
	authGenesis := authtypes.GetGenesisStateFromAppState(app.appCodec, state)
	authGenesis.Accounts = nil
	authJSON, err := app.appCodec.MarshalJSON(&authGenesis)
	if err != nil {
		t.Fatal(err)
	}
	state[authtypes.ModuleName] = authJSON
	matching := []abci.ValidatorUpdate{{
		PubKey: cryptoproto.PublicKey{Sum: &cryptoproto.PublicKey_Ed25519{Ed25519: pubKey}}, Power: 1,
	}}
	if err := ensureConsensusGenesis(app.appCodec, state, matching); err != nil {
		t.Fatalf("matching validator binding rejected: %v", err)
	}
	authGenesis = authtypes.GetGenesisStateFromAppState(app.appCodec, state)
	accounts, err := authtypes.UnpackAccounts(authGenesis.Accounts)
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 || accounts[0].GetAddress().String() != operator {
		t.Fatalf("materialized validator operator auth accounts = %#v, want %s", accounts, operator)
	}
	wrongPower := append([]abci.ValidatorUpdate(nil), matching...)
	wrongPower[0].Power = 2
	if err := ensureConsensusGenesis(app.appCodec, state, wrongPower); err == nil {
		t.Fatal("mismatched consensus power accepted")
	}
	wrongKey := append([]abci.ValidatorUpdate(nil), matching...)
	wrongKey[0].PubKey = cryptoproto.PublicKey{Sum: &cryptoproto.PublicKey_Ed25519{Ed25519: bytes.Repeat([]byte{0x55}, 32)}}
	if err := ensureConsensusGenesis(app.appCodec, state, wrongKey); err == nil {
		t.Fatal("mismatched consensus key accepted")
	}
}

func TestCreateDomainTxHandlerEscrowsToModuleAccount(t *testing.T) {
	app := newGenesisTestApp(t)
	appState, err := json.Marshal(defaultGenesisForApp(app))
	if err != nil {
		t.Fatal(err)
	}
	genesis := &genutiltypes.AppGenesis{
		ChainID:   "create-domain-handler-test",
		AppState:  appState,
		Consensus: &genutiltypes.ConsensusGenesis{},
	}
	validatorPubKey := ed25519.GenPrivKeyFromSecret([]byte("create-domain-handler-validator")).PubKey().Bytes()
	if err := configureGenesisValidatorSet(genesis, []genesisValidatorIdentity{{
		Name:         "validator-1",
		PubKey:       validatorPubKey,
		OperatorAddr: sdk.AccAddress(bytes.Repeat([]byte{0x31}, 20)).String(),
	}}); err != nil {
		t.Fatalf("configure validator set: %v", err)
	}
	var state map[string]json.RawMessage
	if err := json.Unmarshal(genesis.AppState, &state); err != nil {
		t.Fatal(err)
	}
	admin := sdk.AccAddress(bytes.Repeat([]byte{7}, 20))
	addSmokeAccountsToGenesis(t, app, state, []smokeAccount{{
		name:          "admin",
		address:       admin.String(),
		balance:       1_000_000 * token.WholeTokenBaseUnits,
		accountNumber: 1,
	}})
	if err := initGenesisApp(app, state); err != nil {
		t.Fatalf("init genesis: %v", err)
	}
	ctx := app.NewContext(false)
	srv := truedemocracy.NewMsgServer(app.tdKeeper)
	initial := sdk.NewCoins(token.NewCoin(math.NewInt(500_000 * token.WholeTokenBaseUnits)))
	if _, err := srv.CreateDomain(ctx, &truedemocracy.MsgCreateDomain{
		Name:         "Lifecycle",
		Admin:        admin,
		InitialCoins: initial,
	}); err != nil {
		t.Fatalf("create domain: %v", err)
	}
	if _, found := app.tdKeeper.GetDomain(ctx, "Lifecycle"); !found {
		t.Fatal("created domain not found")
	}
	moduleBalance := app.bankKeeper.GetAllBalances(ctx, authtypes.NewModuleAddress(truedemocracy.ModuleName))
	if !moduleBalance.IsAllGTE(initial) {
		t.Fatalf("module escrow balance = %s, want at least %s", moduleBalance, initial)
	}
}

func TestAuthAccountQueryPacksFundedGenesisAccount(t *testing.T) {
	app := newGenesisTestApp(t)
	appState, err := json.Marshal(defaultGenesisForApp(app))
	if err != nil {
		t.Fatal(err)
	}
	genesis := &genutiltypes.AppGenesis{
		ChainID:   "auth-account-query-test",
		AppState:  appState,
		Consensus: &genutiltypes.ConsensusGenesis{},
	}
	validatorPubKey := ed25519.GenPrivKeyFromSecret([]byte("auth-account-query-validator")).PubKey().Bytes()
	if err := configureGenesisValidatorSet(genesis, []genesisValidatorIdentity{{
		Name:         "validator-1",
		PubKey:       validatorPubKey,
		OperatorAddr: sdk.AccAddress(bytes.Repeat([]byte{0x32}, 20)).String(),
	}}); err != nil {
		t.Fatalf("configure validator set: %v", err)
	}
	var state map[string]json.RawMessage
	if err := json.Unmarshal(genesis.AppState, &state); err != nil {
		t.Fatal(err)
	}
	admin := sdk.AccAddress(bytes.Repeat([]byte{9}, 20))
	addSmokeAccountsToGenesis(t, app, state, []smokeAccount{{
		name:          "admin",
		address:       admin.String(),
		balance:       1_000_000 * token.WholeTokenBaseUnits,
		accountNumber: 1,
	}})
	if err := initGenesisApp(app, state); err != nil {
		t.Fatalf("init genesis: %v", err)
	}
	ctx := app.NewContext(false)
	account := app.accountKeeper.GetAccount(ctx, admin)
	if account == nil {
		t.Fatal("funded genesis account missing from account keeper")
	}
	any, err := codectypes.NewAnyWithValue(account)
	if err != nil {
		t.Fatalf("pack account any: %v", err)
	}
	var unpacked sdk.AccountI
	if err := app.appCodec.InterfaceRegistry().UnpackAny(any, &unpacked); err != nil {
		t.Fatalf("unpack account any: %v", err)
	}
	if !unpacked.GetAddress().Equals(admin) {
		t.Fatalf("unpacked address = %s, want %s", unpacked.GetAddress(), admin)
	}
}

func TestCreateDomainMsgSurvivesTxBinaryRoundTrip(t *testing.T) {
	app := newGenesisTestApp(t)
	admin := sdk.AccAddress(bytes.Repeat([]byte{8}, 20))
	initial := sdk.NewCoins(token.NewCoin(math.NewInt(500_000 * token.WholeTokenBaseUnits)))
	builder := app.txConfig.NewTxBuilder()
	if err := builder.SetMsgs(&truedemocracy.MsgCreateDomain{
		Name:         "Lifecycle",
		Admin:        admin,
		InitialCoins: initial,
	}); err != nil {
		t.Fatal(err)
	}
	encoded, err := app.txConfig.TxEncoder()(builder.GetTx())
	if err != nil {
		t.Fatalf("encode tx: %v", err)
	}
	decodedTx, err := app.txConfig.TxDecoder()(encoded)
	if err != nil {
		t.Fatalf("decode tx: %v", err)
	}
	msgs := decodedTx.GetMsgs()
	if len(msgs) != 1 {
		t.Fatalf("decoded msg count = %d, want 1", len(msgs))
	}
	decoded, ok := msgs[0].(*truedemocracy.MsgCreateDomain)
	if !ok {
		t.Fatalf("decoded msg type = %T, want MsgCreateDomain", msgs[0])
	}
	if !decoded.Admin.Equals(admin) || !decoded.InitialCoins.Equal(initial) || decoded.Name != "Lifecycle" {
		t.Fatalf("decoded msg mismatch: %#v", decoded)
	}
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
	ctx := app.NewContext(false)
	if err := app.tdKeeper.ValidateEscrowParity(ctx); err != nil {
		t.Fatalf("consensus bootstrap stake is not bank-backed: %v", err)
	}
	wantStake := math.NewInt(100_000 * truedemocracy.PNYXUnit)
	if supply := app.bankKeeper.GetSupply(ctx, token.BaseDenom).Amount; !supply.Equal(wantStake) {
		t.Fatalf("bootstrap supply = %s, want exact stake %s", supply, wantStake)
	}
	validatorCount := 0
	app.tdKeeper.IterateValidators(ctx, func(validator truedemocracy.Validator) bool {
		validatorCount++
		if !validator.Stake.AmountOf(token.BaseDenom).Equal(wantStake) {
			t.Fatalf("bootstrap validator stake = %s, want %s", validator.Stake, wantStake)
		}
		return false
	})
	if validatorCount != 1 {
		t.Fatalf("bootstrap validator count = %d, want 1", validatorCount)
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

func TestFullAppGenesisRejectsOverCapAndMalformedState(t *testing.T) {
	t.Run("consensus bootstrap exceeds remaining cap", func(t *testing.T) {
		app := newGenesisTestApp(t)
		state := defaultGenesisForApp(app)
		account := sdk.AccAddress("cap-holder")
		setBankGenesis(t, app, state, []banktypes.Balance{{
			Address: account.String(),
			Coins:   sdk.NewCoins(token.NewCoin(token.MaxSupply())),
		}})
		if err := initGenesisApp(app, state); err == nil {
			t.Fatal("consensus bootstrap was allowed to exceed the PNYX cap")
		}
	})

	t.Run("duplicate DEX pool", func(t *testing.T) {
		app := newGenesisTestApp(t)
		state := exactlyBackedGenesisForApp(t, app)
		var dexGenesis dex.GenesisState
		if err := json.Unmarshal(state[dex.ModuleName], &dexGenesis); err != nil {
			t.Fatal(err)
		}
		dexGenesis.Pools = append(dexGenesis.Pools, dexGenesis.Pools[0])
		setJSONGenesis(t, state, dex.ModuleName, dexGenesis)
		if err := initGenesisApp(app, state); err == nil {
			t.Fatal("full app accepted duplicate DEX pool")
		}
	})

	t.Run("negative governance treasury", func(t *testing.T) {
		app := newGenesisTestApp(t)
		state := exactlyBackedGenesisForApp(t, app)
		var democracyGenesis truedemocracy.GenesisState
		if err := json.Unmarshal(state[truedemocracy.ModuleName], &democracyGenesis); err != nil {
			t.Fatal(err)
		}
		democracyGenesis.Domains[0].Treasury = sdk.Coins{{Denom: token.BaseDenom, Amount: math.NewInt(-1)}}
		setJSONGenesis(t, state, truedemocracy.ModuleName, democracyGenesis)
		if err := initGenesisApp(app, state); err == nil {
			t.Fatal("full app accepted negative governance treasury")
		}
	})
}

func TestFullAppGenesisAcceptsExactlyBackedClaims(t *testing.T) {
	app := newGenesisTestApp(t)
	if err := initGenesisApp(app, exactlyBackedGenesisForApp(t, app)); err != nil {
		t.Fatalf("exactly backed genesis failed: %v", err)
	}
}

func TestFullAppGenesisExportImportPreservesNonEmptyCustody(t *testing.T) {
	app := newGenesisTestApp(t)
	if err := initGenesisApp(app, exactlyBackedGenesisForApp(t, app)); err != nil {
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
		t.Fatalf("non-empty exported ledger is not reconciled: %v", err)
	}
	supplyBefore := app.bankKeeper.GetSupply(ctx, token.BaseDenom).Amount

	restored := newGenesisTestApp(t)
	if err := initGenesisApp(restored, exported); err != nil {
		t.Fatalf("non-empty re-import failed: %v", err)
	}
	if _, err := restored.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1, Time: blockTime}); err != nil {
		t.Fatal(err)
	}
	if _, err := restored.Commit(); err != nil {
		t.Fatal(err)
	}
	restoredCtx := restored.NewUncachedContext(false, header)
	if supplyAfter := restored.bankKeeper.GetSupply(restoredCtx, token.BaseDenom).Amount; !supplyAfter.Equal(supplyBefore) {
		t.Fatalf("canonical supply changed across non-empty round trip: before=%s after=%s", supplyBefore, supplyAfter)
	}
	positions := restored.dexKeeper.GetAllLPPositions(restoredCtx)
	if len(positions) != 1 || positions[0].AssetDenom != "atom" || !positions[0].Shares.Equal(math.NewInt(100)) {
		t.Fatalf("LP custody changed across round trip: %+v", positions)
	}
	restored.crisisKeeper.AssertInvariants(restoredCtx)
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

func TestPersistentAppReopensAndContinuesAtNextHeight(t *testing.T) {
	dbDir := t.TempDir()
	db, err := dbm.NewDB("application", dbm.GoLevelDBBackend, dbDir)
	if err != nil {
		t.Fatal(err)
	}
	app := NewTrueRepublicApp(log.NewNopLogger(), db, t.TempDir())
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
	if err := app.Close(); err != nil {
		t.Fatal(err)
	}

	reopenedDB, err := dbm.NewDB("application", dbm.GoLevelDBBackend, dbDir)
	if err != nil {
		t.Fatal(err)
	}
	reopened := NewTrueRepublicApp(log.NewNopLogger(), reopenedDB, t.TempDir())
	t.Cleanup(func() { _ = reopened.Close() })
	if got := reopened.LastBlockHeight(); got != 1 {
		t.Fatalf("reopened height = %d, want 1", got)
	}
	if _, err := reopened.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 2, Time: blockTime.Add(time.Second)}); err != nil {
		t.Fatal(err)
	}
	if _, err := reopened.Commit(); err != nil {
		t.Fatal(err)
	}
	if got := reopened.LastBlockHeight(); got != 2 {
		t.Fatalf("continued height = %d, want 2", got)
	}
}

func requireInvariantPanic(t *testing.T, app *TrueRepublicApp, ctx sdk.Context) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("registered crisis invariants did not halt on divergence")
		}
	}()
	app.crisisKeeper.AssertInvariants(ctx)
}

func TestRegisteredRuntimeInvariantsHaltOnEveryLedgerDivergence(t *testing.T) {
	tests := []struct {
		name       string
		nonEmpty   bool
		corruptApp func(*testing.T, *TrueRepublicApp, sdk.Context)
	}{
		{"supply cap", false, func(t *testing.T, app *TrueRepublicApp, ctx sdk.Context) {
			if err := app.bankKeeper.MintCoins(ctx, truedemocracy.ModuleName, sdk.NewCoins(token.NewCoin(token.MaxSupply()))); err != nil {
				t.Fatal(err)
			}
		}},
		{"escrow parity", false, func(t *testing.T, app *TrueRepublicApp, ctx sdk.Context) {
			admin := sdk.AccAddress("tamper-admin")
			treasury := sdk.NewCoins(token.NewCoin(math.OneInt()))
			app.tdKeeper.CreateDomain(ctx, "tampered", admin, treasury)
			domain, found := app.tdKeeper.GetDomain(ctx, "tampered")
			if !found || !domain.Treasury.Equal(treasury) {
				t.Fatal("failed to create escrow-parity divergence fixture")
			}
		}},
		{"reserve custody", true, func(t *testing.T, app *TrueRepublicApp, ctx sdk.Context) {
			pool, found := app.dexKeeper.GetPool(ctx, "atom")
			if !found {
				t.Fatal("missing atom pool")
			}
			pool.PnyxReserve = pool.PnyxReserve.AddRaw(1)
			app.dexKeeper.SetPool(ctx, pool)
		}},
		{"LP conservation", true, func(t *testing.T, app *TrueRepublicApp, ctx sdk.Context) {
			pool, found := app.dexKeeper.GetPool(ctx, "atom")
			if !found {
				t.Fatal("missing atom pool")
			}
			pool.TotalShares = pool.TotalShares.AddRaw(1)
			app.dexKeeper.SetPool(ctx, pool)
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := newGenesisTestApp(t)
			state := defaultGenesisForApp(app)
			if tc.nonEmpty {
				state = exactlyBackedGenesisForApp(t, app)
			}
			if err := initGenesisApp(app, state); err != nil {
				t.Fatal(err)
			}
			ctx := app.NewContext(false)
			tc.corruptApp(t, app, ctx)
			requireInvariantPanic(t, app, ctx)
		})
	}
}
