package releaseevidence

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testSource = "0123456789abcdef0123456789abcdef01234567"

func TestValidBundleAndAdversarialBindings(t *testing.T) {
	dir, bundle := makeBundle(t)
	build := filepath.Join("..", "configs", "build", "deterministic-linux-daemon.json")
	tools := filepath.Join("..", "configs", "release", "tool-platform.json")
	if report := VerifyDirectory(dir, build, tools); !report.Valid {
		t.Fatalf("valid fixture rejected: %v", report.Violations)
	}

	tests := []struct {
		name   string
		mutate func(*Bundle)
	}{
		{"source mismatch", func(b *Bundle) { b.SourceRef = strings.Repeat("f", 40) }},
		{"tool mismatch", func(b *Bundle) { b.Tools.NPM = "latest" }},
		{"production claim", func(b *Bundle) { b.Claims.Production = true }},
		{"missing target", func(b *Bundle) { b.Targets = b.Targets[:1] }},
		{"extra target", func(b *Bundle) { b.Targets = append(b.Targets, b.Targets[0]) }},
		{"target order", func(b *Bundle) { b.Targets[0], b.Targets[1] = b.Targets[1], b.Targets[0] }},
		{"SBOM order", func(b *Bundle) { b.SBOMs[0], b.SBOMs[1] = b.SBOMs[1], b.SBOMs[0] }},
		{"artifact digest", func(b *Bundle) { b.Targets[0].SHA256 = strings.Repeat("0", 64) }},
		{"path traversal", func(b *Bundle) { b.SBOMs[0].File = "../go.cdx.json" }},
		{"backslash", func(b *Bundle) { b.SBOMs[0].File = `..\go.cdx.json` }},
		{"absolute", func(b *Bundle) { b.SBOMs[0].File = "/tmp/go.cdx.json" }},
		{"provenance digest", func(b *Bundle) { b.Provenance.SHA256 = strings.Repeat("0", 64) }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			changed := bundle
			changed.Targets = append([]Target(nil), bundle.Targets...)
			changed.SBOMs = append([]SBOM(nil), bundle.SBOMs...)
			tc.mutate(&changed)
			writeJSON(t, filepath.Join(dir, "release-evidence.json"), changed)
			if VerifyDirectory(dir, build, tools).Valid {
				t.Fatal("adversarial bundle accepted")
			}
			writeJSON(t, filepath.Join(dir, "release-evidence.json"), bundle)
		})
	}

	t.Run("cross-target source-date epoch mismatch", func(t *testing.T) {
		metadataPath := filepath.Join(dir, "linux-arm64.build-metadata.json")
		var value map[string]any
		contents, err := os.ReadFile(metadataPath)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(contents, &value); err != nil {
			t.Fatal(err)
		}
		value["source_date_epoch"] = 456
		writeJSON(t, metadataPath, value)
		if VerifyDirectory(dir, build, tools).Valid {
			t.Fatal("cross-target source-date epoch mismatch accepted")
		}
		value["source_date_epoch"] = 123
		writeJSON(t, metadataPath, value)
	})

	t.Run("symlink escape", func(t *testing.T) {
		outside := filepath.Join(t.TempDir(), "outside.json")
		os.WriteFile(outside, []byte(`{"bomFormat":"CycloneDX","specVersion":"1.6","components":[]}`), 0600)
		os.Remove(filepath.Join(dir, "go.cdx.json"))
		if err := os.Symlink(outside, filepath.Join(dir, "go.cdx.json")); err != nil {
			t.Fatal(err)
		}
		if VerifyDirectory(dir, build, tools).Valid {
			t.Fatal("symlink escape accepted")
		}
	})
}

func TestStrictJSONBoundary(t *testing.T) {
	valid := `{"schema":"x"}`
	for name, raw := range map[string]string{
		"duplicate": `{"schema":"x","schema":"y"}`,
		"unknown":   `{"schema":"x","unknown":true}`,
		"trailing":  valid + ` {}`,
		"deep":      strings.Repeat("[", 34) + "0" + strings.Repeat("]", 34),
	} {
		t.Run(name, func(t *testing.T) {
			var v struct {
				Schema string `json:"schema"`
			}
			if parseBytes([]byte(raw), &v) == nil {
				t.Fatal("invalid JSON accepted")
			}
		})
	}
	var v struct {
		Schema string `json:"schema"`
	}
	if parseBytes(make([]byte, MaxJSONBytes+1), &v) == nil {
		t.Fatal("oversized JSON accepted")
	}
	var claims Claims
	if parseBytes([]byte(`{"signed":false,"published":false}`), &claims) == nil {
		t.Fatal("incomplete status claims accepted")
	}
}

func makeBundle(t *testing.T) (string, Bundle) {
	t.Helper()
	dir := t.TempDir()
	buildPath := filepath.Join("..", "configs", "build", "deterministic-linux-daemon.json")
	toolPath := filepath.Join("..", "configs", "release", "tool-platform.json")
	bh, _ := hashFile(buildPath)
	th, _ := hashFile(toolPath)
	b := Bundle{Schema: Schema, SourceRef: testSource, BuildContractSHA256: bh, ToolContractSHA256: th, Tools: exactTools(), Claims: Claims{present: true}}
	defs := []BuildTarget{{"linux-amd64", "linux", "amd64", "ubuntu-24.04", "x86_64", "truerepublicd-linux-amd64"}, {"linux-arm64", "linux", "arm64", "ubuntu-24.04-arm", "aarch64", "truerepublicd-linux-arm64"}}
	for i, d := range defs {
		data := []byte("artifact-" + d.ID)
		artifactPath := filepath.Join(dir, d.Artifact)
		os.WriteFile(artifactPath, data, 0600)
		h := sum(data)
		checksum := d.ID + ".CHECKSUMS.sha256"
		metadataName := d.ID + ".build-metadata.json"
		os.WriteFile(filepath.Join(dir, checksum), []byte(h+"  "+d.Artifact+"\n"), 0600)
		m := map[string]any{"schema": BuildEvidenceSchema, "contract_schema": BuildSchema, "contract_sha256": bh, "source_ref": testSource, "target": d.ID, "ci_runner": d.CIRunner, "runner_arch": d.RunnerArch, "artifact": d.Artifact, "sha256": h, "reproducible_pair_sha256": []string{h, h}, "go_version": "1.26.6", "cgo_enabled": "1", "source_date_epoch": 123, "build_flags": map[string]any{"trimpath": true, "buildvcs": false, "mod": "readonly", "buildid": "", "linker_build_id": "none", "version_variable": "main.version"}}
		writeJSON(t, filepath.Join(dir, metadataName), m)
		b.Targets = append(b.Targets, Target{d.ID, d.Artifact, h, checksum, metadataName})
		_ = i
	}
	for _, component := range []string{"go", "client"} {
		name := component + ".cdx.json"
		raw := []byte(`{"bomFormat":"CycloneDX","components":[{"name":"` + component + `"}],"specVersion":"1.6"}`)
		os.WriteFile(filepath.Join(dir, name), raw, 0600)
		b.SBOMs = append(b.SBOMs, SBOM{component, name, sum(raw)})
	}
	p := Provenance{Schema: ProvenanceSchema, SourceRef: testSource, BuildContractSHA256: bh, ToolContractSHA256: th, Claims: Claims{present: true}}
	for _, x := range b.Targets {
		p.Targets = append(p.Targets, ProvenanceEntry{x.ID, x.SHA256})
	}
	for _, x := range b.SBOMs {
		p.SBOMs = append(p.SBOMs, ProvenanceEntry{x.Component, x.SHA256})
	}
	writeJSON(t, filepath.Join(dir, "provenance.json"), p)
	ph, _ := hashFile(filepath.Join(dir, "provenance.json"))
	b.Provenance = FileDigest{"provenance.json", ph}
	writeJSON(t, filepath.Join(dir, "release-evidence.json"), b)
	return dir, b
}
func sum(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
}
