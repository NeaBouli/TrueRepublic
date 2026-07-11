package token

import (
	"strings"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func bankGenesisWithSupply(amount math.Int) banktypes.GenesisState {
	coin := sdk.NewCoin(BaseDenom, amount)
	return *banktypes.NewGenesisState(
		banktypes.DefaultParams(),
		[]banktypes.Balance{{Address: sdk.AccAddress("supply-holder").String(), Coins: sdk.NewCoins(coin)}},
		sdk.NewCoins(coin),
		nil,
		nil,
	)
}

func TestValidateGenesisSupplyBoundaries(t *testing.T) {
	tests := []struct {
		name    string
		amount  math.Int
		wantErr bool
	}{
		{name: "cap minus one", amount: MaxSupply().SubRaw(1)},
		{name: "exact cap", amount: MaxSupply()},
		{name: "cap plus one", amount: MaxSupply().AddRaw(1), wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateGenesisSupply(bankGenesisWithSupply(test.amount))
			if (err != nil) != test.wantErr {
				t.Fatalf("ValidateGenesisSupply() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestValidateGenesisSupplyRejectsLegacyDisplayDenom(t *testing.T) {
	coin := sdk.NewInt64Coin(DisplayDenom, 1)
	genesis := *banktypes.NewGenesisState(
		banktypes.DefaultParams(),
		[]banktypes.Balance{{Address: sdk.AccAddress("legacy-holder").String(), Coins: sdk.NewCoins(coin)}},
		sdk.NewCoins(coin),
		nil,
		nil,
	)

	err := ValidateGenesisSupply(genesis)
	if err == nil || !strings.Contains(err.Error(), BaseDenom) {
		t.Fatalf("expected canonical-denom error, got %v", err)
	}
}

func TestCanonicalSupplyFallsBackToBalances(t *testing.T) {
	genesis := banktypes.GenesisState{
		Params: banktypes.DefaultParams(),
		Balances: []banktypes.Balance{
			{Address: sdk.AccAddress("holder-one").String(), Coins: sdk.NewCoins(sdk.NewInt64Coin(BaseDenom, 10))},
			{Address: sdk.AccAddress("holder-two").String(), Coins: sdk.NewCoins(sdk.NewInt64Coin(BaseDenom, 15))},
		},
	}

	if got := CanonicalSupply(genesis); !got.Equal(math.NewInt(25)) {
		t.Fatalf("CanonicalSupply() = %s, want 25", got)
	}
}

func TestEnsureMetadata(t *testing.T) {
	genesis := banktypes.GenesisState{Params: banktypes.DefaultParams()}
	EnsureMetadata(&genesis)
	EnsureMetadata(&genesis)

	if len(genesis.DenomMetadata) != 1 {
		t.Fatalf("metadata count = %d, want 1", len(genesis.DenomMetadata))
	}
	metadata := genesis.DenomMetadata[0]
	if metadata.Base != BaseDenom || metadata.Display != DisplayDenom || len(metadata.DenomUnits) != 2 {
		t.Fatalf("unexpected metadata: %+v", metadata)
	}
}
