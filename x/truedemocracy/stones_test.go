package truedemocracy

import (
	"bytes"
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// setupDomainWithIssues creates a domain with members and multiple issues/suggestions for stone tests.
func setupDomainWithIssues(t *testing.T, k Keeper, ctx sdk.Context) {
	t.Helper()
	k.CreateDomain(ctx, "StonesDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 500_000)))

	domain, _ := k.GetDomain(ctx, "StonesDomain")
	domain.Members = append(domain.Members, "alice", "bob", "charlie")

	now := ctx.BlockTime().Unix()
	domain.Issues = []Issue{
		{
			Name: "Climate", Stones: 0, CreationDate: now,
			Suggestions: []Suggestion{
				{Name: "GreenDeal", Creator: "alice", Stones: 0, Ratings: []Rating{}, CreationDate: now},
				{Name: "CarbonTax", Creator: "bob", Stones: 0, Ratings: []Rating{}, CreationDate: now + 1},
			},
		},
		{
			Name: "Education", Stones: 0, CreationDate: now + 10,
			Suggestions: []Suggestion{
				{Name: "FreeTuition", Creator: "charlie", Stones: 0, Ratings: []Rating{}, CreationDate: now + 10},
			},
		},
		{
			Name: "Healthcare", Stones: 0, CreationDate: now + 20,
			Suggestions: []Suggestion{},
		},
	}

	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:StonesDomain"), bz)
}

// ---------- PlaceStoneOnIssue ----------

func TestPlaceStoneOnIssue(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupDomainWithIssues(t, k, ctx)

	t.Run("happy path", func(t *testing.T) {
		reward, err := k.PlaceStoneOnIssue(ctx, "StonesDomain", "Climate", "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reward.AmountOf(PNYXDenom).IsPositive() {
			t.Error("reward should be positive")
		}

		domain, _ := k.GetDomain(ctx, "StonesDomain")
		if domain.Issues[0].Stones != 1 {
			t.Errorf("Climate stones = %d, want 1", domain.Issues[0].Stones)
		}

		placed, found := k.GetMemberIssueStone(ctx, "StonesDomain", "alice")
		if !found || placed != "Climate" {
			t.Errorf("alice stone = %q, want 'Climate'", placed)
		}
	})

	t.Run("second member same issue", func(t *testing.T) {
		_, err := k.PlaceStoneOnIssue(ctx, "StonesDomain", "Climate", "bob")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		domain, _ := k.GetDomain(ctx, "StonesDomain")
		if domain.Issues[0].Stones != 2 {
			t.Errorf("Climate stones = %d, want 2", domain.Issues[0].Stones)
		}
	})
}

// ---------- StoneUniqueness ----------

func TestStoneUniqueness(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupDomainWithIssues(t, k, ctx)

	// Place stone.
	k.PlaceStoneOnIssue(ctx, "StonesDomain", "Climate", "alice")

	// Try to place again on same issue — should error.
	_, err := k.PlaceStoneOnIssue(ctx, "StonesDomain", "Climate", "alice")
	if err == nil {
		t.Fatal("expected error when placing stone on same issue twice")
	}
}

// ---------- MoveStone ----------

func TestMoveStone(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupDomainWithIssues(t, k, ctx)

	// Place stone on Climate.
	k.PlaceStoneOnIssue(ctx, "StonesDomain", "Climate", "alice")

	domain, _ := k.GetDomain(ctx, "StonesDomain")
	if domain.Issues[0].Stones != 1 {
		t.Fatalf("Climate stones = %d, want 1", domain.Issues[0].Stones)
	}

	// Move stone to Education (PlaceStone auto-moves).
	reward, err := k.PlaceStoneOnIssue(ctx, "StonesDomain", "Education", "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reward.Empty() {
		t.Errorf("moving stone must not earn another reward, got %s", reward)
	}

	domain, _ = k.GetDomain(ctx, "StonesDomain")
	if domain.Issues[0].Stones != 0 {
		t.Errorf("Climate stones = %d, want 0 after move", domain.Issues[0].Stones)
	}
	if domain.Issues[1].Stones != 1 {
		t.Errorf("Education stones = %d, want 1 after move", domain.Issues[1].Stones)
	}

	placed, _ := k.GetMemberIssueStone(ctx, "StonesDomain", "alice")
	if placed != "Education" {
		t.Errorf("alice stone = %q, want 'Education'", placed)
	}
}

// ---------- NonMemberReject ----------

func TestNonMemberRejectStone(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupDomainWithIssues(t, k, ctx)

	_, err := k.PlaceStoneOnIssue(ctx, "StonesDomain", "Climate", "outsider")
	if err == nil {
		t.Fatal("expected error for non-member stone placement")
	}
}

// ---------- Sorting ----------

func TestSortIssuesByStones(t *testing.T) {
	now := time.Now().Unix()
	issues := []Issue{
		{Name: "A", Stones: 1, CreationDate: now + 10},
		{Name: "B", Stones: 3, CreationDate: now + 20},
		{Name: "C", Stones: 1, CreationDate: now}, // same stones as A, but older
		{Name: "D", Stones: 0, CreationDate: now + 5},
	}

	sorted := SortIssuesByStones(issues)

	want := []string{"B", "C", "A", "D"}
	for i, name := range want {
		if sorted[i].Name != name {
			t.Errorf("sorted[%d] = %s, want %s", i, sorted[i].Name, name)
		}
	}

	// Verify original slice is not mutated.
	if issues[0].Name != "A" {
		t.Error("original slice was mutated")
	}
}

func TestSortSuggestionsByStones(t *testing.T) {
	now := time.Now().Unix()
	suggestions := []Suggestion{
		{Name: "X", Stones: 0, CreationDate: now},
		{Name: "Y", Stones: 2, CreationDate: now + 5},
		{Name: "Z", Stones: 2, CreationDate: now}, // same stones as Y, but older
	}

	sorted := SortSuggestionsByStones(suggestions)

	want := []string{"Z", "Y", "X"}
	for i, name := range want {
		if sorted[i].Name != name {
			t.Errorf("sorted[%d] = %s, want %s", i, sorted[i].Name, name)
		}
	}
}

// ---------- VoteToEarnReward ----------

func TestVoteToEarnReward(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupDomainWithIssues(t, k, ctx)

	domainBefore, _ := k.GetDomain(ctx, "StonesDomain")
	treasuryBefore := domainBefore.Treasury.AmountOf(PNYXDenom)

	// Expected reward = treasury / 1000.
	expectedReward := treasuryBefore.Quo(math.NewInt(1000))

	reward, err := k.PlaceStoneOnIssue(ctx, "StonesDomain", "Climate", "alice")
	if err != nil {
		t.Fatal(err)
	}
	if !reward.AmountOf(PNYXDenom).Equal(expectedReward) {
		t.Errorf("reward = %s, want %s", reward.AmountOf(PNYXDenom), expectedReward)
	}

	domainAfter, _ := k.GetDomain(ctx, "StonesDomain")
	treasuryAfter := domainAfter.Treasury.AmountOf(PNYXDenom)
	wantTreasury := treasuryBefore.Sub(expectedReward)
	if !treasuryAfter.Equal(wantTreasury) {
		t.Errorf("treasury = %s, want %s", treasuryAfter, wantTreasury)
	}
}

// ---------- PlaceStoneOnSuggestion ----------

func TestPlaceStoneOnSuggestion(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupDomainWithIssues(t, k, ctx)

	t.Run("happy path", func(t *testing.T) {
		reward, err := k.PlaceStoneOnSuggestion(ctx, "StonesDomain", "Climate", "GreenDeal", "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reward.AmountOf(PNYXDenom).IsPositive() {
			t.Error("reward should be positive")
		}

		domain, _ := k.GetDomain(ctx, "StonesDomain")
		if domain.Issues[0].Suggestions[0].Stones != 1 {
			t.Errorf("GreenDeal stones = %d, want 1", domain.Issues[0].Suggestions[0].Stones)
		}

		placed, found := k.GetMemberSuggestionStone(ctx, "StonesDomain", "Climate", "alice")
		if !found || placed != "GreenDeal" {
			t.Errorf("alice suggestion stone = %q, want 'GreenDeal'", placed)
		}
	})

	t.Run("move within suggestion list", func(t *testing.T) {
		// Alice moves stone from GreenDeal to CarbonTax.
		_, err := k.PlaceStoneOnSuggestion(ctx, "StonesDomain", "Climate", "CarbonTax", "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		domain, _ := k.GetDomain(ctx, "StonesDomain")
		if domain.Issues[0].Suggestions[0].Stones != 0 {
			t.Errorf("GreenDeal stones = %d, want 0 after move", domain.Issues[0].Suggestions[0].Stones)
		}
		if domain.Issues[0].Suggestions[1].Stones != 1 {
			t.Errorf("CarbonTax stones = %d, want 1 after move", domain.Issues[0].Suggestions[1].Stones)
		}
	})

	t.Run("duplicate rejected", func(t *testing.T) {
		_, err := k.PlaceStoneOnSuggestion(ctx, "StonesDomain", "Climate", "CarbonTax", "alice")
		if err == nil {
			t.Fatal("expected error for duplicate suggestion stone")
		}
	})

	t.Run("non-member rejected", func(t *testing.T) {
		_, err := k.PlaceStoneOnSuggestion(ctx, "StonesDomain", "Climate", "GreenDeal", "outsider")
		if err == nil {
			t.Fatal("expected error for non-member")
		}
	})

	t.Run("unknown suggestion rejected", func(t *testing.T) {
		_, err := k.PlaceStoneOnSuggestion(ctx, "StonesDomain", "Climate", "NoSuchSugg", "bob")
		if err == nil {
			t.Fatal("expected error for unknown suggestion")
		}
	})
}

// ---------- MultipleSuggestionLists ----------

func TestMultipleSuggestionLists(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupDomainWithIssues(t, k, ctx)

	// Alice places stone on GreenDeal (under Climate issue).
	_, err := k.PlaceStoneOnSuggestion(ctx, "StonesDomain", "Climate", "GreenDeal", "alice")
	if err != nil {
		t.Fatal(err)
	}

	// Alice places stone on FreeTuition (under Education issue) — separate list.
	_, err = k.PlaceStoneOnSuggestion(ctx, "StonesDomain", "Education", "FreeTuition", "alice")
	if err != nil {
		t.Fatal(err)
	}

	// Both should be independently placed.
	placed1, _ := k.GetMemberSuggestionStone(ctx, "StonesDomain", "Climate", "alice")
	placed2, _ := k.GetMemberSuggestionStone(ctx, "StonesDomain", "Education", "alice")

	if placed1 != "GreenDeal" {
		t.Errorf("Climate stone = %q, want 'GreenDeal'", placed1)
	}
	if placed2 != "FreeTuition" {
		t.Errorf("Education stone = %q, want 'FreeTuition'", placed2)
	}

	// Verify stone counts are independent.
	domain, _ := k.GetDomain(ctx, "StonesDomain")
	if domain.Issues[0].Suggestions[0].Stones != 1 {
		t.Errorf("GreenDeal stones = %d, want 1", domain.Issues[0].Suggestions[0].Stones)
	}
	if domain.Issues[1].Suggestions[0].Stones != 1 {
		t.Errorf("FreeTuition stones = %d, want 1", domain.Issues[1].Suggestions[0].Stones)
	}
}

// ---------- IssueStoneIndependentOfSuggestionStone ----------

func TestIssueAndSuggestionStonesIndependent(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupDomainWithIssues(t, k, ctx)

	// Alice places stone on Climate issue.
	_, err := k.PlaceStoneOnIssue(ctx, "StonesDomain", "Climate", "alice")
	if err != nil {
		t.Fatal(err)
	}

	// Alice also places stone on GreenDeal suggestion (different list).
	_, err = k.PlaceStoneOnSuggestion(ctx, "StonesDomain", "Climate", "GreenDeal", "alice")
	if err != nil {
		t.Fatal(err)
	}

	// Both should coexist — issue list and suggestion list are independent.
	issueStone, _ := k.GetMemberIssueStone(ctx, "StonesDomain", "alice")
	suggStone, _ := k.GetMemberSuggestionStone(ctx, "StonesDomain", "Climate", "alice")

	if issueStone != "Climate" {
		t.Errorf("issue stone = %q, want 'Climate'", issueStone)
	}
	if suggStone != "GreenDeal" {
		t.Errorf("suggestion stone = %q, want 'GreenDeal'", suggStone)
	}
}

// ---------- GH-306: first-placement VoteToEarn ----------

// TestStoneRewardPaidOncePerScope proves the reward is paid exactly once per
// member and voting scope: the first placement earns it, moves never pay, and
// independent scopes (issue list, each suggestion list, other members) are
// tracked separately.
func TestStoneRewardPaidOncePerScope(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupDomainWithIssues(t, k, ctx)

	// First placement on the issue list pays.
	reward, err := k.PlaceStoneOnIssue(ctx, "StonesDomain", "Climate", "alice")
	if err != nil {
		t.Fatal(err)
	}
	if !reward.AmountOf(PNYXDenom).IsPositive() {
		t.Fatal("first issue placement must pay the VoteToEarn reward")
	}
	if !k.HasStoneRewardRecord(ctx, StoneRewardRecord{DomainName: "StonesDomain", MemberAddr: "alice"}) {
		t.Fatal("first issue placement did not consume the reward marker")
	}

	// Moving the stone within the same scope never pays again.
	reward, err = k.PlaceStoneOnIssue(ctx, "StonesDomain", "Education", "alice")
	if err != nil {
		t.Fatal(err)
	}
	if !reward.Empty() {
		t.Fatalf("issue move paid %s, want no reward", reward)
	}

	// The suggestion list of one issue is an independent scope.
	reward, err = k.PlaceStoneOnSuggestion(ctx, "StonesDomain", "Climate", "GreenDeal", "alice")
	if err != nil {
		t.Fatal(err)
	}
	if !reward.AmountOf(PNYXDenom).IsPositive() {
		t.Fatal("first suggestion placement must pay the VoteToEarn reward")
	}

	// Moving inside that suggestion list never pays again.
	reward, err = k.PlaceStoneOnSuggestion(ctx, "StonesDomain", "Climate", "CarbonTax", "alice")
	if err != nil {
		t.Fatal(err)
	}
	if !reward.Empty() {
		t.Fatalf("suggestion move paid %s, want no reward", reward)
	}

	// Another issue's suggestion list is yet another independent scope.
	reward, err = k.PlaceStoneOnSuggestion(ctx, "StonesDomain", "Education", "FreeTuition", "alice")
	if err != nil {
		t.Fatal(err)
	}
	if !reward.AmountOf(PNYXDenom).IsPositive() {
		t.Fatal("first placement in a second suggestion scope must pay")
	}

	// The marker is per member: bob's first issue placement still pays.
	reward, err = k.PlaceStoneOnIssue(ctx, "StonesDomain", "Climate", "bob")
	if err != nil {
		t.Fatal(err)
	}
	if !reward.AmountOf(PNYXDenom).IsPositive() {
		t.Fatal("another member's first placement must pay")
	}

	// Eligibility is tied to the first placement, not to a later treasury
	// refill. A zero-valued first reward is still consumed fail-closed.
	emptyAdmin := sdk.AccAddress("empty-treasury-admin")
	k.CreateDomain(ctx, "EmptyTreasury", emptyAdmin, sdk.NewCoins())
	emptyDomain, _ := k.GetDomain(ctx, "EmptyTreasury")
	emptyDomain.Issues = []Issue{{Name: "Zero", CreationDate: ctx.BlockTime().Unix()}}
	saveDomain(t, k, ctx, emptyDomain)
	reward, err = k.PlaceStoneOnIssue(ctx, "EmptyTreasury", "Zero", emptyAdmin.String())
	if err != nil {
		t.Fatal(err)
	}
	if !reward.Empty() {
		t.Fatalf("zero-treasury placement paid %s, want no reward", reward)
	}
	if !k.HasStoneRewardRecord(ctx, StoneRewardRecord{DomainName: "EmptyTreasury", MemberAddr: emptyAdmin.String()}) {
		t.Fatal("zero-treasury first placement did not consume eligibility")
	}
	emptyDomain, _ = k.GetDomain(ctx, "EmptyTreasury")
	emptyDomain.Treasury = sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 100_000))
	saveDomain(t, k, ctx, emptyDomain)
	ctx.KVStore(k.StoreKey).Delete(issueStoneKey("EmptyTreasury", emptyAdmin.String()))
	reward, err = k.PlaceStoneOnIssue(ctx, "EmptyTreasury", "Zero", emptyAdmin.String())
	if err != nil {
		t.Fatal(err)
	}
	if !reward.Empty() {
		t.Fatalf("treasury refill replay paid %s, want consumed eligibility", reward)
	}
}

// TestStoneRewardMarkerSurvivesExclusionAndReentry proves the consumed marker
// persists through exclusion cleanup and re-onboarding: a member who re-enters
// the domain and places a stone in the same scope is never rewarded twice.
func TestStoneRewardMarkerSurvivesExclusionAndReentry(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupGovernanceDomain(t, k, ctx)
	admin := sdk.AccAddress("admin1")

	reward, err := k.PlaceStoneOnIssue(ctx, "GovDomain", "Climate", "judy")
	if err != nil {
		t.Fatal(err)
	}
	if !reward.AmountOf(PNYXDenom).IsPositive() {
		t.Fatal("first placement must pay")
	}
	if _, err := k.PlaceStoneOnSuggestion(ctx, "GovDomain", "Climate", "GreenDeal", "judy"); err != nil {
		t.Fatal(err)
	}
	domainBefore, _ := k.GetDomain(ctx, "GovDomain")
	treasuryBefore := domainBefore.Treasury.AmountOf(PNYXDenom)
	payoutsBefore := domainBefore.TotalPayouts

	// Exclude judy (2/3 of the 9 remaining voters): stone keys are cleaned up.
	for _, m := range []string{"alice", "bob", "charlie", "dave", "eve", "frank", "grace"} {
		excluded, err := k.VoteToExclude(ctx, "GovDomain", "judy", m)
		if err != nil {
			t.Fatal(err)
		}
		_ = excluded
	}
	if _, found := k.GetMemberIssueStone(ctx, "GovDomain", "judy"); found {
		t.Fatal("exclusion did not clean up the issue stone")
	}

	// The consumed markers survive the cleanup.
	if !k.HasStoneRewardRecord(ctx, StoneRewardRecord{DomainName: "GovDomain", MemberAddr: "judy"}) {
		t.Fatal("exclusion cleanup deleted the issue-scope reward marker")
	}
	if !k.HasStoneRewardRecord(ctx, StoneRewardRecord{DomainName: "GovDomain", IssueName: "Climate", MemberAddr: "judy"}) {
		t.Fatal("exclusion cleanup deleted the suggestion-scope reward marker")
	}

	// Re-onboard and re-place: no second reward in either scope.
	if err := k.AddMember(ctx, "GovDomain", "judy", admin); err != nil {
		t.Fatal(err)
	}
	reward, err = k.PlaceStoneOnIssue(ctx, "GovDomain", "Climate", "judy")
	if err != nil {
		t.Fatal(err)
	}
	if !reward.Empty() {
		t.Fatalf("re-entry issue placement paid %s, want no reward", reward)
	}
	reward, err = k.PlaceStoneOnSuggestion(ctx, "GovDomain", "Climate", "GreenDeal", "judy")
	if err != nil {
		t.Fatal(err)
	}
	if !reward.Empty() {
		t.Fatalf("re-entry suggestion placement paid %s, want no reward", reward)
	}
	placed, found := k.GetMemberIssueStone(ctx, "GovDomain", "judy")
	if !found || placed != "Climate" {
		t.Fatalf("re-entry stone = %q, want Climate", placed)
	}
	domainAfter, _ := k.GetDomain(ctx, "GovDomain")
	if !domainAfter.Treasury.AmountOf(PNYXDenom).Equal(treasuryBefore) || domainAfter.TotalPayouts != payoutsBefore {
		t.Fatal("re-entry moved treasury or payout accounting")
	}
}

// TestBaselineStoneRewardMarkersMigration proves the deterministic GH-306
// 2→3 migration marks every stone that exists at the upgrade boundary as
// already rewarded, stays idempotent, and does not block first-placement
// rewards for scopes that had no stone at the boundary.
func TestBaselineStoneRewardMarkersMigration(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupDomainWithIssues(t, k, ctx)

	if got := (AppModule{}).ConsensusVersion(); got != 3 {
		t.Fatalf("ConsensusVersion = %d, want 3", got)
	}

	// Pre-upgrade state: stones exist without consumed markers.
	if _, err := k.PlaceStoneOnIssue(ctx, "StonesDomain", "Climate", "alice"); err != nil {
		t.Fatal(err)
	}
	if _, err := k.PlaceStoneOnSuggestion(ctx, "StonesDomain", "Climate", "GreenDeal", "bob"); err != nil {
		t.Fatal(err)
	}
	store := ctx.KVStore(k.StoreKey)
	// A second domain makes the legacy delimiter key below ambiguous. The
	// migration must fail closed by consuming both interpretations, while the
	// new length-prefixed marker keys remain distinct.
	k.CreateDomain(ctx, "StonesDomain:Cleaned", sdk.AccAddress("admin2"), sdk.NewCoins())
	// Lifecycle cleanup may leave a historical suggestion-stone key after its
	// issue disappears. It must still be baselined, including ':' in the scope.
	orphanRecord := StoneRewardRecord{DomainName: "StonesDomain", IssueName: "Cleaned:Issue", MemberAddr: "bob"}
	ambiguousRecord := StoneRewardRecord{DomainName: "StonesDomain:Cleaned", IssueName: "Issue", MemberAddr: "bob"}
	if bytes.Equal(stoneRewardRecordKey(orphanRecord), stoneRewardRecordKey(ambiguousRecord)) {
		t.Fatal("length-prefixed reward marker keys collided across domain/issue boundaries")
	}
	store.Set(suggestionStoneKey(orphanRecord.DomainName, orphanRecord.IssueName, orphanRecord.MemberAddr), []byte("LegacySuggestion"))
	var markerKeys [][]byte
	k.IterateStoneRewardRecords(ctx, func(record StoneRewardRecord) bool {
		markerKeys = append(markerKeys, stoneRewardRecordKey(record))
		return false
	})
	if len(markerKeys) != 2 {
		t.Fatalf("consumed markers before reset = %d, want 2", len(markerKeys))
	}
	for _, key := range markerKeys {
		store.Delete(key)
	}

	// The migration baselines exactly the existing stones.
	k.BaselineStoneRewardMarkers(ctx)
	if !k.HasStoneRewardRecord(ctx, StoneRewardRecord{DomainName: "StonesDomain", MemberAddr: "alice"}) {
		t.Fatal("migration did not baseline alice's issue-scope stone")
	}
	if !k.HasStoneRewardRecord(ctx, StoneRewardRecord{DomainName: "StonesDomain", IssueName: "Climate", MemberAddr: "bob"}) {
		t.Fatal("migration did not baseline bob's suggestion-scope stone")
	}
	if !k.HasStoneRewardRecord(ctx, orphanRecord) {
		t.Fatal("migration did not baseline a lifecycle-cleaned suggestion scope")
	}
	if !k.HasStoneRewardRecord(ctx, ambiguousRecord) {
		t.Fatal("migration did not conservatively baseline an ambiguous legacy suggestion key")
	}
	if k.HasStoneRewardRecord(ctx, StoneRewardRecord{DomainName: "StonesDomain", MemberAddr: "charlie"}) {
		t.Fatal("migration baselined a scope without a stone")
	}

	// Idempotent: a second run adds nothing.
	k.BaselineStoneRewardMarkers(ctx)
	count := 0
	k.IterateStoneRewardRecords(ctx, func(StoneRewardRecord) bool {
		count++
		return false
	})
	if count != 4 {
		t.Fatalf("markers after idempotent re-run = %d, want 4", count)
	}

	// Baselined scopes never pay again, even after a move.
	reward, err := k.PlaceStoneOnIssue(ctx, "StonesDomain", "Education", "alice")
	if err != nil {
		t.Fatal(err)
	}
	if !reward.Empty() {
		t.Fatalf("baselined scope paid %s after migration", reward)
	}

	// A scope without a stone at the boundary still pays its first placement.
	reward, err = k.PlaceStoneOnIssue(ctx, "StonesDomain", "Climate", "charlie")
	if err != nil {
		t.Fatal(err)
	}
	if !reward.AmountOf(PNYXDenom).IsPositive() {
		t.Fatal("new scope after migration must pay its first placement")
	}
}
