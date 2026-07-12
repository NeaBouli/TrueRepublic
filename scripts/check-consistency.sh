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

VERSION=$(jq -r '.version' "$STATUS_FILE")
TESTS=$(jq -r '.tests.total' "$STATUS_FILE")
SUPPLY=$(jq -r '.token.max_supply' "$STATUS_FILE")
GO_TESTS=$(jq -r '.tests.go' "$STATUS_FILE")
RUST_TESTS=$(jq -r '.tests.rust' "$STATUS_FILE")
FRONTEND_TESTS=$(jq -r '.tests.frontend' "$STATUS_FILE")
MODULE_TESTS=$(jq '[.modules[]] | add' "$STATUS_FILE")
BASE_CAP=$(jq -r '.token.max_supply_base_units' "$STATUS_FILE")
DECIMALS=$(jq -r '.token.decimals' "$STATUS_FILE")

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
  grep -Fq "$TESTS" "$file" && echo "  OK Tests" || { echo "  FAIL Tests ($TESTS not found)"; ERRORS=$((ERRORS+1)); }
  echo ""
}

check_file "README.md" "README"
check_file "CLAUDE.md" "Agent Guide"
check_file "docs/index.html" "Landing Page"
check_file "wiki/Home.md" "Wiki Home"
check_file "wiki/status/Current-Status.md" "Wiki Current Status"
check_file "wiki/status/Testing-Status.md" "Wiki Testing Status"

if [ "$ERRORS" -gt 0 ]; then
  echo "FAILED: $ERRORS inconsistencies found"
  exit 1
else
  echo "PASSED: All docs consistent"
fi
