package truedemocracy

import (
	"testing"

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
	beforeWithdrawal := accountBalance(bank, admin)
	if err := keeper.RemoveValidatorWithEscrow(ctx, admin, admin.String()); err != nil {
		t.Fatalf("withdraw full validator stake: %v", err)
	}
	if got := accountBalance(bank, admin); got != beforeWithdrawal+stake {
		t.Fatalf("stake withdrawal amount = %d, want %d", got-beforeWithdrawal, stake)
	}
	if _, found := keeper.GetValidator(ctx, admin.String()); found {
		t.Fatal("validator record remained after full withdrawal")
	}
	if err := keeper.ValidateEscrowParity(ctx); err != nil {
		t.Fatalf("parity after stake withdrawal: %v", err)
	}
	if got, want := moduleBalance(bank), initial+fee; got != want {
		t.Fatalf("module balance = %d, want remaining treasury %d", got, want)
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
	stake := rewards.StakeMin
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

	if err := keeper.WithdrawStakeWithEscrow(ctx, admin, admin.String(), stake); err == nil {
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
