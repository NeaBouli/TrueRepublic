package main

// Explicit compatibility adapters for upstream wasmd interfaces that
// TrueRepublic deliberately does not implement. Proof of Domain is not Cosmos
// delegated staking, so these surfaces must never fabricate staking state.

import (
	"context"
	"errors"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var errUnsupportedStakingSurface = errors.New("unsupported protocol surface: TrueRepublic uses Proof of Domain and does not mount x/staking or x/distribution")

// UnsupportedWasmStakingKeeper satisfies wasmd's constructor interface while
// rejecting every x/staking-shaped operation. Query plugins are overridden as
// well, so this is defense in depth against upstream wiring changes.
type UnsupportedWasmStakingKeeper struct{}

var _ wasmtypes.StakingKeeper = UnsupportedWasmStakingKeeper{}

func (UnsupportedWasmStakingKeeper) BondDenom(_ context.Context) (string, error) {
	return "", errUnsupportedStakingSurface
}

func (UnsupportedWasmStakingKeeper) GetValidator(_ context.Context, _ sdk.ValAddress) (stakingtypes.Validator, error) {
	return stakingtypes.Validator{}, errUnsupportedStakingSurface
}

func (UnsupportedWasmStakingKeeper) GetBondedValidatorsByPower(_ context.Context) ([]stakingtypes.Validator, error) {
	return nil, errUnsupportedStakingSurface
}

func (UnsupportedWasmStakingKeeper) GetAllDelegatorDelegations(_ context.Context, _ sdk.AccAddress) ([]stakingtypes.Delegation, error) {
	return nil, errUnsupportedStakingSurface
}

func (UnsupportedWasmStakingKeeper) GetDelegation(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) (stakingtypes.Delegation, error) {
	return stakingtypes.Delegation{}, errUnsupportedStakingSurface
}

func (UnsupportedWasmStakingKeeper) HasReceivingRedelegation(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) (bool, error) {
	return false, errUnsupportedStakingSurface
}

// UnsupportedWasmDistributionKeeper rejects every standard distribution
// query instead of returning fabricated empty rewards or withdraw addresses.
type UnsupportedWasmDistributionKeeper struct{}

var _ wasmtypes.DistributionKeeper = UnsupportedWasmDistributionKeeper{}

func (UnsupportedWasmDistributionKeeper) DelegatorWithdrawAddress(_ context.Context, _ *distrtypes.QueryDelegatorWithdrawAddressRequest) (*distrtypes.QueryDelegatorWithdrawAddressResponse, error) {
	return nil, errUnsupportedStakingSurface
}

func (UnsupportedWasmDistributionKeeper) DelegationRewards(_ context.Context, _ *distrtypes.QueryDelegationRewardsRequest) (*distrtypes.QueryDelegationRewardsResponse, error) {
	return nil, errUnsupportedStakingSurface
}

func (UnsupportedWasmDistributionKeeper) DelegationTotalRewards(_ context.Context, _ *distrtypes.QueryDelegationTotalRewardsRequest) (*distrtypes.QueryDelegationTotalRewardsResponse, error) {
	return nil, errUnsupportedStakingSurface
}

func (UnsupportedWasmDistributionKeeper) DelegatorValidators(_ context.Context, _ *distrtypes.QueryDelegatorValidatorsRequest) (*distrtypes.QueryDelegatorValidatorsResponse, error) {
	return nil, errUnsupportedStakingSurface
}

func rejectWasmStakingQuery(_ sdk.Context, _ *wasmvmtypes.StakingQuery) ([]byte, error) {
	return nil, wasmvmtypes.UnsupportedRequest{Kind: errUnsupportedStakingSurface.Error()}
}

func rejectWasmDistributionQuery(_ sdk.Context, _ *wasmvmtypes.DistributionQuery) ([]byte, error) {
	return nil, wasmvmtypes.UnsupportedRequest{Kind: errUnsupportedStakingSurface.Error()}
}

func rejectWasmStakingMessage(_ sdk.AccAddress, _ *wasmvmtypes.StakingMsg) ([]sdk.Msg, error) {
	return nil, errUnsupportedStakingSurface
}

func rejectWasmDistributionMessage(_ sdk.AccAddress, _ *wasmvmtypes.DistributionMsg) ([]sdk.Msg, error) {
	return nil, errUnsupportedStakingSurface
}

// NoOpWasmValidatorSetSource satisfies the pinned wasmd v0.53.4 AppModule
// constructor. That version stores this source but does not invoke it; PoD
// validator updates remain owned exclusively by x/truedemocracy.
type NoOpWasmValidatorSetSource struct{}

var _ wasmkeeper.ValidatorSetSource = NoOpWasmValidatorSetSource{}

func (NoOpWasmValidatorSetSource) ApplyAndReturnValidatorSetUpdates(_ context.Context) ([]abci.ValidatorUpdate, error) {
	return nil, nil
}
