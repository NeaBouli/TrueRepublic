package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// Multi-process harnesses hand fixed loopback ports to child nodes. The former
// allocator opened 127.0.0.1:0, recorded the kernel-assigned port and closed
// the listener (GH-324/GH-325). That port came from the ephemeral range, so the
// kernel could hand it out again, e.g. as the source port of a peer dial,
// before the child bound it. The allocator below closes that window:
//
//   - ports come from a fixed range below every default ephemeral range
//     (Linux 32768+, macOS/BSD 49152+) and never overlap the host's configured
//     ephemeral range, so implicit binds and outgoing connections cannot take
//     them;
//   - a process-wide registry never returns a port twice while it is reserved;
//   - an exclusive lock file per port reserves it across concurrent test
//     processes on the same host (shared self-hosted runners);
//   - every candidate is probed for bindability before it is returned.
//
// Reservations are released through t.Cleanup, after the harness cleanups that
// stop the child processes registered later.
const (
	harnessPortRangeFirst = 20000
	harnessPortRangeLast  = 29999
)

var harnessPortRegistry = struct {
	sync.Mutex
	reserved map[int]struct{}
	next     int
}{reserved: make(map[int]struct{})}

// freeTCPPort reserves a loopback TCP port for a child process for the rest of
// the test. It fails the test instead of returning a port that is not
// exclusively reserved.
func freeTCPPort(t *testing.T) int {
	t.Helper()
	port, release, err := reserveHarnessPort(harnessPortLockDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(release)
	return port
}

func harnessPortLockDir() string {
	return filepath.Join(os.TempDir(), "truerepublic-harness-ports")
}

func reserveHarnessPort(lockDir string) (int, func(), error) {
	if err := os.MkdirAll(lockDir, 0o700); err != nil {
		return 0, nil, fmt.Errorf("create harness port lock directory: %w", err)
	}
	ephemeralFirst, ephemeralLast := hostEphemeralPortRange()

	harnessPortRegistry.Lock()
	defer harnessPortRegistry.Unlock()
	span := harnessPortRangeLast - harnessPortRangeFirst + 1
	if harnessPortRegistry.next == 0 {
		// Spread concurrent processes across the range to reduce lock contention.
		seed := os.Getpid()*7919 + time.Now().Nanosecond()
		if seed < 0 {
			seed = -seed
		}
		harnessPortRegistry.next = harnessPortRangeFirst + seed%span
	}
	start := harnessPortRegistry.next - harnessPortRangeFirst
	for offset := 0; offset < span; offset++ {
		port := harnessPortRangeFirst + (start+offset)%span
		if port >= ephemeralFirst && port <= ephemeralLast {
			continue
		}
		if _, reserved := harnessPortRegistry.reserved[port]; reserved {
			continue
		}
		lockPath := filepath.Join(lockDir, strconv.Itoa(port)+".lock")
		if !claimHarnessPortLock(lockPath) {
			continue
		}
		if !loopbackPortBindable(port) {
			releaseHarnessPortLock(lockPath)
			continue
		}
		harnessPortRegistry.reserved[port] = struct{}{}
		harnessPortRegistry.next = harnessPortRangeFirst + (port-harnessPortRangeFirst+1)%span
		release := func() {
			harnessPortRegistry.Lock()
			delete(harnessPortRegistry.reserved, port)
			harnessPortRegistry.Unlock()
			releaseHarnessPortLock(lockPath)
		}
		return port, release, nil
	}
	return 0, nil, fmt.Errorf("no reservable harness port in %d-%d", harnessPortRangeFirst, harnessPortRangeLast)
}

// claimHarnessPortLock creates the per-port lock exclusively. A lock whose
// owning process no longer exists is stale and is reclaimed once.
func claimHarnessPortLock(lockPath string) bool {
	for attempt := 0; attempt < 2; attempt++ {
		file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			_, writeErr := file.WriteString(strconv.Itoa(os.Getpid()))
			closeErr := file.Close()
			if writeErr != nil || closeErr != nil {
				_ = os.Remove(lockPath)
				return false
			}
			return true
		}
		if !errors.Is(err, os.ErrExist) || !harnessPortLockIsStale(lockPath) {
			return false
		}
		_ = os.Remove(lockPath)
	}
	return false
}

func harnessPortLockIsStale(lockPath string) bool {
	content, err := os.ReadFile(lockPath)
	if err != nil {
		return false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(content)))
	if err != nil || pid <= 0 {
		// A lock being written right now may still be empty; never steal it.
		return false
	}
	if pid == os.Getpid() {
		// This process only holds registered ports, which were skipped above.
		// An unregistered own lock is a leftover from an earlier release race.
		return true
	}
	if runtime.GOOS == "windows" {
		// No signal-0 liveness probe; keep foreign locks.
		return false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return true
	}
	signalErr := process.Signal(syscall.Signal(0))
	return signalErr != nil && !errors.Is(signalErr, syscall.EPERM)
}

// releaseHarnessPortLock removes the lock only when this process owns it.
func releaseHarnessPortLock(lockPath string) {
	content, err := os.ReadFile(lockPath)
	if err != nil || strings.TrimSpace(string(content)) != strconv.Itoa(os.Getpid()) {
		return
	}
	_ = os.Remove(lockPath)
}

func loopbackPortBindable(port int) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return false
	}
	return listener.Close() == nil
}

// hostEphemeralPortRange returns the configured Linux ephemeral range, or the
// IANA dynamic range elsewhere. Both defaults lie above the harness range; a
// host configured to overlap it simply has those ports skipped.
func hostEphemeralPortRange() (int, int) {
	content, err := os.ReadFile("/proc/sys/net/ipv4/ip_local_port_range")
	if err == nil {
		fields := strings.Fields(string(content))
		if len(fields) == 2 {
			first, firstErr := strconv.Atoi(fields[0])
			last, lastErr := strconv.Atoi(fields[1])
			if firstErr == nil && lastErr == nil && first <= last {
				return first, last
			}
		}
	}
	return 49152, 65535
}

func TestHarnessPortAllocatorReservesUniqueNonEphemeralPorts(t *testing.T) {
	lockDir := t.TempDir()
	ephemeralFirst, ephemeralLast := hostEphemeralPortRange()
	const count = 64
	seen := make(map[int]struct{}, count)
	releases := make([]func(), 0, count)
	for i := 0; i < count; i++ {
		port, release, err := reserveHarnessPort(lockDir)
		if err != nil {
			t.Fatal(err)
		}
		releases = append(releases, release)
		if port < harnessPortRangeFirst || port > harnessPortRangeLast {
			t.Fatalf("port %d outside harness range", port)
		}
		if port >= ephemeralFirst && port <= ephemeralLast {
			t.Fatalf("port %d inside host ephemeral range %d-%d", port, ephemeralFirst, ephemeralLast)
		}
		if _, duplicate := seen[port]; duplicate {
			t.Fatalf("port %d returned twice while reserved", port)
		}
		seen[port] = struct{}{}
		owner, err := os.ReadFile(filepath.Join(lockDir, strconv.Itoa(port)+".lock"))
		if err != nil || string(owner) != strconv.Itoa(os.Getpid()) {
			t.Fatalf("port %d lock = %q, %v; want own pid", port, owner, err)
		}
	}
	for _, release := range releases {
		release()
	}
	entries, err := os.ReadDir(lockDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("%d lock files left after release", len(entries))
	}
	harnessPortRegistry.Lock()
	defer harnessPortRegistry.Unlock()
	for port := range seen {
		if _, reserved := harnessPortRegistry.reserved[port]; reserved {
			t.Fatalf("port %d still registered after release", port)
		}
	}
}

func TestHarnessPortAllocatorHonorsForeignLocksAndReclaimsStaleOnes(t *testing.T) {
	lockDir := t.TempDir()
	ephemeralFirst, ephemeralLast := hostEphemeralPortRange()
	usable := func(port int) bool {
		harnessPortRegistry.Lock()
		_, reserved := harnessPortRegistry.reserved[port]
		harnessPortRegistry.Unlock()
		return !reserved && !(port >= ephemeralFirst && port <= ephemeralLast) && loopbackPortBindable(port)
	}
	base := 0
	for port := harnessPortRangeFirst; port < harnessPortRangeLast; port++ {
		if usable(port) && usable(port+1) {
			base = port
			break
		}
	}
	if base == 0 {
		t.Fatal("no two adjacent usable harness ports")
	}

	// A lock held by a live foreign process (the go test driver) is honored.
	livePID := os.Getppid()
	if err := os.WriteFile(filepath.Join(lockDir, strconv.Itoa(base)+".lock"), []byte(strconv.Itoa(livePID)), 0o600); err != nil {
		t.Fatal(err)
	}
	// A lock whose process has exited is stale and reclaimed.
	exited := exec.Command(os.Args[0], "-test.run=^$")
	if err := exited.Run(); err != nil {
		t.Fatal(err)
	}
	stalePath := filepath.Join(lockDir, strconv.Itoa(base+1)+".lock")
	if err := os.WriteFile(stalePath, []byte(strconv.Itoa(exited.Process.Pid)), 0o600); err != nil {
		t.Fatal(err)
	}

	harnessPortRegistry.Lock()
	harnessPortRegistry.next = base
	harnessPortRegistry.Unlock()
	port, release, err := reserveHarnessPort(lockDir)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if port != base+1 {
		t.Fatalf("reserved port %d, want stale-lock port %d after skipping live foreign lock %d", port, base+1, base)
	}
	if owner, _ := os.ReadFile(stalePath); string(owner) != strconv.Itoa(os.Getpid()) {
		t.Fatalf("stale lock not reclaimed: %q", owner)
	}
	if owner, _ := os.ReadFile(filepath.Join(lockDir, strconv.Itoa(base)+".lock")); string(owner) != strconv.Itoa(livePID) {
		t.Fatalf("foreign lock was modified: %q", owner)
	}
}

func TestHarnessPortAllocatorSkipsOccupiedPorts(t *testing.T) {
	lockDir := t.TempDir()
	occupied, release, err := reserveHarnessPort(lockDir)
	if err != nil {
		t.Fatal(err)
	}
	// Simulate an unrelated process binding the port without a lock.
	release()
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", occupied))
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	harnessPortRegistry.Lock()
	harnessPortRegistry.next = occupied
	harnessPortRegistry.Unlock()
	port, releaseNext, err := reserveHarnessPort(lockDir)
	if err != nil {
		t.Fatal(err)
	}
	defer releaseNext()
	if port == occupied {
		t.Fatalf("allocator returned bound port %d", port)
	}
	if _, err := os.Stat(filepath.Join(lockDir, strconv.Itoa(occupied)+".lock")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("lock for unbindable port %d was kept: %v", occupied, err)
	}
}

// TestSmokeValidatorExitIsReportedWithLogTail proves GH-325 fail-fast: a child
// that dies at startup is reported with its log tail, and stop() still
// observes the same exit through the done channel.
func TestSmokeValidatorExitIsReportedWithLogTail(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a POSIX shell stub")
	}
	dir := t.TempDir()
	binary := filepath.Join(dir, "exiting-node")
	script := "#!/bin/sh\necho 'bind: address already in use'\nexit 3\n"
	if err := os.WriteFile(binary, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	validator := &smokeValidator{
		name:    "validator-stub",
		home:    dir,
		rpcPort: freeTCPPort(t),
		p2pPort: freeTCPPort(t),
		logPath: filepath.Join(dir, "node.log"),
	}
	if err := validator.start(t.Context(), binary, ""); err != nil {
		t.Fatal(err)
	}
	select {
	case <-validator.exit.closed:
	case <-time.After(10 * time.Second):
		t.Fatal("stub process did not exit")
	}
	err := validator.exitedError()
	if err == nil {
		t.Fatal("exited child was not reported")
	}
	if !strings.Contains(err.Error(), "validator-stub process exited unexpectedly") ||
		!strings.Contains(err.Error(), "bind: address already in use") {
		t.Fatalf("report lacks name or log tail: %v", err)
	}
	if stopErr := validator.stop(true); stopErr == nil {
		t.Fatal("stop(true) did not surface the non-zero exit")
	}
	if validator.exit != nil || validator.exitedError() != nil {
		t.Fatal("exit state survived stop")
	}
}
