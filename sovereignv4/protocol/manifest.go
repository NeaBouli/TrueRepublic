package protocol

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"fmt"
	"slices"
)

const ManifestSchema = "truerepublic.domain-app/v1"

const (
	CapabilityDiscussionRead     = "discussion.read"
	CapabilityDocumentsLocal     = "documents.local"
	CapabilityRegisteredQuery    = "chain.query.registered"
	CapabilityTransactionPropose = "transaction.propose"
)

var (
	manifestDomain    = []byte("truerepublic.domain-app/v1\n")
	knownCapabilities = map[string]struct{}{
		CapabilityDiscussionRead: {}, CapabilityDocumentsLocal: {},
		CapabilityRegisteredQuery: {}, CapabilityTransactionPropose: {},
	}
)

// Manifest declares the exact deny-by-default capability surface of a domain app.
type Manifest struct {
	Version         uint16
	Schema          string
	AppID           string
	AppVersion      string
	ArtifactHash    [sha256.Size]byte
	DomainID        string
	Capabilities    []string
	MinCoreProtocol uint16
	MaxCoreProtocol uint16
	PublisherKey    [ed25519.PublicKeySize]byte
	Signature       [ed25519.SignatureSize]byte
}

// PublisherFact is the exact publisher registry result already verified by a
// caller. This package cross-binds it but does not verify registry proofs.
type PublisherFact struct {
	DomainID     string
	AppID        string
	AppVersion   string
	PublisherKey [ed25519.PublicKeySize]byte
}

func (m Manifest) validate() error {
	if err := validateVersion(m.Version, ManifestVersion); err != nil {
		return err
	}
	if m.Schema != ManifestSchema {
		return fmt.Errorf("%w: manifest schema", ErrMismatch)
	}
	if err := validateID("app_id", m.AppID, MaxAppIDLen); err != nil {
		return err
	}
	if err := validateID("app_version", m.AppVersion, MaxAppVersionLen); err != nil {
		return err
	}
	if err := validateID("domain_id", m.DomainID, MaxDomainIDLen); err != nil {
		return err
	}
	if len(m.Capabilities) == 0 || len(m.Capabilities) > MaxCapabilities {
		return fmt.Errorf("%w: capabilities", ErrBoundExceeded)
	}
	if m.MinCoreProtocol == 0 || m.MaxCoreProtocol < m.MinCoreProtocol {
		return fmt.Errorf("%w: protocol range", ErrInvalidState)
	}
	if !slices.IsSorted(m.Capabilities) {
		return fmt.Errorf("%w: capabilities not sorted", ErrNonCanonical)
	}
	for i, capability := range m.Capabilities {
		if err := validateID("capability", capability, MaxCapabilityLen); err != nil {
			return err
		}
		if _, ok := knownCapabilities[capability]; !ok {
			return fmt.Errorf("%w: %s", ErrUnknownCapability, capability)
		}
		if i > 0 && capability == m.Capabilities[i-1] {
			return fmt.Errorf("%w: duplicate capability", ErrNonCanonical)
		}
	}
	return nil
}

func (m Manifest) encode(includeSignature bool) ([]byte, error) {
	if err := m.validate(); err != nil {
		return nil, err
	}
	var e encoder
	e.u16(manifestTypeTag)
	e.u16(m.Version)
	e.str(m.Schema)
	e.str(m.AppID)
	e.str(m.AppVersion)
	e.Write(m.ArtifactHash[:])
	e.str(m.DomainID)
	e.u16(uint16(len(m.Capabilities)))
	for _, capability := range m.Capabilities {
		e.str(capability)
	}
	e.u16(m.MinCoreProtocol)
	e.u16(m.MaxCoreProtocol)
	e.Write(m.PublisherKey[:])
	if includeSignature {
		e.Write(m.Signature[:])
	}
	return e.Bytes(), nil
}

func (m Manifest) signingBytes() ([]byte, error) {
	b, err := m.encode(false)
	if err != nil {
		return nil, err
	}
	return append(append([]byte(nil), manifestDomain...), b...), nil
}

// Sign sets the publisher signature and requires key ownership.
func (m *Manifest) Sign(key ed25519.PrivateKey) error {
	if len(key) != ed25519.PrivateKeySize || !bytes.Equal(key.Public().(ed25519.PublicKey), m.PublisherKey[:]) {
		return ErrMismatch
	}
	b, err := m.signingBytes()
	if err != nil {
		return err
	}
	copy(m.Signature[:], ed25519.Sign(key, b))
	return nil
}

// VerifyPublisher verifies the signed manifest.
func (m Manifest) VerifyPublisher() error {
	b, err := m.signingBytes()
	if err != nil {
		return err
	}
	if !ed25519.Verify(m.PublisherKey[:], b, m.Signature[:]) {
		return ErrInvalidSignature
	}
	return nil
}

// DeclaresCapability reports a canonical declaration, not authorization.
func (m Manifest) DeclaresCapability(capability string) bool {
	if _, ok := knownCapabilities[capability]; !ok {
		return false
	}
	return slices.Contains(m.Capabilities, capability)
}

// Authorizes cross-binds a caller-verified publisher registry fact, exact
// artifact hash, core version, capability and publisher signature.
func (m Manifest) Authorizes(f PublisherFact, artifactHash [sha256.Size]byte, coreProtocol uint16, capability string) error {
	if f.DomainID != m.DomainID || f.AppID != m.AppID || f.AppVersion != m.AppVersion || f.PublisherKey != m.PublisherKey ||
		artifactHash != m.ArtifactHash || coreProtocol < m.MinCoreProtocol || coreProtocol > m.MaxCoreProtocol ||
		!m.DeclaresCapability(capability) {
		return ErrMismatch
	}
	return m.VerifyPublisher()
}

func (m Manifest) MarshalBinary() ([]byte, error) { return m.encode(true) }

// UnmarshalManifest strictly decodes one canonical manifest.
func UnmarshalManifest(b []byte) (Manifest, error) {
	d, err := newDecoder(b)
	if err != nil {
		return Manifest{}, err
	}
	if err := readHeader(d, manifestTypeTag, ManifestVersion); err != nil {
		return Manifest{}, err
	}
	m := Manifest{Version: ManifestVersion}
	if m.Schema, err = d.str(len(ManifestSchema)); err != nil {
		return Manifest{}, err
	}
	if m.AppID, err = d.str(MaxAppIDLen); err != nil {
		return Manifest{}, err
	}
	if m.AppVersion, err = d.str(MaxAppVersionLen); err != nil {
		return Manifest{}, err
	}
	if p, er := d.take(sha256.Size); er != nil {
		return Manifest{}, er
	} else {
		copy(m.ArtifactHash[:], p)
	}
	if m.DomainID, err = d.str(MaxDomainIDLen); err != nil {
		return Manifest{}, err
	}
	n, err := d.u16()
	if err != nil {
		return Manifest{}, err
	}
	if n == 0 || n > MaxCapabilities {
		return Manifest{}, fmt.Errorf("%w: capabilities", ErrBoundExceeded)
	}
	m.Capabilities = make([]string, n)
	for i := range m.Capabilities {
		if m.Capabilities[i], err = d.str(MaxCapabilityLen); err != nil {
			return Manifest{}, err
		}
	}
	if m.MinCoreProtocol, err = d.u16(); err != nil {
		return Manifest{}, err
	}
	if m.MaxCoreProtocol, err = d.u16(); err != nil {
		return Manifest{}, err
	}
	if p, er := d.take(ed25519.PublicKeySize); er != nil {
		return Manifest{}, er
	} else {
		copy(m.PublisherKey[:], p)
	}
	if p, er := d.take(ed25519.SignatureSize); er != nil {
		return Manifest{}, er
	} else {
		copy(m.Signature[:], p)
	}
	if err := d.finish(); err != nil {
		return Manifest{}, err
	}
	if err := m.validate(); err != nil {
		return Manifest{}, err
	}
	reencoded, err := m.MarshalBinary()
	if err != nil || !bytes.Equal(reencoded, b) {
		return Manifest{}, ErrNonCanonical
	}
	return m, nil
}
