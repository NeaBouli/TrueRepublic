package truedemocracy

import (
	"errors"
	"fmt"
	"sort"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Member stone tracking uses KV keys:
//   "stone:m:{domainName}:{voterAddr}" → target member address
//
// Member exclusion votes use:
//   "exclvote:{domainName}:{targetMember}:{voterAddr}" → []byte{1}

func memberStoneKey(domainName, voterAddr string) []byte {
	return []byte("stone:m:" + domainName + ":" + voterAddr)
}

func excludeVoteKey(domainName, targetMember, voterAddr string) []byte {
	return []byte("exclvote:" + domainName + ":" + targetMember + ":" + voterAddr)
}

// --- Member Stone Voting (WP §3.6) ---

// PlaceStoneOnMember places (or moves) the voter's stone on a domain member.
// Each member has exactly one stone on the member list. Placing on a new
// target automatically moves it from the old one. Members cannot vote for
// themselves.
func (k Keeper) PlaceStoneOnMember(ctx sdk.Context, domainName, targetMember, voterAddr string) error {
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "domain %s not found", domainName)
	}

	if !isMember(domain, voterAddr) {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "only domain members can place stones")
	}

	if !isMember(domain, targetMember) {
		return errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "target is not a domain member")
	}

	if voterAddr == targetMember {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "cannot place stone on yourself")
	}

	store := ctx.KVStore(k.StoreKey)
	key := memberStoneKey(domainName, voterAddr)

	if existing := store.Get(key); existing != nil {
		oldTarget := string(existing)
		if oldTarget == targetMember {
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "stone already placed on this member")
		}
	}

	store.Set(key, []byte(targetMember))
	return nil
}

// GetMemberStone returns who the voter's stone is placed on in the member list.
func (k Keeper) GetMemberStone(ctx sdk.Context, domainName, voterAddr string) (string, bool) {
	store := ctx.KVStore(k.StoreKey)
	bz := store.Get(memberStoneKey(domainName, voterAddr))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}

// MemberRank pairs a member address with their stone count.
type MemberRank struct {
	Address string
	Stones  int
}

// countMemberStones builds a map of member → stone count by checking each
// member's vote in the KV store. O(members) KV reads.
func (k Keeper) countMemberStones(ctx sdk.Context, domain Domain) map[string]int {
	counts := make(map[string]int)
	for _, member := range domain.Members {
		target, found := k.GetMemberStone(ctx, domain.Name, member)
		if found {
			counts[target]++
		}
	}
	return counts
}

// SortMembersByStones returns members sorted by stone count descending.
// Ties are broken by original order (stable sort).
func (k Keeper) SortMembersByStones(ctx sdk.Context, domain Domain) []MemberRank {
	counts := k.countMemberStones(ctx, domain)
	ranks := make([]MemberRank, len(domain.Members))
	for i, addr := range domain.Members {
		ranks[i] = MemberRank{Address: addr, Stones: counts[addr]}
	}
	sort.SliceStable(ranks, func(i, j int) bool {
		return ranks[i].Stones > ranks[j].Stones
	})
	return ranks
}

// --- Admin Election (WP §3.6) ---

// EventTypeDomainAdminQuarantine is emitted by ProcessGovernance when a domain
// is quarantined because its stored admin failed the read-only integrity
// classifier (GH-304). Attributes: "domain" and "reason" (a stable
// DomainAdminReason* code).
const EventTypeDomainAdminQuarantine = "domain_admin_quarantine"

// domainAdminIntegrityError fails an election closed when the domain's stored
// current admin does not pass the read-only integrity classifier (GH-304).
// Unwrap exposes sdkerrors.ErrInvalidAddress so callers keep deterministic
// invalid-address semantics, while errors.As lets ProcessGovernance quarantine
// exactly this failure domain-locally instead of halting the chain — unrelated
// invalid-address errors still abort the whole pass.
type domainAdminIntegrityError struct {
	DomainName string
	Reason     string
}

func (e *domainAdminIntegrityError) Error() string {
	return fmt.Sprintf("domain %s stored admin failed integrity classification: %s", e.DomainName, e.Reason)
}

func (e *domainAdminIntegrityError) Unwrap() error { return sdkerrors.ErrInvalidAddress }

// ElectAdmin sets the domain admin to the member with the most stones when
// AdminElectable is true. If no valid member has stones, admin remains
// unchanged. Members are stored as canonical bech32 strings while Admin holds
// the decoded address bytes, so the winner is compared against Admin.String()
// and decoded with sdk.AccAddressFromBech32 before any state mutation (GH-304).
// Legacy member strings that are not parseable bech32 are ignored while
// ranking candidates so they can never halt the chain. A stored current admin
// that fails the read-only integrity classifier fails the election closed with
// a deterministic domainAdminIntegrityError before any mutation — elections
// never silently repair corrupted admin state (GH-304).
func (k Keeper) ElectAdmin(ctx sdk.Context, domainName string) error {
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return nil
	}

	if !domain.Options.AdminElectable {
		return nil
	}

	// Fail closed before any state change when the stored current admin is
	// corrupt; repair/migration is a separate explicitly approved action.
	if reason := classifyDomainAdmin(domain); reason != "" {
		return &domainAdminIntegrityError{DomainName: domainName, Reason: reason}
	}

	counts := k.countMemberStones(ctx, domain)
	if len(counts) == 0 {
		return nil // no stones placed yet
	}

	// Find valid member with most stones. Stable: first member with max wins.
	// Unparseable legacy member entries can never become admin and are
	// ignored so they cannot halt the chain (GH-304).
	bestAddr := ""
	bestCount := 0
	for _, member := range domain.Members {
		if _, err := sdk.AccAddressFromBech32(member); err != nil {
			continue
		}
		if counts[member] > bestCount {
			bestCount = counts[member]
			bestAddr = member
		}
	}

	if bestAddr == "" || bestAddr == domain.Admin.String() {
		return nil // no change
	}

	// Defense in depth: the winner was already parsed while ranking, but the
	// address is decoded again immediately before the mutation.
	newAdmin, err := sdk.AccAddressFromBech32(bestAddr)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "domain %s elected admin member %q is not a valid bech32 address", domainName, bestAddr)
	}

	domain.Admin = newAdmin
	store := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	store.Set([]byte("domain:"+domainName), bz)
	return nil
}

// --- Member Exclusion by 2/3 Vote (WP §3.6) ---

// VoteToExclude records a vote to exclude a member from the domain. When 2/3
// of members have voted, the target is removed from the member list and their
// stones are cleaned up. Returns (excluded bool, error).
// The current domain admin can never be excluded: the vote is rejected with a
// deterministic invalid-request error before any vote key or other state is
// written, so normal governance cannot create admin_not_in_members/quarantined
// state. A replacement admin must be elected first; the former admin is then
// an ordinary member and excludable under the same threshold (GH-304).
func (k Keeper) VoteToExclude(ctx sdk.Context, domainName, targetMember, voterAddr string) (bool, error) {
	if domainName == ReservedGovernanceDomain {
		return false, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "governance electorate is immutable after genesis")
	}
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return false, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "domain %s not found", domainName)
	}

	if !isMember(domain, voterAddr) {
		return false, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "only domain members can vote to exclude")
	}

	if !isMember(domain, targetMember) {
		return false, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "target is not a domain member")
	}

	if voterAddr == targetMember {
		return false, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "cannot vote to exclude yourself")
	}

	// The admin ∈ Members invariant (GH-304): normal governance must never
	// exclude the current domain admin. Reject before any vote key or domain
	// state is written; a replacement admin must be elected first.
	if targetMember == domain.Admin.String() {
		return false, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "cannot exclude the current domain admin: elect a replacement admin first")
	}

	store := ctx.KVStore(k.StoreKey)
	voteKey := excludeVoteKey(domainName, targetMember, voterAddr)

	if store.Has(voteKey) {
		return false, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "already voted to exclude this member")
	}

	store.Set(voteKey, []byte{1})

	// Count total votes for this exclusion.
	votes := 0
	for _, member := range domain.Members {
		if member == targetMember {
			continue // target doesn't count
		}
		if store.Has(excludeVoteKey(domainName, targetMember, member)) {
			votes++
		}
	}

	totalVoters := len(domain.Members) - 1 // exclude the target from the denominator
	excluded := int64(votes)*10000 >= int64(totalVoters)*ExcludeMajorityBps

	if excluded {
		k.removeMember(ctx, &domain, targetMember)

		bz := k.cdc.MustMarshalLengthPrefixed(&domain)
		store.Set([]byte("domain:"+domainName), bz)
	}

	return excluded, nil
}

// removeMember removes a member from the domain and cleans up their stones.
func (k Keeper) removeMember(ctx sdk.Context, domain *Domain, memberAddr string) {
	// Remove from member list.
	for i, m := range domain.Members {
		if m == memberAddr {
			domain.Members = append(domain.Members[:i], domain.Members[i+1:]...)
			break
		}
	}

	store := ctx.KVStore(k.StoreKey)

	// Clean up issue stone.
	issueKey := issueStoneKey(domain.Name, memberAddr)
	if oldIssue := store.Get(issueKey); oldIssue != nil {
		issueName := string(oldIssue)
		for i, issue := range domain.Issues {
			if issue.Name == issueName && domain.Issues[i].Stones > 0 {
				domain.Issues[i].Stones--
				break
			}
		}
		store.Delete(issueKey)
	}

	// Clean up suggestion stones (one per issue).
	for _, issue := range domain.Issues {
		suggKey := suggestionStoneKey(domain.Name, issue.Name, memberAddr)
		if oldSugg := store.Get(suggKey); oldSugg != nil {
			suggName := string(oldSugg)
			for i, s := range issue.Suggestions {
				if s.Name == suggName {
					// Find the issue index to modify the domain's data.
					for ii, di := range domain.Issues {
						if di.Name == issue.Name && ii < len(domain.Issues) {
							for jj, ds := range domain.Issues[ii].Suggestions {
								if ds.Name == suggName && domain.Issues[ii].Suggestions[jj].Stones > 0 {
									domain.Issues[ii].Suggestions[jj].Stones--
								}
							}
						}
					}
					_ = i
					break
				}
			}
			store.Delete(suggKey)
		}
	}

	// Clean up member stone (who they voted for).
	store.Delete(memberStoneKey(domain.Name, memberAddr))
}

// --- Inactivity Cleanup (WP §3.1) ---

// CleanupInactiveIssues removes issues (and their suggestions) that have had
// no activity for InactivityTimeoutSecs (360 days). Issues with
// LastActivityAt == 0 use CreationDate as fallback.
func (k Keeper) CleanupInactiveIssues(ctx sdk.Context, domainName string) error {
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return nil
	}

	now := ctx.BlockTime().Unix()
	var kept []Issue
	modified := false

	for _, issue := range domain.Issues {
		lastActivity := issue.LastActivityAt
		if lastActivity == 0 {
			lastActivity = issue.CreationDate
		}
		if lastActivity > 0 && now-lastActivity > InactivityTimeoutSecs {
			modified = true
			continue // drop this issue
		}
		kept = append(kept, issue)
	}

	if modified {
		domain.Issues = kept
		store := ctx.KVStore(k.StoreKey)
		bz := k.cdc.MustMarshalLengthPrefixed(&domain)
		store.Set([]byte("domain:"+domainName), bz)
	}
	return nil
}

// --- PoD Transfer Limit (WP §7) ---

// TrackPayout increments a domain's cumulative TotalPayouts. This is used to
// calculate the 10% stake transfer limit for validators.
func (k Keeper) TrackPayout(ctx sdk.Context, domainName string, amount int64) error {
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "domain %s not found", domainName)
	}
	domain.TotalPayouts += amount
	store := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	store.Set([]byte("domain:"+domainName), bz)
	return nil
}

// ValidateStakeTransfer checks whether a validator can withdraw the given
// amount of stake from the domain without exceeding the 10% transfer limit
// (WP §7). The limit is: cumulative transferred stake ≤ 10% of domain's
// total payouts. If domain has zero payouts, transfers are blocked.
func (k Keeper) ValidateStakeTransfer(ctx sdk.Context, domainName, operatorAddr string, amount int64) error {
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "domain %s not found", domainName)
	}

	if domain.TotalPayouts == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "no payouts yet — stake transfers not allowed")
	}

	transferLimit := domain.TotalPayouts * StakeTransferLimitBps / 10000
	newTotal := domain.TransferredStake + amount
	if newTotal > transferLimit {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"transfer %d would exceed 10%% of domain payouts (limit %d, already transferred %d)",
			amount, transferLimit, domain.TransferredStake)
	}
	return nil
}

// ProcessGovernance runs admin election and inactivity cleanup for all domains.
// Called from EndBlock. All election and cleanup work runs in a single cache
// context and is committed only after every domain is processed (GH-304).
// A domain whose stored admin fails the integrity classifier is quarantined
// domain-locally: a stable event records the domain and reason, election and
// cleanup are skipped for that domain only, processing continues with the
// remaining domains, and the valid domains' work still commits atomically.
// The quarantine event is emitted on the cache context, so events and
// valid-domain state share the same atomic commit boundary: write() flushes
// both to the parent context, and any future non-quarantine abort discards
// both. Any other error aborts the pass without committing, so EndBlock fails
// closed instead of persisting partial governance work.
func (k Keeper) ProcessGovernance(ctx sdk.Context) error {
	var domainNames []string
	k.IterateDomains(ctx, func(d Domain) bool {
		domainNames = append(domainNames, d.Name)
		return false
	})

	cacheCtx, write := ctx.CacheContext()
	for _, name := range domainNames {
		if err := k.ElectAdmin(cacheCtx, name); err != nil {
			var integrityErr *domainAdminIntegrityError
			if errors.As(err, &integrityErr) {
				// Emit on the cache context: write() flushes the event together
				// with the valid domains' state, and an abort discards both.
				cacheCtx.EventManager().EmitEvent(sdk.NewEvent(
					EventTypeDomainAdminQuarantine,
					sdk.NewAttribute("domain", integrityErr.DomainName),
					sdk.NewAttribute("reason", integrityErr.Reason),
				))
				continue
			}
			return err
		}
		if err := k.CleanupInactiveIssues(cacheCtx, name); err != nil {
			return err
		}
	}
	write()
	return nil
}
