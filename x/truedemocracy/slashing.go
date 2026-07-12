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

	cacheCtx, write := ctx.CacheContext()
	var err error
	val, err = k.slashValidatorStake(cacheCtx, val, SlashFractionDoubleSign)
	if err != nil {
		return err
	}
	val.Jailed = true
	val.JailedUntil = cacheCtx.BlockTime().Unix() + DowntimeJailDuration*10

	if val.Stake.AmountOf(PNYXDenom).LT(math.NewInt(rewards.StakeMin)) {
		val.Power = 0
	} else {
		val.Power = val.Stake.AmountOf(PNYXDenom).Int64() / rewards.StakeMin
	}

	k.SetValidator(cacheCtx, val)
	write()
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

	cacheCtx, write := ctx.CacheContext()
	val.MissedBlocks++

	threshold := SignedBlocksWindow - MinSignedPerWindow
	if val.MissedBlocks > threshold {
		var err error
		val, err = k.slashValidatorStake(cacheCtx, val, SlashFractionDowntime)
		if err != nil {
			return err
		}
		val.Jailed = true
		val.JailedUntil = cacheCtx.BlockTime().Unix() + DowntimeJailDuration
		val.MissedBlocks = 0

		if val.Stake.AmountOf(PNYXDenom).LT(math.NewInt(rewards.StakeMin)) {
			val.Power = 0
		} else {
			val.Power = val.Stake.AmountOf(PNYXDenom).Int64() / rewards.StakeMin
		}
	}

	k.SetValidator(cacheCtx, val)
	write()
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
	if val.Stake.AmountOf(PNYXDenom).LT(math.NewInt(rewards.StakeMin)) {
		return errorsmod.Wrap(sdkerrors.ErrInsufficientFunds, "stake below minimum after slash")
	}
	if !k.EnforceDomainMembership(ctx, operatorAddr) {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "operator no longer a domain member")
	}

	val, _ = k.GetValidator(ctx, operatorAddr) // re-read after membership check
	val.Jailed = false
	val.JailedUntil = 0
	val.Power = val.Stake.AmountOf(PNYXDenom).Int64() / rewards.StakeMin
	k.SetValidator(ctx, val)
	return nil
}

// slashStake reduces a validator's PNYX stake by the given percentage.
func slashStake(val Validator, pct int64) Validator {
	pnyxAmt := val.Stake.AmountOf(PNYXDenom)
	penalty := pnyxAmt.Mul(math.NewInt(pct)).Quo(math.NewInt(100))
	if penalty.IsZero() {
		penalty = math.OneInt() // slash at least 1
	}
	remaining := pnyxAmt.Sub(penalty)
	if remaining.IsNegative() {
		remaining = math.ZeroInt()
	}
	val.Stake = sdk.NewCoins(sdk.NewCoin(PNYXDenom, remaining))
	return val
}

// slashValidatorStake removes the slashed claim from validator stake and burns
// the same amount from module escrow. The whitepaper requires slashed PNYX to
// leave circulation; crediting it to an admin-withdrawable domain treasury
// would let a colluding validator recover the penalty.
func (k Keeper) slashValidatorStake(ctx sdk.Context, val Validator, pct int64) (Validator, error) {
	if err := requireBankKeeper(k.bankKeeper); err != nil {
		return Validator{}, err
	}

	before := val.Stake.AmountOf(PNYXDenom)
	val = slashStake(val, pct)
	penalty := before.Sub(val.Stake.AmountOf(PNYXDenom))
	if penalty.IsPositive() {
		coins := sdk.NewCoins(sdk.NewCoin(PNYXDenom, penalty))
		if err := k.bankKeeper.BurnCoins(ctx, ModuleName, coins); err != nil {
			return Validator{}, errorsmod.Wrap(err, "validator slash burn failed")
		}
	}
	return val, nil
}
