package truedemocracy

import (
	"context"
	"encoding/json"

	"cosmossdk.io/math"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	abci "github.com/cometbft/cometbft/abci/types"
	cryptoproto "github.com/cometbft/cometbft/proto/tendermint/crypto"

	rewards "truerepublic/treasury/keeper"
)

var (
	_ module.AppModuleBasic  = AppModuleBasic{}
	_ module.AppModule       = AppModule{}
	_ module.HasABCIEndBlock = AppModule{}
)

// AppModuleBasic

type AppModuleBasic struct{}

func (AppModuleBasic) Name() string { return ModuleName }

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	RegisterCodec(cdc)
}

func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateDomain{},
		&MsgSubmitProposal{},
		&MsgRegisterValidator{},
		&MsgWithdrawStake{},
		&MsgRemoveValidator{},
		&MsgUnjail{},
		&MsgJoinPermissionRegister{},
		&MsgPurgePermissionRegister{},
		&MsgPlaceStoneOnIssue{},
		&MsgPlaceStoneOnSuggestion{},
		&MsgPlaceStoneOnMember{},
		&MsgVoteToExclude{},
		&MsgVoteToDelete{},
		&MsgRateProposal{},
		&MsgCastElectionVote{},
		&MsgAddMember{},
		&MsgOnboardToDomain{},
		&MsgApproveOnboarding{},
		&MsgRejectOnboarding{},
		&MsgRegisterIdentity{},
		&MsgRateWithProof{},
	)
}

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	genesis := DefaultGenesisState()
	bz, _ := json.Marshal(genesis)
	return bz
}

func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	return nil
}

func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {}

func (AppModuleBasic) GetTxCmd() *cobra.Command   { return GetTxCmd() }
func (AppModuleBasic) GetQueryCmd() *cobra.Command { return GetQueryCmd(codec.NewLegacyAmino()) }

// AppModule

type AppModule struct {
	AppModuleBasic
	keeper Keeper
	cdc    *codec.LegacyAmino
}

func NewAppModule(cdc *codec.LegacyAmino, keeper Keeper) AppModule {
	return AppModule{keeper: keeper, cdc: cdc}
}

func (AppModule) IsOnePerModuleType() {}
func (AppModule) IsAppModule()        {}

func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	RegisterMsgServer(cfg.MsgServer(), NewMsgServer(am.keeper))
	RegisterQueryServer(cfg.QueryServer(), am.keeper)
}

func (am AppModule) ConsensusVersion() uint64 { return 1 }

func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState GenesisState
	if err := json.Unmarshal(data, &genesisState); err != nil {
		return nil
	}

	// Create domains.
	for _, domain := range genesisState.Domains {
		am.keeper.CreateDomain(ctx, domain.Name, domain.Admin, domain.Treasury)
	}

	// Register genesis validators and build initial validator set.
	var updates []abci.ValidatorUpdate
	for _, gv := range genesisState.Validators {
		stake := sdk.NewCoins(sdk.NewInt64Coin("pnyx", gv.Stake))
		if err := am.keeper.RegisterValidator(ctx, gv.OperatorAddr, gv.PubKey, stake, gv.Domain); err != nil {
			continue
		}
		power := gv.Stake / rewards.StakeMin
		pk := cryptoproto.PublicKey{
			Sum: &cryptoproto.PublicKey_Ed25519{Ed25519: gv.PubKey},
		}
		updates = append(updates, abci.ValidatorUpdate{PubKey: pk, Power: power})
	}

	// Initialize PoD reward tracking state.
	store := ctx.KVStore(am.keeper.StoreKey)
	timeBz := am.cdc.MustMarshalLengthPrefixed(ctx.BlockTime().Unix())
	store.Set([]byte("pod:last-reward-time"), timeBz)
	store.Set([]byte("dom:last-interest-time"), timeBz)
	zeroInt := math.ZeroInt()
	releaseBz := am.cdc.MustMarshalLengthPrefixed(&zeroInt)
	store.Set([]byte("pod:total-release"), releaseBz)

	return updates
}

// EndBlock implements module.HasABCIEndBlock. It distributes staking rewards
// and domain interest, enforces domain membership and minimum stake, and
// returns CometBFT validator set updates.
func (am AppModule) EndBlock(goCtx context.Context) ([]abci.ValidatorUpdate, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// 1. Distribute staking rewards if the interval has elapsed.
	if err := am.keeper.DistributeStakingRewards(ctx); err != nil {
		return nil, err
	}

	// 1b. Distribute domain interest (eq.4) alongside staking rewards.
	if err := am.keeper.DistributeDomainInterest(ctx); err != nil {
		return nil, err
	}

	// 2. Enforce domain membership — remove validators no longer in any domain.
	var toRemove []string
	am.keeper.IterateValidators(ctx, func(v Validator) bool {
		if !am.keeper.EnforceDomainMembership(ctx, v.OperatorAddr) {
			toRemove = append(toRemove, v.OperatorAddr)
		}
		return false
	})
	for _, addr := range toRemove {
		am.keeper.RemoveValidator(ctx, addr)
	}

	// 3. Enforce minimum stake.
	var underStaked []string
	am.keeper.IterateValidators(ctx, func(v Validator) bool {
		if v.Stake.AmountOf("pnyx").LT(math.NewInt(rewards.StakeMin)) {
			underStaked = append(underStaked, v.OperatorAddr)
		}
		return false
	})
	for _, addr := range underStaked {
		am.keeper.RemoveValidator(ctx, addr)
	}

	// 4. Evaluate suggestion lifecycle zones (green/yellow/red → auto-delete).
	am.keeper.ProcessAllLifecycles(ctx)

	// 5. Governance: admin election and inactivity cleanup.
	am.keeper.ProcessGovernance(ctx)

	// 6. Check and execute Big Purges (WP S4: periodic permission register cleanup).
	am.keeper.CheckAndExecuteBigPurges(ctx)

	// 7. Build and return validator updates.
	updates := am.keeper.BuildValidatorUpdates(ctx)
	return updates, nil
}

func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	return nil
}
