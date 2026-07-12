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

	if _, err := k.EnsureVerifyingKey(ctx); err == nil {
		t.Fatal("missing consensus-configured VK must fail closed")
	}
	keys := getTestZKPKeys(t)
	vk1, err := SerializeVerifyingKey(keys.VerifyingKey)
	if err != nil {
		t.Fatalf("SerializeVerifyingKey failed: %v", err)
	}
	k.SetVerifyingKey(ctx, vk1)
	vk1, err = k.EnsureVerifyingKey(ctx)
	if err != nil {
		t.Fatalf("configured EnsureVerifyingKey failed: %v", err)
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

func setTestVerifyingKey(t *testing.T, k Keeper, ctx sdk.Context) []byte {
	t.Helper()
	keys := getTestZKPKeys(t)
	vkBytes, err := SerializeVerifyingKey(keys.VerifyingKey)
	if err != nil {
		t.Fatalf("SerializeVerifyingKey failed: %v", err)
	}
	k.SetVerifyingKey(ctx, vkBytes)
	return vkBytes
}

// setupDomainWithZKPIdentity creates a domain with members, registers identity
// commitments, and returns the identity secrets for proof generation.
func setupDomainWithZKPIdentity(t *testing.T, k Keeper, ctx sdk.Context, domainName string, numMembers int) [][]byte {
	t.Helper()
	setTestVerifyingKey(t, k, ctx)
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, domainName, admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 500_000)))

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
func generateZKPRating(t *testing.T, k Keeper, ctx sdk.Context, domainName string, secrets [][]byte, memberIndex int, issueName, suggestionName string, rating int) (string, string) {
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

	extNullifier := ComputeVoteNullifierScope(ctx.ChainID(), domainName, issueName, suggestionName)
	signalHash := ComputeVoteSignal(ctx.ChainID(), domainName, issueName, suggestionName, rating)

	proofBytes, nullifierHash, err := GenerateMembershipProofForSignal(
		keys,
		secrets[memberIndex],
		tree.Root,
		siblings,
		pathIndices,
		extNullifier,
		signalHash,
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
	fee := sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 10_000_000))
	if err := k.SubmitProposal(ctx, domainName, issueName, suggestionName, admin.String(), fee, ""); err != nil {
		t.Fatalf("SubmitProposal failed: %v", err)
	}
}

// ---------- RateProposalWithZKP Keeper Tests ----------

func TestRateProposalWithZKP(t *testing.T) {
	k, ctx := setupKeeper(t)
	secrets := setupDomainWithZKPIdentity(t, k, ctx, "ZKPDomain", 3)
	addProposal(t, k, ctx, "ZKPDomain", "Climate", "GreenDeal")

	proofHex, nullifierHex := generateZKPRating(t, k, ctx, "ZKPDomain", secrets, 1, "Climate", "GreenDeal", 3)

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

func TestRateProposalWithZKPProofBindsRating(t *testing.T) {
	k, ctx := setupKeeper(t)
	ctx = ctx.WithChainID("truerepublic-test-1")
	secrets := setupDomainWithZKPIdentity(t, k, ctx, "ZKPDomain", 2)
	addProposal(t, k, ctx, "ZKPDomain", "Climate", "GreenDeal")

	proofHex, nullifierHex := generateZKPRating(t, k, ctx, "ZKPDomain", secrets, 0, "Climate", "GreenDeal", 3)
	if _, err := k.RateProposalWithZKP(ctx, "ZKPDomain", "Climate", "GreenDeal", -4, proofHex, nullifierHex, ""); err == nil {
		t.Fatal("proof replay with altered rating must fail")
	}
	domain, _ := k.GetDomain(ctx, "ZKPDomain")
	if got := len(domain.Issues[0].Suggestions[0].Ratings); got != 0 {
		t.Fatalf("altered rating mutated state: %d ratings", got)
	}
	if k.IsNullifierUsed(ctx, "ZKPDomain", nullifierHex) {
		t.Fatal("failed altered-rating proof consumed nullifier")
	}
	if _, err := k.RateProposalWithZKP(ctx, "ZKPDomain", "Climate", "GreenDeal", 3, proofHex, nullifierHex, ""); err != nil {
		t.Fatalf("proof with bound rating should succeed: %v", err)
	}
}

func TestRateProposalWithZKPProofBindsChainID(t *testing.T) {
	k, ctx := setupKeeper(t)
	ctx = ctx.WithChainID("truerepublic-test-1")
	secrets := setupDomainWithZKPIdentity(t, k, ctx, "ZKPDomain", 2)
	addProposal(t, k, ctx, "ZKPDomain", "Climate", "GreenDeal")

	proofHex, nullifierHex := generateZKPRating(t, k, ctx, "ZKPDomain", secrets, 0, "Climate", "GreenDeal", 3)
	otherChain := ctx.WithChainID("truerepublic-test-2")
	if _, err := k.RateProposalWithZKP(otherChain, "ZKPDomain", "Climate", "GreenDeal", 3, proofHex, nullifierHex, ""); err == nil {
		t.Fatal("proof replay on another chain must fail")
	}
	if _, err := k.RateProposalWithZKP(ctx, "ZKPDomain", "Climate", "GreenDeal", 3, proofHex, nullifierHex, ""); err != nil {
		t.Fatalf("proof on its bound chain should succeed: %v", err)
	}
}

func TestVoteNullifierStableAcrossRatings(t *testing.T) {
	k, ctx := setupKeeper(t)
	ctx = ctx.WithChainID("truerepublic-test-1")
	secrets := setupDomainWithZKPIdentity(t, k, ctx, "ZKPDomain", 2)
	addProposal(t, k, ctx, "ZKPDomain", "Climate", "GreenDeal")

	_, nullifier3 := generateZKPRating(t, k, ctx, "ZKPDomain", secrets, 0, "Climate", "GreenDeal", 3)
	_, nullifier4 := generateZKPRating(t, k, ctx, "ZKPDomain", secrets, 0, "Climate", "GreenDeal", 4)
	if nullifier3 != nullifier4 {
		t.Fatal("rating changes must not produce a new one-vote nullifier")
	}
}

func TestRateProposalWithZKPMissingVKFailsWithoutMutation(t *testing.T) {
	k, ctx := setupKeeper(t)
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "NoVKDomain", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 500_000)))
	member := sdk.AccAddress("memberA").String()
	k.AddMember(ctx, "NoVKDomain", member, admin)
	secret := big.NewInt(200).Bytes()
	commitment, _ := ComputeCommitment(secret)
	if err := k.RegisterIdentityCommitment(ctx, "NoVKDomain", member, hex.EncodeToString(commitment)); err != nil {
		t.Fatal(err)
	}
	addProposal(t, k, ctx, "NoVKDomain", "Climate", "GreenDeal")

	domain, _ := k.GetDomain(ctx, "NoVKDomain")
	if _, err := k.RateProposalWithZKP(ctx, "NoVKDomain", "Climate", "GreenDeal", 3, "aa", hex.EncodeToString(make([]byte, 32)), domain.MerkleRoot); err == nil {
		t.Fatal("rating without a configured VK must fail")
	}
	if _, found := k.GetVerifyingKey(ctx); found {
		t.Fatal("transaction execution must not generate or store a VK")
	}
	domain, _ = k.GetDomain(ctx, "NoVKDomain")
	if got := len(domain.Issues[0].Suggestions[0].Ratings); got != 0 {
		t.Fatalf("missing-VK failure mutated ratings: %d", got)
	}
}

func TestRateProposalWithZKPDoubleVoteBlocked(t *testing.T) {
	k, ctx := setupKeeper(t)
	secrets := setupDomainWithZKPIdentity(t, k, ctx, "ZKPDomain", 3)
	addProposal(t, k, ctx, "ZKPDomain", "Climate", "GreenDeal")

	proofHex, nullifierHex := generateZKPRating(t, k, ctx, "ZKPDomain", secrets, 0, "Climate", "GreenDeal", 5)

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
	proof1, null1 := generateZKPRating(t, k, ctx, "ZKPDomain", secrets, 0, "Climate", "GreenDeal", 5)
	proof2, null2 := generateZKPRating(t, k, ctx, "ZKPDomain", secrets, 0, "Climate", "BlueDeal", -2)

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

	_, nullifierHex := generateZKPRating(t, k, ctx, "ZKPDomain", secrets, 0, "Climate", "GreenDeal", 3)

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
	k.CreateDomain(ctx, "EmptyDomain", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 500_000)))

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
	treasuryBefore := domainBefore.Treasury.AmountOf(PNYXDenom).Int64()

	proofHex, nullifierHex := generateZKPRating(t, k, ctx, "ZKPDomain", secrets, 2, "Climate", "GreenDeal", 4)
	reward, err := k.RateProposalWithZKP(ctx, "ZKPDomain", "Climate", "GreenDeal", 4, proofHex, nullifierHex, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	domainAfter, _ := k.GetDomain(ctx, "ZKPDomain")
	treasuryAfter := domainAfter.Treasury.AmountOf(PNYXDenom).Int64()

	if treasuryAfter >= treasuryBefore {
		t.Fatal("treasury should decrease after reward payout")
	}
	if reward.AmountOf(PNYXDenom).Int64() != treasuryBefore-treasuryAfter {
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

	secrets := setupDomainWithZKPIdentity(t, k, ctx, "ZKPDomain", 3)
	addProposal(t, k, ctx, "ZKPDomain", "Climate", "GreenDeal")
	bank := backExistingEscrow(&k, ctx)
	srv := NewMsgServer(k)

	proofHex, nullifierHex := generateZKPRating(t, k, ctx, "ZKPDomain", secrets, 1, "Climate", "GreenDeal", 4)

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
	if got := accountBalance(bank, msg.Sender); got != 0 {
		t.Fatalf("unbound anonymous reward was paid to transaction sender: %d", got)
	}
	if err := k.ValidateEscrowParity(ctx); err != nil {
		t.Fatalf("deferred anonymous reward broke escrow parity: %v", err)
	}
}

func TestMsgServerRateWithProofDoubleVote(t *testing.T) {
	k, ctx := setupKeeper(t)

	secrets := setupDomainWithZKPIdentity(t, k, ctx, "ZKPDomain", 3)
	addProposal(t, k, ctx, "ZKPDomain", "Climate", "GreenDeal")
	backExistingEscrow(&k, ctx)
	srv := NewMsgServer(k)

	proofHex, nullifierHex := generateZKPRating(t, k, ctx, "ZKPDomain", secrets, 0, "Climate", "GreenDeal", 3)

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
	k.CreateDomain(ctx, "E2EDomain", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 500_000)))

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
	fee := sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 10_000_000))
	if err := k.SubmitProposal(ctx, "E2EDomain", "Energy", "Solar", admin.String(), fee, ""); err != nil {
		t.Fatalf("SubmitProposal failed: %v", err)
	}
	backExistingEscrow(&k, ctx)
	srv = NewMsgServer(k)

	// 5. Generate ZKP proof.
	keys := getTestZKPKeys(t)
	setTestVerifyingKey(t, k, ctx)
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
	extNullifier := ComputeVoteNullifierScope(ctx.ChainID(), "E2EDomain", "Energy", "Solar")
	signalHash := ComputeVoteSignal(ctx.ChainID(), "E2EDomain", "Energy", "Solar", 5)
	proofBytes, nullifierHash, err := GenerateMembershipProofForSignal(keys, secret, tree.Root, siblings, pathIndices, extNullifier, signalHash)
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
	k.CreateDomain(ctx, "HistDomain", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 500_000)))

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
	k.CreateDomain(ctx, "CapDomain", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 500_000)))

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
	proofHex, nullifierHex := generateZKPRating(t, k, ctx, "HistRateDomain", secrets, 1, "Climate", "GreenDeal", 3)

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
	proofHex, nullifierHex := generateZKPRating(t, k, ctx, "ExpiredDomain", secrets, 0, "Climate", "GreenDeal", 3)

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

	proofHex, nullifierHex := generateZKPRating(t, k, ctx, "EmptyRootDomain", secrets, 0, "Climate", "GreenDeal", 4)

	// Empty merkleRootHex → uses current domain root (existing behavior).
	_, err := k.RateProposalWithZKP(ctx, "EmptyRootDomain", "Climate", "GreenDeal", 4, proofHex, nullifierHex, "")
	if err != nil {
		t.Fatalf("rating with empty merkle root should use current root: %v", err)
	}
}

func TestBigPurgeClearsRootHistory(t *testing.T) {
	k, ctx := setupKeeper(t)
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "PurgeHistDomain", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 500_000)))

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
	proof1, null1 := generateZKPRating(t, k, ctx, "PurgeDomain", secrets, 0, "Climate", "GreenDeal", 3)
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
	proof2, null2 := generateZKPRating(t, k, ctx, "PurgeDomain", newSecrets, 0, "Climate", "GreenDeal", -2)
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
