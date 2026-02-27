package truedemocracy

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
)

// MerkleTreeDepth is the fixed depth of the Merkle tree.
// Supports up to 2^20 = 1,048,576 leaves per domain (Semaphore standard).
const MerkleTreeDepth = 20

// MiMCHash computes MiMC(data...) using the BN254 native hasher.
// Each element is a big.Int that must be < BN254 field modulus.
// Returns the hash as a 32-byte big-endian slice.
func MiMCHash(data ...*big.Int) []byte {
	hasher := mimc.NewMiMC()
	for _, d := range data {
		var buf [32]byte
		b := d.Bytes()
		copy(buf[32-len(b):], b)
		hasher.Write(buf[:])
	}
	return hasher.Sum(nil)
}

// MiMCHashBytes is a convenience wrapper that hashes raw 32-byte
// big-endian field elements.
func MiMCHashBytes(data ...[]byte) ([]byte, error) {
	vals := make([]*big.Int, len(data))
	for i, d := range data {
		if len(d) > 32 {
			return nil, fmt.Errorf("element %d exceeds 32 bytes", i)
		}
		vals[i] = new(big.Int).SetBytes(d)
	}
	return MiMCHash(vals...), nil
}

// ComputeCommitment computes commitment = MiMC(identitySecret).
func ComputeCommitment(identitySecret []byte) ([]byte, error) {
	if len(identitySecret) == 0 || len(identitySecret) > 32 {
		return nil, fmt.Errorf("identitySecret must be 1-32 bytes")
	}
	return MiMCHashBytes(identitySecret)
}

// ComputeNullifier computes nullifier = MiMC(identitySecret, externalNullifier).
func ComputeNullifier(identitySecret, externalNullifier []byte) ([]byte, error) {
	if len(identitySecret) == 0 || len(identitySecret) > 32 {
		return nil, fmt.Errorf("identitySecret must be 1-32 bytes")
	}
	if len(externalNullifier) == 0 || len(externalNullifier) > 32 {
		return nil, fmt.Errorf("externalNullifier must be 1-32 bytes")
	}
	return MiMCHashBytes(identitySecret, externalNullifier)
}

// zeroValues returns precomputed zero hashes for each tree level.
// Level 0 = MiMC(0) (empty leaf hash).
// Level i = MiMC(zeroValues[i-1], zeroValues[i-1]) (empty subtree).
func zeroValues(depth int) [][]byte {
	zeros := make([][]byte, depth+1)
	zeros[0] = MiMCHash(big.NewInt(0))
	for i := 1; i <= depth; i++ {
		prev := new(big.Int).SetBytes(zeros[i-1])
		zeros[i] = MiMCHash(prev, prev)
	}
	return zeros
}

// MerkleTree is a fixed-depth binary Merkle tree using MiMC hash.
type MerkleTree struct {
	Depth  int
	Leaves [][]byte // actual leaf values (commitments)
	Root   []byte   // current root hash
	zeros  [][]byte // precomputed zero hashes per level
}

// NewMerkleTree creates a new empty Merkle tree of the given depth.
func NewMerkleTree(depth int) *MerkleTree {
	t := &MerkleTree{
		Depth:  depth,
		Leaves: [][]byte{},
		zeros:  zeroValues(depth),
	}
	t.Root = t.zeros[depth]
	return t
}

// BuildFromLeaves rebuilds the tree from a list of leaf values.
func (t *MerkleTree) BuildFromLeaves(leaves [][]byte) error {
	maxLeaves := 1 << t.Depth
	if len(leaves) > maxLeaves {
		return fmt.Errorf("too many leaves: %d > max %d", len(leaves), maxLeaves)
	}
	t.Leaves = make([][]byte, len(leaves))
	copy(t.Leaves, leaves)
	t.Root = t.computeRoot()
	return nil
}

// computeRoot computes the Merkle root from the current leaves using
// sparse tree optimization (precomputed zero hashes for empty subtrees).
func (t *MerkleTree) computeRoot() []byte {
	if len(t.Leaves) == 0 {
		return t.zeros[t.Depth]
	}

	// Start with leaf level.
	currentLevel := make([][]byte, len(t.Leaves))
	copy(currentLevel, t.Leaves)

	for level := 0; level < t.Depth; level++ {
		var nextLevel [][]byte
		levelSize := len(currentLevel)

		for i := 0; i < levelSize; i += 2 {
			left := currentLevel[i]
			var right []byte
			if i+1 < levelSize {
				right = currentLevel[i+1]
			} else {
				// Odd node — pair with zero hash at this level.
				right = t.zeros[level]
			}
			lVal := new(big.Int).SetBytes(left)
			rVal := new(big.Int).SetBytes(right)
			nextLevel = append(nextLevel, MiMCHash(lVal, rVal))
		}

		if len(nextLevel) == 0 {
			return t.zeros[t.Depth]
		}
		currentLevel = nextLevel
	}
	return currentLevel[0]
}

// GetRoot returns the current Merkle root as hex string.
func (t *MerkleTree) GetRoot() string {
	return hex.EncodeToString(t.Root)
}

// GenerateProof returns the Merkle proof for the leaf at the given index.
// Returns siblings (one per level) and pathIndices (0=left, 1=right).
func (t *MerkleTree) GenerateProof(leafIndex int) (siblings [][]byte, pathIndices []int, err error) {
	if leafIndex < 0 || leafIndex >= len(t.Leaves) {
		return nil, nil, fmt.Errorf("leaf index %d out of range [0, %d)", leafIndex, len(t.Leaves))
	}

	siblings = make([][]byte, t.Depth)
	pathIndices = make([]int, t.Depth)

	// Rebuild levels to extract siblings.
	currentLevel := make([][]byte, len(t.Leaves))
	copy(currentLevel, t.Leaves)
	idx := leafIndex

	for level := 0; level < t.Depth; level++ {
		// Determine sibling index.
		var siblingIdx int
		if idx%2 == 0 {
			siblingIdx = idx + 1
			pathIndices[level] = 0 // current node is left child
		} else {
			siblingIdx = idx - 1
			pathIndices[level] = 1 // current node is right child
		}

		// Get sibling value.
		if siblingIdx < len(currentLevel) {
			siblings[level] = currentLevel[siblingIdx]
		} else {
			siblings[level] = t.zeros[level]
		}

		// Compute next level.
		var nextLevel [][]byte
		for i := 0; i < len(currentLevel); i += 2 {
			left := currentLevel[i]
			var right []byte
			if i+1 < len(currentLevel) {
				right = currentLevel[i+1]
			} else {
				right = t.zeros[level]
			}
			lVal := new(big.Int).SetBytes(left)
			rVal := new(big.Int).SetBytes(right)
			nextLevel = append(nextLevel, MiMCHash(lVal, rVal))
		}
		currentLevel = nextLevel
		idx = idx / 2
	}

	return siblings, pathIndices, nil
}

// VerifyMerkleProof verifies a Merkle proof against a root.
func VerifyMerkleProof(root, leaf []byte, siblings [][]byte, pathIndices []int) bool {
	if len(siblings) != len(pathIndices) {
		return false
	}
	currentHash := leaf
	for i := 0; i < len(siblings); i++ {
		var left, right *big.Int
		if pathIndices[i] == 0 {
			left = new(big.Int).SetBytes(currentHash)
			right = new(big.Int).SetBytes(siblings[i])
		} else {
			left = new(big.Int).SetBytes(siblings[i])
			right = new(big.Int).SetBytes(currentHash)
		}
		currentHash = MiMCHash(left, right)
	}
	return new(big.Int).SetBytes(currentHash).Cmp(new(big.Int).SetBytes(root)) == 0
}

// ComputeExternalNullifier hashes a context string into a BN254 field element.
// Used to derive deterministic nullifiers: nullifier = MiMC(secret, externalNullifier).
// Both client and chain must compute the same value for proof verification.
// The context is typically "domainName|issueName|suggestionName" for ratings.
func ComputeExternalNullifier(context string) ([]byte, error) {
	h := sha256.Sum256([]byte(context))
	n := new(big.Int).SetBytes(h[:])
	n.Mod(n, ecc.BN254.ScalarField())
	b := make([]byte, 32)
	nBytes := n.Bytes()
	copy(b[32-len(nBytes):], nBytes)
	return b, nil
}

// HexToFieldElement converts a hex string to a 32-byte big-endian
// field element, validating it is < BN254 field modulus.
func HexToFieldElement(hexStr string) ([]byte, error) {
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hex: %w", err)
	}
	if len(b) > 32 {
		return nil, fmt.Errorf("value exceeds 32 bytes")
	}
	val := new(big.Int).SetBytes(b)
	if val.Cmp(fr.Modulus()) >= 0 {
		return nil, fmt.Errorf("value exceeds BN254 field modulus")
	}
	// Pad to 32 bytes.
	var padded [32]byte
	copy(padded[32-len(b):], b)
	return padded[:], nil
}
