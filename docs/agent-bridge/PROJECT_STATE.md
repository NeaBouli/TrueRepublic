# Project State

Updated: 2026-07-11 22:08 EEST

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
- Recovery worktree: `/Users/gio/Desktop/repos/TrueRepublic-recovery`
- Legacy local checkout: preserved at `/Users/gio/Desktop/repos/TrueRepublic`
- GitHub epic: #4

## Verified state

- Documentation consistency script: PASS.
- Rust workspace: 26 tests PASS; Clippy PASS.
- Rust audit: fixable vulnerabilities removed; six transitive dev-tooling warnings remain.
- v0.4 client: reproducible `npm ci`; npm audit reports zero vulnerabilities after upgrades.
- v0.4 client: `npm ci`, lint, six regression tests, production build, and
  `npm audit` all PASS. Main bundle is 1.68 MB before gzip (performance warning).
- Current GH-14 branch test count is 589: 557 Go, 26 Rust, and six
  maintained-client tests. The prior 577 figure is retained only as historical.
- Go 1.26.5: build, race tests, and vet PASS on GH-11.
  Coverage: root 5.5%, token 88.5%, treasury 97.0%, DEX 34.2%, governance 53.5%.
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
  plus conflicting legacy metadata cleanup. See `PR15_AUDIT.md`.
- GH-14 backs domain treasury and validator stake claims with exact bank escrow,
  uses cached atomic settlement, binds claimed identities to authenticated
  signers across CLI and CosmWasm paths, and burns validator slash penalties.
  Local Go build/vet/race/coverage and 557 Go cases pass; Rust and maintained
  client gates remain green; see `PR16_AUDIT.md`.

## Public-status warning

`docs/status.json`, README, limitations, and the landing page now mark recovery
as active and separate 589 verified tests from the historical 577 figure.
`CLAUDE.md` still needs reconciliation.

## Blocking audit result

The first GH-7 token/ledger slice is FAIL: the 21M cap is not enforced against
bank supply, six-decimal client semantics conflict with chain denominations,
and treasury/stake/DEX/reward ledgers are not consistently bank-backed. The
repository remains simulation/recovery-only until `CODEX_AUDIT.md` blockers close.

GH-11 implements the canonical denomination metadata (`upnyx`, six decimal
places, 21,000,000,000,000 base-unit cap) and pre-init bank-genesis cap checks.
Its final audit corrections are locally verified and rebased onto PR #9 head
`acfc3d5`; refreshed PR #15 GitHub checks are required before readiness.
GH-14 closes the declared treasury/stake custody slice on its stacked branch.
These remediations remain unmerged and do not close reward issuance, DEX
custody, or custom-genesis/runtime-invariant blockers.
