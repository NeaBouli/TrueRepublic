# Recovery Queue

## P0 - security and reproducibility

- [ ] GH-4: keep the recovery epic and acceptance criteria current.
- [x] GH-5: Go/Rust toolchains, tests, static checks, vulnerability gates, and
  GitHub security CI are green.
- [x] GH-6: v0.4 client lint, tests, build, exact amount handling, maintained
  wallet crypto, npm audit, and GitHub CI are green.
- [x] GH-8: reproduce legacy web wallet and mobile wallet CI/security state.

## P1 - consensus and wallet audit

- [x] GH-7: audit PNYX 21M cap, denomination, and ledger conservation paths.
- [x] GH-11: denomination/cap branch was locally and GitHub verified and merged
  via PR #15.
- [x] GH-14: PR #16 was locally/GitHub green with zero unresolved review
  threads and is merged.
- [x] GH-13: PR #17 final review remediation is locally/GitHub green, both
  Docker builds and security pass, all five review threads are resolved, and
  the PR is merged.
- [x] GH-10: DEX bank custody, provider LP ownership, canonical burns, registry
  authority, and rollback evidence are locally green on stacked PR #18.
- [x] GH-12: exact custom genesis, non-empty custody round trip, and registered
  supply/escrow/reserve/LP invariants are locally green on stacked PR #19.
- [x] GH-20: bind ZKP proofs/signatures to chain and vote context, pin genesis
  VK identity, preserve active nullifiers, validate canonical fields, and make
  mock clients non-submittable on stacked PR #22.
- [ ] GH-20: obtain independent cryptographic review and deliver a compatible
  real prover/ceremony artifact before enabling anonymous submission.
- [x] GH-21: replace the MemDB/`select {}` placeholder and legacy `x/staking`
  bootstrap with persistent Cosmos/Comet lifecycle and generated-key,
  bank-backed PoD genesis; native restart/export and local gates pass.
- [x] GH-26: make the operator init wrapper delegate only to the supported PoD
  daemon init; remove mnemonic/account/gentx side effects and add regression
  coverage.
- [ ] GH-26: complete full local/GitHub/Docker/security verification and keep
  the stacked PR behind final PR #24.
- [x] GH-21: publish audited head `ec1ce17`; refreshed Docker restart, Go
  race/coverage, docs, static, and Security Scan gates are green.
- [ ] GH-21: obtain independent multi-node, backup/restore, IBC/upgrade
  operations review before public-network approval.
- [x] GH-7: DEX rounding, slippage, pool accounting, custody, and authorization
  audit completed in GH-10; GH-12 retains genesis/runtime invariants.

## P2 - delivery

- [x] GH-8: review the preserved legacy checkout; no code is safe/useful for
  wholesale merge, and the checkout remains preserved pending final archive.
- [x] GH-8: align CLAUDE, install, FAQ, README, status JSON, limitations,
  landing page, and real wiki status/security claims to the 684-test recovery
  source of truth.
- [x] GH-8: enforce suite/module/cap arithmetic and real wiki/agent/public
  status through docs CI; modernize Action runtimes without credential or
  duplicate-run regression.
- [x] GH-8: publish audited head `3964f4a`; every updated GitHub Action and
  manual Security Scan `29171476126` passes.
- [ ] GH-8: complete refreshed docs/recovery CI and merge PR #24, then rebase
  PR #27 and replace the obsolete PR #25 visibility track.
- [x] PRs #9, #15, #16, #17, #18, #19, #22, and #23 are merged to `main`.
