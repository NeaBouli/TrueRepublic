package main

import (
	"os"
	"strings"
	"testing"
)

func TestGenerativeQualityRepositoryContract(t *testing.T) {
	scriptPath := "scripts/check-generative-quality.sh"
	script := readRepositoryFile(t, scriptPath)
	workflow := readRepositoryFile(t, ".github/workflows/go-ci.yml")
	makefile := readRepositoryFile(t, "Makefile")

	info, err := os.Stat(scriptPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatal("generative quality gate must be executable")
	}

	for _, required := range []string{
		"go test -race ./x/dex ./token ./x/truedemocracy",
		"TestValidateGenesisSupplyRejectsMalformedStructures",
		"TestConsensusSlashingReplayProperty",
		"FuzzComputeSwapOutput",
		"FuzzPNYXCapAndGenesisValidation",
		"TRUEREPUBLIC_FUZZ_ITERATIONS",
		"FUZZ_ITERATIONS > 60000",
		`-fuzztime="${FUZZ_ITERATIONS}x"`,
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("generative quality script missing %q", required)
		}
	}
	if strings.Contains(script, `-fuzztime="${FUZZ_ITERATIONS}s"`) {
		t.Fatal("generative quality gate must not use a wall-clock fuzz deadline")
	}

	for _, required := range []string{
		"quality-depth:",
		"timeout-minutes: 15",
		"./scripts/check-generative-quality.sh",
	} {
		if !strings.Contains(workflow, required) {
			t.Fatalf("Go CI missing generative quality contract %q", required)
		}
	}

	if !strings.Contains(makefile, "quality-depth:\n\t./scripts/check-generative-quality.sh") {
		t.Fatal("Makefile must expose the repository generative quality gate")
	}
}
