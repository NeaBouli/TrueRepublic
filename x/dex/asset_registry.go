package dex

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Storage key prefix for asset registry entries.
const assetRegistryPrefix = "asset:"

func assetRegistryKey(ibcDenom string) []byte {
	return []byte(assetRegistryPrefix + ibcDenom)
}

// RegisteredAsset represents a whitelisted trading asset on the DEX.
type RegisteredAsset struct {
	// IBC denom (e.g., "ibc/27394FB..." or "pnyx" for native)
	IBCDenom string `json:"ibc_denom"`
	// Display symbol (e.g., "BTC", "ETH", "LUSD")
	Symbol string `json:"symbol"`
	// Full name (e.g., "Bitcoin", "Ethereum")
	Name string `json:"name"`
	// Decimal places for display (e.g., 8 for BTC, 18 for ETH, 6 for PNYX)
	Decimals uint32 `json:"decimals"`
	// Origin chain ID (e.g., "cosmoshub-4", "ethereum")
	OriginChain string `json:"origin_chain"`
	// IBC channel used for this asset (e.g., "channel-0")
	IBCChannel string `json:"ibc_channel"`
	// Whether trading is currently enabled
	TradingEnabled bool `json:"trading_enabled"`
	// Block height at which asset was registered
	RegisteredHeight int64 `json:"registered_height"`
	// Address of the admin who registered the asset
	RegisteredBy string `json:"registered_by"`
}

// ValidateBasic performs stateless validation of a RegisteredAsset.
func (a RegisteredAsset) ValidateBasic() error {
	if a.IBCDenom == "" {
		return fmt.Errorf("ibc_denom cannot be empty")
	}
	if a.Symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}
	if a.Decimals > 18 {
		return fmt.Errorf("decimals cannot exceed 18, got %d", a.Decimals)
	}
	return nil
}

// RegisterAsset adds a new asset to the registry.
func (k Keeper) RegisterAsset(ctx sdk.Context, asset RegisteredAsset) error {
	if err := asset.ValidateBasic(); err != nil {
		return err
	}

	if _, exists := k.GetAssetByDenom(ctx, asset.IBCDenom); exists {
		return fmt.Errorf("asset already registered: %s", asset.IBCDenom)
	}

	store := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&asset)
	store.Set(assetRegistryKey(asset.IBCDenom), bz)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"asset_registered",
		sdk.NewAttribute("ibc_denom", asset.IBCDenom),
		sdk.NewAttribute("symbol", asset.Symbol),
		sdk.NewAttribute("origin_chain", asset.OriginChain),
	))

	return nil
}

// GetAssetByDenom retrieves a registered asset by its IBC denom.
func (k Keeper) GetAssetByDenom(ctx sdk.Context, ibcDenom string) (RegisteredAsset, bool) {
	store := ctx.KVStore(k.StoreKey)
	bz := store.Get(assetRegistryKey(ibcDenom))
	if bz == nil {
		return RegisteredAsset{}, false
	}

	var asset RegisteredAsset
	k.cdc.MustUnmarshalLengthPrefixed(bz, &asset)
	return asset, true
}

// GetAssetBySymbol retrieves a registered asset by its display symbol.
func (k Keeper) GetAssetBySymbol(ctx sdk.Context, symbol string) (RegisteredAsset, bool) {
	allAssets := k.GetAllAssets(ctx)
	for _, asset := range allAssets {
		if asset.Symbol == symbol {
			return asset, true
		}
	}
	return RegisteredAsset{}, false
}

// GetAllAssets returns all registered assets.
func (k Keeper) GetAllAssets(ctx sdk.Context) []RegisteredAsset {
	store := ctx.KVStore(k.StoreKey)
	prefix := []byte(assetRegistryPrefix)
	iter := store.Iterator(prefix, prefixEnd(prefix))
	defer iter.Close()

	var assets []RegisteredAsset
	for ; iter.Valid(); iter.Next() {
		var asset RegisteredAsset
		k.cdc.MustUnmarshalLengthPrefixed(iter.Value(), &asset)
		assets = append(assets, asset)
	}
	return assets
}

// UpdateAssetTradingStatus enables or disables trading for a registered asset.
func (k Keeper) UpdateAssetTradingStatus(ctx sdk.Context, ibcDenom string, enabled bool) error {
	asset, exists := k.GetAssetByDenom(ctx, ibcDenom)
	if !exists {
		return fmt.Errorf("asset not found: %s", ibcDenom)
	}

	asset.TradingEnabled = enabled

	store := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&asset)
	store.Set(assetRegistryKey(ibcDenom), bz)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"asset_trading_status_updated",
		sdk.NewAttribute("ibc_denom", ibcDenom),
		sdk.NewAttribute("enabled", fmt.Sprintf("%t", enabled)),
	))

	return nil
}

// DeregisterAsset removes an asset from the registry.
func (k Keeper) DeregisterAsset(ctx sdk.Context, ibcDenom string) error {
	if _, exists := k.GetAssetByDenom(ctx, ibcDenom); !exists {
		return fmt.Errorf("asset not found: %s", ibcDenom)
	}

	store := ctx.KVStore(k.StoreKey)
	store.Delete(assetRegistryKey(ibcDenom))

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"asset_deregistered",
		sdk.NewAttribute("ibc_denom", ibcDenom),
	))

	return nil
}
