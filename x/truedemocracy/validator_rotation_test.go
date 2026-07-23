package truedemocracy

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	rewards "truerepublic/treasury/keeper"
)

func rotationTestAddress(seed byte) sdk.AccAddress {
	return sdk.AccAddress(bytes.Repeat([]byte{seed}, 20))
}

func setupRotationValidator(t *testing.T, k Keeper, ctx sdk.Context) (sdk.AccAddress, Validator) {
	t.Helper()
	operator := rotationTestAddress(1)
	k.CreateDomain(ctx, "Rotation", operator, sdk.NewCoins())
	stake := sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, rewards.StakeMin))
	if err := k.RegisterValidator(ctx, operator.String(), testPubKey("rotation-old"), stake, "Rotation"); err != nil {
		t.Fatal(err)
	}
	validator, found := k.GetValidator(ctx, operator.String())
	if !found {
		t.Fatal("validator missing")
	}
	return operator, validator
}

func TestRotateValidatorKeyPreservesClaimsAndActivationGuard(t *testing.T) {
	k, baseCtx := setupKeeper(t)
	ctx := baseCtx.WithBlockHeight(10)
	operator, before := setupRotationValidator(t, k, ctx)
	oldKey := append([]byte(nil), before.PubKey...)
	newKey := testPubKey("rotation-new")

	rotatedOld, err := k.RotateValidatorKey(ctx, operator, operator.String(), newKey)
	if err != nil {
		t.Fatalf("rotate: %v", err)
	}
	if !bytes.Equal(rotatedOld, oldKey) {
		t.Fatal("returned old key mismatch")
	}
	after, found := k.GetValidator(ctx, operator.String())
	if !found || !bytes.Equal(after.PubKey, newKey) {
		t.Fatal("new key was not installed")
	}
	before.PubKey = newKey
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("rotation changed validator claims:\n got %#v\nwant %#v", after, before)
	}
	if !k.IsValidatorKeyRevoked(ctx, oldKey) {
		t.Fatal("old key was not permanently revoked")
	}
	if got, found := k.GetValidatorByPubKey(ctx, oldKey); !found || got.OperatorAddr != operator.String() {
		t.Fatal("old key lost slashing attribution during activation window")
	}
	if got, found := k.GetValidatorByPubKey(ctx, newKey); !found || got.OperatorAddr != operator.String() {
		t.Fatal("new key reverse index missing")
	}

	updates := k.BuildValidatorUpdates(ctx)
	assertValidatorUpdatePower(t, updates, oldKey, 0)
	assertValidatorUpdatePower(t, updates, newKey, before.Power)
	if _, err := k.RotateValidatorKey(ctx, operator, operator.String(), testPubKey("rotation-too-soon")); err == nil {
		t.Fatal("second rotation succeeded inside activation window")
	}
	if err := k.HandleDowntime(ctx, oldKey); err != nil {
		t.Fatalf("old-key downtime attribution failed: %v", err)
	}
	tracked, _ := k.GetValidator(ctx, operator.String())
	if tracked.MissedBlocks != before.MissedBlocks+1 {
		t.Fatalf("missed blocks = %d, want %d", tracked.MissedBlocks, before.MissedBlocks+1)
	}

	k.BuildValidatorUpdates(ctx.WithBlockHeight(11))
	if _, found := k.GetValidatorByPubKey(ctx, oldKey); !found {
		t.Fatal("old key attribution cleared before H+2")
	}
	k.BuildValidatorUpdates(ctx.WithBlockHeight(12))
	if _, found := k.GetValidatorByPubKey(ctx, oldKey); found {
		t.Fatal("old key attribution remained after H+2")
	}
	if _, err := k.RotateValidatorKey(ctx.WithBlockHeight(12), operator, operator.String(), testPubKey("rotation-third")); err != nil {
		t.Fatalf("rotation remained blocked after activation: %v", err)
	}
}

func TestRotateValidatorKeyFailsClosedAndAtomically(t *testing.T) {
	k, ctx := setupKeeper(t)
	operator, before := setupRotationValidator(t, k, ctx)
	other := rotationTestAddress(2)
	domain, _ := k.GetDomain(ctx, "Rotation")
	domain.Members = append(domain.Members, other.String())
	ctx.KVStore(k.StoreKey).Set([]byte("domain:Rotation"), k.cdc.MustMarshalLengthPrefixed(&domain))
	duplicate := testPubKey("rotation-duplicate")
	if err := k.RegisterValidator(ctx, other.String(), duplicate, before.Stake, "Rotation"); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		sender sdk.AccAddress
		key    []byte
	}{
		{name: "spoofed operator", sender: other, key: testPubKey("spoof")},
		{name: "malformed", sender: operator, key: []byte{1}},
		{name: "same key", sender: operator, key: before.PubKey},
		{name: "active duplicate", sender: operator, key: duplicate},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := k.RotateValidatorKey(ctx, tc.sender, operator.String(), tc.key); err == nil {
				t.Fatal("expected rotation error")
			}
			got, _ := k.GetValidator(ctx, operator.String())
			if !reflect.DeepEqual(got, before) {
				t.Fatal("failed rotation mutated validator")
			}
		})
	}

	oldKey, err := k.RotateValidatorKey(ctx, operator, operator.String(), testPubKey("valid-new"))
	if err != nil {
		t.Fatal(err)
	}
	third := rotationTestAddress(3)
	domain, _ = k.GetDomain(ctx, "Rotation")
	domain.Members = append(domain.Members, third.String())
	ctx.KVStore(k.StoreKey).Set([]byte("domain:Rotation"), k.cdc.MustMarshalLengthPrefixed(&domain))
	if err := k.RegisterValidator(ctx, third.String(), oldKey, before.Stake, "Rotation"); err == nil {
		t.Fatal("revoked key registration succeeded")
	}
}

func TestValidatorRotationGenesisRoundTrip(t *testing.T) {
	am1, k1, ctx1 := setupModuleForGenesis(t)
	ctx1 = ctx1.WithBlockHeight(20)
	operator, _ := setupRotationValidator(t, k1, ctx1)
	oldKey, err := k1.RotateValidatorKey(ctx1, operator, operator.String(), testPubKey("roundtrip-new"))
	if err != nil {
		t.Fatal(err)
	}
	exported := am1.ExportGenesis(ctx1, nil)
	var genesis GenesisState
	if err := json.Unmarshal(exported, &genesis); err != nil {
		t.Fatal(err)
	}
	if len(genesis.RevokedValidatorKeys) != 1 || len(genesis.PendingValidatorRotations) != 1 {
		t.Fatalf("exported revoked/pending = %d/%d, want 1/1", len(genesis.RevokedValidatorKeys), len(genesis.PendingValidatorRotations))
	}
	if err := ValidateGenesisState(genesis); err != nil {
		t.Fatalf("exported genesis invalid: %v", err)
	}

	am2, k2, ctx2 := setupModuleForGenesis(t)
	ctx2 = ctx2.WithBlockHeight(20)
	am2.InitGenesis(ctx2, nil, exported)
	if !k2.IsValidatorKeyRevoked(ctx2, oldKey) {
		t.Fatal("revocation did not survive genesis round trip")
	}
	if got, found := k2.GetValidatorByPubKey(ctx2, oldKey); !found || got.OperatorAddr != operator.String() {
		t.Fatal("pending old-key attribution did not survive genesis round trip")
	}
}

func TestMsgRotateValidatorKeyValidationAndDescriptor(t *testing.T) {
	operator := rotationTestAddress(4)
	msg := MsgRotateValidatorKey{
		Sender:       operator,
		OperatorAddr: operator.String(),
		NewPubKey:    hex.EncodeToString(testPubKey("message-key")),
	}
	if err := msg.ValidateBasic(); err != nil {
		t.Fatalf("valid message rejected: %v", err)
	}
	if bz, indexes := msg.Descriptor(); len(bz) == 0 || len(indexes) == 0 {
		t.Fatal("message descriptor missing")
	}
	msg.Sender = rotationTestAddress(5)
	if err := msg.ValidateBasic(); err == nil {
		t.Fatal("spoofed operator accepted")
	}
	msg.Sender = operator
	msg.NewPubKey = "ABCDEF"
	if err := msg.ValidateBasic(); err == nil {
		t.Fatal("malformed key accepted")
	}
}

func TestMsgServerRotateValidatorKeyEmitsDeterministicEvent(t *testing.T) {
	k, baseCtx := setupKeeper(t)
	ctx := baseCtx.WithBlockHeight(7).WithEventManager(sdk.NewEventManager())
	operator, before := setupRotationValidator(t, k, ctx)
	newKey := testPubKey("event-new")
	msg := &MsgRotateValidatorKey{
		Sender:       operator,
		OperatorAddr: operator.String(),
		NewPubKey:    hex.EncodeToString(newKey),
	}
	if _, err := NewMsgServer(k).RotateValidatorKey(sdk.WrapSDKContext(ctx), msg); err != nil {
		t.Fatal(err)
	}
	events := ctx.EventManager().Events()
	if len(events) != 1 || events[0].Type != "rotate_validator_key" {
		t.Fatalf("events = %#v", events)
	}
	want := [][2]string{
		{"operator", operator.String()},
		{"old_pubkey", hex.EncodeToString(before.PubKey)},
		{"new_pubkey", hex.EncodeToString(newKey)},
		{"activation_height", "9"},
	}
	if len(events[0].Attributes) != len(want) {
		t.Fatalf("attribute count = %d, want %d", len(events[0].Attributes), len(want))
	}
	for i, attribute := range events[0].Attributes {
		if attribute.Key != want[i][0] || attribute.Value != want[i][1] {
			t.Fatalf("attribute %d = %q:%q, want %q:%q", i, attribute.Key, attribute.Value, want[i][0], want[i][1])
		}
	}
}

func TestValidateGenesisRejectsMalformedValidatorRotationState(t *testing.T) {
	operator := rotationTestAddress(6)
	oldKey := testPubKey("genesis-old")
	newKey := testPubKey("genesis-new")
	base := GenesisState{
		Domains: []Domain{{
			Name:          "Rotation",
			Admin:         operator,
			Members:       []string{operator.String()},
			Treasury:      sdk.NewCoins(),
			Issues:        []Issue{},
			PermissionReg: []string{},
		}},
		Validators: []GenesisValidator{{
			OperatorAddr: operator.String(),
			PubKey:       newKey,
			Stake:        rewards.StakeMin,
			Domain:       "Rotation",
		}},
		RevokedValidatorKeys: []RevokedValidatorKey{{
			PubKey: oldKey, OperatorAddr: operator.String(), RevokedAtHeight: 4,
		}},
		PendingValidatorRotations: []PendingValidatorKeyRotation{{
			OperatorAddr: operator.String(), OldPubKey: oldKey, NewPubKey: newKey,
			StartedHeight: 4, ClearAfterHeight: 6,
		}},
	}
	if err := ValidateGenesisState(base); err != nil {
		t.Fatalf("valid rotation genesis rejected: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*GenesisState)
	}{
		{name: "active key revoked", mutate: func(g *GenesisState) { g.RevokedValidatorKeys[0].PubKey = newKey }},
		{name: "old key not revoked", mutate: func(g *GenesisState) { g.RevokedValidatorKeys = nil }},
		{name: "new key not active", mutate: func(g *GenesisState) { g.PendingValidatorRotations[0].NewPubKey = testPubKey("not-active") }},
		{name: "window too short", mutate: func(g *GenesisState) { g.PendingValidatorRotations[0].ClearAfterHeight = 5 }},
		{name: "duplicate pending", mutate: func(g *GenesisState) {
			g.PendingValidatorRotations = append(g.PendingValidatorRotations, g.PendingValidatorRotations[0])
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			candidate := base
			candidate.Validators = append([]GenesisValidator(nil), base.Validators...)
			candidate.RevokedValidatorKeys = append([]RevokedValidatorKey(nil), base.RevokedValidatorKeys...)
			candidate.PendingValidatorRotations = append([]PendingValidatorKeyRotation(nil), base.PendingValidatorRotations...)
			tc.mutate(&candidate)
			if err := ValidateGenesisState(candidate); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestValidateGenesisBootstrapOperatorCeremonyInput(t *testing.T) {
	first := rotationTestAddress(7).String()
	second := rotationTestAddress(8).String()
	if err := ValidateGenesisState(GenesisState{BootstrapOperatorAddresses: []string{first, second}}); err != nil {
		t.Fatalf("valid bootstrap operators rejected: %v", err)
	}
	if err := ValidateGenesisState(GenesisState{BootstrapOperatorAddresses: []string{first, first}}); err == nil {
		t.Fatal("duplicate bootstrap operator accepted")
	}
	if err := ValidateGenesisState(GenesisState{BootstrapOperatorAddresses: []string{"not-an-address"}}); err == nil {
		t.Fatal("malformed bootstrap operator accepted")
	}
	genesis := GenesisState{BootstrapOperatorAddresses: []string{first}, Domains: []Domain{{Name: "materialized"}}}
	if err := ValidateGenesisState(genesis); err == nil {
		t.Fatal("bootstrap operators accepted with materialized domain state")
	}
}
