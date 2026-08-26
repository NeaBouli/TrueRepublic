#!/usr/bin/env bash
# License decision policy check (GH-219).
# Deterministic, tracked-files-only repository gate for the project license
# decision recorded in configs/legal/license-decision.json. While the
# decision status is pending this fails closed on any tracked license artifact,
# SPDX claim, package-metadata license, or non-allowlisted public license
# assertion. Scanning uses git grep/git ls-files, so build caches, .git,
# node_modules and other untracked material are never inspected.
set -euo pipefail

ROOT_DIR=${LICENSE_POLICY_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}
MANIFEST="$ROOT_DIR/configs/legal/license-decision.json"
SCHEMA="truerepublic.license-decision/v1"
EXPECTED_DECISION_RECORD="https://github.com/NeaBouli/TrueRepublic/issues/219#issuecomment-5423337355"
EXPECTED_PATTERN='SPDX-License-Identifier|license[[:space:]]*:[[:space:]]*(Apache-2.0|AGPL-3.0-only|MIT|BSD-[0-9]-Clause|GPL-[0-9.]+|MPL-[0-9.]+)|is licensed under|are licensed under|under the same license|same license as the project|licensed under the (apache|mit|gnu|bsd|mozilla)|under the apache license|under the mit license|apache license, version 2\.0|gnu affero general public license|gnu general public license|\[LICENSE\]\(LICENSE'

for tool in git jq; do
  if ! command -v "$tool" >/dev/null 2>&1; then
    echo "FAIL: $tool is required for the license policy check" >&2
    exit 1
  fi
done

ERRORS=0
fail() { echo "FAIL: $1"; ERRORS=$((ERRORS + 1)); }
ok() { echo "  OK $1"; }

if ! git -C "$ROOT_DIR" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "FAIL: $ROOT_DIR is not a Git work tree; tracked-file enumeration is mandatory" >&2
  exit 1
fi

if [ ! -f "$MANIFEST" ]; then
  echo "FAIL: license decision manifest $MANIFEST not found" >&2
  exit 1
fi
if ! jq empty "$MANIFEST" 2>/dev/null; then
  echo "FAIL: license decision manifest is not well-formed JSON" >&2
  exit 1
fi
if ! jq -e --arg expected_pattern "$EXPECTED_PATTERN" '
  (type == "object") and
  (.schema | type == "string") and
  (.issue | type == "string") and
  (.status | type == "string") and
  (.selected_spdx_id | type == "string") and
  (.copyright_line | type == "string") and
  (.decision_record | type == "string") and
  (.decision_package == "docs/legal/GH-219-LICENSE-DECISION.md") and
  (.scope | type == "object") and
  (.scope.covers | type == "string") and
  (.scope.reuse_file | type == "string") and
  (.scope.excludes | type == "array") and
  (.candidates | type == "array" and length == 3) and
  ([.candidates[].id] | sort == ["AGPL-3.0-only", "Apache-2.0", "dual-license"]) and
  (all(.candidates[];
    (.kind | type == "string" and length > 0) and
    (.reference | type == "string" and length > 0))) and
  (.components | type == "array" and length >= 3) and
  (all(.components[];
    (.id | type == "string" and length > 0) and
    (.kind | type == "string" and length > 0) and
    (.path | type == "string" and length > 0) and
    (.metadata | type == "string" and length > 0) and
    (.declared_license | type == "string" and length > 0) and
    (.coverage == "maintained" or .coverage == "excluded"))) and
  (([.components[].id] | length) == ([.components[].id] | unique | length)) and
  (([.components[].path] | length) == ([.components[].path] | unique | length)) and
  (.scan | type == "object") and
  (.scan.forbidden_assertion_pattern == $expected_pattern) and
  ([.scan.exempt_paths[]] | sort == [
    "REUSE.toml",
    "configs/legal/license-decision.json",
    "docs/legal/GH-219-LICENSE-DECISION.md",
    "scripts/check-license-policy.sh",
    "scripts/test-license-policy.sh"
  ]) and
  (.scan.allowed_historical_paths | type == "array" and length > 0) and
  (all(.scan.allowed_historical_paths[];
    ((.path == "RELEASE_NOTES_v0.3.0.md") or
     (.path == "docs/archive/releases/v0.3.0-draft-original-release-notes.md")) and
    (.marker | type == "string" and length > 0))) and
  (([.scan.allowed_historical_paths[].path] | length) ==
   ([.scan.allowed_historical_paths[].path] | unique | length))
' "$MANIFEST" >/dev/null 2>&1; then
  echo "FAIL: license decision manifest does not satisfy the closed schema" >&2
  exit 1
fi

[ "$(jq -r '.schema' "$MANIFEST")" = "$SCHEMA" ] && ok "manifest schema" || fail "manifest schema must be $SCHEMA"
[ "$(jq -r '.issue' "$MANIFEST")" = "GH-219" ] && ok "issue binding" || fail "manifest must be bound to GH-219"

STATUS=$(jq -r '.status' "$MANIFEST")
case "$STATUS" in
  pending|decided) ok "status $STATUS" ;;
  *) fail "status must be pending or decided, got: $STATUS" ;;
esac

SELECTED=$(jq -r '.selected_spdx_id' "$MANIFEST")
COPYRIGHT_LINE=$(jq -r '.copyright_line' "$MANIFEST")
DECISION_RECORD=$(jq -r '.decision_record' "$MANIFEST")

for candidate in Apache-2.0 AGPL-3.0-only dual-license; do
  jq -e --arg c "$candidate" '[.candidates[].id] | index($c) != null' "$MANIFEST" >/dev/null &&
    ok "candidate $candidate documented" || fail "candidate $candidate missing from manifest"
done

DECISION_PACKAGE="$ROOT_DIR/$(jq -r '.decision_package' "$MANIFEST")"
if [ -f "$DECISION_PACKAGE" ]; then
  ok "decision package exists"
  grep -Fq "$SCHEMA" "$DECISION_PACKAGE" && ok "decision package schema binding" ||
    fail "decision package must reference $SCHEMA"
  if [ "$STATUS" = "pending" ]; then
    grep -Fqi "pending" "$DECISION_PACKAGE" && ok "decision package pending state" ||
      fail "decision package must record the pending decision state"
  else
    grep -Fqi "decided" "$DECISION_PACKAGE" && ok "decision package decided state" ||
      fail "decision package must record the decided state"
  fi
else
  fail "decision package $DECISION_PACKAGE not found"
fi

# Exempt paths (checker, manifest, decision package) must exist so the
# exemption list cannot silently cover removed or renamed files.
while IFS= read -r exempt; do
  [ -n "$exempt" ] || continue
  [ -e "$ROOT_DIR/$exempt" ] || fail "exempt path $exempt does not exist"
done < <(jq -r '.scan.exempt_paths[]' "$MANIFEST")
[ "$ERRORS" -eq 0 ] && ok "exempt paths present"

# Explicitly allowlisted historical files must exist and carry their
# quarantine marker; historical wording is permitted nowhere else.
while IFS= read -r entry; do
  [ -n "$entry" ] || continue
  hist_path=${entry%%:*}
  hist_marker=${entry#*:}
  if [ ! -f "$ROOT_DIR/$hist_path" ]; then
    fail "allowlisted historical file $hist_path missing"
    continue
  fi
  grep -Fq "$hist_marker" "$ROOT_DIR/$hist_path" ||
    fail "allowlisted historical file $hist_path lost its quarantine marker"
done < <(jq -r '.scan.allowed_historical_paths[] | .path + ":" + .marker' "$MANIFEST")
[ "$ERRORS" -eq 0 ] && ok "historical allowlist intact"

LICENSE_ARTIFACTS=$(git -C "$ROOT_DIR" ls-files |
  grep -iE '(^|/)((LICENSE|LICENCE)[^/]*|(COPYING|NOTICE|COPYRIGHT)(\.(md|txt))?)$' |
  grep -v '^configs/legal/license-decision.json$' || true)

# Maintained-component metadata inventory.
npm_license=$(jq -r '.license // "NONE"' "$ROOT_DIR/client-web/package.json" 2>/dev/null || echo "UNREADABLE")
npm_private=$(jq -r '.private' "$ROOT_DIR/client-web/package.json" 2>/dev/null || echo "UNREADABLE")
CARGO_MANIFESTS=$(git -C "$ROOT_DIR" ls-files |
  grep -E '^contracts/(.*/)?Cargo\.toml$' || true)
cargo_with_license=""
while IFS= read -r cargo_manifest; do
  [ -n "$cargo_manifest" ] || continue
  if grep -Eq '^[[:space:]]*license[[:space:]]*=' "$ROOT_DIR/$cargo_manifest"; then
    cargo_with_license="${cargo_with_license}${cargo_with_license:+$'\n'}$cargo_manifest"
  fi
done <<<"$CARGO_MANIFESTS"

if [ "$STATUS" = "pending" ]; then
  [ "$SELECTED" = "NONE" ] && ok "no selected license while pending" ||
    fail "selected_spdx_id must be NONE while pending, got: $SELECTED"
  [ "$COPYRIGHT_LINE" = "NONE" ] && ok "no copyright line while pending" ||
    fail "copyright_line must be NONE while pending"
  [ "$DECISION_RECORD" = "NONE" ] && ok "no governance decision record while pending" ||
    fail "decision_record must be NONE while pending"
  if [ -z "$LICENSE_ARTIFACTS" ]; then
    ok "no LICENSE/COPYING/NOTICE artifact while pending"
  else
    fail "license or notice artifact tracked while pending: $LICENSE_ARTIFACTS"
  fi
  while IFS= read -r component; do
    component_id=${component%%=*}
    component_license=${component#*=}
    [ "$component_license" = "NONE" ] ||
      fail "component $component_id declares a license while pending: $component_license"
  done < <(jq -r '.components[] | .id + "=" + .declared_license' "$MANIFEST")
  [ "$ERRORS" -eq 0 ] && ok "component inventory declares no license"
  [ "$npm_license" = "NONE" ] && [ "$npm_private" = "true" ] &&
    ok "client-web package metadata has no license claim" ||
    fail "client-web/package.json must stay private with no license field while pending (license=$npm_license private=$npm_private)"
  if [ -z "$cargo_with_license" ]; then
    ok "no Cargo crate declares a license while pending"
  else
    fail "Cargo manifest declares a license while pending: $cargo_with_license"
  fi
else
  case "$SELECTED" in
    Apache-2.0|AGPL-3.0-only) ok "supported selected license recorded" ;;
    *) fail "decided status requires Apache-2.0 or AGPL-3.0-only; extend the contract before a dual-license decision" ;;
  esac
  [ "$COPYRIGHT_LINE" = "Copyright 2026 TrueRepublic contributors" ] &&
    ok "community contributor attribution recorded" ||
    fail "decided status requires the collective TrueRepublic contributors attribution"
  if [ "$DECISION_RECORD" = "$EXPECTED_DECISION_RECORD" ]; then
      ok "governance decision record bound to the exact GH-219 comment"
  else
    fail "decided status requires the exact approved GH-219 governance comment URL"
  fi
  if [ "$(jq -r '.scope.covers' "$MANIFEST")" = "maintained-source-and-documentation" ] &&
    [ "$(jq -r '.scope.reuse_file' "$MANIFEST")" = "REUSE.toml" ] &&
    jq -e '.scope.excludes == [
      "brand-assets",
      "artwork",
      "historical-pdfs",
      "archived-historical-evidence",
      "third-party-materials"
    ]' "$MANIFEST" >/dev/null; then
    ok "approved scope and provenance exclusions recorded"
  else
    fail "decided status requires the exact maintained-code/docs scope and exclusions"
  fi
  if [ -s "$ROOT_DIR/REUSE.toml" ] &&
    grep -Fq 'SPDX-FileCopyrightText = "Copyright 2026 TrueRepublic contributors"' "$ROOT_DIR/REUSE.toml" &&
    grep -Fq "SPDX-License-Identifier = \"$SELECTED\"" "$ROOT_DIR/REUSE.toml" &&
    ! grep -Eq '"(assets|docs/assets|docs/archive|ui)/\*\*"' "$ROOT_DIR/REUSE.toml" &&
    ! grep -Fq '"*.pdf"' "$ROOT_DIR/REUSE.toml" &&
    ! grep -Fq '"RELEASE_NOTES_v0.3.0.md"' "$ROOT_DIR/REUSE.toml" &&
    ! grep -Fq '"client-web/CHANGELOG.md"' "$ROOT_DIR/REUSE.toml"; then
    ok "REUSE-compatible maintained scope excludes provenance-gated material"
  else
    fail "REUSE.toml must record Apache-2.0 without covering excluded material"
  fi
  if git -C "$ROOT_DIR" ls-files --error-unmatch LICENSE >/dev/null 2>&1 &&
    [ -s "$ROOT_DIR/LICENSE" ]; then
    ok "non-empty root LICENSE tracked"
    case "$SELECTED" in
      Apache-2.0)
        grep -Fq 'Apache License' "$ROOT_DIR/LICENSE" &&
          grep -Fq 'Version 2.0' "$ROOT_DIR/LICENSE" ||
          fail "root LICENSE does not identify Apache-2.0"
        if command -v shasum >/dev/null 2>&1; then
          apache_hash=$(shasum -a 256 "$ROOT_DIR/LICENSE" | awk '{print $1}')
        elif command -v sha256sum >/dev/null 2>&1; then
          apache_hash=$(sha256sum "$ROOT_DIR/LICENSE" | awk '{print $1}')
        else
          apache_hash="UNAVAILABLE"
        fi
        [ "$apache_hash" = "cfc7749b96f63bd31c3c42b5c471bf756814053e847c10f3eb003417bc523d30" ] &&
          ok "canonical Apache-2.0 text hash" ||
          fail "root LICENSE is not the exact canonical Apache-2.0 text"
        ;;
      AGPL-3.0-only)
        grep -Fq 'GNU AFFERO GENERAL PUBLIC LICENSE' "$ROOT_DIR/LICENSE" &&
          grep -Fq 'Version 3' "$ROOT_DIR/LICENSE" ||
          fail "root LICENSE does not identify AGPL-3.0-only"
        ;;
    esac
  else
    fail "decided status requires a tracked, non-empty root LICENSE"
  fi
  if git -C "$ROOT_DIR" ls-files --error-unmatch NOTICE >/dev/null 2>&1 &&
    [ -s "$ROOT_DIR/NOTICE" ] &&
    grep -Fq "$COPYRIGHT_LINE" "$ROOT_DIR/NOTICE" &&
    grep -Fq "without a central corporate owner" "$ROOT_DIR/NOTICE" &&
    grep -Fq "Brand assets, artwork, historical PDFs" "$ROOT_DIR/NOTICE"; then
    ok "community NOTICE and exclusions tracked"
  else
    fail "decided status requires a tracked NOTICE with attribution and exclusions"
  fi
  while IFS='=' read -r component_kind component_license; do
    if [ "$component_kind" = "repository-authored" ]; then
      [ "$component_license" = "$SELECTED" ] ||
        fail "repository-authored component must match $SELECTED (got $component_license)"
    else
      [ "$component_license" = "NONE" ] ||
        fail "excluded/non-maintained component must remain NONE (got $component_license)"
    fi
  done < <(jq -r '.components[] | .kind + "=" + .declared_license' "$MANIFEST")
  [ "$npm_license" = "$SELECTED" ] && ok "client-web metadata matches decision" ||
    fail "client-web/package.json license must equal the decided SPDX id (got $npm_license)"
  cargo_packages=""
  while IFS= read -r cargo_manifest; do
    [ -n "$cargo_manifest" ] || continue
    if grep -Eq '^\[package\][[:space:]]*$' "$ROOT_DIR/$cargo_manifest"; then
      cargo_packages="${cargo_packages}${cargo_packages:+$'\n'}$cargo_manifest"
    fi
  done <<<"$CARGO_MANIFESTS"
  cargo_total=$(printf '%s\n' "$cargo_packages" | grep -c . || true)
  cargo_matching=0
  while IFS= read -r cargo_manifest; do
    [ -n "$cargo_manifest" ] || continue
    if grep -Eq "^[[:space:]]*license[[:space:]]*=[[:space:]]*\"$SELECTED\"[[:space:]]*$" \
      "$ROOT_DIR/$cargo_manifest"; then
      cargo_matching=$((cargo_matching + 1))
    fi
  done <<<"$cargo_packages"
  [ "$cargo_total" -gt 0 ] && [ "$cargo_matching" -eq "$cargo_total" ] &&
    ok "all Cargo packages match the decision" ||
    fail "decided status requires $SELECTED in all $cargo_total Cargo package manifests (matched $cargo_matching)"
fi

# Component paths must exist so the inventory cannot drift.
while IFS= read -r component_path; do
  [ -n "$component_path" ] || continue
  [ -e "$ROOT_DIR/$component_path" ] || fail "component path $component_path missing"
done < <(jq -r '.components[].path' "$MANIFEST")
[ "$ERRORS" -eq 0 ] && ok "component paths present"

# Contradiction scan over tracked files only.
PATTERN=$EXPECTED_PATTERN
EXEMPT_GREP=$(jq -r '.scan.exempt_paths[]' "$MANIFEST" | sed 's/[].[^$*\/|]/\\&/g' | paste -sd'|' -)
HIST_GREP=$(jq -r '.scan.allowed_historical_paths[].path' "$MANIFEST" | sed 's/[].[^$*\/|]/\\&/g' | paste -sd'|' -)
if MATCHES=$(git -C "$ROOT_DIR" grep -IlEi -e "$PATTERN" -- .); then
  :
else
  grep_status=$?
  if [ "$grep_status" -eq 1 ]; then
    MATCHES=""
  else
    fail "tracked-file license assertion scan failed with git grep exit $grep_status"
    MATCHES=""
  fi
fi
CONTRADICTIONS=$(printf '%s\n' "$MATCHES" | grep -Ev "^($EXEMPT_GREP|$HIST_GREP)$" | grep -v '^$' || true)
if [ "$STATUS" = "decided" ]; then
  ROOT_DECIDED_ARTIFACTS=$(printf '%s\n' "$LICENSE_ARTIFACTS" |
    grep -E '^(LICENSE|NOTICE)$' || true)
  while IFS= read -r root_artifact; do
    [ -n "$root_artifact" ] || continue
    while IFS= read -r spdx_line; do
      [ -n "$spdx_line" ] || continue
      spdx_value=$(printf '%s' "${spdx_line#*:}" |
        sed 's/^[[:space:]]*//; s/[[:space:]]*$//')
      [ "$spdx_value" = "$SELECTED" ] ||
        fail "$root_artifact contains conflicting SPDX identity: $spdx_value"
    done < <(grep -iE 'SPDX-License-Identifier[[:space:]]*:' \
      "$ROOT_DIR/$root_artifact" 2>/dev/null || true)
    case "$SELECTED" in
      Apache-2.0)
        grep -Eiq 'MIT License|GNU (AFFERO )?GENERAL PUBLIC LICENSE|Mozilla Public License' \
          "$ROOT_DIR/$root_artifact" &&
          fail "$root_artifact contains text for a conflicting license" || true
        ;;
      AGPL-3.0-only)
        grep -Eiq 'Apache License|MIT License|Mozilla Public License' \
          "$ROOT_DIR/$root_artifact" &&
          fail "$root_artifact contains text for a conflicting license" || true
        ;;
    esac
  done <<<"$ROOT_DECIDED_ARTIFACTS"
  LICENSE_GREP=$(printf '%s\n' "$ROOT_DECIDED_ARTIFACTS" |
    sed 's/[].[^$*\/|]/\\&/g' | paste -sd'|' -)
  CONTRADICTIONS=$(printf '%s\n' "$CONTRADICTIONS" | grep -Ev "^($LICENSE_GREP)$" | grep -v '^$' || true)
fi
if [ -z "$CONTRADICTIONS" ]; then
  ok "no non-allowlisted license assertion in tracked files"
else
  while IFS= read -r contradiction; do
    fail "non-allowlisted license assertion in tracked file: $contradiction"
  done <<<"$CONTRADICTIONS"
fi

if [ "$ERRORS" -gt 0 ]; then
  echo "FAILED: $ERRORS license policy violation(s)"
  exit 1
fi
echo "PASSED: license decision policy ($STATUS)"
