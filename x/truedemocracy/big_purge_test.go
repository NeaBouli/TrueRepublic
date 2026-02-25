package truedemocracy

import (
	"encoding/hex"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ---------- Big Purge EndBlock Execution ----------

func TestBigPurgeNotTriggeredBeforeTime(t *testing.T) {
	k, ctx := setupKeeper(t)

	k.CreateDomain(ctx, "PurgeDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))

	// Add member and register a domain key.
	domain, _ := k.GetDomain(ctx, "PurgeDomain")
	domain.Members = append(domain.Members, "alice")
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:PurgeDomain"), bz)

	aliceKey := domainKey("alice-purge-test")
	if err := k.JoinPermissionRegister(ctx, "PurgeDomain", "alice", aliceKey.PubKey().Bytes()); err != nil {
		t.Fatalf("failed to register key: %v", err)
	}

	// Run purge check — should NOT purge (90 days haven't passed).
	k.CheckAndExecuteBigPurges(ctx)

	pubHex := hex.EncodeToString(aliceKey.PubKey().Bytes())
	if !k.IsKeyAuthorized(ctx, "PurgeDomain", pubHex) {
		t.Fatal("key should still be authorized — purge time not reached")
	}
}

func TestBigPurgeTriggeredAtScheduledTime(t *testing.T) {
	k, ctx := setupKeeper(t)

	k.CreateDomain(ctx, "PurgeDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))

	// Add members and register keys.
	domain, _ := k.GetDomain(ctx, "PurgeDomain")
	domain.Members = append(domain.Members, "alice", "bob")
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:PurgeDomain"), bz)

	aliceKey := domainKey("alice-purge-trigger")
	bobKey := domainKey("bob-purge-trigger")
	k.JoinPermissionRegister(ctx, "PurgeDomain", "alice", aliceKey.PubKey().Bytes())
	k.JoinPermissionRegister(ctx, "PurgeDomain", "bob", bobKey.PubKey().Bytes())

	// Advance time past the purge schedule (90+ days).
	schedule, _ := k.GetBigPurgeSchedule(ctx, "PurgeDomain")
	futureTime := time.Unix(schedule.NextPurgeTime+1, 0)
	ctx = ctx.WithBlockTime(futureTime)

	k.CheckAndExecuteBigPurges(ctx)

	// Keys should be gone.
	aliceHex := hex.EncodeToString(aliceKey.PubKey().Bytes())
	bobHex := hex.EncodeToString(bobKey.PubKey().Bytes())
	if k.IsKeyAuthorized(ctx, "PurgeDomain", aliceHex) {
		t.Error("alice's key should be purged")
	}
	if k.IsKeyAuthorized(ctx, "PurgeDomain", bobHex) {
		t.Error("bob's key should be purged")
	}
}

func TestBigPurgeMembersPreserved(t *testing.T) {
	k, ctx := setupKeeper(t)

	k.CreateDomain(ctx, "PurgeDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))

	domain, _ := k.GetDomain(ctx, "PurgeDomain")
	domain.Members = append(domain.Members, "alice", "bob")
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:PurgeDomain"), bz)

	aliceKey := domainKey("alice-member-preserve")
	k.JoinPermissionRegister(ctx, "PurgeDomain", "alice", aliceKey.PubKey().Bytes())

	// Advance past purge time.
	schedule, _ := k.GetBigPurgeSchedule(ctx, "PurgeDomain")
	ctx = ctx.WithBlockTime(time.Unix(schedule.NextPurgeTime+1, 0))

	k.CheckAndExecuteBigPurges(ctx)

	// Permission register should be empty.
	domain, _ = k.GetDomain(ctx, "PurgeDomain")
	if len(domain.PermissionReg) != 0 {
		t.Errorf("PermissionReg length = %d, want 0", len(domain.PermissionReg))
	}

	// Members list must be intact.
	if len(domain.Members) != 3 { // admin1 + alice + bob
		t.Errorf("Members count = %d, want 3", len(domain.Members))
	}
}

func TestBigPurgeReschedules(t *testing.T) {
	k, ctx := setupKeeper(t)

	k.CreateDomain(ctx, "PurgeDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))

	schedule, _ := k.GetBigPurgeSchedule(ctx, "PurgeDomain")
	originalInterval := schedule.PurgeInterval

	// Advance past purge time.
	purgeAt := time.Unix(schedule.NextPurgeTime+100, 0)
	ctx = ctx.WithBlockTime(purgeAt)

	k.CheckAndExecuteBigPurges(ctx)

	// Verify rescheduled.
	updated, exists := k.GetBigPurgeSchedule(ctx, "PurgeDomain")
	if !exists {
		t.Fatal("schedule should still exist after purge")
	}
	expectedNext := purgeAt.Unix() + originalInterval
	if updated.NextPurgeTime != expectedNext {
		t.Errorf("NextPurgeTime = %d, want %d (current + interval)", updated.NextPurgeTime, expectedNext)
	}
}

func TestBigPurgeAnnouncementEmitted(t *testing.T) {
	k, ctx := setupKeeper(t)

	k.CreateDomain(ctx, "PurgeDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))

	schedule, _ := k.GetBigPurgeSchedule(ctx, "PurgeDomain")

	// Advance to announcement window (within 7 days of purge, but before purge).
	announcementTime := schedule.NextPurgeTime - schedule.AnnouncementLead + 1
	ctx = ctx.WithBlockTime(time.Unix(announcementTime, 0))
	ctx = ctx.WithEventManager(sdk.NewEventManager())

	k.CheckAndExecuteBigPurges(ctx)

	// Check for announcement event.
	found := false
	for _, event := range ctx.EventManager().Events() {
		if event.Type == "big_purge_announcement" {
			found = true
			for _, attr := range event.Attributes {
				if string(attr.Key) == "domain" && string(attr.Value) != "PurgeDomain" {
					t.Errorf("domain attribute = %s, want PurgeDomain", string(attr.Value))
				}
			}
		}
	}
	if !found {
		t.Fatal("expected big_purge_announcement event")
	}
}

func TestBigPurgeAnnouncementNotDuplicated(t *testing.T) {
	k, ctx := setupKeeper(t)

	k.CreateDomain(ctx, "PurgeDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))

	schedule, _ := k.GetBigPurgeSchedule(ctx, "PurgeDomain")

	// Advance to announcement window.
	announcementTime := schedule.NextPurgeTime - schedule.AnnouncementLead + 1
	ctx = ctx.WithBlockTime(time.Unix(announcementTime, 0))

	// First call — should announce.
	ctx = ctx.WithEventManager(sdk.NewEventManager())
	k.CheckAndExecuteBigPurges(ctx)

	count1 := 0
	for _, event := range ctx.EventManager().Events() {
		if event.Type == "big_purge_announcement" {
			count1++
		}
	}
	if count1 != 1 {
		t.Errorf("first call: announcement count = %d, want 1", count1)
	}

	// Second call — should NOT announce again.
	ctx = ctx.WithEventManager(sdk.NewEventManager())
	k.CheckAndExecuteBigPurges(ctx)

	count2 := 0
	for _, event := range ctx.EventManager().Events() {
		if event.Type == "big_purge_announcement" {
			count2++
		}
	}
	if count2 != 0 {
		t.Errorf("second call: announcement count = %d, want 0", count2)
	}
}

func TestBigPurgeExecutionEvent(t *testing.T) {
	k, ctx := setupKeeper(t)

	k.CreateDomain(ctx, "PurgeDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))

	schedule, _ := k.GetBigPurgeSchedule(ctx, "PurgeDomain")
	ctx = ctx.WithBlockTime(time.Unix(schedule.NextPurgeTime+1, 0))
	ctx = ctx.WithEventManager(sdk.NewEventManager())

	k.CheckAndExecuteBigPurges(ctx)

	found := false
	for _, event := range ctx.EventManager().Events() {
		if event.Type == "big_purge_executed" {
			found = true
			for _, attr := range event.Attributes {
				if string(attr.Key) == "domain" && string(attr.Value) != "PurgeDomain" {
					t.Errorf("domain attribute = %s, want PurgeDomain", string(attr.Value))
				}
			}
		}
	}
	if !found {
		t.Fatal("expected big_purge_executed event")
	}
}

func TestMultipleDomainsPurgeIndependently(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Create domain 1.
	k.CreateDomain(ctx, "Domain1", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))
	domain1, _ := k.GetDomain(ctx, "Domain1")
	domain1.Members = append(domain1.Members, "alice")
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain1)
	st.Set([]byte("domain:Domain1"), bz)

	aliceKey := domainKey("alice-multi-purge")
	k.JoinPermissionRegister(ctx, "Domain1", "alice", aliceKey.PubKey().Bytes())

	// Create domain 2 one day later (different purge schedule).
	ctx2 := ctx.WithBlockTime(ctx.BlockTime().Add(24 * time.Hour))
	k.CreateDomain(ctx2, "Domain2", sdk.AccAddress("admin2"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))
	domain2, _ := k.GetDomain(ctx2, "Domain2")
	domain2.Members = append(domain2.Members, "bob")
	bz = k.cdc.MustMarshalLengthPrefixed(&domain2)
	st.Set([]byte("domain:Domain2"), bz)

	bobKey := domainKey("bob-multi-purge")
	k.JoinPermissionRegister(ctx2, "Domain2", "bob", bobKey.PubKey().Bytes())

	// Advance to domain 1's purge time (domain 2 should not be purged yet).
	schedule1, _ := k.GetBigPurgeSchedule(ctx, "Domain1")
	ctx3 := ctx.WithBlockTime(time.Unix(schedule1.NextPurgeTime+1, 0))

	k.CheckAndExecuteBigPurges(ctx3)

	// Domain 1 should be purged.
	aliceHex := hex.EncodeToString(aliceKey.PubKey().Bytes())
	if k.IsKeyAuthorized(ctx3, "Domain1", aliceHex) {
		t.Error("Domain1: alice's key should be purged")
	}

	// Domain 2 should NOT be purged yet (created 1 day later).
	bobHex := hex.EncodeToString(bobKey.PubKey().Bytes())
	if !k.IsKeyAuthorized(ctx3, "Domain2", bobHex) {
		t.Error("Domain2: bob's key should NOT be purged yet")
	}

	// Advance to domain 2's purge time.
	schedule2, _ := k.GetBigPurgeSchedule(ctx3, "Domain2")
	ctx4 := ctx.WithBlockTime(time.Unix(schedule2.NextPurgeTime+1, 0))

	k.CheckAndExecuteBigPurges(ctx4)

	// Now domain 2 should also be purged.
	if k.IsKeyAuthorized(ctx4, "Domain2", bobHex) {
		t.Error("Domain2: bob's key should be purged now")
	}
}
