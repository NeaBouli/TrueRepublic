package dex

import (
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"

	"truerepublic/token"
)

const ModuleName = "dex"
const pnyxDenom = token.BaseDenom

// SwapFeeBps is the swap fee in basis points (0.3%).
const SwapFeeBps int64 = 30

// BurnBps is the PNYX burn rate on swaps TO PNYX in basis points (1%).
const BurnBps int64 = 100

// Pool represents an AMM liquidity pool pairing PNYX with another asset.
// The constant-product invariant x * y = k governs pricing.
type Pool struct {
	PnyxReserve     math.Int `json:"pnyx_reserve"`
	AssetReserve    math.Int `json:"asset_reserve"`
	AssetDenom      string   `json:"asset_denom"`
	TotalShares     math.Int `json:"total_shares"`
	TotalBurned     math.Int `json:"total_burned"`           // cumulative PNYX burned on swaps
	AssetSymbol     string   `json:"asset_symbol,omitempty"` // display name from registry (populated in queries)
	SwapCount       int64    `json:"swap_count"`             // cumulative swap count
	TotalVolumePnyx math.Int `json:"total_volume_pnyx"`      // cumulative PNYX volume
}

type GenesisState struct {
	Pools            []Pool            `json:"pools"`
	RegisteredAssets []RegisteredAsset `json:"registered_assets"`
	LPPositions      []LPPosition      `json:"lp_positions"`
}

// LPPosition is the exportable ownership record for one provider in one pool.
type LPPosition struct {
	AssetDenom string   `json:"asset_denom"`
	Provider   string   `json:"provider"`
	Shares     math.Int `json:"shares"`
}

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(Pool{}, "dex/Pool", nil)
	cdc.RegisterConcrete(RegisteredAsset{}, "dex/RegisteredAsset", nil)
	cdc.RegisterConcrete(LPPosition{}, "dex/LPPosition", nil)
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
		Pools: []Pool{},
		RegisteredAssets: []RegisteredAsset{
			{
				IBCDenom:       pnyxDenom,
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
		LPPositions: []LPPosition{},
	}
}
