# TrueRepublic Agent Bridge

Canonical coordination lives in [`docs/agent-bridge/`](docs/agent-bridge/README.md).

- Current state: [`PROJECT_STATE.md`](docs/agent-bridge/PROJECT_STATE.md)
- Work queue: [`TODO.md`](docs/agent-bridge/TODO.md)
- Audit trail: [`ACTION_LOG.md`](docs/agent-bridge/ACTION_LOG.md)
- GH-11 cap audit: [`PR15_AUDIT.md`](docs/agent-bridge/PR15_AUDIT.md)
- GH-14 escrow audit: [`PR16_AUDIT.md`](docs/agent-bridge/PR16_AUDIT.md)
- GH-13 issuance audit: [`PR17_AUDIT.md`](docs/agent-bridge/PR17_AUDIT.md)
- GH-10 DEX custody audit: [`PR18_AUDIT.md`](docs/agent-bridge/PR18_AUDIT.md)
- GH-12 genesis/invariant audit: [`PR19_AUDIT.md`](docs/agent-bridge/PR19_AUDIT.md)
- GH-20 ZKP/auth audit: [`PR22_AUDIT.md`](docs/agent-bridge/PR22_AUDIT.md)
- GH-21 node lifecycle audit: [`PR23_AUDIT.md`](docs/agent-bridge/PR23_AUDIT.md)
- GH-8 docs/CI audit: [`PR24_AUDIT.md`](docs/agent-bridge/PR24_AUDIT.md)
- GH-26 operator init audit: [`PR27_AUDIT.md`](docs/agent-bridge/PR27_AUDIT.md)
- GH-32 multi-validator audit: [`GH32_AUDIT.md`](docs/agent-bridge/GH32_AUDIT.md)
- Decisions: [`DECISIONS.md`](docs/agent-bridge/DECISIONS.md)
- Security: [`SECURITY_NOTES.md`](docs/agent-bridge/SECURITY_NOTES.md)

GitHub recovery epic: [#4](https://github.com/NeaBouli/TrueRepublic/issues/4)

## 2026-07-13 02:50 EEST GH-29 Road to Rollout → Done

- **Branch:** `main`
- **Issue:** [GH-29](https://github.com/NeaBouli/TrueRepublic/issues/29)
  remains open as the rollout execution tracker
- **PR:** [#30](https://github.com/NeaBouli/TrueRepublic/pull/30), merged as
  `162038f`
- **Changed:** published the English seven-phase Road to Rollout board, the
  full evidence checklist, rollout exit gates, staged-launch sequence, and
  explicit non-production safety boundary
- **Tests:** documentation consistency, diff/content/conflict checks, browser
  DOM/console, desktop and 375px mobile rendering, Docs CI, Go/Rust security,
  Node audits, DeepScan, and CodeRabbit → PASS
- **Risk:** Low runtime risk; the roadmap does not grant production, mainnet,
  real-funds, or real-key approval
- **Next:** execute and continuously update the open GH-29 workstreams until
  every rollout gate has linked evidence and an explicit go/no-go decision

### Codex review feedback

Approved and merged. The public GitHub page and issue tracker now expose the
complete known path from the recovered foundation to a controlled rollout.

## 2026-07-13 02:35 EEST GH-29 Road to Rollout → Review

- **Branch:** `docs/GH-29-rollout-roadmap`
- **Issue:** [GH-29](https://github.com/NeaBouli/TrueRepublic/issues/29)
- **PR:** [#30](https://github.com/NeaBouli/TrueRepublic/pull/30)
- **Changed:** adding an English public rollout board and detailed checklist
  covering network recovery, production ZKP/privacy, IBC completeness, client
  consolidation, quality/security, operations, release engineering, staged
  testnets, and the final go/no-go gate
- **Tests:** `bash scripts/check-consistency.sh` → PASS; `git diff --check`
  → PASS; required content/link and conflict-marker checks → PASS; browser
  DOM/console → PASS; desktop and 375px mobile rendering → PASS without
  horizontal overflow
- **Risk:** Low for runtime; high communication importance because incomplete
  work must not be mistaken for production approval
- **Ready for:** GitHub CI, review, merge, and Pages deploy

### Codex review feedback

The seven phases cover the known technical and operational blockers through a
staged rollout and explicit go/no-go decision. The page preserves the recovery
safety boundary and does not imply a release date or production approval.

## 2026-07-13 00:37 EEST Recovery merge chain → Done

- **Branch:** `main`
- **Issues:** [GH-4](https://github.com/NeaBouli/TrueRepublic/issues/4),
  [GH-26](https://github.com/NeaBouli/TrueRepublic/issues/26)
- **Merged:** PRs #9, #15, #16, #17, #18, #19, #22, #23, #24, and #27;
  canonical head `513716c`
- **Changed:** completed the ordered recovery foundation, ZKP binding, node
  lifecycle, docs/CI reconciliation, and safe daemon-only operator init path
- **Tests:** PR #27 local Go race/coverage → PASS; GitHub Go/Docker run
  `29190764808`, Docs/Pages run `29190763221`, Security run `29190764842`,
  DeepScan, and CodeRabbit → PASS
- **Risk:** High — this is a recovered engineering foundation, not a production
  or public-network approval; documented crypto and operations blockers remain
- **Next:** serve GitHub Pages from canonical `main:/docs` and continue GH-4
  release-qualification work from the clean baseline

### Codex review feedback

Approved and merged. The ordered PR chain is complete on `main`; the 21M PNYX
cap, 684-test source of truth, remaining blockers, and non-production boundary
are synchronized in repository documentation.

## 2026-07-13 00:15 EEST GH-26 safe operator init → Review

- **Branch:** `fix/GH-26-pod-init-script`
- **Issue:** [GH-26](https://github.com/NeaBouli/TrueRepublic/issues/26)
- **PR:** [#27](https://github.com/NeaBouli/TrueRepublic/pull/27)
- **Changed:** rebased the safe daemon-only initialization wrapper and its
  regression/evidence commits onto the verified PR #24 merge; documentation
  conflicts were reconciled to preserve the current `main` recovery status
- **Tests:** `go test ./... -race -cover -count=1 -timeout=600s` → PASS;
  shell syntax, docs consistency, JSON/YAML parsing, conflict-marker scan, and
  diff checks → PASS; GitHub Docker/security gates are being refreshed
  against `main`
- **Risk:** High — operator genesis, validator identity, token supply, and key
  safety; production readiness remains explicitly false
- **Ready for:** local verification, refreshed GitHub CI, and squash merge

### Codex review feedback

The implementation commit is patch-equivalent to the previously green PR #27.
Only stale pre-merge documentation required conflict resolution; no runtime
behavior was changed during the rebase.

## 2026-07-12 13:14 EEST Public GitHub Pages → Recovery status live

- **Site:** [neabouli.github.io/TrueRepublic](https://neabouli.github.io/TrueRepublic/)
- **Source:** `fix/GH-26-pod-init-script:/docs` at `50b0d9a`; GitHub Pages
  build `1090733247` → BUILT without error
- **Live verification:** page reports `Recovery audit active`, `not approved
  for production`, `21M` maximum PNYX supply, and `684 recovery-verified tests`
- **Safety:** no runtime commit was merged to `main`, no branch-protection rule
  was bypassed, and red PR #25 remains untouched against unrecovered `main`
- **Remaining protected gate:** PR #9 is fully green/mergeable and awaits one
  independent approval before the ordered stack can reach `main`

## 2026-07-12 13:01 EEST GH-26 safe operator init → GitHub green

- **Branch:** `fix/GH-26-pod-init-script`
- **Issue:** [GH-26](https://github.com/NeaBouli/TrueRepublic/issues/26)
- **PR:** [#27](https://github.com/NeaBouli/TrueRepublic/pull/27) (stacked draft
  against final PR #24)
- **Changed:** removed every keyring-account, mnemonic-file, `gentx`,
  `collect-gentxs`, and extra-genesis-supply action from `scripts/init-node.sh`;
  it now delegates only to the generated-key, exact bank-backed PoD daemon init
- **Regression:** asserts the exact daemon command, forbidden-command absence,
  gas/Prometheus edits, no mnemonic artifact, and supported-path status output
- **Docs:** quick start, native install, wiki, limitations, decisions, security,
  audit, public status, and test source of truth now describe one init boundary
- **Tests:** focused wrapper regression, real compiled-daemon init/genesis
  assertions, full 650-case Go suite, vet, shell syntax, docs/JSON/diff → PASS
- **Risk:** High — operator genesis, validator identity, token supply, key safety
- **GitHub:** implementation/audit head `86ff1c8`; Go race/coverage and Docker
  restart run `29172845624`, Docs `29172845627`, Security `29172846057`,
  DeepScan, and CodeRabbit → PASS; zero unresolved review threads
- **Ready for:** independent operations review and ordered stack merge; not
  production

## 2026-07-12 12:35 EEST GH-12 review remediation → Stack green

- **Branch:** `fix/GH-12-genesis-invariants`
- **Issue:** [GH-12](https://github.com/NeaBouli/TrueRepublic/issues/12)
- **PR:** [#19](https://github.com/NeaBouli/TrueRepublic/pull/19)
- **Changed:** verified that the no-error-return `CreateDomain` helper actually
  creates the intended escrow divergence instead of silently continuing; made
  production-bootstrap requirements explicit: a real Ed25519 key, positive
  CometBFT voting power, sufficient exact bank-backed stake, and canonical
  supply within the 21,000,000 PNYX cap
- **Tests:** focused registered-invariant regression, `go test ./... -count=1`,
  `go vet ./...`, and `go build ./...` → PASS
- **GitHub:** commit `eec91c7` published; both review threads answered and
  resolved; PR #22 published at `0c72ad0`, PR #23 published at `49938a3`, and
  PR #24 published with the propagated fix; all refreshed PR #19/#22/#23/#24
  checks and Security runs `29172007410`, `29172246257`, `29172246373`, and
  `29172246235` pass; all four PRs are mergeable with zero unresolved threads
- **Ready for:** independent approval of PR #9 and ordered stack merge; the
  active recovery goal continues and is not blocked

## 2026-07-12 12:09 EEST GH-8 docs/CI reconciliation → GitHub green

- **Branch:** `fix/GH-8-docs-final`
- **Issue:** [GH-8](https://github.com/NeaBouli/TrueRepublic/issues/8)
- **PR:** [#24](https://github.com/NeaBouli/TrueRepublic/pull/24) (stacked draft against final GH-21)
- **Changed:** Node-24-backed official Action majors with read-only/non-persisted
  checkout credentials; non-duplicate workflow triggers; strengthened suite,
  module, cap, agent-guide, landing-page, and real-wiki consistency gates;
  replaced stale CLAUDE/install/FAQ/wiki recovery and security claims
- **Audit fixes:** rebased only the six GH-8 CI/docs commits onto final GH-21
  `49938a3`; preserved Node 22 for the maintained client; corrected false
  anonymous-voting/mobile-wallet availability; made installation explicitly
  select the unmerged recovery branch; replaced the skipped `wiki-github/`
  checks and created missing current/testing status pages
- **Tests:** every workflow YAML parses; docs consistency, JSON, relative wiki
  target, stale-current-claim, and diff checks → PASS; underlying 683-case
  GH-21 code head remains unchanged
- **Risk:** Medium — public security/readiness claims and CI trust/runtime
- **GitHub:** Go race/coverage + Docker restart `29172243080`, Rust
  `29172243094`, Web `29172243172`, Mobile `29172243069`, Docs
  `29172243125`, DeepScan, CodeRabbit, and all five Security Scan
  `29172246235` jobs → PASS
- **Ready for:** independent documentation/recovery review and ordered stack
  merge; PR #25 remains separately blocked on old-main security

### Codex review feedback

Conditional PASS. The initial rebased draft still contained 636-test,
Go-1.23, anonymous-voting, mobile-wallet, and Testnet-Ready claims and a docs
gate pointed at a nonexistent wiki directory. Those findings are remediated.
All modernized Action majors now pass on GitHub. PR #25 remains a separate
default-branch visibility track and must not bypass the vulnerable current
`main` or the ordered recovery stack.

---

## 2026-07-12 11:41 EEST GH-21 node lifecycle → GitHub green

- **Branch:** `fix/GH-21-node-lifecycle`
- **Issue:** [GH-21](https://github.com/NeaBouli/TrueRepublic/issues/21)
- **PR:** [#23](https://github.com/NeaBouli/TrueRepublic/pull/23) (stacked draft against GH-20), audited code head `ec1ce17`
- **Changed:** standard Cosmos server lifecycle with persistent home/database,
  generated CometBFT-key PoD genesis with exact bank-backed stake, consensus
  parameter keeper, clean signal shutdown, export, non-root Debian/glibc image,
  persistent container restart gate, and restored CLI build-version metadata
- **Audit fixes:** rebased without content drift onto final GH-20 head `0c72ad0`;
  rejected existing/conflicting consensus validator sets without mutation;
  wrote genesis atomically with mode `0600`; fixed the reproduced blank/failing
  `version` and `--version` interfaces before publication
- **Tests:** `go test ./... -count=1 -timeout=600s` → PASS; 649 Go cases and
  coverage → PASS (root 64.3%, token 92.6%, treasury 97.0%, DEX 45.3%,
  governance 58.9%); targeted lifecycle race, vet, CGO build, CLI version,
  shell syntax, and diff checks → PASS
- **Risk:** Critical — consensus key identity, bank-backed bootstrap stake,
  persistent application state, restart safety, and operator container runtime
- **GitHub:** Go build/vet/race/coverage and Docker block/restart pass in run
  `29172166826`; Docs, DeepScan, Web, CodeRabbit, and manual Security Scan
  `29172246373` pass
- **Ready for:** independent multi-node operations review and ordered stack
  merge; not production

### Codex review feedback

Conditional PASS. The former MemDB/`select {}` placeholder and invalid
`x/staking` gentx bootstrap are no longer the node path. A real subprocess
produced blocks, stopped on SIGINT, restarted from the same home, advanced
height, preserved invariants, and exported state. GitHub independently repeated
the image build and same-container restart. IBC staking/upgrade and multi-node
operations remain explicit non-production boundaries.

## 2026-07-12 05:28 EEST GH-20 ZKP/authentication → GitHub green

- **Branch:** `fix/GH-20-zkp-binding`
- **Issue:** [GH-20](https://github.com/NeaBouli/TrueRepublic/issues/20)
- **PR:** [#22](https://github.com/NeaBouli/TrueRepublic/pull/22) (stacked draft against GH-12)
- **Changed:** versioned chain/proposal/rating proof signal; chain-scoped,
  rating-independent one-vote nullifier; fail-closed genesis VK; canonical
  BN254 fields; exact active-nullifier export/import; disabled mock submission
- **Audit fixes:** rebased to final PR #19; pinned circuit ID, VK SHA-256,
  curve/public-input shape, and canonical encoding; recomputed genesis Merkle
  roots; rejected malformed ZKP state; preserved Big-Purge nullifier semantics;
  removed both public clients' false mock-proof submission path
- **Tests:** Go build/vet, 643 cases, race, and coverage → PASS (root 66.1%,
  token 92.6%, treasury 97.0%, DEX 45.3%, governance 58.9%); Rust 26 tests/
  audit; maintained client lint/8 tests/build/audit; legacy ZKP 4 tests/build →
  PASS
- **Risk:** Critical — anonymous vote integrity, cross-chain replay, trusted
  setup determinism, genesis identity roots, and double-vote state
- **GitHub:** Docs, DeepScan, Web, Mobile, Rust, both Go/Docker runs, and manual
  Security Scan run `29159603247` pass; CodeRabbit is check-green but explicitly
  rate-limited and produced no substantive review
- **Ready for:** independent cryptographic review; compatible real client prover
  remains intentionally unavailable

### Codex review feedback

Conditional PASS for GH-20's on-chain binding and fail-closed client scope.
Anonymous rewards remain deferred because the proof does not bind a safe bank
recipient. A real prover plus external ceremony/circuit review is still required
before advertising or enabling anonymous voting.

## 2026-07-12 04:52 EEST GH-12 genesis and invariants → GitHub green

- **Branch:** `fix/GH-12-genesis-invariants`
- **Issue:** [GH-12](https://github.com/NeaBouli/TrueRepublic/issues/12)
- **PR:** [#19](https://github.com/NeaBouli/TrueRepublic/pull/19) (stacked draft against GH-10)
- **Changed:** pre-mutation custom-genesis validation and exact module-bank
  reconciliation, provider LP export, non-empty round-trip preservation,
  every-block supply/escrow/reserve/LP crisis invariants, and repaired custom
  service/app startup wiring
- **Audit fixes:** rebased onto final PR #18; adapted LP export/invariants to
  collision-free keys; removed a publicly derivable bootstrap validator secret;
  bootstraps only from real CometBFT Ed25519 public keys with exact stake; made
  InitGenesis failures explicit; added four full-app divergence regressions
- **Tests:** Go build/vet, 615 cases, race, and coverage → PASS (root 66.1%,
  token 92.6%, treasury 97.0%, DEX 45.3%, governance 56.6%); Rust 26 tests/
  audit and maintained client install/lint/6 tests/build/audit → PASS
- **Risk:** Critical — InitChain, validator keys, canonical supply, module
  escrow, DEX reserves, and consensus-halting invariants
- **GitHub:** Docs, DeepScan, Web, Mobile, Rust, Go build/vet/test, and both
  Docker builds pass; manual Security Scan run `29158360390` passes all five
  jobs
- **Ready for:** independent cryptographic/stacked review; CodeRabbit is
  requested and still pending

### Codex review feedback

Conditional PASS for the ledger/genesis scope. The old default bootstrap would
have exposed a reproducible consensus private key; it is removed. GH-21 must
replace the still-invalid legacy `x/staking` gentx script with a PoD-aware real
validator-key flow before production node launch.

## 2026-07-12 03:34 EEST GH-10 DEX custody → Local verification

- **Branch:** `fix/GH-10-dex-custody`
- **Issue:** [GH-10](https://github.com/NeaBouli/TrueRepublic/issues/10)
- **PR:** [#18](https://github.com/NeaBouli/TrueRepublic/pull/18) (stacked draft against GH-13)
- **Changed:** bank-backed pool custody, atomic create/add/remove/swap
  settlement, provider-indexed LP ownership, governance authority for registry
  mutation, and canonical PNYX burns through `token.IssuanceService`
- **Audit fixes:** rebased onto final PR #17, retained both module burn
  permissions, replaced collision-prone textual LP prefixes with
  length-prefixed keys, and added rollback regressions for every custody flow
- **Tests:** Go build/vet, 578 cases, and race → PASS; Rust 26 tests/audit →
  PASS with six tracked transitive warnings; maintained client install/lint/6
  tests/build/audit → PASS; docs/module/diff consistency → PASS
- **Risk:** High — user funds, pool reserves, LP ownership, canonical supply,
  and chain-wide asset authorization
- **GitHub:** docs, DeepScan, Go build/vet/race/coverage, and the real Docker
  build pass on `3234741`; manual Security Scan run `29156922464` passes all
  five jobs
- **Ready for:** independent review; CodeRabbit is temporarily rate-limited
  and did not produce a substantive review

### Codex review feedback

Conditional PASS for GH-10. Every public DEX value transition now reconciles
bank custody, pool reserves, provider shares, and canonical burns before commit.
GH-12 custom-genesis reconciliation/runtime invariants still block production.

## 2026-07-11 20:09 EEST GH-4 foundation merge audit → Review

- **Branch:** `fix/GH-4-recovery-foundation`
- **Issue:** [GH-4](https://github.com/NeaBouli/TrueRepublic/issues/4)
- **PR:** [#9](https://github.com/NeaBouli/TrueRepublic/pull/9)
- **Changed:** audit-only follow-up adds Docker build coverage, records the
  merge review, and removes whitespace-only diff errors
- **Tests:** Go build/test/race/vet → PASS; govulncheck fixable gate → PASS;
  Rust 26 tests/Clippy/audit → PASS with six allowed transitive warnings;
  maintained client lint/5 tests/build/audit → PASS; docs consistency → PASS
- **Risk:** Medium — dependency/toolchain foundation; no consensus or ledger
  implementation changes
- **Ready for:** refreshed GitHub CI and an independent GitHub approval

### Codex review feedback

Conditional approval for the recovery-foundation scope. The seven ledger and
token-economy blockers in `CODEX_AUDIT.md` remain explicitly out of scope and
must stay non-production until the ordered implementation PRs land. Do not
bypass the required independent GitHub approval.

---

## 2026-07-12 23:53 EEST GH-21 node lifecycle → Review

- **Branch:** `fix/GH-21-node-lifecycle`
- **Issue:** [GH-21](https://github.com/NeaBouli/TrueRepublic/issues/21)
- **PR:** [#23](https://github.com/NeaBouli/TrueRepublic/pull/23)
- **Changed:** rebased the ten GH-21 lifecycle commits onto `main` after the
  verified squash merge of PR #22; no implementation delta was introduced
- **Tests:** `git range-diff 0c72ad0..backup/GH-21-before-main-20260712 origin/main..HEAD`
  → 10/10 commits patch-equivalent; `go test ./... -race -cover -count=1
  -timeout=600s` → PASS (sandbox-exempt run required for localhost bind)
- **Risk:** High — persistent node lifecycle, genesis reconciliation, and
  container restart behavior; Docker is verified by the required GitHub gate
- **Ready for:** refreshed GitHub CI, review, and ordered squash merge

### Codex review feedback

The rebased code is patch-equivalent to the previously reviewed GH-21 stack.
Local Go race/coverage verification passes; merge remains gated on refreshed
GitHub build, Docker restart smoke, lint, and analysis checks.

---

## 2026-07-11 22:08 EEST GH-14 escrow audit → Local verification

- **Branch:** `fix/GH-14-bank-escrow`
- **Issue:** [GH-14](https://github.com/NeaBouli/TrueRepublic/issues/14)
- **PR:** [#16](https://github.com/NeaBouli/TrueRepublic/pull/16)
- **Changed:** bank-backed domain/stake claims, atomic transfers, authenticated
  signer claims, signer-safe CosmWasm bindings, and real validator slash burns
- **Audit fixes:** closed a contract-message signer regression and prevented
  slashed PNYX from being recycled through admin-withdrawable domain treasury
- **Tests:** Go build/vet/race/coverage and 557 Go cases → PASS; Rust 26
  tests/Clippy/audit → PASS with six allowed transitive warnings; maintained
  client lint/6 tests/build/audit and docs consistency → PASS
- **Risk:** High — consensus-adjacent bank custody and validator accounting
- **Ready for:** force-push of the rebased stacked branch and refreshed GitHub
  CI/review

### Codex review feedback

The GH-14 custody boundary is locally coherent after remediation. Runtime
issuance, DEX custody, and custom-genesis invariants remain isolated in GH-13,
GH-10, and GH-12 and keep the repository non-production.

---

## 2026-07-11 21:25 EEST PR #9 review remediation → Verification

- **Branch:** `fix/GH-4-recovery-foundation`
- **Issue:** [GH-4](https://github.com/NeaBouli/TrueRepublic/issues/4)
- **PR:** [#9](https://github.com/NeaBouli/TrueRepublic/pull/9)
- **Changed:** hardened checkout credentials and workflow permissions, aligned
  canonical CosmJS CI with Node 22, updated current Go security dependencies,
  removed the hard-coded wasmvm module version, and synchronized public/bridge
  recovery status
- **Tests:** Go build/vet/race/coverage, govulncheck fixable gate, Rust
  tests/Clippy/audit, Node-22 client install/lint/tests/build/audit, docs,
  workflow hygiene, and dynamic wasmvm-path checks → PASS locally; refreshed
  GitHub CI pending for this remediation commit
- **Risk:** Medium — dependency and CI hardening, without consensus/ledger code
- **Ready for:** verification, thread-by-thread review responses, then the
  already-requested independent approval

### Codex review feedback

All 12 unresolved CodeRabbit threads were verified and mapped to six focused
remediation clusters. No administrative branch-protection bypass is permitted.

---

## 2026-07-11 20:54 EEST GH-4 wasmvm Docker linkage → Local verification

- **Branch:** `fix/GH-4-recovery-foundation`
- **Issue:** [GH-4](https://github.com/NeaBouli/TrueRepublic/issues/4)
- **PR:** [#9](https://github.com/NeaBouli/TrueRepublic/pull/9)
- **Changed:** replaced the Alpine/musl builder and runtime with Debian/glibc,
  copied the architecture-specific `libwasmvm` shared object into the runtime,
  and registered it with `ldconfig`
- **Tests:** reproduced GitHub musl/GLIBC linker failure twice; local Go build,
  docs consistency, workflow YAML, and diff checks → PASS; both corrected
  GitHub Docker builds → PASS
- **Risk:** Medium — node container build/runtime linkage
- **Ready for:** independent GitHub approval after review remediation checks

### Codex review feedback

The patch matches the glibc/wasmvm linkage already proven by GH-21 while keeping
GH-4's existing entrypoint and root data path unchanged. Both GitHub Docker
jobs now prove the corrected image builds.

---

## 2026-07-11 23:02 EEST GH-13 canonical reward issuance → GitHub verification

- **Branch:** `fix/GH-13-cap-issuance`
- **Issue:** [GH-13](https://github.com/NeaBouli/TrueRepublic/issues/13)
- **PR:** [#17](https://github.com/NeaBouli/TrueRepublic/pull/17) (stacked draft against GH-14)
- **Changed:** canonical bank-supply issuance service, cap-clipped validator and
  domain inflation, supply-neutral treasury payouts, interval payout snapshots,
  centralized slash burns, atomic two-phase EndBlock reward settlement, and an
  architecture-safe/reproducible wasmvm node image with a reduced build context
- **Audit fixes:** rejected invalid canonical supply, closed the partial
  EndBlock commit boundary between staking issuance and domain issuance, and
  removed a duplicate Amino registration that panicked every CLI/node startup;
  final review also made bank-mock issuance rollback-aware and baselined restored
  domain payout snapshots
- **Tests:** Go build/vet, 569 cases, race, and coverage → PASS; token 93.5%,
  governance 55.8%; Rust 26 tests/Clippy/audit, client lint/6 tests/build/audit,
  documentation consistency, Dockerfile/YAML/diff checks → PASS. Both GitHub
  Docker builds, both Go jobs, DeepScan, docs, and the manual security workflow
  → PASS on the final remediation head; the complete security workflow also
  passes and all five review threads are resolved.
- **Risk:** High — canonical supply, mint/burn authority, reward inflation,
  validator power, and treasury claims
- **Ready for:** ordered stacked review after PRs #9, #15, and #16

### Codex review feedback

Conditional PASS for the GH-13 scope after the three audit hardenings. DEX
custody/burn integration, custom-genesis reconciliation, runtime invariants,
and anonymous recipient binding remain separately blocking.

---

## 2026-07-12 23:53 EEST GH-8 recovery documentation → Review

- **Branch:** `fix/GH-8-docs-final`
- **Issue:** [GH-8](https://github.com/NeaBouli/TrueRepublic/issues/8)
- **PR:** [#24](https://github.com/NeaBouli/TrueRepublic/pull/24)
- **Changed:** rebased all eight GH-8 commits patch-equivalently onto the PR
  #23 merge; reconciled README, machine status, wiki, security notes, project
  state, and queue with the recovery foundation now present on `main`
- **Tests:** `git range-diff 49938a3..backup/GH-8-before-main-20260713
  origin/main..HEAD` → 8/8 commits patch-equivalent;
  `bash scripts/check-consistency.sh` → PASS; workflow YAML parse → PASS;
  `git diff --check` → PASS after removing two Markdown trailing spaces
- **Risk:** Medium — public recovery truth and CI definitions; production
  readiness remains explicitly false
- **Ready for:** refreshed GitHub CI, review, and ordered squash merge

### Codex review feedback

The documentation now distinguishes a merged recovery foundation from a
production release. Cryptographic, multi-node operations, legacy-client, and
release-process blockers remain prominent and unchanged.

---

## 2026-07-14 04:03 EEST GH-32 multi-validator recovery → Done

- **Branch:** `main` at `9d68a6f`
- **Issue:** [GH-32](https://github.com/NeaBouli/TrueRepublic/issues/32), closed;
  child of open rollout tracker [GH-29](https://github.com/NeaBouli/TrueRepublic/issues/29)
- **PR:** [#33](https://github.com/NeaBouli/TrueRepublic/pull/33), squash merged
- **Changed:** added internal public-key-only multi-validator genesis assembly;
  four independent validator homes; consensus, failure, restart/catch-up,
  common-height app-hash, clean shutdown, export, and ledger checks; dedicated
  CI job; operator runbook; synchronized public recovery status
- **Tests:** genesis/binder regressions → PASS; full normal Go suite → 651 PASS;
  separately gated four-validator harness → PASS three times locally, latest
  hardened run 68.90s; final PR #33 GitHub matrix → PASS;
  full Go race/coverage (root 64.9%), build, vet, docs consistency, workflow
  YAML, JSON, and diff checks → PASS; final `main` Security run `29261145077`
  and Pages build `1093339877` at `2851759` → PASS
- **Risk:** High — consensus identity, PoD/bank genesis parity, quorum recovery,
  persistent state, and CI process cleanup
- **Next:** continue the remaining GH-29 Phase 1 network/disaster-recovery gates;
  this bounded result is not production or public-network approval

### Lead Dev notes

The first real run exposed strict-loopback address-book rejection and a shared
pprof port. Both adjustments are confined to temporary localhost test config;
production CometBFT defaults and the single-node `init` refusal boundary remain
unchanged. CodeRabbit then correctly identified that the recovered node was
only compared at a pre-rejoin height and that subprocess/HTTP work lacked test
cancellation. The harness now requires all four nodes to reach and agree at a
height selected two blocks after the stopped process restarts, and threads
`testing.T` context through every child process and request.
This closes only the bounded four-validator Phase 1 checklist item.

### Codex review feedback

Three findings were accepted and locally/GitHub verified. The proposed
`actions/checkout@v7` change was rejected because no official v7 release
exists; the job remains aligned with the repository's existing v5 baseline.
Four review threads are resolved. GitHub's first final-head attempt returned
`Service Unavailable` before action download for four jobs; failed-job reruns
then passed without code changes. PR #33 merged as `9d68a6f` and closed GH-32.
Bridge closure PR #34 merged as `2851759`; Pages is live from `main:/docs` with
the 685-case count and bounded four-validator evidence.

---

## 2026-07-15 20:10 EEST GH-37 Codex subagent role configuration -> Done

- **Branch:** `chore/GH-37-codex-agent-roles`
- **Issue:** [GH-37](https://github.com/NeaBouli/TrueRepublic/issues/37)
- **PR:** [#38](https://github.com/NeaBouli/TrueRepublic/pull/38)
- **Changed:**
  - `.codex/config.toml` - project subagent concurrency and depth limits
  - `.codex/agents/spark-worker.toml` - narrow Spark worker role
  - `docs/agent-bridge/COOPERATION_RULES.md` - Sol/main and Spark role split
  - `docs/agent-bridge/DECISIONS.md` - durable delegation decision
  - `docs/agent-bridge/ACTION_LOG.md` - task progress log
  - `docs/agent-bridge/TODO.md` - GH-37 queue entry
- **Tests:** `python3`/`tomllib` parse for both `.codex` TOML files -> PASS;
  `git diff --check` -> PASS; `bash scripts/check-consistency.sh` -> PASS;
  GitHub Docs Consistency, Security Scan, and DeepScan -> PASS
- **Risk:** Low - workflow configuration and documentation only; no runtime,
  consensus, wallet, or public status behavior changes
- **Ready for:** merged workflow baseline

### Lead Dev notes

The main agent remains responsible for architecture, security/risk judgment,
final verification, GitHub issue/PR state, merges, pushes, and Bridge updates.
`spark_worker` is intentionally limited to small delegated patches, file search,
and focused checks, then reports back to the main agent.

### Codex review feedback

Local and GitHub verification pass. CodeRabbit remained pending without
comments at merge time; branch protection has no required status checks and the
change is workflow/documentation-only. No runtime or public status behavior is
affected.

---
