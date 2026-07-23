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
