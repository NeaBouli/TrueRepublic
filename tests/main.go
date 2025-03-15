package main

import (
    "fmt"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
    "time"
    "truerepublic/x/truedemocracy"
    "truerepublic/x/dex"
    tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func main() {
    admin, _ := sdk.AccAddressFromBech32("cosmos1adminaddress")
    ctx := sdk.NewContext(nil, tmproto.Header{}, false, nil)
    nodes := []*truedemocracy.Node{{Address: "node1", PubKey: "pubkey1", Staked: sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100000))}}
    tdKeeper := truedemocracy.NewKeeper(nil, nodes)
    dexKeeper := dex.NewKeeper(nil)

    err := tdKeeper.CreateDomain(ctx, "testdomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 2000)))
    if err != nil {
        fmt.Println("CreateDomain failed:", err)
        return
    }
    fmt.Println("CreateDomain succeeded")

    err = tdKeeper.SubmitProposal(ctx, "testdomain", "testissue", "testsuggestion", admin.String(), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 15)))
    if err != nil {
        fmt.Println("SubmitProposal failed:", err)
        return
    }
    fmt.Println("SubmitProposal succeeded")

    privKey := ed25519.GenPrivKey()
    reward, cache, err := tdKeeper.RateProposal(ctx, "testdomain", "testissue", "testsuggestion", "voter1", 5, privKey)
    if err != nil {
        fmt.Println("RateProposal failed:", err)
        return
    }
    fmt.Println("RateProposal succeeded - Reward:", reward, "Cache:", cache)

    stoneReward, err := tdKeeper.AddStones(ctx, "testdomain", "testissue", "testsuggestion", "voter1")
    if err != nil {
        fmt.Println("AddStones failed:", err)
        return
    }
    fmt.Println("AddStones succeeded - Reward:", stoneReward)

    err = tdKeeper.FinalizeIssue(ctx, "testdomain", "testissue")
    if err != nil {
        fmt.Println("FinalizeIssue failed:", err)
        return
    }
    fmt.Println("FinalizeIssue succeeded")

    liquidity, err := dexKeeper.AddLiquidity(ctx, admin, sdk.NewInt(1000), sdk.NewInt(1))
    if err != nil {
        fmt.Println("AddLiquidity failed:", err)
        return
    }
    fmt.Println("AddLiquidity succeeded - Pool Tokens:", liquidity)
    swap, err := dexKeeper.SwapBTCtoPNYX(ctx, sdk.NewInt(1))
    if err != nil {
        fmt.Println("SwapBTCtoPNYX failed:", err)
        return
    }
    fmt.Println("SwapBTCtoPNYX succeeded - PNYX:", swap)
}
