# Recovery Queue

## P0 - security and reproducibility

- [ ] GH-4: keep the recovery epic and acceptance criteria current.
- [ ] GH-29: complete the seven-phase production-readiness roadmap and attach
  evidence for every rollout exit gate before any public-network launch.
- [x] GH-29: reopen the issue as the execution tracker; PR #31 completed only
  the roadmap handoff, not the rollout phases.
- [x] GH-5: Go/Rust toolchains, tests, static checks, vulnerability gates, and
  GitHub security CI are green.
- [x] GH-47: bound the Go CI `build-and-test` job with a 20-minute timeout so
  stuck GitHub runners cannot leave `main` indefinitely in progress.
- [x] GH-48: reconcile live post-merge contributor/operator/audit status,
  publish the verified fast-audit evidence, and close after final-head GitHub
  checks pass.
- [x] GH-50: update `golang.org/x/text` to v0.39.0 for reachable GO-2026-5970,
  rerun full Go verification, and close only after final-head Security Scan is
  green.
- [x] GH-51: isolate root Go package selection from installed frontend
  `node_modules` trees and align local/CI verification commands.
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
  authority, and rollback evidence are verified and merged via PR #18.
- [x] GH-12: exact custom genesis, non-empty custody round trip, and registered
  supply/escrow/reserve/LP invariants are verified and merged via PR #19.
- [x] GH-20: bind ZKP proofs/signatures to chain and vote context, pin genesis
  VK identity, preserve active nullifiers, validate canonical fields, and make
  mock clients non-submittable in merged PR #22.
- [ ] GH-20: obtain independent cryptographic review and deliver a compatible
  real prover/ceremony artifact before enabling anonymous submission.
- [x] GH-21: replace the MemDB/`select {}` placeholder and legacy `x/staking`
  bootstrap with persistent Cosmos/Comet lifecycle and generated-key,
  bank-backed PoD genesis; native restart/export and local gates pass.
- [x] GH-26: make the operator init wrapper delegate only to the supported PoD
  daemon init; remove mnemonic/account/gentx side effects and add regression
  coverage.
- [x] GH-26: rebase PR #27 onto the verified PR #24 merge.
- [x] GH-26: GitHub Go/Docker/Docs/static/security verification is green; PR
  #27 was merged to `main` with zero unresolved review threads.
- [x] GH-21: publish audited head `ec1ce17`; refreshed Docker restart, Go
  race/coverage, docs, static, and Security Scan gates are green.
- [ ] GH-21: obtain independent multi-node, IBC/upgrade, rollback, and
  release-operations review before public-network approval.
- [x] GH-32: build and locally verify the four-validator bank-backed PoD
  consensus/failure/restart/catch-up/export harness.
- [x] GH-32: publish the implementation, obtain green Go multi-validator,
  Docker, docs, static-analysis, and security gates, then merge and close via
  PR #33 / merge `9d68a6f`.
- [x] GH-39: locally verify validator join/replacement lifecycle evidence with
  a gated six-node process harness, full-node catch-up, delivered tx checks,
  and Keeper/ABCI power-zero removal regression coverage.
- [x] GH-39: publish the branch, obtain green GitHub checks/review, merge via
  PR #40 / `ad30d188c7956c28cff5bf53304bc04848ba569a`, and update GH-29/GH-39
  closure evidence.
- [x] GH-41: locally verify network partitions, delayed peers, validator
  failure, and recovery without ledger divergence.
- [x] GH-41: publish PR, obtain green GitHub checks/review, merge via PR #42 /
  `8544943dd6fab483884392f1f04e83acbeb8f3f7`, and update GH-29/GH-41 closure
  evidence.
- [x] GH-43: locally verify trusted snapshot state sync catch-up with derived
  trust height/hash, app-hash convergence, validator-power visibility, and
  exported-ledger validation.
- [x] GH-43: publish PR, obtain green GitHub checks/review, merge via PR #44 /
  `12a37339e9cff957d1b44413aa36160aed4e8d29`, close GH-43, and update
  GH-29/roadmap closure evidence.
- [x] GH-45: locally verify sanitized backup, fresh-home restore, restored
  catch-up, app-hash convergence, export, ledger validation, and re-import.
- [x] GH-45: publish PR, obtain green GitHub checks/review, merge via PR #46 /
  `26bf44b7933c25f379db475fd34d2cfb8e49c626`, close GH-45, and update
  GH-29/roadmap closure evidence.
- [x] GH-53: locally prove compatible rolling binary replacement and
  fail-before-open rollback on persisted four-validator homes, including
  app-hash, power, key identity, monotonic signer-state, export, ledger, and
  re-import evidence; replace unsafe full-home rollback guidance.
- [x] GH-53: publish PR #54, obtain green final-head GitHub checks, resolve all
  six review threads, merge as `3e44905`, close GH-53, and synchronize
  GH-29/Bridge closure evidence.
- [x] GH-55: prove coupled consensus-key/signer-state cold custody and
  single-signer failover with a fresh P2P identity, app-hash convergence,
  monotonic signing position, export, and ledger validation; replace unsafe
  operator backup/recovery guidance.
- [x] GH-55: publish PR #57, pass final-head review and all GitHub checks,
  resolve all eight threads, merge as `e8670c6`, close the issue, and
  synchronize GH-29/Bridge closure evidence.
- [x] GH-56: locally implement and verify authenticated atomic consensus-key
  rotation, permanent old-key revocation, and bootstrap operator-authority
  separation without stake-withdrawing removal/re-registration.
- [x] GH-56: publish the audited branch, pass final-head GitHub review/CI,
  merge as `80ab674`, close the issue, and synchronize GH-29 plus the Bridge.
- [x] GH-7: DEX rounding, slippage, pool accounting, custody, and authorization
  audit completed in GH-10; GH-12 retains genesis/runtime invariants.

## P2 - delivery

- [x] GH-8: review the preserved legacy checkout; no code is safe/useful for
  wholesale merge, and the checkout remains preserved pending final archive.
- [x] GH-8: align CLAUDE, install, FAQ, README, status JSON, limitations,
  landing page, and real wiki status/security claims to the historical
  684-test recovery
  source of truth.
- [x] GH-8: enforce suite/module/cap arithmetic and real wiki/agent/public
  status through docs CI; modernize Action runtimes without credential or
  duplicate-run regression.
- [x] GH-8: publish audited head `3964f4a`; every updated GitHub Action and
  manual Security Scan `29171476126` passes.
- [x] GH-8: refreshed docs/recovery CI passed and PRs #24 and #27 are merged.
  Obsolete PR #25 remains isolated.
- [x] PRs #9, #15, #16, #17, #18, #19, #22, #23, #24, and #27 are merged to
  `main`.
- [ ] GH-29: keep the public Road to Rollout page, detailed checklist, GitHub
  issue, and Bridge status synchronized as workstreams close.
- [x] GH-37: configure project-scoped Codex subagent roles so the main agent can
  delegate small bounded work to `spark_worker` without losing architecture,
  security, GitHub, or Bridge ownership.
