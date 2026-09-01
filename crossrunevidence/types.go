// Package crossrunevidence verifies a strict metadata/digest-only comparison
// of two distinct GitHub Actions executions of the same exact main commit,
// extending the GH-258/GH-261 candidate evidence. The comparison binds run
// receipts and re-parses both GH-261 candidate manifests and reports instead
// of trusting copied summary vectors. Run identity is evidence data, never
// authority: no real tag, ref push, signature, attestation, publication,
// deployment, production, key/fund, rollout-credit, or long-term-hermetic
// claim is performed or accepted.
package crossrunevidence

import (
	"encoding/json"
	"errors"

	candidate "truerepublic/candidateevidence"
)

const (
	ContractSchema       = "truerepublic.cross-run-rebuild/v1"
	Schema               = "truerepublic.cross-run-evidence/v1"
	ReceiptSchema        = "truerepublic.cross-run-receipt/v1"
	ReportSchema         = "truerepublic.cross-run-report/v1"
	RunIDPattern         = `^[1-9][0-9]{0,18}$`
	RetentionDays        = 14
	MaxJSONBytes         = candidate.MaxJSONBytes
	maxJSONDepth         = 32
	maxOCILayers         = 128
	maxRunAttempt        = 10000
	maxArtifactID        = 1 << 53
	retentionSeconds     = RetentionDays * 24 * 60 * 60
	retentionTolerance   = 60
	candidateArtifactPfx = "truerepublic-candidate-"
)

// Contract is the exact versioned cross-run rebuild comparison contract.
type Contract struct {
	Schema                  string         `json:"schema"`
	CandidateContractSHA256 string         `json:"candidate_contract_sha256"`
	Source                  ContractSource `json:"source"`
}

type ContractSource struct {
	Repository    string `json:"repository"`
	WorkflowPath  string `json:"workflow_path"`
	Branch        string `json:"branch"`
	Event         string `json:"event"`
	RunIDPattern  string `json:"run_id_pattern"`
	RetentionDays int    `json:"retention_days"`
}

// Expected carries the Actions-computed comparison identity that the CLI
// requires the evidence to match exactly.
type Expected struct {
	Repository    string
	WorkflowPath  string
	Branch        string
	Commit        string
	BaselineRunID string
	CurrentRunID  string
}

// ComparisonIdentity is the evidence-declared identity of the compared pair.
type ComparisonIdentity struct {
	Repository    string `json:"repository"`
	WorkflowPath  string `json:"workflow_path"`
	Branch        string `json:"branch"`
	Commit        string `json:"commit"`
	BaselineRunID string `json:"baseline_run_id"`
	CurrentRunID  string `json:"current_run_id"`
}

// Claims are the cross-run status claims. Every claim must be explicitly
// present and false.
type Claims struct {
	RealTagCreated   bool `json:"real_tag_created"`
	RefPushed        bool `json:"ref_pushed"`
	Signed           bool `json:"signed"`
	Attested         bool `json:"attested"`
	Published        bool `json:"published"`
	Deployed         bool `json:"deployed"`
	Production       bool `json:"production"`
	LongTermHermetic bool `json:"long_term_hermetic"`
	present          bool
}

func (c *Claims) UnmarshalJSON(data []byte) error {
	var raw struct {
		RealTagCreated   *bool `json:"real_tag_created"`
		RefPushed        *bool `json:"ref_pushed"`
		Signed           *bool `json:"signed"`
		Attested         *bool `json:"attested"`
		Published        *bool `json:"published"`
		Deployed         *bool `json:"deployed"`
		Production       *bool `json:"production"`
		LongTermHermetic *bool `json:"long_term_hermetic"`
	}
	if err := parseBytes(data, &raw); err != nil {
		return err
	}
	if raw.RealTagCreated == nil || raw.RefPushed == nil || raw.Signed == nil || raw.Attested == nil ||
		raw.Published == nil || raw.Deployed == nil || raw.Production == nil || raw.LongTermHermetic == nil {
		return errors.New("cross-run status claims are incomplete")
	}
	c.RealTagCreated, c.RefPushed, c.Signed, c.Attested, c.Published, c.Deployed, c.Production, c.LongTermHermetic, c.present =
		*raw.RealTagCreated, *raw.RefPushed, *raw.Signed, *raw.Attested, *raw.Published, *raw.Deployed, *raw.Production, *raw.LongTermHermetic, true
	return nil
}

func (c Claims) explicitFalse() bool {
	return c.present && !c.RealTagCreated && !c.RefPushed && !c.Signed && !c.Attested &&
		!c.Published && !c.Deployed && !c.Production && !c.LongTermHermetic
}

// Manifest is the digest-only cross-run comparison evidence manifest.
type Manifest struct {
	Schema         string             `json:"schema"`
	ContractSHA256 string             `json:"contract_sha256"`
	Comparison     ComparisonIdentity `json:"comparison"`
	Claims         Claims             `json:"claims"`
	Baseline       RunBinding         `json:"baseline"`
	Current        RunBinding         `json:"current"`
}

// RunBinding binds one run's receipt and its GH-261 candidate manifest and
// report members by base file name and SHA-256. Payloads are never present.
type RunBinding struct {
	Receipt           candidate.FileDigest `json:"receipt"`
	CandidateManifest candidate.FileDigest `json:"candidate_manifest"`
	CandidateReport   candidate.FileDigest `json:"candidate_report"`
}

// Receipt is the untrusted metadata receipt of one workflow execution.
type Receipt struct {
	Schema       string          `json:"schema"`
	Repository   string          `json:"repository"`
	WorkflowPath string          `json:"workflow_path"`
	Branch       string          `json:"branch"`
	Event        string          `json:"event"`
	RunID        string          `json:"run_id"`
	RunAttempt   json.Number     `json:"run_attempt"`
	HeadSHA      string          `json:"head_sha"`
	CreatedAt    string          `json:"created_at"`
	Status       string          `json:"status"`
	Conclusion   string          `json:"conclusion"`
	Artifact     ArtifactBinding `json:"artifact"`
}

// ArtifactBinding binds the run's candidate evidence artifact metadata.
type ArtifactBinding struct {
	Name      string      `json:"name"`
	ID        json.Number `json:"id"`
	Digest    string      `json:"digest"`
	CreatedAt string      `json:"created_at"`
	ExpiresAt string      `json:"expires_at"`
	Expired   bool        `json:"expired"`
}

// Report is the deterministic comparison result.
type Report struct {
	Schema        string   `json:"schema"`
	Valid         bool     `json:"valid"`
	Repository    string   `json:"repository"`
	WorkflowPath  string   `json:"workflow_path"`
	Branch        string   `json:"branch"`
	Commit        string   `json:"commit"`
	Tag           string   `json:"tag"`
	BaselineRunID string   `json:"baseline_run_id"`
	CurrentRunID  string   `json:"current_run_id"`
	BinaryTargets int      `json:"binary_targets"`
	OCITargets    int      `json:"oci_targets"`
	Violations    []string `json:"violations"`
}
