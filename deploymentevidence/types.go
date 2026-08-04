// Package deploymentevidence verifies strict, secret-free, offline
// deployment-evidence manifests bound to an exact GH-89 topology contract
// supplied locally as raw bytes. It never probes, resolves, mutates, or
// deploys infrastructure: it performs no DNS lookups, no network access, no
// environment or home-directory reads, and no node or firewall changes. It
// only parses two explicit local files and checks the manifest against the
// derived topology digest and composition.
package deploymentevidence

const (
	// ManifestVersion is the only accepted deployment-evidence schema.
	ManifestVersion = "truerepublic.deployment-evidence/v1"
	// MaxManifestBytes bounds one manifest document.
	MaxManifestBytes = 256 * 1024
	// GateResultPassed is the only accepted gate result.
	GateResultPassed = "passed"
	// ApprovalRoleOperator and ApprovalRoleIndependentReviewer are the only
	// accepted approval roles, required exactly once each.
	ApprovalRoleOperator            = "operator"
	ApprovalRoleIndependentReviewer = "independent-reviewer"

	maxJSONDepth          = 32
	maxEvidenceAgeSec     = 30 * 24 * 60 * 60
	maxClockSkewSec       = 5 * 60
	strictTimestampLayout = "2006-01-02T15:04:05Z"
)

// GateIDs is the canonical, exact, ordered deployment gate set.
var GateIDs = []string{
	"role-policy",
	"provider-separation",
	"dns-tls",
	"provider-firewall",
	"host-firewall",
	"listener-exposure",
	"proxy-abuse-controls",
	"telemetry",
	"capacity",
	"incident-rehearsal",
	"two-person-review",
}

// Manifest is the strict deployment-evidence document. It intentionally has
// no host, IP, node ID, path, URL, provider, or free-text fields.
type Manifest struct {
	Version        string     `json:"version"`
	ChainID        string     `json:"chain_id"`
	TopologySHA256 string     `json:"topology_sha256"`
	NodeCount      int        `json:"node_count"`
	RoleCounts     RoleCounts `json:"role_counts"`
	PreparedBy     string     `json:"prepared_by"`
	PreparedAt     string     `json:"prepared_at"`
	Gates          []Gate     `json:"gates"`
	Approvals      []Approval `json:"approvals"`
}

// RoleCounts mirrors the four topology roles.
type RoleCounts struct {
	Seed      int `json:"seed"`
	Sentry    int `json:"sentry"`
	Validator int `json:"validator"`
	RPC       int `json:"rpc"`
}

// Gate is one deployment gate with its secret-free evidence digest.
type Gate struct {
	ID             string `json:"id"`
	Result         string `json:"result"`
	StartedAt      string `json:"started_at"`
	CompletedAt    string `json:"completed_at"`
	EvidenceSHA256 string `json:"evidence_sha256"`
}

// Approval is one two-person-review approval bound to the topology digest.
type Approval struct {
	Seat           string `json:"seat"`
	Role           string `json:"role"`
	ApprovedAt     string `json:"approved_at"`
	TopologySHA256 string `json:"topology_sha256"`
}

// Topology carries the derived, secret-free facts of the exact local
// topology contract a manifest is verified against.
type Topology struct {
	SHA256     string
	ChainID    string
	NodeCount  int
	RoleCounts RoleCounts
}

// Violation is one generic, fixed-label verification failure. Messages never
// reflect rejected values, field names, paths, timestamps, digests, or seats.
type Violation struct {
	Check   string `json:"check"`
	Message string `json:"message"`
}

// Report contains only fixed safe fields: no chain ID, paths, digests,
// times, seat names, or input values.
type Report struct {
	Version    string      `json:"version"`
	GateCount  int         `json:"gate_count"`
	NodeCount  int         `json:"node_count"`
	Valid      bool        `json:"valid"`
	Violations []Violation `json:"violations"`
}
