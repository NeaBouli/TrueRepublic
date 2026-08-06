#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

fail() {
  echo "ERROR: $*" >&2
  exit 1
}

while IFS= read -r tracked_legacy_file; do
  if [[ -e "$tracked_legacy_file" ]]; then
    fail "retired file still exists: $tracked_legacy_file"
  fi
done < <(git ls-files 'web-wallet/**')
[[ ! -e web-wallet ]] || fail "retired web-wallet tree must stay absent"

if grep -q 'node-audit-legacy-web' .github/workflows/security-scan.yml; then
  fail "informational vulnerable legacy-web audit job must not be restored"
fi

if grep -Eq '(^|[[:space:]])web-wallet([[:space:]]|$|:)' docker-compose.yml nginx/nginx.conf; then
  fail "runtime configuration must not reference the retired client"
fi

node - <<'NODE'
const fs = require("node:fs");
const status = JSON.parse(fs.readFileSync("docs/status.json", "utf8"));
const actual = status?.recovery?.legacy_clients?.["web-wallet"];
const expected = "retired_removed_on_gh112_git_history_only";
if (actual !== expected) {
  console.error("ERROR: web retirement status must be " + expected + ", got " + actual);
  process.exit(1);
}
NODE

for file in INSTALLATION.md docs/developers/integration-guide/web-wallet.md; do
  if grep -Eq 'cd[[:space:]]+web-wallet|npm (install|start)' "$file"; then
    fail "$file must not instruct users to run the retired client"
  fi
done

echo "Web wallet retirement contract: PASS"
