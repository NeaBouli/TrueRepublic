package main

// IBC-specific keeper stubs for ibc-go v8 integration.
// TrueRepublic uses Proof-of-Domain consensus (x/truedemocracy/validator.go)
// instead of standard x/staking, and does not wire x/upgrade.
// These stubs satisfy the interfaces required by ibckeeper.NewKeeper.

import (
	"context"
	"time"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// IBCStakingKeeper satisfies ibc-go's clienttypes.StakingKeeper interface.
// IBC light-client verification needs historical validator sets and unbonding
// period. We return a sensible default unbonding time and empty historical info.
type IBCStakingKeeper struct{}

func (IBCStakingKeeper) GetHistoricalInfo(_ context.Context, _ int64) (stakingtypes.HistoricalInfo, error) {
	return stakingtypes.HistoricalInfo{}, errNotAvailable
}

func (IBCStakingKeeper) UnbondingTime(_ context.Context) (time.Duration, error) {
	return 3 * 7 * 24 * time.Hour, nil // 3 weeks (standard Cosmos default)
}

// IBCUpgradeKeeper satisfies ibc-go's clienttypes.UpgradeKeeper interface.
// Without x/upgrade wired, all upgrade operations are no-ops. IBC client
// upgrades are not supported until an upgrade module is added.
type IBCUpgradeKeeper struct{}

func (IBCUpgradeKeeper) ClearIBCState(_ context.Context, _ int64) error {
	return nil
}

func (IBCUpgradeKeeper) GetUpgradePlan(_ context.Context) (upgradetypes.Plan, error) {
	return upgradetypes.Plan{}, nil // no active upgrade plan
}

func (IBCUpgradeKeeper) GetUpgradedClient(_ context.Context, _ int64) ([]byte, error) {
	return nil, nil
}

func (IBCUpgradeKeeper) SetUpgradedClient(_ context.Context, _ int64, _ []byte) error {
	return nil
}

func (IBCUpgradeKeeper) GetUpgradedConsensusState(_ context.Context, _ int64) ([]byte, error) {
	return nil, nil
}

func (IBCUpgradeKeeper) SetUpgradedConsensusState(_ context.Context, _ int64, _ []byte) error {
	return nil
}

func (IBCUpgradeKeeper) ScheduleUpgrade(_ context.Context, _ upgradetypes.Plan) error {
	return errNotAvailable // upgrades not supported without x/upgrade
}
