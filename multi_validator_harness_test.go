package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
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

	"cosmossdk.io/math"
	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/privval"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"truerepublic/token"
	"truerepublic/x/truedemocracy"
)

const multiValidatorSmokeEnv = "TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE"

type smokeValidator struct {
	name         string
	home         string
	nodeID       string
	pubKey       []byte
	operatorAddr string
	rpcPort      int
	p2pPort      int
	logPath      string
	command      *exec.Cmd
	done         chan error
	logFile      *os.File
}

type smokeAccount struct {
	name          string
	address       string
	balance       int64
	keyringDir    string
	accountNumber uint64
	sequence      uint64
}

type smokeExportedGenesis struct {
	InitialHeight int64                      `json:"initial_height"`
	AppState      map[string]json.RawMessage `json:"app_state"`
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
			Name:         fmt.Sprintf("validator-%d", i+1),
			PubKey:       bytes.Repeat([]byte{byte(i + 1)}, 32),
			OperatorAddr: sdk.AccAddress(bytes.Repeat([]byte{byte(i + 21)}, 20)).String(),
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

func TestConfigureGenesisValidatorSetRejectsCrossCoupledOperators(t *testing.T) {
	app := newGenesisTestApp(t)
	appState, err := json.Marshal(ModuleBasics.DefaultGenesis(app.appCodec))
	if err != nil {
		t.Fatal(err)
	}
	genesis := &genutiltypes.AppGenesis{
		ChainID: "truerepublic-cross-coupling-1", AppState: appState, Consensus: &genutiltypes.ConsensusGenesis{},
	}
	firstKey := bytes.Repeat([]byte{0x61}, 32)
	secondKey := bytes.Repeat([]byte{0x62}, 32)
	identities := []genesisValidatorIdentity{
		{Name: "validator-1", PubKey: firstKey, OperatorAddr: sdk.AccAddress(cmted25519.PubKey(secondKey).Address()).String()},
		{Name: "validator-2", PubKey: secondKey, OperatorAddr: sdk.AccAddress(cmted25519.PubKey(firstKey).Address()).String()},
	}
	if err := configureGenesisValidatorSet(genesis, identities); err == nil {
		t.Fatal("cross-coupled bootstrap operator authorities were accepted")
	}
}

func TestMultiValidatorConsensusRecovery(t *testing.T) {
	if os.Getenv(multiValidatorSmokeEnv) != "1" {
		t.Skipf("set %s=1 to run the multi-validator process harness", multiValidatorSmokeEnv)
	}
	ctx := t.Context()

	binary := filepath.Join(t.TempDir(), "truerepublicd")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
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
		validator.operatorAddr = smokeOperatorAddress(validator.name)
		initCmd := exec.CommandContext(ctx, binary, "init", validator.name, "--chain-id", chainID, "--home", validator.home, "--bootstrap-operator", validator.operatorAddr)
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
		nodeIDCmd := exec.CommandContext(ctx, binary, "comet", "show-node-id", "--home", validator.home)
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
		if err := validator.start(ctx, binary, persistentPeers(validator, validators)); err != nil {
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

	if err := failed.start(ctx, binary, persistentPeers(failed, validators)); err != nil {
		t.Fatalf("restart %s: %v", failed.name, err)
	}
	postRestartHeight := smokeHeight(t, survivors[0]) + 2
	waitForSmokeHeight(t, validators, postRestartHeight, 90*time.Second)
	assertCommonAppHash(t, validators, postRestartHeight)

	for _, validator := range validators {
		if err := validator.stop(true); err != nil {
			t.Fatalf("stop %s after recovery: %v", validator.name, err)
		}
	}
	exported := exec.CommandContext(ctx, binary, "export", "--home", failed.home)
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
	if exportedGenesis.InitialHeight <= postRestartHeight {
		t.Fatalf("exported initial height = %d, want greater than post-restart height %d", exportedGenesis.InitialHeight, postRestartHeight)
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

func TestMultiValidatorJoinReplacementLifecycle(t *testing.T) {
	if os.Getenv(multiValidatorSmokeEnv) != "1" {
		t.Skipf("set %s=1 to run the multi-validator process harness", multiValidatorSmokeEnv)
	}
	ctx := t.Context()

	binary := filepath.Join(t.TempDir(), "truerepublicd")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build daemon: %v\n%s", err, output)
	}

	const chainID = "truerepublic-validator-lifecycle-1"
	validators := make([]*smokeValidator, 6)
	for i := range validators {
		validator := &smokeValidator{
			name:    fmt.Sprintf("validator-%d", i+1),
			home:    filepath.Join(t.TempDir(), fmt.Sprintf("node-%d", i+1)),
			rpcPort: freeTCPPort(t),
			p2pPort: freeTCPPort(t),
			logPath: filepath.Join(t.TempDir(), fmt.Sprintf("validator-%d.log", i+1)),
		}
		initSmokeValidator(t, ctx, binary, chainID, validator)
		validators[i] = validator
	}
	initialValidators := validators[:4]
	joiningValidator := validators[4]
	replacementValidator := validators[5]

	admin := addSmokeKey(t, ctx, binary, initialValidators[0].home, "lifecycle-admin", 4, 1_500_000*token.WholeTokenBaseUnits)
	joiningOperator := addSmokeKey(t, ctx, binary, joiningValidator.home, "joining-operator", 5, 200_000*token.WholeTokenBaseUnits)
	replacementOperator := addSmokeKey(t, ctx, binary, replacementValidator.home, "replacement-operator", 6, 200_000*token.WholeTokenBaseUnits)

	sharedGenesis := buildSharedSmokeGenesis(t, chainID, initialValidators, admin, joiningOperator, replacementOperator)
	for _, validator := range validators {
		genesisPath := filepath.Join(validator.home, "config", "genesis.json")
		if err := atomicWriteFile(genesisPath, sharedGenesis, 0o600); err != nil {
			t.Fatalf("write %s shared genesis: %v", validator.name, err)
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

	for _, validator := range initialValidators {
		if err := validator.start(ctx, binary, persistentPeers(validator, validators)); err != nil {
			t.Fatalf("start %s: %v", validator.name, err)
		}
	}
	waitForSmokeHeight(t, initialValidators, 2, 90*time.Second)
	assertCommonAppHash(t, initialValidators, 2)

	if err := joiningValidator.start(ctx, binary, persistentPeers(joiningValidator, validators)); err != nil {
		t.Fatalf("start joining validator as catching-up full node: %v", err)
	}
	waitForSmokeHeight(t, []*smokeValidator{joiningValidator}, smokeHeight(t, initialValidators[0]), 90*time.Second)

	runSmokeTx(t, ctx, binary, initialValidators[0], &admin, chainID,
		"create-domain", "Lifecycle", fmt.Sprintf("%d%s", 500_000*token.WholeTokenBaseUnits, token.BaseDenom))
	runSmokeTx(t, ctx, binary, initialValidators[0], &admin, chainID,
		"add-member", "Lifecycle", joiningOperator.address)
	runSmokeTx(t, ctx, binary, joiningValidator, &joiningOperator, chainID,
		"register-validator", hex.EncodeToString(joiningValidator.pubKey),
		fmt.Sprintf("%d%s", token.StakeMinBaseUnits, token.BaseDenom), "Lifecycle")
	waitForSmokeValidatorPower(t, initialValidators[0], joiningValidator.pubKey, "1", 90*time.Second)
	postJoinHeight := smokeHeight(t, initialValidators[0]) + 2
	joinedSet := append([]*smokeValidator{}, initialValidators...)
	joinedSet = append(joinedSet, joiningValidator)
	waitForSmokeHeight(t, joinedSet, postJoinHeight, 90*time.Second)
	assertCommonAppHash(t, joinedSet, postJoinHeight)

	if err := joiningValidator.stop(true); err != nil {
		t.Fatalf("stop joined validator before replacement: %v", err)
	}
	runSmokeTx(t, ctx, binary, initialValidators[0], &admin, chainID,
		"add-member", "Lifecycle", replacementOperator.address)
	if err := replacementValidator.start(ctx, binary, persistentPeers(replacementValidator, validators)); err != nil {
		t.Fatalf("start replacement validator as catching-up full node: %v", err)
	}
	waitForSmokeHeight(t, []*smokeValidator{replacementValidator}, smokeHeight(t, initialValidators[0]), 90*time.Second)
	runSmokeTx(t, ctx, binary, replacementValidator, &replacementOperator, chainID,
		"register-validator", hex.EncodeToString(replacementValidator.pubKey),
		fmt.Sprintf("%d%s", token.StakeMinBaseUnits, token.BaseDenom), "Lifecycle")
	waitForSmokeValidatorPower(t, initialValidators[0], replacementValidator.pubKey, "1", 90*time.Second)
	replacementSet := append([]*smokeValidator{}, initialValidators...)
	replacementSet = append(replacementSet, replacementValidator)
	postReplacementHeight := smokeHeight(t, initialValidators[0]) + 2
	waitForSmokeHeight(t, replacementSet, postReplacementHeight, 90*time.Second)
	assertCommonAppHash(t, replacementSet, postReplacementHeight)
}

func TestMultiValidatorNetworkPartitionRecovery(t *testing.T) {
	if os.Getenv(multiValidatorSmokeEnv) != "1" {
		t.Skipf("set %s=1 to run the multi-validator process harness", multiValidatorSmokeEnv)
	}
	ctx := t.Context()

	binary := filepath.Join(t.TempDir(), "truerepublicd")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build daemon: %v\n%s", err, output)
	}

	const chainID = "truerepublic-network-partition-1"
	validators := make([]*smokeValidator, 4)
	for i := range validators {
		validator := &smokeValidator{
			name:    fmt.Sprintf("validator-%d", i+1),
			home:    filepath.Join(t.TempDir(), fmt.Sprintf("node-%d", i+1)),
			rpcPort: freeTCPPort(t),
			p2pPort: freeTCPPort(t),
			logPath: filepath.Join(t.TempDir(), fmt.Sprintf("validator-%d.log", i+1)),
		}
		initSmokeValidator(t, ctx, binary, chainID, validator)
		validators[i] = validator
	}
	quorum := validators[:3]
	isolated := validators[3]
	admin := addSmokeKey(t, ctx, binary, quorum[0].home, "partition-admin", 4, 800_000*token.WholeTokenBaseUnits)

	sharedGenesis := buildSharedSmokeGenesis(t, chainID, validators, admin)
	for _, validator := range validators {
		genesisPath := filepath.Join(validator.home, "config", "genesis.json")
		if err := atomicWriteFile(genesisPath, sharedGenesis, 0o600); err != nil {
			t.Fatalf("write %s shared genesis: %v", validator.name, err)
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

	for _, validator := range quorum {
		if err := validator.start(ctx, binary, persistentPeers(validator, quorum)); err != nil {
			t.Fatalf("start quorum %s: %v", validator.name, err)
		}
	}
	waitForSmokeHeight(t, quorum, 2, 90*time.Second)
	assertCommonAppHash(t, quorum, 2)
	assertSmokeValidatorPowers(t, quorum[0], validators, "1")

	if err := isolated.start(ctx, binary, ""); err != nil {
		t.Fatalf("start isolated validator without peers: %v", err)
	}
	waitForSmokeRPC(t, isolated, 30*time.Second)

	partitionStartHeight := smokeHeight(t, quorum[0])
	runSmokeTx(t, ctx, binary, quorum[0], &admin, chainID,
		"create-domain", "PartitionRecovery", fmt.Sprintf("%d%s", 500_000*token.WholeTokenBaseUnits, token.BaseDenom))
	quorumTargetHeight := partitionStartHeight + 3
	waitForSmokeHeight(t, quorum, quorumTargetHeight, 90*time.Second)
	assertCommonAppHash(t, quorum, quorumTargetHeight)

	isolatedHeight := smokeHeight(t, isolated)
	if isolatedHeight >= quorumTargetHeight {
		t.Fatalf("isolated validator reached height %d during partition, want below quorum height %d", isolatedHeight, quorumTargetHeight)
	}

	if err := isolated.stop(true); err != nil {
		t.Fatalf("stop isolated validator before reconnect: %v", err)
	}
	if err := isolated.start(ctx, binary, persistentPeers(isolated, validators)); err != nil {
		t.Fatalf("restart isolated validator with peers: %v", err)
	}
	recoveryTargetHeight := smokeHeight(t, quorum[0]) + 2
	waitForSmokeHeight(t, validators, recoveryTargetHeight, 120*time.Second)
	assertCommonAppHash(t, validators, recoveryTargetHeight)
	assertSmokeValidatorPowers(t, validators[0], validators, "1")

	for _, validator := range validators {
		if err := validator.stop(true); err != nil {
			t.Fatalf("stop %s after partition recovery: %v", validator.name, err)
		}
		exportedGenesis := exportSmokeGenesis(t, ctx, binary, validator, recoveryTargetHeight)
		exportApp := newGenesisTestApp(t)
		if err := validateLedgerGenesis(exportApp.appCodec, exportedGenesis.AppState); err != nil {
			t.Fatalf("%s recovered export is not exactly bank-backed: %v", validator.name, err)
		}
	}
}

func TestMultiValidatorTrustedSnapshotStateSync(t *testing.T) {
	if os.Getenv(multiValidatorSmokeEnv) != "1" {
		t.Skipf("set %s=1 to run the multi-validator process harness", multiValidatorSmokeEnv)
	}
	ctx := t.Context()

	binary := filepath.Join(t.TempDir(), "truerepublicd")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build daemon: %v\n%s", err, output)
	}

	const chainID = "truerepublic-state-sync-1"
	validators := make([]*smokeValidator, 4)
	for i := range validators {
		validator := &smokeValidator{
			name:    fmt.Sprintf("validator-%d", i+1),
			home:    filepath.Join(t.TempDir(), fmt.Sprintf("node-%d", i+1)),
			rpcPort: freeTCPPort(t),
			p2pPort: freeTCPPort(t),
			logPath: filepath.Join(t.TempDir(), fmt.Sprintf("validator-%d.log", i+1)),
		}
		initSmokeValidator(t, ctx, binary, chainID, validator)
		validators[i] = validator
	}
	syncingNode := &smokeValidator{
		name:    "state-sync-node",
		home:    filepath.Join(t.TempDir(), "state-sync-node"),
		rpcPort: freeTCPPort(t),
		p2pPort: freeTCPPort(t),
		logPath: filepath.Join(t.TempDir(), "state-sync-node.log"),
	}
	initSmokeValidator(t, ctx, binary, chainID, syncingNode)

	admin := addSmokeKey(t, ctx, binary, validators[0].home, "state-sync-admin", 4, 800_000*token.WholeTokenBaseUnits)
	sharedGenesis := buildSharedSmokeGenesis(t, chainID, validators, admin)
	for _, validator := range append(append([]*smokeValidator{}, validators...), syncingNode) {
		genesisPath := filepath.Join(validator.home, "config", "genesis.json")
		if err := atomicWriteFile(genesisPath, sharedGenesis, 0o600); err != nil {
			t.Fatalf("write %s shared genesis: %v", validator.name, err)
		}
	}

	t.Cleanup(func() {
		for _, validator := range append(append([]*smokeValidator{}, validators...), syncingNode) {
			_ = validator.stop(false)
		}
		if t.Failed() {
			for _, validator := range append(append([]*smokeValidator{}, validators...), syncingNode) {
				validator.logContents(t)
			}
		}
	})

	for _, validator := range validators {
		if err := validator.startWithArgs(ctx, binary, persistentPeers(validator, validators),
			"--state-sync.snapshot-interval", "2",
			"--state-sync.snapshot-keep-recent", "3",
		); err != nil {
			t.Fatalf("start snapshot provider %s: %v", validator.name, err)
		}
	}
	waitForSmokeHeight(t, validators, 8, 180*time.Second)
	assertCommonAppHash(t, validators, 8)

	runSmokeTx(t, ctx, binary, validators[0], &admin, chainID,
		"create-domain", "TrustedStateSync", fmt.Sprintf("%d%s", 500_000*token.WholeTokenBaseUnits, token.BaseDenom))
	waitForSmokeHeight(t, validators, 10, 180*time.Second)
	assertCommonAppHash(t, validators, 10)

	trustHeight := smokeHeight(t, validators[0]) - 2
	if trustHeight < 4 {
		t.Fatalf("trust height = %d, want at least 4", trustHeight)
	}
	trustHash := smokeBlockHash(t, validators[0], trustHeight)
	if trustHash == "" {
		t.Fatalf("%s returned an empty block hash at trust height %d", validators[0].name, trustHeight)
	}
	configureSmokeStateSync(t, filepath.Join(syncingNode.home, "config", "config.toml"), trustHeight, trustHash, validators[0], validators[1])

	if err := syncingNode.start(ctx, binary, persistentPeers(syncingNode, validators)); err != nil {
		t.Fatalf("start state-sync node: %v", err)
	}
	waitForSmokeHeight(t, []*smokeValidator{syncingNode}, trustHeight, 180*time.Second)

	convergenceHeight := smokeHeight(t, validators[0]) + 2
	allNodes := append(append([]*smokeValidator{}, validators...), syncingNode)
	waitForSmokeHeight(t, allNodes, convergenceHeight, 180*time.Second)
	assertCommonAppHash(t, allNodes, convergenceHeight)
	assertSmokeValidatorPowers(t, syncingNode, validators, "1")

	if err := syncingNode.stop(true); err != nil {
		t.Fatalf("stop state-sync node: %v", err)
	}
	exportedGenesis := exportSmokeGenesis(t, ctx, binary, syncingNode, convergenceHeight)
	exportApp := newGenesisTestApp(t)
	if err := validateLedgerGenesis(exportApp.appCodec, exportedGenesis.AppState); err != nil {
		t.Fatalf("state-synced export is not exactly bank-backed: %v", err)
	}
}

func TestMultiValidatorBackupRestoreExportImport(t *testing.T) {
	if os.Getenv(multiValidatorSmokeEnv) != "1" {
		t.Skipf("set %s=1 to run the multi-validator process harness", multiValidatorSmokeEnv)
	}
	ctx := t.Context()

	binary := filepath.Join(t.TempDir(), "truerepublicd")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build daemon: %v\n%s", err, output)
	}

	const chainID = "truerepublic-backup-restore-1"
	validators := make([]*smokeValidator, 4)
	for i := range validators {
		validator := &smokeValidator{
			name:    fmt.Sprintf("validator-%d", i+1),
			home:    filepath.Join(t.TempDir(), fmt.Sprintf("node-%d", i+1)),
			rpcPort: freeTCPPort(t),
			p2pPort: freeTCPPort(t),
			logPath: filepath.Join(t.TempDir(), fmt.Sprintf("validator-%d.log", i+1)),
		}
		initSmokeValidator(t, ctx, binary, chainID, validator)
		validators[i] = validator
	}
	fullNode := &smokeValidator{
		name:    "backup-source-full-node",
		home:    filepath.Join(t.TempDir(), "backup-source-full-node"),
		rpcPort: freeTCPPort(t),
		p2pPort: freeTCPPort(t),
		logPath: filepath.Join(t.TempDir(), "backup-source-full-node.log"),
	}
	initSmokeValidator(t, ctx, binary, chainID, fullNode)
	restoredNode := &smokeValidator{
		name:    "backup-restored-full-node",
		home:    filepath.Join(t.TempDir(), "backup-restored-full-node"),
		rpcPort: freeTCPPort(t),
		p2pPort: freeTCPPort(t),
		logPath: filepath.Join(t.TempDir(), "backup-restored-full-node.log"),
	}
	initSmokeValidator(t, ctx, binary, chainID, restoredNode)

	admin := addSmokeKey(t, ctx, binary, validators[0].home, "backup-admin", 4, 800_000*token.WholeTokenBaseUnits)
	sharedGenesis := buildSharedSmokeGenesis(t, chainID, validators, admin)
	allHomes := append(append([]*smokeValidator{}, validators...), fullNode, restoredNode)
	for _, validator := range allHomes {
		genesisPath := filepath.Join(validator.home, "config", "genesis.json")
		if err := atomicWriteFile(genesisPath, sharedGenesis, 0o600); err != nil {
			t.Fatalf("write %s shared genesis: %v", validator.name, err)
		}
	}

	t.Cleanup(func() {
		for _, validator := range allHomes {
			_ = validator.stop(false)
		}
		if t.Failed() {
			for _, validator := range allHomes {
				validator.logContents(t)
			}
		}
	})

	for _, validator := range validators {
		if err := validator.start(ctx, binary, persistentPeers(validator, validators)); err != nil {
			t.Fatalf("start %s: %v", validator.name, err)
		}
	}
	waitForSmokeHeight(t, validators, 2, 90*time.Second)
	assertCommonAppHash(t, validators, 2)

	if err := fullNode.start(ctx, binary, persistentPeers(fullNode, validators)); err != nil {
		t.Fatalf("start backup source full node: %v", err)
	}
	waitForSmokeHeight(t, []*smokeValidator{fullNode}, smokeHeight(t, validators[0]), 90*time.Second)
	runSmokeTx(t, ctx, binary, validators[0], &admin, chainID,
		"create-domain", "BackupRestore", fmt.Sprintf("%d%s", 500_000*token.WholeTokenBaseUnits, token.BaseDenom))
	sourceHeight := smokeHeight(t, validators[0]) + 2
	waitForSmokeHeight(t, append(append([]*smokeValidator{}, validators...), fullNode), sourceHeight, 90*time.Second)
	assertCommonAppHash(t, append(append([]*smokeValidator{}, validators...), fullNode), sourceHeight)

	if err := fullNode.stop(true); err != nil {
		t.Fatalf("stop backup source full node: %v", err)
	}
	backupDir := t.TempDir()
	backupCmd := exec.CommandContext(ctx, "bash", filepath.Join("scripts", "backup.sh"), backupDir)
	backupCmd.Env = append(os.Environ(), "CHAIN_HOME="+fullNode.home)
	if output, err := backupCmd.CombinedOutput(); err != nil {
		t.Fatalf("create sanitized backup: %v\n%s", err, output)
	}
	matches, err := filepath.Glob(filepath.Join(backupDir, "truerepublic_*.tar.gz"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 {
		t.Fatalf("backup artifact count = %d, want 1: %v", len(matches), matches)
	}
	backupArchive := matches[0]
	assertSanitizedBackupArchive(t, ctx, backupArchive)

	restoredNodeKeyPath := filepath.Join(restoredNode.home, "config", "node_key.json")
	restoredValidatorKeyPath := filepath.Join(restoredNode.home, "config", "priv_validator_key.json")
	restoredNodeKeyBefore, err := os.ReadFile(restoredNodeKeyPath)
	if err != nil {
		t.Fatal(err)
	}
	restoredValidatorKeyBefore, err := os.ReadFile(restoredValidatorKeyPath)
	if err != nil {
		t.Fatal(err)
	}
	restoreCmd := exec.CommandContext(ctx, "bash", filepath.Join("scripts", "restore.sh"), backupArchive, restoredNode.home)
	if output, err := restoreCmd.CombinedOutput(); err != nil {
		t.Fatalf("restore sanitized backup: %v\n%s", err, output)
	}
	restoredNodeKeyAfter, err := os.ReadFile(restoredNodeKeyPath)
	if err != nil {
		t.Fatal(err)
	}
	restoredValidatorKeyAfter, err := os.ReadFile(restoredValidatorKeyPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(restoredNodeKeyAfter, restoredNodeKeyBefore) {
		t.Fatal("restore replaced the target node key")
	}
	if !bytes.Equal(restoredValidatorKeyAfter, restoredValidatorKeyBefore) {
		t.Fatal("restore replaced the target validator key")
	}

	if err := restoredNode.start(ctx, binary, persistentPeers(restoredNode, validators)); err != nil {
		t.Fatalf("start restored full node: %v", err)
	}
	recoveryHeight := smokeHeight(t, validators[0]) + 2
	waitForSmokeHeight(t, append(append([]*smokeValidator{}, validators...), restoredNode), recoveryHeight, 120*time.Second)
	assertCommonAppHash(t, append(append([]*smokeValidator{}, validators...), restoredNode), recoveryHeight)

	if err := restoredNode.stop(true); err != nil {
		t.Fatalf("stop restored full node: %v", err)
	}
	exportedGenesis := exportSmokeGenesis(t, ctx, binary, restoredNode, recoveryHeight)
	exportApp := newGenesisTestApp(t)
	if err := validateLedgerGenesis(exportApp.appCodec, exportedGenesis.AppState); err != nil {
		t.Fatalf("restored export is not exactly bank-backed: %v", err)
	}
	importApp := newGenesisTestApp(t)
	if err := initGenesisApp(importApp, exportedGenesis.AppState); err != nil {
		t.Fatalf("re-import restored export: %v", err)
	}
}

func buildSharedSmokeGenesis(t *testing.T, chainID string, validators []*smokeValidator, accounts ...smokeAccount) []byte {
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
		operatorAddr := validator.operatorAddr
		if operatorAddr == "" {
			operatorAddr = sdk.AccAddress(bytes.Repeat([]byte{byte(i + 41)}, 20)).String()
			validators[i].operatorAddr = operatorAddr
		}
		identities[i] = genesisValidatorIdentity{Name: validator.name, PubKey: validator.pubKey, OperatorAddr: operatorAddr}
	}
	if err := configureGenesisValidatorSet(genesis, identities); err != nil {
		t.Fatalf("build shared multi-validator genesis: %v", err)
	}
	if len(accounts) > 0 {
		var state map[string]json.RawMessage
		if err := json.Unmarshal(genesis.AppState, &state); err != nil {
			t.Fatal(err)
		}
		addSmokeAccountsToGenesis(t, app, state, accounts)
		updatedState, err := json.Marshal(state)
		if err != nil {
			t.Fatal(err)
		}
		genesis.AppState = updatedState
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

func initSmokeValidator(t *testing.T, ctx context.Context, binary, chainID string, validator *smokeValidator) {
	t.Helper()
	if validator.operatorAddr == "" {
		validator.operatorAddr = smokeOperatorAddress(validator.name)
	}
	initCmd := exec.CommandContext(ctx, binary, "init", validator.name, "--chain-id", chainID, "--home", validator.home, "--bootstrap-operator", validator.operatorAddr)
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
	nodeIDCmd := exec.CommandContext(ctx, binary, "comet", "show-node-id", "--home", validator.home)
	nodeID, err := nodeIDCmd.Output()
	if err != nil {
		t.Fatalf("read %s node id: %v", validator.name, err)
	}
	validator.nodeID = strings.TrimSpace(string(nodeID))
	if validator.nodeID == "" {
		t.Fatalf("%s node id is empty", validator.name)
	}
}

func smokeOperatorAddress(name string) string {
	digest := sha256.Sum256([]byte("truerepublic-smoke-operator:" + name))
	return sdk.AccAddress(digest[:20]).String()
}

func addSmokeKey(t *testing.T, ctx context.Context, binary, home, name string, accountNumber uint64, balance int64) smokeAccount {
	t.Helper()
	keyringDir := filepath.Join(t.TempDir(), name+"-keyring")
	if err := os.MkdirAll(keyringDir, 0o700); err != nil {
		t.Fatalf("create keyring dir for %s: %v", name, err)
	}
	command := exec.CommandContext(ctx, binary, "keys", "add", name,
		"--home", home,
		"--keyring-dir", keyringDir,
		"--keyring-backend", "test",
		"--output", "json",
	)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("add key %s: %v\n%s", name, err, output)
	}
	var key struct {
		Address string `json:"address"`
	}
	if err := json.Unmarshal(output, &key); err != nil {
		t.Fatalf("decode key %s: %v\n%s", name, err, output)
	}
	if key.Address == "" {
		t.Fatalf("key %s returned an empty address", name)
	}
	return smokeAccount{name: name, address: key.Address, balance: balance, keyringDir: keyringDir, accountNumber: accountNumber}
}

func addSmokeAccountsToGenesis(t *testing.T, app *TrueRepublicApp, state map[string]json.RawMessage, accounts []smokeAccount) {
	t.Helper()
	authGenesis := authtypes.GetGenesisStateFromAppState(app.appCodec, state)
	bankGenesis := banktypes.GetGenesisStateFromAppState(app.appCodec, state)
	existingAccounts, err := authtypes.UnpackAccounts(authGenesis.Accounts)
	if err != nil {
		t.Fatal(err)
	}
	existingByAddress := make(map[string]authtypes.GenesisAccount, len(existingAccounts))
	for _, existing := range existingAccounts {
		existingByAddress[existing.GetAddress().String()] = existing
	}
	genesisAccounts := make(authtypes.GenesisAccounts, 0, len(accounts))
	for _, account := range accounts {
		address, err := sdk.AccAddressFromBech32(account.address)
		if err != nil {
			t.Fatalf("decode smoke account %s address: %v", account.name, err)
		}
		if _, exists := existingByAddress[account.address]; !exists {
			baseAccount := authtypes.NewBaseAccountWithAddress(address)
			baseAccount.AccountNumber = account.accountNumber
			genesisAccounts = append(genesisAccounts, baseAccount)
			existingByAddress[account.address] = baseAccount
		}
		coins := sdk.NewCoins(token.NewCoin(math.NewInt(account.balance)))
		bankGenesis.Balances = append(bankGenesis.Balances, banktypes.Balance{Address: account.address, Coins: coins})
		if !bankGenesis.Supply.Empty() {
			bankGenesis.Supply = bankGenesis.Supply.Add(coins...)
		}
	}
	packed, err := authtypes.PackAccounts(genesisAccounts)
	if err != nil {
		t.Fatal(err)
	}
	authGenesis.Accounts = append(authGenesis.Accounts, packed...)
	bankGenesis.Balances = banktypes.SanitizeGenesisBalances(bankGenesis.Balances)
	token.EnsureMetadata(bankGenesis)
	authJSON, err := app.appCodec.MarshalJSON(&authGenesis)
	if err != nil {
		t.Fatal(err)
	}
	bankJSON, err := app.appCodec.MarshalJSON(bankGenesis)
	if err != nil {
		t.Fatal(err)
	}
	state[authtypes.ModuleName] = authJSON
	state[banktypes.ModuleName] = bankJSON
	if err := validateLedgerGenesis(app.appCodec, state); err != nil {
		t.Fatalf("funded smoke genesis is not bank-backed: %v", err)
	}
}

func runSmokeTx(t *testing.T, ctx context.Context, binary string, node *smokeValidator, from *smokeAccount, chainID string, args ...string) {
	t.Helper()
	commandArgs := append([]string{"tx", truedemocracy.ModuleName}, args...)
	commandArgs = append(commandArgs,
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
		"--gas", "500000",
		"--yes",
		"--output", "json",
	)
	command := exec.CommandContext(ctx, binary, commandArgs...)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("tx %s: %v\n%s", strings.Join(args, " "), err, output)
	}
	var result struct {
		Code   uint32 `json:"code"`
		RawLog string `json:"raw_log"`
		TxHash string `json:"txhash"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("decode tx %s: %v\n%s", strings.Join(args, " "), err, output)
	}
	if result.Code != 0 {
		t.Fatalf("tx %s failed with code %d: %s", strings.Join(args, " "), result.Code, result.RawLog)
	}
	if result.TxHash == "" {
		t.Fatalf("tx %s returned an empty hash: %s", strings.Join(args, " "), output)
	}
	waitForSmokeTx(t, ctx, binary, node, result.TxHash, args...)
	from.sequence++
}

func waitForSmokeTx(t *testing.T, ctx context.Context, _ string, node *smokeValidator, txHash string, args ...string) {
	t.Helper()
	deadline := time.Now().Add(45 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		code, height, rawLog, err := querySmokeTx(ctx, node, txHash)
		if err == nil {
			if code != 0 {
				t.Fatalf("delivered tx %s (%s) failed at height %s with code %d: %s", strings.Join(args, " "), txHash, height, code, rawLog)
			}
			return
		}
		lastErr = err
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("tx %s (%s) was not indexed within 45s: %v", strings.Join(args, " "), txHash, lastErr)
}

func querySmokeTx(ctx context.Context, node *smokeValidator, txHash string) (uint32, string, string, error) {
	client := &http.Client{Timeout: time.Second}
	url := fmt.Sprintf("http://127.0.0.1:%d/tx?hash=0x%s", node.rpcPort, txHash)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, "", "", err
	}
	response, err := client.Do(request)
	if err != nil {
		return 0, "", "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return 0, "", "", fmt.Errorf("tx query response = %s", response.Status)
	}
	var result struct {
		Result struct {
			Height   string `json:"height"`
			TxResult struct {
				Code uint32 `json:"code"`
				Log  string `json:"log"`
			} `json:"tx_result"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    string `json:"data"`
		} `json:"error"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return 0, "", "", err
	}
	if result.Error != nil {
		return 0, "", "", fmt.Errorf("tx query error %d %s: %s", result.Error.Code, result.Error.Message, result.Error.Data)
	}
	return result.Result.TxResult.Code, result.Result.Height, result.Result.TxResult.Log, nil
}

func waitForSmokeValidatorPower(t *testing.T, validator *smokeValidator, pubKey []byte, power string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		got, err := querySmokeValidatorPower(t.Context(), validator, pubKey)
		if err == nil && got == power {
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("%s validator set did not include pubkey %x with power %s within %s", validator.name, pubKey, power, timeout)
}

func assertSmokeValidatorPowers(t *testing.T, node *smokeValidator, validators []*smokeValidator, power string) {
	t.Helper()
	for _, validator := range validators {
		waitForSmokeValidatorPower(t, node, validator.pubKey, power, 90*time.Second)
	}
}

func querySmokeValidatorPower(ctx context.Context, validator *smokeValidator, pubKey []byte) (string, error) {
	client := &http.Client{Timeout: time.Second}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/validators", validator.rpcPort), nil)
	if err != nil {
		return "", err
	}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("validators response = %s", response.Status)
	}
	var validators struct {
		Result struct {
			Validators []struct {
				PubKey struct {
					Value string `json:"value"`
				} `json:"pub_key"`
				VotingPower string `json:"voting_power"`
			} `json:"validators"`
		} `json:"result"`
	}
	if err := json.NewDecoder(response.Body).Decode(&validators); err != nil {
		return "", err
	}
	for _, candidate := range validators.Result.Validators {
		decoded, err := decodeCometPubKey(candidate.PubKey.Value)
		if err != nil {
			continue
		}
		if bytes.Equal(decoded, pubKey) {
			return candidate.VotingPower, nil
		}
	}
	return "", errors.New("validator public key not found")
}

func decodeCometPubKey(value string) ([]byte, error) {
	if decoded, err := base64.StdEncoding.DecodeString(value); err == nil && len(decoded) > 0 {
		return decoded, nil
	}
	if decoded, err := hex.DecodeString(value); err == nil && len(decoded) > 0 {
		return decoded, nil
	}
	return []byte(value), nil
}

func assertSanitizedBackupArchive(t *testing.T, ctx context.Context, archivePath string) {
	t.Helper()
	command := exec.CommandContext(ctx, "tar", "-tzf", archivePath)
	output, err := command.Output()
	if err != nil {
		t.Fatalf("list backup archive: %v", err)
	}
	listing := string(output)
	for _, forbidden := range []string{
		"/config/node_key.json",
		"/config/priv_validator_key.json",
		"/data/priv_validator_state.json",
		"/keyring-file",
		"/keyring-test",
	} {
		if strings.Contains(listing, forbidden) {
			t.Fatalf("backup archive contains forbidden private/signer artifact %q:\n%s", forbidden, listing)
		}
	}
	if !strings.Contains(listing, "/data/") {
		t.Fatalf("backup archive does not contain chain data:\n%s", listing)
	}
	if !strings.Contains(listing, "/config/genesis.json") {
		t.Fatalf("backup archive does not contain genesis:\n%s", listing)
	}
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

func configureSmokeStateSync(t *testing.T, configPath string, trustHeight int64, trustHash string, rpcServers ...*smokeValidator) {
	t.Helper()
	if len(rpcServers) < 2 {
		t.Fatal("state sync requires at least two trusted RPC servers")
	}
	config, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	rpcs := make([]string, 0, len(rpcServers))
	for _, server := range rpcServers {
		rpcs = append(rpcs, fmt.Sprintf("http://127.0.0.1:%d", server.rpcPort))
	}
	updated := replaceTomlSectionValue(t, string(config), "[statesync]", "enable = false", "enable = true")
	updated = replaceTomlSectionValue(t, updated, "[statesync]", `rpc_servers = ""`, fmt.Sprintf(`rpc_servers = "%s"`, strings.Join(rpcs, ",")))
	updated = replaceTomlSectionValue(t, updated, "[statesync]", "trust_height = 0", fmt.Sprintf("trust_height = %d", trustHeight))
	updated = replaceTomlSectionValue(t, updated, "[statesync]", `trust_hash = ""`, fmt.Sprintf(`trust_hash = "%s"`, trustHash))
	if err := atomicWriteFile(configPath, []byte(updated), 0o600); err != nil {
		t.Fatal(err)
	}
}

func replaceTomlSectionValue(t *testing.T, content, section, original, replacement string) string {
	t.Helper()
	start := strings.Index(content, section+"\n")
	if start < 0 {
		t.Fatalf("missing TOML section %s", section)
	}
	bodyStart := start + len(section) + 1
	nextRel := strings.Index(content[bodyStart:], "\n[")
	end := len(content)
	if nextRel >= 0 {
		end = bodyStart + nextRel
	}
	sectionBody := content[bodyStart:end]
	if strings.Count(sectionBody, original) != 1 {
		t.Fatalf("section %s contains %d copies of %q, want 1", section, strings.Count(sectionBody, original), original)
	}
	sectionBody = strings.Replace(sectionBody, original, replacement, 1)
	return content[:bodyStart] + sectionBody + content[end:]
}

func (validator *smokeValidator) start(ctx context.Context, binary, peers string) error {
	return validator.startWithArgs(ctx, binary, peers)
}

func (validator *smokeValidator) startWithArgs(ctx context.Context, binary, peers string, extraArgs ...string) error {
	if validator.command != nil {
		return errors.New("validator process is already running")
	}
	logFile, err := os.OpenFile(validator.logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	commandArgs := []string{
		"start",
		"--home", validator.home,
		"--rpc.laddr", fmt.Sprintf("tcp://127.0.0.1:%d", validator.rpcPort),
		"--p2p.laddr", fmt.Sprintf("tcp://127.0.0.1:%d", validator.p2pPort),
		"--p2p.persistent_peers", peers,
		"--rpc.pprof_laddr", "",
		"--grpc.enable=false",
		"--api.enable=false",
		"--minimum-gas-prices", "0" + token.BaseDenom,
	}
	commandArgs = append(commandArgs, extraArgs...)
	command := exec.CommandContext(ctx, binary, commandArgs...)
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
			height, err := querySmokeHeight(t.Context(), validator)
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

func waitForSmokeRPC(t *testing.T, validator *smokeValidator, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		if _, err := querySmokeHeight(t.Context(), validator); err == nil {
			return
		} else {
			lastErr = err
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("%s RPC was not ready within %s: %v", validator.name, timeout, lastErr)
}

func smokeHeight(t *testing.T, validator *smokeValidator) int64 {
	t.Helper()
	height, err := querySmokeHeight(t.Context(), validator)
	if err != nil {
		t.Fatalf("query %s height: %v", validator.name, err)
	}
	return height
}

func querySmokeHeight(ctx context.Context, validator *smokeValidator) (int64, error) {
	client := &http.Client{Timeout: time.Second}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/status", validator.rpcPort), nil)
	if err != nil {
		return 0, err
	}
	response, err := client.Do(request)
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

func exportSmokeGenesis(t *testing.T, ctx context.Context, binary string, validator *smokeValidator, minimumInitialHeight int64) smokeExportedGenesis {
	t.Helper()
	exported := exec.CommandContext(ctx, binary, "export", "--home", validator.home)
	exportOutput, err := exported.Output()
	if err != nil {
		t.Fatalf("export recovered %s state: %v", validator.name, err)
	}
	var exportedGenesis smokeExportedGenesis
	if err := json.Unmarshal(exportOutput, &exportedGenesis); err != nil {
		t.Fatalf("decode recovered %s export: %v", validator.name, err)
	}
	if exportedGenesis.InitialHeight <= minimumInitialHeight {
		t.Fatalf("%s exported initial height = %d, want greater than %d", validator.name, exportedGenesis.InitialHeight, minimumInitialHeight)
	}
	return exportedGenesis
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
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("build %s block %d request: %v", validator.name, height, err)
	}
	response, err := client.Do(request)
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

func smokeBlockHash(t *testing.T, validator *smokeValidator, height int64) string {
	t.Helper()
	client := &http.Client{Timeout: time.Second}
	url := fmt.Sprintf("http://127.0.0.1:%d/commit?height=%d", validator.rpcPort, height)
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("build %s commit %d request: %v", validator.name, height, err)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("query %s commit %d: %v", validator.name, height, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("query %s commit %d: %s", validator.name, height, response.Status)
	}
	var commit struct {
		Result struct {
			SignedHeader struct {
				Commit struct {
					BlockID struct {
						Hash string `json:"hash"`
					} `json:"block_id"`
				} `json:"commit"`
			} `json:"signed_header"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    string `json:"data"`
		} `json:"error"`
	}
	if err := json.NewDecoder(response.Body).Decode(&commit); err != nil {
		t.Fatalf("decode %s commit %d: %v", validator.name, height, err)
	}
	if commit.Error != nil {
		t.Fatalf("query %s commit %d returned error %d %s: %s", validator.name, height, commit.Error.Code, commit.Error.Message, commit.Error.Data)
	}
	return commit.Result.SignedHeader.Commit.BlockID.Hash
}
