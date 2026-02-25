package truedemocracy

import (
	"encoding/hex"
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ConstructOnboardingMessage creates the standard message format for
// two-step onboarding proofs (WP S4).
// Format: "ONBOARD:{requester}:{domain}:{domain_pubkey_hex}"
func ConstructOnboardingMessage(requesterAddr, domainName, domainPubKeyHex string) []byte {
	return []byte(fmt.Sprintf("ONBOARD:%s:%s:%s", requesterAddr, domainName, domainPubKeyHex))
}

// VerifyOnboardingSignature verifies that the global key signed the
// onboarding message, proving the requester controls their account.
// The signature is over the raw message bytes (ed25519 handles internal hashing).
func VerifyOnboardingSignature(
	requesterAddr, domainName, domainPubKeyHex string,
	globalPubKey *ed25519.PubKey,
	signatureBytes []byte,
) error {
	message := ConstructOnboardingMessage(requesterAddr, domainName, domainPubKeyHex)
	if !globalPubKey.VerifySignature(message, signatureBytes) {
		return fmt.Errorf("invalid signature: global key does not verify onboarding message")
	}
	return nil
}

// VerifyKeysAreDifferent ensures the domain key is not the same as the
// global key. Using the same key would break anonymity (WP S4).
func VerifyKeysAreDifferent(globalPubKeyHex, domainPubKeyHex string) error {
	if globalPubKeyHex == domainPubKeyHex {
		return fmt.Errorf("domain key must be different from global key (anonymity requirement)")
	}
	return nil
}

// ParseEd25519PubKeyFromHex parses a hex-encoded ed25519 public key.
func ParseEd25519PubKeyFromHex(pubKeyHex string) (*ed25519.PubKey, error) {
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid hex: %w", err)
	}
	if len(pubKeyBytes) != ed25519.PubKeySize {
		return nil, fmt.Errorf("invalid public key size: got %d, expected %d", len(pubKeyBytes), ed25519.PubKeySize)
	}
	return &ed25519.PubKey{Key: pubKeyBytes}, nil
}

// DeriveAddressFromPubKey derives an SDK address from an ed25519 public key.
func DeriveAddressFromPubKey(pubKey *ed25519.PubKey) sdk.AccAddress {
	return sdk.AccAddress(pubKey.Address())
}
