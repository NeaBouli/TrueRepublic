package main

import (
    "os"
    "github.com/cosmos/cosmos-sdk/server"
    "github.com/cosmos/cosmos-sdk/types/module"
    "github.com/cosmos/cosmos-sdk/x/auth"
    "github.com/tendermint/tendermint/libs/log"
    tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
    abci "github.com/tendermint/tendermint/abci/types"
    "io"
    "github.com/cosmos/cosmos-sdk/baseapp"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "truerepublic/x/truedemocracy"
)

var ModuleBasics = module.NewBasicManager(
    auth.AppModuleBasic{},
    truedemocracy.AppModuleBasic{},
)

type TrueRepublicApp struct {
    *baseapp.BaseApp
    mm *module.Manager
}

func NewTrueRepublicApp(logger log.Logger, db dbm.DB, traceStore io.Writer) *TrueRepublicApp {
    app := &TrueRepublicApp{
        BaseApp: baseapp.NewBaseApp("TrueRepublic", logger, db, nil),
    }
    app.SetCommitMultiStoreTracer(traceStore)
    app.SetAppVersion("v0.1-alpha")

    app.mm = module.NewManager(
        auth.NewAppModule(app),
        truedemocracy.NewAppModule(app),
    )

    app.MountKVStores(ModuleBasics)
    app.SetInitChainer(app.InitChainer)
    app.SetBeginBlocker(app.BeginBlocker)
    app.SetEndBlocker(app.EndBlocker)

    return app
}

func (app *TrueRepublicApp) InitChainer(ctx sdk.Context, req tmproto.RequestInitChain) tmproto.ResponseInitChain {
    var genesisState truedemocracy.GenesisState
    genesisState = truedemocracy.DefaultGenesisState()
    app.mm.InitGenesis(ctx, genesisState)
    return tmproto.ResponseInitChain{}
}

func (app *TrueRepublicApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
    return app.mm.BeginBlock(ctx, req)
}

func (app *TrueRepublicApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
    return app.mm.EndBlock(ctx, req)
}

func main() {
    logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
    app := NewTrueRepublicApp(logger, dbm.NewMemDB(), nil)
    server.StartCmd(app, "0.0.0.0:26657")
}
