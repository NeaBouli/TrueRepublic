#!/usr/bin/env bash
# Verify deterministic fail-closed CI/release tool-bootstrap composition
# evidence (GH-278) offline against built artifacts, the tool-platform
# contract, the security gate contract, and the repository-owned tool locks.
set -euo pipefail
export LC_ALL=C LANG=C

ROOT_DIR=$(git -C "$(dirname "${BASH_SOURCE[0]}")/.." rev-parse --show-toplevel)
usage() { echo "usage: $0 --evidence <dir> --artifacts <dir> [--contract <file>] [--gates <file>] [--locks-root <dir>] [--output text|json]" >&2; }
EVIDENCE=
ARTIFACTS=
CONTRACT="$ROOT_DIR/configs/release/tool-platform.json"
GATES="$ROOT_DIR/configs/security/gates.json"
LOCKS_ROOT="$ROOT_DIR"
OUTPUT=text
while [[ $# -gt 0 ]]; do
  case "$1" in
    --evidence) [[ $# -ge 2 ]] || { usage; exit 2; }; EVIDENCE=$2; shift 2 ;;
    --artifacts) [[ $# -ge 2 ]] || { usage; exit 2; }; ARTIFACTS=$2; shift 2 ;;
    --contract) [[ $# -ge 2 ]] || { usage; exit 2; }; CONTRACT=$2; shift 2 ;;
    --gates) [[ $# -ge 2 ]] || { usage; exit 2; }; GATES=$2; shift 2 ;;
    --locks-root) [[ $# -ge 2 ]] || { usage; exit 2; }; LOCKS_ROOT=$2; shift 2 ;;
    --output) [[ $# -ge 2 ]] || { usage; exit 2; }; OUTPUT=$2; shift 2 ;;
    *) usage; exit 2 ;;
  esac
done
[[ -n "$EVIDENCE" && -n "$ARTIFACTS" && -n "$CONTRACT" && -n "$GATES" && -n "$LOCKS_ROOT" ]] || { usage; exit 2; }
[[ "$OUTPUT" == text || "$OUTPUT" == json ]] || { usage; exit 2; }
cd "$ROOT_DIR"
export GOTOOLCHAIN=local
export GOPROXY=off
export GOFLAGS=-mod=readonly
exec go run ./cmd/tool-bootstrap-evidence verify \
  --evidence "$EVIDENCE" --artifacts "$ARTIFACTS" \
  --contract "$CONTRACT" --gates "$GATES" --locks-root "$LOCKS_ROOT" \
  --output "$OUTPUT"
