package dex

import (
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
)

const ModuleName = "dex"

// SwapFeeBps is the swap fee in basis points (0.3%).
const SwapFeeBps int64 = 30

// BurnBps is the PNYX burn rate on swaps TO PNYX in basis points (1%).
const BurnBps int64 = 100

// Pool represents an AMM liquidity pool pairing PNYX with another asset.
// The constant-product invariant x * y = k governs pricing.
type Pool struct {
	PnyxReserve  math.Int `json:"pnyx_reserve"`
	AssetReserve math.Int `json:"asset_reserve"`
	AssetDenom   string   `json:"asset_denom"`
	TotalShares  math.Int `json:"total_shares"`
	TotalBurned  math.Int `json:"total_burned"`              // cumulative PNYX burned on swaps
	AssetSymbol  string   `json:"asset_symbol,omitempty"` // display name from registry (populated in queries)
}

type GenesisState struct {
	Pools            []Pool            `json:"pools"`
	RegisteredAssets []RegisteredAsset `json:"registered_assets"`
}

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(Pool{}, "dex/Pool", nil)
	cdc.RegisterConcrete(RegisteredAsset{}, "dex/RegisteredAsset", nil)
	cdc.RegisterConcrete(GenesisState{}, "dex/GenesisState", nil)

	// Message types for CLI transactions.
	cdc.RegisterConcrete(MsgCreatePool{}, "dex/MsgCreatePool", nil)
	cdc.RegisterConcrete(MsgSwap{}, "dex/MsgSwap", nil)
	cdc.RegisterConcrete(MsgAddLiquidity{}, "dex/MsgAddLiquidity", nil)
	cdc.RegisterConcrete(MsgRemoveLiquidity{}, "dex/MsgRemoveLiquidity", nil)
	cdc.RegisterConcrete(MsgRegisterAsset{}, "dex/MsgRegisterAsset", nil)
	cdc.RegisterConcrete(MsgUpdateAssetStatus{}, "dex/MsgUpdateAssetStatus", nil)
	cdc.RegisterConcrete(MsgSwapExact{}, "dex/MsgSwapExact", nil)
}

func DefaultGenesisState() GenesisState {
	return GenesisState{
		Pools: []Pool{
			{
				PnyxReserve:  math.NewInt(1_000_000),
				AssetReserve: math.NewInt(1_000_000),
				AssetDenom:   "atom",
				TotalShares:  math.NewInt(1_000_000), // sqrt(1M * 1M) = 1M
				TotalBurned:  math.ZeroInt(),
			},
		},
		RegisteredAssets: []RegisteredAsset{
			{
				IBCDenom:       "pnyx",
				Symbol:         "PNYX",
				Name:           "TrueRepublic Native Token",
				Decimals:       6,
				OriginChain:    "truerepublic-1",
				TradingEnabled: true,
			},
			{
				IBCDenom:       "atom",
				Symbol:         "ATOM",
				Name:           "Cosmos Hub",
				Decimals:       6,
				OriginChain:    "cosmoshub-4",
				TradingEnabled: true,
			},
		},
	}
}
