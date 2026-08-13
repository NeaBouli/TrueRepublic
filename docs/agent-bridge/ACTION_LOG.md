# Action Log

## 2026-08-10 — GH-175 IBC two-chain implementation locally verified

- Added a dedicated proof-driven two-chain gate over two real persistent
  TrueRepublic applications. It opens Tendermint clients, a connection, and an
  unordered ICS-20 channel; verifies PNYX escrow/voucher/ACK; proves idempotent
  receive/ACK/timeout replay; refunds an expired packet; and
  finishes a pending acknowledgement after database reopen and fresh header.
- The first run failed with Cosmos code 2 and exposed a real production defect:
  Tendermint light-client concrete types were missing from the transaction
  interface registry. The registration and a standard regression test now pass.
- PASS: dedicated 1.05s harness, focused IBC suite, 1,441 standard Go cases,
  documentation consistency, JSON/YAML, formatting, and diff integrity.
  Candidate rollout is 25/59; full build/Vet/Race/Coverage, security gates,
  independent review, protected CI, merge, final-main, and Pages remain.
- No external relayer, public network, real account/key/fund, deployment,
  upgrade, production, or infrastructure action occurred.

## 2026-08-10 — GH-172 merged and final-main verified

- PR #173 exact head `671ff37` passed 18/18 contexts with zero open review
  threads and merged as `293ea35`; GH-172 closed and its remote branch is gone.
- Final-main Go `31340763723`, Security `31340763717`, reproducible Linux
  `31340763708`, Docs `31340763710`, and Pages `31340763413` pass. Live Pages
  publishes 1,607 cases, GH-172 evidence, 23/59, and production false.
- GH-29 now links GH-172/PR #173 inside the existing quality-depth item, so the
  canonical rollout denominator and completed count remain unchanged. GH-172
  is Done; no production or real-key/funds action occurred.

## 2026-08-10 — GH-172 published as PR #173

- Reviewed commit `972e324` is pushed and PR #173 targets protected `main`.
  The PR closes GH-172 on merge and contains only the bounded test, CI,
  evidence, audit, and status scope.
- Protected exact-head workflows and review threads remain authoritative; no
  technical gate is bypassed. Rollout and production status are unchanged.

## 2026-08-10 — GH-172 independent review remediated

- Kimi independently reproduced two real harness runs with opposite contention
  winners, exact one-of-three duplicate acceptance, sequence rejection after
  restart, stable app hashes, and valid export/reimport. It found no P0/P1/P2.
- Sol implemented Kimi's P3 hardening: post-restart evidence must now contain
  Cosmos code 32 and `account sequence mismatch`, and the repository contract
  pins that assertion. The remediated focused contract and real 90.93s harness
  pass; no review finding remains open.
- Local review is complete and the tree is PR-ready. Protected exact-head CI,
  merge, GH-29/public synchronization, final-main and Pages remain required.

## 2026-08-10 — GH-172 local implementation verified

- Added one opt-in real-process harness and a fail-closed repository/CI
  contract. Two authenticated accounts contend for the same domain; exactly one
  commits and exactly one escrow mutation occurs. The same signed bytes are
  broadcast three times concurrently, after commit, and after a clean same-home
  restart without producing a second state transition.
- PASS: focused repository contract; build; Vet; full Race/Coverage across all
  maintained Go packages; bounded generative/fuzz quality; security contract;
  pinned Staticcheck v0.7.0; pinned Gitleaks v8.30.1 with no leak; docs, JSON,
  YAML and diff integrity. The real process gate passed in 145.38s with stable
  historical app hash and ledger-valid export/reimport.
- Fresh standard arithmetic is 1,607 cases (1,440 Go, 26 Rust, 141 client).
  The separately gated process case does not inflate that sum or GH-29's 23/59
  rollout. Kimi implemented the bounded harness slice; Sol reviewed and fixed
  test-goroutine safety. Claude's small read-only census attempt exceeded its
  configured budget and produced no output or diff. Independent Kimi review is
  running; protected PR/CI, merge and final-main/Pages evidence remain.

## 2026-08-10 — GH-172 concurrent/replay/restart evidence started

- Reconciled exact main, GH-29, the detailed roadmap, project/global Bridges,
  open issues/PRs, and final-main workflows. GH-172 is the next internally
  closable quality/security block; independent external review and production
  topology remain separate external gates.
- Created a clean exact-main worktree and bounded issue. The implementation is
  test/evidence-first and must add shared-state contention, duplicate/exact
  replay, persistence, ledger/invariant, race, and timeout evidence without
  inflating existing GH-97/GH-145 coverage.
- No runtime behavior, deployment, public network, key, wallet, signing, funds,
  migration, or release action has occurred. Kimi receives one secret-free
  bounded slice after Sol completes the source/test inventory.


## 2026-08-10 — GH-169 merged and final-main verified

- PR #170 exact remediation head `620c41d` passed 18/18 contexts, all five
  CodeRabbit threads were resolved, and owner-authorized admin squash bypassed
  only formal self-approval. Merge `d50c131` closes GH-169; the remote task
  branch was deleted.
- Final-main Go `31336109586`, Security `31336109601`, reproducible Linux
  `31336109590`, Docs `31336109615`, and Pages `31336109276` pass. Live Pages
  confirms 1,606 tests, 23/59 overall, 23/51 phase work and the unchanged
  non-production boundary. GH-29 is synchronized; external review remains open.

## 2026-08-09 — PR #170 review remediation verified

- Reproduced and fixed all five CodeRabbit inline findings: two residual IBC
  overclaims, stale agent-guide versions, a stale treasury count, and raw-text
  rather than parsed Go module version validation. Also removed JSON-format
  coupling and used robust EOF matching in the contract test.
- Repeated focused contract, docs consistency, full build/vet/Race/Coverage,
  Staticcheck, secret scan, module verification, JSON and diff checks; all pass.
  No failed check was bypassed; the remediated exact-head CI remains required.

## 2026-08-09 — GH-169 locally verified

- Added a canonical v1 cross-system threat register with 18 threats across all
  nine required domains, plus a human-readable model and 19-mutation fail-closed
  repository contract. All 33 evidence paths exist; high/critical residuals map
  to GH-7/GH-29 and production readiness is structurally false.
- Corrected public dependency, IBC and GH-131 history drift. Fresh package-scoped
  JSON output records 1,439 Go cases; with 26 Rust and 141 client cases the
  source of truth is 1,606. GH-169 moves candidate rollout to 23/59 overall and
  23/51 phase work; Phase 6 remains 6/7.
- Kimi's independent review found no P0/P1. Sol fixed its stale multi-client P2,
  made mutations ID-stable and bound ibc-go into consistency checks. PASS:
  focused contract, docs consistency, full build/vet/Race/Coverage, security
  contract, Staticcheck, secret and vulnerability positive/negative gates,
  module verification, JSON and diff integrity.
- Protected PR/CI, merge, GH-29/public synchronization, final-main and Pages
  readback remain. No external runtime or production action occurred.

## 2026-08-09 — GH-161 merged and final-main verified

- Policy PR #162 merged as `7b3f183`, compatible npm PR #163 as `3a3919c`,
  and patch-only Go PR #164 as `62ad5e9`. Their final exact heads passed 18,
  15, and 18 protected contexts respectively. Policy-generated patch PR #167
  then passed independent Sol/Kimi review and 16 contexts before merging as
  final code head `87c47c1`; zero open review threads remained.
- Exact merged-main Go CI `31315620763`, reproducible Linux
  `31315620777`, Security `31315620755`, Dependency Graph `31315623437`, and
  Pages `31315620166` pass. Local composed-head Go and npm verification also
  passed; Kimi independently approved the bounded policy and replacement
  design after Sol resolved its Playwright-major finding, then separately
  approved all five #167 patches and their sumdb-anchored module hashes.
- Unsafe grouped npm #157 and Go #158 were closed with exact incompatibility
  evidence and superseding links. Generated replacement attempts #165/#166
  auto-closed as superseded; compatible Cargo #156 and policy-conforming Go
  #167 were the generated groups merged directly. Rollout stays 22/59, phase
  work 22/51, Phase 6 6/7, and production false; no rollout gate or production
  action was claimed.

## 2026-08-09 — GH-161 dependency groups independently reconciled

- Refreshed and merged Cargo PR #156 only after exact local Rust and ten-job
  protected evidence passed; merge is `4aeddc0` and its remote branch is gone.
- Rejected npm PR #157 because React Refresh 0.5.3 fails 19 route exports and
  Playwright 1.62.1 violates the frozen macOS runner boundary. A bounded
  seven-package replacement retains both pins and passes local lint, unit,
  policy, build/budget, audit, and Chromium/Firefox browser evidence.
- Rejected Go PR #158 because its 0.x Cosmos SDK migration breaks runtime
  invariants, exports, recovery/capacity, static analysis, and vulnerability
  policy. A four-direct-dependency patch replacement passes full `make verify`,
  module verification, staticcheck, govulncheck policy, and contract tests.
- Added typed fail-closed Dependabot exceptions for Go patch-only maintenance,
  npm compatibility pins, and ungrouped Cargo `cosmwasm-*` crate updates.
  Independent final review approved the replacements and policy; Sol closed
  its Playwright-major P2 by fully freezing that automatic pin. Protected
  publication remains; no production action occurred.

## 2026-08-09 — GH-154 merged and final-main verified

- PR #159 passed 11/11 protected exact-head checks on `46f687b` and was
  squash-merged as `ac8fbbc`; issue #154 is closed and its remote task branch
  is deleted.
- Exact-main Docs `31310820782`, Security `31310820753`, and Pages
  `31310820227` all pass on `ac8fbbc`. GH-154 is Done with no runtime,
  dependency, workflow, production, chain, signing, key, or funds mutation.
- The six intentionally preserved remote heads and six preserved historical or
  closed-unmerged local worktrees remain explicit boundaries. Dependabot PRs
  #156-#158 are separate follow-up maintenance and were not auto-merged.

## 2026-08-09 — GH-154 exact remote cleanup / local cleanup in progress

- Repeated the inventory against current GitHub state after GH-153 merged.
  Exactly 31 current remote SHAs matched merged PR heads and had no open PR;
  all 31 residue branches were deleted after immediate per-ref revalidation.
- Final remote heads intentionally remain: `main`, active Dependabot PRs
  #156-#158, closed-unmerged `agent/public-recovery-status`, and archival
  `solana-archived`.
- Seventeen local worktrees were clean exact merged-PR heads before removal.
  Non-forced removal is still processing large ignored dependency/build trees.
  Dirty, divergent, active, closed-unmerged, and ambiguous worktrees remain
  preserved. No production, deployment, chain, key, signing, or funds action
  occurred.

## 2026-08-09 — GH-154 final local inventory verified

- Removed 17 clean local worktrees only after their current HEAD exactly
  matched a merged PR head; no forced removal was used. Final inventory has
  seven worktrees: six preserved historical/active states and GH-154.
- Preserved the dirty/divergent legacy checkout (18 status entries, 24 commits
  outside origin/main), four clean but divergent historical worktrees (24, 28,
  31, and 41 commits outside origin/main), closed-unmerged recovery status
  (one), and active GH-154. No cleanup process remains.
- Final fetch/prune confirms exactly six intentional remote heads and three
  open PRs, #156-#158, each matching its active Dependabot branch. Docs
  consistency and diff checks pass. Independent review and protected docs-only
  publication remain before Done.

## 2026-08-09 — GH-154 independent final review approved

- Kimi returned APPROVE with no P0-P3 and no write. It reproduced six remote
  heads, seven worktrees, exact dirty/divergence counts, closed-unmerged PR #25,
  open PRs #156-#158, all cited main workflow results, docs consistency, and
  diff cleanliness.
- A probe across 91 distinct merged-PR head branch names found zero remaining
  remote residue. The initial 16 exact worktrees plus completed GH-153 account
  for the final 17 non-forced removals. Protected publication remains.

## 2026-08-09 — GH-154 PR #159 documentation review remediated

- Corrected project-state/TODO sequencing so independent review is complete
  and only protected publication remains. Split remote and local inventories
  into explicit lists without changing any verified count or preservation
  boundary. No code, dependency, workflow, or runtime change occurred.

## 2026-08-09 — GH-153 protected merge and exact-main checkpoint

- PR #155 passed all 18 exact-head checks on `3f698a2`, then merged as
  `8489d54`; GH-153 closed and its task branch was removed.
- Exact-main Go `31309568219`, Security `31309568223`, reproducible Linux
  `31309568233`, Docs `31309568209`, Dependency Graph `31309569964`, and Pages
  `31309567545` pass, including bounded multi-validator recovery. GH-153 is
  Done.
- The bounded policy generated minor/patch PRs #156-#158. Cargo/Go generator
  runs additionally expose updater/resolution failures for CosmWasm and an
  incompatible Wasmd/SDK candidate; these remain explicit follow-up work and
  no dependency update is auto-merged.

## 2026-08-09 — GH-153 bounded Dependabot maintenance started

- GH-148 merged through PR #149 as `5baeea6`; its security, static, secret,
  Rust, client, docs, reproducible-build, and Pages checks pass on `main`.
- The first Dependabot sweep opened failing PRs #150–#152 because wildcard
  groups combined incompatible major updates with normal maintenance. Opened
  GH-153 and branch `fix/GH-153-bounded-dependabot` from exact clean main.
- Scope is limited to minor/patch update policy, repository contract evidence,
  and classification/closure of the broken generated PRs. No production or
  runtime action occurred.

## 2026-08-09 — GH-153 compatible update policy verified

- Added explicit version-update minor/patch grouping and wildcard semver-major
  ignores for Actions, Go, Cargo, and npm. A repository negative test rejects
  any missing ecosystem exclusion.
- Focused repository contract, YAML parse, and diff checks pass. Closed failing
  PRs #150–#152 with evidence and deleted only their generated Dependabot
  branches; no dependency change was merged.
- Complete repository verification, independent review, protected PR, merge,
  final-main/Pages, and Bridge closeout remain pending.

## 2026-08-09 — GH-153 structural policy hardening

- Replaced Dependabot string-count assertions with typed YAML validation per
  exact ecosystem/group/ignore after Kimi's P3 robustness observation. Invalid
  nesting, duplicate/unexpected ecosystems, and missing or loosened policy now
  fail the repository contract.
- `go mod tidy` reclassified only three existing pinned direct imports; no
  version or checksum changed. Full `make verify` passes all 1,412 Go cases
  under Race/Coverage; docs, YAML, and diff checks pass.
- Ordinary major version updates are suppressed. Security-update PRs remain an
  intentional separately reviewed exception and are never auto-merged.

## 2026-08-09 — GH-153 final local approval

- Kimi's final read-only review returned APPROVE with no P0-P3 and made no
  change. Repeated full Go build/vet/Race/Coverage passes all 1,412 cases;
  module verify/tidy, docs, typed YAML, and diff checks pass.
- Exact pinned gitleaks scans the 4.04 MB maintained tree with zero findings;
  secret negative fixtures and exact pinned staticcheck pass. Publication and
  protected/final-main evidence remain before Done.

## 2026-08-09 — GH-153 PR #155 published

- Pushed `5f188e6` and opened PR #155, linked to close GH-153 after merge.
  Protected exact-head CI/review, merge, and final-main/Pages evidence remain.






## 2026-08-09 — GH-148 security-gate baseline audit and implementation

- Created GH-148 from exact clean `origin/main` `c1a0ed4` for the single GH-29
  Phase-5 dependency/static/secret/supply-chain checkbox. Kimi's independent
  read-only baseline audit found the Go vulnerability step fail-open on scanner
  errors, mutable tools/actions, no repository secret gate, optional staticcheck,
  incomplete Cargo lock enforcement, missing update automation, and unbounded
  jobs; no file was changed by Kimi.
- Centralized exact scanner/toolchain and approved Action SHA pins, pinned all
  35 third-party Action uses, bounded every job, made govulncheck and staticcheck
  fail closed over the repository package selector, added current-tree gitleaks
  scanning plus planted positive/negative fixtures, enforced Go/Cargo/npm locks,
  and added weekly grouped Dependabot updates for Actions/Go/Cargo/npm.
- The repository contract rejects mutable/unapproved Actions, failure bypasses,
  missing secret negative tests, broad secret allowlists, and missing lockfiles.
  Staticcheck exposed and closed 14 baseline findings without changing protocol
  semantics; four gitleaks matches were exact public synthetic strings and have
  exact-regex exceptions only.

## 2026-08-09 — GH-148 focused gates and candidate status verified

- PASS: six positive/negative security-contract cases; blocking staticcheck;
  maintained-tree gitleaks with zero findings; planted credential rejection;
  strict govulncheck with four exact, active, no-fix policy entries;
  cargo-audit with zero vulnerabilities and five warning-only advisories; npm
  install/audit with zero high/critical findings; all documentation consistency,
  retirement, shell, workflow YAML, Dependabot YAML, JSON, and diff checks.
- PASS: maintained client lint, 10 Node policy/budget cases, 131 Vitest cases,
  build at 76.40 kB gzip initial / 5.03 kB max route / 353.56 kB total, and
  guarded audit. Ten-second DEX/token fuzz plus focused deterministic/race gate
  passes. Full Go and Rust verification continue.
- Candidate arithmetic is 1,579 standard cases (1,412 Go + 26 Rust + 141
  maintained client), rollout 22/59 overall and 22/51 phase work, Phase 6 6/7,
  and `production_ready=false`. No production, deployment, release, real key,
  account, fund, signing, or public-network action occurred.

## 2026-08-09 — GH-148 final-review blockers repaired

- Kimi's read-only final review reproduced a P1: bare govulncheck exits 3 for
  the four reachable no-fix findings and would keep protected CI red. It also
  found a local-only secret-scan false positive in ignored Rust build output.
- Replaced the bare invocation with a fail-closed JSON policy wrapper. Only the
  exact four IDs may pass, only with no published fixed version and active
  dated exceptions bounded to 30 days; unknown, fixable, missing, stale,
  expired, malformed, or scanner-error fixtures are rejected.
- Scoped local secret scanning to tracked plus non-ignored changed files. With
  existing Node/Rust build artifacts present it scans 4.02 MB, reports zero
  leaks, and still rejects the planted credential. The live Go policy gate and
  all positive/negative wrapper fixtures pass.

## 2026-08-09 — GH-148 remediation re-review hardened

- Re-review confirmed the original exit-3 and ignored-build-output findings
  closed, then identified standalone metadata-validation and enumeration-error
  gaps. The executable Go gate now validates real UTC dates, ID/reason types,
  duplicates, and the central 1..30-day maximum itself.
- Expanded fixtures reject long, malformed, stale, duplicate, expired, unknown,
  fixable, and scanner-error states. Secret-tree enumeration now has a checked
  manifest, a non-empty invariant, and a failing-Git regression fixture.
- PASS: expanded fixtures, live maintained-tree gitleaks and planted credential,
  repository contract, shell/diff checks, and full repeated Go build/vet/Race/
  Coverage across all 13 maintained packages.

## 2026-08-09 — GH-148 PR #149 exact Rust components

- First exact-head CI exposed one integration-only failure: pinned Rust 1.95.0
  did not include cargo-fmt by default. Added explicit `rustfmt, clippy`
  components to the same immutable toolchain action.
- The failed run changed no source or status claim. Fresh exact-head CI remains
  mandatory; merge stays blocked until every required job passes.

## 2026-08-09 — GH-145 generative quality-depth inventory

- Confirmed zero existing Go fuzz targets and zero generative property-test
  harnesses. Existing invariant, malformed-genesis, replay, and global race
  evidence is substantial but primarily example-driven.
- Scoped GH-145 to DEX constant-product properties, PNYX cap/bank-genesis
  validation, governance replay safety, and bounded CI fuzz/race execution.
- Kimi K3 performs the bounded three-test-file implementation; Sol retains CI,
  security, integration, complete gates, GitHub, and closure. No production or
  real-funds action is in scope.

## 2026-08-09 — GH-145 implementation integrated

- Added two real seeded Go fuzz targets plus deterministic DEX arithmetic,
  PNYX supply/genesis, and consensus-slashing replay properties. Malformed
  genesis cases cover legacy denomination, duplicates, negatives, invalid
  address, mismatched supply, and over-cap structures without production edits.
- Added a bounded repository/CI quality-depth gate: the deterministic critical
  tests run under Race and each fuzz target runs for a configurable 1–60
  seconds with one worker. Repository contract, shell syntax, focused Race,
  and one-second integrated campaigns pass.
- Kimi's independent implementation verification passed full affected-package
  suites, focused Race, vet, and separate 20-second fuzz campaigns with 40,317
  DEX and 25,524 token executions. Sol reviewed and integrated every write.
- Fresh package-scoped JSON arithmetic is 1,406 Go cases, making the candidate
  total 1,573 with unchanged Rust/client counts. Staged rollout is 21/59
  overall and 21/51 phase work; Phase 6 is 6/7 and production remains false.
- Full gates, independent final review, protected CI, merge, GH-29, final-main,
  and Pages evidence remain pending.

## 2026-08-09 — GH-145 full local gates pass

- `make verify` passes package selection, build, vet, and all 1,406 Go cases
  under Race/Coverage. Critical coverage remains above its protected floors;
  token coverage rises to 95.6% while root is 71.9%, DEX 51.1%, and governance
  63.8%.
- The protected-equivalent quality gate passes 10-second campaigns with 12,694
  DEX and 11,537 token executions. Documentation consistency, status/module/
  rollout arithmetic, JSON, workflow YAML, shell syntax, and diff hygiene pass.
- Local implementation is complete without production-code changes or blocker.
  Independent final review and the complete GitHub/merge/final-main/Pages chain
  remain mandatory before Done.

## 2026-08-09 — GH-145 independent review approved

- Kimi's read-only full-diff review reports 0 P0/P1/P2 and independently
  reproduces property validity, replay conflict semantics, package counts,
  token coverage, the 10-second fuzz gate, documentation arithmetic, vet,
  workflow/shell parsing, and diff hygiene. It changed no file.
- Remediated its only technical P3 by increasing the outer quality-depth job
  timeout from 10 to 15 minutes for cold-cache compile headroom. Per-command
  180-second timeouts, 1–60-second fuzz bounds, and single-worker execution
  remain unchanged. The other observations are accurately staged status and
  machine-dependent per-run execution evidence, not defects.

## 2026-08-09 — GH-145 protected PR published

- Published locally verified commit `d4aaa77` through PR #146; the PR links
  `Closes #145` and preserves the tests/CI/docs-only boundary.
- Exact-head protected workflows and review are running. Merge, GH-29 tracker
  synchronization, final-main checks, Pages readback, and final Bridge closure
  remain mandatory before Done.

## 2026-08-09 — GH-145 merged and fully verified

- Every protected exact-head check passed on `76b5d80`; PR #146 merged as
  `50f18e0`, GH-145 closed, and GH-29 now records exactly 21/59 with linked
  evidence. Authorized admin merge bypassed only formal self-approval.
- Final-main Go CI `31280519207`, Security `31280519224`, and Pages
  `31280518754` pass. Go CI includes quality-depth 4m31s, build/Race/coverage
  8m17s, Docker 7m10s, capacity 5m15s, and recovery 14m00s.
- Live Pages readback publishes 1,573 cases (1,406 Go, 26 Rust, 141 maintained
  client), rollout 21/59 overall and 21/51 phase work, Phase 6 6/7, production
  false. GH-145 is Done with no production or real-funds action.

## 2026-08-08 — GH-139 critical-path coverage locally verified

- Added a repository-owned three-package coverage contract and protected Go CI
  gate with floors of 71.8% root/application, 51.0% DEX, and 63.7%
  governance. Fresh measurements pass at 71.9%, 51.1%, and 63.8%.
- Added 15 standard Go cases: atomic output rollback and error joining; DEX
  authority, custody, liquidity, slippage, and legacy fail-closed boundaries;
  governance escrow, signer/admin authorization, VoteToEarn payout, withdrawal;
  and repository contract enforcement.
- Kimi K3 implemented the bounded 14-case test block. Sol reviewed all writes,
  hardened an event assertion against setup-event false positives, integrated
  CI/docs, and reran focused and coverage verification.
- Staged public arithmetic is 1,534 standard cases (1,367 Go, 26 Rust, 141
  maintained client) and rollout 20/59 (20/51 phase work), Phase 6 6/7,
  production false. Full repository gates, independent final review, protected
  PR CI, merge, tracker synchronization, Pages, and final-main evidence remain.
- Full `make verify` passes build, vet, and all 1,367 Go cases under race and
  coverage. Independent Kimi review reports no P0/P1/P2 and independently
  reproduces the 15 focused race cases, exact full-suite count, coverage
  contract, vet, docs consistency, and workflow parsing. Publication remains.
- Commit `a2b8473` is published through PR #140. Protected exact-head checks,
  merge, tracker synchronization, Pages, and final-main verification remain.

## 2026-08-08 — GH-141 client Docker build-context repair

- PR #140 Docker smoke exposed a pre-existing maintained-client image defect:
  `npm run build` requires `scripts/bundle-budget.mjs`, but the builder copied
  only config and `src`.
- The builder now copies `scripts/` before build, and a repository contract pins
  the copy order plus continued budget invocation. Clean install, 10 Node
  policy cases, 131 Vitest cases, build, and bundle budget pass locally.
- Full `make verify` passes all 1,353 Go cases under race/coverage; client lint,
  docs consistency, and diff checks also pass.
- Commit `14d6851` was published through PR #142. Independent read-only review
  found no functional P0/P1/P2. All protected exact-head gates passed,
  including the authoritative Docker image/restart smoke, and PR #142 merged
  as `11ef2f6`; GH-141 is closed.
- Local Docker CLI is unavailable; protected Docker image/restart CI remains
  the real container proof. The merged base is 1,520 cases; rollout stays 19/59
  and production false until GH-139 completes.

## 2026-08-08 — GH-139 rebased on the Docker repair

- Merged repaired `origin/main` (`11ef2f6`) into PR #140 and reconciled the
  public count to 1,535 cases: 1,368 Go, 26 Rust, and 141 maintained client.
- Rollout remains staged at 20/59 overall and 20/51 phase work, Phase 6 6/7,
  production false. Fresh combined local and protected exact-head evidence is
  required before merge.

## 2026-08-08 — GH-139 merged and final-main verified

- PR #140 passed all protected exact-head checks with no unresolved review
  thread and merged as `1b6ccef`; GH-139 is closed.
- Exact merge commit Go CI `31277242596`, Security `31277242587`, and Pages
  `31277242292` pass. Live Pages reads back 1,535 total cases, 1,368 Go,
  rollout 20/59 overall and 20/51 phase work, Phase 6 6/7, production false.
- GH-29 now links GH-139/PR #140 and marks the Phase 5 critical-path coverage
  item complete. GH-139 is Done; remaining rollout and release-security gates
  stay open.

## 2026-08-08 — GH-131 submitted-history implementation verified

- Replaced the placeholder wallet history with newest-first submitted-only
  Cosmos SDK v0.50 REST pagination, typed failure/decode handling, stale-wallet
  protection, and an accessible explicit incoming-exclusion UI.
- Kimi implemented the bounded block. Sol independently corrected deprecated
  request semantics, cache/abort/count boundaries, wallet-change clearing,
  request cancellation, retry/stale-row behavior, component inventory, and
  committed-send reporting. Independent review
  found no P0-P2 after remediation.
- Ten Node policy cases, 131 standard Vitest cases, two separately gated real
  disposable-chain cases, lint, build/budgets, and guarded audit pass. Public
  status stages 1,519 standard cases and rollout 19/59; protected PR and final-
  main evidence remain.
- CodeRabbit's four valid review findings are remediated: browser fetch receiver,
  retryable dynamic-import failures, bounded invalid timestamps, synchronized
  wiki arithmetic, and a self-contained pagination chain fixture. Focused
  37/37, lint/build/budgets, diff check, and the standalone real-chain history
  case pass. The complete sequential disposable-chain file also passes 2/2 in
  84.53s; the protected rerun and merge remain.
- PR #137 merged as `7c751b2` after exact-head CI passed and all four review
  threads were resolved. GH-131 is closed; GH-29 is synchronized at 19/59.
  Final-main Client Web CI `31266901488`, Security `31266901498`, and Pages
  `31266901092` pass; live Pages reads back 1,519 cases, 19/59 overall, 19/51
  phase work, Phase 6 at 6/7, and production false. GH-131 is Done.

## 2026-08-08 — GH-132 protected browser qualification

- Added pinned Chromium, Firefox, and WebKit desktop/mobile qualification for
  the safe unauthenticated maintained-client routes, with serious/critical
  accessibility, desktop keyboard/focus, responsive overflow, delayed lazy
  loading, and third-party-request assertions.
- Exact PR #136 CI passes browser quality, client build, disposable-chain
  integration, docs, security, retirement, and independent review gates.
- Rollout advances to 18/59 (18/51 phase tasks). Authenticated wallet/key,
  signing, real-fund, deployment, and production gates remain open.

## 2026-08-08 14:50 EEST - GH-128 merged and closed

- PR #129 exact head `dcc1113` passed Client build, real disposable-chain
  client integration, Docs, Go/Rust/Node security, all retirement contracts,
  DeepScan, and formal review statuses with no unresolved review thread.
- Squash merge `47f8c08` closed GH-128. GH-29 now marks route splitting and
  its deterministic bundle-performance budget complete.
- Final-main Client Web CI `31255695662`, Security Scan `31255695671`, and
  Pages `31255695086` passed on exact merge commit `47f8c08`.
- Live Pages readback confirms 1,482 recovery-verified cases, rollout 17/59,
  and 19 lazy routes with deterministic budgets. Phase 6 remains 6/7 and
  production readiness remains false.

## 2026-08-08 14:27 EEST - GH-128 locally verified

- Split all 19 maintained-client page routes while preserving the exact
  21-route contract, added an accessible loading boundary and rejected-chunk
  recovery regression, and deferred wallet signing/query/transaction stacks.
- Reduced the deterministic initial JavaScript entry to 75.79 kB gzip from the
  322.63 kB Vite baseline. The build now enforces raw/gzip entry, route, chunk,
  and total budgets from a hash-independent Vite manifest.
- PASS: lint, 104 client cases, TypeScript/Vite production build, budget gate,
  live High audit, and final-state disposable-chain client delivery. Kimi
  independently found the eager dependency trap, Docs CI gap, and a possible
  concurrent first-service duplication; Sol fixed the latter with one cached
  Promise per service. Spark independently reviewed the lazy-boundary test design.
- Public docs stage 1,482 cases and rollout 17/59 after GH-128 publication;
  Phase 6 remains 6/7 and production readiness false. Protected PR checks,
  review, merge, GH-29 synchronization, Pages, and final-main evidence remain.
- Exact implementation commit `b026739` is published through PR #129; all three
  new files called out by final review are included. Exact-head checks remain.

## 2026-08-08 13:13 EEST - GH-121 merged and closed

- PR #126 exact head `d7a41b0` passed Client, real client-chain integration,
  Docs, Go build/test, capacity, Docker restart, multi-validator recovery,
  Security, DeepScan, and CodeRabbit status checks.
- All eight CodeRabbit review findings were answered and resolved. Independent
  Kimi review found no open P0/P1/P2; both final P3 observations were fixed and
  regression-tested by Sol.
- Administratively squash-merged as `239e6c3` after only the formal one-review
  requirement remained. GH-121 closed automatically. Final-main Client
  `31252288314`, Security `31252288325`, Go `31252288303`, and Pages
  `31252287866` all passed.
- Public arithmetic remains 1,476 cases, rollout 16/59, Phase 6 6/7, and
  `production_ready=false`. No production, deployment, key, account, signing,
  broadcast, or fund action occurred.

## 2026-08-08 12:04 EEST - GH-121 locally verified

- Replaced all 23 maintained-browser call sites for 18 unregistered
  custom-module REST aliases with one registered protobuf gRPC-over-CometBFT
  ABCI JSON-RPC client. Canonical Domain data now drives governance projections;
  DEX and validator/domain payloads are validated at runtime and all failures
  remain explicit.
- Added registered MerkleProof and PayToPut query methods, CLI support,
  deterministic/fail-closed handler tests, route execution contracts, and one
  real disposable-chain browser integration test. No listener, gateway,
  signing, broadcast, deployment, account, key, or fund behavior was widened.
- Kimi K3 independently inventoried the broken boundary and implemented the
  bounded Go query block. Sol reviewed its full write diff, integrated every
  browser consumer, hardened response validation, and reran the complete gates.
  A bounded Claude Code docs attempt exceeded its small budget without usable
  output; Sol completed documentation.
- PASS so far: full Go build/vet/race/coverage with 1,352 cases, Rust
  fmt/clippy/build/26 tests/audit, maintained-client clean install/lint/98
  tests/build at 322.63 kB gzip/live High audit, and real client-chain query
  delivery. Independent final-head review and protected GitHub checks remain.
- Kimi's final read-only review found no P0/P1/P2. Sol remediated its four
  substantive P3 observations: fixed32/fixed64 decoder bounds, validator item
  validation, and explicit error UI for both direct LP-position consumers.
- PR #126's first protected head passed all workflows. CodeRabbit then reported
  eight valid inline findings; remediation preserves address-unbound anonymous
  chain state, binds UI identity status only to the locally owned commitment,
  prevents stale Pay-to-Put quotes, bounds RPC waits, and aligns validation,
  canonical hex, documentation, and published arithmetic. Rerun pending.
- Final Kimi review found no P0/P1/P2 and two P3s. Sol fixed response-body
  timeout classification plus the stale privacy comment, added regression
  coverage, and reran every local gate successfully. Protected exact-head CI,
  review-thread resolution, merge, and final-main evidence remain.
- Public arithmetic is 1,476 cases; rollout remains 16/59 overall and Phase 6
  remains 6/7 because GH-121 does not satisfy a production rollout exit.

## 2026-08-07 02:30 EEST - GH-116 merged and closed

- PR #124 final head `df465e8` passed PR-native Go CI `31130371098`, Docs
  `31130368575`, Security `31130368777`, DeepScan, and all exact-head manual
  Client/Rust/Docs/Security/Go runs. No review thread remained.
- CodeRabbit was temporarily rate-limited and performed no substantive review;
  two Kimi read-only reviews plus Sol review found 0 P0 / 0 P1 / 0 P2.
- Administratively squash-merged as `5dc4487` because only the formal one-
  review rule remained. GH-116 closed automatically.
- Final main passed Go `31131404107` attempt 2, Docs `31131405747`, Client
  `31131407330`, Rust `31131408773`, Security `31131410151`, and Pages
  `31131318674`. Go attempt 1 exposed one transient legacy-migration P2P
  startup miss; the unchanged commit passed all eight recovery harnesses on
  the controlled rerun, so no runtime patch was justified.
- Public arithmetic is 1,450 cases, rollout 16/59, Phase 6 6/7, production
  false. GH-121 remains the browser query-transport blocker. No production or
  secret-bearing action occurred.

## 2026-08-07 01:50 EEST - GH-116 locally verified on patched main

- Removed the consumerless custom ABCI routing override and dead module
  queriers while preserving all 17 protobuf gRPC Query routes and proving live
  governance/DEX execution through ABCI. Retired paths fail closed.
- Kimi's bounded independent census found no legacy maintained consumer and no
  P0; DEX coverage and every active API claim were migrated. GH-121 now owns
  the separate pre-existing browser REST transport defect. Its final read-only
  implementation-head review reproduced the route census, build, vet, focused
  tests, exact module counts, and contracts with 0 P0 / 0 P1 / 0 P2.
- PASS: full Go build/vet/race/coverage (1,333 cases), four focused persistence/
  restart/export/import tests, Rust fmt/clippy/build/test (26 cases) and audit,
  maintained-client install/lint/91 tests/build/live High audit, documentation,
  retirement, shell, workflow YAML, JSON, and diff contracts.
- Rebased onto GH-122 / PR #123 merge `1257484`; protected GH-116 review and
  merge remain. No production or secret-bearing action occurred.
- Final Sol documentation reconciliation corrected stale CLI inventories to
  24 governance tx + 7 query and 7 DEX tx + 9 query commands.

## 2026-08-07 01:45 EEST - GH-122 prerequisite merged

- PR #123 merged as `1257484` after exact-head Client Web CI `31129152758` and
  Security Scan `31129153109` passed; its single chronology review thread was
  corrected and resolved.
- The patched `js-yaml` development-tool lock resolution is now on `main`
  without audit-policy or application behavior changes.

## 2026-08-07 01:32 EEST - GH-122 local repair verified

- Updated only the transitive `js-yaml` lock resolution from vulnerable 4.3.0
  to patched 4.3.1; the application and fail-closed audit policy are unchanged.
- PASS: clean install, lint, 85 Vitest plus six audit-policy cases, production
  build at 318.91 kB gzip, and live High audit. Only the exact existing
  BrowserRouter-bound React Router exception remains accepted.
- Protected PR checks and merge remain before GH-116 can resume publication.

## 2026-08-07 01:30 EEST - GH-122 prerequisite started

- GH-116 full client verification reproduced new live High advisory
  GHSA-5p4m-2wfm-xmqj in unchanged transitive `js-yaml` 4.3.0 through ESLint.
- Created clean `fix/GH-122-js-yaml` from exact `origin/main` `aa3ffa6` for a
  compatible lockfile-only repair before GH-116 publication.
- No application source, audit weakening, deployment, production, key,
  signing, broadcast, account, or fund action is in scope.

## 2026-08-07 01:08 EEST - GH-116 started

- Created clean branch `refactor/GH-116-retire-custom-query` from exact
  `origin/main` `aa3ffa6` after confirming GH-116 is open, no PR is open, and
  the latest main Security Scan and Pages workflows pass.
- Scope is limited to an evidence-backed decision on the retired clients'
  `custom/truedemocracy/...` and `custom/dex/...` ABCI compatibility shim,
  preserving supported gRPC/gateway query behavior.
- Kimi will independently inventory consumers and compatibility risks; Sol
  owns the final diff, tests, publication, merge, and handoff.
- No deployment, infrastructure, production, key, signing, broadcast,
  account, or fund action is in scope.

## 2026-07-23 19:40 EEST - GH-59 ABCI++ validator slashing

- Wired BaseApp's deterministic BeginBlock ABCI++ boundary to CometBFT
  misbehavior and decided-last-commit ingestion. Canonical commit cursors,
  normalized infraction IDs, processed markers, consensus-key active intervals,
  tombstones, and operator-scoped rolling liveness bitmaps now survive restart
  and export/import.
- Equivocation burns 5% exactly once per tombstoned consensus key, jails the
  operator, emits power zero, and remains attributable after key rotation.
  Downtime burns 1% only after more than 50 misses in a complete rolling
  100-block window, then jails and emits power zero.
- Closed the exit-before-evidence path: authenticated full removal and full
  withdrawal retain the complete stake in module escrow until both CometBFT
  evidence-age limits are strictly exceeded. Delayed evidence slashes the held
  claim before payout.
- Extended genesis validation/export/import for historical keys, replay state,
  liveness, jail state, processed infractions, and pending exits. A jailed
  below-minimum validator now round-trips without being resurrected or making
  the exported genesis invalid.
- Added the operator slashing/incident runbook and the gated real four-process
  harness. The harness proves 100-block downtime, exact 1% burn, restart and
  catch-up as a full node, real CometBFT duplicate-vote evidence broadcast,
  exact 5% burn, both power-zero transitions, common app hash, and ledger-valid
  export. The first diagnostic run used production's five-second commit
  interval and correctly exceeded a four-minute observation deadline; the
  test-only temporary homes now use bounded accelerated consensus timeouts.
  The complete real-process rerun passes in 200.95 seconds.
- Local normal `go test ./...` passes. Focused ABCI++, replay, rotation,
  liveness, exit-hold, and export/import regressions pass. Recovery arithmetic
  is 726 cases: 692 Go, 26 Rust, and eight maintained-client. The first
  adversarial review exposed an Unjail cursor halt, historical-key/hold reuse,
  partial-withdrawal custody, replay-binding, and genesis-relation gaps; all are
  remediated with focused regressions. The final review found 0 P0 / 0 P1 /
  0 P2. PR #65 passed all 11 final-head checks and merged as `934a042`; GH-59
  closed and GH-29 was synchronized.

## 2026-07-23 06:04 EEST - GH-55 validator identity recovery start

- Confirmed synchronized, clean `main` at `93f0263` with no open PRs before
  opening the next Phase 1 rollout task.
- Opened GH-55 for validator consensus-identity cold custody, single-signer
  failover, and compromise containment; created branch
  `feature/GH-55-validator-identity-recovery`.
- Audited the boundary between `node_key.json` (P2P identity) and the coupled
  `priv_validator_key.json` plus current `priv_validator_state.json`
  consensus-safety unit. Found unsafe key-only restore, plaintext full-home,
  configuration, Docker-volume, and pre-upgrade archive guidance.
- Confirmed the protocol does not yet support safe key rotation:
  `remove-validator` withdraws stake subject to a domain transfer limit, while
  bootstrap operator authority is derived from the consensus key. Opened
  follow-up GH-56 for authenticated atomic rotation, permanent revocation, and
  separated bootstrap operator authority.
- Implementation in progress: real four-validator cold-failover evidence, a
  fail-closed custody/incident runbook, corrected backup/security guidance,
  CI coverage, rollout status, and Bridge synchronization.
- Local evidence is green: the focused real-process failover passes in 62.17s;
  the full 656-case Go suite passes; build, vet, shell syntax, JSON, docs
  consistency, and diff checks pass; and all five multi-process recovery drills
  pass together in 636.342s. The new drill proves exact post-stop signer-state
  transfer, a distinct P2P identity, strict signing-position advance, unchanged
  consensus power, common app hash, source isolation, valid export/ledger, and
  re-import.
- The structured GH-55 pre-ship audit records 0 FAIL / 1 WARN / 5 PASS. Its one
  HIGH warning is the separate GH-56 protocol rotation gap, not a defect in the
  bounded cold-failover path.
- PR #57 is published. CodeRabbit raised eight review threads; four valid
  findings removed a duplicated word, retained compromised-key recovery as a
  separate blocker, made permission examples honor `CHAIN_HOME`, and required
  dedicated service-account ownership verification. The two future-date
  findings conflict with the authoritative July 23 runtime, and the two Go
  wrapper findings are invalid because the multi-package wrapper cannot combine
  `go build -o <file>` with multiple package arguments while the existing `.`
  commands already select the root package explicitly.
- PR #57 final head `aa66aa9` passed Go build/vet/race/coverage (6m55s), all
  five multi-validator recovery drills (8m44s), Docker restart (3m33s), docs,
  Go/Rust/Node security, DeepScan, and CodeRabbit. All eight review threads are
  resolved. The PR squash-merged to `main` as `e8670c6` and closed GH-55.

## 2026-07-23 04:54 EEST - GH-53 persisted binary upgrade and rollback start

- Confirmed `main` was synchronized with `origin/main` at `312752a`, with no
  open PRs and only parent issues #4, #7, and #29 open.
- Opened GH-53 under rollout tracker GH-29 and created branch
  `feature/GH-53-persisted-upgrade-rollback`.
- Audited the upgrade boundary: versioned binary builds are supported, but
  `x/upgrade`, migration handlers, and governance-controlled halt/resume are
  not wired. Scope is therefore compatible binary replacement and
  fail-before-open rollback, not consensus-breaking state migration.
- Found and removed unsafe operator guidance that archived/restored the entire
  validator home and could regress `priv_validator_state.json` after newer
  heights were signed.
- Added `TestMultiValidatorPersistedBinaryUpgradeRollback`: four real
  validators commit a funded domain, roll one-by-one to a separately versioned
  compatible artifact, exercise an intentional candidate failure before state
  is opened, and roll every node back to the baseline artifact on the same
  homes. The drill checks historical/current app hashes, validator power,
  unchanged node/validator identity keys, non-regressing signing positions,
  exported ledger validity, persisted application state, and re-import.
- Local focused compile/SKIP verification passes. The gated real-process drill
  passes in 203.97s. The full repository suite passes (`truerepublic` in
  55.207s plus all module packages). The CI-equivalent combined consensus,
  trusted state-sync, backup/restore, and upgrade/rollback process run passes
  in 453.076s; the new drill completes in 193.69s inside that sequence.
- Published PR #54. Its first final-head run passed all checks. CodeRabbit
  reported six review points; five valid findings hardened the outer job
  timeout, trusted multi-RPC checkpoint comparison, explicit checkpoint-height
  interpolation, systemd `ExecStart` reset, and pre-merge status wording. The
  future-date finding was rejected because the runtime date was already July
  23 in both EEST and UTC. Every review thread was answered and resolved.
- PR #54 final head `750040b` passed Go build/race/coverage in 7m04s, the
  expanded multi-validator recovery job in 8m06s, Docker restart in 3m35s,
  docs consistency, Go/Rust/Node security, DeepScan, and CodeRabbit. Squash-
  merged as `3e44905`; GitHub closed GH-53.

## 2026-07-19 03:31 EEST - GH-47 CI build-and-test timeout

- Opened GH-47 after the PR #46 merge-commit Go CI run left `build-and-test`
  in progress for hours while `multi-validator-recovery` and
  `docker-restart-smoke` were green.
- Reproduced the normal suite locally: `go test ./...` PASS in 58.913s.
- Added a 20-minute timeout to the Go CI `build-and-test` job. The test command
  itself is unchanged.
- Pushed the fix to `main` as `63b76bf`. GitHub `main` checks are green:
  Go CI (`build-and-test`, `multi-validator-recovery`,
  `docker-restart-smoke`), Security Scan, and Pages.

## 2026-07-19 02:52 EEST - GH-45 backup/restore/export/import start

- Confirmed `main` is synchronized with `origin/main` at `d121d34`.
- Confirmed no open PRs; only #4, #7, and #29 were open before creating the
  next Phase 1 task.
- Opened GH-45 for the next GH-29 Phase 1 gap: backup, restore, export, and
  import disaster-recovery drills.
- Created branch `feature/GH-45-backup-restore-drill`.
- Hardened `scripts/backup.sh` so chain-data backups exclude node keys,
  validator keys, validator signing state, and keyrings.
- Added `scripts/restore.sh`, which restores sanitized backup data into an
  already initialized target home while preserving the target's local keys.
- Added `TestMultiValidatorBackupRestoreExportImport`, a gated process harness
  that backs up a live full node, verifies the backup archive contains no
  private/signer artifacts, restores it into a fresh home, restarts/catches up,
  verifies app-hash convergence, exports the restored state, validates ledger
  invariants, and re-imports the exported genesis.
- Local evidence so far:
  `bash -n scripts/backup.sh scripts/restore.sh` PASS; `go test . -run
  'TestMultiValidatorBackupRestoreExportImport|TestConfigureGenesisValidatorSetBuildsExactBankBackedSet'
  -count=1 -timeout=300s -v` PASS/SKIP as expected without the smoke env;
  `TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test . -run
  TestMultiValidatorBackupRestoreExportImport -count=1 -timeout=420s -v` PASS
  in 88.224s; `go test ./...` PASS in 58.843s; `bash
  scripts/check-consistency.sh` PASS; `TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1
  go test . -run
  '^(TestMultiValidatorConsensusRecovery|TestMultiValidatorTrustedSnapshotStateSync|TestMultiValidatorBackupRestoreExportImport)$'
  -count=1 -timeout=720s -v` PASS in 290.498s.
- Published PR #46. The first GitHub `multi-validator-recovery` run exposed an
  existing trusted state-sync timeout in the combined CI-smoke command, not a
  failure in the new backup/restore harness.
- Hardened the state-sync waits from 120s to 180s and raised the CI smoke
  timeout from 720s/12m to 900s/15m. Focused local verification:
  `TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test . -run
  TestMultiValidatorTrustedSnapshotStateSync -count=1 -timeout=420s -v` PASS
  in 127.784s.
- Refreshed PR #46 GitHub checks are green: `build-and-test`,
  `multi-validator-recovery`, `docker-restart-smoke`, docs `check`, CodeRabbit,
  DeepScan, Go/Rust security scans, and Node audits.
- Merged PR #46 to `main` as
  `26bf44b7933c25f379db475fd34d2cfb8e49c626`; GitHub automatically closed
  GH-45. GH-29 remains open as the parent rollout tracker.

## 2026-07-19 01:42 EEST - GH-43 trusted snapshot state sync start

- Opened GH-43 for the next GH-29 Phase 1 gap: trusted snapshot state sync
  catch-up.
- Created branch `feature/GH-43-trusted-snapshot-state-sync`.
- Added `TestMultiValidatorTrustedSnapshotStateSync`, a gated process harness
  that enables snapshots on four trusted validators, commits a real
  `create-domain` transaction, derives trust height and trust hash from a
  trusted RPC endpoint, starts a fresh non-validator node with state sync
  enabled, and verifies catch-up to a common app hash.
- Added harness helpers for state-sync-safe start flags, scoped `[statesync]`
  config patching, and trusted block-hash lookup through CometBFT RPC.
- Local evidence:
  `go test . -run
  'TestMultiValidatorTrustedSnapshotStateSync|TestConfigureGenesisValidatorSetBuildsExactBankBackedSet'
  -count=1 -timeout=300s -v` PASS/SKIP as expected without the smoke env;
  `TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test . -run
  TestMultiValidatorTrustedSnapshotStateSync -count=1 -timeout=360s -v` PASS
  in 130.528s; `TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test . -run
  '^(TestMultiValidatorConsensusRecovery|TestMultiValidatorTrustedSnapshotStateSync)$'
  -count=1 -timeout=480s -v` PASS in 197.835s; `go test ./...` PASS in
  65.114s; `bash scripts/check-consistency.sh` PASS.
- Published PR #44, waited for green GitHub checks (`build-and-test`,
  `multi-validator-recovery`, `docker-restart-smoke`, docs check, CodeRabbit,
  DeepScan, Go/Rust security scans, and Node audits), and merged it to `main`
  as `12a37339e9cff957d1b44413aa36160aed4e8d29`.
- GitHub automatically closed GH-43. GH-29 remains open as the parent rollout
  tracker.

## 2026-07-19 00:54 EEST - GH-41 network partition recovery start

- Confirmed `main` is synchronized with `origin/main` at `464a36a`.
- Confirmed no open PRs and latest `main` Go CI, Security Scan, and Pages
  deployment are green.
- Cleaned stale completed issues #8, #11, #13, #14, #20, and #21 with closure
  comments that preserve remaining rollout boundaries under GH-29/GH-7.
- Left #4, #7, and #29 open intentionally as parent/audit/rollout trackers.
- Opened GH-41 for the next Phase 1 rollout gap: network partitions, delayed
  peers, validator failure, and recovery without ledger divergence.
- Created branch `feature/GH-41-network-partition-recovery`.
- Added `TestMultiValidatorNetworkPartitionRecovery`, which starts a 3-of-4
  quorum without the fourth peer, starts the fourth validator isolated with no
  peers, commits a real `create-domain` transaction on the quorum side, then
  reconnects the isolated validator and verifies catch-up to the same app hash.
- The new harness also verifies all validator powers after recovery and exports
  every node state through `validateLedgerGenesis` to prove bank supply,
  treasury/stake escrow, DEX reserve, and LP invariant consistency.
- Local evidence:
  `TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test . -run
  TestMultiValidatorNetworkPartitionRecovery -count=1 -timeout=300s -v` PASS
  in 104.175s; all three gated process harnesses pass together in 392.147s;
  `go test ./...` PASS.
- Published PR #42, waited for green GitHub checks (`build-and-test`,
  `multi-validator-recovery`, `docker-restart-smoke`, docs check,
  CodeRabbit, DeepScan, Go/Rust security scans, and Node audits), and merged it
  to `main` as `8544943dd6fab483884392f1f04e83acbeb8f3f7`.
- GitHub automatically closed GH-41. GH-29 remains open as the parent rollout
  tracker.

## 2026-07-15 13:20 EEST - GH-39 validator lifecycle evidence

- Continued `feature/GH-39-validator-lifecycle` from the active rollout goal.
- Fixed Cosmos SDK v0.50 custom signer resolution for hand-written
  truedemocracy Msgs, including dynamic protobuf messages used during tx decode.
- Replaced the partial BaseApp router-only registry setup with
  `SetInterfaceRegistry` so tx decoding, event generation, gRPC, and message
  routing share the same address codecs and custom signer functions.
- Added delivered-transaction verification to the process harness using
  CometBFT `/tx` RPC, preventing `broadcast-mode sync` from hiding DeliverTx
  failures.
- Fixed process-harness account-number handling for offline signing and added
  explicit RPC/catch-up waits before replacement validator tx submission.
- Proved validator join/replacement with a gated six-node process harness:
  `TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test . -run
  TestMultiValidatorJoinReplacementLifecycle -count=1 -timeout=300s` PASS in
  117.638s.
- Re-ran focused validator lifecycle tests and full repository tests:
  `go test ./x/truedemocracy -run
  'TestValidatorJoinLeaveReplacementUpdates|TestBuildValidatorUpdates|TestRemoveValidator'
  -count=1` PASS; targeted root lifecycle/tx tests PASS; `go test ./...` PASS.
- Remaining boundary: public validator leave is still economically coupled to
  full stake withdrawal, so process-level join/replacement evidence is paired
  with Keeper/ABCI power-zero tombstone regression coverage for leave.
- Published [PR #40](https://github.com/NeaBouli/TrueRepublic/pull/40), waited
  for green GitHub checks (`build-and-test`, `multi-validator-recovery`,
  `docker-restart-smoke`, docs consistency, CodeRabbit, DeepScan, Go/Rust
  security scans, and Node audits), and merged it to `main` as
  `ad30d188c7956c28cff5bf53304bc04848ba569a`.

## 2026-07-11 11:50 EEST - Recovery initialization

- Created GitHub recovery epic #4.
- Fetched GitHub and discovered the old local checkout was 150 commits behind
  the actual v0.4 codebase.
- Preserved the old checkout and created clean branch/worktree
  `fix/GH-4-recovery-foundation` from `origin/main` (`d8545cf`).
- Verified main branch protection requires linear history and one approval.
- Reproduced ten consecutive weekly Security Scan failures on GitHub.
- Reproduced Go 1.26 build failure caused by `sonic@1.13.1`.
- Applied Go toolchain and fixable dependency updates; recheck running.
- Reproduced v0.4 client lockfile drift; repaired reproducible `npm ci`.
- Updated vulnerable v0.4 client dependencies to zero npm advisories.
- Fixed React lint issues and exact PNYX string/BigInt conversion.
- Added five v0.4 client regression tests; tests pass.
- Ran Rust workspace tests (26 PASS) and Clippy (PASS).
- Updated two fixable Rust vulnerabilities; cargo audit now exits successfully
  with six transitive dev-tooling warnings.
- Documentation consistency script passes, but public status claims require a
  full recovery recheck before approval.
- Added continuous agent bridge files at the repository root and under
  `docs/agent-bridge/`.
- Created GitHub recovery tickets #5 (security/toolchain), #6 (v0.4 client),
  #7 (consensus/security audit), and #8 (CI/docs/local reconciliation).
- Completed Go 1.26.3 build, `go test ./... -race -cover`, and `go vet ./...`.
- Completed v0.4 client lint, 5 Vitest tests, Vite 8 production build, and
  `npm audit` with zero vulnerabilities.
- Recorded the v0.4 client 1.68 MB main-chunk performance warning.
- Upgraded Go to toolchain 1.26.5 and patched current AWS EventStream/S3 and
  go-jose advisories; govulncheck reports no fixable reachable vulnerability.
- Reproduced legacy web/mobile wallet CI and security state: 68 and 51 npm
  advisories respectively; mobile has no tests.
- Updated README, `docs/status.json`, `docs/LIMITATIONS.md`, and GitHub Pages
  source with a visible recovery/non-production warning linked to GH-4.

## 2026-07-11 12:15 EEST - Foundation verification

- Re-ran the recovered Go stack with Go 1.26.5: build PASS, focused race tests
  PASS, and vet PASS.
- Confirmed coverage at root 5.8%, treasury 97.0%, DEX 34.2%, and governance
  53.5%.
- Confirmed `govulncheck` has no reachable finding with an available fix; four
  upstream findings without fixes remain documented for import-path reduction.
- Inspected the complete recovery diff and verified `git diff --check` passes.
- Confirmed GitHub CLI authentication and remote/default-branch targeting before
  publishing the first recovery milestone as a draft PR.
- Counted current passing tests from executable suites and corrected public
  status to 564 recovery-verified tests (533 Go, 26 Rust, five maintained-client),
  with 577 retained only as a historical total.
- Replaced stale public Go, Vite, CosmJS, bundle-size, and feature-completeness
  fields with the versions and audit states reproduced during recovery.

## 2026-07-11 12:30 EEST - First GitHub milestone

- Committed the recovery foundation in three scoped commits: toolchains/security,
  maintained client, and bridge/public status.
- Pushed `fix/GH-4-recovery-foundation` without modifying the protected `main`
  branch or the preserved legacy checkout.
- Opened draft PR #9 against `main`; GitHub checks are pending.
- PR inspection exposed that the React workflow covered only legacy `web-wallet`.
  Repointed it to canonical `client-web` with install, lint, test, build, and
  audit gates; added a blocking canonical-client audit to Security Scan while
  retaining non-blocking legacy audit visibility.
- Audited the 21M PNYX cap, denomination flow, governance/staking treasury,
  reward issuance, DEX custody, and custom genesis handling. Recorded six
  blocking ledger/supply failures in `CODEX_AUDIT.md`; no production-funds claim
  is acceptable until they are resolved.
- Split the ledger recovery into GitHub issues #11 (canonical denom/cap), #14
  (treasury/stake escrow), #13 (cap-checked rewards), #10 (DEX custody/LP/burn),
  and #12 (genesis/runtime invariants).

## 2026-07-11 12:55 EEST - GH-11 denomination and cap implementation

- Created stacked worktree/branch `fix/GH-11-pnyx-cap` from the PR #9 head.
- Added one canonical Go token package for `upnyx`, six decimals, 100,000 PNYX
  minimum stake, and the 21,000,000 PNYX / 21,000,000,000,000 upnyx cap.
- Added bank-genesis validation for cap boundaries and legacy display-denom
  balances plus canonical bank metadata injection before module init.
- Migrated Go modules, CosmWasm examples/testing helpers, maintained client, and
  operational documentation to `upnyx` base-unit semantics; legacy client and
  node-init metadata were aligned as well without changing their deprecated status.
- Go build/race/vet passes with 540 tests; token package coverage is 88.0%.
- Rust tests (26), Clippy, and cargo audit pass with the six already documented
  transitive dev-tooling warnings.
- Maintained client lint, six tests, build, and zero-advisory audit pass.
- Legacy web wallet still builds with warnings and passes 18 tests; legacy
  mobile still has no tests. Their known 68/51-advisory blockers are unchanged.
- Documentation consistency and diff checks pass at 572 recovery-verified tests.

## 2026-07-11 13:10 EEST - PR #9 GitHub verification

- Confirmed every current PR #9 check is green: Go, Rust, canonical client,
  docs consistency, Go vulnerability gate, Rust audit, canonical npm audit,
  and informational legacy npm audits.
- Closed GH-5 and GH-6 after both local and GitHub acceptance criteria passed.
- GH-7, GH-8, and ledger remediation tickets remain open.

## 2026-07-11 13:15 EEST - GH-11 stacked publication

- Rebased GH-11 onto the latest green PR #9 head.
- Pushed `fix/GH-11-pnyx-cap` and opened stacked draft PR #15 against
  `fix/GH-4-recovery-foundation` so its diff contains only the cap/denom block.
- GitHub checks for PR #15 are pending; GH-11 remains open.

## 2026-07-11 13:20 EEST - Preserved checkout reconciliation

- Reviewed every modified/untracked category in the 150-commit-divergent legacy
  checkout without changing it.
- Found no code suitable for wholesale merge: its isolated tokenomics work is
  superseded by GH-11, its C++/protobuf client targets the unsafe v0.1 protocol,
  and its bridge findings are already captured in the canonical audit/tickets.
- Kept the checkout intact pending final archive/hash comparison after recovery
  PRs land; recorded the decision in GH-8.

## 2026-07-11 20:09 EEST - PR #9 merge audit

- Re-inspected all 35 changed files and confirmed no Go consensus, ledger,
  genesis, DEX, or ZKP runtime source is modified in the foundation PR.
- Reproduced Go build, normal tests, race tests, vet, and the exact fixable
  govulncheck gate on Go 1.26.5.
- Reproduced 26 Rust tests, Clippy with warnings denied, and cargo audit; six
  transitive warnings remain documented through upstream/dev-tooling paths.
- Reproduced maintained-client `npm ci`, lint, five tests, production build,
  and npm audit with zero vulnerabilities; confirmed neither compromised Axios
  version is resolved.
- Verified the `golang:1.26.5-alpine` image tag exists. Local Docker is not
  installed, so added a GitHub Docker build job instead of claiming a build.
- Found and corrected six whitespace-only diff errors.
- Recorded the structured review in `PR9_AUDIT.md`. Merge remains conditional
  on refreshed GitHub CI and one independent GitHub approval.

## 2026-07-11 20:54 EEST - Docker glibc linkage fix

- The newly added Docker gate reproduced the image failure on both push and PR
  runs: Alpine/musl attempted to link wasmvm's default glibc shared library and
  failed on unresolved `GLIBC_*` symbols.
- Verified wasmvm v2.2.2's platform matrix: default Linux builds use the bundled
  glibc `.so`; musl requires an explicit static `muslc` build tag/library.
- Compared the later GH-21 container, whose Debian/glibc build and restart smoke
  tests are green, and selected the same minimal linkage model for GH-4.
- Switched builder/runtime to Bookworm, copied the architecture-specific
  `libwasmvm` shared object into `/usr/lib`, and ran `ldconfig`.
- Local Go build, documentation, YAML, and diff checks pass. Docker verification
  remains delegated to GitHub because Docker is unavailable on this workstation.

## 2026-07-11 21:25 EEST - PR #9 review remediation

- Confirmed both Debian/glibc Docker jobs pass and refreshed the PR #9 audit,
  bridge, queue, and handover state.
- Read all unresolved GitHub review threads and grouped 12 findings into CI
  token security, Node runtime compatibility, dependency security, Docker
  maintainability, bridge state, and public status.
- Disabled persisted checkout credentials in affected workflows and restricted
  the Security Scan token to read-only repository contents.
- Raised canonical `client-web` runtime/CI to Node 22 for CosmJS 0.39 while
  retaining Node 20 only for the explicitly deprecated legacy clients.
- Updated `golang.org/x/crypto` to v0.52.0 and the OpenTelemetry module family
  to v1.43.0, then tidied Go module metadata.
- Replaced the hard-coded wasmvm module-cache path with `go list -m` resolution.
- Corrected the public warning card, limitation version, roadmap date, and
  authoritative 564-test recovery total.
- Re-ran the complete local acceptance matrix: Go build/vet/race/coverage,
  Node-22 client install/lint/five tests/build/audit, 26 Rust tests plus
  Clippy/audit, documentation consistency, diff checks, and dynamic wasmvm
  library-path resolution all pass.
- `govulncheck` still reports only the four documented reachable findings with
  `Fixed in: N/A`; neither newly reported fixable dependency advisory remains.

## 2026-07-11 21:09 EEST - Docker recovery and final GH-11 audit

- Replaced the Alpine/musl node image path in PR #9 with a Debian/glibc builder
  and runtime, copied the architecture-specific wasmvm shared library, and ran
  `ldconfig` in the runtime image.
- Verified both GitHub Docker builds and all other PR #9 gates pass; converted
  PR #9 from draft to ready and requested the required independent review from
  `xxlfan72` and `ijuedt`.
- Audited every GH-11 denomination boundary and found a production validator
  tree stake that had been renamed to `upnyx` without base-unit scaling.
- Corrected the default stake from 0.1 PNYX to 100,000 PNYX and added a
  regression test across all seven generated nodes.
- Corrected Compose, node-init, environment, operator-documentation, and
  maintained-client gas prices so the migration preserves economic values.
- Normalized canonical bank metadata by removing conflicting legacy PNYX
  metadata while preserving unrelated asset metadata; expanded tests.
- Full Go build/race/vet, Rust test/Clippy/audit, maintained-client
  test/lint/build/audit, documentation consistency, and diff checks pass.
- The added validator-tree regression brings the recovery-verified total to
  573: 541 Go cases, 26 Rust tests, and six maintained-client tests.

## 2026-07-11 22:08 EEST - GH-14 bank-escrow audit remediation

- Rebased `fix/GH-14-bank-escrow` onto the fully green PR #15 head so PR #16
  remains an ordered, reviewable custody-only stack.
- Audited bank escrow, cache-context atomicity, signer claims, treasury payouts,
  validator stake settlement, slashing, CLI construction, and CosmWasm bindings.
- Found and fixed a high-severity contract regression: three CosmWasm encoders
  omitted the authenticated `Sender` required by hardened messages.
- Found and fixed a high-severity custody flaw: slashed stake was credited to an
  admin-withdrawable domain treasury. Penalties now burn exact module escrow.
- Added the minimum `burner` module permission plus exact-burn, burn-failure
  rollback, permission, missing-signer, and binding-validation regression tests.
- Updated the token/ledger audit to mark the GH-11 bank-genesis cap and GH-14
  custody slices remediated while retaining GH-13/GH-10/GH-12 blockers.
- Go build, vet, race, coverage, focused governance tests, and 557 full Go cases
  pass locally. Rust test/Clippy/audit, maintained-client install/lint/test/build/
  audit, documentation consistency, and diff checks also pass. The branch total
  is now 589 with 26 Rust and six maintained-client tests.
- Published PR #16 head `d3ae4cf`; both Go/Race and Docker runs, Rust, client,
  docs, DeepScan, and the manually triggered full CodeRabbit review completed.
- Accepted CodeRabbit's slash-atomicity hardening and three stale documentation
  findings. Rejected only its obsolete suggestion that refreshed PR #15 checks
  were pending because live GitHub evidence shows them green at `e0ff339`.

## 2026-07-11 23:02 EEST - GH-13 canonical issuance rebase and audit

- Rebasing retained only the GH-13 implementation commit on top of verified PR
  #16 head `fa693a8`; stale GH-13 documentation commits were intentionally
  skipped and reconstructed from current evidence.
- Audited canonical supply reads, minter/burner permissions, cap clipping,
  validator/domain allocation, interval snapshots, vote transfers, slashing,
  EndBlock ordering, and mint/burn failure paths against whitepaper equations
  4/5 and the 21,000,000 PNYX rule.
- Found and fixed a high-severity abstraction bypass: validator slashing now
  uses `token.IssuanceService.Burn` rather than calling the bank keeper directly.
- Found and fixed a high-severity partial-settlement boundary: staking and
  domain issuance now share one outer EndBlock cache, with a second-mint failure
  regression proving claims, timers, and snapshots roll back together.
- Added missing-bank, nil/negative supply, failed burn, and second-mint failure
  coverage. Canonical supply input is validated before cap calculations.
- Go build, vet, 567 cases, race, and coverage pass. Current coverage is root
  10.2%, token 93.5%, treasury 97.0%, DEX 34.2%, and governance 55.8%.
- Updated the structured audit, public DEX limitations, canonical reward docs,
  599-test status, bridge decisions, security notes, and PR #17 audit record.
- Hardened the node Dockerfile to derive `libwasmvm` from Docker's target
  architecture (`amd64`/`arm64`) instead of host `uname`, added runtime linkage
  validation and `libgcc-s1`, and injected the immutable GitHub commit as the
  binary version.
- Added `.dockerignore`; the excluded Rust target and installed client
  dependencies alone reduce the prior 1.6 GB local build context by more than
  1.5 GB and prevent local environment/key files from entering the context.
- Re-ran Rust Clippy and audit after all changes. Clippy passes; audit reports
  no vulnerability and the same six allowed transitive dev-tooling warnings.
  Maintained-client and documentation gates pass; GitHub must execute the
  actual Docker build because this workstation has no Docker engine.
- The first GitHub image execution proved wasmvm selection/linkage correct but
  exposed an older application startup panic: `legacytx.StdTx` was registered
  directly and then a second time by the auth codec. Removed the duplicate,
  unified application/CLI codec construction, exposed the injected build
  version through Cobra, and added the 567th Go regression test.
- Published `b738d70`. Both refreshed Docker builds now pass their linkage and
  CLI-start smoke check; both Go jobs (including race/coverage), docs, DeepScan,
  and the complete manually triggered security workflow are green.
- Thread-aware GitHub inspection reports zero unresolved review threads. The
  prior full CodeRabbit review completed without findings; the incremental
  startup-fix refresh was acknowledged but temporarily rate-limited.
- The accepted final 33-file review found five inline issues plus two additional
  findings. Made mock-bank mint/burn deltas cache-backed and extended the
  second-mint regression to prove supply/escrow rollback and parity.
- Initialized restored-domain payout snapshots at genesis and lazily backfilled
  pre-GH-13 state, preventing historical payout windfalls. Added genesis and
  upgrade-compatible lazy-backfill regressions.
- Added the container `--version` smoke, reconciled PNYX-cap decisions, DEX
  status, equation signatures, and branch-head handover wording. Full local Go
  build/vet, 569 cases, focused race, coverage, and docs consistency pass.
- Published `0e6cf38`, replied to every inline finding with implementation and
  test evidence, and recorded the two non-inline fixes in the PR conversation.
  Both Go jobs, both Docker builds, docs, DeepScan, CodeRabbit, and the manually
  triggered security workflow pass; all five review threads are resolved.
## 2026-07-12 03:34 EEST - GH-10 DEX custody audit and verification

- Rebased `fix/GH-10-dex-custody` onto final PR #17
  head `29fb228`, preserving both governance and DEX module burn permissions.
- Audited direct and two-hop DEX value paths: exact
  account/module transfers, reserve changes, intermediate PNYX routing, burn
  deltas, LP ownership, authority checks, slippage, and cached commit boundaries.
- Fixed a valid-denom LP key collision by replacing
  textual prefixes with length-prefixed keys; added `atom`/`atom:staked`
  conservation regression coverage.
- Added KV-backed failure regressions for create/add/
  remove/swap/burn paths, proving account, pool, LP, analytics, and supply
  rollback on every settlement boundary.
- Verified 578 Go cases plus race/vet/build/CLI smoke,
  26 Rust tests and audit, six maintained-client tests plus lint/build/audit,
  Go module integrity, docs consistency, and clean diff/secret checks. Docker
  remains delegated to GitHub because no local Docker CLI is installed.
- Published rebased PR #18 head `3234741`; updated Issue #10 acceptance state,
  recovery epic #4, PR metadata, BRIDGE, audit, README, status JSON, and GitHub
  Pages source.
- GitHub docs, DeepScan, Go build/vet/race/coverage, and the real Docker build
  pass. Manually dispatched Security Scan run `29156922464`; govulncheck, Rust
  audit, canonical npm audit, and both informational legacy audits all pass.
- Requested focused CodeRabbit review. The service reported its quota exhausted
  for 44 minutes, so substantive external review remains explicitly pending.

## 2026-07-12 04:32 EEST - GH-12 genesis and runtime-invariant audit

- Rebased GH-12's code commit onto final PR #18 and discarded three obsolete
  documentation commits for evidence-based regeneration.
- Preserved the CLI Amino panic fix, both issuance tests, cache-aware bank mocks,
  and GH-10's collision-free LP keys while adapting global LP export/orphan
  detection to the new length-prefixed format.
- Reproduced and fixed the prototype's critical hard-coded validator secret.
  Production defaults are empty; InitChain accepts actual CometBFT Ed25519
  public keys and creates exact cap-checked module stake, or rejects startup.
- Replaced silent InitGenesis skips with explicit failure and fail-closed JSON
  export behavior.
- Expanded full-app evidence from escrow-only to independent supply, escrow,
  reserve, and LP invariant halts. Added non-empty bank/treasury/stake/pool/LP
  export-import preservation plus over-cap, duplicate, negative, and unbacked
  rejection tests.
- Verified 615 Go cases, race, vet, build, coverage (root 66.1%, token 92.6%,
  treasury 97.0%, DEX 45.3%, governance 56.6%), 26 Rust tests/audit, six
  maintained-client tests/lint/build/audit, CLI smoke, and module integrity.
- Recorded the residual GH-21 blocker: `scripts/init-node.sh` still invokes the
  unavailable `x/staking` gentx flow and must not launch production nodes.
- Published rebased PR #19 head `9d521ce`; synchronized Issue #12, recovery
  epic #4, PR metadata, BRIDGE, audit, README, status JSON, and GitHub Pages
  source.
- GitHub Docs, DeepScan, Web, Mobile, Rust, Go build/vet/test, and both Docker
  builds pass. Manually dispatched Security Scan run `29158360390`; all five
  jobs pass. Requested CodeRabbit review; its independent result is pending.

## 2026-07-12 05:12 EEST - GH-20 ZKP/authentication audit

- Rebased GH-20's single code commit onto final PR #19 and discarded two stale
  documentation snapshots for evidence-based regeneration.
- Preserved chain/rating signal binding, rating-independent chain/proposal
  nullifiers, transaction-time setup removal, canonical commitment checks, and
  altered-rating/cross-chain regressions.
- Found and fixed missing active-nullifier export. Genesis now round-trips exact
  nullifier records/heights without resurrecting values cleared by Big Purge.
- Pinned genesis VK configuration by supported circuit ID, SHA-256, BN254 curve,
  four-public-input shape, canonical serialization, and no trailing bytes.
- Recomputed genesis identity Merkle roots and rejected malformed commitments,
  history, rating authentication fields, nullifiers, and proof public inputs.
- Found and disabled a legacy wallet path that generated random mock proof bytes
  and called `signAndBroadcast` while claiming anonymity. Both maintained and
  legacy web clients now fail closed and display non-submittable preview status.
- Verified 643 Go cases, race, vet, build, coverage (root 66.1%, token 92.6%,
  treasury 97.0%, DEX 45.3%, governance 58.9%), 26 Rust tests/audit, eight
  maintained-client tests/lint/build/audit, four focused legacy ZKP tests/build/
  audit, module integrity, and diff/secret/setup checks.
- Published rebased PR #22 head `6732276`; synchronized Issue #20, recovery
  epic #4, PR metadata, BRIDGE, audit, README, status JSON, and GitHub Pages.
- GitHub Docs, DeepScan, Web, Mobile, Rust, both Go/Docker runs, and manual
  Security Scan run `29159603247` pass. CodeRabbit is check-green but explicitly
  rate-limited and produced no substantive independent cryptographic review.
- Rebased the six audited GH-21 implementation commits without content drift
  onto final GH-20 bridge head `fac50a4`.
- Reproduced and fixed a release regression introduced by the lifecycle move:
  `truerepublicd --version` failed and `truerepublicd version` was blank because
  Cobra/Cosmos SDK application metadata was no longer wired.
- Verified 649 Go cases and coverage (root 64.3%, token 92.6%, treasury 97.0%,
  DEX 45.3%, governance 58.9%), a real native block/SIGINT/same-home restart/
  export flow under race, vet, CGO build, both version interfaces, entrypoint
  syntax, and diff checks.
- Prepared the local GH-21 audit/bridge/public-status evidence at 683 total
  recovery cases. Local Docker and ShellCheck are unavailable; refreshed
  GitHub Docker restart and security jobs remain mandatory before approval.
- Published audited GH-21 head `ec1ce17`; synchronized PR #23, Issue #21, and
  recovery epic #4 with the 649 Go / 683 total evidence and version regression.
- GitHub Go build/vet/race/coverage and Docker block/restart run `29170712626`,
  Docs, DeepScan, Web, Mobile, Rust, and all five manual Security Scan
  `29170832988` jobs pass. Independent multi-node operations review remains.
- Created isolated GH-8 checkout and rebased only its CI/docs commits onto
  final GH-21 `b59efa2`, discarding the obsolete link-only 636-case handoff.
- Combined Action-major upgrades with read-only, non-persisted checkout
  credentials, Node 22 for the maintained client, main-only push automation,
  PR/manual feature execution, and no duplicate branch/PR suites.
- Found the docs gate silently checked nonexistent `wiki-github/` paths while
  the real wiki claimed v0.1-alpha, 182 tests, Go 1.23, Testnet Ready, usable
  anonymity/mobile, and no high/critical issues. Replaced those claims with
  evidence-backed 683-case recovery status and explicit blockers.
- Corrected the recovery installation path to select the branch that actually
  provides GH-21 lifecycle commands; created real wiki current/testing pages
  and reconciled linked architecture/operator toolchain facts.
- Local workflow YAML, docs consistency/arithmetic, JSON, wiki target,
  stale-current-claim, and diff checks pass. GitHub Action execution is pending.
- Published rebased GH-8 PR #24 head `3964f4a`; synchronized PR metadata,
  Issue #8, and recovery epic #4 with the 683-case docs/wiki/CI audit findings.
- GitHub Go race/coverage + Docker restart `29171461365`, Rust `29171461357`,
  Web `29171461355`, Mobile `29171461342`, Docs `29171461348`, DeepScan,
  CodeRabbit, and all five Security Scan `29171476126` jobs pass.
- PR #25 remains draft/red against unrecovered main; no gate bypass is allowed.

## 2026-07-12 12:35 EEST - GH-12 review remediation and stack refresh

- Accepted both actionable PR #19 review findings. Because `CreateDomain` has
  no error return, the escrow-divergence test now reads the domain back and
  validates its treasury, so a failed fixture cannot masquerade as invariant
  coverage.
- Expanded the production-node limitation to require a real Ed25519 key bound
  to a positive-power CometBFT validator, sufficient exact bank-backed stake,
  and preservation of the 21,000,000 PNYX canonical supply cap.
- Focused registered-invariant regression, full Go tests, vet, and build pass.
  Published commit `eec91c7`, answered and resolved both PR #19 review threads,
  rebased/published PR #22 at `0c72ad0`, rebased/published PR #23 at `49938a3`,
  and published the propagated PR #24 stack head.
- PRs #19, #22, #23, and #24 are mergeable on exact consecutive bases with no
  unresolved review threads. Their refreshed standard checks pass, as do all
  five jobs in Security runs `29172007410`, `29172246257`, `29172246373`, and
  `29172246235` respectively.
- PR #9 remains technically green and mergeable but correctly requires one
  independent approval. PR #25 remains red against unrecovered `main`; neither
  gate is bypassed, and meaningful recovery work remains available.

## 2026-07-12 12:48 EEST - GH-26 operator init recovery

- Inventoried all open issues/PRs and found a remaining operator footgun:
  `scripts/init-node.sh` still invoked unavailable `x/staking` gentx commands,
  wrote a mnemonic capture file, and was recommended by public install docs.
- Opened Issue #26 and isolated `fix/GH-26-pod-init-script` from final PR #24.
- Replaced the wrapper with a single quoted `truerepublicd init` delegation;
  retained chain/moniker/home, minimum gas price, and Prometheus configuration.
- Added a regression proving the exact command boundary, absence of all legacy
  account/gentx actions and mnemonic files, and both configuration edits.
- Reconciled operator/public/security documentation and advanced the verified
  source of truth to 684 cases (650 Go + 26 Rust + 8 maintained client).
- Focused test, real compiled-daemon wrapper smoke, full Go suite, vet, shell
  syntax, documentation/JSON consistency, and diff checks pass. The real smoke
  proves one consensus validator, matching custom PoD identity, canonical
  `upnyx` bank supply, no mnemonic artifact, and configured gas/Prometheus.
- Publication and GitHub Go/Docker/security verification remain in progress.
- Published implementation/audit head `86ff1c8` and opened stacked draft PR
  #27 against final PR #24. Refreshed GitHub gates are running.
- GitHub Go race/coverage and Docker restart run `29172845624`, Docs
  `29172845627`, DeepScan, CodeRabbit, and all five Security Scan
  `29172846057` jobs pass. PR #27 is mergeable with zero review threads.
- Completion audit found GitHub Pages still built `main:/docs` at old head
  `d8545cf`. Changed only the Pages source to the fully green recovery branch
  `fix/GH-26-pod-init-script:/docs` and explicitly queued build `1090733247`.
- The build completed without error at `50b0d9a`. Live HTTP verification shows
  the recovery warning, non-production boundary, 21M maximum supply, and 684
  verified cases. PR #25 and branch protection were not bypassed.

## 2026-07-14 00:28 EEST - GH-32 multi-validator recovery implementation

- Reopened GH-29 after confirming PR #31 closed the roadmap handoff rather than
  the seven rollout phases; created child Issue #32 with explicit bounded scope.
- Extracted single-validator genesis binding into a reusable internal public-key
  set assembler while preserving `init` refusal to overwrite an existing set.
- Added four independently generated validator homes and keys, one identical
  exactly bank-backed PoD genesis, explicit persistent peers, common-height
  app-hash checks, one-validator failure, three-validator continued quorum,
  restart/catch-up, clean shutdown, export, and exported-ledger validation.
- The first run correctly failed because strict CometBFT address books reject
  loopback peers and every node inherited pprof port 6060. Restricted
  address-book/duplicate-IP relaxation to temporary localhost configs and
  disabled pprof only for harness processes; production defaults are unchanged.
- Targeted genesis/binder tests pass. The full normal Go suite passes with 651
  cases. The separately gated four-validator harness passes twice locally, most
  recently in 55.84 seconds.
- Added a dedicated `multi-validator-recovery` Go CI job, operator runbook,
  685-case public source-of-truth update, and exact remaining-scope warnings.
- Full Go race/coverage passes: root 64.9%, token 92.6%, treasury 97.0%, DEX
  45.3%, and governance 58.9%; build and vet also pass.

## 2026-07-14 02:05 EEST - GH-32 review remediation

- PR #33's initial GitHub matrix passed, including the dedicated four-validator
  job, Go/Docker, documentation, security, Node audit, and static-analysis gates.
- Accepted review findings for an ambiguous public roadmap bullet, missing
  subprocess/request cancellation, and a recovery assertion that only checked
  the last survivor-produced height.
- The harness now threads the test context through every child process and HTTP
  request, selects a target two blocks after the stopped process restarts, then
  requires all four nodes to reach it and agree on that post-restart app hash.
- Rejected the proposed `actions/checkout@v7` edit: the official action has no
  v7 release, and a runtime-major migration is separate from this bounded
  network-recovery ticket. The new job remains on the repository's v5 baseline.
- Genesis/binder regressions pass and the hardened real four-validator harness
  passes in 68.90 seconds, its third successful local run.

## 2026-07-14 04:03 EEST - GH-32 merged and rollout tracker advanced

- Published review-remediation commit `7dd0fbb`; all four CodeRabbit threads
  are answered or automatically recognized as addressed and are resolved.
- The first final-head attempt failed four jobs before checkout because GitHub
  could not resolve action download metadata and returned `Service Unavailable`.
  Job-level logs confirmed no project command ran, so no code was changed.
- Reran only failed jobs. Go run `29253316692` passes build/race/coverage,
  Docker restart, and multi-validator recovery; Security run `29253316707`
  passes Go vulnerability, Rust, maintained-client, and both legacy audits.
  Docs, DeepScan, and CodeRabbit also pass.
- Squash-merged PR #33 as `9d68a6f`; GitHub automatically closed Issue #32.
  GH-29 remains open and its first Phase 1 checkbox now links the evidence.

## 2026-07-14 04:16 EEST - GH-32 final main and Pages proof

- Squash-merged bridge closure PR #34 as `2851759`; no pull requests remain
  open, GH-32 is closed with all nine criteria checked, and GH-29 remains open
  with only the completed multi-validator item checked.
- Final `main` Security run `29261145077` passes all five jobs. The previously
  documented upstream-only govulncheck annotations remain non-blocking.
- GitHub Pages build `1093339877` completed from `main:/docs` at exact commit
  `2851759`. Live HTTP verification confirms the recovery/non-production
  warning, 21M cap, 685 verified cases, and four-validator recovery evidence.

## 2026-07-15 20:10 EEST - GH-37 Codex subagent role configuration

- Opened GH-37 to track the project-scoped Codex agent role split.
- Added `.codex/config.toml` with one-level subagent nesting and a six-thread
  cap for predictable delegation.
- Added `.codex/agents/spark-worker.toml` as a narrow
  `gpt-5.3-codex-spark` worker for small bounded patches, file search, and
  focused checks.
- Documented that the main Codex agent keeps architecture, risk, integration,
  final verification, GitHub updates, merges, and Bridge responsibility.
- Verified both `.codex` TOML files parse with Python `tomllib`; `git diff
  --check` and `bash scripts/check-consistency.sh` pass.
- PR #38 opened for GH-37. GitHub Docs Consistency, Security Scan, and DeepScan
  pass; CodeRabbit remained pending without comments during the final merge
  decision.

## 2026-07-22 23:04 EEST - GH-48 fast post-merge audit

- Confirmed local `main` and `origin/main` were identical at `6a308c5`, with no
  open pull requests and only GH-4, GH-7, and GH-29 open before this task.
- Confirmed the latest scheduled Security Scan at `6a308c5`, the latest Pages
  build, and the latest code-bearing Go CI at `63b76bf` are green.
- Opened GH-48 after finding live documentation that still described the
  already merged PR #9 through PR #27 recovery sequence as drafts/unmerged.
- Corrected the contributor guide, root audit, DEX guide, ZKP limitation,
  security queue, and live project state without altering historical logs.
- Recorded three residual warnings: remaining GH-29 rollout evidence, root Go
  wildcard discovery in installed frontend dependencies, and the 1,678.30 kB
  maintained-client main chunk.
- Maintained-client install/lint/8 tests/build/audit, Rust format/Clippy/26
  tests, isolated Go build/vet/655 tests, docs consistency, shell syntax, JSON,
  workflow YAML, and diff checks pass. `cargo audit` exits cleanly with the six
  already documented allowed transitive warnings. Docker and `govulncheck` are
  unavailable locally, so current green GitHub security/Docker evidence remains
  required on the final published head.

## 2026-07-22 23:14 EEST - GH-50 GO-2026-5970 remediation

- PR #49's current vulnerability database found reachable GO-2026-5970 through
  Unicode normalization used by the ZKP dependency path. The affected indirect
  module was `golang.org/x/text` v0.37.0; the scanner reports v0.39.0 as fixed.
- Opened GH-50 and updated `golang.org/x/text` to v0.39.0. Module resolution
  also advances `golang.org/x/sync` to v0.21.0 and records the already directly
  imported `cosmossdk.io/x/tx` in the direct require block.
- A current local `govulncheck` no longer reports GO-2026-5970. Four reachable
  upstream findings remain with `Fixed in: N/A`; the existing Security Scan
  gate continues to fail only when a reachable fixable version exists.
- Exact CI-filter reproduction passes: no reachable finding with an available
  fix remains. Go build, vet, and the full 655-case suite pass after the update;
  the root package completes in 64.499 seconds.
- Accepted CodeRabbit's audit-tally finding: the document contains 17 distinct
  `[PASS]` findings, so the root audit and live project state now report
  0 FAIL / 3 WARN / 17 PASS.
- PR #49 final head `2bd5efd` passes Go build/race/coverage (6m49s), combined
  multi-validator recovery (5m42s), Docker restart (3m09s), Go vulnerability,
  Rust audit, maintained and legacy Node audits, docs consistency, DeepScan,
  and CodeRabbit. The one valid review thread is resolved.
- Squash-merged PR #49 as `7dbde858d0d3d5410f22d16a1a3bac614325d925`;
  GitHub closed GH-48 and GH-50. At the merge point, only GH-4, GH-7, and
  GH-29 remained open.
- Opened GH-51 for the remaining medium package-selection finding and linked
  the complete audit/remediation evidence from GH-29 and GH-4.

## 2026-07-22 11:24 UTC - GH-51 root Go package isolation

- Reproduced the issue: root `go list ./...` discovers
  `truerepublic/client-web/node_modules/flatted/golang/pkg/flatted` after the
  maintained-client dependencies are installed.
- Added `scripts/go-packages.sh`, which derives explicit package directories
  only from Git-managed, non-ignored Go source and defensively excludes
  `node_modules` and `vendor` trees. Makefile verification, contributor/agent
  guidance, and GitHub Go CI now use that same selector.
- Added `scripts/test-go-packages.sh`; its ignored dependency-source probe
  leaves the selected five repository packages unchanged.
- Ran `npm ci` concurrently with selector verification, build, vet, and the
  complete race/coverage suite. All pass; root coverage is 65.9%, and npm
  reports zero vulnerabilities. The normal 655-case suite also passes.
- The combined consensus recovery, trusted state-sync, and sanitized
  backup/restore/export/import harnesses pass in 265.381s. Compose/Docker cannot
  run locally because the Docker executable is unavailable, so final-head
  GitHub Docker evidence remains required before merge.
- CodeRabbit reported four valid findings: normalize the Bridge timestamp,
  remove the resolved medium-priority duplicate, ignore missing indexed source,
  and return to the repository root in the wiki test sequence. All were fixed
  in `6a668ee`, answered, and resolved.
- PR #52 final head passes all 11 checks: Go build/vet/race/coverage (7m11s),
  multi-validator recovery (5m44s), Docker restart (3m20s), Go vulnerability,
  Rust audit, maintained and legacy Node audits, docs consistency, DeepScan,
  and CodeRabbit. PR #52 was squash-merged as `ae7105a`; GH-51 closed.

## 2026-07-23 12:40 EEST - GH-56 local implementation and audit

- Implemented authenticated atomic validator consensus-key rotation with a
  signed expected-old-key precondition, permanent old-key revocation, and one
  deterministic H to H+2 CometBFT transition while preserving operator claims.
- Separated bootstrap operator authority from consensus identity, rejected
  self/cross/historical consensus-derived authorities and module accounts, and
  materialized missing public operator base accounts without generating secret
  material.
- Added delayed old-key evidence attribution, one-shot/deferred removal
  handling, genesis/export/import validation, CLI/API support, Docker/operator
  plumbing, and a dedicated rotation/compromise-response runbook.
- Added the real five-process rotation harness. Four active validators rotate
  to a pre-synchronized stopped-key replacement; the old signer stays offline,
  activation occurs at H+2, claims and app hash converge, old-key reuse fails,
  and export/re-import succeeds. Final local run PASS in 168.12 seconds.
- Full normal Go suite, focused authority/transition regressions, repository Go
  vet, docs consistency, shell syntax, and diff checks pass. Verified recovery
  arithmetic is now 717 cases: 683 Go, 26 Rust, and eight maintained-client.
- Final adversarial review found no P0. Residual rollout warnings are isolated
  in GH-59 (ABCI++ evidence/slashing), GH-60 (inactive validator round trip),
  and GH-61 (legacy authority migration). See `GH56_AUDIT.md`.
- PR #62's first combined process run exposed a CI-only harness setup defect:
  isolated test selection had not configured the Cosmos SDK global prefix, so
  legacy smoke helpers emitted `cosmos1...` operators that the hardened init
  correctly rejected. `smokeOperatorAddress` now encodes `truerepublic`
  explicitly, independent of test order. The previously first-failing
  four-validator recovery harness passes in isolation in 99.33 seconds.

## 2026-07-23 03:06 EEST - GH-56 GitHub closure

- PR #62 final head `239cc6f` passed all enforceable GitHub checks: Go
  build/vet/race/coverage in 6m44s, the combined multi-validator matrix in
  9m39s, Docker restart in 3m20s, Docs, Go/Rust/Node security, and DeepScan.
- CodeRabbit was rate-limited and produced no content review; the independent
  adversarial final review recorded in `GH56_AUDIT.md` found no P0 and no
  additional P2 finding.
- Squash-merged PR #62 as `80ab6741423049d6faa3bfc8d2671feebb314134`.
  GitHub closed GH-56; GH-59, GH-60, and GH-61 retain the explicit residual
  rollout boundaries.

## 2026-07-24 00:25 EEST - GH-59 final local merge gate

- Completed deterministic ABCI++ evidence and decided-last-commit ingestion,
  exact equivocation/downtime penalties, consensus-key history, tombstones,
  replay cursors, rolling liveness, and evidence-window exit custody.
- Added fail-closed genesis/export/import relations across validators, signing
  state, revocations, rotations, infractions, and pending exits. Mature payout
  now removes signing state atomically; partial withdrawal remains disabled
  until generalized slashable unbonding exists.
- A real four-process harness proves 100-block downtime, restart/catch-up,
  cryptographically valid duplicate-vote evidence, exact burns, validator
  power-zero transitions, app-hash convergence, and ledger-valid export.
- Fresh local checks pass: docs consistency at 726 tests and the 21M cap,
  package selection, all five Go packages (root/process package 158.108s),
  build, vet, diff checks, and the complete race/coverage gate. The latter
  reports no race and 68.5% root / 61.8% `x/truedemocracy` coverage.
- The final independent read-only merge-gate review reports 0 P0, 0 P1, and
  0 P2. Detailed invariants and evidence are recorded in `GH59_AUDIT.md`.
- Published commit `c8a56f8` in PR #65; the closure evidence follows below.

## 2026-07-24 00:50 EEST - GH-59 GitHub closure

- PR #65 final head `313b327` passed all 11 checks: Go build/vet/race/coverage
  in 7m05s, the combined multi-validator process matrix in 11m46s, Docker
  restart in 3m15s, docs, Go/Rust/Node security, DeepScan, and CodeRabbit.
- Zero unresolved review threads remained. CodeRabbit was rate-limited and
  produced no content review; the recorded independent final audit found
  0 P0 / 0 P1 / 0 P2.
- Squash-merged PR #65 as `934a042017e860852e945f1f82ca2da571fcff86`.
  GitHub closed GH-59. GH-60 and GH-61 remain the bounded consensus-state
  follow-ups under the still-open GH-29 rollout tracker.

## 2026-07-26 19:40 EEST - GH-60 inactive validator genesis start

- Reconstructed the task from clean `origin/main` at `996ae05`, GH-60, the
  global and local rules, and the complete current Bridge state. GH-59 remains
  closed and is not being repeated.
- GH-60 requires exact retained inactive validator state across export/import:
  custody, complete domain membership, jail, missed blocks, power semantics,
  rotations, revocations, signing/infraction state, pending exits, ledger
  backing, and cap parity. Resurrection and historical-key reuse must fail
  closed.
- Created `feature/GH-60-inactive-validator-genesis`; no unrelated local user
  changes were present.
- Adopted the Sol/Kimi role split. Kimi receives one bounded, secret-free local
  implementation slice; Sol retains architecture, security, integration,
  external writes, complete testing, review, PR/merge, and Bridge ownership.
- Verification has not run yet for GH-60. Next: inspect exact genesis paths,
  start the Kimi slice, review its diff, add remaining integration/process
  evidence, and run the complete relevant gate chain.

## 2026-07-26 20:50 EEST - GH-60 local merge gate

- Kimi K3 implemented the bounded genesis core: new exports explicitly carry
  active/inactive classification, all domains, and stored power while valid
  legacy records keep their previous single-domain, stake-derived semantics.
- Sol reviewed every changed line and added process-level recovery to the real
  four-validator slashing harness. The drill proves exact inactive export and
  re-import after downtime and double-sign penalties, with ledger parity and
  no stake/domain/jail/liveness/power drift. PASS in 441.52s.
- The first real drill exposed only an existing harness-timeout flake: the
  loaded host could not advance the 100-block liveness window within 120s. The
  liveness phase now has 300s headroom; the repeated full drill passed.
- Kimi's independent read-only review found one P2 fail-closed gap: an explicit
  inactive claim could be unjailed with positive power when its domains were
  empty. Sol now rejects every inactive unjailed positive-power record and
  added a dedicated regression. The refreshed review reports 0 P0 / 0 P1 /
  0 P2.
- Final exact local CI chain passes: package selector; all five Go packages;
  build; vet; race/coverage (68.5% root, 62.2% truedemocracy); docs consistency
  at 733 cases and 21,000,000 PNYX; and `git diff --check`.
- Added `GH60_AUDIT.md`; updated current Bridge/project/TODO state and public
  recovery totals from 726 to 733 (699 Go, 26 Rust, eight maintained-client).
- GH-60 remains open. Next: publish the branch, open/update the PR and GH-29,
  require final-head GitHub CI/review, remediate any finding, then merge and
  record closure. No production rollout approval is claimed.

## 2026-07-26 21:15 EEST - PR #67 maintained-client Security Scan remediation

- Published GH-60 commit `cfe065e` and opened PR #67. Updated GH-60 and GH-29
  with the local process/race/review evidence.
- GitHub's maintained-client npm audit failed twice after successful clean
  installs. The retiring quick endpoint returned HTTP 400; querying the
  official bulk advisory dataset exposed new High findings in
  `brace-expansion <=5.0.7` and `postcss <=8.5.17`.
- Updated direct PostCSS to `^8.5.23`; pinned `brace-expansion` to patched
  `5.0.8`; unified the old ESLint transitive minimatch path at `10.2.5` so the
  patched brace-expansion CommonJS contract remains compatible.
- Fresh `npm ci`, lint, eight tests, TypeScript/Vite production build, and
  `npm audit --audit-level=high` pass. Two Moderate React Router advisories
  remain visible and require a separately reviewed breaking v7 migration;
  they do not fail the existing High gate.
- Next: publish the lockfile remediation and require every final-head GitHub
  check and review to refresh before merge.

## 2026-07-26 21:17 EEST - PR #67 review remediation

- Triaged every CodeRabbit finding. Synchronized GH-60 local-complete versus
  publication-pending records, refreshed the rollout snapshot date, documented
  the uncached Go command and GH-60 recovery gate, pinned validation-test error
  reasons, and scoped the panic assertion only around `InitGenesis`.
- Softened an inaccurate inactive-state comment: stored power zero is itself
  an explicit non-consensus state and remains valid even for an unjailed,
  sufficiently staked member.
- Rejected the suggestion to derive export activity from domain/stake
  eligibility. The runtime consensus classification is exactly unjailed plus
  positive stored power; reclassifying malformed positive-power state as
  inactive would still fail import validation and would hide corruption rather
  than preserve it exactly.
- Focused governance tests, docs consistency at 733 cases / 21M PNYX, and diff
  checks pass. Next: run the complete final-diff Go gate, publish, answer and
  resolve review threads, then require refreshed GitHub checks before merge.

## 2026-07-26 21:25 EEST - PR #67 final review-diff gate

- All five uncached Go packages pass, followed by Vet, CGO build, complete race
  and coverage, docs consistency, and diff checks.
- Coverage is 68.5% root, 92.6% token, 97.0% treasury, 45.3% DEX, and 62.2%
  governance.
- No local blocker remains. Next: publish the review remediation, resolve its
  GitHub threads with evidence, and wait for the refreshed final-head gates.

## 2026-07-26 21:40 EEST - GH-60 closure

- Published final review remediation as `0cad6df`, answered and resolved all
  four CodeRabbit threads, and preserved fail-closed runtime activity
  classification rather than masking corrupt state during export.
- All 12 final-head GitHub checks passed: Go build/vet/race/coverage in 7m14s,
  combined real multi-validator recovery in 11m55s, Docker restart in 3m33s,
  maintained-client build, docs, Go/Rust/Node security, DeepScan, and
  CodeRabbit.
- The normal merge remained blocked only by the formal second-review branch
  rule. With explicit owner authorization and Kimi's independent 0 P0 / 0 P1 /
  0 P2 review recorded, Sol used the scoped admin squash merge. PR #67 merged
  as `c5a3d38`; GitHub closed GH-60.
- GH-29, the roadmap, Bridge, audit, security notes, TODO, and project state
  are synchronized in the closure branch. GH-61 is the next bounded
  consensus-state task; no deployment or production rollout is approved.

## 2026-07-26 21:56 EEST - GH-61 legacy authority migration start

- Reconstructed GH-61 from clean `origin/main` at `8c0c555`, the closed GH-60
  chain, GH-56's coupled-authority boundary, GH-29, and the 72-item rollout
  roadmap (11 complete, 61 open).
- The migration must be a deterministic halt/export/re-import ceremony, not a
  compatible binary replacement. It must authenticate every replacement
  operator independently from all consensus keys and preserve exact validator,
  consensus, domain, escrow, auth-account, revocation, liveness, infraction,
  pending-exit, and supply state.
- Created `feature/GH-61-legacy-authority-migration`; no unrelated user changes
  are present.
- Sol owns architecture, security, integration, external writes, complete
  verification, and closure. Kimi K3 receives one bounded secret-free
  architecture/development block and may not delegate.
- No GH-61 verification has run. Next: define the canonical plan/proof and
  atomic fail-closed rewrite, then add unit, genesis-integration, rollback, and
  real multi-validator migration evidence.

## 2026-07-26 22:09 EEST - GH-61 architecture security review

- Kimi's read-only review mapped the complete rewrite graph and confirmed that
  an offline atomic halt/export/transform/re-import boundary is smaller and
  more honest than claiming nonexistent `x/gov` or `x/upgrade` support.
- Sol rejected the review's proposed legacy-consent and >2/3 validator
  signatures because they were verified with the legacy consensus keys. That
  would reintroduce exactly the consensus-derived authority GH-61 must remove.
- No implementation was accepted or written. The remaining architecture
  question is whether pre-GH-56 state contains any independently controlled
  governance anchor. If not, GH-61 must introduce and authenticate one
  explicitly or keep the reviewed fresh-genesis path as the only supported
  transition.
- Next: run a bounded adversarial authorization review, record the decision,
  then delegate only the approved migration core.

## 2026-07-26 22:20 EEST - GH-61 authorization-anchor decision

- Kimi completed a bounded read-only adversarial audit of every plausible
  pre-GH-56 authorization anchor. No independently controlled signer with
  protocol-granted authority over validator identity exists.
- `x/gov` and `x/upgrade` are absent; module accounts cannot sign; the legacy
  validator operators and bootstrap domain admin are consensus-key-derived;
  domain permission/ZKP keys and arbitrary user accounts carry no chain
  migration authority.
- Sol therefore rejects retroactive “governance approval.” The only honest
  one-time recovery boundary is a reviewed, state-preserving fresh-genesis
  ceremony bound to the exact halt height and source app hash, with
  proof-of-possession from every fresh independent replacement operator.
- A new consensus-independent governance commitment may be introduced in the
  migrated genesis for future decisions, but cannot be presented as authority
  for the initial rewrite.
- No implementation or test run occurred during this review. Next: synchronize
  the finding to GH-61, narrow its acceptance boundary, and implement the
  canonical descriptor/proof verifier plus deterministic transformer and
  recovery drill.

## 2026-07-26 22:38 EEST - GH-61 canonical descriptor and fresh-key proofs

- Added `migration/legacy_authority.go` and its test suite. Descriptor v1 binds
  both chain IDs, halt height, source app hash, transform ID, and the complete
  mapping; each fresh ed25519/secp256k1 operator key signs the same
  domain-separated canonical bytes.
- Verification rejects malformed/mixed-prefix addresses, wrong key derivation
  or proof, descriptor mutation/replay, duplicates, unsorted mappings,
  old/new cross-coupling, malformed consensus-key input, and replacement
  addresses derived from any supplied legacy consensus key.
- Kimi implemented only this bounded package and its tests. Its first test run
  caught a stale-index sorting comparator; Kimi corrected it before handoff.
- Sol reviewed the complete diff, added stable JSON tags and exact account-size
  checks at canonical-encoding time, then independently reran the gate.
- `gofmt -l migration` and `git diff --check` are clean; `go vet ./migration`
  passes; uncached package tests pass in 0.946s; the race run passes in 2.578s.
- Remaining boundary: the future transformer must supply the complete exported
  consensus-key set and reconcile every typed reference. No CLI, transformer,
  integration drill, repository-wide gate, commit, push, or PR exists yet.

## 2026-07-26 22:59 EEST - GH-61 transformer threat-review refinement

- Enforced distinct source/target chain IDs in descriptor verification and
  added the negative regression, preventing legacy account-transaction replay
  on the reviewed fresh-genesis chain.
- Independent field review identified two additional fail-closed requirements:
  nonempty CosmWasm contract state needs a separately reviewed contract-aware
  migration adapter, and auth account numbers must remain globally unique so
  SDK import cannot silently renumber signer identities.
- The generic transformer must repack only mapped `BaseAccount` records with the
  fresh public key and unchanged account number/sequence, rejecting mapped
  module or unknown account types.
- A second Kimi block completed the detailed typed-transform design but was
  stopped before file writes after exceeding its bounded planning window. No
  partial transformer code was accepted.
- Uncached migration tests pass in 1.334s after the chain-ID change and
  `git diff --check` is clean. The transformer, CLI, integration/process drill,
  repo-wide verification, commit, push, PR, and merge remain open.

## 2026-07-27 00:09 EEST - GH-61 truedemocracy transform core

- Added the state-derived transform core and focused regressions. It inventories
  active, historical, revoked, pending-rotation, and pending-removal consensus
  keys directly from the export before verifying any descriptor proof.
- Exact mapping completeness, replacement collision, invalid proof, malformed
  export, atomic failure, input/descriptor immutability, post-transform genesis
  validation, and old-literal removal are covered.
- Spark's first draft was not accepted as written: Sol's gate found duplicate
  test symbols, incorrect secp256k1 address derivation, a short app hash, and an
  invalid fixture. Sol replaced the faulty tests and hardened SDK-prefix and
  domain-admin decoding before rerunning the complete package gate.
- `gofmt -l migration` and `git diff --check` are clean; `go vet ./migration`
  passes; uncached tests pass in 3.726s; race tests pass in 3.617s.
- Auth/Bank/DEX reconciliation, the explicit Wasm gate, outer genesis binding,
  CLI, process drills, repo-wide verification, and GitHub publication remain
  open.

## 2026-07-27 00:24 EEST - GH-61 full app-state reconciliation

- Added `migration/app_state.go` and focused application-genesis regressions.
  The atomic transformer now binds outer chain ID and halt successor height,
  repacks mapped BaseAccounts with fresh keys while preserving account
  identity counters, moves Bank and DEX ownership, and validates each typed
  module before returning output.
- Duplicate auth addresses/account numbers, unsupported mapped account types,
  pre-existing replacement ownership, invalid source/post state, stale legacy
  address literals, and descriptor proof failures all fail without output or
  input mutation.
- Generic migration now refuses nonempty CosmWasm code/contract state. Unknown
  modules remain preserved and are included in the stale-address gate.
- The earlier independent review supplied the critical Auth/Wasm and
  custom-module boundary requirements. Sol owned implementation, complete diff
  review, fixture corrections, and final verification for this slice.
- `gofmt -l migration` and `git diff --check` are clean; `go vet ./migration`
  passes; uncached migration tests pass in 2.177s; race tests pass in 6.398s.
- Remaining: trusted-header app-hash CLI verification, atomic file output,
  operator docs, export/transform/re-import and rollback drills, full repository
  gates, GitHub publication, commit, push, PR, review, and merge.

## 2026-07-27 00:54 EEST - GH-61 canonical offline CLI and re-import

- Added `truerepublicd migration legacy-authority` with strict descriptor
  decoding, explicit trusted source-app-hash comparison, bounded regular-file
  reads, input/output identity checks, no-overwrite private atomic output, and
  content-free success reporting.
- Added the missing App/Comet validator reconciliation: a populated consensus
  set must exactly match active application consensus keys and power; empty
  export validators remain valid for application `InitChain` reconstruction.
  Fresh target genesis rejects an embedded prior app hash.
- The CLI performs the existing exact cross-module ledger validation before
  creating output. The new end-to-end command regression transforms a valid
  legacy-coupled full genesis and reimports it into a fresh application.
- Kimi supplied a detailed bounded file-safety design but produced no write
  diff before Sol stopped the overlong planning run. Sol implemented, reviewed,
  hardened, and verified the complete slice.
- `gofmt -l` and `git diff --check` are clean; `go vet ./migration .` passes;
  uncached package/root tests pass in 3.951s/141.937s; focused race tests pass
  in 13.865s/13.296s; the canonical roundtrip passes again in 12.041s; daemon
  build/help smoke exposes the new command.
- Remaining: operator runbook, real pinned pre-GH-56 four-validator
  halt/export/transform/import and rollback drill, full repository gates,
  independent final security review, GitHub publication, commit, push, PR,
  review, and merge.

## 2026-07-27 01:34 EEST - GH-61 historical migration and rollback drill

- Added a gated four-validator harness that builds pinned pre-GH-56 revision
  `0e51a05b008f395e3f7391358e117f0817d4eb39`, runs a real coupled-authority
  source chain, establishes a stable no-quorum halt/app hash, exports and
  transforms through the current CLI, starts a distinct-ID target chain,
  verifies app-hash/key/power convergence and export/reimport, stops all target
  signers, and resumes the untouched historical source as rollback.
- The first target rehearsal discovered a real initial-height defect:
  genesis-restored consensus keys used runtime H+2 activation and rejected the
  first decided commit when initial height was greater than one. Split genesis
  activation (actual initial height) from runtime activation (H+2) and added a
  focused height 0/1/4 regression without weakening rotation semantics.
- Final live drill passes in 249.13s. The focused activation suite passes in
  3.588s; post-documentation truedemocracy/migration/root checks pass in
  2.537s/5.105s/3.688s; docs consistency, vet, formatting, and diff checks pass.
- Added the operator migration runbook and synchronized the operator index,
  upgrade boundary, rollout roadmap, and limitations. The procedure explicitly
  records that fresh-key proofs prove possession, not retroactive governance,
  and that source/target signers may never run concurrently.
- Remaining: full repository and security matrices, independent final
  adversarial review and remediation, commit/push/PR, final-head GitHub
  workflows/review, merge, and issue/roadmap closure.

## 2026-07-28 20:49 EEST - GH-61 exact export binding and local final gate

- Confirmed GH61-SB-001 in a disposable pre-fix CLI reproduction: a valid
  export could be changed after fresh operators signed the descriptor, and the
  changed governance state was transformed and re-imported because only the
  source header app hash was bound.
- Added `source_genesis_sha256` to descriptor v1 canonical signing bytes and
  require exactly 32 bytes in encoding and verification. The transform hashes
  the exact raw byte slice before JSON parsing and compares it in constant time,
  so direct package callers and the CLI share one fail-closed boundary.
- Updated descriptor, transform, full-application, CLI, and historical-process
  fixtures to sign the exact artifact. Raw-mutating negative fixtures recompute
  and re-sign their commitment so they still exercise the intended deeper
  Auth/Bank/DEX/Wasm/Comet and mapping controls.
- Added the project-local `AGENTS.md` required by the refreshed Bridge workflow
  and synchronized the operator runbook with exact-byte freeze/hash/sign
  ordering. Reformatting the export after signing is now explicitly forbidden.
- Kimi performed an independent secret-free read-only review and made no file
  changes. It found no P0/P1/P2 issue. Sol accepted and fixed one P3 test-
  attribution observation by sorting an added mapping before testing signature
  invalidation; the separate canonical-string sweep boundary remains recorded
  as P3 without expanding this task.
- Focused migration/root regressions pass (`migration` 3.318s, root 6.837s).
  `make verify` passes: repository selector, build, vet, and Race/Coverage for
  root 193.441s/69.4%, migration 18.935s/82.5%, token 5.397s/92.6%, treasury
  9.157s/97.0%, DEX 23.969s/45.3%, and truedemocracy 172.168s/62.2%.
  Documentation consistency, `gofmt -l`, and `git diff --check` pass.
- The exact real historical
  `TestMultiValidatorLegacyAuthorityMigrationRollback` drill passes again in
  174.82s (178.949s package): four-validator halt/export, exact-bound
  transform, distinct target convergence, export/re-import, complete target
  signer shutdown, and untouched source rollback all remain proven.
- GH61-SB-002 remains open but out of scope: a measured 64 MiB padded artifact
  reached about 1.75 GiB maximum resident memory and 33.51s in the focused
  transform path. No push, PR, merge, deployment, or follow-up remediation was
  authorized or performed in this task.

## 2026-07-29 02:06 EEST - GH-61 publication started

- Froze and staged exactly the 26 documented GH-61 files. The staged diff and
  secret-pattern scan were clean; no unrelated user changes were present.
- Created commit `432156fe9090b1f2711461d0a135715b572fd8bb` and pushed
  `feature/GH-61-legacy-authority-migration` to `origin` without force.
- Opened draft PR #69 against exact `main` base
  `8c0c55574389210162db59c2c1103bb2d99e2691`. The description closes GH-61
  only on merge and records the honest fresh-genesis boundary, real local
  evidence, unsupported production claims, and GH61-SB-002 residual risk.
- Initial GitHub checks started; DeepScan is green and CodeRabbit skipped the
  draft as configured. The PR remains draft until required final-head checks
  pass, after which review must be triggered and resolved before merge.

## 2026-07-29 02:09 EEST - PR #69 current-database gRPC finding

- Final-head Security Scan run 30361942572 failed only `go-vuln`: reachable
  GO-2026-6061 affects `google.golang.org/grpc@v1.79.3`, with `v1.82.1`
  reported as fixed. The four other reachable upstream findings remain `N/A`.
- This is a vulnerability-database/dependency gate change, not a GH-61
  transform regression. The bounded remediation is the scanner-named gRPC
  update plus unavoidable module metadata and full vulnerability, Race, and
  network-process verification before another push.

## 2026-07-29 02:29 EEST - GO-2026-6061 remediation locally verified

- Upgraded direct `google.golang.org/grpc` from `v1.79.3` to fixed `v1.82.1`.
  The required xDS, Genproto, Envoy validation, and GCP detector support graph
  followed; `go mod tidy` also classified already directly imported
  `cosmossdk.io/core` correctly. No application source changed.
- Current `govulncheck` removes GO-2026-6061 and shows zero reachable finding
  with an available fix. Four existing reachable `N/A` findings remain
  visible and tracked.
- `go mod verify`, docs consistency, diff check, package selector, build, vet,
  and full Race/Coverage PASS. Root/migration coverage remains 69.4%/82.5%.
- The complete eight-scenario real process matrix passes in 940.823s under the
  new transport graph, including GH-61 migration/rollback, recovery,
  state-sync, backup/restore, binary rollback, cold failover, key rotation,
  and slashing.

## 2026-07-29 03:08 EEST - PR #69 final-head CI green; review recovery

- Exact head `a28f5654390f58feecfd86afeeb4d7f74e204341` passed docs, Go,
  full multi-validator recovery, Docker restart, Go/Rust/Node security, and
  DeepScan checks; DeepScan reported zero new issues.
- CodeRabbit remained pending for about ten hours without a review or inline
  thread. Its explicit review command was acknowledged but did not replace the
  stale run; a draft/ready transition also left the same pending context.
- Recorded the external-review recovery state append-only. This
  documentation-only commit is intended to produce a fresh final head and bot
  event. No implementation/dependency behavior changes and no review bypass,
  merge, deployment, or production action were made.

## 2026-07-29 03:14 EEST - PR #69 review findings accepted for remediation

- CodeRabbit's full review of `a28f565` completed with six unresolved inline
  threads and seven additional review-body findings. The findings cover
  runbook atomicity/halt integrity, genesis consistency, deprecated SDK use,
  test diagnostics/coverage, mapping-sort efficiency, shared key traversal,
  and CI timeout margin.
- Accepted all findings as in-scope review remediation; the repository-wide
  docstring percentage remains informational.
- Assigned the bounded secret-free Go/test/CI block to Kimi K3. Sol retained
  documentation, architecture/security decisions, diff review, complete local
  and GitHub validation, thread resolution, external writes, and merge.

## 2026-07-29 03:33 EEST - Review remediation integration correction

- Kimi completed the bounded Go/test/CI patch; Sol reviewed the full diff.
  Focused tests/vet and `make verify` passed, raising migration coverage to
  85.0%.
- The first full process-matrix attempt failed in the historical GH-61 drill:
  a real halted running-chain export had an empty Comet validator list and four
  active truedemocracy validators. This proved CodeRabbit's proposed rejection
  incompatible with the supported Cosmos SDK export shape.
- Stopped the remaining matrix, restored canonical empty-consensus-export
  acceptance, added a direct regression plus code/runbook rationale, and kept
  strict reconciliation for non-empty consensus lists. A fresh complete matrix
  is required before publication.

## 2026-07-29 03:53 EEST - Corrected full process matrix passes

- Corrected migration tests and docs/diff gates passed.
- All eight real process scenarios passed in 1169.685s. The GH-61 historical
  migration/rollback slice passed in 196.40s; consensus recovery 119.13s; state
  sync 145.57s; backup/restore 90.99s; binary upgrade/rollback 223.52s; cold
  failover 87.03s; key rotation 88.48s; slashing 215.28s.
- The observed total leaves only about 30 seconds under the former 1200-second
  limit, validating the reviewed 1500-second test and 30-minute job budgets.
- The empty-consensus review suggestion is rejected with direct process
  evidence; every other review finding remains implemented pending final-head
  publication and thread resolution.

## 2026-07-29 03:58 EEST - Final review-remediation local gate passes

- Final `make verify` passed package selection, build, vet, and Race/Coverage:
  root 69.4%, migration 84.6%, token 92.6%, treasury 97.0%, DEX 45.3%, and
  truedemocracy 62.2%.
- Together with the corrected 8/8 process matrix and docs checks, the final
  local remediation diff is ready for staged review and publication.

## 2026-07-29 04:18 EEST - GH-61 merged and closed

- PR #69 final head `33895dca627139d541e74236afb63d2b9fb5bff2`
  passed all 11 GitHub checks. The complete process matrix passed in 13m38s;
  Go build/race/coverage, Docker restart, docs, Go/Rust/Node security,
  DeepScan, and CodeRabbit also passed.
- All original and final CodeRabbit threads were answered with evidence and
  resolved; zero unresolved threads remained before merge.
- PR #69 squash-merged to `main` as
  `264ab7c817414e3081149c61d8b5b6fb0ce5e368` and automatically closed
  GH-61.
- Started the documentation-only closure sync from that exact clean main commit
  to align Bridge, Project State, TODO, Road to Rollout, Limitations, Security
  Notes, and GH-29.

## 2026-07-29 04:24 EEST - Public recovery totals reconciled

- Reproduced 819 Go cases on merged main via package-scoped `go test -json`:
  root 62, migration 82, token 12, treasury 36, DEX 116, truedemocracy 511.
- Updated the authoritative public total from the stale GH-60 baseline
  733/699 to 853 total: 819 Go, 26 Rust, and eight maintained-client cases.
- Synchronized status JSON, README, agent guide, Pages source, wiki status,
  module arithmetic, Project State, and final root/migration/governance
  coverage. Separately gated process harnesses remain excluded from the count.

## 2026-07-29 04:35 EEST - PR #70 independent closure review remediated

- Kimi K3 performed a bounded read-only review and made no repository changes.
  It independently verified the PR #69 merge and issue state, all 11 final-head
  checks, zero unresolved review threads, exact 853/819 arithmetic and
  per-module Go counts, coverage evidence, documentation-only scope, and
  retained non-production boundaries.
- Review found no P0/P1, code, security, or production-claim issue. Corrected
  its sole material P2 (stale 733/699 roadmap baseline) and minor P3s (roadmap
  date and GH-61 omission from the wiki harness list).
- CodeRabbit did not rerun after its successful draft-skip result despite the
  ready-for-review/manual request; the independent Kimi review is retained as
  fallback evidence. Final merge still requires Sol diff review, green
  final-head checks, and zero unresolved threads.

## 2026-07-29 12:24 EEST - GH-71 locally verified and independently reviewed

- Implemented deterministic seed, sentry, validator, RPC/API, and private node
  policies with canonical peer validation, validator-initiated sentry
  isolation, listener/CORS/unsafe/metrics controls, fail-closed startup, safe
  Docker/nginx defaults, and an operator runbook.
- Kimi K3 implemented the bounded policy core and later completed an
  independent read-only review. Sol reviewed every diff, corrected topology,
  path-safety, P2P, metrics, Docker, documentation, and test gaps, and retained
  all external/GitHub authority.
- `make verify` passes on the final tree. Coverage is root 69.5%, migration
  84.6%, network policy 95.5%, token 92.6%, treasury 97.0%, DEX 45.3%, and
  governance 62.2%.
- Reproduced 952 Go cases: root 69, migration 82, network policy 126, token 12,
  treasury 36, DEX 116, governance 511. Together with 26 Rust and eight client
  cases, the public branch total is 986.
- Seven process scenarios passed in the combined run; concurrent load exhausted
  its 1500-second global limit during slashing. The isolated slashing rerun
  passed in 303.98 seconds. Exact final-head combined evidence remains a GitHub
  merge gate.
- Local Docker is unavailable. CI now runs the complete Compose stack and
  requires loopback proxy RPC, wallet routing, healthy node metrics, and
  Grafana health before merge.
- Structured audit: `GH71_AUDIT.md`. No P0/P1 or scope blocker remains.

## 2026-07-29 12:52 EEST - GH-71 merged and closed

- PR #72 exact final head `83ffaad86274efdf3d2bb54dfff6ce8a625d5dee`
  passed Go build/race/coverage (6m03s), the complete recovery matrix (14m29s),
  Docker/Compose restart and routing smoke (7m04s), docs, Go/Rust/Node security
  scans, and DeepScan with zero new findings.
- CodeRabbit remained in `Review in progress` for about ten hours without
  producing a review, finding, or thread. The stale external-bot condition was
  recorded on the PR; zero unresolved review threads existed. Independent Kimi
  review had already found no P0/P1, and every lower-severity observation was
  remediated before the exact final-head CI run.
- Under the repository owner's standing merge authorization, Sol used an admin
  squash merge only to bypass the stale bot status. PR #72 merged to `main` as
  `9c369ac0d589f749e055af33a03f5f4981020101`; GH-71 closed automatically.
- No server, firewall, DNS, production topology, deployment, key, credential,
  mainnet, or public-network action was performed. GH-29 remains the rollout
  tracker.

## 2026-07-29 17:45 EEST - GH-74 locally verified

- Added dependency-free `truerepublicd healthcheck live|ready` commands.
  Liveness checks only local RPC responsiveness; readiness additionally
  requires a positive committed height and `catching_up=false`.
- Hardened probes to literal-loopback plain HTTP, positive timeout at most ten
  seconds, 64 KiB responses, no userinfo/query/fragment/non-root path, no
  redirects or environment proxy, strict JSON-RPC 2.0, and stable
  operator-safe errors.
- Docker now uses liveness for its restart signal. GitHub container and
  Compose gates require liveness, readiness, and post-restart height progress.
  Operator documentation includes distinct Kubernetes exec probes.
- Kimi K3 implemented the bounded probe package/CLI/focused tests. Sol reviewed
  the complete write diff and corrected redirect/proxy handling, JSON-RPC
  version validation, an environment-sensitive port test, and race-safety.
- Local evidence: focused race tests PASS twice; focused root/config tests
  PASS; `go vet ./healthcheck` and `go vet .` PASS; `make verify` PASS.
  Package-scoped JSON evidence records 1,009 Go cases: root 71, healthcheck 55,
  migration 82, network policy 126, token 12, treasury 36, DEX 116, and
  governance 511. Public total: 1,043 with 26 Rust and eight maintained-client
  cases.
- Coverage: root 69.5%, healthcheck 97.2%, migration 84.6%, network policy
  95.5%, token 92.6%, treasury 97.0%, DEX 45.3%, and governance 62.2%.
- Docker is unavailable locally. Exact final-head GitHub image/Compose runtime
  evidence, independent review, and merge remain hard gates.
- Structured audit: `GH74_AUDIT.md`; current result 0 FAIL / 1 WARN / 12 PASS.
- Final independent Kimi review found no P0/P1/P2. Sol remediated its
  actionable P3 by validating the separately read block height before
  publishing or comparing the restart baseline. The duplicated timeout wording
  and standard released-port test idiom remain non-blocking P3 observations.
- PR #75 final-head technical gates passed before external review:
  build/race/coverage 7m25s, Docker/Compose 7m27s, complete recovery matrix
  13m58s, docs, Go/Rust/Node security scans, and DeepScan. CodeRabbit then
  opened six actionable threads; all six were accepted for minimal remediation:
  key-compromise runbook boundary, migration wording, fail-closed deployment
  curl, accurate Docker restart semantics, context-aware test listener, and
  complete-number documentation matching.
- No production, server, firewall, DNS, deployment, real topology, secret,
  credential, key, mainnet, or public-network action was performed.

## 2026-07-29 19:02 EEST - GH-74 merged and closed

- PR #75 exact remediation head
  `e02104452fe2ddd201045c0ea0bb9651c053c417` passed build/vet/race/coverage
  in 6m39s, the complete recovery matrix in 13m54s, Docker/Compose liveness,
  readiness, persistent restart, proxy, wallet, metrics, and Grafana in 7m27s,
  plus docs, Go/Rust/Node security, and DeepScan.
- CodeRabbit produced six actionable threads. All six were remediated in
  `e021044`, answered with evidence, and resolved. Its incremental follow-up
  was rate-limited but returned PASS with no new unresolved thread. Independent
  Kimi review recorded 0 P0/P1/P2.
- Owner-authorized admin squash merge bypassed only the repository's
  self-approval requirement; no failed technical or unresolved review gate was
  bypassed.
- PR #75 merged to `main` as `468a43a6cf4a3ed6cc76524a2de212b4fcf7052a`;
  GH-74 closed automatically. GH-29 remains open for later rollout phases.
- Correction to the earlier local milestone wording: Docker `HEALTHCHECK`
  records unhealthy state; `restart: unless-stopped` does not restart an
  unhealthy main process without explicit external automation.
- No production, server, firewall, DNS, deployment, real topology, secret,
  credential, key, mainnet, or public-network action was performed.

## 2026-07-30 05:28 EEST - GH-77 secret-safe structured logging started

- Opened GH-77 under rollout tracker GH-29 and created
  `feature/GH-77-structured-logging` from clean, synchronized `main` at
  `2435335`.
- Bounded the task to repository code, tests, operator documentation, and CI
  evidence. Consensus/application state and production infrastructure remain
  outside scope.
- The central SDK/CometBFT logger construction is the integration boundary.
  The supported node-operation path will use JSON while sanitization preserves
  safe operational metadata and redacts credentials, mnemonic/private-key
  material, raw transactions, proofs, signatures, and private payloads.
- Initial GitHub state: GH-77 is open; GH-4, GH-7, and GH-29 are the only
  other open issues; no pull request is open.
- Next: delegate the bounded logging core to Kimi K3, review its write diff,
  integrate process/policy tests and operator guidance, then run the complete
  relevant verification chain.

## 2026-07-30 06:17 EEST - GH-77 implementation and adversarial review complete

- Added the central structured/redacting logger boundary, JSON-only supported
  node start, raw trace-store refusal, explicit script/image/Compose defaults,
  container JSONL CI evidence, adversarial unit/process/policy tests, operator
  guidance, and `GH77_AUDIT.md`.
- The bounded Kimi implementation attempt returned design analysis but no
  writable diff and was stopped; Sol implemented and integrated the work.
  Kimi then independently reviewed the current tree read-only and reported
  PASS with no P0/P1/must-fix P2.
- Focused evidence is green: vet; normal and race observability tests; start,
  script, and network policy tests; and a 59.78s compiled daemon
  start/stop/restart test covering fail-closed plain/trace-store inputs,
  persistent height advancement, export compatibility, and JSON
  `level`/`message` on every normal-operation log line.
- Docker is unavailable locally. Exact image/Compose evidence and all complete
  repository/security/recovery gates remain required on the published head.
- No production log, collector, server, deployment, credential, key, mainnet,
  consensus/state, or public-network action was performed.

## 2026-07-30 06:42 EEST - GH-77 complete local gates pass

- `make verify` passed build, vet, race, and coverage across the authoritative
  nine Go packages; observability coverage is 77.3%.
- The complete eight-scenario multi-validator recovery matrix passed in
  1161.044s. Documentation consistency, diff/shell/YAML checks, Rust
  format/clippy/build/tests, Cargo audit, and active web-client
  lint/8 tests/build/high-severity audit also passed.
- Docker and `govulncheck` are unavailable locally and remain mandatory
  exact-head GitHub gates. Cargo audit emitted six allowed transitive warnings.
  The unchanged legacy mobile audit's 51 dependency findings remain
  out-of-scope debt under its existing informational workflow.
- GH-77 local implementation TODO is complete; publication, exact-head CI,
  review remediation, merge, and parent/public-status synchronization remain.
- No production log, collector, server, deployment, credential, key, mainnet,
  consensus/state, or public-network action was performed.

## 2026-07-30 06:44 EEST - GH-77 implementation published as PR #78

- Pushed reviewed implementation commit `ed4fb0700d415990eb79dd8e6f1c8b0104327165`
  to `feature/GH-77-structured-logging` and opened draft PR #78 against `main`
  with `Closes #77` and parent GH-29.
- Initial state is mergeable. Docs/recovery and Security/Go/Docker jobs entered
  the queue; CodeRabbit and DeepScan initially reported success.
- This append-only coordination update is the only post-publication local
  change. Its follow-up commit must become the exact head before CI/review
  evidence is accepted.
- No merge or production/server/deployment/public-network action was performed.

## 2026-07-30 06:54 EEST - GH-77 review remediation started

- Exact head `b97c527` reached 10 green checks; only recovery remained running.
  Go build/race/coverage passed in 7m00s and Docker/Compose with the new
  post-restart structured JSONL assertion passed in 7m09s.
- CodeRabbit opened two threads. Its `json.Number`/`fmt.Stringer` ordering
  finding is valid and will receive a minimal code/test fix. Its future-date
  finding is invalid because the records explicitly use EEST and the verified
  local clock was `2026-07-30 06:54:04 EEST +0300`; GitHub's July 29 display is
  UTC.
- Refreshed focused/full local checks and exact-head GitHub evidence remain
  required after the correction.

## 2026-07-30 06:59 EEST - GH-77 review remediation locally verified

- Reordered `json.Number` before `fmt.Stringer` and added a sanitizer-boundary
  regression. The underlying SDK logger serializes `json.Number` as a string,
  so the test accurately verifies sanitizer preservation without making a false
  end-to-end numeric JSON promise.
- Focused normal/race/vet, consistency, diff checks, and refreshed `make verify`
  all pass; observability coverage is now 77.8%.
- No code change is appropriate for the timestamp thread: the records are
  explicitly EEST and the local clock evidence proves they were written after
  the documented events, despite GitHub displaying the prior UTC date.
- Next: publish, respond/resolve with evidence, and require all refreshed
  exact-head checks.

## 2026-07-30 07:16 EEST - GH-77 implementation merged

- PR #78 exact head `fd8378c` passed all 11 checks: Go 7m16s, recovery 14m10s,
  Docker/Compose with post-restart JSONL 6m57s, docs, Go/Rust/Node security,
  DeepScan, and CodeRabbit.
- Both CodeRabbit threads are resolved. The bot confirmed the `json.Number`
  remediation and withdrew its UTC/EEST finding. Kimi full and incremental
  reviews found no P0/P1/P2.
- Owner-authorized admin squash bypassed only formal self-approval; PR #78
  merged as `133fb3b` and GH-77 closed automatically.
- Created `docs/GH-77-closure` for documentation/public-status/parent-tracker
  synchronization only. No runtime code or production action is included.

## 2026-07-30 - GH-77 closure documentation locally verified

- Fresh package-scoped JSON output records 1,058 Go cases: root 79,
  healthcheck 55, migration 82, network policy 126, observability 41, token 12,
  treasury 36, DEX 116, and governance 511. Public total: 1,092 with 26 Rust
  and eight maintained-client cases.
- Synchronized all public status/count surfaces, Phase-6 roadmap, limitations,
  machine state, wiki, agent state/notes/TODO, and finalized GH77_AUDIT at
  0 FAIL / 1 WARN / 16 PASS.
- Documentation consistency, JSON parsing, audit arithmetic, and diff hygiene
  pass locally. Runtime code is unchanged from merged implementation PR #78.
- Next: publish and verify the documentation-only closure PR, merge it, update
  GH-29, and synchronize clean local main.

## 2026-07-30 - GH-77 closed and synchronized

- Documentation-only PR #79 exact head `3d3eb5a` passed all eight applicable
  docs/security/static checks and merged as `69dee43`.
- GH-29 now marks GH-71 network policy, GH-74 liveness/readiness, and GH-77
  secret-safe structured logs complete while remaining open for unfinished
  rollout work. No pull request remains open.
- Public and durable state agrees on 1,092 cases and a final GH-77 audit of
  0 FAIL / 1 WARN / 16 PASS.
- GH-77 is Done / Target Stop. Next Phase-6 work is broader node/application
  metrics and operator evidence, not production deployment.
- No production, server, collector, deployment, credential, key, mainnet,
  consensus/state, or public-network action was performed.

## 2026-07-30 09:50 EEST - GH-80 node/application metrics baseline started

- Opened GH-80 under rollout tracker GH-29 and created
  `feature/GH-80-metrics-baseline` from clean, synchronized `main` at
  `ed9c971`.
- Existing evidence covers a loopback-only CometBFT Prometheus target and
  verifies target health, but application telemetry is disabled and runtime
  evidence does not require Cosmos SDK, Go/process, TrueRepublic application,
  or invariant-success metric families.
- Scope is a private two-source baseline: retain CometBFT consensus/peer/block
  metrics, opt in to Cosmos SDK Prometheus telemetry through the existing
  loopback API surface, add bounded public-chain-safe TrueRepublic metrics, and
  prove required families and progression through repository tests and
  Docker/Compose CI.
- Dashboards, alerts, objectives, escalation ownership, capacity, retention,
  and production collector/topology deployment remain separate rollout gates.
- No production, server, collector, deployment, credential, key, mainnet,
  consensus/state, or public-network action is authorized.
- Next: delegate one bounded secret-free implementation block to Kimi K3 while
  Sol owns the metric contract, network/security boundary, integration,
  complete verification, GitHub writes, and closure.

## 2026-07-30 10:20 EEST - GH-80 local implementation and audit complete

- Kimi K3 implemented the bounded label-free application collector core,
  collision-safe singleton registration, app EndBlock integration, invariant
  coupling guard, and focused tests. Sol reviewed the complete write diff and
  integrated opt-in/opt-out configuration, private two-source scraping,
  native process/restart/disabled evidence, Compose CI assertions, operator
  guidance, and the structured GH80 audit.
- Five custom families cover last successful application height, completed
  application blocks in the process, last successful invariant-cycle height,
  canonical PNYX supply, and exact fixed-cap headroom. No custom label is
  derived from user, address, transaction, request, peer, hostname, service,
  or deployment input.
- Focused build/vet/unit/race/root/process/wrapper/policy/YAML/shell/docs/diff
  checks pass. The initial full root run exposed Sol's concurrent nonportable
  BSD `sed` expression; replacing it with portable range expressions fixed the
  real failure, and the wrapper plus repeated full root suite pass.
- Audit result is 0 FAIL / 3 WARN / 10 PASS. Generic SDK query-path
  cardinality stays private and 60-second retained; EndBlock height is not a
  committed-height claim; exact-head Docker/Compose remains mandatory because
  Docker is unavailable locally.
- Next: independent full-diff review, any remediation, complete local
  `make verify` and recovery matrix, then a reviewed draft PR.

## 2026-07-30 11:22 EEST - GH-80 review findings remediated

- Kimi's independent full-diff review reproduced a policy-string mismatch, the
  Cosmos SDK gateway's raw disabled-route `501 Not Implemented` behavior, and
  the 60-second expiry of the one-shot `truerepublic_server_info` series.
- Sol fixed the policy guard, pinned the verified 501 fail-closed lifecycle,
  removed the expiring SDK series from the durable Compose contract, blocked
  exact `/api/metrics` at nginx with 404, and added `nginx/**` workflow
  triggers.
- Fresh policy/coupling tests and the compiled daemon
  start/restart/progression/disabled lifecycle pass. This supersedes the
  earlier intermediate-tree `404-disable` and complete-root PASS wording.
- The eight-scenario recovery matrix and remaining complete repository gates
  are still running; GH-80 is not Done and has not been published.

## 2026-07-30 12:04 EEST - GH-80 complete local gates pass

- Kimi full and incremental reviews have zero open P0/P1. Sol removed the
  expiring SDK startup series from process assertions and added a portable,
  fail-loud `[telemetry]` postcondition with backup preservation.
- `make verify` passes the exact package selector, build, vet, race, and
  coverage; root is 69.1% covered. All eight required recovery scenarios pass
  in 1155.6 seconds.
- Rust format/clippy/build, 26 tests, audit, and the maintained client-web
  install/lint/eight tests/build/high-audit pass.
- Legacy web/mobile wallet audits remain informationally red with inherited
  high/critical dependency advisories; their existing workflows use
  `|| true`, and GH-80 does not alter those dependencies.
- Docker is unavailable locally. GH-80 is ready for branch publication but
  remains In Progress until exact-head GitHub Docker/Compose and all other
  required checks pass.

## 2026-07-30 12:15 EEST - GH-80 exact-head Compose interpolation fix

- Published implementation commit `ea6adc6` in draft PR #81. All completed
  non-Docker checks are green.
- Docker reached the metrics-contract step but Compose rejected its missing
  required `GRAFANA_PASSWORD` before metrics were read.
- Added the same CI-only Grafana password and bootstrap operator environment to
  the contract step. YAML, static policy, consistency, and diff checks pass.
- Push/rerun and exact-head Docker/recovery evidence remain pending.

## 2026-07-30 12:24 EEST - GH-80 metrics contract made fail-loud

- The environment fix passed both target/proxy gates. The contract still exited
  before restart on a silent family/value assertion.
- Added explicit missing-family diagnostics and public value logging without
  changing any acceptance rule. YAML, policy, and diff checks pass.
- Next exact-head run will identify the precise assertion for focused repair.

## 2026-07-30 12:39 EEST - GH-80 zero-peer contract corrected

- Exact-head diagnostics identified `cometbft_p2p_peers` as the sole missing
  family after every preceding Docker/Compose gate passed.
- CometBFT does not instantiate its peer gauge before the first peer event, so
  the singleton CI topology legitimately omits it.
- CI now requires the present `cometbft_mempool_size` family. The production
  peer dashboard and documented alert coalesce the absent series to zero.
- Workflow YAML, dashboard JSON, focused network policy, documentation
  consistency, and diff checks pass locally.
- Reconciled GitHub's stale in-progress status against the completed runner
  log: all eight recovery scenarios pass in 666.955s; cleanup also completed.

## 2026-07-30 13:10 EEST - GH-80 CodeRabbit review fixes started

- Exact head `24e00eb` is fully green across GitHub checks.
- Verified all three actionable review threads: exact metric samples, bounded
  init subprocess execution, and descendant metrics-path denial.
- Applying the minimal fixes with focused tests before publication.
- All focused checks plus full `make verify` now pass; root coverage is 69.1%.
  Nginx syntax/runtime remains an exact-head Docker requirement.

## 2026-07-30 13:37 EEST - GH-80 implementation merged

- Exact head `4b880b4` passed all GitHub gates; three actionable CodeRabbit
  threads were fixed, answered, and resolved.
- Owner-authorized admin squash bypassed only the formal self-approval rule.
  PR #81 merged as `e629374` and GH-80 closed.
- Fresh merged-main JSON output records 1,070 Go / 1,104 total cases.
- Started documentation-only branch `docs/GH-80-metrics-closure`; merged-main
  Security, Pages, and graph jobs pass while post-merge Go CI continues.

## 2026-07-30 13:49 EEST - GH-80 closure locally verified

- Post-merge main Go CI is fully green.
- Kimi independently reproduced 1,070 Go / 1,104 total cases, coverage,
  PR/job evidence, and audit arithmetic with no P0/P1.
- Remediated its one P2 by correcting the stale dashboard-rendering claim in
  `docs/DEPLOYMENT.md`.
- Documentation consistency, JSON, audit arithmetic, and diff checks pass.

## 2026-07-30 21:56 EEST - GH-29 public rollout status locally verified

- Updated the GitHub Pages landing source with the canonical GH-29 tracker
  progress: 10/59 full checklist (17%), 10/51 phase work (20%), and 3/7
  Phase 6 operations work (43%).
- Replaced the stale active-reverification banner with the evidence-backed
  recovery baseline and unchanged non-production boundary. The public status
  strip now shows the 21M cap, 1,104 verified cases, rollout progress, green CI
  and security state, and v0.4.0 recovery baseline.
- Corrected Phase 1 to mark validator-key custody, single-signer failover,
  authenticated rotation, permanent revocation, and deterministic network
  policy as merged while leaving consensus-breaking migration recovery open.
- Added the rollout counts to `docs/status.json` and extended the repository
  consistency gate so landing-page counts and rounded percentages cannot drift
  silently.
- PASS: `./scripts/check-consistency.sh`, shell syntax, JSON parse, and
  `git diff --check`. Browser DOM and visual checks pass at the default
  1280-pixel desktop viewport and a 390-pixel mobile viewport; the mobile page
  has no horizontal overflow.
- No runtime code, production infrastructure, deployment, consensus, token,
  wallet, key, mainnet, or public-network action was performed.

## 2026-07-30 21:58 EEST - GH-29 GitHub Page update published as PR #83

- Published implementation commit `e8251b2` on
  `docs/GH-29-github-page-status` and opened documentation-only draft PR #83
  against `main`.
- PR scope and body link the canonical GH-29 counts, local consistency and
  responsive-browser evidence, and the unchanged non-production boundary.
- This append-only publication handoff becomes the replacement exact head.
  Pages, security, docs/static checks, review status, merge, and deployed-page
  verification remain pending.

## 2026-07-30 22:08 EEST - GH-29 PR #83 review findings remediated

- Clarified that the final exit gates are included in the 59-item total but
  remain open; the 10 completed items no longer read as if they include those
  gates.
- The docs consistency gate now rejects missing, nonnumeric, fractional, and
  negative rollout counts plus zero/nonpositive totals before any percentage
  division.
- PASS: complete docs consistency, shell syntax, status JSON, diff hygiene,
  positive integer fixtures, and rejection fixtures for negative, fractional,
  string, null, and zero-total inputs.
- Both post-merge PR #83 threads are technically addressed. Publication,
  exact-head checks, thread replies/resolution, merge, and deployed-page
  verification remain.

## 2026-07-30 22:09 EEST - GH-29 review follow-up published as PR #84

- Published `323abad` on `fix/GH-29-page-review-followup` and opened draft
  PR #84 against merged `main`.
- Replied to both unresolved PR #83 threads with the follow-up PR and concrete
  correction evidence. Threads remain open until the fix is merged and
  verified.
- This append-only handoff becomes the replacement exact head. Fresh docs,
  security, static/review checks and merge remain required.

## 2026-07-30 22:23 EEST - Phase 6 observability operations started

- Reconciled clean local `main` with exact `origin/main` `844c498`; no open PR
  exists and the latest main Pages and Security workflows pass.
- Inspected Prometheus, Grafana, Compose, CI, application metrics, operator
  documentation, rollout tracker, TODO, and security boundaries.
- Confirmed the next bounded gap: the dashboard is stored in an API-wrapper
  shape unsuitable for Grafana file provisioning; CI does not prove the
  dashboard, panel expressions, loaded rules, or alert behavior; no rule file,
  service objectives, or escalation ownership is shipped.
- Chosen boundary: repository-only dashboards, rules, objectives, role-based
  ownership, tests, CI, and docs. External paging and production deployment
  remain explicitly excluded.
- Next: create the GitHub child issue and issue branch, assign a bounded
  secret-free implementation block to Kimi K3, and retain architecture,
  security, integration, GitHub, and final verification with Sol.

## 2026-07-30 22:25 EEST - GH-85 ticket and branch created

- Created GH-85 with explicit acceptance, verification, risk, and
  non-production boundaries.
- Created `feature/GH-85-observability-operations` from exact main `844c498`.
- The task will treat dashboard provisioning, panel queries, loaded rules, and
  selected alert transitions as runtime/test contracts rather than
  documentation-only claims.
- Next: publish the initial Bridge handoff, delegate the bounded dashboard/rule
  implementation, and build the independent repository/CI validation path.

## 2026-07-30 22:49 EEST - GH-85 focused implementation checks pass

- Replaced the invalid Grafana API wrapper with a native immutable dashboard
  and stable provisioned datasource; added reviewed CometBFT, application,
  invariant, supply/headroom, and runtime panels.
- Added ten explicit Prometheus alerts, synthetic promtool transitions,
  recovery/testnet objectives, role-based escalation, bounded response
  guidance, Compose integration, repository tests, and GitHub runtime
  dashboard/rule/query evidence.
- The bounded Kimi implementation invocation produced useful independent
  design analysis but remained in planning and emitted no file write before
  being stopped. Sol authored the resulting diff; a separate Kimi read-only
  finished-diff review remains required.
- PASS: JSON/YAML/workflow parsing; focused repository tests; Prometheus
  3.13.1 config syntax; ten-rule validation; all promtool rule tests; every
  dashboard/objective PromQL parse; docs consistency; diff check.
- Docker is not installed locally. Exact Grafana provisioning and live
  Prometheus API/query evidence remain delegated to the GitHub Compose job.
- Next: full local verification and independent finished-diff review.

## 2026-07-30 23:09 EEST - GH-85 review findings remediated

- Kimi's independent moving-tree review reported 0 P0, 0 P1, one P2, and five
  P3 items. The P2 identified cross-instance aggregation that could pair
  signals from different validators during the stated testnet window.
- Replaced cross-instance invariant and PNYX arithmetic with per-instance
  vector matching; added one bounded static `node` label to pair the local
  CometBFT/application sources for divergence without user-derived labels.
- Added exact 16-panel/17-target runtime gates, exact eleven-rule contracts,
  missing-family recovery evidence, pinned upstream monitoring image digests,
  Grafana datasource UID/proxy-query proof, and corrected pin wording.
- Reconciled stale wiki deployment/monitoring examples with the supported
  `cometbft_*`/`truerepublic_*` contract and removed invented Alertmanager
  destinations.
- PASS: current Prometheus config syntax, all eleven rules, all deterministic
  rule tests, and PromQL parsing. Earlier ten-rule log entries are retained as
  historical snapshots and superseded here.
- Next: complete the recovery matrix, full exact-worktree gates, focused Kimi
  remediation recheck, and final audit.

## 2026-07-30 23:22 EEST - GH-85 exact local gates and audit pass

- All eight maintained multi-validator scenarios pass, including the final
  consensus-slashing process case; the complete matrix ran in 1,208.218s.
- `make build` and `make verify` pass on the remediated tree. Race/coverage is
  green across all nine supported packages with root 69.1%, health 97.2%,
  migration 84.6%, network policy 95.5%, observability 80.3%, token 92.6%,
  treasury 97.0%, DEX 45.3%, and governance 62.2%.
- Prometheus 3.13.1 validates the configuration and all eleven rules; every
  deterministic rule fixture, dashboard expression, and objective query
  passes. JSON/YAML, focused contracts, docs consistency, and diff hygiene
  pass.
- Measured package-scoped JSON output contains 1,071 Go cases. Together with
  the unchanged 26 Rust and eight maintained-client cases, GH-85 will raise
  the public merged total from 1,104 to 1,105 after merge.
- Kimi's focused remediation recheck marks all seven reviewed areas PASS with
  no new P0/P1/P2. Sol added the optional static contract for both bounded
  `node: 'local'` labels and `on (node)` pairing.
- Added `docs/agent-bridge/GH85_AUDIT.md`: 0 FAIL / 0 WARN / 7 PASS. Local
  Docker remains unavailable; exact pinned Compose runtime proof is required
  on GitHub before merge.
- Next: publish the implementation PR, collect exact-head GitHub CI/review,
  remediate if necessary, merge, then publish the separate post-merge
  tracker/status synchronization.

## 2026-07-30 23:28 EEST - GH-85 draft PR #86 published

- Pushed reviewed implementation commit `1c44b2f` and opened draft PR #86
  against `main`.
- PR #86 documents the native dashboard, eleven alerts, objective/ownership
  contract, local verification, independent review, and the explicit
  non-production boundary.
- Posted evidence and PR linkage to GH-85 and the GH-29 rollout tracker.
- The tracker/public page remains unchanged until the implementation is
  merged; this prevents an unmerged branch from being reported as delivered.
- Next: push this append-only handoff, require exact-head GitHub checks and
  review, remediate findings, and merge only when green.

## 2026-07-30 23:38 EEST - PR #86 Docker query gate hardened

- Exact head `67d5d4c` passed Go build/race/coverage, docs, DeepScan, and all
  completed Go/Rust/Node security jobs.
- Docker proved config, pinned promtool rules/tests, stack startup, readiness,
  both targets, and Grafana health, then failed the new combined assertion
  step with all `jq` results suppressed.
- Kimi's focused read-only diagnosis identified the gap between target-up and
  lazy application-series readiness, plus first rule evaluation timing, as the
  likely race class.
- Added bounded retries for the Grafana-proxied application height, eleven
  healthy rules, and all four core app/supply series. Every assertion now
  reports a secret-free identity/count/error summary and the exact failing
  expression instead of an opaque exit code.
- Structural mismatch and invalid PromQL remain fail-fast; the change does not
  weaken any count, identity, health, or cardinality assertion.
- Next: validate and push replacement head, then rerun the complete exact-head
  GitHub matrix.

## 2026-07-30 23:45 EEST - GH-85 retry shell edge closed

- Kimi completed the bounded CI diagnosis: static dashboard/rule cardinalities
  and the focused repository test pass; first rule evaluation remains the most
  likely original one-shot race.
- Remediated its low-risk observation that GitHub `bash -e` could abort a
  retrying command substitution on a transient HTTP failure. Retryable curl
  assignments now execute as explicit conditions with initialized safe
  summaries.
- Workflow YAML, focused repository contract, and diff hygiene pass.
- Next: publish replacement head and require its complete GitHub matrix.

## 2026-07-30 23:48 EEST - PR #86 review findings fixed

- Thread-aware CodeRabbit read identified three actionable issues: Markdown
  table delimiter parsing, an application/divergence panel overlap, and
  missing low-peer recovery evidence.
- Escaped the regex delimiter in the objective table, removed the panel
  overlap, added deterministic low-peer recovery at three peers, and extended
  the repository contract to reject invalid grid bounds or any pairwise
  overlap.
- Promtool tests, dashboard JSON, focused Go contract, docs consistency, and
  diff hygiene pass.
- Next: commit/push, reply to and resolve all three threads, then require the
  complete replacement-head GitHub matrix.

## 2026-07-30 23:58 EEST - Docker cardinality root cause fixed

- Exact head `513a122` passed Go build/race/coverage and every completed
  docs/static/security gate. Docker reached the new diagnostic assertion and
  reported `result_count: 2` for the unscoped application-height query.
- The process legitimately exports registered application metrics on both
  monitored surfaces. The last four-series presence loop alone omitted the
  canonical `job="truerepublic-app"` selector used everywhere else.
- Added the explicit app-job selector to all four required-series checks while
  retaining the exact-one cardinality requirement.
- Workflow YAML, selector PromQL, focused Go contract, and diff hygiene pass.
- Next: publish and require a complete replacement-head GitHub matrix; Docker
  and recovery remain mandatory.

## 2026-07-31 00:15 EEST - GH-85 implementation merged

- Exact PR head `021d5c5` passed every required check, including the 13m47s
  eight-scenario recovery matrix and the 7m35s Docker observability proof.
- Confirmed zero unresolved review threads after all three CodeRabbit findings
  were fixed and resolved.
- Squash-merged PR #86 as `cd44fec`; GH-85 closed automatically.
- Post-merge Go CI, Security Scan, and Pages deployment are in progress.
- Next: require the exact merged-main checks to pass, then publish the
  documentation-only rollout/status closure and verify the deployed page.

## 2026-07-31 00:30 EEST - Closure synchronization locally green

- Exact merged-main `cd44fec` passed Go CI, Docker, all recovery scenarios,
  Security Scan, and Pages deployment.
- Fresh local JSON evidence counts 1,071 passing Go tests; the canonical public
  total is now 1,105 with 26 Rust and eight maintained-client tests.
- Updated GH-29's combined GH-85 checkbox. The live tracker now has 11/59
  complete, 11/51 phase work, and Phase 6 at 4/7.
- Spark completed a read-only stale-status scan. Kimi independently verified
  the test and percentage arithmetic and identified the pending live tracker
  write; Sol performed and re-read that write.
- Documentation consistency, JSON validation, and diff hygiene pass.
- Next: publish the documentation-only branch, require exact-head GitHub
  checks, merge, and verify the public Pages deployment.

## 2026-07-31 00:38 EEST - PR #87 review findings fixed

- CodeRabbit correctly identified that public-status publication remains
  branch-local until PR #87 merges and that its completion checkbox was
  premature.
- Reworded the evidence around merged implementation `cd44fec`, reopened the
  publication checkbox through merge and deployed-page verification, and
  aligned the roadmap date with the GitHub publication date.
- Next: validate, publish the corrected head, resolve both threads, and require
  the complete replacement-head GitHub matrix.

## 2026-07-31 00:46 EEST - GH-85 public closure complete

- PR #87 exact head `3ace039` passed all docs/static/security/review gates with
  zero unresolved threads and merged as `0aaba4b`.
- Post-merge Security Scan and GitHub Pages deployment pass.
- The live public page returns HTTP 200 with 1,105 tests, 11/59 overall, 11/51
  phase work, Phase 6 at 4/7, production topology as next, and the
  non-production boundary.
- Marked the GH-85 publication handoff complete. External paging, production
  SLOs/infrastructure, and rollout approval remain separate.

## 2026-07-31 12:15 EEST - GH-89 topology qualification implemented locally

- Created GH-89 and branch `feature/GH-89-topology-abuse-qualification` from
  clean merged `main` `fcf8428`; parent GH-29 records the repository-only
  boundary and unchanged production-deployment gate.
- Added a strict 256 KiB versioned JSON contract parser with duplicate-key,
  unknown-field, trailing-data, and depth rejection; no environment expansion,
  DNS resolution, node-home read, or external mutation exists.
- Added cross-node validation for synthetic public node identities, seed/
  sentry/validator/RPC composition, reciprocal sentry protection, distinct
  validator sentry zones, unique public endpoints, and explicit
  deny-by-default flow coverage.
- Added TLS/proxy-only public ingress ceilings, explicit HTTP method and route
  allowlists, bounded rate/burst/body/timeout/concurrency, RPC-only WebSocket
  routing, and fail-closed metrics/admin/unsafe/debug/peer-dial denial.
- The offline root command initially reached the Cosmos config interceptor.
  A new regression reproduced the empty-home panic; config-independent routing
  now includes `topology-policy`, and the focused root integration suite passes.
- Spark added the bounded parser/graph/ingress regression matrix. Kimi supplied
  the initial read-only architecture/security analysis; its attempted isolated
  write remained diff-free and was stopped, so Sol implemented and reviewed
  the code. A final Kimi read-only diff review is in progress.
- Current focused evidence: topology package race/coverage PASS at 84.8%;
  root command registration/config independence/repository contract PASS;
  workflow YAML and diff hygiene PASS.
- The committed example uses only synthetic node IDs and `.invalid` hosts.
  Real topology contracts and validator identity correlations remain
  operator-private. No production, server, cloud, DNS, firewall, IAM,
  deployment, key, credential, mainnet, or public-network action occurred.

## 2026-07-31 13:02 EEST - GH-89 complete local verification green

- `make verify` passes on the remediated diff: repository package selection,
  build, vet, and race/coverage across all ten Go packages. Root/application
  coverage is 69.1%; topology policy coverage is 86.2%.
- All eight maintained process drills pass. The first bounded run proved
  legacy-authority migration/rollback (631.94s), consensus recovery (228.45s),
  trusted snapshot state sync (225.85s), and backup/restore/export/import
  (199.36s). Its 25-minute global budget then ended during a linker-bound
  upgrade build, not an assertion.
- An unchanged second run proved persisted binary upgrade/rollback (368.21s),
  validator identity cold failover (218.53s), consensus-key rotation (204.70s),
  and consensus slashing/recovery (288.51s).
- Documentation consistency, contract/status JSON, workflow YAML, real daemon
  CLI validation, focused command/path-secrecy regressions, and diff hygiene
  pass. Local Docker is unavailable; exact-head GitHub Docker/Compose remains
  mandatory.
- Branch arithmetic before merge is 1,079 Go top-level cases: root 84,
  healthcheck 55, migration 82, networkpolicy 126, observability 50, token 12,
  topologypolicy 7, treasury 36, DEX 116, governance 511. With 26 Rust and
  eight maintained-client cases, the post-merge candidate total is 1,113.
  Public status remains at the merged 1,105 baseline until implementation
  merges and a separate evidence sync verifies the new main.
- GH89_AUDIT records 0 open FAIL / 0 open WARN / 7 PASS. No production or
  external infrastructure action occurred.

## 2026-07-31 14:40 EEST - GH-89 implementation merged

- PR #90 merged as `2f2acdd` after exact final head `cb14829` passed all
  11 GitHub checks: Go build/vet/race/coverage, eight multi-validator recovery
  drills, Docker/Compose restart and monitoring, docs, Go/Rust/Node security,
  DeepScan, and CodeRabbit status.
- The first CI head exposed machine-readable JSON on stderr; root streams were
  corrected and the exact Linux pipeline then passed. CodeRabbit's two valid
  close-error/API-label findings were fixed, answered, and resolved.
- Fresh package-scoped JSON evidence reproduces 1,130 Go passing cases: root
  86, healthcheck 55, migration 82, networkpolicy 126, observability 50, token
  12, topologypolicy 56, treasury 36, DEX 116, and governance 511. With
  26 Rust and eight maintained-client cases the merged total is 1,164.
- The earlier 1,113 candidate counted the new topology package by test
  functions while the historical public baseline counted passing subtests.
  It was corrected before publication to preserve one consistent method.
- GH-29 rollout counts remain 11/59 overall, 11/51 phase work, and Phase 6 at
  4/7 because no real production topology was deployed.

## 2026-07-31 15:02 EEST - GH-89 public closure complete

- PR #91 replacement head `3a07fa4` passed docs consistency, all security
  scans, DeepScan, and review after its stale `726` total was corrected to the
  reproduced 1,164-case figure; the thread was resolved and the PR merged as
  `69e498f`.
- Post-merge Security Scan and GitHub Pages deployment pass. The live page and
  status JSON return 1,164 cases, 11/59 overall, 11/51 phase work, Phase 6 at
  4/7, `production_ready=false`, and topology `production_deployment=false`.
- GH-89 acceptance criteria are checked and the issue is closed. GH-29 remains
  open with its real production-topology checkbox unchecked.

## 2026-08-01 11:02 EEST - GH-93 incident rehearsal implementation focused green

- Created GH-93 and branch `feature/GH-93-operations-runbook-rehearsal` from
  clean merged `main` `ed1b6bf`; production and live-incident actions remain
  excluded.
- Added a strict secret-free eight-scenario incident rehearsal contract,
  offline root command, synthetic fixture, CI gate, canonical incident-command
  runbook, specialist cross-links, and repository regressions.
- Fixed the stale key-rotation statement that contradicted merged ABCI++
  slashing behavior. Clarified that generic firewall/proxy snippets are
  incomplete non-production examples governed by the strict network/topology
  policies.
- Kimi review was unavailable due provider quota and produced no diff. Terra's
  independent read-only review supplied the accepted safety findings; Spark
  owned only the focused package tests; Sol reviewed and hardened all output.
- Focused incident-policy tests pass. Complete build/vet/race/coverage,
  documentation consistency, recovery-process, and review gates remain before
  publication or completion.

## 2026-08-01 12:19 EEST - GH-93 local release gate complete

- Closed all final independent-review findings with exact synthetic role
  seats, complete second-signer abort coverage, a rehearsal chain-ID namespace,
  and additional secret-leak, slashing, key-compromise, and upgrade-order
  negative tests. Terra final review: 0 open P0/P1.
- Diagnosed one loaded-host lifecycle-test failure. A separate existing
  TrueRepublic process held the inherited pprof port and heavy compilation
  exhausted the 35-second first-block window; no foreign process was changed.
  The harness now disables pprof and uses a 60-second readiness budget.
- Focused lifecycle race PASS (98.573s). Final `make verify` PASS across all 11
  maintained packages: selection, build, vet, race, and coverage. Incident
  policy coverage is 87.9%; root remains 69.1%.
- Documentation consistency, fixture/CLI JSON validation, repository contracts,
  and diff hygiene PASS. `GH93_AUDIT.md` records 0 FAIL / 0 WARN / 7 PASS.
- Docker is unavailable locally; exact-head GitHub recovery, Docker/Compose,
  security, docs, review, and unresolved-thread gates remain mandatory before
  merge. No production mutation or live rehearsal occurred.
- Fresh package-scoped JSON counting reports 1,185 Go cases; combined with the
  unchanged 26 Rust and eight maintained-client cases, the candidate total is
  1,219. Public documentation remains at the merged 1,164 baseline until main
  contains the implementation and a separate status sync is verified.

## 2026-08-01 12:36 EEST - GH-93 pre-ship reflection audit closed

- Sol found and removed rejected JSON/enum reflection. Terra independently
  identified and verified closure of positional-argument, flag, and
  input-derived scenario-count reflection. Reports now expose only fixed
  synthetic metadata and generic diagnostic categories. Final review: 0 P0/P1.
- Final focused incident-policy race/coverage PASS at 90.7%; final complete
  `make verify`, documentation consistency, exact CLI/JSON gate, and diff
  hygiene PASS.
- The new non-reflection regression raises the candidate to 1,186 Go cases and
  1,220 combined cases. This supersedes the earlier 1,219 candidate; public
  status correctly remains 1,164 until merge and independent status sync.

## 2026-08-01 12:40 EEST - GH-93 PR published

- Pushed reviewed implementation `e03823b` and opened PR #94 against `main`.
- The PR closes GH-93 on merge and preserves the no-production/live-rehearsal
  boundary. Exact-head GitHub CI, review, and unresolved-thread gates remain.

## 2026-08-01 12:57 EEST - GH-93 implementation merged

- Exact head `d3f627d` passed all 11 status gates, including Go
  build/vet/race/coverage, the real incident-validator pipeline, eight recovery
  harnesses, Docker/Compose, docs, security, DeepScan, and CodeRabbit status.
- CodeRabbit was rate-limited without content. Independent final Terra review
  reported 0 P0/P1 and GitHub reported zero review threads.
- PR #94 merged as `b6e7c29` and closed GH-93. A separate status branch now
  prepares 1,220 cases and rollout counts 12/59, 12/51, and Phase 6 5/7 while
  retaining `production_ready=false` and every live/production boundary.

## 2026-08-01 13:00 EEST - GH-93 public status candidate green

- GH-29 now marks only the merged runbook item complete; topology and capacity
  remain open.
- Synchronized source of truth and public surfaces to 1,220 cases, 12/59
  overall, 12/51 phase work, and Phase 6 5/7 without changing the false
  production-ready state or claiming a live operator rehearsal.
- Documentation consistency, JSON arithmetic, repository contracts,
  stale-value scan, and diff hygiene PASS on `docs/GH-93-rollout-status`.

## 2026-08-01 13:08 EEST - GH-93 exact-main handoff verified

- PR #94 merged exact implementation head `d3f627d` as `b6e7c29` after 11/11
  gates passed; PR #95 merged exact status head `b278a45` as `1a850d0` after
  8/8 gates passed. Both PRs had zero unresolved review threads.
- Exact `main` Security Scan run `30674576743` and Pages run `30674575864`
  completed successfully on `1a850d0`.
- Live public status was independently read back: 1,220 verified cases,
  rollout 12/59 overall, 12/51 phase work, Phase 6 5/7, with production
  readiness and deployment both false.
- GH-93 is closed and all acceptance criteria are checked. GH-29 keeps the
  real production-topology and capacity/load gates open. No production or live
  incident action was performed.
- This branch contains only the final append-only handoff; protected PR checks
  and merge are the remaining publication step.

## 2026-08-01 21:30 EEST - GH-97 capacity qualification started

- Opened GH-97 under GH-29 for the remaining repository-side Phase 6 capacity,
  disk-growth, retention, and sustained-load evidence gate.
- Created `feature/GH-97-capacity-sustained-load` from clean exact `main`
  `fce724e`.
- Scope is limited to strict secret-free evidence, bounded loopback temporary
  node workloads, tests, CI, and operator documentation. Production sizing,
  live infrastructure, public load, real identities/keys/funds, and rollout
  approval remain excluded.
- Next: inspect the existing lifecycle/metrics/pruning/retention paths, obtain a
  bounded Kimi architecture review, and define the smallest credible real-
  process qualification slice before implementation.

## 2026-08-01 22:11 EEST - GH-97 local implementation checkpoint

- Kimi's secret-free read-only review returned provider HTTP 403 quota and no
  diff. Terra supplied the independent architecture review. Spark added only
  the focused `capacitypolicy` test file; Sol reviewed it, fixed its premature
  command-output read, and tightened the resulting assertions.
- Implemented the strict capacity contract/parser/verifier/CLI, maintained
  fixture, root lifecycle wiring, repository/CI contracts, operator guidance,
  and real four-validator 24-transaction evidence harness.
- `go test ./capacitypolicy -count=1` passed before the final hardening edits.
  The first process run found and led to a fix for the CometBFT instrumentation
  section selector. A local cache-space failure was remediated only by clearing
  Go build/test caches; the current clean retry has entered test execution.
- No production system, public endpoint, real key/identity/funds, deployment,
  paging, or rollout state was changed.

## 2026-08-01 22:47 EEST - GH-97 independent-review remediation

- Removed self-attested Docker/telemetry retention values from runtime
  evidence; scoped static Compose YAML verification to `truerepublic-node` and
  preserved actual runtime-retention as an explicit limitation.
- Expanded the workload from 24 sequential transactions to 96 authenticated
  transactions across 24 consecutive four-signer waves. Added checked
  milli-TPS and average-block-time fields plus transaction-specific concurrent
  delivery latency measurement.
- Fixed fail-closed divide-by-zero handling for stalled evidence and the
  snapshot-pruning filesystem race. Added explicit failed-transaction,
  stalled-progress, throughput, block-time, exact metric, and response-size
  regressions.
- A first standard-bank attempt failed safely because that CLI module is not
  registered. The corrected run uses four independently administered synthetic
  domains and concurrent authorized membership transactions. Kimi's second
  review attempt remained provider quota-blocked (HTTP 403, no diff); Terra's
  findings were reviewed and remediated by Sol.

## 2026-08-01 23:43 EEST - GH-97 real qualification passed

- Reproduced the complete four-validator sustained-load harness: 96 submitted,
  96 committed, zero failed, 24 workload blocks, restart/app-hash/power/ledger
  checks green, and all required metrics present.
- The generated evidence measured 727 milli-TPS, 5,496 ms average block time,
  5,893 ms p95 and 6,599 ms maximum commit latency. Maximum observed node data
  growth was 1,222,727 bytes; maximum log growth was 137,971 bytes; maximum RSS
  was 154,902,528 bytes.
- Two earlier real runs showed that the harness's assumed HTTP `/snapshots`
  endpoint was not part of this CometBFT RPC surface. Replaced that ineffective
  check with strict local snapshot-store height counting after consensus is
  paused; observed retention was exactly three. Added focused fail-closed store
  counting coverage.
- `capacity-policy verify` accepted the produced evidence as valid with 96
  committed transactions and no violations. Full repository gates, audit, PR,
  exact-head CI, and merge are next.

## 2026-08-02 00:42 EEST - GH-97 local gates and audit complete

- Added `GH97_AUDIT.md`: 0 FAIL, 0 WARN, 7 PASS. Terra's final independent
  review has no remaining P0/P1/P2 finding; Kimi remained unavailable due to
  provider HTTP 403 quota and made no change.
- Passed package selection, documentation consistency, YAML parse, strict
  contract CLI/JQ assertion, diff check, CGO build, Vet, focused regressions,
  and the complete normal repository Go suite.
- Race/Coverage passed for all non-root packages. The combined local run's root
  package failed only because the almost-full local volume could not link the
  daemon built inside `TestNodeStartsStopsAndRestartsFromPersistentHome`.
  After Go-cache cleanup, the same race-instrumented test binary passed that
  exact test in 374.08 seconds; all other root tests had passed in the combined
  run. Exact GitHub CI remains the canonical clean-runner confirmation.
- Marked the local GH-97 implementation/review TODO complete. Publication,
  final-head CI, merge, issue closure, and status synchronization remain open.

## 2026-08-02 00:57 EEST - PR #98 review findings remediated

- CodeRabbit completed with one minor `noctx` finding and one trivial test-name
  and coverage observation. Threaded the Go test context into RSS sampling and
  changed the external `ps` fallback to `exec.CommandContext`.
- Renamed the non-overflow supply-cap test and added a separate
  `math.MaxInt64 + 1` supply-addition overflow regression.
- `go test ./capacitypolicy`, focused capacity root tests, `go vet
  ./capacitypolicy`, `go vet .`, gofmt, and `git diff --check` pass. The updated
  head must rerun GitHub checks before merge.

## 2026-08-02 01:15 EEST - GH-97 implementation merged; public sync prepared

- PR #98 exact head `2722e4c` passed all GitHub checks, including the combined
  recovery matrix and the dedicated capacity qualification. No unresolved
  review thread remained; the owner-author review rule was satisfied through
  the authorized admin merge. PR #98 merged as `23e0915` and closed GH-97.
- Reproduced the established full package-scoped JSON passing-case count:
  GH-97 adds 48 capacity-policy and ten root cases. The public candidate
  therefore records 1,278 total cases (1,244 Go + 26 Rust + eight
  maintained-client).
- Marked the repository-side Phase 6 capacity task complete: rollout progress
  becomes 13/59 overall, 13/51 phase work, and 6/7 in Phase 6. Production
  readiness remains false; production sizing/soak, actual runtime retention,
  private-environment evidence, and live deployment remain explicit blockers.
- Created `docs/GH-97-rollout-status` from exact merged `origin/main`. Public
  status, roadmap, landing page, wiki, TODO, project state, and Bridge are now
  synchronized locally; exact-head docs validation, PR, merge, and live Pages
  readback remain pending.

## 2026-08-02 01:20 EEST - GH-97 public status PR opened

- Opened PR #99 from `docs/GH-97-rollout-status`. A complete package-scoped
  JSON recount passed with 1,244 Go cases (root 98, capacity-policy 48, all
  remaining module counts unchanged), correcting the earlier focused-delta
  candidate before publication to 1,278 total cases.
- Documentation consistency, JSON arithmetic, stale-count scan, and diff
  hygiene pass. Terra's read-only delta recheck reports no P0/P1/P2 and confirms
  the non-production claim boundaries. Exact-head GitHub gates and merge remain
  pending.

## 2026-08-02 01:31 EEST - GH-97 public status complete

- PR #99 exact head `047f6e1` passed all applicable GitHub documentation,
  security, static-analysis, and review checks with no unresolved thread; it
  merged as `fe6e693`.
- Updated closed GH-97 so every acceptance criterion is checked and updated
  open parent GH-29 so the bounded repository-side capacity item links GH-97
  and PR #98 while preserving production sizing/soak/runtime-retention/private-
  environment work as exit gates.
- GitHub Pages deployment `30699752124` passed. Live `status.json` and landing
  page readback confirmed 1,278 cases, 13/59 overall, 13/51 phase work, Phase 6
  at 6/7, the GH-97 capacity marker, and `production_ready=false`.
- GH-97 is Done. No production, deployment, key, identity, fund, paging, or
  infrastructure mutation occurred.

## 2026-08-04 12:08 EEST - GH-102 maintained-client audit unblock started

- PR #103 exposed a current required `node-audit-client` failure unrelated to
  its GH-101 Go/docs diff. Local reproduction found high advisories in the
  maintained client's `brace-expansion`/`minimatch` dependency chain plus
  moderate React Router advisories.
- Created `fix/GH-102-maintained-client-audit` from exact `origin/main` for the
  smallest compatible maintained-client remediation. This slice will not use
  forced breaking upgrades or alter wallet/signing/network behavior.
- Mobile legacy dependency advisories remain separately open under GH-102; its
  current security job is informational (`|| true`) and will not be hidden by
  weakening CI.

## 2026-08-04 12:25 EEST - GH-102 maintained-client high findings remediated locally

- Updated only the existing `brace-expansion` override and resolved lock entry
  from 5.0.8 to compatible patch 5.0.9. This removes the eleven cascading high
  ESLint/TypeScript-ESLint/minimatch audit findings without a direct dependency,
  runtime source, or major-version change.
- Exact clean-install gate passes: `npm ci`, lint, eight tests, production
  build, and `npm audit --audit-level=high` exit 0. Two moderate React Router
  findings remain and require a deliberate v7 migration; `npm audit fix --force`
  was not used.
- Independent review and GitHub publication remain open. GH-101 PR #103 stays
  blocked until this protected fix merges and its head is rebased.

## 2026-08-04 12:32 EEST - GH-102 independent review approved

- Kimi read-only review confirms the 5.0.9 override is the smallest compatible
  5.x patch, the lock integrity matches the official cached registry record,
  and the dependency is dev-only through minimatch/ESLint. No runtime source,
  route, wallet, workflow, or signing behavior changes.
- Live `npm audit --audit-level=high` exits 0 with only the two documented
  moderate React Router findings; Kimi independently reran lint and 8/8 tests.
- Non-blocking process note: offline npm audit may falsely return an empty
  result when advisory cache bodies are absent. Only live CI audit evidence is
  accepted. Protected PR and exact-head checks are next.

## 2026-08-04 10:20 EEST - GH-101 private topology evidence gate started

- Re-read global/project agent rules, project Bridge/state/TODO/security notes,
  current GH-29, and exact Git/GitHub status. `origin/main` remains `3f1edd1`,
  no PR is open, and latest scheduled Security Scan `30790264230` passes.
- Opened GH-101 under GH-29 and created
  `feature/GH-101-private-topology-evidence` from exact clean `origin/main`.
- Scope is a strict secret-free offline evidence manifest/verifier, synthetic
  fixtures, tests, CI, audit, and operator documentation only. No real host,
  inventory, provider, firewall, DNS/TLS, production probe, deployment, key,
  identity, fund, or rollout action is authorized or performed.
- GH-29's production-topology checkbox intentionally remains open. Next: Kimi
  architecture review, then Sol-owned implementation and complete verification.

## 2026-08-04 10:47 EEST - GH-101 implementation integrated

- Kimi supplied a bounded secret-free implementation for the strict offline
  deployment-evidence parser, verifier, CLI, fixture, and package tests. Sol
  reviewed all files and changed report counts to use trusted canonical/topology
  values rather than rejected manifest declarations.
- Sol integrated the root command, config-independent lifecycle boundary, CI
  path/verification gate, operator documentation, limitations, and a repository
  contract test that rejects secret/inventory markers in the maintained fixture.
- Package tests/Vet, focused root/repository tests, root Vet, the maintained
  CLI/JQ assertion, workflow YAML parsing, gofmt, and diff hygiene pass.
- Full repository verification, independent final review, audit, protected PR,
  and exact-head GitHub checks remain open. No infrastructure was probed or
  mutated, and GH-29 Phase 6 remains 6/7.

## 2026-08-04 12:02 EEST - GH-101 local verification and review complete

- `make verify` passes all 13 Go packages with Race/Coverage; root coverage is
  69.1% and `deploymentevidence` is 90.8%. Build, Vet, package selection,
  documentation consistency, workflow YAML, maintained CLI/JQ, formatting, and
  diff hygiene pass.
- Rust format, Clippy with warnings denied, workspace build, and all 26 tests
  pass. Maintained client lint, eight tests, and build pass; mobile CI's
  pass-with-no-tests contract passes.
- Kimi's read-only final review is APPROVED with no P0/P1/P2 finding. Formal
  audit `GH101_AUDIT.md`: 0 FAIL / 0 WARN / 7 PASS.
- Claude Code was invoked for a small read-only independent test-count/status
  arithmetic check, but its OAuth session was expired and could not refresh;
  it produced no output or diff. Sol's `go test -json` recount confirms 71 new
  package cases plus one root repository case (72 new Go cases).
- A repository-wide extra check found dependency advisories in unchanged web
  and mobile lockfiles. Opened GH-102 for separate remediation; no lockfile or
  client code was changed here.
- GH-101 is locally complete. Protected PR, exact-head GitHub CI/review, merge,
  issue closure, and final Bridge/GH-29 synchronization remain open. Phase 6
  remains 6/7; no live or production action occurred.

## 2026-08-04 12:47 EEST - GH-101 review remediation

- Verified two exact-head CodeRabbit threads. Added the valid focused
  `deploymentevidence` package test before the workflow manifest gate; retained
  the correct 4 August EEST audit date and will document that timezone in the
  review response.
- Focused package test, workflow YAML parsing, and diff hygiene pass.
- Claude Code was assigned the small one-file patch but its expired OAuth
  session still could not refresh, producing no output or diff. Sol completed
  and verified the bounded fallback.
- Refreshed exact-head CI, review-thread resolution, merge, and public status
  synchronization remain open.

## 2026-08-04 13:15 EEST - GH-101 merged; public status prepared

- PR #103 exact head `e7d016a` passed all required GitHub checks: full Go
  build/Vet/race/coverage, the complete multi-validator recovery matrix,
  capacity qualification, Docker restart/monitoring, docs, Go/Rust/Node
  security, DeepScan, and review status. Both review threads were resolved.
- PR #103 merged as `7924792`; GH-101 closed automatically. No infrastructure,
  inventory, provider, DNS/TLS, firewall, key, fund, or production action
  occurred.
- Reproduced public arithmetic: GH-101 contributes 71 `deploymentevidence`
  cases plus one root repository-contract case, moving the source of truth to
  1,350 total (1,316 Go + 26 Rust + eight maintained-client).
- Kimi's independent read-only claim census confirmed all gated and non-gated
  status surfaces and the unchanged rollout boundary. Claude Code remained
  unavailable because its OAuth session could not refresh and produced no diff.
- Created `docs/GH-101-rollout-status` from exact merged `origin/main`. Rollout
  remains 13/59 overall, 13/51 phase work, Phase 6 6/7, and
  `production_ready=false`; docs verification, protected PR, merge, GH-29, and
  live Pages readback remain pending.

## 2026-08-04 13:26 EEST - GH-101 public status local pass

- Synchronized 1,350 total cases (1,316 Go + 26 Rust + eight client), root 99,
  and `deploymentevidence` 71 across every current public/status surface while
  preserving 13/59, 13/51, Phase 6 6/7, and `production_ready=false`.
- Kimi independently reproduced the new case counts and 90.8% package coverage,
  approved the production boundary, and found one premature completion box
  plus two lower-severity consistency gaps. All were remediated: the final
  publish/sync item stays open, count-bearing docs and every wiki module row are
  now CI-gated, and the merged-PR list label is accurate.
- Final local checks pass: documentation consistency, exact JSON arithmetic,
  shell syntax, and diff hygiene. Protected PR, GH-29 synchronization, merge,
  Pages readback, and final handoff remain pending.

## 2026-08-04 13:32 EEST - PR #105 review findings remediated

- First exact head `3981193` passed every applicable docs/security/static
  check; CodeRabbit reported two valid minor findings and no higher-severity
  issue.
- Refreshed the project-state timestamp and bound each new count check to the
  precise current documentation field, preventing an unrelated or historical
  occurrence from masking stale status text.
- Hardened docs consistency, shell syntax, and diff hygiene pass. GH-29 is
  synchronized with the repository evidence link while its live topology item
  remains open; refreshed exact-head checks and merge remain pending.

## 2026-08-04 13:37 EEST - PR #105 ready-to-merge checkpoint

- Exact head `1c7e34c` passed every applicable GitHub check with zero pending
  or failed status and zero unresolved review threads.
- GH-29 and Bridge are synchronized while the live topology checkbox remains
  open. The GH-101 publish/sync TODO is marked complete only after these
  conditions passed.
- A final exact-head rerun for this bookkeeping commit, merge, Pages readback,
  and post-merge handoff remain.

## 2026-08-04 13:46 EEST - GH-101 public synchronization complete

- Final PR #105 head `ec55fef` passed all applicable checks with zero unresolved
  threads and merged as `26cd19e`.
- Post-merge Security Scan `30866462881` and Pages deployment `30866461897`
  pass. Live readback confirms 1,350 cases, 13/59 overall, 13/51 phase work,
  Phase 6 6/7, `production_ready=false`, and the GH-101 evidence marker.
- GH-101 acceptance criteria and GH-29 handoff are synchronized. GH-101 is
  Done; live private topology/deployment remains open and no production action
  occurred.

## 2026-08-04 14:15 EEST - GH-102 React Router migration independently approved

- Migrated maintained `client-web` routing from 6.30.4 to 7.18.2 and preserved
  the exact 21-route contract. Added focused redirect, static/parameterized,
  search-parameter, and backslash-input coverage without loading page wallet,
  signing, store, or network behavior.
- The two prior moderate advisories are removed. The live database now reports
  only the later RSC-specific `GHSA-qwww-vcr4-c8h2`; patched router 8.3.0
  requires React 19.2.7+, so Sol approved a documented temporary exception for
  this BrowserRouter-only SPA through 2026-09-04 or pre-rollout.
- The replacement audit gate binds the exception to the exact advisory and
  package, rejects non-declarative router APIs, and fails every other
  high/critical, malformed, cyclic, or unresolved result. Independent Kimi
  review found no high or medium issue and reproduced lint, policy tests,
  Vitest, build, live audit, and diff hygiene.
- Protected PR CI, review, merge, issue checkpoint, and final Bridge/project
  state synchronization remain pending. Mobile remediation remains a separate
  GH-102 slice; no wallet, signing, broadcast, fund, deployment, or production
  action occurred.

## 2026-08-04 14:43 EEST - GH-102 maintained-client router slice merged

- PR #107 final head `694c976` passed Client Web CI, Docs Consistency, DeepScan,
  Go/Rust security, every Node audit job, and CodeRabbit. Its sole actionable
  review found a computed dynamic-import attribution gap; the final patch rejects
  every computed `import()`/`require()` specifier and the thread is resolved.
- PR #107 merged as `06f1125`. GH-102 now records both maintained-client slices
  as complete while remaining open for legacy mobile migration or retirement.
- Reproduced the current public count as 1,388 cases: 1,316 Go + 26 Rust + 40
  Vitest + six audit-policy cases. Corrected the maintained-client inventory to
  42 component files, 21 routes, eight production stores, 11 production
  services, and 317.94 kB gzip;
  extended documentation CI to compare those source counts to `status.json`.
- Rollout remains 13/59 overall, 13/51 phase work, Phase 6 6/7, and
  `production_ready=false`. No wallet, key, signing, broadcast, funds, mobile
  release, deployment, infrastructure, or production mutation occurred.

## 2026-08-04 14:56 EEST - GH-102 public router handoff complete

- PR #108 final head `4402577` passed every applicable exact-head check with no
  review thread and merged as `d1783ec`.
- Post-merge Pages `30870238036` and Security Scan `30870238562` pass. Live
  readback matches 1,388 total cases, the 46-case maintained-client breakdown,
  source-backed 42/21/8/11 client inventory, 317.94 kB gzip, 13/59 rollout,
  Phase 6 6/7, and `production_ready=false`.
- Maintained-client GH-102 work and public handoff are Done. Legacy mobile stays
  open; no production, mobile release, wallet, signing, broadcast, funds,
  infrastructure, or deployment action occurred.

## 2026-08-06 01:38 EEST - GH-110 RustSec prerequisite started

- GH-102 full verification exposed three newly published RustSec findings in
  transitive `rkyv` 0.8.15 after format, clippy, build, and all 26 Rust tests
  passed. The finding is inherited dependency drift, not caused by GH-102.
- Opened GH-110 and created clean branch `fix/GH-110-rkyv` from exact main
  `d6ebdf5` for the smallest compatible lockfile-only repair before resuming
  GH-102 publication.
- No contract source, behavior, API, key, signing, fund, deployment,
  infrastructure, or production action is in scope.

## 2026-08-06 01:43 EEST - GH-110 local repair verified

- Updated `rkyv` and `rkyv_derive` from 0.8.15 to patched 0.8.17 in
  `contracts/Cargo.lock`; the update changed ten lockfile lines and no source.
- PASS: format, all-target workspace clippy with warnings denied, workspace
  build, all 26 tests, and live `cargo audit`. The three blocking `rkyv`
  findings are cleared; five known warning-only advisories remain reported.
- Next: protected PR checks and merge, then rebase and resume GH-102.

## 2026-08-06 01:09 EEST - GH-102 legacy mobile dependency block started

- Reconstructed canonical GitHub and Bridge state after the interrupted app
  session. Exact `origin/main` is `d6ebdf5`, no PR is open, and final-main Pages
  `30870685980` plus Security Scan `30870686809` pass.
- Created clean branch `fix/GH-102-legacy-mobile` from that exact main. Scope is
  limited to reproducible dependency/advisory classification and an
  evidence-backed migration-or-retirement implementation for deprecated
  `mobile-wallet`.
- No mobile source or dependency file has changed yet. No key, mnemonic,
  account, signing, broadcast, fund, app-store, deployment, infrastructure, or
  production action is in scope.

## 2026-08-06 01:31 EEST - GH-102 legacy mobile retirement implemented

- Exact clean baseline: `npm ci` installed 1,282 packages and reproduced 51
  advisories (7 low, 16 moderate, 24 high, 4 critical); `npm test --
  --runInBand` passed only because no tests exist; Expo Doctor failed two of 17
  checks; Android export failed because Metro cannot resolve Node `crypto` from
  `@cosmjs/crypto`.
- Source review found a real mnemonic-derived signing/broadcast path against a
  production-looking RPC, plaintext mnemonic retention in React state, obsolete
  `custom/...` governance/DEX queries, and a swap UI stub. Material fixes require
  major Expo/RN/React/Navigation/CosmJS migrations plus a custody/query/test
  rewrite, not a safe lockfile upgrade.
- Kimi's independent read-only review recommends explicit retirement and found
  no P0 because the client has no release path; it records three P1 and three P2
  findings around signing/key handling, dependency exposure, dead query paths,
  zero-test CI, and documentation drift. Claude Code was assigned a small
  read-only documentation census but did not return output and was stopped; it
  made no diff.
- Removed the prototype and its obsolete workflow, replaced the non-blocking
  legacy audit with a retirement contract, and synchronized current public,
  developer, installation, roadmap, whitepaper, wiki, security, and agent
  guidance. Git history remains available for audit only.
- Local retirement/CI/docs/security/full repository verification and independent
  final review are pending. No key, signing, broadcast, fund, app-store,
  deployment, infrastructure, or production action occurred.

## 2026-08-06 01:50 EEST - GH-102 local verification complete

- PASS: mobile retirement contract, docs consistency, workflow YAML and status
  JSON parsing, shell syntax, and `git diff --check`.
- PASS: full Go `make build && make verify`; maintained `client-web` clean
  install, lint, six audit-policy tests, 40 Vitest cases, production build, and
  guarded live audit; Rust format, all-target clippy with warnings denied,
  workspace build, and all 26 tests.
- A refreshed RustSec audit found three new inherited `rkyv` vulnerabilities.
  GH-110/PR #111 isolated the ten-line compatible lockfile repair to 0.8.17;
  local and protected gates passed and it merged as `002299e` before GH-102.
- Kimi's final read-only diff review found no blocking/P0-P2 issue. Removed its
  obsolete `.gitignore` P3 and re-ran focused checks.
- Legacy `web-wallet` separately passes 18 focused tests and its build, but the
  live audit reports 70 advisories (18 low, 20 moderate, 29 high, 3 critical).
  Corrected stale status claims and opened GH-112 for migration or retirement.
- GH-102 is locally complete. Protected PR checks, merge, final issue/status
  synchronization, and final-main verification remain. No key, signing,
  broadcast, funds, app-store, deployment, infrastructure, or production action.

## 2026-08-06 01:54 EEST - GH-102 rebased and reverified

- Rebased cleanly onto GH-110 merge `002299e`, resolving only the append-only
  Bridge/ACTION_LOG overlap and preserving both histories.
- PASS after rebase: main ancestry, diff check, retirement and documentation
  contracts, workflow YAML/status JSON parsing, shell syntax, and live Rust
  audit with no vulnerabilities.
- Ready for protected GH-102 PR checks and merge.

## 2026-08-06 02:01 EEST - GH-102 merged and final-main verified

- PR #113 exact head `e606db9` passed every applicable protected check with no
  review thread and merged as `d621ff9`; GH-102 closed automatically.
- Final-main Pages `31054458625` and Security Scan `31054461930` pass. Live
  status readback confirms the retired mobile state, 1,388 tests, 13/59 rollout,
  Phase 6 6/7, and `production_ready=false`.
- GH-112 remains open only for the separate deprecated legacy-web client.
  GH-102 is complete; no key, signing, broadcast, funds, app-store, deployment,
  infrastructure, or production action occurred.

## 2026-08-06 03:14 EEST - GH-112 legacy web assessment started

- Created clean branch `fix/GH-112-legacy-web` from exact main `1a92f0a` after
  confirming no open PR and green final-main Pages `31054970470` plus Security
  Scan `31054971793`.
- Scope is the deprecated `web-wallet` only: reproduce dependencies, audit,
  tests/build, wallet/key/signing, queries, CI, and release references before an
  evidence-backed migration-versus-retirement decision.
- No key, account, signing, broadcast, funds, deployment, infrastructure, or
  production action is in scope.

## 2026-08-06 03:17 EEST - GH-112 retirement decision recorded

- Clean install reproduced 70 advisories across 1,426 installed packages (18
  low, 20 moderate, 29 high, 3 critical); all 18 focused tests pass.
- Source and chain review found real Keplr signing/broadcast functions without
  a custom message registry, a chain-rejected unsafe legacy DEX swap, and a
  chain compatibility shim retained for the client's legacy query paths.
- Selected explicit removal instead of migrating a second canonical client.
  Maintained `client-web` will replace it in the local Compose/nginx path, and a
  blocking retirement contract plus current-documentation synchronization will
  prevent accidental restoration.
- No key, account, signing, broadcast, funds, deployment, infrastructure, or
  production action occurred.

## 2026-08-06 03:49 EEST - GH-112 local implementation verified

- Removed the complete legacy web tree and replaced its non-blocking audit with
  a fail-closed retirement contract. Maintained `client-web` now owns the
  Compose/nginx route through a non-root container and loopback-only same-origin
  RPC/REST proxies.
- PASS: maintained-client lint, 6 audit-policy tests, 41 Vitest cases, build,
  and guarded audit; both retirement contracts; docs consistency; JSON/YAML,
  shell, and diff checks; focused container policy; complete Go build/verify;
  Rust format, strict clippy, build, 26 tests, and live audit with five
  warning-only advisories.
- Public state now records 1,389 cases, 14/59 rollout, Phase 6 6/7,
  `production_ready=false`, and the 318.01 kB gzip maintained-client bundle.
- Kimi completed the deep baseline review and bounded runtime integration;
  Claude Code completed a small read-only reference census. Sol reviewed every
  diff and fixed the stale Go repository contract plus one transient CI curl
  construction defect before rerunning the gates.
- Opened GH-115 for the maintained-client custom transaction registry/delivery
  gap and GH-116 for the retired legacy ABCI query shim decision. Protected
  Docker/runtime checks, final review, PR, merge, and final-main verification
  remain.
- No key, account, signing, broadcast, funds, deployment, infrastructure, or
  production action occurred.

## 2026-08-06 03:55 EEST - GH-112 independent review closed

- Kimi's final read-only review found no P0/P1 or blocking issue.
- Removed the remaining Keplr-only instructions from current user and wiki
  flows because maintained `client-web` uses a local encrypted test wallet.
- Added matching RPC/REST rate limits and metrics denial to the loopback client
  proxy, with repository-policy assertions.
- PASS: focused container policy, both retirement contracts, documentation
  consistency, workflow YAML parse, shell syntax, and diff integrity.
- Docker remains unavailable locally; the protected GitHub Compose job is the
  required runtime gate before merge.

## 2026-08-06 04:14 EEST - GH-112 PR #117 review remediation

- First protected head `e8f92fb` passed all 13 checks, including Docker/Compose
  runtime and all long Go/recovery/security jobs.
- Verified and fixed all nine CodeRabbit inline findings without expanding the
  rollout scope or test count.
- PASS: maintained-client lint, 41 Vitest plus six policy cases, build, guarded
  audit, focused Go policy, retirement contracts, docs consistency, workflow
  YAML, shell, and diff checks.
- Updated the measured Vite main-bundle value to 318.02 kB gzip.

## 2026-08-06 04:44 EEST - GH-112 merged and final-main verified

- PR #117 merged as `2d776b5369882a0e180a7ef0dad187016efc7d33`; issue
  GH-112 closed and the feature branch was removed.
- PASS on merge commit: Go CI `31062895318`, Client Web CI `31062895236`,
  Security Scan `31062895275`, and Pages `31062894585`.
- Live Pages readback confirms 1,389 tests, rollout 14/59, Phase 6 6/7,
  318.02 kB gzip, and `production_ready=false`.
- GH-115 is next; GH-116 remains open. No production or real-funds action.

## 2026-08-06 12:10 EEST - GH-115 started

- Created `feature/GH-115-custom-tx` from exact clean `origin/main` `05fb5b7`.
- Confirmed GH-115 is open, no PR is open, and the worktree is clean.
- Scope is the maintained-client custom protobuf registry and secret-free local
  client-to-chain encode/simulate/delivery evidence only.
- Risk is high because wallet signing and custom chain messages are involved.
  No real keys, accounts, signing, broadcast, funds, deployment,
  infrastructure, mainnet, or production action is in scope.

## 2026-08-06 13:18 EEST - GH-115 local implementation verified

- Reconciled all 11 maintained custom message identities and encodings against
  Go, centralized every signing path, corrected the canonical Bech32 prefix,
  and made DEX integer validation fail closed inside the registry boundary.
- The real ephemeral local-node gate delivered bank, governance, membership,
  identity, and slippage-bounded DEX transactions; it also proved expected
  authorization failure and rejection of the legacy DEX swap. Debugging this
  gate exposed and fixed missing DEX descriptors, dynamic-message signer
  extraction, first-block readiness, CORS test-environment, and gas-margin gaps.
- PASS: 1,329 package-scoped Go cases; full Go build/vet/race/coverage; 82
  Vitest plus six audit-policy cases; lint; Vite build at 319.32 kB gzip;
  guarded Node audit; separately gated local client-chain test; Rust format,
  strict clippy, build, 26 tests, and live audit with five allowed warnings;
  docs, retirement, JSON/YAML, and diff consistency.
- Public branch state records 1,443 standard cases and 16/59 rollout progress;
  the real client-chain case remains separately gated and excluded from the
  test arithmetic. Phase 6 remains 6/7 and `production_ready=false`.
- Kimi supplied the bounded registry/service implementation and is performing
  the independent final read-only review. Sol reviewed every write, added the
  chain parity/delivery evidence, and reran the full gates. Two bounded Claude
  Code census attempts were unavailable due their configured budget and made
  no diff.
- Protected PR checks, review closure, merge, issue/GH-29 synchronization,
  Pages readback, and final-main verification remain pending. No real key,
  account, fund, public broadcast, deployment, infrastructure, or production
  action occurred.

## 2026-08-06 13:31 EEST - GH-115 independent review remediated

- Kimi's deep read-only working-tree review reproduced the exact registry/wire
  tests, DEX signer test, full maintained-client suite, lint/typecheck, focused
  Go DEX checks, and the 70-second real local client-chain delivery proof. It
  found no P0/P1/P2 and recommended approval after two P3 corrections.
- Synchronized the standard maintained-client arithmetic to 82 always-run
  Vitest plus six audit-policy cases (88), excluding the separately gated real
  chain case: 1,443 total. Rebuilt at the recorded 319.32 kB gzip.
- Hardened setup-error cleanup in the integration harness. Removed two stale
  GH-115 temporary homes from early failed debugging attempts, reran the full
  successful delivery gate, and separately forced an init failure; both paths
  leave zero GH-115 homes and zero node processes.
- Kimi was stopped after its substantive review and remediation verdict because
  it continued re-reading files that Sol was actively updating; no Kimi write
  occurred during the read-only review. Protected publication remains next.

## 2026-08-06 14:13 EEST - GH-115 protected review remediation verified

- PR #119 was published at `15c9829`; every protected exact-head workflow
  passed before CodeRabbit opened 16 review threads.
- Remediated every valid finding: legacy `true` wallet addresses now migrate
  without touching encrypted payloads; RPC and signing connections fail
  boundedly; empty delivery logs retain the chain code; read-only transaction
  lookup uses `StargateClient`; custom type URLs derive from their codecs; all
  seven DEX signer contracts and protobuf wire tags are enforced and tested;
  documentation/status drift is corrected.
- Four comments were rejected with code evidence: onboarding/member addresses
  are Go strings, strict amount re-encoding is the intended fail-closed
  boundary, generated test mnemonics remain local disposable fixtures, and the
  earlier exact head had already passed CI.
- PASS: `make verify` (build/vet/race/coverage and all 1,329 standard Go
  cases), 85 Vitest plus six audit-policy cases, client lint/build at 318.91 kB
  gzip, guarded audit, exact local client-chain delivery in 70.81 seconds,
  docs consistency, both retirement contracts, JSON and diff checks. Cleanup
  leaves zero GH-115 temporary homes.
- Kimi independently rechecked the remediation and found no P0/P1/P2. Its
  status, wording, and rejection-assertion observations were corrected during
  review; no Kimi write occurred.
- Standard status is now 1,446 cases (1,329 Go + 26 Rust + 91 maintained
  client), rollout 16/59, Phase 6 6/7, and `production_ready=false`. Review
  commit, protected rerun, thread resolution, merge, issue synchronization,
  Pages readback, and final-main verification remain pending.

## 2026-08-06 14:46 EEST - GH-115 merged and final-main verified

- Review commit `b6ff9e2` passed every protected exact-head check; all 16
  review threads were answered with evidence and resolved.
- PR #119 merged administratively as
  `5d9e946e2356d2e6ec08cc3b95c1b471b5cc4fb9` because the only remaining block
  was the repository's formal review requirement. GH-115 closed automatically.
- GH-29's three corresponding checkboxes now produce exactly 16/59 overall;
  its comment records 16/51 phase work, Phase 6 at 6/7, and production false.
- PASS on merge commit: Client Web CI `31097605393`, Go CI `31097605358`
  (capacity 5m31s, build/test 6m05s, Docker 6m23s, recovery 13m54s), Security
  Scan `31097605349`, and Pages `31097604652`.
- Live Pages readback publishes 1,446 cases (1,329 Go + 26 Rust + 91 client),
  318.91 kB gzip, rollout 16/59 (31%), Phase 6 6/7 (86%), and
  `production_ready=false`.
- GH-115 is Done. GH-116 is the next bounded cleanup; transaction history,
  IBC, real ZKP, wallet custody review, accessibility/performance, broader
  quality/security depth, external audit, and staged rollout remain open.
## 2026-08-12 00:05 EEST - GH-178 started

- Created GitHub issue #178 and branch `test/GH-178-ibc-channel-recovery`
  from exact clean `origin/main` `5a22f1c`.
- Selected the remaining bounded GH-29 IBC partial-failure slice: channel
  closure, `TimeoutOnClose`, exactly-once refund/commitment cleanup, database
  recovery, and a successful replacement-channel transfer/acknowledgement.
- External relayer qualification, public-network operation, post-upgrade
  recovery, production, real keys/accounts/funds, and deployment remain out of
  scope and open.
- Next: independent Kimi design review, then Sol implementation and full
  focused/repository validation.

## 2026-08-12 00:41 EEST - GH-178 locally verified

- Added a second opt-in TrueRepublic two-chain case covering committed channel
  close, database reopen, proof-driven close-confirm/timeout-on-close,
  exactly-once refund/no-op replay, old-channel rejection, and replacement
  channel transfer/ACK over the existing connection.
- Kimi's bounded independent review confirmed the ICS-20 and ibc-go proof
  boundary. Sol remediated its one actionable P3 by requiring the exact
  `channel is not OPEN` negative reason; no P0/P1/P2 surfaced.
- PASS: focused IBC gate, full Go build/Vet/Race/Coverage, deterministic quality
  and fuzz, security contract, pinned Staticcheck/Gitleaks/Govuln policy,
  maintained-client install/lint/141 tests/build/audit, Rust format/strict
  Clippy/26 tests/audit, documentation consistency, status arithmetic, and diff
  integrity.
- Public candidate state is 1,609 cases (1,442 Go + 26 Rust + 141 client),
  rollout 25/59 and production false. Protected PR checks, merge, final-main,
  GH-29 and Pages readback remain before Done.

## 2026-08-12 01:20 EEST - GH-178 completed

- Protected PR #179 reviewed exact head `02b3aef`; all 20 contexts passed.
  Three CodeRabbit documentation findings were fixed, answered, and resolved,
  leaving zero open review threads.
- PR #179 merged as `ef8d600`; GH-178 closed automatically. Final-main Go
  `31541112827`, Security `31541112905`, reproducible Linux `31541112832`, Docs
  `31541112826`, and Pages `31541111526` all passed.
- Live Pages readback confirms 1,609 standard-suite cases, rollout 25/59,
  Phase 6 6/7, production false, and the GH-178 IBC recovery status.
- GH-29 now credits GH-178 channel close/replacement and timeout-on-close while
  correctly retaining post-upgrade recovery as open. No production or external
  network action occurred.

## 2026-08-12 10:27 EEST - GH-181 started

- Created GitHub issue #181 and branch `test/GH-181-ibc-compatible-upgrade`
  from exact clean `origin/main` `968f334`.
- Selected the remaining bounded IBC recovery slice: preserve a real
  Tendermint-client/connection/unordered ICS-20 channel and pending packet state
  across a compatible application binary restart, then complete ACK and
  timeout/refund paths exactly once.
- This evidence must remain explicitly distinct from an `x/upgrade`, governance
  halt, consensus migration, public relayer, production deployment, or claim of
  arbitrary cross-version compatibility.
- Kimi completed the initial read-only architecture review. It confirmed the
  existing in-process restart coverage and the absence of a production
  `x/upgrade` path; Sol retains implementation, security, integration, GitHub,
  and final verification ownership.

## 2026-08-12 11:15 EEST - GH-181 locally verified

- Added a third opt-in two-chain scenario. A separately linked candidate test
  binary reopens both original LevelDB states in place, verifies preserved
  client/OPEN-connection/unordered-channel/sequence/escrow/packet/receipt/ACK
  state, completes the pending ACK, and proves fresh ACK plus timeout/refund and
  their duplicate no-op economics.
- The exact candidate's intentional exit-42 phase runs before manifest read or
  database open; recursive path-and-byte hashes for both application databases
  remain identical before baseline continuation.
- Kimi found no P0/P1/P2. Sol remediated both P3 hardening notes with immediate
  duplicate-relay balance assertions and a child timeout below the parent gate,
  then reran the focused and complete relevant suites.
- PASS: all three IBC scenarios, build, package selection, Vet, Race/Coverage,
  deterministic/property/fuzz quality, security contract, pinned Staticcheck,
  Gitleaks and Govuln, maintained-client install/lint/141 tests/build/audit,
  Rust format/strict Clippy/26 tests/audit, documentation consistency, JSON and
  diff integrity.
- Candidate public state is 1,610 cases (1,443 Go + 26 Rust + 141 client),
  rollout 26/59, Phase 6 6/7, and production false. Protected PR/CI, merge,
  final-main, GH-29, both Bridges, and live Pages remain before Done.

## 2026-08-12 11:45 EEST - GH-181 completed

- Protected PR #182 reviewed exact head `68341f8`; all 20 reported contexts
  passed. Four CodeRabbit findings were fixed, answered, and resolved, leaving
  zero open review threads.
- PR #182 merged as `2f22b60`; GH-181 closed automatically. Final-main Docs
  `31579387559`, Security `31579387595`, and Pages `31579386407` pass. The
  first Pages deploy hit a hosted-runner certificate error after successful
  build/upload; its unchanged-commit retry passed.
- Live Pages confirms 1,610 standard-suite cases, rollout 26/59, Phase 6 6/7,
  production false, and the GH-181 compatible-restart status.
- GH-29 now credits compatible binary-restart recovery while correctly keeping
  governance-controlled migrations and IBC client upgrades open. No production
  or external-network action occurred.
## 2026-08-12 13:30 EEST - GH-184 locally verified

- Wired real `x/upgrade`, non-signing module authority, release handler, and
  immutable-electorate two-thirds schedule/cancel governance.
- Kimi K3 contributed the bounded secret-free truedemocracy adapter and review;
  Sol reviewed the diff and added adversarial keeper and process evidence.
- The four-validator three-artifact harness passes old-binary halt, intentional
  cached partial-write failure rollback, fixed candidate recovery, common app
  hashes, preserved power, and exact-once restart in 230.63 seconds.
- Updated candidate rollout/status/runbook/limitations/threat/Bridge records to
  1,621 cases, 28/59 overall, 28/51 phase work, Phase 6 6/7, production false.
- Full local gates, protected PR, final-main, tracker, Pages, and closure remain
  pending; no production, public network, real key, fund, or deployment action
  occurred.

## 2026-08-12 14:22 EEST - GH-184 local gates and final review complete

- Remediated immutable electorate, pending-plan export/import, stale and
  post-execution cleanup, malformed-genesis rejection, exact release identity,
  deterministic build verification, dedicated CI ownership and handler
  coverage without lowering any quality floor.
- PASS: 1,624 standard cases (1,457 Go, 26 Rust, 141 client), full Race/Coverage,
  property/fuzz, security scanners, client, Rust/audit, docs/JSON/diff, and the
  fresh 231.39s four-validator governed-upgrade process gate.
- Kimi final read-only review reports no P0/P1/P2. Sol confirmed the fresh count
  and added documented/tested 2/3 post-execution cleanup, resolving actionable
  P3 notes. Protected PR/CI, merge and public closure remain.

## 2026-08-12 14:52 EEST - GH-184 completed

- Protected PR #185 exact reviewed head `9c4f58c` passed all 21 reported
  contexts with zero review threads and merged as `bf32ad5`; GH-184 closed and
  its remote feature branch was deleted.
- Final-main Docs `31592981577`, Security `31592981538`, reproducible Linux
  `31592981655`, Pages `31592980432`, and Go `31592981562` pass.
- GH-29 now credits the bounded governed v0.4.1 application-upgrade recovery.
  Live Pages confirms 1,624 cases, rollout 28/59, Phase 6 6/7, production
  false, and the governed-upgrade status.
- Kimi supplied bounded secret-free implementation and independent review;
  Sol reviewed every write and retained architecture, security, integration,
  full verification, GitHub, merge, tracker, and closure ownership.
- No production deployment, public-network action, real key/account, funds,
  release, or IBC client upgrade occurred. GH-184 is Done.

## 2026-08-12 23:27 EEST - GH-187 implementation and process gates verified

- Replaced success-like staking/distribution compatibility stubs with stable
  fail-closed IBC/CosmWasm adapters and explicit query/message rejection.
- Added four standard-suite boundary cases proving every required adapter
  rejects and `x/staking`/`x/distribution` remain absent from module, store,
  genesis, gRPC, message-router, and CLI surfaces. Fresh arithmetic is 1,628
  total: 1,461 Go, 26 Rust, and 141 maintained-client.
- PASS: focused boundary/IBC/module tests, 51.06s two-chain transfer/channel/
  restart recovery, 236.47s four-validator governed-upgrade recovery, docs
  consistency, JSON, security contract, formatting, and diff checks.
- Kimi supplied the bounded secret-free architecture review; Sol implemented
  and reviewed all writes. Candidate rollout becomes 30/59 overall and 30/51
  phase work; Phase 6 remains 6/7 and production readiness remains false.
- Full local/security gates, final review, protected PR/CI, merge, GH-29,
  final-main, live Pages, and both-Bridge closure remain pending. No production,
  external network, real key/account/fund, migration, release, or deployment
  action occurred.

## 2026-08-12 23:48 EEST - GH-187 full local/security gates pass

- PASS: Go build/vet/full Race/Coverage, critical floors, property/fuzz,
  contention/replay/restart, IBC two-chain, governed upgrade, Staticcheck,
  exact-policy Govuln plus fixtures, Gitleaks, docs/JSON/security contract.
- PASS: maintained-client clean install, lint, 141 tests, production build,
  bundle budgets and high audit; Rust format, strict Clippy, build, 26 tests,
  and cargo audit with the existing allowed warning set.
- Kimi's read-only final review found one test-attribution documentation issue:
  GH-187 standard cases were grouped with separately gated evidence. README and
  both wiki references now correctly count GH-187 inside the 1,461 Go cases.
- Remaining before publication: focused post-remediation confirmation,
  clean-commit deterministic Linux build, protected PR exact-head CI/review,
  merge, final-main/GH-29/live Pages and both-Bridge closure.

## 2026-08-12 23:54 EEST - GH-187 review commit and platform boundary

- Created reviewed implementation commit `e147dce`; the repository was clean.
- The deterministic build contract then rejected this macOS host as designed:
  `linux-amd64 requires a native Linux x86_64 runner`. This is not a product
  failure and cannot be claimed locally; protected reproducible-Linux CI owns
  the required native proof for the exact published head.
- Kimi follow-up reports the documentation finding RESOLVED and no remaining
  P0/P1/P2 in the remediated public-count diffs.

## 2026-08-12 23:58 EEST - GH-187 published as PR #188

- Pushed `refactor/GH-187-protocol-boundary` and opened protected
  [PR #188](https://github.com/NeaBouli/TrueRepublic/pull/188), closing GH-187.
- PR body records implementation, root cause, all real local evidence, the
  native-Linux CI ownership boundary, residual scope, and prohibited actions.
- This append-only handoff becomes the final candidate head; exact-head checks,
  review threads, merge and final public closure remain pending.

## 2026-08-13 00:30 EEST - GH-187 completed

- Exact PR head `ef6afbb` passed all 22 reported contexts with zero review
  threads and PR #188 merged as `6c8d2a8`; GH-187 closed and the remote feature
  branch was deleted.
- Final-main Go `31641633848`, Docs `31641633858`, Security `31641633794`,
  reproducible Linux `31641633845`, and Pages `31641632893` pass.
- GH-29 and live Pages now publish 1,628 cases, rollout 30/59, phase work
  30/51, Phase 6 6/7, and production false.
- Kimi performed bounded secret-free read-only architecture and final reviews;
  Sol owns every implementation write, integration, full test, GitHub action,
  merge, tracker update, and closure decision.
- All scoped gates are satisfied. GH-187 is Done; no production/public-network,
  real key/account/fund, migration, release, or deployment action occurred.

## 2026-08-13 02:23 EEST - GH-190 started

- Created GH-190 and branch `feature/GH-190-ibc-transfer-ux` from exact clean
  `origin/main` `c845e99` to close GH-29 Phase 4's IBC transfer UX item.
- Kimi's bounded read-only architecture review identified two prerequisites:
  strict native amount parsing and typed IBC channel query failure. It designed
  a canonical MsgTransfer registry/service/store/UI boundary with evidence-only
  status and no automatic resubmission.
- Kimi receives a substantial bounded implementation slice; Claude Code gets
  only a small test-impact inventory. Sol retains architecture, security,
  integration, full verification, external writes, merge and closure.
- No external relayer, counterparty, IBC client upgrade, ZKP, wallet-custody,
  contract, migration, production, real-key/account/fund, release or deployment
  action is in scope.

## 2026-08-13 03:16 EEST - GH-190 locally verified

- Implemented canonical native ICS-20 signing, strict channel/amount/fresh
  balance/timeout boundaries, scoped non-secret persistence, recovery UI, and
  manual source-chain packet lifecycle reconciliation with no auto-resubmit.
- PASS: complete client, disposable-chain, Go/Race/Coverage, IBC two-chain,
  security/static/vulnerability/secret, Rust, docs/JSON and retirement gates.
- Candidate public arithmetic is 1,746 cases and 31/59 rollout. GH190_AUDIT.md
  records exact evidence, delegation, limits and remaining protected PR work.

## 2026-08-13 03:18 EEST - GH-190 published as PR #191

- Pushed implementation commit `7eca8eb` and opened protected PR #191 against
  `main`, closing GH-190 and advancing GH-29. Exact-head CI/review, merge,
  final-main/live Pages, tracker and both-Bridge closure remain pending.
