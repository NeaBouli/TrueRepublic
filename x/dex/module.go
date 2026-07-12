package dex

import (
	"encoding/json"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ module.AppModule      = AppModule{}
)

// AppModuleBasic

type AppModuleBasic struct{}

func (AppModuleBasic) Name() string { return ModuleName }

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	RegisterCodec(cdc)
}

func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreatePool{},
		&MsgSwap{},
		&MsgAddLiquidity{},
		&MsgRemoveLiquidity{},
		&MsgRegisterAsset{},
		&MsgUpdateAssetStatus{},
		&MsgSwapExact{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	genesis := DefaultGenesisState()
	bz, err := json.Marshal(genesis)
	if err != nil {
		panic(err)
	}
	return bz
}

func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genesis GenesisState
	if err := json.Unmarshal(bz, &genesis); err != nil {
		return err
	}
	return ValidateGenesisState(genesis)
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

func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	ir.RegisterRoute(ModuleName, "reserve-custody", func(ctx sdk.Context) (string, bool) {
		if err := am.keeper.ValidateReserveCustody(ctx); err != nil {
			return err.Error(), true
		}
		return "DEX bank balances match pool reserve claims", false
	})
	ir.RegisterRoute(ModuleName, "lp-conservation", func(ctx sdk.Context) (string, bool) {
		if err := am.keeper.ValidateLPConservation(ctx); err != nil {
			return err.Error(), true
		}
		return "DEX provider LP shares match pool totals", false
	})
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	RegisterMsgServer(cfg.MsgServer(), NewMsgServer(am.keeper))
	RegisterQueryServer(cfg.QueryServer(), am.keeper)
}

func (am AppModule) ConsensusVersion() uint64 { return 1 }

func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState GenesisState
	if err := json.Unmarshal(data, &genesisState); err != nil {
		panic(err)
	}
	for _, asset := range genesisState.RegisteredAssets {
		if err := am.keeper.RegisterAsset(ctx, asset); err != nil {
			panic(err)
		}
	}
	for _, pool := range genesisState.Pools {
		am.keeper.SetPool(ctx, pool)
	}
	for _, position := range genesisState.LPPositions {
		provider, err := sdk.AccAddressFromBech32(position.Provider)
		if err != nil {
			panic(err)
		}
		am.keeper.setLPBalance(ctx, position.AssetDenom, provider, position.Shares)
	}
	if err := am.keeper.validateCustodyAndShares(ctx); err != nil {
		panic(err)
	}
	return nil
}

func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	var pools []Pool
	am.keeper.IteratePools(ctx, func(p Pool) bool {
		pools = append(pools, p)
		return false
	})
	genesis := GenesisState{
		Pools:            pools,
		RegisteredAssets: am.keeper.GetAllAssets(ctx),
		LPPositions:      am.keeper.GetAllLPPositions(ctx),
	}
	bz, err := json.Marshal(genesis)
	if err != nil {
		panic(err)
	}
	return bz
}
