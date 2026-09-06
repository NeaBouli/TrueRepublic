package toolbootstrapevidence

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fixture struct {
	root       string
	contract   string
	gates      string
	locksRoot  string
	artifacts  string
	evidence   string
	manifest   string
	lockDigest map[string]string
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func digest(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

func digestPath(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

const fixtureGoBinary = "synthetic govulncheck binary\n"
const fixtureNpmPackage = "{\"name\":\"@cyclonedx/cyclonedx-npm\",\"version\":\"6.0.1\"}\n"

func newFixture(t *testing.T) fixture {
	t.Helper()
	f := fixture{lockDigest: map[string]string{}}
	f.root = t.TempDir()
	f.contract = filepath.Join(f.root, "contract.json")
	f.gates = filepath.Join(f.root, "gates.json")
	f.locksRoot = filepath.Join(f.root, "repo")
	f.artifacts = filepath.Join(f.root, "artifacts")
	f.evidence = filepath.Join(f.root, "evidence")

	lockContents := map[string]string{
		"go_module":   "module truerepublic/tools/ci\n",
		"go_sum":      "synthetic go.sum\n",
		"npm_package": "{\"private\":true}\n",
		"npm_lock":    "{\"lockfileVersion\":3}\n",
	}
	for name, content := range lockContents {
		rel := "locks/" + name
		writeFile(t, filepath.Join(f.locksRoot, rel), content)
		f.lockDigest[name] = digest(content)
	}

	writeFile(t, f.contract, `{
  "schema": "truerepublic.release-tool-platform/v1",
  "tools": {"govulncheck": "v1.6.0", "cyclonedx_npm": "6.0.1"},
  "platforms": [{"id": "linux-amd64", "runner": "ubuntu-24.04", "arch": "x86_64"}],
  "base_images": ["example@sha256:`+strings.Repeat("0", 64)+`"],
  "bootstrap": {
    "go_module": "locks/go_module",
    "go_sum": "locks/go_sum",
    "npm_package": "locks/npm_package",
    "npm_lock": "locks/npm_lock",
    "tools": [
      {"id": "govulncheck", "kind": "go", "package": "golang.org/x/vuln/cmd/govulncheck", "module": "golang.org/x/vuln", "gates_key": "govulncheck", "artifact": "govulncheck"},
      {"id": "cyclonedx-npm", "kind": "npm", "package": "@cyclonedx/cyclonedx-npm", "gates_key": "cyclonedx_npm", "artifact": "node_modules/@cyclonedx/cyclonedx-npm/package.json"}
    ]
  }
}`)
	writeFile(t, f.gates, `{
  "version": "truerepublic.security-gates/v1",
  "review_cadence_days": 7,
  "exception_max_days": 30,
  "toolchains": {"go": "1.26.6"},
  "tools": {"govulncheck": "v1.6.0", "cyclonedx_npm": "6.0.1"},
  "actions": {},
  "go_vulnerability_exceptions": []
}`)

	writeFile(t, filepath.Join(f.artifacts, "govulncheck", "govulncheck"), fixtureGoBinary)
	writeFile(t, filepath.Join(f.artifacts, "cyclonedx-npm", "node_modules", "@cyclonedx", "cyclonedx-npm", "package.json"), fixtureNpmPackage)

	if err := os.MkdirAll(f.evidence, 0o755); err != nil {
		t.Fatal(err)
	}
	f.manifest = filepath.Join(f.evidence, ManifestFile)
	f.writeManifest(t, fixtureClaims("false"), "v1.6.0", digest(fixtureGoBinary), digest(fixtureNpmPackage))
	return f
}

func fixtureClaims(value string) string {
	return fmt.Sprintf(`{"signed":%[1]s,"published":%[1]s,"deployed":%[1]s,"production":%[1]s,"long_term_hermetic":%[1]s}`, value)
}

func (f fixture) writeManifest(t *testing.T, claims, goVersion, goDigest, npmDigest string) {
	t.Helper()
	tools := fmt.Sprintf(`{"id": "govulncheck", "kind": "go", "version": "%s", "artifact": {"file": "govulncheck/govulncheck", "sha256": "%s"}},
    {"id": "cyclonedx-npm", "kind": "npm", "version": "6.0.1", "artifact": {"file": "cyclonedx-npm/node_modules/@cyclonedx/cyclonedx-npm/package.json", "sha256": "%s"}}`,
		goVersion, goDigest, npmDigest)
	f.writeManifestTools(t, claims, tools)
}

func (f fixture) writeManifestTools(t *testing.T, claims, tools string) {
	t.Helper()
	writeFile(t, f.manifest, fmt.Sprintf(`{
  "schema": "truerepublic.tool-bootstrap-evidence/v1",
  "contract_sha256": "%s",
  "gates_sha256": "%s",
  "claims": %s,
  "locks": {
    "go_module": {"file": "locks/go_module", "sha256": "%s"},
    "go_sum": {"file": "locks/go_sum", "sha256": "%s"},
    "npm_package": {"file": "locks/npm_package", "sha256": "%s"},
    "npm_lock": {"file": "locks/npm_lock", "sha256": "%s"}
  },
  "tools": [
    %s
  ]
}`, digestPath(t, f.contract), digestPath(t, f.gates), claims,
		f.lockDigest["go_module"], f.lockDigest["go_sum"], f.lockDigest["npm_package"], f.lockDigest["npm_lock"],
		tools))
}

func (f fixture) verify() Report {
	return Verify(f.evidence, f.artifacts, f.contract, f.gates, f.locksRoot)
}

func requireViolation(t *testing.T, report Report, want string) {
	t.Helper()
	if report.Valid {
		t.Fatalf("expected violation %q, report is valid", want)
	}
	for _, violation := range report.Violations {
		if strings.Contains(violation, want) {
			return
		}
	}
	t.Fatalf("missing violation %q in %v", want, report.Violations)
}

func TestVerifyValidFixture(t *testing.T) {
	f := newFixture(t)
	report := f.verify()
	if !report.Valid || report.Tools != 2 || len(report.Violations) != 0 {
		t.Fatalf("valid fixture rejected: %+v", report)
	}
	second := f.verify()
	firstJSON, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	if string(firstJSON) != string(secondJSON) {
		t.Fatal("verification is not deterministic")
	}
}

func TestVerifyRejectsTrueClaims(t *testing.T) {
	f := newFixture(t)
	f.writeManifest(t, fixtureClaims("false")+`,"extra":true`, "v1.6.0", digest(fixtureGoBinary), digest(fixtureNpmPackage))
	// The extra key must fail strict parsing before claim evaluation.
	requireViolation(t, f.verify(), "not strict bounded JSON")

	f = newFixture(t)
	f.writeManifest(t, `{"signed":true,"published":false,"deployed":false,"production":false,"long_term_hermetic":false}`, "v1.6.0", digest(fixtureGoBinary), digest(fixtureNpmPackage))
	requireViolation(t, f.verify(), "claims must be explicitly present and false")

	f = newFixture(t)
	f.writeManifest(t, `{"signed":false,"published":false,"deployed":false,"production":false}`, "v1.6.0", digest(fixtureGoBinary), digest(fixtureNpmPackage))
	requireViolation(t, f.verify(), "not strict bounded JSON")
}

func TestVerifyRejectsMalformedJSON(t *testing.T) {
	f := newFixture(t)
	data, err := os.ReadFile(f.manifest)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, f.manifest, string(data)+" {}")
	requireViolation(t, f.verify(), "not strict bounded JSON")

	f = newFixture(t)
	writeFile(t, f.manifest, `{"schema":"truerepublic.tool-bootstrap-evidence/v1","schema":"truerepublic.tool-bootstrap-evidence/v1"}`)
	requireViolation(t, f.verify(), "not strict bounded JSON")

	f = newFixture(t)
	writeFile(t, f.manifest, `{"schema":"truerepublic.tool-bootstrap-evidence/v2"}`)
	requireViolation(t, f.verify(), "tool-bootstrap evidence schema mismatch")
}

func TestVerifyRejectsSchemaMismatch(t *testing.T) {
	f := newFixture(t)
	writeFile(t, f.contract, strings.Replace(mustRead(t, f.contract), ContractSchema, "truerepublic.release-tool-platform/v2", 1))
	report := f.verify()
	requireViolation(t, report, "tool-platform contract schema mismatch")
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestVerifyRejectsContractDigestDrift(t *testing.T) {
	f := newFixture(t)
	writeFile(t, f.contract, mustRead(t, f.contract)+"\n")
	requireViolation(t, f.verify(), "tool-platform contract digest drift")
}

func TestVerifyRejectsGatesDigestDrift(t *testing.T) {
	f := newFixture(t)
	writeFile(t, f.gates, mustRead(t, f.gates)+"\n")
	requireViolation(t, f.verify(), "security gate contract digest drift")
}

func TestVerifyRejectsVersionDrift(t *testing.T) {
	f := newFixture(t)
	f.writeManifest(t, fixtureClaims("false"), "v9.9.9", digest(fixtureGoBinary), digest(fixtureNpmPackage))
	requireViolation(t, f.verify(), "tool govulncheck version drift")
}

func TestVerifyRejectsArtifactDigestDrift(t *testing.T) {
	f := newFixture(t)
	f.writeManifest(t, fixtureClaims("false"), "v1.6.0", strings.Repeat("0", 64), digest(fixtureNpmPackage))
	requireViolation(t, f.verify(), "tool govulncheck artifact digest drift")
}

func TestVerifyRejectsArtifactContentDrift(t *testing.T) {
	f := newFixture(t)
	writeFile(t, filepath.Join(f.artifacts, "govulncheck", "govulncheck"), "drifted binary\n")
	requireViolation(t, f.verify(), "tool govulncheck artifact digest drift")
}

func TestVerifyRejectsMissingTool(t *testing.T) {
	f := newFixture(t)
	f.writeManifestTools(t, fixtureClaims("false"),
		fmt.Sprintf(`{"id": "govulncheck", "kind": "go", "version": "v1.6.0", "artifact": {"file": "govulncheck/govulncheck", "sha256": "%s"}}`, digest(fixtureGoBinary)))
	report := f.verify()
	requireViolation(t, report, "tool set is incomplete")
	requireViolation(t, report, "configured tool cyclonedx-npm is missing")
}

func TestVerifyRejectsDuplicateTool(t *testing.T) {
	f := newFixture(t)
	entry := fmt.Sprintf(`{"id": "govulncheck", "kind": "go", "version": "v1.6.0", "artifact": {"file": "govulncheck/govulncheck", "sha256": "%s"}}`, digest(fixtureGoBinary))
	f.writeManifestTools(t, fixtureClaims("false"), entry+",\n    "+entry)
	requireViolation(t, f.verify(), "duplicate tool govulncheck")
}

func TestVerifyRejectsUndeclaredArtifactsMember(t *testing.T) {
	f := newFixture(t)
	if err := os.MkdirAll(filepath.Join(f.artifacts, "unexpected"), 0o755); err != nil {
		t.Fatal(err)
	}
	requireViolation(t, f.verify(), "undeclared member unexpected")
}

func TestVerifyRejectsMissingArtifactsTool(t *testing.T) {
	f := newFixture(t)
	if err := os.RemoveAll(filepath.Join(f.artifacts, "cyclonedx-npm")); err != nil {
		t.Fatal(err)
	}
	report := f.verify()
	requireViolation(t, report, "missing tool cyclonedx-npm")
}

func TestVerifyRejectsSymlinkedArtifact(t *testing.T) {
	f := newFixture(t)
	target := filepath.Join(f.artifacts, "govulncheck", "govulncheck")
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(f.artifacts, "cyclonedx-npm", "node_modules", "@cyclonedx", "cyclonedx-npm", "package.json"), target); err != nil {
		t.Fatal(err)
	}
	requireViolation(t, f.verify(), "symlinked path component")
}

func TestVerifyRejectsExtraGoToolMember(t *testing.T) {
	f := newFixture(t)
	writeFile(t, filepath.Join(f.artifacts, "govulncheck", "extra"), "extra\n")
	requireViolation(t, f.verify(), "must build exactly its declared artifact")
}

func TestVerifyRejectsLockDigestDrift(t *testing.T) {
	f := newFixture(t)
	writeFile(t, filepath.Join(f.locksRoot, "locks", "go_sum"), "drifted go.sum\n")
	requireViolation(t, f.verify(), "lock go_sum digest drift")
}

func TestVerifyRejectsSymlinkedLock(t *testing.T) {
	f := newFixture(t)
	target := filepath.Join(f.locksRoot, "locks", "go_module")
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(f.locksRoot, "locks", "go_sum"), target); err != nil {
		t.Fatal(err)
	}
	requireViolation(t, f.verify(), "lock go_module is unavailable or unsafe")
}

func TestVerifyRejectsLockPathEscape(t *testing.T) {
	f := newFixture(t)
	writeFile(t, f.contract, strings.Replace(mustRead(t, f.contract), `"go_sum": "locks/go_sum"`, `"go_sum": "../outside"`, 1))
	report := f.verify()
	requireViolation(t, report, "lock path go_sum is not a clean repository-relative path")
}

func TestVerifyRejectsEvidenceDirExtras(t *testing.T) {
	f := newFixture(t)
	writeFile(t, filepath.Join(f.evidence, "extra.json"), "{}")
	requireViolation(t, f.verify(), "must contain exactly "+ManifestFile)
}

func TestVerifyRejectsDuplicateContractTool(t *testing.T) {
	f := newFixture(t)
	contract := mustRead(t, f.contract)
	entry := `{"id": "govulncheck", "kind": "go", "package": "golang.org/x/vuln/cmd/govulncheck", "module": "golang.org/x/vuln", "gates_key": "govulncheck", "artifact": "govulncheck"}`
	contract = strings.Replace(contract, `"tools": [
      {`, `"tools": [
      `+entry+`,
      {`, 1)
	writeFile(t, f.contract, contract)
	report := f.verify()
	requireViolation(t, report, "duplicate tool id govulncheck")
}

func TestVerifyRejectsMissingEvidenceDir(t *testing.T) {
	f := newFixture(t)
	report := Verify(filepath.Join(f.root, "missing"), f.artifacts, f.contract, f.gates, f.locksRoot)
	requireViolation(t, report, "evidence directory")
}

func TestRunCLI(t *testing.T) {
	f := newFixture(t)
	var stdout, stderr strings.Builder
	code := Run([]string{"verify", "--evidence", f.evidence, "--artifacts", f.artifacts,
		"--contract", f.contract, "--gates", f.gates, "--locks-root", f.locksRoot, "--output", "json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("valid evidence exited %d: %s", code, stderr.String())
	}
	var report Report
	if err := json.Unmarshal([]byte(stdout.String()), &report); err != nil || !report.Valid {
		t.Fatalf("invalid report %q: %v", stdout.String(), err)
	}

	stdout.Reset()
	if code := Run([]string{"verify", "--evidence", f.evidence, "--artifacts", f.artifacts,
		"--contract", f.contract, "--gates", f.gates, "--locks-root", f.locksRoot, "--bogus"}, &stdout, &stderr); code != 2 {
		t.Fatalf("unknown flag exited %d, want 2", code)
	}
	if code := Run([]string{"verify"}, &stdout, &stderr); code != 2 {
		t.Fatalf("missing flags exited %d, want 2", code)
	}
	if code := Run([]string{"explode"}, &stdout, &stderr); code != 2 {
		t.Fatalf("unknown subcommand exited %d, want 2", code)
	}

	writeFile(t, f.manifest, strings.Replace(mustRead(t, f.manifest), `"v1.6.0"`, `"v9.9.9"`, 1))
	if code := Run([]string{"verify", "--evidence", f.evidence, "--artifacts", f.artifacts,
		"--contract", f.contract, "--gates", f.gates, "--locks-root", f.locksRoot}, &stdout, &stderr); code != 1 {
		t.Fatalf("drifted evidence exited %d, want 1", code)
	}
	if !strings.Contains(stdout.String(), "version drift") {
		t.Fatalf("text report missing drift detail: %q", stdout.String())
	}
}
