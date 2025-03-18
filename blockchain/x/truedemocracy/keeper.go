package truedemocracy

import (
    "encoding/json"
    "errors"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "time"
)

type Keeper struct {
    storeKey sdk.StoreKey
}

type Domain struct {
    Name     string    `json:"name"`
    Created  time.Time `json:"created"`
    Members  []string  `json:"members"`  // Member-Listen hinzugefügt
    Issues   []Issue   `json:"issues"`
    Staked   sdk.Int   `json:"staked"`   // PoD: Staked PNYX für Domain
}

type Issue struct {
    Name        string       `json:"name"`
    Suggestions []Suggestion `json:"suggestions"`
}

type Suggestion struct {
    Name  string `json:"name"`
    Votes []int8 `json:"votes"` // -5 bis +5
}

func NewKeeper(storeKey sdk.StoreKey) Keeper {
    return Keeper{storeKey: storeKey}
}

func (k Keeper) CreateDomain(ctx sdk.Context, name string, creator string, stake sdk.Int) error {
    store := ctx.KVStore(k.storeKey)
    if store.Has([]byte("domain:" + name)) {
        return errors.New("Domain exists")
    }
    domain := Domain{
        Name:    name,
        Created: ctx.BlockTime(),
        Members: []string{creator}, // Creator als erster Member
        Staked:  stake,             // PoD: Initialer Stake
    }
    bz, _ := json.Marshal(domain)
    store.Set([]byte("domain:"+name), bz)
    return nil
}

func (k Keeper) AddMember(ctx sdk.Context, domainName, member string, stake sdk.Int) error {
    store := ctx.KVStore(k.storeKey)
    domainBz := store.Get([]byte("domain:" + domainName))
    if domainBz == nil {
        return errors.New("Domain not found")
    }
    var domain Domain
    json.Unmarshal(domainBz, &domain)
    for _, m := range domain.Members {
        if m == member {
            return errors.New("Member already exists")
        }
    }
    domain.Members = append(domain.Members, member)
    domain.Staked = domain.Staked.Add(stake) // PoD: Stake erhöhen
    bz, _ := json.Marshal(domain)
    store.Set([]byte("domain:"+domainName), bz)
    return nil
}

func (k Keeper) SubmitProposal(ctx sdk.Context, domainName, issueName, suggestionName string, fee sdk.Coins) error {
    store := ctx.KVStore(k.storeKey)
    domainBz := store.Get([]byte("domain:" + domainName))
    if domainBz == nil {
        return errors.New("Domain not found")
    }
    var domain Domain
    json.Unmarshal(domainBz, &domain)
    domain.Issues = append(domain.Issues, Issue{
        Name: issueName,
        Suggestions: []Suggestion{{Name: suggestionName}},
    })
    bz, _ := json.Marshal(domain)
    store.Set([]byte("domain:"+domainName), bz)
    return nil
}

func (k Keeper) Vote(ctx sdk.Context, domainName, issueName, suggestionName string, vote int8) error {
    if vote < -5 || vote > 5 {
        return errors.New("Vote must be between -5 and 5")
    }
    store := ctx.KVStore(k.storeKey)
    domainBz := store.Get([]byte("domain:" + domainName))
    if domainBz == nil {
        return errors.New("Domain not found")
    }
    var domain Domain
    json.Unmarshal(domainBz, &domain)
    for i, issue := range domain.Issues {
        if issue.Name == issueName {
            for j, suggestion := range issue.Suggestions {
                if suggestion.Name == suggestionName {
                    domain.Issues[i].Suggestions[j].Votes = append(suggestion.Votes, vote)
                    bz, _ := json.Marshal(domain)
                    store.Set([]byte("domain:"+domainName), bz)
                    return nil
                }
            }
            return errors.New("Suggestion not found")
        }
    }
    return errors.New("Issue not found")
}
