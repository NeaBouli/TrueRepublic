package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

type securityGateContract struct {
	Version                   string                     `json:"version"`
	ReviewCadenceDays         int                        `json:"review_cadence_days"`
	ExceptionMaxDays          int                        `json:"exception_max_days"`
	Toolchains                map[string]string          `json:"toolchains"`
	Tools                     map[string]string          `json:"tools"`
	Actions                   map[string]string          `json:"actions"`
	GoVulnerabilityExceptions []goVulnerabilityException `json:"go_vulnerability_exceptions"`
}

type goVulnerabilityException struct {
	ID         string `json:"id"`
	ApprovedOn string `json:"approved_on"`
	Expires    string `json:"expires"`
	Reason     string `json:"reason"`
}

var actionUsePattern = regexp.MustCompile(`(?m)^\s*(?:-\s*)?uses:\s*([^@\s]+)@([0-9a-f]{40})\s+#\s+\S+`)
var exactVersionPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

func TestSecurityGateRepositoryContract(t *testing.T) {
	contractBytes, err := os.ReadFile("configs/security/gates.json")
	if err != nil {
		t.Fatal(err)
	}
	var contract securityGateContract
	if err := json.Unmarshal(contractBytes, &contract); err != nil {
		t.Fatalf("parse security gate contract: %v", err)
	}

	files := loadSecurityGateFiles(t)
	if violations := securityGateViolations(files, contract); len(violations) != 0 {
		t.Fatalf("security gate contract violations:\n- %s", strings.Join(violations, "\n- "))
	}

	for _, path := range []string{
		"scripts/check-go-vulnerabilities.sh",
		"scripts/test-go-vulnerability-scan.sh",
		"scripts/check-static-analysis.sh",
		"scripts/check-secret-scan.sh",
		"scripts/test-secret-scan.sh",
	} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode()&0o111 == 0 {
			t.Fatalf("%s must be executable", path)
		}
	}

	t.Run("rejects mutable action reference", func(t *testing.T) {
		mutated := cloneSecurityGateFiles(files)
		mutated[".github/workflows/docs-check.yml"] = strings.Replace(
			mutated[".github/workflows/docs-check.yml"],
			"actions/checkout@"+contract.Actions["actions/checkout"],
			"actions/checkout@v7.0.1", 1,
		)
		assertSecurityGateRejected(t, mutated, contract, "immutable full SHA")
	})

	t.Run("rejects disabled scanner", func(t *testing.T) {
		mutated := cloneSecurityGateFiles(files)
		mutated[".github/workflows/security-scan.yml"] += "\n# scanner bypass || true\n"
		assertSecurityGateRejected(t, mutated, contract, "failure bypass")

		mutated = cloneSecurityGateFiles(files)
		mutated[".github/workflows/security-scan.yml"] = strings.Replace(
			mutated[".github/workflows/security-scan.yml"],
			"./scripts/check-go-vulnerabilities.sh", "govulncheck ./...", 1,
		)
		assertSecurityGateRejected(t, mutated, contract, "check-go-vulnerabilities.sh")
	})

	t.Run("rejects missing secret failure test", func(t *testing.T) {
		mutated := cloneSecurityGateFiles(files)
		mutated[".github/workflows/security-scan.yml"] = strings.ReplaceAll(
			mutated[".github/workflows/security-scan.yml"],
			"./scripts/test-secret-scan.sh", "./scripts/missing-secret-test.sh",
		)
		assertSecurityGateRejected(t, mutated, contract, "test-secret-scan.sh")
	})

	t.Run("rejects broad secret allowlist", func(t *testing.T) {
		mutated := cloneSecurityGateFiles(files)
		mutated[".gitleaks.toml"] += "\npaths = ['''docs/.*''']\n"
		assertSecurityGateRejected(t, mutated, contract, "broad path")
	})

	t.Run("rejects missing lockfile", func(t *testing.T) {
		mutated := cloneSecurityGateFiles(files)
		delete(mutated, "contracts/Cargo.lock")
		assertSecurityGateRejected(t, mutated, contract, "contracts/Cargo.lock")
	})
}

func loadSecurityGateFiles(t *testing.T) map[string]string {
	t.Helper()
	paths := []string{
		".github/dependabot.yml",
		".github/workflows/security-scan.yml",
		".gitleaks.toml",
		"client-web/package-lock.json",
		"contracts/Cargo.lock",
		"go.mod",
		"go.sum",
		"scripts/check-go-vulnerabilities.sh",
		"scripts/test-go-vulnerability-scan.sh",
		"scripts/check-secret-scan.sh",
		"scripts/check-static-analysis.sh",
		"scripts/test-secret-scan.sh",
	}
	workflows, err := filepath.Glob(".github/workflows/*.yml")
	if err != nil {
		t.Fatal(err)
	}
	paths = append(paths, workflows...)

	files := make(map[string]string, len(paths))
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		files[path] = string(content)
	}
	return files
}

func securityGateViolations(files map[string]string, contract securityGateContract) []string {
	var violations []string
	if contract.Version != "truerepublic.security-gates/v1" || contract.ReviewCadenceDays != 7 || contract.ExceptionMaxDays < 1 || contract.ExceptionMaxDays > 30 {
		violations = append(violations, "invalid version, review cadence, or exception bound")
	}
	for name, pattern := range map[string]string{
		"govulncheck": `^v\d+\.\d+\.\d+$`,
		"staticcheck": `^v\d+\.\d+\.\d+$`,
		"gitleaks":    `^v\d+\.\d+\.\d+$`,
		"cargo_audit": `^\d+\.\d+\.\d+$`,
	} {
		if matched, _ := regexp.MatchString(pattern, contract.Tools[name]); !matched {
			violations = append(violations, fmt.Sprintf("tool %s is not exactly versioned", name))
		}
	}
	for _, name := range []string{"go", "node", "rust"} {
		if !exactVersionPattern.MatchString(contract.Toolchains[name]) {
			violations = append(violations, fmt.Sprintf("toolchain %s is not exactly versioned", name))
		}
	}
	seenExceptions := make(map[string]struct{}, len(contract.GoVulnerabilityExceptions))
	for _, exception := range contract.GoVulnerabilityExceptions {
		approved, approvedErr := time.Parse(time.DateOnly, exception.ApprovedOn)
		expires, expiresErr := time.Parse(time.DateOnly, exception.Expires)
		if !regexp.MustCompile(`^GO-[0-9]{4}-[0-9]{4}$`).MatchString(exception.ID) ||
			approvedErr != nil || expiresErr != nil || expires.Before(approved) ||
			expires.Sub(approved) > time.Duration(contract.ExceptionMaxDays)*24*time.Hour ||
			len(strings.TrimSpace(exception.Reason)) < 20 {
			violations = append(violations, "invalid Go vulnerability exception "+exception.ID)
		}
		if _, exists := seenExceptions[exception.ID]; exists {
			violations = append(violations, "duplicate Go vulnerability exception "+exception.ID)
		}
		seenExceptions[exception.ID] = struct{}{}
	}

	workflowNames := make([]string, 0)
	for path := range files {
		if strings.HasPrefix(path, ".github/workflows/") && strings.HasSuffix(path, ".yml") {
			workflowNames = append(workflowNames, path)
		}
	}
	sort.Strings(workflowNames)
	for _, path := range workflowNames {
		workflow := files[path]
		if !strings.Contains(workflow, "permissions:\n  contents: read") {
			violations = append(violations, path+" must use contents-read permissions")
		}
		if strings.Count(workflow, "runs-on:") != strings.Count(workflow, "timeout-minutes:") {
			violations = append(violations, path+" must bound every job with timeout-minutes")
		}
		usesCount := strings.Count(workflow, "uses:")
		matches := actionUsePattern.FindAllStringSubmatch(workflow, -1)
		if len(matches) != usesCount {
			violations = append(violations, path+" action references must use an immutable full SHA and version comment")
		}
		checkoutCount := 0
		for _, match := range matches {
			action, sha := match[1], match[2]
			if contract.Actions[action] != sha {
				violations = append(violations, fmt.Sprintf("%s uses unapproved action pin %s@%s", path, action, sha))
			}
			if action == "actions/checkout" {
				checkoutCount++
			}
		}
		if strings.Count(workflow, "persist-credentials: false") != checkoutCount {
			violations = append(violations, path+" must disable persisted credentials for every checkout")
		}
	}

	security := files[".github/workflows/security-scan.yml"]
	for _, required := range []string{
		"go-vuln:", "go-static:", "secret-scan:", "rust-audit:", "node-audit-client:",
		"go mod verify", "./scripts/check-go-vulnerabilities.sh", "./scripts/test-go-vulnerability-scan.sh", "./scripts/check-static-analysis.sh",
		"./scripts/check-secret-scan.sh", "./scripts/test-secret-scan.sh",
		"cargo install cargo-audit --version", "cargo audit", "npm ci", "npm run audit:high",
		"go test . -run '^TestSecurityGateRepositoryContract$'",
	} {
		if !strings.Contains(security, required) {
			violations = append(violations, "security workflow missing "+required)
		}
	}
	if strings.Contains(security, "@latest") || strings.Contains(security, "|| true") {
		violations = append(violations, "security workflow contains a mutable tool or failure bypass")
	}
	goVulnScript := files["scripts/check-go-vulnerabilities.sh"]
	for _, required := range []string{"-json", "go_vulnerability_exceptions", "exception_max_days", "fromdateiso8601", "$found == $allowed", "$found - $fixable", "SECURITY_GATE_TODAY"} {
		if !strings.Contains(goVulnScript, required) {
			violations = append(violations, "Go vulnerability gate missing "+required)
		}
	}
	secretScript := files["scripts/check-secret-scan.sh"]
	if !strings.Contains(secretScript, "git ls-files -z --cached --others --exclude-standard >\"$manifest\"") ||
		!strings.Contains(secretScript, "file_count -eq 0") {
		violations = append(violations, "secret scan must select the maintained Git tree")
	}
	if !strings.Contains(files["scripts/test-secret-scan.sh"], "ignored a Git enumeration failure") {
		violations = append(violations, "secret scan must test Git enumeration failure semantics")
	}

	dependabot := files[".github/dependabot.yml"]
	for _, required := range []string{"github-actions", "gomod", "cargo", "npm", "interval: weekly", "open-pull-requests-limit: 3"} {
		if !strings.Contains(dependabot, required) {
			violations = append(violations, "Dependabot policy missing "+required)
		}
	}
	gitleaks := files[".gitleaks.toml"]
	if strings.Contains(gitleaks, "paths =") || strings.Contains(gitleaks, "commits =") || strings.Contains(gitleaks, "stopwords =") {
		violations = append(violations, "gitleaks policy must not use broad path, commit, or stopword allowlists")
	}
	for _, exactFixture := range []string{"mnemonic/private/signing", "firewall/TLS/DNS", `cosmos1climatekey\.\.\.`, "private-key-material-123456"} {
		if !strings.Contains(gitleaks, exactFixture) {
			violations = append(violations, "gitleaks policy missing exact synthetic fixture "+exactFixture)
		}
	}
	for _, lockfile := range []string{"go.sum", "contracts/Cargo.lock", "client-web/package-lock.json"} {
		if _, exists := files[lockfile]; !exists {
			violations = append(violations, "missing required lockfile "+lockfile)
		}
	}
	return violations
}

func cloneSecurityGateFiles(files map[string]string) map[string]string {
	clone := make(map[string]string, len(files))
	for path, content := range files {
		clone[path] = content
	}
	return clone
}

func assertSecurityGateRejected(t *testing.T, files map[string]string, contract securityGateContract, want string) {
	t.Helper()
	violations := securityGateViolations(files, contract)
	if !strings.Contains(strings.Join(violations, "\n"), want) {
		t.Fatalf("missing rejection %q in %v", want, violations)
	}
}
