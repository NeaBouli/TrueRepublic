package protocol

import (
	"crypto/ed25519"
	"crypto/sha256"
	"sync"
)

type AuthStatus uint8

const (
	AuthValid AuthStatus = iota + 1
	AuthProvisional
	AuthRejected
)

const (
	ReasonCurrent             = "current"
	ReasonNotYetValid         = "not-yet-valid"
	ReasonExpired             = "expired"
	ReasonExpiredOffline      = "expired-offline"
	ReasonRevocationUnchecked = "revocation-unchecked"
	ReasonRevocationNotSynced = "revocation-not-synced"
	ReasonRevoked             = "revoked"
	ReasonSuperseded          = "superseded"
	ReasonEpochUnverified     = "epoch-unverified"
	ReasonInvalidState        = "invalid-state"
)

const MaxReplayEntries = 1 << 16

type AuthDecision struct {
	Status AuthStatus
	Reason string
	// envelopeID is deliberately unexported: only this package can bind an
	// AuthValid decision to the exact canonical envelope that was verified.
	envelopeID [sha256.Size]byte
}

// RevocationState is caller-supplied committed-chain state for one member.
type RevocationState struct {
	HighestEpoch        uint64
	RevokedThroughEpoch uint64
	RevocationHeight    uint64
	SyncedHeight        uint64
	Online              bool
}

// Evaluate never upgrades stale/offline state to authenticated.
func (s RevocationState) Evaluate(c Certificate, now uint64) AuthDecision {
	if s.RevokedThroughEpoch > 0 && s.RevocationHeight == 0 {
		return AuthDecision{Status: AuthRejected, Reason: ReasonInvalidState}
	}
	if now < c.NotBefore {
		return AuthDecision{Status: AuthRejected, Reason: ReasonNotYetValid}
	}
	if now > c.NotAfter {
		if !s.Online {
			return AuthDecision{Status: AuthProvisional, Reason: ReasonExpiredOffline}
		}
		return AuthDecision{Status: AuthRejected, Reason: ReasonExpired}
	}
	if !s.Online {
		return AuthDecision{Status: AuthProvisional, Reason: ReasonRevocationUnchecked}
	}
	if s.RevocationHeight > 0 && s.SyncedHeight < s.RevocationHeight {
		return AuthDecision{Status: AuthProvisional, Reason: ReasonRevocationNotSynced}
	}
	if s.RevocationHeight > 0 && c.KeyEpoch <= s.RevokedThroughEpoch {
		return AuthDecision{Status: AuthRejected, Reason: ReasonRevoked}
	}
	if c.KeyEpoch < s.HighestEpoch {
		return AuthDecision{Status: AuthRejected, Reason: ReasonSuperseded}
	}
	if c.KeyEpoch > s.HighestEpoch {
		return AuthDecision{Status: AuthProvisional, Reason: ReasonEpochUnverified}
	}
	return AuthDecision{Status: AuthValid, Reason: ReasonCurrent}
}

// VerifyAuthenticatedEnvelope performs all package-owned checks. The caller is
// still responsible for verifying the light-client proof behind fact.
func VerifyAuthenticatedEnvelope(e Envelope, c Certificate, fact MembershipFact, state RevocationState, now uint64) (AuthDecision, error) {
	if err := c.VerifyAccountSignature(); err != nil {
		return AuthDecision{Status: AuthRejected, Reason: "certificate-signature"}, err
	}
	if err := e.verifyBinding(c, fact); err != nil {
		return AuthDecision{Status: AuthRejected, Reason: "envelope-binding"}, err
	}
	decision := state.Evaluate(c, now)
	if decision.Status == AuthValid {
		id, err := e.MessageID()
		if err != nil {
			return AuthDecision{Status: AuthRejected, Reason: "envelope-id"}, err
		}
		decision.envelopeID = id
	}
	return decision, nil
}

type Conflict uint8

const (
	ConflictNew Conflict = iota + 1
	ConflictDuplicate
	ConflictEquivocation
	ConflictCorruption
)

type sequenceKey struct {
	Scope    [sha256.Size]byte
	Author   [ed25519.PublicKeySize]byte
	Epoch    uint64
	Sequence uint64
}

// ReplayStore detects idempotent duplicates and same-sequence equivocation.
type ReplayStore struct {
	mu         sync.Mutex
	byID       map[[sha256.Size]byte][sha256.Size]byte
	bySequence map[sequenceKey][sha256.Size]byte
	limit      int
}

// NewReplayStore creates a fail-closed bounded store. Invalid limits return nil.
func NewReplayStore(limit int) *ReplayStore {
	if limit <= 0 || limit > MaxReplayEntries {
		return nil
	}
	return &ReplayStore{byID: make(map[[sha256.Size]byte][sha256.Size]byte), bySequence: make(map[sequenceKey][sha256.Size]byte), limit: limit}
}

// Observe records an envelope only when authentication was produced for that
// exact envelope by VerifyAuthenticatedEnvelope. Capacity is a hard fail-closed
// signal: callers must checkpoint verified history and rebuild a new bounded
// store rather than silently evict replay evidence.
func (s *ReplayStore) Observe(e Envelope, authentication AuthDecision) (Conflict, error) {
	if s == nil {
		return ConflictCorruption, ErrInvalidState
	}
	if authentication.Status != AuthValid {
		return ConflictCorruption, ErrInvalidState
	}
	if err := e.VerifySignature(); err != nil {
		return ConflictCorruption, err
	}
	b, err := e.MarshalBinary()
	if err != nil {
		return ConflictCorruption, err
	}
	id, err := e.MessageID()
	if err != nil {
		return ConflictCorruption, err
	}
	if authentication.envelopeID != id {
		return ConflictCorruption, ErrMismatch
	}
	scope := sha256.Sum256([]byte(e.ChainID + "\x00" + e.DomainID))
	key := sequenceKey{scope, e.AuthorKey, e.KeyEpoch, e.Sequence}
	canonicalHash := sha256.Sum256(b)
	s.mu.Lock()
	defer s.mu.Unlock()
	if previous, ok := s.byID[id]; ok {
		if previous == canonicalHash {
			return ConflictDuplicate, nil
		}
		return ConflictCorruption, nil
	}
	if previousID, ok := s.bySequence[key]; ok && previousID != id {
		return ConflictEquivocation, nil
	}
	if len(s.byID) >= s.limit {
		return ConflictCorruption, ErrBoundExceeded
	}
	s.byID[id] = canonicalHash
	s.bySequence[key] = id
	return ConflictNew, nil
}
