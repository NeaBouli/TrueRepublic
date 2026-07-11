package truedemocracy

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestValidateGenesisStateRejectsMalformedAndDuplicateDemocracyState(t *testing.T) {
	if err := ValidateGenesisState(DefaultGenesisState()); err != nil {
		t.Fatalf("valid default genesis rejected: %v", err)
	}
	tests := []struct {
		name   string
		mutate func(*GenesisState)
	}{
		{"duplicate domain", func(g *GenesisState) { g.Domains = append(g.Domains, g.Domains[0]) }},
		{"duplicate member", func(g *GenesisState) { g.Domains[0].Members = append(g.Domains[0].Members, g.Domains[0].Members[0]) }},
		{"negative treasury", func(g *GenesisState) {
			g.Domains[0].Treasury = sdk.Coins{{Denom: PNYXDenom, Amount: math.NewInt(-1)}}
		}},
		{"duplicate validator", func(g *GenesisState) { g.Validators = append(g.Validators, g.Validators[0]) }},
		{"missing validator domain", func(g *GenesisState) { g.Validators[0].Domain = "missing" }},
		{"invalid validator pubkey", func(g *GenesisState) { g.Validators[0].PubKey = []byte{1} }},
		{"invalid member", func(g *GenesisState) { g.Domains[0].Members[0] = "not-an-address" }},
		{"admin not member", func(g *GenesisState) { g.Domains[0].Members = []string{sdk.AccAddress("other-member").String()} }},
		{"transfer limit exceeded", func(g *GenesisState) { g.Domains[0].TransferredStake = 1 }},
		{"invalid rating", func(g *GenesisState) {
			g.Domains[0].Issues = []Issue{{Name: "issue", Suggestions: []Suggestion{{Name: "proposal", Creator: g.Domains[0].Admin.String(), Ratings: []Rating{{Value: 6}}}}}}
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			genesis := DefaultGenesisState()
			tc.mutate(&genesis)
			if err := ValidateGenesisState(genesis); err == nil {
				t.Fatal("malformed genesis was accepted")
			}
		})
	}
}

func TestGenesisEscrowClaimsIncludesTreasuryAndStake(t *testing.T) {
	genesis := DefaultGenesisState()
	genesis.Domains[0].Treasury = sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, 123))
	claims, err := GenesisEscrowClaims(genesis)
	if err != nil {
		t.Fatal(err)
	}
	want := math.NewInt(genesis.Validators[0].Stake + 123)
	if !claims.Equal(want) {
		t.Fatalf("claims=%s want=%s", claims, want)
	}
}
