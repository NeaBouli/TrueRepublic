package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"truerepublic/token"
	"truerepublic/x/truedemocracy"
)

type validatorIdentityState struct {
	nodeKey      []byte
	validatorKey []byte
	lastSignRaw  []byte
	lastSign     validatorSignPosition
}

type validatorSignPosition struct {
	Height int64
	Round  int32
	Step   int8
}

func TestMultiValidatorPersistedBinaryUpgradeRollback(t *testing.T) {
	if os.Getenv(multiValidatorSmokeEnv) != "1" {
		t.Skipf("set %s=1 to run the multi-validator process harness", multiValidatorSmokeEnv)
	}
	ctx := t.Context()

	baselineBinary := filepath.Join(t.TempDir(), "truerepublicd-baseline")
	compatibleBinary := filepath.Join(t.TempDir(), "truerepublicd-compatible")
	buildVersionedSmokeBinary(t, ctx, baselineBinary, "v0.4.0-rollback-baseline")
	buildVersionedSmokeBinary(t, ctx, compatibleBinary, "v0.4.1-compatible-drill")
	assertSmokeBinaryVersion(t, ctx, baselineBinary, "v0.4.0-rollback-baseline")
	assertSmokeBinaryVersion(t, ctx, compatibleBinary, "v0.4.1-compatible-drill")

	const chainID = "truerepublic-upgrade-rollback-1"
	validators := make([]*smokeValidator, 4)
	for i := range validators {
		validator := &smokeValidator{
			name:    fmt.Sprintf("validator-%d", i+1),
			home:    filepath.Join(t.TempDir(), fmt.Sprintf("node-%d", i+1)),
			rpcPort: freeTCPPort(t),
			p2pPort: freeTCPPort(t),
			logPath: filepath.Join(t.TempDir(), fmt.Sprintf("validator-%d.log", i+1)),
		}
		initSmokeValidator(t, ctx, baselineBinary, chainID, validator)
		validators[i] = validator
	}

	admin := addSmokeKey(t, ctx, baselineBinary, validators[0].home, "upgrade-admin", 4, 800_000*token.WholeTokenBaseUnits)
	sharedGenesis := buildSharedSmokeGenesis(t, chainID, validators, admin)
	for _, validator := range validators {
		if err := atomicWriteFile(filepath.Join(validator.home, "config", "genesis.json"), sharedGenesis, 0o600); err != nil {
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

	for _, validator := range validators {
		if err := validator.start(ctx, baselineBinary, persistentPeers(validator, validators)); err != nil {
			t.Fatalf("start %s baseline: %v", validator.name, err)
		}
	}
	waitForSmokeHeight(t, validators, 2, 90*time.Second)
	runSmokeTx(t, ctx, baselineBinary, validators[0], &admin, chainID,
		"create-domain", "UpgradeRollback", fmt.Sprintf("%d%s", 500_000*token.WholeTokenBaseUnits, token.BaseDenom))
	preUpgradeHeight := smokeHeight(t, validators[0]) + 2
	waitForSmokeHeight(t, validators, preUpgradeHeight, 90*time.Second)
	assertCommonAppHash(t, validators, preUpgradeHeight)
	assertSmokeValidatorPowers(t, validators[0], validators, "1")

	beforeUpgrade := make(map[string]validatorIdentityState, len(validators))
	for i, validator := range validators {
		if err := validator.stop(true); err != nil {
			t.Fatalf("stop %s for compatible upgrade: %v", validator.name, err)
		}
		beforeUpgrade[validator.name] = readValidatorIdentityState(t, validator)
		if err := validator.start(ctx, compatibleBinary, persistentPeers(validator, validators)); err != nil {
			t.Fatalf("start %s compatible candidate: %v", validator.name, err)
		}
		targetHeight := smokeHeight(t, validators[(i+1)%len(validators)]) + 2
		waitForSmokeHeight(t, validators, targetHeight, 120*time.Second)
		assertCommonAppHash(t, validators, targetHeight)
	}
	assertCommonAppHash(t, validators, preUpgradeHeight)
	assertSmokeValidatorPowers(t, validators[0], validators, "1")

	failedValidator := validators[0]
	if err := failedValidator.stop(true); err != nil {
		t.Fatalf("stop %s for failed candidate drill: %v", failedValidator.name, err)
	}
	beforeFailure := readValidatorIdentityState(t, failedValidator)
	failedBinary := writeFailFastCandidate(t)
	failedStart := exec.CommandContext(ctx, failedBinary, "start", "--home", failedValidator.home)
	failedOutput, failedErr := failedStart.CombinedOutput()
	if failedErr == nil {
		t.Fatal("fail-fast candidate unexpectedly started successfully")
	}
	if !strings.Contains(string(failedOutput), "intentional pre-open failure") {
		t.Fatalf("fail-fast candidate output = %q", failedOutput)
	}
	if _, err := querySmokeHeight(ctx, failedValidator); err == nil {
		t.Fatal("failed candidate unexpectedly exposed the validator RPC endpoint")
	}
	afterFailure := readValidatorIdentityState(t, failedValidator)
	assertIdentityStateEqual(t, failedValidator.name, afterFailure, beforeFailure)

	if err := failedValidator.start(ctx, baselineBinary, persistentPeers(failedValidator, validators)); err != nil {
		t.Fatalf("rollback %s to baseline: %v", failedValidator.name, err)
	}
	recoveryHeight := smokeHeight(t, validators[1]) + 2
	waitForSmokeHeight(t, validators, recoveryHeight, 120*time.Second)
	assertCommonAppHash(t, validators, recoveryHeight)
	assertSmokeValidatorPowers(t, failedValidator, validators, "1")

	for _, validator := range validators[1:] {
		if err := validator.stop(true); err != nil {
			t.Fatalf("stop %s for baseline rollback: %v", validator.name, err)
		}
		if err := validator.start(ctx, baselineBinary, persistentPeers(validator, validators)); err != nil {
			t.Fatalf("rollback %s to baseline: %v", validator.name, err)
		}
		targetHeight := smokeHeight(t, failedValidator) + 2
		waitForSmokeHeight(t, validators, targetHeight, 120*time.Second)
		assertCommonAppHash(t, validators, targetHeight)
	}
	finalHeight := smokeHeight(t, validators[0]) + 2
	waitForSmokeHeight(t, validators, finalHeight, 90*time.Second)
	assertCommonAppHash(t, validators, preUpgradeHeight)
	assertCommonAppHash(t, validators, finalHeight)
	assertSmokeValidatorPowers(t, validators[0], validators, "1")

	for _, validator := range validators {
		if err := validator.stop(true); err != nil {
			t.Fatalf("stop recovered %s: %v", validator.name, err)
		}
		afterRollback := readValidatorIdentityState(t, validator)
		assertIdentityKeysEqual(t, validator.name, afterRollback, beforeUpgrade[validator.name])
		assertSigningStateNotRegressed(t, validator.name, afterRollback.lastSign, beforeUpgrade[validator.name].lastSign)
	}

	exportedGenesis := exportSmokeGenesis(t, ctx, baselineBinary, validators[0], preUpgradeHeight)
	assertExportedDomain(t, exportedGenesis, "UpgradeRollback")
	exportApp := newGenesisTestApp(t)
	if err := validateLedgerGenesis(exportApp.appCodec, exportedGenesis.AppState); err != nil {
		t.Fatalf("rollback export is not exactly bank-backed: %v", err)
	}
	importApp := newGenesisTestApp(t)
	if err := initGenesisApp(importApp, exportedGenesis.AppState); err != nil {
		t.Fatalf("re-import rollback export: %v", err)
	}
}

func buildVersionedSmokeBinary(t *testing.T, ctx context.Context, path, binaryVersion string) {
	t.Helper()
	command := exec.CommandContext(ctx, "go", "build", "-ldflags", "-X main.version="+binaryVersion, "-o", path, ".")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build %s: %v\n%s", binaryVersion, err, output)
	}
}

func assertSmokeBinaryVersion(t *testing.T, ctx context.Context, binary, expected string) {
	t.Helper()
	command := exec.CommandContext(ctx, binary, "version")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("read %s version: %v\n%s", binary, err, output)
	}
	if !strings.Contains(string(output), expected) {
		t.Fatalf("%s version output = %q, want %q", binary, output, expected)
	}
}

func writeFailFastCandidate(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "truerepublicd-fail-fast")
	content := []byte("#!/bin/sh\necho 'intentional pre-open failure' >&2\nexit 42\n")
	if err := os.WriteFile(path, content, 0o700); err != nil {
		t.Fatalf("write fail-fast candidate: %v", err)
	}
	return path
}

func readValidatorIdentityState(t *testing.T, validator *smokeValidator) validatorIdentityState {
	t.Helper()
	read := func(path string) []byte {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		return content
	}
	state := validatorIdentityState{
		nodeKey:      read(filepath.Join(validator.home, "config", "node_key.json")),
		validatorKey: read(filepath.Join(validator.home, "config", "priv_validator_key.json")),
	}
	lastSignJSON := read(filepath.Join(validator.home, "data", "priv_validator_state.json"))
	var encoded struct {
		Height string `json:"height"`
		Round  int32  `json:"round"`
		Step   int8   `json:"step"`
	}
	if err := json.Unmarshal(lastSignJSON, &encoded); err != nil {
		t.Fatalf("decode %s signing state: %v", validator.name, err)
	}
	height, err := strconv.ParseInt(encoded.Height, 10, 64)
	if err != nil {
		t.Fatalf("decode %s signing height %q: %v", validator.name, encoded.Height, err)
	}
	state.lastSignRaw = lastSignJSON
	state.lastSign = validatorSignPosition{Height: height, Round: encoded.Round, Step: encoded.Step}
	return state
}

func assertIdentityStateEqual(t *testing.T, name string, got, want validatorIdentityState) {
	t.Helper()
	assertIdentityKeysEqual(t, name, got, want)
	if !bytes.Equal(got.lastSignRaw, want.lastSignRaw) {
		t.Fatalf("%s signing state changed during fail-before-open attempt", name)
	}
}

func assertIdentityKeysEqual(t *testing.T, name string, got, want validatorIdentityState) {
	t.Helper()
	if !bytes.Equal(got.nodeKey, want.nodeKey) {
		t.Fatalf("%s node identity key changed", name)
	}
	if !bytes.Equal(got.validatorKey, want.validatorKey) {
		t.Fatalf("%s validator identity key changed", name)
	}
}

func assertSigningStateNotRegressed(t *testing.T, name string, got, want validatorSignPosition) {
	t.Helper()
	if got.Height < want.Height ||
		(got.Height == want.Height && got.Round < want.Round) ||
		(got.Height == want.Height && got.Round == want.Round && got.Step < want.Step) {
		t.Fatalf("%s signing state regressed from %d/%d/%d to %d/%d/%d", name,
			want.Height, want.Round, want.Step, got.Height, got.Round, got.Step)
	}
}

func assertExportedDomain(t *testing.T, exported smokeExportedGenesis, domainName string) {
	t.Helper()
	raw, ok := exported.AppState[truedemocracy.ModuleName]
	if !ok {
		t.Fatalf("export is missing %s state", truedemocracy.ModuleName)
	}
	var state truedemocracy.GenesisState
	if err := json.Unmarshal(raw, &state); err != nil {
		t.Fatalf("decode exported %s state: %v", truedemocracy.ModuleName, err)
	}
	for _, domain := range state.Domains {
		if domain.Name == domainName {
			return
		}
	}
	t.Fatalf("export is missing persisted domain %q", domainName)
}
