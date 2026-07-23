package truedemocracy

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
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

	rotatedOld, err := k.RotateValidatorKey(ctx, operator, operator.String(), oldKey, newKey)
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
	if _, err := k.RotateValidatorKey(ctx, operator, operator.String(), newKey, testPubKey("rotation-too-soon")); err == nil {
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
	if got, found := k.GetValidatorForDoubleSignEvidence(ctx.WithBlockHeight(12), oldKey); !found || got.OperatorAddr != operator.String() {
		t.Fatal("permanently revoked key lost delayed evidence attribution")
	}
	if _, err := k.RotateValidatorKey(ctx.WithBlockHeight(12), operator, operator.String(), newKey, testPubKey("rotation-third")); err != nil {
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
			if _, err := k.RotateValidatorKey(ctx, tc.sender, operator.String(), before.PubKey, tc.key); err == nil {
				t.Fatal("expected rotation error")
			}
			got, _ := k.GetValidator(ctx, operator.String())
			if !reflect.DeepEqual(got, before) {
				t.Fatal("failed rotation mutated validator")
			}
		})
	}
	if _, err := k.RotateValidatorKey(ctx, operator, operator.String(), testPubKey("stale-old"), testPubKey("stale-new")); err == nil {
		t.Fatal("stale expected old key was accepted")
	}
	inactive := before
	inactive.Jailed = true
	inactive.Power = 0
	k.SetValidator(ctx, inactive)
	if _, err := k.RotateValidatorKey(ctx, operator, operator.String(), before.PubKey, testPubKey("inactive-new")); err == nil {
		t.Fatal("inactive zero-power validator rotation succeeded")
	}
	k.SetValidator(ctx, before)

	oldKey, err := k.RotateValidatorKey(ctx, operator, operator.String(), before.PubKey, testPubKey("valid-new"))
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

func TestRotationThenSameBlockInactivationOnlyRemovesOldKey(t *testing.T) {
	k, baseCtx := setupKeeper(t)
	ctx := baseCtx.WithBlockHeight(30)
	operator, before := setupRotationValidator(t, k, ctx)
	newKey := testPubKey("same-block-inactive-new")
	if _, err := k.RotateValidatorKey(ctx, operator, operator.String(), before.PubKey, newKey); err != nil {
		t.Fatal(err)
	}
	rotated, _ := k.GetValidator(ctx, operator.String())
	k.QueueValidatorPowerZero(ctx, rotated)
	rotated.Jailed = true
	rotated.Power = 0
	k.SetValidator(ctx, rotated)
	updates := k.BuildValidatorUpdates(ctx)
	assertValidatorUpdatePower(t, updates, before.PubKey, 0)
	cometSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{
		cmttypes.NewValidator(cmted25519.PubKey(before.PubKey), before.Power),
		cmttypes.NewValidator(cmted25519.PubKey(testPubKey("same-block-stable")), 1),
	})
	if err := cometSet.UpdateWithChangeSet(cometChanges(updates)); err != nil {
		t.Fatalf("Comet rejected same-block inactivation updates: %v", err)
	}
	for _, update := range updates {
		if bytes.Equal(update.PubKey.GetEd25519(), newKey) {
			t.Fatalf("never-active replacement key received invalid power-zero update: %#v", update)
		}
	}
	if later := k.BuildValidatorUpdates(ctx.WithBlockHeight(31)); len(later) != 0 {
		t.Fatalf("never-active replacement emitted a later removal: %#v", later)
	}
}

func TestRotationThenNextBlockInactivationDefersNewKeyRemovalOnce(t *testing.T) {
	k, baseCtx := setupKeeper(t)
	ctx := baseCtx.WithBlockHeight(40)
	operator, before := setupRotationValidator(t, k, ctx)
	newKey := testPubKey("next-block-inactive-new")
	if _, err := k.RotateValidatorKey(ctx, operator, operator.String(), before.PubKey, newKey); err != nil {
		t.Fatal(err)
	}
	initial := k.BuildValidatorUpdates(ctx)
	assertValidatorUpdatePower(t, initial, before.PubKey, 0)
	assertValidatorUpdatePower(t, initial, newKey, before.Power)
	stableKey := testPubKey("stable-comet-validator")
	cometSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{
		cmttypes.NewValidator(cmted25519.PubKey(before.PubKey), before.Power),
		cmttypes.NewValidator(cmted25519.PubKey(stableKey), 1),
	})
	if err := cometSet.UpdateWithChangeSet(cometChanges(initial)); err != nil {
		t.Fatalf("Comet rejected normal rotation update set: %v", err)
	}

	nextCtx := ctx.WithBlockHeight(41)
	rotated, _ := k.GetValidator(nextCtx, operator.String())
	k.QueueValidatorPowerZero(nextCtx, rotated)
	rotated.Jailed = true
	rotated.Power = 0
	k.SetValidator(nextCtx, rotated)
	deferred := k.BuildValidatorUpdates(nextCtx)
	assertValidatorUpdatePower(t, deferred, newKey, 0)
	if err := cometSet.UpdateWithChangeSet(cometChanges(deferred)); err != nil {
		t.Fatalf("Comet rejected deferred replacement removal: %v", err)
	}
	if repeated := k.BuildValidatorUpdates(ctx.WithBlockHeight(42)); len(repeated) != 0 {
		t.Fatalf("deferred replacement removal repeated: %#v", repeated)
	}
}

func cometChanges(updates []abci.ValidatorUpdate) []*cmttypes.Validator {
	changes := make([]*cmttypes.Validator, 0, len(updates))
	for _, update := range updates {
		changes = append(changes, cmttypes.NewValidator(cmted25519.PubKey(update.PubKey.GetEd25519()), update.Power))
	}
	return changes
}

func TestRuntimeValidatorAuthoritySeparation(t *testing.T) {
	k, ctx := setupKeeper(t)
	selfKey := testPubKey("self-coupled-runtime")
	selfOperator := sdk.AccAddress((&ed25519.PubKey{Key: selfKey}).Address())
	k.CreateDomain(ctx, "Coupled", selfOperator, sdk.NewCoins())
	stake := sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, rewards.StakeMin))
	if err := k.RegisterValidator(ctx, selfOperator.String(), selfKey, stake, "Coupled"); err == nil {
		t.Fatal("self-coupled runtime validator registration succeeded")
	}
	legacy := Validator{
		OperatorAddr: selfOperator.String(), PubKey: selfKey, Stake: stake,
		Domains: []string{"Coupled"}, Power: 1,
	}
	k.SetValidator(ctx, legacy)
	ctx.KVStore(k.StoreKey).Set(valPubKeyKey(selfKey), []byte(selfOperator.String()))
	if _, err := k.RotateValidatorKey(ctx, selfOperator, selfOperator.String(), selfKey, testPubKey("legacy-migration-attempt")); err == nil {
		t.Fatal("legacy coupled validator rotated without an explicit authority migration")
	}

	operator, before := setupRotationValidator(t, k, ctx)
	collisionKey := testPubKey("cross-coupled-runtime")
	collisionOperator := sdk.AccAddress((&ed25519.PubKey{Key: collisionKey}).Address())
	domain, _ := k.GetDomain(ctx, "Rotation")
	domain.Members = append(domain.Members, collisionOperator.String())
	ctx.KVStore(k.StoreKey).Set([]byte("domain:Rotation"), k.cdc.MustMarshalLengthPrefixed(&domain))
	if err := k.RegisterValidator(ctx, collisionOperator.String(), testPubKey("safe-other-consensus"), stake, "Rotation"); err != nil {
		t.Fatal(err)
	}
	if _, err := k.RotateValidatorKey(ctx, operator, operator.String(), before.PubKey, collisionKey); err == nil {
		t.Fatal("rotation coupled a consensus key to another validator operator")
	}
	rotatedKey := testPubKey("runtime-revocation-target")
	if _, err := k.RotateValidatorKey(ctx, operator, operator.String(), before.PubKey, rotatedKey); err != nil {
		t.Fatal(err)
	}
	k.BuildValidatorUpdates(ctx)
	k.BuildValidatorUpdates(ctx.WithBlockHeight(ctx.BlockHeight() + 2))
	revokedDerivedOperator := sdk.AccAddress((&ed25519.PubKey{Key: before.PubKey}).Address())
	domain, _ = k.GetDomain(ctx, "Rotation")
	domain.Members = append(domain.Members, revokedDerivedOperator.String())
	ctx.KVStore(k.StoreKey).Set([]byte("domain:Rotation"), k.cdc.MustMarshalLengthPrefixed(&domain))
	if err := k.RegisterValidator(ctx, revokedDerivedOperator.String(), testPubKey("post-revocation-safe-key"), stake, "Rotation"); err == nil {
		t.Fatal("operator derived from a revoked consensus key was registered")
	}
}

func TestValidatorRotationGenesisRoundTrip(t *testing.T) {
	am1, k1, ctx1 := setupModuleForGenesis(t)
	ctx1 = ctx1.WithBlockHeight(20)
	operator, before := setupRotationValidator(t, k1, ctx1)
	oldKey, err := k1.RotateValidatorKey(ctx1, operator, operator.String(), before.PubKey, testPubKey("roundtrip-new"))
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
		Sender:            operator,
		OperatorAddr:      operator.String(),
		ExpectedOldPubKey: hex.EncodeToString(testPubKey("message-old-key")),
		NewPubKey:         hex.EncodeToString(testPubKey("message-key")),
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
	msg.ExpectedOldPubKey = "BAD"
	if err := msg.ValidateBasic(); err == nil || !strings.Contains(err.Error(), "expected_old_pub_key") {
		t.Fatalf("both-invalid key error was not deterministic: %v", err)
	}
}

func TestMsgServerRotateValidatorKeyEmitsDeterministicEvent(t *testing.T) {
	k, baseCtx := setupKeeper(t)
	ctx := baseCtx.WithBlockHeight(7).WithEventManager(sdk.NewEventManager())
	operator, before := setupRotationValidator(t, k, ctx)
	newKey := testPubKey("event-new")
	msg := &MsgRotateValidatorKey{
		Sender:            operator,
		OperatorAddr:      operator.String(),
		ExpectedOldPubKey: hex.EncodeToString(before.PubKey),
		NewPubKey:         hex.EncodeToString(newKey),
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
		{"scheduled_activation_height", "9"},
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

func TestValidateGenesisRejectsConsensusDerivedOperator(t *testing.T) {
	pubKey := testPubKey("coupled-genesis-key")
	operator := sdk.AccAddress((&ed25519.PubKey{Key: pubKey}).Address())
	genesis := GenesisState{
		Domains: []Domain{{
			Name: "Coupled", Admin: operator, Members: []string{operator.String()},
			Treasury: sdk.NewCoins(), Issues: []Issue{}, PermissionReg: []string{},
		}},
		Validators: []GenesisValidator{{
			OperatorAddr: operator.String(), PubKey: pubKey, Stake: rewards.StakeMin, Domain: "Coupled",
		}},
	}
	if err := ValidateGenesisState(genesis); err == nil || !strings.Contains(err.Error(), "collides") {
		t.Fatalf("consensus-derived operator genesis was not rejected: %v", err)
	}
}

func TestValidateGenesisRejectsCrossCoupledOperators(t *testing.T) {
	firstKey := testPubKey("cross-genesis-first")
	secondKey := testPubKey("cross-genesis-second")
	firstOperator := sdk.AccAddress((&ed25519.PubKey{Key: secondKey}).Address())
	secondOperator := sdk.AccAddress((&ed25519.PubKey{Key: firstKey}).Address())
	genesis := GenesisState{
		Domains: []Domain{{
			Name: "Cross", Admin: firstOperator,
			Members:  []string{firstOperator.String(), secondOperator.String()},
			Treasury: sdk.NewCoins(), Issues: []Issue{}, PermissionReg: []string{},
		}},
		Validators: []GenesisValidator{
			{OperatorAddr: firstOperator.String(), PubKey: firstKey, Stake: rewards.StakeMin, Domain: "Cross"},
			{OperatorAddr: secondOperator.String(), PubKey: secondKey, Stake: rewards.StakeMin, Domain: "Cross"},
		},
	}
	if err := ValidateGenesisState(genesis); err == nil || !strings.Contains(err.Error(), "collides") {
		t.Fatalf("cross-coupled validator genesis was not rejected: %v", err)
	}
}

func TestValidateGenesisRejectsOperatorDerivedFromRevokedKey(t *testing.T) {
	revokedKey := testPubKey("revoked-genesis-authority")
	operator := sdk.AccAddress((&ed25519.PubKey{Key: revokedKey}).Address())
	genesis := GenesisState{
		Domains: []Domain{{
			Name: "RevokedAuthority", Admin: operator, Members: []string{operator.String()},
			Treasury: sdk.NewCoins(), Issues: []Issue{}, PermissionReg: []string{},
		}},
		Validators: []GenesisValidator{{
			OperatorAddr: operator.String(), PubKey: testPubKey("independent-active-key"),
			Stake: rewards.StakeMin, Domain: "RevokedAuthority",
		}},
		RevokedValidatorKeys: []RevokedValidatorKey{{
			PubKey: revokedKey, OperatorAddr: operator.String(), RevokedAtHeight: 1,
		}},
	}
	if err := ValidateGenesisState(genesis); err == nil || !strings.Contains(err.Error(), "revoked") {
		t.Fatalf("operator derived from revoked key was not rejected: %v", err)
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
