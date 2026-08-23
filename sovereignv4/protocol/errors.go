package protocol

import "errors"

var (
	ErrUnsupportedVersion = errors.New("sovereign v4 protocol: unsupported version")
	ErrWrongType          = errors.New("sovereign v4 protocol: wrong message type")
	ErrTrailingData       = errors.New("sovereign v4 protocol: trailing data")
	ErrNonCanonical       = errors.New("sovereign v4 protocol: non-canonical value")
	ErrBoundExceeded      = errors.New("sovereign v4 protocol: size bound exceeded")
	ErrMalformedID        = errors.New("sovereign v4 protocol: malformed identifier")
	ErrInvalidSignature   = errors.New("sovereign v4 protocol: invalid signature")
	ErrMismatch           = errors.New("sovereign v4 protocol: binding mismatch")
	ErrUnknownCapability  = errors.New("sovereign v4 protocol: unknown capability")
	ErrInvalidState       = errors.New("sovereign v4 protocol: invalid state")
)
