#!/usr/bin/env bash
set -euo pipefail
export LC_ALL=C LANG=C

ROOT_DIR=$(git -C "$(dirname "${BASH_SOURCE[0]}")/.." rev-parse --show-toplevel)
cd "$ROOT_DIR"
export GOTOOLCHAIN=local
export GOPROXY=off
export GOFLAGS=-mod=readonly

bash -n scripts/generate-cross-run-evidence.sh
bash -n scripts/verify-cross-run-evidence.sh
bash -n scripts/test-cross-run-evidence.sh

go test ./crossrunevidence ./cmd/cross-run-evidence -count=1

FIXTURES="$ROOT_DIR/testdata/crossrunevidence"
REPOSITORY=NeaBouli/TrueRepublic
WORKFLOW_PATH=.github/workflows/reproducible-daemon.yml
BRANCH=main
COMMIT=0123456789abcdef0123456789abcdef01234567
BASELINE_RUN_ID=33400000001
CURRENT_RUN_ID=33400000002

verify_fixture() {
  ./scripts/verify-cross-run-evidence.sh \
    --evidence "$1" \
    --repository "$REPOSITORY" --workflow-path "$WORKFLOW_PATH" --branch "$BRANCH" --commit "$COMMIT" \
    --baseline-run-id "$BASELINE_RUN_ID" --current-run-id "$CURRENT_RUN_ID" "${@:2}"
}

verify_fixture "$FIXTURES/valid" >/dev/null

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

first="$tmp/first.json"
second="$tmp/second.json"
verify_fixture "$FIXTURES/valid" --output json >"$first"
verify_fixture "$FIXTURES/valid" --output json >"$second"
cmp "$first" "$second"

if verify_fixture "$FIXTURES/invalid-claims" >/dev/null 2>&1; then
  echo "cross-run verifier accepted true status claims" >&2
  exit 1
fi
if verify_fixture "$FIXTURES/valid" --expected-extra value >/dev/null 2>&1; then
  echo "cross-run verifier accepted an unknown flag" >&2
  exit 1
fi
if ./scripts/verify-cross-run-evidence.sh --evidence "$FIXTURES/valid" >/dev/null 2>&1; then
  echo "cross-run verifier accepted missing expected values" >&2
  exit 1
fi
if verify_fixture "$ROOT_DIR/does-not-exist" >/dev/null 2>&1; then
  echo "cross-run verifier accepted a missing evidence directory" >&2
  exit 1
fi

make_inputs() {
  local prefix=$1
  mkdir -p "$tmp/in-$prefix-baseline" "$tmp/in-$prefix-current"
  cp "$FIXTURES/valid/candidate-manifest-baseline.json" "$tmp/in-$prefix-baseline/candidate-evidence.json"
  cp "$FIXTURES/valid/candidate-report-baseline.json" "$tmp/in-$prefix-baseline/candidate-evidence-report.json"
  cp "$FIXTURES/valid/candidate-manifest-current.json" "$tmp/in-$prefix-current/candidate-evidence.json"
  cp "$FIXTURES/valid/candidate-report-current.json" "$tmp/in-$prefix-current/candidate-evidence-report.json"
  cp "$FIXTURES/valid/receipt-baseline.json" "$tmp/receipt-$prefix-baseline.json"
  cp "$FIXTURES/valid/receipt-current.json" "$tmp/receipt-$prefix-current.json"
}

generate() {
  ./scripts/generate-cross-run-evidence.sh \
    --baseline-receipt "$tmp/receipt-main-baseline.json" --current-receipt "$tmp/receipt-main-current.json" \
    --baseline-dir "$tmp/in-main-baseline" --current-dir "$tmp/in-main-current" \
    --output-dir "$1" \
    --repository "$REPOSITORY" --workflow-path "$WORKFLOW_PATH" --branch "$BRANCH" --commit "$COMMIT" \
    --baseline-run-id "$BASELINE_RUN_ID" --current-run-id "$CURRENT_RUN_ID" "${@:2}"
}

make_inputs main
generate "$tmp/cross-run-a" >/dev/null
verify_fixture "$tmp/cross-run-a" >/dev/null
generate "$tmp/cross-run-b" >/dev/null
cmp "$tmp/cross-run-a/cross-run-evidence.json" "$tmp/cross-run-b/cross-run-evidence.json"

expect_generator_failure() {
  local name=$1 out="$tmp/fail-$1"
  shift
  if generate "$out" "$@" >/dev/null 2>&1; then
    echo "cross-run generator accepted $name" >&2
    exit 1
  fi
  if [[ -e "$out" ]]; then
    echo "cross-run generator left a complete-looking bundle after $name" >&2
    exit 1
  fi
}

expect_generator_failure "a missing input directory" --baseline-dir "$tmp/does-not-exist"
expect_generator_failure "duplicate input directories" --current-dir "$tmp/in-main-baseline"
expect_generator_failure "aliased duplicate input directories" --current-dir "$tmp/in-main-baseline/."
mkdir "$tmp/already-exists"
expect_generator_failure "an existing output directory" --output-dir "$tmp/already-exists"
expect_generator_failure "identical run IDs" --current-run-id "$BASELINE_RUN_ID"
expect_generator_failure "a zero run ID" --baseline-run-id 0
expect_generator_failure "a malformed run ID" --baseline-run-id run-1
expect_generator_failure "a huge run ID" --baseline-run-id 1234567890123456789012345
expect_generator_failure "a malformed commit" --commit v0.5.0

make_inputs bad-receipt
jq '.run_id="33400000009"' "$tmp/receipt-bad-receipt-baseline.json" >"$tmp/mutated-receipt.json"
mv "$tmp/mutated-receipt.json" "$tmp/receipt-bad-receipt-baseline.json"
if ./scripts/generate-cross-run-evidence.sh \
  --baseline-receipt "$tmp/receipt-bad-receipt-baseline.json" --current-receipt "$tmp/receipt-bad-receipt-current.json" \
  --baseline-dir "$tmp/in-bad-receipt-baseline" --current-dir "$tmp/in-bad-receipt-current" \
  --output-dir "$tmp/fail-bad-receipt" \
  --repository "$REPOSITORY" --workflow-path "$WORKFLOW_PATH" --branch "$BRANCH" --commit "$COMMIT" \
  --baseline-run-id "$BASELINE_RUN_ID" --current-run-id "$CURRENT_RUN_ID" >/dev/null 2>&1; then
  echo "cross-run generator accepted a receipt run ID mismatch" >&2
  exit 1
fi

make_inputs extra
printf 'unexpected\n' >"$tmp/in-extra-baseline/extra.json"
expect_generator_failure "an unexpected extra input member" --baseline-dir "$tmp/in-extra-baseline"

make_inputs symlink
rm "$tmp/in-symlink-baseline/candidate-evidence.json"
ln -s "$tmp/in-main-baseline/candidate-evidence.json" "$tmp/in-symlink-baseline/candidate-evidence.json"
expect_generator_failure "a symlinked candidate input" --baseline-dir "$tmp/in-symlink-baseline"

make_inputs drift
jq '.binary_targets[0].artifact_sha256=("0"*64)' "$tmp/in-drift-current/candidate-evidence.json" >"$tmp/drifted.json"
mv "$tmp/drifted.json" "$tmp/in-drift-current/candidate-evidence.json"
expect_generator_failure "a binary digest drift between runs" --current-dir "$tmp/in-drift-current"
