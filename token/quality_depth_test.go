package token

import (
	"context"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// capBoundaryAmounts covers the exact 21,000,000 PNYX cap boundary.
func capBoundaryAmounts() []struct {
	name    string
	amount  math.Int
	wantErr bool
} {
	return []struct {
		name    string
		amount  math.Int
		wantErr bool
	}{
		{name: "cap minus one", amount: MaxSupply().SubRaw(1)},
		{name: "exact cap", amount: MaxSupply()},
		{name: "cap plus one", amount: MaxSupply().AddRaw(1), wantErr: true},
	}
}

// TestPNYXCapBoundaryProperties asserts the exact cap boundary for both bank
// genesis validation and the issuance service mint path.
func TestPNYXCapBoundaryProperties(t *testing.T) {
	for _, test := range capBoundaryAmounts() {
		t.Run("genesis/"+test.name, func(t *testing.T) {
			err := ValidateGenesisSupply(bankGenesisWithSupply(test.amount))
			if (err != nil) != test.wantErr {
				t.Fatalf("ValidateGenesisSupply() error = %v, wantErr %v", err, test.wantErr)
			}
		})

		t.Run("issuance/"+test.name, func(t *testing.T) {
			bank := newIssuanceBank(test.amount)
			service := NewIssuanceService(bank, "rewards")
			minted, err := service.MintUpToCap(context.Background(), math.OneInt())
			if test.wantErr {
				if err == nil {
					t.Fatalf("over-cap supply %s accepted a mint", test.amount)
				}
				return
			}
			if err != nil {
				t.Fatalf("mint at supply %s failed: %v", test.amount, err)
			}
			if minted.IsNegative() || minted.GT(math.OneInt()) {
				t.Fatalf("minted %s outside [0, requested 1]", minted)
			}
			if bank.supply.GT(MaxSupply()) {
				t.Fatalf("supply %s exceeded cap %s after mint", bank.supply, MaxSupply())
			}
		})
	}
}

// TestCanonicalSupplyPrefersExplicitOverBalances asserts the explicit bank
// genesis supply takes precedence and the balance-derived fallback mirrors the
// Cosmos SDK derivation.
func TestCanonicalSupplyPrefersExplicitOverBalances(t *testing.T) {
	holder := sdk.AccAddress("explicit-holder").String()

	explicit := banktypes.GenesisState{
		Params:   banktypes.DefaultParams(),
		Balances: []banktypes.Balance{{Address: holder, Coins: sdk.NewCoins(NewCoin(math.NewInt(10)))}},
		Supply:   sdk.NewCoins(NewCoin(math.NewInt(42))),
	}
	if got := CanonicalSupply(explicit); !got.Equal(math.NewInt(42)) {
		t.Fatalf("CanonicalSupply() = %s, want explicit 42", got)
	}

	derived := banktypes.GenesisState{
		Params: banktypes.DefaultParams(),
		Balances: []banktypes.Balance{
			{Address: holder, Coins: sdk.NewCoins(NewCoin(math.NewInt(10)))},
			{Address: sdk.AccAddress("other-holder").String(), Coins: sdk.NewCoins(NewCoin(math.NewInt(15)))},
		},
	}
	if got := CanonicalSupply(derived); !got.Equal(math.NewInt(25)) {
		t.Fatalf("CanonicalSupply() = %s, want balance-derived 25", got)
	}

	// Both representations of the exact cap must validate; a balance-derived
	// supply above the cap must not.
	explicitCap := bankGenesisWithSupply(MaxSupply())
	if err := ValidateGenesisSupply(explicitCap); err != nil {
		t.Fatalf("explicit cap rejected: %v", err)
	}
	derivedCap := banktypes.GenesisState{
		Params: banktypes.DefaultParams(),
		Balances: []banktypes.Balance{
			{Address: holder, Coins: sdk.NewCoins(NewCoin(MaxSupply().SubRaw(5)))},
			{Address: sdk.AccAddress("other-holder").String(), Coins: sdk.NewCoins(NewCoin(math.NewInt(5)))},
		},
	}
	if err := ValidateGenesisSupply(derivedCap); err != nil {
		t.Fatalf("balance-derived cap rejected: %v", err)
	}
	derivedCap.Balances[1].Coins = sdk.NewCoins(NewCoin(math.NewInt(6)))
	if err := ValidateGenesisSupply(derivedCap); err == nil {
		t.Fatal("balance-derived cap+1 accepted")
	}
}

// TestValidateGenesisSupplyRejectsMalformedStructures asserts invalid,
// negative, duplicate, and legacy-denom structures are rejected without a
// panic.
func TestValidateGenesisSupplyRejectsMalformedStructures(t *testing.T) {
	holder := sdk.AccAddress("malformed-holder").String()

	tests := []struct {
		name    string
		genesis banktypes.GenesisState
	}{
		{
			name: "legacy denom supply",
			genesis: banktypes.GenesisState{
				Params: banktypes.DefaultParams(),
				Supply: sdk.NewCoins(sdk.NewInt64Coin(DisplayDenom, 1)),
			},
		},
		{
			name: "legacy denom balance",
			genesis: banktypes.GenesisState{
				Params:   banktypes.DefaultParams(),
				Balances: []banktypes.Balance{{Address: holder, Coins: sdk.NewCoins(sdk.NewInt64Coin(DisplayDenom, 1))}},
			},
		},
		{
			name: "duplicate balance addresses",
			genesis: banktypes.GenesisState{
				Params: banktypes.DefaultParams(),
				Balances: []banktypes.Balance{
					{Address: holder, Coins: sdk.NewCoins(NewCoin(math.NewInt(1)))},
					{Address: holder, Coins: sdk.NewCoins(NewCoin(math.NewInt(2)))},
				},
			},
		},
		{
			name: "invalid balance address",
			genesis: banktypes.GenesisState{
				Params:   banktypes.DefaultParams(),
				Balances: []banktypes.Balance{{Address: "not-a-valid-address", Coins: sdk.NewCoins(NewCoin(math.NewInt(1)))}},
			},
		},
		{
			name: "negative balance coin",
			genesis: banktypes.GenesisState{
				Params: banktypes.DefaultParams(),
				Balances: []banktypes.Balance{{
					Address: holder,
					Coins:   sdk.Coins{sdk.Coin{Denom: BaseDenom, Amount: math.NewInt(-1)}},
				}},
			},
		},
		{
			name: "negative explicit supply",
			genesis: banktypes.GenesisState{
				Params: banktypes.DefaultParams(),
				Supply: sdk.Coins{sdk.Coin{Denom: BaseDenom, Amount: math.NewInt(-1)}},
			},
		},
		{
			name: "duplicate denom within one balance",
			genesis: banktypes.GenesisState{
				Params: banktypes.DefaultParams(),
				Balances: []banktypes.Balance{{
					Address: holder,
					Coins: sdk.Coins{
						sdk.Coin{Denom: BaseDenom, Amount: math.NewInt(1)},
						sdk.Coin{Denom: BaseDenom, Amount: math.NewInt(2)},
					},
				}},
			},
		},
		{
			name:    "explicit supply above cap with matching balances",
			genesis: bankGenesisWithSupply(MaxSupply().AddRaw(1)),
		},
		{
			name: "explicit supply mismatches balances",
			genesis: banktypes.GenesisState{
				Params:   banktypes.DefaultParams(),
				Balances: []banktypes.Balance{{Address: holder, Coins: sdk.NewCoins(NewCoin(math.NewInt(1)))}},
				Supply:   sdk.NewCoins(NewCoin(math.NewInt(2))),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := ValidateGenesisSupply(test.genesis); err == nil {
				t.Fatal("malformed genesis accepted")
			}
		})
	}
}

// TestPNYXCapGenerativeSweep sweeps deterministic amounts around the cap and
// asserts validation rejects exactly the over-cap amounts while accepted mints
// never push canonical supply above the cap.
func TestPNYXCapGenerativeSweep(t *testing.T) {
	seed := uint64(0x243f6a8885a308d3)
	next := func() uint64 {
		seed = seed*6364136223846793005 + 1442695040888963407
		return seed >> 16
	}
	window := uint64(4*WholeTokenBaseUnits + 1)

	for i := 0; i < 2000; i++ {
		delta := int64(next() % window)
		amount := MaxSupply().SubRaw(2 * WholeTokenBaseUnits).AddRaw(delta)

		err := ValidateGenesisSupply(bankGenesisWithSupply(amount))
		if wantErr := amount.GT(MaxSupply()); (err != nil) != wantErr {
			t.Fatalf("ValidateGenesisSupply(%s) error = %v, wantErr %t", amount, err, wantErr)
		}

		if amount.GT(MaxSupply()) {
			continue
		}
		requested := math.NewInt(int64(next() % uint64(3*WholeTokenBaseUnits)))
		bank := newIssuanceBank(amount)
		minted, err := NewIssuanceService(bank, "rewards").MintUpToCap(context.Background(), requested)
		if err != nil {
			t.Fatalf("mint of %s at supply %s failed: %v", requested, amount, err)
		}
		if minted.IsNegative() || minted.GT(requested) {
			t.Fatalf("minted %s outside [0, requested %s]", minted, requested)
		}
		if bank.supply.GT(MaxSupply()) {
			t.Fatalf("supply %s exceeded cap %s after accepted mint", bank.supply, MaxSupply())
		}
	}
}

// FuzzPNYXCapAndGenesisValidation fuzzes the 21,000,000 PNYX cap from both
// directions: arbitrary issuance mints and arbitrary bank genesis structures
// must never panic, and any accepted result must keep canonical supply within
// the cap.
func FuzzPNYXCapAndGenesisValidation(f *testing.F) {
	f.Add(uint64(MaxSupplyBaseUnits-1), uint64(1), false, true)
	f.Add(uint64(MaxSupplyBaseUnits), uint64(MaxSupplyBaseUnits), false, false)
	f.Add(uint64(MaxSupplyBaseUnits+1), uint64(0), false, true)
	f.Add(uint64(0), uint64(1), true, true)
	f.Add(^uint64(0), ^uint64(0), true, false)
	f.Fuzz(func(t *testing.T, rawSupply, rawRequested uint64, legacy, explicit bool) {
		// Map raw input into [-2, 2*cap+1] to cover invalid, below-cap,
		// boundary, and above-cap values.
		span := uint64(2*MaxSupplyBaseUnits + 4)
		supply := int64(rawSupply%span) - 2
		requested := int64(rawRequested%span) - 2

		// Issuance: an accepted mint must be bounded by the request and must
		// never push canonical supply above the cap.
		bank := newIssuanceBank(math.NewInt(supply))
		minted, err := NewIssuanceService(bank, "fuzz-rewards").MintUpToCap(context.Background(), math.NewInt(requested))
		if err == nil {
			if minted.IsNegative() || minted.GT(math.NewInt(requested)) {
				t.Fatalf("minted %s outside [0, requested %d]", minted, requested)
			}
			if bank.supply.GT(MaxSupply()) {
				t.Fatalf("supply %s exceeded cap %s after accepted mint", bank.supply, MaxSupply())
			}
		}

		// Genesis: arbitrary structures must not panic, and any accepted
		// genesis must stay within the cap and off the legacy display denom.
		coin := sdk.Coin{Denom: BaseDenom, Amount: math.NewInt(supply)}
		if legacy {
			coin.Denom = DisplayDenom
		}
		genesis := banktypes.GenesisState{
			Params:   banktypes.DefaultParams(),
			Balances: []banktypes.Balance{{Address: sdk.AccAddress("fuzz-holder").String(), Coins: sdk.Coins{coin}}},
		}
		if explicit {
			genesis.Supply = sdk.Coins{coin}
		}
		if err := ValidateGenesisSupply(genesis); err == nil {
			if got := CanonicalSupply(genesis); got.GT(MaxSupply()) {
				t.Fatalf("accepted genesis with canonical supply %s above cap %s", got, MaxSupply())
			}
			if genesis.Supply.AmountOf(DisplayDenom).IsPositive() {
				t.Fatal("accepted genesis with legacy display denom supply")
			}
			for _, balance := range genesis.Balances {
				if balance.Coins.AmountOf(DisplayDenom).IsPositive() {
					t.Fatal("accepted genesis with legacy display denom balance")
				}
			}
		}
	})
}
