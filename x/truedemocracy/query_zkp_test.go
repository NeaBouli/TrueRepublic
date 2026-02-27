package truedemocracy

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ---------- QueryNullifier Tests ----------

func TestQueryNullifierUsed(t *testing.T) {
	k, ctx := setupKeeper(t)
	secrets := setupDomainWithZKPIdentity(t, k, ctx, "ZKPDomain", 3)
	addProposal(t, k, ctx, "ZKPDomain", "Climate", "GreenDeal")

	proofHex, nullifierHex := generateZKPRating(t, k, ctx, "ZKPDomain", secrets, 0, "Climate", "GreenDeal")
	_, err := k.RateProposalWithZKP(ctx, "ZKPDomain", "Climate", "GreenDeal", 3, proofHex, nullifierHex)
	if err != nil {
		t.Fatalf("rating failed: %v", err)
	}

	resp, err := k.Nullifier(ctx, &QueryNullifierRequest{
		DomainName:    "ZKPDomain",
		NullifierHash: nullifierHex,
	})
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if result["used"] != true {
		t.Fatal("expected nullifier to be used")
	}
}

func TestQueryNullifierNotUsed(t *testing.T) {
	k, ctx := setupKeeper(t)
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "TestDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))

	resp, err := k.Nullifier(ctx, &QueryNullifierRequest{
		DomainName:    "TestDomain",
		NullifierHash: hex.EncodeToString(make([]byte, 32)),
	})
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if result["used"] != false {
		t.Fatal("expected nullifier to be unused")
	}
}

func TestQueryNullifierInvalidRequest(t *testing.T) {
	k, ctx := setupKeeper(t)

	_, err := k.Nullifier(ctx, &QueryNullifierRequest{
		DomainName:    "",
		NullifierHash: "aabb",
	})
	if err == nil {
		t.Fatal("expected error for empty domain name")
	}
}

// ---------- QueryPurgeSchedule Tests ----------

func TestQueryPurgeSchedule(t *testing.T) {
	k, ctx := setupKeeper(t)
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "TestDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))
	k.InitializeBigPurgeSchedule(ctx, "TestDomain")

	resp, err := k.PurgeSchedule(ctx, &QueryPurgeScheduleRequest{
		DomainName: "TestDomain",
	})
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	var schedule BigPurgeSchedule
	if err := json.Unmarshal(resp.Result, &schedule); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if schedule.DomainName != "TestDomain" {
		t.Fatalf("expected domain TestDomain, got %s", schedule.DomainName)
	}
	if schedule.PurgeInterval != DefaultPurgeInterval {
		t.Fatalf("expected interval %d, got %d", DefaultPurgeInterval, schedule.PurgeInterval)
	}
	if schedule.AnnouncementLead != DefaultAnnouncementLead {
		t.Fatalf("expected lead %d, got %d", DefaultAnnouncementLead, schedule.AnnouncementLead)
	}
}

func TestQueryPurgeScheduleMissingDomain(t *testing.T) {
	k, ctx := setupKeeper(t)

	_, err := k.PurgeSchedule(ctx, &QueryPurgeScheduleRequest{
		DomainName: "NoDomain",
	})
	if err == nil {
		t.Fatal("expected error for missing domain")
	}
}

// ---------- QueryZKPState Tests ----------

func TestQueryZKPState(t *testing.T) {
	k, ctx := setupKeeper(t)
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "ZKPDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))

	// Add 3 members with identity commitments.
	for i := 0; i < 3; i++ {
		memberAddr := sdk.AccAddress("member" + string(rune('A'+i))).String()
		k.AddMember(ctx, "ZKPDomain", memberAddr, admin)

		secret := big.NewInt(int64(i + 100)).Bytes()
		commitment, _ := ComputeCommitment(secret)
		commitHex := hex.EncodeToString(commitment)
		k.RegisterIdentityCommitment(ctx, "ZKPDomain", memberAddr, commitHex)
	}

	resp, err := k.ZKPState(ctx, &QueryZKPStateRequest{
		DomainName: "ZKPDomain",
	})
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	var state ZKPDomainState
	if err := json.Unmarshal(resp.Result, &state); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if state.DomainName != "ZKPDomain" {
		t.Fatalf("expected domain ZKPDomain, got %s", state.DomainName)
	}
	if state.CommitmentCount != 3 {
		t.Fatalf("expected 3 commitments, got %d", state.CommitmentCount)
	}
	if state.MemberCount != 4 { // admin + 3 added members
		t.Fatalf("expected 4 members, got %d", state.MemberCount)
	}
	if state.MerkleRoot == "" {
		t.Fatal("expected non-empty Merkle root")
	}
}

func TestQueryZKPStateEmpty(t *testing.T) {
	k, ctx := setupKeeper(t)
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "EmptyDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))

	resp, err := k.ZKPState(ctx, &QueryZKPStateRequest{
		DomainName: "EmptyDomain",
	})
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	var state ZKPDomainState
	if err := json.Unmarshal(resp.Result, &state); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if state.CommitmentCount != 0 {
		t.Fatalf("expected 0 commitments, got %d", state.CommitmentCount)
	}
	if state.MerkleRoot != "" {
		t.Fatal("expected empty Merkle root")
	}
	if state.VKInitialized {
		t.Fatal("VK should not be initialized")
	}
}

func TestQueryZKPStateMissingDomain(t *testing.T) {
	k, ctx := setupKeeper(t)

	_, err := k.ZKPState(ctx, &QueryZKPStateRequest{
		DomainName: "NoDomain",
	})
	if err == nil {
		t.Fatal("expected error for missing domain")
	}
}
