package dex

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// requireDexMsgEvent fails the test unless the context event manager recorded
// an event of the given type.
func requireDexMsgEvent(t *testing.T, ctx sdk.Context, eventType string) {
	t.Helper()
	for _, event := range ctx.EventManager().Events() {
		if event.Type == eventType {
			return
		}
	}
	t.Fatalf("event %q was not emitted; got %v", eventType, ctx.EventManager().Events())
}

func countDexMsgEvents(ctx sdk.Context, eventType string) int {
	count := 0
	for _, event := range ctx.EventManager().Events() {
		if event.Type == eventType {
			count++
		}
	}
	return count
}

// TestMsgServerSwapFailsClosed proves the legacy slippage-free swap entry
// point always fails and never touches pool state or trader funds.
func TestMsgServerSwapFailsClosed(t *testing.T) {
	keeper, ctx, bank, _ := setupCustodyKeeper(t)
	provider := sdk.AccAddress("legacy-swap-provider")
	bank.fundAccount(ctx, provider, sdk.NewCoins(
		sdk.NewInt64Coin(pnyxDenom, 2_000_000),
		sdk.NewInt64Coin("atom", 2_000_000),
	))
	if err := keeper.CreatePoolWithCustody(ctx, provider, "atom", math.NewInt(1_000_000), math.NewInt(1_000_000)); err != nil {
		t.Fatal(err)
	}
	server := NewMsgServer(keeper)
	trader := sdk.AccAddress("legacy-swap-trader")
	bank.fundAccount(ctx, trader, sdk.NewCoins(sdk.NewInt64Coin(pnyxDenom, 100_000)))

	poolBefore, found := keeper.GetPool(ctx, "atom")
	if !found {
		t.Fatal("pool missing after custodied creation")
	}
	traderBefore := bank.balance(ctx, accountOwner(trader), pnyxDenom)

	if _, err := server.Swap(sdk.WrapSDKContext(ctx), &MsgSwap{
		Sender:      trader,
		InputDenom:  pnyxDenom,
		InputAmt:    10_000,
		OutputDenom: "atom",
	}); err == nil || !containsStr(err.Error(), "disabled") {
		t.Fatalf("legacy swap was not rejected as disabled: %v", err)
	}

	poolAfter, _ := keeper.GetPool(ctx, "atom")
	if !poolAfter.PnyxReserve.Equal(poolBefore.PnyxReserve) ||
		!poolAfter.AssetReserve.Equal(poolBefore.AssetReserve) ||
		poolAfter.SwapCount != poolBefore.SwapCount {
		t.Fatal("rejected legacy swap mutated pool state")
	}
	if !bank.balance(ctx, accountOwner(trader), pnyxDenom).Equal(traderBefore) {
		t.Fatal("rejected legacy swap moved trader funds")
	}
}

// TestMsgServerRegisterAssetAuthorityBoundary proves only the chain authority
// can register assets and that failures leave the registry untouched.
func TestMsgServerRegisterAssetAuthorityBoundary(t *testing.T) {
	keeper, ctx, _, authority := setupCustodyKeeper(t)
	ctx = ctx.WithEventManager(sdk.NewEventManager())
	server := NewMsgServer(keeper)
	goCtx := sdk.WrapSDKContext(ctx)

	registration := &MsgRegisterAsset{
		Sender:      sdk.AccAddress("registry-attacker"),
		IBCDenom:    "ibc/eth",
		Symbol:      "ETH",
		Name:        "Ether",
		Decimals:    18,
		OriginChain: "ethereum",
		IBCChannel:  "channel-0",
	}
	if _, err := server.RegisterAsset(goCtx, registration); err == nil {
		t.Fatal("unauthorized asset registration succeeded")
	}
	if _, found := keeper.GetAssetByDenom(ctx, "ibc/eth"); found {
		t.Fatal("unauthorized registration committed registry state")
	}

	registration.Sender = authority
	if _, err := server.RegisterAsset(goCtx, registration); err != nil {
		t.Fatalf("authority registration failed: %v", err)
	}
	asset, found := keeper.GetAssetByDenom(ctx, "ibc/eth")
	if !found {
		t.Fatal("authorized registration did not persist the asset")
	}
	if !asset.TradingEnabled ||
		asset.Symbol != "ETH" ||
		asset.Name != "Ether" ||
		asset.RegisteredBy != authority.String() ||
		asset.RegisteredHeight != ctx.BlockHeight() {
		t.Fatalf("registered asset mismatch: %+v", asset)
	}
	requireDexMsgEvent(t, ctx, "asset_registered")

	// A duplicate registration through the handler must fail without
	// overwriting the existing record.
	duplicate := *registration
	duplicate.Name = "Overwrite Attempt"
	if _, err := server.RegisterAsset(goCtx, &duplicate); err == nil {
		t.Fatal("duplicate asset registration succeeded")
	}
	asset, _ = keeper.GetAssetByDenom(ctx, "ibc/eth")
	if asset.Name != "Ether" {
		t.Fatalf("duplicate registration overwrote the asset: %+v", asset)
	}
}

// TestMsgServerUpdateAssetStatusAuthorityBoundary proves only the chain
// authority can toggle trading status and that unknown assets are rejected.
func TestMsgServerUpdateAssetStatusAuthorityBoundary(t *testing.T) {
	keeper, ctx, _, authority := setupCustodyKeeper(t)
	server := NewMsgServer(keeper)
	goCtx := sdk.WrapSDKContext(ctx)

	if _, err := server.UpdateAssetStatus(goCtx, &MsgUpdateAssetStatus{
		Sender:   sdk.AccAddress("status-attacker"),
		IBCDenom: "atom",
		Enabled:  false,
	}); err == nil {
		t.Fatal("unauthorized status update succeeded")
	}
	asset, found := keeper.GetAssetByDenom(ctx, "atom")
	if !found || !asset.TradingEnabled {
		t.Fatal("unauthorized status update mutated registry state")
	}

	if _, err := server.UpdateAssetStatus(goCtx, &MsgUpdateAssetStatus{
		Sender:   authority,
		IBCDenom: "ibc/missing",
		Enabled:  false,
	}); err == nil {
		t.Fatal("status update for an unregistered asset succeeded")
	}

	if _, err := server.UpdateAssetStatus(goCtx, &MsgUpdateAssetStatus{
		Sender:   authority,
		IBCDenom: "atom",
		Enabled:  false,
	}); err != nil {
		t.Fatalf("authority status update failed: %v", err)
	}
	asset, _ = keeper.GetAssetByDenom(ctx, "atom")
	if asset.TradingEnabled {
		t.Fatal("authority status update did not disable trading")
	}
	requireDexMsgEvent(t, ctx, "asset_trading_status_updated")

	if _, err := server.UpdateAssetStatus(goCtx, &MsgUpdateAssetStatus{
		Sender:   authority,
		IBCDenom: "atom",
		Enabled:  true,
	}); err != nil {
		t.Fatalf("authority status re-enable failed: %v", err)
	}
	asset, _ = keeper.GetAssetByDenom(ctx, "atom")
	if !asset.TradingEnabled {
		t.Fatal("authority status update did not re-enable trading")
	}
}

// TestMsgServerCreatePoolCustodyBoundary proves the custody-backed pool
// creation handler settles funds on success and commits nothing on failure.
func TestMsgServerCreatePoolCustodyBoundary(t *testing.T) {
	keeper, ctx, bank, _ := setupCustodyKeeper(t)
	server := NewMsgServer(keeper)
	goCtx := sdk.WrapSDKContext(ctx)
	provider := sdk.AccAddress("msg-pool-provider")

	// An unfunded provider must fail without committing pool or LP state.
	if _, err := server.CreatePool(goCtx, &MsgCreatePool{
		Sender:     provider,
		AssetDenom: "atom",
		PnyxAmt:    1_000_000,
		AssetAmt:   1_000_000,
	}); err == nil {
		t.Fatal("pool creation without funds succeeded")
	}
	if _, found := keeper.GetPool(ctx, "atom"); found {
		t.Fatal("failed pool creation committed pool state")
	}
	if !keeper.GetLPBalance(ctx, "atom", provider).IsZero() {
		t.Fatal("failed pool creation committed LP shares")
	}
	if count := countDexMsgEvents(ctx, "create_pool"); count != 0 {
		t.Fatalf("failed pool creation emitted %d create_pool events", count)
	}

	bank.fundAccount(ctx, provider, sdk.NewCoins(
		sdk.NewInt64Coin(pnyxDenom, 1_000_000),
		sdk.NewInt64Coin("atom", 1_000_000),
	))
	if _, err := server.CreatePool(goCtx, &MsgCreatePool{
		Sender:     provider,
		AssetDenom: "atom",
		PnyxAmt:    1_000_000,
		AssetAmt:   500_000,
	}); err != nil {
		t.Fatalf("custodied pool creation failed: %v", err)
	}
	pool, found := keeper.GetPool(ctx, "atom")
	if !found {
		t.Fatal("pool missing after custodied creation")
	}
	if !keeper.GetLPBalance(ctx, "atom", provider).Equal(pool.TotalShares) {
		t.Fatal("pool shares were not assigned to the provider")
	}
	requireDexMsgEvent(t, ctx, "create_pool")
	if err := keeper.validateCustodyAndShares(ctx); err != nil {
		t.Fatal(err)
	}
}

// TestMsgServerAddAndRemoveLiquidityCustody proves the custody-backed
// liquidity handlers mint/burn exactly the sender's shares, emit events on
// success, and roll back every state change on failure.
func TestMsgServerAddAndRemoveLiquidityCustody(t *testing.T) {
	keeper, ctx, bank, _ := setupCustodyKeeper(t)
	server := NewMsgServer(keeper)
	goCtx := sdk.WrapSDKContext(ctx)
	provider := sdk.AccAddress("msg-lp-provider")
	second := sdk.AccAddress("msg-lp-second")
	bank.fundAccount(ctx, provider, sdk.NewCoins(
		sdk.NewInt64Coin(pnyxDenom, 2_000_000),
		sdk.NewInt64Coin("atom", 2_000_000),
	))
	bank.fundAccount(ctx, second, sdk.NewCoins(
		sdk.NewInt64Coin(pnyxDenom, 100_000),
		sdk.NewInt64Coin("atom", 100_000),
	))
	if err := keeper.CreatePoolWithCustody(ctx, provider, "atom", math.NewInt(1_000_000), math.NewInt(1_000_000)); err != nil {
		t.Fatal(err)
	}

	// A deposit exceeding the provider's funds must fail without state changes.
	poolBefore, _ := keeper.GetPool(ctx, "atom")
	if _, err := server.AddLiquidity(goCtx, &MsgAddLiquidity{
		Sender:     second,
		AssetDenom: "atom",
		PnyxAmt:    200_000,
		AssetAmt:   200_000,
	}); err == nil {
		t.Fatal("liquidity deposit without funds succeeded")
	}
	poolAfter, _ := keeper.GetPool(ctx, "atom")
	if !poolAfter.PnyxReserve.Equal(poolBefore.PnyxReserve) ||
		!poolAfter.AssetReserve.Equal(poolBefore.AssetReserve) ||
		!poolAfter.TotalShares.Equal(poolBefore.TotalShares) {
		t.Fatal("failed liquidity deposit mutated pool state")
	}
	if !keeper.GetLPBalance(ctx, "atom", second).IsZero() {
		t.Fatal("failed liquidity deposit committed LP shares")
	}

	// A proportional deposit succeeds and mints shares to the sender.
	if _, err := server.AddLiquidity(goCtx, &MsgAddLiquidity{
		Sender:     second,
		AssetDenom: "atom",
		PnyxAmt:    100_000,
		AssetAmt:   100_000,
	}); err != nil {
		t.Fatalf("custodied liquidity deposit failed: %v", err)
	}
	shares := keeper.GetLPBalance(ctx, "atom", second)
	if !shares.IsPositive() {
		t.Fatal("successful deposit did not assign LP shares")
	}
	requireDexMsgEvent(t, ctx, "add_liquidity")

	// Removing more shares than owned must fail without state changes.
	if _, err := server.RemoveLiquidity(goCtx, &MsgRemoveLiquidity{
		Sender:     second,
		AssetDenom: "atom",
		Shares:     shares.Int64() + 1,
	}); err == nil {
		t.Fatal("removal of unowned shares succeeded")
	}
	if !keeper.GetLPBalance(ctx, "atom", second).Equal(shares) {
		t.Fatal("failed removal mutated LP ownership")
	}

	// Removing all owned shares succeeds and returns both assets.
	if _, err := server.RemoveLiquidity(goCtx, &MsgRemoveLiquidity{
		Sender:     second,
		AssetDenom: "atom",
		Shares:     shares.Int64(),
	}); err != nil {
		t.Fatalf("custodied liquidity removal failed: %v", err)
	}
	if !keeper.GetLPBalance(ctx, "atom", second).IsZero() {
		t.Fatal("full removal left LP shares behind")
	}
	requireDexMsgEvent(t, ctx, "remove_liquidity")
	if err := keeper.validateCustodyAndShares(ctx); err != nil {
		t.Fatal(err)
	}
}

// TestMsgServerSwapExactCustodyBoundary proves the slippage-protected swap
// handler settles the exact output on success and rejects unreachable
// min_output without moving funds, pool state, or events.
func TestMsgServerSwapExactCustodyBoundary(t *testing.T) {
	keeper, ctx, bank, _ := setupCustodyKeeper(t)
	server := NewMsgServer(keeper)
	goCtx := sdk.WrapSDKContext(ctx)
	provider := sdk.AccAddress("msg-swap-provider")
	trader := sdk.AccAddress("msg-swap-trader")
	bank.fundAccount(ctx, provider, sdk.NewCoins(
		sdk.NewInt64Coin(pnyxDenom, 2_000_000),
		sdk.NewInt64Coin("atom", 2_000_000),
	))
	bank.fundAccount(ctx, trader, sdk.NewCoins(sdk.NewInt64Coin(pnyxDenom, 100_000)))
	if err := keeper.CreatePoolWithCustody(ctx, provider, "atom", math.NewInt(1_000_000), math.NewInt(1_000_000)); err != nil {
		t.Fatal(err)
	}

	if _, err := server.SwapExact(goCtx, &MsgSwapExact{
		Sender:      trader,
		InputDenom:  pnyxDenom,
		InputAmt:    10_000,
		OutputDenom: "atom",
		MinOutput:   1,
	}); err != nil {
		t.Fatalf("custodied swap failed: %v", err)
	}
	if !bank.balance(ctx, accountOwner(trader), "atom").IsPositive() {
		t.Fatal("swap did not settle the asset output")
	}
	requireDexMsgEvent(t, ctx, "swap_exact")

	// An unreachable min_output must fail without moving funds or pool state.
	poolBefore, _ := keeper.GetPool(ctx, "atom")
	pnyxBefore := bank.balance(ctx, accountOwner(trader), pnyxDenom)
	eventsBefore := countDexMsgEvents(ctx, "swap_exact")
	if _, err := server.SwapExact(goCtx, &MsgSwapExact{
		Sender:      trader,
		InputDenom:  pnyxDenom,
		InputAmt:    10_000,
		OutputDenom: "atom",
		MinOutput:   999_999_999,
	}); err == nil || !containsStr(err.Error(), "slippage") {
		t.Fatalf("swap with unreachable min_output was not rejected: %v", err)
	}
	poolAfter, _ := keeper.GetPool(ctx, "atom")
	if !poolAfter.PnyxReserve.Equal(poolBefore.PnyxReserve) ||
		!poolAfter.AssetReserve.Equal(poolBefore.AssetReserve) ||
		poolAfter.SwapCount != poolBefore.SwapCount {
		t.Fatal("rejected swap mutated pool state")
	}
	if !bank.balance(ctx, accountOwner(trader), pnyxDenom).Equal(pnyxBefore) {
		t.Fatal("rejected swap moved trader funds")
	}
	if got := countDexMsgEvents(ctx, "swap_exact"); got != eventsBefore {
		t.Fatal("rejected swap emitted a swap_exact event")
	}
	if err := keeper.validateCustodyAndShares(ctx); err != nil {
		t.Fatal(err)
	}
}
