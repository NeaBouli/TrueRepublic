# TrueRepublic Agent Bridge

## 2026-08-16 14:54 EEST GH-222 install lifecycle → In Progress

- **Branch:** `feature/GH-222-install-lifecycle`
- **Issue:** [GH-222](https://github.com/NeaBouli/TrueRepublic/issues/222)
- **Scope:** add one versioned, repository-owned native daemon install,
  compatible-upgrade, pre-start rollback and safe-uninstall contract with
  disposable fail-closed tests; reconcile operator documentation and CI.
- **Roles:** Kimi K3 owns one bounded secret-free implementation/audit block.
  Sol owns architecture, security boundaries, integration, complete tests,
  GitHub writes, merge and closure.
- **Risk:** High. Installer mistakes can replace the wrong binary, lose operator
  state, bypass artifact verification, or make rollback impossible.
- **Boundary:** isolated local prefixes and synthetic fixtures only; no tag,
  release, registry, published artifact, production host/network, real key/fund,
  destructive state removal or go/no-go action.
- **Next:** Kimi/Sol audit current release evidence and unsafe stale installation
  guidance, then implement the smallest fail-closed lifecycle contract.

### 2026-08-16 15:46 EEST local implementation / review remediation

- Added a generic `truerepublic.install-lifecycle/v1` contract and isolated
  `installlifecycle` CLI/package for install, status, pre-start, compatible
  upgrade, newest-snapshot rollback and allowlisted uninstall. Identity binds
  exact digest, source commit, Linux target and runtime; operator state must be
  outside the managed prefix and is never read or removed.
- Kimi contributed the bounded safety design and a final read-only review. Its
  P2 inside-prefix upgrade-artifact finding and P3 orphan-status, macOS-doc and
  CI-trigger findings were remediated. Residual local path-swap risk is bounded
  by full-chain symlink checks, opened-file identity comparison, post-copy
  digest verification and the documented exclusive root-owned prefix parent.
- Claude Code's small documentation audit was attempted but its OAuth session
  was expired; it produced no output or file change. A bounded Codex worker
  implemented only the assigned new core paths; Sol reviewed and hardened every
  line and owns all integration changes.
- Removed unsupported GOPATH installation, in-place replacement, fictitious
  migration, destructive validator reset/volume deletion and unreachable macOS
  service guidance. The historical v0.3.0 root note now carries an explicit
  supersession warning and no operational quick-start commands.
- PASS: lifecycle/CLI tests, root repository contract, Make target, JSON, YAML,
  gofmt, vet and diff hygiene. An earlier full `make verify` passed all 19 Go
  packages with Race/Coverage; it predates the final review remediations and
  therefore does not close the gate. Fresh full gates are next.
- Published status intentionally remains 1,896 and 33/59 until protected merge.
  No tag, release, artifact publication, deployment, real operator state,
  network, key, fund or production action occurred.

### 2026-08-16 16:02 EEST fresh full local gates passed

- PASS on the final remediated tree: complete Go package selection/build/vet/
  Race/Coverage (`installlifecycle` 76.6%), exact pinned staticcheck, Go
  vulnerability policy plus negative fixtures, secret scan, lifecycle and root
  repository contracts, release-evidence and deterministic-daemon fixtures.
- PASS: maintained client lint, 319 cases, TypeScript/Vite production build and
  bundle budgets; Rust format, Clippy with warnings denied and all 26 cases;
  documentation consistency, JSON/YAML, formatting and diff hygiene.
- Fresh package-scoped JSON evidence adds 23 install-lifecycle cases plus one
  root repository case. Candidate arithmetic is 1,920 = 1,575 Go + 26 Rust +
  319 client; public files deliberately retain merged-main 1,896 until the
  protected implementation merge is real.
- Docker CLI is unavailable locally, so hosted Docker/Compose evidence remains
  authoritative. GH-222 is locally ready for commit, protected PR and review.

---

## 2026-08-16 12:58 EEST GH-218 GitHub repository hygiene → In Progress

- **Branch:** `docs/GH-218-github-hygiene`
- **Issue:** [GH-218](https://github.com/NeaBouli/TrueRepublic/issues/218)
- **Scope:** reconcile misleading historical releases, completed recovery
  epics, stale/archival branches, merged-branch cleanup, universal protected
  status checks, the unresolved repository-license decision and GH-29 status.
- **Roles:** Kimi K3 performs one bounded secret-free read-only audit of the
  exact cleanup and protection plan. Sol owns every external GitHub write,
  policy decision, deletion, integration, verification and closeout.
- **Risk:** Medium. Release/issue/branch cleanup is externally visible and
  branch-protection mistakes can block future work. Historical commits/tags
  and all product code must remain recoverable and unchanged.
- **Boundary:** no product/consensus/client change, rollout credit, license
  selection, artifact publication, deployment, network, key, fund or
  production action.
- **Next:** obtain Kimi review, apply only verified cleanup operations, prove
  required checks against a real protected PR, then merge and verify main/
  Pages/both Bridges.

### 2026-08-16 13:05 EEST Kimi review / Sol correction

- Kimi independently verified all ten cleanup actions and the exact reported
  context names. No P0 blocker exists.
- Kimi found that `check` was still pull-request path-filtered and therefore
  not universal. Sol removed only that PR path filter; main pushes remain
  filtered. The GH-218 PR will prove the context on the exact head.
- Kimi also required recoverability symmetry. Both unmerged recovery-status
  history and unrelated Solana history will receive verified annotated archive
  tags before their active branches are deleted.
- Original v0.4.0 and unpublished v0.3.0 release bodies are now preserved under
  `docs/archive/releases/` before any GitHub release text is replaced/deleted.
- Branch protection must preserve linear history, stale-review dismissal, one
  approval, force-push/deletion denial and the deliberate admin-bypass boundary;
  only required status contexts are added.

### 2026-08-16 13:12 EEST external cleanup / local synchronization

- v0.4.0 is now a historical, artifact-free, non-production prerelease with
  current 1,896 / 34-of-59 / production-false boundaries. The obsolete
  unpublished v0.3.0 draft release is deleted; both version tags remain.
- Annotated archive tags preserve exact tips `e756d0b` (closed PR #25 recovery
  status) and `043c62e` (unrelated Solana/pnyx history). Both tags were read
  back before the two active branches were deleted; remote heads now contain
  only `main` plus the current GH-218 task once pushed.
- Completed recovery issues #4 and #7 are closed with evidence. GH-29 remains
  the active rollout epic and carries the exact current snapshot. GH-219 now
  tracks the unresolved license decision without choosing a license.
- `delete_branch_on_merge` is enabled. Main protection now requires exact
  universal Docs/Security contexts while retaining linear history, one stale-
  dismissing review, disabled admin enforcement, and force-push/deletion denial.
- Public status references now distinguish completed recovery foundation #4
  from active rollout #29. Local documentation consistency/PR verification is
  next; no rollout or production state changed.

### 2026-08-16 14:04 EEST Kimi final audit / rollout arithmetic correction

- Kimi's second independent read-only pass verified the release, archive tags,
  removed remote branches, issue states, repository settings, workflow context
  names and passing local documentation check. It found one P2 discrepancy:
  GH-29 contains exactly 33 checked and 26 unchecked boxes, while recent status
  text said 34/59.
- Sol traced the mismatch through GitHub body-edit history and repository status
  commits. No currently unchecked rollout item has completion evidence. The
  conservative source of truth is therefore corrected to 33/59 overall and
  33/51 phase work; Phase 6 remains 6/7 and production remains false.
- This is arithmetic reconciliation, not lost implementation or a rollout
  reversal. No checkbox was unchecked and no open work was represented as done.

### 2026-08-16 14:42 EEST protected merge / exact-main verification

- PR #220 exact head `ca18671c121e3323de54a5b69144768f1ab4a21e`
  passed all nine required universal Docs/Security contexts plus DeepScan. All
  four CodeRabbit threads were remediated and resolved before admin squash.
- PR #220 merged as exact main
  `ae20e65934a3401b7487651afa67d08bbcf7ba5b`; merged-branch auto-deletion
  removed its remote head and only `main` remains.
- Exact-main Docs `31942112014`, Security `31942112026`, and Pages
  `31942111796` passed. Cache-busted live Pages and raw `docs/status.json`
  publish 1,896 / 33-of-59 / 33-of-51 / Phase 6 6-of-7 / production false and
  distinguish completed recovery #4 from active rollout #29.
- This closeout commit records evidence only. No product, license, artifact,
  deployment, public network, key, fund or production action occurred.

---

## 2026-08-16 03:11 EEST GH-212 release-evidence contract → Merged / exact main green

- **Protected merge:** PR #213 replacement head
  `b72f29784ed6d1cd7340d580d527b4883f4ca8a5` passed all 26 contexts with zero
  unresolved review threads and squash-merged as
  `1e4cdb814eefa1441e70c75cf61ff81bbd081b1b`. GH-212 is closed and the remote
  feature branch is deleted.
- **Exact-main PASS:** Docs `31915999520`, Client `31915999527`, Security
  `31915999568`, reproducible Linux `31915999607`, Go `31915999593` and Pages
  `31915999145` all completed successfully for the merge commit.
- **Tracker/public truth:** GH-29 now checks only Phase 7 “Pinned toolchains and
  supported-platform documentation.” Public arithmetic remains 1,896 total,
  rollout 34/59, phase work 34/51, Phase 6 6/7 and production false.
- **Review split:** Kimi K3 supplied architecture, deep review and the bounded
  CodeRabbit remediation; Sol reviewed/integrated every writing diff and ran
  the complete local/protected/final-main gates. Claude Code OAuth remained
  expired and Claude made no change.
- **Boundaries:** no tag, GitHub Release, signed/published artifact, authenticated
  provenance, registry push, container-reproducibility claim, deployment,
  genesis/network mutation or production approval occurred. Remaining Phase 7
  signing/publication/reproducible-tagged-release gates stay open.

---

## 2026-08-16 10:38 EEST GH-215 sovereign decentralized Alpha architecture → In Progress

- **Branch:** `docs/GH-215-sovereign-alpha`
- **Issue:** [GH-215](https://github.com/NeaBouli/TrueRepublic/issues/215)
- **Scope:** preserve `client-web` as the current Beta and define a complete
  Telegram-like, website-independent Alpha architecture for domains,
  discussion, governance, ratings, stones, administration and an integrated
  non-custodial PNYX wallet.
- **Boundary:** architecture and roadmap documentation only. No Telegram fork,
  dependency addition, wallet/cryptography/consensus/protocol change, real
  key/fund use, deployment, release, production action or rollout credit.
- **Roles:** Sol owns architecture, security, integration and publication;
  Kimi K3 receives the bounded deep platform/architecture analysis; Claude
  Code may perform one small read-only consistency inventory if available.
- **Risk:** High future implementation risk because identity, wallet custody,
  messaging privacy, moderation and governance meet in one client. This
  documentation task itself is medium risk and must keep every unproven choice
  explicit.
- **Next:** audit current repository boundaries, research primary upstream
  sources and licenses, then write and independently review the decision-ready
  architecture and phased migration plan.

---

## 2026-08-16 11:22 EEST GH-215 sovereign Alpha architecture → Local PASS

- **Decision:** preserve `client-web` as Beta; future Alpha = Flutter UI over a
  Go sovereign core, Waku conditional on A0, RFC 9420 MLS conditional on DG-2,
  Status as reference rather than fork, and no optional-web decision yet.
- **Security truth:** secrets stay in the Go core; private outer frames contain
  no plaintext account/domain/moderation identifiers, while observable topic/
  traffic correlation remains explicit; BIP-39 does not restore chat epochs;
  Store history, push, spam resistance and metadata privacy are not overclaimed.
- **Review:** Kimi produced the deep draft and independent review. Sol
  remediated its P2 state citation and P3 editorial/version findings; zero
  P0/P1 remain. Claude OAuth was expired and Claude changed nothing.
- **PASS:** docs consistency, root Go package, client lint, 10 Node + 309
  Vitest cases, production build/budgets, diff check and 22 external links.
- **State correction:** stale GH-212 `PROJECT_STATE.md` now matches the already
  published 1,896 tests, rollout 34/59, phase work 34/51, Phase 6 6/7 and
  production false.
- **Next:** protected PR, exact-head review/checks, squash merge and exact-main/
  Pages/GitHub/Bridge closeout. No release, deploy or production write.

---

## 2026-08-16 11:40 EEST GH-215 PR #216 review → Findings remediated

- First head passed all technical contexts; CodeRabbit then reported nine valid
  documentation/protocol gaps, so merge remained blocked.
- All nine are fixed: post-GH-121 golden-vector scope, license/count/TODO truth,
  transport metadata wording, signed topic binding, deterministic device-control
  sequencing and revocation, latest-observed authorization, monotonic update
  anti-rollback and the corresponding adversarial tests.
- Kimi's focused independent review found no P0/P1/P2; its only P3 genesis-hash
  detail is also fixed. Docs consistency and diff checks pass.
- **Next:** replacement commit/checks, thread resolution, squash merge and exact
  main/Pages/GitHub closeout. Rollout remains 34/59; production remains false.

---

## 2026-08-16 11:53 EEST GH-215 sovereign Alpha architecture → Done

- **Protected merge:** PR #216 replacement head
  `ac11bf59bc5168ccb9c50475c71cc7c55b8bcbce` passed every reported Docs,
  Client, Security, DeepScan and CodeRabbit context with all nine review
  threads resolved, then admin-squash-merged as
  `1190ae2a1b7b577fdd89fd325d608854874cbbfb`. GH-215 is closed.
- **Exact-main PASS:** Docs `31937409146`, Client `31937409059`, Security
  `31937409105` and Pages `31937408558` all completed successfully for the
  merge commit.
- **Published GitHub truth:** README, maintained-client documentation,
  architecture, limitations, rollout roadmap and the public project page now
  preserve `client-web` as Beta and publish the sovereign Alpha direction. The
  live Pages site was read back successfully after deployment.
- **Status:** documentation-only completion earns no rollout checkbox. Public
  arithmetic remains 1,896 tests, rollout 34/59, phase work 34/51, Phase 6
  6/7 and production false.
- **Boundary/next:** no application implementation, license selection, release,
  deployment, key/fund use or production action occurred. The next Alpha task
  is A0: qualify the native Go/Waku transport boundary and all license/SBOM,
  reproducibility, lifecycle, device, battery and bandwidth gates before any
  dependency is adopted.

---

## 2026-08-16 02:40 EEST GH-212 replacement candidate → Local PASS

- Sol completed the post-Kimi integration review and reran `make build` plus
  full `make verify`: all 1,551 maintained Go cases pass under race/coverage,
  root coverage remains 73.6% and release-evidence coverage is now 74.5%.
- Exact pinned staticcheck v0.7.0, gitleaks v8.30.1 with positive/negative
  fixtures, the security repository contract, deterministic-daemon contract,
  complete release generator/verifier adversarial flow, docs consistency,
  shell syntax, formatting and diff checks pass.
- The local environment still has no Docker engine and no standalone
  ShellCheck binary; Docker restart/build and shell/static validation therefore
  remain mandatory protected replacement-head jobs. No local claim fills those
  evidence gaps.
- Next: commit and push the reviewed replacement candidate, resolve the nine
  CodeRabbit threads, wait for all checks on the exact new head, then merge and
  perform final-main/Pages/GH-29/Bridge closeout. No release or production write.

---

## 2026-08-16 02:27 EEST GH-212 protected review → Findings remediated

- **PR/head:** PR #213 originally passed every protected context on exact head
  `44e13c8`; CodeRabbit then reported seven actionable findings and two useful
  nitpicks. Merge was correctly withheld.
- **Kimi contribution:** Kimi K3 implemented the bounded review-fix block. Sol
  reviewed the complete writing diff; Claude Code remains unavailable because
  its local OAuth session is expired and changed no files.
- **Hardening:** active Dockerfile `FROM` references are now parsed and exactly
  cross-bound to the release contract; tool pins are checked across release,
  security, deterministic-build and workflow contracts; generator metadata
  derives from the reviewed contract; fixture errors and ambiguous negative
  cases fail explicitly; Linux-only evidence boundaries and real race-command
  requirements are stated precisely.
- **Focused PASS:** release-evidence and root adversarial Go tests, root/package
  vet, executed generator/verifier tamper flow, docs consistency, formatting
  and diff checks. ShellCheck is unavailable locally; protected CI remains the
  authoritative shell/static gate.
- **Next:** run Sol's relevant complete gates, commit/push the replacement head,
  resolve every review thread, require all replacement-head checks green, then
  squash-merge and complete exact-main/Pages/tracker closeout. Production,
  release, signing, publication and deployment remain false/not performed.

---

## 2026-08-16 02:00 EEST GH-212 release-evidence contract → Local PASS

- **Implementation:** exact source/build/tool/platform-bound amd64+arm64
  artifact, checksum, metadata, normalized CycloneDX SBOM and separate unsigned
  provenance bundle with a strict network-free verifier and standalone CLI.
- **Supply-chain pins:** Go 1.26.6, Node 22.22.2/npm 10.9.7,
  `cyclonedx-gomod` v1.10.0, `cyclonedx-npm` 6.0.1 and four official
  multi-arch container manifests. Final read-only registry inspection caught
  and replaced a moving Node 22 image with exact `node:22.22.2-alpine`.
- **PASS:** fresh `make build` + complete `make verify` twice (1,551 Go cases,
  race/vet/build; root 73.6%, release evidence 74.0%); 10 Node + 309 Vitest,
  lint/build/budgets/high audit; 26 Rust tests/fmt/Clippy/build and cargo audit;
  staticcheck, gitleaks plus positive/negative fixtures, exact four-ID no-fix
  govuln policy plus fixtures, security repository contract, deterministic
  build contract, real repeated SBOM parity, synthetic bundle/tamper/cleanup,
  JSON/docs consistency and diff checks.
- **Review:** Kimi architecture plus independent deep review found no P0/P1;
  all actionable P2/P3 were remediated. Sol reviewed/hardened the complete
  writing diff. Claude OAuth remained expired and made no changes.
- **Local limitation:** Docker is not installed in this environment, so no
  local container build ran. Official registry index/config data verifies all
  four digests cover amd64+arm64 and the Node image/tool versions; protected CI
  and repository contracts remain required.
- **Public truth:** 1,896 = 1,551 Go + 26 Rust + 319 client; rollout 34/59,
  phase work 34/51, Phase 6 6/7, production false. No tag, signed/published
  artifact, authenticated provenance, container reproducibility, deployment or
  go/no-go claim. Next: commit, push, protected PR/review/CI and final-main closeout.

---

## 2026-08-16 01:30 EEST GH-212 release-evidence contract → Local candidate

- **Delivered:** strict `truerepublic.release-evidence/v1` offline verification
  binds exact source/build/tool contracts, both native Linux targets, artifacts,
  per-target checksums/metadata, repeated normalized Go/client CycloneDX SBOMs
  and a separate explicitly unsigned provenance record.
- **Hardening:** exact Go 1.26.6, Node 22.22.2/npm 10.9.7,
  `cyclonedx-gomod` v1.10.0 and `cyclonedx-npm` 6.0.1 pins; four official
  maintained image manifest digests cross-bound to both Dockerfiles; strict
  size/depth/duplicate/unknown/trailing/path/symlink/order/digest/source/status
  rejection; failed generation leaves no complete-looking bundle.
- **CI boundary:** real Go/client SBOMs are generated twice and normalized for
  parity; a synthetic two-architecture positive/adversarial bundle is verified
  offline. No daemon binary, bundle, image, signature or release is uploaded.
- **Review:** Kimi K3 supplied the architecture and independent deep diff
  review. All current P2/P3 findings were fixed or explicitly bounded. Its two
  requested implementation runs produced design only, so the bounded core was
  executed by a worker and fully reviewed/hardened by Sol. Claude Code remained
  unavailable because its OAuth session is expired and changed nothing.
- **Verified so far:** exact real SBOM commands/parity, focused package/root/
  security/workflow tests, end-to-end generator/verifier tamper tests, JSON,
  docs consistency and diff checks. Fresh arithmetic is 1,896 = 1,551 Go + 26
  Rust + 319 client. Full build/verify/client/Rust/security gates still run next.
- **Rollout:** Phase 7 toolchain/platform checkbox only: 34/59 overall and
  34/51 phase work; Phase 6 remains 6/7; production remains false. Signing,
  authenticated provenance/SBOM truth, tagged artifacts, publication,
  container reproducibility, staged deployment and go/no-go remain open.

---

## 2026-08-16 00:45 EEST GH-212 release-evidence contract → In Progress

- **Branch:** `feature/GH-212-release-evidence`
- **Issue:** https://github.com/NeaBouli/TrueRepublic/issues/212
- **Baseline:** clean exact `origin/main`/HEAD `79c2a6d`; GH-209 and its
  documentation closeout are complete and no TrueRepublic PR was open.
- **Scope:** add a strict repository-only release-candidate evidence manifest,
  deterministic normalized Go/client CycloneDX SBOMs, unsigned provenance,
  explicit local offline verification, digest-pinned maintained container base
  images, and a machine-readable Linux amd64/arm64 supported-platform contract.
- **Owner split:** Kimi K3 completed the secret-free read-only architecture
  analysis and owns the bounded primary implementation. Sol owns scope,
  release/security boundaries, full diff review, integration, complete tests,
  public claims and all GitHub writes. Claude Code receives one small focused
  helper check only.
- **Risk:** release-adjacent but repository-only. No tag, GitHub Release,
  registry push, signing key, signature, external attestation, binary/image
  publication, deploy, genesis freeze, network action or production approval.
- **Rollout claim:** only Phase 7 pinned toolchains/supported platforms may be
  credited if fully proven. Signed artifacts, published SBOM/provenance and
  reproducible tagged binaries/containers remain open.
- **Tests:** not run yet; this is the pre-implementation ownership record.
- **Next:** Kimi implements the bounded evidence/verifier/CI/docs block; Sol
  reviews and hardens it, Claude performs a small read-only cross-check, then
  the complete relevant repository/security/client/Rust/build gates run.

---

## 2026-08-15 19:45 EEST GH-209 protected closeout → PASS

- PR #210 exact reviewed head `297e927` passed all 25 reported contexts after
  all four valid documentation findings were fixed and resolved, then
  admin-squash-merged as exact main `3cc392e`; GH-209 is closed and its remote
  feature branch is deleted.
- Final-main Client `31895690591`, reproducible Linux `31895690584`, Docs
  `31895690565`, Go `31895690592`, Security `31895690558` and Pages
  `31895690006` all pass on the exact merge commit. GH-29 credits the completed
  recipient-substitution-safe binding while independent cryptographic/privacy/
  setup review remains open.
- Published status is 1,874 = 1,529 Go + 26 Rust + 319 client, rollout 33/59,
  phase work 33/51, Phase 6 6/7 and `production_ready: false`. Client anonymous
  submission stays hard false; direct payout-address linkability is explicit.
- Kimi K3 supplied the main architecture and implementation block; Sol reviewed
  and hardened every writing diff, executed the full local/protected chain and
  owned GitHub closure. Claude Code was unavailable due expired OAuth and made
  no changes.
- No production ceremony/artifact, private key, real account/fund,
  RPC/broadcast, deployment, release or network mutation occurred.

---

## 2026-08-15 18:56 EEST GH-209 recipient-bound anonymous rewards → Local PASS

- **Delivered:** both anonymous rating paths bind the canonical
  `truerepublic` reward recipient into the versioned `TrueRepublic/vote/v2`
  signal while the external nullifier stays recipient- and rating-independent.
  Reward payout is atomic with vote mutation and rejects malformed,
  non-canonical and blocked module-account recipients.
- **Compatibility:** the existing public `SignalHash` is reused, so the frozen
  circuit, CS, PK and VK are unchanged. Consensus version 2 has a registered
  governed v1→v2 no-op migration and an explicit upgrade regression test.
- **Review:** Kimi K3 supplied the large bounded implementation after its
  architecture/security review. Sol reviewed the complete writing diff,
  added the migration registration and checksum-strict client validation,
  reconciled threat/spec/rollout claims and reran all relevant gates. Claude
  Code was attempted for the small independent check but its OAuth session
  had expired; it made no changes and is not a project blocker.
- **PASS:** 1,529 Go cases with Build/Vet/Race/Coverage (`x/truedemocracy`
  64.3%), 26 Rust tests plus fmt/Clippy, 10 Node + 309 Vitest cases (4 skipped),
  client lint/build/budgets/high audit, real WASM-to-native ZKP compatibility,
  staticcheck, gitleaks, govuln policy/fixtures, docs consistency, governed
  four-validator upgrade recovery, all three IBC two-chain cases and the full
  eight-case multi-validator recovery matrix.
- **Status:** candidate arithmetic is 1,874 = 1,529 Go + 26 Rust + 319 client;
  rollout 33/59, phase work 33/51, Phase 6 6/7, production false. Direct payout
  linkability and independent cryptographic/privacy review remain explicit.
- **Next:** commit and push the reviewed candidate, open the protected PR,
  resolve all review threads, require exact-head CI green, then merge and
  synchronize GH-29, final main, Pages and both Bridges.

---

## 2026-08-15 17:16 EEST GH-209 recipient-bound anonymous rewards → In Progress

- **Branch:** `feature/GH-209-zkp-reward-recipient`
- **Issue:** https://github.com/NeaBouli/TrueRepublic/issues/209
- **Baseline:** clean exact `origin/main`/HEAD `de098ff`; GH-206 and its
  closeout are complete, no pull request was open when this task started.
- **Scope:** bind a canonical reward recipient into a new domain-separated vote
  signal while preserving the recipient-independent external nullifier; cover
  Groth16 and domain-key anonymous ratings, atomic payout, adversarial replay/
  substitution/accounting evidence, encoding parity, consensus-version and
  privacy documentation.
- **Kimi contribution:** completed the required secret-free read-only
  architecture/security review. It confirmed the existing `SignalHash` public
  input is sufficient, so the frozen circuit, CS, PK and VK need no change.
- **Safety:** high risk because message/consensus, cryptography and token
  accounting are involved. Client submission remains hard false. No ceremony,
  production artifacts, keys/accounts/funds, RPC/broadcast, release,
  deployment or network mutation is in scope.
- **Tests:** not run yet; this is the pre-implementation ownership record.
- **Next:** delegate the bounded main implementation to Kimi, review every
  writing diff in Sol, use Claude Code for one small independent check, then
  run focused adversarial and complete repository gates.

---

## 2026-08-14 02:56 EEST GH-206 protected closeout → PASS

- PR #207 exact head `a2b3d17` passed all 25 reported contexts, including
  CodeRabbit/DeepScan, real ZKP WASM compatibility, Go vulnerability/static/
  secret/Rust/client audits, both reproducible Linux architectures, IBC,
  upgrade, Docker, capacity, concurrency, full Race/Coverage and the 15-minute
  multi-validator recovery gate. All nine review threads were resolved with
  evidence before merge.
- PR #207 squash-merged as exact main `b9e0a06`; GH-206 closed and the remote
  feature branch was removed. Final-main Client `31754724759`, Go
  `31754724837`, Security `31754724800`, Docs `31754724828`, reproducible Linux
  `31754724745`, dependency graph `31754727509`, and Pages `31754723966` pass.
- Cache-busted live Pages publishes 1,854 = 1,511 Go + 26 Rust + 317 client,
  rollout 32/59, Phase work 32/51, Phase 6 6/7, and `production_ready: false`.
  `isSubmittable` remains literal hard false; no production ceremony, keys,
  funds, wallet/RPC/broadcast, deployment, release or network mutation occurred.
- Kimi K3 was attempted twice through the required wrapper but provider quota
  HTTP 403 prevented contribution. Independent review plus hosted review found
  the remediated issues; Sol owned all changes, testing, merge and verification.
- **Next:** publish this append-only closeout record through its own protected
  documentation PR, update GH-29 without crediting a rollout checkbox, and
  verify the resulting final exact main and Pages.

---

## 2026-08-14 02:25 EEST GH-206 PR review remediation → Locally verified

- Reviewed all nine CodeRabbit threads against exact PR #207 head `dfc7fe3`.
  Hardened strict request decoding with a bounded JSON nesting depth, made both
  native compatibility checks consume the committed manifest's fixed VK
  fingerprint, and made the WASM gate execute only the `npm ci`-installed
  Vitest binary. The already-correct 594/317 detailed counts were retained.
- Split completed local evidence from pending publication work and aligned the
  ZKP status, threat model, roadmap, FAQ and testing wiki on one explicit
  test-only/non-production boundary.
- **PASS:** focused prover tests, real maintained-client WASM-to-native proof,
  docs consistency, diff hygiene, and complete `make verify` Build/Vet/Race/
  Coverage over all 16 selected packages under Go 1.26.6. Prover coverage
  increased from 72.6% to 74.8%.
- **Next:** publish the review-fix head, resolve the verified threads with
  evidence, and require the replacement exact head to pass every protected
  check before merge.

---

## 2026-08-14 02:18 EEST GH-206 Go 1.26.6 remediation → Locally verified

- Synchronized the current Go toolchain contract from 1.26.5 to 1.26.6 across
  `go.mod`, hosted workflows, Docker, deterministic-build metadata/verifier,
  the security source of truth, the dedicated WASM compatibility gate and
  current operator/developer documentation. Historical append-only records
  remain unchanged and no vulnerability exception was added.
- **PASS:** complete `make verify` build/Vet/Race/Coverage over all 16 selected
  Go packages under Go 1.26.6; the real maintained-client WASM-to-native proof
  gate; deterministic-daemon verification; exact active govuln policy plus
  positive/negative fixtures; documentation consistency; workflow YAML, JSON,
  shell and diff hygiene.
- The refreshed official database now returns exactly the four already
  documented reachable upstream findings without available fixes. The
  maintained `golang.org/x/net` remains fixed at v0.55.0.
- **Next:** commit and push this bounded hosted-security remediation, then
  require every check and review thread on the new exact PR #207 head to pass
  before merge.

---

## 2026-08-14 01:39 EEST GH-206 hosted security remediation → In Progress

- PR #207 exact head `def3ee1` made the false-positive secret scan green by
  disambiguating public pinned-artifact digest identifiers without changing
  bytes or verification behavior.
- The refreshed official Go vulnerability database then rejected Go 1.26.5
  for six reachable standard-library findings, all fixed in Go 1.26.6. The
  maintained `golang.org/x/net` is already fixed at v0.55.0; no new exception
  was added.
- Candidate remediation synchronizes Go 1.26.6 across toolchain, all hosted
  workflows, Docker, deterministic-build contract/verifier, security source of
  truth and current operator/developer docs. Focused vulnerability policy plus
  its negative contract, ZKP WASM end-to-end gate, docs and diff hygiene pass.
- **Next:** rerun the complete local chain under Go 1.26.6, commit/push, then
  require a fully green exact PR head before any merge.

---

## 2026-08-14 01:16 EEST GH-206 test-only maintained-client prover → Local review complete

- **Delivered locally:** extracted the frozen circuit into the dependency-light
  `zkpcircuit` leaf, added an exact-artifact Groth16 engine plus a dedicated
  Go `js/wasm` command, and added a strict maintained-client runtime/prover
  adapter without wiring it into the production bundle or transaction flow.
- **Compatibility evidence:** `./scripts/test-zkp-wasm-client.sh` builds an
  isolated 18.8 MB Go WASM command with Go 1.26.5, rejects Cosmos/Comet/Pebble/
  wasmd imports, generates a fresh proof through TypeScript from the unchanged
  synthetic GH-198 fixture, and proves native Go verification. Tampered proof,
  artifact drift, unsafe metadata, invalid witness/path/field, leaf/identity
  mismatch, malformed/trapped runtime, and concurrent stateless calls are
  covered fail closed.
- **Complete local evidence:** `make verify` passes build, Vet, and Race/
  Coverage across all 16 repository Go packages; 26 Rust tests plus fmt/Clippy,
  10 Node + 307 Vitest standard cases, client lint/build/budgets/high audit,
  dedicated WASM compatibility, strict circuit parity, docs consistency,
  JSON/YAML/shell/format/diff hygiene all pass. Public arithmetic is 1,854 =
  1,511 Go + 26 Rust + 317 maintained-client; rollout intentionally remains
  32/59 and `production_ready` remains false.
- **Review:** independent read-only review found and Sol remediated strict-spec
  path drift, proof-tamper/shape, identity-leaf binding, runtime-concurrency,
  CI-filter, and bundle-metric findings. Kimi K3 was invoked twice through the
  required wrapper but both attempts returned provider quota HTTP 403 and made
  no changes; this is not a project blocker.
- **Safety:** exact pinned artifacts were neither modified nor regenerated.
  `isSubmittable` remains literal hard false; no production ceremony, key,
  account/fund, wallet, RPC, broadcast, deploy, release or network action.
- **Next:** create the reviewed commit and protected PR, require all exact-head
  hosted checks and zero unresolved review threads, merge, then synchronize
  final main, Pages, GH-206/GH-29 and both Bridges.

---

## 2026-08-14 00:29 EEST GH-206 test-only maintained-client prover → In Progress

- **Branch:** `feature/GH-206-zkp-client-prover`
- **Issue:** https://github.com/NeaBouli/TrueRepublic/issues/206
- **Scope:** add an isolated real Groth16 proving path against the exact
  GH-198/GH-203 synthetic fixtures, with strict cross-language/browser-to-Go
  compatibility evidence and fail-closed malformed/drift/runtime handling.
- **Owner split:** Kimi K3 receives the larger secret-free feasibility and
  implementation slice; Sol owns architecture, cryptographic boundaries,
  writing-diff review, full local/protected gates, GitHub writes and closure.
- **Risk:** High because this handles private witness material and Groth16
  artifacts. `isSubmittable` must remain hard false; no ceremony, production
  key, real identity/account/fund, public RPC, broadcast, release or deployment.
- **Initial evidence:** exact `origin/main`/HEAD `09f6eeb`, clean worktree, no
  open PR, GH-203/Main/Pages closure green. No implementation file changed yet.
- **Next:** independently inventory the smallest browser-compatible prover
  architecture, choose an auditable bounded design, then implement and test it.

---

## 2026-08-13 22:30 EEST GH-203 protected closeout → PASS

- PR #204 exact head `87c2e8d` passed all 24 reported contexts with both review
  threads resolved, then admin-squash-merged as `c518fdc`; GH-203 closed and no
  pull request remains open.
- Final-main Go CI `31734145539`, Security `31734145526`, Client Web
  `31734145572`, Docs `31734145563`, Reproducible Linux `31734145553`,
  Dependency Graph `31734148836`, and Pages `31734144842` all pass on the exact
  merge commit. Cache-busted live Pages publishes 1,837 cases and 32/59 with
  `production_ready: false`.
- Kimi implemented and independently reviewed the bounded circuit/encoding
  freeze; Sol reviewed every writing diff, remediated all valid findings, and
  reran the complete local and protected integration/security chain. This is a
  compatibility foundation only: no production ceremony, prover, submission,
  keys, funds, release, deployment, or rollout checkbox was authorized.

## 2026-08-13 21:38 EEST GH-203 PR #204 → Review fixes in progress

- Exact head `184139f` passed all 24 reported PR contexts, including the
  14m35s multi-validator recovery gate, CodeRabbit and DeepScan. CodeRabbit
  then left two actionable threads plus three valid hardening nits.
- Sol applied the bounded review fixes: distinguish 1-32-byte inputs from their
  32-byte canonical field encoding, document fresh `-count=1` tests, preserve
  tokenizer errors, parse `go.mod` structurally with `x/mod/modfile` while
  rejecting replacements/exclusions, and explicitly classify committed keys
  as randomized single-party toxic-waste test artifacts.
- Focused GH-203 tests still produce exactly 27 PASS lines; 20 randomized-order
  drift repetitions pass. Kimi's targeted review approved the fixes; Sol also
  rejected mixed direct/indirect module duplicates. Full Go race/coverage,
  pinned security tools, module verification, docs consistency and diff hygiene
  pass again. Rollout stays 32/59 and production false.
- **Next:** commit/push a new PR head,
  resolve both review threads, and require all exact-head checks again.

## 2026-08-13 21:18 EEST GH-203 ZKP circuit-identity freeze → Locally merge-ready

- Sol reviewed Kimi's implementation line by line. Kimi's independent review
  found a P2 canonical-field gap in the pre-existing Go hash helper and P3
  strict-JSON/comment truth gaps; Sol fixed them by rejecting BN254 values at
  or above the scalar modulus, rejecting duplicate JSON keys recursively, and
  correcting the disabled mock-flow comment. Kimi's final re-review reports no
  remaining P0/P1/P2/P3 and a merge-ready verdict.
- Final local evidence passes: `make build && make verify` (1,500 Go PASS lines,
  race/coverage), 26 Rust tests, 10 Node + 301 Vitest tests, client lint/build/
  bundle/high-audit, pinned govulncheck/staticcheck/gitleaks positive and
  negative contracts, Rust audit with five allowed warnings and no blocking
  vulnerability, Go module verification, docs consistency, JSON and diff
  hygiene.
- Source-of-truth arithmetic is 1,837 = 1,500 Go + 26 Rust + 311 maintained
  client. Rollout remains 32/59 and production remains false. No PK/VK/proof
  regeneration, ceremony, prover, submission, keys/funds or deployment.
- **Next:** commit and publish the protected GH-203 PR, require green exact-head
  checks and zero unresolved review threads, merge, then verify final main,
  Pages, GH-29 and both Bridges.

## 2026-08-13 20:35 EEST GH-203 ZKP circuit-identity freeze → Local candidate delivered

- Kimi delivered the bounded secret-free slice on `test/GH-203-zkp-circuit-freeze`:
  strict versioned `configs/security/zkp-circuit.json` plus fail-closed
  `x/truedemocracy/zkp_circuit_spec_test.go` cross-checks against Go
  constants/behavior, the GH-198 fixture manifest/golden vector, the active
  `go.mod` gnark v0.14.0 / gnark-crypto v0.19.2 direct requires, and the
  maintained-client `zkpEncoding` contract with hard-false `isSubmittable`.
- In-memory recompilation of `MembershipCircuit` serializes byte/hash-identical
  to pinned `membership_v2.cs` non-destructively; PK/VK/proof bytes are never
  regenerated or compared (groth16.Setup samples random toxic waste).
- Focused gates pass: six new Go test functions (27 PASS lines with subtests),
  full truedemocracy package, 5 maintained-client zkp Vitest cases, vet/gofmt,
  docs consistency, diff hygiene. Rollout remains 32/59, production false.
- **Next:** Sol line-by-line integration/security review, complete local gates,
  candidate-arithmetic recount, independent re-review, protected PR and
  final-main verification.

## 2026-08-13 19:55 EEST GH-203 ZKP circuit-identity freeze → In Progress

- **Branch/issue:** `test/GH-203-zkp-circuit-freeze`,
  https://github.com/NeaBouli/TrueRepublic/issues/203 from exact verified main
  `21c2e1f` with no open PR or local diff.
- **Scope:** add a strict versioned circuit/encoding contract, deterministic
  compiled constraint-system parity, active `go.mod` toolchain binding, and
  fail-closed Go/fixture/client drift tests. Kimi K3 owns one bounded
  secret-free implementation slice; Sol owns architecture, cryptographic
  safety judgment, diff review, full verification, publication and closure.
- **Hard boundary:** Groth16 setup samples random toxic waste. PK/VK byte
  reproducibility, ceremony provenance, real prover integration, submission,
  reward payout, real keys/funds, deployment, release and network mutation are
  excluded. `isSubmittable` remains hard false; rollout remains 32/59 and
  production remains false.
- **Next:** Kimi implementation, Sol line-by-line integration/security review,
  complete local gates, independent re-review, protected PR and final-main
  verification.

## 2026-08-13 — Public security audit handoff — OPEN

- A private VLABS operator audit records unresolved CI-hygiene and dependency-
  coverage work. No live secret compromise was confirmed in the covered current
  checkout.
- Obtain a bounded, no-secrets task from the operator and record only sanitized
  completion evidence. Do not publish finding details, raw matches or private
  runtime evidence in this Bridge.

## 2026-08-10 02:15 EEST GH-172 → Merged / Final Main Green / Done

- **Merge:** PR [#173](https://github.com/NeaBouli/TrueRepublic/pull/173)
  exact head `671ff37` passed all 18 contexts with zero open review threads and
  squash-merged as `293ea35`. GH-172 is closed and the remote task branch is
  deleted. Admin merge bypassed only formal self-review, never a technical gate.
- **Final `main`:** Go CI `31340763723`, Security Scan `31340763717`,
  reproducible Linux `31340763708`, Docs `31340763710`, and Pages
  `31340763413` pass. The complete final-main Go job matrix is green.
- **Public/tracker:** live Pages publishes 1,607 cases (1,440 Go, 26 Rust, 141
  client), GH-172 evidence, 23/59 rollout, and `production_ready=false`. GH-29
  links GH-172/PR #173 without creating a 60th tracker unit.
- **Outcome:** bounded concurrent-submission, exact-replay and same-home
  restart evidence is Done. No production, deployment, public network, real
  key/account/funds, migration or release action occurred.

## 2026-08-10 01:45 EEST GH-172 → PR #173 Published

- Commit `972e324` is pushed on `test/GH-172-concurrency-replay`; protected
  [PR #173](https://github.com/NeaBouli/TrueRepublic/pull/173) targets `main`
  and closes GH-172 on merge.
- The PR contains only the reviewed GH-172 test/CI/evidence scope. Protected
  exact-head workflows and review threads are now the active gate; no check
  will be bypassed. Rollout remains 23/59 and production remains false.

## 2026-08-10 01:40 EEST GH-172 → Review Complete / PR Ready

- **Independent review:** Kimi reproduced the real harness twice with opposite
  contention winners and found no P0/P1/P2. Sol remediated its one P3 by
  requiring post-restart Cosmos code 32 plus `account sequence mismatch`.
- **Remediated evidence:** focused repository contract and the full real
  harness pass; the latter completed in 90.93s with exact one-of-three duplicate
  acceptance, unchanged state/economics and valid restart/export/reimport.
- **Status:** no open local finding or blocker. Protected exact-head PR checks,
  review-thread closure, merge, GH-29/public synchronization, final-main and
  Pages remain before Done. Rollout remains 23/59; production remains false.

## 2026-08-10 01:30 EEST GH-172 → Locally Verified / Review

- **Implementation:** repository-owned opt-in harness proves single-winner
  authenticated contention, exact signed-byte duplicate/replay rejection,
  same-home restart persistence, historical app-hash stability, singular
  state/economic effects, ledger reconciliation and export/reimport.
- **Evidence:** focused contract, build, Vet, full Race/Coverage, generative
  quality, security contract, pinned Staticcheck and Gitleaks, docs, JSON, YAML
  and diff checks pass. The real harness passed in 145.38s; standard arithmetic
  is 1,607 (1,440 Go + 26 Rust + 141 client), with the process gate separate.
- **Roles:** Kimi produced the bounded initial harness; Sol reviewed every
  write and corrected fatal test polling onto the main test goroutine. Claude's
  small read-only census attempt exceeded its budget and made no diff. Kimi is
  now performing the required independent read-only final review.
- **Status/safety:** no blocker. Rollout stays 23/59 and production false;
  protected PR/CI, review closure, merge, tracker synchronization, final-main
  and Pages verification remain before Done. No production, deployment, public
  network, real key/account/funds, migration or release action occurred.

## 2026-08-10 00:47 EEST GH-172 concurrent/replay/restart evidence → In Progress

- **Issue/branch:** [GH-172](https://github.com/NeaBouli/TrueRepublic/issues/172),
  `test/GH-172-concurrency-replay` from exact `origin/main` `348d622`.
- **Scope:** close the detailed Phase-5 evidence gap for shared-state concurrent
  authenticated submissions, duplicate messages, exact replay attempts, and
  deterministic persistence across restart/export-import with bounded,
  repository-owned tests and ephemeral localhost evidence.
- **Baseline:** no open PR; canonical main is green on Docs `31336988633`,
  Security `31336988646`, and Pages `31336987812`. Existing GH-97 load uses
  independent accounts and GH-145 adds focused replay/race properties, so
  GH-172 must add non-duplicative shared-state and persistence assertions.
- **Roles/safety:** Sol owns architecture, security, integration, every external
  write, complete verification, merge, and closure. Kimi receives one bounded
  secret-free implementation/review slice and may not delegate. No production,
  deployment, public network, real keys/accounts/funds, wallet custody,
  consensus migration, or release action is authorized.


## 2026-08-10 00:25 EEST GH-169 merged and final-main verified → Done

- PR [#170](https://github.com/NeaBouli/TrueRepublic/pull/170) exact head
  `620c41d` passed 18/18 checks after all five review threads were resolved and
  merged by owner-authorized admin squash as `d50c131`. The bypass covered only
  formal self-approval; no technical check or review finding was bypassed.
- Exact merged-main runs pass: Go CI `31336109586`, Security Scan `31336109601`,
  Reproducible Linux Daemon `31336109590`, Docs `31336109615`, and Pages
  `31336109276`. GH-169 is closed and its remote implementation branch is gone.
- GH-29 now checks the Phase-5 threat-model item and links GH-169/PR #170. Live
  Pages readback confirms 1,606 tests, 23/59 overall, 23/51 phase work, 39%/45%,
  GH-169, Vite 8.2 and the unchanged non-production boundary. Phase 5 is 5/6;
  Phase 6 remains 6/7; `production_ready=false`.
- No deployment, mainnet, production, wallet/key/signing, consensus-state
  migration, release publishing or real-funds action occurred. The independent
  external security review remains the final open Phase-5 item.

## 2026-08-09 23:54 EEST PR #170 review remediation → Verified

- CodeRabbit's five inline findings were reproduced and fixed: remaining IBC
  claims are explicitly unverified, CLAUDE/treasury versions and counts match
  source truth, module checks parse active `go.mod` requirements instead of raw
  text, and the threat-model parser/test wording is format-independent and uses
  robust EOF comparison. Two related low-value test nits were also resolved.
- The remediated tree passes the focused 20-case threat-model contract,
  documentation consistency including parsed SDK/CometBFT/wasmd/ibc-go/Vite
  parity, complete build/vet/Race/Coverage, Staticcheck, maintained-tree secret
  scan, module verification, JSON and diff integrity.
- The first PR head reached 17/18 green before remediation; no failed check was
  ignored. A new exact head and its complete protected CI/review chain are next.

## 2026-08-09 23:38 EEST GH-169 cross-system threat model → Locally Verified

- **Implementation:** added canonical `truerepublic.threat-model/v1` JSON and
  explanatory Markdown for nine maintained security domains, 18 stable threat
  IDs, explicit assets/actors/trust boundaries/data flows, verified controls,
  evidence paths, residual risk, durable owners and GH-7/GH-29 next gates.
- **Fail-closed contract:** the root repository test strictly rejects malformed
  or trailing JSON, unknown schema/enums, missing/unsafe/nonexistent evidence,
  duplicate/empty IDs, missing controls/owners/gates, invalid umbrella mappings,
  high residual risk on closed threats and any production-ready claim. All 19
  negative mutations pass.
- **Truth synchronization:** corrected stale SDK/CometBFT/wasmd/Vite versions,
  GH-131 submitted-history status and IBC wording; consistency now binds SDK,
  CometBFT, wasmd, ibc-go and Vite to lock/module truth. Fresh package-scoped
  JSON evidence records 1,439 Go cases (root 153), making 1,606 total. Rollout
  candidate is 23/59 overall and 23/51 phase work; Phase 6 stays 6/7 and
  `production_ready=false`.
- **Review/tests:** Kimi independently verified all 33 evidence paths and found
  no P0/P1. Its one P2 stale multi-client claim and two actionable P3 hardening
  notes were fixed. Sol reran the focused contract, documentation consistency,
  complete build/vet/Race/Coverage `make verify`, security contract,
  Staticcheck, maintained-tree secret positive/negative gates, Go vulnerability
  positive/negative gates, module verification, JSON parsing and diff checks;
  all pass.
- **Risk/next:** this is repository self-assessment, not independent audit or
  production approval. Protected PR/CI, merge, GH-29 synchronization, final-main
  and Pages readback remain before Done. No production, mainnet, deployment,
  key, wallet, signing, migration or funds action occurred.

## 2026-08-09 23:12 EEST GH-169 cross-system threat model → In Progress

- **Issue/branch:** [GH-169](https://github.com/NeaBouli/TrueRepublic/issues/169),
  `security/GH-169-threat-model` from exact verified `origin/main` `f6619e6`.
- **Scope:** define maintained assets, actors, trust boundaries and data flows;
  add a versioned secret-free threat register for consensus/P2P, governance,
  token/treasury/DEX, ZKP/privacy, IBC/upgrades, client/wallet/RPC, operations,
  dependencies/CI and release artifacts; enforce it with fail-closed repository
  tests and negative mutations.
- **Safety boundary:** repository documentation, configuration and tests only.
  This is not an external audit and does not authorize consensus-state,
  cryptography, wallet-custody, keys, signing, funds, deployment, mainnet,
  production, migration or release activity.
- **Roles/next:** Sol owns architecture, security judgment, integration,
  complete verification and GitHub closure. Kimi receives one bounded
  secret-free implementation/review slice; Sol reviews every diff and reruns
  the complete relevant gates before any status or rollout claim changes.

## 2026-08-09 14:45 EEST GH-161 bounded dependency reconciliation → In Progress

- **Issue/branch:** [GH-161](https://github.com/NeaBouli/TrueRepublic/issues/161),
  `fix/GH-161-dependabot-reconciliation` from exact verified `origin/main`
  `b152fb13dc4303859e078a94c0b8e526752c7ea7`.
- **Scope:** independently reproduce and reconcile Cargo PR #156, npm PR #157,
  and Go PR #158. Merge only compatible, fully reviewed updates; otherwise
  repair or replace the generated group and record the concrete incompatibility.
- **Safety boundary:** no automatic merge, semver-major widening, production,
  deployment, network, mainnet, wallet, key, signing, migration, or funds
  action. The six documented historical/closed-unmerged preservation
  worktrees and branches remain untouched.
- **Current evidence:** #156 is green; #157 has one failing client build; #158
  has failing Go build/recovery/capacity plus vulnerability/static gates. Kimi
  receives a secret-free read-only compatibility review while Sol owns all
  edits, integration, protected publication, and final verification.

### 2026-08-09 15:00 EEST triage and safe replacement evidence

- PR #156 was refreshed onto current `main`, passed all ten exact-head checks
  and no review thread, then merged as `4aeddc0267bffd8105ff2182892b72419c6fb88b`.
  Local exact-head Rust fmt, Clippy, build, all 26 unit tests, and doc tests pass.
- PR #157 is rejected as a group: `eslint-plugin-react-refresh` 0.5.3 makes 19
  existing lazy-route exports fail lint, and Playwright 1.62.1 exceeds the
  frozen macOS 12 browser boundary. A seven-package replacement retaining both
  pins passes lint, 141 unit/policy cases, build/budget, audit, chain harness
  compilation, and 26 local Chromium/Firefox browser cases. Ubuntu WebKit CI
  remains the authoritative engine gate.
- PR #158 is rejected as a protocol migration: Cosmos SDK 0.50.14→0.53.0 and
  related 0.x changes disable expected invariant behavior, alter export output,
  break recovery/capacity suites, add static deprecations, and expose three new
  reachable vulnerability IDs. A four-direct-dependency patch-only replacement
  passes full `make verify`, module verification, pinned staticcheck, pinned
  govulncheck policy, and the repository security contract.
- Spark implemented the bounded typed policy patch in `.github/dependabot.yml`
  and `security_gate_repository_test.go`; Sol reviewed the diff and repeated
  the focused tests. Kimi approved the three-slice review; Sol also closed its
  P2 by fully freezing automatic Playwright version updates. Protected
  publication remains.

### 2026-08-09 16:46 EEST merge chain and final-main verified

- Policy PR #162 passed all 18 exact-head contexts and merged as `7b3f183`.
  The compatible npm replacement PR #163 was then rebased, passed all 15
  exact-head contexts including browser and real client-chain integration, and
  merged as `3a3919c`. Go replacement PR #164 passed all 18 contexts including
  14m09s multi-validator recovery and merged as `62ad5e9`. The policy then
  generated patch-only Go PR #167; independent Sol/Kimi review found no
  P0/P1/P2, all 16 contexts passed, and it merged as final code head `87c47c1`.
  No review thread remained open on any merged replacement.
- Exact merged `main` is green: Go CI `31315620763`, reproducible Linux daemon
  `31315620777`, Security Scan `31315620755`, Dependency Graph `31315623437`,
  and Pages `31315620166`. Local composed-head Go build/vet/race/coverage,
  module and sumdb verification, typed security contract, npm
  install/lint/141 tests, build/budget, and audit also pass.
- Closeout PR #168 failed closed once while RustSec's live advisory database had
  a malformed `gettext-rs` path. Upstream corrected it as `e0bc1e8`; an
  unchanged-head failed-job rerun then passed, without weakening the gate.
- PRs #157 and #158 were closed with concrete incompatibility evidence and
  links to #163/#164. The maintained policy now keeps Go automation patch-only,
  freezes automatic Playwright movement, keeps React Refresh on compatible
  patches, and removes Cargo `cosmwasm-*` crates from grouped maintenance.
  Automatically superseded PRs #165/#166 closed without merge; only this
  closeout PR remains open.
- Rollout remains honestly unchanged at 22/59 overall, 22/51 phase work,
  Phase 6 6/7, 1,579 recovery-verified cases, and `production_ready=false`.
  GH-161 is complete after this docs-only closeout reaches protected `main`.
  No production, deployment, mainnet, wallet, key, signing, migration, or
  funds action occurred.

## 2026-08-09 14:32 EEST GH-154 protected publication → Done

- PR #159 passed all 11 exact-head checks on
  `46f687be0f57774dcc8c970154246371a90759f3` and was squash-merged as
  `ac8fbbcfabdc87e9293c30ddffc1cf320cac3aca`; issue #154 is closed and the
  remote task branch is deleted.
- Fresh exact-main Docs Consistency Check `31310820782`, Security Scan
  `31310820753`, and Pages `31310820227` all completed successfully on the
  same merge commit. The cleanup audit is therefore fully published and
  GH-154 is Done.
- The preservation boundaries remain unchanged: six intentional remote heads
  and six historical/closed-unmerged local worktrees after this clean closeout
  worktree is retired. Active Dependabot PRs #156-#158 are separate bounded
  maintenance tasks and were not modified or auto-merged by GH-154.

## 2026-08-09 13:58 EEST GH-154 lossless repository reconciliation → In Progress

- **Issue/branch:** [GH-154](https://github.com/NeaBouli/TrueRepublic/issues/154),
  `chore/GH-154-reconcile-residue` from exact merged `origin/main`
  `8489d5477db695b81cb6626a96f44204aa1ade5a`.
- **Precondition:** GH-153/PR #155 is merged and issue #153 is closed. Fresh
  exact-main workflows are still running, so GH-153 is not yet recorded as
  fully closed.
- **Safety invariant:** re-fetch and revalidate each candidate immediately
  before cleanup. Remove only a remote ref whose current SHA exactly equals a
  merged PR head, or a clean local worktree whose HEAD exactly equals that
  merged PR head. Preserve all dirty, unique, archival, closed-unmerged,
  active, or ambiguous state.
- **Current action:** repeat the prior read-only inventory against current
  GitHub and local state. No residue has been deleted in GH-154 yet.

### 2026-08-09 14:05 EEST exact-remote reconciliation complete

- A fresh GitHub/remote comparison reconfirmed 31 branch heads as byte-exact
  heads of merged PRs. Every candidate was checked again immediately before
  deletion and none had an open PR. All 31 exact merged residues are now
  absent from the remote.
- The final remote inventory intentionally preserves exactly six heads:
  `main`; active Dependabot PR branches #156, #157, and #158; closed-unmerged
  `agent/public-recovery-status`; and `solana-archived`.
- Local cleanup is still running. Seventeen candidates were clean and exact
  merged-PR heads before removal began. The dirty/divergent original checkout,
  four clean but historically divergent worktrees, closed-unmerged recovery
  status, and the active GH-154 worktree remain preservation boundaries.
- Older worktrees contain large ignored `node_modules`/Rust `target` trees, so
  safe non-forced `git worktree remove` operations are filesystem-bound. No
  additional cleanup is started while those operations remain active.

### 2026-08-09 14:05 EEST GH-153 merged / final-main verification

- PR #155 exact head `3f698a2924493579dafde7e91a9b9ee1cd6a847b`
  passed 18/18 protected checks and was squash-merged as
  `8489d5477db695b81cb6626a96f44204aa1ade5a`; issue #153 is closed and its
  remote task branch is deleted.
- Fresh exact-main Go `31309568219`, Security `31309568223`, Reproducible Linux
  `31309568233`, Docs `31309568209`, Dependency Graph `31309569964`, and Pages
  `31309567545` all pass; the five-job Go run includes successful bounded
  multi-validator recovery. GH-153 is Done.
- The bounded policy correctly generated minor/patch PRs #156-#158. Cargo and
  Go generator runs also report separate updater/resolution failures for
  CosmWasm and an incompatible Wasmd/SDK candidate. These are explicit
  follow-up supply-chain findings; no generated update is auto-merged.

### 2026-08-09 14:12 EEST final inventory → Local Complete / PR Ready

- **Remote:** after final fetch/prune exactly six heads remain: `main`
  `8489d54`; active PR #156 `acc37e8`; active PR #157 `2b59b32`; active PR
  #158 `f124156`; closed-unmerged recovery status `e756d0b`; archival Solana
  `043c62e`. There is no stale exact-merged remote residue.
- **Removed locally:** 17 worktrees whose pre-removal state was clean and whose
  HEAD exactly matched a merged PR head. Removal used `git worktree remove`
  without force. Large ignored dependency/build trees explained the bounded
  filesystem delay; no cleanup process remains.
- **Preserved locally:** the legacy checkout (`78d7ce3`, 18 status entries, 24
  commits outside current origin/main); clean but historically divergent GH-10
  (`63960e3`, 24), GH-12 (`6d307d8`, 28), GH-20 (`cd4f3e6`, 31), and GH-8
  (`71dfa9d`, 41); closed-unmerged recovery status (`e756d0b`, one); and the
  active GH-154 worktree. Parenthesized counts are commits outside current
  origin/main. None was modified.
- **Checks:** final remote/open-PR/worktree/dirty inventory, documentation
  consistency, and `git diff --check` pass. Next: independent final review,
  protected docs-only PR, merge, issue closure, exact-main Docs/Pages readback,
  then GH-154 Done.

### 2026-08-09 14:16 EEST independent final review → Approved

- Kimi returned **APPROVE** with no P0/P1/P2/P3 and changed no file. The review
  independently reproduced all six remote SHAs, all seven worktrees and their
  dirty/divergence counts, PR #25 as closed-unmerged, PRs #156-#158 as open,
  every cited exact-main run, docs consistency, and clean diff checks.
- An independent probe of 91 distinct merged-PR head branch names found zero
  remaining remote residue. The original 16-worktree baseline plus the GH-153
  worktree that became exact-merged after PR #155 explains the final 17 safe
  removals.
- Remaining work is publication only: commit, protected docs-only PR, merge,
  issue closure, final-main Docs/Pages and Bridge readback.

### 2026-08-09 14:22 EEST PR #159 review remediation

- CodeRabbit's two valid documentation findings are fixed: project state now
  separates the six remote heads from the seven local worktrees, and the TODO
  records independent review as complete while publication remains open.
- GH-153 closure evidence remains in this closeout because it is the missing
  final-main/Bridge handoff that immediately precedes and establishes GH-154's
  exact base. No runtime, dependency, workflow, or inventory claim changed.

## 2026-08-12 12:01 EEST GH-184 governed deterministic upgrade path → In Progress

- **Ticket:** `GIO-20260812-TR-GOVERNED-UPGRADE`; [GH-184](https://github.com/NeaBouli/TrueRepublic/issues/184).
- **Branch/baseline:** `feature/GH-184-governed-upgrade` from exact clean
  `origin/main` `a87eb5871b4c590a20621a463d37607889636723`; no unrelated
  work was present at takeover.
- **Scope:** replace the application-upgrade no-op boundary with a persisted
  `x/upgrade` plan, a fail-closed TrueRepublic 2/3 governance authorization,
  named deterministic migration, old-binary halt, same-state restart, and
  rollback/partial-failure evidence.
- **Delegation:** Kimi K3 is performing a bounded, secret-free read-only
  architecture/risk review. Sol owns architecture, every write, security,
  integration, complete verification, GitHub actions, merge, and closure.
- **Risk:** High/consensus-critical. Store layout, pre-block execution,
  authority selection, vote identity, threshold arithmetic, exact-once
  migration, and multi-validator recovery must remain deterministic.
- **Safety boundary:** repository-only; no production/public network,
  deployment, release, real validator/relayer, real key/account/funds,
  arbitrary migration, or IBC client-upgrade action. Production remains false.
- **Ready for:** bounded implementation, focused adversarial tests, full local
  gates, independent review, protected PR/CI, and final-main closeout.

## 2026-08-10 03:10 EEST GH-175 IBC two-chain evidence → In Progress

- **Ticket:** `GIO-20260810-TR-IBC-TWO-CHAIN`; GitHub issue #175.
- **Scope:** two isolated TrueRepublic application states with Tendermint IBC
  clients, connection and unordered ICS-20 channel handshakes, native PNYX
  transfer/escrow/voucher acknowledgement, timeout refund, duplicate packet
  rejection, and persistent-state interruption/recovery evidence.
- **Baseline:** exact `origin/main` `10629f1`; no open PR; 13 existing focused
  IBC configuration/stub tests pass. They do not yet exercise a two-chain
  packet lifecycle.
- **Boundary:** governance-controlled client upgrades, replacement of the
  explicitly unsupported x/upgrade path, external relayer qualification,
  deployment, real accounts, keys, or funds are not part of GH-175.
- **Delegation:** Kimi K3 receives a bounded, secret-free implementation block;
  Sol owns architecture, security, integration, full verification, GitHub
  writes, and closure.
- **Status:** In Progress; no blocker identified.

## 2026-08-10 04:05 EEST GH-175 → Local Implementation Verified

- **Finding/fix:** the first real client-creation transaction failed with
  Cosmos code 2 because the concrete Tendermint light-client types were absent
  from the application interface registry. `app.go` now registers them and a
  standard regression test pins the contract.
- **Evidence:** two persistent TrueRepublic apps complete client, connection,
  unordered channel, native escrow/voucher, acknowledgement, timeout refund,
  duplicate receive rejection, idempotent ACK/timeout replay, and a
  pending-ack database reopen followed by a fresh verified header.
- **PASS:** dedicated `make ibc-two-chain` (1.05s), focused IBC suite, 1,441
  standard Go cases, docs consistency, JSON/YAML, formatting, and diff checks.
- **Status:** Local Review; candidate rollout 25/59 and phase work 25/51.
  Full build/Vet/Race/Coverage, security gates, independent review, protected
  CI, merge, final-main, and Pages remain. External relayers, channel
  close/replacement, cross-upgrade behavior, production, real keys/accounts,
  and funds remain outside GH-175.

## 2026-08-10 04:32 EEST GH-175 → Replay Review Remediated

- **Independent review:** Kimi reproduced the dedicated harness and found one
  P2 evidence-quality issue: duplicate receive was stopping at a stale client
  proof before reaching ibc-go's receipt replay check.
- **Fix:** the harness now refreshes the destination client first, then proves
  the duplicate receive succeeds as ibc-go's intentional no-op with code 0 and
  no second voucher mint. Documentation now describes receive, ACK, and timeout
  replays consistently as economically idempotent no-ops.
- **PASS:** remediated `make ibc-two-chain` (1.17s), focused IBC suite, and diff
  integrity. No production, deployment, real account/key/fund, or external
  relayer action occurred.

## 2026-08-10 04:40 EEST GH-175 → Local Gates Green / Ready for PR

- **PASS:** remediated `make ibc-two-chain` (1.17s); focused IBC suite; full
  build/Vet/Race/Coverage (`truerepublic` 71.9%); Staticcheck v0.7.0; security
  repository contract; Gitleaks v8.30.1 with no leaks; govulncheck v1.6.0 with
  exact active no-fix policy and positive/negative fixtures.
- **PASS:** 1,441 standard Go cases (1,608 maintained total), documentation
  consistency, `status.json`, workflow YAML, `go mod tidy`, formatting, and diff
  integrity.
- **Review:** Kimi's single P2 evidence-quality finding is remediated; no P0/P1
  or unresolved P2 remains. Status advances to Ready for PR. Protected exact-
  head CI, review threads, merge, final-main, tracker, and Pages remain.

## 2026-08-10 12:48 EEST GH-175 / PR #176 → Review Remediated

- **First protected head:** commit `4033b9a` passed all 18 contexts, including
  the new `ibc-two-chain` job (3m39s) and multi-validator recovery (14m36s).
- **Review:** CodeRabbit opened nine actionable threads. Eight are fixed:
  historical header AppHash/time now come from height-keyed persistent commit
  data and the provider is rebound after database reopen; chain bootstrap
  operators are distinct; CI has five minutes of timeout headroom; the target
  uses the repository package selector; relayers are explicitly unqualified;
  the executable example and PNYX units are exact; and TM-IBC-001 uses the
  allowed `deferred` state.
- **Append-only disposition:** the thread on older `duplicate ... rejection`
  wording is answered by the later 04:32 correction entry and is not rewritten;
  the canonical result is idempotent receive/ACK/timeout no-op handling.
- **PASS on review diff:** `make ibc-two-chain` (root scenario 2.03s), full
  build/Vet/Race/Coverage, Staticcheck, security contract, docs consistency,
  JSON/YAML, formatting, and diff integrity. New exact-head CI remains before
  merge.

## 2026-08-10 13:18 EEST GH-175 / PR #176 → Merged / Final Main Green / Done

- **Merge:** reviewed head `bc03215` passed 18/18 protected contexts with zero
  unresolved review threads. Authorized administrative merge produced
  `e8ea2eb`; PR #176 merged, GH-175 closed, and the remote branch was deleted.
- **Final `main`:** Go CI `31377429576`, Security `31377429590`, reproducible
  Linux `31377429630`, Docs `31377429560`, and Pages `31377428474` PASS.
- **Public readback:** Pages publishes 1,608 maintained cases, rollout 25/59,
  phase work 25/51, Phase 6 6/7, and production false. GH-29 marks only the two
  proven Phase-3 IBC units complete; pending-ACK recovery is noted while
  post-upgrade/channel-close/timeout-on-close work stays open.
- **Status:** GH-175 Done. No deployment, public counterparty, external
  relayer, production, real account/key/fund, or infrastructure mutation
  occurred. Next IBC work is the still-open residual boundary above.

---

## 2026-08-09 13:19 EEST GH-153 bounded Dependabot maintenance → In Progress

- **Issue/branch:** [GH-153](https://github.com/NeaBouli/TrueRepublic/issues/153),
  `fix/GH-153-bounded-dependabot` from exact merged `origin/main` `5baeea6`.
- **Trigger:** GH-148's first scheduled update sweep opened PRs #150–#152 and
  exposed that wildcard groups also combine incompatible major dependency
  lines. Cargo attempted `schemars` 0.8→1.2 and unresolved CosmWasm updates;
  Go attempted incompatible Cosmos/CometBFT candidates; npm combined 23
  updates and fails build, browser, integration, and audit gates.
- **Scope:** restrict automatic grouped maintenance to compatible minor/patch
  updates, add a repository regression contract, classify/close the broken
  generated PRs, and require local plus protected GitHub evidence.
- **Boundary/risk:** Medium supply-chain configuration only. Major upgrades
  remain separately reviewed tasks. No runtime, protocol, consensus,
  wallet/key/signing, funds, deployment, network, mainnet, or production
  action is authorized.
- **Retro audit:** independent read-only audit classified 31 remote branches
  and 16 local worktrees as exact merged-PR residue; two remote branches and
  ten local worktrees remain active, unique, dirty, or ambiguous and are
  preserved. No cleanup has yet been performed.
- **Ready for:** Sol implementation, focused/full verification, independent
  review, protected PR, merge, final-main/Pages and Bridge closure.

---

### 2026-08-09 13:28 EEST policy correction and generated-PR cleanup

- **Fix:** every ecosystem group now explicitly applies only to version
  updates at `minor` or `patch` level, while a wildcard ignore prevents
  automatic `semver-major` PRs. Major dependency changes remain standalone,
  architecture-aware review tasks.
- **Regression:** `TestSecurityGateRepositoryContract` now rejects removal of
  any of the four major-update exclusions and requires all four compatible
  groups. Official GitHub option semantics were checked before implementation.
- **Focused PASS:** repository security-gate contract including the new
  negative case; Dependabot YAML parse; `git diff --check`.
- **GitHub cleanup:** failing generated PRs #150, #151, and #152 were closed
  with exact failure classification and only their automatic Dependabot
  branches were deleted. No dependency diff was merged.
- **Pending:** full repository verification, independent read-only review,
  commit/push/protected CI/merge, final-main evidence, then proven-safe merged
  branch/worktree cleanup.
- **Cleanup ticket:** [GH-154](https://github.com/NeaBouli/TrueRepublic/issues/154)
  now owns the separate lossless branch/worktree reconciliation so GH-153 stays
  an atomic supply-chain policy repair.

### 2026-08-09 13:28 EEST review hardening

- **Review response:** Kimi's first read-only pass identified one P3 robustness
  concern: literal string counts could accept correctly repeated rules in the
  wrong YAML entries. Sol replaced them with typed `yaml.v3` parsing that
  rejects invalid YAML, missing/duplicate/unexpected ecosystems, wrong group
  names, wrong nesting, extra/missing ignores, and non-minor/patch policy.
- **Module hygiene:** `go mod tidy` only reclassifies three already-pinned,
  already-used modules (`zerolog`, `testify`, `yaml.v3`) as direct; no version,
  checksum, runtime dependency graph, or production source changed.
- **Full PASS:** `make verify` passes build, vet, and all 1,412 Go cases under
  Race/Coverage across 13 packages; root 71.9%, DEX 51.1%, governance 63.8%.
  Documentation consistency, YAML parse, and diff checks also pass.
- **Security-update caveat:** the ordinary version-update policy suppresses
  automatic majors. Dependabot security updates may still cross a major line
  and remain separately reviewed, never auto-merged.
- **Pending:** final re-review on this hardened exact diff, then publication and
  protected/final-main evidence.

### 2026-08-09 13:39 EEST final local review → Approved

- **Kimi verdict:** APPROVE; no P0/P1/P2/P3 remains. The final read-only review
  reproduced the contract case, verified all four exact ecosystem structures,
  confirmed zero dependency/checksum/runtime change, and found Bridge state
  consistent. Kimi changed no file.
- **Final local PASS:** repeated `make verify` passes build, vet, and 1,412 Go
  cases under Race/Coverage; `go mod verify` and `go mod tidy -diff` pass;
  documentation consistency, typed Dependabot YAML, and diff checks pass.
- **Security PASS:** exact pinned gitleaks 8.30.1 scans 4.04 MB with zero leaks;
  planted-secret and Git-failure fixtures pass; exact pinned staticcheck 2026.1
  passes. The first attempt correctly stopped because the new worktree lacked
  scanner binaries; no false PASS was recorded.
- **Ready for:** commit, push, protected exact-head PR checks and merge. GH-153
  is not Done until final-main verification succeeds.

### 2026-08-09 13:43 EEST PR #155 published

- Commit `5f188e6` is pushed through
  [PR #155](https://github.com/NeaBouli/TrueRepublic/pull/155), closing GH-153
  only after merge. The PR records full local evidence and Kimi approval.
- Exact-head protected workflows and review are now mandatory. No check is
  bypassed; merge, final-main/Pages, issue closure, and Bridge synchronization
  remain pending.




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
- GH-55 validator identity recovery audit: [`GH55_AUDIT.md`](docs/agent-bridge/GH55_AUDIT.md)
- GH-56 consensus-key rotation audit: [`GH56_AUDIT.md`](docs/agent-bridge/GH56_AUDIT.md)
- GH-74 node health audit: [`GH74_AUDIT.md`](docs/agent-bridge/GH74_AUDIT.md)
- GH-77 structured logging audit:
  [`GH77_AUDIT.md`](docs/agent-bridge/GH77_AUDIT.md)
- GH-80 metrics baseline audit:
  [`GH80_AUDIT.md`](docs/agent-bridge/GH80_AUDIT.md)
- GH-85 observability operations audit:
  [`GH85_AUDIT.md`](docs/agent-bridge/GH85_AUDIT.md)
- GH-89 topology qualification audit:
  [`GH89_AUDIT.md`](docs/agent-bridge/GH89_AUDIT.md)
- GH-121 browser query transport audit:
  [`GH121_AUDIT.md`](docs/agent-bridge/GH121_AUDIT.md)
- GH-128 client bundle audit:
  [`GH128_AUDIT.md`](docs/agent-bridge/GH128_AUDIT.md)
- Decisions: [`DECISIONS.md`](docs/agent-bridge/DECISIONS.md)
- Security: [`SECURITY_NOTES.md`](docs/agent-bridge/SECURITY_NOTES.md)

GitHub recovery epic: [#4](https://github.com/NeaBouli/TrueRepublic/issues/4)

## 2026-08-09 00:14 EEST GH-145 generative quality depth → In Progress

- **Branch:** `test/GH-145-quality-depth` from clean `origin/main` `9481b87`.
- **Issue:** [GH-145](https://github.com/NeaBouli/TrueRepublic/issues/145)
- **Scope:** deterministic DEX constant-product properties, generative PNYX
  21,000,000 supply/malformed-genesis validation, governance replay safety,
  and bounded protected fuzz/race execution.
- **Delegation:** Kimi K3 owns only the three new secret-free Go test files.
  Sol owns architecture, CI contract, public arithmetic, security/integration
  review, all external writes, complete verification, and closure.
- **Boundary/risk:** Medium, tests/CI/evidence only. No production behavior,
  consensus transition, accounting or DEX formula, governance authority,
  wallet/key/signing, deployment, network, or real-funds change. A property
  failure becomes a separately scoped finding rather than an in-scope code fix.
- **Status/blockers:** In Progress; no blocker. Rollout remains 20/59 overall,
  20/51 phase work, Phase 6 6/7, and `production_ready=false`.

### 2026-08-09 00:24 EEST implementation integrated → full gates running

- **Kimi contribution:** added only the three assigned test files: deterministic
  and fuzzed DEX swap properties, PNYX cap/canonical-supply/malformed-genesis
  properties, and governance replay/no-drift properties. Sol reviewed every
  line, corrected the CI target names, kept fuzzing bounded, and added explicit
  invalid-address and supply/balance-mismatch genesis cases.
- **Focused evidence:** package tests, focused Race execution, the repository
  contract, and a one-second integrated campaign for each real fuzz target pass.
  Kimi's separate 20-second campaigns pass with 40,317 DEX and 25,524 token
  executions and no crasher.
- **Authoritative recount:** package-scoped `go test -json` records 1,406
  standard Go cases: root 120, token 37, DEX 138, governance 533, and unchanged
  remaining modules. Candidate public arithmetic is 1,573 total (1,406 Go + 26
  Rust + 141 maintained client), rollout 21/59 overall and 21/51 phase work;
  Phase 6 remains 6/7 and production remains false.
- **Pending:** full repository verification, critical coverage, documentation
  consistency, independent final review, protected exact-head PR checks,
  merge, GH-29 synchronization, final-main checks, and Pages readback.

### 2026-08-09 00:31 EEST full local verification → review ready

- **Full Go:** `make verify` passes package selection, build, vet, and all
  1,406 standard cases under Race/Coverage. Coverage is root 71.9%, token
  95.6%, DEX 51.1%, and governance 63.8%.
- **Quality gate:** the protected-equivalent 10-second campaigns pass with
  12,694 DEX and 11,537 token executions; focused deterministic tests remain
  under Race. Critical coverage, documentation arithmetic/inventory/rollout,
  JSON, workflow YAML, shell syntax, and diff checks pass.
- **Status:** local implementation and integration are complete with no
  production-code diff and no blocker. Independent final review is running;
  protected publication and post-merge evidence still gate Done.

### 2026-08-09 00:38 EEST independent review → approved

- **Verdict:** Kimi's read-only full-diff review independently reproduced the
  properties, replay-conflict semantics, package counts, token coverage,
  10-second fuzz gate, documentation arithmetic, vet, YAML/shell, and diff
  hygiene. Findings: 0 P0, 0 P1, 0 P2; no file was changed by the reviewer.
- **P3 handling:** increased only the outer GitHub job timeout from 10 to 15
  minutes for cold-cache compile headroom; the deterministic run and each fuzz
  campaign retain their 180-second and 1–60-second hard bounds. Staged rollout
  arithmetic remains explicitly conditional on protected merge, and observed
  fuzz execution counts remain run evidence rather than fixed thresholds.
- **Next:** rerun the affected repository/YAML/docs contracts, then publish the
  protected PR and require exact-head CI plus review closure before merge.

### 2026-08-09 00:41 EEST protected PR published

- **GitHub:** commit `d4aaa77` is pushed on
  `test/GH-145-quality-depth`; PR
  [#146](https://github.com/NeaBouli/TrueRepublic/pull/146) links and closes
  GH-145 only after merge.
- **Status:** exact-head protected workflows and review are running. No local
  or GitHub gate has been bypassed; merge, GH-29 synchronization, final-main
  checks, Pages readback, and final Bridge closeout remain before Done.

### 2026-08-09 01:11 EEST merged / final main and Pages green → Done

- **Exact-head:** every protected PR #146 check passed on `76b5d80`, including
  quality-depth 4m19s, build/Race/coverage 8m58s, Docker 6m46s, capacity 6m19s,
  recovery 14m11s, both reproducible Linux builds, docs, security, DeepScan,
  and retirement gates. GitHub had zero review thread; Kimi's fallback review
  found 0 P0/P1/P2 because CodeRabbit was rate-limited.
- **Merge/tracker:** authorized administrative squash bypassed only formal
  self-approval. PR #146 merged as `50f18e0`; GH-145 closed and GH-29 now
  records exactly 21/59 with GH-145/PR #146 evidence.
- **Final main:** Go CI
  [31280519207](https://github.com/NeaBouli/TrueRepublic/actions/runs/31280519207)
  passes with quality-depth 4m31s, build 8m17s, Docker 7m10s, capacity 5m15s,
  and recovery 14m00s. Security
  [31280519224](https://github.com/NeaBouli/TrueRepublic/actions/runs/31280519224)
  and Pages
  [31280518754](https://github.com/NeaBouli/TrueRepublic/actions/runs/31280518754)
  pass on the same merge commit.
- **Live readback:** Pages publishes 1,573 cases (1,406 Go + 26 Rust + 141
  maintained client), 21/59 overall, 21/51 phase work, Phase 6 6/7, and
  `production_ready=false`.
- **Boundary/status:** no production code, infrastructure, key, wallet,
  account, deployment, network, or real-funds action occurred. GH-145 is Done;
  remaining rollout/security work stays open.

## 2026-08-08 19:46 EEST GH-139 critical-path coverage → In Progress

- **Branch:** `test/GH-139-critical-coverage` from clean `origin/main`
  `be67d72`.
- **Issue:** [GH-139](https://github.com/NeaBouli/TrueRepublic/issues/139)
- **Scope:** measure reproducible root/application, `x/dex`, and
  `x/truedemocracy` coverage; add deterministic security/integrity regression
  tests for meaningful uncovered paths; and establish a bounded repository
  coverage-regression contract.
- **Delegation:** Kimi K3 receives a bounded secret-free test implementation;
  Sol owns path selection, architecture/security review, all external writes,
  complete verification, and closure.
- **Boundary/risk:** Medium, tests and evidence only. No production behavior,
  consensus, accounting, DEX formula, governance authority, wallet/key,
  deployment, network, or real-funds change. Any production defect is reported
  and separately scoped.
- **Status/blockers:** In Progress; no blocker. Rollout remains 19/59,
  Phase 6 remains 6/7, and `production_ready=false`.

### 2026-08-08 20:28 EEST implementation and local verification → PR ready

- **Coverage contract:** the repository now enforces bounded statement floors
  for root/application (71.8%), DEX (51.0%), and governance (63.7%) in local
  development and Go CI. Fresh measured totals are 71.9%, 51.1%, and 63.8%,
  improving the prior 70.8%, 49.1%, and 63.0% baselines.
- **Critical regressions:** 15 standard Go cases cover atomic output rollback,
  DEX authority/custody/slippage/fail-closed behavior, governance escrow,
  authorization, reward payout, withdrawal, and the repository contract.
- **Delegation/review:** Kimi K3 implemented the bounded 14-case root/DEX/
  governance test block and passed focused plus race checks. Sol reviewed every
  line, reset the DEX event manager to prevent a false-positive assertion,
  integrated the coverage contract/CI/docs, and reran focused and coverage
  gates.
- **Public staged state:** 1,534 standard cases (1,367 Go + 26 Rust + 141
  maintained client); 20/59 overall and 20/51 phase work; Phase 6 remains 6/7;
  `production_ready=false`.
- **Boundary:** no production code, consensus, accounting formula, governance
  authority, wallet/key/signing, deployment, network, or real-funds behavior
  changed. This proves bounded regression depth, not complete security review.
- **Status/blockers:** PR ready after the remaining full repository gates and
  independent final review; no blocker. Protected exact-head CI, merge,
  GH-29/issue synchronization, Pages readback, and final-main verification
  remain before Done.

### 2026-08-08 20:42 EEST full gates and independent review → PASS

- **Sol full gate:** `make verify` passes package selection, build, vet, and
  the complete 1,367-case Go race/coverage suite. Root/application is 71.9%,
  DEX 51.1%, governance 63.8%; all other package coverage remains green.
- **Additional PASS:** coverage contract from repository and `/tmp`, focused
  new tests, workflow YAML parse, documentation consistency, shell syntax,
  formatting, and `git diff --check`.
- **Independent Kimi review:** no P0/P1/P2 finding. Fresh read-only evidence
  independently reproduces all 15 focused cases under race, 1,367 full-suite
  passing cases with zero failures, exact module arithmetic, the three coverage
  totals, vet, docs consistency, and YAML parsing. P3 notes are intentional
  ratchet proximity, duplicate CI execution, and two-source floor pinning.
- **Status/blockers:** locally complete and ready to publish; no blocker.
  Protected exact-head CI, merge, tracker/Bridge/public synchronization, Pages,
  and final-main evidence remain before Done.

### 2026-08-08 20:49 EEST PR #140 published

- **Commit/PR:** implementation commit `a2b8473` is pushed through
  [PR #140](https://github.com/NeaBouli/TrueRepublic/pull/140), closing GH-139
  after merge.
- **Status/blockers:** protected exact-head workflows and review are pending;
  no local blocker. Merge, GH-29/public synchronization, Pages, and final-main
  evidence remain before Done.
## 2026-08-08 20:56 EEST GH-141 client Docker build scripts → In Progress

- **Branch:** `fix/GH-141-client-docker-scripts` from clean `origin/main`
  `be67d72`.
- **Issue:** [GH-141](https://github.com/NeaBouli/TrueRepublic/issues/141)
- **Trigger:** PR #140 protected Docker smoke exposed that
  `client-web/Dockerfile` omits the `scripts/` directory while `npm run build`
  requires `scripts/bundle-budget.mjs`.
- **Scope:** copy the maintained-client build scripts into the Docker builder,
  add a deterministic repository regression contract, and verify the direct
  client image plus relevant repository gates.
- **Boundary/risk:** Low/medium container-build repair only. No application
  production logic, deployment, network, wallet/key/signing, account, or funds
  action.
- **Status/blockers:** In Progress; no blocker. This is separate from GH-139's
  test/evidence scope and must merge independently before GH-140 is refreshed.

### 2026-08-08 21:08 EEST root cause fixed → locally verified

- **Fix:** `client-web/Dockerfile` now copies `scripts/` before
  `npm run build`; a repository contract pins both the copy-before-build order
  and continued bundle-budget invocation.
- **PASS:** focused repository test; clean `npm ci` with zero vulnerabilities;
  10 Node policy cases; 131 Vitest cases with two gated skips; TypeScript/Vite
  production build; enforced bundle budget at 76,401-byte gzip entry, 19 lazy
  routes, 5,028-byte maximum route, and 353,562-byte total JavaScript gzip.
- **Environment limitation:** local Docker CLI is unavailable, so the actual
  image/restart proof remains assigned to protected GitHub Docker CI. The
  failing path's underlying command now passes directly.
- **Staged status:** 1,520 standard cases (1,353 Go + 26 Rust + 141 maintained
  client); rollout stays 19/59 and production stays false.
- **Status/blockers:** locally PR-ready; no code blocker. Protected Docker image
  and restart evidence remain before Done.

### 2026-08-08 21:14 EEST complete local gates → PASS

- **Go/repository:** `make verify` passes package selection, build, vet, and all
  1,353 Go cases under race/coverage; docs consistency and diff checks pass.
- **Client:** clean install/audit, lint, 10 Node policy cases, 131 Vitest cases,
  TypeScript/Vite production build, and bundle budget pass.
- **Status/blockers:** no local blocker. Docker CLI absence is an environment
  limitation, not a code failure; protected image/restart CI is the remaining
  acceptance evidence.

### 2026-08-08 21:20 EEST PR #142 published

- **Commit/PR:** `14d6851` is pushed through
  [PR #142](https://github.com/NeaBouli/TrueRepublic/pull/142). Independent
  read-only review found no functional P0/P1/P2; its untracked-file process
  observation was resolved by the committed regression test.
- **Status/blockers:** protected Docker image/restart and exact-head checks are
  pending; no blocker. After merge, GH-140 must incorporate repaired `main` and
  synchronize its +15-case arithmetic on top of the 1,520-case base.

### 2026-08-08 23:18 EEST GH-141 merged; GH-139 base refresh in progress

- **GH-141 closure:** PR #142 passed every protected exact-head gate, including
  the authoritative Docker image/restart smoke (7m08s), complete Go build/test
  (7m23s), and multi-validator recovery (14m06s), then squash-merged as
  `11ef2f6`; GH-141 is closed.
- **GH-139 integration:** repaired `origin/main` is being merged into PR #140.
  The combined public arithmetic is 1,535 standard cases (1,368 Go + 26 Rust +
  141 maintained client), rollout 20/59 overall and 20/51 phase work, Phase 6
  6/7, with `production_ready=false`.
- **Status/blockers:** no blocker. Conflict resolution, fresh full local gates,
  protected exact-head CI, merge, tracker synchronization, Pages readback, and
  final-main verification remain before GH-139 is Done.

### 2026-08-08 23:50 EEST GH-139 critical-path coverage → Done

- **Merge:** [PR #140](https://github.com/NeaBouli/TrueRepublic/pull/140)
  passed every protected exact-head check with no unresolved review thread and
  squash-merged as `1b6ccef`; GH-139 closed automatically.
- **Exact-head evidence:** Go build/test including the new coverage gate passed
  in 8m32s; Docker restart 7m07s; capacity 5m58s; multi-validator recovery
  14m04s; Linux AMD64/ARM64, docs, Go/Rust/Node security, retirement,
  DeepScan, and review gates all passed.
- **Final-main evidence:** on exact merge commit `1b6ccef`, Go CI
  `31277242596` passed build/test (9m12s), capacity (6m11s), Docker restart
  (7m12s), and multi-validator recovery (13m04s); Security Scan `31277242587`
  and Pages `31277242292` passed.
- **Public/tracker readback:** GH-29 marks its Phase 5 critical-path coverage
  item complete. Live Pages reports 1,535 cases (1,368 Go + 26 Rust + 141
  maintained client), rollout 20/59 overall and 20/51 phase work, Phase 6 6/7,
  and `production_ready=false`.
- **Boundary/blockers:** Done with no blocker. This block strengthens tests,
  CI, and evidence; it is not production, mainnet, or real-funds approval.

## 2026-08-08 18:01 EEST GH-131 submitted transaction history → In Progress

- **Branch:** `feature/GH-131-transaction-history` after the shared Bridge
  baseline merges.
- **Issue:** [GH-131](https://github.com/NeaBouli/TrueRepublic/issues/131)
- **Scope:** bounded newest-first server-paginated history for transactions
  submitted by the validated unlocked wallet; explicit unavailable, timeout,
  protocol, decode, authoritative-empty, retry, and stale-wallet handling.
- **Boundary:** this is honestly labelled submitted history. Incoming-only
  account activity, a participant index, wallet/key/signing-policy changes,
  deployment, production, and funds are out of scope.
- **Risk/delegation:** Medium untrusted REST/UI boundary. Kimi K3 receives the
  bounded implementation block; Sol owns query architecture, diff review,
  disposable-chain evidence, full integration, GitHub writes, and closure.
- **Status/blockers:** In Progress; no blocker. GH-29 remains 17/59 and
  `production_ready=false`.

### 2026-08-08 18:49 EEST implementation and real-chain proof → Sol review

- **Changed:** Cosmos SDK v0.50 `query/page/limit/order_by` submitted-history
  service with typed failures and strict decoding; dedicated stale-safe wallet
  history state; accessible newest-first paginated UI; service/store/component
  tests; and a second case in the existing disposable-chain integration.
- **Real evidence:** the first attempt exposed that SDK v0.50 ignores deprecated
  `events/pagination` fields. The corrected live REST contract returns top-level
  `total`; both disposable-chain cases pass in 75.58s. Page 1 begins with a
  real committed slippage failure preserving code, bounded log, timestamp,
  memo, and exact fee; page 2 is disjoint; a receiver-only address is an
  authoritative empty submitted-history result.
- **Delegation/review:** Kimi K3 implemented the bounded block and ran its full
  client gates. Sol reviewed every changed file and corrected service caching,
  caller-abort handling during decode, malformed response-count rejection,
  wallet-create history clearing, failed-page retry state, stale-row rendering,
  and the committed-send/history-refresh failure boundary. Independent Spark
  review found no P0-P2 and identified the latter two P3 UX risks before fix.
- **Current PASS:** lint and 37/37 focused service/store/component cases after
  Sol remediation. Kimi's prior full gate passed ten Node policy cases, 129
  Vitest cases with two gated skips, production build/budgets, guarded audit,
  diff check, and both real-chain cases. Full post-remediation rerun remains.
- **Boundary:** incoming-only transfers remain explicitly excluded; configured
  REST-provider trust remains unchanged. No real account/key/fund, public
  broadcast, listener, deployment, or production action occurred.
- **Status/blockers:** Review; no blocker. Full Sol rerun, public status/rollout
  synchronization after the preceding GH-132 merge, protected PR CI, merge,
  and final-main evidence remain before Done.

### 2026-08-08 18:51 EEST full post-remediation verification → PASS

- **Client gates:** ten Node policy/budget cases and 131 Vitest cases pass;
  two real-chain cases remain gated from standard arithmetic and also pass in
  76.88s. Lint, TypeScript build, production bundle budgets, guarded High
  audit, focused 37/37, and diff check pass.
- **Bundle:** 76,219-byte gzip initial entry, 19 lazy routes, 5,028-byte gzip
  maximum direct route, and 353,061-byte gzip total JavaScript on the GH-131
  branch before incorporating GH-132. Final combined metrics will be recorded
  after merging exact `main`.
- **Status/blockers:** Implementation, independent review, Sol remediation, and
  local integration are complete; no blocker. Incorporate GH-132 final main,
  synchronize 19/59 public status, rerun combined gates, then publish PR.

### 2026-08-08 18:55 EEST combined-main and public status → PR ready

- **Predecessors:** GH-133 is merged as `53326d9` with native Linux evidence
  and green final-main Security/Pages. GH-132 is merged as `795ebd2`; its final
  combined head passed the full protected browser/client/security matrix,
  GH-132 is closed, GH-29 records the completed item, and live Pages reads back
  18/59 and 18/51.
- **Combined GH-131 PASS:** fresh install (zero vulnerabilities), lint, ten Node
  policy/budget cases, 131 standard Vitest cases with two gated skips,
  TypeScript/production build, and guarded audit. The two gated disposable-
  chain cases pass separately in 76.88s. Final combined bundle is 76,291-byte
  gzip initial, 19 lazy routes, 5,027-byte maximum route, and 353,442-byte total
  JavaScript gzip.
- **Staged public state:** 1,519 standard cases (1,352 Go + 26 Rust + 141
  maintained client); 19/59 overall and 19/51 phase work; Phase 6 stays 6/7;
  `production_ready=false`. Roadmap, landing page, README, wiki, project state,
  security notes, action log, and source-of-truth status are synchronized.
- **Status/blockers:** PR ready; no blocker. Documentation/retirement/diff
  gates, commit/push, protected exact-head CI, merge, GH-29 closure sync,
  public readback, and final-main evidence remain before Done.

### 2026-08-08 19:08 EEST PR #137 / Kimi final review → remediation

- **Protected first head:** `c166df7` passes browser quality (2m04s), client
  build (34s), both real client-chain cases (3m55s), docs, Go/Rust/Node
  security, all retirement contracts, DeepScan, and review statuses.
- **Independent Kimi review:** no P0/P1/P2. Kimi independently reproduced 10/10
  Node, 131 Vitest plus two skips, lint, exact bundle budgets, docs consistency,
  and both real-chain cases (74.40s), and reverified SDK v0.50.14 plus CometBFT
  v0.38.22 semantics from pinned source.
- **P3 remediation:** source inventory now excludes `*.test.tsx` and records 43
  actual components instead of counting the component test as runtime surface.
  The store now passes a caller abort signal and cancels superseded, lock, and
  wallet-change history requests in addition to generation invalidation.
- **Accepted P3:** the second disposable-chain case intentionally shares the
  first case's established delivery corpus in the standard sequential file run;
  a filtered single-test invocation is not a supported acceptance command.
- **Status/blockers:** remediation verification and one final protected rerun
  remain; no blocker. Production remains false.

### 2026-08-08 19:20 EEST external review remediation → locally verified

- **Review triage:** all four actionable CodeRabbit findings were valid and
  corrected. The default browser fetch preserves its global receiver; rejected
  transaction-module/service imports are no longer cached and the history
  action records a retryable typed error without a second dynamic import;
  invalid timestamps use the bounded unavailable fallback; and the wiki client
  arithmetic now matches the canonical 141-case total.
- **Independent chain fixture:** the pagination case now commits its own two
  successful sends followed by the expected failed swap. It no longer depends
  on a preceding test and passes alone in 27.13s with one test pass and the
  unrelated delivery case skipped.
- **Verification PASS:** lint, 37/37 focused service/store/component cases,
  TypeScript production build, bundle budgets (76,401-byte gzip entry, 19 lazy
  routes, 5,028-byte maximum route, 353,562-byte total JavaScript), and diff
  check. The default-fetch receiver path has explicit regression coverage
  without changing the standard case count.
- **Status/blockers:** local review remediation is complete; no blocker. The
  full sequential real-chain file, protected exact-head rerun, review-thread
  resolution, merge, tracker sync, Pages readback, and final-main evidence
  remain before Done. Production remains false.

### 2026-08-08 19:22 EEST final local review gate → PASS

- **Full real-chain file:** both sequential disposable-chain cases pass in
  84.53s: canonical maintained transaction delivery in 57.36s and the now
  self-contained submitted-history pagination case in 15.11s.
- **Final local state:** all external findings are fixed; lint, focused 37/37,
  complete standard client tests, production build/budgets, documentation,
  retirement contracts, diff check, standalone history-chain evidence, and
  full sequential chain evidence pass. Exact-head protected CI is next; no
  blocker and production remains false.

### 2026-08-08 19:32 EEST merge and final-main verification → Done

- **Merge/tracker:** PR #137 exact head `cc1e3dd` passed every protected check;
  all four review threads were answered and resolved. The authorized squash
  merge produced `7c751b2`; GH-131 closed automatically and GH-29 now checks
  the transaction-history item at exactly 19/59 overall and 19/51 phase work.
- **Final `main` PASS:** Client Web CI `31266901488` (build 31s, protected
  Chromium/Firefox/WebKit browser quality 1m53s, two-case disposable-chain
  integration 4m06s), Security Scan `31266901498`, and Pages
  `31266901092` all pass on `7c751b2`.
- **Public readback:** live GitHub Pages publishes 1,519 standard cases, 19 of
  59 rollout items (32%), 19 of 51 phase-work items (37%), Phase 6 at 6/7,
  and `production_ready=false`.
- **Closure:** GH-131 is Done with no blocker. Submitted-only history remains
  intentionally distinct from incoming participant history, and configured
  REST-provider trust remains documented. No production, deployment, real
  key/account/fund, or infrastructure action occurred.

## 2026-08-08 18:01 EEST GH-132 client browser qualification → In Progress

- **Branch:** `test/GH-132-client-browser-quality` after the shared Bridge
  baseline merges.
- **Issue:** [GH-132](https://github.com/NeaBouli/TrueRepublic/issues/132)
- **Scope:** real-browser accessibility, keyboard/focus, viewport-overflow,
  delayed-lazy-loading, supported-browser, and low-bandwidth evidence for the
  safe unauthenticated `/unlock`, `/create`, and `/import` surfaces.
- **Boundary:** tests must not execute wallet creation/import, authenticated
  wallet operations, signing, crypto, network mutation, or production actions.
- **Risk/delegation:** Low/Medium client quality and dependency surface. A
  bounded worker owns tests and minimal shared-control fixes; Sol owns tooling
  approval, audit/budget review, complete gates, GitHub writes, and closure.
- **Status/blockers:** In Progress; no blocker. This block does not claim the
  deeper authenticated Phase-4 exit.

### 2026-08-08 18:38 EEST local implementation → Sol review

- **Changed:** pinned Playwright/axe tooling; five desktop/mobile browser
  profiles; protected browser-quality CI; serious/critical accessibility,
  keyboard/focus, viewport-overflow, delayed-lazy-route, and third-party
  request tests; minimal shared label/error/loading/live-region/focus fixes;
  and an explicit browser-support boundary.
- **Local PASS:** clean install with zero vulnerabilities; lint; 94 Vitest and
  ten Node policy/budget cases (one pre-existing skip); production build and
  budgets (75,861-byte gzip entry, 19 lazy routes, 349,730-byte total JS);
  guarded high-severity audit; Chromium desktop 9/9, Firefox desktop 9/9,
  Chromium mobile 8/8 plus one intentional physical-keyboard skip; workflow
  parsing and diff check.
- **Review:** bounded worker implemented the block. Sol reviewed the complete
  diff. Independent Spark review found no P0-P2; its documentation-clarity P3
  was corrected by explicitly distinguishing desktop physical-keyboard checks
  from mobile profiles. The stable Vite component chunk-name intercept remains
  a non-blocking maintenance sensitivity.
- **Risk/boundary:** The secure pinned Playwright 1.55.1 WebKit cannot run on
  the frozen local macOS-12 runner. Downgrading to a locally compatible release
  was rejected because it has two High advisories. Protected Ubuntu CI is the
  mandatory full Chromium/Firefox/WebKit acceptance gate.
- **Status/blockers:** Review; no product blocker. Sol rerun, commit, protected
  full browser CI, merge, and final-main evidence remain before Done.

### 2026-08-08 18:42 EEST Sol rerun → PASS / PR ready

- **Sol PASS:** fresh `npm ci` (389 packages audited, zero vulnerabilities),
  lint, ten Node policy/budget cases, 94 Vitest cases plus one pre-existing
  skip, production build/budgets, guarded high-severity audit, Chromium
  desktop 9/9, Firefox desktop 9/9, and Chromium mobile 8/8 plus the documented
  physical-keyboard skip. Documentation consistency, all retirement contracts,
  workflow YAML parsing, and diff check also pass.
- **Status/blockers:** Local implementation and review are complete; no
  product blocker. Protected Ubuntu full-matrix CI, merge, and final-main
  evidence remain before Done.

### 2026-08-08 18:44 EEST PR #136 exact-head CI → PASS

- **Protected evidence:** exact head `2ce1b59` passes the complete pinned
  Chromium/Firefox/WebKit desktop/mobile browser-quality matrix in 2m02s,
  client build in 25s, disposable client-chain integration in 3m41s,
  Go/Rust/Node security, all retirement contracts, docs consistency,
  DeepScan, and review statuses.
- **Boundary confirmed:** Browser tests exercised only the documented safe
  unauthenticated surfaces and made no wallet, signing, broadcast, deployment,
  or production action. The protected Linux WebKit gate removes the local host
  limitation without weakening the dependency-security floor.
- **Status/blockers:** PR #136 is technically green; repository review policy
  requires the already authorized administrative merge. Merge, final-main
  workflows, and final Bridge closure remain before Done.

### 2026-08-08 18:47 EEST public rollout synchronization staged

- **Public state:** the completed Phase-4 browser/accessibility checklist item
  is linked to GH-132/PR #136; staged arithmetic advances from 17/59 to 18/59
  overall and from 17/51 to 18/51 phase work. Phase 6 stays 6/7 and
  `production_ready=false`.
- **Build evidence:** exact current client output is 75.86 kB gzip initial,
  5.03 kB maximum lazy route, and 349.73 kB total JavaScript gzip. Status,
  roadmap, landing page, project state, security notes, and action log agree;
  documentation consistency passes.
- **Boundary:** this closes only the safe unauthenticated browser-quality item.
  Authenticated wallet/key/signing checks, real ZKP, IBC UX, deployment, real
  funds, and production approval remain open.

## 2026-08-08 18:01 EEST GH-133 deterministic Linux daemon build → In Progress

- **Branch:** `build/GH-133-deterministic-linux-daemon` after the shared Bridge
  baseline merges.
- **Issue:** [GH-133](https://github.com/NeaBouli/TrueRepublic/issues/133)
- **Scope:** tracked Linux amd64/arm64 and Go 1.26.5 build contract,
  deterministic double-build verification, version check, SHA-256 manifest,
  focused tests, and CI-only artifact upload.
- **Boundary:** no tag, GitHub Release, registry push, signing key, SBOM,
  provenance, deployment, production, or rollout-approval claim/action.
- **Risk/delegation:** Low/Medium release-engineering evidence. A bounded worker
  owns the isolated build/scripts/workflow block; Sol owns claim boundaries,
  diff review, full gates, GitHub writes, and closure.
- **Status/blockers:** In Progress; no blocker. Container reproducibility and
  signed/SBOM/provenance artifacts remain separate GH-29 work.

### 2026-08-08 18:25 EEST local implementation → Sol review

- **Changed:** tracked `truerepublic.daemon-build/v1` contract; deterministic
  build, verifier, and negative-test scripts; Make targets; native Linux
  amd64/arm64 PR workflow; narrowly bounded operator documentation.
- **Evidence:** two clean-cache builds per native architecture use Go 1.26.5,
  CGO, readonly modules, `-trimpath`, disabled VCS/build IDs, and exact
  `main.version=<40-char commit>`. Verification requires identical SHA-256,
  native ELF architecture, embedded Go version, and exact CLI version before
  emitting checksums/metadata. CI uploads no binary.
- **Local PASS:** contract success/hash/version/contract/ref/target tests;
  shell syntax; JSON and workflow YAML parsing; focused root-command Go test;
  diff check. A platform-equivalent macOS double build was byte-identical;
  exact Linux evidence intentionally belongs to the protected PR matrix.
- **Delegation/review:** bounded worker implemented the isolated block. Spark
  found no P0-P2. Its P3 notes that unit negatives mock binary inspection;
  accepted because the immediately following matrix step executes the exact
  full binary double-build/inspection path on both Linux architectures.
- **Risk/boundary:** Low/Medium CI evidence. ARM uses GitHub's public-preview
  native runner. This proves same-run repeatability, not cross-environment
  provenance. No tag, release, binary upload, registry, signature, SBOM,
  provenance, deployment, production artifact, or rollout approval.
- **Status/blockers:** Review; no blocker. Full local Go verification, commit,
  protected CI, merge, and final-main evidence remain.

### 2026-08-08 18:34 EEST full local verification → PASS / PR ready

- **Full gate:** `make verify` passes, including package-selection policy,
  CGO build, vet, and race/coverage tests (root 70.8%; maintained policy
  packages 80.3–97.2%; DEX 49.1%; truedemocracy 63.0%).
- **Additional gates:** documentation consistency, all three retirement
  contracts, shell syntax, JSON/workflow parsing, the deterministic-build
  contract suite, focused root-command test, and `git diff --check` pass.
- **Status/blockers:** Local review is complete and no blocker is known.
  Exact native Linux amd64/arm64 repeatability remains the protected PR-CI
  acceptance gate; merge and final-main evidence remain before Done.

### 2026-08-08 18:37 EEST PR #135 exact-head CI → PASS

- **Protected evidence:** exact head `f3f60b6` passes native Linux amd64
  (4m53s) and arm64 (3m29s) double-build/hash/version verification in run
  `31264671102`. Docs, Go/Rust/Node security, all retirement contracts,
  DeepScan, and review statuses also pass.
- **Artifact boundary:** both matrix jobs upload only repository evidence
  checksums and metadata; no daemon binary, release, tag, signature, registry,
  deployment, or production artifact/action is produced.
- **Status/blockers:** PR #135 is technically green; repository review policy
  still requires the already authorized administrative merge. Merge,
  final-main workflows, public tracker synchronization, and final Bridge
  closure remain before Done.

## 2026-08-08 14:15 EEST GH-128 client route splitting → In Progress

- **Branch:** `perf/GH-128-client-route-splitting`
- **Issue:** [GH-128](https://github.com/NeaBouli/TrueRepublic/issues/128)
- **Baseline:** exact clean `origin/main` `e7a6aa6`; GH-121 and its handoff are
  merged, final-main and Pages are green.
- **Scope:** preserve all 21 maintained-client routes while loading non-entry
  screens on demand; add a deterministic emitted-build budget and regression
  coverage; publish measured before/after evidence.
- **Risk:** Medium client architecture/performance. Wallet, signing, broadcast,
  query semantics, chain protocol, and production configuration are out of
  scope and must remain unchanged.
- **Delegation:** Kimi K3 receives one bounded, secret-free, read-only route and
  build-output analysis. Sol owns architecture, implementation integration,
  full tests, GitHub writes, and closure.
- **Status/blockers:** in progress; no blocker. Rollout remains 16/59, Phase 6
  remains 6/7, and production readiness remains false.

### 2026-08-08 14:27 EEST local verification

- **Implementation:** all 19 page routes now use `React.lazy`; one accessible
  `Suspense` boundary sits inside the existing application `ErrorBoundary`.
  Wallet signing, chain-query, and transaction services load on demand, which
  removes the eager `MobileNav → walletStore → CosmJS` entry dependency without
  changing wallet, signing, registry, simulation, broadcast, or query behavior.
- **Budget:** Vite emits a hash-independent manifest and `npm run build` now
  enforces raw plus pinned Node-zlib gzip limits for the entry, every direct
  route, every chunk, and total JavaScript. Measured initial entry: 234,322 raw
  / 75,786 gzip bytes, down from 1,719,256 raw / 322.63 kB Vite gzip. Nineteen
  lazy routes are present; max direct route is 5,031 gzip bytes; total deferred
  JavaScript is 1,733,603 raw / 349,422 gzip bytes.
- **Tests:** PASS `npm run lint`; PASS 104 maintained-client cases (94 Vitest +
  10 Node policy/budget); PASS production build and budget; PASS live High
  audit. Route order/params/redirects stay pinned; loading and rejected-chunk
  behavior have regressions.
- **Delegation/review:** Kimi K3 independently identified the eager CosmJS trap,
  recommended service-level dynamic imports and hash-independent emitted-file
  budgets, and found no P0. Spark independently confirmed the minimal lazy-route
  fallback/error test boundary. Sol implemented, reviewed, and reran all gates.
- **Files:** `client-web/src/{App.tsx,routes.tsx,routes.test.tsx}`, wallet
  service/store, Vite/package/build-budget scripts, React/Docs CI trigger, public
  status/roadmap/wiki, and agent-bridge evidence.
- **Status/blockers:** local implementation is Review-ready with no blocker.
  Protected PR checks, exact-head review, merge, GH-29 synchronization, Pages,
  and final-main verification remain. Public rollout staging is 17/59 overall,
  Phase 6 remains 6/7, and `production_ready=false`.

### 2026-08-08 14:35 EEST integration follow-up

- **Review remediation:** Kimi's final pass identified a possible duplicate
  first-service construction under concurrent store calls. Both dynamic service
  loaders now cache one shared Promise, preserving the original singleton
  behavior without restoring the eager dependency.
- **Fresh result:** PASS lint; PASS 104 client cases; PASS production build with
  level-9 budget measurements (entry 234,322 raw / 75,786 gzip; max route 5,031
  gzip; max chunk 1,057,764 raw / 134,769 gzip; total 1,733,603 raw / 349,422
  gzip); PASS live High audit; PASS real disposable-chain client delivery (one
  case, 69.84 seconds) on the final source state.
- **Transparent failed invocation:** the preceding combined command successfully
  built `truerepublicd` but then ran `npm` from the repository root, which has no
  `package.json`, and exited ENOENT. The corrected command ran from `client-web`
  and passed; this was an invocation-path error, not a product failure.

### 2026-08-08 14:39 EEST publication

- **Commit/PR:** exact implementation commit `b026739` pushed on
  `perf/GH-128-client-route-splitting`; protected
  [PR #129](https://github.com/NeaBouli/TrueRepublic/pull/129) opened against the
  unchanged exact `origin/main` `e7a6aa6` and links `Closes #128`.
- **Review:** final Kimi review reports no P0/P2 code finding after the cached-
  Promise remediation. Its only P1 handoff was to include the three new budget/
  audit files; commit `b026739` includes all three.
- **Status:** protected exact-head Client, real client-chain, Docs, security,
  and review checks are pending. No merge or rollout closure is claimed yet.

### 2026-08-08 14:50 EEST closure

- **Merge/issue:** PR #129 exact head `dcc1113` passed every required check and
  was squash-merged as `47f8c08`; GH-128 closed automatically. GH-29 now marks
  the route-splitting and bundle-budget item complete.
- **Exact-head evidence:** Client build and disposable-chain integration, Docs,
  Go vulnerability, Rust audit, Node audit, all three retirement contracts,
  DeepScan, and formal review statuses passed. No review thread remains.
- **Final-main evidence:** Client Web CI
  [`31255695662`](https://github.com/NeaBouli/TrueRepublic/actions/runs/31255695662),
  Security Scan
  [`31255695671`](https://github.com/NeaBouli/TrueRepublic/actions/runs/31255695671),
  and Pages
  [`31255695086`](https://github.com/NeaBouli/TrueRepublic/actions/runs/31255695086)
  passed on exact merge commit `47f8c08`.
- **Live readback:** GitHub Pages exposes 1,482 recovery-verified cases, 17 of
  59 rollout items complete, and the 19-route deterministic bundle-budget
  evidence. Phase 6 remains 6/7 and `production_ready=false`.
- **Status/blockers:** GH-128 is Done with no blocker. No deployment,
  production, key, account, signing, broadcast, or fund action occurred.

---

## 2026-08-09 11:02 EEST GH-148 reproducible security gates → In Progress

- **Issue/parent:** [GH-148](https://github.com/NeaBouli/TrueRepublic/issues/148)
  closes the single open GH-29 Phase-5 checkbox for maintained dependency,
  static-analysis, secret-scanning, and supply-chain gates.
- **Branch/base:** `security/GH-148-reproducible-gates` from exact clean
  `origin/main` `c1a0ed42c543c9a664f4e4fa3653fac406957c90`; no open PR and no
  inherited working-tree change.
- **Scope:** immutable Action/tool pins, blocking static and secret scanning,
  repository-owned fail-closed policy tests, bounded dependency-update
  automation, operator/contributor guidance, and complete local/protected-CI
  evidence.
- **Boundary:** runtime, protocol, consensus, wallet/signing, ZKP, release
  signing, SBOM/provenance publishing, deployment, infrastructure, mainnet,
  production keys/accounts/funds, and external-audit claims are excluded.
- **Delegation:** Kimi K3 is performing a bounded secret-free read-only
  baseline audit. Sol owns architecture, every write/diff, security decisions,
  full verification, GitHub writes, merge, tracker/status, and closure. Kimi
  may not delegate.
- **Next:** reconcile Kimi findings with the source inventory, freeze exact
  tool/action versions and the repository contract, then implement the
  smallest coherent gate set with negative fixtures.

### 2026-08-09 11:45 EEST implementation and focused verification

- **Kimi audit reconciled:** its read-only report found the fail-open
  govulncheck pipeline, mutable Action/scanner/toolchain inputs, absent secret
  gate and update automation, optional staticcheck, incomplete Cargo locking,
  trigger gaps, and missing job timeouts. Sol independently reproduced and
  closed every in-scope P0-P2 repository finding; Kimi wrote no file.
- **Implementation:** all 35 Action uses are approved SHA pins; one versioned
  contract owns exact Go/Node/Rust and scanner versions; every job is bounded;
  dependency/static/secret/lock and planted-failure gates block; weekly grouped
  Dependabot updates cover Actions, Go, Cargo, and maintained npm.
- **Focused PASS:** six repository contract cases, staticcheck after 14 minimal
  baseline corrections, zero maintained-tree secret findings plus planted
  credential rejection, govulncheck with four reachable no-fix entries,
  cargo-audit with zero vulnerabilities/five warning-only advisories, npm with
  zero high/critical advisories, client lint/141 cases/build/audit, ten-second
  quality campaigns, docs/retirement/YAML/shell/JSON/diff checks.
- **Candidate status:** 1,579 cases (1,412 Go + 26 Rust + 141 client), rollout
  22/59, phase work 22/51, Phase 6 6/7, and production false. Fresh client
  build is 76.40 kB gzip initial, 5.03 kB max route, 353.56 kB total.
- **Running/pending:** complete Go/Rust gates are still running; then final Kimi
  diff review, commit/push/PR, exact-head CI/review, merge, GH-29/Pages/final-main
  closure, followed by the user-requested retroactive PR/branch/worktree audit.
- **Boundary:** no deployment, release, production, mainnet, real key/account/
  fund/signing, private infrastructure, or external-audit claim.

### 2026-08-09 12:26 EEST final-review remediation

- **Independent finding:** Kimi reproduced one P1: bare govulncheck exits 3 on
  the four reachable no-fix findings, which would leave protected CI red. It
  also found local gitleaks false positives in ignored Rust build artifacts.
- **Repair:** the Go gate now parses JSON and permits only the exact four active
  no-fix IDs in the central contract. Exceptions expire within 30 days; new,
  fixable, missing, stale, expired, malformed, or scanner-error cases fail.
  Secret scanning now selects tracked plus non-ignored changed files only.
- **Reverification:** live Go policy gate PASS; allowed/unknown/fixable/expired/
  scanner-error fixtures PASS; maintained-tree gitleaks PASS at 4.02 MB with
  existing build outputs present; planted-secret failure PASS; repository
  contract, staticcheck, docs, shell, JSON/YAML, and diff checks PASS.
- **Next:** obtain independent re-review of the repaired diff, refresh full
  relevant verification if required, then commit/push/PR/exact-head CI/merge and
  final-main/Pages/GH-29 closure before retroactive cleanup.

### 2026-08-09 12:37 EEST remediation re-review findings closed

- **Re-review:** prior P1/P3 closed, but Kimi found three P2 edge cases: the
  standalone Go wrapper did not itself validate real calendar dates or the
  configured 30-day maximum, and the secret-tree process substitution could
  hide a Git enumeration failure.
- **Repair:** the Go wrapper now parses UTC dates, validates ID/reason/type,
  rejects duplicates, and enforces the central integer 1..30-day maximum. Its
  fixtures now cover over-long, malformed, stale, duplicate, expired, unknown,
  fixable, scanner-error, and allowed cases. Secret enumeration is captured by
  a checked command, requires at least one file, and has a dedicated failing-Git
  fixture.
- **PASS:** expanded Go/secret fixtures, repository contract, shell syntax,
  maintained-tree gitleaks plus planted credential, diff check, and repeated
  full `make verify` (build, vet, Race/Coverage across all 13 Go packages).
- **Pending:** final independent approval of this second remediation, then
  publish and complete protected CI/merge/final-main closure.

### 2026-08-09 12:45 EEST final review approved and P3s closed

- **Review:** Kimi returned APPROVE with no P0/P1/P2. It independently confirmed
  the live JSON exit semantics, date/span/type/duplicate policy, Git enumeration
  failure and empty-tree rejection, contract/workflow wiring, arithmetic, and
  documentation consistency.
- **P3 closure:** the wrapper now also requires canonical zero-padded ISO dates;
  an invalid-ID fixture is committed; PROJECT_STATE timestamp and the second
  hardening summary are current.
- **Next:** final local sweep, commit/push/PR, exact-head CI/review, merge, and
  final-main/Pages/GH-29 synchronization. Retroactive repository reconciliation
  remains the immediately following block.

### 2026-08-09 12:52 EEST PR #149 first exact-head CI correction

- **Published:** commit `970a198`, PR #149. DeepScan and docs passed on the first
  head; Rust CI failed immediately because the exact Rust 1.95.0 toolchain action
  did not install `rustfmt` by default (`cargo-fmt is not installed`).
- **Correction:** explicitly install the `rustfmt, clippy` components through the
  same immutable toolchain action. No Rust source, lockfile, runtime, protocol,
  or public status changed.
- **Verification:** local Rust 1.95.0 fmt/clippy/build/test/audit had already
  passed; workflow YAML, repository contract, and exact new-head CI are required
  again before merge.

---

## 2026-08-08 11:22 EEST GH-121 browser query transport → Review

- **Branch:** `fix/GH-121-browser-query-transport`
- **Issue:** [GH-121](https://github.com/NeaBouli/TrueRepublic/issues/121)
- **Baseline:** exact clean `origin/main` `2ec713c`; no open pull request;
  latest final-main Docs, Security, and Pages workflows pass.
- **Scope:** inventory every maintained-browser custom-module read, reproduce
  the unsupported REST-alias and fail-soft behavior, select one registered
  typed transport, implement it without widening public listeners, and prove
  governance/validator/DEX/membership-admin/ZKP reads plus unavailable and
  empty-state behavior against a disposable chain.
- **Risk:** High client/chain compatibility boundary. A transport failure must
  never become authoritative empty chain state, and query-only work must not
  weaken signing, broadcast, listener, identity, or production safeguards.
- **Delegation:** Kimi K3 receives one bounded secret-free repository analysis
  and implementation-design block. Sol owns architecture, security, every
  external write, diff review, full verification, publication, and closure.
- **Safety:** no deployment, infrastructure, production, key, signing,
  broadcast, account, fund, or proof-system action.

### 2026-08-08 11:41 EEST progress

- **Inventory:** 18 unique unsupported custom-module HTTP aliases across 23
  maintained-browser call sites. Five governance subresources are derived from
  the canonical Domain response; every DEX read maps directly to a registered
  query. Merkle proof and Pay-to-Put required two new registered read methods.
- **Implemented:** centralized protobuf/JSON `ModuleQueryClient` over the
  existing CometBFT JSON-RPC endpoint; all maintained custom-module reads now
  use registered `truedemocracy.Query/*` or `dex.Query/*` paths. Unavailable,
  non-zero ABCI, malformed protobuf, and malformed JSON remain explicit errors.
- **Kimi contribution:** bounded Go ownership added `MerkleProof` and
  `PayToPut`, CLI commands, service handlers/clients, protocol route/execution
  checks, and focused fail-closed tests. Sol reviewed the complete diff and
  independently reran focused tests.
- **Evidence so far:** client production build PASS; ESLint PASS; six focused
  transport/ZKP cases PASS; all 13 new Go query cases PASS; route registration
  and live in-process gRPC-over-ABCI tests PASS. Full repository gates and the
  real disposable-process client-chain test remain pending.
- **Helper status:** bounded Claude Code documentation review exhausted its
  budget without a usable result; Sol owns the documentation reconciliation.
- **Blockers:** none. Rollout remains 16/59 and production readiness false.

### 2026-08-08 12:04 EEST local verification

- **Result:** all maintained custom-module reads use registered protobuf
  gRPC-over-ABCI paths; nested governance, DEX, Merkle, nullifier, and economic
  payloads fail closed. The compatible `nanoid` 3.3.18 lock resolution clears
  the newly published live High advisory.
- **Evidence:** Go build/vet/race/coverage and 1,352 cases PASS; Rust
  fmt/clippy/build/26 tests/audit PASS; maintained client clean install,
  lint/98 tests/build at 322.63 kB gzip/live High audit PASS; the real
  disposable-chain browser query test PASS.
- **Audit:** 0 FAIL, 2 documented WARN (configured-RPC trust and existing
  bundle/large-domain scalability), 6 PASS. No blocking implementation finding.
- **Independent review:** Kimi found no P0/P1/P2. Sol remediated all four
  substantive P3 observations: strict fixed-width protobuf bounds, per-field
  validator validation, and visible rejection handling in both direct LP-query
  consumers. Cosmetic legacy nullable signatures are non-behavioral.
- **Protected review:** PR #126 exact head passed every workflow. CodeRabbit's
  eight inline findings are being remediated before merge: stale pricing race,
  unbound member-state claims, bounded RPC timeout, timestamp/depth/hex
  normalization, one documentation boundary, and one stale count.
- **Status:** implementation is locally complete; independent final-head
  review, protected GitHub checks, merge, issue closure, and final-main
  verification remain. Rollout stays 16/59 and production readiness false.

### 2026-08-08 12:56 EEST protected-review remediation

- **Status:** all eight CodeRabbit findings and both final Kimi P3 findings are
  remediated locally; no P0/P1/P2 finding remains.
- **Changed:** Pay-to-Put quotes and LP errors are request-keyed; anonymous
  identity state is no longer attributed to member addresses; query timeout,
  timestamp, protobuf, Merkle depth/canonical hex, fail-closed voting, response
  errors, documentation, and published arithmetic are hardened.
- **Kimi contribution:** independent read-only review found only response-body
  timeout classification and one stale privacy comment at P3. Sol corrected
  both and added the body-read timeout regression case.
- **Tests:** `make verify` PASS (1,352 Go cases plus Rust fmt/clippy/build/26
  tests/audit); maintained-client 98 tests, lint, 322.63 kB gzip build, and live
  High audit PASS; real disposable-chain integration PASS; consistency,
  retirement, shell, JSON, workflow-YAML, and diff checks PASS.
- **Risk/blockers:** no implementation blocker; configured-RPC trust and bundle/
  whole-domain scalability remain documented rollout risks. Exact-head CI,
  thread resolution, merge, issue closure, and final-main proof remain.

### Sol review feedback

Approved for protected review. Kimi's initial implementation review found no
P0/P1/P2; all four substantive P3 observations were remediated and its focused
re-review found no new P0/P1/P2/P3. Sol reviewed the full diff and reproduced
all complete local gates. Merge and Done remain conditional on protected
exact-head and final-main verification.

### 2026-08-08 13:27 EEST merge handoff → Done

- **PR/merge:** PR #126 exact head `d7a41b0` passed every protected check and
  was administratively squash-merged as `239e6c3`; GH-121 closed automatically.
- **Review:** all eight CodeRabbit threads were answered and resolved; DeepScan
  reports zero new and eight fixed issues; independent Kimi review leaves no
  open P0/P1/P2 finding.
- **Final main:** merge commit `239e6c3` passed Client `31252288314`, Security
  `31252288325`, Go `31252288303`, and Pages `31252287866`.
- **Status:** implementation, protected publication, issue closure, final-main
  workflows, and Pages are complete. GH-29 receives the unchanged rollout
  evidence; no rollout exit was claimed.
- **Rollout:** unchanged at 16/59 overall and Phase 6 at 6/7; production remains
  false and no production-affecting action occurred.

---

## 2026-08-07 02:30 EEST GH-116 retired custom ABCI query shim → Done

- **Issue/PR:** GH-116, PR #124; administratively squash-merged to `main` as
  `5dc4487` after the only remaining block was the formal one-review rule.
- **Final head:** `df465e8`; mergeable; 0 unresolved review threads; DeepScan
  PASS. CodeRabbit did not run a substantive review because its temporary
  project limit was active; two independent Kimi read-only reviews found no
  P0/P1/P2 and Sol completed the final diff review.
- **Protected exact-head evidence:** PR-native Go CI `31130371098`, Docs
  `31130368575`, and Security Scan `31130368777` PASS. Exact-head manual Client
  `31129653076`, Rust `31129654208`, Docs `31129651938`, Security
  `31129655164`, and Go `31129650918` also PASS.
- **Final-main evidence:** on merge commit `5dc4487`, Go CI `31131404107`
  attempt 2, Docs `31131405747`, Client `31131407330`, Rust `31131408773`,
  Security `31131410151`, and Pages `31131318674` PASS. Go attempt 1 had one
  transient legacy-migration P2P startup miss while its other jobs and seven
  later recovery harnesses passed; the unchanged commit then passed all eight
  recovery harnesses on the controlled rerun.
- **Result:** legacy `custom/truedemocracy/...` and `custom/dex/...` routing is
  removed and fail closed; all 17 protobuf gRPC Query methods remain supported
  and tested; custom grpc-gateway routes remain intentionally unregistered.
- **Public state:** 1,450 standard cases (1,333 Go + 26 Rust + 91 maintained
  client), rollout 16/59 overall, Phase 6 at 6/7, and production readiness
  false. GH-121 remains the explicit browser query-transport rollout blocker.
- **Safety:** no deployment, infrastructure, production, key, signing,
  broadcast, account, fund, or proof-system action occurred.

### Sol review feedback

Approved and merged after complete local, independent, exact-head, and
PR-native verification.

---

## 2026-08-07 01:50 EEST GH-116 retired custom ABCI query shim → Review

- **Changed:** removed the application `Query` override plus both dead legacy
  module queriers; ported DEX compatibility tests to the supported QueryServer;
  added registration, live gRPC-over-ABCI, retired-path, documentation, and CI
  contracts; documented the exact seven-governance/ten-DEX gRPC method
  boundary and the intentional absence of custom grpc-gateway routes.
- **Consumer census:** no maintained source uses `custom/truedemocracy/...` or
  `custom/dex/...`. The maintained browser client's separate unregistered
  `/truerepublic/...` REST aliases are explicitly isolated in GH-121 and remain
  a rollout blocker, not a reason to retain this unrelated shim.
- **Kimi contribution:** bounded secret-free independent review found no P0
  and confirmed removal is safe; its valid P1 recommendations (port the DEX
  tests and correct every active API document) are implemented. The
  pre-existing browser transport defect became GH-121. A final read-only
  implementation-head re-review independently reproduced the route census,
  focused tests, counts, build, vet, and contracts with 0 P0 / 0 P1 / 0 P2.
- **Tests:** `make verify` PASS (build, vet, race/coverage, 1,333 Go cases);
  focused restart/export/import PASS including the real 47.09-second node
  restart; Rust fmt/clippy/build/test PASS (26 cases), `cargo audit` PASS with
  five existing warning-only advisories; maintained-client clean install,
  lint, 85 Vitest + six audit-policy cases, build at 318.91 kB gzip, and live
  High audit PASS; consistency, both client-retirement contracts, custom-query
  retirement, shell syntax, workflow YAML, status JSON, and diff checks PASS.
- **Prerequisite:** GH-122 / PR #123 merged as `1257484`; this branch is
  rebased on the patched main and its client/security exact-head runs pass.
- **Risk:** Medium protocol cleanup, locally mitigated. Production readiness
  remains false; no deployment, infrastructure, production, key, signing,
  broadcast, account, fund, or proof-system action occurred.
- **Ready for:** protected PR checks, review remediation if any, Sol merge,
  issue/Bridge/GH-29 synchronization, and final-main readback.
- **Documentation reconciliation:** final Sol review also corrected the stale
  CLI inventories to the real 24 governance tx + 7 query and 7 DEX tx + 9
  query commands.

### Sol review feedback

Approved locally after complete rebased diff review and full relevant gates.

---

## 2026-08-07 01:45 EEST GH-122 js-yaml advisory prerequisite → Done

- **Issue/PR:** GH-122, PR #123; merged to `main` as `1257484`.
- **GitHub evidence:** Client Web CI `31129152758` PASS; Security Scan
  `31129153109` PASS; CodeRabbit and DeepScan PASS; the only review thread is
  resolved.
- **Boundary:** compatible dev-tool lock resolution only; audit policy remains
  fail closed.

---

## 2026-08-07 01:32 EEST GH-122 js-yaml advisory prerequisite → Review

- **Changed:** `client-web/package-lock.json` resolves transitive dev-only
  `js-yaml` 4.3.1 instead of vulnerable 4.3.0; no application source or audit
  policy changed.
- **Tests:** `npm ci` PASS; `npm run lint` PASS; `npm test -- --run` PASS
  (85 Vitest + six audit-policy cases); `npm run build` PASS at 318.91 kB gzip;
  `npm run audit:high` PASS with only the existing exact BrowserRouter-bound
  advisory exception.
- **Risk:** Low. One compatible transitive development-tooling lock resolution.
- **Ready for:** protected PR checks and Sol merge.

### Sol review feedback

Approved locally after exact lockfile diff review and clean reproduction.

---

## 2026-08-07 01:30 EEST GH-122 js-yaml advisory prerequisite → In Progress

- **Branch:** `fix/GH-122-js-yaml`
- **Issue:** [GH-122](https://github.com/NeaBouli/TrueRepublic/issues/122)
- **Cause:** live npm audit now reports High GHSA-5p4m-2wfm-xmqj against
  transitive `js-yaml` 4.3.0 through ESLint; GH-116 did not introduce it.
- **Scope:** compatible lockfile-only update to patched `js-yaml`, followed by
  clean install, lint, 91 tests, build, live audit, and protected CI.
- **Risk:** Low. Development-tooling transitive only; no application source,
  audit policy, runtime behavior, deployment, production, key, signing,
  broadcast, account, or fund action.

---

## 2026-08-07 01:08 EEST GH-116 retired custom ABCI query shim → In Progress

- **Branch:** `refactor/GH-116-retire-custom-query`
- **Issue:** [GH-116](https://github.com/NeaBouli/TrueRepublic/issues/116)
- **Baseline:** exact clean `origin/main` `aa3ffa6`; no open pull request;
  current main Security Scan and Pages workflows pass.
- **Scope:** inventory every maintained repository consumer and compatibility
  claim for `custom/truedemocracy/...` and `custom/dex/...`; remove only the
  consumerless legacy application shim and dead module queriers if current
  gRPC/gateway behavior is preserved; document and test the supported query
  boundary.
- **Risk:** Medium. This intentionally removes a legacy protocol surface, so
  consumer absence, modern-route preservation, restart/export/import, and
  protected exact-head checks are mandatory.
- **Delegation:** Kimi K3 receives a bounded, secret-free independent census
  and compatibility review. Sol retains design, security, implementation
  review, complete verification, GitHub writes, merge, and closure.
- **Safety:** no deployment, infrastructure, production, key, signing,
  broadcast, account, or fund action.

### Sol review feedback

Pending inventory, implementation, complete local gates, and protected review.

---

## 2026-07-23 03:10 EEST GH-56 authenticated consensus-key rotation → Done

- **Issue/PR:** [GH-56](https://github.com/NeaBouli/TrueRepublic/issues/56),
  [PR #62](https://github.com/NeaBouli/TrueRepublic/pull/62), squash-merged to
  `main` as `80ab674`; parent rollout tracker
  [GH-29](https://github.com/NeaBouli/TrueRepublic/issues/29) is synchronized.
- **Protocol:** an authenticated atomic key swap preserves validator claims,
  emits one old-key power-zero transition, activates the replacement at H+2,
  permanently revokes the old key, and retains delayed evidence attribution.
- **Authority boundary:** bootstrap and runtime operators are independent of
  consensus identity. Self-, cross-, historical-consensus-derived, and module
  authorities fail closed without generating or embedding account secrets.
- **Evidence:** 717 verified cases; local rotation harness PASS in 168.12s;
  final-head Go build/vet/race/coverage PASS in 6m44s; combined recovery,
  rotation, state-sync, backup/restore, identity, and upgrade process matrix
  PASS in 9m39s; Docker, docs, security, and DeepScan PASS. The audit records
  0 FAIL / 3 WARN / 16 PASS and no P0.
- **Residual rollout boundaries:** [GH-59](https://github.com/NeaBouli/TrueRepublic/issues/59)
  wires ABCI++ evidence to slashing, [GH-60](https://github.com/NeaBouli/TrueRepublic/issues/60)
  completes inactive-validator export/import, and
  [GH-61](https://github.com/NeaBouli/TrueRepublic/issues/61) migrates legacy
  consensus-derived operator authorities.
- **Separate repository update:** developer BTC support was merged through
  [PR #58](https://github.com/NeaBouli/TrueRepublic/pull/58) after green checks;
  the existing team multisig remains unchanged.

## 2026-07-23 06:45 EEST GH-55 validator identity custody/failover → Done

- **Issues/branch:** [GH-55](https://github.com/NeaBouli/TrueRepublic/issues/55),
  follow-up [GH-56](https://github.com/NeaBouli/TrueRepublic/issues/56), parent
  [GH-29](https://github.com/NeaBouli/TrueRepublic/issues/29), branch
  `feature/GH-55-validator-identity-recovery`.
- **Scope:** prove a planned cold transfer of the coupled consensus key and
  current signer state only after the source validator stops, into a fresh
  recovery home with a new P2P identity. Add fail-closed custody and compromise
  guidance and remove unsafe plaintext/key-only restore instructions.
- **Boundary:** on-chain consensus-key rotation, permanent old-key revocation,
  and bootstrap operator-authority separation do not exist yet and remain open
  in GH-56. `remove-validator` plus re-registration is not a supported rotation
  path.
- **Local evidence:** focused real-process failover PASS in 62.17s; full
  656-case Go suite PASS; all five real recovery drills PASS together in
  636.342s; build, vet, shell syntax, JSON, docs consistency, and diff checks
  PASS. The structured audit records 0 FAIL / 1 WARN / 5 PASS, with only the
  separate GH-56 protocol gap warned.
- **Review/merge:** PR #57 final head `aa66aa9` passed Go race/coverage in
  6m55s, the five-drill multi-validator job in 8m44s, Docker in 3m33s, docs,
  Go/Rust/Node security, DeepScan, and CodeRabbit. Four valid review findings
  were fixed, four invalid findings were answered with evidence, and all eight
  threads are resolved. Squash-merged to `main` as `e8670c6`; GH-55 is closed.
- **Current state:** the cold-failover slice is complete. GH-56 remains open for
  authenticated rotation/revocation and compromised-key recovery remains a
  separate production blocker.

## 2026-07-23 05:27 EEST GH-53 persisted binary upgrade/rollback → Done

- **Issue/PR:** [GH-53](https://github.com/NeaBouli/TrueRepublic/issues/53),
  [PR #54](https://github.com/NeaBouli/TrueRepublic/pull/54), merged to `main`
  as `3e44905`; parent tracker
  [GH-29](https://github.com/NeaBouli/TrueRepublic/issues/29).
- **Changed:** added a gated four-validator process drill for compatible rolling
  binary replacement, deterministic failure before opening state, and full
  return to the baseline artifact on unchanged persisted homes. Replaced unsafe
  full-home rollback guidance with a multi-RPC, signer-state-safe runbook.
- **Evidence:** unchanged node/validator identity keys, non-regressing signing
  positions, historical/current app-hash convergence, validator power, funded
  domain persistence, ledger-valid export, and re-import. Local combined
  process run passed in 453.076s. PR final-head Go race/coverage,
  multi-validator recovery, Docker, docs, security, DeepScan, and CodeRabbit
  checks are green; all six review threads are resolved.
- **Boundary:** this closes compatible binary replacement and fail-before-open
  rollback evidence only. `x/upgrade`, consensus-breaking state migration, and
  recovery after a partially applied migration remain open rollout gates.

## 2026-07-19 03:31 EEST GH-47 CI build-and-test timeout → Done

- **Issue:** [GH-47](https://github.com/NeaBouli/TrueRepublic/issues/47),
  parent tracker [GH-29](https://github.com/NeaBouli/TrueRepublic/issues/29)
- **Changed:** added an explicit 20-minute job timeout to the Go CI
  `build-and-test` job so a stuck GitHub runner cannot leave `main`
  indefinitely in progress.
- **Tests:** local `go test ./...` → PASS (`58.913s`); `python3 -m json.tool
  docs/status.json >/dev/null` → PASS; `bash -n scripts/backup.sh
  scripts/restore.sh` → PASS; `git diff --check` → PASS; `bash
  scripts/check-consistency.sh` → PASS. GitHub `main` checks at `63b76bf` are
  green: Go CI (`build-and-test`, `multi-validator-recovery`,
  `docker-restart-smoke`), Security Scan, and Pages.
- **Risk:** Low. This changes CI runner bounds only; code and test commands are
  unchanged.
- **Closed:** GH-47 evidence is merged; GitHub main checks are green.

### Lead Dev notes

PR #46's merge commit produced a stale GitHub `build-and-test` job that stayed
in progress for hours even though local `go test ./...` passed and the
dedicated `multi-validator-recovery` and Docker jobs were green. The explicit
job timeout prevents that state from recurring.

### Codex review feedback

Approved after local verification and refreshed green GitHub main checks.

## 2026-07-19 02:52 EEST GH-45 backup/restore/export/import → Done

- **Branch/PR:** `feature/GH-45-backup-restore-drill`,
  [PR #46](https://github.com/NeaBouli/TrueRepublic/pull/46), merged to
  `main` as `26bf44b7933c25f379db475fd34d2cfb8e49c626`
- **Issues:** [GH-45](https://github.com/NeaBouli/TrueRepublic/issues/45),
  parent tracker [GH-29](https://github.com/NeaBouli/TrueRepublic/issues/29)
- **Changed:** hardening the operator backup path into a sanitized chain-data
  artifact that excludes node keys, validator keys, validator signing state,
  and keyrings; adding `scripts/restore.sh` to restore sanitized data into an
  already initialized target home while preserving local keys; adding
  `TestMultiValidatorBackupRestoreExportImport`, a gated process drill that
  backs up a live full node, verifies the archive contains no private/signer
  artifacts, restores into a fresh home, restarts/catches up, checks common app
  hash, exports restored state, validates ledger invariants, and re-imports the
  export.
- **Tests:** `bash -n scripts/backup.sh scripts/restore.sh` → PASS; `go test .
  -run 'TestMultiValidatorBackupRestoreExportImport|TestConfigureGenesisValidatorSetBuildsExactBankBackedSet'
  -count=1 -timeout=300s -v` → PASS/SKIP as expected without the smoke env;
  `TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test . -run
  TestMultiValidatorBackupRestoreExportImport -count=1 -timeout=420s -v` →
  PASS (`88.224s` package runtime); `go test ./...` → PASS (`58.843s`);
  `bash scripts/check-consistency.sh` → PASS;
  `TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test . -run
  '^(TestMultiValidatorConsensusRecovery|TestMultiValidatorTrustedSnapshotStateSync|TestMultiValidatorBackupRestoreExportImport)$'
  -count=1 -timeout=720s -v` → PASS (`290.498s`); after the CI timing
  hardening, `TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test . -run
  TestMultiValidatorTrustedSnapshotStateSync -count=1 -timeout=420s -v` → PASS
  (`127.784s`). GitHub PR #46 checks are green: `build-and-test`,
  `multi-validator-recovery`, `docker-restart-smoke`, docs check, `go-vuln`,
  Rust audit, maintained and legacy Node audits, DeepScan, and CodeRabbit.
- **Risk:** High. Backup/restore touches operator disaster recovery and key
  safety; the backup artifact must not leak private validator/node material or
  silently replace target keys during restore.
- **Closed:** GH-45 evidence is merged and closed; GH-29 remains open as the
  parent rollout tracker.

### Lead Dev notes

The restore path intentionally requires an initialized target home. It restores
chain data/config while preserving locally generated identity and signing
material. This is rollout evidence only, not production approval.

### Codex review feedback

The current evidence proves sanitized backup artifacts, fresh-home restore
without key replacement, restored catch-up/app-hash convergence, exported
ledger validation, and re-import. This remains rollout evidence only, not
production approval.

## 2026-07-19 01:42 EEST GH-43 trusted snapshot state sync → Done

- **Branch/PR:** `feature/GH-43-trusted-snapshot-state-sync`,
  [PR #44](https://github.com/NeaBouli/TrueRepublic/pull/44), merged to
  `main` as `12a37339e9cff957d1b44413aa36160aed4e8d29`
- **Issues:** [GH-43](https://github.com/NeaBouli/TrueRepublic/issues/43),
  parent tracker [GH-29](https://github.com/NeaBouli/TrueRepublic/issues/29)
- **Changed:** adding a gated process-level drill for a fresh node catching up
  through CometBFT/Cosmos state sync from trusted snapshot metadata. The drill
  enables snapshots on the trusted validators, derives trust height and trust
  hash from a running trusted RPC endpoint, starts a fresh non-validator node
  with state sync enabled, then verifies catch-up, app-hash convergence,
  validator-power visibility, and exported ledger validity.
- **Tests:** `go test . -run
  'TestMultiValidatorTrustedSnapshotStateSync|TestConfigureGenesisValidatorSetBuildsExactBankBackedSet'
  -count=1 -timeout=300s -v` → PASS/SKIP as expected without the smoke env;
  `TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test . -run
  TestMultiValidatorTrustedSnapshotStateSync -count=1 -timeout=360s -v` → PASS
  (`130.528s` package runtime); `TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test
  . -run '^(TestMultiValidatorConsensusRecovery|TestMultiValidatorTrustedSnapshotStateSync)$'
  -count=1 -timeout=480s -v` → PASS (`197.835s`); `go test ./...` → PASS
  (`65.114s`); `bash scripts/check-consistency.sh` → PASS. GitHub PR #44
  checks are green: `build-and-test`, `multi-validator-recovery`,
  `docker-restart-smoke`, docs check, `go-vuln`, Rust audit, maintained and
  legacy Node audits, DeepScan, and CodeRabbit.
- **Risk:** Medium-high. State sync is disaster-recovery critical and must not
  trust hardcoded metadata, stale RPC endpoints, or app-state that diverges from
  the trusted validator set.
- **Closed:** GH-43 evidence is merged and closed; GH-29 remains open as the
  parent rollout tracker.

### Lead Dev notes

This remains gated behind `TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1` because it
starts real CometBFT processes. It is rollout evidence only, not production
approval.

### Codex review feedback

The current evidence proves trusted snapshot state-sync catch-up from derived
trust metadata, app-hash convergence after sync, validator-power visibility from
the synced node, and bank-backed exported state. This remains rollout evidence
only, not production approval.

## 2026-07-19 00:54 EEST GH-41 network partition recovery → Done

- **Branch/PR:** `feature/GH-41-network-partition-recovery`,
  [PR #42](https://github.com/NeaBouli/TrueRepublic/pull/42), merged to
  `main` as `8544943dd6fab483884392f1f04e83acbeb8f3f7`
- **Issues:** [GH-41](https://github.com/NeaBouli/TrueRepublic/issues/41),
  parent tracker [GH-29](https://github.com/NeaBouli/TrueRepublic/issues/29)
- **Changed:** added `TestMultiValidatorNetworkPartitionRecovery`, a gated
  four-validator process harness where a 3-of-4 quorum starts without the fourth
  peer, the isolated validator starts with no peers, the quorum commits a real
  governance transaction during the partition, and the isolated validator then
  reconnects, catches up, shares the same app hash, and exports
  ledger-valid state.
- **Tests:** `TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test . -run
  TestMultiValidatorNetworkPartitionRecovery -count=1 -timeout=300s -v` → PASS
  (`104.175s` package runtime); `TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test .
  -run 'TestMultiValidatorConsensusRecovery|TestMultiValidatorJoinReplacementLifecycle|TestMultiValidatorNetworkPartitionRecovery'
  -count=1 -timeout=600s -v` → PASS (`392.147s`); `go test ./...` → PASS.
  GitHub checks for PR #42 are green: `build-and-test`,
  `multi-validator-recovery`, `docker-restart-smoke`, `check`, `go-vuln`,
  Rust audit, Node audits, DeepScan, and CodeRabbit.
- **Risk:** Medium-high. The task exercises consensus/process behavior and must
  prove no app-hash, bank-supply, module-escrow, reserve, or validator-power
  divergence after recovery.
- **Closed:** GH-41 evidence is merged and closed; GH-29 remains open as the
  parent rollout tracker.

### Lead Dev notes

The implementation intentionally keeps the network-failure evidence behind
`TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1`. The public recovery count is unchanged
because this is a separately gated process harness, not part of the normal
recovery-verified arithmetic.

### Codex review feedback

The current evidence proves quorum-side progress during a minority partition,
delayed/isolated peer recovery, app-hash convergence, validator-power
preservation, and bank-backed export validation after recovery. This remains
rollout evidence only, not production approval.

## 2026-07-15 13:20 EEST GH-39 validator lifecycle evidence → Done

- **Branch/PR:** `feature/GH-39-validator-lifecycle`,
  [PR #40](https://github.com/NeaBouli/TrueRepublic/pull/40), merged to
  `main` as `ad30d188c7956c28cff5bf53304bc04848ba569a`
- **Issues:** [GH-39](https://github.com/NeaBouli/TrueRepublic/issues/39),
  parent tracker [GH-29](https://github.com/NeaBouli/TrueRepublic/issues/29)
- **Changed:** fixed validator-removal ABCI evidence by retaining one-shot
  power-zero tombstones; wired Cosmos SDK v0.50 signing contexts consistently
  across InterfaceRegistry, TxConfig, BaseApp, CLI, event generation, and
  dynamic Msg decoding; added a gated six-node process harness for validator
  join, replacement, full-node catch-up, and restart-safe validator-set
  convergence; hardened the smoke harness to verify delivered tx results via
  CometBFT RPC.
- **Evidence:** GitHub checks for PR #40 are green: `build-and-test`,
  `multi-validator-recovery`, `docker-restart-smoke`, docs consistency,
  CodeRabbit, DeepScan, Go/Rust security scans, and maintained/legacy Node
  audits. Local evidence: `TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test . -run
  TestMultiValidatorJoinReplacementLifecycle -count=1 -timeout=300s` → PASS
  (`117.638s`); `go test ./...` → PASS; focused validator removal/update and
  tx round-trip tests → PASS.
- **Boundary:** public validator leave/removal remains constrained by existing
  stake-withdrawal economics, so process-level join/replacement/restart evidence
  is paired with Keeper/ABCI power-zero regression coverage for leave.
- **Closed:** GH-39 evidence is merged; GH-29 remains open as the parent
  rollout tracker.

### Codex review feedback

The previous GH-39 blockers were real infrastructure bugs: custom hand-written
messages were missing SDK v0.50 signer resolvers in multiple signing contexts,
and sync broadcast hid DeliverTx failures. The branch now verifies delivered
transactions and proves a new validator joins the CometBFT set before a
replacement validator catches up and joins cleanly.

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

## 2026-08-04 12:08 EEST GH-102 maintained-client audit unblock → In Progress

- **Branch:** `fix/GH-102-maintained-client-audit`
- **Issue:** [GH-102](https://github.com/NeaBouli/TrueRepublic/issues/102)
- **Trigger:** GH-101 PR #103 exact head is blocked by the required
  `node-audit-client` check on newly published dependency advisories in the
  unchanged maintained client.
- **Scope:** first bounded remediation slice: remove every high/critical
  maintained-client audit finding with the smallest compatible lock/package
  update, preserving lint, eight tests, production build, routing behavior, and
  exact security workflow. Mobile legacy advisories stay explicitly open in
  GH-102 for a later reviewed migration.
- **Boundary:** repository dependency metadata/tests only. No wallet key,
  signing, broadcast, funds, server, deployment, app-store, or production
  action.
- **Ready for:** exact dependency-tree analysis, minimal update, focused client
  gates, review, protected PR, merge, then GH-101 rebase.

---

### 2026-08-01 23:43 EEST real capacity qualification checkpoint

- The maintained four-validator loopback qualification passed with 96/96
  authenticated transactions committed in 24 consecutive four-signer waves,
  zero transaction failures, and an exact 24-block workload delta.
- Observed evidence: 131,924 ms workload duration, 727 milli-TPS, 5,496 ms
  average block time, 5,893 ms p95 commit latency, 6,599 ms maximum commit
  latency, maximum node data growth 1,222,727 bytes, maximum log growth
  137,971 bytes, and maximum RSS 154,902,528 bytes.
- Restart, common app hash, equal validator power, ledger validation, required
  app/Comet metrics, and runtime snapshot retention all passed. The settled
  snapshot store contained exactly three retained heights.
- The separate strict CLI verifier accepted the generated evidence with
  `valid: true`, `committed_count: 96`, and no violations.
- Two preceding real runs exposed a harness defect rather than a chain defect:
  an assumed public `/snapshots` RPC route does not exist in this CometBFT
  surface. Sol replaced it with a fail-closed count of the real local snapshot
  height directories after halting new commits, and added a focused regression
  test. Full repository gates and publication remain pending.

### 2026-08-02 00:42 EEST local completion and audit checkpoint

- Final implementation review is complete. `docs/agent-bridge/GH97_AUDIT.md`
  records `0 FAIL / 0 WARN / 7 PASS`; Terra's independent final review found no
  remaining P0/P1/P2 issue. Both bounded Kimi attempts remained provider-quota
  blocked with HTTP 403 and produced no output or diff.
- PASS: package-selector contract, documentation consistency, workflow YAML,
  maintained contract CLI/JQ contract, `git diff --check`, full CGO build, full
  Vet, focused policy/repository/metrics/snapshot tests, and the complete normal
  Go suite over every repository-owned package.
- PASS: all non-root packages under Race/Coverage and every root test under the
  same instrumentation. The first combined local Race command exhausted the
  nearly-full workstation volume while the existing lifecycle test performed
  its nested daemon build; after safe Go-cache cleanup, the race-instrumented
  lifecycle test passed separately in 374.08 seconds. This was an environment
  capacity failure, not a test assertion or data race; GitHub will rerun the
  canonical combined command on its clean runner.
- The maintained real capacity run remains PASS with 96/96 commits and strict
  CLI evidence verification. Local implementation/review is Done; protected
  PR, exact-head GitHub checks, merge, issue closure, and public-status sync are
  next. Production and live rollout gates remain open.

### 2026-08-02 00:57 EEST PR #98 review remediation

- Exact-head GitHub capacity qualification passed in 6m09s; build/race/coverage
  passed in 7m29s; Docker/Compose/monitoring passed in 7m15s; docs, Go/Rust/Node
  security, and DeepScan also passed before the review update.
- CodeRabbit raised one minor inline finding and one trivial test observation.
  Sol verified both: the macOS RSS `ps` fallback now inherits `t.Context()` via
  `exec.CommandContext`, and the supply-cap mismatch test is correctly named
  while a distinct `math.MaxInt64 + 1` case covers the checked-add overflow
  branch.
- Focused policy/root tests, package/root Vet, formatting, and diff checks pass.
  A small remediation commit and a fresh exact-head GitHub run are next; the
  prior independent audit result remains unchanged after fixing both findings.

## 2026-07-31 00:15 EEST PR #86 merged; GH-85 closed

- Exact head `021d5c5` passed the complete GitHub matrix: build/race/coverage,
  Docker observability runtime proof, all eight multi-validator recovery
  scenarios, docs consistency, Go/Rust/Node security gates, and DeepScan.
- Thread-aware review reported zero unresolved threads; all three CodeRabbit
  findings were fixed, replied to, and resolved.
- Owner-authorized admin squash merged PR #86 as `cd44fec`; GitHub
  automatically closed GH-85.
- Post-merge `main` Go CI, Security Scan, and Pages deployment are running.
- **Next:** require those exact-main checks to pass, then publish the
  documentation-only closure update for GH-29, rollout counts, test counts,
  Bridge, and the public GitHub Pages status.

---

## 2026-07-31 00:30 EEST GH-85 closure synchronization → Locally green

- Exact merged-main `cd44fec` passed post-merge Go CI, including Docker and all
  recovery scenarios, plus Security Scan and GitHub Pages deployment.
- Fresh local JSON output confirms 1,071 passing Go tests; with 26 Rust and
  eight maintained-client tests, the public total is 1,105.
- Canonical GH-29 now records 11/59 complete, 11/51 phase work, and Phase 6
  4/7; the GH-85 combined checkbox links GH-85/PR #86.
- Public status sources now use 19%, 22%, and 57% respectively and retain the
  explicit recovery-only, no-production-SLO, and no-external-paging boundary.
- **Independent review:** Spark identified every current/public stale count
  without touching historical logs; Kimi verified the arithmetic and found the
  then-pending GH-29 synchronization, which is now resolved.
- **PASS:** documentation consistency, JSON validation, diff hygiene, and the
  1,071-case local Go suite.
- **Next:** publish this documentation-only closure branch, require its exact
  GitHub checks, merge, and verify the deployed public page.

---

## 2026-07-31 00:38 EEST PR #87 review findings → Fixed

- CodeRabbit identified two valid publication-state issues on exact head
  `f600cf7`: the current-state wording could imply the 1,105 public status was
  already merged, and the publication checkbox was prematurely complete.
- Reworded the evidence as fresh tests against merged implementation
  `cd44fec`, explicitly branch-local until PR #87 merges, and reopened the
  publication checkbox through exact-head merge and deployed-page verification.
- Aligned the roadmap update date with the GitHub publication date
  `2026-07-30`.
- **Next:** publish the corrected head, resolve both review threads, and require
  its full GitHub matrix.

---

## 2026-07-31 00:46 EEST GH-85 public closure → Done

- Documentation PR #87 passed its exact-head docs, DeepScan, CodeRabbit,
  Go/Rust/Node security, and review gates with zero unresolved threads.
- Owner-authorized squash merge published it as `0aaba4b` on `main`.
- Exact merged-main Security Scan and GitHub Pages deployment pass.
- Live `https://neabouli.github.io/TrueRepublic/` returns HTTP 200 and shows
  1,105 tests, 11/59 full rollout (19%), 11/51 phase work (22%), Phase 6 at
  4/7 (57%), production topology as next, and the explicit non-production
  boundary.
- GH-29 remains synchronized at 11/59 with its GH-85/PR #86 checkbox complete.
- **Status:** GH-85 implementation and public closure are complete; external
  paging, production SLOs, production infrastructure, and rollout approval
  remain outside the completed scope.

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

## 2026-07-22 23:04 EEST GH-48/GH-50 post-merge audit remediation → Done

- **Branch:** `fix/GH-48-post-merge-status`
- **Issues:** [GH-48](https://github.com/NeaBouli/TrueRepublic/issues/48),
  [GH-50](https://github.com/NeaBouli/TrueRepublic/issues/50)
- **PR:** [#49](https://github.com/NeaBouli/TrueRepublic/pull/49), squash
  merged to `main` as `7dbde858d0d3d5410f22d16a1a3bac614325d925`
- **Changed:** current contributor, audit, operator, limitations, security, and
  bridge status now distinguish the merged recovery foundation from remaining
  production-rollout work; `golang.org/x/text` is updated from v0.37.0 to
  v0.39.0 for GO-2026-5970; historical bridge entries remain unchanged
- **Tests:** maintained client `npm ci`, lint, 8 tests, build, and audit → PASS;
  Rust format, Clippy `-D warnings`, and 26 workspace tests → PASS; docs,
  shell, JSON, workflow YAML, secret-literal, and dependency-version checks →
  PASS; isolated Go build/vet/655 tests → PASS (`75.399s` root package);
  `cargo audit` → PASS with six known allowed transitive warnings; after the
  GH-50 update, exact Security Scan filtering → PASS and isolated Go build,
  vet, and 655 tests → PASS (`64.499s` root package)
- **Risk:** Medium — security dependency update plus documentation/audit truth;
  no intended consensus, ledger, wallet, contract, or deployment behavior change
- **Ready for:** next GH-29 rollout task

### Lead Dev notes

The audit found no recovery-foundation failure. Three residual warnings remain
explicit: incomplete GH-29 rollout evidence, root Go wildcard discovery under
an installed frontend dependency tree ([GH-51](https://github.com/NeaBouli/TrueRepublic/issues/51)),
and the maintained-client bundle size.
Docker is unavailable in this local environment; the latest canonical GitHub
Docker restart check is green.

### Codex review feedback

PR #49's first Security Scan exposed newly published GO-2026-5970 in
`golang.org/x/text` v0.37.0. The minimal v0.39.0 remediation removes that
finding locally; the exact CI gate, build, vet, and full tests pass. Final
head checks are green: Go build/race/coverage, multi-validator recovery, Docker
restart, Go/Rust/Node security, docs, DeepScan, and CodeRabbit. The valid audit-
tally finding was fixed and its review thread resolved before merge.

---

## 2026-07-22 11:24 UTC GH-51 isolate root Go package selection → Done

- **Branch:** `fix/GH-51-go-package-selector`
- **Issue:** [GH-51](https://github.com/NeaBouli/TrueRepublic/issues/51)
- **Changed:**
  - `scripts/go-packages.sh` — explicit repository-owned Go package selector
    and generic command wrapper
  - `scripts/test-go-packages.sh` — ignored `node_modules` regression probe
  - `Makefile`, `.github/workflows/go-ci.yml` — shared local/CI verification
  - contributor, installation, testing, wiki, audit, and agent-bridge guidance
- **Tests:** selector regression → PASS; concurrent `npm ci` plus Go build/vet/
  race/coverage → PASS; normal 655-case Go suite → PASS; consensus recovery,
  state-sync, and backup/restore harnesses → PASS (`265.381s`); docs consistency,
  workflow YAML, shell syntax, and diff checks → PASS
- **Risk:** Low — developer tooling and CI package discovery only; no runtime,
  consensus, ledger, wallet, contract, dependency, or deployment behavior change
- **PR:** [#52](https://github.com/NeaBouli/TrueRepublic/pull/52), squash
  merged to `main` as `ae7105afb564547c2eb17933d4a325267be51d46`
- **Ready for:** next GH-29 rollout task

### Lead Dev notes

The wrapper uses Git's tracked/untracked-but-not-ignored source view, so local
new packages remain visible while ignored dependency installations do not.
Docker is unavailable locally; the PR Docker restart gate is mandatory.

### Codex review feedback

Approved and merged. CodeRabbit's four valid findings were fixed in `6a668ee`,
answered, and all threads resolved. Final head passes all 11 GitHub checks:
Go build/vet/race/coverage, multi-validator recovery, Docker restart,
Go/Rust/Node security, docs, DeepScan, and CodeRabbit. GH-51 is closed.

---

## 2026-07-26 19:40 EEST GH-60 inactive validator genesis round-trip → In Progress

- **Branch:** `feature/GH-60-inactive-validator-genesis`
- **Issue:** [GH-60](https://github.com/NeaBouli/TrueRepublic/issues/60);
  parent rollout tracker
  [GH-29](https://github.com/NeaBouli/TrueRepublic/issues/29)
- **Scope:** preserve legitimate inactive, excluded, jailed, and under-staked
  validator claims exactly across export/import, including all domains, jail,
  missed blocks, power semantics, revocations, rotations, signing/infraction
  state, pending exits, ledger backing, and supply parity. Malformed
  resurrection and revoked-key reuse must fail closed.
- **Roles:** Codex Sol owns scope, architecture, security, integration,
  external writes, complete verification, PR/merge, and closure. Kimi K3 owns
  one bounded secret-free local implementation block and may not delegate.
- **Baseline:** clean `origin/main` at `996ae05`; GH-59 is already closed.
- **Risk:** High — consensus state, validator custody, genesis validation, and
  ledger parity.
- **Current verification:** not run yet for GH-60.
- **Next:** inspect the exact export/validation/import paths, delegate the
  bounded representation/test slice to Kimi, review the diff, then run focused,
  complete Go, race/coverage, docs, and real process recovery gates.

---

## 2026-07-26 20:50 EEST GH-60 local merge gate → PASS, publication pending

- **Branch:** `feature/GH-60-inactive-validator-genesis`
- **Issue:** [GH-60](https://github.com/NeaBouli/TrueRepublic/issues/60);
  parent [GH-29](https://github.com/NeaBouli/TrueRepublic/issues/29)
- **Status:** local implementation and review complete; GitHub publication,
  final-head CI/review, merge, issue closure, and tracker synchronization remain.
- **Changed:** explicit active/inactive genesis representation, complete domains
  and stored power, exact inactive custody/jail/liveness recovery, legacy
  compatibility, fail-closed resurrection/key reuse, real process re-import
  assertions, reliable liveness-window timeout, GH-60 audit, Bridge, rollout,
  wiki, and verified 733-case public status.
- **Kimi contribution:** bounded genesis-core implementation and independent
  read-only deep review. The review found one P2 malformed-state gap: inactive,
  unjailed positive power with no domains. Sol tightened the invariant, added a
  regression, and reran all gates. Final review result: 0 P0 / 0 P1 / 0 P2.
- **Tests:** focused GH-60 tests PASS; real four-process downtime/double-sign
  export/import drill PASS in 441.52s; package selector, Go build, vet, complete
  race/coverage, docs consistency at 733 cases and 21M PNYX, JSON, and diff
  checks PASS. Coverage: root 68.5%, truedemocracy 62.2%.
- **Risk/blocker:** no local code blocker. Final GitHub Docker, security,
  review, and branch-head checks remain mandatory before merge. No production
  or deployment approval is claimed.
- **Next:** commit, push, open PR, update GH-60/GH-29, then monitor and remediate
  final-head GitHub checks/review before merge and Bridge closure.

---

## 2026-07-26 21:15 EEST PR #67 maintained-client security remediation → PASS locally

- **Trigger:** GitHub `node-audit-client` failed twice after clean `npm ci`.
  The retiring quick-audit endpoint returned HTTP 400; the official bulk
  advisory response exposed two newly published High findings rather than a
  GH-60 code regression.
- **Changed:** `postcss` moves to `^8.5.23`; `brace-expansion` is pinned to
  patched `5.0.8`; `minimatch` is consistently overridden to `10.2.5` so the
  old ESLint transitive path remains API-compatible.
- **Verification:** clean `npm ci`, ESLint, eight Vitest cases, TypeScript/Vite
  production build, and `npm audit --audit-level=high` PASS. Audit still reports
  two Moderate React Router findings below the existing High gate.
- **GitHub status before refresh:** Docs, Docker, Go vulnerability, Rust audit,
  legacy Node audits, and DeepScan PASS; Go build/race, combined process matrix,
  CodeRabbit, and refreshed maintained-client Security Scan remain pending.
- **Risk:** dependency-only security remediation in the maintained client; no
  consensus, ledger, validator, wallet-key, deployment, or production action.

---

## 2026-07-26 21:17 EEST PR #67 review remediation → PASS locally

- **Status:** all four CodeRabbit review threads were triaged. Documentation,
  reproduction commands, GH-60 local/publication status, expected validation
  errors, and panic-recovery test scope were corrected.
- **Preserved invariant:** the request to classify a malformed unjailed,
  positive-power runtime validator as inactive during export was rejected.
  Doing so would neither round-trip nor pass explicit inactive validation and
  would conceal corrupt active state instead of failing closed.
- **Verification:** `go test ./x/truedemocracy -count=1`, documentation
  consistency at 733 cases / 21M PNYX, and `git diff --check` PASS.
- **Kimi contribution:** unchanged from the bounded implementation and
  independent review recorded above; Sol owns this review remediation,
  integration, final tests, GitHub replies, and merge decision.
- **Risk/blocker:** no local blocker. Full final-diff Go gates and refreshed
  final-head GitHub CI/review remain required before merge.

---

## 2026-07-26 21:25 EEST PR #67 final review-diff gate → PASS

- **Tests:** all five uncached Go package tests, Vet, CGO build, full race and
  coverage, docs consistency, and diff checks PASS.
- **Coverage:** root 68.5%, token 92.6%, treasury 97.0%, DEX 45.3%, governance
  62.2%.
- **Status:** review remediation is ready to publish. No local code, test,
  security, or documentation blocker remains.
- **Next:** commit/push the review fixes, answer and resolve every review
  thread, then require the refreshed GitHub head to pass before merge.

---

## 2026-07-26 21:40 EEST GH-60 / PR #67 → Done

- **Merge:** final head `0cad6df` passed all 12 GitHub checks and PR #67 was
  squash-merged to `main` as `c5a3d38`; GitHub closed GH-60.
- **GitHub evidence:** Go build/vet/race/coverage PASS in 7m14s; combined real
  multi-validator recovery PASS in 11m55s; Docker restart PASS in 3m33s; docs,
  Go/Rust/Node security, maintained-client build, DeepScan, and CodeRabbit PASS.
- **Review:** all four review threads are resolved. Kimi's independent final
  review recorded 0 P0 / 0 P1 / 0 P2. The admin merge was used only because
  the owner-approved PR had no second GitHub account available to satisfy the
  formal approval rule.
- **Status:** inactive validator claims now round-trip on `main`; Bridge,
  roadmap, audit, issue, and GH-29 closure evidence are being synchronized.
  GH-61 is the next bounded consensus-state task. No rollout or deployment is
  approved.

---

## 2026-07-26 21:56 EEST GH-61 legacy operator-authority migration → In Progress

- **Branch:** `feature/GH-61-legacy-authority-migration`
- **Issue:** [GH-61](https://github.com/NeaBouli/TrueRepublic/issues/61);
  parent rollout tracker
  [GH-29](https://github.com/NeaBouli/TrueRepublic/issues/29)
- **Scope:** design and prove a deterministic, governance-controlled halt /
  export / migration / re-import ceremony for pre-GH-56 validators whose
  operator addresses were derived from consensus keys. Every replacement
  operator must authenticate independently; consensus keys may never confer
  account authority.
- **Required invariants:** preserve Comet validator identity/power, validator
  stake and module escrow, domains, revocations, signing/infraction history,
  pending exits, auth accounts, bank supply, and replay safety. Partial,
  duplicate, cross-coupled, wrong-height, or unauthenticated plans fail before
  mutation; rollback may not create a dual signer.
- **Roles:** Codex Sol owns architecture, security, integration, full tests,
  GitHub writes, merge, and closure. Kimi K3 receives a bounded secret-free
  architecture/development block and may not delegate.
- **Baseline:** clean `origin/main` at `8c0c555`; GH-60 is closed and no pull
  request is open.
- **Risk:** High — consensus identity, operator authorization, genesis
  migration, escrow ownership, and replay safety.
- **Current verification:** not run yet for GH-61.
- **Next:** independently review the safest migration boundary, define the
  canonical signed plan and fail-closed state rewrite, then implement focused
  unit/integration/process evidence before the complete repository gate.

---

## 2026-07-26 22:09 EEST GH-61 architecture review → Security decision pending

- **Kimi contribution:** read-only repo-wide review mapped every operator
  reference that a migration must rewrite: validators and indexes, domains,
  consensus history, revocations, signing/infraction state, pending rotations
  and exits, auth/bank balances, DEX provider claims, Comet identity/power, and
  the replay cursor. It recommends a pure atomic halt/export/transform/re-import
  boundary because neither `x/gov` nor `x/upgrade` is wired.
- **Sol security finding:** the proposed legacy-consent and validator-quorum
  signatures used the old consensus keys. That would infer account/governance
  authority from consensus identity and directly contradict GH-61's acceptance
  criterion. This design is rejected and no code was written from it.
- **Confirmed constraints:** current genesis validation cannot re-import a
  pre-GH-56 coupled validator; current registration and rotation deliberately
  refuse the same coupling. Stake remains module-escrowed and heights must stay
  monotonic across export/import.
- **Verification:** read-only architecture inspection only; no GH-61 tests
  exist or ran yet.
- **Risk/blocker:** the legacy state may contain no independently controlled
  governance signer at all. Before implementation, the authorization anchor
  must be proven independent from every active, historical, and revoked
  consensus key; otherwise the task must add an explicit governance anchor or
  remain fresh-genesis-only.
- **Next:** adversarially review the authorization-anchor question, then record
  the final architecture decision before assigning a bounded implementation.

---

## 2026-07-26 22:20 EEST GH-61 authorization-anchor review → No legacy anchor

- **Decision:** pre-GH-56 exportable state contains no independently
  controlled, machine-verifiable authority that the old protocol empowered to
  replace validator operators. The first transition therefore cannot honestly
  be described as governance-authorized from legacy state.
- **Evidence:** `x/gov` and `x/upgrade` are not wired; the `gov` module address
  cannot sign; legacy validator operators and the bootstrap domain admin were
  derived from consensus keys; domain permission/ZKP keys and arbitrary
  auth/bank accounts have no protocol-granted authority over validator
  identity. Reusing any of them would either violate the no-consensus-authority
  invariant or invent authority retroactively.
- **Kimi contribution:** bounded read-only adversarial review independently
  enumerated every candidate anchor, confirmed the legacy bootstrap coupling
  from pre-GH-56 source history, and found no valid old-state signer. No Kimi
  write diff was produced.
- **Supported boundary:** the one-time recovery must be a published, reviewed,
  state-preserving fresh-genesis ceremony bound to an exact halt height and
  source app hash. Each replacement account must provide an independent
  proof-of-possession over the canonical migration descriptor. A new,
  consensus-independent governance commitment may begin in the migrated
  genesis for future authority, but it cannot retroactively authorize this
  first rewrite.
- **Verification:** repository and history inspection only; no code changed and
  no GH-61 tests ran. Existing local modifications remain Bridge/audit entries
  only.
- **Risk/blocker:** the original phrase “governance-controlled migration” is
  impossible if it means authority rooted in pre-GH-56 state. Implementation
  must explicitly document the reviewed-genesis trust boundary and must not
  claim stronger provenance.
- **Next:** publish the architecture finding to GH-61, narrow the acceptance
  wording to the honest trust boundary, then implement the canonical
  descriptor/proof verifier and deterministic exact-state transformer with
  rollback and multi-validator drill evidence.

---

## 2026-07-26 22:38 EEST GH-61 canonical descriptor/proofs → Local PASS

- **Implemented:** new `migration` package defines descriptor v1 binding source
  and target chain IDs, positive halt height, 32-byte source app hash,
  transform ID, and the complete sorted one-to-one operator mapping. Every
  replacement proves possession of its fresh SDK ed25519 or secp256k1 account
  key over one domain-separated deterministic encoding of the full descriptor.
- **Fail-closed checks:** malformed or mixed-prefix account addresses, wrong key
  derivation/type/signature, duplicate or unsorted mappings, old/new
  cross-coupling, field mutation/replay, malformed supplied consensus keys, and
  every new address derived from a caller-supplied active, historical, pending,
  or revoked consensus key are rejected.
- **Kimi contribution:** owned only
  `migration/legacy_authority.go` and
  `migration/legacy_authority_test.go`; implemented the bounded core and 36
  negative cases. Its tests exposed and it fixed a stale-index comparator bug
  in `SortMappings` before handoff.
- **Sol review:** read the complete write diff, added stable JSON field names,
  required exact 20-byte address payloads already in `SigningBytes`, and added
  regressions for both. No governance authority is inferred: the public API and
  package documentation state that proofs establish fresh-key possession only.
- **Verification:** `gofmt -l migration` clean; `go vet ./migration` PASS;
  `go test ./migration -count=1` PASS in 0.946s;
  `go test -race ./migration -count=1` PASS in 2.578s;
  `git diff --check` PASS.
- **Risk/blocker:** consensus exclusion is only complete when the transformer
  supplies every exported active, historical, pending, and revoked key. No
  transformer, CLI, full-genesis reconciliation, rollback drill, or repo-wide
  gate exists yet, so GH-61 remains In Progress.
- **Next:** implement the typed, atomic full-genesis transformer and prove that
  it discovers the complete legacy key set, rewrites every operator-bearing
  claim, rejects stale/unknown references, and preserves all non-authority
  state exactly.

---

## 2026-07-26 22:59 EEST GH-61 transformer threat review → Boundary hardened

- **Security refinement:** the target chain ID must differ from the signed
  source chain ID so pre-migration account transactions cannot replay on the
  recovered chain. Descriptor verification and its negative tests now enforce
  that rule.
- **Wasm boundary:** typed Wasm creator/admin/access-list fields and arbitrary
  contract state can retain operator authority in binary or application-defined
  form. The generic transformer must reject nonempty contract state unless a
  separately reviewed contract-aware adapter proves its migration; byte-search
  alone is insufficient.
- **Auth boundary:** the transformer must reconstruct and repack only mapped
  `BaseAccount` records, preserve account number/sequence, replace the old
  public key, reject mapped module/unknown account types, and independently
  reject duplicate account numbers before and after the rewrite so SDK import
  cannot silently renumber signer identities.
- **Independent review:** Sol's read-only field audit confirmed the complete
  custom-module rewrite list and identified the Wasm/account-number gaps before
  transformer code was accepted.
- **Kimi contribution:** a second bounded implementation session developed the
  typed transform/test design but was stopped before writing files when its
  planning exceeded the bounded execution window. No partial transformer diff
  exists.
- **Verification:** after the chain-ID hardening,
  `go test ./migration -count=1` PASS in 1.334s and
  `git diff --check` PASS.
- **Status:** descriptor/proof core remains locally green; transformer remains
  In Progress. No commit, push, PR, merge, rollout, or deployment occurred.
- **Next:** resume from the reviewed design with a smaller implementation slice:
  state-derived key inventory and typed truedemocracy rewrite first, followed by
  Auth/Bank/DEX reconciliation and the explicit empty-Wasm gate.

---

## 2026-07-27 00:09 EEST GH-61 truedemocracy transform core → Local PASS

- **Implemented:** `migration/transform_core.go` accepts a complete exported
  genesis document, derives its forbidden consensus-key inventory from active
  validators, history, revocations, both pending-rotation keys, and pending
  removals, then verifies the signed descriptor against that state-derived set.
- **Exact mapping gate:** every typed truedemocracy address that is both present
  and consensus-key-derived must be mapped exactly once; missing, extra,
  non-coupled, stale, invalid-proof, and replacement-collision cases fail before
  output is returned.
- **Atomic rewrite:** validators, bootstrap operators, domain admin/members,
  suggestion creators, revocations, rotations, key history, signing infos,
  infractions, and pending-removal operator/recipient fields are rewritten on
  decoded state copies. Consensus keys, stake, domains, evidence state, heights,
  and unrelated modules remain unchanged. The rewritten module passes
  `ValidateGenesisState`, and old operator literals are rejected in its JSON.
- **Delegated contribution:** the Spark worker produced an initial core/test
  draft, but its tests failed Sol's gate because they duplicated helpers,
  derived secp256k1 addresses incorrectly, used a 20-byte app hash, and built an
  invalid genesis fixture. Sol replaced the faulty suite and hardened prefix,
  panic, collision, and literal-remnant handling before acceptance.
- **Verification:** `gofmt -l migration` clean; `go vet ./migration` PASS;
  `go test ./migration -count=1` PASS in 3.726s;
  `go test -race ./migration -count=1` PASS in 3.617s;
  `git diff --check` PASS.
- **Status:** the typed custom-module core is locally green. Auth, Bank, DEX,
  Wasm, outer chain-id/height binding, CLI, process drills, full repository
  gates, commit, push, and PR remain open.
- **Next:** compose the full app-state reconciler around this core, enforcing
  unique auth account numbers, fresh BaseAccount repacking, exact balances/LP
  ownership, unchanged supply/escrow, and an empty-contract Wasm gate.

---

## 2026-07-27 00:24 EEST GH-61 full app-state reconciliation → Local PASS

- **Implemented:** `migration/app_state.go` composes the verified
  truedemocracy rewrite with typed Auth, Bank, and DEX reconciliation, binds the
  outer source/target chain IDs and exact `halt_height + 1` initial height, and
  requires all safety-critical modules to be present.
- **Auth integrity:** mapped `BaseAccount` records retain account number and
  sequence, receive the signed fresh public key, and are repacked atomically.
  Duplicate addresses/numbers, pre-existing target accounts, mapped module or
  unsupported account types, and invalid post-state fail closed. A mapped
  operator absent from Auth remains absent instead of receiving invented
  identity state.
- **Bank/DEX integrity:** operator balances, LP providers, and asset
  registration authorities move to the exact signed replacements. Supply,
  module balances, DEX invariants, and unrelated state are preserved; target
  collisions are rejected.
- **Wasm and unknown-module boundary:** any exported CosmWasm code or contract
  state is rejected until a contract-aware adapter exists. Unknown modules are
  retained, but the final application-wide scan rejects any remaining literal
  legacy operator address.
- **Review contribution:** the earlier independent transformer review supplied
  the complete Auth/Wasm and custom-module risk inventory. Sol implemented and
  reviewed this integration slice and corrected the Wasm test fixture and
  malformed-module regression before acceptance; no unreviewed delegated write
  was accepted.
- **Verification:** `gofmt -l migration` clean; `go vet ./migration` PASS;
  `go test ./migration -count=1` PASS in 2.177s;
  `go test -race ./migration -count=1` PASS in 6.398s;
  `git diff --check` PASS.
- **Risk/blocker:** trusted source-header app-hash comparison remains an
  explicit ceremony/CLI responsibility because it cannot be derived from
  genesis JSON. Nonempty Wasm state is intentionally unsupported. CLI,
  export/transform/re-import and rollback drills, repository-wide gates,
  commit, push, PR, and merge remain open.
- **Next:** expose one fail-closed offline CLI that verifies the trusted app
  hash, reads descriptor and export files without mutation, writes atomically,
  and supports the documented rehearsal and rollback procedure.

---

## 2026-07-27 00:54 EEST GH-61 canonical offline CLI and re-import → Local PASS

- **Implemented:** `truerepublicd migration legacy-authority` requires a strict
  signed descriptor, full export, new output path, and independently obtained
  trusted halt-height app hash. Descriptor unknown/trailing fields, malformed
  hashes, non-regular or aliased inputs, existing outputs, and over-limit files
  fail closed.
- **Atomic output:** the command writes and syncs a private `0600` temporary
  file, installs it through a no-overwrite hard link, syncs the directory, and
  removes partial output/temp files on every reported failure. It never prints
  descriptor proofs, public keys, addresses, or genesis content.
- **Full boundary hardening:** a nonempty embedded genesis `app_hash` is
  rejected. A present CometBFT validator set must match every active
  application validator's ed25519 key and exact power; an empty exported set is
  permitted because `InitChain` reconstructs it from application state.
  Consensus data is otherwise preserved.
- **Import gate:** before any file is created, the canonical CLI reruns the
  application's cross-module ledger validator over the transformed Auth, Bank,
  DEX, and truedemocracy state.
- **End-to-end regression:** the Cobra command consumes a valid coupled legacy
  fixture and fresh-key proof, checks the trusted app hash, transforms it,
  writes privately, removes every legacy address literal, and reimports the
  output into a fresh `TrueRepublicApp`.
- **Kimi contribution:** a bounded Kimi implementation attempt reconstructed
  the file-safety design but was stopped before writes after remaining in
  planning. Sol implemented the bounded patch, reviewed every line, identified
  and closed the missing Comet/App validator boundary, and ran all gates.
- **Verification:** `gofmt -l` and `git diff --check` clean;
  `go vet ./migration .` PASS; `go test ./migration . -count=1` PASS
  (`migration` 3.951s, root 141.937s); focused race PASS
  (`migration` 13.865s, root 13.296s); canonical transform/reimport regression
  PASS in 12.041s; built daemon help exposes the command and all four required
  flags.
- **Risk/blocker:** this proves the deterministic offline command and in-memory
  re-import, not yet a live pre-GH-56 four-validator halt/transform/restart/
  rollback ceremony. Operator documentation and that real process drill remain
  mandatory before GH-61 can close.
- **Next:** add the pinned pre-GH-56 binary process harness and operator
  runbook, prove source halt/app-hash capture, target-chain convergence, source
  rollback without dual signers, and recovered export/re-import.

---

## 2026-07-27 01:34 EEST GH-61 historical process drill and runbook → Local PASS

- **Real historical source:** the gated process harness archives and builds
  pinned pre-GH-56 revision
  `0e51a05b008f395e3f7391358e117f0817d4eb39`, starts four real coupled-authority
  validators, removes quorum, records one stable common halt-height app hash,
  stops every source signer, and exports the exact halt successor.
- **Target proof:** four fresh secp256k1 operator keys sign one canonical
  descriptor. The current daemon CLI verifies and transforms the historical
  export, then four clean target homes with a distinct chain ID progress,
  converge on one app hash, retain every expected consensus key at positive
  power, and produce a ledger-valid export that reimports into a fresh app.
- **Rollback proof:** every target signer is cleanly stopped before the
  untouched historical homes restart. The source chain resumes beyond the halt
  and reconverges, proving rollback without concurrent copies or target/source
  state mixing.
- **Protocol defect found and fixed:** the first target run exposed that
  restored genesis consensus-key history used runtime H+2 activation. At a
  non-one initial height, block H+1 therefore rejected the decided commit for
  H. Genesis activation now uses the actual initial height while runtime
  validator registration retains H+2; a focused regression covers heights
  0/1/4.
- **Harness corrections:** the first rehearsal fixed an explicit test-process
  bech32 prefix; the final assertion now requires each consensus key with
  positive power because normal validator rewards increase exact power after
  restart. Neither correction weakened application/Comet reconciliation.
- **Documentation:** added the English operator runbook with the honest
  reviewed-fresh-genesis authorization boundary, descriptor/proof format,
  halt/export/transform/start/rollback sequence, signer isolation, fail-closed
  conditions, and exact automated-evidence command. Roadmap, limitations,
  upgrades, and operator index now link the supported boundary.
- **Verification:** final
  `TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test . -run
  '^TestMultiValidatorLegacyAuthorityMigrationRollback$' -count=1
  -timeout=900s -v` PASS in 249.13s; activation regression and focused
  truedemocracy tests PASS in 3.588s; post-doc focused packages PASS
  (`truedemocracy` 2.537s, `migration` 5.105s, root 3.688s);
  docs consistency PASS; `go vet .`, `gofmt -l`, and `git diff --check` PASS.
- **Risk/blocker:** GH-61 is locally implemented and process-proven, but not
  Done. Full repository/security gates, independent final adversarial review,
  commit/push, PR, final-head GitHub workflows, review remediation, merge, and
  issue/roadmap closure remain. Nonempty Wasm state and generic in-place
  migrations remain explicitly unsupported.
- **Next:** run the complete local verification/security matrix, commission an
  independent read-only final review of the full GH-61 diff, remediate every
  valid finding, then publish the reviewable PR.

---

## 2026-07-28 20:12 EEST GH-61 exact export-artifact binding → In Progress

- **Branch:** `feature/GH-61-legacy-authority-migration`
- **Issue:** [GH-61](https://github.com/NeaBouli/TrueRepublic/issues/61)
- **Authorization:** Gio explicitly approved this exactly bounded local
  security remediation after the 2026-07-28 Bridge workflow refresh. No push,
  merge, deployment, public-network action, or follow-up task is included.
- **Definition of Ready:** bind the exact source-export bytes to the canonical
  descriptor signed by every fresh replacement operator; reject any
  post-signing byte mutation before transformation or output.
- **Scope:** `AGENTS.md` workflow prerequisite; descriptor/canonical signing
  bytes; offline CLI verification; descriptor, CLI, historical harness, and
  operator-runbook regressions; complete relevant local verification and
  independent read-only review.
- **Non-goals:** no app-hash-to-export cryptographic proof claim, no production
  ceremony, no deployment, no generic Wasm migration, and no remediation of
  the separate resource-amplification candidate in this task.
- **Acceptance:** the signed descriptor contains one exact 32-byte SHA-256
  commitment to the raw export; malformed/missing/mutated commitments fail
  closed; the CLI hashes the bytes it actually transforms; the real process
  harness and documentation use the same artifact; full relevant gates pass.
- **Risk:** High — changes canonical recovery authorization and genesis
  migration semantics. Codex Sol owns design, implementation, integration, and
  full verification; Kimi K3 performs an independent secret-free read-only
  final review.
- **Current evidence:** the pre-fix disposable CLI reproduction accepted a
  post-signing governance-state mutation and re-imported it successfully,
  confirming the missing artifact binding. Repository files were not modified
  by that reproduction.
- **Next:** add the signed export digest and fail-closed CLI check, run focused
  regressions, review the full diff, then execute the complete relevant gate.

---

## 2026-07-28 20:49 EEST GH-61 exact export-artifact binding → Done locally

- **Outcome:** GH61-SB-001 is fixed locally. Descriptor v1 now contains
  `source_genesis_sha256`; the canonical bytes signed by every fresh operator
  bind the exact raw export that the transform consumes. Missing, malformed,
  or changed commitments fail before JSON parsing, state rewrite, ledger
  validation, or output creation.
- **Changed for this bounded task:** project-local `AGENTS.md`;
  `migration/legacy_authority.go`, `migration/transform_core.go`, and related
  descriptor/core/application tests; CLI and historical harness fixtures;
  `docs/node-operators/operations/legacy-authority-migration.md`; Bridge,
  Action Log, Security Notes, Project State, and TODO.
- **Regression depth:** the CLI mutation regression proves that changing the
  signed export leaves no output. Other raw-mutating negative fixtures recompute
  and re-sign their exact digest so source-chain, height, consensus, Auth, Bank,
  DEX, Wasm, stale-reference, and malformed-JSON gates remain independently
  exercised.
- **Kimi contribution:** independent secret-free read-only review; write guard
  before/after remained
  `8beb30ec6fe6d22368bf7587190dfe31f06a47a69ceeeb2690ef528df18620b5`.
  Kimi reported no P0/P1/P2 issue. Sol fixed the one accepted P3 test-
  attribution observation and retained ownership of design, diff review, and
  all final gates.
- **Verification:** focused migration/root suite PASS (`migration` 3.318s,
  root 6.837s); `make verify` PASS, including selector, build, vet, and
  Race/Coverage (root 193.441s/69.4%, migration 18.935s/82.5%, token
  5.397s/92.6%, treasury 9.157s/97.0%, DEX 23.969s/45.3%,
  truedemocracy 172.168s/62.2%); docs consistency PASS; `gofmt -l` empty;
  `git diff --check` PASS.
- **Real process evidence:**
  `TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test . -run
  '^TestMultiValidatorLegacyAuthorityMigrationRollback$' -count=1
  -timeout=900s -v` PASS in 174.82s (package 178.949s), preserving the exact
  four-validator migration, target convergence, export/re-import, signer
  isolation, and untouched-source rollback evidence.
- **Residual risks:** GH61-SB-002 resource amplification and the P3
  canonical-string limitation for preserved unknown-module address scans are
  documented but intentionally not changed. Nonempty Wasm, generic in-place
  migration, public ceremony, and app-hash-to-export proof remain unsupported.
- **Blockers:** none for this authorized local remediation. Publication,
  GitHub PR/CI/review, merge, issue closure, production actions, and any
  follow-up hardening require their own current task authorization under the
  refreshed workflow; none were performed.
- **Next authorized step:** none. Target Stop is active.

---

## 2026-07-29 02:02 EEST GH-61 publication and closure → In Progress

- **Branch:** `feature/GH-61-legacy-authority-migration`
- **Issue:** [GH-61](https://github.com/NeaBouli/TrueRepublic/issues/61)
- **Authorization:** Gio approved the immediately proposed publication stage
  with “Ok dann weitermachen”: commit and push the already verified GH-61
  scope, open a draft PR, inspect final-head GitHub CI and review, remediate
  valid findings, merge only when green, and synchronize issue/roadmap/Bridge.
- **Definition of Ready:** branch HEAD and `origin/main` both equal
  `8c0c55574389210162db59c2c1103bb2d99e2691`; authenticated `gh` targets
  `NeaBouli/TrueRepublic`; the complete dirty tree is the documented GH-61
  implementation with no unrelated user changes.
- **Acceptance:** intentional commit contains only GH-61; remote branch and PR
  target `main`; all required final-head checks and actionable review are
  resolved; merge and GH-61/GH-29/project status are synchronized with exact
  evidence.
- **Risk:** High — publishes and merges consensus/genesis migration code.
  Codex Sol owns staging, external writes, review, CI inspection, remediation,
  merge, and closure.
- **Non-goals:** no deployment, public-network launch, mainnet or live
  migration ceremony, server/IAM/secret/wallet action, force-push, GH61-SB-002
  implementation, or unrelated follow-up.
- **Current evidence:** local focused tests, `make verify`, docs consistency,
  formatting/diff checks, exact-binding Kimi review, and the historical
  four-validator migration/rollback drill are green.
- **Next:** stage the explicit verified file set, review the staged diff,
  commit, push, and open the draft PR.

---

## 2026-07-29 02:06 EEST GH-61 PR #69 → Draft / CI

- **Commit:** `432156fe9090b1f2711461d0a135715b572fd8bb`
  (`feat(migration): recover legacy validator authorities`), one intentional
  26-file GH-61 commit; staged secret-pattern and diff checks were clean.
- **Remote:** `origin/feature/GH-61-legacy-authority-migration` tracks the
  pushed branch; no force-push was used.
- **Pull request:** [PR #69](https://github.com/NeaBouli/TrueRepublic/pull/69)
  is a draft against exact base
  `8c0c55574389210162db59c2c1103bb2d99e2691` on `main`, links GH-61, and
  records implementation, root cause, local evidence, security limits, and
  unsupported production boundaries.
- **Initial GitHub state:** mergeable; Go, Docker, docs, and security jobs
  started. DeepScan is green. CodeRabbit correctly skipped the draft and will
  be triggered after required CI is green and the PR becomes ready for review.
- **Next:** commit this PR pointer, observe only the new final head, inspect
  every failed check/review thread, and do not merge until all required
  evidence is green and actionable review is resolved.

---

## 2026-07-29 02:09 EEST PR #69 go-vuln → Changes Required

- **Failed check:** final-head Security Scan `go-vuln` on
  `59884edde7637d9bce315618c2b0e5b7bcc2cb48`, run
  [30361942572](https://github.com/NeaBouli/TrueRepublic/actions/runs/30361942572).
- **Root cause:** the current vulnerability database added reachable
  GO-2026-6061 in `google.golang.org/grpc@v1.79.3`; the scanner reports
  `v1.82.1` as the fixed version. Four other reachable upstream findings still
  report `Fixed in: N/A` and remain tracked.
- **Approved remediation scope:** update only gRPC to the scanner-named fixed
  version plus unavoidable module metadata, review the dependency diff, rerun
  the exact vulnerability gate, `make verify`, and the complete
  multi-validator process matrix because gRPC is network-transport critical.
- **Non-goals:** no GH61-SB-002 work, application/API redesign, unrelated
  dependency modernization, production action, or merge before the new final
  head is green.
- **Next:** inspect the module graph, apply the minimal version update, and run
  the complete local gate before pushing.

---

## 2026-07-29 02:29 EEST PR #69 GO-2026-6061 remediation → Local PASS

- **Dependency fix:** `google.golang.org/grpc` upgraded from `v1.79.3` to the
  scanner-named fixed `v1.82.1`. `go mod tidy` updated only the required
  gRPC/xDS/Genproto/Telemetry support graph and promoted already directly
  imported `cosmossdk.io/core`; no TrueRepublic source/API/consensus behavior
  was changed.
- **Security evidence:** current `govulncheck ./...` no longer reports
  GO-2026-6061 and reports no reachable vulnerability with an available fix.
  The remaining four reachable findings are unchanged and all state
  `Fixed in: N/A`.
- **Module/docs gates:** `go mod verify`, docs consistency, `git diff --check`,
  repository package selection, build, and vet PASS.
- **Race/Coverage:** `make verify` PASS: root 65.266s/69.4%, migration
  4.081s/82.5%, token 4.007s/92.6%, treasury 5.337s/97.0%, DEX
  5.639s/45.3%, truedemocracy 53.976s/62.2%.
- **Full network/process matrix:** all eight gated scenarios PASS in 940.823s:
  legacy migration/rollback 124.35s, consensus recovery 72.32s, trusted
  state-sync 125.42s, backup/restore/export/import 79.44s, persisted binary
  upgrade/rollback 208.71s, cold failover 59.63s, consensus-key rotation
  92.73s, and consensus slashing 175.71s.
- **Next:** commit and push the exact module/Bridge remediation, update PR #69,
  and evaluate only the new final-head GitHub runs.

---

## 2026-07-29 03:08 EEST PR #69 Final Head → CI PASS / Review Recovery

- **Final head:** `a28f5654390f58feecfd86afeeb4d7f74e204341`
  (`fix(deps): update grpc for GO-2026-6061`) is pushed without force.
- **GitHub evidence:** docs consistency, Go build/race/coverage,
  multi-validator recovery, Docker restart, Go/Rust/maintained-client and
  informational legacy-client security scans, and DeepScan all PASS. DeepScan
  reports zero new issues.
- **Review state:** PR #69 is ready for review and mergeable. CodeRabbit's
  status remained `Review in progress` on an unchanged timestamp for about ten
  hours, with no review or inline thread emitted. A bounded
  `@coderabbitai review` command was acknowledged but reused the stale run.
  Converting the PR to draft and immediately back to ready also preserved the
  stale status.
- **Recovery action:** this append-only coordination update will create a
  fresh final head so GitHub and CodeRabbit receive a new commit event without
  changing migration, consensus, API, or dependency behavior.
- **Safety boundary:** no review requirement is bypassed and no merge,
  deployment, public-network, secret, wallet, or production action occurs
  while the independent review check is pending.
- **Next:** commit/push this documentation-only recovery head, evaluate every
  check and review thread on that exact head, then merge only after required
  final-head evidence is terminal and acceptable.

---

## 2026-07-29 03:14 EEST PR #69 CodeRabbit → Changes Required

- **Recovered review:** CodeRabbit completed its full review of implementation
  head `a28f5654390f58feecfd86afeeb4d7f74e204341`. The apparent long-running
  status was an external completion-display delay; the fresh Bridge commit
  exposed the finished review and produced a new incremental event.
- **Actionable inline threads (6):** portable repository instructions; halt
  guidance without validator-set mutation; atomic/fail-closed source export;
  reject empty Comet validator state when active application validators exist;
  replace deprecated `authtypes.ModuleAccountI`; and make the supply snapshot
  independent instead of slice-aliased.
- **Review-body findings (7):** re-evaluate the eight-harness timeout margin;
  check every legacy operator in the process drill; capture stderr for
  stdout-sensitive subprocesses; document the ephemeral test-only key copy;
  cover foreign account-prefix rejection; decode mappings once before sorting;
  and consolidate duplicated consensus-key traversal.
- **Assessment:** all findings are within GH-61 review-remediation scope. The
  docstring-coverage warning is repository-wide/informational and is not a
  migration correctness gate.
- **Kimi assignment:** Kimi K3 receives the bounded, secret-free Go/test/CI
  remediation block. Sol owns the runbook/instruction fixes, integration,
  security review, full tests, GitHub replies/resolution, final-head CI, and
  merge. Delegated agents may not perform external writes or start agents.
- **Review service:** the subsequent documentation-only incremental review is
  temporarily rate-limited; the completed full implementation review and all
  its findings remain available and must be resolved before merge.
- **Next:** implement and review every valid finding, run focused and full
  local gates, commit/push one reviewed remediation, resolve threads with
  evidence, and wait for terminal final-head GitHub checks.

---

## 2026-07-29 03:33 EEST PR #69 Review Remediation → Integration Correction

- **Kimi contribution:** Kimi K3 implemented the bounded Go/test/CI review
  block without external writes. Focused migration tests and root/migration vet
  passed. Sol reviewed every delegated diff and corrected the test-only key-copy
  wording before integration.
- **Accepted fixes:** repository-path portability; non-mutating halt guidance;
  atomic private source export; independent bank-supply snapshot; current SDK
  module-account interface; all-operator process assertion; stderr diagnostics;
  explicit ephemeral-key warning; foreign-prefix and malformed-key regressions;
  single five-collection consensus-key traversal; decode-once mapping sort; and
  coherent 1500-second test / 30-minute CI timeout budgets.
- **Focused/full unit evidence:** docs consistency, diff check, migration tests,
  root migration-command tests, vet, and `make verify` PASS. Race/Coverage:
  root 69.4%, migration 85.0%, token 92.6%, treasury 97.0%, DEX 45.3%, and
  truedemocracy 62.2%.
- **Integration finding:** the real historical process drill correctly proved
  that a halted Cosmos SDK running-chain export omits
  `consensus.validators` while retaining four active truedemocracy validators.
  The review suggestion to reject that canonical shape caused the transform to
  fail and was therefore invalid for GH-61.
- **Correction:** restored acceptance of an empty exported Comet validator list
  after validating the consensus envelope; retained strict reconciliation when
  a list is present; added a regression and runbook/code explanation that
  truedemocracy supplies target InitChain validator updates. The interrupted
  first matrix run is recorded as FAIL evidence, not hidden.
- **Next:** rerun focused gates, then the complete eight-scenario process matrix
  from a clean test invocation before any commit or push.

---

## 2026-07-29 03:53 EEST PR #69 Review Remediation → Full Matrix PASS

- **Corrected focused gates:** migration tests PASS in 2.231s; documentation
  consistency and `git diff --check` PASS.
- **Historical migration/rollback:** PASS in 196.40s with the canonical empty
  exported Comet validator list, four active application validators, exact
  transform, target convergence, export/reimport, and untouched source rollback.
- **Complete real process matrix:** 8/8 PASS in 1169.685s: consensus recovery
  119.13s, trusted state sync 145.57s, backup/restore/export/import 90.99s,
  persisted binary upgrade/rollback 223.52s, cold failover 87.03s,
  consensus-key rotation 88.48s, and consensus slashing 215.28s in addition to
  the GH-61 drill.
- **Timeout evidence:** the prior 1200-second Go test limit would have left only
  about 30 seconds of local margin. The reviewed 1500-second test cap and
  30-minute job cap are therefore required, coherent, and not a convenience
  relaxation.
- **Review disposition:** the empty-consensus rejection is rejected with direct
  integration evidence; all other CodeRabbit findings are remediated. No
  review thread is resolved until the final diff and final local gate are
  committed and pushed.
- **Next:** rerun `make verify` on the final code, inspect/stage the exact diff,
  commit/push, respond to and resolve review threads with evidence, then require
  terminal final-head GitHub CI before merge.

---

## 2026-07-29 03:58 EEST PR #69 Final Remediation Diff → Local PASS

- **Final `make verify`:** repository selector, build, vet, and Race/Coverage
  PASS: root 109.768s/69.4%, migration 9.775s/84.6%, token 5.440s/92.6%,
  treasury 3.564s/97.0%, DEX 11.708s/45.3%, and truedemocracy
  95.818s/62.2%.
- **Complete evidence:** final focused tests/docs plus 8/8 real process
  scenarios are green on the corrected review-remediation worktree.
- **Next:** stage only the documented review-remediation files, inspect the
  staged diff and secret patterns, commit/push without force, then resolve PR
  feedback only with exact final-head evidence.

---

## 2026-07-29 04:18 EEST GH-61 → Merged / Closure Sync In Progress

- **Final PR head:** `33895dca627139d541e74236afb63d2b9fb5bff2`;
  all 11 GitHub checks PASS, including the complete multi-validator recovery
  matrix in 13m38s, Go build/race/coverage, Docker, docs, Go/Rust/Node security,
  DeepScan, and CodeRabbit.
- **Review:** seven original review clusters plus the final timezone false
  positive are answered with evidence; zero unresolved threads remain.
- **Merge:** PR #69 squash-merged to `main` as
  `264ab7c817414e3081149c61d8b5b6fb0ce5e368`. GH-61 is closed.
- **Closure branch:** `docs/GH-61-closure` starts from the exact clean merge
  commit. Scope is documentation/ticket synchronization only.
- **Next:** update Project State, TODO, Road to Rollout, Limitations, Action Log,
  Security Notes, GH-29, and the global Bridge; run docs/diff checks; publish and
  merge the closure PR only on green final-head evidence.

---

## 2026-07-29 04:24 EEST GH-61 Closure → Public Count Reconciled

- Reproduced the merged repository's package-case count with normal
  `go test -json`: root 62, migration 82, token 12, treasury 36, DEX 116, and
  truedemocracy 511, totaling 819 Go cases.
- The public authoritative total is therefore 853: 819 Go + 26 Rust + eight
  maintained-client cases. The prior 733/699 values were the pre-GH-61 GH-60
  baseline and were stale after PR #69.
- Updated status JSON, README badge/status, agent guide, GitHub Pages source,
  wiki home/current/testing status, module arithmetic, and final coverage
  values. Process harnesses remain separately gated and are not double-counted.
- Next: run the exact docs-consistency and diff gates, review/stage the closure
  diff, then publish the documentation-only closure PR.

---

## 2026-07-29 04:35 EEST PR #70 Independent Closure Review → Remediated

- **Review:** Kimi K3 completed a bounded, read-only independent review of the
  documentation-only closure diff and changed no repository files. It verified
  PR #69/GH-61/GH-29 state, all 11 final-head checks, zero unresolved review
  threads, the 853 = 819 + 26 + 8 arithmetic, per-module Go counts, coverage,
  scope, and retained non-production boundaries.
- **Findings:** no P0/P1, code, security, or production-claim findings. The only
  material P2 was a stale 733/699 baseline in `docs/ROLLOUT_ROADMAP.md`; minor
  P3s were its old update date and the wiki harness list omitting GH-61.
- **Remediation:** corrected the roadmap to 853/819 and 2026-07-29, clarified
  that separately gated process harnesses are never double-counted, and added
  GH-61 legacy-authority migration/rollback to the wiki harness inventory.
- **Review fallback:** CodeRabbit retained its successful draft-skip result
  after PR #70 was marked ready and manually requested. Kimi supplies the
  independent closure review; Sol owns final diff review and all final-head
  gates.
- **Next:** run documentation consistency, arithmetic, stale-value, and diff
  checks; commit/push the remediation; require green final-head PR #70 checks
  and zero unresolved threads before merge.

---

## 2026-07-29 10:40 EEST GH-71 Role-Based Network Policy → In Progress

- **Branch:** `feature/GH-71-network-policy`
- **Issue:** [GH-71](https://github.com/NeaBouli/TrueRepublic/issues/71);
  parent rollout tracker
  [GH-29](https://github.com/NeaBouli/TrueRepublic/issues/29).
- **Scope:** define and validate fail-closed seed, sentry, validator, RPC/API,
  and private-operator profiles covering canonical peer endpoints, PEX and
  private/unconditional peers, listener exposure, unsafe RPC/CORS/profiling,
  least-privilege firewall ports, and reverse-proxy rate-limit requirements.
  Add deterministic repository tooling, regression/process evidence, and an
  operator runbook.
- **Risk:** High. Incorrect peer or listener policy can expose validator
  identities, privileged RPC, profiling, or denial-of-service surface.
- **Safety boundary:** repository code/tests/fixtures/docs only. No server
  access, real topology, firewall/DNS mutation, deployment, production,
  mainnet, credentials, secrets, keys, or public-network launch.
- **Current evidence:** exact clean `origin/main`
  `35827e48c733ba931601aa14cc4ea31d1a14ba4d`; GH-71 is the last unchecked
  Phase-1 network/disaster-recovery item in GH-29.
- **Delegation:** Kimi K3 will receive one bounded, secret-free
  implementation/analysis block. Sol retains architecture, security, complete
  diff review, tests, GitHub writes, and closure.
- **Next:** inspect the node configuration lifecycle and existing Docker/
  operator documentation; freeze the typed profile schema and test plan before
  implementation.

---

## 2026-07-29 11:20 EEST GH-71 Implementation Milestone

- **Status:** In Progress.
- **Kimi K3 contribution:** implemented the bounded `networkpolicy/`
  validation core, CLI registration, and focused unit/CLI tests; no external
  writes or delegated agents.
- **Sol integration/review:** removed any operator-accessible local-peer
  bypass, made validation output path/content-safe, corrected validator/sentry
  trust relationships, added external-address and Prometheus checks, tightened
  the RPC role, and integrated fail-closed startup/Docker/nginx defaults.
- **Documentation:** role profiles, listener boundaries, PEX/topology rules,
  reverse-proxy guidance, firewall/rate-limit guidance, and local-only
  monitoring exposure are being synchronized across operator docs and wiki.
- **Verified so far:** focused network-policy tests PASS; root startup/init/
  repository policy tests PASS; `go vet` for root and networkpolicy PASS; shell
  syntax PASS; YAML parse PASS; `git diff --check` PASS.
- **Environment limitation:** Docker is not installed locally, so the
  repository Docker/Compose smoke test remains delegated to GitHub Actions
  rather than being claimed locally.
- **Risks/blockers:** no scope blocker. Full repository gates, eight-scenario
  process matrix, independent Kimi review, GitHub CI, and final closure sync
  remain pending.
- **Next:** complete the documentation safety scan, run full local verification
  and process evidence, then review/publish/merge only on green evidence.

---

## 2026-07-29 12:24 EEST GH-71 Local Gates and Review → Green

- **Status:** In Progress; ready for publication and final-head GitHub gates.
- **Implementation:** five deterministic node roles, canonical peer/topology
  validation, validator-initiated sentry isolation, loopback-only client/
  profiling/metrics listeners, fail-closed native startup, safe local Docker/
  nginx/monitoring defaults, CI Compose runtime evidence, and operator runbook.
- **Kimi K3:** bounded implementation contribution plus a later independent
  read-only security/integration review. No P0/P1 or review blocker; no files
  changed by the review. Sol remediated every remaining P3 observation.
- **Final local Go gate:** `make verify` PASS. Coverage: root 69.5%, migration
  84.6%, network policy 95.5%, token 92.6%, treasury 97.0%, DEX 45.3%,
  governance 62.2%.
- **Counts:** 952 Go cases (69 + 82 + 126 + 12 + 36 + 116 + 511), plus 26
  Rust and eight maintained-client cases = 986 authoritative branch cases.
- **Process evidence:** seven scenarios PASS in the combined run; concurrent
  review/test load exhausted the 1500-second global limit during slashing.
  The isolated slashing rerun PASS in 303.98s. The exact final-head combined
  matrix remains mandatory on GitHub.
- **Other gates:** focused tests, shell syntax, YAML/JSON parsing, gofmt, docs
  consistency, and `git diff --check` PASS.
- **Docker boundary:** Docker unavailable locally; CI now starts node, nginx,
  wallet, Prometheus, and Grafana and requires proxy RPC, wallet routing,
  healthy node metrics, and Grafana health.
- **Audit:** `docs/agent-bridge/GH71_AUDIT.md`.
- **No production action:** no server, firewall, DNS, deployment, real
  topology, key, credential, mainnet, or public-network change.
- **Next:** commit/push, open PR, require exact final-head CI/review, remediate
  any valid finding, merge, and synchronize GH-71/GH-29/public closure state.

---

## 2026-07-29 12:28 EEST GH-71 Published as Draft PR #72

- **Commit/head:** `3e0a9f521e9a7d3c623d6417922f0b63559c3028`
  on `feature/GH-71-network-policy`.
- **PR:** [#72](https://github.com/NeaBouli/TrueRepublic/pull/72), draft
  against `main`; `Closes #71` is declared only for merge.
- **State:** exact final-head GitHub CI and external review are running/pending.
  No merge readiness is claimed.
- **Hard gates:** complete eight-scenario process matrix, Go
  build/vet/Race/coverage, full Docker/Compose runtime smoke, docs/static/
  security checks, review, and zero unresolved actionable threads.
- **Next:** push this append-only publication record, monitor the new exact
  head, remediate valid failures/findings, then mark ready/merge only when all
  gates are green.

---

## 2026-07-29 12:52 EEST GH-71 → Done / Target Stop

- **Status:** Done. [PR #72](https://github.com/NeaBouli/TrueRepublic/pull/72)
  squash-merged to `main` as
  `9c369ac0d589f749e055af33a03f5f4981020101`; GH-71 closed automatically.
- **Exact verified head:** `83ffaad86274efdf3d2bb54dfff6ce8a625d5dee`.
- **GitHub evidence:** build/race/coverage PASS 6m03s; complete recovery matrix
  PASS 14m29s; Docker/Compose restart, proxy, wallet, metrics, and Grafana smoke
  PASS 7m04s; docs, Go/Rust/Node security scans, and DeepScan PASS.
- **Review:** independent Kimi K3 review found no P0/P1 and Sol remediated all
  lower-severity observations before final CI. CodeRabbit remained stuck for
  about ten hours without a review, finding, or thread; the infrastructure
  condition was documented on the PR. Zero unresolved review threads existed.
- **Merge decision:** owner-authorized admin squash merge bypassed only the
  stale external-bot status; no failed technical gate was bypassed.
- **Safety boundary:** no server, firewall, DNS, deployment, real topology,
  secret, key, credential, production, mainnet, or public-network action.
- **Remaining parent work:** GH-29 stays open for later rollout phases.
- **Next:** publish this documentation-only closure synchronization, require
  its final-head docs/static gates, merge, update GH-29, and stop GH-71.

---

## 2026-07-29 17:22 EEST GH-74 Node Health Signals → In Progress

- **Branch:** `feature/GH-74-node-health`
- **Issue:** [GH-74](https://github.com/NeaBouli/TrueRepublic/issues/74);
  parent [GH-29](https://github.com/NeaBouli/TrueRepublic/issues/29).
- **Scope:** add distinct dependency-free local `live` and `ready` probes,
  strict timeout/response handling, Docker and real-container CI integration,
  focused/process regressions, and operator guidance.
- **Semantics:** liveness proves only that local CometBFT RPC responds;
  readiness additionally requires valid status, positive block height, and
  `catching_up=false`. Synchronization must not become a restart condition.
- **Risk:** Medium. Incorrect probe semantics can cause restart loops or route
  traffic to an unsynchronized node; consensus and transaction behavior remain
  unchanged.
- **Safety boundary:** repository code/tests/docs and GitHub CI only. No
  production, server, firewall, DNS, deployment, real topology, key, secret,
  credential, mainnet, or public-network action.
- **Baseline:** exact clean `origin/main`
  `71164459eecfff12de8671d8d3220cc0c3fb1181`; no open PR at task start.
- **Delegation:** Kimi K3 receives one bounded, secret-free implementation
  block for the probe package/CLI/focused tests. Sol retains architecture,
  security, Docker/CI/docs integration, complete diff review, full tests,
  GitHub writes, and closure.
- **Next:** freeze the probe API and failure taxonomy, then run the bounded
  implementation while Sol independently designs integration and policy gates.

---

## 2026-07-29 17:45 EEST GH-74 Local Verification Milestone

- **Implementation:** `truerepublicd healthcheck live|ready` now separates
  restart-worthy local RPC failure from synchronization-aware readiness.
- **Security:** probe targets are literal loopback plain HTTP only; userinfo,
  query, fragment, non-root paths, redirects, environment proxies, oversized
  bodies, invalid JSON-RPC, and unbounded timeouts fail closed with safe errors.
- **Integration:** Docker uses liveness only. GitHub container/Compose gates
  require live, ready, and post-restart height progress. Operator guidance uses
  separate Kubernetes exec probes.
- **Kimi contribution:** bounded probe package, CLI, and focused tests. Sol
  reviewed every write and corrected redirect/proxy handling, strict JSON-RPC
  validation, test portability, and race-safety.
- **Tests:** focused health race tests PASS twice; focused root/config tests
  PASS; both focused and root vet PASS; `make verify` PASS. Authoritative local
  count is 1,009 Go plus 26 Rust and eight maintained-client cases = 1,043.
- **Coverage:** root 69.5%, healthcheck 97.2%, migration 84.6%, network policy
  95.5%, token 92.6%, treasury 97.0%, DEX 45.3%, governance 62.2%.
- **Audit:** `docs/agent-bridge/GH74_AUDIT.md` records 0 FAIL / 1 WARN /
  12 PASS.
- **Independent review:** Kimi K3 found no P0/P1/P2. Sol remediated the one
  actionable P3 by refusing to publish or compare an empty separately fetched
  block height; duplicated timeout wording and the standard released-port test
  idiom remain non-blocking P3 observations.
- **Remaining gate:** Docker is unavailable locally; exact final-head GitHub
  Docker/Compose evidence remains mandatory.
- **Safety boundary:** no production, server, firewall, DNS, deployment, real
  topology, secret, credential, key, mainnet, or public-network action.
- **Next:** freeze the local diff, complete the independent review and final
  local gates, then publish a draft PR for exact final-head CI.

---

## 2026-07-29 18:22 EEST GH-74 Published as Draft PR #75

- **Commit/head:** `6de3bb1311593fa143ebcfc3bcd603948d6ffcc9`
  on `feature/GH-74-node-health`.
- **PR:** [#75](https://github.com/NeaBouli/TrueRepublic/pull/75), draft
  against `main`; `Closes #74` applies only when merged.
- **Independent review:** 0 P0/P1/P2. The one actionable P3 was remediated and
  the exact follow-up diff was independently approved; no review-authoring
  agent changed repository files.
- **Local final tree:** `make verify`, focused health Race, root/config/policy,
  vet, docs consistency, shell syntax, JSON/YAML parse, formatting, and diff
  gates PASS.
- **State:** final-head GitHub CI/security/static/external review and real
  Docker/Compose runtime evidence are now pending. No merge readiness is
  claimed.
- **First published-head evidence:** build/race/coverage PASS 7m25s;
  Docker/Compose PASS 7m27s; complete recovery matrix PASS 13m58s; docs,
  Go/Rust/Node security, and DeepScan PASS.
- **External review:** CodeRabbit opened six actionable threads. Sol accepted
  all six for minimal remediation: preserve the key-compromise runbook gate,
  correct migration wording, make the deployment curl fail closed, correct
  Docker unhealthy/restart semantics, use context-aware ephemeral listening in
  the test, and match documentation totals only as complete numbers.
- **Next:** run focused local verification, commit/push the six review
  remediations, require refreshed exact-head CI/review, resolve every confirmed
  thread, and merge only with zero unresolved actionable findings.

---

## 2026-07-29 19:02 EEST GH-74 → Done / Target Stop

- **Status:** Done. [PR #75](https://github.com/NeaBouli/TrueRepublic/pull/75)
  squash-merged to `main` as
  `468a43a6cf4a3ed6cc76524a2de212b4fcf7052a`; GH-74 closed automatically.
- **Exact verified head:** `e02104452fe2ddd201045c0ea0bb9651c053c417`.
- **GitHub evidence:** build/vet/race/coverage PASS 6m39s; complete recovery
  matrix PASS 13m54s; Docker/Compose live, ready, persistent restart, proxy,
  wallet, metrics, and Grafana PASS 7m27s; docs, Go/Rust/Node security, and
  DeepScan PASS.
- **Review:** independent Kimi review found no P0/P1/P2. All six CodeRabbit
  findings were remediated, answered, and resolved; incremental follow-up was
  rate-limited but passed with zero new unresolved threads. The independent
  documentation-only closure review also approved with no P0/P1/P2; its
  audit-count traceability P3 was remediated by enumerating all 13 PASS items.
- **Merge decision:** owner-authorized admin squash merge bypassed only the
  formal self-approval requirement. No failed technical gate or unresolved
  actionable finding was bypassed.
- **Semantic correction:** Docker health status uses liveness, but
  `restart: unless-stopped` does not restart an unhealthy process without
  explicit external automation. Readiness remains traffic-routing only.
- **Safety boundary:** no production, server, firewall, DNS, deployment, real
  topology, secret, key, credential, mainnet, or public-network action.
- **Remaining parent work:** GH-29 stays open for later rollout phases.
- **Next:** publish this documentation-only closure synchronization, require
  its final-head docs/static checks, merge, update GH-29, and stop GH-74.

---

## 2026-07-30 05:28 EEST GH-77 Secret-Safe Structured Logging → In Progress

- **Issue/branch:** [GH-77](https://github.com/NeaBouli/TrueRepublic/issues/77),
  `feature/GH-77-structured-logging`; parent rollout tracker
  [GH-29](https://github.com/NeaBouli/TrueRepublic/issues/29).
- **Scope:** make the supported daemon-operation path emit structured JSON and
  sanitize SDK/CometBFT messages and fields at the central logger boundary.
  Preserve safe operational metadata while redacting credentials, mnemonic and
  private-key material, raw transactions, proofs, signatures, and other
  private transaction data.
- **Planned evidence:** focused unit and real-process redaction/structure
  regressions, full Go build/vet/race/coverage, repository consistency,
  recovery matrix, Docker/Compose, docs/security gates, and independent Kimi
  review on the exact final head.
- **Safety boundary:** repository code, tests, and documentation only. No
  consensus/state change, production-log inspection, collector deployment,
  server, credential, key, mainnet, or public-network action.
- **Owner:** Codex Sol; Kimi K3 receives one bounded secret-free
  implementation/review block. Sol retains architecture, security, integration,
  GitHub writes, complete verification, and merge responsibility.
- **Next:** map the SDK logger construction and sensitive field surface,
  delegate the bounded logging core, review the full diff, and run focused
  gates before broader integration.

---

## 2026-07-30 06:17 EEST GH-77 Local Implementation Review → In Progress

- **Implementation:** the supported `start` path now enforces JSON at the
  SDK/CometBFT logger boundary, wraps every logger level and child context with
  bounded defensive sanitization, rejects raw `--trace-store` output, and keeps
  health/network-policy commands config-independent. Script, image, Compose,
  CI policy, regression tests, operator guidance, and the 0 FAIL / 2 WARN /
  14 PASS audit were updated in the same bounded scope.
- **Kimi contribution:** the initial bounded implementation attempt produced
  design analysis but no file diff and was stopped. Sol implemented and
  integrated the design. A fresh read-only Kimi K3 adversarial review of the
  resulting tree found no P0, P1, or must-fix P2; its P3 residuals are
  fail-safe over-redaction, heuristic-not-DLP limits, caller-side concurrent
  map mutation, and documented unstructured Go crash output.
- **Focused evidence:** `go vet ./observability .` PASS;
  `go test ./observability -count=1` PASS; `go test -race ./observability
  -count=1` PASS; focused start/script/network policy tests PASS; real compiled
  daemon start/stop/restart PASS in 59.78s with CLI/environment plain formats
  and CLI/environment trace stores rejected, height advancement preserved, and
  every captured normal-operation line valid JSON with string `level` and
  `message`.
- **Risk/blocker:** no code blocker. Docker is not installed locally, so the
  exact image/Compose runtime remains a mandatory GitHub CI gate. Redaction is
  defensive minimization, not a substitute for forbidding secrets at call
  sites or treating retained logs as sensitive.
- **Safety boundary:** no production log, collector, server, deployment,
  credential, key, mainnet, consensus/state, or public-network action.
- **Next:** run the repository's complete local verification and recovery
  matrix, publish the reviewed head, require every exact-head GitHub gate and
  actionable review thread to pass, then merge and synchronize GH-29/docs.

---

## 2026-07-30 06:42 EEST GH-77 Complete Local Gates → Ready for PR

- **Go:** `make verify` PASS across the authoritative nine-package selector:
  build, vet, race, and coverage all green; root 68.9%, observability 77.3%,
  healthcheck 97.2%, migration 84.6%, network policy 95.5%, token 92.6%,
  treasury keeper 97.0%, DEX 45.3%, and truedemocracy 62.2%.
- **Recovery:** all eight multi-validator scenarios PASS in 1161.044s:
  migration rollback, consensus recovery, trusted snapshot state sync,
  backup/restore/export/import, persisted binary upgrade rollback, validator
  cold failover, key rotation, and slashing.
- **Repository/integration:** documentation consistency, `git diff --check`,
  shell syntax, workflow/Compose YAML parsing, Rust format/clippy/build/tests,
  Cargo audit, and active web-client lint/8 tests/build/high-severity audit
  PASS. Cargo audit reported six allowed transitive maintenance/unsoundness
  warnings without a failing advisory.
- **Existing out-of-scope debt:** the legacy mobile audit reports 51 dependency
  findings (7 low, 16 moderate, 24 high, 4 critical); its official workflow is
  explicitly informational and the unchanged legacy client is outside GH-77.
  `govulncheck` is not installed locally; the exact-head GitHub Security Scan
  remains mandatory and fails on reachable fixable Go vulnerabilities.
- **Blocker:** none for publication. Docker is not installed locally, so image,
  Compose, and post-restart two-stream JSONL evidence must pass on GitHub.
- **Safety boundary:** no production log, collector, server, deployment,
  credential, key, mainnet, consensus/state, or public-network action.
- **Next:** freeze and publish this reviewed tree, open the GH-77 PR, require all
  exact-head CI/security/review gates, remediate findings, then merge.

---

## 2026-07-30 06:44 EEST GH-77 Published → PR #78 In Review

- **Published implementation:** commit `ed4fb0700d415990eb79dd8e6f1c8b0104327165`
  is on `origin/feature/GH-77-structured-logging`; draft
  [PR #78](https://github.com/NeaBouli/TrueRepublic/pull/78) targets `main` and
  closes GH-77.
- **Initial GitHub state:** PR is mergeable; Docs Consistency and the recovery
  matrix started, remaining Go/Security/Docker jobs queued or running.
  CodeRabbit and DeepScan initially report success; exact final-head evidence
  is still required after this append-only coordination update.
- **Local evidence:** all previously recorded complete gates remain green; this
  update changes coordination documentation only.
- **Risk/blocker:** no implementation blocker. Do not merge until the final
  published head has green build/race/coverage, recovery, Docker/Compose
  structured-log, docs, static/security, and review evidence with zero
  unresolved actionable thread.
- **Next:** publish this coordination-only follow-up, monitor exact-head checks,
  address findings, mark ready, and merge only after every technical gate is
  green.

---

## 2026-07-30 06:54 EEST GH-77 Review Remediation → In Progress

- **Exact reviewed head:** `b97c527cfc9ccfcf1be3ea85c6e214e3d1643a69`;
  10 checks green and only the recovery matrix still running when remediation
  began. Build/race/coverage passed in 7m00s and Docker/Compose, including
  post-restart structured JSONL, passed in 7m09s.
- **CodeRabbit:** two unresolved threads. The valid minor finding is that
  `json.Number` implements `fmt.Stringer`, making the numeric-preservation case
  unreachable; reorder the cases and add a numeric-type regression. The
  Bridge/audit “future date” finding is invalid: every entry explicitly uses
  local `EEST` (`UTC+03:00`), and at review time the verified local clock was
  `2026-07-30 06:54:04 EEST` while GitHub displayed July 29 UTC.
- **Scope:** minimal sanitizer/test correction plus append-only evidence.
  Re-run focused and authoritative local gates, publish a new exact head, reply
  to both threads with evidence, and require refreshed CI/review.
- **Safety boundary:** unchanged; no production, server, collector, deployment,
  credential, key, mainnet, consensus/state, or public-network action.

---

## 2026-07-30 06:59 EEST GH-77 Review Remediation → Locally Verified

- **Correction:** `json.Number` now precedes `fmt.Stringer` at the sanitizer
  type boundary. `TestSanitizeValuePreservesJSONNumber` proves the reviewed
  value reaches the underlying SDK logger without being converted by the
  sanitizer. The SDK logger itself serializes `json.Number` as a string, so the
  regression intentionally does not overclaim end-to-end numeric JSON typing.
- **Verification:** focused observability normal/race tests and vet PASS;
  repository consistency and diff checks PASS; refreshed `make verify` PASS
  across all nine packages with observability coverage increased to 77.8%.
- **Review handling:** publish the minimal correction, respond to the valid
  CodeRabbit thread with the test evidence, answer the EEST/UTC false positive
  with the verified local timestamp, and resolve both only after the new head
  is visible.
- **Next:** push the remediation head and require a complete refreshed
  exact-head CI/review cycle before merge.

---

## 2026-07-30 07:16 EEST GH-77 Implementation Merged → Closure Sync

- **Merge:** [PR #78](https://github.com/NeaBouli/TrueRepublic/pull/78)
  squash-merged to `main` as
  `133fb3be9b2a1c571d14b8b2847670d653c43f61`; GH-77 closed automatically.
- **Exact verified head:** `fd8378c3e58a273e69e4a2d752d3b1f41c4a80ac`.
  All 11 checks PASS: build/vet/race/coverage 7m16s, recovery matrix 14m10s,
  Docker/Compose with post-restart structured JSONL 6m57s, docs, Go/Rust/Node
  security, DeepScan, and CodeRabbit.
- **Review:** both CodeRabbit threads were answered and resolved; the bot
  confirmed the `json.Number` fix and withdrew its UTC/EEST false positive.
  Independent Kimi full and exact incremental reviews found no P0/P1/P2.
- **Merge decision:** owner-authorized admin squash bypassed only the formal
  self-approval requirement. No failed technical gate or unresolved actionable
  finding was bypassed.
- **Closure scope:** branch `docs/GH-77-closure` updates the Phase-6 roadmap,
  public status/limitations, exact test count, GH-29 checklist, TODO/ACTION_LOG,
  audit conclusion, and Bridge. No runtime code or production action.
- **Next:** calculate the merged authoritative test count, publish and verify
  the documentation-only closure PR, merge it, update GH-29, and stop GH-77.

---

## 2026-07-30 GH-77 Closure Documentation → Locally Verified

- **Authoritative count:** fresh package-scoped `go test -json` evidence records
  1,058 Go cases: root 79, healthcheck 55, migration 82, network policy 126,
  observability 41, token 12, treasury 36, DEX 116, and governance 511.
  Combined public total is 1,092 with 26 Rust and eight maintained-client cases.
- **Synchronization:** README/badge, CLAUDE guide, machine status/modules,
  landing page, rollout roadmap, limitations, wiki home/current/testing,
  PROJECT_STATE, SECURITY_NOTES, TODO, action log, and the final 0 FAIL /
  1 WARN / 16 PASS GH-77 audit now reflect the merged structured-logging gate.
- **Remaining Phase 6 work:** broader metrics, dashboards/alerts/objectives,
  production topology/abuse protection, rehearsed runbooks, and sustained
  capacity/retention evidence. Secret-safe structured logging is complete but
  remains defense in depth, not general DLP or rollout approval.
- **Local checks:** documentation consistency PASS, JSON parse PASS, audit
  arithmetic 16 PASS / 1 WARN, and `git diff --check` PASS.
- **Next:** publish a documentation-only closure PR, require exact-head
  docs/static/security review, merge, update GH-29, and synchronize local main.

---

## 2026-07-30 GH-77 → Done / Target Stop

- **Implementation:** PR #78 merged as `133fb3b`; GH-77 closed automatically.
  Exact implementation head `fd8378c` passed all 11 checks, including Go
  build/vet/race/coverage, the complete recovery matrix, Docker/Compose
  post-restart structured JSONL, docs/security/static analysis, and review.
- **Closure:** documentation-only PR #79 exact head `3d3eb5a` passed all eight
  applicable docs/security/static checks and merged as `69dee43`.
- **Review:** both CodeRabbit threads are resolved and acknowledged; independent
  Kimi full, incremental, and closure reviews found no P0/P1/P2. The final audit
  records 0 FAIL / 1 WARN / 16 PASS.
- **Public state:** 1,092 authoritative cases (1,058 Go, 26 Rust, eight
  maintained-client). Roadmap, status JSON, README, landing page, wiki,
  limitations, agent state, TODO, audit, and GH-29 are synchronized. GH-29
  remains open for the unfinished rollout gates.
- **Safety boundary:** secret-safe structured logging is defensive minimization,
  not general DLP, production approval, or public-network authorization. No
  production, server, collector, deployment, credential, key, mainnet,
  consensus/state, or public-network action occurred.
- **Next project work:** continue GH-29 Phase 6 with broader consensus/peer/block/
  invariant/resource/application metrics and their operator evidence.

`TASK COMPLETE — TARGET STOP ACTIVE`

---

## 2026-07-30 09:50 EEST GH-80 Node/Application Metrics Baseline → In Progress

- **Issue/branch:** [GH-80](https://github.com/NeaBouli/TrueRepublic/issues/80),
  `feature/GH-80-metrics-baseline`; parent rollout tracker
  [GH-29](https://github.com/NeaBouli/TrueRepublic/issues/29).
- **Baseline:** clean synchronized `main`/`origin/main` at `ed9c971`; no open
  PR and current main Security Scan green.
- **Verified gap:** CometBFT already exposes loopback-only Prometheus metrics,
  but Cosmos SDK application telemetry is disabled. Existing Compose evidence
  checks only target health and does not require SDK/runtime, TrueRepublic
  application, or successful invariant-cycle families.
- **Scope:** retain the private CometBFT consensus/peer/block source; add an
  opt-in loopback-only SDK/application source, bounded custom metrics without
  user/transaction/address/peer-derived labels, and real runtime assertions
  including post-restart progression.
- **Boundary:** repository code, tests, local recovery Compose configuration,
  documentation, and CI evidence only. Dashboards/alerts/objectives/ownership,
  production collectors/topology, capacity, and retention remain separate.
- **Owner:** Codex Sol + Kimi K3. Kimi receives one bounded secret-free local
  implementation block; Sol retains architecture, security, integration,
  full tests, GitHub writes, review, and closure.
- **Next:** freeze the metric contract and telemetry enable/disable semantics,
  review Kimi's diff, integrate repository/process/Compose evidence, then run
  the complete relevant gate chain.

---

## 2026-07-30 10:20 EEST GH-80 Local Implementation and Audit → In Progress

- **Implementation:** retained the private CometBFT source and added an opt-in
  loopback SDK/application scrape source. Five fixed, label-free
  `truerepublic_*` families expose successful EndBlock/invariant heights,
  process-local completed blocks, canonical `upnyx` supply, and exact
  21-million-PNYX headroom.
- **Semantics:** custom collectors update only after successful module
  EndBlock. Crisis is last with check period one, so invariant success is
  mechanically coupled and regression-guarded. Application height is not
  overclaimed as durable commit height.
- **Kimi contribution:** Kimi K3 implemented the bounded collector/registration
  core, app wiring, coupling guard, and focused tests. Sol reviewed every line,
  corrected the portable native telemetry configuration, added real-process
  enable/restart/disable evidence, private scrape/CI integration, documentation,
  and the structured audit.
- **Focused evidence:** build/vet, observability normal/race tests, full root
  tests, wrapper/policy tests, real compiled daemon start/restart/404-disable
  path, YAML/shell parsing, docs consistency, and diff checks PASS. The first
  root run correctly exposed a nonportable BSD `sed` expression in Sol's
  concurrent script work; the portable correction passes in isolation and in
  the repeated full root run.
- **Audit:** `docs/agent-bridge/GH80_AUDIT.md` records 0 FAIL / 3 WARN /
  10 PASS. Residuals are private generic SDK query-path cardinality, EndBlock
  versus committed-height interpretation, and Docker/Compose evidence pending
  because Docker is unavailable locally.
- **Boundary:** dashboard provisioning/panels, alerts, objectives, paging,
  ownership, production collectors/topology, capacity, and retention remain
  separate Phase-6 gates.
- **Next:** obtain a fresh independent full-diff Kimi review, remediate valid
  findings, then run the complete local repository/recovery gate chain before
  publication.

---

## 2026-07-30 11:22 EEST GH-80 Independent Review Corrections → In Progress

- **Kimi review:** the fresh read-only full-diff review reproduced two local
  gate failures and one CI flake risk in the moving working tree. Sol
  independently verified each path before changing files.
- **Corrections:** the documentation policy guard now matches the actual
  baseline disclaimer; the disabled raw SDK route is correctly pinned to the
  SDK gateway's fail-closed `501 Not Implemented` response; the durable
  Compose contract no longer requires the one-shot
  `truerepublic_server_info` series after its 60-second retention window.
- **Security:** nginx now explicitly returns 404 for exact
  `/api/metrics`, `nginx/**` triggers the runtime workflow, and the CI wait
  proves that the query proxy does not expose raw metrics.
- **Fresh evidence:** policy/coupling tests PASS; the real compiled daemon
  start, persistent restart, metric progression, exact cap arithmetic, and
  disabled-501 lifecycle PASS (`ok truerepublic 71.676s`);
  observability race and diff checks remain PASS.
- **Correction to 10:20 entry:** its `404-disable` wording and complete-root
  PASS claim described an intermediate tree and are superseded here. The raw
  SDK fallback is 501; nginx's public proxy boundary is 404. Full final gates
  are still running, so GH-80 remains In Progress.
- **Next:** complete the already-running eight-scenario recovery matrix, obtain
  Kimi's final severity report against the corrected diff, then run the
  complete repository verification chain before publication.

---

## 2026-07-30 12:04 EEST GH-80 Complete Local Gates → Ready for Publication

- **Review:** Kimi's full and incremental read-only reviews now report zero
  open P0/P1. Sol remediated both P2 findings: the process contract no longer
  depends on expiring `truerepublic_server_info`, and native telemetry
  configuration validates the complete real `[telemetry]` postcondition
  fail-loud while preserving its backup on failure.
- **Go evidence:** `make verify` PASS across the exact nine-package selector:
  manifest, build, vet, race, and coverage. Root PASS in `138.814s` at 69.1%
  coverage; the eight-scenario recovery matrix PASS in `1155.6s`.
- **Runtime/security evidence:** compiled daemon start, persistent restart,
  application/invariant progression, exact 21-trillion-`upnyx` reconciliation,
  and raw disabled SDK `501 Not Implemented` PASS. Exact nginx query-proxy
  boundary remains `404`.
- **Rust/client evidence:** Rust format/clippy/build, 26 tests, and audit PASS
  with six allowed warnings. Maintained `client-web` install/lint/eight tests/
  build/high-audit PASS; two moderate React Router advisories and the existing
  bundle-size warning remain non-blocking.
- **Inherited debt:** informational legacy `web-wallet` and `mobile-wallet`
  audits still report high/critical dependency advisories and exit non-zero,
  exactly as their existing Security Scan steps tolerate with `|| true`.
  GH-80 changes none of those dependency trees; remediation is a separate
  rollout/security task, not hidden as a metrics change.
- **Environment:** a disk-full false failure was traced to the regenerable
  9.6-GiB Go build cache. `go clean -cache` restored capacity; the cold full
  root suite and official verify gate then passed.
- **Remaining GH-80 gate:** Docker is unavailable locally. Publish the reviewed
  branch and require every exact-head GitHub check, especially
  Docker/Compose metrics/restart evidence, before merge.

---

## 2026-07-30 12:15 EEST GH-80 Exact-Head CI Fix → In Progress

- **Publication:** implementation commit `ea6adc6` is pushed and draft
  [PR #81](https://github.com/NeaBouli/TrueRepublic/pull/81) targets `main`.
- **Green exact-head checks:** build/test, docs consistency, Go vulnerability,
  Rust audit, maintained-client audit, both informational legacy-wallet audit
  jobs, and DeepScan.
- **Reproduced failure:** Docker reached the new metrics-contract step after
  the loopback stack became ready, then Compose interpolation rejected the
  missing required `GRAFANA_PASSWORD` before any metric assertion executed.
- **Focused fix:** the contract step now carries the same CI-only Grafana
  password and bootstrap operator as the stack-start step. YAML parse, static
  network policy, consistency, and diff checks PASS locally.
- **Remaining:** push the focused fix, require the rerun Docker metrics/restart
  step and the still-running recovery matrix to pass, then enable full review.

---

## 2026-07-30 12:24 EEST GH-80 Metrics Contract Diagnostics → In Progress

- **Second exact-head result:** Compose environment and both Prometheus targets
  now pass. The contract exits inside its pre-restart family/value assertions,
  but GitHub's default `bash -e` log does not reveal the silent failing
  `grep`/`awk`.
- **Diagnostic hardening:** required-family checks now emit the missing family
  plus the available public metric names; the exact height/invariant/supply/
  headroom values are logged before arithmetic assertions. Acceptance rules
  are unchanged.
- **Local evidence:** workflow YAML parse, static network-policy test, and diff
  check PASS.
- **Next:** publish the fail-loud workflow and use its exact-head evidence to
  fix the concrete family/value mismatch without weakening the contract.

---

## 2026-07-30 12:39 EEST GH-80 Zero-Peer Semantics → Fix In Progress

- **Exact evidence:** run `30499586251`, job `90736070552`, passed the image,
  persistent-restart, loopback stack, proxy, and both private target gates.
  The contract then reported only `cometbft_p2p_peers` missing.
- **Root cause:** CometBFT instantiates that gauge on the first peer
  add/remove event. A deliberate single-validator stack with zero peers omits
  the series rather than exporting a zero sample.
- **Focused repair:** the singleton CI contract now requires the proven
  always-instantiated `cometbft_mempool_size` family. Production peer
  observability is retained, while its dashboard and documented alert use
  `or vector(0)` so an absent zero-peer series is treated as zero.
- **Acceptance remains strict:** no endpoint, restart, progression,
  application, supply, or headroom assertion is weakened.
- **Next:** run focused policy/docs/JSON validation, publish, and require a
  fully green new exact-head Docker and Recovery result.

---

## 2026-07-30 12:42 EEST GH-80 Zero-Peer Repair → Locally Verified

- Workflow YAML and Grafana dashboard JSON parse PASS.
- `TestContainerNetworkDefaultsFailClosed` PASS in 3.131s.
- Documentation consistency and `git diff --check` PASS.
- The current exact-head Go root suite and eight-scenario recovery matrix are
  still running independently; their result will be recorded before replacing
  that head with the focused repair.
- **Next:** commit and push the five-file repair, then require all checks on the
  replacement exact head.

---

## 2026-07-30 12:47 EEST GH-80 Recovery Evidence Reconciled

- GitHub's status API still labels recovery job `90736070604` in progress, but
  its completed runner log is authoritative: all eight named scenarios PASS in
  666.955s, followed by normal cache and checkout cleanup.
- The same head's root build/test job and all security jobs pass. Its sole
  product-relevant failure remains the now-understood zero-peer family
  assumption in Docker.
- No recovery timeout defect exists: the command already enforces Go's
  1500-second test timeout. The stale job state is a GitHub control-plane
  reporting lag, not a hanging TrueRepublic process.
- **Next:** publish the locally verified zero-peer repair and verify a new
  exact head end to end.

---

## 2026-07-30 13:10 EEST GH-80 External Review → Fixes In Progress

- Exact head `24e00eb` passed Docker/Compose, root build/test, all eight
  recovery scenarios, docs, security, DeepScan, and CodeRabbit's status gate.
- CodeRabbit completed its first non-draft review with three unresolved,
  independently verified findings:
  1. prefix/substring metric checks could accept a sibling sample or HELP line;
  2. the fail-closed init subprocess lacked a local timeout;
  3. nginx blocked `/api/metrics` but not its descendant paths.
- All three are within GH-80's fail-closed observability scope. Minimal fixes
  will use exact sample parsing, a bounded command context, and a descendant-
  aware 404 rule with runtime coverage.
- The generic docstring-coverage warning is not a changed-code defect or a
  repository-required gate and will not trigger unrelated repo-wide churn.
- **Next:** validate the focused review fixes, publish them, resolve only the
  supported threads, and require the replacement exact head to pass.

---

## 2026-07-30 13:18 EEST GH-80 Review Fixes → Locally Verified

- Exact Prometheus parsing now rejects HELP text and sibling names; histogram
  `_sum` and `_count` samples are required explicitly.
- Both init-script subprocess tests use 10-second command contexts.
- Nginx denies `/api/metrics` and every descendant path; CI now proves both the
  query-string form and `/api/metrics/` return 404.
- Focused four-test set PASS in 4.719s; the false-positive parser regression,
  standalone exact-sample shell proof, workflow YAML, `go vet .`, docs
  consistency, and diff checks PASS.
- Full `make verify` PASS: build, vet, race, coverage, and all nine repository
  packages; root coverage remains 69.1%.
- Local nginx binary remains unavailable. Exact syntax/runtime proof is
  delegated to the mandatory Docker/Compose gate on the replacement head.
- **Next:** publish the review-fix commit, await all exact-head gates and the
  incremental CodeRabbit review, then resolve supported threads.

---

## 2026-07-30 13:37 EEST GH-80 Implementation Merged → Closure In Progress

- **Merge:** PR #81 exact head `4b880b4` passed all checks and squash-merged to
  `main` as `e629374`; GH-80 closed automatically and its remote branch was
  deleted.
- **Exact-head evidence:** root build/test 7m32s, Docker/Compose 7m43s, all
  eight recovery scenarios 14m09s, docs, Go/Rust/Node security, DeepScan, and
  CodeRabbit status PASS. All three actionable review threads were remediated,
  answered, and resolved.
- **Merged-main evidence:** Security Scan, Pages, and dependency graph PASS;
  post-merge Go CI is still running and remains a closure gate.
- **Authoritative count:** fresh package-scoped JSON output on merged main
  records 1,070 Go cases: root 82, healthcheck 55, migration 82, network policy
  126, observability 50, token 12, treasury 36, DEX 116, governance 511. With
  26 Rust and eight maintained-client cases, the public total is 1,104.
- **Closure branch:** `docs/GH-80-metrics-closure` from `origin/main`.
  Documentation-only synchronization covers counts, Phase 6, limitations,
  machine status, wiki, audit, TODO, Project State, Bridge, and GH-29.
- **Safety boundary:** dashboards/application panels, alert rules, objectives,
  escalation ownership, capacity/topology, runbooks, and production deployment
  remain open. No production action occurred.
- **Next:** finish consistency/audit validation, await post-merge Go CI, publish
  the closure PR, verify it, merge, and synchronize GH-29/main.

---

## 2026-07-30 13:49 EEST GH-80 Closure → Locally Verified

- **Post-merge main:** Go CI is fully green on `e629374`: build/test, complete
  eight-scenario recovery matrix, and Docker/Compose. Security, Pages, and
  dependency graph are also green.
- **Independent review:** Kimi K3 reproduced all 1,070 Go cases and coverage,
  confirmed 1,104 total arithmetic, PR/job evidence, audit arithmetic, and the
  non-production boundary. Result: no P0/P1; one pre-existing P2.
- **P2 remediation:** `docs/DEPLOYMENT.md` no longer claims rendered Grafana
  panels. It now distinguishes verified container/datasource health and private
  scrape targets from still-open provisioning, panel, alert, objective, and
  ownership gates.
- **Local closure checks:** package-scoped JSON count PASS; documentation
  consistency PASS; status JSON parse PASS; audit arithmetic 0 FAIL / 2 WARN /
  11 PASS; diff check PASS. The diff is documentation/status only.
- **GH-29:** Phase 6 metrics checkbox now links GH-80 / PR #81; dashboard,
  alert, objective, and ownership work remains unchecked.
- **Next:** commit/push, open documentation-only closure PR, require all
  applicable exact-head checks, merge, and synchronize clean `main`.

---

## 2026-07-30 21:52 EEST GH-29 GitHub Page rollout status → In Progress

- **Branch:** `docs/GH-29-github-page-status`
- **Issue:** [GH-29](https://github.com/NeaBouli/TrueRepublic/issues/29)
- **Scope:** update the public GitHub Page to show the canonical tracker
  progress (10/59, approximately 17%), Phase 6 progress (3/7), current green
  engineering baseline, and the next operations milestone.
- **Correction:** Phase 1 still presents validator-key rotation and compromise
  response as open even though GH-55/GH-56 are merged. The page will
  distinguish that completed evidence from the still-open consensus-breaking
  migration gate.
- **Risk:** Low. Documentation-only public status update; no runtime,
  consensus, token, wallet, infrastructure, or production mutation.
- **Tests required:** documentation consistency, HTML/status assertions, diff
  hygiene, exact-head Pages/security CI, and deployed-page verification.
- **Next:** implement the bounded page change, validate locally, publish a
  reviewable PR, merge only after green exact-head checks, then verify the
  deployed GitHub Page and synchronize GH-29/Bridge.

---

## 2026-07-30 21:56 EEST GH-29 GitHub Page rollout status → Review

- **Changed:** `docs/index.html`, `docs/status.json`,
  `scripts/check-consistency.sh`, `BRIDGE.md`, and
  `docs/agent-bridge/ACTION_LOG.md`.
- **Public result:** the landing page now exposes 10/59 (17%) full rollout,
  10/51 (20%) phase work, 3/7 (43%) Phase 6, 1,104 verified cases, green
  CI/security, the 21M cap, and the next dashboards/alerts milestone.
- **Correction:** completed validator custody/failover/rotation/revocation and
  network-policy evidence no longer appears open in Phase 1.
- **Tests:** consistency PASS; shell syntax PASS; status JSON PASS; diff check
  PASS; desktop 1280 and mobile 390 browser/DOM rendering PASS with no mobile
  horizontal overflow.
- **Risk:** Low, documentation-only. Exact-head Pages/security checks and the
  deployed GitHub Page remain required before Done.
- **Ready for:** Sol publication and final GitHub verification.

---

## 2026-07-30 21:58 EEST GH-29 GitHub Page rollout status → Review

- **PR:** [#83](https://github.com/NeaBouli/TrueRepublic/pull/83)
- **Published implementation:** `e8251b2` on
  `docs/GH-29-github-page-status`.
- **State:** documentation-only draft PR is open. This append-only handoff
  becomes the replacement exact head after push.
- **Remaining gates:** exact-head docs/static/security checks, review-thread
  reconciliation, merge, clean-main synchronization, and verification of the
  deployed GitHub Page.
- **Boundary:** no production deployment or runtime code change.

---

## 2026-07-30 22:06 EEST GH-29 PR #83 review follow-up → In Progress

- **Source PR:** [#83](https://github.com/NeaBouli/TrueRepublic/pull/83),
  merged as `4546906`.
- **Finding:** the rate-limited full review returned green but created two
  unresolved minor threads immediately before merge. Both are valid:
  1. clarify that exit gates are counted in 59 but remain open;
  2. validate rollout integers and positive totals before division.
- **Branch:** `fix/GH-29-page-review-followup` from merged `origin/main`.
- **Scope:** minimal page wording and fail-closed consistency-script
  validation, targeted regression checks, thread replies/resolution, and a
  follow-up PR.
- **Risk:** Low, documentation and docs-gate hardening only. No runtime or
  production change.
- **Next:** implement both review corrections, prove malformed/negative/zero
  status input fails safely, publish, and require fresh exact-head checks.

---

## 2026-07-30 22:08 EEST GH-29 PR #83 review follow-up → Review

- **Changed:** `docs/index.html`, `scripts/check-consistency.sh`,
  `BRIDGE.md`, and `docs/agent-bridge/ACTION_LOG.md`.
- **Finding 1:** wording now says exit gates are included in the 59 total but
  remain open.
- **Finding 2:** all six rollout fields are type/range validated before the
  first division; totals must be positive integers and completed counts must be
  nonnegative integers.
- **Tests:** full consistency PASS; shell/JSON/diff PASS; valid integer fixtures
  PASS; negative, fractional, string, null, and zero-total rejection PASS.
- **Risk:** Low, documentation/docs-gate only. Ready for a focused follow-up PR
  and fresh exact-head verification.

---

## 2026-07-30 22:09 EEST GH-29 PR #83 review follow-up → Review

- **PR:** [#84](https://github.com/NeaBouli/TrueRepublic/pull/84)
- **Published fix:** `323abad` on
  `fix/GH-29-page-review-followup`.
- Both PR #83 threads have evidence-backed replies linking the follow-up; they
  remain open until the correction is merged.
- This append-only handoff becomes the replacement exact head.
- **Remaining:** fresh exact-head checks/review, merge, resolve both source
  threads, synchronize clean main, and verify deployed Pages.

---

## 2026-07-30 22:23 EEST Phase 6 dashboards, alerts, objectives, and ownership → In Progress

- **Baseline:** clean `main` equals `origin/main` at `844c498`; the latest
  Pages and Security workflows are green, no pull request is open, and GH-29
  remains the canonical rollout tracker.
- **Scope:** ship a provisioned Grafana dashboard with CometBFT plus bounded
  application panels, reviewed Prometheus alert rules, recovery/testnet
  service objectives, role-based escalation ownership, deterministic
  validation, operator guidance, and GitHub runtime evidence.
- **Finding:** the existing dashboard file uses Grafana's API wrapper shape
  while file provisioning expects a native dashboard document. Current CI
  proves only Grafana health and the datasource, not dashboard provisioning,
  panel expressions, loaded rules, or alert behavior.
- **Boundary:** repository configuration, tests, CI, and documentation only.
  No Alertmanager destination, external paging, production topology, server,
  deployment, IAM, credentials, keys, mainnet, or public-network mutation.
- **Risk:** medium. Incorrect queries or thresholds can create silent gaps or
  alert noise; file provisioning and rule evaluation therefore require
  machine-checked and Compose runtime evidence.
- **Next:** create one bounded GitHub child issue, branch from this exact
  baseline, delegate the dashboard/rule implementation block to Kimi K3, then
  have Sol review, integrate, and run focused plus complete verification.

---

## 2026-07-30 22:25 EEST GH-85 observability operations → In Progress

- **Issue:** [GH-85](https://github.com/NeaBouli/TrueRepublic/issues/85)
- **Branch:** `feature/GH-85-observability-operations` from exact clean main
  `844c498`.
- Acceptance criteria cover native dashboard provisioning, reviewed
  CometBFT/application panels, Prometheus rule loading and deterministic rule
  tests, recovery/testnet objectives, role-based escalation ownership,
  Compose API/query evidence, documentation, and final GitHub verification.
- External notification routing, production deployment, and production SLO
  approval are explicitly excluded.
- **Next:** record the queue/state transition, publish this initial handoff,
  then assign the bounded dashboard/rule file implementation to Kimi K3 while
  Sol owns validation architecture, CI integration, documentation, and review.

---

## 2026-07-30 22:49 EEST GH-85 implementation → Locally Focused Green

- **Implemented:** native immutable Grafana dashboard with 16 reviewed
  CometBFT/application/supply/runtime panels; stable datasource/provisioning;
  ten Prometheus alerts; deterministic promtool transitions; service
  objectives; role-based escalation and response guidance; Compose mounts;
  repository contracts; and GitHub runtime API/query checks.
- **Kimi contribution:** the bounded implementation invocation independently
  analyzed the exact metric/job contract, native dashboard shape, alert
  coverage, PromQL matching, and synthetic transition design. It remained in
  planning without writing a diff and was stopped; Sol implemented the
  bounded design and therefore has no unreviewed Kimi write to integrate.
- **Focused PASS:** JSON and all YAML parse; workflow YAML parses; two
  repository contract tests pass; Prometheus 3.13.1 validates config syntax
  and all ten rules; all promtool alert transition tests pass; every dashboard
  and objective PromQL expression parses; docs consistency and diff hygiene
  pass.
- **Local limitation:** Docker is unavailable, so actual Grafana file
  provisioning, Prometheus absolute rule-path loading, live panel results, and
  rule API health remain mandatory GitHub Compose gates.
- **Next:** run the complete local build/verify/security-relevant gates,
  request an independent read-only Kimi review of the finished diff, remediate
  findings, and only then publish the implementation PR.

---

## 2026-07-30 23:09 EEST GH-85 review remediation → In Progress

- **Independent review:** Kimi found 0 P0, 0 P1, one P2, and five P3 items
  against the moving worktree. The substantive P2 was cross-instance signal
  mixing; P3s covered stale Bridge counts, pin wording, missing Grafana
  datasource runtime proof, weak dashboard count floors, and one missing
  alert-recovery fixture.
- **Remediated:** application/invariant/supply alerts now preserve per-instance
  label pairing; consensus/application divergence pairs the two scrape sources
  through one bounded static `node` label. Prometheus config documents the
  pairing contract for future multi-node targets.
- **Hardened:** eleven exact alert names/counts, a required-metric-family alert,
  deterministic cases for every rule including missing-family recovery, exact
  16-panel/17-target assertions, tag+digest-pinned Prometheus/Grafana images,
  Grafana datasource UID plus proxy-query runtime proof, and aligned canonical
  docs/wiki examples.
- **PASS after remediation:** Prometheus 3.13.1 config syntax, eleven-rule
  validation, all promtool cases, JSON/YAML, and PromQL parsing.
- The earlier ten-rule entries remain valid historical snapshots; this
  append-only entry supersedes them with the current eleven-rule state.
- **Next:** finish the running eight-scenario recovery matrix, rerun full
  exact-worktree verification, obtain a bounded Kimi remediation recheck, and
  create the audit artifact before publication.

---

## 2026-07-30 23:22 EEST GH-85 exact local verification → Review Ready

- **Recovery PASS:** all eight maintained multi-validator process scenarios
  passed on the current code path: legacy-authority migration/rollback,
  consensus recovery, trusted snapshot state sync, backup/restore
  export/import, binary upgrade/rollback, validator identity failover,
  consensus-key rotation, and consensus slashing.
- **Full PASS:** `make build`; package build and vet; race/coverage across all
  nine supported Go packages; Prometheus 3.13.1 config/rules/tests; dashboard
  JSON and provisioning/workflow YAML; all 17 panel expressions and six
  objective queries; focused repository contracts; docs consistency; and diff
  hygiene.
- **Measured evidence:** the Go suite now contains 1,071 passing cases (83
  root/application plus the unchanged package counts). Public status remains
  at the merged 1,104-case baseline until GH-85 and its one new Go contract
  test merge; the post-merge synchronization will publish 1,105 total cases.
- **Independent recheck:** Kimi verified all seven remediation areas as PASS
  and found 0 new P0, 0 P1, and 0 P2 defects. Sol additionally hardened the
  repository test to pin both bounded static `node` labels and `on (node)`
  divergence matching.
- **Audit:** `docs/agent-bridge/GH85_AUDIT.md` records 0 FAIL / 0 WARN / 7
  PASS. Docker is unavailable locally, so the pinned Compose provisioning,
  datasource query, rule API, and panel-query proof remains a mandatory
  exact-head GitHub gate.
- **Next:** publish the implementation commit and draft PR, obtain all
  exact-head GitHub checks and review evidence, remediate any findings, then
  merge and synchronize GH-29, public rollout status, test counts, and Bridge.

---

## 2026-07-30 23:28 EEST GH-85 published as draft PR #86

- **Commit:** `1c44b2f` (`feat: add observability operations baseline`) is
  pushed on `feature/GH-85-observability-operations`.
- **PR:** [#86](https://github.com/NeaBouli/TrueRepublic/pull/86) targets
  `main`, closes GH-85 after merge, and carries the implementation, risk
  boundary, local evidence, and mandatory GitHub runtime checks.
- GH-85 and GH-29 now link the PR and record that rollout/public status stays
  open until exact-head checks pass and the implementation merges.
- **State:** review in progress. No production, server, mainnet, IAM, paging,
  or notification action occurred.
- **Next:** publish this append-only PR handoff as the replacement head and
  monitor every GitHub check/review; fix valid findings before merge.

---

## 2026-07-30 23:38 EEST PR #86 Docker runtime finding → Fix In Progress

- Exact head `67d5d4c` passed Go build/race/coverage in 7m26s, docs,
  DeepScan, Go/Rust/maintained-client/legacy-client security, while recovery
  and CodeRabbit continued.
- The Docker job passed Compose config, pinned Prometheus config, eleven-rule
  validation and rule tests, image builds, node readiness, both scrape targets,
  and Grafana health. Its new combined dashboard/rule/query assertion then
  exited without identifying which suppressed `jq` cardinality check failed.
- **Root risk:** target health does not prove that a lazily initialized
  application series or first rule evaluation is already queryable. The
  original one-shot assertions also discarded the safe response summary
  needed to distinguish timing from a real provisioning mismatch.
- **Fix:** dashboard/datasource/rule/query contracts now emit bounded,
  secret-free failure summaries. Datasource application-height, rule-health,
  and all four required application-series assertions retry for up to 30
  seconds; structural dashboard/datasource contracts and every PromQL query
  still fail immediately on a real mismatch.
- **Kimi:** focused read-only diagnosis independently identified the
  target-up-versus-series-ready race and first-evaluation timing as the likely
  failure class; Sol owns and validates the workflow change.
- **Next:** run local workflow/repository contracts, publish a replacement
  head, and require the complete GitHub matrix again.

---

## 2026-07-30 23:45 EEST GH-85 bounded retry review → Hardened

- Kimi's completed focused review confirmed 16 panels, 17 targets, eleven
  alerts, and the repository contract test, and ranked first rule evaluation
  as the most likely original runtime race.
- It identified one remaining low-risk shell edge: under GitHub's `bash -e`,
  a transient HTTP failure inside command substitution could bypass the new
  retry loop before a JSON response existed.
- Sol remediated the edge by placing each retryable curl assignment inside an
  explicit conditional and initializing a bounded unavailable response.
  Transient HTTP failures now retry; structural or PromQL failures remain
  strict and diagnostic.
- **Next:** publish the hardened replacement head and use only its complete
  GitHub matrix as merge evidence.

---

## 2026-07-30 23:48 EEST PR #86 CodeRabbit findings → Remediated

- Thread-aware review found three valid issues on the implementation diff:
  an unescaped regex pipe breaking the three-column objectives table, a
  four-row overlap between application-rate and divergence panels, and no
  explicit recovery transition for the low-peer rule.
- **Fixed:** escaped the Markdown table delimiter without changing the rendered
  PromQL; reduced panel 9 to the intended four-row height; added a low-peer
  firing/recovery fixture at the three-peer threshold.
- **Prevented:** the repository dashboard contract now validates bounded
  24-column grid positions and rejects every pairwise panel overlap.
- **PASS:** dashboard JSON; promtool rule suite including the new recovery
  transition; focused repository contract; docs consistency; and diff hygiene.
- **Review state:** all three threads are technically addressed locally; reply
  and resolution follow after publication.
- **Next:** publish the replacement head, reply/resolve the exact threads, and
  require its full GitHub matrix.

---

## 2026-07-30 23:58 EEST PR #86 Docker cardinality root cause → Fixed

- Exact review head `513a122` passed Go build/race/coverage in 7m28s and all
  completed docs/static/security gates. Docker reached the fully diagnostic
  observability assertion and reported two results for an unscoped
  application-height query.
- **Root cause:** the final four-series presence loop omitted the canonical
  `job="truerepublic-app"` selector. The running process legitimately exposes
  the registered application metrics through both monitored surfaces, so an
  unscoped query returns two series. This was a deterministic CI assertion
  error, not missing telemetry and not a relaxed cardinality boundary.
- **Fix:** all four final required-series queries now use the same explicit
  `truerepublic-app` job selector as the dashboard, alerts, objectives, and
  Grafana proxy proof. The gate still requires exactly one canonical series.
- **PASS:** workflow YAML, all four selector PromQL expressions, focused
  repository contract, and diff hygiene.
- **Next:** publish the corrected head and require the complete GitHub matrix
  once more; no merge until Docker and recovery both pass.

---

## 2026-07-31 11:37 EEST Phase 6 topology qualification → In Progress

- **Owner:** Codex Sol + Kimi K3; **Ticket:** `GIO-20260731-TR-033`.
- **Issue/branch:** [GH-89](https://github.com/NeaBouli/TrueRepublic/issues/89),
  `feature/GH-89-topology-abuse-qualification`.
- **Baseline:** clean `main` at `fcf8428`; GH-71 already provides fail-closed
  validation for each individual seed, sentry, validator, RPC, and private-node
  home. GH-85 observability is complete.
- **Gap:** the repository does not yet define or validate one coherent
  multi-node deployment inventory, role relationships, validator isolation,
  public RPC ingress, firewall flow matrix, or aggregate abuse-protection
  contract.
- **Scope:** introduce a secret-free example topology schema and deterministic
  validator, negative/positive regressions, CI/repository contracts, and an
  operator qualification/rollback procedure that composes the existing GH-71
  role policy without duplicating it.
- **Safety boundary:** repository artifacts only. No production/server/cloud,
  DNS, firewall, IAM, real host/IP, deployment, mainnet, key, credential, or
  external-network change. The GH-29 production-deployment checkbox stays
  unchecked until separately authorized live deployment evidence exists.
- **Next:** obtain a bounded secret-free Kimi topology review, then implement
  under Sol architecture,
  security, integration, testing, GitHub, and closure ownership.

---

## 2026-07-31 11:42 EEST GH-89 architecture review → Core implementation

- **Kimi review:** read-only gap analysis confirmed GH-71 is strictly
  per-initialized-home and cannot prove reciprocal sentry protection, unique
  identities, failure-domain diversity, aggregate flow denial, or public
  ingress controls across one intended topology.
- **Accepted design:** a separate versioned, strict JSON topology-policy
  package; synthetic committed example; deterministic reports; explicit
  deny-by-default flows; abuse-control ceilings; no DNS lookup or environment
  expansion. The package composes GH-71 concepts and does not replace its
  effective-home validation.
- **Privacy correction:** a real contract containing validator public node IDs
  is operator-private evidence. Only synthetic IDs and `.invalid` hosts may be
  committed, preventing the public repository from defeating sentry identity
  privacy.
- **Delegation:** Kimi owns only the new core package and synthetic fixture,
  may not delegate, and has no production/external authority. Sol owns every
  existing-file edit, CLI/docs/CI integration, diff review, complete tests,
  GitHub publication, and closure.
- **External sync:** GH-89 and parent GH-29 contain the active scope and
  explicit statement that the production-deployment checkbox remains open.

---

## 2026-07-31 13:02 EEST GH-89 local implementation → Review Ready

- **Implemented:** strict `truerepublic.topology/v1` contract parser,
  deterministic cross-node validator, offline `topology-policy validate` CLI,
  five-node synthetic example, explicit CI command, repository contracts, and
  operator qualification/rollback guidance.
- **Review remediated:** reserved external-principal collision, trailing-value
  reflection, non-public special addresses, IPv6 endpoint aliases,
  percent/escape-ambiguous routes, stale disabled ingress, and the initial
  Cosmos config-interceptor panic all have regressions.
- **Full PASS:** package selection, build, vet, race/coverage across ten Go
  packages; topology policy 86.2%, root/application 69.1%; docs consistency,
  JSON/YAML, real CLI, and diff hygiene.
- **Process PASS:** all eight maintained recovery scenarios passed across two
  unchanged sequential runs. Four passed before the first run's 25-minute
  global budget was consumed by local linker contention; the remaining four
  then passed explicitly.
- **Candidate arithmetic:** 1,079 Go + 26 Rust + eight maintained-client =
  1,113 top-level cases. Public status remains at 1,105 until merged-main
  reproduction and the closure sync.
- **Boundary:** local Docker is unavailable, so GitHub Docker/Compose,
  exact-head security/static checks, review, and zero unresolved threads remain
  mandatory before merge. GH-29's real production deployment stays open.
- **Next:** commit/push the implementation, open the review PR, and require
  every exact-head GitHub gate before merge.

---

## 2026-07-31 13:10 EEST GH-89 published as draft PR #90

- **Commit:** `035993c` is pushed on
  `feature/GH-89-topology-abuse-qualification`.
- **PR:** [#90](https://github.com/NeaBouli/TrueRepublic/pull/90) targets
  `main`, closes GH-89 only after merge, and carries the complete local
  evidence and unchanged production boundary.
- GH-89 and parent GH-29 now link PR #90. The production-deployment checkbox,
  rollout counts, and public Pages test count remain unchanged until verified
  merged-main evidence exists.
- **State:** exact-head GitHub build/recovery/Docker/docs/security/review gates
  are running. No production or external infrastructure action occurred.

---

## 2026-07-31 14:01 EEST PR #90 exact-head CI remediation → Revalidation

- **Observed:** the first exact-head `build-and-test` run failed only at the
  maintained topology JSON assertion. The validator produced
  `valid=true`, five nodes, and an empty violation list, but the daemon root
  routed the report to stderr, so `jq` received no document (exit 4).
- **Fix:** `server_lifecycle.go` now binds normal root-command output to stdout
  and errors to stderr. `server_lifecycle_test.go` adds the regression; the
  GH-89 audit records the evidence.
- **Local PASS:** formatting/diff hygiene, focused root stream and command
  registration tests, and the exact real-daemon CI pipeline returning `true`.
- **GitHub PASS on prior head:** docs consistency, Go vulnerability scan, all
  three Node audits, Rust audit, DeepScan, CodeRabbit with zero review
  threads, Docker/restart smoke, and all eight multi-validator recovery drills.
- **State:** the corrected head still requires the complete exact-head GitHub
  matrix. No production action occurred, rollout counts remain unchanged, and
  GH-29's real deployment checkbox remains open.

---

## 2026-07-31 14:21 EEST PR #90 corrected-head local gate → PASS

- Documentation/rollout consistency, package selection, full build, vet, and
  complete ten-package race/coverage suite all pass after the stream fix.
- Coverage remains root/application 69.1%, topology policy 86.2%, healthcheck
  97.2%, migration 84.6%, network policy 95.5%, observability 80.3%, token
  92.6%, treasury 97.0%, DEX 45.3%, and governance 62.2%.
- The real daemon pipeline returns `true`; diff hygiene passes.
- **Next:** commit and push the correction, then require every check and zero
  unresolved review threads on that exact new head before merge.

---

## 2026-07-31 14:23 EEST PR #90 CodeRabbit findings → Remediated

- CodeRabbit reported two valid minor findings on corrected head `a0b3a3b`:
  unchecked contract-file close handling and an API evidence message that said
  RPC.
- `Load` now preserves parse failures and fails closed with an operator-safe
  generic error on a later close failure. The API branch reports
  `internet-to-API`, with an exact regression.
- Focused topology vet and race/coverage PASS (85.8%); diff hygiene PASS.
- **Next:** publish the reviewed fix, reply to and resolve both threads, then
  require the complete exact-head GitHub matrix again.

---

## 2026-07-31 14:40 EEST PR #90 → Merged; public closure started

- Final head `cb14829` passed all 11 checks: build/vet/race/coverage, eight
  recovery drills (14m59s), Docker/Compose (7m52s), docs, all dependency/
  security scans, DeepScan, and CodeRabbit status. Both valid review findings
  were fixed, answered, and resolved; unresolved threads: zero.
- PR #90 merged to `main` as `2f2acdd`. Package-scoped JSON evidence reproduces
  1,130 Go + 26 Rust + eight maintained-client = 1,164 passing cases. The
  earlier 1,113 candidate mixed test-function and historical subtest counting
  and was corrected before publication.
- Branch `docs/GH-89-rollout-status` now synchronizes the public status,
  roadmap, wiki, audit, TODO, and Bridge from merged evidence.
- Rollout arithmetic remains 11/59 overall, 11/51 phase work, and Phase 6 4/7.
  GH-29's live deployment item remains unchecked; no production action occurred.

---

## 2026-07-31 14:55 EEST PR #91 review finding → Remediated

- CodeRabbit found one stale `726` public-status sentence inside the historical
  Project State section. It conflicted with the reproduced 1,164-case status.
- The sentence now preserves historical 577 explicitly while naming 1,164 as
  the active passing-case total. Documentation consistency, JSON/module/
  rollout arithmetic, and diff hygiene pass.
- **Next:** publish the correction, resolve the thread, and require the
  replacement-head status/security/review checks before merge.

---

## 2026-07-31 15:02 EEST GH-89 → Done

- PR #91 corrected head `3a07fa4` passed all docs/static/security/review gates
  with its only finding fixed and zero unresolved threads; it merged as
  `69e498f`.
- Exact merged implementation `2f2acdd` has green Go build/test, all eight
  recovery drills, Docker/Compose, Security Scan, and Pages. Exact merged
  status `69e498f` has green Security Scan and Pages.
- Live Pages and status JSON return 1,164 cases, 11/59 overall, 11/51 phase
  work, Phase 6 4/7, `production_ready=false`, and
  `topology_policy.production_deployment=false`.
- GH-89 acceptance criteria are checked and the issue is closed. GH-29's real
  topology/deployment item remains unchecked. No production mutation occurred.
- **Status:** Done; final append-only handoff publication only.

---

## 2026-08-01 10:29 EEST GH-93 operations runbook rehearsal → In Progress

- **Owner:** Codex Sol + bounded Kimi K3 review; global ticket
  `GIO-20260801-TR-034`.
- **Branch:** `feature/GH-93-operations-runbook-rehearsal` from clean,
  synchronized `origin/main` `ed1b6bf`.
- **Issue:** [GH-93](https://github.com/NeaBouli/TrueRepublic/issues/93), child
  of the rollout tracker GH-29.
- **Scope:** compose the existing specialist procedures into a strict,
  secret-free incident-command and rehearsal contract covering halt,
  validator/slashing failure, consensus-key or operator-authority compromise,
  sanitized backup/restore, compatible binary upgrade/rollback, and the
  separately bounded legacy migration path.
- **Acceptance:** deterministic offline validation, synthetic positive and
  negative scenarios, phase/authority/evidence/abort gates, repository and CI
  tests, canonical operator runbook, audit, independent review, and complete
  relevant verification.
- **Risk:** high in subject matter because incorrect recovery guidance can
  cause double signing, stale signer-state restoration, divergent restart, or
  evidence loss; execution risk is bounded to repository artifacts only.
- **Safety boundary:** no production/server/cloud/IAM/DNS/firewall mutation,
  no real topology, keys, signer state, identities, paging destinations,
  deployment, mainnet, or external incident action. Repository rehearsal is
  not evidence that a live production process is active.
- **Next:** inventory the existing procedures, obtain a secret-free Kimi gap
  review, specify the minimal contract, then implement focused tests before
  integration.

### 2026-08-01 10:34 EEST delegation update

- Kimi K3 `review` was invoked with the bounded secret-free GH-93 scope but
  returned provider HTTP 403 because its billing-cycle quota is exhausted.
  It produced no findings, writes, or repository diff.
- This is not a project blocker. Sol retains the task; a bounded Terra
  read-only review covers the independent gap analysis while Sol continues
  architecture and implementation.

### 2026-08-01 11:02 EEST architecture and focused implementation

- Added strict version `truerepublic.incident-rehearsal/v1`, an offline
  `incident-rehearsal validate` command, and one synthetic eight-scenario
  fixture with seven forward-only phases per scenario.
- The schema permits controlled enums and logical rehearsal-seat aliases only;
  real evidence values, hosts, URLs, commands, paths outside exact maintained
  runbooks, identities, contacts, and secrets have no schema field.
- Scenario gates cover chain halt, validator failure/slashing, consensus-key
  compromise, operator-authority compromise, sanitized backup/restore,
  compatible binary upgrade/rollback, and legacy fresh-genesis migration.
  Cross-scenario actions are rejected; critical approval, single-signer,
  signer-state freshness, fail-before-state-open, source/target isolation,
  target-stop-before-source-restart, invariant, and safe-stop boundaries are
  mandatory.
- Terra's independent read-only review found one P0 documentation conflict:
  the key-rotation guide still claimed ABCI++ slashing was unwired despite the
  merged GH-59 implementation. The claim now matches code and the slashing
  runbook, and a repository regression prevents recurrence. Its network-guide
  exposure warning finding is also remediated without changing infrastructure.
- Spark added only focused `incidentpolicy` tests. Sol reviewed and hardened
  that suite with cross-role alias, wrong-scenario action, command non-leak,
  upgrade, and source/target rollback ordering cases.
- Focused `go test ./incidentpolicy -count=1` PASS; maintained fixture validates
  eight scenarios with an empty violation array. Root/repository and complete
  verification continue next.

### 2026-08-01 12:19 EEST final local verification and review

- Hardened the final contract after independent review: every role now uses an
  exact synthetic seat alias, every signer-relevant scenario requires the
  second-signer abort, the rehearsal chain ID is namespaced, and negative tests
  pin secret-like aliases plus upgrade, slashing, and key-compromise failures.
- Terra's final read-only review reports 0 open P0/P1 findings. Kimi remained
  unavailable because of its provider quota and produced no diff; Spark's
  bounded test contribution was fully reviewed and integrated by Sol.
- The first complete rerun exposed one local lifecycle-harness isolation issue
  under simultaneous heavy compilation: an unrelated TrueRepublic process was
  already using pprof port 6060 and the first-block window was only 35 seconds.
  The harness now disables pprof and uses a 60-second window. The foreign
  process was not touched; the focused race regression passes in 98.573s.
- Final `make verify` PASS: package selection, complete build, vet, race, and
  coverage across all 11 maintained Go packages. Relevant coverage is root
  69.1%, incident policy 87.9%, network policy 95.5%, topology policy 85.8%,
  and governance 62.2%.
- Documentation consistency, strict JSON fixture validation, `git diff
  --check`, root/repository tests, and the real CLI-to-`jq` pipeline PASS.
  `GH93_AUDIT.md` records 0 open FAIL / 0 open WARN / 7 PASS.
- Local Docker is unavailable. Exact-head GitHub Linux, eight-scenario
  recovery, Docker/Compose, docs/static/security, review, and thread gates are
  mandatory before merge. No production or external incident action occurred.
- **Status:** In Progress; next is reviewed commit, PR publication, exact-head
  GitHub gates, merge, and separate public rollout/status synchronization.

### 2026-08-01 12:23 EEST reproducible candidate count

- Fresh package-scoped JSON evidence reports 1,185 passing Go cases: root 88,
  healthcheck 55, incident policy 53, migration 82, network policy 126,
  observability 50, token 12, topology policy 56, treasury 36, DEX 116, and
  governance 511. With the unchanged 26 Rust and eight maintained-client cases,
  the post-merge candidate is 1,219. Public status remains 1,164 until the
  implementation is merged and main is independently synchronized.

### 2026-08-01 12:36 EEST pre-ship non-reflection closure

- Sol's structured pre-ship audit found that rejected JSON keys and enum values
  could be reflected in diagnostics. Terra then found the same class in Cobra
  positional/flag errors and input-derived scenario counts. All paths now use
  generic errors and fixed synthetic report metadata; broad sentinel tests pass.
- Final independent re-review reports 0 open P0/P1. Focused incident-policy
  race/coverage PASS at 90.7%; the final complete `make verify` rerun PASS across
  all 11 packages, followed by docs consistency, CLI/JSON, and diff checks.
- Corrected fresh candidate count after the added regression: 1,186 Go cases
  (incident policy 54), plus 26 Rust and eight maintained-client cases = 1,220.
  The earlier 1,219 candidate is superseded; public status remains 1,164 until
  merge and the separate verified status sync.

### 2026-08-01 12:40 EEST GitHub publication

- Reviewed implementation commit `e03823b` is pushed and PR
  [#94](https://github.com/NeaBouli/TrueRepublic/pull/94) is open against
  `main`, closing GH-93 on merge.
- PR body records the synthetic/repository-only boundary and the complete local
  gate. Exact-head GitHub CI, review, and unresolved-thread checks are now the
  remaining merge conditions.

### 2026-08-01 12:57 EEST implementation merged; public sync in progress

- Exact PR #94 head `d3f627d` passed all 11 GitHub status gates: docs,
  build/vet/race/coverage with the real incident CLI pipeline, all eight
  recovery harnesses, Docker/Compose, Go/Rust/Node security scans, DeepScan,
  and CodeRabbit status. CodeRabbit was rate-limited and produced no content;
  the independent final Terra review recorded 0 open P0/P1 and GitHub had zero
  review threads.
- PR #94 merged as `b6e7c29`; GH-93 closed automatically. No production or
  live incident action occurred.
- Created `docs/GH-93-rollout-status` from exact merged main. Candidate public
  state is 1,220 verified cases, 12/59 overall, 12/51 phase work, and Phase 6
  5/7, while `production_ready=false` and real topology/live rehearsal/capacity
  gates remain open.
- **Status:** In Progress; verify the documentation diff, update GH-29, publish
  a separate status PR, check exact-head CI, merge, and verify Pages/main.

### 2026-08-01 13:00 EEST public status candidate locally green

- GH-29 now checks the GH-93/PR #94 runbook item while production topology and
  capacity remain unchecked.
- Public source of truth, README, landing page, rollout roadmap, wiki, agent
  state, quickstart, FAQ, and deployment boundaries consistently publish 1,220
  cases (1,186 Go + 26 Rust + eight maintained-client), 12/59 overall, 12/51
  phase work, and Phase 6 at 5/7.
- `production_ready=false`; no live rehearsal, deployment, topology, paging,
  or capacity claim was added.
- `check-consistency.sh`, status JSON parsing/arithmetic, repository contracts,
  stale-value scan, and diff hygiene PASS. Next: commit and protected docs PR.

### 2026-08-01 13:08 EEST GH-93 merged-main and public handoff verified

- Implementation PR [#94](https://github.com/NeaBouli/TrueRepublic/pull/94)
  merged exact head `d3f627d` as `b6e7c29` after all 11 required checks passed.
  Status PR [#95](https://github.com/NeaBouli/TrueRepublic/pull/95) merged exact
  head `b278a45` as `1a850d0` after all eight required checks passed.
- Exact merged-main Security Scan
  [30674576743](https://github.com/NeaBouli/TrueRepublic/actions/runs/30674576743)
  and Pages deployment
  [30674575864](https://github.com/NeaBouli/TrueRepublic/actions/runs/30674575864)
  both PASS on `1a850d0`; live Pages and `status.json` publish 1,220 verified
  cases, 12/59 overall, 12/51 phase work, and Phase 6 at 5/7.
- GH-93 is closed with every acceptance criterion checked. GH-29 marks only
  the incident/runbook gate complete; production topology and sustained-load
  capacity evidence remain open. `production_ready=false` and
  `production_deployment=false` remain explicit.
- This documentation-only branch publishes the final append-only handoff. No
  production infrastructure, live rehearsal, paging, identity, key, or chain
  mutation occurred.
- **Status:** Done after this protected handoff PR merges with exact-head gates
  green; the next repository-side Phase 6 task is capacity, disk-growth,
  retention, and sustained-load evidence.

---

## 2026-08-01 21:30 EEST GH-97 capacity qualification → In Progress

- **Branch:** `feature/GH-97-capacity-sustained-load`
- **Issue:** [GH-97](https://github.com/NeaBouli/TrueRepublic/issues/97)
- **Changed:**
  - `BRIDGE.md` / `docs/agent-bridge/ACTION_LOG.md` — task ownership and
    recovery-only scope recorded before implementation.
- **Tests:** baseline `main` `fce724e` is clean and exact with `origin/main`;
  focused and complete gates have not yet run on this new branch.
- **Risk:** Medium — the task exercises real temporary local validators and
  resource measurements, but may not alter consensus, production systems,
  real identities, keys, funds, topology, or rollout readiness.
- **Ready for:** bounded architecture analysis and Kimi review, then Sol-owned
  implementation/integration.

### Delegate notes

Kimi receives a secret-free read-only architecture review. Sol retains the
evidence contract, security boundaries, GitHub writes, diff review, complete
tests, and merge. Repository measurements must be reproducible mechanics and
regression evidence, never production sizing claims.

### Sol review feedback

Pending.

### 2026-08-01 22:11 EEST implementation checkpoint

- Kimi K3's bounded read-only review was attempted first but the provider
  returned HTTP 403 quota exhaustion; it produced no output or file change.
  Terra independently reviewed the design and recommended the implemented
  four-validator, 24-transaction, resource/retention/restart/ledger slice.
- Spark contributed only `capacitypolicy/capacitypolicy_test.go`; its focused
  package test passed. Sol reviewed that diff, fixed the command helper's
  premature buffer read, and tightened the output assertions.
- The strict contract/parser/verifier/CLI, maintained synthetic fixture,
  lifecycle registration, repository contract, CI job, operator guide, and
  real-process capacity harness now exist locally. The first real run exposed
  a CometBFT TOML section-name mismatch; Sol fixed it. A second attempt was
  blocked before execution by an exhausted local Go build cache, which was
  safely cleared without deleting repository or user data.
- The third real four-validator run has reached test execution. Focused policy
  tests passed before the final hardening edits; the full clean gates remain
  pending after the process harness finishes.
- **Risk/claim boundary:** candidate evidence is synthetic loopback regression
  data only. It proves neither multi-day soak nor Docker/Prometheus retention,
  production sizing, public topology, live infrastructure, or rollout
  readiness.

### 2026-08-01 22:47 EEST review remediation and load expansion

- Terra's final read-only review found that copied log/telemetry settings were
  configuration, not runtime-retention evidence; the evidence schema no longer
  carries those self-attested fields. The repository contract now parses
  Compose YAML and verifies finite `truerepublic-node` log settings exactly.
  Actual Docker/Prometheus retention remains explicitly open.
- The original 24 sequential transactions did not satisfy GH-97's sustained
  load criterion. The maintained candidate now defines 24 consecutive waves,
  four independent authenticated signers per wave, and 96 transactions total,
  with exact checked milli-TPS and average-block-time evidence.
- Negative tests exposed and Sol fixed a divide-by-zero panic for stalled
  manipulated evidence. Snapshot measurement now tolerates only the expected
  child-disappearance race caused by active retention while keeping root and
  other filesystem failures strict.
- Terra also found cumulative wave latency; Sol replaced it with per-submit
  timers and concurrent per-hash delivery observation. The first parallel
  attempt correctly failed before evidence because standard bank transactions
  are not registered in this app's root CLI. The corrected harness creates one
  synthetic domain per signer during setup and submits four concurrent
  authorized `add-member` transactions per workload wave.
- Kimi was retried after the expansion and again returned provider HTTP 403
  quota with no output or diff. Focused policy/root/repository/Vet/diff gates
  pass; the corrected real 96-transaction run is in progress.

---

## 2026-08-02 01:15 EEST GH-97 implementation merged → Status Sync

- **Merged:** [PR #98](https://github.com/NeaBouli/TrueRepublic/pull/98) as
  `23e0915` after every exact-head GitHub build, race/coverage, recovery,
  capacity, Docker, documentation, security, static-analysis, and review check
  passed; zero review threads remained. GH-97 closed automatically.
- **Evidence:** 96/96 authenticated transactions committed across 24 workload
  blocks on four temporary validators, with resource, retention, restart,
  app-hash, validator-power, metrics, snapshot, and bank-ledger checks green.
  Audit: 0 FAIL / 0 WARN / 7 PASS.
- **Public candidate:** `docs/GH-97-rollout-status` publishes 1,278 verified
  cases (1,244 Go + 26 Rust + 8 maintained-client), 13/59 overall, 13/51 phase
  work, and Phase 6 at 6/7. The 58-case increase is 48 capacity-policy and ten
  root cases; the real process harness remains a separate gate.
- **Boundary:** `production_ready=false`. Production sizing, multi-day soak,
  actual Docker/Prometheus retention, private-environment capacity evidence,
  live topology/deployment, external paging, and independent live operations
  remain open.
- **Next:** verify, publish, and merge this documentation-only exact-head PR;
  then update GH-29, GH-97, live Pages, and the final Bridge handoff.

### 2026-08-02 01:20 EEST public status PR checkpoint

- [PR #99](https://github.com/NeaBouli/TrueRepublic/pull/99) is open from
  `docs/GH-97-rollout-status`. Documentation consistency, full package-scoped
  1,244-Go-case recount, arithmetic checks, diff hygiene, and two independent
  read-only reviews pass with no P0/P1/P2 finding.
- The authoritative candidate is 1,278 total cases, 13/59 overall, 13/51 phase
  work, and Phase 6 at 6/7. PR #99 exact-head security/review checks and merge
  remain pending; no production or live-environment action occurred.

### 2026-08-02 01:31 EEST GH-97 final handoff → Done

- PR #99 exact head `047f6e1` passed documentation consistency, Go/Rust/Node
  security, DeepScan, CodeRabbit, and independent read-only review with no open
  thread or P0/P1/P2 finding. It merged as `fe6e693`.
- GH-97 is closed with every acceptance criterion checked. GH-29 records the
  repository-side capacity gate complete and remains open for later rollout
  work. The merged status is 1,278 cases, 13/59 overall, 13/51 phase work, and
  Phase 6 at 6/7; `production_ready=false` remains unchanged.
- GitHub Pages deployment `30699752124` passed. Live `status.json` and the
  landing page were read back successfully with the exact merged totals,
  percentages, capacity marker, and false production-ready flag.
- Production sizing, multi-day soak, actual Docker/Prometheus runtime
  retention, private-environment capacity evidence, live topology/deployment,
  external paging, and independent live operations remain open. No production,
  deployment, key, identity, fund, or infrastructure action occurred.
- **Status:** Done. The foundation is clean for the next bounded GH-29 task.

---

## 2026-08-04 12:25 EEST GH-102 maintained-client high audit gate → Local Pass

- **Changed:** only `client-web/package.json` and its lockfile update the
  pre-existing `brace-expansion` override from 5.0.8 to compatible patch 5.0.9.
- **Root cause:** all eleven high findings cascaded through the dev-only
  ESLint/TypeScript-ESLint/minimatch tree from that vulnerable override.
- **Passed:** clean `npm ci`, lint, eight Vitest cases, production build, and
  required `npm audit --audit-level=high` (exit 0). Only two moderate React
  Router advisories remain; their offered fix is a separately reviewed major
  v6→v7 migration, not a forced audit rewrite.
- **Boundary:** no runtime source, routing, wallet, signing, network, key, fund,
  server, deployment, or production change. Mobile legacy advisories stay open
  under GH-102.
- **Next:** independent read-only review, commit/protected PR/CI/merge, then
  rebase and rerun GH-101 PR #103.

### Independent review

- Kimi K3: APPROVED; no P0/P1 finding and no blocking diff defect. Lock
  integrity, dev-only dependency closure, minimal 5.0.9 patch, live high-level
  audit exit 0, lint, and 8/8 tests were independently confirmed.
- Review note: do not use `npm audit --offline` as release evidence because
  pruned advisory-cache bodies can yield a false empty result. Repository CI
  already uses the required live audit, unchanged.

---

## 2026-08-04 10:20 EEST GH-101 private topology evidence gate → In Progress

- **Branch:** `feature/GH-101-private-topology-evidence`
- **Issue:** [GH-101](https://github.com/NeaBouli/TrueRepublic/issues/101)
- **Changed:** task and non-production scope recorded before implementation.
- **Tests:** baseline `origin/main` is exact at `3f1edd1`; latest scheduled
  Security Scan `30790264230` passed. Focused GH-101 gates have not run yet.
- **Risk:** Medium — the package will validate deployment evidence and may
  affect future rollout decisions, but this task is repository-only and may
  not inspect or mutate any real infrastructure, inventory, identity, key,
  firewall, DNS, TLS, provider, funds, or production state.
- **Ready for:** bounded Kimi architecture review, Sol implementation and
  integration, focused/full verification, independent review, PR, and merge.

### Delegate notes

Kimi receives only the public repository and a secret-free bounded request.
The evidence verifier must bind a private topology contract by digest and
cross-check it only when supplied locally; public reports may contain counts,
gate names, digests, time bounds, and pass/fail state, never inventory values.
Passing the repository fixture must not close GH-29's live Phase 6 checkbox.

### Sol review feedback

Kimi's implementation was reviewed line by line. Sol changed report counters to
derive from trusted policy and topology rather than untrusted manifest values,
then added a repository contract test covering fixture binding, secret/inventory
exclusion, CI wiring, and documentation boundaries.

### 2026-08-04 10:47 EEST implementation checkpoint

- **Changed:** strict `deploymentevidence` parser/verifier/CLI, synthetic
  digest-bound fixture, root-command and CI integration, operator guide,
  limitations, and repository contract coverage.
- **Kimi contribution:** secret-free architecture review and bounded package
  implementation. Kimi made no commit or external change; Sol owns and reviewed
  every diff.
- **Passed:** package tests, package Vet, focused root/repository tests, root
  Vet, maintained CLI/JQ verification, workflow YAML parsing, gofmt, and diff
  hygiene.
- **Boundary:** the verifier proves envelope structure, freshness, digest
  binding, and claimed logical-seat separation only. It does not inspect source
  artifacts or prove real infrastructure, operators, observations, or rollout.
  Phase 6 remains 6/7 and overall rollout remains 13/59.
- **Next:** full repository verification, independent final review, audit,
  exact-head PR/CI, merge, and public handoff.

### 2026-08-04 12:02 EEST local release gate

- Full Go build/Vet and all 13 packages with Race/Coverage pass; root is 69.1%
  and `deploymentevidence` is 90.8%. Docs consistency, deterministic CLI/JQ,
  workflow YAML, formatting, and diff checks pass.
- Rust format/Clippy/build and 26 tests pass. Web lint/eight tests/build and the
  mobile pass-with-no-tests workflow contract pass.
- Kimi final read-only review: APPROVED, no P0/P1/P2. Audit: 0 FAIL / 0 WARN /
  7 PASS in `docs/agent-bridge/GH101_AUDIT.md`.
- Claude Code was assigned the small read-only test-count/status-arithmetic
  cross-check requested by Gio, but its local OAuth session had expired and
  could not refresh; it produced no output or diff. Sol's measured JSON recount
  is 71 new package cases plus one root case, or 72 new Go cases total.
- Separate observation: current unchanged client dependencies expose audit
  advisories; [GH-102](https://github.com/NeaBouli/TrueRepublic/issues/102)
  tracks the dedicated remediation and is not folded into GH-101.
- **Status:** locally complete; protected PR/exact-head CI/merge pending.
  Rollout remains 13/59 and Phase 6 remains 6/7.

---

## 2026-08-04 12:47 EEST GH-101 exact-head review remediation → In Progress

- **Review:** two CodeRabbit threads were verified against exact head
  `7c8f087`. The CI-ordering finding was valid; the date finding was not,
  because the audit was performed on 4 August in the documented EEST project
  timezone while GitHub still displayed 3 August UTC.
- **Changed:** `.github/workflows/go-ci.yml` now derives
  `./deploymentevidence` from `scripts/go-packages.sh` and runs its focused
  tests immediately before the maintained manifest gate.
- **Tests:** focused package test PASS, workflow YAML parse PASS, and
  `git diff --check` PASS.
- **Claude contribution:** Claude Code received this small, bounded one-file
  task but its OAuth session remained expired; it produced no output or diff.
  Sol applied and verified the minimal fix under the documented fallback.
- **Risk/blocker:** no implementation blocker. Exact-head CI must restart for
  the new commit, and both review threads must be answered/resolved before
  merge. No live or production action occurred.
- **Next:** commit/push the remediation, resolve the verified review threads,
  pass the refreshed exact-head checks, merge PR #103, and synchronize public
  rollout status without closing the live Phase 6 topology checkbox.

---

## 2026-08-04 13:15 EEST GH-101 implementation merged → Status Sync

- **Merged:** [PR #103](https://github.com/NeaBouli/TrueRepublic/pull/103) as
  `7924792`; GH-101 closed automatically. Exact head `e7d016a` passed every
  required build, full/race test, recovery, capacity, Docker, docs, dependency,
  static-analysis, and review status, with zero unresolved review threads.
- **Public candidate:** `docs/GH-101-rollout-status` starts from exact merged
  `origin/main` and will publish 1,350 verified cases (1,316 Go + 26 Rust + 8
  maintained-client). GH-101 contributes 71 `deploymentevidence` cases and one
  root repository-contract case.
- **Kimi contribution:** independent read-only claim census and arithmetic
  review identified every gated and non-gated public surface; no file was
  changed by Kimi. Claude Code remained OAuth-unavailable and produced no diff.
- **Boundary:** rollout stays 13/59 overall, 13/51 phase work, Phase 6 at 6/7,
  and `production_ready=false`. The verifier proves only a secret-free evidence
  envelope; no private/live deployment or infrastructure evidence exists.
- **Next:** synchronize repository/public status and GH-29, validate exact
  arithmetic and docs, publish/merge the protected docs PR, verify Pages
  readback, then append the final handoff.

---

## 2026-08-04 13:26 EEST GH-101 public sync → Local Pass

- **Changed:** public totals are synchronized at 1,350 (1,316 Go + 26 Rust +
  8 maintained-client), with root 99 and `deploymentevidence` 71. README,
  status JSON, landing page, roadmap, FAQ, quickstart, wiki, project state,
  TODO, action log, and the consistency gate now agree.
- **Kimi review:** found one premature GH-101 completion checkbox and two P2
  consistency/document-label gaps. Sol reopened the pending publish/sync item,
  extended the gate across all current count surfaces and every wiki module
  row, and renamed the mixed PR list accurately.
- **Tests:** docs consistency PASS; status/module/rollout arithmetic PASS;
  shell syntax PASS; diff hygiene PASS. Kimi independently reproduced 99 root
  and 71 deployment-evidence passing cases, 90.8% deployment-evidence
  coverage, eight client cases, and the two remaining moderate React Router
  advisories with zero high/critical maintained-client finding.
- **Boundary:** rollout remains 13/59, phase work 13/51, Phase 6 6/7, and
  `production_ready=false`; live topology and deployment evidence remain open.
- **Next:** commit/push the docs candidate, open and verify the protected PR,
  synchronize GH-29, then mark the still-open publish/sync checkbox complete
  only when its conditions are actually satisfied.

---

## 2026-08-04 13:32 EEST PR #105 review remediation → In Progress

- **GitHub:** first exact head `3981193` passed docs, Go/Rust/Node security,
  DeepScan, and maintained-client high-level audit gates. GH-29 now links
  GH-101/PR #103 while its production-topology checkbox remains open.
- **Review:** CodeRabbit's two valid minor findings were remediated: project
  state now carries the real post-sync UTC timestamp, and each added count
  check is tied to its specific documentation field rather than any matching
  historical number elsewhere in the file.
- **Tests:** hardened documentation consistency PASS, shell syntax PASS, and
  diff hygiene PASS.
- **Next:** push the remediation, resolve both threads, rerun the exact-head
  checks, then complete the still-open publish/sync item and merge only after
  the refreshed head is fully green.

---

## 2026-08-04 13:37 EEST PR #105 exact head → Ready to Merge

- **Head:** `1c7e34c` passed every applicable docs, Go/Rust/Node security,
  maintained-client audit, DeepScan, and review status. Both CodeRabbit threads
  are resolved; zero failed or pending checks remain.
- **Synchronization:** GH-29 links GH-101/PR #103, the project handoff is
  current through this checkpoint, and the live production-topology item
  remains unchecked. The GH-101 publish/sync TODO is now complete.
- **Boundary:** 1,350 verified cases; rollout 13/59, phase work 13/51, Phase 6
  6/7, and `production_ready=false` remain exact.
- **Next:** run this final bookkeeping head through its exact checks, merge PR
  #105, verify Pages/live status, and record the post-merge handoff.

---

## 2026-08-04 13:46 EEST GH-101 public synchronization → Done

- **Merged:** implementation PR #103 as `7924792`; public status PR #105 as
  `26cd19e`. GH-101 is closed with every acceptance criterion checked, and
  GH-29 links the repository evidence while its live topology item stays open.
- **GitHub evidence:** both PRs passed their exact-head required checks with
  zero unresolved review threads. Post-merge Security Scan `30866462881` and
  Pages deployment `30866461897` pass.
- **Live readback:** GitHub Pages publishes 1,350 cases (1,316 Go + 26 Rust +
  8 maintained-client), rollout 13/59, phase work 13/51, Phase 6 6/7,
  `production_ready=false`, and the GH-101 offline evidence-gate marker.
- **Review/delegation:** Kimi's independent census, count reproduction, and
  final review findings were integrated and reverified. Claude Code received
  two small bounded assignments but remained OAuth-unavailable and produced no
  output or diff.
- **Residual:** GH-102 remains open for two moderate React Router advisories and
  legacy mobile migration/retirement. Real private topology deployment,
  inventory, DNS/TLS, firewall, provider, and live evidence remain separately
  authorized work.
- **Status:** Done. No infrastructure, key, identity, fund, deployment, or
  production mutation occurred.

---

## 2026-08-04 13:55 EEST GH-102 React Router remediation → In Progress

- **Branch:** `fix/GH-102-react-router-v7` from exact clean `origin/main`
  `fe2ed26`; issue [GH-102](https://github.com/NeaBouli/TrueRepublic/issues/102).
- **Scope:** maintained `client-web` only for this slice: classify and remove
  the two remaining moderate React Router advisories through a reviewed v7
  migration, preserve routes/navigation/wallet/signing/chain behavior, add
  focused router coverage, and pass clean install/lint/tests/build/live audit.
- **Delegation:** Kimi receives a bounded secret-free implementation block.
  Sol retains architecture, diff review, security, integration, complete tests,
  GitHub actions, and closure. Claude Code may receive only a small focused
  helper check if its expired OAuth session becomes available.
- **Boundary:** repository code/dependency/test/docs only. No wallet key,
  account, signing, broadcast, funds, mobile release, deployment, or production
  action. Legacy mobile remediation remains a separate GH-102 slice.
- **Status:** baseline and migration review pending; no file changed yet.

---

## 2026-08-04 14:12 EEST GH-102 React Router migration → Verification

- **Kimi contribution:** migrated `client-web` from React Router 6.30.4 to
  7.18.2, extracted the exact 21-route contract, and added focused coverage for
  redirects, static/parameterized routes, search parameters, and backslash-shaped
  navigation input. Kimi did not touch Bridge, GitHub, wallet/signing code, or
  external systems.
- **Live advisory change:** `npm audit` now reports only
  `GHSA-qwww-vcr4-c8h2`, published 2026-07-24 after the original 7.18.2 target
  was selected. The patched `react-router` 8.3.0 requires React/React DOM
  19.2.7+ and `react-router-dom` has no v8 release, so no React-18-compatible
  patched version exists.
- **Sol decision:** accept this single advisory temporarily for the client-only
  BrowserRouter SPA because it has no RSC, server actions, route actions, data
  router, or server runtime. Added a fail-closed audit gate that accepts only the
  exact advisory and fails every other high/critical or malformed result. Owner,
  deadline (2026-09-04 or pre-rollout), rationale, and removal condition are in
  `docs/developers/DEPENDENCY_RISK_ACCEPTANCE.md`.
- **Changed files:** `client-web/package{,-lock}.json`, `client-web/src/App.tsx`,
  new `client-web/src/routes.tsx` and `routes.test.tsx`, new audit policy and Node
  policy tests under `client-web/scripts/`, both client audit workflow steps,
  and the dependency-risk record.
- **Verification:** `npm run lint` passed; Node policy tests 6/6 passed; Vitest
  40/40 passed across 5 files; `npm run build` passed; `npm run audit:high`
  passed with only the exact documented RSC advisory accepted; `git diff
  --check` passed. Independent review, protected PR CI, and GitHub/Bridge
  publication remain pending.

---

## 2026-08-04 14:19 EEST GH-102 review remediation → PR Ready

- Kimi's independent review found no high or medium issue and approved the
  migration after protected PR gates. Sol resolved its process and hardening
  observations before publication: added the ACTION_LOG entry and developer-doc
  index link, clarified the affected dependency chain, bound the exception to
  the `react-router` package, and made CI reject router APIs outside the reviewed
  declarative SPA surface.
- **Final local results:** audit-policy/boundary tests 6/6, Vitest 40/40 across
  five files, lint PASS, TypeScript/Vite production build PASS, guarded live npm
  audit PASS with only `GHSA-qwww-vcr4-c8h2`, and diff hygiene PASS.
- Protected PR exact-head CI/review, merge, GH-102 checkpoint, and final
  Bridge/project-state synchronization remain. Legacy mobile remediation stays
  open as the next separate GH-102 slice.

---

### 2026-08-04 14:27 EEST final audit-boundary hardening

- Replaced the provisional text matcher with the installed TypeScript parser.
  The gate now covers named/default/namespace imports, router re-exports,
  wildcard exports, static/dynamic imports, and CommonJS `require`, while
  allowing only the reviewed declarative-SPA API set.
- Kimi's focused final review approved the exact final script/test state with no
  blocking finding. Sol reran policy tests 6/6, Vitest 40/40, lint, build, live
  guarded audit, documentation consistency, and diff hygiene successfully.

### 2026-08-04 14:36 EEST PR #107 review remediation

- Exact head `3b5d47e` passed Client Web CI, Docs, DeepScan, Go/Rust security,
  and all Node audit jobs. CodeRabbit raised one valid boundary finding: a
  variable-backed dynamic module import could evade static module attribution.
- The gate now rejects every computed `import()`/`require()` specifier before
  module classification, with a regression case in the six-test policy suite.
  Lint, policy tests 6/6, Vitest 40/40, production build, guarded live audit,
  docs consistency, and diff hygiene pass after the patch. Refreshed exact-head
  CI/review and thread resolution remain pending.

## 2026-08-04 14:43 EEST GH-102 maintained-client router slice → Merged

- PR #107 final head `694c976` passed every applicable GitHub check with its
  sole review thread resolved and merged as `06f1125`.
- GH-102 and its issue evidence now show both maintained-client slices complete;
  the issue remains open only for legacy mobile migration/retirement and final
  repository-wide closure.
- Public recovery evidence is synchronized to 1,388 cases (1,316 Go + 26 Rust +
  46 maintained client). Source-backed client inventory is 42 component files,
  21 routes, eight production stores, 11 production services, and a 317.94 kB
  gzip bundle. Rollout stays
  13/59 overall, 13/51 phase work, Phase 6 6/7, `production_ready=false`.
- Final handoff branch updates status/docs/Bridge only. No production, mobile
  release, wallet, signing, broadcast, fund, or infrastructure action occurred.

### 2026-08-04 14:50 EEST handoff review remediation

- Kimi independently reproduced 1,316 + 26 + 46 = 1,388, Vitest 40/40,
  audit-policy 6/6, 42 component files, 21 routes, and the unchanged rollout
  boundary. Its semantic inventory finding was corrected by excluding test files:
  eight production stores and 11 production services.
- Root audit bundle/test claims are synchronized. Client inventory fields now
  require positive integers, and a zero-route result reports a controlled
  fail-closed inconsistency instead of aborting before the summary.

---

## 2026-08-04 14:56 EEST GH-102 public handoff → Complete

- Public status PR #108 exact head `4402577` passed Docs Consistency, DeepScan,
  CodeRabbit, Go/Rust security, and every Node audit job with zero review threads;
  it merged as `d1783ec`.
- Post-merge Pages run `30870238036` and Security Scan `30870238562` pass. Live
  `status.json` readback confirms 1,388 cases (1,316 Go + 26 Rust + 46 client),
  42 component files, 21 routes, eight production stores, 11 production services,
  317.94 kB gzip, rollout 13/59, Phase 6 6/7, and `production_ready=false`.
- GH-102 is synchronized and remains open only for legacy mobile migration or
  retirement plus final repository-wide closure. Maintained-client router work
  is Done. No production/runtime, wallet, signing, broadcast, fund, app-store,
  deployment, infrastructure, or provider mutation occurred.

---

## 2026-08-06 01:38 EEST GH-110 RustSec rkyv unblock → In Progress

- **Branch:** `fix/GH-110-rkyv` from exact `origin/main` `d6ebdf5`; issue
  [GH-110](https://github.com/NeaBouli/TrueRepublic/issues/110).
- **Trigger:** GH-102 full verification passed Rust format, clippy, build, and
  all 26 workspace tests, then the refreshed RustSec database rejected
  transitive `rkyv` 0.8.15 for RUSTSEC-2026-0233, -0234, and -0235.
- **Scope:** smallest compatible lockfile-only update to the patched `rkyv`
  line, full Rust verification, protected PR, and merge before GH-102 resumes.
- **Boundary:** no contract behavior, source API, keys, signing, funds,
  deployment, infrastructure, or production action.

---

## 2026-08-06 01:43 EEST GH-110 local verification → Complete

- Updated only the compatible lockfile line: `rkyv` and `rkyv_derive` 0.8.15 →
  0.8.17. No Rust source, contract API, or behavior changed.
- PASS: `cargo fmt --all -- --check`, workspace clippy for all targets with
  warnings denied, workspace build, all 26 workspace tests, and `cargo audit`.
  The three `rkyv` vulnerabilities are gone; five pre-existing allowed
  unmaintained/unsound warnings remain visible.
- Diff is limited to ten lockfile lines plus append-only Bridge evidence.
  Protected GitHub PR checks and merge remain before GH-102 resumes.

---

## 2026-08-06 01:09 EEST GH-102 legacy mobile dependency block → In Progress

- **Branch:** `fix/GH-102-legacy-mobile` from exact clean `origin/main`
  `d6ebdf5`; issue [GH-102](https://github.com/NeaBouli/TrueRepublic/issues/102).
- **Scope:** reproduce and classify the current `mobile-wallet` dependency and
  advisory inventory, then choose the smallest evidence-backed safe outcome:
  a compatible migration or explicit retirement. Preserve all maintained-client,
  wallet/signing, chain/API, and rollout boundaries.
- **Baseline:** no open PR; latest `main` Pages `30870685980` and Security Scan
  `30870686809` pass. The issue remains open only for this legacy-mobile slice
  and final repository-wide closure.
- **Delegation:** Kimi K3 receives one bounded, secret-free repo analysis or
  implementation block after the baseline is reproduced. Claude Code may receive
  only a small focused helper check if its authentication is available. Sol owns
  architecture, security judgment, all diffs, complete verification, GitHub
  writes, merge, and closure.
- **Boundary:** repository code, dependencies, tests, CI, and documentation only.
  No keys, mnemonics, accounts, signing, broadcast, funds, app-store action,
  deployment, infrastructure, or production mutation.
- **Next:** reproduce exact clean-install/audit/build behavior; inventory mobile
  imports, crypto/key storage, testability, Expo/CosmJS compatibility, and CI;
  then record the migration-versus-retirement decision before implementation.

---

## 2026-08-06 01:50 EEST GH-102 legacy mobile retirement → Locally Complete

- **Decision:** explicitly retired and removed the unsafe prototype and its
  pass-with-no-tests workflow. A blocking CI retirement contract keeps the
  source, workflow, stale status, and run instructions absent; Git history is
  audit-only. Any future native client is a separately reviewed rebuild.
- **Baseline evidence:** 51 advisories (7 low, 16 moderate, 24 high, 4 critical),
  no tests, two Expo Doctor failures, Android bundle failure at CosmJS Node
  `crypto`, mnemonic retained in UI state, live signing/broadcast code, obsolete
  governance/DEX queries, and a swap stub.
- **Verification:** mobile retirement contract, documentation consistency,
  workflow YAML/status JSON parsing, shell syntax, and diff checks pass. Full Go
  `make build && make verify` passes; maintained `client-web` lint, 6 audit-policy
  tests, 40 Vitest cases, build, and guarded live audit pass; Rust format, clippy,
  build, and all 26 tests pass.
- **Security prerequisite:** the refreshed RustSec database exposed three new
  transitive `rkyv` findings. GH-110/PR #111 updated only the compatible lockfile
  line to 0.8.17; all protected checks passed and it merged as `002299e`.
- **Independent review:** Kimi found no P0-P2 or blocking issue. Its single dead
  `.gitignore` P3 was removed and rechecked. Claude Code's bounded documentation
  census returned no output, was stopped, and made no diff.
- **Separate inherited debt:** legacy `web-wallet` tests/build pass, while the
  live audit reports 70 advisories (18 low, 20 moderate, 29 high, 3 critical).
  GH-112 tracks migration or retirement; this does not weaken or block the
  mobile-retirement decision.
- **Boundary:** no key, mnemonic, account, signing, broadcast, fund, app-store,
  deployment, infrastructure, or production action occurred. Protected GH-102
  PR checks, merge, final issue/status sync, and final-main verification remain.

---

## 2026-08-06 01:54 EEST GH-102 exact-head preparation → Complete

- Rebased the retirement commit onto merged prerequisite main `002299e` and
  preserved both append-only GH-110 and GH-102 Bridge histories.
- Post-rebase PASS: main ancestry, clean diff, retirement contract, docs
  consistency, workflow YAML/status JSON parsing, shell syntax, and live Rust
  audit with zero vulnerabilities and five visible warning-only advisories.
- Branch is ready for push and protected PR checks. No runtime, deployment,
  infrastructure, wallet, key, signing, broadcast, fund, or production action.

---

## 2026-08-06 02:01 EEST GH-102 legacy mobile retirement → Complete

- PR [#113](https://github.com/NeaBouli/TrueRepublic/pull/113) exact head
  `e606db9` passed docs consistency, mobile retirement, Go/Rust security,
  maintained-client and legacy-web audit jobs, DeepScan, and the available
  CodeRabbit gate with no review thread. It merged as `d621ff9`; GH-102 closed.
- Final-main Pages `31054458625` and Security Scan `31054461930` pass. Live
  `status.json` confirms 1,388 cases, rollout 13/59, Phase 6 6/7,
  `production_ready=false`, maintained `client-web`, deprecated `web-wallet`,
  and retired/removed `mobile-wallet` with Git-history-only status.
- GH-112 separately tracks the inherited legacy-web migration/retirement debt.
  There is no supported native mobile client; any replacement is a separately
  reviewed high-risk implementation.
- No key, signing, broadcast, fund, app-store, deployment, infrastructure, or
  production action occurred. GH-102 is Done.

---

## 2026-08-06 03:14 EEST GH-112 legacy web wallet → In Progress

- **Branch:** `fix/GH-112-legacy-web` from exact clean `origin/main` `1a92f0a`.
- **Issue:** [GH-112](https://github.com/NeaBouli/TrueRepublic/issues/112).
- **Scope:** reproduce and classify the deprecated `web-wallet` dependency,
  crypto/key/signing, chain-query, build, test, and release surfaces; then
  migrate only if it can meet the canonical-client security boundary, otherwise
  explicitly retire and remove it with a blocking absence contract.
- **Risk:** High because the tree contains wallet/cryptography code. Repository
  evidence only: no real keys, accounts, signing, broadcast, funds, deployment,
  infrastructure, or production action.
- **Delegation:** Kimi K3 receives a bounded secret-free deep analysis. Claude
  Code may receive one small helper census if available. Sol owns the decision,
  every diff, complete verification, GitHub writes, merge, and closure.
- **Baseline:** no open PR; final-main Pages `31054970470` and Security Scan
  `31054971793` pass. `client-web` remains the sole maintained client.
- **Next:** reproduce exact clean install, tests, build, live audit, dependency
  paths, wallet/key behavior, chain compatibility, CI, and public references.

---

## 2026-08-06 03:49 EEST GH-112 legacy web retirement → Locally Complete

- **Decision/evidence:** explicit retirement selected. The final clean legacy
  baseline installed 1,426 packages, reproduced 70 advisories (18 low, 20
  moderate, 29 high, 3 critical), passed 18 API-isolated tests, and built with
  extensive source-map warnings. Review proved broken `queryAbci` calls,
  unregistered/nonexistent custom messages, a chain-rejected unsafe legacy swap,
  and a functional bank-send path.
- **Implementation:** removed the complete `web-wallet` tree and non-blocking
  audit; added a blocking retirement contract; containerized maintained
  `client-web` with non-root nginx; moved Compose/root-nginx/CI to the canonical
  client; and added same-origin loopback-only RPC/REST proxy configuration plus
  a repository policy regression.
- **Public state:** synchronized current install, developer, user, operator,
  security, wiki, roadmap, agent, status, and Pages sources. The verified total
  is 1,389 (1,316 Go + 26 Rust + 47 maintained-client), rollout is 14/59 and
  Phase 6 correctly remains 6/7, `production_ready=false`, and current bundle
  size is 318.01 kB gzip.
- **Verification PASS:** maintained-client lint, 6 audit-policy tests, 41
  Vitest cases, production build, and guarded live audit; web/mobile retirement
  contracts; docs consistency; JSON/YAML/shell/diff checks; focused container
  policy test; complete `make build && make verify`; Rust format, all-target
  clippy with warnings denied, workspace build, all 26 tests, and live audit
  with five visible warning-only advisories.
- **Delegation:** Kimi supplied the independent baseline/security review and the
  bounded container/Compose/nginx/CI implementation; Sol reviewed and extended
  every diff and reran all gates. Claude Code's small read-only census identified
  the exact runtime/CI/docs retirement surfaces and made no diff.
- **Follow-ups:** [GH-115](https://github.com/NeaBouli/TrueRepublic/issues/115)
  tracks maintained-client custom transaction registration/delivery; [GH-116](https://github.com/NeaBouli/TrueRepublic/issues/116)
  tracks the now-consumerless legacy `custom/...` query shim.
- **Pending:** independent final diff review, protected PR Docker/runtime and
  exact-head checks, merge, issue/status closure, and final-main verification.
  Local Docker/nginx binaries were unavailable, so no local image/runtime claim
  is made.
- **Safety:** no real key, mnemonic, account, signing, broadcast, funds,
  deployment, infrastructure, or production action occurred.

---

## 2026-08-06 03:55 EEST GH-112 legacy web retirement → Review Complete / PR Ready

- **Independent review:** Kimi's fresh read-only diff review found no P0/P1 and
  no blocking defect. The stale Keplr-only user path it identified was removed
  from current user, wiki, architecture, and integration guidance; reference
  examples are now explicitly local/test-only.
- **Hardening:** the direct loopback client proxy now applies the same bounded
  RPC/REST request rates and metrics denial as the root proxy. The repository
  policy test asserts these controls.
- **Reverification PASS:** focused container-network policy test, both client
  retirement contracts, documentation consistency, workflow YAML parse, shell
  syntax, and diff integrity. The earlier complete Go, Rust, and maintained
  client gates remain valid; later changes were limited to documentation, nginx
  policy, and its focused Go regression.
- **Remaining external gate:** Docker is unavailable locally, so the client
  image and shared-namespace runtime must be proven by the protected GitHub
  Compose job. No production or external runtime action occurred.
- **Next:** commit, publish the ready PR, wait for exact-head protected checks,
  merge, and verify final-main status.

---

## 2026-08-06 04:14 EEST GH-112 PR #117 → Review Remediation

- **Protected first head:** all 13 checks passed on `e8f92fb`, including the
  previously unavailable Docker/Compose runtime, client image, direct RPC/REST
  probes, complete Go/recovery/capacity jobs, security scans, DeepScan, and
  CodeRabbit.
- **Review handling:** verified and addressed all nine CodeRabbit inline
  findings: absolute endpoint normalization, non-watch contributor command,
  recovery-only signer language, fail-closed ZKP status, completed retirement
  state, executable local routes/setup, and Markdown fence hygiene.
- **Local revalidation PASS:** 41 Vitest plus six audit-policy cases, lint,
  Vite build, guarded audit, focused Go policy, both retirement contracts,
  documentation consistency, YAML/shell, and diff integrity. Vite now reports
  the current main bundle as 318.02 kB gzip; no test-count change.
- **Next:** publish the review commit and require the complete exact-head check
  set to return green before merge.

---

## 2026-08-06 04:44 EEST GH-112 legacy web retirement → Complete

- **GitHub:** PR [#117](https://github.com/NeaBouli/TrueRepublic/pull/117)
  merged as `2d776b5369882a0e180a7ef0dad187016efc7d33`; GH-112 closed and its
  feature branch was removed. All nine review threads are resolved.
- **Final-main PASS:** Go CI `31062895318`, Client Web CI `31062895236`,
  Security Scan `31062895275`, and Pages `31062894585` all completed
  successfully on the merge commit.
- **Published state verified:** live Pages reports 1,389 recovery tests, rollout
  14/59, Phase 6 6/7, maintained-client bundle 318.02 kB gzip, and
  `production_ready=false`.
- **Outcome:** unsafe legacy `web-wallet` is absent and guarded by a blocking
  retirement contract; maintained `client-web` is the sole repository web
  client and its loopback container/proxy path passed protected runtime CI.
- **Next:** GH-115 is the next implementation target; GH-116 remains the
  protocol-cleanup decision. No production action or real signing/funds flow
  occurred.

---

## 2026-08-06 12:10 EEST GH-115 canonical custom transactions → In Progress

- **Branch:** `feature/GH-115-custom-tx` from exact `origin/main` `05fb5b7`.
- **Issue:** [GH-115](https://github.com/NeaBouli/TrueRepublic/issues/115).
- **Scope:** inventory and reconcile every maintained-client custom transaction
  with the chain's registered protobuf identity and field encoding; introduce
  one fail-closed registry/client boundary; and add deterministic encoding plus
  secret-free local delivery evidence for supported transaction families.
- **Risk:** High. This touches wallet signing, custom protobuf messages, DEX
  amounts/slippage, membership/governance/admin authority, and chain delivery.
  Sol owns architecture, security, every diff, external writes, and closure.
- **Delegation:** Kimi K3 receives one bounded secret-free implementation block
  after Sol completes the source/type inventory. Kimi may not delegate, access
  external systems, or use real keys/accounts. Claude Code remains limited to a
  later small helper test/docs census if useful.
- **Safety boundary:** repository and ephemeral local-test work only. No real
  key, mnemonic, account, signing, broadcast, funds, deployment,
  infrastructure, mainnet, or production action is authorized or performed.
- **Next:** map all client `typeUrl` values to Go message registration and
  existing tx harnesses, then freeze the canonical registry design before code.

---

## 2026-08-06 13:18 EEST GH-115 canonical custom transactions → Local Verification Complete

- **Implementation:** one exact registry now owns all 11 maintained custom
  messages; all signing paths share fail-closed simulation, conservative
  config-priced fees, and delivery handling. Canonical addresses use
  `truerepublic`, and DEX amounts are validated as positive signed-int64 decimal
  strings inside the registry encoder.
- **Chain compatibility:** added Go/TypeScript byte-for-byte vectors, complete
  hand-written DEX protobuf descriptors, and dynamic-message signer extraction
  required by Cosmos SDK v0.50 transaction decoding.
- **Real local evidence PASS:** a temporary loopback-only chain with generated
  disposable accounts simulated, signed, delivered, and committed bank,
  governance, membership, identity, and slippage-bounded DEX flows. Expected
  approval failure and unsupported legacy swap paths fail closed; all temporary
  state is removed.
- **Full gates PASS:** 1,329 Go cases with build/vet/race/coverage; 82 Vitest +
  six audit-policy cases; lint; build (319.32 kB gzip); guarded audit; Rust
  format/clippy/build/26 tests/audit; docs/retirement/JSON/YAML/diff checks.
- **Status:** 1,443 standard cases, rollout 16/59, phase work 16/51, Phase 6
  6/7, `production_ready=false`. The real client-chain case is separately gated
  and not added to the standard arithmetic.
- **Delegation:** Kimi implemented the bounded registry/service block and is
  completing a read-only final review. Sol reviewed and extended every diff.
  Claude Code was attempted twice for a small read-only census but its budget
  prevented execution; no output or diff resulted.
- **Pending:** final independent review, protected PR checks, merge, GH-115 and
  GH-29 synchronization, Pages readback, and final-main verification. No real
  keys/accounts/funds, public broadcast, deployment, infrastructure, or
  production action occurred.

---

## 2026-08-06 13:31 EEST GH-115 canonical custom transactions → Review Complete / PR Ready

- **Independent review:** Kimi reproduced registry/wire, signer, client, lint,
  typecheck, focused Go, and real local-chain evidence. No P0/P1/P2 was found;
  approval was recommended after correcting two P3 observations.
- **P3 remediation:** public arithmetic now exactly records 82 standard Vitest
  plus six audit-policy cases (88 maintained-client; 1,443 overall), while the
  real chain case remains separately gated. The latest production bundle is
  319.32 kB gzip.
- **Cleanup hardening:** setup failures now terminate any started node and
  remove the temporary home before rethrowing. A successful full delivery run
  and a deliberately failed `/bin/false` init both leave zero GH-115 temporary
  directories and zero node processes.
- **Status:** local implementation and review are complete. Protected PR CI,
  review threads, merge, GH-115/GH-29 synchronization, Pages readback, and
  final-main verification remain before Done. Production remains false.

---

## 2026-08-06 14:13 EEST GH-115 → PR Review Remediation Verified

- **PR state:** PR #119 is published. Its original exact head passed every
  protected workflow; 16 subsequent CodeRabbit review threads were triaged.
- **Valid fixes:** lossless legacy-prefix wallet migration, bounded RPC
  readiness, connection-error result handling, delivery-code fallback,
  read-only query client, codec-owned type URLs, explicit seven-message DEX
  signer contracts, protobuf wire-tag validation, and synchronized status/docs.
- **Evidence:** `make verify` PASS (1,329 Go cases plus build/vet/race/coverage),
  85 Vitest + six audit cases, lint/build/audit PASS, real disposable-chain
  delivery PASS in 70.81 seconds, docs/retirement/JSON/diff checks PASS, and
  zero leaked GH-115 temporary homes.
- **Independent review:** Kimi found no P0/P1/P2. Sol corrected all remaining
  P3/P4 status, wording, and assertion observations; Kimi made no write.
- **Status:** 1,446 standard cases, rollout 16/59, phase work 16/51, Phase 6
  6/7, `production_ready=false`. Next: commit/push remediation, respond to and
  resolve all 16 threads, await exact-head CI, merge, close GH-115, update
  GH-29, verify Pages and final `main`.

---

## 2026-08-06 14:46 EEST GH-115 → Merged / Final Main Green / Done

- **Merge:** PR #119 merged as `5d9e946`; GH-115 is closed. All 16 review
  threads were resolved before the authorized administrative merge.
- **Final `main`:** Client Web CI `31097605393`, Go CI `31097605358`, Security
  Scan `31097605349`, and Pages `31097604652` pass. The merge-commit recovery
  matrix completes in 13m54s.
- **Public readback:** Pages publishes 1,446 cases, 318.91 kB gzip, rollout
  16/59 (31%), phase work 16/51, Phase 6 6/7 (86%), and production false.
- **Tracker:** GH-29 is synchronized at exactly 16/59. No production, real
  account, real key, fund, deployment, or infrastructure action occurred.
- **Next:** GH-116 is the next bounded cleanup. GH-115 is Done.

---
## 2026-08-12 00:05 EEST GH-178 IBC channel recovery → In Progress

- **Branch:** `test/GH-178-ibc-channel-recovery`
- **Issue:** [GH-178](https://github.com/NeaBouli/TrueRepublic/issues/178)
- **Baseline:** exact clean `origin/main` `5a22f1c`; no unrelated local work and
  no open pull request was present at takeover.
- **Scope:** prove a real TrueRepublic two-chain ICS-20 packet can be refunded
  exactly once through counterparty channel closure and `TimeoutOnClose`, then
  prove safe continuation through a distinct replacement channel and ACK.
- **Delegation:** Kimi K3 receives a bounded, secret-free read-only design and
  risk review. Sol retains architecture, every write, security judgment,
  complete verification, GitHub actions, merge, and closure.
- **Risk:** High because packet commitments, channel state, bank escrow,
  refunds, replay behavior, and persistent recovery are protocol-critical.
- **Safety boundary:** ephemeral local applications only; no external relayer,
  public network, real key/account/funds, upgrade, deployment, release, or
  production action.
- **Ready for:** bounded design review, implementation, and focused validation.

---

## 2026-08-12 00:41 EEST GH-178 IBC channel recovery → Review

- **Changed:** real two-chain CloseConfirm/TimeoutOnClose/refund/replay/restart/
  replacement evidence; protected CI target; public IBC, test-count, threat,
  limitation, rollout, wiki, status, and agent documentation.
- **Tests:** focused IBC PASS; complete Go Build/Vet/Race/Coverage PASS;
  quality/fuzz/security/static/secret/vulnerability gates PASS; maintained
  client lint/141 tests/build/audit PASS; Rust format/Clippy/26 tests/audit PASS;
  docs/status/diff consistency PASS.
- **Independent review:** Kimi confirmed the committed-fixture boundary and real
  proof paths. No P0/P1/P2 surfaced; its P3 exact negative-error assertion was
  remediated and the focused gate rerun successfully.
- **Risk:** High protocol/accounting test surface; no production state changed.
- **Residual:** external relayer/counterparty, post-upgrade recovery, protocol
  stubs, deployment, real keys/funds, and release approval remain open.
- **Ready for:** commit, protected PR/exact-head checks, merge, final-main and
  live Pages verification. Rollout remains 25/59; production remains false.

---

## 2026-08-12 01:20 EEST GH-178 IBC channel recovery → Done

- **Issue/PR:** GH-178 closed through protected PR #179; merge commit
  `ef8d600fa8f0a769d8ff3dd117c76b8dfd13a9bd` is on `main`.
- **Reviewed head:** `02b3aefff623c0b79ec336c4fc1b777df64a3987`
  passed all 20 reported contexts. Three CodeRabbit documentation findings were
  fixed in `02b3aef`, answered, and resolved; zero review threads remain open.
- **Final-main:** Go CI `31541112827`, Security `31541112905`, reproducible
  Linux `31541112832`, Docs `31541112826`, and Pages `31541111526` all PASS on
  the merge commit.
- **Live Pages:** readback publishes 1,609 standard-suite cases (1,442 Go + 26
  Rust + 141 maintained client), rollout 25/59, Phase 6 6/7,
  `production_ready=false`, and IBC status
  `two_chain_channel_recovery_verified_locally_on_gh178`.
- **Tracker:** GH-29 now records pending-ACK recovery through GH-175 and channel
  close/replacement plus timeout-on-close through GH-178. Its combined item
  remains open for post-upgrade recovery.
- **Kimi contribution:** bounded secret-free read-only design and final diff
  reviews confirmed the ibc-go fixture/proof boundary; Sol fixed Kimi's one P3
  assertion-precision finding and retained every write, test, GitHub, merge,
  and security decision.
- **Residual:** external relayer/public counterparty qualification,
  post-upgrade IBC recovery, upgrade/staking/distribution stubs, production
  deployment, real keys/funds, and release approval remain open.
- **Safety:** no production, deployment, public-network, real-key, account, or
  fund action occurred. GH-178 is Done; production remains false.

---

## 2026-08-12 10:27 EEST GH-181 compatible IBC post-upgrade recovery → In Progress

- **Branch:** `test/GH-181-ibc-compatible-upgrade`
- **Issue:** [GH-181](https://github.com/NeaBouli/TrueRepublic/issues/181)
- **Baseline:** exact clean `origin/main` `968f334`; no open pull request and no
  unrelated work in this worktree at takeover.
- **Scope:** prove established two-chain IBC state and pending packets survive
  the supported state-compatible application/binary restart boundary; complete
  post-reopen ACK and timeout paths, and prove fail-before-open rollback does
  not mutate state.
- **Delegation:** Kimi K3 receives a bounded, secret-free read-only architecture
  and risk review. Sol owns scope, architecture, every write, security,
  complete verification, GitHub actions, merge, and closure.
- **Risk:** High evidence surface because IBC proofs, packet commitments,
  escrow, replay behavior, persistent databases, and upgrade/rollback wording
  are protocol-critical.
- **Safety boundary:** no `x/upgrade`, consensus-breaking migration,
  governance-controlled upgrade, IBC client upgrade, external relayer/public
  network, production, deployment, real key/account/funds, or release action.
- **Ready for:** architecture review, bounded implementation, focused and full
  local verification, protected PR, final-main, tracker/Page/Bridge closure.

## 2026-08-12 11:15 EEST GH-181 compatible IBC binary restart → Locally Verified

- **Implementation:** a separately linked candidate test binary reopens the
  same two LevelDB states without `InitChain`, export/import, or height reset.
  Clients, OPEN connections, unordered channel, sequence, escrow, commitment,
  receipt, and ACK commitment persist.
- **Recovery:** pending and fresh ACK plus timeout/refund complete exactly once;
  duplicate ACK/timeout relays are balance-checked no-ops. A fail-before-open
  exit-42 candidate leaves both recursive database hashes byte-identical.
- **Review:** Kimi no P0/P1/P2; both P3 notes remediated and rerun.
- **PASS:** full IBC, Go/Race/Coverage, quality/fuzz, security scanners, client,
  Rust/audit, docs consistency, JSON and diff integrity.
- **Candidate:** 1,610 = 1,443 Go + 26 Rust + 141 client; rollout 26/59, Phase 6
  6/7, production false.
- **Boundary:** same-source compatible binary restart only; `x/upgrade`, governed
  migrations, IBC client upgrades, external relayers, daemon/public deployment,
  arbitrary version compatibility, real keys and funds remain open.
- **Next:** protected PR/CI, merge, final-main/GH-29/Pages readback, then Done.

## 2026-08-12 11:22 EEST GH-181 → Published for protected review

- **Commits:** `0f49d88` test/CI implementation; `4255f54` audit, public status,
  tracker candidate, and Bridge evidence.
- **Pull request:** [PR #182](https://github.com/NeaBouli/TrueRepublic/pull/182)
  targets `main`; exact final head and protected contexts remain pending.
- **Local state:** clean after publication; all recorded local gates pass.
- **Next:** mark ready, review exact-head CI/review threads, remediate in scope,
  merge only when green, then final-main/GH-29/Pages/Bridge closure.

---

## 2026-08-12 11:45 EEST GH-181 compatible IBC binary restart → Done

- **Issue/PR:** GH-181 closed through protected PR #182; merge commit
  `2f22b60d0dc75e25657f70165d7209668d3cb957` is on `main`.
- **Reviewed head:** `68341f8c78d887681167e1cedc93c006b7ca6831`
  passed all 20 reported contexts. Four CodeRabbit findings were fixed,
  answered, and resolved; zero review threads remain open.
- **Final-main:** Docs `31579387559`, Security `31579387595`, and Pages
  `31579386407` pass on the merge commit. Pages initially hit a hosted-runner
  self-signed-certificate error after a successful build/upload; the failed
  deploy retry passed without a repository change.
- **Live Pages:** readback publishes 1,610 standard-suite cases (1,443 Go + 26
  Rust + 141 maintained client), rollout 26/59, Phase 6 6/7,
  `production_ready=false`, and IBC status
  `compatible_binary_restart_recovery_verified_locally_on_gh181`.
- **Tracker:** GH-29 marks partial-failure and compatible binary-restart
  recovery complete through GH-175/GH-178/GH-181. Governance-controlled
  consensus migrations and IBC client upgrades remain separate open work.
- **Kimi contribution:** bounded secret-free read-only design and final reviews;
  no Kimi writes. Sol remediated both P3 hardening notes and retained every
  write, test, security, GitHub, tracker, merge, and closure decision.
- **Residual:** external relayer/public counterparty qualification, governed
  `x/upgrade` and migration rollback, IBC client upgrades,
  staking/distribution/upgrade stubs, production deployment, real keys/funds,
  and release approval remain open.
- **Safety:** no production, public-network, real-key, account, fund, or release
  action occurred. GH-181 is Done; production remains false.

---

## 2026-08-12 13:30 EEST GH-184 governed upgrade → Locally Verified

- **Implementation:** real `x/upgrade` keeper/module/PreBlock replaces the IBC
  upgrade stub. Its authority is the non-signing `truedemocracy` module account;
  only the reserved genesis governance Domain's immutable electorate may
  schedule or cancel an exact plan at a two-thirds threshold.
- **Release boundary:** old artifacts halt; only an artifact with exact
  `main.upgradePlan=v0.4.1` registers the migration handler while
  `main.version` remains the source commit. Pre-GH-184 store introduction, arbitrary
  plans, and IBC client upgrades remain unsupported.
- **PASS:** focused governance suite and the four-validator three-artifact
  process harness. The latter proves old-binary halt, intentional partial
  cached-write failure rollback, fixed-candidate recovery, common app hashes,
  unchanged validator power, and exact-once restart in 230.63 seconds.
- **Kimi contribution:** bounded secret-free architecture review and
  `x/truedemocracy` adapter implementation. Sol reviewed every write, added
  adversarial tests/harness, and owns security, integration, full gates, and
  all external actions.
- **Candidate status:** 1,621 standard cases (1,454 Go + 26 Rust + 141 client),
  rollout 28/59, phase work 28/51, Phase 6 6/7, production false.
- **Pending:** full local/security gates, final Kimi review, protected PR/CI,
  merge, final-main, GH-29, live Pages, and both-Bridge closure.
- **Safety:** no production, deployment, public network, real keys/accounts,
  funds, or release action occurred.

---

## 2026-08-12 13:58 EEST GH-184 governed upgrade → Hardened verification

- **Review remediations:** reserved governance membership is genesis-immutable;
  pending scheduled plans and authenticated votes round-trip through export/import;
  stale unscheduled proposals remain cancellable; malformed upgrade genesis is
  rejected; import inside the original lead window is safe while plans at/past
  execution height fail closed.
- **Release identity:** reproducible artifacts keep `main.version` bound to the
  source commit and independently inject exact `main.upgradePlan=v0.4.1`; the
  build contract and verifier agree on both values.
- **CI ownership:** `make governed-upgrade` and the protected
  `governed-upgrade` workflow job now own the three-artifact process gate.
- **PASS:** fresh 1,454-case Go arithmetic (551 governance), deterministic-build
  contract, focused governance suite, documentation consistency, JSON/diff
  integrity, and fresh four-validator harness in 231.39 seconds.
- **Candidate:** 1,621 standard cases (1,454 Go + 26 Rust + 141 client), rollout
  28/59, phase work 28/51, Phase 6 6/7, production false.
- **Pending:** full race/coverage/security/client/Rust gates, final independent
  review, protected PR/CI, merge, final-main/GH-29/Pages readback and Done.

---

## 2026-08-12 14:22 EEST GH-184 governed upgrade → Local gates complete

- **PASS:** build, package selector, Vet, 1,457-case Race/Coverage, maintained
  coverage floors (root 72.2%, governance 64.0%, DEX 51.1%), property/fuzz,
  security contract, deterministic build, Staticcheck, exact-policy
  Govulncheck, Gitleaks and their negative fixtures.
- **PASS:** maintained client fresh install/lint/141 tests/build/bundle/audit;
  Rust fmt/clippy/build/26 tests/audit; docs consistency, JSON and diff checks;
  fresh four-validator three-artifact process gate in 231.39 seconds.
- **Review:** Kimi final read-only review reports no P0/P1/P2. Sol resolved the
  actionable P3s with fresh exact counting and tested/documented authenticated
  post-execution cleanup. Historical audit statements remain historical.
- **Candidate:** 1,624 standard cases (1,457 Go + 26 Rust + 141 client), rollout
  28/59, phase work 28/51, Phase 6 6/7, production false.
- **Pending:** commit, protected PR/exact-head CI, review threads, merge,
  final-main/GH-29/live Pages and both-Bridge Done closure.
- **Safety:** no deployment, public network, real keys/accounts, funds or
  release action occurred.

---

## 2026-08-12 14:27 EEST GH-184 → PR #185 published

- **Implementation commit:** `5a236527333ec21f8cec8213d3cb62dce9bfcf73`
  publishes the complete reviewed governed-upgrade implementation and local
  evidence on `feature/GH-184-governed-upgrade`.
- **Pull request:** [PR #185](https://github.com/NeaBouli/TrueRepublic/pull/185)
  targets protected `main` and closes GH-184 on merge.
- **Local evidence:** every required Go/Race/Coverage, process, security,
  deterministic build, client, Rust/audit, documentation and review gate passes.
- **Next:** this append-only handoff becomes the final PR head; wait for every
  exact-head context and review thread, remediate without bypassing gates, then
  merge and perform final-main/GH-29/Pages/both-Bridge closure.

---

## 2026-08-12 14:52 EEST GH-184 governed upgrade → Done

- **GitHub:** PR #185 exact reviewed head `9c4f58c` passed all 21 reported
  contexts with zero review threads and merged as `bf32ad5`; GH-184 closed and
  the remote feature branch was deleted.
- **Final main:** Docs `31592981577`, Security `31592981538`, reproducible
  Linux `31592981655`, Pages `31592980432`, and Go `31592981562` pass.
- **Public readback:** live Pages reports 1,624 standard cases (1,457 Go + 26
  Rust + 141 client), rollout 28/59, phase work 28/51, Phase 6 6/7,
  production false, and `governed_application_upgrade_recovery_verified_locally_on_gh184`.
- **Tracker:** GH-29 credits governed application-upgrade recovery only within
  the documented fresh-genesis v0.4.1 boundary. Pre-GH-184 store introduction,
  IBC client upgrades, external relayers, remaining staking/distribution stubs,
  production deployment, and release approval remain open.
- **Agents:** Kimi performed bounded secret-free implementation/review with no
  final P0/P1/P2; Sol reviewed every write and retained full architecture,
  security, integration, test, GitHub, merge, tracker, and closure ownership.
- **Safety:** no production, public-network, real-key/account, fund, release,
  or deployment action occurred. GH-184 is Done.

---

## 2026-08-12 22:53 EEST GH-187 protocol compatibility boundary → In Progress

- **Branch:** `refactor/GH-187-protocol-boundary`
- **Issue:** [GH-187](https://github.com/NeaBouli/TrueRepublic/issues/187)
- **Scope:** replace ambiguous staking/distribution stubs with explicit
  fail-closed compatibility adapters; prove unsupported `x/staking` and
  `x/distribution` module/query/transaction surfaces stay absent; freeze the
  exact supported Cosmos SDK, CosmWasm, IBC and governed-upgrade boundary.
- **Invariants:** preserve PoD validator/reward accounting, ICS-20 transfer,
  governed `x/upgrade`, genesis/export/restart behavior, and the 21M PNYX cap;
  do not introduce a consensus-state migration or delegated staking semantics.
- **Delegation:** Kimi K3 receives a bounded, secret-free architecture and
  implementation analysis. Sol owns scope, architecture, every accepted write,
  security, complete verification, GitHub, merge, tracker and closure.
- **Risk:** High because upstream keeper interfaces and protocol exposure are
  consensus-adjacent. The implementation must remain behavior-preserving and
  fail closed on unsupported surfaces.
- **Safety:** repository-only. No production/public network, relayer,
  counterparty, real key/account/fund, IBC client upgrade, store migration,
  release or deployment action.
- **Ready for:** code/interface census, Kimi review, bounded implementation and
  focused/full verification.

---

## 2026-08-12 23:27 EEST GH-187 local implementation/process gates → Verified

- **Implementation:** required ibc-go/wasmd staking and distribution adapters
  now reject through a stable unsupported-surface boundary; explicit Wasm
  query/message plugins prevent empty or zero-valued success responses.
- **Absence proof:** four new standard cases cover every adapter and prove
  `x/staking`/`x/distribution` are absent from module basics/runtime, stores,
  default genesis, gRPC, message routers, and root CLI query/transaction trees.
- **Real gates:** IBC two-chain transfer/ACK/timeout/replay/channel/restart PASS
  (51.06s root); four-validator governed upgrade halt/failure/rollback/recovery
  PASS (236.47s). Fresh Go arithmetic PASS: 1,461.
- **Candidate status:** 1,628 standard cases (1,461 Go + 26 Rust + 141 client),
  rollout 30/59, phase work 30/51, Phase 6 6/7, production false.
- **Agents:** Kimi delivered bounded secret-free read-only architecture review
  and no accepted write diff. Sol implemented/reviewed all writes and owns full
  verification, GitHub, merge, tracker, and closure.
- **Pending:** full repository/security/client/Rust/deterministic gates, final
  independent review, protected PR/exact-head CI, merge, GH-29/final-main/live
  Pages, and both-Bridge Done closure.
- **Safety:** no production/public network, relayer/counterparty, real
  key/account/fund, store or IBC client migration, release, or deployment
  action occurred.

---

## 2026-08-12 23:48 EEST GH-187 full local/security gates → PASS

- Full Go build/vet/Race/Coverage, critical floors, property/fuzz,
  contention/replay/restart, real IBC and governed-upgrade gates pass.
- Pinned Staticcheck, exact-policy Govuln plus negative fixtures, Gitleaks,
  docs/JSON/security contract, maintained-client lint/141 tests/build/audit,
  and Rust format/Clippy/build/26 tests/audit pass.
- Kimi's independent review found one public test-attribution error; Sol fixed
  README and two wiki references so GH-187 is correctly inside the 1,461 Go
  standard cases rather than described as a separate process gate.
- Clean-commit deterministic Linux build, protected PR/exact-head CI and
  review, merge, final-main/GH-29/live Pages and both-Bridge Done closure remain.

---

## 2026-08-12 23:54 EEST GH-187 review commit → Ready for protected PR

- Reviewed implementation commit: `e147dce796fa3c237e154bfd9667f6eda21b332a`.
- Kimi follow-up: attribution finding RESOLVED; no P0/P1/P2 remains in the
  remediated public-count diffs.
- The clean-commit deterministic command correctly refused the macOS host with
  `linux-amd64 requires a native Linux x86_64 runner`. Protected native-Linux CI
  must produce that exact-head evidence; it is not claimed locally.
- All other local repository, process, client, Rust, security, consistency and
  review gates pass. Next: publish branch/PR, monitor exact-head checks and
  threads, remediate if needed, then merge and complete public closure.

---

## 2026-08-12 23:58 EEST GH-187 → PR #188 published

- **PR:** [#188](https://github.com/NeaBouli/TrueRepublic/pull/188), protected
  `main`, closes GH-187.
- **Branch:** `refactor/GH-187-protocol-boundary`.
- **Evidence:** PR records the full local process/security/client/Rust review
  matrix and the explicit macOS/native-Linux reproducibility boundary.
- **Next:** push this final append-only handoff, wait for every exact-head
  context and review thread, remediate without bypassing gates, then merge and
  verify final main, GH-29, live Pages and both Bridges.

---

## 2026-08-13 00:30 EEST GH-187 → Done

- **Publication:** exact PR head `ef6afbbe063ab5b8333d5d866640e031561c27cf`
  passed all 22 reported contexts with zero review threads; protected PR #188
  merged as `6c8d2a8eb9f9271a9182fafe383bd058d68c7263`, its remote branch was
  deleted, and GH-187 closed.
- **Final main:** Go `31641633848`, Docs `31641633858`, Security
  `31641633794`, reproducible Linux `31641633845`, and Pages `31641632893`
  pass.
- **Public synchronization:** GH-29 now credits the completed governed-upgrade,
  fail-closed compatibility-adapter, and exact feature-boundary work. Live
  Pages reports 1,628 cases, rollout 30/59, phase work 30/51, Phase 6 6/7, and
  production readiness false.
- **Agents:** Kimi supplied bounded, secret-free read-only architecture and
  final reviews; no Kimi write diff was accepted. Sol implemented, reviewed,
  tested, published, and merged every change.
- **Safety:** no production/public network, relayer/counterparty, real
  key/account/fund, migration, release, or deployment action occurred.
- **Closure:** every repository, security, process, native-build, protected
  publication, tracker, live-site, and Bridge gate is satisfied. GH-187 is
  Done; remaining rollout work stays explicitly separate.

---

## 2026-08-13 02:23 EEST GH-190 IBC transfer UX and recovery status → In Progress

- **Branch:** `feature/GH-190-ibc-transfer-ux`
- **Issue:** [GH-190](https://github.com/NeaBouli/TrueRepublic/issues/190)
- **Scope:** complete the maintained client's canonical ICS-20 transfer form,
  fail-closed channel/amount/timeout validation, source-chain-evidenced packet
  lifecycle, persistence, timeout/refund and uncertain-broadcast recovery UX.
- **Starting point:** exact clean `origin/main`
  `c845e99fa88f36d99bcee44e427ba7198d613133`; no open PRs. GH-29 remains the
  parent rollout tracker at 30/59 overall and 30/51 phase work.
- **Initial findings:** the native Send form can accept non-numeric input as a
  zero amount, and the current IBC channel query conflates transport/schema
  failure with an empty channel set. Both must be fixed before reuse.
- **Delegation:** Kimi K3 completed a bounded secret-free read-only architecture
  review and will implement a substantial, named client slice. Claude Code is
  limited to one small read-only/test-impact helper task. Neither may delegate
  further or perform external writes. Sol owns architecture, security, every
  accepted diff, integration, complete tests, GitHub, merge and closure.
- **Risk:** High because misleading cross-chain status or replay UX can cause
  duplicate escrow. Broadcast success must never be presented as delivery;
  uncertain outcomes must be checked before manual resubmission.
- **Safety:** repository-only. External relayer/counterparty qualification,
  IBC client upgrades, ZKP, wallet custody, contracts, consensus migration,
  production, real keys/accounts/funds, release and deployment are excluded.
- **Ready for:** bounded Kimi implementation, small Claude test inventory, Sol
  integration/security review, focused tests and the complete relevant gates.

---

## 2026-08-13 03:16 EEST GH-190 implementation → Locally verified

- Canonical native MsgTransfer, strict channel/amount/fresh-balance/timeout
  checks, wallet-scoped non-secret records, lazy transfer/status UI, and manual
  source-chain send_packet/ACK/timeout recovery are implemented. Broadcast is
  never delivery and no path automatically resubmits.
- PASS: client lint, 10 Node + 249 Vitest cases, production build/budgets and
  audit; disposable local-chain delivery/rejection; full Go build/vet/Race/
  Coverage; real two-chain IBC recovery; security/static/vulnerability/secret
  gates; Rust format/Clippy/26 tests/audit; docs/JSON/retirement consistency.
- Public candidate status is 1,746 standard cases, rollout 31/59, phase work
  31/51, Phase 6 6/7, production false.
- Kimi implemented the large core slice; Sol reviewed and extended every
  accepted diff. Claude Code performed only the small route/bundle inventory.
  A final Kimi review did not complete; Sol's final review found and added
  canonical packet_ack_hex decoding. No known P0/P1/P2 remains.
- Safety remains repository/local-chain only. Next: commit, publish protected
  PR, wait for exact-head checks/review, merge, synchronize GH-29/live Pages and
  both Bridges, then close GH-190.

---

## 2026-08-13 03:18 EEST GH-190 → PR #191 published

- Implementation commit `7eca8eb` was pushed and protected
  [PR #191](https://github.com/NeaBouli/TrueRepublic/pull/191) opened against
  `main`, closing GH-190 and advancing GH-29.
- PR evidence records the complete local client, disposable-chain, Go/IBC,
  security, Rust, consistency and safety boundary. The candidate is ready for
  exact-head protected checks and review; it is not yet merged or Done.

---

## 2026-08-13 03:27 EEST GH-190 IBC transfer UX and recovery → Done

- **Publication:** exact PR head `847540f` passed all 14 reported contexts with
  zero review threads; PR #191 merged as `ed669a2`, GH-190 closed, and the
  remote feature branch was deleted.
- **Final main:** Client Web CI `31654208105`, Security `31654208135`, Docs
  `31654208130`, and Pages `31654207520` pass on the merge commit.
- **Public synchronization:** GH-29 credits GH-190. Live Pages reports 1,746
  standard cases, rollout 31/59, phase work 31/51, Phase 6 6/7, 20 lazy routes,
  the GH-190 IBC recovery boundary, and production readiness false.
- **Agents:** Kimi supplied the large bounded core implementation; Sol reviewed
  and extended every accepted write and owns all integration, testing,
  publication, merge, tracker and closure work. Claude Code supplied only a
  small read-only inventory.
- **Safety:** no external relayer/public counterparty, production/public
  network, real key/account/fund, release, migration, or deployment action
  occurred. The scoped task is fully complete and verified.

---

## 2026-08-13 03:41 EEST GH-193 wallet and signing safety → In Progress

- **Issue/branch:** [GH-193](https://github.com/NeaBouli/TrueRepublic/issues/193)
  on `feature/GH-193-wallet-safety`, based on exact clean `origin/main`
  `5cc5edc` after the fully verified GH-190 closeout.
- **Scope:** qualify maintained-client wallet creation/import, encrypted
  persistence, unlock/lock/switch/delete/reload invalidation, signer and chain
  scoping, and sensitive-data exclusion with negative and lifecycle evidence.
- **Delegation:** Kimi K3 receives a large bounded secret-free architecture and
  implementation block. Claude Code receives only a small read-only inventory
  task. Neither may delegate or write externally. Sol owns security decisions,
  every accepted diff, integration, full tests, GitHub, merge, and closure.
- **Safety:** repository and disposable local-chain work only; no real key,
  account, fund, public RPC, production wallet, release, deployment, external
  message, or production-readiness claim.
- **Status:** task and durable handoff are created before code changes. Next is
  parallel inventory/review followed by bounded implementation and full gates.

---

## 2026-08-13 03:58 EEST GH-193 client slice → Locally verified

- Kimi delivered the large bounded wallet/service/test implementation; Sol
  reviewed every write and added versioned PBKDF2-600k legacy re-encryption,
  fail-closed malformed storage handling, stale-unlock protection, and
  session-bound in-flight signer invalidation. Claude Code supplied only the
  small read-only signing inventory.
- PASS: 10 Node + 295 Vitest cases, lint, TypeScript production build/budgets,
  and three real disposable-chain client deliveries. Build is 71.13 kB gzip
  entry, 4.91 kB maximum lazy route, 355.08 kB total JavaScript.
- Candidate status: 1,792 standard cases, rollout 32/59, phase work 32/51,
  Phase 6 6/7, production false. Browser compromise/XSS, hardware custody, real
  keys/accounts/funds, public RPC, release and deployment remain excluded.
- Next: full Go/Race/Coverage, IBC, security/tooling, Rust, consistency and
  independent-review gates; then protected PR/merge/public closure.

---

## 2026-08-13 04:17 EEST GH-193 full local gates → PASS

- Full Go verify and real two-chain IBC transfer/ACK/timeout/replay/recovery
  pass. Pinned Govulncheck, Staticcheck and Gitleaks pass; Rust fmt/Clippy/26
  tests/audit pass with no vulnerability and five allowed existing warnings;
  npm audit, docs, JSON and retirement contracts pass.
- Kimi's independent review found no P0/P1 and one P2 concurrent-save race. Sol
  fixed it, a stale balance-refresh-after-lock race, malformed-JSON error
  bounding, and post-lock signer mnemonic access, then reran the relevant full
  client gates: 10 Node + 298 Vitest, lint and build/budgets all pass.
- Candidate status is now 1,795 standard cases, rollout 32/59, phase work 32/51,
  Phase 6 6/7, production false. No known P0/P1/P2 remains. Next is protected
  PR, exact-head checks/review, merge, GH-29/live Pages and both-Bridge closure.

---

## 2026-08-13 04:20 EEST GH-193 → PR #194 published

- Reviewed commit `700eb18` is pushed and protected
  [PR #194](https://github.com/NeaBouli/TrueRepublic/pull/194) targets `main`,
  closes GH-193 and carries the complete local evidence and safety boundary.
- Exact-head protected checks and review are now pending; merge, GH-29, final
  main, live Pages and both-Bridge closure remain required before Done.

---

## 2026-08-13 04:30 EEST GH-193 wallet and signing safety → Done

- Exact PR head `16d752f` passed all 14 reported contexts with zero review
  threads; protected PR #194 merged as exact main `10a9a83`, GH-193 closed,
  and the remote feature branch was deleted.
- Final-main Client Web CI `31657679062`, Security `31657679041`, Docs
  `31657679040`, and Pages `31657678393` pass. GH-29 and live Pages publish
  1,795 cases, rollout 32/59, phase work 32/51 and Phase 6 6/7.
- Kimi delivered the large implementation and independent review; Sol fixed
  the review's P2 plus the stale-balance race, reviewed every accepted write,
  and owns complete verification/publication. No known P0/P1/P2 remains.
- No production/public network, real key/account/fund, release or deployment
  action occurred. GH-193 is complete; remaining rollout work is separate.

---

## 2026-08-13 04:40 EEST GH-193 closeout → Published and verified

- Closeout PR #195 exact head `8d8731d` passed all 11 reported contexts with
  zero review threads and merged as `65a3c9c`; its remote branch is deleted.
- Final-closeout main Security `31658266050`, Docs `31658266070`, and Pages
  `31658262780` pass. Cache-busted live readback confirms 1,795 cases, 308
  maintained-client cases, rollout 32/59, phase work 32/51, Phase 6 6/7 and the
  GH-193 wallet-safety status. GH-29 is checked and no PR remains open except
  this final append-only handoff publication.
- GH-193 is fully closed with no known P0/P1/P2 and no production/public-network,
  real-key/account/fund, release or deployment action.

---

## 2026-08-13 12:55 EEST GH-198 ZKP compatibility foundation → In Progress

- User granted an exact bounded approval for GH-198. Scope is repository-only,
  explicitly test-only Groth16 artifacts, synthetic golden vectors, Go/client
  encoding and proof compatibility, negative tests, and deterministic CI gates.
- Safety boundary: no consensus wire-format or active genesis-VK change, no
  production ceremony or keys, no public network/RPC, no real account/funds,
  no reward-recipient design, no deployment, and no anonymous UI/transaction
  enablement. `isSubmittable` must remain false.
- Baseline is clean `origin/main` `954b11b`. A separate owner-authored docs-only
  PR #197 modifies only `BRIDGE.md`; it was inspected read-only and is preserved
  without merge, edit, or attribution to GH-198. GH-198 uses isolated branch
  `feat/GH-198-zkp-compatibility`.
- Kimi's read-only inventory confirmed the chain verifier and randomized
  test-time setup exist, while no committed artifact set, deterministic vector
  replay, maintained-client prover, or cross-language encoding gate exists.
  Claude independently identified a test-injectable client prover seam that
  preserves the fail-closed production boundary.
- Next: Kimi implements the isolated Go artifact/vector slice; Sol reviews every
  byte, integrates client/CI/security work, runs the complete relevant gates,
  and reconciles PR #197 before GH-198 publication so Bridge history cannot slip.

---

## 2026-08-13 15:18 EEST GH-198 implementation and local gates → PASS

- Added immutable test-only CS (1,656,948 bytes), PK (2,382,883), VK (460) and
  synthetic golden-vector/manifest artifacts. Exact size/SHA-256, circuit ID,
  public-input order, CS/PK decoding, canonical VK shape/fingerprint, proof and
  keeper replay, corruption, wrong root/chain/rating, manifest drift and trailing
  data are fail-closed. The env-gated generator intentionally performs a new
  randomized single-party setup; regeneration is not a production ceremony.
- Added a maintained-client BN254 MiMC and vote-context compatibility module plus
  shared-vector tests. The only service integration is a test-injectable prover;
  `isSubmittable` remains hard false and no transaction registration/UI broadcast
  path was added.
- PASS: focused Go/client tests; full 1,473-case Go race/coverage suite; 10 Node +
  301 Vitest = 311 client cases; ESLint; TypeScript/Vite build and bundle budgets
  (355.09 kB total); high audit; vet; critical coverage; docs/JSON/retirement
  checks; pinned Gitleaks v8.30.1 maintained-tree scan with no leak.
- Kimi independently verified MiMC constants/rounds/compression, Keccak variant,
  field/context/rating encoding, fail-closed client and artifact boundaries and
  found no P0/P1/P2. Sol fixed its P3 clean-failure notes for unexpected fixture
  directories and malformed 32-byte hex. The accepted residual P3 is ~3.9 MB Git
  growth for forge-capable test material; production ZKP risk remains critical.
- Public arithmetic is 1,810 standard cases. Rollout remains 32/59 because no
  production ZKP checkbox is claimed. Next: reconcile separate PR #197, commit,
  protected PR checks/review, merge and final main/Pages/Bridge verification.

---

## 2026-08-13 15:22 EEST PR #197 reconciled → GH-198 publication-ready

- Owner-authored docs-only PR #197 passed its reported security contexts and was
  admin-squash-merged as `fc5b418`; its remote branch was deleted. Its public
  audit handoff remains intact at the top of the append-only root Bridge.
- GH-198 rebased cleanly onto exact `origin/main` `fc5b418` with no conflict or
  lost line. Candidate commit is now `028a555`; protected publication is next.

---

## 2026-08-13 15:41 EEST GH-198 review follow-up → PASS

- Independently verified the GitHub review findings before changing code. The
  fixture loader now exercises its exact checksum rejection helper, keeper replay
  checks `AddMember`, the client prover seam proves exact input forwarding, and
  empty MiMC elements match Go field-zero behavior. Strict JSON now uses one
  explicit EOF check; no production or consensus path changed.
- Synchronized the current Bridge state for completed PR #197 reconciliation.
  Full post-review Go `-race ./...` passes. Full client evidence remains 10 Node
  + 301 Vitest cases, 3 skipped; lint, TypeScript/Vite build, bundle budgets
  (entry 72,868 bytes, 20 lazy routes, max 5,030, total 363,613), and high audit
  all pass. Public arithmetic therefore remains 1,810 and rollout remains 32/59.
- No known P0/P1/P2 remains. The next protected head must rerun every GitHub
  context and close all review threads before merge.

---

## 2026-08-13 16:03 EEST GH-200 final-main quality flake → Diagnosed

- PR #199 merged as `f7a9a52` after exact head `ab9a3cd` passed all 24
  contexts with zero open review threads; GH-198 closed and its branch was
  deleted. Final-main Docs and Pages pass, and live Pages publishes 1,810 cases,
  rollout 32/59, and production false.
- Final-main Go run `31702374261` diverged only in `quality-depth`:
  `FuzzComputeSwapOutput` completed 15,459 executions but exceeded its exact
  10-second wall-clock fuzz deadline by ~80 ms under hosted-runner contention.
  No invariant, DEX, ZKP, race, or product assertion failed. The exact command
  then passed five consecutive local reproductions.
- Opened GH-200 for a bounded harness-only remediation: replace wall-clock fuzz
  duration with an equivalent finite iteration budget and retain every
  property/fuzz/race gate. No product, consensus, client, artifact, production,
  key/account/fund, release or deployment path may change.

---

## 2026-08-13 16:09 EEST GH-200 iteration-bounded harness → Local PASS

- Replaced both 10-second wall-clock fuzz deadlines with exact 10,000-execution
  budgets. The optional override is numeric, fail-closed, and capped at 60,000
  so each invocation retains margin under its fixed 180-second timeout even on
  substantially slower runners. Race/property packages, fuzz targets,
  assertions, parallelism, workflow wiring and job timeout are unchanged.
- PASS: repository contract, invalid `0` and `60001` configuration rejection,
  and complete generative-quality script. Both fuzz targets reached exactly
  10,000 executions and exited normally; no wall-clock context governs them.
- Claude Code's focused read-only review first found the excessive 1,000,000
  override cap as P2. Sol reduced it to 60,000; Claude's follow-up is APPROVED
  with no P0/P1/P2/P3. No product or production path changed.

---

## 2026-08-13 16:13 EEST GH-200 complete local gates → PASS

- Post-review full `go test -race ./...`, `go vet ./...`, documentation
  consistency and diff hygiene pass. The complete iteration-bounded generative
  gate already passed both 10,000-execution fuzz targets plus its unchanged race
  property stage.
- Candidate changes remain limited to the quality harness, its existing
  repository contract test, and append-only handoff records. No public test or
  rollout count changes; protected GH-200 publication is next.

## 2026-08-13 16:42 EEST GH-198/GH-200 protected closeout → PASS

- PR #201 exact head `60cc641` passed every reported context, including
  iteration-bounded `quality-depth` and multi-validator recovery, with zero open
  review threads. Squash merge `b0e963f` closed GH-200; its local and remote
  feature branches were removed safely.
- Final-main Go CI `31705055413`, Security `31705055382`, Docs
  `31705055493`, Reproducible Linux `31705055420`, and Pages `31705054350`
  pass. The cache-busted public page still reports 1,810 tests and 32/59; it
  explicitly remains not production-ready.
- GH-198 is complete through PR #199 (`f7a9a52`) only as a test compatibility
  foundation. No production ceremony, prover, real keys/funds, public network,
  deployment or rollout authorization occurred. The divergent dirty legacy
  Desktop worktree was preserved and not modified.

---

## 2026-08-16 16:20 EEST - GH-222 protected review remediation

- Verified and addressed all six CodeRabbit threads plus Kimi's final P3
  hardening: exclusive marker errors retain causes, all independent managed
  directories are pruned before returning a joined error, dedicated
  transaction directories are allowlisted and tested, and systemd contract
  readability is explicit. Kimi's final review found no P0-P2.
- Fresh full Go Race/Coverage, vet, staticcheck, vulnerability policy and
  fixtures, gitleaks, lifecycle/repository, release-evidence, deterministic,
  docs and diff gates pass. The additional concurrency regression corrects
  candidate arithmetic to 1,921 = 1,576 Go + 26 Rust + 319 client.
- No release, tag, artifact publication, deployment, real state, network,
  key, fund or production action occurred. Public status remains 1,896 and
  rollout 33/59 until protected merge and closeout synchronization.

## 2026-08-16 16:43 EEST - GH-222 implementation merged / closeout in progress

- PR #223 exact head `c6cbcb7e634c1c3f6b4d8c38e768465bd075cc80`
  passed all 22 reported protected contexts with zero unresolved review threads.
  Admin squash merge `1bf9ce45cf55f0ee8c2f0c30e7b6a5a6818114a3`
  closed GH-222 and removed the remote feature branch; no technical gate was
  bypassed.
- GH-29 now credits the exact Phase 7 install/migrate/upgrade/rollback/uninstall
  documentation item with GH-222/PR #223 evidence and retains arbitrary
  migrations as unsupported. Public closeout candidate is 1,921 cases, rollout
  34/59, phase work 34/51, Phase 6 6/7, production false.
- Docs-only branch `docs/GH-222-closeout` synchronizes status.json, README,
  roadmap, Pages, wiki, limitations and handoff records. Protected docs review,
  merge and exact-main/Pages verification remain pending. No tag, release,
  publication, deployment or production action occurred.

## 2026-08-16 16:45 EEST - GH-222 docs closeout local PASS

- Synchronized all maintained public status surfaces to 1,921 cases (1,576 Go,
  26 Rust, 319 client), rollout 34/59, phase work 34/51, Phase 6 6/7 and
  production false. Added root 159 and install-lifecycle 24 to the canonical
  module table and extended the consistency gate to enforce both release and
  install-lifecycle rows.
- PASS: documentation consistency, JSON arithmetic/schema queries, stale-live-
  status scan and diff hygiene. GH-29 and the public roadmap now link the
  merged GH-222/PR #223 evidence with arbitrary migrations still unsupported.
- Next: protected docs PR, review/merge, exact-main workflows and cache-busted
  live Pages verification. No release, deployment or production action.
