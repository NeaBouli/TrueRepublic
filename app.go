package main

import (
    "os"
    "github.com/cosmos/cosmos-sdk/server"
    "github.com/cosmos/cosmos-sdk/types/module"
    "github.com/cosmos/cosmos-sdk/x/auth"
    "github.com/tendermint/tendermint/libs/log"
    tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
    "github.com/tendermint/tendermint/abci/types"
    "github.com/cosmos/cosmos-sdk/baseapp"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/store"
    "truerepublic/x/truedemocracy"
)

var ModuleBasics = module.NewBasicManager(
    auth.AppModuleBasic{},
)

type TrueRepublicApp struct {
    *baseapp.BaseApp
    keeper truedemocracy.Keeper
}

func NewTrueRepublicApp(logger log.Logger) *TrueRepublicApp {
    db := store.NewCommitMultiStore(nil) // In-memory store für Tests
    baseApp := baseapp.NewBaseApp("TrueRepublic", logger, db, nil)
    nodes := []*truedemocracy.Node{{Address: "node1", PubKey: "pubkey1"}}
    keeper := truedemocracy.NewKeeper(baseApp.CommitMultiStore().GetKVStoreKey(), nodes)
    return &TrueRepublicApp{
        BaseApp: baseApp,
        keeper:  keeper,
    }
}

func (app *TrueRepublicApp) InitChain(req types.RequestInitChain) types.ResponseInitChain {
    return types.ResponseInitChain{}
}

func (app *TrueRepublicApp) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
    return types.ResponseBeginBlock{}
}

func (app *TrueRepublicApp) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
    return types.ResponseEndBlock{}
}

func (app *TrueRepublicApp) DeliverTx(req types.RequestDeliverTx) types.ResponseDeliverTx {
    return types.ResponseDeliverTx{Code: 0}
}

func (app *TrueRepublicApp) CheckTx(req types.RequestCheckTx) types.ResponseCheckTx {
    return types.ResponseCheckTx{Code: 0}
}

func main() {
    logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
    app := NewTrueRepublicApp(logger)
    server := server.NewServer(app, "TrueRepublic", "home")
    ctx := sdk.NewContext(app.CommitMultiStore(), tmproto.Header{}, false, logger)

    // Beispiel: Domain erstellen
    admin, _ := sdk.AccAddressFromBech32("cosmos1adminaddress")
    initialCoins := sdk.NewCoins(sdk.NewInt64Coin("pnyx", 1000))
    app.keeper.CreateDomain(ctx, "testdomain", admin, initialCoins)

    if err := server.Start(); err != nil {
        logger.Error("Failed to start server", "error", err)
        os.Exit(1)
    }
    defer server.Stop()
}
