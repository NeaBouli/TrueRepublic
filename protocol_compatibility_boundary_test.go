package main

import (
	"context"
	"errors"
	"testing"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestUnsupportedWasmStakingKeeperFailsClosed(t *testing.T) {
	keeper := UnsupportedWasmStakingKeeper{}
	checks := []func() error{
		func() error { _, err := keeper.BondDenom(context.Background()); return err },
		func() error { _, err := keeper.GetValidator(context.Background(), nil); return err },
		func() error { _, err := keeper.GetBondedValidatorsByPower(context.Background()); return err },
		func() error { _, err := keeper.GetAllDelegatorDelegations(context.Background(), nil); return err },
		func() error { _, err := keeper.GetDelegation(context.Background(), nil, nil); return err },
		func() error { _, err := keeper.HasReceivingRedelegation(context.Background(), nil, nil); return err },
	}
	for index, check := range checks {
		if err := check(); !errors.Is(err, errUnsupportedStakingSurface) {
			t.Errorf("staking adapter method %d returned %v", index, err)
		}
	}
}

func TestUnsupportedWasmDistributionKeeperFailsClosed(t *testing.T) {
	keeper := UnsupportedWasmDistributionKeeper{}
	checks := []func() error{
		func() error {
			_, err := keeper.DelegatorWithdrawAddress(context.Background(), &distrtypes.QueryDelegatorWithdrawAddressRequest{})
			return err
		},
		func() error {
			_, err := keeper.DelegationRewards(context.Background(), &distrtypes.QueryDelegationRewardsRequest{})
			return err
		},
		func() error {
			_, err := keeper.DelegationTotalRewards(context.Background(), &distrtypes.QueryDelegationTotalRewardsRequest{})
			return err
		},
		func() error {
			_, err := keeper.DelegatorValidators(context.Background(), &distrtypes.QueryDelegatorValidatorsRequest{})
			return err
		},
	}
	for index, check := range checks {
		if err := check(); !errors.Is(err, errUnsupportedStakingSurface) {
			t.Errorf("distribution adapter method %d returned %v", index, err)
		}
	}
}

func TestWasmStakingAndDistributionPluginsRejectEveryVariant(t *testing.T) {
	stakingQueries := []*wasmvmtypes.StakingQuery{
		{BondedDenom: &struct{}{}},
		{AllValidators: &wasmvmtypes.AllValidatorsQuery{}},
		{Validator: &wasmvmtypes.ValidatorQuery{}},
		{AllDelegations: &wasmvmtypes.AllDelegationsQuery{}},
		{Delegation: &wasmvmtypes.DelegationQuery{}},
	}
	for _, query := range stakingQueries {
		if _, err := rejectWasmStakingQuery(sdk.Context{}, query); !isUnsupportedWasmRequest(err) {
			t.Errorf("staking query %#v returned %v", query, err)
		}
	}

	distributionQueries := []*wasmvmtypes.DistributionQuery{
		{DelegatorWithdrawAddress: &wasmvmtypes.DelegatorWithdrawAddressQuery{}},
		{DelegationRewards: &wasmvmtypes.DelegationRewardsQuery{}},
		{DelegationTotalRewards: &wasmvmtypes.DelegationTotalRewardsQuery{}},
		{DelegatorValidators: &wasmvmtypes.DelegatorValidatorsQuery{}},
	}
	for _, query := range distributionQueries {
		if _, err := rejectWasmDistributionQuery(sdk.Context{}, query); !isUnsupportedWasmRequest(err) {
			t.Errorf("distribution query %#v returned %v", query, err)
		}
	}

	if _, err := rejectWasmStakingMessage(nil, &wasmvmtypes.StakingMsg{}); !errors.Is(err, errUnsupportedStakingSurface) {
		t.Errorf("staking message returned %v", err)
	}
	if _, err := rejectWasmDistributionMessage(nil, &wasmvmtypes.DistributionMsg{}); !errors.Is(err, errUnsupportedStakingSurface) {
		t.Errorf("distribution message returned %v", err)
	}
}

func TestUnsupportedStakingAndDistributionSurfacesRemainUnmounted(t *testing.T) {
	app := newGenesisTestApp(t)
	for _, moduleName := range []string{stakingtypes.ModuleName, distrtypes.ModuleName} {
		if _, found := ModuleBasics[moduleName]; found {
			t.Errorf("unsupported module %q is registered in ModuleBasics", moduleName)
		}
		if _, found := app.mm.Modules[moduleName]; found {
			t.Errorf("unsupported module %q is mounted in the module manager", moduleName)
		}
		if _, found := app.keys[moduleName]; found {
			t.Errorf("unsupported module %q has a store key", moduleName)
		}
		if _, found := ModuleBasics.DefaultGenesis(app.appCodec)[moduleName]; found {
			t.Errorf("unsupported module %q has default genesis", moduleName)
		}
	}

	queryRoutes := []string{
		"/cosmos.staking.v1beta1.Query/Validators",
		"/cosmos.staking.v1beta1.Query/Params",
		"/cosmos.distribution.v1beta1.Query/DelegationRewards",
		"/cosmos.distribution.v1beta1.Query/CommunityPool",
	}
	for _, route := range queryRoutes {
		if app.GRPCQueryRouter().Route(route) != nil {
			t.Errorf("unsupported gRPC query route %q is registered", route)
		}
	}

	messageRoutes := []string{
		"/cosmos.staking.v1beta1.MsgDelegate",
		"/cosmos.staking.v1beta1.MsgUndelegate",
		"/cosmos.staking.v1beta1.MsgBeginRedelegate",
		"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward",
		"/cosmos.distribution.v1beta1.MsgSetWithdrawAddress",
	}
	for _, route := range messageRoutes {
		if app.MsgServiceRouter().HandlerByTypeURL(route) != nil {
			t.Errorf("unsupported message route %q is registered", route)
		}
	}

	root := newRootCmd()
	for _, parentName := range []string{"query", "tx"} {
		parent, _, err := root.Find([]string{parentName})
		if err != nil || parent == nil || parent.Name() != parentName {
			t.Fatalf("find %s command: %v", parentName, err)
		}
		for _, child := range parent.Commands() {
			if child.Name() == stakingtypes.ModuleName || child.Name() == distrtypes.ModuleName {
				t.Errorf("unsupported CLI command %s %s is registered", parentName, child.Name())
			}
		}
	}
}

func isUnsupportedWasmRequest(err error) bool {
	var unsupported wasmvmtypes.UnsupportedRequest
	return errors.As(err, &unsupported) && unsupported.Kind == errUnsupportedStakingSurface.Error()
}
