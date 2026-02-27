package truedemocracy

import (
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

// setupModuleForGenesis creates a fresh AppModule, Keeper, and sdk.Context for genesis tests.
func setupModuleForGenesis(t *testing.T) (AppModule, Keeper, sdk.Context) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(ModuleName)
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	ms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, nil)
	if err := ms.LoadLatestVersion(); err != nil {
		t.Fatal(err)
	}

	cdc := codec.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(cdc)
	RegisterCodec(cdc)

	nodes := BuildTree()
	keeper := NewKeeper(cdc, storeKey, nodes)
	am := NewAppModule(cdc, keeper)

	ctx := sdk.NewContext(ms, cmtproto.Header{
		Time: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}, false, log.NewNopLogger())

	return am, keeper, ctx
}

func TestExportGenesisNotNil(t *testing.T) {
	am, k, ctx := setupModuleForGenesis(t)

	// Create a domain so there is state to export.
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "TestDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500_000)))

	exported := am.ExportGenesis(ctx, nil)
	if exported == nil {
		t.Fatal("ExportGenesis should not return nil")
	}
}

func TestExportGenesisContainsDomains(t *testing.T) {
	am, k, ctx := setupModuleForGenesis(t)

	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "DomainA", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))
	k.CreateDomain(ctx, "DomainB", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 200_000)))

	exported := am.ExportGenesis(ctx, nil)
	var genesis GenesisState
	if err := json.Unmarshal(exported, &genesis); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(genesis.Domains) != 2 {
		t.Fatalf("expected 2 domains, got %d", len(genesis.Domains))
	}

	names := map[string]bool{}
	for _, d := range genesis.Domains {
		names[d.Name] = true
	}
	if !names["DomainA"] || !names["DomainB"] {
		t.Fatal("exported genesis should contain DomainA and DomainB")
	}
}

func TestExportGenesisContainsValidators(t *testing.T) {
	am, k, ctx := setupModuleForGenesis(t)

	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "ValDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500_000)))
	k.AddMember(ctx, "ValDomain", "validator1", admin)

	pubKey := testPubKey("genesis-val-1")
	stake := sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000))
	if err := k.RegisterValidator(ctx, "validator1", pubKey, stake, "ValDomain"); err != nil {
		t.Fatalf("RegisterValidator failed: %v", err)
	}

	exported := am.ExportGenesis(ctx, nil)
	var genesis GenesisState
	if err := json.Unmarshal(exported, &genesis); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(genesis.Validators) != 1 {
		t.Fatalf("expected 1 validator, got %d", len(genesis.Validators))
	}
	if genesis.Validators[0].OperatorAddr != "validator1" {
		t.Fatalf("expected operator addr validator1, got %s", genesis.Validators[0].OperatorAddr)
	}
	if genesis.Validators[0].Stake != 100_000 {
		t.Fatalf("expected stake 100000, got %d", genesis.Validators[0].Stake)
	}
}

func TestExportGenesisContainsVK(t *testing.T) {
	am, k, ctx := setupModuleForGenesis(t)

	// Initialize the verifying key.
	_, err := k.EnsureVerifyingKey(ctx)
	if err != nil {
		t.Fatalf("EnsureVerifyingKey failed: %v", err)
	}

	exported := am.ExportGenesis(ctx, nil)
	var genesis GenesisState
	if err := json.Unmarshal(exported, &genesis); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if genesis.VerifyingKeyHex == "" {
		t.Fatal("exported genesis should contain non-empty VK hex")
	}

	// Verify the hex is valid.
	vkBytes, err := hex.DecodeString(genesis.VerifyingKeyHex)
	if err != nil {
		t.Fatalf("invalid VK hex: %v", err)
	}
	if len(vkBytes) == 0 {
		t.Fatal("VK bytes should not be empty")
	}
}

func TestExportGenesisNoVKWhenNotInitialized(t *testing.T) {
	am, k, ctx := setupModuleForGenesis(t)

	// Create a domain but do NOT use ZKP.
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "NoZKPDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500_000)))

	exported := am.ExportGenesis(ctx, nil)
	var genesis GenesisState
	if err := json.Unmarshal(exported, &genesis); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if genesis.VerifyingKeyHex != "" {
		t.Fatal("VK hex should be empty when ZKP was not initialized")
	}
}

func TestInitGenesisWithVK(t *testing.T) {
	// Setup first module instance and generate VK.
	_, k1, ctx1 := setupModuleForGenesis(t)
	vkBytes1, err := k1.EnsureVerifyingKey(ctx1)
	if err != nil {
		t.Fatalf("EnsureVerifyingKey failed: %v", err)
	}
	vkHex := hex.EncodeToString(vkBytes1)

	// Create a new module instance and InitGenesis with VK.
	am2, k2, ctx2 := setupModuleForGenesis(t)
	genesisData := GenesisState{
		Domains:         []Domain{},
		Validators:      []GenesisValidator{},
		VerifyingKeyHex: vkHex,
	}
	bz, _ := json.Marshal(genesisData)
	am2.InitGenesis(ctx2, nil, bz)

	// Verify VK was loaded.
	vkBytes2, found := k2.GetVerifyingKey(ctx2)
	if !found {
		t.Fatal("VK should exist after InitGenesis with VK hex")
	}
	if hex.EncodeToString(vkBytes2) != vkHex {
		t.Fatal("loaded VK should match genesis VK")
	}
}

func TestInitGenesisWithoutVK(t *testing.T) {
	am, k, ctx := setupModuleForGenesis(t)
	genesisData := GenesisState{
		Domains:    []Domain{},
		Validators: []GenesisValidator{},
	}
	bz, _ := json.Marshal(genesisData)
	am.InitGenesis(ctx, nil, bz)

	_, found := k.GetVerifyingKey(ctx)
	if found {
		t.Fatal("VK should not exist after InitGenesis without VK hex")
	}
}

func TestGenesisRoundTrip(t *testing.T) {
	// Create state in first module.
	am1, k1, ctx1 := setupModuleForGenesis(t)

	admin := sdk.AccAddress("admin1")
	k1.CreateDomain(ctx1, "RoundTripDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500_000)))
	k1.AddMember(ctx1, "RoundTripDomain", "validator1", admin)

	pubKey := testPubKey("roundtrip-val")
	stake := sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000))
	k1.RegisterValidator(ctx1, "validator1", pubKey, stake, "RoundTripDomain")

	// Initialize VK.
	_, err := k1.EnsureVerifyingKey(ctx1)
	if err != nil {
		t.Fatalf("EnsureVerifyingKey failed: %v", err)
	}

	// Export.
	exported := am1.ExportGenesis(ctx1, nil)

	// Import into fresh module.
	am2, k2, ctx2 := setupModuleForGenesis(t)
	am2.InitGenesis(ctx2, nil, exported)

	// Verify domain.
	domain, found := k2.GetDomain(ctx2, "RoundTripDomain")
	if !found {
		t.Fatal("domain should exist after round-trip")
	}
	if domain.Name != "RoundTripDomain" {
		t.Fatalf("expected RoundTripDomain, got %s", domain.Name)
	}

	// Verify validator.
	v, found := k2.GetValidator(ctx2, "validator1")
	if !found {
		t.Fatal("validator should exist after round-trip")
	}
	if v.Stake.AmountOf("pnyx").Int64() != 100_000 {
		t.Fatalf("expected stake 100000, got %d", v.Stake.AmountOf("pnyx").Int64())
	}

	// Verify VK.
	vkBytes, found := k2.GetVerifyingKey(ctx2)
	if !found {
		t.Fatal("VK should exist after round-trip")
	}
	if len(vkBytes) == 0 {
		t.Fatal("VK bytes should not be empty after round-trip")
	}

	// Verify VK can be deserialized.
	_, err = DeserializeVerifyingKey(vkBytes)
	if err != nil {
		t.Fatalf("DeserializeVerifyingKey failed after round-trip: %v", err)
	}
}
