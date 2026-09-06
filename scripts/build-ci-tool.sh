#!/usr/bin/env bash
# Build one repository-locked CI/release tool (GH-278) from the nested
# tools/ci module or npm package-lock into a caller-provided NEW output
# directory. Resolution is fail-closed: the tool id must be allowlisted in
# configs/release/tool-platform.json, the module/npm version must equal the
# exact pin in configs/security/gates.json, Go builds run go mod verify and
# go build -mod=readonly with deterministic flags, and npm installs run only
# npm ci from the repository-owned lock. The script reads no secrets and
# mutates no project state; it writes only inside the new output directory.
set -euo pipefail
export LC_ALL=C LANG=C

ROOT_DIR=$(git -C "$(dirname "${BASH_SOURCE[0]}")/.." rev-parse --show-toplevel)
CONTRACT="$ROOT_DIR/configs/release/tool-platform.json"
GATES="$ROOT_DIR/configs/security/gates.json"
TOOL=
OUTPUT_DIR=
usage() { echo "usage: $0 --tool <id> --output-dir <new-dir>" >&2; }
while [[ $# -gt 0 ]]; do
  case "$1" in
    --tool) [[ $# -ge 2 ]] || { usage; exit 2; }; TOOL=$2; shift 2;;
    --output-dir) [[ $# -ge 2 ]] || { usage; exit 2; }; OUTPUT_DIR=$2; shift 2;;
    *) usage; exit 2;;
  esac
done
[[ -n "$TOOL" && -n "$OUTPUT_DIR" ]] || { usage; exit 2; }
[[ "$TOOL" =~ ^[a-z0-9][a-z0-9-]{0,63}$ ]] || { echo "tool id is malformed" >&2; exit 2; }
output_parent=$(dirname "$OUTPUT_DIR")
output_name=$(basename "$OUTPUT_DIR")
[[ "$output_name" != "." && "$output_name" != ".." && -n "$output_name" ]] || {
  echo "output directory is malformed" >&2; exit 2; }
[[ -d "$output_parent" ]] || { echo "output parent directory is unavailable" >&2; exit 1; }
output_parent=$(cd "$output_parent" && pwd -P)
OUTPUT_DIR="$output_parent/$output_name"
[[ ! -e "$OUTPUT_DIR" && ! -L "$OUTPUT_DIR" ]] || { echo "output directory already exists" >&2; exit 1; }

[[ $(jq -er '.schema' "$CONTRACT") == "truerepublic.release-tool-platform/v1" ]] || { echo "tool-platform contract schema mismatch" >&2; exit 1; }
[[ $(jq -er '.version' "$GATES") == "truerepublic.security-gates/v1" ]] || { echo "security gate contract schema mismatch" >&2; exit 1; }
[[ $(jq -er --arg id "$TOOL" '[.bootstrap.tools[] | select(.id == $id)] | length' "$CONTRACT") == "1" ]] || {
  echo "tool $TOOL is not in the tool-platform bootstrap allowlist" >&2; exit 1; }
kind=$(jq -er --arg id "$TOOL" '.bootstrap.tools[] | select(.id == $id) | .kind' "$CONTRACT")
gates_key=$(jq -er --arg id "$TOOL" '.bootstrap.tools[] | select(.id == $id) | .gates_key' "$CONTRACT")
artifact=$(jq -er --arg id "$TOOL" '.bootstrap.tools[] | select(.id == $id) | .artifact' "$CONTRACT")
expected=$(jq -er --arg key "$gates_key" '.tools[$key]' "$GATES")
[[ "$expected" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo "security gate version for $gates_key is not exact" >&2; exit 1; }
[[ "$artifact" != /* && "$artifact" != ./* &&
   "$artifact" != "." && "$artifact" != ".." && "$artifact" != */. &&
   "$artifact" != ../* && "$artifact" != */../* && "$artifact" != */.. &&
   "$artifact" != */./* && "$artifact" != *//* && "$artifact" != *\\* ]] || {
  echo "tool artifact path is not a clean relative path" >&2; exit 1; }

mkdir "$OUTPUT_DIR"
complete=0
cleanup() {
  if [[ "$complete" != 1 && -d "$OUTPUT_DIR" && ! -L "$OUTPUT_DIR" ]]; then
    rm -rf "$OUTPUT_DIR"
  fi
}
trap cleanup EXIT

case "$kind" in
  go)
    module=$(jq -er --arg id "$TOOL" '.bootstrap.tools[] | select(.id == $id) | .module' "$CONTRACT")
    package=$(jq -er --arg id "$TOOL" '.bootstrap.tools[] | select(.id == $id) | .package' "$CONTRACT")
    [[ "$artifact" != */* ]] || { echo "go tool artifact must be a bare file name" >&2; exit 1; }
    cd "$ROOT_DIR/tools/ci"
    export GOTOOLCHAIN=local
    export GOFLAGS=-mod=readonly
    export CGO_ENABLED=0
    go mod verify >/dev/null
    locked=$(go list -mod=readonly -m -f '{{.Version}}' "$module")
    [[ "$locked" == "$expected" ]] || {
      echo "tools/ci module $module locks $locked, security gate requires $expected" >&2; exit 1; }
    version_args=()
    ldflags='-s -w -buildid='
    case "$TOOL" in
      cyclonedx-gomod) version_args=(version);;
      govulncheck|staticcheck) version_args=(-version);;
      gitleaks)
        version_args=(version)
        # gitleaks embeds its version only through this ldflags symbol;
        # inject the exact configured pin so the binary reports it.
        ldflags="$ldflags -X github.com/zricethezav/gitleaks/v8/version.Version=${expected#v}"
        ;;
      *) echo "go tool $TOOL has no configured version check" >&2; exit 1;;
    esac
    go build -mod=readonly -trimpath -buildvcs=false -ldflags="$ldflags" \
      -o "$OUTPUT_DIR/$artifact" "$package"
    version_output=$("$OUTPUT_DIR/$artifact" "${version_args[@]}" 2>&1 || true)
    numeric_version=${expected#v}
    version_pattern=${numeric_version//./\\.}
    [[ "$version_output" =~ (^|[^0-9.])${version_pattern}([^0-9.]|$) ]] || {
      echo "built $TOOL reports version ${version_output:-<none>}, security gate requires $expected" >&2; exit 1; }
    ;;
  npm)
    package=$(jq -er --arg id "$TOOL" '.bootstrap.tools[] | select(.id == $id) | .package' "$CONTRACT")
    npm_expected=$(jq -er '.tools.npm' "$GATES")
    [[ $(npm --version) == "$npm_expected" ]] || {
      echo "npm toolchain $(npm --version) does not match the security gate contract $npm_expected" >&2; exit 1; }
    [[ $(jq -er --arg pkg "$package" '.dependencies[$pkg]' "$ROOT_DIR/tools/ci/npm/package.json") == "$expected" ]] || {
      echo "tools/ci npm package.json does not pin $package $expected exactly" >&2; exit 1; }
    [[ $(jq -er --arg pkg "node_modules/$package" '.packages[$pkg].version' "$ROOT_DIR/tools/ci/npm/package-lock.json") == "$expected" ]] || {
      echo "tools/ci npm package-lock does not lock $package $expected exactly" >&2; exit 1; }
    cp "$ROOT_DIR/tools/ci/npm/package.json" "$ROOT_DIR/tools/ci/npm/package-lock.json" "$OUTPUT_DIR/"
    cd "$OUTPUT_DIR"
    npm ci --ignore-scripts --no-audit --no-fund >/dev/null
    [[ $(jq -er '.version' "$OUTPUT_DIR/$artifact") == "$expected" ]] || {
      echo "installed $package version drifted from $expected" >&2; exit 1; }
    ;;
  *)
    echo "tool $TOOL has an unsupported kind $kind" >&2; exit 1
    ;;
esac

complete=1
trap - EXIT
