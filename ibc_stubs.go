package main

// IBC-specific keeper stub for ibc-go v8 integration.
// TrueRepublic uses Proof-of-Domain consensus (x/truedemocracy/validator.go)
// instead of standard x/staking. The real x/upgrade keeper is wired directly
// in app.go, so only the staking compatibility boundary remains here.

import (
	"context"
	"time"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// IBCStakingKeeper satisfies ibc-go's clienttypes.StakingKeeper interface.
// IBC light-client verification needs historical validator sets and unbonding
// period. We return a sensible default unbonding time and empty historical info.
type IBCStakingKeeper struct{ initialized bool }

func (IBCStakingKeeper) GetHistoricalInfo(_ context.Context, _ int64) (stakingtypes.HistoricalInfo, error) {
	return stakingtypes.HistoricalInfo{}, errNotAvailable
}

func (IBCStakingKeeper) UnbondingTime(_ context.Context) (time.Duration, error) {
	return 3 * 7 * 24 * time.Hour, nil // 3 weeks (standard Cosmos default)
}
