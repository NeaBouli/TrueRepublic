package treasury

import (
    "errors"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct {
    storeKey sdk.StoreKey
}

const (
    TotalSupply = 21000000 // 21M PNYX
)

func NewKeeper(storeKey sdk.StoreKey) Keeper {
    return Keeper{storeKey: storeKey}
}

func (k Keeper) Deposit(ctx sdk.Context, amount sdk.Coins) error {
    store := ctx.KVStore(k.storeKey)
    bz := store.Get([]byte("balance"))
    balance := sdk.NewCoins()
    if bz != nil {
        json.Unmarshal(bz, &balance)
    }
    total := balance.AmountOf("pnyx").Add(amount.AmountOf("pnyx"))
    if total.GT(sdk.NewInt(TotalSupply)) {
        return errors.New("Total supply exceeded")
    }
    balance = balance.Add(amount...) // PayToPut: Einfaches Deposit
    bz, _ = json.Marshal(balance)
    store.Set([]byte("balance"), bz)
    return nil
}

func (k Keeper) Withdraw(ctx sdk.Context, amount sdk.Coins) error {
    store := ctx.KVStore(k.storeKey)
    bz := store.Get([]byte("balance"))
    balance := sdk.NewCoins()
    if bz != nil {
        json.Unmarshal(bz, &balance)
    }
    if !balance.IsAllGTE(amount) {
        return errors.New("Insufficient funds")
    }
    balance = balance.Sub(amount)
    bz, _ = json.Marshal(balance)
    store.Set([]byte("balance"), bz)
    return nil
}

func (k Keeper) RateToEarn(ctx sdk.Context, rater string, amount sdk.Int) error {
    // RateToEarn: Belohnung für Bewertung
    store := ctx.KVStore(k.storeKey)
    bz := store.Get([]byte("balance"))
    balance := sdk.NewCoins()
    if bz != nil {
        json.Unmarshal(bz, &balance)
    }
    reward := sdk.NewCoin("pnyx", amount.Quo(sdk.NewInt(10))) // 10% von eingezahltem Betrag
    balance = balance.Add(reward)
    bz, _ = json.Marshal(balance)
    store.Set([]byte("balance"), bz)
    return nil
}

func (k Keeper) VoteToEarn(ctx sdk.Context, voter string, amount sdk.Int) error {
    // VoteToEarn: Belohnung für Abstimmung
    store := ctx.KVStore(k.storeKey)
    bz := store.Get([]byte("balance"))
    balance := sdk.NewCoins()
    if bz != nil {
        json.Unmarshal(bz, &balance)
    }
    reward := sdk.NewCoin("pnyx", amount.Quo(sdk.NewInt(20))) // 5% von eingezahltem Betrag
    balance = balance.Add(reward)
    bz, _ = json.Marshal(balance)
    store.Set([]byte("balance"), bz)
    return nil
}
