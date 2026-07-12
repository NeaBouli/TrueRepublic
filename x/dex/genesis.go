package dex

import (
	"fmt"
	"sort"
	"strings"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValidateGenesisState validates DEX structure and LP ownership. Bank reserve
// backing is validated at the application boundary where x/bank genesis is
// available.
func ValidateGenesisState(genesis GenesisState) error {
	assets := make(map[string]RegisteredAsset, len(genesis.RegisteredAssets))
	symbols := make(map[string]string, len(genesis.RegisteredAssets))
	for _, asset := range genesis.RegisteredAssets {
		if err := asset.ValidateBasic(); err != nil {
			return fmt.Errorf("invalid registered asset %q: %w", asset.IBCDenom, err)
		}
		if _, exists := assets[asset.IBCDenom]; exists {
			return fmt.Errorf("duplicate registered asset denom %q", asset.IBCDenom)
		}
		symbol := strings.ToUpper(asset.Symbol)
		if denom, exists := symbols[symbol]; exists {
			return fmt.Errorf("duplicate registered asset symbol %q for %q and %q", symbol, denom, asset.IBCDenom)
		}
		assets[asset.IBCDenom] = asset
		symbols[symbol] = asset.IBCDenom
	}

	pools := make(map[string]Pool, len(genesis.Pools))
	for _, pool := range genesis.Pools {
		if err := validateGenesisPool(pool, assets); err != nil {
			return err
		}
		if _, exists := pools[pool.AssetDenom]; exists {
			return fmt.Errorf("duplicate pool for %q", pool.AssetDenom)
		}
		pools[pool.AssetDenom] = pool
	}

	totals := make(map[string]math.Int, len(pools))
	positions := make(map[string]struct{}, len(genesis.LPPositions))
	for _, position := range genesis.LPPositions {
		if _, found := pools[position.AssetDenom]; !found {
			return fmt.Errorf("LP position references missing pool %q", position.AssetDenom)
		}
		if _, err := sdk.AccAddressFromBech32(position.Provider); err != nil {
			return fmt.Errorf("invalid LP provider for %q: %w", position.AssetDenom, err)
		}
		if position.Shares.IsNil() || !position.Shares.IsPositive() {
			return fmt.Errorf("LP position for %q/%q must have positive shares", position.AssetDenom, position.Provider)
		}
		key := position.AssetDenom + "\x00" + position.Provider
		if _, exists := positions[key]; exists {
			return fmt.Errorf("duplicate LP position for %q/%q", position.AssetDenom, position.Provider)
		}
		positions[key] = struct{}{}
		current, found := totals[position.AssetDenom]
		if !found {
			current = math.ZeroInt()
		}
		totals[position.AssetDenom] = current.Add(position.Shares)
	}

	denoms := make([]string, 0, len(pools))
	for denom := range pools {
		denoms = append(denoms, denom)
	}
	sort.Strings(denoms)
	for _, denom := range denoms {
		total, found := totals[denom]
		if !found {
			total = math.ZeroInt()
		}
		if !total.Equal(pools[denom].TotalShares) {
			return fmt.Errorf("LP shares for %q total %s, want %s", denom, total, pools[denom].TotalShares)
		}
	}
	return nil
}

// GenesisReserveClaims returns the exact bank coins required to back every
// declared pool reserve.
func GenesisReserveClaims(genesis GenesisState) (sdk.Coins, error) {
	if err := ValidateGenesisState(genesis); err != nil {
		return nil, err
	}
	claims := sdk.NewCoins()
	for _, pool := range genesis.Pools {
		claims = claims.Add(sdk.NewCoin(pnyxDenom, pool.PnyxReserve))
		claims = claims.Add(sdk.NewCoin(pool.AssetDenom, pool.AssetReserve))
	}
	return claims, nil
}

func validateGenesisPool(pool Pool, assets map[string]RegisteredAsset) error {
	if err := sdk.ValidateDenom(pool.AssetDenom); err != nil {
		return fmt.Errorf("invalid pool asset denom %q: %w", pool.AssetDenom, err)
	}
	if pool.AssetDenom == pnyxDenom {
		return fmt.Errorf("pool asset must differ from %s", pnyxDenom)
	}
	asset, found := assets[pool.AssetDenom]
	if !found {
		return fmt.Errorf("pool asset %q is not registered", pool.AssetDenom)
	}
	if !asset.TradingEnabled {
		return fmt.Errorf("pool asset %q is not enabled for trading", pool.AssetDenom)
	}
	if pool.PnyxReserve.IsNil() || !pool.PnyxReserve.IsPositive() ||
		pool.AssetReserve.IsNil() || !pool.AssetReserve.IsPositive() ||
		pool.TotalShares.IsNil() || !pool.TotalShares.IsPositive() {
		return fmt.Errorf("pool %q reserves and total shares must be positive", pool.AssetDenom)
	}
	if pool.TotalBurned.IsNil() || pool.TotalBurned.IsNegative() {
		return fmt.Errorf("pool %q total burned cannot be negative", pool.AssetDenom)
	}
	if pool.TotalVolumePnyx.IsNil() || pool.TotalVolumePnyx.IsNegative() {
		return fmt.Errorf("pool %q total volume cannot be negative", pool.AssetDenom)
	}
	if pool.SwapCount < 0 {
		return fmt.Errorf("pool %q swap count cannot be negative", pool.AssetDenom)
	}
	return nil
}
