#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
EXPECTED_GO_VERSION=go1.26.6
ACTUAL_GO_VERSION=$(go env GOVERSION)

if [[ "$ACTUAL_GO_VERSION" != "$EXPECTED_GO_VERSION" ]]; then
  echo "Go toolchain mismatch: got $ACTUAL_GO_VERSION, want $EXPECTED_GO_VERSION" >&2
  exit 1
fi

WORK_DIR=$(mktemp -d "${TMPDIR:-/tmp}/truerepublic-zkp-wasm.XXXXXX")
cleanup() {
  rm -rf "$WORK_DIR"
}
trap cleanup EXIT

WASM_PATH="$WORK_DIR/zkp-prover.wasm"
RESULT_PATH="$WORK_DIR/result.json"
WASM_EXEC_PATH=$(go env GOROOT)/lib/wasm/wasm_exec.js

cd "$ROOT_DIR"
DEPENDENCIES=$(GOOS=js GOARCH=wasm CGO_ENABLED=0 go list -deps ./cmd/zkp-prover-wasm)
if grep -Eiq 'cosmos|cometbft|pebble|wasmd|cosmos-db' <<<"$DEPENDENCIES"; then
  echo "test-only WASM prover unexpectedly imports chain/runtime dependencies" >&2
  grep -Ei 'cosmos|cometbft|pebble|wasmd|cosmos-db' <<<"$DEPENDENCIES" >&2
  exit 1
fi

GOOS=js GOARCH=wasm CGO_ENABLED=0 go build \
  -trimpath -buildvcs=false -ldflags='-buildid=' \
  -o "$WASM_PATH" ./cmd/zkp-prover-wasm

cd "$ROOT_DIR/client-web"
TRUEREPUBLIC_ZKP_WASM_INTEGRATION=1 \
TRUEREPUBLIC_ZKP_WASM_PATH="$WASM_PATH" \
TRUEREPUBLIC_WASM_EXEC_PATH="$WASM_EXEC_PATH" \
TRUEREPUBLIC_ZKP_RESULT_PATH="$RESULT_PATH" \
  ./node_modules/.bin/vitest run src/services/zkpWasmProver.integration.test.ts

if [[ ! -s "$RESULT_PATH" ]]; then
  echo "maintained-client integration produced no keeper handoff at $RESULT_PATH" >&2
  exit 1
fi

cd "$ROOT_DIR"
TRUEREPUBLIC_ZKP_RESULT_PATH="$RESULT_PATH" \
  go test ./internal/zkpprover \
    -run '^TestWASMClientOutputIsAcceptedByNativeVerifier$' \
    -count=1 -timeout=60s -v

# GH-266: the fresh client proof must also traverse the real keeper reward
# boundary. The env var is set above, so these tests run rather than skip;
# a missing or malformed handoff fails the gate.
TRUEREPUBLIC_ZKP_RESULT_PATH="$RESULT_PATH" \
  go test ./x/truedemocracy \
    -run '^TestWASMClient' \
    -count=1 -timeout=300s -v
