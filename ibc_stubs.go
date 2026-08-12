package main

// Explicit fail-closed adapter required by the pinned ibc-go v8 constructor.
// TrueRepublic uses Proof of Domain instead of x/staking; it must not invent
// historical staking state or a PoS unbonding period.

import (
	"context"
	"time"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
)

// UnsupportedIBCStakingKeeper stays non-zero when passed to ibc-go because
// ibckeeper.NewKeeper rejects zero-valued compatibility keepers. Both methods
// fail closed if a future ibc-go path attempts self-consensus validation.
type UnsupportedIBCStakingKeeper struct{ registered bool }

var _ ibcclienttypes.StakingKeeper = UnsupportedIBCStakingKeeper{}

func (UnsupportedIBCStakingKeeper) GetHistoricalInfo(_ context.Context, _ int64) (stakingtypes.HistoricalInfo, error) {
	return stakingtypes.HistoricalInfo{}, errUnsupportedStakingSurface
}

func (UnsupportedIBCStakingKeeper) UnbondingTime(_ context.Context) (time.Duration, error) {
	return 0, errUnsupportedStakingSurface
}
