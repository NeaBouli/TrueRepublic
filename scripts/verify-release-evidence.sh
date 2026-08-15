#!/usr/bin/env bash
set -euo pipefail
export LC_ALL=C LANG=C

ROOT_DIR=$(git -C "$(dirname "${BASH_SOURCE[0]}")/.." rev-parse --show-toplevel)
usage() { echo "usage: $0 --bundle <dir> [--build-contract <file>] [--tool-contract <file>]" >&2; }
BUNDLE=
BUILD_CONTRACT="$ROOT_DIR/configs/build/deterministic-linux-daemon.json"
TOOL_CONTRACT="$ROOT_DIR/configs/release/tool-platform.json"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --bundle) BUNDLE=${2-}; shift 2 ;;
    --build-contract) BUILD_CONTRACT=${2-}; shift 2 ;;
    --tool-contract) TOOL_CONTRACT=${2-}; shift 2 ;;
    *) usage; exit 2 ;;
  esac
done
[[ -n "$BUNDLE" ]] || { usage; exit 2; }
cd "$ROOT_DIR"
export GOTOOLCHAIN=local
export GOPROXY=off
export GOFLAGS=-mod=readonly
exec go run ./cmd/release-evidence verify --bundle "$BUNDLE" --build-contract "$BUILD_CONTRACT" --tool-contract "$TOOL_CONTRACT"
