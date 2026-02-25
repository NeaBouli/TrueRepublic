package truedemocracy

import (
	"encoding/hex"

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
