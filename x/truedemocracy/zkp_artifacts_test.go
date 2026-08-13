package truedemocracy

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	zkpFixtureDir    = "testdata/zkp"
	zkpFixtureSchema = "truerepublic/zkp-test-fixture/v1"
)

type zkpArtifactRecord struct {
	Path   string `json:"path"`
	Size   int    `json:"size_bytes"`
	SHA256 string `json:"sha256"`
}

type zkpFixtureManifest struct {
	Schema            string              `json:"schema"`
	CircuitID         string              `json:"circuit_id"`
	Curve             string              `json:"curve"`
	MerkleDepth       int                 `json:"merkle_depth"`
	PublicInputOrder  []string            `json:"public_input_order"`
	Gnark             string              `json:"gnark"`
	GnarkCrypto       string              `json:"gnark_crypto"`
	Classification    string              `json:"classification"`
	ProductionAllowed bool                `json:"production_allowed"`
	Artifacts         []zkpArtifactRecord `json:"artifacts"`
}

type zkpGoldenVector struct {
	Schema               string   `json:"schema"`
	CircuitID            string   `json:"circuit_id"`
	ChainID              string   `json:"chain_id"`
	DomainName           string   `json:"domain_name"`
	IssueName            string   `json:"issue_name"`
	SuggestionName       string   `json:"suggestion_name"`
	Rating               int      `json:"rating"`
	SyntheticWitnessHex  string   `json:"synthetic_witness_hex"`
	CommitmentHex        string   `json:"commitment_hex"`
	MerkleRootHex        string   `json:"merkle_root_hex"`
	SiblingsHex          []string `json:"siblings_hex"`
	PathIndices          []int    `json:"path_indices"`
	ExternalNullifierHex string   `json:"external_nullifier_hex"`
	SignalHashHex        string   `json:"signal_hash_hex"`
	NullifierHashHex     string   `json:"nullifier_hash_hex"`
	ProofHex             string   `json:"proof_hex"`
	ProofSHA256          string   `json:"proof_sha256"`
	SyntheticAndTestOnly bool     `json:"synthetic_and_test_only"`
}

func strictJSON(data []byte, target any) error {
	if err := rejectDuplicateJSONKeys(json.NewDecoder(bytes.NewReader(data))); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("trailing JSON value")
	}
	return nil
}

func rejectDuplicateJSONKeys(decoder *json.Decoder) error {
	if err := rejectDuplicateJSONValue(decoder); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return fmt.Errorf("trailing JSON value")
	}
	return nil
}

func rejectDuplicateJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := make(map[string]struct{})
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return fmt.Errorf("JSON object key is not a string")
			}
			if _, exists := seen[key]; exists {
				return fmt.Errorf("duplicate JSON key %q", key)
			}
			seen[key] = struct{}{}
			if err := rejectDuplicateJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	case '[':
		for decoder.More() {
			if err := rejectDuplicateJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delim)
	}
}

func artifactRecord(path string, data []byte) zkpArtifactRecord {
	digest := sha256.Sum256(data)
	return zkpArtifactRecord{Path: path, Size: len(data), SHA256: hex.EncodeToString(digest[:])}
}

func verifyArtifact(record zkpArtifactRecord, data []byte) error {
	if got := artifactRecord(record.Path, data); got != record {
		return fmt.Errorf("artifact mismatch for %s: got %+v want %+v", record.Path, got, record)
	}
	return nil
}

func readVerifiedZKPFixture(t *testing.T) (zkpFixtureManifest, zkpGoldenVector, groth16.VerifyingKey, []byte) {
	t.Helper()
	manifestBytes, err := os.ReadFile(filepath.Join(zkpFixtureDir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest zkpFixtureManifest
	if err := strictJSON(manifestBytes, &manifest); err != nil {
		t.Fatalf("strict manifest decode: %v", err)
	}
	if manifest.Schema != zkpFixtureSchema || manifest.CircuitID != MembershipCircuitID ||
		manifest.Curve != "BN254" || manifest.MerkleDepth != MerkleTreeDepth ||
		manifest.ProductionAllowed || manifest.Classification != "TEST-ONLY SINGLE-PARTY TOXIC WASTE" {
		t.Fatalf("unsafe or incompatible fixture manifest: %+v", manifest)
	}
	wantInputs := []string{"merkle_root", "nullifier_hash", "external_nullifier", "signal_hash"}
	if fmt.Sprint(manifest.PublicInputOrder) != fmt.Sprint(wantInputs) {
		t.Fatalf("public input order = %v", manifest.PublicInputOrder)
	}
	artifactFiles := []string{"golden_vector.json", "membership_v2.cs", "membership_v2.pk", "membership_v2.vk"}
	wantFiles := []string{"golden_vector.json", "manifest.json", "membership_v2.cs", "membership_v2.pk", "membership_v2.vk"}
	entries, err := os.ReadDir(zkpFixtureDir)
	if err != nil {
		t.Fatal(err)
	}
	var gotFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			t.Fatalf("unexpected fixture subdirectory %q", entry.Name())
		}
		gotFiles = append(gotFiles, entry.Name())
	}
	sort.Strings(gotFiles)
	if fmt.Sprint(gotFiles) != fmt.Sprint(wantFiles) {
		t.Fatalf("fixture file set = %v, want %v", gotFiles, wantFiles)
	}
	if len(manifest.Artifacts) != 4 {
		t.Fatalf("artifact count = %d", len(manifest.Artifacts))
	}
	artifactData := make(map[string][]byte, 4)
	for _, record := range manifest.Artifacts {
		if filepath.Base(record.Path) != record.Path || record.Path == "manifest.json" {
			t.Fatalf("unsafe artifact path %q", record.Path)
		}
		data, err := os.ReadFile(filepath.Join(zkpFixtureDir, record.Path))
		if err != nil {
			t.Fatal(err)
		}
		if err := verifyArtifact(record, data); err != nil {
			t.Fatal(err)
		}
		if _, duplicate := artifactData[record.Path]; duplicate {
			t.Fatalf("duplicate artifact %q", record.Path)
		}
		artifactData[record.Path] = data
	}
	for _, name := range artifactFiles {
		if artifactData[name] == nil {
			t.Fatalf("manifest omits %s", name)
		}
	}

	cs := groth16.NewCS(ecc.BN254)
	reader := bytes.NewReader(artifactData["membership_v2.cs"])
	if _, err := cs.ReadFrom(reader); err != nil || reader.Len() != 0 {
		t.Fatalf("constraint system decode: err=%v trailing=%d", err, reader.Len())
	}
	pk := groth16.NewProvingKey(ecc.BN254)
	reader = bytes.NewReader(artifactData["membership_v2.pk"])
	if _, err := pk.ReadFrom(reader); err != nil || reader.Len() != 0 {
		t.Fatalf("proving key decode: err=%v trailing=%d", err, reader.Len())
	}
	vkBytes := artifactData["membership_v2.vk"]
	vk, err := ValidateMembershipVerifyingKey(vkBytes, manifest.CircuitID, VerifyingKeyFingerprint(vkBytes))
	if err != nil {
		t.Fatalf("verifying key validation: %v", err)
	}
	var vector zkpGoldenVector
	if err := strictJSON(artifactData["golden_vector.json"], &vector); err != nil {
		t.Fatalf("strict vector decode: %v", err)
	}
	if vector.Schema != "truerepublic/zkp-golden-vector/v1" || vector.CircuitID != MembershipCircuitID ||
		!vector.SyntheticAndTestOnly || len(vector.SiblingsHex) != MerkleTreeDepth || len(vector.PathIndices) != MerkleTreeDepth {
		t.Fatalf("invalid golden-vector metadata")
	}
	proof, err := hex.DecodeString(vector.ProofHex)
	if err != nil {
		t.Fatal(err)
	}
	proofDigest := sha256.Sum256(proof)
	if hex.EncodeToString(proofDigest[:]) != vector.ProofSHA256 {
		t.Fatal("golden proof checksum mismatch")
	}
	return manifest, vector, vk, proof
}

func fixtureHex(t *testing.T, label, value string) []byte {
	t.Helper()
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != 32 {
		t.Fatalf("%s must be canonical 32-byte hex: %v", label, err)
	}
	return decoded
}

func TestZKPFixtureArtifactsAndGoldenProof(t *testing.T) {
	_, vector, vk, proof := readVerifiedZKPFixture(t)
	root := fixtureHex(t, "merkle root", vector.MerkleRootHex)
	nullifier := fixtureHex(t, "nullifier", vector.NullifierHashHex)
	external := fixtureHex(t, "external nullifier", vector.ExternalNullifierHex)
	signal := fixtureHex(t, "signal", vector.SignalHashHex)
	if err := VerifyMembershipProofForSignal(vk, proof, root, nullifier, external, signal); err != nil {
		t.Fatalf("golden proof rejected: %v", err)
	}

	t.Run("corrupted proof", func(t *testing.T) {
		candidate := append([]byte(nil), proof...)
		candidate[len(candidate)/2] ^= 1
		if err := VerifyMembershipProofForSignal(vk, candidate, root, nullifier, external, signal); err == nil {
			t.Fatal("corrupted proof accepted")
		}
	})
	t.Run("wrong root", func(t *testing.T) {
		candidate := append([]byte(nil), root...)
		candidate[31] ^= 1
		if err := VerifyMembershipProofForSignal(vk, proof, candidate, nullifier, external, signal); err == nil {
			t.Fatal("wrong root accepted")
		}
	})
	t.Run("wrong rating signal", func(t *testing.T) {
		candidate := ComputeVoteSignal(vector.ChainID, vector.DomainName, vector.IssueName, vector.SuggestionName, vector.Rating+1)
		if err := VerifyMembershipProofForSignal(vk, proof, root, nullifier, external, candidate); err == nil {
			t.Fatal("wrong rating signal accepted")
		}
	})
	t.Run("wrong chain scope", func(t *testing.T) {
		candidate := ComputeVoteNullifierScope(vector.ChainID+"-other", vector.DomainName, vector.IssueName, vector.SuggestionName)
		if err := VerifyMembershipProofForSignal(vk, proof, root, nullifier, candidate, signal); err == nil {
			t.Fatal("wrong chain scope accepted")
		}
	})
	t.Run("trailing proof bytes", func(t *testing.T) {
		if err := VerifyMembershipProofForSignal(vk, append(proof, 0), root, nullifier, external, signal); err == nil {
			t.Fatal("proof with trailing data accepted")
		}
	})
}

func TestZKPFixtureMetadataRejectsDriftAndMalformedKeys(t *testing.T) {
	manifestBytes, err := os.ReadFile(filepath.Join(zkpFixtureDir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest zkpFixtureManifest
	if err := strictJSON(manifestBytes, &manifest); err != nil {
		t.Fatal(err)
	}

	t.Run("unknown manifest field", func(t *testing.T) {
		candidate := bytes.Replace(manifestBytes, []byte("\n}"), []byte(",\n  \"unexpected\": true\n}"), 1)
		var decoded zkpFixtureManifest
		if err := strictJSON(candidate, &decoded); err == nil {
			t.Fatal("unknown manifest field accepted")
		}
	})
	t.Run("artifact checksum drift", func(t *testing.T) {
		record := manifest.Artifacts[0]
		data, err := os.ReadFile(filepath.Join(zkpFixtureDir, record.Path))
		if err != nil {
			t.Fatal(err)
		}
		candidate := append([]byte(nil), data...)
		candidate[0] ^= 1
		if err := verifyArtifact(record, candidate); err == nil {
			t.Fatal("loader checksum validation accepted corrupted artifact")
		}
	})
	t.Run("verifying key fingerprint mismatch", func(t *testing.T) {
		vkBytes, err := os.ReadFile(filepath.Join(zkpFixtureDir, "membership_v2.vk"))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ValidateMembershipVerifyingKey(vkBytes, MembershipCircuitID, string(bytes.Repeat([]byte{'0'}, 64))); err == nil {
			t.Fatal("wrong VK fingerprint accepted")
		}
	})
	t.Run("verifying key trailing bytes", func(t *testing.T) {
		vkBytes, err := os.ReadFile(filepath.Join(zkpFixtureDir, "membership_v2.vk"))
		if err != nil {
			t.Fatal(err)
		}
		candidate := append(vkBytes, 0)
		if _, err := ValidateMembershipVerifyingKey(candidate, MembershipCircuitID, VerifyingKeyFingerprint(candidate)); err == nil {
			t.Fatal("VK with trailing data accepted")
		}
	})
}

func TestZKPGoldenProofKeeperReplay(t *testing.T) {
	_, vector, _, proof := readVerifiedZKPFixture(t)
	k, ctx := setupKeeper(t)
	ctx = ctx.WithChainID(vector.ChainID)
	admin := sdk.AccAddress("fixture-admin")
	member := sdk.AccAddress("fixture-member")
	k.CreateDomain(ctx, vector.DomainName, admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 500_000)))
	if err := k.AddMember(ctx, vector.DomainName, member.String(), admin); err != nil {
		t.Fatal(err)
	}
	if err := k.RegisterIdentityCommitment(ctx, vector.DomainName, member.String(), vector.CommitmentHex); err != nil {
		t.Fatal(err)
	}
	domain, found := k.GetDomain(ctx, vector.DomainName)
	if !found || domain.MerkleRoot != vector.MerkleRootHex {
		t.Fatalf("fixture root mismatch: got %q want %q", domain.MerkleRoot, vector.MerkleRootHex)
	}
	addProposal(t, k, ctx, vector.DomainName, vector.IssueName, vector.SuggestionName)
	vkBytes, err := os.ReadFile(filepath.Join(zkpFixtureDir, "membership_v2.vk"))
	if err != nil {
		t.Fatal(err)
	}
	k.SetVerifyingKey(ctx, vkBytes)
	if _, err := k.RateProposalWithZKP(
		ctx, vector.DomainName, vector.IssueName, vector.SuggestionName, vector.Rating,
		hex.EncodeToString(proof), vector.NullifierHashHex, vector.MerkleRootHex,
	); err != nil {
		t.Fatalf("keeper rejected golden proof: %v", err)
	}
	if !k.IsNullifierUsed(ctx, vector.DomainName, vector.NullifierHashHex) {
		t.Fatal("keeper did not persist golden nullifier")
	}
}

func TestRegenerateZKPFixture(t *testing.T) {
	if os.Getenv("TRUEREPUBLIC_REGENERATE_ZKP_FIXTURE") != "1" {
		t.Skip("explicit fixture regeneration only")
	}
	keys, err := SetupMembershipCircuit()
	if err != nil {
		t.Fatal(err)
	}
	identity := hashToField([]byte("TrueRepublic/GH-198/synthetic-test-only-identity"))
	commitment, err := ComputeCommitment(identity)
	if err != nil {
		t.Fatal(err)
	}
	tree := NewMerkleTree(MerkleTreeDepth)
	if err := tree.BuildFromLeaves([][]byte{commitment}); err != nil {
		t.Fatal(err)
	}
	siblings, indices, err := tree.GenerateProof(0)
	if err != nil {
		t.Fatal(err)
	}
	vector := zkpGoldenVector{
		Schema: "truerepublic/zkp-golden-vector/v1", CircuitID: MembershipCircuitID,
		ChainID: "truerepublic-zkp-fixture-1", DomainName: "FixtureDomain", IssueName: "FixtureIssue",
		SuggestionName: "FixtureSuggestion", Rating: 3, SyntheticWitnessHex: hex.EncodeToString(identity),
		CommitmentHex: hex.EncodeToString(commitment), MerkleRootHex: hex.EncodeToString(tree.Root),
		PathIndices: indices, SyntheticAndTestOnly: true,
	}
	for _, sibling := range siblings {
		vector.SiblingsHex = append(vector.SiblingsHex, hex.EncodeToString(sibling))
	}
	external := ComputeVoteNullifierScope(vector.ChainID, vector.DomainName, vector.IssueName, vector.SuggestionName)
	signal := ComputeVoteSignal(vector.ChainID, vector.DomainName, vector.IssueName, vector.SuggestionName, vector.Rating)
	proof, nullifier, err := GenerateMembershipProofForSignal(keys, identity, tree.Root, siblings, indices, external, signal)
	if err != nil {
		t.Fatal(err)
	}
	vector.ExternalNullifierHex = hex.EncodeToString(external)
	vector.SignalHashHex = hex.EncodeToString(signal)
	vector.NullifierHashHex = hex.EncodeToString(nullifier)
	vector.ProofHex = hex.EncodeToString(proof)
	proofDigest := sha256.Sum256(proof)
	vector.ProofSHA256 = hex.EncodeToString(proofDigest[:])

	var cs, pk bytes.Buffer
	if _, err := keys.CS.WriteTo(&cs); err != nil {
		t.Fatal(err)
	}
	if _, err := keys.ProvingKey.WriteTo(&pk); err != nil {
		t.Fatal(err)
	}
	vk, err := SerializeVerifyingKey(keys.VerifyingKey)
	if err != nil {
		t.Fatal(err)
	}
	vectorBytes, err := json.MarshalIndent(vector, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	vectorBytes = append(vectorBytes, '\n')
	files := map[string][]byte{
		"golden_vector.json": vectorBytes, "membership_v2.cs": cs.Bytes(),
		"membership_v2.pk": pk.Bytes(), "membership_v2.vk": vk,
	}
	if err := os.MkdirAll(zkpFixtureDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := zkpFixtureManifest{
		Schema: zkpFixtureSchema, CircuitID: MembershipCircuitID, Curve: "BN254", MerkleDepth: MerkleTreeDepth,
		PublicInputOrder: []string{"merkle_root", "nullifier_hash", "external_nullifier", "signal_hash"},
		Gnark:            "v0.14.0", GnarkCrypto: "v0.19.2", Classification: "TEST-ONLY SINGLE-PARTY TOXIC WASTE",
		ProductionAllowed: false,
	}
	for _, name := range []string{"golden_vector.json", "membership_v2.cs", "membership_v2.pk", "membership_v2.vk"} {
		if err := os.WriteFile(filepath.Join(zkpFixtureDir, name), files[name], 0o644); err != nil {
			t.Fatal(err)
		}
		manifest.Artifacts = append(manifest.Artifacts, artifactRecord(name, files[name]))
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(zkpFixtureDir, "manifest.json"), append(manifestBytes, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}
