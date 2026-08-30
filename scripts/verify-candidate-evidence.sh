#!/usr/bin/env bash
set -euo pipefail
export LC_ALL=C LANG=C

ROOT_DIR=$(git -C "$(dirname "${BASH_SOURCE[0]}")/.." rev-parse --show-toplevel)
usage() { echo "usage: $0 --evidence <dir> [--contract <file>] [--output text|json]" >&2; }
EVIDENCE=
CONTRACT="$ROOT_DIR/configs/release/candidate-evidence.json"
OUTPUT=text
while [[ $# -gt 0 ]]; do
  case "$1" in
    --evidence) [[ $# -ge 2 ]] || { usage; exit 2; }; EVIDENCE=$2; shift 2 ;;
    --contract) [[ $# -ge 2 ]] || { usage; exit 2; }; CONTRACT=$2; shift 2 ;;
    --output) [[ $# -ge 2 ]] || { usage; exit 2; }; OUTPUT=$2; shift 2 ;;
    *) usage; exit 2 ;;
  esac
done
[[ -n "$EVIDENCE" ]] || { usage; exit 2; }
[[ -n "$CONTRACT" ]] || { usage; exit 2; }
[[ "$OUTPUT" == text || "$OUTPUT" == json ]] || { usage; exit 2; }
cd "$ROOT_DIR"
export GOTOOLCHAIN=local
export GOPROXY=off
export GOFLAGS=-mod=readonly
exec go run ./cmd/candidate-evidence verify --evidence "$EVIDENCE" --contract "$CONTRACT" --output "$OUTPUT"
