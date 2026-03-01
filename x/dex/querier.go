package dex

import (
	"encoding/json"
	"strconv"

	"cosmossdk.io/math"
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/cometbft/cometbft/abci/types"
)

// Query route constants.
const (
	QueryPool           = "pool"
	QueryPools          = "pools"
	QueryPoolStats      = "pool_stats"
	QuerySpotPrice      = "spot_price"
	QueryLiquidityDepth = "liquidity_depth"
	QueryLPPosition     = "lp_position"
	QueryEstimateSwap   = "estimate_swap"
)

// NewQuerier returns an ABCI querier for the dex module.
func NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case QueryPool:
			return queryPool(ctx, path[1:], k, legacyQuerierCdc)
		case QueryPools:
			return queryAllPools(ctx, k)
		case QueryPoolStats:
			return queryPoolStats(ctx, path[1:], k)
		case QuerySpotPrice:
			return querySpotPrice(ctx, path[1:], k)
		case QueryLiquidityDepth:
			return queryLiquidityDepth(ctx, path[1:], k)
		case QueryLPPosition:
			return queryLPPosition(ctx, path[1:], k)
		case QueryEstimateSwap:
			return queryEstimateSwap(ctx, path[1:], k)
		default:
			return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "unknown query path: %s", path[0])
		}
	}
}

func queryPool(ctx sdk.Context, path []string, k Keeper, cdc *codec.LegacyAmino) ([]byte, error) {
	if len(path) < 1 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "missing asset denom")
	}
	pool, found := k.GetPool(ctx, path[0])
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "pool for %s not found", path[0])
	}
	bz, err := cdc.MarshalJSON(pool)
	if err != nil {
		return nil, err
	}
	return bz, nil
}

func queryAllPools(ctx sdk.Context, k Keeper) ([]byte, error) {
	var pools []Pool
	k.IteratePools(ctx, func(p Pool) bool {
		pools = append(pools, p)
		return false
	})
	if pools == nil {
		pools = []Pool{}
	}
	bz, err := json.Marshal(pools)
	if err != nil {
		return nil, err
	}
	return bz, nil
}

// queryPoolStats returns aggregated pool statistics.
// Path: pool_stats/{assetDenom}
func queryPoolStats(ctx sdk.Context, path []string, k Keeper) ([]byte, error) {
	if len(path) < 1 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "missing asset denom")
	}
	pool, found := k.GetPool(ctx, path[0])
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "pool for %s not found", path[0])
	}

	totalFeesEarned := pool.TotalVolumePnyx.Mul(math.NewInt(SwapFeeBps)).Quo(math.NewInt(10000))
	spotPrice, _ := k.ComputeSpotPrice(ctx, pnyxDenom, path[0])

	result := struct {
		AssetDenom          string `json:"asset_denom"`
		AssetSymbol         string `json:"asset_symbol"`
		SwapCount           int64  `json:"swap_count"`
		TotalVolumePnyx     string `json:"total_volume_pnyx"`
		TotalFeesEarned     string `json:"total_fees_earned"`
		TotalBurned         string `json:"total_burned"`
		PnyxReserve         string `json:"pnyx_reserve"`
		AssetReserve        string `json:"asset_reserve"`
		SpotPricePerMillion string `json:"spot_price_per_million"`
		TotalShares         string `json:"total_shares"`
	}{
		AssetDenom:          pool.AssetDenom,
		AssetSymbol:         k.GetSymbolForDenom(ctx, pool.AssetDenom),
		SwapCount:           pool.SwapCount,
		TotalVolumePnyx:     pool.TotalVolumePnyx.String(),
		TotalFeesEarned:     totalFeesEarned.String(),
		TotalBurned:         pool.TotalBurned.String(),
		PnyxReserve:         pool.PnyxReserve.String(),
		AssetReserve:        pool.AssetReserve.String(),
		SpotPricePerMillion: spotPrice.String(),
		TotalShares:         pool.TotalShares.String(),
	}
	return json.Marshal(result)
}

// querySpotPrice returns the instantaneous price between two denoms.
// Path: spot_price/{inputDenom}/{outputDenom}
func querySpotPrice(ctx sdk.Context, path []string, k Keeper) ([]byte, error) {
	if len(path) < 2 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "usage: spot_price/{inputDenom}/{outputDenom}")
	}
	inputDenom, outputDenom := path[0], path[1]

	price, err := k.ComputeSpotPrice(ctx, inputDenom, outputDenom)
	if err != nil {
		return nil, err
	}

	var route []string
	if inputDenom == pnyxDenom || outputDenom == pnyxDenom {
		route = []string{inputDenom, outputDenom}
	} else {
		route = []string{inputDenom, pnyxDenom, outputDenom}
	}

	result := struct {
		InputDenom      string   `json:"input_denom"`
		OutputDenom     string   `json:"output_denom"`
		PricePerMillion string   `json:"price_per_million"`
		InputSymbol     string   `json:"input_symbol"`
		OutputSymbol    string   `json:"output_symbol"`
		Route           []string `json:"route"`
	}{
		InputDenom:      inputDenom,
		OutputDenom:     outputDenom,
		PricePerMillion: price.String(),
		InputSymbol:     k.GetSymbolForDenom(ctx, inputDenom),
		OutputSymbol:    k.GetSymbolForDenom(ctx, outputDenom),
		Route:           route,
	}
	return json.Marshal(result)
}

// queryLiquidityDepth returns the slippage curve.
// Path: liquidity_depth/{inputDenom}/{outputDenom}
func queryLiquidityDepth(ctx sdk.Context, path []string, k Keeper) ([]byte, error) {
	if len(path) < 2 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "usage: liquidity_depth/{inputDenom}/{outputDenom}")
	}
	levels, err := k.ComputeLiquidityDepth(ctx, path[0], path[1])
	if err != nil {
		return nil, err
	}

	result := struct {
		InputDenom  string       `json:"input_denom"`
		OutputDenom string       `json:"output_denom"`
		Levels      []DepthLevel `json:"levels"`
	}{
		InputDenom:  path[0],
		OutputDenom: path[1],
		Levels:      levels,
	}
	return json.Marshal(result)
}

// queryLPPosition returns the underlying token values for LP shares.
// Path: lp_position/{assetDenom}/{shares}
func queryLPPosition(ctx sdk.Context, path []string, k Keeper) ([]byte, error) {
	if len(path) < 2 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "usage: lp_position/{assetDenom}/{shares}")
	}
	shares, err := strconv.ParseInt(path[1], 10, 64)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid shares: %s", path[1])
	}

	pnyxVal, assetVal, shareBps, err := k.ComputeLPPosition(ctx, path[0], math.NewInt(shares))
	if err != nil {
		return nil, err
	}

	result := struct {
		AssetDenom     string `json:"asset_denom"`
		Shares         string `json:"shares"`
		PnyxValue      string `json:"pnyx_value"`
		AssetValue     string `json:"asset_value"`
		ShareOfPoolBps int64  `json:"share_of_pool_bps"`
	}{
		AssetDenom:     path[0],
		Shares:         strconv.FormatInt(shares, 10),
		PnyxValue:      pnyxVal.String(),
		AssetValue:     assetVal.String(),
		ShareOfPoolBps: shareBps,
	}
	return json.Marshal(result)
}

// queryEstimateSwap returns the expected output for a swap.
// Path: estimate_swap/{inputDenom}/{amount}/{outputDenom}
func queryEstimateSwap(ctx sdk.Context, path []string, k Keeper) ([]byte, error) {
	if len(path) < 3 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "usage: estimate_swap/{inputDenom}/{amount}/{outputDenom}")
	}
	inputDenom, outputDenom := path[0], path[2]
	amt, err := strconv.ParseInt(path[1], 10, 64)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid amount: %s", path[1])
	}

	expectedOutput, route, err := k.EstimateSwapOutput(ctx, inputDenom, math.NewInt(amt), outputDenom)
	if err != nil {
		return nil, err
	}

	routeSymbols := make([]string, len(route))
	for i, d := range route {
		routeSymbols[i] = k.GetSymbolForDenom(ctx, d)
	}

	result := struct {
		ExpectedOutput string   `json:"expected_output"`
		Route          []string `json:"route"`
		RouteSymbols   []string `json:"route_symbols"`
		Hops           int      `json:"hops"`
	}{
		ExpectedOutput: expectedOutput.String(),
		Route:          route,
		RouteSymbols:   routeSymbols,
		Hops:           len(route) - 1,
	}
	return json.Marshal(result)
}
