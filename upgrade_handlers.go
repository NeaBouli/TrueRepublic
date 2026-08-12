package main

import (
	"context"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"truerepublic/x/truedemocracy"
)

const governedUpgradePlanV041 = "v0.4.1"

const governedUpgradeFailureFixtureV041 = "v0.4.1-gh184-failure-fixture"

var governedUpgradeMarkerV041 = []byte("truerepublic:upgrade:v0.4.1")

// registerReleaseUpgradeHandlers keeps the binary/plan binding explicit. The
// infrastructure binary has no handler and therefore halts at the agreed
// height. Only an artifact built with main.upgradePlan=v0.4.1 can apply this
// migration; main.version remains the immutable source commit identity.
func (app *TrueRepublicApp) registerReleaseUpgradeHandlers() {
	switch upgradePlan {
	case governedUpgradePlanV041:
		app.registerUpgradeHandler(governedUpgradePlanV041, app.v041UpgradeHandler)
	case governedUpgradeFailureFixtureV041:
		// This non-release build identity exists only for the opt-in GH-184
		// process harness. It proves a write followed by a handler error is
		// discarded before the fixed v0.4.1 artifact is started.
		app.registerUpgradeHandler(governedUpgradePlanV041, app.v041FailingFixtureHandler)
	}
}

// v041UpgradeHandler runs registered module migrations and records a
// deterministic application marker. x/upgrade executes this inside the cached
// FinalizeBlock, so any error discards both module and marker writes.
func (app *TrueRepublicApp) v041UpgradeHandler(
	ctx context.Context,
	plan upgradetypes.Plan,
	fromVM module.VersionMap,
) (module.VersionMap, error) {
	if plan.Name != governedUpgradePlanV041 {
		return nil, fmt.Errorf("unexpected upgrade plan %q", plan.Name)
	}
	updatedVM, err := app.mm.RunMigrations(ctx, app.configurator, fromVM)
	if err != nil {
		return nil, err
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(app.keys[truedemocracy.ModuleName])
	if store.Has(governedUpgradeMarkerV041) {
		return nil, fmt.Errorf("upgrade %q migration marker already exists", plan.Name)
	}
	store.Set(governedUpgradeMarkerV041, []byte{1})
	return updatedVM, nil
}

func (app *TrueRepublicApp) v041FailingFixtureHandler(
	ctx context.Context,
	plan upgradetypes.Plan,
	fromVM module.VersionMap,
) (module.VersionMap, error) {
	if plan.Name != governedUpgradePlanV041 {
		return nil, fmt.Errorf("unexpected upgrade plan %q", plan.Name)
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.KVStore(app.keys[truedemocracy.ModuleName]).Set(governedUpgradeMarkerV041, []byte{0xff})
	return nil, fmt.Errorf("intentional GH-184 migration failure after partial cached write")
}
