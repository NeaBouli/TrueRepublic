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
