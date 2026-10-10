package genesisevidence

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
	"testing"

	"truerepublic/x/truedemocracy/zkpadmission"
)

// GH300B4B: the zkp-verifying-key check passes only for an explicitly recorded
// absent key; the frozen TEST-ONLY key, unadmitted keys and any partial,
// mismatched or non-canonical triple fail closed.

func frozenTestOnlyVK(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("../x/truedemocracy/testdata/zkp/membership_v2.vk")
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func vkDigest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func evidenceWithTD(t *testing.T, edit func(td map[string]any)) Evidence {
	t.Helper()
	m, g := objects(t)
	td := g["app_state"].(map[string]any)["truedemocracy"].(map[string]any)
	edit(td)
	mb, gb := encoded(t, m, g)
	return Verify(mb, gb)
}

func setTriple(circuit, vkHex, digest string) func(map[string]any) {
	return func(td map[string]any) {
		td["zkp_circuit_id"] = circuit
		td["verifying_key_hex"] = vkHex
		td["verifying_key_sha256"] = digest
	}
}

func TestZKPVerifyingKeyAbsentPassesWithExplicitClassification(t *testing.T) {
	for name, edit := range map[string]func(map[string]any){
		"fields omitted": func(map[string]any) {},
		"fields empty":   setTriple("", "", ""),
	} {
		t.Run(name, func(t *testing.T) {
			e := evidenceWithTD(t, edit)
			c := check(e, "zkp-verifying-key")
			if !e.Valid || !c.Pass || c.Classification != string(zkpadmission.Absent) || len(c.Violations) != 0 {
				t.Fatalf("absent key not recorded as passing absent: valid=%v %+v", e.Valid, c)
			}
		})
	}
}

func TestZKPVerifyingKeyRejectsDeniedUnadmittedAndMalformedTriples(t *testing.T) {
	frozen := frozenTestOnlyVK(t)
	frozenHex := hex.EncodeToString(frozen)
	other := append([]byte(nil), frozen...)
	other[0] ^= 1
	circuit := zkpadmission.CircuitID

	tests := []struct {
		name           string
		edit           func(map[string]any)
		violation      string
		classification string
	}{
		{"frozen test-only key", setTriple(circuit, frozenHex, vkDigest(frozen)), "test-only-verifying-key", "test-only-denied"},
		{"unadmitted key", setTriple(circuit, hex.EncodeToString(other), vkDigest(other)), "unadmitted-verifying-key", "unadmitted"},
		{"key without digest", func(td map[string]any) {
			td["zkp_circuit_id"] = circuit
			td["verifying_key_hex"] = frozenHex
		}, "partial-verifying-key-triple", ""},
		{"digest without key", func(td map[string]any) { td["verifying_key_sha256"] = vkDigest(frozen) }, "partial-verifying-key-triple", ""},
		{"circuit mismatch", setTriple("truerepublic/membership-vote/v1", frozenHex, vkDigest(frozen)), "verifying-key-circuit-mismatch", ""},
		{"digest mismatch", setTriple(circuit, frozenHex, vkDigest(other)), "verifying-key-digest-mismatch", ""},
		{"uppercase key hex", setTriple(circuit, strings.ToUpper(frozenHex), vkDigest(frozen)), "non-canonical-verifying-key-hex", ""},
		{"uppercase digest", setTriple(circuit, frozenHex, strings.ToUpper(vkDigest(frozen))), "invalid-verifying-key-digest", ""},
		{"oversized key", setTriple(circuit, strings.Repeat("00", zkpadmission.MaxVerifyingKeyBytes+1), vkDigest(frozen)), "oversized-verifying-key", ""},
		{"non-string field", func(td map[string]any) {
			setTriple(circuit, frozenHex, vkDigest(frozen))(td)
			td["verifying_key_hex"] = []any{frozenHex}
		}, "invalid-verifying-key-field", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := evidenceWithTD(t, tc.edit)
			c := check(e, "zkp-verifying-key")
			if e.Valid || c.Pass || !hasViolation(c, tc.violation) || c.Classification != tc.classification {
				t.Fatalf("valid=%v check=%+v, want violation %q classification %q", e.Valid, c, tc.violation, tc.classification)
			}
		})
	}
}

func TestZKPVerifyingKeyNotEvaluatedCarriesNoClassification(t *testing.T) {
	m, g := objects(t)
	m.TotalSupplyUPNYX = "not-a-number"
	mb, gb := encoded(t, m, g)
	c := check(Verify(mb, gb), "zkp-verifying-key")
	if c.Pass || !hasViolation(c, "not-evaluated") || c.Classification != "" {
		t.Fatalf("skipped check must fail closed without classification: %+v", c)
	}

	m, g = objects(t)
	delete(g["app_state"].(map[string]any), "truedemocracy")
	mb, gb = encoded(t, m, g)
	c = check(Verify(mb, gb), "zkp-verifying-key")
	if c.Pass || !hasViolation(c, "missing-truedemocracy-state") || c.Classification != "" {
		t.Fatalf("missing truedemocracy state must fail closed: %+v", c)
	}
}
