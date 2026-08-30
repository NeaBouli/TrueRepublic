#!/usr/bin/env bash
set -euo pipefail
export LC_ALL=C LANG=C

ROOT_DIR=$(git -C "$(dirname "${BASH_SOURCE[0]}")/.." rev-parse --show-toplevel)
cd "$ROOT_DIR"
export GOTOOLCHAIN=local
export GOPROXY=off
export GOFLAGS=-mod=readonly

bash -n scripts/generate-candidate-evidence.sh
bash -n scripts/verify-candidate-evidence.sh
bash -n scripts/test-candidate-evidence.sh

go test ./candidateevidence ./cmd/candidate-evidence -count=1

FIXTURES="$ROOT_DIR/testdata/candidateevidence"
./scripts/verify-candidate-evidence.sh --evidence "$FIXTURES/valid" >/dev/null

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

first="$tmp/first.json"
second="$tmp/second.json"
./scripts/verify-candidate-evidence.sh --evidence "$FIXTURES/valid" --output json >"$first"
./scripts/verify-candidate-evidence.sh --evidence "$FIXTURES/valid" --output json >"$second"
cmp "$first" "$second"

if ./scripts/verify-candidate-evidence.sh --evidence "$FIXTURES/invalid-claims" >/dev/null 2>&1; then
  echo "candidate verifier accepted true status claims" >&2
  exit 1
fi
if ./scripts/verify-candidate-evidence.sh --unknown value >/dev/null 2>&1; then
  echo "candidate verifier accepted an unknown flag" >&2
  exit 1
fi
if ./scripts/verify-candidate-evidence.sh --evidence >/dev/null 2>&1; then
  echo "candidate verifier accepted a flag without a value" >&2
  exit 1
fi
if ./scripts/verify-candidate-evidence.sh --evidence "$ROOT_DIR/does-not-exist" >/dev/null 2>&1; then
  echo "candidate verifier accepted a missing evidence directory" >&2
  exit 1
fi

sha256_file() {
  if command -v sha256sum >/dev/null; then
    sha256sum "$1" | awk '{print $1}'
  else
    shasum -a 256 "$1" | awk '{print $1}'
  fi
}
sha256_stdin() {
  if command -v sha256sum >/dev/null; then
    sha256sum | awk '{print $1}'
  else
    shasum -a 256 | awk '{print $1}'
  fi
}

source_ref=0123456789abcdef0123456789abcdef01234567
tag=v0.5.0
build_hash=$(sha256_file configs/build/deterministic-linux-daemon.json)
oci_hash=$(sha256_file configs/build/reproducible-oci.json)

make_binary() {
  local id=$1 runner=$2 arch=$3 artifact=$4
  local dir="$tmp/in-$id"
  mkdir -p "$dir"
  local digest
  digest=$(printf 'synthetic-binary-%s\n' "$id" | sha256_stdin)
  printf '%s  %s\n' "$digest" "$artifact" >"$dir/CHECKSUMS.sha256"
  jq -n -S \
    --arg contract_hash "$build_hash" --arg source "$source_ref" --arg target "$id" \
    --arg runner "$runner" --arg runner_arch "$arch" --arg artifact "$artifact" --arg digest "$digest" \
    '{schema:"truerepublic.daemon-build-evidence/v1",contract_schema:"truerepublic.daemon-build/v1",contract_sha256:$contract_hash,source_ref:$source,target:$target,ci_runner:$runner,runner_arch:$runner_arch,artifact:$artifact,sha256:$digest,reproducible_pair_sha256:[$digest,$digest],go_version:"1.26.6",cgo_enabled:"1",source_date_epoch:1,build_flags:{trimpath:true,buildvcs:false,mod:"readonly",buildid:"",linker_build_id:"none",version_variable:"main.version"}}' \
    >"$dir/build-metadata.json"
}

report_image() {
  local id=$1
  jq -n -S --arg id "$id" \
    --arg index "$(printf 'index-%s\n' "$id" | sha256_stdin)" \
    --arg manifest "$(printf 'manifest-%s\n' "$id" | sha256_stdin)" \
    --arg config "$(printf 'config-%s\n' "$id" | sha256_stdin)" \
    --arg layer "$(printf 'layer-%s\n' "$id" | sha256_stdin)" \
    '{id:$id,index_sha256:$index,manifest_sha256:$manifest,config_sha256:$config,layer_sha256:[$layer]}'
}

make_oci() {
  local suffix=$1 platform=$2
  local dir="$tmp/in-oci-$suffix"
  mkdir -p "$dir"
  local daemon="daemon-linux-$suffix" client="client-web-linux-$suffix"
  jq -n -S \
    --arg source "$source_ref" --arg contract_hash "$oci_hash" --arg platform "$platform" \
    --arg daemon "$daemon" --arg client "$client" \
    --arg d1 "$(printf 'archive-%s-1\n' "$daemon" | sha256_stdin)" \
    --arg d2 "$(printf 'archive-%s-2\n' "$daemon" | sha256_stdin)" \
    --arg c1 "$(printf 'archive-%s-1\n' "$client" | sha256_stdin)" \
    --arg c2 "$(printf 'archive-%s-2\n' "$client" | sha256_stdin)" \
    '{schema:"truerepublic.oci-evidence/v1",source_ref:$source,contract_sha256:$contract_hash,platform:$platform,claims:{signed:false,published:false,production:false},targets:[{id:$daemon,builds:[{file:($daemon+"-1.oci.tar"),sha256:$d1},{file:($daemon+"-2.oci.tar"),sha256:$d2}]},{id:$client,builds:[{file:($client+"-1.oci.tar"),sha256:$c1},{file:($client+"-2.oci.tar"),sha256:$c2}]}]}' \
    >"$dir/oci-evidence.json"
  jq -n -S --arg platform "$platform" \
    --argjson daemon "$(report_image "$daemon")" --argjson client "$(report_image "$client")" \
    '{schema:"truerepublic.oci-evidence-report/v1",valid:true,platform:$platform,targets:2,violations:[],images:[$daemon,$client]}' \
    >"$dir/oci-evidence-report-$suffix.json"
}

make_binary linux-amd64 ubuntu-24.04 x86_64 truerepublicd-linux-amd64
make_binary linux-arm64 ubuntu-24.04-arm aarch64 truerepublicd-linux-arm64
make_oci amd64 linux/amd64
make_oci arm64 linux/arm64

generate() {
  ./scripts/generate-candidate-evidence.sh \
    --tag "$tag" --commit "$source_ref" \
    --amd64-dir "$tmp/in-linux-amd64" --arm64-dir "$tmp/in-linux-arm64" \
    --oci-amd64-dir "$tmp/in-oci-amd64" --oci-arm64-dir "$tmp/in-oci-arm64" \
    --output-dir "$1"
}

generate "$tmp/candidate-a" >/dev/null
./scripts/verify-candidate-evidence.sh --evidence "$tmp/candidate-a" >/dev/null
generate "$tmp/candidate-b" >/dev/null
cmp "$tmp/candidate-a/candidate-evidence.json" "$tmp/candidate-b/candidate-evidence.json"

expect_generator_failure() {
  local name=$1 out="$tmp/fail-$1"
  shift
  if ./scripts/generate-candidate-evidence.sh \
    --tag "$tag" --commit "$source_ref" \
    --amd64-dir "$tmp/in-linux-amd64" --arm64-dir "$tmp/in-linux-arm64" \
    --oci-amd64-dir "$tmp/in-oci-amd64" --oci-arm64-dir "$tmp/in-oci-arm64" \
    --output-dir "$out" "$@" >/dev/null 2>&1; then
    echo "candidate generator accepted $name" >&2
    exit 1
  fi
  if [[ -e "$out" ]]; then
    echo "candidate generator left a complete-looking bundle after $name" >&2
    exit 1
  fi
}

expect_generator_failure "a missing input directory" --arm64-dir "$tmp/does-not-exist"
expect_generator_failure "duplicate input directories" --arm64-dir "$tmp/in-linux-amd64"
expect_generator_failure "aliased duplicate input directories" --arm64-dir "$tmp/in-linux-amd64/."
mkdir "$tmp/already-exists"
expect_generator_failure "an existing output directory" --output-dir "$tmp/already-exists"
expect_generator_failure "a malformed simulated tag" --tag ../escape
expect_generator_failure "a malformed commit" --commit v0.4.0

cp -R "$tmp/in-linux-arm64" "$tmp/in-bad-checksum"
printf '%064d  truerepublicd-linux-arm64\n' 0 >"$tmp/in-bad-checksum/CHECKSUMS.sha256"
expect_generator_failure "a mismatched binary checksum" --arm64-dir "$tmp/in-bad-checksum"

cp -R "$tmp/in-linux-amd64" "$tmp/in-bad-source"
jq '.source_ref="ffffffffffffffffffffffffffffffffffffffff"' "$tmp/in-bad-source/build-metadata.json" >"$tmp/bad-source.json"
mv "$tmp/bad-source.json" "$tmp/in-bad-source/build-metadata.json"
expect_generator_failure "a metadata source mismatch" --amd64-dir "$tmp/in-bad-source"

cp -R "$tmp/in-oci-amd64" "$tmp/in-bad-claims"
jq '.claims.signed=true' "$tmp/in-bad-claims/oci-evidence.json" >"$tmp/bad-claims.json"
mv "$tmp/bad-claims.json" "$tmp/in-bad-claims/oci-evidence.json"
expect_generator_failure "true OCI status claims" --oci-amd64-dir "$tmp/in-bad-claims"

cp -R "$tmp/in-oci-arm64" "$tmp/in-bad-report"
jq '.valid=false' "$tmp/in-bad-report/oci-evidence-report-arm64.json" >"$tmp/bad-report.json"
mv "$tmp/bad-report.json" "$tmp/in-bad-report/oci-evidence-report-arm64.json"
expect_generator_failure "an invalid OCI digest report" --oci-arm64-dir "$tmp/in-bad-report"

cp -R "$tmp/in-linux-amd64" "$tmp/in-symlink"
rm "$tmp/in-symlink/build-metadata.json"
ln -s "$tmp/in-linux-amd64/build-metadata.json" "$tmp/in-symlink/build-metadata.json"
expect_generator_failure "a symlinked metadata input" --amd64-dir "$tmp/in-symlink"

cp -R "$tmp/in-oci-amd64" "$tmp/in-extra"
printf 'unexpected\n' >"$tmp/in-extra/extra.json"
expect_generator_failure "an unexpected extra input member" --oci-amd64-dir "$tmp/in-extra"
