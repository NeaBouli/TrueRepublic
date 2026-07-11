package truedemocracy

import (
	"encoding/hex"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	rewards "truerepublic/treasury/keeper"
)

// ValidateGenesisState validates all custom state that can create treasury or
// stake claims. Cross-module bank backing is checked by the application.
func ValidateGenesisState(genesis GenesisState) error {
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
		if len(validator.PubKey) != 32 {
			return fmt.Errorf("validator %q pubkey must be 32 bytes", validator.OperatorAddr)
		}
		pubKey := hex.EncodeToString(validator.PubKey)
		if operator, exists := pubKeys[pubKey]; exists {
			return fmt.Errorf("duplicate validator pubkey for %q and %q", operator, validator.OperatorAddr)
		}
		if validator.Stake < rewards.StakeMin {
			return fmt.Errorf("validator %q stake %d is below minimum %d", validator.OperatorAddr, validator.Stake, rewards.StakeMin)
		}
		domain, found := domains[validator.Domain]
		if !found {
			return fmt.Errorf("validator %q references missing domain %q", validator.OperatorAddr, validator.Domain)
		}
		if !containsString(domain.Members, validator.OperatorAddr) {
			return fmt.Errorf("validator %q is not a member of domain %q", validator.OperatorAddr, validator.Domain)
		}
		operators[validator.OperatorAddr] = struct{}{}
		pubKeys[pubKey] = validator.OperatorAddr
	}

	if genesis.VerifyingKeyHex != "" {
		verifyingKey, err := hex.DecodeString(genesis.VerifyingKeyHex)
		if err != nil {
			return fmt.Errorf("invalid verifying key hex: %w", err)
		}
		if _, err := DeserializeVerifyingKey(verifyingKey); err != nil {
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
			}
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
