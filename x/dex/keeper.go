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
	if _, exists := k.GetPool(ctx, assetDenom); exists {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "pool for %s already exists", assetDenom)
	}

	shares := intSqrt(pnyxAmt.Mul(assetAmt))

	pool := Pool{
		PnyxReserve:  pnyxAmt,
		AssetReserve: assetAmt,
		AssetDenom:   assetDenom,
		TotalShares:  shares,
	}
	k.SetPool(ctx, pool)
	return nil
}

// Swap executes a constant-product AMM swap with a 0.3% fee.
//
// The output amount is:
//
//	out = (outReserve * in * (10000 - fee)) / (inReserve * 10000 + in * (10000 - fee))
//
// One of inputDenom/outputDenom must be "pnyx" and the other the pool's
// asset denom.
func (k Keeper) Swap(ctx sdk.Context, inputDenom string, inputAmt math.Int, outputDenom string) (math.Int, error) {
	if !inputAmt.IsPositive() {
		return math.Int{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "input amount must be positive")
	}

	// Determine which denom is the asset side.
	assetDenom, err := resolveAssetDenom(inputDenom, outputDenom)
	if err != nil {
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

	// Constant-product formula with fee.
	feeMultiplier := math.NewInt(10000 - SwapFeeBps) // 9970
	numerator := outReserve.Mul(inputAmt).Mul(feeMultiplier)
	denominator := inReserve.Mul(math.NewInt(10000)).Add(inputAmt.Mul(feeMultiplier))
	outputAmt := numerator.Quo(denominator)

	if !outputAmt.IsPositive() {
		return math.Int{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "output amount is zero")
	}
	if outputAmt.GTE(outReserve) {
		return math.Int{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "swap would drain the pool")
	}

	// Update reserves.
	if inputDenom == pnyxDenom {
		pool.PnyxReserve = pool.PnyxReserve.Add(inputAmt)
		pool.AssetReserve = pool.AssetReserve.Sub(outputAmt)
	} else {
		pool.AssetReserve = pool.AssetReserve.Add(inputAmt)
		pool.PnyxReserve = pool.PnyxReserve.Sub(outputAmt)
	}

	k.SetPool(ctx, pool)
	return outputAmt, nil
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
