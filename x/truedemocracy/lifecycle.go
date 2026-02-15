package truedemocracy

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Suggestion lifecycle (whitepaper §3.1.2):
//
//   GREEN  (approval >= threshold)  →  stays permanently
//   YELLOW (approval < threshold)   →  dwell time to recover
//   RED    (still < threshold)      →  dwell time, then auto-deleted
//
// Transitions are evaluated every EndBlock. A suggestion can recover
// to green from yellow or red at any time by gaining enough stones.

// MeetsApprovalThreshold checks whether a suggestion's stone count meets
// the domain's approval threshold. Uses integer math to avoid floats:
//
//	stones * 10000 >= totalMembers * thresholdBps
func MeetsApprovalThreshold(stones, totalMembers int, thresholdBps int64) bool {
	if totalMembers <= 0 {
		return false
	}
	return int64(stones)*10000 >= int64(totalMembers)*thresholdBps
}

// effectiveThreshold returns the domain's approval threshold or the default.
func effectiveThreshold(opts DomainOptions) int64 {
	if opts.ApprovalThreshold > 0 {
		return opts.ApprovalThreshold
	}
	return DefaultApprovalThresholdBps
}

// effectiveDwellTime returns the suggestion's own dwell time, falling back to
// the domain default, then the global default.
func effectiveDwellTime(s Suggestion, opts DomainOptions) int64 {
	if s.DwellTime > 0 {
		return s.DwellTime
	}
	if opts.DefaultDwellTime > 0 {
		return opts.DefaultDwellTime
	}
	return DefaultDwellTimeSecs
}

// EvaluateSuggestionZones processes zone transitions for all suggestions in
// a domain. Called from EndBlock. Suggestions that expire in the red zone
// are automatically deleted.
func (k Keeper) EvaluateSuggestionZones(ctx sdk.Context, domainName string) error {
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return nil // domain may have been removed
	}

	now := ctx.BlockTime().Unix()
	totalMembers := len(domain.Members)
	threshold := effectiveThreshold(domain.Options)
	modified := false

	for i := range domain.Issues {
		var kept []Suggestion
		for j := range domain.Issues[i].Suggestions {
			s := &domain.Issues[i].Suggestions[j]

			if MeetsApprovalThreshold(s.Stones, totalMembers, threshold) {
				// Approved → green. Clear any zone timestamps.
				if s.Color != "green" {
					s.Color = "green"
					s.EnteredYellowAt = 0
					s.EnteredRedAt = 0
					modified = true
				}
				kept = append(kept, *s)
				continue
			}

			// Below threshold.
			dwellTime := effectiveDwellTime(*s, domain.Options)

			switch s.Color {
			case "", "green":
				// First drop below threshold → enter yellow.
				s.Color = "yellow"
				s.EnteredYellowAt = now
				s.EnteredRedAt = 0
				modified = true
				kept = append(kept, *s)

			case "yellow":
				if now >= s.EnteredYellowAt+dwellTime {
					// Yellow expired → enter red.
					s.Color = "red"
					s.EnteredRedAt = now
					modified = true
				}
				kept = append(kept, *s)

			case "red":
				if now >= s.EnteredRedAt+dwellTime {
					// Red expired → auto-delete.
					modified = true
					// Don't append → suggestion is removed.
				} else {
					kept = append(kept, *s)
				}
			default:
				kept = append(kept, *s)
			}
		}
		if len(kept) != len(domain.Issues[i].Suggestions) {
			modified = true
		}
		domain.Issues[i].Suggestions = kept
	}

	if modified {
		store := ctx.KVStore(k.StoreKey)
		bz := k.cdc.MustMarshalLengthPrefixed(&domain)
		store.Set([]byte("domain:"+domainName), bz)
	}
	return nil
}

// ProcessAllLifecycles evaluates suggestion zones for every domain.
// Called from EndBlock.
func (k Keeper) ProcessAllLifecycles(ctx sdk.Context) {
	var domainNames []string
	k.IterateDomains(ctx, func(d Domain) bool {
		domainNames = append(domainNames, d.Name)
		return false
	})
	for _, name := range domainNames {
		k.EvaluateSuggestionZones(ctx, name)
	}
}

// IterateDomains iterates over all domains in the KV store.
func (k Keeper) IterateDomains(ctx sdk.Context, fn func(Domain) bool) {
	store := ctx.KVStore(k.StoreKey)
	prefix := []byte("domain:")
	end := prefixEnd(prefix)
	iter := store.Iterator(prefix, end)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var domain Domain
		k.cdc.MustUnmarshalLengthPrefixed(iter.Value(), &domain)
		if fn(domain) {
			break
		}
	}
}

// --- Fast Delete (2/3 majority) ---

// deleteVoteKey returns the KV key for tracking an individual member's
// delete vote on a suggestion.
func deleteVoteKey(domainName, issueName, suggestionName, memberAddr string) []byte {
	return []byte("delvote:" + domainName + ":" + issueName + ":" + suggestionName + ":" + memberAddr)
}

// VoteToDelete records a member's vote to delete a suggestion. If the vote
// count reaches 2/3 of domain members, the suggestion is immediately removed.
// Returns (deleted bool, error).
func (k Keeper) VoteToDelete(ctx sdk.Context, domainName, issueName, suggestionName, memberAddr string) (bool, error) {
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return false, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "domain %s not found", domainName)
	}

	if !isMember(domain, memberAddr) {
		return false, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "only domain members can vote to delete")
	}

	issueIdx := findIssueIndex(domain, issueName)
	if issueIdx == -1 {
		return false, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "issue not found")
	}
	suggIdx := findSuggestionIndex(domain.Issues[issueIdx], suggestionName)
	if suggIdx == -1 {
		return false, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "suggestion not found")
	}

	store := ctx.KVStore(k.StoreKey)
	voteKey := deleteVoteKey(domainName, issueName, suggestionName, memberAddr)

	// Check for duplicate vote.
	if store.Has(voteKey) {
		return false, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "already voted to delete this suggestion")
	}

	// Record vote.
	store.Set(voteKey, []byte{1})
	domain.Issues[issueIdx].Suggestions[suggIdx].DeleteVotes++

	totalMembers := len(domain.Members)
	votes := domain.Issues[issueIdx].Suggestions[suggIdx].DeleteVotes

	// Check if 2/3 majority reached.
	deleted := int64(votes)*10000 >= int64(totalMembers)*DeleteMajorityBps
	if deleted {
		// Remove the suggestion.
		suggs := domain.Issues[issueIdx].Suggestions
		domain.Issues[issueIdx].Suggestions = append(suggs[:suggIdx], suggs[suggIdx+1:]...)
	}

	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	store.Set([]byte("domain:"+domainName), bz)
	return deleted, nil
}
