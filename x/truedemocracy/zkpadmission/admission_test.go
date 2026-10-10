package zkpadmission

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const repoRoot = "../../.."

func readRepoFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot, path))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

func digestHex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func TestPolicyConfigMatchesCompiledPolicy(t *testing.T) {
	policy, err := ParsePolicy(readRepoFile(t, "configs/security/zkp-vk-admission.json"))
	if err != nil {
		t.Fatalf("repository policy rejected: %v", err)
	}
	if !reflect.DeepEqual(policy, DefaultPolicy()) {
		t.Fatalf("config and compiled policy differ:\nconfig   %+v\ncompiled %+v", policy, DefaultPolicy())
	}
	if len(policy.ProductionAdmitted) != 0 {
		t.Fatal("production allowlist must stay empty until a reviewed ceremony entry exists")
	}
}

func TestDeniedKeyIsTheFrozenTestOnlyArtifact(t *testing.T) {
	vk := readRepoFile(t, "x/truedemocracy/testdata/zkp/membership_v2.vk")
	if digestHex(vk) != FrozenTestOnlySHA256 || len(vk) != FrozenTestOnlySizeBytes {
		t.Fatalf("testdata VK digest/size drifted: %s/%d", digestHex(vk), len(vk))
	}

	var browser struct {
		Classification string `json:"classification"`
		VerifyingKey   struct {
			SizeBytes int    `json:"size_bytes"`
			SHA256Hex string `json:"sha256_hex"`
		} `json:"verifying_key"`
	}
	if err := json.Unmarshal(readRepoFile(t, "configs/security/zkp-browser-artifacts.json"), &browser); err != nil {
		t.Fatal(err)
	}
	if browser.Classification != "TEST-ONLY SINGLE-PARTY TOXIC WASTE" ||
		browser.VerifyingKey.SHA256Hex != FrozenTestOnlySHA256 || browser.VerifyingKey.SizeBytes != FrozenTestOnlySizeBytes {
		t.Fatalf("browser manifest VK is not the denied test-only key: %+v", browser)
	}

	var fixture struct {
		Classification string `json:"classification"`
		Artifacts      []struct {
			Path      string `json:"path"`
			SizeBytes int    `json:"size_bytes"`
			SHA256    string `json:"sha256"`
		} `json:"artifacts"`
	}
	if err := json.Unmarshal(readRepoFile(t, "x/truedemocracy/testdata/zkp/manifest.json"), &fixture); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, artifact := range fixture.Artifacts {
		if artifact.Path == "membership_v2.vk" {
			found = artifact.SHA256 == FrozenTestOnlySHA256 && artifact.SizeBytes == FrozenTestOnlySizeBytes
		}
	}
	if fixture.Classification != "TEST-ONLY SINGLE-PARTY TOXIC WASTE" || !found {
		t.Fatal("fixture manifest VK is not the denied test-only key")
	}

	if !strings.Contains(string(readRepoFile(t, "internal/zkpprover/prover.go")), `VerifyingArtifactSHA256 = "`+FrozenTestOnlySHA256+`"`) {
		t.Fatal("WASM prover verifying-key pin is not the denied test-only key")
	}
}

func syntheticProvenance() map[string]string {
	provenance := map[string]string{}
	for _, field := range requiredProvenance {
		provenance[field] = "synthetic-test-value"
	}
	return provenance
}

func encodePolicy(t *testing.T, p Policy) []byte {
	t.Helper()
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestParsePolicyRejectsAdversarialInputs(t *testing.T) {
	// Synthetic digest used only inside this test; never part of the config.
	synthetic := strings.Repeat("ab", 32)
	base := string(readRepoFile(t, "configs/security/zkp-vk-admission.json"))
	mutate := func(fn func(*Policy)) []byte {
		p := DefaultPolicy()
		fn(&p)
		return encodePolicy(t, p)
	}
	tests := map[string][]byte{
		"empty":                    nil,
		"unknown field":            []byte(strings.Replace(base, `"schema":`, `"activation":true,"schema":`, 1)),
		"duplicate key":            []byte(strings.Replace(base, `"schema":`, `"circuit_id":"x","schema":`, 1)),
		"trailing data":            []byte(base + `{}`),
		"null production list":     []byte(strings.Replace(base, `"production_admitted": []`, `"production_admitted": null`, 1)),
		"wrong schema":             mutate(func(p *Policy) { p.Schema = "truerepublic/zkp-vk-admission/v2" }),
		"wrong circuit":            mutate(func(p *Policy) { p.CircuitID = "truerepublic/membership-vote/v3" }),
		"uppercase digest":         mutate(func(p *Policy) { p.TestOnlyDenied[0].SHA256Hex = strings.ToUpper(FrozenTestOnlySHA256) }),
		"short digest":             mutate(func(p *Policy) { p.TestOnlyDenied[0].SHA256Hex = FrozenTestOnlySHA256[:62] }),
		"zero size":                mutate(func(p *Policy) { p.TestOnlyDenied[0].SizeBytes = 0 }),
		"oversized":                mutate(func(p *Policy) { p.TestOnlyDenied[0].SizeBytes = MaxVerifyingKeyBytes + 1 }),
		"duplicate denied":         mutate(func(p *Policy) { p.TestOnlyDenied = append(p.TestOnlyDenied, p.TestOnlyDenied[0]) }),
		"denied without sources":   mutate(func(p *Policy) { p.TestOnlyDenied[0].Sources = []string{} }),
		"duplicate source":         mutate(func(p *Policy) { p.TestOnlyDenied[0].Sources = []string{"a", "a"} }),
		"empty source":             mutate(func(p *Policy) { p.TestOnlyDenied[0].Sources = []string{""} }),
		"weakened provenance list": mutate(func(p *Policy) { p.RequiredProductionProvenance = p.RequiredProductionProvenance[1:] }),
		"admitted overlaps denied": mutate(func(p *Policy) {
			p.ProductionAdmitted = []AdmittedKey{{SHA256Hex: FrozenTestOnlySHA256, SizeBytes: FrozenTestOnlySizeBytes, Provenance: syntheticProvenance()}}
		}),
		"duplicate admitted": mutate(func(p *Policy) {
			entry := AdmittedKey{SHA256Hex: synthetic, SizeBytes: 460, Provenance: syntheticProvenance()}
			p.ProductionAdmitted = []AdmittedKey{entry, entry}
		}),
		"admitted missing provenance": mutate(func(p *Policy) {
			provenance := syntheticProvenance()
			delete(provenance, "random_beacon")
			p.ProductionAdmitted = []AdmittedKey{{SHA256Hex: synthetic, SizeBytes: 460, Provenance: provenance}}
		}),
		"admitted empty provenance value": mutate(func(p *Policy) {
			provenance := syntheticProvenance()
			provenance["review_reference"] = ""
			p.ProductionAdmitted = []AdmittedKey{{SHA256Hex: synthetic, SizeBytes: 460, Provenance: provenance}}
		}),
		"admitted extra provenance field": mutate(func(p *Policy) {
			provenance := syntheticProvenance()
			delete(provenance, "ceremony_id")
			provenance["exception"] = "allow"
			p.ProductionAdmitted = []AdmittedKey{{SHA256Hex: synthetic, SizeBytes: 460, Provenance: provenance}}
		}),
	}
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := ParsePolicy(data); err == nil {
				t.Fatal("policy accepted")
			}
		})
	}
}

func TestClassify(t *testing.T) {
	policy := DefaultPolicy()
	frozen := readRepoFile(t, "x/truedemocracy/testdata/zkp/membership_v2.vk")
	frozenHex := hex.EncodeToString(frozen)
	other := append([]byte(nil), frozen...)
	other[len(other)-1] ^= 1
	otherHex := hex.EncodeToString(other)

	type want struct {
		class Class
		err   error
	}
	tests := []struct {
		name                string
		circuit, vk, digest string
		want                want
	}{
		{"absent", "", "", "", want{class: Absent}},
		{"circuit only", CircuitID, "", "", want{err: ErrPartialTriple}},
		{"key only", "", frozenHex, "", want{err: ErrPartialTriple}},
		{"digest only", "", "", FrozenTestOnlySHA256, want{err: ErrPartialTriple}},
		{"missing digest", CircuitID, frozenHex, "", want{err: ErrPartialTriple}},
		{"circuit mismatch", "truerepublic/membership-vote/v1", frozenHex, FrozenTestOnlySHA256, want{err: ErrCircuitMismatch}},
		{"uppercase key hex", CircuitID, strings.ToUpper(frozenHex), FrozenTestOnlySHA256, want{err: ErrNonCanonicalHex}},
		{"odd key hex", CircuitID, frozenHex + "0", FrozenTestOnlySHA256, want{err: ErrNonCanonicalHex}},
		{"non-hex key", CircuitID, "zz" + frozenHex[2:], FrozenTestOnlySHA256, want{err: ErrNonCanonicalHex}},
		{"oversized key", CircuitID, strings.Repeat("00", MaxVerifyingKeyBytes+1), FrozenTestOnlySHA256, want{err: ErrOversizedKey}},
		{"uppercase digest", CircuitID, frozenHex, strings.ToUpper(FrozenTestOnlySHA256), want{err: ErrInvalidDigest}},
		{"short digest", CircuitID, frozenHex, FrozenTestOnlySHA256[:63], want{err: ErrInvalidDigest}},
		{"digest of other key", CircuitID, frozenHex, digestHex(other), want{err: ErrDigestMismatch}},
		{"trailing byte", CircuitID, frozenHex + "00", FrozenTestOnlySHA256, want{err: ErrDigestMismatch}},
		{"frozen test-only key", CircuitID, frozenHex, FrozenTestOnlySHA256, want{class: TestOnlyDenied}},
		{"unknown key", CircuitID, otherHex, digestHex(other), want{class: Unadmitted}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			class, err := policy.Classify(tc.circuit, tc.vk, tc.digest)
			if class != tc.want.class || !errors.Is(err, tc.want.err) || (tc.want.err == nil && err != nil) {
				t.Fatalf("Classify = (%q, %v), want (%q, %v)", class, err, tc.want.class, tc.want.err)
			}
		})
	}
}

func TestProductionAdmissionRequiresAPolicyEntry(t *testing.T) {
	frozen := readRepoFile(t, "x/truedemocracy/testdata/zkp/membership_v2.vk")
	key := append([]byte(nil), frozen...)
	key[0] ^= 1
	keyHex, keyDigest := hex.EncodeToString(key), digestHex(key)

	if class, _ := DefaultPolicy().Classify(CircuitID, keyHex, keyDigest); class != Unadmitted {
		t.Fatalf("compiled policy classified an unlisted key as %q", class)
	}
	// A synthetic in-test policy proves the admitted branch; the repository
	// policy carries no production entry.
	synthetic := DefaultPolicy()
	synthetic.ProductionAdmitted = []AdmittedKey{{SHA256Hex: keyDigest, SizeBytes: len(key), Provenance: syntheticProvenance()}}
	parsed, err := ParsePolicy(encodePolicy(t, synthetic))
	if err != nil {
		t.Fatalf("synthetic policy rejected: %v", err)
	}
	if class, err := parsed.Classify(CircuitID, keyHex, keyDigest); class != ProductionAdmitted || err != nil {
		t.Fatalf("admitted key classified as (%q, %v)", class, err)
	}
	if class, _ := parsed.Classify(CircuitID, hex.EncodeToString(frozen), FrozenTestOnlySHA256); class != TestOnlyDenied {
		t.Fatalf("denied key classified as %q", class)
	}
}
