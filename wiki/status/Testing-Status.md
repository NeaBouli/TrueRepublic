# Testing Status

The current **v0.4.0 recovery** source of truth records **853 verified cases**.

| Suite | Passing cases |
|---|---:|
| Go root/application | 62 |
| Go migration | 82 |
| Go token | 12 |
| Go treasury | 36 |
| Go DEX | 116 |
| Go governance | 511 |
| Rust/CosmWasm | 26 |
| Maintained client | 8 |
| **Total** | **853** |

## Current Go coverage

| Package | Statements |
|---|---:|
| root/application | 69.4% |
| migration | 84.6% |
| token | 92.6% |
| treasury | 97.0% |
| DEX | 45.3% |
| governance | 62.2% |

## Reproduction commands

```bash
./scripts/go-packages.sh go test -count=1
CGO_ENABLED=1 ./scripts/go-packages.sh go test -race -cover -count=1 -timeout=600s
TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test . \
  -run '^TestMultiValidatorConsensusRecovery$' -count=1 -timeout=300s -v
./scripts/go-packages.sh go vet
CGO_ENABLED=1 ./scripts/go-packages.sh go build
./scripts/check-consistency.sh
```

The maintained client is verified with `npm ci`, lint, 8 tests, production
build, and audit. The CosmWasm workspace is verified with tests, formatting,
Clippy, build, and audit.

GH-32/GH-41/GH-43/GH-45/GH-53/GH-55/GH-56/GH-60 add the separately gated
multi-validator failure/restart/catch-up, partition-recovery, trusted
state-sync, sanitized backup/restore/export/import, compatible binary
replacement/rollback, single-signer identity failover, authenticated
consensus-key rotation, inactive-validator genesis round-trip, and common-height
app-hash tests. They do not replace the remaining
consensus-breaking migration, network policy, load, or independent operations
gates.

Green tests are recovery evidence, not an external security or production
approval. See [Current Status](Current-Status) for remaining gates.
