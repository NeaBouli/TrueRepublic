package crossrunevidence

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	candidate "truerepublic/candidateevidence"
)

const (
	testCommit         = "0123456789abcdef0123456789abcdef01234567"
	testTag            = "v0.5.0"
	testBaselineRunID  = "33400000001"
	testCurrentRunID   = "33400000002"
	testBaselineCreate = "2026-08-30T10:00:00Z"
	testCurrentCreate  = "2026-09-01T10:00:00Z"
)

var (
	testContract          = filepath.Join("..", "configs", "release", "cross-run-rebuild.json")
	testCandidateContract = filepath.Join("..", "configs", "release", "candidate-evidence.json")
)

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("write rejected") }

func testExpected() Expected {
	return Expected{
		Repository:    "NeaBouli/TrueRepublic",
		WorkflowPath:  ".github/workflows/reproducible-daemon.yml",
		Branch:        "main",
		Commit:        testCommit,
		BaselineRunID: testBaselineRunID,
		CurrentRunID:  testCurrentRunID,
	}
}

func TestValidEvidence(t *testing.T) {
	dir, _ := makeEvidence(t)
	report := Compare(dir, testContract, testCandidateContract, testExpected())
	if !report.Valid {
		t.Fatalf("valid evidence rejected: %v", report.Violations)
	}
	if report.Repository != "NeaBouli/TrueRepublic" || report.WorkflowPath != ".github/workflows/reproducible-daemon.yml" ||
		report.Branch != "main" || report.Commit != testCommit || report.Tag != testTag ||
		report.BaselineRunID != testBaselineRunID || report.CurrentRunID != testCurrentRunID ||
		report.BinaryTargets != 2 || report.OCITargets != 4 {
		t.Fatalf("report identity is incomplete: %#v", report)
	}
}

func TestArtifactRetentionTolerance(t *testing.T) {
	t.Run("accepts narrow GitHub timestamp skew", func(t *testing.T) {
		dir, manifest := makeEvidence(t)
		mutateReceipt(t, dir, &manifest, "baseline", func(v map[string]any) {
			v["artifact"].(map[string]any)["expires_at"] = "2026-09-13T10:20:59Z"
		})
		writeJSON(t, filepath.Join(dir, "cross-run-evidence.json"), manifest)
		if report := Compare(dir, testContract, testCandidateContract, testExpected()); !report.Valid {
			t.Fatalf("narrow timestamp skew rejected: %v", report.Violations)
		}
	})

	t.Run("rejects timestamp skew outside tolerance", func(t *testing.T) {
		dir, manifest := makeEvidence(t)
		mutateReceipt(t, dir, &manifest, "baseline", func(v map[string]any) {
			v["artifact"].(map[string]any)["expires_at"] = "2026-09-13T10:21:01Z"
		})
		writeJSON(t, filepath.Join(dir, "cross-run-evidence.json"), manifest)
		report := Compare(dir, testContract, testCandidateContract, testExpected())
		if report.Valid || !strings.Contains(strings.Join(report.Violations, "\n"), "retention window mismatch") {
			t.Fatalf("timestamp skew outside tolerance was not rejected precisely: %v", report.Violations)
		}
	})
}

func TestCommittedFixtures(t *testing.T) {
	fixtures := filepath.Join("..", "testdata", "crossrunevidence")
	expected := testExpected()
	if report := Compare(filepath.Join(fixtures, "valid"), testContract, testCandidateContract, expected); !report.Valid {
		t.Fatalf("committed valid fixture rejected: %v", report.Violations)
	}
	if report := Compare(filepath.Join(fixtures, "invalid-claims"), testContract, testCandidateContract, expected); report.Valid {
		t.Fatal("committed true-claims fixture accepted")
	} else if !strings.Contains(strings.Join(report.Violations, "\n"), "cross-run status claims must remain false") {
		t.Fatalf("true-claims fixture rejected for the wrong reason: %v", report.Violations)
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
		{"attested claim", func(_ *testing.T, _ string, m *Manifest) { m.Claims.Attested = true }},
		{"published claim", func(_ *testing.T, _ string, m *Manifest) { m.Claims.Published = true }},
		{"deployed claim", func(_ *testing.T, _ string, m *Manifest) { m.Claims.Deployed = true }},
		{"production claim", func(_ *testing.T, _ string, m *Manifest) { m.Claims.Production = true }},
		{"long-term-hermetic claim", func(_ *testing.T, _ string, m *Manifest) { m.Claims.LongTermHermetic = true }},
		{"schema drift", func(_ *testing.T, _ string, m *Manifest) { m.Schema = "truerepublic.cross-run-evidence/v2" }},
		{"contract digest drift", func(_ *testing.T, _ string, m *Manifest) { m.ContractSHA256 = strings.Repeat("0", 64) }},
		{"repository drift", func(_ *testing.T, _ string, m *Manifest) { m.Comparison.Repository = "example/fork" }},
		{"workflow drift", func(_ *testing.T, _ string, m *Manifest) { m.Comparison.WorkflowPath = ".github/workflows/other.yml" }},
		{"branch drift", func(_ *testing.T, _ string, m *Manifest) { m.Comparison.Branch = "feature" }},
		{"commit drift", func(_ *testing.T, _ string, m *Manifest) { m.Comparison.Commit = strings.Repeat("f", 40) }},
		{"uppercase commit", func(_ *testing.T, _ string, m *Manifest) { m.Comparison.Commit = strings.ToUpper(testCommit) }},
		{"short commit", func(_ *testing.T, _ string, m *Manifest) { m.Comparison.Commit = testCommit[:39] }},
		{"same run", func(_ *testing.T, _ string, m *Manifest) { m.Comparison.CurrentRunID = m.Comparison.BaselineRunID }},
		{"zero run ID", func(_ *testing.T, _ string, m *Manifest) { m.Comparison.BaselineRunID = "0" }},
		{"leading-zero run ID", func(_ *testing.T, _ string, m *Manifest) { m.Comparison.BaselineRunID = "033400000001" }},
		{"non-decimal run ID", func(_ *testing.T, _ string, m *Manifest) { m.Comparison.BaselineRunID = "run-1" }},
		{"huge run ID", func(_ *testing.T, _ string, m *Manifest) { m.Comparison.BaselineRunID = "1234567890123456789012345" }},
		{"baseline member name drift", func(_ *testing.T, _ string, m *Manifest) { m.Baseline.Receipt.File = "receipt.json" }},
		{"member digest malformed", func(_ *testing.T, _ string, m *Manifest) { m.Baseline.Receipt.SHA256 = "bad" }},
		{"member digest drift", func(_ *testing.T, _ string, m *Manifest) { m.Baseline.Receipt.SHA256 = strings.Repeat("0", 64) }},
		{"member path traversal", func(_ *testing.T, _ string, m *Manifest) { m.Baseline.Receipt.File = "../escape.json" }},
		{"member backslash path", func(_ *testing.T, _ string, m *Manifest) { m.Baseline.Receipt.File = `..\escape.json` }},
		{"member absolute path", func(_ *testing.T, _ string, m *Manifest) { m.Baseline.Receipt.File = "/tmp/escape.json" }},
		{"shared member binding", func(_ *testing.T, _ string, m *Manifest) { m.Current.Receipt = m.Baseline.Receipt }},

		{"receipt schema drift", func(t *testing.T, dir string, m *Manifest) {
			mutateReceipt(t, dir, m, "baseline", func(v map[string]any) { v["schema"] = "truerepublic.cross-run-receipt/v2" })
		}},
		{"receipt repository drift", func(t *testing.T, dir string, m *Manifest) {
			mutateReceipt(t, dir, m, "baseline", func(v map[string]any) { v["repository"] = "example/fork" })
		}},
		{"receipt workflow drift", func(t *testing.T, dir string, m *Manifest) {
			mutateReceipt(t, dir, m, "baseline", func(v map[string]any) { v["workflow_path"] = ".github/workflows/other.yml" })
		}},
		{"receipt branch drift", func(t *testing.T, dir string, m *Manifest) {
			mutateReceipt(t, dir, m, "baseline", func(v map[string]any) { v["branch"] = "feature" })
		}},
		{"receipt event drift", func(t *testing.T, dir string, m *Manifest) {
			mutateReceipt(t, dir, m, "baseline", func(v map[string]any) { v["event"] = "push" })
		}},
		{"receipt run ID drift", func(t *testing.T, dir string, m *Manifest) {
			mutateReceipt(t, dir, m, "baseline", func(v map[string]any) { v["run_id"] = "33400000009" })
		}},
		{"receipt zero attempt", func(t *testing.T, dir string, m *Manifest) {
			mutateReceipt(t, dir, m, "baseline", func(v map[string]any) { v["run_attempt"] = json.Number("0") })
		}},
		{"receipt huge attempt", func(t *testing.T, dir string, m *Manifest) {
			mutateReceipt(t, dir, m, "baseline", func(v map[string]any) { v["run_attempt"] = json.Number("999999999") })
		}},
		{"receipt head SHA drift", func(t *testing.T, dir string, m *Manifest) {
			mutateReceipt(t, dir, m, "baseline", func(v map[string]any) { v["head_sha"] = strings.Repeat("f", 40) })
		}},
		{"receipt status drift", func(t *testing.T, dir string, m *Manifest) {
			mutateReceipt(t, dir, m, "baseline", func(v map[string]any) { v["status"] = "in_progress" })
		}},
		{"receipt conclusion drift", func(t *testing.T, dir string, m *Manifest) {
			mutateReceipt(t, dir, m, "baseline", func(v map[string]any) { v["conclusion"] = "failure" })
		}},
		{"current receipt completed", func(t *testing.T, dir string, m *Manifest) {
			mutateReceipt(t, dir, m, "current", func(v map[string]any) {
				v["status"] = "completed"
				v["conclusion"] = "success"
			})
		}},
		{"baseline artifact name drift", func(t *testing.T, dir string, m *Manifest) {
			mutateReceipt(t, dir, m, "baseline", func(v map[string]any) {
				v["artifact"].(map[string]any)["name"] = "truerepublic-candidate-other"
			})
		}},
		{"baseline artifact ID drift", func(t *testing.T, dir string, m *Manifest) {
			mutateReceipt(t, dir, m, "baseline", func(v map[string]any) {
				v["artifact"].(map[string]any)["id"] = json.Number("0")
			})
		}},
		{"baseline artifact digest malformed", func(t *testing.T, dir string, m *Manifest) {
			mutateReceipt(t, dir, m, "baseline", func(v map[string]any) {
				v["artifact"].(map[string]any)["digest"] = "md5:" + strings.Repeat("0", 32)
			})
		}},
		{"baseline artifact expired", func(t *testing.T, dir string, m *Manifest) {
			mutateReceipt(t, dir, m, "baseline", func(v map[string]any) {
				v["artifact"].(map[string]any)["expired"] = true
			})
		}},
		{"baseline artifact retention drift", func(t *testing.T, dir string, m *Manifest) {
			mutateReceipt(t, dir, m, "baseline", func(v map[string]any) {
				v["artifact"].(map[string]any)["expires_at"] = "2026-09-14T10:20:00Z"
			})
		}},
		{"stale baseline run", func(t *testing.T, dir string, m *Manifest) {
			mutateReceipt(t, dir, m, "baseline", func(v map[string]any) { v["created_at"] = "2026-08-01T10:00:00Z" })
		}},
		{"current run predates baseline", func(t *testing.T, dir string, m *Manifest) {
			mutateReceipt(t, dir, m, "current", func(v map[string]any) { v["created_at"] = "2026-08-29T10:00:00Z" })
		}},

		{"candidate schema drift", func(t *testing.T, dir string, m *Manifest) {
			mutateCandidateManifest(t, dir, m, "current", func(v map[string]any) { v["schema"] = "truerepublic.release-candidate-evidence/v2" })
		}},
		{"candidate contract drift", func(t *testing.T, dir string, m *Manifest) {
			mutateCandidateManifest(t, dir, m, "current", func(v map[string]any) { v["contract_sha256"] = strings.Repeat("0", 64) })
		}},
		{"candidate commit drift", func(t *testing.T, dir string, m *Manifest) {
			mutateCandidateManifest(t, dir, m, "current", func(v map[string]any) {
				v["source"].(map[string]any)["commit"] = strings.Repeat("f", 40)
			})
		}},
		{"candidate claims omitted", func(t *testing.T, dir string, m *Manifest) {
			mutateCandidateManifest(t, dir, m, "current", func(v map[string]any) { delete(v, "claims") })
		}},
		{"candidate tag drift", func(t *testing.T, dir string, m *Manifest) {
			mutateCandidateManifest(t, dir, m, "current", func(v map[string]any) {
				v["source"].(map[string]any)["tag"] = "v0.6.0"
			})
			mutateCandidateReport(t, dir, m, "current", func(v map[string]any) { v["tag"] = "v0.6.0" })
		}},
		{"candidate malformed tag", func(t *testing.T, dir string, m *Manifest) {
			mutateCandidateManifest(t, dir, m, "current", func(v map[string]any) {
				v["source"].(map[string]any)["tag"] = "latest"
			})
		}},
		{"binary digest drift", func(t *testing.T, dir string, m *Manifest) {
			mutateCandidateManifest(t, dir, m, "current", func(v map[string]any) {
				v["binary_targets"].([]any)[0].(map[string]any)["artifact_sha256"] = strings.Repeat("0", 64)
			})
		}},
		{"binary target order drift", func(t *testing.T, dir string, m *Manifest) {
			mutateCandidateManifest(t, dir, m, "current", func(v map[string]any) {
				targets := v["binary_targets"].([]any)
				targets[0], targets[1] = targets[1], targets[0]
			})
		}},
		{"OCI index drift", func(t *testing.T, dir string, m *Manifest) {
			mutateOCIIdentity(t, dir, m, "index_sha256", nil)
		}},
		{"OCI manifest drift", func(t *testing.T, dir string, m *Manifest) {
			mutateOCIIdentity(t, dir, m, "manifest_sha256", nil)
		}},
		{"OCI config drift", func(t *testing.T, dir string, m *Manifest) {
			mutateOCIIdentity(t, dir, m, "config_sha256", nil)
		}},
		{"OCI layer drift", func(t *testing.T, dir string, m *Manifest) {
			mutateOCIIdentity(t, dir, m, "", func(layers []any) []any {
				layers[0] = strings.Repeat("0", 64)
				return layers
			})
		}},
		{"OCI layer append", func(t *testing.T, dir string, m *Manifest) {
			mutateOCIIdentity(t, dir, m, "", func(layers []any) []any {
				return append(layers, sum([]byte("extra-layer")))
			})
		}},
		{"OCI target order drift", func(t *testing.T, dir string, m *Manifest) {
			mutateCandidateManifest(t, dir, m, "current", func(v map[string]any) {
				targets := v["oci_platforms"].([]any)[0].(map[string]any)["targets"].([]any)
				targets[0], targets[1] = targets[1], targets[0]
			})
		}},
		{"candidate report invalid", func(t *testing.T, dir string, m *Manifest) {
			mutateCandidateReport(t, dir, m, "current", func(v map[string]any) { v["valid"] = false })
		}},
		{"candidate report violations", func(t *testing.T, dir string, m *Manifest) {
			mutateCandidateReport(t, dir, m, "current", func(v map[string]any) { v["violations"] = []any{"drift"} })
		}},
		{"candidate report commit drift", func(t *testing.T, dir string, m *Manifest) {
			mutateCandidateReport(t, dir, m, "current", func(v map[string]any) { v["commit"] = strings.Repeat("f", 40) })
		}},
		{"candidate report count drift", func(t *testing.T, dir string, m *Manifest) {
			mutateCandidateReport(t, dir, m, "current", func(v map[string]any) { v["oci_targets"] = json.Number("3") })
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir, m := makeEvidence(t)
			test.mutate(t, dir, &m)
			writeJSON(t, filepath.Join(dir, "cross-run-evidence.json"), m)
			if report := Compare(dir, testContract, testCandidateContract, testExpected()); report.Valid {
				t.Fatal("adversarial evidence accepted")
			}
		})
	}
}

func TestExpectedIdentityMismatch(t *testing.T) {
	dir, _ := makeEvidence(t)
	for _, test := range []struct {
		name   string
		mutate func(*Expected)
	}{
		{"repository", func(e *Expected) { e.Repository = "example/fork" }},
		{"workflow", func(e *Expected) { e.WorkflowPath = ".github/workflows/other.yml" }},
		{"branch", func(e *Expected) { e.Branch = "develop" }},
		{"commit", func(e *Expected) { e.Commit = strings.Repeat("f", 40) }},
		{"baseline run ID", func(e *Expected) { e.BaselineRunID = "33400000009" }},
		{"current run ID", func(e *Expected) { e.CurrentRunID = "33400000009" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			expected := testExpected()
			test.mutate(&expected)
			if report := Compare(dir, testContract, testCandidateContract, expected); report.Valid {
				t.Fatal("evidence accepted with a mismatched expected identity")
			}
		})
	}
}

func TestRawAdversarialManifest(t *testing.T) {
	t.Run("duplicate key", func(t *testing.T) {
		dir, _ := makeEvidence(t)
		raw := mustRead(t, filepath.Join(dir, "cross-run-evidence.json"))
		mutated := bytes.Replace(raw, []byte(`"schema":`), []byte(`"schema":"`+Schema+`","schema":`), 1)
		writeFile(t, filepath.Join(dir, "cross-run-evidence.json"), mutated)
		if Compare(dir, testContract, testCandidateContract, testExpected()).Valid {
			t.Fatal("duplicate manifest key accepted")
		}
	})
	t.Run("unknown key", func(t *testing.T) {
		dir, _ := makeEvidence(t)
		raw := mustRead(t, filepath.Join(dir, "cross-run-evidence.json"))
		mutated := bytes.Replace(raw, []byte(`"schema":`), []byte(`"unknown":true,"schema":`), 1)
		writeFile(t, filepath.Join(dir, "cross-run-evidence.json"), mutated)
		if Compare(dir, testContract, testCandidateContract, testExpected()).Valid {
			t.Fatal("unknown manifest key accepted")
		}
	})
	t.Run("trailing data", func(t *testing.T) {
		dir, _ := makeEvidence(t)
		raw := append(mustRead(t, filepath.Join(dir, "cross-run-evidence.json")), ' ', '{', '}')
		writeFile(t, filepath.Join(dir, "cross-run-evidence.json"), raw)
		if Compare(dir, testContract, testCandidateContract, testExpected()).Valid {
			t.Fatal("trailing manifest data accepted")
		}
	})
	t.Run("missing claim", func(t *testing.T) {
		dir, _ := makeEvidence(t)
		var value map[string]any
		if err := json.Unmarshal(mustRead(t, filepath.Join(dir, "cross-run-evidence.json")), &value); err != nil {
			t.Fatal(err)
		}
		delete(value["claims"].(map[string]any), "long_term_hermetic")
		writeJSON(t, filepath.Join(dir, "cross-run-evidence.json"), value)
		if Compare(dir, testContract, testCandidateContract, testExpected()).Valid {
			t.Fatal("incomplete claims accepted")
		}
	})
	t.Run("symlink escape", func(t *testing.T) {
		dir, m := makeEvidence(t)
		outside := filepath.Join(t.TempDir(), "receipt.json")
		writeFile(t, outside, mustRead(t, filepath.Join(dir, m.Baseline.Receipt.File)))
		if err := os.Remove(filepath.Join(dir, m.Baseline.Receipt.File)); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, filepath.Join(dir, m.Baseline.Receipt.File)); err != nil {
			t.Fatal(err)
		}
		if Compare(dir, testContract, testCandidateContract, testExpected()).Valid {
			t.Fatal("symlink member accepted")
		}
	})
	t.Run("symlinked evidence directory", func(t *testing.T) {
		dir, _ := makeEvidence(t)
		link := filepath.Join(t.TempDir(), "evidence")
		if err := os.Symlink(dir, link); err != nil {
			t.Fatal(err)
		}
		if Compare(link, testContract, testCandidateContract, testExpected()).Valid {
			t.Fatal("symlinked evidence directory accepted")
		}
	})
	t.Run("undeclared evidence member", func(t *testing.T) {
		dir, _ := makeEvidence(t)
		writeFile(t, filepath.Join(dir, "notes.json"), []byte("{}\n"))
		if Compare(dir, testContract, testCandidateContract, testExpected()).Valid {
			t.Fatal("undeclared evidence member accepted")
		}
	})
	t.Run("payload evidence member", func(t *testing.T) {
		dir, _ := makeEvidence(t)
		writeFile(t, filepath.Join(dir, "daemon-linux-amd64-1.oci.tar"), []byte("payload"))
		if Compare(dir, testContract, testCandidateContract, testExpected()).Valid {
			t.Fatal("payload evidence member accepted")
		}
	})
	t.Run("missing evidence member", func(t *testing.T) {
		dir, m := makeEvidence(t)
		if err := os.Remove(filepath.Join(dir, m.Current.CandidateReport.File)); err != nil {
			t.Fatal(err)
		}
		if Compare(dir, testContract, testCandidateContract, testExpected()).Valid {
			t.Fatal("missing evidence member accepted")
		}
	})
	t.Run("missing evidence directory", func(t *testing.T) {
		if Compare(filepath.Join(t.TempDir(), "missing"), testContract, testCandidateContract, testExpected()).Valid {
			t.Fatal("missing evidence directory accepted")
		}
	})
	t.Run("missing manifest", func(t *testing.T) {
		if Compare(t.TempDir(), testContract, testCandidateContract, testExpected()).Valid {
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
		{"schema", func(v map[string]any) { v["schema"] = "truerepublic.cross-run-rebuild/v2" }},
		{"repository", func(v map[string]any) { v["source"].(map[string]any)["repository"] = "example/fork" }},
		{"workflow path", func(v map[string]any) { v["source"].(map[string]any)["workflow_path"] = ".github/workflows/other.yml" }},
		{"branch", func(v map[string]any) { v["source"].(map[string]any)["branch"] = "develop" }},
		{"event", func(v map[string]any) { v["source"].(map[string]any)["event"] = "push" }},
		{"run ID grammar", func(v map[string]any) { v["source"].(map[string]any)["run_id_pattern"] = "^.*$" }},
		{"retention", func(v map[string]any) { v["source"].(map[string]any)["retention_days"] = json.Number("30") }},
		{"candidate digest", func(v map[string]any) { v["candidate_contract_sha256"] = strings.Repeat("0", 64) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			var value map[string]any
			if err := json.Unmarshal(mustRead(t, testContract), &value); err != nil {
				t.Fatal(err)
			}
			test.mutate(value)
			mutated := filepath.Join(t.TempDir(), "contract.json")
			writeJSON(t, mutated, value)
			if Compare(dir, mutated, testCandidateContract, testExpected()).Valid {
				t.Fatal("drifted contract accepted")
			}
		})
	}
	t.Run("candidate contract replaced", func(t *testing.T) {
		if Compare(dir, testContract, testContract, testExpected()).Valid {
			t.Fatal("wrong candidate contract accepted")
		}
	})
}

func TestDeterministicReport(t *testing.T) {
	dir, m := makeEvidence(t)
	first := mustJSON(t, Compare(dir, testContract, testCandidateContract, testExpected()))
	second := mustJSON(t, Compare(dir, testContract, testCandidateContract, testExpected()))
	if !bytes.Equal(first, second) {
		t.Fatal("repeated comparison reports differ")
	}
	m.Claims.LongTermHermetic = true
	writeJSON(t, filepath.Join(dir, "cross-run-evidence.json"), m)
	invalidFirst := mustJSON(t, Compare(dir, testContract, testCandidateContract, testExpected()))
	invalidSecond := mustJSON(t, Compare(dir, testContract, testCandidateContract, testExpected()))
	if !bytes.Equal(invalidFirst, invalidSecond) {
		t.Fatal("repeated violation reports differ")
	}
}

func TestRunCLI(t *testing.T) {
	dir, m := makeEvidence(t)
	expected := testExpected()
	baseArgs := []string{
		"compare", "--evidence", dir, "--contract", testContract, "--candidate-contract", testCandidateContract,
		"--expected-repository", expected.Repository, "--expected-workflow", expected.WorkflowPath,
		"--expected-branch", expected.Branch, "--expected-commit", expected.Commit,
		"--expected-baseline-run-id", expected.BaselineRunID, "--expected-current-run-id", expected.CurrentRunID,
	}
	var stdout, stderr bytes.Buffer
	if code := Run(nil, &stdout, &stderr); code != 2 {
		t.Fatal("missing subcommand accepted")
	}
	if code := Run([]string{"verify", "--evidence", dir}, &stdout, &stderr); code != 2 {
		t.Fatal("unknown subcommand accepted")
	}
	if code := Run(append(baseArgs, "--unknown", "x"), &stdout, &stderr); code != 2 {
		t.Fatal("unknown flag accepted")
	}
	if code := Run(append(baseArgs, "--output", "yaml"), &stdout, &stderr); code != 2 {
		t.Fatal("unknown output format accepted")
	}
	missingExpected := []string{"compare", "--evidence", dir}
	if code := Run(missingExpected, &stdout, &stderr); code != 2 {
		t.Fatal("missing expected values accepted")
	}
	stdout.Reset()
	if code := Run(append(baseArgs, "--output", "json"), &stdout, &stderr); code != 0 {
		t.Fatalf("valid evidence rejected: %s %s", stdout.String(), stderr.String())
	}
	var report Report
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &report); err != nil || !report.Valid ||
		report.BinaryTargets != 2 || report.OCITargets != 4 || report.BaselineRunID != testBaselineRunID {
		t.Fatalf("JSON report is invalid: %s", stdout.String())
	}
	if code := Run(append(baseArgs, "--output", "text"), failingWriter{}, &stderr); code != 1 {
		t.Fatal("text report write failure was ignored")
	}
	m.Claims.Production = true
	writeJSON(t, filepath.Join(dir, "cross-run-evidence.json"), m)
	stdout.Reset()
	if code := Run(append(baseArgs, "--output", "json"), &stdout, &stderr); code != 1 {
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
	oversized := []byte(`{"schema":"` + strings.Repeat("x", MaxJSONBytes) + `"}`)
	if err := parseBytes(oversized, &v); err == nil || !strings.Contains(err.Error(), "byte limit") {
		t.Fatalf("oversized valid JSON did not hit the byte limit: %v", err)
	}
	var claims Claims
	if parseBytes([]byte(`{"real_tag_created":false,"ref_pushed":false,"signed":false,"attested":false,"published":false,"deployed":false,"production":false}`), &claims) == nil {
		t.Fatal("incomplete cross-run claims accepted")
	}
}

// makeEvidence builds a valid synthetic cross-run evidence directory in a
// temporary location and returns it with the manifest for mutation.
func makeEvidence(t *testing.T) (string, Manifest) {
	t.Helper()
	dir := t.TempDir()
	contractHash, err := hashRegularFile(testContract, MaxJSONBytes)
	if err != nil {
		t.Fatal(err)
	}
	candidateHash, err := hashRegularFile(testCandidateContract, MaxJSONBytes)
	if err != nil {
		t.Fatal(err)
	}
	m := Manifest{
		Schema:         Schema,
		ContractSHA256: contractHash,
		Comparison: ComparisonIdentity{
			Repository:    "NeaBouli/TrueRepublic",
			WorkflowPath:  ".github/workflows/reproducible-daemon.yml",
			Branch:        "main",
			Commit:        testCommit,
			BaselineRunID: testBaselineRunID,
			CurrentRunID:  testCurrentRunID,
		},
		Claims: Claims{present: true},
	}
	m.Baseline = writeRunSide(t, dir, "baseline", testBaselineRunID, testBaselineCreate, "completed", "success", 990000001,
		"2026-08-30T10:20:00Z", "2026-09-13T10:20:00Z", candidateHash)
	m.Current = writeRunSide(t, dir, "current", testCurrentRunID, testCurrentCreate, "in_progress", "", 990000002,
		"2026-09-01T10:20:00Z", "2026-09-15T10:20:00Z", candidateHash)
	writeJSON(t, filepath.Join(dir, "cross-run-evidence.json"), m)
	return dir, m
}

func writeRunSide(t *testing.T, dir, side, runID, createdAt, status, conclusion string, artifactID int64, artifactCreated, artifactExpires, candidateHash string) RunBinding {
	t.Helper()
	receipt := Receipt{
		Schema:       ReceiptSchema,
		Repository:   "NeaBouli/TrueRepublic",
		WorkflowPath: ".github/workflows/reproducible-daemon.yml",
		Branch:       "main",
		Event:        "workflow_dispatch",
		RunID:        runID,
		RunAttempt:   json.Number("1"),
		HeadSHA:      testCommit,
		CreatedAt:    createdAt,
		Status:       status,
		Conclusion:   conclusion,
		Artifact: ArtifactBinding{
			Name:      candidateArtifactPfx + testCommit,
			ID:        json.Number(strconv.Itoa(int(artifactID))),
			Digest:    "sha256:" + sum([]byte("artifact-"+side)),
			CreatedAt: artifactCreated,
			ExpiresAt: artifactExpires,
			Expired:   false,
		},
	}
	manifest := makeCandidateManifest(candidateHash)
	report := candidate.Report{
		Schema:        candidate.ReportSchema,
		Valid:         true,
		Tag:           testTag,
		Commit:        testCommit,
		BinaryTargets: 2,
		OCITargets:    4,
		Violations:    []string{},
	}
	receiptFile := "receipt-" + side + ".json"
	manifestFile := "candidate-manifest-" + side + ".json"
	reportFile := "candidate-report-" + side + ".json"
	writeJSON(t, filepath.Join(dir, receiptFile), receipt)
	writeJSON(t, filepath.Join(dir, manifestFile), manifest)
	writeJSON(t, filepath.Join(dir, reportFile), report)
	return RunBinding{
		Receipt:           candidate.FileDigest{File: receiptFile, SHA256: fileHash(t, dir, receiptFile)},
		CandidateManifest: candidate.FileDigest{File: manifestFile, SHA256: fileHash(t, dir, manifestFile)},
		CandidateReport:   candidate.FileDigest{File: reportFile, SHA256: fileHash(t, dir, reportFile)},
	}
}

// makeCandidateManifest builds a GH-261-shaped candidate manifest with
// synthetic deterministic digests.
func makeCandidateManifest(contractHash string) candidate.Manifest {
	m := candidate.Manifest{
		Schema:         candidate.Schema,
		ContractSHA256: contractHash,
		Source:         candidate.SourceIdentity{Tag: testTag, Commit: testCommit},
		Claims:         candidate.Claims{},
	}
	for _, target := range expectedBinaryTargets() {
		digest := sum([]byte("binary-" + target.ID))
		m.BinaryTargets = append(m.BinaryTargets, candidate.BinaryTargetBinding{
			ID:             target.ID,
			Artifact:       target.Artifact,
			ArtifactSHA256: digest,
			Metadata:       candidate.FileDigest{File: "build-metadata-" + target.ID + ".json", SHA256: sum([]byte("metadata-" + target.ID))},
			Checksums:      candidate.FileDigest{File: "checksums-" + target.ID + ".sha256", SHA256: sum([]byte("checksums-" + target.ID))},
		})
	}
	for index, platform := range []string{"linux/amd64", "linux/arm64"} {
		suffix := []string{"amd64", "arm64"}[index]
		binding := candidate.OCIPlatformBinding{
			Platform: platform,
			Evidence: candidate.FileDigest{File: "oci-evidence-" + suffix + ".json", SHA256: sum([]byte("oci-evidence-" + suffix))},
			Report:   candidate.FileDigest{File: "oci-report-" + suffix + ".json", SHA256: sum([]byte("oci-report-" + suffix))},
		}
		for _, target := range []string{"daemon-linux-" + suffix, "client-web-linux-" + suffix} {
			binding.Targets = append(binding.Targets, candidate.OCIImageIdentity{
				ID:             target,
				IndexSHA256:    sum([]byte("index-" + target)),
				ManifestSHA256: sum([]byte("manifest-" + target)),
				ConfigSHA256:   sum([]byte("config-" + target)),
				LayerSHA256:    []string{sum([]byte("layer-" + target + "-0")), sum([]byte("layer-" + target + "-1"))},
			})
		}
		m.OCIPlatforms = append(m.OCIPlatforms, binding)
	}
	return m
}

func sideBinding(m *Manifest, side string) *RunBinding {
	if side == "baseline" {
		return &m.Baseline
	}
	return &m.Current
}

func mutateReceipt(t *testing.T, dir string, m *Manifest, side string, change func(map[string]any)) {
	t.Helper()
	binding := sideBinding(m, side)
	mutateJSON(t, filepath.Join(dir, binding.Receipt.File), change)
	binding.Receipt.SHA256 = fileHash(t, dir, binding.Receipt.File)
}

func mutateCandidateManifest(t *testing.T, dir string, m *Manifest, side string, change func(map[string]any)) {
	t.Helper()
	binding := sideBinding(m, side)
	mutateJSON(t, filepath.Join(dir, binding.CandidateManifest.File), change)
	binding.CandidateManifest.SHA256 = fileHash(t, dir, binding.CandidateManifest.File)
}

func mutateCandidateReport(t *testing.T, dir string, m *Manifest, side string, change func(map[string]any)) {
	t.Helper()
	binding := sideBinding(m, side)
	mutateJSON(t, filepath.Join(dir, binding.CandidateReport.File), change)
	binding.CandidateReport.SHA256 = fileHash(t, dir, binding.CandidateReport.File)
}

// mutateOCIIdentity changes one identity field, or the ordered layer vector
// when field is empty, of the first current-run OCI target.
func mutateOCIIdentity(t *testing.T, dir string, m *Manifest, field string, layers func([]any) []any) {
	t.Helper()
	mutateCandidateManifest(t, dir, m, "current", func(v map[string]any) {
		target := v["oci_platforms"].([]any)[0].(map[string]any)["targets"].([]any)[0].(map[string]any)
		if field != "" {
			target[field] = strings.Repeat("0", 64)
			return
		}
		target["layer_sha256"] = layers(target["layer_sha256"].([]any))
	})
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
