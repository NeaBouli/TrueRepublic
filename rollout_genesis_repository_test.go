package main

import (
	"os"
	"strings"
	"testing"
)

func TestRolloutGenesisRepositoryContract(t *testing.T) {
	required := map[string][]string{
		"Makefile": {
			"rollout-genesis-contract-test:",
			"go test ./genesisevidence ./cmd/genesis-evidence -count=1",
		},
		".github/workflows/docs-check.yml": {
			"Check Rollout Genesis Contract",
			"make rollout-genesis-contract-test",
		},
		".github/workflows/reproducible-daemon.yml": {
			"genesisevidence/**",
			"cmd/genesis-evidence/**",
			"Verify rollout genesis qualification contract",
		},
		"docs/node-operators/configuration/rollout-genesis-qualification.md": {
			"truerepublic.rollout-genesis-manifest/v1",
			"truerepublic.rollout-genesis-evidence/v1",
			"completes no rollout checkbox",
			"production_ready` remains false",
		},
	}
	for path, values := range required {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		body := string(data)
		for _, value := range values {
			if !strings.Contains(body, value) {
				t.Errorf("%s is missing %q", path, value)
			}
		}
	}

	for _, path := range []string{
		"configs/release/rollout-genesis-manifest.json",
		"configs/release/genesis.json",
	} {
		if _, err := os.Lstat(path); err == nil {
			t.Errorf("repository must not ship an unapproved rollout artifact: %s", path)
		} else if !os.IsNotExist(err) {
			t.Fatalf("inspect %s: %v", path, err)
		}
	}
}
