package incidentpolicy

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	rehearsalFixturePath = "../configs/incidents/rehearsal.example.json"
)

func loadFixtureContract(t *testing.T) Contract {
	t.Helper()
	data, err := os.ReadFile(filepath.Clean(rehearsalFixturePath))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	contract, err := Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	return contract
}

func cloneContract(contract Contract) Contract {
	clone := Contract{
		Version:     contract.Version,
		ExerciseID:  contract.ExerciseID,
		ChainID:     contract.ChainID,
		Environment: contract.Environment,
		Defaults:    contract.Defaults,
	}

	clone.Roles = make([]Role, len(contract.Roles))
	for i, role := range contract.Roles {
		clone.Roles[i] = Role{Name: role.Name, Primary: role.Primary, Secondary: role.Secondary}
	}

	clone.Scenarios = make([]Scenario, len(contract.Scenarios))
	for i, scenario := range contract.Scenarios {
		copyScenario := Scenario{
			ID: scenario.ID, Kind: scenario.Kind, Severity: scenario.Severity,
			Authority: scenario.Authority, Runbook: scenario.Runbook, Outcome: scenario.Outcome,
		}
		copyScenario.AbortConditions = append([]AbortCondition(nil), scenario.AbortConditions...)
		copyScenario.Phases = make([]Phase, len(scenario.Phases))
		for j, phase := range scenario.Phases {
			copyPhase := Phase{Name: phase.Name, Owner: phase.Owner}
			copyPhase.Actions = append([]Action(nil), phase.Actions...)
			copyPhase.Evidence = append([]Evidence(nil), phase.Evidence...)
			copyPhase.Approvals = append([]RoleName(nil), phase.Approvals...)
			copyScenario.Phases[j] = copyPhase
		}
		clone.Scenarios[i] = copyScenario
	}

	return clone
}

func hasViolation(report Report, check string) bool {
	for _, violation := range report.Violations {
		if violation.Check == check {
			return true
		}
	}
	return false
}

func hasViolationMessageContains(report Report, fragment string) bool {
	for _, violation := range report.Violations {
		if strings.Contains(violation.Check, fragment) || strings.Contains(violation.Message, fragment) {
			return true
		}
	}
	return false
}

func TestLoadAndValidateRehearsalFixture(t *testing.T) {
	base := loadFixtureContract(t)
	if got := Validate(base); !got.Valid {
		t.Fatalf("fixture must be valid, got %d violations", len(got.Violations))
	}

	fixture, err := os.ReadFile(filepath.Clean(rehearsalFixturePath))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	first, err := Parse(bytes.NewReader(fixture))
	if err != nil {
		t.Fatalf("parse first: %v", err)
	}
	second, err := Parse(bytes.NewReader(fixture))
	if err != nil {
		t.Fatalf("parse second: %v", err)
	}
	if first.Version != second.Version || first.ExerciseID != second.ExerciseID || len(first.Scenarios) != len(second.Scenarios) {
		t.Fatalf("fixture parsing is not deterministic")
	}
}

func TestParseStrictness(t *testing.T) {
	base, err := os.ReadFile(filepath.Clean(rehearsalFixturePath))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	tests := []struct {
		name string
		json string
		want string
	}{
		{
			name: "empty",
			json: "  \n",
			want: "incident rehearsal is empty",
		},
		{
			name: "trailing",
			json: `{"version":"x"}{"version":"x"}`,
			want: "trailing value is forbidden",
		},
		{
			name: "duplicate",
			json: `{"a":1,"a":2}`,
			want: "duplicate object key",
		},
		{
			name: "unknown_field",
			json: `{"version":"x","exercise_id":"x","chain_id":"x","environment":"synthetic-rehearsal","defaults":{"secrets":"forbidden","artifacts":"references-only","live_actions":"forbidden","source_target_concurrency":"forbidden"},"roles":[],"scenarios":[],"extra":"value"}`,
			want: "invalid incident rehearsal schema",
		},
		{
			name: "depth_exceeded",
			json: `{"level_01":{"level_02":{"level_03":{"level_04":{"level_05":{"level_06":{"level_07":{"level_08":{"level_09":{"level_10":{"level_11":{"level_12":{"level_13":{"level_14":{"level_15":{"level_16":{"level_17":{"level_18":{"level_19":{"level_20":{"level_21":{"level_22":{"level_23":{"level_24":{"level_25":{"level_26":{"level_27":{"level_28":{"level_29":{"level_30":{"level_31":{"level_32":{"level_33":{"level_34":0}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}`,
			want: "maximum JSON depth",
		},
		{
			name: "too_large",
			json: func() string {
				return `{"x":"` + strings.Repeat("a", MaxContractBytes) + `"}`
			}(),
			want: "exceeds",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse(strings.NewReader(tc.json))
			if err == nil {
				t.Fatalf("expected error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q, got %v", tc.want, err)
			}
		})
	}

	// baseline must parse
	if _, err := Parse(bytes.NewReader(base)); err != nil {
		t.Fatalf("fixture should parse: %v", err)
	}
}

func TestValidateCases(t *testing.T) {
	base := loadFixtureContract(t)

	tests := []struct {
		name     string
		mutate   func(Contract) Contract
		required string
		hasFrag  string
	}{
		{
			name:     "chain_id_must_be_synthetic_rehearsal",
			required: "chain_id",
			mutate: func(c Contract) Contract {
				c.ChainID = "truerepublic-1"
				return c
			},
		},
		{
			name:     "role_alias_must_be_logical_seat",
			required: "roles[0].primary",
			mutate: func(c Contract) Contract {
				c.Roles[0].Primary = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
				return c
			},
		},
		{
			name:     "role_aliases_must_be_unique_across_roles",
			required: "roles[1].primary",
			mutate: func(c Contract) Contract {
				c.Roles[1].Primary = c.Roles[0].Primary
				return c
			},
		},
		{
			name:    "missing_required_role",
			hasFrag: "required role \"incident-commander\" is missing",
			mutate: func(c Contract) Contract {
				c.Roles = c.Roles[1:]
				return c
			},
		},
		{
			name:     "duplicate_role",
			required: "roles[0].name",
			hasFrag:  "role is duplicated",
			mutate: func(c Contract) Contract {
				c.Roles = append(c.Roles, c.Roles[0])
				return c
			},
		},
		{
			name:     "role_secondary_matches_primary",
			required: "roles[0].secondary",
			mutate: func(c Contract) Contract {
				c.Roles[0].Secondary = c.Roles[0].Primary
				return c
			},
		},
		{
			name:     "phase_count_not_exact",
			required: "scenarios[0].phases",
			mutate: func(c Contract) Contract {
				c.Scenarios[0].Phases = c.Scenarios[0].Phases[:6]
				return c
			},
		},
		{
			name:     "phase_order_wrong",
			required: "scenarios[0].phases[0].name",
			mutate: func(c Contract) Contract {
				c.Scenarios[0].Phases[0], c.Scenarios[0].Phases[1] = c.Scenarios[0].Phases[1], c.Scenarios[0].Phases[0]
				return c
			},
		},
		{
			name:     "missing_detect_action",
			required: "scenarios[0].phases.detect.actions",
			hasFrag:  "classify-incident",
			mutate: func(c Contract) Contract {
				c.Scenarios[0].Phases[0].Actions = nil
				return c
			},
		},
		{
			name:     "missing_validate_evidence",
			required: "scenarios[0].phases.validate.evidence",
			hasFrag:  "validator-power",
			mutate: func(c Contract) Contract {
				c.Scenarios[0].Phases[5].Evidence = c.Scenarios[0].Phases[5].Evidence[:2]
				return c
			},
		},
		{
			name:     "missing_decide_approvals",
			required: "scenarios[0].phases.decide.approvals",
			mutate: func(c Contract) Contract {
				c.Scenarios[0].Phases[3].Approvals = []RoleName{RoleIncidentCommander}
				return c
			},
		},
		{
			name:     "invalid_severity",
			required: "scenarios[1].severity",
			mutate: func(c Contract) Contract {
				c.Scenarios[1].Severity = "medium"
				return c
			},
		},
		{
			name:     "invalid_authority",
			required: "scenarios[1].authority",
			mutate: func(c Contract) Contract {
				c.Scenarios[1].Authority = "manual-ops"
				return c
			},
		},
		{
			name:     "invalid_outcome",
			required: "scenarios[1].outcome",
			mutate: func(c Contract) Contract {
				c.Scenarios[1].Outcome = "lost"
				return c
			},
		},
		{
			name:     "invalid_runbook",
			required: "scenarios[1].runbook",
			mutate: func(c Contract) Contract {
				c.Scenarios[1].Runbook = "docs/invalid.md"
				return c
			},
		},
		{
			name:     "duplicate_scenario_kind",
			required: "scenarios[1].kind",
			mutate: func(c Contract) Contract {
				c.Scenarios[1].Kind = c.Scenarios[0].Kind
				return c
			},
		},
		{
			name:     "missing_scenario_kind",
			required: "scenarios.validator-slashing",
			mutate: func(c Contract) Contract {
				c.Scenarios = append(c.Scenarios[:2], c.Scenarios[3:]...)
				return c
			},
		},
		{
			name:    "unsupported_action",
			hasFrag: "not supported",
			mutate: func(c Contract) Contract {
				c.Scenarios[0].Phases[1].Actions = append(c.Scenarios[0].Phases[1].Actions, Action("nope"))
				return c
			},
		},
		{
			name:    "supported_but_wrong_scenario_action",
			hasFrag: "forbids unexpected action",
			mutate: func(c Contract) Contract {
				c.Scenarios[5].Phases[4].Actions = append(c.Scenarios[5].Phases[4].Actions, Action("rotate-consensus-key"))
				return c
			},
		},
		{
			name:    "unsupported_evidence",
			hasFrag: "not supported",
			mutate: func(c Contract) Contract {
				c.Scenarios[0].Phases[1].Evidence = append(c.Scenarios[0].Phases[1].Evidence, Evidence("nope"))
				return c
			},
		},
		{
			name:    "unsupported_abort",
			hasFrag: "not supported",
			mutate: func(c Contract) Contract {
				c.Scenarios[0].AbortConditions = append(c.Scenarios[0].AbortConditions, AbortCondition("nope"))
				return c
			},
		},
		{
			name:     "critical_decide_approvals",
			required: "scenarios[0].phases.decide.approvals",
			hasFrag:  "protocol-owner",
			mutate: func(c Contract) Contract {
				c.Scenarios[0].Phases[3].Approvals = []RoleName{RoleIncidentCommander}
				return c
			},
		},
		{
			name:     "signer_isolation_boundary_missing",
			required: "scenarios[1].phases.contain.evidence",
			mutate: func(c Contract) Contract {
				c.Scenarios[1].Phases[1].Evidence = []Evidence{Evidence("restart-disabled")}
				return c
			},
		},
		{
			name:     "stale_signer_boundary_missing",
			required: "scenarios[1].abort_conditions",
			hasFrag:  "signer-state-stale",
			mutate: func(c Contract) Contract {
				filtered := make([]AbortCondition, 0, len(c.Scenarios[1].AbortConditions))
				for _, cond := range c.Scenarios[1].AbortConditions {
					if cond != "signer-state-stale" {
						filtered = append(filtered, cond)
					}
				}
				c.Scenarios[1].AbortConditions = filtered
				return c
			},
		},
		{
			name:     "second_signer_boundary_missing",
			required: "scenarios[1].abort_conditions",
			hasFrag:  "second-signer-detected",
			mutate: func(c Contract) Contract {
				filtered := make([]AbortCondition, 0, len(c.Scenarios[1].AbortConditions))
				for _, cond := range c.Scenarios[1].AbortConditions {
					if cond != "second-signer-detected" {
						filtered = append(filtered, cond)
					}
				}
				c.Scenarios[1].AbortConditions = filtered
				return c
			},
		},
		{
			name:     "slashing_second_signer_boundary_missing",
			required: "scenarios[2].abort_conditions",
			hasFrag:  "second-signer-detected",
			mutate: func(c Contract) Contract {
				c.Scenarios[2].AbortConditions = withoutAbortCondition(c.Scenarios[2].AbortConditions, "second-signer-detected")
				return c
			},
		},
		{
			name:     "key_compromise_second_signer_boundary_missing",
			required: "scenarios[3].abort_conditions",
			hasFrag:  "second-signer-detected",
			mutate: func(c Contract) Contract {
				c.Scenarios[3].AbortConditions = withoutAbortCondition(c.Scenarios[3].AbortConditions, "second-signer-detected")
				return c
			},
		},
		{
			name:     "operator_authority_cannot_rotate",
			required: "scenarios[4].phases",
			hasFrag:  "must not authorize consensus-key rotation",
			mutate: func(c Contract) Contract {
				c.Scenarios[4].Phases[5].Actions = append(c.Scenarios[4].Phases[5].Actions, Action("rotate-consensus-key"))
				return c
			},
		},
		{
			name:     "backup_secret_material_abort_missing",
			required: "scenarios[5].abort_conditions",
			hasFrag:  "secret-material-detected",
			mutate: func(c Contract) Contract {
				filtered := make([]AbortCondition, 0, len(c.Scenarios[5].AbortConditions))
				for _, cond := range c.Scenarios[5].AbortConditions {
					if cond != "secret-material-detected" {
						filtered = append(filtered, cond)
					}
				}
				c.Scenarios[5].AbortConditions = filtered
				return c
			},
		},
		{
			name:     "backup_fresh_target_gate_missing",
			required: "scenarios[5].abort_conditions",
			hasFrag:  "restore-target-not-fresh",
			mutate: func(c Contract) Contract {
				filtered := make([]AbortCondition, 0, len(c.Scenarios[5].AbortConditions))
				for _, cond := range c.Scenarios[5].AbortConditions {
					if cond != "restore-target-not-fresh" {
						filtered = append(filtered, cond)
					}
				}
				c.Scenarios[5].AbortConditions = filtered
				return c
			},
		},
		{
			name:     "upgrade_candidate_state_unopened_missing",
			required: "scenarios[6].phases.recover.evidence",
			hasFrag:  "candidate-state-unopened",
			mutate: func(c Contract) Contract {
				c.Scenarios[6].Phases[4].Evidence = nil
				return c
			},
		},
		{
			name:     "upgrade_candidate_opened_state_abort_missing",
			required: "scenarios[6].abort_conditions",
			hasFrag:  "candidate-opened-state",
			mutate: func(c Contract) Contract {
				filtered := make([]AbortCondition, 0, len(c.Scenarios[6].AbortConditions))
				for _, cond := range c.Scenarios[6].AbortConditions {
					if cond != "candidate-opened-state" {
						filtered = append(filtered, cond)
					}
				}
				c.Scenarios[6].AbortConditions = filtered
				return c
			},
		},
		{
			name:    "upgrade_recovery_order_unsafe",
			hasFrag: "safe order",
			mutate: func(c Contract) Contract {
				actions := c.Scenarios[6].Phases[4].Actions
				actions[0], actions[1] = actions[1], actions[0]
				return c
			},
		},
		{
			name:     "legacy_source_target_isolation_missing",
			required: "scenarios[7].phases.contain.evidence",
			hasFrag:  "source-target-isolation",
			mutate: func(c Contract) Contract {
				c.Scenarios[7].Phases[1].Evidence = []Evidence{Evidence("restart-disabled")}
				return c
			},
		},
		{
			name:     "legacy_distinct_chain_missing",
			required: "scenarios[7].phases.preserve.evidence",
			hasFrag:  "distinct-chain-ids",
			mutate: func(c Contract) Contract {
				e := c.Scenarios[7].Phases[2].Evidence
				filtered := make([]Evidence, 0, len(e))
				for _, evidence := range e {
					if evidence != "distinct-chain-ids" {
						filtered = append(filtered, evidence)
					}
				}
				c.Scenarios[7].Phases[2].Evidence = filtered
				return c
			},
		},
		{
			name:     "legacy_fresh_proof_missing",
			required: "scenarios[7].phases.validate.evidence",
			hasFrag:  "fresh-operator-proofs",
			mutate: func(c Contract) Contract {
				e := c.Scenarios[7].Phases[6-1].Evidence
				filtered := make([]Evidence, 0, len(e))
				for _, evidence := range e {
					if evidence != "fresh-operator-proofs" {
						filtered = append(filtered, evidence)
					}
				}
				c.Scenarios[7].Phases[6-1].Evidence = filtered
				return c
			},
		},
		{
			name:     "legacy_empty_wasm_missing",
			required: "scenarios[7].phases.validate.evidence",
			hasFrag:  "empty-wasm-state",
			mutate: func(c Contract) Contract {
				e := c.Scenarios[7].Phases[6-1].Evidence
				filtered := make([]Evidence, 0, len(e))
				for _, evidence := range e {
					if evidence != "empty-wasm-state" {
						filtered = append(filtered, evidence)
					}
				}
				c.Scenarios[7].Phases[6-1].Evidence = filtered
				return c
			},
		},
		{
			name:     "legacy_concurrency_abort_missing",
			required: "scenarios[7].abort_conditions",
			hasFrag:  "source-target-concurrency",
			mutate: func(c Contract) Contract {
				filtered := make([]AbortCondition, 0, len(c.Scenarios[7].AbortConditions))
				for _, cond := range c.Scenarios[7].AbortConditions {
					if cond != "source-target-concurrency" {
						filtered = append(filtered, cond)
					}
				}
				c.Scenarios[7].AbortConditions = filtered
				return c
			},
		},
		{
			name:    "legacy_recovery_order_unsafe",
			hasFrag: "safe order",
			mutate: func(c Contract) Contract {
				actions := c.Scenarios[7].Phases[4].Actions
				actions[2], actions[3] = actions[3], actions[2]
				return c
			},
		},
		{
			name:     "legacy_state_mixing_abort_missing",
			required: "scenarios[7].abort_conditions",
			hasFrag:  "source-target-state-mixing",
			mutate: func(c Contract) Contract {
				filtered := make([]AbortCondition, 0, len(c.Scenarios[7].AbortConditions))
				for _, condition := range c.Scenarios[7].AbortConditions {
					if condition != "source-target-state-mixing" {
						filtered = append(filtered, condition)
					}
				}
				c.Scenarios[7].AbortConditions = filtered
				return c
			},
		},
		{
			name:     "legacy_target_stop_evidence_missing",
			required: "scenarios[7].phases.recover.evidence",
			hasFrag:  "target-stopped",
			mutate: func(c Contract) Contract {
				c.Scenarios[7].Phases[4].Evidence = nil
				return c
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			candidate := tc.mutate(cloneContract(base))
			report := Validate(candidate)
			if report.Valid {
				t.Fatalf("expected invalid scenario %s", tc.name)
			}
			if tc.required != "" && hasViolation(report, tc.required) {
				return
			}
			if tc.hasFrag != "" && hasViolationMessageContains(report, tc.hasFrag) {
				return
			}
			if tc.required == "" && tc.hasFrag == "" {
				return
			}
			for _, violation := range report.Violations {
				t.Logf("%s: %s: %s", tc.name, violation.Check, violation.Message)
			}
			t.Fatalf("expected violation with check=%q or fragment=%q", tc.required, tc.hasFrag)
		})
	}
}

func withoutAbortCondition(conditions []AbortCondition, rejected AbortCondition) []AbortCondition {
	filtered := make([]AbortCondition, 0, len(conditions))
	for _, condition := range conditions {
		if condition != rejected {
			filtered = append(filtered, condition)
		}
	}
	return filtered
}

func TestIncidentPolicyValidateCommandJSON(t *testing.T) {
	base := loadFixtureContract(t)

	contractPath := filepath.Join(t.TempDir(), "contract.json")
	payload, err := json.Marshal(base)
	if err != nil {
		t.Fatalf("marshal base: %v", err)
	}
	if err := os.WriteFile(contractPath, payload, 0o600); err != nil {
		t.Fatalf("write temp: %v", err)
	}

	cmd := NewCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"validate", "--file", contractPath, "--output", "json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("command failed: %v", err)
	}
	var report Report
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("invalid json output: %v", err)
	}
	if !report.Valid || report.ScenarioCount != len(base.Scenarios) {
		t.Fatalf("expected valid command report, got valid=%v scenarios=%d", report.Valid, report.ScenarioCount)
	}
	result := out.String()
	if strings.Contains(result, "incident-secret-primary") {
		t.Fatalf("command output leaked local values")
	}
	if strings.Contains(result, contractPath) {
		t.Fatalf("command output leaked file path")
	}
}

func TestIncidentPolicyRejectsSecretLikeSeatWithoutLeak(t *testing.T) {
	base := loadFixtureContract(t)
	const rejected = "incident-secret-primary"
	base.Roles[0].Primary = rejected
	report := Validate(base)
	if report.Valid || !hasViolation(report, "roles[0].primary") {
		t.Fatalf("secret-like role seat unexpectedly passed: %+v", report)
	}

	path := filepath.Join(t.TempDir(), "private-seats.json")
	payload, err := json.Marshal(base)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := NewCommand()
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"validate", "--file", path, "--output", "json"})
	err = cmd.Execute()
	if err == nil {
		t.Fatal("secret-like role seat unexpectedly passed through CLI")
	}
	combined := out.String() + err.Error()
	if strings.Contains(combined, rejected) || strings.Contains(combined, path) {
		t.Fatalf("validation leaked rejected seat or local path: %q", combined)
	}
}

func TestIncidentPolicyCommandFailureDoesNotLeakRejectedInput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "private-incident.json")
	const secret = "must-not-appear-in-error"
	if err := os.WriteFile(path, []byte(`{"mnemonic":"`+secret+`"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := NewCommand()
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"validate", "--file", path, "--output", "json"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("invalid rehearsal unexpectedly passed")
	}
	combined := out.String() + err.Error()
	if strings.Contains(combined, path) || strings.Contains(combined, secret) {
		t.Fatalf("command failure leaked a local path or rejected value: %q", combined)
	}
}

func TestRejectedContractValuesNeverReachReportsOrErrors(t *testing.T) {
	const rejected = "must-not-appear-in-public-output"
	mutations := []func(Contract) Contract{
		func(c Contract) Contract { c.Version = rejected; return c },
		func(c Contract) Contract { c.ExerciseID = rejected; return c },
		func(c Contract) Contract { c.ChainID = rejected; return c },
		func(c Contract) Contract { c.Roles[0].Name = RoleName(rejected); return c },
		func(c Contract) Contract { c.Roles[0].Primary = rejected; return c },
		func(c Contract) Contract { c.Scenarios[0].ID = rejected; return c },
		func(c Contract) Contract { c.Scenarios[0].Kind = ScenarioKind(rejected); return c },
		func(c Contract) Contract { c.Scenarios[0].Severity = Severity(rejected); return c },
		func(c Contract) Contract { c.Scenarios[0].Authority = Authority(rejected); return c },
		func(c Contract) Contract { c.Scenarios[0].Runbook = rejected; return c },
		func(c Contract) Contract { c.Scenarios[0].Outcome = Outcome(rejected); return c },
		func(c Contract) Contract { c.Scenarios[0].Phases[0].Name = PhaseName(rejected); return c },
		func(c Contract) Contract { c.Scenarios[0].Phases[0].Owner = RoleName(rejected); return c },
		func(c Contract) Contract { c.Scenarios[0].Phases[0].Actions[0] = Action(rejected); return c },
		func(c Contract) Contract { c.Scenarios[0].Phases[0].Evidence[0] = Evidence(rejected); return c },
		func(c Contract) Contract { c.Scenarios[0].Phases[3].Approvals[0] = RoleName(rejected); return c },
		func(c Contract) Contract { c.Scenarios[0].AbortConditions[0] = AbortCondition(rejected); return c },
	}
	for i, mutate := range mutations {
		report := Validate(mutate(loadFixtureContract(t)))
		if report.Valid {
			t.Fatalf("mutation %d unexpectedly passed", i)
		}
		encoded, err := json.Marshal(report)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(encoded), rejected) {
			t.Fatalf("mutation %d leaked rejected value: %s", i, encoded)
		}
	}

	short := loadFixtureContract(t)
	short.Scenarios = short.Scenarios[:1]
	if report := Validate(short); report.ScenarioCount != MaintainedScenarioCount {
		t.Fatalf("invalid contract controlled public scenario count: %d", report.ScenarioCount)
	}

	for _, input := range []string{
		`{"` + rejected + `":true}`,
		`{"` + rejected + `":1,"` + rejected + `":2}`,
	} {
		_, err := Parse(strings.NewReader(input))
		if err == nil {
			t.Fatal("rejected JSON unexpectedly parsed")
		}
		if strings.Contains(err.Error(), rejected) {
			t.Fatalf("parser leaked rejected key: %q", err)
		}
	}

	cmd := NewCommand()
	cmd.SetArgs([]string{"validate", "--file", rehearsalFixturePath, "--output", rejected})
	if err := cmd.Execute(); err == nil || strings.Contains(err.Error(), rejected) {
		t.Fatalf("output-format rejection leaked value or passed: %v", err)
	}

	for _, args := range [][]string{
		{"validate", "--file", rehearsalFixturePath, rejected},
		{"validate", "--file", rehearsalFixturePath, "--" + rejected},
	} {
		cmd = NewCommand()
		cmd.SetArgs(args)
		err := cmd.Execute()
		if err == nil || strings.Contains(err.Error(), rejected) {
			t.Fatalf("CLI rejection leaked value or passed: %v", err)
		}
	}
}
