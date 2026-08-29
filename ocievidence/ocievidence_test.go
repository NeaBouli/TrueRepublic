package ocievidence

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
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

func TestLayerDiffContentAndMetadataAreDigestOnly(t *testing.T) {
	firstBody := []byte("private-marker-one")
	secondBody := []byte("private-marker-two")
	dir, contract := makeLayerDiagnosticEvidence(t, func(repetition int) ([]byte, string) {
		body := firstBody
		if repetition == 2 {
			body = secondBody
		}
		return makeLayerTar(t, []layerEntryFixture{{name: "var/lib/app/state", mode: 0640, body: body}}, false), layerMediaTar
	})
	report := VerifyDirectory(dir, contract)
	if report.Valid || len(report.LayerDiffs) != 1 {
		t.Fatalf("layer mismatch diagnostic missing: %#v", report)
	}
	layer := report.LayerDiffs[0]
	if layer.Target != "daemon-linux-amd64" || layer.LayerIndex != 0 || layer.Truncated || len(layer.Entries) != 1 {
		t.Fatalf("unexpected layer diagnostic: %#v", layer)
	}
	entry := layer.Entries[0]
	if entry.Path != "var/lib/app/state" || entry.Change != "modified" || entry.First == nil || entry.Second == nil {
		t.Fatalf("unexpected entry diagnostic: %#v", entry)
	}
	if entry.First.ContentSHA256 == entry.Second.ContentSHA256 {
		t.Fatal("content-only mismatch retained equal content digests")
	}
	if entry.First.MetadataSHA256 != entry.Second.MetadataSHA256 {
		t.Fatal("content-only mismatch changed metadata digest")
	}
	raw := mustJSON(t, report)
	if bytes.Contains(raw, firstBody) || bytes.Contains(raw, secondBody) {
		t.Fatalf("raw layer content leaked into report: %s", raw)
	}

	dir, contract = makeLayerDiagnosticEvidence(t, func(repetition int) ([]byte, string) {
		mode := int64(0644)
		if repetition == 2 {
			mode = 0600
		}
		return makeLayerTar(t, []layerEntryFixture{{name: "etc/app.conf", mode: mode, body: []byte("same")}}, false), layerMediaTar
	})
	report = VerifyDirectory(dir, contract)
	entry = report.LayerDiffs[0].Entries[0]
	if entry.First.ContentSHA256 != entry.Second.ContentSHA256 || entry.First.MetadataSHA256 == entry.Second.MetadataSHA256 || entry.First.Mode != "0644" || entry.Second.Mode != "0600" {
		t.Fatalf("metadata-only mismatch not isolated: %#v", entry)
	}
}

func TestLayerDiffGzipOrderingAndTruncation(t *testing.T) {
	oldDiffs := maxLayerDiffs
	maxLayerDiffs = 2
	defer func() { maxLayerDiffs = oldDiffs }()
	dir, contract := makeLayerDiagnosticEvidence(t, func(repetition int) ([]byte, string) {
		suffix := byte('a' + repetition - 1)
		entries := []layerEntryFixture{
			{name: "z-last", mode: 0644, body: []byte{suffix}},
			{name: "a-first", mode: 0644, body: []byte{suffix}},
			{name: "m-middle", mode: 0644, body: []byte{suffix}},
		}
		return makeLayerTar(t, entries, true), layerMediaTarGzip
	})
	report := VerifyDirectory(dir, contract)
	if report.Valid || len(report.LayerDiffs) != 1 {
		t.Fatalf("gzip layer mismatch diagnostic missing: %#v", report)
	}
	layer := report.LayerDiffs[0]
	if !layer.Truncated || len(layer.Entries) != 2 || layer.Entries[0].Path != "a-first" || layer.Entries[1].Path != "m-middle" {
		t.Fatalf("layer diff order/bound is not deterministic: %#v", layer)
	}
}

func TestLayerDiffAddedRemovedAndEqual(t *testing.T) {
	dir, contract := makeLayerDiagnosticEvidence(t, func(repetition int) ([]byte, string) {
		entries := []layerEntryFixture{{name: "shared", mode: 0644, body: []byte("same")}}
		if repetition == 1 {
			entries = append(entries, layerEntryFixture{name: "removed", mode: 0644, body: []byte("old")})
		} else {
			entries = append(entries, layerEntryFixture{name: "added", mode: 0644, body: []byte("new")})
		}
		return makeLayerTar(t, entries, false), layerMediaTar
	})
	report := VerifyDirectory(dir, contract)
	if len(report.LayerDiffs) != 1 || len(report.LayerDiffs[0].Entries) != 2 {
		t.Fatalf("added/removed paths missing: %#v", report.LayerDiffs)
	}
	if first := report.LayerDiffs[0].Entries[0]; first.Path != "added" || first.Change != "added" || first.First != nil || first.Second == nil {
		t.Fatalf("added path report invalid: %#v", first)
	}
	if second := report.LayerDiffs[0].Entries[1]; second.Path != "removed" || second.Change != "removed" || second.First == nil || second.Second != nil {
		t.Fatalf("removed path report invalid: %#v", second)
	}

	equalLayer := makeLayerTar(t, []layerEntryFixture{{name: "equal", mode: 0644, body: []byte("same")}}, false)
	dir, contract = makeLayerDiagnosticEvidence(t, func(int) ([]byte, string) { return equalLayer, layerMediaTar })
	report = VerifyDirectory(dir, contract)
	if !report.Valid || len(report.LayerDiffs) != 0 {
		t.Fatalf("equal layers produced diagnostics: %#v", report)
	}
}

func TestLayerDiffAdversarialInputsFailClosed(t *testing.T) {
	tests := []struct {
		name       string
		configure  func() func()
		layer      func(*testing.T, int) ([]byte, string)
		wantReason string
	}{
		{
			name: "unsafe path",
			layer: func(t *testing.T, repetition int) ([]byte, string) {
				return makeLayerTar(t, []layerEntryFixture{{name: "../escape", mode: 0644, body: []byte{byte(repetition)}}}, false), layerMediaTar
			},
			wantReason: "unsafe layer path",
		},
		{
			name: "duplicate path",
			layer: func(t *testing.T, repetition int) ([]byte, string) {
				return makeLayerTar(t, []layerEntryFixture{{name: "dup", mode: 0644, body: []byte("a")}, {name: "dup", mode: 0644, body: []byte{byte(repetition)}}}, false), layerMediaTar
			},
			wantReason: "duplicate layer path",
		},
		{
			name: "unsupported compression",
			layer: func(t *testing.T, repetition int) ([]byte, string) {
				return makeLayerTar(t, []layerEntryFixture{{name: "file", mode: 0644, body: []byte{byte(repetition)}}}, false), layerMediaTar + "+zstd"
			},
			wantReason: "unsupported layer media type",
		},
		{
			name: "stream bound",
			configure: func() func() {
				old := maxLayerStreamBytes
				maxLayerStreamBytes = 512
				return func() { maxLayerStreamBytes = old }
			},
			layer: func(t *testing.T, repetition int) ([]byte, string) {
				return makeLayerTar(t, []layerEntryFixture{{name: "file", mode: 0644, body: bytes.Repeat([]byte{byte(repetition)}, 2048)}}, true), layerMediaTarGzip
			},
			wantReason: "layer stream byte bound",
		},
		{
			name: "entry bound",
			configure: func() func() {
				old := maxLayerEntries
				maxLayerEntries = 1
				return func() { maxLayerEntries = old }
			},
			layer: func(t *testing.T, repetition int) ([]byte, string) {
				return makeLayerTar(t, []layerEntryFixture{{name: "one", mode: 0644, body: []byte("1")}, {name: "two", mode: 0644, body: []byte{byte(repetition)}}}, false), layerMediaTar
			},
			wantReason: "layer entry bound",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			restore := func() {}
			if test.configure != nil {
				restore = test.configure()
			}
			defer restore()
			dir, contract := makeLayerDiagnosticEvidence(t, func(repetition int) ([]byte, string) { return test.layer(t, repetition) })
			report := VerifyDirectory(dir, contract)
			if report.Valid || len(report.LayerDiffs) != 0 || !strings.Contains(strings.Join(report.Violations, " "), test.wantReason) {
				t.Fatalf("adversarial layer did not fail closed: %#v", report)
			}
		})
	}
}

func TestLayerEntryMetadataAndDescriptorBoundaries(t *testing.T) {
	raw := makeLayerTar(t, []layerEntryFixture{
		{name: "legacy-file", legacyRegular: true, mode: 0644, body: []byte("legacy")},
		{name: "dir", typeflag: tar.TypeDir, mode: 0755},
		{name: "dir/symlink", typeflag: tar.TypeSymlink, mode: 0777, linkname: "/target"},
		{name: "dir/hardlink", typeflag: tar.TypeLink, mode: 0644, linkname: "dir/file"},
		{name: "char", typeflag: tar.TypeChar, mode: 0600},
		{name: "block", typeflag: tar.TypeBlock, mode: 0600},
		{name: "fifo", typeflag: tar.TypeFifo, mode: 0600},
		{name: "dir/file", mode: 0644, uname: "root", gname: "root", pax: map[string]string{"comment": "bounded", "SCHILY.xattr.user.test": "value"}, body: []byte("body")},
	}, false)
	entries, err := readLayerTar(bytes.NewReader(raw))
	if err != nil || len(entries) != 8 || entries["legacy-file"].ContentSHA256 == "" || entries["dir/symlink"].LinkTarget != "/target" || entries["dir/file"].MetadataSHA256 == "" {
		t.Fatalf("supported layer metadata rejected: %v %#v", err, entries)
	}
	withoutXattr := makeLayerTar(t, []layerEntryFixture{{name: "dir/file", mode: 0644, uname: "root", gname: "root", pax: map[string]string{"comment": "bounded"}, body: []byte("body")}}, false)
	entriesWithoutXattr, err := readLayerTar(bytes.NewReader(withoutXattr))
	if err != nil || entries["dir/file"].MetadataSHA256 == entriesWithoutXattr["dir/file"].MetadataSHA256 {
		t.Fatalf("PAX extended attribute not represented in metadata digest: %v", err)
	}

	unsupported := makeLayerTar(t, []layerEntryFixture{{name: "unknown", typeflag: tar.TypeCont, mode: 0600}}, false)
	if _, err = readLayerTar(bytes.NewReader(unsupported)); err == nil || !strings.Contains(err.Error(), "unsupported layer entry type") {
		t.Fatalf("unsupported tar type accepted: %v", err)
	}

	oldPath, oldLink, oldMetadata := maxLayerPathBytes, maxLayerLinkBytes, maxLayerMetadataBytes
	defer func() { maxLayerPathBytes, maxLayerLinkBytes, maxLayerMetadataBytes = oldPath, oldLink, oldMetadata }()
	maxLayerPathBytes = 3
	if _, err = readLayerTar(bytes.NewReader(makeLayerTar(t, []layerEntryFixture{{name: "long-path", mode: 0600}}, false))); err == nil || !strings.Contains(err.Error(), "layer path bound") {
		t.Fatalf("oversized path accepted: %v", err)
	}
	maxLayerPathBytes = oldPath
	maxLayerLinkBytes = 3
	if _, err = readLayerTar(bytes.NewReader(makeLayerTar(t, []layerEntryFixture{{name: "link", typeflag: tar.TypeSymlink, mode: 0777, linkname: "long-target"}}, false))); err == nil || !strings.Contains(err.Error(), "layer link target bound") {
		t.Fatalf("oversized link target accepted: %v", err)
	}
	maxLayerLinkBytes = oldLink
	maxLayerMetadataBytes = 3
	if _, err = readLayerTar(bytes.NewReader(makeLayerTar(t, []layerEntryFixture{{name: "file", mode: 0600, uname: "long-owner"}}, false))); err == nil || !strings.Contains(err.Error(), "layer metadata bound") {
		t.Fatalf("oversized metadata accepted: %v", err)
	}

	validDescriptor := Descriptor{Digest: "sha256:" + strings.Repeat("a", 64), Size: json.Number("1")}
	if digest, size, err := layerDescriptorIdentity(validDescriptor); err != nil || digest != strings.Repeat("a", 64) || size != 1 {
		t.Fatalf("valid layer descriptor rejected: %q %d %v", digest, size, err)
	}
	for _, descriptor := range []Descriptor{
		{Digest: "md5:" + strings.Repeat("a", 32), Size: json.Number("1")},
		{Digest: "sha256:bad", Size: json.Number("1")},
		{Digest: "sha256:" + strings.Repeat("a", 64), Size: json.Number("-1")},
	} {
		if _, _, err := layerDescriptorIdentity(descriptor); err == nil {
			t.Fatalf("unsafe layer descriptor accepted: %#v", descriptor)
		}
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

type layerEntryFixture struct {
	name          string
	typeflag      byte
	legacyRegular bool
	mode          int64
	uid           int
	gid           int
	linkname      string
	uname         string
	gname         string
	pax           map[string]string
	body          []byte
}

func makeLayerTar(t *testing.T, entries []layerEntryFixture, compressed bool) []byte {
	t.Helper()
	var raw bytes.Buffer
	writer := tar.NewWriter(&raw)
	for _, entry := range entries {
		typeflag := entry.typeflag
		if typeflag == 0 && !entry.legacyRegular {
			typeflag = tar.TypeReg
		}
		header := &tar.Header{
			Name: entry.name, Mode: entry.mode, Uid: entry.uid, Gid: entry.gid,
			Typeflag: typeflag, Linkname: entry.linkname, Uname: entry.uname,
			Gname: entry.gname, PAXRecords: entry.pax,
			ModTime: time.Unix(0, 0),
		}
		if typeflag == tar.TypeReg || typeflag == 0 {
			header.Size = int64(len(entry.body))
		}
		if err := writer.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if len(entry.body) > 0 {
			if _, err := writer.Write(entry.body); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if !compressed {
		return append([]byte(nil), raw.Bytes()...)
	}
	var encoded bytes.Buffer
	gzipWriter := gzip.NewWriter(&encoded)
	gzipWriter.Header.ModTime = time.Unix(0, 0)
	if _, err := gzipWriter.Write(raw.Bytes()); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	return encoded.Bytes()
}

func makeLayerDiagnosticEvidence(t *testing.T, layerFor func(int) ([]byte, string)) (string, string) {
	t.Helper()
	dir, contract, bundle := makeEvidence(t, "linux/amd64")
	for buildIndex := range bundle.Targets[0].Builds {
		build := &bundle.Targets[0].Builds[buildIndex]
		layer, mediaType := layerFor(buildIndex + 1)
		archive := filepath.Join(dir, build.File)
		writeOCIArchiveWithLayer(t, archive, "amd64", layer, mediaType)
		build.SHA256 = mustHash(t, archive)
	}
	writeJSON(t, filepath.Join(dir, "oci-evidence.json"), bundle)
	return dir, contract
}

func writeOCIArchiveWithLayer(t *testing.T, filePath, arch string, layer []byte, mediaType string) {
	t.Helper()
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
			"mediaType": mediaType, "digest": "sha256:" + layerHash, "size": len(layer),
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
	handle, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	writer := tar.NewWriter(handle)
	for _, entry := range entries {
		header := &tar.Header{Name: entry.name, Mode: 0600, Size: int64(len(entry.data)), ModTime: time.Unix(0, 0)}
		if err = writer.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if _, err = writer.Write(entry.data); err != nil {
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
