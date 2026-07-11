package dex

import (
	"sort"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func lpBalanceKey(assetDenom, provider string) []byte {
	return []byte("lp:" + assetDenom + ":" + provider)
}

func lpPoolPrefix(assetDenom string) []byte {
	return []byte("lp:" + assetDenom + ":")
}

func (k Keeper) requireBank() error {
	if k.bank == nil {
		return errorsmod.Wrap(sdkerrors.ErrLogic, "DEX bank keeper is not available")
	}
	return nil
}

func (k Keeper) RequireAuthority(sender sdk.AccAddress) error {
	if sender.Empty() || k.authority == "" || sender.String() != k.authority {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "DEX registry change requires chain authority")
	}
	return nil
}

func (k Keeper) GetLPBalance(ctx sdk.Context, assetDenom string, provider sdk.AccAddress) math.Int {
	bz := ctx.KVStore(k.StoreKey).Get(lpBalanceKey(assetDenom, provider.String()))
	if bz == nil {
		return math.ZeroInt()
	}
	var shares math.Int
	k.cdc.MustUnmarshalLengthPrefixed(bz, &shares)
	return shares
}

func (k Keeper) setLPBalance(ctx sdk.Context, assetDenom string, provider sdk.AccAddress, shares math.Int) {
	store := ctx.KVStore(k.StoreKey)
	key := lpBalanceKey(assetDenom, provider.String())
	if !shares.IsPositive() {
		store.Delete(key)
		return
	}
	store.Set(key, k.cdc.MustMarshalLengthPrefixed(&shares))
}

func (k Keeper) LPShareTotal(ctx sdk.Context, assetDenom string) math.Int {
	store := ctx.KVStore(k.StoreKey)
	prefix := lpPoolPrefix(assetDenom)
	iterator := store.Iterator(prefix, prefixEnd(prefix))
	defer iterator.Close()
	total := math.ZeroInt()
	for ; iterator.Valid(); iterator.Next() {
		var shares math.Int
		k.cdc.MustUnmarshalLengthPrefixed(iterator.Value(), &shares)
		total = total.Add(shares)
	}
	return total
}

func (k Keeper) ValidateLPConservation(ctx sdk.Context) error {
	var invariantErr error
	k.IteratePools(ctx, func(pool Pool) bool {
		providerShares := k.LPShareTotal(ctx, pool.AssetDenom)
		if !providerShares.Equal(pool.TotalShares) {
			invariantErr = errorsmod.Wrapf(
				sdkerrors.ErrLogic,
				"LP share mismatch for %s: providers=%s total=%s",
				pool.AssetDenom,
				providerShares,
				pool.TotalShares,
			)
			return true
		}
		return false
	})
	return invariantErr
}

func (k Keeper) ReserveClaims(ctx sdk.Context) sdk.Coins {
	claims := sdk.Coins{}
	k.IteratePools(ctx, func(pool Pool) bool {
		if pool.PnyxReserve.IsPositive() {
			claims = claims.Add(sdk.NewCoin(pnyxDenom, pool.PnyxReserve))
		}
		if pool.AssetReserve.IsPositive() {
			claims = claims.Add(sdk.NewCoin(pool.AssetDenom, pool.AssetReserve))
		}
		return false
	})
	return claims
}

func (k Keeper) ValidateReserveCustody(ctx sdk.Context) error {
	if err := k.requireBank(); err != nil {
		return err
	}
	claims := k.ReserveClaims(ctx)
	denoms := map[string]struct{}{pnyxDenom: {}}
	k.IteratePools(ctx, func(pool Pool) bool {
		denoms[pool.AssetDenom] = struct{}{}
		return false
	})
	for _, asset := range k.GetAllAssets(ctx) {
		denoms[asset.IBCDenom] = struct{}{}
	}
	moduleAddress := authtypes.NewModuleAddress(ModuleName)
	orderedDenoms := make([]string, 0, len(denoms))
	for denom := range denoms {
		orderedDenoms = append(orderedDenoms, denom)
	}
	sort.Strings(orderedDenoms)
	for _, denom := range orderedDenoms {
		balance := k.bank.GetBalance(ctx, moduleAddress, denom)
		claim := claims.AmountOf(denom)
		if !balance.Amount.Equal(claim) {
			return errorsmod.Wrapf(
				sdkerrors.ErrLogic,
				"DEX reserve mismatch for %s: bank=%s claims=%s",
				denom,
				balance.Amount,
				claim,
			)
		}
	}
	return nil
}

func (k Keeper) validateCustodyAndShares(ctx sdk.Context) error {
	if err := k.ValidateReserveCustody(ctx); err != nil {
		return err
	}
	return k.ValidateLPConservation(ctx)
}

func (k Keeper) CreatePoolWithCustody(
	ctx sdk.Context,
	provider sdk.AccAddress,
	assetDenom string,
	pnyxAmount, assetAmount math.Int,
) error {
	if err := k.requireBank(); err != nil {
		return err
	}
	if provider.Empty() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "liquidity provider is required")
	}
	if err := sdk.ValidateDenom(assetDenom); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid asset denom")
	}

	cacheCtx, write := ctx.CacheContext()
	if err := k.CreatePool(cacheCtx, assetDenom, pnyxAmount, assetAmount); err != nil {
		return err
	}
	pool, _ := k.GetPool(cacheCtx, assetDenom)
	k.setLPBalance(cacheCtx, assetDenom, provider, pool.TotalShares)
	coins := sdk.NewCoins(
		sdk.NewCoin(pnyxDenom, pnyxAmount),
		sdk.NewCoin(assetDenom, assetAmount),
	)
	if err := k.bank.SendCoinsFromAccountToModule(cacheCtx, provider, ModuleName, coins); err != nil {
		return errorsmod.Wrap(err, "initial DEX liquidity transfer failed")
	}
	if err := k.validateCustodyAndShares(cacheCtx); err != nil {
		return err
	}
	write()
	return nil
}

func (k Keeper) AddLiquidityWithCustody(
	ctx sdk.Context,
	provider sdk.AccAddress,
	assetDenom string,
	pnyxAmount, assetAmount math.Int,
) (math.Int, error) {
	if err := k.requireBank(); err != nil {
		return math.Int{}, err
	}
	if provider.Empty() {
		return math.Int{}, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "liquidity provider is required")
	}
	if err := k.validateAssetForTrading(ctx, assetDenom); err != nil {
		return math.Int{}, err
	}
	cacheCtx, write := ctx.CacheContext()
	shares, err := k.AddLiquidity(cacheCtx, assetDenom, pnyxAmount, assetAmount)
	if err != nil {
		return math.Int{}, err
	}
	current := k.GetLPBalance(cacheCtx, assetDenom, provider)
	k.setLPBalance(cacheCtx, assetDenom, provider, current.Add(shares))
	coins := sdk.NewCoins(
		sdk.NewCoin(pnyxDenom, pnyxAmount),
		sdk.NewCoin(assetDenom, assetAmount),
	)
	if err := k.bank.SendCoinsFromAccountToModule(cacheCtx, provider, ModuleName, coins); err != nil {
		return math.Int{}, errorsmod.Wrap(err, "DEX liquidity transfer failed")
	}
	if err := k.validateCustodyAndShares(cacheCtx); err != nil {
		return math.Int{}, err
	}
	write()
	return shares, nil
}

func (k Keeper) RemoveLiquidityWithCustody(
	ctx sdk.Context,
	provider sdk.AccAddress,
	assetDenom string,
	shares math.Int,
) (math.Int, math.Int, error) {
	if err := k.requireBank(); err != nil {
		return math.Int{}, math.Int{}, err
	}
	owned := k.GetLPBalance(ctx, assetDenom, provider)
	if shares.IsNil() || !shares.IsPositive() || shares.GT(owned) {
		return math.Int{}, math.Int{}, errorsmod.Wrapf(
			sdkerrors.ErrUnauthorized,
			"requested LP shares %s exceed provider balance %s",
			shares,
			owned,
		)
	}
	cacheCtx, write := ctx.CacheContext()
	pnyxOutput, assetOutput, err := k.RemoveLiquidity(cacheCtx, assetDenom, shares)
	if err != nil {
		return math.Int{}, math.Int{}, err
	}
	k.setLPBalance(cacheCtx, assetDenom, provider, owned.Sub(shares))
	coins := sdk.NewCoins(
		sdk.NewCoin(pnyxDenom, pnyxOutput),
		sdk.NewCoin(assetDenom, assetOutput),
	)
	if err := k.bank.SendCoinsFromModuleToAccount(cacheCtx, ModuleName, provider, coins); err != nil {
		return math.Int{}, math.Int{}, errorsmod.Wrap(err, "DEX liquidity withdrawal failed")
	}
	if err := k.validateCustodyAndShares(cacheCtx); err != nil {
		return math.Int{}, math.Int{}, err
	}
	write()
	return pnyxOutput, assetOutput, nil
}

func (k Keeper) SwapWithCustody(
	ctx sdk.Context,
	trader sdk.AccAddress,
	inputDenom string,
	inputAmount math.Int,
	outputDenom string,
	minOutput math.Int,
) (math.Int, error) {
	if err := k.requireBank(); err != nil {
		return math.Int{}, err
	}
	if trader.Empty() {
		return math.Int{}, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "trader is required")
	}
	assetDenom, err := resolveAssetDenom(inputDenom, outputDenom)
	if err != nil {
		return math.Int{}, err
	}
	poolBefore, found := k.GetPool(ctx, assetDenom)
	if !found {
		return math.Int{}, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "no pool for %s", assetDenom)
	}

	cacheCtx, write := ctx.CacheContext()
	output, err := k.Swap(cacheCtx, inputDenom, inputAmount, outputDenom, minOutput)
	if err != nil {
		return math.Int{}, err
	}
	poolAfter, _ := k.GetPool(cacheCtx, assetDenom)
	burn := poolAfter.TotalBurned.Sub(poolBefore.TotalBurned)
	if err := k.settleSwap(cacheCtx, trader, inputDenom, inputAmount, outputDenom, output, burn); err != nil {
		return math.Int{}, err
	}
	if err := k.validateCustodyAndShares(cacheCtx); err != nil {
		return math.Int{}, err
	}
	write()
	return output, nil
}

func (k Keeper) SwapExactWithCustody(
	ctx sdk.Context,
	trader sdk.AccAddress,
	inputDenom string,
	inputAmount math.Int,
	outputDenom string,
	minOutput math.Int,
) (math.Int, error) {
	if minOutput.IsNil() || !minOutput.IsPositive() {
		return math.Int{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "minimum output must be positive")
	}
	if inputDenom == pnyxDenom || outputDenom == pnyxDenom {
		return k.SwapWithCustody(ctx, trader, inputDenom, inputAmount, outputDenom, minOutput)
	}
	if err := k.requireBank(); err != nil {
		return math.Int{}, err
	}
	if trader.Empty() {
		return math.Int{}, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "trader is required")
	}
	poolBefore, found := k.GetPool(ctx, inputDenom)
	if !found {
		return math.Int{}, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "no pool for %s", inputDenom)
	}

	cacheCtx, write := ctx.CacheContext()
	output, err := k.SwapExact(cacheCtx, inputDenom, inputAmount, outputDenom, minOutput)
	if err != nil {
		return math.Int{}, err
	}
	poolAfter, _ := k.GetPool(cacheCtx, inputDenom)
	burn := poolAfter.TotalBurned.Sub(poolBefore.TotalBurned)
	if err := k.settleSwap(cacheCtx, trader, inputDenom, inputAmount, outputDenom, output, burn); err != nil {
		return math.Int{}, err
	}
	if err := k.validateCustodyAndShares(cacheCtx); err != nil {
		return math.Int{}, err
	}
	write()
	return output, nil
}

func (k Keeper) settleSwap(
	ctx sdk.Context,
	trader sdk.AccAddress,
	inputDenom string,
	inputAmount math.Int,
	outputDenom string,
	outputAmount math.Int,
	burn math.Int,
) error {
	if err := k.bank.SendCoinsFromAccountToModule(
		ctx,
		trader,
		ModuleName,
		sdk.NewCoins(sdk.NewCoin(inputDenom, inputAmount)),
	); err != nil {
		return errorsmod.Wrap(err, "DEX swap input transfer failed")
	}
	if err := k.bank.SendCoinsFromModuleToAccount(
		ctx,
		ModuleName,
		trader,
		sdk.NewCoins(sdk.NewCoin(outputDenom, outputAmount)),
	); err != nil {
		return errorsmod.Wrap(err, "DEX swap output transfer failed")
	}
	if burn.IsPositive() {
		if err := k.issuer.Burn(ctx, burn); err != nil {
			return errorsmod.Wrap(err, "DEX swap burn failed")
		}
	}
	return nil
}
