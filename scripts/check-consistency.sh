#!/bin/bash
# Consistency Check Script
# Verifies version/tests/supply across all docs

set -euo pipefail

echo "Checking Documentation Consistency..."

# Load status.json as source of truth
STATUS_FILE="docs/status.json"

if [ ! -f "$STATUS_FILE" ]; then
  echo "ERROR: docs/status.json not found!"
  exit 1
fi

if ! jq -e '
  def nonnegative_integer:
    if type == "number" then (. >= 0 and floor == .) else false end;
  def positive_integer:
    if type == "number" then (. > 0 and floor == .) else false end;
  .rollout as $rollout |
  ($rollout.completed | nonnegative_integer) and
  ($rollout.total | positive_integer) and
  ($rollout.phase_work_completed | nonnegative_integer) and
  ($rollout.phase_work_total | positive_integer) and
  ($rollout.phase_6_completed | nonnegative_integer) and
  ($rollout.phase_6_total | positive_integer) and
  ($rollout.phase_7_completed | nonnegative_integer) and
  ($rollout.phase_7_total | positive_integer)
' "$STATUS_FILE" >/dev/null; then
  echo "FAIL Rollout counts must be nonnegative integers and totals must be positive integers"
  exit 1
fi

if ! jq -e '
  [.web_client.components, .web_client.routes, .web_client.stores, .web_client.services] |
  all(type == "number" and . > 0 and floor == .)
' "$STATUS_FILE" >/dev/null; then
  echo "FAIL Maintained-client inventory counts must be positive integers"
  exit 1
fi

VERSION=$(jq -r '.version' "$STATUS_FILE")
TESTS=$(jq -r '.tests.total' "$STATUS_FILE")
SUPPLY=$(jq -r '.token.max_supply' "$STATUS_FILE")
GO_TESTS=$(jq -r '.tests.go' "$STATUS_FILE")
RUST_TESTS=$(jq -r '.tests.rust' "$STATUS_FILE")
FRONTEND_TESTS=$(jq -r '.tests.frontend' "$STATUS_FILE")
MODULE_TESTS=$(jq '[.modules[]] | add' "$STATUS_FILE")
BASE_CAP=$(jq -r '.token.max_supply_base_units' "$STATUS_FILE")
DECIMALS=$(jq -r '.token.decimals' "$STATUS_FILE")
ROLLOUT_COMPLETED=$(jq -r '.rollout.completed' "$STATUS_FILE")
ROLLOUT_TOTAL=$(jq -r '.rollout.total' "$STATUS_FILE")
ROLLOUT_PHASE_WORK_COMPLETED=$(jq -r '.rollout.phase_work_completed' "$STATUS_FILE")
ROLLOUT_PHASE_WORK_TOTAL=$(jq -r '.rollout.phase_work_total' "$STATUS_FILE")
ROLLOUT_PHASE_6_COMPLETED=$(jq -r '.rollout.phase_6_completed' "$STATUS_FILE")
ROLLOUT_PHASE_6_TOTAL=$(jq -r '.rollout.phase_6_total' "$STATUS_FILE")
ROLLOUT_PHASE_7_COMPLETED=$(jq -r '.rollout.phase_7_completed' "$STATUS_FILE")
ROLLOUT_PHASE_7_TOTAL=$(jq -r '.rollout.phase_7_total' "$STATUS_FILE")
COSMOS_SDK_VERSION=$(jq -r '.tech.cosmos_sdk' "$STATUS_FILE")
COMETBFT_VERSION=$(jq -r '.tech.cometbft' "$STATUS_FILE")
WASMD_VERSION=$(jq -r '.tech.cosmwasm' "$STATUS_FILE")
WASMVM_VERSION=$(jq -r '.tech.wasmvm' "$STATUS_FILE")
IBC_GO_VERSION=$(jq -r '.tech.ibc_go' "$STATUS_FILE")
VITE_VERSION=$(jq -r '.tech.vite' "$STATUS_FILE")
VITE_SERIES=${VITE_VERSION%.*}
ROLLOUT_PERCENT=$(((ROLLOUT_COMPLETED * 100 + ROLLOUT_TOTAL / 2) / ROLLOUT_TOTAL))
ROLLOUT_PHASE_WORK_PERCENT=$(((ROLLOUT_PHASE_WORK_COMPLETED * 100 + ROLLOUT_PHASE_WORK_TOTAL / 2) / ROLLOUT_PHASE_WORK_TOTAL))
ROLLOUT_PHASE_6_PERCENT=$(((ROLLOUT_PHASE_6_COMPLETED * 100 + ROLLOUT_PHASE_6_TOTAL / 2) / ROLLOUT_PHASE_6_TOTAL))
ROLLOUT_PHASE_7_PERCENT=$(((ROLLOUT_PHASE_7_COMPLETED * 100 + ROLLOUT_PHASE_7_TOTAL / 2) / ROLLOUT_PHASE_7_TOTAL))

format_integer() {
  local remaining="$1"
  local formatted=""
  while [ "${#remaining}" -gt 3 ]; do
    formatted=",${remaining: -3}${formatted}"
    remaining="${remaining:0:${#remaining}-3}"
  done
  printf '%s%s' "$remaining" "$formatted"
}

FORMATTED_TESTS=$(format_integer "$TESTS")
FORMATTED_GO_TESTS=$(format_integer "$GO_TESTS")
FORMATTED_FRONTEND_TESTS=$(format_integer "$FRONTEND_TESTS")
WEB_COMPONENTS=$(jq -r '.web_client.components' "$STATUS_FILE")
WEB_ROUTES=$(jq -r '.web_client.routes' "$STATUS_FILE")
WEB_STORES=$(jq -r '.web_client.stores' "$STATUS_FILE")
WEB_SERVICES=$(jq -r '.web_client.services' "$STATUS_FILE")
WEB_BUILD_SIZE_GZIP=$(jq -r '.web_client.build_size_gzip' "$STATUS_FILE")
LICENSE_ID=$(jq -r '.license.spdx_id' "$STATUS_FILE")
LICENSE_DECISION_RECORD=$(jq -r '.license.decision_record' "$STATUS_FILE")
LICENSE_ATTRIBUTION=$(jq -r '.license.attribution' "$STATUS_FILE")

echo "Source of Truth (status.json):"
echo "  Version: $VERSION"
echo "  Tests: $TESTS"
echo "  Supply: $SUPPLY"
echo ""

ERRORS=0

if [ $((GO_TESTS + RUST_TESTS + FRONTEND_TESTS)) -ne "$TESTS" ]; then
  echo "FAIL Test breakdown does not sum to total"
  ERRORS=$((ERRORS+1))
fi
if [ "$MODULE_TESTS" -ne "$GO_TESTS" ]; then
  echo "FAIL Module test counts do not sum to Go total"
  ERRORS=$((ERRORS+1))
fi
if [ $((SUPPLY * 10 ** DECIMALS)) -ne "$BASE_CAP" ]; then
  echo "FAIL Display supply/decimals do not match base-unit cap"
  ERRORS=$((ERRORS+1))
fi
if [ "$ROLLOUT_COMPLETED" -gt "$ROLLOUT_TOTAL" ] ||
   [ "$ROLLOUT_PHASE_WORK_COMPLETED" -gt "$ROLLOUT_PHASE_WORK_TOTAL" ] ||
   [ "$ROLLOUT_PHASE_6_COMPLETED" -gt "$ROLLOUT_PHASE_6_TOTAL" ] ||
   [ "$ROLLOUT_PHASE_7_COMPLETED" -gt "$ROLLOUT_PHASE_7_TOTAL" ]; then
  echo "FAIL Rollout completed counts exceed their totals"
  ERRORS=$((ERRORS+1))
fi

PHASE_7_ROADMAP_COMPLETED=$(awk '
  /^## Phase 7 / { in_phase = 1; next }
  /^## Rollout sequence/ { in_phase = 0 }
  in_phase && /^- \[x\]/ { completed++ }
  END { print completed + 0 }
' docs/ROLLOUT_ROADMAP.md)
PHASE_7_ROADMAP_TOTAL=$(awk '
  /^## Phase 7 / { in_phase = 1; next }
  /^## Rollout sequence/ { in_phase = 0 }
  in_phase && /^- \[[ x]\]/ { total++ }
  END { print total + 0 }
' docs/ROLLOUT_ROADMAP.md)
echo "Checking Phase 7 tracker accounting..."
if [ "$PHASE_7_ROADMAP_COMPLETED" -eq "$ROLLOUT_PHASE_7_COMPLETED" ] &&
   [ "$PHASE_7_ROADMAP_TOTAL" -eq "$ROLLOUT_PHASE_7_TOTAL" ]; then
  echo "  OK roadmap top-level Phase 7 count"
else
  echo "  FAIL roadmap Phase 7 is ${PHASE_7_ROADMAP_COMPLETED}/${PHASE_7_ROADMAP_TOTAL}, status records ${ROLLOUT_PHASE_7_COMPLETED}/${ROLLOUT_PHASE_7_TOTAL}"
  ERRORS=$((ERRORS+1))
fi
PHASE_7_FINAL_REVIEW_BLOCK_VALID=$(awk '
  /^- \[ \] Complete final release review and accountable launch authorization\.$/ {
    parents++
    in_parent = 1
    next
  }
  in_parent && /^  - \[ \] Freeze the release candidate while final evidence is reviewed\.$/ {
    freeze++
    next
  }
  in_parent && /^  - \[ \] Record an explicit go\/no-go decision and accountable approvers\.$/ {
    go_no_go++
    next
  }
  in_parent && /^- \[[ x]\]/ { in_parent = 0 }
  END { print (parents == 1 && freeze == 1 && go_no_go == 1) ? 1 : 0 }
' docs/ROLLOUT_ROADMAP.md)
if [ "$PHASE_7_FINAL_REVIEW_BLOCK_VALID" -eq 1 ]; then
  echo "  OK release freeze and go/no-go remain nested mandatory subchecks"
else
  echo "  FAIL Phase 7 final tracker item must retain both nested subchecks"
  ERRORS=$((ERRORS+1))
fi
echo ""

echo "Checking technology source of truth..."
if ! GO_MODULE_METADATA=$(go mod edit -json); then
  echo "FAIL unable to parse go.mod metadata"
  exit 1
fi
for module_version in \
  "github.com/cosmos/cosmos-sdk:$COSMOS_SDK_VERSION" \
  "github.com/cometbft/cometbft:$COMETBFT_VERSION" \
  "github.com/CosmWasm/wasmd:$WASMD_VERSION" \
  "github.com/CosmWasm/wasmvm/v2:$WASMVM_VERSION" \
  "github.com/cosmos/ibc-go/v8:$IBC_GO_VERSION"; do
  module=${module_version%%:*}
  expected=${module_version#*:}
  if jq -e --arg module "$module" \
    '.Replace[]? | select(.Old.Path == $module)' <<<"$GO_MODULE_METADATA" >/dev/null; then
    echo "  FAIL go.mod replaces monitored module $module"
    ERRORS=$((ERRORS+1))
    continue
  fi
  actual=$(jq -r --arg module "$module" \
    '.Require[] | select(.Path == $module) | .Version' <<<"$GO_MODULE_METADATA")
  if [ "$actual" = "$expected" ]; then
    echo "  OK $module $expected"
  else
    echo "  FAIL go.mod requires $module $actual, status records $expected"
    ERRORS=$((ERRORS+1))
  fi
done
LOCKED_VITE_VERSION=$(jq -r '.packages["node_modules/vite"].version' client-web/package-lock.json)
if [ "$VITE_VERSION" = "$LOCKED_VITE_VERSION" ]; then
  echo "  OK vite $VITE_VERSION"
else
  echo "  FAIL Vite status $VITE_VERSION does not match lockfile $LOCKED_VITE_VERSION"
  ERRORS=$((ERRORS+1))
fi
if jq -e '
  .limitations.tx_history |
  contains("submitted-only") and
  contains("GH-131") and
  (contains("not yet implemented") | not)
' "$STATUS_FILE" >/dev/null; then
  echo "  OK submitted-only transaction-history boundary"
else
  echo "  FAIL transaction-history status does not match GH-131"
  ERRORS=$((ERRORS+1))
fi
echo ""

echo "Checking community licensing source of truth..."
if jq -e '
  .license.status == "decided" and
  .license.spdx_id == "Apache-2.0" and
  .license.decision_issue == 219 and
  .license.decision_record == "https://github.com/NeaBouli/TrueRepublic/issues/219#issuecomment-5423337355" and
  .license.attribution == "TrueRepublic contributors" and
  .license.copyright_model == "individual_contributors_retain_rights" and
  .license.scope == "maintained_source_and_documentation" and
  .license.excluded == [
    "brand_assets",
    "artwork",
    "historical_pdfs",
    "archived_historical_evidence",
    "third_party_materials"
  ]
' "$STATUS_FILE" >/dev/null &&
  [ "$(jq -r '.selected_spdx_id' configs/legal/license-decision.json)" = "$LICENSE_ID" ] &&
  [ "$(jq -r '.decision_record' configs/legal/license-decision.json)" = "$LICENSE_DECISION_RECORD" ] &&
  [ "$(jq -r '.license' client-web/package.json)" = "$LICENSE_ID" ] &&
  [ "$(jq -r '.packages[""].license' client-web/package-lock.json)" = "$LICENSE_ID" ]; then
  echo "  OK machine-readable license state"
else
  echo "  FAIL machine-readable license state is inconsistent"
  ERRORS=$((ERRORS+1))
fi
for license_file in README.md CONTRIBUTING.md client-web/README.md docs/index.html \
  docs/ROLLOUT_ROADMAP.md wiki/Home.md wiki/status/Current-Status.md; do
  if grep -Fq "$LICENSE_ID" "$license_file" &&
    grep -Fq "$LICENSE_ATTRIBUTION" "$license_file"; then
    echo "  OK $license_file license and attribution"
  else
    echo "  FAIL $license_file lacks current license or attribution"
    ERRORS=$((ERRORS+1))
  fi
done
echo ""

check_file() {
  local file="$1"
  local label="$2"

  if [ ! -f "$file" ]; then
    echo "FAIL: required file $file not found"
    ERRORS=$((ERRORS+1))
    return
  fi

  echo "Checking $label ($file)..."
  grep -Fq "$VERSION" "$file" && echo "  OK Version" || { echo "  FAIL Version ($VERSION not found)"; ERRORS=$((ERRORS+1)); }
  if grep -Eq "(^|[^[:digit:],])(${TESTS}|${FORMATTED_TESTS})([^[:digit:],]|$)" "$file"; then
    echo "  OK Tests"
  else
    echo "  FAIL Tests ($TESTS or $FORMATTED_TESTS not found)"
    ERRORS=$((ERRORS+1))
  fi
  echo ""
}

check_file "README.md" "README"
check_file "CLAUDE.md" "Agent Guide"
check_file "docs/index.html" "Landing Page"
check_file "wiki/Home.md" "Wiki Home"
check_file "wiki/status/Current-Status.md" "Wiki Current Status"
check_file "wiki/status/Testing-Status.md" "Wiki Testing Status"

echo "Checking public technology documentation..."
for file_value in \
  "README.md:$COSMOS_SDK_VERSION" \
  "README.md:$COMETBFT_VERSION" \
  "README.md:$WASMD_VERSION" \
  "README.md:$IBC_GO_VERSION" \
  "CLAUDE.md:$COSMOS_SDK_VERSION" \
  "CLAUDE.md:$COMETBFT_VERSION" \
  "CLAUDE.md:$WASMD_VERSION" \
  "CLAUDE.md:$WASMVM_VERSION" \
  "CLAUDE.md:$IBC_GO_VERSION" \
  "CLAUDE.md:Vite $VITE_VERSION" \
  "wiki/Home.md:$COSMOS_SDK_VERSION" \
  "wiki/Home.md:$COMETBFT_VERSION" \
  "wiki/Home.md:$WASMD_VERSION" \
  "wiki/Home.md:$WASMVM_VERSION" \
  "wiki/Home.md:$IBC_GO_VERSION" \
  "wiki/Home.md:Vite $VITE_SERIES" \
  "wiki/develop/Code-Structure.md:$COSMOS_SDK_VERSION" \
  "wiki/develop/Code-Structure.md:$COMETBFT_VERSION" \
  "wiki/develop/Architecture-Overview.md:$COSMOS_SDK_VERSION" \
  "wiki/develop/Architecture-Overview.md:$COMETBFT_VERSION" \
  "docs/FAQ.md:$COSMOS_SDK_VERSION" \
  "docs/FAQ.md:$COMETBFT_VERSION" \
  "docs/FAQ.md:$WASMD_VERSION" \
  "docs/legal/GH-219-LICENSE-DECISION.md:$WASMVM_VERSION" \
  "docs/FAQ.md:Vite $VITE_SERIES"; do
  file=${file_value%%:*}
  value=${file_value#*:}
  if grep -Fq "$value" "$file"; then
    echo "  OK $file contains $value"
  else
    echo "  FAIL $file does not contain $value"
    ERRORS=$((ERRORS+1))
  fi
done
echo ""

check_count_file() {
  local file="$1"
  local label="$2"
  local field="$3"
  local count="$4"
  local formatted="$5"

  if [ ! -f "$file" ]; then
    echo "FAIL: required file $file not found"
    ERRORS=$((ERRORS+1))
    return
  fi

  echo "Checking $label ($file)..."
  if grep -F "$field" "$file" |
    grep -Eq "(^|[^[:digit:],])(${count}|${formatted})([^[:digit:],]|$)"; then
    echo "  OK Count"
  else
    echo "  FAIL Count ($count or $formatted not found beside '$field')"
    ERRORS=$((ERRORS+1))
  fi
  echo ""
}

check_count_file "docs/FAQ.md" "FAQ total" "The recovery baseline has" "$TESTS" "$FORMATTED_TESTS"
check_count_file "docs/FAQ.md" "FAQ Go breakdown" "The recovery baseline has" "$GO_TESTS" "$FORMATTED_GO_TESTS"
check_count_file "docs/QUICKSTART.md" "Quickstart Go total" "All Go tests" "$GO_TESTS" "$FORMATTED_GO_TESTS"
check_count_file "docs/ROLLOUT_ROADMAP.md" "Rollout roadmap total" "The source of truth records" "$TESTS" "$FORMATTED_TESTS"
check_count_file "docs/ROLLOUT_ROADMAP.md" "Rollout roadmap Go breakdown" "The source of truth records" "$GO_TESTS" "$FORMATTED_GO_TESTS"
check_count_file "wiki/develop/Architecture-Overview.md" "Architecture Go total" "| **Test** |" "$GO_TESTS" "$FORMATTED_GO_TESTS"
check_count_file "README.md" "README client breakdown" "tests recovery-verified locally" "$FRONTEND_TESTS" "$FORMATTED_FRONTEND_TESTS"
check_count_file "wiki/status/Testing-Status.md" "Wiki client breakdown" "| Maintained client |" "$FRONTEND_TESTS" "$FORMATTED_FRONTEND_TESTS"

echo "Checking maintained-client inventory..."
ACTUAL_COMPONENTS=$(find client-web/src/components -type f -name '*.tsx' ! -name '*.test.tsx' | wc -l | tr -d ' ')
ACTUAL_ROUTES=$(grep -c '^  { path:' client-web/src/routes.tsx || true)
ACTUAL_STORES=$(find client-web/src/stores -type f -name '*.ts' ! -name '*.test.ts' | wc -l | tr -d ' ')
ACTUAL_SERVICES=$(find client-web/src/services -type f -name '*.ts' ! -name '*.test.ts' | wc -l | tr -d ' ')
for item in \
  "components:$WEB_COMPONENTS:$ACTUAL_COMPONENTS" \
  "routes:$WEB_ROUTES:$ACTUAL_ROUTES" \
  "stores:$WEB_STORES:$ACTUAL_STORES" \
  "services:$WEB_SERVICES:$ACTUAL_SERVICES"; do
  IFS=: read -r label expected actual <<<"$item"
  if [ "$expected" -eq "$actual" ]; then
    echo "  OK $label"
  else
    echo "  FAIL $label (status $expected, source $actual)"
    ERRORS=$((ERRORS+1))
  fi
done
grep -Fq "$WEB_BUILD_SIZE_GZIP" docs/agent-bridge/SECURITY_NOTES.md &&
  echo "  OK build gzip" ||
  { echo "  FAIL build gzip ($WEB_BUILD_SIZE_GZIP not documented)"; ERRORS=$((ERRORS+1)); }
echo ""

echo "Checking wiki module table (wiki/status/Testing-Status.md)..."
while IFS='|' read -r module label; do
  expected=$(jq -r --arg module "$module" '.modules[$module]' "$STATUS_FILE")
  if grep -Fq "| Go $label | $expected |" wiki/status/Testing-Status.md; then
    echo "  OK $module"
  else
    echo "  FAIL $module ($expected not found)"
    ERRORS=$((ERRORS+1))
  fi
done <<'MODULES'
root|root/application
candidateevidence|candidate evidence
crossrunevidence|cross-run evidence
genesisevidence|rollout genesis evidence
capacitypolicy|capacity policy
deploymentevidence|deployment evidence
healthcheck|health checks
incidentpolicy|incident policy
migration|migration
networkpolicy|network policy
observability|observability
token|token
topologypolicy|topology policy
treasury|treasury
dex|DEX
truedemocracy|governance
zkpprover|test-only ZKP prover
releaseevidence|release evidence
installlifecycle|install lifecycle
sovereignv4protocol|Sovereign V4 protocol
ocievidence|OCI evidence
MODULES
echo ""

echo "Checking rollout status (docs/index.html)..."
grep -Fq "${ROLLOUT_COMPLETED} of ${ROLLOUT_TOTAL}" docs/index.html &&
  echo "  OK Full checklist" ||
  { echo "  FAIL Full checklist"; ERRORS=$((ERRORS+1)); }
grep -Fq "${ROLLOUT_PHASE_WORK_COMPLETED} of ${ROLLOUT_PHASE_WORK_TOTAL}" docs/index.html &&
  echo "  OK Phase work" ||
  { echo "  FAIL Phase work"; ERRORS=$((ERRORS+1)); }
grep -Fq "${ROLLOUT_PHASE_6_COMPLETED} of ${ROLLOUT_PHASE_6_TOTAL}" docs/index.html &&
  echo "  OK Phase 6" ||
  { echo "  FAIL Phase 6"; ERRORS=$((ERRORS+1)); }
grep -Fq "${ROLLOUT_PHASE_7_COMPLETED} of ${ROLLOUT_PHASE_7_TOTAL}" docs/index.html &&
  echo "  OK Phase 7" ||
  { echo "  FAIL Phase 7"; ERRORS=$((ERRORS+1)); }
grep -Fq "${ROLLOUT_PERCENT}%" docs/index.html &&
  echo "  OK Rounded rollout percentage" ||
  { echo "  FAIL Rounded rollout percentage"; ERRORS=$((ERRORS+1)); }
grep -Fq "${ROLLOUT_PHASE_WORK_PERCENT}%" docs/index.html &&
  echo "  OK Rounded phase-work percentage" ||
  { echo "  FAIL Rounded phase-work percentage"; ERRORS=$((ERRORS+1)); }
grep -Fq "${ROLLOUT_PHASE_6_PERCENT}%" docs/index.html &&
  echo "  OK Rounded Phase 6 percentage" ||
  { echo "  FAIL Rounded Phase 6 percentage"; ERRORS=$((ERRORS+1)); }
grep -Fq "${ROLLOUT_PHASE_7_PERCENT}%" docs/index.html &&
  echo "  OK Rounded Phase 7 percentage" ||
  { echo "  FAIL Rounded Phase 7 percentage"; ERRORS=$((ERRORS+1)); }
grep -Fq "Phase 7 is ${ROLLOUT_PHASE_7_COMPLETED}/${ROLLOUT_PHASE_7_TOTAL}" README.md &&
  echo "  OK README Phase 7" ||
  { echo "  FAIL README Phase 7"; ERRORS=$((ERRORS+1)); }
grep -Fq "Phase 7: **${ROLLOUT_PHASE_7_COMPLETED}/${ROLLOUT_PHASE_7_TOTAL}**" wiki/status/Roadmap.md &&
  echo "  OK Wiki Phase 7" ||
  { echo "  FAIL Wiki Phase 7"; ERRORS=$((ERRORS+1)); }
grep -Fq "Phase 7 is ${ROLLOUT_PHASE_7_COMPLETED}/${ROLLOUT_PHASE_7_TOTAL}" wiki/Home.md &&
  echo "  OK Wiki Home Phase 7" ||
  { echo "  FAIL Wiki Home Phase 7"; ERRORS=$((ERRORS+1)); }
grep -Fq "Phase 7 is ${ROLLOUT_PHASE_7_COMPLETED}/${ROLLOUT_PHASE_7_TOTAL}" wiki/status/Current-Status.md &&
  echo "  OK Wiki Current Status Phase 7" ||
  { echo "  FAIL Wiki Current Status Phase 7"; ERRORS=$((ERRORS+1)); }
echo ""

if [ "$ERRORS" -gt 0 ]; then
  echo "FAILED: $ERRORS inconsistencies found"
  exit 1
else
  echo "PASSED: All docs consistent"
fi
