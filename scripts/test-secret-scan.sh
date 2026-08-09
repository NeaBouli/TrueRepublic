#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
GITLEAKS_BIN=${GITLEAKS_BIN:-gitleaks}
if ! command -v "$GITLEAKS_BIN" >/dev/null 2>&1; then
  echo "gitleaks is required; install the version pinned in configs/security/gates.json" >&2
  exit 1
fi

fixture=$(mktemp -d "${TMPDIR:-/tmp}/truerepublic-secret-scan.XXXXXX")
trap 'rm -rf "$fixture"' EXIT

printf '%s\n' 'public documentation without credentials' >"$fixture/safe.txt"
"$GITLEAKS_BIN" dir --config "$ROOT_DIR/.gitleaks.toml" --redact --no-banner \
  --exit-code 1 "$fixture" >/dev/null

printf 'token = ghp_' >"$fixture/leak.txt"
candidate='a9F3c7E1b5D8026A4c9E7f1B3d5A8c2E6f0B'
printf '%s' "${candidate:0:36}" >>"$fixture/leak.txt"
printf '\n' >>"$fixture/leak.txt"

if "$GITLEAKS_BIN" dir --config "$ROOT_DIR/.gitleaks.toml" --redact --no-banner \
  --exit-code 1 "$fixture" >/dev/null 2>&1; then
  echo "gitleaks accepted the planted credential fixture" >&2
  exit 1
fi

mkdir -p "$fixture/bin"
cat >"$fixture/bin/git" <<'EOF'
#!/usr/bin/env bash
exit 2
EOF
cat >"$fixture/bin/gitleaks" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
chmod +x "$fixture/bin/git" "$fixture/bin/gitleaks"
if PATH="$fixture/bin:$PATH" GITLEAKS_BIN="$fixture/bin/gitleaks" \
  "$ROOT_DIR/scripts/check-secret-scan.sh" >/dev/null 2>&1; then
  echo "maintained-tree secret scan ignored a Git enumeration failure" >&2
  exit 1
fi

echo "secret scan positive and negative fixtures passed"
