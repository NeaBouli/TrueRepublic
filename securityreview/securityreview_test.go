package securityreview

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestVerifyAcceptsBoundedRepositoryContract(t *testing.T) {
	root, scope, findings := validContract(t)
	if violations := Verify(root, scope, findings); len(violations) != 0 {
		t.Fatalf("valid contract rejected: %v", violations)
	}
}

func TestStrictParsersRejectMalformedTrailingUnknownAndOversizeJSON(t *testing.T) {
	root, scope, findings := validContract(t)
	scopeJSON := marshalJSON(t, scope)
	findingsJSON := marshalJSON(t, findings)
	invalidScopeRaw, err := os.ReadFile("../testdata/securityreview/invalid-claims/scope.json")
	if err != nil {
		t.Fatal(err)
	}
	invalidScope, err := ParseScope(invalidScopeRaw)
	if err != nil {
		t.Fatal(err)
	}
	assertViolation(t, Verify(root, invalidScope, findings), "external_independence_claim")
	invalidFindingsRaw, err := os.ReadFile("../testdata/securityreview/invalid-claims/findings.json")
	if err != nil {
		t.Fatal(err)
	}
	invalidFindings, err := ParseFindings(invalidFindingsRaw)
	if err != nil {
		t.Fatal(err)
	}
	assertViolation(t, Verify(root, scope, invalidFindings), "production_ready")

	for _, tc := range []struct {
		name  string
		data  []byte
		parse func([]byte) error
	}{
		{"scope malformed", scopeJSON[:len(scopeJSON)/2], func(data []byte) error { _, err := ParseScope(data); return err }},
		{"scope trailing", append(append([]byte{}, scopeJSON...), []byte(` {}`)...), func(data []byte) error { _, err := ParseScope(data); return err }},
		{"scope unknown", bytes.Replace(scopeJSON, []byte(`"production_ready":false`), []byte(`"unexpected":1,"production_ready":false`), 1), func(data []byte) error { _, err := ParseScope(data); return err }},
		{"findings malformed", findingsJSON[:len(findingsJSON)/2], func(data []byte) error { _, err := ParseFindings(data); return err }},
		{"findings trailing", append(append([]byte{}, findingsJSON...), []byte(` []`)...), func(data []byte) error { _, err := ParseFindings(data); return err }},
		{"findings unknown", bytes.Replace(findingsJSON, []byte(`"production_ready":false`), []byte(`"unexpected":1,"production_ready":false`), 1), func(data []byte) error { _, err := ParseFindings(data); return err }},
		{"scope oversize", bytes.Repeat([]byte("x"), MaxDocumentBytes+1), func(data []byte) error { _, err := ParseScope(data); return err }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.parse(tc.data); err == nil {
				t.Fatal("invalid document accepted")
			}
		})
	}
}

func TestVerifyRejectsScopeDriftAndUnsafeReferences(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*testing.T, string, *Scope)
		want   string
	}{
		{"true independence claim", func(_ *testing.T, _ string, scope *Scope) { scope.ExternalIndependenceClaim = boolPointer(true) }, "external_independence_claim"},
		{"missing independence claim", func(_ *testing.T, _ string, scope *Scope) { scope.ExternalIndependenceClaim = nil }, "explicitly false"},
		{"true production claim", func(_ *testing.T, _ string, scope *Scope) { scope.ProductionReady = boolPointer(true) }, "production_ready"},
		{"unknown schema", func(_ *testing.T, _ string, scope *Scope) { scope.Document.Version = "v2" }, "scope schema version"},
		{"duplicate in-scope path", func(_ *testing.T, _ string, scope *Scope) { scope.InScope = append(scope.InScope, scope.InScope[0]) }, "duplicates path"},
		{"unsafe in-scope path", func(_ *testing.T, _ string, scope *Scope) { scope.InScope[0] = "../escape" }, "unsafe repository-relative path"},
		{"missing in-scope path", func(_ *testing.T, _ string, scope *Scope) { scope.InScope[0] = "missing" }, "no such file"},
		{"duplicate invariant", func(_ *testing.T, _ string, scope *Scope) {
			scope.Invariants = append(scope.Invariants, scope.Invariants[0])
		}, "duplicate invariant ID"},
		{"invalid invariant ID", func(_ *testing.T, _ string, scope *Scope) { scope.Invariants[0].ID = "INV-1" }, "invalid invariant ID"},
		{"unknown threat", func(_ *testing.T, _ string, scope *Scope) { scope.Invariants[0].ThreatIDs[0] = "TM-ZKP-999" }, "unknown threat ID"},
		{"duplicate threat", func(_ *testing.T, _ string, scope *Scope) {
			scope.Invariants[0].ThreatIDs = append(scope.Invariants[0].ThreatIDs, scope.Invariants[0].ThreatIDs[0])
		}, "duplicates threat ID"},
		{"missing source anchor", func(_ *testing.T, _ string, scope *Scope) { scope.Invariants[0].SourceAnchor = "renamedInvariant" }, "anchor"},
		{"duplicate evidence", func(_ *testing.T, _ string, scope *Scope) { scope.Evidence = append(scope.Evidence, scope.Evidence[0]) }, "duplicate evidence ID"},
		{"unknown evidence kind", func(_ *testing.T, _ string, scope *Scope) { scope.Evidence[0].Kind = "opinion" }, "unknown kind"},
		{"missing evidence anchor", func(_ *testing.T, _ string, scope *Scope) { scope.Evidence[0].RequiredAnchor = "missing-test" }, "anchor"},
		{"oversize source", func(t *testing.T, root string, scope *Scope) {
			writeFile(t, root, scope.Invariants[0].SourcePath, bytes.Repeat([]byte("x"), MaxReferencedBytes+1))
		}, "exceeds"},
		{"symlink source", func(t *testing.T, root string, scope *Scope) {
			target := filepath.Join(root, "actual.go")
			if err := os.WriteFile(target, []byte("InvariantAnchor"), 0o600); err != nil {
				t.Fatal(err)
			}
			source := filepath.Join(root, scope.Invariants[0].SourcePath)
			if err := os.Remove(source); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(target, source); err != nil {
				t.Fatal(err)
			}
		}, "symlink is forbidden"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root, scope, findings := validContract(t)
			tc.mutate(t, root, &scope)
			assertViolation(t, Verify(root, scope, findings), tc.want)
		})
	}
}

func TestVerifyRejectsInvalidFindingLifecycle(t *testing.T) {
	root, scope, findings := validContract(t)
	findings.Findings[0].ThreatIDs = nil
	assertViolation(t, Verify(root, scope, findings), "has no threat IDs")
	findings.Findings[0].ThreatIDs = []string{"TM-ZKP-999"}
	assertViolation(t, Verify(root, scope, findings), "references unknown threat ID")
	findings.Findings[0].ThreatIDs = []string{"TM-CON-001", "TM-CON-001"}
	assertViolation(t, Verify(root, scope, findings), "duplicates threat ID")
	findings.Findings[0].ThreatIDs = []string{"TM-CON-001"}
	findings.Findings[0].VerificationSource = "external"
	assertViolation(t, Verify(root, scope, findings), "outside verified_closed status")

	tests := []struct {
		name   string
		mutate func(*Finding)
		want   string
	}{
		{"true independence claim", nil, "external_independence_claim"},
		{"duplicate ID", func(f *Finding) {}, "duplicate finding ID"},
		{"invalid ID", func(f *Finding) { f.ID = "finding-1" }, "invalid finding ID"},
		{"unknown severity", func(f *Finding) { f.Severity = "urgent" }, "unknown severity"},
		{"unknown status", func(f *Finding) { f.Status = "closed" }, "unknown status"},
		{"unknown source", func(f *Finding) { f.Source = "scanner" }, "unknown source"},
		{"missing owner", func(f *Finding) { f.Owner = "" }, "requires title and owner"},
		{"unsafe affected path", func(f *Finding) { f.AffectedPaths[0] = "/etc/passwd" }, "unsafe repository-relative path"},
		{"high closed without evidence", func(f *Finding) { f.Status = "verified_closed" }, "requires remediation evidence"},
		{"remediated without evidence", func(f *Finding) { f.Status = "remediated" }, "requires remediation evidence"},
		{"accepted risk missing dates", func(f *Finding) { f.Status = "accepted_risk" }, "valid accepted_on and expires"},
		{"accepted risk reversed", func(f *Finding) { f.Status = "accepted_risk"; f.AcceptedOn = "2026-09-13"; f.Expires = "2026-09-12" }, "expiry must be after"},
		{"accepted risk over 30 days", func(f *Finding) { f.Status = "accepted_risk"; f.AcceptedOn = "2026-09-13"; f.Expires = "2026-10-14" }, "at most 30 days"},
		{"open carries risk dates", func(f *Finding) { f.AcceptedOn = "2026-09-13"; f.Expires = "2026-10-13" }, "outside accepted_risk"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root, scope, findings := validContract(t)
			if tc.name == "true independence claim" {
				findings.ExternalIndependenceClaim = boolPointer(true)
			} else if tc.name == "duplicate ID" {
				findings.Findings = append(findings.Findings, findings.Findings[0])
			} else {
				tc.mutate(&findings.Findings[0])
			}
			assertViolation(t, Verify(root, scope, findings), tc.want)
		})
	}
}

func TestVerifyAcceptsEvidenceBackedClosureAndBoundedRisk(t *testing.T) {
	root, scope, findings := validContract(t)
	findings.Findings[0].Status = "verified_closed"
	findings.Findings[0].RemediationEvidence = []string{"evidence_test.go"}
	assertViolation(t, Verify(root, scope, findings), "must contain its finding ID")
	writeFile(t, root, "evidence_test.go", []byte("// SRF-CON-001\nfunc TestInvariantEvidence() {}\n"))
	assertViolation(t, Verify(root, scope, findings), "verification_source external")
	findings.Findings[0].VerificationSource = "internal"
	assertViolation(t, Verify(root, scope, findings), "verification_source external")
	findings.Findings[0].VerificationSource = "external"
	assertViolation(t, Verify(root, scope, findings), "requires external verification evidence")
	findings.Findings[0].VerificationEvidence = []string{"external-review.md"}
	writeFile(t, root, "external-review.md", []byte("Independent verification record without linked identifier.\n"))
	assertViolation(t, Verify(root, scope, findings), "external verification evidence must contain its finding ID")
	writeFile(t, root, "external-review.md", []byte("Independent verification: SRF-CON-001 verified on the reviewed commit.\n"))
	findings.Findings = append(findings.Findings, Finding{
		ID: "SRF-OPS-002", Severity: "medium", Status: "accepted_risk", Source: "internal",
		Title: "Bounded risk", ThreatIDs: []string{"TM-CON-001"}, AffectedPaths: []string{"module"}, Owner: "security-maintainers",
		AcceptedOn: "2026-09-13", Expires: "2026-10-13",
	})
	asOf := time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC)
	if violations := VerifyAt(root, scope, findings, asOf); len(violations) != 0 {
		t.Fatalf("valid finding states rejected: %v", violations)
	}
	assertViolation(t, VerifyAt(root, scope, findings, time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)), "expired on")
}

func TestVerifyFilesAndCommand(t *testing.T) {
	root, scope, findings := validContract(t)
	writeJSON(t, root, "scope.json", scope)
	writeJSON(t, root, "findings.json", findings)
	if err := VerifyFiles(root, "scope.json", "findings.json"); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"--repo-root", root, "--scope", "scope.json", "--findings", "findings.json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("Run code=%d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "independent review and production claims remain false") {
		t.Fatalf("unexpected output %q", stdout.String())
	}
	stderr.Reset()
	if code := Run([]string{"--repo-root", root}, &stdout, &stderr); code != 2 {
		t.Fatalf("missing flags code=%d, want 2", code)
	}
}

func TestReadScopeRejectsSymlinkAndOversizeFile(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target.json")
	if err := os.WriteFile(target, []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "scope.json")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadScope(link); err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("symlink read error=%v", err)
	}
	over := filepath.Join(root, "oversize.json")
	if err := os.WriteFile(over, bytes.Repeat([]byte("x"), MaxDocumentBytes+1), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadScope(over); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("oversize read error=%v", err)
	}
}

func validContract(t *testing.T) (string, Scope, Findings) {
	t.Helper()
	root := t.TempDir()
	writeFile(t, root, "configs/security/threat-model.json", []byte(`{"model":{"version":"truerepublic.threat-model/v1"},"threats":[{"id":"TM-CON-001"}]}`))
	writeFile(t, root, "configs/security/gates.json", []byte(`{"version":"truerepublic.security-gates/v1"}`))
	writeFile(t, root, "module/invariant.go", []byte("package module\nconst InvariantAnchor = true\n"))
	writeFile(t, root, "evidence_test.go", []byte("func TestInvariantEvidence() {}\n"))
	scope := Scope{
		Document:                  Identity{Name: "test-security-review-scope", Version: ScopeSchemaVersion, Updated: "2026-09-13"},
		ExternalIndependenceClaim: boolPointer(false),
		ProductionReady:           boolPointer(false),
		ThreatModelPath:           "configs/security/threat-model.json", SecurityGatePath: "configs/security/gates.json",
		InScope:    []string{"module"},
		Invariants: []Invariant{{ID: "SR-INV-001", Description: "test invariant", SourcePath: "module/invariant.go", SourceAnchor: "InvariantAnchor", ThreatIDs: []string{"TM-CON-001"}}},
		Evidence:   []Evidence{{ID: "SR-EVD-001", Path: "evidence_test.go", Kind: "contract_test", RequiredAnchor: "TestInvariantEvidence"}},
	}
	findings := Findings{
		Document:                  Identity{Name: "test-security-review-findings", Version: FindingsSchemaVersion, Updated: "2026-09-13"},
		ExternalIndependenceClaim: boolPointer(false),
		ProductionReady:           boolPointer(false),
		Findings:                  []Finding{{ID: "SRF-CON-001", Severity: "high", Status: "open", Source: "internal", Title: "Open test limitation", ThreatIDs: []string{"TM-CON-001"}, AffectedPaths: []string{"module"}, Owner: "security-maintainers"}},
	}
	return root, scope, findings
}

func boolPointer(value bool) *bool { return &value }

func writeFile(t *testing.T, root, rel string, data []byte) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func writeJSON(t *testing.T, root, rel string, value any) {
	t.Helper()
	writeFile(t, root, rel, marshalJSON(t, value))
}

func marshalJSON(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func assertViolation(t *testing.T, violations []string, want string) {
	t.Helper()
	if !strings.Contains(strings.Join(violations, "\n"), want) {
		t.Fatalf("missing violation %q in %v", want, violations)
	}
}
