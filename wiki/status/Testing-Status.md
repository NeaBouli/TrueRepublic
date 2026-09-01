# Testing Status

The current **v0.4.0 recovery** source of truth records **2,319 verified
standard-suite cases**. This arithmetic excludes the separately run opt-in
GH-266 Go/WASM-to-keeper replay, GH-175/GH-178/GH-181 IBC two-chain, and
GH-184 governed-upgrade gates.

| Suite | Passing cases |
|---|---:|
| Go root/application | 232 |
| Go candidate evidence | 77 |
| Go cross-run evidence | 107 |
| Go rollout genesis evidence | 68 |
| Go capacity policy | 48 |
| Go deployment evidence | 71 |
| Go health checks | 55 |
| Go incident policy | 54 |
| Go migration | 82 |
| Go network policy | 126 |
| Go observability | 50 |
| Go token | 37 |
| Go topology policy | 56 |
| Go treasury | 36 |
| Go DEX | 138 |
| Go governance | 614 |
| Go test-only ZKP prover | 8 |
| Go release evidence | 20 |
| Go install lifecycle | 24 |
| Go Sovereign V4 protocol | 36 |
| Go OCI evidence | 35 |
| Rust/CosmWasm | 26 |
| Maintained client | 319 |
| **Total** | **2,319** |

The published total is the reproducible standard-suite baseline; the opt-in
GH-175/GH-178/GH-181 IBC recovery, GH-184 upgrade, and GH-206/GH-266 Go/WASM
native-verifier plus keeper-replay (`./scripts/test-zkp-wasm-client.sh`) gates are additional
evidence and are not counted in the 1,974 Go subtotal.

## Current Go coverage

| Package | Statements |
|---|---:|
| root/application | 73.6% |
| candidate evidence | 87.3% |
| cross-run evidence | 86.5% |
| rollout genesis evidence | 84.2% |
| capacity policy | 85.5% |
| deployment evidence | 90.8% |
| health checks | 97.2% |
| incident policy | 90.7% |
| migration | 84.6% |
| network policy | 95.5% |
| observability | 80.3% |
| OCI evidence | 81.3% |
| release evidence | 74.5% |
| install lifecycle | 77.1% |
| token | 95.6% |
| topology policy | 85.8% |
| treasury | 97.0% |
| DEX | 51.1% |
| governance | 64.3% |

## Reproduction commands

```bash
./scripts/go-packages.sh go test -count=1
CGO_ENABLED=1 ./scripts/go-packages.sh go test -race -cover -count=1 -timeout=600s
TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test . \
  -run '^TestMultiValidatorConsensusRecovery$' -count=1 -timeout=300s -v
./scripts/go-packages.sh go vet
CGO_ENABLED=1 ./scripts/go-packages.sh go build
./scripts/test-zkp-wasm-client.sh
./scripts/check-consistency.sh
```

The maintained client is verified with `npm ci`, lint, 319 passing cases (309
Vitest plus ten Node policy/budget cases), production build, and guarded live
audit. The CosmWasm workspace is verified with tests, formatting,
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

GH-212 adds the strict two-target offline release-evidence parser/verifier,
repeated normalized SBOM parity, unsigned provenance binding, exact tool and
platform pins, and synthetic adversarial CI gate. It does not publish binaries,
authenticate SBOM component truth, sign an artifact, or approve rollout.

GH-258 adds 35 standard-suite OCI evidence cases plus a repository contract and
native Linux amd64/arm64 protected matrix. Each daemon and maintained-client
image is exported twice without cache and must match on OCI index, manifest,
config, and ordered layer digests. Only JSON evidence is retained; floating
runner/BuildKit versions and live Debian package sources leave cross-time
hermetic rebuilding open.

GH-261 adds 77 candidate-evidence cases and 16 root repository-contract events.
The verifier binds two deterministic daemon records and four OCI target
identities to one exact commit and explicitly simulated tag, rejects promoted
claims and undeclared payload members, and is exercised by the protected
metadata-only aggregation job. No real tag, release, signature, publication,
deployment, or rollout credit is created.

GH-273 adds 107 cross-run-evidence cases and its protected repository contract.
Hosted baseline run 33465480131 and comparison run 33466167289 matched both
daemon and all four OCI target identities on exact commit
`3b0d1639bb40c7df6733dd13a86252e1c8c9efd3`; the bounded pair report is valid
with zero violations. This is pair-scoped evidence, not a tagged, signed,
published, deployed, production, or long-term-hermetic release claim.

GH-266 extends the separate GH-206 compatibility command with a strict
versioned handoff and two env-gated keeper tests. The fresh maintained-client
proof must pay the bound recipient exactly once through
`RateProposalWithZKPPayout`; replay, proof corruption, chain/rating/root drift,
recipient substitution, blocked/noncanonical recipients, and malformed
handoffs must leave nullifier, rating, treasury, escrow, and balances unchanged.
Because these tests skip in the ordinary package suite and run only in the
dedicated gate, they add evidence without changing the 1,974 Go subtotal.

Green tests are recovery evidence, not an external security or production
approval. See [Current Status](Current-Status) for remaining gates.
