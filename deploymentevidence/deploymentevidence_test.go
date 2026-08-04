package deploymentevidence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var evaluationTime = time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)

func TestParseManifestStrict(t *testing.T) {
	fixture := readFixture(t, filepath.Join("deployment", "evidence.example.json"))

	cases := []struct {
		name     string
		input    []byte
		wantErr  bool
		contains string
	}{
		{name: "valid fixture", input: fixture},
		{name: "empty", input: []byte(" "), contains: "deployment manifest is empty"},
		{name: "trailing", input: []byte("{}\nnull"), contains: "trailing value is forbidden"},
		{name: "duplicate key", input: []byte(`{"version":"x","version":"y"}`), contains: "duplicate object key"},
		{name: "unknown field", input: []byte(`{"version":"` + ManifestVersion + `","unexpected_field":"x"}`), contains: "invalid deployment manifest schema"},
		{name: "max depth", input: []byte(multiDepthJSON(33)), contains: "maximum JSON depth exceeded"},
		{name: "oversize", input: bytes.Repeat([]byte(" "), MaxManifestBytes+1), contains: "deployment manifest exceeds"},
		{name: "non-object array", input: []byte(`["version"]`), contains: "invalid deployment manifest schema"},
		{name: "non-object scalar", input: []byte(`42`), contains: "invalid deployment manifest schema"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseManifest(bytes.NewReader(tc.input))
			if tc.wantErr || tc.contains != "" {
				if err == nil {
					t.Fatal("expected parse failure")
				}
				if tc.contains != "" && !strings.Contains(err.Error(), tc.contains) {
					t.Fatalf("expected %q in %q", tc.contains, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected parse failure: %v", err)
			}
		})
	}

	t.Run("no reflection of rejected keys or values", func(t *testing.T) {
		inputs := [][]byte{
			[]byte(`{"version":"x","super_secret_field":"super-secret-token"}`),
			[]byte(`{"version":"x","secretduplicate":"a","secretduplicate":"b"}`),
			[]byte(`{"version":`),
		}
		for _, input := range inputs {
			_, err := ParseManifest(bytes.NewReader(input))
			if err == nil {
				t.Fatal("expected parse failure")
			}
			for _, planted := range []string{"super_secret_field", "super-secret-token", "secretduplicate"} {
				if strings.Contains(err.Error(), planted) {
					t.Fatalf("error %q reflects rejected input %q", err, planted)
				}
			}
		}
	})
}

func TestLoadTopology(t *testing.T) {
	t.Run("valid fixture", func(t *testing.T) {
		topology := loadFixtureTopology(t)
		if len(topology.SHA256) != 64 {
			t.Fatalf("expected 64 hex digest, got %q", topology.SHA256)
		}
		if topology.NodeCount != 5 {
			t.Fatalf("expected 5 nodes, got %d", topology.NodeCount)
		}
		want := RoleCounts{Seed: 1, Sentry: 2, Validator: 1, RPC: 1}
		if topology.RoleCounts != want {
			t.Fatalf("expected role counts %+v, got %+v", want, topology.RoleCounts)
		}
		if topology.ChainID == "" {
			t.Fatal("expected derived chain ID")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := LoadTopology(filepath.Join(t.TempDir(), "does-not-exist.json"))
		if err == nil || !strings.Contains(err.Error(), "file does not exist") {
			t.Fatalf("expected generic missing-file error, got %v", err)
		}
	})

	t.Run("invalid topology yields generic fixed errors", func(t *testing.T) {
		dir := t.TempDir()
		parseBroken := filepath.Join(dir, "parse-broken.json")
		if err := os.WriteFile(parseBroken, []byte(`{"version":"wrong","nodes":`), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadTopology(parseBroken); err == nil ||
			err.Error() != "invalid topology contract" {
			t.Fatalf("expected fixed parse error, got %v", err)
		}

		validationBroken := filepath.Join(dir, "validation-broken.json")
		raw := readFixture(t, filepath.Join("topology", "qualification.example.json"))
		raw = bytes.Replace(raw, []byte(`"inbound": "deny"`), []byte(`"inbound": "allow"`), 1)
		if err := os.WriteFile(validationBroken, raw, 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := LoadTopology(validationBroken)
		if err == nil || err.Error() != "topology contract validation failed" {
			t.Fatalf("expected fixed validation error, got %v", err)
		}
		if strings.Contains(err.Error(), "inbound") || strings.Contains(err.Error(), "deny") {
			t.Fatalf("topology validator message reflected: %v", err)
		}
	})
}

func TestVerifyValidFixture(t *testing.T) {
	manifest := loadFixtureManifest(t)
	topology := loadFixtureTopology(t)

	report := Verify(manifest, topology, evaluationTime)
	if !report.Valid {
		t.Fatalf("expected valid fixture: %+v", report)
	}
	if report.Violations == nil {
		t.Fatal("valid report must encode violations as an empty array")
	}
	if report.GateCount != len(GateIDs) || report.NodeCount != 5 {
		t.Fatalf("unexpected report counts: %+v", report)
	}
	encoded, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"violations":[]`) {
		t.Fatalf("empty violations must serialize as []: %s", encoded)
	}
	for _, forbidden := range []string{topology.ChainID, topology.SHA256, manifest.PreparedBy} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("report reflects forbidden value %q: %s", forbidden, encoded)
		}
	}
}

func TestVerifyBinding(t *testing.T) {
	topology := loadFixtureTopology(t)

	cases := []struct {
		name   string
		mutate func(*Manifest)
		check  string
	}{
		{name: "digest format", mutate: func(m *Manifest) { m.TopologySHA256 = "not-a-digest" }, check: "topology_sha256.format"},
		{name: "digest mismatch", mutate: func(m *Manifest) { m.TopologySHA256 = strings.Repeat("a", 64) }, check: "topology_sha256.binding"},
		{name: "chain mismatch", mutate: func(m *Manifest) { m.ChainID = "other-chain-1" }, check: "chain_id"},
		{name: "count mismatch", mutate: func(m *Manifest) { m.NodeCount = 6 }, check: "node_count"},
		{name: "role mismatch", mutate: func(m *Manifest) { m.RoleCounts.Sentry = 1 }, check: "role_counts"},
		{name: "version mismatch", mutate: func(m *Manifest) { m.Version = "wrong/version" }, check: "version"},
		{name: "preparer not canonical", mutate: func(m *Manifest) { m.PreparedBy = "Not_A_Seat!" }, check: "prepared_by"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			manifest := loadFixtureManifest(t)
			tc.mutate(&manifest)
			report := Verify(manifest, topology, evaluationTime)
			if report.Valid {
				t.Fatal("expected invalid manifest")
			}
			if !containsCheck(report.Violations, tc.check) {
				t.Fatalf("expected check %q, got %#v", tc.check, report.Violations)
			}
		})
	}
}

func TestVerifyReportCountsComeFromTrustedPolicy(t *testing.T) {
	manifest := loadFixtureManifest(t)
	topology := loadFixtureTopology(t)
	manifest.NodeCount = -1
	manifest.Gates = manifest.Gates[:1]

	report := Verify(manifest, topology, evaluationTime)
	if report.Valid {
		t.Fatal("expected invalid manifest")
	}
	if report.NodeCount != topology.NodeCount || report.GateCount != len(GateIDs) {
		t.Fatalf("report reflected untrusted manifest counts: %+v", report)
	}
}

func TestVerifyGates(t *testing.T) {
	topology := loadFixtureTopology(t)

	cases := []struct {
		name   string
		mutate func(*Manifest)
		check  string
	}{
		{name: "missing gate", mutate: func(m *Manifest) { m.Gates = m.Gates[:len(GateIDs)-1] }, check: "gates.count"},
		{name: "extra gate", mutate: func(m *Manifest) { m.Gates = append(m.Gates, m.Gates[0]) }, check: "gates.count"},
		{name: "duplicate gate id", mutate: func(m *Manifest) { m.Gates[1].ID = m.Gates[0].ID }, check: "gates[1].id"},
		{name: "unknown gate id", mutate: func(m *Manifest) { m.Gates[3].ID = "unknown-gate" }, check: "gates[3].id"},
		{name: "out of order gates", mutate: func(m *Manifest) { m.Gates[0], m.Gates[1] = m.Gates[1], m.Gates[0] }, check: "gates[0].id"},
		{name: "failed gate", mutate: func(m *Manifest) { m.Gates[0].Result = "failed" }, check: "gates[0].result"},
		{name: "malformed digest", mutate: func(m *Manifest) { m.Gates[2].EvidenceSHA256 = "zzzz" }, check: "gates[2].evidence_sha256"},
		{name: "duplicate digest", mutate: func(m *Manifest) { m.Gates[2].EvidenceSHA256 = m.Gates[0].EvidenceSHA256 }, check: "gates[2].evidence_sha256"},
		{name: "started after completed", mutate: func(m *Manifest) {
			m.Gates[0].StartedAt = "2026-07-30T08:25:00Z"
			m.Gates[0].CompletedAt = "2026-07-30T08:00:00Z"
		}, check: "gates[0].order"},
		{name: "completed after preparation", mutate: func(m *Manifest) {
			m.Gates[0].StartedAt = "2026-07-31T09:00:00Z"
			m.Gates[0].CompletedAt = "2026-07-31T09:30:00Z"
		}, check: "gates[0].order"},
		{name: "stale gate", mutate: func(m *Manifest) {
			m.Gates[0].StartedAt = "2026-06-20T08:00:00Z"
			m.Gates[0].CompletedAt = "2026-06-20T08:25:00Z"
		}, check: "gates[0].completed_at"},
		{name: "future gate", mutate: func(m *Manifest) {
			m.Gates[0].StartedAt = "2026-08-01T12:00:00Z"
			m.Gates[0].CompletedAt = "2026-08-01T12:10:00Z"
			m.PreparedAt = "2026-08-01T12:04:00Z"
			for i := range m.Approvals {
				m.Approvals[i].ApprovedAt = "2026-08-01T12:04:00Z"
			}
		}, check: "gates[0].completed_at"},
		{name: "non-canonical gate timestamp", mutate: func(m *Manifest) {
			m.Gates[0].StartedAt = "2026-07-30T8:00:00Z"
		}, check: "gates[0].started_at"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			manifest := loadFixtureManifest(t)
			tc.mutate(&manifest)
			report := Verify(manifest, topology, evaluationTime)
			if report.Valid {
				t.Fatal("expected invalid manifest")
			}
			if !containsCheck(report.Violations, tc.check) {
				t.Fatalf("expected check %q, got %#v", tc.check, report.Violations)
			}
		})
	}
}

func TestVerifyTimestamps(t *testing.T) {
	topology := loadFixtureTopology(t)

	cases := []struct {
		name  string
		value string
		check string
	}{
		{name: "invalid format", value: "2026-07-31 09:00:00Z", check: "prepared_at.format"},
		{name: "offset not UTC canonical", value: "2026-07-31T09:00:00+00:00", check: "prepared_at.format"},
		{name: "out of range day", value: "2026-02-30T09:00:00Z", check: "prepared_at.format"},
		{name: "stale", value: "2026-07-01T11:59:59Z", check: "prepared_at"},
		{name: "future beyond skew", value: "2026-08-01T12:05:01Z", check: "prepared_at"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			manifest := loadFixtureManifest(t)
			manifest.PreparedAt = tc.value
			report := Verify(manifest, topology, evaluationTime)
			if report.Valid {
				t.Fatal("expected invalid manifest")
			}
			if !containsCheck(report.Violations, tc.check) {
				t.Fatalf("expected check %q, got %#v", tc.check, report.Violations)
			}
		})
	}

	t.Run("boundary prepared_at within skew passes freshness", func(t *testing.T) {
		manifest := loadFixtureManifest(t)
		manifest.PreparedAt = "2026-08-01T12:04:59Z"
		for i := range manifest.Gates {
			manifest.Gates[i].StartedAt = "2026-07-30T08:00:00Z"
			manifest.Gates[i].CompletedAt = "2026-07-30T08:25:00Z"
		}
		for i := range manifest.Approvals {
			manifest.Approvals[i].ApprovedAt = "2026-08-01T12:04:59Z"
		}
		report := Verify(manifest, topology, evaluationTime)
		if !report.Valid {
			t.Fatalf("expected boundary timestamps to pass: %+v", report)
		}
	})
}

func TestVerifyApprovals(t *testing.T) {
	topology := loadFixtureTopology(t)

	cases := []struct {
		name   string
		mutate func(*Manifest)
		check  string
	}{
		{name: "missing approval", mutate: func(m *Manifest) { m.Approvals = m.Approvals[:1] }, check: "approvals.count"},
		{name: "extra approval", mutate: func(m *Manifest) { m.Approvals = append(m.Approvals, m.Approvals[0]) }, check: "approvals.count"},
		{name: "same seat", mutate: func(m *Manifest) { m.Approvals[1].Seat = m.Approvals[0].Seat }, check: "approvals[1].seat"},
		{name: "same role", mutate: func(m *Manifest) { m.Approvals[1].Role = ApprovalRoleOperator }, check: "approvals[1].role"},
		{name: "unknown role", mutate: func(m *Manifest) { m.Approvals[1].Role = "auditor" }, check: "approvals[1].role"},
		{name: "preparer reuse", mutate: func(m *Manifest) { m.Approvals[0].Seat = m.PreparedBy }, check: "approvals[0].seat"},
		{name: "bad binding", mutate: func(m *Manifest) { m.Approvals[0].TopologySHA256 = strings.Repeat("b", 64) }, check: "approvals[0].topology_sha256"},
		{name: "malformed binding", mutate: func(m *Manifest) { m.Approvals[0].TopologySHA256 = "xyz" }, check: "approvals[0].topology_sha256"},
		{name: "approved before preparation", mutate: func(m *Manifest) { m.Approvals[0].ApprovedAt = "2026-07-31T08:59:59Z" }, check: "approvals[0].approved_at"},
		{name: "future approval", mutate: func(m *Manifest) { m.Approvals[0].ApprovedAt = "2026-08-01T12:05:01Z" }, check: "approvals[0].approved_at"},
		{name: "non-canonical approval time", mutate: func(m *Manifest) { m.Approvals[0].ApprovedAt = "2026-07-31T10:00:00" }, check: "approvals[0].approved_at"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			manifest := loadFixtureManifest(t)
			tc.mutate(&manifest)
			report := Verify(manifest, topology, evaluationTime)
			if report.Valid {
				t.Fatal("expected invalid manifest")
			}
			if !containsCheck(report.Violations, tc.check) {
				t.Fatalf("expected check %q, got %#v", tc.check, report.Violations)
			}
		})
	}
}

func TestVerifyNoReflection(t *testing.T) {
	topology := loadFixtureTopology(t)
	manifest := loadFixtureManifest(t)
	manifest.ChainID = "secret-chain-value"
	manifest.PreparedBy = "Secret_Seat_Planted"
	manifest.Gates[0].Result = "secretly-passed"
	manifest.Gates[1].StartedAt = "planted-timestamp"
	manifest.Gates[2].EvidenceSHA256 = "planted-digest"
	manifest.Approvals[0].Seat = "Secret_Approver"
	manifest.Approvals[1].ApprovedAt = "planted-approval-time"

	report := Verify(manifest, topology, evaluationTime)
	if report.Valid {
		t.Fatal("expected invalid manifest")
	}
	encoded, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	for _, planted := range []string{
		"secret-chain-value", "Secret_Seat_Planted", "secretly-passed",
		"planted-timestamp", "planted-digest", "Secret_Approver", "planted-approval-time",
	} {
		if strings.Contains(string(encoded), planted) {
			t.Fatalf("report reflects rejected value %q: %s", planted, encoded)
		}
	}
}

func TestDeploymentEvidenceCommands(t *testing.T) {
	dir := t.TempDir()
	contractPath := filepath.Join(dir, "contract.json")
	if err := os.WriteFile(contractPath, readFixture(t, filepath.Join("topology", "qualification.example.json")), 0o600); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(dir, "manifest.json")
	if err := os.WriteFile(manifestPath, readFixture(t, filepath.Join("deployment", "evidence.example.json")), 0o600); err != nil {
		t.Fatal(err)
	}

	badManifest := loadFixtureManifest(t)
	badManifest.Gates[0].Result = "failed"
	badManifestJSON, err := json.MarshalIndent(badManifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	badManifestPath := filepath.Join(dir, "manifest-bad.json")
	if err := os.WriteFile(badManifestPath, badManifestJSON, 0o600); err != nil {
		t.Fatal(err)
	}

	brokenTopologyPath := filepath.Join(dir, "topology-broken.json")
	if err := os.WriteFile(brokenTopologyPath, []byte(`{"version":"wrong"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	sensitivePath := filepath.Join(dir, "secret-identifier.json")
	sensitiveManifest := []byte(`{"version":"x","super_secret_field":"super-secret-token"}`)
	if err := os.WriteFile(sensitivePath, sensitiveManifest, 0o600); err != nil {
		t.Fatal(err)
	}

	at := "2026-08-01T12:00:00Z"

	t.Run("json success", func(t *testing.T) {
		out, err := runDeploymentCommand(t, "verify", "--contract", contractPath,
			"--manifest", manifestPath, "--at", at, "--output", "json")
		if err != nil {
			t.Fatalf("expected success: %v\n%s", err, out)
		}
		if !strings.Contains(out, `"valid": true`) || !strings.Contains(out, `"violations": []`) {
			t.Fatalf("unexpected report output: %s", out)
		}
	})

	t.Run("text success", func(t *testing.T) {
		out, err := runDeploymentCommand(t, "verify", "--contract", contractPath,
			"--manifest", manifestPath, "--at", at, "--output", "text")
		if err != nil {
			t.Fatalf("expected success: %v\n%s", err, out)
		}
		if !strings.Contains(out, "OK: deployment evidence validates 11 gates across 5 nodes") {
			t.Fatalf("unexpected text output: %s", out)
		}
	})

	t.Run("json failure", func(t *testing.T) {
		out, err := runDeploymentCommand(t, "verify", "--contract", contractPath,
			"--manifest", badManifestPath, "--at", at, "--output", "json")
		if err == nil {
			t.Fatal("expected command failure")
		}
		var report Report
		if err := json.Unmarshal([]byte(out), &report); err != nil {
			t.Fatalf("decode report: %v\n%s", err, out)
		}
		if report.Valid || !containsCheck(report.Violations, "gates[0].result") {
			t.Fatalf("expected structured invalid report: %+v", report)
		}
	})

	t.Run("text failure", func(t *testing.T) {
		out, err := runDeploymentCommand(t, "verify", "--contract", contractPath,
			"--manifest", badManifestPath, "--at", at, "--output", "text")
		if err == nil {
			t.Fatal("expected command failure")
		}
		if !strings.Contains(out, "VIOLATION gates[0].result:") {
			t.Fatalf("unexpected text output: %s", out)
		}
	})

	t.Run("missing flags", func(t *testing.T) {
		out, err := runDeploymentCommand(t, "verify")
		if err == nil {
			t.Fatal("expected command failure")
		}
		if strings.Contains(out, "OK:") {
			t.Fatalf("unexpected success output: %s", out)
		}
	})

	t.Run("strict --at", func(t *testing.T) {
		rejected := "2026-08-01 12:00"
		out, err := runDeploymentCommand(t, "verify", "--contract", contractPath,
			"--manifest", manifestPath, "--at", rejected)
		if err == nil || !strings.Contains(err.Error(), "invalid --at") {
			t.Fatalf("expected strict --at failure, got %v", err)
		}
		if strings.Contains(err.Error(), rejected) || strings.Contains(out, rejected) {
			t.Fatalf("rejected --at value reflected: %v %s", err, out)
		}
	})

	t.Run("unknown output not echoed", func(t *testing.T) {
		out, err := runDeploymentCommand(t, "verify", "--contract", contractPath,
			"--manifest", manifestPath, "--at", at, "--output", "yaml")
		if err == nil || !strings.Contains(err.Error(), "unknown --output: expected text or json") {
			t.Fatalf("expected unknown output failure, got %v", err)
		}
		if strings.Contains(err.Error(), "yaml") || strings.Contains(out, "yaml") {
			t.Fatalf("rejected output value reflected: %v %s", err, out)
		}
	})

	t.Run("invalid topology generic error", func(t *testing.T) {
		_, err := runDeploymentCommand(t, "verify", "--contract", brokenTopologyPath,
			"--manifest", manifestPath, "--at", at)
		if err == nil {
			t.Fatal("expected command failure")
		}
		if err.Error() != "invalid topology contract" &&
			err.Error() != "topology contract validation failed" {
			t.Fatalf("expected fixed topology error, got %v", err)
		}
	})

	t.Run("no reflection of secrets or paths", func(t *testing.T) {
		out, err := runDeploymentCommand(t, "verify", "--contract", contractPath,
			"--manifest", sensitivePath, "--at", at, "--output", "json")
		if err == nil {
			t.Fatal("expected command failure")
		}
		combined := out + err.Error()
		for _, planted := range []string{"super-secret-token", "super_secret_field", "secret-identifier.json"} {
			if strings.Contains(combined, planted) {
				t.Fatalf("output reflects %q: %s", planted, combined)
			}
		}
	})

	t.Run("unknown flag not echoed", func(t *testing.T) {
		out, err := runDeploymentCommand(t, "verify", "--contract", contractPath,
			"--manifest", manifestPath, "--at", at, "--planted-flag", "planted-value")
		if err == nil {
			t.Fatal("expected command failure")
		}
		combined := out + err.Error()
		if strings.Contains(combined, "planted-flag") || strings.Contains(combined, "planted-value") {
			t.Fatalf("rejected flag reflected: %s", combined)
		}
	})

	t.Run("config independence", func(t *testing.T) {
		t.Chdir(t.TempDir())
		out, err := runDeploymentCommand(t, "verify", "--contract", contractPath,
			"--manifest", manifestPath, "--at", at, "--output", "text")
		if err != nil {
			t.Fatalf("expected success from unrelated working directory: %v\n%s", err, out)
		}
		if !strings.Contains(out, "OK:") {
			t.Fatalf("unexpected output: %s", out)
		}
	})
}

func runDeploymentCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := NewCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func loadFixtureManifest(t *testing.T) Manifest {
	t.Helper()
	manifest, err := ParseManifest(bytes.NewReader(readFixture(t, filepath.Join("deployment", "evidence.example.json"))))
	if err != nil {
		t.Fatalf("parse manifest fixture: %v", err)
	}
	return manifest
}

func loadFixtureTopology(t *testing.T) Topology {
	t.Helper()
	topology, err := LoadTopology(filepath.Join("..", "configs", "topology", "qualification.example.json"))
	if err != nil {
		t.Fatalf("load topology fixture: %v", err)
	}
	return topology
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("..", "configs", name)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	return raw
}

func containsCheck(violations []Violation, check string) bool {
	for _, violation := range violations {
		if violation.Check == check {
			return true
		}
	}
	return false
}

func multiDepthJSON(depth int) string {
	out := "{}"
	for i := 0; i < depth; i++ {
		out = fmt.Sprintf(`{"a":%s}`, out)
	}
	return out
}
