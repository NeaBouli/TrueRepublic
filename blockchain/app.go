package main

import (
    "os"
    "github.com/cosmos/cosmos-sdk/server"
    "github.com/cosmos/cosmos-sdk/types/module"
    "github.com/cosmos/cosmos-sdk/x/auth"
    "github.com/tendermint/tendermint/libs/log"
    tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
    "github.com/cosmos/cosmos-sdk/baseapp"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/store"
    "truerepublic/x/truedemocracy"
    "truerepublic/x/dex"
    "truerepublic/x/ibc"
    "truerepublic/x/treasury"
)

var ModuleBasics = module.NewBasicManager(auth.AppModuleBasic{})

type TrueRepublicApp struct {
    *baseapp.BaseApp
    tdKeeper       truedemocracy.Keeper
    dexKeeper      dex.Keeper
    ibcHandler     ibc.Handler
    treasuryKeeper treasury.Keeper
}

func NewTrueRepublicApp(logger log.Logger) *TrueRepublicApp {
    db := store.NewCommitMultiStore(nil)
    baseApp := baseapp.NewBaseApp("TrueRepublic", logger, db, nil)
    tdKeeper := truedemocracy.NewKeeper(baseApp.CommitMultiStore().GetKVStoreKey())
    dexKeeper := dex.NewKeeper(baseApp.CommitMultiStore().GetKVStoreKey())
    ibcHandler := ibc.NewHandler(baseApp.CommitMultiStore().GetKVStoreKey())
    treasuryKeeper := treasury.NewKeeper(baseApp.CommitMultiStore().GetKVStoreKey())
    app := &TrueRepublicApp{
        BaseApp:        baseApp,
        tdKeeper:       tdKeeper,
        dexKeeper:      dexKeeper,
        ibcHandler:     ibcHandler,
        treasuryKeeper: treasuryKeeper,
    }
    app.SetInitChainer(app.InitChainer)
    return app
}

func (app *TrueRepublicApp) InitChainer(ctx sdk.Context, req tmproto.InitChainRequest) tmproto.InitChainResponse {
    return tmproto.InitChainResponse{}
}

func main() {
    logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
    app := NewTrueRepublicApp(logger)
    server := server.NewServer(app, "TrueRepublic", "home")
    if err := server.Start(); err != nil {
        logger.Error("Failed to start server", "error", err)
        os.Exit(1)
    }
    defer server.Stop()
}
