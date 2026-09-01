#!/usr/bin/env bash
set -euo pipefail
export LC_ALL=C LANG=C

ROOT_DIR=$(git -C "$(dirname "${BASH_SOURCE[0]}")/.." rev-parse --show-toplevel)
CONTRACT="$ROOT_DIR/configs/release/cross-run-rebuild.json"
BASELINE_RECEIPT=
CURRENT_RECEIPT=
BASELINE_DIR=
CURRENT_DIR=
OUTPUT_DIR=
REPOSITORY=
WORKFLOW_PATH=
BRANCH=
COMMIT=
BASELINE_RUN_ID=
CURRENT_RUN_ID=
usage() { echo "usage: $0 --baseline-receipt <file> --current-receipt <file> --baseline-dir <dir> --current-dir <dir> --output-dir <new-dir> --repository <owner/repo> --workflow-path <path> --branch <branch> --commit <40-hex> --baseline-run-id <id> --current-run-id <id> [--contract <file>]" >&2; }
while [[ $# -gt 0 ]]; do
  case "$1" in
    --baseline-receipt) BASELINE_RECEIPT=${2-}; shift 2;; --current-receipt) CURRENT_RECEIPT=${2-}; shift 2;;
    --baseline-dir) BASELINE_DIR=${2-}; shift 2;; --current-dir) CURRENT_DIR=${2-}; shift 2;;
    --output-dir) OUTPUT_DIR=${2-}; shift 2;; --contract) CONTRACT=${2-}; shift 2;;
    --repository) REPOSITORY=${2-}; shift 2;; --workflow-path) WORKFLOW_PATH=${2-}; shift 2;;
    --branch) BRANCH=${2-}; shift 2;; --commit) COMMIT=${2-}; shift 2;;
    --baseline-run-id) BASELINE_RUN_ID=${2-}; shift 2;; --current-run-id) CURRENT_RUN_ID=${2-}; shift 2;;
    *) usage; exit 2;;
  esac
done
[[ -n "$BASELINE_RECEIPT" && -n "$CURRENT_RECEIPT" && -n "$BASELINE_DIR" && -n "$CURRENT_DIR" && -n "$OUTPUT_DIR" && -n "$REPOSITORY" && -n "$WORKFLOW_PATH" && -n "$BRANCH" && -n "$COMMIT" && -n "$BASELINE_RUN_ID" && -n "$CURRENT_RUN_ID" ]] || { usage; exit 2; }
[[ "$COMMIT" =~ ^[0-9a-f]{40}$ ]] || { echo "commit must be exactly 40 lowercase hexadecimal characters" >&2; exit 1; }
[[ "$BASELINE_RUN_ID" =~ ^[1-9][0-9]{0,18}$ && "$CURRENT_RUN_ID" =~ ^[1-9][0-9]{0,18}$ ]] || { echo "run IDs must be decimal strings of 1-19 digits without leading zeros" >&2; exit 1; }
[[ "$BASELINE_RUN_ID" != "$CURRENT_RUN_ID" ]] || { echo "baseline and current run IDs must be distinct" >&2; exit 1; }
[[ ! -e "$OUTPUT_DIR" && ! -L "$OUTPUT_DIR" ]] || { echo "output directory already exists" >&2; exit 1; }
[[ -d "$(dirname "$OUTPUT_DIR")" ]] || { echo "output parent directory is unavailable" >&2; exit 1; }

sha256_file() { if command -v sha256sum >/dev/null; then sha256sum "$1"|awk '{print $1}'; else shasum -a 256 "$1"|awk '{print $1}'; fi; }

contract_hash=$(sha256_file "$CONTRACT")
[[ $(jq -er '.schema' "$CONTRACT") == "truerepublic.cross-run-rebuild/v1" ]] || { echo "cross-run contract schema mismatch" >&2; exit 1; }

canonical_inputs=()
for dir in "$BASELINE_DIR" "$CURRENT_DIR"; do
  [[ -d "$dir" && ! -L "$dir" ]] || { echo "input directory $dir is unavailable or symlinked" >&2; exit 1; }
  canonical=$(cd "$dir" && pwd -P)
  for prior in "${canonical_inputs[@]-}"; do
    [[ -z "$prior" || "$canonical" != "$prior" ]] || { echo "input directories must be distinct" >&2; exit 1; }
  done
  canonical_inputs+=("$canonical")
done

require_dir() {
  local dir=$1; shift
  [[ $(find "$dir" -mindepth 1 -maxdepth 1 | wc -l | tr -d ' ') -eq $# ]] || { echo "input directory $dir has an unexpected member count" >&2; exit 1; }
  local name
  for name in "$@"; do
    [[ -f "$dir/$name" && ! -L "$dir/$name" ]] || { echo "input $dir/$name is missing or symlinked" >&2; exit 1; }
  done
}
require_json() { [[ -f "$1" && ! -L "$1" ]] || { echo "input $1 is missing or symlinked" >&2; exit 1; }; [[ $(wc -c <"$1" | tr -d ' ') -le 1048576 ]] || { echo "input $1 exceeds the JSON byte limit" >&2; exit 1; }; }

require_dir "$BASELINE_DIR" candidate-evidence.json candidate-evidence-report.json
require_dir "$CURRENT_DIR" candidate-evidence.json candidate-evidence-report.json
require_json "$BASELINE_RECEIPT"
require_json "$CURRENT_RECEIPT"
require_json "$BASELINE_DIR/candidate-evidence.json"
require_json "$BASELINE_DIR/candidate-evidence-report.json"
require_json "$CURRENT_DIR/candidate-evidence.json"
require_json "$CURRENT_DIR/candidate-evidence-report.json"

check_receipt() {
  local file=$1 run_id=$2
  jq -e --arg run_id "$run_id" --arg commit "$COMMIT" --arg repository "$REPOSITORY" --arg workflow_path "$WORKFLOW_PATH" --arg branch "$BRANCH" '
    .schema=="truerepublic.cross-run-receipt/v1" and
    .repository==$repository and
    .workflow_path==$workflow_path and
    .branch==$branch and
    .event=="workflow_dispatch" and
    .run_id==$run_id and
    .head_sha==$commit
  ' "$file" >/dev/null || { echo "receipt $file does not match the expected run identity" >&2; exit 1; }
}
check_receipt "$BASELINE_RECEIPT" "$BASELINE_RUN_ID"
check_receipt "$CURRENT_RECEIPT" "$CURRENT_RUN_ID"

complete=0
trap 'if [[ "$complete" != 1 && -d "$OUTPUT_DIR" ]]; then rm -rf "$OUTPUT_DIR"; fi' EXIT
mkdir "$OUTPUT_DIR"

cp "$BASELINE_RECEIPT" "$OUTPUT_DIR/receipt-baseline.json"
cp "$CURRENT_RECEIPT" "$OUTPUT_DIR/receipt-current.json"
cp "$BASELINE_DIR/candidate-evidence.json" "$OUTPUT_DIR/candidate-manifest-baseline.json"
cp "$BASELINE_DIR/candidate-evidence-report.json" "$OUTPUT_DIR/candidate-report-baseline.json"
cp "$CURRENT_DIR/candidate-evidence.json" "$OUTPUT_DIR/candidate-manifest-current.json"
cp "$CURRENT_DIR/candidate-evidence-report.json" "$OUTPUT_DIR/candidate-report-current.json"

jq -n -S \
  --arg contract_hash "$contract_hash" \
  --arg repository "$REPOSITORY" --arg workflow_path "$WORKFLOW_PATH" --arg branch "$BRANCH" --arg commit "$COMMIT" \
  --arg baseline_run_id "$BASELINE_RUN_ID" --arg current_run_id "$CURRENT_RUN_ID" \
  --arg receipt_baseline "$(sha256_file "$OUTPUT_DIR/receipt-baseline.json")" \
  --arg receipt_current "$(sha256_file "$OUTPUT_DIR/receipt-current.json")" \
  --arg manifest_baseline "$(sha256_file "$OUTPUT_DIR/candidate-manifest-baseline.json")" \
  --arg manifest_current "$(sha256_file "$OUTPUT_DIR/candidate-manifest-current.json")" \
  --arg report_baseline "$(sha256_file "$OUTPUT_DIR/candidate-report-baseline.json")" \
  --arg report_current "$(sha256_file "$OUTPUT_DIR/candidate-report-current.json")" \
  '{
    schema:"truerepublic.cross-run-evidence/v1",
    contract_sha256:$contract_hash,
    comparison:{repository:$repository,workflow_path:$workflow_path,branch:$branch,commit:$commit,baseline_run_id:$baseline_run_id,current_run_id:$current_run_id},
    claims:{real_tag_created:false,ref_pushed:false,signed:false,attested:false,published:false,deployed:false,production:false,long_term_hermetic:false},
    baseline:{receipt:{file:"receipt-baseline.json",sha256:$receipt_baseline},candidate_manifest:{file:"candidate-manifest-baseline.json",sha256:$manifest_baseline},candidate_report:{file:"candidate-report-baseline.json",sha256:$report_baseline}},
    current:{receipt:{file:"receipt-current.json",sha256:$receipt_current},candidate_manifest:{file:"candidate-manifest-current.json",sha256:$manifest_current},candidate_report:{file:"candidate-report-current.json",sha256:$report_current}}
  }' >"$OUTPUT_DIR/cross-run-evidence.json"

"$ROOT_DIR/scripts/verify-cross-run-evidence.sh" \
  --evidence "$OUTPUT_DIR" --contract "$CONTRACT" \
  --repository "$REPOSITORY" --workflow-path "$WORKFLOW_PATH" --branch "$BRANCH" --commit "$COMMIT" \
  --baseline-run-id "$BASELINE_RUN_ID" --current-run-id "$CURRENT_RUN_ID"
complete=1
trap - EXIT
