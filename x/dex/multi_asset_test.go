package dex

import (
	"encoding/json"
	"testing"

	"cosmossdk.io/math"
)

// ---------- Pool creation validation ----------

func TestCreatePoolUnregisteredAssetFails(t *testing.T) {
	k, ctx := setupKeeper(t)

	err := k.CreatePool(ctx, "ibc/UNREGISTERED", math.NewInt(1000), math.NewInt(1000))
	if err == nil {
		t.Fatal("expected error for unregistered asset")
	}
	if !containsStr(err.Error(), "not registered") {
		t.Errorf("error %q should mention 'not registered'", err.Error())
	}
}

func TestCreatePoolTradingDisabledFails(t *testing.T) {
	k, ctx := setupKeeper(t)

	k.RegisterAsset(ctx, RegisteredAsset{
		IBCDenom:       "ibc/BTC",
		Symbol:         "BTC",
		Decimals:       8,
		TradingEnabled: false, // disabled
	})

	err := k.CreatePool(ctx, "ibc/BTC", math.NewInt(1000), math.NewInt(1000))
	if err == nil {
		t.Fatal("expected error for trading-disabled asset")
	}
	if !containsStr(err.Error(), "trading disabled") {
		t.Errorf("error %q should mention 'trading disabled'", err.Error())
	}
}

func TestCreatePoolNativePnyxSkipsValidation(t *testing.T) {
	k, ctx := setupKeeper(t)

	// PNYX is native — no registration needed. But CreatePool always pairs
	// PNYX with an asset, so pnyx can't be the assetDenom. This test
	// confirms that "pnyx" denom validation itself doesn't fail.
	err := k.validateAssetForTrading(ctx, "pnyx")
	if err != nil {
		t.Fatalf("pnyx should skip validation, got: %v", err)
	}
}

// ---------- Swap validation ----------

func TestSwapTradingDisabledFails(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Register and enable, create pool, then disable.
	k.RegisterAsset(ctx, RegisteredAsset{IBCDenom: "ibc/ETH", Symbol: "ETH", Decimals: 18, TradingEnabled: true})
	k.CreatePool(ctx, "ibc/ETH", math.NewInt(1_000_000), math.NewInt(1_000_000))

	// Disable trading.
	k.UpdateAssetTradingStatus(ctx, "ibc/ETH", false)

	// Swap should fail.
	_, err := k.Swap(ctx, "pnyx", math.NewInt(1000), "ibc/ETH")
	if err == nil {
		t.Fatal("expected error for swap on disabled asset")
	}
	if !containsStr(err.Error(), "trading disabled") {
		t.Errorf("error %q should mention 'trading disabled'", err.Error())
	}

	// Reverse direction should also fail.
	_, err = k.Swap(ctx, "ibc/ETH", math.NewInt(1000), "pnyx")
	if err == nil {
		t.Fatal("expected error for reverse swap on disabled asset")
	}
}

func TestReEnableTradingAllowsSwap(t *testing.T) {
	k, ctx := setupKeeper(t)

	k.RegisterAsset(ctx, RegisteredAsset{IBCDenom: "ibc/LUSD", Symbol: "LUSD", Decimals: 18, TradingEnabled: true})
	k.CreatePool(ctx, "ibc/LUSD", math.NewInt(1_000_000), math.NewInt(1_000_000))

	// Disable then re-enable.
	k.UpdateAssetTradingStatus(ctx, "ibc/LUSD", false)
	k.UpdateAssetTradingStatus(ctx, "ibc/LUSD", true)

	// Swap should work again.
	out, err := k.Swap(ctx, "pnyx", math.NewInt(1000), "ibc/LUSD")
	if err != nil {
		t.Fatalf("swap should succeed after re-enabling: %v", err)
	}
	if !out.IsPositive() {
		t.Fatal("output should be positive")
	}
}

// ---------- Symbol resolution ----------

func TestGetSymbolForDenom(t *testing.T) {
	k, ctx := setupKeeper(t)

	k.RegisterAsset(ctx, RegisteredAsset{
		IBCDenom: "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
		Symbol:   "BTC",
		Decimals: 8,
	})

	tests := []struct {
		denom string
		want  string
	}{
		{"pnyx", "PNYX"},
		{"ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2", "BTC"},
		{"unknown_denom", "unknown_denom"}, // fallback to raw denom
	}

	for _, tt := range tests {
		got := k.GetSymbolForDenom(ctx, tt.denom)
		if got != tt.want {
			t.Errorf("GetSymbolForDenom(%q) = %q, want %q", tt.denom, got, tt.want)
		}
	}
}

// ---------- Multi-asset pool creation ----------

func TestMultiAssetPoolCreation(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Register BTC, ETH, LUSD.
	assets := []RegisteredAsset{
		{IBCDenom: "ibc/BTC", Symbol: "BTC", Name: "Bitcoin", Decimals: 8, TradingEnabled: true},
		{IBCDenom: "ibc/ETH", Symbol: "ETH", Name: "Ethereum", Decimals: 18, TradingEnabled: true},
		{IBCDenom: "ibc/LUSD", Symbol: "LUSD", Name: "Liquity USD", Decimals: 18, TradingEnabled: true},
	}
	for _, a := range assets {
		if err := k.RegisterAsset(ctx, a); err != nil {
			t.Fatalf("register %s: %v", a.Symbol, err)
		}
	}

	// Create pools for each (balanced reserves to ensure swaps produce output).
	pools := []struct {
		denom   string
		pnyx    int64
		asset   int64
		swapAmt int64
	}{
		{"ibc/BTC", 1_000_000, 1_000_000, 10_000},  // PNYX/BTC
		{"ibc/ETH", 1_000_000, 1_000_000, 10_000},  // PNYX/ETH
		{"ibc/LUSD", 1_000_000, 1_000_000, 10_000}, // PNYX/LUSD
	}

	for _, p := range pools {
		err := k.CreatePool(ctx, p.denom, math.NewInt(p.pnyx), math.NewInt(p.asset))
		if err != nil {
			t.Fatalf("create pool %s: %v", p.denom, err)
		}
	}

	// Verify all pools exist.
	var count int
	k.IteratePools(ctx, func(p Pool) bool {
		count++
		return false
	})
	if count != 3 {
		t.Fatalf("expected 3 pools, got %d", count)
	}

	// Swap in each pool.
	for _, p := range pools {
		out, err := k.Swap(ctx, "pnyx", math.NewInt(p.swapAmt), p.denom)
		if err != nil {
			t.Fatalf("swap PNYX→%s: %v", p.denom, err)
		}
		if !out.IsPositive() {
			t.Errorf("swap PNYX→%s: output should be positive", p.denom)
		}
	}
}

// ---------- Query enrichment ----------

func TestPoolQueryIncludesSymbol(t *testing.T) {
	k, ctx := setupKeeper(t)

	k.RegisterAsset(ctx, RegisteredAsset{IBCDenom: "ibc/BTC", Symbol: "BTC", Decimals: 8, TradingEnabled: true})
	k.CreatePool(ctx, "ibc/BTC", math.NewInt(500_000), math.NewInt(100))

	// Use the query handler.
	resp, err := k.Pool(ctx, &QueryPoolRequest{AssetDenom: "ibc/BTC"})
	if err != nil {
		t.Fatalf("query pool: %v", err)
	}

	var pool Pool
	if err := json.Unmarshal(resp.Result, &pool); err != nil {
		t.Fatalf("unmarshal pool: %v", err)
	}

	if pool.AssetSymbol != "BTC" {
		t.Errorf("AssetSymbol = %q, want BTC", pool.AssetSymbol)
	}
	if pool.AssetDenom != "ibc/BTC" {
		t.Errorf("AssetDenom = %q, want ibc/BTC", pool.AssetDenom)
	}
}

func TestPoolsQueryIncludesSymbols(t *testing.T) {
	k, ctx := setupKeeper(t)

	k.RegisterAsset(ctx, RegisteredAsset{IBCDenom: "ibc/BTC", Symbol: "BTC", Decimals: 8, TradingEnabled: true})
	k.RegisterAsset(ctx, RegisteredAsset{IBCDenom: "ibc/ETH", Symbol: "ETH", Decimals: 18, TradingEnabled: true})
	k.CreatePool(ctx, "ibc/BTC", math.NewInt(500_000), math.NewInt(100))
	k.CreatePool(ctx, "ibc/ETH", math.NewInt(300_000), math.NewInt(5_000))

	resp, err := k.Pools(ctx, &QueryPoolsRequest{})
	if err != nil {
		t.Fatalf("query pools: %v", err)
	}

	var pools []Pool
	if err := json.Unmarshal(resp.Result, &pools); err != nil {
		t.Fatalf("unmarshal pools: %v", err)
	}

	if len(pools) != 2 {
		t.Fatalf("expected 2 pools, got %d", len(pools))
	}

	symbols := make(map[string]bool)
	for _, p := range pools {
		if p.AssetSymbol == "" {
			t.Errorf("pool %s has empty AssetSymbol", p.AssetDenom)
		}
		symbols[p.AssetSymbol] = true
	}
	for _, want := range []string{"BTC", "ETH"} {
		if !symbols[want] {
			t.Errorf("missing symbol %s in pool query results", want)
		}
	}
}

// ---------- Default genesis includes ATOM ----------

func TestDefaultGenesisIncludesATOM(t *testing.T) {
	genesis := DefaultGenesisState()

	if len(genesis.RegisteredAssets) < 2 {
		t.Fatal("default genesis should include at least PNYX and ATOM")
	}

	atom := genesis.RegisteredAssets[1]
	if atom.Symbol != "ATOM" {
		t.Errorf("second default asset symbol = %s, want ATOM", atom.Symbol)
	}
	if atom.IBCDenom != "atom" {
		t.Errorf("ATOM denom = %s, want atom", atom.IBCDenom)
	}
	if atom.Decimals != 6 {
		t.Errorf("ATOM decimals = %d, want 6", atom.Decimals)
	}
	if !atom.TradingEnabled {
		t.Error("ATOM trading should be enabled by default")
	}
}
