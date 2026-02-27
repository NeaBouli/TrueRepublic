package truedemocracy

import (
	"context"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// --- Mock BankKeeper ---

// mockBankKeeper implements BankKeeper for testing treasury bridge operations.
type mockBankKeeper struct {
	accounts map[string]sdk.Coins // address → balances
	modules  map[string]sdk.Coins // module name → balances
}

func newMockBankKeeper() *mockBankKeeper {
	return &mockBankKeeper{
		accounts: make(map[string]sdk.Coins),
		modules:  make(map[string]sdk.Coins),
	}
}

func (m *mockBankKeeper) fundAccount(addr sdk.AccAddress, coins sdk.Coins) {
	m.accounts[addr.String()] = m.accounts[addr.String()].Add(coins...)
}

func (m *mockBankKeeper) fundModule(moduleName string, coins sdk.Coins) {
	m.modules[moduleName] = m.modules[moduleName].Add(coins...)
}

func (m *mockBankKeeper) SendCoinsFromAccountToModule(_ context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	key := senderAddr.String()
	bal := m.accounts[key]
	for _, coin := range amt {
		if bal.AmountOf(coin.Denom).LT(coin.Amount) {
			return fmt.Errorf("insufficient funds: %s < %s", bal.AmountOf(coin.Denom), coin.Amount)
		}
	}
	m.accounts[key] = bal.Sub(amt...)
	m.modules[recipientModule] = m.modules[recipientModule].Add(amt...)
	return nil
}

func (m *mockBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	bal := m.modules[senderModule]
	for _, coin := range amt {
		if bal.AmountOf(coin.Denom).LT(coin.Amount) {
			return fmt.Errorf("insufficient module funds: %s < %s", bal.AmountOf(coin.Denom), coin.Amount)
		}
	}
	m.modules[senderModule] = bal.Sub(amt...)
	key := recipientAddr.String()
	m.accounts[key] = m.accounts[key].Add(amt...)
	return nil
}

// setupKeeperWithBank creates a Keeper with a mock BankKeeper for bridge tests.
func setupKeeperWithBank(t *testing.T) (Keeper, sdk.Context, *mockBankKeeper) {
	t.Helper()
	k, ctx := setupKeeper(t) // from validator_test.go (bankKeeper=nil)
	bk := newMockBankKeeper()
	k.bankKeeper = bk
	return k, ctx, bk
}

// --- Deposit Tests ---

func TestDepositToDomain(t *testing.T) {
	k, ctx, bk := setupKeeperWithBank(t)

	admin := sdk.AccAddress("admin1")
	user := sdk.AccAddress("user-alice")
	k.CreateDomain(ctx, "TestDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 1000)))

	// Fund user account.
	bk.fundAccount(user, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500)))

	t.Run("success", func(t *testing.T) {
		err := k.DepositToDomain(ctx, user, "TestDomain", sdk.NewInt64Coin("pnyx", 100))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// User balance decreased.
		userBal := bk.accounts[user.String()]
		if userBal.AmountOf("pnyx").Int64() != 400 {
			t.Errorf("user balance = %d, want 400", userBal.AmountOf("pnyx").Int64())
		}

		// Module account received coins.
		modBal := bk.modules[ModuleName]
		if modBal.AmountOf("pnyx").Int64() != 100 {
			t.Errorf("module balance = %d, want 100", modBal.AmountOf("pnyx").Int64())
		}

		// Domain treasury increased.
		domain, _ := k.GetDomain(ctx, "TestDomain")
		if domain.Treasury.AmountOf("pnyx").Int64() != 1100 {
			t.Errorf("treasury = %d, want 1100", domain.Treasury.AmountOf("pnyx").Int64())
		}
	})

	t.Run("domain not found", func(t *testing.T) {
		err := k.DepositToDomain(ctx, user, "NoSuchDomain", sdk.NewInt64Coin("pnyx", 100))
		if err == nil {
			t.Fatal("expected error for missing domain")
		}
	})

	t.Run("insufficient funds", func(t *testing.T) {
		err := k.DepositToDomain(ctx, user, "TestDomain", sdk.NewInt64Coin("pnyx", 9999))
		if err == nil {
			t.Fatal("expected error for insufficient funds")
		}
	})

	t.Run("wrong denom", func(t *testing.T) {
		err := k.DepositToDomain(ctx, user, "TestDomain", sdk.NewInt64Coin("atom", 10))
		if err == nil {
			t.Fatal("expected error for wrong denom")
		}
	})

	t.Run("zero amount", func(t *testing.T) {
		err := k.DepositToDomain(ctx, user, "TestDomain", sdk.NewInt64Coin("pnyx", 0))
		if err == nil {
			t.Fatal("expected error for zero amount")
		}
	})
}

// --- Withdraw Tests ---

func TestWithdrawFromDomain(t *testing.T) {
	k, ctx, bk := setupKeeperWithBank(t)

	admin := sdk.AccAddress("admin1")
	recipient := sdk.AccAddress("recipient1")
	k.CreateDomain(ctx, "WithdrawDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 5000)))

	// Fund the module account to back the treasury.
	bk.fundModule(ModuleName, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 5000)))

	t.Run("success", func(t *testing.T) {
		err := k.WithdrawFromDomain(ctx, "WithdrawDomain", recipient, sdk.NewInt64Coin("pnyx", 1000), admin)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Recipient received coins.
		recipientBal := bk.accounts[recipient.String()]
		if recipientBal.AmountOf("pnyx").Int64() != 1000 {
			t.Errorf("recipient balance = %d, want 1000", recipientBal.AmountOf("pnyx").Int64())
		}

		// Domain treasury decreased.
		domain, _ := k.GetDomain(ctx, "WithdrawDomain")
		if domain.Treasury.AmountOf("pnyx").Int64() != 4000 {
			t.Errorf("treasury = %d, want 4000", domain.Treasury.AmountOf("pnyx").Int64())
		}

		// Module account debited.
		modBal := bk.modules[ModuleName]
		if modBal.AmountOf("pnyx").Int64() != 4000 {
			t.Errorf("module balance = %d, want 4000", modBal.AmountOf("pnyx").Int64())
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		randomUser := sdk.AccAddress("random-user")
		err := k.WithdrawFromDomain(ctx, "WithdrawDomain", recipient, sdk.NewInt64Coin("pnyx", 100), randomUser)
		if err == nil {
			t.Fatal("expected error for unauthorized withdraw")
		}
	})

	t.Run("insufficient treasury", func(t *testing.T) {
		err := k.WithdrawFromDomain(ctx, "WithdrawDomain", recipient, sdk.NewInt64Coin("pnyx", 999999), admin)
		if err == nil {
			t.Fatal("expected error for insufficient treasury")
		}
	})

	t.Run("domain not found", func(t *testing.T) {
		err := k.WithdrawFromDomain(ctx, "NoSuchDomain", recipient, sdk.NewInt64Coin("pnyx", 100), admin)
		if err == nil {
			t.Fatal("expected error for missing domain")
		}
	})

	t.Run("wrong denom", func(t *testing.T) {
		err := k.WithdrawFromDomain(ctx, "WithdrawDomain", recipient, sdk.NewInt64Coin("atom", 10), admin)
		if err == nil {
			t.Fatal("expected error for wrong denom")
		}
	})

	t.Run("zero amount", func(t *testing.T) {
		err := k.WithdrawFromDomain(ctx, "WithdrawDomain", recipient, sdk.NewInt64Coin("pnyx", 0), admin)
		if err == nil {
			t.Fatal("expected error for zero amount")
		}
	})
}

// --- Round-trip Test ---

func TestDepositWithdrawRoundTrip(t *testing.T) {
	k, ctx, bk := setupKeeperWithBank(t)

	admin := sdk.AccAddress("admin1")
	user := sdk.AccAddress("user-round")
	k.CreateDomain(ctx, "RoundTrip", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 0)))

	// Fund user with 1000 PNYX.
	bk.fundAccount(user, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 1000)))

	// Deposit 500.
	err := k.DepositToDomain(ctx, user, "RoundTrip", sdk.NewInt64Coin("pnyx", 500))
	if err != nil {
		t.Fatalf("deposit: %v", err)
	}

	// Verify mid-state.
	userBal := bk.accounts[user.String()]
	if userBal.AmountOf("pnyx").Int64() != 500 {
		t.Errorf("user after deposit = %d, want 500", userBal.AmountOf("pnyx").Int64())
	}
	domain, _ := k.GetDomain(ctx, "RoundTrip")
	if domain.Treasury.AmountOf("pnyx").Int64() != 500 {
		t.Errorf("treasury after deposit = %d, want 500", domain.Treasury.AmountOf("pnyx").Int64())
	}

	// Admin withdraws 200 back to user.
	err = k.WithdrawFromDomain(ctx, "RoundTrip", user, sdk.NewInt64Coin("pnyx", 200), admin)
	if err != nil {
		t.Fatalf("withdraw: %v", err)
	}

	// Verify final state.
	userBal = bk.accounts[user.String()]
	if userBal.AmountOf("pnyx").Int64() != 700 {
		t.Errorf("user after withdraw = %d, want 700 (500+200)", userBal.AmountOf("pnyx").Int64())
	}
	domain, _ = k.GetDomain(ctx, "RoundTrip")
	if domain.Treasury.AmountOf("pnyx").Int64() != 300 {
		t.Errorf("treasury after withdraw = %d, want 300 (500-200)", domain.Treasury.AmountOf("pnyx").Int64())
	}
}

// --- Bank Keeper Nil Guard ---

func TestBridgeWithoutBankKeeper(t *testing.T) {
	k, ctx := setupKeeper(t) // bankKeeper is nil
	admin := sdk.AccAddress("admin1")
	k.CreateDomain(ctx, "NoBankDomain", admin, sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100)))

	t.Run("deposit fails gracefully", func(t *testing.T) {
		err := k.DepositToDomain(ctx, admin, "NoBankDomain", sdk.NewInt64Coin("pnyx", 10))
		if err == nil {
			t.Fatal("expected error when bankKeeper is nil")
		}
	})

	t.Run("withdraw fails gracefully", func(t *testing.T) {
		err := k.WithdrawFromDomain(ctx, "NoBankDomain", admin, sdk.NewInt64Coin("pnyx", 10), admin)
		if err == nil {
			t.Fatal("expected error when bankKeeper is nil")
		}
	})
}

// --- ValidateBasic Tests ---

func TestMsgDepositToDomainValidateBasic(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		msg := MsgDepositToDomain{
			Sender:     sdk.AccAddress("sender1"),
			DomainName: "TestDomain",
			Amount:     sdk.NewInt64Coin("pnyx", 100),
		}
		if err := msg.ValidateBasic(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("empty domain", func(t *testing.T) {
		msg := MsgDepositToDomain{
			Sender: sdk.AccAddress("sender1"),
			Amount: sdk.NewInt64Coin("pnyx", 100),
		}
		if err := msg.ValidateBasic(); err == nil {
			t.Error("expected error for empty domain")
		}
	})

	t.Run("wrong denom", func(t *testing.T) {
		msg := MsgDepositToDomain{
			Sender:     sdk.AccAddress("sender1"),
			DomainName: "TestDomain",
			Amount:     sdk.NewInt64Coin("atom", 100),
		}
		if err := msg.ValidateBasic(); err == nil {
			t.Error("expected error for wrong denom")
		}
	})
}

func TestMsgWithdrawFromDomainValidateBasic(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		msg := MsgWithdrawFromDomain{
			Sender:     sdk.AccAddress("sender1"),
			DomainName: "TestDomain",
			Recipient:  "cosmos1abc",
			Amount:     sdk.NewInt64Coin("pnyx", 100),
		}
		if err := msg.ValidateBasic(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("empty recipient", func(t *testing.T) {
		msg := MsgWithdrawFromDomain{
			Sender:     sdk.AccAddress("sender1"),
			DomainName: "TestDomain",
			Amount:     sdk.NewInt64Coin("pnyx", 100),
		}
		if err := msg.ValidateBasic(); err == nil {
			t.Error("expected error for empty recipient")
		}
	})
}
