package protocol

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"fmt"
)

var (
	envelopeDomain   = []byte("truerepublic.edge-envelope/v1\n")
	envelopeIDDomain = []byte("truerepublic.edge-envelope-id/v1\n")
)

// Envelope is one signed, domain-separated edge message.
type Envelope struct {
	Version          uint16
	ChainID          string
	DomainID         string
	TopicID          string
	CertificateHash  [sha256.Size]byte
	ParentID         [sha256.Size]byte
	AuthorKey        [ed25519.PublicKeySize]byte
	KeyEpoch         uint64
	MembershipHeight uint64
	Sequence         uint64
	Timestamp        uint64
	PayloadHash      [sha256.Size]byte
	Payload          []byte
	Signature        [ed25519.SignatureSize]byte
}

func (e Envelope) validate() error {
	if err := validateVersion(e.Version, EnvelopeVersion); err != nil {
		return err
	}
	if err := validateID("chain_id", e.ChainID, MaxChainIDLen); err != nil {
		return err
	}
	if err := validateID("domain_id", e.DomainID, MaxDomainIDLen); err != nil {
		return err
	}
	if err := validateID("topic_id", e.TopicID, MaxTopicLen); err != nil {
		return err
	}
	if len(e.Payload) > MaxPayloadLen {
		return fmt.Errorf("%w: payload", ErrBoundExceeded)
	}
	want := sha256.Sum256(e.Payload)
	if !bytes.Equal(want[:], e.PayloadHash[:]) {
		return fmt.Errorf("%w: payload hash", ErrMismatch)
	}
	if e.MembershipHeight == 0 || e.Sequence == 0 || e.Timestamp == 0 {
		return fmt.Errorf("%w: zero envelope counter", ErrInvalidState)
	}
	return nil
}

func (e Envelope) encode(includeSignature bool) ([]byte, error) {
	if err := e.validate(); err != nil {
		return nil, err
	}
	var w encoder
	w.u16(envelopeTypeTag)
	w.u16(e.Version)
	w.str(e.ChainID)
	w.str(e.DomainID)
	w.str(e.TopicID)
	w.Write(e.CertificateHash[:])
	w.Write(e.ParentID[:])
	w.Write(e.AuthorKey[:])
	w.u64(e.KeyEpoch)
	w.u64(e.MembershipHeight)
	w.u64(e.Sequence)
	w.u64(e.Timestamp)
	w.Write(e.PayloadHash[:])
	w.sized(e.Payload)
	if includeSignature {
		w.Write(e.Signature[:])
	}
	return w.Bytes(), nil
}

func (e Envelope) signingBytes() ([]byte, error) {
	b, err := e.encode(false)
	if err != nil {
		return nil, err
	}
	return append(append([]byte(nil), envelopeDomain...), b...), nil
}

// Sign sets the message signature and requires the private key to match AuthorKey.
func (e *Envelope) Sign(key ed25519.PrivateKey) error {
	if len(key) != ed25519.PrivateKeySize || !bytes.Equal(key.Public().(ed25519.PublicKey), e.AuthorKey[:]) {
		return ErrMismatch
	}
	b, err := e.signingBytes()
	if err != nil {
		return err
	}
	copy(e.Signature[:], ed25519.Sign(key, b))
	return nil
}

// VerifySignature checks the domain-separated message signature.
func (e Envelope) VerifySignature() error {
	b, err := e.signingBytes()
	if err != nil {
		return err
	}
	if !ed25519.Verify(e.AuthorKey[:], b, e.Signature[:]) {
		return ErrInvalidSignature
	}
	return nil
}

// MessageID deterministically hashes the unsigned canonical envelope.
func (e Envelope) MessageID() ([sha256.Size]byte, error) {
	b, err := e.encode(false)
	if err != nil {
		return [sha256.Size]byte{}, err
	}
	h := sha256.New()
	h.Write(envelopeIDDomain)
	h.Write(b)
	var id [sha256.Size]byte
	copy(id[:], h.Sum(nil))
	return id, nil
}

// verifyBinding cross-binds an envelope, certificate and caller-verified fact.
// The exported VerifyAuthenticatedEnvelope is the complete verification entry.
func (e Envelope) verifyBinding(c Certificate, f MembershipFact) error {
	if err := e.VerifySignature(); err != nil {
		return err
	}
	if err := c.BindMembership(f); err != nil {
		return err
	}
	h, err := c.Hash()
	if err != nil {
		return err
	}
	if e.ChainID != c.ChainID || e.DomainID != c.DomainID || e.KeyEpoch != c.KeyEpoch ||
		e.MembershipHeight != c.MembershipHeight || !bytes.Equal(e.AuthorKey[:], c.MessagingKey[:]) ||
		!bytes.Equal(e.CertificateHash[:], h[:]) {
		return ErrMismatch
	}
	return nil
}

// MarshalBinary returns the canonical envelope bytes.
func (e Envelope) MarshalBinary() ([]byte, error) { return e.encode(true) }

// UnmarshalEnvelope strictly decodes one canonical envelope.
func UnmarshalEnvelope(b []byte) (Envelope, error) {
	d, err := newDecoder(b)
	if err != nil {
		return Envelope{}, err
	}
	if err := readHeader(d, envelopeTypeTag, EnvelopeVersion); err != nil {
		return Envelope{}, err
	}
	e := Envelope{Version: EnvelopeVersion}
	if e.ChainID, err = d.str(MaxChainIDLen); err != nil {
		return Envelope{}, err
	}
	if e.DomainID, err = d.str(MaxDomainIDLen); err != nil {
		return Envelope{}, err
	}
	if e.TopicID, err = d.str(MaxTopicLen); err != nil {
		return Envelope{}, err
	}
	for _, dst := range [][]byte{e.CertificateHash[:], e.ParentID[:], e.AuthorKey[:]} {
		p, er := d.take(len(dst))
		if er != nil {
			return Envelope{}, er
		}
		copy(dst, p)
	}
	if e.KeyEpoch, err = d.u64(); err != nil {
		return Envelope{}, err
	}
	if e.MembershipHeight, err = d.u64(); err != nil {
		return Envelope{}, err
	}
	if e.Sequence, err = d.u64(); err != nil {
		return Envelope{}, err
	}
	if e.Timestamp, err = d.u64(); err != nil {
		return Envelope{}, err
	}
	if p, er := d.take(sha256.Size); er != nil {
		return Envelope{}, er
	} else {
		copy(e.PayloadHash[:], p)
	}
	if e.Payload, err = d.sized(MaxPayloadLen); err != nil {
		return Envelope{}, err
	}
	if p, er := d.take(ed25519.SignatureSize); er != nil {
		return Envelope{}, er
	} else {
		copy(e.Signature[:], p)
	}
	if err := d.finish(); err != nil {
		return Envelope{}, err
	}
	if err := e.validate(); err != nil {
		return Envelope{}, err
	}
	reencoded, err := e.MarshalBinary()
	if err != nil || !bytes.Equal(reencoded, b) {
		return Envelope{}, ErrNonCanonical
	}
	return e, nil
}
