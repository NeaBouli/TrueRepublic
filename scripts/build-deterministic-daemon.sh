#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(git -C "$(dirname "${BASH_SOURCE[0]}")/.." rev-parse --show-toplevel)
source "$ROOT_DIR/scripts/verify-deterministic-daemon.sh"
CONTRACT=
TARGET=
SOURCE_REF=
OUTPUT_DIR=

usage() {
  echo "usage: $0 --contract <path> --target <linux-amd64|linux-arm64> --source-ref <git-sha> --output-dir <dir>" >&2
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --contract) CONTRACT=$2; shift 2 ;;
    --target) TARGET=$2; shift 2 ;;
    --source-ref) SOURCE_REF=$2; shift 2 ;;
    --output-dir) OUTPUT_DIR=$2; shift 2 ;;
    *) usage; exit 2 ;;
  esac
done

[[ -n "$CONTRACT" && -n "$TARGET" && -n "$SOURCE_REF" && -n "$OUTPUT_DIR" ]] || { usage; exit 2; }
[[ "$CONTRACT" = /* ]] || CONTRACT="$ROOT_DIR/$CONTRACT"
[[ "$OUTPUT_DIR" = /* ]] || OUTPUT_DIR="$ROOT_DIR/$OUTPUT_DIR"

validate_contract "$CONTRACT" "$TARGET" "$SOURCE_REF"

[[ $(git -C "$ROOT_DIR" rev-parse HEAD) == "$SOURCE_REF" ]] || {
  echo "source ref must equal the checked-out commit" >&2
  exit 1
}
[[ -z $(git -C "$ROOT_DIR" status --porcelain --untracked-files=normal) ]] || {
  echo "deterministic builds require a clean repository" >&2
  exit 1
}

required_go=$(jq -r '.go_version' "$CONTRACT")
[[ $(go env GOVERSION) == "go${required_go}" ]] || {
  echo "Go ${required_go} is required" >&2
  exit 1
}

goos=$(jq -r --arg target "$TARGET" '.targets[] | select(.id == $target) | .goos' "$CONTRACT")
goarch=$(jq -r --arg target "$TARGET" '.targets[] | select(.id == $target) | .goarch' "$CONTRACT")
ci_runner=$(jq -r --arg target "$TARGET" '.targets[] | select(.id == $target) | .ci_runner' "$CONTRACT")
runner_arch=$(jq -r --arg target "$TARGET" '.targets[] | select(.id == $target) | .runner_arch' "$CONTRACT")
artifact=$(jq -r --arg target "$TARGET" '.targets[] | select(.id == $target) | .artifact' "$CONTRACT")
main_package=$(jq -r '.main_package' "$CONTRACT")
source_date_epoch=$(git -C "$ROOT_DIR" show -s --format=%ct "$SOURCE_REF")
ldflags=$(jq -r --arg source_ref "$SOURCE_REF" '
  .build_flags.ldflags |
  map(if . == "main.version={{source_ref}}" then "main.version=" + $source_ref else . end) |
  join(" ")
' "$CONTRACT")

if [[ $(uname -s) != "Linux" || $(uname -m) != "$runner_arch" ]]; then
  echo "$TARGET requires a native Linux $runner_arch runner" >&2
  exit 1
fi

if [[ -e "$OUTPUT_DIR" ]]; then
  echo "output directory already exists: $OUTPUT_DIR" >&2
  exit 1
fi
mkdir -p "$OUTPUT_DIR/attempt-1" "$OUTPUT_DIR/attempt-2"

build_once() {
  local attempt=$1
  local destination="$OUTPUT_DIR/$attempt/$artifact"
  local cache="$OUTPUT_DIR/.cache-$attempt"
  mkdir -p "$cache"
  echo "Building $TARGET ($attempt) from $SOURCE_REF"
  (
    cd "$ROOT_DIR"
    env \
      CGO_ENABLED=1 \
      GOOS="$goos" \
      GOARCH="$goarch" \
      GOCACHE="$cache" \
      GOFLAGS=-mod=readonly \
      LC_ALL=C \
      SOURCE_DATE_EPOCH="$source_date_epoch" \
      TZ=UTC \
      go build -trimpath -buildvcs=false -ldflags="$ldflags" -o "$destination" "$main_package"
  )
}

build_once attempt-1
build_once attempt-2

hash=$("$ROOT_DIR/scripts/verify-deterministic-daemon.sh" \
  "$CONTRACT" "$TARGET" "$SOURCE_REF" \
  "$OUTPUT_DIR/attempt-1/$artifact" "$OUTPUT_DIR/attempt-2/$artifact")

cp "$OUTPUT_DIR/attempt-1/$artifact" "$OUTPUT_DIR/$artifact"
printf '%s  %s\n' "$hash" "$artifact" >"$OUTPUT_DIR/CHECKSUMS.sha256"

contract_hash=$(sha256_file "$CONTRACT")
jq -n -S \
  --arg schema "truerepublic.daemon-build-evidence/v1" \
  --arg contract_schema "truerepublic.daemon-build/v1" \
  --arg contract_sha256 "$contract_hash" \
  --arg source_ref "$SOURCE_REF" \
  --arg target "$TARGET" \
  --arg ci_runner "$ci_runner" \
  --arg runner_arch "$runner_arch" \
  --arg artifact "$artifact" \
  --arg sha256 "$hash" \
  --arg go_version "$required_go" \
  --arg source_date_epoch "$source_date_epoch" \
  '{
    schema: $schema,
    contract_schema: $contract_schema,
    contract_sha256: $contract_sha256,
    source_ref: $source_ref,
    target: $target,
    ci_runner: $ci_runner,
    runner_arch: $runner_arch,
    artifact: $artifact,
    sha256: $sha256,
    reproducible_pair_sha256: [$sha256, $sha256],
    go_version: $go_version,
    cgo_enabled: "1",
    source_date_epoch: ($source_date_epoch | tonumber),
    build_flags: {
      trimpath: true,
      buildvcs: false,
      mod: "readonly",
      buildid: "",
      linker_build_id: "none",
      version_variable: "main.version"
    }
  }' >"$OUTPUT_DIR/build-metadata.json"

rm -rf "$OUTPUT_DIR/attempt-1" "$OUTPUT_DIR/attempt-2" \
  "$OUTPUT_DIR/.cache-attempt-1" "$OUTPUT_DIR/.cache-attempt-2"
echo "Verified reproducible $TARGET SHA-256: $hash"
