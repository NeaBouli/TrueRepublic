package main

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"truerepublic/x/dex"
	"truerepublic/x/truedemocracy"
)

// bankAppModuleBasic supplies the bank half of the deterministic bootstrap
// validator. Custom genesis files must keep bank/custom claims exactly aligned.
type bankAppModuleBasic struct{ bank.AppModuleBasic }

func (bankAppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	stake := sdk.NewInt64Coin(truedemocracy.PNYXDenom, 100_000*truedemocracy.PNYXUnit)
	balance := banktypes.Balance{
		Address: authtypes.NewModuleAddress(truedemocracy.ModuleName).String(),
		Coins:   sdk.NewCoins(stake),
	}
	genesis := banktypes.NewGenesisState(
		banktypes.DefaultParams(),
		[]banktypes.Balance{balance},
		sdk.NewCoins(stake),
		nil,
		nil,
	)
	return cdc.MustMarshalJSON(genesis)
}

// validateLedgerGenesis reconciles custom claims against exact x/bank module
// balances. It runs before any module mutates consensus state.
func validateLedgerGenesis(cdc codec.Codec, appState map[string]json.RawMessage) error {
	bankGenesis := banktypes.GetGenesisStateFromAppState(cdc, appState)

	var democracyGenesis truedemocracy.GenesisState
	if err := json.Unmarshal(appState[truedemocracy.ModuleName], &democracyGenesis); err != nil {
		return fmt.Errorf("decode %s genesis: %w", truedemocracy.ModuleName, err)
	}
	democracyClaims, err := truedemocracy.GenesisEscrowClaims(democracyGenesis)
	if err != nil {
		return fmt.Errorf("validate %s genesis: %w", truedemocracy.ModuleName, err)
	}
	wantDemocracy := sdk.NewCoins()
	if democracyClaims.IsPositive() {
		wantDemocracy = sdk.NewCoins(sdk.NewCoin(truedemocracy.PNYXDenom, democracyClaims))
	}
	if err := requireModuleGenesisBalance(*bankGenesis, truedemocracy.ModuleName, wantDemocracy); err != nil {
		return err
	}

	var dexGenesis dex.GenesisState
	if err := json.Unmarshal(appState[dex.ModuleName], &dexGenesis); err != nil {
		return fmt.Errorf("decode %s genesis: %w", dex.ModuleName, err)
	}
	dexClaims, err := dex.GenesisReserveClaims(dexGenesis)
	if err != nil {
		return fmt.Errorf("validate %s genesis: %w", dex.ModuleName, err)
	}
	if err := requireModuleGenesisBalance(*bankGenesis, dex.ModuleName, dexClaims); err != nil {
		return err
	}
	return nil
}

func requireModuleGenesisBalance(genesis banktypes.GenesisState, moduleName string, expected sdk.Coins) error {
	moduleAddress := authtypes.NewModuleAddress(moduleName).String()
	actual := sdk.NewCoins()
	for _, balance := range genesis.Balances {
		if balance.Address == moduleAddress {
			actual = balance.Coins
			break
		}
	}
	if !actual.Equal(expected) {
		return fmt.Errorf("%s genesis bank balance %s does not equal custom claims %s", moduleName, actual, expected)
	}
	return nil
}
