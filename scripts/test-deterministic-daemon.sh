#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(git -C "$(dirname "${BASH_SOURCE[0]}")/.." rev-parse --show-toplevel)
source "$ROOT_DIR/scripts/verify-deterministic-daemon.sh"

TEMP_DIR=$(mktemp -d)
trap 'rm -rf "$TEMP_DIR"' EXIT

CONTRACT="$ROOT_DIR/configs/build/deterministic-linux-daemon.json"
SOURCE_REF=0123456789abcdef0123456789abcdef01234567
FIRST="$TEMP_DIR/first"
SECOND="$TEMP_DIR/second"
printf 'identical deterministic payload\n' >"$FIRST"
cp "$FIRST" "$SECOND"

inspect_binary() {
  local binary=$3
  if [[ ${VERSION_MISMATCH:-0} == 1 && "$binary" == "$SECOND" ]]; then
    printf 'truerepublicd version wrong-ref\n'
  else
    printf 'truerepublicd version %s\n' "$SOURCE_REF"
  fi
}

assert_fails() {
  local label=$1
  shift
  if ("$@") >/dev/null 2>&1; then
    echo "expected failure: $label" >&2
    exit 1
  fi
}

verify_pair "$CONTRACT" linux-amd64 "$SOURCE_REF" "$FIRST" "$SECOND" >/dev/null
validate_contract "$CONTRACT" linux-arm64 "$SOURCE_REF"

printf 'different payload\n' >"$SECOND"
assert_fails "hash mismatch" verify_pair "$CONTRACT" linux-amd64 "$SOURCE_REF" "$FIRST" "$SECOND"
cp "$FIRST" "$SECOND"

VERSION_MISMATCH=1
assert_fails "version mismatch" verify_pair "$CONTRACT" linux-amd64 "$SOURCE_REF" "$FIRST" "$SECOND"
unset VERSION_MISMATCH

BROKEN_CONTRACT="$TEMP_DIR/broken-contract.json"
jq '.go_version = "1.26.4"' "$CONTRACT" >"$BROKEN_CONTRACT"
assert_fails "contract mismatch" validate_contract "$BROKEN_CONTRACT" linux-amd64 "$SOURCE_REF"
assert_fails "source ref mismatch" validate_contract "$CONTRACT" linux-amd64 not-a-commit
assert_fails "target mismatch" validate_contract "$CONTRACT" linux-riscv64 "$SOURCE_REF"

echo "deterministic daemon contract tests passed (success, hash, version, contract, ref, target)"
