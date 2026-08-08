# Testing Status

The current **v0.4.0 recovery** source of truth records **1,520 verified cases**.

| Suite | Passing cases |
|---|---:|
| Go root/application | 115 |
| Go capacity policy | 48 |
| Go deployment evidence | 71 |
| Go health checks | 55 |
| Go incident policy | 54 |
| Go migration | 82 |
| Go network policy | 126 |
| Go observability | 50 |
| Go token | 12 |
| Go topology policy | 56 |
| Go treasury | 36 |
| Go DEX | 124 |
| Go governance | 524 |
| Rust/CosmWasm | 26 |
| Maintained client | 141 |
| **Total** | **1,520** |

## Current Go coverage

| Package | Statements |
|---|---:|
| root/application | 70.8% |
| capacity policy | 85.5% |
| deployment evidence | 90.8% |
| health checks | 97.2% |
| incident policy | 90.7% |
| migration | 84.6% |
| network policy | 95.5% |
| observability | 80.3% |
| token | 92.6% |
| topology policy | 85.8% |
| treasury | 97.0% |
| DEX | 49.1% |
| governance | 62.6% |

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

The maintained client is verified with `npm ci`, lint, 141 cases (131 Vitest plus
ten Node policy/budget cases), production build, and guarded live audit. The CosmWasm workspace is verified with tests, formatting,
Clippy, build, and audit.

GH-32/GH-41/GH-43/GH-45/GH-53/GH-55/GH-56/GH-60/GH-61/GH-93/GH-97 add the separately gated
multi-validator failure/restart/catch-up, partition-recovery, trusted
state-sync, sanitized backup/restore/export/import, compatible binary
replacement/rollback, single-signer identity failover, authenticated
consensus-key rotation, inactive-validator genesis round-trip,
legacy-authority migration/rollback, secret-free incident rehearsal, bounded
sustained-load qualification, and common-height app-hash tests. They do
not replace the remaining
consensus-breaking migration, external paging drills, production sizing or
soak, live topology, private live rehearsal, or independent operations gates.

GH-101 adds the strict offline deployment-evidence parser, verifier, repository
contract, and CI manifest gate. It validates only a synthetic secret-free
envelope and does not replace live topology or private deployment evidence.

Green tests are recovery evidence, not an external security or production
approval. See [Current Status](Current-Status) for remaining gates.
