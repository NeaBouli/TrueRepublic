package candidateevidence

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

var (
	digestRE = regexp.MustCompile(`^[0-9a-f]{64}$`)
	commitRE = regexp.MustCompile(CommitPattern)
	tagRE    = regexp.MustCompile(TagPattern)
)

const maxMemberBytes int64 = MaxJSONBytes

// daemonMetadata mirrors the existing deterministic daemon build metadata
// JSON schema for strict cross-binding.
type daemonMetadata struct {
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

// ociClaims mirrors the existing OCI evidence status claims.
type ociClaims struct {
	Signed     bool `json:"signed"`
	Published  bool `json:"published"`
	Production bool `json:"production"`
	present    bool
}

func (c *ociClaims) UnmarshalJSON(data []byte) error {
	var raw struct {
		Signed     *bool `json:"signed"`
		Published  *bool `json:"published"`
		Production *bool `json:"production"`
	}
	if err := parseBytes(data, &raw); err != nil {
		return err
	}
	if raw.Signed == nil || raw.Published == nil || raw.Production == nil {
		return errors.New("OCI status claims are incomplete")
	}
	c.Signed, c.Published, c.Production, c.present = *raw.Signed, *raw.Published, *raw.Production, true
	return nil
}

func (c ociClaims) explicitFalse() bool {
	return c.present && !c.Signed && !c.Published && !c.Production
}

// ociBundle mirrors the existing per-platform OCI evidence manifest.
type ociBundle struct {
	Schema         string             `json:"schema"`
	SourceRef      string             `json:"source_ref"`
	ContractSHA256 string             `json:"contract_sha256"`
	Platform       string             `json:"platform"`
	Claims         ociClaims          `json:"claims"`
	Targets        []ociTargetArchive `json:"targets"`
}

type ociTargetArchive struct {
	ID     string              `json:"id"`
	Builds []ociArchiveBinding `json:"builds"`
}

type ociArchiveBinding struct {
	File   string `json:"file"`
	SHA256 string `json:"sha256"`
}

// ociDigestReport mirrors the existing OCI evidence verification report.
type ociDigestReport struct {
	Schema     string           `json:"schema"`
	Valid      bool             `json:"valid"`
	Platform   string           `json:"platform"`
	Targets    int              `json:"targets"`
	Images     []ociImageReport `json:"images"`
	LayerDiffs []any            `json:"layer_diffs"`
	Violations []string         `json:"violations"`
}

type ociImageReport struct {
	ID         string   `json:"id"`
	Repetition int      `json:"repetition"`
	Index      string   `json:"index_sha256"`
	Manifest   string   `json:"manifest_sha256"`
	Config     string   `json:"config_sha256"`
	Layers     []string `json:"layer_sha256"`
}

func expectedBinaryTargets() []ContractBinaryTarget {
	return []ContractBinaryTarget{
		{ID: "linux-amd64", Artifact: "truerepublicd-linux-amd64", CIRunner: "ubuntu-24.04", RunnerArch: "x86_64"},
		{ID: "linux-arm64", Artifact: "truerepublicd-linux-arm64", CIRunner: "ubuntu-24.04-arm", RunnerArch: "aarch64"},
	}
}

func expectedOCITargets() []ContractOCITarget {
	return []ContractOCITarget{
		{ID: "daemon-linux-amd64", Platform: "linux/amd64"},
		{ID: "client-web-linux-amd64", Platform: "linux/amd64"},
		{ID: "daemon-linux-arm64", Platform: "linux/arm64"},
		{ID: "client-web-linux-arm64", Platform: "linux/arm64"},
	}
}

func expectedOCIPlatforms() []string { return []string{"linux/amd64", "linux/arm64"} }

// VerifyDirectory fail-closed verifies one release-candidate evidence
// directory against the exact candidate contract. It rejects a symlinked
// evidence root and symlinked members, never leaves the evidence directory,
// and never reads payload content beyond the bound JSON and checksum members.
func VerifyDirectory(evidenceDir, contractPath string) Report {
	report := Report{Schema: ReportSchema, Violations: []string{}}
	fail := func(message string) { report.Violations = append(report.Violations, message) }

	base, err := filepath.Abs(evidenceDir)
	if err != nil {
		fail("evidence path is invalid")
		return report
	}
	info, err := os.Lstat(base)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		fail("evidence directory is unavailable or symlinked")
		return report
	}

	manifestPath, err := safeMember(base, "candidate-evidence.json")
	if err != nil {
		fail("candidate manifest is unsafe")
		return report
	}
	var manifest Manifest
	if err := parseFile(manifestPath, &manifest); err != nil {
		fail("candidate manifest is invalid")
		return report
	}
	var contract Contract
	if err := parseFile(contractPath, &contract); err != nil {
		fail("candidate contract is invalid")
		return report
	}
	contractHash, err := hashRegularFile(contractPath, MaxJSONBytes)
	if err != nil {
		fail("candidate contract digest is unavailable")
		return report
	}

	report.Tag = manifest.Source.Tag
	report.Commit = manifest.Source.Commit
	report.BinaryTargets = len(manifest.BinaryTargets)
	for _, platform := range manifest.OCIPlatforms {
		report.OCITargets += len(platform.Targets)
	}

	if manifest.Schema != Schema {
		fail("candidate evidence schema mismatch")
	}
	if !tagRE.MatchString(manifest.Source.Tag) {
		fail("candidate tag is malformed")
	}
	if !commitRE.MatchString(manifest.Source.Commit) {
		fail("candidate commit is malformed")
	}
	if err := validateContract(contract); err != nil {
		fail("candidate contract mismatch: " + err.Error())
		return report
	}
	if manifest.ContractSHA256 != contractHash {
		fail("candidate contract digest mismatch")
	}
	if !manifest.Claims.explicitFalse() {
		fail("candidate status claims must remain false")
	}

	seenFiles := map[string]bool{"candidate-evidence.json": true}
	verifyBinaryTargets(fail, base, &manifest, &contract, seenFiles)
	verifyOCIPlatforms(fail, base, &manifest, &contract, seenFiles)
	verifyExactMembers(fail, base, seenFiles)

	report.Valid = len(report.Violations) == 0
	return report
}

// verifyExactMembers keeps the evidence directory metadata-only and
// self-contained: every regular member must be declared exactly once by the
// candidate manifest, and no binary, archive, symlink, directory, or unrelated
// file may ride alongside the bound evidence.
func verifyExactMembers(fail func(string), base string, seenFiles map[string]bool) {
	entries, err := os.ReadDir(base)
	if err != nil {
		fail("evidence directory cannot be enumerated")
		return
	}
	if len(entries) != len(seenFiles) {
		fail("evidence member count mismatch")
	}
	for _, entry := range entries {
		if !seenFiles[entry.Name()] {
			fail("evidence directory contains an undeclared member")
		}
	}
}

func validateContract(contract Contract) error {
	want := ContractSource{CommitKind: "git-commit", CommitPattern: CommitPattern, TagKind: "simulated-future-tag", TagPattern: TagPattern}
	if contract.Schema != ContractSchema || contract.Source != want {
		return errors.New("schema or source identity")
	}
	if !digestRE.MatchString(contract.BinaryContractSHA256) || !digestRE.MatchString(contract.OCIContractSHA256) {
		return errors.New("pinned contract digests")
	}
	if contract.OCIRepetitions != 2 {
		return errors.New("OCI repetition count")
	}
	if !binaryTargetsEqual(contract.BinaryTargets, expectedBinaryTargets()) {
		return errors.New("binary targets")
	}
	if !ociTargetsEqual(contract.OCITargets, expectedOCITargets()) {
		return errors.New("OCI targets")
	}
	return nil
}

func binaryTargetsEqual(got, want []ContractBinaryTarget) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func ociTargetsEqual(got, want []ContractOCITarget) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func verifyBinaryTargets(fail func(string), base string, manifest *Manifest, contract *Contract, seenFiles map[string]bool) {
	if len(manifest.BinaryTargets) != len(contract.BinaryTargets) {
		fail("binary target count mismatch")
	}
	seen := map[string]bool{}
	for index, binding := range manifest.BinaryTargets {
		if index >= len(contract.BinaryTargets) || binding.ID != contract.BinaryTargets[index].ID {
			fail("binary target order mismatch")
			continue
		}
		expected := contract.BinaryTargets[index]
		if seen[binding.ID] {
			fail("binary target set mismatch")
			continue
		}
		seen[binding.ID] = true
		if binding.Artifact != expected.Artifact || !digestRE.MatchString(binding.ArtifactSHA256) {
			fail("binary artifact declaration mismatch for " + binding.ID)
			continue
		}
		for _, member := range []FileDigest{binding.Metadata, binding.Checksums} {
			if !digestRE.MatchString(member.SHA256) {
				fail("binary evidence digest is malformed for " + binding.ID)
				continue
			}
			if seenFiles[member.File] {
				fail("binary evidence file is duplicated for " + binding.ID)
				continue
			}
			seenFiles[member.File] = true
		}
		metadataPath, err := boundMember(base, binding.Metadata)
		if err != nil {
			fail("binary metadata path is unsafe for " + binding.ID)
			continue
		}
		checksumPath, err := boundMember(base, binding.Checksums)
		if err != nil {
			fail("binary checksum path is unsafe for " + binding.ID)
			continue
		}
		var metadata daemonMetadata
		if err := parseFile(metadataPath, &metadata); err != nil || !metadataMatches(metadata, binding, expected, contract.BinaryContractSHA256, manifest.Source.Commit) {
			fail("binary metadata mismatch for " + binding.ID)
		}
		if err := verifyChecksum(checksumPath, binding.ArtifactSHA256, binding.Artifact); err != nil {
			fail("binary checksum mismatch for " + binding.ID)
		}
	}
	for _, expected := range contract.BinaryTargets {
		if !seen[expected.ID] {
			fail("binary targets are incomplete")
		}
	}
}

func metadataMatches(m daemonMetadata, binding BinaryTargetBinding, expected ContractBinaryTarget, contractHash, commit string) bool {
	epoch, err := strconv.ParseInt(m.SourceDateEpoch.String(), 10, 64)
	return err == nil && epoch > 0 &&
		m.Schema == BuildEvidenceSchema &&
		m.ContractSchema == BuildContractSchema &&
		m.ContractSHA256 == contractHash &&
		m.SourceRef == commit &&
		m.Target == binding.ID &&
		m.CIRunner == expected.CIRunner &&
		m.RunnerArch == expected.RunnerArch &&
		m.Artifact == binding.Artifact &&
		m.SHA256 == binding.ArtifactSHA256 &&
		len(m.Reproducible) == 2 && m.Reproducible[0] == binding.ArtifactSHA256 && m.Reproducible[1] == binding.ArtifactSHA256 &&
		m.GoVersion == "1.26.6" &&
		m.CGOEnabled == "1" &&
		m.BuildFlags.Trimpath && !m.BuildFlags.BuildVCS && m.BuildFlags.Mod == "readonly" &&
		m.BuildFlags.BuildID == "" && m.BuildFlags.LinkerBuildID == "none" && m.BuildFlags.VersionVariable == "main.version"
}

func verifyOCIPlatforms(fail func(string), base string, manifest *Manifest, contract *Contract, seenFiles map[string]bool) {
	platforms := expectedOCIPlatforms()
	if len(manifest.OCIPlatforms) != len(platforms) {
		fail("OCI platform count mismatch")
	}
	seen := map[string]bool{}
	for index, binding := range manifest.OCIPlatforms {
		if index >= len(platforms) || binding.Platform != platforms[index] || seen[binding.Platform] {
			fail("OCI platform order or set mismatch")
			continue
		}
		seen[binding.Platform] = true
		for _, member := range []FileDigest{binding.Evidence, binding.Report} {
			if !digestRE.MatchString(member.SHA256) {
				fail("OCI evidence digest is malformed for " + binding.Platform)
				continue
			}
			if seenFiles[member.File] {
				fail("OCI evidence file is duplicated for " + binding.Platform)
				continue
			}
			seenFiles[member.File] = true
		}
		evidencePath, err := boundMember(base, binding.Evidence)
		if err != nil {
			fail("OCI evidence path is unsafe for " + binding.Platform)
			continue
		}
		reportPath, err := boundMember(base, binding.Report)
		if err != nil {
			fail("OCI report path is unsafe for " + binding.Platform)
			continue
		}
		expectedTargets := ociTargetsForPlatform(contract.OCITargets, binding.Platform)
		if len(binding.Targets) != len(expectedTargets) {
			fail("OCI target count mismatch for " + binding.Platform)
			continue
		}
		var bundle ociBundle
		if err := parseFile(evidencePath, &bundle); err != nil || !bundleMatches(bundle, binding, expectedTargets, contract, manifest.Source.Commit) {
			fail("OCI evidence mismatch for " + binding.Platform)
		}
		var digestReport ociDigestReport
		if err := parseFile(reportPath, &digestReport); err != nil || !reportMatches(digestReport, binding, expectedTargets) {
			fail("OCI digest report mismatch for " + binding.Platform)
		}
	}
	for _, platform := range platforms {
		if !seen[platform] {
			fail("OCI platforms are incomplete")
		}
	}
}

func ociTargetsForPlatform(targets []ContractOCITarget, platform string) []ContractOCITarget {
	var selected []ContractOCITarget
	for _, target := range targets {
		if target.Platform == platform {
			selected = append(selected, target)
		}
	}
	return selected
}

func bundleMatches(bundle ociBundle, binding OCIPlatformBinding, expected []ContractOCITarget, contract *Contract, commit string) bool {
	if bundle.Schema != OCIEvidenceSchema || bundle.SourceRef != commit || bundle.ContractSHA256 != contract.OCIContractSHA256 ||
		bundle.Platform != binding.Platform || !bundle.Claims.explicitFalse() || len(bundle.Targets) != len(expected) {
		return false
	}
	seenBuilds := map[string]bool{}
	for index, target := range bundle.Targets {
		if target.ID != expected[index].ID || len(target.Builds) != contract.OCIRepetitions {
			return false
		}
		for _, build := range target.Builds {
			if !digestRE.MatchString(build.SHA256) || !safeMemberName(build.File) || seenBuilds[build.File] {
				return false
			}
			seenBuilds[build.File] = true
		}
	}
	return true
}

func reportMatches(digestReport ociDigestReport, binding OCIPlatformBinding, expected []ContractOCITarget) bool {
	if digestReport.Schema != OCIReportSchema || !digestReport.Valid || digestReport.Platform != binding.Platform ||
		digestReport.Targets != len(expected) || len(digestReport.Violations) != 0 || len(digestReport.LayerDiffs) != 0 ||
		len(digestReport.Images) != len(expected) {
		return false
	}
	for index, image := range digestReport.Images {
		declared := binding.Targets[index]
		if image.ID != expected[index].ID || declared.ID != image.ID || image.Repetition != 0 ||
			!digestRE.MatchString(image.Index) || !digestRE.MatchString(image.Manifest) || !digestRE.MatchString(image.Config) ||
			len(image.Layers) == 0 || len(image.Layers) > maxOCILayers || len(declared.LayerSHA256) != len(image.Layers) {
			return false
		}
		if declared.IndexSHA256 != image.Index || declared.ManifestSHA256 != image.Manifest || declared.ConfigSHA256 != image.Config {
			return false
		}
		for layer, layerDigest := range image.Layers {
			if !digestRE.MatchString(layerDigest) || declared.LayerSHA256[layer] != layerDigest {
				return false
			}
		}
	}
	return true
}

// safeMemberName reports whether name is a plain base file name without ever
// touching the filesystem.
func safeMemberName(name string) bool {
	return name != "" && !strings.Contains(name, `\`) && name == filepath.Base(name) && !filepath.IsAbs(name) && name != "." && name != ".."
}

// safeMember resolves a fixed evidence member without following symlinks.
func safeMember(base, name string) (string, error) {
	if !safeMemberName(name) {
		return "", errors.New("unsafe member")
	}
	member := filepath.Join(base, name)
	info, err := os.Lstat(member)
	if err != nil || !info.Mode().IsRegular() {
		return "", errors.New("member is unavailable or not regular")
	}
	return member, nil
}

// boundMember resolves one manifest-declared evidence member inside base,
// rejecting symlinks, non-regular files, and any path escape, then verifies
// its declared SHA-256.
func boundMember(base string, member FileDigest) (string, error) {
	if !safeMemberName(member.File) {
		return "", errors.New("unsafe member")
	}
	path := filepath.Join(base, member.File)
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() {
		return "", errors.New("member is unavailable or not regular")
	}
	digest, err := hashRegularFile(path, maxMemberBytes)
	if err != nil || digest != member.SHA256 {
		return "", errors.New("member digest mismatch")
	}
	return path, nil
}

func hashRegularFile(filePath string, limit int64) (string, error) {
	info, err := os.Lstat(filePath)
	if err != nil || !info.Mode().IsRegular() || info.Size() < 0 || info.Size() > limit {
		return "", errors.New("file is unavailable, unsafe, or oversized")
	}
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()
	hash := sha256.New()
	written, err := io.Copy(hash, io.LimitReader(file, limit+1))
	if err != nil || written != info.Size() || written > limit {
		return "", errors.New("file read failed or exceeded limit")
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func verifyChecksum(path, digest, name string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 0, maxChecksumLineBytes), maxChecksumLineBytes)
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

func formatViolations(report Report) string {
	if report.Valid {
		return fmt.Sprintf("candidate evidence valid for %s at %s (%d binaries, %d OCI targets)", report.Tag, report.Commit, report.BinaryTargets, report.OCITargets)
	}
	return strings.Join(report.Violations, "; ")
}
