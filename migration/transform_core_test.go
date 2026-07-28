package migration_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"testing"

	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkbech32 "github.com/cosmos/cosmos-sdk/types/bech32"

	"truerepublic/migration"
	"truerepublic/token"
	"truerepublic/x/truedemocracy"
)

type coreFixture struct {
	raw       json.RawMessage
	desc      *migration.Descriptor
	old       []string
	fresh     []cryptotypes.PrivKey
	unrelated string
}

func TestTransformTrueDemocracyMinimalLegacyState(t *testing.T) {
	fixture := newCoreFixture(t, 2)
	before := append([]byte(nil), fixture.raw...)
	descBefore, err := json.Marshal(fixture.desc)
	if err != nil {
		t.Fatal(err)
	}

	out, err := migration.TransformTrueDemocracy(fixture.desc, fixture.raw)
	if err != nil {
		t.Fatalf("TransformTrueDemocracy: %v", err)
	}
	if !bytes.Equal(before, fixture.raw) {
		t.Fatal("transform mutated the input export")
	}
	afterDesc, _ := json.Marshal(fixture.desc)
	if !bytes.Equal(descBefore, afterDesc) {
		t.Fatal("transform mutated the descriptor")
	}

	state := decodeTransformedState(t, out)
	if err := truedemocracy.ValidateGenesisState(state); err != nil {
		t.Fatalf("transformed genesis is invalid: %v", err)
	}
	for i, validator := range state.Validators {
		if validator.OperatorAddr != fixture.desc.Mappings[indexOfOld(fixture.desc, fixture.old[i])].NewOperator {
			t.Fatalf("validator %d operator was not rewritten", i)
		}
		if !bytes.Equal(validator.PubKey, oldConsensusKey(t, fixture.raw, i)) {
			t.Fatalf("validator %d consensus key changed", i)
		}
		if validator.Stake != token.StakeMinBaseUnits {
			t.Fatalf("validator %d stake changed", i)
		}
	}
	for _, old := range fixture.old {
		if bytes.Contains(out, []byte(old)) {
			t.Fatalf("old operator %s remains in output", old)
		}
	}
}

func TestTransformTrueDemocracyFailsClosed(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*testing.T, *coreFixture)
	}{
		{
			name: "missing mapping",
			mutate: func(t *testing.T, f *coreFixture) {
				rebuildDescriptor(t, f, f.desc.Mappings[:1], f.fresh[:1])
			},
		},
		{
			name: "extra non-coupled mapping",
			mutate: func(t *testing.T, f *coreFixture) {
				key := secp256k1.GenPrivKey()
				extra := migration.OperatorMapping{
					OldOperator: accountAddress(t, ed25519.GenPrivKey().PubKey().Address()),
					NewOperator: accountAddress(t, key.PubKey().Address()),
					PubKeyType:  migration.PubKeyTypeSecp256k1,
					PubKey:      key.PubKey().Bytes(),
				}
				mappings := append(append([]migration.OperatorMapping(nil), f.desc.Mappings...), extra)
				keys := append(append([]cryptotypes.PrivKey(nil), f.fresh...), key)
				rebuildDescriptor(t, f, mappings, keys)
			},
		},
		{
			name: "replacement collision",
			mutate: func(t *testing.T, f *coreFixture) {
				collisionKey := secp256k1.GenPrivKey()
				collisionAddress := accountAddress(t, collisionKey.PubKey().Address())
				setUnrelatedMember(t, f, collisionAddress)
				target := indexOfOld(f.desc, f.old[0])
				f.desc.Mappings[target].NewOperator = collisionAddress
				f.desc.Mappings[target].PubKey = collisionKey.PubKey().Bytes()
				f.fresh[target] = collisionKey
				signCoreDescriptor(t, f.desc, f.fresh)
			},
		},
		{
			name: "invalid proof",
			mutate: func(_ *testing.T, f *coreFixture) {
				f.desc.Mappings[0].Signature = []byte("invalid")
			},
		},
		{
			name: "malformed export",
			mutate: func(t *testing.T, f *coreFixture) {
				f.raw = json.RawMessage(`{"app_state":`)
				signCoreFixtureExport(t, f)
			},
		},
		{
			name: "history-only coupled operator missing",
			mutate: func(t *testing.T, f *coreFixture) {
				addHistoryOnlyCoupledOperator(t, f)
			},
		},
		{
			name: "foreign account prefix",
			mutate: func(t *testing.T, f *coreFixture) {
				for i := range f.desc.Mappings {
					f.desc.Mappings[i].OldOperator = reencodeAddress(t, f.desc.Mappings[i].OldOperator, "foreign")
					f.desc.Mappings[i].NewOperator = reencodeAddress(t, f.desc.Mappings[i].NewOperator, "foreign")
				}
				sortCoreMappingsAndKeys(t, f.desc.Mappings, f.fresh)
				signCoreDescriptor(t, f.desc, f.fresh)
			},
		},
		{
			name: "malformed consensus key in state",
			mutate: func(t *testing.T, f *coreFixture) {
				state := decodeTransformedState(t, f.raw)
				state.Validators[0].PubKey = state.Validators[0].PubKey[:31]
				replaceState(t, f, state)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newCoreFixture(t, 2)
			tc.mutate(t, fixture)
			rawBefore := append([]byte(nil), fixture.raw...)
			descBefore, _ := json.Marshal(fixture.desc)
			if out, err := migration.TransformTrueDemocracy(fixture.desc, fixture.raw); err == nil || out != nil {
				t.Fatalf("expected atomic failure, output=%s error=%v", out, err)
			}
			if !bytes.Equal(rawBefore, fixture.raw) {
				t.Fatal("failure mutated the input export")
			}
			descAfter, _ := json.Marshal(fixture.desc)
			if !bytes.Equal(descBefore, descAfter) {
				t.Fatal("failure mutated the descriptor")
			}
		})
	}
}

func newCoreFixture(t *testing.T, count int) *coreFixture {
	t.Helper()
	hrp := sdk.GetConfig().GetBech32AccountAddrPrefix()
	consensus := make([]cryptotypes.PrivKey, count)
	fresh := make([]cryptotypes.PrivKey, count)
	old := make([]string, count)
	mappings := make([]migration.OperatorMapping, count)
	members := make([]string, 0, count+1)
	validators := make([]truedemocracy.GenesisValidator, count)
	for i := 0; i < count; i++ {
		consensus[i] = ed25519.GenPrivKey()
		fresh[i] = secp256k1.GenPrivKey()
		old[i] = accountAddress(t, tmhash.SumTruncated(consensus[i].PubKey().Bytes()))
		members = append(members, old[i])
		validators[i] = truedemocracy.GenesisValidator{
			OperatorAddr: old[i],
			PubKey:       consensus[i].PubKey().Bytes(),
			Stake:        token.StakeMinBaseUnits,
			Domain:       "legacy",
		}
		mappings[i] = migration.OperatorMapping{
			OldOperator: old[i],
			NewOperator: accountAddress(t, fresh[i].PubKey().Address()),
			PubKeyType:  migration.PubKeyTypeSecp256k1,
			PubKey:      fresh[i].PubKey().Bytes(),
		}
	}
	unrelated := accountAddress(t, ed25519.GenPrivKey().PubKey().Address())
	members = append(members, unrelated)
	state := truedemocracy.GenesisState{
		Domains: []truedemocracy.Domain{{
			Name:    "legacy",
			Admin:   sdk.MustAccAddressFromBech32(old[0]),
			Members: members,
			Issues: []truedemocracy.Issue{{
				Name: "migration",
				Suggestions: []truedemocracy.Suggestion{{
					Name:    "preserve",
					Creator: old[count-1],
				}},
			}},
		}},
		Validators: validators,
	}
	stateJSON, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	appJSON, err := json.Marshal(map[string]json.RawMessage{
		truedemocracy.ModuleName: stateJSON,
		"untouched":              json.RawMessage(`{"value":"preserved"}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(map[string]json.RawMessage{
		"app_state": appJSON,
		"chain_id":  json.RawMessage(`"legacy-chain"`),
	})
	if err != nil {
		t.Fatal(err)
	}
	desc := &migration.Descriptor{
		Version:       migration.DescriptorVersion,
		SourceChainID: "legacy-chain",
		TargetChainID: "recovered-chain",
		HaltHeight:    100,
		SourceAppHash: bytes.Repeat([]byte{0x42}, 32),
		TransformID:   "gh-61-transform-v1",
		Mappings:      mappings,
	}
	sourceGenesisSHA256 := sha256.Sum256(raw)
	desc.SourceGenesisSHA256 = append([]byte(nil), sourceGenesisSHA256[:]...)
	if hrp == "" {
		t.Fatal("SDK account prefix is empty")
	}
	sortCoreMappingsAndKeys(t, desc.Mappings, fresh)
	signCoreDescriptor(t, desc, fresh)
	return &coreFixture{raw: raw, desc: desc, old: old, fresh: fresh, unrelated: unrelated}
}

func rebuildDescriptor(t *testing.T, fixture *coreFixture, mappings []migration.OperatorMapping, keys []cryptotypes.PrivKey) {
	t.Helper()
	fixture.desc.Mappings = append([]migration.OperatorMapping(nil), mappings...)
	fixture.fresh = append([]cryptotypes.PrivKey(nil), keys...)
	sortCoreMappingsAndKeys(t, fixture.desc.Mappings, fixture.fresh)
	signCoreDescriptor(t, fixture.desc, fixture.fresh)
}

func sortCoreMappingsAndKeys(t *testing.T, mappings []migration.OperatorMapping, keys []cryptotypes.PrivKey) {
	t.Helper()
	for i := 1; i < len(mappings); i++ {
		for j := i; j > 0; j-- {
			_, left, _ := sdkbech32.DecodeAndConvert(mappings[j-1].OldOperator)
			_, right, _ := sdkbech32.DecodeAndConvert(mappings[j].OldOperator)
			if bytes.Compare(left, right) <= 0 {
				break
			}
			mappings[j-1], mappings[j] = mappings[j], mappings[j-1]
			keys[j-1], keys[j] = keys[j], keys[j-1]
		}
	}
}

func signCoreDescriptor(t *testing.T, desc *migration.Descriptor, keys []cryptotypes.PrivKey) {
	t.Helper()
	message, err := migration.SigningBytes(desc)
	if err != nil {
		t.Fatal(err)
	}
	for i, key := range keys {
		signature, err := key.Sign(message)
		if err != nil {
			t.Fatal(err)
		}
		desc.Mappings[i].Signature = signature
	}
}

func signCoreFixtureExport(t *testing.T, fixture *coreFixture) {
	t.Helper()
	digest := sha256.Sum256(fixture.raw)
	fixture.desc.SourceGenesisSHA256 = append(fixture.desc.SourceGenesisSHA256[:0], digest[:]...)
	signCoreDescriptor(t, fixture.desc, fixture.fresh)
}

func accountAddress(t *testing.T, address []byte) string {
	t.Helper()
	out, err := sdkbech32.ConvertAndEncode(sdk.GetConfig().GetBech32AccountAddrPrefix(), address)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// reencodeAddress re-encodes a bech32 address under a different human-readable
// prefix while keeping the decoded payload bytes identical.
func reencodeAddress(t *testing.T, address string, hrp string) string {
	t.Helper()
	_, decoded, err := sdkbech32.DecodeAndConvert(address)
	if err != nil {
		t.Fatal(err)
	}
	out, err := sdkbech32.ConvertAndEncode(hrp, decoded)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func indexOfOld(desc *migration.Descriptor, old string) int {
	for i := range desc.Mappings {
		if desc.Mappings[i].OldOperator == old {
			return i
		}
	}
	return -1
}

func decodeTransformedState(t *testing.T, raw json.RawMessage) truedemocracy.GenesisState {
	t.Helper()
	var root map[string]json.RawMessage
	if err := json.Unmarshal(raw, &root); err != nil {
		t.Fatal(err)
	}
	var app map[string]json.RawMessage
	if err := json.Unmarshal(root["app_state"], &app); err != nil {
		t.Fatal(err)
	}
	var state truedemocracy.GenesisState
	if err := json.Unmarshal(app[truedemocracy.ModuleName], &state); err != nil {
		t.Fatal(err)
	}
	return state
}

func oldConsensusKey(t *testing.T, raw json.RawMessage, index int) []byte {
	t.Helper()
	return decodeTransformedState(t, raw).Validators[index].PubKey
}

func setUnrelatedMember(t *testing.T, fixture *coreFixture, member string) {
	t.Helper()
	state := decodeTransformedState(t, fixture.raw)
	state.Domains[0].Members[len(state.Domains[0].Members)-1] = member
	replaceState(t, fixture, state)
}

func addHistoryOnlyCoupledOperator(t *testing.T, fixture *coreFixture) {
	t.Helper()
	state := decodeTransformedState(t, fixture.raw)
	key := ed25519.GenPrivKey().PubKey().Bytes()
	operator := accountAddress(t, tmhash.SumTruncated(key))
	state.ConsensusKeyHistory = append(state.ConsensusKeyHistory, truedemocracy.ConsensusKeyRecord{
		ConsensusAddress: tmhash.SumTruncated(key),
		PubKey:           key,
		OperatorAddr:     operator,
		ActivatedHeight:  1,
		RetiredHeight:    2,
	})
	replaceState(t, fixture, state)
}

func replaceState(t *testing.T, fixture *coreFixture, state truedemocracy.GenesisState) {
	t.Helper()
	var root map[string]json.RawMessage
	if err := json.Unmarshal(fixture.raw, &root); err != nil {
		t.Fatal(err)
	}
	var app map[string]json.RawMessage
	if err := json.Unmarshal(root["app_state"], &app); err != nil {
		t.Fatal(err)
	}
	stateJSON, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	app[truedemocracy.ModuleName] = stateJSON
	appJSON, err := json.Marshal(app)
	if err != nil {
		t.Fatal(err)
	}
	root["app_state"] = appJSON
	fixture.raw, err = json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	signCoreFixtureExport(t, fixture)
}
