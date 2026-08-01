package incidentpolicy

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	logicalNamePattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,62}[a-z0-9])?$`)
	chainIDPattern     = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{2,63}$`)
)

var requiredRoles = []RoleName{
	RoleIncidentCommander,
	RoleProtocolOwner,
	RoleSecurityOwner,
	RoleValidatorOperator,
	RoleEvidenceCustodian,
	RoleCommunicationsOwner,
	RoleReleaseOwner,
}

var requiredRoleAssignments = map[RoleName][2]string{
	RoleIncidentCommander:   {"incident-primary", "incident-secondary"},
	RoleProtocolOwner:       {"protocol-primary", "protocol-secondary"},
	RoleSecurityOwner:       {"security-primary", "security-secondary"},
	RoleValidatorOperator:   {"validator-primary", "validator-secondary"},
	RoleEvidenceCustodian:   {"evidence-primary", "evidence-secondary"},
	RoleCommunicationsOwner: {"communications-primary", "communications-secondary"},
	RoleReleaseOwner:        {"release-primary", "release-secondary"},
}

var requiredPhaseOrder = []PhaseName{
	PhaseDetect,
	PhaseContain,
	PhasePreserve,
	PhaseDecide,
	PhaseRecover,
	PhaseValidate,
	PhaseClose,
}

var requiredScenarioKinds = []ScenarioKind{
	KindChainHalt,
	KindValidatorFailure,
	KindValidatorSlashing,
	KindConsensusKeyCompromise,
	KindOperatorAuthorityCompromise,
	KindBackupRestore,
	KindCompatibleBinaryUpgrade,
	KindLegacyAuthorityMigration,
}

type phaseRequirement struct {
	actions   []Action
	evidence  []Evidence
	approvals []RoleName
}

type scenarioSpec struct {
	severity        Severity
	authority       Authority
	outcome         Outcome
	runbook         string
	recoverOwner    RoleName
	requirements    map[PhaseName]phaseRequirement
	abortConditions []AbortCondition
}

var scenarioSpecs = map[ScenarioKind]scenarioSpec{
	KindChainHalt: {
		severity: SeverityCritical, authority: AuthorityCoordinatedProtocolRelease,
		outcome: OutcomeRestored, runbook: "docs/node-operators/operations/incident-command.md",
		recoverOwner: RoleProtocolOwner,
		requirements: map[PhaseName]phaseRequirement{
			PhaseContain: {actions: []Action{"declare-chain-halt", "disable-automatic-restart"}, evidence: []Evidence{"signer-isolation", "restart-disabled", "validator-power"}},
			PhaseRecover: {actions: []Action{"coordinate-restart"}},
		},
		abortConditions: []AbortCondition{"quorum-unsafe", "trusted-hash-disagreement", "signer-isolation-unproven", "second-signer-detected"},
	},
	KindValidatorFailure: {
		severity: SeverityHigh, authority: AuthorityOperator, outcome: OutcomeRestored,
		runbook:      "docs/node-operators/operations/validator-identity-recovery.md",
		recoverOwner: RoleValidatorOperator,
		requirements: map[PhaseName]phaseRequirement{
			PhaseContain:  {actions: []Action{"isolate-affected-signer", "disable-automatic-restart"}, evidence: []Evidence{"signer-isolation", "restart-disabled"}},
			PhaseRecover:  {actions: []Action{"repair-or-recover-validator"}},
			PhaseValidate: {evidence: []Evidence{"signer-state-monotonic", "one-active-signer", "validator-power"}},
		},
		abortConditions: []AbortCondition{"signer-isolation-unproven", "signer-state-stale", "second-signer-detected"},
	},
	KindValidatorSlashing: {
		severity: SeverityCritical, authority: AuthorityOperator, outcome: OutcomeRestored,
		runbook:      "docs/node-operators/operations/validator-slashing.md",
		recoverOwner: RoleValidatorOperator,
		requirements: map[PhaseName]phaseRequirement{
			PhaseContain:  {actions: []Action{"isolate-affected-signer", "disable-automatic-restart"}, evidence: []Evidence{"signer-isolation", "restart-disabled"}},
			PhasePreserve: {evidence: []Evidence{"offense-height", "evidence-hash", "stake-and-supply"}},
			PhaseDecide:   {actions: []Action{"evaluate-slashing-recovery"}, approvals: []RoleName{RoleSecurityOwner, RoleProtocolOwner}},
			PhaseRecover:  {actions: []Action{"repair-or-rotate-validator"}},
			PhaseValidate: {evidence: []Evidence{"one-active-signer", "validator-power", "stake-and-supply"}},
		},
		abortConditions: []AbortCondition{"signer-isolation-unproven", "signer-state-stale", "second-signer-detected", "evidence-incomplete"},
	},
	KindConsensusKeyCompromise: {
		severity: SeverityCritical, authority: AuthorityOperator, outcome: OutcomeRestored,
		runbook:      "docs/node-operators/operations/validator-key-rotation.md",
		recoverOwner: RoleValidatorOperator,
		requirements: map[PhaseName]phaseRequirement{
			PhaseContain:  {actions: []Action{"isolate-affected-signer", "disable-automatic-restart"}, evidence: []Evidence{"signer-isolation", "restart-disabled"}},
			PhaseDecide:   {approvals: []RoleName{RoleSecurityOwner, RoleProtocolOwner}},
			PhaseRecover:  {actions: []Action{"rotate-consensus-key"}},
			PhaseValidate: {evidence: []Evidence{"old-key-revoked", "rotation-activation-height", "one-active-signer", "signer-state-monotonic"}},
		},
		abortConditions: []AbortCondition{"operator-authority-uncertain", "signer-isolation-unproven", "second-signer-detected", "replacement-key-unavailable"},
	},
	KindOperatorAuthorityCompromise: {
		severity: SeverityCritical, authority: AuthorityManualGovernance, outcome: OutcomeSafeStop,
		runbook:      "docs/node-operators/operations/incident-command.md",
		recoverOwner: RoleSecurityOwner,
		requirements: map[PhaseName]phaseRequirement{
			PhaseContain:  {actions: []Action{"isolate-affected-signer", "disable-automatic-restart", "freeze-authority-actions"}, evidence: []Evidence{"signer-isolation", "restart-disabled", "authority-isolation"}},
			PhaseDecide:   {approvals: []RoleName{RoleSecurityOwner, RoleProtocolOwner, RoleReleaseOwner}},
			PhaseRecover:  {actions: []Action{"coordinate-manual-recovery"}},
			PhaseValidate: {evidence: []Evidence{"safe-stopped", "authority-isolation", "governance-decision"}},
		},
		abortConditions: []AbortCondition{"authority-still-active", "signer-isolation-unproven", "second-signer-detected", "unauthorized-rotation-attempt"},
	},
	KindBackupRestore: {
		severity: SeverityHigh, authority: AuthorityOperator, outcome: OutcomeRestored,
		runbook:      "docs/node-operators/operations/backup-recovery.md",
		recoverOwner: RoleValidatorOperator,
		requirements: map[PhaseName]phaseRequirement{
			PhaseContain:  {actions: []Action{"stop-backup-source"}, evidence: []Evidence{"restart-disabled"}},
			PhasePreserve: {evidence: []Evidence{"backup-sha256", "backup-excludes-keys"}},
			PhaseRecover:  {actions: []Action{"create-sanitized-backup", "restore-fresh-full-node"}},
			PhaseValidate: {evidence: []Evidence{"backup-excludes-keys", "fresh-node-identity", "genesis-sha256"}},
		},
		abortConditions: []AbortCondition{"secret-material-detected", "artifact-mismatch", "restore-target-not-fresh"},
	},
	KindCompatibleBinaryUpgrade: {
		severity: SeverityHigh, authority: AuthorityCoordinatedProtocolRelease,
		outcome: OutcomeRestored, runbook: "docs/node-operators/operations/upgrades.md",
		recoverOwner: RoleReleaseOwner,
		requirements: map[PhaseName]phaseRequirement{
			PhaseContain:  {actions: []Action{"stop-upgrade-node", "disable-automatic-restart"}, evidence: []Evidence{"restart-disabled"}},
			PhasePreserve: {evidence: []Evidence{"binary-sha256", "backup-sha256", "backup-excludes-keys", "signer-state-monotonic"}},
			PhaseRecover:  {actions: []Action{"install-compatible-binary", "rollback-before-state-open"}, evidence: []Evidence{"candidate-state-unopened"}},
			PhaseValidate: {evidence: []Evidence{"signer-state-monotonic", "one-active-signer", "validator-power"}},
		},
		abortConditions: []AbortCondition{"candidate-opened-state", "signer-state-stale", "second-signer-detected", "trusted-hash-disagreement"},
	},
	KindLegacyAuthorityMigration: {
		severity: SeverityCritical, authority: AuthorityManualGovernance,
		outcome: OutcomeRestored, runbook: "docs/node-operators/operations/legacy-authority-migration.md",
		recoverOwner: RoleProtocolOwner,
		requirements: map[PhaseName]phaseRequirement{
			PhaseContain:  {actions: []Action{"halt-source-chain", "disable-automatic-restart"}, evidence: []Evidence{"source-target-isolation", "restart-disabled"}},
			PhasePreserve: {evidence: []Evidence{"raw-export-sha256", "distinct-chain-ids", "source-target-isolation"}},
			PhaseDecide:   {approvals: []RoleName{RoleSecurityOwner, RoleProtocolOwner, RoleReleaseOwner}},
			PhaseRecover:  {actions: []Action{"transform-fresh-genesis", "start-isolated-target", "stop-target-chain", "rollback-separate-chain", "restart-untouched-source"}, evidence: []Evidence{"target-stopped"}},
			PhaseValidate: {evidence: []Evidence{"distinct-chain-ids", "source-target-isolation", "fresh-operator-proofs", "empty-wasm-state", "target-stopped"}},
		},
		abortConditions: []AbortCondition{"source-target-concurrency", "source-target-state-mixing", "target-still-active", "second-signer-detected", "artifact-mismatch", "cosmwasm-state-present", "trusted-hash-disagreement"},
	},
}

var supportedActions = setOf(
	Action("classify-incident"), "freeze-changes", "preserve-evidence", "record-decision",
	"validate-recovery", "close-incident", "declare-chain-halt", "disable-automatic-restart",
	"coordinate-restart", "isolate-affected-signer", "repair-or-recover-validator",
	"evaluate-slashing-recovery", "repair-or-rotate-validator", "rotate-consensus-key",
	"freeze-authority-actions", "coordinate-manual-recovery", "stop-backup-source",
	"create-sanitized-backup", "restore-fresh-full-node", "stop-upgrade-node",
	"install-compatible-binary", "rollback-before-state-open", "halt-source-chain",
	"transform-fresh-genesis", "start-isolated-target", "stop-target-chain",
	"rollback-separate-chain", "restart-untouched-source",
)

var supportedEvidence = setOf(
	Evidence("incident-start"), "alert-snapshot", "chain-id", "logs", "trusted-height",
	"trusted-app-hash", "approval-record", "decision-deadline", "communication-channel",
	"common-app-hash", "height-progress", "ledger-invariants", "postmortem",
	"signer-isolation", "restart-disabled", "validator-power", "signer-state-monotonic",
	"one-active-signer", "offense-height", "evidence-hash", "stake-and-supply",
	"old-key-revoked", "rotation-activation-height", "authority-isolation",
	"governance-decision", "safe-stopped", "backup-sha256", "backup-excludes-keys",
	"fresh-node-identity", "genesis-sha256", "binary-sha256", "candidate-state-unopened",
	"source-target-isolation", "distinct-chain-ids", "raw-export-sha256",
	"fresh-operator-proofs", "empty-wasm-state", "target-stopped",
)

var supportedAbortConditions = setOf(
	AbortCondition("quorum-unsafe"), "trusted-hash-disagreement", "signer-isolation-unproven",
	"signer-state-stale", "second-signer-detected", "evidence-incomplete",
	"operator-authority-uncertain", "replacement-key-unavailable", "authority-still-active",
	"unauthorized-rotation-attempt", "secret-material-detected", "artifact-mismatch",
	"restore-target-not-fresh", "candidate-opened-state", "source-target-concurrency",
	"source-target-state-mixing", "target-still-active", "cosmwasm-state-present",
	"invariant-failure",
)

type validator struct {
	violations []Violation
	roles      map[RoleName]bool
}

func Validate(contract Contract) Report {
	v := validator{roles: make(map[RoleName]bool)}
	v.checkHeader(contract)
	v.checkDefaults(contract.Defaults)
	v.checkRoles(contract.Roles)
	v.checkScenarios(contract.Scenarios)
	violations := make([]Violation, len(v.violations))
	copy(violations, v.violations)
	return Report{
		Version: ContractVersion, ExerciseID: MaintainedExerciseID,
		ScenarioCount: MaintainedScenarioCount, Valid: len(violations) == 0,
		Violations: violations,
	}
}

func (v *validator) fail(check, format string, args ...any) {
	v.violations = append(v.violations, Violation{Check: check, Message: fmt.Sprintf(format, args...)})
}

func (v *validator) checkHeader(contract Contract) {
	if contract.Version != ContractVersion {
		v.fail("version", "version must be %q", ContractVersion)
	}
	if contract.ExerciseID != MaintainedExerciseID {
		v.fail("exercise_id", "exercise_id must use the fixed synthetic rehearsal identifier")
	}
	if !chainIDPattern.MatchString(contract.ChainID) || !strings.HasPrefix(contract.ChainID, "truerepublic-rehearsal-") {
		v.fail("chain_id", "chain_id must be a canonical truerepublic-rehearsal-* logical identifier")
	}
	if contract.Environment != "synthetic-rehearsal" {
		v.fail("environment", "environment must be synthetic-rehearsal")
	}
}

func (v *validator) checkDefaults(defaults Defaults) {
	checks := []struct{ check, got, want string }{
		{"defaults.secrets", defaults.Secrets, "forbidden"},
		{"defaults.artifacts", defaults.Artifacts, "references-only"},
		{"defaults.live_actions", defaults.LiveActions, "forbidden"},
		{"defaults.source_target_concurrency", defaults.SourceTargetConcurrency, "forbidden"},
	}
	for _, check := range checks {
		if check.got != check.want {
			v.fail(check.check, "value must be %q", check.want)
		}
	}
}

func (v *validator) checkRoles(roles []Role) {
	seenAssignments := make(map[string]string)
	for i, role := range roles {
		prefix := fmt.Sprintf("roles[%d]", i)
		if !contains(requiredRoles, role.Name) {
			v.fail(prefix+".name", "role is not supported")
		} else if v.roles[role.Name] {
			v.fail(prefix+".name", "role is duplicated")
		} else {
			v.roles[role.Name] = true
		}
		assignments := []struct{ field, assignment string }{
			{"primary", role.Primary},
			{"secondary", role.Secondary},
		}
		expectedAssignments := requiredRoleAssignments[role.Name]
		for index, item := range assignments {
			field, assignment := item.field, item.assignment
			if !logicalNamePattern.MatchString(assignment) || assignment != expectedAssignments[index] {
				v.fail(prefix+"."+field, "assignment must use the fixed secret-free rehearsal seat for this role")
			}
			if previous, exists := seenAssignments[assignment]; exists {
				_ = previous
				v.fail(prefix+"."+field, "assignment is already used")
			} else {
				seenAssignments[assignment] = string(role.Name) + "." + field
			}
		}
		if role.Primary != "" && role.Primary == role.Secondary {
			v.fail(prefix+".secondary", "primary and secondary must be independent aliases")
		}
	}
	for _, role := range requiredRoles {
		if !v.roles[role] {
			v.fail("roles."+string(role), "required role %q is missing", role)
		}
	}
}

func (v *validator) checkScenarios(scenarios []Scenario) {
	seenIDs := make(map[string]bool)
	seenKinds := make(map[ScenarioKind]bool)
	for i, scenario := range scenarios {
		prefix := fmt.Sprintf("scenarios[%d]", i)
		if !logicalNamePattern.MatchString(scenario.ID) {
			v.fail(prefix+".id", "scenario ID must be a canonical logical identifier")
		} else if seenIDs[scenario.ID] {
			v.fail(prefix+".id", "scenario ID is duplicated")
		} else {
			seenIDs[scenario.ID] = true
		}
		spec, supported := scenarioSpecs[scenario.Kind]
		if !supported {
			v.fail(prefix+".kind", "scenario kind is not supported")
			v.checkPhases(prefix, scenario, scenarioSpec{})
			v.checkAbortConditions(prefix, scenario.AbortConditions, nil)
			continue
		}
		if seenKinds[scenario.Kind] {
			v.fail(prefix+".kind", "scenario kind %q is duplicated", scenario.Kind)
		}
		seenKinds[scenario.Kind] = true
		if scenario.ID != string(scenario.Kind)+"-rehearsal" {
			v.fail(prefix+".id", "scenario ID must use the fixed synthetic identifier for its kind")
		}
		if scenario.Severity != spec.severity {
			v.fail(prefix+".severity", "%s severity must be %q", scenario.Kind, spec.severity)
		}
		if scenario.Authority != spec.authority {
			v.fail(prefix+".authority", "%s authority must be %q", scenario.Kind, spec.authority)
		}
		if scenario.Outcome != spec.outcome {
			v.fail(prefix+".outcome", "%s outcome must be %q", scenario.Kind, spec.outcome)
		}
		if scenario.Runbook != spec.runbook {
			v.fail(prefix+".runbook", "%s runbook must be %q", scenario.Kind, spec.runbook)
		}
		v.checkPhases(prefix, scenario, spec)
		v.checkAbortConditions(prefix, scenario.AbortConditions, spec.abortConditions)
	}
	for _, kind := range requiredScenarioKinds {
		if !seenKinds[kind] {
			v.fail("scenarios."+string(kind), "required scenario kind %q is missing", kind)
		}
	}
}

func (v *validator) checkPhases(prefix string, scenario Scenario, spec scenarioSpec) {
	if len(scenario.Phases) != len(requiredPhaseOrder) {
		v.fail(prefix+".phases", "exactly %d ordered phases are required", len(requiredPhaseOrder))
	}
	seenActions := make(map[Action]bool)
	for i, phase := range scenario.Phases {
		phasePrefix := fmt.Sprintf("%s.phases[%d]", prefix, i)
		if i >= len(requiredPhaseOrder) || phase.Name != requiredPhaseOrder[i] {
			want := "no additional phase"
			if i < len(requiredPhaseOrder) {
				want = string(requiredPhaseOrder[i])
			}
			v.fail(phasePrefix+".name", "phase must be %q at sequence %d", want, i+1)
		}
		if !v.roles[phase.Owner] {
			v.fail(phasePrefix+".owner", "owner is not a declared required role")
		}
		for j, action := range phase.Actions {
			check := fmt.Sprintf("%s.actions[%d]", phasePrefix, j)
			if !supportedActions[action] {
				v.fail(check, "action is not supported")
			}
			if seenActions[action] {
				v.fail(check, "action is duplicated in the scenario")
			}
			seenActions[action] = true
		}
		seenEvidence := make(map[Evidence]bool)
		for j, evidence := range phase.Evidence {
			check := fmt.Sprintf("%s.evidence[%d]", phasePrefix, j)
			if !supportedEvidence[evidence] {
				v.fail(check, "evidence is not supported")
			}
			if seenEvidence[evidence] {
				v.fail(check, "evidence is duplicated in the phase")
			}
			seenEvidence[evidence] = true
		}
		seenApprovals := make(map[RoleName]bool)
		for j, approval := range phase.Approvals {
			check := fmt.Sprintf("%s.approvals[%d]", phasePrefix, j)
			if !v.roles[approval] {
				v.fail(check, "approval role is not declared")
			}
			if seenApprovals[approval] {
				v.fail(check, "approval role is duplicated in the phase")
			}
			seenApprovals[approval] = true
		}
	}

	common := map[PhaseName]phaseRequirement{
		PhaseDetect:   {actions: []Action{"classify-incident"}, evidence: []Evidence{"incident-start", "alert-snapshot", "chain-id"}},
		PhaseContain:  {actions: []Action{"freeze-changes"}},
		PhasePreserve: {actions: []Action{"preserve-evidence"}, evidence: []Evidence{"logs", "trusted-height", "trusted-app-hash"}},
		PhaseDecide:   {actions: []Action{"record-decision"}, evidence: []Evidence{"approval-record", "decision-deadline", "communication-channel"}, approvals: []RoleName{RoleIncidentCommander}},
		PhaseRecover:  {},
		PhaseValidate: {actions: []Action{"validate-recovery"}},
		PhaseClose:    {actions: []Action{"close-incident"}, evidence: []Evidence{"postmortem"}, approvals: []RoleName{RoleIncidentCommander}},
	}
	if scenario.Outcome == OutcomeRestored {
		common[PhaseValidate] = phaseRequirement{
			actions:   []Action{"validate-recovery"},
			evidence:  []Evidence{"common-app-hash", "height-progress", "ledger-invariants"},
			approvals: []RoleName{RoleProtocolOwner, RoleSecurityOwner},
		}
	} else if scenario.Outcome == OutcomeSafeStop {
		common[PhaseValidate] = phaseRequirement{
			actions:   []Action{"validate-recovery"},
			evidence:  []Evidence{"safe-stopped", "signer-isolation", "restart-disabled", "governance-decision"},
			approvals: []RoleName{RoleProtocolOwner, RoleSecurityOwner},
		}
	}

	for _, phaseName := range requiredPhaseOrder {
		requirement := common[phaseName]
		if scenario.Severity == SeverityCritical && phaseName == PhaseDecide {
			requirement = mergeRequirement(requirement, phaseRequirement{
				approvals: []RoleName{RoleSecurityOwner, RoleProtocolOwner},
			})
		}
		if specific, exists := spec.requirements[phaseName]; exists {
			requirement = mergeRequirement(requirement, specific)
		}
		v.requirePhaseValues(prefix, scenario.Phases, phaseName, requirement)
		v.rejectUnexpectedPhaseValues(prefix, scenario.Phases, phaseName, requirement)
	}
	ownerByPhase := map[PhaseName]RoleName{
		PhaseDetect: RoleIncidentCommander, PhaseContain: RoleSecurityOwner,
		PhasePreserve: RoleEvidenceCustodian, PhaseDecide: RoleIncidentCommander,
		PhaseRecover: spec.recoverOwner, PhaseValidate: RoleProtocolOwner,
		PhaseClose: RoleIncidentCommander,
	}
	for i, phase := range scenario.Phases {
		if expected := ownerByPhase[phase.Name]; expected != "" && phase.Owner != expected {
			v.fail(fmt.Sprintf("%s.phases[%d].owner", prefix, i), "%s phase owner must be %q", phase.Name, expected)
		}
	}
	if scenario.Kind == KindOperatorAuthorityCompromise {
		for i, phase := range scenario.Phases {
			if contains(phase.Actions, Action("rotate-consensus-key")) {
				v.fail(fmt.Sprintf("%s.phases[%d].actions", prefix, i), "operator-authority compromise must not authorize consensus-key rotation")
			}
		}
	}
	if scenario.Kind == KindCompatibleBinaryUpgrade {
		v.requireActionOrder(prefix, scenario.Phases, PhaseRecover,
			[]Action{"install-compatible-binary", "rollback-before-state-open"})
	}
	if scenario.Kind == KindLegacyAuthorityMigration {
		v.requireActionOrder(prefix, scenario.Phases, PhaseRecover,
			[]Action{"transform-fresh-genesis", "start-isolated-target", "stop-target-chain", "rollback-separate-chain", "restart-untouched-source"})
	}
}

func mergeRequirement(base, additional phaseRequirement) phaseRequirement {
	for _, action := range additional.actions {
		if !contains(base.actions, action) {
			base.actions = append(base.actions, action)
		}
	}
	for _, evidence := range additional.evidence {
		if !contains(base.evidence, evidence) {
			base.evidence = append(base.evidence, evidence)
		}
	}
	for _, approval := range additional.approvals {
		if !contains(base.approvals, approval) {
			base.approvals = append(base.approvals, approval)
		}
	}
	return base
}

func (v *validator) rejectUnexpectedPhaseValues(prefix string, phases []Phase, name PhaseName, allowed phaseRequirement) {
	for _, phase := range phases {
		if phase.Name != name {
			continue
		}
		for _, action := range phase.Actions {
			if !contains(allowed.actions, action) {
				v.fail(prefix+".phases."+string(name)+".actions", "%s phase forbids unexpected action", name)
			}
		}
		for _, evidence := range phase.Evidence {
			if !contains(allowed.evidence, evidence) {
				v.fail(prefix+".phases."+string(name)+".evidence", "%s phase forbids unexpected evidence", name)
			}
		}
		for _, approval := range phase.Approvals {
			if !contains(allowed.approvals, approval) {
				v.fail(prefix+".phases."+string(name)+".approvals", "%s phase forbids an unexpected approval", name)
			}
		}
		return
	}
}

func (v *validator) requireActionOrder(prefix string, phases []Phase, name PhaseName, required []Action) {
	for _, phase := range phases {
		if phase.Name != name {
			continue
		}
		position := -1
		for _, action := range required {
			found := -1
			for i := position + 1; i < len(phase.Actions); i++ {
				if phase.Actions[i] == action {
					found = i
					break
				}
			}
			if found < 0 {
				v.fail(prefix+".phases."+string(name)+".actions",
					"actions must preserve safe order; %q must follow the preceding recovery gate", action)
				return
			}
			position = found
		}
		return
	}
}

func (v *validator) requirePhaseValues(prefix string, phases []Phase, name PhaseName, requirement phaseRequirement) {
	var phase *Phase
	for i := range phases {
		if phases[i].Name == name {
			phase = &phases[i]
			break
		}
	}
	if phase == nil {
		return
	}
	for _, action := range requirement.actions {
		if !contains(phase.Actions, action) {
			v.fail(prefix+".phases."+string(name)+".actions", "%s phase requires action %q", name, action)
		}
	}
	for _, evidence := range requirement.evidence {
		if !contains(phase.Evidence, evidence) {
			v.fail(prefix+".phases."+string(name)+".evidence", "%s phase requires evidence %q", name, evidence)
		}
	}
	for _, approval := range requirement.approvals {
		if !contains(phase.Approvals, approval) {
			v.fail(prefix+".phases."+string(name)+".approvals", "%s phase requires approval from %q", name, approval)
		}
	}
}

func (v *validator) checkAbortConditions(prefix string, conditions []AbortCondition, required []AbortCondition) {
	seen := make(map[AbortCondition]bool)
	for i, condition := range conditions {
		check := fmt.Sprintf("%s.abort_conditions[%d]", prefix, i)
		if !supportedAbortConditions[condition] {
			v.fail(check, "abort condition is not supported")
		}
		if seen[condition] {
			v.fail(check, "abort condition is duplicated")
		}
		seen[condition] = true
	}
	for _, condition := range append([]AbortCondition{"evidence-incomplete", "invariant-failure"}, required...) {
		if !seen[condition] {
			v.fail(prefix+".abort_conditions", "required abort condition %q is missing", condition)
		}
	}
}

func contains[T comparable](values []T, expected T) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func setOf[T ~string](values ...T) map[T]bool {
	result := make(map[T]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}
