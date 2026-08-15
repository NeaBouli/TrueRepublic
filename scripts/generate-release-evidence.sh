#!/usr/bin/env bash
set -euo pipefail
export LC_ALL=C LANG=C

ROOT_DIR=$(git -C "$(dirname "${BASH_SOURCE[0]}")/.." rev-parse --show-toplevel)
BUILD_CONTRACT="$ROOT_DIR/configs/build/deterministic-linux-daemon.json"
TOOL_CONTRACT="$ROOT_DIR/configs/release/tool-platform.json"
SOURCE_REF=
AMD64_DIR=
ARM64_DIR=
GO_SBOM_A=
GO_SBOM_B=
CLIENT_SBOM_A=
CLIENT_SBOM_B=
OUTPUT_DIR=
usage() { echo "usage: $0 --source-ref <40-hex> --amd64-dir <dir> --arm64-dir <dir> --go-sbom-a <raw.json> --go-sbom-b <raw.json> --client-sbom-a <raw.json> --client-sbom-b <raw.json> --output-dir <new-dir> [--build-contract <file>] [--tool-contract <file>]" >&2; }
while [[ $# -gt 0 ]]; do
  case "$1" in
    --source-ref) SOURCE_REF=${2-}; shift 2;; --amd64-dir) AMD64_DIR=${2-}; shift 2;; --arm64-dir) ARM64_DIR=${2-}; shift 2;;
    --go-sbom-a) GO_SBOM_A=${2-}; shift 2;; --go-sbom-b) GO_SBOM_B=${2-}; shift 2;;
    --client-sbom-a) CLIENT_SBOM_A=${2-}; shift 2;; --client-sbom-b) CLIENT_SBOM_B=${2-}; shift 2;;
    --output-dir) OUTPUT_DIR=${2-}; shift 2;;
    --build-contract) BUILD_CONTRACT=${2-}; shift 2;; --tool-contract) TOOL_CONTRACT=${2-}; shift 2;; *) usage; exit 2;;
  esac
done
[[ -n "$SOURCE_REF" && -n "$AMD64_DIR" && -n "$ARM64_DIR" && -n "$GO_SBOM_A" && -n "$GO_SBOM_B" && -n "$CLIENT_SBOM_A" && -n "$CLIENT_SBOM_B" && -n "$OUTPUT_DIR" ]] || { usage; exit 2; }
[[ "$SOURCE_REF" =~ ^[0-9a-f]{40}$ ]] || { echo "source ref must be exactly 40 lowercase hexadecimal characters" >&2; exit 1; }
[[ ! -e "$OUTPUT_DIR" ]] || { echo "output directory already exists" >&2; exit 1; }
jq -e '.schema=="truerepublic.release-tool-platform/v1" and .tools=={"cyclonedx_gomod":"v1.10.0","cyclonedx_npm":"6.0.1","node":"22.22.2","npm":"10.9.7"} and (.platforms|length)==2 and (.base_images|length)==4' "$TOOL_CONTRACT" >/dev/null
tool_cyclonedx_gomod=$(jq -er '.tools.cyclonedx_gomod' "$TOOL_CONTRACT")
tool_cyclonedx_npm=$(jq -er '.tools.cyclonedx_npm' "$TOOL_CONTRACT")
tool_node=$(jq -er '.tools.node' "$TOOL_CONTRACT")
tool_npm=$(jq -er '.tools.npm' "$TOOL_CONTRACT")
tmp=$(mktemp -d)
complete=0
trap 'rm -rf "$tmp"; if [[ "$complete" != 1 && -d "$OUTPUT_DIR" ]]; then rm -rf "$OUTPUT_DIR"; fi' EXIT
mkdir "$OUTPUT_DIR"
sha256_file() { if command -v sha256sum >/dev/null; then sha256sum "$1"|awk '{print $1}'; else shasum -a 256 "$1"|awk '{print $1}'; fi; }
normalize() { jq -cS 'del(.serialNumber, .metadata.timestamp)' "$1" >"$2"; }
for component in go client; do
  raw_a=$GO_SBOM_A; raw_b=$GO_SBOM_B
  if [[ "$component" == client ]]; then raw_a=$CLIENT_SBOM_A; raw_b=$CLIENT_SBOM_B; fi
  for raw in "$raw_a" "$raw_b"; do
    [[ -f "$raw" && ! -L "$raw" && $(wc -c <"$raw") -le 1048576 ]] || { echo "$component SBOM input is unavailable, symlinked, or oversized" >&2; exit 1; }
  done
  normalize "$raw_a" "$tmp/$component.1.json"; normalize "$raw_b" "$tmp/$component.2.json"
  cmp -s "$tmp/$component.1.json" "$tmp/$component.2.json" || { echo "$component SBOM normalization parity failed" >&2; exit 1; }
  jq -e '.bomFormat=="CycloneDX" and (.specVersion|type=="string") and (.components|type=="array")' "$tmp/$component.1.json" >/dev/null
  cp "$tmp/$component.1.json" "$OUTPUT_DIR/$component.cdx.json"
done
copy_target() {
  local id=$1 dir=$2 artifact=$3
  cp "$dir/$artifact" "$OUTPUT_DIR/$artifact"
  cp "$dir/CHECKSUMS.sha256" "$OUTPUT_DIR/$id.CHECKSUMS.sha256"
  cp "$dir/build-metadata.json" "$OUTPUT_DIR/$id.build-metadata.json"
}
copy_target linux-amd64 "$AMD64_DIR" truerepublicd-linux-amd64
copy_target linux-arm64 "$ARM64_DIR" truerepublicd-linux-arm64
build_hash=$(sha256_file "$BUILD_CONTRACT"); tool_hash=$(sha256_file "$TOOL_CONTRACT")
amd_hash=$(sha256_file "$OUTPUT_DIR/truerepublicd-linux-amd64"); arm_hash=$(sha256_file "$OUTPUT_DIR/truerepublicd-linux-arm64")
go_hash=$(sha256_file "$OUTPUT_DIR/go.cdx.json"); client_hash=$(sha256_file "$OUTPUT_DIR/client.cdx.json")
jq -n -cS --arg source "$SOURCE_REF" --arg bh "$build_hash" --arg th "$tool_hash" --arg ah "$amd_hash" --arg rh "$arm_hash" --arg gh "$go_hash" --arg ch "$client_hash" '{schema:"truerepublic.unsigned-provenance/v1",source_ref:$source,build_contract_sha256:$bh,tool_contract_sha256:$th,claims:{signed:false,published:false,production:false},targets:[{id:"linux-amd64",sha256:$ah},{id:"linux-arm64",sha256:$rh}],sboms:[{id:"go",sha256:$gh},{id:"client",sha256:$ch}]}' >"$OUTPUT_DIR/provenance.json"
provenance_hash=$(sha256_file "$OUTPUT_DIR/provenance.json")
jq -n -S --arg source "$SOURCE_REF" --arg bh "$build_hash" --arg th "$tool_hash" --arg ah "$amd_hash" --arg rh "$arm_hash" --arg gh "$go_hash" --arg ch "$client_hash" --arg ph "$provenance_hash" --arg cxg "$tool_cyclonedx_gomod" --arg cxn "$tool_cyclonedx_npm" --arg node "$tool_node" --arg npm "$tool_npm" '{schema:"truerepublic.release-evidence/v1",source_ref:$source,build_contract_sha256:$bh,tool_contract_sha256:$th,tools:{cyclonedx_gomod:$cxg,cyclonedx_npm:$cxn,node:$node,npm:$npm},claims:{signed:false,published:false,production:false},provenance:{file:"provenance.json",sha256:$ph},targets:[{id:"linux-amd64",artifact:"truerepublicd-linux-amd64",sha256:$ah,checksums_file:"linux-amd64.CHECKSUMS.sha256",metadata_file:"linux-amd64.build-metadata.json"},{id:"linux-arm64",artifact:"truerepublicd-linux-arm64",sha256:$rh,checksums_file:"linux-arm64.CHECKSUMS.sha256",metadata_file:"linux-arm64.build-metadata.json"}],sboms:[{component:"go",file:"go.cdx.json",sha256:$gh},{component:"client",file:"client.cdx.json",sha256:$ch}]}' >"$OUTPUT_DIR/release-evidence.json"
"$ROOT_DIR/scripts/verify-release-evidence.sh" --bundle "$OUTPUT_DIR" --build-contract "$BUILD_CONTRACT" --tool-contract "$TOOL_CONTRACT"
complete=1
trap - EXIT
rm -rf "$tmp"
