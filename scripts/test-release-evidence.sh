#!/usr/bin/env bash
set -euo pipefail
export LC_ALL=C LANG=C

ROOT_DIR=$(git -C "$(dirname "${BASH_SOURCE[0]}")/.." rev-parse --show-toplevel)
cd "$ROOT_DIR"
go test ./releaseevidence

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
source_ref=0123456789abcdef0123456789abcdef01234567

sha256_file() {
  if command -v sha256sum >/dev/null; then
    sha256sum "$1" | awk '{print $1}'
  else
    shasum -a 256 "$1" | awk '{print $1}'
  fi
}

build_hash=$(sha256_file configs/build/deterministic-linux-daemon.json)
make_target() {
  local id=$1 runner=$2 runner_arch=$3 artifact=$4
  local dir="$tmp/$id"
  mkdir -p "$dir"
  printf 'synthetic-%s\n' "$id" >"$dir/$artifact"
  local digest
  digest=$(sha256_file "$dir/$artifact")
  printf '%s  %s\n' "$digest" "$artifact" >"$dir/CHECKSUMS.sha256"
  jq -n -S \
    --arg contract_hash "$build_hash" --arg source "$source_ref" --arg target "$id" \
    --arg runner "$runner" --arg runner_arch "$runner_arch" --arg artifact "$artifact" --arg digest "$digest" \
    '{schema:"truerepublic.daemon-build-evidence/v1",contract_schema:"truerepublic.daemon-build/v1",contract_sha256:$contract_hash,source_ref:$source,target:$target,ci_runner:$runner,runner_arch:$runner_arch,artifact:$artifact,sha256:$digest,reproducible_pair_sha256:[$digest,$digest],go_version:"1.26.6",cgo_enabled:"1",source_date_epoch:1,build_flags:{trimpath:true,buildvcs:false,mod:"readonly",buildid:"",linker_build_id:"none",version_variable:"main.version"}}' \
    >"$dir/build-metadata.json"
}
make_target linux-amd64 ubuntu-24.04 x86_64 truerepublicd-linux-amd64
make_target linux-arm64 ubuntu-24.04-arm aarch64 truerepublicd-linux-arm64

for component in go client; do
  jq -n -c --arg serial "urn:uuid:${component}-first" --arg timestamp '2026-08-16T00:00:00Z' --arg component "$component" \
    '{bomFormat:"CycloneDX",specVersion:"1.6",serialNumber:$serial,metadata:{timestamp:$timestamp},components:[{name:$component}]}' \
    >"$tmp/$component-a.json"
  jq -n -c --arg serial "urn:uuid:${component}-second" --arg timestamp '2026-08-16T00:00:01Z' --arg component "$component" \
    '{components:[{name:$component}],metadata:{timestamp:$timestamp},serialNumber:$serial,specVersion:"1.6",bomFormat:"CycloneDX"}' \
    >"$tmp/$component-b.json"
done

./scripts/generate-release-evidence.sh \
  --source-ref "$source_ref" \
  --amd64-dir "$tmp/linux-amd64" \
  --arm64-dir "$tmp/linux-arm64" \
  --go-sbom-a "$tmp/go-a.json" --go-sbom-b "$tmp/go-b.json" \
  --client-sbom-a "$tmp/client-a.json" --client-sbom-b "$tmp/client-b.json" \
  --output-dir "$tmp/bundle" >/dev/null
./scripts/verify-release-evidence.sh --bundle "$tmp/bundle" >/dev/null

if ./scripts/verify-release-evidence.sh --bundle "$tmp/bundle" --unknown value >/dev/null 2>&1; then
  echo "verifier accepted an unknown flag" >&2
  exit 1
fi
if ./scripts/verify-release-evidence.sh --bundle "$tmp/missing" >/dev/null 2>&1; then
  echo "verifier accepted a missing bundle" >&2
  exit 1
fi

printf '%064d  truerepublicd-linux-arm64\n' 0 >"$tmp/linux-arm64/CHECKSUMS.sha256"
if ./scripts/generate-release-evidence.sh \
  --source-ref "$source_ref" --amd64-dir "$tmp/linux-amd64" --arm64-dir "$tmp/linux-arm64" \
  --go-sbom-a "$tmp/go-a.json" --go-sbom-b "$tmp/go-b.json" \
  --client-sbom-a "$tmp/client-a.json" --client-sbom-b "$tmp/client-b.json" \
  --output-dir "$tmp/invalid-final" >/dev/null 2>&1; then
  echo "generator accepted mismatched input evidence" >&2
  exit 1
fi
if [[ -e "$tmp/invalid-final" ]]; then
  echo "generator left a complete-looking bundle after verification failure" >&2
  exit 1
fi
arm_digest=$(sha256_file "$tmp/linux-arm64/truerepublicd-linux-arm64")
printf '%s  truerepublicd-linux-arm64\n' "$arm_digest" >"$tmp/linux-arm64/CHECKSUMS.sha256"

printf 'tamper\n' >>"$tmp/bundle/truerepublicd-linux-amd64"
if ./scripts/verify-release-evidence.sh --bundle "$tmp/bundle" >/dev/null 2>&1; then
  echo "verifier accepted a tampered artifact" >&2
  exit 1
fi

jq '.components[0].name="different"' "$tmp/client-b.json" >"$tmp/client-mismatch.json"
if ./scripts/generate-release-evidence.sh \
  --source-ref "$source_ref" --amd64-dir "$tmp/linux-amd64" --arm64-dir "$tmp/linux-arm64" \
  --go-sbom-a "$tmp/go-a.json" --go-sbom-b "$tmp/go-b.json" \
  --client-sbom-a "$tmp/client-a.json" --client-sbom-b "$tmp/client-mismatch.json" \
  --output-dir "$tmp/parity-failure" >/dev/null 2>&1; then
  echo "generator accepted divergent repeated SBOM output" >&2
  exit 1
fi

if ./scripts/generate-release-evidence.sh --source-ref ../escape --amd64-dir "$tmp" --arm64-dir "$tmp" --go-sbom-a "$tmp/go-a.json" --go-sbom-b "$tmp/go-b.json" --client-sbom-a "$tmp/client-a.json" --client-sbom-b "$tmp/client-b.json" --output-dir "$tmp/out" >/dev/null 2>&1; then
  echo "generator accepted a malformed source ref" >&2
  exit 1
fi
