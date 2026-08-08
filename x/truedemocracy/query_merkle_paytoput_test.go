package truedemocracy

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	rewards "truerepublic/treasury/keeper"
)

// ---------- Test helpers ----------

// setupDomainWithCommitments creates a domain with count additional members,
// each registering one unique identity commitment, and returns the canonical
// commitment hex strings in registration order.
func setupDomainWithCommitments(t *testing.T, k Keeper, ctx sdk.Context, name string, count int) []string {
	t.Helper()
	admin := sdk.AccAddress(name + "-admin")
	k.CreateDomain(ctx, name, admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1_000_000)))
	commits := make([]string, count)
	for i := 0; i < count; i++ {
		memberAddr := sdk.AccAddress(fmt.Sprintf("%s-member-%d", name, i)).String()
		if err := k.AddMember(ctx, name, memberAddr, admin); err != nil {
			t.Fatalf("add member %d: %v", i, err)
		}
		commitment, err := ComputeCommitment(big.NewInt(int64(i + 1000)).Bytes())
		if err != nil {
			t.Fatalf("compute commitment %d: %v", i, err)
		}
		commits[i] = hex.EncodeToString(commitment)
		if err := k.RegisterIdentityCommitment(ctx, name, memberAddr, commits[i]); err != nil {
			t.Fatalf("register commitment %d: %v", i, err)
		}
	}
	return commits
}

// storeDomain writes domain state directly, bypassing keeper invariants, so
// tests can craft malformed state that must fail closed.
func storeDomain(t *testing.T, k Keeper, ctx sdk.Context, domain Domain) {
	t.Helper()
	store := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	store.Set([]byte("domain:"+domain.Name), bz)
}

// ---------- QueryMerkleProof Tests ----------

func TestQueryMerkleProof(t *testing.T) {
	k, ctx := setupKeeper(t)
	commits := setupDomainWithCommitments(t, k, ctx, "ProofDomain", 3)

	resp, err := k.MerkleProof(ctx, &QueryMerkleProofRequest{
		DomainName: "ProofDomain",
		Commitment: commits[1],
	})
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	var proof MerkleProofResult
	if err := json.Unmarshal(resp.Result, &proof); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if proof.DomainName != "ProofDomain" {
		t.Fatalf("expected domain ProofDomain, got %s", proof.DomainName)
	}
	if proof.Commitment != commits[1] {
		t.Fatalf("expected commitment %s, got %s", commits[1], proof.Commitment)
	}
	domain, found := k.GetDomain(ctx, "ProofDomain")
	if !found {
		t.Fatal("domain not found")
	}
	if proof.Root != domain.MerkleRoot {
		t.Fatalf("proof root %s does not match stored root %s", proof.Root, domain.MerkleRoot)
	}
	if len(proof.PathIndices) != MerkleTreeDepth || len(proof.PathElements) != MerkleTreeDepth {
		t.Fatalf("expected %d path levels, got %d indices and %d elements",
			MerkleTreeDepth, len(proof.PathIndices), len(proof.PathElements))
	}

	rootBytes, err := hex.DecodeString(proof.Root)
	if err != nil {
		t.Fatalf("decode root: %v", err)
	}
	leafBytes, err := hex.DecodeString(proof.Commitment)
	if err != nil {
		t.Fatalf("decode commitment: %v", err)
	}
	siblings := make([][]byte, len(proof.PathElements))
	for i, element := range proof.PathElements {
		siblings[i], err = hex.DecodeString(element)
		if err != nil {
			t.Fatalf("decode path element %d: %v", i, err)
		}
	}
	if !VerifyMerkleProof(rootBytes, leafBytes, siblings, proof.PathIndices) {
		t.Fatal("returned proof does not verify against the stored root")
	}
}

func TestQueryMerkleProofFirstAndLastLeaf(t *testing.T) {
	k, ctx := setupKeeper(t)
	commits := setupDomainWithCommitments(t, k, ctx, "EdgeDomain", 4)

	for i, commitment := range commits {
		resp, err := k.MerkleProof(ctx, &QueryMerkleProofRequest{
			DomainName: "EdgeDomain",
			Commitment: commitment,
		})
		if err != nil {
			t.Fatalf("query for leaf %d failed: %v", i, err)
		}
		var proof MerkleProofResult
		if err := json.Unmarshal(resp.Result, &proof); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		rootBytes, _ := hex.DecodeString(proof.Root)
		leafBytes, _ := hex.DecodeString(commitment)
		siblings := make([][]byte, len(proof.PathElements))
		for j, element := range proof.PathElements {
			siblings[j], _ = hex.DecodeString(element)
		}
		if !VerifyMerkleProof(rootBytes, leafBytes, siblings, proof.PathIndices) {
			t.Fatalf("proof for leaf %d does not verify", i)
		}
	}
}

func TestQueryMerkleProofInvalidRequest(t *testing.T) {
	k, ctx := setupKeeper(t)
	commits := setupDomainWithCommitments(t, k, ctx, "ProofDomain", 2)

	cases := []struct {
		name string
		req  *QueryMerkleProofRequest
	}{
		{"empty domain name", &QueryMerkleProofRequest{DomainName: "", Commitment: commits[0]}},
		{"missing domain", &QueryMerkleProofRequest{DomainName: "NoDomain", Commitment: commits[0]}},
		{"empty commitment", &QueryMerkleProofRequest{DomainName: "ProofDomain", Commitment: ""}},
		{"short commitment", &QueryMerkleProofRequest{DomainName: "ProofDomain", Commitment: "aabb"}},
		{"non-hex commitment", &QueryMerkleProofRequest{DomainName: "ProofDomain", Commitment: strings.Repeat("zz", 32)}},
		{"unknown commitment", &QueryMerkleProofRequest{DomainName: "ProofDomain", Commitment: hex.EncodeToString(make([]byte, 32))}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := k.MerkleProof(ctx, tc.req); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestQueryMerkleProofDomainWithoutCommitments(t *testing.T) {
	k, ctx := setupKeeper(t)
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "EmptyDomain", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 100_000)))

	_, err := k.MerkleProof(ctx, &QueryMerkleProofRequest{
		DomainName: "EmptyDomain",
		Commitment: hex.EncodeToString(make([]byte, 32)),
	})
	if err == nil {
		t.Fatal("expected error for domain without a Merkle root")
	}
}

func TestQueryMerkleProofRejectsDuplicateCommitmentState(t *testing.T) {
	k, ctx := setupKeeper(t)
	commits := setupDomainWithCommitments(t, k, ctx, "ProofDomain", 2)

	domain, found := k.GetDomain(ctx, "ProofDomain")
	if !found {
		t.Fatal("domain not found")
	}
	domain.IdentityCommits = append(domain.IdentityCommits, domain.IdentityCommits[0])
	storeDomain(t, k, ctx, domain)

	if _, err := k.MerkleProof(ctx, &QueryMerkleProofRequest{
		DomainName: "ProofDomain",
		Commitment: commits[0],
	}); err == nil {
		t.Fatal("expected error for duplicate identity commitment state")
	}
}

func TestQueryMerkleProofRejectsMalformedCommitmentState(t *testing.T) {
	k, ctx := setupKeeper(t)
	commits := setupDomainWithCommitments(t, k, ctx, "ProofDomain", 2)

	domain, found := k.GetDomain(ctx, "ProofDomain")
	if !found {
		t.Fatal("domain not found")
	}
	domain.IdentityCommits[1] = "zz"
	storeDomain(t, k, ctx, domain)

	if _, err := k.MerkleProof(ctx, &QueryMerkleProofRequest{
		DomainName: "ProofDomain",
		Commitment: commits[0],
	}); err == nil {
		t.Fatal("expected error for malformed identity commitment state")
	}
}

func TestQueryMerkleProofRejectsRootMismatch(t *testing.T) {
	k, ctx := setupKeeper(t)
	commits := setupDomainWithCommitments(t, k, ctx, "ProofDomain", 2)

	domain, found := k.GetDomain(ctx, "ProofDomain")
	if !found {
		t.Fatal("domain not found")
	}
	domain.MerkleRoot = strings.Repeat("00", 32)
	storeDomain(t, k, ctx, domain)

	if _, err := k.MerkleProof(ctx, &QueryMerkleProofRequest{
		DomainName: "ProofDomain",
		Commitment: commits[0],
	}); err == nil {
		t.Fatal("expected error for Merkle root mismatch")
	}
}

// ---------- QueryPayToPut Tests ----------

func TestQueryPayToPut(t *testing.T) {
	k, ctx := setupKeeper(t)
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "PayDomain", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1_000_000)))

	resp, err := k.PayToPut(ctx, &QueryPayToPutRequest{DomainName: "PayDomain"})
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	var result PayToPutResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	// reward = 1,000,000 / 1000 = 1000; multiplier = min(15, 1 member) = 1.
	if result.BaseCost != "1000" {
		t.Fatalf("expected base cost 1000, got %s", result.BaseCost)
	}
	if result.DomainMultiplier != 1 {
		t.Fatalf("expected multiplier 1, got %d", result.DomainMultiplier)
	}
	if result.FinalCost != "1000" {
		t.Fatalf("expected final cost 1000, got %s", result.FinalCost)
	}
	if !strings.Contains(result.Formula, "CEarn=1000") || !strings.Contains(result.Formula, "CPut=15") {
		t.Fatalf("formula does not expose canonical constants: %s", result.Formula)
	}
}

func TestQueryPayToPutMultiplierCappedByCPut(t *testing.T) {
	k, ctx := setupKeeper(t)
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "PayDomain", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 1_000_000)))
	for i := 0; i < 20; i++ {
		memberAddr := sdk.AccAddress(fmt.Sprintf("pay-member-%d", i)).String()
		if err := k.AddMember(ctx, "PayDomain", memberAddr, admin); err != nil {
			t.Fatalf("add member %d: %v", i, err)
		}
	}

	resp, err := k.PayToPut(ctx, &QueryPayToPutRequest{DomainName: "PayDomain"})
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	var result PayToPutResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	// 21 members > CPut; reward = 1000; final = 1000 * 15.
	if result.DomainMultiplier != rewards.CPut {
		t.Fatalf("expected multiplier %d, got %d", rewards.CPut, result.DomainMultiplier)
	}
	if result.FinalCost != "15000" {
		t.Fatalf("expected final cost 15000, got %s", result.FinalCost)
	}
}

func TestQueryPayToPutMatchesSubmitProposalCalculation(t *testing.T) {
	k, ctx := setupKeeper(t)
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "PayDomain", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 500_000)))
	for i := 0; i < 9; i++ {
		memberAddr := sdk.AccAddress(fmt.Sprintf("pay-member-%d", i)).String()
		if err := k.AddMember(ctx, "PayDomain", memberAddr, admin); err != nil {
			t.Fatalf("add member %d: %v", i, err)
		}
	}

	resp, err := k.PayToPut(ctx, &QueryPayToPutRequest{DomainName: "PayDomain"})
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	var result PayToPutResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	domain, found := k.GetDomain(ctx, "PayDomain")
	if !found {
		t.Fatal("domain not found")
	}
	// The query must expose exactly what Keeper.SubmitProposal enforces.
	want := rewards.CalcPutPrice(domain.Treasury.AmountOf(PNYXDenom), int64(len(domain.Members)))
	if result.FinalCost != want.String() {
		t.Fatalf("query final cost %s does not match canonical put price %s", result.FinalCost, want)
	}
	wantBase := rewards.CalcReward(domain.Treasury.AmountOf(PNYXDenom))
	if result.BaseCost != wantBase.String() {
		t.Fatalf("query base cost %s does not match canonical reward %s", result.BaseCost, wantBase)
	}
}

func TestQueryPayToPutEmptyTreasury(t *testing.T) {
	k, ctx := setupKeeper(t)
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "PayDomain", admin, sdk.NewCoins())

	resp, err := k.PayToPut(ctx, &QueryPayToPutRequest{DomainName: "PayDomain"})
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	var result PayToPutResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if result.BaseCost != "0" || result.FinalCost != "0" {
		t.Fatalf("expected zero costs for empty treasury, got base=%s final=%s", result.BaseCost, result.FinalCost)
	}
}

func TestQueryPayToPutInvalidRequest(t *testing.T) {
	k, ctx := setupKeeper(t)
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "PayDomain", admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 100_000)))

	if _, err := k.PayToPut(ctx, &QueryPayToPutRequest{DomainName: ""}); err == nil {
		t.Fatal("expected error for empty domain name")
	}
	if _, err := k.PayToPut(ctx, &QueryPayToPutRequest{DomainName: "NoDomain"}); err == nil {
		t.Fatal("expected error for missing domain")
	}
}

// ---------- Query CLI registration ----------

func TestQueryCommandsIncludeMerkleProofAndPayToPut(t *testing.T) {
	cmd := GetQueryCmd(codec.NewLegacyAmino())
	found := map[string]bool{}
	for _, sub := range cmd.Commands() {
		found[sub.Name()] = true
	}
	for _, name := range []string{"merkle-proof", "pay-to-put"} {
		if !found[name] {
			t.Errorf("query command %q is not registered", name)
		}
	}
}
