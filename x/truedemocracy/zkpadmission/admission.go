// Package zkpadmission classifies a genesis Groth16 verifying key for offline
// qualification evidence (GH300B4B). It depends only on the standard library so
// offline evidence tools can use it without gnark. It never changes consensus:
// chain genesis validation still accepts any canonical circuit-shaped key, and
// this classification only decides whether evidence may qualify that key.
package zkpadmission

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
)

const (
	PolicySchema = "truerepublic/zkp-vk-admission/v1"
	CircuitID    = "truerepublic/membership-vote/v2-bn254-mimc-depth20"
	// MaxVerifyingKeyBytes bounds hex decoding; the frozen circuit VK is 460 bytes.
	MaxVerifyingKeyBytes = 4096
	maxPolicyBytes       = 64 << 10

	FrozenTestOnlySHA256    = "80b92df9562e48d4b25df9e7105e54f6d79250a3f35171250fcfd45c1489e289"
	FrozenTestOnlySizeBytes = 460
)

// Class is the admission result for one genesis verifying-key triple.
type Class string

const (
	Absent             Class = "absent"
	TestOnlyDenied     Class = "test-only-denied"
	Unadmitted         Class = "unadmitted"
	ProductionAdmitted Class = "production-admitted"
)

// Violation codes for malformed or inconsistent verifying-key evidence.
var (
	ErrPartialTriple   = errors.New("partial-verifying-key-triple")
	ErrCircuitMismatch = errors.New("verifying-key-circuit-mismatch")
	ErrNonCanonicalHex = errors.New("non-canonical-verifying-key-hex")
	ErrOversizedKey    = errors.New("oversized-verifying-key")
	ErrInvalidDigest   = errors.New("invalid-verifying-key-digest")
	ErrDigestMismatch  = errors.New("verifying-key-digest-mismatch")
	errInvalidPolicy   = errors.New("invalid zkp vk admission policy")
	hex64              = regexp.MustCompile(`^[0-9a-f]{64}$`)
	lowercaseEvenHex   = regexp.MustCompile(`^(?:[0-9a-f]{2})+$`)
	requiredProvenance = []string{"ceremony_id", "phase1_source_sha256", "phase2_transcript_sha256", "contributor_count", "random_beacon", "independent_verification_report_sha256", "review_reference"}
)

// DeniedKey is a known single-party TEST-ONLY verifying key.
type DeniedKey struct {
	SHA256Hex string   `json:"sha256_hex"`
	SizeBytes int      `json:"size_bytes"`
	Sources   []string `json:"sources"`
}

// AdmittedKey is an independently qualified production verifying key. Every
// required provenance field must be present and non-empty.
type AdmittedKey struct {
	SHA256Hex  string            `json:"sha256_hex"`
	SizeBytes  int               `json:"size_bytes"`
	Provenance map[string]string `json:"provenance"`
}

// Policy mirrors configs/security/zkp-vk-admission.json exactly.
type Policy struct {
	Schema                       string        `json:"schema"`
	CircuitID                    string        `json:"circuit_id"`
	ProductionAdmitted           []AdmittedKey `json:"production_admitted"`
	TestOnlyDenied               []DeniedKey   `json:"test_only_denied"`
	RequiredProductionProvenance []string      `json:"required_production_provenance"`
}

// DefaultPolicy returns the compiled policy. The production allowlist is empty,
// so ProductionAdmitted is unreachable until a reviewed ceremony entry exists.
func DefaultPolicy() Policy {
	return Policy{
		Schema:             PolicySchema,
		CircuitID:          CircuitID,
		ProductionAdmitted: []AdmittedKey{},
		TestOnlyDenied: []DeniedKey{{
			// GH-198/GH-300 single-party TEST-ONLY toxic-waste verifying key.
			SHA256Hex: FrozenTestOnlySHA256,
			SizeBytes: FrozenTestOnlySizeBytes,
			Sources:   []string{"configs/security/zkp-browser-artifacts.json", "x/truedemocracy/testdata/zkp/manifest.json"},
		}},
		RequiredProductionProvenance: append([]string(nil), requiredProvenance...),
	}
}

// ParsePolicy decodes and validates a policy fail-closed: unknown or duplicate
// keys, trailing data, malformed digests or sizes, duplicates and deny/allow
// overlap are rejected.
func ParsePolicy(data []byte) (Policy, error) {
	if len(data) == 0 || len(data) > maxPolicyBytes {
		return Policy{}, fmt.Errorf("%w: size", errInvalidPolicy)
	}
	if err := rejectDuplicateKeys(data); err != nil {
		return Policy{}, fmt.Errorf("%w: %v", errInvalidPolicy, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var policy Policy
	if err := decoder.Decode(&policy); err != nil {
		return Policy{}, fmt.Errorf("%w: %v", errInvalidPolicy, err)
	}
	if _, err := decoder.Token(); err != io.EOF {
		return Policy{}, fmt.Errorf("%w: trailing data", errInvalidPolicy)
	}
	if err := policy.validate(); err != nil {
		return Policy{}, fmt.Errorf("%w: %v", errInvalidPolicy, err)
	}
	return policy, nil
}

func (p Policy) validate() error {
	if p.Schema != PolicySchema || p.CircuitID != CircuitID {
		return errors.New("schema or circuit id")
	}
	if p.ProductionAdmitted == nil || p.TestOnlyDenied == nil || p.RequiredProductionProvenance == nil {
		return errors.New("missing list")
	}
	if !equalStrings(p.RequiredProductionProvenance, requiredProvenance) {
		return errors.New("required provenance fields")
	}
	seen := map[string]string{}
	validKey := func(digest string, size int) error {
		if !hex64.MatchString(digest) {
			return errors.New("digest form")
		}
		if size <= 0 || size > MaxVerifyingKeyBytes {
			return errors.New("size")
		}
		return nil
	}
	for _, entry := range p.TestOnlyDenied {
		if err := validKey(entry.SHA256Hex, entry.SizeBytes); err != nil {
			return fmt.Errorf("denied key %w", err)
		}
		if _, dup := seen[entry.SHA256Hex]; dup {
			return errors.New("duplicate denied key")
		}
		seen[entry.SHA256Hex] = "denied"
		if len(entry.Sources) == 0 {
			return errors.New("denied key without sources")
		}
		sources := map[string]bool{}
		for _, source := range entry.Sources {
			if source == "" || sources[source] {
				return errors.New("denied key source")
			}
			sources[source] = true
		}
	}
	for _, entry := range p.ProductionAdmitted {
		if err := validKey(entry.SHA256Hex, entry.SizeBytes); err != nil {
			return fmt.Errorf("admitted key %w", err)
		}
		if kind, dup := seen[entry.SHA256Hex]; dup {
			if kind == "denied" {
				return errors.New("admitted key overlaps denied key")
			}
			return errors.New("duplicate admitted key")
		}
		seen[entry.SHA256Hex] = "admitted"
		if len(entry.Provenance) != len(requiredProvenance) {
			return errors.New("admitted key provenance fields")
		}
		for _, field := range requiredProvenance {
			if entry.Provenance[field] == "" {
				return fmt.Errorf("admitted key missing provenance %s", field)
			}
		}
	}
	return nil
}

// Classify evaluates the genesis triple (circuit id, VK hex, VK SHA-256 hex).
// All three empty is Absent. A partial, mismatched or non-canonical triple
// returns an error; a consistent key is denied, unadmitted or admitted.
func (p Policy) Classify(circuitID, verifyingKeyHex, verifyingKeySHA256 string) (Class, error) {
	if circuitID == "" && verifyingKeyHex == "" && verifyingKeySHA256 == "" {
		return Absent, nil
	}
	if circuitID == "" || verifyingKeyHex == "" || verifyingKeySHA256 == "" {
		return "", ErrPartialTriple
	}
	if circuitID != p.CircuitID {
		return "", ErrCircuitMismatch
	}
	if len(verifyingKeyHex) > 2*MaxVerifyingKeyBytes {
		return "", ErrOversizedKey
	}
	if !lowercaseEvenHex.MatchString(verifyingKeyHex) {
		return "", ErrNonCanonicalHex
	}
	if !hex64.MatchString(verifyingKeySHA256) {
		return "", ErrInvalidDigest
	}
	key, err := hex.DecodeString(verifyingKeyHex)
	if err != nil {
		return "", ErrNonCanonicalHex
	}
	digest := sha256.Sum256(key)
	if hex.EncodeToString(digest[:]) != verifyingKeySHA256 {
		return "", ErrDigestMismatch
	}
	for _, entry := range p.TestOnlyDenied {
		if entry.SHA256Hex == verifyingKeySHA256 && entry.SizeBytes == len(key) {
			return TestOnlyDenied, nil
		}
	}
	for _, entry := range p.ProductionAdmitted {
		if entry.SHA256Hex == verifyingKeySHA256 && entry.SizeBytes == len(key) {
			return ProductionAdmitted, nil
		}
	}
	return Unadmitted, nil
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// rejectDuplicateKeys walks the JSON token stream and fails on any object that
// repeats a key, which encoding/json would otherwise silently overwrite.
func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var walk func(depth int) error
	walk = func(depth int) error {
		if depth > 16 {
			return errors.New("maximum depth")
		}
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		delim, ok := token.(json.Delim)
		if !ok {
			return nil
		}
		switch delim {
		case '{':
			keys := []string{}
			for decoder.More() {
				keyToken, err := decoder.Token()
				if err != nil {
					return err
				}
				key, _ := keyToken.(string)
				keys = append(keys, key)
				if err := walk(depth + 1); err != nil {
					return err
				}
			}
			sort.Strings(keys)
			for i := 1; i < len(keys); i++ {
				if keys[i] == keys[i-1] {
					return fmt.Errorf("duplicate key %q", keys[i])
				}
			}
		case '[':
			for decoder.More() {
				if err := walk(depth + 1); err != nil {
					return err
				}
			}
		}
		_, err = decoder.Token()
		return err
	}
	return walk(0)
}
