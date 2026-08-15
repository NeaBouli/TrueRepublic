package truedemocracy

import (
	"encoding/hex"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func validatePNYXCoins(coins sdk.Coins, field string) error {
	if !coins.IsValid() || len(coins) != 1 || coins[0].Denom != PNYXDenom || !coins[0].Amount.IsPositive() {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"%s must contain exactly one positive %s coin",
			field,
			PNYXDenom,
		)
	}
	return nil
}

func requireBankKeeper(bankKeeper BankKeeper) error {
	if bankKeeper == nil {
		return errorsmod.Wrap(sdkerrors.ErrLogic, "bank keeper not available")
	}
	return nil
}

func requireSignerClaim(sender sdk.AccAddress, claimed, field string) error {
	if sender.Empty() || claimed != sender.String() {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s must match the authenticated sender", field)
	}
	return nil
}

// CreateDomainWithEscrow atomically creates a domain and moves its declared
// treasury from the authenticated admin into the module escrow account.
func (k Keeper) CreateDomainWithEscrow(ctx sdk.Context, name string, admin sdk.AccAddress, initialCoins sdk.Coins) error {
	if err := requireBankKeeper(k.bankKeeper); err != nil {
		return err
	}
	if name == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "domain name is required")
	}
	if name == ReservedGovernanceDomain {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "domain %s is reserved and can only be anchored in genesis", ReservedGovernanceDomain)
	}
	if admin.Empty() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "admin address is required")
	}
	if err := validatePNYXCoins(initialCoins, "initial coins"); err != nil {
		return err
	}
	if _, found := k.GetDomain(ctx, name); found {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "domain %s already exists", name)
	}

	cacheCtx, write := ctx.CacheContext()
	k.CreateDomain(cacheCtx, name, admin, initialCoins)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(cacheCtx, admin, ModuleName, initialCoins); err != nil {
		return errorsmod.Wrap(err, "initial treasury escrow transfer failed")
	}
	write()
	return nil
}

// SubmitProposalWithEscrow derives the creator from the signer and atomically
// escrows the exact proposal fee before committing the proposal state.
func (k Keeper) SubmitProposalWithEscrow(
	ctx sdk.Context,
	sender sdk.AccAddress,
	creator string,
	domainName, issueName, suggestionName string,
	fee sdk.Coins,
	externalLink string,
) error {
	if err := requireBankKeeper(k.bankKeeper); err != nil {
		return err
	}
	if err := validatePNYXCoins(fee, "proposal fee"); err != nil {
		return err
	}
	if err := requireSignerClaim(sender, creator, "creator"); err != nil {
		return err
	}

	cacheCtx, write := ctx.CacheContext()
	if err := k.SubmitProposal(
		cacheCtx,
		domainName,
		issueName,
		suggestionName,
		creator,
		fee,
		externalLink,
	); err != nil {
		return err
	}
	if err := k.bankKeeper.SendCoinsFromAccountToModule(cacheCtx, sender, ModuleName, fee); err != nil {
		return errorsmod.Wrap(err, "proposal fee escrow transfer failed")
	}
	write()
	return nil
}

// RegisterValidatorWithEscrow derives the operator from the signer and backs
// the full internal stake claim with coins held by the module account.
func (k Keeper) RegisterValidatorWithEscrow(
	ctx sdk.Context,
	sender sdk.AccAddress,
	operatorAddr string,
	pubKeyBytes []byte,
	stake sdk.Coins,
	domainName string,
) error {
	if err := requireBankKeeper(k.bankKeeper); err != nil {
		return err
	}
	if err := requireSignerClaim(sender, operatorAddr, "operator address"); err != nil {
		return err
	}
	if err := validatePNYXCoins(stake, "validator stake"); err != nil {
		return err
	}

	cacheCtx, write := ctx.CacheContext()
	if err := k.RegisterValidator(cacheCtx, operatorAddr, pubKeyBytes, stake, domainName); err != nil {
		return err
	}
	if err := k.bankKeeper.SendCoinsFromAccountToModule(cacheCtx, sender, ModuleName, stake); err != nil {
		return errorsmod.Wrap(err, "validator stake escrow transfer failed")
	}
	write()
	return nil
}

// WithdrawStakeWithEscrow atomically reduces an authenticated operator's stake
// claim and returns the exact amount from module escrow.
func (k Keeper) WithdrawStakeWithEscrow(ctx sdk.Context, sender sdk.AccAddress, operatorAddr string, amount int64) error {
	if err := requireBankKeeper(k.bankKeeper); err != nil {
		return err
	}
	if err := requireSignerClaim(sender, operatorAddr, "operator address"); err != nil {
		return err
	}
	if amount <= 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "withdrawal amount must be positive")
	}
	validator, found := k.GetValidator(ctx, operatorAddr)
	if found {
		stake := validator.Stake.AmountOf(PNYXDenom)
		if stake.IsInt64() && stake.Int64() == amount {
			return k.RemoveValidatorWithEscrow(ctx, sender, operatorAddr)
		}
		return errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"partial validator withdrawals are disabled until slashable unbonding is implemented; use a full validator exit",
		)
	}

	cacheCtx, write := ctx.CacheContext()
	if err := k.WithdrawStake(cacheCtx, operatorAddr, amount); err != nil {
		return err
	}
	coins := sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, amount))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(cacheCtx, ModuleName, sender, coins); err != nil {
		return errorsmod.Wrap(err, "validator stake escrow withdrawal failed")
	}
	write()
	return nil
}

// RemoveValidatorWithEscrow is a full authenticated exit. It atomically
// removes validator power but retains the stake in module escrow until the
// CometBFT evidence window has expired.
func (k Keeper) RemoveValidatorWithEscrow(ctx sdk.Context, sender sdk.AccAddress, operatorAddr string) error {
	if err := requireBankKeeper(k.bankKeeper); err != nil {
		return err
	}
	if err := requireSignerClaim(sender, operatorAddr, "operator address"); err != nil {
		return err
	}
	validator, found := k.GetValidator(ctx, operatorAddr)
	if !found {
		return errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "validator not found")
	}
	if _, found := k.GetPendingValidatorRemoval(ctx, operatorAddr); found {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "validator removal is already pending")
	}
	amount := validator.Stake.AmountOf(PNYXDenom)
	if !amount.IsPositive() || !amount.IsInt64() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "validator stake is invalid")
	}

	removal, err := newPendingValidatorRemoval(ctx, validator, sender.String())
	if err != nil {
		return err
	}

	cacheCtx, write := ctx.CacheContext()
	if err := k.WithdrawStake(cacheCtx, operatorAddr, amount.Int64()); err != nil {
		return err
	}
	k.SetPendingValidatorRemoval(cacheCtx, removal)
	write()
	return nil
}

type rewardAction func(ctx sdk.Context) (sdk.Coins, error)

func (k Keeper) executeRewardPayout(ctx sdk.Context, recipient sdk.AccAddress, action rewardAction) (sdk.Coins, error) {
	if err := requireBankKeeper(k.bankKeeper); err != nil {
		return nil, err
	}

	cacheCtx, write := ctx.CacheContext()
	reward, err := action(cacheCtx)
	if err != nil {
		return nil, err
	}
	if !reward.Empty() {
		if err := validatePNYXCoins(reward, "reward"); err != nil {
			return nil, err
		}
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(cacheCtx, ModuleName, recipient, reward); err != nil {
			return nil, errorsmod.Wrap(err, "treasury reward payout failed")
		}
	}
	write()
	return reward, nil
}

func (k Keeper) PlaceStoneOnIssueWithPayout(
	ctx sdk.Context,
	sender sdk.AccAddress,
	domainName, issueName, memberAddr string,
) (sdk.Coins, error) {
	if err := requireSignerClaim(sender, memberAddr, "member address"); err != nil {
		return nil, err
	}
	return k.executeRewardPayout(ctx, sender, func(cacheCtx sdk.Context) (sdk.Coins, error) {
		return k.PlaceStoneOnIssue(cacheCtx, domainName, issueName, memberAddr)
	})
}

func (k Keeper) PlaceStoneOnSuggestionWithPayout(
	ctx sdk.Context,
	sender sdk.AccAddress,
	domainName, issueName, suggestionName, memberAddr string,
) (sdk.Coins, error) {
	if err := requireSignerClaim(sender, memberAddr, "member address"); err != nil {
		return nil, err
	}
	return k.executeRewardPayout(ctx, sender, func(cacheCtx sdk.Context) (sdk.Coins, error) {
		return k.PlaceStoneOnSuggestion(cacheCtx, domainName, issueName, suggestionName, memberAddr)
	})
}

// executeDeferredAnonymousReward was the GH-13/GH-7 stopgap: it recorded the
// anonymous rating but restored the reward to treasury because neither path
// bound a bank recipient. GH-209 replaces it with an atomic treasury-funded
// payout to the proof- or signature-bound recipient below.

// ValidateRewardRecipient decodes and fail-closed validates the canonical
// bech32 reward recipient bound into the v2 anonymous-rating payload. The
// recipient must be non-empty, use the configured account prefix, and match
// its own canonical lowercase re-encoding exactly.
func ValidateRewardRecipient(recipient string) (sdk.AccAddress, error) {
	if recipient == "" {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "reward recipient is required")
	}
	if recipient != strings.ToLower(recipient) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "reward recipient must be canonical lowercase bech32")
	}
	prefix := sdk.GetConfig().GetBech32AccountAddrPrefix()
	if !strings.HasPrefix(recipient, prefix+"1") {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "reward recipient must use the %s account prefix", prefix)
	}
	addr, err := sdk.AccAddressFromBech32(recipient)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "reward recipient is not valid bech32")
	}
	if addr.String() != recipient {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "reward recipient is not in canonical bech32 form")
	}
	return addr, nil
}

// validatePayoutRecipient additionally rejects blocked module accounts so a
// reward can never be paid into an address the bank layer forbids receiving.
func (k Keeper) validatePayoutRecipient(recipient string) (sdk.AccAddress, error) {
	if err := requireBankKeeper(k.bankKeeper); err != nil {
		return nil, err
	}
	addr, err := ValidateRewardRecipient(recipient)
	if err != nil {
		return nil, err
	}
	if k.bankKeeper.BlockedAddr(addr) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "reward recipient must not be a blocked module account")
	}
	return addr, nil
}

// RateProposalWithSignaturePayout records a domain-key anonymous rating whose
// signature covers the recipient-bound v2 payload and atomically pays the
// treasury reward to exactly that bound recipient. Any validation, signature,
// or bank-send failure leaves rating, treasury, and escrow state unchanged.
func (k Keeper) RateProposalWithSignaturePayout(
	ctx sdk.Context,
	domainName, issueName, suggestionName string,
	rating int,
	domainPubKeyHex, signatureHex string,
	rewardRecipient string,
) (sdk.Coins, error) {
	recipient, err := k.validatePayoutRecipient(rewardRecipient)
	if err != nil {
		return nil, err
	}
	return k.executeRewardPayout(ctx, recipient, func(cacheCtx sdk.Context) (sdk.Coins, error) {
		return k.RateProposalWithSignature(
			cacheCtx,
			domainName,
			issueName,
			suggestionName,
			rating,
			domainPubKeyHex,
			signatureHex,
			rewardRecipient,
		)
	})
}

// RateProposalWithZKPPayout records a Groth16 anonymous rating whose public
// SignalHash covers the recipient-bound v2 payload and atomically pays the
// treasury reward to exactly that bound recipient. Any validation, proof, or
// bank-send failure leaves rating, nullifier, treasury, and escrow state
// unchanged. The reward never goes to msg.Sender for submitting the
// transaction.
func (k Keeper) RateProposalWithZKPPayout(
	ctx sdk.Context,
	domainName, issueName, suggestionName string,
	rating int,
	proofHex, nullifierHashHex, merkleRootHex string,
	rewardRecipient string,
) (sdk.Coins, error) {
	recipient, err := k.validatePayoutRecipient(rewardRecipient)
	if err != nil {
		return nil, err
	}
	return k.executeRewardPayout(ctx, recipient, func(cacheCtx sdk.Context) (sdk.Coins, error) {
		return k.RateProposalWithZKP(
			cacheCtx,
			domainName,
			issueName,
			suggestionName,
			rating,
			proofHex,
			nullifierHashHex,
			merkleRootHex,
			rewardRecipient,
		)
	})
}

// EscrowClaims returns the aggregate upnyx claims held in domain treasuries,
// active validator stake records, and evidence-window exit holds. Reward
// issuance must fund this same escrow.
func (k Keeper) EscrowClaims(ctx sdk.Context) math.Int {
	claims := math.ZeroInt()
	k.IterateDomains(ctx, func(domain Domain) bool {
		claims = claims.Add(domain.Treasury.AmountOf(PNYXDenom))
		return false
	})
	k.IterateValidators(ctx, func(validator Validator) bool {
		claims = claims.Add(validator.Stake.AmountOf(PNYXDenom))
		return false
	})
	k.IteratePendingValidatorRemovals(ctx, func(removal PendingValidatorRemoval) bool {
		claims = claims.Add(removal.Validator.Stake.AmountOf(PNYXDenom))
		return false
	})
	return claims
}

func (k Keeper) ValidateEscrowParity(ctx sdk.Context) error {
	if err := requireBankKeeper(k.bankKeeper); err != nil {
		return err
	}
	moduleAddress := authtypes.NewModuleAddress(ModuleName)
	claims := k.EscrowClaims(ctx)
	expected := sdk.NewCoins()
	if claims.IsPositive() {
		expected = sdk.NewCoins(sdk.NewCoin(PNYXDenom, claims))
	}
	bankBalance := k.bankKeeper.GetAllBalances(ctx, moduleAddress)
	if !bankBalance.Equal(expected) {
		return errorsmod.Wrapf(
			sdkerrors.ErrLogic,
			"escrow mismatch: bank=%s claims=%s",
			bankBalance,
			expected,
		)
	}
	return nil
}

func decodeValidatorPubKey(pubKey string) ([]byte, error) {
	pubKeyBytes, err := hex.DecodeString(pubKey)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid hex-encoded public key")
	}
	return pubKeyBytes, nil
}
