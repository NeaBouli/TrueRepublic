package truedemocracy

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// setupElectionDomain creates a domain with 5 members and an issue "BoardChair"
// with 3 candidate suggestions for election tests (WP §3.7).
func setupElectionDomain(t *testing.T, k Keeper, ctx sdk.Context, mode VotingMode, abstentionAllowed bool) {
	t.Helper()
	k.CreateDomain(ctx, "ElecDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500_000)))

	domain, _ := k.GetDomain(ctx, "ElecDomain")
	domain.Members = []string{"alice", "bob", "charlie", "dave", "eve"}
	domain.Options.VotingMode = mode
	domain.Options.AbstentionAllowed = abstentionAllowed

	now := ctx.BlockTime().Unix()
	domain.Issues = []Issue{
		{
			Name: "BoardChair", Stones: 0, CreationDate: now, LastActivityAt: now,
			Suggestions: []Suggestion{
				{Name: "Alice", Creator: "bob", Stones: 0, Ratings: []Rating{}, CreationDate: now},
				{Name: "Bob", Creator: "alice", Stones: 0, Ratings: []Rating{}, CreationDate: now},
				{Name: "Charlie", Creator: "dave", Stones: 0, Ratings: []Rating{}, CreationDate: now},
			},
		},
	}

	st := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	st.Set([]byte("domain:ElecDomain"), bz)
}

// ---------- CastElectionVote ----------

func TestCastElectionVote(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupElectionDomain(t, k, ctx, VotingModeSimpleMajority, true)

	t.Run("approve vote", func(t *testing.T) {
		err := k.CastElectionVote(ctx, "ElecDomain", "BoardChair", "Alice", "bob", VoteChoiceApprove)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("abstain vote", func(t *testing.T) {
		err := k.CastElectionVote(ctx, "ElecDomain", "BoardChair", "", "charlie", VoteChoiceAbstain)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("non-member rejected", func(t *testing.T) {
		err := k.CastElectionVote(ctx, "ElecDomain", "BoardChair", "Alice", "outsider", VoteChoiceApprove)
		if err == nil {
			t.Fatal("expected error for non-member")
		}
	})

	t.Run("unknown candidate rejected", func(t *testing.T) {
		err := k.CastElectionVote(ctx, "ElecDomain", "BoardChair", "Unknown", "alice", VoteChoiceApprove)
		if err == nil {
			t.Fatal("expected error for unknown candidate")
		}
	})

	t.Run("abstain blocked when not allowed", func(t *testing.T) {
		k2, ctx2 := setupKeeper(t)
		setupElectionDomain(t, k2, ctx2, VotingModeSimpleMajority, false)

		err := k2.CastElectionVote(ctx2, "ElecDomain", "BoardChair", "", "bob", VoteChoiceAbstain)
		if err == nil {
			t.Fatal("expected error when abstention is not allowed")
		}
	})

	t.Run("vote can be changed", func(t *testing.T) {
		// bob changes vote from Alice to Bob
		err := k.CastElectionVote(ctx, "ElecDomain", "BoardChair", "Bob", "bob", VoteChoiceApprove)
		if err != nil {
			t.Fatalf("unexpected error changing vote: %v", err)
		}
	})
}

// ---------- TallyElection: Simple Majority ----------

func TestTallySimpleMajority(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupElectionDomain(t, k, ctx, VotingModeSimpleMajority, true)

	t.Run("winner with majority", func(t *testing.T) {
		// 3 vote for Alice, 1 abstains, 1 does not vote → Alice 3/3 = 100%
		k.CastElectionVote(ctx, "ElecDomain", "BoardChair", "Alice", "bob", VoteChoiceApprove)
		k.CastElectionVote(ctx, "ElecDomain", "BoardChair", "Alice", "charlie", VoteChoiceApprove)
		k.CastElectionVote(ctx, "ElecDomain", "BoardChair", "Alice", "dave", VoteChoiceApprove)
		k.CastElectionVote(ctx, "ElecDomain", "BoardChair", "", "eve", VoteChoiceAbstain)

		result, err := k.TallyElection(ctx, "ElecDomain", "BoardChair")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Elected {
			t.Error("expected Alice to be elected")
		}
		if result.Candidate != "Alice" {
			t.Errorf("winner = %q, want Alice", result.Candidate)
		}
		if result.Votes != 3 {
			t.Errorf("votes = %d, want 3", result.Votes)
		}
		if result.Abstained != 1 {
			t.Errorf("abstained = %d, want 1", result.Abstained)
		}
		if result.Total != 3 { // 4 voted - 1 abstained = 3 effective
			t.Errorf("total = %d, want 3", result.Total)
		}
	})

	t.Run("no majority when split", func(t *testing.T) {
		k2, ctx2 := setupKeeper(t)
		setupElectionDomain(t, k2, ctx2, VotingModeSimpleMajority, true)

		// 2 for Alice, 2 for Bob, 1 abstain → Alice 2/4 = 50%, not >50%
		k2.CastElectionVote(ctx2, "ElecDomain", "BoardChair", "Alice", "charlie", VoteChoiceApprove)
		k2.CastElectionVote(ctx2, "ElecDomain", "BoardChair", "Alice", "dave", VoteChoiceApprove)
		k2.CastElectionVote(ctx2, "ElecDomain", "BoardChair", "Bob", "alice", VoteChoiceApprove)
		k2.CastElectionVote(ctx2, "ElecDomain", "BoardChair", "Bob", "eve", VoteChoiceApprove)
		k2.CastElectionVote(ctx2, "ElecDomain", "BoardChair", "", "bob", VoteChoiceAbstain)

		result, err := k2.TallyElection(ctx2, "ElecDomain", "BoardChair")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Elected {
			t.Error("expected no winner with 50/50 split")
		}
	})
}

// ---------- TallyElection: Absolute Majority ----------

func TestTallyAbsoluteMajority(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupElectionDomain(t, k, ctx, VotingModeAbsoluteMajority, true)

	t.Run("elected with absolute majority", func(t *testing.T) {
		// 3 out of 5 members vote for Alice → 3/5 = 60% > 50%
		k.CastElectionVote(ctx, "ElecDomain", "BoardChair", "Alice", "bob", VoteChoiceApprove)
		k.CastElectionVote(ctx, "ElecDomain", "BoardChair", "Alice", "charlie", VoteChoiceApprove)
		k.CastElectionVote(ctx, "ElecDomain", "BoardChair", "Alice", "dave", VoteChoiceApprove)

		result, err := k.TallyElection(ctx, "ElecDomain", "BoardChair")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Elected {
			t.Error("expected Alice elected with 3/5 absolute majority")
		}
		if result.Total != 5 { // denominator is total members
			t.Errorf("total = %d, want 5 (all members)", result.Total)
		}
	})

	t.Run("not elected without absolute majority", func(t *testing.T) {
		k2, ctx2 := setupKeeper(t)
		setupElectionDomain(t, k2, ctx2, VotingModeAbsoluteMajority, true)

		// 2 out of 5 vote for Alice, 1 abstains → 2/5 = 40% < 50%
		k2.CastElectionVote(ctx2, "ElecDomain", "BoardChair", "Alice", "bob", VoteChoiceApprove)
		k2.CastElectionVote(ctx2, "ElecDomain", "BoardChair", "Alice", "charlie", VoteChoiceApprove)
		k2.CastElectionVote(ctx2, "ElecDomain", "BoardChair", "", "dave", VoteChoiceAbstain)

		result, err := k2.TallyElection(ctx2, "ElecDomain", "BoardChair")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Elected {
			t.Error("expected Alice NOT elected with 2/5 absolute majority")
		}
	})
}

// ---------- TallyElection: Edge Cases ----------

func TestTallyEdgeCases(t *testing.T) {
	t.Run("no votes cast", func(t *testing.T) {
		k, ctx := setupKeeper(t)
		setupElectionDomain(t, k, ctx, VotingModeSimpleMajority, true)

		result, err := k.TallyElection(ctx, "ElecDomain", "BoardChair")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Elected {
			t.Error("expected no winner with zero votes")
		}
		if result.Total != 0 {
			t.Errorf("total = %d, want 0", result.Total)
		}
	})

	t.Run("all abstain", func(t *testing.T) {
		k, ctx := setupKeeper(t)
		setupElectionDomain(t, k, ctx, VotingModeSimpleMajority, true)

		for _, m := range []string{"alice", "bob", "charlie", "dave", "eve"} {
			k.CastElectionVote(ctx, "ElecDomain", "BoardChair", "", m, VoteChoiceAbstain)
		}

		result, err := k.TallyElection(ctx, "ElecDomain", "BoardChair")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Elected {
			t.Error("expected no winner when all abstain")
		}
		if result.Abstained != 5 {
			t.Errorf("abstained = %d, want 5", result.Abstained)
		}
	})

	t.Run("unknown domain", func(t *testing.T) {
		k, ctx := setupKeeper(t)
		_, err := k.TallyElection(ctx, "NonExistent", "BoardChair")
		if err == nil {
			t.Fatal("expected error for unknown domain")
		}
	})

	t.Run("unknown issue", func(t *testing.T) {
		k, ctx := setupKeeper(t)
		setupElectionDomain(t, k, ctx, VotingModeSimpleMajority, true)

		_, err := k.TallyElection(ctx, "ElecDomain", "NonExistent")
		if err == nil {
			t.Fatal("expected error for unknown issue")
		}
	})
}
