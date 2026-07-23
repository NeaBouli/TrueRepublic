package truedemocracy

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"cosmossdk.io/core/comet"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	rewards "truerepublic/treasury/keeper"
)

// ValidateGenesisState validates all custom state that can create treasury or
// stake claims. Cross-module bank backing is checked by the application.
func ValidateGenesisState(genesis GenesisState) error {
	if len(genesis.BootstrapOperatorAddresses) > 0 {
		if len(genesis.Domains) != 0 || len(genesis.Validators) != 0 {
			return fmt.Errorf("bootstrap operator addresses are only allowed before domains and validators are materialized")
		}
		seen := make(map[string]struct{}, len(genesis.BootstrapOperatorAddresses))
		for _, address := range genesis.BootstrapOperatorAddresses {
			if address == "" {
				return fmt.Errorf("bootstrap operator address is required")
			}
			if _, err := sdk.AccAddressFromBech32(address); err != nil {
				return fmt.Errorf("bootstrap operator address %q is invalid: %w", address, err)
			}
			if _, exists := seen[address]; exists {
				return fmt.Errorf("duplicate bootstrap operator address %q", address)
			}
			seen[address] = struct{}{}
		}
	}

	domains := make(map[string]Domain, len(genesis.Domains))
	for _, domain := range genesis.Domains {
		if err := validateGenesisDomain(domain); err != nil {
			return err
		}
		if _, exists := domains[domain.Name]; exists {
			return fmt.Errorf("duplicate domain %q", domain.Name)
		}
		domains[domain.Name] = domain
	}

	operators := make(map[string]struct{}, len(genesis.Validators))
	activeValidators := make(map[string]GenesisValidator, len(genesis.Validators))
	pubKeys := make(map[string]string, len(genesis.Validators))
	for _, validator := range genesis.Validators {
		if validator.OperatorAddr == "" {
			return fmt.Errorf("validator operator address is required")
		}
		if _, err := sdk.AccAddressFromBech32(validator.OperatorAddr); err != nil {
			return fmt.Errorf("validator operator address %q is invalid: %w", validator.OperatorAddr, err)
		}
		if _, exists := operators[validator.OperatorAddr]; exists {
			return fmt.Errorf("duplicate validator operator %q", validator.OperatorAddr)
		}
		operators[validator.OperatorAddr] = struct{}{}
		activeValidators[validator.OperatorAddr] = validator
	}
	for _, validator := range genesis.Validators {
		if len(validator.PubKey) != 32 {
			return fmt.Errorf("validator %q pubkey must be 32 bytes", validator.OperatorAddr)
		}
		pubKey := hex.EncodeToString(validator.PubKey)
		if operator, exists := pubKeys[pubKey]; exists {
			return fmt.Errorf("duplicate validator pubkey for %q and %q", operator, validator.OperatorAddr)
		}
		if validator.Stake <= 0 || (!validator.Jailed && validator.Stake < rewards.StakeMin) {
			return fmt.Errorf("validator %q stake %d is below minimum %d", validator.OperatorAddr, validator.Stake, rewards.StakeMin)
		}
		if validator.JailedUntil < 0 || validator.MissedBlocks < 0 ||
			validator.MissedBlocks > SignedBlocksWindow ||
			(!validator.Jailed && validator.JailedUntil != 0) {
			return fmt.Errorf("validator %q jail or liveness state is invalid", validator.OperatorAddr)
		}
		domain, found := domains[validator.Domain]
		if !found {
			return fmt.Errorf("validator %q references missing domain %q", validator.OperatorAddr, validator.Domain)
		}
		if !containsString(domain.Members, validator.OperatorAddr) {
			return fmt.Errorf("validator %q is not a member of domain %q", validator.OperatorAddr, validator.Domain)
		}
		pubKeys[pubKey] = validator.OperatorAddr
	}

	revokedPubKeys := make(map[string]RevokedValidatorKey, len(genesis.RevokedValidatorKeys))
	for _, record := range genesis.RevokedValidatorKeys {
		if len(record.PubKey) != 32 {
			return fmt.Errorf("revoked validator pubkey must be 32 bytes")
		}
		if record.OperatorAddr == "" {
			return fmt.Errorf("revoked validator operator address is required")
		}
		if _, err := sdk.AccAddressFromBech32(record.OperatorAddr); err != nil {
			return fmt.Errorf("revoked validator operator address %q is invalid: %w", record.OperatorAddr, err)
		}
		if record.RevokedAtHeight < 0 {
			return fmt.Errorf("revoked validator key height cannot be negative")
		}
		key := hex.EncodeToString(record.PubKey)
		if _, exists := revokedPubKeys[key]; exists {
			return fmt.Errorf("duplicate revoked validator pubkey %q", key)
		}
		if operator, active := pubKeys[key]; active {
			return fmt.Errorf("revoked validator pubkey is active for %q", operator)
		}
		revokedPubKeys[key] = record
	}
	for _, validator := range genesis.Validators {
		derived := consensusKeyDerivedOperator(validator.PubKey)
		if _, coupled := operators[derived]; coupled {
			return fmt.Errorf("validator operator %q collides with an active consensus-key authority", derived)
		}
	}
	for _, record := range genesis.RevokedValidatorKeys {
		derived := consensusKeyDerivedOperator(record.PubKey)
		if _, coupled := operators[derived]; coupled {
			return fmt.Errorf("validator operator %q collides with a revoked consensus-key authority", derived)
		}
	}

	pendingOperators := make(map[string]struct{}, len(genesis.PendingValidatorRotations))
	pendingPubKeys := make(map[string]string, len(genesis.PendingValidatorRotations)*2)
	for _, rotation := range genesis.PendingValidatorRotations {
		if _, err := sdk.AccAddressFromBech32(rotation.OperatorAddr); err != nil {
			return fmt.Errorf("pending rotation operator address %q is invalid: %w", rotation.OperatorAddr, err)
		}
		if _, exists := pendingOperators[rotation.OperatorAddr]; exists {
			return fmt.Errorf("duplicate pending rotation for operator %q", rotation.OperatorAddr)
		}
		if len(rotation.OldPubKey) != 32 || len(rotation.NewPubKey) != 32 {
			return fmt.Errorf("pending rotation keys for %q must be 32 bytes", rotation.OperatorAddr)
		}
		oldKey := hex.EncodeToString(rotation.OldPubKey)
		newKey := hex.EncodeToString(rotation.NewPubKey)
		if oldKey == newKey {
			return fmt.Errorf("pending rotation keys for %q must differ", rotation.OperatorAddr)
		}
		if rotation.StartedHeight < 0 || rotation.ClearAfterHeight < rotation.StartedHeight || rotation.ClearAfterHeight-rotation.StartedHeight < 2 {
			return fmt.Errorf("pending rotation heights for %q are invalid", rotation.OperatorAddr)
		}
		if activeOperator, active := pubKeys[newKey]; !active || activeOperator != rotation.OperatorAddr {
			return fmt.Errorf("pending rotation new key for %q is not its active validator key", rotation.OperatorAddr)
		}
		if _, active := pubKeys[oldKey]; active {
			return fmt.Errorf("pending rotation old key for %q is still active", rotation.OperatorAddr)
		}
		revocation, revoked := revokedPubKeys[oldKey]
		if !revoked || revocation.OperatorAddr != rotation.OperatorAddr {
			return fmt.Errorf("pending rotation old key for %q lacks a matching revocation", rotation.OperatorAddr)
		}
		for _, key := range []string{oldKey, newKey} {
			if other, exists := pendingPubKeys[key]; exists {
				return fmt.Errorf("pending rotation key is shared by %q and %q", other, rotation.OperatorAddr)
			}
			pendingPubKeys[key] = rotation.OperatorAddr
		}
		pendingOperators[rotation.OperatorAddr] = struct{}{}
	}

	pendingRemovalOperators := make(map[string]PendingValidatorRemoval, len(genesis.PendingValidatorRemovals))
	pendingStakeByDomain := make(map[string]math.Int)
	for _, removal := range genesis.PendingValidatorRemovals {
		operator := removal.Validator.OperatorAddr
		if _, err := sdk.AccAddressFromBech32(operator); err != nil {
			return fmt.Errorf("pending removal operator address %q is invalid: %w", operator, err)
		}
		if _, active := operators[operator]; active {
			return fmt.Errorf("pending removal operator %q is still active", operator)
		}
		if _, exists := pendingRemovalOperators[operator]; exists {
			return fmt.Errorf("duplicate pending removal for operator %q", operator)
		}
		if _, err := sdk.AccAddressFromBech32(removal.RecipientAddr); err != nil || removal.RecipientAddr != operator {
			return fmt.Errorf("pending removal recipient for %q must match its operator", operator)
		}
		if len(removal.Validator.PubKey) != 32 {
			return fmt.Errorf("pending removal pubkey for %q must be 32 bytes", operator)
		}
		if err := validatePNYXCoins(removal.Validator.Stake, "pending validator removal stake"); err != nil {
			return fmt.Errorf("pending removal for %q: %w", operator, err)
		}
		if !removal.Validator.Stake.AmountOf(PNYXDenom).IsInt64() {
			return fmt.Errorf("pending removal stake for %q exceeds supported range", operator)
		}
		if len(removal.Validator.Domains) != 1 {
			return fmt.Errorf("pending removal for %q must reference exactly one accounting domain", operator)
		}
		domainName := removal.Validator.Domains[0]
		if _, found := domains[domainName]; !found {
			return fmt.Errorf("pending removal for %q references missing domain %q", operator, domainName)
		}
		pendingStake := math.ZeroInt()
		if existing, found := pendingStakeByDomain[domainName]; found {
			pendingStake = existing
		}
		pendingStakeByDomain[domainName] = pendingStake.Add(removal.Validator.Stake.AmountOf(PNYXDenom))
		if removal.RemovedAtHeight < 0 ||
			removal.ConsensusRetiredHeight <= removal.RemovedAtHeight ||
			removal.ReleaseAfterHeight < removal.ConsensusRetiredHeight {
			return fmt.Errorf("pending removal heights for %q are invalid", operator)
		}
		if removal.RemovedAtTimeNanos < 0 ||
			(removal.ConsensusRetiredAtNanos == 0) != (removal.ReleaseAfterTimeNanos == 0) ||
			removal.ConsensusRetiredAtNanos < 0 ||
			removal.ReleaseAfterTimeNanos < removal.ConsensusRetiredAtNanos {
			return fmt.Errorf("pending removal times for %q are invalid", operator)
		}
		pendingRemovalOperators[operator] = removal
	}
	for domainName, pendingStake := range pendingStakeByDomain {
		domain := domains[domainName]
		if math.NewInt(domain.TransferredStake).LT(pendingStake) {
			return fmt.Errorf("domain %q transferred stake does not cover pending validator holds", domainName)
		}
	}
	if (len(genesis.PendingValidatorRemovals) > 0 ||
		len(genesis.ValidatorSigningInfos) > 0 ||
		len(genesis.ProcessedInfractions) > 0 ||
		len(genesis.RevokedValidatorKeys) > 0 ||
		len(genesis.PendingValidatorRotations) > 0) &&
		len(genesis.ConsensusKeyHistory) == 0 {
		return fmt.Errorf("consensus slashing state requires complete consensus key history")
	}

	consensusAddresses := make(map[string]ConsensusKeyRecord, len(genesis.ConsensusKeyHistory))
	consensusPubKeys := make(map[string]struct{}, len(genesis.ConsensusKeyHistory))
	for _, record := range genesis.ConsensusKeyHistory {
		if len(record.ConsensusAddress) != consensusAddressLength || len(record.PubKey) != 32 {
			return fmt.Errorf("consensus key history entry has invalid key material")
		}
		if !bytes.Equal(record.ConsensusAddress, consensusAddressFromPubKey(record.PubKey)) {
			return fmt.Errorf("consensus key history address does not match its pubkey")
		}
		if _, err := sdk.AccAddressFromBech32(record.OperatorAddr); err != nil {
			return fmt.Errorf("consensus key history operator %q is invalid: %w", record.OperatorAddr, err)
		}
		if record.ActivatedHeight <= 0 ||
			(record.RetiredHeight != 0 && record.RetiredHeight < record.ActivatedHeight) {
			return fmt.Errorf("consensus key history interval for %q is invalid", record.OperatorAddr)
		}
		key := hex.EncodeToString(record.ConsensusAddress)
		if _, exists := consensusAddresses[key]; exists {
			return fmt.Errorf("duplicate consensus key history address %q", key)
		}
		pubKey := hex.EncodeToString(record.PubKey)
		if _, exists := consensusPubKeys[pubKey]; exists {
			return fmt.Errorf("duplicate consensus key history pubkey %q", pubKey)
		}
		consensusAddresses[key] = record
		consensusPubKeys[pubKey] = struct{}{}
	}
	if len(genesis.ConsensusKeyHistory) > 0 {
		for operator, validator := range activeValidators {
			record, found := consensusAddresses[hex.EncodeToString(consensusAddressFromPubKey(validator.PubKey))]
			if !found || record.OperatorAddr != operator || !bytes.Equal(record.PubKey, validator.PubKey) ||
				record.RetiredHeight != 0 {
				return fmt.Errorf("active validator %q lacks a matching open consensus key record", operator)
			}
			if record.Tombstoned && !validator.Jailed {
				return fmt.Errorf("active validator %q uses a tombstoned key without being jailed", operator)
			}
		}
		for operator, removal := range pendingRemovalOperators {
			record, found := consensusAddresses[hex.EncodeToString(consensusAddressFromPubKey(removal.Validator.PubKey))]
			if !found || record.OperatorAddr != operator ||
				record.RetiredHeight != removal.ConsensusRetiredHeight {
				return fmt.Errorf("pending removal %q lacks matching retired consensus history", operator)
			}
		}
		for _, record := range genesis.ConsensusKeyHistory {
			if record.RetiredHeight != 0 {
				continue
			}
			validator, found := activeValidators[record.OperatorAddr]
			if !found || !bytes.Equal(validator.PubKey, record.PubKey) {
				return fmt.Errorf("open consensus key record for %q has no matching active validator", record.OperatorAddr)
			}
		}
		for _, revoked := range genesis.RevokedValidatorKeys {
			history, found := consensusAddresses[hex.EncodeToString(consensusAddressFromPubKey(revoked.PubKey))]
			wantRetiredHeight := revoked.RevokedAtHeight + sdk.ValidatorUpdateDelay + 1
			if !found || history.OperatorAddr != revoked.OperatorAddr ||
				!bytes.Equal(history.PubKey, revoked.PubKey) ||
				history.RetiredHeight != wantRetiredHeight {
				return fmt.Errorf("revoked consensus key for %q is inconsistent with key history", revoked.OperatorAddr)
			}
		}
		for _, rotation := range genesis.PendingValidatorRotations {
			transitionHeight := rotation.StartedHeight + sdk.ValidatorUpdateDelay + 1
			oldHistory, oldFound := consensusAddresses[hex.EncodeToString(consensusAddressFromPubKey(rotation.OldPubKey))]
			newHistory, newFound := consensusAddresses[hex.EncodeToString(consensusAddressFromPubKey(rotation.NewPubKey))]
			if !oldFound || oldHistory.OperatorAddr != rotation.OperatorAddr ||
				oldHistory.RetiredHeight != transitionHeight ||
				!newFound || newHistory.OperatorAddr != rotation.OperatorAddr ||
				newHistory.ActivatedHeight != transitionHeight {
				return fmt.Errorf("pending rotation for %q is inconsistent with key history", rotation.OperatorAddr)
			}
		}
	}

	signingOperators := make(map[string]struct{}, len(genesis.ValidatorSigningInfos))
	for _, info := range genesis.ValidatorSigningInfos {
		if _, err := sdk.AccAddressFromBech32(info.OperatorAddr); err != nil {
			return fmt.Errorf("validator signing operator %q is invalid: %w", info.OperatorAddr, err)
		}
		if err := validateSigningInfo(info); err != nil {
			return fmt.Errorf("validator signing info for %q is invalid: %w", info.OperatorAddr, err)
		}
		if _, exists := signingOperators[info.OperatorAddr]; exists {
			return fmt.Errorf("duplicate validator signing info for %q", info.OperatorAddr)
		}
		if _, active := activeValidators[info.OperatorAddr]; !active {
			if _, pending := pendingRemovalOperators[info.OperatorAddr]; !pending {
				return fmt.Errorf("validator signing info for %q has no active or held claim", info.OperatorAddr)
			}
		}
		signingOperators[info.OperatorAddr] = struct{}{}
	}

	processedIDs := make(map[string]struct{}, len(genesis.ProcessedInfractions))
	for _, record := range genesis.ProcessedInfractions {
		if len(record.ID) != sha256.Size || len(record.ConsensusAddress) != consensusAddressLength {
			return fmt.Errorf("processed infraction has invalid identity material")
		}
		if record.MisbehaviorType != int32(comet.DuplicateVote) &&
			record.MisbehaviorType != int32(comet.LightClientAttack) {
			return fmt.Errorf("processed infraction has unsupported type %d", record.MisbehaviorType)
		}
		if _, err := sdk.AccAddressFromBech32(record.OperatorAddr); err != nil {
			return fmt.Errorf("processed infraction operator %q is invalid: %w", record.OperatorAddr, err)
		}
		if record.InfractionHeight <= 0 ||
			record.ObservedHeight <= record.InfractionHeight ||
			record.InfractionTimeNanos <= 0 ||
			record.ValidatorPower <= 0 ||
			record.TotalVotingPower < record.ValidatorPower ||
			record.BurnedAmount < 0 {
			return fmt.Errorf("processed infraction for %q is malformed", record.OperatorAddr)
		}
		key := hex.EncodeToString(record.ID)
		if _, exists := processedIDs[key]; exists {
			return fmt.Errorf("duplicate processed infraction %q", key)
		}
		if len(genesis.ConsensusKeyHistory) > 0 {
			history, found := consensusAddresses[hex.EncodeToString(record.ConsensusAddress)]
			if !found || history.OperatorAddr != record.OperatorAddr || !history.Tombstoned ||
				record.InfractionHeight < history.ActivatedHeight ||
				(history.RetiredHeight > 0 && record.InfractionHeight >= history.RetiredHeight) {
				return fmt.Errorf("processed infraction %q is inconsistent with consensus key history", key)
			}
		}
		processedIDs[key] = struct{}{}
	}

	if len(genesis.LastCommitCursor.Hash) == 0 {
		if genesis.LastCommitCursor.CommitHeight != 0 {
			return fmt.Errorf("last commit cursor height requires a hash")
		}
	} else if len(genesis.LastCommitCursor.Hash) != sha256.Size || genesis.LastCommitCursor.CommitHeight < 0 {
		return fmt.Errorf("last commit cursor is malformed")
	}

	usedNullifiers := make(map[string]struct{}, len(genesis.UsedNullifiers))
	for _, record := range genesis.UsedNullifiers {
		if _, exists := domains[record.DomainName]; !exists {
			return fmt.Errorf("used nullifier references missing domain %q", record.DomainName)
		}
		if err := validateCanonicalFieldHex(record.NullifierHash, "used nullifier", true); err != nil {
			return fmt.Errorf("domain %q: %w", record.DomainName, err)
		}
		if record.UsedAtHeight < 0 {
			return fmt.Errorf("domain %q used nullifier height cannot be negative", record.DomainName)
		}
		key := record.DomainName + "\x00" + record.NullifierHash
		if _, exists := usedNullifiers[key]; exists {
			return fmt.Errorf("duplicate used nullifier for domain %q", record.DomainName)
		}
		usedNullifiers[key] = struct{}{}
	}

	if genesis.VerifyingKeyHex == "" {
		if genesis.ZKPCircuitID != "" || genesis.VerifyingKeySHA256 != "" {
			return fmt.Errorf("ZKP circuit id and verifying key fingerprint require verifying key bytes")
		}
	} else {
		verifyingKey, err := hex.DecodeString(genesis.VerifyingKeyHex)
		if err != nil {
			return fmt.Errorf("invalid verifying key hex: %w", err)
		}
		if genesis.VerifyingKeyHex != hex.EncodeToString(verifyingKey) {
			return fmt.Errorf("verifying key hex must use canonical lowercase encoding")
		}
		if _, err := ValidateMembershipVerifyingKey(verifyingKey, genesis.ZKPCircuitID, genesis.VerifyingKeySHA256); err != nil {
			return fmt.Errorf("invalid verifying key: %w", err)
		}
	}
	return nil
}

// GenesisEscrowClaims returns all PNYX treasury and validator stake claims.
func GenesisEscrowClaims(genesis GenesisState) (math.Int, error) {
	if err := ValidateGenesisState(genesis); err != nil {
		return math.Int{}, err
	}
	claims := math.ZeroInt()
	for _, domain := range genesis.Domains {
		claims = claims.Add(domain.Treasury.AmountOf(PNYXDenom))
	}
	for _, validator := range genesis.Validators {
		claims = claims.Add(math.NewInt(validator.Stake))
	}
	for _, removal := range genesis.PendingValidatorRemovals {
		claims = claims.Add(removal.Validator.Stake.AmountOf(PNYXDenom))
	}
	return claims, nil
}

func validateGenesisDomain(domain Domain) error {
	if domain.Name == "" {
		return fmt.Errorf("domain name is required")
	}
	if domain.Admin.Empty() {
		return fmt.Errorf("domain %q admin is required", domain.Name)
	}
	if !domain.Treasury.IsValid() {
		return fmt.Errorf("domain %q treasury is invalid", domain.Name)
	}
	if !domain.Treasury.Empty() && (len(domain.Treasury) != 1 || domain.Treasury[0].Denom != PNYXDenom || !domain.Treasury[0].Amount.IsPositive()) {
		return fmt.Errorf("domain %q treasury must contain only positive %s", domain.Name, PNYXDenom)
	}
	if domain.TotalPayouts < 0 || domain.TransferredStake < 0 {
		return fmt.Errorf("domain %q payout counters cannot be negative", domain.Name)
	}
	if math.NewInt(domain.TransferredStake).MulRaw(10_000).GT(math.NewInt(domain.TotalPayouts).MulRaw(StakeTransferLimitBps)) {
		return fmt.Errorf("domain %q transferred stake exceeds its payout-backed limit", domain.Name)
	}
	if domain.Options.ApprovalThreshold < 0 || domain.Options.ApprovalThreshold > 10_000 ||
		domain.Options.DefaultDwellTime < 0 ||
		domain.Options.VotingMode < VotingModeSimpleMajority || domain.Options.VotingMode > VotingModeSystemicConsensing {
		return fmt.Errorf("domain %q options are invalid", domain.Name)
	}
	if err := validateUniqueStrings(domain.Name, "member", domain.Members); err != nil {
		return err
	}
	if !containsString(domain.Members, domain.Admin.String()) {
		return fmt.Errorf("domain %q admin is not a member", domain.Name)
	}
	for _, member := range domain.Members {
		if _, err := sdk.AccAddressFromBech32(member); err != nil {
			return fmt.Errorf("domain %q member %q is invalid: %w", domain.Name, member, err)
		}
	}
	if err := validateUniqueStrings(domain.Name, "permission entry", domain.PermissionReg); err != nil {
		return err
	}
	for _, key := range domain.PermissionReg {
		if err := validateCanonicalFieldHex(key, "permission entry", false); err != nil {
			return fmt.Errorf("domain %q: %w", domain.Name, err)
		}
	}
	if err := validateGenesisZKPState(domain); err != nil {
		return err
	}
	issues := make(map[string]struct{}, len(domain.Issues))
	for _, issue := range domain.Issues {
		if issue.Name == "" || issue.Stones < 0 || issue.CreationDate < 0 || issue.LastActivityAt < 0 {
			return fmt.Errorf("domain %q contains malformed issue %q", domain.Name, issue.Name)
		}
		if _, exists := issues[issue.Name]; exists {
			return fmt.Errorf("domain %q contains duplicate issue %q", domain.Name, issue.Name)
		}
		issues[issue.Name] = struct{}{}
		suggestions := make(map[string]struct{}, len(issue.Suggestions))
		for _, suggestion := range issue.Suggestions {
			if suggestion.Name == "" || suggestion.Creator == "" || suggestion.Stones < 0 || suggestion.DwellTime < 0 ||
				suggestion.CreationDate < 0 || suggestion.EnteredYellowAt < 0 || suggestion.EnteredRedAt < 0 || suggestion.DeleteVotes < 0 {
				return fmt.Errorf("domain %q issue %q contains malformed suggestion %q", domain.Name, issue.Name, suggestion.Name)
			}
			if _, err := sdk.AccAddressFromBech32(suggestion.Creator); err != nil {
				return fmt.Errorf("domain %q suggestion %q creator is invalid: %w", domain.Name, suggestion.Name, err)
			}
			if _, exists := suggestions[suggestion.Name]; exists {
				return fmt.Errorf("domain %q issue %q contains duplicate suggestion %q", domain.Name, issue.Name, suggestion.Name)
			}
			suggestions[suggestion.Name] = struct{}{}
			for _, rating := range suggestion.Ratings {
				if rating.Value < -5 || rating.Value > 5 {
					return fmt.Errorf("domain %q suggestion %q contains rating outside -5..5", domain.Name, suggestion.Name)
				}
				if rating.NullifierHex != "" {
					if rating.DomainPubKeyHex != "" {
						return fmt.Errorf("domain %q suggestion %q rating mixes ZKP and domain-key identity", domain.Name, suggestion.Name)
					}
					if err := validateCanonicalFieldHex(rating.NullifierHex, "rating nullifier", true); err != nil {
						return fmt.Errorf("domain %q suggestion %q: %w", domain.Name, suggestion.Name, err)
					}
				} else if rating.DomainPubKeyHex != "" {
					if err := validateCanonicalFieldHex(rating.DomainPubKeyHex, "rating domain public key", false); err != nil {
						return fmt.Errorf("domain %q suggestion %q: %w", domain.Name, suggestion.Name, err)
					}
				}
			}
		}
	}
	return nil
}

func validateGenesisZKPState(domain Domain) error {
	if len(domain.IdentityCommits) > 1<<MerkleTreeDepth {
		return fmt.Errorf("domain %q has too many identity commitments", domain.Name)
	}
	seenCommitments := make(map[string]struct{}, len(domain.IdentityCommits))
	leaves := make([][]byte, len(domain.IdentityCommits))
	for i, commitment := range domain.IdentityCommits {
		if err := validateCanonicalFieldHex(commitment, "identity commitment", true); err != nil {
			return fmt.Errorf("domain %q: %w", domain.Name, err)
		}
		if _, exists := seenCommitments[commitment]; exists {
			return fmt.Errorf("domain %q contains duplicate identity commitment %q", domain.Name, commitment)
		}
		seenCommitments[commitment] = struct{}{}
		leaves[i], _ = hex.DecodeString(commitment)
	}
	if len(domain.IdentityCommits) == 0 {
		if domain.MerkleRoot != "" || len(domain.MerkleRootHistory) != 0 {
			return fmt.Errorf("domain %q has Merkle roots without identity commitments", domain.Name)
		}
		return nil
	}
	if err := validateCanonicalFieldHex(domain.MerkleRoot, "Merkle root", true); err != nil {
		return fmt.Errorf("domain %q: %w", domain.Name, err)
	}
	tree := NewMerkleTree(MerkleTreeDepth)
	if err := tree.BuildFromLeaves(leaves); err != nil {
		return fmt.Errorf("domain %q identity tree: %w", domain.Name, err)
	}
	if domain.MerkleRoot != tree.GetRoot() {
		return fmt.Errorf("domain %q Merkle root does not match identity commitments", domain.Name)
	}
	if len(domain.MerkleRootHistory) > MerkleRootHistorySize {
		return fmt.Errorf("domain %q Merkle root history exceeds %d entries", domain.Name, MerkleRootHistorySize)
	}
	seenRoots := make(map[string]struct{}, len(domain.MerkleRootHistory))
	for _, root := range domain.MerkleRootHistory {
		if err := validateCanonicalFieldHex(root, "Merkle root history entry", true); err != nil {
			return fmt.Errorf("domain %q: %w", domain.Name, err)
		}
		if _, exists := seenRoots[root]; exists {
			return fmt.Errorf("domain %q contains duplicate Merkle root history entry %q", domain.Name, root)
		}
		seenRoots[root] = struct{}{}
	}
	return nil
}

func validateCanonicalFieldHex(value, field string, requireFieldElement bool) error {
	if len(value) != 64 {
		return fmt.Errorf("%s must be exactly 32 bytes", field)
	}
	decoded, err := hex.DecodeString(value)
	if err != nil {
		return fmt.Errorf("%s is invalid hex: %w", field, err)
	}
	if value != hex.EncodeToString(decoded) {
		return fmt.Errorf("%s must use canonical lowercase hex", field)
	}
	if requireFieldElement {
		if _, err := HexToFieldElement(value); err != nil {
			return fmt.Errorf("%s is not a canonical BN254 field element: %w", field, err)
		}
	}
	return nil
}

func validateUniqueStrings(domainName, field string, values []string) error {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "" {
			return fmt.Errorf("domain %q contains empty %s", domainName, field)
		}
		if _, exists := seen[value]; exists {
			return fmt.Errorf("domain %q contains duplicate %s %q", domainName, field, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
