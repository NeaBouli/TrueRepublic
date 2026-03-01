package dex

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const pnyxDenom = "pnyx"

type Keeper struct {
	StoreKey storetypes.StoreKey
	cdc      *codec.LegacyAmino
}

func NewKeeper(cdc *codec.LegacyAmino, storeKey storetypes.StoreKey) Keeper {
	return Keeper{StoreKey: storeKey, cdc: cdc}
}

func poolKey(assetDenom string) []byte {
	return []byte("pool:" + assetDenom)
}

// GetPool loads a liquidity pool from the store.
func (k Keeper) GetPool(ctx sdk.Context, assetDenom string) (Pool, bool) {
	store := ctx.KVStore(k.StoreKey)
	bz := store.Get(poolKey(assetDenom))
	if bz == nil {
		return Pool{}, false
	}
	var pool Pool
	k.cdc.MustUnmarshalLengthPrefixed(bz, &pool)
	return pool, true
}

// SetPool persists a pool to the store.
func (k Keeper) SetPool(ctx sdk.Context, pool Pool) {
	store := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&pool)
	store.Set(poolKey(pool.AssetDenom), bz)
}

// CreatePool initialises a new PNYX/<asset> liquidity pool.
// Initial shares are set to sqrt(pnyxAmt * assetAmt) using integer sqrt.
func (k Keeper) CreatePool(ctx sdk.Context, assetDenom string, pnyxAmt, assetAmt math.Int) error {
	if !pnyxAmt.IsPositive() || !assetAmt.IsPositive() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "both reserve amounts must be positive")
	}

	// Validate asset is registered and trading enabled.
	if err := k.validateAssetForTrading(ctx, assetDenom); err != nil {
		return err
	}

	if _, exists := k.GetPool(ctx, assetDenom); exists {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "pool for %s already exists", assetDenom)
	}

	shares := intSqrt(pnyxAmt.Mul(assetAmt))

	pool := Pool{
		PnyxReserve:  pnyxAmt,
		AssetReserve: assetAmt,
		AssetDenom:   assetDenom,
		TotalShares:  shares,
		TotalBurned:  math.ZeroInt(),
	}
	k.SetPool(ctx, pool)
	return nil
}

// computeSwapOutput calculates AMM output from reserves without side effects.
// Returns (outputAmt, burnAmt). burnAmt is nonzero only when outputIsPnyx.
func computeSwapOutput(inReserve, outReserve, inputAmt math.Int, outputIsPnyx bool) (math.Int, math.Int) {
	feeMultiplier := math.NewInt(10000 - SwapFeeBps) // 9970
	numerator := outReserve.Mul(inputAmt).Mul(feeMultiplier)
	denominator := inReserve.Mul(math.NewInt(10000)).Add(inputAmt.Mul(feeMultiplier))
	outputAmt := numerator.Quo(denominator)

	burnAmt := math.ZeroInt()
	if outputIsPnyx {
		burnAmt = outputAmt.Mul(math.NewInt(BurnBps)).Quo(math.NewInt(10000))
		if burnAmt.IsPositive() {
			outputAmt = outputAmt.Sub(burnAmt)
		}
	}
	return outputAmt, burnAmt
}

// Swap executes a constant-product AMM swap with a 0.3% fee.
//
// The output amount is:
//
//	out = (outReserve * in * (10000 - fee)) / (inReserve * 10000 + in * (10000 - fee))
//
// One of inputDenom/outputDenom must be "pnyx" and the other the pool's
// asset denom. If minOutput is positive, the swap fails when the output
// would be less than minOutput (slippage protection).
func (k Keeper) Swap(ctx sdk.Context, inputDenom string, inputAmt math.Int, outputDenom string, minOutput math.Int) (math.Int, error) {
	if !inputAmt.IsPositive() {
		return math.Int{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "input amount must be positive")
	}

	// Determine which denom is the asset side.
	assetDenom, err := resolveAssetDenom(inputDenom, outputDenom)
	if err != nil {
		return math.Int{}, err
	}

	// Validate asset trading status.
	if err := k.validateAssetForTrading(ctx, assetDenom); err != nil {
		return math.Int{}, err
	}

	pool, found := k.GetPool(ctx, assetDenom)
	if !found {
		return math.Int{}, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "no pool for %s", assetDenom)
	}

	var inReserve, outReserve math.Int
	if inputDenom == pnyxDenom {
		inReserve = pool.PnyxReserve
		outReserve = pool.AssetReserve
	} else {
		inReserve = pool.AssetReserve
		outReserve = pool.PnyxReserve
	}

	outputAmt, burnAmt := computeSwapOutput(inReserve, outReserve, inputAmt, outputDenom == pnyxDenom)

	if !outputAmt.IsPositive() {
		return math.Int{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "output amount is zero")
	}
	if outputAmt.Add(burnAmt).GTE(outReserve) {
		return math.Int{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "swap would drain the pool")
	}

	// Slippage protection.
	if minOutput.IsPositive() && outputAmt.LT(minOutput) {
		return math.Int{}, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"slippage: output %s below minimum %s", outputAmt, minOutput)
	}

	// Track burn.
	if burnAmt.IsPositive() {
		pool.TotalBurned = pool.TotalBurned.Add(burnAmt)
	}

	// Update reserves.
	if inputDenom == pnyxDenom {
		pool.PnyxReserve = pool.PnyxReserve.Add(inputAmt)
		pool.AssetReserve = pool.AssetReserve.Sub(outputAmt)
	} else {
		pool.AssetReserve = pool.AssetReserve.Add(inputAmt)
		// Subtract output + burn from PNYX reserve (burn removes from circulation).
		pool.PnyxReserve = pool.PnyxReserve.Sub(outputAmt).Sub(burnAmt)
	}

	k.SetPool(ctx, pool)
	return outputAmt, nil
}

// SwapExact executes a swap with slippage protection, automatically routing
// cross-asset swaps through the PNYX hub. If one side is PNYX, it delegates
// to Swap(). If neither side is PNYX, it performs two atomic hops:
// inputDenom -> PNYX -> outputDenom.
func (k Keeper) SwapExact(ctx sdk.Context, inputDenom string, inputAmt math.Int, outputDenom string, minOutput math.Int) (math.Int, error) {
	if inputDenom == outputDenom {
		return math.Int{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "input and output denoms must differ")
	}

	// Direct swap: one side is PNYX.
	if inputDenom == pnyxDenom || outputDenom == pnyxDenom {
		return k.Swap(ctx, inputDenom, inputAmt, outputDenom, minOutput)
	}

	// Cross-asset swap: route through PNYX hub (2 hops).
	// Validate both assets.
	if err := k.validateAssetForTrading(ctx, inputDenom); err != nil {
		return math.Int{}, err
	}
	if err := k.validateAssetForTrading(ctx, outputDenom); err != nil {
		return math.Int{}, err
	}

	// Hop 1: inputDenom -> PNYX (no minOutput on intermediate).
	intermediateAmt, err := k.Swap(ctx, inputDenom, inputAmt, pnyxDenom, math.ZeroInt())
	if err != nil {
		return math.Int{}, errorsmod.Wrapf(err, "hop 1 (%s -> PNYX) failed", inputDenom)
	}

	// Hop 2: PNYX -> outputDenom (no minOutput on intermediate).
	finalAmt, err := k.Swap(ctx, pnyxDenom, intermediateAmt, outputDenom, math.ZeroInt())
	if err != nil {
		return math.Int{}, errorsmod.Wrapf(err, "hop 2 (PNYX -> %s) failed", outputDenom)
	}

	// Check slippage on final output.
	if minOutput.IsPositive() && finalAmt.LT(minOutput) {
		return math.Int{}, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"slippage: output %s below minimum %s", finalAmt, minOutput)
	}

	return finalAmt, nil
}

// EstimateSwapOutput calculates the expected output for a swap without
// executing it. Returns (expectedOutput, route, error) where route is the
// list of denoms traversed (e.g., ["btc", "pnyx", "eth"] for cross-asset).
func (k Keeper) EstimateSwapOutput(ctx sdk.Context, inputDenom string, inputAmt math.Int, outputDenom string) (math.Int, []string, error) {
	if inputDenom == outputDenom {
		return math.Int{}, nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "input and output denoms must differ")
	}

	// Direct swap: one side is PNYX.
	if inputDenom == pnyxDenom || outputDenom == pnyxDenom {
		assetDenom, err := resolveAssetDenom(inputDenom, outputDenom)
		if err != nil {
			return math.Int{}, nil, err
		}
		pool, found := k.GetPool(ctx, assetDenom)
		if !found {
			return math.Int{}, nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "no pool for %s", assetDenom)
		}
		var inReserve, outReserve math.Int
		if inputDenom == pnyxDenom {
			inReserve = pool.PnyxReserve
			outReserve = pool.AssetReserve
		} else {
			inReserve = pool.AssetReserve
			outReserve = pool.PnyxReserve
		}
		outputAmt, _ := computeSwapOutput(inReserve, outReserve, inputAmt, outputDenom == pnyxDenom)
		return outputAmt, []string{inputDenom, outputDenom}, nil
	}

	// Cross-asset: route through PNYX hub.
	// Hop 1: inputDenom -> PNYX.
	pool1, found := k.GetPool(ctx, inputDenom)
	if !found {
		return math.Int{}, nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "no pool for %s", inputDenom)
	}
	intermediateAmt, _ := computeSwapOutput(pool1.AssetReserve, pool1.PnyxReserve, inputAmt, true)

	// Hop 2: PNYX -> outputDenom.
	pool2, found := k.GetPool(ctx, outputDenom)
	if !found {
		return math.Int{}, nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "no pool for %s", outputDenom)
	}
	finalAmt, _ := computeSwapOutput(pool2.PnyxReserve, pool2.AssetReserve, intermediateAmt, false)

	return finalAmt, []string{inputDenom, pnyxDenom, outputDenom}, nil
}

// AddLiquidity deposits PNYX and the paired asset proportionally and mints
// LP shares. The caller receives shares proportional to the smaller ratio
// of the two deposits relative to pool reserves.
func (k Keeper) AddLiquidity(ctx sdk.Context, assetDenom string, pnyxAmt, assetAmt math.Int) (math.Int, error) {
	if !pnyxAmt.IsPositive() || !assetAmt.IsPositive() {
		return math.Int{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "both amounts must be positive")
	}

	pool, found := k.GetPool(ctx, assetDenom)
	if !found {
		return math.Int{}, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "no pool for %s", assetDenom)
	}

	// shares = min(pnyxAmt/pnyxReserve, assetAmt/assetReserve) * totalShares
	// Using cross-multiplication to avoid decimal division:
	//   sharesByPnyx = pnyxAmt * totalShares / pnyxReserve
	//   sharesByAsset = assetAmt * totalShares / assetReserve
	sharesByPnyx := pnyxAmt.Mul(pool.TotalShares).Quo(pool.PnyxReserve)
	sharesByAsset := assetAmt.Mul(pool.TotalShares).Quo(pool.AssetReserve)

	shares := sharesByPnyx
	if sharesByAsset.LT(sharesByPnyx) {
		shares = sharesByAsset
	}

	if !shares.IsPositive() {
		return math.Int{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "deposit too small to mint shares")
	}

	pool.PnyxReserve = pool.PnyxReserve.Add(pnyxAmt)
	pool.AssetReserve = pool.AssetReserve.Add(assetAmt)
	pool.TotalShares = pool.TotalShares.Add(shares)

	k.SetPool(ctx, pool)
	return shares, nil
}

// RemoveLiquidity burns LP shares and returns the proportional amounts of
// PNYX and the paired asset.
func (k Keeper) RemoveLiquidity(ctx sdk.Context, assetDenom string, shares math.Int) (pnyxOut, assetOut math.Int, err error) {
	if !shares.IsPositive() {
		return math.Int{}, math.Int{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "shares must be positive")
	}

	pool, found := k.GetPool(ctx, assetDenom)
	if !found {
		return math.Int{}, math.Int{}, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "no pool for %s", assetDenom)
	}
	if shares.GT(pool.TotalShares) {
		return math.Int{}, math.Int{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "shares exceed total supply")
	}

	pnyxOut = pool.PnyxReserve.Mul(shares).Quo(pool.TotalShares)
	assetOut = pool.AssetReserve.Mul(shares).Quo(pool.TotalShares)

	pool.PnyxReserve = pool.PnyxReserve.Sub(pnyxOut)
	pool.AssetReserve = pool.AssetReserve.Sub(assetOut)
	pool.TotalShares = pool.TotalShares.Sub(shares)

	k.SetPool(ctx, pool)
	return pnyxOut, assetOut, nil
}

// resolveAssetDenom determines which of the two denoms is the non-PNYX asset.
func resolveAssetDenom(denomA, denomB string) (string, error) {
	switch {
	case denomA == pnyxDenom && denomB != pnyxDenom:
		return denomB, nil
	case denomA != pnyxDenom && denomB == pnyxDenom:
		return denomA, nil
	default:
		return "", errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "exactly one denom must be pnyx")
	}
}

// IteratePools iterates over all pools in the store.
func (k Keeper) IteratePools(ctx sdk.Context, cb func(Pool) bool) {
	store := ctx.KVStore(k.StoreKey)
	prefix := []byte("pool:")
	iter := store.Iterator(prefix, prefixEnd(prefix))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var pool Pool
		k.cdc.MustUnmarshalLengthPrefixed(iter.Value(), &pool)
		if cb(pool) {
			break
		}
	}
}

func prefixEnd(prefix []byte) []byte {
	end := make([]byte, len(prefix))
	copy(end, prefix)
	for i := len(end) - 1; i >= 0; i-- {
		end[i]++
		if end[i] != 0 {
			return end
		}
	}
	return nil
}

// GetSymbolForDenom returns the human-readable symbol for a denom.
// Returns "PNYX" for the native denom and falls back to the raw denom
// if the asset is not registered.
func (k Keeper) GetSymbolForDenom(ctx sdk.Context, denom string) string {
	if denom == pnyxDenom {
		return "PNYX"
	}
	asset, found := k.GetAssetByDenom(ctx, denom)
	if !found {
		return denom
	}
	return asset.Symbol
}

// validateAssetForTrading checks that a non-PNYX denom is registered and
// has trading enabled.
func (k Keeper) validateAssetForTrading(ctx sdk.Context, denom string) error {
	if denom == pnyxDenom {
		return nil
	}
	asset, found := k.GetAssetByDenom(ctx, denom)
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "asset not registered: %s", denom)
	}
	if !asset.TradingEnabled {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "trading disabled for %s", asset.Symbol)
	}
	return nil
}

// intSqrt computes the integer square root of n using Newton's method.
func intSqrt(n math.Int) math.Int {
	if !n.IsPositive() {
		return math.ZeroInt()
	}
	x := n
	y := x.Add(math.OneInt()).Quo(math.NewInt(2))
	for y.LT(x) {
		x = y
		y = x.Add(n.Quo(x)).Quo(math.NewInt(2))
	}
	return x
}
