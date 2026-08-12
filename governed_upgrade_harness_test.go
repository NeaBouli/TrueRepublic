package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/log"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"truerepublic/token"
	"truerepublic/x/truedemocracy"
)

type governedUpgradeState struct {
	height     int64
	marker     []byte
	doneHeight int64
	plan       upgradetypes.Plan
	planFound  bool
}

func TestGovernedUpgradeMultiValidatorHaltFailureRecovery(t *testing.T) {
	if testing.Short() || strings.TrimSpace(os.Getenv(multiValidatorSmokeEnv)) != "1" {
		t.Skipf("set %s=1 to run the governed multi-validator upgrade harness", multiValidatorSmokeEnv)
	}
	ctx := t.Context()
	baselineBinary := filepath.Join(t.TempDir(), "truerepublicd-upgrade-base")
	failureBinary := filepath.Join(t.TempDir(), "truerepublicd-upgrade-failure")
	candidateBinary := filepath.Join(t.TempDir(), "truerepublicd-v0.4.1")
	buildGovernedUpgradeBinary(t, ctx, baselineBinary, "v0.4.0-upgrade-base", "")
	buildGovernedUpgradeBinary(t, ctx, failureBinary, governedUpgradeFailureFixtureV041, governedUpgradeFailureFixtureV041)
	buildGovernedUpgradeBinary(t, ctx, candidateBinary, governedUpgradePlanV041, governedUpgradePlanV041)

	const chainID = "truerepublic-governed-upgrade-1"
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

	voters := make([]smokeAccount, 4)
	for i := range voters {
		voters[i] = addSmokeKey(t, ctx, baselineBinary, validators[0].home,
			fmt.Sprintf("upgrade-voter-%d", i+1), uint64(20+i), 10*token.WholeTokenBaseUnits)
	}
	sharedGenesis := addGovernedUpgradeDomain(t, buildSharedSmokeGenesis(t, chainID, validators, voters...), voters)
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
			t.Fatalf("start %s upgrade base: %v", validator.name, err)
		}
	}
	waitForSmokeHeight(t, validators, 2, 90*time.Second)
	// Each vote is committed in its own block. Keep enough headroom for all
	// three transactions while preserving the consensus-enforced 10-block lead.
	targetHeight := smokeHeight(t, validators[0]) + 16
	for i := 0; i < 3; i++ {
		runSmokeTx(t, ctx, baselineBinary, validators[0], &voters[i], chainID,
			"vote-software-upgrade", governedUpgradePlanV041, strconv.FormatInt(targetHeight, 10), "gh184-deterministic-v0.4.1")
	}
	waitForSmokeHeight(t, validators, targetHeight-2, 180*time.Second)
	for _, validator := range validators {
		waitForGovernedUpgradeLog(t, validator, `UPGRADE "v0.4.1" NEEDED`, 120*time.Second)
	}
	stopGovernedUpgradeValidators(t, validators)

	for _, validator := range validators {
		if err := validator.start(ctx, failureBinary, persistentPeers(validator, validators)); err != nil {
			t.Fatalf("start %s failure fixture: %v", validator.name, err)
		}
	}
	for _, validator := range validators {
		waitForGovernedUpgradeLog(t, validator, "intentional GH-184 migration failure", 120*time.Second)
	}
	stopGovernedUpgradeValidators(t, validators)

	for _, validator := range validators {
		if err := validator.start(ctx, candidateBinary, persistentPeers(validator, validators)); err != nil {
			t.Fatalf("start %s fixed candidate: %v", validator.name, err)
		}
	}
	waitForSmokeHeight(t, validators, targetHeight+2, 150*time.Second)
	assertCommonAppHash(t, validators, targetHeight)
	assertCommonAppHash(t, validators, targetHeight+2)
	assertSmokeValidatorPowers(t, validators[0], validators, "1")
	for _, validator := range validators {
		if err := validator.stop(true); err != nil {
			t.Fatalf("stop %s fixed candidate: %v", validator.name, err)
		}
	}

	for _, validator := range validators {
		if err := validator.start(ctx, candidateBinary, persistentPeers(validator, validators)); err != nil {
			t.Fatalf("restart %s candidate for exact-once proof: %v", validator.name, err)
		}
	}
	finalHeight := targetHeight + 4
	waitForSmokeHeight(t, validators, finalHeight, 120*time.Second)
	assertCommonAppHash(t, validators, finalHeight)
	for _, validator := range validators {
		if err := validator.stop(true); err != nil {
			t.Fatalf("stop %s exact-once candidate: %v", validator.name, err)
		}
		state := readGovernedUpgradeState(t, validator.home)
		if state.height < finalHeight || !bytes.Equal(state.marker, []byte{1}) || state.doneHeight != targetHeight || state.planFound {
			t.Fatalf("%s replayed or resurrected completed upgrade: %+v", validator.name, state)
		}
	}
}

func buildGovernedUpgradeBinary(t *testing.T, ctx context.Context, path, binaryVersion, plan string) {
	t.Helper()
	ldflags := "-X main.version=" + binaryVersion
	if plan != "" {
		ldflags += " -X main.upgradePlan=" + plan
	}
	command := exec.CommandContext(ctx, "go", "build", "-ldflags", ldflags, "-o", path, ".")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build %s: %v\n%s", binaryVersion, err, output)
	}
}

func addGovernedUpgradeDomain(t *testing.T, genesisJSON []byte, voters []smokeAccount) []byte {
	t.Helper()
	var genesis genutiltypes.AppGenesis
	if err := json.Unmarshal(genesisJSON, &genesis); err != nil {
		t.Fatal(err)
	}
	var state map[string]json.RawMessage
	if err := json.Unmarshal(genesis.AppState, &state); err != nil {
		t.Fatal(err)
	}
	var democracy truedemocracy.GenesisState
	if err := json.Unmarshal(state[truedemocracy.ModuleName], &democracy); err != nil {
		t.Fatal(err)
	}
	members := make([]string, len(voters))
	for i, voter := range voters {
		members[i] = voter.address
	}
	admin, err := sdk.AccAddressFromBech32(members[0])
	if err != nil {
		t.Fatal(err)
	}
	democracy.Domains = append(democracy.Domains, truedemocracy.Domain{
		Name: truedemocracy.ReservedGovernanceDomain, Admin: admin, Members: members,
		Treasury: sdk.NewCoins(), Issues: []truedemocracy.Issue{},
		Options: truedemocracy.DomainOptions{AdminElectable: true}, PermissionReg: []string{},
	})
	updatedDemocracy, err := json.Marshal(democracy)
	if err != nil {
		t.Fatal(err)
	}
	state[truedemocracy.ModuleName] = updatedDemocracy
	genesis.AppState, err = json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	updated, err := json.MarshalIndent(genesis, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return updated
}

func waitForGovernedUpgradeLog(t *testing.T, validator *smokeValidator, logNeedle string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		content, err := os.ReadFile(validator.logPath)
		logText := string(content)
		matched := err == nil && strings.Contains(logText, logNeedle)
		if err == nil && strings.Contains(logNeedle, "UPGRADE") {
			// The JSON logger escapes the quoted plan name. Match the semantic
			// halt markers as well as the plain-text representation.
			matched = matched || (strings.Contains(logText, "UPGRADE") &&
				strings.Contains(logText, governedUpgradePlanV041) &&
				strings.Contains(logText, "NEEDED"))
		}
		if matched {
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("%s did not record governed consensus halt %q within %s", validator.name, logNeedle, timeout)
}

func stopGovernedUpgradeValidators(t *testing.T, validators []*smokeValidator) {
	t.Helper()
	// Keep every node alive until the full validator set recorded the same
	// consensus result. Early shutdown would remove quorum from the final node.
	for _, validator := range validators {
		if err := validator.stop(false); err != nil {
			t.Fatalf("stop %s after governed halt: %v", validator.name, err)
		}
	}
}

func readGovernedUpgradeState(t *testing.T, home string) governedUpgradeState {
	t.Helper()
	database, err := dbm.NewDB("application", dbm.GoLevelDBBackend, filepath.Join(home, "data"))
	if err != nil {
		t.Fatal(err)
	}
	app := NewTrueRepublicApp(log.NewNopLogger(), database, home)
	defer func() { _ = app.Close() }()
	state := governedUpgradeState{height: app.LastBlockHeight()}
	ctx := app.NewUncachedContext(false, cmtproto.Header{Height: state.height})
	state.marker = append([]byte(nil), ctx.KVStore(app.keys[truedemocracy.ModuleName]).Get(governedUpgradeMarkerV041)...)
	state.doneHeight, err = app.upgradeKeeper.GetDoneHeight(ctx, governedUpgradePlanV041)
	if err != nil {
		t.Fatal(err)
	}
	state.plan, err = app.upgradeKeeper.GetUpgradePlan(ctx)
	switch {
	case err == nil:
		state.planFound = true
	case errors.Is(err, upgradetypes.ErrNoUpgradePlanFound):
	case err != nil:
		t.Fatal(err)
	}
	return state
}
