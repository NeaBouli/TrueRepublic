package installlifecycle

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

const sourceRef = "0123456789abcdef0123456789abcdef01234567"

func TestLifecycleInstallUpgradeRollbackUninstall(t *testing.T) {
	root := t.TempDir()
	operatorState := filepath.Join(t.TempDir(), "state")
	if err := os.MkdirAll(operatorState, 0700); err != nil {
		t.Fatal(err)
	}
	stateFile := filepath.Join(operatorState, "priv_validator_key.json")
	if err := os.WriteFile(stateFile, []byte("must survive"), 0600); err != nil {
		t.Fatal(err)
	}
	one := artifact(t, t.TempDir(), "one", "version-one")
	c := contract(root, operatorState, digest(t, one))

	if err := Install(c, one); err != nil {
		t.Fatalf("install: %v", err)
	}
	if err := PreStart(c); err != nil {
		t.Fatalf("pre-start: %v", err)
	}
	if info, err := os.Stat(c.BinaryPath); err != nil || info.Mode().Perm() != 0755 {
		t.Fatalf("binary permissions: %v %v", info, err)
	}
	if err := Install(c, one); err == nil {
		t.Fatal("duplicate install accepted")
	}

	two := artifact(t, t.TempDir(), "two", "version-two")
	oldDigest := c.ArtifactSHA256
	c.ArtifactSHA256 = digest(t, two)
	if err := Upgrade(c, two, strings.Repeat("f", 64), sourceRef); err == nil {
		t.Fatal("wrong expected digest accepted")
	}
	if err := Upgrade(c, two, oldDigest, strings.Repeat("f", 40)); err == nil {
		t.Fatal("wrong expected source accepted")
	}
	if err := Upgrade(c, two, oldDigest, sourceRef); err != nil {
		t.Fatalf("upgrade: %v", err)
	}
	if err := Upgrade(c, two, c.ArtifactSHA256, sourceRef); err == nil {
		t.Fatal("upgrade to identical candidate accepted")
	}
	three := artifact(t, t.TempDir(), "three", "version-three")
	twoDigest := c.ArtifactSHA256
	c.ArtifactSHA256 = digest(t, three)
	if err := Upgrade(c, three, twoDigest, sourceRef); err != nil {
		t.Fatalf("second upgrade: %v", err)
	}
	if err := Rollback(c); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	var afterRollback Manifest
	if err := parseFile(c.ManifestPath, &afterRollback); err != nil {
		t.Fatal(err)
	}
	if afterRollback.Generation != 4 {
		t.Fatalf("rollback generation = %d, want 4", afterRollback.Generation)
	}
	if err := Rollback(c); err == nil {
		t.Fatal("rollback replay accepted")
	}
	if got := string(mustRead(t, c.BinaryPath)); got != "version-two" {
		t.Fatalf("rollback content %q", got)
	}
	c.ArtifactSHA256 = twoDigest
	if err := Uninstall(c, twoDigest); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	if exists(c.BinaryPath) || exists(c.ManifestPath) {
		t.Fatal("managed files survived uninstall")
	}
	for _, dir := range []string{filepath.Dir(c.BinaryPath), filepath.Dir(c.RollbackPath), filepath.Dir(c.ManifestPath)} {
		if exists(dir) {
			t.Fatalf("empty managed directory survived uninstall: %s", dir)
		}
	}
	if got := string(mustRead(t, stateFile)); got != "must survive" {
		t.Fatalf("operator state changed: %q", got)
	}
	if err := Install(c, two); err != nil {
		t.Fatalf("reinstall after uninstall: %v", err)
	}
}

func TestTransactionMarkerCreationIsExclusive(t *testing.T) {
	path := filepath.Join(canonicalTestPath(t.TempDir()), "lib", "install-transaction.json")
	start := make(chan struct{})
	results := make(chan error, 2)
	var ready sync.WaitGroup
	ready.Add(2)
	for range 2 {
		go func() {
			ready.Done()
			<-start
			results <- exclusiveJSON(path, transaction{Schema: TransactionSchema, Operation: "install"}, 0600)
		}()
	}
	ready.Wait()
	close(start)
	succeeded := 0
	for range 2 {
		if err := <-results; err == nil {
			succeeded++
		}
	}
	if succeeded != 1 {
		t.Fatalf("exclusive marker successes = %d, want 1", succeeded)
	}
}

func TestFailClosedAdversarialBoundaries(t *testing.T) {
	root := t.TempDir()
	one := artifact(t, t.TempDir(), "artifact", "trusted")
	c := contract(root, filepath.Join(t.TempDir(), "state"), digest(t, one))

	t.Run("digest mismatch", func(t *testing.T) {
		bad := c
		bad.ArtifactSHA256 = strings.Repeat("0", 64)
		if err := Install(bad, one); err == nil {
			t.Fatal("mismatch accepted")
		}
		if exists(bad.TransactionPath) {
			t.Fatal("preflight rejection mutated the managed prefix")
		}
	})

	t.Run("symlink component", func(t *testing.T) {
		dir := t.TempDir()
		outside := t.TempDir()
		if err := os.Symlink(outside, filepath.Join(dir, "managed")); err != nil {
			t.Fatal(err)
		}
		bad := contractRaw(filepath.Join(dir, "managed"), canonicalTestPath(filepath.Join(t.TempDir(), "state")), digest(t, one))
		if err := Install(bad, one); err == nil {
			t.Fatal("symlink prefix accepted")
		}
	})

	t.Run("symlink ancestor", func(t *testing.T) {
		dir := t.TempDir()
		outside := t.TempDir()
		link := filepath.Join(dir, "ancestor")
		if err := os.Symlink(outside, link); err != nil {
			t.Fatal(err)
		}
		bad := contractRaw(filepath.Join(link, "managed"), canonicalTestPath(filepath.Join(t.TempDir(), "state")), digest(t, one))
		if err := Install(bad, one); err == nil {
			t.Fatal("symlink ancestor accepted")
		}
	})

	t.Run("path traversal and overlapping state", func(t *testing.T) {
		bad := c
		bad.BinaryPath = filepath.Join(c.Prefix, "..", "escape")
		if validateContract(bad) == nil {
			t.Fatal("unclean path accepted")
		}
		bad = c
		bad.OperatorStatePath = filepath.Join(c.Prefix, "state")
		if validateContract(bad) == nil {
			t.Fatal("managed operator state accepted")
		}
	})

	t.Run("unsupported target runtime binding", func(t *testing.T) {
		if _, err := Bind(templateContract(), canonicalTestPath(t.TempDir()), canonicalTestPath(t.TempDir()), digest(t, one), sourceRef, "linux-amd64", "musl"); err == nil {
			t.Fatal("wrong target runtime accepted")
		}
	})

	t.Run("transaction marker", func(t *testing.T) {
		dir := t.TempDir()
		good := contract(dir, filepath.Join(t.TempDir(), "state"), digest(t, one))
		if err := os.MkdirAll(filepath.Dir(good.TransactionPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(good.TransactionPath, []byte(`{"schema":"truerepublic.install-transaction/v1","operation":"install"}`), 0600); err != nil {
			t.Fatal(err)
		}
		if err := Install(good, one); err == nil {
			t.Fatal("existing transaction accepted")
		}
		if err := PreStart(good); err == nil {
			t.Fatal("pre-start accepted transaction")
		}

		dedicated := contract(t.TempDir(), filepath.Join(t.TempDir(), "state"), digest(t, one))
		dedicated.TransactionPath = filepath.Join(dedicated.Prefix, "transaction", "install.json")
		if err := Install(dedicated, one); err != nil {
			t.Fatalf("install with dedicated transaction directory: %v", err)
		}
		if err := Uninstall(dedicated, dedicated.ArtifactSHA256); err != nil {
			t.Fatalf("uninstall with dedicated transaction directory: %v", err)
		}
		if exists(filepath.Dir(dedicated.TransactionPath)) {
			t.Fatal("dedicated transaction directory survived uninstall")
		}
	})

	t.Run("tampered installed binary", func(t *testing.T) {
		dir := t.TempDir()
		good := contract(dir, filepath.Join(t.TempDir(), "state"), digest(t, one))
		if err := Install(good, one); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(good.BinaryPath, []byte("tampered"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := PreStart(good); err == nil {
			t.Fatal("tampered binary accepted")
		}
	})

	t.Run("unsafe artifact permissions", func(t *testing.T) {
		unsafe := artifact(t, t.TempDir(), "unsafe", "binary")
		if err := os.Chmod(unsafe, 0777); err != nil {
			t.Fatal(err)
		}
		good := contract(t.TempDir(), filepath.Join(t.TempDir(), "state"), digest(t, unsafe))
		if err := Install(good, unsafe); err == nil {
			t.Fatal("world-writable artifact accepted")
		}
	})

	t.Run("owner executable is required", func(t *testing.T) {
		unsafe := artifact(t, t.TempDir(), "unsafe-owner", "binary")
		if err := os.Chmod(unsafe, 0655); err != nil {
			t.Fatal(err)
		}
		good := contract(t.TempDir(), filepath.Join(t.TempDir(), "state"), digest(t, unsafe))
		if err := Install(good, unsafe); err == nil {
			t.Fatal("artifact without owner execute accepted")
		}
	})

	t.Run("partial prefix", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "unexpected"), []byte("state"), 0600); err != nil {
			t.Fatal(err)
		}
		good := contract(dir, filepath.Join(t.TempDir(), "state"), digest(t, one))
		if err := Install(good, one); err == nil {
			t.Fatal("partial prefix accepted")
		}
	})

	t.Run("orphan rollback is partial state", func(t *testing.T) {
		good := contract(t.TempDir(), filepath.Join(t.TempDir(), "state"), digest(t, one))
		if err := os.MkdirAll(filepath.Dir(good.RollbackPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(good.RollbackPath, []byte("orphan"), 0755); err != nil {
			t.Fatal(err)
		}
		if _, err := Check(good); err == nil {
			t.Fatal("orphan rollback reported as an absent clean install")
		}
	})

	t.Run("upgrade artifact inside prefix", func(t *testing.T) {
		good := contract(t.TempDir(), filepath.Join(t.TempDir(), "state"), digest(t, one))
		if err := Install(good, one); err != nil {
			t.Fatal(err)
		}
		candidate := filepath.Join(good.Prefix, "candidate")
		if err := os.WriteFile(candidate, []byte("candidate"), 0755); err != nil {
			t.Fatal(err)
		}
		old := good.ArtifactSHA256
		good.ArtifactSHA256 = digest(t, candidate)
		if err := Upgrade(good, candidate, old, sourceRef); err == nil {
			t.Fatal("upgrade artifact inside managed prefix accepted")
		}
	})

	t.Run("oversized artifact", func(t *testing.T) {
		large := artifact(t, t.TempDir(), "large", "too-large")
		good := contract(t.TempDir(), filepath.Join(t.TempDir(), "state"), digest(t, large))
		good.Limits.MaxArtifactBytes = 3
		if err := Install(good, large); err == nil {
			t.Fatal("oversized artifact accepted")
		}
	})

	t.Run("unknown manifest field", func(t *testing.T) {
		good := contract(t.TempDir(), filepath.Join(t.TempDir(), "state"), digest(t, one))
		if err := Install(good, one); err != nil {
			t.Fatal(err)
		}
		manifest := bytes.TrimSpace(mustRead(t, good.ManifestPath))
		manifest = append(append([]byte(nil), bytes.TrimSuffix(manifest, []byte("}"))...), []byte(`,"unknown":true}`)...)
		if err := os.WriteFile(good.ManifestPath, manifest, 0644); err != nil {
			t.Fatal(err)
		}
		if err := PreStart(good); err == nil {
			t.Fatal("unknown manifest field accepted")
		}
	})

	t.Run("tampered rollback and unknown uninstall entry", func(t *testing.T) {
		dir := t.TempDir()
		first := artifact(t, t.TempDir(), "first", "first")
		second := artifact(t, t.TempDir(), "second", "second")
		good := contract(dir, filepath.Join(t.TempDir(), "state"), digest(t, first))
		if err := Install(good, first); err != nil {
			t.Fatal(err)
		}
		old := good.ArtifactSHA256
		good.ArtifactSHA256 = digest(t, second)
		if err := Upgrade(good, second, old, sourceRef); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(good.RollbackPath, []byte("tampered"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := Rollback(good); err == nil {
			t.Fatal("tampered rollback accepted")
		}
		if err := os.WriteFile(filepath.Join(good.Prefix, "operator-key.json"), []byte("key"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := Uninstall(good, good.ArtifactSHA256); err == nil {
			t.Fatal("unknown prefix entry accepted")
		}
		if !exists(good.BinaryPath) {
			t.Fatal("uninstall deleted files before allowlist validation")
		}
	})
}

func TestStrictContractJSONAndCLI(t *testing.T) {
	root := t.TempDir()
	one := artifact(t, root, "artifact", "binary")
	c := contract(filepath.Join(root, "prefix"), filepath.Join(root, "state"), digest(t, one))
	b, err := json.Marshal(templateContract())
	if err != nil {
		t.Fatal(err)
	}
	for name, raw := range map[string][]byte{
		"duplicate": bytes.Replace(b, []byte(`"schema":`), []byte(`"schema":"truerepublic.install-lifecycle/v1","schema":`), 1),
		"unknown":   append(append([]byte(nil), bytes.TrimSuffix(b, []byte("}"))...), []byte(`,"unknown":true}`)...),
		"trailing":  append(append([]byte(nil), b...), []byte(` {}`)...),
		"deep":      []byte(strings.Repeat("[", 34) + "0" + strings.Repeat("]", 34)),
		"oversize":  []byte(`{"schema":"` + strings.Repeat("x", MaxJSONBytes) + `"}`),
	} {
		t.Run(name, func(t *testing.T) {
			p := filepath.Join(root, name+".json")
			if err := os.WriteFile(p, raw, 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadContract(p); err == nil {
				t.Fatal("invalid JSON accepted")
			}
		})
	}
	contractPath := filepath.Join(root, "contract.json")
	if err := os.WriteFile(contractPath, b, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := newCommandForHost(c.Target)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--contract", contractPath, "--prefix", c.Prefix, "--operator-state", c.OperatorStatePath, "--sha256", c.ArtifactSHA256, "--source-ref", c.SourceRef, "--target", c.Target, "--runtime", c.Runtime, "status"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"installed": false`) {
		t.Fatalf("unexpected status: %s", out.String())
	}
	wrongHost := newCommandForHost("linux-arm64")
	wrongHost.SetArgs([]string{"--contract", contractPath, "--prefix", c.Prefix, "--operator-state", c.OperatorStatePath, "--sha256", c.ArtifactSHA256, "--source-ref", c.SourceRef, "--target", c.Target, "--runtime", c.Runtime, "status"})
	if err := wrongHost.Execute(); err == nil {
		t.Fatal("CLI accepted a target that differs from the host")
	}
	if cmd := NewCommand(); commandMissingRequiredFlagSucceeds(cmd, []string{"install"}) {
		t.Fatal("install without required flags accepted")
	}
}

func commandMissingRequiredFlagSucceeds(cmd interface {
	SetArgs([]string)
	Execute() error
}, args []string) bool {
	cmd.SetArgs(args)
	return cmd.Execute() == nil
}

func contract(prefix, state, digest string) Contract {
	prefix = canonicalTestPath(prefix)
	state = canonicalTestPath(state)
	return contractRaw(prefix, state, digest)
}

func contractRaw(prefix, state, digest string) Contract {
	c, err := Bind(templateContract(), prefix, state, digest, sourceRef, "linux-amd64", "linux-glibc")
	if err != nil {
		panic(err)
	}
	return c
}

func templateContract() Contract {
	return Contract{Schema: ContractSchema, SupportedTargets: []string{"linux-amd64", "linux-arm64"}, TargetRuntimes: map[string]string{"linux-amd64": "linux-glibc", "linux-arm64": "linux-glibc"}, Layout: Layout{Binary: "bin/truerepublicd", Manifest: "lib/install-manifest.json", RollbackBinary: "lib/rollback/truerepublicd", Transaction: "lib/install-transaction.json"}, Modes: Modes{Binary: 0755, Metadata: 0644, Transaction: 0600}, Limits: Limits{MaxArtifactBytes: 1 << 20}, OperatorStateBoundary: "outside-prefix"}
}

func artifact(t *testing.T, dir, name, contents string) string {
	t.Helper()
	dir = canonicalTestPath(dir)
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(contents), 0755); err != nil {
		t.Fatal(err)
	}
	return p
}

func canonicalTestPath(path string) string {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(path))
	if err != nil {
		return path
	}
	return filepath.Join(parent, filepath.Base(path))
}
func digest(t *testing.T, path string) string {
	t.Helper()
	sum := sha256.Sum256(mustRead(t, path))
	return hex.EncodeToString(sum[:])
}
func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
