package main

// Stub keeper implementations for wasmd integration.
// These stubs provide no-op/error responses for keepers not yet wired
// (staking, distribution). IBC keepers are now real (wired in app.go).

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/cometbft/cometbft/abci/types"
)

var errNotAvailable = errors.New("feature not available: requires module integration")

// --- Staking Keeper Stub (no x/staking module) ---

type StubStakingKeeper struct{}

func (StubStakingKeeper) BondDenom(_ context.Context) (string, error) {
	return "pnyx", nil
}

func (StubStakingKeeper) GetValidator(_ context.Context, _ sdk.ValAddress) (stakingtypes.Validator, error) {
	return stakingtypes.Validator{}, errNotAvailable
}

func (StubStakingKeeper) GetBondedValidatorsByPower(_ context.Context) ([]stakingtypes.Validator, error) {
	return nil, nil
}

func (StubStakingKeeper) GetAllDelegatorDelegations(_ context.Context, _ sdk.AccAddress) ([]stakingtypes.Delegation, error) {
	return nil, nil
}

func (StubStakingKeeper) GetDelegation(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) (stakingtypes.Delegation, error) {
	return stakingtypes.Delegation{}, errNotAvailable
}

func (StubStakingKeeper) HasReceivingRedelegation(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) (bool, error) {
	return false, nil
}

// --- Distribution Keeper Stub (no x/distribution module) ---

type StubDistributionKeeper struct{}

func (StubDistributionKeeper) DelegatorWithdrawAddress(_ context.Context, _ *distrtypes.QueryDelegatorWithdrawAddressRequest) (*distrtypes.QueryDelegatorWithdrawAddressResponse, error) {
	return &distrtypes.QueryDelegatorWithdrawAddressResponse{}, nil
}

func (StubDistributionKeeper) DelegationRewards(_ context.Context, _ *distrtypes.QueryDelegationRewardsRequest) (*distrtypes.QueryDelegationRewardsResponse, error) {
	return &distrtypes.QueryDelegationRewardsResponse{}, nil
}

func (StubDistributionKeeper) DelegationTotalRewards(_ context.Context, _ *distrtypes.QueryDelegationTotalRewardsRequest) (*distrtypes.QueryDelegationTotalRewardsResponse, error) {
	return &distrtypes.QueryDelegationTotalRewardsResponse{}, nil
}

func (StubDistributionKeeper) DelegatorValidators(_ context.Context, _ *distrtypes.QueryDelegatorValidatorsRequest) (*distrtypes.QueryDelegatorValidatorsResponse, error) {
	return &distrtypes.QueryDelegatorValidatorsResponse{}, nil
}

// --- Validator Set Source Stub (used by wasm module genesis) ---

type StubValidatorSetSource struct{}

func (StubValidatorSetSource) ApplyAndReturnValidatorSetUpdates(_ context.Context) ([]abci.ValidatorUpdate, error) {
	return nil, nil
}
