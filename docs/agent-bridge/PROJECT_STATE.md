# Project State

Updated: 2026-08-09 23:15 UTC

## Repository

- GH-172 is complete through PR #173. Exact reviewed head `671ff37` passed
  18/18 contexts with zero open review threads and merged as `293ea35`.
  Final-main Go `31340763723`, Security `31340763717`, reproducible Linux
  `31340763708`, Docs `31340763710`, and Pages `31340763413` pass. Live Pages
  publishes 1,607 cases (1,440 Go, 26 Rust, 141 client), GH-172 evidence,
  23/59 overall, 23/51 phase work, Phase 5 5/6, Phase 6 6/7, and production
  false. GH-29 folds this evidence into its existing quality item.

- GH-169 is complete through PR #170. Exact reviewed head `620c41d` passed
  18/18 contexts with zero unresolved review threads and squash-merged as
  `d50c131`. Final-main Go `31336109586`, Security `31336109601`, reproducible
  Linux `31336109590`, Docs `31336109615`, and Pages `31336109276` pass. GH-29
  and live Pages publish 1,606 cases, 23/59 overall, 23/51 phase work, Phase 5
  5/6, Phase 6 6/7 and production false. Independent external security review
  remains open.

- GH-169 is locally verified on `security/GH-169-threat-model`. The canonical
  v1 register covers nine domains with 18 threats and 33 repository evidence
  paths; its fail-closed root contract passes 19 negative mutations. Kimi's
  independent review found no P0/P1, and Sol remediated its one P2 plus both
  actionable P3 hardening notes. Fresh source-of-truth arithmetic is 1,606
  cases (1,439 Go, 26 Rust, 141 client), rollout 23/59 overall and 23/51 phase
  work, Phase 6 6/7, production false. Protected PR/CI, merge, GH-29, final-main
  and Pages verification remain pending.

- GH-161 is complete. Safe Cargo PR #156 merged as `4aeddc0`; incompatible
  grouped npm #157 and Go #158 were closed with concrete evidence and replaced
  by bounded PRs #163 (`3a3919c`) and #164 (`62ad5e9`). Typed policy PR #162
  merged as `7b3f183` and enforces Go patch-only maintenance, the Playwright and
  React Refresh compatibility boundaries, and ungrouped Cargo `cosmwasm-*`
  crate updates. Policy-generated Go patch PR #167 received separate Sol/Kimi
  review and merged as final code head `87c47c1`; superseded generated PRs
  #165/#166 auto-closed. Kimi approved the reviewed slices with no remaining
  P0/P1/P2; every final exact-head context passed with no open review thread.
  Exact merged-main Go `31315620763`, reproducible Linux `31315620777`,
  Security `31315620755`, Dependency Graph `31315623437`, and Pages
  `31315620166` pass. Public rollout arithmetic remains 1,579 cases, 22/59
  overall, 22/51 phase work, Phase 6 6/7, and production false because
  maintenance does not complete a rollout gate.

- GH-148 is closed through merged PR #149 (`5baeea6`). All 35 third-party
  Action uses are immutable SHA pins; toolchains/scanners are centralized and
  exact; every job is bounded; Go vulnerability/static, maintained-tree secret,
  Cargo/Node audit, and lock gates fail closed with committed negative fixtures.
  Kimi's independent final review found no P0/P1/P2 after remediation. Exact PR
  head passed 22/22 checks; final-main Go `31307697100`, Security
  `31307697098`, Rust `31307697110`, Client `31307697087`, reproducible Linux
  `31307697093`, Docs `31307697083`, and Pages `31307696631` pass. Live Pages
  publishes 1,579 cases (1,412 Go, 26 Rust, 141 client), rollout 22/59 overall
  and 22/51 phase work, Phase 6 6/7, and production false. GH-29 is synchronized.
- GH-153 is closed through PR #155: exact head `3f698a2` passed 18/18 protected
  checks and merged as `8489d54`. The policy groups only minor/patch version
  updates, ignores ordinary automatic semver-major updates in all four
  ecosystems, and validates exact YAML structure in a repository regression.
  Final-main Go `31309568219`, Security `31309568223`, reproducible Linux
  `31309568233`, Docs `31309568209`, Dependency Graph `31309569964`, and Pages
  `31309567545` pass. The new bounded sweep generated minor/patch PRs
  #156-#158; separate Cargo/Go updater-resolution failures remain explicit
  follow-up findings and no dependency update is auto-merged.
- GH-154 is closed through PR #159. All 31 remote branches whose current
  SHA exactly matched a merged PR head and all 17 clean exact-merged local
  worktrees were removed after immediate revalidation; no forced worktree
  removal was used.
  - **Remote inventory (six heads):** `main`; active Dependabot PR branches
    #156-#158; closed-unmerged `agent/public-recovery-status`; and
    `solana-archived`.
  - **Preserved local inventory (six worktrees):** the dirty/divergent legacy checkout;
    four clean but historically divergent worktrees; closed-unmerged recovery
    status.
  Independent final review approved the audit with no P0-P3. PR #159 exact head
  `46f687b` passed 11/11 checks and merged as `ac8fbbc`; final-main Docs
  `31310820782`, Security `31310820753`, and Pages `31310820227` pass on that
  commit. Issue #154 is closed, its remote task branch is gone, and GH-154 is
  Done. Dependabot PRs #156-#158 remain separate bounded maintenance work.

- GH-145 is closed through merged PR #146 (`50f18e0`). Two real seeded fuzz
  targets plus deterministic DEX, PNYX cap/malformed-genesis, governance
  replay/no-drift, and focused Race evidence pass local, protected exact-head,
  and final-main gates. Final-main Go `31280519207`, Security `31280519224`,
  and Pages `31280518754` pass; live Pages publishes 1,573 cases (1,406 Go, 26
  Rust, 141 maintained client), rollout 21/59 overall and 21/51 phase work,
  Phase 6 6/7, and production false. GH-29 is synchronized; GH-145 is Done.
- GH-141 is closed through merged PR #142 (`11ef2f6`) after every protected
  exact-head gate passed, including the authoritative Docker image/restart
  smoke. The maintained-client Docker builder now includes its required build
  scripts and retains bundle-budget enforcement.
- GH-139 is closed through merged PR #140 (`1b6ccef`): a repository-owned
  coverage contract guards root/application at 71.8%, DEX at 51.0%, and
  governance at 63.7%. Fresh
  results are 71.9%, 51.1%, and 63.8%, backed by 15 new standard Go cases for
  atomic rollback, DEX authority/custody/slippage, governance escrow/auth/reward
  payout/withdrawal, and the repository contract. Current public status
  is 1,535 cases (1,368 Go, 26 Rust, 141 maintained client), rollout 20/59
  (20/51 phase work), Phase 6 6/7, and production false. Exact-head checks and
  final-main Go `31277242596`, Security `31277242587`, and Pages `31277242292`
  pass; live Pages and GH-29 are synchronized. Remaining rollout/release gates
  still prohibit production readiness.

- GH-131 is PR-ready after full local review: the maintained client now exposes
  an honestly labelled newest-first submitted-transaction history over the
  Cosmos SDK v0.50 `query/page/limit/total` contract, with typed failures,
  stale-wallet protection, committed-failure preservation, and accessible
  pagination. Both disposable-chain cases pass; the standard baseline is
  1,519 cases (1,352 Go, 26 Rust, 141 maintained client). The combined GH-132/
  GH-131 build is 76.29 kB gzip initial, 5.03 kB maximum lazy route, and
  353.44 kB total JavaScript gzip. Staged rollout is 19/59 (19/51 phase work),
  Phase 6 remains 6/7, and production readiness remains false.
- GH-128 is closed through merged PR #129 (`47f8c08`): all 19 page routes
  are lazy entries, the eager MobileNav-to-CosmJS dependency chain is broken,
  and chunk rejection reaches the application error boundary. The initial
  entry falls from 322.63 kB gzip to a deterministic 75.79 kB gzip; raw/gzip
  entry, route, chunk, and total budgets now fail the build on regressions. The
  fresh local baseline is 1,482 cases (1,352 Go, 26 Rust, 104 maintained
  client). Exact-head checks passed without an unresolved review thread; final-
  main Client `31255695662`, Security `31255695671`, and Pages `31255695086`
  pass on merge commit `47f8c08`. GH-132 adds a pinned protected
  Chromium/Firefox/WebKit accessibility, responsive, delayed-loading, and
  third-party-request matrix for safe unauthenticated routes; exact PR #136
  CI is green. The GH-132 rollout state is 18/59. Phase 6 remains 6/7 and
  production readiness remains false.
- GH-121 is closed through merged PR #126 (`239e6c3`): all 23
  maintained-browser consumers of
  18 unregistered custom-module REST aliases now use one fail-closed registered
  protobuf gRPC-over-CometBFT-ABCI JSON-RPC transport. Two missing canonical
  reads, Merkle proof and Pay-to-Put, are registered without adding a listener
  or grpc-gateway route. The disposable-chain browser test proves governance,
  membership, identity, DEX, Merkle, Pay-to-Put, and unavailable-pool behavior.
  The fresh local baseline is 1,476 cases (1,352 Go, 26 Rust, 98 maintained
  client). Exact-head PR checks passed, all eight review threads were resolved,
  and independent review found no open P0/P1/P2 finding. Merge commit `239e6c3`
  passed final-main Client `31252288314`, Security `31252288325`, Go
  `31252288303`, and Pages `31252287866`; rollout is still 16/59 and production
  readiness remains false.
- GH-116 is closed through merged PR #124 (`5dc4487`): the consumerless legacy
  `custom/truedemocracy/...` and `custom/dex/...` ABCI shim is removed, all 17
  supported module query methods remain registered through protobuf gRPC, and
  protocol-boundary tests prove both live gRPC-over-ABCI execution and
  fail-closed retired paths. The fresh baseline is 1,450 cases (1,333 Go, 26
  Rust, 91 maintained client). Browser REST aliases for these custom modules
  were never registered and remain a separately bounded compatibility defect
  in GH-121; production readiness remains false. The newly published
  GHSA-5p4m-2wfm-xmqj prerequisite is separately fixed by merged GH-122 / PR
  #123 (`1257484`). Final head `df465e8` passed PR-native Go, Docs, Security,
  DeepScan, independent review, and exact-head Client/Rust verification with
  zero review threads. Merge commit `5dc4487` then passed final-main Go
  `31131404107` attempt 2, Docs `31131405747`, Client `31131407330`, Rust
  `31131408773`, Security `31131410151`, and Pages `31131318674`. Rollout
  remains 16/59 and production readiness false.
- GH-102 is closed through merged PR #113 (`d621ff9`), GH-112 is closed
  through merged PR #117 (`2d776b5`), and GH-115 is closed through merged PR
  #119 (`5d9e946`): both legacy clients are removed, maintained `client-web` is
  the sole public client, and its exact fail-closed registry plus centralized
  simulate/sign/deliver boundary are proved against a disposable chain. All 16
  review threads are resolved, exact-head and merge-commit workflows pass, and
  its GH-115 merge published 1,446 standard cases with rollout 16/59. Independent
  review found no P0/P1/P2. Production readiness remains false.
  GH-116 subsequently removed the retired client's consumerless legacy
  `custom/...` query shim through PR #124.
- GH-56 is closed through merged PR #62 (`80ab674`). Authenticated atomic
  consensus-key rotation, permanent revocation, separate bootstrap operator
  authority, deterministic CometBFT H+2 activation, and real five-process
  export/import evidence are now on `main`.
- PR #58 separately merged the requested developer BTC support address through
  GitHub's supported custom funding link; the team multisig remains unchanged.

- GitHub: `NeaBouli/TrueRepublic`
- Baseline: canonical `origin/main`; exact implementation and evidence commits
  are recorded in `ACTION_LOG.md` so this live state does not self-expire after
  documentation-only merges.
- GH-89 implementation is merged through PR #90 (`2f2acdd`). The strict,
  versioned, secret-free cross-node topology qualification contract composes
  GH-71 per-home role policy and validates reciprocal sentry protection,
  distinct sentry zones, validator isolation, relationship-backed
  deny-by-default flows, and bounded method/route/query ingress. The current
  docs branch publishes the merged 1,164-case status. This is repository
  evidence only; GH-29's actual production-deployment checkbox remains open.
- GH-93 implementation is merged through PR #94 (`b6e7c29`). The strict
  secret-free incident-command contract, offline validator, eight synthetic
  rehearsal scenarios, specialist runbook links, repository/CI gates, and
  non-reflecting diagnostics passed all exact-head checks. The current docs
  branch publishes 1,220 cases and 12/59 rollout progress. Private live
  operator rehearsal, production topology, paging, and capacity remain open.
- GH-97 is merged through PR #98 (`23e0915`). Its strict secret-free capacity
  contract, real temporary four-validator 96-transaction workload, bounded
  resource/retention evidence, restart and ledger checks, fail-closed verifier,
  CI gate, and operator guide passed local and exact-head GitHub verification.
  Its GH-97 status sync recorded 1,278 cases and 13/59 rollout progress while
  preserving production sizing, multi-day soak, private-environment evidence,
  and deployment as open boundaries.
- GH-101 is merged through PR #103 (`7924792`). Its strict secret-free,
  digest-bound deployment-evidence envelope, offline verifier, synthetic
  fixture, CI gate, operator procedure, and 0 FAIL / 0 WARN audit passed local
  and exact-head GitHub verification. Public status records 1,350 cases while
  rollout remains 13/59 overall and Phase 6 remains 6/7 because no private live
  deployment or infrastructure evidence was produced.
- GH-102's maintained-client dependency slices are merged through PR #104
  (`91ee957`) and PR #107 (`06f1125`). The high-severity dev-only chain and both
  prior moderate React Router advisories are cleared; the exact 21-route
  contract, 40 Vitest cases, six audit-policy cases, and guarded live audit pass.
  A time-boxed exception for the later RSC-only `GHSA-qwww-vcr4-c8h2` is bound
  to the BrowserRouter-only architecture through 2026-09-04 or pre-rollout.
  Public status now records 1,388 cases (1,316 Go + 26 Rust + 46 maintained
  client). Its production inventory is 42 component files, 21 routes, eight
  stores, and 11 services. The legacy mobile prototype was subsequently
  retired and removed under GH-102; this paragraph preserves the historical
  dependency-slice evidence.
- Merged evidence and recovery PRs: #9, #15, #16, #17, #18, #19, #22, #23, #24, #27,
  #28, #30, #31, #33, #34, #35, #40, #42, #44, #46, #49, #52, #54, #57,
  #58, #62, #65, #67, #69, #72, #75, #78, #81, #86, #87, #88, #90, #94,
  #98, #103, #104, and #107.
- Current work: GH-60 inactive validator claim round-trip is closed through
  merged PR #67 (`c5a3d38`). Explicit active/inactive representation, legacy
  compatibility, complete domains, exact stake/jail/missed-block/power
  recovery, fail-closed malformed resurrection, and the real four-process
  export/import drill pass locally and on the final GitHub head. GH-61 is
  closed through merged PR #69
  (`264ab7c817414e3081149c61d8b5b6fb0ce5e368`). The merged bounded
  fresh-genesis migration includes canonical fresh-key proofs, exact raw-export
  binding, typed state/Comet reconciliation, an atomic trusted-hash CLI, the
  historical pre-GH-56 four-validator migration/rollback drill, and the
  operator runbook. Final-head GitHub CI passed all 11 checks and all review
  threads are resolved. GH-71 is closed through merged PR #72
  (`9c369ac0d589f749e055af33a03f5f4981020101`): deterministic role-based
  node policy, fail-closed startup, safe container/proxy/monitoring defaults,
  and the operator runbook are now on `main`. Its exact final head passed Go
  build/race/coverage, the complete recovery matrix, full Compose runtime,
  docs, security scans, and DeepScan. CodeRabbit remained stuck without a
  finding or review thread for about ten hours; the independent Kimi review
  found no P0/P1 and all lower-severity observations were remediated before
  the final green run. GH-29 remains open as the rollout execution tracker.
  GH-32 and
  PR #33 close its first Phase 1 gate with local and GitHub evidence. GH-39 is
  now merged via PR #40 with green GitHub CI for validator
  join/replacement/restart-catch-up evidence plus Keeper/ABCI power-zero leave
  coverage. GH-41 network partition, delayed peer, quorum progress,
  reconnect/catch-up, app-hash convergence, validator-power, and bank-backed
  export evidence is merged through PR #42 with green GitHub checks. GH-43
  trusted snapshot state-sync catch-up is merged through PR #44 with green
  local and GitHub evidence. GH-45 sanitized backup/restore/export/import
  drills are merged through PR #46 with green local and GitHub evidence. GH-47
  bounds the Go CI `build-and-test` job with a 20-minute timeout after a stale
  GitHub runner left the PR #46 merge-commit run in progress despite local Go
  tests passing; refreshed `main` Go CI passes at `63b76bf`.
- GH-53 is closed through merged PR #54 (`3e44905`). Its four-validator process
  drill proves compatible versioned binaries rolling across persisted homes, a
  deterministic candidate failure before state opens, return of every
  validator to the baseline binary, unchanged identity keys, non-regressing
  signing positions, historical/current app-hash and validator-power
  convergence, and recovered non-empty export/validation/re-import. The
  operator runbook now forbids stale signer-state restoration. PR #54 final
  head passed Go race/coverage, multi-validator recovery, Docker restart,
  docs, Go/Rust/Node security, DeepScan, and CodeRabbit; all six review threads
  are resolved.
- GH-56 is closed through merged PR #62 (`80ab674`). Final head `239cc6f`
  passed Go build/vet/race/coverage in 6m44s, the combined recovery/rotation/
  state-sync/backup/identity/upgrade process matrix in 9m39s, Docker restart,
  Docs, Go/Rust/Node security scans, and DeepScan. CodeRabbit was rate-limited;
  the recorded independent adversarial review found no P0 and no additional
  P2. GH-59 and GH-60 are closed; residual consensus-state rollout work is
  bounded by GH-61.
- GH-59 is closed through merged PR #65 (`934a042`). Final head `313b327`
  passed all 11 GitHub checks: Go build/vet/race/coverage in 7m05s, the combined
  recovery/slashing/rotation/state-sync/backup/identity/upgrade process matrix
  in 11m46s, Docker restart in 3m15s, docs, Go/Rust/Node security scans,
  DeepScan, and the rate-limited CodeRabbit status. The independent final review
  found 0 P0 / 0 P1 / 0 P2 and no unresolved review thread remained.
- GH-48 closed the 2026-07-22 fast audit reconciliation. Local/GitHub state was
  synchronized with no open PRs before the task; the audit found no recovery-
  foundation failure, corrected live documentation that still called merged
  PRs unmerged, and records three residual warnings: incomplete rollout gates,
  root Go wildcard discovery under installed frontend dependencies, and the
  maintained-client bundle size.
- GH-50 closed GO-2026-5970, first surfaced by PR #49 after the vulnerability
  database changed since the last green scheduled scan. The minimal remediation
  updates `golang.org/x/text` from v0.37.0 to the scanner-reported fixed
  v0.39.0. Exact local Security Scan filtering, Go build, vet, and all 655 Go
  tests pass. PR #49 final-head Go build/race/coverage, multi-validator recovery,
  Docker restart, Go/Rust/Node security, docs, DeepScan, and CodeRabbit checks
  are green; the PR merged as `7dbde85` and closed GH-48/GH-50.
- GH-51 is closed through merged PR #52 (`ae7105a`). One repository-owned
  wrapper derives the root-module package set from Git-managed, non-ignored Go
  sources, excluding installed frontend dependency trees. Local selector,
  concurrent frontend install plus build/vet/race/coverage, normal tests, and
  all three multi-validator recovery harnesses pass. Final-head GitHub Go,
  Docker, security, docs, and static-analysis checks are green; all four valid
  CodeRabbit findings were corrected and their threads resolved.
- Active recovery checkout:
  `/Users/gio/Documents/Codex/2026-07-11/erkunden/TrueRepublic-gh20`
- GH-26 branch: `fix/GH-26-pod-init-script`
- GH-26 issue: #26; PR #27 is verified and merged to `main`.
- GH-26 recovery checkout:
  `/Users/gio/Documents/Codex/2026-07-11/erkunden/TrueRepublic-gh26`
- Recovery worktree: `/Users/gio/Desktop/repos/TrueRepublic-recovery`
- Legacy local checkout: preserved at `/Users/gio/Desktop/repos/TrueRepublic`
- GitHub epic: #4
- Current open GitHub issue set after GH-89 closure: #4 recovery epic, #7
  audit/review parent, and #29 rollout tracker.

## Verified state

- GH-89 is implemented through merged PR #90 (`2f2acdd`). Final head
  `cb14829` passed Go build/vet/race/coverage, the strict real-daemon topology
  JSON pipeline, all eight recovery drills, Docker/Compose restart and
  monitoring, docs, Go/Rust/Node security scans, DeepScan, and review with both
  findings remediated and zero unresolved threads. Package-scoped JSON output
  reproduces 1,164 passing cases: 1,130 Go, 26 Rust, and eight
  maintained-client tests.
  Production inventory, firewall/TLS/DNS, live deployment, and GH-29 rollout
  approval remain unexecuted. Public status PR #91 merged as `69e498f`; its
  post-merge Security Scan and Pages deployment pass, the live page publishes
  1,164 cases with unchanged rollout arithmetic, and GH-89 is closed.
- GH-85 is closed through merged PR #86 (`cd44fec`). The supported private
  observability stack now provides an immutable 16-panel Grafana dashboard,
  eleven actionable and deterministically tested Prometheus rules,
  recovery/testnet objectives, role-based escalation ownership, and CI runtime
  proof of every required query. Fresh tests against merged implementation
  `cd44fec` record 1,105 cases: 1,071 Go, 26 Rust, and eight maintained-client
  tests; PR #87 published that status on `main` as `0aaba4b`. External paging,
  production SLOs, and deployment remain outside this evidence.
- GH-80 is closed through merged PR #81 (`e629374`). The supported node now
  exposes a private two-source CometBFT plus SDK/application metrics baseline,
  fixed-cardinality TrueRepublic EndBlock/invariant/supply signals, exact
  start/restart/disable behavior, and fail-closed public-proxy denial.
  Merged `main` records 1,104 cases: 1,070 Go, 26 Rust, and eight
  maintained-client tests.
- GH-77 is closed through merged PR #78 (`133fb3b`). Supported daemon
  operation now enforces structured JSON at the central SDK/CometBFT logger,
  defensively redacts reviewed credential/key/private-transaction patterns,
  rejects raw KV trace stores, and validates normal native/container output.
  Merged `main` records 1,092 cases: 1,058 Go, 26 Rust, and eight
  maintained-client tests.
- GH-74 is closed through merged PR #75 (`468a43a`). It adds dependency-free
  `healthcheck live|ready` commands with
  literal-loopback-only RPC targets, bounded time/body handling, no environment
  proxy or redirects, strict JSON-RPC validation, and distinct restart versus
  synchronization semantics. Docker, CI, operator guidance, and repository
  policy tests use the same commands. Merged `main` records 1,043 cases:
  1,009 Go, 26 Rust, and eight maintained-client tests. Exact final head
  `e021044` passed all technical gates; all six CodeRabbit threads were
  remediated and resolved.
- GH-71 merged a deterministic read-only network-policy command for seed,
  sentry, validator, RPC/API, and private roles; fail-closed startup integration;
  canonical peer/topology checks; loopback-only client, profiling, and metrics
  listeners; explicit public P2P interfaces; safe Docker/nginx defaults; and
  operator firewall/rate-limit guidance. Its merged `main` baseline recorded
  986 cases: 952 Go, 26 Rust, and eight maintained-client tests. Production
  topology, firewall, DNS, server, and deployment actions remain unexecuted.
- GH-14 local documentation consistency script: PASS.
- GH-14 local Rust workspace: 26 tests PASS; Clippy PASS.
- GH-14 local Rust audit: no blocking advisory; six allowed transitive
  dev-tooling warnings remain.
- GH-14 local v0.4 client: reproducible `npm ci`; npm audit reports zero
  vulnerabilities after upgrades.
- GH-14 local v0.4 client: `npm ci`, lint, six regression tests, production build, and
  `npm audit` all PASS. Main bundle is 1.68 MB before gzip (performance warning).
- The pre-GH-32 `main` baseline was 684: 650 Go, 26 Rust, and eight
  maintained-client tests. Four focused legacy-web ZKP regressions pass
  separately and are not included in that authoritative total. The prior 577
  figure is retained only as historical.
- The pre-GH-71 main count is 853: 819 Go, 26 Rust, and eight maintained-client
  tests. Separately gated process harnesses are tracked only after they pass
  explicitly and are never added to that arithmetic total. The latest
  hardened four-validator run requires new post-rejoin blocks and passed in
  68.90 seconds. Full Go race/coverage passes with root/application coverage at
  65.9% on PR #40.
- GH-56 final-head CI proves the complete combined multi-validator process
  matrix in 9m39s. Its rotation slice stops the old signer, authenticates the
  operator transaction, applies old-key zero and replacement activation at
  H+2, preserves validator claims and app-hash convergence, rejects revoked-key
  reuse, and round-trips export/import.
- GH-32 uses four independently generated CometBFT Ed25519 keys, one identical
  bank-backed PoD genesis, explicit localhost persistent peers, common-height
  app-hash checks, one-validator failure with continued quorum, restart/catch-up,
  clean SIGINT shutdown, recovered export, and post-export ledger validation.
  Child processes and RPC requests inherit the test context so a canceled or
  timed-out test cannot orphan network work.
  Localhost address-book relaxation and duplicate-IP allowance are confined to
  temporary test configuration; production defaults are unchanged.
- GH-39 merged evidence adds custom SDK v0.50 signer resolution for hand-written
  truedemocracy Msgs, shares the configured InterfaceRegistry with BaseApp and
  tx/event paths, verifies delivered tx results through CometBFT RPC, and passes
  a gated six-node join/replacement lifecycle smoke in 117.638 seconds. Full
  `go test ./...` passes locally, and PR #40 GitHub checks are green:
  `build-and-test`, `multi-validator-recovery`, `docker-restart-smoke`, docs
  consistency, CodeRabbit, DeepScan, Go/Rust security scans, and Node audits.
- GH-41 merged evidence adds a gated four-validator network partition recovery
  harness. A 3-of-4 quorum progresses and commits a real governance transaction
  while the fourth validator is isolated with no peers; the isolated validator
  then reconnects, catches up, shares the same app hash, retains validator
  power, and exports ledger-valid state. Local targeted run passes in 104.175s;
  all gated process harnesses pass together in 392.147s; full local
  `go test ./...` passes; PR #42 GitHub checks are green.
- GH-43 merged evidence adds a gated trusted snapshot state-sync harness. Four
  trusted validators serve snapshots, a real governance transaction commits
  before sync, a fresh non-validator node derives trust height/hash from trusted
  RPC, catches up through state sync, converges on the same app hash, sees all
  validator powers, and exports ledger-valid state. Local targeted run passes
  in 130.528s; the combined CI-smoke equivalent passes in 197.835s; full local
  `go test ./...` passes in 65.114s; PR #44 GitHub checks are green.
- GH-45 merged evidence adds a gated sanitized backup/restore/export/import
  harness. A live full node is backed up without node key, validator key,
  validator signing state, or keyring material; the artifact restores into a
  freshly initialized home while preserving local keys; the restored node
  catches up, converges on app hash, exports ledger-valid state, and re-imports
  the exported genesis. Local targeted run passes in 88.224s; full local
  `go test ./...` passes in 58.843s; the combined CI-smoke equivalent passes
  in 290.498s; a focused state-sync timeout hardening recheck passes in
  127.784s; PR #46 GitHub checks are green.
- GH-53 local evidence adds a gated compatible binary upgrade and failed-
  candidate rollback harness. It uses separately versioned artifacts on four
  persisted validator homes, preserves quorum during rolling replacement,
  proves a fail-before-open candidate leaves keys and signing state byte-for-
  byte unchanged, rolls all validators back to the baseline artifact, verifies
  non-regressing signing positions, app-hash convergence and validator power,
  and validates/re-imports the non-empty exported ledger. This does not claim
  `x/upgrade` or consensus-breaking migration support.
- GH-13 local Go 1.26.5: build, vet, normal tests, race tests, and coverage PASS.
  Coverage: root 10.2%, token 93.5%, treasury 97.0%, DEX 34.2%, governance 55.8%.
- Go vulnerability gate: no reachable finding with an available fix remains;
  four upstream `N/A` findings are tracked for import-path reduction.
- Former `web-wallet`: locally retired and removed under GH-112 after the
  final baseline reproduced 70 advisories, 18 passing but API-isolated tests, a
  warning-heavy build, broken custom-query calls, obsolete/unregistered custom
  transaction paths, and a real bank-send path. Git history is audit-only.
- Former `mobile-wallet`: retired and removed under GH-102 after the final
  baseline reproduced 51 advisories (7 low, 16 moderate, 24 high, 4 critical),
  no tests, a broken Android bundle, obsolete chain queries, and unsafe
  mnemonic-in-UI handling. Git history is audit-only; there is no supported
  native mobile client.
- Public README, status JSON, limitations, and GitHub Pages source now display
  an active recovery warning and link to GH-4.
- Public GitHub Pages is configured from `main:/docs`. The latest source update
  records the 726-case and validator-lifecycle evidence, recovery/non-production
  warning, and 21M cap.
- Canonical `client-web` now has dedicated GitHub install/lint/test/build/audit
  gates; legacy client audits remain informational during migration.
- PR #9's final GitHub checks passed before the ordered recovery chain was
  merged to `main`: Go CI, Rust CI, Client Web CI, documentation consistency,
  govulncheck, Rust audit, canonical npm audit, and informational legacy audits.
- Both Debian/glibc Docker builds pass with the architecture-specific wasmvm
  shared library; the module path is resolved dynamically from Go metadata.
- Codex merge audit: conditional approval with 0 FAIL / 3 WARN / 7 PASS.
- GitHub branch protection currently requests one approval but defines no
  required status-check contexts and does not enforce the rule for admins.
  Project workflow therefore continues to require final-head CI evidence and a
  reviewable PR even when GitHub would technically allow an administrator merge.
- CodeRabbit review remediation passes locally and on GitHub: checkout credentials are
  disabled, security workflow permissions are read-only, canonical client CI
  uses Node 22, current Go security releases are applied, and the full local
  and GitHub verification matrices are green.
- Final GH-11 audit found and fixed validator-stake and gas-price scaling gaps,
  plus conflicting legacy metadata cleanup. Inherited PR #15 checks are green
  at head `e0ff339`; see `PR15_AUDIT.md`.
- GH-14 backs domain treasury and validator stake claims with exact bank escrow,
  uses cached atomic settlement, binds claimed identities to authenticated
  signers across CLI and CosmWasm paths, and burns validator slash penalties.
  GH-14 local Go build/vet/race/coverage and 557 Go cases pass; Rust and
  maintained-client gates pass locally. PR #16 GitHub Go/Rust/client/docs/
  Docker/DeepScan/CodeRabbit checks are green; see `PR16_AUDIT.md`.
- GH-13 derives reward decay from canonical bank supply, clips aggregate mints
  at the 21M cap, backs validator/domain claims with exact module mints, routes
  slash burns through the same service, and commits both inflation phases under
  one EndBlock cache. Full local Go/Rust/client/docs gates pass. Its Dockerfile
  now maps Docker target architecture to the correct wasmvm library, verifies
  runtime linkage during image construction, and excludes 1.5+ GB of local
  build artifacts/dependencies from the context. The image build and
  CLI startup are proven by both GitHub Docker jobs. PR #17 was merged; both
  Go jobs, docs, DeepScan, the manual security matrix, and the prior full
  CodeRabbit review completed with five inline and two additional findings.
  Rollback-aware mock-bank evidence, restored payout snapshot baselines,
  container version smoke, and documentation corrections pass locally and on
  GitHub. Both Go/Docker jobs, docs, DeepScan, CodeRabbit, and the manual
  security workflow are green; all five review threads are resolved. See
  `PR17_AUDIT.md`.
- GH-10 is rebased onto final PR #17 and moves every public DEX reserve through
  exact module-bank custody. Provider-indexed LP shares gate withdrawals,
  direct and cross-asset swaps settle atomically, PNYX burns reduce canonical
  supply, and registry/status mutation requires chain authority. Length-prefixed
  LP keys prevent valid denom-prefix collisions. Local Go build/vet/578 tests/
  race, Rust 26 tests/audit, maintained-client install/lint/6 tests/build/audit,
  CLI smoke, module verification, and docs/diff checks pass. GitHub docs,
  DeepScan, Go build/vet/race/coverage, and Docker pass at `3234741`; manual
  Security Scan run `29156922464` passes all five jobs. CodeRabbit is
  rate-limited and substantive external review remains pending; see
  `PR18_AUDIT.md`.
- GH-12 is rebased onto final PR #18 and validates all custom genesis before
  mutation, reconciles complete module bank balances, exports provider LP
  ownership, preserves non-empty custody across export/import, and checks cap,
  escrow, reserves, and LP totals every block. Audit remediation removed a
  publicly derivable bootstrap-validator secret and now bootstraps only from
  real CometBFT Ed25519 public keys with exact cap-checked stake. Local Go
  build/vet/615 cases/race/coverage, Rust 26 tests/audit, maintained-client
  install/lint/6 tests/build/audit, CLI smoke, module integrity, and docs/diff
  checks pass; see `PR19_AUDIT.md`.
- GH-12 GitHub Docs, DeepScan, Web, Mobile, Rust, Go, both Docker jobs, and
  refreshed Security Scan `29172007410` are green. Both actionable review
  threads are answered/resolved at head `eec91c7`.
- GH-20 is rebased onto final PR #19. Proofs bind versioned chain/proposal/rating
  signals while one-vote nullifiers remain rating-independent and chain-scoped.
  Random trusted setup is removed from consensus. Genesis pins circuit ID, VK
  SHA-256, BN254/public-input shape, and canonical bytes; recomputes identity
  roots; and round-trips exact active nullifiers without undoing Big Purges.
  Both web clients now reject mock proof submission. Local Go build/vet/643
  cases/race/coverage, Rust 26 tests/audit, maintained-client lint/46 tests/build/
  audit, four focused legacy tests/build/audit, module integrity, and diff
  checks pass; see `PR22_AUDIT.md`.
- GH-21 was rebased without implementation drift onto PR #22 head `0c72ad0`.
  Standard Cosmos/Comet lifecycle now uses the configured persistent database
  and home; `init` binds the generated CometBFT key to exactly bank-backed PoD
  genesis and refuses conflicting validator sets. Native block production,
  SIGINT shutdown, same-home restart, height advancement, invariants, export,
  649 Go cases, targeted race, vet, build, CLI version, shell syntax, and diff
  checks pass locally. Root coverage is 64.3%. GitHub Go/Docker run
  `29172166826`, Docs, DeepScan, Web, CodeRabbit, and Security Scan
  `29172246373` passed before PR #23 merged; see `PR23_AUDIT.md`.
- GH-8 was rebased onto final GH-21 `49938a3`. It modernizes official
  Action runtimes without credential persistence or duplicate feature runs,
  strengthens suite/module/cap consistency, and reconciles CLAUDE, install,
  FAQ, landing, and real wiki status/security claims to 684 cases. Workflow
  YAML, docs, JSON, wiki target, stale-current-claim, and diff checks passed.
  GitHub Go/Docker, Rust, Web, Mobile, Docs, DeepScan, CodeRabbit, and all five
  Security Scan `29172246235` jobs passed before PR #24 merged. See
  `PR24_AUDIT.md`.
- GH-26 removes the last public `x/staking` bootstrap footgun. The operator
  wrapper now invokes only daemon `init`; its regression and a real compiled
  init prove generated-key, exact bank-backed PoD genesis without mnemonic,
  account, gentx, or extra-supply side effects. Full Go/vet/docs/shell gates
  pass locally. Rebased PR #27 passed GitHub Go/Docker run `29190764808`,
  Docs/Pages run `29190763221`, Security run `29190764842`, DeepScan, and
  CodeRabbit before squash merge `513716c`. See `PR27_AUDIT.md`.

## Public-status warning

`docs/status.json`, README, limitations, and the landing page now mark recovery
as active and separate 1,388 verified cases from the historical 577 figure.
`CLAUDE.md`, install guidance, FAQ, landing page, wiki, and the root audit are
reconciled with the merged recovery foundation while retaining the explicit
non-production boundary.

## Blocking audit result

The current recovery-foundation audit records 0 FAIL / 2 WARN / 18 PASS across
denomination/cap, governance custody, reward issuance, DEX custody, custom
genesis, runtime invariants, ZKP boundaries, maintained-client safety, and node
lifecycle. The repository remains recovery-only because GH-20 still needs a
real prover and external cryptographic review. GH-32/GH-41/GH-43/GH-45/GH-53/GH-55
close bounded four-validator failure/restart/catch-up, partition-recovery,
trusted state-sync, sanitized backup/restore/export/import, compatible binary
replacement, fail-before-open rollback, and validator identity cold-failover
slices. Consensus-breaking state migration, partially applied migration
recovery, paging drills, load/capacity evidence, live topology deployment,
IBC, and independent operations review remain open. Authenticated
consensus-key rotation, compromised-key eviction, and deterministic network
policy are closed by GH-56 and GH-71.

GH-11 implements the canonical denomination metadata (`upnyx`, six decimal
places, 21,000,000,000,000 base-unit cap) and pre-init bank-genesis cap checks.
Its final audit corrections and PR #15 checks passed before the ordered merge
to `main`.
GH-14 closes the declared treasury/stake custody slice on `main`.
GH-13 closes cap-checked reward issuance, GH-10 closes DEX custody/LP/burn/
authority, and GH-12 closes custom-genesis/runtime-invariant findings on
`main`. GH-20 closes the on-chain ZKP binding and mock-client safety
implementation on `main`. GH-21 closes the native single-node persistence/
restart implementation locally and in GitHub CI. These remediations are merged
to `main`.
