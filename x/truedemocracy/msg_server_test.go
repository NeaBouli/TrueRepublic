package truedemocracy

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// requireGovMsgEvent fails the test unless the context event manager recorded
// an event of the given type.
func requireGovMsgEvent(t *testing.T, ctx sdk.Context, eventType string) {
	t.Helper()
	for _, event := range ctx.EventManager().Events() {
		if event.Type == eventType {
			return
		}
	}
	t.Fatalf("event %q was not emitted; got %v", eventType, ctx.EventManager().Events())
}

func findSuggestion(domain Domain, issueName, suggestionName string) bool {
	for _, issue := range domain.Issues {
		if issue.Name != issueName {
			continue
		}
		for _, suggestion := range issue.Suggestions {
			if suggestion.Name == suggestionName {
				return true
			}
		}
	}
	return false
}

// TestMsgServerCreateDomainEscrowBoundary proves domain creation escrows the
// declared treasury atomically: failures commit nothing, success moves the
// exact coins into the module account.
func TestMsgServerCreateDomainEscrowBoundary(t *testing.T) {
	k, ctx, bk := setupKeeperWithBank(t)
	server := NewMsgServer(k)
	goCtx := sdk.WrapSDKContext(ctx)
	admin := sdk.AccAddress("domain-admin")
	initial := sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1_000_000))

	// An unfunded admin must fail without committing the domain.
	if _, err := server.CreateDomain(goCtx, &MsgCreateDomain{
		Name: "Gov", Admin: admin, InitialCoins: initial,
	}); err == nil {
		t.Fatal("domain creation without escrow funds succeeded")
	}
	if _, found := k.GetDomain(ctx, "Gov"); found {
		t.Fatal("failed domain creation committed state")
	}

	// Coins that are not exactly one positive PNYX coin must be rejected.
	bk.fundAccount(admin, initial)
	if _, err := server.CreateDomain(goCtx, &MsgCreateDomain{
		Name: "Empty", Admin: admin, InitialCoins: sdk.NewCoins(),
	}); err == nil {
		t.Fatal("domain creation with empty initial coins succeeded")
	}
	if _, found := k.GetDomain(ctx, "Empty"); found {
		t.Fatal("rejected domain creation committed state")
	}

	if _, err := server.CreateDomain(goCtx, &MsgCreateDomain{
		Name: "Gov", Admin: admin, InitialCoins: initial,
	}); err != nil {
		t.Fatalf("escrowed domain creation failed: %v", err)
	}
	domain, found := k.GetDomain(ctx, "Gov")
	if !found {
		t.Fatal("domain missing after escrowed creation")
	}
	if !domain.Treasury.Equal(initial) {
		t.Fatalf("domain treasury = %s, want %s", domain.Treasury, initial)
	}
	if len(domain.Members) != 1 || domain.Members[0] != admin.String() {
		t.Fatalf("domain members = %v, want the admin only", domain.Members)
	}
	if got := bk.accounts[admin.String()].AmountOf(PNYXDenom); !got.IsZero() {
		t.Fatalf("admin still holds %s after escrow", got)
	}
	if got := bk.modules[ModuleName].AmountOf(PNYXDenom); !got.Equal(initial[0].Amount) {
		t.Fatalf("module escrow = %s, want %s", got, initial[0].Amount)
	}
	requireGovMsgEvent(t, ctx, "create_domain")

	// A duplicate domain must fail without changing the stored domain.
	if _, err := server.CreateDomain(goCtx, &MsgCreateDomain{
		Name: "Gov", Admin: admin, InitialCoins: initial,
	}); err == nil {
		t.Fatal("duplicate domain creation succeeded")
	}
	domain, _ = k.GetDomain(ctx, "Gov")
	if !domain.Treasury.Equal(initial) || len(domain.Members) != 1 {
		t.Fatal("duplicate domain creation mutated the stored domain")
	}
}

// TestMsgServerSubmitProposalEscrowBoundary proves the proposal handler
// rejects creator spoofing, rolls back failed fee escrow, and commits the
// suggestion plus fee on success.
func TestMsgServerSubmitProposalEscrowBoundary(t *testing.T) {
	k, ctx, bk := setupKeeperWithBank(t)
	server := NewMsgServer(k)
	goCtx := sdk.WrapSDKContext(ctx)
	admin := sdk.AccAddress("proposal-admin")
	creator := sdk.AccAddress("proposal-creator")
	initial := sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1_000_000))
	// With a 1,000,000 upnyx treasury and one member the put price (eq.3) is
	// exactly 1,000 upnyx.
	fee := sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1_000))
	bk.fundAccount(admin, initial)
	if err := k.CreateDomainWithEscrow(ctx, "Gov", admin, initial); err != nil {
		t.Fatal(err)
	}

	// A claimed creator that is not the authenticated sender must fail.
	if _, err := server.SubmitProposal(goCtx, &MsgSubmitProposal{
		Sender: creator, Creator: "spoofed-creator",
		DomainName: "Gov", IssueName: "Climate", SuggestionName: "GreenDeal",
		Fee: fee,
	}); err == nil {
		t.Fatal("creator spoofing was accepted")
	}
	if domain, _ := k.GetDomain(ctx, "Gov"); len(domain.Issues) != 0 {
		t.Fatal("spoofed proposal committed domain state")
	}

	// A valid claim without funds must fail the escrow without committing.
	msg := &MsgSubmitProposal{
		Sender: creator, Creator: creator.String(),
		DomainName: "Gov", IssueName: "Climate", SuggestionName: "GreenDeal",
		Fee: fee,
	}
	if _, err := server.SubmitProposal(goCtx, msg); err == nil {
		t.Fatal("proposal without escrow funds succeeded")
	}
	domain, _ := k.GetDomain(ctx, "Gov")
	if len(domain.Issues) != 0 || !domain.Treasury.Equal(initial) {
		t.Fatal("failed proposal escrow mutated domain state")
	}

	// The funded, authenticated proposal succeeds end to end.
	bk.fundAccount(creator, fee)
	if _, err := server.SubmitProposal(goCtx, msg); err != nil {
		t.Fatalf("escrowed proposal failed: %v", err)
	}
	domain, _ = k.GetDomain(ctx, "Gov")
	if !findSuggestion(domain, "Climate", "GreenDeal") {
		t.Fatal("committed proposal is missing the suggestion")
	}
	wantTreasury := initial.Add(fee...)
	if !domain.Treasury.Equal(wantTreasury) {
		t.Fatalf("domain treasury = %s, want %s", domain.Treasury, wantTreasury)
	}
	if got := bk.accounts[creator.String()].AmountOf(PNYXDenom); !got.IsZero() {
		t.Fatalf("creator still holds %s after fee escrow", got)
	}
	if got := bk.modules[ModuleName].AmountOf(PNYXDenom); !got.Equal(wantTreasury[0].Amount) {
		t.Fatalf("module escrow = %s, want %s", got, wantTreasury[0].Amount)
	}
	requireGovMsgEvent(t, ctx, "submit_proposal")

	// The same suggestion name must be rejected without a state change.
	if _, err := server.SubmitProposal(goCtx, msg); err == nil {
		t.Fatal("duplicate suggestion was accepted")
	}
	domain, _ = k.GetDomain(ctx, "Gov")
	if !domain.Treasury.Equal(wantTreasury) || len(domain.Issues) != 1 {
		t.Fatal("duplicate proposal mutated domain state")
	}
}

// TestMsgServerAddMemberAuthorization proves only the domain admin can add
// members through the handler and that rejections never change membership.
func TestMsgServerAddMemberAuthorization(t *testing.T) {
	k, ctx, _ := setupKeeperWithBank(t)
	server := NewMsgServer(k)
	goCtx := sdk.WrapSDKContext(ctx)
	admin := sdk.AccAddress("member-admin")
	k.CreateDomain(ctx, "Gov", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1_000)))

	if _, err := server.AddMember(goCtx, &MsgAddMember{
		Sender: sdk.AccAddress("member-attacker"), DomainName: "Gov", NewMember: "mallory",
	}); err == nil {
		t.Fatal("non-admin added a member")
	}
	if domain, _ := k.GetDomain(ctx, "Gov"); len(domain.Members) != 1 {
		t.Fatal("unauthorized add mutated membership")
	}

	if _, err := server.AddMember(goCtx, &MsgAddMember{
		Sender: admin, DomainName: "Gov", NewMember: "new-member",
	}); err != nil {
		t.Fatalf("admin add member failed: %v", err)
	}
	domain, _ := k.GetDomain(ctx, "Gov")
	if len(domain.Members) != 2 || domain.Members[1] != "new-member" {
		t.Fatalf("members = %v, want admin plus new-member", domain.Members)
	}
	requireGovMsgEvent(t, ctx, "add_member")

	if _, err := server.AddMember(goCtx, &MsgAddMember{
		Sender: admin, DomainName: "Gov", NewMember: "new-member",
	}); err == nil {
		t.Fatal("duplicate member was accepted")
	}
	if _, err := server.AddMember(goCtx, &MsgAddMember{
		Sender: admin, DomainName: "NoSuchDomain", NewMember: "ghost",
	}); err == nil {
		t.Fatal("add member on a missing domain succeeded")
	}
	if domain, _ := k.GetDomain(ctx, "Gov"); len(domain.Members) != 2 {
		t.Fatal("rejected adds mutated membership")
	}
}

// TestMsgServerPlaceStoneOnIssueBoundary proves stone placement requires the
// authenticated member, pays the VoteToEarn reward from escrow, and rejects
// duplicates without a state change.
func TestMsgServerPlaceStoneOnIssueBoundary(t *testing.T) {
	k, ctx, bk := setupKeeperWithBank(t)
	server := NewMsgServer(k)
	goCtx := sdk.WrapSDKContext(ctx)
	admin := sdk.AccAddress("stone-admin")
	creator := sdk.AccAddress("stone-creator")
	alice := sdk.AccAddress("stone-alice")
	bk.fundAccount(admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1_000_000)))
	if err := k.CreateDomainWithEscrow(ctx, "Gov", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1_000_000))); err != nil {
		t.Fatal(err)
	}
	fee := sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1_000))
	bk.fundAccount(creator, fee)
	if err := k.SubmitProposalWithEscrow(ctx, creator, creator.String(), "Gov", "Climate", "GreenDeal", fee, ""); err != nil {
		t.Fatal(err)
	}
	if err := k.AddMember(ctx, "Gov", alice.String(), admin); err != nil {
		t.Fatal(err)
	}

	// A member claim that is not the authenticated sender must fail.
	if _, err := server.PlaceStoneOnIssue(goCtx, &MsgPlaceStoneOnIssue{
		Sender: alice, DomainName: "Gov", IssueName: "Climate", MemberAddr: "mallory",
	}); err == nil {
		t.Fatal("member spoofing was accepted")
	}
	if got := ctx.KVStore(k.StoreKey).Get(issueStoneKey("Gov", "mallory")); got != nil {
		t.Fatal("spoofed placement committed a stone")
	}

	// A non-member cannot place a stone.
	bob := sdk.AccAddress("stone-outsider")
	if _, err := server.PlaceStoneOnIssue(goCtx, &MsgPlaceStoneOnIssue{
		Sender: bob, DomainName: "Gov", IssueName: "Climate", MemberAddr: bob.String(),
	}); err == nil {
		t.Fatal("non-member placed a stone")
	}

	// The treasury is 1,001,000 upnyx after the fee escrow, so the VoteToEarn
	// reward (eq.2) is exactly 1,001 upnyx.
	if _, err := server.PlaceStoneOnIssue(goCtx, &MsgPlaceStoneOnIssue{
		Sender: alice, DomainName: "Gov", IssueName: "Climate", MemberAddr: alice.String(),
	}); err != nil {
		t.Fatalf("member stone placement failed: %v", err)
	}
	if got := ctx.KVStore(k.StoreKey).Get(issueStoneKey("Gov", alice.String())); string(got) != "Climate" {
		t.Fatalf("alice stone = %q, want Climate", got)
	}
	if got := bk.accounts[alice.String()].AmountOf(PNYXDenom); got.Int64() != 1_001 {
		t.Fatalf("alice reward = %s, want 1001", got)
	}
	domain, _ := k.GetDomain(ctx, "Gov")
	if got := domain.Treasury.AmountOf(PNYXDenom); got.Int64() != 999_999 {
		t.Fatalf("domain treasury = %s, want 999999", got)
	}
	if domain.TotalPayouts != 1_001 {
		t.Fatalf("total payouts = %d, want 1001", domain.TotalPayouts)
	}
	requireGovMsgEvent(t, ctx, "place_stone_issue")

	// Re-placing on the same issue must fail without moving funds or stones.
	if _, err := server.PlaceStoneOnIssue(goCtx, &MsgPlaceStoneOnIssue{
		Sender: alice, DomainName: "Gov", IssueName: "Climate", MemberAddr: alice.String(),
	}); err == nil {
		t.Fatal("duplicate stone placement was accepted")
	}
	if got := bk.accounts[alice.String()].AmountOf(PNYXDenom); got.Int64() != 1_001 {
		t.Fatalf("duplicate placement changed alice balance to %s", got)
	}
	domain, _ = k.GetDomain(ctx, "Gov")
	if got := domain.Treasury.AmountOf(PNYXDenom); got.Int64() != 999_999 {
		t.Fatalf("duplicate placement changed treasury to %s", got)
	}
}

// TestMsgServerWithdrawFromDomainBoundary proves withdrawals require a valid
// bech32 recipient, the domain admin, and a covered amount; failures never
// move treasury or funds.
func TestMsgServerWithdrawFromDomainBoundary(t *testing.T) {
	k, ctx, bk := setupKeeperWithBank(t)
	server := NewMsgServer(k)
	goCtx := sdk.WrapSDKContext(ctx)
	admin := sdk.AccAddress("withdraw-admin")
	recipient := sdk.AccAddress("withdraw-recipient")
	bk.fundAccount(admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1_000_000)))
	if err := k.CreateDomainWithEscrow(ctx, "Gov", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1_000_000))); err != nil {
		t.Fatal(err)
	}

	assertTreasury := func(want int64) {
		t.Helper()
		domain, _ := k.GetDomain(ctx, "Gov")
		if got := domain.Treasury.AmountOf(PNYXDenom); got.Int64() != want {
			t.Fatalf("domain treasury = %s, want %d", got, want)
		}
	}

	// An invalid bech32 recipient is rejected before any state transition.
	if _, err := server.WithdrawFromDomain(goCtx, &MsgWithdrawFromDomain{
		Sender: admin, DomainName: "Gov",
		Recipient: "not-a-bech32-address",
		Amount:    sdk.NewInt64Coin(PNYXDenom, 10),
	}); err == nil {
		t.Fatal("withdrawal with an invalid recipient succeeded")
	}
	assertTreasury(1_000_000)

	// A non-admin authorizer cannot withdraw.
	if _, err := server.WithdrawFromDomain(goCtx, &MsgWithdrawFromDomain{
		Sender: sdk.AccAddress("withdraw-stranger"), DomainName: "Gov",
		Recipient: recipient.String(),
		Amount:    sdk.NewInt64Coin(PNYXDenom, 10),
	}); err == nil {
		t.Fatal("non-admin withdrawal succeeded")
	}
	assertTreasury(1_000_000)

	// The amount must be covered by the treasury.
	if _, err := server.WithdrawFromDomain(goCtx, &MsgWithdrawFromDomain{
		Sender: admin, DomainName: "Gov",
		Recipient: recipient.String(),
		Amount:    sdk.NewInt64Coin(PNYXDenom, 2_000_000),
	}); err == nil {
		t.Fatal("overdrawn withdrawal succeeded")
	}
	assertTreasury(1_000_000)

	// The admin withdrawal succeeds and settles exactly.
	if _, err := server.WithdrawFromDomain(goCtx, &MsgWithdrawFromDomain{
		Sender: admin, DomainName: "Gov",
		Recipient: recipient.String(),
		Amount:    sdk.NewInt64Coin(PNYXDenom, 100_000),
	}); err != nil {
		t.Fatalf("admin withdrawal failed: %v", err)
	}
	assertTreasury(900_000)
	if got := bk.accounts[recipient.String()].AmountOf(PNYXDenom); got.Int64() != 100_000 {
		t.Fatalf("recipient balance = %s, want 100000", got)
	}
	if got := bk.modules[ModuleName].AmountOf(PNYXDenom); got.Int64() != 900_000 {
		t.Fatalf("module escrow = %s, want 900000", got)
	}
	requireGovMsgEvent(t, ctx, "withdraw_from_domain")
}
