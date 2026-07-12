# Test Coverage

Current executable evidence is maintained in
[`docs/status.json`](https://github.com/NeaBouli/TrueRepublic/blob/fix/GH-8-docs-final/docs/status.json)
and summarized in [Testing Status](../status/Testing-Status).

## Recovery suites

- Go: 650 passing cases across root/application, token, treasury, DEX, and
  governance packages.
- Rust/CosmWasm: 26 passing cases.
- Maintained `client-web`: 8 passing cases plus lint, build, and audit.
- Legacy clients are not counted in the authoritative total; focused legacy
  checks do not make them safe for real keys or funds.

## Coverage snapshot

| Package | Statement coverage |
|---|---:|
| root/application | 64.3% |
| token | 92.6% |
| treasury | 97.0% |
| DEX | 45.3% |
| governance | 58.9% |

The critical recovery paths also include explicit cap boundaries, bank/custom
ledger reconciliation, atomic rollback, invariant corruption, ZKP replay and
encoding failures, generated validator-key binding, persistent restart, and
Docker restart tests.

Coverage percentages and test counts are evidence for the current stacked
branch only. They are not a production-readiness threshold or external audit.
