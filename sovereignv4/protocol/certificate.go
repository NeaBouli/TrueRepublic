package protocol

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"fmt"
)

var certificateDomain = []byte("truerepublic.edge-certificate/v1\n")

// MembershipFact is the exact membership result already verified by a caller's
// TRChain light client. ProofHash binds the external proof bytes without
// claiming that this package validated their cryptography.
type MembershipFact struct {
	ChainID       string
	DomainID      string
	MemberAccount string
	AccountKey    [ed25519.PublicKeySize]byte
	Height        uint64
	ProofHash     [sha256.Size]byte
}

func (f MembershipFact) validate() error {
	if err := validateID("chain_id", f.ChainID, MaxChainIDLen); err != nil {
		return err
	}
	if err := validateID("domain_id", f.DomainID, MaxDomainIDLen); err != nil {
		return err
	}
	if err := validateID("member_account", f.MemberAccount, MaxAccountLen); err != nil {
		return err
	}
	if f.Height == 0 {
		return fmt.Errorf("%w: zero membership height", ErrInvalidState)
	}
	return nil
}

// Certificate authorizes one domain-scoped Ed25519 messaging key. Signature is
// produced by the chain account key over the canonical unsigned certificate.
type Certificate struct {
	Version             uint16
	ChainID             string
	DomainID            string
	MemberAccount       string
	AccountKey          [ed25519.PublicKeySize]byte
	MembershipHeight    uint64
	MembershipProofHash [sha256.Size]byte
	MessagingKey        [ed25519.PublicKeySize]byte
	KeyEpoch            uint64
	NotBefore           uint64
	NotAfter            uint64
	Signature           [ed25519.SignatureSize]byte
}

func (c Certificate) validate() error {
	if err := validateVersion(c.Version, CertificateVersion); err != nil {
		return err
	}
	if err := (MembershipFact{c.ChainID, c.DomainID, c.MemberAccount, c.AccountKey, c.MembershipHeight, c.MembershipProofHash}).validate(); err != nil {
		return err
	}
	if c.NotAfter <= c.NotBefore {
		return fmt.Errorf("%w: invalid validity interval", ErrInvalidState)
	}
	return nil
}

func (c Certificate) encode(includeSignature bool) ([]byte, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}
	var e encoder
	e.u16(certificateTypeTag)
	e.u16(c.Version)
	e.str(c.ChainID)
	e.str(c.DomainID)
	e.str(c.MemberAccount)
	e.Write(c.AccountKey[:])
	e.u64(c.MembershipHeight)
	e.Write(c.MembershipProofHash[:])
	e.Write(c.MessagingKey[:])
	e.u64(c.KeyEpoch)
	e.u64(c.NotBefore)
	e.u64(c.NotAfter)
	if includeSignature {
		e.Write(c.Signature[:])
	}
	return e.Bytes(), nil
}

func (c Certificate) signingBytes() ([]byte, error) {
	b, err := c.encode(false)
	if err != nil {
		return nil, err
	}
	return append(append([]byte(nil), certificateDomain...), b...), nil
}

// Sign sets the account authorization signature.
func (c *Certificate) Sign(accountKey ed25519.PrivateKey) error {
	if len(accountKey) != ed25519.PrivateKeySize || !bytes.Equal(accountKey.Public().(ed25519.PublicKey), c.AccountKey[:]) {
		return ErrMismatch
	}
	b, err := c.signingBytes()
	if err != nil {
		return err
	}
	copy(c.Signature[:], ed25519.Sign(accountKey, b))
	return nil
}

// VerifyAccountSignature verifies the chain account's authorization signature.
func (c Certificate) VerifyAccountSignature() error {
	b, err := c.signingBytes()
	if err != nil {
		return err
	}
	if !ed25519.Verify(c.AccountKey[:], b, c.Signature[:]) {
		return ErrInvalidSignature
	}
	return nil
}

// BindMembership requires every caller-verified membership field to match.
func (c Certificate) BindMembership(f MembershipFact) error {
	if err := c.validate(); err != nil {
		return err
	}
	if err := f.validate(); err != nil {
		return err
	}
	if c.ChainID != f.ChainID || c.DomainID != f.DomainID ||
		c.MemberAccount != f.MemberAccount || c.AccountKey != f.AccountKey || c.MembershipHeight != f.Height ||
		!bytes.Equal(c.MembershipProofHash[:], f.ProofHash[:]) {
		return ErrMismatch
	}
	return nil
}

// MarshalBinary returns the canonical certificate bytes.
func (c Certificate) MarshalBinary() ([]byte, error) { return c.encode(true) }

// Hash returns the SHA-256 digest of the complete canonical certificate.
func (c Certificate) Hash() ([sha256.Size]byte, error) {
	b, err := c.MarshalBinary()
	if err != nil {
		return [sha256.Size]byte{}, err
	}
	return sha256.Sum256(b), nil
}

// UnmarshalCertificate strictly decodes one canonical certificate.
func UnmarshalCertificate(b []byte) (Certificate, error) {
	d, err := newDecoder(b)
	if err != nil {
		return Certificate{}, err
	}
	if err := readHeader(d, certificateTypeTag, CertificateVersion); err != nil {
		return Certificate{}, err
	}
	c := Certificate{Version: CertificateVersion}
	if c.ChainID, err = d.str(MaxChainIDLen); err != nil {
		return Certificate{}, err
	}
	if c.DomainID, err = d.str(MaxDomainIDLen); err != nil {
		return Certificate{}, err
	}
	if c.MemberAccount, err = d.str(MaxAccountLen); err != nil {
		return Certificate{}, err
	}
	if p, e := d.take(ed25519.PublicKeySize); e != nil {
		return Certificate{}, e
	} else {
		copy(c.AccountKey[:], p)
	}
	if c.MembershipHeight, err = d.u64(); err != nil {
		return Certificate{}, err
	}
	if p, e := d.take(sha256.Size); e != nil {
		return Certificate{}, e
	} else {
		copy(c.MembershipProofHash[:], p)
	}
	if p, e := d.take(ed25519.PublicKeySize); e != nil {
		return Certificate{}, e
	} else {
		copy(c.MessagingKey[:], p)
	}
	if c.KeyEpoch, err = d.u64(); err != nil {
		return Certificate{}, err
	}
	if c.NotBefore, err = d.u64(); err != nil {
		return Certificate{}, err
	}
	if c.NotAfter, err = d.u64(); err != nil {
		return Certificate{}, err
	}
	if p, e := d.take(ed25519.SignatureSize); e != nil {
		return Certificate{}, e
	} else {
		copy(c.Signature[:], p)
	}
	if err := d.finish(); err != nil {
		return Certificate{}, err
	}
	if err := c.validate(); err != nil {
		return Certificate{}, err
	}
	reencoded, err := c.MarshalBinary()
	if err != nil || !bytes.Equal(reencoded, b) {
		return Certificate{}, ErrNonCanonical
	}
	return c, nil
}
