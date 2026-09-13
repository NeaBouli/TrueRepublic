package truedemocracy

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
)

// GH-297 freezes the production-target ZKP protocol contract in a strict
// versioned manifest that digest-binds the existing GH-203/GH-209 test-only
// circuit specification and pins the versioned public inputs, field encoding,
// vote signal and nullifier semantics plus fail-closed change-control rules.
// This test is read-only: it never runs Groth16 setup, never regenerates or
// compares proving artifacts, and never mutates the manifest, the source
// spec, or any runtime behavior. production_ready and submission_allowed stay
// hard false.

const (
	zkpProtocolFreezePath   = "../../configs/security/zkp-protocol-freeze.json"
	zkpProtocolFreezeSchema = "truerepublic/zkp-protocol-freeze/v1"

	zkpFreezeStatus         = "frozen"
	zkpFreezeClassification = "FROZEN PRODUCTION-CANDIDATE SPECIFICATION"
	zkpFreezeRolloutItem    = "phase-2-frozen-versioned-circuit-inputs-encodings-nullifier-rules"

	zkpFreezeSourcePath      = "configs/security/zkp-circuit.json"
	zkpFreezeSourceSchema    = "truerepublic/zkp-circuit/v2"
	zkpFreezeSourceSHA256    = "27fb8e4eaeb365706535f107c0d6a39a282e423f01385c14495f3f06414c4ab7"
	zkpFreezeArtifactClass   = "TEST-ONLY SINGLE-PARTY TOXIC WASTE"
	zkpFreezeClientGuardPath = "../../client-web/src/services/zkp.ts"

	zkpFreezeProfileProtocol   = "truerepublic/anonymous-rating/v2"
	zkpFreezeProfilePublicIn   = "truerepublic/zkp-public-inputs/v1"
	zkpFreezeProfileFieldEnc   = "truerepublic/bn254-field-be32-canonical/v1"
	zkpFreezeNullifierScope    = "TrueRepublic/vote/v1"
	zkpFreezeSignalProfile     = "TrueRepublic/vote/v2"
	zkpFreezeProfileRecipient  = "truerepublic/bech32-canonical-recipient/v1"
	zkpFreezeConsensusVersion  = 2
	zkpFreezeFieldElementCanon = "exactly 32-byte big-endian canonical BN254 scalar field element"
	zkpFreezeLegacyUse         = "nullifier-scope derivation and frozen synthetic fixture compatibility only"
	zkpFreezeCandidateUse      = "canonical recipient-bound anonymous-rating signal"

	zkpFreezeActivation   = "fresh genesis or an explicit governed consensus upgrade with deterministic migration evidence"
	zkpFreezeReviewForced = "publish a new version and migration plan; never mutate this frozen profile in place"
)

var zkpFreezePublicInputOrder = []string{"merkle_root", "nullifier_hash", "external_nullifier", "signal_hash"}
var zkpFreezeNullifierBoundFields = []string{"chain_id", "domain_name", "issue_name", "suggestion_name"}
var zkpFreezeNullifierExcludedFields = []string{"rating", "reward_recipient"}
var zkpFreezeSignalBoundFields = []string{"chain_id", "domain_name", "issue_name", "suggestion_name", "rating", "reward_recipient"}

var zkpFreezeCircuitBumpRequiredFor = []string{
	"constraint_system", "curve", "hash_construction", "merkle_depth",
	"identity_commitment", "nullifier_hash", "public_input_order",
}
var zkpFreezeProtocolBumpRequiredFor = []string{
	"field_element_encoding", "vote_context_domain", "vote_context_field_order",
	"rating_encoding", "reward_recipient_encoding",
}
var zkpFreezeNullifierKeyspaceRequiredFor = []string{
	"nullifier_scope_domain", "nullifier_bound_fields", "nullifier_hash_construction",
}
var zkpFreezeOpenGates = []string{
	"compatible real Groth16 maintained-client prover",
	"reproducible production proving and verification artifacts",
	"ceremony provenance and artifact rotation policy",
	"browser-to-chain production proof compatibility",
	"independent cryptographic privacy and setup review",
}
var zkpFreezeEvidencePaths = []string{
	"configs/security/zkp-circuit.json",
	"x/truedemocracy/zkpcircuit/circuit.go",
	"x/truedemocracy/zkp.go",
	"x/truedemocracy/merkle.go",
	"x/truedemocracy/zkp_circuit_spec_test.go",
	"x/truedemocracy/zkp_genesis_security_test.go",
	"client-web/src/services/zkpEncoding.ts",
	"client-web/src/services/zkp.ts",
}

type zkpProtocolFreezeManifest struct {
	Schema            string                 `json:"schema"`
	Status            string                 `json:"status"`
	Classification    string                 `json:"classification"`
	ProductionReady   bool                   `json:"production_ready"`
	SubmissionAllowed bool                   `json:"submission_allowed"`
	RolloutItem       string                 `json:"rollout_item"`
	SourceSpec        zkpFreezeSourceSpec    `json:"source_spec"`
	Profiles          zkpFreezeProfiles      `json:"profiles"`
	Semantics         zkpFreezeSemantics     `json:"semantics"`
	ChangeControl     zkpFreezeChangeControl `json:"change_control"`
	OpenGates         []string               `json:"open_gates"`
	Evidence          []string               `json:"evidence"`
}

type zkpFreezeSourceSpec struct {
	Path                      string `json:"path"`
	Schema                    string `json:"schema"`
	SHA256                    string `json:"sha256"`
	ArtifactClassification    string `json:"artifact_classification"`
	ArtifactProductionAllowed bool   `json:"artifact_production_allowed"`
}

type zkpFreezeProfiles struct {
	Protocol         string `json:"protocol"`
	Circuit          string `json:"circuit"`
	ConsensusVersion int    `json:"consensus_version"`
	PublicInputs     string `json:"public_inputs"`
	FieldEncoding    string `json:"field_encoding"`
	NullifierScope   string `json:"nullifier_scope"`
	Signal           string `json:"signal"`
	RewardRecipient  string `json:"reward_recipient"`
}

type zkpFreezeSemantics struct {
	PublicInputOrder        []string             `json:"public_input_order"`
	FieldElementEncoding    string               `json:"field_element_encoding"`
	NullifierBoundFields    []string             `json:"nullifier_bound_fields"`
	NullifierExcludedFields []string             `json:"nullifier_excluded_fields"`
	SignalBoundFields       []string             `json:"signal_bound_fields"`
	LegacyContext           zkpFreezeVoteProfile `json:"legacy_context"`
	CandidateContext        zkpFreezeVoteProfile `json:"candidate_context"`
}

type zkpFreezeVoteProfile struct {
	Profile          string `json:"profile"`
	Use              string `json:"use"`
	ProductionSignal bool   `json:"production_signal"`
}

type zkpFreezeChangeControl struct {
	InPlaceSemanticChangeAllowed    bool     `json:"in_place_semantic_change_allowed"`
	CircuitIDBumpRequiredFor        []string `json:"circuit_id_bump_required_for"`
	ProtocolProfileBumpRequiredFor  []string `json:"protocol_profile_bump_required_for"`
	NewNullifierKeyspaceRequiredFor []string `json:"new_nullifier_keyspace_required_for"`
	Activation                      string   `json:"activation"`
	ReviewForcedChange              string   `json:"review_forced_change"`
}

// zkpFreezeUnsafePathReason rejects evidence paths that are not normalized,
// safe, repository-relative forward-slash paths.
func zkpFreezeUnsafePathReason(path string) string {
	switch {
	case path == "":
		return "empty path"
	case filepath.IsAbs(path) || strings.HasPrefix(path, "/"):
		return "absolute path"
	case strings.Contains(path, "\\"):
		return "backslash separator"
	case filepath.ToSlash(filepath.Clean(path)) != path:
		return "not normalized"
	}
	for _, segment := range strings.Split(path, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return "unsafe segment " + strconv.Quote(segment)
		}
	}
	return ""
}

// zkpFreezeEvidenceFileReason reports why a resolved filesystem path is not a
// regular non-symlink evidence file, or "" when it is acceptable.
func zkpFreezeEvidenceFileReason(absPath string) string {
	info, err := os.Lstat(absPath)
	if err != nil {
		return "missing"
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "symlink"
	}
	if !info.Mode().IsRegular() {
		return "not a regular file"
	}
	return ""
}

// zkpFreezeEvidenceListViolations pins the evidence set exactly: unique,
// normalized safe repository-relative paths in the declared order.
func zkpFreezeEvidenceListViolations(paths []string) []string {
	var violations []string
	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		if reason := zkpFreezeUnsafePathReason(path); reason != "" {
			violations = append(violations, fmt.Sprintf("evidence path %q is unsafe: %s", path, reason))
		}
		if _, duplicate := seen[path]; duplicate {
			violations = append(violations, fmt.Sprintf("evidence path %q is duplicated", path))
		}
		seen[path] = struct{}{}
	}
	if fmt.Sprint(paths) != fmt.Sprint(zkpFreezeEvidencePaths) {
		violations = append(violations, fmt.Sprintf("evidence = %v, want exactly %v", paths, zkpFreezeEvidencePaths))
	}
	return violations
}

// zkpProtocolFreezeViolations fail-closed pins every frozen value in the
// manifest and binds it to the actual SHA-256 digest of the source circuit
// specification. Any drift returns a non-empty violation list.
func zkpProtocolFreezeViolations(m zkpProtocolFreezeManifest, sourceSpecDigest string) []string {
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
	wantFalse := func(field string, got bool) {
		if got {
			violations = append(violations, field+" must be false")
		}
	}

	wantString("schema", m.Schema, zkpProtocolFreezeSchema)
	wantString("status", m.Status, zkpFreezeStatus)
	wantString("classification", m.Classification, zkpFreezeClassification)
	wantFalse("production_ready", m.ProductionReady)
	wantFalse("submission_allowed", m.SubmissionAllowed)
	wantString("rollout_item", m.RolloutItem, zkpFreezeRolloutItem)

	wantString("source_spec.path", m.SourceSpec.Path, zkpFreezeSourcePath)
	wantString("source_spec.schema", m.SourceSpec.Schema, zkpFreezeSourceSchema)
	wantString("source_spec.sha256", m.SourceSpec.SHA256, zkpFreezeSourceSHA256)
	wantString("source_spec.artifact_classification", m.SourceSpec.ArtifactClassification, zkpFreezeArtifactClass)
	wantFalse("source_spec.artifact_production_allowed", m.SourceSpec.ArtifactProductionAllowed)
	if sourceSpecDigest != m.SourceSpec.SHA256 {
		violations = append(violations, fmt.Sprintf("source spec digest = %q, manifest pins %q", sourceSpecDigest, m.SourceSpec.SHA256))
	}

	wantString("profiles.protocol", m.Profiles.Protocol, zkpFreezeProfileProtocol)
	wantString("profiles.circuit", m.Profiles.Circuit, MembershipCircuitID)
	wantInt("profiles.consensus_version", m.Profiles.ConsensusVersion, zkpFreezeConsensusVersion)
	wantString("profiles.public_inputs", m.Profiles.PublicInputs, zkpFreezeProfilePublicIn)
	wantString("profiles.field_encoding", m.Profiles.FieldEncoding, zkpFreezeProfileFieldEnc)
	wantString("profiles.nullifier_scope", m.Profiles.NullifierScope, zkpFreezeNullifierScope)
	wantString("profiles.signal", m.Profiles.Signal, zkpFreezeSignalProfile)
	wantString("profiles.reward_recipient", m.Profiles.RewardRecipient, zkpFreezeProfileRecipient)

	wantStrings("semantics.public_input_order", m.Semantics.PublicInputOrder, zkpFreezePublicInputOrder)
	wantString("semantics.field_element_encoding", m.Semantics.FieldElementEncoding, zkpFreezeFieldElementCanon)
	wantStrings("semantics.nullifier_bound_fields", m.Semantics.NullifierBoundFields, zkpFreezeNullifierBoundFields)
	wantStrings("semantics.nullifier_excluded_fields", m.Semantics.NullifierExcludedFields, zkpFreezeNullifierExcludedFields)
	wantStrings("semantics.signal_bound_fields", m.Semantics.SignalBoundFields, zkpFreezeSignalBoundFields)

	wantString("semantics.legacy_context.profile", m.Semantics.LegacyContext.Profile, zkpFreezeNullifierScope)
	wantString("semantics.legacy_context.use", m.Semantics.LegacyContext.Use, zkpFreezeLegacyUse)
	wantFalse("semantics.legacy_context.production_signal", m.Semantics.LegacyContext.ProductionSignal)
	wantString("semantics.candidate_context.profile", m.Semantics.CandidateContext.Profile, zkpFreezeSignalProfile)
	wantString("semantics.candidate_context.use", m.Semantics.CandidateContext.Use, zkpFreezeCandidateUse)
	if !m.Semantics.CandidateContext.ProductionSignal {
		violations = append(violations, "semantics.candidate_context.production_signal must be true")
	}
	if m.Semantics.LegacyContext.Profile != m.Profiles.NullifierScope {
		violations = append(violations, "legacy context profile must equal profiles.nullifier_scope")
	}
	if m.Semantics.CandidateContext.Profile != m.Profiles.Signal {
		violations = append(violations, "candidate context profile must equal profiles.signal")
	}

	// The nullifier scope must bind exactly the signal fields minus the
	// excluded rating and reward recipient, so the one-vote rule stays
	// recipient- and rating-independent.
	if len(m.Semantics.NullifierBoundFields)+len(m.Semantics.NullifierExcludedFields) != len(m.Semantics.SignalBoundFields) {
		violations = append(violations, "nullifier bound+excluded fields must partition the signal bound fields")
	}
	bound := make(map[string]struct{}, len(m.Semantics.NullifierBoundFields))
	for _, field := range m.Semantics.NullifierBoundFields {
		bound[field] = struct{}{}
	}
	signalBound := make(map[string]struct{}, len(m.Semantics.SignalBoundFields))
	for _, field := range m.Semantics.SignalBoundFields {
		signalBound[field] = struct{}{}
	}
	for _, field := range m.Semantics.NullifierExcludedFields {
		if _, isBound := bound[field]; isBound {
			violations = append(violations, fmt.Sprintf("nullifier excluded field %q is also bound", field))
		}
		if _, isSignaled := signalBound[field]; !isSignaled {
			violations = append(violations, fmt.Sprintf("nullifier excluded field %q is not signal-bound", field))
		}
	}

	wantFalse("change_control.in_place_semantic_change_allowed", m.ChangeControl.InPlaceSemanticChangeAllowed)
	wantStrings("change_control.circuit_id_bump_required_for", m.ChangeControl.CircuitIDBumpRequiredFor, zkpFreezeCircuitBumpRequiredFor)
	wantStrings("change_control.protocol_profile_bump_required_for", m.ChangeControl.ProtocolProfileBumpRequiredFor, zkpFreezeProtocolBumpRequiredFor)
	wantStrings("change_control.new_nullifier_keyspace_required_for", m.ChangeControl.NewNullifierKeyspaceRequiredFor, zkpFreezeNullifierKeyspaceRequiredFor)
	wantString("change_control.activation", m.ChangeControl.Activation, zkpFreezeActivation)
	wantString("change_control.review_forced_change", m.ChangeControl.ReviewForcedChange, zkpFreezeReviewForced)

	wantStrings("open_gates", m.OpenGates, zkpFreezeOpenGates)
	violations = append(violations, zkpFreezeEvidenceListViolations(m.Evidence)...)
	return violations
}

// loadZKPProtocolFreeze strict-decodes the Sol-owned freeze manifest with the
// existing duplicate/unknown/trailing rejection helper, verifies every pinned
// value, and digest-binds the source circuit specification without changing
// it. It also returns the fully cross-checked circuit spec loaded by the
// existing GH-203 helper.
func loadZKPProtocolFreeze(t *testing.T) (zkpProtocolFreezeManifest, []byte, zkpCircuitSpec) {
	t.Helper()
	raw, err := os.ReadFile(zkpProtocolFreezePath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest zkpProtocolFreezeManifest
	if err := strictJSON(raw, &manifest); err != nil {
		t.Fatalf("strict protocol freeze decode: %v", err)
	}
	sourceBytes, err := os.ReadFile(filepath.Join("../..", zkpFreezeSourcePath))
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(sourceBytes)
	spec := loadZKPCircuitSpec(t)
	if violations := zkpProtocolFreezeViolations(manifest, hex.EncodeToString(digest[:])); len(violations) != 0 {
		t.Fatalf("zkp protocol freeze violations:\n- %s", strings.Join(violations, "\n- "))
	}
	return manifest, sourceBytes, spec
}

func TestZKPProtocolFreezeManifestStrictContract(t *testing.T) {
	loadZKPProtocolFreeze(t)
}

func TestZKPProtocolFreezeDigestBindsCircuitSpec(t *testing.T) {
	manifest, sourceBytes, spec := loadZKPProtocolFreeze(t)

	// Decode the exact digest-bound bytes again and cross-check identity,
	// schema, public-input order and artifact classification against both the
	// manifest and the live Go constants. The source spec is never modified.
	var decoded zkpCircuitSpec
	if err := strictJSON(sourceBytes, &decoded); err != nil {
		t.Fatalf("strict source spec decode: %v", err)
	}
	if decoded.Schema != manifest.SourceSpec.Schema || decoded.Schema != spec.Schema {
		t.Fatalf("source spec schema = %q, manifest pins %q", decoded.Schema, manifest.SourceSpec.Schema)
	}
	if decoded.CircuitID != manifest.Profiles.Circuit || decoded.CircuitID != MembershipCircuitID || spec.CircuitID != MembershipCircuitID {
		t.Fatalf("source spec circuit id %q drifted from manifest %q / Go %q",
			decoded.CircuitID, manifest.Profiles.Circuit, MembershipCircuitID)
	}
	if fmt.Sprint(decoded.PublicInputOrder) != fmt.Sprint(manifest.Semantics.PublicInputOrder) ||
		len(decoded.PublicInputOrder) != membershipPublicWitnessCount {
		t.Fatalf("source spec public input order = %v, manifest pins %v", decoded.PublicInputOrder, manifest.Semantics.PublicInputOrder)
	}
	if decoded.Classification != manifest.SourceSpec.ArtifactClassification {
		t.Fatalf("source spec classification = %q, manifest pins %q", decoded.Classification, manifest.SourceSpec.ArtifactClassification)
	}
	if decoded.ProductionAllowed || manifest.SourceSpec.ArtifactProductionAllowed {
		t.Fatal("test-only circuit artifacts must stay production-forbidden")
	}

	after, err := os.ReadFile(filepath.Join("../..", zkpFreezeSourcePath))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, sourceBytes) {
		t.Fatal("protocol freeze verification mutated the source circuit spec")
	}
}

func TestZKPProtocolFreezeMatchesGoConsensusAndSemantics(t *testing.T) {
	manifest, _, spec := loadZKPProtocolFreeze(t)

	if manifest.Profiles.Circuit != MembershipCircuitID {
		t.Fatalf("frozen circuit profile %q drifted from Go circuit id %q", manifest.Profiles.Circuit, MembershipCircuitID)
	}
	if spec.Merkle.Depth != MerkleTreeDepth || MerkleTreeDepth != 20 {
		t.Fatalf("frozen merkle depth drifted from Go constant %d", MerkleTreeDepth)
	}
	if got := (AppModule{}).ConsensusVersion(); got != zkpFreezeConsensusVersion || manifest.Profiles.ConsensusVersion != int(got) {
		t.Fatalf("frozen consensus version %d drifted from AppModule ConsensusVersion %d", manifest.Profiles.ConsensusVersion, got)
	}
	if len(manifest.Semantics.PublicInputOrder) != membershipPublicWitnessCount {
		t.Fatalf("frozen public input count %d drifted from circuit witness count %d",
			len(manifest.Semantics.PublicInputOrder), membershipPublicWitnessCount)
	}

	// Digest-bind the declared v1 nullifier/v2 signal profiles to the Go
	// encoders: the preimages are rebuilt from the manifest-declared domain
	// separators, bound fields and their declared order, then hashed with the
	// spec formula. Drift in either the manifest or the Go encoding fails.
	vector := spec.VoteContextV2.Vector
	fieldString := func(field string) (string, bool) {
		switch field {
		case "chain_id":
			return vector.ChainID, true
		case "domain_name":
			return vector.DomainName, true
		case "issue_name":
			return vector.IssueName, true
		case "suggestion_name":
			return vector.SuggestionName, true
		case "reward_recipient":
			return vector.RewardRecipient, true
		}
		return "", false
	}
	reduceToField := func(data []byte) []byte {
		digest := sha256.Sum256(data)
		reduced := new(big.Int).Mod(new(big.Int).SetBytes(digest[:]), ecc.BN254.ScalarField())
		out := make([]byte, 32)
		reduced.FillBytes(out)
		return out
	}
	writeStringField := func(preimage *bytes.Buffer, field string) {
		value, ok := fieldString(field)
		if !ok {
			t.Fatalf("bound field %q is not a string vote-context field", field)
		}
		_ = binary.Write(preimage, binary.BigEndian, uint32(len(value)))
		preimage.WriteString(value)
	}

	var scopePreimage bytes.Buffer
	scopePreimage.WriteString(manifest.Profiles.NullifierScope)
	for _, field := range manifest.Semantics.NullifierBoundFields {
		if field == "rating" {
			t.Fatal("nullifier scope must stay rating-independent")
		}
		writeStringField(&scopePreimage, field)
	}
	scope := ComputeVoteNullifierScope(vector.ChainID, vector.DomainName, vector.IssueName, vector.SuggestionName)
	if !bytes.Equal(reduceToField(scopePreimage.Bytes()), scope) {
		t.Fatal("frozen v1 nullifier-scope semantics drifted from the Go encoder")
	}

	var signalPreimage bytes.Buffer
	signalPreimage.WriteString(manifest.Profiles.Signal)
	for _, field := range manifest.Semantics.SignalBoundFields {
		if field == "rating" {
			_ = binary.Write(&signalPreimage, binary.BigEndian, int64(vector.Rating))
			continue
		}
		writeStringField(&signalPreimage, field)
	}
	signal := ComputeVoteSignalV2(vector.ChainID, vector.DomainName, vector.IssueName, vector.SuggestionName, vector.Rating, vector.RewardRecipient)
	if !bytes.Equal(reduceToField(signalPreimage.Bytes()), signal) {
		t.Fatal("frozen v2 signal semantics drifted from the Go encoder")
	}
	if bytes.Equal(scope, signal) {
		t.Fatal("nullifier scope and vote signal must never collide")
	}
}

func TestZKPProtocolFreezeEvidencePathsAreSafeAndComplete(t *testing.T) {
	manifest, _, _ := loadZKPProtocolFreeze(t)

	for _, path := range manifest.Evidence {
		current := filepath.Join("..", "..")
		segments := strings.Split(path, "/")
		for i, segment := range segments {
			current = filepath.Join(current, segment)
			info, err := os.Lstat(current)
			if err != nil {
				t.Fatalf("evidence path %q: %v", path, err)
			}
			if info.Mode()&os.ModeSymlink != 0 {
				t.Fatalf("evidence path %q traverses symlink component %q", path, current)
			}
			if i == len(segments)-1 && !info.Mode().IsRegular() {
				t.Fatalf("evidence path %q is not a regular file", path)
			}
		}
	}

	t.Run("non-regular evidence rejected", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "target.json")
		if err := os.WriteFile(target, []byte("{}"), 0o600); err != nil {
			t.Fatal(err)
		}
		link := filepath.Join(dir, "link.json")
		if err := os.Symlink(target, link); err != nil {
			t.Fatal(err)
		}
		if reason := zkpFreezeEvidenceFileReason(link); reason == "" {
			t.Fatal("symlink evidence path accepted")
		}
		if reason := zkpFreezeEvidenceFileReason(dir); reason == "" {
			t.Fatal("directory evidence path accepted")
		}
		if reason := zkpFreezeEvidenceFileReason(filepath.Join(dir, "missing.json")); reason == "" {
			t.Fatal("missing evidence path accepted")
		}
		if reason := zkpFreezeEvidenceFileReason(target); reason != "" {
			t.Fatalf("regular evidence file rejected: %s", reason)
		}
	})
}

func TestZKPProtocolFreezeClientSubmissionStaysDisabled(t *testing.T) {
	manifest, _, _ := loadZKPProtocolFreeze(t)
	if manifest.ProductionReady || manifest.SubmissionAllowed {
		t.Fatal("protocol freeze must keep production and submission disabled")
	}
	guard, err := os.ReadFile(zkpFreezeClientGuardPath)
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`get isSubmittable\(\): boolean \{\s*return false;\s*\}`).Match(guard) {
		t.Fatal("client isSubmittable is not an exact hard-false getter")
	}
}

func TestZKPProtocolFreezeRejectsDrift(t *testing.T) {
	manifest, sourceBytes, _ := loadZKPProtocolFreeze(t)
	digest := sha256.Sum256(sourceBytes)
	validDigest := hex.EncodeToString(digest[:])

	clone := func() zkpProtocolFreezeManifest {
		candidate := manifest
		candidate.Semantics.PublicInputOrder = append([]string(nil), manifest.Semantics.PublicInputOrder...)
		candidate.Semantics.NullifierBoundFields = append([]string(nil), manifest.Semantics.NullifierBoundFields...)
		candidate.Semantics.NullifierExcludedFields = append([]string(nil), manifest.Semantics.NullifierExcludedFields...)
		candidate.Semantics.SignalBoundFields = append([]string(nil), manifest.Semantics.SignalBoundFields...)
		candidate.ChangeControl.CircuitIDBumpRequiredFor = append([]string(nil), manifest.ChangeControl.CircuitIDBumpRequiredFor...)
		candidate.ChangeControl.ProtocolProfileBumpRequiredFor = append([]string(nil), manifest.ChangeControl.ProtocolProfileBumpRequiredFor...)
		candidate.ChangeControl.NewNullifierKeyspaceRequiredFor = append([]string(nil), manifest.ChangeControl.NewNullifierKeyspaceRequiredFor...)
		candidate.OpenGates = append([]string(nil), manifest.OpenGates...)
		candidate.Evidence = append([]string(nil), manifest.Evidence...)
		return candidate
	}

	cases := map[string]func(*zkpProtocolFreezeManifest){
		"wrong schema":                     func(m *zkpProtocolFreezeManifest) { m.Schema = "truerepublic/zkp-protocol-freeze/v2" },
		"wrong status":                     func(m *zkpProtocolFreezeManifest) { m.Status = "draft" },
		"weakened classification":          func(m *zkpProtocolFreezeManifest) { m.Classification = "PRODUCTION READY" },
		"production ready flip":            func(m *zkpProtocolFreezeManifest) { m.ProductionReady = true },
		"submission allowed flip":          func(m *zkpProtocolFreezeManifest) { m.SubmissionAllowed = true },
		"wrong rollout item":               func(m *zkpProtocolFreezeManifest) { m.RolloutItem = "phase-2-done" },
		"wrong source path":                func(m *zkpProtocolFreezeManifest) { m.SourceSpec.Path = "configs/security/zkp-protocol-freeze.json" },
		"wrong source schema":              func(m *zkpProtocolFreezeManifest) { m.SourceSpec.Schema = "truerepublic/zkp-circuit/v1" },
		"wrong source sha256":              func(m *zkpProtocolFreezeManifest) { m.SourceSpec.SHA256 = strings.Repeat("00", 32) },
		"weakened artifact classification": func(m *zkpProtocolFreezeManifest) { m.SourceSpec.ArtifactClassification = "PRODUCTION" },
		"artifact production allowed":      func(m *zkpProtocolFreezeManifest) { m.SourceSpec.ArtifactProductionAllowed = true },
		"wrong protocol profile":           func(m *zkpProtocolFreezeManifest) { m.Profiles.Protocol = "truerepublic/anonymous-rating/v1" },
		"wrong circuit profile": func(m *zkpProtocolFreezeManifest) {
			m.Profiles.Circuit = "truerepublic/membership-vote/v3-bn254-mimc-depth20"
		},
		"consensus version downgrade":   func(m *zkpProtocolFreezeManifest) { m.Profiles.ConsensusVersion = 1 },
		"consensus version skip":        func(m *zkpProtocolFreezeManifest) { m.Profiles.ConsensusVersion = 3 },
		"wrong public inputs profile":   func(m *zkpProtocolFreezeManifest) { m.Profiles.PublicInputs = "truerepublic/zkp-public-inputs/v2" },
		"wrong field encoding profile":  func(m *zkpProtocolFreezeManifest) { m.Profiles.FieldEncoding = "truerepublic/bn254-field-le32/v1" },
		"nullifier scope swapped to v2": func(m *zkpProtocolFreezeManifest) { m.Profiles.NullifierScope = zkpFreezeSignalProfile },
		"signal swapped to v1":          func(m *zkpProtocolFreezeManifest) { m.Profiles.Signal = zkpFreezeNullifierScope },
		"wrong reward recipient profile": func(m *zkpProtocolFreezeManifest) {
			m.Profiles.RewardRecipient = "truerepublic/bech32-any-recipient/v1"
		},
		"reordered public inputs": func(m *zkpProtocolFreezeManifest) {
			m.Semantics.PublicInputOrder[0], m.Semantics.PublicInputOrder[1] = m.Semantics.PublicInputOrder[1], m.Semantics.PublicInputOrder[0]
		},
		"wrong field element encoding": func(m *zkpProtocolFreezeManifest) { m.Semantics.FieldElementEncoding = "32-byte little-endian" },
		"reordered nullifier bound fields": func(m *zkpProtocolFreezeManifest) {
			m.Semantics.NullifierBoundFields[0], m.Semantics.NullifierBoundFields[1] = m.Semantics.NullifierBoundFields[1], m.Semantics.NullifierBoundFields[0]
		},
		"rating added to nullifier scope": func(m *zkpProtocolFreezeManifest) {
			m.Semantics.NullifierBoundFields = append(m.Semantics.NullifierBoundFields, "rating")
		},
		"excluded reward recipient dropped": func(m *zkpProtocolFreezeManifest) {
			m.Semantics.NullifierExcludedFields = m.Semantics.NullifierExcludedFields[:1]
		},
		"bound field excluded": func(m *zkpProtocolFreezeManifest) {
			m.Semantics.NullifierExcludedFields = append(m.Semantics.NullifierExcludedFields, "chain_id")
		},
		"signal recipient binding dropped": func(m *zkpProtocolFreezeManifest) {
			m.Semantics.SignalBoundFields = m.Semantics.SignalBoundFields[:5]
		},
		"reordered signal bound fields": func(m *zkpProtocolFreezeManifest) {
			m.Semantics.SignalBoundFields[4], m.Semantics.SignalBoundFields[5] = m.Semantics.SignalBoundFields[5], m.Semantics.SignalBoundFields[4]
		},
		"legacy profile drift":          func(m *zkpProtocolFreezeManifest) { m.Semantics.LegacyContext.Profile = zkpFreezeSignalProfile },
		"legacy promoted to production": func(m *zkpProtocolFreezeManifest) { m.Semantics.LegacyContext.ProductionSignal = true },
		"candidate profile drift":       func(m *zkpProtocolFreezeManifest) { m.Semantics.CandidateContext.Profile = zkpFreezeNullifierScope },
		"candidate demoted from production": func(m *zkpProtocolFreezeManifest) {
			m.Semantics.CandidateContext.ProductionSignal = false
		},
		"in-place semantic change allowed": func(m *zkpProtocolFreezeManifest) {
			m.ChangeControl.InPlaceSemanticChangeAllowed = true
		},
		"circuit bump list shrunk": func(m *zkpProtocolFreezeManifest) {
			m.ChangeControl.CircuitIDBumpRequiredFor = append([]string(nil), m.ChangeControl.CircuitIDBumpRequiredFor[:6]...)
		},
		"circuit bump list reordered": func(m *zkpProtocolFreezeManifest) {
			m.ChangeControl.CircuitIDBumpRequiredFor[0], m.ChangeControl.CircuitIDBumpRequiredFor[1] =
				m.ChangeControl.CircuitIDBumpRequiredFor[1], m.ChangeControl.CircuitIDBumpRequiredFor[0]
		},
		"protocol bump list shrunk": func(m *zkpProtocolFreezeManifest) {
			m.ChangeControl.ProtocolProfileBumpRequiredFor = append([]string(nil), m.ChangeControl.ProtocolProfileBumpRequiredFor[:4]...)
		},
		"nullifier keyspace list shrunk": func(m *zkpProtocolFreezeManifest) {
			m.ChangeControl.NewNullifierKeyspaceRequiredFor = append([]string(nil), m.ChangeControl.NewNullifierKeyspaceRequiredFor[:2]...)
		},
		"weakened activation boundary": func(m *zkpProtocolFreezeManifest) { m.ChangeControl.Activation = "any coordinated restart" },
		"weakened review rule": func(m *zkpProtocolFreezeManifest) {
			m.ChangeControl.ReviewForcedChange = "edit the frozen profile in place"
		},
		"open gates emptied":           func(m *zkpProtocolFreezeManifest) { m.OpenGates = nil },
		"open gate dropped":            func(m *zkpProtocolFreezeManifest) { m.OpenGates = m.OpenGates[:4] },
		"duplicate evidence":           func(m *zkpProtocolFreezeManifest) { m.Evidence = append(m.Evidence, m.Evidence[0]) },
		"missing evidence":             func(m *zkpProtocolFreezeManifest) { m.Evidence = m.Evidence[:len(m.Evidence)-1] },
		"extra evidence":               func(m *zkpProtocolFreezeManifest) { m.Evidence = append(m.Evidence, "go.mod") },
		"unsafe evidence traversal":    func(m *zkpProtocolFreezeManifest) { m.Evidence[0] = "../secrets/keys.json" },
		"absolute evidence path":       func(m *zkpProtocolFreezeManifest) { m.Evidence[0] = "/etc/passwd" },
		"non-normalized evidence path": func(m *zkpProtocolFreezeManifest) { m.Evidence[0] = "./configs/security/zkp-circuit.json" },
		"backslash evidence path":      func(m *zkpProtocolFreezeManifest) { m.Evidence[0] = `configs\security\zkp-circuit.json` },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			candidate := clone()
			mutate(&candidate)
			if violations := zkpProtocolFreezeViolations(candidate, validDigest); len(violations) == 0 {
				t.Fatal("drifted protocol freeze manifest accepted")
			}
		})
	}

	t.Run("source digest drift", func(t *testing.T) {
		if violations := zkpProtocolFreezeViolations(manifest, strings.Repeat("00", 32)); len(violations) == 0 {
			t.Fatal("source spec digest drift accepted")
		}
	})
}

func TestZKPProtocolFreezeRejectsMalformedJSON(t *testing.T) {
	raw, err := os.ReadFile(zkpProtocolFreezePath)
	if err != nil {
		t.Fatal(err)
	}

	cases := map[string][]byte{
		"unknown top-level field": bytes.Replace(raw, []byte("\n}"), []byte(",\n  \"unexpected\": true\n}"), 1),
		"unknown nested field": bytes.Replace(raw, []byte("\"source_spec\": {"),
			[]byte("\"source_spec\": {\n    \"unexpected\": true,"), 1),
		"duplicate top-level key": bytes.Replace(raw, []byte("{\n"), []byte("{\n  \"schema\": \"duplicate\",\n"), 1),
		"duplicate nested key": bytes.Replace(raw,
			[]byte("\"path\": \"configs/security/zkp-circuit.json\""),
			[]byte("\"path\": \"configs/security/zkp-circuit.json\",\n    \"path\": \"other.json\""), 1),
		"trailing JSON value": append(append([]byte(nil), raw...), []byte(" {}")...),
	}
	for name, candidate := range cases {
		t.Run(name, func(t *testing.T) {
			var decoded zkpProtocolFreezeManifest
			if err := strictJSON(candidate, &decoded); err == nil {
				t.Fatal("malformed protocol freeze manifest accepted")
			}
		})
	}
}
