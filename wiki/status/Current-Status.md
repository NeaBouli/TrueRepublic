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
  identity; GH-206 generates a real proof through isolated Go/WASM from exact
  synthetic toxic-waste fixtures and proves native verifier compatibility,
  while GH-209 binds an atomic anonymous reward to one canonical recipient
  without changing the circuit/setup artifacts or recipient-independent
  nullifier. Direct payout is publicly linkable; mock and transaction
  submission remain fail-closed.
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
  GH-212 adds a strict offline two-target release-evidence contract with exact
  checksums, metadata, normalized SBOM and unsigned-provenance bindings, pinned
  release tools/platforms and container bases, without publishing or signing
  an artifact or claiming production rollout.
- 1,959 recovery-verified tests: 1,614 Go, 26 Rust, and 319 maintained-client,
  plus the separately gated GH-175/GH-178/GH-181 two-chain IBC packet/channel/compatible-restart recovery and GH-172
  contention/replay/restart process proofs.

## Recovery sequence

PR #9 → #15 → #16 → #17 → #18 → #19 → #22 → #23 → #24 → #27.

The recovery foundation and safe deployment-initialization wrapper were
reviewed, verified, and merged to `main` in this order.

## Release blockers

- Release qualification and independent security review.
- Signed and published release artifacts, authenticated provenance, exact
  tagged-candidate qualification, and staged deployment evidence; GH-212 is an
  unsigned repository-only foundation, not completion of these gates.
- Production-qualified Groth16 client prover, ceremony, submission path, and external circuit review.
- Independent privacy review of the implemented recipient binding and its
  documented direct-payout linkability.
- IBC/consensus-breaking migration recovery, external paging drills,
  private-live-capacity/live-topology deployment, private live rehearsal, and
  independent live operations evidence.
- The deprecated legacy web client was retired and removed under GH-112; the
  former mobile prototype was retired and removed under GH-102.

See the completed recovery foundation in
[Issue #4](https://github.com/NeaBouli/TrueRepublic/issues/4), the active
[rollout tracker #29](https://github.com/NeaBouli/TrueRepublic/issues/29),
[`BRIDGE.md`](https://github.com/NeaBouli/TrueRepublic/blob/main/BRIDGE.md),
and [`docs/status.json`](https://github.com/NeaBouli/TrueRepublic/blob/main/docs/status.json).
