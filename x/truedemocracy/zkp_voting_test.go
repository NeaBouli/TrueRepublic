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

	reward, err := k.RateProposalWithZKP(ctx, "ZKPDomain", "Climate", "GreenDeal", 3, proofHex, nullifierHex, "")
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
	_, err := k.RateProposalWithZKP(ctx, "ZKPDomain", "Climate", "GreenDeal", 5, proofHex, nullifierHex, "")
	if err != nil {
		t.Fatalf("first vote should succeed: %v", err)
	}

	// Second vote with same nullifier rejected.
	_, err = k.RateProposalWithZKP(ctx, "ZKPDomain", "Climate", "GreenDeal", -3, proofHex, nullifierHex, "")
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

	_, err := k.RateProposalWithZKP(ctx, "ZKPDomain", "Climate", "GreenDeal", 5, proof1, null1, "")
	if err != nil {
		t.Fatalf("first suggestion rating failed: %v", err)
	}

	_, err = k.RateProposalWithZKP(ctx, "ZKPDomain", "Climate", "BlueDeal", -2, proof2, null2, "")
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

	_, err := k.RateProposalWithZKP(ctx, "ZKPDomain", "Climate", "GreenDeal", 3, badProofHex, nullifierHex, "")
	if err == nil {
		t.Fatal("expected error for wrong proof")
	}
}

func TestRateProposalWithZKPEmptyMerkleRoot(t *testing.T) {
	k, ctx := setupKeeper(t)
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "EmptyDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500_000)))

	_, err := k.RateProposalWithZKP(ctx, "EmptyDomain", "Issue", "Sugg", 3, "aabb", "aabb", "")
	if err == nil {
		t.Fatal("expected error for domain with no identity commitments")
	}
}

func TestRateProposalWithZKPInvalidRating(t *testing.T) {
	k, ctx := setupKeeper(t)

	_, err := k.RateProposalWithZKP(ctx, "D", "I", "S", 6, "aa", "bb", "")
	if err == nil {
		t.Fatal("expected error for rating > 5")
	}
	_, err = k.RateProposalWithZKP(ctx, "D", "I", "S", -6, "aa", "bb", "")
	if err == nil {
		t.Fatal("expected error for rating < -5")
	}
}

func TestRateProposalWithZKPUnknownDomain(t *testing.T) {
	k, ctx := setupKeeper(t)

	_, err := k.RateProposalWithZKP(ctx, "NoDomain", "I", "S", 3, "aa", "bb", "")
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
	reward, err := k.RateProposalWithZKP(ctx, "ZKPDomain", "Climate", "GreenDeal", 4, proofHex, nullifierHex, "")
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

// ---------- MsgRateWithProof ValidateBasic Tests ----------

func TestMsgRateWithProofValidateBasic(t *testing.T) {
	validProof := hex.EncodeToString(make([]byte, 128))
	validNullifier := hex.EncodeToString(make([]byte, 32))

	t.Run("valid message", func(t *testing.T) {
		msg := MsgRateWithProof{
			Sender:         sdk.AccAddress("sender1"),
			DomainName:     "TestDomain",
			IssueName:      "Climate",
			SuggestionName: "GreenDeal",
			Rating:         3,
			Proof:          validProof,
			NullifierHash:  validNullifier,
		}
		if err := msg.ValidateBasic(); err != nil {
			t.Fatalf("expected valid, got: %v", err)
		}
	})

	t.Run("empty domain rejected", func(t *testing.T) {
		msg := MsgRateWithProof{
			Sender:         sdk.AccAddress("sender1"),
			DomainName:     "",
			IssueName:      "Climate",
			SuggestionName: "GreenDeal",
			Rating:         3,
			Proof:          validProof,
			NullifierHash:  validNullifier,
		}
		if err := msg.ValidateBasic(); err == nil {
			t.Fatal("expected error for empty domain")
		}
	})

	t.Run("empty proof rejected", func(t *testing.T) {
		msg := MsgRateWithProof{
			Sender:         sdk.AccAddress("sender1"),
			DomainName:     "TestDomain",
			IssueName:      "Climate",
			SuggestionName: "GreenDeal",
			Rating:         3,
			Proof:          "",
			NullifierHash:  validNullifier,
		}
		if err := msg.ValidateBasic(); err == nil {
			t.Fatal("expected error for empty proof")
		}
	})

	t.Run("invalid nullifier length rejected", func(t *testing.T) {
		msg := MsgRateWithProof{
			Sender:         sdk.AccAddress("sender1"),
			DomainName:     "TestDomain",
			IssueName:      "Climate",
			SuggestionName: "GreenDeal",
			Rating:         3,
			Proof:          validProof,
			NullifierHash:  "aabb", // only 4 hex chars, need 64
		}
		if err := msg.ValidateBasic(); err == nil {
			t.Fatal("expected error for invalid nullifier length")
		}
	})

	t.Run("invalid hex in proof rejected", func(t *testing.T) {
		msg := MsgRateWithProof{
			Sender:         sdk.AccAddress("sender1"),
			DomainName:     "TestDomain",
			IssueName:      "Climate",
			SuggestionName: "GreenDeal",
			Rating:         3,
			Proof:          "not-valid-hex!!",
			NullifierHash:  validNullifier,
		}
		if err := msg.ValidateBasic(); err == nil {
			t.Fatal("expected error for invalid hex in proof")
		}
	})

	t.Run("rating out of range rejected", func(t *testing.T) {
		msg := MsgRateWithProof{
			Sender:         sdk.AccAddress("sender1"),
			DomainName:     "TestDomain",
			IssueName:      "Climate",
			SuggestionName: "GreenDeal",
			Rating:         6,
			Proof:          validProof,
			NullifierHash:  validNullifier,
		}
		if err := msg.ValidateBasic(); err == nil {
			t.Fatal("expected error for rating > 5")
		}

		msg.Rating = -6
		if err := msg.ValidateBasic(); err == nil {
			t.Fatal("expected error for rating < -5")
		}
	})
}

// ---------- MsgServer RateWithProof Tests ----------

func TestMsgServerRateWithProof(t *testing.T) {
	k, ctx := setupKeeper(t)
	srv := NewMsgServer(k)

	secrets := setupDomainWithZKPIdentity(t, k, ctx, "ZKPDomain", 3)
	addProposal(t, k, ctx, "ZKPDomain", "Climate", "GreenDeal")

	proofHex, nullifierHex := generateZKPRating(t, k, ctx, "ZKPDomain", secrets, 1, "Climate", "GreenDeal")

	msg := &MsgRateWithProof{
		Sender:         sdk.AccAddress("sender1"),
		DomainName:     "ZKPDomain",
		IssueName:      "Climate",
		SuggestionName: "GreenDeal",
		Rating:         4,
		Proof:          proofHex,
		NullifierHash:  nullifierHex,
	}

	_, err := srv.RateWithProof(ctx, msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify rating stored.
	domain, _ := k.GetDomain(ctx, "ZKPDomain")
	ratings := domain.Issues[0].Suggestions[0].Ratings
	if len(ratings) != 1 {
		t.Fatalf("expected 1 rating, got %d", len(ratings))
	}
	if ratings[0].Value != 4 {
		t.Fatalf("expected rating value 4, got %d", ratings[0].Value)
	}
}

func TestMsgServerRateWithProofDoubleVote(t *testing.T) {
	k, ctx := setupKeeper(t)
	srv := NewMsgServer(k)

	secrets := setupDomainWithZKPIdentity(t, k, ctx, "ZKPDomain", 3)
	addProposal(t, k, ctx, "ZKPDomain", "Climate", "GreenDeal")

	proofHex, nullifierHex := generateZKPRating(t, k, ctx, "ZKPDomain", secrets, 0, "Climate", "GreenDeal")

	msg := &MsgRateWithProof{
		Sender:         sdk.AccAddress("sender1"),
		DomainName:     "ZKPDomain",
		IssueName:      "Climate",
		SuggestionName: "GreenDeal",
		Rating:         3,
		Proof:          proofHex,
		NullifierHash:  nullifierHex,
	}

	// First vote succeeds.
	_, err := srv.RateWithProof(ctx, msg)
	if err != nil {
		t.Fatalf("first vote should succeed: %v", err)
	}

	// Second vote with same nullifier rejected.
	msg.Rating = -2
	_, err = srv.RateWithProof(ctx, msg)
	if err == nil {
		t.Fatal("expected error for double vote")
	}
}

// ---------- E2E Tests ----------

func TestE2EZKPRatingFlow(t *testing.T) {
	k, ctx := setupKeeper(t)
	srv := NewMsgServer(k)

	// 1. Create domain.
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "E2EDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500_000)))

	// 2. Add member.
	memberAddr := sdk.AccAddress("memberA").String()
	k.AddMember(ctx, "E2EDomain", memberAddr, admin)

	// 3. Register identity commitment.
	secret := big.NewInt(777).Bytes()
	commitment, err := ComputeCommitment(secret)
	if err != nil {
		t.Fatalf("ComputeCommitment failed: %v", err)
	}
	commitHex := hex.EncodeToString(commitment)

	identityMsg := &MsgRegisterIdentity{
		Sender:     sdk.AccAddress("memberA"),
		DomainName: "E2EDomain",
		Commitment: commitHex,
	}
	_, err = srv.RegisterIdentity(ctx, identityMsg)
	if err != nil {
		t.Fatalf("RegisterIdentity failed: %v", err)
	}

	// 4. Submit a proposal.
	fee := sdk.NewCoins(sdk.NewInt64Coin("pnyx", 10_000_000))
	if err := k.SubmitProposal(ctx, "E2EDomain", "Energy", "Solar", admin.String(), fee, ""); err != nil {
		t.Fatalf("SubmitProposal failed: %v", err)
	}

	// 5. Generate ZKP proof.
	keys := getTestZKPKeys(t)
	domain, _ := k.GetDomain(ctx, "E2EDomain")
	commitments := make([][]byte, len(domain.IdentityCommits))
	for i, h := range domain.IdentityCommits {
		b, _ := hex.DecodeString(h)
		commitments[i] = b
	}
	tree := NewMerkleTree(MerkleTreeDepth)
	if err := tree.BuildFromLeaves(commitments); err != nil {
		t.Fatalf("BuildFromLeaves failed: %v", err)
	}
	siblings, pathIndices, err := tree.GenerateProof(0)
	if err != nil {
		t.Fatalf("GenerateProof failed: %v", err)
	}
	extNullifier, _ := ComputeExternalNullifier("E2EDomain|Energy|Solar")
	proofBytes, nullifierHash, err := GenerateMembershipProof(keys, secret, tree.Root, siblings, pathIndices, extNullifier)
	if err != nil {
		t.Fatalf("GenerateMembershipProof failed: %v", err)
	}

	// 6. Submit MsgRateWithProof.
	rateMsg := &MsgRateWithProof{
		Sender:         sdk.AccAddress("memberA"),
		DomainName:     "E2EDomain",
		IssueName:      "Energy",
		SuggestionName: "Solar",
		Rating:         5,
		Proof:          hex.EncodeToString(proofBytes),
		NullifierHash:  hex.EncodeToString(nullifierHash),
	}
	_, err = srv.RateWithProof(ctx, rateMsg)
	if err != nil {
		t.Fatalf("RateWithProof failed: %v", err)
	}

	// 7. Verify rating stored + nullifier used.
	domain, _ = k.GetDomain(ctx, "E2EDomain")
	ratings := domain.Issues[0].Suggestions[0].Ratings
	if len(ratings) != 1 {
		t.Fatalf("expected 1 rating, got %d", len(ratings))
	}
	if ratings[0].Value != 5 {
		t.Fatalf("expected rating 5, got %d", ratings[0].Value)
	}
	if !k.IsNullifierUsed(ctx, "E2EDomain", hex.EncodeToString(nullifierHash)) {
		t.Fatal("nullifier should be marked as used")
	}
}

// ---------- Merkle Root History Tests ----------

func TestMerkleRootHistory(t *testing.T) {
	k, ctx := setupKeeper(t)
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "HistDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500_000)))

	memberAddr := sdk.AccAddress("memberA").String()
	k.AddMember(ctx, "HistDomain", memberAddr, admin)

	// Register 3 commitments — each should push the previous root into history.
	var prevRoots []string
	for i := 0; i < 3; i++ {
		secret := big.NewInt(int64(i + 100)).Bytes()
		commitment, err := ComputeCommitment(secret)
		if err != nil {
			t.Fatalf("ComputeCommitment failed: %v", err)
		}
		domain, _ := k.GetDomain(ctx, "HistDomain")
		if domain.MerkleRoot != "" {
			prevRoots = append(prevRoots, domain.MerkleRoot)
		}
		if err := k.RegisterIdentityCommitment(ctx, "HistDomain", memberAddr, hex.EncodeToString(commitment)); err != nil {
			t.Fatalf("RegisterIdentityCommitment failed at %d: %v", i, err)
		}
	}

	domain, _ := k.GetDomain(ctx, "HistDomain")
	// After 3 registrations, history should contain the first 2 roots.
	if len(domain.MerkleRootHistory) != 2 {
		t.Fatalf("expected 2 roots in history, got %d", len(domain.MerkleRootHistory))
	}
	for i, expected := range prevRoots {
		if domain.MerkleRootHistory[i] != expected {
			t.Fatalf("history[%d] mismatch: expected %s, got %s", i, expected, domain.MerkleRootHistory[i])
		}
	}
}

func TestMerkleRootHistoryCap(t *testing.T) {
	k, ctx := setupKeeper(t)
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "CapDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500_000)))

	memberAddr := sdk.AccAddress("memberA").String()
	k.AddMember(ctx, "CapDomain", memberAddr, admin)

	// Register MerkleRootHistorySize + 5 commitments.
	total := MerkleRootHistorySize + 5
	for i := 0; i < total; i++ {
		secret := big.NewInt(int64(i + 300)).Bytes()
		commitment, err := ComputeCommitment(secret)
		if err != nil {
			t.Fatalf("ComputeCommitment failed: %v", err)
		}
		if err := k.RegisterIdentityCommitment(ctx, "CapDomain", memberAddr, hex.EncodeToString(commitment)); err != nil {
			t.Fatalf("RegisterIdentityCommitment failed at %d: %v", i, err)
		}
	}

	domain, _ := k.GetDomain(ctx, "CapDomain")
	if len(domain.MerkleRootHistory) != MerkleRootHistorySize {
		t.Fatalf("expected history capped at %d, got %d", MerkleRootHistorySize, len(domain.MerkleRootHistory))
	}
}

func TestRateWithHistoricalRoot(t *testing.T) {
	k, ctx := setupKeeper(t)
	secrets := setupDomainWithZKPIdentity(t, k, ctx, "HistRateDomain", 3)
	addProposal(t, k, ctx, "HistRateDomain", "Climate", "GreenDeal")

	// Capture current root before adding more commitments.
	domain, _ := k.GetDomain(ctx, "HistRateDomain")
	rootBeforeNewCommit := domain.MerkleRoot

	// Generate proof against the current root.
	proofHex, nullifierHex := generateZKPRating(t, k, ctx, "HistRateDomain", secrets, 1, "Climate", "GreenDeal")

	// Now register a NEW commitment — this changes the current root.
	newMemberAddr := sdk.AccAddress("memberX").String()
	k.AddMember(ctx, "HistRateDomain", newMemberAddr, sdk.AccAddress("admin1"))
	newSecret := big.NewInt(999).Bytes()
	newCommitment, _ := ComputeCommitment(newSecret)
	if err := k.RegisterIdentityCommitment(ctx, "HistRateDomain", newMemberAddr, hex.EncodeToString(newCommitment)); err != nil {
		t.Fatalf("RegisterIdentityCommitment failed: %v", err)
	}

	// Verify root has changed.
	domain, _ = k.GetDomain(ctx, "HistRateDomain")
	if domain.MerkleRoot == rootBeforeNewCommit {
		t.Fatal("root should have changed after new commitment")
	}

	// Rate with the OLD root (now in history) — should succeed.
	_, err := k.RateProposalWithZKP(ctx, "HistRateDomain", "Climate", "GreenDeal", 3, proofHex, nullifierHex, rootBeforeNewCommit)
	if err != nil {
		t.Fatalf("rating with historical root should succeed: %v", err)
	}
}

func TestRateWithExpiredRoot(t *testing.T) {
	k, ctx := setupKeeper(t)
	secrets := setupDomainWithZKPIdentity(t, k, ctx, "ExpiredDomain", 2)
	addProposal(t, k, ctx, "ExpiredDomain", "Climate", "GreenDeal")

	// Capture current root.
	domain, _ := k.GetDomain(ctx, "ExpiredDomain")
	oldRoot := domain.MerkleRoot

	// Generate proof against the current root.
	proofHex, nullifierHex := generateZKPRating(t, k, ctx, "ExpiredDomain", secrets, 0, "Climate", "GreenDeal")

	// Register MerkleRootHistorySize + 2 more commitments to push old root out of history.
	memberAddr := sdk.AccAddress("memberA").String()
	for i := 0; i < MerkleRootHistorySize+2; i++ {
		secret := big.NewInt(int64(i + 700)).Bytes()
		commitment, _ := ComputeCommitment(secret)
		if err := k.RegisterIdentityCommitment(ctx, "ExpiredDomain", memberAddr, hex.EncodeToString(commitment)); err != nil {
			t.Fatalf("RegisterIdentityCommitment failed at %d: %v", i, err)
		}
	}

	// Verify old root is NOT in history anymore.
	domain, _ = k.GetDomain(ctx, "ExpiredDomain")
	if domain.MerkleRoot == oldRoot {
		t.Fatal("root should have changed")
	}
	for _, h := range domain.MerkleRootHistory {
		if h == oldRoot {
			t.Fatal("old root should have been evicted from history")
		}
	}

	// Rate with expired root — should fail.
	_, err := k.RateProposalWithZKP(ctx, "ExpiredDomain", "Climate", "GreenDeal", 3, proofHex, nullifierHex, oldRoot)
	if err == nil {
		t.Fatal("expected error for expired root")
	}
}

func TestRateWithEmptyMerkleRootUsesCurrentRoot(t *testing.T) {
	k, ctx := setupKeeper(t)
	secrets := setupDomainWithZKPIdentity(t, k, ctx, "EmptyRootDomain", 3)
	addProposal(t, k, ctx, "EmptyRootDomain", "Climate", "GreenDeal")

	proofHex, nullifierHex := generateZKPRating(t, k, ctx, "EmptyRootDomain", secrets, 0, "Climate", "GreenDeal")

	// Empty merkleRootHex → uses current domain root (existing behavior).
	_, err := k.RateProposalWithZKP(ctx, "EmptyRootDomain", "Climate", "GreenDeal", 4, proofHex, nullifierHex, "")
	if err != nil {
		t.Fatalf("rating with empty merkle root should use current root: %v", err)
	}
}

func TestBigPurgeClearsRootHistory(t *testing.T) {
	k, ctx := setupKeeper(t)
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "PurgeHistDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500_000)))

	memberAddr := sdk.AccAddress("memberA").String()
	k.AddMember(ctx, "PurgeHistDomain", memberAddr, admin)

	// Register a few commitments to build up history.
	for i := 0; i < 5; i++ {
		secret := big.NewInt(int64(i + 400)).Bytes()
		commitment, _ := ComputeCommitment(secret)
		if err := k.RegisterIdentityCommitment(ctx, "PurgeHistDomain", memberAddr, hex.EncodeToString(commitment)); err != nil {
			t.Fatalf("RegisterIdentityCommitment failed at %d: %v", i, err)
		}
	}

	domain, _ := k.GetDomain(ctx, "PurgeHistDomain")
	if len(domain.MerkleRootHistory) == 0 {
		t.Fatal("history should be non-empty before purge")
	}
	if domain.MerkleRoot == "" {
		t.Fatal("root should be non-empty before purge")
	}

	// Execute Big Purge.
	k.executeBigPurge(ctx, "PurgeHistDomain")

	domain, _ = k.GetDomain(ctx, "PurgeHistDomain")
	if domain.MerkleRoot != "" {
		t.Fatal("purge should clear Merkle root")
	}
	if len(domain.MerkleRootHistory) != 0 {
		t.Fatalf("purge should clear root history, got %d entries", len(domain.MerkleRootHistory))
	}
	if len(domain.IdentityCommits) != 0 {
		t.Fatal("purge should clear identity commitments")
	}
}

func TestE2EBigPurgeClearsAndReallowsVoting(t *testing.T) {
	k, ctx := setupKeeper(t)
	secrets := setupDomainWithZKPIdentity(t, k, ctx, "PurgeDomain", 2)
	addProposal(t, k, ctx, "PurgeDomain", "Climate", "GreenDeal")

	// 1. Rate with ZKP — succeeds.
	proof1, null1 := generateZKPRating(t, k, ctx, "PurgeDomain", secrets, 0, "Climate", "GreenDeal")
	_, err := k.RateProposalWithZKP(ctx, "PurgeDomain", "Climate", "GreenDeal", 3, proof1, null1, "")
	if err != nil {
		t.Fatalf("first rating should succeed: %v", err)
	}

	// 2. Execute Big Purge — clears identities, nullifiers, Merkle root.
	k.executeBigPurge(ctx, "PurgeDomain")

	domain, _ := k.GetDomain(ctx, "PurgeDomain")
	if len(domain.IdentityCommits) != 0 {
		t.Fatal("purge should clear identity commitments")
	}
	if domain.MerkleRoot != "" {
		t.Fatal("purge should clear Merkle root")
	}
	if k.IsNullifierUsed(ctx, "PurgeDomain", null1) {
		t.Fatal("purge should clear nullifiers")
	}

	// 3. Re-register identity commitments (fresh secrets).
	newSecrets := make([][]byte, 2)
	for i := 0; i < 2; i++ {
		memberAddr := sdk.AccAddress("member" + string(rune('A'+i))).String()
		newSecret := big.NewInt(int64(i + 500)).Bytes()
		commitment, err := ComputeCommitment(newSecret)
		if err != nil {
			t.Fatalf("ComputeCommitment failed: %v", err)
		}
		commitHex := hex.EncodeToString(commitment)
		if err := k.RegisterIdentityCommitment(ctx, "PurgeDomain", memberAddr, commitHex); err != nil {
			t.Fatalf("RegisterIdentityCommitment failed: %v", err)
		}
		newSecrets[i] = newSecret
	}

	// 4. Rate again with new proof — succeeds.
	proof2, null2 := generateZKPRating(t, k, ctx, "PurgeDomain", newSecrets, 0, "Climate", "GreenDeal")
	_, err = k.RateProposalWithZKP(ctx, "PurgeDomain", "Climate", "GreenDeal", -2, proof2, null2, "")
	if err != nil {
		t.Fatalf("rating after purge + re-register should succeed: %v", err)
	}

	// Verify: two ratings total (one before purge, one after).
	domain, _ = k.GetDomain(ctx, "PurgeDomain")
	ratings := domain.Issues[0].Suggestions[0].Ratings
	if len(ratings) != 2 {
		t.Fatalf("expected 2 ratings, got %d", len(ratings))
	}
}
