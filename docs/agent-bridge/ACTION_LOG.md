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
