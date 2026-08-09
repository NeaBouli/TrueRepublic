# Current Status

**Version:** v0.4.0 recovery
**Release state:** recovery foundation merged to `main`
**Production-ready:** no

## Verified foundation

- Canonical `upnyx` base denomination with six decimals.
- Exact maximum supply of 21,000,000 PNYX / 21,000,000,000,000 `upnyx`.
- Bank-backed governance escrow, capped issuance, DEX custody, genesis
  reconciliation, and every-block supply/custody invariants on the ordered
  recovery stack.
- Chain/proposal/rating-bound ZKP statement and pinned genesis verification-key
  identity; the maintained web client rejects mock proof generation and
  submission fail-closed.
- Persistent Cosmos/Comet lifecycle with generated-key, bank-backed PoD
  genesis, native/Docker restart evidence, and a bounded four-validator
  failure/restart/catch-up, partition-recovery, trusted state-sync, and
  sanitized backup/restore, compatible binary replacement/rollback, and
  single-signer validator-identity failover harness. Supported node operation
  emits secret-minimized structured JSON logs through the central SDK/CometBFT
  logger boundary and exposes a private two-source CometBFT plus
  SDK/application metrics baseline. GH-85 adds an immutable Grafana dashboard,
  eleven deterministically tested Prometheus rules, recovery/testnet
  objectives, role ownership, and CI runtime query proof. GH-89 adds a strict
  synthetic cross-node contract that verifies sentry diversity, validator
  isolation, relationship-backed deny-by-default flows, and bounded public
  query ingress without publishing or deploying a real operator inventory.
  GH-93 adds a strict secret-free incident-command contract and eight synthetic
  halt, validator, key, backup/restore, upgrade/rollback, and migration
  rehearsals without claiming a private live operator drill.
  GH-97 adds a bounded four-validator sustained-load contract with 96 committed
  transactions plus resource, retention, restart, and ledger evidence without
  claiming production sizing or multi-day soak behavior.
  GH-101 adds a strict secret-free digest-bound deployment-evidence envelope
  and offline verifier without claiming or performing a live deployment.
- 1,607 recovery-verified tests: 1,440 Go, 26 Rust, and 141 maintained-client,
  plus the separately gated GH-172 contention/replay/restart process proof.

## Recovery sequence

PR #9 → #15 → #16 → #17 → #18 → #19 → #22 → #23 → #24 → #27.

The recovery foundation and safe deployment-initialization wrapper were
reviewed, verified, and merged to `main` in this order.

## Release blockers

- Release qualification and independent security review.
- Compatible real Groth16 client prover and external circuit/ceremony review.
- Privacy-preserving anonymous reward recipient binding.
- IBC/consensus-breaking migration recovery, external paging drills,
  private-live-capacity/live-topology deployment, private live rehearsal, and
  independent live operations evidence.
- The deprecated legacy web client was retired and removed under GH-112; the
  former mobile prototype was retired and removed under GH-102.

See [Issue #4](https://github.com/NeaBouli/TrueRepublic/issues/4),
[`BRIDGE.md`](https://github.com/NeaBouli/TrueRepublic/blob/main/BRIDGE.md),
and [`docs/status.json`](https://github.com/NeaBouli/TrueRepublic/blob/main/docs/status.json).
