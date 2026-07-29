package healthcheck

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// runCLI executes the healthcheck command group with the given args and
// returns stdout, stderr, and the execution error.
func runCLI(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := NewCommand()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func TestCLILiveSuccessOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(validHealthBody))
	}))
	t.Cleanup(srv.Close)

	stdout, stderr, err := runCLI(t, "live", "--rpc-url", srv.URL)
	if err != nil {
		t.Fatalf("live command failed: %v", err)
	}
	if stdout != "live\n" {
		t.Fatalf("live stdout must be exactly %q, got %q", "live\n", stdout)
	}
	if stderr != "" {
		t.Fatalf("live stderr must be empty, got %q", stderr)
	}
}

func TestCLIReadySuccessOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(validStatusBody))
	}))
	t.Cleanup(srv.Close)

	stdout, stderr, err := runCLI(t, "ready", "--rpc-url", srv.URL, "--timeout", "5s")
	if err != nil {
		t.Fatalf("ready command failed: %v", err)
	}
	if stdout != "ready\n" {
		t.Fatalf("ready stdout must be exactly %q, got %q", "ready\n", stdout)
	}
	if stderr != "" {
		t.Fatalf("ready stderr must be empty, got %q", stderr)
	}
}

func TestCLIProbeDefaults(t *testing.T) {
	for _, name := range []string{"live", "ready"} {
		cmd := NewCommand()
		probe, _, err := cmd.Find([]string{name})
		if err != nil {
			t.Fatal(err)
		}
		if got := probe.Flag("rpc-url").DefValue; got != DefaultRPCURL {
			t.Fatalf("%s rpc-url default = %q, want %q", name, got, DefaultRPCURL)
		}
		if got := probe.Flag("timeout").DefValue; got != DefaultTimeout.String() {
			t.Fatalf("%s timeout default = %q, want %q", name, got, DefaultTimeout)
		}
	}
}

func TestCLIProbeFailures(t *testing.T) {
	failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","result":{"sync_info":{"latest_block_height":"7","catching_up":true}}}`))
	}))
	t.Cleanup(failing.Close)

	cases := map[string]struct {
		args []string
		err  error
	}{
		"live invalid url":      {[]string{"live", "--rpc-url", "https://127.0.0.1:26657"}, errScheme},
		"ready invalid url":     {[]string{"ready", "--rpc-url", "http://127.0.0.1:26657/x"}, errPath},
		"live bad timeout":      {[]string{"live", "--timeout", "0s"}, errTimeoutRange},
		"ready timeout too big": {[]string{"ready", "--timeout", "30s"}, errTimeoutRange},
		"ready catching up":     {[]string{"ready", "--rpc-url", failing.URL}, errCatchingUp},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			stdout, stderr, err := runCLI(t, tc.args...)
			if !errors.Is(err, tc.err) {
				t.Fatalf("expected %v, got %v", tc.err, err)
			}
			if stdout != "" || stderr != "" {
				t.Fatalf("failed probe must print nothing, got stdout=%q stderr=%q", stdout, stderr)
			}
		})
	}
}

func TestCLIRejectsUnexpectedArgs(t *testing.T) {
	if _, _, err := runCLI(t, "live", "extra"); err == nil {
		t.Fatal("live must reject positional arguments")
	}
	if _, _, err := runCLI(t, "ready", "extra"); err == nil {
		t.Fatal("ready must reject positional arguments")
	}
}
