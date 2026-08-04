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
  ($rollout.phase_6_total | positive_integer)
' "$STATUS_FILE" >/dev/null; then
  echo "FAIL Rollout counts must be nonnegative integers and totals must be positive integers"
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
ROLLOUT_PERCENT=$(((ROLLOUT_COMPLETED * 100 + ROLLOUT_TOTAL / 2) / ROLLOUT_TOTAL))
ROLLOUT_PHASE_WORK_PERCENT=$(((ROLLOUT_PHASE_WORK_COMPLETED * 100 + ROLLOUT_PHASE_WORK_TOTAL / 2) / ROLLOUT_PHASE_WORK_TOTAL))
ROLLOUT_PHASE_6_PERCENT=$(((ROLLOUT_PHASE_6_COMPLETED * 100 + ROLLOUT_PHASE_6_TOTAL / 2) / ROLLOUT_PHASE_6_TOTAL))

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
   [ "$ROLLOUT_PHASE_6_COMPLETED" -gt "$ROLLOUT_PHASE_6_TOTAL" ]; then
  echo "FAIL Rollout completed counts exceed their totals"
  ERRORS=$((ERRORS+1))
fi

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

check_count_file() {
  local file="$1"
  local label="$2"
  local count="$3"
  local formatted="$4"

  if [ ! -f "$file" ]; then
    echo "FAIL: required file $file not found"
    ERRORS=$((ERRORS+1))
    return
  fi

  echo "Checking $label ($file)..."
  if grep -Eq "(^|[^[:digit:],])(${count}|${formatted})([^[:digit:],]|$)" "$file"; then
    echo "  OK Count"
  else
    echo "  FAIL Count ($count or $formatted not found)"
    ERRORS=$((ERRORS+1))
  fi
  echo ""
}

check_count_file "docs/FAQ.md" "FAQ total" "$TESTS" "$FORMATTED_TESTS"
check_count_file "docs/FAQ.md" "FAQ Go breakdown" "$GO_TESTS" "$FORMATTED_GO_TESTS"
check_count_file "docs/QUICKSTART.md" "Quickstart Go total" "$GO_TESTS" "$FORMATTED_GO_TESTS"
check_count_file "docs/ROLLOUT_ROADMAP.md" "Rollout roadmap total" "$TESTS" "$FORMATTED_TESTS"
check_count_file "docs/ROLLOUT_ROADMAP.md" "Rollout roadmap Go breakdown" "$GO_TESTS" "$FORMATTED_GO_TESTS"
check_count_file "wiki/develop/Architecture-Overview.md" "Architecture Go total" "$GO_TESTS" "$FORMATTED_GO_TESTS"

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
grep -Fq "${ROLLOUT_PERCENT}%" docs/index.html &&
  echo "  OK Rounded rollout percentage" ||
  { echo "  FAIL Rounded rollout percentage"; ERRORS=$((ERRORS+1)); }
grep -Fq "${ROLLOUT_PHASE_WORK_PERCENT}%" docs/index.html &&
  echo "  OK Rounded phase-work percentage" ||
  { echo "  FAIL Rounded phase-work percentage"; ERRORS=$((ERRORS+1)); }
grep -Fq "${ROLLOUT_PHASE_6_PERCENT}%" docs/index.html &&
  echo "  OK Rounded Phase 6 percentage" ||
  { echo "  FAIL Rounded Phase 6 percentage"; ERRORS=$((ERRORS+1)); }
echo ""

if [ "$ERRORS" -gt 0 ]; then
  echo "FAILED: $ERRORS inconsistencies found"
  exit 1
else
  echo "PASSED: All docs consistent"
fi
