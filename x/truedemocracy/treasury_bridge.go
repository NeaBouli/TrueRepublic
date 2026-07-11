package truedemocracy

// Treasury bridge: connects x/bank user accounts with Domain.Treasury
// (custom accounting). Deposits move PNYX from a user's bank balance into
// the truedemocracy module account and increment Domain.Treasury. Withdrawals
// do the reverse (admin authorization required).

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DepositToDomain transfers PNYX from a user's bank account to a domain's
// treasury. The coins move: user → truedemocracy module account (via x/bank),
// and Domain.Treasury is incremented by the same amount.
func (k Keeper) DepositToDomain(ctx sdk.Context, depositor sdk.AccAddress, domainName string, amount sdk.Coin) error {
	if k.bankKeeper == nil {
		return errorsmod.Wrap(sdkerrors.ErrLogic, "bank keeper not available")
	}

	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrNotFound, "domain %s not found", domainName)
	}

	if !amount.IsPositive() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "amount must be positive")
	}
	if amount.Denom != PNYXDenom {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "only upnyx deposits supported, got %s", amount.Denom)
	}

	cacheCtx, write := ctx.CacheContext()

	// Transfer from user account to module account.
	coins := sdk.NewCoins(amount)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(cacheCtx, depositor, ModuleName, coins); err != nil {
		return errorsmod.Wrap(err, "bank transfer failed")
	}

	// Credit domain treasury.
	domain.Treasury = domain.Treasury.Add(amount)

	store := cacheCtx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	store.Set([]byte("domain:"+domainName), bz)

	cacheCtx.EventManager().EmitEvent(sdk.NewEvent(
		"domain_deposit",
		sdk.NewAttribute("domain", domainName),
		sdk.NewAttribute("depositor", depositor.String()),
		sdk.NewAttribute("amount", amount.String()),
	))

	write()
	return nil
}

// WithdrawFromDomain transfers PNYX from a domain's treasury to a recipient's
// bank account. Only the domain admin may authorize withdrawals.
// The coins move: truedemocracy module account → recipient (via x/bank),
// and Domain.Treasury is decremented.
func (k Keeper) WithdrawFromDomain(ctx sdk.Context, domainName string, recipient sdk.AccAddress, amount sdk.Coin, authorizer sdk.AccAddress) error {
	if k.bankKeeper == nil {
		return errorsmod.Wrap(sdkerrors.ErrLogic, "bank keeper not available")
	}

	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrNotFound, "domain %s not found", domainName)
	}

	// Only admin can withdraw.
	if !authorizer.Equals(domain.Admin) {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "only domain admin can withdraw from treasury")
	}

	if !amount.IsPositive() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "amount must be positive")
	}
	if amount.Denom != PNYXDenom {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "only upnyx withdrawals supported, got %s", amount.Denom)
	}

	// Check sufficient treasury balance.
	if domain.Treasury.AmountOf(PNYXDenom).LT(amount.Amount) {
		return errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds,
			"domain treasury has %s upnyx, requested %s",
			domain.Treasury.AmountOf(PNYXDenom), amount.Amount)
	}

	cacheCtx, write := ctx.CacheContext()

	// Debit domain treasury first in the transaction cache.
	domain.Treasury = domain.Treasury.Sub(amount)

	// Transfer from module account to recipient.
	coins := sdk.NewCoins(amount)
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(cacheCtx, ModuleName, recipient, coins); err != nil {
		return errorsmod.Wrap(err, "bank transfer failed")
	}

	store := cacheCtx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	store.Set([]byte("domain:"+domainName), bz)

	cacheCtx.EventManager().EmitEvent(sdk.NewEvent(
		"domain_withdrawal",
		sdk.NewAttribute("domain", domainName),
		sdk.NewAttribute("recipient", recipient.String()),
		sdk.NewAttribute("amount", amount.String()),
		sdk.NewAttribute("authorizer", authorizer.String()),
	))

	write()
	return nil
}
