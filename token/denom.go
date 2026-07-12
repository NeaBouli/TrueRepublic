package token

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	BaseDenom           = "upnyx"
	DisplayDenom        = "pnyx"
	Symbol              = "PNYX"
	Decimals     uint32 = 6

	WholeTokenBaseUnits int64 = 1_000_000
	MaxSupplyWhole      int64 = 21_000_000
	MaxSupplyBaseUnits  int64 = MaxSupplyWhole * WholeTokenBaseUnits
	StakeMinWhole       int64 = 100_000
	StakeMinBaseUnits   int64 = StakeMinWhole * WholeTokenBaseUnits
)

func MaxSupply() math.Int {
	return math.NewInt(MaxSupplyBaseUnits)
}

func Metadata() banktypes.Metadata {
	return banktypes.Metadata{
		Description: "The native governance and utility token of TrueRepublic",
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: BaseDenom, Exponent: 0},
			{Denom: DisplayDenom, Exponent: Decimals},
		},
		Base:    BaseDenom,
		Display: DisplayDenom,
		Name:    "PNYX",
		Symbol:  Symbol,
	}
}

// CanonicalSupply returns the PNYX supply represented by bank genesis. When
// the explicit supply is absent, Cosmos SDK derives it from account balances;
// this function mirrors that behavior for pre-init validation.
func CanonicalSupply(genesis banktypes.GenesisState) math.Int {
	if !genesis.Supply.Empty() {
		return genesis.Supply.AmountOf(BaseDenom)
	}

	total := math.ZeroInt()
	for _, balance := range genesis.Balances {
		total = total.Add(balance.Coins.AmountOf(BaseDenom))
	}
	return total
}

func ValidateGenesisSupply(genesis banktypes.GenesisState) error {
	if err := genesis.Validate(); err != nil {
		return fmt.Errorf("invalid bank genesis: %w", err)
	}

	if legacySupply := genesis.Supply.AmountOf(DisplayDenom); legacySupply.IsPositive() {
		return fmt.Errorf("native supply must use %s, found %s%s", BaseDenom, legacySupply, DisplayDenom)
	}
	for _, balance := range genesis.Balances {
		if legacyBalance := balance.Coins.AmountOf(DisplayDenom); legacyBalance.IsPositive() {
			return fmt.Errorf("native balances must use %s, found %s%s", BaseDenom, legacyBalance, DisplayDenom)
		}
	}

	supply := CanonicalSupply(genesis)
	if supply.GT(MaxSupply()) {
		return fmt.Errorf("PNYX supply %s%s exceeds maximum %s%s", supply, BaseDenom, MaxSupply(), BaseDenom)
	}
	return nil
}

func EnsureMetadata(genesis *banktypes.GenesisState) {
	metadata := make([]banktypes.Metadata, 0, len(genesis.DenomMetadata)+1)
	for _, existing := range genesis.DenomMetadata {
		// A pre-migration metadata entry with pnyx as its base conflicts with
		// pnyx becoming the canonical display denom. Drop it during genesis
		// normalization and preserve metadata for unrelated assets.
		if existing.Base == BaseDenom || existing.Base == DisplayDenom {
			continue
		}
		metadata = append(metadata, existing)
	}
	genesis.DenomMetadata = append(metadata, Metadata())
}

func NewCoin(amount math.Int) sdk.Coin {
	return sdk.NewCoin(BaseDenom, amount)
}
