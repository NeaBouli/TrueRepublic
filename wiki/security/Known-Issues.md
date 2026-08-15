# Known Issues and Release Blockers

## Critical release blockers

### Recovery foundation is not a production approval

The 21M cap, custody, issuance, DEX, genesis/invariant, ZKP, and node-lifecycle
remediations were verified and merged to `main` through the ordered recovery
PRs. This does not replace independent cryptographic, multi-node operations,
or release-security review.

### Anonymous voting is not client-ready

The maintained web client rejects mock proof generation/submission. GH-209
implements recipient-bound atomic rewards without changing the frozen circuit
or setup artifacts, but direct payout publicly links the vote/nullifier event
to the chosen address. A production-qualified prover, trusted-setup/circuit
review, privacy analysis, and audited submission path are still required.

## High-priority operational gaps

- Single-node native and Docker restart pass. Bounded four-validator failure,
  restart, catch-up, partition recovery, trusted state sync, and sanitized
  backup/restore/export/import, compatible binary rollback, and single-signer
  identity failover now pass. IBC relaying/upgrades, persisted-state
  consensus-breaking migration recovery, authenticated consensus-key rotation,
  compromised consensus-key eviction/recovery, and network-policy drills remain
  open.
- GH-187 makes unsupported IBC/CosmWasm staking and distribution adapters
  fail closed and proves `x/staking`/`x/distribution` remain unmounted. External
  relayers/counterparties, IBC client upgrades, and arbitrary migrations remain
  unqualified.
- Production monitoring, alerting, incident response, validator key custody,
  and release procedures are not independently verified.

## Legacy client blockers

- `web-wallet` carried 70 dependency advisories plus broken/obsolete
  transaction and query paths. It was retired and removed under GH-112.
- The former `mobile-wallet` prototype had high/critical dependency advisories,
  no meaningful tests, unsafe mnemonic handling, and a broken Android bundle.
  It was retired and removed under GH-102 and must not be recovered for real keys.
- No legacy or retired client is approved for real keys or funds.

## Review boundaries

Green CI, CodeRabbit, and DeepScan checks are not an external consensus or
cryptographic audit. CodeRabbit was rate-limited on parts of the recovery
stack, so a green status must not be described as substantive independent
review where no findings were produced.

Track authoritative progress in
[Issue #4](https://github.com/NeaBouli/TrueRepublic/issues/4) and the repository
`BRIDGE.md`.
