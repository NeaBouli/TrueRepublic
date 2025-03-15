package truedemocracy

import (
    "encoding/json"
    "math"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
    "github.com/cosmos/cosmos-sdk/store/types"
    typeserrors "github.com/cosmos/cosmos-sdk/types/errors"
    "time"
)

const (
    C_dmn     = 2
    C_put     = 15
    Stake     = 100000
    C_earn    = 1000
    APY_dom   = 0.25
    APY_node  = 0.1
    SupplyMax = 21000000
)

type Keeper struct {
    storeKey types.StoreKey
    nodes    []*Node
    members  map[string]Member
}

func NewKeeper(storeKey types.StoreKey, nodes []*Node) Keeper {
    return Keeper{storeKey: storeKey, nodes: nodes, members: make(map[string]Member)}
}

func (k Keeper) CreateDomain(ctx sdk.Context, name string, admin sdk.AccAddress, initialCoins sdk.Coins) error {
    store := ctx.KVStore(k.storeKey)
    if store.Has([]byte("domain:" + name)) {
        return typeserrors.Wrap(typeserrors.ErrInvalidRequest, "Domain already exists")
    }
    fee := ctx.GasPrices().AmountOf("pnyx")
    p_dmn := fee.Mul(sdk.NewInt(C_dmn)).Mul(sdk.NewInt(C_earn))
    if initialCoins.AmountOf("pnyx").LT(p_dmn) {
        return typeserrors.Wrapf(typeserrors.ErrInsufficientFunds, "Need at least %s PNYX", p_dmn.String())
    }
    domain := Domain{
        Name:          name,
        Admin:         admin,
        Members:       []Member{{Avatar: admin.String(), IsActive: true, Weight: sdk.NewDecWithPrec(2, 1)}},
        Treasury:      initialCoins,
        Issues:        []Issue{},
        Options:       DomainOptions{AdminElectable: true, MaxGreenZone: 12, DwellTime: 24 * time.Hour},
        PermissionReg: []string{},
        GlobalKeys:    make(map[string]string),
    }
    bz, err := json.Marshal(domain)
    if err != nil {
        return typeserrors.Wrap(err, "Failed to marshal domain")
    }
    store.Set([]byte("domain:"+name), bz)
    return nil
}

func (k Keeper) SubmitProposal(ctx sdk.Context, domainName, issueName, suggestionName string, creator string, fee sdk.Coins) error {
    store := ctx.KVStore(k.storeKey)
    domainBz := store.Get([]byte("domain:" + domainName))
    if domainBz == nil {
        return typeserrors.Wrap(typeserrors.ErrUnknownRequest, "Domain not found")
    }
    var domain Domain
    if err := json.Unmarshal(domainBz, &domain); err != nil {
        return typeserrors.Wrap(err, "Failed to unmarshal domain")
    }
    if domain.Options.OnlyAdminIssues && creator != domain.Admin.String() {
        return typeserrors.Wrap(typeserrors.ErrUnauthorized, "Only admin can submit")
    }
    p_rew := domain.Treasury.AmountOf("pnyx").Quo(sdk.NewInt(C_earn))
    p_put := sdk.MinInt(p_rew.Mul(sdk.NewInt(C_put)), p_rew.Mul(sdk.NewInt(int64(len(domain.Members)))))
    if fee.AmountOf("pnyx").LT(p_put) {
        return typeserrors.Wrapf(typeserrors.ErrInsufficientFunds, "Need at least %s PNYX", p_put.String())
    }
    domain.Treasury = domain.Treasury.Add(fee...)
    found := false
    for i, issue := range domain.Issues {
        if issue.Name == issueName {
            domain.Issues[i].Suggestions = append(domain.Issues[i].Suggestions, Suggestion{
                Name:      suggestionName,
                Stones:    sdk.ZeroInt(),
                Zone:      "yellow",
                Created:   ctx.BlockTime(),
                ShortDesc: "Submitted by " + creator,
            })
            found = true
            break
        }
    }
    if !found {
        domain.Issues = append(domain.Issues, Issue{
            Name:        issueName,
            Suggestions: []Suggestion{{Name: suggestionName, Stones: sdk.ZeroInt(), Zone: "yellow", Created: ctx.BlockTime(), ShortDesc: "Submitted by " + creator}},
            Stones:      sdk.ZeroInt(),
            Consensus:   "open",
            Created:     ctx.BlockTime(),
        })
    }
    bz, err := json.Marshal(domain)
    if err != nil {
        return typeserrors.Wrap(err, "Failed to marshal domain")
    }
    store.Set([]byte("domain:"+domainName), bz)
    return nil
}

func (k Keeper) RateProposal(ctx sdk.Context, domainName, issueName, suggestionName, voter string, rating int, privKey *ed25519.PrivKey) (sdk.Coins, map[string]interface{}, error) {
    if rating < -5 || rating > 5 {
        return nil, nil, typeserrors.Wrap(typeserrors.ErrInvalidRequest, "Rating must be between -5 and 5")
    }
    store := ctx.KVStore(k.storeKey)
    domainBz := store.Get([]byte("domain:" + domainName))
    if domainBz == nil {
        return nil, nil, typeserrors.Wrap(typeserrors.ErrUnknownRequest, "Domain not found")
    }
    var domain Domain
    if err := json.Unmarshal(domainBz, &domain); err != nil {
        return nil, nil, typeserrors.Wrap(err, "Failed to unmarshal domain")
    }
    voterKey := sdk.AccAddress(privKey.PubKey().Bytes()).String()
    p_rew := domain.Treasury.AmountOf("pnyx").Quo(sdk.NewInt(C_earn))
    if p_rew.LT(sdk.OneInt()) {
        return nil, nil, typeserrors.Wrap(typeserrors.ErrInsufficientFunds, "Treasury too low for reward")
    }
    reward := sdk.NewCoins(sdk.NewCoin("pnyx", p_rew))
    for i, issue := range domain.Issues {
        if issue.Name == issueName && issue.Consensus == "open" {
            for j, suggestion := range issue.Suggestions {
                if suggestion.Name == suggestionName {
                    domain.Issues[i].Suggestions[j].Ratings = append(suggestion.Ratings, Rating{
                        VoterKey: voterKey,
                        Value:    rating,
                    })
                    k.updateSuggestionZone(&domain, i, j)
                    avg := k.calculateSystemicConsensus(domain.Issues[i].Suggestions)
                    cache := map[string]interface{}{
                        "avg_rating": avg,
                        "stones":     suggestion.Stones,
                        "treasury":   domain.Treasury,
                    }
                    bz, err := json.Marshal(domain)
                    if err != nil {
                        return nil, nil, typeserrors.Wrap(err, "Failed to marshal domain")
                    }
                    store.Set([]byte("domain:"+domainName), bz)
                    k.updateMemberVotes(&domain, voter)
                    return reward, cache, nil
                }
            }
        }
    }
    return nil, nil, typeserrors.Wrap(typeserrors.ErrUnknownRequest, "Suggestion not found")
}

func (k Keeper) AddStones(ctx sdk.Context, domainName, issueName, suggestionName string, voter string) (sdk.Coins, error) {
    store := ctx.KVStore(k.storeKey)
    domainBz := store.Get([]byte("domain:" + domainName))
    if domainBz == nil {
        return nil, typeserrors.Wrap(typeserrors.ErrUnknownRequest, "Domain not found")
    }
    var domain Domain
    if err := json.Unmarshal(domainBz, &domain); err != nil {
        return nil, typeserrors.Wrap(err, "Failed to unmarshal domain")
    }
    p_rew := domain.Treasury.AmountOf("pnyx").Quo(sdk.NewInt(C_earn))
    if p_rew.LT(sdk.OneInt()) {
        return nil, typeserrors.Wrap(typeserrors.ErrInsufficientFunds, "Treasury too low for reward")
    }
    reward := sdk.NewCoins(sdk.NewCoin("pnyx", p_rew))
    for i, issue := range domain.Issues {
        if issue.Name == issueName {
            for j, suggestion := range issue.Suggestions {
                if suggestion.Name == suggestionName {
                    if suggestion.Stones.LT(sdk.NewInt(int64(len(domain.Members)))) {
                        domain.Issues[i].Suggestions[j].Stones = suggestion.Stones.Add(sdk.OneInt())
                        domain.Issues[i].Stones = issue.Stones.Add(sdk.OneInt())
                        k.updateSuggestionZone(&domain, i, j)
                        bz, err := json.Marshal(domain)
                        if err != nil {
                            return nil, typeserrors.Wrap(err, "Failed to marshal domain")
                        }
                        store.Set([]byte("domain:"+domainName), bz)
                        k.updateMemberVotes(&domain, voter)
                        return reward, nil
                    }
                    return nil, typeserrors.Wrap(typeserrors.ErrInvalidRequest, "Stone limit reached")
                }
            }
        }
    }
    return nil, typeserrors.Wrap(typeserrors.ErrUnknownRequest, "Suggestion not found")
}

func (k Keeper) FinalizeIssue(ctx sdk.Context, domainName, issueName string) error {
    store := ctx.KVStore(k.storeKey)
    domainBz := store.Get([]byte("domain:" + domainName))
    if domainBz == nil {
        return typeserrors.Wrap(typeserrors.ErrUnknownRequest, "Domain not found")
    }
    var domain Domain
    if err := json.Unmarshal(domainBz, &domain); err != nil {
        return typeserrors.Wrap(err, "Failed to unmarshal domain")
    }
    for i, issue := range domain.Issues {
        if issue.Name == issueName && issue.Consensus == "open" {
            consensus := k.calculateSystemicConsensus(issue.Suggestions)
            if consensus > 0 {
                domain.Issues[i].Consensus = "accepted"
            } else {
                domain.Issues[i].Consensus = "rejected"
            }
            bz, err := json.Marshal(domain)
            if err != nil {
                return typeserrors.Wrap(err, "Failed to marshal domain")
            }
            store.Set([]byte("domain:"+domainName), bz)
            return nil
        }
    }
    return typeserrors.Wrap(typeserrors.ErrUnknownRequest, "Issue not found or already finalized")
}

func (k Keeper) Stake(ctx sdk.Context, delegator sdk.AccAddress, nodeAddress string, amount sdk.Coins) error {
    store := ctx.KVStore(k.storeKey)
    domainBz := store.Get([]byte("domain:governance"))
    if domainBz == nil {
        return typeserrors.Wrap(typeserrors.ErrUnknownRequest, "Governance domain not found")
    }
    var domain Domain
    if err := json.Unmarshal(domainBz, &domain); err != nil {
        return typeserrors.Wrap(err, "Failed to unmarshal domain")
    }
    totalPayouts := sdk.ZeroInt()
    for _, issue := range domain.Issues {
        for _, suggestion := range issue.Suggestions {
            totalPayouts = totalPayouts.Add(sdk.NewInt(int64(len(suggestion.Ratings))))
        }
    }
    maxStake := totalPayouts.Quo(sdk.NewInt(10))
    if amount.AmountOf("pnyx").GT(maxStake) {
        return typeserrors.Wrapf(typeserrors.ErrInvalidRequest, "Stake exceeds PoD limit of %s PNYX", maxStake.String())
    }
    for i, node := range k.nodes {
        if node.Address == nodeAddress {
            if node.Staked.AmountOf("pnyx").Add(amount.AmountOf("pnyx")).LT(sdk.NewInt(Stake)) {
                return typeserrors.Wrap(typeserrors.ErrInsufficientFunds, "Stake below minimum")
            }
            k.nodes[i].Staked = node.Staked.Add(amount...)
            k.nodes[i].DomainOrigin = "governance"
            k.nodes[i].LastActive = ctx.BlockTime()
            return nil
        }
    }
    return typeserrors.Wrap(typeserrors.ErrUnknownRequest, "Node not found")
}

func (k Keeper) FundNode(ctx sdk.Context, nodeAddress string, amount sdk.Coins) error {
    for i, node := range k.nodes {
        if node.Address == nodeAddress {
            k.nodes[i].Staked = node.Staked.Add(amount...)
            k.nodes[i].LastActive = ctx.BlockTime()
            return nil
        }
    }
    return typeserrors.Wrap(typeserrors.ErrUnknownRequest, "Node not found")
}

func (k Keeper) updateSuggestionZone(domain *Domain, issueIdx, sugIdx int) {
    suggestion := domain.Issues[issueIdx].Suggestions[sugIdx]
    approval := suggestion.Stones.ToDec().Quo(sdk.NewDec(int64(len(domain.Members)))).Mul(sdk.NewDec(100))
    if approval.GTE(sdk.NewDec(5)) && len(domain.Issues[issueIdx].Suggestions) <= domain.Options.MaxGreenZone {
        domain.Issues[issueIdx].Suggestions[sugIdx].Zone = "green"
    } else if time.Since(suggestion.Created) > domain.Options.DwellTime {
        domain.Issues[issueIdx].Suggestions[sugIdx].Zone = "red"
    }
}

func (k Keeper) calculateSystemicConsensus(suggestions []Suggestion) float64 {
    totalResistance := 0
    totalVotes := 0
    for _, s := range suggestions {
        for _, r := range s.Ratings {
            totalResistance += int(math.Abs(float64(r.Value)))
            totalVotes++
        }
    }
    if totalVotes == 0 {
        return 0
    }
    return float64(totalResistance) / float64(totalVotes) * -1
}

func (k Keeper) updateMemberVotes(domain *Domain, voter string) {
    for i, member := range domain.Members {
        if member.Avatar == voter {
            domain.Members[i].VotesCast++
            if !member.IsActive && member.VotesCast <= 10 {
                domain.Members[i].Weight = sdk.NewDecWithPrec(1, 1).Quo(sdk.NewDec(10)).Mul(sdk.NewDec(int64(member.VotesCast)))
            } else if !member.IsActive {
                domain.Members[i].Weight = sdk.NewDecWithPrec(1, 1)
            }
            k.members[voter] = domain.Members[i]
            break
        }
    }
}
