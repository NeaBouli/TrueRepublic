package truedemocracy

import (
	"encoding/hex"
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ---------- RegisterIdentityCommitment Keeper Tests ----------

func TestRegisterIdentityCommitment(t *testing.T) {
	k, ctx := setupKeeper(t)

	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "ZKPDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))
	k.AddMember(ctx, "ZKPDomain", "alice", admin)

	// Compute a valid commitment.
	secret := big.NewInt(42).Bytes()
	commitment, _ := ComputeCommitment(secret)
	commitHex := hex.EncodeToString(commitment)

	err := k.RegisterIdentityCommitment(ctx, "ZKPDomain", "alice", commitHex)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	domain, _ := k.GetDomain(ctx, "ZKPDomain")
	if len(domain.IdentityCommits) != 1 {
		t.Fatalf("expected 1 commitment, got %d", len(domain.IdentityCommits))
	}
	if domain.IdentityCommits[0] != commitHex {
		t.Fatal("commitment mismatch")
	}
	if domain.MerkleRoot == "" {
		t.Fatal("MerkleRoot should be set after registration")
	}
}

func TestRegisterIdentityNonMemberRejected(t *testing.T) {
	k, ctx := setupKeeper(t)

	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "ZKPDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))

	secret := big.NewInt(42).Bytes()
	commitment, _ := ComputeCommitment(secret)
	commitHex := hex.EncodeToString(commitment)

	err := k.RegisterIdentityCommitment(ctx, "ZKPDomain", "outsider", commitHex)
	if err == nil {
		t.Fatal("expected error for non-member")
	}
}

func TestRegisterIdentityDuplicateRejected(t *testing.T) {
	k, ctx := setupKeeper(t)

	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "ZKPDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))
	k.AddMember(ctx, "ZKPDomain", "alice", admin)

	secret := big.NewInt(42).Bytes()
	commitment, _ := ComputeCommitment(secret)
	commitHex := hex.EncodeToString(commitment)

	k.RegisterIdentityCommitment(ctx, "ZKPDomain", "alice", commitHex)
	err := k.RegisterIdentityCommitment(ctx, "ZKPDomain", "alice", commitHex)
	if err == nil {
		t.Fatal("expected error for duplicate commitment")
	}
}

func TestRegisterIdentityInvalidHex(t *testing.T) {
	k, ctx := setupKeeper(t)

	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "ZKPDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))
	k.AddMember(ctx, "ZKPDomain", "alice", admin)

	err := k.RegisterIdentityCommitment(ctx, "ZKPDomain", "alice", "not-valid-hex!!!")
	if err == nil {
		t.Fatal("expected error for invalid hex")
	}
}

func TestRegisterIdentityWrongLength(t *testing.T) {
	k, ctx := setupKeeper(t)

	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "ZKPDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))
	k.AddMember(ctx, "ZKPDomain", "alice", admin)

	err := k.RegisterIdentityCommitment(ctx, "ZKPDomain", "alice", "aabb") // too short
	if err == nil {
		t.Fatal("expected error for wrong-length commitment")
	}
}

func TestRegisterIdentityUnknownDomain(t *testing.T) {
	k, ctx := setupKeeper(t)

	secret := big.NewInt(42).Bytes()
	commitment, _ := ComputeCommitment(secret)
	commitHex := hex.EncodeToString(commitment)

	err := k.RegisterIdentityCommitment(ctx, "NoDomain", "alice", commitHex)
	if err == nil {
		t.Fatal("expected error for unknown domain")
	}
}

func TestMerkleRootUpdatesOnRegistration(t *testing.T) {
	k, ctx := setupKeeper(t)

	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "ZKPDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))
	k.AddMember(ctx, "ZKPDomain", "alice", admin)
	k.AddMember(ctx, "ZKPDomain", "bob", admin)
	k.AddMember(ctx, "ZKPDomain", "charlie", admin)

	// Register three commitments, root should change each time.
	var roots []string
	for i, member := range []string{"alice", "bob", "charlie"} {
		secret := big.NewInt(int64(i + 100)).Bytes()
		commitment, _ := ComputeCommitment(secret)
		commitHex := hex.EncodeToString(commitment)
		k.RegisterIdentityCommitment(ctx, "ZKPDomain", member, commitHex)

		domain, _ := k.GetDomain(ctx, "ZKPDomain")
		roots = append(roots, domain.MerkleRoot)
	}

	// All roots should be different.
	if roots[0] == roots[1] || roots[1] == roots[2] || roots[0] == roots[2] {
		t.Fatalf("Merkle roots should change with each registration: %v", roots)
	}
}

// ---------- Nullifier Store Tests ----------

func TestNullifierStorage(t *testing.T) {
	k, ctx := setupKeeper(t)

	k.SetNullifierUsed(ctx, "TestDomain", "aabbccdd", 100)
	if !k.IsNullifierUsed(ctx, "TestDomain", "aabbccdd") {
		t.Fatal("nullifier should be marked as used")
	}
}

func TestNullifierNotUsedInitially(t *testing.T) {
	k, ctx := setupKeeper(t)

	if k.IsNullifierUsed(ctx, "TestDomain", "unknown") {
		t.Fatal("nullifier should not be used initially")
	}
}

func TestPurgeNullifiers(t *testing.T) {
	k, ctx := setupKeeper(t)

	k.SetNullifierUsed(ctx, "TestDomain", "null1", 100)
	k.SetNullifierUsed(ctx, "TestDomain", "null2", 101)
	k.SetNullifierUsed(ctx, "TestDomain", "null3", 102)

	k.PurgeNullifiers(ctx, "TestDomain")

	if k.IsNullifierUsed(ctx, "TestDomain", "null1") {
		t.Fatal("null1 should be purged")
	}
	if k.IsNullifierUsed(ctx, "TestDomain", "null2") {
		t.Fatal("null2 should be purged")
	}
	if k.IsNullifierUsed(ctx, "TestDomain", "null3") {
		t.Fatal("null3 should be purged")
	}
}

// ---------- Big Purge Integration Tests ----------

func TestBigPurgeClearsIdentityCommits(t *testing.T) {
	k, ctx := setupKeeper(t)

	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "PurgeDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))
	k.AddMember(ctx, "PurgeDomain", "alice", admin)

	// Register a commitment.
	secret := big.NewInt(42).Bytes()
	commitment, _ := ComputeCommitment(secret)
	commitHex := hex.EncodeToString(commitment)
	k.RegisterIdentityCommitment(ctx, "PurgeDomain", "alice", commitHex)

	domain, _ := k.GetDomain(ctx, "PurgeDomain")
	if len(domain.IdentityCommits) != 1 {
		t.Fatal("should have 1 commitment before purge")
	}

	// Execute big purge.
	k.executeBigPurge(ctx, "PurgeDomain")

	domain, _ = k.GetDomain(ctx, "PurgeDomain")
	if len(domain.IdentityCommits) != 0 {
		t.Fatalf("expected 0 commitments after purge, got %d", len(domain.IdentityCommits))
	}
	if domain.MerkleRoot != "" {
		t.Fatalf("MerkleRoot should be empty after purge, got %s", domain.MerkleRoot)
	}
	// Members should still be present.
	if len(domain.Members) == 0 {
		t.Fatal("members should be preserved after purge")
	}
}

func TestBigPurgeClearsNullifiers(t *testing.T) {
	k, ctx := setupKeeper(t)

	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "PurgeDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))

	// Store some nullifiers.
	k.SetNullifierUsed(ctx, "PurgeDomain", "n1", 100)
	k.SetNullifierUsed(ctx, "PurgeDomain", "n2", 101)

	// Execute big purge.
	k.executeBigPurge(ctx, "PurgeDomain")

	if k.IsNullifierUsed(ctx, "PurgeDomain", "n1") {
		t.Fatal("n1 should be purged")
	}
	if k.IsNullifierUsed(ctx, "PurgeDomain", "n2") {
		t.Fatal("n2 should be purged")
	}
}

// ---------- MsgRegisterIdentity ValidateBasic ----------

func TestMsgRegisterIdentityValidateBasic(t *testing.T) {
	validCommitment := hex.EncodeToString(make([]byte, 32))

	t.Run("valid message", func(t *testing.T) {
		msg := MsgRegisterIdentity{
			Sender:     sdk.AccAddress("alice"),
			DomainName: "TestDomain",
			Commitment: validCommitment,
		}
		if err := msg.ValidateBasic(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty domain rejected", func(t *testing.T) {
		msg := MsgRegisterIdentity{
			Sender:     sdk.AccAddress("alice"),
			DomainName: "",
			Commitment: validCommitment,
		}
		if err := msg.ValidateBasic(); err == nil {
			t.Fatal("expected error for empty domain")
		}
	})

	t.Run("empty commitment rejected", func(t *testing.T) {
		msg := MsgRegisterIdentity{
			Sender:     sdk.AccAddress("alice"),
			DomainName: "TestDomain",
			Commitment: "",
		}
		if err := msg.ValidateBasic(); err == nil {
			t.Fatal("expected error for empty commitment")
		}
	})

	t.Run("short commitment rejected", func(t *testing.T) {
		msg := MsgRegisterIdentity{
			Sender:     sdk.AccAddress("alice"),
			DomainName: "TestDomain",
			Commitment: "aabb",
		}
		if err := msg.ValidateBasic(); err == nil {
			t.Fatal("expected error for short commitment")
		}
	})

	t.Run("invalid hex rejected", func(t *testing.T) {
		msg := MsgRegisterIdentity{
			Sender:     sdk.AccAddress("alice"),
			DomainName: "TestDomain",
			Commitment: "gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg",
		}
		if err := msg.ValidateBasic(); err == nil {
			t.Fatal("expected error for invalid hex")
		}
	})
}

// ---------- MsgServer RegisterIdentity ----------

func TestMsgServerRegisterIdentity(t *testing.T) {
	k, ctx := setupKeeper(t)
	srv := NewMsgServer(k)

	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "ZKPDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))
	k.AddMember(ctx, "ZKPDomain", sdk.AccAddress("alice").String(), admin)

	secret := big.NewInt(42).Bytes()
	commitment, _ := ComputeCommitment(secret)
	commitHex := hex.EncodeToString(commitment)

	msg := &MsgRegisterIdentity{
		Sender:     sdk.AccAddress("alice"),
		DomainName: "ZKPDomain",
		Commitment: commitHex,
	}

	_, err := srv.RegisterIdentity(ctx, msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify commitment is stored.
	domain, _ := k.GetDomain(ctx, "ZKPDomain")
	if len(domain.IdentityCommits) != 1 {
		t.Fatalf("expected 1 commitment, got %d", len(domain.IdentityCommits))
	}
	if domain.MerkleRoot == "" {
		t.Fatal("MerkleRoot should be set")
	}
}
