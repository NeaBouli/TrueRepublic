package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
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

// --- IBC integration tests (Milestone 7.2) ---

func TestIBCDefaultGenesis(t *testing.T) {
	// Verify ModuleBasics produces valid default genesis JSON for IBC modules.
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	ModuleBasics.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	defaults := ModuleBasics.DefaultGenesis(cdc)

	// All IBC modules must have default genesis entries.
	required := []string{
		ibcexported.ModuleName,
		transfertypes.ModuleName,
		capabilitytypes.ModuleName,
	}
	for _, name := range required {
		data, ok := defaults[name]
		if !ok {
			t.Fatalf("DefaultGenesis missing module: %s", name)
		}
		if len(data) == 0 {
			t.Fatalf("DefaultGenesis for %s is empty", name)
		}
		// Verify it's valid JSON.
		if !json.Valid(data) {
			t.Fatalf("DefaultGenesis for %s is not valid JSON", name)
		}
	}
}

func TestIBCTransferPortID(t *testing.T) {
	// The ICS-20 transfer module uses port "transfer".
	if transfertypes.PortID != "transfer" {
		t.Fatalf("expected PortID 'transfer', got %q", transfertypes.PortID)
	}
}

func TestIBCTransferDenomTrace(t *testing.T) {
	// Verify IBC denom trace hash computation.
	trace := transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "pnyx",
	}

	ibcDenom := trace.IBCDenom()
	// IBC denom format: ibc/<SHA256-HASH>
	if len(ibcDenom) < 5 || ibcDenom[:4] != "ibc/" {
		t.Fatalf("expected ibc/ prefix, got %q", ibcDenom)
	}

	// Hash should be deterministic.
	ibcDenom2 := trace.IBCDenom()
	if ibcDenom != ibcDenom2 {
		t.Fatal("IBCDenom() is not deterministic")
	}

	// Different path should produce different denom.
	trace2 := transfertypes.DenomTrace{
		Path:      "transfer/channel-1",
		BaseDenom: "pnyx",
	}
	if trace.IBCDenom() == trace2.IBCDenom() {
		t.Fatal("different paths should produce different IBC denoms")
	}
}

func TestIBCEscrowAddress(t *testing.T) {
	// Verify escrow address generation is deterministic and non-empty.
	addr := transfertypes.GetEscrowAddress("transfer", "channel-0")
	if addr.Empty() {
		t.Fatal("escrow address should not be empty")
	}

	// Same inputs should produce same address.
	addr2 := transfertypes.GetEscrowAddress("transfer", "channel-0")
	if !addr.Equals(addr2) {
		t.Fatal("escrow address should be deterministic")
	}

	// Different channel should produce different address.
	addr3 := transfertypes.GetEscrowAddress("transfer", "channel-1")
	if addr.Equals(addr3) {
		t.Fatal("different channels should have different escrow addresses")
	}
}

func TestIBCStoreKeys(t *testing.T) {
	// Verify the expected store key constants.
	if ibcexported.StoreKey != "ibc" {
		t.Fatalf("expected IBC store key 'ibc', got %q", ibcexported.StoreKey)
	}
	if transfertypes.StoreKey != "transfer" {
		t.Fatalf("expected transfer store key 'transfer', got %q", transfertypes.StoreKey)
	}
	if capabilitytypes.StoreKey != "capability" {
		t.Fatalf("expected capability store key 'capability', got %q", capabilitytypes.StoreKey)
	}
	if capabilitytypes.MemStoreKey != "memory:capability" {
		t.Fatalf("expected capability mem key 'memory:capability', got %q", capabilitytypes.MemStoreKey)
	}
}

func TestIBCTransferDefaultParams(t *testing.T) {
	// Verify default transfer params have send and receive enabled.
	params := transfertypes.DefaultParams()
	if !params.SendEnabled {
		t.Fatal("default transfer params should have SendEnabled=true")
	}
	if !params.ReceiveEnabled {
		t.Fatal("default transfer params should have ReceiveEnabled=true")
	}
}
