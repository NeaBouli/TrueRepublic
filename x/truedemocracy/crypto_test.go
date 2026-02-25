package truedemocracy

import (
	"encoding/hex"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
)

func TestConstructOnboardingMessage(t *testing.T) {
	msg := ConstructOnboardingMessage("alice", "TestDomain", "aabbccdd")
	expected := "ONBOARD:alice:TestDomain:aabbccdd"
	if string(msg) != expected {
		t.Errorf("message = %s, want %s", string(msg), expected)
	}
}

func TestVerifyOnboardingSignature(t *testing.T) {
	globalPriv := ed25519.GenPrivKeyFromSecret([]byte("global-key"))
	globalPub := &ed25519.PubKey{Key: globalPriv.PubKey().Bytes()}

	domainPriv := ed25519.GenPrivKeyFromSecret([]byte("domain-key"))
	domainPubHex := hex.EncodeToString(domainPriv.PubKey().Bytes())

	// Sign the onboarding message with the global key.
	message := ConstructOnboardingMessage("alice", "TestDomain", domainPubHex)
	sig, err := globalPriv.Sign(message)
	if err != nil {
		t.Fatalf("failed to sign: %v", err)
	}

	t.Run("valid signature", func(t *testing.T) {
		err := VerifyOnboardingSignature("alice", "TestDomain", domainPubHex, globalPub, sig)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("wrong key rejects", func(t *testing.T) {
		wrongPriv := ed25519.GenPrivKeyFromSecret([]byte("wrong-key"))
		wrongPub := &ed25519.PubKey{Key: wrongPriv.PubKey().Bytes()}
		err := VerifyOnboardingSignature("alice", "TestDomain", domainPubHex, wrongPub, sig)
		if err == nil {
			t.Fatal("expected error for wrong key")
		}
	})

	t.Run("wrong message rejects", func(t *testing.T) {
		// Signature was for "alice", verifying for "bob" should fail.
		err := VerifyOnboardingSignature("bob", "TestDomain", domainPubHex, globalPub, sig)
		if err == nil {
			t.Fatal("expected error for wrong message")
		}
	})
}

func TestVerifyKeysAreDifferent(t *testing.T) {
	t.Run("different keys pass", func(t *testing.T) {
		err := VerifyKeysAreDifferent("aabbcc", "ddeeff")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("same keys fail", func(t *testing.T) {
		err := VerifyKeysAreDifferent("aabbcc", "aabbcc")
		if err == nil {
			t.Fatal("expected error for same keys")
		}
	})
}

func TestParseEd25519PubKeyFromHex(t *testing.T) {
	privKey := ed25519.GenPrivKeyFromSecret([]byte("parse-test"))
	pubKeyHex := hex.EncodeToString(privKey.PubKey().Bytes())

	t.Run("valid hex", func(t *testing.T) {
		parsed, err := ParseEd25519PubKeyFromHex(pubKeyHex)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if hex.EncodeToString(parsed.Bytes()) != pubKeyHex {
			t.Error("parsed key doesn't match original")
		}
	})

	t.Run("invalid hex", func(t *testing.T) {
		_, err := ParseEd25519PubKeyFromHex("not-valid-hex!!!")
		if err == nil {
			t.Fatal("expected error for invalid hex")
		}
	})

	t.Run("wrong size", func(t *testing.T) {
		_, err := ParseEd25519PubKeyFromHex("0102030405")
		if err == nil {
			t.Fatal("expected error for wrong size")
		}
	})
}

func TestDeriveAddressFromPubKey(t *testing.T) {
	privKey := ed25519.GenPrivKeyFromSecret([]byte("addr-test"))
	pubKey := &ed25519.PubKey{Key: privKey.PubKey().Bytes()}

	addr := DeriveAddressFromPubKey(pubKey)
	if len(addr) != 20 {
		t.Errorf("address length = %d, want 20", len(addr))
	}
}
