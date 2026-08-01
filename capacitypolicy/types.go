// Package capacitypolicy validates secret-free capacity qualification plans
// and evidence. It never starts nodes, reads node homes, or performs load.
package capacitypolicy

const (
	ContractVersion              = "truerepublic.capacity-qualification/v1"
	EvidenceVersion              = "truerepublic.capacity-evidence/v1"
	MaintainedQualificationID    = "phase-6-capacity-qualification"
	MaintainedEnvironment        = "synthetic-loopback"
	MaintainedValidatorCount     = 4
	MaintainedTransactionCount   = 96
	MaintainedTransactionWaves   = 24
	MaintainedMinimumBlockDelta  = 24
	MaintainedProjectionBlocks   = 1_000_000
	MaintainedSnapshotInterval   = 2
	MaintainedSnapshotKeepRecent = 3
	MaintainedTelemetryRetention = 60
	MaintainedLogMaxSizeBytes    = 50 * 1024 * 1024
	MaintainedLogMaxFiles        = 3
	MaxDocumentBytes             = 256 * 1024
)

var MaintainedMetrics = []string{
	"process_resident_memory_bytes",
	"go_goroutines",
	"cometbft_consensus_height",
	"truerepublic_app_last_successful_block_height",
	"truerepublic_app_last_successful_invariant_cycle_height",
	"truerepublic_app_completed_blocks_total",
	"truerepublic_token_pnyx_supply_base_units",
	"truerepublic_token_pnyx_supply_headroom_base_units",
}

var MaintainedAbortConditions = []string{
	"transaction-failure",
	"consensus-progress-stalled",
	"app-hash-divergence",
	"validator-power-divergence",
	"ledger-invariant-failure",
	"resource-envelope-breach",
	"retention-unbounded",
	"measurement-overflow",
	"secret-material-detected",
}

type Contract struct {
	Version         string          `json:"version"`
	QualificationID string          `json:"qualification_id"`
	Environment     string          `json:"environment"`
	Workload        Workload        `json:"workload"`
	Storage         StoragePolicy   `json:"storage"`
	Telemetry       TelemetryPolicy `json:"telemetry"`
	Logging         LoggingPolicy   `json:"logging"`
	Limits          Limits          `json:"limits"`
	AbortConditions []string        `json:"abort_conditions"`
}

type Workload struct {
	ValidatorCount         int   `json:"validator_count"`
	TransactionCount       int   `json:"transaction_count"`
	MinimumBlockDelta      int64 `json:"minimum_block_delta"`
	MaximumDurationSeconds int64 `json:"maximum_duration_seconds"`
	MaximumCommitLatencyMS int64 `json:"maximum_commit_latency_ms"`
}

type StoragePolicy struct {
	Pruning                string `json:"pruning"`
	SnapshotInterval       int64  `json:"snapshot_interval"`
	SnapshotKeepRecent     int    `json:"snapshot_keep_recent"`
	ProjectionBlocks       int64  `json:"projection_blocks"`
	MaxNodeDataGrowthBytes int64  `json:"max_node_data_growth_bytes"`
}

type TelemetryPolicy struct {
	Enabled                    bool     `json:"enabled"`
	Bind                       string   `json:"bind"`
	PrometheusRetentionSeconds int      `json:"prometheus_retention_seconds"`
	RequiredMetrics            []string `json:"required_metrics"`
}

type LoggingPolicy struct {
	Driver       string `json:"driver"`
	MaxSizeBytes int64  `json:"max_size_bytes"`
	MaxFiles     int    `json:"max_files"`
}

type Limits struct {
	MaxLogGrowthBytesPerNode int64 `json:"max_log_growth_bytes_per_node"`
	MaxRSSBytesPerNode       int64 `json:"max_rss_bytes_per_node"`
	MaxGoroutines            int64 `json:"max_goroutines"`
}

type Evidence struct {
	Version         string             `json:"version"`
	QualificationID string             `json:"qualification_id"`
	Environment     string             `json:"environment"`
	Workload        WorkloadEvidence   `json:"workload"`
	Nodes           []NodeEvidence     `json:"nodes"`
	Metrics         MetricsEvidence    `json:"metrics"`
	Retention       RetentionEvidence  `json:"retention"`
	Consensus       ConsensusEvidence  `json:"consensus"`
	Projection      ProjectionEvidence `json:"projection"`
}

type WorkloadEvidence struct {
	Validators            int   `json:"validators"`
	TransactionsSubmitted int   `json:"transactions_submitted"`
	TransactionsCommitted int   `json:"transactions_committed"`
	TransactionsFailed    int   `json:"transactions_failed"`
	StartHeight           int64 `json:"start_height"`
	EndHeight             int64 `json:"end_height"`
	BlockDelta            int64 `json:"block_delta"`
	DurationMS            int64 `json:"duration_ms"`
	ThroughputMilliTPS    int64 `json:"throughput_milli_tps"`
	AverageBlockTimeMS    int64 `json:"average_block_time_ms"`
	P95CommitLatencyMS    int64 `json:"p95_commit_latency_ms"`
	MaxCommitLatencyMS    int64 `json:"max_commit_latency_ms"`
}

type NodeEvidence struct {
	Name            string `json:"name"`
	DataBytesStart  int64  `json:"data_bytes_start"`
	DataBytesPeak   int64  `json:"data_bytes_peak"`
	DataGrowthBytes int64  `json:"data_growth_bytes"`
	LogBytesStart   int64  `json:"log_bytes_start"`
	LogBytesPeak    int64  `json:"log_bytes_peak"`
	LogGrowthBytes  int64  `json:"log_growth_bytes"`
	MaxRSSBytes     int64  `json:"max_rss_bytes"`
}

type MetricsEvidence struct {
	ConsensusHeight     int64 `json:"consensus_height"`
	ApplicationHeight   int64 `json:"application_height"`
	InvariantHeight     int64 `json:"invariant_height"`
	CompletedBlocks     int64 `json:"completed_blocks"`
	SupplyBaseUnits     int64 `json:"supply_base_units"`
	SupplyHeadroomUnits int64 `json:"supply_headroom_base_units"`
	Goroutines          int64 `json:"goroutines"`
	ResidentMemoryBytes int64 `json:"resident_memory_bytes"`
}

type RetentionEvidence struct {
	Pruning               string `json:"pruning"`
	SnapshotInterval      int64  `json:"snapshot_interval"`
	SnapshotKeepRecent    int    `json:"snapshot_keep_recent"`
	ObservedSnapshotCount int    `json:"observed_snapshot_count"`
}

type ConsensusEvidence struct {
	CommonAppHash            bool `json:"common_app_hash"`
	ValidatorPowerConsistent bool `json:"validator_power_consistent"`
	RestartVerified          bool `json:"restart_verified"`
	LedgerValid              bool `json:"ledger_valid"`
}

type ProjectionEvidence struct {
	Blocks                  int64 `json:"blocks"`
	MaxProjectedGrowthBytes int64 `json:"max_projected_growth_bytes"`
}

type Violation struct {
	Check   string `json:"check"`
	Message string `json:"message"`
}

type Report struct {
	Version          string      `json:"version"`
	QualificationID  string      `json:"qualification_id"`
	ValidatorCount   int         `json:"validator_count"`
	TransactionCount int         `json:"transaction_count"`
	Valid            bool        `json:"valid"`
	Violations       []Violation `json:"violations"`
}

type EvidenceReport struct {
	Version         string      `json:"version"`
	QualificationID string      `json:"qualification_id"`
	ValidatorCount  int         `json:"validator_count"`
	CommittedCount  int         `json:"committed_count"`
	Valid           bool        `json:"valid"`
	Violations      []Violation `json:"violations"`
}
