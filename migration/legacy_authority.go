// Package migration implements the GH-61 canonical legacy-authority migration
// descriptor and fresh replacement-operator proof verification.
//
// A Descriptor is the single canonical artifact of the offline
// halt/export/transform/re-import ceremony that moves the chain away from
// consensus-key-coupled operator authority. It binds the source chain ID,
// distinct target chain ID, positive halt height, 32-byte source app hash,
// SHA-256 of the exact raw source genesis export, a nonempty transform ID, and
// the complete, sorted, one-to-one old-to-new operator mapping. Every mapping
// entry carries a proof-of-possession signature made by the fresh,
// independently controlled replacement account key over the canonical signing
// bytes of the whole descriptor.
//
// Authorization boundary: these proofs authenticate ONLY possession of each
// fresh replacement account key at ceremony time. They are not, and must
// never be presented as, retroactive governance authorization by the legacy
// validator set. A legacy consensus-key signature is never accepted as
// authorization here: signatures are verified exclusively against the fresh
// account public key declared in each entry, and any new operator address
// derivable from a caller-supplied consensus public key (active, historical,
// pending, or revoked) is rejected. Callers MUST supply the complete set of
// consensus public keys from the exported state; keys withheld by the caller
// cannot be excluded by this package.
package migration

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"fmt"
	"sort"

	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdkbech32 "github.com/cosmos/cosmos-sdk/types/bech32"
)

// DescriptorVersion is the only supported descriptor version.
const DescriptorVersion uint32 = 1

// Supported fresh account public-key types.
const (
	// PubKeyTypeEd25519 is a Cosmos SDK ed25519 account public key (32 bytes).
	PubKeyTypeEd25519 = "ed25519"
	// PubKeyTypeSecp256k1 is a Cosmos SDK secp256k1 account public key
	// (33 bytes, compressed).
	PubKeyTypeSecp256k1 = "secp256k1"
)

const (
	// addressLength is the length of a decoded bech32 account address.
	addressLength = 20
	// appHashLength is the required length of the source app hash.
	appHashLength = 32
	// genesisHashLength is the length of a SHA-256 commitment to the exact
	// raw source genesis export.
	genesisHashLength = sha256.Size
	// consensusKeyLength is the length of a raw CometBFT ed25519 consensus
	// public key as supplied by the caller.
	consensusKeyLength = 32
	// domainSeparationTag prefixes all canonical signing bytes so a
	// descriptor proof can never be replayed as a signature for any other
	// message family.
	domainSeparationTag = "truerepublic/GH-61/legacy-authority-descriptor/v1"
)

// OperatorMapping is one old-to-new operator replacement entry. Signature is
// made by the fresh account private key belonging to PubKey over the
// canonical SigningBytes of the entire descriptor (all entries included).
type OperatorMapping struct {
	// OldOperator is the legacy operator account address (bech32).
	OldOperator string `json:"old_operator"`
	// NewOperator is the fresh replacement operator account address
	// (bech32). It must equal the address derived from PubKey.
	NewOperator string `json:"new_operator"`
	// PubKeyType is PubKeyTypeEd25519 or PubKeyTypeSecp256k1.
	PubKeyType string `json:"pub_key_type"`
	// PubKey is the raw fresh account public key.
	PubKey []byte `json:"pub_key"`
	// Signature proves possession of the fresh account private key. It is
	// excluded from the canonical signing bytes.
	Signature []byte `json:"signature"`
}

// Descriptor is the canonical GH-61 legacy-authority migration descriptor.
type Descriptor struct {
	Version       uint32 `json:"version"`
	SourceChainID string `json:"source_chain_id"`
	TargetChainID string `json:"target_chain_id"`
	// HaltHeight is the positive source-chain height at which the export was
	// halted.
	HaltHeight int64 `json:"halt_height"`
	// SourceAppHash is the 32-byte app hash of the source chain at
	// HaltHeight.
	SourceAppHash []byte `json:"source_app_hash"`
	// SourceGenesisSHA256 is the SHA-256 digest of the exact raw source
	// genesis export bytes that every replacement operator reviewed and
	// signed. Reformatting or otherwise changing the export changes this
	// commitment.
	SourceGenesisSHA256 []byte `json:"source_genesis_sha256"`
	// TransformID identifies the deterministic transform applied to the
	// export.
	TransformID string `json:"transform_id"`
	// Mappings is the complete old-to-new operator mapping, sorted strictly
	// ascending by decoded old operator address bytes.
	Mappings []OperatorMapping `json:"mappings"`
}

// SigningBytes returns the canonical, domain-separated signing bytes of the
// descriptor. This is the one deterministic encoding every proof signs; it
// covers the entire descriptor and every mapping entry, and excludes all
// signatures.
//
// Encoding (all integers little-endian, lp(x) = uint32(len(x)) || x):
//
//	domainSeparationTag || 0x00 ||
//	uint32(version) ||
//	lp(sourceChainID) || lp(targetChainID) ||
//	uint64(haltHeight) ||
//	sourceAppHash (32 raw bytes) ||
//	sourceGenesisSHA256 (32 raw bytes) ||
//	lp(transformID) ||
//	uint32(mappingCount) ||
//	for each mapping in descriptor order:
//	    oldOperator (20 raw decoded address bytes) ||
//	    newOperator (20 raw decoded address bytes) ||
//	    lp(pubKeyType) || lp(pubKey)
//
// Addresses are encoded as their decoded 20-byte payloads, never as bech32
// strings, so encoding is independent of bech32 casing. SigningBytes does not
// fully validate the descriptor; it fails only when a field cannot be
// encoded. Verify performs the complete fail-closed validation.
func SigningBytes(desc *Descriptor) ([]byte, error) {
	if desc == nil {
		return nil, fmt.Errorf("migration: nil descriptor")
	}
	if desc.HaltHeight < 0 {
		return nil, fmt.Errorf("migration: halt height %d cannot be encoded", desc.HaltHeight)
	}
	if len(desc.SourceAppHash) != appHashLength {
		return nil, fmt.Errorf("migration: source app hash must be %d bytes, got %d", appHashLength, len(desc.SourceAppHash))
	}
	if len(desc.SourceGenesisSHA256) != genesisHashLength {
		return nil, fmt.Errorf(
			"migration: source genesis SHA-256 must be %d bytes, got %d",
			genesisHashLength,
			len(desc.SourceGenesisSHA256),
		)
	}

	buf := bytes.NewBuffer(make([]byte, 0, 256+len(desc.Mappings)*128))
	buf.WriteString(domainSeparationTag)
	buf.WriteByte(0x00)

	writeUint32(buf, desc.Version)
	writeLengthPrefixed(buf, []byte(desc.SourceChainID))
	writeLengthPrefixed(buf, []byte(desc.TargetChainID))
	writeUint64(buf, uint64(desc.HaltHeight))
	buf.Write(desc.SourceAppHash)
	buf.Write(desc.SourceGenesisSHA256)
	writeLengthPrefixed(buf, []byte(desc.TransformID))

	writeUint32(buf, uint32(len(desc.Mappings)))
	for i, m := range desc.Mappings {
		_, oldAddr, err := sdkbech32.DecodeAndConvert(m.OldOperator)
		if err != nil {
			return nil, fmt.Errorf("migration: mapping %d old operator: %w", i, err)
		}
		if len(oldAddr) != addressLength {
			return nil, fmt.Errorf("migration: mapping %d old operator address must be %d bytes, got %d", i, addressLength, len(oldAddr))
		}
		_, newAddr, err := sdkbech32.DecodeAndConvert(m.NewOperator)
		if err != nil {
			return nil, fmt.Errorf("migration: mapping %d new operator: %w", i, err)
		}
		if len(newAddr) != addressLength {
			return nil, fmt.Errorf("migration: mapping %d new operator address must be %d bytes, got %d", i, addressLength, len(newAddr))
		}
		buf.Write(oldAddr)
		buf.Write(newAddr)
		writeLengthPrefixed(buf, []byte(m.PubKeyType))
		writeLengthPrefixed(buf, m.PubKey)
	}
	return buf.Bytes(), nil
}

// Verify performs the complete fail-closed validation of a descriptor and
// its fresh replacement-operator proofs. It returns nil only when every
// check passes:
//
//   - version is DescriptorVersion;
//   - source and target chain IDs are nonempty;
//   - source and target chain IDs differ, preventing legacy transaction replay;
//   - halt height is positive;
//   - source app hash is exactly 32 bytes;
//   - source genesis SHA-256 is exactly 32 bytes;
//   - transform ID is nonempty;
//   - at least one mapping exists;
//   - every old/new operator address is valid bech32 with a 20-byte payload
//     and all addresses share one human-readable prefix;
//   - each new operator address equals the address derived from the entry's
//     fresh account public key;
//   - each entry's signature verifies against the entry's fresh account
//     public key over SigningBytes (proof of possession);
//   - entries are sorted strictly ascending by old operator address;
//   - old operator addresses are unique, new operator addresses are unique,
//     and no address appears as both an old and a new operator anywhere in
//     the descriptor (no self- or cross-coupling);
//   - no new operator address is derivable from any caller-supplied
//     consensus public key (active, historical, pending, or revoked).
//
// consensusPubKeys must contain the complete set of raw CometBFT ed25519
// consensus public keys (32 bytes each) from the exported state; a malformed
// entry fails closed. Signatures are verified only against the fresh account
// keys declared in the descriptor — a legacy consensus-key signature is never
// accepted as authorization. The proofs authenticate fresh-key possession
// only, not retroactive governance.
func Verify(desc *Descriptor, consensusPubKeys [][]byte) error {
	if desc == nil {
		return fmt.Errorf("migration: nil descriptor")
	}
	if desc.Version != DescriptorVersion {
		return fmt.Errorf("migration: unsupported descriptor version %d", desc.Version)
	}
	if desc.SourceChainID == "" {
		return fmt.Errorf("migration: empty source chain ID")
	}
	if desc.TargetChainID == "" {
		return fmt.Errorf("migration: empty target chain ID")
	}
	if desc.SourceChainID == desc.TargetChainID {
		return fmt.Errorf("migration: target chain ID must differ from source chain ID")
	}
	if desc.HaltHeight <= 0 {
		return fmt.Errorf("migration: halt height must be positive, got %d", desc.HaltHeight)
	}
	if len(desc.SourceAppHash) != appHashLength {
		return fmt.Errorf("migration: source app hash must be %d bytes, got %d", appHashLength, len(desc.SourceAppHash))
	}
	if len(desc.SourceGenesisSHA256) != genesisHashLength {
		return fmt.Errorf(
			"migration: source genesis SHA-256 must be %d bytes, got %d",
			genesisHashLength,
			len(desc.SourceGenesisSHA256),
		)
	}
	if desc.TransformID == "" {
		return fmt.Errorf("migration: empty transform ID")
	}
	if len(desc.Mappings) == 0 {
		return fmt.Errorf("migration: descriptor must contain at least one operator mapping")
	}

	forbidden, err := consensusDerivedAddresses(consensusPubKeys)
	if err != nil {
		return err
	}

	signingBytes, err := SigningBytes(desc)
	if err != nil {
		return err
	}

	var hrp string
	var prevOld []byte
	oldSeen := make(map[string]struct{}, len(desc.Mappings))
	newSeen := make(map[string]struct{}, len(desc.Mappings))
	oldAddrs := make(map[string]int, len(desc.Mappings))

	for i, m := range desc.Mappings {
		oldHRP, oldAddr, err := sdkbech32.DecodeAndConvert(m.OldOperator)
		if err != nil {
			return fmt.Errorf("migration: mapping %d old operator: %w", i, err)
		}
		newHRP, newAddr, err := sdkbech32.DecodeAndConvert(m.NewOperator)
		if err != nil {
			return fmt.Errorf("migration: mapping %d new operator: %w", i, err)
		}
		if len(oldAddr) != addressLength {
			return fmt.Errorf("migration: mapping %d old operator address must be %d bytes, got %d", i, addressLength, len(oldAddr))
		}
		if len(newAddr) != addressLength {
			return fmt.Errorf("migration: mapping %d new operator address must be %d bytes, got %d", i, addressLength, len(newAddr))
		}
		if i == 0 {
			hrp = oldHRP
		}
		if oldHRP != hrp || newHRP != hrp {
			return fmt.Errorf("migration: mapping %d uses inconsistent bech32 prefix", i)
		}
		if bytes.Equal(oldAddr, newAddr) {
			return fmt.Errorf("migration: mapping %d reuses the old operator address as the new operator address", i)
		}
		if prevOld != nil && bytes.Compare(prevOld, oldAddr) >= 0 {
			return fmt.Errorf("migration: mappings are not sorted strictly ascending by old operator address at index %d", i)
		}
		prevOld = oldAddr
		if _, dup := oldSeen[string(oldAddr)]; dup {
			return fmt.Errorf("migration: duplicate old operator address at mapping %d", i)
		}
		oldSeen[string(oldAddr)] = struct{}{}
		if _, dup := newSeen[string(newAddr)]; dup {
			return fmt.Errorf("migration: duplicate new operator address at mapping %d", i)
		}
		newSeen[string(newAddr)] = struct{}{}
		oldAddrs[string(oldAddr)] = i

		pubKey, err := freshPubKey(m.PubKeyType, m.PubKey)
		if err != nil {
			return fmt.Errorf("migration: mapping %d: %w", i, err)
		}
		if derived := pubKey.Address(); !bytes.Equal(derived, newAddr) {
			return fmt.Errorf("migration: mapping %d new operator address is not derived from the supplied fresh public key", i)
		}
		if _, coupled := forbidden[string(newAddr)]; coupled {
			return fmt.Errorf("migration: mapping %d new operator address is derivable from a legacy consensus public key", i)
		}
		// The signature is verified exclusively against the fresh account
		// public key declared in this entry. No legacy consensus key is ever
		// consulted as a signer.
		if !pubKey.VerifySignature(signingBytes, m.Signature) {
			return fmt.Errorf("migration: mapping %d has an invalid fresh-key proof-of-possession signature", i)
		}
	}

	for newAddr := range newSeen {
		if j, coupled := oldAddrs[newAddr]; coupled {
			return fmt.Errorf("migration: new operator address of one mapping is the old operator address of mapping %d (cross-coupling)", j)
		}
	}
	return nil
}

// VerifySourceGenesis checks that raw contains the exact source genesis export
// bytes committed by the signed descriptor. It deliberately hashes the raw
// bytes before any JSON decoding or normalization so whitespace, key ordering,
// and every state byte remain part of the reviewed artifact.
func VerifySourceGenesis(desc *Descriptor, raw []byte) error {
	if desc == nil {
		return fmt.Errorf("migration: nil descriptor")
	}
	if len(desc.SourceGenesisSHA256) != genesisHashLength {
		return fmt.Errorf(
			"migration: source genesis SHA-256 must be %d bytes, got %d",
			genesisHashLength,
			len(desc.SourceGenesisSHA256),
		)
	}
	digest := sha256.Sum256(raw)
	if subtle.ConstantTimeCompare(desc.SourceGenesisSHA256, digest[:]) != 1 {
		return fmt.Errorf("migration: source genesis SHA-256 does not match the signed descriptor")
	}
	return nil
}

// DeriveAddress derives the account address for a supported fresh public-key
// type and raw key bytes. It exposes the same derivation Verify uses so the
// transformer and signing tooling can construct entries without duplicating
// the logic.
func DeriveAddress(pubKeyType string, pubKey []byte) ([]byte, error) {
	key, err := freshPubKey(pubKeyType, pubKey)
	if err != nil {
		return nil, err
	}
	return append([]byte(nil), key.Address()...), nil
}

// freshPubKey builds the Cosmos SDK account public key for a supported key
// type, failing closed on unknown types and malformed key bytes.
func freshPubKey(pubKeyType string, pubKey []byte) (cryptotypes.PubKey, error) {
	switch pubKeyType {
	case PubKeyTypeEd25519:
		if len(pubKey) != ed25519.PubKeySize {
			return nil, fmt.Errorf("ed25519 public key must be %d bytes, got %d", ed25519.PubKeySize, len(pubKey))
		}
		return &ed25519.PubKey{Key: append([]byte(nil), pubKey...)}, nil
	case PubKeyTypeSecp256k1:
		if len(pubKey) != secp256k1.PubKeySize {
			return nil, fmt.Errorf("secp256k1 public key must be %d bytes (compressed), got %d", secp256k1.PubKeySize, len(pubKey))
		}
		return &secp256k1.PubKey{Key: append([]byte(nil), pubKey...)}, nil
	default:
		return nil, fmt.Errorf("unsupported public key type %q", pubKeyType)
	}
}

// consensusDerivedAddresses derives the account address of every supplied
// consensus public key. Consensus keys are raw CometBFT ed25519 public keys;
// their address is SHA-256 truncated to 20 bytes, matching the derivation
// used for consensus-coupled authority in x/truedemocracy.
func consensusDerivedAddresses(consensusPubKeys [][]byte) (map[string]struct{}, error) {
	out := make(map[string]struct{}, len(consensusPubKeys))
	for i, key := range consensusPubKeys {
		if len(key) != consensusKeyLength {
			return nil, fmt.Errorf("migration: consensus public key %d must be %d bytes (ed25519), got %d", i, consensusKeyLength, len(key))
		}
		out[string(tmhash.SumTruncated(key))] = struct{}{}
	}
	return out, nil
}

// SortMappings sorts the mappings strictly ascending by decoded old operator
// address bytes, as required by the canonical encoding. It is a convenience
// for descriptor construction; Verify rejects unsorted input.
func SortMappings(mappings []OperatorMapping) error {
	type decodedMapping struct {
		mapping OperatorMapping
		oldAddr []byte
	}
	decoded := make([]decodedMapping, len(mappings))
	for i, m := range mappings {
		_, oldAddr, err := sdkbech32.DecodeAndConvert(m.OldOperator)
		if err != nil {
			return fmt.Errorf("migration: mapping %d old operator: %w", i, err)
		}
		decoded[i] = decodedMapping{mapping: m, oldAddr: oldAddr}
	}
	sort.SliceStable(decoded, func(a, b int) bool {
		return bytes.Compare(decoded[a].oldAddr, decoded[b].oldAddr) < 0
	})
	for i := range decoded {
		mappings[i] = decoded[i].mapping
	}
	return nil
}

func writeUint32(buf *bytes.Buffer, v uint32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	buf.Write(b[:])
}

func writeUint64(buf *bytes.Buffer, v uint64) {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], v)
	buf.Write(b[:])
}

func writeLengthPrefixed(buf *bytes.Buffer, bz []byte) {
	writeUint32(buf, uint32(len(bz)))
	buf.Write(bz)
}
