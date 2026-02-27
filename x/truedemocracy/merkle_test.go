package truedemocracy

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// ---------- MiMC Hash Tests ----------

func TestMiMCHashDeterministic(t *testing.T) {
	val := big.NewInt(42)
	h1 := MiMCHash(val)
	h2 := MiMCHash(val)
	if !bytes.Equal(h1, h2) {
		t.Fatal("MiMCHash should be deterministic")
	}
	if len(h1) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(h1))
	}
}

func TestMiMCHashDifferentInputs(t *testing.T) {
	h1 := MiMCHash(big.NewInt(1))
	h2 := MiMCHash(big.NewInt(2))
	if bytes.Equal(h1, h2) {
		t.Fatal("different inputs should produce different hashes")
	}
}

func TestMiMCHashMultipleInputs(t *testing.T) {
	h1 := MiMCHash(big.NewInt(1), big.NewInt(2))
	h2 := MiMCHash(big.NewInt(2), big.NewInt(1))
	if bytes.Equal(h1, h2) {
		t.Fatal("different input order should produce different hashes")
	}
}

// ---------- Commitment & Nullifier Tests ----------

func TestComputeCommitment(t *testing.T) {
	secret := make([]byte, 32)
	secret[31] = 0x42
	commitment, err := ComputeCommitment(secret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(commitment) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(commitment))
	}
	// Non-zero.
	allZero := true
	for _, b := range commitment {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		t.Fatal("commitment should not be all zeros")
	}
}

func TestComputeCommitmentDeterministic(t *testing.T) {
	secret := []byte{0x01, 0x02, 0x03}
	c1, _ := ComputeCommitment(secret)
	c2, _ := ComputeCommitment(secret)
	if !bytes.Equal(c1, c2) {
		t.Fatal("same secret should produce same commitment")
	}
}

func TestComputeNullifier(t *testing.T) {
	secret := []byte{0x01, 0x02, 0x03}
	ext := []byte{0x04, 0x05, 0x06}
	nullifier, err := ComputeNullifier(secret, ext)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nullifier) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(nullifier))
	}
}

func TestComputeNullifierDeterministic(t *testing.T) {
	secret := []byte{0xAA}
	ext := []byte{0xBB}
	n1, _ := ComputeNullifier(secret, ext)
	n2, _ := ComputeNullifier(secret, ext)
	if !bytes.Equal(n1, n2) {
		t.Fatal("same inputs should produce same nullifier")
	}
}

func TestComputeNullifierDifferentContext(t *testing.T) {
	secret := []byte{0xAA}
	ext1 := []byte{0xBB}
	ext2 := []byte{0xCC}
	n1, _ := ComputeNullifier(secret, ext1)
	n2, _ := ComputeNullifier(secret, ext2)
	if bytes.Equal(n1, n2) {
		t.Fatal("different external nullifiers should produce different nullifiers")
	}
}

// ---------- Merkle Tree Tests ----------

func TestNewMerkleTree(t *testing.T) {
	tree := NewMerkleTree(MerkleTreeDepth)
	if tree.Depth != MerkleTreeDepth {
		t.Fatalf("expected depth %d, got %d", MerkleTreeDepth, tree.Depth)
	}
	if len(tree.Root) != 32 {
		t.Fatalf("expected 32-byte root, got %d", len(tree.Root))
	}
	// Empty tree root should be deterministic.
	tree2 := NewMerkleTree(MerkleTreeDepth)
	if !bytes.Equal(tree.Root, tree2.Root) {
		t.Fatal("empty tree roots should be equal")
	}
}

func TestMerkleTreeSingleLeaf(t *testing.T) {
	tree := NewMerkleTree(MerkleTreeDepth)
	leaf := MiMCHash(big.NewInt(42))
	if err := tree.BuildFromLeaves([][]byte{leaf}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Root should differ from empty tree.
	empty := NewMerkleTree(MerkleTreeDepth)
	if bytes.Equal(tree.Root, empty.Root) {
		t.Fatal("tree with one leaf should have different root than empty tree")
	}
}

func TestMerkleTreeBuildFromLeaves(t *testing.T) {
	secrets := [][]byte{{0x01}, {0x02}, {0x03}, {0x04}, {0x05}}
	leaves := make([][]byte, len(secrets))
	for i, s := range secrets {
		c, _ := ComputeCommitment(s)
		leaves[i] = c
	}

	tree := NewMerkleTree(MerkleTreeDepth)
	if err := tree.BuildFromLeaves(leaves); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Root should change when we add another leaf.
	leaves2 := append(leaves, MiMCHash(big.NewInt(6)))
	tree2 := NewMerkleTree(MerkleTreeDepth)
	tree2.BuildFromLeaves(leaves2)
	if bytes.Equal(tree.Root, tree2.Root) {
		t.Fatal("different leaves should produce different roots")
	}
}

func TestMerkleProofGeneration(t *testing.T) {
	leaves := make([][]byte, 5)
	for i := range leaves {
		c, _ := ComputeCommitment([]byte{byte(i + 1)})
		leaves[i] = c
	}

	tree := NewMerkleTree(MerkleTreeDepth)
	tree.BuildFromLeaves(leaves)

	siblings, pathIndices, err := tree.GenerateProof(2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(siblings) != MerkleTreeDepth {
		t.Fatalf("expected %d siblings, got %d", MerkleTreeDepth, len(siblings))
	}
	if len(pathIndices) != MerkleTreeDepth {
		t.Fatalf("expected %d path indices, got %d", MerkleTreeDepth, len(pathIndices))
	}
	// All path indices should be 0 or 1.
	for i, pi := range pathIndices {
		if pi != 0 && pi != 1 {
			t.Fatalf("path index %d is %d, expected 0 or 1", i, pi)
		}
	}
}

func TestMerkleProofVerification(t *testing.T) {
	leaves := make([][]byte, 5)
	for i := range leaves {
		c, _ := ComputeCommitment([]byte{byte(i + 1)})
		leaves[i] = c
	}

	tree := NewMerkleTree(MerkleTreeDepth)
	tree.BuildFromLeaves(leaves)

	for i := 0; i < len(leaves); i++ {
		siblings, pathIndices, err := tree.GenerateProof(i)
		if err != nil {
			t.Fatalf("leaf %d: proof generation failed: %v", i, err)
		}
		if !VerifyMerkleProof(tree.Root, leaves[i], siblings, pathIndices) {
			t.Fatalf("leaf %d: proof verification failed", i)
		}
	}
}

func TestMerkleProofWrongLeafFails(t *testing.T) {
	leaves := make([][]byte, 3)
	for i := range leaves {
		c, _ := ComputeCommitment([]byte{byte(i + 1)})
		leaves[i] = c
	}

	tree := NewMerkleTree(MerkleTreeDepth)
	tree.BuildFromLeaves(leaves)

	siblings, pathIndices, _ := tree.GenerateProof(0)

	// Use a wrong leaf.
	wrongLeaf := MiMCHash(big.NewInt(999))
	if VerifyMerkleProof(tree.Root, wrongLeaf, siblings, pathIndices) {
		t.Fatal("verification should fail with wrong leaf")
	}
}

func TestMerkleProofWrongRootFails(t *testing.T) {
	leaves := make([][]byte, 3)
	for i := range leaves {
		c, _ := ComputeCommitment([]byte{byte(i + 1)})
		leaves[i] = c
	}

	tree := NewMerkleTree(MerkleTreeDepth)
	tree.BuildFromLeaves(leaves)

	siblings, pathIndices, _ := tree.GenerateProof(0)

	wrongRoot := MiMCHash(big.NewInt(999))
	if VerifyMerkleProof(wrongRoot, leaves[0], siblings, pathIndices) {
		t.Fatal("verification should fail with wrong root")
	}
}

// ---------- HexToFieldElement Tests ----------

func TestHexToFieldElement(t *testing.T) {
	t.Run("valid hex", func(t *testing.T) {
		hexStr := hex.EncodeToString([]byte{0x01, 0x02, 0x03})
		result, err := HexToFieldElement(hexStr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 32 {
			t.Fatalf("expected 32 bytes, got %d", len(result))
		}
	})

	t.Run("invalid hex rejected", func(t *testing.T) {
		_, err := HexToFieldElement("gggg")
		if err == nil {
			t.Fatal("expected error for invalid hex")
		}
	})

	t.Run("overflow rejected", func(t *testing.T) {
		// BN254 modulus is ~21888..., use a value larger than modulus.
		modulus := fr.Modulus()
		overflow := new(big.Int).Add(modulus, big.NewInt(1))
		hexStr := hex.EncodeToString(overflow.Bytes())
		_, err := HexToFieldElement(hexStr)
		if err == nil {
			t.Fatal("expected error for value exceeding field modulus")
		}
	})
}
