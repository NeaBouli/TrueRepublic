package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cometbft/cometbft/privval"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"truerepublic/token"
	"truerepublic/x/truedemocracy"
)

const multiValidatorSmokeEnv = "TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE"

type smokeValidator struct {
	name    string
	home    string
	nodeID  string
	pubKey  []byte
	rpcPort int
	p2pPort int
	logPath string
	command *exec.Cmd
	done    chan error
	logFile *os.File
}

func TestConfigureGenesisValidatorSetBuildsExactBankBackedSet(t *testing.T) {
	app := newGenesisTestApp(t)
	appState, err := json.Marshal(ModuleBasics.DefaultGenesis(app.appCodec))
	if err != nil {
		t.Fatal(err)
	}
	genesis := &genutiltypes.AppGenesis{
		ChainID:   "truerepublic-multi-genesis-1",
		AppState:  appState,
		Consensus: &genutiltypes.ConsensusGenesis{},
	}
	identities := make([]genesisValidatorIdentity, 4)
	for i := range identities {
		identities[i] = genesisValidatorIdentity{
			Name:   fmt.Sprintf("validator-%d", i+1),
			PubKey: bytes.Repeat([]byte{byte(i + 1)}, 32),
		}
	}
	if err := configureGenesisValidatorSet(genesis, identities); err != nil {
		t.Fatal(err)
	}
	if len(genesis.Consensus.Validators) != len(identities) {
		t.Fatalf("consensus validator count = %d, want %d", len(genesis.Consensus.Validators), len(identities))
	}
	var state map[string]json.RawMessage
	if err := json.Unmarshal(genesis.AppState, &state); err != nil {
		t.Fatal(err)
	}
	var democracyGenesis truedemocracy.GenesisState
	if err := json.Unmarshal(state[truedemocracy.ModuleName], &democracyGenesis); err != nil {
		t.Fatal(err)
	}
	if len(democracyGenesis.Validators) != len(identities) {
		t.Fatalf("PoD validator count = %d, want %d", len(democracyGenesis.Validators), len(identities))
	}
	for i, identity := range identities {
		if !bytes.Equal(genesis.Consensus.Validators[i].PubKey.Bytes(), identity.PubKey) {
			t.Fatalf("consensus validator %d key does not match", i)
		}
		if !bytes.Equal(democracyGenesis.Validators[i].PubKey, identity.PubKey) {
			t.Fatalf("PoD validator %d key does not match", i)
		}
	}
	if err := validateLedgerGenesis(app.appCodec, state); err != nil {
		t.Fatalf("multi-validator genesis is not exactly bank-backed: %v", err)
	}
}

func TestMultiValidatorConsensusRecovery(t *testing.T) {
	if os.Getenv(multiValidatorSmokeEnv) != "1" {
		t.Skipf("set %s=1 to run the multi-validator process harness", multiValidatorSmokeEnv)
	}

	binary := filepath.Join(t.TempDir(), "truerepublicd")
	build := exec.Command("go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build daemon: %v\n%s", err, output)
	}

	const chainID = "truerepublic-multi-recovery-1"
	validators := make([]*smokeValidator, 4)
	for i := range validators {
		validator := &smokeValidator{
			name:    fmt.Sprintf("validator-%d", i+1),
			home:    filepath.Join(t.TempDir(), fmt.Sprintf("node-%d", i+1)),
			rpcPort: freeTCPPort(t),
			p2pPort: freeTCPPort(t),
			logPath: filepath.Join(t.TempDir(), fmt.Sprintf("validator-%d.log", i+1)),
		}
		initCmd := exec.Command(binary, "init", validator.name, "--chain-id", chainID, "--home", validator.home)
		if output, err := initCmd.CombinedOutput(); err != nil {
			t.Fatalf("init %s: %v\n%s", validator.name, err, output)
		}
		configureLocalhostSmokeP2P(t, filepath.Join(validator.home, "config", "config.toml"))
		filePV := privval.LoadFilePV(
			filepath.Join(validator.home, "config", "priv_validator_key.json"),
			filepath.Join(validator.home, "data", "priv_validator_state.json"),
		)
		pubKey, err := filePV.GetPubKey()
		if err != nil {
			t.Fatalf("read %s public key: %v", validator.name, err)
		}
		validator.pubKey = append([]byte(nil), pubKey.Bytes()...)
		nodeIDCmd := exec.Command(binary, "comet", "show-node-id", "--home", validator.home)
		nodeID, err := nodeIDCmd.Output()
		if err != nil {
			t.Fatalf("read %s node id: %v", validator.name, err)
		}
		validator.nodeID = strings.TrimSpace(string(nodeID))
		if validator.nodeID == "" {
			t.Fatalf("%s node id is empty", validator.name)
		}
		validators[i] = validator
	}

	sharedGenesis := buildSharedSmokeGenesis(t, chainID, validators)
	for _, validator := range validators {
		genesisPath := filepath.Join(validator.home, "config", "genesis.json")
		if err := atomicWriteFile(genesisPath, sharedGenesis, 0o600); err != nil {
			t.Fatalf("write %s shared genesis: %v", validator.name, err)
		}
		written, err := os.ReadFile(genesisPath)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(written, sharedGenesis) {
			t.Fatalf("%s did not receive the identical shared genesis", validator.name)
		}
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

	for _, validator := range validators {
		if err := validator.start(binary, persistentPeers(validator, validators)); err != nil {
			t.Fatalf("start %s: %v", validator.name, err)
		}
	}
	waitForSmokeHeight(t, validators, 2, 90*time.Second)
	assertCommonAppHash(t, validators, 2)

	failed := validators[len(validators)-1]
	if err := failed.stop(true); err != nil {
		t.Fatalf("stop %s: %v", failed.name, err)
	}
	survivors := validators[:len(validators)-1]
	failureHeight := smokeHeight(t, survivors[0])
	recoveryHeight := failureHeight + 2
	waitForSmokeHeight(t, survivors, recoveryHeight, 90*time.Second)

	if err := failed.start(binary, persistentPeers(failed, validators)); err != nil {
		t.Fatalf("restart %s: %v", failed.name, err)
	}
	waitForSmokeHeight(t, validators, recoveryHeight, 90*time.Second)
	assertCommonAppHash(t, validators, recoveryHeight)

	for _, validator := range validators {
		if err := validator.stop(true); err != nil {
			t.Fatalf("stop %s after recovery: %v", validator.name, err)
		}
	}
	exported := exec.Command(binary, "export", "--home", failed.home)
	exportOutput, err := exported.Output()
	if err != nil {
		t.Fatalf("export recovered %s state: %v", failed.name, err)
	}
	var exportedGenesis struct {
		InitialHeight int64                      `json:"initial_height"`
		AppState      map[string]json.RawMessage `json:"app_state"`
	}
	if err := json.Unmarshal(exportOutput, &exportedGenesis); err != nil {
		t.Fatalf("decode recovered export: %v", err)
	}
	if exportedGenesis.InitialHeight <= recoveryHeight {
		t.Fatalf("exported initial height = %d, want greater than recovery height %d", exportedGenesis.InitialHeight, recoveryHeight)
	}
	var democracyGenesis truedemocracy.GenesisState
	if err := json.Unmarshal(exportedGenesis.AppState[truedemocracy.ModuleName], &democracyGenesis); err != nil {
		t.Fatalf("decode recovered PoD genesis: %v", err)
	}
	if len(democracyGenesis.Validators) != len(validators) {
		t.Fatalf("recovered validator count = %d, want %d", len(democracyGenesis.Validators), len(validators))
	}
	exportApp := newGenesisTestApp(t)
	if err := validateLedgerGenesis(exportApp.appCodec, exportedGenesis.AppState); err != nil {
		t.Fatalf("recovered export is not exactly bank-backed: %v", err)
	}
}

func buildSharedSmokeGenesis(t *testing.T, chainID string, validators []*smokeValidator) []byte {
	t.Helper()
	app := newGenesisTestApp(t)
	appState, err := json.Marshal(ModuleBasics.DefaultGenesis(app.appCodec))
	if err != nil {
		t.Fatal(err)
	}
	genesis := &genutiltypes.AppGenesis{
		ChainID:   chainID,
		AppState:  appState,
		Consensus: &genutiltypes.ConsensusGenesis{},
	}
	identities := make([]genesisValidatorIdentity, len(validators))
	for i, validator := range validators {
		identities[i] = genesisValidatorIdentity{Name: validator.name, PubKey: validator.pubKey}
	}
	if err := configureGenesisValidatorSet(genesis, identities); err != nil {
		t.Fatalf("build shared multi-validator genesis: %v", err)
	}
	shared, err := json.MarshalIndent(genesis, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(shared, []byte(`"priv_key"`)) {
		t.Fatal("shared genesis contains private validator material")
	}
	return shared
}

func persistentPeers(self *smokeValidator, validators []*smokeValidator) string {
	peers := make([]string, 0, len(validators)-1)
	for _, validator := range validators {
		if validator == self {
			continue
		}
		peers = append(peers, fmt.Sprintf("%s@127.0.0.1:%d", validator.nodeID, validator.p2pPort))
	}
	return strings.Join(peers, ",")
}

func configureLocalhostSmokeP2P(t *testing.T, configPath string) {
	t.Helper()
	config, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	for original, replacement := range map[string]string{
		"addr_book_strict = true":    "addr_book_strict = false",
		"allow_duplicate_ip = false": "allow_duplicate_ip = true",
	} {
		if bytes.Count(config, []byte(original)) != 1 {
			t.Fatalf("%s does not contain exactly one %q setting", configPath, original)
		}
		config = bytes.Replace(config, []byte(original), []byte(replacement), 1)
	}
	if err := atomicWriteFile(configPath, config, 0o600); err != nil {
		t.Fatal(err)
	}
}

func (validator *smokeValidator) start(binary, peers string) error {
	if validator.command != nil {
		return errors.New("validator process is already running")
	}
	logFile, err := os.OpenFile(validator.logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	command := exec.Command(binary,
		"start",
		"--home", validator.home,
		"--rpc.laddr", fmt.Sprintf("tcp://127.0.0.1:%d", validator.rpcPort),
		"--p2p.laddr", fmt.Sprintf("tcp://127.0.0.1:%d", validator.p2pPort),
		"--p2p.persistent_peers", peers,
		"--rpc.pprof_laddr", "",
		"--grpc.enable=false",
		"--api.enable=false",
		"--minimum-gas-prices", "0"+token.BaseDenom,
	)
	command.Stdout = logFile
	command.Stderr = logFile
	if err := command.Start(); err != nil {
		_ = logFile.Close()
		return err
	}
	validator.command = command
	validator.done = make(chan error, 1)
	validator.logFile = logFile
	go func() {
		validator.done <- command.Wait()
	}()
	return nil
}

func (validator *smokeValidator) stop(requireClean bool) error {
	if validator.command == nil {
		return nil
	}
	command := validator.command
	if err := command.Process.Signal(os.Interrupt); err != nil {
		_ = command.Process.Kill()
	}
	var processErr error
	select {
	case processErr = <-validator.done:
	case <-time.After(20 * time.Second):
		_ = command.Process.Kill()
		select {
		case <-validator.done:
			processErr = errors.New("process required a forced shutdown")
		case <-time.After(5 * time.Second):
			processErr = errors.New("process did not exit after forced shutdown")
		}
	}
	closeErr := validator.logFile.Close()
	validator.command = nil
	validator.done = nil
	validator.logFile = nil
	if requireClean && processErr != nil {
		return processErr
	}
	return closeErr
}

func (validator *smokeValidator) logContents(t *testing.T) {
	t.Helper()
	content, err := os.ReadFile(validator.logPath)
	if err != nil {
		t.Logf("%s log unavailable: %v", validator.name, err)
		return
	}
	t.Logf("%s log:\n%s", validator.name, content)
}

func waitForSmokeHeight(t *testing.T, validators []*smokeValidator, minimum int64, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ready := true
		for _, validator := range validators {
			height, err := querySmokeHeight(validator)
			if err != nil || height < minimum {
				ready = false
				break
			}
		}
		if ready {
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("validators did not all reach height %d within %s", minimum, timeout)
}

func smokeHeight(t *testing.T, validator *smokeValidator) int64 {
	t.Helper()
	height, err := querySmokeHeight(validator)
	if err != nil {
		t.Fatalf("query %s height: %v", validator.name, err)
	}
	return height
}

func querySmokeHeight(validator *smokeValidator) (int64, error) {
	client := &http.Client{Timeout: time.Second}
	response, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/status", validator.rpcPort))
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("status response = %s", response.Status)
	}
	var status struct {
		Result struct {
			SyncInfo struct {
				LatestBlockHeight string `json:"latest_block_height"`
			} `json:"sync_info"`
		} `json:"result"`
	}
	if err := json.NewDecoder(response.Body).Decode(&status); err != nil {
		return 0, err
	}
	return strconv.ParseInt(status.Result.SyncInfo.LatestBlockHeight, 10, 64)
}

func assertCommonAppHash(t *testing.T, validators []*smokeValidator, height int64) {
	t.Helper()
	var expected string
	for _, validator := range validators {
		hash := smokeAppHash(t, validator, height)
		if hash == "" {
			t.Fatalf("%s returned an empty app hash at height %d", validator.name, height)
		}
		if expected == "" {
			expected = hash
			continue
		}
		if hash != expected {
			t.Fatalf("%s app hash at height %d = %s, want %s", validator.name, height, hash, expected)
		}
	}
}

func smokeAppHash(t *testing.T, validator *smokeValidator, height int64) string {
	t.Helper()
	client := &http.Client{Timeout: time.Second}
	url := fmt.Sprintf("http://127.0.0.1:%d/block?height=%d", validator.rpcPort, height)
	response, err := client.Get(url)
	if err != nil {
		t.Fatalf("query %s block %d: %v", validator.name, height, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("query %s block %d: %s", validator.name, height, response.Status)
	}
	var block struct {
		Result struct {
			Block struct {
				Header struct {
					AppHash string `json:"app_hash"`
				} `json:"header"`
			} `json:"block"`
		} `json:"result"`
	}
	if err := json.NewDecoder(response.Body).Decode(&block); err != nil {
		t.Fatalf("decode %s block %d: %v", validator.name, height, err)
	}
	return block.Result.Block.Header.AppHash
}
