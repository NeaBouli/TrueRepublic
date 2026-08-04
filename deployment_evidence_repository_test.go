package main

import (
	"os"
	"strings"
	"testing"
	"time"

	"truerepublic/deploymentevidence"
)

func TestDeploymentEvidenceRepositoryContract(t *testing.T) {
	topology, err := deploymentevidence.LoadTopology("configs/topology/qualification.example.json")
	if err != nil {
		t.Fatalf("load maintained topology: %v", err)
	}
	manifest, err := deploymentevidence.LoadManifest("configs/deployment/evidence.example.json")
	if err != nil {
		t.Fatalf("load maintained deployment evidence: %v", err)
	}
	evaluation, err := time.Parse(time.RFC3339, "2026-08-01T12:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	report := deploymentevidence.Verify(manifest, topology, evaluation)
	if !report.Valid || report.GateCount != 11 || report.NodeCount != 5 || len(report.Violations) != 0 {
		t.Fatalf("maintained deployment evidence must remain valid: %+v", report)
	}
	if manifest.TopologySHA256 != topology.SHA256 {
		t.Fatal("maintained manifest is not bound to the raw topology contract")
	}

	assertFileOmitsDeploymentInventory(t, "configs/deployment/evidence.example.json")
	assertFileContainsAll(t, ".github/workflows/go-ci.yml", []string{
		"configs/deployment/**",
		"Verify maintained deployment evidence manifest",
		"deployment-evidence verify",
		"--contract configs/topology/qualification.example.json",
		"--manifest configs/deployment/evidence.example.json",
		"--at 2026-08-01T12:00:00Z",
		".gate_count == 11",
	})
	assertFileContainsAll(t, "docs/node-operators/operations/deployment-evidence.md", []string{
		"offline",
		"two-person",
		"private",
		"deploy, probe, approve, or prove",
		"checkbox remains open",
	})
}

func assertFileOmitsDeploymentInventory(t *testing.T, path string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	lower := strings.ToLower(string(raw))
	for _, forbidden := range []string{
		"mnemonic", "private_key", "priv_validator_state", "password", "token",
		"http://", "https://", "@", "-----begin", "${", "/users/", "/home/",
		`"host"`, `"node_id"`, `"provider_id"`,
	} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("%s contains forbidden inventory or secret marker %q", path, forbidden)
		}
	}
}

func assertFileContainsAll(t *testing.T, path string, required []string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	contents := string(raw)
	for _, value := range required {
		if !strings.Contains(contents, value) {
			t.Fatalf("%s must contain %q", path, value)
		}
	}
}
