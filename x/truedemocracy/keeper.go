package truedemocracy

import (
    "encoding/json"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
    "github.com/cosmos/cosmos-sdk/store/types"
    typeserrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type Keeper struct {
    storeKey types.StoreKey
    nodes    []*Node
}

func NewKeeper(storeKey types.StoreKey, nodes []*Node) Keeper {
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
    bz, _ := json.Marshal(domain)
    store.Set([]byte("domain:"+name), bz)
}

func (k Keeper) SubmitProposal(ctx sdk.Context, domainName, issueName, suggestionName string, creator string, fee sdk.Coins) error {
    store := ctx.KVStore(k.storeKey)
    domainBz := store.Get([]byte("domain:" + domainName))
    if domainBz == nil {
        return typeserrors.ErrUnknownRequest
    }
    
    var domain Domain
    if err := json.Unmarshal(domainBz, &domain); err != nil {
        return typeserrors.ErrInvalidRequest
    }

    if domain.Options.OnlyAdminIssues && creator != domain.Admin.String() {
        return typeserrors.ErrUnauthorized
    }
    if domain.Options.CoinBurnRequired && fee.AmountOf("pnyx").LT(sdk.NewInt(100)) {
        return typeserrors.ErrInsufficientFunds
    }
    if fee.AmountOf("pnyx").LT(sdk.NewInt(15)) {
        return typeserrors.ErrInsufficientFunds
    }
    
    domain.Treasury = domain.Treasury.Add(fee...)

    found := false
    for i, issue := range domain.Issues {
        if issue.Name == issueName {
            domain.Issues[i].Suggestions = append(domain.Issues[i].Suggestions, Suggestion{
                Name: suggestionName,
            })
            found = true
            break
        }
    }
    if !found {
        domain.Issues = append(domain.Issues, Issue{
            Name:        issueName,
            Suggestions: []Suggestion{{Name: suggestionName}},
        })
    }

    bz, _ := json.Marshal(domain)
    store.Set([]byte("domain:"+domainName), bz)
    
    return nil
}

func (k Keeper) RateProposal(ctx sdk.Context, domainName, issueName, suggestionName, voter string, rating int, privKey *ed25519.PrivKey) (sdk.Coins, map[string]interface{}, error) {
    store := ctx.KVStore(k.storeKey)
    domainBz := store.Get([]byte("domain:" + domainName))
    if domainBz == nil {
        return nil, nil, typeserrors.ErrUnknownRequest
    }

    var domain Domain
    if err := json.Unmarshal(domainBz, &domain); err != nil {
        return nil, nil, typeserrors.ErrInvalidRequest
    }

    for i, issue := range domain.Issues {
        if issue.Name == issueName {
            for j, suggestion := range issue.Suggestions {
                if suggestion.Name == suggestionName {
                    domain.Issues[i].Suggestions[j].Ratings = append(suggestion.Ratings, Rating{
                        Voter: voter,
                        Value: rating,
                    })
                    reward := sdk.NewCoins(sdk.NewInt64Coin("pnyx", 10))
                    cache := map[string]interface{}{
                        "avg_rating": rating,
                        "stones":     suggestion.Stones,
                        "treasury":   domain.Treasury,
                    }
                    bz, _ := json.Marshal(domain)
                    store.Set([]byte("domain:"+domainName), bz)
                    return reward, cache, nil
                }
            }
            return nil, nil, typeserrors.ErrUnknownRequest
        }
    }
    return nil, nil, typeserrors.ErrUnknownRequest
}
