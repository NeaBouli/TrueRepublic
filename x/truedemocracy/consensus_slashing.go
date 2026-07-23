package truedemocracy

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/bits"
	"sort"

	"cosmossdk.io/core/comet"
	errorsmod "cosmossdk.io/errors"
	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	consensusAddressLength = 20
	livenessBitmapLength   = int((SignedBlocksWindow + 7) / 8)
)

func consensusKeyHistoryKey(address []byte) []byte {
	return []byte("consensus-key-history:" + hex.EncodeToString(address))
}

func validatorSigningInfoKey(operatorAddr string) []byte {
	return []byte("validator-signing:" + operatorAddr)
}

func processedInfractionKey(id []byte) []byte {
	return []byte("processed-infraction:" + hex.EncodeToString(id))
}

func lastCommitCursorKey() []byte {
	return []byte("last-commit-cursor")
}

func consensusAddressFromPubKey(pubKey []byte) []byte {
	return append([]byte(nil), cmted25519.PubKey(pubKey).Address()...)
}

func initialConsensusActivationHeight(ctx sdk.Context) int64 {
	if ctx.BlockHeight() <= 0 {
		return 1
	}
	return ctx.BlockHeight() + sdk.ValidatorUpdateDelay + 1
}

func (k Keeper) setConsensusKeyRecord(ctx sdk.Context, record ConsensusKeyRecord) {
	ctx.KVStore(k.StoreKey).Set(
		consensusKeyHistoryKey(record.ConsensusAddress),
		k.cdc.MustMarshalLengthPrefixed(&record),
	)
}

func (k Keeper) GetConsensusKeyRecord(ctx sdk.Context, address []byte) (ConsensusKeyRecord, bool) {
	if len(address) != consensusAddressLength {
		return ConsensusKeyRecord{}, false
	}
	store := ctx.KVStore(k.StoreKey)
	if bz := store.Get(consensusKeyHistoryKey(address)); bz != nil {
		var record ConsensusKeyRecord
		k.cdc.MustUnmarshalLengthPrefixed(bz, &record)
		return record, true
	}

	// Backfill pre-GH-59 recovery state deterministically. GH-56 permanently
	// indexed every active and revoked consensus address to its operator.
	operatorBz := store.Get(consensusAuthorityIndexKey(sdk.AccAddress(address).String()))
	if operatorBz == nil {
		return ConsensusKeyRecord{}, false
	}
	operatorAddr := string(operatorBz)
	if validator, found := k.GetValidator(ctx, operatorAddr); found &&
		bytes.Equal(consensusAddressFromPubKey(validator.PubKey), address) {
		record := ConsensusKeyRecord{
			ConsensusAddress: append([]byte(nil), address...),
			PubKey:           append([]byte(nil), validator.PubKey...),
			OperatorAddr:     operatorAddr,
			ActivatedHeight:  1,
		}
		k.setConsensusKeyRecord(ctx, record)
		return record, true
	}
	if pending, found := k.GetPendingValidatorRemoval(ctx, operatorAddr); found &&
		bytes.Equal(consensusAddressFromPubKey(pending.Validator.PubKey), address) {
		record := ConsensusKeyRecord{
			ConsensusAddress: append([]byte(nil), address...),
			PubKey:           append([]byte(nil), pending.Validator.PubKey...),
			OperatorAddr:     operatorAddr,
			ActivatedHeight:  1,
			RetiredHeight:    pending.ConsensusRetiredHeight,
		}
		k.setConsensusKeyRecord(ctx, record)
		return record, true
	}
	var matched RevokedValidatorKey
	k.IterateRevokedValidatorKeys(ctx, func(record RevokedValidatorKey) bool {
		if bytes.Equal(consensusAddressFromPubKey(record.PubKey), address) {
			matched = record
			return true
		}
		return false
	})
	if len(matched.PubKey) == 0 {
		return ConsensusKeyRecord{}, false
	}
	record := ConsensusKeyRecord{
		ConsensusAddress: append([]byte(nil), address...),
		PubKey:           append([]byte(nil), matched.PubKey...),
		OperatorAddr:     matched.OperatorAddr,
		ActivatedHeight:  1,
		RetiredHeight:    matched.RevokedAtHeight + sdk.ValidatorUpdateDelay + 1,
	}
	k.setConsensusKeyRecord(ctx, record)
	return record, true
}

func (k Keeper) IterateConsensusKeyRecords(ctx sdk.Context, fn func(ConsensusKeyRecord) bool) {
	prefix := []byte("consensus-key-history:")
	iter := ctx.KVStore(k.StoreKey).Iterator(prefix, prefixEnd(prefix))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var record ConsensusKeyRecord
		k.cdc.MustUnmarshalLengthPrefixed(iter.Value(), &record)
		if fn(record) {
			return
		}
	}
}

func (k Keeper) registerConsensusKeyRecord(ctx sdk.Context, pubKey []byte, operatorAddr string, activatedHeight int64) {
	k.setConsensusKeyRecord(ctx, ConsensusKeyRecord{
		ConsensusAddress: consensusAddressFromPubKey(pubKey),
		PubKey:           append([]byte(nil), pubKey...),
		OperatorAddr:     operatorAddr,
		ActivatedHeight:  activatedHeight,
	})
}

func (k Keeper) retireConsensusKeyRecord(ctx sdk.Context, pubKey []byte, retiredHeight int64) {
	address := consensusAddressFromPubKey(pubKey)
	record, found := k.GetConsensusKeyRecord(ctx, address)
	if !found {
		return
	}
	if record.RetiredHeight == 0 || retiredHeight < record.RetiredHeight {
		record.RetiredHeight = retiredHeight
		k.setConsensusKeyRecord(ctx, record)
	}
}

func (k Keeper) resolveConsensusKeyAtHeight(ctx sdk.Context, address []byte, height int64) (ConsensusKeyRecord, error) {
	if len(address) != consensusAddressLength {
		return ConsensusKeyRecord{}, errorsmod.Wrapf(
			sdkerrors.ErrInvalidAddress,
			"consensus address must be %d bytes",
			consensusAddressLength,
		)
	}
	record, found := k.GetConsensusKeyRecord(ctx, address)
	if !found {
		return ConsensusKeyRecord{}, errorsmod.Wrap(sdkerrors.ErrUnknownAddress, "consensus address is not registered")
	}
	if height < record.ActivatedHeight || (record.RetiredHeight > 0 && height >= record.RetiredHeight) {
		return ConsensusKeyRecord{}, errorsmod.Wrapf(
			sdkerrors.ErrUnauthorized,
			"consensus address was not active at height %d",
			height,
		)
	}
	return record, nil
}

func (k Keeper) getValidatorSigningInfo(ctx sdk.Context, operatorAddr string) (ValidatorSigningInfo, bool) {
	bz := ctx.KVStore(k.StoreKey).Get(validatorSigningInfoKey(operatorAddr))
	if bz == nil {
		return ValidatorSigningInfo{}, false
	}
	var info ValidatorSigningInfo
	k.cdc.MustUnmarshalLengthPrefixed(bz, &info)
	return info, true
}

func (k Keeper) setValidatorSigningInfo(ctx sdk.Context, info ValidatorSigningInfo) {
	ctx.KVStore(k.StoreKey).Set(
		validatorSigningInfoKey(info.OperatorAddr),
		k.cdc.MustMarshalLengthPrefixed(&info),
	)
}

func (k Keeper) deleteValidatorSigningInfo(ctx sdk.Context, operatorAddr string) {
	ctx.KVStore(k.StoreKey).Delete(validatorSigningInfoKey(operatorAddr))
}

func (k Keeper) IterateValidatorSigningInfos(ctx sdk.Context, fn func(ValidatorSigningInfo) bool) {
	prefix := []byte("validator-signing:")
	iter := ctx.KVStore(k.StoreKey).Iterator(prefix, prefixEnd(prefix))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var info ValidatorSigningInfo
		k.cdc.MustUnmarshalLengthPrefixed(iter.Value(), &info)
		if fn(info) {
			return
		}
	}
}

func getMissedBit(bitmap []byte, index int64) bool {
	return bitmap[index/8]&(1<<uint(index%8)) != 0
}

func setMissedBit(bitmap []byte, index int64, missed bool) {
	mask := byte(1 << uint(index%8))
	if missed {
		bitmap[index/8] |= mask
		return
	}
	bitmap[index/8] &^= mask
}

func missedBitCount(bitmap []byte) int64 {
	var count int64
	for _, value := range bitmap {
		count += int64(bits.OnesCount8(value))
	}
	return count
}

func (k Keeper) resetValidatorSigningInfo(ctx sdk.Context, operatorAddr string, commitHeight int64) {
	k.setValidatorSigningInfo(ctx, ValidatorSigningInfo{
		OperatorAddr:             operatorAddr,
		StartCommitHeight:        commitHeight,
		MissedBitmap:             make([]byte, livenessBitmapLength),
		LastObservedCommitHeight: commitHeight,
	})
}

func (k Keeper) getProcessedInfraction(ctx sdk.Context, id []byte) (ProcessedInfraction, bool) {
	bz := ctx.KVStore(k.StoreKey).Get(processedInfractionKey(id))
	if bz == nil {
		return ProcessedInfraction{}, false
	}
	var record ProcessedInfraction
	k.cdc.MustUnmarshalLengthPrefixed(bz, &record)
	return record, true
}

func (k Keeper) setProcessedInfraction(ctx sdk.Context, record ProcessedInfraction) {
	ctx.KVStore(k.StoreKey).Set(
		processedInfractionKey(record.ID),
		k.cdc.MustMarshalLengthPrefixed(&record),
	)
}

func (k Keeper) IterateProcessedInfractions(ctx sdk.Context, fn func(ProcessedInfraction) bool) {
	prefix := []byte("processed-infraction:")
	iter := ctx.KVStore(k.StoreKey).Iterator(prefix, prefixEnd(prefix))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var record ProcessedInfraction
		k.cdc.MustUnmarshalLengthPrefixed(iter.Value(), &record)
		if fn(record) {
			return
		}
	}
}

func (k Keeper) getLastCommitCursor(ctx sdk.Context) (LastCommitCursor, bool) {
	bz := ctx.KVStore(k.StoreKey).Get(lastCommitCursorKey())
	if bz == nil {
		return LastCommitCursor{}, false
	}
	var cursor LastCommitCursor
	k.cdc.MustUnmarshalLengthPrefixed(bz, &cursor)
	return cursor, true
}

func (k Keeper) setLastCommitCursor(ctx sdk.Context, cursor LastCommitCursor) {
	ctx.KVStore(k.StoreKey).Set(lastCommitCursorKey(), k.cdc.MustMarshalLengthPrefixed(&cursor))
}

type consensusEvidenceEnvelope struct {
	Type             comet.MisbehaviorType
	Address          []byte
	ValidatorPower   int64
	Height           int64
	TimeNanos        int64
	TotalVotingPower int64
	ID               []byte
	Canonical        []byte
}

func appendInt64(dst []byte, value int64) []byte {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], uint64(value))
	return append(dst, encoded[:]...)
}

func appendInt32(dst []byte, value int32) []byte {
	var encoded [4]byte
	binary.BigEndian.PutUint32(encoded[:], uint32(value))
	return append(dst, encoded[:]...)
}

func normalizedInfractionID(chainID string, address []byte, height int64) []byte {
	payload := append([]byte("truerepublic/equivocation/v1\x00"), []byte(chainID)...)
	payload = append(payload, 0)
	payload = append(payload, address...)
	payload = appendInt64(payload, height)
	sum := sha256.Sum256(payload)
	return sum[:]
}

func canonicalEvidenceBytes(envelope consensusEvidenceEnvelope) []byte {
	payload := appendInt32(nil, int32(envelope.Type))
	payload = append(payload, envelope.Address...)
	payload = appendInt64(payload, envelope.ValidatorPower)
	payload = appendInt64(payload, envelope.Height)
	payload = appendInt64(payload, envelope.TimeNanos)
	return appendInt64(payload, envelope.TotalVotingPower)
}

func collectConsensusEvidence(ctx sdk.Context) ([]consensusEvidenceEnvelope, error) {
	evidence := ctx.CometInfo().GetEvidence()
	envelopes := make([]consensusEvidenceEnvelope, 0, evidence.Len())
	for i := 0; i < evidence.Len(); i++ {
		item := evidence.Get(i)
		if item.Type() != comet.DuplicateVote && item.Type() != comet.LightClientAttack {
			return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "unsupported misbehavior type %d", item.Type())
		}
		validator := item.Validator()
		address := validator.Address()
		if len(address) != consensusAddressLength {
			return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "misbehavior consensus address must be %d bytes", consensusAddressLength)
		}
		if validator.Power() <= 0 || item.TotalVotingPower() < validator.Power() {
			return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "misbehavior voting power is invalid")
		}
		if item.Height() <= 0 || item.Height() >= ctx.BlockHeight() {
			return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "misbehavior height is outside the committed history")
		}
		if item.Time().IsZero() || item.Time().After(ctx.BlockTime()) {
			return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "misbehavior time is invalid")
		}
		envelope := consensusEvidenceEnvelope{
			Type:             item.Type(),
			Address:          append([]byte(nil), address...),
			ValidatorPower:   validator.Power(),
			Height:           item.Height(),
			TimeNanos:        item.Time().UnixNano(),
			TotalVotingPower: item.TotalVotingPower(),
		}
		envelope.ID = normalizedInfractionID(ctx.ChainID(), envelope.Address, envelope.Height)
		envelope.Canonical = canonicalEvidenceBytes(envelope)
		envelopes = append(envelopes, envelope)
	}
	sort.Slice(envelopes, func(i, j int) bool {
		if comparison := bytes.Compare(envelopes[i].ID, envelopes[j].ID); comparison != 0 {
			return comparison < 0
		}
		return bytes.Compare(envelopes[i].Canonical, envelopes[j].Canonical) < 0
	})
	return envelopes, nil
}

type consensusVoteEnvelope struct {
	Address []byte
	Power   int64
	Flag    comet.BlockIDFlag
}

func collectConsensusVotes(ctx sdk.Context) ([]consensusVoteEnvelope, []byte, error) {
	lastCommit := ctx.CometInfo().GetLastCommit()
	votes := lastCommit.Votes()
	envelopes := make([]consensusVoteEnvelope, 0, votes.Len())
	seen := make(map[string]struct{}, votes.Len())
	for i := 0; i < votes.Len(); i++ {
		vote := votes.Get(i)
		address := vote.Validator().Address()
		if len(address) != consensusAddressLength {
			return nil, nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "last-commit consensus address must be %d bytes", consensusAddressLength)
		}
		if vote.Validator().Power() <= 0 {
			return nil, nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "last-commit validator power must be positive")
		}
		if vote.GetBlockIDFlag() != comet.BlockIDFlagAbsent &&
			vote.GetBlockIDFlag() != comet.BlockIDFlagCommit &&
			vote.GetBlockIDFlag() != comet.BlockIDFlagNil {
			return nil, nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "last-commit vote flag is invalid")
		}
		key := string(address)
		if _, exists := seen[key]; exists {
			return nil, nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "last-commit contains duplicate validator address")
		}
		seen[key] = struct{}{}
		envelopes = append(envelopes, consensusVoteEnvelope{
			Address: append([]byte(nil), address...),
			Power:   vote.Validator().Power(),
			Flag:    vote.GetBlockIDFlag(),
		})
	}
	sort.Slice(envelopes, func(i, j int) bool {
		return bytes.Compare(envelopes[i].Address, envelopes[j].Address) < 0
	})

	payload := append([]byte("truerepublic/last-commit/v1\x00"), []byte(ctx.ChainID())...)
	payload = append(payload, 0)
	payload = appendInt64(payload, ctx.BlockHeight()-1)
	payload = appendInt32(payload, lastCommit.Round())
	for _, vote := range envelopes {
		payload = append(payload, vote.Address...)
		payload = appendInt64(payload, vote.Power)
		payload = appendInt32(payload, int32(vote.Flag))
	}
	sum := sha256.Sum256(payload)
	return envelopes, sum[:], nil
}

func consensusSignalHash(commitHash []byte, evidence []consensusEvidenceEnvelope) []byte {
	payload := append([]byte("truerepublic/consensus-signals/v1\x00"), commitHash...)
	for _, item := range evidence {
		payload = append(payload, item.ID...)
		payload = appendInt64(payload, int64(len(item.Canonical)))
		payload = append(payload, item.Canonical...)
	}
	sum := sha256.Sum256(payload)
	return sum[:]
}

// ProcessConsensusSignals consumes the ABCI++ data BaseApp placed in the
// sdk.Context before module BeginBlock. Evidence, economic penalties, liveness
// state, replay markers and the commit cursor are committed atomically.
func (k Keeper) ProcessConsensusSignals(ctx sdk.Context) error {
	evidence, err := collectConsensusEvidence(ctx)
	if err != nil {
		return err
	}
	votes, commitHash, err := collectConsensusVotes(ctx)
	if err != nil {
		return err
	}
	signalHash := consensusSignalHash(commitHash, evidence)

	cacheCtx, write := ctx.CacheContext()
	commitHeight := cacheCtx.BlockHeight() - 1
	if cursor, found := k.getLastCommitCursor(cacheCtx); found {
		switch {
		case cursor.CommitHeight == commitHeight && bytes.Equal(cursor.Hash, signalHash):
			return nil
		case cursor.CommitHeight == commitHeight:
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "last-commit replay hash changed")
		case commitHeight != cursor.CommitHeight+1:
			return errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"last-commit cursor discontinuity: got %d after %d",
				commitHeight,
				cursor.CommitHeight,
			)
		}
	}

	processedInBatch := make(map[string]struct{}, len(evidence))
	for _, item := range evidence {
		idKey := string(item.ID)
		if _, duplicate := processedInBatch[idKey]; duplicate {
			continue
		}
		processedInBatch[idKey] = struct{}{}
		if _, processed := k.getProcessedInfraction(cacheCtx, item.ID); processed {
			continue
		}
		keyRecord, err := k.resolveConsensusKeyAtHeight(cacheCtx, item.Address, item.Height)
		if err != nil {
			return err
		}
		var burned int64
		if !keyRecord.Tombstoned {
			burned, err = k.handleDoubleSignForRecord(cacheCtx, keyRecord)
			if err != nil {
				return err
			}
			keyRecord.Tombstoned = true
			k.setConsensusKeyRecord(cacheCtx, keyRecord)
		}
		k.setProcessedInfraction(cacheCtx, ProcessedInfraction{
			ID:                  append([]byte(nil), item.ID...),
			MisbehaviorType:     int32(item.Type),
			ConsensusAddress:    append([]byte(nil), item.Address...),
			OperatorAddr:        keyRecord.OperatorAddr,
			InfractionHeight:    item.Height,
			InfractionTimeNanos: item.TimeNanos,
			ObservedHeight:      cacheCtx.BlockHeight(),
			ValidatorPower:      item.ValidatorPower,
			TotalVotingPower:    item.TotalVotingPower,
			BurnedAmount:        burned,
		})
	}

	for _, vote := range votes {
		keyRecord, err := k.resolveConsensusKeyAtHeight(cacheCtx, vote.Address, commitHeight)
		if err != nil {
			return err
		}
		if err := k.recordValidatorSignature(
			cacheCtx,
			keyRecord.OperatorAddr,
			commitHeight,
			vote.Flag == comet.BlockIDFlagAbsent,
		); err != nil {
			return err
		}
	}
	k.setLastCommitCursor(cacheCtx, LastCommitCursor{
		CommitHeight: commitHeight,
		Hash:         append([]byte(nil), signalHash...),
	})
	write()
	return nil
}

func validateSigningInfo(info ValidatorSigningInfo) error {
	if info.OperatorAddr == "" || info.StartCommitHeight < 0 || info.IndexOffset < 0 ||
		info.MissedBlocks < 0 || info.MissedBlocks > SignedBlocksWindow ||
		info.LastObservedCommitHeight < 0 || len(info.MissedBitmap) != livenessBitmapLength {
		return fmt.Errorf("validator signing info is malformed")
	}
	if info.MissedBlocks != missedBitCount(info.MissedBitmap) {
		return fmt.Errorf("validator signing missed counter does not match bitmap")
	}
	return nil
}
