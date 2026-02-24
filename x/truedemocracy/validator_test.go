package truedemocracy

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	rewards "truerepublic/treasury/keeper"
)

// setupKeeper creates an in-memory Keeper and sdk.Context for testing.
func setupKeeper(t *testing.T) (Keeper, sdk.Context) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(ModuleName)
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	ms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, nil)
	if err := ms.LoadLatestVersion(); err != nil {
		t.Fatal(err)
	}

	cdc := codec.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(cdc)
	RegisterCodec(cdc)

	nodes := BuildTree()
	keeper := NewKeeper(cdc, storeKey, nodes)

	ctx := sdk.NewContext(ms, cmtproto.Header{
		Time: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}, false, log.NewNopLogger())

	return keeper, ctx
}

// testPubKey returns a deterministic ed25519 public key for testing.
func testPubKey(seed string) []byte {
	return ed25519.GenPrivKeyFromSecret([]byte(seed)).PubKey().Bytes()
}

// setupDomainWithValidator creates a domain and registers a validator in it.
func setupDomainWithValidator(t *testing.T, k Keeper, ctx sdk.Context) (string, []byte) {
	t.Helper()
	k.CreateDomain(ctx, "TestDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500_000)))

	// Re-save domain with the validator operator as a member.
	domain, _ := k.GetDomain(ctx, "TestDomain")
	domain.Members = append(domain.Members, "oper1")
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:TestDomain"), bz)

	pk := testPubKey("oper1")
	stake := sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000))
	if err := k.RegisterValidator(ctx, "oper1", pk, stake, "TestDomain"); err != nil {
		t.Fatal(err)
	}
	return "oper1", pk
}

func TestRegisterValidator(t *testing.T) {
	k, ctx := setupKeeper(t)

	k.CreateDomain(ctx, "Party", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500_000)))
	// Add member to domain.
	domain, _ := k.GetDomain(ctx, "Party")
	domain.Members = append(domain.Members, "val1")
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:Party"), bz)

	pk := testPubKey("val1")

	t.Run("happy path", func(t *testing.T) {
		stake := sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000))
		err := k.RegisterValidator(ctx, "val1", pk, stake, "Party")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		val, found := k.GetValidator(ctx, "val1")
		if !found {
			t.Fatal("validator not found after registration")
		}
		if val.Power != 1 {
			t.Errorf("power = %d, want 1", val.Power)
		}
		if val.Jailed {
			t.Error("new validator should not be jailed")
		}
	})

	t.Run("duplicate registration", func(t *testing.T) {
		stake := sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000))
		err := k.RegisterValidator(ctx, "val1", pk, stake, "Party")
		if err == nil {
			t.Fatal("expected error for duplicate registration")
		}
	})

	t.Run("insufficient stake", func(t *testing.T) {
		pk2 := testPubKey("val2-low")
		domain, _ := k.GetDomain(ctx, "Party")
		domain.Members = append(domain.Members, "val2")
		bz := k.cdc.MustMarshalLengthPrefixed(&domain)
		st.Set([]byte("domain:Party"), bz)

		stake := sdk.NewCoins(sdk.NewInt64Coin("pnyx", 99_999))
		err := k.RegisterValidator(ctx, "val2", pk2, stake, "Party")
		if err == nil {
			t.Fatal("expected error for insufficient stake")
		}
	})

	t.Run("non-member", func(t *testing.T) {
		pk3 := testPubKey("outsider")
		stake := sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000))
		err := k.RegisterValidator(ctx, "outsider", pk3, stake, "Party")
		if err == nil {
			t.Fatal("expected error for non-member")
		}
	})

	t.Run("bad pubkey length", func(t *testing.T) {
		domain, _ := k.GetDomain(ctx, "Party")
		domain.Members = append(domain.Members, "val3")
		bz := k.cdc.MustMarshalLengthPrefixed(&domain)
		st.Set([]byte("domain:Party"), bz)

		stake := sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000))
		err := k.RegisterValidator(ctx, "val3", []byte("short"), stake, "Party")
		if err == nil {
			t.Fatal("expected error for bad pubkey length")
		}
	})
}

func TestGetValidatorByPubKey(t *testing.T) {
	k, ctx := setupKeeper(t)
	addr, pk := setupDomainWithValidator(t, k, ctx)

	val, found := k.GetValidatorByPubKey(ctx, pk)
	if !found {
		t.Fatal("validator not found by pubkey")
	}
	if val.OperatorAddr != addr {
		t.Errorf("operator = %s, want %s", val.OperatorAddr, addr)
	}
}

func TestRemoveValidator(t *testing.T) {
	k, ctx := setupKeeper(t)
	addr, pk := setupDomainWithValidator(t, k, ctx)

	err := k.RemoveValidator(ctx, addr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, found := k.GetValidator(ctx, addr)
	if found {
		t.Error("validator should be removed")
	}
	_, found = k.GetValidatorByPubKey(ctx, pk)
	if found {
		t.Error("pubkey reverse index should be removed")
	}
}

func TestIterateValidators(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupDomainWithValidator(t, k, ctx)

	// Register a second validator.
	domain, _ := k.GetDomain(ctx, "TestDomain")
	domain.Members = append(domain.Members, "oper2")
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:TestDomain"), bz)

	pk2 := testPubKey("oper2")
	stake := sdk.NewCoins(sdk.NewInt64Coin("pnyx", 200_000))
	if err := k.RegisterValidator(ctx, "oper2", pk2, stake, "TestDomain"); err != nil {
		t.Fatal(err)
	}

	count := 0
	k.IterateValidators(ctx, func(v Validator) bool {
		count++
		return false
	})
	if count != 2 {
		t.Errorf("iterated %d validators, want 2", count)
	}
}

func TestEnforceDomainMembership(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupDomainWithValidator(t, k, ctx)

	t.Run("member present", func(t *testing.T) {
		ok := k.EnforceDomainMembership(ctx, "oper1")
		if !ok {
			t.Error("expected true — operator is a domain member")
		}
	})

	t.Run("member removed", func(t *testing.T) {
		// Remove oper1 from domain members.
		domain, _ := k.GetDomain(ctx, "TestDomain")
		var newMembers []string
		for _, m := range domain.Members {
			if m != "oper1" {
				newMembers = append(newMembers, m)
			}
		}
		domain.Members = newMembers
		st := ctx.KVStore(k.StoreKey)
		bz := k.cdc.MustMarshalLengthPrefixed(&domain)
		st.Set([]byte("domain:TestDomain"), bz)

		ok := k.EnforceDomainMembership(ctx, "oper1")
		if ok {
			t.Error("expected false — operator was removed from domain")
		}
	})
}

func TestDistributeStakingRewards(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupDomainWithValidator(t, k, ctx)

	// Initialize reward tracking.
	st := ctx.KVStore(k.StoreKey)
	initTime := ctx.BlockTime().Unix()
	st.Set([]byte("pod:last-reward-time"), k.cdc.MustMarshalLengthPrefixed(initTime))
	zeroInt := math.ZeroInt()
	st.Set([]byte("pod:total-release"), k.cdc.MustMarshalLengthPrefixed(&zeroInt))

	t.Run("no reward before interval", func(t *testing.T) {
		// Advance by less than RewardInterval.
		ctx2 := ctx.WithBlockTime(ctx.BlockTime().Add(30 * time.Minute))
		if err := k.DistributeStakingRewards(ctx2); err != nil {
			t.Fatal(err)
		}
		val, _ := k.GetValidator(ctx2, "oper1")
		if !val.Stake.AmountOf("pnyx").Equal(math.NewInt(100_000)) {
			t.Errorf("stake changed before interval: %s", val.Stake)
		}
	})

	t.Run("reward after interval", func(t *testing.T) {
		// Advance by exactly RewardInterval (1 hour).
		ctx2 := ctx.WithBlockTime(ctx.BlockTime().Add(time.Duration(RewardInterval) * time.Second))
		if err := k.DistributeStakingRewards(ctx2); err != nil {
			t.Fatal(err)
		}
		val, _ := k.GetValidator(ctx2, "oper1")
		stakeAmt := val.Stake.AmountOf("pnyx")

		// Expected: CalcNodeReward(100000, 0, 3600)
		expected := rewards.CalcNodeReward(math.NewInt(100_000), math.ZeroInt(), RewardInterval)
		wantStake := math.NewInt(100_000).Add(expected)
		if !stakeAmt.Equal(wantStake) {
			t.Errorf("stake = %s, want %s (reward = %s)", stakeAmt, wantStake, expected)
		}
	})
}

func TestBuildValidatorUpdates(t *testing.T) {
	k, ctx := setupKeeper(t)
	_, pk := setupDomainWithValidator(t, k, ctx)

	updates := k.BuildValidatorUpdates(ctx)
	if len(updates) != 1 {
		t.Fatalf("got %d updates, want 1", len(updates))
	}
	if updates[0].Power != 1 {
		t.Errorf("power = %d, want 1", updates[0].Power)
	}
	gotPK := updates[0].PubKey.GetEd25519()
	if len(gotPK) == 0 {
		t.Fatal("pubkey is empty in update")
	}
	for i := range pk {
		if gotPK[i] != pk[i] {
			t.Fatalf("pubkey mismatch at byte %d", i)
		}
	}

	t.Run("jailed validator has power 0", func(t *testing.T) {
		val, _ := k.GetValidator(ctx, "oper1")
		val.Jailed = true
		k.SetValidator(ctx, val)

		updates := k.BuildValidatorUpdates(ctx)
		if len(updates) != 1 {
			t.Fatalf("got %d updates, want 1", len(updates))
		}
		if updates[0].Power != 0 {
			t.Errorf("jailed validator power = %d, want 0", updates[0].Power)
		}
	})
}

// ---------- PoD Transfer Limit (WP §7) ----------

// helper: register a validator with given stake in a domain.
func registerVal(t *testing.T, k Keeper, ctx sdk.Context, domainName, operAddr, seed string, stakeAmt int64) {
	t.Helper()
	domain, _ := k.GetDomain(ctx, domainName)
	domain.Members = append(domain.Members, operAddr)
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:"+domainName), bz)

	pk := testPubKey(seed)
	stake := sdk.NewCoins(sdk.NewInt64Coin("pnyx", stakeAmt))
	if err := k.RegisterValidator(ctx, operAddr, pk, stake, domainName); err != nil {
		t.Fatal(err)
	}
}

func TestTransferLimitEnforcement(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateDomain(ctx, "LimitDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 5_000_000)))

	registerVal(t, k, ctx, "LimitDomain", "val1", "val1-lim", 200_000)

	// Set domain payouts to 1,000,000 → 10% limit = 100,000.
	domain, _ := k.GetDomain(ctx, "LimitDomain")
	domain.TotalPayouts = 1_000_000
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:LimitDomain"), bz)

	// Try to withdraw 150,000 → exceeds 10% limit (100,000). Should fail.
	err := k.WithdrawStake(ctx, "val1", 150_000)
	if err == nil {
		t.Fatal("expected error — withdraw exceeds 10% of domain payouts")
	}
}

func TestTransferWithinLimit(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateDomain(ctx, "OkDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 5_000_000)))

	registerVal(t, k, ctx, "OkDomain", "val1", "val1-ok", 200_000)

	// Set domain payouts to 2,000,000 → 10% limit = 200,000.
	domain, _ := k.GetDomain(ctx, "OkDomain")
	domain.TotalPayouts = 2_000_000
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:OkDomain"), bz)

	// Withdraw 100,000 → within limit. Should succeed.
	err := k.WithdrawStake(ctx, "val1", 100_000)
	if err != nil {
		t.Fatalf("withdrawal within limit should succeed: %v", err)
	}

	// Verify TransferredStake was incremented.
	domain, _ = k.GetDomain(ctx, "OkDomain")
	if domain.TransferredStake != 100_000 {
		t.Errorf("TransferredStake = %d, want 100000", domain.TransferredStake)
	}

	// Verify validator stake was reduced.
	val, found := k.GetValidator(ctx, "val1")
	if !found {
		t.Fatal("validator should still exist")
	}
	if val.Stake.AmountOf("pnyx").Int64() != 100_000 {
		t.Errorf("validator stake = %d, want 100000", val.Stake.AmountOf("pnyx").Int64())
	}
}

func TestPayoutIncreasesLimit(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateDomain(ctx, "GrowDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 5_000_000)))

	registerVal(t, k, ctx, "GrowDomain", "val1", "val1-grow", 300_000)

	// Initial payouts = 1,000,000 → limit = 100,000.
	domain, _ := k.GetDomain(ctx, "GrowDomain")
	domain.TotalPayouts = 1_000_000
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:GrowDomain"), bz)

	// Withdraw 100,000 → exactly at limit, should succeed.
	err := k.WithdrawStake(ctx, "val1", 100_000)
	if err != nil {
		t.Fatalf("withdraw at limit should succeed: %v", err)
	}

	// Now another 1,000 should fail (limit exhausted).
	err = k.WithdrawStake(ctx, "val1", 1_000)
	if err == nil {
		t.Fatal("expected error — limit exhausted")
	}

	// Increase payouts → limit increases.
	domain, _ = k.GetDomain(ctx, "GrowDomain")
	domain.TotalPayouts = 3_000_000 // new limit = 300,000; already transferred 100,000 → 200,000 left
	bz = k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:GrowDomain"), bz)

	// Now withdraw 100,000 more — should succeed (200,000 remaining capacity).
	err = k.WithdrawStake(ctx, "val1", 100_000)
	if err != nil {
		t.Fatalf("after payout increase, withdraw should succeed: %v", err)
	}
}

func TestMultipleValidators(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateDomain(ctx, "MultiDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 5_000_000)))

	registerVal(t, k, ctx, "MultiDomain", "val1", "val1-multi", 200_000)
	registerVal(t, k, ctx, "MultiDomain", "val2", "val2-multi", 200_000)

	// Set payouts = 2,000,000 → 10% limit = 200,000 total across all validators.
	domain, _ := k.GetDomain(ctx, "MultiDomain")
	domain.TotalPayouts = 2_000_000
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:MultiDomain"), bz)

	// val1 withdraws 100,000 → success (100,000 of 200,000 used).
	err := k.WithdrawStake(ctx, "val1", 100_000)
	if err != nil {
		t.Fatalf("val1 withdraw should succeed: %v", err)
	}

	// val2 withdraws 100,000 → success (200,000 of 200,000 used).
	err = k.WithdrawStake(ctx, "val2", 100_000)
	if err != nil {
		t.Fatalf("val2 withdraw should succeed: %v", err)
	}

	// val1 tries another 1,000 → fail (limit exhausted).
	err = k.WithdrawStake(ctx, "val1", 1_000)
	if err == nil {
		t.Fatal("expected error — domain transfer limit exhausted")
	}
}

func TestWithdrawBelowMinRemovesValidator(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateDomain(ctx, "RemDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 5_000_000)))

	registerVal(t, k, ctx, "RemDomain", "val1", "val1-rem", 150_000)

	// Set payouts high enough that limit is not an issue.
	domain, _ := k.GetDomain(ctx, "RemDomain")
	domain.TotalPayouts = 10_000_000 // limit = 1,000,000
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:RemDomain"), bz)

	// Withdraw 60,000 → remaining 90,000 < StakeMin (100,000) → validator removed.
	err := k.WithdrawStake(ctx, "val1", 60_000)
	if err != nil {
		t.Fatalf("withdraw should succeed: %v", err)
	}
	_, found := k.GetValidator(ctx, "val1")
	if found {
		t.Error("validator should be removed when stake drops below minimum")
	}
}

func TestWithdrawNoPayoutsBlocked(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateDomain(ctx, "NoPay", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 5_000_000)))

	registerVal(t, k, ctx, "NoPay", "val1", "val1-nopay", 100_000)

	// TotalPayouts = 0 → no transfers allowed.
	err := k.WithdrawStake(ctx, "val1", 50_000)
	if err == nil {
		t.Fatal("expected error — no payouts, transfers blocked")
	}
}

func TestPayoutTrackingFromStoneReward(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupDomainWithIssues(t, k, ctx)

	// Place a stone → triggers VoteToEarn reward → should update TotalPayouts.
	_, err := k.PlaceStoneOnIssue(ctx, "StonesDomain", "Climate", "alice")
	if err != nil {
		t.Fatal(err)
	}

	domain, _ := k.GetDomain(ctx, "StonesDomain")
	if domain.TotalPayouts <= 0 {
		t.Error("TotalPayouts should be positive after stone reward")
	}
}

func TestPowerScalesWithStake(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateDomain(ctx, "Big", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500_000)))
	domain, _ := k.GetDomain(ctx, "Big")
	domain.Members = append(domain.Members, "whale")
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:Big"), bz)

	pk := testPubKey("whale")
	stake := sdk.NewCoins(sdk.NewInt64Coin("pnyx", 300_000))
	if err := k.RegisterValidator(ctx, "whale", pk, stake, "Big"); err != nil {
		t.Fatal(err)
	}
	val, _ := k.GetValidator(ctx, "whale")
	if val.Power != 3 {
		t.Errorf("power = %d, want 3 for 300k stake", val.Power)
	}
}

// ---------- Domain Interest (eq.4) ----------

func TestDistributeDomainInterest(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateDomain(ctx, "Active", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500_000)))

	// Set up domain with payouts so it qualifies for interest.
	domain, _ := k.GetDomain(ctx, "Active")
	domain.TotalPayouts = 100_000 // active domain
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:Active"), bz)

	// Initialize tracking state (mimics InitGenesis).
	initTime := ctx.BlockTime().Unix()
	st.Set([]byte("dom:last-interest-time"), k.cdc.MustMarshalLengthPrefixed(initTime))
	zeroInt := math.ZeroInt()
	st.Set([]byte("pod:total-release"), k.cdc.MustMarshalLengthPrefixed(&zeroInt))

	t.Run("no interest before interval", func(t *testing.T) {
		ctx2 := ctx.WithBlockTime(ctx.BlockTime().Add(30 * time.Minute))
		if err := k.DistributeDomainInterest(ctx2); err != nil {
			t.Fatal(err)
		}
		domain, _ := k.GetDomain(ctx2, "Active")
		if !domain.Treasury.AmountOf("pnyx").Equal(math.NewInt(500_000)) {
			t.Errorf("treasury changed before interval: %s", domain.Treasury)
		}
	})

	t.Run("interest after interval", func(t *testing.T) {
		ctx2 := ctx.WithBlockTime(ctx.BlockTime().Add(time.Duration(RewardInterval) * time.Second))
		if err := k.DistributeDomainInterest(ctx2); err != nil {
			t.Fatal(err)
		}
		domain, _ := k.GetDomain(ctx2, "Active")
		treasuryAmt := domain.Treasury.AmountOf("pnyx")

		// Expected: CalcDomainInterest(500000, 100000, 0, 3600)
		expected := rewards.CalcDomainInterest(
			math.NewInt(500_000), math.NewInt(100_000), math.ZeroInt(), RewardInterval,
		)
		wantTreasury := math.NewInt(500_000).Add(expected)
		if !treasuryAmt.Equal(wantTreasury) {
			t.Errorf("treasury = %s, want %s (interest = %s)", treasuryAmt, wantTreasury, expected)
		}
	})
}

func TestDomainInterestZeroPayouts(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateDomain(ctx, "Idle", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500_000)))

	// Domain with zero payouts — should earn no interest.
	st := ctx.KVStore(k.StoreKey)
	initTime := ctx.BlockTime().Unix()
	st.Set([]byte("dom:last-interest-time"), k.cdc.MustMarshalLengthPrefixed(initTime))
	zeroInt := math.ZeroInt()
	st.Set([]byte("pod:total-release"), k.cdc.MustMarshalLengthPrefixed(&zeroInt))

	ctx2 := ctx.WithBlockTime(ctx.BlockTime().Add(time.Duration(RewardInterval) * time.Second))
	if err := k.DistributeDomainInterest(ctx2); err != nil {
		t.Fatal(err)
	}
	domain, _ := k.GetDomain(ctx2, "Idle")
	if !domain.Treasury.AmountOf("pnyx").Equal(math.NewInt(500_000)) {
		t.Errorf("idle domain treasury should not change: got %s", domain.Treasury)
	}
}

func TestDomainInterestCappedByPayout(t *testing.T) {
	k, ctx := setupKeeper(t)
	// Large treasury, small payouts → interest capped by payout.
	k.CreateDomain(ctx, "Big", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 10_000_000)))

	domain, _ := k.GetDomain(ctx, "Big")
	domain.TotalPayouts = 1 // tiny payout → caps interest at 1
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:Big"), bz)

	initTime := ctx.BlockTime().Unix()
	st.Set([]byte("dom:last-interest-time"), k.cdc.MustMarshalLengthPrefixed(initTime))
	zeroInt := math.ZeroInt()
	st.Set([]byte("pod:total-release"), k.cdc.MustMarshalLengthPrefixed(&zeroInt))

	ctx2 := ctx.WithBlockTime(ctx.BlockTime().Add(time.Duration(RewardInterval) * time.Second))
	if err := k.DistributeDomainInterest(ctx2); err != nil {
		t.Fatal(err)
	}
	domain, _ = k.GetDomain(ctx2, "Big")
	interest := domain.Treasury.AmountOf("pnyx").Sub(math.NewInt(10_000_000))

	// The raw interest on 10M at 25% for 1 hour is large, but it must be capped at payout=1.
	if interest.GT(math.NewInt(1)) {
		t.Errorf("interest %s exceeds payout cap of 1", interest)
	}
}

func TestDomainInterestDecaysWithRelease(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateDomain(ctx, "Decay", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500_000)))

	domain, _ := k.GetDomain(ctx, "Decay")
	domain.TotalPayouts = 1_000_000 // high cap so it doesn't constrain
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:Decay"), bz)

	initTime := ctx.BlockTime().Unix()
	st.Set([]byte("dom:last-interest-time"), k.cdc.MustMarshalLengthPrefixed(initTime))

	// Set total release to 50% of supply → decay = 0.5.
	halfSupply := math.NewInt(rewards.SupplyMax / 2)
	st.Set([]byte("pod:total-release"), k.cdc.MustMarshalLengthPrefixed(&halfSupply))

	ctx2 := ctx.WithBlockTime(ctx.BlockTime().Add(time.Duration(RewardInterval) * time.Second))
	if err := k.DistributeDomainInterest(ctx2); err != nil {
		t.Fatal(err)
	}
	domain, _ = k.GetDomain(ctx2, "Decay")
	interestWithDecay := domain.Treasury.AmountOf("pnyx").Sub(math.NewInt(500_000))

	// Compare with zero-release interest.
	interestNoDecay := rewards.CalcDomainInterest(
		math.NewInt(500_000), math.NewInt(1_000_000), math.ZeroInt(), RewardInterval,
	)

	if !interestWithDecay.LT(interestNoDecay) {
		t.Errorf("interest with 50%% release (%s) should be less than with 0%% (%s)",
			interestWithDecay, interestNoDecay)
	}
}

func TestDomainInterestFirstCallInitializes(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateDomain(ctx, "Fresh", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500_000)))

	// Don't set dom:last-interest-time — first call should initialize it.
	st := ctx.KVStore(k.StoreKey)
	zeroInt := math.ZeroInt()
	st.Set([]byte("pod:total-release"), k.cdc.MustMarshalLengthPrefixed(&zeroInt))

	if err := k.DistributeDomainInterest(ctx); err != nil {
		t.Fatal(err)
	}

	// Should have initialized the timer without changing treasury.
	domain, _ := k.GetDomain(ctx, "Fresh")
	if !domain.Treasury.AmountOf("pnyx").Equal(math.NewInt(500_000)) {
		t.Errorf("first call should not pay interest: %s", domain.Treasury)
	}

	// Timer should now be set.
	if bz := st.Get([]byte("dom:last-interest-time")); bz == nil {
		t.Error("dom:last-interest-time should be initialized after first call")
	}
}
