package truedemocracy

import (
	"encoding/hex"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkbech32 "github.com/cosmos/cosmos-sdk/types/bech32"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	rewards "truerepublic/treasury/keeper"
)

// GH-209 acceptance coverage: recipient-bound v2 payloads, atomic
// treasury-funded payout to only the bound recipient, fail-closed rollback,
// and state round-trip without deferred-reward resurrection.

func TestValidateRewardRecipient(t *testing.T) {
	valid := testRewardRecipient()
	raw, err := sdk.AccAddressFromBech32(valid)
	if err != nil {
		t.Fatal(err)
	}
	wrongPrefix, err := sdkbech32.ConvertAndEncode("foreignchain", raw)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := ValidateRewardRecipient(valid); err != nil {
		t.Fatalf("canonical recipient rejected: %v", err)
	}
	cases := map[string]string{
		"empty":        "",
		"uppercase":    strings.ToUpper(valid),
		"wrong prefix": wrongPrefix,
		"malformed":    "not-a-bech32-address",
		"truncated":    valid[:len(valid)-1],
	}
	for name, candidate := range cases {
		if _, err := ValidateRewardRecipient(candidate); err == nil {
			t.Fatalf("%s recipient accepted: %q", name, candidate)
		}
	}
}

// zkpPayoutFixture builds a ZKP domain with a backed escrow and returns the
// exact pre-rating accounting snapshot.
type payoutSnapshot struct {
	treasury      int64
	totalPayouts  int64
	moduleBalance int64
}

func snapshotAccounting(t *testing.T, k Keeper, ctx sdk.Context, bank *mockBankKeeper, domainName string) payoutSnapshot {
	t.Helper()
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		t.Fatalf("domain %s not found", domainName)
	}
	return payoutSnapshot{
		treasury:      domain.Treasury.AmountOf(PNYXDenom).Int64(),
		totalPayouts:  domain.TotalPayouts,
		moduleBalance: moduleBalance(bank),
	}
}

func expectedReward(treasuryBefore int64) int64 {
	return rewards.CalcReward(sdk.NewInt64Coin(PNYXDenom, treasuryBefore).Amount).Int64()
}

func assertUnchangedAfterFailure(t *testing.T, k Keeper, ctx sdk.Context, bank *mockBankKeeper, domainName, nullifierHex string, before payoutSnapshot, watcher sdk.AccAddress) {
	t.Helper()
	domain, _ := k.GetDomain(ctx, domainName)
	if got := domain.Treasury.AmountOf(PNYXDenom).Int64(); got != before.treasury {
		t.Fatalf("treasury mutated on failure: %d, want %d", got, before.treasury)
	}
	if domain.TotalPayouts != before.totalPayouts {
		t.Fatalf("total payouts mutated on failure: %d, want %d", domain.TotalPayouts, before.totalPayouts)
	}
	for _, issue := range domain.Issues {
		for _, suggestion := range issue.Suggestions {
			if got := len(suggestion.Ratings); got != 0 {
				t.Fatalf("failure recorded %d ratings", got)
			}
		}
	}
	if nullifierHex != "" && k.IsNullifierUsed(ctx, domainName, nullifierHex) {
		t.Fatal("failure consumed the nullifier")
	}
	if got := moduleBalance(bank); got != before.moduleBalance {
		t.Fatalf("module balance mutated on failure: %d, want %d", got, before.moduleBalance)
	}
	if got := accountBalance(bank, watcher); got != 0 {
		t.Fatalf("watcher received funds on failure: %d", got)
	}
	if err := k.ValidateEscrowParity(ctx); err != nil {
		t.Fatalf("escrow parity broken after failure: %v", err)
	}
}

func TestRateWithProofRecipientBoundPayoutAccounting(t *testing.T) {
	k, ctx := setupKeeper(t)
	secrets := setupDomainWithZKPIdentity(t, k, ctx, "ZKPDomain", 3)
	addProposal(t, k, ctx, "ZKPDomain", "Climate", "GreenDeal")
	bank := backExistingEscrow(&k, ctx)
	srv := NewMsgServer(k)

	recipient := sdk.AccAddress("bound-recipient")
	sender := sdk.AccAddress("relayer-sender")
	proofHex, nullifierHex := generateZKPRatingForRecipient(t, k, ctx, "ZKPDomain", secrets, 1, "Climate", "GreenDeal", 4, recipient.String())

	before := snapshotAccounting(t, k, ctx, bank, "ZKPDomain")
	reward := expectedReward(before.treasury)
	if reward <= 0 {
		t.Fatal("expected a positive reward for this fixture")
	}

	msg := &MsgRateWithProof{
		Sender:          sender,
		DomainName:      "ZKPDomain",
		IssueName:       "Climate",
		SuggestionName:  "GreenDeal",
		Rating:          4,
		Proof:           proofHex,
		NullifierHash:   nullifierHex,
		RewardRecipient: recipient.String(),
	}
	if _, err := srv.RateWithProof(ctx, msg); err != nil {
		t.Fatalf("bound-recipient rating failed: %v", err)
	}

	domain, _ := k.GetDomain(ctx, "ZKPDomain")
	if got := domain.Treasury.AmountOf(PNYXDenom).Int64(); got != before.treasury-reward {
		t.Fatalf("treasury = %d, want %d", got, before.treasury-reward)
	}
	if domain.TotalPayouts != before.totalPayouts+reward {
		t.Fatalf("total payouts = %d, want %d", domain.TotalPayouts, before.totalPayouts+reward)
	}
	if got := moduleBalance(bank); got != before.moduleBalance-reward {
		t.Fatalf("module balance = %d, want %d", got, before.moduleBalance-reward)
	}
	if got := accountBalance(bank, recipient); got != reward {
		t.Fatalf("bound recipient balance = %d, want exact reward %d", got, reward)
	}
	if got := accountBalance(bank, sender); got != 0 {
		t.Fatalf("reward paid to the transaction sender: %d", got)
	}
	if !k.IsNullifierUsed(ctx, "ZKPDomain", nullifierHex) {
		t.Fatal("nullifier not consumed after success")
	}
	ratings := domain.Issues[0].Suggestions[0].Ratings
	if len(ratings) != 1 || ratings[0].NullifierHex != nullifierHex {
		t.Fatalf("rating not recorded exactly once: %+v", ratings)
	}
	if err := k.ValidateEscrowParity(ctx); err != nil {
		t.Fatalf("escrow parity after payout: %v", err)
	}
}

func TestRateProposalSignatureRecipientBoundPayout(t *testing.T) {
	k, ctx := setupKeeper(t)
	ctx = ctx.WithChainID("truerepublic-test-1")
	setupDomainWithIssue(t, k, ctx)
	aliceKey := domainKey("alice-sig-payout")
	if err := k.JoinPermissionRegister(ctx, "AnonDomain", "alice", aliceKey.PubKey().Bytes()); err != nil {
		t.Fatal(err)
	}
	bank := backExistingEscrow(&k, ctx)
	srv := NewMsgServer(k)

	recipient := sdk.AccAddress("sig-recipient")
	sender := sdk.AccAddress("relayer-sender")
	payload := encodeVoteContextV2(ctx.ChainID(), "AnonDomain", "Climate", "GreenDeal", 3, recipient.String())
	sig, err := aliceKey.Sign(payload)
	if err != nil {
		t.Fatal(err)
	}

	before := snapshotAccounting(t, k, ctx, bank, "AnonDomain")
	reward := expectedReward(before.treasury)

	msg := &MsgRateProposal{
		Sender:          sender,
		DomainName:      "AnonDomain",
		IssueName:       "Climate",
		SuggestionName:  "GreenDeal",
		Rating:          3,
		DomainPubKey:    hex.EncodeToString(aliceKey.PubKey().Bytes()),
		Signature:       hex.EncodeToString(sig),
		RewardRecipient: recipient.String(),
	}
	if _, err := srv.RateProposal(ctx, msg); err != nil {
		t.Fatalf("bound-recipient signature rating failed: %v", err)
	}

	domain, _ := k.GetDomain(ctx, "AnonDomain")
	if got := domain.Treasury.AmountOf(PNYXDenom).Int64(); got != before.treasury-reward {
		t.Fatalf("treasury = %d, want %d", got, before.treasury-reward)
	}
	if domain.TotalPayouts != before.totalPayouts+reward {
		t.Fatalf("total payouts = %d, want %d", domain.TotalPayouts, before.totalPayouts+reward)
	}
	if got := moduleBalance(bank); got != before.moduleBalance-reward {
		t.Fatalf("module balance = %d, want %d", got, before.moduleBalance-reward)
	}
	if got := accountBalance(bank, recipient); got != reward {
		t.Fatalf("bound recipient balance = %d, want exact reward %d", got, reward)
	}
	if got := accountBalance(bank, sender); got != 0 {
		t.Fatalf("reward paid to the transaction sender: %d", got)
	}
	if err := k.ValidateEscrowParity(ctx); err != nil {
		t.Fatalf("escrow parity after payout: %v", err)
	}
}

func TestRateWithProofRecipientSubstitutionRejected(t *testing.T) {
	k, ctx := setupKeeper(t)
	secrets := setupDomainWithZKPIdentity(t, k, ctx, "ZKPDomain", 3)
	addProposal(t, k, ctx, "ZKPDomain", "Climate", "GreenDeal")
	bank := backExistingEscrow(&k, ctx)
	srv := NewMsgServer(k)

	bound := sdk.AccAddress("bound-recipient")
	attacker := sdk.AccAddress("attacker-recipient")
	proofHex, nullifierHex := generateZKPRatingForRecipient(t, k, ctx, "ZKPDomain", secrets, 0, "Climate", "GreenDeal", 3, bound.String())

	before := snapshotAccounting(t, k, ctx, bank, "ZKPDomain")

	// A copied submission with a substituted recipient cannot redirect the
	// reward: the proof no longer matches the v2 signal.
	msg := &MsgRateWithProof{
		Sender:          attacker,
		DomainName:      "ZKPDomain",
		IssueName:       "Climate",
		SuggestionName:  "GreenDeal",
		Rating:          3,
		Proof:           proofHex,
		NullifierHash:   nullifierHex,
		RewardRecipient: attacker.String(),
	}
	if _, err := srv.RateWithProof(ctx, msg); err == nil {
		t.Fatal("recipient-substituted submission accepted")
	}
	assertUnchangedAfterFailure(t, k, ctx, bank, "ZKPDomain", nullifierHex, before, attacker)
	if got := accountBalance(bank, bound); got != 0 {
		t.Fatalf("bound recipient paid before a valid submission: %d", got)
	}

	// The honest copy still verifies and pays only the bound recipient, even
	// when relayed by a different sender.
	msg.RewardRecipient = bound.String()
	if _, err := srv.RateWithProof(ctx, msg); err != nil {
		t.Fatalf("bound-recipient submission failed: %v", err)
	}
	if got := accountBalance(bank, bound); got != expectedReward(before.treasury) {
		t.Fatalf("bound recipient balance = %d, want %d", got, expectedReward(before.treasury))
	}
	if got := accountBalance(bank, attacker); got != 0 {
		t.Fatalf("relaying attacker received funds: %d", got)
	}
}

func TestRateProposalSignatureCopiedSubmissionCannotRedirect(t *testing.T) {
	k, ctx := setupKeeper(t)
	ctx = ctx.WithChainID("truerepublic-test-1")
	setupDomainWithIssue(t, k, ctx)
	aliceKey := domainKey("alice-copied-sig")
	if err := k.JoinPermissionRegister(ctx, "AnonDomain", "alice", aliceKey.PubKey().Bytes()); err != nil {
		t.Fatal(err)
	}
	bank := backExistingEscrow(&k, ctx)
	srv := NewMsgServer(k)

	bound := sdk.AccAddress("sig-bound-recipient")
	attacker := sdk.AccAddress("sig-attacker")
	payload := encodeVoteContextV2(ctx.ChainID(), "AnonDomain", "Climate", "GreenDeal", 2, bound.String())
	sig, err := aliceKey.Sign(payload)
	if err != nil {
		t.Fatal(err)
	}

	before := snapshotAccounting(t, k, ctx, bank, "AnonDomain")
	msg := &MsgRateProposal{
		Sender:          attacker,
		DomainName:      "AnonDomain",
		IssueName:       "Climate",
		SuggestionName:  "GreenDeal",
		Rating:          2,
		DomainPubKey:    hex.EncodeToString(aliceKey.PubKey().Bytes()),
		Signature:       hex.EncodeToString(sig),
		RewardRecipient: attacker.String(),
	}
	if _, err := srv.RateProposal(ctx, msg); err == nil {
		t.Fatal("recipient-substituted signature submission accepted")
	}
	assertUnchangedAfterFailure(t, k, ctx, bank, "AnonDomain", "", before, attacker)

	// A copied verbatim submission still pays only the bound recipient.
	msg.RewardRecipient = bound.String()
	if _, err := srv.RateProposal(ctx, msg); err != nil {
		t.Fatalf("bound-recipient signature submission failed: %v", err)
	}
	if got := accountBalance(bank, bound); got != expectedReward(before.treasury) {
		t.Fatalf("bound recipient balance = %d, want %d", got, expectedReward(before.treasury))
	}
	if got := accountBalance(bank, attacker); got != 0 {
		t.Fatalf("copying attacker received funds: %d", got)
	}
}

func TestRateWithProofModuleAccountRecipientRejected(t *testing.T) {
	k, ctx := setupKeeper(t)
	setupDomainWithZKPIdentity(t, k, ctx, "ZKPDomain", 2)
	addProposal(t, k, ctx, "ZKPDomain", "Climate", "GreenDeal")
	bank := backExistingEscrow(&k, ctx)
	bank.blockModuleAccount(ModuleName)
	srv := NewMsgServer(k)

	moduleRecipient := authtypes.NewModuleAddress(ModuleName)
	before := snapshotAccounting(t, k, ctx, bank, "ZKPDomain")
	msg := &MsgRateWithProof{
		Sender:          sdk.AccAddress("sender1"),
		DomainName:      "ZKPDomain",
		IssueName:       "Climate",
		SuggestionName:  "GreenDeal",
		Rating:          3,
		Proof:           hex.EncodeToString(make([]byte, 128)),
		NullifierHash:   hex.EncodeToString(make([]byte, 32)),
		RewardRecipient: moduleRecipient.String(),
	}
	if _, err := srv.RateWithProof(ctx, msg); err == nil || !strings.Contains(err.Error(), "module account") {
		t.Fatalf("module-account recipient result = %v", err)
	}
	assertUnchangedAfterFailure(t, k, ctx, bank, "ZKPDomain", "", before, moduleRecipient)
}

func TestRateWithProofBankSendFailureRollsBack(t *testing.T) {
	k, ctx := setupKeeper(t)
	secrets := setupDomainWithZKPIdentity(t, k, ctx, "ZKPDomain", 2)
	addProposal(t, k, ctx, "ZKPDomain", "Climate", "GreenDeal")
	bank := backExistingEscrow(&k, ctx)
	srv := NewMsgServer(k)

	recipient := sdk.AccAddress("bound-recipient")
	proofHex, nullifierHex := generateZKPRatingForRecipient(t, k, ctx, "ZKPDomain", secrets, 0, "Climate", "GreenDeal", 3, recipient.String())
	before := snapshotAccounting(t, k, ctx, bank, "ZKPDomain")

	bank.failModuleToAccount = true
	msg := &MsgRateWithProof{
		Sender:          sdk.AccAddress("sender1"),
		DomainName:      "ZKPDomain",
		IssueName:       "Climate",
		SuggestionName:  "GreenDeal",
		Rating:          3,
		Proof:           proofHex,
		NullifierHash:   nullifierHex,
		RewardRecipient: recipient.String(),
	}
	if _, err := srv.RateWithProof(ctx, msg); err == nil {
		t.Fatal("bank-send failure accepted")
	}
	assertUnchangedAfterFailure(t, k, ctx, bank, "ZKPDomain", nullifierHex, before, recipient)

	// Once the bank failure clears, the same proof still verifies: nothing was
	// consumed by the rolled-back attempt.
	bank.failModuleToAccount = false
	if _, err := srv.RateWithProof(ctx, msg); err != nil {
		t.Fatalf("retry after cleared bank failure rejected: %v", err)
	}
	if got := accountBalance(bank, recipient); got != expectedReward(before.treasury) {
		t.Fatalf("bound recipient balance = %d, want %d", got, expectedReward(before.treasury))
	}
}

func TestRateProposalSignatureBankSendFailureRollsBack(t *testing.T) {
	k, ctx := setupKeeper(t)
	ctx = ctx.WithChainID("truerepublic-test-1")
	setupDomainWithIssue(t, k, ctx)
	aliceKey := domainKey("alice-rollback-sig")
	if err := k.JoinPermissionRegister(ctx, "AnonDomain", "alice", aliceKey.PubKey().Bytes()); err != nil {
		t.Fatal(err)
	}
	bank := backExistingEscrow(&k, ctx)
	srv := NewMsgServer(k)

	recipient := sdk.AccAddress("sig-recipient")
	payload := encodeVoteContextV2(ctx.ChainID(), "AnonDomain", "Climate", "GreenDeal", 1, recipient.String())
	sig, err := aliceKey.Sign(payload)
	if err != nil {
		t.Fatal(err)
	}
	before := snapshotAccounting(t, k, ctx, bank, "AnonDomain")

	bank.failModuleToAccount = true
	msg := &MsgRateProposal{
		Sender:          sdk.AccAddress("sender1"),
		DomainName:      "AnonDomain",
		IssueName:       "Climate",
		SuggestionName:  "GreenDeal",
		Rating:          1,
		DomainPubKey:    hex.EncodeToString(aliceKey.PubKey().Bytes()),
		Signature:       hex.EncodeToString(sig),
		RewardRecipient: recipient.String(),
	}
	if _, err := srv.RateProposal(ctx, msg); err == nil {
		t.Fatal("bank-send failure accepted")
	}
	assertUnchangedAfterFailure(t, k, ctx, bank, "AnonDomain", "", before, recipient)

	bank.failModuleToAccount = false
	if _, err := srv.RateProposal(ctx, msg); err != nil {
		t.Fatalf("retry after cleared bank failure rejected: %v", err)
	}
	if got := accountBalance(bank, recipient); got != expectedReward(before.treasury) {
		t.Fatalf("bound recipient balance = %d, want %d", got, expectedReward(before.treasury))
	}
}

func TestRateWithProofReplayAndMalleationCannotDuplicatePayout(t *testing.T) {
	k, ctx := setupKeeper(t)
	secrets := setupDomainWithZKPIdentity(t, k, ctx, "ZKPDomain", 2)
	addProposal(t, k, ctx, "ZKPDomain", "Climate", "GreenDeal")
	bank := backExistingEscrow(&k, ctx)
	srv := NewMsgServer(k)

	recipient := sdk.AccAddress("bound-recipient")
	proofHex, nullifierHex := generateZKPRatingForRecipient(t, k, ctx, "ZKPDomain", secrets, 0, "Climate", "GreenDeal", 3, recipient.String())
	msg := &MsgRateWithProof{
		Sender:          sdk.AccAddress("sender1"),
		DomainName:      "ZKPDomain",
		IssueName:       "Climate",
		SuggestionName:  "GreenDeal",
		Rating:          3,
		Proof:           proofHex,
		NullifierHash:   nullifierHex,
		RewardRecipient: recipient.String(),
	}
	if _, err := srv.RateWithProof(ctx, msg); err != nil {
		t.Fatalf("first rating failed: %v", err)
	}
	paidBalance := accountBalance(bank, recipient)
	if paidBalance <= 0 {
		t.Fatal("expected a positive first payout")
	}
	after := snapshotAccounting(t, k, ctx, bank, "ZKPDomain")

	// An exact replay is blocked by the consumed nullifier and cannot pay twice.
	if _, err := srv.RateWithProof(ctx, msg); err == nil {
		t.Fatal("replayed proof accepted")
	}
	// A malleated proof fails verification and cannot pay either.
	malleated, err := hex.DecodeString(proofHex)
	if err != nil {
		t.Fatal(err)
	}
	malleated[len(malleated)/2] ^= 1
	malleatedMsg := *msg
	malleatedMsg.Proof = hex.EncodeToString(malleated)
	if _, err := srv.RateWithProof(ctx, &malleatedMsg); err == nil {
		t.Fatal("malleated proof accepted")
	}

	if got := accountBalance(bank, recipient); got != paidBalance {
		t.Fatalf("duplicate attempts changed recipient balance: %d, want %d", got, paidBalance)
	}
	current := snapshotAccounting(t, k, ctx, bank, "ZKPDomain")
	if current != after {
		t.Fatalf("duplicate attempts changed accounting: %+v, want %+v", current, after)
	}
	if err := k.ValidateEscrowParity(ctx); err != nil {
		t.Fatalf("escrow parity after duplicate attempts: %v", err)
	}
}

func TestRateWithProofV1SignalRejected(t *testing.T) {
	k, ctx := setupKeeper(t)
	secrets := setupDomainWithZKPIdentity(t, k, ctx, "ZKPDomain", 2)
	addProposal(t, k, ctx, "ZKPDomain", "Climate", "GreenDeal")
	bank := backExistingEscrow(&k, ctx)
	srv := NewMsgServer(k)

	// Build a proof over the historical recipient-independent v1 signal.
	keys := getTestZKPKeys(t)
	domain, _ := k.GetDomain(ctx, "ZKPDomain")
	commitments := make([][]byte, len(domain.IdentityCommits))
	for i, h := range domain.IdentityCommits {
		b, _ := hex.DecodeString(h)
		commitments[i] = b
	}
	tree := NewMerkleTree(MerkleTreeDepth)
	if err := tree.BuildFromLeaves(commitments); err != nil {
		t.Fatal(err)
	}
	siblings, pathIndices, err := tree.GenerateProof(0)
	if err != nil {
		t.Fatal(err)
	}
	extNullifier := ComputeVoteNullifierScope(ctx.ChainID(), "ZKPDomain", "Climate", "GreenDeal")
	v1Signal := ComputeVoteSignal(ctx.ChainID(), "ZKPDomain", "Climate", "GreenDeal", 3)
	proofBytes, nullifierHash, err := GenerateMembershipProofForSignal(keys, secrets[0], tree.Root, siblings, pathIndices, extNullifier, v1Signal)
	if err != nil {
		t.Fatal(err)
	}

	recipient := sdk.AccAddress("bound-recipient")
	before := snapshotAccounting(t, k, ctx, bank, "ZKPDomain")
	msg := &MsgRateWithProof{
		Sender:          sdk.AccAddress("sender1"),
		DomainName:      "ZKPDomain",
		IssueName:       "Climate",
		SuggestionName:  "GreenDeal",
		Rating:          3,
		Proof:           hex.EncodeToString(proofBytes),
		NullifierHash:   hex.EncodeToString(nullifierHash),
		RewardRecipient: recipient.String(),
	}
	if _, err := srv.RateWithProof(ctx, msg); err == nil {
		t.Fatal("v1-signal proof accepted after GH-209")
	}
	assertUnchangedAfterFailure(t, k, ctx, bank, "ZKPDomain", hex.EncodeToString(nullifierHash), before, recipient)
}

func TestRateProposalSignatureV1PayloadRejected(t *testing.T) {
	k, ctx := setupKeeper(t)
	ctx = ctx.WithChainID("truerepublic-test-1")
	setupDomainWithIssue(t, k, ctx)
	aliceKey := domainKey("alice-v1-sig")
	if err := k.JoinPermissionRegister(ctx, "AnonDomain", "alice", aliceKey.PubKey().Bytes()); err != nil {
		t.Fatal(err)
	}
	bank := backExistingEscrow(&k, ctx)
	srv := NewMsgServer(k)

	recipient := sdk.AccAddress("sig-recipient")
	rating := 3
	v1Payload := encodeVoteContext(ctx.ChainID(), "AnonDomain", "Climate", "GreenDeal", &rating)
	sig, err := aliceKey.Sign(v1Payload)
	if err != nil {
		t.Fatal(err)
	}
	before := snapshotAccounting(t, k, ctx, bank, "AnonDomain")
	msg := &MsgRateProposal{
		Sender:          sdk.AccAddress("sender1"),
		DomainName:      "AnonDomain",
		IssueName:       "Climate",
		SuggestionName:  "GreenDeal",
		Rating:          3,
		DomainPubKey:    hex.EncodeToString(aliceKey.PubKey().Bytes()),
		Signature:       hex.EncodeToString(sig),
		RewardRecipient: recipient.String(),
	}
	if _, err := srv.RateProposal(ctx, msg); err == nil {
		t.Fatal("v1-payload signature accepted after GH-209")
	}
	assertUnchangedAfterFailure(t, k, ctx, bank, "AnonDomain", "", before, recipient)
}

func TestRateWithProofWrongSuggestionRejected(t *testing.T) {
	k, ctx := setupKeeper(t)
	secrets := setupDomainWithZKPIdentity(t, k, ctx, "ZKPDomain", 2)
	addProposal(t, k, ctx, "ZKPDomain", "Climate", "GreenDeal")
	addProposal(t, k, ctx, "ZKPDomain", "Climate", "BlueDeal")
	bank := backExistingEscrow(&k, ctx)
	srv := NewMsgServer(k)

	recipient := sdk.AccAddress("bound-recipient")
	proofHex, nullifierHex := generateZKPRatingForRecipient(t, k, ctx, "ZKPDomain", secrets, 0, "Climate", "GreenDeal", 3, recipient.String())
	before := snapshotAccounting(t, k, ctx, bank, "ZKPDomain")
	msg := &MsgRateWithProof{
		Sender:          sdk.AccAddress("sender1"),
		DomainName:      "ZKPDomain",
		IssueName:       "Climate",
		SuggestionName:  "BlueDeal",
		Rating:          3,
		Proof:           proofHex,
		NullifierHash:   nullifierHex,
		RewardRecipient: recipient.String(),
	}
	if _, err := srv.RateWithProof(ctx, msg); err == nil {
		t.Fatal("proof replayed against a different suggestion")
	}

	domain, _ := k.GetDomain(ctx, "ZKPDomain")
	if got := domain.Treasury.AmountOf(PNYXDenom).Int64(); got != before.treasury {
		t.Fatalf("treasury mutated on failure: %d, want %d", got, before.treasury)
	}
	if k.IsNullifierUsed(ctx, "ZKPDomain", nullifierHex) {
		t.Fatal("wrong-suggestion failure consumed the nullifier")
	}
}

func TestAnonymousRewardRecipientExportImportRoundTrip(t *testing.T) {
	am1, k1, ctx1 := setupModuleForGenesis(t)
	secrets := setupDomainWithZKPIdentity(t, k1, ctx1, "ZKPDomain", 2)
	addProposal(t, k1, ctx1, "ZKPDomain", "Climate", "GreenDeal")
	backExistingEscrow(&k1, ctx1)

	recipient := sdk.AccAddress("roundtrip-recipient")
	proofHex, nullifierHex := generateZKPRatingForRecipient(t, k1, ctx1, "ZKPDomain", secrets, 0, "Climate", "GreenDeal", 3, recipient.String())
	if _, err := k1.RateProposalWithZKPPayout(ctx1, "ZKPDomain", "Climate", "GreenDeal", 3, proofHex, nullifierHex, "", recipient.String()); err != nil {
		t.Fatalf("payout failed: %v", err)
	}
	domainBefore, _ := k1.GetDomain(ctx1, "ZKPDomain")

	exported := am1.ExportGenesis(ctx1, nil)
	if strings.Contains(string(exported), recipient.String()) {
		t.Fatal("exported genesis persists the anonymous reward recipient")
	}

	am2, k2, ctx2 := setupModuleForGenesis(t)
	am2.InitGenesis(ctx2, nil, exported)

	if !k2.IsNullifierUsed(ctx2, "ZKPDomain", nullifierHex) {
		t.Fatal("nullifier not consumed after export/import")
	}
	domainAfter, found := k2.GetDomain(ctx2, "ZKPDomain")
	if !found {
		t.Fatal("domain missing after import")
	}
	if !domainAfter.Treasury.Equal(domainBefore.Treasury) || domainAfter.TotalPayouts != domainBefore.TotalPayouts {
		t.Fatalf("accounting drift after round-trip: treasury %s vs %s, payouts %d vs %d",
			domainAfter.Treasury, domainBefore.Treasury, domainAfter.TotalPayouts, domainBefore.TotalPayouts)
	}

	// A replay after import is rejected by the persisted nullifier: no
	// deferred reward can resurrect and no second payout can occur.
	if _, err := k2.RateProposalWithZKP(ctx2, "ZKPDomain", "Climate", "GreenDeal", 3, proofHex, nullifierHex, "", recipient.String()); err == nil {
		t.Fatal("replayed proof accepted after export/import")
	}
	domainReplay, _ := k2.GetDomain(ctx2, "ZKPDomain")
	if !domainReplay.Treasury.Equal(domainBefore.Treasury) || domainReplay.TotalPayouts != domainBefore.TotalPayouts {
		t.Fatal("rejected replay mutated accounting after import")
	}
	if got := len(domainReplay.Issues[0].Suggestions[0].Ratings); got != 1 {
		t.Fatalf("ratings after round-trip = %d, want exactly 1", got)
	}
}
