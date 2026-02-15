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
