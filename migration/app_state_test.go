package migration_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"testing"

	"cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"truerepublic/migration"
	"truerepublic/token"
	"truerepublic/x/dex"
)

type applicationFixture struct {
	core    *coreFixture
	cdc     codec.Codec
	raw     json.RawMessage
	unknown json.RawMessage
}

func TestTransformApplicationGenesisReconcilesOwnership(t *testing.T) {
	fixture := newApplicationFixture(t)
	inputBefore := append([]byte(nil), fixture.raw...)

	output, err := migration.TransformApplicationGenesis(fixture.core.desc, fixture.raw, fixture.cdc)
	if err != nil {
		t.Fatalf("TransformApplicationGenesis: %v", err)
	}
	if !bytes.Equal(inputBefore, fixture.raw) {
		t.Fatal("application transform mutated the input")
	}
	root, app := decodeApplicationGenesis(t, output)
	var chainID string
	if err := json.Unmarshal(root["chain_id"], &chainID); err != nil {
		t.Fatal(err)
	}
	if chainID != fixture.core.desc.TargetChainID {
		t.Fatalf("target chain ID = %q", chainID)
	}
	inputRoot, _ := decodeApplicationGenesis(t, inputBefore)
	if !bytes.Equal(root["consensus"], inputRoot["consensus"]) {
		t.Fatal("consensus genesis changed")
	}
	if !bytes.Equal(app["unknown"], fixture.unknown) {
		t.Fatalf("unknown module changed: got=%s want=%s", app["unknown"], fixture.unknown)
	}

	var authState authtypes.GenesisState
	if err := fixture.cdc.UnmarshalJSON(app[authtypes.ModuleName], &authState); err != nil {
		t.Fatal(err)
	}
	accounts := unpackAccounts(t, fixture.cdc, authState.Accounts)
	old0 := fixture.core.old[0]
	new0 := mappingForOld(t, fixture.core.desc, old0)
	if findAccount(accounts, old0) != nil {
		t.Fatal("old auth account remains")
	}
	rewritten := findAccount(accounts, new0.NewOperator)
	if rewritten == nil {
		t.Fatal("rewritten auth account is missing")
	}
	if rewritten.GetAccountNumber() != 7 || rewritten.GetSequence() != 3 {
		t.Fatalf("auth identity changed: number=%d sequence=%d", rewritten.GetAccountNumber(), rewritten.GetSequence())
	}
	if !bytes.Equal(rewritten.GetPubKey().Bytes(), new0.PubKey) {
		t.Fatal("fresh auth public key was not installed")
	}
	if findAccount(accounts, mappingForOld(t, fixture.core.desc, fixture.core.old[1]).NewOperator) != nil {
		t.Fatal("transform invented an auth account for an absent legacy account")
	}

	var bankState banktypes.GenesisState
	if err := fixture.cdc.UnmarshalJSON(app[banktypes.ModuleName], &bankState); err != nil {
		t.Fatal(err)
	}
	if err := bankState.Validate(); err != nil {
		t.Fatalf("bank output invalid: %v", err)
	}
	for _, old := range fixture.core.old {
		if findBalance(bankState.Balances, old) != nil {
			t.Fatalf("old bank balance %s remains", old)
		}
		mapping := mappingForOld(t, fixture.core.desc, old)
		if findBalance(bankState.Balances, mapping.NewOperator) == nil {
			t.Fatalf("new bank balance %s missing", mapping.NewOperator)
		}
	}

	var dexState dex.GenesisState
	if err := json.Unmarshal(app[dex.ModuleName], &dexState); err != nil {
		t.Fatal(err)
	}
	if dexState.LPPositions[0].Provider != mappingForOld(t, fixture.core.desc, fixture.core.old[1]).NewOperator {
		t.Fatal("DEX LP provider was not rewritten")
	}
	if dexState.RegisteredAssets[1].RegisteredBy != new0.NewOperator {
		t.Fatal("DEX asset authority was not rewritten")
	}
}

func TestTransformApplicationGenesisFailsClosed(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*testing.T, *applicationFixture)
	}{
		{
			name: "source chain mismatch",
			mutate: func(t *testing.T, f *applicationFixture) {
				setRootField(t, f, "chain_id", json.RawMessage(`"other-chain"`))
			},
		},
		{
			name: "halt height mismatch",
			mutate: func(t *testing.T, f *applicationFixture) {
				setRootField(t, f, "initial_height", json.RawMessage(`999`))
			},
		},
		{
			name: "embedded source app hash",
			mutate: func(t *testing.T, f *applicationFixture) {
				setRootField(t, f, "app_hash", json.RawMessage(`"AQ=="`))
			},
		},
		{
			name: "missing consensus state",
			mutate: func(t *testing.T, f *applicationFixture) {
				root, _ := decodeApplicationGenesis(t, f.raw)
				delete(root, "consensus")
				var err error
				f.raw, err = json.Marshal(root)
				if err != nil {
					t.Fatal(err)
				}
				signApplicationFixtureExport(t, f)
			},
		},
		{
			name: "consensus key mismatch",
			mutate: func(t *testing.T, f *applicationFixture) {
				mutateConsensus(t, f, func(state *genutiltypes.ConsensusGenesis) {
					key := cmted25519.GenPrivKey().PubKey()
					state.Validators[0].PubKey = key
					state.Validators[0].Address = key.Address()
				})
			},
		},
		{
			name: "consensus power mismatch",
			mutate: func(t *testing.T, f *applicationFixture) {
				mutateConsensus(t, f, func(state *genutiltypes.ConsensusGenesis) {
					state.Validators[0].Power++
				})
			},
		},
		{
			name: "duplicate auth account number",
			mutate: func(t *testing.T, f *applicationFixture) {
				mutateAuth(t, f, func(state *authtypes.GenesisState) {
					account := authtypes.NewBaseAccount(
						sdk.MustAccAddressFromBech32(f.core.unrelated),
						nil,
						7,
						0,
					)
					packed, err := codectypes.NewAnyWithValue(account)
					if err != nil {
						t.Fatal(err)
					}
					state.Accounts = append(state.Accounts, packed)
				})
			},
		},
		{
			name: "mapped module account",
			mutate: func(t *testing.T, f *applicationFixture) {
				mutateAuth(t, f, func(state *authtypes.GenesisState) {
					module := &authtypes.ModuleAccount{
						BaseAccount: &authtypes.BaseAccount{
							Address:       f.core.old[0],
							AccountNumber: 7,
						},
						Name: "legacy-mapped",
					}
					packed, err := codectypes.NewAnyWithValue(module)
					if err != nil {
						t.Fatal(err)
					}
					state.Accounts[0] = packed
				})
			},
		},
		{
			name: "existing target auth account",
			mutate: func(t *testing.T, f *applicationFixture) {
				mapping := mappingForOld(t, f.core.desc, f.core.old[0])
				mutateAuth(t, f, func(state *authtypes.GenesisState) {
					key := &ed25519.PubKey{Key: bytes.Repeat([]byte{0x11}, ed25519.PubKeySize)}
					account := authtypes.NewBaseAccount(
						sdk.MustAccAddressFromBech32(mapping.NewOperator),
						key,
						99,
						0,
					)
					packed, err := codectypes.NewAnyWithValue(account)
					if err != nil {
						t.Fatal(err)
					}
					state.Accounts = append(state.Accounts, packed)
				})
			},
		},
		{
			name: "existing target bank balance",
			mutate: func(t *testing.T, f *applicationFixture) {
				mapping := mappingForOld(t, f.core.desc, f.core.old[0])
				mutateBank(t, f, func(state *banktypes.GenesisState) {
					coins := sdk.NewCoins(sdk.NewInt64Coin("upnyx", 1))
					state.Balances = append(state.Balances, banktypes.Balance{Address: mapping.NewOperator, Coins: coins})
					state.Supply = state.Supply.Add(coins...)
				})
			},
		},
		{
			name: "existing target DEX authority",
			mutate: func(t *testing.T, f *applicationFixture) {
				mapping := mappingForOld(t, f.core.desc, f.core.old[0])
				mutateDEX(t, f, func(state *dex.GenesisState) {
					state.RegisteredAssets[1].RegisteredBy = mapping.NewOperator
				})
			},
		},
		{
			name: "nonempty wasm code",
			mutate: func(t *testing.T, f *applicationFixture) {
				mutateWasm(t, f, func(state *wasmtypes.GenesisState) {
					state.Codes = []wasmtypes.Code{{CodeID: 1}}
				})
			},
		},
		{
			name: "old address in unknown module",
			mutate: func(t *testing.T, f *applicationFixture) {
				setModule(t, f, "unknown", json.RawMessage(`{"authority":"`+f.core.old[0]+`"}`))
			},
		},
		{
			name: "malformed auth module",
			mutate: func(t *testing.T, f *applicationFixture) {
				setModule(t, f, authtypes.ModuleName, json.RawMessage(`{"accounts":"invalid"}`))
			},
		},
		{
			name: "invalid descriptor proof",
			mutate: func(_ *testing.T, f *applicationFixture) {
				f.core.desc.Mappings[0].Signature = []byte("invalid")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newApplicationFixture(t)
			tc.mutate(t, fixture)
			before := append([]byte(nil), fixture.raw...)
			if output, err := migration.TransformApplicationGenesis(fixture.core.desc, fixture.raw, fixture.cdc); err == nil || output != nil {
				t.Fatalf("expected atomic failure, output=%s error=%v", output, err)
			}
			if !bytes.Equal(before, fixture.raw) {
				t.Fatal("failure mutated the input")
			}
		})
	}
}

// TestTransformApplicationGenesisAllowsCanonicalExportWithoutConsensusValidators
// pins the canonical Cosmos SDK export form: a halted running-chain export
// omits CometBFT validators while truedemocracy retains the active application
// validator inventory used to emit InitChain validator updates.
func TestTransformApplicationGenesisAllowsCanonicalExportWithoutConsensusValidators(t *testing.T) {
	fixture := newApplicationFixture(t)
	mutateConsensus(t, fixture, func(state *genutiltypes.ConsensusGenesis) {
		state.Validators = nil
	})
	if _, err := migration.TransformApplicationGenesis(fixture.core.desc, fixture.raw, fixture.cdc); err != nil {
		t.Fatalf("TransformApplicationGenesis rejected canonical running-chain export: %v", err)
	}
}

func newApplicationFixture(t *testing.T) *applicationFixture {
	t.Helper()
	core := newCoreFixture(t, 2)
	cdc := applicationCodec()
	_, app := decodeApplicationGenesis(t, core.raw)
	state := decodeTransformedState(t, core.raw)

	oldAccount := authtypes.NewBaseAccount(
		sdk.MustAccAddressFromBech32(core.old[0]),
		&ed25519.PubKey{Key: state.Validators[0].PubKey},
		7,
		3,
	)
	unrelated := authtypes.NewBaseAccount(
		sdk.MustAccAddressFromBech32(core.unrelated),
		nil,
		8,
		0,
	)
	module := authtypes.NewEmptyModuleAccount(dex.ModuleName)
	packedAccounts, err := authtypes.PackAccounts(authtypes.GenesisAccounts{oldAccount, unrelated, module})
	if err != nil {
		t.Fatal(err)
	}
	authState := authtypes.GenesisState{Params: authtypes.DefaultParams(), Accounts: packedAccounts}
	app[authtypes.ModuleName], err = cdc.MarshalJSON(&authState)
	if err != nil {
		t.Fatal(err)
	}

	moduleAddress := module.GetAddress().String()
	balances := []banktypes.Balance{
		{Address: core.old[0], Coins: sdk.NewCoins(sdk.NewInt64Coin("upnyx", 11))},
		{Address: core.old[1], Coins: sdk.NewCoins(sdk.NewInt64Coin("upnyx", 12))},
		{Address: core.unrelated, Coins: sdk.NewCoins(sdk.NewInt64Coin("upnyx", 13))},
		{Address: moduleAddress, Coins: sdk.NewCoins(sdk.NewInt64Coin("upnyx", 100))},
	}
	supply := sdk.NewCoins()
	for _, balance := range balances {
		supply = supply.Add(balance.Coins...)
	}
	bankState := banktypes.NewGenesisState(banktypes.DefaultParams(), balances, supply, nil, nil)
	app[banktypes.ModuleName], err = cdc.MarshalJSON(bankState)
	if err != nil {
		t.Fatal(err)
	}

	dexState := dex.DefaultGenesisState()
	dexState.RegisteredAssets[1].RegisteredBy = core.old[0]
	dexState.Pools = []dex.Pool{{
		PnyxReserve:     math.NewInt(1_000),
		AssetReserve:    math.NewInt(2_000),
		AssetDenom:      "atom",
		TotalShares:     math.NewInt(100),
		TotalBurned:     math.ZeroInt(),
		TotalVolumePnyx: math.ZeroInt(),
	}}
	dexState.LPPositions = []dex.LPPosition{{
		AssetDenom: "atom",
		Provider:   core.old[1],
		Shares:     math.NewInt(100),
	}}
	app[dex.ModuleName], err = json.Marshal(dexState)
	if err != nil {
		t.Fatal(err)
	}
	wasmState := wasmtypes.GenesisState{Params: wasmtypes.DefaultParams()}
	app[wasmtypes.ModuleName], err = cdc.MarshalJSON(&wasmState)
	if err != nil {
		t.Fatal(err)
	}
	unknown := json.RawMessage(`{"preserve":["exact",1]}`)
	app["unknown"] = unknown

	root, _ := decodeApplicationGenesis(t, core.raw)
	appJSON, err := json.Marshal(app)
	if err != nil {
		t.Fatal(err)
	}
	root["app_state"] = appJSON
	root["initial_height"] = json.RawMessage(`101`)
	root["app_hash"] = json.RawMessage(`null`)
	consensusValidators := make([]cmttypes.GenesisValidator, len(state.Validators))
	for i, validator := range state.Validators {
		key := cmted25519.PubKey(append([]byte(nil), validator.PubKey...))
		consensusValidators[i] = cmttypes.GenesisValidator{
			Address: key.Address(),
			PubKey:  key,
			Power:   validator.Stake / token.StakeMinBaseUnits,
			Name:    "legacy-validator",
		}
	}
	consensusJSON, err := json.Marshal(&genutiltypes.ConsensusGenesis{
		Validators: consensusValidators,
		Params:     cmttypes.DefaultConsensusParams(),
	})
	if err != nil {
		t.Fatal(err)
	}
	root["consensus"] = consensusJSON
	raw, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	fixture := &applicationFixture{core: core, cdc: cdc, raw: raw, unknown: unknown}
	signApplicationFixtureExport(t, fixture)
	return fixture
}

func applicationCodec() codec.Codec {
	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	authtypes.RegisterInterfaces(registry)
	return codec.NewProtoCodec(registry)
}

func decodeApplicationGenesis(t *testing.T, raw json.RawMessage) (map[string]json.RawMessage, map[string]json.RawMessage) {
	t.Helper()
	var root map[string]json.RawMessage
	if err := json.Unmarshal(raw, &root); err != nil {
		t.Fatal(err)
	}
	var app map[string]json.RawMessage
	if err := json.Unmarshal(root["app_state"], &app); err != nil {
		t.Fatal(err)
	}
	return root, app
}

func unpackAccounts(t *testing.T, cdc codec.Codec, anys []*codectypes.Any) []sdk.AccountI {
	t.Helper()
	accounts := make([]sdk.AccountI, len(anys))
	for i, anyAccount := range anys {
		if err := cdc.UnpackAny(anyAccount, &accounts[i]); err != nil {
			t.Fatal(err)
		}
	}
	return accounts
}

func findAccount(accounts []sdk.AccountI, address string) sdk.AccountI {
	for _, account := range accounts {
		if account.GetAddress().String() == address {
			return account
		}
	}
	return nil
}

func findBalance(balances []banktypes.Balance, address string) *banktypes.Balance {
	for i := range balances {
		if balances[i].Address == address {
			return &balances[i]
		}
	}
	return nil
}

func mappingForOld(t *testing.T, desc *migration.Descriptor, old string) migration.OperatorMapping {
	t.Helper()
	for _, mapping := range desc.Mappings {
		if mapping.OldOperator == old {
			return mapping
		}
	}
	t.Fatalf("mapping for %s not found", old)
	return migration.OperatorMapping{}
}

func setRootField(t *testing.T, fixture *applicationFixture, key string, value json.RawMessage) {
	t.Helper()
	root, _ := decodeApplicationGenesis(t, fixture.raw)
	root[key] = value
	var err error
	fixture.raw, err = json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	signApplicationFixtureExport(t, fixture)
}

func setModule(t *testing.T, fixture *applicationFixture, name string, value json.RawMessage) {
	t.Helper()
	root, app := decodeApplicationGenesis(t, fixture.raw)
	app[name] = value
	appJSON, err := json.Marshal(app)
	if err != nil {
		t.Fatal(err)
	}
	root["app_state"] = appJSON
	fixture.raw, err = json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	signApplicationFixtureExport(t, fixture)
}

func signApplicationFixtureExport(t *testing.T, fixture *applicationFixture) {
	t.Helper()
	digest := sha256.Sum256(fixture.raw)
	fixture.core.desc.SourceGenesisSHA256 = append(
		fixture.core.desc.SourceGenesisSHA256[:0],
		digest[:]...,
	)
	signCoreDescriptor(t, fixture.core.desc, fixture.core.fresh)
}

func mutateAuth(t *testing.T, fixture *applicationFixture, mutate func(*authtypes.GenesisState)) {
	t.Helper()
	_, app := decodeApplicationGenesis(t, fixture.raw)
	var state authtypes.GenesisState
	if err := fixture.cdc.UnmarshalJSON(app[authtypes.ModuleName], &state); err != nil {
		t.Fatal(err)
	}
	mutate(&state)
	output, err := fixture.cdc.MarshalJSON(&state)
	if err != nil {
		t.Fatal(err)
	}
	setModule(t, fixture, authtypes.ModuleName, output)
}

func mutateBank(t *testing.T, fixture *applicationFixture, mutate func(*banktypes.GenesisState)) {
	t.Helper()
	_, app := decodeApplicationGenesis(t, fixture.raw)
	var state banktypes.GenesisState
	if err := fixture.cdc.UnmarshalJSON(app[banktypes.ModuleName], &state); err != nil {
		t.Fatal(err)
	}
	mutate(&state)
	output, err := fixture.cdc.MarshalJSON(&state)
	if err != nil {
		t.Fatal(err)
	}
	setModule(t, fixture, banktypes.ModuleName, output)
}

func mutateDEX(t *testing.T, fixture *applicationFixture, mutate func(*dex.GenesisState)) {
	t.Helper()
	_, app := decodeApplicationGenesis(t, fixture.raw)
	var state dex.GenesisState
	if err := json.Unmarshal(app[dex.ModuleName], &state); err != nil {
		t.Fatal(err)
	}
	mutate(&state)
	output, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	setModule(t, fixture, dex.ModuleName, output)
}

func mutateWasm(t *testing.T, fixture *applicationFixture, mutate func(*wasmtypes.GenesisState)) {
	t.Helper()
	_, app := decodeApplicationGenesis(t, fixture.raw)
	var state wasmtypes.GenesisState
	if err := fixture.cdc.UnmarshalJSON(app[wasmtypes.ModuleName], &state); err != nil {
		t.Fatal(err)
	}
	mutate(&state)
	output, err := fixture.cdc.MarshalJSON(&state)
	if err != nil {
		t.Fatal(err)
	}
	setModule(t, fixture, wasmtypes.ModuleName, output)
}

func mutateConsensus(
	t *testing.T,
	fixture *applicationFixture,
	mutate func(*genutiltypes.ConsensusGenesis),
) {
	t.Helper()
	root, _ := decodeApplicationGenesis(t, fixture.raw)
	var state genutiltypes.ConsensusGenesis
	if err := json.Unmarshal(root["consensus"], &state); err != nil {
		t.Fatal(err)
	}
	mutate(&state)
	output, err := json.Marshal(&state)
	if err != nil {
		t.Fatal(err)
	}
	setRootField(t, fixture, "consensus", output)
}
