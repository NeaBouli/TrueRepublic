package main

import (
	"os"
	"strings"
	"testing"

	"truerepublic/securityreview"
)

const (
	securityReviewScopePath    = "configs/security/review-scope.json"
	securityReviewFindingsPath = "configs/security/review-findings.json"
)

func TestSecurityReviewRepositoryContract(t *testing.T) {
	if err := securityreview.VerifyFiles(".", securityReviewScopePath, securityReviewFindingsPath); err != nil {
		t.Fatal(err)
	}

	scopeRaw := readSecurityReviewFile(t, securityReviewScopePath)
	findingsRaw := readSecurityReviewFile(t, securityReviewFindingsPath)
	scope, err := securityreview.ParseScope(scopeRaw)
	if err != nil {
		t.Fatal(err)
	}
	findings, err := securityreview.ParseFindings(findingsRaw)
	if err != nil {
		t.Fatal(err)
	}

	requiredInvariants := map[string]bool{
		"SR-INV-001": false,
		"SR-INV-002": false,
		"SR-INV-003": false,
		"SR-INV-004": false,
		"SR-INV-005": false,
	}
	for _, invariant := range scope.Invariants {
		if _, ok := requiredInvariants[invariant.ID]; ok {
			requiredInvariants[invariant.ID] = true
		}
	}
	for id, present := range requiredInvariants {
		if !present {
			t.Fatalf("review scope missing required invariant %s", id)
		}
	}

	requiredFindings := map[string]bool{
		"SRF-REV-001": false,
		"SRF-ZKP-001": false,
		"SRF-CLI-001": false,
		"SRF-OPS-001": false,
		"SRF-REL-001": false,
	}
	for _, finding := range findings.Findings {
		if _, ok := requiredFindings[finding.ID]; ok {
			requiredFindings[finding.ID] = true
		}
	}
	for id, present := range requiredFindings {
		if !present {
			t.Fatalf("findings register missing known limitation %s", id)
		}
	}

	requiredDocs := map[string][]string{
		"docs/security/INDEPENDENT_REVIEW_GUIDE.md": {
			securityreview.ScopeSchemaVersion,
			securityreview.FindingsSchemaVersion,
			"not an independent audit",
			"external_independence_claim=false",
			"production_ready=false",
			"make security-review-contract-test",
			"verification_source=external",
			"clean, immutable checkout",
			"complete threat-model and security-gate repository contracts",
		},
		"CODEX_AUDIT.md": {
			"Historical snapshot — superseded for current status",
			"docs/security/INDEPENDENT_REVIEW_GUIDE.md",
		},
		"docs/ROLLOUT_ROADMAP.md": {
			"GH-294 provides a versioned repository-prepared scope",
			"grants no rollout credit",
		},
		"README.md": {
			"Independent Review Guide and Findings Contract",
		},
		"wiki/security/Audit-Reports.md": {
			"GH-294 adds a machine-verifiable",
			"does not claim reviewer independence",
		},
	}
	for path, anchors := range requiredDocs {
		contents := string(readSecurityReviewFile(t, path))
		for _, anchor := range anchors {
			if !strings.Contains(contents, anchor) {
				t.Fatalf("%s must contain %q", path, anchor)
			}
		}
	}

	wiring := map[string][]string{
		"Makefile": {
			"security-review-contract-test:",
			"TestSecurityReviewRepositoryContract",
			"TestThreatModelRepositoryContract",
			"TestSecurityGateRepositoryContract",
			"go run ./cmd/security-review",
		},
		".github/workflows/security-scan.yml": {
			"Verify security-review readiness contract",
			"make security-review-contract-test",
		},
		"scripts/check-consistency.sh": {
			"Checking security-review readiness contract",
			"go run ./cmd/security-review",
		},
	}
	for path, anchors := range wiring {
		contents := string(readSecurityReviewFile(t, path))
		for _, anchor := range anchors {
			if !strings.Contains(contents, anchor) {
				t.Fatalf("%s must contain %q", path, anchor)
			}
		}
	}

	t.Run("rejects omitted or promoted safety claims", func(t *testing.T) {
		promoted := true
		scope.ExternalIndependenceClaim = &promoted
		if got := strings.Join(securityreview.Verify(".", scope, findings), "\n"); !strings.Contains(got, "external_independence_claim") {
			t.Fatalf("promoted independence claim accepted: %s", got)
		}
		scope.ExternalIndependenceClaim = nil
		if got := strings.Join(securityreview.Verify(".", scope, findings), "\n"); !strings.Contains(got, "explicitly false") {
			t.Fatalf("omitted independence claim accepted: %s", got)
		}
	})
}

func readSecurityReviewFile(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
