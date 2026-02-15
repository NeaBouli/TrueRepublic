package truedemocracy

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// setupGovernanceDomain creates a domain with 10 members for governance tests.
func setupGovernanceDomain(t *testing.T, k Keeper, ctx sdk.Context) {
	t.Helper()
	k.CreateDomain(ctx, "GovDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 1_000_000)))

	domain, _ := k.GetDomain(ctx, "GovDomain")
	domain.Members = []string{"alice", "bob", "charlie", "dave", "eve", "frank", "grace", "heidi", "ivan", "judy"}
	domain.Options.AdminElectable = true

	now := ctx.BlockTime().Unix()
	domain.Issues = []Issue{
		{
			Name: "Climate", Stones: 0, CreationDate: now, LastActivityAt: now,
			Suggestions: []Suggestion{
				{Name: "GreenDeal", Creator: "alice", Stones: 0, Ratings: []Rating{}, CreationDate: now},
			},
		},
		{
			Name: "Education", Stones: 0, CreationDate: now, LastActivityAt: now,
			Suggestions: []Suggestion{
				{Name: "FreeTuition", Creator: "bob", Stones: 0, Ratings: []Rating{}, CreationDate: now},
			},
		},
	}

	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:GovDomain"), bz)
}

// ---------- PlaceStoneOnMember ----------

func TestPlaceStoneOnMember(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupGovernanceDomain(t, k, ctx)

	t.Run("happy path", func(t *testing.T) {
		err := k.PlaceStoneOnMember(ctx, "GovDomain", "bob", "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		target, found := k.GetMemberStone(ctx, "GovDomain", "alice")
		if !found || target != "bob" {
			t.Errorf("alice stone = %q, want 'bob'", target)
		}
	})

	t.Run("move stone", func(t *testing.T) {
		err := k.PlaceStoneOnMember(ctx, "GovDomain", "charlie", "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		target, _ := k.GetMemberStone(ctx, "GovDomain", "alice")
		if target != "charlie" {
			t.Errorf("alice stone = %q, want 'charlie' after move", target)
		}
	})

	t.Run("duplicate rejected", func(t *testing.T) {
		err := k.PlaceStoneOnMember(ctx, "GovDomain", "charlie", "alice")
		if err == nil {
			t.Fatal("expected error for duplicate stone placement")
		}
	})

	t.Run("self-vote rejected", func(t *testing.T) {
		err := k.PlaceStoneOnMember(ctx, "GovDomain", "alice", "alice")
		if err == nil {
			t.Fatal("expected error for self-vote")
		}
	})

	t.Run("non-member rejected", func(t *testing.T) {
		err := k.PlaceStoneOnMember(ctx, "GovDomain", "bob", "outsider")
		if err == nil {
			t.Fatal("expected error for non-member")
		}
	})

	t.Run("target not member rejected", func(t *testing.T) {
		err := k.PlaceStoneOnMember(ctx, "GovDomain", "outsider", "alice")
		if err == nil {
			t.Fatal("expected error for non-member target")
		}
	})
}

// ---------- SortMembersByStones ----------

func TestSortMembersByStones(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupGovernanceDomain(t, k, ctx)

	// alice, bob, charlie all vote for dave. eve votes for alice.
	k.PlaceStoneOnMember(ctx, "GovDomain", "dave", "alice")
	k.PlaceStoneOnMember(ctx, "GovDomain", "dave", "bob")
	k.PlaceStoneOnMember(ctx, "GovDomain", "dave", "charlie")
	k.PlaceStoneOnMember(ctx, "GovDomain", "alice", "eve")

	domain, _ := k.GetDomain(ctx, "GovDomain")
	ranks := k.SortMembersByStones(ctx, domain)

	if ranks[0].Address != "dave" || ranks[0].Stones != 3 {
		t.Errorf("rank 0 = %v, want dave with 3 stones", ranks[0])
	}
	if ranks[1].Address != "alice" || ranks[1].Stones != 1 {
		t.Errorf("rank 1 = %v, want alice with 1 stone", ranks[1])
	}
	// Everyone else should have 0 stones.
	for _, r := range ranks[2:] {
		if r.Stones != 0 {
			t.Errorf("%s has %d stones, want 0", r.Address, r.Stones)
		}
	}
}

// ---------- ElectAdmin ----------

func TestElectAdmin(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupGovernanceDomain(t, k, ctx)

	t.Run("highest stone count becomes admin", func(t *testing.T) {
		// 3 members vote for dave.
		k.PlaceStoneOnMember(ctx, "GovDomain", "dave", "alice")
		k.PlaceStoneOnMember(ctx, "GovDomain", "dave", "bob")
		k.PlaceStoneOnMember(ctx, "GovDomain", "dave", "charlie")

		err := k.ElectAdmin(ctx, "GovDomain")
		if err != nil {
			t.Fatal(err)
		}

		domain, _ := k.GetDomain(ctx, "GovDomain")
		if string(domain.Admin) != "dave" {
			t.Errorf("admin = %q, want 'dave'", string(domain.Admin))
		}
	})

	t.Run("no stones means no change", func(t *testing.T) {
		// Create a fresh domain with no stones.
		k.CreateDomain(ctx, "EmptyDomain", sdk.AccAddress("origadmin"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100)))
		domain, _ := k.GetDomain(ctx, "EmptyDomain")
		domain.Options.AdminElectable = true
		domain.Members = []string{"a", "b", "c"}
		st := ctx.KVStore(k.StoreKey)
		bz := k.cdc.MustMarshalLengthPrefixed(&domain)
		st.Set([]byte("domain:EmptyDomain"), bz)

		k.ElectAdmin(ctx, "EmptyDomain")

		domain, _ = k.GetDomain(ctx, "EmptyDomain")
		if string(domain.Admin) != "origadmin" {
			t.Error("admin should not change when no stones are placed")
		}
	})

	t.Run("not electable skips election", func(t *testing.T) {
		k.CreateDomain(ctx, "FixedAdmin", sdk.AccAddress("boss"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100)))
		domain, _ := k.GetDomain(ctx, "FixedAdmin")
		domain.Options.AdminElectable = false
		domain.Members = []string{"boss", "worker"}
		st := ctx.KVStore(k.StoreKey)
		bz := k.cdc.MustMarshalLengthPrefixed(&domain)
		st.Set([]byte("domain:FixedAdmin"), bz)

		k.PlaceStoneOnMember(ctx, "FixedAdmin", "worker", "boss")
		k.ElectAdmin(ctx, "FixedAdmin")

		domain, _ = k.GetDomain(ctx, "FixedAdmin")
		if string(domain.Admin) != "boss" {
			t.Error("admin should not change when AdminElectable is false")
		}
	})
}

// ---------- VoteToExclude ----------

func TestVoteToExclude(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupGovernanceDomain(t, k, ctx)

	t.Run("single vote does not exclude", func(t *testing.T) {
		excluded, err := k.VoteToExclude(ctx, "GovDomain", "judy", "alice")
		if err != nil {
			t.Fatal(err)
		}
		if excluded {
			t.Error("single vote should not exclude (need 2/3)")
		}
	})

	t.Run("duplicate vote rejected", func(t *testing.T) {
		_, err := k.VoteToExclude(ctx, "GovDomain", "judy", "alice")
		if err == nil {
			t.Fatal("expected error for duplicate vote")
		}
	})

	t.Run("self-exclusion rejected", func(t *testing.T) {
		_, err := k.VoteToExclude(ctx, "GovDomain", "judy", "judy")
		if err == nil {
			t.Fatal("expected error for self-exclusion vote")
		}
	})

	t.Run("non-member rejected", func(t *testing.T) {
		_, err := k.VoteToExclude(ctx, "GovDomain", "judy", "outsider")
		if err == nil {
			t.Fatal("expected error for non-member")
		}
	})

	t.Run("2/3 majority excludes", func(t *testing.T) {
		// alice already voted. 9 voters (judy excluded from count).
		// Need 6 votes: 6*10000=60000 >= 9*6667=60003... need 7.
		// Actually: 7*10000=70000 >= 9*6667=60003 → yes.
		for _, m := range []string{"bob", "charlie", "dave", "eve", "frank"} {
			excluded, err := k.VoteToExclude(ctx, "GovDomain", "judy", m)
			if err != nil {
				t.Fatalf("vote by %s failed: %v", m, err)
			}
			if excluded {
				t.Errorf("vote by %s should not have triggered exclusion yet", m)
			}
		}

		// 7th vote should trigger (alice + 5 above + grace = 7).
		excluded, err := k.VoteToExclude(ctx, "GovDomain", "judy", "grace")
		if err != nil {
			t.Fatal(err)
		}
		if !excluded {
			t.Error("7/9 votes should reach 2/3 majority and exclude")
		}

		// Verify judy is gone.
		domain, _ := k.GetDomain(ctx, "GovDomain")
		for _, m := range domain.Members {
			if m == "judy" {
				t.Error("judy should have been excluded")
			}
		}
		if len(domain.Members) != 9 {
			t.Errorf("members count = %d, want 9", len(domain.Members))
		}
	})
}

// ---------- Exclusion stone cleanup ----------

func TestExclusionCleansUpStones(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupGovernanceDomain(t, k, ctx)

	// judy places stones.
	k.PlaceStoneOnIssue(ctx, "GovDomain", "Climate", "judy")
	k.PlaceStoneOnSuggestion(ctx, "GovDomain", "Climate", "GreenDeal", "judy")
	k.PlaceStoneOnMember(ctx, "GovDomain", "alice", "judy")

	// Verify stones are placed.
	domain, _ := k.GetDomain(ctx, "GovDomain")
	if domain.Issues[0].Stones != 1 {
		t.Fatalf("Climate stones = %d, want 1", domain.Issues[0].Stones)
	}

	// Exclude judy (need 2/3 of 9 other voters = 7 votes).
	for _, m := range []string{"alice", "bob", "charlie", "dave", "eve", "frank", "grace"} {
		k.VoteToExclude(ctx, "GovDomain", "judy", m)
	}

	domain, _ = k.GetDomain(ctx, "GovDomain")

	// Issue stone should be decremented.
	if domain.Issues[0].Stones != 0 {
		t.Errorf("Climate stones = %d, want 0 after judy exclusion", domain.Issues[0].Stones)
	}

	// Suggestion stone should be decremented.
	if domain.Issues[0].Suggestions[0].Stones != 0 {
		t.Errorf("GreenDeal stones = %d, want 0 after judy exclusion", domain.Issues[0].Suggestions[0].Stones)
	}

	// Member stone KV entry should be gone.
	_, found := k.GetMemberStone(ctx, "GovDomain", "judy")
	if found {
		t.Error("judy's member stone should be cleaned up")
	}
}

// ---------- CleanupInactiveIssues ----------

func TestCleanupInactiveIssues(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupGovernanceDomain(t, k, ctx)

	t.Run("active issues kept", func(t *testing.T) {
		err := k.CleanupInactiveIssues(ctx, "GovDomain")
		if err != nil {
			t.Fatal(err)
		}

		domain, _ := k.GetDomain(ctx, "GovDomain")
		if len(domain.Issues) != 2 {
			t.Errorf("issues count = %d, want 2 (recent issues should be kept)", len(domain.Issues))
		}
	})

	t.Run("old inactive issues removed", func(t *testing.T) {
		// Advance time 361 days.
		futureCtx := ctx.WithBlockTime(ctx.BlockTime().Add(361 * 24 * time.Hour))

		err := k.CleanupInactiveIssues(futureCtx, "GovDomain")
		if err != nil {
			t.Fatal(err)
		}

		domain, _ := k.GetDomain(futureCtx, "GovDomain")
		if len(domain.Issues) != 0 {
			t.Errorf("issues count = %d, want 0 (inactive issues should be removed)", len(domain.Issues))
		}
	})
}

func TestInactivityPartialCleanup(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupGovernanceDomain(t, k, ctx)

	// Set Climate activity to now, but Education to 400 days ago.
	domain, _ := k.GetDomain(ctx, "GovDomain")
	domain.Issues[1].LastActivityAt = ctx.BlockTime().Unix() - 400*86400
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:GovDomain"), bz)

	err := k.CleanupInactiveIssues(ctx, "GovDomain")
	if err != nil {
		t.Fatal(err)
	}

	domain, _ = k.GetDomain(ctx, "GovDomain")
	if len(domain.Issues) != 1 {
		t.Fatalf("issues count = %d, want 1", len(domain.Issues))
	}
	if domain.Issues[0].Name != "Climate" {
		t.Errorf("remaining issue = %q, want Climate", domain.Issues[0].Name)
	}
}

func TestActivityResetsInactivityTimer(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupGovernanceDomain(t, k, ctx)

	// Set Climate activity to 350 days ago (almost expired).
	domain, _ := k.GetDomain(ctx, "GovDomain")
	domain.Issues[0].LastActivityAt = ctx.BlockTime().Unix() - 350*86400
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:GovDomain"), bz)

	// Place a stone — this should update LastActivityAt to now.
	k.PlaceStoneOnIssue(ctx, "GovDomain", "Climate", "alice")

	// Advance 350 days — issue should survive because activity was reset.
	futureCtx := ctx.WithBlockTime(ctx.BlockTime().Add(350 * 24 * time.Hour))
	k.CleanupInactiveIssues(futureCtx, "GovDomain")

	domain, _ = k.GetDomain(futureCtx, "GovDomain")
	found := false
	for _, issue := range domain.Issues {
		if issue.Name == "Climate" {
			found = true
		}
	}
	if !found {
		t.Error("Climate should survive — stone placement reset the activity timer")
	}
}

// ---------- ExternalLink ----------

func TestExternalLink(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateDomain(ctx, "LinkDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500_000)))

	domain, _ := k.GetDomain(ctx, "LinkDomain")
	domain.Members = append(domain.Members, "alice")
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:LinkDomain"), bz)

	// Submit with external link.
	err := k.SubmitProposal(ctx, "LinkDomain", "Policy", "Plan", "alice",
		sdk.NewCoins(sdk.NewInt64Coin("pnyx", 1000)), "https://forum.example.com/policy")
	if err != nil {
		t.Fatal(err)
	}

	domain, _ = k.GetDomain(ctx, "LinkDomain")
	if domain.Issues[0].Suggestions[0].ExternalLink != "https://forum.example.com/policy" {
		t.Errorf("external link = %q, want forum URL", domain.Issues[0].Suggestions[0].ExternalLink)
	}
}

func TestExternalLinkEmpty(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateDomain(ctx, "NoLinkDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500_000)))

	domain, _ := k.GetDomain(ctx, "NoLinkDomain")
	domain.Members = append(domain.Members, "alice")
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:NoLinkDomain"), bz)

	// Submit without external link.
	err := k.SubmitProposal(ctx, "NoLinkDomain", "Policy", "Plan", "alice",
		sdk.NewCoins(sdk.NewInt64Coin("pnyx", 1000)), "")
	if err != nil {
		t.Fatal(err)
	}

	domain, _ = k.GetDomain(ctx, "NoLinkDomain")
	if domain.Issues[0].Suggestions[0].ExternalLink != "" {
		t.Errorf("external link = %q, want empty", domain.Issues[0].Suggestions[0].ExternalLink)
	}
}

// ---------- ProcessGovernance ----------

func TestProcessGovernance(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupGovernanceDomain(t, k, ctx)

	// Vote for alice as admin.
	k.PlaceStoneOnMember(ctx, "GovDomain", "alice", "bob")
	k.PlaceStoneOnMember(ctx, "GovDomain", "alice", "charlie")

	k.ProcessGovernance(ctx)

	domain, _ := k.GetDomain(ctx, "GovDomain")
	if string(domain.Admin) != "alice" {
		t.Errorf("admin = %q, want 'alice' after governance processing", string(domain.Admin))
	}
}
