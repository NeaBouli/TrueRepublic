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
- Decisions: [`DECISIONS.md`](docs/agent-bridge/DECISIONS.md)
- Security: [`SECURITY_NOTES.md`](docs/agent-bridge/SECURITY_NOTES.md)

GitHub recovery epic: [#4](https://github.com/NeaBouli/TrueRepublic/issues/4)

## 2026-07-12 12:48 EEST GH-26 safe operator init → Locally verified

- **Branch:** `fix/GH-26-pod-init-script`
- **Issue:** [GH-26](https://github.com/NeaBouli/TrueRepublic/issues/26)
- **Planned PR:** stacked against final PR #24
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
- **Ready for:** publication and GitHub Go/Docker/security verification

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
