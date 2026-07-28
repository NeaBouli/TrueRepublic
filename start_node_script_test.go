package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestStartNodeScriptRequiresValidatedRole(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		extraEnv   []string
		wantOK     bool
		wantOutput string
		wantCalls  []string
		forbidCall string
	}{
		{
			name:       "missing role fails before binary",
			wantOutput: "NETWORK_ROLE is required",
		},
		{
			name: "environment topology mutation rejected",
			extraEnv: []string{
				"NETWORK_ROLE=validator",
				"SEEDS=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa@seed.example:26656",
			},
			wantOutput: "SEEDS/PERSISTENT_PEERS environment mutation is disabled",
		},
		{
			name:       "validation failure prevents start",
			extraEnv:   []string{"NETWORK_ROLE=validator", "VALIDATION_FAIL=1"},
			wantCalls:  []string{"network-policy validate --role validator --home "},
			forbidCall: "start --home ",
		},
		{
			name:     "validated role starts node",
			extraEnv: []string{"NETWORK_ROLE=validator"},
			wantOK:   true,
			wantCalls: []string{
				"network-policy validate --role validator --home ",
				"start --home ",
				"--log_level info",
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			callsPath := filepath.Join(tempDir, "calls.log")
			binaryPath := filepath.Join(tempDir, "fake-truerepublicd")
			binary := `#!/bin/sh
printf '%s\n' "$*" >> "$CALLS_PATH"
if [ "${VALIDATION_FAIL:-}" = "1" ] &&
   [ "${1:-}" = "network-policy" ] &&
   [ "${2:-}" = "validate" ]; then
  exit 23
fi
`
			if err := os.WriteFile(binaryPath, []byte(binary), 0o700); err != nil {
				t.Fatalf("write fake binary: %v", err)
			}

			cmd := exec.Command("bash", "scripts/start-node.sh")
			cmd.Env = append([]string{
				"PATH=" + os.Getenv("PATH"),
				"BINARY=" + binaryPath,
				"CALLS_PATH=" + callsPath,
				"CHAIN_HOME=" + filepath.Join(tempDir, "home"),
			}, test.extraEnv...)
			output, err := cmd.CombinedOutput()
			if test.wantOK && err != nil {
				t.Fatalf("start-node.sh failed: %v\n%s", err, output)
			}
			if !test.wantOK && err == nil {
				t.Fatalf("start-node.sh unexpectedly passed:\n%s", output)
			}
			if test.wantOutput != "" && !strings.Contains(string(output), test.wantOutput) {
				t.Fatalf("output %q does not contain %q", output, test.wantOutput)
			}

			calls, readErr := os.ReadFile(callsPath)
			if len(test.wantCalls) == 0 {
				if readErr == nil && len(calls) > 0 {
					t.Fatalf("binary was called before validation: %q", calls)
				}
				return
			}
			if readErr != nil {
				t.Fatalf("read fake binary calls: %v", readErr)
			}
			for _, want := range test.wantCalls {
				if !strings.Contains(string(calls), want) {
					t.Fatalf("calls %q do not contain %q", calls, want)
				}
			}
			if test.forbidCall != "" && strings.Contains(string(calls), test.forbidCall) {
				t.Fatalf("calls %q unexpectedly contain %q", calls, test.forbidCall)
			}
		})
	}
}
