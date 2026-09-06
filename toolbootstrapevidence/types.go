// Package toolbootstrapevidence verifies deterministic, fail-closed
// composition evidence for the repository-owned locked CI/release tool
// bootstrap (GH-278). The evidence binds the exact configured tool versions
// from configs/security/gates.json, the tool-platform contract, the nested
// tools/ci Go module/go.sum and npm package/package-lock digests, and the
// digests of twice-built tool artifacts. Verification is offline and strict:
// unknown, duplicate, trailing, or missing JSON fields, path or symlink
// escape, version or digest drift, incomplete or duplicate tools, and any true
// signed/published/deployed/production/long-term-hermetic claim fail closed.
// Tool identity is evidence data, never authority: no tag, signature,
// publication, registry push, deployment, production, or rollout-credit claim
// is performed or accepted.
package toolbootstrapevidence

import (
	"errors"
)

const (
	ContractSchema = "truerepublic.release-tool-platform/v1"
	GatesSchema    = "truerepublic.security-gates/v1"
	Schema         = "truerepublic.tool-bootstrap-evidence/v1"
	ReportSchema   = "truerepublic.tool-bootstrap-report/v1"
	ManifestFile   = "tool-bootstrap-evidence.json"

	MaxJSONBytes     = 1 << 20
	MaxLockBytes     = 16 << 20
	MaxArtifactBytes = 512 << 20
	maxJSONDepth     = 32
	maxTools         = 16
)

// Contract is the strictly parsed tool-platform contract. Only the bootstrap
// composition is interpreted; the legacy pin/platform/base-image sections are
// validated for shape so drift cannot hide in unparsed fields.
type Contract struct {
	Schema     string             `json:"schema"`
	Tools      map[string]string  `json:"tools"`
	Platforms  []ContractPlatform `json:"platforms"`
	BaseImages []string           `json:"base_images"`
	Bootstrap  Bootstrap          `json:"bootstrap"`
}

type ContractPlatform struct {
	ID     string `json:"id"`
	Runner string `json:"runner"`
	Arch   string `json:"arch"`
}

// Bootstrap declares the repository-owned lock files and the exact allowlist
// of buildable CI/release tools.
type Bootstrap struct {
	GoModule   string          `json:"go_module"`
	GoSum      string          `json:"go_sum"`
	NpmPackage string          `json:"npm_package"`
	NpmLock    string          `json:"npm_lock"`
	Tools      []BootstrapTool `json:"tools"`
}

// BootstrapTool binds one tool id to its build source, the security-gates
// version key, and the artifact path (relative to the tool's output
// directory) whose digest the evidence binds.
type BootstrapTool struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Package  string `json:"package"`
	Module   string `json:"module,omitempty"`
	GatesKey string `json:"gates_key"`
	Artifact string `json:"artifact"`
}

// Gates is the strictly parsed security gate contract section the evidence
// binds tool versions to.
type Gates struct {
	Version                   string            `json:"version"`
	ReviewCadenceDays         int               `json:"review_cadence_days"`
	ExceptionMaxDays          int               `json:"exception_max_days"`
	Toolchains                map[string]string `json:"toolchains"`
	Tools                     map[string]string `json:"tools"`
	Actions                   map[string]string `json:"actions"`
	GoVulnerabilityExceptions []GatesException  `json:"go_vulnerability_exceptions"`
}

type GatesException struct {
	ID         string `json:"id"`
	ApprovedOn string `json:"approved_on"`
	Expires    string `json:"expires"`
	Reason     string `json:"reason"`
}

// Claims are the tool-bootstrap status claims. Every claim must be explicitly
// present and false.
type Claims struct {
	Signed           bool `json:"signed"`
	Published        bool `json:"published"`
	Deployed         bool `json:"deployed"`
	Production       bool `json:"production"`
	LongTermHermetic bool `json:"long_term_hermetic"`
	present          bool
}

func (c *Claims) UnmarshalJSON(data []byte) error {
	var raw struct {
		Signed           *bool `json:"signed"`
		Published        *bool `json:"published"`
		Deployed         *bool `json:"deployed"`
		Production       *bool `json:"production"`
		LongTermHermetic *bool `json:"long_term_hermetic"`
	}
	if err := parseBytes(data, &raw); err != nil {
		return err
	}
	if raw.Signed == nil || raw.Published == nil || raw.Deployed == nil || raw.Production == nil || raw.LongTermHermetic == nil {
		return errors.New("tool-bootstrap status claims are incomplete")
	}
	c.Signed, c.Published, c.Deployed, c.Production, c.LongTermHermetic, c.present =
		*raw.Signed, *raw.Published, *raw.Deployed, *raw.Production, *raw.LongTermHermetic, true
	return nil
}

func (c Claims) explicitFalse() bool {
	return c.present && !c.Signed && !c.Published && !c.Deployed && !c.Production && !c.LongTermHermetic
}

// FileDigest binds one evidence member by repository-relative file path and
// SHA-256.
type FileDigest struct {
	File   string `json:"file"`
	SHA256 string `json:"sha256"`
}

// Locks binds the digests of the four repository-owned tool lock files.
type Locks struct {
	GoModule   FileDigest `json:"go_module"`
	GoSum      FileDigest `json:"go_sum"`
	NpmPackage FileDigest `json:"npm_package"`
	NpmLock    FileDigest `json:"npm_lock"`
}

// ToolEvidence binds one built tool to its exact configured version and
// twice-built artifact digest.
type ToolEvidence struct {
	ID       string     `json:"id"`
	Kind     string     `json:"kind"`
	Version  string     `json:"version"`
	Artifact FileDigest `json:"artifact"`
}

// Manifest is the digest-only tool-bootstrap composition evidence manifest.
type Manifest struct {
	Schema         string         `json:"schema"`
	ContractSHA256 string         `json:"contract_sha256"`
	GatesSHA256    string         `json:"gates_sha256"`
	Claims         Claims         `json:"claims"`
	Locks          Locks          `json:"locks"`
	Tools          []ToolEvidence `json:"tools"`
}

// Report is the deterministic verification result.
type Report struct {
	Schema     string   `json:"schema"`
	Valid      bool     `json:"valid"`
	Tools      int      `json:"tools"`
	Violations []string `json:"violations"`
}
