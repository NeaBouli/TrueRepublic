package main

// Stub keeper implementations for wasmd integration.
// These stubs provide no-op/error responses for keepers not yet wired
// (staking, distribution, IBC). Real implementations will be added when
// those features are integrated (v0.3.0 Weeks 7-9 for IBC).

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/cometbft/cometbft/abci/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types" //nolint:staticcheck
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
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

// --- IBC Stubs (no IBC until v0.3.0 Weeks 7-9) ---

type StubChannelKeeper struct{}

func (StubChannelKeeper) GetChannel(_ sdk.Context, _, _ string) (channeltypes.Channel, bool) {
	return channeltypes.Channel{}, false
}

func (StubChannelKeeper) GetNextSequenceSend(_ sdk.Context, _, _ string) (uint64, bool) {
	return 0, false
}

func (StubChannelKeeper) ChanCloseInit(_ sdk.Context, _, _ string, _ *capabilitytypes.Capability) error {
	return errNotAvailable
}

func (StubChannelKeeper) GetAllChannels(_ sdk.Context) []channeltypes.IdentifiedChannel {
	return nil
}

func (StubChannelKeeper) SetChannel(_ sdk.Context, _, _ string, _ channeltypes.Channel) {}

func (StubChannelKeeper) GetAllChannelsWithPortPrefix(_ sdk.Context, _ string) []channeltypes.IdentifiedChannel {
	return nil
}

type StubICS4Wrapper struct{}

func (StubICS4Wrapper) SendPacket(_ sdk.Context, _ *capabilitytypes.Capability, _, _ string, _ clienttypes.Height, _ uint64, _ []byte) (uint64, error) {
	return 0, errNotAvailable
}

func (StubICS4Wrapper) WriteAcknowledgement(_ sdk.Context, _ *capabilitytypes.Capability, _ ibcexported.PacketI, _ ibcexported.Acknowledgement) error {
	return errNotAvailable
}

type StubPortKeeper struct{}

func (StubPortKeeper) BindPort(_ sdk.Context, _ string) *capabilitytypes.Capability {
	return &capabilitytypes.Capability{Index: 1}
}

type StubCapabilityKeeper struct{}

func (StubCapabilityKeeper) GetCapability(_ sdk.Context, _ string) (*capabilitytypes.Capability, bool) {
	return nil, false
}

func (StubCapabilityKeeper) ClaimCapability(_ sdk.Context, _ *capabilitytypes.Capability, _ string) error {
	return nil
}

func (StubCapabilityKeeper) AuthenticateCapability(_ sdk.Context, _ *capabilitytypes.Capability, _ string) bool {
	return false
}

type StubICS20TransferPortSource struct{}

func (StubICS20TransferPortSource) GetPort(_ sdk.Context) string {
	return "transfer"
}

// --- Validator Set Source Stub (used by wasm module genesis) ---

type StubValidatorSetSource struct{}

func (StubValidatorSetSource) ApplyAndReturnValidatorSetUpdates(_ context.Context) ([]abci.ValidatorUpdate, error) {
	return nil, nil
}
