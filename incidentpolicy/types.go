// Package incidentpolicy validates a secret-free operations rehearsal
// contract. It describes roles, ordered phases, actions, evidence classes,
// approvals, and abort boundaries; it never performs an incident action.
package incidentpolicy

const (
	ContractVersion         = "truerepublic.incident-rehearsal/v1"
	MaintainedExerciseID    = "phase-6-operations-rehearsal"
	MaintainedScenarioCount = 8
	MaxContractBytes        = 256 * 1024
)

type RoleName string
type ScenarioKind string
type Severity string
type Outcome string
type Authority string
type PhaseName string
type Action string
type Evidence string
type AbortCondition string

const (
	RoleIncidentCommander   RoleName = "incident-commander"
	RoleProtocolOwner       RoleName = "protocol-owner"
	RoleSecurityOwner       RoleName = "security-owner"
	RoleValidatorOperator   RoleName = "validator-operator"
	RoleEvidenceCustodian   RoleName = "evidence-custodian"
	RoleCommunicationsOwner RoleName = "communications-owner"
	RoleReleaseOwner        RoleName = "release-owner"
)

const (
	AuthorityOperator                   Authority = "operator"
	AuthorityCoordinatedProtocolRelease Authority = "coordinated-protocol-release"
	AuthorityManualGovernance           Authority = "manual-governance"
)

const (
	KindChainHalt                   ScenarioKind = "chain-halt"
	KindValidatorFailure            ScenarioKind = "validator-failure"
	KindValidatorSlashing           ScenarioKind = "validator-slashing"
	KindConsensusKeyCompromise      ScenarioKind = "consensus-key-compromise"
	KindOperatorAuthorityCompromise ScenarioKind = "operator-authority-compromise"
	KindBackupRestore               ScenarioKind = "backup-restore"
	KindCompatibleBinaryUpgrade     ScenarioKind = "compatible-binary-upgrade"
	KindLegacyAuthorityMigration    ScenarioKind = "legacy-authority-migration"
)

const (
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
	OutcomeRestored  Outcome  = "service-restored"
	OutcomeSafeStop  Outcome  = "safe-stopped"
)

const (
	PhaseDetect   PhaseName = "detect"
	PhaseContain  PhaseName = "contain"
	PhasePreserve PhaseName = "preserve"
	PhaseDecide   PhaseName = "decide"
	PhaseRecover  PhaseName = "recover"
	PhaseValidate PhaseName = "validate"
	PhaseClose    PhaseName = "close"
)

type Contract struct {
	Version     string     `json:"version"`
	ExerciseID  string     `json:"exercise_id"`
	ChainID     string     `json:"chain_id"`
	Environment string     `json:"environment"`
	Defaults    Defaults   `json:"defaults"`
	Roles       []Role     `json:"roles"`
	Scenarios   []Scenario `json:"scenarios"`
}

type Defaults struct {
	Secrets                 string `json:"secrets"`
	Artifacts               string `json:"artifacts"`
	LiveActions             string `json:"live_actions"`
	SourceTargetConcurrency string `json:"source_target_concurrency"`
}

type Role struct {
	Name      RoleName `json:"name"`
	Primary   string   `json:"primary"`
	Secondary string   `json:"secondary"`
}

type Scenario struct {
	ID              string           `json:"id"`
	Kind            ScenarioKind     `json:"kind"`
	Severity        Severity         `json:"severity"`
	Authority       Authority        `json:"authority"`
	Runbook         string           `json:"runbook"`
	Outcome         Outcome          `json:"outcome"`
	Phases          []Phase          `json:"phases"`
	AbortConditions []AbortCondition `json:"abort_conditions"`
}

type Phase struct {
	Name      PhaseName  `json:"name"`
	Owner     RoleName   `json:"owner"`
	Actions   []Action   `json:"actions"`
	Evidence  []Evidence `json:"evidence"`
	Approvals []RoleName `json:"approvals"`
}

type Violation struct {
	Check   string `json:"check"`
	Message string `json:"message"`
}

type Report struct {
	Version       string      `json:"version"`
	ExerciseID    string      `json:"exercise_id"`
	ScenarioCount int         `json:"scenario_count"`
	Valid         bool        `json:"valid"`
	Violations    []Violation `json:"violations"`
}
