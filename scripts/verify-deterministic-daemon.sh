#!/usr/bin/env bash
set -euo pipefail

fail() {
  echo "deterministic daemon verification failed: $*" >&2
  return 1
}

sha256_file() {
  local path=$1
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$path" | awk '{print $1}'
  else
    shasum -a 256 "$path" | awk '{print $1}'
  fi
}

validate_contract() {
  local contract=$1
  local target=$2
  local source_ref=$3

  if ! jq -e --arg target "$target" --arg source_ref "$source_ref" '
    .schema == "truerepublic.daemon-build/v1" and
    .binary == "truerepublicd" and
    .main_package == "." and
    .go_version == "1.26.5" and
    .cgo_enabled == "1" and
    .source_ref == {
      "kind": "git-commit",
      "pattern": "^[0-9a-f]{40}$"
    } and
    .build_flags == {
      "trimpath": true,
      "buildvcs": false,
      "mod": "readonly",
      "ldflags": [
        "-s",
        "-w",
        "-buildid=",
        "-X",
        "main.version={{source_ref}}",
        "-X",
        "main.upgradePlan=v0.4.1",
        "-linkmode=external",
        "-extldflags=-Wl,--build-id=none"
      ]
    } and
    .targets == [
      {
        "id": "linux-amd64",
        "goos": "linux",
        "goarch": "amd64",
        "ci_runner": "ubuntu-24.04",
        "runner_arch": "x86_64",
        "artifact": "truerepublicd-linux-amd64"
      },
      {
        "id": "linux-arm64",
        "goos": "linux",
        "goarch": "arm64",
        "ci_runner": "ubuntu-24.04-arm",
        "runner_arch": "aarch64",
        "artifact": "truerepublicd-linux-arm64"
      }
    ] and
    any(.targets[]; .id == $target) and
    (.source_ref.pattern as $pattern | $source_ref | test($pattern))
  ' "$contract" >/dev/null; then
    fail "contract, target, or source ref does not match the maintained v1 contract"
    return 1
  fi
}

inspect_binary() {
  local contract=$1
  local target=$2
  local binary=$3
  local expected_runner expected_file go_version go_version_output version_output

  if [[ $(uname -s) != "Linux" ]]; then
    fail "binary inspection requires a native Linux runner"
    return 1
  fi
  expected_runner=$(jq -r --arg target "$target" '.targets[] | select(.id == $target) | .runner_arch' "$contract")
  if [[ $(uname -m) != "$expected_runner" ]]; then
    fail "runner architecture $(uname -m) does not match $expected_runner"
    return 1
  fi

  expected_file="x86-64"
  [[ "$target" == "linux-arm64" ]] && expected_file="ARM aarch64"
  if ! file "$binary" | grep -Fq "ELF 64-bit"; then
    fail "$binary is not a 64-bit ELF binary"
    return 1
  fi
  if ! file "$binary" | grep -Fq "$expected_file"; then
    fail "$binary does not match $target"
    return 1
  fi

  go_version=$(jq -r '.go_version' "$contract")
  go_version_output=$(go version -m "$binary") || return 1
  if ! sed -n '1p' <<<"$go_version_output" | grep -Fq "go${go_version}"; then
    fail "$binary was not built with Go $go_version"
    return 1
  fi

  version_output=$("$binary" --version | tr -d '\r')
  printf '%s\n' "$version_output"
}

verify_pair() {
  local contract=$1
  local target=$2
  local source_ref=$3
  local first=$4
  local second=$5
  local first_hash second_hash expected_version first_version second_version

  validate_contract "$contract" "$target" "$source_ref" || return 1
  if [[ ! -f "$first" || ! -f "$second" ]]; then
    fail "both independent build attempts are required"
    return 1
  fi

  first_hash=$(sha256_file "$first")
  second_hash=$(sha256_file "$second")
  if [[ "$first_hash" != "$second_hash" ]]; then
    fail "independent build SHA-256 values differ"
    return 1
  fi

  expected_version="truerepublicd version ${source_ref}"
  first_version=$(inspect_binary "$contract" "$target" "$first") || return 1
  second_version=$(inspect_binary "$contract" "$target" "$second") || return 1
  if [[ "$first_version" != "$expected_version" ]]; then
    fail "first build version does not match the declared source ref"
    return 1
  fi
  if [[ "$second_version" != "$expected_version" ]]; then
    fail "second build version does not match the declared source ref"
    return 1
  fi

  printf '%s\n' "$first_hash"
}

main() {
  if [[ $# -ne 5 ]]; then
    echo "usage: $0 <contract> <target> <source-ref> <first-binary> <second-binary>" >&2
    exit 2
  fi
  verify_pair "$@"
}

if [[ ${BASH_SOURCE[0]} == "$0" ]]; then
  main "$@"
fi
