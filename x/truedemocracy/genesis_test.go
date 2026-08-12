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
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	rewards "truerepublic/treasury/keeper"
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
	keeper := NewKeeper(cdc, storeKey, nodes, nil, nil)
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
	k.CreateDomain(ctx, "TestDomain", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 500_000*PNYXUnit)))

	exported := am.ExportGenesis(ctx, nil)
	if exported == nil {
		t.Fatal("ExportGenesis should not return nil")
	}
}

func TestExportGenesisContainsDomains(t *testing.T) {
	am, k, ctx := setupModuleForGenesis(t)

	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "DomainA", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 100_000*PNYXUnit)))
	k.CreateDomain(ctx, "DomainB", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 200_000*PNYXUnit)))

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
	k.CreateDomain(ctx, "ValDomain", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 500_000*PNYXUnit)))
	k.AddMember(ctx, "ValDomain", "validator1", admin)

	pubKey := testPubKey("genesis-val-1")
	stake := sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 100_000*PNYXUnit))
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
	if genesis.Validators[0].Stake != 100_000*PNYXUnit {
		t.Fatalf("expected stake 100000, got %d", genesis.Validators[0].Stake)
	}
}

func TestExportGenesisContainsVK(t *testing.T) {
	am, k, ctx := setupModuleForGenesis(t)

	// Install the consensus-configured verifying key.
	setTestVerifyingKey(t, k, ctx)

	exported := am.ExportGenesis(ctx, nil)
	var genesis GenesisState
	if err := json.Unmarshal(exported, &genesis); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if genesis.VerifyingKeyHex == "" {
		t.Fatal("exported genesis should contain non-empty VK hex")
	}
	if genesis.ZKPCircuitID != MembershipCircuitID {
		t.Fatalf("exported circuit id = %q", genesis.ZKPCircuitID)
	}

	// Verify the hex is valid.
	vkBytes, err := hex.DecodeString(genesis.VerifyingKeyHex)
	if err != nil {
		t.Fatalf("invalid VK hex: %v", err)
	}
	if len(vkBytes) == 0 {
		t.Fatal("VK bytes should not be empty")
	}
	if genesis.VerifyingKeySHA256 != VerifyingKeyFingerprint(vkBytes) {
		t.Fatal("exported VK fingerprint mismatch")
	}
}

func TestExportGenesisNoVKWhenNotInitialized(t *testing.T) {
	am, k, ctx := setupModuleForGenesis(t)

	// Create a domain but do NOT use ZKP.
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "NoZKPDomain", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 500_000*PNYXUnit)))

	exported := am.ExportGenesis(ctx, nil)
	var genesis GenesisState
	if err := json.Unmarshal(exported, &genesis); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if genesis.VerifyingKeyHex != "" || genesis.VerifyingKeySHA256 != "" || genesis.ZKPCircuitID != "" {
		t.Fatal("VK configuration should be empty when ZKP was not initialized")
	}
}

func TestInitGenesisWithVK(t *testing.T) {
	// Setup first module instance and generate VK.
	_, k1, ctx1 := setupModuleForGenesis(t)
	vkBytes1 := setTestVerifyingKey(t, k1, ctx1)
	vkHex := hex.EncodeToString(vkBytes1)

	// Create a new module instance and InitGenesis with VK.
	am2, k2, ctx2 := setupModuleForGenesis(t)
	genesisData := GenesisState{
		Domains:            []Domain{},
		Validators:         []GenesisValidator{},
		ZKPCircuitID:       MembershipCircuitID,
		VerifyingKeyHex:    vkHex,
		VerifyingKeySHA256: VerifyingKeyFingerprint(vkBytes1),
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

func TestInitGenesisBaselinesDomainPayoutSnapshots(t *testing.T) {
	am, keeper, ctx := setupModuleForGenesis(t)
	const payouts int64 = 42_000
	admin := sdk.AccAddress("snapshot-admin")
	genesisData := GenesisState{Domains: []Domain{{
		Name: "Restored", Admin: admin, Members: []string{admin.String()},
		Treasury: sdk.NewCoins(), Issues: []Issue{}, PermissionReg: []string{},
		TotalPayouts: payouts,
	}}}
	bz, err := json.Marshal(genesisData)
	if err != nil {
		t.Fatal(err)
	}
	am.InitGenesis(ctx, nil, bz)

	var snapshot int64
	stored := ctx.KVStore(keeper.StoreKey).Get(domainPayoutSnapshotKey("Restored"))
	if stored == nil {
		t.Fatal("missing restored domain payout snapshot")
	}
	keeper.cdc.MustUnmarshalLengthPrefixed(stored, &snapshot)
	if snapshot != payouts {
		t.Fatalf("snapshot = %d, want %d", snapshot, payouts)
	}
}

func TestGenesisRoundTrip(t *testing.T) {
	// Create state in first module.
	am1, k1, ctx1 := setupModuleForGenesis(t)

	admin := sdk.AccAddress("admin1")
	k1.CreateDomain(ctx1, "RoundTripDomain", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 500_000*PNYXUnit)))
	validatorAddr := sdk.AccAddress("validator1").String()
	k1.AddMember(ctx1, "RoundTripDomain", validatorAddr, admin)

	pubKey := testPubKey("roundtrip-val")
	stake := sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 100_000*PNYXUnit))
	k1.RegisterValidator(ctx1, validatorAddr, pubKey, stake, "RoundTripDomain")

	// Initialize VK.
	setTestVerifyingKey(t, k1, ctx1)

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
	v, found := k2.GetValidator(ctx2, validatorAddr)
	if !found {
		t.Fatal("validator should exist after round-trip")
	}
	if v.Stake.AmountOf(PNYXDenom).Int64() != 100_000*PNYXUnit {
		t.Fatalf("expected stake 100000, got %d", v.Stake.AmountOf(PNYXDenom).Int64())
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
	_, err := DeserializeVerifyingKey(vkBytes)
	if err != nil {
		t.Fatalf("DeserializeVerifyingKey failed after round-trip: %v", err)
	}
}

func TestDefaultGenesisStateValidates(t *testing.T) {
	if err := ValidateGenesisState(DefaultGenesisState()); err != nil {
		t.Fatalf("default genesis rejected: %v", err)
	}
}

func TestGenesisRoundTripPreservesMultiDomainActiveValidator(t *testing.T) {
	am1, k1, ctx1 := setupModuleForGenesis(t)

	admin := sdk.AccAddress("admin-md")
	operator := sdk.AccAddress("operator-md").String()
	k1.CreateDomain(ctx1, "DomainA", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 500_000*PNYXUnit)))
	k1.CreateDomain(ctx1, "DomainB", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 500_000*PNYXUnit)))
	if err := k1.AddMember(ctx1, "DomainA", operator, admin); err != nil {
		t.Fatal(err)
	}
	if err := k1.AddMember(ctx1, "DomainB", operator, admin); err != nil {
		t.Fatal(err)
	}
	pubKey := testPubKey("gh60-multidomain")
	stake := sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 2*rewards.StakeMin))
	if err := k1.RegisterValidator(ctx1, operator, pubKey, stake, "DomainA"); err != nil {
		t.Fatalf("RegisterValidator failed: %v", err)
	}
	validator, found := k1.GetValidator(ctx1, operator)
	if !found {
		t.Fatal("validator missing after registration")
	}
	validator.Domains = []string{"DomainA", "DomainB"}
	k1.SetValidator(ctx1, validator)

	exported := am1.ExportGenesis(ctx1, nil)
	var genesis GenesisState
	if err := json.Unmarshal(exported, &genesis); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(genesis.Validators) != 1 {
		t.Fatalf("expected 1 validator, got %d", len(genesis.Validators))
	}
	gv := genesis.Validators[0]
	if gv.Domain != "DomainA" || len(gv.Domains) != 2 || gv.Domains[0] != "DomainA" || gv.Domains[1] != "DomainB" {
		t.Fatalf("export flattened multi-domain state: domain=%q domains=%v", gv.Domain, gv.Domains)
	}
	if gv.Active == nil || !*gv.Active {
		t.Fatal("active validator was not classified active")
	}
	if gv.Power != 2 {
		t.Fatalf("exported power = %d, want 2", gv.Power)
	}

	am2, k2, ctx2 := setupModuleForGenesis(t)
	updates := am2.InitGenesis(ctx2, nil, exported)
	if len(updates) != 1 || updates[0].Power != 2 {
		t.Fatalf("init updates = %v, want one power-2 update", updates)
	}
	restored, found := k2.GetValidator(ctx2, operator)
	if !found {
		t.Fatal("validator should exist after round-trip")
	}
	if len(restored.Domains) != 2 || restored.Domains[0] != "DomainA" || restored.Domains[1] != "DomainB" {
		t.Fatalf("restored domains = %v, want [DomainA DomainB]", restored.Domains)
	}
	if restored.Power != 2 || restored.Jailed ||
		restored.Stake.AmountOf(PNYXDenom).Int64() != 2*rewards.StakeMin {
		t.Fatalf("restored validator drifted: %+v", restored)
	}
}

func TestGenesisRoundTripRetainsInactiveClaimsExactly(t *testing.T) {
	am1, k1, ctx1 := setupModuleForGenesis(t)

	admin := sdk.AccAddress("admin-ia")
	k1.CreateDomain(ctx1, "RetainDomain", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 500_000*PNYXUnit)))
	operators := []string{
		sdk.AccAddress("operator-under").String(),
		sdk.AccAddress("operator-excluded").String(),
		sdk.AccAddress("operator-jailed").String(),
	}
	for i, operator := range operators {
		if err := k1.AddMember(ctx1, "RetainDomain", operator, admin); err != nil {
			t.Fatal(err)
		}
		stake := sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, rewards.StakeMin))
		if err := k1.RegisterValidator(ctx1, operator, testPubKey("gh60-inactive-"+string(rune('a'+i))), stake, "RetainDomain"); err != nil {
			t.Fatalf("RegisterValidator failed: %v", err)
		}
	}

	// Under-staked claim: slashed below minimum, disabled, stake retained.
	under, _ := k1.GetValidator(ctx1, operators[0])
	under.Stake = sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, rewards.StakeMin/2))
	under.Jailed = true
	under.Power = 0
	k1.SetValidator(ctx1, under)

	// Excluded claim: no domain membership remains, custody retained.
	excluded, _ := k1.GetValidator(ctx1, operators[1])
	excluded.Domains = nil
	excluded.Jailed = true
	excluded.Power = 0
	k1.SetValidator(ctx1, excluded)

	// Downtime-jailed claim: domain and stored power retained until unjail.
	jailed, _ := k1.GetValidator(ctx1, operators[2])
	jailed.Jailed = true
	jailed.JailedUntil = 1_700_000_600
	jailed.MissedBlocks = 7
	k1.SetValidator(ctx1, jailed)

	exported := am1.ExportGenesis(ctx1, nil)
	var genesis GenesisState
	if err := json.Unmarshal(exported, &genesis); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(genesis.Validators) != 3 {
		t.Fatalf("expected 3 validators, got %d", len(genesis.Validators))
	}
	for _, gv := range genesis.Validators {
		if gv.Active == nil || *gv.Active {
			t.Fatalf("inactive claim %q was classified active", gv.OperatorAddr)
		}
	}

	am2, k2, ctx2 := setupModuleForGenesis(t)
	updates := am2.InitGenesis(ctx2, nil, exported)
	if len(updates) != 0 {
		t.Fatalf("inactive claims emitted consensus updates: %v", updates)
	}

	restoredUnder, found := k2.GetValidator(ctx2, operators[0])
	if !found || restoredUnder.Stake.AmountOf(PNYXDenom).Int64() != rewards.StakeMin/2 ||
		!restoredUnder.Jailed || restoredUnder.Power != 0 ||
		len(restoredUnder.Domains) != 1 || restoredUnder.Domains[0] != "RetainDomain" {
		t.Fatalf("under-staked claim drifted: %+v", restoredUnder)
	}
	restoredExcluded, found := k2.GetValidator(ctx2, operators[1])
	if !found || restoredExcluded.Stake.AmountOf(PNYXDenom).Int64() != rewards.StakeMin ||
		!restoredExcluded.Jailed || restoredExcluded.Power != 0 || len(restoredExcluded.Domains) != 0 {
		t.Fatalf("excluded claim drifted: %+v", restoredExcluded)
	}
	restoredJailed, found := k2.GetValidator(ctx2, operators[2])
	if !found || !restoredJailed.Jailed || restoredJailed.JailedUntil != 1_700_000_600 ||
		restoredJailed.MissedBlocks != 7 || restoredJailed.Power != 1 ||
		len(restoredJailed.Domains) != 1 || restoredJailed.Domains[0] != "RetainDomain" {
		t.Fatalf("jailed claim drifted: %+v", restoredJailed)
	}
}

func TestInitGenesisAcceptsLegacyValidatorRecord(t *testing.T) {
	admin := sdk.AccAddress("legacy-admin")
	operator := sdk.AccAddress("legacy-operator").String()
	legacy := GenesisState{
		Domains: []Domain{{
			Name: "Legacy", Admin: admin,
			Members:  []string{admin.String(), operator},
			Treasury: sdk.NewCoins(), Issues: []Issue{}, PermissionReg: []string{},
		}},
		Validators: []GenesisValidator{{
			OperatorAddr: operator,
			PubKey:       testPubKey("gh60-legacy"),
			Stake:        rewards.StakeMin,
			Domain:       "Legacy",
		}},
	}
	bz, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}

	// The legacy wire format must not grow GH-60 keys.
	var wire struct {
		Validators []map[string]json.RawMessage `json:"validators"`
	}
	if err := json.Unmarshal(bz, &wire); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"active", "domains", "power"} {
		if _, present := wire.Validators[0][key]; present {
			t.Fatalf("legacy record unexpectedly carries %q", key)
		}
	}

	am, k, ctx := setupModuleForGenesis(t)
	updates := am.InitGenesis(ctx, nil, bz)
	if len(updates) != 1 || updates[0].Power != 1 {
		t.Fatalf("legacy init updates = %v, want one power-1 update", updates)
	}
	restored, found := k.GetValidator(ctx, operator)
	if !found {
		t.Fatal("legacy validator missing after init")
	}
	if len(restored.Domains) != 1 || restored.Domains[0] != "Legacy" || restored.Power != 1 || restored.Jailed {
		t.Fatalf("legacy validator restored incorrectly: %+v", restored)
	}
}

func TestInitGenesisRejectsResurrectedInactiveClaim(t *testing.T) {
	am1, k1, ctx1 := setupModuleForGenesis(t)

	admin := sdk.AccAddress("admin-rs")
	operator := sdk.AccAddress("operator-rs").String()
	k1.CreateDomain(ctx1, "ResurrectDomain", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 500_000*PNYXUnit)))
	if err := k1.AddMember(ctx1, "ResurrectDomain", operator, admin); err != nil {
		t.Fatal(err)
	}
	stake := sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, rewards.StakeMin))
	if err := k1.RegisterValidator(ctx1, operator, testPubKey("gh60-resurrect"), stake, "ResurrectDomain"); err != nil {
		t.Fatalf("RegisterValidator failed: %v", err)
	}
	validator, _ := k1.GetValidator(ctx1, operator)
	validator.Domains = nil
	validator.Jailed = true
	validator.Power = 0
	k1.SetValidator(ctx1, validator)

	exported := am1.ExportGenesis(ctx1, nil)
	var genesis GenesisState
	if err := json.Unmarshal(exported, &genesis); err != nil {
		t.Fatal(err)
	}
	active := true
	genesis.Validators[0].Active = &active
	tampered, err := json.Marshal(genesis)
	if err != nil {
		t.Fatal(err)
	}

	am2, _, ctx2 := setupModuleForGenesis(t)
	func() {
		defer func() {
			if recover() == nil {
				t.Error("resurrected inactive claim was imported")
			}
		}()
		am2.InitGenesis(ctx2, nil, tampered)
	}()
}
