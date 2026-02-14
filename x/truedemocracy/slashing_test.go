package truedemocracy

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	rewards "truerepublic/treasury/keeper"
)

func TestHandleDoubleSign(t *testing.T) {
	k, ctx := setupKeeper(t)
	_, pk := setupDomainWithValidator(t, k, ctx)

	if err := k.HandleDoubleSign(ctx, pk); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, found := k.GetValidator(ctx, "oper1")
	if !found {
		t.Fatal("validator should still exist after slash")
	}
	if !val.Jailed {
		t.Error("validator should be jailed after double sign")
	}

	// 5% of 100_000 = 5_000 slashed → 95_000 remaining.
	want := math.NewInt(95_000)
	got := val.Stake.AmountOf("pnyx")
	if !got.Equal(want) {
		t.Errorf("stake after slash = %s, want %s", got, want)
	}
}

func TestHandleDoubleSignBelowMinimum(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateDomain(ctx, "D", sdk.AccAddress("a"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500_000)))
	domain, _ := k.GetDomain(ctx, "D")
	domain.Members = append(domain.Members, "lowval")
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:D"), bz)

	// Register with exactly minimum stake.
	pk := testPubKey("lowval")
	stake := sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000))
	k.RegisterValidator(ctx, "lowval", pk, stake, "D")

	k.HandleDoubleSign(ctx, pk)

	val, _ := k.GetValidator(ctx, "lowval")
	// After 5% slash: 95_000 < 100_000 → power should be 0.
	if val.Power != 0 {
		t.Errorf("power = %d, want 0 (stake below minimum after slash)", val.Power)
	}
}

func TestHandleDowntime(t *testing.T) {
	k, ctx := setupKeeper(t)
	_, pk := setupDomainWithValidator(t, k, ctx)

	// Miss blocks below threshold — no slash.
	threshold := SignedBlocksWindow - MinSignedPerWindow // 50
	for i := int64(0); i < threshold; i++ {
		if err := k.HandleDowntime(ctx, pk); err != nil {
			t.Fatal(err)
		}
	}
	val, _ := k.GetValidator(ctx, "oper1")
	if val.Jailed {
		t.Error("should not be jailed before exceeding threshold")
	}
	if val.MissedBlocks != threshold {
		t.Errorf("missed blocks = %d, want %d", val.MissedBlocks, threshold)
	}

	// One more miss exceeds threshold → slash and jail.
	if err := k.HandleDowntime(ctx, pk); err != nil {
		t.Fatal(err)
	}
	val, _ = k.GetValidator(ctx, "oper1")
	if !val.Jailed {
		t.Error("should be jailed after exceeding threshold")
	}
	// 1% of 100_000 = 1_000 slashed → 99_000 remaining.
	want := math.NewInt(99_000)
	got := val.Stake.AmountOf("pnyx")
	if !got.Equal(want) {
		t.Errorf("stake after downtime slash = %s, want %s", got, want)
	}
	if val.MissedBlocks != 0 {
		t.Errorf("missed blocks should be reset, got %d", val.MissedBlocks)
	}
}

func TestUnjail(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupDomainWithValidator(t, k, ctx)

	// Jail the validator.
	val, _ := k.GetValidator(ctx, "oper1")
	val.Jailed = true
	val.JailedUntil = ctx.BlockTime().Unix() + DowntimeJailDuration
	val.Power = 0
	k.SetValidator(ctx, val)

	t.Run("too early", func(t *testing.T) {
		err := k.Unjail(ctx, "oper1")
		if err == nil {
			t.Fatal("expected error — jail duration not elapsed")
		}
	})

	t.Run("after jail duration", func(t *testing.T) {
		futureCtx := ctx.WithBlockTime(ctx.BlockTime().Add(time.Duration(DowntimeJailDuration+1) * time.Second))
		err := k.Unjail(futureCtx, "oper1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		val, _ := k.GetValidator(futureCtx, "oper1")
		if val.Jailed {
			t.Error("validator should be unjailed")
		}
		if val.Power != 1 {
			t.Errorf("power = %d, want 1", val.Power)
		}
	})
}

func TestUnjailFailsStakeBelowMin(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupDomainWithValidator(t, k, ctx)

	// Manually set stake below minimum and jail.
	val, _ := k.GetValidator(ctx, "oper1")
	val.Jailed = true
	val.JailedUntil = ctx.BlockTime().Unix() - 1 // already expired
	val.Stake = sdk.NewCoins(sdk.NewInt64Coin("pnyx", rewards.StakeMin-1))
	val.Power = 0
	k.SetValidator(ctx, val)

	err := k.Unjail(ctx, "oper1")
	if err == nil {
		t.Fatal("expected error — stake below minimum")
	}
}

func TestUnjailFailsNotDomainMember(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupDomainWithValidator(t, k, ctx)

	// Jail the validator.
	val, _ := k.GetValidator(ctx, "oper1")
	val.Jailed = true
	val.JailedUntil = ctx.BlockTime().Unix() - 1
	k.SetValidator(ctx, val)

	// Remove oper1 from domain.
	domain, _ := k.GetDomain(ctx, "TestDomain")
	domain.Members = []string{domain.Admin.String()} // only admin
	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:TestDomain"), bz)

	err := k.Unjail(ctx, "oper1")
	if err == nil {
		t.Fatal("expected error — operator no longer a domain member")
	}
}
