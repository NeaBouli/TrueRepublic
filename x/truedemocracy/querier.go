package truedemocracy

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
	QueryDomain     = "domain"
	QueryDomains    = "domains"
	QueryValidator  = "validator"
	QueryValidators = "validators"
)

// NewQuerier returns an ABCI querier for the truedemocracy module.
func NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case QueryDomain:
			return queryDomain(ctx, path[1:], k, legacyQuerierCdc)
		case QueryDomains:
			return queryAllDomains(ctx, k)
		case QueryValidator:
			return queryValidator(ctx, path[1:], k, legacyQuerierCdc)
		case QueryValidators:
			return queryAllValidators(ctx, k)
		default:
			return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "unknown query path: %s", path[0])
		}
	}
}

func queryDomain(ctx sdk.Context, path []string, k Keeper, cdc *codec.LegacyAmino) ([]byte, error) {
	if len(path) < 1 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "missing domain name")
	}
	domain, found := k.GetDomain(ctx, path[0])
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "domain %s not found", path[0])
	}
	bz, err := cdc.MarshalJSON(domain)
	if err != nil {
		return nil, err
	}
	return bz, nil
}

func queryAllDomains(ctx sdk.Context, k Keeper) ([]byte, error) {
	var domains []Domain
	k.IterateDomains(ctx, func(d Domain) bool {
		domains = append(domains, d)
		return false
	})
	if domains == nil {
		domains = []Domain{}
	}
	bz, err := json.Marshal(domains)
	if err != nil {
		return nil, err
	}
	return bz, nil
}

func queryValidator(ctx sdk.Context, path []string, k Keeper, cdc *codec.LegacyAmino) ([]byte, error) {
	if len(path) < 1 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "missing operator address")
	}
	val, found := k.GetValidator(ctx, path[0])
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "validator %s not found", path[0])
	}
	bz, err := cdc.MarshalJSON(val)
	if err != nil {
		return nil, err
	}
	return bz, nil
}

func queryAllValidators(ctx sdk.Context, k Keeper) ([]byte, error) {
	var validators []Validator
	k.IterateValidators(ctx, func(v Validator) bool {
		validators = append(validators, v)
		return false
	})
	if validators == nil {
		validators = []Validator{}
	}
	bz, err := json.Marshal(validators)
	if err != nil {
		return nil, err
	}
	return bz, nil
}
