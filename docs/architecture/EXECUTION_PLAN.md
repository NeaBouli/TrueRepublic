# TrueRepublic — Remaining-Work Execution Plan (GH-327B)

**Status:** planning document — documentation only. It earns no rollout
checkbox, changes no code, and authorizes no deployment, migration, release,
audit, or production action.
**Base:** verified against `origin/main` `1283a4452d36784ea962f7fc8ba13f9ad45472cd`
(2026-09-19) plus read-only GitHub state read on 2026-09-27.
**Author:** Fleet Worker B (Kimi), brief `.fleet/tasks/GH327B.md`, parent
issue [#327](https://github.com/NeaBouli/TrueRepublic/issues/327).
**Companion:** GH-327A publishes the code-grounded architecture map
(`docs/architecture/MAP.md`); that map is the mandatory scope boundary for
every implementation block ordered here.

---

## 1. Verified current state (2026-09-27)

### 1.1 Verified facts

| Fact | Value | Source |
|---|---|---|
| Current `main` | `1283a44` — PR #302, audit series publication (2026-09-19 23:40 EEST) | git log |
| Security stack | merged via PR #320 (`b26b2b9`): gRPC v1.83.2 (GO-2026-6348/6443), rustls 0.23.45 (RUSTSEC-2026-0285), npm audit-gate hardening (GH-321) | git log, `contracts/Cargo.lock` |
| TRR-01 fix | merged via PR #319 (`2376daf`), GH-304 closed 2026-09-19 | git log, GitHub |
| Rollout | **36/59** = 36/51 phase work + 0/8 exit criteria | #29 body, `docs/status.json`, `docs/ROLLOUT_ROADMAP.md` |
| Phase split | P1 7/7 · P2 2/7 · P3 6/6 · P4 7/8 · P5 5/6 · P6 6/7 · P7 3/10 (sums to 36/51) | #29 body |
| Verified tests | 2,486 = 2,132 Go + 26 Rust + 328 maintained client | `docs/status.json`, README badge |
| `production_ready` | **false** everywhere | `docs/status.json` |
| Audit register #303 | TRR-01…TRR-42: 3C/4H/15M/16L/4I; **1 fixed** (TRR-01), **38 open** ticketed #306–#315, 3 informational strengths (TRR-28/37/42) need no fix | #303, `docs/community-audits/` |
| Open issues (18) | #29, #232, #300, #303, #306–#316, #324, #325, #327 | `gh issue list` |
| Open PRs (2) | #317 (docs status checkpoint, CONFLICTING), #326 (Dependabot client-maintenance, MERGEABLE, 1 red check) | `gh pr list` |
| Preserved candidates | dirty GH-306 worktree (current implementation block); dirty paused GH-300 checkout (resume gated on #309) | `.fleet/PLAN.md`, `TODO.md` |
| Hosted CI quota | reset expected **2026-10-01**; no avoidable hosted runs before | `.fleet/PLAN.md`, #327 |
| Go no-fix exceptions | GO-2026-5932, GO-2026-4740, GO-2023-1881, GO-2023-1821 — window expires **2026-10-13** | `PROJECT_STATE.md` |

### 1.2 Stale or conflicting coordination entries (reconciled, not silently resolved)

1. **Bridge files are frozen at 2026-09-19 ~23:40.** `PROJECT_STATE.md`,
   `TODO.md`, `ACTION_LOG.md` record GH-304 and GH-318 as "merge pending" and
   name `b26b2b9` as latest main. Both PRs merged the same day; main is two
   commits ahead (`2376daf`, `1283a44`).
2. **TODO.md internal contradiction:** "[x] Land GH-318/GH-305/GH-321" is true
   in effect — all three landed cumulatively via PR #320 — but the GH-305 and
   GH-321 sections still carry open "protected CI/merge" boxes, and their
   recorded PR heads (#322 `f854bdc`, #323 `4e86e8c`) exist only on unmerged
   remote branches. The content is on main; the recorded merge path is stale.
3. **#29's latest comment (2026-09-19)** names #304/#305 as blockers; both
   closed that day. Current critical/high blockers are **#306, #307, #308**.
4. **SECURITY_NOTES.md predates the entire 2026-09-15 TRR audit series** — no
   mention of TRR-01…TRR-42, #303–#316, GO-2026-6348/6443, or
   RUSTSEC-2026-0285.
5. **No bridge entry exists** for the audit-series publication (PR #302),
   GH-306–GH-316, GH-324/GH-325, PR #317, or PR #326.
6. **PR #317's body** claims "18 open issues" and merge gates on GH-318 and
   PR #302 — both resolved; the count still happens to be 18 (#327 added,
   #304/#305/#318 closed) with different composition.
7. Per-phase rollout counts at 36/59 are not restated in the bridge (last
   full split is recorded at 35/59); #29's body is the canonical split.
8. Codex should reconcile 1–6 in the append-only coordination files when
   integrating GH-327; this plan is the verified baseline for that update.

---

## 2. Work classification

### 2.1 Must finish — Basic rollout (the 59-item tracker, #29)

Open phase items (15) mapped to their execution vehicle:

| Phase item | Vehicle |
|---|---|
| P2 ×5 (real prover, reproducible PK/VK, ceremony provenance, browser-to-chain real proofs, independent crypto/privacy review) | #300 resume + new Phase-2 tickets (§5 M5) |
| P4 ×1 (connect audited real ZKP path, remove preview dead paths) | after Phase-2 exit gate |
| P5 ×1 (independent security review + resolve every critical/high) | #306/#307/#308 + verification evidence (§8) |
| P6 ×1 (real production topology + abuse protection) | private deployment evidence (M6) |
| P7 ×7 (tagged reproducible builds; signed artifacts/SBOM/provenance; genesis freeze; genesis re-qualification; private testnet + drills; public testnet/canary; final review + go/no-go) | staged release program (M7) |

Plus the 8 final exit criteria (§10).

### 2.2 Audit remediation (register #303, all prerequisite to P5 closure)

- **In progress:** #306 (TRR-02 High, TRR-04 Medium) — preserved candidate.
- **Live-code High:** #307 (TRR-03 High, TRR-05 M, TRR-22 Info-semantics).
- **Conditional Critical:** #308 (TRR-13…TRR-18; 2C/2H/2M, all conditional on
  CosmWasm instantiation — none observed; quarantine-first).
- **Client custody:** #309 (TRR-19/20 M) — gates GH-300 resume.
- **Protocol hardening:** #310 (TRR-06/07/08 M), #311 (TRR-09…12 L + TRR-21 L).
- **Client hardening:** #312 (TRR-23 M, TRR-26/27 L).
- **Documentation:** #313 (TRR-24/25/33/34), #314 (TRR-29…36), #315 (TRR-38…41).
- **Checkpoint:** #316 (via PR #317). **Register closure:** #303 per finding
  with external verification evidence (GH-294 `verified_closed` rule).
- Dependency-security children #304/#305/#318 are **done** (§1.1).

### 2.3 Deferred — Sovereign Alpha ("V3", not counted in 59)

GH-215 Alpha track (slices A0–A6). The docs never use the label "V3"; the
intermediate generation is the Sovereign Alpha. No native Alpha implementation
exists; architecture is decision-ready documentation. Starts only after Basic
rollout stabilization per the 2026-08-23 decision record.

### 2.4 Deferred — Sovereign V4 (not counted in 59)

V4-0 (`sovereignv4/protocol`, GH-241) is implemented but deliberately unwired.
V4-1…V4-5 and decision gates DG-V4-1…DG-V4-6 remain proposed. V4-1 runtime
wiring is explicitly blocked until rollout stabilization or a new documented
decision (AGENTS.md, DECISIONS.md 2026-08-23).

### 2.5 Optional / non-goal

- **#232** domain ballot engine: deferred behind rollout stability, production
  ZKP, and jurisdiction review; authorizes no implementation.
- **truerepublic.network DNS registration** (#313): not authorized; remove or
  quarantine references instead.
- **File/binary deletions** (#314: "Kopie 2.pdf", committed Mach-O binary):
  require separate explicit approval; the ticket fixes claims only.
- **Legacy-state repair** of TRR-01-quarantined domains: detector/quarantine
  only; any mutating repair is a separate, explicitly approved task.
- V4-1+, GH-232, Alpha: out of the Basic completion definition (§10).

---

## 3. Ticket mapping

Worker types: **Sol** (integration/review/GitHub owner), **K** (Kimi — large
bounded blocks), **C** (Claude — bounded blocks), **G** (Grok — small bounded
only). CI tiers: **Full** = make verify + security/static/secret gates +
multi-validator recovery gate; **Mod** = full minus multi-validator (single
lane); **Docs** = consistency/license/diff contexts only.

| Ticket | Architecture node | Findings / scope | Risk | Prerequisites | Acceptance evidence | Worker | CI |
|---|---|---|---|---|---|---|---|
| #324+#325 | test harness (`server_lifecycle_test.go`, multi-validator/governed-upgrade harnesses) | freeTCPPort TOCTOU flakes (runs 35462066576, 35466470548) | Low (test-only), blocks CI reliability of every later gate | none | one fix closes both: reserved port allocation, fail-fast child-exit logs, ≥3 repeated harness runs green, full multi-validator gate | G or C | Mod + repeated harness runs |
| #306 | `x/truedemocracy` validator exits (`validator.go`), stone rewards (`stones.go`), treasury invariants | TRR-02 (H), TRR-04 (M) | High — consensus, token accounting, treasury | #327 map; preserved candidate reconciled with main | frozen exit semantics, slash-safe zero-payout full exit, bounded non-repeatable rewards, escrow/power/export-import invariants, independent review, Full gates | Sol integrates candidate + K support | Full |
| #307 | `x/truedemocracy` anonymity/membership (`anonymity.go`, `msg_server.go`, `governance.go` exclusion) | TRR-03 (H), TRR-05 (M), TRR-22 (Info semantics) | High — governance authority, payout multiplication | #306 merged (same package) | versioned voting-key rule, join-bypass closed, multi-vote/multi-payout impossible, privacy-preserving revocation, Big Purge semantics recorded, replay/rotation/purge/export tests | K | Full |
| #308 | `contracts/` (core treasury/governance, examples zkp-aggregator/governance-dao/dex-bot) | TRR-13…18 (2C/2H/2M, conditional) | High if ever instantiated; currently dormant | none beyond map; parallel to Go lane | non-instantiation proven, fail-closed non-production packaging boundary, per-contract remove/archive/rewrite decision, adversarial tests, cargo audit clean, independent review | C | Rust lane (fmt/clippy/test/audit) |
| #309 | `client-web/src/stores/identityStore.ts`, `services/zkp.ts`, onboarding | TRR-19/20 (M) | Medium-high — secret custody, on-chain mock registration | map | no plaintext secret at rest + migration, validated import/warned export, canonical BN254/MiMC commitment + chain-bound nullifier, MsgRegisterIdentity blocked from preview, `isSubmittable` stays hard false | K (after #306) or C | Client lane (lint/vitest/build budgets/audit/browser matrix) |
| #310 | `x/truedemocracy` genesis anchoring, `x/dex` registry authority, EndBlock bounds | TRR-06/07/08 (M) | Medium-high — consensus performance, dead authority | #307 merged | reserved governance domain anchored on every genesis path, DEX authority spendable or registry frozen, bounded state growth, scheduled/bounded EndBlock scans, scale/gas/upgrade tests | K | Full |
| #311 | `x/truedemocracy` keeper/tree/crypto, `x/dex` LP paths | TRR-09…12 (L), TRR-21 (L) | Medium — correctness cleanup, touches consensus paths | #310 merged | async mutation + hardcoded nodes removed/quarantined, CoinBurnRequired fixed/retired, chain-bound onboarding signatures, exclusion authority validity, zero-output LP removal rejected, fail-closed GetAllLPPositions/legacy RateProposal | K | Full |
| #312 | `client-web` nginx/deploy, chains config, unlock flows, landing assets | TRR-23 (M), TRR-26/27 (L) | Medium — client security posture | #309 merged (same lane) | browser-tested strict CSP, documented single-RPC `prove:false` boundary, unlock throttling + password lifetime, local assets only, browser security/a11y/budget gates | C or K | Client lane |
| #313 | operator docs + wiki operations pages | TRR-24/25/33/34 (M/L) | Low-medium — operator harm from wrong guidance | none | env/genesis/Docker/ports corrected, unmounted staking/distribution commands removed, CLI order + PNYX units (×10⁶) fixed, truerepublic.network references removed (no DNS purchase), SECURITY.md-only reporting, link check | G or C | Docs |
| #314 | whitepapers EN/DE, stale roadmaps, content surfaces | TRR-29…36 (M/L) | Low — misleading public claims | none | vision/recovery framing, PoD/anti-whale aligned with code, no financial-performance language, counts/toolchain reconciled, EN/DE fee/trustee text aligned, consistency checks extended; **no deletions without separate approval** | G or C | Docs |
| #315 | `docs/index.html` (landing) | TRR-38…41 (L) | Low | none | canonical/OG/Twitter/JSON-LD metadata preserving production=false, sitemap/robots/llms.txt, keyboard-accessible mobile nav ≤768px, design-system consolidation, AA contrast + zero-script preserved, local/live byte identity | G or C | Docs (+ Pages readback post-merge) |
| #316 / PR #317 | `docs/status/` checkpoint | durable status reconciliation | Low | §6 treatment; after M1–M2 merges | refreshed checkpoint citing current main, GH-327 map/plan, register state; protected exact-head matrix, zero threads; closes #316 | Sol | Docs |
| #300 | `client-web` ZKP prover (WASM/worker), `internal/zkpprover` | Phase-2 prover integration (paused) | High — cryptographic surface; compatibility-only | **#309 merged**; preserved checkout | per its 9 acceptance boxes; TEST-ONLY toxic-waste manifest, lazy same-origin loading, Worker prover, canonical encodings, real-browser fail-closed tests, `isSubmittable` false; no GH-29 checkbox alone | K | Client lane + ZKP/WASM contexts |
| #303 | audit register | register hygiene | Low | remediation merges | per-finding checkbox + remediation/verification links; TRR-28/37/42 marked verified-strength/no-action; closure per GH-294 external-verification rule | Sol | Docs |
| #29 | rollout tracker | tracker hygiene | Low | per-item evidence | item credit only after protected merge + public readback; exit criteria only with linked evidence | Sol | Docs |

---

## 4. Dependency graph and stop conditions

### 4.1 Dependency graph (textual; arrows = "must land before")

```
#327 (map+plan) ──► every implementation block (scope boundary)
#324/#325 ──► all later hosted multi-validator gates (flake protection)
#306 ──► #307 ──► #310 ──► #311            (Lane G: same Go package, sequential)
#308 (Lane R, parallel) ──┐
#306 + #307 ──────────────┼──► Phase-5 crit/high closure ──► §8 verification run
                          ┘         ──► #303 register closure (external evidence)
#309 ──► #300 resume ──► Phase-2 real-proof items ──► P2 exit gate
        ──► #312 (same client lane)
P2 exit gate ──► P4 "connect audited ZKP path" ──► P4 exit gate
Lane G merged ──► genesis freeze ──► genesis re-qualification (GH-244 contract)
tagged reproducible builds + genesis freeze ──► private testnet + drills
private testnet ──► public testnet/canary ──► release freeze ──► go/no-go
#313/#314/#315 (docs lane, parallel, no blockers)
#316/PR #317 after M1–M2 (content freshness)
#326 superseded by bounded reconciliation PR (post-reset, client lane)
```

### 4.2 Explicit stop conditions

1. **New or regressed critical/high finding** → halt rollout progression;
   open a remediation ticket before any further Phase 6/7 work.
2. **GH-306 candidate cannot be reconciled with main** → stop and escalate to
   Sol; no silent rewrite of the preserved candidate.
3. **Scope drift into gated territory** — mutating legacy-state repair, file
   deletions (#314), DNS registration (#313), V4-1 wiring, GH-232
   implementation, any mainnet action → stop; each needs a separate explicit
   decision.
4. **Consensus-breaking change without governed-upgrade evidence** → blocked
   by definition (AGENTS.md).
5. **Hosted CI quota exhausted** → local-only mode; merges wait for reset.
6. **Proven worker outage** → fleet fallback protocol only (probe-proven);
   security reviews are never silently replaced.
7. **Coordination contradiction** (stale bridge vs verified state) → reconcile
   append-only files first; this plan is the reconciliation baseline.
8. **Worker report `partial`/`failed` or filled `risks`/`security`** → no
   merge; Sol triage before the next block in that lane.

---

## 5. Milestones: GH-306 completion → production go/no-go

**M0 — now → 2026-10-01 (local-only, no hosted CI spend)**
GH-327A map + GH-327B plan delivered; Codex integrates and reconciles
coordination files (§1.2). GH-306 candidate completed and locally gated from
its preserved worktree (`make verify`, focused race/coverage, local
multi-validator where resources allow). PR #326 left untouched. No pushes
that trigger hosted workflows.

**M1 — post-reset, CI foundation**
One PR fixes #324+#325 (flake protection). GH-306 protected PR on exact main
→ full matrix → merge. Register #303: TRR-02/TRR-04 remediation links.

**M2 — parallel remediation wave 1**
Lane G: #307. Lane R: #308 (quarantine boundary first, then per-contract
decisions). Lane C: #309. Merges sequenced one at a time onto main
(G → R → C), each rebased to exact head with its tier matrix.

**M3 — remediation wave 2 + docs sweep**
Lane G: #310, #311. Lane C: #312. Lane D: #313, #314, #315 (batched, cheap
contexts). PR #317 rebased/refreshed → merged (closes #316). Client-maintenance
reconciliation PR supersedes #326 (§6.2). Renew the four Go no-fix exceptions
before **2026-10-13** if upstream still ships no fix.

**M4 — Phase 5 closure → 37/59**
All critical/high register items (#306/#307/#308) merged → Phase-5
critical/high resolution complete. §8 verification run + external evidence per
GH-294; #303 register reconciled per finding. Phase 5: 6/6.

**M5 — Phase 2 + Phase 4 completion → 43/59**
#300 resumed (compatibility-only). Production prover integration, reproducible
PK/VK artifacts, ceremony provenance + rotation policy, browser-to-chain real
-proof compatibility, independent cryptographic/privacy/setup review →
P2 7/7 (42/59). Connect the audited real ZKP path, remove preview dead paths →
P4 8/8 (43/59). Anonymous submission stays fail-closed until the P2 exit gate.

**M6 — Phase 6 topology → 44/59**
Real private seed/sentry/validator/RPC deployment with abuse protection and
live operator rehearsal evidence → P6 7/7.

**M7 — Phase 7 staged release → 50/59 overall; 50/51 phase work**
Reproducible tagged binaries/images; signed artifacts + SBOM + provenance;
genesis/chain-ID/parameters/authorities/allocations freeze + independent
review; genesis re-qualification via the GH-244 contract against the exact
artifact; private multi-validator testnet + failure drills; public testnet or
controlled canary with monitoring and rollback window → P7 9/10 (50/59).

**M8 — Final authorization → 59/59**
Release-candidate freeze while final evidence is reviewed; final release
review; explicit go/no-go with accountable approvers recorded → P7 10/10 and
the 8 exit criteria evidenced. Only then does the staged public rollout begin.

---

## 6. Treatment of open PRs

### 6.1 PR #317 (docs status checkpoint, closes #316)

**Do not merge as-is.** The head CONFLICTS with main and its red checks
(go-vuln, rust-audit) are the pre-#320 advisories on a stale base. Its body
documents a 2026-09-19 state that is already superseded (§1.2).

Exact treatment:
1. Keep open, unmerged, no hosted reruns before the reset.
2. After M1–M2 merges, rebase `docs/GH-316-comprehensive-status` onto
   then-current main and refresh content: current main hash, GH-327 map/plan
   references, audit-register state, the #326 decision, corrected blocker list.
3. Require the protected exact-head matrix and zero unresolved review threads.
4. Merge as the durable checkpoint (closes #316). No rollout credit.

### 6.2 PR #326 (Dependabot client-maintenance, 7 dev updates)

**Do not merge the bot head directly.** Project precedent (GH-235, GH-254):
bot heads are rebuilt as bounded current-main candidates. Its one red check is
consistent with the fail-closed exact-Vite binding (`vite 8.2.2 → 8.3.0`
conflicts with `docs/status.json` and `CLAUDE.md`, which pin 8.2.2).

Exact treatment:
1. Pre-reset: no rebase, no comment churn, no CI spend.
2. Post-reset: one bounded client-maintenance reconciliation PR on current
   main containing the 7 updates (autoprefixer, react-router-dom,
   @testing-library/dom, @types/node, happy-dom, typescript-eslint,
   vite 8.3.0) **plus** synchronized `docs/status.json`/`CLAUDE.md` Vite
   bindings, Apache-2.0 lockfile metadata preservation, registry/integrity
   review, re-measured bundle budgets, and the full client lane
   (`npm ci`, lint, Vitest, build, budgets, audit).
3. Close #326 as superseded with the replacement-merge evidence link.
4. Keep it a separate PR from #309/#312 — dependency reconciliation is never
   mixed into security-remediation scope.

---

## 7. CI strategy around the 2026-10-01 quota reset

**Before 2026-10-01 (local-only):**
- No pushes or PR updates that trigger hosted workflows; GH-327 publication
  timing is Codex's decision around the reset.
- All gates runnable locally: `make verify`, focused race/coverage,
  `scripts/check-consistency.sh`, `git diff --check`, client/Rust lanes.
- GH-306 candidate completion and review preparation only.

**After 2026-10-01 (priority-ordered spend):**
1. #324/#325 fix first — it protects every later multi-validator run.
2. GH-306 full protected matrix (largest single spend, furthest along).
3. Lane merges sequenced one at a time; each PR rebased to exact head before
   its matrix; no two implementation PRs race for exact-main.
4. Multi-validator recovery gate runs at Lane-G merge points and M5–M8
   checkpoints — not for Rust/client/docs-only merges.
5. Docs lane batched into few PRs (docs contexts are cheap; path filters
   already skip most heavy jobs for docs-only changes).
6. Resolve review threads before re-push to avoid duplicate exact-head
   matrices; target one hosted matrix per candidate.
7. **2026-10-13:** the four Go no-fix exceptions expire; re-validate upstream
   fix availability and renew or remediate before the gate goes red.
8. Exact-main verification (Go/Rust/client/security/docs/reproducible/Pages)
   after each merge; public status (`docs/status.json`, README, roadmap,
   Pages, Wiki) is republished only from verified exact-main state.

---

## 8. Security-audit timing decision (Cloudflare-derived skill)

Method source: `~/.codex/skills/security-audit` — adopted unchanged from
Cloudflare's MIT-licensed `security-audit-skill` on 2026-09-26 (HERKUNFT.md).
The skill is **source-only and sandboxed** (`sandboxed-source-and-local-only`);
it categorically forbids probing live endpoints, and it contains **no**
Cloudflare-infrastructure, WAF, or rate-limit scheduling constraint —
"Cloudflare" names only the skill's upstream author.

**Decision — now: no audit.**
1. The GH-327B brief explicitly forbids starting a full audit or probing any
   endpoint.
2. 38 of 42 register findings are open; a pass now would spend its budget
   re-discovering known, already-ticketed issues.
3. The skill's budget gate requires an explicit full-audit-mode trigger and an
   agent-invocation budget funding reconnaissance plus critic and verifier
   reserves; neither exists.
4. The community TRR series (2026-09-15) is 12 days old and its baseline
   (`f5de5a1`) is only two merges behind; its register is the live prior-run
   input.

**Decision — later: one bounded verification run, then external closure.**
- **When:** after M4 prerequisites — #306, #307, #308 merged on exact main
  (plus #309 if the client surface is in scope). Earlier runs are waste;
  later runs block Phase-5 closure.
- **What:** a skill-conformant scoped run (profile `standard`, scoped to
  `x/truedemocracy`, `contracts/`, and `client-web/src`) that carries the TRR
  register as prior findings: every remediated finding becomes a revalidation
  unit (`confirmed` only if current independent validation re-establishes the
  result); TRR-13…TRR-18 stay conditional units until any contract is actually
  instantiated. The run must state that no skill-format prior coverage ledger
  exists and must not imply exhaustive coverage.
- **External closure:** GH-294 accepts `verified_closed` only with
  `verification_source=external`. The internal skill run supplements but does
  not replace external verification — register closure per finding requires
  the external auditor (Collateral Web3 re-audit) or comparably attributed
  external evidence. That external engagement is also the natural completion
  evidence for the Phase-5 independent-review item.

**Precise prerequisites for the run:**
1. #306/#307/#308 (and #309 for client scope) merged; exact-main gates green.
2. Explicit owner trigger naming mode, profile, scope, and budget (agent
   invocations with the mandated reserves: recon + critics + ≥1 verifier per
   candidate, ~30% balance reserve when in doubt).
3. Output directory outside the target; `sandboxed-source-and-local-only`
   execution; dummy fixtures/secrets only; no dependency installs; no network.
4. No live endpoint probing inside the run. Any landing/Pages probe (as the
   TRR series used for reports 3/5) is outside the skill and needs separate
   explicit owner authorization.
5. Hosted CI quota available if hosted evidence is required.
6. Attribution decided up front (internal supplement vs external closure
   evidence) so GH-294 closure criteria are met without re-runs.

---

## 9. Token-efficient Fleet allocation

### 9.1 Lanes with non-overlapping file ownership

| Lane | Tickets | File ownership (exclusive) | Worker |
|---|---|---|---|
| T (harness) | #324+#325 | `server_lifecycle_test.go`, multi-validator/governed-upgrade test harnesses | G or C |
| G (Go consensus) | #306 → #307 → #310 → #311 | `x/truedemocracy/**`, `x/dex/**`, `token/**`, app wiring | K (Sol integrates #306 candidate) |
| R (contracts) | #308 | `contracts/**` | C |
| C (client) | #309 → #312 → #300 resume; #326 reconciliation | `client-web/**` | C primary; K only after Lane G if explicitly reassigned |
| D (docs) | #313, #314, #315 | `docs/**` (operator/wiki sources), whitepapers, `docs/index.html` — disjoint per ticket | G |
| S (status) | #316/#317, #303/#29 hygiene | `docs/status/`, register/tracker bodies, `docs/status.json`, BRIDGE/agent-bridge | Sol only |

Two lanes never hold write ownership of the same file set concurrently; the
GH-327A MAP.md is the boundary reference for every block. Lane G is strictly
sequential (same package, exact-main discipline). Lanes R/C/D/T run in
parallel with Lane G where worker availability allows.

### 9.2 Review discipline (no routine double-review)

- Sol reviews every delegated diff and owns all external writes — unchanged
  project rule.
- One independent read-only cross-review **only** for high-risk merges: Lane G
  (consensus/treasury), #308 contract decisions, #309 custody, and any
  cryptographic surface (#300).
- Docs-lane and harness-lane merges: Sol review only.
- A clean `status: ok` report with empty `risks`/`security` is accepted
  without a second independent review; Sol's integration review remains
  mandatory. `partial`/`failed` or filled risk fields trigger triage, not
  routine re-work.
- No delegated agent delegates further; no worker touches secrets, production,
  or external writes.

### 9.3 Token rules

- One issue = one brief = one branch; briefs list exact paths; no exploratory
  repo reads beyond them.
- Bulk evidence collection goes through bounded subagent passes (as this
  plan's research did), not repeated orchestrator reads.
- Review threads resolved before re-push → one hosted matrix per candidate.
- Preserved candidates (GH-306, GH-300) are continued, never rewritten.
- Worker availability is never probed routinely; fallback only on proven
  outage (fleet-probe), and security reviews are never silently replaced.

---

## 10. Definition of project completion and production-readiness gates

**Basic TrueRepublic is complete when:**
1. #29 reads **59/59**: all 51 phase items plus all 8 exit criteria, each with
   linked protected-merge and public-readback evidence.
2. The final go/no-go checklist holds: all seven phase exit gates evidenced;
   CI/security green on the tagged release commit; no unresolved critical/high
   finding; the real ZKP submission path audited and fail-closed;
   DR/upgrade/rollback independently repeatable; monitoring/alerting/incident
   ownership active; artifacts reproducible, signed, with SBOM/provenance; the
   maintained client and protocol surface unambiguous.
3. An explicit, accountable go/no-go decision is recorded.

**Production readiness additionally requires the executed staged sequence**
(roadmap "Rollout sequence"): reproducible release candidate → private
multi-validator testnet with disaster-recovery drills → public testnet or
controlled canary with monitoring and a rollback window → release freeze with
independent evidence review → recorded go/no-go. Green CI alone is not rollout
approval; until every gate passes, TrueRepublic remains a recovery-stage
project with `production_ready: false`.

**Explicitly outside the completion definition:** Sovereign Alpha (V3),
V4-1…V4-5, and the GH-232 ballot engine — separate deferred programs with
their own decision gates.

---

## Appendix — sources

`AGENTS.md`, `README.md`, `BRIDGE.md`, `CLAUDE.md`, `.fleet/PLAN.md`,
`docs/agent-bridge/{README,COOPERATION_RULES,PROJECT_STATE,TODO,SECURITY_NOTES,DECISIONS,ACTION_LOG}.md`,
`docs/ROLLOUT_ROADMAP.md`, `docs/status.json`, `docs/ARCHITECTURE.md`,
`docs/SOVEREIGN_ALPHA_ARCHITECTURE.md`, `docs/SOVEREIGN_V4_EDGE_ARCHITECTURE.md`,
`docs/GOVERNANCE_BALLOT_ARCHITECTURE.md`, `docs/LIMITATIONS.md`,
`docs/community-audits/README.md` + five TRR reports (2026-09-15),
GitHub issues #29, #232, #300, #303, #306–#316, #324–#325, #327 and PRs
#317/#326 (read-only, 2026-09-27), `~/.codex/skills/security-audit/{SKILL,HERKUNFT}.md`.
