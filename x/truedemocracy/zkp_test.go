package truedemocracy

import (
	"bytes"
	"math/big"
	"sync"
	"testing"
)

// Shared ZKP keys (expensive setup, run once).
var (
	testZKPKeys     *ZKPKeys
	testZKPKeysOnce sync.Once
	testZKPKeysErr  error
)

func getTestZKPKeys(t *testing.T) *ZKPKeys {
	t.Helper()
	testZKPKeysOnce.Do(func() {
		testZKPKeys, testZKPKeysErr = SetupMembershipCircuit()
	})
	if testZKPKeysErr != nil {
		t.Fatalf("ZKP setup failed: %v", testZKPKeysErr)
	}
	return testZKPKeys
}

// buildTestTree creates a Merkle tree with the given number of members
// and returns the tree, secrets, and commitments.
func buildTestTree(t *testing.T, numMembers int) (*MerkleTree, [][]byte, [][]byte) {
	t.Helper()
	secrets := make([][]byte, numMembers)
	commitments := make([][]byte, numMembers)
	for i := 0; i < numMembers; i++ {
		secrets[i] = big.NewInt(int64(i + 100)).Bytes()
		c, err := ComputeCommitment(secrets[i])
		if err != nil {
			t.Fatalf("ComputeCommitment failed: %v", err)
		}
		commitments[i] = c
	}
	tree := NewMerkleTree(MerkleTreeDepth)
	if err := tree.BuildFromLeaves(commitments); err != nil {
		t.Fatalf("BuildFromLeaves failed: %v", err)
	}
	return tree, secrets, commitments
}

func TestSetupMembershipCircuit(t *testing.T) {
	keys := getTestZKPKeys(t)
	if keys.ProvingKey == nil {
		t.Fatal("ProvingKey should not be nil")
	}
	if keys.VerifyingKey == nil {
		t.Fatal("VerifyingKey should not be nil")
	}
	if keys.CS == nil {
		t.Fatal("ConstraintSystem should not be nil")
	}
}

func TestGenerateAndVerifyProof(t *testing.T) {
	keys := getTestZKPKeys(t)
	tree, secrets, _ := buildTestTree(t, 5)

	// Prove membership for member at index 2.
	memberIdx := 2
	siblings, pathIndices, err := tree.GenerateProof(memberIdx)
	if err != nil {
		t.Fatalf("GenerateProof failed: %v", err)
	}

	extNullifier := []byte("TestDomain|Issue1|Suggestion1")
	proofBytes, nullifierHash, err := GenerateMembershipProof(
		keys, secrets[memberIdx], tree.Root, siblings, pathIndices, extNullifier,
	)
	if err != nil {
		t.Fatalf("GenerateMembershipProof failed: %v", err)
	}
	if len(proofBytes) == 0 {
		t.Fatal("proof should not be empty")
	}
	if len(nullifierHash) != 32 {
		t.Fatalf("nullifierHash should be 32 bytes, got %d", len(nullifierHash))
	}

	// Verify.
	err = VerifyMembershipProof(keys.VerifyingKey, proofBytes, tree.Root, nullifierHash, extNullifier)
	if err != nil {
		t.Fatalf("VerifyMembershipProof failed: %v", err)
	}
}

func TestProofWithWrongRootFails(t *testing.T) {
	keys := getTestZKPKeys(t)
	tree, secrets, _ := buildTestTree(t, 3)

	siblings, pathIndices, _ := tree.GenerateProof(0)
	extNullifier := []byte("TestDomain|Issue1|Suggestion1")

	proofBytes, nullifierHash, err := GenerateMembershipProof(
		keys, secrets[0], tree.Root, siblings, pathIndices, extNullifier,
	)
	if err != nil {
		t.Fatalf("proof generation failed: %v", err)
	}

	// Verify against wrong root.
	wrongRoot := MiMCHash(big.NewInt(999))
	err = VerifyMembershipProof(keys.VerifyingKey, proofBytes, wrongRoot, nullifierHash, extNullifier)
	if err == nil {
		t.Fatal("verification should fail with wrong root")
	}
}

func TestProofWithWrongNullifierFails(t *testing.T) {
	keys := getTestZKPKeys(t)
	tree, secrets, _ := buildTestTree(t, 3)

	siblings, pathIndices, _ := tree.GenerateProof(0)
	extNullifier := []byte("TestDomain|Issue1|Suggestion1")

	proofBytes, _, err := GenerateMembershipProof(
		keys, secrets[0], tree.Root, siblings, pathIndices, extNullifier,
	)
	if err != nil {
		t.Fatalf("proof generation failed: %v", err)
	}

	// Tamper with nullifier hash.
	wrongNullifier := MiMCHash(big.NewInt(888))
	err = VerifyMembershipProof(keys.VerifyingKey, proofBytes, tree.Root, wrongNullifier, extNullifier)
	if err == nil {
		t.Fatal("verification should fail with tampered nullifier")
	}
}

func TestZKPNullifierDeterminism(t *testing.T) {
	keys := getTestZKPKeys(t)
	tree, secrets, _ := buildTestTree(t, 3)

	siblings, pathIndices, _ := tree.GenerateProof(0)
	extNullifier := []byte("TestDomain|Issue1|Suggestion1")

	_, null1, err := GenerateMembershipProof(
		keys, secrets[0], tree.Root, siblings, pathIndices, extNullifier,
	)
	if err != nil {
		t.Fatalf("first proof failed: %v", err)
	}

	_, null2, err := GenerateMembershipProof(
		keys, secrets[0], tree.Root, siblings, pathIndices, extNullifier,
	)
	if err != nil {
		t.Fatalf("second proof failed: %v", err)
	}

	if !bytes.Equal(null1, null2) {
		t.Fatal("same inputs should produce same nullifier (critical for double-vote detection)")
	}
}

func TestZKPNullifierUniqueness(t *testing.T) {
	keys := getTestZKPKeys(t)
	tree, secrets, _ := buildTestTree(t, 3)

	siblings, pathIndices, _ := tree.GenerateProof(0)

	_, null1, err := GenerateMembershipProof(
		keys, secrets[0], tree.Root, siblings, pathIndices,
		[]byte("TestDomain|Issue1|Suggestion1"),
	)
	if err != nil {
		t.Fatalf("first proof failed: %v", err)
	}

	_, null2, err := GenerateMembershipProof(
		keys, secrets[0], tree.Root, siblings, pathIndices,
		[]byte("TestDomain|Issue1|Suggestion2"),
	)
	if err != nil {
		t.Fatalf("second proof failed: %v", err)
	}

	if bytes.Equal(null1, null2) {
		t.Fatal("different external nullifiers should produce different nullifiers")
	}
}

func TestDifferentSecretsUnlinkable(t *testing.T) {
	keys := getTestZKPKeys(t)
	tree, secrets, _ := buildTestTree(t, 3)
	extNullifier := []byte("TestDomain|Issue1|Suggestion1")

	siblings0, pathIndices0, _ := tree.GenerateProof(0)
	_, null0, err := GenerateMembershipProof(
		keys, secrets[0], tree.Root, siblings0, pathIndices0, extNullifier,
	)
	if err != nil {
		t.Fatalf("proof 0 failed: %v", err)
	}

	siblings1, pathIndices1, _ := tree.GenerateProof(1)
	_, null1, err := GenerateMembershipProof(
		keys, secrets[1], tree.Root, siblings1, pathIndices1, extNullifier,
	)
	if err != nil {
		t.Fatalf("proof 1 failed: %v", err)
	}

	if bytes.Equal(null0, null1) {
		t.Fatal("different secrets should produce different nullifiers (unlinkability)")
	}
}

func TestProofSerialization(t *testing.T) {
	keys := getTestZKPKeys(t)
	tree, secrets, _ := buildTestTree(t, 3)

	siblings, pathIndices, _ := tree.GenerateProof(0)
	extNullifier := []byte("TestDomain|Issue1|Suggestion1")

	proofBytes, nullifierHash, err := GenerateMembershipProof(
		keys, secrets[0], tree.Root, siblings, pathIndices, extNullifier,
	)
	if err != nil {
		t.Fatalf("proof generation failed: %v", err)
	}

	// Roundtrip: deserialize then re-serialize.
	proof, err := DeserializeProof(proofBytes)
	if err != nil {
		t.Fatalf("DeserializeProof failed: %v", err)
	}
	reserialized, err := SerializeProof(proof)
	if err != nil {
		t.Fatalf("SerializeProof failed: %v", err)
	}
	if !bytes.Equal(proofBytes, reserialized) {
		t.Fatal("proof serialization roundtrip mismatch")
	}

	// Verify the deserialized proof still works.
	err = VerifyMembershipProof(keys.VerifyingKey, reserialized, tree.Root, nullifierHash, extNullifier)
	if err != nil {
		t.Fatalf("verification after roundtrip failed: %v", err)
	}
}

func TestVerifyingKeySerialization(t *testing.T) {
	keys := getTestZKPKeys(t)

	vkBytes, err := SerializeVerifyingKey(keys.VerifyingKey)
	if err != nil {
		t.Fatalf("SerializeVerifyingKey failed: %v", err)
	}
	if len(vkBytes) == 0 {
		t.Fatal("serialized verifying key should not be empty")
	}

	vk2, err := DeserializeVerifyingKey(vkBytes)
	if err != nil {
		t.Fatalf("DeserializeVerifyingKey failed: %v", err)
	}

	// Verify a proof with the deserialized key.
	tree, secrets, _ := buildTestTree(t, 3)
	siblings, pathIndices, _ := tree.GenerateProof(0)
	extNullifier := []byte("TestDomain|Issue1|Suggestion1")

	proofBytes, nullifierHash, err := GenerateMembershipProof(
		keys, secrets[0], tree.Root, siblings, pathIndices, extNullifier,
	)
	if err != nil {
		t.Fatalf("proof generation failed: %v", err)
	}

	err = VerifyMembershipProof(vk2, proofBytes, tree.Root, nullifierHash, extNullifier)
	if err != nil {
		t.Fatalf("verification with deserialized VK failed: %v", err)
	}
}
