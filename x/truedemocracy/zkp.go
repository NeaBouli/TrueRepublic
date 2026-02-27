package truedemocracy

import (
	"bytes"
	"fmt"
	"math/big"
	"sync"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/consensys/gnark/std/hash/mimc"
)

// MembershipCircuit is the Groth16 circuit for anonymous set membership.
// It proves knowledge of an identitySecret whose MiMC hash is a leaf
// in the Merkle tree, and computes a deterministic nullifier.
type MembershipCircuit struct {
	// Public inputs (known to verifier).
	MerkleRoot        frontend.Variable `gnark:",public"`
	NullifierHash     frontend.Variable `gnark:",public"`
	ExternalNullifier frontend.Variable `gnark:",public"`

	// Private inputs (known only to prover).
	IdentitySecret frontend.Variable
	Siblings       [MerkleTreeDepth]frontend.Variable
	PathIndices    [MerkleTreeDepth]frontend.Variable
}

// Define implements frontend.Circuit. It constrains:
// 1. commitment = MiMC(identitySecret)
// 2. Merkle path from commitment to root
// 3. nullifier = MiMC(identitySecret, externalNullifier)
// 4. nullifier == NullifierHash
// 5. All PathIndices are boolean
func (c *MembershipCircuit) Define(api frontend.API) error {
	// 1. Compute commitment = MiMC(identitySecret).
	commitHasher, err := mimc.NewMiMC(api)
	if err != nil {
		return fmt.Errorf("mimc init for commitment: %w", err)
	}
	commitHasher.Write(c.IdentitySecret)
	commitment := commitHasher.Sum()

	// 2. Verify Merkle path from commitment to root.
	currentHash := commitment
	for i := 0; i < MerkleTreeDepth; i++ {
		// PathIndices[i] == 0: current is left child, sibling is right.
		// PathIndices[i] == 1: current is right child, sibling is left.
		api.AssertIsBoolean(c.PathIndices[i])

		left := api.Select(c.PathIndices[i], c.Siblings[i], currentHash)
		right := api.Select(c.PathIndices[i], currentHash, c.Siblings[i])

		levelHasher, err := mimc.NewMiMC(api)
		if err != nil {
			return fmt.Errorf("mimc init for level %d: %w", i, err)
		}
		levelHasher.Write(left, right)
		currentHash = levelHasher.Sum()
	}
	api.AssertIsEqual(currentHash, c.MerkleRoot)

	// 3. Compute nullifier = MiMC(identitySecret, externalNullifier).
	nullHasher, err := mimc.NewMiMC(api)
	if err != nil {
		return fmt.Errorf("mimc init for nullifier: %w", err)
	}
	nullHasher.Write(c.IdentitySecret, c.ExternalNullifier)
	nullifier := nullHasher.Sum()

	// 4. Assert nullifier matches public input.
	api.AssertIsEqual(nullifier, c.NullifierHash)

	return nil
}

// ZKPKeys holds the compiled circuit artifacts for Groth16.
type ZKPKeys struct {
	ProvingKey   groth16.ProvingKey
	VerifyingKey groth16.VerifyingKey
	CS           constraint.ConstraintSystem
}

// Global cached keys (setup is expensive, ~seconds).
var (
	cachedKeys     *ZKPKeys
	cachedKeysOnce sync.Once
	cachedKeysErr  error
)

// SetupMembershipCircuit compiles the circuit and runs the Groth16
// trusted setup. This is expensive and results are cached globally.
func SetupMembershipCircuit() (*ZKPKeys, error) {
	cachedKeysOnce.Do(func() {
		var circuit MembershipCircuit
		cs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
		if err != nil {
			cachedKeysErr = fmt.Errorf("circuit compilation failed: %w", err)
			return
		}
		pk, vk, err := groth16.Setup(cs)
		if err != nil {
			cachedKeysErr = fmt.Errorf("groth16 setup failed: %w", err)
			return
		}
		cachedKeys = &ZKPKeys{ProvingKey: pk, VerifyingKey: vk, CS: cs}
	})
	return cachedKeys, cachedKeysErr
}

// GenerateMembershipProof creates a Groth16 proof that the prover
// knows an identitySecret whose commitment is in the Merkle tree.
func GenerateMembershipProof(
	keys *ZKPKeys,
	identitySecret []byte,
	merkleRoot []byte,
	siblings [][]byte,
	pathIndices []int,
	externalNullifier []byte,
) (proofBytes []byte, nullifierHash []byte, err error) {
	if len(siblings) != MerkleTreeDepth || len(pathIndices) != MerkleTreeDepth {
		return nil, nil, fmt.Errorf("siblings and pathIndices must have length %d", MerkleTreeDepth)
	}

	// Compute nullifier off-chain for return value.
	nullifierHash, err = ComputeNullifier(identitySecret, externalNullifier)
	if err != nil {
		return nil, nil, fmt.Errorf("nullifier computation failed: %w", err)
	}

	// Build witness assignment.
	assignment := MembershipCircuit{
		MerkleRoot:        new(big.Int).SetBytes(merkleRoot),
		NullifierHash:     new(big.Int).SetBytes(nullifierHash),
		ExternalNullifier: new(big.Int).SetBytes(externalNullifier),
		IdentitySecret:    new(big.Int).SetBytes(identitySecret),
	}
	for i := 0; i < MerkleTreeDepth; i++ {
		assignment.Siblings[i] = new(big.Int).SetBytes(siblings[i])
		assignment.PathIndices[i] = pathIndices[i]
	}

	// Create witness.
	witness, err := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
	if err != nil {
		return nil, nil, fmt.Errorf("witness creation failed: %w", err)
	}

	// Generate proof.
	proof, err := groth16.Prove(keys.CS, keys.ProvingKey, witness)
	if err != nil {
		return nil, nil, fmt.Errorf("proof generation failed: %w", err)
	}

	proofBytes, err = SerializeProof(proof)
	if err != nil {
		return nil, nil, err
	}
	return proofBytes, nullifierHash, nil
}

// VerifyMembershipProof verifies a Groth16 membership proof.
func VerifyMembershipProof(
	vk groth16.VerifyingKey,
	proofBytes []byte,
	merkleRoot []byte,
	nullifierHash []byte,
	externalNullifier []byte,
) error {
	proof, err := DeserializeProof(proofBytes)
	if err != nil {
		return fmt.Errorf("proof deserialization failed: %w", err)
	}

	// Build public witness (only public inputs).
	publicAssignment := MembershipCircuit{
		MerkleRoot:        new(big.Int).SetBytes(merkleRoot),
		NullifierHash:     new(big.Int).SetBytes(nullifierHash),
		ExternalNullifier: new(big.Int).SetBytes(externalNullifier),
	}
	publicWitness, err := frontend.NewWitness(&publicAssignment, ecc.BN254.ScalarField(),
		frontend.PublicOnly())
	if err != nil {
		return fmt.Errorf("public witness creation failed: %w", err)
	}

	return groth16.Verify(proof, vk, publicWitness)
}

// SerializeProof serializes a Groth16 proof to bytes.
func SerializeProof(proof groth16.Proof) ([]byte, error) {
	var buf bytes.Buffer
	if _, err := proof.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("proof serialization failed: %w", err)
	}
	return buf.Bytes(), nil
}

// DeserializeProof deserializes bytes back to a Groth16 proof.
func DeserializeProof(data []byte) (groth16.Proof, error) {
	proof := groth16.NewProof(ecc.BN254)
	if _, err := proof.ReadFrom(bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("proof deserialization failed: %w", err)
	}
	return proof, nil
}

// SerializeVerifyingKey serializes a Groth16 verifying key to bytes.
func SerializeVerifyingKey(vk groth16.VerifyingKey) ([]byte, error) {
	var buf bytes.Buffer
	if _, err := vk.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("verifying key serialization failed: %w", err)
	}
	return buf.Bytes(), nil
}

// DeserializeVerifyingKey deserializes bytes back to a Groth16 verifying key.
func DeserializeVerifyingKey(data []byte) (groth16.VerifyingKey, error) {
	vk := groth16.NewVerifyingKey(ecc.BN254)
	if _, err := vk.ReadFrom(bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("verifying key deserialization failed: %w", err)
	}
	return vk, nil
}
