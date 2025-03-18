package dex

import (
    "errors"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct {
    storeKey sdk.StoreKey
}

type Pool struct {
    Reserves map[string]sdk.Int // Asset -> Reserve
}

func NewKeeper(storeKey sdk.StoreKey) Keeper {
    return Keeper{storeKey: storeKey}
}

func (k Keeper) GetPool(ctx sdk.Context, fromAsset, toAsset string) Pool {
    store := ctx.KVStore(k.storeKey)
    bz := store.Get([]byte("pool:" + fromAsset + ":" + toAsset))
    if bz == nil {
        return Pool{Reserves: map[string]sdk.Int{
            "pnyx": sdk.NewInt(500000),
            "btc":  sdk.NewInt(500),
            "atom": sdk.NewInt(10000),
        }}
    }
    var pool Pool
    json.Unmarshal(bz, &pool)
    return pool
}

func (k Keeper) SetPool(ctx sdk.Context, pool Pool, fromAsset, toAsset string) {
    store := ctx.KVStore(k.storeKey)
    bz, _ := json.Marshal(pool)
    store.Set([]byte("pool:"+fromAsset+":"+toAsset), bz)
}

func (k Keeper) Swap(ctx sdk.Context, fromAsset, toAsset string, amount sdk.Int) (sdk.Coins, error) {
    if amount.LTE(sdk.ZeroInt()) {
        return nil, errors.New("Amount must be positive")
    }
    pool := k.GetPool(ctx, fromAsset, toAsset)
    kValue := pool.Reserves[fromAsset].Mul(pool.Reserves[toAsset])
    newFromRes := pool.Reserves[fromAsset].Add(amount)
    newToRes := kValue.Quo(newFromRes)
    out := pool.Reserves[toAsset].Sub(newToRes)
    pool.Reserves[fromAsset] = newFromRes
    pool.Reserves[toAsset] = newToRes
    k.SetPool(ctx, pool, fromAsset, toAsset)
    return sdk.NewCoins(sdk.NewCoin(toAsset, out)), nil
}
