package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"cosmossdk.io/log"
	wasm "github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/spf13/cobra"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtcfg "github.com/cometbft/cometbft/config"
	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/privval"
	cryptoproto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/gogoproto/grpc"

	"github.com/cosmos/cosmos-sdk/client"
	clientconfig "github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	cmtservice "github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	sdkversion "github.com/cosmos/cosmos-sdk/version"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	crisis "github.com/cosmos/cosmos-sdk/x/crisis"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"truerepublic/token"
	"truerepublic/x/dex"
	"truerepublic/x/truedemocracy"
)

const envPrefix = "TRUEREPUBLIC"

var defaultNodeHome = filepath.Join(userHomeDir(), ".truerepublic")

func init() {
	sdkversion.Name = "TrueRepublic"
	sdkversion.AppName = "truerepublicd"
	sdkversion.Version = version
}

func userHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".truerepublic"
	}
	return home
}

// RegisterAPIRoutes implements servertypes.Application.
func (app *TrueRepublicApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig serverconfig.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	if err := server.RegisterSwaggerAPI(clientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
}

func (app *TrueRepublicApp) RegisterGRPCServer(grpcServer grpc.Server) {
	app.BaseApp.RegisterGRPCServer(grpcServer)
}

func (app *TrueRepublicApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.GRPCQueryRouter(), clientCtx, app.Simulate, app.appCodec.InterfaceRegistry())
}

func (app *TrueRepublicApp) RegisterTendermintService(clientCtx client.Context) {
	cmtApp := server.NewCometABCIWrapper(app)
	cmtservice.RegisterTendermintService(clientCtx, app.GRPCQueryRouter(), app.appCodec.InterfaceRegistry(), cmtApp.Query)
}

func (app *TrueRepublicApp) RegisterNodeService(clientCtx client.Context, cfg serverconfig.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer, appOpts servertypes.AppOptions) servertypes.Application {
	homeDir, _ := appOpts.Get(flags.FlagHome).(string)
	if homeDir == "" {
		homeDir = defaultNodeHome
	}
	baseAppOptions := server.DefaultBaseappOptions(appOpts)
	return NewTrueRepublicApp(logger, db, homeDir, baseAppOptions...)
}

func appExport(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	if forZeroHeight {
		return servertypes.ExportedApp{}, errors.New("zero-height export is not supported by the PoD validator model")
	}
	app, ok := newApp(logger, db, traceStore, appOpts).(*TrueRepublicApp)
	if !ok {
		return servertypes.ExportedApp{}, errors.New("unexpected application type")
	}
	if height >= 0 && height != app.LastBlockHeight() {
		if err := app.LoadVersion(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	}
	ctx := app.NewContextLegacy(true, cmtproto.Header{Height: app.LastBlockHeight()})
	genesis, err := app.mm.ExportGenesisForModules(ctx, app.appCodec, modulesToExport)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}
	appState, err := json.MarshalIndent(genesis, "", "  ")
	if err != nil {
		return servertypes.ExportedApp{}, err
	}
	return servertypes.ExportedApp{
		AppState:        appState,
		Height:          app.LastBlockHeight() + 1,
		ConsensusParams: app.GetConsensusParams(ctx),
	}, nil
}

func initAppConfig() (string, interface{}) {
	type appConfig struct {
		serverconfig.Config
		Wasm wasmtypes.WasmConfig `mapstructure:"wasm"`
	}
	cfg := serverconfig.DefaultConfig()
	cfg.MinGasPrices = "1000" + token.BaseDenom
	return serverconfig.DefaultConfigTemplate + wasmtypes.DefaultConfigTemplate(), appConfig{
		Config: *cfg,
		Wasm:   wasmtypes.DefaultWasmConfig(),
	}
}

func newRootCmd() *cobra.Command {
	legacyAmino := makeAminoCodec()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	ModuleBasics.RegisterInterfaces(interfaceRegistry)
	appCodec := codec.NewProtoCodec(interfaceRegistry)
	txConfig := authtx.NewTxConfig(appCodec, []signingtypes.SignMode{signingtypes.SignMode_SIGN_MODE_DIRECT})

	initClientCtx := client.Context{}.
		WithCodec(appCodec).
		WithInterfaceRegistry(interfaceRegistry).
		WithLegacyAmino(legacyAmino).
		WithTxConfig(txConfig).
		WithInput(os.Stdin).
		WithOutput(os.Stdout).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithHomeDir(defaultNodeHome).
		WithViper("")

	rootCmd := &cobra.Command{
		Use:           "truerepublicd",
		Short:         "TrueRepublic blockchain daemon and CLI",
		Version:       version,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())
			var err error
			initClientCtx = initClientCtx.WithCmdContext(cmd.Context())
			initClientCtx, err = client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			initClientCtx, err = clientconfig.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}
			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}
			template, config := initAppConfig()
			return server.InterceptConfigsPreRunHandler(cmd, template, config, cmtcfg.DefaultConfig())
		},
	}

	rootCmd.AddCommand(initNodeCmd(ModuleBasics, defaultNodeHome))
	server.AddCommands(rootCmd, defaultNodeHome, newApp, appExport, func(startCmd *cobra.Command) {
		crisis.AddModuleInitFlags(startCmd)
		wasm.AddModuleInitFlags(startCmd)
	})

	txCmd := &cobra.Command{Use: "tx", Short: "Transaction commands", RunE: client.ValidateCmd}
	txCmd.AddCommand(truedemocracy.GetTxCmd(), dex.GetTxCmd())
	queryCmd := &cobra.Command{Use: "query", Aliases: []string{"q"}, Short: "Query commands", RunE: client.ValidateCmd}
	queryCmd.AddCommand(truedemocracy.GetQueryCmd(legacyAmino), dex.GetQueryCmd(legacyAmino))
	rootCmd.AddCommand(txCmd, queryCmd, keys.Commands(), server.StatusCommand())
	return rootCmd
}

// initNodeCmd delegates file creation to the SDK, then builds both CometBFT
// and exactly bank-backed PoD genesis from this node's generated public key.
func initNodeCmd(basicManager module.BasicManager, home string) *cobra.Command {
	cmd := genutilcli.InitCmd(basicManager, home)
	defaultDenom := cmd.Flags().Lookup(genutilcli.FlagDefaultBondDenom)
	if defaultDenom != nil {
		defaultDenom.DefValue = token.BaseDenom
		_ = defaultDenom.Value.Set(token.BaseDenom)
	}
	initRun := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := initRun(cmd, args); err != nil {
			return err
		}
		serverCtx := server.GetServerContextFromCmd(cmd)
		clientCtx := client.GetClientContextFromCmd(cmd)
		serverCtx.Config.SetRoot(clientCtx.HomeDir)
		filePV := privval.LoadFilePV(serverCtx.Config.PrivValidatorKeyFile(), serverCtx.Config.PrivValidatorStateFile())
		pubKey, err := filePV.GetPubKey()
		if err != nil {
			return fmt.Errorf("read generated validator public key: %w", err)
		}
		return bindGenesisValidatorKey(serverCtx.Config.GenesisFile(), pubKey.Bytes())
	}
	return cmd
}

func bindGenesisValidatorKey(genesisPath string, pubKey []byte) error {
	if len(pubKey) != cmted25519.PubKeySize {
		return fmt.Errorf("generated validator public key must be %d bytes", cmted25519.PubKeySize)
	}
	genesis, err := genutiltypes.AppGenesisFromFile(genesisPath)
	if err != nil {
		return err
	}
	var appState map[string]json.RawMessage
	if err := json.Unmarshal(genesis.AppState, &appState); err != nil {
		return err
	}
	if genesis.Consensus == nil {
		return errors.New("consensus genesis is missing")
	}
	if len(genesis.Consensus.Validators) != 0 {
		return errors.New("refusing to replace an existing consensus validator set")
	}
	validatorUpdate := abci.ValidatorUpdate{
		PubKey: cryptoproto.PublicKey{Sum: &cryptoproto.PublicKey_Ed25519{Ed25519: append([]byte(nil), pubKey...)}},
		Power:  1,
	}
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	ModuleBasics.RegisterInterfaces(interfaceRegistry)
	appCodec := codec.NewProtoCodec(interfaceRegistry)
	if err := ensureConsensusGenesis(appCodec, appState, []abci.ValidatorUpdate{validatorUpdate}); err != nil {
		return fmt.Errorf("build bank-backed PoD genesis: %w", err)
	}
	var democracyGenesis truedemocracy.GenesisState
	if err := json.Unmarshal(appState[truedemocracy.ModuleName], &democracyGenesis); err != nil {
		return err
	}
	if len(democracyGenesis.Validators) != 1 || !bytes.Equal(democracyGenesis.Validators[0].PubKey, pubKey) {
		return errors.New("generated PoD validator does not match the node consensus key")
	}
	if err := validateLedgerGenesis(appCodec, appState); err != nil {
		return fmt.Errorf("generated PoD genesis is not bank-backed: %w", err)
	}
	genesis.AppState, err = json.Marshal(appState)
	if err != nil {
		return err
	}
	consensusKey := cmted25519.PubKey(append([]byte(nil), pubKey...))
	genesis.Consensus.Validators = []cmttypes.GenesisValidator{{
		Address: consensusKey.Address(),
		PubKey:  consensusKey,
		Power:   1,
		Name:    "truerepublic-bootstrap",
	}}
	if err := genesis.ValidateAndComplete(); err != nil {
		return err
	}
	updated, err := json.MarshalIndent(genesis, "", "  ")
	if err != nil {
		return err
	}
	return atomicWriteFile(genesisPath, updated, 0o600)
}

func atomicWriteFile(path string, data []byte, mode os.FileMode) (err error) {
	temp, err := os.CreateTemp(filepath.Dir(path), ".genesis-*.tmp")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer func() {
		_ = temp.Close()
		if err != nil {
			_ = os.Remove(tempPath)
		}
	}()
	if err = temp.Chmod(mode); err != nil {
		return err
	}
	if _, err = temp.Write(data); err != nil {
		return err
	}
	if err = temp.Sync(); err != nil {
		return err
	}
	if err = temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}

func executeRootCommand(rootCmd *cobra.Command) {
	if err := svrcmd.Execute(rootCmd, envPrefix, defaultNodeHome); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var _ servertypes.Application = (*TrueRepublicApp)(nil)
