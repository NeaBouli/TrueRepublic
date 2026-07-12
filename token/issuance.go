package token

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IssuanceBankKeeper is the canonical bank surface used for PNYX supply
// changes. Modules must not mutate supply through any other path.
type IssuanceBankKeeper interface {
	GetSupply(ctx context.Context, denom string) sdk.Coin
	MintCoins(ctx context.Context, moduleName string, amounts sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amounts sdk.Coins) error
}

// IssuanceService applies the global PNYX cap to one authorized module account.
// Cosmos executes state transitions serially, so reading supply and minting in
// the same cached context makes the remaining-cap check atomic.
type IssuanceService struct {
	bank       IssuanceBankKeeper
	moduleName string
}

func NewIssuanceService(bank IssuanceBankKeeper, moduleName string) IssuanceService {
	return IssuanceService{bank: bank, moduleName: moduleName}
}

func (s IssuanceService) Supply(ctx context.Context) (math.Int, error) {
	if s.bank == nil {
		return math.Int{}, fmt.Errorf("issuance bank keeper is not available")
	}
	supply := s.bank.GetSupply(ctx, BaseDenom).Amount
	if supply.IsNil() || supply.IsNegative() {
		return math.Int{}, fmt.Errorf("canonical PNYX supply must be valid and non-negative")
	}
	return supply, nil
}

// MintUpToCap mints min(requested, remaining supply capacity) into the
// configured module escrow. Reaching the cap returns zero instead of halting
// consensus; negative requests are rejected.
func (s IssuanceService) MintUpToCap(ctx context.Context, requested math.Int) (math.Int, error) {
	if requested.IsNil() || requested.IsNegative() {
		return math.Int{}, fmt.Errorf("mint amount must be valid and non-negative")
	}
	if s.moduleName == "" {
		return math.Int{}, fmt.Errorf("issuance module name is required")
	}

	supply, err := s.Supply(ctx)
	if err != nil {
		return math.Int{}, err
	}
	if supply.GT(MaxSupply()) {
		return math.Int{}, fmt.Errorf("canonical PNYX supply %s exceeds cap %s", supply, MaxSupply())
	}
	if requested.IsZero() {
		return math.ZeroInt(), nil
	}
	remaining := MaxSupply().Sub(supply)
	if !remaining.IsPositive() {
		return math.ZeroInt(), nil
	}
	minted := requested
	if minted.GT(remaining) {
		minted = remaining
	}
	if err := s.bank.MintCoins(ctx, s.moduleName, sdk.NewCoins(NewCoin(minted))); err != nil {
		return math.Int{}, fmt.Errorf("mint PNYX into %s escrow: %w", s.moduleName, err)
	}
	return minted, nil
}

// Burn removes exact PNYX base units from the configured module account and
// therefore from canonical bank supply.
func (s IssuanceService) Burn(ctx context.Context, amount math.Int) error {
	if amount.IsNil() || !amount.IsPositive() {
		return fmt.Errorf("burn amount must be positive")
	}
	if s.bank == nil {
		return fmt.Errorf("issuance bank keeper is not available")
	}
	if s.moduleName == "" {
		return fmt.Errorf("issuance module name is required")
	}
	if err := s.bank.BurnCoins(ctx, s.moduleName, sdk.NewCoins(NewCoin(amount))); err != nil {
		return fmt.Errorf("burn PNYX from %s escrow: %w", s.moduleName, err)
	}
	return nil
}
