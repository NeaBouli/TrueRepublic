package truedemocracy

import (
	"encoding/hex"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	cryptoproto "github.com/cometbft/cometbft/proto/tendermint/crypto"

	rewards "truerepublic/treasury/keeper"
)

// KV store key helpers for validator state.
func validatorKey(operatorAddr string) []byte {
	return []byte("validator:" + operatorAddr)
}

func valPubKeyKey(pubKeyBytes []byte) []byte {
	return []byte("val-pubkey:" + hex.EncodeToString(pubKeyBytes))
}

// RegisterValidator registers a new Proof of Domain validator.
// The operator must be a member of the given domain and stake >= StakeMin PNYX.
func (k Keeper) RegisterValidator(ctx sdk.Context, operatorAddr string, pubKeyBytes []byte, stake sdk.Coins, domainName string) error {
	if len(pubKeyBytes) != 32 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "pubkey must be 32 bytes (ed25519)")
	}

	pnyxAmt := stake.AmountOf("pnyx")
	if pnyxAmt.LT(math.NewInt(rewards.StakeMin)) {
		return errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, "stake %s below minimum %d", pnyxAmt, rewards.StakeMin)
	}

	// Verify the operator is a member of the domain.
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "domain not found")
	}
	isMember := false
	for _, m := range domain.Members {
		if m == operatorAddr {
			isMember = true
			break
		}
	}
	if !isMember {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "operator is not a member of the domain")
	}

	// Check for duplicate registration.
	store := ctx.KVStore(k.StoreKey)
	if store.Has(validatorKey(operatorAddr)) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "validator already registered")
	}

	power := pnyxAmt.Int64() / rewards.StakeMin

	val := Validator{
		OperatorAddr: operatorAddr,
		PubKey:       pubKeyBytes,
		Stake:        stake,
		Domains:      []string{domainName},
		Power:        power,
		Jailed:       false,
		JailedUntil:  0,
		MissedBlocks: 0,
	}

	valBz := k.cdc.MustMarshalLengthPrefixed(&val)
	store.Set(validatorKey(operatorAddr), valBz)
	store.Set(valPubKeyKey(pubKeyBytes), []byte(operatorAddr))

	return nil
}

// GetValidator loads a validator from the store by operator address.
func (k Keeper) GetValidator(ctx sdk.Context, operatorAddr string) (Validator, bool) {
	store := ctx.KVStore(k.StoreKey)
	bz := store.Get(validatorKey(operatorAddr))
	if bz == nil {
		return Validator{}, false
	}
	var val Validator
	k.cdc.MustUnmarshalLengthPrefixed(bz, &val)
	return val, true
}

// SetValidator persists a validator to the store.
func (k Keeper) SetValidator(ctx sdk.Context, val Validator) {
	store := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&val)
	store.Set(validatorKey(val.OperatorAddr), bz)
}

// GetValidatorByPubKey looks up a validator via the reverse pubkey index.
func (k Keeper) GetValidatorByPubKey(ctx sdk.Context, pubKeyBytes []byte) (Validator, bool) {
	store := ctx.KVStore(k.StoreKey)
	addrBz := store.Get(valPubKeyKey(pubKeyBytes))
	if addrBz == nil {
		return Validator{}, false
	}
	return k.GetValidator(ctx, string(addrBz))
}

// WithdrawStake allows a validator to withdraw some or all of their staked
// PNYX. The withdrawal is subject to the PoD transfer limit (WP §7):
// cumulative withdrawals across all domain validators cannot exceed 10% of
// the domain's total historical payouts.
func (k Keeper) WithdrawStake(ctx sdk.Context, operatorAddr string, amount int64) error {
	val, found := k.GetValidator(ctx, operatorAddr)
	if !found {
		return errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "validator not found")
	}

	stakeAmt := val.Stake.AmountOf("pnyx").Int64()
	if amount > stakeAmt {
		return errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds,
			"withdraw %d exceeds current stake %d", amount, stakeAmt)
	}

	// Check transfer limit against the validator's primary domain.
	if len(val.Domains) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "validator has no domains")
	}
	domainName := val.Domains[0]

	if err := k.ValidateStakeTransfer(ctx, domainName, operatorAddr, amount); err != nil {
		return err
	}

	// Update domain's cumulative transferred stake.
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "domain %s not found", domainName)
	}
	domain.TransferredStake += amount
	store := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	store.Set([]byte("domain:"+domainName), bz)

	// Reduce validator stake.
	newStake := stakeAmt - amount
	if newStake < rewards.StakeMin {
		// Stake drops below minimum — remove the validator entirely.
		return k.RemoveValidator(ctx, operatorAddr)
	}

	val.Stake = sdk.NewCoins(sdk.NewInt64Coin("pnyx", newStake))
	val.Power = newStake / rewards.StakeMin
	k.SetValidator(ctx, val)
	return nil
}

// RemoveValidator deletes a validator and its reverse index.
func (k Keeper) RemoveValidator(ctx sdk.Context, operatorAddr string) error {
	val, found := k.GetValidator(ctx, operatorAddr)
	if !found {
		return errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "validator not found")
	}
	store := ctx.KVStore(k.StoreKey)
	store.Delete(validatorKey(operatorAddr))
	store.Delete(valPubKeyKey(val.PubKey))
	return nil
}

// IterateValidators calls fn for each validator in the store.
// If fn returns true, iteration stops early.
func (k Keeper) IterateValidators(ctx sdk.Context, fn func(Validator) bool) {
	store := ctx.KVStore(k.StoreKey)
	prefix := []byte("validator:")
	end := prefixEnd(prefix)
	iter := store.Iterator(prefix, end)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var val Validator
		k.cdc.MustUnmarshalLengthPrefixed(iter.Value(), &val)
		if fn(val) {
			break
		}
	}
}

// prefixEnd returns the end key for a prefix scan (increment last byte).
func prefixEnd(prefix []byte) []byte {
	if len(prefix) == 0 {
		return nil
	}
	end := make([]byte, len(prefix))
	copy(end, prefix)
	for i := len(end) - 1; i >= 0; i-- {
		end[i]++
		if end[i] != 0 {
			return end[:i+1]
		}
	}
	return nil // overflow: prefix was all 0xFF
}

// EnforceDomainMembership checks that a validator is still a member of at
// least one of its registered domains. Returns false if no domains remain.
func (k Keeper) EnforceDomainMembership(ctx sdk.Context, operatorAddr string) bool {
	val, found := k.GetValidator(ctx, operatorAddr)
	if !found {
		return false
	}

	var active []string
	for _, domName := range val.Domains {
		domain, ok := k.GetDomain(ctx, domName)
		if !ok {
			continue
		}
		for _, m := range domain.Members {
			if m == operatorAddr {
				active = append(active, domName)
				break
			}
		}
	}

	if len(active) == 0 {
		return false
	}

	val.Domains = active
	k.SetValidator(ctx, val)
	return true
}

// DistributeStakingRewards distributes node staking rewards (eq.5) to all
// bonded validators if at least RewardInterval seconds have elapsed.
func (k Keeper) DistributeStakingRewards(ctx sdk.Context) error {
	store := ctx.KVStore(k.StoreKey)
	blockTime := ctx.BlockTime().Unix()

	// Load last reward time.
	var lastRewardTime int64
	if bz := store.Get([]byte("pod:last-reward-time")); bz != nil {
		k.cdc.MustUnmarshalLengthPrefixed(bz, &lastRewardTime)
	} else {
		// First call — initialize and return.
		bz := k.cdc.MustMarshalLengthPrefixed(blockTime)
		store.Set([]byte("pod:last-reward-time"), bz)
		return nil
	}

	elapsed := blockTime - lastRewardTime
	if elapsed < RewardInterval {
		return nil
	}

	// Load total coins released so far.
	var totalRelease math.Int
	if bz := store.Get([]byte("pod:total-release")); bz != nil {
		k.cdc.MustUnmarshalLengthPrefixed(bz, &totalRelease)
	} else {
		totalRelease = math.ZeroInt()
	}

	var newRelease math.Int = math.ZeroInt()

	k.IterateValidators(ctx, func(val Validator) bool {
		if val.Jailed {
			return false
		}
		stakeAmt := val.Stake.AmountOf("pnyx")
		reward := rewards.CalcNodeReward(stakeAmt, totalRelease, elapsed)
		if reward.IsPositive() {
			val.Stake = val.Stake.Add(sdk.NewCoin("pnyx", reward))
			val.Power = val.Stake.AmountOf("pnyx").Int64() / rewards.StakeMin
			k.SetValidator(ctx, val)
			newRelease = newRelease.Add(reward)
		}
		return false
	})

	totalRelease = totalRelease.Add(newRelease)
	store.Set([]byte("pod:total-release"), k.cdc.MustMarshalLengthPrefixed(&totalRelease))
	store.Set([]byte("pod:last-reward-time"), k.cdc.MustMarshalLengthPrefixed(blockTime))

	return nil
}

// BuildValidatorUpdates constructs the CometBFT ValidatorUpdate slice for
// the current validator set. Jailed validators are reported with Power 0.
func (k Keeper) BuildValidatorUpdates(ctx sdk.Context) []abci.ValidatorUpdate {
	var updates []abci.ValidatorUpdate

	k.IterateValidators(ctx, func(val Validator) bool {
		pk := cryptoproto.PublicKey{
			Sum: &cryptoproto.PublicKey_Ed25519{Ed25519: val.PubKey},
		}
		power := val.Power
		if val.Jailed {
			power = 0
		}
		updates = append(updates, abci.ValidatorUpdate{PubKey: pk, Power: power})
		return false
	})

	return updates
}
