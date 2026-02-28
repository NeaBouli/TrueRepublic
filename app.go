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
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	abci "github.com/cometbft/cometbft/abci/types"

	// IBC
	capability "github.com/cosmos/ibc-go/modules/capability"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	transfer "github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	transferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	"truerepublic/x/dex"
	"truerepublic/x/truedemocracy"
)

// maccPerms defines module account permissions for the auth keeper.
// Each entry maps a module account name to its allowed permissions.
var maccPerms = map[string][]string{
	authtypes.FeeCollectorName: nil,
	wasmtypes.ModuleName:       {authtypes.Burner},
	truedemocracy.ModuleName:   nil, // treasury bridge module account
	transfertypes.ModuleName:   {authtypes.Minter, authtypes.Burner},
}

var ModuleBasics = module.NewBasicManager(
	auth.AppModuleBasic{},
	bank.AppModuleBasic{},
	capability.AppModuleBasic{},
	ibc.AppModuleBasic{},
	transfer.AppModuleBasic{},
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
	tkeys         map[string]*storetypes.TransientStoreKey
	memKeys       map[string]*storetypes.MemoryStoreKey
	paramsKeeper  paramskeeper.Keeper
	accountKeeper authkeeper.AccountKeeper
	bankKeeper    bankkeeper.BaseKeeper
	capKeeper     *capabilitykeeper.Keeper
	ibcKeeper     *ibckeeper.Keeper
	transferKeeper transferkeeper.Keeper
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

	// Store keys: KV, transient, and memory stores.
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey,       // "acc"
		banktypes.StoreKey,       // "bank"
		paramstypes.StoreKey,     // "params"
		capabilitytypes.StoreKey, // "capability"
		ibcexported.StoreKey,     // "ibc"
		transfertypes.StoreKey,   // "transfer"
		wasmtypes.StoreKey,       // "wasm"
		truedemocracy.ModuleName,
		dex.ModuleName,
	)
	tkeys := storetypes.NewTransientStoreKeys(paramstypes.TStoreKey)
	memKeys := storetypes.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	app := &TrueRepublicApp{
		BaseApp:  baseapp.NewBaseApp("TrueRepublic", logger, db, txCfg.TxDecoder()),
		cdc:      cdc,
		appCodec: appCodec,
		keys:     keys,
		tkeys:    tkeys,
		memKeys:  memKeys,
	}

	// Set interface registry on the message service router so it can resolve
	// message type URLs during RegisterService.
	app.MsgServiceRouter().SetInterfaceRegistry(interfaceRegistry)

	// Authority address for governance operations (standard pattern).
	authority := authtypes.NewModuleAddress("gov").String()

	// --- Params keeper (legacy — required by IBC for parameter subspaces) ---
	app.paramsKeeper = paramskeeper.NewKeeper(
		appCodec, cdc, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey],
	)
	app.paramsKeeper.Subspace(ibcexported.ModuleName)
	app.paramsKeeper.Subspace(transfertypes.ModuleName)

	// --- Account keeper ---
	app.accountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		sdkaddress.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
		authority,
	)

	// --- Bank keeper ---
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

	// --- Capability keeper (IBC port/channel capability management) ---
	app.capKeeper = capabilitykeeper.NewKeeper(
		appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey],
	)
	scopedIBCKeeper := app.capKeeper.ScopeToModule(ibcexported.ModuleName)
	scopedTransferKeeper := app.capKeeper.ScopeToModule(transfertypes.ModuleName)
	scopedWasmKeeper := app.capKeeper.ScopeToModule(wasmtypes.ModuleName)
	app.capKeeper.Seal()

	// --- IBC keeper (core IBC: clients, connections, channels) ---
	ibcSubspace, _ := app.paramsKeeper.GetSubspace(ibcexported.ModuleName)
	app.ibcKeeper = ibckeeper.NewKeeper(
		appCodec,
		keys[ibcexported.StoreKey],
		ibcSubspace,
		IBCStakingKeeper{},
		IBCUpgradeKeeper{},
		scopedIBCKeeper,
		authority,
	)

	// --- Transfer keeper (ICS-20 token transfers) ---
	transferSubspace, _ := app.paramsKeeper.GetSubspace(transfertypes.ModuleName)
	app.transferKeeper = transferkeeper.NewKeeper(
		appCodec,
		keys[transfertypes.StoreKey],
		transferSubspace,
		app.ibcKeeper.ChannelKeeper,
		app.ibcKeeper.ChannelKeeper,
		app.ibcKeeper.PortKeeper,
		app.accountKeeper,
		app.bankKeeper,
		scopedTransferKeeper,
		authority,
	)

	// --- IBC Router (routes packets to IBC modules) ---
	ibcRouter := porttypes.NewRouter()
	transferIBCModule := transfer.NewIBCModule(app.transferKeeper)
	ibcRouter.AddRoute(transfertypes.ModuleName, transferIBCModule)
	app.ibcKeeper.SetRouter(ibcRouter)

	// --- Governance module keepers ---
	tdKeeper := truedemocracy.NewKeeper(cdc, keys[truedemocracy.ModuleName], truedemocracy.BuildTree(), app.bankKeeper)
	dexKeeper := dex.NewKeeper(cdc, keys[dex.ModuleName])
	app.tdKeeper = tdKeeper
	app.dexKeeper = dexKeeper

	// --- CosmWasm keeper (now using real IBC keepers instead of stubs) ---
	wasmConfig := wasmtypes.DefaultWasmConfig()
	app.wasmKeeper = wasmkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[wasmtypes.StoreKey]),
		app.accountKeeper,
		app.bankKeeper,
		StubStakingKeeper{},
		StubDistributionKeeper{},
		app.ibcKeeper.ChannelKeeper,
		app.ibcKeeper.ChannelKeeper,
		app.ibcKeeper.PortKeeper,
		scopedWasmKeeper,
		app.transferKeeper,
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

	// --- Module manager ---
	authModule := auth.NewAppModule(appCodec, app.accountKeeper, nil, nil)
	bankModule := bank.NewAppModule(appCodec, app.bankKeeper, app.accountKeeper, nil)
	capModule := capability.NewAppModule(appCodec, *app.capKeeper, false)
	ibcModule := ibc.NewAppModule(app.ibcKeeper)
	transferModule := transfer.NewAppModule(app.transferKeeper)
	wasmModule := wasm.NewAppModule(appCodec, &app.wasmKeeper, StubValidatorSetSource{}, app.accountKeeper, app.bankKeeper, app.MsgServiceRouter(), nil)

	app.mm = module.NewManager(
		authModule,
		bankModule,
		capModule,
		ibcModule,
		transferModule,
		wasmModule,
		app.tdModule,
		app.dexModule,
	)

	// Genesis order: capability first (port binding), then auth/bank, IBC, transfer, wasm, custom modules.
	app.mm.SetOrderInitGenesis(
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		ibcexported.ModuleName,
		transfertypes.ModuleName,
		wasmtypes.ModuleName,
		truedemocracy.ModuleName,
		dex.ModuleName,
	)

	// BeginBlock order: IBC client updates run first.
	app.mm.SetOrderBeginBlockers(
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		ibcexported.ModuleName,
		transfertypes.ModuleName,
		wasmtypes.ModuleName,
		truedemocracy.ModuleName,
		dex.ModuleName,
	)

	// EndBlock order: truedemocracy last (returns validator updates).
	app.mm.SetOrderEndBlockers(
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		ibcexported.ModuleName,
		transfertypes.ModuleName,
		wasmtypes.ModuleName,
		truedemocracy.ModuleName,
		dex.ModuleName,
	)

	// Register gRPC message handlers via module Configurator.
	configurator := module.NewConfigurator(appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(configurator)

	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

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

	// Fill in default genesis for modules not present in the genesis state.
	// This ensures IBC, capability, and transfer modules are properly
	// initialized even with minimal genesis (e.g., tests).
	defaults := ModuleBasics.DefaultGenesis(app.appCodec)
	for name, data := range defaults {
		if _, ok := genesisState[name]; !ok {
			genesisState[name] = data
		}
	}

	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// BeginBlocker runs begin-block logic (IBC client updates).
func (app *TrueRepublicApp) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.mm.BeginBlock(ctx)
}

// EndBlocker runs end-block logic (staking rewards, PoD enforcement, governance).
func (app *TrueRepublicApp) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.mm.EndBlock(ctx)
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
