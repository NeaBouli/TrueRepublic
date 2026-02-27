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

	wasm "github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkaddress "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	abci "github.com/cometbft/cometbft/abci/types"

	"truerepublic/x/dex"
	"truerepublic/x/truedemocracy"
)

// maccPerms defines module account permissions for the auth keeper.
// Each entry maps a module account name to its allowed permissions.
var maccPerms = map[string][]string{
	authtypes.FeeCollectorName: nil,
	wasmtypes.ModuleName:       {authtypes.Burner},
}

var ModuleBasics = module.NewBasicManager(
	auth.AppModuleBasic{},
	bank.AppModuleBasic{},
	wasm.AppModuleBasic{},
	truedemocracy.AppModuleBasic{},
	dex.AppModuleBasic{},
)

type TrueRepublicApp struct {
	*baseapp.BaseApp
	mm            *module.Manager
	cdc           *codec.LegacyAmino
	appCodec      codec.Codec
	keys          map[string]*storetypes.KVStoreKey
	accountKeeper authkeeper.AccountKeeper
	bankKeeper    bankkeeper.BaseKeeper
	wasmKeeper    wasmkeeper.Keeper
	tdKeeper      truedemocracy.Keeper
	dexKeeper     dex.Keeper
	tdModule      truedemocracy.AppModule
	dexModule     dex.AppModule
}

func NewTrueRepublicApp(logger log.Logger, db dbm.DB, homeDir string) *TrueRepublicApp {
	cdc := codec.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(cdc)
	legacytx.RegisterLegacyAminoCodec(cdc)
	authtypes.RegisterLegacyAminoCodec(cdc)
	banktypes.RegisterLegacyAminoCodec(cdc)
	truedemocracy.RegisterCodec(cdc)
	dex.RegisterCodec(cdc)

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	ModuleBasics.RegisterInterfaces(interfaceRegistry)
	appCodec := codec.NewProtoCodec(interfaceRegistry)

	txCfg := authtx.NewTxConfig(appCodec, []signingtypes.SignMode{signingtypes.SignMode_SIGN_MODE_DIRECT})
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey,      // "acc"
		banktypes.StoreKey,      // "bank"
		wasmtypes.StoreKey,      // "wasm"
		truedemocracy.ModuleName,
		dex.ModuleName,
	)

	app := &TrueRepublicApp{
		BaseApp:  baseapp.NewBaseApp("TrueRepublic", logger, db, txCfg.TxDecoder()),
		cdc:      cdc,
		appCodec: appCodec,
		keys:     keys,
	}

	// Set interface registry on the message service router so it can resolve
	// message type URLs during RegisterService.
	app.MsgServiceRouter().SetInterfaceRegistry(interfaceRegistry)

	// Authority address for governance operations (standard pattern).
	authority := authtypes.NewModuleAddress("gov").String()

	// Account keeper — manages on-chain accounts (addresses, sequences, pub keys).
	app.accountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		sdkaddress.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
		authority,
	)

	// Bank keeper — manages coin balances (PNYX token transfers, minting).
	// Domain.Treasury stays as internal accounting; x/bank handles user/contract balances.
	blockedAddrs := make(map[string]bool)
	for acc := range maccPerms {
		blockedAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}
	app.bankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		app.accountKeeper,
		blockedAddrs,
		authority,
		logger,
	)

	// Governance module keepers (created before wasm so custom bindings can reference them).
	tdKeeper := truedemocracy.NewKeeper(cdc, keys[truedemocracy.ModuleName], truedemocracy.BuildTree())
	dexKeeper := dex.NewKeeper(cdc, keys[dex.ModuleName])
	app.tdKeeper = tdKeeper
	app.dexKeeper = dexKeeper

	// CosmWasm keeper — executes WASM smart contracts.
	// Staking, distribution, and IBC keepers are stubbed (integrated in future milestones).
	// Custom query/message bindings let contracts read domain state and submit governance actions.
	wasmConfig := wasmtypes.DefaultWasmConfig()
	app.wasmKeeper = wasmkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[wasmtypes.StoreKey]),
		app.accountKeeper,
		app.bankKeeper,
		StubStakingKeeper{},
		StubDistributionKeeper{},
		StubICS4Wrapper{},
		StubChannelKeeper{},
		StubPortKeeper{},
		StubCapabilityKeeper{},
		StubICS20TransferPortSource{},
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		homeDir,
		wasmConfig,
		[]string{"iterator", "stargate", "cosmwasm_1_1", "cosmwasm_1_2", "cosmwasm_1_3", "cosmwasm_1_4", "cosmwasm_2_0"},
		authority,
		wasmkeeper.WithQueryPlugins(&wasmkeeper.QueryPlugins{
			Custom: truedemocracy.CustomQueryHandler(tdKeeper),
		}),
		wasmkeeper.WithMessageEncoders(&wasmkeeper.MessageEncoders{
			Custom: truedemocracy.CustomMessageEncoder(),
		}),
	)

	app.tdModule = truedemocracy.NewAppModule(cdc, tdKeeper)
	app.dexModule = dex.NewAppModule(cdc, dexKeeper)

	// Auth, bank, and wasm AppModules for genesis and gRPC service registration.
	authModule := auth.NewAppModule(appCodec, app.accountKeeper, nil, nil)
	bankModule := bank.NewAppModule(appCodec, app.bankKeeper, app.accountKeeper, nil)
	wasmModule := wasm.NewAppModule(appCodec, &app.wasmKeeper, StubValidatorSetSource{}, app.accountKeeper, app.bankKeeper, app.MsgServiceRouter(), nil)

	app.mm = module.NewManager(
		authModule,
		bankModule,
		wasmModule,
		app.tdModule,
		app.dexModule,
	)
	app.mm.SetOrderInitGenesis(
		authtypes.ModuleName,
		banktypes.ModuleName,
		wasmtypes.ModuleName,
		truedemocracy.ModuleName,
		dex.ModuleName,
	)

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

	// Initialize auth and bank with default params so account/balance
	// infrastructure is ready before governance module genesis.
	if err := app.accountKeeper.Params.Set(ctx, authtypes.DefaultParams()); err != nil {
		return nil, err
	}
	if err := app.bankKeeper.SetParams(ctx, banktypes.DefaultParams()); err != nil {
		return nil, err
	}

	// Initialize wasm params.
	if err := app.wasmKeeper.SetParams(ctx, wasmtypes.DefaultParams()); err != nil {
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
	authtypes.RegisterLegacyAminoCodec(cdc)
	banktypes.RegisterLegacyAminoCodec(cdc)
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
			homeDir := os.ExpandEnv("$HOME/.truerepublic")
			logger := log.NewLogger(os.Stdout)
			db := dbm.NewMemDB()
			app := NewTrueRepublicApp(logger, db, homeDir)
			_ = app
			logger.Info("TrueRepublic v0.1-alpha started")
			select {}
		},
	})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
