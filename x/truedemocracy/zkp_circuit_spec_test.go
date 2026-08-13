package truedemocracy

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

// GH-203 freezes the test-only ZKP circuit/encoding contract in a strict
// versioned machine-readable specification and cross-checks it fail-closed
// against Go constants and behavior, the GH-198 fixture manifest and golden
// vector, the active go.mod gnark toolchain, and the maintained-client
// zkpEncoding contract. Groth16 proving/verifying keys and proofs are never
// regenerated or compared here: groth16.Setup samples random toxic waste, so
// only the deterministic compiled constraint system is pinned by parity.

const (
	zkpCircuitSpecPath   = "../../configs/security/zkp-circuit.json"
	zkpCircuitSpecSchema = "truerepublic/zkp-circuit/v1"
	zkpGoModPath         = "../../go.mod"

	zkpSpecClassification = "TEST-ONLY SINGLE-PARTY TOXIC WASTE"
	zkpSpecCanonicalRule  = "exactly 32 big-endian bytes; values >= scalar field modulus are rejected"
	zkpSpecConstants      = "110 round constants: keccak256 iterated starting from keccak256(\"seed\"), each digest reduced modulo the BN254 scalar field"
	zkpSpecNodeRule       = "MiMC(left, right) per level; pathIndex 0 = current node is the left child, 1 = right child"
	zkpSpecCommitment     = "commitment = MiMC(identitySecret); identitySecret is 1-32 bytes, big-endian"
	zkpSpecExternal       = "externalNullifier = hashToField(voteContext(chainID, domainName, issueName, suggestionName)); rating-independent, stable per suggestion"
	zkpSpecNullifier      = "nullifierHash = MiMC(identitySecret, externalNullifier)"
	zkpSpecSignal         = "signalHash = hashToField(voteContext(chainID, domainName, issueName, suggestionName, rating)); binds chain, domain, issue, suggestion and the exact rating"
	zkpSpecRatingEncoding = "int64 big-endian, appended only for the signal"
	zkpSpecHashToField    = "SHA-256(data) as big-endian integer reduced modulo the BN254 scalar field, serialized as 32 big-endian bytes"
	zkpSpecClientContract = "client-web/src/services/zkpEncoding.ts"
	zkpSpecClientGuard    = "client-web/src/services/zkp.ts isSubmittable must return false"
)

var zkpSpecPublicInputOrder = []string{"merkle_root", "nullifier_hash", "external_nullifier", "signal_hash"}
var zkpSpecVoteContextFields = []string{"chain_id", "domain_name", "issue_name", "suggestion_name"}

type zkpCircuitSpec struct {
	Schema                 string             `json:"schema"`
	CircuitID              string             `json:"circuit_id"`
	Classification         string             `json:"classification"`
	ProductionAllowed      bool               `json:"production_allowed"`
	Curve                  zkpSpecCurve       `json:"curve"`
	Hash                   zkpSpecHash        `json:"hash"`
	Merkle                 zkpSpecMerkle      `json:"merkle"`
	IdentityCommitment     string             `json:"identity_commitment"`
	ExternalNullifier      string             `json:"external_nullifier"`
	NullifierHash          string             `json:"nullifier_hash"`
	SignalHash             string             `json:"signal_hash"`
	VoteContextEncoding    zkpSpecVoteContext `json:"vote_context_encoding"`
	HashToField            string             `json:"hash_to_field"`
	PublicInputOrder       []string           `json:"public_input_order"`
	Toolchain              zkpSpecToolchain   `json:"toolchain"`
	Fixtures               zkpSpecFixtures    `json:"fixtures"`
	ClientEncodingContract string             `json:"client_encoding_contract"`
	ClientSubmissionGuard  string             `json:"client_submission_guard"`
	Exclusions             []string           `json:"exclusions"`
}

type zkpSpecCurve struct {
	Name               string `json:"name"`
	ScalarFieldModulus string `json:"scalar_field_modulus"`
	ScalarFieldBytes   int    `json:"scalar_field_bytes"`
	ByteEncoding       string `json:"byte_encoding"`
	CanonicalRule      string `json:"canonical_rule"`
}

type zkpSpecHash struct {
	Algorithm    string `json:"algorithm"`
	Rounds       int    `json:"rounds"`
	Exponent     int    `json:"exponent"`
	Seed         string `json:"seed"`
	Constants    string `json:"constants"`
	Construction string `json:"construction"`
}

type zkpSpecMerkle struct {
	Depth    int    `json:"depth"`
	Leaf     string `json:"leaf"`
	ZeroLeaf string `json:"zero_leaf"`
	NodeRule string `json:"node_rule"`
}

type zkpSpecVoteContext struct {
	DomainSeparator    string   `json:"domain_separator"`
	Fields             []string `json:"fields"`
	StringLengthPrefix string   `json:"string_length_prefix"`
	RatingEncoding     string   `json:"rating_encoding"`
}

type zkpSpecToolchain struct {
	Gnark       string `json:"gnark"`
	GnarkCrypto string `json:"gnark_crypto"`
}

type zkpSpecFixtures struct {
	Directory          string `json:"directory"`
	Manifest           string `json:"manifest"`
	ManifestSchema     string `json:"manifest_schema"`
	GoldenVector       string `json:"golden_vector"`
	GoldenVectorSchema string `json:"golden_vector_schema"`
	ConstraintSystem   string `json:"constraint_system"`
}

// parseGnarkToolchain extracts the active direct go.mod requirements for
// gnark and gnark-crypto. Each module must appear exactly once as a direct
// (non-indirect) require line.
func parseGnarkToolchain(goMod string) (gnark string, gnarkCrypto string, err error) {
	find := func(module string) (string, error) {
		pattern := regexp.MustCompile(`(?m)^\t` + regexp.QuoteMeta(module) + ` (v[0-9]+\.[0-9]+\.[0-9]+)$`)
		matches := pattern.FindAllStringSubmatch(goMod, -1)
		if len(matches) != 1 {
			return "", fmt.Errorf("go.mod must contain exactly one direct require for %s, found %d", module, len(matches))
		}
		return matches[0][1], nil
	}
	gnark, err = find("github.com/consensys/gnark")
	if err != nil {
		return "", "", err
	}
	gnarkCrypto, err = find("github.com/consensys/gnark-crypto")
	if err != nil {
		return "", "", err
	}
	return gnark, gnarkCrypto, nil
}

// zkpCircuitSpecViolations cross-checks the decoded specification against live
// Go constants and the active go.mod toolchain versions. Any drift fails
// closed with a non-empty violation list.
func zkpCircuitSpecViolations(spec zkpCircuitSpec, gnarkVersion, gnarkCryptoVersion string) []string {
	var violations []string
	wantString := func(field, got, want string) {
		if got != want {
			violations = append(violations, fmt.Sprintf("%s = %q, want %q", field, got, want))
		}
	}
	wantInt := func(field string, got, want int) {
		if got != want {
			violations = append(violations, fmt.Sprintf("%s = %d, want %d", field, got, want))
		}
	}
	wantStrings := func(field string, got, want []string) {
		if fmt.Sprint(got) != fmt.Sprint(want) {
			violations = append(violations, fmt.Sprintf("%s = %v, want %v", field, got, want))
		}
	}

	wantString("schema", spec.Schema, zkpCircuitSpecSchema)
	wantString("circuit_id", spec.CircuitID, MembershipCircuitID)
	wantString("classification", spec.Classification, zkpSpecClassification)
	if spec.ProductionAllowed {
		violations = append(violations, "production_allowed must be false")
	}

	wantString("curve.name", spec.Curve.Name, "BN254")
	wantString("curve.scalar_field_modulus", spec.Curve.ScalarFieldModulus, ecc.BN254.ScalarField().String())
	if fr.Modulus().String() != ecc.BN254.ScalarField().String() {
		violations = append(violations, "gnark-crypto fr modulus disagrees with ecc.BN254 scalar field")
	}
	wantInt("curve.scalar_field_bytes", spec.Curve.ScalarFieldBytes, 32)
	wantString("curve.byte_encoding", spec.Curve.ByteEncoding, "big-endian")
	wantString("curve.canonical_rule", spec.Curve.CanonicalRule, zkpSpecCanonicalRule)

	wantString("hash.algorithm", spec.Hash.Algorithm, "MiMC")
	wantInt("hash.rounds", spec.Hash.Rounds, 110)
	wantInt("hash.exponent", spec.Hash.Exponent, 5)
	wantString("hash.seed", spec.Hash.Seed, "seed")
	wantString("hash.constants", spec.Hash.Constants, zkpSpecConstants)
	wantString("hash.construction", spec.Hash.Construction, "Miyaguchi-Preneel")

	wantInt("merkle.depth", spec.Merkle.Depth, MerkleTreeDepth)
	wantString("merkle.leaf", spec.Merkle.Leaf, "identity_commitment")
	wantString("merkle.zero_leaf", spec.Merkle.ZeroLeaf, "MiMC(0)")
	wantString("merkle.node_rule", spec.Merkle.NodeRule, zkpSpecNodeRule)

	wantString("identity_commitment", spec.IdentityCommitment, zkpSpecCommitment)
	wantString("external_nullifier", spec.ExternalNullifier, zkpSpecExternal)
	wantString("nullifier_hash", spec.NullifierHash, zkpSpecNullifier)
	wantString("signal_hash", spec.SignalHash, zkpSpecSignal)
	wantString("vote_context_encoding.domain_separator", spec.VoteContextEncoding.DomainSeparator, "TrueRepublic/vote/v1")
	wantStrings("vote_context_encoding.fields", spec.VoteContextEncoding.Fields, zkpSpecVoteContextFields)
	wantString("vote_context_encoding.string_length_prefix", spec.VoteContextEncoding.StringLengthPrefix, "uint32 big-endian")
	wantString("vote_context_encoding.rating_encoding", spec.VoteContextEncoding.RatingEncoding, zkpSpecRatingEncoding)
	wantString("hash_to_field", spec.HashToField, zkpSpecHashToField)

	wantStrings("public_input_order", spec.PublicInputOrder, zkpSpecPublicInputOrder)
	if len(spec.PublicInputOrder) != membershipPublicWitnessCount {
		violations = append(violations, fmt.Sprintf("public input count = %d, want %d", len(spec.PublicInputOrder), membershipPublicWitnessCount))
	}

	wantString("toolchain.gnark", spec.Toolchain.Gnark, gnarkVersion)
	wantString("toolchain.gnark_crypto", spec.Toolchain.GnarkCrypto, gnarkCryptoVersion)

	wantString("fixtures.directory", spec.Fixtures.Directory, "x/truedemocracy/testdata/zkp")
	wantString("fixtures.manifest", spec.Fixtures.Manifest, "manifest.json")
	wantString("fixtures.manifest_schema", spec.Fixtures.ManifestSchema, zkpFixtureSchema)
	wantString("fixtures.golden_vector", spec.Fixtures.GoldenVector, "golden_vector.json")
	wantString("fixtures.golden_vector_schema", spec.Fixtures.GoldenVectorSchema, "truerepublic/zkp-golden-vector/v1")
	wantString("fixtures.constraint_system", spec.Fixtures.ConstraintSystem, "membership_v2.cs")

	wantString("client_encoding_contract", spec.ClientEncodingContract, zkpSpecClientContract)
	wantString("client_submission_guard", spec.ClientSubmissionGuard, zkpSpecClientGuard)

	joinedExclusions := strings.Join(spec.Exclusions, "\n")
	for _, required := range []string{
		"groth16.Setup samples random toxic waste",
		"NOT reproducible",
		"no production ceremony",
		"isSubmittable remains hard false",
	} {
		if !strings.Contains(joinedExclusions, required) {
			violations = append(violations, fmt.Sprintf("exclusions must cover %q", required))
		}
	}
	return violations
}

func loadZKPCircuitSpec(t *testing.T) zkpCircuitSpec {
	t.Helper()
	raw, err := os.ReadFile(zkpCircuitSpecPath)
	if err != nil {
		t.Fatal(err)
	}
	var spec zkpCircuitSpec
	if err := strictJSON(raw, &spec); err != nil {
		t.Fatalf("strict circuit spec decode: %v", err)
	}
	goMod, err := os.ReadFile(zkpGoModPath)
	if err != nil {
		t.Fatal(err)
	}
	gnark, gnarkCrypto, err := parseGnarkToolchain(string(goMod))
	if err != nil {
		t.Fatal(err)
	}
	if violations := zkpCircuitSpecViolations(spec, gnark, gnarkCrypto); len(violations) != 0 {
		t.Fatalf("zkp circuit spec violations:\n- %s", strings.Join(violations, "\n- "))
	}
	return spec
}

func TestZKPCircuitSpecMatchesGoConstantsAndBehavior(t *testing.T) {
	spec := loadZKPCircuitSpec(t)
	_, vector, _, _ := readVerifiedZKPFixture(t)

	witness := fixtureHex(t, "synthetic witness", vector.SyntheticWitnessHex)
	commitment, err := ComputeCommitment(witness)
	if err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(commitment); got != vector.CommitmentHex {
		t.Fatalf("identity commitment = %s, want %s", got, vector.CommitmentHex)
	}

	external := ComputeVoteNullifierScope(vector.ChainID, vector.DomainName, vector.IssueName, vector.SuggestionName)
	if got := hex.EncodeToString(external); got != vector.ExternalNullifierHex {
		t.Fatalf("external nullifier = %s, want %s", got, vector.ExternalNullifierHex)
	}

	// Independently re-derive hashToField from the spec formula instead of
	// calling hashToField, so an implementation drift cannot self-agree.
	digest := sha256.Sum256(encodeVoteContext(vector.ChainID, vector.DomainName, vector.IssueName, vector.SuggestionName, nil))
	reduced := new(big.Int).Mod(new(big.Int).SetBytes(digest[:]), ecc.BN254.ScalarField())
	var independent [32]byte
	reduced.FillBytes(independent[:])
	if !bytes.Equal(independent[:], external) {
		t.Fatalf("spec hash_to_field formula disagrees with ComputeVoteNullifierScope")
	}

	signal := ComputeVoteSignal(vector.ChainID, vector.DomainName, vector.IssueName, vector.SuggestionName, vector.Rating)
	if got := hex.EncodeToString(signal); got != vector.SignalHashHex {
		t.Fatalf("signal hash = %s, want %s", got, vector.SignalHashHex)
	}
	nullifier, err := ComputeNullifier(witness, external)
	if err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(nullifier); got != vector.NullifierHashHex {
		t.Fatalf("nullifier hash = %s, want %s", got, vector.NullifierHashHex)
	}
	nonCanonical := ecc.BN254.ScalarField().FillBytes(make([]byte, 32))
	if _, err := ComputeCommitment(nonCanonical); err == nil {
		t.Fatal("spec requires rejection of non-canonical BN254 field elements")
	}

	siblings := make([][]byte, len(vector.SiblingsHex))
	for i, sibling := range vector.SiblingsHex {
		siblings[i] = fixtureHex(t, "sibling", sibling)
	}
	root := fixtureHex(t, "merkle root", vector.MerkleRootHex)
	if !VerifyMerkleProof(root, commitment, siblings, vector.PathIndices) {
		t.Fatal("golden merkle path does not reproduce the pinned root")
	}

	if spec.CircuitID != MembershipCircuitID || spec.Merkle.Depth != MerkleTreeDepth {
		t.Fatal("spec identity drifted from Go circuit constants")
	}
}

func TestZKPCircuitSpecMatchesFixtureManifestAndVector(t *testing.T) {
	spec := loadZKPCircuitSpec(t)
	manifest, vector, _, _ := readVerifiedZKPFixture(t)

	if manifest.Schema != spec.Fixtures.ManifestSchema || manifest.CircuitID != spec.CircuitID ||
		manifest.Curve != spec.Curve.Name || manifest.MerkleDepth != spec.Merkle.Depth ||
		manifest.Gnark != spec.Toolchain.Gnark || manifest.GnarkCrypto != spec.Toolchain.GnarkCrypto ||
		manifest.Classification != spec.Classification || manifest.ProductionAllowed != spec.ProductionAllowed {
		t.Fatal("fixture manifest drifted from the circuit spec")
	}
	if fmt.Sprint(manifest.PublicInputOrder) != fmt.Sprint(spec.PublicInputOrder) {
		t.Fatalf("manifest public input order = %v, want %v", manifest.PublicInputOrder, spec.PublicInputOrder)
	}
	if vector.Schema != spec.Fixtures.GoldenVectorSchema || vector.CircuitID != spec.CircuitID || !vector.SyntheticAndTestOnly {
		t.Fatal("golden vector metadata drifted from the circuit spec")
	}
}

func TestZKPCircuitSpecMatchesGoModToolchain(t *testing.T) {
	spec := loadZKPCircuitSpec(t)
	goMod, err := os.ReadFile(zkpGoModPath)
	if err != nil {
		t.Fatal(err)
	}
	gnark, gnarkCrypto, err := parseGnarkToolchain(string(goMod))
	if err != nil {
		t.Fatal(err)
	}
	if gnark != spec.Toolchain.Gnark || gnarkCrypto != spec.Toolchain.GnarkCrypto {
		t.Fatalf("active go.mod toolchain %s/%s drifted from spec %s/%s",
			gnark, gnarkCrypto, spec.Toolchain.Gnark, spec.Toolchain.GnarkCrypto)
	}

	t.Run("indirect requirement rejected", func(t *testing.T) {
		candidate := "module example\nrequire (\n\tgithub.com/consensys/gnark v0.14.0 // indirect\n\tgithub.com/consensys/gnark-crypto v0.19.2\n)\n"
		if _, _, err := parseGnarkToolchain(candidate); err == nil {
			t.Fatal("indirect gnark require accepted as direct")
		}
	})
	t.Run("missing module rejected", func(t *testing.T) {
		candidate := "module example\nrequire (\n\tgithub.com/consensys/gnark v0.14.0\n)\n"
		if _, _, err := parseGnarkToolchain(candidate); err == nil {
			t.Fatal("missing gnark-crypto require accepted")
		}
	})
	t.Run("duplicate module rejected", func(t *testing.T) {
		candidate := "module example\nrequire (\n\tgithub.com/consensys/gnark v0.14.0\n\tgithub.com/consensys/gnark v0.14.0\n\tgithub.com/consensys/gnark-crypto v0.19.2\n)\n"
		if _, _, err := parseGnarkToolchain(candidate); err == nil {
			t.Fatal("duplicate gnark require accepted")
		}
	})
	t.Run("prefixed module not matched", func(t *testing.T) {
		candidate := "module example\nrequire (\n\tgithub.com/consensys/gnark-crypto v0.19.2\n\tgithub.com/ingonyama-zk/icicle-gnark/v3 v3.2.2\n)\n"
		if _, _, err := parseGnarkToolchain(candidate); err == nil {
			t.Fatal("icicle-gnark path matched the gnark module")
		}
	})
}

func TestZKPCircuitSpecMatchesMaintainedClientContract(t *testing.T) {
	spec := loadZKPCircuitSpec(t)
	encoding, err := os.ReadFile(filepath.Join("../..", spec.ClientEncodingContract))
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		spec.Curve.ScalarFieldModulus,
		fmt.Sprintf("MIMC_ROUNDS = %d", spec.Hash.Rounds),
		"'" + spec.VoteContextEncoding.DomainSeparator + "'",
		"keccak256",
		"sha256",
		"setUint32(0, value, false)",
		"setBigInt64(0, BigInt(value), false)",
	} {
		if !strings.Contains(string(encoding), required) {
			t.Fatalf("maintained-client zkpEncoding contract missing %q", required)
		}
	}

	guard, err := os.ReadFile("../../client-web/src/services/zkp.ts")
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`get isSubmittable\(\): boolean \{\s*return false;\s*\}`).Match(guard) {
		t.Fatal("client isSubmittable is not a hard-false getter")
	}
}

// TestZKPCircuitConstraintSystemParity compiles MembershipCircuit in memory,
// serializes only the constraint system, and proves byte/hash parity with the
// pinned GH-198 fixture without modifying it. Proving keys, verifying keys
// and proofs are deliberately excluded: groth16.Setup samples random toxic
// waste, so their bytes are not reproducible.
func TestZKPCircuitConstraintSystemParity(t *testing.T) {
	spec := loadZKPCircuitSpec(t)
	pinnedPath := filepath.Join(zkpFixtureDir, spec.Fixtures.ConstraintSystem)
	before, err := os.ReadFile(pinnedPath)
	if err != nil {
		t.Fatal(err)
	}
	beforeDigest := sha256.Sum256(before)

	var circuit MembershipCircuit
	cs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	if err != nil {
		t.Fatalf("in-memory circuit compilation failed: %v", err)
	}
	var compiled bytes.Buffer
	if _, err := cs.WriteTo(&compiled); err != nil {
		t.Fatalf("constraint system serialization failed: %v", err)
	}
	compiledDigest := sha256.Sum256(compiled.Bytes())
	if !bytes.Equal(compiled.Bytes(), before) {
		t.Fatalf("compiled constraint system drifted from pinned fixture:\ncompiled sha256 %x (%d bytes)\npinned   sha256 %x (%d bytes)",
			compiledDigest, compiled.Len(), beforeDigest, len(before))
	}

	after, err := os.ReadFile(pinnedPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, before) {
		t.Fatal("parity check mutated the pinned constraint-system fixture")
	}
}

func TestZKPCircuitSpecRejectsDrift(t *testing.T) {
	spec := loadZKPCircuitSpec(t)
	gnark, gnarkCrypto := spec.Toolchain.Gnark, spec.Toolchain.GnarkCrypto

	cases := map[string]func(*zkpCircuitSpec){
		"production allowed":      func(s *zkpCircuitSpec) { s.ProductionAllowed = true },
		"weakened classification": func(s *zkpCircuitSpec) { s.Classification = "PRODUCTION" },
		"wrong modulus": func(s *zkpCircuitSpec) {
			s.Curve.ScalarFieldModulus = "21888242871839275222246405745257275088548364400416034343698204186575808495618"
		},
		"wrong byte encoding": func(s *zkpCircuitSpec) { s.Curve.ByteEncoding = "little-endian" },
		"wrong rounds":        func(s *zkpCircuitSpec) { s.Hash.Rounds = 109 },
		"wrong seed":          func(s *zkpCircuitSpec) { s.Hash.Seed = "random" },
		"wrong depth":         func(s *zkpCircuitSpec) { s.Merkle.Depth = 19 },
		"wrong circuit id":    func(s *zkpCircuitSpec) { s.CircuitID = "truerepublic/membership-vote/v3-bn254-mimc-depth20" },
		"reordered public inputs": func(s *zkpCircuitSpec) {
			s.PublicInputOrder[0], s.PublicInputOrder[1] = s.PublicInputOrder[1], s.PublicInputOrder[0]
		},
		"wrong gnark version":        func(s *zkpCircuitSpec) { s.Toolchain.Gnark = "v0.15.0" },
		"wrong gnark-crypto version": func(s *zkpCircuitSpec) { s.Toolchain.GnarkCrypto = "v0.20.0" },
		"missing exclusions":         func(s *zkpCircuitSpec) { s.Exclusions = nil },
		"wrong client contract":      func(s *zkpCircuitSpec) { s.ClientEncodingContract = "client-web/src/services/zkp.ts" },
		"wrong fixture directory":    func(s *zkpCircuitSpec) { s.Fixtures.Directory = "testdata/zkp" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			candidate := spec
			// The decoded spec contains slices. Clone them before mutation so a
			// negative case cannot contaminate the shared valid baseline or make
			// later cases pass for the wrong reason under randomized map order.
			candidate.PublicInputOrder = append([]string(nil), spec.PublicInputOrder...)
			candidate.VoteContextEncoding.Fields = append([]string(nil), spec.VoteContextEncoding.Fields...)
			candidate.Exclusions = append([]string(nil), spec.Exclusions...)
			mutate(&candidate)
			if violations := zkpCircuitSpecViolations(candidate, gnark, gnarkCrypto); len(violations) == 0 {
				t.Fatal("drifted circuit spec accepted")
			}
		})
	}

	raw, err := os.ReadFile(zkpCircuitSpecPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("unknown field rejected", func(t *testing.T) {
		candidate := bytes.Replace(raw, []byte("\n}"), []byte(",\n  \"unexpected\": true\n}"), 1)
		var decoded zkpCircuitSpec
		if err := strictJSON(candidate, &decoded); err == nil {
			t.Fatal("unknown spec field accepted")
		}
		candidate = bytes.Replace(raw, []byte("{\n"), []byte("{\n  \"schema\": \"duplicate\",\n"), 1)
		if err := strictJSON(candidate, &decoded); err == nil {
			t.Fatal("duplicate spec field accepted")
		}
	})
	t.Run("trailing JSON rejected", func(t *testing.T) {
		candidate := append(append([]byte(nil), raw...), []byte(" {}")...)
		var decoded zkpCircuitSpec
		if err := strictJSON(candidate, &decoded); err == nil {
			t.Fatal("trailing JSON value accepted")
		}
	})
	t.Run("toolchain drift detected", func(t *testing.T) {
		if violations := zkpCircuitSpecViolations(spec, "v0.15.0", gnarkCrypto); len(violations) == 0 {
			t.Fatal("go.mod gnark drift accepted")
		}
		if violations := zkpCircuitSpecViolations(spec, gnark, "v0.20.0"); len(violations) == 0 {
			t.Fatal("go.mod gnark-crypto drift accepted")
		}
	})
}
