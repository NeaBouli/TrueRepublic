package main

import (
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"truerepublic/capacitypolicy"
)

func TestCapacityQualificationRepositoryContract(t *testing.T) {
	t.Parallel()

	fixturePath := "configs/capacity/qualification.example.json"
	fixture, err := os.Open(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	contract, parseErr := capacitypolicy.ParseContract(fixture)
	closeErr := fixture.Close()
	if parseErr != nil {
		t.Fatalf("parse maintained capacity qualification: %v", parseErr)
	}
	if closeErr != nil {
		t.Fatalf("close maintained capacity qualification: %v", closeErr)
	}
	report := capacitypolicy.ValidateContract(contract)
	if !report.Valid || report.ValidatorCount != 4 || report.TransactionCount != 96 || len(report.Violations) != 0 {
		t.Fatalf("maintained capacity qualification is invalid: %+v", report)
	}

	raw := readRepositoryFile(t, fixturePath)
	for _, forbidden := range []string{
		"mnemonic", "private_key", "priv_validator_state", "password",
		"http://", "https://", "@", "BEGIN ", "${", "/home/", "/Users/",
	} {
		if strings.Contains(raw, forbidden) {
			t.Fatalf("synthetic capacity contract contains forbidden secret/host/value shape %q", forbidden)
		}
	}

	workflow := readRepositoryFile(t, ".github/workflows/go-ci.yml")
	for _, required := range []string{
		"'configs/capacity/**'",
		"Validate maintained capacity qualification contract",
		"go run . capacity-policy validate",
		"--file configs/capacity/qualification.example.json",
		"TestMultiValidatorCapacitySustainedLoad",
		"go run . capacity-policy verify",
	} {
		if !strings.Contains(workflow, required) {
			t.Fatalf("Go CI must contain capacity qualification gate %q", required)
		}
	}

	var compose struct {
		Services map[string]struct {
			Logging struct {
				Driver  string            `yaml:"driver"`
				Options map[string]string `yaml:"options"`
			} `yaml:"logging"`
		} `yaml:"services"`
	}
	if err := yaml.Unmarshal([]byte(readRepositoryFile(t, "docker-compose.yml")), &compose); err != nil {
		t.Fatalf("parse Compose configuration: %v", err)
	}
	node, exists := compose.Services["truerepublic-node"]
	if !exists || node.Logging.Driver != contract.Logging.Driver ||
		node.Logging.Options["max-size"] != "50m" ||
		node.Logging.Options["max-file"] != "3" {
		t.Fatalf("truerepublic-node must retain the bounded static logging configuration: %+v", node.Logging)
	}

	guide := readRepositoryFile(t, "docs/node-operators/operations/capacity-qualification.md")
	for _, required := range []string{
		"synthetic loopback regression evidence",
		"exactly four temporary validators",
		"96 deterministic transactions",
		"default pruning",
		"not establish production sizing",
		"TRUEREPUBLIC_CAPACITY_EVIDENCE_OUT",
		"capacity-policy verify",
		"requires an operator to abort",
	} {
		if !strings.Contains(guide, required) {
			t.Fatalf("capacity qualification guide must explain %q", required)
		}
	}
}
