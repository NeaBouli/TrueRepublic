package main

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"

	wasm "github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkaddress "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensus "github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisis "github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	gogoproto "github.com/cosmos/gogoproto/proto"

	// IBC
	capability "github.com/cosmos/ibc-go/modules/capability"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	transfer "github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	transferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"

	"truerepublic/token"
	"truerepublic/x/dex"
	"truerepublic/x/truedemocracy"
)

// maccPerms defines module account permissions for the auth keeper.
// Each entry maps a module account name to its allowed permissions.
var maccPerms = map[string][]string{
	authtypes.FeeCollectorName: nil,
	wasmtypes.ModuleName:       {authtypes.Burner},
	truedemocracy.ModuleName:   {authtypes.Minter, authtypes.Burner}, // capped issuance, escrow, and slash burns
	dex.ModuleName:             {authtypes.Burner},                   // canonical swap burn
	transfertypes.ModuleName:   {authtypes.Minter, authtypes.Burner},
}

// version is injected by release and container builds via -ldflags.
var version = "dev"

var ModuleBasics = module.NewBasicManager(
	auth.AppModuleBasic{},
	bank.AppModuleBasic{},
	crisis.AppModuleBasic{},
	consensus.AppModuleBasic{},
	capability.AppModuleBasic{},
	ibc.AppModuleBasic{},
	transfer.AppModuleBasic{},
	wasm.AppModuleBasic{},
	truedemocracy.AppModuleBasic{},
	dex.AppModuleBasic{},
)

var sdkConfigOnce sync.Once

func configureSDKConfig() {
	sdkConfigOnce.Do(func() {
		config := sdk.GetConfig()
		config.SetBech32PrefixForAccount("truerepublic", "truerepublicpub")
		config.SetBech32PrefixForValidator("truerepublicvaloper", "truerepublicvaloperpub")
		config.SetBech32PrefixForConsensusNode("truerepublicvalcons", "truerepublicvalconspub")
		config.Seal()
	})
}

func makeInterfaceRegistry() codectypes.InterfaceRegistry {
	signingOptions, err := authtx.NewDefaultSigningOptions()
	if err != nil {
		panic(err)
	}
	truedemocracy.RegisterCustomGetSigners(signingOptions)

	interfaceRegistry, err := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles:     gogoproto.HybridResolver,
		SigningOptions: *signingOptions,
	})
	if err != nil {
		panic(err)
	}
	std.RegisterInterfaces(interfaceRegistry)
	ModuleBasics.RegisterInterfaces(interfaceRegistry)
	return interfaceRegistry
}

func makeTxConfig(appCodec codec.Codec) client.TxConfig {
	txConfig, err := authtx.NewTxConfigWithOptions(appCodec, authtx.ConfigOptions{
		EnabledSignModes: []signingtypes.SignMode{signingtypes.SignMode_SIGN_MODE_DIRECT},
		SigningContext:   appCodec.InterfaceRegistry().SigningContext(),
	})
	if err != nil {
		panic(err)
	}
	return txConfig
}

type TrueRepublicApp struct {
	*baseapp.BaseApp
	mm              *module.Manager
	cdc             *codec.LegacyAmino
	appCodec        codec.Codec
	txConfig        client.TxConfig
	keys            map[string]*storetypes.KVStoreKey
	tkeys           map[string]*storetypes.TransientStoreKey
	memKeys         map[string]*storetypes.MemoryStoreKey
	paramsKeeper    paramskeeper.Keeper
	accountKeeper   authkeeper.AccountKeeper
	bankKeeper      bankkeeper.BaseKeeper
	crisisKeeper    *crisiskeeper.Keeper
	consensusKeeper consensusparamkeeper.Keeper
	capKeeper       *capabilitykeeper.Keeper
	ibcKeeper       *ibckeeper.Keeper
	transferKeeper  transferkeeper.Keeper
	wasmKeeper      wasmkeeper.Keeper
	tdKeeper        truedemocracy.Keeper
	dexKeeper       dex.Keeper
	tdModule        truedemocracy.AppModule
	dexModule       dex.AppModule
}

func NewTrueRepublicApp(logger log.Logger, db dbm.DB, homeDir string, baseAppOptions ...func(*baseapp.BaseApp)) *TrueRepublicApp {
	configureSDKConfig()
	cdc := makeAminoCodec()

	interfaceRegistry := makeInterfaceRegistry()
	appCodec := codec.NewProtoCodec(interfaceRegistry)
	txCfg := makeTxConfig(appCodec)

	// Store keys: KV, transient, and memory stores.
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey,   // "acc"
		banktypes.StoreKey,   // "bank"
		crisistypes.StoreKey, // "crisis"
		consensusparamtypes.StoreKey,
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
		BaseApp:  baseapp.NewBaseApp("TrueRepublic", logger, db, txCfg.TxDecoder(), baseAppOptions...),
		cdc:      cdc,
		appCodec: appCodec,
		txConfig: txCfg,
		keys:     keys,
		tkeys:    tkeys,
		memKeys:  memKeys,
	}

	// Set the shared interface registry on BaseApp and its routers so tx
	// decoding, message routing, event generation, and gRPC all use the same
	// address codecs and custom signer resolvers.
	app.SetInterfaceRegistry(interfaceRegistry)

	// Authority address for governance operations (standard pattern).
	authority := authtypes.NewModuleAddress("gov").String()

	// --- Params keeper (legacy — required by IBC for parameter subspaces) ---
	app.paramsKeeper = paramskeeper.NewKeeper(
		appCodec, cdc, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey],
	)
	app.paramsKeeper.Subspace(ibcexported.ModuleName)
	app.paramsKeeper.Subspace(transfertypes.ModuleName)
	app.consensusKeeper = consensusparamkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[consensusparamtypes.StoreKey]),
		authority,
		runtime.EventService{},
	)
	app.SetParamStore(app.consensusKeeper.ParamsStore)

	// --- Account keeper ---
	accountAddressCodec := sdkaddress.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
	app.accountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		accountAddressCodec,
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
	app.crisisKeeper = crisiskeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[crisistypes.StoreKey]),
		1,
		app.bankKeeper,
		authtypes.FeeCollectorName,
		authority,
		accountAddressCodec,
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
		IBCStakingKeeper{initialized: true},
		IBCUpgradeKeeper{initialized: true},
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
	dexKeeper := dex.NewKeeper(cdc, keys[dex.ModuleName], app.bankKeeper, authority)
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
	crisisModule := crisis.NewAppModule(app.crisisKeeper, false, nil)
	consensusModule := consensus.NewAppModule(appCodec, app.consensusKeeper)
	capModule := capability.NewAppModule(appCodec, *app.capKeeper, false)
	ibcModule := ibc.NewAppModule(app.ibcKeeper)
	transferModule := transfer.NewAppModule(app.transferKeeper)
	wasmModule := wasm.NewAppModule(appCodec, &app.wasmKeeper, StubValidatorSetSource{}, app.accountKeeper, app.bankKeeper, app.MsgServiceRouter(), nil)

	app.mm = module.NewManager(
		authModule,
		bankModule,
		crisisModule,
		consensusModule,
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
		consensusparamtypes.ModuleName,
		ibcexported.ModuleName,
		transfertypes.ModuleName,
		wasmtypes.ModuleName,
		truedemocracy.ModuleName,
		dex.ModuleName,
		crisistypes.ModuleName,
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
	app.crisisKeeper.RegisterRoute("token", "supply-cap", token.SupplyCapInvariant(app.bankKeeper))
	app.mm.RegisterInvariants(app.crisisKeeper)

	// EndBlock order: custom state first, then crisis asserts every registered invariant.
	app.mm.SetOrderEndBlockers(
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		ibcexported.ModuleName,
		transfertypes.ModuleName,
		wasmtypes.ModuleName,
		truedemocracy.ModuleName,
		dex.ModuleName,
		crisistypes.ModuleName,
	)

	// Register gRPC message handlers via module Configurator.
	configurator := module.NewConfigurator(appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(configurator)

	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)
	anteHandler, err := authante.NewAnteHandler(authante.HandlerOptions{
		AccountKeeper:   app.accountKeeper,
		BankKeeper:      app.bankKeeper,
		SignModeHandler: txCfg.SignModeHandler(),
		SigGasConsumer:  authante.DefaultSigVerificationGasConsumer,
	})
	if err != nil {
		panic(err)
	}
	app.SetAnteHandler(anteHandler)

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
	if err := ensureConsensusGenesis(app.appCodec, genesisState, req.Validators); err != nil {
		return nil, err
	}

	bankGenesis := banktypes.GetGenesisStateFromAppState(app.appCodec, genesisState)
	if err := token.ValidateGenesisSupply(*bankGenesis); err != nil {
		return nil, err
	}
	token.EnsureMetadata(bankGenesis)
	bankGenesisJSON, err := app.appCodec.MarshalJSON(bankGenesis)
	if err != nil {
		return nil, err
	}
	genesisState[banktypes.ModuleName] = bankGenesisJSON
	if err := ModuleBasics.ValidateGenesis(app.appCodec, app.txConfig, genesisState); err != nil {
		return nil, err
	}
	if err := validateLedgerGenesis(app.appCodec, genesisState); err != nil {
		return nil, err
	}
	// The legacy x/params store backs IBC subspaces but has no module genesis
	// writer in this PoD app. Persist a version marker so an otherwise empty
	// IAVL store can be reopened at the committed root-store version.
	paramsStore := ctx.KVStore(app.keys[paramstypes.StoreKey])
	paramsStore.Set([]byte("truerepublic:params-store-version"), []byte{1})

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
	std.RegisterLegacyAminoCodec(cdc)
	// auth registers legacytx.StdTx itself. Registering legacytx directly before
	// auth panics during every CLI/node startup because Amino rejects duplicate
	// concrete type registrations.
	authtypes.RegisterLegacyAminoCodec(cdc)
	banktypes.RegisterLegacyAminoCodec(cdc)
	crisistypes.RegisterLegacyAminoCodec(cdc)
	truedemocracy.RegisterCodec(cdc)
	dex.RegisterCodec(cdc)
	return cdc
}

func main() {
	executeRootCommand(newRootCmd())
}
