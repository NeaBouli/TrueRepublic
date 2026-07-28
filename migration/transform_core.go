package migration

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/cometbft/cometbft/crypto/tmhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkbech32 "github.com/cosmos/cosmos-sdk/types/bech32"

	"truerepublic/x/truedemocracy"
)

// TransformTrueDemocracy applies a GH-61 descriptor-based rewrite to the
// truedemocracy module section of a full app-state export.
func TransformTrueDemocracy(desc *Descriptor, raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("migration: empty export payload")
	}
	if err := VerifySourceGenesis(desc, raw); err != nil {
		return nil, err
	}

	var root map[string]json.RawMessage
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, fmt.Errorf("migration: malformed export JSON: %w", err)
	}

	appStateRaw, ok := root["app_state"]
	if !ok {
		return nil, fmt.Errorf("migration: export is missing app_state")
	}

	var appState map[string]json.RawMessage
	if err := json.Unmarshal(appStateRaw, &appState); err != nil {
		return nil, fmt.Errorf("migration: malformed app_state: %w", err)
	}

	truedStateRaw, ok := appState["truedemocracy"]
	if !ok {
		return nil, fmt.Errorf("migration: app_state is missing truedemocracy")
	}

	var state truedemocracy.GenesisState
	if err := json.Unmarshal(truedStateRaw, &state); err != nil {
		return nil, fmt.Errorf("migration: malformed truedemocracy state: %w", err)
	}

	consensusKeys, err := collectConsensusInventory(state)
	if err != nil {
		return nil, err
	}
	if err := Verify(desc, consensusKeys); err != nil {
		return nil, err
	}
	descriptorHRP, _, err := sdkbech32.DecodeAndConvert(desc.Mappings[0].OldOperator)
	if err != nil {
		return nil, fmt.Errorf("migration: decode descriptor account prefix: %w", err)
	}
	if configured := sdk.GetConfig().GetBech32AccountAddrPrefix(); descriptorHRP != configured {
		return nil, fmt.Errorf(
			"migration: descriptor account prefix %q does not match configured prefix %q",
			descriptorHRP,
			configured,
		)
	}

	rewritten, err := rewriteTrueDemocracy(state, desc.Mappings)
	if err != nil {
		return nil, err
	}

	truedStateRawOut, err := json.Marshal(rewritten)
	if err != nil {
		return nil, fmt.Errorf("migration: failed to encode transformed truedemocracy state: %w", err)
	}
	for _, mapping := range desc.Mappings {
		if bytes.Contains(truedStateRawOut, []byte(mapping.OldOperator)) {
			return nil, fmt.Errorf(
				"migration: old operator %q remains in transformed truedemocracy JSON",
				mapping.OldOperator,
			)
		}
	}
	appState["truedemocracy"] = truedStateRawOut

	appStateRawOut, err := json.Marshal(appState)
	if err != nil {
		return nil, fmt.Errorf("migration: failed to encode transformed app_state: %w", err)
	}
	root["app_state"] = appStateRawOut

	out, err := json.Marshal(root)
	if err != nil {
		return nil, fmt.Errorf("migration: failed to encode transformed export: %w", err)
	}
	return out, nil
}

func collectConsensusInventory(state truedemocracy.GenesisState) ([][]byte, error) {
	refs := consensusKeyRefs(state)
	keys := make([][]byte, 0, len(refs))
	for _, ref := range refs {
		if len(ref.key) != 32 {
			return nil, fmt.Errorf(
				"migration: %s: %w",
				ref.label,
				fmt.Errorf("migration: consensus key must be 32 bytes, got %d", len(ref.key)),
			)
		}
		keys = append(keys, append([]byte(nil), ref.key...))
	}
	return keys, nil
}

// consensusKeyRef pairs a raw consensus public key with the exact state field
// it was read from, so validation diagnostics can name the offending field.
type consensusKeyRef struct {
	label string
	key   []byte
}

// consensusKeyRefs is the single traversal over every consensus public key
// collection in the truedemocracy state: validators, revoked validator keys,
// consensus key history, pending validator rotations (old and new keys), and
// pending validator removals. Both consensus inventory collectors share it so
// the covered collections and their order can never drift apart.
func consensusKeyRefs(state truedemocracy.GenesisState) []consensusKeyRef {
	refs := make([]consensusKeyRef, 0)
	for i, v := range state.Validators {
		refs = append(refs, consensusKeyRef{fmt.Sprintf("validator[%d].pub_key", i), v.PubKey})
	}
	for i, v := range state.RevokedValidatorKeys {
		refs = append(refs, consensusKeyRef{fmt.Sprintf("revoked_validator_keys[%d].pub_key", i), v.PubKey})
	}
	for i, v := range state.ConsensusKeyHistory {
		refs = append(refs, consensusKeyRef{fmt.Sprintf("consensus_key_history[%d].pub_key", i), v.PubKey})
	}
	for i, v := range state.PendingValidatorRotations {
		refs = append(refs, consensusKeyRef{fmt.Sprintf("pending_validator_rotations[%d].old_pub_key", i), v.OldPubKey})
		refs = append(refs, consensusKeyRef{fmt.Sprintf("pending_validator_rotations[%d].new_pub_key", i), v.NewPubKey})
	}
	for i, v := range state.PendingValidatorRemovals {
		refs = append(refs, consensusKeyRef{fmt.Sprintf("pending_validator_removals[%d].validator.pub_key", i), v.Validator.PubKey})
	}
	return refs
}

func rewriteTrueDemocracy(state truedemocracy.GenesisState, mappings []OperatorMapping) (truedemocracy.GenesisState, error) {
	mapping := make(map[string]string, len(mappings))
	mappedOld := make(map[string]struct{}, len(mappings))
	for _, m := range mappings {
		mapping[m.OldOperator] = m.NewOperator
		mappedOld[m.OldOperator] = struct{}{}
	}

	coupled := collectCoupledFromConsensus(state)
	for old := range mappedOld {
		if _, ok := coupled[old]; !ok {
			return truedemocracy.GenesisState{}, fmt.Errorf("migration: old operator %q is not present in coupled truedemocracy operator fields", old)
		}
	}
	for old := range coupled {
		if _, ok := mappedOld[old]; !ok {
			return truedemocracy.GenesisState{}, fmt.Errorf("migration: mapping for coupled old operator %q is missing", old)
		}
	}

	allTyped := collectOperatorOccurrences(state)
	for _, replacement := range mapping {
		if _, exists := allTyped[replacement]; exists {
			return truedemocracy.GenesisState{}, fmt.Errorf("migration: replacement operator %q collides with an existing typed address", replacement)
		}
	}

	rewrite := state
	rewrite.Validators = append([]truedemocracy.GenesisValidator(nil), state.Validators...)
	for i := range rewrite.Validators {
		rewrite.Validators[i].OperatorAddr = mapOperator(rewrite.Validators[i].OperatorAddr, mapping)
	}

	rewrite.BootstrapOperatorAddresses = append([]string(nil), state.BootstrapOperatorAddresses...)
	for i := range rewrite.BootstrapOperatorAddresses {
		rewrite.BootstrapOperatorAddresses[i] = mapOperator(rewrite.BootstrapOperatorAddresses[i], mapping)
	}

	rewrite.Domains = append([]truedemocracy.Domain(nil), state.Domains...)
	for i := range rewrite.Domains {
		if mapped := mapOptAddress(rewrite.Domains[i].Admin.String(), mapping); mapped != "" {
			_, address, err := sdkbech32.DecodeAndConvert(mapped)
			if err != nil || len(address) != addressLength {
				return truedemocracy.GenesisState{}, fmt.Errorf(
					"migration: replacement domain admin %q is invalid",
					mapped,
				)
			}
			rewrite.Domains[i].Admin = sdk.AccAddress(address)
		}
		members := append([]string(nil), rewrite.Domains[i].Members...)
		for j := range members {
			members[j] = mapOperator(members[j], mapping)
		}
		rewrite.Domains[i].Members = members

		issues := append([]truedemocracy.Issue(nil), rewrite.Domains[i].Issues...)
		for j := range issues {
			suggestions := append([]truedemocracy.Suggestion(nil), issues[j].Suggestions...)
			for k := range suggestions {
				suggestions[k].Creator = mapOperator(suggestions[k].Creator, mapping)
			}
			issues[j].Suggestions = suggestions
		}
		rewrite.Domains[i].Issues = issues
	}

	rewrite.RevokedValidatorKeys = append([]truedemocracy.RevokedValidatorKey(nil), state.RevokedValidatorKeys...)
	for i := range rewrite.RevokedValidatorKeys {
		rewrite.RevokedValidatorKeys[i].OperatorAddr = mapOperator(rewrite.RevokedValidatorKeys[i].OperatorAddr, mapping)
	}

	rewrite.PendingValidatorRotations = append([]truedemocracy.PendingValidatorKeyRotation(nil), state.PendingValidatorRotations...)
	for i := range rewrite.PendingValidatorRotations {
		rewrite.PendingValidatorRotations[i].OperatorAddr = mapOperator(rewrite.PendingValidatorRotations[i].OperatorAddr, mapping)
	}

	rewrite.ConsensusKeyHistory = append([]truedemocracy.ConsensusKeyRecord(nil), state.ConsensusKeyHistory...)
	for i := range rewrite.ConsensusKeyHistory {
		rewrite.ConsensusKeyHistory[i].OperatorAddr = mapOperator(rewrite.ConsensusKeyHistory[i].OperatorAddr, mapping)
	}

	rewrite.ValidatorSigningInfos = append([]truedemocracy.ValidatorSigningInfo(nil), state.ValidatorSigningInfos...)
	for i := range rewrite.ValidatorSigningInfos {
		rewrite.ValidatorSigningInfos[i].OperatorAddr = mapOperator(rewrite.ValidatorSigningInfos[i].OperatorAddr, mapping)
	}

	rewrite.ProcessedInfractions = append([]truedemocracy.ProcessedInfraction(nil), state.ProcessedInfractions...)
	for i := range rewrite.ProcessedInfractions {
		rewrite.ProcessedInfractions[i].OperatorAddr = mapOperator(rewrite.ProcessedInfractions[i].OperatorAddr, mapping)
	}

	rewrite.PendingValidatorRemovals = append([]truedemocracy.PendingValidatorRemoval(nil), state.PendingValidatorRemovals...)
	for i := range rewrite.PendingValidatorRemovals {
		rewrite.PendingValidatorRemovals[i].Validator.OperatorAddr = mapOperator(rewrite.PendingValidatorRemovals[i].Validator.OperatorAddr, mapping)
		rewrite.PendingValidatorRemovals[i].RecipientAddr = mapOperator(rewrite.PendingValidatorRemovals[i].RecipientAddr, mapping)
	}

	if err := ensureNoOldAddressesRemain(rewrite, mappedOld); err != nil {
		return truedemocracy.GenesisState{}, err
	}

	if err := truedemocracy.ValidateGenesisState(rewrite); err != nil {
		return truedemocracy.GenesisState{}, fmt.Errorf("migration: transformed truedemocracy genesis invalid: %w", err)
	}
	return rewrite, nil
}

func collectCoupledFromConsensus(state truedemocracy.GenesisState) map[string]struct{} {
	inventory := collectConsensusKeysRaw(state)
	invSet := make(map[string]struct{}, len(inventory))
	for _, key := range inventory {
		invSet[string(tmhash.SumTruncated(key))] = struct{}{}
	}

	all := collectOperatorOccurrences(state)
	coupled := make(map[string]struct{}, len(all))
	for addr := range all {
		_, addrBytes, err := sdkbech32.DecodeAndConvert(addr)
		if err != nil {
			continue
		}
		if _, ok := invSet[string(addrBytes)]; ok {
			coupled[addr] = struct{}{}
		}
	}
	return coupled
}

func collectOperatorOccurrences(state truedemocracy.GenesisState) map[string]int {
	occ := make(map[string]int)
	inc := func(addr string) {
		if addr == "" {
			return
		}
		occ[addr]++
	}

	for _, v := range state.Validators {
		inc(v.OperatorAddr)
	}
	for _, v := range state.BootstrapOperatorAddresses {
		inc(v)
	}
	for _, d := range state.Domains {
		inc(d.Admin.String())
		for _, m := range d.Members {
			inc(m)
		}
		for _, issue := range d.Issues {
			for _, suggestion := range issue.Suggestions {
				inc(suggestion.Creator)
			}
		}
	}
	for _, v := range state.RevokedValidatorKeys {
		inc(v.OperatorAddr)
	}
	for _, v := range state.PendingValidatorRotations {
		inc(v.OperatorAddr)
	}
	for _, v := range state.ConsensusKeyHistory {
		inc(v.OperatorAddr)
	}
	for _, v := range state.ValidatorSigningInfos {
		inc(v.OperatorAddr)
	}
	for _, v := range state.ProcessedInfractions {
		inc(v.OperatorAddr)
	}
	for _, v := range state.PendingValidatorRemovals {
		inc(v.Validator.OperatorAddr)
		inc(v.RecipientAddr)
	}
	return occ
}

func collectConsensusKeysRaw(state truedemocracy.GenesisState) [][]byte {
	refs := consensusKeyRefs(state)
	out := make([][]byte, 0, len(refs))
	for _, ref := range refs {
		out = append(out, append([]byte(nil), ref.key...))
	}
	return out
}

func ensureNoOldAddressesRemain(state truedemocracy.GenesisState, mappedOld map[string]struct{}) error {
	written := collectOperatorOccurrences(state)
	for old := range mappedOld {
		if count := written[old]; count > 0 {
			return fmt.Errorf("migration: old operator %q remains in transformed truedemocracy state", old)
		}
	}
	return nil
}

func mapOptAddress(address string, mapping map[string]string) string {
	if replacement, ok := mapping[address]; ok {
		return replacement
	}
	return ""
}

func mapOperator(address string, mapping map[string]string) string {
	if replacement, ok := mapping[address]; ok {
		return replacement
	}
	return address
}
