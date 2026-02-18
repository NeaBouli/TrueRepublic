package dex

import (
	"encoding/json"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	abci "github.com/cometbft/cometbft/abci/types"
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
	for _, pool := range genesisState.Pools {
		am.keeper.CreatePool(ctx, pool.AssetDenom, pool.PnyxReserve, pool.AssetReserve)
	}
	return nil
}

func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	return nil
}
