#!/usr/bin/env bash
set -euo pipefail
export LC_ALL=C LANG=C

ROOT_DIR=$(git -C "$(dirname "${BASH_SOURCE[0]}")/.." rev-parse --show-toplevel)
usage() { echo "usage: $0 --evidence <dir> --repository <owner/repo> --workflow-path <path> --branch <branch> --commit <40-hex> --baseline-run-id <id> --current-run-id <id> [--contract <file>] [--candidate-contract <file>] [--output text|json]" >&2; }
EVIDENCE=
CONTRACT="$ROOT_DIR/configs/release/cross-run-rebuild.json"
CANDIDATE_CONTRACT="$ROOT_DIR/configs/release/candidate-evidence.json"
REPOSITORY=
WORKFLOW_PATH=
BRANCH=
COMMIT=
BASELINE_RUN_ID=
CURRENT_RUN_ID=
OUTPUT=text
while [[ $# -gt 0 ]]; do
  case "$1" in
    --evidence) [[ $# -ge 2 ]] || { usage; exit 2; }; EVIDENCE=$2; shift 2 ;;
    --contract) [[ $# -ge 2 ]] || { usage; exit 2; }; CONTRACT=$2; shift 2 ;;
    --candidate-contract) [[ $# -ge 2 ]] || { usage; exit 2; }; CANDIDATE_CONTRACT=$2; shift 2 ;;
    --repository) [[ $# -ge 2 ]] || { usage; exit 2; }; REPOSITORY=$2; shift 2 ;;
    --workflow-path) [[ $# -ge 2 ]] || { usage; exit 2; }; WORKFLOW_PATH=$2; shift 2 ;;
    --branch) [[ $# -ge 2 ]] || { usage; exit 2; }; BRANCH=$2; shift 2 ;;
    --commit) [[ $# -ge 2 ]] || { usage; exit 2; }; COMMIT=$2; shift 2 ;;
    --baseline-run-id) [[ $# -ge 2 ]] || { usage; exit 2; }; BASELINE_RUN_ID=$2; shift 2 ;;
    --current-run-id) [[ $# -ge 2 ]] || { usage; exit 2; }; CURRENT_RUN_ID=$2; shift 2 ;;
    --output) [[ $# -ge 2 ]] || { usage; exit 2; }; OUTPUT=$2; shift 2 ;;
    *) usage; exit 2 ;;
  esac
done
[[ -n "$EVIDENCE" && -n "$CONTRACT" && -n "$CANDIDATE_CONTRACT" && -n "$REPOSITORY" && -n "$WORKFLOW_PATH" && -n "$BRANCH" && -n "$COMMIT" && -n "$BASELINE_RUN_ID" && -n "$CURRENT_RUN_ID" ]] || { usage; exit 2; }
[[ "$OUTPUT" == text || "$OUTPUT" == json ]] || { usage; exit 2; }
cd "$ROOT_DIR"
export GOTOOLCHAIN=local
export GOPROXY=off
export GOFLAGS=-mod=readonly
exec go run ./cmd/cross-run-evidence compare \
  --evidence "$EVIDENCE" --contract "$CONTRACT" --candidate-contract "$CANDIDATE_CONTRACT" \
  --expected-repository "$REPOSITORY" --expected-workflow "$WORKFLOW_PATH" --expected-branch "$BRANCH" \
  --expected-commit "$COMMIT" --expected-baseline-run-id "$BASELINE_RUN_ID" --expected-current-run-id "$CURRENT_RUN_ID" \
  --output "$OUTPUT"
