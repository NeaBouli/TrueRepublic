package main

import (
	"os"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"truerepublic/x/truedemocracy"
)

var ModuleBasics = module.NewBasicManager(
	truedemocracy.AppModuleBasic{},
)

type TrueRepublicApp struct {
	*baseapp.BaseApp
	mm   *module.Manager
	cdc  *codec.LegacyAmino
	keys map[string]*storetypes.KVStoreKey
}

func NewTrueRepublicApp(logger log.Logger, db dbm.DB) *TrueRepublicApp {
	cdc := codec.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(cdc)
	truedemocracy.RegisterCodec(cdc)

	keys := storetypes.NewKVStoreKeys(truedemocracy.ModuleName)

	app := &TrueRepublicApp{
		BaseApp: baseapp.NewBaseApp("TrueRepublic", logger, db, nil),
		cdc:     cdc,
		keys:    keys,
	}

	tdKeeper := truedemocracy.NewKeeper(cdc, keys[truedemocracy.ModuleName], truedemocracy.BuildTree())

	app.mm = module.NewManager(
		truedemocracy.NewAppModule(cdc, tdKeeper),
	)

	app.MountKVStores(keys)

	if err := app.LoadLatestVersion(); err != nil {
		panic(err)
	}

	return app
}

func main() {
	logger := log.NewLogger(os.Stdout)
	db := dbm.NewMemDB()
	app := NewTrueRepublicApp(logger, db)
	_ = app
	logger.Info("TrueRepublic v0.1-alpha started")
}
