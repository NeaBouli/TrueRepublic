package truedemocracy

// GH-266: the dedicated GH-206 Go/WASM gate now also feeds the fresh
// maintained-client v2 Groth16 proof through the real keeper
// RateProposalWithZKP reward boundary via RateProposalWithZKPPayout.
// Everything here is env-gated, fixture-pinned and test-only: synthetic
// single-party toxic-waste artifacts, an in-memory keeper and disposable
// local accounts. No production, consensus, migration or runtime behavior
// changes; the maintained client remains non-submittable.

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

const (
	zkpWASMHandoffEnv      = "TRUEREPUBLIC_ZKP_RESULT_PATH"
	zkpWASMHandoffSchema   = "truerepublic/zkp-wasm-handoff/v1"
	zkpWASMHandoffMaxBytes = 16 * 1024
	zkpWASMProofHexChars   = 2 * 164 // canonical Groth16 proof serialization
	zkpWASMMaxNameLength   = 128
)

// zkpWASMHandoff is the strict GH-266 handoff written by the maintained
// client's GH-206 WASM integration test. Every field is required; unknown,
// duplicate or trailing JSON fails closed.
type zkpWASMHandoff struct {
	Schema          string   `json:"schema"`
	ChainID         string   `json:"chainId"`
	DomainName      string   `json:"domainName"`
	IssueName       string   `json:"issueName"`
	SuggestionName  string   `json:"suggestionName"`
	Rating          int      `json:"rating"`
	RewardRecipient string   `json:"rewardRecipient"`
	Proof           string   `json:"proof"`
	NullifierHash   string   `json:"nullifierHash"`
	MerkleRoot      string   `json:"merkleRoot"`
	PublicSignals   []string `json:"publicSignals"`
}

// The x/truedemocracy test binary otherwise runs with the SDK default bech32
// prefix, while the pinned GH-209 v2 vector recipient is a canonical
// truerepublic address exactly as in production (app.go configureSDKConfig).
// The gate must therefore mirror the production prefix before any address is
// encoded or decoded. This runs only inside the env-gated tests, which the
// dedicated script executes in their own filtered test binary, so ordinary
// package runs are unaffected.
var zkpWASMAddressPrefixOnce sync.Once

func configureWASMTestAddressPrefix() {
	zkpWASMAddressPrefixOnce.Do(func() {
		sdk.GetConfig().SetBech32PrefixForAccount("truerepublic", "truerepublicpub")
	})
}

// decodeWASMHandoffStrict bounds and strictly decodes the maintained-client
// handoff: exact schema, no unknown/duplicate/trailing JSON, canonical
// lowercase hex, canonical BN254 field elements, the pinned public-signal
// binding and a canonical reward recipient.
func decodeWASMHandoffStrict(data []byte) (zkpWASMHandoff, error) {
	var handoff zkpWASMHandoff
	if len(data) == 0 || len(data) > zkpWASMHandoffMaxBytes {
		return handoff, fmt.Errorf("handoff size %d is empty or exceeds the %d-byte bound", len(data), zkpWASMHandoffMaxBytes)
	}
	if err := strictJSON(data, &handoff); err != nil {
		return handoff, fmt.Errorf("strict handoff decode: %w", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return handoff, fmt.Errorf("inspect handoff fields: %w", err)
	}
	requiredFields := []string{
		"schema", "chainId", "domainName", "issueName", "suggestionName",
		"rating", "rewardRecipient", "proof", "nullifierHash", "merkleRoot",
		"publicSignals",
	}
	if len(fields) != len(requiredFields) {
		return handoff, fmt.Errorf("handoff carries %d fields, want exactly %d", len(fields), len(requiredFields))
	}
	for _, field := range requiredFields {
		if _, ok := fields[field]; !ok {
			return handoff, fmt.Errorf("handoff field %q is required", field)
		}
	}
	if handoff.Schema != zkpWASMHandoffSchema {
		return handoff, fmt.Errorf("handoff schema %q is not %q", handoff.Schema, zkpWASMHandoffSchema)
	}
	for _, name := range []struct {
		label string
		value string
	}{
		{"chain id", handoff.ChainID},
		{"domain name", handoff.DomainName},
		{"issue name", handoff.IssueName},
		{"suggestion name", handoff.SuggestionName},
	} {
		if name.value == "" || len(name.value) > zkpWASMMaxNameLength {
			return handoff, fmt.Errorf("handoff %s is empty or exceeds %d bytes", name.label, zkpWASMMaxNameLength)
		}
	}
	if handoff.Rating < -5 || handoff.Rating > 5 {
		return handoff, fmt.Errorf("handoff rating %d is out of range", handoff.Rating)
	}
	if _, err := ValidateRewardRecipient(handoff.RewardRecipient); err != nil {
		return handoff, fmt.Errorf("handoff reward recipient: %w", err)
	}
	if len(handoff.Proof) != zkpWASMProofHexChars || !isWASMLowerHex(handoff.Proof) {
		return handoff, fmt.Errorf("handoff proof must be exactly %d lowercase hex characters", zkpWASMProofHexChars)
	}
	if _, err := hex.DecodeString(handoff.Proof); err != nil {
		return handoff, fmt.Errorf("handoff proof is not hex: %w", err)
	}
	canonicalField := func(label, value string) error {
		if len(value) != 64 || !isWASMLowerHex(value) {
			return fmt.Errorf("handoff %s must be exactly 64 lowercase hex characters", label)
		}
		if _, err := HexToFieldElement(value); err != nil {
			return fmt.Errorf("handoff %s is not a canonical BN254 field element: %w", label, err)
		}
		return nil
	}
	if err := canonicalField("merkle root", handoff.MerkleRoot); err != nil {
		return handoff, err
	}
	if err := canonicalField("nullifier hash", handoff.NullifierHash); err != nil {
		return handoff, err
	}
	if len(handoff.PublicSignals) != membershipPublicWitnessCount {
		return handoff, fmt.Errorf("handoff carries %d public signals, want %d", len(handoff.PublicSignals), membershipPublicWitnessCount)
	}
	for i, signal := range handoff.PublicSignals {
		if err := canonicalField(fmt.Sprintf("public signal %d", i), signal); err != nil {
			return handoff, err
		}
	}
	if handoff.PublicSignals[0] != handoff.MerkleRoot || handoff.PublicSignals[1] != handoff.NullifierHash {
		return handoff, fmt.Errorf("handoff public signals do not bind the merkle root and nullifier")
	}
	return handoff, nil
}

func isWASMLowerHex(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}

// verifyWASMHandoffPinnedBindings requires the handoff vote context to match
// the pinned GH-198 golden vector and the GH-203/GH-209 v2 circuit vector
// exactly, so a drifting client cannot steer the keeper toward another
// chain, proposal, rating, root, nullifier or recipient.
func verifyWASMHandoffPinnedBindings(handoff zkpWASMHandoff, vector zkpGoldenVector, v2 zkpSpecVoteContextV2Vector) error {
	mismatch := func(field, got, want string) error {
		return fmt.Errorf("handoff %s %q does not match the pinned fixture %q", field, got, want)
	}
	if handoff.ChainID != vector.ChainID || handoff.ChainID != v2.ChainID {
		return mismatch("chain id", handoff.ChainID, vector.ChainID)
	}
	if handoff.DomainName != vector.DomainName || handoff.DomainName != v2.DomainName {
		return mismatch("domain name", handoff.DomainName, vector.DomainName)
	}
	if handoff.IssueName != vector.IssueName || handoff.IssueName != v2.IssueName {
		return mismatch("issue name", handoff.IssueName, vector.IssueName)
	}
	if handoff.SuggestionName != vector.SuggestionName || handoff.SuggestionName != v2.SuggestionName {
		return mismatch("suggestion name", handoff.SuggestionName, vector.SuggestionName)
	}
	if handoff.Rating != vector.Rating || handoff.Rating != v2.Rating {
		return fmt.Errorf("handoff rating %d does not match the pinned fixture %d", handoff.Rating, vector.Rating)
	}
	if handoff.RewardRecipient != v2.RewardRecipient {
		return mismatch("reward recipient", handoff.RewardRecipient, v2.RewardRecipient)
	}
	if handoff.MerkleRoot != vector.MerkleRootHex {
		return mismatch("merkle root", handoff.MerkleRoot, vector.MerkleRootHex)
	}
	if handoff.NullifierHash != vector.NullifierHashHex {
		return mismatch("nullifier hash", handoff.NullifierHash, vector.NullifierHashHex)
	}
	if handoff.PublicSignals[2] != vector.ExternalNullifierHex {
		return mismatch("external nullifier signal", handoff.PublicSignals[2], vector.ExternalNullifierHex)
	}
	if handoff.PublicSignals[3] != v2.SignalHashHex {
		return mismatch("v2 signal hash", handoff.PublicSignals[3], v2.SignalHashHex)
	}
	return nil
}

// loadVerifiedWASMHandoff reads the env-gated handoff, strictly decodes it
// and cross-checks it against the pinned fixtures. The fixture VK bytes are
// returned for keeper pinning; their checksum was already manifest-verified
// by readVerifiedZKPFixture.
func loadVerifiedWASMHandoff(t *testing.T) (zkpWASMHandoff, zkpGoldenVector, []byte) {
	t.Helper()
	path := os.Getenv(zkpWASMHandoffEnv)
	if path == "" {
		t.Skipf("set %s in the dedicated WASM integration gate", zkpWASMHandoffEnv)
	}
	// Do not mutate the process-global SDK prefix on the normal skip path.
	// The dedicated script runs only these tests in an isolated test binary.
	configureWASMTestAddressPrefix()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	handoff, err := decodeWASMHandoffStrict(data)
	if err != nil {
		t.Fatalf("malformed maintained-client handoff: %v", err)
	}
	_, vector, _, _ := readVerifiedZKPFixture(t)
	spec := loadZKPCircuitSpec(t)
	if err := verifyWASMHandoffPinnedBindings(handoff, vector, spec.VoteContextV2.Vector); err != nil {
		t.Fatal(err)
	}
	vkBytes, err := os.ReadFile(filepath.Join(zkpFixtureDir, "membership_v2.vk"))
	if err != nil {
		t.Fatal(err)
	}
	return handoff, vector, vkBytes
}

// setupWASMFixtureKeeper builds an in-memory keeper whose domain, identity
// tree, proposal and pinned VK exactly match the fixture the maintained
// client proved against, backed by an escrowed mock bank.
func setupWASMFixtureKeeper(t *testing.T, handoff zkpWASMHandoff, vector zkpGoldenVector, vkBytes []byte, chainID string) (Keeper, sdk.Context, *mockBankKeeper) {
	t.Helper()
	k, ctx := setupKeeper(t)
	ctx = ctx.WithChainID(chainID)
	admin := sdk.AccAddress("wasm-fixture-admin")
	member := sdk.AccAddress("wasm-fixture-member")
	k.CreateDomain(ctx, handoff.DomainName, admin, sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 500_000)))
	if err := k.AddMember(ctx, handoff.DomainName, member.String(), admin); err != nil {
		t.Fatal(err)
	}
	if err := k.RegisterIdentityCommitment(ctx, handoff.DomainName, member.String(), vector.CommitmentHex); err != nil {
		t.Fatal(err)
	}
	domain, found := k.GetDomain(ctx, handoff.DomainName)
	if !found || domain.MerkleRoot != handoff.MerkleRoot {
		t.Fatalf("fixture root mismatch: got %q want %q", domain.MerkleRoot, handoff.MerkleRoot)
	}
	addProposal(t, k, ctx, handoff.DomainName, handoff.IssueName, handoff.SuggestionName)
	k.SetVerifyingKey(ctx, vkBytes)
	bank := backExistingEscrow(&k, ctx)
	return k, ctx, bank
}

// wasmKeeperRatingCall carries the mutable fields of one keeper replay so
// each adversarial case can deviate from the valid handoff in exactly one
// dimension.
type wasmKeeperRatingCall struct {
	rating    int
	proof     string
	nullifier string
	root      string
	recipient string
}

func (call wasmKeeperRatingCall) rate(k Keeper, ctx sdk.Context, handoff zkpWASMHandoff) error {
	_, err := k.RateProposalWithZKPPayout(
		ctx, handoff.DomainName, handoff.IssueName, handoff.SuggestionName,
		call.rating, call.proof, call.nullifier, call.root, call.recipient,
	)
	return err
}

func TestWASMClientProofTraversesKeeperRewardBoundary(t *testing.T) {
	handoff, vector, vkBytes := loadVerifiedWASMHandoff(t)

	recipientAddr, err := ValidateRewardRecipient(handoff.RewardRecipient)
	if err != nil {
		t.Fatal(err)
	}
	baseCall := func() wasmKeeperRatingCall {
		return wasmKeeperRatingCall{
			rating:    handoff.Rating,
			proof:     handoff.Proof,
			nullifier: handoff.NullifierHash,
			root:      handoff.MerkleRoot,
			recipient: handoff.RewardRecipient,
		}
	}

	t.Run("valid proof pays the bound recipient exactly once", func(t *testing.T) {
		k, ctx, bank := setupWASMFixtureKeeper(t, handoff, vector, vkBytes, handoff.ChainID)
		before := snapshotAccounting(t, k, ctx, bank, handoff.DomainName)
		reward := expectedReward(before.treasury)
		if reward <= 0 {
			t.Fatal("expected a positive reward for this fixture")
		}

		paid, err := k.RateProposalWithZKPPayout(
			ctx, handoff.DomainName, handoff.IssueName, handoff.SuggestionName,
			handoff.Rating, handoff.Proof, handoff.NullifierHash, handoff.MerkleRoot, handoff.RewardRecipient,
		)
		if err != nil {
			t.Fatalf("fresh maintained-client proof rejected: %v", err)
		}
		if got := paid.AmountOf(PNYXDenom).Int64(); got != reward {
			t.Fatalf("paid reward = %d, want exact %d", got, reward)
		}

		domain, _ := k.GetDomain(ctx, handoff.DomainName)
		if got := domain.Treasury.AmountOf(PNYXDenom).Int64(); got != before.treasury-reward {
			t.Fatalf("treasury = %d, want %d", got, before.treasury-reward)
		}
		if domain.TotalPayouts != before.totalPayouts+reward {
			t.Fatalf("total payouts = %d, want %d", domain.TotalPayouts, before.totalPayouts+reward)
		}
		if got := moduleBalance(bank); got != before.moduleBalance-reward {
			t.Fatalf("module balance = %d, want %d", got, before.moduleBalance-reward)
		}
		if got := accountBalance(bank, recipientAddr); got != reward {
			t.Fatalf("bound recipient balance = %d, want exact reward %d", got, reward)
		}
		if !k.IsNullifierUsed(ctx, handoff.DomainName, handoff.NullifierHash) {
			t.Fatal("nullifier not consumed after success")
		}
		ratings := domain.Issues[0].Suggestions[0].Ratings
		if len(ratings) != 1 || ratings[0].NullifierHex != handoff.NullifierHash || ratings[0].Value != handoff.Rating {
			t.Fatalf("rating not recorded exactly once: %+v", ratings)
		}
		if err := k.ValidateEscrowParity(ctx); err != nil {
			t.Fatalf("escrow parity after payout: %v", err)
		}

		// An exact replay hits the consumed nullifier and cannot pay twice.
		if err := baseCall().rate(k, ctx, handoff); err == nil || !strings.Contains(err.Error(), "nullifier already used") {
			t.Fatalf("replay result = %v", err)
		}
		if got := accountBalance(bank, recipientAddr); got != reward {
			t.Fatalf("replay changed recipient balance: %d, want %d", got, reward)
		}
		if after, want := snapshotAccounting(t, k, ctx, bank, handoff.DomainName), (payoutSnapshot{
			treasury:      before.treasury - reward,
			totalPayouts:  before.totalPayouts + reward,
			moduleBalance: before.moduleBalance - reward,
		}); after != want {
			t.Fatalf("replay mutated accounting: %+v, want %+v", after, want)
		}
		if err := k.ValidateEscrowParity(ctx); err != nil {
			t.Fatalf("escrow parity after replay: %v", err)
		}
	})

	t.Run("corrupted proof fails closed", func(t *testing.T) {
		k, ctx, bank := setupWASMFixtureKeeper(t, handoff, vector, vkBytes, handoff.ChainID)
		before := snapshotAccounting(t, k, ctx, bank, handoff.DomainName)
		call := baseCall()
		proofBytes, err := hex.DecodeString(call.proof)
		if err != nil {
			t.Fatal(err)
		}
		proofBytes[len(proofBytes)/2] ^= 1
		call.proof = hex.EncodeToString(proofBytes)
		if err := call.rate(k, ctx, handoff); err == nil {
			t.Fatal("corrupted proof accepted")
		}
		assertUnchangedAfterFailure(t, k, ctx, bank, handoff.DomainName, handoff.NullifierHash, before, recipientAddr)
	})

	t.Run("recipient substitution fails closed", func(t *testing.T) {
		k, ctx, bank := setupWASMFixtureKeeper(t, handoff, vector, vkBytes, handoff.ChainID)
		before := snapshotAccounting(t, k, ctx, bank, handoff.DomainName)
		attacker := sdk.AccAddress("wasm-attacker")
		call := baseCall()
		call.recipient = attacker.String()
		if err := call.rate(k, ctx, handoff); err == nil {
			t.Fatal("recipient-substituted replay accepted")
		}
		assertUnchangedAfterFailure(t, k, ctx, bank, handoff.DomainName, handoff.NullifierHash, before, attacker)
		if got := accountBalance(bank, recipientAddr); got != 0 {
			t.Fatalf("bound recipient paid on a rejected substitution: %d", got)
		}
	})

	t.Run("rating drift fails closed", func(t *testing.T) {
		k, ctx, bank := setupWASMFixtureKeeper(t, handoff, vector, vkBytes, handoff.ChainID)
		before := snapshotAccounting(t, k, ctx, bank, handoff.DomainName)
		call := baseCall()
		call.rating = handoff.Rating + 1
		if err := call.rate(k, ctx, handoff); err == nil {
			t.Fatal("rating-drifted replay accepted")
		}
		assertUnchangedAfterFailure(t, k, ctx, bank, handoff.DomainName, handoff.NullifierHash, before, recipientAddr)
	})

	t.Run("wrong chain scope fails closed", func(t *testing.T) {
		k, ctx, bank := setupWASMFixtureKeeper(t, handoff, vector, vkBytes, handoff.ChainID+"-foreign")
		before := snapshotAccounting(t, k, ctx, bank, handoff.DomainName)
		if err := baseCall().rate(k, ctx, handoff); err == nil {
			t.Fatal("cross-chain replay accepted")
		}
		assertUnchangedAfterFailure(t, k, ctx, bank, handoff.DomainName, handoff.NullifierHash, before, recipientAddr)
	})

	t.Run("wrong merkle root fails closed", func(t *testing.T) {
		k, ctx, bank := setupWASMFixtureKeeper(t, handoff, vector, vkBytes, handoff.ChainID)
		before := snapshotAccounting(t, k, ctx, bank, handoff.DomainName)
		call := baseCall()
		rootBytes, err := hex.DecodeString(call.root)
		if err != nil {
			t.Fatal(err)
		}
		rootBytes[31] ^= 1
		call.root = hex.EncodeToString(rootBytes)
		if err := call.rate(k, ctx, handoff); err == nil || !strings.Contains(err.Error(), "merkle root not recognized") {
			t.Fatalf("wrong-root result = %v", err)
		}
		assertUnchangedAfterFailure(t, k, ctx, bank, handoff.DomainName, handoff.NullifierHash, before, recipientAddr)
	})

	t.Run("noncanonical recipient fails closed", func(t *testing.T) {
		k, ctx, bank := setupWASMFixtureKeeper(t, handoff, vector, vkBytes, handoff.ChainID)
		before := snapshotAccounting(t, k, ctx, bank, handoff.DomainName)
		call := baseCall()
		call.recipient = strings.ToUpper(handoff.RewardRecipient)
		if err := call.rate(k, ctx, handoff); err == nil {
			t.Fatal("noncanonical recipient accepted")
		}
		assertUnchangedAfterFailure(t, k, ctx, bank, handoff.DomainName, handoff.NullifierHash, before, recipientAddr)
	})

	t.Run("blocked module recipient fails closed", func(t *testing.T) {
		k, ctx, bank := setupWASMFixtureKeeper(t, handoff, vector, vkBytes, handoff.ChainID)
		bank.blockModuleAccount(ModuleName)
		before := snapshotAccounting(t, k, ctx, bank, handoff.DomainName)
		moduleRecipient := authtypes.NewModuleAddress(ModuleName)
		call := baseCall()
		call.recipient = moduleRecipient.String()
		if err := call.rate(k, ctx, handoff); err == nil || !strings.Contains(err.Error(), "module account") {
			t.Fatalf("module-account recipient result = %v", err)
		}
		assertUnchangedAfterFailure(t, k, ctx, bank, handoff.DomainName, handoff.NullifierHash, before, moduleRecipient)
		if got := accountBalance(bank, recipientAddr); got != 0 {
			t.Fatalf("bound recipient paid on a rejected module recipient: %d", got)
		}
	})
}

func TestWASMClientHandoffStrictDecoding(t *testing.T) {
	path := os.Getenv(zkpWASMHandoffEnv)
	if path == "" {
		t.Skipf("set %s in the dedicated WASM integration gate", zkpWASMHandoffEnv)
	}
	// Keep ordinary package runs byte-for-byte free of global prefix changes.
	configureWASMTestAddressPrefix()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	handoff, err := decodeWASMHandoffStrict(raw)
	if err != nil {
		t.Fatalf("fresh maintained-client handoff rejected: %v", err)
	}
	_, vector, _, _ := readVerifiedZKPFixture(t)
	spec := loadZKPCircuitSpec(t)
	v2 := spec.VoteContextV2.Vector
	if err := verifyWASMHandoffPinnedBindings(handoff, vector, v2); err != nil {
		t.Fatal(err)
	}

	trimmed := strings.TrimSpace(string(raw))
	mutateJSON := func(t *testing.T, drop string, edit func(map[string]any)) []byte {
		t.Helper()
		var decoded map[string]any
		if err := json.Unmarshal([]byte(trimmed), &decoded); err != nil {
			t.Fatal(err)
		}
		if drop != "" {
			delete(decoded, drop)
		}
		if edit != nil {
			edit(decoded)
		}
		mutated, err := json.Marshal(decoded)
		if err != nil {
			t.Fatal(err)
		}
		return mutated
	}

	cases := map[string][]byte{
		"extra field":     mutateJSON(t, "", func(m map[string]any) { m["unexpected"] = true }),
		"missing proof":   mutateJSON(t, "proof", nil),
		"missing signals": mutateJSON(t, "publicSignals", nil),
		"duplicate key":   []byte(strings.Replace(trimmed, "{", `{"schema":"duplicate",`, 1)),
		"trailing value":  []byte(trimmed + " {}"),
		"wrong schema":    mutateJSON(t, "", func(m map[string]any) { m["schema"] = "truerepublic/zkp-wasm-handoff/v0" }),
		"uppercase hex":   mutateJSON(t, "", func(m map[string]any) { m["nullifierHash"] = strings.ToUpper(handoff.NullifierHash) }),
		"swapped signals": mutateJSON(t, "", func(m map[string]any) {
			m["publicSignals"] = []string{handoff.PublicSignals[1], handoff.PublicSignals[0], handoff.PublicSignals[2], handoff.PublicSignals[3]}
		}),
		"oversized input": append(append([]byte(nil), raw...), make([]byte, zkpWASMHandoffMaxBytes)...),
	}
	for name, candidate := range cases {
		if _, err := decodeWASMHandoffStrict(candidate); err == nil {
			t.Fatalf("%s accepted", name)
		}
	}

	// A well-formed handoff that drifts from any pinned binding is rejected
	// before it can steer the keeper.
	drift := map[string]func(h *zkpWASMHandoff){
		"chain drift":     func(h *zkpWASMHandoff) { h.ChainID += "-other" },
		"rating drift":    func(h *zkpWASMHandoff) { h.Rating++ },
		"recipient drift": func(h *zkpWASMHandoff) { h.RewardRecipient = sdk.AccAddress("wasm-drift").String() },
		"root drift":      func(h *zkpWASMHandoff) { h.MerkleRoot = "00" + h.MerkleRoot[2:] },
		"nullifier drift": func(h *zkpWASMHandoff) { h.NullifierHash = "00" + h.NullifierHash[2:] },
		"scope drift":     func(h *zkpWASMHandoff) { h.PublicSignals[2] = h.PublicSignals[3] },
		"signal drift":    func(h *zkpWASMHandoff) { h.PublicSignals[3] = h.PublicSignals[2] },
	}
	for name, apply := range drift {
		candidate := handoff
		candidate.PublicSignals = append([]string(nil), handoff.PublicSignals...)
		apply(&candidate)
		if err := verifyWASMHandoffPinnedBindings(candidate, vector, v2); err == nil {
			t.Fatalf("%s accepted", name)
		}
	}
}
