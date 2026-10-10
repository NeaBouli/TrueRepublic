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

// governedUpgradeV042BaseCommit is exact main before GH-306: its truedemocracy
// module runs at consensus version 2, which is the only source version the
// v0.4.2 plan accepts.
const governedUpgradeV042BaseCommit = "1283a4452d36784ea962f7fc8ba13f9ad45472cd"

const governedUpgradeV042StoneDomain = "V042Stones"

type governedUpgradeV042State struct {
	Height        int64                             `json:"height"`
	Marker        []byte                            `json:"marker"`
	DoneHeight    int64                             `json:"done_height"`
	PlanFound     bool                              `json:"plan_found"`
	ModuleVersion uint64                            `json:"module_version"`
	RewardRecords []truedemocracy.StoneRewardRecord `json:"reward_records"`
}

const governedUpgradeV042ProbeHomeEnv = "TRUEREPUBLIC_V042_STATE_PROBE_HOME"

const governedUpgradeV042ProbePrefix = "V042_STATE_PROBE "

// TestGovernedUpgradeV042FromVersion2MultiValidator proves the GH-306 F2 fix on
// persisted multi-validator state: a version-2 chain built from the pre-GH-306
// source halts at the governed v0.4.2 height, a failing-fixture binary
// discards its cached write, and the fixed v0.4.2 candidate migrates 2→3
// exactly once, baselining the stones placed before the upgrade.
func TestGovernedUpgradeV042FromVersion2MultiValidator(t *testing.T) {
	if testing.Short() || strings.TrimSpace(os.Getenv(multiValidatorSmokeEnv)) != "1" {
		t.Skipf("set %s=1 to run the governed v0.4.2 multi-validator upgrade harness", multiValidatorSmokeEnv)
	}
	ctx := t.Context()
	baselineBinary := filepath.Join(t.TempDir(), "truerepublicd-v2-base")
	failureBinary := filepath.Join(t.TempDir(), "truerepublicd-v042-failure")
	candidateBinary := filepath.Join(t.TempDir(), "truerepublicd-v0.4.2")
	buildGovernedUpgradeBinaryFromCommit(t, ctx, baselineBinary, governedUpgradeV042BaseCommit, "v0.4.1-v2-base")
	buildGovernedUpgradeBinary(t, ctx, failureBinary, governedUpgradeFailureFixtureV042, governedUpgradeFailureFixtureV042)
	buildGovernedUpgradeBinary(t, ctx, candidateBinary, governedUpgradePlanV042, governedUpgradePlanV042)

	const chainID = "truerepublic-governed-upgrade-v042"
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
			fmt.Sprintf("v042-voter-%d", i+1), uint64(20+i), 10*token.WholeTokenBaseUnits)
	}
	sharedGenesis := addGovernedUpgradeDomain(t, buildSharedSmokeGenesis(t, chainID, validators, voters...), voters)
	sharedGenesis = addGovernedUpgradeV042StoneDomain(t, sharedGenesis, voters)
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
			t.Fatalf("start %s v2 base: %v", validator.name, err)
		}
	}
	waitForSmokeHeight(t, validators, 2, 90*time.Second)

	// A pre-upgrade placement on the version-2 chain: the legacy stone key
	// exists without a consumed reward marker until the 2→3 baseline.
	stoneVoter := &voters[3]
	runSmokeTx(t, ctx, baselineBinary, validators[0], stoneVoter, chainID,
		"place-stone-issue", governedUpgradeV042StoneDomain, "Harbor")

	targetHeight := smokeHeight(t, validators[0]) + 16
	for i := 0; i < 3; i++ {
		runSmokeTx(t, ctx, baselineBinary, validators[0], &voters[i], chainID,
			"vote-software-upgrade", governedUpgradePlanV042, strconv.FormatInt(targetHeight, 10), "gh306-deterministic-v0.4.2")
	}
	waitForSmokeHeight(t, validators, targetHeight-2, 180*time.Second)
	for _, validator := range validators {
		waitForGovernedUpgradeV042Log(t, validator, "UPGRADE", "NEEDED", 120*time.Second)
	}
	stopGovernedUpgradeValidators(t, validators)
	for _, validator := range validators {
		state := readGovernedUpgradeV042StateIsolated(t, ctx, validator.home)
		if state.ModuleVersion != governedUpgradeV042FromVersion || len(state.Marker) != 0 || len(state.RewardRecords) != 0 {
			t.Fatalf("%s halted state is not the untouched version-2 state: %+v", validator.name, state)
		}
	}

	for _, validator := range validators {
		if err := validator.start(ctx, failureBinary, persistentPeers(validator, validators)); err != nil {
			t.Fatalf("start %s v0.4.2 failure fixture: %v", validator.name, err)
		}
	}
	for _, validator := range validators {
		waitForGovernedUpgradeV042Log(t, validator, "intentional GH-306 migration failure", "", 120*time.Second)
	}
	stopGovernedUpgradeValidators(t, validators)
	for _, validator := range validators {
		state := readGovernedUpgradeV042StateIsolated(t, ctx, validator.home)
		if state.ModuleVersion != governedUpgradeV042FromVersion || len(state.Marker) != 0 || state.DoneHeight != 0 {
			t.Fatalf("%s failed v0.4.2 migration leaked state: %+v", validator.name, state)
		}
	}

	for _, validator := range validators {
		if err := validator.start(ctx, candidateBinary, persistentPeers(validator, validators)); err != nil {
			t.Fatalf("start %s v0.4.2 candidate: %v", validator.name, err)
		}
	}
	waitForSmokeHeight(t, validators, targetHeight+2, 120*time.Second)
	assertCommonAppHash(t, validators, targetHeight)
	assertCommonAppHash(t, validators, targetHeight+2)
	for _, validator := range validators {
		if err := validator.stop(true); err != nil {
			t.Fatalf("stop %s v0.4.2 candidate: %v", validator.name, err)
		}
	}

	for _, validator := range validators {
		if err := validator.start(ctx, candidateBinary, persistentPeers(validator, validators)); err != nil {
			t.Fatalf("restart %s v0.4.2 candidate for exact-once proof: %v", validator.name, err)
		}
	}
	finalHeight := targetHeight + 4
	waitForSmokeHeight(t, validators, finalHeight, 120*time.Second)
	assertCommonAppHash(t, validators, finalHeight)
	wantRecord := truedemocracy.StoneRewardRecord{DomainName: governedUpgradeV042StoneDomain, MemberAddr: stoneVoter.address}
	for _, validator := range validators {
		if err := validator.stop(true); err != nil {
			t.Fatalf("stop %s exact-once v0.4.2 candidate: %v", validator.name, err)
		}
		state := readGovernedUpgradeV042State(t, validator.home)
		if state.Height < finalHeight || !bytes.Equal(state.Marker, []byte{1}) || state.DoneHeight != targetHeight || state.PlanFound {
			t.Fatalf("%s replayed or resurrected completed v0.4.2 upgrade: %+v", validator.name, state)
		}
		if state.ModuleVersion != 3 {
			t.Fatalf("%s truedemocracy module version = %d, want 3", validator.name, state.ModuleVersion)
		}
		if len(state.RewardRecords) != 1 || state.RewardRecords[0] != wantRecord {
			t.Fatalf("%s 2→3 baseline records = %+v, want exactly %+v", validator.name, state.RewardRecords, wantRecord)
		}
	}
}

// buildGovernedUpgradeBinaryFromCommit builds an infrastructure binary (no
// upgrade plan) from an exact historical commit, extracted with git archive so
// the working tree and other worktrees stay untouched.
func buildGovernedUpgradeBinaryFromCommit(t *testing.T, ctx context.Context, path, commit, binaryVersion string) {
	t.Helper()
	source := t.TempDir()
	archive := exec.CommandContext(ctx, "git", "archive", "--format=tar", commit)
	extract := exec.CommandContext(ctx, "tar", "-x", "-C", source)
	pipe, err := archive.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	extract.Stdin = pipe
	var archiveErr, extractErr bytes.Buffer
	archive.Stderr = &archiveErr
	extract.Stderr = &extractErr
	if err := extract.Start(); err != nil {
		t.Fatal(err)
	}
	if err := archive.Run(); err != nil {
		t.Fatalf("git archive %s: %v\n%s", commit, err, archiveErr.String())
	}
	if err := extract.Wait(); err != nil {
		t.Fatalf("extract %s: %v\n%s", commit, err, extractErr.String())
	}
	command := exec.CommandContext(ctx, "go", "build", "-ldflags", "-X main.version="+binaryVersion, "-o", path, ".")
	command.Dir = source
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build %s from %s: %v\n%s", binaryVersion, commit, err, output)
	}
}

// addGovernedUpgradeV042StoneDomain adds a treasury-less domain with one issue
// so a pre-upgrade stone placement exists at the version boundary.
func addGovernedUpgradeV042StoneDomain(t *testing.T, genesisJSON []byte, voters []smokeAccount) []byte {
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
	now := time.Now().Unix()
	democracy.Domains = append(democracy.Domains, truedemocracy.Domain{
		Name: governedUpgradeV042StoneDomain, Admin: admin, Members: members,
		Treasury: sdk.NewCoins(),
		Issues: []truedemocracy.Issue{{
			Name: "Harbor", Suggestions: []truedemocracy.Suggestion{}, CreationDate: now, LastActivityAt: now,
		}},
		PermissionReg: []string{},
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

func waitForGovernedUpgradeV042Log(t *testing.T, validator *smokeValidator, needle, alsoNeeded string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		content, err := os.ReadFile(validator.logPath)
		text := string(content)
		if err == nil && strings.Contains(text, needle) &&
			(alsoNeeded == "" || (strings.Contains(text, alsoNeeded) && strings.Contains(text, governedUpgradePlanV042))) {
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("%s log did not contain %q within %s", validator.name, needle, timeout)
}

func readGovernedUpgradeV042State(t *testing.T, home string) governedUpgradeV042State {
	t.Helper()
	database, err := dbm.NewDB("application", dbm.GoLevelDBBackend, filepath.Join(home, "data"))
	if err != nil {
		t.Fatal(err)
	}
	app := NewTrueRepublicApp(log.NewNopLogger(), database, home)
	defer func() { _ = app.Close() }()
	state := governedUpgradeV042State{Height: app.LastBlockHeight()}
	ctx := app.NewUncachedContext(false, cmtproto.Header{Height: state.Height})
	state.Marker = append([]byte(nil), ctx.KVStore(app.keys[truedemocracy.ModuleName]).Get(governedUpgradeMarkerV042)...)
	state.DoneHeight, err = app.upgradeKeeper.GetDoneHeight(ctx, governedUpgradePlanV042)
	if err != nil {
		t.Fatal(err)
	}
	_, err = app.upgradeKeeper.GetUpgradePlan(ctx)
	switch {
	case err == nil:
		state.PlanFound = true
	case errors.Is(err, upgradetypes.ErrNoUpgradePlanFound):
	default:
		t.Fatal(err)
	}
	versions, err := app.upgradeKeeper.GetModuleVersionMap(ctx)
	if err != nil {
		t.Fatal(err)
	}
	state.ModuleVersion = versions[truedemocracy.ModuleName]
	app.tdKeeper.IterateStoneRewardRecords(ctx, func(record truedemocracy.StoneRewardRecord) bool {
		state.RewardRecords = append(state.RewardRecords, record)
		return false
	})
	return state
}

// readGovernedUpgradeV042StateIsolated reads a stopped node's state in a
// short-lived child test process. Opening the app in-process would keep the
// CosmWasm VM lock on <home>/wasm for the rest of the test and block the next
// binary that starts on that home.
func readGovernedUpgradeV042StateIsolated(t *testing.T, ctx context.Context, home string) governedUpgradeV042State {
	t.Helper()
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestGovernedUpgradeV042StateProbe$", "-test.count=1")
	command.Env = append(os.Environ(), governedUpgradeV042ProbeHomeEnv+"="+home)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("state probe for %s: %v\n%s", home, err, output)
	}
	for _, line := range strings.Split(string(output), "\n") {
		if payload, found := strings.CutPrefix(strings.TrimSpace(line), governedUpgradeV042ProbePrefix); found {
			var state governedUpgradeV042State
			if err := json.Unmarshal([]byte(payload), &state); err != nil {
				t.Fatalf("decode state probe for %s: %v", home, err)
			}
			return state
		}
	}
	t.Fatalf("state probe for %s printed no state\n%s", home, output)
	return governedUpgradeV042State{}
}

// TestGovernedUpgradeV042StateProbe is the child side of
// readGovernedUpgradeV042StateIsolated; it is a no-op unless the probe home is set.
func TestGovernedUpgradeV042StateProbe(t *testing.T) {
	home := os.Getenv(governedUpgradeV042ProbeHomeEnv)
	if home == "" {
		t.Skip("state probe runs only as a child of the v0.4.2 upgrade harness")
	}
	encoded, err := json.Marshal(readGovernedUpgradeV042State(t, home))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(governedUpgradeV042ProbePrefix + string(encoded))
}
