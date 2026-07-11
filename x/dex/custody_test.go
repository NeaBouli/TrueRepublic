package dex

import (
	"context"
	"fmt"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type storeBankKeeper struct {
	key                 *storetypes.KVStoreKey
	failAccountToModule bool
	failModuleToAccount bool
	failBurn            bool
}

func bankAmountKey(owner, denom string) []byte {
	return []byte("balance:" + owner + ":" + denom)
}

func bankSupplyKey(denom string) []byte {
	return []byte("supply:" + denom)
}

func readBankAmount(ctx sdk.Context, key *storetypes.KVStoreKey, storageKey []byte) math.Int {
	bz := ctx.KVStore(key).Get(storageKey)
	if bz == nil {
		return math.ZeroInt()
	}
	amount, ok := math.NewIntFromString(string(bz))
	if !ok {
		panic("invalid mock bank amount")
	}
	return amount
}

func writeBankAmount(ctx sdk.Context, key *storetypes.KVStoreKey, storageKey []byte, amount math.Int) {
	store := ctx.KVStore(key)
	if amount.IsZero() {
		store.Delete(storageKey)
		return
	}
	store.Set(storageKey, []byte(amount.String()))
}

func accountOwner(address sdk.AccAddress) string {
	return "account:" + address.String()
}

func moduleOwner(moduleName string) string {
	return "module:" + moduleName
}

func (bank *storeBankKeeper) balance(ctx sdk.Context, owner, denom string) math.Int {
	return readBankAmount(ctx, bank.key, bankAmountKey(owner, denom))
}

func (bank *storeBankKeeper) setBalance(ctx sdk.Context, owner, denom string, amount math.Int) {
	writeBankAmount(ctx, bank.key, bankAmountKey(owner, denom), amount)
}

func (bank *storeBankKeeper) fundAccount(ctx sdk.Context, address sdk.AccAddress, coins sdk.Coins) {
	for _, coin := range coins {
		bank.setBalance(ctx, accountOwner(address), coin.Denom, bank.balance(ctx, accountOwner(address), coin.Denom).Add(coin.Amount))
		supply := readBankAmount(ctx, bank.key, bankSupplyKey(coin.Denom))
		writeBankAmount(ctx, bank.key, bankSupplyKey(coin.Denom), supply.Add(coin.Amount))
	}
}

func (bank *storeBankKeeper) transfer(ctx context.Context, from, to string, coins sdk.Coins) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	for _, coin := range coins {
		if bank.balance(sdkCtx, from, coin.Denom).LT(coin.Amount) {
			return fmt.Errorf("insufficient %s funds", coin.Denom)
		}
	}
	for _, coin := range coins {
		bank.setBalance(sdkCtx, from, coin.Denom, bank.balance(sdkCtx, from, coin.Denom).Sub(coin.Amount))
		bank.setBalance(sdkCtx, to, coin.Denom, bank.balance(sdkCtx, to, coin.Denom).Add(coin.Amount))
	}
	return nil
}

func (bank *storeBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, sender sdk.AccAddress, moduleName string, coins sdk.Coins) error {
	if bank.failAccountToModule {
		return fmt.Errorf("injected account-to-module failure")
	}
	return bank.transfer(ctx, accountOwner(sender), moduleOwner(moduleName), coins)
}

func (bank *storeBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, moduleName string, recipient sdk.AccAddress, coins sdk.Coins) error {
	if bank.failModuleToAccount {
		return fmt.Errorf("injected module-to-account failure")
	}
	return bank.transfer(ctx, moduleOwner(moduleName), accountOwner(recipient), coins)
}

func (bank *storeBankKeeper) GetBalance(ctx context.Context, address sdk.AccAddress, denom string) sdk.Coin {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	owner := accountOwner(address)
	if address.Equals(authtypes.NewModuleAddress(ModuleName)) {
		owner = moduleOwner(ModuleName)
	}
	return sdk.NewCoin(denom, bank.balance(sdkCtx, owner, denom))
}

func (bank *storeBankKeeper) GetAllBalances(ctx context.Context, address sdk.AccAddress) sdk.Coins {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	owner := accountOwner(address)
	if address.Equals(authtypes.NewModuleAddress(ModuleName)) {
		owner = moduleOwner(ModuleName)
	}
	prefix := []byte("balance:" + owner + ":")
	iterator := sdkCtx.KVStore(bank.key).Iterator(prefix, prefixEnd(prefix))
	defer iterator.Close()
	balances := sdk.NewCoins()
	for ; iterator.Valid(); iterator.Next() {
		denom := string(iterator.Key()[len(prefix):])
		balances = balances.Add(sdk.NewCoin(denom, readBankAmount(sdkCtx, bank.key, iterator.Key())))
	}
	return balances
}

func (bank *storeBankKeeper) GetSupply(ctx context.Context, denom string) sdk.Coin {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdk.NewCoin(denom, readBankAmount(sdkCtx, bank.key, bankSupplyKey(denom)))
}

func (bank *storeBankKeeper) MintCoins(ctx context.Context, moduleName string, coins sdk.Coins) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	for _, coin := range coins {
		bank.setBalance(sdkCtx, moduleOwner(moduleName), coin.Denom, bank.balance(sdkCtx, moduleOwner(moduleName), coin.Denom).Add(coin.Amount))
		supply := readBankAmount(sdkCtx, bank.key, bankSupplyKey(coin.Denom))
		writeBankAmount(sdkCtx, bank.key, bankSupplyKey(coin.Denom), supply.Add(coin.Amount))
	}
	return nil
}

func (bank *storeBankKeeper) BurnCoins(ctx context.Context, moduleName string, coins sdk.Coins) error {
	if bank.failBurn {
		return fmt.Errorf("injected burn failure")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	for _, coin := range coins {
		balance := bank.balance(sdkCtx, moduleOwner(moduleName), coin.Denom)
		if balance.LT(coin.Amount) {
			return fmt.Errorf("insufficient module funds to burn")
		}
		bank.setBalance(sdkCtx, moduleOwner(moduleName), coin.Denom, balance.Sub(coin.Amount))
		supply := readBankAmount(sdkCtx, bank.key, bankSupplyKey(coin.Denom))
		writeBankAmount(sdkCtx, bank.key, bankSupplyKey(coin.Denom), supply.Sub(coin.Amount))
	}
	return nil
}

func setupCustodyKeeper(t *testing.T) (Keeper, sdk.Context, *storeBankKeeper, sdk.AccAddress) {
	t.Helper()
	dexKey := storetypes.NewKVStoreKey(ModuleName)
	bankKey := storetypes.NewKVStoreKey("mock-bank")
	database := dbm.NewMemDB()
	multiStore := store.NewCommitMultiStore(database, log.NewNopLogger(), metrics.NewNoOpMetrics())
	multiStore.MountStoreWithDB(dexKey, storetypes.StoreTypeIAVL, nil)
	multiStore.MountStoreWithDB(bankKey, storetypes.StoreTypeIAVL, nil)
	if err := multiStore.LoadLatestVersion(); err != nil {
		t.Fatal(err)
	}
	codex := codec.NewLegacyAmino()
	RegisterCodec(codex)
	authority := sdk.AccAddress("dex-authority")
	bank := &storeBankKeeper{key: bankKey}
	keeper := NewKeeper(codex, dexKey, bank, authority.String())
	ctx := sdk.NewContext(multiStore, cmtproto.Header{}, false, log.NewNopLogger())
	if err := keeper.RegisterAsset(ctx, RegisteredAsset{IBCDenom: "atom", Symbol: "ATOM", Decimals: 6, TradingEnabled: true}); err != nil {
		t.Fatal(err)
	}
	if err := keeper.RegisterAsset(ctx, RegisteredAsset{IBCDenom: "btc", Symbol: "BTC", Decimals: 8, TradingEnabled: true}); err != nil {
		t.Fatal(err)
	}
	return keeper, ctx, bank, authority
}

func TestCustodyLiquidityLifecycleAndOwnership(t *testing.T) {
	keeper, ctx, bank, _ := setupCustodyKeeper(t)
	provider := sdk.AccAddress("provider-one")
	second := sdk.AccAddress("provider-two")
	attacker := sdk.AccAddress("lp-attacker")
	bank.fundAccount(ctx, provider, sdk.NewCoins(sdk.NewInt64Coin(pnyxDenom, 2_000_000), sdk.NewInt64Coin("atom", 2_000_000)))
	bank.fundAccount(ctx, second, sdk.NewCoins(sdk.NewInt64Coin(pnyxDenom, 200_000), sdk.NewInt64Coin("atom", 200_000)))

	if err := keeper.CreatePoolWithCustody(ctx, provider, "atom", math.NewInt(1_000_000), math.NewInt(1_000_000)); err != nil {
		t.Fatal(err)
	}
	pool, _ := keeper.GetPool(ctx, "atom")
	if !keeper.GetLPBalance(ctx, "atom", provider).Equal(pool.TotalShares) {
		t.Fatal("initial LP shares were not assigned to provider")
	}
	secondPnyxBefore := bank.balance(ctx, accountOwner(second), pnyxDenom)
	secondAtomBefore := bank.balance(ctx, accountOwner(second), "atom")
	if _, err := keeper.AddLiquidityWithCustody(ctx, second, "atom", math.NewInt(100_000), math.NewInt(50_000)); err == nil {
		t.Fatal("expected imbalanced custodial deposit to fail")
	}
	if !bank.balance(ctx, accountOwner(second), pnyxDenom).Equal(secondPnyxBefore) ||
		!bank.balance(ctx, accountOwner(second), "atom").Equal(secondAtomBefore) {
		t.Fatal("failed imbalanced deposit moved provider funds")
	}
	shares, err := keeper.AddLiquidityWithCustody(ctx, second, "atom", math.NewInt(100_000), math.NewInt(100_000))
	if err != nil {
		t.Fatal(err)
	}
	if !keeper.GetLPBalance(ctx, "atom", second).Equal(shares) {
		t.Fatal("added LP shares were not assigned to second provider")
	}
	if _, _, err := keeper.RemoveLiquidityWithCustody(ctx, attacker, "atom", math.OneInt()); err == nil {
		t.Fatal("attacker removed shares they do not own")
	}
	remove := shares.QuoRaw(2)
	if _, _, err := keeper.RemoveLiquidityWithCustody(ctx, second, "atom", remove); err != nil {
		t.Fatal(err)
	}
	if !keeper.GetLPBalance(ctx, "atom", second).Equal(shares.Sub(remove)) {
		t.Fatal("provider LP balance was not reduced exactly")
	}
	if err := keeper.validateCustodyAndShares(ctx); err != nil {
		t.Fatal(err)
	}
	remainingSecond := keeper.GetLPBalance(ctx, "atom", second)
	if _, _, err := keeper.RemoveLiquidityWithCustody(ctx, second, "atom", remainingSecond); err != nil {
		t.Fatal(err)
	}
	remainingProvider := keeper.GetLPBalance(ctx, "atom", provider)
	if _, _, err := keeper.RemoveLiquidityWithCustody(ctx, provider, "atom", remainingProvider); err != nil {
		t.Fatal(err)
	}
	if _, found := keeper.GetPool(ctx, "atom"); found {
		t.Fatal("fully withdrawn pool should be removed")
	}
	if err := keeper.validateCustodyAndShares(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestCustodySwapsSettleAndBurnCanonicalSupply(t *testing.T) {
	keeper, ctx, bank, _ := setupCustodyKeeper(t)
	provider := sdk.AccAddress("swap-provider")
	trader := sdk.AccAddress("swap-trader")
	bank.fundAccount(ctx, provider, sdk.NewCoins(sdk.NewInt64Coin(pnyxDenom, 2_000_000), sdk.NewInt64Coin("atom", 2_000_000)))
	bank.fundAccount(ctx, trader, sdk.NewCoins(sdk.NewInt64Coin(pnyxDenom, 100_000), sdk.NewInt64Coin("atom", 100_000)))
	if err := keeper.CreatePoolWithCustody(ctx, provider, "atom", math.NewInt(1_000_000), math.NewInt(1_000_000)); err != nil {
		t.Fatal(err)
	}

	supplyBefore := bank.GetSupply(ctx, pnyxDenom).Amount
	atomBefore := bank.balance(ctx, accountOwner(trader), "atom")
	poolInvariantBefore, _ := keeper.GetPool(ctx, "atom")
	constantProductBefore := poolInvariantBefore.PnyxReserve.Mul(poolInvariantBefore.AssetReserve)
	output, err := keeper.SwapWithCustody(ctx, trader, pnyxDenom, math.NewInt(10_000), "atom", math.OneInt())
	if err != nil {
		t.Fatal(err)
	}
	if got := bank.balance(ctx, accountOwner(trader), "atom").Sub(atomBefore); !got.Equal(output) {
		t.Fatalf("asset output settled %s, want %s", got, output)
	}
	if supply := bank.GetSupply(ctx, pnyxDenom).Amount; !supply.Equal(supplyBefore) {
		t.Fatal("PNYX-input swap changed canonical supply")
	}
	poolInvariantAfter, _ := keeper.GetPool(ctx, "atom")
	if product := poolInvariantAfter.PnyxReserve.Mul(poolInvariantAfter.AssetReserve); product.LT(constantProductBefore) {
		t.Fatalf("constant product decreased: before=%s after=%s", constantProductBefore, product)
	}

	poolBefore, _ := keeper.GetPool(ctx, "atom")
	pnyxBefore := bank.balance(ctx, accountOwner(trader), pnyxDenom)
	supplyBefore = bank.GetSupply(ctx, pnyxDenom).Amount
	output, err = keeper.SwapWithCustody(ctx, trader, "atom", math.NewInt(10_000), pnyxDenom, math.OneInt())
	if err != nil {
		t.Fatal(err)
	}
	poolAfter, _ := keeper.GetPool(ctx, "atom")
	burn := poolAfter.TotalBurned.Sub(poolBefore.TotalBurned)
	if got := bank.balance(ctx, accountOwner(trader), pnyxDenom).Sub(pnyxBefore); !got.Equal(output) {
		t.Fatalf("PNYX output settled %s, want %s", got, output)
	}
	if supply := bank.GetSupply(ctx, pnyxDenom).Amount; !supply.Equal(supplyBefore.Sub(burn)) {
		t.Fatalf("canonical supply did not burn %s: got %s", burn, supply)
	}
	if err := keeper.validateCustodyAndShares(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestCrossAssetSwapIsAtomicAndCustodied(t *testing.T) {
	keeper, ctx, bank, _ := setupCustodyKeeper(t)
	provider := sdk.AccAddress("cross-provider")
	trader := sdk.AccAddress("cross-trader")
	bank.fundAccount(ctx, provider, sdk.NewCoins(
		sdk.NewInt64Coin(pnyxDenom, 4_000_000),
		sdk.NewInt64Coin("atom", 2_000_000),
		sdk.NewInt64Coin("btc", 2_000_000),
	))
	bank.fundAccount(ctx, trader, sdk.NewCoins(sdk.NewInt64Coin("atom", 100_000)))
	if err := keeper.CreatePoolWithCustody(ctx, provider, "atom", math.NewInt(1_000_000), math.NewInt(1_000_000)); err != nil {
		t.Fatal(err)
	}
	if err := keeper.CreatePoolWithCustody(ctx, provider, "btc", math.NewInt(1_000_000), math.NewInt(1_000_000)); err != nil {
		t.Fatal(err)
	}

	btcBefore := bank.balance(ctx, accountOwner(trader), "btc")
	supplyBefore := bank.GetSupply(ctx, pnyxDenom).Amount
	atomPoolBefore, _ := keeper.GetPool(ctx, "atom")
	output, err := keeper.SwapExactWithCustody(ctx, trader, "atom", math.NewInt(10_000), "btc", math.OneInt())
	if err != nil {
		t.Fatal(err)
	}
	if got := bank.balance(ctx, accountOwner(trader), "btc").Sub(btcBefore); !got.Equal(output) {
		t.Fatalf("cross-asset output = %s, want %s", got, output)
	}
	atomPoolAfter, _ := keeper.GetPool(ctx, "atom")
	burn := atomPoolAfter.TotalBurned.Sub(atomPoolBefore.TotalBurned)
	if supply := bank.GetSupply(ctx, pnyxDenom).Amount; !supply.Equal(supplyBefore.Sub(burn)) {
		t.Fatalf("cross-asset burn did not reduce supply by %s: %s", burn, supply)
	}
	if err := keeper.validateCustodyAndShares(ctx); err != nil {
		t.Fatal(err)
	}

	atomBalance := bank.balance(ctx, accountOwner(trader), "atom")
	atomPool, _ := keeper.GetPool(ctx, "atom")
	btcPool, _ := keeper.GetPool(ctx, "btc")
	if _, err := keeper.SwapExactWithCustody(ctx, trader, "atom", math.NewInt(1_000), "btc", math.NewInt(1_000_000)); err == nil {
		t.Fatal("expected final slippage failure")
	}
	if !bank.balance(ctx, accountOwner(trader), "atom").Equal(atomBalance) {
		t.Fatal("failed cross swap changed trader balance")
	}
	atomAfter, _ := keeper.GetPool(ctx, "atom")
	btcAfter, _ := keeper.GetPool(ctx, "btc")
	if !atomAfter.PnyxReserve.Equal(atomPool.PnyxReserve) || !atomAfter.AssetReserve.Equal(atomPool.AssetReserve) ||
		!btcAfter.PnyxReserve.Equal(btcPool.PnyxReserve) || !btcAfter.AssetReserve.Equal(btcPool.AssetReserve) {
		t.Fatal("failed cross swap committed pool state")
	}
}

func TestCustodyBankFailuresRollbackAllState(t *testing.T) {
	keeper, ctx, bank, _ := setupCustodyKeeper(t)
	provider := sdk.AccAddress("rollback-provider")
	bank.fundAccount(ctx, provider, sdk.NewCoins(sdk.NewInt64Coin(pnyxDenom, 1_000_000), sdk.NewInt64Coin("atom", 1_000_000)))
	bank.failAccountToModule = true
	if err := keeper.CreatePoolWithCustody(ctx, provider, "atom", math.NewInt(500_000), math.NewInt(500_000)); err == nil {
		t.Fatal("expected injected funding failure")
	}
	if _, found := keeper.GetPool(ctx, "atom"); found {
		t.Fatal("failed funding committed pool")
	}
	if !keeper.GetLPBalance(ctx, "atom", provider).IsZero() {
		t.Fatal("failed funding committed LP shares")
	}

	bank.failAccountToModule = false
	if err := keeper.CreatePoolWithCustody(ctx, provider, "atom", math.NewInt(500_000), math.NewInt(500_000)); err != nil {
		t.Fatal(err)
	}
	trader := sdk.AccAddress("rollback-trader")
	bank.fundAccount(ctx, trader, sdk.NewCoins(sdk.NewInt64Coin("atom", 10_000)))
	bank.failBurn = true
	traderBefore := bank.balance(ctx, accountOwner(trader), "atom")
	poolBefore, _ := keeper.GetPool(ctx, "atom")
	supplyBefore := bank.GetSupply(ctx, pnyxDenom).Amount
	if _, err := keeper.SwapWithCustody(ctx, trader, "atom", math.NewInt(10_000), pnyxDenom, math.OneInt()); err == nil {
		t.Fatal("expected injected burn failure")
	}
	poolAfter, _ := keeper.GetPool(ctx, "atom")
	if !poolAfter.PnyxReserve.Equal(poolBefore.PnyxReserve) || !poolAfter.AssetReserve.Equal(poolBefore.AssetReserve) {
		t.Fatal("failed burn committed pool reserves")
	}
	if !bank.balance(ctx, accountOwner(trader), "atom").Equal(traderBefore) {
		t.Fatal("failed burn committed trader input")
	}
	if !bank.GetSupply(ctx, pnyxDenom).Amount.Equal(supplyBefore) {
		t.Fatal("failed burn changed canonical supply")
	}
}

func TestCustodyAddRemoveAndSwapTransferFailuresRollback(t *testing.T) {
	keeper, ctx, bank, _ := setupCustodyKeeper(t)
	provider := sdk.AccAddress("transfer-provider")
	second := sdk.AccAddress("transfer-second")
	trader := sdk.AccAddress("transfer-trader")
	bank.fundAccount(ctx, provider, sdk.NewCoins(
		sdk.NewInt64Coin(pnyxDenom, 2_000_000),
		sdk.NewInt64Coin("atom", 2_000_000),
	))
	bank.fundAccount(ctx, second, sdk.NewCoins(
		sdk.NewInt64Coin(pnyxDenom, 100_000),
		sdk.NewInt64Coin("atom", 100_000),
	))
	bank.fundAccount(ctx, trader, sdk.NewCoins(sdk.NewInt64Coin("atom", 100_000)))
	if err := keeper.CreatePoolWithCustody(ctx, provider, "atom", math.NewInt(1_000_000), math.NewInt(1_000_000)); err != nil {
		t.Fatal(err)
	}

	poolBefore, _ := keeper.GetPool(ctx, "atom")
	secondPnyxBefore := bank.balance(ctx, accountOwner(second), pnyxDenom)
	secondAtomBefore := bank.balance(ctx, accountOwner(second), "atom")
	bank.failAccountToModule = true
	if _, err := keeper.AddLiquidityWithCustody(ctx, second, "atom", math.NewInt(100_000), math.NewInt(100_000)); err == nil {
		t.Fatal("expected injected add-liquidity transfer failure")
	}
	bank.failAccountToModule = false
	poolAfter, _ := keeper.GetPool(ctx, "atom")
	if !poolAfter.PnyxReserve.Equal(poolBefore.PnyxReserve) ||
		!poolAfter.AssetReserve.Equal(poolBefore.AssetReserve) ||
		!poolAfter.TotalShares.Equal(poolBefore.TotalShares) {
		t.Fatal("failed liquidity deposit committed pool state")
	}
	if !keeper.GetLPBalance(ctx, "atom", second).IsZero() {
		t.Fatal("failed liquidity deposit committed LP ownership")
	}
	if !bank.balance(ctx, accountOwner(second), pnyxDenom).Equal(secondPnyxBefore) ||
		!bank.balance(ctx, accountOwner(second), "atom").Equal(secondAtomBefore) {
		t.Fatal("failed liquidity deposit changed provider balances")
	}

	providerShares := keeper.GetLPBalance(ctx, "atom", provider)
	providerPnyxBefore := bank.balance(ctx, accountOwner(provider), pnyxDenom)
	providerAtomBefore := bank.balance(ctx, accountOwner(provider), "atom")
	bank.failModuleToAccount = true
	if _, _, err := keeper.RemoveLiquidityWithCustody(ctx, provider, "atom", providerShares.QuoRaw(2)); err == nil {
		t.Fatal("expected injected remove-liquidity transfer failure")
	}
	bank.failModuleToAccount = false
	poolAfter, _ = keeper.GetPool(ctx, "atom")
	if !poolAfter.PnyxReserve.Equal(poolBefore.PnyxReserve) ||
		!poolAfter.AssetReserve.Equal(poolBefore.AssetReserve) ||
		!poolAfter.TotalShares.Equal(poolBefore.TotalShares) {
		t.Fatal("failed liquidity withdrawal committed pool state")
	}
	if !keeper.GetLPBalance(ctx, "atom", provider).Equal(providerShares) {
		t.Fatal("failed liquidity withdrawal changed LP ownership")
	}
	if !bank.balance(ctx, accountOwner(provider), pnyxDenom).Equal(providerPnyxBefore) ||
		!bank.balance(ctx, accountOwner(provider), "atom").Equal(providerAtomBefore) {
		t.Fatal("failed liquidity withdrawal changed provider balances")
	}

	traderAtomBefore := bank.balance(ctx, accountOwner(trader), "atom")
	traderPnyxBefore := bank.balance(ctx, accountOwner(trader), pnyxDenom)
	supplyBefore := bank.GetSupply(ctx, pnyxDenom).Amount
	bank.failModuleToAccount = true
	if _, err := keeper.SwapWithCustody(ctx, trader, "atom", math.NewInt(10_000), pnyxDenom, math.OneInt()); err == nil {
		t.Fatal("expected injected swap output transfer failure")
	}
	bank.failModuleToAccount = false
	poolAfter, _ = keeper.GetPool(ctx, "atom")
	if !poolAfter.PnyxReserve.Equal(poolBefore.PnyxReserve) ||
		!poolAfter.AssetReserve.Equal(poolBefore.AssetReserve) ||
		!poolAfter.TotalBurned.Equal(poolBefore.TotalBurned) ||
		poolAfter.SwapCount != poolBefore.SwapCount {
		t.Fatal("failed swap output transfer committed pool state")
	}
	if !bank.balance(ctx, accountOwner(trader), "atom").Equal(traderAtomBefore) ||
		!bank.balance(ctx, accountOwner(trader), pnyxDenom).Equal(traderPnyxBefore) {
		t.Fatal("failed swap output transfer changed trader balances")
	}
	if !bank.GetSupply(ctx, pnyxDenom).Amount.Equal(supplyBefore) {
		t.Fatal("failed swap output transfer changed canonical supply")
	}
	if err := keeper.validateCustodyAndShares(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestRegistryRequiresChainAuthority(t *testing.T) {
	keeper, ctx, _, authority := setupCustodyKeeper(t)
	server := NewMsgServer(keeper)
	asset := &MsgRegisterAsset{Sender: sdk.AccAddress("attacker"), IBCDenom: "eth", Symbol: "ETH", Decimals: 18}
	if _, err := server.RegisterAsset(ctx, asset); err == nil {
		t.Fatal("unauthorized asset registration succeeded")
	}
	asset.Sender = authority
	if _, err := server.RegisterAsset(ctx, asset); err != nil {
		t.Fatalf("authority registration failed: %v", err)
	}
	status := &MsgUpdateAssetStatus{Sender: sdk.AccAddress("attacker"), IBCDenom: "eth", Enabled: false}
	if _, err := server.UpdateAssetStatus(ctx, status); err == nil {
		t.Fatal("unauthorized status update succeeded")
	}
	status.Sender = authority
	if _, err := server.UpdateAssetStatus(ctx, status); err != nil {
		t.Fatalf("authority status update failed: %v", err)
	}
}

func TestCustodyInvariantsDetectDivergence(t *testing.T) {
	keeper, ctx, bank, _ := setupCustodyKeeper(t)
	provider := sdk.AccAddress("invariant-provider")
	bank.fundAccount(ctx, provider, sdk.NewCoins(
		sdk.NewInt64Coin(pnyxDenom, 1_000_001),
		sdk.NewInt64Coin("atom", 1_000_001),
		sdk.NewInt64Coin("mystery", 1),
	))
	if err := keeper.CreatePoolWithCustody(ctx, provider, "atom", math.NewInt(1_000_000), math.NewInt(1_000_000)); err != nil {
		t.Fatal(err)
	}

	owned := keeper.GetLPBalance(ctx, "atom", provider)
	keeper.setLPBalance(ctx, "atom", provider, owned.AddRaw(1))
	if err := keeper.ValidateLPConservation(ctx); err == nil {
		t.Fatal("LP invariant missed provider/total divergence")
	}
	keeper.setLPBalance(ctx, "atom", provider, owned)

	if err := bank.SendCoinsFromAccountToModule(
		ctx,
		provider,
		ModuleName,
		sdk.NewCoins(sdk.NewInt64Coin("mystery", 1)),
	); err != nil {
		t.Fatal(err)
	}
	if err := keeper.ValidateReserveCustody(ctx); err == nil {
		t.Fatal("reserve invariant missed excess module balance")
	}
}

func TestLPShareTotalsDoNotCollideAcrossDenomPrefixes(t *testing.T) {
	keeper, ctx, bank, _ := setupCustodyKeeper(t)
	provider := sdk.AccAddress("prefix-provider")
	const prefixedDenom = "atom:staked"
	if err := keeper.RegisterAsset(ctx, RegisteredAsset{
		IBCDenom:       prefixedDenom,
		Symbol:         "stATOM",
		Decimals:       6,
		TradingEnabled: true,
	}); err != nil {
		t.Fatal(err)
	}
	bank.fundAccount(ctx, provider, sdk.NewCoins(
		sdk.NewInt64Coin(pnyxDenom, 4_000_000),
		sdk.NewInt64Coin("atom", 2_000_000),
		sdk.NewInt64Coin(prefixedDenom, 2_000_000),
	))
	if err := keeper.CreatePoolWithCustody(ctx, provider, "atom", math.NewInt(1_000_000), math.NewInt(1_000_000)); err != nil {
		t.Fatal(err)
	}
	if err := keeper.CreatePoolWithCustody(ctx, provider, prefixedDenom, math.NewInt(1_000_000), math.NewInt(1_000_000)); err != nil {
		t.Fatal(err)
	}

	for _, denom := range []string{"atom", prefixedDenom} {
		pool, found := keeper.GetPool(ctx, denom)
		if !found {
			t.Fatalf("pool %s not found", denom)
		}
		if total := keeper.LPShareTotal(ctx, denom); !total.Equal(pool.TotalShares) {
			t.Fatalf("LP total for %s = %s, want %s", denom, total, pool.TotalShares)
		}
	}
	if err := keeper.ValidateLPConservation(ctx); err != nil {
		t.Fatal(err)
	}
}
