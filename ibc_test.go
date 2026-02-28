package main

import (
	"context"
	"testing"
	"time"

	upgradetypes "cosmossdk.io/x/upgrade/types"
)

// --- IBCStakingKeeper Tests ---

func TestIBCStakingKeeper_UnbondingTime(t *testing.T) {
	k := IBCStakingKeeper{}
	dur, err := k.UnbondingTime(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := 3 * 7 * 24 * time.Hour
	if dur != expected {
		t.Fatalf("expected %v, got %v", expected, dur)
	}
}

func TestIBCStakingKeeper_GetHistoricalInfo(t *testing.T) {
	k := IBCStakingKeeper{}
	_, err := k.GetHistoricalInfo(context.Background(), 100)
	if err == nil {
		t.Fatal("expected error for GetHistoricalInfo (stub)")
	}
}

// --- IBCUpgradeKeeper Tests ---

func TestIBCUpgradeKeeper_GetUpgradePlan(t *testing.T) {
	k := IBCUpgradeKeeper{}
	plan, err := k.GetUpgradePlan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.Name != "" {
		t.Fatal("expected empty plan name")
	}
}

func TestIBCUpgradeKeeper_ClearIBCState(t *testing.T) {
	k := IBCUpgradeKeeper{}
	if err := k.ClearIBCState(context.Background(), 100); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIBCUpgradeKeeper_ScheduleUpgrade(t *testing.T) {
	k := IBCUpgradeKeeper{}
	err := k.ScheduleUpgrade(context.Background(), upgradetypes.Plan{Name: "test"})
	if err == nil {
		t.Fatal("expected error for ScheduleUpgrade (not supported)")
	}
}

func TestIBCUpgradeKeeper_ClientState(t *testing.T) {
	k := IBCUpgradeKeeper{}

	// GetUpgradedClient returns nil (no upgraded client)
	bz, err := k.GetUpgradedClient(context.Background(), 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bz != nil {
		t.Fatal("expected nil for GetUpgradedClient")
	}

	// SetUpgradedClient is a no-op
	if err := k.SetUpgradedClient(context.Background(), 100, []byte("test")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIBCUpgradeKeeper_ConsensusState(t *testing.T) {
	k := IBCUpgradeKeeper{}

	// GetUpgradedConsensusState returns nil
	bz, err := k.GetUpgradedConsensusState(context.Background(), 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bz != nil {
		t.Fatal("expected nil for GetUpgradedConsensusState")
	}

	// SetUpgradedConsensusState is a no-op
	if err := k.SetUpgradedConsensusState(context.Background(), 100, []byte("test")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Module configuration tests ---

func TestModuleBasicsIncludesIBC(t *testing.T) {
	// Verify IBC-related modules are registered in ModuleBasics.
	names := make(map[string]bool)
	for name := range ModuleBasics {
		names[name] = true
	}

	required := []string{"ibc", "transfer", "capability"}
	for _, name := range required {
		if !names[name] {
			t.Fatalf("ModuleBasics missing required module: %s", name)
		}
	}
}

func TestMaccPermsIncludesTransfer(t *testing.T) {
	perms, ok := maccPerms["transfer"]
	if !ok {
		t.Fatal("maccPerms missing transfer module")
	}
	// Transfer module needs Minter and Burner permissions for ICS-20.
	hasMinter, hasBurner := false, false
	for _, p := range perms {
		if p == "minter" {
			hasMinter = true
		}
		if p == "burner" {
			hasBurner = true
		}
	}
	if !hasMinter || !hasBurner {
		t.Fatalf("transfer perms should include minter+burner, got %v", perms)
	}
}
