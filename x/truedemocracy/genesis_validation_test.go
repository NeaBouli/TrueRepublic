package truedemocracy

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"

	rewards "truerepublic/treasury/keeper"
)

func validDemocracyGenesis() GenesisState {
	admin := sdk.AccAddress("genesis-admin")
	return GenesisState{
		Domains: []Domain{{
			Name:          "Test",
			Admin:         admin,
			Members:       []string{admin.String()},
			Treasury:      sdk.NewCoins(),
			Issues:        []Issue{},
			Options:       DomainOptions{AdminElectable: true},
			PermissionReg: []string{},
		}},
		Validators: []GenesisValidator{{
			OperatorAddr: admin.String(),
			PubKey:       ed25519.GenPrivKeyFromSecret([]byte("genesis-validation-test")).PubKey().Bytes(),
			Stake:        100_000 * PNYXUnit,
			Domain:       "Test",
		}},
	}
}

func TestValidateGenesisStateRejectsMalformedAndDuplicateDemocracyState(t *testing.T) {
	if err := ValidateGenesisState(validDemocracyGenesis()); err != nil {
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
			genesis := validDemocracyGenesis()
			tc.mutate(&genesis)
			if err := ValidateGenesisState(genesis); err == nil {
				t.Fatal("malformed genesis was accepted")
			}
		})
	}
}

func TestValidateGenesisStateRejectsInconsistentConsensusSlashingRelations(t *testing.T) {
	withActiveHistory := func() GenesisState {
		genesis := validDemocracyGenesis()
		validator := genesis.Validators[0]
		genesis.ConsensusKeyHistory = []ConsensusKeyRecord{{
			ConsensusAddress: consensusAddressFromPubKey(validator.PubKey),
			PubKey:           append([]byte(nil), validator.PubKey...),
			OperatorAddr:     validator.OperatorAddr,
			ActivatedHeight:  1,
		}}
		return genesis
	}
	if err := ValidateGenesisState(withActiveHistory()); err != nil {
		t.Fatalf("consistent active history rejected: %v", err)
	}
	withPendingRemoval := func() GenesisState {
		genesis := withActiveHistory()
		active := genesis.Validators[0]
		genesis.Validators = nil
		genesis.Domains[0].TotalPayouts = active.Stake * 10
		genesis.Domains[0].TransferredStake = active.Stake
		genesis.ConsensusKeyHistory[0].RetiredHeight = 12
		genesis.PendingValidatorRemovals = []PendingValidatorRemoval{{
			Validator: Validator{
				OperatorAddr: active.OperatorAddr,
				PubKey:       active.PubKey,
				Stake:        sdk.NewCoins(sdk.NewInt64Coin(PNYXDenom, active.Stake)),
				Domains:      []string{active.Domain},
				Power:        1,
			},
			RecipientAddr:          active.OperatorAddr,
			RemovedAtHeight:        10,
			RemovedAtTimeNanos:     1,
			ConsensusRetiredHeight: 12,
			ReleaseAfterHeight:     20,
		}}
		return genesis
	}
	if err := ValidateGenesisState(withPendingRemoval()); err != nil {
		t.Fatalf("consistent pending removal rejected: %v", err)
	}

	tests := []struct {
		name    string
		genesis func() GenesisState
	}{
		{"active key is retired", func() GenesisState {
			genesis := withActiveHistory()
			genesis.ConsensusKeyHistory[0].RetiredHeight = 2
			return genesis
		}},
		{"orphan signing info", func() GenesisState {
			genesis := withActiveHistory()
			genesis.ValidatorSigningInfos = []ValidatorSigningInfo{{
				OperatorAddr:             sdk.AccAddress("orphan-signer").String(),
				StartCommitHeight:        1,
				IndexOffset:              1,
				MissedBitmap:             make([]byte, livenessBitmapLength),
				LastObservedCommitHeight: 1,
			}}
			return genesis
		}},
		{"processed infraction lacks tombstone", func() GenesisState {
			genesis := withActiveHistory()
			record := genesis.ConsensusKeyHistory[0]
			genesis.ProcessedInfractions = []ProcessedInfraction{{
				ID:                  make([]byte, 32),
				MisbehaviorType:     1,
				ConsensusAddress:    append([]byte(nil), record.ConsensusAddress...),
				OperatorAddr:        record.OperatorAddr,
				InfractionHeight:    1,
				InfractionTimeNanos: 1,
				ObservedHeight:      2,
				ValidatorPower:      1,
				TotalVotingPower:    1,
				BurnedAmount:        1,
			}}
			return genesis
		}},
		{"pending removal retirement mismatch", func() GenesisState {
			genesis := withPendingRemoval()
			genesis.PendingValidatorRemovals[0].ConsensusRetiredHeight = 13
			return genesis
		}},
		{"pending removal lacks history", func() GenesisState {
			genesis := withPendingRemoval()
			genesis.ConsensusKeyHistory = nil
			return genesis
		}},
		{"pending removal has multiple domains", func() GenesisState {
			genesis := withPendingRemoval()
			genesis.PendingValidatorRemovals[0].Validator.Domains = []string{"Test", "Test"}
			return genesis
		}},
		{"pending removal exceeds transferred stake", func() GenesisState {
			genesis := withPendingRemoval()
			genesis.Domains[0].TransferredStake--
			return genesis
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateGenesisState(tc.genesis()); err == nil {
				t.Fatal("inconsistent consensus slashing state was accepted")
			}
		})
	}
}

func TestGenesisEscrowClaimsIncludesTreasuryAndStake(t *testing.T) {
	genesis := validDemocracyGenesis()
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

func explicitActiveFlag(active bool) *bool { return &active }

// validActiveInactiveGenesis carries one explicit active multi-domain
// validator and one explicit inactive retained claim (excluded, jailed).
func validActiveInactiveGenesis() GenesisState {
	admin := sdk.AccAddress("genesis-admin")
	activeOperator := sdk.AccAddress("gh60-active-operator").String()
	inactiveOperator := sdk.AccAddress("gh60-inactive-operator").String()
	return GenesisState{
		Domains: []Domain{
			{
				Name: "Test", Admin: admin,
				Members:  []string{admin.String(), activeOperator},
				Treasury: sdk.NewCoins(), Issues: []Issue{}, PermissionReg: []string{},
			},
			{
				Name: "Second", Admin: admin,
				Members:  []string{admin.String(), activeOperator},
				Treasury: sdk.NewCoins(), Issues: []Issue{}, PermissionReg: []string{},
			},
		},
		Validators: []GenesisValidator{
			{
				OperatorAddr: activeOperator,
				PubKey:       testPubKey("gh60-active"),
				Stake:        2 * rewards.StakeMin,
				Domain:       "Test",
				Domains:      []string{"Test", "Second"},
				Power:        2,
				Active:       explicitActiveFlag(true),
			},
			{
				OperatorAddr: inactiveOperator,
				PubKey:       testPubKey("gh60-inactive"),
				Stake:        rewards.StakeMin,
				Active:       explicitActiveFlag(false),
				Jailed:       true,
			},
		},
	}
}

func TestValidateGenesisStateAcceptsExplicitActiveAndInactiveClaims(t *testing.T) {
	if err := ValidateGenesisState(validActiveInactiveGenesis()); err != nil {
		t.Fatalf("explicit active/inactive genesis rejected: %v", err)
	}

	admin := sdk.AccAddress("genesis-admin")
	jailedOperator := sdk.AccAddress("gh60-jailed-operator").String()
	underStakedOperator := sdk.AccAddress("gh60-under-operator").String()
	nonMemberOperator := sdk.AccAddress("gh60-nonmember-operator").String()
	genesis := GenesisState{
		Domains: []Domain{{
			Name: "Test", Admin: admin,
			Members:  []string{admin.String(), jailedOperator},
			Treasury: sdk.NewCoins(), Issues: []Issue{}, PermissionReg: []string{},
		}},
		Validators: []GenesisValidator{
			{ // under-staked retained claim keeps its exact stake
				OperatorAddr: underStakedOperator,
				PubKey:       testPubKey("gh60-under-staked"),
				Stake:        rewards.StakeMin / 2,
				Active:       explicitActiveFlag(false),
				Jailed:       true,
			},
			{ // downtime-jailed claim retains stored power and domain
				OperatorAddr: jailedOperator,
				PubKey:       testPubKey("gh60-jailed"),
				Stake:        rewards.StakeMin,
				Domain:       "Test",
				Domains:      []string{"Test"},
				Power:        1,
				Active:       explicitActiveFlag(false),
				Jailed:       true,
				JailedUntil:  600,
			},
			{ // excluded from membership but still listing the domain
				OperatorAddr: nonMemberOperator,
				PubKey:       testPubKey("gh60-non-member"),
				Stake:        rewards.StakeMin,
				Domain:       "Test",
				Domains:      []string{"Test"},
				Active:       explicitActiveFlag(false),
				Jailed:       true,
			},
		},
	}
	if err := ValidateGenesisState(genesis); err != nil {
		t.Fatalf("legitimate inactive retained claims rejected: %v", err)
	}
}

func TestValidateGenesisStateRejectsMalformedResurrection(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*GenesisState)
	}{
		{"resurrect excluded claim as active", func(g *GenesisState) {
			g.Validators[1].Active = explicitActiveFlag(true)
		}},
		{"active flag with jail", func(g *GenesisState) {
			g.Validators[0].Jailed = true
		}},
		{"active flag with zero power", func(g *GenesisState) {
			g.Validators[0].Power = 0
		}},
		{"active power inconsistent with stake", func(g *GenesisState) {
			g.Validators[0].Power = 5
		}},
		{"active flag with under-minimum stake", func(g *GenesisState) {
			g.Validators[0].Stake = rewards.StakeMin / 2
		}},
		{"inactive flag contradicts active state", func(g *GenesisState) {
			g.Validators[0].Active = explicitActiveFlag(false)
		}},
		{"inactive unjailed positive power without domains", func(g *GenesisState) {
			g.Validators[1].Jailed = false
			g.Validators[1].Power = 1
		}},
		{"inactive power inconsistent with stake", func(g *GenesisState) {
			g.Validators[1].Power = 7
		}},
		{"legacy record mixed with domain list", func(g *GenesisState) {
			g.Validators[0].Active = nil
		}},
		{"legacy record mixed with explicit power", func(g *GenesisState) {
			g.Validators[1].Active = nil
			g.Validators[1].Power = 1
		}},
		{"primary domain contradicts domain list", func(g *GenesisState) {
			g.Validators[0].Domain = "Second"
		}},
		{"domain list without primary domain", func(g *GenesisState) {
			g.Validators[0].Domain = ""
		}},
		{"duplicate domain in list", func(g *GenesisState) {
			g.Validators[0].Domains = []string{"Test", "Test"}
		}},
		{"inactive references missing domain", func(g *GenesisState) {
			g.Validators[1].Domain = "missing"
			g.Validators[1].Domains = []string{"missing"}
		}},
		{"active lacks membership in one domain", func(g *GenesisState) {
			g.Domains[1].Members = []string{g.Domains[1].Admin.String()}
		}},
		{"revoked key reused by inactive claim", func(g *GenesisState) {
			g.RevokedValidatorKeys = []RevokedValidatorKey{{
				PubKey:          append([]byte(nil), g.Validators[1].PubKey...),
				OperatorAddr:    g.Validators[1].OperatorAddr,
				RevokedAtHeight: 1,
			}}
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			genesis := validActiveInactiveGenesis()
			tc.mutate(&genesis)
			if err := ValidateGenesisState(genesis); err == nil {
				t.Fatal("malformed active/inactive genesis was accepted")
			}
		})
	}
}
