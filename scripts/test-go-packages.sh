#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(git -C "$(dirname "${BASH_SOURCE[0]}")/.." rev-parse --show-toplevel)
SELECTOR="$ROOT_DIR/scripts/go-packages.sh"
PROBE_DIR="$ROOT_DIR/client-web/node_modules/.truerepublic-go-selector-probe"
PROBE_FILE="$PROBE_DIR/probe.go"

cleanup() {
  rm -f "$PROBE_FILE"
  rmdir "$PROBE_DIR" 2>/dev/null || true
}
trap cleanup EXIT HUP INT TERM

before=$("$SELECTOR" --list)

for required in . ./token ./treasury/keeper ./x/dex ./x/truedemocracy; do
  if ! grep -Fxq "$required" <<<"$before"; then
    echo "required Go package directory missing: $required" >&2
    exit 1
  fi
done

if grep -Fq '/node_modules/' <<<"$before"; then
  echo "selector included a node_modules package" >&2
  exit 1
fi

mkdir -p "$PROBE_DIR"
printf 'package ignoredprobe\n' >"$PROBE_FILE"

after=$("$SELECTOR" --list)
if [[ "$before" != "$after" ]]; then
  echo "ignored node_modules source changed repository Go package selection" >&2
  diff -u <(printf '%s\n' "$before") <(printf '%s\n' "$after") || true
  exit 1
fi

printf 'Go package selector excludes ignored dependency trees:\n%s\n' "$after"
