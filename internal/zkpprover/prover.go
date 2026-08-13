// Package zkpprover implements the isolated, test-only Groth16 prover used by
// the maintained-client compatibility harness. It must never be treated as a
// production ceremony or submission boundary.
package zkpprover

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	bn254mimc "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"

	"truerepublic/x/truedemocracy/zkpcircuit"
)

const (
	RequestSchema = "truerepublic/zkp-prover-request/v1"
	ResultSchema  = "truerepublic/zkp-prover-result/v1"

	ConstraintSystemSize   = 1_656_948
	ConstraintSystemSHA256 = "2a9bc20a39f076fe39d931164c6b29b857f3372cb573c54e5381840f77abaa10"
	ProvingKeySize         = 2_382_883
	ProvingKeySHA256       = "4775ed0a77f64e105a0a995958c45344eeda888ca26bafa07abf141b9d9952ed"
	VerifyingKeySize       = 460
	VerifyingKeySHA256     = "80b92df9562e48d4b25df9e7105e54f6d79250a3f35171250fcfd45c1489e289"
)

// Request contains one synthetic witness and its frozen public context. All
// field elements use canonical lowercase big-endian hex.
type Request struct {
	Schema               string   `json:"schema"`
	CircuitID            string   `json:"circuit_id"`
	SyntheticAndTestOnly bool     `json:"synthetic_and_test_only"`
	IdentitySecretHex    string   `json:"identity_secret_hex"`
	MerkleRootHex        string   `json:"merkle_root_hex"`
	SiblingsHex          []string `json:"siblings_hex"`
	PathIndices          []int    `json:"path_indices"`
	ExternalNullifierHex string   `json:"external_nullifier_hex"`
	SignalHashHex        string   `json:"signal_hash_hex"`
}

// Result is a canonical proof response. PublicSignalsHex follows the frozen
// circuit order: root, nullifier, external nullifier, signal.
type Result struct {
	Schema               string   `json:"schema"`
	CircuitID            string   `json:"circuit_id"`
	SyntheticAndTestOnly bool     `json:"synthetic_and_test_only"`
	ProofHex             string   `json:"proof_hex"`
	NullifierHashHex     string   `json:"nullifier_hash_hex"`
	MerkleRootHex        string   `json:"merkle_root_hex"`
	PublicSignalsHex     []string `json:"public_signals_hex"`
}

// DecodeRequestStrict rejects unknown fields, duplicate keys, and trailing
// values before any private witness is accepted by the prover.
func DecodeRequestStrict(data []byte) (Request, error) {
	if err := rejectDuplicateJSONKeys(json.NewDecoder(bytes.NewReader(data))); err != nil {
		return Request{}, fmt.Errorf("strict request JSON: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var request Request
	if err := decoder.Decode(&request); err != nil {
		return Request{}, fmt.Errorf("strict request JSON: %w", err)
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		return Request{}, fmt.Errorf("strict request JSON: trailing value")
	}
	return request, nil
}

// Prove loads only the exact pinned synthetic artifacts, proves the validated
// witness, and verifies the generated proof before returning it.
func Prove(csBytes, pkBytes, vkBytes []byte, request Request) (Result, error) {
	if err := verifyArtifact("constraint system", csBytes, ConstraintSystemSize, ConstraintSystemSHA256); err != nil {
		return Result{}, err
	}
	if err := verifyArtifact("proving key", pkBytes, ProvingKeySize, ProvingKeySHA256); err != nil {
		return Result{}, err
	}
	if err := verifyArtifact("verifying key", vkBytes, VerifyingKeySize, VerifyingKeySHA256); err != nil {
		return Result{}, err
	}

	assignment, publicInputs, identitySecret, err := validateRequest(request)
	if err != nil {
		return Result{}, err
	}
	defer clear(identitySecret)

	cs := groth16.NewCS(ecc.BN254)
	if err := readCanonical("constraint system", csBytes, cs); err != nil {
		return Result{}, err
	}
	pk := groth16.NewProvingKey(ecc.BN254)
	if err := readCanonical("proving key", pkBytes, pk); err != nil {
		return Result{}, err
	}
	vk := groth16.NewVerifyingKey(ecc.BN254)
	if err := readCanonical("verifying key", vkBytes, vk); err != nil {
		return Result{}, err
	}
	if vk.CurveID() != ecc.BN254 || vk.NbPublicWitness() != zkpcircuit.PublicWitnessCount {
		return Result{}, fmt.Errorf("verifying key public-input shape mismatch")
	}

	witness, err := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
	if err != nil {
		return Result{}, fmt.Errorf("create witness: %w", err)
	}
	proof, err := groth16.Prove(cs, pk, witness)
	if err != nil {
		return Result{}, fmt.Errorf("prove witness: %w", err)
	}
	publicWitness, err := frontend.NewWitness(&assignment, ecc.BN254.ScalarField(), frontend.PublicOnly())
	if err != nil {
		return Result{}, fmt.Errorf("create public witness: %w", err)
	}
	if err := groth16.Verify(proof, vk, publicWitness); err != nil {
		return Result{}, fmt.Errorf("self-verify generated proof: %w", err)
	}

	var proofBuffer bytes.Buffer
	if _, err := proof.WriteTo(&proofBuffer); err != nil {
		return Result{}, fmt.Errorf("serialize proof: %w", err)
	}
	return Result{
		Schema:               ResultSchema,
		CircuitID:            zkpcircuit.ID,
		SyntheticAndTestOnly: true,
		ProofHex:             hex.EncodeToString(proofBuffer.Bytes()),
		NullifierHashHex:     hex.EncodeToString(publicInputs[1]),
		MerkleRootHex:        hex.EncodeToString(publicInputs[0]),
		PublicSignalsHex: []string{
			hex.EncodeToString(publicInputs[0]),
			hex.EncodeToString(publicInputs[1]),
			hex.EncodeToString(publicInputs[2]),
			hex.EncodeToString(publicInputs[3]),
		},
	}, nil
}

type canonicalReader interface {
	ReadFrom(io.Reader) (int64, error)
	WriteTo(io.Writer) (int64, error)
}

func readCanonical(label string, data []byte, target canonicalReader) error {
	reader := bytes.NewReader(data)
	if _, err := target.ReadFrom(reader); err != nil {
		return fmt.Errorf("decode %s: %w", label, err)
	}
	if reader.Len() != 0 {
		return fmt.Errorf("decode %s: %d trailing bytes", label, reader.Len())
	}
	var canonical bytes.Buffer
	if _, err := target.WriteTo(&canonical); err != nil {
		return fmt.Errorf("re-encode %s: %w", label, err)
	}
	if !bytes.Equal(canonical.Bytes(), data) {
		return fmt.Errorf("%s encoding is not canonical", label)
	}
	return nil
}

func verifyArtifact(label string, data []byte, expectedSize int, expectedSHA256 string) error {
	if len(data) != expectedSize {
		return fmt.Errorf("%s size = %d, want %d", label, len(data), expectedSize)
	}
	digest := sha256.Sum256(data)
	if hex.EncodeToString(digest[:]) != expectedSHA256 {
		return fmt.Errorf("%s SHA-256 mismatch", label)
	}
	return nil
}

func validateRequest(request Request) (zkpcircuit.MembershipCircuit, [4][]byte, []byte, error) {
	if request.Schema != RequestSchema || request.CircuitID != zkpcircuit.ID || !request.SyntheticAndTestOnly {
		return zkpcircuit.MembershipCircuit{}, [4][]byte{}, nil, fmt.Errorf("unsafe or incompatible prover request metadata")
	}
	if len(request.SiblingsHex) != zkpcircuit.MerkleDepth || len(request.PathIndices) != zkpcircuit.MerkleDepth {
		return zkpcircuit.MembershipCircuit{}, [4][]byte{}, nil, fmt.Errorf("siblings and path_indices must contain exactly %d elements", zkpcircuit.MerkleDepth)
	}

	identity, err := decodeField("identity secret", request.IdentitySecretHex, false)
	if err != nil {
		return zkpcircuit.MembershipCircuit{}, [4][]byte{}, nil, err
	}
	root, err := decodeField("merkle root", request.MerkleRootHex, true)
	if err != nil {
		return zkpcircuit.MembershipCircuit{}, [4][]byte{}, nil, err
	}
	external, err := decodeField("external nullifier", request.ExternalNullifierHex, true)
	if err != nil {
		return zkpcircuit.MembershipCircuit{}, [4][]byte{}, nil, err
	}
	signal, err := decodeField("signal hash", request.SignalHashHex, true)
	if err != nil {
		return zkpcircuit.MembershipCircuit{}, [4][]byte{}, nil, err
	}
	if new(big.Int).SetBytes(signal).Sign() == 0 {
		return zkpcircuit.MembershipCircuit{}, [4][]byte{}, nil, fmt.Errorf("signal hash must be non-zero")
	}
	nullifier, err := hashElements(identity, external)
	if err != nil {
		return zkpcircuit.MembershipCircuit{}, [4][]byte{}, nil, fmt.Errorf("compute nullifier: %w", err)
	}

	assignment := zkpcircuit.MembershipCircuit{
		MerkleRoot:        new(big.Int).SetBytes(root),
		NullifierHash:     new(big.Int).SetBytes(nullifier),
		ExternalNullifier: new(big.Int).SetBytes(external),
		SignalHash:        new(big.Int).SetBytes(signal),
		IdentitySecret:    new(big.Int).SetBytes(identity),
	}
	for index := 0; index < zkpcircuit.MerkleDepth; index++ {
		sibling, err := decodeField(fmt.Sprintf("sibling %d", index), request.SiblingsHex[index], true)
		if err != nil {
			return zkpcircuit.MembershipCircuit{}, [4][]byte{}, nil, err
		}
		if request.PathIndices[index] != 0 && request.PathIndices[index] != 1 {
			return zkpcircuit.MembershipCircuit{}, [4][]byte{}, nil, fmt.Errorf("path index %d must be 0 or 1", index)
		}
		assignment.Siblings[index] = new(big.Int).SetBytes(sibling)
		assignment.PathIndices[index] = request.PathIndices[index]
	}
	return assignment, [4][]byte{root, nullifier, external, signal}, identity, nil
}

func decodeField(label, value string, exact32 bool) ([]byte, error) {
	if value == "" || value != bytesToLowerASCII(value) || len(value)%2 != 0 {
		return nil, fmt.Errorf("%s must be non-empty canonical lowercase hex", label)
	}
	decoded, err := hex.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("%s must be non-empty canonical lowercase hex", label)
	}
	if (exact32 && len(decoded) != 32) || (!exact32 && (len(decoded) == 0 || len(decoded) > 32)) {
		if exact32 {
			return nil, fmt.Errorf("%s must be exactly 32 bytes", label)
		}
		return nil, fmt.Errorf("%s must contain 1-32 bytes", label)
	}
	if new(big.Int).SetBytes(decoded).Cmp(ecc.BN254.ScalarField()) >= 0 {
		return nil, fmt.Errorf("%s is not a canonical BN254 field element", label)
	}
	return decoded, nil
}

func bytesToLowerASCII(value string) string {
	result := make([]byte, len(value))
	for index := range value {
		character := value[index]
		if character >= 'A' && character <= 'Z' {
			character += 'a' - 'A'
		}
		result[index] = character
	}
	return string(result)
}

func hashElements(elements ...[]byte) ([]byte, error) {
	hasher := bn254mimc.NewMiMC()
	for _, element := range elements {
		var padded [32]byte
		copy(padded[len(padded)-len(element):], element)
		if _, err := hasher.Write(padded[:]); err != nil {
			return nil, err
		}
	}
	return hasher.Sum(nil), nil
}

func clear(data []byte) {
	for index := range data {
		data[index] = 0
	}
}

func rejectDuplicateJSONKeys(decoder *json.Decoder) error {
	if err := rejectDuplicateJSONValue(decoder); err != nil {
		return err
	}
	switch _, err := decoder.Token(); {
	case err == io.EOF:
		return nil
	case err != nil:
		return err
	default:
		return fmt.Errorf("trailing JSON value")
	}
}

func rejectDuplicateJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		seen := make(map[string]struct{})
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return fmt.Errorf("JSON object key is not a string")
			}
			if _, duplicate := seen[key]; duplicate {
				return fmt.Errorf("duplicate JSON key %q", key)
			}
			seen[key] = struct{}{}
			if err := rejectDuplicateJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	case '[':
		for decoder.More() {
			if err := rejectDuplicateJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delimiter)
	}
}
