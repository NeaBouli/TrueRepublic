package dex

import (
	"encoding/json"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func validDEXGenesis() GenesisState {
	provider := sdk.AccAddress("genesis-lp")
	genesis := DefaultGenesisState()
	genesis.Pools = []Pool{{
		PnyxReserve: math.NewInt(2_000), AssetReserve: math.NewInt(1_000), AssetDenom: "atom",
		TotalShares: math.NewInt(100), TotalBurned: math.ZeroInt(), TotalVolumePnyx: math.ZeroInt(),
	}}
	genesis.LPPositions = []LPPosition{{AssetDenom: "atom", Provider: provider.String(), Shares: math.NewInt(100)}}
	return genesis
}

func TestValidateGenesisStateRejectsMalformedAndDuplicateDEXState(t *testing.T) {
	if err := ValidateGenesisState(validDEXGenesis()); err != nil {
		t.Fatalf("valid genesis rejected: %v", err)
	}
	tests := []struct {
		name   string
		mutate func(*GenesisState)
	}{
		{"duplicate asset", func(g *GenesisState) { g.RegisteredAssets = append(g.RegisteredAssets, g.RegisteredAssets[0]) }},
		{"duplicate pool", func(g *GenesisState) { g.Pools = append(g.Pools, g.Pools[0]) }},
		{"negative reserve", func(g *GenesisState) { g.Pools[0].AssetReserve = math.NewInt(-1) }},
		{"missing LP ownership", func(g *GenesisState) { g.LPPositions = nil }},
		{"duplicate LP ownership", func(g *GenesisState) { g.LPPositions = append(g.LPPositions, g.LPPositions[0]) }},
		{"orphan LP ownership", func(g *GenesisState) { g.LPPositions[0].AssetDenom = "btc" }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			genesis := validDEXGenesis()
			tc.mutate(&genesis)
			if err := ValidateGenesisState(genesis); err == nil {
				t.Fatal("malformed genesis was accepted")
			}
		})
	}
}

func TestDEXGenesisExportIncludesProviderLPOwnership(t *testing.T) {
	keeper, ctx, bank, _ := setupCustodyKeeper(t)
	provider := sdk.AccAddress("export-provider")
	bank.fundAccount(ctx, provider, sdk.NewCoins(sdk.NewInt64Coin(pnyxDenom, 1_000), sdk.NewInt64Coin("atom", 1_000)))
	if err := keeper.CreatePoolWithCustody(ctx, provider, "atom", math.NewInt(1_000), math.NewInt(1_000)); err != nil {
		t.Fatal(err)
	}
	exported := NewAppModule(keeper.cdc, keeper).ExportGenesis(ctx, nil)
	var genesis GenesisState
	if err := json.Unmarshal(exported, &genesis); err != nil {
		t.Fatal(err)
	}
	if err := ValidateGenesisState(genesis); err != nil {
		t.Fatal(err)
	}
	if len(genesis.LPPositions) != 1 || genesis.LPPositions[0].Provider != provider.String() {
		t.Fatalf("provider LP ownership not exported: %+v", genesis.LPPositions)
	}
}
