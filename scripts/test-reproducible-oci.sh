#!/usr/bin/env bash
set -euo pipefail
export LC_ALL=C LANG=C

ROOT_DIR=$(git -C "$(dirname "${BASH_SOURCE[0]}")/.." rev-parse --show-toplevel)
cd "$ROOT_DIR"
export GOTOOLCHAIN=local
export GOPROXY=off
export GOFLAGS=-mod=readonly

go test ./ocievidence ./cmd/oci-evidence -count=1

if ./scripts/verify-reproducible-oci.sh --unknown value >/dev/null 2>&1; then
  echo "OCI verifier accepted an unknown flag" >&2
  exit 1
fi
if ./scripts/verify-reproducible-oci.sh --evidence >/dev/null 2>&1; then
  echo "OCI verifier accepted a flag without a value" >&2
  exit 1
fi
if ./scripts/verify-reproducible-oci.sh --evidence "$ROOT_DIR/does-not-exist" >/dev/null 2>&1; then
  echo "OCI verifier accepted a missing evidence directory" >&2
  exit 1
fi
