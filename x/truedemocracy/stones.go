package truedemocracy

import (
	"sort"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	rewards "truerepublic/treasury/keeper"
)

// Stone placement is tracked in the KV store with these key patterns:
//   "stone:i:{domainName}:{memberAddr}"                     → issue name
//   "stone:s:{domainName}:{issueName}:{memberAddr}"         → suggestion name
//
// Each member has exactly ONE stone per list:
//   - One stone on the domain's issue list
//   - One stone per suggestion list (one per issue)
//
// Placing a stone on a new entry automatically moves it from the old one.

func issueStoneKey(domainName, memberAddr string) []byte {
	return []byte("stone:i:" + domainName + ":" + memberAddr)
}

func suggestionStoneKey(domainName, issueName, memberAddr string) []byte {
	return []byte("stone:s:" + domainName + ":" + issueName + ":" + memberAddr)
}

// PlaceStoneOnIssue places (or moves) the member's stone on an issue in the
// domain's issue list. If the member already has a stone on a different issue,
// it is moved automatically (old issue -1, new issue +1). A VoteToEarn reward
// is paid from the domain treasury (whitepaper eq.2).
func (k Keeper) PlaceStoneOnIssue(ctx sdk.Context, domainName, issueName, memberAddr string) (sdk.Coins, error) {
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return sdk.Coins{}, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "domain %s not found", domainName)
	}

	if !isMember(domain, memberAddr) {
		return sdk.Coins{}, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "only domain members can place stones")
	}

	targetIdx := findIssueIndex(domain, issueName)
	if targetIdx == -1 {
		return sdk.Coins{}, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "issue not found")
	}

	store := ctx.KVStore(k.StoreKey)
	key := issueStoneKey(domainName, memberAddr)

	// Check if member already has a stone placed.
	if existing := store.Get(key); existing != nil {
		oldIssue := string(existing)
		if oldIssue == issueName {
			return sdk.Coins{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "stone already placed on this issue")
		}
		// Move: decrement old issue.
		for i, issue := range domain.Issues {
			if issue.Name == oldIssue {
				domain.Issues[i].Stones--
				break
			}
		}
	}

	// Increment target issue and update activity.
	domain.Issues[targetIdx].Stones++
	domain.Issues[targetIdx].LastActivityAt = ctx.BlockTime().Unix()
	store.Set(key, []byte(issueName))

	// VoteToEarn reward (eq.2).
	reward := k.payStoneReward(&domain)

	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	store.Set([]byte("domain:"+domainName), bz)
	return reward, nil
}

// PlaceStoneOnSuggestion places (or moves) the member's stone on a suggestion
// within an issue's suggestion list. Each issue has its own independent
// suggestion list, so a member can have one stone per issue's suggestion list.
func (k Keeper) PlaceStoneOnSuggestion(ctx sdk.Context, domainName, issueName, suggestionName, memberAddr string) (sdk.Coins, error) {
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return sdk.Coins{}, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "domain %s not found", domainName)
	}

	if !isMember(domain, memberAddr) {
		return sdk.Coins{}, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "only domain members can place stones")
	}

	issueIdx := findIssueIndex(domain, issueName)
	if issueIdx == -1 {
		return sdk.Coins{}, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "issue not found")
	}

	targetIdx := findSuggestionIndex(domain.Issues[issueIdx], suggestionName)
	if targetIdx == -1 {
		return sdk.Coins{}, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "suggestion not found")
	}

	store := ctx.KVStore(k.StoreKey)
	key := suggestionStoneKey(domainName, issueName, memberAddr)

	// Check if member already has a stone in this suggestion list.
	if existing := store.Get(key); existing != nil {
		oldSugg := string(existing)
		if oldSugg == suggestionName {
			return sdk.Coins{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "stone already placed on this suggestion")
		}
		// Move: decrement old suggestion.
		for j, s := range domain.Issues[issueIdx].Suggestions {
			if s.Name == oldSugg {
				domain.Issues[issueIdx].Suggestions[j].Stones--
				break
			}
		}
	}

	// Increment target suggestion and update issue activity.
	domain.Issues[issueIdx].Suggestions[targetIdx].Stones++
	domain.Issues[issueIdx].LastActivityAt = ctx.BlockTime().Unix()
	store.Set(key, []byte(suggestionName))

	// VoteToEarn reward (eq.2).
	reward := k.payStoneReward(&domain)

	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	store.Set([]byte("domain:"+domainName), bz)
	return reward, nil
}

// GetMemberIssueStone returns the issue the member's stone is currently on,
// or ("", false) if no stone is placed.
func (k Keeper) GetMemberIssueStone(ctx sdk.Context, domainName, memberAddr string) (string, bool) {
	store := ctx.KVStore(k.StoreKey)
	bz := store.Get(issueStoneKey(domainName, memberAddr))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}

// GetMemberSuggestionStone returns the suggestion the member's stone is on
// within a specific issue's suggestion list.
func (k Keeper) GetMemberSuggestionStone(ctx sdk.Context, domainName, issueName, memberAddr string) (string, bool) {
	store := ctx.KVStore(k.StoreKey)
	bz := store.Get(suggestionStoneKey(domainName, issueName, memberAddr))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}

// SortIssuesByStones sorts issues by Stones descending, then CreationDate
// ascending (oldest first on tie). Returns a new sorted slice.
func SortIssuesByStones(issues []Issue) []Issue {
	sorted := make([]Issue, len(issues))
	copy(sorted, issues)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Stones != sorted[j].Stones {
			return sorted[i].Stones > sorted[j].Stones
		}
		return sorted[i].CreationDate < sorted[j].CreationDate
	})
	return sorted
}

// SortSuggestionsByStones sorts suggestions by Stones descending, then
// CreationDate ascending (oldest first on tie). Returns a new sorted slice.
func SortSuggestionsByStones(suggestions []Suggestion) []Suggestion {
	sorted := make([]Suggestion, len(suggestions))
	copy(sorted, suggestions)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Stones != sorted[j].Stones {
			return sorted[i].Stones > sorted[j].Stones
		}
		return sorted[i].CreationDate < sorted[j].CreationDate
	})
	return sorted
}

// --- helpers ---

func isMember(domain Domain, addr string) bool {
	for _, m := range domain.Members {
		if m == addr {
			return true
		}
	}
	return false
}

func findIssueIndex(domain Domain, issueName string) int {
	for i, issue := range domain.Issues {
		if issue.Name == issueName {
			return i
		}
	}
	return -1
}

func findSuggestionIndex(issue Issue, suggestionName string) int {
	for j, s := range issue.Suggestions {
		if s.Name == suggestionName {
			return j
		}
	}
	return -1
}

// payStoneReward calculates and deducts the VoteToEarn reward (eq.2) from the
// domain treasury. Returns the reward coins (may be empty if treasury is low).
func (k Keeper) payStoneReward(domain *Domain) sdk.Coins {
	rewardAmt := rewards.CalcReward(domain.Treasury.AmountOf("pnyx"))
	if !rewardAmt.IsPositive() {
		return sdk.Coins{}
	}
	reward := sdk.NewCoins(sdk.NewCoin("pnyx", rewardAmt))
	domain.Treasury = domain.Treasury.Sub(reward...)
	domain.TotalPayouts += rewardAmt.Int64()
	return reward
}
