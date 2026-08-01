package capacitypolicy

import (
	"math"
	"sort"
	"strconv"
)

const maxSupplyBaseUnits int64 = 21_000_000_000_000

type validator struct {
	violations []Violation
}

func ValidateContract(contract Contract) Report {
	v := validator{}
	v.require(contract.Version == ContractVersion, "version", "version must use the maintained capacity contract")
	v.require(contract.QualificationID == MaintainedQualificationID, "qualification_id", "qualification_id must use the fixed synthetic identifier")
	v.require(contract.Environment == MaintainedEnvironment, "environment", "environment must be synthetic-loopback")
	v.validateWorkload(contract.Workload)
	v.validateStorage(contract.Storage)
	v.validateTelemetry(contract.Telemetry)
	v.validateLogging(contract.Logging)
	v.validateLimits(contract.Limits, contract.Logging)
	v.requireExactSet(contract.AbortConditions, MaintainedAbortConditions, "abort_conditions")
	return Report{
		Version: ContractVersion, QualificationID: MaintainedQualificationID,
		ValidatorCount: MaintainedValidatorCount, TransactionCount: MaintainedTransactionCount,
		Valid: len(v.violations) == 0, Violations: append([]Violation{}, v.violations...),
	}
}

func ValidateEvidence(contract Contract, evidence Evidence) EvidenceReport {
	v := validator{}
	contractReport := ValidateContract(contract)
	if !contractReport.Valid {
		v.violations = append(v.violations, Violation{Check: "contract", Message: "capacity contract must validate before evidence"})
	}
	v.require(evidence.Version == EvidenceVersion, "version", "version must use the maintained capacity evidence contract")
	v.require(evidence.QualificationID == MaintainedQualificationID, "qualification_id", "qualification_id must use the fixed synthetic identifier")
	v.require(evidence.Environment == MaintainedEnvironment, "environment", "environment must be synthetic-loopback")
	v.validateWorkloadEvidence(contract, evidence.Workload)
	v.validateNodes(contract, evidence.Nodes)
	v.validateMetrics(contract, evidence.Metrics, evidence.Workload.EndHeight)
	v.validateRetention(contract, evidence.Retention)
	v.require(evidence.Consensus.CommonAppHash, "consensus.common_app_hash", "all validators must share one app hash")
	v.require(evidence.Consensus.ValidatorPowerConsistent, "consensus.validator_power_consistent", "validator power must remain consistent")
	v.require(evidence.Consensus.RestartVerified, "consensus.restart_verified", "one validator restart and catch-up must be verified")
	v.require(evidence.Consensus.LedgerValid, "consensus.ledger_valid", "the exported ledger must validate")
	v.validateProjection(contract, evidence)
	return EvidenceReport{
		Version: EvidenceVersion, QualificationID: MaintainedQualificationID,
		ValidatorCount: MaintainedValidatorCount, CommittedCount: MaintainedTransactionCount,
		Valid: len(v.violations) == 0, Violations: append([]Violation{}, v.violations...),
	}
}

func (v *validator) validateWorkload(workload Workload) {
	v.require(workload.ValidatorCount == MaintainedValidatorCount, "workload.validator_count", "exactly four validators are required")
	v.require(workload.TransactionCount == MaintainedTransactionCount, "workload.transaction_count", "exactly 96 transactions are required")
	v.require(workload.MinimumBlockDelta == MaintainedMinimumBlockDelta, "workload.minimum_block_delta", "minimum block delta must equal the maintained workload")
	v.require(workload.MaximumDurationSeconds >= 60 && workload.MaximumDurationSeconds <= 900, "workload.maximum_duration_seconds", "duration must be bounded between 60 and 900 seconds")
	v.require(workload.MaximumCommitLatencyMS >= 5_000 && workload.MaximumCommitLatencyMS <= 60_000, "workload.maximum_commit_latency_ms", "commit latency must be bounded between 5 and 60 seconds")
}

func (v *validator) validateStorage(storage StoragePolicy) {
	v.require(storage.Pruning == "default", "storage.pruning", "only bounded default pruning is supported")
	v.require(storage.SnapshotInterval == MaintainedSnapshotInterval, "storage.snapshot_interval", "snapshot interval must use the maintained value")
	v.require(storage.SnapshotKeepRecent == MaintainedSnapshotKeepRecent, "storage.snapshot_keep_recent", "snapshot retention must use the maintained finite value")
	v.require(storage.ProjectionBlocks == MaintainedProjectionBlocks, "storage.projection_blocks", "projection horizon must use the maintained finite value")
	v.require(storage.MaxNodeDataGrowthBytes > 0 && storage.MaxNodeDataGrowthBytes <= 2*1024*1024*1024, "storage.max_node_data_growth_bytes", "node data growth envelope must be positive and at most 2 GiB")
}

func (v *validator) validateTelemetry(telemetry TelemetryPolicy) {
	v.require(telemetry.Enabled, "telemetry.enabled", "telemetry must be enabled for qualification")
	v.require(telemetry.Bind == "127.0.0.1", "telemetry.bind", "telemetry must bind to literal IPv4 loopback")
	v.require(telemetry.PrometheusRetentionSeconds == MaintainedTelemetryRetention, "telemetry.prometheus_retention_seconds", "telemetry retention must use the maintained finite value")
	v.requireExactSet(telemetry.RequiredMetrics, MaintainedMetrics, "telemetry.required_metrics")
}

func (v *validator) validateLogging(logging LoggingPolicy) {
	v.require(logging.Driver == "json-file", "logging.driver", "local qualification uses the json-file driver")
	v.require(logging.MaxSizeBytes == MaintainedLogMaxSizeBytes, "logging.max_size_bytes", "local log rotation must use the maintained finite size")
	v.require(logging.MaxFiles == MaintainedLogMaxFiles, "logging.max_files", "local log rotation must use the maintained finite file count")
}

func (v *validator) validateLimits(limits Limits, logging LoggingPolicy) {
	v.require(limits.MaxLogGrowthBytesPerNode > 0 && limits.MaxLogGrowthBytesPerNode <= logging.MaxSizeBytes, "limits.max_log_growth_bytes_per_node", "log growth envelope must be positive and no larger than one retained file")
	v.require(limits.MaxRSSBytesPerNode >= 64*1024*1024 && limits.MaxRSSBytesPerNode <= 4*1024*1024*1024, "limits.max_rss_bytes_per_node", "RSS envelope must be between 64 MiB and 4 GiB")
	v.require(limits.MaxGoroutines > 0 && limits.MaxGoroutines <= 100_000, "limits.max_goroutines", "goroutine envelope must be positive and finite")
}

func (v *validator) validateWorkloadEvidence(contract Contract, evidence WorkloadEvidence) {
	v.require(evidence.Validators == MaintainedValidatorCount, "workload.validators", "evidence must cover exactly four validators")
	v.require(evidence.TransactionsSubmitted == MaintainedTransactionCount, "workload.transactions_submitted", "evidence must submit the maintained transaction count")
	v.require(evidence.TransactionsCommitted == MaintainedTransactionCount, "workload.transactions_committed", "every submitted transaction must commit")
	v.require(evidence.TransactionsFailed == 0, "workload.transactions_failed", "failed transactions are forbidden")
	v.require(evidence.StartHeight > 0 && evidence.EndHeight > evidence.StartHeight, "workload.height", "workload heights must show forward progress")
	delta, ok := checkedSubtract(evidence.EndHeight, evidence.StartHeight)
	v.require(ok && evidence.BlockDelta == delta && evidence.BlockDelta >= contract.Workload.MinimumBlockDelta, "workload.block_delta", "block delta must be exact and meet the maintained minimum")
	v.require(evidence.DurationMS > 0 && evidence.DurationMS <= contract.Workload.MaximumDurationSeconds*1000, "workload.duration_ms", "workload duration must be positive and within its envelope")
	expectedThroughput, throughputOK := checkedMultiplyDivide(int64(evidence.TransactionsCommitted), 1_000_000, evidence.DurationMS)
	v.require(throughputOK && evidence.ThroughputMilliTPS == expectedThroughput && evidence.ThroughputMilliTPS > 0, "workload.throughput_milli_tps", "throughput must be the exact positive committed transaction rate")
	expectedAverageBlockTime := int64(0)
	averageBlockTimeOK := evidence.DurationMS > 0 && evidence.BlockDelta > 0
	if averageBlockTimeOK {
		expectedAverageBlockTime = evidence.DurationMS / evidence.BlockDelta
	}
	v.require(averageBlockTimeOK && evidence.AverageBlockTimeMS == expectedAverageBlockTime && evidence.AverageBlockTimeMS > 0, "workload.average_block_time_ms", "average block time must be the exact positive workload duration ratio")
	v.require(evidence.P95CommitLatencyMS > 0 && evidence.P95CommitLatencyMS <= evidence.MaxCommitLatencyMS, "workload.p95_commit_latency_ms", "p95 commit latency must be positive and no larger than the maximum")
	v.require(evidence.MaxCommitLatencyMS <= contract.Workload.MaximumCommitLatencyMS, "workload.max_commit_latency_ms", "maximum commit latency breached its envelope")
}

func (v *validator) validateNodes(contract Contract, nodes []NodeEvidence) {
	v.require(len(nodes) == MaintainedValidatorCount, "nodes", "evidence must contain exactly four node summaries")
	seen := make(map[string]bool)
	for index, node := range nodes {
		expected := "validator-" + strconv.Itoa(index+1)
		v.require(node.Name == expected, "nodes.name", "node summaries must use fixed ordered synthetic names")
		v.require(!seen[node.Name], "nodes.name", "node summary names must be unique")
		seen[node.Name] = true
		v.require(node.DataBytesStart >= 0 && node.DataBytesPeak >= node.DataBytesStart, "nodes.data_bytes", "data measurements must be non-negative and monotonic at peak")
		growth, ok := checkedSubtract(node.DataBytesPeak, node.DataBytesStart)
		v.require(ok && growth == node.DataGrowthBytes && growth <= contract.Storage.MaxNodeDataGrowthBytes, "nodes.data_growth_bytes", "data growth must be exact and within its envelope")
		v.require(node.LogBytesStart >= 0 && node.LogBytesPeak >= node.LogBytesStart, "nodes.log_bytes", "log measurements must be non-negative and monotonic at peak")
		logGrowth, ok := checkedSubtract(node.LogBytesPeak, node.LogBytesStart)
		v.require(ok && logGrowth == node.LogGrowthBytes && logGrowth <= contract.Limits.MaxLogGrowthBytesPerNode, "nodes.log_growth_bytes", "log growth must be exact and within its envelope")
		v.require(node.MaxRSSBytes > 0 && node.MaxRSSBytes <= contract.Limits.MaxRSSBytesPerNode, "nodes.max_rss_bytes", "RSS observation must be positive and within its envelope")
	}
}

func (v *validator) validateMetrics(contract Contract, metrics MetricsEvidence, endHeight int64) {
	v.require(metrics.ConsensusHeight >= endHeight, "metrics.consensus_height", "consensus metric must cover the workload end height")
	v.require(metrics.ApplicationHeight == metrics.InvariantHeight && metrics.ApplicationHeight >= endHeight, "metrics.application_height", "application and invariant heights must align through the workload")
	v.require(metrics.CompletedBlocks >= contract.Workload.MinimumBlockDelta, "metrics.completed_blocks", "completed-block counter must cover the workload")
	sum, ok := checkedAdd(metrics.SupplyBaseUnits, metrics.SupplyHeadroomUnits)
	v.require(ok && metrics.SupplyBaseUnits >= 0 && metrics.SupplyHeadroomUnits >= 0 && sum == maxSupplyBaseUnits, "metrics.supply", "supply and headroom must preserve the fixed cap")
	v.require(metrics.Goroutines > 0 && metrics.Goroutines <= contract.Limits.MaxGoroutines, "metrics.goroutines", "goroutine observation must be positive and within its envelope")
	v.require(metrics.ResidentMemoryBytes > 0 && metrics.ResidentMemoryBytes <= contract.Limits.MaxRSSBytesPerNode, "metrics.resident_memory_bytes", "resident-memory metric must be positive and within its envelope")
}

func (v *validator) validateRetention(contract Contract, evidence RetentionEvidence) {
	v.require(evidence.Pruning == contract.Storage.Pruning, "retention.pruning", "evidence must use the qualified pruning mode")
	v.require(evidence.SnapshotInterval == contract.Storage.SnapshotInterval, "retention.snapshot_interval", "snapshot interval must match the contract")
	v.require(evidence.SnapshotKeepRecent == contract.Storage.SnapshotKeepRecent, "retention.snapshot_keep_recent", "snapshot retention must match the contract")
	v.require(evidence.ObservedSnapshotCount > 0 && evidence.ObservedSnapshotCount <= contract.Storage.SnapshotKeepRecent, "retention.observed_snapshot_count", "observed snapshot count must be positive and within retention")
}

func (v *validator) validateProjection(contract Contract, evidence Evidence) {
	v.require(evidence.Projection.Blocks == contract.Storage.ProjectionBlocks, "projection.blocks", "projection must use the maintained finite horizon")
	maxGrowth := int64(0)
	for _, node := range evidence.Nodes {
		if node.DataGrowthBytes > maxGrowth {
			maxGrowth = node.DataGrowthBytes
		}
	}
	projected, ok := ProjectedGrowthBytes(maxGrowth, evidence.Workload.BlockDelta, contract.Storage.ProjectionBlocks)
	v.require(ok && evidence.Projection.MaxProjectedGrowthBytes == projected, "projection.max_projected_growth_bytes", "projected growth must use checked maintained arithmetic")
}

func ProjectedGrowthBytes(observedGrowth, observedBlocks, projectionBlocks int64) (int64, bool) {
	if observedGrowth < 0 || observedBlocks <= 0 || projectionBlocks <= 0 {
		return 0, false
	}
	if observedGrowth == 0 {
		return 0, true
	}
	if observedGrowth > math.MaxInt64/projectionBlocks {
		return 0, false
	}
	product := observedGrowth * projectionBlocks
	if product > math.MaxInt64-(observedBlocks-1) {
		return 0, false
	}
	return (product + observedBlocks - 1) / observedBlocks, true
}

func (v *validator) requireExactSet(got, want []string, check string) {
	if len(got) != len(want) {
		v.require(false, check, "values must contain the exact maintained set")
		return
	}
	gotCopy, wantCopy := append([]string(nil), got...), append([]string(nil), want...)
	sort.Strings(gotCopy)
	sort.Strings(wantCopy)
	for index := range gotCopy {
		if index > 0 && gotCopy[index] == gotCopy[index-1] {
			v.require(false, check, "values must be unique")
			return
		}
		if gotCopy[index] != wantCopy[index] {
			v.require(false, check, "values must contain the exact maintained set")
			return
		}
	}
}

func (v *validator) require(condition bool, check, message string) {
	if !condition {
		v.violations = append(v.violations, Violation{Check: check, Message: message})
	}
}

func checkedAdd(left, right int64) (int64, bool) {
	if (right > 0 && left > math.MaxInt64-right) || (right < 0 && left < math.MinInt64-right) {
		return 0, false
	}
	return left + right, true
}

func checkedSubtract(left, right int64) (int64, bool) {
	if (right < 0 && left > math.MaxInt64+right) || (right > 0 && left < math.MinInt64+right) {
		return 0, false
	}
	return left - right, true
}

func checkedMultiplyDivide(value, multiplier, divisor int64) (int64, bool) {
	if value < 0 || multiplier < 0 || divisor <= 0 || (value > 0 && multiplier > math.MaxInt64/value) {
		return 0, false
	}
	return value * multiplier / divisor, true
}
