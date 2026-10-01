#!/usr/bin/env bash
# Prove that every quarantined CosmWasm prototype crate (issue #308) fails a
# real wasm32 compilation with the QUARANTINED marker. Any successful build,
# a failure without the marker, or a missing package fails this verifier.
#
# Requires the wasm32-unknown-unknown target of the pinned toolchain. Build
# output goes to a private temporary directory that is removed on exit; shared
# Cargo caches are never touched.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONTRACTS_DIR="$ROOT_DIR/contracts"
TARGET="wasm32-unknown-unknown"
MARKER="QUARANTINED (TrueRepublic #308)"
PACKAGES=(
  truerepublic-contracts
  governance-dao
  zkp-aggregator
  dex-bot
  token-vesting
)

WORK_DIR="$(mktemp -d "${TMPDIR:-/tmp}/truerepublic-quarantine-wasm.XXXXXX")"
cleanup() {
  rm -rf -- "$WORK_DIR"
}
trap cleanup EXIT INT TERM

if ! rustup target list --installed 2>/dev/null | grep -qx "$TARGET"; then
  echo "FAIL $TARGET target is not installed for the active toolchain" >&2
  exit 1
fi

failures=0
for package in "${PACKAGES[@]}"; do
  log="$WORK_DIR/$package.log"
  if (cd "$CONTRACTS_DIR" && cargo check --locked --package "$package" \
    --target "$TARGET" --target-dir "$WORK_DIR/target") >"$log" 2>&1; then
    echo "FAIL $package compiled for $TARGET; the quarantine guard is not effective" >&2
    failures=$((failures + 1))
  elif grep -Fq "$MARKER" "$log"; then
    echo "OK   $package rejects $TARGET with the quarantine marker"
  else
    echo "FAIL $package failed for $TARGET without the quarantine marker:" >&2
    tail -n 20 "$log" >&2
    failures=$((failures + 1))
  fi
done

if [ "$failures" -ne 0 ]; then
  echo "FAIL $failures of ${#PACKAGES[@]} quarantined packages are not fail-closed for $TARGET" >&2
  exit 1
fi
echo "PASSED: all ${#PACKAGES[@]} quarantined packages are fail-closed for $TARGET"
