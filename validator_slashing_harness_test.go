package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
	"time"

	cmtbytes "github.com/cometbft/cometbft/libs/bytes"
	"github.com/cometbft/cometbft/privval"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	cmttypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"truerepublic/token"
	"truerepublic/x/truedemocracy"
)

// TestMultiValidatorConsensusSlashing proves that real CometBFT
// DecidedLastCommit and duplicate-vote evidence reach the application, change
// economic state exactly once, remove validator power, and preserve network
// convergence plus an exportable, bank-backed ledger.
func TestMultiValidatorConsensusSlashing(t *testing.T) {
	if os.Getenv(multiValidatorSmokeEnv) != "1" {
		t.Skipf("set %s=1 to run the validator-slashing process harness", multiValidatorSmokeEnv)
	}
	ctx := t.Context()

	binary := filepath.Join(t.TempDir(), "truerepublicd")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build daemon: %v\n%s", err, output)
	}

	const chainID = "truerepublic-validator-slashing-1"
	validators := make([]*smokeValidator, 4)
	for i := range validators {
		validator := &smokeValidator{
			name:    fmt.Sprintf("slashing-validator-%d", i+1),
			home:    filepath.Join(t.TempDir(), fmt.Sprintf("slashing-node-%d", i+1)),
			rpcPort: freeTCPPort(t),
			p2pPort: freeTCPPort(t),
			logPath: filepath.Join(t.TempDir(), fmt.Sprintf("slashing-validator-%d.log", i+1)),
		}
		validator.operatorAddr = smokeOperatorAddress(validator.name)
		initSmokeValidator(t, ctx, binary, chainID, validator)
		configureFastSlashingConsensus(t, filepath.Join(validator.home, "config", "config.toml"))
		validators[i] = validator
	}

	sharedGenesis := buildSharedSmokeGenesis(t, chainID, validators)
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
		if err := validator.start(ctx, binary, persistentPeers(validator, validators)); err != nil {
			t.Fatalf("start %s: %v", validator.name, err)
		}
	}
	waitForSmokeHeight(t, validators, 3, 90*time.Second)
	assertCommonAppHash(t, validators, 3)

	downtimeTarget := validators[3]
	if err := downtimeTarget.stop(true); err != nil {
		t.Fatalf("stop downtime target: %v", err)
	}
	startHeight := smokeHeight(t, validators[0])
	livenessHeight := startHeight + truedemocracy.SignedBlocksWindow + 3
	// This phase deliberately advances the full liveness window. Leave enough
	// headroom for loaded CI hosts where consensus rounds exceed the tuned
	// sub-second happy path.
	waitForSmokeHeight(t, validators[:3], livenessHeight, 300*time.Second)

	downtimeValidator := querySmokeApplicationValidator(t, ctx, binary, validators[0], downtimeTarget.operatorAddr)
	if !downtimeValidator.Jailed || downtimeValidator.Power != 0 {
		t.Fatalf("downtime validator jail/power = %t/%d, want true/0", downtimeValidator.Jailed, downtimeValidator.Power)
	}
	wantDowntimeStake := int64(99_000 * token.WholeTokenBaseUnits)
	if got := downtimeValidator.Stake.AmountOf(truedemocracy.PNYXDenom).Int64(); got != wantDowntimeStake {
		t.Fatalf("downtime stake = %d, want %d", got, wantDowntimeStake)
	}
	if _, found, err := querySmokeValidatorPowerAtHeight(ctx, validators[0], downtimeTarget.pubKey, 0); err != nil {
		t.Fatalf("query downtime consensus power: %v", err)
	} else if found {
		t.Fatal("downtime validator remained in the current consensus set")
	}

	if err := downtimeTarget.start(ctx, binary, persistentPeers(downtimeTarget, validators)); err != nil {
		t.Fatalf("restart downtime target as full node: %v", err)
	}
	waitForSmokeHeight(t, validators, livenessHeight+2, 120*time.Second)
	assertCommonAppHash(t, validators, livenessHeight+2)

	equivocationTarget := validators[2]
	evidenceHeight := smokeHeight(t, validators[0]) - 1
	broadcastDuplicateVoteEvidence(t, ctx, chainID, validators[0], equivocationTarget, evidenceHeight)
	evidenceObservedHeight := evidenceHeight + 4
	waitForSmokeHeight(t, validators, evidenceObservedHeight, 120*time.Second)

	equivocationValidator := querySmokeApplicationValidator(t, ctx, binary, validators[0], equivocationTarget.operatorAddr)
	if !equivocationValidator.Jailed || equivocationValidator.Power != 0 {
		t.Fatalf("equivocation validator jail/power = %t/%d, want true/0", equivocationValidator.Jailed, equivocationValidator.Power)
	}
	wantEquivocationStake := int64(95_000 * token.WholeTokenBaseUnits)
	if got := equivocationValidator.Stake.AmountOf(truedemocracy.PNYXDenom).Int64(); got != wantEquivocationStake {
		t.Fatalf("equivocation stake = %d, want %d", got, wantEquivocationStake)
	}
	if _, found, err := querySmokeValidatorPowerAtHeight(ctx, validators[0], equivocationTarget.pubKey, 0); err != nil {
		t.Fatalf("query equivocation consensus power: %v", err)
	} else if found {
		t.Fatal("equivocating validator remained in the current consensus set")
	}

	convergenceHeight := smokeHeight(t, validators[0]) + 2
	waitForSmokeHeight(t, validators, convergenceHeight, 120*time.Second)
	assertCommonAppHash(t, validators, convergenceHeight)
	downtimeValidator = querySmokeApplicationValidator(t, ctx, binary, validators[0], downtimeTarget.operatorAddr)
	equivocationValidator = querySmokeApplicationValidator(t, ctx, binary, validators[0], equivocationTarget.operatorAddr)
	for _, validator := range validators {
		if err := validator.stop(true); err != nil {
			t.Fatalf("stop %s before export: %v", validator.name, err)
		}
	}

	exported := exportSmokeGenesis(t, ctx, binary, validators[0], convergenceHeight)
	exportApp := newGenesisTestApp(t)
	if err := validateLedgerGenesis(exportApp.appCodec, exported.AppState); err != nil {
		t.Fatalf("slashing export is not exactly bank-backed: %v", err)
	}
	var democracyGenesis truedemocracy.GenesisState
	if err := json.Unmarshal(exported.AppState[truedemocracy.ModuleName], &democracyGenesis); err != nil {
		t.Fatalf("decode slashing export: %v", err)
	}
	if len(democracyGenesis.ProcessedInfractions) != 1 {
		t.Fatalf("processed infractions = %d, want 1", len(democracyGenesis.ProcessedInfractions))
	}
	if len(democracyGenesis.LastCommitCursor.Hash) != sha256.Size {
		t.Fatal("exported slashing state lacks the replay cursor")
	}
	assertExportedInactiveValidator(t, democracyGenesis, downtimeValidator)
	assertExportedInactiveValidator(t, democracyGenesis, equivocationValidator)

	recoveredApp := newGenesisTestApp(t)
	if err := initGenesisApp(recoveredApp, exported.AppState); err != nil {
		t.Fatalf("re-import slashing export: %v", err)
	}
	recoveredCtx := recoveredApp.NewContext(false)
	assertRecoveredInactiveValidator(t, recoveredApp.tdKeeper, recoveredCtx, downtimeValidator)
	assertRecoveredInactiveValidator(t, recoveredApp.tdKeeper, recoveredCtx, equivocationValidator)
}

func assertExportedInactiveValidator(
	t *testing.T,
	genesis truedemocracy.GenesisState,
	want truedemocracy.Validator,
) {
	t.Helper()
	for _, validator := range genesis.Validators {
		if validator.OperatorAddr != want.OperatorAddr {
			continue
		}
		if validator.Active == nil || *validator.Active {
			t.Fatalf("exported validator %q active flag = %v, want explicit inactive", want.OperatorAddr, validator.Active)
		}
		if validator.Stake != want.Stake.AmountOf(truedemocracy.PNYXDenom).Int64() ||
			validator.Power != want.Power ||
			validator.Jailed != want.Jailed ||
			validator.JailedUntil != want.JailedUntil ||
			validator.MissedBlocks != want.MissedBlocks ||
			!slices.Equal(validator.Domains, want.Domains) {
			t.Fatalf("exported inactive validator %q drifted: got %+v want %+v", want.OperatorAddr, validator, want)
		}
		return
	}
	t.Fatalf("export omitted inactive validator %q", want.OperatorAddr)
}

func assertRecoveredInactiveValidator(
	t *testing.T,
	keeper truedemocracy.Keeper,
	ctx sdk.Context,
	want truedemocracy.Validator,
) {
	t.Helper()
	validator, found := keeper.GetValidator(ctx, want.OperatorAddr)
	if !found {
		t.Fatalf("re-import omitted inactive validator %q", want.OperatorAddr)
	}
	if !validator.Stake.Equal(want.Stake) ||
		!bytes.Equal(validator.PubKey, want.PubKey) ||
		validator.Power != want.Power ||
		validator.Jailed != want.Jailed ||
		validator.JailedUntil != want.JailedUntil ||
		validator.MissedBlocks != want.MissedBlocks ||
		!slices.Equal(validator.Domains, want.Domains) {
		t.Fatalf("re-imported inactive validator %q drifted: got %+v want %+v", want.OperatorAddr, validator, want)
	}
}

func broadcastDuplicateVoteEvidence(
	t *testing.T,
	ctx context.Context,
	chainID string,
	rpcNode, target *smokeValidator,
	height int64,
) {
	t.Helper()
	client, err := rpchttp.New(fmt.Sprintf("http://127.0.0.1:%d", rpcNode.rpcPort), "/websocket")
	if err != nil {
		t.Fatal(err)
	}
	block, err := client.Block(ctx, &height)
	if err != nil {
		t.Fatalf("query evidence block %d: %v", height, err)
	}
	perPage := 100
	validatorResult, err := client.Validators(ctx, &height, nil, &perPage)
	if err != nil {
		t.Fatalf("query evidence validator set %d: %v", height, err)
	}
	validatorSet := cmttypes.NewValidatorSet(validatorResult.Validators)
	index, validator := validatorSet.GetByAddress(target.pubKeyAddress())
	if validator == nil {
		t.Fatalf("target %s was not in validator set at height %d", target.name, height)
	}

	filePV := privval.LoadFilePV(
		filepath.Join(target.home, "config", "priv_validator_key.json"),
		filepath.Join(target.home, "data", "priv_validator_state.json"),
	)
	mockPV := cmttypes.NewMockPVWithParams(filePV.Key.PrivKey, false, false)
	firstID := block.BlockID
	secondID := firstID
	conflictHash := sha256.Sum256(append([]byte("truerepublic/process-equivocation/v1"), firstID.Hash...))
	secondID.Hash = cmtbytes.HexBytes(conflictHash[:])
	if bytes.Equal(firstID.Hash, secondID.Hash) {
		t.Fatal("duplicate-vote block IDs did not conflict")
	}

	voteA, err := cmttypes.MakeVote(
		mockPV,
		chainID,
		index,
		height,
		0,
		cmtproto.PrecommitType,
		firstID,
		block.Block.Time,
	)
	if err != nil {
		t.Fatalf("sign first duplicate vote: %v", err)
	}
	voteB, err := cmttypes.MakeVote(
		mockPV,
		chainID,
		index,
		height,
		0,
		cmtproto.PrecommitType,
		secondID,
		block.Block.Time,
	)
	if err != nil {
		t.Fatalf("sign second duplicate vote: %v", err)
	}
	evidence, err := cmttypes.NewDuplicateVoteEvidence(voteA, voteB, block.Block.Time, validatorSet)
	if err != nil {
		t.Fatalf("construct duplicate-vote evidence: %v", err)
	}
	if _, err := client.BroadcastEvidence(ctx, evidence); err != nil {
		t.Fatalf("broadcast duplicate-vote evidence: %v", err)
	}
}

func (validator *smokeValidator) pubKeyAddress() []byte {
	return cmttypes.NewValidator(
		privval.LoadFilePV(
			filepath.Join(validator.home, "config", "priv_validator_key.json"),
			filepath.Join(validator.home, "data", "priv_validator_state.json"),
		).Key.PrivKey.PubKey(),
		1,
	).Address
}

func configureFastSlashingConsensus(t *testing.T, configPath string) {
	t.Helper()
	config, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	updated := string(config)
	for original, replacement := range map[string]string{
		`timeout_propose = "3s"`:            `timeout_propose = "500ms"`,
		`timeout_propose_delta = "500ms"`:   `timeout_propose_delta = "100ms"`,
		`timeout_prevote = "1s"`:            `timeout_prevote = "200ms"`,
		`timeout_prevote_delta = "500ms"`:   `timeout_prevote_delta = "100ms"`,
		`timeout_precommit = "1s"`:          `timeout_precommit = "200ms"`,
		`timeout_precommit_delta = "500ms"`: `timeout_precommit_delta = "100ms"`,
		`timeout_commit = "5s"`:             `timeout_commit = "200ms"`,
	} {
		updated = replaceTomlSectionValue(t, updated, "[consensus]", original, replacement)
	}
	if err := atomicWriteFile(configPath, []byte(updated), 0o600); err != nil {
		t.Fatal(err)
	}
}
