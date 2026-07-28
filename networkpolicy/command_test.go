package networkpolicy

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// runCommand executes the network-policy command tree with the given args and
// captures its output, exactly as an operator script would.
func runCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := NewCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func TestCommandValidateJSONPass(t *testing.T) {
	configTOML, appTOML := goodValidatorConfig("")
	home, selfID := makeHome(t, configTOML, appTOML)
	configTOML, appTOML = goodValidatorConfig(selfID)
	rewriteConfig(t, home, configTOML, appTOML)

	out, err := runCommand(t, "validate", "--role", "validator", "--home", home, "--output", "json")
	if err != nil {
		t.Fatalf("validate command failed on compliant home: %v\n%s", err, out)
	}
	var report Report
	if err := json.Unmarshal([]byte(out), &report); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, out)
	}
	if !report.Valid || report.Role != RoleValidator || len(report.Violations) != 0 {
		t.Fatalf("unexpected report: %+v", report)
	}
	if strings.Contains(out, home) || strings.Contains(out, `"home"`) {
		t.Fatalf("JSON output leaks the local node-home path: %s", out)
	}
}

func TestCommandValidateTextFailure(t *testing.T) {
	// Validator with pex on and a wildcard RPC listener must fail closed.
	config, app := goodValidatorConfig("")
	config = strings.Replace(config, "pex = false", "pex = true", 1)
	config = strings.Replace(config, goodRPC, "tcp://0.0.0.0:26657", 1)
	home, _ := makeHome(t, config, app)

	out, err := runCommand(t, "validate", "--role", "validator", "--home", home)
	if err == nil {
		t.Fatalf("expected non-zero exit on policy violation, output:\n%s", out)
	}
	if !strings.Contains(out, "VIOLATION p2p.pex") || !strings.Contains(out, "VIOLATION rpc.laddr") {
		t.Fatalf("output misses expected violations:\n%s", out)
	}
}

func TestCommandValidateJSONFailureIsStructured(t *testing.T) {
	config, app := goodSentryConfig()
	config = strings.Replace(config, "pex = true", "pex = false", 1)
	home, _ := makeHome(t, config, app)

	out, err := runCommand(t, "validate", "--role", "sentry", "--home", home, "--output", "json")
	if err == nil {
		t.Fatal("expected non-zero exit on policy violation")
	}
	var report Report
	if jsonErr := json.Unmarshal([]byte(out), &report); jsonErr != nil {
		t.Fatalf("output is not JSON: %v\n%s", jsonErr, out)
	}
	if report.Valid || len(report.Violations) == 0 {
		t.Fatalf("unexpected report: %+v", report)
	}
	if report.Violations[0].Check == "" || report.Violations[0].Message == "" {
		t.Fatalf("violation is not actionable: %+v", report.Violations[0])
	}
}

func TestCommandValidateRejectsBadInput(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"unknown-role", []string{"validate", "--role", "operator", "--home", t.TempDir()}},
		{"unknown-output", []string{"validate", "--role", "seed", "--home", t.TempDir(), "--output", "yaml"}},
		{"missing-role", []string{"validate", "--home", t.TempDir()}},
		{"missing-home", []string{"validate", "--role", "seed"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := runCommand(t, tc.args...); err == nil {
				t.Fatal("expected command to reject the input")
			}
		})
	}
}

func TestCommandDoesNotExposeLocalPeerBypass(t *testing.T) {
	if _, err := runCommand(t, "validate", "--role", "private", "--home", t.TempDir(), "--allow-local-peers"); err == nil {
		t.Fatal("operator CLI exposed the repository-test-only local peer bypass")
	}
}
