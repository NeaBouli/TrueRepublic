# Known Issues and Release Blockers

## Critical release blockers

### Recovery foundation is not a production approval

The 21M cap, custody, issuance, DEX, genesis/invariant, ZKP, and node-lifecycle
remediations were verified and merged to `main` through the ordered recovery
PRs. This does not replace independent cryptographic, multi-node operations,
or release-security review.

### Anonymous voting is not client-ready

Both web clients reject mock proof generation/submission. A compatible real
prover, trusted-setup/circuit review, privacy analysis, and safe anonymous
reward recipient binding are still required.

## High-priority operational gaps

- Single-node native and Docker restart pass. Bounded four-validator failure,
  restart, catch-up, partition recovery, trusted state sync, and sanitized
  backup/restore/export/import, compatible binary rollback, and single-signer
  identity failover now pass. IBC relaying/upgrades, persisted-state
  consensus-breaking migration recovery, authenticated consensus-key rotation,
  compromised consensus-key eviction/recovery, and network-policy drills remain
  open.
- IBC staking/upgrade and standard CosmWasm staking/distribution remain explicit
  stubs.
- Production monitoring, alerting, incident response, validator key custody,
  and release procedures are not independently verified.

## Legacy client blockers

- `web-wallet` uses an obsolete client/toolchain architecture and is preserved
  only for migration/reference.
- `mobile-wallet` has unresolved high/critical dependency advisories and no
  meaningful test suite.
- Neither legacy client is approved for real keys or funds.

## Review boundaries

Green CI, CodeRabbit, and DeepScan checks are not an external consensus or
cryptographic audit. CodeRabbit was rate-limited on parts of the recovery
stack, so a green status must not be described as substantive independent
review where no findings were produced.

Track authoritative progress in
[Issue #4](https://github.com/NeaBouli/TrueRepublic/issues/4) and the repository
`BRIDGE.md`.
