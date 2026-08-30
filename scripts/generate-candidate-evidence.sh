#!/usr/bin/env bash
set -euo pipefail
export LC_ALL=C LANG=C

ROOT_DIR=$(git -C "$(dirname "${BASH_SOURCE[0]}")/.." rev-parse --show-toplevel)
CONTRACT="$ROOT_DIR/configs/release/candidate-evidence.json"
TAG=
COMMIT=
AMD64_DIR=
ARM64_DIR=
OCI_AMD64_DIR=
OCI_ARM64_DIR=
OUTPUT_DIR=
usage() { echo "usage: $0 --tag <vX.Y.Z> --commit <40-hex> --amd64-dir <dir> --arm64-dir <dir> --oci-amd64-dir <dir> --oci-arm64-dir <dir> --output-dir <new-dir> [--contract <file>]" >&2; }
while [[ $# -gt 0 ]]; do
  case "$1" in
    --tag) TAG=${2-}; shift 2;; --commit) COMMIT=${2-}; shift 2;;
    --amd64-dir) AMD64_DIR=${2-}; shift 2;; --arm64-dir) ARM64_DIR=${2-}; shift 2;;
    --oci-amd64-dir) OCI_AMD64_DIR=${2-}; shift 2;; --oci-arm64-dir) OCI_ARM64_DIR=${2-}; shift 2;;
    --output-dir) OUTPUT_DIR=${2-}; shift 2;; --contract) CONTRACT=${2-}; shift 2;;
    *) usage; exit 2;;
  esac
done
[[ -n "$TAG" && -n "$COMMIT" && -n "$AMD64_DIR" && -n "$ARM64_DIR" && -n "$OCI_AMD64_DIR" && -n "$OCI_ARM64_DIR" && -n "$OUTPUT_DIR" ]] || { usage; exit 2; }
[[ "$TAG" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]] || { echo "tag must be a simulated semantic version like v0.4.0" >&2; exit 1; }
[[ "$COMMIT" =~ ^[0-9a-f]{40}$ ]] || { echo "commit must be exactly 40 lowercase hexadecimal characters" >&2; exit 1; }
[[ ! -e "$OUTPUT_DIR" && ! -L "$OUTPUT_DIR" ]] || { echo "output directory already exists" >&2; exit 1; }
[[ -d "$(dirname "$OUTPUT_DIR")" ]] || { echo "output parent directory is unavailable" >&2; exit 1; }

sha256_file() { if command -v sha256sum >/dev/null; then sha256sum "$1"|awk '{print $1}'; else shasum -a 256 "$1"|awk '{print $1}'; fi; }

contract_hash=$(sha256_file "$CONTRACT")
binary_contract_hash=$(jq -er '.binary_contract_sha256' "$CONTRACT")
oci_contract_hash=$(jq -er '.oci_contract_sha256' "$CONTRACT")
[[ "$binary_contract_hash" =~ ^[0-9a-f]{64}$ && "$oci_contract_hash" =~ ^[0-9a-f]{64}$ ]] || { echo "candidate contract pins are malformed" >&2; exit 1; }

canonical_inputs=()
for dir in "$AMD64_DIR" "$ARM64_DIR" "$OCI_AMD64_DIR" "$OCI_ARM64_DIR"; do
  [[ -d "$dir" && ! -L "$dir" ]] || { echo "input directory $dir is unavailable or symlinked" >&2; exit 1; }
  canonical=$(cd "$dir" && pwd -P)
  for prior in "${canonical_inputs[@]-}"; do
    [[ -z "$prior" || "$canonical" != "$prior" ]] || { echo "input directories must be distinct" >&2; exit 1; }
  done
  canonical_inputs+=("$canonical")
done

require_dir() {
  local dir=$1; shift
  [[ -d "$dir" && ! -L "$dir" ]] || { echo "input directory $dir is unavailable or symlinked" >&2; exit 1; }
  [[ $(find "$dir" -mindepth 1 -maxdepth 1 | wc -l | tr -d ' ') -eq $# ]] || { echo "input directory $dir has an unexpected member count" >&2; exit 1; }
  local name
  for name in "$@"; do
    [[ -f "$dir/$name" && ! -L "$dir/$name" ]] || { echo "input $dir/$name is missing or symlinked" >&2; exit 1; }
  done
}
require_json() { [[ $(wc -c <"$1" | tr -d ' ') -le 1048576 ]] || { echo "input $1 exceeds the JSON byte limit" >&2; exit 1; }; }

tmp=$(mktemp -d)
complete=0
trap 'rm -rf "$tmp"; if [[ "$complete" != 1 && -d "$OUTPUT_DIR" ]]; then rm -rf "$OUTPUT_DIR"; fi' EXIT
mkdir "$OUTPUT_DIR"

bind_binary() {
  local id=$1 runner=$2 arch=$3 artifact=$4 dir=$5
  require_dir "$dir" CHECKSUMS.sha256 build-metadata.json
  require_json "$dir/build-metadata.json"
  [[ $(wc -c <"$dir/CHECKSUMS.sha256" | tr -d ' ') -le 4096 ]] || { echo "checksum input for $id exceeds the byte limit" >&2; exit 1; }
  jq -e --arg hash "$binary_contract_hash" --arg commit "$COMMIT" --arg id "$id" --arg runner "$runner" --arg arch "$arch" --arg artifact "$artifact" '
    .schema=="truerepublic.daemon-build-evidence/v1" and
    .contract_schema=="truerepublic.daemon-build/v1" and
    .contract_sha256==$hash and
    .source_ref==$commit and
    .target==$id and
    .ci_runner==$runner and
    .runner_arch==$arch and
    .artifact==$artifact and
    (.sha256|test("^[0-9a-f]{64}$")) and
    .reproducible_pair_sha256==[.sha256,.sha256] and
    .go_version=="1.26.6" and
    .cgo_enabled=="1" and
    (.source_date_epoch|type=="number" and floor==. and .>0) and
    .build_flags=={trimpath:true,buildvcs:false,mod:"readonly",buildid:"",linker_build_id:"none",version_variable:"main.version"}
  ' "$dir/build-metadata.json" >/dev/null || { echo "binary metadata mismatch for $id" >&2; exit 1; }
  local digest
  digest=$(jq -er '.sha256' "$dir/build-metadata.json")
  printf '%s  %s\n' "$digest" "$artifact" | cmp -s - "$dir/CHECKSUMS.sha256" || { echo "binary checksum mismatch for $id" >&2; exit 1; }
  cp "$dir/build-metadata.json" "$OUTPUT_DIR/build-metadata-$id.json"
  cp "$dir/CHECKSUMS.sha256" "$OUTPUT_DIR/checksums-$id.sha256"
  BINARY_DIGESTS+=("$digest")
}

bind_oci() {
  local suffix=$1 platform=$2 first=$3 second=$4 dir=$5
  require_dir "$dir" oci-evidence.json "oci-evidence-report-$suffix.json"
  require_json "$dir/oci-evidence.json"
  require_json "$dir/oci-evidence-report-$suffix.json"
  jq -e --arg hash "$oci_contract_hash" --arg commit "$COMMIT" --arg platform "$platform" --arg first "$first" --arg second "$second" '
    .schema=="truerepublic.oci-evidence/v1" and
    .source_ref==$commit and
    .contract_sha256==$hash and
    .platform==$platform and
    .claims=={signed:false,published:false,production:false} and
    (.targets|length)==2 and
    .targets[0].id==$first and .targets[1].id==$second and
    all(.targets[]; (.builds|length)==2 and
      all(.builds[]; (.sha256|test("^[0-9a-f]{64}$")) and (.file|test("^[a-z0-9][a-z0-9.-]*\\.oci\\.tar$"))))
  ' "$dir/oci-evidence.json" >/dev/null || { echo "OCI evidence mismatch for $platform" >&2; exit 1; }
  jq -e --arg platform "$platform" --arg first "$first" --arg second "$second" '
    .schema=="truerepublic.oci-evidence-report/v1" and
    .valid==true and
    .platform==$platform and
    .targets==2 and
    .violations==[] and
    ((.layer_diffs // [])==[]) and
    (.images|length)==2 and
    .images[0].id==$first and .images[1].id==$second and
    ((.images[0].repetition // 0)==0) and ((.images[1].repetition // 0)==0) and
    all(.images[];
      (.index_sha256|test("^[0-9a-f]{64}$")) and
      (.manifest_sha256|test("^[0-9a-f]{64}$")) and
      (.config_sha256|test("^[0-9a-f]{64}$")) and
      (.layer_sha256|length)>=1 and (.layer_sha256|length)<=128 and
      all(.layer_sha256[]; test("^[0-9a-f]{64}$")))
  ' "$dir/oci-evidence-report-$suffix.json" >/dev/null || { echo "OCI digest report mismatch for $platform" >&2; exit 1; }
  cp "$dir/oci-evidence.json" "$OUTPUT_DIR/oci-evidence-$suffix.json"
  cp "$dir/oci-evidence-report-$suffix.json" "$OUTPUT_DIR/oci-report-$suffix.json"
  jq -cS '[.images[] | {config_sha256:.config_sha256,id:.id,index_sha256:.index_sha256,layer_sha256:.layer_sha256,manifest_sha256:.manifest_sha256}]' \
    "$dir/oci-evidence-report-$suffix.json" >"$tmp/oci-targets-$suffix.json"
}

BINARY_DIGESTS=()
bind_binary linux-amd64 ubuntu-24.04 x86_64 truerepublicd-linux-amd64 "$AMD64_DIR"
bind_binary linux-arm64 ubuntu-24.04-arm aarch64 truerepublicd-linux-arm64 "$ARM64_DIR"
bind_oci amd64 linux/amd64 daemon-linux-amd64 client-web-linux-amd64 "$OCI_AMD64_DIR"
bind_oci arm64 linux/arm64 daemon-linux-arm64 client-web-linux-arm64 "$OCI_ARM64_DIR"

amd_meta_hash=$(sha256_file "$OUTPUT_DIR/build-metadata-linux-amd64.json")
amd_sum_hash=$(sha256_file "$OUTPUT_DIR/checksums-linux-amd64.sha256")
arm_meta_hash=$(sha256_file "$OUTPUT_DIR/build-metadata-linux-arm64.json")
arm_sum_hash=$(sha256_file "$OUTPUT_DIR/checksums-linux-arm64.sha256")
amd_ev_hash=$(sha256_file "$OUTPUT_DIR/oci-evidence-amd64.json")
amd_rep_hash=$(sha256_file "$OUTPUT_DIR/oci-report-amd64.json")
arm_ev_hash=$(sha256_file "$OUTPUT_DIR/oci-evidence-arm64.json")
arm_rep_hash=$(sha256_file "$OUTPUT_DIR/oci-report-arm64.json")

jq -n -S \
  --arg contract_hash "$contract_hash" --arg tag "$TAG" --arg commit "$COMMIT" \
  --arg amd_sha "${BINARY_DIGESTS[0]}" --arg arm_sha "${BINARY_DIGESTS[1]}" \
  --arg amd_meta_hash "$amd_meta_hash" --arg amd_sum_hash "$amd_sum_hash" \
  --arg arm_meta_hash "$arm_meta_hash" --arg arm_sum_hash "$arm_sum_hash" \
  --arg amd_ev_hash "$amd_ev_hash" --arg amd_rep_hash "$amd_rep_hash" \
  --arg arm_ev_hash "$arm_ev_hash" --arg arm_rep_hash "$arm_rep_hash" \
  --slurpfile amd_targets "$tmp/oci-targets-amd64.json" \
  --slurpfile arm_targets "$tmp/oci-targets-arm64.json" \
  '{
    schema:"truerepublic.release-candidate-evidence/v1",
    contract_sha256:$contract_hash,
    source:{tag:$tag,commit:$commit},
    claims:{real_tag_created:false,ref_pushed:false,signed:false,published:false,deployed:false,production:false},
    binary_targets:[
      {id:"linux-amd64",artifact:"truerepublicd-linux-amd64",artifact_sha256:$amd_sha,
       metadata:{file:"build-metadata-linux-amd64.json",sha256:$amd_meta_hash},
       checksums:{file:"checksums-linux-amd64.sha256",sha256:$amd_sum_hash}},
      {id:"linux-arm64",artifact:"truerepublicd-linux-arm64",artifact_sha256:$arm_sha,
       metadata:{file:"build-metadata-linux-arm64.json",sha256:$arm_meta_hash},
       checksums:{file:"checksums-linux-arm64.sha256",sha256:$arm_sum_hash}}
    ],
    oci_platforms:[
      {platform:"linux/amd64",
       evidence:{file:"oci-evidence-amd64.json",sha256:$amd_ev_hash},
       report:{file:"oci-report-amd64.json",sha256:$amd_rep_hash},
       targets:$amd_targets[0]},
      {platform:"linux/arm64",
       evidence:{file:"oci-evidence-arm64.json",sha256:$arm_ev_hash},
       report:{file:"oci-report-arm64.json",sha256:$arm_rep_hash},
       targets:$arm_targets[0]}
    ]
  }' >"$OUTPUT_DIR/candidate-evidence.json"

"$ROOT_DIR/scripts/verify-candidate-evidence.sh" --evidence "$OUTPUT_DIR" --contract "$CONTRACT"
complete=1
trap - EXIT
rm -rf "$tmp"
