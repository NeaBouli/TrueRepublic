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

// generateZKPRating generates a Groth16 proof and nullifier for rating a suggestion.
// Returns proofHex and nullifierHashHex ready for RateProposalWithZKP.
func generateZKPRating(t *testing.T, k Keeper, ctx sdk.Context, domainName string, secrets [][]byte, memberIndex int, issueName, suggestionName string) (string, string) {
	t.Helper()
	keys := getTestZKPKeys(t)

	// Rebuild tree from domain's commitments.
	domain, _ := k.GetDomain(ctx, domainName)
	commitments := make([][]byte, len(domain.IdentityCommits))
	for i, h := range domain.IdentityCommits {
		b, _ := hex.DecodeString(h)
		commitments[i] = b
	}
	tree := NewMerkleTree(MerkleTreeDepth)
	if err := tree.BuildFromLeaves(commitments); err != nil {
		t.Fatalf("BuildFromLeaves failed: %v", err)
	}

	siblings, pathIndices, err := tree.GenerateProof(memberIndex)
	if err != nil {
		t.Fatalf("GenerateProof failed: %v", err)
	}

	extNullifier, err := ComputeExternalNullifier(domainName + "|" + issueName + "|" + suggestionName)
	if err != nil {
		t.Fatalf("ComputeExternalNullifier failed: %v", err)
	}

	proofBytes, nullifierHash, err := GenerateMembershipProof(
		keys,
		secrets[memberIndex],
		tree.Root,
		siblings,
		pathIndices,
		extNullifier,
	)
	if err != nil {
		t.Fatalf("GenerateMembershipProof failed: %v", err)
	}

	return hex.EncodeToString(proofBytes), hex.EncodeToString(nullifierHash)
}

// addProposal creates an issue with a suggestion in the domain.
func addProposal(t *testing.T, k Keeper, ctx sdk.Context, domainName, issueName, suggestionName string) {
	t.Helper()
	admin := sdk.AccAddress("admin1")
	fee := sdk.NewCoins(sdk.NewInt64Coin("pnyx", 10_000_000))
	if err := k.SubmitProposal(ctx, domainName, issueName, suggestionName, admin.String(), fee, ""); err != nil {
		t.Fatalf("SubmitProposal failed: %v", err)
	}
}

// ---------- RateProposalWithZKP Keeper Tests ----------

func TestRateProposalWithZKP(t *testing.T) {
	k, ctx := setupKeeper(t)
	secrets := setupDomainWithZKPIdentity(t, k, ctx, "ZKPDomain", 3)
	addProposal(t, k, ctx, "ZKPDomain", "Climate", "GreenDeal")

	proofHex, nullifierHex := generateZKPRating(t, k, ctx, "ZKPDomain", secrets, 1, "Climate", "GreenDeal")

	reward, err := k.RateProposalWithZKP(ctx, "ZKPDomain", "Climate", "GreenDeal", 3, proofHex, nullifierHex)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reward.IsZero() {
		t.Fatal("expected non-zero reward")
	}

	// Verify rating stored.
	domain, _ := k.GetDomain(ctx, "ZKPDomain")
	ratings := domain.Issues[0].Suggestions[0].Ratings
	if len(ratings) != 1 {
		t.Fatalf("expected 1 rating, got %d", len(ratings))
	}
	if ratings[0].Value != 3 {
		t.Fatalf("expected rating value 3, got %d", ratings[0].Value)
	}
	if ratings[0].NullifierHex != nullifierHex {
		t.Fatal("rating should store nullifier hex")
	}
	if ratings[0].DomainPubKeyHex != "" {
		t.Fatal("ZKP rating should have empty DomainPubKeyHex")
	}
}

func TestRateProposalWithZKPDoubleVoteBlocked(t *testing.T) {
	k, ctx := setupKeeper(t)
	secrets := setupDomainWithZKPIdentity(t, k, ctx, "ZKPDomain", 3)
	addProposal(t, k, ctx, "ZKPDomain", "Climate", "GreenDeal")

	proofHex, nullifierHex := generateZKPRating(t, k, ctx, "ZKPDomain", secrets, 0, "Climate", "GreenDeal")

	// First vote succeeds.
	_, err := k.RateProposalWithZKP(ctx, "ZKPDomain", "Climate", "GreenDeal", 5, proofHex, nullifierHex)
	if err != nil {
		t.Fatalf("first vote should succeed: %v", err)
	}

	// Second vote with same nullifier rejected.
	_, err = k.RateProposalWithZKP(ctx, "ZKPDomain", "Climate", "GreenDeal", -3, proofHex, nullifierHex)
	if err == nil {
		t.Fatal("expected error for double vote")
	}
}

func TestRateProposalWithZKPDifferentSuggestions(t *testing.T) {
	k, ctx := setupKeeper(t)
	secrets := setupDomainWithZKPIdentity(t, k, ctx, "ZKPDomain", 3)
	addProposal(t, k, ctx, "ZKPDomain", "Climate", "GreenDeal")
	addProposal(t, k, ctx, "ZKPDomain", "Climate", "BlueDeal")

	// Same member rates two different suggestions — different nullifiers.
	proof1, null1 := generateZKPRating(t, k, ctx, "ZKPDomain", secrets, 0, "Climate", "GreenDeal")
	proof2, null2 := generateZKPRating(t, k, ctx, "ZKPDomain", secrets, 0, "Climate", "BlueDeal")

	if null1 == null2 {
		t.Fatal("different suggestions should produce different nullifiers")
	}

	_, err := k.RateProposalWithZKP(ctx, "ZKPDomain", "Climate", "GreenDeal", 5, proof1, null1)
	if err != nil {
		t.Fatalf("first suggestion rating failed: %v", err)
	}

	_, err = k.RateProposalWithZKP(ctx, "ZKPDomain", "Climate", "BlueDeal", -2, proof2, null2)
	if err != nil {
		t.Fatalf("second suggestion rating failed: %v", err)
	}
}

func TestRateProposalWithZKPWrongProof(t *testing.T) {
	k, ctx := setupKeeper(t)
	secrets := setupDomainWithZKPIdentity(t, k, ctx, "ZKPDomain", 3)
	addProposal(t, k, ctx, "ZKPDomain", "Climate", "GreenDeal")

	_, nullifierHex := generateZKPRating(t, k, ctx, "ZKPDomain", secrets, 0, "Climate", "GreenDeal")

	// Use garbage proof bytes.
	badProofHex := hex.EncodeToString(make([]byte, 256))

	_, err := k.RateProposalWithZKP(ctx, "ZKPDomain", "Climate", "GreenDeal", 3, badProofHex, nullifierHex)
	if err == nil {
		t.Fatal("expected error for wrong proof")
	}
}

func TestRateProposalWithZKPEmptyMerkleRoot(t *testing.T) {
	k, ctx := setupKeeper(t)
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "EmptyDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500_000)))

	_, err := k.RateProposalWithZKP(ctx, "EmptyDomain", "Issue", "Sugg", 3, "aabb", "aabb")
	if err == nil {
		t.Fatal("expected error for domain with no identity commitments")
	}
}

func TestRateProposalWithZKPInvalidRating(t *testing.T) {
	k, ctx := setupKeeper(t)

	_, err := k.RateProposalWithZKP(ctx, "D", "I", "S", 6, "aa", "bb")
	if err == nil {
		t.Fatal("expected error for rating > 5")
	}
	_, err = k.RateProposalWithZKP(ctx, "D", "I", "S", -6, "aa", "bb")
	if err == nil {
		t.Fatal("expected error for rating < -5")
	}
}

func TestRateProposalWithZKPUnknownDomain(t *testing.T) {
	k, ctx := setupKeeper(t)

	_, err := k.RateProposalWithZKP(ctx, "NoDomain", "I", "S", 3, "aa", "bb")
	if err == nil {
		t.Fatal("expected error for unknown domain")
	}
}

func TestRateProposalWithZKPRewardsDistributed(t *testing.T) {
	k, ctx := setupKeeper(t)
	secrets := setupDomainWithZKPIdentity(t, k, ctx, "ZKPDomain", 3)
	addProposal(t, k, ctx, "ZKPDomain", "Climate", "GreenDeal")

	domainBefore, _ := k.GetDomain(ctx, "ZKPDomain")
	treasuryBefore := domainBefore.Treasury.AmountOf("pnyx").Int64()

	proofHex, nullifierHex := generateZKPRating(t, k, ctx, "ZKPDomain", secrets, 2, "Climate", "GreenDeal")
	reward, err := k.RateProposalWithZKP(ctx, "ZKPDomain", "Climate", "GreenDeal", 4, proofHex, nullifierHex)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	domainAfter, _ := k.GetDomain(ctx, "ZKPDomain")
	treasuryAfter := domainAfter.Treasury.AmountOf("pnyx").Int64()

	if treasuryAfter >= treasuryBefore {
		t.Fatal("treasury should decrease after reward payout")
	}
	if reward.AmountOf("pnyx").Int64() != treasuryBefore-treasuryAfter {
		t.Fatal("reward amount should match treasury decrease")
	}
}
