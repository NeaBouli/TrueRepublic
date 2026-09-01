# Test Coverage

Current executable evidence is maintained in
[`docs/status.json`](https://github.com/NeaBouli/TrueRepublic/blob/main/docs/status.json)
and summarized in [Testing Status](../status/Testing-Status).

## Recovery suites

- Go: 1,974 passing standard-suite cases across the maintained application and
  evidence packages.
- Rust/CosmWasm: 26 passing cases.
- Maintained `client-web`: 319 passing cases plus lint, production build,
  enforced bundle budgets, and guarded live audit.
- Standard-suite total: 2,319. Separately gated multi-process, IBC, governed
  upgrade, Go/WASM verifier, and keeper-replay proofs are additional evidence
  and are not counted in that total.
- Legacy clients are retired and are not part of the maintained evidence.

## Coverage snapshot

| Package | Statement coverage |
|---|---:|
| root/application | 73.6% |
| candidate evidence | 87.3% |
| cross-run evidence | 86.5% |
| token | 95.6% |
| treasury | 97.0% |
| DEX | 51.1% |
| governance | 64.3% |

The critical recovery paths also include explicit cap boundaries, bank/custom
ledger reconciliation, atomic rollback, invariant corruption, ZKP replay and
encoding failures, generated validator-key binding, persistent restart, Docker
restart, and separately gated four-validator failure/restart/catch-up,
join/replacement, partition recovery, trusted state-sync, and sanitized
backup/restore/export/import tests.

Coverage percentages and test counts are recovery evidence, not a
production-readiness threshold or external audit. The complete package table
and reproduction commands are maintained in [Testing Status](../status/Testing-Status).
