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

// governedUpgradePlanV042 carries a truedemocracy module that is already at
// consensus version 2 (after v0.4.1 or a version-2 genesis) through the GH-306
// 2→3 migration. v0.4.1 stays unchanged and still reaches version 3 from 1.
const governedUpgradePlanV042 = "v0.4.2"

const governedUpgradeFailureFixtureV042 = "v0.4.2-gh306-failure-fixture"

var governedUpgradeMarkerV042 = []byte("truerepublic:upgrade:v0.4.2")

// governedUpgradeV042FromVersion is the only truedemocracy module version the
// v0.4.2 plan accepts; any other stored version fails closed.
const governedUpgradeV042FromVersion = 2

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
	case governedUpgradePlanV042:
		app.registerUpgradeHandler(governedUpgradePlanV042, app.v042UpgradeHandler)
	case governedUpgradeFailureFixtureV042:
		// Non-release identity for the v0.4.2 rollback proof: a cached write
		// followed by a handler error must be discarded before the fixed
		// v0.4.2 artifact starts.
		app.registerUpgradeHandler(governedUpgradePlanV042, app.v042FailingFixtureHandler)
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

// v042UpgradeHandler runs the GH-306 truedemocracy 2→3 migration, which
// baselines every existing Stones Voting scope as already rewarded, and records
// a deterministic marker. It fails closed for a wrong plan, a module that is
// not at version 2, or a repeated execution. x/upgrade executes it inside the
// cached FinalizeBlock, so any error discards the migration and marker writes.
func (app *TrueRepublicApp) v042UpgradeHandler(
	ctx context.Context,
	plan upgradetypes.Plan,
	fromVM module.VersionMap,
) (module.VersionMap, error) {
	if plan.Name != governedUpgradePlanV042 {
		return nil, fmt.Errorf("unexpected upgrade plan %q", plan.Name)
	}
	if got, found := fromVM[truedemocracy.ModuleName]; !found || got != governedUpgradeV042FromVersion {
		return nil, fmt.Errorf(
			"upgrade %q requires %s module version %d, found %d",
			plan.Name, truedemocracy.ModuleName, governedUpgradeV042FromVersion, got,
		)
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(app.keys[truedemocracy.ModuleName])
	if store.Has(governedUpgradeMarkerV042) {
		return nil, fmt.Errorf("upgrade %q migration marker already exists", plan.Name)
	}
	updatedVM, err := app.mm.RunMigrations(ctx, app.configurator, fromVM)
	if err != nil {
		return nil, err
	}
	if got := updatedVM[truedemocracy.ModuleName]; got != (truedemocracy.AppModule{}).ConsensusVersion() {
		return nil, fmt.Errorf("upgrade %q left %s at version %d", plan.Name, truedemocracy.ModuleName, got)
	}
	store.Set(governedUpgradeMarkerV042, []byte{1})
	return updatedVM, nil
}

func (app *TrueRepublicApp) v042FailingFixtureHandler(
	ctx context.Context,
	plan upgradetypes.Plan,
	fromVM module.VersionMap,
) (module.VersionMap, error) {
	if plan.Name != governedUpgradePlanV042 {
		return nil, fmt.Errorf("unexpected upgrade plan %q", plan.Name)
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.KVStore(app.keys[truedemocracy.ModuleName]).Set(governedUpgradeMarkerV042, []byte{0xff})
	return nil, fmt.Errorf("intentional GH-306 migration failure after partial cached write")
}
