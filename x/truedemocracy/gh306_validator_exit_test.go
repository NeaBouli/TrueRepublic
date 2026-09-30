package truedemocracy

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	rewards "truerepublic/treasury/keeper"
)

// TestFullValidatorExitBypassesBudgetAtZeroPayouts proves a bootstrap-domain
// validator can fully exit before any payout exists: the WP §7 transfer budget
// is skipped, TransferredStake is never touched, the stake stays in the
// slashable dual height/time evidence hold, and the mature hold pays the
// authenticated operator exactly.
func TestFullValidatorExitBypassesBudgetAtZeroPayouts(t *testing.T) {
	keeper, ctx, bank := setupKeeperWithBank(t)
	operator := sdk.AccAddress("bootstrap-exit-operator")
	stake := int64(rewards.StakeMin)
	ctx = withEvidenceWindow(ctx.WithBlockHeight(50), 5, 10*time.Minute)
	bank.fundAccount(operator, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, stake+1)))
	if err := keeper.CreateDomainWithEscrow(ctx, "Bootstrap", operator, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1))); err != nil {
		t.Fatal(err)
	}
	if err := keeper.RegisterValidatorWithEscrow(
		ctx,
		operator,
		operator.String(),
		testPubKey("bootstrap-exit-key"),
		sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, stake)),
		"Bootstrap",
	); err != nil {
		t.Fatal(err)
	}

	// TotalPayouts is zero: the legacy transfer check would fail closed here.
	if err := keeper.RemoveValidatorWithEscrow(ctx, operator, operator.String()); err != nil {
		t.Fatalf("zero-payout full exit blocked: %v", err)
	}
	removal, found := keeper.GetPendingValidatorRemoval(ctx, operator.String())
	if !found {
		t.Fatal("full exit did not create the evidence hold")
	}
	if !removal.BudgetExempt {
		t.Fatal("new exit hold is not marked budget-exempt")
	}
	if got := removal.Validator.Stake.AmountOf(PNYXDenom).Int64(); got != stake {
		t.Fatalf("held stake = %d, want %d", got, stake)
	}
	domain, _ := keeper.GetDomain(ctx, "Bootstrap")
	if domain.TransferredStake != 0 {
		t.Fatalf("TransferredStake = %d, want untouched 0", domain.TransferredStake)
	}
	if err := keeper.ValidateEscrowParity(ctx); err != nil {
		t.Fatalf("parity after budget-exempt exit: %v", err)
	}

	// The hold still matures through both evidence limits and pays the
	// authenticated operator.
	before := accountBalance(bank, operator)
	retirementCtx := ctx.
		WithBlockHeight(removal.ConsensusRetiredHeight).
		WithBlockTime(ctx.BlockTime().Add(time.Minute))
	if err := keeper.ProcessPendingValidatorRemovals(retirementCtx); err != nil {
		t.Fatal(err)
	}
	removal, _ = keeper.GetPendingValidatorRemoval(retirementCtx, operator.String())
	releaseCtx := retirementCtx.
		WithBlockHeight(removal.ReleaseAfterHeight + 1).
		WithBlockTime(time.Unix(0, removal.ReleaseAfterTimeNanos+1))
	if err := keeper.ProcessPendingValidatorRemovals(releaseCtx); err != nil {
		t.Fatal(err)
	}
	if got, want := accountBalance(bank, operator), before+stake; got != want {
		t.Fatalf("mature payout balance = %d, want %d", got, want)
	}
	if err := keeper.ValidateEscrowParity(releaseCtx); err != nil {
		t.Fatalf("parity after budget-exempt payout: %v", err)
	}
	domain, _ = keeper.GetDomain(releaseCtx, "Bootstrap")
	if domain.TransferredStake != 0 {
		t.Fatalf("TransferredStake after payout = %d, want untouched 0", domain.TransferredStake)
	}
}

// TestFullValidatorExitDoesNotConsumeTransferBudget proves an over-limit full
// exit succeeds without spending the domain's legacy 10% budget, leaving the
// complete budget available to the remaining validators' legacy accounting.
func TestFullValidatorExitDoesNotConsumeTransferBudget(t *testing.T) {
	keeper, ctx, bank := setupKeeperWithBank(t)
	admin := sdk.AccAddress("budget-admin")
	exiting := sdk.AccAddress("budget-exiting-operator")
	remaining := sdk.AccAddress("budget-remaining-operator")
	stake := int64(rewards.StakeMin)
	ctx = withEvidenceWindow(ctx.WithBlockHeight(50), 5, 10*time.Minute)

	bank.fundAccount(admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1)))
	if err := keeper.CreateDomainWithEscrow(ctx, "Budget", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1))); err != nil {
		t.Fatal(err)
	}
	for _, operator := range []sdk.AccAddress{exiting, remaining} {
		if err := keeper.AddMember(ctx, "Budget", operator.String(), admin); err != nil {
			t.Fatal(err)
		}
	}
	bank.fundAccount(exiting, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, stake)))
	bank.fundAccount(remaining, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 2*stake)))
	if err := keeper.RegisterValidatorWithEscrow(ctx, exiting, exiting.String(), testPubKey("budget-exit-key"), sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, stake)), "Budget"); err != nil {
		t.Fatal(err)
	}
	if err := keeper.RegisterValidatorWithEscrow(ctx, remaining, remaining.String(), testPubKey("budget-remain-key"), sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 2*stake)), "Budget"); err != nil {
		t.Fatal(err)
	}

	// 10% limit is 1,000 upnyx: far below the exiting stake.
	domain, _ := keeper.GetDomain(ctx, "Budget")
	domain.TotalPayouts = 10_000
	saveDomain(t, keeper, ctx, domain)

	if err := keeper.RemoveValidatorWithEscrow(ctx, exiting, exiting.String()); err != nil {
		t.Fatalf("over-limit full exit blocked: %v", err)
	}
	domain, _ = keeper.GetDomain(ctx, "Budget")
	if domain.TransferredStake != 0 {
		t.Fatalf("TransferredStake = %d, want 0 after budget-exempt exit", domain.TransferredStake)
	}

	// The legacy budget is untouched: the remaining validator's nominal
	// withdrawal accounting still enforces the exact 10% limit.
	if err := keeper.WithdrawStake(ctx, remaining.String(), 1_001); err == nil {
		t.Fatal("legacy withdrawal exceeded the untouched 10% budget")
	}
	if err := keeper.WithdrawStake(ctx, remaining.String(), 1_000); err != nil {
		t.Fatalf("legacy withdrawal within the untouched budget rejected: %v", err)
	}
	domain, _ = keeper.GetDomain(ctx, "Budget")
	if domain.TransferredStake != 1_000 {
		t.Fatalf("TransferredStake = %d, want 1000 from the legacy withdrawal only", domain.TransferredStake)
	}
}

// TestPendingRemovalSlashAccountingLegacyVsExempt proves the additive marker
// split: slashing a legacy hold decrements TransferredStake symmetrically,
// while slashing a budget-exempt hold burns the penalty without ever touching
// unrelated historical TransferredStake.
func TestPendingRemovalSlashAccountingLegacyVsExempt(t *testing.T) {
	keeper, ctx, bank := setupKeeperWithBank(t)
	admin := sdk.AccAddress("slash-mixed-admin")
	legacyOp := sdk.AccAddress("slash-legacy-operator")
	exemptOp := sdk.AccAddress("slash-exempt-operator")
	stake := int64(rewards.StakeMin)
	ctx = withEvidenceWindow(ctx.WithBlockHeight(50), 5, 10*time.Minute)

	bank.fundAccount(admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1)))
	if err := keeper.CreateDomainWithEscrow(ctx, "MixedSlash", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1))); err != nil {
		t.Fatal(err)
	}
	for _, operator := range []sdk.AccAddress{legacyOp, exemptOp} {
		if err := keeper.AddMember(ctx, "MixedSlash", operator.String(), admin); err != nil {
			t.Fatal(err)
		}
		bank.fundAccount(operator, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, stake)))
	}
	legacyKey := testPubKey("slash-legacy-key")
	exemptKey := testPubKey("slash-exempt-key")
	if err := keeper.RegisterValidatorWithEscrow(ctx, legacyOp, legacyOp.String(), legacyKey, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, stake)), "MixedSlash"); err != nil {
		t.Fatal(err)
	}
	if err := keeper.RegisterValidatorWithEscrow(ctx, exemptOp, exemptOp.String(), exemptKey, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, stake)), "MixedSlash"); err != nil {
		t.Fatal(err)
	}
	domain, _ := keeper.GetDomain(ctx, "MixedSlash")
	domain.TotalPayouts = stake * 100
	saveDomain(t, keeper, ctx, domain)

	// Reconstruct a pre-GH-306 legacy exit: TransferredStake was incremented
	// and the hold carries no marker.
	legacyVal, _ := keeper.GetValidator(ctx, legacyOp.String())
	if err := keeper.WithdrawStake(ctx, legacyOp.String(), stake); err != nil {
		t.Fatalf("legacy full withdrawal: %v", err)
	}
	legacyRemoval, err := newPendingValidatorRemoval(ctx, legacyVal, legacyOp.String())
	if err != nil {
		t.Fatal(err)
	}
	legacyRemoval.BudgetExempt = false
	keeper.SetPendingValidatorRemoval(ctx, legacyRemoval)

	// The new exit path is budget-exempt and leaves TransferredStake alone.
	if err := keeper.RemoveValidatorWithEscrow(ctx, exemptOp, exemptOp.String()); err != nil {
		t.Fatalf("exempt full exit: %v", err)
	}
	domain, _ = keeper.GetDomain(ctx, "MixedSlash")
	if domain.TransferredStake != stake {
		t.Fatalf("TransferredStake = %d, want legacy-only %d", domain.TransferredStake, stake)
	}

	if err := keeper.HandleDoubleSign(ctx, exemptKey); err != nil {
		t.Fatal(err)
	}
	domain, _ = keeper.GetDomain(ctx, "MixedSlash")
	if domain.TransferredStake != stake {
		t.Fatalf("exempt slash moved TransferredStake to %d, want %d", domain.TransferredStake, stake)
	}
	exemptRemoval, _ := keeper.GetPendingValidatorRemoval(ctx, exemptOp.String())
	wantExemptStake := int64(95_000 * PNYXUnit)
	if got := exemptRemoval.Validator.Stake.AmountOf(PNYXDenom).Int64(); got != wantExemptStake {
		t.Fatalf("exempt held stake after slash = %d, want %d", got, wantExemptStake)
	}

	if err := keeper.HandleDoubleSign(ctx, legacyKey); err != nil {
		t.Fatal(err)
	}
	legacyPenalty := stake - wantExemptStake
	domain, _ = keeper.GetDomain(ctx, "MixedSlash")
	if got, want := domain.TransferredStake, stake-legacyPenalty; got != want {
		t.Fatalf("legacy slash TransferredStake = %d, want symmetric %d", got, want)
	}
	legacyRemoval, _ = keeper.GetPendingValidatorRemoval(ctx, legacyOp.String())
	if got := legacyRemoval.Validator.Stake.AmountOf(PNYXDenom).Int64(); got != wantExemptStake {
		t.Fatalf("legacy held stake after slash = %d, want %d", got, wantExemptStake)
	}
	if got, want := bank.burned.AmountOf(PNYXDenom).Int64(), 2*legacyPenalty; got != want {
		t.Fatalf("burned = %d, want %d", got, want)
	}
	if err := keeper.ValidateEscrowParity(ctx); err != nil {
		t.Fatalf("parity after mixed slashes: %v", err)
	}
}
