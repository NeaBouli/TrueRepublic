// Package candidateevidence verifies a strict digest-only release-candidate
// manifest that binds one explicitly simulated future tag string and one exact
// checked-out commit to the deterministic daemon target records and the
// repeated OCI target identities. Tag and commit are data, never authority:
// no real tag, ref push, signature, publication, deployment, or production
// action is performed or accepted.
package candidateevidence

import "errors"

const (
	ContractSchema       = "truerepublic.release-candidate/v1"
	Schema               = "truerepublic.release-candidate-evidence/v1"
	ReportSchema         = "truerepublic.release-candidate-report/v1"
	BuildContractSchema  = "truerepublic.daemon-build/v1"
	BuildEvidenceSchema  = "truerepublic.daemon-build-evidence/v1"
	OCIContractSchema    = "truerepublic.oci-build/v1"
	OCIEvidenceSchema    = "truerepublic.oci-evidence/v1"
	OCIReportSchema      = "truerepublic.oci-evidence-report/v1"
	TagPattern           = `^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`
	CommitPattern        = `^[0-9a-f]{40}$`
	MaxJSONBytes         = 1 << 20
	maxJSONDepth         = 32
	maxOCILayers         = 128
	maxChecksumLineBytes = 4096
)

// Contract is the exact versioned release-candidate contract.
type Contract struct {
	Schema               string                 `json:"schema"`
	Source               ContractSource         `json:"source"`
	BinaryContractSHA256 string                 `json:"binary_contract_sha256"`
	OCIContractSHA256    string                 `json:"oci_contract_sha256"`
	BinaryTargets        []ContractBinaryTarget `json:"binary_targets"`
	OCIRepetitions       int                    `json:"oci_repetitions"`
	OCITargets           []ContractOCITarget    `json:"oci_targets"`
}

type ContractSource struct {
	CommitKind    string `json:"commit_kind"`
	CommitPattern string `json:"commit_pattern"`
	TagKind       string `json:"tag_kind"`
	TagPattern    string `json:"tag_pattern"`
}

type ContractBinaryTarget struct {
	ID         string `json:"id"`
	Artifact   string `json:"artifact"`
	CIRunner   string `json:"ci_runner"`
	RunnerArch string `json:"runner_arch"`
}

type ContractOCITarget struct {
	ID       string `json:"id"`
	Platform string `json:"platform"`
}

// Manifest is the digest-only release-candidate evidence manifest.
type Manifest struct {
	Schema         string                `json:"schema"`
	ContractSHA256 string                `json:"contract_sha256"`
	Source         SourceIdentity        `json:"source"`
	Claims         Claims                `json:"claims"`
	BinaryTargets  []BinaryTargetBinding `json:"binary_targets"`
	OCIPlatforms   []OCIPlatformBinding  `json:"oci_platforms"`
}

// SourceIdentity couples the simulated future tag string with the exact
// checked-out commit. Both are evidence data; neither grants any authority.
type SourceIdentity struct {
	Tag    string `json:"tag"`
	Commit string `json:"commit"`
}

// Claims are the release-candidate status claims. Every claim must be
// explicitly present and false.
type Claims struct {
	RealTagCreated bool `json:"real_tag_created"`
	RefPushed      bool `json:"ref_pushed"`
	Signed         bool `json:"signed"`
	Published      bool `json:"published"`
	Deployed       bool `json:"deployed"`
	Production     bool `json:"production"`
	present        bool
}

func (c *Claims) UnmarshalJSON(data []byte) error {
	var raw struct {
		RealTagCreated *bool `json:"real_tag_created"`
		RefPushed      *bool `json:"ref_pushed"`
		Signed         *bool `json:"signed"`
		Published      *bool `json:"published"`
		Deployed       *bool `json:"deployed"`
		Production     *bool `json:"production"`
	}
	if err := parseBytes(data, &raw); err != nil {
		return err
	}
	if raw.RealTagCreated == nil || raw.RefPushed == nil || raw.Signed == nil || raw.Published == nil || raw.Deployed == nil || raw.Production == nil {
		return errors.New("candidate status claims are incomplete")
	}
	c.RealTagCreated, c.RefPushed, c.Signed, c.Published, c.Deployed, c.Production, c.present =
		*raw.RealTagCreated, *raw.RefPushed, *raw.Signed, *raw.Published, *raw.Deployed, *raw.Production, true
	return nil
}

func (c Claims) explicitFalse() bool {
	return c.present && !c.RealTagCreated && !c.RefPushed && !c.Signed && !c.Published && !c.Deployed && !c.Production
}

// FileDigest binds one evidence member by base file name and SHA-256.
type FileDigest struct {
	File   string `json:"file"`
	SHA256 string `json:"sha256"`
}

// BinaryTargetBinding cross-binds the deterministic daemon metadata and
// checksum evidence for one maintained binary target. The artifact payload
// itself is never present, copied, or emitted.
type BinaryTargetBinding struct {
	ID             string     `json:"id"`
	Artifact       string     `json:"artifact"`
	ArtifactSHA256 string     `json:"artifact_sha256"`
	Metadata       FileDigest `json:"metadata"`
	Checksums      FileDigest `json:"checksums"`
}

// OCIPlatformBinding cross-binds the per-platform repeated OCI evidence and
// digest-report JSON plus the normalized per-target image identities.
type OCIPlatformBinding struct {
	Platform string             `json:"platform"`
	Evidence FileDigest         `json:"evidence"`
	Report   FileDigest         `json:"report"`
	Targets  []OCIImageIdentity `json:"targets"`
}

// OCIImageIdentity is the digest-only identity of one repeated OCI target.
type OCIImageIdentity struct {
	ID             string   `json:"id"`
	IndexSHA256    string   `json:"index_sha256"`
	ManifestSHA256 string   `json:"manifest_sha256"`
	ConfigSHA256   string   `json:"config_sha256"`
	LayerSHA256    []string `json:"layer_sha256"`
}

// Report is the deterministic verification result.
type Report struct {
	Schema        string   `json:"schema"`
	Valid         bool     `json:"valid"`
	Tag           string   `json:"tag"`
	Commit        string   `json:"commit"`
	BinaryTargets int      `json:"binary_targets"`
	OCITargets    int      `json:"oci_targets"`
	Violations    []string `json:"violations"`
}
