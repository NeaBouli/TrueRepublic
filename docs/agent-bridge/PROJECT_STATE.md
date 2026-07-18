# Project State

Updated: 2026-07-19 00:54 EEST

## Repository

- GitHub: `NeaBouli/TrueRepublic`
- Baseline: canonical `origin/main`; exact implementation and evidence commits
  are recorded in `ACTION_LOG.md` so this live state does not self-expire after
  documentation-only merges.
- Merged recovery PRs: #9, #15, #16, #17, #18, #19, #22, #23, #24, #27,
  #28, #30, #31, #33, #34, #35, and #40.
- Current work: GH-29 remains open as the rollout execution tracker. GH-32 and
  PR #33 close its first Phase 1 gate with local and GitHub evidence. GH-39 is
  now merged via PR #40 with green GitHub CI for validator
  join/replacement/restart-catch-up evidence plus Keeper/ABCI power-zero leave
  coverage. GH-41 is the active Phase 1 task for network partitions, delayed
  peers, validator failure, and ledger-safe recovery.
- Active recovery checkout:
  `/Users/gio/Documents/Codex/2026-07-11/erkunden/TrueRepublic-gh20`
- GH-26 branch: `fix/GH-26-pod-init-script`
- GH-26 issue: #26; PR #27 is verified and merged to `main`.
- GH-26 recovery checkout:
  `/Users/gio/Documents/Codex/2026-07-11/erkunden/TrueRepublic-gh26`
- Recovery worktree: `/Users/gio/Desktop/repos/TrueRepublic-recovery`
- Legacy local checkout: preserved at `/Users/gio/Desktop/repos/TrueRepublic`
- GitHub epic: #4
- Open GitHub issue set after cleanup: #4 recovery epic, #7 audit/review
  parent, #29 rollout tracker, and #41 active network-failure child task.

## Verified state

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
- Current `main` count is 689: 655 Go, 26 Rust, and eight maintained-client
  tests. The separately gated four-validator and six-node validator lifecycle
  process harnesses are not added again to that arithmetic total. The latest
  hardened four-validator run requires new post-rejoin blocks and passed in
  68.90 seconds. Full Go race/coverage passes with root/application coverage at
  65.9% on PR #40.
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
- GH-13 local Go 1.26.5: build, vet, normal tests, race tests, and coverage PASS.
  Coverage: root 10.2%, token 93.5%, treasury 97.0%, DEX 34.2%, governance 55.8%.
- Go vulnerability gate: no reachable finding with an available fix remains;
  four upstream `N/A` findings are tracked for import-path reduction.
- Legacy `web-wallet`: focused ZKP tests, build, and current npm audit pass, but
  obsolete CosmJS crypto/Create React App and source-map warnings remain; mock
  proof submission is disabled and it is not approved for keys or funds.
- Legacy `mobile-wallet`: no tests exist and 51 advisories remain (22 high,
  3 critical); not approved for keys or funds.
- Public README, status JSON, limitations, and GitHub Pages source now display
  an active recovery warning and link to GH-4.
- Public GitHub Pages is configured from `main:/docs`. The latest source update
  records the 689-case and validator-lifecycle evidence, recovery/non-production
  warning, and 21M cap.
- Canonical `client-web` now has dedicated GitHub install/lint/test/build/audit
  gates; legacy client audits remain informational during migration.
- PR #9 GitHub checks are all green: Go CI, Rust CI, Client Web CI,
  documentation consistency, govulncheck, Rust audit, canonical npm audit, and
  informational legacy npm audits.
- Both Debian/glibc Docker builds pass with the architecture-specific wasmvm
  shared library; the module path is resolved dynamically from Go metadata.
- Codex merge audit: conditional approval with 0 FAIL / 3 WARN / 7 PASS.
- GitHub branch protection requires one approval; PR #9 currently has none and
  must not be merged through the administrator bypass.
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
  CLI startup are proven by both GitHub Docker jobs. PR #17 is mergeable; both
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
  cases/race/coverage, Rust 26 tests/audit, maintained-client lint/8 tests/build/
  audit, four focused legacy tests/build/audit, module integrity, and diff checks
  pass; see `PR22_AUDIT.md`.
- GH-21 is rebased without implementation drift onto PR #22 head `0c72ad0`.
  Standard Cosmos/Comet lifecycle now uses the configured persistent database
  and home; `init` binds the generated CometBFT key to exactly bank-backed PoD
  genesis and refuses conflicting validator sets. Native block production,
  SIGINT shutdown, same-home restart, height advancement, invariants, export,
  649 Go cases, targeted race, vet, build, CLI version, shell syntax, and diff
  checks pass locally. Root coverage is 64.3%. Published head `49938a3` is
  mergeable; GitHub Go/Docker run `29172166826`, Docs, DeepScan, Web,
  CodeRabbit, and Security Scan `29172246373` pass; see `PR23_AUDIT.md`.
- GH-8 is rebased onto final GH-21 `49938a3`. It modernizes official
  Action runtimes without credential persistence or duplicate feature runs,
  strengthens suite/module/cap consistency, and reconciles CLAUDE, install,
  FAQ, landing, and real wiki status/security claims to 684 cases. Workflow
  YAML, docs, JSON, wiki target, stale-current-claim, and diff checks pass;
  Published stack head is mergeable. GitHub Go/Docker, Rust, Web, Mobile,
  Docs, DeepScan, CodeRabbit, and all five Security Scan `29172246235` jobs
  pass. See `PR24_AUDIT.md`.
- GH-26 removes the last public `x/staking` bootstrap footgun. The operator
  wrapper now invokes only daemon `init`; its regression and a real compiled
  init prove generated-key, exact bank-backed PoD genesis without mnemonic,
  account, gentx, or extra-supply side effects. Full Go/vet/docs/shell gates
  pass locally. Rebased PR #27 passed GitHub Go/Docker run `29190764808`,
  Docs/Pages run `29190763221`, Security run `29190764842`, DeepScan, and
  CodeRabbit before squash merge `513716c`. See `PR27_AUDIT.md`.

## Public-status warning

`docs/status.json`, README, limitations, and the landing page now mark recovery
as active and separate 684 verified tests from the historical 577 figure.
`CLAUDE.md`, install guidance, FAQ, landing page, and wiki are reconciled on
PR #24 and are visible on `main`.

## Blocking audit result

The token/ledger audit is 12/12 PASS locally across the merged recovery work:
denomination/cap, governance custody, reward issuance, DEX custody, custom
genesis, and runtime invariants. The repository remains recovery-only because
GH-20 still needs a real prover/external
cryptographic review. GH-32 closes only the bounded four-validator
failure/restart/catch-up slice; partitions, state sync, backup/restore,
upgrades, load, topology, IBC, and independent operations review remain open.

GH-11 implements the canonical denomination metadata (`upnyx`, six decimal
places, 21,000,000,000,000 base-unit cap) and pre-init bank-genesis cap checks.
Its final audit corrections are locally verified and rebased onto PR #9 head
`acfc3d5`; refreshed PR #15 GitHub checks are green at head `e0ff339`.
GH-14 closes the declared treasury/stake custody slice on `main`.
GH-13 closes cap-checked reward issuance, GH-10 closes DEX custody/LP/burn/
authority, and GH-12 closes custom-genesis/runtime-invariant findings locally.
GH-20 closes the on-chain ZKP binding and mock-client safety implementation
locally. GH-21 closes the native single-node persistence/restart implementation
locally and in GitHub CI. These remediations are merged to `main`.
