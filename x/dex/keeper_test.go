package dex

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

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
	RegisterCodec(cdc)

	keeper := NewKeeper(cdc, storeKey)
	ctx := sdk.NewContext(ms, cmtproto.Header{}, false, log.NewNopLogger())

	return keeper, ctx
}

// ---------- CreatePool ----------

func TestCreatePool(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("happy path", func(t *testing.T) {
		err := k.CreatePool(ctx, "atom", math.NewInt(1_000_000), math.NewInt(500_000))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		pool, found := k.GetPool(ctx, "atom")
		if !found {
			t.Fatal("pool not found after creation")
		}
		if !pool.PnyxReserve.Equal(math.NewInt(1_000_000)) {
			t.Errorf("pnyx reserve = %s, want 1000000", pool.PnyxReserve)
		}
		if !pool.AssetReserve.Equal(math.NewInt(500_000)) {
			t.Errorf("asset reserve = %s, want 500000", pool.AssetReserve)
		}
		// sqrt(1_000_000 * 500_000) = sqrt(500_000_000_000) ≈ 707106
		if !pool.TotalShares.IsPositive() {
			t.Error("total shares should be positive")
		}
	})

	t.Run("duplicate pool", func(t *testing.T) {
		err := k.CreatePool(ctx, "atom", math.NewInt(100), math.NewInt(100))
		if err == nil {
			t.Fatal("expected error for duplicate pool")
		}
	})

	t.Run("zero amounts", func(t *testing.T) {
		err := k.CreatePool(ctx, "btc", math.ZeroInt(), math.NewInt(100))
		if err == nil {
			t.Fatal("expected error for zero pnyx amount")
		}
		err = k.CreatePool(ctx, "btc", math.NewInt(100), math.ZeroInt())
		if err == nil {
			t.Fatal("expected error for zero asset amount")
		}
	})
}

// ---------- Swap ----------

func TestSwap(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreatePool(ctx, "atom", math.NewInt(1_000_000), math.NewInt(1_000_000))

	t.Run("PNYX to ATOM", func(t *testing.T) {
		poolBefore, _ := k.GetPool(ctx, "atom")
		kBefore := poolBefore.PnyxReserve.Mul(poolBefore.AssetReserve)

		out, err := k.Swap(ctx, "pnyx", math.NewInt(10_000), "atom")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !out.IsPositive() {
			t.Fatal("output should be positive")
		}

		poolAfter, _ := k.GetPool(ctx, "atom")
		kAfter := poolAfter.PnyxReserve.Mul(poolAfter.AssetReserve)

		// k should increase (fees stay in pool).
		if kAfter.LT(kBefore) {
			t.Errorf("constant product decreased: %s < %s", kAfter, kBefore)
		}
	})

	t.Run("ATOM to PNYX", func(t *testing.T) {
		out, err := k.Swap(ctx, "atom", math.NewInt(5_000), "pnyx")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !out.IsPositive() {
			t.Fatal("output should be positive")
		}
	})

	t.Run("unknown denom", func(t *testing.T) {
		_, err := k.Swap(ctx, "pnyx", math.NewInt(100), "xxx")
		if err == nil {
			t.Fatal("expected error for unknown pool")
		}
	})

	t.Run("zero input", func(t *testing.T) {
		_, err := k.Swap(ctx, "pnyx", math.ZeroInt(), "atom")
		if err == nil {
			t.Fatal("expected error for zero input")
		}
	})

	t.Run("both pnyx", func(t *testing.T) {
		_, err := k.Swap(ctx, "pnyx", math.NewInt(100), "pnyx")
		if err == nil {
			t.Fatal("expected error when both denoms are pnyx")
		}
	})
}

func TestSwapFeeDeduction(t *testing.T) {
	k, ctx := setupKeeper(t)
	// Equal reserves for simpler math.
	k.CreatePool(ctx, "atom", math.NewInt(1_000_000), math.NewInt(1_000_000))

	// With 0.3% fee, swapping 10000 PNYX into a 1M/1M pool:
	// out = 1000000 * 10000 * 9970 / (1000000 * 10000 + 10000 * 9970)
	//     = 9970000000000 / (10000000000 + 99700000)
	//     = 9970000000000 / 10099700000
	//     ≈ 9871 (less than ~9901 without fee)
	out, err := k.Swap(ctx, "pnyx", math.NewInt(10_000), "atom")
	if err != nil {
		t.Fatal(err)
	}
	// Without fee: 10000 * 1000000 / (1000000 + 10000) = 9900.99 ≈ 9900
	// With 0.3% fee, output should be less.
	noFeeApprox := math.NewInt(9901)
	if out.GTE(noFeeApprox) {
		t.Errorf("output %s should be less than no-fee output %s", out, noFeeApprox)
	}
	if out.LT(math.NewInt(9800)) {
		t.Errorf("output %s is unexpectedly low", out)
	}
}

func TestSwapFeeAccumulation(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreatePool(ctx, "atom", math.NewInt(1_000_000), math.NewInt(1_000_000))

	poolBefore, _ := k.GetPool(ctx, "atom")
	kBefore := poolBefore.PnyxReserve.Mul(poolBefore.AssetReserve)

	// Perform many swaps in both directions.
	for i := 0; i < 20; i++ {
		k.Swap(ctx, "pnyx", math.NewInt(1_000), "atom")
		k.Swap(ctx, "atom", math.NewInt(1_000), "pnyx")
	}

	poolAfter, _ := k.GetPool(ctx, "atom")
	kAfter := poolAfter.PnyxReserve.Mul(poolAfter.AssetReserve)

	// After many swaps, k should grow due to accumulated fees.
	if !kAfter.GT(kBefore) {
		t.Errorf("constant product should grow from fees: before=%s, after=%s", kBefore, kAfter)
	}
}

// ---------- AddLiquidity ----------

func TestAddLiquidity(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreatePool(ctx, "atom", math.NewInt(1_000_000), math.NewInt(1_000_000))

	poolBefore, _ := k.GetPool(ctx, "atom")

	t.Run("proportional deposit", func(t *testing.T) {
		// Deposit 10% of both reserves.
		shares, err := k.AddLiquidity(ctx, "atom", math.NewInt(100_000), math.NewInt(100_000))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should get 10% of total shares.
		expectedShares := poolBefore.TotalShares.Quo(math.NewInt(10))
		if !shares.Equal(expectedShares) {
			t.Errorf("shares = %s, want %s", shares, expectedShares)
		}

		pool, _ := k.GetPool(ctx, "atom")
		if !pool.PnyxReserve.Equal(math.NewInt(1_100_000)) {
			t.Errorf("pnyx reserve = %s, want 1100000", pool.PnyxReserve)
		}
	})

	t.Run("imbalanced deposit uses smaller ratio", func(t *testing.T) {
		pool, _ := k.GetPool(ctx, "atom")
		totalBefore := pool.TotalShares

		// Deposit disproportionate amounts — shares based on smaller ratio.
		shares, err := k.AddLiquidity(ctx, "atom", math.NewInt(110_000), math.NewInt(55_000))
		if err != nil {
			t.Fatal(err)
		}

		// assetAmt/assetReserve = 55000/1100000 = 5%
		// pnyxAmt/pnyxReserve = 110000/1100000 = 10%
		// min = 5%, so shares = 5% of total.
		expectedShares := totalBefore.Quo(math.NewInt(20)) // 5%
		if !shares.Equal(expectedShares) {
			t.Errorf("shares = %s, want %s (5%% of %s)", shares, expectedShares, totalBefore)
		}
	})

	t.Run("no pool", func(t *testing.T) {
		_, err := k.AddLiquidity(ctx, "xxx", math.NewInt(100), math.NewInt(100))
		if err == nil {
			t.Fatal("expected error for unknown pool")
		}
	})
}

// ---------- RemoveLiquidity ----------

func TestRemoveLiquidity(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreatePool(ctx, "atom", math.NewInt(1_000_000), math.NewInt(500_000))

	pool, _ := k.GetPool(ctx, "atom")
	totalShares := pool.TotalShares

	t.Run("partial withdrawal", func(t *testing.T) {
		// Remove 10% of shares.
		sharesToRemove := totalShares.Quo(math.NewInt(10))
		pnyxOut, assetOut, err := k.RemoveLiquidity(ctx, "atom", sharesToRemove)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should get ~10% of reserves (integer rounding may lose 1).
		wantPnyx := pool.PnyxReserve.Mul(sharesToRemove).Quo(totalShares)
		wantAsset := pool.AssetReserve.Mul(sharesToRemove).Quo(totalShares)
		if !pnyxOut.Equal(wantPnyx) {
			t.Errorf("pnyxOut = %s, want %s", pnyxOut, wantPnyx)
		}
		if !assetOut.Equal(wantAsset) {
			t.Errorf("assetOut = %s, want %s", assetOut, wantAsset)
		}
		if !pnyxOut.IsPositive() || !assetOut.IsPositive() {
			t.Error("outputs should be positive")
		}
	})

	t.Run("full withdrawal", func(t *testing.T) {
		pool, _ := k.GetPool(ctx, "atom")
		pnyxOut, assetOut, err := k.RemoveLiquidity(ctx, "atom", pool.TotalShares)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !pnyxOut.Equal(pool.PnyxReserve) {
			t.Errorf("pnyxOut = %s, want %s", pnyxOut, pool.PnyxReserve)
		}
		if !assetOut.Equal(pool.AssetReserve) {
			t.Errorf("assetOut = %s, want %s", assetOut, pool.AssetReserve)
		}
	})

	t.Run("excessive shares", func(t *testing.T) {
		// Pool is empty now, recreate.
		k.CreatePool(ctx, "btc", math.NewInt(100), math.NewInt(100))
		pool, _ := k.GetPool(ctx, "btc")
		_, _, err := k.RemoveLiquidity(ctx, "btc", pool.TotalShares.Add(math.OneInt()))
		if err == nil {
			t.Fatal("expected error for excessive shares")
		}
	})
}

// ---------- intSqrt ----------

func TestIntSqrt(t *testing.T) {
	tests := []struct {
		input int64
		want  int64
	}{
		{0, 0},
		{1, 1},
		{4, 2},
		{9, 3},
		{10, 3}, // floor
		{100, 10},
		{1_000_000, 1_000},
		{1_000_000_000_000, 1_000_000}, // sqrt(1e12) = 1e6
	}
	for _, tt := range tests {
		got := intSqrt(math.NewInt(tt.input))
		if !got.Equal(math.NewInt(tt.want)) {
			t.Errorf("intSqrt(%d) = %s, want %d", tt.input, got, tt.want)
		}
	}
}
