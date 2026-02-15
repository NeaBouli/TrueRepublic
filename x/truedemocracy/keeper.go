package truedemocracy

import (
    "encoding/hex"
    "fmt"

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

// GetDomain loads a domain from the KV store by name.
func (k Keeper) GetDomain(ctx sdk.Context, name string) (Domain, bool) {
    store := ctx.KVStore(k.StoreKey)
    bz := store.Get([]byte("domain:" + name))
    if bz == nil {
        return Domain{}, false
    }
    var domain Domain
    k.cdc.MustUnmarshalLengthPrefixed(bz, &domain)
    return domain, true
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
                Name:         suggestionName,
                Creator:      creator,
                Ratings:      []Rating{},
                Stones:       0,
                Color:        "",
                DwellTime:    0,
                CreationDate: ctx.BlockTime().Unix(),
            })
            found = true
            break
        }
    }
    if !found {
        domain.Issues = append(domain.Issues, Issue{
            Name:         issueName,
            Suggestions:  []Suggestion{{Name: suggestionName, Creator: creator, Ratings: []Rating{}, Stones: 0, Color: "", DwellTime: 0, CreationDate: ctx.BlockTime().Unix()}},
            Stones:       0,
            CreationDate: ctx.BlockTime().Unix(),
        })
    }

    bz := k.cdc.MustMarshalLengthPrefixed(&domain)
    store.Set([]byte("domain:"+domainName), bz)
    return nil
}

// RateProposal records an anonymous rating on a suggestion. The caller proves
// they control a key in the domain's permission register by providing their
// domain-specific private key. The voter's avatar name is never stored —
// only the domain public key hex appears on-chain (whitepaper Section 4).
func (k Keeper) RateProposal(ctx sdk.Context, domainName, issueName, suggestionName string, rating int, domainPrivKey *ed25519.PrivKey) (sdk.Coins, map[string]interface{}, error) {
    if domainPrivKey == nil {
        return sdk.Coins{}, nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "domain private key is required for anonymous voting")
    }

    // Derive domain public key (anonymous identity).
    domainPubKeyHex := hex.EncodeToString(domainPrivKey.PubKey().Bytes())

    store := ctx.KVStore(k.StoreKey)
    domainBz := store.Get([]byte("domain:" + domainName))
    if domainBz == nil {
        return sdk.Coins{}, nil, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "Domain not found")
    }
    var domain Domain
    k.cdc.MustUnmarshalLengthPrefixed(domainBz, &domain)

    // Verify domain key is in the permission register.
    if !k.IsKeyAuthorized(ctx, domainName, domainPubKeyHex) {
        return sdk.Coins{}, nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "domain key not in permission register")
    }

    // Sign the vote payload to prove key ownership.
    payload := []byte(fmt.Sprintf("%s|%s|%s|%d", domainName, issueName, suggestionName, rating))
    sig, err := domainPrivKey.Sign(payload)
    if err != nil {
        return sdk.Coins{}, nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "failed to sign vote payload")
    }
    // Verify the signature (proves caller controls the key).
    if !domainPrivKey.PubKey().VerifySignature(payload, sig) {
        return sdk.Coins{}, nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "vote signature verification failed")
    }

    // Prevent double-voting with the same domain key.
    if HasDomainKeyVoted(domain, issueName, suggestionName, domainPubKeyHex) {
        return sdk.Coins{}, nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "domain key has already voted on this suggestion")
    }

    foundIssue := false
    foundSuggestion := false
    for i, issue := range domain.Issues {
        if issue.Name == issueName {
            foundIssue = true
            for j, suggestion := range issue.Suggestions {
                if suggestion.Name == suggestionName {
                    domain.Issues[i].Suggestions[j].Ratings = append(domain.Issues[i].Suggestions[j].Ratings, Rating{
                        DomainPubKeyHex: domainPubKeyHex,
                        Value:           rating,
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
