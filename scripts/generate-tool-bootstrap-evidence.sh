#!/usr/bin/env bash
# Generate deterministic fail-closed CI/release tool-bootstrap composition
# evidence (GH-278). Two independently built artifact trees (--artifacts-a and
# --artifacts-b) must bind the exact same SHA-256 per configured tool; the
# manifest additionally binds the exact configured versions, the tool-platform
# and security-gate contract digests, and the repository-owned Go/npm lock
# digests. All signed/published/deployed/production/long-term-hermetic claims
# are explicitly false. The generator retains JSON metadata only and writes a
# caller-provided NEW output directory.
set -euo pipefail
export LC_ALL=C LANG=C

ROOT_DIR=$(git -C "$(dirname "${BASH_SOURCE[0]}")/.." rev-parse --show-toplevel)
CONTRACT="$ROOT_DIR/configs/release/tool-platform.json"
GATES="$ROOT_DIR/configs/security/gates.json"
LOCKS_ROOT="$ROOT_DIR"
ARTIFACTS_A=
ARTIFACTS_B=
OUTPUT_DIR=
usage() { echo "usage: $0 --artifacts-a <dir> --artifacts-b <dir> --output-dir <new-dir> [--contract <file>] [--gates <file>] [--locks-root <dir>]" >&2; }
while [[ $# -gt 0 ]]; do
  case "$1" in
    --artifacts-a) [[ $# -ge 2 ]] || { usage; exit 2; }; ARTIFACTS_A=$2; shift 2;;
    --artifacts-b) [[ $# -ge 2 ]] || { usage; exit 2; }; ARTIFACTS_B=$2; shift 2;;
    --output-dir) [[ $# -ge 2 ]] || { usage; exit 2; }; OUTPUT_DIR=$2; shift 2;;
    --contract) [[ $# -ge 2 ]] || { usage; exit 2; }; CONTRACT=$2; shift 2;;
    --gates) [[ $# -ge 2 ]] || { usage; exit 2; }; GATES=$2; shift 2;;
    --locks-root) [[ $# -ge 2 ]] || { usage; exit 2; }; LOCKS_ROOT=$2; shift 2;;
    *) usage; exit 2;;
  esac
done
[[ -n "$ARTIFACTS_A" && -n "$ARTIFACTS_B" && -n "$OUTPUT_DIR" && -n "$CONTRACT" && -n "$GATES" && -n "$LOCKS_ROOT" ]] || { usage; exit 2; }
[[ ! -e "$OUTPUT_DIR" && ! -L "$OUTPUT_DIR" ]] || { echo "output directory already exists" >&2; exit 1; }
[[ -d "$(dirname "$OUTPUT_DIR")" ]] || { echo "output parent directory is unavailable" >&2; exit 1; }

sha256_file() { if command -v sha256sum >/dev/null; then sha256sum "$1"|awk '{print $1}'; else shasum -a 256 "$1"|awk '{print $1}'; fi; }

[[ $(jq -er '.schema' "$CONTRACT") == "truerepublic.release-tool-platform/v1" ]] || { echo "tool-platform contract schema mismatch" >&2; exit 1; }
[[ $(jq -er '.version' "$GATES") == "truerepublic.security-gates/v1" ]] || { echo "security gate contract schema mismatch" >&2; exit 1; }

canonical_inputs=()
for dir in "$ARTIFACTS_A" "$ARTIFACTS_B"; do
  [[ -d "$dir" && ! -L "$dir" ]] || { echo "artifacts directory $dir is unavailable or symlinked" >&2; exit 1; }
  canonical=$(cd "$dir" && pwd -P)
  for prior in "${canonical_inputs[@]-}"; do
    [[ -z "$prior" || "$canonical" != "$prior" ]] || { echo "artifacts directories must be distinct" >&2; exit 1; }
  done
  canonical_inputs+=("$canonical")
done

contract_hash=$(sha256_file "$CONTRACT")
gates_hash=$(sha256_file "$GATES")

lock_digest() { # <contract key> -> digest of the contract-declared lock path
  local rel
  rel=$(jq -er --arg key "$1" '.bootstrap[$key]' "$CONTRACT")
  [[ "$rel" != /* && "$rel" != *..* && "$rel" != *\\* ]] || { echo "lock path $rel is not repository-relative" >&2; exit 1; }
  local file="$LOCKS_ROOT/$rel"
  [[ -f "$file" && ! -L "$file" ]] || { echo "lock $file is missing or symlinked" >&2; exit 1; }
  sha256_file "$file"
}
go_module_hash=$(lock_digest go_module)
go_sum_hash=$(lock_digest go_sum)
npm_package_hash=$(lock_digest npm_package)
npm_lock_hash=$(lock_digest npm_lock)

mapfile -t tool_ids < <(jq -r '.bootstrap.tools[].id' "$CONTRACT")
[[ ${#tool_ids[@]} -ge 1 ]] || { echo "tool-platform contract declares no bootstrap tools" >&2; exit 1; }

tools_json='[]'
for id in "${tool_ids[@]}"; do
  kind=$(jq -er --arg id "$id" '.bootstrap.tools[] | select(.id == $id) | .kind' "$CONTRACT")
  gates_key=$(jq -er --arg id "$id" '.bootstrap.tools[] | select(.id == $id) | .gates_key' "$CONTRACT")
  artifact=$(jq -er --arg id "$id" '.bootstrap.tools[] | select(.id == $id) | .artifact' "$CONTRACT")
  version=$(jq -er --arg key "$gates_key" '.tools[$key]' "$GATES")
  [[ "$version" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo "security gate version for $gates_key is not exact" >&2; exit 1; }
  file_a="$ARTIFACTS_A/$id/$artifact"
  file_b="$ARTIFACTS_B/$id/$artifact"
  for file in "$file_a" "$file_b"; do
    [[ -f "$file" && ! -L "$file" ]] || { echo "artifact $file is missing or symlinked" >&2; exit 1; }
  done
  digest_a=$(sha256_file "$file_a")
  digest_b=$(sha256_file "$file_b")
  [[ "$digest_a" == "$digest_b" ]] || {
    echo "tool $id artifact digest differs between the two builds" >&2; exit 1; }
  tools_json=$(jq -cn -S \
    --argjson prior "$tools_json" \
    --arg id "$id" --arg kind "$kind" --arg version "$version" \
    --arg file "$id/$artifact" --arg digest "$digest_a" \
    '$prior + [{id:$id,kind:$kind,version:$version,artifact:{file:$file,sha256:$digest}}]')
done

complete=0
trap 'if [[ "$complete" != 1 && -d "$OUTPUT_DIR" ]]; then rm -rf "$OUTPUT_DIR"; fi' EXIT
mkdir "$OUTPUT_DIR"

jq -n -S \
  --arg contract_hash "$contract_hash" \
  --arg gates_hash "$gates_hash" \
  --arg go_module_rel "$(jq -er '.bootstrap.go_module' "$CONTRACT")" \
  --arg go_sum_rel "$(jq -er '.bootstrap.go_sum' "$CONTRACT")" \
  --arg npm_package_rel "$(jq -er '.bootstrap.npm_package' "$CONTRACT")" \
  --arg npm_lock_rel "$(jq -er '.bootstrap.npm_lock' "$CONTRACT")" \
  --arg go_module_hash "$go_module_hash" \
  --arg go_sum_hash "$go_sum_hash" \
  --arg npm_package_hash "$npm_package_hash" \
  --arg npm_lock_hash "$npm_lock_hash" \
  --argjson tools "$tools_json" \
  '{
    schema:"truerepublic.tool-bootstrap-evidence/v1",
    contract_sha256:$contract_hash,
    gates_sha256:$gates_hash,
    claims:{signed:false,published:false,deployed:false,production:false,long_term_hermetic:false},
    locks:{
      go_module:{file:$go_module_rel,sha256:$go_module_hash},
      go_sum:{file:$go_sum_rel,sha256:$go_sum_hash},
      npm_package:{file:$npm_package_rel,sha256:$npm_package_hash},
      npm_lock:{file:$npm_lock_rel,sha256:$npm_lock_hash}
    },
    tools:$tools
  }' >"$OUTPUT_DIR/tool-bootstrap-evidence.json"

"$ROOT_DIR/scripts/verify-tool-bootstrap-evidence.sh" \
  --evidence "$OUTPUT_DIR" --artifacts "$ARTIFACTS_A" \
  --contract "$CONTRACT" --gates "$GATES" --locks-root "$LOCKS_ROOT" >/dev/null
"$ROOT_DIR/scripts/verify-tool-bootstrap-evidence.sh" \
  --evidence "$OUTPUT_DIR" --artifacts "$ARTIFACTS_B" \
  --contract "$CONTRACT" --gates "$GATES" --locks-root "$LOCKS_ROOT"
complete=1
trap - EXIT
