package truedemocracy

import (
    "time"
    "sort"
    "math"
    "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct {
    storeKey sdk.StoreKey
    nodes    []*Node
}

func NewKeeper(storeKey sdk.StoreKey, nodes []*Node) Keeper {
    return Keeper{storeKey: storeKey, nodes: nodes}
}

func (k Keeper) CreateDomain(ctx sdk.Context, name string, admin sdk.AccAddress, initialCoins sdk.Coins) {
    store := ctx.KVStore(k.storeKey)
    domain := Domain{
        Name:          name,
        Admin:         admin,
        Members:       []string{admin.String()},
        Treasury:      initialCoins,
        Issues:        []Issue{},
        Options:       DomainOptions{AdminElectable: true, AnyoneCanJoin: false},
        PermissionReg: []string{},
    }
    bz, _ := sdk.MarshalJSON(domain)
    store.Set([]byte("domain:"+name), bz)
}

func (k Keeper) SubmitProposal(ctx sdk.Context, domainName, issueName, suggestionName, creator string, fee sdk.Coins) error {
    store := ctx.KVStore(k.storeKey)
    domainBz := store.Get([]byte("domain:" + domainName))
    if domainBz == nil {
        return sdk.ErrUnknownRequest("Domain not found")
    }
    var domain Domain
    sdk.UnmarshalJSON(domainBz, &domain)

    if domain.Options.OnlyAdminIssues && creator != domain.Admin.String() {
        return sdk.ErrUnauthorized("Only admin can submit issues")
    }
    if domain.Options.CoinBurnRequired && fee.AmountOf("pnyx").LT(sdk.NewInt(100)) {
        return sdk.ErrInsufficientFunds("Coin burn of 100 PNYX required")
    }
    if fee.AmountOf("pnyx").LT(sdk.NewInt(15)) {
        return sdk.ErrInsufficientFunds("Minimum fee is 15 PNYX")
    }
    domain.Treasury = domain.Treasury.Add(fee...)

    found := false
    for i, issue := range domain.Issues {
        if issue.Name == issueName {
            domain.Issues[i].Suggestions = append(domain.Issues[i].Suggestions, Suggestion{
                Name:      suggestionName,
                C...
