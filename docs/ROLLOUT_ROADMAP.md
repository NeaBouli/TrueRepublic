# TrueRepublic Road to Rollout

Updated: 2026-07-15

TrueRepublic has a recovered and CI-verified v0.4 engineering foundation. It
is **not production-ready, mainnet-ready, or approved for real funds or keys**.
This document is the public checklist from the current recovery baseline to a
controlled rollout. Progress is tracked in
[GitHub issue #29](https://github.com/NeaBouli/TrueRepublic/issues/29); the
parent recovery record remains
[issue #4](https://github.com/NeaBouli/TrueRepublic/issues/4).

## Current baseline

- The ordered recovery merge chain is on `main`.
- The maximum supply is fixed at 21,000,000 PNYX.
- The source of truth records 689 recovery-verified tests: 655 Go, 26 Rust,
  and 8 maintained-client tests.
- Ledger, escrow, issuance, DEX custody, genesis, runtime invariants, ZKP
  statement binding, node persistence, and the safe operator-init boundary
  have CI-backed recovery evidence.
- GitHub Pages, security scans, and the current CI matrix are green.

The baseline is suitable for continued engineering. It is not a rollout
approval.

## Phase 1 — Network and disaster-recovery evidence

- [x] Build a reproducible four-validator consensus and recovery harness
  ([GH-32](https://github.com/NeaBouli/TrueRepublic/issues/32)).
- [x] Verify validator join, replacement, restart/catch-up, and leave
  power-zero evidence
  ([GH-39](https://github.com/NeaBouli/TrueRepublic/issues/39)).
- [x] Exercise network partitions, delayed peers, validator failure, and
  recovery without ledger divergence.
- [x] Verify state sync and catch-up from a trusted snapshot.
- [x] Run backup, restore, export, and import drills from documented artifacts.
- [ ] Prove application and consensus upgrades on persisted state.
- [ ] Prove rollback after a failed upgrade within a defined recovery window.
- [ ] Define and test validator-key backup, rotation, and compromise response.
- [ ] Define seed, persistent-peer, RPC, API, firewall, and rate-limit policy.

**Exit gate:** the same documented procedure must reproduce consensus,
recovery, upgrade, and rollback results on clean infrastructure.

## Phase 2 — Production ZKP and privacy path

- [ ] Integrate a compatible real Groth16 prover into the maintained client.
- [ ] Freeze and version the circuit, public-input order, field encodings,
  nullifier rules, and chain/proposal/rating binding.
- [ ] Produce reproducible circuit, proving-key, verification-key, and checksum
  artifacts.
- [ ] Document ceremony provenance, participant assumptions, and artifact
  rotation or circuit-upgrade rules.
- [ ] Add browser-to-chain proof compatibility tests using real proofs.
- [ ] Design and implement a front-running-safe anonymous reward-recipient
  binding without leaking voter identity.
- [ ] Complete independent cryptographic, privacy, and trusted-setup review.
- [ ] Keep anonymous submission fail-closed until every item above passes.

**Exit gate:** a real maintained-client proof must verify on-chain under the
published circuit identity, with no unresolved critical or high audit finding.

## Phase 3 — IBC and protocol completeness

- [ ] Run two-chain relayer tests for PNYX transfers.
- [ ] Test acknowledgements, timeouts, channel closure, replay resistance, and
  relayer interruption.
- [ ] Prove safe recovery after partial IBC and network failures.
- [ ] Test IBC behavior across application upgrades.
- [ ] Complete the supported upgrade path and its governance controls.
- [ ] Implement, replace, or explicitly remove remaining staking,
  distribution, and upgrade stubs.
- [ ] Document the exact supported Cosmos/IBC feature boundary.

**Exit gate:** supported IBC and upgrade flows pass automated failure and
recovery tests; unsupported surfaces are absent or unmistakably disabled.

## Phase 4 — Canonical client and legacy retirement

- [ ] Keep `client-web` as the single canonical public client.
- [ ] Migrate or archive `web-wallet` and `mobile-wallet`; remove them from
  public release paths.
- [ ] Complete transaction history with pagination and failure handling.
- [ ] Complete IBC transfer UX, status tracking, timeout handling, and recovery
  messaging.
- [ ] Connect the real audited ZKP prover and remove preview-only dead paths.
- [ ] Verify wallet creation, import, locking, signing, and key-storage safety.
- [ ] Add accessibility, responsive-layout, low-bandwidth, and browser support
  checks.
- [ ] Split oversized routes and establish a bundle-performance budget.

**Exit gate:** one maintained client completes every supported critical flow
against the rollout candidate; legacy clients cannot be mistaken for supported
software.

## Phase 5 — Quality and security depth

- [ ] Raise critical-path coverage for the root package, DEX, and governance
  modules, prioritizing rollback, authorization, arithmetic, and failure paths.
- [ ] Add end-to-end tests from wallet/client actions through committed chain
  state and query results.
- [ ] Add property, fuzz, invariant, and malformed-genesis tests where they
  provide stronger guarantees than example tests.
- [ ] Test concurrent submissions, duplicate messages, replay attempts, and
  deterministic restart behavior.
- [ ] Maintain dependency, static-analysis, secret, and supply-chain gates.
- [ ] Refresh the threat model for consensus, governance, DEX, ZKP, IBC,
  operator, and client boundaries.
- [ ] Complete an independent security review and resolve every critical/high
  finding.

**Exit gate:** the release matrix is reproducibly green, critical paths have
defensible coverage, and no unresolved critical/high security finding remains.

## Phase 6 — Operations and observability

- [ ] Add separate liveness and readiness signals for node operation.
- [ ] Define structured logs without secrets, mnemonic material, or private
  transaction data.
- [ ] Export consensus, peer, block, transaction, invariant, resource, and
  application metrics.
- [ ] Provide dashboards and actionable alert thresholds.
- [ ] Define service objectives and escalation ownership.
- [ ] Deploy the intended production topology, including seed nodes, sentries,
  validator isolation, RPC exposure, firewalling, and abuse protection.
- [ ] Write and rehearse incident, validator failure, key compromise, backup,
  restore, upgrade, rollback, and chain-halt runbooks.
- [ ] Validate resource limits, disk growth, log retention, and capacity
  assumptions under sustained load.

**Exit gate:** operators can detect, diagnose, contain, recover, and document a
failure using the published runbooks and telemetry.

## Phase 7 — Release engineering and staged rollout

- [ ] Produce reproducible binaries and container images from a tagged commit.
- [ ] Publish signed artifacts, checksums, software bill of materials (SBOM),
  provenance, and dependency reports.
- [ ] Pin release toolchains and document supported platforms.
- [ ] Provide installation, configuration, migration, upgrade, rollback, and
  uninstallation instructions.
- [ ] Publish release notes with compatibility and breaking-change statements.
- [ ] Freeze and independently review chain ID, genesis, consensus parameters,
  governance authorities, initial validator set, and all initial allocations.
- [ ] Re-run supply, balance, escrow, DEX, and validator-power checks against
  the exact rollout genesis artifact.
- [ ] Run a private multi-validator testnet and complete failure drills.
- [ ] Run a public testnet or controlled canary with monitoring and a defined
  rollback window.
- [ ] Freeze the release candidate while final evidence is reviewed.
- [ ] Record an explicit go/no-go decision and accountable approvers.

**Exit gate:** the exact signed release candidate survives the staged rollout
and all earlier phase gates remain satisfied.

## Rollout sequence

1. Reproducible local and CI release candidate.
2. Private multi-validator testnet with disaster-recovery drills.
3. Public testnet or tightly controlled canary with active monitoring.
4. Release freeze, independent evidence review, and explicit go/no-go record.
5. Public-network rollout only after every blocking checkbox is complete.

## Final go/no-go checklist

- [ ] All seven phase exit gates have linked evidence.
- [ ] CI and security workflows are green on the tagged release commit.
- [ ] No unresolved critical/high security or privacy finding exists.
- [ ] Real ZKP submission is compatible, audited, and fail-closed.
- [ ] Disaster recovery, upgrade, and rollback are independently repeatable.
- [ ] Monitoring, alerting, incident ownership, and runbooks are active.
- [ ] Release artifacts are reproducible, signed, and accompanied by SBOM and
  provenance.
- [ ] The maintained client and supported protocol surface are unambiguous.
- [ ] A documented go/no-go approval authorizes the staged rollout.

Green CI alone does not satisfy this checklist. Until all gates pass,
TrueRepublic remains a recovery-stage project.
