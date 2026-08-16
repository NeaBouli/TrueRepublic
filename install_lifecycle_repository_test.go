package main

import (
	"os"
	"strings"
	"testing"
)

func TestInstallLifecycleRepositoryContract(t *testing.T) {
	t.Parallel()

	required := map[string][]string{
		"Makefile": {
			"install-lifecycle-contract-test:",
			"go test ./installlifecycle ./cmd/install-lifecycle",
		},
		".github/workflows/reproducible-daemon.yml": {
			"'installlifecycle/**'",
			"'cmd/install-lifecycle/**'",
			"'docs/node-operators/installation/**'",
			"make install-lifecycle-contract-test",
			"go test . -run '^TestInstallLifecycleRepositoryContract$' -count=1",
		},
		"docs/node-operators/installation/lifecycle.md": {
			"Operator state boundary",
			"pre-start",
			"rollback",
			"uninstall",
			"production",
		},
	}
	for path, needles := range required {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for _, needle := range needles {
			if !strings.Contains(string(body), needle) {
				t.Errorf("%s must contain %q", path, needle)
			}
		}
	}

	for _, path := range []string{
		"Makefile",
		"INSTALLATION.md",
		"docs/DEPLOYMENT.md",
		"docs/node-operators/installation/native-build.md",
		"docs/node-operators/installation/docker-setup.md",
	} {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for _, forbidden := range []string{
			"go install -ldflags=\"$(LDFLAGS)\" ./",
			"tendermint unsafe-reset-all",
			"docker volume rm truerepublic_node-data",
			"truerepublicd-new migrate",
			"Install to $GOPATH/bin",
		} {
			if strings.Contains(string(body), forbidden) {
				t.Errorf("%s contains unsafe or unsupported guidance %q", path, forbidden)
			}
		}
	}
}
