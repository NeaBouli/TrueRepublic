package dex

import (
	"encoding/json"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	abci "github.com/cometbft/cometbft/abci/types"
)

// ---------- Pool Stats ----------

func TestQuerierPoolStats(t *testing.T) {
	k, ctx := setupKeeperWithDefaults(t)
	err := k.CreatePool(ctx, "atom", math.NewInt(1_000_000), math.NewInt(1_000_000))
	if err != nil {
		t.Fatalf("CreatePool: %v", err)
	}

	cdc := codec.NewLegacyAmino()
	RegisterCodec(cdc)
	querier := NewQuerier(k, cdc)

	bz, err := querier(ctx, []string{"pool_stats", "atom"}, abci.RequestQuery{})
	if err != nil {
		t.Fatalf("querier returned error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bz, &result); err != nil {
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

func TestQuerierSpotPrice(t *testing.T) {
	k, ctx := setupKeeperWithDefaults(t)
	err := k.CreatePool(ctx, "atom", math.NewInt(1_000_000), math.NewInt(1_000_000))
	if err != nil {
		t.Fatalf("CreatePool: %v", err)
	}

	cdc := codec.NewLegacyAmino()
	RegisterCodec(cdc)
	querier := NewQuerier(k, cdc)

	bz, err := querier(ctx, []string{"spot_price", "pnyx", "atom"}, abci.RequestQuery{})
	if err != nil {
		t.Fatalf("querier returned error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bz, &result); err != nil {
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
	if route[0] != "pnyx" || route[1] != "atom" {
		t.Errorf("route = %v, want [pnyx atom]", route)
	}

	// Price per million should be positive and reasonable for an equal pool.
	priceStr, _ := result["price_per_million"].(string)
	if priceStr == "" || priceStr == "0" {
		t.Errorf("price_per_million should be positive, got %q", priceStr)
	}
}

// ---------- Liquidity Depth ----------

func TestQuerierLiquidityDepth(t *testing.T) {
	k, ctx := setupKeeperWithDefaults(t)
	err := k.CreatePool(ctx, "atom", math.NewInt(1_000_000), math.NewInt(1_000_000))
	if err != nil {
		t.Fatalf("CreatePool: %v", err)
	}

	cdc := codec.NewLegacyAmino()
	RegisterCodec(cdc)
	querier := NewQuerier(k, cdc)

	bz, err := querier(ctx, []string{"liquidity_depth", "pnyx", "atom"}, abci.RequestQuery{})
	if err != nil {
		t.Fatalf("querier returned error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bz, &result); err != nil {
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

func TestQuerierLPPosition(t *testing.T) {
	k, ctx := setupKeeperWithDefaults(t)
	err := k.CreatePool(ctx, "atom", math.NewInt(1_000_000), math.NewInt(1_000_000))
	if err != nil {
		t.Fatalf("CreatePool: %v", err)
	}

	cdc := codec.NewLegacyAmino()
	RegisterCodec(cdc)
	querier := NewQuerier(k, cdc)

	// Query for half the pool's shares (500000 for a 1M/1M pool with TotalShares=1000000).
	bz, err := querier(ctx, []string{"lp_position", "atom", "500000"}, abci.RequestQuery{})
	if err != nil {
		t.Fatalf("querier returned error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bz, &result); err != nil {
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

func TestQuerierEstimateSwap(t *testing.T) {
	k, ctx := setupKeeperWithDefaults(t)
	err := k.CreatePool(ctx, "atom", math.NewInt(1_000_000), math.NewInt(1_000_000))
	if err != nil {
		t.Fatalf("CreatePool: %v", err)
	}

	cdc := codec.NewLegacyAmino()
	RegisterCodec(cdc)
	querier := NewQuerier(k, cdc)

	bz, err := querier(ctx, []string{"estimate_swap", "pnyx", "1000", "atom"}, abci.RequestQuery{})
	if err != nil {
		t.Fatalf("querier returned error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bz, &result); err != nil {
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

func TestQuerierPoolStatsMissingArgs(t *testing.T) {
	k, ctx := setupKeeperWithDefaults(t)

	cdc := codec.NewLegacyAmino()
	RegisterCodec(cdc)
	querier := NewQuerier(k, cdc)

	// Call pool_stats without the asset denom argument.
	_, err := querier(ctx, []string{"pool_stats"}, abci.RequestQuery{})
	if err == nil {
		t.Fatal("expected error for missing path args")
	}
}

func TestQuerierUnknownPath(t *testing.T) {
	k, ctx := setupKeeperWithDefaults(t)

	cdc := codec.NewLegacyAmino()
	RegisterCodec(cdc)
	querier := NewQuerier(k, cdc)

	_, err := querier(ctx, []string{"nonexistent_route"}, abci.RequestQuery{})
	if err == nil {
		t.Fatal("expected error for unknown query path")
	}
}

func TestQuerierPoolNotFound(t *testing.T) {
	k, ctx := setupKeeperWithDefaults(t)

	cdc := codec.NewLegacyAmino()
	RegisterCodec(cdc)
	querier := NewQuerier(k, cdc)

	// Query pool_stats for a pool that does not exist.
	_, err := querier(ctx, []string{"pool_stats", "nonexistent"}, abci.RequestQuery{})
	if err == nil {
		t.Fatal("expected error for nonexistent pool")
	}
}
