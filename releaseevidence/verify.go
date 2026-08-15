package releaseevidence

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var digestRE = regexp.MustCompile(`^[0-9a-f]{64}$`)
var sourceRE = regexp.MustCompile(`^[0-9a-f]{40}$`)

const maxBundleMemberBytes int64 = 1 << 30

type metadata struct {
	Schema, ContractSchema, ContractSHA256, SourceRef, Target, CIRunner, RunnerArch, Artifact, SHA256, GoVersion, CGOEnabled string
	Reproducible                                                                                                             []string
	SourceDateEpoch                                                                                                          json.Number
	BuildFlags                                                                                                               metadataFlags
}
type metadataFlags struct {
	Trimpath                                     bool
	BuildVCS                                     bool
	Mod, BuildID, LinkerBuildID, VersionVariable string
}

func (m *metadata) UnmarshalJSON(data []byte) error {
	type raw struct {
		Schema          string      `json:"schema"`
		ContractSchema  string      `json:"contract_schema"`
		ContractSHA256  string      `json:"contract_sha256"`
		SourceRef       string      `json:"source_ref"`
		Target          string      `json:"target"`
		CIRunner        string      `json:"ci_runner"`
		RunnerArch      string      `json:"runner_arch"`
		Artifact        string      `json:"artifact"`
		SHA256          string      `json:"sha256"`
		Reproducible    []string    `json:"reproducible_pair_sha256"`
		GoVersion       string      `json:"go_version"`
		CGOEnabled      string      `json:"cgo_enabled"`
		SourceDateEpoch json.Number `json:"source_date_epoch"`
		BuildFlags      struct {
			Trimpath        bool   `json:"trimpath"`
			BuildVCS        bool   `json:"buildvcs"`
			Mod             string `json:"mod"`
			BuildID         string `json:"buildid"`
			LinkerBuildID   string `json:"linker_build_id"`
			VersionVariable string `json:"version_variable"`
		} `json:"build_flags"`
	}
	var r raw
	if err := parseBytes(data, &r); err != nil {
		return err
	}
	m.Schema = r.Schema
	m.ContractSchema = r.ContractSchema
	m.ContractSHA256 = r.ContractSHA256
	m.SourceRef = r.SourceRef
	m.Target = r.Target
	m.CIRunner = r.CIRunner
	m.RunnerArch = r.RunnerArch
	m.Artifact = r.Artifact
	m.SHA256 = r.SHA256
	m.Reproducible = r.Reproducible
	m.GoVersion = r.GoVersion
	m.CGOEnabled = r.CGOEnabled
	m.SourceDateEpoch = r.SourceDateEpoch
	m.BuildFlags = metadataFlags{r.BuildFlags.Trimpath, r.BuildFlags.BuildVCS, r.BuildFlags.Mod, r.BuildFlags.BuildID, r.BuildFlags.LinkerBuildID, r.BuildFlags.VersionVariable}
	return nil
}

func VerifyDirectory(bundleDir, buildContractPath, toolContractPath string) Report {
	r := Report{Schema: Schema, Violations: []string{}}
	fail := func(s string) { r.Violations = append(r.Violations, s) }
	base, err := filepath.Abs(bundleDir)
	if err != nil {
		fail("bundle path is invalid")
		return r
	}
	base, err = filepath.EvalSymlinks(base)
	if err != nil {
		fail("bundle directory is unavailable")
		return r
	}
	manifestPath, err := safeMember(base, "release-evidence.json")
	if err != nil {
		fail("release manifest path is unsafe")
		return r
	}
	var bundle Bundle
	if err := parseFile(manifestPath, &bundle); err != nil {
		fail("release manifest is invalid")
		return r
	}
	r.Targets = len(bundle.Targets)
	r.SBOMs = len(bundle.SBOMs)
	var build BuildContract
	if err := parseFile(buildContractPath, &build); err != nil {
		fail("build contract is invalid")
		return r
	}
	var tools ToolContract
	if err := parseFile(toolContractPath, &tools); err != nil {
		fail("tool contract is invalid")
		return r
	}
	buildHash, e1 := hashFile(buildContractPath)
	toolHash, e2 := hashFile(toolContractPath)
	if e1 != nil || e2 != nil {
		fail("contract digest is unavailable")
		return r
	}
	if bundle.Schema != Schema {
		fail("bundle schema mismatch")
	}
	if !sourceRE.MatchString(bundle.SourceRef) {
		fail("source ref is malformed")
	}
	if build.Schema != BuildSchema || !exactBuildContract(build) {
		fail("build contract mismatch")
	}
	if tools.Schema != ToolSchema || tools.Tools != exactTools() || !exactPlatforms(tools.Platforms) || !exactImages(tools.BaseImages) {
		fail("tool contract mismatch")
	}
	if bundle.BuildContractSHA256 != buildHash || bundle.ToolContractSHA256 != toolHash {
		fail("contract digest mismatch")
	}
	if bundle.Tools != tools.Tools {
		fail("tool pin mismatch")
	}
	if !bundle.Claims.explicitFalse() {
		fail("release status claims must remain false")
	}
	buildTargets := map[string]BuildTarget{}
	for _, t := range build.Targets {
		if _, ok := buildTargets[t.ID]; ok {
			fail("duplicate build target")
		}
		buildTargets[t.ID] = t
	}
	seen := map[string]bool{}
	sourceDateEpoch := ""
	if len(bundle.Targets) != 2 {
		fail("release target count mismatch")
	}
	for i, t := range bundle.Targets {
		if i >= len(targetIDs) || t.ID != targetIDs[i] {
			fail("release target order mismatch")
		}
		bt, ok := buildTargets[t.ID]
		if !ok || seen[t.ID] {
			fail("release target set mismatch")
			continue
		}
		seen[t.ID] = true
		if t.Artifact != bt.Artifact || !digestRE.MatchString(t.SHA256) {
			fail("release artifact declaration mismatch")
		}
		artifact, err := safeMember(base, t.Artifact)
		if err != nil {
			fail("release artifact path is unsafe")
			continue
		}
		checksum, err1 := safeMember(base, t.ChecksumsFile)
		metadataPath, err2 := safeMember(base, t.MetadataFile)
		if err1 != nil || err2 != nil {
			fail("release metadata path is unsafe")
			continue
		}
		h, err := hashFile(artifact)
		if err != nil || h != t.SHA256 {
			fail("release artifact digest mismatch")
		}
		if err := verifyChecksum(checksum, t.SHA256, t.Artifact); err != nil {
			fail("checksum file mismatch")
		}
		var m metadata
		if err := parseFile(metadataPath, &m); err != nil || !metadataMatches(m, t, bt, buildHash, bundle.SourceRef, build.GoVersion) {
			fail("build metadata mismatch")
		} else if sourceDateEpoch == "" {
			sourceDateEpoch = m.SourceDateEpoch.String()
		} else if sourceDateEpoch != m.SourceDateEpoch.String() {
			fail("build metadata source-date epoch mismatch")
		}
	}
	if !seen["linux-amd64"] || !seen["linux-arm64"] {
		fail("release targets are incomplete")
	}
	components := map[string]bool{}
	sbomFiles := map[string]bool{}
	sbomDigests := map[string]bool{}
	if len(bundle.SBOMs) != 2 {
		fail("SBOM count mismatch")
	}
	expectedComponents := []string{"go", "client"}
	for i, s := range bundle.SBOMs {
		if i >= len(expectedComponents) || s.Component != expectedComponents[i] {
			fail("SBOM order mismatch")
		}
		if (s.Component != "go" && s.Component != "client") || components[s.Component] || !digestRE.MatchString(s.SHA256) {
			fail("SBOM set mismatch")
			continue
		}
		if sbomFiles[s.File] || sbomDigests[s.SHA256] {
			fail("SBOM bindings must be distinct")
			continue
		}
		components[s.Component] = true
		sbomFiles[s.File], sbomDigests[s.SHA256] = true, true
		p, err := safeMember(base, s.File)
		if err != nil {
			fail("SBOM path is unsafe")
			continue
		}
		h, err := hashFile(p)
		if err != nil || h != s.SHA256 {
			fail("SBOM digest mismatch")
		}
		if err := validateCycloneDX(p); err != nil {
			fail("SBOM is not normalized CycloneDX")
		}
	}
	if !components["go"] || !components["client"] {
		fail("SBOMs are incomplete")
	}
	provenancePath, err := safeMember(base, bundle.Provenance.File)
	if err != nil || !digestRE.MatchString(bundle.Provenance.SHA256) {
		fail("provenance declaration is invalid")
	} else if h, hashErr := hashFile(provenancePath); hashErr != nil || h != bundle.Provenance.SHA256 {
		fail("provenance digest mismatch")
	} else {
		var p Provenance
		if parseFile(provenancePath, &p) != nil || !provenanceMatches(p, bundle) {
			fail("provenance binding mismatch")
		}
	}
	r.Valid = len(r.Violations) == 0
	return r
}

func exactTools() Tools { return Tools{"v1.10.0", "6.0.1", "22.22.2", "10.9.7"} }
func exactPlatforms(p []Platform) bool {
	return len(p) == 2 && p[0] == (Platform{"linux-amd64", "ubuntu-24.04", "x86_64"}) && p[1] == (Platform{"linux-arm64", "ubuntu-24.04-arm", "aarch64"})
}
func exactImages(v []string) bool {
	expected := []string{"golang:1.26.6-bookworm@sha256:116d58cbd88c1297624acc6e967a060012422bacf9930927e23fb719189c6f36", "debian:bookworm-slim@sha256:abd67ffcfa541b485a3dff59865ab629aa048a6c613e639d36e7456b0b229241", "node:22.22.2-alpine@sha256:8ea2348b068a9544dae7317b4f3aafcdc032df1647bb7d768a05a5cad1a7683f", "nginx:alpine@sha256:4a73073bd557c65b759505da037898b61f1be6cbcc3c2c3aeac22d2a470c1752"}
	return len(v) == len(expected) && strings.Join(v, "\x00") == strings.Join(expected, "\x00")
}
func provenanceMatches(p Provenance, b Bundle) bool {
	if len(b.Targets) != 2 || len(b.SBOMs) != 2 || p.Schema != ProvenanceSchema || p.SourceRef != b.SourceRef || p.BuildContractSHA256 != b.BuildContractSHA256 || p.ToolContractSHA256 != b.ToolContractSHA256 || !p.Claims.explicitFalse() || len(p.Targets) != 2 || len(p.SBOMs) != 2 {
		return false
	}
	for i := range p.Targets {
		if p.Targets[i] != (ProvenanceEntry{b.Targets[i].ID, b.Targets[i].SHA256}) {
			return false
		}
	}
	for i := range p.SBOMs {
		if p.SBOMs[i] != (ProvenanceEntry{b.SBOMs[i].Component, b.SBOMs[i].SHA256}) {
			return false
		}
	}
	return true
}
func exactBuildContract(b BuildContract) bool {
	return b.Binary == "truerepublicd" && b.MainPackage == "." && b.GoVersion == "1.26.6" && b.CGOEnabled == "1" && b.SourceRef.Kind == "git-commit" && b.SourceRef.Pattern == "^[0-9a-f]{40}$" && len(b.Targets) == 2 && b.Targets[0] == (BuildTarget{"linux-amd64", "linux", "amd64", "ubuntu-24.04", "x86_64", "truerepublicd-linux-amd64"}) && b.Targets[1] == (BuildTarget{"linux-arm64", "linux", "arm64", "ubuntu-24.04-arm", "aarch64", "truerepublicd-linux-arm64"}) && b.BuildFlags.Trimpath && !b.BuildFlags.BuildVCS && b.BuildFlags.Mod == "readonly" && strings.Join(b.BuildFlags.LDFlags, "\x00") == strings.Join([]string{"-s", "-w", "-buildid=", "-X", "main.version={{source_ref}}", "-X", "main.upgradePlan=v0.4.1", "-linkmode=external", "-extldflags=-Wl,--build-id=none"}, "\x00")
}
func metadataMatches(m metadata, t Target, b BuildTarget, ch, source, goVersion string) bool {
	epoch, err := strconv.ParseInt(m.SourceDateEpoch.String(), 10, 64)
	return err == nil && epoch > 0 && m.Schema == BuildEvidenceSchema && m.ContractSchema == BuildSchema && m.ContractSHA256 == ch && m.SourceRef == source && m.Target == t.ID && m.CIRunner == b.CIRunner && m.RunnerArch == b.RunnerArch && m.Artifact == t.Artifact && m.SHA256 == t.SHA256 && len(m.Reproducible) == 2 && m.Reproducible[0] == t.SHA256 && m.Reproducible[1] == t.SHA256 && m.GoVersion == goVersion && m.CGOEnabled == "1" && m.BuildFlags == (metadataFlags{true, false, "readonly", "", "none", "main.version"})
}
func safeMember(base, name string) (string, error) {
	if name == "" || strings.Contains(name, `\`) || name != filepath.Base(name) || filepath.IsAbs(name) || name == "." || name == ".." {
		return "", errors.New("unsafe")
	}
	p := filepath.Join(base, name)
	resolved, err := filepath.EvalSymlinks(p)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(base, resolved)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", errors.New("escape")
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.Mode().IsRegular() {
		return "", errors.New("not regular")
	}
	return resolved, nil
}
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	written, err := io.CopyN(h, f, maxBundleMemberBytes+1)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	if written > maxBundleMemberBytes {
		return "", errors.New("bundle member exceeds byte limit")
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func verifyChecksum(path, digest, name string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	if !s.Scan() {
		return errors.New("empty")
	}
	if s.Text() != digest+"  "+name {
		return errors.New("line")
	}
	if s.Scan() || s.Err() != nil {
		return errors.New("extra")
	}
	return nil
}
func validateCycloneDX(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	value, err := parseValue(f)
	if err != nil {
		return err
	}
	v, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid")
	}
	format, fok := v["bomFormat"].(string)
	spec, sok := v["specVersion"].(string)
	_, cok := v["components"].([]any)
	if !fok || !sok || !cok || format != "CycloneDX" || spec == "" {
		return fmt.Errorf("invalid")
	}
	if _, exists := v["serialNumber"]; exists {
		return fmt.Errorf("not normalized")
	}
	if metadata, ok := v["metadata"].(map[string]any); ok {
		if _, exists := metadata["timestamp"]; exists {
			return fmt.Errorf("not normalized")
		}
	}
	if format != "CycloneDX" {
		return fmt.Errorf("invalid")
	}
	return nil
}
