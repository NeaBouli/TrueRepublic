package capacitypolicy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseContractAndEvidenceStrict(t *testing.T) {
	fixtureContract := readFixture(t, "qualification.example.json")
	contract, err := ParseContract(bytes.NewReader(fixtureContract))
	if err != nil {
		t.Fatalf("parse fixture contract: %v", err)
	}
	fixtureEvidence, err := json.MarshalIndent(evidenceFromContract(contract), "", "  ")
	if err != nil {
		t.Fatalf("marshal evidence fixture: %v", err)
	}

	cases := []struct {
		name     string
		parse    func([]byte) error
		input    []byte
		wantErr  bool
		contains string
	}{
		{
			name:  "contract valid fixture",
			parse: func(raw []byte) error { _, err := ParseContract(bytes.NewReader(raw)); return err },
			input: fixtureContract,
		},
		{
			name:     "contract empty",
			parse:    func(raw []byte) error { _, err := ParseContract(bytes.NewReader(raw)); return err },
			input:    []byte(" "),
			wantErr:  true,
			contains: "capacity contract is empty",
		},
		{
			name:     "contract trailing",
			parse:    func(raw []byte) error { _, err := ParseContract(bytes.NewReader(raw)); return err },
			input:    []byte(`{}` + "\n" + `null`),
			wantErr:  true,
			contains: "trailing value is forbidden",
		},
		{
			name:     "contract duplicate key",
			parse:    func(raw []byte) error { _, err := ParseContract(bytes.NewReader(raw)); return err },
			input:    []byte(`{"version":"x","version":"y"}`),
			wantErr:  true,
			contains: "duplicate object key",
		},
		{
			name:     "contract unknown field",
			parse:    func(raw []byte) error { _, err := ParseContract(bytes.NewReader(raw)); return err },
			input:    []byte(`{"version":"` + ContractVersion + `","qualification_id":"` + MaintainedQualificationID + `","environment":"` + MaintainedEnvironment + `","workload":{"validator_count":4,"transaction_count":96,"minimum_block_delta":24,"maximum_duration_seconds":600,"maximum_commit_latency_ms":60000},"storage":{"pruning":"default","snapshot_interval":2,"snapshot_keep_recent":3,"projection_blocks":1000000,"max_node_data_growth_bytes":268435456},"telemetry":{"enabled":true,"bind":"127.0.0.1","prometheus_retention_seconds":60,"required_metrics":["process_resident_memory_bytes","go_goroutines","cometbft_consensus_height","truerepublic_app_last_successful_block_height","truerepublic_app_last_successful_invariant_cycle_height","truerepublic_app_completed_blocks_total","truerepublic_token_pnyx_supply_base_units","truerepublic_token_pnyx_supply_headroom_base_units"]},"logging":{"driver":"json-file","max_size_bytes":52428800,"max_files":3},"limits":{"max_log_growth_bytes_per_node":52428800,"max_rss_bytes_per_node":2147483648,"max_goroutines":8192},"abort_conditions":["transaction-failure","consensus-progress-stalled","app-hash-divergence","validator-power-divergence","ledger-invariant-failure","resource-envelope-breach","retention-unbounded","measurement-overflow","secret-material-detected"],"secret_material_detected":"redacted"}`),
			wantErr:  true,
			contains: "invalid capacity contract schema",
		},
		{
			name:     "contract max depth",
			parse:    func(raw []byte) error { _, err := ParseContract(bytes.NewReader(raw)); return err },
			input:    []byte(multiDepthJSON(33)),
			wantErr:  true,
			contains: "maximum JSON depth exceeded",
		},
		{
			name:     "contract oversize",
			parse:    func(raw []byte) error { _, err := ParseContract(bytes.NewReader(raw)); return err },
			input:    bytes.Repeat([]byte(" "), MaxDocumentBytes+1),
			wantErr:  true,
			contains: "capacity contract exceeds",
		},
		{
			name:  "evidence valid fixture",
			parse: func(raw []byte) error { _, err := ParseEvidence(bytes.NewReader(raw)); return err },
			input: fixtureEvidence,
		},
		{
			name:     "evidence empty",
			parse:    func(raw []byte) error { _, err := ParseEvidence(bytes.NewReader(raw)); return err },
			input:    []byte(""),
			wantErr:  true,
			contains: "capacity evidence is empty",
		},
		{
			name:     "evidence max depth",
			parse:    func(raw []byte) error { _, err := ParseEvidence(bytes.NewReader(raw)); return err },
			input:    []byte(multiDepthJSON(33)),
			wantErr:  true,
			contains: "maximum JSON depth exceeded",
		},
		{
			name:     "evidence unknown field",
			parse:    func(raw []byte) error { _, err := ParseEvidence(bytes.NewReader(raw)); return err },
			input:    []byte(`{"version":"` + EvidenceVersion + `","secret_material_detected":"redacted"}`),
			wantErr:  true,
			contains: "invalid capacity evidence schema",
		},
		{
			name:     "evidence trailing",
			parse:    func(raw []byte) error { _, err := ParseEvidence(bytes.NewReader(raw)); return err },
			input:    []byte(`{}` + "\n" + `null`),
			wantErr:  true,
			contains: "trailing value is forbidden",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.parse(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected parse failure")
				}
				if tc.contains != "" && !strings.Contains(err.Error(), tc.contains) {
					t.Fatalf("expected %q in %q", tc.contains, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected parse failure: %v", err)
			}
		})
	}
}

func TestValidateContractAndEvidence_GatedFields(t *testing.T) {
	contract, err := ParseContract(bytes.NewReader(readFixture(t, "qualification.example.json")))
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	evidenceFixture := evidenceFromContract(contract)
	evidenceJSON, err := json.Marshal(evidenceFixture)
	if err != nil {
		t.Fatalf("marshal evidence fixture: %v", err)
	}
	evidence, err := ParseEvidence(bytes.NewReader(evidenceJSON))
	if err != nil {
		t.Fatalf("parse evidence fixture: %v", err)
	}

	t.Run("contract valid fixture", func(t *testing.T) {
		report := ValidateContract(contract)
		if !report.Valid {
			t.Fatalf("expected valid fixture contract: %+v", report)
		}
		if report.Violations == nil {
			t.Fatal("valid contract report must encode violations as an empty array")
		}
	})

	contractCases := []struct {
		name   string
		mutate func(*Contract)
		check  string
	}{
		{name: "fixed metadata version", mutate: func(c *Contract) { c.Version = "wrong/version" }, check: "version"},
		{name: "fixed metadata qualification", mutate: func(c *Contract) { c.QualificationID = "wrong-id" }, check: "qualification_id"},
		{name: "fixed metadata environment", mutate: func(c *Contract) { c.Environment = "mainnet" }, check: "environment"},
		{name: "exact metrics set", mutate: func(c *Contract) { c.Telemetry.RequiredMetrics = append(c.Telemetry.RequiredMetrics, "extra.metric") }, check: "telemetry.required_metrics"},
		{name: "exact metrics set duplicate", mutate: func(c *Contract) {
			c.Telemetry.RequiredMetrics = append(append([]string(nil), c.Telemetry.RequiredMetrics...), c.Telemetry.RequiredMetrics[0])
		}, check: "telemetry.required_metrics"},
		{name: "unbounded pruning", mutate: func(c *Contract) { c.Storage.Pruning = "nothing" }, check: "storage.pruning"},
	}

	for _, tc := range contractCases {
		t.Run(tc.name, func(t *testing.T) {
			c := contract
			tc.mutate(&c)
			report := ValidateContract(c)
			if report.Valid {
				t.Fatal("expected invalid contract")
			}
			if !containsCheck(report.Violations, tc.check) {
				t.Fatalf("expected check %q, got %#v", tc.check, report.Violations)
			}
		})
	}

	evidenceCases := []struct {
		name   string
		mutate func(*Evidence, *Contract)
		checks []string
	}{
		{name: "workload validators", mutate: func(e *Evidence, _ *Contract) { e.Workload.Validators = 3 }, checks: []string{"workload.validators"}},
		{name: "workload failed transaction", mutate: func(e *Evidence, _ *Contract) { e.Workload.TransactionsFailed = 1 }, checks: []string{"workload.transactions_failed"}},
		{name: "workload heights", mutate: func(e *Evidence, _ *Contract) { e.Workload.EndHeight = e.Workload.StartHeight }, checks: []string{"workload.height"}},
		{name: "workload stalled progress", mutate: func(e *Evidence, _ *Contract) { e.Workload.BlockDelta = 0 }, checks: []string{"workload.block_delta"}},
		{name: "workload throughput", mutate: func(e *Evidence, _ *Contract) { e.Workload.ThroughputMilliTPS++ }, checks: []string{"workload.throughput_milli_tps"}},
		{name: "workload block time", mutate: func(e *Evidence, _ *Contract) { e.Workload.AverageBlockTimeMS++ }, checks: []string{"workload.average_block_time_ms"}},
		{name: "workload latency", mutate: func(e *Evidence, c *Contract) { e.Workload.MaxCommitLatencyMS = c.Workload.MaximumCommitLatencyMS + 1 }, checks: []string{"workload.max_commit_latency_ms"}},
		{name: "node disk", mutate: func(e *Evidence, c *Contract) {
			e.Nodes[0].DataGrowthBytes = c.Storage.MaxNodeDataGrowthBytes + 1
			e.Nodes[0].DataBytesPeak = e.Nodes[0].DataBytesStart + e.Nodes[0].DataGrowthBytes
		}, checks: []string{"nodes.data_growth_bytes"}},
		{name: "node log", mutate: func(e *Evidence, c *Contract) {
			e.Nodes[1].LogGrowthBytes = c.Limits.MaxLogGrowthBytesPerNode + 1
			e.Nodes[1].LogBytesPeak = e.Nodes[1].LogBytesStart + e.Nodes[1].LogGrowthBytes
		}, checks: []string{"nodes.log_growth_bytes"}},
		{name: "node rss", mutate: func(e *Evidence, _ *Contract) { e.Nodes[2].MaxRSSBytes = -1 }, checks: []string{"nodes.max_rss_bytes"}},
		{name: "metrics height", mutate: func(e *Evidence, _ *Contract) { e.Metrics.ConsensusHeight = e.Workload.EndHeight - 1 }, checks: []string{"metrics.consensus_height"}},
		{name: "metrics goroutines", mutate: func(e *Evidence, c *Contract) { e.Metrics.Goroutines = c.Limits.MaxGoroutines + 1 }, checks: []string{"metrics.goroutines"}},
		{name: "evidence consensus", mutate: func(e *Evidence, _ *Contract) { e.Consensus.ValidatorPowerConsistent = false }, checks: []string{"consensus.validator_power_consistent"}},
		{name: "evidence restart", mutate: func(e *Evidence, _ *Contract) { e.Consensus.RestartVerified = false }, checks: []string{"consensus.restart_verified"}},
		{name: "evidence ledger", mutate: func(e *Evidence, _ *Contract) { e.Consensus.LedgerValid = false }, checks: []string{"consensus.ledger_valid"}},
		{name: "evidence projection", mutate: func(e *Evidence, _ *Contract) { e.Projection.MaxProjectedGrowthBytes = 0 }, checks: []string{"projection.max_projected_growth_bytes"}},
		{name: "evidence retention", mutate: func(e *Evidence, c *Contract) { e.Retention.Pruning = c.Storage.Pruning + "-none" }, checks: []string{"retention.pruning"}},
	}

	for _, tc := range evidenceCases {
		t.Run(tc.name, func(t *testing.T) {
			c := contract
			e := cloneEvidence(evidence)
			tc.mutate(&e, &c)
			report := ValidateEvidence(c, e)
			if report.Valid {
				t.Fatal("expected invalid evidence")
			}
			for _, want := range tc.checks {
				if !containsCheck(report.Violations, want) {
					t.Fatalf("expected check %q, got %#v", want, report.Violations)
				}
			}
		})
	}

	t.Run("evidence valid fixture", func(t *testing.T) {
		report := ValidateEvidence(contract, evidence)
		if !report.Valid {
			t.Fatalf("expected valid evidence: %+v", report)
		}
		if report.Violations == nil {
			t.Fatal("valid evidence report must encode violations as an empty array")
		}
	})

	t.Run("checked overflow", func(t *testing.T) {
		overflowEvidence := cloneEvidence(evidence)
		overflowEvidence.Metrics.SupplyBaseUnits = maxSupplyBaseUnits
		overflowEvidence.Metrics.SupplyHeadroomUnits = maxSupplyBaseUnits
		report := ValidateEvidence(contract, overflowEvidence)
		if report.Valid || !containsCheck(report.Violations, "metrics.supply") {
			t.Fatalf("expected checked overflow rejection: %+v", report)
		}
	})
}

func TestCapacityPolicyCommands(t *testing.T) {
	dir := t.TempDir()
	contractPath := filepath.Join(dir, "contract.json")
	if err := os.WriteFile(contractPath, readFixture(t, "qualification.example.json"), 0o600); err != nil {
		t.Fatalf("write contract fixture: %v", err)
	}

	contract, err := ParseContract(bytes.NewReader(readFixture(t, "qualification.example.json")))
	if err != nil {
		t.Fatalf("parse contract: %v", err)
	}
	evidence := evidenceFromContract(contract)
	evidenceBytes, err := json.Marshal(evidence)
	if err != nil {
		t.Fatalf("marshal evidence fixture: %v", err)
	}
	validEvidencePath := filepath.Join(dir, "evidence.json")
	if err := os.WriteFile(validEvidencePath, evidenceBytes, 0o600); err != nil {
		t.Fatalf("write evidence fixture: %v", err)
	}
	badEvidence := cloneEvidence(evidence)
	badEvidence.Metrics.CompletedBlocks = 0
	badEvidenceJSON, err := json.Marshal(badEvidence)
	if err != nil {
		t.Fatalf("marshal bad evidence: %v", err)
	}
	badEvidencePath := filepath.Join(dir, "evidence-bad.json")
	if err := os.WriteFile(badEvidencePath, badEvidenceJSON, 0o600); err != nil {
		t.Fatalf("write bad evidence: %v", err)
	}
	sensitiveContract := []byte(`{"version":"wrong","secret_material_detected":"super-secret-token"}`)
	sensitivePath := filepath.Join(dir, "secret-identifier.json")
	if err := os.WriteFile(sensitivePath, sensitiveContract, 0o600); err != nil {
		t.Fatalf("write sensitive input: %v", err)
	}

	t.Run("validate json success", func(t *testing.T) {
		out, err := runCapacityCommand(t, "validate", "--file", contractPath, "--output", "json")
		if err != nil {
			t.Fatalf("expected success: %v\n%s", err, out)
		}
		if !strings.Contains(out, "\"valid\": true") {
			t.Fatalf("unexpected report output: %s", out)
		}
	})

	t.Run("verify json failure", func(t *testing.T) {
		out, err := runCapacityCommand(t, "verify", "--contract", contractPath, "--evidence", badEvidencePath, "--output", "json")
		if err == nil {
			t.Fatal("expected command failure")
		}
		var report EvidenceReport
		if err := json.Unmarshal([]byte(out), &report); err != nil {
			t.Fatalf("decode report: %v", err)
		}
		if report.Valid || !containsCheck(report.Violations, "metrics.completed_blocks") {
			t.Fatalf("expected structured invalid report: %+v", report)
		}
	})

	t.Run("command no-reflection of secrets/path", func(t *testing.T) {
		out, err := runCapacityCommand(t, "validate", "--file", sensitivePath, "--output", "json")
		if err == nil {
			t.Fatal("expected failure")
		}
		if strings.Contains(out, "super-secret-token") {
			t.Fatalf("secret leaked: %s", out)
		}
		if strings.Contains(out, filepath.Base(sensitivePath)) {
			t.Fatalf("path leaked: %s", out)
		}
	})

	t.Run("proof command output stays text/json stable", func(t *testing.T) {
		out, err := runCapacityCommand(t, "verify", "--contract", contractPath, "--evidence", validEvidencePath, "--output", "text")
		if err != nil {
			t.Fatalf("expected success: %v\n%s", err, out)
		}
		if !strings.Contains(out, "OK: capacity evidence validates 4 validators and 96 committed transactions") {
			t.Fatalf("unexpected proof text output: %s", out)
		}
	})
}

func TestProjectedGrowthBytesOverflow(t *testing.T) {
	t.Run("overflow", func(t *testing.T) {
		if _, ok := ProjectedGrowthBytes(math.MaxInt64, 1, math.MaxInt64); ok {
			t.Fatal("expected overflow")
		}
	})
}

func runCapacityCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := NewCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func evidenceFromContract(contract Contract) Evidence {
	startHeight := int64(1000)
	endHeight := startHeight + contract.Workload.MinimumBlockDelta
	dataGrowth := int64(10_000)
	maxProjected, _ := ProjectedGrowthBytes(dataGrowth, contract.Workload.MinimumBlockDelta, contract.Storage.ProjectionBlocks)
	nodes := make([]NodeEvidence, 0, MaintainedValidatorCount)
	for i := 1; i <= MaintainedValidatorCount; i++ {
		nodes = append(nodes, NodeEvidence{
			Name:            fmt.Sprintf("validator-%d", i),
			DataBytesStart:  10_000,
			DataBytesPeak:   10_000 + dataGrowth,
			DataGrowthBytes: dataGrowth,
			LogBytesStart:   10_000,
			LogBytesPeak:    10_100,
			LogGrowthBytes:  100,
			MaxRSSBytes:     contract.Limits.MaxRSSBytesPerNode / 4,
		})
	}

	return Evidence{
		Version:         EvidenceVersion,
		QualificationID: MaintainedQualificationID,
		Environment:     MaintainedEnvironment,
		Workload: WorkloadEvidence{
			Validators:            MaintainedValidatorCount,
			TransactionsSubmitted: MaintainedTransactionCount,
			TransactionsCommitted: MaintainedTransactionCount,
			TransactionsFailed:    0,
			StartHeight:           startHeight,
			EndHeight:             endHeight,
			BlockDelta:            endHeight - startHeight,
			DurationMS:            300_000,
			ThroughputMilliTPS:    int64(MaintainedTransactionCount) * 1_000_000 / 300_000,
			AverageBlockTimeMS:    300_000 / contract.Workload.MinimumBlockDelta,
			P95CommitLatencyMS:    2_000,
			MaxCommitLatencyMS:    2_500,
		},
		Nodes: nodes,
		Metrics: MetricsEvidence{
			ConsensusHeight:     endHeight,
			ApplicationHeight:   endHeight,
			InvariantHeight:     endHeight,
			CompletedBlocks:     contract.Workload.MinimumBlockDelta,
			SupplyBaseUnits:     1_000_000_000,
			SupplyHeadroomUnits: maxSupplyBaseUnits - 1_000_000_000,
			Goroutines:          contract.Limits.MaxGoroutines / 2,
			ResidentMemoryBytes: contract.Limits.MaxRSSBytesPerNode / 2,
		},
		Retention: RetentionEvidence{
			Pruning:               contract.Storage.Pruning,
			SnapshotInterval:      contract.Storage.SnapshotInterval,
			SnapshotKeepRecent:    contract.Storage.SnapshotKeepRecent,
			ObservedSnapshotCount: 1,
		},
		Consensus: ConsensusEvidence{
			CommonAppHash:            true,
			ValidatorPowerConsistent: true,
			RestartVerified:          true,
			LedgerValid:              true,
		},
		Projection: ProjectionEvidence{
			Blocks:                  contract.Storage.ProjectionBlocks,
			MaxProjectedGrowthBytes: maxProjected,
		},
	}
}

func cloneEvidence(e Evidence) Evidence {
	c := e
	c.Nodes = append([]NodeEvidence(nil), e.Nodes...)
	return c
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("..", "configs", "capacity", name)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	return raw
}

func containsCheck(violations []Violation, check string) bool {
	for _, violation := range violations {
		if violation.Check == check {
			return true
		}
	}
	return false
}

func multiDepthJSON(depth int) string {
	out := "{}"
	for i := 0; i < depth; i++ {
		out = `{"a":` + out + `}`
	}
	return out
}
