package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	rewards "truerepublic/treasury/keeper"
	"truerepublic/x/dex"
	"truerepublic/x/truedemocracy"
)

// ensureConsensusGenesis creates a bank-backed PoD bootstrap only from the
// public keys supplied by CometBFT's consensus genesis. It never derives a
// production validator from a hard-coded or otherwise shared private secret.
func ensureConsensusGenesis(cdc codec.Codec, appState map[string]json.RawMessage, validators []abci.ValidatorUpdate) error {
	var democracyGenesis truedemocracy.GenesisState
	if err := json.Unmarshal(appState[truedemocracy.ModuleName], &democracyGenesis); err != nil {
		return fmt.Errorf("decode %s genesis: %w", truedemocracy.ModuleName, err)
	}
	if len(democracyGenesis.Validators) > 0 {
		operators := make([]string, 0, len(democracyGenesis.Validators))
		for _, validator := range democracyGenesis.Validators {
			if _, err := validateOperatorAccountAddress(validator.OperatorAddr); err != nil {
				return err
			}
			operators = append(operators, validator.OperatorAddr)
		}
		if len(validators) == 0 {
			return ensureOperatorAuthAccounts(cdc, appState, operators)
		}
		if len(democracyGenesis.Validators) != len(validators) {
			return fmt.Errorf("application and consensus genesis validator counts differ")
		}
		matched := make([]bool, len(validators))
		for _, appValidator := range democracyGenesis.Validators {
			found := false
			for i, consensusValidator := range validators {
				if matched[i] || !bytes.Equal(appValidator.PubKey, consensusValidator.PubKey.GetEd25519()) {
					continue
				}
				wantPower := appValidator.Stake / rewards.StakeMin
				if consensusValidator.Power != wantPower {
					return fmt.Errorf("application and consensus power differ for validator %q", appValidator.OperatorAddr)
				}
				matched[i] = true
				found = true
				break
			}
			if !found {
				return fmt.Errorf("application validator %q has no matching consensus key", appValidator.OperatorAddr)
			}
		}
		return ensureOperatorAuthAccounts(cdc, appState, operators)
	}
	if len(democracyGenesis.Domains) > 0 {
		return fmt.Errorf("%s genesis defines domains but no validators", truedemocracy.ModuleName)
	}
	if len(validators) == 0 {
		return fmt.Errorf("consensus genesis must provide at least one validator")
	}
	if len(democracyGenesis.BootstrapOperatorAddresses) != len(validators) {
		return fmt.Errorf("consensus genesis requires exactly one independent bootstrap operator address per validator")
	}

	members := make([]string, 0, len(validators))
	genesisValidators := make([]truedemocracy.GenesisValidator, 0, len(validators))
	seen := make(map[string]struct{}, len(validators))
	totalStake := math.ZeroInt()
	var bootstrapAdmin sdk.AccAddress
	for i, validator := range validators {
		pubKey := validator.PubKey.GetEd25519()
		if validator.Power <= 0 || len(pubKey) != ed25519.PubKeySize {
			return fmt.Errorf("consensus validator %d must have positive power and a 32-byte ed25519 key", i)
		}
		operatorAddress, err := validateOperatorAccountAddress(democracyGenesis.BootstrapOperatorAddresses[i])
		if err != nil {
			return fmt.Errorf("bootstrap operator %d is invalid: %w", i, err)
		}
		consensusDerived := sdk.AccAddress((&ed25519.PubKey{Key: pubKey}).Address())
		if operatorAddress.Equals(consensusDerived) {
			return fmt.Errorf("bootstrap operator %d must be independent from its consensus key", i)
		}
		operator := operatorAddress.String()
		if _, exists := seen[operator]; exists {
			return fmt.Errorf("duplicate consensus validator %q", operator)
		}
		seen[operator] = struct{}{}
		if bootstrapAdmin.Empty() {
			bootstrapAdmin = operatorAddress
		}
		members = append(members, operator)
		genesisValidators = append(genesisValidators, truedemocracy.GenesisValidator{
			OperatorAddr: operator,
			PubKey:       pubKey,
			Stake:        rewards.StakeMin,
			Domain:       "Bootstrap",
		})
		totalStake = totalStake.AddRaw(rewards.StakeMin)
	}
	for _, validator := range genesisValidators {
		derived := sdk.AccAddress((&ed25519.PubKey{Key: validator.PubKey}).Address()).String()
		if _, coupled := seen[derived]; coupled {
			return fmt.Errorf("bootstrap operator %q collides with a consensus-key authority", derived)
		}
	}
	democracyGenesis = truedemocracy.GenesisState{
		Domains: []truedemocracy.Domain{{
			Name:          "Bootstrap",
			Admin:         bootstrapAdmin,
			Members:       members,
			Treasury:      sdk.NewCoins(),
			Issues:        []truedemocracy.Issue{},
			Options:       truedemocracy.DomainOptions{AdminElectable: true},
			PermissionReg: []string{},
		}},
		Validators: genesisValidators,
	}
	democracyJSON, err := json.Marshal(democracyGenesis)
	if err != nil {
		return err
	}
	appState[truedemocracy.ModuleName] = democracyJSON

	if err := ensureOperatorAuthAccounts(cdc, appState, members); err != nil {
		return err
	}

	bankGenesis := banktypes.GetGenesisStateFromAppState(cdc, appState)
	moduleAddress := authtypes.NewModuleAddress(truedemocracy.ModuleName).String()
	stakeCoins := sdk.NewCoins(sdk.NewCoin(truedemocracy.PNYXDenom, totalStake))
	found := false
	for i := range bankGenesis.Balances {
		if bankGenesis.Balances[i].Address != moduleAddress {
			continue
		}
		found = true
		if !bankGenesis.Balances[i].Coins.Empty() && !bankGenesis.Balances[i].Coins.Equal(stakeCoins) {
			return fmt.Errorf("existing %s module balance does not match consensus bootstrap stake", truedemocracy.ModuleName)
		}
		if bankGenesis.Balances[i].Coins.Empty() {
			bankGenesis.Balances[i].Coins = stakeCoins
			bankGenesis.Supply = bankGenesis.Supply.Add(stakeCoins...)
		}
		break
	}
	if !found {
		bankGenesis.Balances = append(bankGenesis.Balances, banktypes.Balance{Address: moduleAddress, Coins: stakeCoins})
		bankGenesis.Supply = bankGenesis.Supply.Add(stakeCoins...)
	}
	bankJSON, err := cdc.MarshalJSON(bankGenesis)
	if err != nil {
		return err
	}
	appState[banktypes.ModuleName] = bankJSON
	return nil
}

func validateOperatorAccountAddress(operator string) (sdk.AccAddress, error) {
	address, err := sdk.AccAddressFromBech32(operator)
	if err != nil {
		return nil, err
	}
	for moduleName := range maccPerms {
		if address.Equals(authtypes.NewModuleAddress(moduleName)) {
			return nil, fmt.Errorf("operator address %q is reserved for module %q", operator, moduleName)
		}
	}
	return address, nil
}

func ensureOperatorAuthAccounts(cdc codec.Codec, appState map[string]json.RawMessage, operators []string) error {
	authGenesis := authtypes.GetGenesisStateFromAppState(cdc, appState)
	existingAccounts, err := authtypes.UnpackAccounts(authGenesis.Accounts)
	if err != nil {
		return fmt.Errorf("decode auth genesis accounts: %w", err)
	}
	existingAddresses := make(map[string]struct{}, len(existingAccounts))
	nextAccountNumber := uint64(0)
	for _, account := range existingAccounts {
		existingAddresses[account.GetAddress().String()] = struct{}{}
		if account.GetAccountNumber() >= nextAccountNumber {
			if account.GetAccountNumber() == ^uint64(0) {
				return fmt.Errorf("auth genesis account number space is exhausted")
			}
			nextAccountNumber = account.GetAccountNumber() + 1
		}
	}
	for _, operator := range operators {
		address, err := validateOperatorAccountAddress(operator)
		if err != nil {
			return err
		}
		if _, exists := existingAddresses[operator]; exists {
			for _, account := range existingAccounts {
				if account.GetAddress().String() == operator {
					if _, reserved := account.(authtypes.ModuleAccountI); reserved {
						return fmt.Errorf("operator address %q is a module account", operator)
					}
					break
				}
			}
			continue
		}
		account := authtypes.NewBaseAccountWithAddress(address)
		account.AccountNumber = nextAccountNumber
		nextAccountNumber++
		existingAccounts = append(existingAccounts, account)
		existingAddresses[operator] = struct{}{}
	}
	packedAccounts, err := authtypes.PackAccounts(existingAccounts)
	if err != nil {
		return fmt.Errorf("encode auth genesis accounts: %w", err)
	}
	authGenesis.Accounts = packedAccounts
	authJSON, err := cdc.MarshalJSON(&authGenesis)
	if err != nil {
		return err
	}
	appState[authtypes.ModuleName] = authJSON
	return nil
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
