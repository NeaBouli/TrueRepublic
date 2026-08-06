#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

fail() {
  echo "ERROR: $*" >&2
  exit 1
}

for file in x/truedemocracy/querier.go x/dex/querier.go; do
  [[ ! -e "$file" ]] || fail "retired legacy querier still exists: $file"
done

if grep -Eq '^func \(app \*TrueRepublicApp\) Query\(' app.go; then
  fail "app.go must not override BaseApp.Query for retired custom paths"
fi

if git grep -n -E 'custom/(truedemocracy|dex)' -- \
  '*.go' '*.ts' '*.tsx' '*.js' '*.mjs' \
  ':!query_protocol_boundary_test.go' >/dev/null; then
  fail "maintained source still consumes a retired custom query path"
fi

grep -Fq '/truedemocracy.Query/Domains' docs/developers/api-reference/abci-queries.md ||
  fail "supported truedemocracy gRPC route is not documented"
grep -Fq '/dex.Query/Pools' docs/developers/api-reference/abci-queries.md ||
  fail "supported DEX gRPC route is not documented"
grep -Fq 'not registered for custom modules' docs/developers/api-reference/abci-queries.md ||
  fail "custom-module HTTP gateway boundary is not documented"

for file in \
  docs/API.md \
  docs/developers/api-reference/rest-rpc.md \
  docs/developers/integration-guide/cosmjs-examples.md \
  docs/developers/architecture/system-overview.md \
  wiki/develop/Architecture-Overview.md \
  wiki/develop/Code-Structure.md \
  wiki/develop/Module-Deep-Dive.md; do
  if grep -Eq 'custom/(truedemocracy|dex)' "$file"; then
    fail "$file still advertises a retired custom query path"
  fi
done

if grep -Eq '/truerepublic/(truedemocracy|dex)' docs/API_REFERENCE.md; then
  fail "docs/API_REFERENCE.md advertises unregistered custom-module HTTP aliases"
fi

echo "Custom query retirement contract: PASS"
