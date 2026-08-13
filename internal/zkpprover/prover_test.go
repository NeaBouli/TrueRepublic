package zkpprover_test

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"truerepublic/internal/zkpprover"
	"truerepublic/x/truedemocracy"
)

type goldenVector struct {
	CircuitID            string   `json:"circuit_id"`
	SyntheticWitnessHex  string   `json:"synthetic_witness_hex"`
	MerkleRootHex        string   `json:"merkle_root_hex"`
	SiblingsHex          []string `json:"siblings_hex"`
	PathIndices          []int    `json:"path_indices"`
	ExternalNullifierHex string   `json:"external_nullifier_hex"`
	SignalHashHex        string   `json:"signal_hash_hex"`
	NullifierHashHex     string   `json:"nullifier_hash_hex"`
}

type fixtureManifest struct {
	Artifacts []struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
	} `json:"artifacts"`
}

func TestPinnedSyntheticProofIsAcceptedByNativeVerifier(t *testing.T) {
	request, cs, pk, vk := loadFixture(t)
	result, err := zkpprover.Prove(cs, pk, vk, request)
	if err != nil {
		t.Fatalf("generate pinned synthetic proof: %v", err)
	}
	if result.Schema != zkpprover.ResultSchema || result.CircuitID != truedemocracy.MembershipCircuitID ||
		!result.SyntheticAndTestOnly || len(result.PublicSignalsHex) != 4 {
		t.Fatalf("unsafe or incompatible result: %+v", result)
	}
	proof := decodeHex(t, "proof", result.ProofHex)
	root := decodeHex(t, "root", result.MerkleRootHex)
	nullifier := decodeHex(t, "nullifier", result.NullifierHashHex)
	external := decodeHex(t, "external nullifier", result.PublicSignalsHex[2])
	signal := decodeHex(t, "signal", result.PublicSignalsHex[3])
	validatedVK, err := truedemocracy.ValidateMembershipVerifyingKey(
		vk,
		truedemocracy.MembershipCircuitID,
		pinnedVerifyingKeyFingerprint(t),
	)
	if err != nil {
		t.Fatalf("validate pinned verifying key: %v", err)
	}
	if err := truedemocracy.VerifyMembershipProofForSignal(validatedVK, proof, root, nullifier, external, signal); err != nil {
		t.Fatalf("native verifier rejected maintained-client proof: %v", err)
	}
}

func TestProverFailsClosedBeforeProofGeneration(t *testing.T) {
	request, cs, pk, vk := loadFixture(t)

	t.Run("artifact drift", func(t *testing.T) {
		candidate := append([]byte(nil), pk...)
		candidate[len(candidate)/2] ^= 1
		if _, err := zkpprover.Prove(cs, candidate, vk, request); err == nil || !strings.Contains(err.Error(), "SHA-256 mismatch") {
			t.Fatalf("corrupted proving key result = %v", err)
		}
	})
	t.Run("unsafe metadata", func(t *testing.T) {
		candidate := request
		candidate.SyntheticAndTestOnly = false
		if _, err := zkpprover.Prove(cs, pk, vk, candidate); err == nil || !strings.Contains(err.Error(), "unsafe or incompatible") {
			t.Fatalf("production-marked request result = %v", err)
		}
	})
	t.Run("path shape", func(t *testing.T) {
		candidate := request
		candidate.PathIndices = append([]int(nil), request.PathIndices...)
		candidate.PathIndices[0] = 2
		if _, err := zkpprover.Prove(cs, pk, vk, candidate); err == nil || !strings.Contains(err.Error(), "must be 0 or 1") {
			t.Fatalf("invalid path result = %v", err)
		}
	})
	t.Run("noncanonical field", func(t *testing.T) {
		candidate := request
		candidate.SignalHashHex = strings.ToUpper(request.SignalHashHex)
		if _, err := zkpprover.Prove(cs, pk, vk, candidate); err == nil || !strings.Contains(err.Error(), "canonical lowercase") {
			t.Fatalf("uppercase field result = %v", err)
		}
	})
	t.Run("wrong witness", func(t *testing.T) {
		candidate := request
		candidate.IdentitySecretHex = strings.Repeat("01", 32)
		if _, err := zkpprover.Prove(cs, pk, vk, candidate); err == nil || !strings.Contains(err.Error(), "prove witness") {
			t.Fatalf("invalid witness result = %v", err)
		}
	})
}

func TestDecodeRequestStrictRejectsAmbiguousJSON(t *testing.T) {
	tests := []string{
		`{"schema":"a","schema":"b"}`,
		`{"unknown":true}`,
		`{} {}`,
	}
	for _, candidate := range tests {
		if _, err := zkpprover.DecodeRequestStrict([]byte(candidate)); err == nil {
			t.Fatalf("ambiguous JSON accepted: %s", candidate)
		}
	}
	tooDeep := strings.Repeat("[", maxTestJSONNestingDepth+2) + "0" +
		strings.Repeat("]", maxTestJSONNestingDepth+2)
	if _, err := zkpprover.DecodeRequestStrict([]byte(tooDeep)); err == nil ||
		!strings.Contains(err.Error(), "JSON nesting exceeds") {
		t.Fatalf("deeply nested JSON result = %v", err)
	}
}

const maxTestJSONNestingDepth = 8

func TestWASMClientOutputIsAcceptedByNativeVerifier(t *testing.T) {
	path := os.Getenv("TRUEREPUBLIC_ZKP_RESULT_PATH")
	if path == "" {
		t.Skip("set TRUEREPUBLIC_ZKP_RESULT_PATH in the dedicated WASM integration gate")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Proof         string   `json:"proof"`
		NullifierHash string   `json:"nullifierHash"`
		MerkleRoot    string   `json:"merkleRoot"`
		PublicSignals []string `json:"publicSignals"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	if len(result.PublicSignals) != 4 || result.PublicSignals[0] != result.MerkleRoot || result.PublicSignals[1] != result.NullifierHash {
		t.Fatalf("malformed public signal binding: %+v", result)
	}
	_, _, _, vk := loadFixture(t)
	validatedVK, err := truedemocracy.ValidateMembershipVerifyingKey(
		vk,
		truedemocracy.MembershipCircuitID,
		pinnedVerifyingKeyFingerprint(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := truedemocracy.VerifyMembershipProofForSignal(
		validatedVK,
		decodeHex(t, "proof", result.Proof),
		decodeHex(t, "root", result.MerkleRoot),
		decodeHex(t, "nullifier", result.NullifierHash),
		decodeHex(t, "external nullifier", result.PublicSignals[2]),
		decodeHex(t, "signal", result.PublicSignals[3]),
	); err != nil {
		t.Fatalf("native Go verifier rejected WASM-client proof: %v", err)
	}
	tamperedProof := decodeHex(t, "proof", result.Proof)
	tamperedProof[len(tamperedProof)/2] ^= 1
	if err := truedemocracy.VerifyMembershipProofForSignal(
		validatedVK,
		tamperedProof,
		decodeHex(t, "root", result.MerkleRoot),
		decodeHex(t, "nullifier", result.NullifierHash),
		decodeHex(t, "external nullifier", result.PublicSignals[2]),
		decodeHex(t, "signal", result.PublicSignals[3]),
	); err == nil {
		t.Fatal("native Go verifier accepted tampered WASM-client proof")
	}
}

func loadFixture(t *testing.T) (zkpprover.Request, []byte, []byte, []byte) {
	t.Helper()
	directory := filepath.Join("..", "..", "x", "truedemocracy", "testdata", "zkp")
	vectorBytes, err := os.ReadFile(filepath.Join(directory, "golden_vector.json"))
	if err != nil {
		t.Fatal(err)
	}
	var vector goldenVector
	if err := json.Unmarshal(vectorBytes, &vector); err != nil {
		t.Fatal(err)
	}
	read := func(name string) []byte {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	return zkpprover.Request{
		Schema:               zkpprover.RequestSchema,
		CircuitID:            vector.CircuitID,
		SyntheticAndTestOnly: true,
		IdentitySecretHex:    vector.SyntheticWitnessHex,
		MerkleRootHex:        vector.MerkleRootHex,
		SiblingsHex:          vector.SiblingsHex,
		PathIndices:          vector.PathIndices,
		ExternalNullifierHex: vector.ExternalNullifierHex,
		SignalHashHex:        vector.SignalHashHex,
	}, read("membership_v2.cs"), read("membership_v2.pk"), read("membership_v2.vk")
}

func pinnedVerifyingKeyFingerprint(t *testing.T) string {
	t.Helper()
	directory := filepath.Join("..", "..", "x", "truedemocracy", "testdata", "zkp")
	data, err := os.ReadFile(filepath.Join(directory, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest fixtureManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	for _, artifact := range manifest.Artifacts {
		if artifact.Path == "membership_v2.vk" && artifact.SHA256 != "" {
			return artifact.SHA256
		}
	}
	t.Fatal("manifest does not pin membership_v2.vk SHA-256")
	return ""
}

func decodeHex(t *testing.T, label, value string) []byte {
	t.Helper()
	decoded, err := hex.DecodeString(value)
	if err != nil {
		t.Fatalf("decode %s: %v", label, err)
	}
	return decoded
}
