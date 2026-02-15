package truedemocracy

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// setupLifecycleDomain creates a domain with 10 members and one issue with
// two suggestions, suitable for lifecycle zone testing.
func setupLifecycleDomain(t *testing.T, k Keeper, ctx sdk.Context) {
	t.Helper()
	k.CreateDomain(ctx, "LifeDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 1_000_000)))

	domain, _ := k.GetDomain(ctx, "LifeDomain")
	domain.Members = []string{"m1", "m2", "m3", "m4", "m5", "m6", "m7", "m8", "m9", "m10"}

	now := ctx.BlockTime().Unix()
	domain.Issues = []Issue{
		{
			Name: "PolicyA", Stones: 0, CreationDate: now,
			Suggestions: []Suggestion{
				{Name: "S1", Creator: "m1", Stones: 0, Ratings: []Rating{}, CreationDate: now, Color: ""},
				{Name: "S2", Creator: "m2", Stones: 0, Ratings: []Rating{}, CreationDate: now, Color: ""},
			},
		},
	}

	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:LifeDomain"), bz)
}

// setSuggestionStones directly sets the stone count on a suggestion.
func setSuggestionStones(t *testing.T, k Keeper, ctx sdk.Context, domainName string, issueIdx, suggIdx, stones int) {
	t.Helper()
	domain, _ := k.GetDomain(ctx, domainName)
	domain.Issues[issueIdx].Suggestions[suggIdx].Stones = stones
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:"+domainName), bz)
}

// setSuggestionColor directly sets the color/zone state on a suggestion.
func setSuggestionColor(t *testing.T, k Keeper, ctx sdk.Context, domainName string, issueIdx, suggIdx int, color string, yellowAt, redAt int64) {
	t.Helper()
	domain, _ := k.GetDomain(ctx, domainName)
	domain.Issues[issueIdx].Suggestions[suggIdx].Color = color
	domain.Issues[issueIdx].Suggestions[suggIdx].EnteredYellowAt = yellowAt
	domain.Issues[issueIdx].Suggestions[suggIdx].EnteredRedAt = redAt
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:"+domainName), bz)
}

// ---------- MeetsApprovalThreshold ----------

func TestMeetsApprovalThreshold(t *testing.T) {
	// Default threshold is 500 bps = 5%.
	// With 10 members, need stones*10000 >= 10*500 = 5000.
	// So need stones >= 1 (1*10000=10000 >= 5000).

	t.Run("above threshold", func(t *testing.T) {
		if !MeetsApprovalThreshold(1, 10, 500) {
			t.Error("1 stone out of 10 members at 5% should meet threshold")
		}
	})

	t.Run("exactly at threshold", func(t *testing.T) {
		// 5 stones, 100 members, 500 bps: 5*10000=50000 >= 100*500=50000 → true
		if !MeetsApprovalThreshold(5, 100, 500) {
			t.Error("5 stones out of 100 at 5% should meet threshold")
		}
	})

	t.Run("below threshold", func(t *testing.T) {
		// 0 stones out of 10 → 0 < 5000
		if MeetsApprovalThreshold(0, 10, 500) {
			t.Error("0 stones should not meet threshold")
		}
	})

	t.Run("zero members", func(t *testing.T) {
		if MeetsApprovalThreshold(1, 0, 500) {
			t.Error("zero members should always return false")
		}
	})

	t.Run("high threshold", func(t *testing.T) {
		// 50% threshold (5000 bps). 4 stones out of 10: 4*10000=40000 < 10*5000=50000
		if MeetsApprovalThreshold(4, 10, 5000) {
			t.Error("4/10 should not meet 50% threshold")
		}
		if !MeetsApprovalThreshold(5, 10, 5000) {
			t.Error("5/10 should meet 50% threshold")
		}
	})
}

// ---------- Green zone: suggestion stays when approved ----------

func TestGreenZoneStays(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupLifecycleDomain(t, k, ctx)

	// Give S1 enough stones (1 stone with 10 members at 5% = approved).
	setSuggestionStones(t, k, ctx, "LifeDomain", 0, 0, 1)

	err := k.EvaluateSuggestionZones(ctx, "LifeDomain")
	if err != nil {
		t.Fatal(err)
	}

	domain, _ := k.GetDomain(ctx, "LifeDomain")
	s := domain.Issues[0].Suggestions[0]
	if s.Color != "green" {
		t.Errorf("S1 color = %q, want green", s.Color)
	}
	if s.EnteredYellowAt != 0 || s.EnteredRedAt != 0 {
		t.Error("green zone should have zero timestamps")
	}
}

// ---------- Yellow zone: drop below threshold ----------

func TestYellowTransition(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupLifecycleDomain(t, k, ctx)

	// S1 has 0 stones — below threshold.
	err := k.EvaluateSuggestionZones(ctx, "LifeDomain")
	if err != nil {
		t.Fatal(err)
	}

	domain, _ := k.GetDomain(ctx, "LifeDomain")
	s := domain.Issues[0].Suggestions[0]
	if s.Color != "yellow" {
		t.Errorf("S1 color = %q, want yellow", s.Color)
	}
	if s.EnteredYellowAt != ctx.BlockTime().Unix() {
		t.Errorf("EnteredYellowAt = %d, want %d", s.EnteredYellowAt, ctx.BlockTime().Unix())
	}
}

// ---------- Yellow → Red transition after dwell time ----------

func TestYellowToRedTransition(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupLifecycleDomain(t, k, ctx)

	now := ctx.BlockTime().Unix()

	// Set S1 to yellow, entered 2 days ago (dwell time is 1 day = 86400s).
	setSuggestionColor(t, k, ctx, "LifeDomain", 0, 0, "yellow", now-2*86400, 0)

	err := k.EvaluateSuggestionZones(ctx, "LifeDomain")
	if err != nil {
		t.Fatal(err)
	}

	domain, _ := k.GetDomain(ctx, "LifeDomain")
	s := domain.Issues[0].Suggestions[0]
	if s.Color != "red" {
		t.Errorf("S1 color = %q, want red", s.Color)
	}
	if s.EnteredRedAt != now {
		t.Errorf("EnteredRedAt = %d, want %d", s.EnteredRedAt, now)
	}
}

// ---------- Red → auto-delete after dwell time ----------

func TestRedAutoDelete(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupLifecycleDomain(t, k, ctx)

	now := ctx.BlockTime().Unix()

	// Set S1 to red, entered 2 days ago.
	setSuggestionColor(t, k, ctx, "LifeDomain", 0, 0, "red", 0, now-2*86400)

	// S2 is still valid (will go yellow).
	err := k.EvaluateSuggestionZones(ctx, "LifeDomain")
	if err != nil {
		t.Fatal(err)
	}

	domain, _ := k.GetDomain(ctx, "LifeDomain")
	if len(domain.Issues[0].Suggestions) != 1 {
		t.Fatalf("suggestions count = %d, want 1 (S1 should be deleted)", len(domain.Issues[0].Suggestions))
	}
	if domain.Issues[0].Suggestions[0].Name != "S2" {
		t.Errorf("remaining suggestion = %q, want S2", domain.Issues[0].Suggestions[0].Name)
	}
}

// ---------- Recovery: yellow → green ----------

func TestRecoveryFromYellow(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupLifecycleDomain(t, k, ctx)

	now := ctx.BlockTime().Unix()

	// S1 is yellow but now has enough stones.
	setSuggestionColor(t, k, ctx, "LifeDomain", 0, 0, "yellow", now-100, 0)
	setSuggestionStones(t, k, ctx, "LifeDomain", 0, 0, 1)

	err := k.EvaluateSuggestionZones(ctx, "LifeDomain")
	if err != nil {
		t.Fatal(err)
	}

	domain, _ := k.GetDomain(ctx, "LifeDomain")
	s := domain.Issues[0].Suggestions[0]
	if s.Color != "green" {
		t.Errorf("S1 color = %q, want green after recovery", s.Color)
	}
	if s.EnteredYellowAt != 0 {
		t.Error("EnteredYellowAt should be cleared after recovery")
	}
}

// ---------- Recovery: red → green ----------

func TestRecoveryFromRed(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupLifecycleDomain(t, k, ctx)

	now := ctx.BlockTime().Unix()

	// S1 is red but now has enough stones.
	setSuggestionColor(t, k, ctx, "LifeDomain", 0, 0, "red", now-200, now-100)
	setSuggestionStones(t, k, ctx, "LifeDomain", 0, 0, 1)

	err := k.EvaluateSuggestionZones(ctx, "LifeDomain")
	if err != nil {
		t.Fatal(err)
	}

	domain, _ := k.GetDomain(ctx, "LifeDomain")
	s := domain.Issues[0].Suggestions[0]
	if s.Color != "green" {
		t.Errorf("S1 color = %q, want green after recovery from red", s.Color)
	}
	if s.EnteredRedAt != 0 || s.EnteredYellowAt != 0 {
		t.Error("zone timestamps should be cleared after recovery")
	}
}

// ---------- Yellow stays yellow before dwell time ----------

func TestYellowStaysBeforeDwell(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupLifecycleDomain(t, k, ctx)

	now := ctx.BlockTime().Unix()

	// S1 entered yellow 1 hour ago (dwell time is 1 day).
	setSuggestionColor(t, k, ctx, "LifeDomain", 0, 0, "yellow", now-3600, 0)

	err := k.EvaluateSuggestionZones(ctx, "LifeDomain")
	if err != nil {
		t.Fatal(err)
	}

	domain, _ := k.GetDomain(ctx, "LifeDomain")
	s := domain.Issues[0].Suggestions[0]
	if s.Color != "yellow" {
		t.Errorf("S1 color = %q, want yellow (dwell time not expired)", s.Color)
	}
}

// ---------- Red stays red before dwell time ----------

func TestRedStaysBeforeDwell(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupLifecycleDomain(t, k, ctx)

	now := ctx.BlockTime().Unix()

	// S1 entered red 1 hour ago.
	setSuggestionColor(t, k, ctx, "LifeDomain", 0, 0, "red", 0, now-3600)

	err := k.EvaluateSuggestionZones(ctx, "LifeDomain")
	if err != nil {
		t.Fatal(err)
	}

	domain, _ := k.GetDomain(ctx, "LifeDomain")
	// S1 should still exist (not deleted yet).
	found := false
	for _, s := range domain.Issues[0].Suggestions {
		if s.Name == "S1" {
			found = true
			if s.Color != "red" {
				t.Errorf("S1 color = %q, want red", s.Color)
			}
		}
	}
	if !found {
		t.Error("S1 should not be deleted yet (red dwell not expired)")
	}
}

// ---------- Custom domain threshold ----------

func TestCustomApprovalThreshold(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupLifecycleDomain(t, k, ctx)

	// Set a high threshold: 50% (5000 bps). Now need 5 stones out of 10.
	domain, _ := k.GetDomain(ctx, "LifeDomain")
	domain.Options.ApprovalThreshold = 5000
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:LifeDomain"), bz)

	// Give S1 only 1 stone — not enough at 50%.
	setSuggestionStones(t, k, ctx, "LifeDomain", 0, 0, 1)

	err := k.EvaluateSuggestionZones(ctx, "LifeDomain")
	if err != nil {
		t.Fatal(err)
	}

	domain, _ = k.GetDomain(ctx, "LifeDomain")
	if domain.Issues[0].Suggestions[0].Color != "yellow" {
		t.Errorf("S1 color = %q, want yellow (1/10 at 50%% threshold)", domain.Issues[0].Suggestions[0].Color)
	}

	// Give S1 5 stones — should be green now.
	setSuggestionStones(t, k, ctx, "LifeDomain", 0, 0, 5)

	k.EvaluateSuggestionZones(ctx, "LifeDomain")

	domain, _ = k.GetDomain(ctx, "LifeDomain")
	if domain.Issues[0].Suggestions[0].Color != "green" {
		t.Errorf("S1 color = %q, want green (5/10 at 50%% threshold)", domain.Issues[0].Suggestions[0].Color)
	}
}

// ---------- Custom dwell time ----------

func TestCustomDwellTime(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupLifecycleDomain(t, k, ctx)

	// Set domain dwell time to 1 hour.
	domain, _ := k.GetDomain(ctx, "LifeDomain")
	domain.Options.DefaultDwellTime = 3600
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:LifeDomain"), bz)

	now := ctx.BlockTime().Unix()

	// S1 entered yellow 2 hours ago — with 1h dwell, should transition to red.
	setSuggestionColor(t, k, ctx, "LifeDomain", 0, 0, "yellow", now-7200, 0)

	err := k.EvaluateSuggestionZones(ctx, "LifeDomain")
	if err != nil {
		t.Fatal(err)
	}

	domain, _ = k.GetDomain(ctx, "LifeDomain")
	if domain.Issues[0].Suggestions[0].Color != "red" {
		t.Errorf("S1 color = %q, want red (1h dwell expired)", domain.Issues[0].Suggestions[0].Color)
	}
}

// ---------- Suggestion-specific dwell time ----------

func TestSuggestionSpecificDwellTime(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupLifecycleDomain(t, k, ctx)

	now := ctx.BlockTime().Unix()

	// Set S1's own dwell time to 30 minutes.
	domain, _ := k.GetDomain(ctx, "LifeDomain")
	domain.Issues[0].Suggestions[0].DwellTime = 1800
	domain.Issues[0].Suggestions[0].Color = "yellow"
	domain.Issues[0].Suggestions[0].EnteredYellowAt = now - 2000 // 33 min ago
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:LifeDomain"), bz)

	err := k.EvaluateSuggestionZones(ctx, "LifeDomain")
	if err != nil {
		t.Fatal(err)
	}

	domain, _ = k.GetDomain(ctx, "LifeDomain")
	if domain.Issues[0].Suggestions[0].Color != "red" {
		t.Errorf("S1 color = %q, want red (30min dwell expired)", domain.Issues[0].Suggestions[0].Color)
	}
}

// ---------- ProcessAllLifecycles ----------

func TestProcessAllLifecycles(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupLifecycleDomain(t, k, ctx)

	// Create a second domain.
	k.CreateDomain(ctx, "Domain2", sdk.AccAddress("admin2"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))
	domain2, _ := k.GetDomain(ctx, "Domain2")
	domain2.Members = []string{"m1", "m2", "m3"}
	domain2.Issues = []Issue{
		{
			Name: "Issue2", Stones: 0, CreationDate: ctx.BlockTime().Unix(),
			Suggestions: []Suggestion{
				{Name: "Prop1", Creator: "m1", Stones: 0, Ratings: []Rating{}, CreationDate: ctx.BlockTime().Unix(), Color: ""},
			},
		},
	}
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain2)
	st.Set([]byte("domain:Domain2"), bz)

	// Process all — both domains should be evaluated.
	k.ProcessAllLifecycles(ctx)

	d1, _ := k.GetDomain(ctx, "LifeDomain")
	d2, _ := k.GetDomain(ctx, "Domain2")

	// Both should have transitioned to yellow (0 stones, below threshold).
	if d1.Issues[0].Suggestions[0].Color != "yellow" {
		t.Errorf("LifeDomain S1 = %q, want yellow", d1.Issues[0].Suggestions[0].Color)
	}
	if d2.Issues[0].Suggestions[0].Color != "yellow" {
		t.Errorf("Domain2 Prop1 = %q, want yellow", d2.Issues[0].Suggestions[0].Color)
	}
}

// ---------- VoteToDelete ----------

func TestVoteToDelete(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupLifecycleDomain(t, k, ctx)

	t.Run("single vote does not delete", func(t *testing.T) {
		deleted, err := k.VoteToDelete(ctx, "LifeDomain", "PolicyA", "S1", "m1")
		if err != nil {
			t.Fatal(err)
		}
		if deleted {
			t.Error("single vote should not delete (need 2/3)")
		}
	})

	t.Run("duplicate vote rejected", func(t *testing.T) {
		_, err := k.VoteToDelete(ctx, "LifeDomain", "PolicyA", "S1", "m1")
		if err == nil {
			t.Fatal("expected error for duplicate vote")
		}
	})

	t.Run("non-member rejected", func(t *testing.T) {
		_, err := k.VoteToDelete(ctx, "LifeDomain", "PolicyA", "S1", "outsider")
		if err == nil {
			t.Fatal("expected error for non-member")
		}
	})

	t.Run("2/3 majority deletes", func(t *testing.T) {
		// m1 already voted. Need 7 members total to reach 7/10 >= 66.67%.
		// (7*10000=70000 >= 10*6667=66670)
		for _, m := range []string{"m2", "m3", "m4", "m5", "m6", "m7"} {
			deleted, err := k.VoteToDelete(ctx, "LifeDomain", "PolicyA", "S1", m)
			if err != nil {
				t.Fatalf("vote by %s failed: %v", m, err)
			}
			if m == "m7" {
				if !deleted {
					t.Error("7/10 votes should reach 2/3 majority and delete")
				}
			} else {
				if deleted {
					t.Errorf("vote by %s should not have triggered delete", m)
				}
			}
		}

		// Verify S1 is gone.
		domain, _ := k.GetDomain(ctx, "LifeDomain")
		for _, s := range domain.Issues[0].Suggestions {
			if s.Name == "S1" {
				t.Error("S1 should have been deleted")
			}
		}
	})
}

// ---------- VoteToDelete edge cases ----------

func TestVoteToDeleteUnknownIssue(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupLifecycleDomain(t, k, ctx)

	_, err := k.VoteToDelete(ctx, "LifeDomain", "NoSuchIssue", "S1", "m1")
	if err == nil {
		t.Fatal("expected error for unknown issue")
	}
}

func TestVoteToDeleteUnknownSuggestion(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupLifecycleDomain(t, k, ctx)

	_, err := k.VoteToDelete(ctx, "LifeDomain", "PolicyA", "NoSuchSugg", "m1")
	if err == nil {
		t.Fatal("expected error for unknown suggestion")
	}
}

// ---------- Full lifecycle: green → yellow → red → deleted ----------

func TestFullLifecycleProgression(t *testing.T) {
	k, baseCtx := setupKeeper(t)
	setupLifecycleDomain(t, k, baseCtx)

	// Give S1 1 stone → green.
	setSuggestionStones(t, k, baseCtx, "LifeDomain", 0, 0, 1)
	k.EvaluateSuggestionZones(baseCtx, "LifeDomain")

	domain, _ := k.GetDomain(baseCtx, "LifeDomain")
	if domain.Issues[0].Suggestions[0].Color != "green" {
		t.Fatalf("expected green, got %q", domain.Issues[0].Suggestions[0].Color)
	}

	// Remove stone → should go yellow.
	setSuggestionStones(t, k, baseCtx, "LifeDomain", 0, 0, 0)
	k.EvaluateSuggestionZones(baseCtx, "LifeDomain")

	domain, _ = k.GetDomain(baseCtx, "LifeDomain")
	if domain.Issues[0].Suggestions[0].Color != "yellow" {
		t.Fatalf("expected yellow, got %q", domain.Issues[0].Suggestions[0].Color)
	}

	// Advance time past dwell (1 day + 1 second).
	ctx2 := baseCtx.WithBlockTime(baseCtx.BlockTime().Add(86401 * time.Second))
	k.EvaluateSuggestionZones(ctx2, "LifeDomain")

	domain, _ = k.GetDomain(ctx2, "LifeDomain")
	if domain.Issues[0].Suggestions[0].Color != "red" {
		t.Fatalf("expected red, got %q", domain.Issues[0].Suggestions[0].Color)
	}

	// Advance time past another dwell period.
	ctx3 := ctx2.WithBlockTime(ctx2.BlockTime().Add(86401 * time.Second))
	k.EvaluateSuggestionZones(ctx3, "LifeDomain")

	domain, _ = k.GetDomain(ctx3, "LifeDomain")
	// S1 should be deleted, only S2 remains.
	foundS1 := false
	for _, s := range domain.Issues[0].Suggestions {
		if s.Name == "S1" {
			foundS1 = true
		}
	}
	if foundS1 {
		t.Error("S1 should have been auto-deleted after red dwell expired")
	}
}
