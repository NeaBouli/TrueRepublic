# Project State

Updated: 2026-07-12 04:32 EEST

## Repository

- GitHub: `NeaBouli/TrueRepublic`
- Baseline: `origin/main` at `d8545cf`
- Recovery branch: `fix/GH-4-recovery-foundation`
- Ready PR: #9 (`fix/GH-4-recovery-foundation` -> `main`), head `acfc3d5`
- Stacked implementation branch: `fix/GH-11-pnyx-cap`
- Stacked draft PR: #15 (`fix/GH-11-pnyx-cap` -> `fix/GH-4-recovery-foundation`)
- Stacked worktree: `/Users/gio/Desktop/repos/TrueRepublic-gh11`
- GH-14 branch: `fix/GH-14-bank-escrow`
- GH-14 draft PR: #16 (`fix/GH-14-bank-escrow` -> `fix/GH-11-pnyx-cap`)
- GH-14 worktree: `/Users/gio/Desktop/repos/TrueRepublic-gh14`
- GH-13 branch: `fix/GH-13-cap-issuance`
- GH-13 draft PR: #17 (`fix/GH-13-cap-issuance` -> `fix/GH-14-bank-escrow`)
- GH-13 worktree: `/Users/gio/Desktop/repos/TrueRepublic-gh13`
- GH-10 branch: `fix/GH-10-dex-custody`
- GH-10 draft PR: #18 (`fix/GH-10-dex-custody` -> `fix/GH-13-cap-issuance`)
- GH-10 recovery checkout:
  `/Users/gio/Documents/Codex/2026-07-11/erkunden/TrueRepublic-gh10`
- GH-12 branch: `fix/GH-12-genesis-invariants`
- GH-12 draft PR: #19 (`fix/GH-12-genesis-invariants` -> `fix/GH-10-dex-custody`)
- GH-12 recovery checkout:
  `/Users/gio/Documents/Codex/2026-07-11/erkunden/TrueRepublic-gh12`
- Recovery worktree: `/Users/gio/Desktop/repos/TrueRepublic-recovery`
- Legacy local checkout: preserved at `/Users/gio/Desktop/repos/TrueRepublic`
- GitHub epic: #4

## Verified state

- GH-14 local documentation consistency script: PASS.
- GH-14 local Rust workspace: 26 tests PASS; Clippy PASS.
- GH-14 local Rust audit: no blocking advisory; six allowed transitive
  dev-tooling warnings remain.
- GH-14 local v0.4 client: reproducible `npm ci`; npm audit reports zero
  vulnerabilities after upgrades.
- GH-14 local v0.4 client: `npm ci`, lint, six regression tests, production build, and
  `npm audit` all PASS. Main bundle is 1.68 MB before gzip (performance warning).
- Current GH-12 branch test count is 647: 615 Go, 26 Rust, and six
  maintained-client tests. The prior 577 figure is retained only as historical.
- GH-13 local Go 1.26.5: build, vet, normal tests, race tests, and coverage PASS.
  Coverage: root 10.2%, token 93.5%, treasury 97.0%, DEX 34.2%, governance 55.8%.
- Go vulnerability gate: no reachable finding with an available fix remains;
  four upstream `N/A` findings are tracked for import-path reduction.
- Legacy `web-wallet`: build/test command reaches audit, but 68 advisories remain
  (26 high, 2 critical); not approved for keys or funds.
- Legacy `mobile-wallet`: no tests exist and 51 advisories remain (22 high,
  3 critical); not approved for keys or funds.
- Public README, status JSON, limitations, and GitHub Pages source now display
  an active recovery warning and link to GH-4.
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

## Public-status warning

`docs/status.json`, README, limitations, and the landing page now mark recovery
as active and separate 647 verified tests from the historical 577 figure.
`CLAUDE.md` still needs reconciliation.

## Blocking audit result

The GH-7 token/ledger audit is 12/12 PASS locally across the ordered stack:
denomination/cap, governance custody, reward issuance, DEX custody, custom
genesis, and runtime invariants. The repository remains recovery-only because
the stack is unmerged/unreviewed and GH-20 ZKP plus GH-21 node-lifecycle work
remain outside this ledger slice.

GH-11 implements the canonical denomination metadata (`upnyx`, six decimal
places, 21,000,000,000,000 base-unit cap) and pre-init bank-genesis cap checks.
Its final audit corrections are locally verified and rebased onto PR #9 head
`acfc3d5`; refreshed PR #15 GitHub checks are green at head `e0ff339`.
GH-14 closes the declared treasury/stake custody slice on its stacked branch.
GH-13 closes cap-checked reward issuance, GH-10 closes DEX custody/LP/burn/
authority, and GH-12 closes custom-genesis/runtime-invariant findings locally.
These remediations remain stacked and unmerged.
