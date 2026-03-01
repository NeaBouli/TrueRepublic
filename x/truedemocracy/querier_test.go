package truedemocracy

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/cometbft/cometbft/abci/types"
)

// ---------- QueryZKPState (ABCI querier) ----------

func TestQuerierZKPState(t *testing.T) {
	k, ctx := setupKeeper(t)
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "TestDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))

	cdc := codec.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(cdc)
	RegisterCodec(cdc)
	querier := NewQuerier(k, cdc)

	bz, err := querier(ctx, []string{"zkp_state", "TestDomain"}, abci.RequestQuery{})
	if err != nil {
		t.Fatalf("querier returned error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bz, &result); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if result["domain_name"] != "TestDomain" {
		t.Fatalf("expected domain_name TestDomain, got %v", result["domain_name"])
	}
	if result["commitment_count"] != float64(0) {
		t.Fatalf("expected commitment_count 0, got %v", result["commitment_count"])
	}
	// Creator is auto-added as a member.
	if result["member_count"] != float64(1) {
		t.Fatalf("expected member_count 1, got %v", result["member_count"])
	}
	if _, ok := result["merkle_root"]; !ok {
		t.Fatal("expected merkle_root field in response")
	}
	if _, ok := result["vk_initialized"]; !ok {
		t.Fatal("expected vk_initialized field in response")
	}
}

// ---------- QueryNullifier (ABCI querier) ----------

func TestQuerierNullifier(t *testing.T) {
	k, ctx := setupKeeper(t)
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "TestDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))

	cdc := codec.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(cdc)
	RegisterCodec(cdc)
	querier := NewQuerier(k, cdc)

	bz, err := querier(ctx, []string{"nullifier", "TestDomain", "abc123"}, abci.RequestQuery{})
	if err != nil {
		t.Fatalf("querier returned error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bz, &result); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if result["domain_name"] != "TestDomain" {
		t.Fatalf("expected domain_name TestDomain, got %v", result["domain_name"])
	}
	if result["nullifier_hash"] != "abc123" {
		t.Fatalf("expected nullifier_hash abc123, got %v", result["nullifier_hash"])
	}
	if result["used"] != false {
		t.Fatal("expected used to be false for unused nullifier")
	}
}

// ---------- QueryPurgeSchedule (ABCI querier) ----------

func TestQuerierPurgeSchedule(t *testing.T) {
	k, ctx := setupKeeper(t)
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "TestDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))

	k.SetBigPurgeSchedule(ctx, BigPurgeSchedule{
		DomainName:       "TestDomain",
		NextPurgeTime:    1000,
		PurgeInterval:    7776000,
		AnnouncementLead: 604800,
	})

	cdc := codec.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(cdc)
	RegisterCodec(cdc)
	querier := NewQuerier(k, cdc)

	bz, err := querier(ctx, []string{"purge_schedule", "TestDomain"}, abci.RequestQuery{})
	if err != nil {
		t.Fatalf("querier returned error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bz, &result); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if result["domain_name"] != "TestDomain" {
		t.Fatalf("expected domain_name TestDomain, got %v", result["domain_name"])
	}
	if result["next_purge_time"] != float64(1000) {
		t.Fatalf("expected next_purge_time 1000, got %v", result["next_purge_time"])
	}
}

// ---------- Error cases ----------

func TestQuerierZKPStateDomainNotFound(t *testing.T) {
	k, ctx := setupKeeper(t)

	cdc := codec.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(cdc)
	RegisterCodec(cdc)
	querier := NewQuerier(k, cdc)

	_, err := querier(ctx, []string{"zkp_state", "NonExistent"}, abci.RequestQuery{})
	if err == nil {
		t.Fatal("expected error for non-existent domain")
	}
}

func TestQuerierNullifierMissingPath(t *testing.T) {
	k, ctx := setupKeeper(t)

	cdc := codec.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(cdc)
	RegisterCodec(cdc)
	querier := NewQuerier(k, cdc)

	// Only provide domain name, missing the nullifier hash.
	_, err := querier(ctx, []string{"nullifier", "TestDomain"}, abci.RequestQuery{})
	if err == nil {
		t.Fatal("expected error for missing nullifier hash in path")
	}
}

func TestQuerierPurgeScheduleNotFound(t *testing.T) {
	k, ctx := setupKeeper(t)

	cdc := codec.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(cdc)
	RegisterCodec(cdc)
	querier := NewQuerier(k, cdc)

	// No domain created, so no purge schedule exists.
	_, err := querier(ctx, []string{"purge_schedule", "NoDomain"}, abci.RequestQuery{})
	if err == nil {
		t.Fatal("expected error for missing purge schedule")
	}
}
