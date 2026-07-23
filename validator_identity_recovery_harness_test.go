package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestValidatorIdentityColdFailover(t *testing.T) {
	if os.Getenv(multiValidatorSmokeEnv) != "1" && os.Getenv("TRUEREPUBLIC_VALIDATOR_IDENTITY_RECOVERY") != "1" {
		t.Skipf("set %s=1 or TRUEREPUBLIC_VALIDATOR_IDENTITY_RECOVERY=1 to run the validator identity failover harness", multiValidatorSmokeEnv)
	}
	ctx := t.Context()

	binary := filepath.Join(t.TempDir(), "truerepublicd")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build daemon: %v\n%s", err, output)
	}

	const chainID = "truerepublic-validator-identity-recovery-1"
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

	sharedGenesis := buildSharedSmokeGenesis(t, chainID, validators)
	for _, validator := range validators {
		if err := atomicWriteFile(filepath.Join(validator.home, "config", "genesis.json"), sharedGenesis, 0o600); err != nil {
			t.Fatalf("write %s shared genesis: %v", validator.name, err)
		}
	}

	var recovery *smokeValidator
	t.Cleanup(func() {
		for _, validator := range validators {
			_ = validator.stop(false)
		}
		if recovery != nil {
			_ = recovery.stop(false)
		}
		if t.Failed() {
			for _, validator := range validators {
				validator.logContents(t)
			}
			if recovery != nil {
				recovery.logContents(t)
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

	source := validators[len(validators)-1]
	if err := source.stop(true); err != nil {
		t.Fatalf("stop source %s: %v", source.name, err)
	}
	if _, err := querySmokeHeight(ctx, source); err == nil {
		t.Fatalf("source %s should remain stopped", source.name)
	}
	sourceIdentity := readValidatorIdentityState(t, source)
	beforeRecoveryHeight := smokeHeight(t, validators[0])

	recovery = &smokeValidator{
		name:    fmt.Sprintf("%s-recovered", source.name),
		home:    filepath.Join(t.TempDir(), fmt.Sprintf("%s-recovered", source.name)),
		rpcPort: freeTCPPort(t),
		p2pPort: freeTCPPort(t),
		logPath: filepath.Join(t.TempDir(), fmt.Sprintf("%s-recovered.log", source.name)),
	}
	recovery.operatorAddr = smokeOperatorAddress(recovery.name)
	initCmd := exec.CommandContext(ctx, binary, "init", recovery.name, "--chain-id", chainID, "--home", recovery.home, "--bootstrap-operator", recovery.operatorAddr)
	if output, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("init %s: %v\n%s", recovery.name, err, output)
	}
	configureLocalhostSmokeP2P(t, filepath.Join(recovery.home, "config", "config.toml"))
	if err := atomicWriteFile(filepath.Join(recovery.home, "config", "genesis.json"), sharedGenesis, 0o600); err != nil {
		t.Fatalf("write %s shared genesis: %v", recovery.name, err)
	}
	backupDir := t.TempDir()
	backupCmd := exec.CommandContext(ctx, "bash", filepath.Join("scripts", "backup.sh"), backupDir)
	backupCmd.Env = append(os.Environ(), "CHAIN_HOME="+source.home)
	if output, err := backupCmd.CombinedOutput(); err != nil {
		t.Fatalf("create %s sanitized backup: %v\n%s", source.name, err, output)
	}
	backupMatches, err := filepath.Glob(filepath.Join(backupDir, "truerepublic_*.tar.gz"))
	if err != nil {
		t.Fatal(err)
	}
	if len(backupMatches) != 1 {
		t.Fatalf("backup artifact count = %d, want 1: %v", len(backupMatches), backupMatches)
	}
	assertSanitizedBackupArchive(t, ctx, backupMatches[0])
	restoreCmd := exec.CommandContext(ctx, "bash", filepath.Join("scripts", "restore.sh"), backupMatches[0], recovery.home)
	if output, err := restoreCmd.CombinedOutput(); err != nil {
		t.Fatalf("restore %s sanitized data: %v\n%s", recovery.name, err, output)
	}

	if err := atomicWriteFile(filepath.Join(recovery.home, "config", "priv_validator_key.json"), sourceIdentity.validatorKey, 0o600); err != nil {
		t.Fatalf("write %s validator key: %v", recovery.name, err)
	}
	if err := atomicWriteFile(filepath.Join(recovery.home, "data", "priv_validator_state.json"), sourceIdentity.lastSignRaw, 0o600); err != nil {
		t.Fatalf("write %s validator state: %v", recovery.name, err)
	}
	assertPathMode(t, filepath.Join(recovery.home, "config", "priv_validator_key.json"), 0o600)
	assertPathMode(t, filepath.Join(recovery.home, "data", "priv_validator_state.json"), 0o600)

	recovery.pubKey = append([]byte(nil), source.pubKey...)
	recovery.nodeID = smokeNodeID(t, ctx, binary, recovery.home)
	if recovery.nodeID == source.nodeID {
		t.Fatalf("recovery %s reused source node id", recovery.name)
	}
	recoveredState := readValidatorIdentityState(t, recovery)
	if !bytes.Equal(recoveredState.validatorKey, sourceIdentity.validatorKey) {
		t.Fatalf("%s validator key changed after transfer", recovery.name)
	}
	if !bytes.Equal(recoveredState.lastSignRaw, sourceIdentity.lastSignRaw) {
		t.Fatalf("%s signer state changed before the recovered signer started", recovery.name)
	}
	if bytes.Equal(recoveredState.nodeKey, sourceIdentity.nodeKey) {
		t.Fatalf("%s should keep a distinct node key after recovery home init", recovery.name)
	}

	if err := recovery.start(ctx, binary, persistentPeers(recovery, validators)); err != nil {
		t.Fatalf("start %s: %v", recovery.name, err)
	}
	recoverySet := append([]*smokeValidator{}, validators[:len(validators)-1]...)
	recoverySet = append(recoverySet, recovery)

	postFailureHeight := beforeRecoveryHeight + 2
	waitForSmokeHeight(t, recoverySet, postFailureHeight, 120*time.Second)
	assertCommonAppHash(t, recoverySet, postFailureHeight)
	afterRecoveryState := readValidatorIdentityState(t, recovery)
	assertSigningStateAdvanced(t, recovery.name, afterRecoveryState.lastSign, recoveredState.lastSign)
	assertSmokeValidatorPowers(t, validators[0], recoverySet, "1")
	if source.command != nil {
		t.Fatalf("source %s restarted while the recovery signer was active", source.name)
	}
	if _, err := querySmokeHeight(ctx, source); err == nil {
		t.Fatalf("source %s RPC became reachable while the recovery signer was active", source.name)
	}
	for _, validator := range recoverySet {
		if err := validator.stop(true); err != nil {
			t.Fatalf("stop %s after convergence: %v", validator.name, err)
		}
	}

	exported := exportSmokeGenesis(t, ctx, binary, recovery, postFailureHeight)
	exportApp := newGenesisTestApp(t)
	if err := validateLedgerGenesis(exportApp.appCodec, exported.AppState); err != nil {
		t.Fatalf("identity-recovery export is not exactly bank-backed: %v", err)
	}
	importApp := newGenesisTestApp(t)
	if err := initGenesisApp(importApp, exported.AppState); err != nil {
		t.Fatalf("re-import identity-recovery export: %v", err)
	}
}

func assertSigningStateAdvanced(t *testing.T, name string, got, previous validatorSignPosition) {
	t.Helper()
	if got.Height > previous.Height ||
		(got.Height == previous.Height && got.Round > previous.Round) ||
		(got.Height == previous.Height && got.Round == previous.Round && got.Step > previous.Step) {
		return
	}
	t.Fatalf("%s signing state did not advance from %d/%d/%d: got %d/%d/%d", name,
		previous.Height, previous.Round, previous.Step, got.Height, got.Round, got.Step)
}

func smokeNodeID(t *testing.T, ctx context.Context, binary, home string) string {
	t.Helper()
	nodeIDCmd := exec.CommandContext(ctx, binary, "comet", "show-node-id", "--home", home)
	nodeID, err := nodeIDCmd.Output()
	if err != nil {
		t.Fatalf("read node id for %s: %v", home, err)
	}
	return strings.TrimSpace(string(nodeID))
}

func assertPathMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if info.Mode().Perm() != want {
		t.Fatalf("file mode %s = %#o, want %#o", path, info.Mode().Perm(), want)
	}
}
