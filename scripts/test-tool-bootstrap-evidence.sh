#!/usr/bin/env bash
# Offline fail-closed contract test for the GH-278 tool-bootstrap evidence:
# package tests, positive/adversarial fixtures, deterministic regeneration,
# and generator failure semantics. Builds no real tools and needs no network.
set -euo pipefail
export LC_ALL=C LANG=C

ROOT_DIR=$(git -C "$(dirname "${BASH_SOURCE[0]}")/.." rev-parse --show-toplevel)
cd "$ROOT_DIR"
export GOTOOLCHAIN=local
export GOPROXY=off
export GOFLAGS=-mod=readonly

bash -n scripts/build-ci-tool.sh
bash -n scripts/generate-tool-bootstrap-evidence.sh
bash -n scripts/verify-tool-bootstrap-evidence.sh
bash -n scripts/test-tool-bootstrap-evidence.sh

expect_usage_exit_2() {
  local script=$1 option=$2 status
  set +e
  "$script" "$option" >/dev/null 2>&1
  status=$?
  set -e
  [[ "$status" -eq 2 ]] || {
    echo "$script returned $status instead of usage exit 2 for $option without a value" >&2
    exit 1
  }
}
expect_usage_exit_2 ./scripts/build-ci-tool.sh --tool
expect_usage_exit_2 ./scripts/build-ci-tool.sh --output-dir
for option in --artifacts-a --artifacts-b --output-dir --contract --gates --locks-root; do
  expect_usage_exit_2 ./scripts/generate-tool-bootstrap-evidence.sh "$option"
done

go test ./toolbootstrapevidence ./cmd/tool-bootstrap-evidence -count=1

FIXTURES="$ROOT_DIR/testdata/toolbootstrapevidence"

verify_fixture() {
  ./scripts/verify-tool-bootstrap-evidence.sh \
    --evidence "$1/evidence" --artifacts "$1/artifacts" \
    --contract "$1/contract.json" --gates "$1/gates.json" --locks-root "$1" "${@:2}"
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
  echo "tool-bootstrap verifier accepted true status claims" >&2
  exit 1
fi
if verify_fixture "$FIXTURES/valid" --expected-extra value >/dev/null 2>&1; then
  echo "tool-bootstrap verifier accepted an unknown flag" >&2
  exit 1
fi
if ./scripts/verify-tool-bootstrap-evidence.sh --evidence "$FIXTURES/valid/evidence" >/dev/null 2>&1; then
  echo "tool-bootstrap verifier accepted missing arguments" >&2
  exit 1
fi
if verify_fixture "$ROOT_DIR/does-not-exist" >/dev/null 2>&1; then
  echo "tool-bootstrap verifier accepted a missing evidence directory" >&2
  exit 1
fi

# Generator round-trip with synthetic artifacts against the fixture contract.
make_artifacts() { # <dir> [binary-content-suffix]
  local dir=$1 suffix=${2:-}
  local tool
  for tool in cyclonedx-gomod govulncheck staticcheck gitleaks; do
    mkdir -p "$dir/$tool"
    printf 'synthetic %s binary%s\n' "$tool" "$suffix" >"$dir/$tool/$tool"
  done
  mkdir -p "$dir/cyclonedx-npm/node_modules/@cyclonedx/cyclonedx-npm"
  printf '{"name":"@cyclonedx/cyclonedx-npm","version":"6.0.1"}\n' \
    >"$dir/cyclonedx-npm/node_modules/@cyclonedx/cyclonedx-npm/package.json"
}

generate() {
  ./scripts/generate-tool-bootstrap-evidence.sh \
    --artifacts-a "$tmp/artifacts-a" --artifacts-b "$tmp/artifacts-b" \
    --output-dir "$1" \
    --contract "$FIXTURES/valid/contract.json" --gates "$FIXTURES/valid/gates.json" \
    --locks-root "$FIXTURES/valid" "${@:2}"
}

make_artifacts "$tmp/artifacts-a"
make_artifacts "$tmp/artifacts-b"
generate "$tmp/evidence-a" >/dev/null
./scripts/verify-tool-bootstrap-evidence.sh \
  --evidence "$tmp/evidence-a" --artifacts "$tmp/artifacts-b" \
  --contract "$FIXTURES/valid/contract.json" --gates "$FIXTURES/valid/gates.json" \
  --locks-root "$FIXTURES/valid" >/dev/null
generate "$tmp/evidence-b" >/dev/null
cmp "$tmp/evidence-a/tool-bootstrap-evidence.json" "$tmp/evidence-b/tool-bootstrap-evidence.json"

expect_generator_failure() {
  local name=$1 out="$tmp/fail-$1"
  shift
  if generate "$out" "$@" >/dev/null 2>&1; then
    echo "tool-bootstrap generator accepted $name" >&2
    exit 1
  fi
  if [[ -e "$out" ]]; then
    echo "tool-bootstrap generator left a complete-looking bundle after $name" >&2
    exit 1
  fi
}

mkdir "$tmp/already-exists"
printf 'preserve\n' >"$tmp/already-exists/sentinel"
if generate "$tmp/unused-existing-output" --output-dir "$tmp/already-exists" >/dev/null 2>&1; then
  echo "tool-bootstrap generator accepted an existing output directory" >&2
  exit 1
fi
[[ -f "$tmp/already-exists/sentinel" && $(find "$tmp/already-exists" -mindepth 1 -maxdepth 1 | wc -l | tr -d ' ') -eq 1 ]] || {
  echo "tool-bootstrap generator mutated the effective existing output directory" >&2
  exit 1
}
[[ ! -e "$tmp/unused-existing-output" ]] || {
  echo "tool-bootstrap generator created the shadowed output directory" >&2
  exit 1
}
expect_generator_failure "a missing artifacts directory" --artifacts-a "$tmp/does-not-exist"
expect_generator_failure "duplicate artifacts directories" --artifacts-b "$tmp/artifacts-a"
expect_generator_failure "aliased duplicate artifacts directories" --artifacts-b "$tmp/artifacts-a/."

printf 'drifted\n' >>"$tmp/artifacts-b/gitleaks/gitleaks"
expect_generator_failure "an artifact digest drift between builds"
make_artifacts "$tmp/artifacts-b"

rm "$tmp/artifacts-b/staticcheck/staticcheck"
expect_generator_failure "a missing tool artifact"
make_artifacts "$tmp/artifacts-b"

rm "$tmp/artifacts-b/govulncheck/govulncheck"
ln -s "$tmp/artifacts-a/govulncheck/govulncheck" "$tmp/artifacts-b/govulncheck/govulncheck"
expect_generator_failure "a symlinked tool artifact"
make_artifacts "$tmp/artifacts-b"

rm "$tmp/artifacts-a/govulncheck/govulncheck"
ln -s "$tmp/artifacts-b/govulncheck/govulncheck" "$tmp/artifacts-a/govulncheck/govulncheck"
expect_generator_failure "a symlinked first-build tool artifact"
