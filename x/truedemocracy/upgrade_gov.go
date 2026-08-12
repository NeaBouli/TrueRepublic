package truedemocracy

import (
	"fmt"
	"sort"

	errorsmod "cosmossdk.io/errors"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ReservedGovernanceDomain is the chain-wide governance domain. It can only
// be anchored in genesis; runtime transaction-based creation of this exact
// name is rejected so no account can claim the software-upgrade electorate.
const ReservedGovernanceDomain = "governance"

// Software-upgrade governance bounds (GH-184).
const (
	// UpgradePlanNameMaxLength bounds the plan name.
	UpgradePlanNameMaxLength = 64
	// UpgradePlanInfoMaxBytes bounds the free-form plan info.
	UpgradePlanInfoMaxBytes = 4096
	// UpgradeMinLeadBlocks is the minimum number of blocks between the current
	// height and a schedulable upgrade height.
	UpgradeMinLeadBlocks int64 = 10
)

// SoftwareUpgradeProposal is the single active governance upgrade record. The
// eligible member snapshot is taken on the first valid vote and never changes
// afterwards, so later membership changes cannot alter eligibility or the
// threshold. The record (and its votes) is preserved after scheduling so a
// cancellation vote authenticates against the original snapshot.
type SoftwareUpgradeProposal struct {
	Name      string   `json:"name"`
	Height    int64    `json:"height"`
	Info      string   `json:"info"`
	Eligible  []string `json:"eligible"` // sorted, deduplicated, non-empty snapshot
	Scheduled bool     `json:"scheduled"`
}

// SoftwareUpgradeCancelProposal tracks cancellation votes bound to the exact
// scheduled plan name and the original eligibility snapshot.
type SoftwareUpgradeCancelProposal struct {
	Name     string   `json:"name"`
	Eligible []string `json:"eligible"` // copied from the scheduled proposal
}

// KV layout:
//
//	"upgradegov:proposal"             → SoftwareUpgradeProposal
//	"upgradegov:vote:{voter}"         → []byte{1}
//	"upgradegov:cancel:proposal"      → SoftwareUpgradeCancelProposal
//	"upgradegov:cancel:vote:{voter}"  → []byte{1}

var softwareUpgradeProposalKey = []byte("upgradegov:proposal")
var softwareUpgradeCancelProposalKey = []byte("upgradegov:cancel:proposal")

func softwareUpgradeVoteKey(voter string) []byte {
	return []byte("upgradegov:vote:" + voter)
}

func softwareUpgradeCancelVoteKey(voter string) []byte {
	return []byte("upgradegov:cancel:vote:" + voter)
}

// validateUpgradePlanName enforces the bounded plan-name alphabet: 1..64
// chars, first character a lowercase ASCII alphanumeric, remaining characters
// lowercase ASCII alphanumeric, '.', '_' or '-'.
func validateUpgradePlanName(name string) error {
	if len(name) == 0 || len(name) > UpgradePlanNameMaxLength {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "plan name must be 1..%d characters", UpgradePlanNameMaxLength)
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		lowerAlnum := (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')
		if i == 0 {
			if !lowerAlnum {
				return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "plan name must start with a lowercase alphanumeric character")
			}
			continue
		}
		if !lowerAlnum && c != '.' && c != '_' && c != '-' {
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "plan name may only contain lowercase alphanumeric characters, '.', '_' or '-'")
		}
	}
	return nil
}

// validateUpgradePlan checks name, info bound, and the minimum lead time
// against the current block height.
func validateUpgradePlan(ctx sdk.Context, name string, height int64, info string) error {
	if err := validateUpgradePlanName(name); err != nil {
		return err
	}
	if height < ctx.BlockHeight()+UpgradeMinLeadBlocks {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"plan height %d is too soon or in the past (current %d, minimum lead %d blocks)",
			height, ctx.BlockHeight(), UpgradeMinLeadBlocks,
		)
	}
	if len(info) > UpgradePlanInfoMaxBytes {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "plan info exceeds %d bytes", UpgradePlanInfoMaxBytes)
	}
	return nil
}

// validateRestoredUpgradePlan validates an already-approved pending plan at
// genesis import. The original lead-time vote was enforced when the plan was
// scheduled; import may legitimately happen inside that window, but never at
// or after the planned height.
func validateRestoredUpgradePlan(ctx sdk.Context, proposal SoftwareUpgradeProposal) error {
	if err := validateUpgradePlanName(proposal.Name); err != nil {
		return err
	}
	if proposal.Height <= ctx.BlockHeight() {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"restored plan height %d must be after current height %d",
			proposal.Height, ctx.BlockHeight(),
		)
	}
	if len(proposal.Info) > UpgradePlanInfoMaxBytes {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "plan info exceeds %d bytes", UpgradePlanInfoMaxBytes)
	}
	return nil
}

// snapshotEligibleMembers returns the sorted, deduplicated member list used
// as the immutable eligibility snapshot.
func snapshotEligibleMembers(domain Domain) []string {
	eligible := append([]string(nil), domain.Members...)
	sort.Strings(eligible)
	out := eligible[:0]
	for i, member := range eligible {
		if i > 0 && member == eligible[i-1] {
			continue
		}
		out = append(out, member)
	}
	return out
}

func isEligibleSnapshotMember(eligible []string, addr string) bool {
	idx := sort.SearchStrings(eligible, addr)
	return idx < len(eligible) && eligible[idx] == addr
}

// GetSoftwareUpgradeProposal returns the active governance upgrade proposal,
// including a scheduled one retained for cancellation authentication.
func (k Keeper) GetSoftwareUpgradeProposal(ctx sdk.Context) (SoftwareUpgradeProposal, bool) {
	store := ctx.KVStore(k.StoreKey)
	bz := store.Get(softwareUpgradeProposalKey)
	if bz == nil {
		return SoftwareUpgradeProposal{}, false
	}
	var proposal SoftwareUpgradeProposal
	k.cdc.MustUnmarshalLengthPrefixed(bz, &proposal)
	return proposal, true
}

func (k Keeper) setSoftwareUpgradeProposal(ctx sdk.Context, proposal SoftwareUpgradeProposal) {
	store := ctx.KVStore(k.StoreKey)
	store.Set(softwareUpgradeProposalKey, k.cdc.MustMarshalLengthPrefixed(&proposal))
}

// GetSoftwareUpgradeCancelProposal returns the active cancellation record.
func (k Keeper) GetSoftwareUpgradeCancelProposal(ctx sdk.Context) (SoftwareUpgradeCancelProposal, bool) {
	store := ctx.KVStore(k.StoreKey)
	bz := store.Get(softwareUpgradeCancelProposalKey)
	if bz == nil {
		return SoftwareUpgradeCancelProposal{}, false
	}
	var proposal SoftwareUpgradeCancelProposal
	k.cdc.MustUnmarshalLengthPrefixed(bz, &proposal)
	return proposal, true
}

// HasSoftwareUpgradeVote reports whether the voter already voted for the
// active upgrade proposal.
func (k Keeper) HasSoftwareUpgradeVote(ctx sdk.Context, voter string) bool {
	return ctx.KVStore(k.StoreKey).Has(softwareUpgradeVoteKey(voter))
}

// HasSoftwareUpgradeCancelVote reports whether the voter already voted to
// cancel the scheduled plan.
func (k Keeper) HasSoftwareUpgradeCancelVote(ctx sdk.Context, voter string) bool {
	return ctx.KVStore(k.StoreKey).Has(softwareUpgradeCancelVoteKey(voter))
}

// countSoftwareUpgradeVotes counts recorded schedule votes over the snapshot.
func (k Keeper) countSoftwareUpgradeVotes(ctx sdk.Context, eligible []string) int {
	store := ctx.KVStore(k.StoreKey)
	votes := 0
	for _, member := range eligible {
		if store.Has(softwareUpgradeVoteKey(member)) {
			votes++
		}
	}
	return votes
}

func (k Keeper) countSoftwareUpgradeCancelVotes(ctx sdk.Context, eligible []string) int {
	store := ctx.KVStore(k.StoreKey)
	votes := 0
	for _, member := range eligible {
		if store.Has(softwareUpgradeCancelVoteKey(member)) {
			votes++
		}
	}
	return votes
}

// ExportSoftwareUpgradeGovernance returns only an actually pending scheduled
// plan and its authenticated votes. A completed proposal retained solely for
// cancellation cleanup must not be resurrected by genesis export/import.
func (k Keeper) ExportSoftwareUpgradeGovernance(ctx sdk.Context) (*SoftwareUpgradeProposal, []string, *SoftwareUpgradeCancelProposal, []string) {
	proposal, found := k.GetSoftwareUpgradeProposal(ctx)
	if !found || !proposal.Scheduled || k.upgradeScheduler == nil {
		return nil, nil, nil, nil
	}
	plan, err := k.upgradeScheduler.GetUpgradePlan(ctx)
	if err != nil || plan.Name != proposal.Name || plan.Height != proposal.Height || plan.Info != proposal.Info {
		return nil, nil, nil, nil
	}
	votes := make([]string, 0, len(proposal.Eligible))
	for _, member := range proposal.Eligible {
		if k.HasSoftwareUpgradeVote(ctx, member) {
			votes = append(votes, member)
		}
	}
	cancelProposal, cancelFound := k.GetSoftwareUpgradeCancelProposal(ctx)
	var cancelPtr *SoftwareUpgradeCancelProposal
	var cancelVotes []string
	if cancelFound {
		cancelCopy := cancelProposal
		cancelPtr = &cancelCopy
		for _, member := range cancelProposal.Eligible {
			if k.HasSoftwareUpgradeCancelVote(ctx, member) {
				cancelVotes = append(cancelVotes, member)
			}
		}
	}
	proposalCopy := proposal
	return &proposalCopy, votes, cancelPtr, cancelVotes
}

// InitSoftwareUpgradeGovernance restores a pending scheduled plan atomically.
// The exported electorate must exactly match the immutable genesis Domain.
func (k Keeper) InitSoftwareUpgradeGovernance(ctx sdk.Context, proposal *SoftwareUpgradeProposal, votes []string, cancelProposal *SoftwareUpgradeCancelProposal, cancelVotes []string) error {
	if proposal == nil {
		if len(votes) != 0 || cancelProposal != nil || len(cancelVotes) != 0 {
			return fmt.Errorf("software-upgrade votes require a proposal")
		}
		return nil
	}
	if k.upgradeScheduler == nil || !proposal.Scheduled {
		return fmt.Errorf("scheduled software-upgrade proposal requires an upgrade scheduler")
	}
	domain, found := k.GetDomain(ctx, ReservedGovernanceDomain)
	if !found {
		return fmt.Errorf("reserved governance domain not found")
	}
	wantEligible := snapshotEligibleMembers(domain)
	if len(wantEligible) == 0 || len(wantEligible) != len(proposal.Eligible) {
		return fmt.Errorf("software-upgrade electorate does not match genesis governance domain")
	}
	for i := range wantEligible {
		if wantEligible[i] != proposal.Eligible[i] {
			return fmt.Errorf("software-upgrade electorate does not match genesis governance domain")
		}
	}
	if err := validateRestoredUpgradePlan(ctx, *proposal); err != nil {
		return err
	}
	if !upgradeThresholdReached(len(votes), len(proposal.Eligible)) {
		return fmt.Errorf("scheduled software-upgrade proposal lacks two-thirds vote evidence")
	}
	cacheCtx, write := ctx.CacheContext()
	k.setSoftwareUpgradeProposal(cacheCtx, *proposal)
	seen := make(map[string]struct{}, len(votes))
	for _, voter := range votes {
		if !isEligibleSnapshotMember(proposal.Eligible, voter) {
			return fmt.Errorf("software-upgrade voter is outside the electorate")
		}
		if _, duplicate := seen[voter]; duplicate {
			return fmt.Errorf("duplicate software-upgrade voter")
		}
		seen[voter] = struct{}{}
		cacheCtx.KVStore(k.StoreKey).Set(softwareUpgradeVoteKey(voter), []byte{1})
	}
	if cancelProposal != nil {
		if cancelProposal.Name != proposal.Name || len(cancelProposal.Eligible) != len(proposal.Eligible) {
			return fmt.Errorf("software-upgrade cancellation does not match proposal")
		}
		for i := range proposal.Eligible {
			if cancelProposal.Eligible[i] != proposal.Eligible[i] {
				return fmt.Errorf("software-upgrade cancellation electorate mismatch")
			}
		}
		cacheCtx.KVStore(k.StoreKey).Set(softwareUpgradeCancelProposalKey, k.cdc.MustMarshalLengthPrefixed(cancelProposal))
		seen = make(map[string]struct{}, len(cancelVotes))
		for _, voter := range cancelVotes {
			if !isEligibleSnapshotMember(proposal.Eligible, voter) {
				return fmt.Errorf("software-upgrade cancellation voter is outside the electorate")
			}
			if _, duplicate := seen[voter]; duplicate {
				return fmt.Errorf("duplicate software-upgrade cancellation voter")
			}
			seen[voter] = struct{}{}
			cacheCtx.KVStore(k.StoreKey).Set(softwareUpgradeCancelVoteKey(voter), []byte{1})
		}
	} else if len(cancelVotes) != 0 {
		return fmt.Errorf("software-upgrade cancellation votes require a cancellation proposal")
	}
	plan := upgradetypes.Plan{Name: proposal.Name, Height: proposal.Height, Info: proposal.Info}
	if err := k.upgradeScheduler.ScheduleUpgrade(cacheCtx, plan); err != nil {
		return fmt.Errorf("restore software-upgrade plan: %w", err)
	}
	write()
	return nil
}

// clearSoftwareUpgradeGovernance removes the proposal, the cancel proposal,
// and every recorded vote over the snapshot.
func (k Keeper) clearSoftwareUpgradeGovernance(ctx sdk.Context, eligible []string) {
	store := ctx.KVStore(k.StoreKey)
	store.Delete(softwareUpgradeProposalKey)
	store.Delete(softwareUpgradeCancelProposalKey)
	for _, member := range eligible {
		store.Delete(softwareUpgradeVoteKey(member))
		store.Delete(softwareUpgradeCancelVoteKey(member))
	}
}

// upgradeThresholdReached reports the exact 2/3 ceiling: votes*3 >= eligible*2.
func upgradeThresholdReached(votes, eligible int) bool {
	return eligible > 0 && int64(votes)*3 >= int64(eligible)*2
}

// VoteSoftwareUpgrade records an authenticated member's vote for an exact
// software-upgrade plan and schedules it through the UpgradeScheduler once
// votes reach two thirds of the snapshotted electorate. The first valid vote
// requires the reserved genesis domain to exist and the sender to be a
// current member, then snapshots a sorted, deduplicated, non-empty member
// list; later votes authenticate against that snapshot only. At most one
// governance plan is active: conflicting names, heights, or info fail closed.
// Vote state and the scheduler write commit atomically through a cache
// context, so a scheduler failure rolls the vote back.
func (k Keeper) VoteSoftwareUpgrade(
	ctx sdk.Context,
	sender sdk.AccAddress,
	name string,
	height int64,
	info string,
) (votes, eligibleCount int, scheduled bool, err error) {
	if k.upgradeScheduler == nil {
		return 0, 0, false, errorsmod.Wrap(sdkerrors.ErrLogic, "upgrade scheduler not available")
	}
	if sender.Empty() {
		return 0, 0, false, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "sender address is required")
	}
	if err := validateUpgradePlan(ctx, name, height, info); err != nil {
		return 0, 0, false, err
	}

	domain, found := k.GetDomain(ctx, ReservedGovernanceDomain)
	if !found {
		return 0, 0, false, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "reserved domain %s not found", ReservedGovernanceDomain)
	}

	cacheCtx, write := ctx.CacheContext()

	proposal, exists := k.GetSoftwareUpgradeProposal(cacheCtx)
	if !exists {
		// First valid vote: the sender must be a current member, then the
		// electorate is snapshotted once and never re-derived.
		if !isMember(domain, sender.String()) {
			return 0, 0, false, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "only governance domain members can vote for software upgrades")
		}
		eligible := snapshotEligibleMembers(domain)
		if len(eligible) == 0 {
			return 0, 0, false, errorsmod.Wrap(sdkerrors.ErrLogic, "governance domain has no members")
		}
		proposal = SoftwareUpgradeProposal{
			Name:     name,
			Height:   height,
			Info:     info,
			Eligible: eligible,
		}
	} else {
		// One active plan at a time: later votes must match the exact plan.
		if proposal.Name != name || proposal.Height != height || proposal.Info != info {
			return 0, 0, false, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "conflicting software-upgrade plan; only one governance plan can be active")
		}
		if proposal.Scheduled {
			return 0, 0, false, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "software-upgrade plan is already scheduled")
		}
		if !isEligibleSnapshotMember(proposal.Eligible, sender.String()) {
			return 0, 0, false, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "sender is not in the upgrade eligibility snapshot")
		}
	}

	if k.HasSoftwareUpgradeVote(cacheCtx, sender.String()) {
		return 0, 0, false, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "sender already voted for this software-upgrade plan")
	}
	cacheCtx.KVStore(k.StoreKey).Set(softwareUpgradeVoteKey(sender.String()), []byte{1})

	votes = k.countSoftwareUpgradeVotes(cacheCtx, proposal.Eligible)
	eligibleCount = len(proposal.Eligible)

	if upgradeThresholdReached(votes, eligibleCount) {
		plan := upgradetypes.Plan{Name: proposal.Name, Height: proposal.Height, Info: proposal.Info}
		if err := k.upgradeScheduler.ScheduleUpgrade(cacheCtx, plan); err != nil {
			// The cache is dropped: neither the vote nor the proposal persists.
			return 0, 0, false, errorsmod.Wrap(err, "software-upgrade scheduling failed")
		}
		proposal.Scheduled = true
		scheduled = true
	}
	k.setSoftwareUpgradeProposal(cacheCtx, proposal)
	write()
	return votes, eligibleCount, scheduled, nil
}

// VoteCancelSoftwareUpgrade records an authenticated snapshot member's vote
// to cancel the exact active plan, including an unscheduled proposal whose
// voting window expired. Only members of the original
// eligibility snapshot may vote, and the name must match the scheduled plan.
// At two thirds the plan is cleared through the UpgradeScheduler and all
// governance state is removed atomically; a scheduler failure rolls back.
func (k Keeper) VoteCancelSoftwareUpgrade(
	ctx sdk.Context,
	sender sdk.AccAddress,
	name string,
) (votes, eligibleCount int, cancelled bool, err error) {
	if k.upgradeScheduler == nil {
		return 0, 0, false, errorsmod.Wrap(sdkerrors.ErrLogic, "upgrade scheduler not available")
	}
	if sender.Empty() {
		return 0, 0, false, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "sender address is required")
	}
	if err := validateUpgradePlanName(name); err != nil {
		return 0, 0, false, err
	}

	proposal, exists := k.GetSoftwareUpgradeProposal(ctx)
	if !exists {
		return 0, 0, false, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "no software-upgrade plan to cancel")
	}
	if proposal.Name != name {
		return 0, 0, false, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "cancel vote does not match the scheduled plan name")
	}

	cacheCtx, write := ctx.CacheContext()

	cancelProposal, cancelExists := k.GetSoftwareUpgradeCancelProposal(cacheCtx)
	if !cancelExists {
		cancelProposal = SoftwareUpgradeCancelProposal{
			Name:     proposal.Name,
			Eligible: append([]string(nil), proposal.Eligible...),
		}
	} else if cancelProposal.Name != proposal.Name {
		return 0, 0, false, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "conflicting software-upgrade cancellation")
	}

	if !isEligibleSnapshotMember(cancelProposal.Eligible, sender.String()) {
		return 0, 0, false, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "sender is not in the upgrade eligibility snapshot")
	}
	if k.HasSoftwareUpgradeCancelVote(cacheCtx, sender.String()) {
		return 0, 0, false, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "sender already voted to cancel this software-upgrade plan")
	}
	cacheCtx.KVStore(k.StoreKey).Set(softwareUpgradeCancelVoteKey(sender.String()), []byte{1})

	votes = k.countSoftwareUpgradeCancelVotes(cacheCtx, cancelProposal.Eligible)
	eligibleCount = len(cancelProposal.Eligible)

	if upgradeThresholdReached(votes, eligibleCount) {
		if proposal.Scheduled {
			if err := k.upgradeScheduler.ClearUpgradePlan(cacheCtx); err != nil {
				// The cache is dropped: the cancel vote and record roll back.
				return 0, 0, false, errorsmod.Wrap(err, "software-upgrade cancellation failed")
			}
		}
		k.clearSoftwareUpgradeGovernance(cacheCtx, cancelProposal.Eligible)
		cancelled = true
	} else {
		cacheCtx.KVStore(k.StoreKey).Set(
			softwareUpgradeCancelProposalKey,
			k.cdc.MustMarshalLengthPrefixed(&cancelProposal),
		)
	}
	write()
	return votes, eligibleCount, cancelled, nil
}
