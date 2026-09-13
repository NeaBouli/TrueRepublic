package securityreview

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	datePattern      = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)
	scopeIDPattern   = regexp.MustCompile(`^SR-(INV|EVD)-[0-9]{3}$`)
	findingIDPattern = regexp.MustCompile(`^SRF-[A-Z]{3}-[0-9]{3}$`)
	threatIDPattern  = regexp.MustCompile(`^TM-[A-Z]{3}-[0-9]{3}$`)
)

var evidenceKinds = map[string]bool{
	"audit_record": true, "configuration": true, "contract_test": true,
	"policy": true, "runbook": true, "test": true, "workflow": true,
}

func VerifyFiles(repoRoot, scopePath, findingsPath string) error {
	root, err := canonicalRoot(repoRoot)
	if err != nil {
		return err
	}
	scopeFile, err := resolvePath(root, scopePath, true)
	if err != nil {
		return fmt.Errorf("scope path: %w", err)
	}
	findingsFile, err := resolvePath(root, findingsPath, true)
	if err != nil {
		return fmt.Errorf("findings path: %w", err)
	}
	scope, err := ReadScope(scopeFile)
	if err != nil {
		return err
	}
	findings, err := ReadFindings(findingsFile)
	if err != nil {
		return err
	}
	violations := Verify(root, scope, findings)
	if len(violations) != 0 {
		return fmt.Errorf("security review contract violations:\n- %s", strings.Join(violations, "\n- "))
	}
	return nil
}

func Verify(repoRoot string, scope Scope, findings Findings) []string {
	return VerifyAt(repoRoot, scope, findings, time.Now().UTC())
}

func VerifyAt(repoRoot string, scope Scope, findings Findings, now time.Time) []string {
	root, err := canonicalRoot(repoRoot)
	if err != nil {
		return []string{err.Error()}
	}
	var violations []string
	violations = append(violations, verifyIdentity(scope.Document, ScopeSchemaVersion, "scope")...)
	violations = append(violations, verifyIdentity(findings.Document, FindingsSchemaVersion, "findings")...)
	if scope.ExternalIndependenceClaim == nil || *scope.ExternalIndependenceClaim || findings.ExternalIndependenceClaim == nil || *findings.ExternalIndependenceClaim {
		violations = append(violations, "external_independence_claim must be explicitly false")
	}
	if scope.ProductionReady == nil || *scope.ProductionReady || findings.ProductionReady == nil || *findings.ProductionReady {
		violations = append(violations, "production_ready must be explicitly false")
	}

	threats, err := loadThreatIDs(root, scope.ThreatModelPath)
	if err != nil {
		violations = append(violations, "threat model: "+err.Error())
	}
	if scope.SecurityGatePath != "configs/security/gates.json" {
		violations = append(violations, "security gate path must be configs/security/gates.json")
	} else if err := verifySecurityGate(root, scope.SecurityGatePath); err != nil {
		violations = append(violations, "security gate path: "+err.Error())
	}
	if len(scope.InScope) == 0 {
		violations = append(violations, "in_scope must not be empty")
	}
	violations = append(violations, verifyPaths(root, "in_scope", scope.InScope, false)...)

	seenInvariant := map[string]bool{}
	for _, invariant := range scope.Invariants {
		if !scopeIDPattern.MatchString(invariant.ID) || !strings.HasPrefix(invariant.ID, "SR-INV-") {
			violations = append(violations, fmt.Sprintf("invalid invariant ID %q", invariant.ID))
		}
		if seenInvariant[invariant.ID] {
			violations = append(violations, fmt.Sprintf("duplicate invariant ID %q", invariant.ID))
		}
		seenInvariant[invariant.ID] = true
		if strings.TrimSpace(invariant.Description) == "" {
			violations = append(violations, fmt.Sprintf("invariant %q has empty description", invariant.ID))
		}
		violations = append(violations, verifyAnchor(root, "invariant "+invariant.ID, invariant.SourcePath, invariant.SourceAnchor)...)
		if len(invariant.ThreatIDs) == 0 {
			violations = append(violations, fmt.Sprintf("invariant %q has no threat IDs", invariant.ID))
		}
		seenThreat := map[string]bool{}
		for _, id := range invariant.ThreatIDs {
			if !threatIDPattern.MatchString(id) || !threats[id] {
				violations = append(violations, fmt.Sprintf("invariant %q references unknown threat ID %q", invariant.ID, id))
			}
			if seenThreat[id] {
				violations = append(violations, fmt.Sprintf("invariant %q duplicates threat ID %q", invariant.ID, id))
			}
			seenThreat[id] = true
		}
	}
	if len(scope.Invariants) == 0 {
		violations = append(violations, "invariants must not be empty")
	}

	seenEvidence := map[string]bool{}
	for _, evidence := range scope.Evidence {
		if !scopeIDPattern.MatchString(evidence.ID) || !strings.HasPrefix(evidence.ID, "SR-EVD-") {
			violations = append(violations, fmt.Sprintf("invalid evidence ID %q", evidence.ID))
		}
		if seenEvidence[evidence.ID] {
			violations = append(violations, fmt.Sprintf("duplicate evidence ID %q", evidence.ID))
		}
		seenEvidence[evidence.ID] = true
		if !evidenceKinds[evidence.Kind] {
			violations = append(violations, fmt.Sprintf("evidence %q has unknown kind %q", evidence.ID, evidence.Kind))
		}
		violations = append(violations, verifyAnchor(root, "evidence "+evidence.ID, evidence.Path, evidence.RequiredAnchor)...)
	}
	if len(scope.Evidence) == 0 {
		violations = append(violations, "evidence must not be empty")
	}

	seenFinding := map[string]bool{}
	for _, finding := range findings.Findings {
		violations = append(violations, verifyFinding(root, finding, seenFinding, threats, now)...)
	}
	sort.Strings(violations)
	return violations
}

func verifyIdentity(identity Identity, version, label string) []string {
	var violations []string
	if strings.TrimSpace(identity.Name) == "" {
		violations = append(violations, label+" document name must not be empty")
	}
	if identity.Version != version {
		violations = append(violations, fmt.Sprintf("%s schema version must be %q", label, version))
	}
	if _, err := parseDate(identity.Updated); err != nil {
		violations = append(violations, label+" updated date: "+err.Error())
	}
	return violations
}

func verifyFinding(root string, finding Finding, seen, threats map[string]bool, now time.Time) []string {
	var violations []string
	if !findingIDPattern.MatchString(finding.ID) {
		violations = append(violations, fmt.Sprintf("invalid finding ID %q", finding.ID))
	}
	if seen[finding.ID] {
		violations = append(violations, fmt.Sprintf("duplicate finding ID %q", finding.ID))
	}
	seen[finding.ID] = true
	if !contains([]string{"critical", "high", "medium", "low"}, finding.Severity) {
		violations = append(violations, fmt.Sprintf("finding %q has unknown severity %q", finding.ID, finding.Severity))
	}
	if !contains([]string{"open", "remediated", "verified_closed", "accepted_risk"}, finding.Status) {
		violations = append(violations, fmt.Sprintf("finding %q has unknown status %q", finding.ID, finding.Status))
	}
	if !contains([]string{"internal", "external"}, finding.Source) {
		violations = append(violations, fmt.Sprintf("finding %q has unknown source %q", finding.ID, finding.Source))
	}
	if strings.TrimSpace(finding.Title) == "" || strings.TrimSpace(finding.Owner) == "" {
		violations = append(violations, fmt.Sprintf("finding %q requires title and owner", finding.ID))
	}
	if len(finding.ThreatIDs) == 0 {
		violations = append(violations, fmt.Sprintf("finding %q has no threat IDs", finding.ID))
	}
	seenThreat := map[string]bool{}
	for _, id := range finding.ThreatIDs {
		if !threatIDPattern.MatchString(id) || !threats[id] {
			violations = append(violations, fmt.Sprintf("finding %q references unknown threat ID %q", finding.ID, id))
		}
		if seenThreat[id] {
			violations = append(violations, fmt.Sprintf("finding %q duplicates threat ID %q", finding.ID, id))
		}
		seenThreat[id] = true
	}
	if len(finding.AffectedPaths) == 0 {
		violations = append(violations, fmt.Sprintf("finding %q has no affected paths", finding.ID))
	}
	violations = append(violations, verifyPaths(root, "finding "+finding.ID+" affected_paths", finding.AffectedPaths, false)...)
	if finding.Status == "remediated" || finding.Status == "verified_closed" {
		if len(finding.RemediationEvidence) == 0 {
			violations = append(violations, fmt.Sprintf("finding %q status %s requires remediation evidence", finding.ID, finding.Status))
		}
	}
	violations = append(violations, verifyPaths(root, "finding "+finding.ID+" remediation_evidence", finding.RemediationEvidence, true)...)
	if (finding.Status == "remediated" || finding.Status == "verified_closed") && len(finding.RemediationEvidence) != 0 && !remediationLinksFinding(root, finding) {
		violations = append(violations, fmt.Sprintf("finding %q remediation evidence must contain its finding ID", finding.ID))
	}
	if finding.Status == "verified_closed" {
		if finding.VerificationSource != "external" {
			violations = append(violations, fmt.Sprintf("finding %q verified_closed requires verification_source external", finding.ID))
		}
		if len(finding.VerificationEvidence) == 0 {
			violations = append(violations, fmt.Sprintf("finding %q verified_closed requires external verification evidence", finding.ID))
		}
		violations = append(violations, verifyPaths(root, "finding "+finding.ID+" verification_evidence", finding.VerificationEvidence, true)...)
		if len(finding.VerificationEvidence) != 0 && !evidenceLinksFinding(root, finding.VerificationEvidence, finding.ID) {
			violations = append(violations, fmt.Sprintf("finding %q external verification evidence must contain its finding ID", finding.ID))
		}
	} else if finding.VerificationSource != "" || len(finding.VerificationEvidence) != 0 {
		violations = append(violations, fmt.Sprintf("finding %q carries verification attribution outside verified_closed status", finding.ID))
	}
	if finding.Status == "accepted_risk" {
		accepted, acceptedErr := parseDate(finding.AcceptedOn)
		expires, expiresErr := parseDate(finding.Expires)
		if acceptedErr != nil || expiresErr != nil {
			violations = append(violations, fmt.Sprintf("accepted-risk finding %q requires valid accepted_on and expires dates", finding.ID))
		} else if !expires.After(accepted) || expires.Sub(accepted) > 30*24*time.Hour {
			violations = append(violations, fmt.Sprintf("accepted-risk finding %q expiry must be after acceptance and at most 30 days", finding.ID))
		} else if expires.Before(dateOnly(now)) {
			violations = append(violations, fmt.Sprintf("accepted-risk finding %q expired on %s", finding.ID, finding.Expires))
		}
	} else if finding.AcceptedOn != "" || finding.Expires != "" {
		violations = append(violations, fmt.Sprintf("finding %q carries accepted-risk dates outside accepted_risk status", finding.ID))
	}
	return violations
}

func remediationLinksFinding(root string, finding Finding) bool {
	return evidenceLinksFinding(root, finding.RemediationEvidence, finding.ID)
}

func evidenceLinksFinding(root string, paths []string, findingID string) bool {
	for _, rel := range paths {
		resolved, err := resolvePath(root, rel, true)
		if err != nil {
			continue
		}
		raw, err := readBoundedFile(resolved, MaxReferencedBytes)
		if err == nil && strings.Contains(string(raw), findingID) {
			return true
		}
	}
	return false
}

func verifyAnchor(root, label, rel, anchor string) []string {
	resolved, err := resolvePath(root, rel, true)
	if err != nil {
		return []string{label + " path: " + err.Error()}
	}
	if strings.TrimSpace(anchor) == "" {
		return []string{label + " required anchor must not be empty"}
	}
	raw, err := readBoundedFile(resolved, MaxReferencedBytes)
	if err != nil {
		return []string{label + " reference: " + err.Error()}
	}
	if !strings.Contains(string(raw), anchor) {
		return []string{fmt.Sprintf("%s anchor %q is absent from %q", label, anchor, rel)}
	}
	return nil
}

func verifyPaths(root, label string, paths []string, filesOnly bool) []string {
	var violations []string
	seen := map[string]bool{}
	for _, rel := range paths {
		if seen[rel] {
			violations = append(violations, fmt.Sprintf("%s duplicates path %q", label, rel))
		}
		seen[rel] = true
		if _, err := resolvePath(root, rel, filesOnly); err != nil {
			violations = append(violations, fmt.Sprintf("%s path %q: %v", label, rel, err))
		}
	}
	return violations
}

func canonicalRoot(root string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", fmt.Errorf("repository root must not be empty")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve repository root: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("stat repository root: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("repository root is not a directory")
	}
	return filepath.Clean(abs), nil
}

func resolvePath(root, rel string, fileOnly bool) (string, error) {
	if rel == "" || len(rel) > 256 || strings.Contains(rel, "\\") || strings.ContainsRune(rel, 0) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("unsafe repository-relative path %q", rel)
	}
	clean := filepath.Clean(rel)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || clean != rel {
		return "", fmt.Errorf("unsafe repository-relative path %q", rel)
	}
	current := root
	for _, part := range strings.Split(clean, string(filepath.Separator)) {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("symlink is forbidden at %q", rel)
		}
	}
	info, err := os.Stat(current)
	if err != nil {
		return "", err
	}
	if fileOnly && !info.Mode().IsRegular() {
		return "", fmt.Errorf("%q is not a regular file", rel)
	}
	return current, nil
}

func loadThreatIDs(root, rel string) (map[string]bool, error) {
	if rel != "configs/security/threat-model.json" {
		return nil, fmt.Errorf("threat model path must be configs/security/threat-model.json")
	}
	resolved, err := resolvePath(root, rel, true)
	if err != nil {
		return nil, err
	}
	raw, err := readBoundedFile(resolved, MaxReferencedBytes)
	if err != nil {
		return nil, err
	}
	var register struct {
		Model struct {
			Version string `json:"version"`
		} `json:"model"`
		Threats []struct {
			ID string `json:"id"`
		} `json:"threats"`
	}
	if err := json.Unmarshal(raw, &register); err != nil {
		return nil, err
	}
	if register.Model.Version != "truerepublic.threat-model/v1" {
		return nil, fmt.Errorf("unsupported threat model version %q", register.Model.Version)
	}
	if len(register.Threats) == 0 {
		return nil, fmt.Errorf("threat register is empty")
	}
	ids := make(map[string]bool, len(register.Threats))
	for _, threat := range register.Threats {
		if !threatIDPattern.MatchString(threat.ID) || ids[threat.ID] {
			return nil, fmt.Errorf("invalid or duplicate threat ID %q", threat.ID)
		}
		ids[threat.ID] = true
	}
	return ids, nil
}

func verifySecurityGate(root, rel string) error {
	resolved, err := resolvePath(root, rel, true)
	if err != nil {
		return err
	}
	raw, err := readBoundedFile(resolved, MaxReferencedBytes)
	if err != nil {
		return err
	}
	var gate struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(raw, &gate); err != nil {
		return err
	}
	if gate.Version != "truerepublic.security-gates/v1" {
		return fmt.Errorf("unsupported security gate version %q", gate.Version)
	}
	return nil
}

func parseDate(value string) (time.Time, error) {
	if !datePattern.MatchString(value) {
		return time.Time{}, fmt.Errorf("must use YYYY-MM-DD")
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil || parsed.Format("2006-01-02") != value {
		return time.Time{}, fmt.Errorf("must be a valid calendar date")
	}
	return parsed, nil
}

func dateOnly(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
