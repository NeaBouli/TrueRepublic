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

	"gopkg.in/yaml.v3"
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

type dependabotConfig struct {
	Version int                `yaml:"version"`
	Updates []dependabotUpdate `yaml:"updates"`
}

type dependabotUpdate struct {
	Ecosystem string                      `yaml:"package-ecosystem"`
	Groups    map[string]dependabotGroup  `yaml:"groups"`
	Ignore    []dependabotIgnoreCondition `yaml:"ignore"`
}

type dependabotGroup struct {
	AppliesTo   string   `yaml:"applies-to"`
	Patterns    []string `yaml:"patterns"`
	Exclude     []string `yaml:"exclude-patterns"`
	UpdateTypes []string `yaml:"update-types"`
}

type dependabotIgnoreCondition struct {
	DependencyName string   `yaml:"dependency-name"`
	UpdateTypes    []string `yaml:"update-types"`
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

	t.Run("rejects automatic major dependency updates", func(t *testing.T) {
		mutated := cloneSecurityGateFiles(files)
		mutated[".github/dependabot.yml"] = mutateDependabotYAML(t, mutated[".github/dependabot.yml"], func(updates []map[string]interface{}) {
			for _, update := range updates {
				if ecosystem, ok := update["package-ecosystem"].(string); !ok || ecosystem != "gomod" {
					continue
				}
				ignore, ok := update["ignore"].([]interface{})
				if !ok {
					continue
				}
				for _, ignoreEntry := range ignore {
					condition, ok := ignoreEntry.(map[string]interface{})
					if !ok {
						continue
					}
					if condition["dependency-name"] == "*" {
						condition["update-types"] = []interface{}{"version-update:semver-major"}
					}
				}
			}
		})
		assertSecurityGateRejected(t, mutated, contract, "Dependabot policy must use exact exception set for gomod")
	})

	t.Run("rejects widening Go maintenance beyond patch-only", func(t *testing.T) {
		mutated := cloneSecurityGateFiles(files)
		mutated[".github/dependabot.yml"] = mutateDependabotYAML(t, mutated[".github/dependabot.yml"], func(updates []map[string]interface{}) {
			for _, update := range updates {
				if ecosystem, ok := update["package-ecosystem"].(string); !ok || ecosystem != "gomod" {
					continue
				}
				groups, ok := update["groups"].(map[string]interface{})
				if !ok {
					continue
				}
				group, ok := groups["go-maintenance"].(map[string]interface{})
				if !ok {
					continue
				}
				group["update-types"] = []interface{}{"minor", "patch"}
			}
		})
		assertSecurityGateRejected(t, mutated, contract, "Dependabot policy must use exact grouped update-types for gomod")
	})

	t.Run("rejects missing npm package exceptions", func(t *testing.T) {
		mutated := cloneSecurityGateFiles(files)
		mutated[".github/dependabot.yml"] = mutateDependabotYAML(t, mutated[".github/dependabot.yml"], func(updates []map[string]interface{}) {
			for _, update := range updates {
				if ecosystem, ok := update["package-ecosystem"].(string); !ok || ecosystem != "npm" {
					continue
				}
				filtered := make([]interface{}, 0, 1)
				ignore, ok := update["ignore"].([]interface{})
				if !ok {
					continue
				}
				for _, entry := range ignore {
					condition, ok := entry.(map[string]interface{})
					if !ok || condition["dependency-name"] == "@playwright/test" {
						continue
					}
					filtered = append(filtered, condition)
				}
				update["ignore"] = filtered
			}
		})
		assertSecurityGateRejected(t, mutated, contract, "Dependabot policy must use exact exception set for npm")
	})

	t.Run("rejects incomplete Playwright freeze", func(t *testing.T) {
		mutated := cloneSecurityGateFiles(files)
		mutated[".github/dependabot.yml"] = mutateDependabotYAML(t, mutated[".github/dependabot.yml"], func(updates []map[string]interface{}) {
			for _, update := range updates {
				if ecosystem, ok := update["package-ecosystem"].(string); !ok || ecosystem != "npm" {
					continue
				}
				ignore, ok := update["ignore"].([]interface{})
				if !ok {
					continue
				}
				for _, entry := range ignore {
					condition, ok := entry.(map[string]interface{})
					if ok && condition["dependency-name"] == "@playwright/test" {
						condition["update-types"] = []interface{}{"version-update:semver-minor", "version-update:semver-patch"}
					}
				}
			}
		})
		assertSecurityGateRejected(t, mutated, contract, "Dependabot policy must use exact exception set for npm")
	})

	t.Run("rejects missing npm React Refresh exception", func(t *testing.T) {
		mutated := cloneSecurityGateFiles(files)
		mutated[".github/dependabot.yml"] = mutateDependabotYAML(t, mutated[".github/dependabot.yml"], func(updates []map[string]interface{}) {
			for _, update := range updates {
				if ecosystem, ok := update["package-ecosystem"].(string); !ok || ecosystem != "npm" {
					continue
				}
				filtered := make([]interface{}, 0, 2)
				ignore, ok := update["ignore"].([]interface{})
				if !ok {
					continue
				}
				for _, entry := range ignore {
					condition, ok := entry.(map[string]interface{})
					if !ok || condition["dependency-name"] == "eslint-plugin-react-refresh" {
						continue
					}
					filtered = append(filtered, condition)
				}
				update["ignore"] = filtered
			}
		})
		assertSecurityGateRejected(t, mutated, contract, "Dependabot policy must use exact exception set for npm")
	})

	t.Run("rejects missing Cargo cosmwasm grouped exclusion", func(t *testing.T) {
		mutated := cloneSecurityGateFiles(files)
		mutated[".github/dependabot.yml"] = mutateDependabotYAML(t, mutated[".github/dependabot.yml"], func(updates []map[string]interface{}) {
			for _, update := range updates {
				if ecosystem, ok := update["package-ecosystem"].(string); !ok || ecosystem != "cargo" {
					continue
				}
				groups, ok := update["groups"].(map[string]interface{})
				if !ok {
					continue
				}
				group, ok := groups["rust-maintenance"].(map[string]interface{})
				if !ok {
					continue
				}
				delete(group, "exclude-patterns")
			}
		})
		assertSecurityGateRejected(t, mutated, contract, "Dependabot policy must exclude cosmwasm-* from rust maintenance grouping for cargo")
	})

	t.Run("rejects missing Cargo cosmwasm minor exception", func(t *testing.T) {
		mutated := cloneSecurityGateFiles(files)
		mutated[".github/dependabot.yml"] = mutateDependabotYAML(t, mutated[".github/dependabot.yml"], func(updates []map[string]interface{}) {
			for _, update := range updates {
				if ecosystem, ok := update["package-ecosystem"].(string); !ok || ecosystem != "cargo" {
					continue
				}
				filtered := make([]interface{}, 0, 1)
				ignore, ok := update["ignore"].([]interface{})
				if !ok {
					continue
				}
				for _, entry := range ignore {
					condition, ok := entry.(map[string]interface{})
					if !ok || condition["dependency-name"] == "cosmwasm-*" {
						continue
					}
					filtered = append(filtered, condition)
				}
				update["ignore"] = filtered
			}
		})
		assertSecurityGateRejected(t, mutated, contract, "Dependabot policy must use exact exception set for cargo")
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
	violations = append(violations, dependabotPolicyViolations(dependabot)...)
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

func dependabotPolicyViolations(content string) []string {
	var config dependabotConfig
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return []string{"Dependabot policy is invalid YAML"}
	}
	if config.Version != 2 {
		return []string{"Dependabot policy must use version 2"}
	}

	expectedGroups := map[string]string{
		"github-actions": "github-actions",
		"gomod":          "go-maintenance",
		"cargo":          "rust-maintenance",
		"npm":            "client-maintenance",
	}
	expectedUpdateTypes := map[string][]string{
		"github-actions": {"minor", "patch"},
		"gomod":          {"patch"},
		"cargo":          {"minor", "patch"},
		"npm":            {"minor", "patch"},
	}
	expectedIgnores := map[string][]dependabotIgnoreCondition{
		"github-actions": {
			{DependencyName: "*", UpdateTypes: []string{"version-update:semver-major"}},
		},
		"gomod": {
			{DependencyName: "*", UpdateTypes: []string{"version-update:semver-major", "version-update:semver-minor"}},
		},
		"cargo": {
			{DependencyName: "*", UpdateTypes: []string{"version-update:semver-major"}},
			{DependencyName: "cosmwasm-*", UpdateTypes: []string{"version-update:semver-major", "version-update:semver-minor"}},
		},
		"npm": {
			{DependencyName: "*", UpdateTypes: []string{"version-update:semver-major"}},
			{DependencyName: "@playwright/test", UpdateTypes: []string{"version-update:semver-major", "version-update:semver-minor", "version-update:semver-patch"}},
			{DependencyName: "eslint-plugin-react-refresh", UpdateTypes: []string{"version-update:semver-major", "version-update:semver-minor"}},
		},
	}
	seen := make(map[string]bool, len(expectedGroups))
	var violations []string
	for _, update := range config.Updates {
		groupName, expected := expectedGroups[update.Ecosystem]
		if !expected || seen[update.Ecosystem] {
			violations = append(violations, "Dependabot policy contains an unexpected or duplicate ecosystem "+update.Ecosystem)
			continue
		}
		seen[update.Ecosystem] = true
		group, exists := update.Groups[groupName]
		if !exists || group.AppliesTo != "version-updates" || strings.Join(group.Patterns, ",") != "*" {
			violations = append(violations, "Dependabot policy must configure grouped version updates for "+update.Ecosystem)
			continue
		}
		if !slicesEqualInsensitive(group.UpdateTypes, expectedUpdateTypes[update.Ecosystem]) {
			violations = append(violations, "Dependabot policy must use exact grouped update-types for "+update.Ecosystem)
		}
		if update.Ecosystem == "cargo" {
			if !slicesEqualInsensitive(group.Exclude, []string{"cosmwasm-*"}) {
				violations = append(violations, "Dependabot policy must exclude cosmwasm-* from rust maintenance grouping for cargo")
			}
		} else if len(group.Exclude) != 0 {
			violations = append(violations, "Dependabot policy must not set exclude-patterns for "+update.Ecosystem)
		}
		if !dependabotIgnoreConditionsEqual(update.Ignore, expectedIgnores[update.Ecosystem]) {
			violations = append(violations, "Dependabot policy must use exact exception set for "+update.Ecosystem)
		}
	}
	for ecosystem := range expectedGroups {
		if !seen[ecosystem] {
			violations = append(violations, "Dependabot policy missing compatible update rules for "+ecosystem)
		}
	}
	return violations
}

func mutateDependabotYAML(t *testing.T, content string, mutate func([]map[string]interface{})) string {
	t.Helper()
	var root map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &root); err != nil {
		t.Fatalf("parse dependabot yaml: %v", err)
	}
	rawUpdates, ok := root["updates"].([]interface{})
	if !ok {
		t.Fatalf("dependabot yaml missing updates list")
	}
	updates := make([]map[string]interface{}, 0, len(rawUpdates))
	for _, raw := range rawUpdates {
		update, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		updates = append(updates, update)
	}
	mutate(updates)
	bytes, err := yaml.Marshal(root)
	if err != nil {
		t.Fatalf("render dependabot yaml: %v", err)
	}
	return string(bytes)
}

func slicesEqualInsensitive(lhs, rhs []string) bool {
	if len(lhs) != len(rhs) {
		return false
	}
	left := make(map[string]int, len(lhs))
	for _, value := range lhs {
		left[value]++
	}
	right := make(map[string]int, len(rhs))
	for _, value := range rhs {
		right[value]++
	}
	for key, lhsCount := range left {
		if right[key] != lhsCount {
			return false
		}
	}
	return true
}

func dependabotIgnoreConditionsEqual(lhs, rhs []dependabotIgnoreCondition) bool {
	if len(lhs) != len(rhs) {
		return false
	}
	lhsIndex := make(map[string]dependabotIgnoreCondition, len(lhs))
	for _, item := range lhs {
		lhsIndex[item.DependencyName] = item
	}
	for _, item := range rhs {
		lhsItem, ok := lhsIndex[item.DependencyName]
		if !ok || !slicesEqualInsensitive(lhsItem.UpdateTypes, item.UpdateTypes) {
			return false
		}
	}
	for _, item := range lhs {
		if item.DependencyName == "" {
			continue
		}
		found := false
		for _, expected := range rhs {
			if expected.DependencyName == item.DependencyName {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
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
