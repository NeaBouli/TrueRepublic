package truedemocracy

import (
	"math"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const pendingValidatorRemovalPrefix = "pending-validator-removal:"

func pendingValidatorRemovalKey(operatorAddr string) []byte {
	return []byte(pendingValidatorRemovalPrefix + operatorAddr)
}

// GetPendingValidatorRemoval returns the evidence-window escrow hold for an
// operator that has fully exited the validator set.
func (k Keeper) GetPendingValidatorRemoval(ctx sdk.Context, operatorAddr string) (PendingValidatorRemoval, bool) {
	bz := ctx.KVStore(k.StoreKey).Get(pendingValidatorRemovalKey(operatorAddr))
	if bz == nil {
		return PendingValidatorRemoval{}, false
	}
	var removal PendingValidatorRemoval
	k.cdc.MustUnmarshalLengthPrefixed(bz, &removal)
	return removal, true
}

// SetPendingValidatorRemoval persists an exit hold under the immutable
// operator identity carried by its validator snapshot.
func (k Keeper) SetPendingValidatorRemoval(ctx sdk.Context, removal PendingValidatorRemoval) {
	bz := k.cdc.MustMarshalLengthPrefixed(&removal)
	ctx.KVStore(k.StoreKey).Set(pendingValidatorRemovalKey(removal.Validator.OperatorAddr), bz)
}

func (k Keeper) DeletePendingValidatorRemoval(ctx sdk.Context, operatorAddr string) {
	ctx.KVStore(k.StoreKey).Delete(pendingValidatorRemovalKey(operatorAddr))
}

// IteratePendingValidatorRemovals visits exit holds in deterministic operator
// key order. Returning true stops iteration.
func (k Keeper) IteratePendingValidatorRemovals(ctx sdk.Context, fn func(PendingValidatorRemoval) bool) {
	store := ctx.KVStore(k.StoreKey)
	prefix := []byte(pendingValidatorRemovalPrefix)
	iter := store.Iterator(prefix, prefixEnd(prefix))
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var removal PendingValidatorRemoval
		k.cdc.MustUnmarshalLengthPrefixed(iter.Value(), &removal)
		if fn(removal) {
			return
		}
	}
}

func validatorRetirementHeight(removalHeight, maxAgeNumBlocks int64) (int64, int64, error) {
	if removalHeight < 0 {
		return 0, 0, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "validator removal height cannot be negative")
	}
	if maxAgeNumBlocks <= 0 {
		return 0, 0, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "consensus evidence max-age blocks must be positive")
	}

	// An update returned after block H is first reflected in the validator set
	// at H + 1 + ValidatorUpdateDelay. The preceding block is therefore the
	// last block the removed key can sign.
	retirementDelay := sdk.ValidatorUpdateDelay + 1
	if removalHeight > math.MaxInt64-retirementDelay {
		return 0, 0, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "validator retirement height overflows")
	}
	retirementHeight := removalHeight + retirementDelay
	// The key's final possible signing height is retirementHeight-1.
	// Evidence expires only once currentHeight-finalSigningHeight is strictly
	// greater than MaxAgeNumBlocks.
	releaseOffset := maxAgeNumBlocks - 1
	if retirementHeight > math.MaxInt64-releaseOffset {
		return 0, 0, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "validator release height overflows")
	}
	return retirementHeight, retirementHeight + releaseOffset, nil
}

// newPendingValidatorRemoval snapshots the complete validator claim and the
// consensus evidence limits at exit time. The time-based boundary is
// deliberately unset until the retirement height is actually observed.
func newPendingValidatorRemoval(ctx sdk.Context, validator Validator, recipientAddr string) (PendingValidatorRemoval, error) {
	evidence := ctx.ConsensusParams().Evidence
	if evidence == nil {
		return PendingValidatorRemoval{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "consensus evidence parameters are unavailable")
	}
	if evidence.MaxAgeDuration <= 0 {
		return PendingValidatorRemoval{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "consensus evidence max-age duration must be positive")
	}
	retiredHeight, releaseHeight, err := validatorRetirementHeight(ctx.BlockHeight(), evidence.MaxAgeNumBlocks)
	if err != nil {
		return PendingValidatorRemoval{}, err
	}
	return PendingValidatorRemoval{
		Validator:              validator,
		RecipientAddr:          recipientAddr,
		RemovedAtHeight:        ctx.BlockHeight(),
		RemovedAtTimeNanos:     ctx.BlockTime().UnixNano(),
		ConsensusRetiredHeight: retiredHeight,
		ReleaseAfterHeight:     releaseHeight,
	}, nil
}

func observeValidatorRetirement(ctx sdk.Context, removal PendingValidatorRemoval) (PendingValidatorRemoval, error) {
	if removal.ConsensusRetiredAtNanos != 0 || ctx.BlockHeight() < removal.ConsensusRetiredHeight {
		return removal, nil
	}
	evidence := ctx.ConsensusParams().Evidence
	if evidence == nil {
		return PendingValidatorRemoval{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "consensus evidence parameters are unavailable")
	}
	if evidence.MaxAgeDuration <= 0 {
		return PendingValidatorRemoval{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "consensus evidence max-age duration must be positive")
	}
	retiredAt := ctx.BlockTime().UnixNano()
	if retiredAt > math.MaxInt64-int64(evidence.MaxAgeDuration) {
		return PendingValidatorRemoval{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "validator release time overflows")
	}
	removal.ConsensusRetiredAtNanos = retiredAt
	removal.ReleaseAfterTimeNanos = retiredAt + int64(evidence.MaxAgeDuration)
	return removal, nil
}

// ProcessPendingValidatorRemovals observes the real retirement block and
// releases mature holds. A hold is paid only after both consensus evidence
// limits have been strictly exceeded.
func (k Keeper) ProcessPendingValidatorRemovals(ctx sdk.Context) error {
	if err := requireBankKeeper(k.bankKeeper); err != nil {
		return err
	}

	cacheCtx, write := ctx.CacheContext()
	var removals []PendingValidatorRemoval
	k.IteratePendingValidatorRemovals(cacheCtx, func(removal PendingValidatorRemoval) bool {
		removals = append(removals, removal)
		return false
	})

	for _, stored := range removals {
		removal, err := observeValidatorRetirement(cacheCtx, stored)
		if err != nil {
			return err
		}
		if removal.ConsensusRetiredAtNanos == 0 ||
			cacheCtx.BlockHeight() <= removal.ReleaseAfterHeight ||
			cacheCtx.BlockTime().UnixNano() <= removal.ReleaseAfterTimeNanos {
			if removal.ConsensusRetiredAtNanos != stored.ConsensusRetiredAtNanos {
				k.SetPendingValidatorRemoval(cacheCtx, removal)
			}
			continue
		}

		recipient, err := sdk.AccAddressFromBech32(removal.RecipientAddr)
		if err != nil {
			return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "pending validator removal recipient is invalid")
		}
		stake := removal.Validator.Stake.AmountOf(PNYXDenom)
		if !stake.IsPositive() {
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "pending validator removal stake is not positive")
		}
		coins := sdk.NewCoins(sdk.NewCoin(PNYXDenom, stake))
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(cacheCtx, ModuleName, recipient, coins); err != nil {
			return errorsmod.Wrap(err, "pending validator removal payout failed")
		}
		k.deleteValidatorSigningInfo(cacheCtx, removal.Validator.OperatorAddr)
		k.DeletePendingValidatorRemoval(cacheCtx, removal.Validator.OperatorAddr)
	}

	write()
	return nil
}
