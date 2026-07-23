package truedemocracy

import (
	"encoding/hex"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	cryptoproto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	rewards "truerepublic/treasury/keeper"
)

// KV store key helpers for validator state.
func validatorKey(operatorAddr string) []byte {
	return []byte("validator:" + operatorAddr)
}

func valPubKeyKey(pubKeyBytes []byte) []byte {
	return []byte("val-pubkey:" + hex.EncodeToString(pubKeyBytes))
}

func removedValidatorKey(pubKeyBytes []byte) []byte {
	return []byte("validator-removed:" + hex.EncodeToString(pubKeyBytes))
}

func revokedValidatorKey(pubKeyBytes []byte) []byte {
	return []byte("validator-revoked:" + hex.EncodeToString(pubKeyBytes))
}

func pendingValidatorRotationKey(operatorAddr string) []byte {
	return []byte("validator-rotation:" + operatorAddr)
}

func pendingValidatorPubKeyKey(pubKeyBytes []byte) []byte {
	return []byte("validator-rotation-pubkey:" + hex.EncodeToString(pubKeyBytes))
}

func consensusAuthorityIndexKey(operatorAddr string) []byte {
	return []byte("validator-consensus-authority:" + operatorAddr)
}

func consensusKeyDerivedOperator(pubKeyBytes []byte) string {
	return sdk.AccAddress((&ed25519.PubKey{Key: pubKeyBytes}).Address()).String()
}

func (k Keeper) validatorAuthorityIsCoupled(ctx sdk.Context, operatorAddr string, pubKeyBytes []byte) bool {
	store := ctx.KVStore(k.StoreKey)
	derived := consensusKeyDerivedOperator(pubKeyBytes)
	return operatorAddr == derived || store.Has(validatorKey(derived)) || store.Has(consensusAuthorityIndexKey(operatorAddr))
}

// RegisterValidator registers a new Proof of Domain validator.
// The operator must be a member of the given domain and stake >= StakeMin PNYX.
func (k Keeper) RegisterValidator(ctx sdk.Context, operatorAddr string, pubKeyBytes []byte, stake sdk.Coins, domainName string) error {
	if len(pubKeyBytes) != 32 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "pubkey must be 32 bytes (ed25519)")
	}
	if k.validatorAuthorityIsCoupled(ctx, operatorAddr, pubKeyBytes) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "validator operator authority must remain independent from active and revoked consensus keys")
	}

	pnyxAmt := stake.AmountOf(PNYXDenom)
	if pnyxAmt.LT(math.NewInt(rewards.StakeMin)) {
		return errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, "stake %s below minimum %d", pnyxAmt, rewards.StakeMin)
	}
	if !pnyxAmt.IsInt64() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "stake exceeds supported range")
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
	if store.Has(valPubKeyKey(pubKeyBytes)) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "validator public key already registered")
	}
	if store.Has(removedValidatorKey(pubKeyBytes)) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "validator public key removal is pending")
	}
	if store.Has(revokedValidatorKey(pubKeyBytes)) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "validator public key is permanently revoked")
	}
	if store.Has(pendingValidatorPubKeyKey(pubKeyBytes)) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "validator public key rotation is pending")
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
	store.Set(consensusAuthorityIndexKey(consensusKeyDerivedOperator(pubKeyBytes)), []byte(operatorAddr))

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

// QueueValidatorPowerZero records a one-shot removal only for a key that may
// already exist in CometBFT's validator set. A just-rotated replacement key is
// not active until H+2; queuing power zero for it in the rotation block would
// make CometBFT reject an attempt to remove a validator it has never seen.
func (k Keeper) QueueValidatorPowerZero(ctx sdk.Context, val Validator) {
	if val.Jailed || val.Power <= 0 {
		return
	}
	store := ctx.KVStore(k.StoreKey)
	if bz := store.Get(pendingValidatorRotationKey(val.OperatorAddr)); bz != nil {
		var pending PendingValidatorKeyRotation
		k.cdc.MustUnmarshalLengthPrefixed(bz, &pending)
		if string(pending.NewPubKey) == string(val.PubKey) {
			if ctx.BlockHeight() > pending.StartedHeight {
				pending.DeactivateNewKey = true
				store.Set(pendingValidatorRotationKey(val.OperatorAddr), k.cdc.MustMarshalLengthPrefixed(&pending))
			}
			return
		}
	}
	store.Set(removedValidatorKey(val.PubKey), append([]byte(nil), val.PubKey...))
}

// GetValidatorByPubKey looks up a validator via the reverse pubkey index.
func (k Keeper) GetValidatorByPubKey(ctx sdk.Context, pubKeyBytes []byte) (Validator, bool) {
	store := ctx.KVStore(k.StoreKey)
	addrBz := store.Get(valPubKeyKey(pubKeyBytes))
	if addrBz == nil {
		// CometBFT applies validator updates after a delay. Preserve attribution
		// of the old key during that window so evidence and missed-block records
		// are charged to the same operator after rotation.
		addrBz = store.Get(pendingValidatorPubKeyKey(pubKeyBytes))
		if addrBz == nil {
			return Validator{}, false
		}
	}
	return k.GetValidator(ctx, string(addrBz))
}

// RotateValidatorKey atomically replaces an authenticated operator's
// consensus key while preserving every stake, domain, power and jail claim.
func (k Keeper) RotateValidatorKey(ctx sdk.Context, sender sdk.AccAddress, operatorAddr string, expectedOldPubKey, newPubKey []byte) ([]byte, error) {
	if err := requireSignerClaim(sender, operatorAddr, "operator address"); err != nil {
		return nil, err
	}
	if len(newPubKey) != 32 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "pubkey must be 32 bytes (ed25519)")
	}

	cacheCtx, write := ctx.CacheContext()
	store := cacheCtx.KVStore(k.StoreKey)
	val, found := k.GetValidator(cacheCtx, operatorAddr)
	if !found {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "validator not found")
	}
	if val.Jailed || val.Power <= 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "only an active positive-power validator can rotate its consensus key")
	}
	if operatorAddr == consensusKeyDerivedOperator(val.PubKey) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "legacy consensus-derived operator authority requires an explicit migration")
	}
	if string(val.PubKey) != string(expectedOldPubKey) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "expected old validator public key does not match current key")
	}
	if store.Has(pendingValidatorRotationKey(operatorAddr)) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "validator key rotation is already pending")
	}
	if string(val.PubKey) == string(newPubKey) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "new validator public key must differ from current key")
	}
	if cacheCtx.BlockHeight() < 0 || cacheCtx.BlockHeight() > int64(^uint64(0)>>1)-2 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "block height cannot represent validator key activation window")
	}
	if store.Has(removedValidatorKey(val.PubKey)) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "current validator public key removal is pending")
	}
	if store.Has(valPubKeyKey(newPubKey)) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "validator public key already registered")
	}
	if store.Has(removedValidatorKey(newPubKey)) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "validator public key removal is pending")
	}
	if store.Has(revokedValidatorKey(newPubKey)) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "validator public key is permanently revoked")
	}
	if store.Has(pendingValidatorPubKeyKey(newPubKey)) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "validator public key rotation is pending")
	}
	if store.Has(validatorKey(consensusKeyDerivedOperator(newPubKey))) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "new consensus key must remain independent from every validator operator authority")
	}

	oldPubKey := append([]byte(nil), val.PubKey...)
	revoked := RevokedValidatorKey{
		PubKey:          oldPubKey,
		OperatorAddr:    operatorAddr,
		RevokedAtHeight: cacheCtx.BlockHeight(),
	}
	pending := PendingValidatorKeyRotation{
		OperatorAddr:     operatorAddr,
		OldPubKey:        oldPubKey,
		NewPubKey:        append([]byte(nil), newPubKey...),
		StartedHeight:    cacheCtx.BlockHeight(),
		ClearAfterHeight: cacheCtx.BlockHeight() + 2,
	}

	store.Delete(valPubKeyKey(oldPubKey))
	store.Set(removedValidatorKey(oldPubKey), oldPubKey)
	store.Set(revokedValidatorKey(oldPubKey), k.cdc.MustMarshalLengthPrefixed(&revoked))
	store.Set(consensusAuthorityIndexKey(consensusKeyDerivedOperator(oldPubKey)), []byte(operatorAddr))
	store.Set(consensusAuthorityIndexKey(consensusKeyDerivedOperator(newPubKey)), []byte(operatorAddr))
	store.Set(pendingValidatorRotationKey(operatorAddr), k.cdc.MustMarshalLengthPrefixed(&pending))
	store.Set(pendingValidatorPubKeyKey(oldPubKey), []byte(operatorAddr))
	store.Set(pendingValidatorPubKeyKey(newPubKey), []byte(operatorAddr))
	val.PubKey = append([]byte(nil), newPubKey...)
	k.SetValidator(cacheCtx, val)
	store.Set(valPubKeyKey(newPubKey), []byte(operatorAddr))
	write()
	return oldPubKey, nil
}

func (k Keeper) IsValidatorKeyRevoked(ctx sdk.Context, pubKey []byte) bool {
	return ctx.KVStore(k.StoreKey).Has(revokedValidatorKey(pubKey))
}

// GetValidatorForDoubleSignEvidence resolves active and pending keys first,
// then uses the permanent revocation owner. Equivocation evidence can arrive
// after the H→H+2 activation window, while downtime attribution must not.
func (k Keeper) GetValidatorForDoubleSignEvidence(ctx sdk.Context, pubKey []byte) (Validator, bool) {
	if validator, found := k.GetValidatorByPubKey(ctx, pubKey); found {
		return validator, true
	}
	bz := ctx.KVStore(k.StoreKey).Get(revokedValidatorKey(pubKey))
	if bz == nil {
		return Validator{}, false
	}
	var record RevokedValidatorKey
	k.cdc.MustUnmarshalLengthPrefixed(bz, &record)
	return k.GetValidator(ctx, record.OperatorAddr)
}

func (k Keeper) IterateRevokedValidatorKeys(ctx sdk.Context, fn func(RevokedValidatorKey) bool) {
	store := ctx.KVStore(k.StoreKey)
	prefix := []byte("validator-revoked:")
	iter := store.Iterator(prefix, prefixEnd(prefix))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var record RevokedValidatorKey
		k.cdc.MustUnmarshalLengthPrefixed(iter.Value(), &record)
		if fn(record) {
			return
		}
	}
}

func (k Keeper) IteratePendingValidatorKeyRotations(ctx sdk.Context, fn func(PendingValidatorKeyRotation) bool) {
	store := ctx.KVStore(k.StoreKey)
	prefix := []byte("validator-rotation:")
	iter := store.Iterator(prefix, prefixEnd(prefix))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var record PendingValidatorKeyRotation
		k.cdc.MustUnmarshalLengthPrefixed(iter.Value(), &record)
		if fn(record) {
			return
		}
	}
}

func (k Keeper) restoreRevokedValidatorKey(ctx sdk.Context, record RevokedValidatorKey) {
	store := ctx.KVStore(k.StoreKey)
	store.Set(revokedValidatorKey(record.PubKey), k.cdc.MustMarshalLengthPrefixed(&record))
	store.Set(consensusAuthorityIndexKey(consensusKeyDerivedOperator(record.PubKey)), []byte(record.OperatorAddr))
}

func (k Keeper) restorePendingValidatorKeyRotation(ctx sdk.Context, record PendingValidatorKeyRotation) {
	store := ctx.KVStore(k.StoreKey)
	store.Set(pendingValidatorRotationKey(record.OperatorAddr), k.cdc.MustMarshalLengthPrefixed(&record))
	store.Set(pendingValidatorPubKeyKey(record.OldPubKey), []byte(record.OperatorAddr))
	store.Set(pendingValidatorPubKeyKey(record.NewPubKey), []byte(record.OperatorAddr))
}

// WithdrawStake allows a validator to withdraw some or all of their staked
// PNYX. The withdrawal is subject to the PoD transfer limit (WP §7):
// cumulative withdrawals across all domain validators cannot exceed 10% of
// the domain's total historical payouts.
func (k Keeper) WithdrawStake(ctx sdk.Context, operatorAddr string, amount int64) error {
	if amount <= 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "withdrawal amount must be positive")
	}
	val, found := k.GetValidator(ctx, operatorAddr)
	if !found {
		return errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "validator not found")
	}

	stake := val.Stake.AmountOf(PNYXDenom)
	if !stake.IsInt64() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "validator stake exceeds supported range")
	}
	stakeAmt := stake.Int64()
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

	newStake := stakeAmt - amount
	if newStake > 0 && newStake < rewards.StakeMin {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"withdrawal would leave dust stake %d below minimum %d",
			newStake,
			rewards.StakeMin,
		)
	}

	// Update domain accounting only after every withdrawal precondition passed.
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "domain %s not found", domainName)
	}
	domain.TransferredStake += amount
	store := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	store.Set([]byte("domain:"+domainName), bz)

	if newStake == 0 {
		return k.RemoveValidator(ctx, operatorAddr)
	}

	val.Stake = sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, newStake))
	val.Power = newStake / rewards.StakeMin
	k.SetValidator(ctx, val)
	return nil
}

// RemoveValidator deletes a validator, its reverse index, and records a
// one-shot CometBFT power-zero update for the removed consensus key.
func (k Keeper) RemoveValidator(ctx sdk.Context, operatorAddr string) error {
	val, found := k.GetValidator(ctx, operatorAddr)
	if !found {
		return errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "validator not found")
	}
	store := ctx.KVStore(k.StoreKey)
	if store.Has(pendingValidatorRotationKey(operatorAddr)) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "validator key rotation is pending")
	}
	store.Delete(validatorKey(operatorAddr))
	store.Delete(valPubKeyKey(val.PubKey))
	store.Set(removedValidatorKey(val.PubKey), append([]byte(nil), val.PubKey...))
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
	cacheCtx, write := ctx.CacheContext()
	store := cacheCtx.KVStore(k.StoreKey)
	blockTime := ctx.BlockTime().Unix()

	// Load last reward time.
	var lastRewardTime int64
	if bz := store.Get([]byte("pod:last-reward-time")); bz != nil {
		k.cdc.MustUnmarshalLengthPrefixed(bz, &lastRewardTime)
	} else {
		// First call — initialize and return.
		bz := k.cdc.MustMarshalLengthPrefixed(blockTime)
		store.Set([]byte("pod:last-reward-time"), bz)
		write()
		return nil
	}

	elapsed := blockTime - lastRewardTime
	if elapsed < RewardInterval {
		return nil
	}

	supply, err := k.issuer.Supply(cacheCtx)
	if err != nil {
		return errorsmod.Wrap(err, "read canonical supply for staking rewards")
	}

	type allocation struct {
		validator Validator
		requested math.Int
	}
	var allocations []allocation
	totalRequested := math.ZeroInt()
	k.IterateValidators(cacheCtx, func(val Validator) bool {
		if val.Jailed {
			return false
		}
		stakeAmt := val.Stake.AmountOf(PNYXDenom)
		reward := rewards.CalcNodeReward(stakeAmt, supply, elapsed)
		if reward.IsPositive() {
			allocations = append(allocations, allocation{validator: val, requested: reward})
			totalRequested = totalRequested.Add(reward)
		}
		return false
	})

	minted, err := k.issuer.MintUpToCap(cacheCtx, totalRequested)
	if err != nil {
		return errorsmod.Wrap(err, "mint staking rewards")
	}
	remaining := minted
	for _, allocation := range allocations {
		grant := allocation.requested
		if grant.GT(remaining) {
			grant = remaining
		}
		if !grant.IsPositive() {
			break
		}
		allocation.validator.Stake = allocation.validator.Stake.Add(sdk.NewCoin(PNYXDenom, grant))
		allocation.validator.Power = allocation.validator.Stake.AmountOf(PNYXDenom).Int64() / rewards.StakeMin
		k.SetValidator(cacheCtx, allocation.validator)
		remaining = remaining.Sub(grant)
	}
	store.Set([]byte("pod:last-reward-time"), k.cdc.MustMarshalLengthPrefixed(blockTime))

	write()
	return nil
}

// DistributeDomainInterest credits domain treasuries with interest (eq.4)
// from the token release mechanism. This runs every RewardInterval alongside
// node staking rewards. Only active domains (with payouts in this interval)
// receive interest, capped by their payout amount.
func (k Keeper) DistributeDomainInterest(ctx sdk.Context) error {
	cacheCtx, write := ctx.CacheContext()
	store := cacheCtx.KVStore(k.StoreKey)
	blockTime := ctx.BlockTime().Unix()

	// Load last domain interest time.
	var lastInterestTime int64
	if bz := store.Get([]byte("dom:last-interest-time")); bz != nil {
		k.cdc.MustUnmarshalLengthPrefixed(bz, &lastInterestTime)
	} else {
		bz := k.cdc.MustMarshalLengthPrefixed(blockTime)
		store.Set([]byte("dom:last-interest-time"), bz)
		k.IterateDomains(cacheCtx, func(domain Domain) bool {
			store.Set(domainPayoutSnapshotKey(domain.Name), k.cdc.MustMarshalLengthPrefixed(domain.TotalPayouts))
			return false
		})
		write()
		return nil
	}

	elapsed := blockTime - lastInterestTime
	if elapsed < RewardInterval {
		return nil
	}

	supply, err := k.issuer.Supply(cacheCtx)
	if err != nil {
		return errorsmod.Wrap(err, "read canonical supply for domain interest")
	}

	type allocation struct {
		domain    Domain
		requested math.Int
	}
	var allocations []allocation
	totalRequested := math.ZeroInt()
	k.IterateDomains(cacheCtx, func(domain Domain) bool {
		treasure := domain.Treasury.AmountOf(PNYXDenom)
		snapshotKey := domainPayoutSnapshotKey(domain.Name)
		previousPayouts := domain.TotalPayouts
		if bz := store.Get(snapshotKey); bz != nil {
			k.cdc.MustUnmarshalLengthPrefixed(bz, &previousPayouts)
		}
		intervalPayouts := domain.TotalPayouts - previousPayouts
		if intervalPayouts < 0 {
			intervalPayouts = 0
		}
		interest := rewards.CalcDomainInterest(treasure, math.NewInt(intervalPayouts), supply, elapsed)
		if interest.IsPositive() {
			allocations = append(allocations, allocation{domain: domain, requested: interest})
			totalRequested = totalRequested.Add(interest)
		}
		// Missing snapshots belong to pre-GH-13 state. Baseline them lazily at
		// the current cumulative payout so historical payouts are never rewarded.
		store.Set(snapshotKey, k.cdc.MustMarshalLengthPrefixed(domain.TotalPayouts))
		return false
	})

	minted, err := k.issuer.MintUpToCap(cacheCtx, totalRequested)
	if err != nil {
		return errorsmod.Wrap(err, "mint domain interest")
	}
	remaining := minted
	for _, allocation := range allocations {
		grant := allocation.requested
		if grant.GT(remaining) {
			grant = remaining
		}
		if !grant.IsPositive() {
			break
		}
		allocation.domain.Treasury = allocation.domain.Treasury.Add(sdk.NewCoin(PNYXDenom, grant))
		store.Set([]byte("domain:"+allocation.domain.Name), k.cdc.MustMarshalLengthPrefixed(&allocation.domain))
		remaining = remaining.Sub(grant)
	}
	store.Set([]byte("dom:last-interest-time"), k.cdc.MustMarshalLengthPrefixed(blockTime))
	write()
	return nil
}

func domainPayoutSnapshotKey(domainName string) []byte {
	return []byte("dom:last-payouts:" + domainName)
}

// BuildValidatorUpdates constructs the CometBFT ValidatorUpdate slice for the
// current validator set. Jailed validators are reported with Power 0. Validators
// removed since the last call are emitted once with Power 0 so CometBFT can
// evict them from the consensus set.
func (k Keeper) BuildValidatorUpdates(ctx sdk.Context) []abci.ValidatorUpdate {
	var updates []abci.ValidatorUpdate

	k.IterateValidators(ctx, func(val Validator) bool {
		if val.Jailed || val.Power <= 0 {
			return false
		}
		pk := cryptoproto.PublicKey{
			Sum: &cryptoproto.PublicKey_Ed25519{Ed25519: val.PubKey},
		}
		updates = append(updates, abci.ValidatorUpdate{PubKey: pk, Power: val.Power})
		return false
	})

	store := ctx.KVStore(k.StoreKey)
	removedPrefix := []byte("validator-removed:")
	iter := store.Iterator(removedPrefix, prefixEnd(removedPrefix))
	var removalKeys [][]byte
	for ; iter.Valid(); iter.Next() {
		pubKey := append([]byte(nil), iter.Value()...)
		pk := cryptoproto.PublicKey{
			Sum: &cryptoproto.PublicKey_Ed25519{Ed25519: pubKey},
		}
		updates = append(updates, abci.ValidatorUpdate{PubKey: pk, Power: 0})
		removalKeys = append(removalKeys, append([]byte(nil), iter.Key()...))
	}
	iter.Close()
	for _, key := range removalKeys {
		store.Delete(key)
	}

	// If inactivation happens at H+1 after H scheduled the replacement, emit a
	// one-shot power-zero update while that key is present in NextValidators.
	// Same-H inactivation never schedules the replacement and needs no removal.
	var deferred []PendingValidatorKeyRotation
	k.IteratePendingValidatorKeyRotations(ctx, func(rotation PendingValidatorKeyRotation) bool {
		if rotation.DeactivateNewKey && ctx.BlockHeight() >= rotation.StartedHeight+1 {
			deferred = append(deferred, rotation)
		}
		return false
	})
	for _, rotation := range deferred {
		pk := cryptoproto.PublicKey{Sum: &cryptoproto.PublicKey_Ed25519{Ed25519: rotation.NewPubKey}}
		updates = append(updates, abci.ValidatorUpdate{PubKey: pk, Power: 0})
		rotation.DeactivateNewKey = false
		store.Set(pendingValidatorRotationKey(rotation.OperatorAddr), k.cdc.MustMarshalLengthPrefixed(&rotation))
	}

	// The old consensus key must remain attributable through CometBFT's H→H+2
	// activation window. Clear the durable guard only after all processing at
	// the terminal height has completed.
	var completed []PendingValidatorKeyRotation
	k.IteratePendingValidatorKeyRotations(ctx, func(rotation PendingValidatorKeyRotation) bool {
		if ctx.BlockHeight() >= rotation.ClearAfterHeight {
			completed = append(completed, rotation)
		}
		return false
	})
	for _, rotation := range completed {
		store.Delete(pendingValidatorRotationKey(rotation.OperatorAddr))
		store.Delete(pendingValidatorPubKeyKey(rotation.OldPubKey))
		store.Delete(pendingValidatorPubKeyKey(rotation.NewPubKey))
	}

	return updates
}
