# Known Issues and Release Blockers

> **Non-production (recovery status).** TrueRepublic v0.4 is not approved for
> production, mainnet, real keys or real funds, and the project operates no
> public seeds, RPC, snapshot or status services. See
> [`SECURITY.md`](https://github.com/NeaBouli/TrueRepublic/security/policy) and
> the [rollout roadmap](https://github.com/NeaBouli/TrueRepublic/blob/main/docs/ROLLOUT_ROADMAP.md).

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
  identity failover now pass. Since then the canonical
  [rollout roadmap](../../docs/ROLLOUT_ROADMAP.md) records as completed:
  governed consensus-breaking migration with partial-migration rollback
  (GH-184), authenticated consensus-key rotation with permanent old-key
  revocation (GH-56), coupled key/signer custody and compromise containment
  (GH-55), and the role-based peer/listener/firewall policy (GH-71). These are
  local, bounded proofs; live multi-operator drills, external IBC relayers and
  production operations remain open.
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

The recovery foundation is preserved in completed
[Issue #4](https://github.com/NeaBouli/TrueRepublic/issues/4). Track remaining
production-readiness work in
[Issue #29](https://github.com/NeaBouli/TrueRepublic/issues/29) and the
repository `BRIDGE.md`.
