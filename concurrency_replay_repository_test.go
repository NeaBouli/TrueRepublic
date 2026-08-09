package main

import (
	"os"
	"strings"
	"testing"
)

func TestConcurrencyReplayRepositoryContract(t *testing.T) {
	harness := readRepositoryFile(t, "concurrency_replay_harness_test.go")
	workflow := readRepositoryFile(t, ".github/workflows/go-ci.yml")
	makefile := readRepositoryFile(t, "Makefile")
	roadmap := readRepositoryFile(t, "docs/ROLLOUT_ROADMAP.md")

	for _, required := range []string{
		`TRUEREPUBLIC_CONCURRENCY_REPLAY_SMOKE`,
		`TestConcurrentSharedStateReplayRestart`,
		`broadcastSmokeTxBytes(ctx, validator, signedTxBytes, signedTxHash)`,
		`validator.stop(true)`,
		`smokeAppHash(t, validator, preStopHeight)`,
		`strings.Contains(postRestart.evidence, "account sequence mismatch")`,
		`validateLedgerGenesis`,
		`initGenesisApp`,
	} {
		if !strings.Contains(harness, required) {
			t.Fatalf("concurrency/replay harness missing %q", required)
		}
	}

	for _, required := range []string{
		"concurrency-replay:",
		"TRUEREPUBLIC_CONCURRENCY_REPLAY_SMOKE=1 go test .",
		"-timeout=900s -v",
	} {
		if !strings.Contains(makefile, required) {
			t.Fatalf("Makefile concurrency/replay gate missing %q", required)
		}
	}

	for _, required := range []string{
		"concurrency-replay:",
		"timeout-minutes: 15",
		"run: make concurrency-replay",
	} {
		if !strings.Contains(workflow, required) {
			t.Fatalf("Go CI concurrency/replay gate missing %q", required)
		}
	}

	if !strings.Contains(roadmap, "GH-172") {
		t.Fatal("Phase 5 roadmap must link the verified GH-172 evidence")
	}

	info, err := os.Stat("concurrency_replay_harness_test.go")
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o022 != 0 {
		t.Fatalf("concurrency/replay harness must not be group/world writable: %o", info.Mode().Perm())
	}
}
