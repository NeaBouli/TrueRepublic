package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"truerepublic/capacitypolicy"
	"truerepublic/token"
)

const capacityEvidenceOutputEnv = "TRUEREPUBLIC_CAPACITY_EVIDENCE_OUT"

func TestMultiValidatorCapacitySustainedLoad(t *testing.T) {
	if os.Getenv(multiValidatorSmokeEnv) != "1" {
		t.Skipf("set %s=1 to run the multi-validator process harness", multiValidatorSmokeEnv)
	}
	ctx := t.Context()
	contract, err := capacitypolicy.LoadContract("configs/capacity/qualification.example.json")
	if err != nil {
		t.Fatal(err)
	}
	if report := capacitypolicy.ValidateContract(contract); !report.Valid {
		t.Fatalf("maintained capacity contract is invalid: %+v", report)
	}

	binary := filepath.Join(t.TempDir(), "truerepublicd")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build daemon: %v\n%s", err, output)
	}

	const chainID = "truerepublic-capacity-1"
	validators := make([]*smokeValidator, capacitypolicy.MaintainedValidatorCount)
	for index := range validators {
		validator := &smokeValidator{
			name:    fmt.Sprintf("validator-%d", index+1),
			home:    filepath.Join(t.TempDir(), fmt.Sprintf("node-%d", index+1)),
			rpcPort: freeTCPPort(t),
			p2pPort: freeTCPPort(t),
			logPath: filepath.Join(t.TempDir(), fmt.Sprintf("validator-%d.log", index+1)),
		}
		initSmokeValidator(t, ctx, binary, chainID, validator)
		validators[index] = validator
	}
	accounts := make([]smokeAccount, capacitypolicy.MaintainedValidatorCount)
	for index := range accounts {
		accounts[index] = addSmokeKey(
			t, ctx, binary, validators[0].home, fmt.Sprintf("capacity-sender-%d", index+1),
			uint64(index+4), 1_000_000*token.WholeTokenBaseUnits,
		)
	}
	sharedGenesis := buildSharedSmokeGenesis(t, chainID, validators, accounts...)
	for _, validator := range validators {
		if err := atomicWriteFile(filepath.Join(validator.home, "config", "genesis.json"), sharedGenesis, 0o600); err != nil {
			t.Fatalf("write %s shared genesis: %v", validator.name, err)
		}
	}

	cometMetricsPort := freeTCPPort(t)
	configureCapacityPrometheus(t, filepath.Join(validators[0].home, "config", "config.toml"), cometMetricsPort)
	apiPort := freeTCPPort(t)
	t.Setenv("TRUEREPUBLICD_TELEMETRY_ENABLED", "true")
	t.Setenv("TRUEREPUBLICD_TELEMETRY_PROMETHEUS_RETENTION_TIME", strconv.Itoa(capacitypolicy.MaintainedTelemetryRetention))
	t.Setenv("TRUEREPUBLICD_TELEMETRY_SERVICE_NAME", "truerepublic")
	t.Setenv("TRUEREPUBLICD_TELEMETRY_ENABLE_HOSTNAME", "false")
	t.Setenv("TRUEREPUBLICD_TELEMETRY_ENABLE_HOSTNAME_LABEL", "false")
	t.Setenv("TRUEREPUBLICD_TELEMETRY_ENABLE_SERVICE_LABEL", "false")

	startValidator := func(validator *smokeValidator, exposeAPI bool) error {
		extra := []string{
			"--pruning", contract.Storage.Pruning,
			"--state-sync.snapshot-interval", strconv.FormatInt(contract.Storage.SnapshotInterval, 10),
			"--state-sync.snapshot-keep-recent", strconv.Itoa(contract.Storage.SnapshotKeepRecent),
		}
		if exposeAPI {
			extra = append(extra,
				"--api.enable=true",
				"--api.address", fmt.Sprintf("tcp://127.0.0.1:%d", apiPort),
			)
		}
		return validator.startWithArgs(ctx, binary, persistentPeers(validator, validators), extra...)
	}

	t.Cleanup(func() {
		for _, validator := range validators {
			_ = validator.stop(false)
		}
		if t.Failed() {
			for _, validator := range validators {
				validator.logContents(t)
			}
		}
	})
	for index, validator := range validators {
		if err := startValidator(validator, index == 0); err != nil {
			t.Fatalf("start %s: %v", validator.name, err)
		}
	}
	waitForSmokeHeight(t, validators, 3, 120*time.Second)
	for index := range accounts {
		runSmokeTx(t, ctx, binary, validators[0], &accounts[index], chainID,
			"create-domain", fmt.Sprintf("CapacityQualification%d", index+1),
			fmt.Sprintf("%d%s", 500_000*token.WholeTokenBaseUnits, token.BaseDenom))
	}
	startHeight := smokeHeight(t, validators[0])

	nodeMeasurements := make([]capacityNodeMeasurement, len(validators))
	for index, validator := range validators {
		nodeMeasurements[index] = capacityNodeMeasurement{
			name:      validator.name,
			dataStart: directoryRegularBytes(t, filepath.Join(validator.home, "data")),
			logStart:  regularFileBytes(t, validator.logPath),
		}
		sampleCapacityNode(t, validator, &nodeMeasurements[index])
	}

	latencies := make([]time.Duration, 0, capacitypolicy.MaintainedTransactionCount)
	workloadStart := time.Now()
	for wave := 0; wave < capacitypolicy.MaintainedTransactionWaves; wave++ {
		results := make([]capacityTxSubmission, len(accounts))
		var submissions sync.WaitGroup
		for accountIndex := range accounts {
			submissions.Add(1)
			go func(index int) {
				defer submissions.Done()
				digest := sha256.Sum256([]byte(fmt.Sprintf("truerepublic-capacity-recipient:%02d:%d", wave, index)))
				recipient := sdk.MustBech32ifyAddressBytes("truerepublic", digest[:20])
				results[index] = submitCapacityTx(ctx, binary, validators[0], accounts[index], chainID, fmt.Sprintf("CapacityQualification%d", index+1), recipient)
			}(accountIndex)
		}
		submissions.Wait()
		for index, result := range results {
			if result.err != nil {
				t.Fatalf("capacity transaction wave %d sender %d: %v", wave+1, index+1, result.err)
			}
			accounts[index].sequence++
		}
		deliveries := make([]capacityTxDelivery, len(results))
		var deliveryChecks sync.WaitGroup
		for resultIndex := range results {
			deliveryChecks.Add(1)
			go func(index int) {
				defer deliveryChecks.Done()
				deliveries[index].err = waitForCapacityTx(ctx, validators[0], results[index].hash)
				deliveries[index].latency = time.Since(results[index].started)
			}(resultIndex)
		}
		deliveryChecks.Wait()
		for index, delivery := range deliveries {
			if delivery.err != nil {
				t.Fatalf("capacity transaction wave %d sender %d delivery failed: %v", wave+1, index+1, delivery.err)
			}
			latencies = append(latencies, delivery.latency)
		}
		for measurementIndex, validator := range validators {
			sampleCapacityNode(t, validator, &nodeMeasurements[measurementIndex])
		}
	}
	workloadDuration := time.Since(workloadStart)
	workloadEndHeight := smokeHeight(t, validators[0])
	waitForSmokeHeight(t, validators, workloadEndHeight, 120*time.Second)
	assertCommonAppHash(t, validators, workloadEndHeight)
	assertSmokeValidatorPowers(t, validators[0], validators, "1")

	if err := validators[len(validators)-1].stop(true); err != nil {
		t.Fatalf("stop validator for restart qualification: %v", err)
	}
	restartTarget := smokeHeight(t, validators[0]) + 2
	if err := startValidator(validators[len(validators)-1], false); err != nil {
		t.Fatalf("restart validator: %v", err)
	}
	waitForSmokeHeight(t, validators, restartTarget, 120*time.Second)
	finalHeight := smokeHeight(t, validators[0])
	assertCommonAppHash(t, validators, finalHeight)
	assertSmokeValidatorPowers(t, validators[0], validators, "1")
	for index, validator := range validators {
		sampleCapacityNode(t, validator, &nodeMeasurements[index])
	}

	appMetricsURL := fmt.Sprintf("http://127.0.0.1:%d/metrics?format=prometheus", apiPort)
	cometMetricsURL := fmt.Sprintf("http://127.0.0.1:%d/metrics", cometMetricsPort)
	metrics := waitForCapacityMetrics(t, appMetricsURL, cometMetricsURL, finalHeight)
	// Snapshot creation and pruning run asynchronously after a committed block.
	// Stop enough validators to halt new commits while keeping validator-1's RPC
	// available, then observe the settled retention state rather than a transient
	// create-before-prune window.
	for _, validator := range validators[2:] {
		if err := validator.stop(true); err != nil {
			t.Fatalf("stop %s before snapshot retention observation: %v", validator.name, err)
		}
	}
	snapshotCount := waitForSnapshotCount(t, validators[0].home, contract.Storage.SnapshotKeepRecent, 30*time.Second)

	for index, validator := range validators {
		if err := validator.stop(true); err != nil {
			t.Fatalf("stop %s after capacity workload: %v", validator.name, err)
		}
		sampleCapacityNode(t, validator, &nodeMeasurements[index])
		assertStructuredNodeLogs(t, validator.logPath)
	}
	exported := exportSmokeGenesis(t, ctx, binary, validators[0], finalHeight)
	exportApp := newGenesisTestApp(t)
	if err := validateLedgerGenesis(exportApp.appCodec, exported.AppState); err != nil {
		t.Fatalf("capacity export is not exactly bank-backed: %v", err)
	}

	sort.Slice(latencies, func(left, right int) bool { return latencies[left] < latencies[right] })
	p95Index := (95*len(latencies)+99)/100 - 1
	nodes := make([]capacitypolicy.NodeEvidence, len(nodeMeasurements))
	maxDataGrowth := int64(0)
	for index, measurement := range nodeMeasurements {
		dataGrowth := measurement.dataPeak - measurement.dataStart
		logGrowth := measurement.logPeak - measurement.logStart
		if dataGrowth > maxDataGrowth {
			maxDataGrowth = dataGrowth
		}
		nodes[index] = capacitypolicy.NodeEvidence{
			Name:           measurement.name,
			DataBytesStart: measurement.dataStart, DataBytesPeak: measurement.dataPeak, DataGrowthBytes: dataGrowth,
			LogBytesStart: measurement.logStart, LogBytesPeak: measurement.logPeak, LogGrowthBytes: logGrowth,
			MaxRSSBytes: measurement.maxRSS,
		}
	}
	blockDelta := workloadEndHeight - startHeight
	if blockDelta <= 0 {
		t.Fatal("capacity workload made no block progress")
	}
	projected, ok := capacitypolicy.ProjectedGrowthBytes(maxDataGrowth, blockDelta, contract.Storage.ProjectionBlocks)
	if !ok {
		t.Fatal("capacity disk-growth projection overflowed")
	}
	evidence := capacitypolicy.Evidence{
		Version: capacitypolicy.EvidenceVersion, QualificationID: capacitypolicy.MaintainedQualificationID,
		Environment: capacitypolicy.MaintainedEnvironment,
		Workload: capacitypolicy.WorkloadEvidence{
			Validators: len(validators), TransactionsSubmitted: len(latencies), TransactionsCommitted: len(latencies),
			TransactionsFailed: 0, StartHeight: startHeight, EndHeight: workloadEndHeight, BlockDelta: blockDelta,
			DurationMS:         maxInt64(1, workloadDuration.Milliseconds()),
			ThroughputMilliTPS: int64(len(latencies)) * 1_000_000 / maxInt64(1, workloadDuration.Milliseconds()),
			AverageBlockTimeMS: maxInt64(1, workloadDuration.Milliseconds()) / blockDelta,
			P95CommitLatencyMS: maxInt64(1, latencies[p95Index].Milliseconds()),
			MaxCommitLatencyMS: maxInt64(1, latencies[len(latencies)-1].Milliseconds()),
		},
		Nodes:   nodes,
		Metrics: metrics,
		Retention: capacitypolicy.RetentionEvidence{
			Pruning: contract.Storage.Pruning, SnapshotInterval: contract.Storage.SnapshotInterval,
			SnapshotKeepRecent: contract.Storage.SnapshotKeepRecent, ObservedSnapshotCount: snapshotCount,
		},
		Consensus: capacitypolicy.ConsensusEvidence{
			CommonAppHash: true, ValidatorPowerConsistent: true, RestartVerified: true, LedgerValid: true,
		},
		Projection: capacitypolicy.ProjectionEvidence{Blocks: contract.Storage.ProjectionBlocks, MaxProjectedGrowthBytes: projected},
	}
	report := capacitypolicy.ValidateEvidence(contract, evidence)
	if !report.Valid {
		t.Fatalf("capacity evidence is invalid: %+v", report)
	}
	encoded, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if outputPath := os.Getenv(capacityEvidenceOutputEnv); outputPath != "" {
		file, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err != nil {
			t.Fatalf("create capacity evidence: output is unavailable")
		}
		if _, err := file.Write(append(encoded, '\n')); err != nil {
			_ = file.Close()
			t.Fatalf("write capacity evidence: output is unavailable")
		}
		if err := file.Close(); err != nil {
			t.Fatalf("close capacity evidence: output is unavailable")
		}
	}
	t.Logf("capacity_evidence=%s", encoded)
}

type capacityTxSubmission struct {
	hash    string
	started time.Time
	err     error
}

type capacityTxDelivery struct {
	latency time.Duration
	err     error
}

func submitCapacityTx(ctx context.Context, binary string, node *smokeValidator, from smokeAccount, chainID, domain, recipient string) capacityTxSubmission {
	started := time.Now()
	command := exec.CommandContext(ctx, binary,
		"tx", "truedemocracy", "add-member", domain, recipient,
		"--home", node.home,
		"--keyring-dir", from.keyringDir,
		"--keyring-backend", "test",
		"--from", from.name,
		"--chain-id", chainID,
		"--node", fmt.Sprintf("tcp://127.0.0.1:%d", node.rpcPort),
		"--offline",
		"--account-number", strconv.FormatUint(from.accountNumber, 10),
		"--sequence", strconv.FormatUint(from.sequence, 10),
		"--broadcast-mode", "sync",
		"--fees", "0"+token.BaseDenom,
		"--gas", "200000",
		"--yes",
		"--output", "json",
	)
	output, err := command.Output()
	if err != nil {
		return capacityTxSubmission{started: started, err: fmt.Errorf("authenticated transaction submission failed")}
	}
	var result struct {
		Code   uint32 `json:"code"`
		TxHash string `json:"txhash"`
	}
	if json.Unmarshal(output, &result) != nil || result.Code != 0 || result.TxHash == "" {
		return capacityTxSubmission{started: started, err: fmt.Errorf("authenticated transaction response is invalid")}
	}
	return capacityTxSubmission{hash: result.TxHash, started: started}
}

func waitForCapacityTx(ctx context.Context, node *smokeValidator, txHash string) error {
	deadline := time.Now().Add(45 * time.Second)
	for time.Now().Before(deadline) {
		code, _, _, err := querySmokeTx(ctx, node, txHash)
		if err == nil {
			if code != 0 {
				return fmt.Errorf("delivered transaction failed")
			}
			return nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	return fmt.Errorf("transaction was not indexed before the deadline")
}

type capacityNodeMeasurement struct {
	name      string
	dataStart int64
	dataPeak  int64
	logStart  int64
	logPeak   int64
	maxRSS    int64
}

func sampleCapacityNode(t *testing.T, validator *smokeValidator, measurement *capacityNodeMeasurement) {
	t.Helper()
	dataBytes := directoryRegularBytes(t, filepath.Join(validator.home, "data"))
	if dataBytes > measurement.dataPeak {
		measurement.dataPeak = dataBytes
	}
	logBytes := regularFileBytes(t, validator.logPath)
	if logBytes > measurement.logPeak {
		measurement.logPeak = logBytes
	}
	if validator.command == nil || validator.command.Process == nil {
		return
	}
	rss, err := processRSSBytes(validator.command.Process.Pid)
	if err != nil {
		t.Fatalf("sample %s resident memory: %v", validator.name, err)
	}
	if rss > measurement.maxRSS {
		measurement.maxRSS = rss
	}
}

func directoryRegularBytes(t *testing.T, root string) int64 {
	t.Helper()
	var total int64
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Snapshot retention may remove a completed snapshot after Walk has
			// listed it but before the entry is inspected. Keep the root strict
			// and tolerate only that expected child-disappearance race.
			if path != root && os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.Mode().IsRegular() {
			if info.Size() > 0 && total > int64(^uint64(0)>>1)-info.Size() {
				return fmt.Errorf("directory size overflow")
			}
			total += info.Size()
		}
		return nil
	})
	if err != nil {
		t.Fatalf("measure node data: %v", err)
	}
	return total
}

func regularFileBytes(t *testing.T, path string) int64 {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("measure node log: %v", err)
	}
	return info.Size()
}

func processRSSBytes(pid int) (int64, error) {
	if runtime.GOOS == "linux" {
		content, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
		if err != nil {
			return 0, err
		}
		scanner := bufio.NewScanner(bytes.NewReader(content))
		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) >= 2 && fields[0] == "VmRSS:" {
				kilobytes, err := strconv.ParseInt(fields[1], 10, 64)
				if err != nil || kilobytes <= 0 || kilobytes > int64(^uint64(0)>>1)/1024 {
					return 0, fmt.Errorf("invalid resident-memory sample")
				}
				return kilobytes * 1024, nil
			}
		}
		return 0, fmt.Errorf("resident-memory sample is missing")
	}
	output, err := exec.Command("ps", "-o", "rss=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return 0, err
	}
	kilobytes, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil || kilobytes <= 0 || kilobytes > int64(^uint64(0)>>1)/1024 {
		return 0, fmt.Errorf("invalid resident-memory sample")
	}
	return kilobytes * 1024, nil
}

func configureCapacityPrometheus(t *testing.T, configPath string, port int) {
	t.Helper()
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	updated := replaceTomlSectionValue(t, string(content), "[instrumentation]", "prometheus = false", "prometheus = true")
	updated = replaceTomlSectionValue(t, updated, "[instrumentation]", `prometheus_listen_addr = ":26660"`, fmt.Sprintf(`prometheus_listen_addr = "127.0.0.1:%d"`, port))
	if err := atomicWriteFile(configPath, []byte(updated), 0o600); err != nil {
		t.Fatal(err)
	}
}

func waitForCapacityMetrics(t *testing.T, appURL, cometURL string, minimumHeight int64) capacitypolicy.MetricsEvidence {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	client := &http.Client{Timeout: 2 * time.Second}
	var lastErr error
	for time.Now().Before(deadline) {
		appBody, err := readBoundedURL(t.Context(), client, appURL)
		if err != nil {
			lastErr = err
			time.Sleep(250 * time.Millisecond)
			continue
		}
		cometBody, err := readBoundedURL(t.Context(), client, cometURL)
		if err != nil {
			lastErr = err
			time.Sleep(250 * time.Millisecond)
			continue
		}
		readApp := func(name string) (int64, error) { return parseIntegralMetric(appBody, name, false) }
		readComet := func(name string) (int64, error) { return parseIntegralMetric(cometBody, name, true) }
		consensusHeight, err := readComet("cometbft_consensus_height")
		if err != nil {
			lastErr = err
			continue
		}
		applicationHeight, err := readApp("truerepublic_app_last_successful_block_height")
		if err != nil {
			lastErr = err
			continue
		}
		invariantHeight, err := readApp("truerepublic_app_last_successful_invariant_cycle_height")
		if err != nil {
			lastErr = err
			continue
		}
		completedBlocks, err := readApp("truerepublic_app_completed_blocks_total")
		if err != nil {
			lastErr = err
			continue
		}
		supply, err := readApp("truerepublic_token_pnyx_supply_base_units")
		if err != nil {
			lastErr = err
			continue
		}
		headroom, err := readApp("truerepublic_token_pnyx_supply_headroom_base_units")
		if err != nil {
			lastErr = err
			continue
		}
		goroutines, err := readApp("go_goroutines")
		if err != nil {
			lastErr = err
			continue
		}
		rss, err := readApp("process_resident_memory_bytes")
		if err != nil {
			lastErr = err
			continue
		}
		if consensusHeight >= minimumHeight && applicationHeight >= minimumHeight && invariantHeight == applicationHeight {
			return capacitypolicy.MetricsEvidence{
				ConsensusHeight: consensusHeight, ApplicationHeight: applicationHeight, InvariantHeight: invariantHeight,
				CompletedBlocks: completedBlocks, SupplyBaseUnits: supply, SupplyHeadroomUnits: headroom,
				Goroutines: goroutines, ResidentMemoryBytes: rss,
			}
		}
		lastErr = fmt.Errorf("metrics have not reached the final height")
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("capacity metrics did not converge: %v", lastErr)
	return capacitypolicy.MetricsEvidence{}
}

func readBoundedURL(ctx context.Context, client *http.Client, url string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("metrics endpoint returned non-success status")
	}
	const maximumMetricsBytes = 2 << 20
	body, err := io.ReadAll(io.LimitReader(response.Body, maximumMetricsBytes+1))
	if err != nil {
		return "", err
	}
	if len(body) > maximumMetricsBytes {
		return "", fmt.Errorf("metrics response exceeds the maintained size limit")
	}
	return string(body), nil
}

func parseIntegralMetric(body, name string, allowLabels bool) (int64, error) {
	scanner := bufio.NewScanner(strings.NewReader(body))
	found := false
	value := int64(0)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 {
			continue
		}
		familyMatches := fields[0] == name
		if allowLabels && strings.HasPrefix(fields[0], name+"{") && strings.HasSuffix(fields[0], "}") {
			familyMatches = true
		}
		if !familyMatches {
			continue
		}
		if found {
			return 0, fmt.Errorf("required metric %s has multiple series", name)
		}
		parsed, err := strconv.ParseFloat(fields[1], 64)
		if err != nil || parsed < 0 || parsed > float64(int64(^uint64(0)>>1)) || parsed != float64(int64(parsed)) {
			return 0, fmt.Errorf("metric is not a bounded integer")
		}
		found = true
		value = int64(parsed)
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	if found {
		return value, nil
	}
	return 0, fmt.Errorf("required metric %s is missing", name)
}

func TestCapacityMetricParsingIsExactAndBounded(t *testing.T) {
	t.Run("exact unlabelled metric", func(t *testing.T) {
		value, err := parseIntegralMetric("required_metric 42\nrequired_metric_suffix 99\n", "required_metric", false)
		if err != nil || value != 42 {
			t.Fatalf("parse exact metric: value=%d err=%v", value, err)
		}
	})
	t.Run("labelled sibling is not substituted", func(t *testing.T) {
		if _, err := parseIntegralMetric(`required_metric{node="other"} 42`, "required_metric", false); err == nil {
			t.Fatal("expected exact-name failure")
		}
	})
	t.Run("one labelled consensus series", func(t *testing.T) {
		value, err := parseIntegralMetric(`required_metric{chain_id="synthetic"} 42`, "required_metric", true)
		if err != nil || value != 42 {
			t.Fatalf("parse labelled metric: value=%d err=%v", value, err)
		}
	})
	t.Run("multiple labelled consensus series fail", func(t *testing.T) {
		body := "required_metric{chain_id=\"one\"} 1\nrequired_metric{chain_id=\"two\"} 2\n"
		if _, err := parseIntegralMetric(body, "required_metric", true); err == nil {
			t.Fatal("expected multiple-series failure")
		}
	})
	t.Run("fraction and negative values fail", func(t *testing.T) {
		for _, body := range []string{"required_metric 1.5", "required_metric -1"} {
			if _, err := parseIntegralMetric(body, "required_metric", false); err == nil {
				t.Fatalf("expected bounded-integer failure for %q", body)
			}
		}
	})
}

func TestCapacityMetricsResponseLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write(bytes.Repeat([]byte("x"), (2<<20)+1))
	}))
	t.Cleanup(server.Close)
	client := &http.Client{Timeout: 2 * time.Second}
	if _, err := readBoundedURL(t.Context(), client, server.URL); err == nil || !strings.Contains(err.Error(), "size limit") {
		t.Fatalf("expected response-size rejection, got %v", err)
	}
}

func TestCapacitySnapshotStoreCounting(t *testing.T) {
	snapshotDir := t.TempDir()
	for _, name := range []string{"metadata.db", "10", "12", "14"} {
		if err := os.Mkdir(filepath.Join(snapshotDir, name), 0o700); err != nil {
			t.Fatalf("create snapshot-store entry: %v", err)
		}
	}
	count, err := countSnapshotHeightDirectories(snapshotDir)
	if err != nil || count != 3 {
		t.Fatalf("count retained snapshot heights: count=%d err=%v", count, err)
	}
	if err := os.Mkdir(filepath.Join(snapshotDir, "unexpected"), 0o700); err != nil {
		t.Fatalf("create invalid snapshot-store entry: %v", err)
	}
	if _, err := countSnapshotHeightDirectories(snapshotDir); err == nil {
		t.Fatal("expected an unexpected snapshot-store entry to fail closed")
	}
}

func waitForSnapshotCount(t *testing.T, home string, maximum int, timeout time.Duration) int {
	t.Helper()
	deadline := time.Now().Add(timeout)
	lastCount := -1
	lastErr := error(nil)
	for time.Now().Before(deadline) {
		count, err := countSnapshotHeightDirectories(filepath.Join(home, "data", "snapshots"))
		lastErr = err
		if err == nil {
			lastCount = count
			if count > 0 && count <= maximum {
				return count
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("snapshot retention did not settle within 1..%d snapshots (last observed count: %d, last error: %v)", maximum, lastCount, lastErr)
	return 0
}

func countSnapshotHeightDirectories(snapshotDir string) (int, error) {
	entries, err := os.ReadDir(snapshotDir)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, entry := range entries {
		if entry.Name() == "metadata.db" {
			continue
		}
		if !entry.IsDir() {
			return 0, fmt.Errorf("unexpected snapshot-store entry")
		}
		height, err := strconv.ParseUint(entry.Name(), 10, 64)
		if err != nil || height == 0 {
			return 0, fmt.Errorf("unexpected snapshot height directory")
		}
		count++
	}
	return count, nil
}

func maxInt64(left, right int64) int64 {
	if left > right {
		return left
	}
	return right
}
