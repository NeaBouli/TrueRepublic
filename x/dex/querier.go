package dex

import (
	"encoding/json"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/cometbft/cometbft/abci/types"
)

// Query route constants.
const (
	QueryPool  = "pool"
	QueryPools = "pools"
)

// NewQuerier returns an ABCI querier for the dex module.
func NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case QueryPool:
			return queryPool(ctx, path[1:], k, legacyQuerierCdc)
		case QueryPools:
			return queryAllPools(ctx, k)
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
