package truedemocracy

import (
	"encoding/hex"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ---------- MsgAddMember ValidateBasic ----------

func TestMsgAddMemberValidateBasic(t *testing.T) {
	t.Run("valid message", func(t *testing.T) {
		msg := MsgAddMember{
			Sender:     sdk.AccAddress("admin1"),
			DomainName: "TestDomain",
			NewMember:  "alice",
		}
		if err := msg.ValidateBasic(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty domain rejected", func(t *testing.T) {
		msg := MsgAddMember{
			Sender:     sdk.AccAddress("admin1"),
			DomainName: "",
			NewMember:  "alice",
		}
		if err := msg.ValidateBasic(); err == nil {
			t.Fatal("expected error for empty domain_name")
		}
	})

	t.Run("empty member rejected", func(t *testing.T) {
		msg := MsgAddMember{
			Sender:     sdk.AccAddress("admin1"),
			DomainName: "TestDomain",
			NewMember:  "",
		}
		if err := msg.ValidateBasic(); err == nil {
			t.Fatal("expected error for empty new_member")
		}
	})
}

// ---------- MsgOnboardToDomain ValidateBasic ----------

func TestMsgOnboardToDomainValidateBasic(t *testing.T) {
	t.Run("valid message", func(t *testing.T) {
		msg := MsgOnboardToDomain{
			Sender:          sdk.AccAddress("alice"),
			DomainName:      "TestDomain",
			DomainPubKeyHex: "aabb",
			GlobalPubKeyHex: "ccdd",
			SignatureHex:    "eeff",
		}
		if err := msg.ValidateBasic(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty domain rejected", func(t *testing.T) {
		msg := MsgOnboardToDomain{
			Sender:          sdk.AccAddress("alice"),
			DomainName:      "",
			DomainPubKeyHex: "aabb",
			GlobalPubKeyHex: "ccdd",
			SignatureHex:    "eeff",
		}
		if err := msg.ValidateBasic(); err == nil {
			t.Fatal("expected error for empty domain_name")
		}
	})

	t.Run("empty keys rejected", func(t *testing.T) {
		msg := MsgOnboardToDomain{
			Sender:          sdk.AccAddress("alice"),
			DomainName:      "TestDomain",
			DomainPubKeyHex: "",
			GlobalPubKeyHex: "ccdd",
			SignatureHex:    "eeff",
		}
		if err := msg.ValidateBasic(); err == nil {
			t.Fatal("expected error for empty domain_pub_key_hex")
		}
	})
}

// ---------- Keeper AddMember ----------

func TestAddMember(t *testing.T) {
	k, ctx := setupKeeper(t)

	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "TestDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))

	t.Run("admin can add member", func(t *testing.T) {
		err := k.AddMember(ctx, "TestDomain", "alice", admin)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		domain, _ := k.GetDomain(ctx, "TestDomain")
		found := false
		for _, m := range domain.Members {
			if m == "alice" {
				found = true
			}
		}
		if !found {
			t.Fatal("alice should be in Members after AddMember")
		}
	})

	t.Run("non-admin rejected", func(t *testing.T) {
		err := k.AddMember(ctx, "TestDomain", "bob", sdk.AccAddress("not-admin"))
		if err == nil {
			t.Fatal("expected error for non-admin")
		}
	})

	t.Run("duplicate member rejected", func(t *testing.T) {
		err := k.AddMember(ctx, "TestDomain", "alice", admin)
		if err == nil {
			t.Fatal("expected error for duplicate member")
		}
	})

	t.Run("unknown domain rejected", func(t *testing.T) {
		err := k.AddMember(ctx, "NoDomain", "bob", admin)
		if err == nil {
			t.Fatal("expected error for unknown domain")
		}
	})
}

// ---------- Two-Step Onboarding Full Flow ----------

func TestTwoStepOnboardingFullFlow(t *testing.T) {
	k, ctx := setupKeeper(t)

	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "OnboardDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))

	// Step 1: Admin adds alice.
	err := k.AddMember(ctx, "OnboardDomain", "alice", admin)
	if err != nil {
		t.Fatalf("step 1 failed: %v", err)
	}

	// Step 2: Alice registers a domain key.
	aliceDomainKey := domainKey("alice-onboard-domain-key")
	err = k.JoinPermissionRegister(ctx, "OnboardDomain", "alice", aliceDomainKey.PubKey().Bytes())
	if err != nil {
		t.Fatalf("step 2 failed: %v", err)
	}

	// Verify alice can now vote.
	aliceHex := hex.EncodeToString(aliceDomainKey.PubKey().Bytes())
	if !k.IsKeyAuthorized(ctx, "OnboardDomain", aliceHex) {
		t.Fatal("alice's domain key should be authorized after full onboarding")
	}
}

func TestOnboardingWithoutMembershipFails(t *testing.T) {
	k, ctx := setupKeeper(t)

	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "OnboardDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))

	// Skip step 1 — try to register domain key directly (not a member).
	outsiderKey := domainKey("outsider-key")
	err := k.JoinPermissionRegister(ctx, "OnboardDomain", "outsider", outsiderKey.PubKey().Bytes())
	if err == nil {
		t.Fatal("expected error — outsider is not a member, step 1 was skipped")
	}
}

// ---------- MsgServer OnboardToDomain ----------

func TestMsgServerOnboardToDomain(t *testing.T) {
	k, ctx := setupKeeper(t)
	srv := NewMsgServer(k)

	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "SigDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))

	// Add alice as member.
	k.AddMember(ctx, "SigDomain", sdk.AccAddress("alice").String(), admin)

	// Generate keys.
	globalPriv := ed25519.GenPrivKeyFromSecret([]byte("alice-global"))
	globalPub := &ed25519.PubKey{Key: globalPriv.PubKey().Bytes()}
	globalPubHex := hex.EncodeToString(globalPub.Bytes())

	domainPriv := ed25519.GenPrivKeyFromSecret([]byte("alice-domain"))
	domainPubHex := hex.EncodeToString(domainPriv.PubKey().Bytes())

	// Sign the onboarding message.
	message := ConstructOnboardingMessage(sdk.AccAddress("alice").String(), "SigDomain", domainPubHex)
	sig, err := globalPriv.Sign(message)
	if err != nil {
		t.Fatalf("failed to sign: %v", err)
	}
	sigHex := hex.EncodeToString(sig)

	t.Run("valid onboarding", func(t *testing.T) {
		msg := &MsgOnboardToDomain{
			Sender:          sdk.AccAddress("alice"),
			DomainName:      "SigDomain",
			DomainPubKeyHex: domainPubHex,
			GlobalPubKeyHex: globalPubHex,
			SignatureHex:    sigHex,
		}
		_, err := srv.OnboardToDomain(ctx, msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify key is now authorized.
		if !k.IsKeyAuthorized(ctx, "SigDomain", domainPubHex) {
			t.Fatal("domain key should be authorized after onboarding")
		}
	})

	t.Run("same key as global rejected", func(t *testing.T) {
		msg := &MsgOnboardToDomain{
			Sender:          sdk.AccAddress("alice"),
			DomainName:      "SigDomain",
			DomainPubKeyHex: globalPubHex, // same as global!
			GlobalPubKeyHex: globalPubHex,
			SignatureHex:    sigHex,
		}
		_, err := srv.OnboardToDomain(ctx, msg)
		if err == nil {
			t.Fatal("expected error for same domain and global key")
		}
	})

	t.Run("invalid signature rejected", func(t *testing.T) {
		msg := &MsgOnboardToDomain{
			Sender:          sdk.AccAddress("alice"),
			DomainName:      "SigDomain",
			DomainPubKeyHex: domainPubHex,
			GlobalPubKeyHex: globalPubHex,
			SignatureHex:    "deadbeef", // wrong signature
		}
		_, err := srv.OnboardToDomain(ctx, msg)
		if err == nil {
			t.Fatal("expected error for invalid signature")
		}
	})
}

// ---------- MsgServer ApproveOnboarding / RejectOnboarding ----------

func TestMsgServerApproveOnboarding(t *testing.T) {
	k, ctx := setupKeeper(t)
	srv := NewMsgServer(k)

	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "ApproveDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))

	// Add alice as member.
	k.AddMember(ctx, "ApproveDomain", "alice", admin)

	// Create a pending onboarding request.
	aliceDomainKey := domainKey("alice-approve-key")
	domainPubHex := hex.EncodeToString(aliceDomainKey.PubKey().Bytes())
	request := OnboardingRequest{
		DomainName:      "ApproveDomain",
		RequesterAddr:   "alice",
		DomainPubKeyHex: domainPubHex,
		RequestedAt:     ctx.BlockTime().Unix(),
		Status:          "pending",
	}
	k.SetOnboardingRequest(ctx, request)

	t.Run("admin approves", func(t *testing.T) {
		msg := &MsgApproveOnboarding{
			Sender:        admin,
			DomainName:    "ApproveDomain",
			RequesterAddr: "alice",
		}
		_, err := srv.ApproveOnboarding(ctx, msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !k.IsKeyAuthorized(ctx, "ApproveDomain", domainPubHex) {
			t.Fatal("key should be authorized after approval")
		}
	})
}

func TestMsgServerRejectOnboarding(t *testing.T) {
	k, ctx := setupKeeper(t)
	srv := NewMsgServer(k)

	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "RejectDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100_000)))

	// Create a pending request.
	request := OnboardingRequest{
		DomainName:      "RejectDomain",
		RequesterAddr:   "bob",
		DomainPubKeyHex: "aabbccdd",
		RequestedAt:     ctx.BlockTime().Unix(),
		Status:          "pending",
	}
	k.SetOnboardingRequest(ctx, request)

	t.Run("admin rejects", func(t *testing.T) {
		msg := &MsgRejectOnboarding{
			Sender:        admin,
			DomainName:    "RejectDomain",
			RequesterAddr: "bob",
		}
		_, err := srv.RejectOnboarding(ctx, msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		updated, _ := k.GetOnboardingRequest(ctx, "RejectDomain", "bob")
		if updated.Status != "rejected" {
			t.Errorf("status = %s, want rejected", updated.Status)
		}
	})

	t.Run("non-admin rejected", func(t *testing.T) {
		// Create another pending request.
		req2 := OnboardingRequest{
			DomainName:      "RejectDomain",
			RequesterAddr:   "charlie",
			DomainPubKeyHex: "eeff0011",
			RequestedAt:     ctx.BlockTime().Unix(),
			Status:          "pending",
		}
		k.SetOnboardingRequest(ctx, req2)

		msg := &MsgRejectOnboarding{
			Sender:        sdk.AccAddress("not-admin"),
			DomainName:    "RejectDomain",
			RequesterAddr: "charlie",
		}
		_, err := srv.RejectOnboarding(ctx, msg)
		if err == nil {
			t.Fatal("expected error for non-admin rejection")
		}
	})
}
