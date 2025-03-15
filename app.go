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
    "truerepublic/x/dex"
)

var ModuleBasics = module.NewBasicManager(
    auth.AppModuleBasic{},
)

type TrueRepublicApp struct {
    *baseapp.BaseApp
    tdKeeper  truedemocracy.Keeper
    dexKeeper dex.Keeper
}

func NewTrueRepublicApp(logger log.Logger) *TrueRepublicApp {
    db := store.NewCommitMultiStore(nil)
    baseApp := baseapp.NewBaseApp("TrueRepublic", logger, db, nil)
    nodes := make([]*truedemocracy.Node, 7)
    for i := 0; i < 7; i++ {
        nodes[i] = &truedemocracy.Node{Address: "node" + string(rune(i+1)), PubKey: "pubkey" + string(rune(i+1)), Staked: sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100000))}
    }
    tdKeeper := truedemocracy.NewKeeper(baseApp.CommitMultiStore().GetKVStoreKey(), nodes)
    dexKeeper := dex.NewKeeper(baseApp.CommitMultiStore().GetKVStoreKey())
    return &TrueRepublicApp{
        BaseApp:   baseApp,
        tdKeeper:  tdKeeper,
        dexKeeper: dexKeeper,
    }
}

func (app *TrueRepublicApp) InitChain(req types.RequestInitChain) types.ResponseInitChain {
    ctx := sdk.NewContext(app.CommitMultiStore(), tmproto.Header{}, false, app.Logger())
    admin, _ := sdk.AccAddressFromBech32("cosmos1adminaddress")
    app.tdKeeper.CreateDomain(ctx, "governance", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500000)))
    return types.ResponseInitChain{}
}

func (app *TrueRepublicApp) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
    return types.ResponseBeginBlock{}
}

func (app *TrueRepublicApp) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
    ctx := sdk.NewContext(app.CommitMultiStore(), tmproto.Header{Height: req.Height}, false, app.Logger())
    store := ctx.KVStore(app.tdKeeper.storeKey)
    iterator := sdk.KVStorePrefixIterator(store, []byte("domain:"))
    defer iterator.Close()
    for ; iterator.Valid(); iterator.Next() {
        var domain truedemocracy.Domain
        json.Unmarshal(iterator.Value(), &domain)
        if len(domain.Issues) > 0 && time.Since(domain.Issues[len(domain.Issues)-1].Created) > 360*24*time.Hour {
            burn := domain.Treasury.AmountOf("pnyx").Quo(sdk.NewInt(2))
            domain.Treasury = domain.Treasury.Sub(sdk.NewCoins(sdk.NewCoin("pnyx", burn)))
        }
        if req.Height%129600 == 0 {
            domain.PermissionReg = []string{}
        }
        bz, _ := json.Marshal(domain)
        store.Set(iterator.Key(), bz)
    }
    for i, node := range app.tdKeeper.nodes {
        if time.Since(node.LastActive) > 90*24*time.Hour {
            slash := node.Staked.AmountOf("pnyx").Quo(sdk.NewInt(10))
            app.tdKeeper.nodes[i].Staked = node.Staked.Sub(sdk.NewCoins(sdk.NewCoin("pnyx", slash)))
        }
        reward := node.Staked.AmountOf("pnyx").Mul(sdk.NewInt(int64(truedemocracy.APY_node*100))).Quo(sdk.NewInt(365*100))
        app.tdKeeper.nodes[i].Staked = node.Staked.Add(sdk.NewCoins(sdk.NewCoin("pnyx", reward)))
    }
    return types.ResponseEndBlock{}
}

func (app *TrueRepublicApp) DeliverTx(req types.RequestDeliverTx) types.ResponseDeliverTx {
    ctx := sdk.NewContext(app.CommitMultiStore(), tmproto.Header{}, false, app.Logger())
    return types.ResponseDeliverTx{Code: 0}
}

func (app *TrueRepublicApp) CheckTx(req types.RequestCheckTx) types.ResponseCheckTx {
    return types.ResponseCheckTx{Code: 0}
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
