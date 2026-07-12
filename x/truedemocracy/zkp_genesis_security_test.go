package truedemocracy

import (
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func testGenesisVerifyingKey(t *testing.T) (string, string) {
	t.Helper()
	bytes, err := SerializeVerifyingKey(getTestZKPKeys(t).VerifyingKey)
	if err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(bytes), VerifyingKeyFingerprint(bytes)
}

func TestValidateGenesisPinsMembershipVerifyingKey(t *testing.T) {
	vkHex, fingerprint := testGenesisVerifyingKey(t)
	valid := validDemocracyGenesis()
	valid.ZKPCircuitID = MembershipCircuitID
	valid.VerifyingKeyHex = vkHex
	valid.VerifyingKeySHA256 = fingerprint
	if err := ValidateGenesisState(valid); err != nil {
		t.Fatalf("valid pinned VK rejected: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*GenesisState)
	}{
		{"missing circuit id", func(g *GenesisState) { g.ZKPCircuitID = "" }},
		{"wrong circuit id", func(g *GenesisState) { g.ZKPCircuitID = "truerepublic/membership-v1" }},
		{"missing fingerprint", func(g *GenesisState) { g.VerifyingKeySHA256 = "" }},
		{"wrong fingerprint", func(g *GenesisState) { g.VerifyingKeySHA256 = strings.Repeat("0", 64) }},
		{"uppercase encoding", func(g *GenesisState) { g.VerifyingKeyHex = strings.ToUpper(g.VerifyingKeyHex) }},
		{"trailing bytes", func(g *GenesisState) {
			g.VerifyingKeyHex += "00"
			decoded, _ := hex.DecodeString(g.VerifyingKeyHex)
			g.VerifyingKeySHA256 = VerifyingKeyFingerprint(decoded)
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			genesis := valid
			tc.mutate(&genesis)
			if err := ValidateGenesisState(genesis); err == nil {
				t.Fatal("invalid VK configuration accepted")
			}
		})
	}
}

func TestValidateGenesisRecomputesCanonicalIdentityTree(t *testing.T) {
	genesis := validDemocracyGenesis()
	commitments := make([]string, 2)
	leaves := make([][]byte, 2)
	for i := range commitments {
		secret := make([]byte, 32)
		secret[31] = byte(i + 1)
		commitment, err := ComputeCommitment(secret)
		if err != nil {
			t.Fatal(err)
		}
		leaves[i] = commitment
		commitments[i] = hex.EncodeToString(commitment)
	}
	tree := NewMerkleTree(MerkleTreeDepth)
	if err := tree.BuildFromLeaves(leaves); err != nil {
		t.Fatal(err)
	}
	genesis.Domains[0].IdentityCommits = commitments
	genesis.Domains[0].MerkleRoot = tree.GetRoot()
	if err := ValidateGenesisState(genesis); err != nil {
		t.Fatalf("valid identity tree rejected: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*GenesisState)
	}{
		{"short commitment", func(g *GenesisState) { g.Domains[0].IdentityCommits[0] = "01" }},
		{"uppercase commitment", func(g *GenesisState) {
			g.Domains[0].IdentityCommits[0] = strings.ToUpper(g.Domains[0].IdentityCommits[0])
		}},
		{"duplicate commitment", func(g *GenesisState) { g.Domains[0].IdentityCommits[1] = g.Domains[0].IdentityCommits[0] }},
		{"mismatched root", func(g *GenesisState) { g.Domains[0].MerkleRoot = strings.Repeat("0", 64) }},
		{"short history root", func(g *GenesisState) { g.Domains[0].MerkleRootHistory = []string{"01"} }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			candidate := genesis
			candidate.Domains = append([]Domain(nil), genesis.Domains...)
			candidate.Domains[0].IdentityCommits = append([]string(nil), genesis.Domains[0].IdentityCommits...)
			tc.mutate(&candidate)
			if err := ValidateGenesisState(candidate); err == nil {
				t.Fatal("malformed ZKP genesis accepted")
			}
		})
	}
}

func TestInitGenesisRestoresUsedZKPNullifiers(t *testing.T) {
	am, keeper, ctx := setupModuleForGenesis(t)
	admin := sdk.AccAddress("nullifier-admin")
	nullifier := strings.Repeat("0", 63) + "1"
	genesis := GenesisState{Domains: []Domain{{
		Name: "Restored", Admin: admin, Members: []string{admin.String()},
		Treasury: sdk.NewCoins(), PermissionReg: []string{},
		Issues: []Issue{{Name: "Issue", Suggestions: []Suggestion{{
			Name: "Suggestion", Creator: admin.String(),
			Ratings: []Rating{{NullifierHex: nullifier, Value: 3}},
		}}}},
	}}, UsedNullifiers: []NullifierRecord{{
		DomainName: "Restored", NullifierHash: nullifier, UsedAtHeight: 42,
	}}}
	bz, err := json.Marshal(genesis)
	if err != nil {
		t.Fatal(err)
	}
	am.InitGenesis(ctx, nil, bz)
	if !keeper.IsNullifierUsed(ctx, "Restored", nullifier) {
		t.Fatal("exported rating nullifier was not restored into the used set")
	}
	recordBz := ctx.KVStore(keeper.StoreKey).Get([]byte("nullifier:Restored:" + nullifier))
	var record NullifierRecord
	keeper.cdc.MustUnmarshalLengthPrefixed(recordBz, &record)
	if record.UsedAtHeight != 42 {
		t.Fatalf("used height = %d, want 42", record.UsedAtHeight)
	}

	exported := am.ExportGenesis(ctx, nil)
	var roundTrip GenesisState
	if err := json.Unmarshal(exported, &roundTrip); err != nil {
		t.Fatal(err)
	}
	if err := ValidateGenesisState(roundTrip); err != nil {
		t.Fatalf("round-trip genesis rejected: %v", err)
	}
	if len(roundTrip.UsedNullifiers) != 1 || roundTrip.UsedNullifiers[0].UsedAtHeight != 42 {
		t.Fatalf("active nullifier set not preserved: %+v", roundTrip.UsedNullifiers)
	}
}

func TestInitGenesisDoesNotResurrectPurgedRatingNullifier(t *testing.T) {
	am, keeper, ctx := setupModuleForGenesis(t)
	admin := sdk.AccAddress("purged-nullifier-admin")
	nullifier := strings.Repeat("0", 63) + "2"
	genesis := GenesisState{Domains: []Domain{{
		Name: "Purged", Admin: admin, Members: []string{admin.String()},
		Treasury: sdk.NewCoins(), PermissionReg: []string{},
		Issues: []Issue{{Name: "Issue", Suggestions: []Suggestion{{
			Name: "Suggestion", Creator: admin.String(),
			Ratings: []Rating{{NullifierHex: nullifier, Value: 3}},
		}}}},
	}}, UsedNullifiers: []NullifierRecord{}}
	bz, err := json.Marshal(genesis)
	if err != nil {
		t.Fatal(err)
	}
	am.InitGenesis(ctx, nil, bz)
	if keeper.IsNullifierUsed(ctx, "Purged", nullifier) {
		t.Fatal("historical rating resurrected a nullifier absent from the active genesis set")
	}
}

func TestValidateGenesisRejectsMalformedUsedNullifiers(t *testing.T) {
	genesis := validDemocracyGenesis()
	genesis.UsedNullifiers = []NullifierRecord{{
		DomainName: "Test", NullifierHash: strings.Repeat("0", 63) + "1", UsedAtHeight: 1,
	}}
	if err := ValidateGenesisState(genesis); err != nil {
		t.Fatalf("valid used nullifier rejected: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*GenesisState)
	}{
		{"missing domain", func(g *GenesisState) { g.UsedNullifiers[0].DomainName = "missing" }},
		{"short hash", func(g *GenesisState) { g.UsedNullifiers[0].NullifierHash = "01" }},
		{"negative height", func(g *GenesisState) { g.UsedNullifiers[0].UsedAtHeight = -1 }},
		{"duplicate", func(g *GenesisState) { g.UsedNullifiers = append(g.UsedNullifiers, g.UsedNullifiers[0]) }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			candidate := genesis
			candidate.UsedNullifiers = append([]NullifierRecord(nil), genesis.UsedNullifiers...)
			tc.mutate(&candidate)
			if err := ValidateGenesisState(candidate); err == nil {
				t.Fatal("malformed used nullifier accepted")
			}
		})
	}
}

func TestValidateGenesisRejectsNonCanonicalRatingNullifier(t *testing.T) {
	admin := sdk.AccAddress("rating-admin")
	nullifier := strings.Repeat("0", 63) + "1"
	genesis := GenesisState{Domains: []Domain{{
		Name: "Ratings", Admin: admin, Members: []string{admin.String()},
		Treasury: sdk.NewCoins(), PermissionReg: []string{},
		Issues: []Issue{{Name: "Issue", Suggestions: []Suggestion{{
			Name: "Suggestion", Creator: admin.String(),
			Ratings: []Rating{{NullifierHex: nullifier, Value: 1}},
		}}}},
	}}}
	if err := ValidateGenesisState(genesis); err != nil {
		t.Fatalf("canonical rating nullifier rejected: %v", err)
	}
	genesis.Domains[0].Issues[0].Suggestions[0].Ratings = []Rating{{NullifierHex: "01", Value: 1}}
	if err := ValidateGenesisState(genesis); err == nil {
		t.Fatal("short rating nullifier accepted")
	}
}

func TestProofAPIsRejectNonCanonicalPublicInputs(t *testing.T) {
	canonical := hashToField([]byte("canonical-public-input"))
	modulus := ecc.BN254.ScalarField().FillBytes(make([]byte, 32))
	keys := getTestZKPKeys(t)

	if err := VerifyMembershipProofForSignal(keys.VerifyingKey, nil, []byte{1}, canonical, canonical, canonical); err == nil {
		t.Fatal("short Merkle root accepted")
	}
	if err := VerifyMembershipProofForSignal(keys.VerifyingKey, nil, canonical, modulus, canonical, canonical); err == nil {
		t.Fatal("non-canonical nullifier accepted")
	}
	if err := VerifyMembershipProofForSignal(keys.VerifyingKey, nil, canonical, canonical, modulus, canonical); err == nil {
		t.Fatal("non-canonical external nullifier accepted")
	}
	if err := VerifyMembershipProofForSignal(keys.VerifyingKey, nil, canonical, canonical, canonical, modulus); err == nil {
		t.Fatal("non-canonical signal accepted")
	}
	if err := VerifyMembershipProofForSignal(keys.VerifyingKey, nil, canonical, canonical, canonical, make([]byte, 32)); err == nil {
		t.Fatal("zero signal accepted")
	}
}
