package crossrunevidence

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	candidate "truerepublic/candidateevidence"
)

var (
	digestRE         = regexp.MustCompile(`^[0-9a-f]{64}$`)
	commitRE         = regexp.MustCompile(candidate.CommitPattern)
	tagRE            = regexp.MustCompile(candidate.TagPattern)
	runIDRE          = regexp.MustCompile(RunIDPattern)
	artifactDigestRE = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
)

const maxMemberBytes int64 = MaxJSONBytes

// expectedContractSource pins the exact comparison source identity. The
// contract may declare it, but only these values are accepted.
func expectedContractSource() ContractSource {
	return ContractSource{
		Repository:    "NeaBouli/TrueRepublic",
		WorkflowPath:  ".github/workflows/reproducible-daemon.yml",
		Branch:        "main",
		Event:         "workflow_dispatch",
		RunIDPattern:  RunIDPattern,
		RetentionDays: RetentionDays,
	}
}

type binaryTargetExpectation struct {
	ID       string
	Artifact string
}

func expectedBinaryTargets() []binaryTargetExpectation {
	return []binaryTargetExpectation{
		{ID: "linux-amd64", Artifact: "truerepublicd-linux-amd64"},
		{ID: "linux-arm64", Artifact: "truerepublicd-linux-arm64"},
	}
}

func expectedOCITargets() []struct{ Platform, ID string } {
	return []struct{ Platform, ID string }{
		{Platform: "linux/amd64", ID: "daemon-linux-amd64"},
		{Platform: "linux/amd64", ID: "client-web-linux-amd64"},
		{Platform: "linux/arm64", ID: "daemon-linux-arm64"},
		{Platform: "linux/arm64", ID: "client-web-linux-arm64"},
	}
}

func expectedMemberNames() [6]string {
	return [6]string{
		"receipt-baseline.json",
		"receipt-current.json",
		"candidate-manifest-baseline.json",
		"candidate-report-baseline.json",
		"candidate-manifest-current.json",
		"candidate-report-current.json",
	}
}

// Compare fail-closed verifies one cross-run comparison evidence directory
// against the exact cross-run contract, the pinned GH-261 candidate contract,
// and the explicit expected comparison identity. It rejects a symlinked
// evidence root and symlinked members, never leaves the evidence directory,
// and never reads payload content beyond the bound JSON members.
func Compare(evidenceDir, contractPath, candidateContractPath string, expected Expected) Report {
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

	manifestPath, err := safeMember(base, "cross-run-evidence.json")
	if err != nil {
		fail("cross-run manifest is unsafe")
		return report
	}
	var manifest Manifest
	if err := parseFile(manifestPath, &manifest); err != nil {
		fail("cross-run manifest is invalid")
		return report
	}
	var contract Contract
	if err := parseFile(contractPath, &contract); err != nil {
		fail("cross-run contract is invalid")
		return report
	}
	contractHash, err := hashRegularFile(contractPath, MaxJSONBytes)
	if err != nil {
		fail("cross-run contract digest is unavailable")
		return report
	}
	candidateContractHash, err := hashRegularFile(candidateContractPath, MaxJSONBytes)
	if err != nil {
		fail("candidate contract digest is unavailable")
		return report
	}

	report.Repository = manifest.Comparison.Repository
	report.WorkflowPath = manifest.Comparison.WorkflowPath
	report.Branch = manifest.Comparison.Branch
	report.Commit = manifest.Comparison.Commit
	report.BaselineRunID = manifest.Comparison.BaselineRunID
	report.CurrentRunID = manifest.Comparison.CurrentRunID

	if manifest.Schema != Schema {
		fail("cross-run evidence schema mismatch")
	}
	if err := validateContract(contract); err != nil {
		fail("cross-run contract mismatch: " + err.Error())
		return report
	}
	if contract.CandidateContractSHA256 != candidateContractHash {
		fail("pinned candidate contract digest mismatch")
		return report
	}
	if manifest.ContractSHA256 != contractHash {
		fail("cross-run contract digest mismatch")
	}
	if !manifest.Claims.explicitFalse() {
		fail("cross-run status claims must remain false")
	}
	validateComparisonIdentity(fail, manifest.Comparison, contract, expected)

	seenFiles := map[string]bool{"cross-run-evidence.json": true}
	baseline := verifyRunBinding(fail, base, "baseline", manifest.Baseline, expectedMemberNames()[0], expectedMemberNames()[2], expectedMemberNames()[3], seenFiles)
	current := verifyRunBinding(fail, base, "current", manifest.Current, expectedMemberNames()[1], expectedMemberNames()[4], expectedMemberNames()[5], seenFiles)

	if baseline != nil && current != nil {
		compareRuns(fail, &report, manifest.Comparison, contract, candidateContractHash, baseline, current)
	}
	verifyExactMembers(fail, base, seenFiles)

	report.Valid = len(report.Violations) == 0
	return report
}

// runEvidence carries one side's re-parsed receipt and candidate evidence.
type runEvidence struct {
	receipt  Receipt
	manifest candidate.Manifest
	report   candidate.Report
}

// validateComparisonIdentity binds the declared comparison to the contract
// and to the explicit expected values passed in by the caller.
func validateComparisonIdentity(fail func(string), comparison ComparisonIdentity, contract Contract, expected Expected) {
	want := contract.Source
	if comparison.Repository != want.Repository || comparison.WorkflowPath != want.WorkflowPath || comparison.Branch != want.Branch {
		fail("comparison source identity mismatch")
	}
	if !commitRE.MatchString(comparison.Commit) {
		fail("comparison commit is malformed")
	}
	if !runIDRE.MatchString(comparison.BaselineRunID) || !runIDRE.MatchString(comparison.CurrentRunID) {
		fail("comparison run ID is malformed")
	}
	if comparison.BaselineRunID == comparison.CurrentRunID {
		fail("comparison runs must be distinct")
	}
	got := Expected(comparison)
	if got != expected {
		fail("comparison identity does not match the expected values")
	}
}

// verifyRunBinding binds one run's three members by exact base name and
// SHA-256, then re-parses the receipt, candidate manifest, and candidate
// report. It returns nil when any member cannot be bound or parsed.
func verifyRunBinding(fail func(string), base, side string, binding RunBinding, receiptName, manifestName, reportName string, seenFiles map[string]bool) *runEvidence {
	members := []struct {
		kind string
		want string
		file candidate.FileDigest
	}{
		{kind: "receipt", want: receiptName, file: binding.Receipt},
		{kind: "candidate manifest", want: manifestName, file: binding.CandidateManifest},
		{kind: "candidate report", want: reportName, file: binding.CandidateReport},
	}
	paths := map[string]string{}
	ok := true
	for _, member := range members {
		if member.file.File != member.want {
			fail(side + " " + member.kind + " member name mismatch")
			ok = false
			continue
		}
		if !digestRE.MatchString(member.file.SHA256) {
			fail(side + " " + member.kind + " member digest is malformed")
			ok = false
			continue
		}
		if seenFiles[member.file.File] {
			fail(side + " " + member.kind + " member is duplicated")
			ok = false
			continue
		}
		seenFiles[member.file.File] = true
		path, err := boundMember(base, member.file)
		if err != nil {
			fail(side + " " + member.kind + " member is unsafe or drifted")
			ok = false
			continue
		}
		paths[member.kind] = path
	}
	if !ok {
		return nil
	}
	evidence := &runEvidence{}
	if err := parseFile(paths["receipt"], &evidence.receipt); err != nil {
		fail(side + " receipt is invalid")
		return nil
	}
	if err := parseFile(paths["candidate manifest"], &evidence.manifest); err != nil {
		fail(side + " candidate manifest is invalid")
		return nil
	}
	if err := parseFile(paths["candidate report"], &evidence.report); err != nil {
		fail(side + " candidate report is invalid")
		return nil
	}
	return evidence
}

// compareRuns validates both runs against the contract and expected identity
// and compares the complete re-parsed binary and OCI identities.
func compareRuns(fail func(string), report *Report, comparison ComparisonIdentity, contract Contract, candidateContractHash string, baseline, current *runEvidence) {
	validateReceipt(fail, "baseline", baseline.receipt, contract, comparison, comparison.BaselineRunID, "completed", "success")
	validateReceipt(fail, "current", current.receipt, contract, comparison, comparison.CurrentRunID, "in_progress", "")
	validateCandidateSide(fail, "baseline", candidateContractHash, comparison.Commit, baseline)
	validateCandidateSide(fail, "current", candidateContractHash, comparison.Commit, current)

	baselineCreated, baselineErr := time.Parse(time.RFC3339, baseline.receipt.CreatedAt)
	currentCreated, currentErr := time.Parse(time.RFC3339, current.receipt.CreatedAt)
	if baselineErr == nil && currentErr == nil {
		if currentCreated.Before(baselineCreated) {
			fail("current run predates the baseline run")
		} else if currentCreated.Sub(baselineCreated) > retentionSeconds*time.Second {
			fail("baseline run is outside the retention window")
		}
	}

	if baseline.manifest.Source.Tag != current.manifest.Source.Tag {
		fail("candidate tags differ between runs")
	}
	report.Tag = current.manifest.Source.Tag
	report.BinaryTargets = len(current.manifest.BinaryTargets)
	report.OCITargets = countOCITargets(current.manifest)
	for index := range current.manifest.BinaryTargets {
		if index >= len(baseline.manifest.BinaryTargets) {
			break
		}
		if baseline.manifest.BinaryTargets[index].ArtifactSHA256 != current.manifest.BinaryTargets[index].ArtifactSHA256 {
			fail("binary artifact digest differs between runs for " + current.manifest.BinaryTargets[index].ID)
		}
	}
	compareOCIIdentities(fail, baseline.manifest, current.manifest)
}

// validateReceipt treats receipt data as untrusted and binds it to the
// contract, the comparison identity, and the retention policy.
func validateReceipt(fail func(string), side string, receipt Receipt, contract Contract, comparison ComparisonIdentity, wantRunID, wantStatus, wantConclusion string) {
	prefix := side + " receipt "
	if receipt.Schema != ReceiptSchema {
		fail(prefix + "schema mismatch")
	}
	if receipt.Repository != contract.Source.Repository || receipt.WorkflowPath != contract.Source.WorkflowPath ||
		receipt.Branch != contract.Source.Branch || receipt.Event != contract.Source.Event {
		fail(prefix + "source identity mismatch")
	}
	if receipt.RunID != wantRunID {
		fail(prefix + "run ID mismatch")
	}
	attempt, err := strconv.ParseInt(receipt.RunAttempt.String(), 10, 64)
	if err != nil || attempt < 1 || attempt > maxRunAttempt {
		fail(prefix + "run attempt is malformed")
	}
	if receipt.HeadSHA != comparison.Commit {
		fail(prefix + "head commit mismatch")
	}
	if _, err := time.Parse(time.RFC3339, receipt.CreatedAt); err != nil {
		fail(prefix + "creation time is malformed")
	}
	if receipt.Status != wantStatus || receipt.Conclusion != wantConclusion {
		fail(prefix + "status or conclusion mismatch")
	}
	validateArtifact(fail, side, receipt.Artifact, comparison.Commit)
}

func validateArtifact(fail func(string), side string, artifact ArtifactBinding, commit string) {
	prefix := side + " artifact "
	if artifact.Name != candidateArtifactPfx+commit {
		fail(prefix + "name mismatch")
	}
	id, err := strconv.ParseInt(artifact.ID.String(), 10, 64)
	if err != nil || id < 1 || id > maxArtifactID {
		fail(prefix + "ID is malformed")
	}
	if !artifactDigestRE.MatchString(artifact.Digest) {
		fail(prefix + "digest is malformed")
	}
	if artifact.Expired {
		fail(prefix + "is expired")
	}
	created, createdErr := time.Parse(time.RFC3339, artifact.CreatedAt)
	expires, expiresErr := time.Parse(time.RFC3339, artifact.ExpiresAt)
	if createdErr != nil || expiresErr != nil {
		fail(prefix + "timestamps are malformed")
		return
	}
	if expires.Sub(created) != retentionSeconds*time.Second {
		fail(prefix + "retention window mismatch")
	}
}

// validateCandidateSide re-validates one run's GH-261 candidate manifest and
// report against the pinned candidate contract and the exact commit. Declared
// summary vectors are checked for shape and order only; the pairwise digest
// comparison happens between the two re-parsed manifests.
func validateCandidateSide(fail func(string), side, candidateContractHash, commit string, evidence *runEvidence) {
	manifest := evidence.manifest
	prefix := side + " candidate manifest "
	if manifest.Schema != candidate.Schema {
		fail(prefix + "schema mismatch")
	}
	if manifest.ContractSHA256 != candidateContractHash {
		fail(prefix + "contract digest mismatch")
	}
	if !tagRE.MatchString(manifest.Source.Tag) {
		fail(prefix + "tag is malformed")
	}
	if manifest.Source.Commit != commit {
		fail(prefix + "commit mismatch")
	}
	if !candidateClaimsFalse(manifest.Claims) {
		fail(prefix + "status claims must remain false")
	}
	expectedBinary := expectedBinaryTargets()
	if len(manifest.BinaryTargets) != len(expectedBinary) {
		fail(prefix + "binary target count mismatch")
	} else {
		for index, want := range expectedBinary {
			got := manifest.BinaryTargets[index]
			if got.ID != want.ID || got.Artifact != want.Artifact || !digestRE.MatchString(got.ArtifactSHA256) ||
				!digestRE.MatchString(got.Metadata.SHA256) || !digestRE.MatchString(got.Checksums.SHA256) {
				fail(prefix + "binary target mismatch for " + want.ID)
			}
		}
	}
	expectedOCI := expectedOCITargets()
	if len(manifest.OCIPlatforms) != 2 {
		fail(prefix + "OCI platform count mismatch")
	} else {
		for platformIndex, platform := range []string{"linux/amd64", "linux/arm64"} {
			binding := manifest.OCIPlatforms[platformIndex]
			if binding.Platform != platform || len(binding.Targets) != 2 ||
				!digestRE.MatchString(binding.Evidence.SHA256) || !digestRE.MatchString(binding.Report.SHA256) {
				fail(prefix + "OCI platform mismatch for " + platform)
				continue
			}
			for targetIndex := range binding.Targets {
				want := expectedOCI[platformIndex*2+targetIndex]
				identity := binding.Targets[targetIndex]
				if identity.ID != want.ID || !digestRE.MatchString(identity.IndexSHA256) ||
					!digestRE.MatchString(identity.ManifestSHA256) || !digestRE.MatchString(identity.ConfigSHA256) ||
					len(identity.LayerSHA256) == 0 || len(identity.LayerSHA256) > maxOCILayers {
					fail(prefix + "OCI target mismatch for " + want.ID)
					continue
				}
				for _, layer := range identity.LayerSHA256 {
					if !digestRE.MatchString(layer) {
						fail(prefix + "OCI layer digest is malformed for " + want.ID)
						break
					}
				}
			}
		}
	}

	candidateReport := evidence.report
	reportPrefix := side + " candidate report "
	if candidateReport.Schema != candidate.ReportSchema || !candidateReport.Valid || len(candidateReport.Violations) != 0 {
		fail(reportPrefix + "is not a clean verification report")
	}
	if candidateReport.Tag != manifest.Source.Tag || candidateReport.Commit != commit {
		fail(reportPrefix + "identity mismatch")
	}
	if candidateReport.BinaryTargets != len(manifest.BinaryTargets) || candidateReport.OCITargets != countOCITargets(manifest) {
		fail(reportPrefix + "target count mismatch")
	}
}

func candidateClaimsFalse(claims candidate.Claims) bool {
	return !claims.RealTagCreated && !claims.RefPushed && !claims.Signed &&
		!claims.Published && !claims.Deployed && !claims.Production
}

func countOCITargets(manifest candidate.Manifest) int {
	total := 0
	for _, platform := range manifest.OCIPlatforms {
		total += len(platform.Targets)
	}
	return total
}

// compareOCIIdentities compares the complete ordered OCI index, manifest,
// config, and ordered-layer vectors of the two re-parsed candidate manifests.
func compareOCIIdentities(fail func(string), baseline, current candidate.Manifest) {
	if len(baseline.OCIPlatforms) != len(current.OCIPlatforms) {
		fail("OCI platform sets differ between runs")
		return
	}
	for platformIndex := range current.OCIPlatforms {
		currentPlatform := current.OCIPlatforms[platformIndex]
		baselinePlatform := baseline.OCIPlatforms[platformIndex]
		if baselinePlatform.Platform != currentPlatform.Platform {
			fail("OCI platform order differs between runs")
			continue
		}
		if len(baselinePlatform.Targets) != len(currentPlatform.Targets) {
			fail("OCI target sets differ between runs for " + currentPlatform.Platform)
			continue
		}
		for targetIndex := range currentPlatform.Targets {
			baselineIdentity := baselinePlatform.Targets[targetIndex]
			currentIdentity := currentPlatform.Targets[targetIndex]
			if baselineIdentity.ID != currentIdentity.ID {
				fail("OCI target order differs between runs for " + currentPlatform.Platform)
				continue
			}
			if baselineIdentity.IndexSHA256 != currentIdentity.IndexSHA256 ||
				baselineIdentity.ManifestSHA256 != currentIdentity.ManifestSHA256 ||
				baselineIdentity.ConfigSHA256 != currentIdentity.ConfigSHA256 ||
				!stringSlicesEqual(baselineIdentity.LayerSHA256, currentIdentity.LayerSHA256) {
				fail("OCI identity differs between runs for " + currentIdentity.ID)
			}
		}
	}
}

func stringSlicesEqual(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func validateContract(contract Contract) error {
	if contract.Schema != ContractSchema || contract.Source != expectedContractSource() {
		return errors.New("schema or source identity")
	}
	if !digestRE.MatchString(contract.CandidateContractSHA256) {
		return errors.New("pinned candidate contract digest")
	}
	if _, err := regexp.Compile(contract.Source.RunIDPattern); err != nil {
		return errors.New("run ID grammar")
	}
	return nil
}

// verifyExactMembers keeps the evidence directory metadata-only and
// self-contained: every regular member must be declared exactly once by the
// cross-run manifest, and no binary, archive, symlink, directory, or unrelated
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
func boundMember(base string, member candidate.FileDigest) (string, error) {
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

func formatViolations(report Report) string {
	if report.Valid {
		return fmt.Sprintf("cross-run evidence valid for %s %s %s at %s (baseline run %s, current run %s; %d binaries, %d OCI targets match)",
			report.Repository, report.WorkflowPath, report.Branch, report.Commit, report.BaselineRunID, report.CurrentRunID, report.BinaryTargets, report.OCITargets)
	}
	return strings.Join(report.Violations, "; ")
}
