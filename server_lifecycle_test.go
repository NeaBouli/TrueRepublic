package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"truerepublic/token"
	"truerepublic/x/truedemocracy"
)

func TestRootUsesStandardCosmosServerCommands(t *testing.T) {
	root := newRootCmd()
	for _, path := range []string{"init", "start", "export", "comet", "keys"} {
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
}

func TestNodeStartsStopsAndRestartsFromPersistentHome(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "truerepublicd")
	build := exec.Command("go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build daemon: %v\n%s", err, output)
	}
	home := filepath.Join(t.TempDir(), "node")
	initCmd := exec.Command(binary, "init", "restart-node", "--chain-id", "truerepublic-restart-1", "--home", home)
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
	rpcURL := fmt.Sprintf("http://127.0.0.1:%d/status", rpcPort)
	start := func() (*exec.Cmd, *os.File) {
		t.Helper()
		logFile, err := os.CreateTemp(t.TempDir(), "node-*.log")
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(binary,
			"start",
			"--home", home,
			"--rpc.laddr", fmt.Sprintf("tcp://127.0.0.1:%d", rpcPort),
			"--p2p.laddr", fmt.Sprintf("tcp://127.0.0.1:%d", p2pPort),
			"--grpc.enable=false",
			"--api.enable=false",
			"--minimum-gas-prices", "0"+token.BaseDenom,
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

	first, firstLog := start()
	firstHeight := waitForNodeHeight(t, rpcURL, 1, first, firstLog)
	stop(first, firstLog)

	second, secondLog := start()
	secondHeight := waitForNodeHeight(t, rpcURL, firstHeight+1, second, secondLog)
	stop(second, secondLog)
	if secondHeight <= firstHeight {
		t.Fatalf("restart did not advance height: first=%d second=%d", firstHeight, secondHeight)
	}
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
	if exportedGenesis.InitialHeight <= secondHeight {
		t.Fatalf("exported initial height = %d, want greater than committed height %d", exportedGenesis.InitialHeight, secondHeight)
	}
	if exportedGenesis.AppState[truedemocracy.ModuleName] == nil {
		t.Fatal("exported persistent state is missing truedemocracy genesis")
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
	deadline := time.Now().Add(35 * time.Second)
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
	if err := bindGenesisValidatorKey(path, generatedPubKey); err != nil {
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
	if len(tdGenesis.Domains) != 1 || len(tdGenesis.Validators) != 1 {
		t.Fatalf("unexpected PoD bootstrap state: %+v", tdGenesis)
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
	if err := bindGenesisValidatorKey(path, []byte{1}); err == nil {
		t.Fatal("invalid consensus key accepted")
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(after, before) {
		t.Fatal("invalid key mutated genesis")
	}
}
