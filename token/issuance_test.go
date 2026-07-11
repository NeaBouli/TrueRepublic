package token

import (
	"context"
	"fmt"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type issuanceBank struct {
	supply  math.Int
	modules map[string]math.Int
	fail    bool
}

func newIssuanceBank(supply math.Int) *issuanceBank {
	return &issuanceBank{supply: supply, modules: make(map[string]math.Int)}
}

func (bank *issuanceBank) GetSupply(context.Context, string) sdk.Coin {
	return NewCoin(bank.supply)
}

func (bank *issuanceBank) MintCoins(_ context.Context, moduleName string, amounts sdk.Coins) error {
	if bank.fail {
		return fmt.Errorf("injected mint failure")
	}
	amount := amounts.AmountOf(BaseDenom)
	bank.supply = bank.supply.Add(amount)
	moduleBalance := bank.modules[moduleName]
	if moduleBalance.IsNil() {
		moduleBalance = math.ZeroInt()
	}
	bank.modules[moduleName] = moduleBalance.Add(amount)
	return nil
}

func (bank *issuanceBank) BurnCoins(_ context.Context, moduleName string, amounts sdk.Coins) error {
	if bank.fail {
		return fmt.Errorf("injected burn failure")
	}
	amount := amounts.AmountOf(BaseDenom)
	moduleBalance := bank.modules[moduleName]
	if moduleBalance.IsNil() {
		moduleBalance = math.ZeroInt()
	}
	if moduleBalance.LT(amount) {
		return fmt.Errorf("insufficient module balance")
	}
	bank.modules[moduleName] = moduleBalance.Sub(amount)
	bank.supply = bank.supply.Sub(amount)
	return nil
}

func TestIssuanceServiceCapsAggregateMints(t *testing.T) {
	bank := newIssuanceBank(MaxSupply().SubRaw(3))
	service := NewIssuanceService(bank, "rewards")

	first, err := service.MintUpToCap(context.Background(), math.NewInt(2))
	if err != nil || !first.Equal(math.NewInt(2)) {
		t.Fatalf("first mint = %s, %v", first, err)
	}
	second, err := service.MintUpToCap(context.Background(), math.NewInt(5))
	if err != nil || !second.Equal(math.OneInt()) {
		t.Fatalf("second mint = %s, %v; want 1", second, err)
	}
	third, err := service.MintUpToCap(context.Background(), math.OneInt())
	if err != nil || !third.IsZero() {
		t.Fatalf("mint at cap = %s, %v; want zero", third, err)
	}
	if !bank.supply.Equal(MaxSupply()) {
		t.Fatalf("supply = %s, want cap %s", bank.supply, MaxSupply())
	}
}

func TestIssuanceServiceBurnChangesCanonicalSupply(t *testing.T) {
	bank := newIssuanceBank(MaxSupply())
	bank.modules["rewards"] = math.NewInt(10)
	service := NewIssuanceService(bank, "rewards")

	if err := service.Burn(context.Background(), math.NewInt(4)); err != nil {
		t.Fatal(err)
	}
	if !bank.supply.Equal(MaxSupply().SubRaw(4)) {
		t.Fatalf("supply after burn = %s", bank.supply)
	}
	minted, err := service.MintUpToCap(context.Background(), math.NewInt(10))
	if err != nil || !minted.Equal(math.NewInt(4)) {
		t.Fatalf("remint after burn = %s, %v; want 4", minted, err)
	}
}

func TestIssuanceServiceRejectsInvalidAndFailedOperations(t *testing.T) {
	bank := newIssuanceBank(math.ZeroInt())
	service := NewIssuanceService(bank, "rewards")
	if _, err := service.MintUpToCap(context.Background(), math.NewInt(-1)); err == nil {
		t.Fatal("expected negative mint rejection")
	}
	if err := service.Burn(context.Background(), math.ZeroInt()); err == nil {
		t.Fatal("expected zero burn rejection")
	}
	if _, err := NewIssuanceService(bank, "").MintUpToCap(context.Background(), math.OneInt()); err == nil {
		t.Fatal("expected empty module rejection")
	}
	bank.fail = true
	if _, err := service.MintUpToCap(context.Background(), math.OneInt()); err == nil {
		t.Fatal("expected bank mint failure")
	}
	overCap := newIssuanceBank(MaxSupply().AddRaw(1))
	if _, err := NewIssuanceService(overCap, "rewards").MintUpToCap(context.Background(), math.OneInt()); err == nil {
		t.Fatal("expected over-cap canonical supply rejection")
	}
	if _, err := NewIssuanceService(overCap, "rewards").MintUpToCap(context.Background(), math.ZeroInt()); err == nil {
		t.Fatal("expected over-cap rejection even without a new reward")
	}
}
