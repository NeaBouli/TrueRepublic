package dex

import (
	"encoding/json"
	"testing"

	"cosmossdk.io/math"
)

// ---------- Pool Stats ----------

func TestQueryServerPoolStats(t *testing.T) {
	k, ctx := setupKeeperWithDefaults(t)
	err := k.CreatePool(ctx, "atom", math.NewInt(1_000_000), math.NewInt(1_000_000))
	if err != nil {
		t.Fatalf("CreatePool: %v", err)
	}

	resp, err := k.PoolStats(ctx, &QueryPoolStatsRequest{AssetDenom: "atom"})
	if err != nil {
		t.Fatalf("query server returned error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Check expected fields exist.
	for _, field := range []string{
		"asset_denom", "swap_count", "total_volume_pnyx",
		"pnyx_reserve", "asset_reserve", "spot_price_per_million", "total_shares",
	} {
		if _, ok := result[field]; !ok {
			t.Errorf("missing field %q in pool_stats response", field)
		}
	}

	if result["asset_denom"] != "atom" {
		t.Errorf("asset_denom = %v, want atom", result["asset_denom"])
	}
	if result["pnyx_reserve"] != "1000000" {
		t.Errorf("pnyx_reserve = %v, want 1000000", result["pnyx_reserve"])
	}
	if result["asset_reserve"] != "1000000" {
		t.Errorf("asset_reserve = %v, want 1000000", result["asset_reserve"])
	}
}

// ---------- Spot Price ----------

func TestQueryServerSpotPrice(t *testing.T) {
	k, ctx := setupKeeperWithDefaults(t)
	err := k.CreatePool(ctx, "atom", math.NewInt(1_000_000), math.NewInt(1_000_000))
	if err != nil {
		t.Fatalf("CreatePool: %v", err)
	}

	resp, err := k.SpotPrice(ctx, &QuerySpotPriceRequest{InputDenom: pnyxDenom, OutputDenom: "atom"})
	if err != nil {
		t.Fatalf("query server returned error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	for _, field := range []string{
		"price_per_million", "input_symbol", "output_symbol", "route",
	} {
		if _, ok := result[field]; !ok {
			t.Errorf("missing field %q in spot_price response", field)
		}
	}

	// Direct route pnyx -> atom should have 2 elements.
	route, ok := result["route"].([]interface{})
	if !ok {
		t.Fatal("route is not an array")
	}
	if len(route) != 2 {
		t.Errorf("route length = %d, want 2", len(route))
	}
	if route[0] != pnyxDenom || route[1] != "atom" {
		t.Errorf("route = %v, want [pnyx atom]", route)
	}

	// Price per million should be positive and reasonable for an equal pool.
	priceStr, _ := result["price_per_million"].(string)
	if priceStr == "" || priceStr == "0" {
		t.Errorf("price_per_million should be positive, got %q", priceStr)
	}
}

// ---------- Liquidity Depth ----------

func TestQueryServerLiquidityDepth(t *testing.T) {
	k, ctx := setupKeeperWithDefaults(t)
	err := k.CreatePool(ctx, "atom", math.NewInt(1_000_000), math.NewInt(1_000_000))
	if err != nil {
		t.Fatalf("CreatePool: %v", err)
	}

	resp, err := k.LiquidityDepth(ctx, &QueryLiquidityDepthRequest{InputDenom: pnyxDenom, OutputDenom: "atom"})
	if err != nil {
		t.Fatalf("query server returned error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if _, ok := result["levels"]; !ok {
		t.Fatal("missing field 'levels' in liquidity_depth response")
	}

	levels, ok := result["levels"].([]interface{})
	if !ok {
		t.Fatal("levels is not an array")
	}
	if len(levels) == 0 {
		t.Error("expected at least one depth level")
	}
}

// ---------- LP Position ----------

func TestQueryServerLPPosition(t *testing.T) {
	k, ctx := setupKeeperWithDefaults(t)
	err := k.CreatePool(ctx, "atom", math.NewInt(1_000_000), math.NewInt(1_000_000))
	if err != nil {
		t.Fatalf("CreatePool: %v", err)
	}

	// Query for half the pool's shares (500000 for a 1M/1M pool with TotalShares=1000000).
	resp, err := k.LPPosition(ctx, &QueryLPPositionRequest{AssetDenom: "atom", Shares: 500000})
	if err != nil {
		t.Fatalf("query server returned error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	for _, field := range []string{
		"pnyx_value", "asset_value", "share_of_pool_bps",
	} {
		if _, ok := result[field]; !ok {
			t.Errorf("missing field %q in lp_position response", field)
		}
	}

	if result["pnyx_value"] != "500000" {
		t.Errorf("pnyx_value = %v, want 500000", result["pnyx_value"])
	}
	if result["asset_value"] != "500000" {
		t.Errorf("asset_value = %v, want 500000", result["asset_value"])
	}

	// share_of_pool_bps for 50% should be 5000.
	bpsFloat, ok := result["share_of_pool_bps"].(float64)
	if !ok {
		t.Fatalf("share_of_pool_bps is not a number: %T", result["share_of_pool_bps"])
	}
	if int64(bpsFloat) != 5000 {
		t.Errorf("share_of_pool_bps = %v, want 5000", bpsFloat)
	}
}

// ---------- Estimate Swap ----------

func TestQueryServerEstimateSwap(t *testing.T) {
	k, ctx := setupKeeperWithDefaults(t)
	err := k.CreatePool(ctx, "atom", math.NewInt(1_000_000), math.NewInt(1_000_000))
	if err != nil {
		t.Fatalf("CreatePool: %v", err)
	}

	resp, err := k.EstimateSwap(ctx, &QueryEstimateSwapRequest{InputDenom: pnyxDenom, InputAmt: 1000, OutputDenom: "atom"})
	if err != nil {
		t.Fatalf("query server returned error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	for _, field := range []string{
		"expected_output", "route", "hops",
	} {
		if _, ok := result[field]; !ok {
			t.Errorf("missing field %q in estimate_swap response", field)
		}
	}

	// Expected output should be a positive number string.
	outStr, _ := result["expected_output"].(string)
	if outStr == "" || outStr == "0" {
		t.Errorf("expected_output should be positive, got %q", outStr)
	}

	// Direct swap pnyx -> atom: route should have 2 elements, 1 hop.
	route, ok := result["route"].([]interface{})
	if !ok {
		t.Fatal("route is not an array")
	}
	if len(route) != 2 {
		t.Errorf("route length = %d, want 2", len(route))
	}

	hopsFloat, ok := result["hops"].(float64)
	if !ok {
		t.Fatalf("hops is not a number: %T", result["hops"])
	}
	if int(hopsFloat) != 1 {
		t.Errorf("hops = %v, want 1", hopsFloat)
	}
}

// ---------- Error Cases ----------

func TestQueryServerPoolStatsMissingRequest(t *testing.T) {
	k, ctx := setupKeeperWithDefaults(t)

	_, err := k.PoolStats(ctx, &QueryPoolStatsRequest{})
	if err == nil {
		t.Fatal("expected error for missing asset denom")
	}
}

func TestQueryServerEstimateSwapRejectsInvalidRequest(t *testing.T) {
	k, ctx := setupKeeperWithDefaults(t)

	_, err := k.EstimateSwap(ctx, &QueryEstimateSwapRequest{InputDenom: pnyxDenom, OutputDenom: "atom"})
	if err == nil {
		t.Fatal("expected error for non-positive input amount")
	}
}

func TestQueryServerPoolStatsNotFound(t *testing.T) {
	k, ctx := setupKeeperWithDefaults(t)

	_, err := k.PoolStats(ctx, &QueryPoolStatsRequest{AssetDenom: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent pool")
	}
}
