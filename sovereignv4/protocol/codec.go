package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unicode/utf8"
)

const (
	CertificateVersion uint16 = 1
	EnvelopeVersion    uint16 = 1
	ManifestVersion    uint16 = 1

	MaxChainIDLen      = 128
	MaxDomainIDLen     = 128
	MaxAccountLen      = 128
	MaxTopicLen        = 256
	MaxAppIDLen        = 128
	MaxAppVersionLen   = 64
	MaxCapabilityLen   = 64
	MaxCapabilities    = 16
	MaxPayloadLen      = 64 << 10
	MaxEncodedMessage  = 1 << 20
	certificateTypeTag = 1
	envelopeTypeTag    = 2
	manifestTypeTag    = 3
)

type encoder struct{ bytes.Buffer }

func (e *encoder) u16(v uint16) {
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], v)
	e.Write(b[:])
}

func (e *encoder) u64(v uint64) {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], v)
	e.Write(b[:])
}

func (e *encoder) sized(b []byte) {
	var n [4]byte
	binary.BigEndian.PutUint32(n[:], uint32(len(b)))
	e.Write(n[:])
	e.Write(b)
}

func (e *encoder) str(v string) { e.sized([]byte(v)) }

type decoder struct {
	b   []byte
	off int
}

func newDecoder(b []byte) (*decoder, error) {
	if len(b) > MaxEncodedMessage {
		return nil, fmt.Errorf("%w: encoded message", ErrBoundExceeded)
	}
	return &decoder{b: b}, nil
}

func (d *decoder) take(n int) ([]byte, error) {
	if n < 0 || n > len(d.b)-d.off {
		return nil, fmt.Errorf("%w: truncated message", ErrNonCanonical)
	}
	v := d.b[d.off : d.off+n]
	d.off += n
	return v, nil
}

func (d *decoder) u16() (uint16, error) {
	b, err := d.take(2)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(b), nil
}

func (d *decoder) u64() (uint64, error) {
	b, err := d.take(8)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(b), nil
}

func (d *decoder) sized(max int) ([]byte, error) {
	b, err := d.take(4)
	if err != nil {
		return nil, err
	}
	n := uint64(binary.BigEndian.Uint32(b))
	if n > uint64(max) {
		return nil, fmt.Errorf("%w: length %d exceeds %d", ErrBoundExceeded, n, max)
	}
	return d.take(int(n))
}

func (d *decoder) str(max int) (string, error) {
	b, err := d.sized(max)
	if err != nil {
		return "", err
	}
	if !utf8.Valid(b) {
		return "", fmt.Errorf("%w: invalid UTF-8", ErrNonCanonical)
	}
	return string(b), nil
}

func (d *decoder) finish() error {
	if d.off != len(d.b) {
		return ErrTrailingData
	}
	return nil
}

func validateID(name, value string, max int) error {
	if value == "" || len(value) > max {
		return fmt.Errorf("%w: %s", ErrMalformedID, name)
	}
	for i := range len(value) {
		c := value[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || bytes.ContainsRune([]byte("-._:/@+"), rune(c)) {
			continue
		}
		return fmt.Errorf("%w: %s", ErrMalformedID, name)
	}
	return nil
}

func validateVersion(got, want uint16) error {
	if got != want {
		return fmt.Errorf("%w: got %d want %d", ErrUnsupportedVersion, got, want)
	}
	return nil
}

func readHeader(d *decoder, tag, version uint16) error {
	gotTag, err := d.u16()
	if err != nil {
		return err
	}
	if gotTag != tag {
		return fmt.Errorf("%w: got %d want %d", ErrWrongType, gotTag, tag)
	}
	gotVersion, err := d.u16()
	if err != nil {
		return err
	}
	return validateVersion(gotVersion, version)
}
