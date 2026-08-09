#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$ROOT_DIR"

GOVULNCHECK_BIN=${GOVULNCHECK_BIN:-govulncheck}
GATES_FILE=${GATES_FILE:-configs/security/gates.json}
SECURITY_GATE_TODAY=${SECURITY_GATE_TODAY:-$(date -u +%F)}

for command in "$GOVULNCHECK_BIN" jq; do
  if ! command -v "$command" >/dev/null 2>&1; then
    echo "$command is required for the Go vulnerability gate" >&2
    exit 1
  fi
done

mapfile -t packages < <(./scripts/go-packages.sh --list)
if [[ ${#packages[@]} -eq 0 ]]; then
  echo "Go vulnerability gate found no maintained packages" >&2
  exit 1
fi

report=$(mktemp "${TMPDIR:-/tmp}/truerepublic-govuln.XXXXXX")
trap 'rm -f "$report"' EXIT

# JSON mode separates scanner execution errors from policy evaluation. The
# wrapper below then fails on every finding except the exact, active, no-fix
# entries declared in the bounded central policy.
CGO_ENABLED=1 "$GOVULNCHECK_BIN" -json "${packages[@]}" >"$report"

if ! jq -e -n \
  --slurpfile stream "$report" \
  --slurpfile config "$GATES_FILE" \
  --arg today "$SECURITY_GATE_TODAY" '
    [$stream[] |
      select(.finding?.trace[0]?.function? != null and .finding.trace[0].function != "") |
      .finding.osv] | unique as $found |
    [$stream[] |
      select(.finding?.trace[0]?.function? != null and .finding.trace[0].function != "") |
      select(.finding.fixed_version? != null and .finding.fixed_version != "") |
      .finding.osv] | unique as $fixable |
    ($config[0].go_vulnerability_exceptions // []) as $exceptions |
    [$exceptions[].id] as $raw_allowed |
    ($raw_allowed | unique) as $allowed |
    ($config[0].exception_max_days // 0) as $max_days |
    ($max_days | type == "number" and . >= 1 and . <= 30 and floor == .) and
    (($raw_allowed | length) == ($allowed | length)) and
    ($exceptions | all(
      ((.id | type) == "string") and (.id | test("^GO-[0-9]{4}-[0-9]{4}$")) and
      ((.approved_on | type) == "string") and
      ((.expires | type) == "string") and
      (.approved_on | test("^[0-9]{4}-[0-9]{2}-[0-9]{2}$")) and
      (.expires | test("^[0-9]{4}-[0-9]{2}-[0-9]{2}$")) and
      ((.reason | type) == "string") and ((.reason | length) >= 20) and
      ((try (.approved_on + "T00:00:00Z" | fromdateiso8601) catch null) as $approved |
       (try (.expires + "T00:00:00Z" | fromdateiso8601) catch null) as $expires |
       (try ($today + "T00:00:00Z" | fromdateiso8601) catch null) as $now |
       $approved != null and $expires != null and $now != null and
       $approved <= $now and $expires >= $now and
       ($expires - $approved) >= 0 and
       ($expires - $approved) <= ($max_days * 86400))
    )) and
    ($found == $allowed) and
    (($found - $fixable) == $found)
  ' >/dev/null; then
  echo "Go vulnerability policy rejected the current scan." >&2
  echo "Found reachable IDs:" >&2
  jq -r 'select(.finding?.trace[0]?.function? != null and .finding.trace[0].function != "") | .finding.osv' "$report" | sort -u >&2
  echo "Active allowed no-fix IDs:" >&2
  jq -r --arg today "$SECURITY_GATE_TODAY" \
    '.go_vulnerability_exceptions[] | select(.approved_on <= $today and .expires >= $today) | .id' \
    "$GATES_FILE" | sort -u >&2
  exit 1
fi

echo "Go vulnerability gate passed with exact active no-fix policy."
