# Project State

Updated: 2026-07-11 21:25 EEST

## Repository

- GitHub: `NeaBouli/TrueRepublic`
- Baseline: `origin/main` at `d8545cf`
- Recovery branch: `fix/GH-4-recovery-foundation`
- Ready PR: #9 (`fix/GH-4-recovery-foundation` -> `main`)
- Recovery worktree: `/Users/gio/Desktop/repos/TrueRepublic-recovery`
- Legacy local checkout: preserved at `/Users/gio/Desktop/repos/TrueRepublic`
- GitHub epic: #4

## Verified state

- Documentation consistency script: PASS.
- Rust workspace: 26 tests PASS; Clippy PASS.
- Rust audit: fixable vulnerabilities removed; six transitive dev-tooling warnings remain.
- v0.4 client: reproducible `npm ci`; npm audit reports zero vulnerabilities after upgrades.
- v0.4 client: `npm ci`, lint, five regression tests, production build, and
  `npm audit` all PASS. Main bundle is 1.68 MB before gzip (performance warning).
- Current recovery-verified test count is 564: 533 Go, 26 Rust, and five
  maintained-client tests. The prior 577 figure is retained only as historical.
- Go 1.26.5: build, race tests, and vet PASS after dependency/toolchain recovery.
  Coverage: root 5.8%, treasury 97.0%, DEX 34.2%, governance 53.5%.
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
- CodeRabbit review remediation passes locally: checkout credentials are
  disabled, security workflow permissions are read-only, canonical client CI
  uses Node 22, current Go security releases are applied, and the full local
  verification matrix remains green. Refreshed GitHub checks are pending push.

## Public-status warning

`docs/status.json`, README, limitations, and the landing page now mark recovery
as active and separate 564 verified tests from the historical 577 figure.
`CLAUDE.md` still needs reconciliation.

## Blocking audit result

The first GH-7 token/ledger slice is FAIL: the 21M cap is not enforced against
bank supply, six-decimal client semantics conflict with chain denominations,
and treasury/stake/DEX/reward ledgers are not consistently bank-backed. The
repository remains simulation/recovery-only until `CODEX_AUDIT.md` blockers close.
