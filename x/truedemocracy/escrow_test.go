package truedemocracy

import (
	"encoding/json"
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"truerepublic/token"
	rewards "truerepublic/treasury/keeper"
)

func accountBalance(bank *mockBankKeeper, address sdk.AccAddress) int64 {
	return bank.accounts[address.String()].AmountOf(PNYXDenom).Int64()
}

func moduleBalance(bank *mockBankKeeper) int64 {
	return bank.modules[ModuleName].AmountOf(PNYXDenom).Int64()
}

func withEvidenceWindow(ctx sdk.Context, maxAgeBlocks int64, maxAgeDuration time.Duration) sdk.Context {
	return ctx.WithConsensusParams(cmtproto.ConsensusParams{
		Evidence: &cmtproto.EvidenceParams{
			MaxAgeNumBlocks: maxAgeBlocks,
			MaxAgeDuration:  maxAgeDuration,
		},
	})
}

func saveDomain(t *testing.T, keeper Keeper, ctx sdk.Context, domain Domain) {
	t.Helper()
	store := ctx.KVStore(keeper.StoreKey)
	store.Set([]byte("domain:"+domain.Name), keeper.cdc.MustMarshalLengthPrefixed(&domain))
}

func backExistingEscrow(keeper *Keeper, ctx sdk.Context) *mockBankKeeper {
	bank := newMockBankKeeper()
	bank.storeKey = keeper.StoreKey
	bank.fundModule(ModuleName, sdk.NewCoins(sdk.NewCoin(PNYXDenom, keeper.EscrowClaims(ctx))))
	keeper.bankKeeper = bank
	keeper.issuer = token.NewIssuanceService(bank, ModuleName)
	return bank
}

func TestEscrowLifecycleMaintainsExactParity(t *testing.T) {
	keeper, ctx, bank := setupKeeperWithBank(t)
	admin := sdk.AccAddress("escrow-admin")
	initial := int64(1_000_000 * PNYXUnit)
	fee := rewards.CalcPutPrice(sdk.NewInt64Coin(PNYXDenom, initial).Amount, 1).Int64()
	stake := rewards.StakeMin
	bank.fundAccount(admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, initial+fee+stake)))

	if err := keeper.CreateDomainWithEscrow(ctx, "Escrow", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, initial))); err != nil {
		t.Fatalf("create domain: %v", err)
	}
	if err := keeper.ValidateEscrowParity(ctx); err != nil {
		t.Fatalf("parity after domain creation: %v", err)
	}

	balanceBeforeDuplicate := accountBalance(bank, admin)
	if err := keeper.CreateDomainWithEscrow(ctx, "Escrow", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, initial))); err == nil {
		t.Fatal("expected duplicate domain to fail")
	}
	if got := accountBalance(bank, admin); got != balanceBeforeDuplicate {
		t.Fatalf("duplicate domain debited account: got %d, want %d", got, balanceBeforeDuplicate)
	}

	proposalFee := sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, fee))
	if err := keeper.SubmitProposalWithEscrow(ctx, admin, admin.String(), "Escrow", "Issue", "Suggestion", proposalFee, ""); err != nil {
		t.Fatalf("submit proposal: %v", err)
	}
	if err := keeper.ValidateEscrowParity(ctx); err != nil {
		t.Fatalf("parity after proposal: %v", err)
	}

	balanceBeforeDuplicate = accountBalance(bank, admin)
	if err := keeper.SubmitProposalWithEscrow(ctx, admin, admin.String(), "Escrow", "Issue", "Suggestion", proposalFee, ""); err == nil {
		t.Fatal("expected duplicate suggestion to fail")
	}
	if got := accountBalance(bank, admin); got != balanceBeforeDuplicate {
		t.Fatalf("duplicate suggestion debited account: got %d, want %d", got, balanceBeforeDuplicate)
	}

	stakeCoins := sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, stake))
	if err := keeper.RegisterValidatorWithEscrow(ctx, admin, admin.String(), testPubKey("escrow-validator"), stakeCoins, "Escrow"); err != nil {
		t.Fatalf("register validator: %v", err)
	}
	if err := keeper.ValidateEscrowParity(ctx); err != nil {
		t.Fatalf("parity after validator registration: %v", err)
	}

	balanceBeforeDuplicate = accountBalance(bank, admin)
	if err := keeper.RegisterValidatorWithEscrow(ctx, admin, admin.String(), testPubKey("escrow-validator"), stakeCoins, "Escrow"); err == nil {
		t.Fatal("expected duplicate validator to fail")
	}
	if got := accountBalance(bank, admin); got != balanceBeforeDuplicate {
		t.Fatalf("duplicate validator debited account: got %d, want %d", got, balanceBeforeDuplicate)
	}

	domain, _ := keeper.GetDomain(ctx, "Escrow")
	domain.TotalPayouts = stake * 10
	saveDomain(t, keeper, ctx, domain)
	ctx = withEvidenceWindow(ctx.WithBlockHeight(100), 5, 10*time.Minute)
	beforeWithdrawal := accountBalance(bank, admin)
	if err := keeper.RemoveValidatorWithEscrow(ctx, admin, admin.String()); err != nil {
		t.Fatalf("begin full validator exit: %v", err)
	}
	if got := accountBalance(bank, admin); got != beforeWithdrawal {
		t.Fatalf("validator exit paid stake before evidence window: got balance %d, want %d", got, beforeWithdrawal)
	}
	if _, found := keeper.GetValidator(ctx, admin.String()); found {
		t.Fatal("validator record remained after full withdrawal")
	}
	removal, found := keeper.GetPendingValidatorRemoval(ctx, admin.String())
	if !found {
		t.Fatal("validator exit did not create pending removal")
	}
	if got, want := removal.Validator.Stake.AmountOf(PNYXDenom).Int64(), stake; got != want {
		t.Fatalf("pending removal stake = %d, want %d", got, want)
	}
	if err := keeper.ValidateEscrowParity(ctx); err != nil {
		t.Fatalf("parity during pending validator removal: %v", err)
	}
	if got, want := moduleBalance(bank), initial+fee+stake; got != want {
		t.Fatalf("module balance during hold = %d, want %d", got, want)
	}

	retirementCtx := ctx.
		WithBlockHeight(removal.ConsensusRetiredHeight).
		WithBlockTime(ctx.BlockTime().Add(2 * time.Minute))
	if err := keeper.ProcessPendingValidatorRemovals(retirementCtx); err != nil {
		t.Fatalf("observe validator retirement: %v", err)
	}
	removal, _ = keeper.GetPendingValidatorRemoval(retirementCtx, admin.String())
	releaseCtx := retirementCtx.
		WithBlockHeight(removal.ReleaseAfterHeight + 1).
		WithBlockTime(time.Unix(0, removal.ReleaseAfterTimeNanos+1))
	if err := keeper.ProcessPendingValidatorRemovals(releaseCtx); err != nil {
		t.Fatalf("release validator exit: %v", err)
	}
	if got := accountBalance(bank, admin); got != beforeWithdrawal+stake {
		t.Fatalf("released stake amount = %d, want %d", got-beforeWithdrawal, stake)
	}
	if _, found := keeper.GetPendingValidatorRemoval(releaseCtx, admin.String()); found {
		t.Fatal("mature pending removal remained after payout")
	}
	if err := keeper.ValidateEscrowParity(releaseCtx); err != nil {
		t.Fatalf("parity after validator exit release: %v", err)
	}
	if got, want := moduleBalance(bank), initial+fee; got != want {
		t.Fatalf("module balance = %d, want remaining treasury %d", got, want)
	}
}

func TestPendingValidatorRemovalRequiresBothEvidenceLimits(t *testing.T) {
	keeper, ctx, bank := setupKeeperWithBank(t)
	operator := sdk.AccAddress("exit-window-operator")
	stake := int64(rewards.StakeMin)
	maxAgeBlocks := int64(5)
	maxAgeDuration := 10 * time.Minute
	ctx = withEvidenceWindow(ctx.WithBlockHeight(50), maxAgeBlocks, maxAgeDuration)

	keeper.CreateDomain(ctx, "ExitWindow", operator, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1)))
	domain, _ := keeper.GetDomain(ctx, "ExitWindow")
	domain.TotalPayouts = stake * 10
	saveDomain(t, keeper, ctx, domain)
	if err := keeper.RegisterValidator(ctx, operator.String(), testPubKey("exit-window"), sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, stake)), "ExitWindow"); err != nil {
		t.Fatal(err)
	}
	bank.fundModule(ModuleName, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, stake+1)))

	before := accountBalance(bank, operator)
	// A nominal stake withdrawal cannot bypass the hold by requesting the
	// validator's complete balance.
	if err := keeper.WithdrawStakeWithEscrow(ctx, operator, operator.String(), stake); err != nil {
		t.Fatal(err)
	}
	removal, found := keeper.GetPendingValidatorRemoval(ctx, operator.String())
	if !found {
		t.Fatal("pending removal not found")
	}
	if got, want := removal.ConsensusRetiredHeight, int64(52); got != want {
		t.Fatalf("retirement height = %d, want %d", got, want)
	}
	if got, want := removal.ReleaseAfterHeight, int64(56); got != want {
		t.Fatalf("release height = %d, want %d", got, want)
	}
	if removal.ReleaseAfterTimeNanos != 0 {
		t.Fatal("time boundary was set before consensus retirement was observed")
	}

	beforeRetirement := ctx.WithBlockHeight(51).WithBlockTime(ctx.BlockTime().Add(time.Minute))
	if err := keeper.ProcessPendingValidatorRemovals(beforeRetirement); err != nil {
		t.Fatal(err)
	}
	removal, _ = keeper.GetPendingValidatorRemoval(beforeRetirement, operator.String())
	if removal.ReleaseAfterTimeNanos != 0 {
		t.Fatal("time boundary was set before retirement height")
	}

	retirementTime := ctx.BlockTime().Add(2 * time.Minute)
	atRetirement := ctx.WithBlockHeight(removal.ConsensusRetiredHeight).WithBlockTime(retirementTime)
	if err := keeper.ProcessPendingValidatorRemovals(atRetirement); err != nil {
		t.Fatal(err)
	}
	removal, _ = keeper.GetPendingValidatorRemoval(atRetirement, operator.String())
	if got, want := removal.ConsensusRetiredAtNanos, retirementTime.UnixNano(); got != want {
		t.Fatalf("observed retirement time = %d, want %d", got, want)
	}
	if got, want := removal.ReleaseAfterTimeNanos, retirementTime.Add(maxAgeDuration).UnixNano(); got != want {
		t.Fatalf("release time = %d, want %d", got, want)
	}

	// The time limit has passed, but equality at the height limit is not
	// sufficient: both evidence limits must be strictly exceeded.
	atHeightBoundary := atRetirement.
		WithBlockHeight(removal.ReleaseAfterHeight).
		WithBlockTime(retirementTime.Add(maxAgeDuration + time.Minute))
	if err := keeper.ProcessPendingValidatorRemovals(atHeightBoundary); err != nil {
		t.Fatal(err)
	}
	if got := accountBalance(bank, operator); got != before {
		t.Fatalf("stake released at height boundary: balance %d, want %d", got, before)
	}

	// The height limit has passed, but equality at the time limit also keeps
	// the hold pending.
	atTimeBoundary := atRetirement.
		WithBlockHeight(removal.ReleaseAfterHeight + 1).
		WithBlockTime(time.Unix(0, removal.ReleaseAfterTimeNanos))
	if err := keeper.ProcessPendingValidatorRemovals(atTimeBoundary); err != nil {
		t.Fatal(err)
	}
	if got := accountBalance(bank, operator); got != before {
		t.Fatalf("stake released at time boundary: balance %d, want %d", got, before)
	}

	mature := atTimeBoundary.WithBlockTime(time.Unix(0, removal.ReleaseAfterTimeNanos+1))
	if err := keeper.ProcessPendingValidatorRemovals(mature); err != nil {
		t.Fatal(err)
	}
	if got, want := accountBalance(bank, operator), before+stake; got != want {
		t.Fatalf("mature stake payout balance = %d, want %d", got, want)
	}
	if _, found := keeper.GetPendingValidatorRemoval(mature, operator.String()); found {
		t.Fatal("mature hold was not deleted")
	}
	if err := keeper.ValidateEscrowParity(mature); err != nil {
		t.Fatalf("mature release broke escrow parity: %v", err)
	}
}

func TestPendingValidatorRemovalBlocksReregistrationAndRemainsSlashable(t *testing.T) {
	keeper, ctx, bank := setupKeeperWithBank(t)
	operator := sdk.AccAddress("held-validator-owner")
	stake := int64(rewards.StakeMin)
	ctx = withEvidenceWindow(ctx.WithBlockHeight(50), 5, 10*time.Minute)
	bank.fundAccount(operator, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, stake+1)))
	if err := keeper.CreateDomainWithEscrow(ctx, "HeldExit", operator, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1))); err != nil {
		t.Fatal(err)
	}
	domain, _ := keeper.GetDomain(ctx, "HeldExit")
	domain.TotalPayouts = stake * 10
	saveDomain(t, keeper, ctx, domain)
	oldKey := testPubKey("held-exit-key")
	if err := keeper.RegisterValidatorWithEscrow(
		ctx,
		operator,
		operator.String(),
		oldKey,
		sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, stake)),
		"HeldExit",
	); err != nil {
		t.Fatal(err)
	}
	keeper.setValidatorSigningInfo(ctx, ValidatorSigningInfo{
		OperatorAddr:             operator.String(),
		StartCommitHeight:        1,
		IndexOffset:              40,
		MissedBitmap:             make([]byte, livenessBitmapLength),
		LastObservedCommitHeight: 40,
	})
	if err := keeper.RemoveValidatorWithEscrow(ctx, operator, operator.String()); err != nil {
		t.Fatal(err)
	}
	if err := keeper.RegisterValidator(
		ctx,
		operator.String(),
		testPubKey("held-exit-replacement"),
		sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, stake)),
		"HeldExit",
	); err == nil {
		t.Fatal("operator re-registered while its evidence hold was pending")
	}

	if err := keeper.HandleDoubleSign(ctx, oldKey); err != nil {
		t.Fatal(err)
	}
	removal, found := keeper.GetPendingValidatorRemoval(ctx, operator.String())
	if !found {
		t.Fatal("slashing deleted the pending exit")
	}
	wantStake := int64(95_000 * PNYXUnit)
	if got := removal.Validator.Stake.AmountOf(PNYXDenom).Int64(); got != wantStake {
		t.Fatalf("held stake after slash = %d, want %d", got, wantStake)
	}
	domain, _ = keeper.GetDomain(ctx, "HeldExit")
	if domain.TransferredStake != wantStake {
		t.Fatalf("transferred stake = %d, want slash-adjusted %d", domain.TransferredStake, wantStake)
	}
	if got := bank.burned.AmountOf(PNYXDenom).Int64(); got != stake-wantStake {
		t.Fatalf("held stake burn = %d, want %d", got, stake-wantStake)
	}

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
	if _, found := keeper.getValidatorSigningInfo(releaseCtx, operator.String()); found {
		t.Fatal("mature exit payout retained orphaned signing state")
	}
	exported := NewAppModule(keeper.cdc, keeper).ExportGenesis(releaseCtx, nil)
	var exportedGenesis GenesisState
	if err := json.Unmarshal(exported, &exportedGenesis); err != nil {
		t.Fatal(err)
	}
	if err := ValidateGenesisState(exportedGenesis); err != nil {
		t.Fatalf("post-payout export is invalid: %v", err)
	}
	if err := keeper.RegisterValidator(
		releaseCtx,
		operator.String(),
		oldKey,
		sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, stake)),
		"HeldExit",
	); err == nil {
		t.Fatal("historical tombstoned consensus key was re-registered after hold release")
	}
	newKey := testPubKey("post-held-exit-new-key")
	if err := keeper.RegisterValidator(
		releaseCtx,
		operator.String(),
		newKey,
		sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, stake)),
		"HeldExit",
	); err != nil {
		t.Fatalf("operator could not re-register a fresh key after payout: %v", err)
	}
	if err := keeper.recordValidatorSignature(releaseCtx, operator.String(), releaseCtx.BlockHeight()+10, false); err != nil {
		t.Fatalf("fresh post-exit validator inherited a stale commit cursor: %v", err)
	}
}

func TestPartialValidatorWithdrawalIsFailClosedUntilSlashableUnbonding(t *testing.T) {
	keeper, ctx, bank := setupKeeperWithBank(t)
	operator := sdk.AccAddress("partial-exit-owner")
	stake := int64(200_000 * PNYXUnit)
	bank.fundAccount(operator, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, stake+1)))
	if err := keeper.CreateDomainWithEscrow(ctx, "PartialExit", operator, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1))); err != nil {
		t.Fatal(err)
	}
	if err := keeper.RegisterValidatorWithEscrow(
		ctx,
		operator,
		operator.String(),
		testPubKey("partial-exit-key"),
		sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, stake)),
		"PartialExit",
	); err != nil {
		t.Fatal(err)
	}
	accountBefore := accountBalance(bank, operator)
	moduleBefore := moduleBalance(bank)
	if err := keeper.WithdrawStakeWithEscrow(ctx, operator, operator.String(), 50_000*PNYXUnit); err == nil {
		t.Fatal("partial withdrawal bypassed slashable evidence custody")
	}
	validator, found := keeper.GetValidator(ctx, operator.String())
	if !found || validator.Stake.AmountOf(PNYXDenom).Int64() != stake {
		t.Fatal("rejected partial withdrawal changed validator stake")
	}
	if accountBalance(bank, operator) != accountBefore || moduleBalance(bank) != moduleBefore {
		t.Fatal("rejected partial withdrawal changed bank custody")
	}
}

func TestEscrowRejectsUnfundedAndSpoofedClaims(t *testing.T) {
	keeper, ctx, bank := setupKeeperWithBank(t)
	admin := sdk.AccAddress("claim-admin")
	attacker := sdk.AccAddress("claim-attacker")
	initialCoins := sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1_000*PNYXUnit))

	if err := keeper.CreateDomainWithEscrow(ctx, "Unfunded", admin, initialCoins); err == nil {
		t.Fatal("expected zero-balance domain declaration to fail")
	}
	if _, found := keeper.GetDomain(ctx, "Unfunded"); found {
		t.Fatal("failed domain transfer left internal domain state")
	}
	if moduleBalance(bank) != 0 {
		t.Fatal("failed domain transfer changed module balance")
	}

	bank.fundAccount(admin, initialCoins)
	if err := keeper.CreateDomainWithEscrow(ctx, "Claims", admin, initialCoins); err != nil {
		t.Fatal(err)
	}
	fee := sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, PNYXUnit))
	bank.fundAccount(attacker, fee)
	if err := keeper.SubmitProposalWithEscrow(ctx, attacker, admin.String(), "Claims", "I", "S", fee, ""); err == nil {
		t.Fatal("expected spoofed creator to fail")
	}
	if err := keeper.RegisterValidatorWithEscrow(ctx, attacker, admin.String(), testPubKey("spoof"), sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, rewards.StakeMin)), "Claims"); err == nil {
		t.Fatal("expected spoofed operator to fail")
	}
	if err := keeper.ValidateEscrowParity(ctx); err != nil {
		t.Fatalf("spoof rejection changed parity: %v", err)
	}
}

func TestEscrowPayoutFailureRollsBackCustomState(t *testing.T) {
	keeper, ctx, bank := setupKeeperWithBank(t)
	admin := sdk.AccAddress("rollback-admin")
	initial := int64(1_000_000 * PNYXUnit)
	fee := rewards.CalcPutPrice(sdk.NewInt64Coin(PNYXDenom, initial).Amount, 1).Int64()
	withdrawal := int64(rewards.StakeMin)
	stake := 2 * withdrawal
	bank.fundAccount(admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, initial+fee+stake)))

	if err := keeper.CreateDomainWithEscrow(ctx, "Rollback", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, initial))); err != nil {
		t.Fatal(err)
	}
	if err := keeper.SubmitProposalWithEscrow(ctx, admin, admin.String(), "Rollback", "Issue", "Suggestion", sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, fee)), ""); err != nil {
		t.Fatal(err)
	}
	if err := keeper.RegisterValidatorWithEscrow(ctx, admin, admin.String(), testPubKey("rollback-validator"), sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, stake)), "Rollback"); err != nil {
		t.Fatal(err)
	}

	domainBefore, _ := keeper.GetDomain(ctx, "Rollback")
	domainBefore.TotalPayouts = stake * 10
	saveDomain(t, keeper, ctx, domainBefore)
	moduleBefore := moduleBalance(bank)
	accountBefore := accountBalance(bank, admin)
	bank.failModuleToAccount = true

	if _, err := keeper.PlaceStoneOnIssueWithPayout(ctx, admin, "Rollback", "Issue", admin.String()); err == nil {
		t.Fatal("expected injected reward payout failure")
	}
	domainAfter, _ := keeper.GetDomain(ctx, "Rollback")
	if domainAfter.Issues[0].Stones != domainBefore.Issues[0].Stones ||
		!domainAfter.Treasury.AmountOf(PNYXDenom).Equal(domainBefore.Treasury.AmountOf(PNYXDenom)) {
		t.Fatal("failed reward payout committed custom state")
	}
	if _, found := keeper.GetMemberIssueStone(ctx, "Rollback", admin.String()); found {
		t.Fatal("failed reward payout committed stone index")
	}

	if err := keeper.WithdrawStakeWithEscrow(ctx, admin, admin.String(), withdrawal); err == nil {
		t.Fatal("expected injected stake payout failure")
	}
	validator, found := keeper.GetValidator(ctx, admin.String())
	if !found || validator.Stake.AmountOf(PNYXDenom).Int64() != stake {
		t.Fatal("failed stake payout changed validator claim")
	}
	domainAfter, _ = keeper.GetDomain(ctx, "Rollback")
	if domainAfter.TransferredStake != domainBefore.TransferredStake {
		t.Fatal("failed stake payout changed transfer accounting")
	}
	if moduleBalance(bank) != moduleBefore || accountBalance(bank, admin) != accountBefore {
		t.Fatal("failed module payouts changed bank balances")
	}
	if err := keeper.ValidateEscrowParity(ctx); err != nil {
		t.Fatalf("payout rollback broke parity: %v", err)
	}
}

func TestSlashBurnsEscrowAndPreservesParity(t *testing.T) {
	keeper, ctx, bank := setupKeeperWithBank(t)
	admin := sdk.AccAddress("slash-admin")
	initial := int64(500_000 * PNYXUnit)
	stake := rewards.StakeMin
	bank.fundAccount(admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, initial+stake)))
	if err := keeper.CreateDomainWithEscrow(ctx, "Slash", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, initial))); err != nil {
		t.Fatal(err)
	}
	pubKey := testPubKey("slash-validator")
	if err := keeper.RegisterValidatorWithEscrow(ctx, admin, admin.String(), pubKey, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, stake)), "Slash"); err != nil {
		t.Fatal(err)
	}

	if err := keeper.HandleDoubleSign(ctx, pubKey); err != nil {
		t.Fatal(err)
	}
	validator, _ := keeper.GetValidator(ctx, admin.String())
	penalty := stake - validator.Stake.AmountOf(PNYXDenom).Int64()
	domain, _ := keeper.GetDomain(ctx, "Slash")
	if got := domain.Treasury.AmountOf(PNYXDenom).Int64(); got != initial {
		t.Fatalf("slash changed domain treasury to %d, want %d", got, initial)
	}
	if got := bank.burned.AmountOf(PNYXDenom).Int64(); got != penalty {
		t.Fatalf("burned amount = %d, want penalty %d", got, penalty)
	}
	if err := keeper.ValidateEscrowParity(ctx); err != nil {
		t.Fatalf("slash broke escrow parity: %v", err)
	}
}

func TestSlashBurnFailureDoesNotChangeValidatorClaim(t *testing.T) {
	keeper, ctx, bank := setupKeeperWithBank(t)
	admin := sdk.AccAddress("slash-failure-admin")
	initial := int64(500_000 * PNYXUnit)
	stake := rewards.StakeMin
	bank.fundAccount(admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, initial+stake)))
	if err := keeper.CreateDomainWithEscrow(ctx, "SlashFailure", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, initial))); err != nil {
		t.Fatal(err)
	}
	pubKey := testPubKey("slash-failure-validator")
	if err := keeper.RegisterValidatorWithEscrow(ctx, admin, admin.String(), pubKey, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, stake)), "SlashFailure"); err != nil {
		t.Fatal(err)
	}

	moduleBefore := moduleBalance(bank)
	bank.failBurn = true
	if err := keeper.HandleDoubleSign(ctx, pubKey); err == nil {
		t.Fatal("expected injected slash burn failure")
	}
	validator, found := keeper.GetValidator(ctx, admin.String())
	if !found || validator.Stake.AmountOf(PNYXDenom).Int64() != stake || validator.Jailed {
		t.Fatal("failed slash burn changed validator state")
	}
	if moduleBalance(bank) != moduleBefore || !bank.burned.Empty() {
		t.Fatal("failed slash burn changed bank accounting")
	}
	if err := keeper.ValidateEscrowParity(ctx); err != nil {
		t.Fatalf("failed slash burn broke escrow parity: %v", err)
	}
}

func TestEscrowMessagesRejectInvalidCoinSetsAndSignerClaims(t *testing.T) {
	sender := sdk.AccAddress("message-sender")
	validCoin := sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1))
	tests := []struct {
		name string
		msg  interface{ ValidateBasic() error }
	}{
		{"domain without funding", &MsgCreateDomain{Name: "D", Admin: sender}},
		{"proposal wrong denom", &MsgSubmitProposal{Sender: sender, Creator: sender.String(), DomainName: "D", IssueName: "I", SuggestionName: "S", Fee: sdk.NewCoins(sdk.NewInt64Coin("atom", 1))}},
		{"proposal spoof", &MsgSubmitProposal{Sender: sender, Creator: "other", DomainName: "D", IssueName: "I", SuggestionName: "S", Fee: validCoin}},
		{"validator mixed denom", &MsgRegisterValidator{Sender: sender, OperatorAddr: sender.String(), PubKey: "00", DomainName: "D", Stake: sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1), sdk.NewInt64Coin("atom", 1))}},
		{"validator spoof", &MsgRegisterValidator{Sender: sender, OperatorAddr: "other", PubKey: "00", DomainName: "D", Stake: validCoin}},
		{"voter spoof", &MsgCastElectionVote{Sender: sender, DomainName: "D", IssueName: "I", CandidateName: "C", VoterAddr: "other", Choice: 0}},
		{"deposit without sender", &MsgDepositToDomain{DomainName: "D", Amount: validCoin[0]}},
		{"withdrawal without sender", &MsgWithdrawFromDomain{DomainName: "D", Recipient: sender.String(), Amount: validCoin[0]}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.msg.ValidateBasic(); err == nil {
				t.Fatal("expected ValidateBasic error")
			}
		})
	}
}

func TestEscrowParityRejectsUnexpectedDenomination(t *testing.T) {
	keeper, ctx, bank := setupKeeperWithBank(t)
	bank.fundModule(ModuleName, sdk.NewCoins(sdk.NewInt64Coin("unexpected", 1)))
	if err := keeper.ValidateEscrowParity(ctx); err == nil {
		t.Fatal("escrow parity accepted an unclaimed denomination")
	}
}
