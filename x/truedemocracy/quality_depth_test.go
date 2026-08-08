package truedemocracy

import (
	"reflect"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// replayPropertySnapshot captures the slashing state an evidence replay must
// never drift: canonical burn, processed-infraction audit trail, validator
// record, and the historical key tombstone.
type replayPropertySnapshot struct {
	burned     int64
	processed  int
	validator  Validator
	tombstoned bool
}

// TestConsensusSlashingReplayProperty asserts the replay-safety property of
// consensus evidence processing: replaying identical evidence at the same
// block height is idempotent, while a changed replay for the same identity is
// rejected without any state drift.
func TestConsensusSlashingReplayProperty(t *testing.T) {
	cases := []struct {
		name           string
		blockHeight    int64
		evidenceHeight int64
		changedHeight  int64
	}{
		{name: "baseline delayed evidence", blockHeight: 20, evidenceHeight: 11, changedHeight: 10},
		{name: "later block", blockHeight: 30, evidenceHeight: 11, changedHeight: 12},
		{name: "older changed evidence height", blockHeight: 25, evidenceHeight: 11, changedHeight: 9},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			k, ctx := setupKeeper(t)
			operator, before := setupRotationValidator(t, k, ctx)
			newKey := testPubKey("replay-property-new")
			if _, err := k.RotateValidatorKey(
				ctx.WithBlockHeight(10),
				operator,
				operator.String(),
				before.PubKey,
				newKey,
			); err != nil {
				t.Fatal(err)
			}
			bank := backExistingEscrow(&k, ctx)

			signals := func(evidenceHeight int64, flag cmtproto.BlockIDFlag) sdk.Context {
				return withConsensusSignals(
					ctx.WithBlockTime(ctx.BlockTime().Add(time.Hour)),
					test.blockHeight,
					[]abci.Misbehavior{{
						Type:             abci.MisbehaviorType_DUPLICATE_VOTE,
						Validator:        abci.Validator{Address: consensusAddressFromPubKey(before.PubKey), Power: 1},
						Height:           evidenceHeight,
						Time:             ctx.BlockTime(),
						TotalVotingPower: 1,
					}},
					[]abci.VoteInfo{validatorVote(newKey, flag)},
				)
			}
			snapshot := func(blockCtx sdk.Context) replayPropertySnapshot {
				var processed int
				k.IterateProcessedInfractions(blockCtx, func(ProcessedInfraction) bool {
					processed++
					return false
				})
				validator, found := k.GetValidator(blockCtx, operator.String())
				if !found {
					t.Fatal("validator missing after processing")
				}
				record, found := k.GetConsensusKeyRecord(blockCtx, consensusAddressFromPubKey(before.PubKey))
				if !found {
					t.Fatal("historical key record missing after processing")
				}
				return replayPropertySnapshot{
					burned:     bank.burned.AmountOf(PNYXDenom).Int64(),
					processed:  processed,
					validator:  validator,
					tombstoned: record.Tombstoned,
				}
			}

			blockCtx := signals(test.evidenceHeight, cmtproto.BlockIDFlagCommit)
			if err := k.ProcessConsensusSignals(blockCtx); err != nil {
				t.Fatal(err)
			}
			want := snapshot(blockCtx)
			if want.burned <= 0 || want.processed != 1 || !want.validator.Jailed || !want.tombstoned {
				t.Fatalf("initial evidence processing incomplete: %+v", want)
			}

			// Identical evidence replay must be idempotent, repeatedly.
			for replay := 0; replay < 3; replay++ {
				if err := k.ProcessConsensusSignals(blockCtx); err != nil {
					t.Fatalf("identical replay %d failed: %v", replay, err)
				}
				if got := snapshot(blockCtx); !reflect.DeepEqual(got, want) {
					t.Fatalf("identical replay %d drifted state:\n got %+v\nwant %+v", replay, got, want)
				}
			}

			// A changed replay for the same identity and block height must be
			// rejected and must leave every tracked state component untouched.
			changedReplays := map[string]sdk.Context{
				"changed evidence height": signals(test.changedHeight, cmtproto.BlockIDFlagCommit),
				"changed commit votes":    signals(test.evidenceHeight, cmtproto.BlockIDFlagAbsent),
			}
			for name, changed := range changedReplays {
				if err := k.ProcessConsensusSignals(changed); err == nil {
					t.Fatalf("%s replay for the same identity was accepted", name)
				}
				if got := snapshot(blockCtx); !reflect.DeepEqual(got, want) {
					t.Fatalf("rejected %s replay drifted state:\n got %+v\nwant %+v", name, got, want)
				}
			}
		})
	}
}
