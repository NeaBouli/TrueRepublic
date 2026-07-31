package topologypolicy

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()
	command := NewCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetErr(output)
	command.SetArgs(args)
	err := command.Execute()
	return output.String(), err
}

func TestCommandValidateJSON(t *testing.T) {
	output, err := runCommand(t,
		"validate",
		"--file", exampleContractPath(),
		"--output", "json",
	)
	if err != nil {
		t.Fatalf("validate example: %v\n%s", err, output)
	}
	var report Report
	if err := json.Unmarshal([]byte(output), &report); err != nil {
		t.Fatalf("decode report: %v\n%s", err, output)
	}
	if !report.Valid || report.NodeCount != 5 || report.Violations == nil {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestCommandFailureDoesNotPrintContractPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "private-inventory.json")
	if err := os.WriteFile(path, []byte(`{"mnemonic":"must-not-be-accepted"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	output, err := runCommand(t, "validate", "--file", path)
	if err == nil {
		t.Fatal("invalid contract unexpectedly passed")
	}
	if strings.Contains(output, path) || strings.Contains(err.Error(), path) ||
		strings.Contains(output, "must-not-be-accepted") ||
		strings.Contains(err.Error(), "must-not-be-accepted") {
		t.Fatalf("failure leaked a local path or rejected value: output=%q err=%v", output, err)
	}
}

func TestCommandRejectsUnknownOutput(t *testing.T) {
	if _, err := runCommand(t,
		"validate",
		"--file", exampleContractPath(),
		"--output", "yaml",
	); err == nil {
		t.Fatal("unknown output format unexpectedly passed")
	}
}
