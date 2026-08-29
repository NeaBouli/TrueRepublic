package ocievidence

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const fixtureSource = "0123456789abcdef0123456789abcdef01234567"

func TestValidRepeatedOCIArchives(t *testing.T) {
	dir, contract, _ := makeEvidence(t, "linux/amd64")
	report := VerifyDirectory(dir, contract)
	if !report.Valid {
		t.Fatalf("valid evidence rejected: %v", report.Violations)
	}
	if len(report.Images) != 2 || report.Images[0].ID != "daemon-linux-amd64" || len(report.Images[0].Layers) != 1 {
		t.Fatalf("verified image identities are incomplete: %#v", report.Images)
	}

	var bundle Bundle
	mustReadJSON(t, filepath.Join(dir, "oci-evidence.json"), &bundle)
	for _, target := range bundle.Targets {
		if target.Builds[0].SHA256 == target.Builds[1].SHA256 {
			t.Fatalf("fixture archives for %s should differ at tar level", target.ID)
		}
	}
}

func TestOCIArchiveDirectoryEntriesAccepted(t *testing.T) {
	dir, contract, bundle := makeEvidence(t, "linux/amd64")
	for targetIndex := range bundle.Targets {
		for buildIndex := range bundle.Targets[targetIndex].Builds {
			build := &bundle.Targets[targetIndex].Builds[buildIndex]
			archive := filepath.Join(dir, build.File)
			writeOCIArchiveWithDirectories(t, archive, "amd64", bundle.Targets[targetIndex].ID, buildIndex+1)
			build.SHA256 = mustHash(t, archive)
		}
	}
	writeJSON(t, filepath.Join(dir, "oci-evidence.json"), bundle)
	if report := VerifyDirectory(dir, contract); !report.Valid {
		t.Fatalf("canonical OCI directory entries rejected: %v", report.Violations)
	}
}

func TestAdversarialEvidenceRejected(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*testing.T, string, *Bundle)
	}{
		{"production claim", func(_ *testing.T, _ string, b *Bundle) { b.Claims.Production = true }},
		{"wrong platform", func(_ *testing.T, _ string, b *Bundle) { b.Platform = "linux/s390x" }},
		{"missing target", func(_ *testing.T, _ string, b *Bundle) { b.Targets = b.Targets[:1] }},
		{"duplicate target", func(_ *testing.T, _ string, b *Bundle) { b.Targets[1] = b.Targets[0] }},
		{"target order", func(_ *testing.T, _ string, b *Bundle) { b.Targets[0], b.Targets[1] = b.Targets[1], b.Targets[0] }},
		{"missing repetition", func(_ *testing.T, _ string, b *Bundle) { b.Targets[0].Builds = b.Targets[0].Builds[:1] }},
		{"malformed archive hash", func(_ *testing.T, _ string, b *Bundle) { b.Targets[0].Builds[0].SHA256 = "bad" }},
		{"unsafe archive path", func(_ *testing.T, _ string, b *Bundle) { b.Targets[0].Builds[0].File = "../escape.tar" }},
		{"swapped archive", func(_ *testing.T, _ string, b *Bundle) { b.Targets[0].Builds[0] = b.Targets[1].Builds[0] }},
		{"tampered archive", func(t *testing.T, dir string, b *Bundle) {
			file := filepath.Join(dir, b.Targets[0].Builds[0].File)
			handle, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = handle.Write([]byte("tamper")); err != nil {
				t.Fatal(err)
			}
			if err = handle.Close(); err != nil {
				t.Fatal(err)
			}
		}},
		{"unequal image identity", func(t *testing.T, dir string, b *Bundle) {
			build := &b.Targets[0].Builds[1]
			writeOCIArchive(t, filepath.Join(dir, build.File), "amd64", b.Targets[0].ID+"-different", 9, false)
			build.SHA256 = mustHash(t, filepath.Join(dir, build.File))
		}},
		{"unsafe tar entry", func(t *testing.T, dir string, b *Bundle) {
			build := &b.Targets[0].Builds[1]
			writeOCIArchive(t, filepath.Join(dir, build.File), "amd64", b.Targets[0].ID, 9, true)
			build.SHA256 = mustHash(t, filepath.Join(dir, build.File))
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir, contract, bundle := makeEvidence(t, "linux/amd64")
			test.mutate(t, dir, &bundle)
			writeJSON(t, filepath.Join(dir, "oci-evidence.json"), bundle)
			report := VerifyDirectory(dir, contract)
			if report.Valid {
				t.Fatal("adversarial evidence accepted")
			}
			if test.name == "unequal image identity" && (len(report.Images) < 2 || report.Images[0].Repetition != 1 || report.Images[1].Repetition != 2) {
				t.Fatalf("mismatch diagnostics omit per-repetition identities: %#v", report.Images)
			}
		})
	}
}

func TestSymlinkArchiveRejected(t *testing.T) {
	dir, contract, bundle := makeEvidence(t, "linux/amd64")
	archive := filepath.Join(dir, bundle.Targets[0].Builds[0].File)
	outside := filepath.Join(t.TempDir(), "outside.tar")
	data, err := os.ReadFile(archive)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(outside, data, 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(archive); err != nil {
		t.Fatal(err)
	}
	if err = os.Symlink(outside, archive); err != nil {
		t.Fatal(err)
	}
	if VerifyDirectory(dir, contract).Valid {
		t.Fatal("symlinked archive accepted")
	}
}

func TestStrictJSONBoundary(t *testing.T) {
	for name, raw := range map[string]string{
		"duplicate": `{"schema":"x","schema":"y"}`,
		"unknown":   `{"schema":"x","unknown":true}`,
		"trailing":  `{"schema":"x"}{}`,
		"deep":      strings.Repeat("[", maxJSONDepth+2) + "0" + strings.Repeat("]", maxJSONDepth+2),
	} {
		t.Run(name, func(t *testing.T) {
			var value struct {
				Schema string `json:"schema"`
			}
			if parseBytes([]byte(raw), &value) == nil {
				t.Fatal("invalid JSON accepted")
			}
		})
	}
	var value struct {
		Schema string `json:"schema"`
	}
	if parseBytes(make([]byte, MaxJSONBytes+1), &value) == nil {
		t.Fatal("oversized JSON accepted")
	}
	var claims Claims
	if parseBytes([]byte(`{"signed":false,"published":false}`), &claims) == nil {
		t.Fatal("incomplete claims accepted")
	}
}

func TestCommandBoundary(t *testing.T) {
	dir, contract, _ := makeEvidence(t, "linux/arm64")
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"verify", "--evidence", dir, "--contract", contract, "--output", "json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("valid command failed (%d): %s", code, stderr.String())
	}
	var report Report
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil || !report.Valid {
		t.Fatalf("invalid JSON report: %s", stdout.String())
	}
	if code := Run([]string{"verify", "--unknown"}, &stdout, &stderr); code != 2 {
		t.Fatalf("unknown flag exit code = %d, want 2", code)
	}
}

func TestContractRejectsUnconsumedBuildArguments(t *testing.T) {
	var contract Contract
	if err := parseFile(filepath.Join("..", "configs", "build", "reproducible-oci.json"), &contract); err != nil {
		t.Fatal(err)
	}
	contract.Targets[0].BuildArgs = map[string]string{"VERSION": "dev"}
	if err := validateContract(contract); err == nil {
		t.Fatal("daemon version drift accepted")
	}
	if err := parseFile(filepath.Join("..", "configs", "build", "reproducible-oci.json"), &contract); err != nil {
		t.Fatal(err)
	}
	contract.Targets[1].BuildArgs = map[string]string{"UNCONSUMED": "value"}
	if err := validateContract(contract); err == nil {
		t.Fatal("unconsumed client build argument accepted")
	}
}

func TestOCIConfigPlatformBinding(t *testing.T) {
	if err := validateOCIConfig([]byte(`{"architecture":"amd64","os":"linux"}`), "linux/amd64"); err != nil {
		t.Fatalf("matching OCI config rejected: %v", err)
	}
	for _, config := range [][]byte{
		[]byte(`{"architecture":"arm64","os":"linux"}`),
		[]byte(`{"architecture":"amd64","os":"windows"}`),
		[]byte(`{"architecture":"amd64"}`),
		[]byte(`[]`),
	} {
		if err := validateOCIConfig(config, "linux/amd64"); err == nil {
			t.Fatalf("mismatched OCI config accepted: %s", config)
		}
	}
}

func TestNormalizedTarPathBoundary(t *testing.T) {
	for _, raw := range []string{"blobs", "blobs/", "./blobs", "./blobs/"} {
		if got, err := normalizedTarPath(raw, true); err != nil || got != "blobs" {
			t.Fatalf("canonical directory path %q rejected: %q, %v", raw, got, err)
		}
	}
	for _, raw := range []string{"", "/blobs", "../blobs", "blobs//", "././blobs", `blobs\sha256`} {
		if _, err := normalizedTarPath(raw, true); err == nil {
			t.Fatalf("unsafe directory path %q accepted", raw)
		}
	}
	if _, err := normalizedTarPath("blobs/", false); err == nil {
		t.Fatal("regular file with trailing slash accepted")
	}
}

func makeEvidence(t *testing.T, platform string) (string, string, Bundle) {
	t.Helper()
	dir := t.TempDir()
	contract := filepath.Join("..", "configs", "build", "reproducible-oci.json")
	arch := strings.TrimPrefix(platform, "linux/")
	contractHash := mustHash(t, contract)
	bundle := Bundle{
		Schema:         Schema,
		SourceRef:      fixtureSource,
		ContractSHA256: contractHash,
		Platform:       platform,
		Claims:         Claims{present: true},
	}
	prefixes := []string{"daemon", "client-web"}
	for _, prefix := range prefixes {
		id := prefix + "-linux-" + arch
		target := TargetEvidence{ID: id}
		for repetition := 1; repetition <= 2; repetition++ {
			name := id + "-" + string(rune('a'+repetition-1)) + ".oci.tar"
			file := filepath.Join(dir, name)
			writeOCIArchive(t, file, arch, id, repetition, false)
			target.Builds = append(target.Builds, BuildEvidence{File: name, SHA256: mustHash(t, file)})
		}
		bundle.Targets = append(bundle.Targets, target)
	}
	writeJSON(t, filepath.Join(dir, "oci-evidence.json"), bundle)
	return dir, contract, bundle
}

func writeOCIArchive(t *testing.T, filePath, arch, identity string, variant int, unsafe bool) {
	writeOCIArchiveFixture(t, filePath, arch, identity, variant, unsafe, false)
}

func writeOCIArchiveWithDirectories(t *testing.T, filePath, arch, identity string, variant int) {
	writeOCIArchiveFixture(t, filePath, arch, identity, variant, false, true)
}

func writeOCIArchiveFixture(t *testing.T, filePath, arch, identity string, variant int, unsafe, directories bool) {
	t.Helper()
	layer := []byte("layer-" + identity)
	config := []byte(`{"architecture":"` + arch + `","os":"linux","rootfs":{"type":"layers","diff_ids":[]}}`)
	layerHash := byteHash(layer)
	configHash := byteHash(config)
	manifest := mustJSON(t, map[string]any{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.oci.image.manifest.v1+json",
		"config": map[string]any{
			"mediaType": "application/vnd.oci.image.config.v1+json", "digest": "sha256:" + configHash, "size": len(config),
		},
		"layers": []map[string]any{{
			"mediaType": "application/vnd.oci.image.layer.v1.tar", "digest": "sha256:" + layerHash, "size": len(layer),
		}},
	})
	manifestHash := byteHash(manifest)
	index := mustJSON(t, map[string]any{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.oci.image.index.v1+json",
		"manifests": []map[string]any{{
			"mediaType": "application/vnd.oci.image.manifest.v1+json", "digest": "sha256:" + manifestHash, "size": len(manifest), "platform": map[string]any{"architecture": arch, "os": "linux"},
		}},
	})
	entries := []struct {
		name string
		data []byte
	}{
		{"oci-layout", []byte(`{"imageLayoutVersion":"1.0.0"}`)},
		{"index.json", index},
		{"blobs/sha256/" + configHash, config},
		{"blobs/sha256/" + layerHash, layer},
		{"blobs/sha256/" + manifestHash, manifest},
	}
	if variant%2 == 0 {
		entries[0], entries[1] = entries[1], entries[0]
	}
	handle, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	writer := tar.NewWriter(handle)
	if directories {
		for _, name := range []string{"blobs/", "blobs/sha256/"} {
			if err = writer.WriteHeader(&tar.Header{Name: name, Mode: 0755, Typeflag: tar.TypeDir, ModTime: time.Unix(int64(variant), 0)}); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, entry := range entries {
		header := &tar.Header{Name: entry.name, Mode: 0600, Size: int64(len(entry.data)), ModTime: time.Unix(int64(variant), 0)}
		if err = writer.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if _, err = writer.Write(entry.data); err != nil {
			t.Fatal(err)
		}
	}
	if unsafe {
		if err = writer.WriteHeader(&tar.Header{Name: "unsafe-link", Typeflag: tar.TypeSymlink, Linkname: "../outside"}); err != nil {
			t.Fatal(err)
		}
	}
	if err = writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err = handle.Close(); err != nil {
		t.Fatal(err)
	}
}

func byteHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func mustHash(t *testing.T, path string) string {
	t.Helper()
	hash, err := hashRegularFile(path, maxArchiveBytes)
	if err != nil {
		t.Fatal(err)
	}
	return hash
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func writeJSON(t *testing.T, path string, value any) {
	t.Helper()
	if err := os.WriteFile(path, mustJSON(t, value), 0600); err != nil {
		t.Fatal(err)
	}
}

func mustReadJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, value); err != nil {
		t.Fatal(err)
	}
}
