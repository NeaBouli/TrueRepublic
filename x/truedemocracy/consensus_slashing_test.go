package truedemocracy

import (
	"encoding/json"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func withConsensusSignals(
	ctx sdk.Context,
	blockHeight int64,
	evidence []abci.Misbehavior,
	votes []abci.VoteInfo,
) sdk.Context {
	return ctx.
		WithBlockHeight(blockHeight).
		WithChainID("truerepublic-consensus-test-1").
		WithCometInfo(baseapp.NewBlockInfo(
			evidence,
			nil,
			nil,
			abci.CommitInfo{Round: 0, Votes: votes},
		))
}

func validatorVote(pubKey []byte, flag cmtproto.BlockIDFlag) abci.VoteInfo {
	return abci.VoteInfo{
		Validator: abci.Validator{
			Address: consensusAddressFromPubKey(pubKey),
			Power:   1,
		},
		BlockIdFlag: flag,
	}
}

func TestProcessConsensusSignalsSlashesDelayedRotatedKeyExactlyOnce(t *testing.T) {
	k, ctx := setupKeeper(t)
	operator, before := setupRotationValidator(t, k, ctx)
	newKey := testPubKey("delayed-evidence-new")
	rotationCtx := ctx.WithBlockHeight(10)
	if _, err := k.RotateValidatorKey(
		rotationCtx,
		operator,
		operator.String(),
		before.PubKey,
		newKey,
	); err != nil {
		t.Fatal(err)
	}
	bank := backExistingEscrow(&k, ctx)

	blockCtx := withConsensusSignals(
		ctx.WithBlockTime(ctx.BlockTime().Add(time.Hour)),
		20,
		[]abci.Misbehavior{{
			Type:             abci.MisbehaviorType_DUPLICATE_VOTE,
			Validator:        abci.Validator{Address: consensusAddressFromPubKey(before.PubKey), Power: 1},
			Height:           11,
			Time:             ctx.BlockTime(),
			TotalVotingPower: 1,
		}},
		[]abci.VoteInfo{validatorVote(newKey, cmtproto.BlockIDFlagCommit)},
	)
	if err := k.ProcessConsensusSignals(blockCtx); err != nil {
		t.Fatal(err)
	}

	validator, found := k.GetValidator(blockCtx, operator.String())
	if !found || !validator.Jailed {
		t.Fatal("delayed equivocation did not jail the rotated validator")
	}
	wantBurned := int64(5_000 * PNYXUnit)
	if got := bank.burned.AmountOf(PNYXDenom).Int64(); got != wantBurned {
		t.Fatalf("burned = %d, want %d", got, wantBurned)
	}
	oldRecord, found := k.GetConsensusKeyRecord(blockCtx, consensusAddressFromPubKey(before.PubKey))
	if !found || !oldRecord.Tombstoned {
		t.Fatal("equivocating historical key was not tombstoned")
	}

	if err := k.ProcessConsensusSignals(blockCtx); err != nil {
		t.Fatalf("identical replay failed: %v", err)
	}
	if got := bank.burned.AmountOf(PNYXDenom).Int64(); got != wantBurned {
		t.Fatalf("replay burned %d, want unchanged %d", got, wantBurned)
	}
	var processed int
	k.IterateProcessedInfractions(blockCtx, func(ProcessedInfraction) bool {
		processed++
		return false
	})
	if processed != 1 {
		t.Fatalf("processed infractions = %d, want 1", processed)
	}

	changedReplay := withConsensusSignals(
		blockCtx,
		20,
		[]abci.Misbehavior{{
			Type:             abci.MisbehaviorType_DUPLICATE_VOTE,
			Validator:        abci.Validator{Address: consensusAddressFromPubKey(before.PubKey), Power: 1},
			Height:           10,
			Time:             ctx.BlockTime(),
			TotalVotingPower: 1,
		}},
		[]abci.VoteInfo{validatorVote(newKey, cmtproto.BlockIDFlagCommit)},
	)
	if err := k.ProcessConsensusSignals(changedReplay); err == nil {
		t.Fatal("same-height replay changed canonical evidence without rejection")
	}
	if got := bank.burned.AmountOf(PNYXDenom).Int64(); got != wantBurned {
		t.Fatal("rejected changed replay altered the burn")
	}

	nextBlockCtx := withConsensusSignals(
		blockCtx.WithBlockTime(blockCtx.BlockTime().Add(time.Second)),
		21,
		[]abci.Misbehavior{{
			Type:             abci.MisbehaviorType_DUPLICATE_VOTE,
			Validator:        abci.Validator{Address: consensusAddressFromPubKey(before.PubKey), Power: 1},
			Height:           10,
			Time:             ctx.BlockTime(),
			TotalVotingPower: 1,
		}},
		[]abci.VoteInfo{validatorVote(newKey, cmtproto.BlockIDFlagCommit)},
	)
	if err := k.ProcessConsensusSignals(nextBlockCtx); err != nil {
		t.Fatalf("second tombstoned-key offense failed: %v", err)
	}
	if got := bank.burned.AmountOf(PNYXDenom).Int64(); got != wantBurned {
		t.Fatalf("tombstoned key burned %d, want capped %d", got, wantBurned)
	}
	processed = 0
	k.IterateProcessedInfractions(nextBlockCtx, func(ProcessedInfraction) bool {
		processed++
		return false
	})
	if processed != 2 {
		t.Fatalf("processed infractions = %d, want two audit records", processed)
	}
}

func TestProcessConsensusSignalsUsesCompleteRollingLivenessWindow(t *testing.T) {
	k, ctx := setupKeeper(t)
	operator, pubKey := setupDomainWithValidator(t, k, ctx)
	bank := backExistingEscrow(&k, ctx)

	for commitHeight := int64(1); commitHeight <= SignedBlocksWindow; commitHeight++ {
		flag := cmtproto.BlockIDFlagCommit
		if commitHeight <= SignedBlocksWindow-MinSignedPerWindow+1 {
			flag = cmtproto.BlockIDFlagAbsent
		}
		blockCtx := withConsensusSignals(
			ctx.WithBlockTime(ctx.BlockTime().Add(time.Duration(commitHeight)*time.Second)),
			commitHeight+1,
			nil,
			[]abci.VoteInfo{validatorVote(pubKey, flag)},
		)
		if err := k.ProcessConsensusSignals(blockCtx); err != nil {
			t.Fatalf("commit %d: %v", commitHeight, err)
		}
		if commitHeight == SignedBlocksWindow-MinSignedPerWindow+1 {
			validator, _ := k.GetValidator(blockCtx, operator)
			if validator.Jailed {
				t.Fatal("validator was punished before a complete liveness window")
			}
		}
	}

	validator, found := k.GetValidator(ctx, operator)
	if !found || !validator.Jailed {
		t.Fatal("validator was not jailed after 51 misses in a complete window")
	}
	if validator.MissedBlocks != 0 {
		t.Fatalf("missed blocks = %d, want reset to 0", validator.MissedBlocks)
	}
	wantBurned := int64(1_000 * PNYXUnit)
	if got := bank.burned.AmountOf(PNYXDenom).Int64(); got != wantBurned {
		t.Fatalf("burned = %d, want %d", got, wantBurned)
	}
}

func TestConsensusSlashingGenesisRoundTripPreservesReplayAndJailState(t *testing.T) {
	am1, k1, ctx1 := setupModuleForGenesis(t)
	operator, before := setupRotationValidator(t, k1, ctx1)
	newKey := testPubKey("roundtrip-evidence-new")
	if _, err := k1.RotateValidatorKey(
		ctx1.WithBlockHeight(10),
		operator,
		operator.String(),
		before.PubKey,
		newKey,
	); err != nil {
		t.Fatal(err)
	}
	backExistingEscrow(&k1, ctx1)
	blockCtx1 := withConsensusSignals(
		ctx1.WithBlockTime(ctx1.BlockTime().Add(time.Hour)),
		20,
		[]abci.Misbehavior{{
			Type:             abci.MisbehaviorType_DUPLICATE_VOTE,
			Validator:        abci.Validator{Address: consensusAddressFromPubKey(before.PubKey), Power: 1},
			Height:           11,
			Time:             ctx1.BlockTime(),
			TotalVotingPower: 1,
		}},
		[]abci.VoteInfo{validatorVote(newKey, cmtproto.BlockIDFlagCommit)},
	)
	if err := k1.ProcessConsensusSignals(blockCtx1); err != nil {
		t.Fatal(err)
	}

	exported := am1.ExportGenesis(blockCtx1, nil)
	var genesis GenesisState
	if err := json.Unmarshal(exported, &genesis); err != nil {
		t.Fatal(err)
	}
	if err := ValidateGenesisState(genesis); err != nil {
		t.Fatalf("exported slashing state is invalid: %v", err)
	}

	am2, k2, ctx2 := setupModuleForGenesis(t)
	if updates := am2.InitGenesis(ctx2, nil, exported); len(updates) != 0 {
		t.Fatalf("jailed validator produced %d genesis updates, want 0", len(updates))
	}
	validator, found := k2.GetValidator(ctx2, operator.String())
	if !found || !validator.Jailed || validator.Stake.AmountOf(PNYXDenom).Int64() != 95_000*PNYXUnit {
		t.Fatal("validator jail and slashed stake did not survive export/import")
	}
	record, found := k2.GetConsensusKeyRecord(ctx2, consensusAddressFromPubKey(before.PubKey))
	if !found || !record.Tombstoned {
		t.Fatal("historical key tombstone did not survive export/import")
	}
	cursor, found := k2.getLastCommitCursor(ctx2)
	if !found || cursor.CommitHeight != 19 {
		t.Fatal("last-commit cursor did not survive export/import")
	}
	var processed int
	k2.IterateProcessedInfractions(ctx2, func(ProcessedInfraction) bool {
		processed++
		return false
	})
	if processed != 1 {
		t.Fatalf("processed infractions after import = %d, want 1", processed)
	}

	replayCtx := withConsensusSignals(
		ctx2.WithBlockTime(blockCtx1.BlockTime()),
		20,
		[]abci.Misbehavior{{
			Type:             abci.MisbehaviorType_DUPLICATE_VOTE,
			Validator:        abci.Validator{Address: consensusAddressFromPubKey(before.PubKey), Power: 1},
			Height:           11,
			Time:             ctx1.BlockTime(),
			TotalVotingPower: 1,
		}},
		[]abci.VoteInfo{validatorVote(newKey, cmtproto.BlockIDFlagCommit)},
	)
	if err := k2.ProcessConsensusSignals(replayCtx); err != nil {
		t.Fatalf("post-import replay was not idempotent: %v", err)
	}
}
