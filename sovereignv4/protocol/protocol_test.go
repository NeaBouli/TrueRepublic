package protocol

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"sync"
	"testing"
)

type fixtures struct {
	accountPrivate   ed25519.PrivateKey
	messagingPrivate ed25519.PrivateKey
	publisherPrivate ed25519.PrivateKey
	fact             MembershipFact
	certificate      Certificate
	envelope         Envelope
	manifest         Manifest
}

func privateKey(fill byte) ed25519.PrivateKey {
	seed := bytes.Repeat([]byte{fill}, ed25519.SeedSize)
	return ed25519.NewKeyFromSeed(seed)
}

func makeFixtures(t testing.TB) fixtures {
	t.Helper()
	f := fixtures{
		accountPrivate:   privateKey(1),
		messagingPrivate: privateKey(2),
		publisherPrivate: privateKey(3),
	}
	f.fact = MembershipFact{
		ChainID: "truerepublic-4", DomainID: "athens.civic", MemberAccount: "tr1member",
		Height: 42, ProofHash: sha256.Sum256([]byte("verified-membership-proof")),
	}
	copy(f.fact.AccountKey[:], f.accountPrivate.Public().(ed25519.PublicKey))
	f.certificate = Certificate{
		Version: CertificateVersion, ChainID: f.fact.ChainID, DomainID: f.fact.DomainID,
		MemberAccount: f.fact.MemberAccount, MembershipHeight: f.fact.Height,
		MembershipProofHash: f.fact.ProofHash, AccountKey: f.fact.AccountKey,
		KeyEpoch: 7, NotBefore: 1_700_000_000, NotAfter: 1_800_000_000,
	}
	copy(f.certificate.MessagingKey[:], f.messagingPrivate.Public().(ed25519.PublicKey))
	if err := f.certificate.Sign(f.accountPrivate); err != nil {
		t.Fatal(err)
	}
	certificateHash, err := f.certificate.Hash()
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte("canonical civic message")
	f.envelope = Envelope{
		Version: EnvelopeVersion, ChainID: f.fact.ChainID, DomainID: f.fact.DomainID,
		TopicID: "bill.2026.0042", CertificateHash: certificateHash, KeyEpoch: 7,
		MembershipHeight: 42, Sequence: 9, Timestamp: 1_710_000_000,
		Payload: payload, PayloadHash: sha256.Sum256(payload),
	}
	copy(f.envelope.AuthorKey[:], f.messagingPrivate.Public().(ed25519.PublicKey))
	if err := f.envelope.Sign(f.messagingPrivate); err != nil {
		t.Fatal(err)
	}
	f.manifest = Manifest{
		Version: ManifestVersion, Schema: ManifestSchema, AppID: "civic.reader", AppVersion: "1.0.0",
		ArtifactHash: sha256.Sum256([]byte("domain-app-artifact")), DomainID: f.fact.DomainID,
		Capabilities:    []string{CapabilityRegisteredQuery, CapabilityDiscussionRead},
		MinCoreProtocol: 1, MaxCoreProtocol: 1,
	}
	copy(f.manifest.PublisherKey[:], f.publisherPrivate.Public().(ed25519.PublicKey))
	if err := f.manifest.Sign(f.publisherPrivate); err != nil {
		t.Fatal(err)
	}
	return f
}

func TestCanonicalRoundTrip(t *testing.T) {
	f := makeFixtures(t)
	tests := []struct {
		name    string
		marshal func() ([]byte, error)
		decode  func([]byte) ([]byte, error)
	}{
		{"certificate", f.certificate.MarshalBinary, func(b []byte) ([]byte, error) {
			v, e := UnmarshalCertificate(b)
			if e != nil {
				return nil, e
			}
			return v.MarshalBinary()
		}},
		{"envelope", f.envelope.MarshalBinary, func(b []byte) ([]byte, error) {
			v, e := UnmarshalEnvelope(b)
			if e != nil {
				return nil, e
			}
			return v.MarshalBinary()
		}},
		{"manifest", f.manifest.MarshalBinary, func(b []byte) ([]byte, error) {
			v, e := UnmarshalManifest(b)
			if e != nil {
				return nil, e
			}
			return v.MarshalBinary()
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original, err := tt.marshal()
			if err != nil {
				t.Fatal(err)
			}
			got, err := tt.decode(original)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, original) {
				t.Fatal("round trip changed canonical bytes")
			}
		})
	}
}

type goldenVectors struct {
	CertificateHex string `json:"certificate_hex"`
	CertificateSHA string `json:"certificate_sha256"`
	EnvelopeHex    string `json:"envelope_hex"`
	MessageID      string `json:"message_id"`
	ManifestHex    string `json:"manifest_hex"`
}

func TestGoldenVectors(t *testing.T) {
	f := makeFixtures(t)
	certificate, _ := f.certificate.MarshalBinary()
	envelope, _ := f.envelope.MarshalBinary()
	manifest, _ := f.manifest.MarshalBinary()
	id, _ := f.envelope.MessageID()
	want := goldenVectors{
		CertificateHex: hex.EncodeToString(certificate), CertificateSHA: hashHex(certificate),
		EnvelopeHex: hex.EncodeToString(envelope), MessageID: hex.EncodeToString(id[:]), ManifestHex: hex.EncodeToString(manifest),
	}
	b, err := os.ReadFile("testdata/vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var got goldenVectors
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.CertificateHex == "" {
		encoded, _ := json.MarshalIndent(want, "", "  ")
		t.Fatalf("populate testdata/vectors.json with:\n%s", encoded)
	}
	if got != want {
		t.Fatalf("golden vectors changed\n got: %+v\nwant: %+v", got, want)
	}
}

func hashHex(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

func TestCertificateMembershipAndSignature(t *testing.T) {
	f := makeFixtures(t)
	if err := f.certificate.VerifyAccountSignature(); err != nil {
		t.Fatal(err)
	}
	if err := f.certificate.BindMembership(f.fact); err != nil {
		t.Fatal(err)
	}
	bad := f.fact
	bad.Height++
	if !errors.Is(f.certificate.BindMembership(bad), ErrMismatch) {
		t.Fatal("height mismatch accepted")
	}
	bad = f.fact
	copy(bad.AccountKey[:], privateKey(9).Public().(ed25519.PublicKey))
	if !errors.Is(f.certificate.BindMembership(bad), ErrMismatch) {
		t.Fatal("unbound account key accepted")
	}
	badCertificate := f.certificate
	copy(badCertificate.AccountKey[:], privateKey(9).Public().(ed25519.PublicKey))
	if !errors.Is(badCertificate.VerifyAccountSignature(), ErrInvalidSignature) {
		t.Fatal("wrong account key accepted")
	}
}

func TestAuthenticatedEnvelopeDecisions(t *testing.T) {
	f := makeFixtures(t)
	state := RevocationState{HighestEpoch: 7, SyncedHeight: 50, Online: true}
	decision, err := VerifyAuthenticatedEnvelope(f.envelope, f.certificate, f.fact, state, 1_720_000_000)
	if err != nil || decision.Status != AuthValid || decision.Reason != ReasonCurrent {
		t.Fatalf("got %+v, %v", decision, err)
	}
	state.Online = false
	decision, err = VerifyAuthenticatedEnvelope(f.envelope, f.certificate, f.fact, state, 1_720_000_000)
	if err != nil || decision.Status != AuthProvisional || decision.Reason != ReasonRevocationUnchecked {
		t.Fatalf("got %+v, %v", decision, err)
	}
}

func TestRevocationRotationExpiryMatrix(t *testing.T) {
	f := makeFixtures(t)
	tests := []struct {
		name  string
		state RevocationState
		now   uint64
		want  AuthDecision
	}{
		{"not-yet-valid", RevocationState{Online: true}, f.certificate.NotBefore - 1, AuthDecision{Status: AuthRejected, Reason: ReasonNotYetValid}},
		{"expired-online", RevocationState{Online: true}, f.certificate.NotAfter + 1, AuthDecision{Status: AuthRejected, Reason: ReasonExpired}},
		{"expired-offline", RevocationState{}, f.certificate.NotAfter + 1, AuthDecision{Status: AuthProvisional, Reason: ReasonExpiredOffline}},
		{"revocation-unchecked", RevocationState{}, f.certificate.NotBefore + 1, AuthDecision{Status: AuthProvisional, Reason: ReasonRevocationUnchecked}},
		{"revocation-not-synced", RevocationState{Online: true, RevocationHeight: 50, SyncedHeight: 49}, f.certificate.NotBefore + 1, AuthDecision{Status: AuthProvisional, Reason: ReasonRevocationNotSynced}},
		{"revoked", RevocationState{Online: true, RevocationHeight: 50, SyncedHeight: 50, RevokedThroughEpoch: 7}, f.certificate.NotBefore + 1, AuthDecision{Status: AuthRejected, Reason: ReasonRevoked}},
		{"superseded", RevocationState{Online: true, HighestEpoch: 8}, f.certificate.NotBefore + 1, AuthDecision{Status: AuthRejected, Reason: ReasonSuperseded}},
		{"future-epoch", RevocationState{Online: true, HighestEpoch: 6}, f.certificate.NotBefore + 1, AuthDecision{Status: AuthProvisional, Reason: ReasonEpochUnverified}},
		{"invalid-state", RevocationState{Online: true, RevokedThroughEpoch: 1}, f.certificate.NotBefore + 1, AuthDecision{Status: AuthRejected, Reason: ReasonInvalidState}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.Evaluate(f.certificate, tt.now); got != tt.want {
				t.Fatalf("got %+v want %+v", got, tt.want)
			}
		})
	}
	attackerCertificate := f.certificate
	attackerCertificate.KeyEpoch = 1_000
	if err := attackerCertificate.Sign(f.accountPrivate); err != nil {
		t.Fatal(err)
	}
	attackerEnvelope := f.envelope
	attackerEnvelope.KeyEpoch = attackerCertificate.KeyEpoch
	attackerEnvelope.CertificateHash, _ = attackerCertificate.Hash()
	if err := attackerEnvelope.Sign(f.messagingPrivate); err != nil {
		t.Fatal(err)
	}
	state := RevocationState{Online: true, HighestEpoch: 8, RevokedThroughEpoch: 7, RevocationHeight: 50, SyncedHeight: 50}
	decision, err := VerifyAuthenticatedEnvelope(attackerEnvelope, attackerCertificate, f.fact, state, f.certificate.NotBefore+1)
	if err != nil || decision.Status != AuthProvisional || decision.Reason != ReasonEpochUnverified {
		t.Fatalf("future epoch bypassed revocation: %+v, %v", decision, err)
	}
}

func TestCrossDomainAndChainRejected(t *testing.T) {
	f := makeFixtures(t)
	for _, mutate := range []func(*Envelope){
		func(e *Envelope) { e.ChainID = "foreign-1" },
		func(e *Envelope) { e.DomainID = "other.domain" },
		func(e *Envelope) { e.KeyEpoch++ },
		func(e *Envelope) { e.MembershipHeight++ },
	} {
		e := f.envelope
		mutate(&e)
		if err := e.Sign(f.messagingPrivate); err != nil {
			t.Fatal(err)
		}
		if _, err := VerifyAuthenticatedEnvelope(e, f.certificate, f.fact, RevocationState{HighestEpoch: 7, Online: true}, f.certificate.NotBefore+1); !errors.Is(err, ErrMismatch) {
			t.Fatal("cross-binding mutation accepted")
		}
	}
}

func TestEnvelopeStableIDAndPayloadBinding(t *testing.T) {
	f := makeFixtures(t)
	id1, err := f.envelope.MessageID()
	if err != nil {
		t.Fatal(err)
	}
	id2, _ := f.envelope.MessageID()
	if id1 != id2 {
		t.Fatal("message ID is not stable")
	}
	e := f.envelope
	e.Payload = []byte("changed")
	e.PayloadHash = sha256.Sum256(e.Payload)
	if err := e.Sign(f.messagingPrivate); err != nil {
		t.Fatal(err)
	}
	id3, _ := e.MessageID()
	if id1 == id3 {
		t.Fatal("payload change did not change message ID")
	}
	e.Payload[0] ^= 1
	if _, err := e.MarshalBinary(); !errors.Is(err, ErrMismatch) {
		t.Fatal("payload hash mismatch accepted")
	}
}

func TestReplayDuplicateEquivocationAndConcurrency(t *testing.T) {
	f := makeFixtures(t)
	store := NewReplayStore(16)
	valid, err := VerifyAuthenticatedEnvelope(f.envelope, f.certificate, f.fact, RevocationState{HighestEpoch: 7, Online: true}, f.certificate.NotBefore+1)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := store.Observe(f.envelope, valid); err != nil || got != ConflictNew {
		t.Fatalf("got %v, %v", got, err)
	}
	if got, err := store.Observe(f.envelope, valid); err != nil || got != ConflictDuplicate {
		t.Fatalf("got %v, %v", got, err)
	}
	otherDomain := f.envelope
	otherDomain.DomainID = "other.domain"
	if err := otherDomain.Sign(f.messagingPrivate); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Observe(otherDomain, valid); !errors.Is(err, ErrMismatch) {
		t.Fatalf("authentication decision reused across domains: %v", err)
	}
	otherFact := f.fact
	otherFact.DomainID = otherDomain.DomainID
	otherCertificate := f.certificate
	otherCertificate.DomainID = otherDomain.DomainID
	if err := otherCertificate.Sign(f.accountPrivate); err != nil {
		t.Fatal(err)
	}
	otherDomain.CertificateHash, _ = otherCertificate.Hash()
	if err := otherDomain.Sign(f.messagingPrivate); err != nil {
		t.Fatal(err)
	}
	otherValid, err := VerifyAuthenticatedEnvelope(otherDomain, otherCertificate, otherFact, RevocationState{HighestEpoch: 7, Online: true}, f.certificate.NotBefore+1)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := store.Observe(otherDomain, otherValid); err != nil || got != ConflictNew {
		t.Fatalf("cross-domain sequence falsely equivocated: %v, %v", got, err)
	}
	e := f.envelope
	e.Payload = []byte("equivocation")
	e.PayloadHash = sha256.Sum256(e.Payload)
	if err := e.Sign(f.messagingPrivate); err != nil {
		t.Fatal(err)
	}
	eValid := valid
	eValid.envelopeID, _ = e.MessageID()
	if got, err := store.Observe(e, eValid); err != nil || got != ConflictEquivocation {
		t.Fatalf("got %v, %v", got, err)
	}
	parallel := NewReplayStore(64)
	var wg sync.WaitGroup
	for range 32 {
		wg.Add(1)
		go func() { defer wg.Done(); _, _ = parallel.Observe(f.envelope, valid) }()
	}
	wg.Wait()
}

func TestReplayStoreFailsClosedAtBoundAndOnInvalidSignature(t *testing.T) {
	f := makeFixtures(t)
	store := NewReplayStore(1)
	valid, err := VerifyAuthenticatedEnvelope(f.envelope, f.certificate, f.fact, RevocationState{HighestEpoch: 7, Online: true}, f.certificate.NotBefore+1)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := store.Observe(f.envelope, valid); err != nil || got != ConflictNew {
		t.Fatalf("got %v, %v", got, err)
	}
	invalid := f.envelope
	invalid.Signature[0] ^= 1
	if _, err := store.Observe(invalid, valid); !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("invalid signature entered store: %v", err)
	}
	next := f.envelope
	next.Sequence++
	next.Timestamp++
	if err := next.Sign(f.messagingPrivate); err != nil {
		t.Fatal(err)
	}
	nextValid := valid
	nextValid.envelopeID, _ = next.MessageID()
	if _, err := store.Observe(next, nextValid); !errors.Is(err, ErrBoundExceeded) {
		t.Fatalf("replay store exceeded bound: %v", err)
	}
	if NewReplayStore(0) != nil || NewReplayStore(MaxReplayEntries+1) != nil {
		t.Fatal("invalid replay limit accepted")
	}
	if _, err := NewReplayStore(1).Observe(f.envelope, AuthDecision{Status: AuthProvisional, Reason: ReasonRevocationUnchecked}); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("provisional envelope entered store: %v", err)
	}
	if _, err := NewReplayStore(1).Observe(f.envelope, AuthDecision{Status: AuthValid, Reason: ReasonCurrent}); !errors.Is(err, ErrMismatch) {
		t.Fatalf("unbound valid decision entered store: %v", err)
	}
}

func TestManifestDenyByDefaultAndSignature(t *testing.T) {
	f := makeFixtures(t)
	if err := f.manifest.VerifyPublisher(); err != nil {
		t.Fatal(err)
	}
	if !f.manifest.DeclaresCapability(CapabilityDiscussionRead) || f.manifest.DeclaresCapability("network.raw") {
		t.Fatal("capability policy incorrect")
	}
	publisher := PublisherFact{DomainID: f.fact.DomainID, AppID: f.manifest.AppID, AppVersion: f.manifest.AppVersion, PublisherKey: f.manifest.PublisherKey}
	if err := f.manifest.Authorizes(publisher, f.manifest.ArtifactHash, 1, CapabilityDiscussionRead); err != nil {
		t.Fatal(err)
	}
	publisher.DomainID = "other.domain"
	if !errors.Is(f.manifest.Authorizes(publisher, f.manifest.ArtifactHash, 1, CapabilityDiscussionRead), ErrMismatch) {
		t.Fatal("cross-domain manifest accepted")
	}
	publisher = PublisherFact{DomainID: f.fact.DomainID, AppID: f.manifest.AppID, AppVersion: f.manifest.AppVersion, PublisherKey: f.manifest.PublisherKey}
	publisher.AppVersion = "2.0.0"
	if !errors.Is(f.manifest.Authorizes(publisher, f.manifest.ArtifactHash, 1, CapabilityDiscussionRead), ErrMismatch) {
		t.Fatal("wrong app version accepted")
	}
	publisher.AppVersion = f.manifest.AppVersion
	wrongArtifact := sha256.Sum256([]byte("different-artifact"))
	if !errors.Is(f.manifest.Authorizes(publisher, wrongArtifact, 1, CapabilityDiscussionRead), ErrMismatch) {
		t.Fatal("wrong artifact accepted")
	}
	publisher.PublisherKey[0] ^= 1
	if !errors.Is(f.manifest.Authorizes(publisher, f.manifest.ArtifactHash, 1, CapabilityDiscussionRead), ErrMismatch) {
		t.Fatal("unregistered publisher accepted")
	}
	m := f.manifest
	m.Capabilities = append(m.Capabilities, "network.raw")
	if err := m.Sign(f.publisherPrivate); !errors.Is(err, ErrUnknownCapability) {
		t.Fatal("unknown capability accepted")
	}
}

func TestManifestCanonicalCapabilityOrder(t *testing.T) {
	f := makeFixtures(t)
	m := f.manifest
	m.Capabilities[0], m.Capabilities[1] = m.Capabilities[1], m.Capabilities[0]
	if _, err := m.MarshalBinary(); !errors.Is(err, ErrNonCanonical) {
		t.Fatal("unsorted capabilities accepted")
	}
	m = f.manifest
	m.Capabilities = []string{CapabilityDiscussionRead, CapabilityDiscussionRead}
	if _, err := m.MarshalBinary(); !errors.Is(err, ErrNonCanonical) {
		t.Fatal("duplicate capabilities accepted")
	}
}

func TestStrictDecodersRejectTrailingWrongTypeAndVersion(t *testing.T) {
	f := makeFixtures(t)
	certificate, _ := f.certificate.MarshalBinary()
	if _, err := UnmarshalCertificate(append(certificate, 0)); !errors.Is(err, ErrTrailingData) {
		t.Fatalf("trailing data: %v", err)
	}
	wrongType := append([]byte(nil), certificate...)
	wrongType[1] = envelopeTypeTag
	if _, err := UnmarshalCertificate(wrongType); !errors.Is(err, ErrWrongType) {
		t.Fatalf("wrong type: %v", err)
	}
	wrongVersion := append([]byte(nil), certificate...)
	wrongVersion[3] = 2
	if _, err := UnmarshalCertificate(wrongVersion); !errors.Is(err, ErrUnsupportedVersion) {
		t.Fatalf("wrong version: %v", err)
	}
}

func TestBoundsAndMalformedIdentifiers(t *testing.T) {
	f := makeFixtures(t)
	c := f.certificate
	c.ChainID = string(bytes.Repeat([]byte{'a'}, MaxChainIDLen+1))
	if _, err := c.MarshalBinary(); !errors.Is(err, ErrMalformedID) {
		t.Fatalf("oversized ID: %v", err)
	}
	c = f.certificate
	c.DomainID = "bad domain"
	if _, err := c.MarshalBinary(); !errors.Is(err, ErrMalformedID) {
		t.Fatalf("malformed ID: %v", err)
	}
	e := f.envelope
	e.Payload = make([]byte, MaxPayloadLen+1)
	e.PayloadHash = sha256.Sum256(e.Payload)
	if _, err := e.MarshalBinary(); !errors.Is(err, ErrBoundExceeded) {
		t.Fatalf("oversized payload: %v", err)
	}
	if _, err := UnmarshalEnvelope(make([]byte, MaxEncodedMessage+1)); !errors.Is(err, ErrBoundExceeded) {
		t.Fatalf("oversized wire: %v", err)
	}
}

func TestMutationsNeverVerify(t *testing.T) {
	f := makeFixtures(t)
	t.Run("certificate", func(t *testing.T) {
		original, _ := f.certificate.MarshalBinary()
		for i := range original {
			mutated := append([]byte(nil), original...)
			mutated[i] ^= 1
			c, err := UnmarshalCertificate(mutated)
			if err == nil && c.VerifyAccountSignature() == nil {
				t.Fatalf("mutation %d verified", i)
			}
		}
	})
	t.Run("envelope", func(t *testing.T) {
		original, _ := f.envelope.MarshalBinary()
		for i := range original {
			mutated := append([]byte(nil), original...)
			mutated[i] ^= 1
			e, err := UnmarshalEnvelope(mutated)
			if err == nil {
				decision, verifyErr := VerifyAuthenticatedEnvelope(e, f.certificate, f.fact, RevocationState{HighestEpoch: 7, Online: true}, f.certificate.NotBefore+1)
				if verifyErr == nil && decision.Status == AuthValid {
					t.Fatalf("mutation %d verified", i)
				}
			}
		}
	})
	t.Run("manifest", func(t *testing.T) {
		original, _ := f.manifest.MarshalBinary()
		for i := range original {
			mutated := append([]byte(nil), original...)
			mutated[i] ^= 1
			m, err := UnmarshalManifest(mutated)
			if err == nil && m.VerifyPublisher() == nil {
				t.Fatalf("mutation %d verified", i)
			}
		}
	})
}

func TestDeterministicPropertyRoundTrips(t *testing.T) {
	f := makeFixtures(t)
	rng := rand.New(rand.NewSource(241))
	for i := 1; i <= 300; i++ {
		e := f.envelope
		e.Sequence = uint64(i)
		e.Timestamp += uint64(i)
		e.Payload = make([]byte, rng.Intn(1024))
		_, _ = rng.Read(e.Payload)
		e.PayloadHash = sha256.Sum256(e.Payload)
		if err := e.Sign(f.messagingPrivate); err != nil {
			t.Fatal(err)
		}
		first, err := e.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		second, _ := e.MarshalBinary()
		if !bytes.Equal(first, second) {
			t.Fatal("non-deterministic encoding")
		}
		decoded, err := UnmarshalEnvelope(first)
		if err != nil {
			t.Fatal(err)
		}
		decision, err := VerifyAuthenticatedEnvelope(decoded, f.certificate, f.fact, RevocationState{HighestEpoch: 7, Online: true}, f.certificate.NotBefore+1)
		if err != nil || decision.Status != AuthValid {
			t.Fatalf("round-trip verification: %+v, %v", decision, err)
		}
	}
}
