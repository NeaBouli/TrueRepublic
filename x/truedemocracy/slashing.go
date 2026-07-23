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
	record, found := k.GetConsensusKeyRecord(ctx, consensusAddressFromPubKey(pubKeyBytes))
	if !found {
		return errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "validator not found")
	}

	cacheCtx, write := ctx.CacheContext()
	if record.Tombstoned {
		return nil
	}
	if _, err := k.handleDoubleSignForRecord(cacheCtx, record); err != nil {
		return err
	}
	record.Tombstoned = true
	k.setConsensusKeyRecord(cacheCtx, record)
	write()
	return nil
}

// HandleDowntime is retained for direct keeper callers and tests. Production
// liveness ingestion uses recordValidatorSignature from the decided last
// commit passed to BeginBlock.
func (k Keeper) HandleDowntime(ctx sdk.Context, pubKeyBytes []byte) error {
	val, found := k.GetValidatorByPubKey(ctx, pubKeyBytes)
	if !found {
		return errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "validator not found")
	}
	nextHeight := int64(1)
	if info, exists := k.getValidatorSigningInfo(ctx, val.OperatorAddr); exists {
		nextHeight = info.LastObservedCommitHeight + 1
	}
	return k.recordValidatorSignature(ctx, val.OperatorAddr, nextHeight, true)
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
	if record, found := k.GetConsensusKeyRecord(ctx, consensusAddressFromPubKey(val.PubKey)); found && record.Tombstoned {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "tombstoned consensus key must be rotated before unjail")
	}
	val.Jailed = false
	val.JailedUntil = 0
	val.Power = val.Stake.AmountOf(PNYXDenom).Int64() / rewards.StakeMin
	k.SetValidator(ctx, val)
	k.deleteValidatorSigningInfo(ctx, operatorAddr)
	return nil
}

func validatorPowerFromStake(val Validator) int64 {
	if val.Stake.AmountOf(PNYXDenom).LT(math.NewInt(rewards.StakeMin)) {
		return 0
	}
	return val.Stake.AmountOf(PNYXDenom).Int64() / rewards.StakeMin
}

// handleDoubleSignForRecord applies the economic penalty to either an active
// validator or an evidence-window exit hold. The caller owns the tombstone and
// replay marker so the full ABCI++ batch remains atomic.
func (k Keeper) handleDoubleSignForRecord(ctx sdk.Context, record ConsensusKeyRecord) (int64, error) {
	if val, found := k.GetValidator(ctx, record.OperatorAddr); found {
		before := val.Stake.AmountOf(PNYXDenom)
		slashed, err := k.slashValidatorStake(ctx, val, SlashFractionDoubleSign)
		if err != nil {
			return 0, err
		}
		k.QueueValidatorPowerZero(ctx, val)
		slashed.Jailed = true
		slashed.JailedUntil = ctx.BlockTime().Unix() + DowntimeJailDuration*10
		slashed.Power = validatorPowerFromStake(slashed)
		k.SetValidator(ctx, slashed)
		return before.Sub(slashed.Stake.AmountOf(PNYXDenom)).Int64(), nil
	}

	removal, found := k.GetPendingValidatorRemoval(ctx, record.OperatorAddr)
	if !found {
		return 0, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "validator claim not found")
	}
	before := removal.Validator.Stake.AmountOf(PNYXDenom)
	slashed, err := k.slashValidatorStake(ctx, removal.Validator, SlashFractionDoubleSign)
	if err != nil {
		return 0, err
	}
	slashed.Jailed = true
	slashed.JailedUntil = ctx.BlockTime().Unix() + DowntimeJailDuration*10
	slashed.Power = 0
	removal.Validator = slashed
	penalty := before.Sub(slashed.Stake.AmountOf(PNYXDenom)).Int64()
	if err := k.reducePendingTransferAccounting(ctx, removal, penalty); err != nil {
		return 0, err
	}
	k.SetPendingValidatorRemoval(ctx, removal)
	return penalty, nil
}

// recordValidatorSignature advances the operator-scoped 100-block rolling
// window and applies one downtime penalty after a complete window exceeds the
// allowed miss threshold. It also covers validators already in an exit hold.
func (k Keeper) recordValidatorSignature(ctx sdk.Context, operatorAddr string, commitHeight int64, missed bool) error {
	if commitHeight < 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "commit height cannot be negative")
	}

	val, active := k.GetValidator(ctx, operatorAddr)
	removal, pending := k.GetPendingValidatorRemoval(ctx, operatorAddr)
	if !active && !pending {
		return errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "validator claim not found")
	}
	if active && (val.Jailed || val.Power <= 0) {
		return nil
	}
	if pending && removal.Validator.Jailed {
		return nil
	}

	info, found := k.getValidatorSigningInfo(ctx, operatorAddr)
	if !found {
		info = ValidatorSigningInfo{
			OperatorAddr:             operatorAddr,
			StartCommitHeight:        commitHeight,
			MissedBitmap:             make([]byte, livenessBitmapLength),
			LastObservedCommitHeight: commitHeight - 1,
		}
	}
	if err := validateSigningInfo(info); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	if commitHeight != info.LastObservedCommitHeight+1 {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"validator commit discontinuity: got %d after %d",
			commitHeight,
			info.LastObservedCommitHeight,
		)
	}

	index := info.IndexOffset % SignedBlocksWindow
	previouslyMissed := getMissedBit(info.MissedBitmap, index)
	if previouslyMissed && !missed {
		info.MissedBlocks--
	} else if !previouslyMissed && missed {
		info.MissedBlocks++
	}
	setMissedBit(info.MissedBitmap, index, missed)
	info.IndexOffset++
	info.LastObservedCommitHeight = commitHeight

	if active {
		val.MissedBlocks = info.MissedBlocks
		k.SetValidator(ctx, val)
	} else {
		removal.Validator.MissedBlocks = info.MissedBlocks
		k.SetPendingValidatorRemoval(ctx, removal)
	}

	threshold := SignedBlocksWindow - MinSignedPerWindow
	if info.IndexOffset < SignedBlocksWindow || info.MissedBlocks <= threshold {
		k.setValidatorSigningInfo(ctx, info)
		return nil
	}

	if active {
		slashed, err := k.slashValidatorStake(ctx, val, SlashFractionDowntime)
		if err != nil {
			return err
		}
		k.QueueValidatorPowerZero(ctx, val)
		slashed.Jailed = true
		slashed.JailedUntil = ctx.BlockTime().Unix() + DowntimeJailDuration
		slashed.MissedBlocks = 0
		slashed.Power = validatorPowerFromStake(slashed)
		k.SetValidator(ctx, slashed)
	} else {
		before := removal.Validator.Stake.AmountOf(PNYXDenom)
		slashed, err := k.slashValidatorStake(ctx, removal.Validator, SlashFractionDowntime)
		if err != nil {
			return err
		}
		slashed.Jailed = true
		slashed.JailedUntil = ctx.BlockTime().Unix() + DowntimeJailDuration
		slashed.MissedBlocks = 0
		slashed.Power = 0
		removal.Validator = slashed
		penalty := before.Sub(slashed.Stake.AmountOf(PNYXDenom)).Int64()
		if err := k.reducePendingTransferAccounting(ctx, removal, penalty); err != nil {
			return err
		}
		k.SetPendingValidatorRemoval(ctx, removal)
	}
	k.resetValidatorSigningInfo(ctx, operatorAddr, commitHeight)
	return nil
}

func (k Keeper) reducePendingTransferAccounting(ctx sdk.Context, removal PendingValidatorRemoval, penalty int64) error {
	if penalty <= 0 || len(removal.Validator.Domains) == 0 {
		return nil
	}
	domainName := removal.Validator.Domains[0]
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "pending validator domain %s not found", domainName)
	}
	if domain.TransferredStake < penalty {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "pending validator transfer accounting underflows")
	}
	domain.TransferredStake -= penalty
	ctx.KVStore(k.StoreKey).Set([]byte("domain:"+domainName), k.cdc.MustMarshalLengthPrefixed(&domain))
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
		if err := k.issuer.Burn(ctx, penalty); err != nil {
			return Validator{}, errorsmod.Wrap(err, "validator slash burn failed")
		}
	}
	return val, nil
}
