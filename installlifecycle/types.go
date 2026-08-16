package installlifecycle

const (
	ContractSchema    = "truerepublic.install-lifecycle/v1"
	ManifestSchema    = "truerepublic.install-manifest/v1"
	TransactionSchema = "truerepublic.install-transaction/v1"
	MaxJSONBytes      = 64 << 10
)

type Contract struct {
	Schema                string            `json:"schema"`
	SupportedTargets      []string          `json:"supported_targets"`
	TargetRuntimes        map[string]string `json:"target_runtimes"`
	Layout                Layout            `json:"layout"`
	Modes                 Modes             `json:"modes"`
	Limits                Limits            `json:"limits"`
	OperatorStateBoundary string            `json:"operator_state_boundary"`
	SourceRef             string            `json:"-"`
	Target                string            `json:"-"`
	Runtime               string            `json:"-"`
	ArtifactSHA256        string            `json:"-"`
	Prefix                string            `json:"-"`
	BinaryPath            string            `json:"-"`
	ManifestPath          string            `json:"-"`
	RollbackPath          string            `json:"-"`
	TransactionPath       string            `json:"-"`
	OperatorStatePath     string            `json:"-"`
}

type Layout struct {
	Binary         string `json:"binary"`
	Manifest       string `json:"manifest"`
	RollbackBinary string `json:"rollback_binary"`
	Transaction    string `json:"transaction"`
}
type Modes struct {
	Binary      uint32 `json:"binary"`
	Metadata    uint32 `json:"metadata"`
	Transaction uint32 `json:"transaction"`
}
type Limits struct {
	MaxArtifactBytes int64 `json:"max_artifact_bytes"`
}

type Identity struct {
	SHA256    string `json:"sha256"`
	SourceRef string `json:"source_ref"`
	Target    string `json:"target"`
	Runtime   string `json:"runtime"`
}

type Manifest struct {
	Schema             string    `json:"schema"`
	BinaryPath         string    `json:"binary_path"`
	OperatorStatePath  string    `json:"operator_state_path"`
	Current            Identity  `json:"current"`
	Rollback           *Identity `json:"rollback,omitempty"`
	Generation         uint64    `json:"generation"`
	RollbackGeneration *uint64   `json:"rollback_generation,omitempty"`
}

type transaction struct {
	Schema    string `json:"schema"`
	Operation string `json:"operation"`
}

type Status struct {
	Installed bool      `json:"installed"`
	Healthy   bool      `json:"healthy"`
	Manifest  *Manifest `json:"manifest,omitempty"`
	Problem   string    `json:"problem,omitempty"`
}
