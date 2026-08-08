package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestCriticalCoverageRepositoryContract(t *testing.T) {
	t.Parallel()

	contract, err := os.Open("configs/quality/critical-coverage.tsv")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := contract.Close(); err != nil {
			t.Errorf("close critical coverage contract: %v", err)
		}
	}()

	want := map[string]float64{
		".":                 71.8,
		"./x/dex":           51.0,
		"./x/truedemocracy": 63.7,
	}
	seen := make(map[string]bool, len(want))
	scanner := bufio.NewScanner(contract)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 3 {
			t.Fatalf("invalid critical coverage row %q", line)
		}
		minimum, err := strconv.ParseFloat(fields[2], 64)
		if err != nil {
			t.Fatalf("parse coverage minimum in %q: %v", line, err)
		}
		baseline, exists := want[fields[1]]
		if !exists || seen[fields[1]] || minimum < baseline || minimum > 100 {
			t.Fatalf("unsafe critical coverage row %q", line)
		}
		seen[fields[1]] = true
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan critical coverage contract: %v", err)
	}
	if len(seen) != len(want) {
		t.Fatalf("critical coverage package count = %d, want %d", len(seen), len(want))
	}

	assertFileContainsAll(t, "scripts/check-critical-coverage.sh", []string{
		"set -euo pipefail",
		"configs/quality/critical-coverage.tsv",
		"cd \"$ROOT_DIR\"",
		"go test -covermode=atomic",
		"go tool cover -func",
		"must define exactly three packages",
	})
	assertFileContainsAll(t, ".github/workflows/go-ci.yml", []string{
		"'configs/quality/**'",
		"'scripts/check-critical-coverage.sh'",
		"Enforce critical package coverage floors",
		"./scripts/check-critical-coverage.sh",
	})
}
