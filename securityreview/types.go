package securityreview

const (
	ScopeSchemaVersion    = "truerepublic.security-review-scope/v1"
	FindingsSchemaVersion = "truerepublic.security-review-findings/v1"
	MaxDocumentBytes      = 1 << 20
	MaxReferencedBytes    = 4 << 20
)

type Identity struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Updated string `json:"updated"`
}

type Scope struct {
	Document                  Identity    `json:"document"`
	ExternalIndependenceClaim *bool       `json:"external_independence_claim"`
	ProductionReady           *bool       `json:"production_ready"`
	ThreatModelPath           string      `json:"threat_model_path"`
	SecurityGatePath          string      `json:"security_gate_path"`
	InScope                   []string    `json:"in_scope"`
	Invariants                []Invariant `json:"invariants"`
	Evidence                  []Evidence  `json:"evidence"`
}

type Invariant struct {
	ID           string   `json:"id"`
	Description  string   `json:"description"`
	SourcePath   string   `json:"source_path"`
	SourceAnchor string   `json:"source_anchor"`
	ThreatIDs    []string `json:"threat_ids"`
}

type Evidence struct {
	ID             string `json:"id"`
	Path           string `json:"path"`
	Kind           string `json:"kind"`
	RequiredAnchor string `json:"required_anchor"`
}

type Findings struct {
	Document                  Identity  `json:"document"`
	ExternalIndependenceClaim *bool     `json:"external_independence_claim"`
	ProductionReady           *bool     `json:"production_ready"`
	Findings                  []Finding `json:"findings"`
}

type Finding struct {
	ID                   string   `json:"id"`
	Severity             string   `json:"severity"`
	Status               string   `json:"status"`
	Source               string   `json:"source"`
	Title                string   `json:"title"`
	ThreatIDs            []string `json:"threat_ids"`
	AffectedPaths        []string `json:"affected_paths"`
	RemediationEvidence  []string `json:"remediation_evidence"`
	VerificationSource   string   `json:"verification_source"`
	VerificationEvidence []string `json:"verification_evidence"`
	Owner                string   `json:"owner"`
	AcceptedOn           string   `json:"accepted_on"`
	Expires              string   `json:"expires"`
}
