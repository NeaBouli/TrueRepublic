#!/usr/bin/env bash
# Adversarial fixtures for scripts/check-license-policy.sh (GH-219).
# Builds synthetic repositories in a temporary directory and proves the
# policy check passes an honest pending tree and fails closed on every
# planted contradiction. No real repository file is touched.
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
CHECK="$ROOT_DIR/scripts/check-license-policy.sh"

for tool in git jq; do
  if ! command -v "$tool" >/dev/null 2>&1; then
    echo "git and jq are required for the license policy fixtures" >&2
    exit 1
  fi
done

WORK=$(mktemp -d "${TMPDIR:-/tmp}/truerepublic-license-policy.XXXXXX")
trap 'rm -rf "$WORK"' EXIT

PATTERN='SPDX-License-Identifier|is licensed under|are licensed under|under the same license|same license as the project|licensed under the (apache|mit|gnu|bsd|mozilla)|under the apache license|under the mit license|apache license, version 2\.0|gnu affero general public license|gnu general public license|\[LICENSE\]\(LICENSE'

write_manifest() {
  local dir=$1 status=$2 selected=$3
  # JSON requires doubled backslashes for regex escapes such as `2\.0`.
  local pattern_json=${PATTERN//\\/\\\\}
  cat >"$dir/configs/legal/license-decision.json" <<EOF
{
  "schema": "truerepublic.license-decision/v1",
  "issue": "GH-219",
  "status": "$status",
  "selected_spdx_id": "$selected",
  "copyright_line": "NONE",
  "decision_record": "NONE",
  "decision_package": "docs/legal/GH-219-LICENSE-DECISION.md",
  "candidates": [
    {"id": "Apache-2.0", "kind": "permissive", "reference": "https://www.apache.org/licenses/LICENSE-2.0"},
    {"id": "AGPL-3.0-only", "kind": "network-copyleft", "reference": "https://www.gnu.org/licenses/agpl-3.0.html"},
    {"id": "dual-license", "kind": "combination", "reference": "docs/legal/GH-219-LICENSE-DECISION.md"}
  ],
  "components": [
    {"id": "daemon-go", "kind": "repository-authored", "path": "go.mod", "metadata": "go-module", "declared_license": "NONE"},
    {"id": "contracts-rust", "kind": "repository-authored", "path": "contracts", "metadata": "cargo-workspace", "declared_license": "NONE"},
    {"id": "client-web", "kind": "repository-authored", "path": "client-web/package.json", "metadata": "npm-package", "declared_license": "NONE"}
  ],
  "scan": {
    "forbidden_assertion_pattern": "$pattern_json",
    "exempt_paths": [
      "configs/legal/license-decision.json",
      "docs/legal/GH-219-LICENSE-DECISION.md",
      "scripts/check-license-policy.sh",
      "scripts/test-license-policy.sh"
    ],
    "allowed_historical_paths": [
      {"path": "RELEASE_NOTES_v0.3.0.md", "marker": "Historical draft only"}
    ]
  }
}
EOF
}

build_fixture() {
  local dir=$1
  mkdir -p "$dir"/{configs/legal,docs/legal,scripts,contracts/core,client-web}
  write_manifest "$dir" pending NONE
  cat >"$dir/docs/legal/GH-219-LICENSE-DECISION.md" <<'EOF'
Decision package for truerepublic.license-decision/v1. Decision status: pending.
EOF
  : >"$dir/scripts/check-license-policy.sh"
  : >"$dir/scripts/test-license-policy.sh"
  printf 'module truerepublic\n' >"$dir/go.mod"
  printf '[package]\nname = "truerepublic-contracts"\nversion = "0.1.0"\n' >"$dir/contracts/core/Cargo.toml"
  printf '[workspace]\nmembers = ["core"]\n' >"$dir/contracts/Cargo.toml"
  printf '{"name":"client-web","private":true,"version":"0.4.0"}\n' >"$dir/client-web/package.json"
  printf '# Historical draft only.\nSee [LICENSE](LICENSE) file.\n' >"$dir/RELEASE_NOTES_v0.3.0.md"
  printf '# Contributing\n\nThe project license decision is pending.\n' >"$dir/CONTRIBUTING.md"
  git -C "$dir" init -q
  git -C "$dir" add -A
}

expect_pass() {
  local label=$1 dir=$2
  if LICENSE_POLICY_ROOT="$dir" "$CHECK" >"$dir.out" 2>&1; then
    echo "  OK pass: $label"
  else
    echo "fixture unexpectedly failed ($label):" >&2
    cat "$dir.out" >&2
    exit 1
  fi
}

expect_fail() {
  local label=$1 dir=$2
  if LICENSE_POLICY_ROOT="$dir" "$CHECK" >"$dir.out" 2>&1; then
    echo "fixture unexpectedly passed ($label)" >&2
    cat "$dir.out" >&2
    exit 1
  fi
  echo "  OK fail-closed: $label"
}

commit_change() { git -C "$1" add -A; }

prepare_decided() {
  local dir=$1 selected=$2 decision_record=$3
  jq --arg selected "$selected" --arg decision_record "$decision_record" '
    .status = "decided" |
    .selected_spdx_id = $selected |
    .copyright_line = "Copyright 2026 TrueRepublic" |
    .decision_record = $decision_record |
    (.components[].declared_license) = $selected
  ' "$dir/configs/legal/license-decision.json" \
    >"$dir/configs/legal/license-decision.json.tmp"
  mv "$dir/configs/legal/license-decision.json.tmp" \
    "$dir/configs/legal/license-decision.json"
  printf 'Decision package for truerepublic.license-decision/v1. Decision status: decided.\n' \
    >"$dir/docs/legal/GH-219-LICENSE-DECISION.md"
  printf '{"name":"client-web","private":true,"version":"0.4.0","license":"%s"}\n' \
    "$selected" >"$dir/client-web/package.json"
  printf '[package]\nname = "truerepublic-contracts"\nversion = "0.1.0"\nlicense = "%s"\n' \
    "$selected" >"$dir/contracts/core/Cargo.toml"
}

# 1. Honest pending tree passes.
build_fixture "$WORK/honest"
expect_pass "honest pending tree" "$WORK/honest"

# 2. Tracked root LICENSE while pending fails.
build_fixture "$WORK/root-license"
printf 'Apache License\nVersion 2.0\n' >"$WORK/root-license/LICENSE"
commit_change "$WORK/root-license"
expect_fail "tracked root LICENSE while pending" "$WORK/root-license"

# 3. SPDX identifier in a source file fails.
build_fixture "$WORK/spdx"
printf '// SPDX-License-Identifier: MIT\npackage main\n' >"$WORK/spdx/main.go"
commit_change "$WORK/spdx"
expect_fail "SPDX identifier in tracked source" "$WORK/spdx"

# 4. Historical CONTRIBUTING-style Apache claim fails.
build_fixture "$WORK/contributing"
printf 'contributions will be licensed\nunder the same license as the project (Apache 2.0).\n' \
  >>"$WORK/contributing/CONTRIBUTING.md"
commit_change "$WORK/contributing"
expect_fail "non-allowlisted project-license assertion" "$WORK/contributing"

# 5. The same claim is permitted only inside the allowlisted historical file.
build_fixture "$WORK/allowlisted"
printf 'See [LICENSE](LICENSE) file for the terms.\n' >>"$WORK/allowlisted/RELEASE_NOTES_v0.3.0.md"
commit_change "$WORK/allowlisted"
expect_pass "allowlisted historical wording" "$WORK/allowlisted"

# 6. Untracked material is never scanned (build caches, node_modules, .git).
build_fixture "$WORK/untracked"
mkdir -p "$WORK/untracked/node_modules/pkg" "$WORK/untracked/build"
printf 'SPDX-License-Identifier: MIT\n' >"$WORK/untracked/node_modules/pkg/index.js"
printf 'SPDX-License-Identifier: Apache-2.0\n' >"$WORK/untracked/build/cache.txt"
printf 'SPDX-License-Identifier: GPL-3.0\n' >"$WORK/untracked/scratch.md"
expect_pass "untracked external material ignored" "$WORK/untracked"

# 7. npm package metadata license claim fails while pending.
build_fixture "$WORK/npm-license"
printf '{"name":"client-web","private":true,"version":"0.4.0","license":"Apache-2.0"}\n' \
  >"$WORK/npm-license/client-web/package.json"
commit_change "$WORK/npm-license"
expect_fail "npm license metadata while pending" "$WORK/npm-license"

# 8. Cargo crate license claim fails while pending.
build_fixture "$WORK/cargo-license"
printf '[package]\nname = "truerepublic-contracts"\nversion = "0.1.0"\nlicense = "Apache-2.0"\n' \
  >"$WORK/cargo-license/contracts/core/Cargo.toml"
commit_change "$WORK/cargo-license"
expect_fail "Cargo license metadata while pending" "$WORK/cargo-license"

# 9. Malformed manifest fails closed.
build_fixture "$WORK/malformed"
printf '{"schema":' >"$WORK/malformed/configs/legal/license-decision.json"
commit_change "$WORK/malformed"
expect_fail "malformed manifest JSON" "$WORK/malformed"

# 10. Unknown status fails closed.
build_fixture "$WORK/bad-status"
write_manifest "$WORK/bad-status" exploratory NONE
commit_change "$WORK/bad-status"
expect_fail "unknown decision status" "$WORK/bad-status"

# 11. Decided status without the recorded license artifacts fails.
build_fixture "$WORK/decided-incomplete"
write_manifest "$WORK/decided-incomplete" decided NONE
commit_change "$WORK/decided-incomplete"
expect_fail "decided status without selected license" "$WORK/decided-incomplete"

# 12. Allowlisted historical file without its quarantine marker fails.
build_fixture "$WORK/stale-historical"
printf '# Release notes\nSee [LICENSE](LICENSE) file.\n' >"$WORK/stale-historical/RELEASE_NOTES_v0.3.0.md"
commit_change "$WORK/stale-historical"
expect_fail "historical file without quarantine marker" "$WORK/stale-historical"

# 13. A component declaring a license while pending fails.
build_fixture "$WORK/component-license"
jq '(.components[0].declared_license) = "Apache-2.0"' \
  "$WORK/component-license/configs/legal/license-decision.json" \
  >"$WORK/component-license/configs/legal/license-decision.json.tmp"
mv "$WORK/component-license/configs/legal/license-decision.json.tmp" \
  "$WORK/component-license/configs/legal/license-decision.json"
commit_change "$WORK/component-license"
expect_fail "component license declaration while pending" "$WORK/component-license"

# 14. Git enumeration failure fails closed (no work tree).
build_fixture "$WORK/no-git"
rm -rf "$WORK/no-git/.git"
expect_fail "manifest without a Git work tree" "$WORK/no-git"

# 15. The manifest cannot broaden the historical exception surface.
build_fixture "$WORK/overbroad-history"
jq '.scan.allowed_historical_paths += [{"path":"docs/new-claim.md","marker":"Historical"}]' \
  "$WORK/overbroad-history/configs/legal/license-decision.json" \
  >"$WORK/overbroad-history/configs/legal/license-decision.json.tmp"
mv "$WORK/overbroad-history/configs/legal/license-decision.json.tmp" \
  "$WORK/overbroad-history/configs/legal/license-decision.json"
printf '# Historical\nSPDX-License-Identifier: MIT\n' >"$WORK/overbroad-history/docs/new-claim.md"
commit_change "$WORK/overbroad-history"
expect_fail "overbroad historical allowlist" "$WORK/overbroad-history"

# 16. The checker exemption list is fixed and cannot hide a new claim.
build_fixture "$WORK/overbroad-exempt"
jq '.scan.exempt_paths += ["README.md"]' \
  "$WORK/overbroad-exempt/configs/legal/license-decision.json" \
  >"$WORK/overbroad-exempt/configs/legal/license-decision.json.tmp"
mv "$WORK/overbroad-exempt/configs/legal/license-decision.json.tmp" \
  "$WORK/overbroad-exempt/configs/legal/license-decision.json"
printf '# Project\nSPDX-License-Identifier: MIT\n' >"$WORK/overbroad-exempt/README.md"
commit_change "$WORK/overbroad-exempt"
expect_fail "overbroad checker exemption" "$WORK/overbroad-exempt"

# 17. The contradiction pattern is immutable and cannot be weakened.
build_fixture "$WORK/weakened-pattern"
jq '.scan.forbidden_assertion_pattern = "zzz-no-match"' \
  "$WORK/weakened-pattern/configs/legal/license-decision.json" \
  >"$WORK/weakened-pattern/configs/legal/license-decision.json.tmp"
mv "$WORK/weakened-pattern/configs/legal/license-decision.json.tmp" \
  "$WORK/weakened-pattern/configs/legal/license-decision.json"
printf '// SPDX-License-Identifier: MIT\npackage main\n' >"$WORK/weakened-pattern/main.go"
commit_change "$WORK/weakened-pattern"
expect_fail "weakened contradiction pattern" "$WORK/weakened-pattern"

# 18. An invalid contradiction pattern cannot turn scan errors into success.
build_fixture "$WORK/invalid-pattern"
jq '.scan.forbidden_assertion_pattern = "(["' \
  "$WORK/invalid-pattern/configs/legal/license-decision.json" \
  >"$WORK/invalid-pattern/configs/legal/license-decision.json.tmp"
mv "$WORK/invalid-pattern/configs/legal/license-decision.json.tmp" \
  "$WORK/invalid-pattern/configs/legal/license-decision.json"
printf '// SPDX-License-Identifier: MIT\npackage main\n' >"$WORK/invalid-pattern/main.go"
commit_change "$WORK/invalid-pattern"
expect_fail "invalid contradiction pattern" "$WORK/invalid-pattern"

# 19. Nested license artifacts are forbidden while the decision is pending.
build_fixture "$WORK/nested-license"
mkdir -p "$WORK/nested-license/docs/sub"
printf 'MIT License\n\nPermission is hereby granted, free of charge.\n' \
  >"$WORK/nested-license/docs/sub/LICENSE"
commit_change "$WORK/nested-license"
expect_fail "nested license artifact while pending" "$WORK/nested-license"

# 20. A complete single-license decided state passes the dormant branch.
build_fixture "$WORK/decided-apache"
prepare_decided "$WORK/decided-apache" Apache-2.0 \
  'https://github.com/NeaBouli/TrueRepublic/issues/219#issuecomment-123'
printf 'Apache License\nVersion 2.0\n' >"$WORK/decided-apache/LICENSE"
commit_change "$WORK/decided-apache"
expect_pass "complete Apache-2.0 decided state" "$WORK/decided-apache"

# 21. The AGPL-3.0-only decided state is independently exercised.
build_fixture "$WORK/decided-agpl"
prepare_decided "$WORK/decided-agpl" AGPL-3.0-only \
  'https://github.com/NeaBouli/TrueRepublic/issues/219#issuecomment-456'
printf 'GNU AFFERO GENERAL PUBLIC LICENSE\nVersion 3\n' >"$WORK/decided-agpl/LICENSE"
commit_change "$WORK/decided-agpl"
expect_pass "complete AGPL-3.0-only decided state" "$WORK/decided-agpl"

# 22. NOTICE alone cannot impersonate the required root LICENSE.
build_fixture "$WORK/decided-notice-only"
prepare_decided "$WORK/decided-notice-only" Apache-2.0 \
  'https://github.com/NeaBouli/TrueRepublic/issues/219#issuecomment-789'
printf 'TrueRepublic notice\n' >"$WORK/decided-notice-only/NOTICE"
commit_change "$WORK/decided-notice-only"
expect_fail "decided state with NOTICE but no LICENSE" "$WORK/decided-notice-only"

# 23. The owner decision URL must include a numeric comment identifier.
build_fixture "$WORK/decided-empty-comment"
prepare_decided "$WORK/decided-empty-comment" Apache-2.0 \
  'https://github.com/NeaBouli/TrueRepublic/issues/219#issuecomment-'
printf 'Apache License\nVersion 2.0\n' >"$WORK/decided-empty-comment/LICENSE"
commit_change "$WORK/decided-empty-comment"
expect_fail "decided state without numeric owner comment id" "$WORK/decided-empty-comment"

# 24. The canonical root text must match the selected SPDX identity.
build_fixture "$WORK/decided-wrong-text"
prepare_decided "$WORK/decided-wrong-text" Apache-2.0 \
  'https://github.com/NeaBouli/TrueRepublic/issues/219#issuecomment-1011'
printf 'GNU AFFERO GENERAL PUBLIC LICENSE\nVersion 3\n' >"$WORK/decided-wrong-text/LICENSE"
commit_change "$WORK/decided-wrong-text"
expect_fail "decided state with mismatched root license text" "$WORK/decided-wrong-text"

# 25. Case variants cannot hide a license artifact while pending.
build_fixture "$WORK/lowercase-license"
printf 'MIT License\n\nPermission is hereby granted, free of charge.\n' \
  >"$WORK/lowercase-license/license"
commit_change "$WORK/lowercase-license"
expect_fail "lowercase license artifact while pending" "$WORK/lowercase-license"

# 26. A decided state cannot hide a conflicting nested license artifact.
build_fixture "$WORK/decided-nested-conflict"
prepare_decided "$WORK/decided-nested-conflict" Apache-2.0 \
  'https://github.com/NeaBouli/TrueRepublic/issues/219#issuecomment-1213'
printf 'Apache License\nVersion 2.0\n' >"$WORK/decided-nested-conflict/LICENSE"
mkdir -p "$WORK/decided-nested-conflict/docs/sub"
printf 'SPDX-License-Identifier: MIT\n' >"$WORK/decided-nested-conflict/docs/sub/LICENSE"
commit_change "$WORK/decided-nested-conflict"
expect_fail "decided state with conflicting nested license" "$WORK/decided-nested-conflict"

echo "license policy positive and negative fixtures passed"
