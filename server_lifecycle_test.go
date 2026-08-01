package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkversion "github.com/cosmos/cosmos-sdk/version"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"

	"truerepublic/token"
	"truerepublic/x/truedemocracy"
)

func TestRootUsesStandardCosmosServerCommands(t *testing.T) {
	root := newRootCmd()
	if root.Version != version {
		t.Fatalf("root version = %q, want injected version %q", root.Version, version)
	}
	if sdkversion.Name != "TrueRepublic" || sdkversion.AppName != "truerepublicd" || sdkversion.Version != version {
		t.Fatalf("SDK version metadata = (%q, %q, %q), want TrueRepublic/truerepublicd/%s", sdkversion.Name, sdkversion.AppName, sdkversion.Version, version)
	}
	for _, path := range []string{
		"init", "start", "export", "comet", "keys",
		"network-policy", "topology-policy", "incident-rehearsal", "capacity-policy", "healthcheck",
	} {
		cmd, _, err := root.Find([]string{path})
		if err != nil || cmd == root {
			t.Fatalf("standard server command %q is not registered", path)
		}
	}
	start, _, err := root.Find([]string{"start"})
	if err != nil {
		t.Fatal(err)
	}
	for _, flag := range []string{"home", "db_backend", "with-comet", "shutdown-grace"} {
		if start.Flags().Lookup(flag) == nil {
			t.Fatalf("standard start flag %q is missing", flag)
		}
	}
	initCmd, _, err := root.Find([]string{"init"})
	if err != nil {
		t.Fatal(err)
	}
	if got := initCmd.Flags().Lookup("default-denom").DefValue; got != token.BaseDenom {
		t.Fatalf("init default denom = %q, want %q", got, token.BaseDenom)
	}
	if initCmd.Flags().Lookup("bootstrap-operator") == nil {
		t.Fatal("init bootstrap-operator flag is missing")
	}
}

func TestRootUsesProcessOutputStreams(t *testing.T) {
	root := newRootCmd()
	if root.OutOrStdout() != os.Stdout {
		t.Fatal("root command output must use stdout for machine-readable pipelines")
	}
	if root.ErrOrStderr() != os.Stderr {
		t.Fatal("root command errors must use stderr")
	}
}

func TestStartRequiresStructuredLogFormat(t *testing.T) {
	for _, environmentName := range structuredLogEnvironmentNames() {
		t.Setenv(environmentName, "")
	}

	t.Run("default becomes JSON", func(t *testing.T) {
		cmd := &cobra.Command{Use: "start"}
		cmd.Flags().String(flags.FlagLogFormat, "plain", "")
		if err := requireStructuredStartLogging(cmd); err != nil {
			t.Fatal(err)
		}
		if got, err := cmd.Flags().GetString(flags.FlagLogFormat); err != nil || got != flags.OutputFormatJSON {
			t.Fatalf("log format = %q, %v; want %q", got, err, flags.OutputFormatJSON)
		}
	})

	t.Run("explicit plain fails closed", func(t *testing.T) {
		cmd := &cobra.Command{Use: "start"}
		cmd.Flags().String(flags.FlagLogFormat, "plain", "")
		if err := cmd.Flags().Set(flags.FlagLogFormat, "plain"); err != nil {
			t.Fatal(err)
		}
		err := requireStructuredStartLogging(cmd)
		if err == nil || !strings.Contains(err.Error(), `requires "json" structured logs`) {
			t.Fatalf("plain log format error = %v", err)
		}
	})

	t.Run("effective environment format fails closed", func(t *testing.T) {
		cmd := &cobra.Command{Use: "start"}
		err := validateStructuredStartLogFormat(cmd, "plain")
		if err == nil || !strings.Contains(err.Error(), `requires "json" structured logs`) {
			t.Fatalf("effective plain log format error = %v", err)
		}
	})

	t.Run("plain environment override fails closed", func(t *testing.T) {
		t.Setenv(envPrefix+"_LOG_FORMAT", "plain")
		cmd := &cobra.Command{Use: "start"}
		cmd.Flags().String(flags.FlagLogFormat, "plain", "")
		err := requireStructuredStartLogging(cmd)
		if err == nil || !strings.Contains(err.Error(), envPrefix+"_LOG_FORMAT") {
			t.Fatalf("environment log format error = %v", err)
		}
	})

	t.Run("raw trace store fails closed", func(t *testing.T) {
		cmd := &cobra.Command{Use: "start"}
		cmd.Flags().String(flags.FlagLogFormat, "plain", "")
		cmd.Flags().String(traceStoreFlag, "", "")
		if err := cmd.Flags().Set(traceStoreFlag, "/tmp/raw-kv.log"); err != nil {
			t.Fatal(err)
		}
		err := requireStructuredStartLogging(cmd)
		if err == nil || !strings.Contains(err.Error(), "bypasses the structured logging boundary") {
			t.Fatalf("trace-store error = %v", err)
		}
	})

	t.Run("effective raw trace store fails closed", func(t *testing.T) {
		cmd := &cobra.Command{Use: "start"}
		err := validateStructuredStartTraceStore(cmd, "/tmp/raw-kv.log")
		if err == nil || !strings.Contains(err.Error(), "bypasses the structured logging boundary") {
			t.Fatalf("effective trace-store error = %v", err)
		}
	})

	t.Run("non-start command is unchanged", func(t *testing.T) {
		cmd := &cobra.Command{Use: "export"}
		if err := requireStructuredStartLogging(cmd); err != nil {
			t.Fatal(err)
		}
	})
}

func TestNodeStartsStopsAndRestartsFromPersistentHome(t *testing.T) {
	for _, environmentName := range structuredLogEnvironmentNames() {
		t.Setenv(environmentName, "")
	}
	t.Setenv("TRUEREPUBLICD_TRACE_STORE", "")
	t.Setenv("truerepublicd_TRACE_STORE", "")

	binary := filepath.Join(t.TempDir(), "truerepublicd")
	build := exec.Command("go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build daemon: %v\n%s", err, output)
	}
	home := filepath.Join(t.TempDir(), "node")
	for _, test := range []struct {
		name        string
		command     *exec.Cmd
		wantMessage string
	}{
		{
			name:        "explicit plain",
			command:     exec.Command(binary, "start", "--home", home, "--log_format=plain"),
			wantMessage: `requires "json" structured logs`,
		},
		{
			name: "environment plain",
			command: func() *exec.Cmd {
				cmd := exec.Command(binary, "start", "--home", home)
				cmd.Env = append(os.Environ(), "TRUEREPUBLIC_LOG_FORMAT=plain")
				return cmd
			}(),
			wantMessage: `requires "json" structured logs`,
		},
		{
			name: "raw trace store",
			command: exec.Command(
				binary,
				"start",
				"--home", home,
				"--trace-store", filepath.Join(t.TempDir(), "raw-kv.log"),
			),
			wantMessage: "bypasses the structured logging boundary",
		},
		{
			name: "environment raw trace store",
			command: func() *exec.Cmd {
				cmd := exec.Command(binary, "start", "--home", home)
				cmd.Env = append(os.Environ(), "TRUEREPUBLICD_TRACE_STORE="+filepath.Join(t.TempDir(), "raw-kv.log"))
				return cmd
			}(),
			wantMessage: "bypasses the structured logging boundary",
		},
	} {
		output, err := test.command.CombinedOutput()
		if err == nil || !strings.Contains(string(output), test.wantMessage) {
			t.Fatalf("%s start was not rejected before opening state: err=%v\n%s", test.name, err, output)
		}
	}
	operatorAddr := "truerepublic13hgqwy9986x5nk6jt23ns5v7j0acs8qmhchhtw"
	initCmd := exec.Command(binary, "init", "restart-node", "--chain-id", "truerepublic-restart-1", "--home", home, "--bootstrap-operator", operatorAddr)
	if output, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("init node: %v\n%s", err, output)
	}
	initialized, err := genutiltypes.AppGenesisFromFile(filepath.Join(home, "config", "genesis.json"))
	if err != nil {
		t.Fatal(err)
	}
	var initializedState map[string]json.RawMessage
	if err := json.Unmarshal(initialized.AppState, &initializedState); err != nil {
		t.Fatal(err)
	}
	var crisisGenesis crisistypes.GenesisState
	if err := json.Unmarshal(initializedState[crisistypes.ModuleName], &crisisGenesis); err != nil {
		t.Fatal(err)
	}
	if crisisGenesis.ConstantFee.Denom != token.BaseDenom {
		t.Fatalf("crisis fee denom = %q, want %q", crisisGenesis.ConstantFee.Denom, token.BaseDenom)
	}
	keyCmd := exec.Command(binary, "keys", "add", "smoke", "--keyring-backend", "test", "--home", home, "--output", "json")
	if output, err := keyCmd.CombinedOutput(); err != nil {
		t.Fatalf("create smoke key: %v\n%s", err, output)
	}

	rpcPort := freeTCPPort(t)
	p2pPort := freeTCPPort(t)
	apiPort := freeTCPPort(t)
	rpcURL := fmt.Sprintf("http://127.0.0.1:%d/status", rpcPort)
	metricsURL := fmt.Sprintf("http://127.0.0.1:%d/metrics?format=prometheus", apiPort)
	start := func(telemetryEnabled bool) (*exec.Cmd, *os.File) {
		t.Helper()
		logFile, err := os.CreateTemp(t.TempDir(), "node-*.log")
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(binary,
			"start",
			"--home", home,
			"--rpc.laddr", fmt.Sprintf("tcp://127.0.0.1:%d", rpcPort),
			"--rpc.pprof_laddr", "",
			"--p2p.laddr", fmt.Sprintf("tcp://127.0.0.1:%d", p2pPort),
			"--grpc.enable=false",
			"--api.enable=true",
			"--api.address", fmt.Sprintf("tcp://127.0.0.1:%d", apiPort),
			"--minimum-gas-prices", "0"+token.BaseDenom,
		)
		telemetryValue := strconv.FormatBool(telemetryEnabled)
		retention := "0"
		if telemetryEnabled {
			retention = "60"
		}
		cmd.Env = append(os.Environ(),
			"TRUEREPUBLICD_TELEMETRY_ENABLED="+telemetryValue,
			"TRUEREPUBLICD_TELEMETRY_PROMETHEUS_RETENTION_TIME="+retention,
			"TRUEREPUBLICD_TELEMETRY_SERVICE_NAME=truerepublic",
			"TRUEREPUBLICD_TELEMETRY_ENABLE_HOSTNAME=false",
			"TRUEREPUBLICD_TELEMETRY_ENABLE_HOSTNAME_LABEL=false",
			"TRUEREPUBLICD_TELEMETRY_ENABLE_SERVICE_LABEL=false",
		)
		cmd.Stdout = logFile
		cmd.Stderr = logFile
		if err := cmd.Start(); err != nil {
			_ = logFile.Close()
			t.Fatal(err)
		}
		return cmd, logFile
	}
	stop := func(cmd *exec.Cmd, logFile *os.File) {
		t.Helper()
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			_ = cmd.Process.Kill()
		}
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()
		select {
		case err := <-done:
			if err != nil {
				_ = logFile.Close()
				content, _ := os.ReadFile(logFile.Name())
				t.Fatalf("node did not stop cleanly: %v\n%s", err, content)
			}
		case <-time.After(15 * time.Second):
			_ = cmd.Process.Kill()
			_ = logFile.Close()
			t.Fatal("node did not stop within 15 seconds")
		}
		_ = logFile.Close()
	}

	first, firstLog := start(true)
	firstHeight := waitForNodeHeight(t, rpcURL, 1, first, firstLog)
	firstMetrics := waitForApplicationMetrics(t, metricsURL, firstHeight, first, firstLog)
	firstLogPath := firstLog.Name()
	stop(first, firstLog)
	assertStructuredNodeLogs(t, firstLogPath)

	second, secondLog := start(true)
	secondHeight := waitForNodeHeight(t, rpcURL, firstHeight+1, second, secondLog)
	secondMetrics := waitForApplicationMetrics(t, metricsURL, secondHeight, second, secondLog)
	secondLogPath := secondLog.Name()
	stop(second, secondLog)
	assertStructuredNodeLogs(t, secondLogPath)
	if secondHeight <= firstHeight {
		t.Fatalf("restart did not advance height: first=%d second=%d", firstHeight, secondHeight)
	}
	if secondMetrics.blockHeight <= firstMetrics.blockHeight {
		t.Fatalf(
			"application metrics did not advance after restart: first=%v second=%v",
			firstMetrics.blockHeight,
			secondMetrics.blockHeight,
		)
	}

	disabled, disabledLog := start(false)
	disabledHeight := waitForNodeHeight(t, rpcURL, secondHeight+1, disabled, disabledLog)
	waitForApplicationMetricsDisabled(t, metricsURL, disabled, disabledLog)
	disabledLogPath := disabledLog.Name()
	stop(disabled, disabledLog)
	assertStructuredNodeLogs(t, disabledLogPath)

	exportCmd := exec.Command(binary, "export", "--home", home)
	exported, err := exportCmd.Output()
	if err != nil {
		t.Fatalf("export persistent state: %v", err)
	}
	var exportedGenesis struct {
		InitialHeight int64                      `json:"initial_height"`
		AppState      map[string]json.RawMessage `json:"app_state"`
	}
	if err := json.Unmarshal(exported, &exportedGenesis); err != nil {
		t.Fatalf("decode exported genesis: %v", err)
	}
	if exportedGenesis.InitialHeight <= disabledHeight {
		t.Fatalf("exported initial height = %d, want greater than committed height %d", exportedGenesis.InitialHeight, disabledHeight)
	}
	if exportedGenesis.AppState[truedemocracy.ModuleName] == nil {
		t.Fatal("exported persistent state is missing truedemocracy genesis")
	}
}

type applicationMetricsSnapshot struct {
	blockHeight     float64
	invariantHeight float64
	completedBlocks float64
	supply          float64
	headroom        float64
}

func waitForApplicationMetrics(
	t *testing.T,
	url string,
	minimumHeight int64,
	cmd *exec.Cmd,
	logFile *os.File,
) applicationMetricsSnapshot {
	t.Helper()
	client := &http.Client{Timeout: time.Second}
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		response, err := client.Get(url)
		if err == nil {
			body, readErr := io.ReadAll(io.LimitReader(response.Body, 2<<20))
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK && readErr == nil {
				snapshot, parseErr := parseApplicationMetrics(string(body))
				if parseErr == nil &&
					snapshot.blockHeight >= float64(minimumHeight) &&
					snapshot.invariantHeight == snapshot.blockHeight &&
					snapshot.completedBlocks >= 1 &&
					snapshot.supply >= 0 &&
					snapshot.headroom >= 0 &&
					snapshot.supply+snapshot.headroom == float64(token.MaxSupplyBaseUnits) {
					return snapshot
				}
			}
		}
		if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	_ = cmd.Process.Kill()
	_ = cmd.Wait()
	_ = logFile.Close()
	content, _ := os.ReadFile(logFile.Name())
	t.Fatalf("application metrics did not reach height %d\n%s", minimumHeight, content)
	return applicationMetricsSnapshot{}
}

func parseApplicationMetrics(body string) (applicationMetricsSnapshot, error) {
	read := func(name string) (float64, error) {
		scanner := bufio.NewScanner(strings.NewReader(body))
		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) == 2 && fields[0] == name {
				return strconv.ParseFloat(fields[1], 64)
			}
		}
		if err := scanner.Err(); err != nil {
			return 0, err
		}
		return 0, fmt.Errorf("metric %s is missing", name)
	}
	for _, required := range []string{
		"go_goroutines",
		"process_resident_memory_bytes",
	} {
		if _, err := read(required); err != nil {
			return applicationMetricsSnapshot{}, fmt.Errorf("required SDK/runtime metric: %w", err)
		}
	}
	var snapshot applicationMetricsSnapshot
	values := []struct {
		name   string
		target *float64
	}{
		{"truerepublic_app_last_successful_block_height", &snapshot.blockHeight},
		{"truerepublic_app_last_successful_invariant_cycle_height", &snapshot.invariantHeight},
		{"truerepublic_app_completed_blocks_total", &snapshot.completedBlocks},
		{"truerepublic_token_pnyx_supply_base_units", &snapshot.supply},
		{"truerepublic_token_pnyx_supply_headroom_base_units", &snapshot.headroom},
	}
	for _, value := range values {
		parsed, err := read(value.name)
		if err != nil {
			return applicationMetricsSnapshot{}, err
		}
		*value.target = parsed
	}
	return snapshot, nil
}

func TestParseApplicationMetricsRequiresExactRuntimeSamples(t *testing.T) {
	body := `# HELP go_goroutines Number of goroutines.
go_goroutines_extra 7
process_resident_memory_bytes_total 1024
truerepublic_app_last_successful_block_height 1
truerepublic_app_last_successful_invariant_cycle_height 1
truerepublic_app_completed_blocks_total 1
truerepublic_token_pnyx_supply_base_units 0
truerepublic_token_pnyx_supply_headroom_base_units 21000000000000
`
	if _, err := parseApplicationMetrics(body); err == nil {
		t.Fatal("metric comments and sibling names must not satisfy exact runtime samples")
	}
}

func waitForApplicationMetricsDisabled(t *testing.T, url string, cmd *exec.Cmd, logFile *os.File) {
	t.Helper()
	client := &http.Client{Timeout: time.Second}
	deadline := time.Now().Add(10 * time.Second)
	exposed := false
	lastStatus := 0
	var lastErr error
	for time.Now().Before(deadline) {
		response, err := client.Get(url)
		if err == nil {
			lastStatus = response.StatusCode
			lastErr = nil
			_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 64<<10))
			_ = response.Body.Close()
			if response.StatusCode == http.StatusNotImplemented {
				return
			}
			if response.StatusCode == http.StatusOK {
				exposed = true
				break
			}
		} else {
			lastErr = err
		}
		if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	_ = cmd.Process.Kill()
	_ = cmd.Wait()
	_ = logFile.Close()
	content, _ := os.ReadFile(logFile.Name())
	if exposed {
		t.Fatalf("disabled application telemetry exposed the metrics endpoint\n%s", content)
	}
	t.Fatalf(
		"disabled application telemetry did not settle on a 501 endpoint (last status=%d, last error=%v)\n%s",
		lastStatus,
		lastErr,
		content,
	)
}

func assertStructuredNodeLogs(t *testing.T, path string) {
	t.Helper()
	logFile, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer logFile.Close()

	scanner := bufio.NewScanner(logFile)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	lines := 0
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		lines++
		var event map[string]any
		if err := json.Unmarshal(line, &event); err != nil {
			t.Fatalf("node log line %d is not structured JSON: %v\n%s", lines, err, line)
		}
		if event["level"] == nil || event["message"] == nil {
			t.Fatalf("node log line %d lacks level/message fields: %s", lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if lines == 0 {
		t.Fatal("node emitted no structured log lines")
	}
}

func freeTCPPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func waitForNodeHeight(t *testing.T, url string, minimum int64, cmd *exec.Cmd, logFile *os.File) int64 {
	t.Helper()
	client := &http.Client{Timeout: time.Second}
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		response, err := client.Get(url)
		if err == nil {
			var status struct {
				Result struct {
					SyncInfo struct {
						LatestBlockHeight string `json:"latest_block_height"`
					} `json:"sync_info"`
				} `json:"result"`
			}
			decodeErr := json.NewDecoder(response.Body).Decode(&status)
			_ = response.Body.Close()
			if decodeErr == nil {
				height, parseErr := strconv.ParseInt(status.Result.SyncInfo.LatestBlockHeight, 10, 64)
				if parseErr == nil && height >= minimum {
					return height
				}
			}
		}
		if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	_ = cmd.Process.Kill()
	_ = cmd.Wait()
	_ = logFile.Close()
	content, _ := os.ReadFile(logFile.Name())
	t.Fatalf("node did not reach height %d\n%s", minimum, content)
	return 0
}

func TestBindGenesisValidatorKeyUsesGeneratedNodeKey(t *testing.T) {
	app := newGenesisTestApp(t)
	appState := ModuleBasics.DefaultGenesis(app.appCodec)
	appStateJSON, err := json.Marshal(appState)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "genesis.json")
	genesisDoc := &genutiltypes.AppGenesis{
		ChainID: "test-bind-chain", AppState: appStateJSON, Consensus: &genutiltypes.ConsensusGenesis{},
	}
	if err := genesisDoc.SaveAs(path); err != nil {
		t.Fatal(err)
	}

	generatedPubKey := bytes.Repeat([]byte{0x42}, 32)
	operatorAddr := sdk.AccAddress(bytes.Repeat([]byte{0x24}, 20)).String()
	if err := bindGenesisValidatorKey(path, generatedPubKey, operatorAddr); err != nil {
		t.Fatal(err)
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var genesis map[string]json.RawMessage
	if err := json.Unmarshal(updated, &genesis); err != nil {
		t.Fatal(err)
	}
	var updatedState map[string]json.RawMessage
	if err := json.Unmarshal(genesis["app_state"], &updatedState); err != nil {
		t.Fatal(err)
	}
	var tdGenesis truedemocracy.GenesisState
	if err := json.Unmarshal(updatedState[truedemocracy.ModuleName], &tdGenesis); err != nil {
		t.Fatal(err)
	}
	if got := tdGenesis.Validators[0].PubKey; !bytes.Equal(got, generatedPubKey) {
		t.Fatalf("bootstrap pubkey = %x, want generated key %x", got, generatedPubKey)
	}
	if got := tdGenesis.Validators[0].OperatorAddr; got != operatorAddr {
		t.Fatalf("bootstrap operator = %q, want independent account %q", got, operatorAddr)
	}
	if len(tdGenesis.Domains) != 1 || len(tdGenesis.Validators) != 1 {
		t.Fatalf("unexpected PoD bootstrap state: %+v", tdGenesis)
	}
	authGenesis := authtypes.GetGenesisStateFromAppState(app.appCodec, updatedState)
	accounts, err := authtypes.UnpackAccounts(authGenesis.Accounts)
	if err != nil {
		t.Fatal(err)
	}
	foundOperatorAccount := false
	for _, account := range accounts {
		if account.GetAddress().String() == operatorAddr {
			foundOperatorAccount = true
			break
		}
	}
	if !foundOperatorAccount {
		t.Fatal("bootstrap operator authority is missing from auth genesis")
	}
	bankGenesis := banktypes.GetGenesisStateFromAppState(app.appCodec, updatedState)
	moduleAddress := authtypes.NewModuleAddress(truedemocracy.ModuleName).String()
	wantStake := tdGenesis.Validators[0].Stake
	var backedStake int64
	for _, balance := range bankGenesis.Balances {
		if balance.Address == moduleAddress {
			backedStake = balance.Coins.AmountOf(token.BaseDenom).Int64()
			break
		}
	}
	if backedStake != wantStake {
		t.Fatalf("module stake backing = %d, want %d", backedStake, wantStake)
	}
	if err := validateLedgerGenesis(app.appCodec, updatedState); err != nil {
		t.Fatalf("generated genesis is not ledger-backed: %v", err)
	}
	cometGenesis, err := genutiltypes.AppGenesisFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cometGenesis.Consensus.Validators) != 1 || !bytes.Equal(cometGenesis.Consensus.Validators[0].PubKey.Bytes(), generatedPubKey) {
		t.Fatalf("CometBFT validator set does not use generated key: %+v", cometGenesis.Consensus.Validators)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("genesis mode = %o, want 600", info.Mode().Perm())
	}
}

func TestBindGenesisValidatorKeyRejectsConsensusDerivedOperator(t *testing.T) {
	appState := ModuleBasics.DefaultGenesis(newGenesisTestApp(t).appCodec)
	appStateJSON, err := json.Marshal(appState)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "genesis.json")
	genesisDoc := &genutiltypes.AppGenesis{
		ChainID: "test-coupled-operator-chain", AppState: appStateJSON, Consensus: &genutiltypes.ConsensusGenesis{},
	}
	if err := genesisDoc.SaveAs(path); err != nil {
		t.Fatal(err)
	}
	pubKey := bytes.Repeat([]byte{0x43}, 32)
	coupledOperator := sdk.AccAddress(cmted25519.PubKey(pubKey).Address()).String()
	before, _ := os.ReadFile(path)
	if err := bindGenesisValidatorKey(path, pubKey, coupledOperator); err == nil {
		t.Fatal("consensus-derived operator authority accepted")
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(after, before) {
		t.Fatal("rejected coupled operator mutated genesis")
	}
}

func TestBindGenesisValidatorKeyRejectsReservedModuleOperator(t *testing.T) {
	appState := ModuleBasics.DefaultGenesis(newGenesisTestApp(t).appCodec)
	appStateJSON, err := json.Marshal(appState)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "genesis.json")
	genesisDoc := &genutiltypes.AppGenesis{
		ChainID: "test-module-operator-chain", AppState: appStateJSON, Consensus: &genutiltypes.ConsensusGenesis{},
	}
	if err := genesisDoc.SaveAs(path); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(path)
	moduleOperator := authtypes.NewModuleAddress(truedemocracy.ModuleName).String()
	if err := bindGenesisValidatorKey(path, bytes.Repeat([]byte{0x44}, 32), moduleOperator); err == nil {
		t.Fatal("reserved module account accepted as bootstrap operator")
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(after, before) {
		t.Fatal("rejected module operator mutated genesis")
	}
}

func TestBindGenesisValidatorKeyRejectsInvalidKeyWithoutMutation(t *testing.T) {
	appState := ModuleBasics.DefaultGenesis(newGenesisTestApp(t).appCodec)
	appStateJSON, err := json.Marshal(appState)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "genesis.json")
	genesisDoc := &genutiltypes.AppGenesis{
		ChainID: "test-invalid-key-chain", AppState: appStateJSON, Consensus: &genutiltypes.ConsensusGenesis{},
	}
	if err := genesisDoc.SaveAs(path); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(path)
	if err := bindGenesisValidatorKey(path, []byte{1}, sdk.AccAddress(bytes.Repeat([]byte{0x25}, 20)).String()); err == nil {
		t.Fatal("invalid consensus key accepted")
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(after, before) {
		t.Fatal("invalid key mutated genesis")
	}
}

func TestBindGenesisValidatorKeyRefusesExistingConsensusSet(t *testing.T) {
	appState := ModuleBasics.DefaultGenesis(newGenesisTestApp(t).appCodec)
	appStateJSON, err := json.Marshal(appState)
	if err != nil {
		t.Fatal(err)
	}
	existingKey := bytes.Repeat([]byte{0x11}, 32)
	genesisDoc := &genutiltypes.AppGenesis{
		ChainID: "existing-validator-chain", AppState: appStateJSON,
		Consensus: &genutiltypes.ConsensusGenesis{Validators: []cmttypes.GenesisValidator{{
			PubKey: cmted25519.PubKey(existingKey), Power: 1,
		}}},
	}
	path := filepath.Join(t.TempDir(), "genesis.json")
	if err := genesisDoc.SaveAs(path); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(path)
	if err := bindGenesisValidatorKey(path, bytes.Repeat([]byte{0x22}, 32), sdk.AccAddress(bytes.Repeat([]byte{0x26}, 20)).String()); err == nil {
		t.Fatal("existing consensus validator set was replaced")
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(after, before) {
		t.Fatal("rejected existing validator set mutated genesis")
	}
}
