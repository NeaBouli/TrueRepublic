#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(git -C "$(dirname "${BASH_SOURCE[0]}")/.." rev-parse --show-toplevel)
FUZZ_ITERATIONS=${TRUEREPUBLIC_FUZZ_ITERATIONS:-10000}

if [[ ! "$FUZZ_ITERATIONS" =~ ^[1-9][0-9]*$ ]] || ((FUZZ_ITERATIONS > 60000)); then
  echo "TRUEREPUBLIC_FUZZ_ITERATIONS must be an integer from 1 through 60000" >&2
  exit 2
fi

cd "$ROOT_DIR"

CGO_ENABLED=1 go test -race ./x/dex ./token ./x/truedemocracy \
  -run '^(TestComputeSwapOutputDeterministicProperties|TestPNYXCapBoundaryProperties|TestCanonicalSupplyPrefersExplicitOverBalances|TestValidateGenesisSupplyRejectsMalformedStructures|TestPNYXCapGenerativeSweep|TestConsensusSlashingReplayProperty)$' \
  -count=1 -timeout=180s

CGO_ENABLED=1 go test ./x/dex -run '^$' \
  -fuzz '^FuzzComputeSwapOutput$' -fuzztime="${FUZZ_ITERATIONS}x" \
  -parallel=1 -timeout=180s

CGO_ENABLED=1 go test ./token -run '^$' \
  -fuzz '^FuzzPNYXCapAndGenesisValidation$' -fuzztime="${FUZZ_ITERATIONS}x" \
  -parallel=1 -timeout=180s
