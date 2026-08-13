# TrueRepublic Cross-System Threat Model

Model: `truerepublic-cross-system-threat-model` · Version:
`truerepublic.threat-model/v1` · Updated: 2026-08-09

The canonical, machine-readable register is
[`configs/security/threat-model.json`](../../configs/security/threat-model.json).
This document explains the model; where prose and JSON differ, the JSON wins
and the repository contract test (`threat_model_repository_test.go`) fails
closed on drift.

**`production_ready=false`.** TrueRepublic remains a recovery-phase
repository. This threat model is a versioned self-assessment of repository
evidence. It is **not an independent external audit**, and it does not
authorize or claim consensus-state, cryptography, wallet-custody, key,
signing, funds, deployment, mainnet, migration, release, or production
activity.

## Scope and safety boundary

The model covers maintained assets, actors, trust boundaries, and data flows
visible in this repository. The repository holds no secrets, private keys,
real topology, or customer data; all committed configuration (topology,
incident, capacity, deployment evidence) is synthetic. Threat severities and
statuses describe repository-verifiable evidence only.

### Evidence classes

Every threat entry distinguishes three evidence classes:

- **Verified repository controls** — controls backed by merged code, tests,
  contracts, and CI evidence cited under `evidence` paths in the JSON.
- **Deferred or blocked work** — threats whose mitigation depends on open
  engineering tracked under umbrella issues GH-7 (audit/review parent) or
  GH-29 (rollout execution tracker).
- **External or live evidence** — anything requiring a real network, private
  infrastructure, external reviewers, real funds, or production systems. The
  register never claims this class; entries that need it carry `deferred` or
  `blocked` status and map to GH-7 or GH-29.

## Maintained assets

- **PNYX ledger** — canonical bank state, the 21,000,000 PNYX supply cap,
  module escrow, and treasury accounting.
- **Governance state** — domains, proposals, votes, validator claims, stake,
  and slash state in `x/truedemocracy`.
- **DEX custody** — pool reserves, provider-owned LP shares, the asset
  registry, and swap burn settlement.
- **Consensus identity** — CometBFT validator identity, consensus-key history
  and revocation, and signer state (operator-held, never in the repository).
- **ZKP anonymity** — pinned genesis verification-key state, identity Merkle
  roots, and the active nullifier set.
- **Maintained client** — the `client-web` browser client; user key material
  stays in user custody.
- **Node software** — the `truerepublicd` daemon, the deterministic build
  contract, and derived release artifacts.
- **CI supply chain** — workflows, SHA-pinned Actions, lockfiles, and the
  security gate contract.
- **Operations evidence** — structured logs, private metrics, dashboards,
  alert rules, and recovery drill evidence.

## Actors

Adversaries: an external network attacker; a malicious or compromised
validator; a malicious governance/treasury/DEX participant; a malicious or
compromised RPC provider serving the browser client; a supply-chain attacker
targeting dependencies, toolchains, or Actions; and a compromised operator.
Legitimate roles: reviewed maintainers acting through protected CI and
append-only Bridge records, and node operators running seed, sentry,
validator, RPC, or private roles under documented policy.

## Trust boundaries and data flows

- **P2P network** — untrusted internet versus sentry-shielded validators.
  Flows: consensus votes/proposals/blocks through sentries, sanctioned peer
  exchange, trusted state-sync snapshots.
- **Consensus ↔ application (ABCI)** — deterministic block transitions,
  ABCI++ misbehavior evidence feeding slashing, and export/import or
  migration state across binary versions.
- **Module ↔ bank custody** — exact escrow/stake/fee/reward/slash bank
  movements, DEX settlement and canonical burns, every-block crisis
  invariants.
- **ZKP prover ↔ on-chain verifier** — versioned chain/proposal/rating-bound
  signals, nullifier publication, genesis-pinned verification key. No real
  prover exists; clients stay fail-closed.
- **Client ↔ RPC** — registered protobuf gRPC-over-ABCI queries and the
  centralized simulate/sign/deliver flow. The configured provider is trusted
  for completeness until a light client exists; this trust is an explicit
  residual, not a control.
- **Repository ↔ CI** — SHA-pinned Actions with read-only permissions,
  lockfile-bound dependency resolution, fail-closed vulnerability/static/
  secret gates.
- **Operator ↔ node** — generated-key genesis bootstrap, sanitized backups
  excluding key material, secret-minimized logs and private-only metrics.
- **IBC counterparties** — no live packet flow exists; the boundary is a
  documented non-goal today.

## Threat register summary

Statuses: `mitigated`, `deferred`, `blocked`, `accepted`, `not_applicable`.
Every critical/high residual maps to GH-7 or GH-29. Full attack paths,
controls, evidence paths, residual risk, owners, and next gates live in the
JSON register.

### consensus_p2p

- **TM-CON-001** (high/medium, deferred → GH-29): P2P eclipse or flooding of
  validators. Verified: GH-71 role policy and GH-89 topology qualification.
  Deferred: no applied firewall, proxy, DNS, or private live topology
  evidence exists.
- **TM-CON-002** (critical/medium, deferred → GH-29): consensus-breaking
  migration or partially applied migration without rehearsed rollback.
  Verified: compatible replacement/export drills plus GH-184 two-thirds
  governance, deterministic old-binary halt, cached-write failure rollback,
  fixed-candidate recovery, and exact-once four-validator restart. Residual:
  pre-GH-184 store introduction, clean private-infrastructure reproduction,
  independent review, and production operation remain unqualified.

### governance_identity

- **TM-GOV-001** (high/medium, mitigated): spoofed identity or authority.
  Verified: signer-derived identity, exact bank escrow, cached atomic
  settlement, chain-authority registry checks.
- **TM-GOV-002** (high/medium, deferred → GH-7): front-running of anonymous
  rating rewards. Verified: rewards deferred, bound proof/signature payloads,
  fail-closed clients. Deferred: no reviewed recipient-binding design.

### token_treasury_dex

- **TM-TOK-001** (critical/low, mitigated): supply-cap violation.
  Verified: one cap-checked issuance service, pre-mutation genesis cap
  checks, every-block crisis invariants.
- **TM-TOK-002** (high/low, mitigated): DEX reserve drain or unauthorized
  registry mutation. Verified: module-bank custody exclusivity,
  provider-owned LP shares, authority-gated registry.

### zkp_privacy

- **TM-ZKP-001** (critical/high, blocked → GH-29): no production-qualified Groth16 prover.
  Verified: fail-closed clients, pinned genesis VK, no randomized consensus
  setup, and GH-206 real synthetic Go/WASM-to-native-verifier compatibility.
  Blocked: no production ceremony, no reproducible proving artifacts for
  production, no audited submission path, or real-network browser-to-chain
  evidence is claimed.
- **TM-ZKP-002** (high/medium, deferred → GH-7): compromised or unaudited
  trusted setup or circuit. Verified: genesis ceremony artifacts pinned as
  trust anchor. Deferred: ceremony provenance and independent cryptographic,
  privacy, and trusted-setup review are absent and not claimed.

### ibc_upgrades

- **TM-IBC-001** (high/medium, deferred → GH-29): IBC value
  transfer or relay failure. Verified locally: two proof-driven TrueRepublic
  chains complete client/connection/channel handshakes, native escrow, voucher
  mint, acknowledgement, timeout refund, replay-safe duplicate handling, and
  pending-ack database recovery. GH-178 additionally verifies a committed
  closed counterparty end, proof-driven close-confirm and timeout-on-close,
  exactly-once refund, persistent recovery, and replacement-channel transfer.
  GH-181 verifies that the supported in-place compatible binary replacement
  preserves open IBC and pending packet state, then completes pending/fresh ACK
  and timeout/refund paths exactly once. GH-184 adds the governed application
  migration failure/recovery path. Residual: external relayer/counterparty,
  pre-existing-chain store introduction, IBC client upgrades, and arbitrary
  cross-version evidence remain absent, so rollout remains exit-gated.
- **TM-IBC-002** (high/medium, mitigated → GH-29): unsupported staking and
  distribution compatibility surfaces. GH-187 replaces ambiguous success-like
  stubs with deterministic fail-closed IBC/CosmWasm adapters and proves
  `x/staking` and `x/distribution` remain absent from module, store, genesis,
  gRPC, message-router, and CLI surfaces. Residual: upstream interface changes
  must preserve this boundary under dependency upgrades.

### client_wallet_rpc

- **TM-CLI-001** (high/medium, deferred → GH-29): malicious RPC provider.
  Verified: registered gRPC-only transport, fail-closed typed errors, capped
  sanitized failure output. Deferred: the configured provider is still
  trusted; no browser light client exists.
- **TM-CLI-002** (critical/medium, deferred → GH-7): browser wallet custody.
  Verified: exact fail-closed transaction registry, centralized signing
  boundary, no seed-phrase handling in the test matrix. Deferred: no
  independent wallet custody or cryptographic review exists or is claimed.

### operations_observability

- **TM-OPS-001** (high/medium, deferred → GH-29): missed incidents without
  paging, SLOs, or live monitoring. Verified: immutable dashboard, eleven
  deterministic alerts, health probes, incident rehearsals. Deferred:
  external paging, production SLOs, private live rehearsal, and soak remain
  open.
- **TM-OPS-002** (medium/low, mitigated): secret leakage through logs,
  metrics, or configuration. Verified: secret-minimized structured logging,
  reviewed-pattern redaction, fail-closed secret scanning.

### dependencies_ci

- **TM-DEP-001** (high/medium, accepted): reachable dependency
  vulnerabilities without upstream fixes. Verified: fail-closed
  govulncheck/cargo-audit/npm gates, exact-ID 30-day-bounded exceptions.
  Accepted residual: four no-fix Go findings and monitored cargo warnings
  stay visible under the weekly cadence.
- **TM-DEP-002** (high/low, mitigated): CI poisoning through mutable Actions
  or credential persistence. Verified: full-SHA pins, contents-read
  permissions, per-job timeouts, contract-tested rejection of mutable or
  bypass shapes.

### release_artifacts

- **TM-REL-001** (high/medium, blocked → GH-29): unsigned release artifacts
  without SBOM or provenance. Verified: deterministic Linux build contract
  and reproducible-build CI. Blocked: signing, SBOM, provenance, and
  publishing do not exist and are not claimed.
- **TM-REL-002** (medium/low, mitigated): non-deterministic build drift.
  Verified: pinned toolchains, repository-owned deterministic build script,
  lockfiles bound into the security contract.

## Explicitly not claimed

This model does not claim: a real ZKP prover or ceremony; any independent
external audit; wallet custody proof; IBC or testnet evidence; private
topology, firewall, or DNS deployment; release artifact signing or
provenance; production or mainnet readiness; or real-funds/real-key safety.
Entries needing those carry a `deferred` or `blocked` status and map to GH-7
or GH-29.

## Review and update triggers

Update the JSON register (and this document in the same change) when any of
the following occurs:

1. A GH-7 or GH-29 gate closes or a new umbrella-level risk is opened.
2. A new maintained asset, actor, trust boundary, or data flow is added to
   the repository.
3. A security gate, audit, or drill changes a threat's verified controls or
   residual severity.
4. A dependency exception is approved, renewed, or expires.
5. Any claim about ZKP, IBC, client custody, operations, or release evidence
   changes state.
6. At minimum, every 90 days from the `updated` field, even if no trigger
   fired.

Version bumps follow `truerepublic.threat-model/vN`; the repository contract
test pins the exact current version and fails closed on schema, enum,
evidence-path, umbrella-mapping, or parity drift.
