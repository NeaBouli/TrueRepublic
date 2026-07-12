package truedemocracy

import (
	"encoding/hex"

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

// RemoveValidatorWithEscrow is a full authenticated withdrawal. It reuses the
// transfer-limit and accounting checks applied by WithdrawStake.
func (k Keeper) RemoveValidatorWithEscrow(ctx sdk.Context, sender sdk.AccAddress, operatorAddr string) error {
	if err := requireSignerClaim(sender, operatorAddr, "operator address"); err != nil {
		return err
	}
	validator, found := k.GetValidator(ctx, operatorAddr)
	if !found {
		return errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "validator not found")
	}
	amount := validator.Stake.AmountOf(PNYXDenom)
	if !amount.IsPositive() || !amount.IsInt64() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "validator stake is invalid")
	}
	return k.WithdrawStakeWithEscrow(ctx, sender, operatorAddr, amount.Int64())
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

// executeDeferredAnonymousReward records an anonymous rating but restores the
// calculated reward to treasury. Current legacy signatures and ZKP public
// inputs do not bind a bank recipient, so paying msg.Sender would be
// front-runnable. GH-13/GH-7 must add a recipient-bound claim before payout.
func (k Keeper) executeDeferredAnonymousReward(
	ctx sdk.Context,
	domainName string,
	action rewardAction,
) (sdk.Coins, error) {
	cacheCtx, write := ctx.CacheContext()
	reward, err := action(cacheCtx)
	if err != nil {
		return nil, err
	}
	if !reward.Empty() {
		if err := validatePNYXCoins(reward, "deferred reward"); err != nil {
			return nil, err
		}
		domain, found := k.GetDomain(cacheCtx, domainName)
		if !found {
			return nil, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "domain disappeared while deferring reward")
		}
		rewardAmount := reward.AmountOf(PNYXDenom)
		if !rewardAmount.IsInt64() || domain.TotalPayouts < rewardAmount.Int64() {
			return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "invalid deferred reward accounting")
		}
		domain.Treasury = domain.Treasury.Add(reward...)
		domain.TotalPayouts -= rewardAmount.Int64()
		store := cacheCtx.KVStore(k.StoreKey)
		store.Set([]byte("domain:"+domainName), k.cdc.MustMarshalLengthPrefixed(&domain))
	}
	write()
	return sdk.Coins{}, nil
}

func (k Keeper) RateProposalWithSignatureDeferredReward(
	ctx sdk.Context,
	domainName, issueName, suggestionName string,
	rating int,
	domainPubKeyHex, signatureHex string,
) (sdk.Coins, error) {
	return k.executeDeferredAnonymousReward(ctx, domainName, func(cacheCtx sdk.Context) (sdk.Coins, error) {
		return k.RateProposalWithSignature(
			cacheCtx,
			domainName,
			issueName,
			suggestionName,
			rating,
			domainPubKeyHex,
			signatureHex,
		)
	})
}

func (k Keeper) RateProposalWithZKPDeferredReward(
	ctx sdk.Context,
	domainName, issueName, suggestionName string,
	rating int,
	proofHex, nullifierHashHex, merkleRootHex string,
) (sdk.Coins, error) {
	return k.executeDeferredAnonymousReward(ctx, domainName, func(cacheCtx sdk.Context) (sdk.Coins, error) {
		return k.RateProposalWithZKP(
			cacheCtx,
			domainName,
			issueName,
			suggestionName,
			rating,
			proofHex,
			nullifierHashHex,
			merkleRootHex,
		)
	})
}

// EscrowClaims returns the aggregate upnyx claims held in domain treasuries
// and validator stake records. Reward issuance must fund this same escrow.
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
	return claims
}

func (k Keeper) ValidateEscrowParity(ctx sdk.Context) error {
	if err := requireBankKeeper(k.bankKeeper); err != nil {
		return err
	}
	moduleAddress := authtypes.NewModuleAddress(ModuleName)
	bankBalance := k.bankKeeper.GetBalance(ctx, moduleAddress, PNYXDenom).Amount
	claims := k.EscrowClaims(ctx)
	if !bankBalance.Equal(claims) {
		return errorsmod.Wrapf(
			sdkerrors.ErrLogic,
			"escrow mismatch: bank=%s%s claims=%s%s",
			bankBalance,
			PNYXDenom,
			claims,
			PNYXDenom,
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
