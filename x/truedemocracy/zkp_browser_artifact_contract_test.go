package truedemocracy

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

const (
	zkpBrowserManifestSchema = "truerepublic/zkp-artifact-manifest/v1"
	zkpTestClassification    = "TEST-ONLY SINGLE-PARTY TOXIC WASTE"
	zkpBrowserWASMSize       = 18778260
	zkpBrowserWASMSHA256     = "d2cf62c436055de543562347f877554c631fbffbdafa26155c96c07d6dc67d66"
)

type zkpBrowserArtifactDescriptor struct {
	Path      string `json:"path"`
	SizeBytes int    `json:"size_bytes"`
	SHA256Hex string `json:"sha256_hex"`
}

type zkpBrowserArtifactManifest struct {
	Schema            string                       `json:"schema"`
	CircuitID         string                       `json:"circuit_id"`
	Classification    string                       `json:"classification"`
	ProductionAllowed bool                         `json:"production_allowed"`
	ConstraintSystem  zkpBrowserArtifactDescriptor `json:"constraint_system"`
	ProvingKey        zkpBrowserArtifactDescriptor `json:"proving_key"`
	VerifyingKey      zkpBrowserArtifactDescriptor `json:"verifying_key"`
	WASM              zkpBrowserArtifactDescriptor `json:"wasm"`
}

// TestZKPBrowserArtifactRepositoryContract prevents the maintained browser
// harness from drifting away from the frozen, explicitly forge-capable test
// fixture. The WASM bytes are built and digest-checked by the browser gate.
func TestZKPBrowserArtifactRepositoryContract(t *testing.T) {
	manifestPath := filepath.Join("..", "..", "configs", "security", "zkp-browser-artifacts.json")
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest zkpBrowserArtifactManifest
	if err := strictJSON(manifestBytes, &manifest); err != nil {
		t.Fatalf("strict browser artifact manifest decode: %v", err)
	}
	if manifest.Schema != zkpBrowserManifestSchema ||
		manifest.CircuitID != MembershipCircuitID ||
		manifest.Classification != zkpTestClassification ||
		manifest.ProductionAllowed {
		t.Fatalf("unsafe browser artifact manifest identity: %+v", manifest)
	}

	wantFixtures := map[string]string{
		"__zkp/membership_v2.cs": "membership_v2.cs",
		"__zkp/membership_v2.pk": "membership_v2.pk",
		"__zkp/membership_v2.vk": "membership_v2.vk",
	}
	descriptors := []zkpBrowserArtifactDescriptor{
		manifest.ConstraintSystem,
		manifest.ProvingKey,
		manifest.VerifyingKey,
		manifest.WASM,
	}
	seenPaths := make(map[string]struct{}, len(descriptors))
	seenDigests := make(map[string]struct{}, len(descriptors))
	for _, descriptor := range descriptors {
		if _, duplicate := seenPaths[descriptor.Path]; duplicate {
			t.Fatalf("duplicate browser artifact path %q", descriptor.Path)
		}
		if _, duplicate := seenDigests[descriptor.SHA256Hex]; duplicate {
			t.Fatalf("duplicate browser artifact digest %q", descriptor.SHA256Hex)
		}
		seenPaths[descriptor.Path] = struct{}{}
		seenDigests[descriptor.SHA256Hex] = struct{}{}
	}

	for browserPath, fixtureName := range wantFixtures {
		var descriptor zkpBrowserArtifactDescriptor
		for _, candidate := range descriptors {
			if candidate.Path == browserPath {
				descriptor = candidate
				break
			}
		}
		if descriptor.Path == "" {
			t.Fatalf("browser manifest omits %s", browserPath)
		}
		fixture, err := os.ReadFile(filepath.Join(zkpFixtureDir, fixtureName))
		if err != nil {
			t.Fatal(err)
		}
		digest := sha256.Sum256(fixture)
		if descriptor.SizeBytes != len(fixture) || descriptor.SHA256Hex != hex.EncodeToString(digest[:]) {
			t.Fatalf("browser descriptor for %s does not bind the frozen fixture", browserPath)
		}
	}

	if manifest.WASM.Path != "__zkp/zkp-prover.wasm" ||
		manifest.WASM.SizeBytes != zkpBrowserWASMSize ||
		manifest.WASM.SHA256Hex != zkpBrowserWASMSHA256 {
		t.Fatalf("browser WASM descriptor drifted: %+v", manifest.WASM)
	}
}
