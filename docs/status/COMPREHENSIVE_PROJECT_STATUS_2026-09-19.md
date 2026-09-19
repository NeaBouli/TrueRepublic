# TrueRepublic Comprehensive Project Status and Completion Backlog

**Checkpoint:** 2026-09-19 12:41 EEST

**Status issue:** [GH-316](https://github.com/NeaBouli/TrueRepublic/issues/316)

**Canonical baseline:** `origin/main` at
`f5de5a150b9ff0edc51d68411b7829e78bcf78e6`

**Public version:** v0.4.0 recovery baseline

**Production readiness:** **false**

**Basic rollout:** **36/59 complete (61.0%)**

**Phase work:** **36/51 complete (70.6%)**

**Purpose:** durable, evidence-backed checkpoint of what exists, what has been
verified, what is unfinished, and what must happen before a controlled rollout
or any later Alpha/V4 completion claim.

---

## 1. Executive verdict

TrueRepublic is a substantial recovered engineering foundation, not a finished
or production-approved network. The repository has a real Cosmos SDK chain,
governance and treasury modules, a bank-backed DEX, CosmWasm and IBC
integration, a maintained React/TypeScript browser client, extensive recovery
and release-evidence tooling, and 2,462 counted recovery-verification cases.
The 21,000,000 PNYX cap and major ledger/escrow/custody invariants have
machine-checked evidence.

The project must nevertheless remain `production_ready=false` because:

1. the external audit report identifies seven unresolved Critical/High
   findings, including a reachable Critical domain-admin corruption defect;
2. the maintained browser ZKP path and production trusted setup are incomplete;
3. one dependency advisory currently makes PR #302's `rust-audit` red;
4. real private topology, operational rehearsal and sizing evidence do not
   exist;
5. tagged, signed and published release artifacts do not exist;
6. rollout genesis, public/private testnets, canary, release freeze and
   accountable go/no-go are incomplete; and
7. eight final rollout exit criteria are all still open.

The correct immediate order is audit remediation and safety recovery, not V4-1
feature expansion. The next implementation task is GH-304.

---

## 2. Scope and definitions

This report separates four different completion meanings:

| Level | Meaning | Current state |
|---|---|---|
| Recovery foundation | Repository builds, core recovery flows and evidence exist | Substantially complete |
| Basic-59 rollout candidate | All 51 phase tasks plus eight exit criteria pass | 36/59; not complete |
| Controlled production rollout | Signed candidate survives private/public staged operation and explicit go/no-go | Not started |
| Long-term product roadmap | Sovereign Alpha, V4 edge runtime and optional ballot system are implemented and qualified | Architecture only, except unwired V4-0 protocol |

“Project completely finished” is therefore not one checkbox. The first
well-defined finish line is **Basic-59 at 59/59 with production authorization**.
Sovereign Alpha, V4 and the optional ballot engine remain separate later
programs and must not be counted as Basic rollout progress.

### Source-of-truth order

1. GitHub `origin/main` and [GH-29](https://github.com/NeaBouli/TrueRepublic/issues/29).
2. `docs/status.json` and `docs/ROLLOUT_ROADMAP.md`.
3. Project-local `AGENTS.md`, `BRIDGE.md` and `docs/agent-bridge/`.
4. Open GitHub issues and pull requests.
5. The external audit reports in PR #302 and umbrella register #303.
6. Local unfinished branches only as work-in-progress evidence, never as
   canonical project truth.

---

## 3. Current repository and public state

### 3.1 Canonical main

| Item | Verified state |
|---|---|
| Repository | [NeaBouli/TrueRepublic](https://github.com/NeaBouli/TrueRepublic) |
| Visibility | Public |
| Default branch | `main` |
| Main commit | `f5de5a150b9ff0edc51d68411b7829e78bcf78e6` |
| Last main closeout | GH-297 ZKP protocol freeze |
| License | Apache-2.0 for maintained source and maintained documentation |
| Attribution | “TrueRepublic contributors”; individual contributors retain copyright |
| Exclusions | Brand assets, artwork, historical PDFs, archived historical evidence and third-party material until provenance permits inclusion |
| Canonical client | `client-web` |
| Retired clients | `web-wallet` and `mobile-wallet` exist only in Git history |
| Release state | `recovery_active` |
| Production/mainnet | false / not authorized |
| Open issues | 17 including this checkpoint: #29, #232, #300, #303-#316 |
| Open pull requests | Two: #301 and #302 |

The repository is monorepo-style: one Git repository contains the Go daemon
and modules, Rust/CosmWasm workspace, TypeScript client, V4 protocol package,
release/security tooling, documentation, landing page and wiki source. It is
not a Turborepo-style JavaScript-only monorepo.

### 3.2 Local worktrees

Two local states must not be confused:

- The canonical reporting checkout for GH-316 starts clean from exact
  `origin/main`.
- The original checkout remains on
  `feature/GH-300-browser-zkp-prover` with preserved modified and untracked
  browser-prover files. Those changes are unfinished, uncommitted, not on main
  and not rollout evidence. They were deliberately not reset or folded into
  this report.

No data loss was observed. The isolation exists specifically to avoid treating
unfinished GH-300 files as canonical history.

### 3.3 Public documentation and community state

- README, landing, wiki and `docs/status.json` publish 2,462 tests and 36/59.
- Apache-2.0 is detected by GitHub.
- The project is community-governed and has no central corporate owner.
- Developer BTC support is published in README and `.github/FUNDING.yml`:
  `bc1p2kh7reqf7l5zmssk8kdx2fx0q6suzfmsgdrphl9wjsxarns56jkq4tqvjm`.
- The current public documentation predates the 2026-09-15 audit remediation.
  PR #302 and issues #303-#315 contain the audit truth, but the report PR is not
  merged and no finding is resolved merely because it has been catalogued.
- GH-316 adds only this durable checkpoint. It changes neither product behavior,
  rollout arithmetic nor production status.

---

## 4. Implemented and verified foundation

### 4.1 Technology and test inventory

| Area | Current recorded baseline |
|---|---|
| Go | 1.26.6 |
| Cosmos SDK | v0.50.15 |
| CometBFT | v0.38.26 |
| Wasmd | v0.53.4 |
| WasmVM | v2.2.8 |
| IBC-Go | v8.7.0 |
| React / TypeScript / Vite | React 18.2 / TypeScript 5.9 / Vite 8.2.2 on main |
| Counted tests | 2,462 total |
| Go | 2,117 |
| Rust | 26 |
| Maintained client | 319 |

Important qualification: several multi-validator, IBC, browser and release
process harnesses are separately gated and intentionally excluded from the
2,462 arithmetic. The count is useful evidence, not proof of production safety.

### 4.2 Token and ledger foundation

- Maximum supply is fixed at 21,000,000 whole PNYX
  (21,000,000,000,000 `upnyx`).
- Single-issuance and per-block crisis/invariant enforcement exists.
- Bank supply, governance escrow, DEX reserves and LP conservation are
  reconciled by registered invariants.
- Domain treasury and bank custody have recovery evidence.
- Front-running-safe ZKP reward-recipient binding is implemented on chain.
- Direct anonymous reward payout remains publicly linkable to its recipient;
  it is not shielded-payment privacy.

### 4.3 Consensus, recovery and validator evidence

Completed evidence includes:

- reproducible four-validator consensus/recovery;
- join, replacement, restart/catch-up and leave power-zero behavior;
- partition recovery without app-hash divergence;
- trusted snapshot state sync;
- sanitized backup, restore, export and import;
- compatible persisted-state binary replacement and fail-before-open rollback;
- bounded governed fresh-genesis v0.4.1 migration/recovery;
- validator identity cold failover;
- authenticated consensus-key rotation and permanent old-key revocation;
- ABCI++ slashing evidence and custody holds;
- inactive validator round-trip preservation; and
- seed/sentry/validator/RPC policy enforced through repository contracts.

These are strong repository and synthetic/process results. They do not replace
a real private deployment, real operator rehearsal or production topology.

### 4.4 Protocol, IBC and DEX foundation

- Native PNYX ICS-20 transfer, voucher minting, escrow and cleanup are tested.
- Acknowledgement, timeout refund, duplicate replay and interruption paths are
  tested.
- Pending-ACK database recovery, channel-close timeout/refund and replacement
  channel flows are tested.
- Compatible binary restart with pending IBC state is tested.
- The supported Cosmos/IBC boundary is documented and unsupported standard
  staking/distribution surfaces fail closed.
- The DEX uses bank-backed custody, provider-indexed LP ownership, multi-asset
  pools, routing and conservation evidence.

Still unqualified: external relayers, real public counterparties, arbitrary
cross-version compatibility, IBC client upgrades and real public operation.

### 4.5 Maintained Beta client

The single maintained client is `client-web`:

- React/TypeScript/Vite with 23 routes, 44 components, nine stores and 17
  services in the current machine-readable baseline;
- fail-closed custom transaction registry;
- local bank, governance, membership, identity and DEX delivery evidence;
- registered protobuf gRPC-over-ABCI query transport;
- transaction history with committed-failure preservation;
- native ICS-20 transfer and source-chain recovery messaging;
- BIP-39 import validation and versioned AES-GCM/PBKDF2 local wallet storage;
- wallet/chain/account session invalidation;
- accessibility, responsive, low-bandwidth and browser gates;
- route splitting and bundle budgets.

Residual client limitations include same-origin/XSS exposure, no hardware
custody qualification, single-RPC trust, `prove:false` query reliance,
plaintext ZKP identity-secret storage in the preview path, no production ZKP
submission and no supported native mobile application.

### 4.6 Security, release and operations foundation

Completed repository capabilities include:

- dependency, vulnerability, static, secret and supply-chain gates;
- versioned cross-system threat model;
- independent-review readiness verifier;
- locked CI/release tool bootstrap;
- deterministic daemon and repeated OCI evidence;
- exact-commit and simulated-tag candidate aggregation;
- cross-run metadata comparison;
- offline release evidence, checksums/SBOM/provenance schema;
- liveness/readiness, secret-safe structured logs and metrics;
- dashboards, alerts, recovery objectives and escalation roles;
- incident/backup/upgrade/rollback runbooks; and
- bounded sustained-load/resource/retention evidence.

These prepare a release. They do not create a real tag, signature, published
artifact, deployment, production inventory or authorization.

### 4.7 Architecture-only and parallel foundations already completed

- GH-215: decision-ready Sovereign Alpha architecture.
- GH-231: optional domain ballot architecture.
- GH-236: edge-native V4 architecture with TRChain as sole authority.
- GH-241: V4-0 canonical protocol library and vectors, deliberately unwired.
- GH-219: repository/community license decision and publication.

None of these completed architecture tasks adds Basic-59 rollout credit.

---

## 5. Exact Basic-59 arithmetic

The 59 items comprise **51 phase tasks plus eight rollout exit criteria**.

| Phase | Complete | Total | Open | Status |
|---|---:|---:|---:|---|
| 1 — Network and disaster recovery | 7 | 7 | 0 | Phase task list complete |
| 2 — Production ZKP and privacy | 2 | 7 | 5 | Major blocker |
| 3 — IBC and protocol completeness | 6 | 6 | 0 | Supported boundary complete |
| 4 — Canonical client and legacy retirement | 7 | 8 | 1 | Real audited ZKP path open |
| 5 — Quality and security depth | 5 | 6 | 1 | External Critical/High closure open |
| 6 — Operations and observability | 6 | 7 | 1 | Real production topology open |
| 7 — Release engineering and staged rollout | 3 | 10 | 7 | Staged release work open |
| Phase subtotal | **36** | **51** | **15** | 70.6% |
| Final exit criteria | **0** | **8** | **8** | All blocking |
| Overall | **36** | **59** | **23** | 61.0% |

Audit issues #304-#315 overlap these open rollout gates. They are not twelve
additional rollout checkboxes and must not be added to 23.

The Phase-2 roadmap also displays “keep anonymous submission fail-closed” as a
continuing safety control. It is mandatory but is not an eighth Phase-2 tracker
unit; GH-29 and `docs/status.json` define the canonical denominator as seven.

---

## 6. Detailed remaining Basic-59 work

### 6.1 Phase 2 — Production ZKP and privacy (five open)

1. **Compatible real Groth16 maintained-client prover.**
   GH-300 contains an unfinished compatibility-only browser integration.
   It must remain non-submittable and now depends on GH-309 safety work.
2. **Reproducible production proving and verification artifacts.**
   Circuit, CS, proving key, verification key and checksums must be built from
   a reproducible approved pipeline, not the current forge-capable fixture.
3. **Ceremony provenance and artifact rotation.**
   Participant assumptions, toxic-waste handling, transcript verification,
   contribution validation, circuit upgrade and emergency rotation must be
   governed and documented.
4. **Real browser-to-chain proof compatibility.**
   The maintained browser must create a proof under production artifacts that
   verifies through the real keeper path with replay/context/recipient
   adversarial coverage.
5. **Independent cryptographic, privacy and setup review.**
   This must cover circuit statements, nullifier/linkability model, ceremony,
   artifacts, client custody and chain integration.

The chain-side recipient binding and frozen/versioned protocol are already
complete. Submission must stay hard-disabled until all five open items pass.

### 6.2 Phase 4 — Canonical client (one open)

Connect the **real audited** ZKP path and remove or quarantine preview-only dead
paths. This is broader than GH-300 alone because:

- GH-300 uses test-only artifacts;
- GH-309 must eliminate plaintext identity custody and mock commitment
  registration;
- production artifacts/ceremony are still absent; and
- real submission cannot be enabled before independent review.

### 6.3 Phase 5 — Security depth (one open)

Complete the independent security review and resolve every Critical/High
finding. PR #302 supplies an external audit report, and GH-294 supplies the
lifecycle machinery, but the rollout's independent-review/closure gate remains
open until remediation and a separately attributable independent retest verify
the fixes.

Minimum blocking audit closures:

- GH-304 / TRR-01;
- GH-306 / TRR-02;
- GH-307 / TRR-03;
- GH-308 / TRR-13 through TRR-16 where applicable; and
- a fresh independent retest confirming closure rather than self-assertion.

### 6.4 Phase 6 — Operations (one open)

Deploy and qualify the intended production topology:

- real seed and sentry inventory;
- isolated validators behind at least the approved sentry boundary;
- bounded public RPC/API exposure;
- TLS, DNS, firewall, allowlist, rate/burst/body/timeout/concurrency controls;
- private metrics and paging;
- capacity sizing and multi-day soak;
- live incident, key compromise, backup, restore, upgrade and rollback drills;
- signed, secret-free deployment evidence consumed by the offline verifier.

This requires separately authorized infrastructure work. Repository contracts
cannot complete it alone.

### 6.5 Phase 7 — Release and staged rollout (seven open)

1. Produce reproducible binaries and images from a **real tagged commit**.
2. Sign and publish artifacts, checksums, SBOM, provenance and dependency
   reports.
3. Freeze and independently review chain ID, exact genesis, consensus and
   governance parameters, authorities, initial validator set and allocations.
4. Re-run supply, balance, escrow, DEX and validator-power checks against the
   exact frozen rollout genesis.
5. Operate a private multi-validator testnet and complete failure drills.
6. Operate a public testnet or controlled canary with monitoring and rollback
   window.
7. Complete final release review and accountable launch authorization,
   including both a frozen candidate and an explicit go/no-go record.

### 6.6 Eight final rollout exit criteria

All remain open:

1. Every phase gate has linked evidence and green CI.
2. No unresolved Critical/High security or privacy finding exists.
3. Real ZKP flow is compatible, audited and fail-closed.
4. Disaster recovery, upgrade and rollback are independently repeatable.
5. Monitoring, alerting, runbooks and incident ownership are active.
6. Exact release artifacts are reproducible, signed and documented.
7. Maintained client and supported protocol surface are unambiguous.
8. Explicit accountable go/no-go approval authorizes staged rollout.

---

## 7. External audit: complete 42-finding register

Source: PR #302 and GH-303. Totals are **3 Critical, 4 High, 15 Medium,
16 Low and 4 Informational**.

### 7.1 Critical

| Finding | Meaning | Remediation |
|---|---|---|
| TRR-01 | EndBlock admin election compares bech32 text with raw address bytes and stores text as raw bytes, corrupting domain admin, treasury/admin operations and export/import | GH-304 |
| TRR-13 | Legacy `core/treasury.rs` is an unauthenticated, unbacked fiction ledger if instantiated | GH-308 |
| TRR-14 | Example ZKP aggregator trusts caller-provided nullifiers and permits Sybil votes if instantiated | GH-308 |

TRR-01 is live-code and reachable. TRR-13/14 are conditional; the audit found
no evidence that those contracts were instantiated.

### 7.2 High

| Finding | Meaning | Remediation |
|---|---|---|
| TRR-02 | Cumulative 10% payout limit can prevent validator exit, including bootstrap validators | GH-306 |
| TRR-03 | One member can create multiple pseudonymous voting keys/commitments and receive repeated payouts | GH-307 |
| TRR-15 | Example DAO permits zero-vote execution and lacks voting-period enforcement; dangerous if deployed/admin | GH-308 |
| TRR-16 | Legacy core governance permits repeat voting and effectively only instantiator voting | GH-308 |

### 7.3 Medium

| Finding | Summary | Remediation |
|---|---|---|
| TRR-04 | Repeated stone moves can farm rewards | GH-306 |
| TRR-05 | Member exclusion does not promptly revoke anonymous authority | GH-307 |
| TRR-06 | Reserved governance domain may be unanchored/unreachable | GH-310 |
| TRR-07 | DEX registry authority can be an unspendable module account; hardcoded atom behavior | GH-310 |
| TRR-08 | Multiple whole-domain EndBlock scans create unbounded consensus work | GH-310 |
| TRR-17 | Example DEX bot has shadow orders without real escrow/settlement | GH-308 |
| TRR-18 | Contracts use monolithic state and lack cw2/migration discipline | GH-308 |
| TRR-19 | ZKP identity secret persists plaintext in localStorage while documentation claims encryption | GH-309 |
| TRR-20 | FNV mock commitments can be registered by real signed transaction, creating unusable leaves | GH-309 |
| TRR-23 | Maintained wallet client lacks Content-Security-Policy | GH-312 |
| TRR-24 | Operator/wiki instructions contain nonfunctional or unsafe commands/configuration | GH-313 |
| TRR-25 | Fictional `truerepublic.network` references create squatting/security-contact risk | GH-313 |
| TRR-29 | Whitepaper overclaims live contracts, anonymity, TSS and cross-chain behavior | GH-314 |
| TRR-30 | PoD/anti-whale description does not match implemented eligibility | GH-314 |
| TRR-31 | Financial/store-of-value/price-increase promises are unsafe and unsupported | GH-314 |

### 7.4 Low

| Finding | Summary | Remediation |
|---|---|---|
| TRR-09 | Production tree code includes async state mutation/fake nodes | GH-311 |
| TRR-10 | `CoinBurnRequired` is unsatisfiable/misnamed treasury credit | GH-311 |
| TRR-11 | Onboarding signature lacks chain binding | GH-311 |
| TRR-12 | Exclusion/admin, IBC recovery, zero-output LP burn and panic cluster | GH-311 |
| TRR-21 | Legacy rating path lacks range check and uses obsolete signing semantics | GH-311 |
| TRR-26 | Single-RPC trust, plaintext defaults, password/KDF/throttling limitations | GH-312 |
| TRR-27 | Landing hotlinks raw GitHub images despite local assets | GH-312 |
| TRR-32 | Stale test/module counts on multiple surfaces | GH-314 |
| TRR-33 | Stale known-issues and missing high-risk recovery caveats | GH-313 |
| TRR-34 | Stale structure/toolchain/CLI/lifecycle descriptions | GH-313 |
| TRR-35 | Stale planning and contradictory fee/trustee material not quarantined | GH-314 |
| TRR-36 | Duplicate PDF, compiled binary, weak CI/CD PDF, bounty/team wording hygiene | GH-314 |
| TRR-38 | Landing lacks canonical/social/structured discovery metadata | GH-315 |
| TRR-39 | No sitemap, robots or llms discovery files | GH-315 |
| TRR-40 | Mobile navigation disappears and minor accessibility defects exist | GH-315 |
| TRR-41 | Repeated inline styles and no small landing design system | GH-315 |

### 7.5 Informational / verified strengths

| Finding | Meaning | Action |
|---|---|---|
| TRR-22 | Big Purge re-vote semantics require an explicit design decision | Included in GH-307 |
| TRR-28 | Client/deployment strengths verified | Preserve |
| TRR-37 | Headline status/count discipline and PNYX naming verified | Preserve |
| TRR-42 | Landing has zero scripts, AA contrast and CI-grep status discipline | Preserve in GH-315 |

---

## 8. Remediation issue map and acceptance boundaries

| Issue | Scope | Required result | Priority/dependency |
|---|---|---|---|
| [#304](https://github.com/NeaBouli/TrueRepublic/issues/304) | ElectAdmin corruption and legacy-state detector/recovery design | Real bech32 regression, canonical parse/store, no partial mutation, admin/treasury/export/import proof, read-only detector | **P0, first** |
| [#305](https://github.com/NeaBouli/TrueRepublic/issues/305) | rustls advisory | Minimal patched lock, no vulnerable duplicate, complete Rust/security gates | Before PR #302 merge |
| [#306](https://github.com/NeaBouli/TrueRepublic/issues/306) | Validator exit and stone rewards | Slash-safe exit semantics, bounded non-farmable rewards, conservation/recovery tests | Critical/High closure |
| [#307](https://github.com/NeaBouli/TrueRepublic/issues/307) | One-member authority/revocation | Versioned key rule, one vote/reward, exclusion/purge/rotation semantics | Critical/High closure; feeds future ballot invariants |
| [#308](https://github.com/NeaBouli/TrueRepublic/issues/308) | Unsafe legacy/example contracts | Prove deployment status, quarantine/remove/rewrite, adversarial tests, cw2/migration if retained | Critical/High closure |
| [#309](https://github.com/NeaBouli/TrueRepublic/issues/309) | ZKP identity custody/mock registration | No plaintext persistent secret, cleanup/migration, canonical commitments, preview registration blocked | Before GH-300 claim or ZKP activation |
| [#310](https://github.com/NeaBouli/TrueRepublic/issues/310) | Governance/DEX authority and bounded EndBlock | Reachable authorities, bounded growth/work, scale and migration tests | Before genesis freeze |
| [#311](https://github.com/NeaBouli/TrueRepublic/issues/311) | Latent chain correctness | Remove unsafe async paths, bind signatures, fail-closed DEX/rating/admin paths | Before release freeze |
| [#312](https://github.com/NeaBouli/TrueRepublic/issues/312) | Client CSP/RPC/unlock/assets | CSP including WASM/Worker, explicit trust boundary, throttling, local assets | Before production client |
| [#313](https://github.com/NeaBouli/TrueRepublic/issues/313) | Operator docs/fictitious infrastructure | Executable commands, real contacts/references, synchronized wiki/docs | Before operator/testnet handoff |
| [#314](https://github.com/NeaBouli/TrueRepublic/issues/314) | Whitepaper/history/claims | Honest vision labels, correct PoD/economics/counts, historical quarantine | Before public rollout communications |
| [#315](https://github.com/NeaBouli/TrueRepublic/issues/315) | Landing discovery/mobile/a11y | Metadata/discovery, accessible mobile nav, preserved no-tracking strengths | Before polished public launch |

No issue authorizes deployment, live migration, real funds, contract
instantiation, production identity use or deletion of historical assets.

---

## 9. Open pull requests and CI

### PR #301 — client dependency group

Updates five development dependencies:

- autoprefixer 10.5.4 → 10.5.6;
- @types/node 22.20.1 → 22.20.2;
- happy-dom 20.14.0 → 20.14.3;
- typescript-eslint 8.69.0 → 8.70.0;
- Vite 8.2.2 → 8.3.0.

All substantive client/browser/security/reproducibility checks are green. Docs
Consistency is red for one exact reason: `docs/status.json` still truthfully
declares main's Vite 8.2.2 while the PR lockfile declares 8.3.0. This is not a
GitHub quota failure. The PR should be rebuilt/reconciled from the then-current
main, documentation synchronized, full exact-head checks rerun, reviewed and
merged only if all checks are green.

### PR #302 — external audit reports

- Report-only: six documentation files, 524 added lines.
- All five report SHA-256 pins were independently recomputed and match.
- Docs, Go vulnerability/static, secret, Node and retirement checks pass.
- `rust-audit` fails on
  [RUSTSEC-2026-0285](https://rustsec.org/advisories/RUSTSEC-2026-0285.html)
  because main pins rustls 0.23.38; patched releases begin at 0.23.45.
- Keep unmerged until GH-305 fixes the base dependency without an ignore.
- Then rebase/rerun exact head and merge the report through normal review.

### CI and rate limits

- GitHub API core quota was 5,000/5,000 remaining at this checkpoint.
- New Actions runs start normally.
- Standard hosted runners for this public repository are not blocked by a
  private-repository minutes allowance; see
  [GitHub Actions billing](https://docs.github.com/en/billing/concepts/product-billing/github-actions).
- The visible failures are real repository checks, not a run-rate limit.

Main's latest recorded scheduled security run is green only because it ran
before the new rustls advisory entered the database. The next fresh scan is
expected to expose the advisory until GH-305 lands.

---

## 10. GH-300 unfinished maintained-browser prover work

### Intended scope

- strict versioned same-origin artifact manifest;
- exact size and SHA-256 checks;
- toxic-waste/test-only classification;
- lazy browser WASM/artifact loading;
- dedicated serialized Worker with cancellation/stale-generation handling;
- canonical BN254/MiMC helpers;
- fresh real-browser synthetic proof accepted by native verifier/keeper replay;
- `isSubmittable=false` and no signing/RPC/broadcast path.

### Current status

- Local implementation files exist on `feature/GH-300-browser-zkp-prover`.
- They are modified/untracked and have not passed a final integration chain.
- They are not committed, pushed, reviewed or merged.
- They do not earn rollout credit.
- Work is paused to avoid building on an unsafe identity-custody/mock-
  commitment foundation identified by GH-309.

### Resume rule

GH-300 may resume as compatibility-only work after the GH-304 first-response
block and with GH-309 boundaries explicitly integrated. Sol and Kimi must own
non-overlapping slices. Final acceptance requires complete client/browser/Go
ZKP/security/docs gates and protected exact-head review.

---

## 11. Deferred product programs beyond Basic-59

### 11.1 Optional domain ballot engine — GH-232

Architecture is complete; implementation is deferred. Planned stages:

1. B0: approve ADR, legal/process matrix, enums, exact integer math and
   migration plan.
2. B1: pure deterministic tally library with table/property/fuzz tests.
3. B2: opt-in public yes/no/abstain lifecycle with immutable ballot/electorate
   snapshots, genesis and export/import.
4. B3: person elections, explicit no-winner/tie/runoff semantics and legacy
   election retirement.
5. B4: systemic-consensing and hybrid consensing-to-ratification.
6. B5: ballot-scoped private proofs/nullifiers, ceremony, client support and
   privacy-safe reward decision.
7. B6: independent consensus/crypto/privacy/legal review and staged testnets.

Blocking gates: stable rollout, approved consensus migration, production ZKP
for secret profiles, and jurisdiction-specific review for legally binding use.

### 11.2 Sovereign Alpha application

The target is an installable Telegram-like TrueRepublic experience without a
mandatory hosted website or custodial signer. The Beta remains supported until
Alpha qualifies.

Planned slices:

- A0: go-waku/mobile/Flutter/Go/FFI qualification;
- A1: Go codecs, identities, key vault, encrypted local store and transport
  conformance;
- A2: Android-first chat/discussion with qualified MLS group encryption;
- A3: wallet and governance integration;
- A4: signed reproducible distribution and updates;
- A5: hardening, iOS/desktop/hardware signer and independent review;
- A6: full readiness, portability, coexistence soak, rollback and go/no-go.

The architecture lists 20 future issue-sized blocks. They are not yet opened as
an executable GitHub issue chain. Major decisions remain open: exact MLS
implementation, Waku qualification/fallback, desktop order, mobile light
verification, hardware signer, iOS distribution, push privacy, history
recovery, attachment storage, local DB encryption and spam control.

Alpha additionally requires all DG-2 through DG-8 work: exact dependency
compatibility/provenance and SBOM review; MLS interoperability, mobile
performance and independent cryptographic qualification; encrypted local-store
and topic-metadata design; moderation plus a verifiable moderation log; a
signed chain-authority-verified domain-policy record; attachment-backend
selection; anti-spam protocol qualification; and an Alpha-specific threat-model
extension.

### 11.3 Sovereign V4 edge program

- The V4-0 protocol-library slice is merged and protected-CI-qualified, but
  deliberately unwired; no V4 runtime exists.
- V4-1 discussion runtime is open: Waku adapter, encrypted store,
  offline/gap/replay and two-device recovery tests.
- V4-2 civic lifecycle is open: bills, amendments, consensing and local metrics.
- V4-3 device profiles are open: mobile verifier and pruned citizen/home nodes.
- V4-4 domain apps are open: signed manifests, sandbox/capability broker.
- V4-5 hardening is open: recovery, packaging, security/privacy review and
  explicit go/no-go.

V4-2 formal ballots depend on GH-232. V4-5 anonymous-production claims depend
on the Phase-2 ZKP exit. No second chain, token or foreign settlement authority
is selected; TRChain remains the sole authority.

Before any V4 runtime claim, DG-V4-1 through DG-V4-6 require exact component
compatibility/provenance/SBOM review, Waku and real-device qualification, a
proved mobile light-verification path, sandbox selection with key-isolation and
capability evidence, civic/privacy/legal review of bill and metric schemas, and
approval of measured device budgets.

### 11.4 Optional future web interface

Whether a website remains after Alpha is intentionally undecided. Inputs to
that decision include browser-wallet XSS/custody limits, maintenance cost of
two front ends and the value of a read-only public-chain viewer.

---

## 12. Work that requires more than repository code

The following cannot be completed by ordinary local coding alone:

- production trusted-setup ceremony and independent transcript verification;
- independent cryptographic/privacy/security audits;
- domain/DNS/security-contact decision if `truerepublic.network` is retained;
- real private seed/sentry/validator/RPC infrastructure;
- TLS/firewall/paging/monitoring ownership;
- multi-day capacity soak and live incident exercises;
- release signing identities and protected signing environment;
- real tagged build, publication channels and provenance;
- private and public testnets/canary;
- rollout genesis/validator/allocation governance;
- accountable release freeze and go/no-go;
- jurisdiction-specific review for binding ballots or political-personal data;
- future Alpha third-party dependency, mobile distribution and license review.

Each requires an exact bounded authorization at execution time. General
development permission is not production or infrastructure permission.

---

## 13. Recommended non-overlapping execution sequence

### Block 0 — Preserve and control state

- Keep GH-300 work isolated.
- Keep this report and issue/Bridge state synchronized.
- Do not merge red PRs.

### Block 1 — Immediate P0 recovery

1. GH-304 reproduce and fix ElectAdmin corruption.
2. Add read-only corrupted-state detection.
3. Specify legacy repair separately; do not execute migration.
4. Full Go, race/coverage, export/import, recovery and independent review.

### Block 2 — Restore security gate and publish audit

1. GH-305 minimally update rustls.
2. Run full Rust and repository security gates.
3. Rebase/rerun PR #302.
4. Merge audit reports only when exact head is green.

### Block 3 — Close Critical/High governance/contracts findings

- GH-306 validator exits/reward farming.
- GH-307 one-member authority/revocation.
- GH-308 contract deployment inventory and fail-closed quarantine.
- Fresh independent retest of every Critical/High closure.

### Block 4 — Harden protocol and client

- GH-309 identity custody/mock registration.
- GH-310 governance/DEX authority and bounded EndBlock.
- GH-311 latent chain correctness.
- GH-312 CSP/RPC/unlock/assets.

### Block 5 — Documentation and landing truth

- GH-313 operator/wikis/infrastructure.
- GH-314 whitepaper/history/claims.
- GH-315 landing metadata/mobile/a11y/design.
- Synchronize repository wiki and live GitHub Wiki only after checks pass.

### Block 6 — Dependency reconciliation

- Rebase/review PR #301 from the post-remediation main.
- Synchronize Vite metadata and rerun full client/repository matrix.

### Block 7 — Finish production ZKP

- Resume GH-300 compatibility integration with GH-309 boundaries.
- Produce reproducible production artifacts.
- Conduct governed ceremony and establish rotation.
- Complete browser-to-chain real proof path.
- Independent cryptographic/privacy/setup review.

### Block 8 — Close remaining Basic-59 operations/release gates

- Real private topology and operator evidence.
- Tagged reproducible builds.
- Signed publication/SBOM/provenance.
- Freeze and verify rollout genesis.
- Private testnet and failure drills.
- Public canary.
- Release freeze, final audit and accountable go/no-go.
- Close all eight final exit criteria.

### Block 9 — Post-Basic programs

Only after explicit prioritization:

- GH-232 ballot stages;
- Sovereign Alpha A0-A6;
- V4-1 through V4-5;
- optional web-interface decision.

---

## 14. Verification required for every future block

At minimum:

- focused regression first;
- complete relevant Go/Rust/client tests;
- build, lint, typecheck, format and vet;
- race/coverage where applicable;
- repository security, vulnerability, secret, license and supply-chain gates;
- docs consistency and link checks;
- export/import/restart for state changes;
- multi-validator agreement for consensus work;
- browser matrix and bundle budgets for client work;
- deterministic/reproducible checks for release work;
- Kimi bounded implementation or independent deep review for large work;
- Sol line-by-line diff/security/integration review;
- exact-head protected CI, zero unresolved review threads;
- exact-main/public readback after merge;
- append-only Bridge and GitHub ticket closeout.

“Implemented locally” is not Done. “Green CI” is not production approval.

---

## 15. Known uncertainties and decisions still needed

1. Exact safe semantics and migration boundary for validator exits/rewards.
2. Exact one-member anonymous authority and revocation design.
3. Fate of each unsafe legacy/example contract.
4. Legacy-state repair policy for already-corrupted domain admins.
5. Production ZKP ceremony governance and artifact rotation.
6. Private ballot cryptography, coercion/stimmenkauf trade-offs and reward
   privacy.
7. Real governance-domain and DEX registry authority.
8. Production chain ID/genesis/validator/allocation set.
9. Real infrastructure ownership without contradicting community governance.
10. Alpha Waku/MLS/mobile verification/distribution choices.
11. V4 sandbox/device budget/transport decisions.
12. Whether a web interface survives after Alpha.

Ballot-specific approvals still include quorum/threshold defaults and who may
set them; Stage-B electorate reuse versus refresh; missing-rating semantics;
public-ballot eligibility warnings; sealed-secret recovery; unlinkable rewards;
cancellation authority; and migration/retirement of historical `elecvote:`
keys.

Until decisions are approved and verified, fail-closed behavior and honest
non-production wording are mandatory.

---

## 16. What “Basic rollout complete” will require

The project may claim Basic-59 completion only when:

- GH-29 is 59/59;
- audit Critical/High findings have independent verified closure;
- production ZKP and client submission are real, audited and fail-closed;
- the exact frozen genesis and release artifacts are reproducible and signed;
- private and public staged operation has succeeded with recovery drills;
- monitoring and accountable operators are active;
- every protected exact-head and exact-main gate is green;
- an explicit community-governance go/no-go exists; and
- public README, landing, wiki, release notes and machine-readable status all
  match that evidence.

Until then, the only accurate statement is:

> TrueRepublic v0.4 is a recovery-stage, community-governed open-source
> engineering foundation. It is not production-ready, mainnet-ready or
> approved for real funds or keys.

---

## 17. Snapshot sources

- [Rollout tracker #29](https://github.com/NeaBouli/TrueRepublic/issues/29)
- [Audit report PR #302](https://github.com/NeaBouli/TrueRepublic/pull/302)
- [Audit register #303](https://github.com/NeaBouli/TrueRepublic/issues/303)
- [Status checkpoint issue #316](https://github.com/NeaBouli/TrueRepublic/issues/316)
- `docs/status.json`
- `docs/ROLLOUT_ROADMAP.md`
- `docs/LIMITATIONS.md`
- `docs/SOVEREIGN_ALPHA_ARCHITECTURE.md`
- `docs/SOVEREIGN_V4_EDGE_ARCHITECTURE.md`
- `docs/GOVERNANCE_BALLOT_ARCHITECTURE.md`
- `docs/security/THREAT_MODEL.md`
- `docs/agent-bridge/PROJECT_STATE.md`
- `docs/agent-bridge/TODO.md`
- root `BRIDGE.md`

This is a point-in-time checkpoint. Future status changes must update GitHub
tickets and append to the Bridge; this historical report should not be silently
rewritten to imply work existed at the checkpoint when it did not.
