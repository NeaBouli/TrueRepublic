package main

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"truerepublic/x/truedemocracy"
)

func TestV041UpgradeHandlerAppliesMarkerExactlyOnce(t *testing.T) {
	app := newGenesisTestApp(t)
	if err := initGenesisApp(app, defaultGenesisForApp(app)); err != nil {
		t.Fatal(err)
	}
	sdkCtx := app.NewUncachedContext(false, types.Header{Height: 1})
	ctx := sdkCtx
	plan := upgradetypes.Plan{Name: governedUpgradePlanV041, Height: 1}

	fromVM := app.mm.GetVersionMap()
	fromVM[truedemocracy.ModuleName] = 1
	updated, err := app.v041UpgradeHandler(ctx, plan, fromVM)
	if err != nil || updated == nil {
		t.Fatalf("v0.4.1 handler failed: versions=%v err=%v", updated, err)
	}
	wantVersion := (truedemocracy.AppModule{}).ConsensusVersion()
	if got := updated[truedemocracy.ModuleName]; got != wantVersion {
		t.Fatalf("truedemocracy module version = %d, want current %d", got, wantVersion)
	}
	marker := sdkCtx.KVStore(app.keys[truedemocracy.ModuleName]).Get(governedUpgradeMarkerV041)
	if !bytes.Equal(marker, []byte{1}) {
		t.Fatalf("migration marker = %x", marker)
	}
	if _, err := app.v041UpgradeHandler(ctx, plan, updated); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("duplicate migration did not fail closed: %v", err)
	}
	if _, err := app.v041UpgradeHandler(ctx, upgradetypes.Plan{Name: "v0.4.2"}, updated); err == nil {
		t.Fatal("handler accepted a different plan name")
	}
}

func TestV041FailingFixtureWritesOnlyToProvidedContext(t *testing.T) {
	app := newGenesisTestApp(t)
	if err := initGenesisApp(app, defaultGenesisForApp(app)); err != nil {
		t.Fatal(err)
	}
	baseCtx := app.NewUncachedContext(false, types.Header{Height: 1})
	cacheCtx, _ := baseCtx.CacheContext()
	plan := upgradetypes.Plan{Name: governedUpgradePlanV041, Height: 1}

	if _, err := app.v041FailingFixtureHandler(cacheCtx, plan, app.mm.GetVersionMap()); err == nil || !strings.Contains(err.Error(), "intentional") {
		t.Fatalf("failure fixture result = %v", err)
	}
	if marker := cacheCtx.KVStore(app.keys[truedemocracy.ModuleName]).Get(governedUpgradeMarkerV041); !bytes.Equal(marker, []byte{0xff}) {
		t.Fatalf("cached failure marker = %x", marker)
	}
	if marker := baseCtx.KVStore(app.keys[truedemocracy.ModuleName]).Get(governedUpgradeMarkerV041); marker != nil {
		t.Fatalf("discarded cache leaked marker = %x", marker)
	}
	if _, err := app.v041FailingFixtureHandler(context.Background(), upgradetypes.Plan{Name: "wrong"}, nil); err == nil {
		t.Fatal("failure fixture accepted a different plan name")
	}
}

// seedV2StoneState writes legacy (pre-GH-306) stone keys without consumed
// reward markers, exactly as a version-2 chain stores them.
func seedV2StoneState(t *testing.T, app *TrueRepublicApp, ctx sdk.Context) (sdk.AccAddress, sdk.AccAddress) {
	t.Helper()
	alice := sdk.AccAddress(bytes.Repeat([]byte{0x51}, 20))
	bob := sdk.AccAddress(bytes.Repeat([]byte{0x52}, 20))
	app.tdKeeper.CreateDomain(ctx, "V042Domain", alice, sdk.NewCoins())
	if err := app.tdKeeper.AddMember(ctx, "V042Domain", bob.String(), alice); err != nil {
		t.Fatal(err)
	}
	store := ctx.KVStore(app.keys[truedemocracy.ModuleName])
	store.Set([]byte("stone:i:V042Domain:"+alice.String()), []byte("LegacyIssue"))
	store.Set([]byte("stone:s:V042Domain:LegacyIssue:"+bob.String()), []byte("LegacySuggestion"))
	var markers int
	app.tdKeeper.IterateStoneRewardRecords(ctx, func(truedemocracy.StoneRewardRecord) bool {
		markers++
		return false
	})
	if markers != 0 {
		t.Fatalf("v2 seed state already has %d reward markers", markers)
	}
	return alice, bob
}

func TestV042UpgradeHandlerMigratesV2ToV3ExactlyOnce(t *testing.T) {
	app := newGenesisTestApp(t)
	if err := initGenesisApp(app, defaultGenesisForApp(app)); err != nil {
		t.Fatal(err)
	}
	ctx := app.NewUncachedContext(false, types.Header{Height: 1})
	alice, bob := seedV2StoneState(t, app, ctx)
	plan := upgradetypes.Plan{Name: governedUpgradePlanV042, Height: 1}

	// Wrong source versions fail closed before any write.
	for _, from := range []uint64{1, 3} {
		fromVM := app.mm.GetVersionMap()
		fromVM[truedemocracy.ModuleName] = from
		if _, err := app.v042UpgradeHandler(ctx, plan, fromVM); err == nil || !strings.Contains(err.Error(), "requires") {
			t.Fatalf("v0.4.2 from version %d did not fail closed: %v", from, err)
		}
	}
	if _, err := app.v042UpgradeHandler(ctx, upgradetypes.Plan{Name: governedUpgradePlanV041}, app.mm.GetVersionMap()); err == nil {
		t.Fatal("v0.4.2 handler accepted the v0.4.1 plan name")
	}
	if ctx.KVStore(app.keys[truedemocracy.ModuleName]).Has(governedUpgradeMarkerV042) {
		t.Fatal("rejected executions wrote the v0.4.2 marker")
	}

	fromVM := app.mm.GetVersionMap()
	fromVM[truedemocracy.ModuleName] = 2
	updated, err := app.v042UpgradeHandler(ctx, plan, fromVM)
	if err != nil {
		t.Fatalf("v0.4.2 handler failed: %v", err)
	}
	if got, want := updated[truedemocracy.ModuleName], (truedemocracy.AppModule{}).ConsensusVersion(); got != want || want != 3 {
		t.Fatalf("truedemocracy version = %d, want %d (3)", got, want)
	}
	if marker := ctx.KVStore(app.keys[truedemocracy.ModuleName]).Get(governedUpgradeMarkerV042); !bytes.Equal(marker, []byte{1}) {
		t.Fatalf("v0.4.2 marker = %x", marker)
	}
	if ctx.KVStore(app.keys[truedemocracy.ModuleName]).Has(governedUpgradeMarkerV041) {
		t.Fatal("v0.4.2 wrote the v0.4.1 marker")
	}
	issueScope := truedemocracy.StoneRewardRecord{DomainName: "V042Domain", MemberAddr: alice.String()}
	suggestionScope := truedemocracy.StoneRewardRecord{DomainName: "V042Domain", IssueName: "LegacyIssue", MemberAddr: bob.String()}
	var records []truedemocracy.StoneRewardRecord
	app.tdKeeper.IterateStoneRewardRecords(ctx, func(record truedemocracy.StoneRewardRecord) bool {
		records = append(records, record)
		return false
	})
	if len(records) != 2 || !app.tdKeeper.HasStoneRewardRecord(ctx, issueScope) || !app.tdKeeper.HasStoneRewardRecord(ctx, suggestionScope) {
		t.Fatalf("v2->v3 baseline records = %+v, want exactly the two seeded scopes", records)
	}

	// Duplicate execution, even with a version-2 map, fails closed.
	if _, err := app.v042UpgradeHandler(ctx, plan, fromVM); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("duplicate v0.4.2 migration did not fail closed: %v", err)
	}

	// The baseline persists through export, validation and import.
	exported := app.tdModule.ExportGenesis(ctx, app.appCodec)
	var genesis truedemocracy.GenesisState
	if err := json.Unmarshal(exported, &genesis); err != nil {
		t.Fatal(err)
	}
	if err := truedemocracy.ValidateGenesisState(genesis); err != nil {
		t.Fatalf("post-migration export is invalid: %v", err)
	}
	if len(genesis.StoneRewardRecords) != 2 {
		t.Fatalf("exported reward records = %+v, want 2", genesis.StoneRewardRecords)
	}
	// Module-level import into a fresh store: the full-app path additionally
	// requires consensus validators, which this state-only fixture omits.
	imported := newGenesisTestApp(t)
	importedCtx := imported.NewUncachedContext(false, types.Header{Height: 1})
	imported.tdModule.InitGenesis(importedCtx, imported.appCodec, exported)
	if !imported.tdKeeper.HasStoneRewardRecord(importedCtx, issueScope) || !imported.tdKeeper.HasStoneRewardRecord(importedCtx, suggestionScope) {
		t.Fatal("reward records lost across export/import")
	}
}

func TestV042FailingFixtureWritesOnlyToProvidedContext(t *testing.T) {
	app := newGenesisTestApp(t)
	if err := initGenesisApp(app, defaultGenesisForApp(app)); err != nil {
		t.Fatal(err)
	}
	baseCtx := app.NewUncachedContext(false, types.Header{Height: 1})
	cacheCtx, _ := baseCtx.CacheContext()
	plan := upgradetypes.Plan{Name: governedUpgradePlanV042, Height: 1}

	if _, err := app.v042FailingFixtureHandler(cacheCtx, plan, app.mm.GetVersionMap()); err == nil || !strings.Contains(err.Error(), "intentional") {
		t.Fatalf("failure fixture result = %v", err)
	}
	if marker := cacheCtx.KVStore(app.keys[truedemocracy.ModuleName]).Get(governedUpgradeMarkerV042); !bytes.Equal(marker, []byte{0xff}) {
		t.Fatalf("cached failure marker = %x", marker)
	}
	if marker := baseCtx.KVStore(app.keys[truedemocracy.ModuleName]).Get(governedUpgradeMarkerV042); marker != nil {
		t.Fatalf("discarded cache leaked marker = %x", marker)
	}
	if _, err := app.v042FailingFixtureHandler(context.Background(), upgradetypes.Plan{Name: "wrong"}, nil); err == nil {
		t.Fatal("failure fixture accepted a different plan name")
	}
}

// TestReleaseUpgradePlanIdentityRegistersExactlyOneHandler binds each build
// identity to exactly one governed plan name, so a v0.4.1 artifact can never
// apply v0.4.2 and vice versa, and the infrastructure build registers none.
func TestReleaseUpgradePlanIdentityRegistersExactlyOneHandler(t *testing.T) {
	original := upgradePlan
	t.Cleanup(func() { upgradePlan = original })
	cases := []struct {
		identity string
		v041     bool
		v042     bool
	}{
		{identity: "", v041: false, v042: false},
		{identity: governedUpgradePlanV041, v041: true, v042: false},
		{identity: governedUpgradeFailureFixtureV041, v041: true, v042: false},
		{identity: governedUpgradePlanV042, v041: false, v042: true},
		{identity: governedUpgradeFailureFixtureV042, v041: false, v042: true},
	}
	for _, tc := range cases {
		upgradePlan = tc.identity
		app := newGenesisTestApp(t)
		if got := app.upgradeKeeper.HasHandler(governedUpgradePlanV041); got != tc.v041 {
			t.Fatalf("identity %q: v0.4.1 handler registered = %v, want %v", tc.identity, got, tc.v041)
		}
		if got := app.upgradeKeeper.HasHandler(governedUpgradePlanV042); got != tc.v042 {
			t.Fatalf("identity %q: v0.4.2 handler registered = %v, want %v", tc.identity, got, tc.v042)
		}
	}
}
