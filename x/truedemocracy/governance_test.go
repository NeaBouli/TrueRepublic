package truedemocracy

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// setupGovernanceDomain creates a domain with 10 members for governance tests.
func setupGovernanceDomain(t *testing.T, k Keeper, ctx sdk.Context) {
	t.Helper()
	k.CreateDomain(ctx, "GovDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1_000_000)))

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

	t.Run("highest stone count becomes admin", func(t *testing.T) {
		addrs := realTestAddrs(4)
		_, alice, bob, carol := addrs[0], addrs[1], addrs[2], addrs[3]
		setupBech32ElectionDomain(t, k, ctx, "ElectDomain", addrs)
		electBobByBech32Stones(t, k, ctx, "ElectDomain", alice, bob, carol)

		if err := k.ElectAdmin(ctx, "ElectDomain"); err != nil {
			t.Fatal(err)
		}

		domain, _ := k.GetDomain(ctx, "ElectDomain")
		if !domain.Admin.Equals(bob) {
			t.Errorf("admin = %q, want decoded winner %s", domain.Admin.String(), bob.String())
		}
	})

	t.Run("tie keeps first member in list order", func(t *testing.T) {
		addrs := realTestAddrs(3)
		alice, bob := addrs[1], addrs[2]
		setupBech32ElectionDomain(t, k, ctx, "TieDomain", addrs)

		// alice and bob tie with one stone each; alice appears earlier in the
		// member list and must win deterministically.
		if err := k.PlaceStoneOnMember(ctx, "TieDomain", bob.String(), alice.String()); err != nil {
			t.Fatal(err)
		}
		if err := k.PlaceStoneOnMember(ctx, "TieDomain", alice.String(), bob.String()); err != nil {
			t.Fatal(err)
		}

		if err := k.ElectAdmin(ctx, "TieDomain"); err != nil {
			t.Fatal(err)
		}

		domain, _ := k.GetDomain(ctx, "TieDomain")
		if !domain.Admin.Equals(alice) {
			t.Errorf("tie-break admin = %q, want earlier member %s", domain.Admin.String(), alice.String())
		}
	})

	t.Run("no stones means no change", func(t *testing.T) {
		// Create a fresh domain with no stones.
		k.CreateDomain(ctx, "EmptyDomain", sdk.AccAddress("origadmin"), sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 100)))
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
		k.CreateDomain(ctx, "FixedAdmin", sdk.AccAddress("boss"), sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 100)))
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

// ---------- GH-304: exclusion of the current admin ----------

// TestVoteToExcludeRejectsCurrentAdmin proves normal governance can never
// exclude the current domain admin: a full 2/3 attempt is rejected with a
// deterministic invalid-request error before any vote key or domain state is
// written, so exclusion cannot create admin_not_in_members/quarantined state.
// Once a replacement admin has been elected, the former admin is an ordinary
// member and can be excluded under the existing threshold semantics (GH-304).
func TestVoteToExcludeRejectsCurrentAdmin(t *testing.T) {
	k, ctx := setupKeeper(t)
	addrs := realTestAddrs(10)
	admin, replacement := addrs[0], addrs[1]
	setupBech32ElectionDomain(t, k, ctx, "AdminExclusionDomain", addrs)

	before, _ := k.GetDomain(ctx, "AdminExclusionDomain")

	// A legacy all-uppercase member spelling is semantically the same address.
	// The admin guard must compare decoded identity, not raw text, and reject
	// before writing an exclusion vote. Restore canonical state afterward so
	// the remainder exercises the normal election/exclusion flow.
	legacy := before
	legacy.Members = append([]string(nil), before.Members...)
	legacy.Members[0] = strings.ToUpper(admin.String())
	saveDomain(t, k, ctx, legacy)
	legacyTarget := legacy.Members[0]
	if excluded, err := k.VoteToExclude(ctx, "AdminExclusionDomain", legacyTarget, replacement.String()); err == nil || excluded {
		t.Fatalf("legacy-spelled current admin exclusion = (%t, %v), want rejected", excluded, err)
	}
	if key := excludeVoteKey("AdminExclusionDomain", legacyTarget, replacement.String()); ctx.KVStore(k.StoreKey).Has(key) {
		t.Fatalf("legacy-spelled admin vote key %q must not be written", key)
	}
	saveDomain(t, k, ctx, before)

	// 10 members → 9 eligible voters → the 2/3 threshold is 7 votes
	// (7*10000 >= 9*6667). Every vote targeting the current admin must be
	// rejected before any state is written.
	attemptVoters := addrs[1:8]
	for _, voter := range attemptVoters {
		excluded, err := k.VoteToExclude(ctx, "AdminExclusionDomain", admin.String(), voter.String())
		if err == nil {
			t.Fatalf("vote by %s to exclude the current admin must be rejected", voter.String())
		}
		if excluded {
			t.Fatalf("vote by %s must not report exclusion", voter.String())
		}
		if !errors.Is(err, sdkerrors.ErrInvalidRequest) {
			t.Fatalf("error = %v, want deterministic wrapped ErrInvalidRequest", err)
		}
		if !strings.Contains(err.Error(), "elect a replacement admin first") {
			t.Fatalf("error = %v, want guidance to elect a replacement admin first", err)
		}
	}

	// No exclusion vote key was persisted for any attempted voter.
	store := ctx.KVStore(k.StoreKey)
	for _, voter := range attemptVoters {
		if key := excludeVoteKey("AdminExclusionDomain", admin.String(), voter.String()); store.Has(key) {
			t.Fatalf("vote key %q must not exist after rejection", key)
		}
	}

	// The persisted domain stayed byte-identical: admin remains in Members.
	assertDomainUnchanged(t, k, ctx, "AdminExclusionDomain", before)

	// Elect the replacement admin through the normal stone election.
	for _, voter := range addrs[2:4] {
		if err := k.PlaceStoneOnMember(ctx, "AdminExclusionDomain", replacement.String(), voter.String()); err != nil {
			t.Fatalf("place stone on replacement: %v", err)
		}
	}
	if err := k.ElectAdmin(ctx, "AdminExclusionDomain"); err != nil {
		t.Fatalf("ElectAdmin: %v", err)
	}
	domain, _ := k.GetDomain(ctx, "AdminExclusionDomain")
	if !domain.Admin.Equals(replacement) {
		t.Fatalf("admin = %q, want elected replacement %s", domain.Admin.String(), replacement.String())
	}

	// The former admin is now an ordinary member: the same 7 of 9 voters
	// exclude them under the existing threshold semantics.
	var excluded bool
	for i, voter := range attemptVoters {
		var err error
		excluded, err = k.VoteToExclude(ctx, "AdminExclusionDomain", admin.String(), voter.String())
		if err != nil {
			t.Fatalf("vote %d by %s against the former admin failed: %v", i, voter.String(), err)
		}
	}
	if !excluded {
		t.Fatal("7/9 votes must exclude the former admin under the existing threshold")
	}

	domain, _ = k.GetDomain(ctx, "AdminExclusionDomain")
	if containsString(domain.Members, admin.String()) {
		t.Error("former admin should have been excluded from the member list")
	}
	if len(domain.Members) != 9 {
		t.Errorf("members count = %d, want 9", len(domain.Members))
	}
	// The new admin is still in the member set: the invariant holds.
	if reason := classifyDomainAdmin(domain); reason != "" {
		t.Errorf("classifyDomainAdmin = %q, want no finding after election and exclusion", reason)
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
	k.CreateDomain(ctx, "LinkDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 500_000)))

	domain, _ := k.GetDomain(ctx, "LinkDomain")
	domain.Members = append(domain.Members, "alice")
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:LinkDomain"), bz)

	// Submit with external link.
	err := k.SubmitProposal(ctx, "LinkDomain", "Policy", "Plan", "alice",
		sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1000)), "https://forum.example.com/policy")
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
	k.CreateDomain(ctx, "NoLinkDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 500_000)))

	domain, _ := k.GetDomain(ctx, "NoLinkDomain")
	domain.Members = append(domain.Members, "alice")
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:NoLinkDomain"), bz)

	// Submit without external link.
	err := k.SubmitProposal(ctx, "NoLinkDomain", "Policy", "Plan", "alice",
		sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1000)), "")
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
	addrs := realTestAddrs(4)
	_, alice, bob, carol := addrs[0], addrs[1], addrs[2], addrs[3]
	setupBech32ElectionDomain(t, k, ctx, "GovDomain", addrs)

	// Vote for alice as admin.
	if err := k.PlaceStoneOnMember(ctx, "GovDomain", alice.String(), bob.String()); err != nil {
		t.Fatal(err)
	}
	if err := k.PlaceStoneOnMember(ctx, "GovDomain", alice.String(), carol.String()); err != nil {
		t.Fatal(err)
	}

	if err := k.ProcessGovernance(ctx); err != nil {
		t.Fatalf("ProcessGovernance: %v", err)
	}

	domain, _ := k.GetDomain(ctx, "GovDomain")
	if !domain.Admin.Equals(alice) {
		t.Errorf("admin = %q, want %s after governance processing", domain.Admin.String(), alice.String())
	}
}

// ---------- GH-304: real bech32 election regression ----------

// realTestAddrs returns deterministic 20-byte account addresses so governance
// tests use the production bech32 member representation. ASCII pseudo-
// addresses mask address-handling defects because raw bytes and canonical
// text coincide for them (GH-304).
func realTestAddrs(n int) []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, n)
	for i := 0; i < n; i++ {
		b := make([]byte, 20)
		b[0] = byte(i + 1)
		addrs[i] = sdk.AccAddress(b)
	}
	return addrs
}

// setupBech32ElectionDomain creates an electable domain whose member list
// holds canonical bech32 addresses, exactly as production stores them.
func setupBech32ElectionDomain(t *testing.T, k Keeper, ctx sdk.Context, name string, members []sdk.AccAddress) {
	t.Helper()
	k.CreateDomain(ctx, name, members[0], sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1_000_000)))
	domain, found := k.GetDomain(ctx, name)
	if !found {
		t.Fatalf("domain %s not found after creation", name)
	}
	domain.Members = make([]string, len(members))
	for i, member := range members {
		domain.Members[i] = member.String()
	}
	domain.Options.AdminElectable = true
	saveDomain(t, k, ctx, domain)
}

// electBobByBech32Stones lets alice and carol place their stones on bob and
// bob place his stone on alice, so bob wins the election.
func electBobByBech32Stones(t *testing.T, k Keeper, ctx sdk.Context, domainName string, alice, bob, carol sdk.AccAddress) {
	t.Helper()
	for _, voter := range []sdk.AccAddress{alice, carol} {
		if err := k.PlaceStoneOnMember(ctx, domainName, bob.String(), voter.String()); err != nil {
			t.Fatalf("place stone on bob: %v", err)
		}
	}
	if err := k.PlaceStoneOnMember(ctx, domainName, alice.String(), bob.String()); err != nil {
		t.Fatalf("place stone on alice: %v", err)
	}
}

func TestElectAdminRealBech32StoresDecodedAdmin(t *testing.T) {
	k, ctx := setupKeeper(t)
	addrs := realTestAddrs(4)
	_, alice, bob, carol := addrs[0], addrs[1], addrs[2], addrs[3]
	setupBech32ElectionDomain(t, k, ctx, "Bech32Domain", addrs)
	electBobByBech32Stones(t, k, ctx, "Bech32Domain", alice, bob, carol)

	if err := k.ElectAdmin(ctx, "Bech32Domain"); err != nil {
		t.Fatalf("ElectAdmin: %v", err)
	}

	domain, _ := k.GetDomain(ctx, "Bech32Domain")
	if !domain.Admin.Equals(bob) {
		t.Fatalf("admin = %q (raw %x), want the decoded address of winner %s", domain.Admin.String(), []byte(domain.Admin), bob.String())
	}
	if !containsString(domain.Members, domain.Admin.String()) {
		t.Fatalf("stored admin %q is not in the canonical member set", domain.Admin.String())
	}
}

func TestElectAdminRealBech32RecognizesNoChange(t *testing.T) {
	k, ctx := setupKeeper(t)
	addrs := realTestAddrs(3)
	bob, alice, carol := addrs[0], addrs[1], addrs[2]

	// bob is already admin and also wins the stone count: ElectAdmin must be
	// a strict no-op.
	setupBech32ElectionDomain(t, k, ctx, "StableDomain", addrs)
	for _, voter := range []sdk.AccAddress{alice, carol} {
		if err := k.PlaceStoneOnMember(ctx, "StableDomain", bob.String(), voter.String()); err != nil {
			t.Fatalf("place stone on bob: %v", err)
		}
	}

	before, _ := k.GetDomain(ctx, "StableDomain")
	if err := k.ElectAdmin(ctx, "StableDomain"); err != nil {
		t.Fatalf("ElectAdmin: %v", err)
	}
	after, _ := k.GetDomain(ctx, "StableDomain")

	if !after.Admin.Equals(bob) {
		t.Fatalf("no-change election rewrote admin to %q (raw %x), want %s untouched", after.Admin.String(), []byte(after.Admin), bob.String())
	}
	want := k.cdc.MustMarshalLengthPrefixed(&before)
	got := k.cdc.MustMarshalLengthPrefixed(&after)
	if !bytes.Equal(got, want) {
		t.Fatal("no-change election mutated persisted domain state")
	}
}

func TestElectAdminRealBech32ExportImportValid(t *testing.T) {
	k, ctx := setupKeeper(t)
	addrs := realTestAddrs(4)
	_, alice, bob, carol := addrs[0], addrs[1], addrs[2], addrs[3]
	setupBech32ElectionDomain(t, k, ctx, "Bech32Domain", addrs)
	electBobByBech32Stones(t, k, ctx, "Bech32Domain", alice, bob, carol)

	if err := k.ElectAdmin(ctx, "Bech32Domain"); err != nil {
		t.Fatalf("ElectAdmin: %v", err)
	}

	module := NewAppModule(k.cdc, k)
	exported := module.ExportGenesis(ctx, nil)
	var genesis GenesisState
	if err := json.Unmarshal(exported, &genesis); err != nil {
		t.Fatalf("unmarshal exported genesis: %v", err)
	}
	if err := ValidateGenesisState(genesis); err != nil {
		t.Fatalf("exported genesis after real bech32 election must validate for import: %v", err)
	}
}

// ---------- GH-304: quarantined corruption and ignored legacy candidates ----------

// invalidLegacyMember is an unparseable historical member string as found in
// legacy/corrupted domain state (GH-304).
const invalidLegacyMember = "not-a-bech32-address"

// addInvalidMemberWithStones appends the unparseable legacy member string to
// the domain and lets the given voters place their stones on it, so the
// invalid entry would win the election if it were ranked (GH-304).
func addInvalidMemberWithStones(t *testing.T, k Keeper, ctx sdk.Context, domainName string, voters ...sdk.AccAddress) {
	t.Helper()
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		t.Fatalf("domain %s not found", domainName)
	}
	domain.Members = append(domain.Members, invalidLegacyMember)
	saveDomain(t, k, ctx, domain)
	for _, voter := range voters {
		if err := k.PlaceStoneOnMember(ctx, domainName, invalidLegacyMember, voter.String()); err != nil {
			t.Fatalf("place stone on invalid member: %v", err)
		}
	}
}

// corruptDomainAdmin rewrites the stored admin to the GH-304 double-encoding
// signature: the bech32 text bytes of a member instead of the decoded address.
func corruptDomainAdmin(t *testing.T, k Keeper, ctx sdk.Context, domainName string, member sdk.AccAddress) {
	t.Helper()
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		t.Fatalf("domain %s not found", domainName)
	}
	domain.Admin = sdk.AccAddress([]byte(member.String()))
	saveDomain(t, k, ctx, domain)
}

func assertDomainUnchanged(t *testing.T, k Keeper, ctx sdk.Context, domainName string, before Domain) {
	t.Helper()
	after, found := k.GetDomain(ctx, domainName)
	if !found {
		t.Fatalf("domain %s disappeared", domainName)
	}
	want := k.cdc.MustMarshalLengthPrefixed(&before)
	got := k.cdc.MustMarshalLengthPrefixed(&after)
	if !bytes.Equal(got, want) {
		t.Fatalf("persisted domain %s state mutated", domainName)
	}
}

// requireQuarantineEvent fails unless a domain_admin_quarantine event with the
// exact domain and reason attributes was emitted.
func requireQuarantineEvent(t *testing.T, ctx sdk.Context, domainName, reason string) {
	t.Helper()
	for _, event := range ctx.EventManager().Events() {
		if event.Type != EventTypeDomainAdminQuarantine {
			continue
		}
		var gotDomain, gotReason string
		for _, attr := range event.Attributes {
			switch attr.Key {
			case "domain":
				gotDomain = attr.Value
			case "reason":
				gotReason = attr.Value
			}
		}
		if gotDomain == domainName && gotReason == reason {
			return
		}
	}
	t.Fatalf("quarantine event for domain %s with reason %s not emitted; got %v", domainName, reason, ctx.EventManager().Events())
}

// TestElectAdminIgnoresInvalidLegacyCandidate proves unparseable and
// noncanonical legacy member strings never halt the election and never win:
// the valid candidate with the next-highest stone count is elected
// deterministically (GH-304).
func TestElectAdminIgnoresInvalidLegacyCandidate(t *testing.T) {
	k, ctx := setupKeeper(t)
	addrs := realTestAddrs(6)
	admin, alice, bob, carol, dave, erin := addrs[0], addrs[1], addrs[2], addrs[3], addrs[4], addrs[5]
	setupBech32ElectionDomain(t, k, ctx, "LegacyDomain", addrs)

	// Both invalid entries would beat bob 2:1 if either were ranked. Full
	// uppercase bech32 decodes successfully but is not canonical storage text.
	addInvalidMemberWithStones(t, k, ctx, "LegacyDomain", alice, carol)
	uppercaseCandidate := strings.ToUpper(dave.String())
	domain, _ := k.GetDomain(ctx, "LegacyDomain")
	domain.Members = append(domain.Members, uppercaseCandidate)
	saveDomain(t, k, ctx, domain)
	for _, voter := range []sdk.AccAddress{dave, erin} {
		if err := k.PlaceStoneOnMember(ctx, "LegacyDomain", uppercaseCandidate, voter.String()); err != nil {
			t.Fatalf("place stone on noncanonical member: %v", err)
		}
	}
	if err := k.PlaceStoneOnMember(ctx, "LegacyDomain", bob.String(), admin.String()); err != nil {
		t.Fatalf("place stone on bob: %v", err)
	}

	if err := k.ElectAdmin(ctx, "LegacyDomain"); err != nil {
		t.Fatalf("ElectAdmin must ignore the invalid candidate, got %v", err)
	}
	domain, _ = k.GetDomain(ctx, "LegacyDomain")
	if !domain.Admin.Equals(bob) {
		t.Fatalf("admin = %q, want highest-count valid candidate %s", domain.Admin.String(), bob.String())
	}
}

// TestElectAdminOnlyInvalidCandidatesMakeNoChange proves an election where
// only invalid legacy entries carry stones is a strict no-op (GH-304).
func TestElectAdminOnlyInvalidCandidatesMakeNoChange(t *testing.T) {
	k, ctx := setupKeeper(t)
	addrs := realTestAddrs(2)
	_, alice := addrs[0], addrs[1]
	setupBech32ElectionDomain(t, k, ctx, "OnlyInvalidDomain", addrs)
	addInvalidMemberWithStones(t, k, ctx, "OnlyInvalidDomain", alice)

	before, _ := k.GetDomain(ctx, "OnlyInvalidDomain")
	if err := k.ElectAdmin(ctx, "OnlyInvalidDomain"); err != nil {
		t.Fatalf("ElectAdmin with only invalid candidates must be a no-op, got %v", err)
	}
	assertDomainUnchanged(t, k, ctx, "OnlyInvalidDomain", before)
}

// TestElectAdminCorruptAdminFailsClosedWithoutMutation proves a corrupted
// stored admin returns the dedicated deterministic integrity error and that
// ElectAdmin never silently repairs the corrupt state, even when a valid
// winner with stones exists (GH-304).
func TestElectAdminCorruptAdminFailsClosedWithoutMutation(t *testing.T) {
	k, ctx := setupKeeper(t)
	addrs := realTestAddrs(4)
	_, alice, bob, carol := addrs[0], addrs[1], addrs[2], addrs[3]
	setupBech32ElectionDomain(t, k, ctx, "CorruptAdminDomain", addrs)
	electBobByBech32Stones(t, k, ctx, "CorruptAdminDomain", alice, bob, carol)
	corruptDomainAdmin(t, k, ctx, "CorruptAdminDomain", alice)

	before, _ := k.GetDomain(ctx, "CorruptAdminDomain")
	err := k.ElectAdmin(ctx, "CorruptAdminDomain")
	if err == nil {
		t.Fatal("ElectAdmin must fail closed on a corrupt stored admin")
	}
	if !errors.Is(err, sdkerrors.ErrInvalidAddress) {
		t.Fatalf("ElectAdmin error = %v, want deterministic wrapped ErrInvalidAddress", err)
	}
	var integrityErr *domainAdminIntegrityError
	if !errors.As(err, &integrityErr) {
		t.Fatalf("ElectAdmin error = %v, want dedicated domainAdminIntegrityError", err)
	}
	if integrityErr.DomainName != "CorruptAdminDomain" || integrityErr.Reason != DomainAdminReasonBech32Text {
		t.Fatalf("integrity error = %+v, want domain CorruptAdminDomain with reason %s", integrityErr, DomainAdminReasonBech32Text)
	}

	// No silent legacy repair: the corrupt admin bytes and every other
	// persisted field stay byte-identical although stones exist.
	assertDomainUnchanged(t, k, ctx, "CorruptAdminDomain", before)
}

// TestProcessGovernanceQuarantinesCorruptDomainAndCommitsHealthy proves the
// GH-304 domain-local quarantine: the domain with the corrupt stored admin is
// skipped with a stable event — no election, no cleanup, no repair even though
// a valid winner with stones exists — while the healthy domain's election and
// cleanup still commit in the same pass.
//
// The earlier full-pass rollback test is intentionally retired: after this
// remediation no non-quarantine error is reachable in production code
// (invalid candidates are ignored, admin corruption quarantines, the
// winner-parse guard is unreachable for filtered candidates, and
// CleanupInactiveIssues never fails), so a genuine abort cannot be induced
// without test-only production hooks, which GH-304 forbids. The commit
// boundary is instead proven directly here — the quarantined domain remains
// byte-identical while the healthy domain's work is persisted — and module.go
// still propagates any returned ProcessGovernance error to EndBlock.
func TestProcessGovernanceQuarantinesCorruptDomainAndCommitsHealthy(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Key order: "ACorruptDomain" is processed before "ZHealthyDomain", so
	// the pass must continue past the quarantined domain.
	addrs := realTestAddrs(4)
	_, alice, bob, carol := addrs[0], addrs[1], addrs[2], addrs[3]
	setupBech32ElectionDomain(t, k, ctx, "ZHealthyDomain", addrs)
	electBobByBech32Stones(t, k, ctx, "ZHealthyDomain", alice, bob, carol)

	// An expired issue proves cleanup commits for the healthy domain and is
	// skipped for the quarantined one.
	expired := Issue{Name: "StaleIssue", CreationDate: 1, Suggestions: []Suggestion{}}
	healthy, _ := k.GetDomain(ctx, "ZHealthyDomain")
	healthy.Issues = append(healthy.Issues, expired)
	saveDomain(t, k, ctx, healthy)

	setupBech32ElectionDomain(t, k, ctx, "ACorruptDomain", addrs)
	corruptDomainAdmin(t, k, ctx, "ACorruptDomain", alice)
	corrupt, _ := k.GetDomain(ctx, "ACorruptDomain")
	corrupt.Issues = append(corrupt.Issues, expired)
	saveDomain(t, k, ctx, corrupt)
	// A valid winner with stones exists in the corrupt domain too.
	if err := k.PlaceStoneOnMember(ctx, "ACorruptDomain", bob.String(), alice.String()); err != nil {
		t.Fatalf("place stone on bob in corrupt domain: %v", err)
	}
	corruptBefore, _ := k.GetDomain(ctx, "ACorruptDomain")

	if err := k.ProcessGovernance(ctx); err != nil {
		t.Fatalf("ProcessGovernance must quarantine instead of failing the pass: %v", err)
	}

	// Healthy domain: election committed (bob) and the stale issue cleaned up.
	healthyAfter, _ := k.GetDomain(ctx, "ZHealthyDomain")
	if !healthyAfter.Admin.Equals(bob) {
		t.Fatalf("healthy domain admin = %q, want elected %s", healthyAfter.Admin.String(), bob.String())
	}
	if len(healthyAfter.Issues) != 0 {
		t.Fatalf("healthy domain issues = %v, want the stale issue cleaned up", healthyAfter.Issues)
	}

	// Corrupt domain: no election, no cleanup, no silent repair.
	assertDomainUnchanged(t, k, ctx, "ACorruptDomain", corruptBefore)

	// Stable quarantine event with the domain and reason.
	requireQuarantineEvent(t, ctx, "ACorruptDomain", DomainAdminReasonBech32Text)
}

// TestEndBlockQuarantinesCorruptDomainWithoutHalting proves a corrupt stored
// domain admin neither panics nor halts EndBlock: the block succeeds, the
// corrupt domain is left untouched, and the healthy domain's election commits.
func TestEndBlockQuarantinesCorruptDomainWithoutHalting(t *testing.T) {
	k, ctx, _ := setupKeeperWithBank(t)
	addrs := realTestAddrs(4)
	_, alice, bob, carol := addrs[0], addrs[1], addrs[2], addrs[3]
	setupBech32ElectionDomain(t, k, ctx, "ZHealthyDomain", addrs)
	electBobByBech32Stones(t, k, ctx, "ZHealthyDomain", alice, bob, carol)

	setupBech32ElectionDomain(t, k, ctx, "ACorruptDomain", addrs)
	corruptDomainAdmin(t, k, ctx, "ACorruptDomain", alice)
	corruptBefore, _ := k.GetDomain(ctx, "ACorruptDomain")

	module := NewAppModule(k.cdc, k)
	if _, err := module.EndBlock(ctx); err != nil {
		t.Fatalf("EndBlock must not halt on a corrupt domain admin: %v", err)
	}

	assertDomainUnchanged(t, k, ctx, "ACorruptDomain", corruptBefore)
	healthyAfter, _ := k.GetDomain(ctx, "ZHealthyDomain")
	if !healthyAfter.Admin.Equals(bob) {
		t.Fatalf("healthy domain admin = %q, want elected %s", healthyAfter.Admin.String(), bob.String())
	}
	requireQuarantineEvent(t, ctx, "ACorruptDomain", DomainAdminReasonBech32Text)
}

// TestElectAdminRealBech32KeepsAdminOperationsIntact proves that after a valid
// real-bech32 election every admin-gated operation and the export/import
// round-trip keep working with the decoded winner address (GH-304).
func TestElectAdminRealBech32KeepsAdminOperationsIntact(t *testing.T) {
	k, ctx, bank := setupKeeperWithBank(t)
	addrs := realTestAddrs(5)
	_, alice, bob, carol, newMember := addrs[0], addrs[1], addrs[2], addrs[3], addrs[4]
	setupBech32ElectionDomain(t, k, ctx, "OpsDomain", addrs[:4])
	electBobByBech32Stones(t, k, ctx, "OpsDomain", alice, bob, carol)

	if err := k.ElectAdmin(ctx, "OpsDomain"); err != nil {
		t.Fatalf("ElectAdmin: %v", err)
	}
	domain, _ := k.GetDomain(ctx, "OpsDomain")
	if !domain.Admin.Equals(bob) {
		t.Fatalf("admin = %q, want decoded winner %s", domain.Admin.String(), bob.String())
	}

	// Member administration: only the decoded winner is authorized.
	if err := k.AddMember(ctx, "OpsDomain", newMember.String(), bob); err != nil {
		t.Fatalf("AddMember by elected admin: %v", err)
	}
	outsider := sdk.AccAddress(bytes.Repeat([]byte{0xff}, 20))
	if err := k.AddMember(ctx, "OpsDomain", outsider.String(), alice); err == nil {
		t.Fatal("AddMember by non-admin must be rejected")
	}

	// Permission-register administration.
	if err := k.JoinPermissionRegister(ctx, "OpsDomain", bob.String(), testPubKey("ops-domain-key")); err != nil {
		t.Fatalf("JoinPermissionRegister by elected admin member: %v", err)
	}
	if err := k.PurgePermissionRegister(ctx, "OpsDomain", alice); err == nil {
		t.Fatal("PurgePermissionRegister by non-admin must be rejected")
	}
	if err := k.PurgePermissionRegister(ctx, "OpsDomain", bob); err != nil {
		t.Fatalf("PurgePermissionRegister by elected admin: %v", err)
	}

	// Treasury authorization.
	bank.fundModule(ModuleName, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1_000_000)))
	if err := k.WithdrawFromDomain(ctx, "OpsDomain", bob, sdk.NewInt64Coin(PNYXDenom, 100), alice); err == nil {
		t.Fatal("WithdrawFromDomain authorized by non-admin must be rejected")
	}
	if err := k.WithdrawFromDomain(ctx, "OpsDomain", bob, sdk.NewInt64Coin(PNYXDenom, 100), bob); err != nil {
		t.Fatalf("WithdrawFromDomain authorized by elected admin: %v", err)
	}

	// Export/import validity after the election and the admin operations.
	module := NewAppModule(k.cdc, k)
	exported := module.ExportGenesis(ctx, nil)
	var genesis GenesisState
	if err := json.Unmarshal(exported, &genesis); err != nil {
		t.Fatalf("unmarshal exported genesis: %v", err)
	}
	if err := ValidateGenesisState(genesis); err != nil {
		t.Fatalf("exported genesis after election and admin operations must validate: %v", err)
	}
}
