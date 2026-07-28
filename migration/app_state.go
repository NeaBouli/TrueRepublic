package migration

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	rewards "truerepublic/treasury/keeper"
	"truerepublic/x/dex"
	"truerepublic/x/truedemocracy"
)

// TransformApplicationGenesis composes the verified truedemocracy rewrite with
// exact auth, bank, and DEX ownership reconciliation. It also changes the outer
// chain ID, enforces the signed halt/export height, and rejects CosmWasm code or
// contract state that a generic address transformer cannot safely interpret.
//
// SourceAppHash is deliberately not derived from genesis JSON: callers must
// compare the descriptor's hash with the trusted source header at HaltHeight.
// TransformTrueDemocracy separately hashes the exact raw genesis bytes and
// requires them to match the signed SourceGenesisSHA256 commitment.
func TransformApplicationGenesis(
	desc *Descriptor,
	raw json.RawMessage,
	cdc codec.Codec,
) (json.RawMessage, error) {
	if cdc == nil {
		return nil, fmt.Errorf("migration: nil application codec")
	}

	transformed, err := TransformTrueDemocracy(desc, raw)
	if err != nil {
		return nil, err
	}

	var root map[string]json.RawMessage
	if err := json.Unmarshal(transformed, &root); err != nil {
		return nil, fmt.Errorf("migration: decode transformed genesis: %w", err)
	}
	var sourceChainID string
	if err := json.Unmarshal(root["chain_id"], &sourceChainID); err != nil {
		return nil, fmt.Errorf("migration: outer genesis has invalid chain_id: %w", err)
	}
	if sourceChainID != desc.SourceChainID {
		return nil, fmt.Errorf(
			"migration: outer chain ID %q does not match descriptor source %q",
			sourceChainID,
			desc.SourceChainID,
		)
	}
	var initialHeight int64
	if err := json.Unmarshal(root["initial_height"], &initialHeight); err != nil {
		return nil, fmt.Errorf("migration: outer genesis has invalid initial_height: %w", err)
	}
	if want := desc.HaltHeight + 1; initialHeight != want {
		return nil, fmt.Errorf(
			"migration: initial height %d does not match halt-height successor %d",
			initialHeight,
			want,
		)
	}
	var embeddedAppHash []byte
	appHashJSON, found := root["app_hash"]
	if !found {
		return nil, fmt.Errorf("migration: outer genesis is missing app_hash")
	}
	if err := json.Unmarshal(appHashJSON, &embeddedAppHash); err != nil {
		return nil, fmt.Errorf("migration: outer genesis has invalid app_hash: %w", err)
	}
	if len(embeddedAppHash) != 0 {
		return nil, fmt.Errorf("migration: fresh target genesis must not embed a prior app_hash")
	}

	var appState map[string]json.RawMessage
	if err := json.Unmarshal(root["app_state"], &appState); err != nil {
		return nil, fmt.Errorf("migration: decode transformed app_state: %w", err)
	}
	for _, name := range []string{
		authtypes.ModuleName,
		banktypes.ModuleName,
		dex.ModuleName,
		wasmtypes.ModuleName,
	} {
		if _, found := appState[name]; !found {
			return nil, fmt.Errorf("migration: app_state is missing module %q", name)
		}
	}
	if err := validateConsensusGenesis(root, appState[truedemocracy.ModuleName]); err != nil {
		return nil, err
	}

	replacements := make(map[string]OperatorMapping, len(desc.Mappings))
	targets := make(map[string]struct{}, len(desc.Mappings))
	for _, mapping := range desc.Mappings {
		replacements[mapping.OldOperator] = mapping
		targets[mapping.NewOperator] = struct{}{}
	}

	authJSON, moduleAccounts, err := reconcileAuth(cdc, appState[authtypes.ModuleName], replacements, targets)
	if err != nil {
		return nil, err
	}
	bankJSON, err := reconcileBank(cdc, appState[banktypes.ModuleName], replacements, targets, moduleAccounts)
	if err != nil {
		return nil, err
	}
	dexJSON, err := reconcileDEX(appState[dex.ModuleName], replacements, targets)
	if err != nil {
		return nil, err
	}
	if err := enforceEmptyWasm(cdc, appState[wasmtypes.ModuleName]); err != nil {
		return nil, err
	}

	appState[authtypes.ModuleName] = authJSON
	appState[banktypes.ModuleName] = bankJSON
	appState[dex.ModuleName] = dexJSON
	for name, moduleJSON := range appState {
		for _, mapping := range desc.Mappings {
			if bytes.Contains(moduleJSON, []byte(mapping.OldOperator)) {
				return nil, fmt.Errorf(
					"migration: module %q still contains old operator %q",
					name,
					mapping.OldOperator,
				)
			}
		}
	}

	appStateJSON, err := json.Marshal(appState)
	if err != nil {
		return nil, fmt.Errorf("migration: encode reconciled app_state: %w", err)
	}
	targetChainID, err := json.Marshal(desc.TargetChainID)
	if err != nil {
		return nil, fmt.Errorf("migration: encode target chain ID: %w", err)
	}
	root["chain_id"] = targetChainID
	root["app_state"] = appStateJSON
	output, err := json.Marshal(root)
	if err != nil {
		return nil, fmt.Errorf("migration: encode reconciled genesis: %w", err)
	}
	return output, nil
}

func validateConsensusGenesis(
	root map[string]json.RawMessage,
	truedemocracyJSON json.RawMessage,
) error {
	consensusJSON, found := root["consensus"]
	if !found {
		return fmt.Errorf("migration: outer genesis is missing consensus state")
	}
	var consensus genutiltypes.ConsensusGenesis
	if err := json.Unmarshal(consensusJSON, &consensus); err != nil {
		return fmt.Errorf("migration: decode consensus genesis: %w", err)
	}
	consensusCopy := consensus
	consensusCopy.Validators = append(consensusCopy.Validators[:0:0], consensus.Validators...)
	if err := consensusCopy.ValidateAndComplete(); err != nil {
		return fmt.Errorf("migration: invalid consensus genesis: %w", err)
	}
	if len(consensus.Validators) == 0 {
		return nil
	}

	var state truedemocracy.GenesisState
	if err := json.Unmarshal(truedemocracyJSON, &state); err != nil {
		return fmt.Errorf("migration: decode reconciled truedemocracy validators: %w", err)
	}
	expected := make(map[string]int64, len(state.Validators))
	for _, validator := range state.Validators {
		power, active := exportedValidatorPower(validator)
		if !active {
			continue
		}
		key := hex.EncodeToString(validator.PubKey)
		if _, duplicate := expected[key]; duplicate {
			return fmt.Errorf("migration: duplicate active application consensus key %q", key)
		}
		expected[key] = power
	}
	if len(consensus.Validators) != len(expected) {
		return fmt.Errorf(
			"migration: consensus validator count %d does not match active application count %d",
			len(consensus.Validators),
			len(expected),
		)
	}
	seen := make(map[string]struct{}, len(consensus.Validators))
	for i, validator := range consensus.Validators {
		pubKey, ok := validator.PubKey.(cmted25519.PubKey)
		if !ok {
			return fmt.Errorf(
				"migration: consensus validator %d uses unsupported key type %T",
				i,
				validator.PubKey,
			)
		}
		key := hex.EncodeToString(pubKey)
		if _, duplicate := seen[key]; duplicate {
			return fmt.Errorf("migration: duplicate consensus validator key %q", key)
		}
		seen[key] = struct{}{}
		power, found := expected[key]
		if !found {
			return fmt.Errorf("migration: consensus validator %d is absent from application state", i)
		}
		if validator.Power != power {
			return fmt.Errorf(
				"migration: consensus validator %d power %d does not match application power %d",
				i,
				validator.Power,
				power,
			)
		}
	}
	return nil
}

func exportedValidatorPower(validator truedemocracy.GenesisValidator) (int64, bool) {
	if validator.Active == nil {
		power := validator.Stake / rewards.StakeMin
		return power, !validator.Jailed && power > 0
	}
	return validator.Power, *validator.Active
}

func reconcileAuth(
	cdc codec.Codec,
	raw json.RawMessage,
	replacements map[string]OperatorMapping,
	targets map[string]struct{},
) (json.RawMessage, map[string]sdk.Coins, error) {
	var state authtypes.GenesisState
	if err := cdc.UnmarshalJSON(raw, &state); err != nil {
		return nil, nil, fmt.Errorf("migration: decode auth genesis: %w", err)
	}

	addresses := make(map[string]struct{}, len(state.Accounts))
	accountNumbers := make(map[uint64]struct{}, len(state.Accounts))
	moduleAccounts := make(map[string]sdk.Coins)
	for i, anyAccount := range state.Accounts {
		var account sdk.AccountI
		if err := cdc.UnpackAny(anyAccount, &account); err != nil {
			return nil, nil, fmt.Errorf("migration: unpack auth account %d: %w", i, err)
		}
		address := account.GetAddress().String()
		if address == "" {
			return nil, nil, fmt.Errorf("migration: auth account %d has no valid address", i)
		}
		if _, duplicate := addresses[address]; duplicate {
			return nil, nil, fmt.Errorf("migration: duplicate auth account %q", address)
		}
		addresses[address] = struct{}{}
		if _, duplicate := accountNumbers[account.GetAccountNumber()]; duplicate {
			return nil, nil, fmt.Errorf(
				"migration: duplicate auth account number %d",
				account.GetAccountNumber(),
			)
		}
		accountNumbers[account.GetAccountNumber()] = struct{}{}
		if _, target := targets[address]; target {
			return nil, nil, fmt.Errorf("migration: replacement auth account %q already exists", address)
		}
		if _, module := account.(authtypes.ModuleAccountI); module {
			moduleAccounts[address] = nil
		}
	}

	for i, anyAccount := range state.Accounts {
		var account sdk.AccountI
		if err := cdc.UnpackAny(anyAccount, &account); err != nil {
			return nil, nil, fmt.Errorf("migration: unpack auth account %d: %w", i, err)
		}
		mapping, mapped := replacements[account.GetAddress().String()]
		if !mapped {
			continue
		}
		base, ok := account.(*authtypes.BaseAccount)
		if !ok {
			return nil, nil, fmt.Errorf(
				"migration: mapped operator %q uses unsupported auth account type %T",
				mapping.OldOperator,
				account,
			)
		}
		freshKey, err := freshPubKey(mapping.PubKeyType, mapping.PubKey)
		if err != nil {
			return nil, nil, fmt.Errorf("migration: build fresh auth public key: %w", err)
		}
		packedKey, err := codectypes.NewAnyWithValue(freshKey)
		if err != nil {
			return nil, nil, fmt.Errorf("migration: pack fresh auth public key: %w", err)
		}
		rewritten := &authtypes.BaseAccount{
			Address:       mapping.NewOperator,
			PubKey:        packedKey,
			AccountNumber: base.AccountNumber,
			Sequence:      base.Sequence,
		}
		packedAccount, err := codectypes.NewAnyWithValue(rewritten)
		if err != nil {
			return nil, nil, fmt.Errorf("migration: pack rewritten auth account: %w", err)
		}
		state.Accounts[i] = packedAccount
	}
	if err := validateAuthAccountNumbers(cdc, state.Accounts); err != nil {
		return nil, nil, err
	}
	if err := authtypes.ValidateGenesis(state); err != nil {
		return nil, nil, fmt.Errorf("migration: rewritten auth genesis invalid: %w", err)
	}
	output, err := cdc.MarshalJSON(&state)
	if err != nil {
		return nil, nil, fmt.Errorf("migration: encode rewritten auth genesis: %w", err)
	}
	return output, moduleAccounts, nil
}

func validateAuthAccountNumbers(cdc codec.Codec, accounts []*codectypes.Any) error {
	numbers := make(map[uint64]struct{}, len(accounts))
	for i, anyAccount := range accounts {
		var account sdk.AccountI
		if err := cdc.UnpackAny(anyAccount, &account); err != nil {
			return fmt.Errorf("migration: unpack rewritten auth account %d: %w", i, err)
		}
		if _, duplicate := numbers[account.GetAccountNumber()]; duplicate {
			return fmt.Errorf("migration: duplicate rewritten auth account number %d", account.GetAccountNumber())
		}
		numbers[account.GetAccountNumber()] = struct{}{}
	}
	return nil
}

func reconcileBank(
	cdc codec.Codec,
	raw json.RawMessage,
	replacements map[string]OperatorMapping,
	targets map[string]struct{},
	moduleAccounts map[string]sdk.Coins,
) (json.RawMessage, error) {
	var state banktypes.GenesisState
	if err := cdc.UnmarshalJSON(raw, &state); err != nil {
		return nil, fmt.Errorf("migration: decode bank genesis: %w", err)
	}
	if err := state.Validate(); err != nil {
		return nil, fmt.Errorf("migration: source bank genesis invalid: %w", err)
	}
	sourceSupply := state.Supply
	sourceModuleBalances := make(map[string]sdk.Coins, len(moduleAccounts))
	for i := range state.Balances {
		balance := &state.Balances[i]
		if _, target := targets[balance.Address]; target {
			return nil, fmt.Errorf("migration: replacement bank balance %q already exists", balance.Address)
		}
		if _, module := moduleAccounts[balance.Address]; module {
			sourceModuleBalances[balance.Address] = balance.Coins
		}
		if mapping, mapped := replacements[balance.Address]; mapped {
			balance.Address = mapping.NewOperator
		}
	}
	if err := state.Validate(); err != nil {
		return nil, fmt.Errorf("migration: rewritten bank genesis invalid: %w", err)
	}
	if !state.Supply.Equal(sourceSupply) {
		return nil, fmt.Errorf("migration: bank supply changed during reconciliation")
	}
	for _, balance := range state.Balances {
		if before, module := sourceModuleBalances[balance.Address]; module && !balance.Coins.Equal(before) {
			return nil, fmt.Errorf("migration: module balance %q changed", balance.Address)
		}
	}
	output, err := cdc.MarshalJSON(&state)
	if err != nil {
		return nil, fmt.Errorf("migration: encode rewritten bank genesis: %w", err)
	}
	return output, nil
}

func reconcileDEX(
	raw json.RawMessage,
	replacements map[string]OperatorMapping,
	targets map[string]struct{},
) (json.RawMessage, error) {
	var state dex.GenesisState
	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, fmt.Errorf("migration: decode dex genesis: %w", err)
	}
	if err := dex.ValidateGenesisState(state); err != nil {
		return nil, fmt.Errorf("migration: source dex genesis invalid: %w", err)
	}
	for i := range state.LPPositions {
		if _, target := targets[state.LPPositions[i].Provider]; target {
			return nil, fmt.Errorf(
				"migration: replacement DEX provider %q already exists",
				state.LPPositions[i].Provider,
			)
		}
		if mapping, mapped := replacements[state.LPPositions[i].Provider]; mapped {
			state.LPPositions[i].Provider = mapping.NewOperator
		}
	}
	for i := range state.RegisteredAssets {
		if _, target := targets[state.RegisteredAssets[i].RegisteredBy]; target {
			return nil, fmt.Errorf(
				"migration: replacement DEX authority %q already exists",
				state.RegisteredAssets[i].RegisteredBy,
			)
		}
		if mapping, mapped := replacements[state.RegisteredAssets[i].RegisteredBy]; mapped {
			state.RegisteredAssets[i].RegisteredBy = mapping.NewOperator
		}
	}
	if err := dex.ValidateGenesisState(state); err != nil {
		return nil, fmt.Errorf("migration: rewritten dex genesis invalid: %w", err)
	}
	output, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("migration: encode rewritten dex genesis: %w", err)
	}
	return output, nil
}

func enforceEmptyWasm(cdc codec.Codec, raw json.RawMessage) error {
	var state wasmtypes.GenesisState
	if err := cdc.UnmarshalJSON(raw, &state); err != nil {
		return fmt.Errorf("migration: decode wasm genesis: %w", err)
	}
	if len(state.Codes) != 0 || len(state.Contracts) != 0 {
		return fmt.Errorf(
			"migration: generic authority transform refuses wasm state with %d codes and %d contracts",
			len(state.Codes),
			len(state.Contracts),
		)
	}
	return nil
}
