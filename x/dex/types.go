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
	TotalBurned  math.Int `json:"total_burned"` // cumulative PNYX burned on swaps
}

type GenesisState struct {
	Pools []Pool `json:"pools"`
}

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(Pool{}, "dex/Pool", nil)
	cdc.RegisterConcrete(GenesisState{}, "dex/GenesisState", nil)

	// Message types for CLI transactions.
	cdc.RegisterConcrete(MsgCreatePool{}, "dex/MsgCreatePool", nil)
	cdc.RegisterConcrete(MsgSwap{}, "dex/MsgSwap", nil)
	cdc.RegisterConcrete(MsgAddLiquidity{}, "dex/MsgAddLiquidity", nil)
	cdc.RegisterConcrete(MsgRemoveLiquidity{}, "dex/MsgRemoveLiquidity", nil)
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
	}
}
