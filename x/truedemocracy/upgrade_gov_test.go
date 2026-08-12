package truedemocracy

import (
	"context"
	"errors"
	"strings"
	"testing"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type testUpgradeScheduler struct {
	plan         upgradetypes.Plan
	hasPlan      bool
	failSchedule bool
	failClear    bool
}

func (s *testUpgradeScheduler) ScheduleUpgrade(_ context.Context, plan upgradetypes.Plan) error {
	if s.failSchedule {
		return errors.New("injected schedule failure")
	}
	s.plan, s.hasPlan = plan, true
	return nil
}

func (s *testUpgradeScheduler) ClearUpgradePlan(_ context.Context) error {
	if s.failClear {
		return errors.New("injected clear failure")
	}
	s.plan, s.hasPlan = upgradetypes.Plan{}, false
	return nil
}

func (s *testUpgradeScheduler) GetUpgradePlan(_ context.Context) (upgradetypes.Plan, error) {
	if !s.hasPlan {
		return upgradetypes.Plan{}, upgradetypes.ErrNoUpgradePlanFound
	}
	return s.plan, nil
}

func upgradeMembers() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.AccAddress("upgrade-member-1"),
		sdk.AccAddress("upgrade-member-2"),
		sdk.AccAddress("upgrade-member-3"),
		sdk.AccAddress("upgrade-member-4"),
	}
}

func createUpgradeDomain(t *testing.T, k Keeper, ctx sdk.Context, members []sdk.AccAddress) {
	t.Helper()
	k.CreateDomain(ctx, ReservedGovernanceDomain, members[0], sdk.NewCoins())
	domain, _ := k.GetDomain(ctx, ReservedGovernanceDomain)
	for _, member := range members[1:] {
		domain.Members = append(domain.Members, member.String())
	}
	ctx.KVStore(k.StoreKey).Set([]byte("domain:"+ReservedGovernanceDomain), k.cdc.MustMarshalLengthPrefixed(&domain))
}

func TestSoftwareUpgradeGovernanceFailsClosedAndValidatesPlan(t *testing.T) {
	k, ctx := setupKeeper(t)
	ctx = ctx.WithBlockHeight(100)
	member := upgradeMembers()[0]
	if _, _, _, err := k.VoteSoftwareUpgrade(ctx, member, "v0.4.1", 110, "info"); err == nil {
		t.Fatal("missing scheduler accepted an upgrade vote")
	}
	scheduler := &testUpgradeScheduler{}
	k.upgradeScheduler = scheduler
	if _, _, _, err := k.VoteSoftwareUpgrade(ctx, member, "v0.4.1", 110, "info"); err == nil {
		t.Fatal("missing genesis governance domain accepted an upgrade vote")
	}
	createUpgradeDomain(t, k, ctx, []sdk.AccAddress{member})
	for _, tc := range []struct {
		name   string
		height int64
		info   string
	}{
		{"", 110, ""}, {"V0.4.1", 110, ""}, {"v/1", 110, ""},
		{strings.Repeat("a", UpgradePlanNameMaxLength+1), 110, ""},
		{"v0.4.1", 109, ""}, {"v0.4.1", 110, strings.Repeat("x", UpgradePlanInfoMaxBytes+1)},
	} {
		if _, _, _, err := k.VoteSoftwareUpgrade(ctx, member, tc.name, tc.height, tc.info); err == nil {
			t.Fatalf("invalid plan accepted: %+v", tc)
		}
	}
	if err := (MsgCreateDomain{Name: ReservedGovernanceDomain, Admin: member}).ValidateBasic(); err == nil {
		t.Fatal("runtime reserved governance-domain message validated")
	}
	bankKeeper, bankCtx, _ := setupKeeperWithBank(t)
	if err := bankKeeper.CreateDomainWithEscrow(bankCtx, ReservedGovernanceDomain, member, sdk.NewCoins()); err == nil {
		t.Fatal("keeper created the genesis-only governance domain at runtime")
	}
	if err := k.AddMember(ctx, ReservedGovernanceDomain, sdk.AccAddress("late-member").String(), member); err == nil {
		t.Fatal("keeper changed the immutable governance electorate")
	}
	if _, err := k.VoteToExclude(ctx, ReservedGovernanceDomain, member.String(), upgradeMembers()[1].String()); err == nil {
		t.Fatal("exclusion vote changed the immutable governance electorate")
	}
}

func TestSoftwareUpgradeThresholdConflictAndImmutableSnapshot(t *testing.T) {
	k, ctx := setupKeeper(t)
	ctx = ctx.WithBlockHeight(100)
	scheduler := &testUpgradeScheduler{}
	k.upgradeScheduler = scheduler
	members := upgradeMembers()
	createUpgradeDomain(t, k, ctx, members)

	for i := 0; i < 2; i++ {
		votes, eligible, scheduled, err := k.VoteSoftwareUpgrade(ctx, members[i], "v0.4.1", 120, "exact")
		if err != nil || votes != i+1 || eligible != 4 || scheduled {
			t.Fatalf("vote %d = %d/%d scheduled=%t err=%v", i+1, votes, eligible, scheduled, err)
		}
	}
	if _, _, _, err := k.VoteSoftwareUpgrade(ctx, members[0], "v0.4.1", 120, "exact"); err == nil {
		t.Fatal("duplicate vote accepted")
	}
	if _, _, _, err := k.VoteSoftwareUpgrade(ctx, members[2], "v0.4.2", 120, "exact"); err == nil {
		t.Fatal("conflicting plan accepted")
	}
	late := sdk.AccAddress("late-governance-member")
	if _, _, _, err := k.VoteSoftwareUpgrade(ctx, late, "v0.4.1", 120, "exact"); err == nil {
		t.Fatal("late member entered immutable electorate")
	}
	votes, eligible, scheduled, err := k.VoteSoftwareUpgrade(ctx, members[2], "v0.4.1", 120, "exact")
	if err != nil || votes != 3 || eligible != 4 || !scheduled {
		t.Fatalf("threshold vote = %d/%d scheduled=%t err=%v", votes, eligible, scheduled, err)
	}
	if !scheduler.hasPlan || scheduler.plan != (upgradetypes.Plan{Name: "v0.4.1", Height: 120, Info: "exact"}) {
		t.Fatalf("scheduled plan = %+v", scheduler.plan)
	}
	proposal, found := k.GetSoftwareUpgradeProposal(ctx)
	if !found || !proposal.Scheduled || len(proposal.Eligible) != 4 {
		t.Fatalf("retained proposal = %+v found=%t", proposal, found)
	}
}

func TestSoftwareUpgradeScheduleFailureRollsBackThresholdVote(t *testing.T) {
	k, ctx := setupKeeper(t)
	ctx = ctx.WithBlockHeight(100)
	scheduler := &testUpgradeScheduler{failSchedule: true}
	k.upgradeScheduler = scheduler
	members := upgradeMembers()[:2]
	createUpgradeDomain(t, k, ctx, members)
	if _, _, _, err := k.VoteSoftwareUpgrade(ctx, members[0], "v0.4.1", 120, ""); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := k.VoteSoftwareUpgrade(ctx, members[1], "v0.4.1", 120, ""); err == nil {
		t.Fatal("injected schedule failure was ignored")
	}
	if k.HasSoftwareUpgradeVote(ctx, members[1].String()) {
		t.Fatal("threshold-reaching vote persisted after scheduler failure")
	}
	proposal, found := k.GetSoftwareUpgradeProposal(ctx)
	if !found || proposal.Scheduled {
		t.Fatalf("proposal after failed schedule = %+v found=%t", proposal, found)
	}
	scheduler.failSchedule = false
	if _, _, scheduled, err := k.VoteSoftwareUpgrade(ctx, members[1], "v0.4.1", 120, ""); err != nil || !scheduled {
		t.Fatalf("fixed scheduler did not recover: scheduled=%t err=%v", scheduled, err)
	}
}

func TestSoftwareUpgradeCancellationUsesOriginalSnapshotAndIsAtomic(t *testing.T) {
	k, ctx := setupKeeper(t)
	ctx = ctx.WithBlockHeight(100)
	scheduler := &testUpgradeScheduler{}
	k.upgradeScheduler = scheduler
	members := upgradeMembers()[:3]
	createUpgradeDomain(t, k, ctx, members)
	for i := 0; i < 2; i++ {
		if _, _, _, err := k.VoteSoftwareUpgrade(ctx, members[i], "v0.4.1", 120, ""); err != nil {
			t.Fatal(err)
		}
	}
	late := sdk.AccAddress("late-cancel-member")
	if _, _, _, err := k.VoteCancelSoftwareUpgrade(ctx, late, "v0.4.1"); err == nil {
		t.Fatal("late member cancelled against a changed electorate")
	}
	if _, _, _, err := k.VoteCancelSoftwareUpgrade(ctx, members[0], "wrong-plan"); err == nil {
		t.Fatal("conflicting cancellation accepted")
	}
	if _, _, _, err := k.VoteCancelSoftwareUpgrade(ctx, members[0], "v0.4.1"); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := k.VoteCancelSoftwareUpgrade(ctx, members[0], "v0.4.1"); err == nil {
		t.Fatal("duplicate cancellation vote accepted")
	}
	scheduler.failClear = true
	if _, _, _, err := k.VoteCancelSoftwareUpgrade(ctx, members[1], "v0.4.1"); err == nil {
		t.Fatal("injected clear failure was ignored")
	}
	if k.HasSoftwareUpgradeCancelVote(ctx, members[1].String()) {
		t.Fatal("threshold cancellation vote persisted after clear failure")
	}
	if _, found := k.GetSoftwareUpgradeProposal(ctx); !found || !scheduler.hasPlan {
		t.Fatal("failed cancellation removed scheduled plan")
	}
	scheduler.failClear = false
	votes, eligible, cancelled, err := k.VoteCancelSoftwareUpgrade(ctx, members[1], "v0.4.1")
	if err != nil || votes != 2 || eligible != 3 || !cancelled {
		t.Fatalf("cancel threshold = %d/%d cancelled=%t err=%v", votes, eligible, cancelled, err)
	}
	if scheduler.hasPlan {
		t.Fatal("successful cancellation retained scheduler plan")
	}
	if _, found := k.GetSoftwareUpgradeProposal(ctx); found {
		t.Fatal("successful cancellation retained governance proposal")
	}
}

func TestSoftwareUpgradePostExecutionCleanupAllowsNextPlan(t *testing.T) {
	k, ctx := setupKeeper(t)
	ctx = ctx.WithBlockHeight(100)
	scheduler := &testUpgradeScheduler{}
	k.upgradeScheduler = scheduler
	members := upgradeMembers()[:3]
	createUpgradeDomain(t, k, ctx, members)
	for _, member := range members[:2] {
		if _, _, _, err := k.VoteSoftwareUpgrade(ctx, member, "v0.4.1", 120, ""); err != nil {
			t.Fatal(err)
		}
	}
	// x/upgrade clears the scheduler plan after successful execution, while the
	// governance record remains available to authenticate cleanup votes.
	scheduler.plan, scheduler.hasPlan = upgradetypes.Plan{}, false
	for _, member := range members[:2] {
		if _, _, _, err := k.VoteCancelSoftwareUpgrade(ctx, member, "v0.4.1"); err != nil {
			t.Fatal(err)
		}
	}
	if _, found := k.GetSoftwareUpgradeProposal(ctx); found {
		t.Fatal("post-execution cleanup retained the completed proposal")
	}
	if _, _, scheduled, err := k.VoteSoftwareUpgrade(ctx, members[0], "v0.4.2", 130, "next"); err != nil || scheduled {
		t.Fatalf("next plan first vote scheduled=%t err=%v", scheduled, err)
	}
}

func TestSoftwareUpgradeGovernanceExportImportPreservesPendingPlan(t *testing.T) {
	members := upgradeMembers()[:3]
	source, sourceCtx := setupKeeper(t)
	sourceCtx = sourceCtx.WithBlockHeight(100)
	sourceScheduler := &testUpgradeScheduler{}
	source.upgradeScheduler = sourceScheduler
	createUpgradeDomain(t, source, sourceCtx, members)
	for _, member := range members[:2] {
		if _, _, _, err := source.VoteSoftwareUpgrade(sourceCtx, member, "v0.4.1", 120, "round-trip"); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, _, err := source.VoteCancelSoftwareUpgrade(sourceCtx, members[0], "v0.4.1"); err != nil {
		t.Fatal(err)
	}
	proposal, votes, cancelProposal, cancelVotes := source.ExportSoftwareUpgradeGovernance(sourceCtx)
	if proposal == nil || cancelProposal == nil || len(votes) != 2 || len(cancelVotes) != 1 {
		t.Fatalf("export = proposal=%+v votes=%v cancel=%+v cancelVotes=%v", proposal, votes, cancelProposal, cancelVotes)
	}

	target, targetCtx := setupKeeper(t)
	// Import inside the original ten-block lead window must remain possible:
	// the plan was already approved and only needs to remain in the future.
	targetCtx = targetCtx.WithBlockHeight(115)
	targetScheduler := &testUpgradeScheduler{}
	target.upgradeScheduler = targetScheduler
	createUpgradeDomain(t, target, targetCtx, members)
	if err := target.InitSoftwareUpgradeGovernance(targetCtx, proposal, votes, cancelProposal, cancelVotes); err != nil {
		t.Fatal(err)
	}
	if !targetScheduler.hasPlan || targetScheduler.plan != sourceScheduler.plan {
		t.Fatalf("restored scheduler plan = %+v", targetScheduler.plan)
	}
	for _, voter := range votes {
		if !target.HasSoftwareUpgradeVote(targetCtx, voter) {
			t.Fatalf("missing restored schedule vote for %s", voter)
		}
	}
	if !target.HasSoftwareUpgradeCancelVote(targetCtx, cancelVotes[0]) {
		t.Fatal("missing restored cancellation vote")
	}
	lateTarget, lateCtx := setupKeeper(t)
	lateTarget.upgradeScheduler = &testUpgradeScheduler{}
	lateCtx = lateCtx.WithBlockHeight(proposal.Height)
	createUpgradeDomain(t, lateTarget, lateCtx, members)
	if err := lateTarget.InitSoftwareUpgradeGovernance(lateCtx, proposal, votes, cancelProposal, cancelVotes); err == nil {
		t.Fatal("restored plan at its execution height was accepted")
	}
}

func TestSoftwareUpgradeUnscheduledProposalCanBeCancelledAfterLeadWindow(t *testing.T) {
	k, ctx := setupKeeper(t)
	ctx = ctx.WithBlockHeight(100)
	scheduler := &testUpgradeScheduler{}
	k.upgradeScheduler = scheduler
	members := upgradeMembers()[:3]
	createUpgradeDomain(t, k, ctx, members)
	if _, _, scheduled, err := k.VoteSoftwareUpgrade(ctx, members[0], "v0.4.1", 120, ""); err != nil || scheduled {
		t.Fatalf("first vote scheduled=%t err=%v", scheduled, err)
	}
	ctx = ctx.WithBlockHeight(115)
	if _, _, _, err := k.VoteSoftwareUpgrade(ctx, members[1], "v0.4.1", 120, ""); err == nil {
		t.Fatal("late schedule vote unexpectedly passed the lead-time boundary")
	}
	for _, member := range members[:2] {
		if _, _, _, err := k.VoteCancelSoftwareUpgrade(ctx, member, "v0.4.1"); err != nil {
			t.Fatal(err)
		}
	}
	if _, found := k.GetSoftwareUpgradeProposal(ctx); found {
		t.Fatal("two-thirds cancellation retained an expired unscheduled proposal")
	}
}
