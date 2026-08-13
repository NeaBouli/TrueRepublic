# TrueRepublic Road to Rollout

Updated: 2026-08-12

TrueRepublic has a recovered and CI-verified v0.4 engineering foundation. It
is **not production-ready, mainnet-ready, or approved for real funds or keys**.
This document is the public checklist from the current recovery baseline to a
controlled rollout. Progress is tracked in
[GitHub issue #29](https://github.com/NeaBouli/TrueRepublic/issues/29); the
parent recovery record remains
[issue #4](https://github.com/NeaBouli/TrueRepublic/issues/4).

Published progress arithmetic follows the canonical issue #29 tracker. This
detailed roadmap splits its combined dashboard/alerts/objectives/ownership
item into two evidence bullets, so raw checkbox counts in this file are
intentionally more granular than the public 59-item tracker.

## Current baseline

- The ordered recovery merge chain is on `main`.
- The maximum supply is fixed at 21,000,000 PNYX.
- The source of truth records 1,746 recovery-verified tests: 1,461 Go, 26 Rust,
  and 259 maintained-client tests. GH-121's real registered browser-query
  boundary and GH-115's local client-chain delivery
  case are separately gated and excluded from this arithmetic, as is GH-172's
  real contention/replay/restart process harness. GH-175/GH-178/GH-181's
  two-chain IBC packet, channel, and compatible binary-restart recovery
  harnesses are also separately gated. GH-184 adds the opt-in four-validator
  governed upgrade/failure/recovery process harness.
- Ledger, escrow, issuance, DEX custody, genesis, runtime invariants, ZKP
  statement binding, node persistence, and the safe operator-init boundary
  have CI-backed recovery evidence.
- GH-89/PR #90 add repository-owned cross-node topology and abuse-control
  qualification evidence; real private deployment remains open.
- GH-93/PR #94 add a strict secret-free incident-command and eight-scenario
  rehearsal contract; private live operator rehearsal remains open.
- GH-97/PR #98 add bounded four-validator sustained-load, resource, retention,
  restart, and ledger evidence; private-environment sizing and soak remain open.
- GH-101/PR #103 add a strict secret-free digest-bound deployment-evidence
  envelope and offline verifier; private live deployment evidence remains open.
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
- [x] Prove compatible rolling binary replacement on persisted validator state
  ([GH-53](https://github.com/NeaBouli/TrueRepublic/issues/53)).
- [x] Prove rollback after a candidate fails before opening state, without
  replacing validator identity or regressing signer state
  ([GH-53](https://github.com/NeaBouli/TrueRepublic/issues/53)).
- [x] Implement and prove governance-controlled consensus-breaking state
  migrations and rollback after a partially applied migration
  ([GH-184](https://github.com/NeaBouli/TrueRepublic/issues/184)).
- [x] Define and test coupled validator-key/signer-state cold custody,
  single-signer failover, and compromise containment
  ([GH-55](https://github.com/NeaBouli/TrueRepublic/issues/55)).
- [x] Implement authenticated atomic consensus-key rotation, permanent old-key
  revocation, and bootstrap operator-authority separation
  ([GH-56](https://github.com/NeaBouli/TrueRepublic/issues/56)).
- [x] Define and regression-test seed, sentry, validator, RPC/API, and private
  node peer/listener/firewall/rate-limit policy
  ([GH-71](https://github.com/NeaBouli/TrueRepublic/issues/71)).
- [x] Complete final review and merge of the bounded fresh-genesis migration
  for pre-GH-56 coupled validator authorities through
  [PR #69](https://github.com/NeaBouli/TrueRepublic/pull/69); no independent
  legacy governance anchor exists, so this is not retroactive governance
  authorization ([GH-61](https://github.com/NeaBouli/TrueRepublic/issues/61)).
- [x] Preserve inactive, excluded, jailed, and under-staked validator claims in
  round-trip-safe export/import state
  ([GH-60](https://github.com/NeaBouli/TrueRepublic/issues/60)).

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

- [x] Run deterministic two-chain proof-relay tests for native PNYX transfer,
  escrow, destination voucher minting, and commitment cleanup ([GH-175](https://github.com/NeaBouli/TrueRepublic/issues/175)).
- [x] Test acknowledgements, timeout refunds, open unordered channel state,
  idempotent duplicate receive/ACK/timeout replay, and relay
  interruption ([GH-175](https://github.com/NeaBouli/TrueRepublic/issues/175)).
- [x] Reopen a persisted destination application while an acknowledgement is
  pending and complete the relay without balance or invariant drift (GH-175).
- [x] Test a committed counterparty channel closure, proof-driven
  close-confirm and timeout-on-close, exactly-once refund, database recovery,
  and transfer/acknowledgement on a distinct replacement channel
  ([GH-178](https://github.com/NeaBouli/TrueRepublic/issues/178)). ICS-20
  intentionally rejects user-initiated close, so this does not claim a public
  channel-close authority or external relayer qualification.
- [x] Test IBC packet behavior across the currently supported compatible,
  in-place application binary replacement without height reset or genesis
  export/import ([GH-181](https://github.com/NeaBouli/TrueRepublic/issues/181)).
  This does not qualify `x/upgrade`, consensus migrations, arbitrary version
  changes, a daemon relayer, or public-network operation.
- [x] Complete the supported fresh-genesis application-upgrade path and its
  immutable-snapshot two-thirds governance controls (GH-184).
- [x] Replace remaining staking and distribution compatibility stubs with
  explicit fail-closed adapters (GH-187).
- [x] Document and regression-test the exact supported Cosmos/IBC feature
  boundary (GH-187).

**Exit gate:** supported IBC and upgrade flows pass automated failure and
recovery tests; unsupported surfaces are absent or unmistakably disabled.

## Phase 4 — Canonical client and legacy retirement

- [x] Keep `client-web` as the single canonical public client, with one exact
  fail-closed custom transaction registry and local client-to-chain proof
  ([GH-115](https://github.com/NeaBouli/TrueRepublic/issues/115)).
- [x] Retire and remove `web-wallet` under GH-112; keep the GH-102-retired
  mobile prototype absent from public release paths.
- [x] Complete newest-first submitted-transaction history with server
  pagination, committed-failure preservation, typed unavailable/timeout/
  protocol/decode handling, stale-wallet protection, and real disposable-chain
  evidence ([GH-131](https://github.com/NeaBouli/TrueRepublic/issues/131)).
  Incoming-only activity remains explicitly outside this submitted-history
  boundary.
- [x] Complete canonical native ICS-20 signing, strict open-channel/amount/
  balance/timeout validation, wallet-scoped status persistence, and manual
  source-chain acknowledgement/timeout recovery without automatic resubmission
  ([GH-190](https://github.com/NeaBouli/TrueRepublic/issues/190)). External
  relayer and public-counterparty qualification remain Phase 3 work.
- [ ] Connect the real audited ZKP prover and remove preview-only dead paths.
- [ ] Verify wallet creation, import, locking, signing, and key-storage safety.
- [x] Add accessibility, responsive-layout, low-bandwidth, and browser support
  checks for the safe unauthenticated maintained-client surfaces, with a pinned
  protected Chromium/Firefox/WebKit matrix and explicit authenticated-flow
  boundary ([GH-132](https://github.com/NeaBouli/TrueRepublic/issues/132),
  [PR #136](https://github.com/NeaBouli/TrueRepublic/pull/136)).
- [x] Split all 20 page routes and enforce raw/gzip entry, route, chunk, and
  total-JavaScript budgets during every maintained-client build
  ([GH-128](https://github.com/NeaBouli/TrueRepublic/issues/128)).

**Exit gate:** one maintained client completes every supported critical flow
against the rollout candidate; legacy clients cannot be mistaken for supported
software.

## Phase 5 — Quality and security depth

- [x] Raise critical-path coverage for the root package, DEX, and governance
  modules, prioritizing rollback, authorization, arithmetic, and failure paths
  ([GH-139](https://github.com/NeaBouli/TrueRepublic/issues/139)).
- [x] Add end-to-end tests from maintained-client signing actions through
  committed local-chain delivery for the supported transaction families
  ([GH-115](https://github.com/NeaBouli/TrueRepublic/issues/115)); expanded
  state/query assertions remain future depth work.
- [x] Add property, fuzz, invariant, replay, malformed-genesis, and focused
  race tests where they provide stronger guarantees than example tests
  ([GH-145](https://github.com/NeaBouli/TrueRepublic/issues/145)).
- [x] Test concurrent submissions, duplicate messages, exact transaction
  replay attempts, and deterministic same-home restart behavior
  ([GH-172](https://github.com/NeaBouli/TrueRepublic/issues/172)). This
  detailed evidence item remains folded into GH-29's canonical quality-depth
  tracker item and therefore does not create a sixtieth rollout unit.
- [x] Wire ABCI++ misbehavior and last-commit data into the economic slashing
  handlers and test evidence-window custody after validator removal
  ([GH-59](https://github.com/NeaBouli/TrueRepublic/issues/59)).
- [x] Maintain dependency, static-analysis, secret, and supply-chain gates
  with immutable Action pins, exact scanner/toolchain versions, blocking
  static and secret scans, lockfile enforcement, bounded weekly dependency
  updates, and a repository-owned fail-closed contract
  ([GH-148](https://github.com/NeaBouli/TrueRepublic/issues/148)).
- [x] Refresh the threat model for consensus, governance, DEX, ZKP, IBC,
  operator, and client boundaries with a versioned register and fail-closed
  repository contract
  ([GH-169](https://github.com/NeaBouli/TrueRepublic/issues/169)).
- [ ] Complete an independent security review and resolve every critical/high
  finding.

**Exit gate:** the release matrix is reproducibly green, critical paths have
defensible coverage, and no unresolved critical/high security finding remains.

## Phase 6 — Operations and observability

- [x] Add separate liveness and readiness signals for node operation
  ([GH-74](https://github.com/NeaBouli/TrueRepublic/issues/74)).
- [x] Define structured logs without secrets, mnemonic material, or private
  transaction data
  ([GH-77](https://github.com/NeaBouli/TrueRepublic/issues/77)).
- [x] Export consensus, peer, block, transaction, invariant, resource, and
  application metrics through a private two-source baseline
  ([GH-80](https://github.com/NeaBouli/TrueRepublic/issues/80)).
- [x] Provide dashboards and actionable alert thresholds
  ([GH-85](https://github.com/NeaBouli/TrueRepublic/issues/85),
  [PR #86](https://github.com/NeaBouli/TrueRepublic/pull/86)).
- [x] Define recovery/testnet service objectives and role-based escalation
  ownership
  ([GH-85](https://github.com/NeaBouli/TrueRepublic/issues/85),
  [PR #86](https://github.com/NeaBouli/TrueRepublic/pull/86)).
- [ ] Deploy the intended production topology, including seed nodes, sentries,
  validator isolation, RPC exposure, firewalling, and abuse protection.
  GH-89/PR #90 provide the strict synthetic qualification contract and CI
  validator. GH-101/PR #103 provide the strict offline deployment-evidence
  envelope and verifier; real private inventory and deployment evidence remain
  required.
- [x] Write and synthetically rehearse incident, validator failure, key
  compromise, backup, restore, upgrade, rollback, and chain-halt runbooks
  ([GH-93](https://github.com/NeaBouli/TrueRepublic/issues/93),
  [PR #94](https://github.com/NeaBouli/TrueRepublic/pull/94)); private live
  operator rehearsal remains part of deployment and final exit evidence.
- [x] Validate repository-side resource limits, disk growth, log retention,
  and capacity assumptions under bounded sustained load
  ([GH-97](https://github.com/NeaBouli/TrueRepublic/issues/97),
  [PR #98](https://github.com/NeaBouli/TrueRepublic/pull/98)); production sizing,
  multi-day soak, and private-environment evidence remain rollout exit work.

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
