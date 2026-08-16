package installlifecycle

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	hex40 = regexp.MustCompile(`^[0-9a-f]{40}$`)
	hex64 = regexp.MustCompile(`^[0-9a-f]{64}$`)
	name  = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)
)

func LoadContract(path string) (Contract, error) {
	var c Contract
	if err := parseFile(path, &c); err != nil {
		return c, err
	}
	return c, validateTemplate(c)
}

func validateTemplate(c Contract) error {
	if c.Schema != ContractSchema || len(c.SupportedTargets) == 0 || len(c.TargetRuntimes) != len(c.SupportedTargets) || c.OperatorStateBoundary != "outside-prefix" || c.Modes.Binary != 0755 || c.Modes.Metadata != 0644 || c.Modes.Transaction != 0600 || c.Limits.MaxArtifactBytes <= 0 || c.Limits.MaxArtifactBytes > 1<<30 {
		return errors.New("lifecycle contract is invalid")
	}
	seenTarget := map[string]bool{}
	for _, target := range c.SupportedTargets {
		if !name.MatchString(target) || seenTarget[target] || !name.MatchString(c.TargetRuntimes[target]) {
			return errors.New("supported targets are invalid")
		}
		seenTarget[target] = true
	}
	for target := range c.TargetRuntimes {
		if !seenTarget[target] {
			return errors.New("target runtime mapping is invalid")
		}
	}
	seenPath := map[string]bool{}
	paths := []string{c.Layout.Binary, c.Layout.Manifest, c.Layout.RollbackBinary, c.Layout.Transaction}
	for _, p := range paths {
		if p == "" || filepath.IsAbs(p) || filepath.Clean(p) != p || p == "." || strings.HasPrefix(p, ".."+string(filepath.Separator)) || seenPath[p] {
			return errors.New("managed layout is invalid")
		}
		seenPath[p] = true
	}
	for i, left := range paths {
		for j, right := range paths {
			if i != j && within(left, right) {
				return errors.New("managed layout paths must not overlap")
			}
		}
	}
	return nil
}

func Bind(c Contract, prefix, operatorState, sha, sourceRef, target, runtime string) (Contract, error) {
	if err := validateTemplate(c); err != nil {
		return c, err
	}
	if !hex40.MatchString(sourceRef) || !hex64.MatchString(sha) || !contains(c.SupportedTargets, target) || c.TargetRuntimes[target] != runtime {
		return c, errors.New("requested identity is invalid or unsupported")
	}
	c.Prefix, c.OperatorStatePath = prefix, operatorState
	c.BinaryPath, c.ManifestPath = filepath.Join(prefix, c.Layout.Binary), filepath.Join(prefix, c.Layout.Manifest)
	c.RollbackPath, c.TransactionPath = filepath.Join(prefix, c.Layout.RollbackBinary), filepath.Join(prefix, c.Layout.Transaction)
	c.ArtifactSHA256, c.SourceRef, c.Target, c.Runtime = sha, sourceRef, target, runtime
	if err := validateContract(c); err != nil {
		return c, err
	}
	return c, nil
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func validateContract(c Contract) error {
	if err := validateTemplate(c); err != nil {
		return err
	}
	if !hex40.MatchString(c.SourceRef) || !hex64.MatchString(c.ArtifactSHA256) || !contains(c.SupportedTargets, c.Target) || !name.MatchString(c.Runtime) {
		return errors.New("bound identity is invalid")
	}
	paths := []string{c.Prefix, c.BinaryPath, c.ManifestPath, c.RollbackPath, c.TransactionPath, c.OperatorStatePath}
	for _, p := range paths {
		if p == "" || !filepath.IsAbs(p) || filepath.Clean(p) != p || p == string(filepath.Separator) {
			return errors.New("contract paths must be absolute and clean")
		}
	}
	for _, p := range []string{c.BinaryPath, c.ManifestPath, c.RollbackPath, c.TransactionPath} {
		if !within(c.Prefix, p) || p == c.Prefix {
			return errors.New("managed paths must be isolated beneath prefix")
		}
	}
	if within(c.Prefix, c.OperatorStatePath) || within(c.OperatorStatePath, c.Prefix) {
		return errors.New("operator state must be outside the managed prefix")
	}
	seen := map[string]bool{}
	for _, p := range []string{c.BinaryPath, c.ManifestPath, c.RollbackPath, c.TransactionPath} {
		if seen[p] {
			return errors.New("managed paths must be distinct")
		}
		seen[p] = true
	}
	return nil
}

func within(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func Install(c Contract, artifact string) error {
	if err := validateContract(c); err != nil {
		return err
	}
	if within(c.Prefix, artifact) {
		return errors.New("artifact must be outside managed prefix")
	}
	if entries, err := os.ReadDir(c.Prefix); err == nil && len(entries) != 0 {
		return errors.New("managed prefix must be empty")
	} else if err != nil && !os.IsNotExist(err) {
		return errors.New("managed prefix is unavailable")
	}
	if err := verifyArtifact(artifact, c.ArtifactSHA256, c.Limits.MaxArtifactBytes); err != nil {
		return err
	}
	return mutate(c, "install", func() error {
		if err := atomicCopy(artifact, c.BinaryPath, os.FileMode(c.Modes.Binary)); err != nil {
			return err
		}
		if err := verifyArtifact(c.BinaryPath, c.ArtifactSHA256, c.Limits.MaxArtifactBytes); err != nil {
			return errors.New("installed artifact re-verification failed")
		}
		m := Manifest{Schema: ManifestSchema, BinaryPath: c.BinaryPath, OperatorStatePath: c.OperatorStatePath, Current: identity(c), Generation: 1}
		return atomicJSON(c.ManifestPath, m, os.FileMode(c.Modes.Metadata))
	})
}

func Upgrade(c Contract, artifact, expectedCurrentSHA, expectedCurrentSource string) error {
	if err := validateContract(c); err != nil {
		return err
	}
	if !hex64.MatchString(expectedCurrentSHA) {
		return errors.New("expected current digest is invalid")
	}
	if within(c.Prefix, artifact) {
		return errors.New("artifact must be outside managed prefix")
	}
	m, err := verifiedManifest(c)
	if err != nil {
		return err
	}
	if !hex40.MatchString(expectedCurrentSource) || m.Current.SHA256 != expectedCurrentSHA || m.Current.SourceRef != expectedCurrentSource {
		return errors.New("installed identity does not match expected current digest")
	}
	if m.Current.Target != c.Target || m.Current.Runtime != c.Runtime {
		return errors.New("upgrade target/runtime differs from installed identity")
	}
	if c.ArtifactSHA256 == m.Current.SHA256 {
		return errors.New("candidate must differ from installed artifact")
	}
	if err := verifyArtifact(artifact, c.ArtifactSHA256, c.Limits.MaxArtifactBytes); err != nil {
		return err
	}
	return mutate(c, "upgrade", func() error {
		if err := atomicCopy(c.BinaryPath, c.RollbackPath, os.FileMode(c.Modes.Binary)); err != nil {
			return err
		}
		old, oldGeneration := m.Current, m.Generation
		if err := atomicCopy(artifact, c.BinaryPath, os.FileMode(c.Modes.Binary)); err != nil {
			return err
		}
		if err := verifyArtifact(c.BinaryPath, c.ArtifactSHA256, c.Limits.MaxArtifactBytes); err != nil {
			return errors.New("upgraded artifact re-verification failed")
		}
		m.Current, m.Rollback, m.RollbackGeneration = identity(c), &old, &oldGeneration
		m.Generation++
		return atomicJSON(c.ManifestPath, m, os.FileMode(c.Modes.Metadata))
	})
}

func Rollback(c Contract) error {
	if err := validateContract(c); err != nil {
		return err
	}
	m, err := verifiedManifest(c)
	if err != nil {
		return err
	}
	if m.Current != identity(c) {
		return errors.New("installed identity does not match rollback request")
	}
	if m.Rollback == nil || m.RollbackGeneration == nil || *m.RollbackGeneration >= m.Generation || !exists(c.RollbackPath) {
		return errors.New("no rollback snapshot is available")
	}
	if err := verifyArtifact(c.RollbackPath, m.Rollback.SHA256, c.Limits.MaxArtifactBytes); err != nil {
		return errors.New("rollback snapshot identity mismatch")
	}
	return mutate(c, "rollback", func() error {
		if err := atomicCopy(c.RollbackPath, c.BinaryPath, os.FileMode(c.Modes.Binary)); err != nil {
			return err
		}
		if err := verifyArtifact(c.BinaryPath, m.Rollback.SHA256, c.Limits.MaxArtifactBytes); err != nil {
			return errors.New("restored artifact re-verification failed")
		}
		m.Current, m.Rollback, m.Generation, m.RollbackGeneration = *m.Rollback, nil, m.Generation+1, nil
		if err := atomicJSON(c.ManifestPath, m, os.FileMode(c.Modes.Metadata)); err != nil {
			return err
		}
		return os.Remove(c.RollbackPath)
	})
}

func Uninstall(c Contract, expectedCurrentSHA string) error {
	if err := validateContract(c); err != nil {
		return err
	}
	m, err := verifiedManifest(c)
	if err != nil {
		return err
	}
	if !hex64.MatchString(expectedCurrentSHA) || m.Current.SHA256 != expectedCurrentSHA {
		return errors.New("installed identity does not match expected current digest")
	}
	if m.Current != identity(c) {
		return errors.New("installed identity does not match uninstall request")
	}
	if err := ensureExactAllowlist(c); err != nil {
		return err
	}
	return mutate(c, "uninstall", func() error {
		for _, p := range []string{c.RollbackPath, c.BinaryPath, c.ManifestPath} {
			if err := removeRegularOnly(p); err != nil {
				return err
			}
		}
		return nil
	})
}

func Check(c Contract) (Status, error) {
	if err := validateContract(c); err != nil {
		return Status{}, err
	}
	if exists(c.TransactionPath) {
		return Status{Installed: exists(c.ManifestPath), Problem: "incomplete transaction marker exists"}, errors.New("incomplete transaction marker exists")
	}
	if !exists(c.ManifestPath) && !exists(c.BinaryPath) && !exists(c.RollbackPath) {
		return Status{}, nil
	}
	m, err := verifiedManifest(c)
	if err != nil {
		return Status{Installed: true, Problem: err.Error()}, err
	}
	if m.Current != identity(c) {
		return Status{Installed: true, Problem: "installed identity differs from requested binding"}, errors.New("installed identity differs from requested binding")
	}
	return Status{Installed: true, Healthy: true, Manifest: &m}, nil
}

func PreStart(c Contract) error {
	s, err := Check(c)
	if err != nil {
		return err
	}
	if !s.Installed || !s.Healthy {
		return errors.New("verified installation is unavailable")
	}
	return nil
}

func identity(c Contract) Identity {
	return Identity{SHA256: c.ArtifactSHA256, SourceRef: c.SourceRef, Target: c.Target, Runtime: c.Runtime}
}

func verifiedManifest(c Contract) (Manifest, error) {
	var m Manifest
	if exists(c.TransactionPath) {
		return m, errors.New("incomplete transaction marker exists")
	}
	if err := rejectAbsoluteSymlinks(c.ManifestPath); err != nil {
		return m, err
	}
	if err := parseFile(c.ManifestPath, &m); err != nil {
		return m, errors.New("install manifest is invalid")
	}
	if m.Schema != ManifestSchema || m.Generation == 0 || m.BinaryPath != c.BinaryPath || m.OperatorStatePath != c.OperatorStatePath || !validIdentity(m.Current) {
		return m, errors.New("install manifest binding is invalid")
	}
	if (m.Rollback == nil) != (m.RollbackGeneration == nil) || (m.Rollback != nil && (!validIdentity(*m.Rollback) || *m.RollbackGeneration >= m.Generation)) {
		return m, errors.New("rollback identity is invalid")
	}
	if err := verifyArtifact(c.BinaryPath, m.Current.SHA256, c.Limits.MaxArtifactBytes); err != nil {
		return m, errors.New("installed binary identity mismatch")
	}
	return m, nil
}

func validIdentity(i Identity) bool {
	return hex64.MatchString(i.SHA256) && hex40.MatchString(i.SourceRef) && name.MatchString(i.Target) && name.MatchString(i.Runtime)
}

func mutate(c Contract, operation string, fn func() error) error {
	if exists(c.TransactionPath) {
		return errors.New("incomplete transaction marker exists")
	}
	for _, p := range []string{c.Prefix, c.BinaryPath, c.ManifestPath, c.RollbackPath, c.TransactionPath} {
		if err := rejectAbsoluteSymlinks(p); err != nil {
			return err
		}
	}
	if err := atomicJSON(c.TransactionPath, transaction{Schema: TransactionSchema, Operation: operation}, os.FileMode(c.Modes.Transaction)); err != nil {
		return err
	}
	if err := fn(); err != nil {
		return fmt.Errorf("%s failed; transaction marker retained: %w", operation, err)
	}
	if err := os.Remove(c.TransactionPath); err != nil {
		return fmt.Errorf("operation completed but transaction marker could not be removed: %w", err)
	}
	return syncDir(filepath.Dir(c.TransactionPath))
}

func verifyArtifact(path, want string, maxBytes int64) error {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return errors.New("artifact path must be absolute and clean")
	}
	if err := rejectAbsoluteSymlinks(path); err != nil {
		return err
	}
	pathInfo, err := os.Lstat(path)
	if err != nil || !pathInfo.Mode().IsRegular() {
		return errors.New("artifact must be a regular file")
	}
	f, err := os.Open(path)
	if err != nil {
		return errors.New("artifact is unavailable")
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !os.SameFile(pathInfo, info) || !info.Mode().IsRegular() || info.Size() <= 0 || info.Size() > maxBytes || info.Mode().Perm()&0100 == 0 || info.Mode().Perm()&0022 != 0 || info.Mode()&(os.ModeSetuid|os.ModeSetgid|os.ModeSticky) != 0 {
		return errors.New("artifact must be a regular file")
	}
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return errors.New("artifact cannot be read")
	}
	if hex.EncodeToString(h.Sum(nil)) != want {
		return errors.New("artifact digest mismatch")
	}
	return nil
}

func ensureExactAllowlist(c Contract) error {
	allowedFiles := map[string]bool{c.BinaryPath: true, c.ManifestPath: true, c.RollbackPath: true}
	allowedDirs := map[string]bool{c.Prefix: true}
	for p := range allowedFiles {
		for d := filepath.Dir(p); within(c.Prefix, d) || d == c.Prefix; d = filepath.Dir(d) {
			allowedDirs[d] = true
			if d == c.Prefix {
				break
			}
		}
	}
	return filepath.WalkDir(c.Prefix, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return errors.New("symlink in managed prefix")
		}
		if entry.IsDir() {
			if !allowedDirs[path] {
				return errors.New("unknown directory in managed prefix")
			}
			return nil
		}
		if !allowedFiles[path] {
			return errors.New("unknown file in managed prefix")
		}
		if info, err := entry.Info(); err != nil || !info.Mode().IsRegular() {
			return errors.New("managed entry is not a regular file")
		}
		return nil
	})
}

func atomicCopy(source, destination string, mode os.FileMode) error {
	s, err := os.Open(source)
	if err != nil {
		return err
	}
	defer s.Close()
	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return err
	}
	t, err := os.CreateTemp(filepath.Dir(destination), ".lifecycle-*")
	if err != nil {
		return err
	}
	tmp := t.Name()
	defer os.Remove(tmp)
	if _, err = io.Copy(t, s); err == nil {
		err = t.Chmod(mode)
	}
	if err == nil {
		err = t.Sync()
	}
	if closeErr := t.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err = os.Rename(tmp, destination); err != nil {
		return err
	}
	return syncDir(filepath.Dir(destination))
}

func atomicJSON(path string, value any, mode os.FileMode) error {
	b, err := marshal(value)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	t, err := os.CreateTemp(filepath.Dir(path), ".lifecycle-*")
	if err != nil {
		return err
	}
	tmp := t.Name()
	defer os.Remove(tmp)
	if _, err = t.Write(b); err == nil {
		err = t.Chmod(mode)
	}
	if err == nil {
		err = t.Sync()
	}
	if closeErr := t.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err = os.Rename(tmp, path); err != nil {
		return err
	}
	return syncDir(filepath.Dir(path))
}

func removeRegularOnly(path string) error {
	i, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !i.Mode().IsRegular() {
		return errors.New("managed path is not a regular file")
	}
	return os.Remove(path)
}

func rejectAbsoluteSymlinks(path string) error {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return errors.New("path must be absolute and clean")
	}
	volume := filepath.VolumeName(path)
	cur := volume + string(filepath.Separator)
	rel := strings.TrimPrefix(path, cur)
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		if part == "" {
			continue
		}
		cur = filepath.Join(cur, part)
		info, err := os.Lstat(cur)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return errors.New("symlink path rejected")
		}
	}
	return nil
}

func exists(path string) bool { _, err := os.Lstat(path); return err == nil }
func syncDir(path string) error {
	d, err := os.Open(path)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}
