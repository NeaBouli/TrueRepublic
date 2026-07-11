package truedemocracy

import (
	"context"
	"encoding/hex"
	"encoding/json"

	"cosmossdk.io/math"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	abci "github.com/cometbft/cometbft/abci/types"
	cryptoproto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

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
		&MsgDepositToDomain{},
		&MsgWithdrawFromDomain{},
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

func (AppModuleBasic) GetTxCmd() *cobra.Command    { return GetTxCmd() }
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

	// Restore domains from genesis (full state, not just name/admin/treasury).
	store := ctx.KVStore(am.keeper.StoreKey)
	for _, domain := range genesisState.Domains {
		bz := am.cdc.MustMarshalLengthPrefixed(&domain)
		store.Set([]byte("domain:"+domain.Name), bz)
		am.keeper.InitializeBigPurgeSchedule(ctx, domain.Name)
	}

	// Register genesis validators and build initial validator set.
	var updates []abci.ValidatorUpdate
	for _, gv := range genesisState.Validators {
		stake := sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, gv.Stake))
		if err := am.keeper.RegisterValidator(ctx, gv.OperatorAddr, gv.PubKey, stake, gv.Domain); err != nil {
			continue
		}
		power := gv.Stake / rewards.StakeMin
		pk := cryptoproto.PublicKey{
			Sum: &cryptoproto.PublicKey_Ed25519{Ed25519: gv.PubKey},
		}
		updates = append(updates, abci.ValidatorUpdate{PubKey: pk, Power: power})
	}

	// Load verifying key from genesis if present.
	if genesisState.VerifyingKeyHex != "" {
		vkBytes, err := hex.DecodeString(genesisState.VerifyingKeyHex)
		if err == nil {
			if _, err := DeserializeVerifyingKey(vkBytes); err == nil {
				am.keeper.SetVerifyingKey(ctx, vkBytes)
			}
		}
	}

	// Initialize PoD reward tracking state.
	timeBz := am.cdc.MustMarshalLengthPrefixed(ctx.BlockTime().Unix())
	store.Set([]byte("pod:last-reward-time"), timeBz)
	store.Set([]byte("dom:last-interest-time"), timeBz)
	return updates
}

// EndBlock implements module.HasABCIEndBlock. It distributes staking rewards
// and domain interest, disables ineligible validators without destroying
// their escrow claims, and returns CometBFT validator set updates.
func (am AppModule) EndBlock(goCtx context.Context) ([]abci.ValidatorUpdate, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Commit both inflation paths as one reward phase. If domain issuance fails
	// after staking issuance, neither the bank mint nor matching claims/timers
	// become visible to the rest of EndBlock.
	rewardCtx, writeRewards := ctx.CacheContext()
	if err := am.keeper.DistributeStakingRewards(rewardCtx); err != nil {
		return nil, err
	}

	if err := am.keeper.DistributeDomainInterest(rewardCtx); err != nil {
		return nil, err
	}
	writeRewards()

	// 2. Disable validators no longer in any domain. Their stake record stays
	// intact so custody remains withdrawable and escrow parity is preserved.
	var validatorAddresses []string
	am.keeper.IterateValidators(ctx, func(v Validator) bool {
		validatorAddresses = append(validatorAddresses, v.OperatorAddr)
		return false
	})
	for _, address := range validatorAddresses {
		if am.keeper.EnforceDomainMembership(ctx, address) {
			continue
		}
		validator, found := am.keeper.GetValidator(ctx, address)
		if !found {
			continue
		}
		validator.Jailed = true
		validator.Power = 0
		am.keeper.SetValidator(ctx, validator)
	}

	// 3. Disable under-staked validators while retaining their withdrawal claim.
	var underStaked []string
	am.keeper.IterateValidators(ctx, func(v Validator) bool {
		if v.Stake.AmountOf(PNYXDenom).LT(math.NewInt(rewards.StakeMin)) {
			underStaked = append(underStaked, v.OperatorAddr)
		}
		return false
	})
	for _, address := range underStaked {
		validator, found := am.keeper.GetValidator(ctx, address)
		if !found {
			continue
		}
		validator.Jailed = true
		validator.Power = 0
		am.keeper.SetValidator(ctx, validator)
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
	var domains []Domain
	am.keeper.IterateDomains(ctx, func(d Domain) bool {
		domains = append(domains, d)
		return false
	})
	if domains == nil {
		domains = []Domain{}
	}

	var validators []GenesisValidator
	am.keeper.IterateValidators(ctx, func(v Validator) bool {
		domain := ""
		if len(v.Domains) > 0 {
			domain = v.Domains[0]
		}
		validators = append(validators, GenesisValidator{
			OperatorAddr: v.OperatorAddr,
			PubKey:       v.PubKey,
			Stake:        v.Stake.AmountOf(PNYXDenom).Int64(),
			Domain:       domain,
		})
		return false
	})
	if validators == nil {
		validators = []GenesisValidator{}
	}

	vkHex := ""
	if vkBytes, found := am.keeper.GetVerifyingKey(ctx); found {
		vkHex = hex.EncodeToString(vkBytes)
	}

	genesis := GenesisState{
		Domains:         domains,
		Validators:      validators,
		VerifyingKeyHex: vkHex,
	}
	bz, _ := json.Marshal(genesis)
	return bz
}
