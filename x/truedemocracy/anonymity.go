package truedemocracy

import (
	"encoding/hex"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// JoinPermissionRegister adds a domain-specific public key to the domain's
// permission register. The caller must be a domain member. The system never
// stores which member owns which domain key — that link exists only on the
// client side, ensuring anonymous voting (whitepaper Section 4).
func (k Keeper) JoinPermissionRegister(ctx sdk.Context, domainName, memberAddr string, domainPubKey []byte) error {
	if len(domainPubKey) != 32 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "domain public key must be 32 bytes (ed25519)")
	}

	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "domain %s not found", domainName)
	}

	// Verify the caller is a domain member.
	isMember := false
	for _, m := range domain.Members {
		if m == memberAddr {
			isMember = true
			break
		}
	}
	if !isMember {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "only domain members can join the permission register")
	}

	// Check for duplicate key.
	keyHex := hex.EncodeToString(domainPubKey)
	for _, existing := range domain.PermissionReg {
		if existing == keyHex {
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "domain key already registered")
		}
	}

	domain.PermissionReg = append(domain.PermissionReg, keyHex)

	store := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	store.Set([]byte("domain:"+domainName), bz)
	return nil
}

// PurgePermissionRegister clears all keys from the domain's permission
// register. After a purge, all members must re-register with fresh domain
// keys before they can vote. Members who have been removed from the domain
// cannot re-register. Only the domain admin can trigger a purge.
func (k Keeper) PurgePermissionRegister(ctx sdk.Context, domainName string, caller sdk.AccAddress) error {
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "domain %s not found", domainName)
	}

	if !caller.Equals(domain.Admin) {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "only the domain admin can purge the permission register")
	}

	domain.PermissionReg = []string{}

	store := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	store.Set([]byte("domain:"+domainName), bz)
	return nil
}

// IsKeyAuthorized checks whether a hex-encoded domain public key is present
// in the domain's permission register.
func (k Keeper) IsKeyAuthorized(ctx sdk.Context, domainName string, domainPubKeyHex string) bool {
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return false
	}
	for _, key := range domain.PermissionReg {
		if key == domainPubKeyHex {
			return true
		}
	}
	return false
}

// HasDomainKeyVoted checks whether a domain key has already voted on a
// specific suggestion, preventing double-voting.
func HasDomainKeyVoted(domain Domain, issueName, suggestionName, domainPubKeyHex string) bool {
	for _, issue := range domain.Issues {
		if issue.Name == issueName {
			for _, suggestion := range issue.Suggestions {
				if suggestion.Name == suggestionName {
					for _, r := range suggestion.Ratings {
						if r.DomainPubKeyHex == domainPubKeyHex {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

// ---------- Big Purge Schedule ----------

// Default Big Purge parameters (WP S4).
const (
	DefaultPurgeInterval    int64 = 7_776_000 // 90 days in seconds
	DefaultAnnouncementLead int64 = 604_800   // 7 days in seconds
)

// GetBigPurgeSchedule retrieves the automated purge schedule for a domain.
func (k Keeper) GetBigPurgeSchedule(ctx sdk.Context, domainName string) (BigPurgeSchedule, bool) {
	store := ctx.KVStore(k.StoreKey)
	bz := store.Get([]byte("purge-schedule:" + domainName))
	if bz == nil {
		return BigPurgeSchedule{}, false
	}
	var schedule BigPurgeSchedule
	k.cdc.MustUnmarshalLengthPrefixed(bz, &schedule)
	return schedule, true
}

// SetBigPurgeSchedule stores the purge schedule for a domain.
func (k Keeper) SetBigPurgeSchedule(ctx sdk.Context, schedule BigPurgeSchedule) {
	store := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&schedule)
	store.Set([]byte("purge-schedule:"+schedule.DomainName), bz)
}

// InitializeBigPurgeSchedule sets up the default purge schedule for a new domain.
// Default: 90-day purge interval, 7-day announcement lead.
func (k Keeper) InitializeBigPurgeSchedule(ctx sdk.Context, domainName string) {
	now := ctx.BlockTime().Unix()
	schedule := BigPurgeSchedule{
		DomainName:       domainName,
		NextPurgeTime:    now + DefaultPurgeInterval,
		PurgeInterval:    DefaultPurgeInterval,
		AnnouncementLead: DefaultAnnouncementLead,
	}
	k.SetBigPurgeSchedule(ctx, schedule)
}

// ---------- Onboarding Request ----------

// GetOnboardingRequest retrieves a pending domain key registration request.
func (k Keeper) GetOnboardingRequest(ctx sdk.Context, domainName, requesterAddr string) (OnboardingRequest, bool) {
	store := ctx.KVStore(k.StoreKey)
	bz := store.Get([]byte("onboarding:" + domainName + ":" + requesterAddr))
	if bz == nil {
		return OnboardingRequest{}, false
	}
	var request OnboardingRequest
	k.cdc.MustUnmarshalLengthPrefixed(bz, &request)
	return request, true
}

// SetOnboardingRequest stores a domain key registration request.
func (k Keeper) SetOnboardingRequest(ctx sdk.Context, request OnboardingRequest) {
	store := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&request)
	store.Set([]byte("onboarding:"+request.DomainName+":"+request.RequesterAddr), bz)
}

// DeleteOnboardingRequest removes a completed or rejected onboarding request.
func (k Keeper) DeleteOnboardingRequest(ctx sdk.Context, domainName, requesterAddr string) {
	store := ctx.KVStore(k.StoreKey)
	store.Delete([]byte("onboarding:" + domainName + ":" + requesterAddr))
}

// ApproveOnboardingRequest validates a pending request and adds the domain
// key to the permission register. Only the domain admin can approve (WP S4).
func (k Keeper) ApproveOnboardingRequest(ctx sdk.Context, domainName, requesterAddr string, adminAddr sdk.AccAddress) error {
	request, exists := k.GetOnboardingRequest(ctx, domainName, requesterAddr)
	if !exists {
		return errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "onboarding request not found")
	}
	if request.Status != "pending" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "request is not pending: %s", request.Status)
	}

	// Verify caller is domain admin.
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "domain %s not found", domainName)
	}
	if !adminAddr.Equals(domain.Admin) {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "only domain admin can approve onboarding")
	}

	// Decode domain pub key from hex.
	domainPubKeyBytes, err := hex.DecodeString(request.DomainPubKeyHex)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid domain public key hex in request")
	}

	// JoinPermissionRegister validates key length, membership, and duplicates.
	if err := k.JoinPermissionRegister(ctx, domainName, requesterAddr, domainPubKeyBytes); err != nil {
		return err
	}

	// Update request status.
	request.Status = "approved"
	k.SetOnboardingRequest(ctx, request)
	return nil
}

// RejectOnboardingRequest rejects a pending onboarding request.
// Only the domain admin can reject (WP S4).
func (k Keeper) RejectOnboardingRequest(ctx sdk.Context, domainName, requesterAddr string, adminAddr sdk.AccAddress) error {
	request, exists := k.GetOnboardingRequest(ctx, domainName, requesterAddr)
	if !exists {
		return errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "onboarding request not found")
	}
	if request.Status != "pending" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "request is not pending: %s", request.Status)
	}

	// Verify caller is domain admin.
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "domain %s not found", domainName)
	}
	if !adminAddr.Equals(domain.Admin) {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "only domain admin can reject onboarding")
	}

	request.Status = "rejected"
	k.SetOnboardingRequest(ctx, request)
	return nil
}

// ---------- ZKP Verifying Key Storage (v0.3.0) ----------

// GetVerifyingKey retrieves the serialized Groth16 verifying key from the KV store.
func (k Keeper) GetVerifyingKey(ctx sdk.Context) ([]byte, bool) {
	store := ctx.KVStore(k.StoreKey)
	bz := store.Get([]byte("zkp:verifying-key"))
	if bz == nil {
		return nil, false
	}
	return bz, true
}

// SetVerifyingKey stores the serialized Groth16 verifying key.
func (k Keeper) SetVerifyingKey(ctx sdk.Context, vkBytes []byte) {
	store := ctx.KVStore(k.StoreKey)
	store.Set([]byte("zkp:verifying-key"), vkBytes)
}

// EnsureVerifyingKey returns the Groth16 verifying key, lazily initializing
// it on first use. For v0.3.0-dev (single-node testnet). Multi-node
// production requires a genesis-time VK or trusted setup ceremony.
func (k Keeper) EnsureVerifyingKey(ctx sdk.Context) ([]byte, error) {
	if vkBytes, found := k.GetVerifyingKey(ctx); found {
		return vkBytes, nil
	}

	// First use — run circuit setup and store the VK.
	keys, err := SetupMembershipCircuit()
	if err != nil {
		return nil, fmt.Errorf("ZKP circuit setup failed: %w", err)
	}
	vkBytes, err := SerializeVerifyingKey(keys.VerifyingKey)
	if err != nil {
		return nil, fmt.Errorf("verifying key serialization failed: %w", err)
	}
	k.SetVerifyingKey(ctx, vkBytes)
	return vkBytes, nil
}

// ---------- ZKP Identity Commitments (v0.3.0) ----------

// RegisterIdentityCommitment adds a MiMC commitment to the domain's
// identity commitment set and rebuilds the Merkle root.
// The caller must be a domain member. The commitment is not linked
// to the member's identity on-chain (WP S4 ZKP extension).
func (k Keeper) RegisterIdentityCommitment(ctx sdk.Context, domainName, memberAddr, commitmentHex string) error {
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "domain %s not found", domainName)
	}

	// Verify caller is a domain member.
	isMember := false
	for _, m := range domain.Members {
		if m == memberAddr {
			isMember = true
			break
		}
	}
	if !isMember {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "only domain members can register identity commitments")
	}

	// Validate commitment hex (must be 64 hex chars = 32 bytes).
	commitBytes, err := hex.DecodeString(commitmentHex)
	if err != nil || len(commitBytes) != 32 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "commitment must be 32 bytes hex-encoded (64 hex chars)")
	}

	// Check for duplicate commitment.
	for _, existing := range domain.IdentityCommits {
		if existing == commitmentHex {
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "commitment already registered")
		}
	}

	// Append commitment.
	domain.IdentityCommits = append(domain.IdentityCommits, commitmentHex)

	// Save current root to history before overwriting.
	if domain.MerkleRoot != "" {
		domain.MerkleRootHistory = append(domain.MerkleRootHistory, domain.MerkleRoot)
		if len(domain.MerkleRootHistory) > MerkleRootHistorySize {
			domain.MerkleRootHistory = domain.MerkleRootHistory[len(domain.MerkleRootHistory)-MerkleRootHistorySize:]
		}
	}

	// Rebuild Merkle root.
	root, err := k.computeMerkleRoot(domain.IdentityCommits)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "failed to compute Merkle root: "+err.Error())
	}
	domain.MerkleRoot = root

	// Persist domain.
	store := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	store.Set([]byte("domain:"+domainName), bz)
	return nil
}

// computeMerkleRoot builds a Merkle tree from hex-encoded commitments
// and returns the root as a hex string.
func (k Keeper) computeMerkleRoot(commitHexes []string) (string, error) {
	tree := NewMerkleTree(MerkleTreeDepth)
	leaves := make([][]byte, len(commitHexes))
	for i, h := range commitHexes {
		b, err := hex.DecodeString(h)
		if err != nil {
			return "", fmt.Errorf("invalid commitment hex at index %d: %w", i, err)
		}
		leaves[i] = b
	}
	if err := tree.BuildFromLeaves(leaves); err != nil {
		return "", err
	}
	return tree.GetRoot(), nil
}

// isAcceptedMerkleRoot checks if the given root hex matches the domain's current
// root or any root in the history window.
func isAcceptedMerkleRoot(domain Domain, rootHex string) bool {
	if domain.MerkleRoot == rootHex {
		return true
	}
	for _, h := range domain.MerkleRootHistory {
		if h == rootHex {
			return true
		}
	}
	return false
}

// ---------- Nullifier Store (v0.3.0) ----------

// IsNullifierUsed checks whether a nullifier has already been consumed
// in this domain (prevents ZKP double-voting).
func (k Keeper) IsNullifierUsed(ctx sdk.Context, domainName, nullifierHex string) bool {
	store := ctx.KVStore(k.StoreKey)
	return store.Has([]byte("nullifier:" + domainName + ":" + nullifierHex))
}

// SetNullifierUsed records a nullifier as consumed.
func (k Keeper) SetNullifierUsed(ctx sdk.Context, domainName, nullifierHex string, blockHeight int64) {
	record := NullifierRecord{
		DomainName:    domainName,
		NullifierHash: nullifierHex,
		UsedAtHeight:  blockHeight,
	}
	store := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&record)
	store.Set([]byte("nullifier:"+domainName+":"+nullifierHex), bz)
}

// PurgeNullifiers clears all nullifiers for a domain.
// Called during Big Purge when identity commitments are also wiped.
func (k Keeper) PurgeNullifiers(ctx sdk.Context, domainName string) {
	store := ctx.KVStore(k.StoreKey)
	prefix := []byte("nullifier:" + domainName + ":")
	end := prefixEnd(prefix)
	iter := store.Iterator(prefix, end)
	defer iter.Close()
	var keys [][]byte
	for ; iter.Valid(); iter.Next() {
		keys = append(keys, append([]byte{}, iter.Key()...))
	}
	for _, key := range keys {
		store.Delete(key)
	}
}
