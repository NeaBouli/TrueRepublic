package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cometbft/cometbft/proto/tendermint/types"

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
	if got := updated[truedemocracy.ModuleName]; got != 2 {
		t.Fatalf("truedemocracy module version = %d, want 2", got)
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
