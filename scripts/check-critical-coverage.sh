#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(git -C "$(dirname "${BASH_SOURCE[0]}")/.." rev-parse --show-toplevel)
CONTRACT="$ROOT_DIR/configs/quality/critical-coverage.tsv"
cd "$ROOT_DIR"

if [[ ! -f "$CONTRACT" ]]; then
  echo "critical coverage contract is missing: $CONTRACT" >&2
  exit 1
fi

profile_dir=$(mktemp -d "${TMPDIR:-/tmp}/truerepublic-critical-coverage.XXXXXX")
cleanup() {
  rm -f "$profile_dir"/*.cover
  rmdir "$profile_dir"
}
trap cleanup EXIT

seen='|'
count=0

while IFS=$'\t' read -r label package minimum extra; do
  [[ -z "$label" || "$label" == \#* ]] && continue
  if [[ -n "${extra:-}" || ! "$minimum" =~ ^[0-9]+([.][0-9]+)?$ ]]; then
    echo "invalid critical coverage contract row for $label" >&2
    exit 1
  fi

  case "$package" in
    .) profile_name=root ;;
    ./x/dex) profile_name=dex ;;
    ./x/truedemocracy) profile_name=governance ;;
    *)
      echo "unsupported critical coverage package: $package" >&2
      exit 1
      ;;
  esac

  if [[ "$seen" == *"|$package|"* ]]; then
    echo "duplicate critical coverage package: $package" >&2
    exit 1
  fi
  seen+="$package|"
  count=$((count + 1))

  profile="$profile_dir/$profile_name.cover"
  CGO_ENABLED=1 go test -covermode=atomic -coverprofile="$profile" \
    -count=1 -timeout=600s "$package"
  actual=$(go tool cover -func="$profile" | awk '$1 == "total:" { gsub(/%/, "", $3); print $3 }')
  if [[ -z "$actual" ]]; then
    echo "coverage total missing for $label ($package)" >&2
    exit 1
  fi

  if ! awk -v actual="$actual" -v minimum="$minimum" \
    'BEGIN { exit !(actual + 0.000001 >= minimum) }'; then
    echo "$label coverage $actual% is below the required $minimum%" >&2
    exit 1
  fi
  echo "$label coverage contract passed: $actual% >= $minimum%"
done < "$CONTRACT"

if [[ $count -ne 3 ]]; then
  echo "critical coverage contract must define exactly three packages, found $count" >&2
  exit 1
fi
