package truedemocracy

import (
	"encoding/hex"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// domainKey generates a deterministic ed25519 private key for domain voting.
func domainKey(seed string) *ed25519.PrivKey {
	return ed25519.GenPrivKeyFromSecret([]byte(seed))
}

// setupDomainWithIssue creates a domain with members and a proposal for testing.
func setupDomainWithIssue(t *testing.T, k Keeper, ctx sdk.Context) {
	t.Helper()
	k.CreateDomain(ctx, "AnonDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500_000)))

	// Add members.
	domain, _ := k.GetDomain(ctx, "AnonDomain")
	domain.Members = append(domain.Members, "alice", "bob", "charlie")
	// Add an issue with a suggestion.
	domain.Issues = []Issue{
		{
			Name:   "Climate",
			Stones: 0,
			Suggestions: []Suggestion{
				{Name: "GreenDeal", Creator: "alice", Ratings: []Rating{}, Stones: 0},
			},
		},
	}
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:AnonDomain"), bz)
}

// ---------- JoinPermissionRegister ----------

func TestJoinPermissionRegister(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupDomainWithIssue(t, k, ctx)

	aliceKey := domainKey("alice-domain-key")

	t.Run("happy path", func(t *testing.T) {
		err := k.JoinPermissionRegister(ctx, "AnonDomain", "alice", aliceKey.PubKey().Bytes())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		domain, _ := k.GetDomain(ctx, "AnonDomain")
		if len(domain.PermissionReg) != 1 {
			t.Fatalf("permission register length = %d, want 1", len(domain.PermissionReg))
		}
		wantHex := hex.EncodeToString(aliceKey.PubKey().Bytes())
		if domain.PermissionReg[0] != wantHex {
			t.Errorf("registered key = %s, want %s", domain.PermissionReg[0], wantHex)
		}
	})

	t.Run("non-member rejected", func(t *testing.T) {
		outsiderKey := domainKey("outsider-key")
		err := k.JoinPermissionRegister(ctx, "AnonDomain", "outsider", outsiderKey.PubKey().Bytes())
		if err == nil {
			t.Fatal("expected error for non-member")
		}
	})

	t.Run("duplicate key rejected", func(t *testing.T) {
		err := k.JoinPermissionRegister(ctx, "AnonDomain", "alice", aliceKey.PubKey().Bytes())
		if err == nil {
			t.Fatal("expected error for duplicate key")
		}
	})

	t.Run("bad key length rejected", func(t *testing.T) {
		err := k.JoinPermissionRegister(ctx, "AnonDomain", "bob", []byte("short"))
		if err == nil {
			t.Fatal("expected error for bad key length")
		}
	})

	t.Run("unknown domain rejected", func(t *testing.T) {
		bobKey := domainKey("bob-domain-key")
		err := k.JoinPermissionRegister(ctx, "NoSuchDomain", "bob", bobKey.PubKey().Bytes())
		if err == nil {
			t.Fatal("expected error for unknown domain")
		}
	})
}

// ---------- PurgePermissionRegister ----------

func TestPurgePermissionRegister(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupDomainWithIssue(t, k, ctx)

	// Register two keys.
	aliceKey := domainKey("alice-purge-key")
	bobKey := domainKey("bob-purge-key")
	k.JoinPermissionRegister(ctx, "AnonDomain", "alice", aliceKey.PubKey().Bytes())
	k.JoinPermissionRegister(ctx, "AnonDomain", "bob", bobKey.PubKey().Bytes())

	t.Run("non-admin rejected", func(t *testing.T) {
		err := k.PurgePermissionRegister(ctx, "AnonDomain", sdk.AccAddress("not-admin"))
		if err == nil {
			t.Fatal("expected error for non-admin purge")
		}
		// Keys should still be there.
		domain, _ := k.GetDomain(ctx, "AnonDomain")
		if len(domain.PermissionReg) != 2 {
			t.Errorf("permission register length = %d, want 2", len(domain.PermissionReg))
		}
	})

	t.Run("admin can purge", func(t *testing.T) {
		err := k.PurgePermissionRegister(ctx, "AnonDomain", sdk.AccAddress("admin1"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		domain, _ := k.GetDomain(ctx, "AnonDomain")
		if len(domain.PermissionReg) != 0 {
			t.Errorf("permission register length = %d, want 0 after purge", len(domain.PermissionReg))
		}
	})

	t.Run("keys are unauthorized after purge", func(t *testing.T) {
		aliceHex := hex.EncodeToString(aliceKey.PubKey().Bytes())
		if k.IsKeyAuthorized(ctx, "AnonDomain", aliceHex) {
			t.Error("alice's key should be unauthorized after purge")
		}
	})
}

// ---------- Anonymous Rating ----------

func TestAnonymousRating(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupDomainWithIssue(t, k, ctx)

	aliceKey := domainKey("alice-vote-key")
	k.JoinPermissionRegister(ctx, "AnonDomain", "alice", aliceKey.PubKey().Bytes())

	t.Run("happy path", func(t *testing.T) {
		reward, cache, err := k.RateProposal(ctx, "AnonDomain", "Climate", "GreenDeal", 3, aliceKey)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reward.AmountOf("pnyx").IsPositive() {
			t.Error("reward should be positive")
		}
		if cache["avg_rating"] != 3 {
			t.Errorf("avg_rating = %v, want 3", cache["avg_rating"])
		}

		// Verify rating is stored with domain key, not avatar name.
		domain, _ := k.GetDomain(ctx, "AnonDomain")
		ratings := domain.Issues[0].Suggestions[0].Ratings
		if len(ratings) != 1 {
			t.Fatalf("ratings count = %d, want 1", len(ratings))
		}
		wantHex := hex.EncodeToString(aliceKey.PubKey().Bytes())
		if ratings[0].DomainPubKeyHex != wantHex {
			t.Errorf("rating key = %s, want %s", ratings[0].DomainPubKeyHex, wantHex)
		}
	})

	t.Run("unregistered key rejected", func(t *testing.T) {
		unregKey := domainKey("unregistered-key")
		_, _, err := k.RateProposal(ctx, "AnonDomain", "Climate", "GreenDeal", 2, unregKey)
		if err == nil {
			t.Fatal("expected error for unregistered domain key")
		}
	})

	t.Run("nil key rejected", func(t *testing.T) {
		_, _, err := k.RateProposal(ctx, "AnonDomain", "Climate", "GreenDeal", 2, nil)
		if err == nil {
			t.Fatal("expected error for nil domain key")
		}
	})
}

// ---------- Double Vote Prevention ----------

func TestDoubleVotePrevention(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupDomainWithIssue(t, k, ctx)

	aliceKey := domainKey("alice-double-vote-key")
	k.JoinPermissionRegister(ctx, "AnonDomain", "alice", aliceKey.PubKey().Bytes())

	// First vote should succeed.
	_, _, err := k.RateProposal(ctx, "AnonDomain", "Climate", "GreenDeal", 3, aliceKey)
	if err != nil {
		t.Fatalf("first vote failed: %v", err)
	}

	// Second vote with same key should fail.
	_, _, err = k.RateProposal(ctx, "AnonDomain", "Climate", "GreenDeal", -2, aliceKey)
	if err == nil {
		t.Fatal("expected error for double-voting with same domain key")
	}
}

// ---------- Rating After Purge ----------

func TestRatingAfterPurge(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupDomainWithIssue(t, k, ctx)

	oldKey := domainKey("alice-old-key")
	k.JoinPermissionRegister(ctx, "AnonDomain", "alice", oldKey.PubKey().Bytes())

	// Vote with old key.
	_, _, err := k.RateProposal(ctx, "AnonDomain", "Climate", "GreenDeal", 4, oldKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Admin purges the permission register.
	err = k.PurgePermissionRegister(ctx, "AnonDomain", sdk.AccAddress("admin1"))
	if err != nil {
		t.Fatalf("purge failed: %v", err)
	}

	// Old key no longer works.
	_, _, err = k.RateProposal(ctx, "AnonDomain", "Climate", "GreenDeal", 2, oldKey)
	if err == nil {
		t.Fatal("expected error for purged key")
	}

	// Re-register with a fresh key.
	newKey := domainKey("alice-new-key")
	err = k.JoinPermissionRegister(ctx, "AnonDomain", "alice", newKey.PubKey().Bytes())
	if err != nil {
		t.Fatalf("re-registration failed: %v", err)
	}

	// New key can vote (it's a different key, so not a double vote).
	_, _, err = k.RateProposal(ctx, "AnonDomain", "Climate", "GreenDeal", 5, newKey)
	if err != nil {
		t.Fatalf("vote with new key failed: %v", err)
	}
}

// ---------- Voter Unlinkability ----------

func TestVoterUnlinkability(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupDomainWithIssue(t, k, ctx)

	// Alice and Bob each register with separate domain keys.
	aliceKey := domainKey("alice-unlink-key")
	bobKey := domainKey("bob-unlink-key")
	k.JoinPermissionRegister(ctx, "AnonDomain", "alice", aliceKey.PubKey().Bytes())
	k.JoinPermissionRegister(ctx, "AnonDomain", "bob", bobKey.PubKey().Bytes())

	// Both vote.
	k.RateProposal(ctx, "AnonDomain", "Climate", "GreenDeal", 3, aliceKey)
	// Add another suggestion for Bob to vote on the same one.
	domain, _ := k.GetDomain(ctx, "AnonDomain")
	domain.Issues[0].Suggestions = append(domain.Issues[0].Suggestions, Suggestion{
		Name: "CarbonTax", Creator: "bob", Ratings: []Rating{}, Stones: 0,
	})
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:AnonDomain"), bz)

	k.RateProposal(ctx, "AnonDomain", "Climate", "CarbonTax", -1, bobKey)

	// Verify: no avatar name ("alice", "bob") appears anywhere in stored ratings.
	domain, _ = k.GetDomain(ctx, "AnonDomain")
	for _, issue := range domain.Issues {
		for _, suggestion := range issue.Suggestions {
			for _, rating := range suggestion.Ratings {
				if rating.DomainPubKeyHex == "alice" || rating.DomainPubKeyHex == "bob" {
					t.Error("avatar name leaked into rating data")
				}
				// Verify it's a valid hex string (domain pubkey).
				if _, err := hex.DecodeString(rating.DomainPubKeyHex); err != nil {
					t.Errorf("rating key is not valid hex: %s", rating.DomainPubKeyHex)
				}
			}
		}
	}

	// Verify permission register contains only hex keys, not avatar names.
	for _, key := range domain.PermissionReg {
		if key == "alice" || key == "bob" {
			t.Error("avatar name found in permission register")
		}
	}
}

// ---------- Excluded Member Cannot Re-register ----------

func TestExcludedMemberCannotReregister(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupDomainWithIssue(t, k, ctx)

	charlieKey := domainKey("charlie-key")
	// Charlie is a member, so registration works.
	err := k.JoinPermissionRegister(ctx, "AnonDomain", "charlie", charlieKey.PubKey().Bytes())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Admin purges.
	k.PurgePermissionRegister(ctx, "AnonDomain", sdk.AccAddress("admin1"))

	// Remove charlie from members.
	domain, _ := k.GetDomain(ctx, "AnonDomain")
	var newMembers []string
	for _, m := range domain.Members {
		if m != "charlie" {
			newMembers = append(newMembers, m)
		}
	}
	domain.Members = newMembers
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:AnonDomain"), bz)

	// Charlie tries to re-register after being removed — should fail.
	newCharlieKey := domainKey("charlie-new-key")
	err = k.JoinPermissionRegister(ctx, "AnonDomain", "charlie", newCharlieKey.PubKey().Bytes())
	if err == nil {
		t.Fatal("expected error — charlie was removed from the domain and cannot re-register")
	}
}
