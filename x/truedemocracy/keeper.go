package truedemocracy

import (
	"context"
	"encoding/hex"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"truerepublic/token"
	rewards "truerepublic/treasury/keeper"
)

// BankKeeper defines the x/bank methods required for the treasury bridge.
type BankKeeper interface {
	token.IssuanceBankKeeper
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}

type Keeper struct {
	StoreKey   storetypes.StoreKey
	nodes      []*Node
	cdc        *codec.LegacyAmino
	bankKeeper BankKeeper // nil until x/bank is wired (bridge functions check)
	issuer     token.IssuanceService
}

func NewKeeper(cdc *codec.LegacyAmino, storeKey storetypes.StoreKey, nodes []*Node, bankKeeper BankKeeper) Keeper {
	return Keeper{
		StoreKey:   storeKey,
		nodes:      nodes,
		cdc:        cdc,
		bankKeeper: bankKeeper,
		issuer:     token.NewIssuanceService(bankKeeper, ModuleName),
	}
}

// GetDomain loads a domain from the KV store by name.
func (k Keeper) GetDomain(ctx sdk.Context, name string) (Domain, bool) {
	store := ctx.KVStore(k.StoreKey)
	bz := store.Get([]byte("domain:" + name))
	if bz == nil {
		return Domain{}, false
	}
	var domain Domain
	k.cdc.MustUnmarshalLengthPrefixed(bz, &domain)
	return domain, true
}

func (k Keeper) CreateDomain(ctx sdk.Context, name string, admin sdk.AccAddress, initialCoins sdk.Coins) {
	store := ctx.KVStore(k.StoreKey)
	domain := Domain{
		Name:          name,
		Admin:         admin,
		Members:       []string{admin.String()},
		Treasury:      initialCoins,
		Issues:        []Issue{},
		Options:       DomainOptions{AdminElectable: true, AnyoneCanJoin: false},
		PermissionReg: []string{},
	}
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	store.Set([]byte("domain:"+name), bz)

	// Initialize automated Big Purge schedule (WP S4).
	k.InitializeBigPurgeSchedule(ctx, name)
}

// AddMember adds a new member to a domain. Only the domain admin can add
// members. This is step 1 of the two-step onboarding flow (WP S4).
func (k Keeper) AddMember(ctx sdk.Context, domainName, newMember string, caller sdk.AccAddress) error {
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "domain %s not found", domainName)
	}

	if !caller.Equals(domain.Admin) {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "only domain admin can add members")
	}

	for _, m := range domain.Members {
		if m == newMember {
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "member already exists in domain")
		}
	}

	domain.Members = append(domain.Members, newMember)

	store := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	store.Set([]byte("domain:"+domainName), bz)
	return nil
}

func (k Keeper) SubmitProposal(ctx sdk.Context, domainName, issueName, suggestionName, creator string, fee sdk.Coins, externalLink string) error {
	store := ctx.KVStore(k.StoreKey)
	domainBz := store.Get([]byte("domain:" + domainName))
	if domainBz == nil {
		return errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "Domain not found")
	}
	var domain Domain
	k.cdc.MustUnmarshalLengthPrefixed(domainBz, &domain)

	if domain.Options.OnlyAdminIssues && creator != domain.Admin.String() {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "Only admin can submit issues")
	}
	if domain.Options.CoinBurnRequired && fee.AmountOf(PNYXDenom).LT(rewards.CalcDomainCost(fee.AmountOf(PNYXDenom))) {
		return errorsmod.Wrap(sdkerrors.ErrInsufficientFunds, "Coin burn requirement not met")
	}
	putPrice := rewards.CalcPutPrice(domain.Treasury.AmountOf(PNYXDenom), int64(len(domain.Members)))
	if putPrice.IsPositive() && fee.AmountOf(PNYXDenom).LT(putPrice) {
		return errorsmod.Wrap(sdkerrors.ErrInsufficientFunds, "Fee below put price (eq.3)")
	}
	for _, issue := range domain.Issues {
		if issue.Name != issueName {
			continue
		}
		for _, suggestion := range issue.Suggestions {
			if suggestion.Name == suggestionName {
				return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "suggestion already exists")
			}
		}
	}
	domain.Treasury = domain.Treasury.Add(fee...)

	now := ctx.BlockTime().Unix()

	found := false
	for i, issue := range domain.Issues {
		if issue.Name == issueName {
			domain.Issues[i].Suggestions = append(domain.Issues[i].Suggestions, Suggestion{
				Name:         suggestionName,
				Creator:      creator,
				Ratings:      []Rating{},
				Stones:       0,
				Color:        "",
				DwellTime:    0,
				CreationDate: now,
				ExternalLink: externalLink,
			})
			domain.Issues[i].LastActivityAt = now
			found = true
			break
		}
	}
	if !found {
		domain.Issues = append(domain.Issues, Issue{
			Name:           issueName,
			Suggestions:    []Suggestion{{Name: suggestionName, Creator: creator, Ratings: []Rating{}, Stones: 0, Color: "", DwellTime: 0, CreationDate: now, ExternalLink: externalLink}},
			Stones:         0,
			CreationDate:   now,
			LastActivityAt: now,
		})
	}

	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	store.Set([]byte("domain:"+domainName), bz)
	return nil
}

// RateProposal records an anonymous rating on a suggestion. The caller proves
// they control a key in the domain's permission register by providing their
// domain-specific private key. The voter's avatar name is never stored —
// only the domain public key hex appears on-chain (whitepaper Section 4).
func (k Keeper) RateProposal(ctx sdk.Context, domainName, issueName, suggestionName string, rating int, domainPrivKey *ed25519.PrivKey) (sdk.Coins, map[string]interface{}, error) {
	if domainPrivKey == nil {
		return sdk.Coins{}, nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "domain private key is required for anonymous voting")
	}

	// Derive domain public key (anonymous identity).
	domainPubKeyHex := hex.EncodeToString(domainPrivKey.PubKey().Bytes())

	store := ctx.KVStore(k.StoreKey)
	domainBz := store.Get([]byte("domain:" + domainName))
	if domainBz == nil {
		return sdk.Coins{}, nil, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "Domain not found")
	}
	var domain Domain
	k.cdc.MustUnmarshalLengthPrefixed(domainBz, &domain)

	// Verify domain key is in the permission register.
	if !k.IsKeyAuthorized(ctx, domainName, domainPubKeyHex) {
		return sdk.Coins{}, nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "domain key not in permission register")
	}

	// Sign the vote payload to prove key ownership.
	payload := encodeVoteContext(ctx.ChainID(), domainName, issueName, suggestionName, &rating)
	sig, err := domainPrivKey.Sign(payload)
	if err != nil {
		return sdk.Coins{}, nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "failed to sign vote payload")
	}
	// Verify the signature (proves caller controls the key).
	if !domainPrivKey.PubKey().VerifySignature(payload, sig) {
		return sdk.Coins{}, nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "vote signature verification failed")
	}

	// Prevent double-voting with the same domain key.
	if HasDomainKeyVoted(domain, issueName, suggestionName, domainPubKeyHex) {
		return sdk.Coins{}, nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "domain key has already voted on this suggestion")
	}

	now := ctx.BlockTime().Unix()
	foundIssue := false
	foundSuggestion := false
	for i, issue := range domain.Issues {
		if issue.Name == issueName {
			foundIssue = true
			for j, suggestion := range issue.Suggestions {
				if suggestion.Name == suggestionName {
					domain.Issues[i].Suggestions[j].Ratings = append(domain.Issues[i].Suggestions[j].Ratings, Rating{
						DomainPubKeyHex: domainPubKeyHex,
						Value:           rating,
					})
					domain.Issues[i].LastActivityAt = now
					foundSuggestion = true
					break
				}
			}
			break
		}
	}
	if !foundIssue || !foundSuggestion {
		return sdk.Coins{}, nil, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "Issue or suggestion not found")
	}

	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	store.Set([]byte("domain:"+domainName), bz)

	rewardAmt := rewards.CalcReward(domain.Treasury.AmountOf(PNYXDenom))
	reward := sdk.NewCoins(sdk.NewCoin(PNYXDenom, rewardAmt))
	domain.Treasury = domain.Treasury.Sub(reward...)
	domain.TotalPayouts += rewardAmt.Int64()

	bz = k.cdc.MustMarshalLengthPrefixed(&domain)
	store.Set([]byte("domain:"+domainName), bz)

	cache := map[string]interface{}{
		"avg_rating": rating,
		"stones":     0,
		"treasury":   domain.Treasury.String(),
	}
	return reward, cache, nil
}

// RateProposalWithSignature records a rating using a pre-computed signature.
// This is the message-handler variant: the client signs the payload offline
// and submits pubkey + signature (private key never leaves the client).
func (k Keeper) RateProposalWithSignature(ctx sdk.Context, domainName, issueName, suggestionName string, rating int, domainPubKeyHex, signatureHex string) (sdk.Coins, error) {
	if rating < -5 || rating > 5 {
		return sdk.Coins{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "rating must be between -5 and +5")
	}

	// Decode and verify the public key.
	pubKeyBytes, err := hex.DecodeString(domainPubKeyHex)
	if err != nil || len(pubKeyBytes) != ed25519.PubKeySize {
		return sdk.Coins{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid domain public key hex")
	}
	pubKey := &ed25519.PubKey{Key: pubKeyBytes}

	// Decode and verify the signature.
	sigBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		return sdk.Coins{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid signature hex")
	}
	payload := encodeVoteContext(ctx.ChainID(), domainName, issueName, suggestionName, &rating)
	if !pubKey.VerifySignature(payload, sigBytes) {
		return sdk.Coins{}, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "signature verification failed")
	}

	// Verify key is in permission register.
	if !k.IsKeyAuthorized(ctx, domainName, domainPubKeyHex) {
		return sdk.Coins{}, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "domain key not in permission register")
	}

	store := ctx.KVStore(k.StoreKey)
	domainBz := store.Get([]byte("domain:" + domainName))
	if domainBz == nil {
		return sdk.Coins{}, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "domain not found")
	}
	var domain Domain
	k.cdc.MustUnmarshalLengthPrefixed(domainBz, &domain)

	// Prevent double-voting.
	if HasDomainKeyVoted(domain, issueName, suggestionName, domainPubKeyHex) {
		return sdk.Coins{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "domain key has already voted on this suggestion")
	}

	// Find and rate the suggestion.
	now := ctx.BlockTime().Unix()
	found := false
	for i, issue := range domain.Issues {
		if issue.Name == issueName {
			for j, suggestion := range issue.Suggestions {
				if suggestion.Name == suggestionName {
					domain.Issues[i].Suggestions[j].Ratings = append(domain.Issues[i].Suggestions[j].Ratings, Rating{
						DomainPubKeyHex: domainPubKeyHex,
						Value:           rating,
					})
					domain.Issues[i].LastActivityAt = now
					found = true
					break
				}
			}
			break
		}
	}
	if !found {
		return sdk.Coins{}, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "issue or suggestion not found")
	}

	// RateToEarn reward (eq.2).
	rewardAmt := rewards.CalcReward(domain.Treasury.AmountOf(PNYXDenom))
	reward := sdk.NewCoins(sdk.NewCoin(PNYXDenom, rewardAmt))
	domain.Treasury = domain.Treasury.Sub(reward...)
	domain.TotalPayouts += rewardAmt.Int64()

	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	store.Set([]byte("domain:"+domainName), bz)
	return reward, nil
}

// RateProposalWithZKP records a rating using a Groth16 ZKP membership proof.
// The voter proves membership in the domain's identity commitment set without
// revealing which commitment is theirs. A deterministic nullifier prevents
// double-voting while preserving full anonymity (WP S4 ZKP extension).
func (k Keeper) RateProposalWithZKP(ctx sdk.Context, domainName, issueName, suggestionName string, rating int, proofHex, nullifierHashHex, merkleRootHex string) (sdk.Coins, error) {
	if rating < -5 || rating > 5 {
		return sdk.Coins{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "rating must be between -5 and +5")
	}

	// Get domain and verify identity commitments exist.
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return sdk.Coins{}, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "domain %s not found", domainName)
	}
	if domain.MerkleRoot == "" {
		return sdk.Coins{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "no identity commitments registered in domain")
	}

	// Determine which Merkle root to verify against.
	effectiveRoot := domain.MerkleRoot
	if merkleRootHex != "" {
		if !isAcceptedMerkleRoot(domain, merkleRootHex) {
			return sdk.Coins{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "merkle root not recognized (not current and not in history window)")
		}
		effectiveRoot = merkleRootHex
	}
	merkleRootBytes, err := HexToFieldElement(effectiveRoot)
	if err != nil {
		return sdk.Coins{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid or non-canonical Merkle root")
	}

	// Decode and validate nullifier hash.
	nullifierBytes, err := HexToFieldElement(nullifierHashHex)
	if err != nil || len(nullifierHashHex) != 64 {
		return sdk.Coins{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "nullifier hash must be 32 bytes hex-encoded (64 hex chars)")
	}
	nullifierHashHex = hex.EncodeToString(nullifierBytes)

	// Compute external nullifier from voting context.
	externalNullifier := ComputeVoteNullifierScope(ctx.ChainID(), domainName, issueName, suggestionName)
	signalHash := ComputeVoteSignal(ctx.ChainID(), domainName, issueName, suggestionName, rating)

	// Decode proof.
	proofBytes, err := hex.DecodeString(proofHex)
	if err != nil {
		return sdk.Coins{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid proof hex encoding")
	}

	// Get or initialize the verifying key.
	vkBytes, err := k.EnsureVerifyingKey(ctx)
	if err != nil {
		return sdk.Coins{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "failed to load verifying key: "+err.Error())
	}
	vk, err := DeserializeVerifyingKey(vkBytes)
	if err != nil {
		return sdk.Coins{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "failed to deserialize verifying key")
	}

	// Verify the Groth16 membership proof.
	if err := VerifyMembershipProofForSignal(vk, proofBytes, merkleRootBytes, nullifierBytes, externalNullifier, signalHash); err != nil {
		return sdk.Coins{}, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "ZKP membership proof verification failed: "+err.Error())
	}

	// Check nullifier has not been used (prevents double-voting).
	if k.IsNullifierUsed(ctx, domainName, nullifierHashHex) {
		return sdk.Coins{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "nullifier already used (double-vote prevented)")
	}

	// Find and rate the suggestion.
	now := ctx.BlockTime().Unix()
	suggestionFound := false
	for i, issue := range domain.Issues {
		if issue.Name == issueName {
			for j, suggestion := range issue.Suggestions {
				if suggestion.Name == suggestionName {
					domain.Issues[i].Suggestions[j].Ratings = append(domain.Issues[i].Suggestions[j].Ratings, Rating{
						NullifierHex: nullifierHashHex,
						Value:        rating,
					})
					domain.Issues[i].LastActivityAt = now
					suggestionFound = true
					break
				}
			}
			break
		}
	}
	if !suggestionFound {
		return sdk.Coins{}, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "issue or suggestion not found")
	}

	// Mark nullifier as used.
	k.SetNullifierUsed(ctx, domainName, nullifierHashHex, ctx.BlockHeight())

	// RateToEarn reward (eq.2).
	rewardAmt := rewards.CalcReward(domain.Treasury.AmountOf(PNYXDenom))
	reward := sdk.NewCoins(sdk.NewCoin(PNYXDenom, rewardAmt))
	domain.Treasury = domain.Treasury.Sub(reward...)
	domain.TotalPayouts += rewardAmt.Int64()

	store := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	store.Set([]byte("domain:"+domainName), bz)
	return reward, nil
}
