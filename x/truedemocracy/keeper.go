package truedemocracy

import (
    errorsmod "cosmossdk.io/errors"
    storetypes "cosmossdk.io/store/types"
    "github.com/cosmos/cosmos-sdk/codec"
    "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    rewards "truerepublic/treasury/keeper"
)

type Keeper struct {
    StoreKey storetypes.StoreKey
    nodes    []*Node
    cdc      *codec.LegacyAmino
}

func NewKeeper(cdc *codec.LegacyAmino, storeKey storetypes.StoreKey, nodes []*Node) Keeper {
    return Keeper{StoreKey: storeKey, nodes: nodes, cdc: cdc}
}

func (k Keeper) CreateDomain(ctx sdk.Context, name string, admin sdk.AccAddress, initialCoins sdk.Coins) {
    store := ctx.KVStore(k.StoreKey)
    domain := Domain{
        Name:          name,
        Admin:         admin,
        Members:       []string{admin.String()},
        Treasury:      initialCoins,
        Issues:        []Issue{},
        Options:       DomainOptions{AdminElectable: true, AnyoneCanJoin: false},
        PermissionReg: []string{},
    }
    bz := k.cdc.MustMarshalLengthPrefixed(&domain)
    store.Set([]byte("domain:"+name), bz)
}

func (k Keeper) SubmitProposal(ctx sdk.Context, domainName, issueName, suggestionName, creator string, fee sdk.Coins) error {
    store := ctx.KVStore(k.StoreKey)
    domainBz := store.Get([]byte("domain:" + domainName))
    if domainBz == nil {
        return errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "Domain not found")
    }
    var domain Domain
    k.cdc.MustUnmarshalLengthPrefixed(domainBz, &domain)

    if domain.Options.OnlyAdminIssues && creator != domain.Admin.String() {
        return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "Only admin can submit issues")
    }
    if domain.Options.CoinBurnRequired && fee.AmountOf("pnyx").LT(rewards.CalcDomainCost(fee.AmountOf("pnyx"))) {
        return errorsmod.Wrap(sdkerrors.ErrInsufficientFunds, "Coin burn requirement not met")
    }
    putPrice := rewards.CalcPutPrice(domain.Treasury.AmountOf("pnyx"), int64(len(domain.Members)))
    if putPrice.IsPositive() && fee.AmountOf("pnyx").LT(putPrice) {
        return errorsmod.Wrap(sdkerrors.ErrInsufficientFunds, "Fee below put price (eq.3)")
    }
    domain.Treasury = domain.Treasury.Add(fee...)

    found := false
    for i, issue := range domain.Issues {
        if issue.Name == issueName {
            domain.Issues[i].Suggestions = append(domain.Issues[i].Suggestions, Suggestion{
                Name:      suggestionName,
                Creator:   creator,
                Ratings:   []Rating{},
                Stones:    0,
                Color:     "",
                DwellTime: 0,
            })
            found = true
            break
        }
    }
    if !found {
        domain.Issues = append(domain.Issues, Issue{
            Name:        issueName,
            Suggestions: []Suggestion{{Name: suggestionName, Creator: creator, Ratings: []Rating{}, Stones: 0, Color: "", DwellTime: 0}},
            Stones:      0,
        })
    }

    bz := k.cdc.MustMarshalLengthPrefixed(&domain)
    store.Set([]byte("domain:"+domainName), bz)
    return nil
}

func (k Keeper) RateProposal(ctx sdk.Context, domainName, issueName, suggestionName, voter string, rating int, privKey *ed25519.PrivKey) (sdk.Coins, map[string]interface{}, error) {
    store := ctx.KVStore(k.StoreKey)
    domainBz := store.Get([]byte("domain:" + domainName))
    if domainBz == nil {
        return sdk.Coins{}, nil, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "Domain not found")
    }
    var domain Domain
    k.cdc.MustUnmarshalLengthPrefixed(domainBz, &domain)

    foundIssue := false
    foundSuggestion := false
    for i, issue := range domain.Issues {
        if issue.Name == issueName {
            foundIssue = true
            for j, suggestion := range issue.Suggestions {
                if suggestion.Name == suggestionName {
                    domain.Issues[i].Suggestions[j].Ratings = append(domain.Issues[i].Suggestions[j].Ratings, Rating{
                        Voter: voter,
                        Value: rating,
                    })
                    foundSuggestion = true
                    break
                }
            }
            break
        }
    }
    if !foundIssue || !foundSuggestion {
        return sdk.Coins{}, nil, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "Issue or suggestion not found")
    }

    bz := k.cdc.MustMarshalLengthPrefixed(&domain)
    store.Set([]byte("domain:"+domainName), bz)

    rewardAmt := rewards.CalcReward(domain.Treasury.AmountOf("pnyx"))
    reward := sdk.NewCoins(sdk.NewCoin("pnyx", rewardAmt))
    domain.Treasury = domain.Treasury.Sub(reward...)

    bz = k.cdc.MustMarshalLengthPrefixed(&domain)
    store.Set([]byte("domain:"+domainName), bz)

    cache := map[string]interface{}{
        "avg_rating": rating,
        "stones":     0,
        "treasury":   domain.Treasury.String(),
    }
    return reward, cache, nil
}
