# Action Log

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
