package main

import (
	"os"
	"strings"
	"testing"

	"truerepublic/incidentpolicy"
)

func TestIncidentRehearsalRepositoryContract(t *testing.T) {
	t.Parallel()

	fixturePath := "configs/incidents/rehearsal.example.json"
	fixture, err := os.Open(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	contract, parseErr := incidentpolicy.Parse(fixture)
	closeErr := fixture.Close()
	if parseErr != nil {
		t.Fatalf("parse maintained incident rehearsal: %v", parseErr)
	}
	if closeErr != nil {
		t.Fatalf("close maintained incident rehearsal: %v", closeErr)
	}
	report := incidentpolicy.Validate(contract)
	if !report.Valid || report.ScenarioCount != 8 || len(report.Violations) != 0 {
		t.Fatalf("maintained incident rehearsal is invalid: %+v", report)
	}

	raw := readRepositoryFile(t, fixturePath)
	for _, forbidden := range []string{
		"mnemonic", "private_key", "priv_validator_state", "password", "token",
		"http://", "https://", "@", "BEGIN ", "${", "/home/", "/Users/",
	} {
		if strings.Contains(raw, forbidden) {
			t.Fatalf("synthetic incident rehearsal contains forbidden secret/host/value shape %q", forbidden)
		}
	}

	workflow := readRepositoryFile(t, ".github/workflows/go-ci.yml")
	for _, required := range []string{
		"'configs/incidents/**'",
		"Validate maintained incident rehearsal contract",
		"go run . incident-rehearsal validate",
		"--file configs/incidents/rehearsal.example.json",
		".scenario_count == 8",
	} {
		if !strings.Contains(workflow, required) {
			t.Fatalf("Go CI must contain incident rehearsal gate %q", required)
		}
	}

	guide := readRepositoryFile(t, "docs/node-operators/operations/incident-command.md")
	for _, required := range []string{
		"exactly seven forward-only phases",
		"One consensus identity may have exactly one active signer",
		"compatible binary rollback is allowed only",
		"before opening or mutating state",
		"Source and target validators may never run together",
		"operator-authority compromise cannot be repaired",
		"do not configure a live on-call rota",
		"truerepublicd incident-rehearsal validate",
		"failed or aborted rehearsal remains valid evidence",
	} {
		if !strings.Contains(guide, required) {
			t.Fatalf("incident command guide must explain %q", required)
		}
	}

	rotation := readRepositoryFile(t, "docs/node-operators/operations/validator-key-rotation.md")
	if strings.Contains(rotation, "does not feed CometBFT ABCI++ misbehavior") {
		t.Fatal("validator key-rotation guide retains stale pre-GH-59 slashing claim")
	}
	for _, required := range []string{
		"does feed CometBFT ABCI++ misbehavior",
		"Validator Slashing and Recovery",
		"do not treat key rotation alone as",
	} {
		if !strings.Contains(rotation, required) {
			t.Fatalf("validator key-rotation guide must contain corrected slashing boundary %q", required)
		}
	}

	network := readRepositoryFile(t, "docs/node-operators/configuration/network-config.md")
	for _, required := range []string{
		"illustrative host-local deny defaults",
		"private validator interface",
		"public proxy",
		"rate/burst/body/timeout/concurrency",
	} {
		if !strings.Contains(network, required) {
			t.Fatalf("generic network guide must retain safety warning %q", required)
		}
	}

	for _, path := range []string{
		"docs/node-operators/operations/monitoring.md",
		"docs/node-operators/operations/validator-slashing.md",
		"docs/node-operators/operations/validator-identity-recovery.md",
		"docs/node-operators/operations/validator-key-rotation.md",
		"docs/node-operators/operations/backup-recovery.md",
		"docs/node-operators/operations/upgrades.md",
		"docs/node-operators/operations/legacy-authority-migration.md",
	} {
		if !strings.Contains(readRepositoryFile(t, path), "Incident Command and Rehearsal") {
			t.Fatalf("specialist runbook %s must cross-link the incident command", path)
		}
	}
}
