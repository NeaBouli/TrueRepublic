#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(git -C "$(dirname "${BASH_SOURCE[0]}")/.." rev-parse --show-toplevel)
FUZZ_SECONDS=${TRUEREPUBLIC_FUZZ_SECONDS:-10}

if [[ ! "$FUZZ_SECONDS" =~ ^[1-9][0-9]*$ ]] || ((FUZZ_SECONDS > 60)); then
  echo "TRUEREPUBLIC_FUZZ_SECONDS must be an integer from 1 through 60" >&2
  exit 2
fi

cd "$ROOT_DIR"

CGO_ENABLED=1 go test -race ./x/dex ./token ./x/truedemocracy \
  -run '^(TestComputeSwapOutputDeterministicProperties|TestPNYXCapBoundaryProperties|TestCanonicalSupplyPrefersExplicitOverBalances|TestValidateGenesisSupplyRejectsMalformedStructures|TestPNYXCapGenerativeSweep|TestConsensusSlashingReplayProperty)$' \
  -count=1 -timeout=180s

CGO_ENABLED=1 go test ./x/dex -run '^$' \
  -fuzz '^FuzzComputeSwapOutput$' -fuzztime="${FUZZ_SECONDS}s" \
  -parallel=1 -timeout=180s

CGO_ENABLED=1 go test ./token -run '^$' \
  -fuzz '^FuzzPNYXCapAndGenesisValidation$' -fuzztime="${FUZZ_SECONDS}s" \
  -parallel=1 -timeout=180s
