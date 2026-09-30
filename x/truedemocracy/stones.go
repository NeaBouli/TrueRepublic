package truedemocracy

import (
	"encoding/binary"
	"sort"
	"strings"

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
//
// The first-placement VoteToEarn reward (GH-306) is tracked separately with
// persistent consumed markers that survive moves, exclusion cleanup, and
// genesis export/import:
//   "stone-reward:" + scope byte + uint32-length-prefixed fields → StoneRewardRecord

func issueStoneKey(domainName, memberAddr string) []byte {
	return []byte("stone:i:" + domainName + ":" + memberAddr)
}

func suggestionStoneKey(domainName, issueName, memberAddr string) []byte {
	return []byte("stone:s:" + domainName + ":" + issueName + ":" + memberAddr)
}

const stoneRewardRecordPrefix = "stone-reward:"

// stoneRewardRecordKey derives the consumed-marker key for one member and
// voting scope. An empty IssueName selects the domain issue-list scope.
func stoneRewardRecordKey(record StoneRewardRecord) []byte {
	appendPart := func(key []byte, value string) []byte {
		var size [4]byte
		binary.BigEndian.PutUint32(size[:], uint32(len(value)))
		key = append(key, size[:]...)
		return append(key, value...)
	}
	key := []byte(stoneRewardRecordPrefix)
	if record.IssueName == "" {
		key = append(key, 'i')
		key = appendPart(key, record.DomainName)
		return appendPart(key, record.MemberAddr)
	}
	key = append(key, 's')
	key = appendPart(key, record.DomainName)
	key = appendPart(key, record.IssueName)
	return appendPart(key, record.MemberAddr)
}

// HasStoneRewardRecord reports whether the first-placement reward for the
// member's voting scope was already consumed.
func (k Keeper) HasStoneRewardRecord(ctx sdk.Context, record StoneRewardRecord) bool {
	return ctx.KVStore(k.StoreKey).Has(stoneRewardRecordKey(record))
}

// SetStoneRewardRecord persists the consumed marker. The marshaled record is
// stored as the value so export never has to parse ambiguous key segments.
func (k Keeper) SetStoneRewardRecord(ctx sdk.Context, record StoneRewardRecord) {
	ctx.KVStore(k.StoreKey).Set(stoneRewardRecordKey(record), k.cdc.MustMarshalLengthPrefixed(&record))
}

// IterateStoneRewardRecords visits every consumed marker in deterministic
// store key order. Returning true stops iteration.
func (k Keeper) IterateStoneRewardRecords(ctx sdk.Context, fn func(StoneRewardRecord) bool) {
	store := ctx.KVStore(k.StoreKey)
	prefix := []byte(stoneRewardRecordPrefix)
	iter := store.Iterator(prefix, prefixEnd(prefix))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var record StoneRewardRecord
		k.cdc.MustUnmarshalLengthPrefixed(iter.Value(), &record)
		if fn(record) {
			return
		}
	}
}

// BaselineStoneRewardMarkers is the deterministic GH-306 2→3 migration step:
// every stone that exists at the upgrade boundary is marked as already
// rewarded, so a legacy placement can never earn the first-placement reward a
// second time. Keys are scanned globally and resolved against known domain
// names, so lifecycle-cleaned suggestion scopes are included and names
// containing ':' remain deterministic. The step is idempotent: existing
// markers are left untouched.
func (k Keeper) BaselineStoneRewardMarkers(ctx sdk.Context) {
	store := ctx.KVStore(k.StoreKey)
	baseline := func(record StoneRewardRecord) {
		if key := stoneRewardRecordKey(record); !store.Has(key) {
			store.Set(key, k.cdc.MustMarshalLengthPrefixed(&record))
		}
	}
	var domainNames []string
	k.IterateDomains(ctx, func(domain Domain) bool {
		domainNames = append(domainNames, domain.Name)
		return false
	})
	issuePrefix := []byte("stone:i:")
	issueIter := store.Iterator(issuePrefix, prefixEnd(issuePrefix))
	for ; issueIter.Valid(); issueIter.Next() {
		rest := string(issueIter.Key()[len(issuePrefix):])
		for _, domainName := range domainNames {
			prefix := domainName + ":"
			if !strings.HasPrefix(rest, prefix) {
				continue
			}
			member := strings.TrimPrefix(rest, prefix)
			// Runtime member addresses cannot contain ':'. Skipping the shorter
			// ambiguous prefix therefore selects the only valid issue-key split.
			if member != "" && !strings.ContainsRune(member, ':') {
				baseline(StoneRewardRecord{DomainName: domainName, MemberAddr: member})
			}
		}
	}
	issueIter.Close()

	// Lifecycle cleanup can leave suggestion-stone keys whose issue no longer
	// exists. They still consumed their first-placement reward and must be
	// baselined at the version boundary.
	suggestionPrefix := []byte("stone:s:")
	suggestionIter := store.Iterator(suggestionPrefix, prefixEnd(suggestionPrefix))
	for ; suggestionIter.Valid(); suggestionIter.Next() {
		rest := string(suggestionIter.Key()[len(suggestionPrefix):])
		for _, domainName := range domainNames {
			prefix := domainName + ":"
			if !strings.HasPrefix(rest, prefix) {
				continue
			}
			scopeAndMember := strings.TrimPrefix(rest, prefix)
			lastSeparator := strings.LastIndexByte(scopeAndMember, ':')
			if lastSeparator <= 0 || lastSeparator == len(scopeAndMember)-1 {
				continue
			}
			// Legacy delimiter keys can be ambiguous when domain and issue
			// names contain ':'. Baselining every valid interpretation is the
			// conservative, deterministic choice: it can only consume a reward,
			// never create a second payout.
			baseline(StoneRewardRecord{
				DomainName: domainName,
				IssueName:  scopeAndMember[:lastSeparator],
				MemberAddr: scopeAndMember[lastSeparator+1:],
			})
		}
	}
	suggestionIter.Close()
}

// PlaceStoneOnIssue places (or moves) the member's stone on an issue in the
// domain's issue list. If the member already has a stone on a different issue,
// it is moved automatically (old issue -1, new issue +1). The VoteToEarn
// reward (whitepaper eq.2) is paid from the domain treasury at most once per
// member and scope: only the very first placement earns it; moves and
// re-entries after cleanup never pay again (GH-306).
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
	firstPlacement := true
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
		firstPlacement = false
	}

	// Increment target issue and update activity.
	domain.Issues[targetIdx].Stones++
	domain.Issues[targetIdx].LastActivityAt = ctx.BlockTime().Unix()
	store.Set(key, []byte(issueName))

	// VoteToEarn reward (eq.2) only for the first placement in this scope.
	reward := sdk.Coins{}
	if firstPlacement {
		record := StoneRewardRecord{DomainName: domainName, MemberAddr: memberAddr}
		if !k.HasStoneRewardRecord(ctx, record) {
			reward = k.payStoneReward(&domain)
			k.SetStoneRewardRecord(ctx, record)
		}
	}

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
	firstPlacement := true
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
		firstPlacement = false
	}

	// Increment target suggestion and update issue activity.
	domain.Issues[issueIdx].Suggestions[targetIdx].Stones++
	domain.Issues[issueIdx].LastActivityAt = ctx.BlockTime().Unix()
	store.Set(key, []byte(suggestionName))

	// VoteToEarn reward (eq.2) only for the first placement in this scope.
	reward := sdk.Coins{}
	if firstPlacement {
		record := StoneRewardRecord{DomainName: domainName, IssueName: issueName, MemberAddr: memberAddr}
		if !k.HasStoneRewardRecord(ctx, record) {
			reward = k.payStoneReward(&domain)
			k.SetStoneRewardRecord(ctx, record)
		}
	}

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
	rewardAmt := rewards.CalcReward(domain.Treasury.AmountOf(PNYXDenom))
	if !rewardAmt.IsPositive() {
		return sdk.Coins{}
	}
	reward := sdk.NewCoins(sdk.NewCoin(PNYXDenom, rewardAmt))
	domain.Treasury = domain.Treasury.Sub(reward...)
	domain.TotalPayouts += rewardAmt.Int64()
	return reward
}
