package migration_test

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdkbech32 "github.com/cosmos/cosmos-sdk/types/bech32"

	"truerepublic/migration"
)

const testHRP = "truerepublic"

func newPrivKey(keyType string) cryptotypes.PrivKey {
	switch keyType {
	case migration.PubKeyTypeEd25519:
		return ed25519.GenPrivKey()
	case migration.PubKeyTypeSecp256k1:
		return secp256k1.GenPrivKey()
	default:
		panic("unsupported key type in test: " + keyType)
	}
}

func randomBytes(t *testing.T, n int) []byte {
	t.Helper()
	bz := make([]byte, n)
	if _, err := rand.Read(bz); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}
	return bz
}

func randomAddress(t *testing.T) string {
	t.Helper()
	addr, err := sdkbech32.ConvertAndEncode(testHRP, randomBytes(t, 20))
	if err != nil {
		t.Fatalf("ConvertAndEncode: %v", err)
	}
	return addr
}

// sortEntries sorts mappings and their parallel private keys ascending by
// decoded old operator address, mirroring the canonical ordering.
func sortEntries(t *testing.T, mappings []migration.OperatorMapping, privs []cryptotypes.PrivKey) {
	t.Helper()
	type pair struct {
		m migration.OperatorMapping
		p cryptotypes.PrivKey
	}
	pairs := make([]pair, len(mappings))
	for i := range mappings {
		pairs[i] = pair{mappings[i], privs[i]}
	}
	for i := 1; i < len(pairs); i++ {
		for j := i; j > 0; j-- {
			_, addrA, err := sdkbech32.DecodeAndConvert(pairs[j-1].m.OldOperator)
			if err != nil {
				t.Fatalf("DecodeAndConvert: %v", err)
			}
			_, addrB, err := sdkbech32.DecodeAndConvert(pairs[j].m.OldOperator)
			if err != nil {
				t.Fatalf("DecodeAndConvert: %v", err)
			}
			if bytes.Compare(addrA, addrB) <= 0 {
				break
			}
			pairs[j-1], pairs[j] = pairs[j], pairs[j-1]
		}
	}
	for i := range pairs {
		mappings[i] = pairs[i].m
		privs[i] = pairs[i].p
	}
}

func signDescriptor(t *testing.T, desc *migration.Descriptor, privs []cryptotypes.PrivKey) {
	t.Helper()
	msg, err := migration.SigningBytes(desc)
	if err != nil {
		t.Fatalf("SigningBytes: %v", err)
	}
	for i, priv := range privs {
		sig, err := priv.Sign(msg)
		if err != nil {
			t.Fatalf("Sign: %v", err)
		}
		desc.Mappings[i].Signature = sig
	}
}

// buildDescriptor returns a valid, fully signed descriptor with one mapping
// per requested key type, plus the parallel fresh private keys.
func buildDescriptor(t *testing.T, keyTypes []string) (*migration.Descriptor, []cryptotypes.PrivKey) {
	t.Helper()
	mappings := make([]migration.OperatorMapping, len(keyTypes))
	privs := make([]cryptotypes.PrivKey, len(keyTypes))
	for i, keyType := range keyTypes {
		priv := newPrivKey(keyType)
		newAddr, err := sdkbech32.ConvertAndEncode(testHRP, priv.PubKey().Address())
		if err != nil {
			t.Fatalf("ConvertAndEncode: %v", err)
		}
		mappings[i] = migration.OperatorMapping{
			OldOperator: randomAddress(t),
			NewOperator: newAddr,
			PubKeyType:  keyType,
			PubKey:      priv.PubKey().Bytes(),
		}
		privs[i] = priv
	}
	sortEntries(t, mappings, privs)
	desc := &migration.Descriptor{
		Version:             migration.DescriptorVersion,
		SourceChainID:       "truerepublic-legacy-1",
		TargetChainID:       "truerepublic-2",
		HaltHeight:          987654,
		SourceAppHash:       randomBytes(t, 32),
		SourceGenesisSHA256: randomBytes(t, sha256.Size),
		TransformID:         "gh-61-legacy-authority-v1",
		Mappings:            mappings,
	}
	signDescriptor(t, desc, privs)
	return desc, privs
}

// unrelatedConsensusKeys returns validly shaped consensus public keys that
// are not connected to any descriptor entry.
func unrelatedConsensusKeys(t *testing.T, n int) [][]byte {
	t.Helper()
	keys := make([][]byte, n)
	for i := range keys {
		keys[i] = randomBytes(t, 32)
	}
	return keys
}

func TestVerifyValidDescriptors(t *testing.T) {
	cases := []struct {
		name      string
		keyTypes  []string
		consensus int
	}{
		{"single ed25519 mapping, no consensus keys supplied", []string{migration.PubKeyTypeEd25519}, 0},
		{"single secp256k1 mapping with unrelated consensus keys", []string{migration.PubKeyTypeSecp256k1}, 3},
		{"multiple mixed key types with unrelated consensus keys", []string{
			migration.PubKeyTypeEd25519,
			migration.PubKeyTypeSecp256k1,
			migration.PubKeyTypeEd25519,
			migration.PubKeyTypeSecp256k1,
		}, 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			desc, _ := buildDescriptor(t, tc.keyTypes)
			if err := migration.Verify(desc, unrelatedConsensusKeys(t, tc.consensus)); err != nil {
				t.Fatalf("Verify rejected a valid descriptor: %v", err)
			}
		})
	}
}

func TestVerifyRejects(t *testing.T) {
	defaultKeyTypes := []string{
		migration.PubKeyTypeEd25519,
		migration.PubKeyTypeSecp256k1,
		migration.PubKeyTypeEd25519,
	}
	cases := []struct {
		name     string
		keyTypes []string
		mutate   func(t *testing.T, desc *migration.Descriptor, privs []cryptotypes.PrivKey, consensus *[][]byte)
	}{
		{
			name: "version zero",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.Version = 0
			},
		},
		{
			name: "unsupported version",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.Version = 2
			},
		},
		{
			name: "empty source chain ID",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.SourceChainID = ""
			},
		},
		{
			name: "empty target chain ID",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.TargetChainID = ""
			},
		},
		{
			name: "target chain ID reuses source chain ID",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.TargetChainID = desc.SourceChainID
			},
		},
		{
			name: "zero halt height",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.HaltHeight = 0
			},
		},
		{
			name: "negative halt height",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.HaltHeight = -1
			},
		},
		{
			name: "nil source app hash",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.SourceAppHash = nil
			},
		},
		{
			name: "short source app hash",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.SourceAppHash = desc.SourceAppHash[:31]
			},
		},
		{
			name: "long source app hash",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.SourceAppHash = append(desc.SourceAppHash, 0x00)
			},
		},
		{
			name: "nil source genesis SHA-256",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.SourceGenesisSHA256 = nil
			},
		},
		{
			name: "short source genesis SHA-256",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.SourceGenesisSHA256 = desc.SourceGenesisSHA256[:sha256.Size-1]
			},
		},
		{
			name: "long source genesis SHA-256",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.SourceGenesisSHA256 = append(desc.SourceGenesisSHA256, 0x00)
			},
		},
		{
			name: "empty transform ID",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.TransformID = ""
			},
		},
		{
			name: "no mappings",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.Mappings = nil
			},
		},
		{
			name: "malformed old operator bech32",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.Mappings[0].OldOperator = "not-a-bech32-address"
			},
		},
		{
			name: "malformed new operator bech32",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.Mappings[0].NewOperator = "not-a-bech32-address"
			},
		},
		{
			name: "inconsistent bech32 prefix",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				other, err := sdkbech32.ConvertAndEncode("other", randomBytes(t, 20))
				if err != nil {
					t.Fatalf("ConvertAndEncode: %v", err)
				}
				desc.Mappings[0].OldOperator = other
			},
		},
		{
			name: "new operator address not derived from supplied key",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.Mappings[0].NewOperator = randomAddress(t)
			},
		},
		{
			name: "wrong public key type",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.Mappings[0].PubKeyType = "sr25519"
			},
		},
		{
			name: "empty public key type",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.Mappings[0].PubKeyType = ""
			},
		},
		{
			name:     "malformed ed25519 public key",
			keyTypes: []string{migration.PubKeyTypeEd25519},
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.Mappings[0].PubKey = desc.Mappings[0].PubKey[:31]
			},
		},
		{
			name:     "malformed secp256k1 public key",
			keyTypes: []string{migration.PubKeyTypeSecp256k1},
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.Mappings[0].PubKey = desc.Mappings[0].PubKey[:32]
			},
		},
		{
			name: "garbage signature",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.Mappings[0].Signature = randomBytes(t, 64)
			},
		},
		{
			name: "missing signature",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.Mappings[0].Signature = nil
			},
		},
		{
			name: "signatures swapped between entries",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.Mappings[0].Signature, desc.Mappings[1].Signature =
					desc.Mappings[1].Signature, desc.Mappings[0].Signature
			},
		},
		{
			name:     "proof made by a legacy consensus key instead of the fresh key",
			keyTypes: []string{migration.PubKeyTypeEd25519},
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, consensus *[][]byte) {
				legacyConsensusKey := ed25519.GenPrivKey()
				*consensus = [][]byte{legacyConsensusKey.PubKey().Bytes()}
				msg, err := migration.SigningBytes(desc)
				if err != nil {
					t.Fatalf("SigningBytes: %v", err)
				}
				sig, err := legacyConsensusKey.Sign(msg)
				if err != nil {
					t.Fatalf("Sign: %v", err)
				}
				desc.Mappings[0].Signature = sig
			},
		},
		{
			name: "mutation after signing: transform ID",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.TransformID = "gh-61-legacy-authority-v2"
			},
		},
		{
			name: "mutation after signing: source chain ID",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.SourceChainID = "other-legacy-1"
			},
		},
		{
			name: "mutation after signing: target chain ID",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.TargetChainID = "truerepublic-3"
			},
		},
		{
			name: "mutation after signing: halt height",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.HaltHeight++
			},
		},
		{
			name: "mutation after signing: source app hash",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.SourceAppHash[0] ^= 0xff
			},
		},
		{
			name: "mutation after signing: source genesis SHA-256",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				desc.SourceGenesisSHA256[0] ^= 0xff
			},
		},
		{
			name: "mutation after signing: mapping added",
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, _ *[][]byte) {
				extra, _ := buildDescriptor(t, []string{migration.PubKeyTypeEd25519})
				desc.Mappings = append(desc.Mappings, extra.Mappings[0])
				if err := migration.SortMappings(desc.Mappings); err != nil {
					t.Fatalf("SortMappings: %v", err)
				}
			},
		},
		{
			name: "unsorted mappings",
			mutate: func(t *testing.T, desc *migration.Descriptor, privs []cryptotypes.PrivKey, _ *[][]byte) {
				desc.Mappings[0], desc.Mappings[1] = desc.Mappings[1], desc.Mappings[0]
				privs[0], privs[1] = privs[1], privs[0]
				signDescriptor(t, desc, privs)
			},
		},
		{
			name: "duplicate old operator address",
			mutate: func(t *testing.T, desc *migration.Descriptor, privs []cryptotypes.PrivKey, _ *[][]byte) {
				desc.Mappings[1].OldOperator = desc.Mappings[0].OldOperator
				signDescriptor(t, desc, privs)
			},
		},
		{
			name: "duplicate new operator address",
			mutate: func(t *testing.T, desc *migration.Descriptor, privs []cryptotypes.PrivKey, _ *[][]byte) {
				desc.Mappings[1].NewOperator = desc.Mappings[0].NewOperator
				desc.Mappings[1].PubKeyType = desc.Mappings[0].PubKeyType
				desc.Mappings[1].PubKey = desc.Mappings[0].PubKey
				privs[1] = privs[0]
				signDescriptor(t, desc, privs)
			},
		},
		{
			name:     "old and new operator identical within entry",
			keyTypes: []string{migration.PubKeyTypeEd25519},
			mutate: func(t *testing.T, desc *migration.Descriptor, privs []cryptotypes.PrivKey, _ *[][]byte) {
				desc.Mappings[0].OldOperator = desc.Mappings[0].NewOperator
				signDescriptor(t, desc, privs)
			},
		},
		{
			name: "cross-coupled old and new operator addresses",
			keyTypes: []string{
				migration.PubKeyTypeEd25519,
				migration.PubKeyTypeSecp256k1,
			},
			mutate: func(t *testing.T, desc *migration.Descriptor, privs []cryptotypes.PrivKey, _ *[][]byte) {
				desc.Mappings[1].OldOperator = desc.Mappings[0].NewOperator
				sortEntries(t, desc.Mappings, privs)
				signDescriptor(t, desc, privs)
			},
		},
		{
			name:     "new operator address derived from an active consensus key",
			keyTypes: []string{migration.PubKeyTypeEd25519},
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, consensus *[][]byte) {
				*consensus = [][]byte{desc.Mappings[0].PubKey}
			},
		},
		{
			name:     "new operator address derived from a historical pending or revoked consensus key",
			keyTypes: []string{migration.PubKeyTypeEd25519},
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, consensus *[][]byte) {
				*consensus = append(unrelatedConsensusKeys(t, 3), desc.Mappings[0].PubKey)
			},
		},
		{
			name:     "malformed consensus public key supplied by caller",
			keyTypes: []string{migration.PubKeyTypeEd25519},
			mutate: func(t *testing.T, desc *migration.Descriptor, _ []cryptotypes.PrivKey, consensus *[][]byte) {
				*consensus = [][]byte{randomBytes(t, 31)}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			keyTypes := tc.keyTypes
			if keyTypes == nil {
				keyTypes = defaultKeyTypes
			}
			desc, privs := buildDescriptor(t, keyTypes)
			consensus := unrelatedConsensusKeys(t, 2)
			tc.mutate(t, desc, privs, &consensus)
			if err := migration.Verify(desc, consensus); err == nil {
				t.Fatalf("Verify accepted an invalid descriptor (%s)", tc.name)
			}
		})
	}
}

func TestSigningBytesDeterministicAndDomainSeparated(t *testing.T) {
	desc, _ := buildDescriptor(t, []string{
		migration.PubKeyTypeEd25519,
		migration.PubKeyTypeSecp256k1,
	})

	first, err := migration.SigningBytes(desc)
	if err != nil {
		t.Fatalf("SigningBytes: %v", err)
	}
	second, err := migration.SigningBytes(desc)
	if err != nil {
		t.Fatalf("SigningBytes: %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("SigningBytes is not deterministic")
	}

	if !bytes.HasPrefix(first, []byte("truerepublic/GH-61/legacy-authority-descriptor/v1\x00")) {
		t.Fatal("SigningBytes is not domain separated")
	}

	// Signatures are excluded from the canonical encoding.
	desc.Mappings[0].Signature = randomBytes(t, 64)
	withMutatedSig, err := migration.SigningBytes(desc)
	if err != nil {
		t.Fatalf("SigningBytes: %v", err)
	}
	if !bytes.Equal(first, withMutatedSig) {
		t.Fatal("SigningBytes must exclude signatures")
	}

	// Any covered field changes the encoding.
	desc.Mappings[0].PubKey = randomBytes(t, 32)
	withMutatedKey, err := migration.SigningBytes(desc)
	if err != nil {
		t.Fatalf("SigningBytes: %v", err)
	}
	if bytes.Equal(first, withMutatedKey) {
		t.Fatal("SigningBytes must cover every mapping field")
	}
}

func TestSigningBytesRejectsUnencodableInput(t *testing.T) {
	desc, _ := buildDescriptor(t, []string{migration.PubKeyTypeEd25519})

	bad := *desc
	bad.SourceAppHash = bad.SourceAppHash[:16]
	if _, err := migration.SigningBytes(&bad); err == nil {
		t.Fatal("SigningBytes accepted a short app hash")
	}

	bad = *desc
	bad.SourceGenesisSHA256 = bad.SourceGenesisSHA256[:16]
	if _, err := migration.SigningBytes(&bad); err == nil {
		t.Fatal("SigningBytes accepted a short source genesis SHA-256")
	}

	bad = *desc
	bad.HaltHeight = -5
	if _, err := migration.SigningBytes(&bad); err == nil {
		t.Fatal("SigningBytes accepted a negative halt height")
	}

	bad = *desc
	bad.Mappings = append([]migration.OperatorMapping(nil), desc.Mappings...)
	bad.Mappings[0].OldOperator = "not-bech32"
	if _, err := migration.SigningBytes(&bad); err == nil {
		t.Fatal("SigningBytes accepted a malformed operator address")
	}

	bad = *desc
	bad.Mappings = append([]migration.OperatorMapping(nil), desc.Mappings...)
	shortAddress, err := sdkbech32.ConvertAndEncode(testHRP, randomBytes(t, 19))
	if err != nil {
		t.Fatalf("ConvertAndEncode: %v", err)
	}
	bad.Mappings[0].NewOperator = shortAddress
	if _, err := migration.SigningBytes(&bad); err == nil {
		t.Fatal("SigningBytes accepted a non-account-sized operator address")
	}

	if _, err := migration.SigningBytes(nil); err == nil {
		t.Fatal("SigningBytes accepted a nil descriptor")
	}
}

func TestDescriptorJSONUsesStableFieldNames(t *testing.T) {
	desc, _ := buildDescriptor(t, []string{migration.PubKeyTypeEd25519})
	bz, err := json.Marshal(desc)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	for _, field := range []string{
		`"source_chain_id"`,
		`"target_chain_id"`,
		`"halt_height"`,
		`"source_app_hash"`,
		`"source_genesis_sha256"`,
		`"transform_id"`,
		`"old_operator"`,
		`"new_operator"`,
		`"pub_key_type"`,
	} {
		if !bytes.Contains(bz, []byte(field)) {
			t.Fatalf("descriptor JSON is missing stable field %s: %s", field, bz)
		}
	}
}

func TestVerifySourceGenesis(t *testing.T) {
	raw := []byte("{\"chain_id\":\"legacy\",\"app_state\":{}}")
	digest := sha256.Sum256(raw)
	desc := &migration.Descriptor{
		SourceGenesisSHA256: append([]byte(nil), digest[:]...),
	}
	if err := migration.VerifySourceGenesis(desc, raw); err != nil {
		t.Fatalf("VerifySourceGenesis rejected exact bytes: %v", err)
	}

	mutated := append([]byte(nil), raw...)
	mutated[len(mutated)-1] = ' '
	if err := migration.VerifySourceGenesis(desc, mutated); err == nil {
		t.Fatal("VerifySourceGenesis accepted post-signing byte mutation")
	}

	desc.SourceGenesisSHA256 = desc.SourceGenesisSHA256[:sha256.Size-1]
	if err := migration.VerifySourceGenesis(desc, raw); err == nil {
		t.Fatal("VerifySourceGenesis accepted a malformed commitment")
	}
	if err := migration.VerifySourceGenesis(nil, raw); err == nil {
		t.Fatal("VerifySourceGenesis accepted a nil descriptor")
	}
}

func TestDeriveAddress(t *testing.T) {
	edPriv := ed25519.GenPrivKey()
	addr, err := migration.DeriveAddress(migration.PubKeyTypeEd25519, edPriv.PubKey().Bytes())
	if err != nil {
		t.Fatalf("DeriveAddress ed25519: %v", err)
	}
	if !bytes.Equal(addr, edPriv.PubKey().Address()) {
		t.Fatal("DeriveAddress ed25519 mismatch with SDK derivation")
	}

	secpPriv := secp256k1.GenPrivKey()
	addr, err = migration.DeriveAddress(migration.PubKeyTypeSecp256k1, secpPriv.PubKey().Bytes())
	if err != nil {
		t.Fatalf("DeriveAddress secp256k1: %v", err)
	}
	if !bytes.Equal(addr, secpPriv.PubKey().Address()) {
		t.Fatal("DeriveAddress secp256k1 mismatch with SDK derivation")
	}

	if _, err := migration.DeriveAddress("sr25519", randomBytes(t, 32)); err == nil {
		t.Fatal("DeriveAddress accepted an unsupported key type")
	}
	if _, err := migration.DeriveAddress(migration.PubKeyTypeEd25519, randomBytes(t, 31)); err == nil {
		t.Fatal("DeriveAddress accepted a malformed key")
	}
}

func TestSortMappings(t *testing.T) {
	desc, _ := buildDescriptor(t, []string{
		migration.PubKeyTypeEd25519,
		migration.PubKeyTypeSecp256k1,
		migration.PubKeyTypeEd25519,
	})
	shuffled := []migration.OperatorMapping{desc.Mappings[2], desc.Mappings[0], desc.Mappings[1]}
	if err := migration.SortMappings(shuffled); err != nil {
		t.Fatalf("SortMappings: %v", err)
	}
	for i := range shuffled {
		if shuffled[i].OldOperator != desc.Mappings[i].OldOperator {
			t.Fatal("SortMappings did not restore canonical order")
		}
	}

	shuffled[0].OldOperator = "not-bech32"
	if err := migration.SortMappings(shuffled); err == nil {
		t.Fatal("SortMappings accepted a malformed operator address")
	}
}

// TestVerifyErrorMessagesExercise ensures failures are attributable: every
// rejection carries the mapping index or the violated field.
func TestVerifyErrorMessagesExercise(t *testing.T) {
	desc, _ := buildDescriptor(t, []string{migration.PubKeyTypeEd25519})
	desc.Mappings[0].Signature = randomBytes(t, 64)
	err := migration.Verify(desc, nil)
	if err == nil || !strings.Contains(err.Error(), "mapping 0") {
		t.Fatalf("expected attributable error, got: %v", err)
	}
}
