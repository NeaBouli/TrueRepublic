package truedemocracy

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"truerepublic/token"
	rewards "truerepublic/treasury/keeper"
)

func initializeRewardTimers(keeper Keeper, ctx sdk.Context) {
	store := ctx.KVStore(keeper.StoreKey)
	blockTime := ctx.BlockTime().Unix()
	store.Set([]byte("pod:last-reward-time"), keeper.cdc.MustMarshalLengthPrefixed(blockTime))
	store.Set([]byte("dom:last-interest-time"), keeper.cdc.MustMarshalLengthPrefixed(blockTime))
	keeper.IterateDomains(ctx, func(domain Domain) bool {
		store.Set(domainPayoutSnapshotKey(domain.Name), keeper.cdc.MustMarshalLengthPrefixed(domain.TotalPayouts))
		return false
	})
}

func TestEndBlockAggregateIssuanceStopsExactlyAtCap(t *testing.T) {
	originalNodeAPY := rewards.ApyNode
	rewards.ApyNode = math.LegacyNewDec(1_000_000_000)
	defer func() { rewards.ApyNode = originalNodeAPY }()

	keeper, ctx := setupKeeper(t)
	setupDomainWithValidator(t, keeper, ctx)
	initializeRewardTimers(keeper, ctx)
	domain, _ := keeper.GetDomain(ctx, "TestDomain")
	domain.TotalPayouts = 100_000 * PNYXUnit
	saveDomain(t, keeper, ctx, domain)

	bank := backExistingEscrow(&keeper, ctx)
	currentSupply := bank.GetSupply(ctx, PNYXDenom).Amount
	remainingBeforeFinalUnit := token.MaxSupply().Sub(currentSupply).SubRaw(1)
	bank.fundAccount(
		sdk.AccAddress("cap-holder"),
		sdk.NewCoins(sdk.NewCoin(PNYXDenom, remainingBeforeFinalUnit)),
	)

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Duration(RewardInterval) * time.Second))
	module := NewAppModule(keeper.cdc, keeper)
	if _, err := module.EndBlock(ctx); err != nil {
		t.Fatalf("end block near cap: %v", err)
	}

	if supply := bank.GetSupply(ctx, PNYXDenom).Amount; !supply.Equal(token.MaxSupply()) {
		t.Fatalf("supply after aggregate rewards = %s, want %s", supply, token.MaxSupply())
	}
	validator, _ := keeper.GetValidator(ctx, "oper1")
	if got := validator.Stake.AmountOf(PNYXDenom); !got.Equal(math.NewInt(rewards.StakeMin + 1)) {
		t.Fatalf("validator received %s stake, want final cap unit", got)
	}
	domain, _ = keeper.GetDomain(ctx, "TestDomain")
	if got := domain.Treasury.AmountOf(PNYXDenom); !got.Equal(math.NewInt(500_000 * PNYXUnit)) {
		t.Fatalf("domain minted beyond exhausted cap: treasury=%s", got)
	}
	if err := keeper.ValidateEscrowParity(ctx); err != nil {
		t.Fatalf("aggregate issuance broke escrow parity: %v", err)
	}
}

func TestStakingMintFailureRollsBackClaimsAndTimer(t *testing.T) {
	keeper, ctx := setupKeeper(t)
	setupDomainWithValidator(t, keeper, ctx)
	initializeRewardTimers(keeper, ctx)
	bank := backExistingEscrow(&keeper, ctx)
	bank.failMint = true
	supplyBefore := bank.GetSupply(ctx, PNYXDenom).Amount
	initialTime := ctx.BlockTime().Unix()

	rewardCtx := ctx.WithBlockTime(ctx.BlockTime().Add(time.Duration(RewardInterval) * time.Second))
	if err := keeper.DistributeStakingRewards(rewardCtx); err == nil {
		t.Fatal("expected injected mint failure")
	}
	validator, _ := keeper.GetValidator(ctx, "oper1")
	if got := validator.Stake.AmountOf(PNYXDenom); !got.Equal(math.NewInt(rewards.StakeMin)) {
		t.Fatalf("failed mint changed stake claim: %s", got)
	}
	var storedTime int64
	store := ctx.KVStore(keeper.StoreKey)
	keeper.cdc.MustUnmarshalLengthPrefixed(store.Get([]byte("pod:last-reward-time")), &storedTime)
	if storedTime != initialTime {
		t.Fatalf("failed mint advanced reward timer: %d", storedTime)
	}
	if supply := bank.GetSupply(ctx, PNYXDenom).Amount; !supply.Equal(supplyBefore) {
		t.Fatalf("failed mint changed supply: %s", supply)
	}
	if err := keeper.ValidateEscrowParity(ctx); err != nil {
		t.Fatalf("failed mint broke parity: %v", err)
	}
}

func TestEndBlockDomainMintFailureRollsBackRewardClaimsAndTimers(t *testing.T) {
	keeper, ctx := setupKeeper(t)
	setupDomainWithValidator(t, keeper, ctx)
	initializeRewardTimers(keeper, ctx)
	domain, _ := keeper.GetDomain(ctx, "TestDomain")
	domain.TotalPayouts = 100_000 * PNYXUnit
	saveDomain(t, keeper, ctx, domain)
	bank := backExistingEscrow(&keeper, ctx)
	bank.failMintAt = 2
	supplyBefore := bank.GetSupply(ctx, PNYXDenom).Amount
	moduleBefore := bank.GetBalance(ctx, authtypes.NewModuleAddress(ModuleName), PNYXDenom).Amount
	initialTime := ctx.BlockTime().Unix()

	rewardCtx := ctx.WithBlockTime(ctx.BlockTime().Add(time.Duration(RewardInterval) * time.Second))
	module := NewAppModule(keeper.cdc, keeper)
	if _, err := module.EndBlock(rewardCtx); err == nil {
		t.Fatal("expected second issuance call to fail")
	}

	validator, _ := keeper.GetValidator(ctx, "oper1")
	if got := validator.Stake.AmountOf(PNYXDenom); !got.Equal(math.NewInt(rewards.StakeMin)) {
		t.Fatalf("failed reward phase changed validator claim: %s", got)
	}
	afterDomain, _ := keeper.GetDomain(ctx, "TestDomain")
	if !afterDomain.Treasury.Equal(domain.Treasury) {
		t.Fatalf("failed reward phase changed domain claim: %s", afterDomain.Treasury)
	}

	store := ctx.KVStore(keeper.StoreKey)
	for _, key := range [][]byte{[]byte("pod:last-reward-time"), []byte("dom:last-interest-time")} {
		var storedTime int64
		keeper.cdc.MustUnmarshalLengthPrefixed(store.Get(key), &storedTime)
		if storedTime != initialTime {
			t.Fatalf("failed reward phase advanced %s to %d", key, storedTime)
		}
	}
	if store.Has(domainPayoutSnapshotKey("TestDomain")) {
		var payoutSnapshot int64
		keeper.cdc.MustUnmarshalLengthPrefixed(store.Get(domainPayoutSnapshotKey("TestDomain")), &payoutSnapshot)
		if payoutSnapshot != 0 {
			t.Fatalf("failed reward phase changed domain payout snapshot: %d", payoutSnapshot)
		}
	}
	if supply := bank.GetSupply(ctx, PNYXDenom).Amount; !supply.Equal(supplyBefore) {
		t.Fatalf("failed reward phase changed canonical supply: %s", supply)
	}
	if moduleBalance := bank.GetBalance(ctx, authtypes.NewModuleAddress(ModuleName), PNYXDenom).Amount; !moduleBalance.Equal(moduleBefore) {
		t.Fatalf("failed reward phase changed module escrow: %s", moduleBalance)
	}
	if err := keeper.ValidateEscrowParity(ctx); err != nil {
		t.Fatalf("failed reward phase broke escrow parity: %v", err)
	}
}

func TestDomainInterestUsesOnlyNewIntervalPayouts(t *testing.T) {
	keeper, ctx := setupKeeper(t)
	keeper.CreateDomain(ctx, "Interval", sdk.AccAddress("admin"), sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 500_000*PNYXUnit)))
	initializeRewardTimers(keeper, ctx)
	domain, _ := keeper.GetDomain(ctx, "Interval")
	domain.TotalPayouts = 100_000 * PNYXUnit
	saveDomain(t, keeper, ctx, domain)
	bank := backExistingEscrow(&keeper, ctx)

	firstCtx := ctx.WithBlockTime(ctx.BlockTime().Add(time.Duration(RewardInterval) * time.Second))
	if err := keeper.DistributeDomainInterest(firstCtx); err != nil {
		t.Fatal(err)
	}
	afterFirst, _ := keeper.GetDomain(firstCtx, "Interval")
	if !afterFirst.Treasury.AmountOf(PNYXDenom).GT(domain.Treasury.AmountOf(PNYXDenom)) {
		t.Fatal("expected first interval interest")
	}
	supplyAfterFirst := bank.GetSupply(firstCtx, PNYXDenom).Amount

	secondCtx := firstCtx.WithBlockTime(firstCtx.BlockTime().Add(time.Duration(RewardInterval) * time.Second))
	if err := keeper.DistributeDomainInterest(secondCtx); err != nil {
		t.Fatal(err)
	}
	afterSecond, _ := keeper.GetDomain(secondCtx, "Interval")
	if !afterSecond.Treasury.AmountOf(PNYXDenom).Equal(afterFirst.Treasury.AmountOf(PNYXDenom)) {
		t.Fatalf("cumulative payout was rewarded twice: first=%s second=%s", afterFirst.Treasury, afterSecond.Treasury)
	}
	if supply := bank.GetSupply(secondCtx, PNYXDenom).Amount; !supply.Equal(supplyAfterFirst) {
		t.Fatalf("idle interval changed canonical supply: %s", supply)
	}
	if err := keeper.ValidateEscrowParity(secondCtx); err != nil {
		t.Fatalf("interval issuance broke parity: %v", err)
	}
}

func TestDomainInterestBackfillsMissingSnapshotWithoutHistoricalReward(t *testing.T) {
	keeper, ctx := setupKeeper(t)
	keeper.CreateDomain(ctx, "Restored", sdk.AccAddress("admin"), sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 500_000*PNYXUnit)))
	domain, _ := keeper.GetDomain(ctx, "Restored")
	domain.TotalPayouts = 100_000 * PNYXUnit
	saveDomain(t, keeper, ctx, domain)

	store := ctx.KVStore(keeper.StoreKey)
	store.Set([]byte("dom:last-interest-time"), keeper.cdc.MustMarshalLengthPrefixed(ctx.BlockTime().Unix()))
	bank := backExistingEscrow(&keeper, ctx)
	supplyBefore := bank.GetSupply(ctx, PNYXDenom).Amount

	rewardCtx := ctx.WithBlockTime(ctx.BlockTime().Add(time.Duration(RewardInterval) * time.Second))
	if err := keeper.DistributeDomainInterest(rewardCtx); err != nil {
		t.Fatal(err)
	}
	after, _ := keeper.GetDomain(ctx, "Restored")
	if !after.Treasury.Equal(domain.Treasury) {
		t.Fatalf("historical payouts earned interest: before=%s after=%s", domain.Treasury, after.Treasury)
	}
	if supply := bank.GetSupply(ctx, PNYXDenom).Amount; !supply.Equal(supplyBefore) {
		t.Fatalf("historical payout backfill changed supply: %s", supply)
	}
	var snapshot int64
	keeper.cdc.MustUnmarshalLengthPrefixed(store.Get(domainPayoutSnapshotKey("Restored")), &snapshot)
	if snapshot != domain.TotalPayouts {
		t.Fatalf("snapshot = %d, want %d", snapshot, domain.TotalPayouts)
	}
}

func TestVoteRewardTransfersEscrowWithoutChangingSupply(t *testing.T) {
	keeper, ctx, bank := setupKeeperWithBank(t)
	member := sdk.AccAddress("reward-member")
	initial := int64(1_000_000 * PNYXUnit)
	fee := rewards.CalcPutPrice(math.NewInt(initial), 1).Int64()
	bank.fundAccount(member, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, initial+fee)))
	if err := keeper.CreateDomainWithEscrow(ctx, "Rewards", member, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, initial))); err != nil {
		t.Fatal(err)
	}
	if err := keeper.SubmitProposalWithEscrow(
		ctx,
		member,
		member.String(),
		"Rewards",
		"Issue",
		"Suggestion",
		sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, fee)),
		"",
	); err != nil {
		t.Fatal(err)
	}

	supplyBefore := bank.GetSupply(ctx, PNYXDenom).Amount
	accountBefore := accountBalance(bank, member)
	reward, err := keeper.PlaceStoneOnIssueWithPayout(ctx, member, "Rewards", "Issue", member.String())
	if err != nil {
		t.Fatal(err)
	}
	if !reward.AmountOf(PNYXDenom).IsPositive() {
		t.Fatal("expected bank-paid vote reward")
	}
	if got := accountBalance(bank, member) - accountBefore; got != reward.AmountOf(PNYXDenom).Int64() {
		t.Fatalf("account reward = %d, want %s", got, reward.AmountOf(PNYXDenom))
	}
	if supply := bank.GetSupply(ctx, PNYXDenom).Amount; !supply.Equal(supplyBefore) {
		t.Fatalf("treasury-funded vote reward changed supply: %s", supply)
	}
	if err := keeper.ValidateEscrowParity(ctx); err != nil {
		t.Fatalf("vote reward broke escrow parity: %v", err)
	}
}
