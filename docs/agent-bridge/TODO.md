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
- [x] GH-11: denomination/cap branch is locally and GitHub verified in PR #15;
  keep it stacked until PR #9 receives independent approval.
- [x] GH-14: PR #16 is rebased, mergeable, locally/GitHub green, and has zero
  unresolved review threads; keep it stacked until its bases merge.
- [ ] GH-13: issuance implementation, audit hardening, and the full local
  polyglot matrix pass; republish PR #17 and complete Docker/CI review.
- [ ] GH-10: implement DEX bank custody, provider LP ownership, and real burns.
- [ ] GH-12: validate genesis and add ledger conservation invariants.
- [ ] GH-7: audit ZKP/nullifier/domain-key authentication and client-side mock boundaries.
- [ ] GH-7: audit DEX rounding, slippage, pool accounting, and authorization.

## P2 - delivery

- [x] GH-8: review the preserved legacy checkout; no code is safe/useful for
  wholesale merge, and the checkout remains preserved pending final archive.
- [ ] GH-8: align README, CLAUDE.md, status JSON, limitations, website, and test counts.
- [ ] GH-8: add bridge/docs consistency to CI.
- [x] PR #9 is ready, mergeable, fully green, and awaiting the required
  independent approval.
