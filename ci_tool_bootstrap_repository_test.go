package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"truerepublic/toolbootstrapevidence"
)

type ciToolSpec struct {
	id       string
	kind     string
	module   string
	pkg      string
	gatesKey string
	artifact string
}

var expectedCITools = []ciToolSpec{
	{"cyclonedx-gomod", "go", "github.com/CycloneDX/cyclonedx-gomod", "github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod", "cyclonedx_gomod", "cyclonedx-gomod"},
	{"govulncheck", "go", "golang.org/x/vuln", "golang.org/x/vuln/cmd/govulncheck", "govulncheck", "govulncheck"},
	{"staticcheck", "go", "honnef.co/go/tools", "honnef.co/go/tools/cmd/staticcheck", "staticcheck", "staticcheck"},
	{"gitleaks", "go", "github.com/zricethezav/gitleaks/v8", "github.com/zricethezav/gitleaks/v8", "gitleaks", "gitleaks"},
	{"cyclonedx-npm", "npm", "", "@cyclonedx/cyclonedx-npm", "cyclonedx_npm", "node_modules/@cyclonedx/cyclonedx-npm/package.json"},
}

var coveredBootstrapWorkflows = []string{
	".github/workflows/reproducible-daemon.yml",
	".github/workflows/security-scan.yml",
	".github/workflows/docs-check.yml",
}

var (
	goInstallPattern   = regexp.MustCompile(`\bgo install\b`)
	npmInstallPattern  = regexp.MustCompile(`\bnpm install\b`)
	npxPattern         = regexp.MustCompile(`\bnpx\b`)
	ciToolVersionRegex = regexp.MustCompile(`^v?[0-9]+\.[0-9]+\.[0-9]+$`)
)

type ciToolBootstrapState struct {
	contract toolbootstrapevidence.Contract
	gates    securityGateContract
	files    map[string]string
	packages []string
}

func TestCIToolBootstrapRepositoryContract(t *testing.T) {
	state := loadCIToolBootstrapState(t)
	if violations := ciToolBootstrapViolations(state); len(violations) != 0 {
		t.Fatalf("CI tool bootstrap contract violations:\n- %s", strings.Join(violations, "\n- "))
	}

	t.Run("rejects live go install in a covered workflow", func(t *testing.T) {
		mutated := cloneCIToolBootstrapState(state)
		mutated.files[".github/workflows/security-scan.yml"] += "\n# go install golang.org/x/vuln/cmd/govulncheck@v1.6.0\n"
		assertCIToolBootstrapRejected(t, mutated, "live go install")
	})

	t.Run("rejects live npm install in a covered workflow", func(t *testing.T) {
		mutated := cloneCIToolBootstrapState(state)
		mutated.files[".github/workflows/reproducible-daemon.yml"] += "\n# npm install --global @cyclonedx/cyclonedx-npm@6.0.1\n"
		assertCIToolBootstrapRejected(t, mutated, "live npm install")
	})

	t.Run("rejects npx in a covered workflow", func(t *testing.T) {
		mutated := cloneCIToolBootstrapState(state)
		mutated.files[".github/workflows/docs-check.yml"] += "\n# npx cyclonedx-npm\n"
		assertCIToolBootstrapRejected(t, mutated, "live npx")
	})

	t.Run("rejects security gate version drift", func(t *testing.T) {
		mutated := cloneCIToolBootstrapState(state)
		mutated.gates.Tools["govulncheck"] = "v9.9.9"
		mutated.files["configs/security/gates.json"] = strings.Replace(
			mutated.files["configs/security/gates.json"], `"govulncheck": "v1.6.0"`, `"govulncheck": "v9.9.9"`, 1)
		assertCIToolBootstrapRejected(t, mutated, "govulncheck")
	})

	t.Run("rejects Go module lock drift", func(t *testing.T) {
		mutated := cloneCIToolBootstrapState(state)
		mutated.files["tools/ci/go.mod"] = strings.Replace(
			mutated.files["tools/ci/go.mod"], "golang.org/x/vuln v1.6.0", "golang.org/x/vuln v9.9.9", 1)
		assertCIToolBootstrapRejected(t, mutated, "govulncheck")
	})

	t.Run("rejects npm lock drift", func(t *testing.T) {
		mutated := cloneCIToolBootstrapState(state)
		mutated.files["tools/ci/npm/package-lock.json"] = strings.Replace(
			mutated.files["tools/ci/npm/package-lock.json"], `"version": "6.0.1"`, `"version": "9.9.9"`, 1)
		assertCIToolBootstrapRejected(t, mutated, "cyclonedx-npm")
	})

	t.Run("rejects nested module in root package selection", func(t *testing.T) {
		mutated := cloneCIToolBootstrapState(state)
		mutated.packages = append(append([]string{}, mutated.packages...), "./tools/ci")
		assertCIToolBootstrapRejected(t, mutated, "root package selection must ignore the nested tools/ci module")
	})

	t.Run("rejects missing Dependabot npm coverage", func(t *testing.T) {
		mutated := cloneCIToolBootstrapState(state)
		mutated.files[".github/dependabot.yml"] = strings.Replace(
			mutated.files[".github/dependabot.yml"], "directory: /tools/ci/npm", "directory: /tools/ci/removed", 1)
		assertCIToolBootstrapRejected(t, mutated, "Dependabot policy must cover /tools/ci/npm")
	})

	t.Run("rejects a duplicate contract tool", func(t *testing.T) {
		mutated := cloneCIToolBootstrapState(state)
		mutated.contract.Bootstrap.Tools = append(mutated.contract.Bootstrap.Tools, mutated.contract.Bootstrap.Tools[0])
		assertCIToolBootstrapRejected(t, mutated, "exactly the configured tool allowlist")
	})

	t.Run("rejects a missing workflow evidence step", func(t *testing.T) {
		mutated := cloneCIToolBootstrapState(state)
		mutated.files[".github/workflows/reproducible-daemon.yml"] = strings.Replace(
			mutated.files[".github/workflows/reproducible-daemon.yml"],
			"./scripts/generate-tool-bootstrap-evidence.sh", "./scripts/missing-generator.sh", 1)
		assertCIToolBootstrapRejected(t, mutated, "generate-tool-bootstrap-evidence.sh")
	})
}

func loadCIToolBootstrapState(t *testing.T) ciToolBootstrapState {
	t.Helper()
	paths := []string{
		"configs/release/tool-platform.json",
		"configs/security/gates.json",
		"tools/ci/go.mod",
		"tools/ci/go.sum",
		"tools/ci/tools.go",
		"tools/ci/npm/package.json",
		"tools/ci/npm/package-lock.json",
		"Makefile",
		".gitignore",
		".github/dependabot.yml",
		"scripts/build-ci-tool.sh",
		"scripts/generate-tool-bootstrap-evidence.sh",
		"scripts/verify-tool-bootstrap-evidence.sh",
		"scripts/test-tool-bootstrap-evidence.sh",
	}
	paths = append(paths, coveredBootstrapWorkflows...)
	files := make(map[string]string, len(paths))
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		files[path] = string(content)
	}

	var contract toolbootstrapevidence.Contract
	if err := json.Unmarshal([]byte(files["configs/release/tool-platform.json"]), &contract); err != nil {
		t.Fatalf("parse tool-platform contract: %v", err)
	}
	var gates securityGateContract
	if err := json.Unmarshal([]byte(files["configs/security/gates.json"]), &gates); err != nil {
		t.Fatalf("parse security gate contract: %v", err)
	}

	listing, err := exec.Command("scripts/go-packages.sh", "--list").Output()
	if err != nil {
		t.Fatalf("list repository Go packages: %v", err)
	}
	var packages []string
	for _, line := range strings.Split(string(listing), "\n") {
		if line != "" {
			packages = append(packages, line)
		}
	}
	return ciToolBootstrapState{contract: contract, gates: gates, files: files, packages: packages}
}

func ciToolBootstrapViolations(state ciToolBootstrapState) []string {
	var violations []string
	contract := state.contract
	if contract.Schema != toolbootstrapevidence.ContractSchema {
		violations = append(violations, "tool-platform contract schema mismatch")
	}
	expectedLocks := map[string]string{
		"go_module":   "tools/ci/go.mod",
		"go_sum":      "tools/ci/go.sum",
		"npm_package": "tools/ci/npm/package.json",
		"npm_lock":    "tools/ci/npm/package-lock.json",
	}
	actualLocks := map[string]string{
		"go_module":   contract.Bootstrap.GoModule,
		"go_sum":      contract.Bootstrap.GoSum,
		"npm_package": contract.Bootstrap.NpmPackage,
		"npm_lock":    contract.Bootstrap.NpmLock,
	}
	for name, expected := range expectedLocks {
		if actualLocks[name] != expected {
			violations = append(violations, "tool-platform lock path "+name+" must be "+expected)
		}
	}

	if len(contract.Bootstrap.Tools) != len(expectedCITools) {
		violations = append(violations, "tool-platform contract must declare exactly the configured tool allowlist")
	}
	byID := make(map[string]toolbootstrapevidence.BootstrapTool, len(contract.Bootstrap.Tools))
	for _, tool := range contract.Bootstrap.Tools {
		if _, exists := byID[tool.ID]; exists {
			violations = append(violations, "tool-platform contract must declare exactly the configured tool allowlist")
		}
		byID[tool.ID] = tool
	}
	for _, expected := range expectedCITools {
		tool, exists := byID[expected.id]
		if !exists {
			violations = append(violations, "tool-platform contract is missing tool "+expected.id)
			continue
		}
		if tool.Kind != expected.kind || tool.Module != expected.module || tool.Package != expected.pkg ||
			tool.GatesKey != expected.gatesKey || tool.Artifact != expected.artifact {
			violations = append(violations, "tool-platform tool "+expected.id+" binding drift")
		}
		version := state.gates.Tools[expected.gatesKey]
		if !ciToolVersionRegex.MatchString(version) {
			violations = append(violations, "security gate version for "+expected.gatesKey+" is not exact")
		}
		if expected.kind == "go" {
			if !strings.Contains(state.files["tools/ci/go.mod"], expected.module+" "+version) {
				violations = append(violations, "tools/ci/go.mod does not lock "+expected.gatesKey+" module "+expected.module+" "+version)
			}
			if !strings.Contains(state.files["tools/ci/go.sum"], expected.module+" "+version+" ") {
				violations = append(violations, "tools/ci/go.sum does not lock "+expected.gatesKey+" module "+expected.module+" "+version)
			}
			if !strings.Contains(state.files["tools/ci/tools.go"], `"`+expected.pkg+`"`) {
				violations = append(violations, "tools/ci/tools.go does not pin "+expected.gatesKey+" package "+expected.pkg)
			}
		}
	}
	npmVersion := state.gates.Tools["cyclonedx_npm"]
	var npmPackage struct {
		Private      bool              `json:"private"`
		Dependencies map[string]string `json:"dependencies"`
	}
	if err := json.Unmarshal([]byte(state.files["tools/ci/npm/package.json"]), &npmPackage); err != nil {
		violations = append(violations, "tools/ci/npm/package.json is invalid")
	} else {
		if !npmPackage.Private {
			violations = append(violations, "tools/ci/npm/package.json must stay private")
		}
		if len(npmPackage.Dependencies) != 1 || npmPackage.Dependencies["@cyclonedx/cyclonedx-npm"] != npmVersion {
			violations = append(violations, "tools/ci/npm/package.json must pin @cyclonedx/cyclonedx-npm exactly to the security gate version")
		}
	}
	var npmLock struct {
		LockfileVersion int `json:"lockfileVersion"`
		Packages        map[string]struct {
			Version string `json:"version"`
		} `json:"packages"`
	}
	if err := json.Unmarshal([]byte(state.files["tools/ci/npm/package-lock.json"]), &npmLock); err != nil {
		violations = append(violations, "tools/ci/npm/package-lock.json is invalid")
	} else {
		if npmLock.LockfileVersion != 3 {
			violations = append(violations, "tools/ci/npm/package-lock.json must use lockfileVersion 3")
		}
		locked, exists := npmLock.Packages["node_modules/@cyclonedx/cyclonedx-npm"]
		if !exists || locked.Version != npmVersion {
			violations = append(violations, "tools/ci/npm/package-lock.json does not lock @cyclonedx/cyclonedx-npm "+npmVersion)
		}
	}

	for _, pkg := range state.packages {
		if pkg == "./tools/ci" || strings.HasPrefix(pkg, "./tools/ci/") {
			violations = append(violations, "root package selection must ignore the nested tools/ci module")
		}
	}

	for _, workflowPath := range coveredBootstrapWorkflows {
		workflow := state.files[workflowPath]
		if goInstallPattern.MatchString(workflow) {
			violations = append(violations, workflowPath+" must not contain a live go install")
		}
		if npmInstallPattern.MatchString(workflow) {
			violations = append(violations, workflowPath+" must not contain a live npm install")
		}
		if npxPattern.MatchString(workflow) {
			violations = append(violations, workflowPath+" must not contain a live npx")
		}
	}
	reproducible := state.files[".github/workflows/reproducible-daemon.yml"]
	for _, required := range []string{
		"./scripts/build-ci-tool.sh --tool cyclonedx-gomod",
		"./scripts/build-ci-tool.sh --tool cyclonedx-npm",
		"./scripts/generate-tool-bootstrap-evidence.sh",
		"./scripts/verify-tool-bootstrap-evidence.sh",
		"make ci-tool-bootstrap-contract-test",
		"tool-bootstrap-evidence.json",
		"retention-days: 14",
	} {
		if !strings.Contains(reproducible, required) {
			violations = append(violations, "reproducible-daemon workflow missing "+required)
		}
	}
	security := state.files[".github/workflows/security-scan.yml"]
	for _, required := range []string{
		"./scripts/build-ci-tool.sh --tool govulncheck",
		"./scripts/build-ci-tool.sh --tool staticcheck",
		"./scripts/build-ci-tool.sh --tool gitleaks",
		"go test . -run '^TestCIToolBootstrapRepositoryContract$'",
	} {
		if !strings.Contains(security, required) {
			violations = append(violations, "security workflow missing "+required)
		}
	}
	if !strings.Contains(state.files[".github/workflows/docs-check.yml"], "make ci-tool-bootstrap-contract-test") {
		violations = append(violations, "docs-check workflow missing make ci-tool-bootstrap-contract-test")
	}
	makefile := state.files["Makefile"]
	for _, required := range []string{
		"ci-tool-bootstrap-contract-test:",
		"./scripts/test-tool-bootstrap-evidence.sh",
		"TestCIToolBootstrapRepositoryContract",
	} {
		if !strings.Contains(makefile, required) {
			violations = append(violations, "Makefile missing "+required)
		}
	}
	dependabot := state.files[".github/dependabot.yml"]
	if !strings.Contains(dependabot, "directory: /tools/ci/npm") || !strings.Contains(dependabot, "ci-tool-maintenance") {
		violations = append(violations, "Dependabot policy must cover /tools/ci/npm")
	}
	gitignore := state.files[".gitignore"]
	for _, required := range []string{
		"node_modules/",
		"!testdata/toolbootstrapevidence/*/artifacts/cyclonedx-npm/node_modules/",
		"testdata/toolbootstrapevidence/*/artifacts/cyclonedx-npm/node_modules/*",
		"!testdata/toolbootstrapevidence/*/artifacts/cyclonedx-npm/node_modules/@cyclonedx/",
		"testdata/toolbootstrapevidence/*/artifacts/cyclonedx-npm/node_modules/@cyclonedx/*",
		"!testdata/toolbootstrapevidence/*/artifacts/cyclonedx-npm/node_modules/@cyclonedx/cyclonedx-npm/",
		"testdata/toolbootstrapevidence/*/artifacts/cyclonedx-npm/node_modules/@cyclonedx/cyclonedx-npm/*",
		"!testdata/toolbootstrapevidence/*/artifacts/cyclonedx-npm/node_modules/@cyclonedx/cyclonedx-npm/package.json",
	} {
		if !strings.Contains(gitignore, required) {
			violations = append(violations, ".gitignore must preserve tracked npm evidence fixture path "+required)
		}
	}
	buildScript := state.files["scripts/build-ci-tool.sh"]
	for _, required := range []string{"go mod verify", "-mod=readonly", "-trimpath", "npm ci --ignore-scripts", "configs/release/tool-platform.json", `output_parent=$(cd "$output_parent" && pwd -P)`} {
		if !strings.Contains(buildScript, required) {
			violations = append(violations, "locked tool bootstrap script missing "+required)
		}
	}
	return violations
}

func cloneCIToolBootstrapState(state ciToolBootstrapState) ciToolBootstrapState {
	clone := ciToolBootstrapState{
		contract: state.contract,
		gates:    state.gates,
		files:    make(map[string]string, len(state.files)),
		packages: append([]string{}, state.packages...),
	}
	for path, content := range state.files {
		clone.files[path] = content
	}
	clone.contract.Bootstrap.Tools = append([]toolbootstrapevidence.BootstrapTool{}, state.contract.Bootstrap.Tools...)
	clone.gates.Tools = make(map[string]string, len(state.gates.Tools))
	for key, value := range state.gates.Tools {
		clone.gates.Tools[key] = value
	}
	return clone
}

func assertCIToolBootstrapRejected(t *testing.T, state ciToolBootstrapState, want string) {
	t.Helper()
	violations := ciToolBootstrapViolations(state)
	if !strings.Contains(strings.Join(violations, "\n"), want) {
		t.Fatalf("missing rejection %q in %v", want, violations)
	}
}
