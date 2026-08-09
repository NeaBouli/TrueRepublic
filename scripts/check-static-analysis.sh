#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$ROOT_DIR"

STATICCHECK_BIN=${STATICCHECK_BIN:-staticcheck}
if ! command -v "$STATICCHECK_BIN" >/dev/null 2>&1; then
  echo "staticcheck is required; install the version pinned in configs/security/gates.json" >&2
  exit 1
fi

mapfile -t packages < <(./scripts/go-packages.sh --list)
if [[ ${#packages[@]} -eq 0 ]]; then
  echo "repository package selector returned no Go packages" >&2
  exit 1
fi

"$STATICCHECK_BIN" "${packages[@]}"
