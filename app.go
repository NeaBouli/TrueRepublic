package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	abci "github.com/cometbft/cometbft/abci/types"

	"truerepublic/x/dex"
	"truerepublic/x/truedemocracy"
)

var ModuleBasics = module.NewBasicManager(
	truedemocracy.AppModuleBasic{},
	dex.AppModuleBasic{},
)

type TrueRepublicApp struct {
	*baseapp.BaseApp
	mm        *module.Manager
	cdc       *codec.LegacyAmino
	appCodec  codec.Codec
	keys      map[string]*storetypes.KVStoreKey
	tdKeeper  truedemocracy.Keeper
	dexKeeper dex.Keeper
	tdModule  truedemocracy.AppModule
	dexModule dex.AppModule
}

func NewTrueRepublicApp(logger log.Logger, db dbm.DB) *TrueRepublicApp {
	cdc := codec.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(cdc)
	legacytx.RegisterLegacyAminoCodec(cdc)
	truedemocracy.RegisterCodec(cdc)
	dex.RegisterCodec(cdc)

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	ModuleBasics.RegisterInterfaces(interfaceRegistry)
	appCodec := codec.NewProtoCodec(interfaceRegistry)

	txCfg := authtx.NewTxConfig(appCodec, []signingtypes.SignMode{signingtypes.SignMode_SIGN_MODE_DIRECT})
	keys := storetypes.NewKVStoreKeys(truedemocracy.ModuleName, dex.ModuleName)

	app := &TrueRepublicApp{
		BaseApp:  baseapp.NewBaseApp("TrueRepublic", logger, db, txCfg.TxDecoder()),
		cdc:      cdc,
		appCodec: appCodec,
		keys:     keys,
	}

	// Set interface registry on the message service router so it can resolve
	// message type URLs during RegisterService.
	app.MsgServiceRouter().SetInterfaceRegistry(interfaceRegistry)

	tdKeeper := truedemocracy.NewKeeper(cdc, keys[truedemocracy.ModuleName], truedemocracy.BuildTree())
	dexKeeper := dex.NewKeeper(cdc, keys[dex.ModuleName])
	app.tdKeeper = tdKeeper
	app.dexKeeper = dexKeeper

	app.tdModule = truedemocracy.NewAppModule(cdc, tdKeeper)
	app.dexModule = dex.NewAppModule(cdc, dexKeeper)

	app.mm = module.NewManager(app.tdModule, app.dexModule)
	app.mm.SetOrderInitGenesis(truedemocracy.ModuleName, dex.ModuleName)

	// Register gRPC message handlers via module Configurator.
	configurator := module.NewConfigurator(appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(configurator)

	app.SetInitChainer(app.InitChainer)
	app.SetEndBlocker(app.EndBlocker)

	app.MountKVStores(keys)

	if err := app.LoadLatestVersion(); err != nil {
		panic(err)
	}

	return app
}

// Query intercepts ABCI queries to support legacy "custom/" paths that were
// removed in Cosmos SDK v0.50. All other queries fall through to BaseApp.
func (app *TrueRepublicApp) Query(ctx context.Context, req *abci.RequestQuery) (*abci.ResponseQuery, error) {
	path := strings.Split(strings.TrimPrefix(req.Path, "/"), "/")
	if len(path) >= 2 && path[0] == "custom" {
		sdkCtx, err := app.CreateQueryContext(req.Height, false)
		if err != nil {
			return nil, err
		}

		switch path[1] {
		case truedemocracy.ModuleName:
			querier := truedemocracy.NewQuerier(app.tdKeeper, app.cdc)
			bz, err := querier(sdkCtx, path[2:], *req)
			if err != nil {
				return nil, err
			}
			return &abci.ResponseQuery{Value: bz}, nil
		case dex.ModuleName:
			querier := dex.NewQuerier(app.dexKeeper, app.cdc)
			bz, err := querier(sdkCtx, path[2:], *req)
			if err != nil {
				return nil, err
			}
			return &abci.ResponseQuery{Value: bz}, nil
		}
	}
	return app.BaseApp.Query(ctx, req)
}

// InitChainer initializes the chain from genesis state.
func (app *TrueRepublicApp) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState map[string]json.RawMessage
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		return nil, err
	}

	var validatorUpdates []abci.ValidatorUpdate

	if data, ok := genesisState[truedemocracy.ModuleName]; ok {
		updates := app.tdModule.InitGenesis(ctx, app.appCodec, data)
		if len(updates) > 0 {
			validatorUpdates = updates
		}
	}
	if data, ok := genesisState[dex.ModuleName]; ok {
		app.dexModule.InitGenesis(ctx, app.appCodec, data)
	}

	return &abci.ResponseInitChain{Validators: validatorUpdates}, nil
}

// EndBlocker runs end-block logic (staking rewards, PoD enforcement, governance).
func (app *TrueRepublicApp) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	updates, err := app.tdModule.EndBlock(ctx)
	if err != nil {
		return sdk.EndBlock{}, err
	}
	return sdk.EndBlock{ValidatorUpdates: updates}, nil
}

// makeAminoCodec creates and configures the amino codec for CLI operations.
func makeAminoCodec() *codec.LegacyAmino {
	cdc := codec.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(cdc)
	legacytx.RegisterLegacyAminoCodec(cdc)
	truedemocracy.RegisterCodec(cdc)
	dex.RegisterCodec(cdc)
	return cdc
}

func main() {
	cdc := makeAminoCodec()

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	ModuleBasics.RegisterInterfaces(interfaceRegistry)
	appCodec := codec.NewProtoCodec(interfaceRegistry)

	txCfg := authtx.NewTxConfig(appCodec, []signingtypes.SignMode{signingtypes.SignMode_SIGN_MODE_DIRECT})

	initClientCtx := client.Context{}.
		WithCodec(appCodec).
		WithLegacyAmino(cdc).
		WithTxConfig(txCfg)

	rootCmd := &cobra.Command{
		Use:   "truerepublicd",
		Short: "TrueRepublic blockchain daemon and CLI",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			ctx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			return client.SetCmdClientContext(cmd, ctx)
		},
	}

	rootCmd.PersistentFlags().String(flags.FlagChainID, "truerepublic-1", "Chain ID")
	rootCmd.PersistentFlags().String(flags.FlagNode, "tcp://localhost:26657", "RPC endpoint")
	rootCmd.PersistentFlags().String(flags.FlagKeyringBackend, "test", "Keyring backend")
	rootCmd.PersistentFlags().String(flags.FlagHome, os.ExpandEnv("$HOME/.truerepublic"), "Home directory")
	rootCmd.PersistentFlags().String(flags.FlagOutput, "text", "Output format (text|json)")

	// Transaction commands.
	txCmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transaction commands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(truedemocracy.GetTxCmd())
	txCmd.AddCommand(dex.GetTxCmd())
	rootCmd.AddCommand(txCmd)

	// Query commands.
	queryCmd := &cobra.Command{
		Use:                        "query",
		Short:                      "Query commands",
		Aliases:                    []string{"q"},
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	queryCmd.AddCommand(truedemocracy.GetQueryCmd(cdc))
	queryCmd.AddCommand(dex.GetQueryCmd(cdc))
	rootCmd.AddCommand(queryCmd)

	// Start command.
	rootCmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Start the TrueRepublic node",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := log.NewLogger(os.Stdout)
			db := dbm.NewMemDB()
			app := NewTrueRepublicApp(logger, db)
			_ = app
			logger.Info("TrueRepublic v0.1-alpha started")
			select {}
		},
	})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
