package dex

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/store/types"
    typeserrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type Keeper struct {
    storeKey types.StoreKey
    pnyxRes  sdk.Int
    btcRes   sdk.Int
}

func NewKeeper(storeKey types.StoreKey) Keeper {
    return Keeper{storeKey: storeKey, pnyxRes: sdk.NewInt(500000), btcRes: sdk.NewInt(500)}
}

func (k Keeper) AddLiquidity(ctx sdk.Context, provider sdk.AccAddress, pnyxAmt, btcAmt sdk.Int) (sdk.Coins, error) {
    if pnyxAmt.LTE(sdk.ZeroInt()) || btcAmt.LTE(sdk.ZeroInt()) {
        return nil, typeserrors.Wrap(typeserrors.ErrInvalidRequest, "Amounts must be positive")
    }
    poolTokens := sdk.MinInt(pnyxAmt, btcAmt.Mul(k.pnyxRes).Quo(k.btcRes))
    k.pnyxRes = k.pnyxRes.Add(pnyxAmt)
    k.btcRes = k.btcRes.Add(btcAmt)
    return sdk.NewCoins(sdk.NewCoin("pool", poolTokens)), nil
}

func (k Keeper) RemoveLiquidity(ctx sdk.Context, provider sdk.AccAddress, poolTokens sdk.Int) (sdk.Coins, error) {
    if poolTokens.LTE(sdk.ZeroInt()) {
        return nil, typeserrors.Wrap(typeserrors.ErrInvalidRequest, "Pool tokens must be positive")
    }
    pnyxShare := poolTokens.Mul(k.pnyxRes).Quo(sdk.NewInt(1000000))
    btcShare := poolTokens.Mul(k.btcRes).Quo(sdk.NewInt(1000000))
    k.pnyxRes = k.pnyxRes.Sub(pnyxShare)
    k.btcRes = k.btcRes.Sub(btcShare)
    return sdk.NewCoins(sdk.NewCoin("pnyx", pnyxShare), sdk.NewCoin("btc", btcShare)), nil
}

func (k Keeper) SwapBTCtoPNYX(ctx sdk.Context, btcAmt sdk.Int) (sdk.Coins, error) {
    if btcAmt.LTE(sdk.ZeroInt()) {
        return nil, typeserrors.Wrap(typeserrors.ErrInvalidRequest, "Amount must be positive")
    }
    if k.btcRes.LT(btcAmt) {
        return nil, typeserrors.Wrap(typeserrors.ErrInsufficientFunds, "Insufficient BTC reserve")
    }
    pnyxOut := k.pnyxRes.Mul(k.btcRes).Quo(k.btcRes.Add(btcAmt))
    fee := pnyxOut.Mul(sdk.NewInt(3)).Quo(sdk.NewInt(1000))
    burn := pnyxOut.Quo(sdk.NewInt(100))
    k.pnyxRes = k.pnyxRes.Sub(pnyxOut)
    k.btcRes = k.btcRes.Add(btcAmt)
    return sdk.NewCoins(sdk.NewCoin("pnyx", pnyxOut.Sub(fee).Sub(burn))), nil
}

func (k Keeper) SwapPNYXtoBTC(ctx sdk.Context, pnyxAmt sdk.Int) (sdk.Coins, error) {
    if pnyxAmt.LTE(sdk.ZeroInt()) {
        return nil, typeserrors.Wrap(typeserrors.ErrInvalidRequest, "Amount must be positive")
    }
    if k.pnyxRes.LT(pnyxAmt) {
        return nil, typeserrors.Wrap(typeserrors.ErrInsufficientFunds, "Insufficient PNYX reserve")
    }
    btcOut := k.btcRes.Mul(k.pnyxRes).Quo(k.pnyxRes.Add(pnyxAmt))
    fee := btcOut.Mul(sdk.NewInt(3)).Quo(sdk.NewInt(1000))
    k.pnyxRes = k.pnyxRes.Add(pnyxAmt)
    k.btcRes = k.btcRes.Sub(btcOut)
    return sdk.NewCoins(sdk.NewCoin("btc", btcOut.Sub(fee))), nil
}
