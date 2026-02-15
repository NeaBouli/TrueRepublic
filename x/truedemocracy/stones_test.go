package truedemocracy

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// setupDomainWithIssues creates a domain with members and multiple issues/suggestions for stone tests.
func setupDomainWithIssues(t *testing.T, k Keeper, ctx sdk.Context) {
	t.Helper()
	k.CreateDomain(ctx, "StonesDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500_000)))

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
		if !reward.AmountOf("pnyx").IsPositive() {
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
	if !reward.AmountOf("pnyx").IsPositive() {
		t.Error("moving stone should also earn reward")
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
		{Name: "C", Stones: 1, CreationDate: now},      // same stones as A, but older
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
		{Name: "Z", Stones: 2, CreationDate: now},       // same stones as Y, but older
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
	treasuryBefore := domainBefore.Treasury.AmountOf("pnyx")

	// Expected reward = treasury / 1000.
	expectedReward := treasuryBefore.Quo(math.NewInt(1000))

	reward, err := k.PlaceStoneOnIssue(ctx, "StonesDomain", "Climate", "alice")
	if err != nil {
		t.Fatal(err)
	}
	if !reward.AmountOf("pnyx").Equal(expectedReward) {
		t.Errorf("reward = %s, want %s", reward.AmountOf("pnyx"), expectedReward)
	}

	domainAfter, _ := k.GetDomain(ctx, "StonesDomain")
	treasuryAfter := domainAfter.Treasury.AmountOf("pnyx")
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
		if !reward.AmountOf("pnyx").IsPositive() {
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
