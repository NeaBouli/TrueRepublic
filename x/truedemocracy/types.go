package truedemocracy

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"truerepublic/token"
)

const ModuleName = "truedemocracy"
const PNYXDenom = token.BaseDenom
const PNYXUnit = token.WholeTokenBaseUnits

// Proof of Domain slashing and reward parameters.
const (
	SlashFractionDowntime   int64 = 1    // 1% of stake slashed for downtime
	SlashFractionDoubleSign int64 = 5    // 5% of stake slashed for equivocation
	DowntimeJailDuration    int64 = 600  // seconds (10 min)
	SignedBlocksWindow      int64 = 100  // blocks tracked for liveness
	MinSignedPerWindow      int64 = 50   // must sign ≥50% of blocks in window
	RewardInterval          int64 = 3600 // distribute rewards every hour (seconds)
)

// Suggestion lifecycle parameters (whitepaper §3.1.2).
const (
	DefaultApprovalThresholdBps int64 = 500   // 5% in basis points
	DefaultDwellTimeSecs        int64 = 86400 // 1 day
	DeleteMajorityBps           int64 = 6667  // 2/3 ≈ 66.67% in basis points
)

// Governance parameters (whitepaper §3, §3.6).
const (
	// MerkleRootHistorySize is the maximum number of recent Merkle roots retained.
	// Proofs generated against any root in this window are accepted.
	MerkleRootHistorySize = 10

	InactivityTimeoutSecs int64 = 31_104_000 // 360 days
	ExcludeMajorityBps    int64 = 6667       // 2/3 ≈ 66.67% in basis points
	StakeTransferLimitBps int64 = 1000       // 10% of domain total payouts (WP §7)
)

// VotingMode determines how a winner is decided in person elections (WP §3.7).
type VotingMode int32

const (
	VotingModeSimpleMajority     VotingMode = 0 // >50% of votes cast (excl. abstentions)
	VotingModeAbsoluteMajority   VotingMode = 1 // >50% of all eligible members
	VotingModeSystemicConsensing VotingMode = 2 // -5 to +5 rating scale (WP §3.2)
)

// VoteChoice represents a member's vote in a person election (WP §3.7).
type VoteChoice int32

const (
	VoteChoiceApprove VoteChoice = 0 // place stone / vote for candidate
	VoteChoiceAbstain VoteChoice = 1 // explicit abstention
)

type Domain struct {
	Name             string         `json:"name"`
	Admin            sdk.AccAddress `json:"admin"`
	Members          []string       `json:"members"`
	Treasury         sdk.Coins      `json:"treasury"`
	Issues           []Issue        `json:"issues"`
	Options          DomainOptions  `json:"options"`
	PermissionReg    []string       `json:"permission_reg"`
	TotalPayouts     int64          `json:"total_payouts"`     // cumulative PNYX paid out (rewards, etc.)
	TransferredStake int64          `json:"transferred_stake"` // cumulative PNYX withdrawn by validators
	// v0.3.0 ZKP fields (backward compatible — zero values for existing domains).
	IdentityCommits   []string `json:"identity_commits"`    // MiMC commitments (hex)
	MerkleRoot        string   `json:"merkle_root"`         // current Merkle root (hex)
	MerkleRootHistory []string `json:"merkle_root_history"` // recent past Merkle roots
}

type DomainOptions struct {
	AdminElectable    bool       `json:"admin_electable"`
	AnyoneCanJoin     bool       `json:"anyone_can_join"`
	OnlyAdminIssues   bool       `json:"only_admin_issues"`
	CoinBurnRequired  bool       `json:"coin_burn_required"`
	ApprovalThreshold int64      `json:"approval_threshold"` // basis points; 0 = use default (500 = 5%)
	DefaultDwellTime  int64      `json:"default_dwell_time"` // seconds; 0 = use default (86400 = 1 day)
	VotingMode        VotingMode `json:"voting_mode"`        // person election mode (WP §3.7); 0 = simple majority
	AbstentionAllowed bool       `json:"abstention_allowed"` // allow explicit abstention in elections (WP §3.7); default true
}

type Issue struct {
	Name           string       `json:"name"`
	Stones         int          `json:"stones"`
	Suggestions    []Suggestion `json:"suggestions"`
	CreationDate   int64        `json:"creation_date"`    // unix timestamp
	LastActivityAt int64        `json:"last_activity_at"` // updated on any interaction
	ExternalLink   string       `json:"external_link"`    // optional URL to forum/discussion
}

type Suggestion struct {
	Name            string   `json:"name"`
	Creator         string   `json:"creator"`
	Stones          int      `json:"stones"`
	Ratings         []Rating `json:"ratings"`
	Color           string   `json:"color"`
	DwellTime       int64    `json:"dwell_time"`
	CreationDate    int64    `json:"creation_date"`     // unix timestamp
	ExternalLink    string   `json:"external_link"`     // optional URL to details/arguments
	EnteredYellowAt int64    `json:"entered_yellow_at"` // when suggestion entered yellow zone
	EnteredRedAt    int64    `json:"entered_red_at"`    // when suggestion entered red zone
	DeleteVotes     int      `json:"delete_votes"`      // fast-delete vote counter
}

type Rating struct {
	DomainPubKeyHex string `json:"domain_pub_key_hex"` // legacy ed25519 domain key (hex), empty for ZKP
	NullifierHex    string `json:"nullifier_hex"`      // ZKP nullifier (hex), empty for legacy
	Value           int    `json:"value"`
}

// VoteCommitment records a domain-key-signed vote without revealing voter identity.
type VoteCommitment struct {
	DomainPubKey string `json:"domain_pub_key"` // hex-encoded domain ed25519 pubkey
	Signature    string `json:"signature"`      // hex-encoded signature over vote payload
}

// Validator represents an active Proof of Domain validator node.
type Validator struct {
	OperatorAddr string    `json:"operator_addr"`
	PubKey       []byte    `json:"pub_key"`
	Stake        sdk.Coins `json:"stake"`
	Domains      []string  `json:"domains"`
	Power        int64     `json:"power"`
	Jailed       bool      `json:"jailed"`
	JailedUntil  int64     `json:"jailed_until"`
	MissedBlocks int64     `json:"missed_blocks"`
}

// BigPurgeSchedule tracks automated purge timing for a domain (WP S4).
// After each purge, all members must re-register fresh domain keys.
type BigPurgeSchedule struct {
	DomainName       string `json:"domain_name"`
	NextPurgeTime    int64  `json:"next_purge_time"`   // unix timestamp
	PurgeInterval    int64  `json:"purge_interval"`    // seconds, default 7776000 (90 days)
	AnnouncementLead int64  `json:"announcement_lead"` // seconds, default 604800 (7 days)
}

// OnboardingRequest tracks a pending domain key registration (WP S4).
// Step 1: member submits request with new domain key.
// Step 2: admin approves, key is added to permission register.
type OnboardingRequest struct {
	DomainName      string `json:"domain_name"`
	RequesterAddr   string `json:"requester_addr"`
	DomainPubKeyHex string `json:"domain_pub_key_hex"`
	RequestedAt     int64  `json:"requested_at"` // unix timestamp
	Status          string `json:"status"`       // "pending", "approved", "rejected"
}

// ZKPDomainState is a lightweight projection of domain ZKP fields for query responses.
type ZKPDomainState struct {
	DomainName        string   `json:"domain_name"`
	MerkleRoot        string   `json:"merkle_root"`
	MerkleRootHistory []string `json:"merkle_root_history"`
	CommitmentCount   int      `json:"commitment_count"`
	MemberCount       int      `json:"member_count"`
	VKInitialized     bool     `json:"vk_initialized"`
}

// NullifierRecord tracks a used nullifier to prevent double-voting with ZKP.
// KV key: "nullifier:{domain}:{nullifierHex}"
type NullifierRecord struct {
	DomainName    string `json:"domain_name"`
	NullifierHash string `json:"nullifier_hash"` // hex-encoded
	UsedAtHeight  int64  `json:"used_at_height"` // block height when consumed
}

// GenesisValidator is the genesis-file representation of a validator.
type GenesisValidator struct {
	OperatorAddr string `json:"operator_addr"`
	PubKey       []byte `json:"pub_key"`
	Stake        int64  `json:"stake"`
	Domain       string `json:"domain"`
}

// RevokedValidatorKey permanently retires a consensus key. Retired keys can
// never be registered again, which prevents a compromised key from returning
// after an operator has rotated away from it.
type RevokedValidatorKey struct {
	PubKey          []byte `json:"pub_key"`
	OperatorAddr    string `json:"operator_addr"`
	RevokedAtHeight int64  `json:"revoked_at_height"`
}

// PendingValidatorKeyRotation covers CometBFT's delayed validator-update
// activation window. During this window the old consensus key still resolves
// to the operator for evidence and liveness attribution, and another rotation
// is rejected.
type PendingValidatorKeyRotation struct {
	OperatorAddr     string `json:"operator_addr"`
	OldPubKey        []byte `json:"old_pub_key"`
	NewPubKey        []byte `json:"new_pub_key"`
	StartedHeight    int64  `json:"started_height"`
	ClearAfterHeight int64  `json:"clear_after_height"`
}

type GenesisState struct {
	Domains                    []Domain                      `json:"domains"`
	Validators                 []GenesisValidator            `json:"validators"`
	RevokedValidatorKeys       []RevokedValidatorKey         `json:"revoked_validator_keys"`
	PendingValidatorRotations  []PendingValidatorKeyRotation `json:"pending_validator_rotations"`
	BootstrapOperatorAddresses []string                      `json:"bootstrap_operator_addresses,omitempty"`
	UsedNullifiers             []NullifierRecord             `json:"used_nullifiers"`
	ZKPCircuitID               string                        `json:"zkp_circuit_id,omitempty"`
	VerifyingKeyHex            string                        `json:"verifying_key_hex,omitempty"`
	VerifyingKeySHA256         string                        `json:"verifying_key_sha256,omitempty"`
}

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(Domain{}, "truedemocracy/Domain", nil)
	cdc.RegisterConcrete(DomainOptions{}, "truedemocracy/DomainOptions", nil)
	cdc.RegisterConcrete(Issue{}, "truedemocracy/Issue", nil)
	cdc.RegisterConcrete(Suggestion{}, "truedemocracy/Suggestion", nil)
	cdc.RegisterConcrete(Rating{}, "truedemocracy/Rating", nil)
	cdc.RegisterConcrete(VoteCommitment{}, "truedemocracy/VoteCommitment", nil)
	cdc.RegisterConcrete(GenesisState{}, "truedemocracy/GenesisState", nil)
	cdc.RegisterConcrete(BigPurgeSchedule{}, "truedemocracy/BigPurgeSchedule", nil)
	cdc.RegisterConcrete(OnboardingRequest{}, "truedemocracy/OnboardingRequest", nil)
	cdc.RegisterConcrete(NullifierRecord{}, "truedemocracy/NullifierRecord", nil)
	cdc.RegisterConcrete(Validator{}, "truedemocracy/Validator", nil)
	cdc.RegisterConcrete(GenesisValidator{}, "truedemocracy/GenesisValidator", nil)
	cdc.RegisterConcrete(RevokedValidatorKey{}, "truedemocracy/RevokedValidatorKey", nil)
	cdc.RegisterConcrete(PendingValidatorKeyRotation{}, "truedemocracy/PendingValidatorKeyRotation", nil)

	// Message types for CLI transactions.
	cdc.RegisterConcrete(MsgCreateDomain{}, "truedemocracy/MsgCreateDomain", nil)
	cdc.RegisterConcrete(MsgSubmitProposal{}, "truedemocracy/MsgSubmitProposal", nil)
	cdc.RegisterConcrete(MsgRegisterValidator{}, "truedemocracy/MsgRegisterValidator", nil)
	cdc.RegisterConcrete(MsgWithdrawStake{}, "truedemocracy/MsgWithdrawStake", nil)
	cdc.RegisterConcrete(MsgRemoveValidator{}, "truedemocracy/MsgRemoveValidator", nil)
	cdc.RegisterConcrete(MsgRotateValidatorKey{}, "truedemocracy/MsgRotateValidatorKey", nil)
	cdc.RegisterConcrete(MsgUnjail{}, "truedemocracy/MsgUnjail", nil)
	cdc.RegisterConcrete(MsgJoinPermissionRegister{}, "truedemocracy/MsgJoinPermissionRegister", nil)
	cdc.RegisterConcrete(MsgPurgePermissionRegister{}, "truedemocracy/MsgPurgePermissionRegister", nil)
	cdc.RegisterConcrete(MsgPlaceStoneOnIssue{}, "truedemocracy/MsgPlaceStoneOnIssue", nil)
	cdc.RegisterConcrete(MsgPlaceStoneOnSuggestion{}, "truedemocracy/MsgPlaceStoneOnSuggestion", nil)
	cdc.RegisterConcrete(MsgPlaceStoneOnMember{}, "truedemocracy/MsgPlaceStoneOnMember", nil)
	cdc.RegisterConcrete(MsgVoteToExclude{}, "truedemocracy/MsgVoteToExclude", nil)
	cdc.RegisterConcrete(MsgVoteToDelete{}, "truedemocracy/MsgVoteToDelete", nil)
	cdc.RegisterConcrete(MsgRateProposal{}, "truedemocracy/MsgRateProposal", nil)
	cdc.RegisterConcrete(MsgCastElectionVote{}, "truedemocracy/MsgCastElectionVote", nil)
	cdc.RegisterConcrete(MsgAddMember{}, "truedemocracy/MsgAddMember", nil)
	cdc.RegisterConcrete(MsgOnboardToDomain{}, "truedemocracy/MsgOnboardToDomain", nil)
	cdc.RegisterConcrete(MsgApproveOnboarding{}, "truedemocracy/MsgApproveOnboarding", nil)
	cdc.RegisterConcrete(MsgRejectOnboarding{}, "truedemocracy/MsgRejectOnboarding", nil)
	cdc.RegisterConcrete(MsgRegisterIdentity{}, "truedemocracy/MsgRegisterIdentity", nil)
	cdc.RegisterConcrete(MsgRateWithProof{}, "truedemocracy/MsgRateWithProof", nil)
	cdc.RegisterConcrete(MsgDepositToDomain{}, "truedemocracy/MsgDepositToDomain", nil)
	cdc.RegisterConcrete(MsgWithdrawFromDomain{}, "truedemocracy/MsgWithdrawFromDomain", nil)
}

func DefaultGenesisState() GenesisState {
	return GenesisState{
		Domains:                   []Domain{},
		Validators:                []GenesisValidator{},
		RevokedValidatorKeys:      []RevokedValidatorKey{},
		PendingValidatorRotations: []PendingValidatorKeyRotation{},
	}
}
