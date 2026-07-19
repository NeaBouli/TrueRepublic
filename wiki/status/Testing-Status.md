# Testing Status

The current **v0.4.0 recovery** source of truth records **689 verified cases**.

| Suite | Passing cases |
|---|---:|
| Go root/application | 41 |
| Go token | 12 |
| Go treasury | 36 |
| Go DEX | 116 |
| Go governance | 446 |
| Rust/CosmWasm | 26 |
| Maintained client | 8 |
| **Total** | **689** |

## Current Go coverage

| Package | Statements |
|---|---:|
| root/application | 64.9% |
| token | 92.6% |
| treasury | 97.0% |
| DEX | 45.3% |
| governance | 58.9% |

## Reproduction commands

```bash
CGO_ENABLED=1 go test ./... -race -cover -count=1 -timeout=600s
TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test . \
  -run '^TestMultiValidatorConsensusRecovery$' -count=1 -timeout=300s -v
go vet ./...
CGO_ENABLED=1 go build ./...
./scripts/check-consistency.sh
```

The maintained client is verified with `npm ci`, lint, 8 tests, production
build, and audit. The CosmWasm workspace is verified with tests, formatting,
Clippy, build, and audit.

GH-32/GH-41/GH-43/GH-45 add the separately gated four-validator
failure/restart/catch-up, partition-recovery, trusted state-sync, sanitized
backup/restore/export/import, and common-height app-hash tests. They do not
replace the remaining upgrade, rollback, validator-key operations, network
policy, load, or independent operations gates.

Green tests are recovery evidence, not an external security or production
approval. See [Current Status](Current-Status) for remaining gates.
