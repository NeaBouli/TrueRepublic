#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$ROOT_DIR"

GITLEAKS_BIN=${GITLEAKS_BIN:-gitleaks}
if ! command -v "$GITLEAKS_BIN" >/dev/null 2>&1; then
  echo "gitleaks is required; install the version pinned in configs/security/gates.json" >&2
  exit 1
fi

scan_root=$(mktemp -d "${TMPDIR:-/tmp}/truerepublic-secret-scan.XXXXXX")
manifest=$(mktemp "${TMPDIR:-/tmp}/truerepublic-secret-files.XXXXXX")
trap 'rm -rf "$scan_root"; rm -f "$manifest"' EXIT

# Scan exactly the maintained tree: tracked files plus non-ignored files that
# are part of a local change. Build outputs, dependencies, and other ignored
# artifacts must neither hide a leak nor create an unactionable false positive.
git ls-files -z --cached --others --exclude-standard >"$manifest"

file_count=0
while IFS= read -r -d '' path; do
  if [[ -L "$path" ]]; then
    echo "Tracked symlinks are not supported by the maintained-tree secret scan: $path" >&2
    exit 1
  fi
  [[ -f "$path" ]] || continue
  mkdir -p "$scan_root/$(dirname "$path")"
  cp "$path" "$scan_root/$path"
  ((file_count += 1))
done <"$manifest"

if [[ $file_count -eq 0 ]]; then
  echo "Maintained-tree enumeration returned no files" >&2
  exit 1
fi

"$GITLEAKS_BIN" dir --config "$ROOT_DIR/.gitleaks.toml" --redact --no-banner --exit-code 1 "$scan_root"
