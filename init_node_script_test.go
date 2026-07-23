package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitNodeScriptUsesOnlyPoDBootstrap(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	home := filepath.Join(tempDir, "node-home")
	invocations := filepath.Join(tempDir, "invocations")
	fakeBinary := filepath.Join(tempDir, "truerepublicd")
	fake := `#!/bin/sh
set -eu
printf '%s\n' "$*" >> "$INVOCATIONS"
mkdir -p "$FAKE_HOME/config"
printf 'minimum-gas-prices = ""\n' > "$FAKE_HOME/config/app.toml"
printf 'prometheus = false\n' > "$FAKE_HOME/config/config.toml"
`
	if err := os.WriteFile(fakeBinary, []byte(fake), 0o700); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("bash", "scripts/init-node.sh")
	cmd.Env = append(os.Environ(),
		"BINARY="+fakeBinary,
		"CHAIN_ID=truerepublic-wrapper-test",
		"MONIKER=wrapper-node",
		"CHAIN_HOME="+home,
		"BOOTSTRAP_OPERATOR=truerepublic13hgqwy9986x5nk6jt23ns5v7j0acs8qmhchhtw",
		"INVOCATIONS="+invocations,
		"FAKE_HOME="+home,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("init wrapper failed: %v\n%s", err, output)
	}

	called, err := os.ReadFile(invocations)
	if err != nil {
		t.Fatal(err)
	}
	want := "init wrapper-node --chain-id truerepublic-wrapper-test --home " + home + " --bootstrap-operator truerepublic13hgqwy9986x5nk6jt23ns5v7j0acs8qmhchhtw\n"
	if string(called) != want {
		t.Fatalf("unexpected daemon commands:\n got: %q\nwant: %q", called, want)
	}
	for _, forbidden := range []string{"gentx", "collect-gentxs", "add-genesis-account", "keys add"} {
		if strings.Contains(string(called), forbidden) {
			t.Fatalf("wrapper invoked unsupported command %q: %s", forbidden, called)
		}
	}
	if _, err := os.Stat(filepath.Join(home, "genesis-key.txt")); !os.IsNotExist(err) {
		t.Fatalf("wrapper must not persist a mnemonic file: %v", err)
	}

	appConfig, err := os.ReadFile(filepath.Join(home, "config", "app.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(appConfig), `minimum-gas-prices = "1000upnyx"`) {
		t.Fatalf("minimum gas price was not configured: %s", appConfig)
	}
	cometConfig, err := os.ReadFile(filepath.Join(home, "config", "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(cometConfig), "prometheus = true") {
		t.Fatalf("Prometheus was not enabled: %s", cometConfig)
	}
	if !strings.Contains(string(output), "bank-backed PoD genesis") {
		t.Fatalf("wrapper did not report the supported bootstrap path: %s", output)
	}
}

func TestDockerEntrypointRequiresBootstrapOperatorOnlyForFreshHome(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	if err := os.MkdirAll(binDir, 0o700); err != nil {
		t.Fatal(err)
	}
	invocations := filepath.Join(tempDir, "invocations")
	fakeBinary := filepath.Join(binDir, "truerepublicd")
	fake := `#!/bin/sh
set -eu
printf '%s\n' "$*" >> "$INVOCATIONS"
if [ "${1:-}" = "init" ]; then
  mkdir -p "$HOME/.truerepublic/config"
  printf '{}\n' > "$HOME/.truerepublic/config/genesis.json"
fi
`
	if err := os.WriteFile(fakeBinary, []byte(fake), 0o700); err != nil {
		t.Fatal(err)
	}

	run := func(operator string) ([]byte, error) {
		cmd := exec.Command("sh", "scripts/docker-entrypoint.sh", "start")
		cmd.Env = append(os.Environ(),
			"HOME="+tempDir,
			"PATH="+binDir+":"+os.Getenv("PATH"),
			"INVOCATIONS="+invocations,
			"BOOTSTRAP_OPERATOR="+operator,
		)
		return cmd.CombinedOutput()
	}
	if output, err := run(""); err == nil || !strings.Contains(string(output), "BOOTSTRAP_OPERATOR is required") {
		t.Fatalf("fresh-home entrypoint did not fail closed: err=%v output=%s", err, output)
	}
	if _, err := os.Stat(invocations); !os.IsNotExist(err) {
		t.Fatal("entrypoint invoked daemon before validating bootstrap operator")
	}

	operator := "truerepublic13hgqwy9986x5nk6jt23ns5v7j0acs8qmhchhtw"
	if output, err := run(operator); err != nil {
		t.Fatalf("entrypoint with operator failed: %v\n%s", err, output)
	}
	called, err := os.ReadFile(invocations)
	if err != nil {
		t.Fatal(err)
	}
	want := "init truerepublic-node --chain-id truerepublic-1 --bootstrap-operator " + operator + " --home " + filepath.Join(tempDir, ".truerepublic") + "\nstart --home " + filepath.Join(tempDir, ".truerepublic") + "\n"
	if string(called) != want {
		t.Fatalf("unexpected entrypoint commands:\n got: %q\nwant: %q", called, want)
	}

	if err := os.WriteFile(invocations, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if output, err := run(""); err != nil {
		t.Fatalf("initialized home unexpectedly required operator: %v\n%s", err, output)
	}
	called, err = os.ReadFile(invocations)
	if err != nil {
		t.Fatal(err)
	}
	want = "start --home " + filepath.Join(tempDir, ".truerepublic") + "\n"
	if string(called) != want {
		t.Fatalf("initialized-home entrypoint commands:\n got: %q\nwant: %q", called, want)
	}
}
