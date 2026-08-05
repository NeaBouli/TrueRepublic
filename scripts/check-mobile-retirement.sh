#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

fail() {
  echo "ERROR: $*" >&2
  exit 1
}

while IFS= read -r tracked_mobile_file; do
  if [[ -e "$tracked_mobile_file" ]]; then
    fail "retired file still exists: $tracked_mobile_file"
  fi
done < <(git ls-files 'mobile-wallet/**')
[[ ! -e .github/workflows/react-native-ci.yml ]] || fail "retired mobile CI workflow must stay absent"

if grep -q 'node-audit-legacy-mobile' .github/workflows/security-scan.yml; then
  fail "informational vulnerable mobile audit job must not be restored"
fi

node - <<'NODE'
const fs = require("node:fs");
const status = JSON.parse(fs.readFileSync("docs/status.json", "utf8"));
const actual = status?.recovery?.legacy_clients?.["mobile-wallet"];
const expected = "retired_removed_on_gh102_git_history_only";
if (actual !== expected) {
  console.error(`ERROR: mobile retirement status must be ${expected}, got ${actual}`);
  process.exit(1);
}
NODE

for file in INSTALLATION.md docs/developers/integration-guide/mobile.md; do
  if grep -q 'cd mobile-wallet' "$file"; then
    fail "$file must not instruct users to run the retired client"
  fi
done

echo "Mobile retirement contract: PASS"
