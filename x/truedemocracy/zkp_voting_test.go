package truedemocracy

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ---------- ComputeExternalNullifier Tests ----------

func TestComputeExternalNullifier(t *testing.T) {
	result, err := ComputeExternalNullifier("TestDomain|Climate|GreenDeal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(result))
	}

	// Deterministic: same input → same output.
	result2, _ := ComputeExternalNullifier("TestDomain|Climate|GreenDeal")
	if hex.EncodeToString(result) != hex.EncodeToString(result2) {
		t.Fatal("ComputeExternalNullifier must be deterministic")
	}
}

func TestComputeExternalNullifierDifferentContexts(t *testing.T) {
	a, _ := ComputeExternalNullifier("Domain1|Issue1|Suggestion1")
	b, _ := ComputeExternalNullifier("Domain1|Issue1|Suggestion2")
	c, _ := ComputeExternalNullifier("Domain2|Issue1|Suggestion1")

	ah := hex.EncodeToString(a)
	bh := hex.EncodeToString(b)
	ch := hex.EncodeToString(c)

	if ah == bh || ah == ch || bh == ch {
		t.Fatal("different contexts must produce different external nullifiers")
	}
}

func TestComputeExternalNullifierLongContext(t *testing.T) {
	// Context string longer than 32 bytes.
	long := "VeryLongDomainName|VeryLongIssueName|VeryLongSuggestionNameThatExceeds32Bytes"
	result, err := ComputeExternalNullifier(long)
	if err != nil {
		t.Fatalf("unexpected error for long context: %v", err)
	}
	if len(result) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(result))
	}
}

func TestComputeExternalNullifierFieldElement(t *testing.T) {
	result, _ := ComputeExternalNullifier("test-context")
	n := new(big.Int).SetBytes(result)
	modulus := ecc.BN254.ScalarField()
	if n.Cmp(modulus) >= 0 {
		t.Fatal("result must be < BN254 scalar field modulus")
	}
}

// ---------- Verifying Key Storage Tests ----------

func TestStoreAndRetrieveVerifyingKey(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Initially no VK stored.
	_, found := k.GetVerifyingKey(ctx)
	if found {
		t.Fatal("VK should not exist initially")
	}

	// Setup and store VK.
	keys := getTestZKPKeys(t)
	vkBytes, err := SerializeVerifyingKey(keys.VerifyingKey)
	if err != nil {
		t.Fatalf("SerializeVerifyingKey failed: %v", err)
	}
	k.SetVerifyingKey(ctx, vkBytes)

	// Retrieve and verify.
	retrieved, found := k.GetVerifyingKey(ctx)
	if !found {
		t.Fatal("VK should exist after storing")
	}
	if len(retrieved) != len(vkBytes) {
		t.Fatalf("retrieved VK length mismatch: %d vs %d", len(retrieved), len(vkBytes))
	}

	// Verify the retrieved VK can be deserialized.
	_, err = DeserializeVerifyingKey(retrieved)
	if err != nil {
		t.Fatalf("DeserializeVerifyingKey failed: %v", err)
	}
}

func TestEnsureVerifyingKeyIdempotent(t *testing.T) {
	k, ctx := setupKeeper(t)

	// First call initializes.
	vk1, err := k.EnsureVerifyingKey(ctx)
	if err != nil {
		t.Fatalf("first EnsureVerifyingKey failed: %v", err)
	}
	if len(vk1) == 0 {
		t.Fatal("VK bytes should not be empty")
	}

	// Second call returns stored VK (no re-setup).
	vk2, err := k.EnsureVerifyingKey(ctx)
	if err != nil {
		t.Fatalf("second EnsureVerifyingKey failed: %v", err)
	}
	if hex.EncodeToString(vk1) != hex.EncodeToString(vk2) {
		t.Fatal("EnsureVerifyingKey must return the same VK on subsequent calls")
	}
}

// ---------- Test Helpers ----------

// setupDomainWithZKPIdentity creates a domain with members, registers identity
// commitments, and returns the identity secrets for proof generation.
func setupDomainWithZKPIdentity(t *testing.T, k Keeper, ctx sdk.Context, domainName string, numMembers int) [][]byte {
	t.Helper()
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, domainName, admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500_000)))

	secrets := make([][]byte, numMembers)
	for i := 0; i < numMembers; i++ {
		memberAddr := sdk.AccAddress("member" + string(rune('A'+i))).String()
		k.AddMember(ctx, domainName, memberAddr, admin)

		secret := big.NewInt(int64(i + 200)).Bytes()
		commitment, err := ComputeCommitment(secret)
		if err != nil {
			t.Fatalf("ComputeCommitment failed for member %d: %v", i, err)
		}
		commitHex := hex.EncodeToString(commitment)
		if err := k.RegisterIdentityCommitment(ctx, domainName, memberAddr, commitHex); err != nil {
			t.Fatalf("RegisterIdentityCommitment failed for member %d: %v", i, err)
		}
		secrets[i] = secret
	}
	return secrets
}
