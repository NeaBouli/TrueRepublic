// Package zkpcircuit defines the versioned membership-vote circuit shared by
// the chain verifier and the isolated test-only maintained-client prover.
package zkpcircuit

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
)

const (
	// ID is the frozen circuit identifier pinned by the repository ZKP spec.
	ID = "truerepublic/membership-vote/v2-bn254-mimc-depth20"
	// MerkleDepth supports up to 2^20 identity commitments per domain.
	MerkleDepth = 20
	// PublicWitnessCount is the exact number of public circuit inputs.
	PublicWitnessCount = 4
)

// MembershipCircuit proves knowledge of an identity secret whose commitment
// belongs to a fixed-depth Merkle tree and binds one stable nullifier and one
// non-zero signal to the proof.
type MembershipCircuit struct {
	MerkleRoot        frontend.Variable `gnark:",public"`
	NullifierHash     frontend.Variable `gnark:",public"`
	ExternalNullifier frontend.Variable `gnark:",public"`
	SignalHash        frontend.Variable `gnark:",public"`

	IdentitySecret frontend.Variable
	Siblings       [MerkleDepth]frontend.Variable
	PathIndices    [MerkleDepth]frontend.Variable
}

// Define implements frontend.Circuit.
func (c *MembershipCircuit) Define(api frontend.API) error {
	commitHasher, err := mimc.NewMiMC(api)
	if err != nil {
		return fmt.Errorf("mimc init for commitment: %w", err)
	}
	commitHasher.Write(c.IdentitySecret)
	commitment := commitHasher.Sum()

	currentHash := commitment
	for i := 0; i < MerkleDepth; i++ {
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

	nullifierHasher, err := mimc.NewMiMC(api)
	if err != nil {
		return fmt.Errorf("mimc init for nullifier: %w", err)
	}
	nullifierHasher.Write(c.IdentitySecret, c.ExternalNullifier)
	api.AssertIsEqual(nullifierHasher.Sum(), c.NullifierHash)
	api.AssertIsDifferent(c.SignalHash, 0)
	return nil
}
