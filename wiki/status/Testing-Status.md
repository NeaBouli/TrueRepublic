# Testing Status

The current **v0.4.0 recovery** source of truth records **684 verified cases**.

| Suite | Passing cases |
|---|---:|
| Go root/application | 40 |
| Go token | 12 |
| Go treasury | 36 |
| Go DEX | 116 |
| Go governance | 446 |
| Rust/CosmWasm | 26 |
| Maintained client | 8 |
| **Total** | **684** |

## Current Go coverage

| Package | Statements |
|---|---:|
| root/application | 64.3% |
| token | 92.6% |
| treasury | 97.0% |
| DEX | 45.3% |
| governance | 58.9% |

## Reproduction commands

```bash
CGO_ENABLED=1 go test ./... -race -cover -count=1 -timeout=600s
go vet ./...
CGO_ENABLED=1 go build ./...
./scripts/check-consistency.sh
```

The maintained client is verified with `npm ci`, lint, 8 tests, production
build, and audit. The CosmWasm workspace is verified with tests, formatting,
Clippy, build, and audit.

GH-21 final-head run `29170968611` passes Go race/coverage and the Docker
first-block/same-container-restart/height-advance gate. Security Scan
`29170832988` passes all five jobs on the audited code head.

Green tests are recovery evidence, not an external security or production
approval. See [Current Status](Current-Status) for remaining gates.
