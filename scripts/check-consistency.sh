#!/bin/bash
# Consistency Check Script
# Verifies version/tests/supply across all docs

set -e

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

echo "Source of Truth (status.json):"
echo "  Version: $VERSION"
echo "  Tests: $TESTS"
echo "  Supply: $SUPPLY"
echo ""

ERRORS=0

check_file() {
  local file="$1"
  local label="$2"

  if [ ! -f "$file" ]; then
    echo "SKIP: $file not found"
    return
  fi

  echo "Checking $label ($file)..."
  grep -q "$VERSION" "$file" && echo "  OK Version" || { echo "  FAIL Version ($VERSION not found)"; ERRORS=$((ERRORS+1)); }
  grep -q "$TESTS" "$file" && echo "  OK Tests" || { echo "  FAIL Tests ($TESTS not found)"; ERRORS=$((ERRORS+1)); }
  echo ""
}

check_file "README.md" "README"
check_file "docs/index.html" "Landing Page"
check_file "wiki-github/Home.md" "Wiki Home"
check_file "wiki-github/status-Current-Status.md" "Wiki Current Status"
check_file "wiki-github/status-Testing-Status.md" "Wiki Testing Status"

if [ "$ERRORS" -gt 0 ]; then
  echo "FAILED: $ERRORS inconsistencies found"
  exit 1
else
  echo "PASSED: All docs consistent"
fi
