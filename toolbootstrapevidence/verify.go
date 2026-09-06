package toolbootstrapevidence

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	toolIDPattern   = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)
	gatesKeyPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_]{0,63}$`)
	versionPattern  = regexp.MustCompile(`^v?[0-9]+\.[0-9]+\.[0-9]+$`)
	sha256Pattern   = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// Verify checks one tool-bootstrap evidence directory against the artifacts of
// a completed build, the tool-platform contract, the security gate contract,
// and the repository-owned lock files below locksRoot. It never trusts
// evidence-declared values without recomputation and collects every violation.
func Verify(evidenceDir, artifactsDir, contractPath, gatesPath, locksRoot string) Report {
	report := Report{Schema: ReportSchema, Violations: []string{}}
	v := &reporter{report: &report}

	contract, contractOK := loadContract(contractPath, v)
	gates, gatesOK := loadGates(gatesPath, v)

	manifestPath, manifestOK := v.check("evidence directory", func() (string, error) {
		return evidenceManifestPath(evidenceDir)
	})
	var manifest Manifest
	if manifestOK {
		if err := parseFile(manifestPath, &manifest); err != nil {
			v.add("tool-bootstrap manifest is not strict bounded JSON")
			manifestOK = false
		}
	}

	if contractOK {
		if digest, err := digestFile(contractPath, MaxJSONBytes); err != nil {
			v.add("tool-platform contract digest is unavailable")
		} else if manifestOK && manifest.ContractSHA256 != digest {
			v.add("tool-platform contract digest drift")
		}
	}
	if gatesOK {
		if digest, err := digestFile(gatesPath, MaxJSONBytes); err != nil {
			v.add("security gate contract digest is unavailable")
		} else if manifestOK && manifest.GatesSHA256 != digest {
			v.add("security gate contract digest drift")
		}
	}

	if manifestOK {
		if manifest.Schema != Schema {
			v.add("tool-bootstrap evidence schema mismatch")
		}
		if !sha256Pattern.MatchString(manifest.ContractSHA256) || !sha256Pattern.MatchString(manifest.GatesSHA256) {
			v.add("contract digests must be lowercase SHA-256")
		}
		if !manifest.Claims.explicitFalse() {
			v.add("tool-bootstrap status claims must be explicitly present and false")
		}
	}

	if contractOK && manifestOK {
		verifyLocks(contract, manifest, locksRoot, v)
	}
	if contractOK && gatesOK && manifestOK {
		verifyTools(contract, gates, manifest, artifactsDir, v)
	}

	report.Tools = len(manifest.Tools)
	report.Valid = len(report.Violations) == 0
	return report
}

type reporter struct {
	report *Report
}

func (r *reporter) add(message string) {
	r.report.Violations = append(r.report.Violations, message)
}

func (r *reporter) check(context string, fn func() (string, error)) (string, bool) {
	result, err := fn()
	if err != nil {
		r.add(context + ": " + err.Error())
		return "", false
	}
	return result, true
}

func loadContract(path string, v *reporter) (Contract, bool) {
	var contract Contract
	if err := parseFile(path, &contract); err != nil {
		v.add("tool-platform contract is not strict bounded JSON")
		return contract, false
	}
	if contract.Schema != ContractSchema {
		v.add("tool-platform contract schema mismatch")
		return contract, false
	}
	for name, rel := range contract.lockPaths() {
		if !cleanRelPath(rel) {
			v.add("tool-platform lock path " + name + " is not a clean repository-relative path")
		}
	}
	if len(contract.Bootstrap.Tools) == 0 || len(contract.Bootstrap.Tools) > maxTools {
		v.add("tool-platform bootstrap tool count is out of bounds")
		return contract, false
	}
	seen := make(map[string]struct{}, len(contract.Bootstrap.Tools))
	valid := true
	for _, tool := range contract.Bootstrap.Tools {
		if !toolIDPattern.MatchString(tool.ID) {
			v.add("tool id " + tool.ID + " is malformed")
			valid = false
		}
		if _, exists := seen[tool.ID]; exists {
			v.add("duplicate tool id " + tool.ID + " in the tool-platform contract")
			valid = false
		}
		seen[tool.ID] = struct{}{}
		if tool.Kind != "go" && tool.Kind != "npm" {
			v.add("tool " + tool.ID + " has an unsupported kind")
			valid = false
		}
		if tool.Kind == "go" && tool.Module == "" {
			v.add("go tool " + tool.ID + " must declare its module")
			valid = false
		}
		if tool.Kind == "npm" && tool.Module != "" {
			v.add("npm tool " + tool.ID + " must not declare a Go module")
			valid = false
		}
		if tool.Package == "" || len(tool.Package) > 256 || strings.ContainsAny(tool.Package, " \t\n\"'") {
			v.add("tool " + tool.ID + " package is malformed")
			valid = false
		}
		if !gatesKeyPattern.MatchString(tool.GatesKey) {
			v.add("tool " + tool.ID + " gates key is malformed")
			valid = false
		}
		if !cleanRelPath(tool.Artifact) {
			v.add("tool " + tool.ID + " artifact is not a clean relative path")
			valid = false
		}
	}
	return contract, valid
}

func loadGates(path string, v *reporter) (Gates, bool) {
	var gates Gates
	if err := parseFile(path, &gates); err != nil {
		v.add("security gate contract is not strict bounded JSON")
		return gates, false
	}
	if gates.Version != GatesSchema {
		v.add("security gate contract schema mismatch")
		return gates, false
	}
	return gates, true
}

func (c Contract) lockPaths() map[string]string {
	return map[string]string{
		"go_module":   c.Bootstrap.GoModule,
		"go_sum":      c.Bootstrap.GoSum,
		"npm_package": c.Bootstrap.NpmPackage,
		"npm_lock":    c.Bootstrap.NpmLock,
	}
}

func verifyLocks(contract Contract, manifest Manifest, locksRoot string, v *reporter) {
	declared := map[string]FileDigest{
		"go_module":   manifest.Locks.GoModule,
		"go_sum":      manifest.Locks.GoSum,
		"npm_package": manifest.Locks.NpmPackage,
		"npm_lock":    manifest.Locks.NpmLock,
	}
	for name, rel := range contract.lockPaths() {
		binding, exists := declared[name]
		if !exists {
			v.add("tool-bootstrap locks are incomplete")
			continue
		}
		if binding.File != rel {
			v.add("lock " + name + " path does not match the tool-platform contract")
			continue
		}
		if !sha256Pattern.MatchString(binding.SHA256) {
			v.add("lock " + name + " digest must be lowercase SHA-256")
			continue
		}
		digest, err := digestLockedFile(locksRoot, rel, MaxLockBytes)
		if err != nil {
			v.add("lock " + name + " is unavailable or unsafe: " + err.Error())
			continue
		}
		if digest != binding.SHA256 {
			v.add("lock " + name + " digest drift")
		}
	}
}

func verifyTools(contract Contract, gates Gates, manifest Manifest, artifactsDir string, v *reporter) {
	if len(manifest.Tools) != len(contract.Bootstrap.Tools) {
		v.add("tool-bootstrap tool set is incomplete")
	}
	seen := make(map[string]struct{}, len(manifest.Tools))
	byID := make(map[string]ToolEvidence, len(manifest.Tools))
	for _, tool := range manifest.Tools {
		if _, exists := seen[tool.ID]; exists {
			v.add("duplicate tool " + tool.ID + " in the evidence manifest")
		}
		seen[tool.ID] = struct{}{}
		byID[tool.ID] = tool
		if !sha256Pattern.MatchString(tool.Artifact.SHA256) {
			v.add("tool " + tool.ID + " artifact digest must be lowercase SHA-256")
		}
	}
	for _, declared := range contract.Bootstrap.Tools {
		tool, exists := byID[declared.ID]
		if !exists {
			v.add("configured tool " + declared.ID + " is missing from the evidence")
			continue
		}
		if tool.Kind != declared.Kind {
			v.add("tool " + declared.ID + " kind drift")
		}
		expectedVersion := gates.Tools[declared.GatesKey]
		if !versionPattern.MatchString(expectedVersion) {
			v.add("security gate version for " + declared.GatesKey + " is not exact")
			continue
		}
		if tool.Version != expectedVersion {
			v.add("tool " + declared.ID + " version drift")
		}
		expectedFile := declared.ID + "/" + declared.Artifact
		if tool.Artifact.File != expectedFile {
			v.add("tool " + declared.ID + " artifact path does not match the tool-platform contract")
			continue
		}
		digest, err := digestLockedFile(artifactsDir, expectedFile, MaxArtifactBytes)
		if err != nil {
			v.add("tool " + declared.ID + " artifact is unavailable or unsafe: " + err.Error())
			continue
		}
		if digest != tool.Artifact.SHA256 {
			v.add("tool " + declared.ID + " artifact digest drift")
		}
	}
	verifyArtifactsRoot(contract, artifactsDir, v)
}

// verifyArtifactsRoot requires the artifacts root to contain exactly one
// non-symlink directory per configured tool and nothing else. Go tool
// directories must contain exactly the declared binary.
func verifyArtifactsRoot(contract Contract, artifactsDir string, v *reporter) {
	members, err := os.ReadDir(artifactsDir)
	if err != nil {
		v.add("artifacts directory is unavailable")
		return
	}
	expected := make(map[string]BootstrapTool, len(contract.Bootstrap.Tools))
	for _, tool := range contract.Bootstrap.Tools {
		expected[tool.ID] = tool
	}
	seen := make(map[string]struct{}, len(members))
	for _, member := range members {
		if member.Type()&os.ModeSymlink != 0 || !member.IsDir() {
			v.add("artifacts member " + member.Name() + " is not a regular directory")
			continue
		}
		tool, exists := expected[member.Name()]
		if !exists {
			v.add("artifacts directory contains undeclared member " + member.Name())
			continue
		}
		seen[member.Name()] = struct{}{}
		if tool.Kind != "go" {
			continue
		}
		toolMembers, err := os.ReadDir(filepath.Join(artifactsDir, member.Name()))
		if err != nil {
			v.add("tool " + tool.ID + " artifact directory is unavailable")
			continue
		}
		if len(toolMembers) != 1 || toolMembers[0].Name() != tool.Artifact ||
			toolMembers[0].Type()&os.ModeSymlink != 0 || !toolMembers[0].Type().IsRegular() {
			v.add("go tool " + tool.ID + " must build exactly its declared artifact")
		}
	}
	for id := range expected {
		if _, exists := seen[id]; !exists {
			v.add("artifacts directory is missing tool " + id)
		}
	}
}

func evidenceManifestPath(evidenceDir string) (string, error) {
	info, err := os.Lstat(evidenceDir)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("unavailable or symlinked")
	}
	members, err := os.ReadDir(evidenceDir)
	if err != nil {
		return "", errors.New("unreadable")
	}
	if len(members) != 1 || members[0].Name() != ManifestFile ||
		members[0].Type()&os.ModeSymlink != 0 || !members[0].Type().IsRegular() {
		return "", errors.New("must contain exactly " + ManifestFile)
	}
	return filepath.Join(evidenceDir, ManifestFile), nil
}

// cleanRelPath reports whether rel is a bounded relative path that cannot
// escape its root: no absolute prefix, no dot or dot-dot components, no
// backslashes, and a bounded length.
func cleanRelPath(rel string) bool {
	if rel == "" || len(rel) > 512 || path.IsAbs(rel) || strings.Contains(rel, "\\") {
		return false
	}
	if path.Clean(rel) != rel {
		return false
	}
	for _, component := range strings.Split(rel, "/") {
		if component == "" || component == "." || component == ".." {
			return false
		}
	}
	return true
}

// digestLockedFile recomputes the SHA-256 of root/rel without following
// symlinks in any path component and with a hard byte bound.
func digestLockedFile(root, rel string, maxBytes int64) (string, error) {
	if !cleanRelPath(rel) {
		return "", errors.New("unclean relative path")
	}
	current := root
	components := strings.Split(rel, "/")
	for i, component := range components {
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if err != nil {
			return "", errors.New("unavailable")
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", errors.New("symlinked path component")
		}
		if i < len(components)-1 {
			if !info.IsDir() {
				return "", errors.New("non-directory path component")
			}
			continue
		}
		if !info.Mode().IsRegular() {
			return "", errors.New("not a regular file")
		}
		if info.Size() > maxBytes {
			return "", errors.New("exceeds the byte limit")
		}
	}
	return digestFile(current, maxBytes)
}

func digestFile(path string, maxBytes int64) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(f, maxBytes+1))
	if err != nil {
		return "", err
	}
	if n > maxBytes {
		return "", fmt.Errorf("exceeds the byte limit")
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func formatViolations(report Report) string {
	if report.Valid {
		return fmt.Sprintf("tool-bootstrap evidence valid: %d tools, 0 violations", report.Tools)
	}
	lines := make([]string, 0, len(report.Violations)+1)
	lines = append(lines, fmt.Sprintf("tool-bootstrap evidence invalid: %d violation(s)", len(report.Violations)))
	for _, violation := range report.Violations {
		lines = append(lines, "- "+violation)
	}
	sort.Strings(lines[1:])
	return strings.Join(lines, "\n")
}
