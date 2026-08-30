package candidateevidence

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	testCommit = "0123456789abcdef0123456789abcdef01234567"
	testTag    = "v0.5.0"
)

var testContract = filepath.Join("..", "configs", "release", "candidate-evidence.json")

func TestValidEvidence(t *testing.T) {
	dir, _ := makeEvidence(t)
	report := VerifyDirectory(dir, testContract)
	if !report.Valid {
		t.Fatalf("valid fixture rejected: %v", report.Violations)
	}
	if report.Tag != testTag || report.Commit != testCommit || report.BinaryTargets != 2 || report.OCITargets != 4 {
		t.Fatalf("report identity is incomplete: %#v", report)
	}
}

func TestCommittedFixtures(t *testing.T) {
	fixtures := filepath.Join("..", "testdata", "candidateevidence")
	if report := VerifyDirectory(filepath.Join(fixtures, "valid"), testContract); !report.Valid {
		t.Fatalf("committed valid fixture rejected: %v", report.Violations)
	}
	if report := VerifyDirectory(filepath.Join(fixtures, "invalid-claims"), testContract); report.Valid {
		t.Fatal("committed true-claims fixture accepted")
	}
}

func TestTagGrammar(t *testing.T) {
	for _, tag := range []string{"0.5.0", "v0.5", "v0.5.0.1", "v01.0.0", "v0.5.0-rc1", "V0.5.0", "", "latest"} {
		t.Run("reject "+tag, func(t *testing.T) {
			dir, m := makeEvidence(t)
			m.Source.Tag = tag
			writeJSON(t, filepath.Join(dir, "candidate-evidence.json"), m)
			if VerifyDirectory(dir, testContract).Valid {
				t.Fatal("malformed tag accepted")
			}
		})
	}
}

func TestAdversarialEvidenceRejected(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*testing.T, string, *Manifest)
	}{
		{"real tag claim", func(_ *testing.T, _ string, m *Manifest) { m.Claims.RealTagCreated = true }},
		{"ref pushed claim", func(_ *testing.T, _ string, m *Manifest) { m.Claims.RefPushed = true }},
		{"signed claim", func(_ *testing.T, _ string, m *Manifest) { m.Claims.Signed = true }},
		{"published claim", func(_ *testing.T, _ string, m *Manifest) { m.Claims.Published = true }},
		{"deployed claim", func(_ *testing.T, _ string, m *Manifest) { m.Claims.Deployed = true }},
		{"production claim", func(_ *testing.T, _ string, m *Manifest) { m.Claims.Production = true }},
		{"schema drift", func(_ *testing.T, _ string, m *Manifest) { m.Schema = "truerepublic.release-candidate-evidence/v2" }},
		{"contract digest drift", func(_ *testing.T, _ string, m *Manifest) { m.ContractSHA256 = strings.Repeat("0", 64) }},
		{"uppercase commit", func(_ *testing.T, _ string, m *Manifest) { m.Source.Commit = strings.ToUpper(testCommit) }},
		{"short commit", func(_ *testing.T, _ string, m *Manifest) { m.Source.Commit = testCommit[:39] }},
		{"artifact digest drift", func(_ *testing.T, _ string, m *Manifest) { m.BinaryTargets[0].ArtifactSHA256 = strings.Repeat("0", 64) }},
		{"missing binary target", func(_ *testing.T, _ string, m *Manifest) { m.BinaryTargets = m.BinaryTargets[:1] }},
		{"extra binary target", func(_ *testing.T, _ string, m *Manifest) {
			m.BinaryTargets = append(m.BinaryTargets, m.BinaryTargets[1])
		}},
		{"duplicate binary target", func(_ *testing.T, _ string, m *Manifest) { m.BinaryTargets[1] = m.BinaryTargets[0] }},
		{"swapped binary targets", func(_ *testing.T, _ string, m *Manifest) {
			m.BinaryTargets[0], m.BinaryTargets[1] = m.BinaryTargets[1], m.BinaryTargets[0]
		}},
		{"missing OCI platform", func(_ *testing.T, _ string, m *Manifest) { m.OCIPlatforms = m.OCIPlatforms[:1] }},
		{"duplicate OCI platform", func(_ *testing.T, _ string, m *Manifest) { m.OCIPlatforms[1] = m.OCIPlatforms[0] }},
		{"swapped OCI platforms", func(_ *testing.T, _ string, m *Manifest) {
			m.OCIPlatforms[0], m.OCIPlatforms[1] = m.OCIPlatforms[1], m.OCIPlatforms[0]
		}},
		{"missing OCI target", func(_ *testing.T, _ string, m *Manifest) { m.OCIPlatforms[0].Targets = m.OCIPlatforms[0].Targets[:1] }},
		{"swapped OCI targets", func(_ *testing.T, _ string, m *Manifest) {
			m.OCIPlatforms[0].Targets[0], m.OCIPlatforms[0].Targets[1] = m.OCIPlatforms[0].Targets[1], m.OCIPlatforms[0].Targets[0]
		}},
		{"OCI identity drift", func(_ *testing.T, _ string, m *Manifest) {
			m.OCIPlatforms[0].Targets[0].IndexSHA256 = strings.Repeat("0", 64)
		}},
		{"malformed metadata digest", func(_ *testing.T, _ string, m *Manifest) { m.BinaryTargets[0].Metadata.SHA256 = "bad" }},
		{"shared file binding", func(_ *testing.T, _ string, m *Manifest) {
			m.BinaryTargets[1].Metadata = m.BinaryTargets[0].Metadata
		}},
		{"metadata path traversal", func(_ *testing.T, _ string, m *Manifest) { m.BinaryTargets[0].Metadata.File = "../escape.json" }},
		{"metadata backslash path", func(_ *testing.T, _ string, m *Manifest) { m.BinaryTargets[0].Metadata.File = `..\escape.json` }},
		{"metadata absolute path", func(_ *testing.T, _ string, m *Manifest) { m.BinaryTargets[0].Metadata.File = "/tmp/escape.json" }},
		{"OCI report path traversal", func(_ *testing.T, _ string, m *Manifest) { m.OCIPlatforms[0].Report.File = "../escape.json" }},

		{"tag commit mismatch", func(t *testing.T, dir string, m *Manifest) {
			mutateMetadata(t, dir, m, 0, func(v map[string]any) { v["source_ref"] = strings.Repeat("f", 40) })
		}},
		{"metadata target mismatch", func(t *testing.T, dir string, m *Manifest) {
			mutateMetadata(t, dir, m, 0, func(v map[string]any) { v["target"] = "linux-arm64" })
		}},
		{"metadata digest drift", func(t *testing.T, dir string, m *Manifest) {
			mutateMetadata(t, dir, m, 0, func(v map[string]any) { v["go_version"] = "1.99.0" })
			m.BinaryTargets[0].Metadata.SHA256 = fileHash(t, dir, m.BinaryTargets[0].Metadata.File)
		}},
		{"metadata stale digest", func(t *testing.T, dir string, m *Manifest) {
			mutateMetadata(t, dir, m, 0, func(v map[string]any) { v["go_version"] = "1.99.0" })
		}},
		{"metadata contract drift", func(t *testing.T, dir string, m *Manifest) {
			mutateMetadata(t, dir, m, 0, func(v map[string]any) { v["contract_sha256"] = strings.Repeat("0", 64) })
			m.BinaryTargets[0].Metadata.SHA256 = fileHash(t, dir, m.BinaryTargets[0].Metadata.File)
		}},
		{"checksum drift", func(t *testing.T, dir string, m *Manifest) {
			writeFile(t, filepath.Join(dir, m.BinaryTargets[0].Checksums.File), []byte(strings.Repeat("0", 64)+"  "+m.BinaryTargets[0].Artifact+"\n"))
		}},
		{"checksum extra line", func(t *testing.T, dir string, m *Manifest) {
			path := filepath.Join(dir, m.BinaryTargets[0].Checksums.File)
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			writeFile(t, path, append(raw, raw...))
		}},
		{"OCI bundle source mismatch", func(t *testing.T, dir string, m *Manifest) {
			mutateOCIEvidence(t, dir, m, 0, func(v map[string]any) { v["source_ref"] = strings.Repeat("f", 40) })
			m.OCIPlatforms[0].Evidence.SHA256 = fileHash(t, dir, m.OCIPlatforms[0].Evidence.File)
		}},
		{"OCI bundle claims true", func(t *testing.T, dir string, m *Manifest) {
			mutateOCIEvidence(t, dir, m, 0, func(v map[string]any) {
				v["claims"] = map[string]any{"signed": true, "published": false, "production": false}
			})
			m.OCIPlatforms[0].Evidence.SHA256 = fileHash(t, dir, m.OCIPlatforms[0].Evidence.File)
		}},
		{"OCI bundle contract drift", func(t *testing.T, dir string, m *Manifest) {
			mutateOCIEvidence(t, dir, m, 0, func(v map[string]any) { v["contract_sha256"] = strings.Repeat("0", 64) })
			m.OCIPlatforms[0].Evidence.SHA256 = fileHash(t, dir, m.OCIPlatforms[0].Evidence.File)
		}},
		{"OCI bundle stale digest", func(t *testing.T, dir string, m *Manifest) {
			mutateOCIEvidence(t, dir, m, 0, func(v map[string]any) { v["platform"] = "linux/s390x" })
		}},
		{"OCI report invalid", func(t *testing.T, dir string, m *Manifest) {
			mutateOCIReport(t, dir, m, 0, func(v map[string]any) { v["valid"] = false })
			m.OCIPlatforms[0].Report.SHA256 = fileHash(t, dir, m.OCIPlatforms[0].Report.File)
		}},
		{"OCI report layer diff", func(t *testing.T, dir string, m *Manifest) {
			mutateOCIReport(t, dir, m, 0, func(v map[string]any) { v["layer_diffs"] = []any{map[string]any{"target": "daemon-linux-amd64"}} })
			m.OCIPlatforms[0].Report.SHA256 = fileHash(t, dir, m.OCIPlatforms[0].Report.File)
		}},
		{"OCI report platform mismatch", func(t *testing.T, dir string, m *Manifest) {
			mutateOCIReport(t, dir, m, 0, func(v map[string]any) { v["platform"] = "linux/arm64" })
			m.OCIPlatforms[0].Report.SHA256 = fileHash(t, dir, m.OCIPlatforms[0].Report.File)
		}},
		{"OCI report stale digest", func(t *testing.T, dir string, m *Manifest) {
			mutateOCIReport(t, dir, m, 0, func(v map[string]any) { v["valid"] = false })
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir, m := makeEvidence(t)
			test.mutate(t, dir, &m)
			writeJSON(t, filepath.Join(dir, "candidate-evidence.json"), m)
			if report := VerifyDirectory(dir, testContract); report.Valid {
				t.Fatal("adversarial evidence accepted")
			}
		})
	}
}

func TestRawAdversarialManifest(t *testing.T) {
	t.Run("duplicate key", func(t *testing.T) {
		dir, _ := makeEvidence(t)
		raw := mustRead(t, filepath.Join(dir, "candidate-evidence.json"))
		mutated := bytes.Replace(raw, []byte(`"schema":`), []byte(`"schema":"`+Schema+`","schema":`), 1)
		writeFile(t, filepath.Join(dir, "candidate-evidence.json"), mutated)
		if VerifyDirectory(dir, testContract).Valid {
			t.Fatal("duplicate manifest key accepted")
		}
	})
	t.Run("unknown key", func(t *testing.T) {
		dir, _ := makeEvidence(t)
		raw := mustRead(t, filepath.Join(dir, "candidate-evidence.json"))
		mutated := bytes.Replace(raw, []byte(`"schema":`), []byte(`"unknown":true,"schema":`), 1)
		writeFile(t, filepath.Join(dir, "candidate-evidence.json"), mutated)
		if VerifyDirectory(dir, testContract).Valid {
			t.Fatal("unknown manifest key accepted")
		}
	})
	t.Run("missing claim", func(t *testing.T) {
		dir, _ := makeEvidence(t)
		var value map[string]any
		if err := json.Unmarshal(mustRead(t, filepath.Join(dir, "candidate-evidence.json")), &value); err != nil {
			t.Fatal(err)
		}
		delete(value["claims"].(map[string]any), "production")
		writeJSON(t, filepath.Join(dir, "candidate-evidence.json"), value)
		if VerifyDirectory(dir, testContract).Valid {
			t.Fatal("incomplete claims accepted")
		}
	})
	t.Run("symlink escape", func(t *testing.T) {
		dir, m := makeEvidence(t)
		outside := filepath.Join(t.TempDir(), "metadata.json")
		writeFile(t, outside, mustRead(t, filepath.Join(dir, m.BinaryTargets[0].Metadata.File)))
		if err := os.Remove(filepath.Join(dir, m.BinaryTargets[0].Metadata.File)); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, filepath.Join(dir, m.BinaryTargets[0].Metadata.File)); err != nil {
			t.Fatal(err)
		}
		if VerifyDirectory(dir, testContract).Valid {
			t.Fatal("symlink member accepted")
		}
	})
	t.Run("symlinked evidence directory", func(t *testing.T) {
		dir, _ := makeEvidence(t)
		link := filepath.Join(t.TempDir(), "evidence")
		if err := os.Symlink(dir, link); err != nil {
			t.Fatal(err)
		}
		if VerifyDirectory(link, testContract).Valid {
			t.Fatal("symlinked evidence directory accepted")
		}
	})
	t.Run("undeclared evidence member", func(t *testing.T) {
		dir, _ := makeEvidence(t)
		writeFile(t, filepath.Join(dir, "daemon-linux-amd64.oci.tar"), []byte("payload"))
		if VerifyDirectory(dir, testContract).Valid {
			t.Fatal("undeclared evidence member accepted")
		}
	})
	t.Run("missing evidence directory", func(t *testing.T) {
		if VerifyDirectory(filepath.Join(t.TempDir(), "missing"), testContract).Valid {
			t.Fatal("missing evidence directory accepted")
		}
	})
	t.Run("missing manifest", func(t *testing.T) {
		if VerifyDirectory(t.TempDir(), testContract).Valid {
			t.Fatal("missing manifest accepted")
		}
	})
}

func TestContractDriftRejected(t *testing.T) {
	dir, _ := makeEvidence(t)
	for _, test := range []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"tag grammar", func(v map[string]any) { v["source"].(map[string]any)["tag_pattern"] = "^.*$" }},
		{"commit grammar", func(v map[string]any) { v["source"].(map[string]any)["commit_pattern"] = "^.*$" }},
		{"schema", func(v map[string]any) { v["schema"] = "truerepublic.release-candidate/v2" }},
		{"binary target", func(v map[string]any) { v["binary_targets"].([]any)[0].(map[string]any)["id"] = "darwin-amd64" }},
		{"OCI target", func(v map[string]any) { v["oci_targets"].([]any)[0].(map[string]any)["platform"] = "linux/s390x" }},
		{"repetitions", func(v map[string]any) { v["oci_repetitions"] = json.Number("3") }},
	} {
		t.Run(test.name, func(t *testing.T) {
			var value map[string]any
			if err := json.Unmarshal(mustRead(t, testContract), &value); err != nil {
				t.Fatal(err)
			}
			test.mutate(value)
			mutated := filepath.Join(t.TempDir(), "contract.json")
			writeJSON(t, mutated, value)
			if VerifyDirectory(dir, mutated).Valid {
				t.Fatal("drifted contract accepted")
			}
		})
	}
}

func TestDeterministicReport(t *testing.T) {
	dir, m := makeEvidence(t)
	first := mustJSON(t, VerifyDirectory(dir, testContract))
	second := mustJSON(t, VerifyDirectory(dir, testContract))
	if !bytes.Equal(first, second) {
		t.Fatal("repeated verification reports differ")
	}
	m.Claims.Production = true
	writeJSON(t, filepath.Join(dir, "candidate-evidence.json"), m)
	invalidFirst := mustJSON(t, VerifyDirectory(dir, testContract))
	invalidSecond := mustJSON(t, VerifyDirectory(dir, testContract))
	if !bytes.Equal(invalidFirst, invalidSecond) {
		t.Fatal("repeated violation reports differ")
	}
}

func TestRunCLI(t *testing.T) {
	dir, m := makeEvidence(t)
	var stdout, stderr bytes.Buffer
	if code := Run(nil, &stdout, &stderr); code != 2 {
		t.Fatal("missing subcommand accepted")
	}
	if code := Run([]string{"verify", "--unknown", "x"}, &stdout, &stderr); code != 2 {
		t.Fatal("unknown flag accepted")
	}
	if code := Run([]string{"verify", "--evidence", dir, "--contract", testContract, "--output", "yaml"}, &stdout, &stderr); code != 2 {
		t.Fatal("unknown output format accepted")
	}
	stdout.Reset()
	if code := Run([]string{"verify", "--evidence", dir, "--contract", testContract, "--output", "json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("valid evidence rejected: %s %s", stdout.String(), stderr.String())
	}
	var report Report
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &report); err != nil || !report.Valid || report.BinaryTargets != 2 || report.OCITargets != 4 {
		t.Fatalf("JSON report is invalid: %s", stdout.String())
	}
	m.Claims.Production = true
	writeJSON(t, filepath.Join(dir, "candidate-evidence.json"), m)
	stdout.Reset()
	if code := Run([]string{"verify", "--evidence", dir, "--contract", testContract, "--output", "json"}, &stdout, &stderr); code != 1 {
		t.Fatal("invalid evidence accepted")
	}
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
	if parseBytes([]byte(`{"real_tag_created":false,"ref_pushed":false,"signed":false,"published":false,"deployed":false}`), &claims) == nil {
		t.Fatal("incomplete candidate claims accepted")
	}
}

func makeEvidence(t *testing.T) (string, Manifest) {
	t.Helper()
	dir := t.TempDir()
	var contract Contract
	if err := parseFile(testContract, &contract); err != nil {
		t.Fatal(err)
	}
	contractHash, err := hashRegularFile(testContract, MaxJSONBytes)
	if err != nil {
		t.Fatal(err)
	}
	m := Manifest{
		Schema:         Schema,
		ContractSHA256: contractHash,
		Source:         SourceIdentity{Tag: testTag, Commit: testCommit},
		Claims:         Claims{present: true},
	}
	for _, target := range expectedBinaryTargets() {
		digest := sum([]byte("synthetic-" + target.ID))
		writeFile(t, filepath.Join(dir, "checksums-"+target.ID+".sha256"), []byte(digest+"  "+target.Artifact+"\n"))
		metadata := daemonMetadata{
			Schema:          BuildEvidenceSchema,
			ContractSchema:  BuildContractSchema,
			ContractSHA256:  contract.BinaryContractSHA256,
			SourceRef:       testCommit,
			Target:          target.ID,
			CIRunner:        target.CIRunner,
			RunnerArch:      target.RunnerArch,
			Artifact:        target.Artifact,
			SHA256:          digest,
			Reproducible:    []string{digest, digest},
			GoVersion:       "1.26.6",
			CGOEnabled:      "1",
			SourceDateEpoch: json.Number("1767225600"),
		}
		metadata.BuildFlags.Trimpath = true
		metadata.BuildFlags.Mod = "readonly"
		metadata.BuildFlags.LinkerBuildID = "none"
		metadata.BuildFlags.VersionVariable = "main.version"
		metadataFile := "build-metadata-" + target.ID + ".json"
		writeJSON(t, filepath.Join(dir, metadataFile), metadata)
		m.BinaryTargets = append(m.BinaryTargets, BinaryTargetBinding{
			ID:             target.ID,
			Artifact:       target.Artifact,
			ArtifactSHA256: digest,
			Metadata:       FileDigest{File: metadataFile, SHA256: fileHash(t, dir, metadataFile)},
			Checksums:      FileDigest{File: "checksums-" + target.ID + ".sha256", SHA256: fileHash(t, dir, "checksums-"+target.ID+".sha256")},
		})
	}
	for index, platform := range expectedOCIPlatforms() {
		suffix := []string{"amd64", "arm64"}[index]
		bundle := ociBundle{
			Schema:         OCIEvidenceSchema,
			SourceRef:      testCommit,
			ContractSHA256: contract.OCIContractSHA256,
			Platform:       platform,
			Claims:         ociClaims{present: true},
		}
		digestReport := ociDigestReport{
			Schema:     OCIReportSchema,
			Valid:      true,
			Platform:   platform,
			Targets:    2,
			Violations: []string{},
		}
		binding := OCIPlatformBinding{Platform: platform}
		for _, target := range ociTargetsForPlatform(expectedOCITargets(), platform) {
			builds := []ociArchiveBinding{}
			for repetition := 1; repetition <= 2; repetition++ {
				builds = append(builds, ociArchiveBinding{
					File:   target.ID + "-" + string(rune('0'+repetition)) + ".oci.tar",
					SHA256: sum([]byte("archive-" + target.ID + "-" + string(rune('0'+repetition)))),
				})
			}
			bundle.Targets = append(bundle.Targets, ociTargetArchive{ID: target.ID, Builds: builds})
			identity := OCIImageIdentity{
				ID:             target.ID,
				IndexSHA256:    sum([]byte("index-" + target.ID)),
				ManifestSHA256: sum([]byte("manifest-" + target.ID)),
				ConfigSHA256:   sum([]byte("config-" + target.ID)),
				LayerSHA256:    []string{sum([]byte("layer-" + target.ID + "-0"))},
			}
			binding.Targets = append(binding.Targets, identity)
			digestReport.Images = append(digestReport.Images, ociImageReport{
				ID:       target.ID,
				Index:    identity.IndexSHA256,
				Manifest: identity.ManifestSHA256,
				Config:   identity.ConfigSHA256,
				Layers:   identity.LayerSHA256,
			})
		}
		evidenceFile := "oci-evidence-" + suffix + ".json"
		reportFile := "oci-report-" + suffix + ".json"
		writeJSON(t, filepath.Join(dir, evidenceFile), bundle)
		writeJSON(t, filepath.Join(dir, reportFile), digestReport)
		binding.Evidence = FileDigest{File: evidenceFile, SHA256: fileHash(t, dir, evidenceFile)}
		binding.Report = FileDigest{File: reportFile, SHA256: fileHash(t, dir, reportFile)}
		m.OCIPlatforms = append(m.OCIPlatforms, binding)
	}
	writeJSON(t, filepath.Join(dir, "candidate-evidence.json"), m)
	return dir, m
}

func mutateMetadata(t *testing.T, dir string, m *Manifest, target int, change func(map[string]any)) {
	t.Helper()
	mutateJSON(t, filepath.Join(dir, m.BinaryTargets[target].Metadata.File), change)
}

func mutateOCIEvidence(t *testing.T, dir string, m *Manifest, platform int, change func(map[string]any)) {
	t.Helper()
	mutateJSON(t, filepath.Join(dir, m.OCIPlatforms[platform].Evidence.File), change)
}

func mutateOCIReport(t *testing.T, dir string, m *Manifest, platform int, change func(map[string]any)) {
	t.Helper()
	mutateJSON(t, filepath.Join(dir, m.OCIPlatforms[platform].Report.File), change)
}

func mutateJSON(t *testing.T, path string, change func(map[string]any)) {
	t.Helper()
	var value map[string]any
	if err := json.Unmarshal(mustRead(t, path), &value); err != nil {
		t.Fatal(err)
	}
	change(value)
	writeJSON(t, path, value)
}

func sum(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

func fileHash(t *testing.T, dir, name string) string {
	t.Helper()
	h, err := hashRegularFile(filepath.Join(dir, name), maxMemberBytes)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	writeFile(t, path, mustJSON(t, v))
}
