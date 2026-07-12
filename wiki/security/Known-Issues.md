# Known Issues and Release Blockers

## Critical release blockers

### Recovery stack is unmerged and independently unapproved

The 21M cap, custody, issuance, DEX, genesis/invariant, ZKP, and node-lifecycle
remediations live in ordered draft PRs. Do not use administrator bypass or
present branch evidence as deployed `main` behavior.

### Anonymous voting is not client-ready

Both web clients reject mock proof generation/submission. A compatible real
prover, trusted-setup/circuit review, privacy analysis, and safe anonymous
reward recipient binding are still required.

## High-priority operational gaps

- Single-node native and Docker restart pass, but multi-node consensus, peer
  failure, IBC relaying/upgrades, backup/restore, and rollback drills do not.
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
