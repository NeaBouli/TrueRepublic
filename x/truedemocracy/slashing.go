package truedemocracy

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	rewards "truerepublic/treasury/keeper"
)

// HandleDoubleSign slashes a validator for equivocation (signing conflicting
// blocks). The validator loses SlashFractionDoubleSign percent of its stake
// and is jailed for 10× the standard downtime jail duration.
func (k Keeper) HandleDoubleSign(ctx sdk.Context, pubKeyBytes []byte) error {
	val, found := k.GetValidatorByPubKey(ctx, pubKeyBytes)
	if !found {
		return errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "validator not found")
	}

	val = slashStake(val, SlashFractionDoubleSign)
	val.Jailed = true
	val.JailedUntil = ctx.BlockTime().Unix() + DowntimeJailDuration*10

	if val.Stake.AmountOf("pnyx").LT(math.NewInt(rewards.StakeMin)) {
		val.Power = 0
	} else {
		val.Power = val.Stake.AmountOf("pnyx").Int64() / rewards.StakeMin
	}

	k.SetValidator(ctx, val)
	return nil
}

// HandleDowntime increments a validator's missed-block counter and slashes
// if the validator has missed more than the allowed threshold within the
// signed-blocks window.
func (k Keeper) HandleDowntime(ctx sdk.Context, pubKeyBytes []byte) error {
	val, found := k.GetValidatorByPubKey(ctx, pubKeyBytes)
	if !found {
		return errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "validator not found")
	}

	val.MissedBlocks++

	threshold := SignedBlocksWindow - MinSignedPerWindow
	if val.MissedBlocks > threshold {
		val = slashStake(val, SlashFractionDowntime)
		val.Jailed = true
		val.JailedUntil = ctx.BlockTime().Unix() + DowntimeJailDuration
		val.MissedBlocks = 0

		if val.Stake.AmountOf("pnyx").LT(math.NewInt(rewards.StakeMin)) {
			val.Power = 0
		} else {
			val.Power = val.Stake.AmountOf("pnyx").Int64() / rewards.StakeMin
		}
	}

	k.SetValidator(ctx, val)
	return nil
}

// Unjail releases a jailed validator back to bonded status, provided the
// jail duration has passed, the stake is still above StakeMin, and the
// operator is still a domain member.
func (k Keeper) Unjail(ctx sdk.Context, operatorAddr string) error {
	val, found := k.GetValidator(ctx, operatorAddr)
	if !found {
		return errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "validator not found")
	}
	if !val.Jailed {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "validator is not jailed")
	}
	if ctx.BlockTime().Unix() < val.JailedUntil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "jail duration has not elapsed")
	}
	if val.Stake.AmountOf("pnyx").LT(math.NewInt(rewards.StakeMin)) {
		return errorsmod.Wrap(sdkerrors.ErrInsufficientFunds, "stake below minimum after slash")
	}
	if !k.EnforceDomainMembership(ctx, operatorAddr) {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "operator no longer a domain member")
	}

	val, _ = k.GetValidator(ctx, operatorAddr) // re-read after membership check
	val.Jailed = false
	val.JailedUntil = 0
	val.Power = val.Stake.AmountOf("pnyx").Int64() / rewards.StakeMin
	k.SetValidator(ctx, val)
	return nil
}

// slashStake reduces a validator's PNYX stake by the given percentage.
func slashStake(val Validator, pct int64) Validator {
	pnyxAmt := val.Stake.AmountOf("pnyx")
	penalty := pnyxAmt.Mul(math.NewInt(pct)).Quo(math.NewInt(100))
	if penalty.IsZero() {
		penalty = math.OneInt() // slash at least 1
	}
	remaining := pnyxAmt.Sub(penalty)
	if remaining.IsNegative() {
		remaining = math.ZeroInt()
	}
	val.Stake = sdk.NewCoins(sdk.NewCoin("pnyx", remaining))
	return val
}
