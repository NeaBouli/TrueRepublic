# Action Log

## 2026-07-23 06:04 EEST - GH-55 validator identity recovery start

- Confirmed synchronized, clean `main` at `93f0263` with no open PRs before
  opening the next Phase 1 rollout task.
- Opened GH-55 for validator consensus-identity cold custody, single-signer
  failover, and compromise containment; created branch
  `feature/GH-55-validator-identity-recovery`.
- Audited the boundary between `node_key.json` (P2P identity) and the coupled
  `priv_validator_key.json` plus current `priv_validator_state.json`
  consensus-safety unit. Found unsafe key-only restore, plaintext full-home,
  configuration, Docker-volume, and pre-upgrade archive guidance.
- Confirmed the protocol does not yet support safe key rotation:
  `remove-validator` withdraws stake subject to a domain transfer limit, while
  bootstrap operator authority is derived from the consensus key. Opened
  follow-up GH-56 for authenticated atomic rotation, permanent revocation, and
  separated bootstrap operator authority.
- Implementation in progress: real four-validator cold-failover evidence, a
  fail-closed custody/incident runbook, corrected backup/security guidance,
  CI coverage, rollout status, and Bridge synchronization.
- Local evidence is green: the focused real-process failover passes in 62.17s;
  the full 656-case Go suite passes; build, vet, shell syntax, JSON, docs
  consistency, and diff checks pass; and all five multi-process recovery drills
  pass together in 636.342s. The new drill proves exact post-stop signer-state
  transfer, a distinct P2P identity, strict signing-position advance, unchanged
  consensus power, common app hash, source isolation, valid export/ledger, and
  re-import.
- The structured GH-55 pre-ship audit records 0 FAIL / 1 WARN / 5 PASS. Its one
  HIGH warning is the separate GH-56 protocol rotation gap, not a defect in the
  bounded cold-failover path.
- PR #57 is published. CodeRabbit raised eight review threads; four valid
  findings removed a duplicated word, retained compromised-key recovery as a
  separate blocker, made permission examples honor `CHAIN_HOME`, and required
  dedicated service-account ownership verification. The two future-date
  findings conflict with the authoritative July 23 runtime, and the two Go
  wrapper findings are invalid because the multi-package wrapper cannot combine
  `go build -o <file>` with multiple package arguments while the existing `.`
  commands already select the root package explicitly.
- PR #57 final head `aa66aa9` passed Go build/vet/race/coverage (6m55s), all
  five multi-validator recovery drills (8m44s), Docker restart (3m33s), docs,
  Go/Rust/Node security, DeepScan, and CodeRabbit. All eight review threads are
  resolved. The PR squash-merged to `main` as `e8670c6` and closed GH-55.

## 2026-07-23 04:54 EEST - GH-53 persisted binary upgrade and rollback start

- Confirmed `main` was synchronized with `origin/main` at `312752a`, with no
  open PRs and only parent issues #4, #7, and #29 open.
- Opened GH-53 under rollout tracker GH-29 and created branch
  `feature/GH-53-persisted-upgrade-rollback`.
- Audited the upgrade boundary: versioned binary builds are supported, but
  `x/upgrade`, migration handlers, and governance-controlled halt/resume are
  not wired. Scope is therefore compatible binary replacement and
  fail-before-open rollback, not consensus-breaking state migration.
- Found and removed unsafe operator guidance that archived/restored the entire
  validator home and could regress `priv_validator_state.json` after newer
  heights were signed.
- Added `TestMultiValidatorPersistedBinaryUpgradeRollback`: four real
  validators commit a funded domain, roll one-by-one to a separately versioned
  compatible artifact, exercise an intentional candidate failure before state
  is opened, and roll every node back to the baseline artifact on the same
  homes. The drill checks historical/current app hashes, validator power,
  unchanged node/validator identity keys, non-regressing signing positions,
  exported ledger validity, persisted application state, and re-import.
- Local focused compile/SKIP verification passes. The gated real-process drill
  passes in 203.97s. The full repository suite passes (`truerepublic` in
  55.207s plus all module packages). The CI-equivalent combined consensus,
  trusted state-sync, backup/restore, and upgrade/rollback process run passes
  in 453.076s; the new drill completes in 193.69s inside that sequence.
- Published PR #54. Its first final-head run passed all checks. CodeRabbit
  reported six review points; five valid findings hardened the outer job
  timeout, trusted multi-RPC checkpoint comparison, explicit checkpoint-height
  interpolation, systemd `ExecStart` reset, and pre-merge status wording. The
  future-date finding was rejected because the runtime date was already July
  23 in both EEST and UTC. Every review thread was answered and resolved.
- PR #54 final head `750040b` passed Go build/race/coverage in 7m04s, the
  expanded multi-validator recovery job in 8m06s, Docker restart in 3m35s,
  docs consistency, Go/Rust/Node security, DeepScan, and CodeRabbit. Squash-
  merged as `3e44905`; GitHub closed GH-53.

## 2026-07-19 03:31 EEST - GH-47 CI build-and-test timeout

- Opened GH-47 after the PR #46 merge-commit Go CI run left `build-and-test`
  in progress for hours while `multi-validator-recovery` and
  `docker-restart-smoke` were green.
- Reproduced the normal suite locally: `go test ./...` PASS in 58.913s.
- Added a 20-minute timeout to the Go CI `build-and-test` job. The test command
  itself is unchanged.
- Pushed the fix to `main` as `63b76bf`. GitHub `main` checks are green:
  Go CI (`build-and-test`, `multi-validator-recovery`,
  `docker-restart-smoke`), Security Scan, and Pages.

## 2026-07-19 02:52 EEST - GH-45 backup/restore/export/import start

- Confirmed `main` is synchronized with `origin/main` at `d121d34`.
- Confirmed no open PRs; only #4, #7, and #29 were open before creating the
  next Phase 1 task.
- Opened GH-45 for the next GH-29 Phase 1 gap: backup, restore, export, and
  import disaster-recovery drills.
- Created branch `feature/GH-45-backup-restore-drill`.
- Hardened `scripts/backup.sh` so chain-data backups exclude node keys,
  validator keys, validator signing state, and keyrings.
- Added `scripts/restore.sh`, which restores sanitized backup data into an
  already initialized target home while preserving the target's local keys.
- Added `TestMultiValidatorBackupRestoreExportImport`, a gated process harness
  that backs up a live full node, verifies the backup archive contains no
  private/signer artifacts, restores it into a fresh home, restarts/catches up,
  verifies app-hash convergence, exports the restored state, validates ledger
  invariants, and re-imports the exported genesis.
- Local evidence so far:
  `bash -n scripts/backup.sh scripts/restore.sh` PASS; `go test . -run
  'TestMultiValidatorBackupRestoreExportImport|TestConfigureGenesisValidatorSetBuildsExactBankBackedSet'
  -count=1 -timeout=300s -v` PASS/SKIP as expected without the smoke env;
  `TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test . -run
  TestMultiValidatorBackupRestoreExportImport -count=1 -timeout=420s -v` PASS
  in 88.224s; `go test ./...` PASS in 58.843s; `bash
  scripts/check-consistency.sh` PASS; `TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1
  go test . -run
  '^(TestMultiValidatorConsensusRecovery|TestMultiValidatorTrustedSnapshotStateSync|TestMultiValidatorBackupRestoreExportImport)$'
  -count=1 -timeout=720s -v` PASS in 290.498s.
- Published PR #46. The first GitHub `multi-validator-recovery` run exposed an
  existing trusted state-sync timeout in the combined CI-smoke command, not a
  failure in the new backup/restore harness.
- Hardened the state-sync waits from 120s to 180s and raised the CI smoke
  timeout from 720s/12m to 900s/15m. Focused local verification:
  `TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test . -run
  TestMultiValidatorTrustedSnapshotStateSync -count=1 -timeout=420s -v` PASS
  in 127.784s.
- Refreshed PR #46 GitHub checks are green: `build-and-test`,
  `multi-validator-recovery`, `docker-restart-smoke`, docs `check`, CodeRabbit,
  DeepScan, Go/Rust security scans, and Node audits.
- Merged PR #46 to `main` as
  `26bf44b7933c25f379db475fd34d2cfb8e49c626`; GitHub automatically closed
  GH-45. GH-29 remains open as the parent rollout tracker.

## 2026-07-19 01:42 EEST - GH-43 trusted snapshot state sync start

- Opened GH-43 for the next GH-29 Phase 1 gap: trusted snapshot state sync
  catch-up.
- Created branch `feature/GH-43-trusted-snapshot-state-sync`.
- Added `TestMultiValidatorTrustedSnapshotStateSync`, a gated process harness
  that enables snapshots on four trusted validators, commits a real
  `create-domain` transaction, derives trust height and trust hash from a
  trusted RPC endpoint, starts a fresh non-validator node with state sync
  enabled, and verifies catch-up to a common app hash.
- Added harness helpers for state-sync-safe start flags, scoped `[statesync]`
  config patching, and trusted block-hash lookup through CometBFT RPC.
- Local evidence:
  `go test . -run
  'TestMultiValidatorTrustedSnapshotStateSync|TestConfigureGenesisValidatorSetBuildsExactBankBackedSet'
  -count=1 -timeout=300s -v` PASS/SKIP as expected without the smoke env;
  `TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test . -run
  TestMultiValidatorTrustedSnapshotStateSync -count=1 -timeout=360s -v` PASS
  in 130.528s; `TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test . -run
  '^(TestMultiValidatorConsensusRecovery|TestMultiValidatorTrustedSnapshotStateSync)$'
  -count=1 -timeout=480s -v` PASS in 197.835s; `go test ./...` PASS in
  65.114s; `bash scripts/check-consistency.sh` PASS.
- Published PR #44, waited for green GitHub checks (`build-and-test`,
  `multi-validator-recovery`, `docker-restart-smoke`, docs check, CodeRabbit,
  DeepScan, Go/Rust security scans, and Node audits), and merged it to `main`
  as `12a37339e9cff957d1b44413aa36160aed4e8d29`.
- GitHub automatically closed GH-43. GH-29 remains open as the parent rollout
  tracker.

## 2026-07-19 00:54 EEST - GH-41 network partition recovery start

- Confirmed `main` is synchronized with `origin/main` at `464a36a`.
- Confirmed no open PRs and latest `main` Go CI, Security Scan, and Pages
  deployment are green.
- Cleaned stale completed issues #8, #11, #13, #14, #20, and #21 with closure
  comments that preserve remaining rollout boundaries under GH-29/GH-7.
- Left #4, #7, and #29 open intentionally as parent/audit/rollout trackers.
- Opened GH-41 for the next Phase 1 rollout gap: network partitions, delayed
  peers, validator failure, and recovery without ledger divergence.
- Created branch `feature/GH-41-network-partition-recovery`.
- Added `TestMultiValidatorNetworkPartitionRecovery`, which starts a 3-of-4
  quorum without the fourth peer, starts the fourth validator isolated with no
  peers, commits a real `create-domain` transaction on the quorum side, then
  reconnects the isolated validator and verifies catch-up to the same app hash.
- The new harness also verifies all validator powers after recovery and exports
  every node state through `validateLedgerGenesis` to prove bank supply,
  treasury/stake escrow, DEX reserve, and LP invariant consistency.
- Local evidence:
  `TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test . -run
  TestMultiValidatorNetworkPartitionRecovery -count=1 -timeout=300s -v` PASS
  in 104.175s; all three gated process harnesses pass together in 392.147s;
  `go test ./...` PASS.
- Published PR #42, waited for green GitHub checks (`build-and-test`,
  `multi-validator-recovery`, `docker-restart-smoke`, docs check,
  CodeRabbit, DeepScan, Go/Rust security scans, and Node audits), and merged it
  to `main` as `8544943dd6fab483884392f1f04e83acbeb8f3f7`.
- GitHub automatically closed GH-41. GH-29 remains open as the parent rollout
  tracker.

## 2026-07-15 13:20 EEST - GH-39 validator lifecycle evidence

- Continued `feature/GH-39-validator-lifecycle` from the active rollout goal.
- Fixed Cosmos SDK v0.50 custom signer resolution for hand-written
  truedemocracy Msgs, including dynamic protobuf messages used during tx decode.
- Replaced the partial BaseApp router-only registry setup with
  `SetInterfaceRegistry` so tx decoding, event generation, gRPC, and message
  routing share the same address codecs and custom signer functions.
- Added delivered-transaction verification to the process harness using
  CometBFT `/tx` RPC, preventing `broadcast-mode sync` from hiding DeliverTx
  failures.
- Fixed process-harness account-number handling for offline signing and added
  explicit RPC/catch-up waits before replacement validator tx submission.
- Proved validator join/replacement with a gated six-node process harness:
  `TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test . -run
  TestMultiValidatorJoinReplacementLifecycle -count=1 -timeout=300s` PASS in
  117.638s.
- Re-ran focused validator lifecycle tests and full repository tests:
  `go test ./x/truedemocracy -run
  'TestValidatorJoinLeaveReplacementUpdates|TestBuildValidatorUpdates|TestRemoveValidator'
  -count=1` PASS; targeted root lifecycle/tx tests PASS; `go test ./...` PASS.
- Remaining boundary: public validator leave is still economically coupled to
  full stake withdrawal, so process-level join/replacement evidence is paired
  with Keeper/ABCI power-zero tombstone regression coverage for leave.
- Published [PR #40](https://github.com/NeaBouli/TrueRepublic/pull/40), waited
  for green GitHub checks (`build-and-test`, `multi-validator-recovery`,
  `docker-restart-smoke`, docs consistency, CodeRabbit, DeepScan, Go/Rust
  security scans, and Node audits), and merged it to `main` as
  `ad30d188c7956c28cff5bf53304bc04848ba569a`.

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

## 2026-07-11 21:09 EEST - Docker recovery and final GH-11 audit

- Replaced the Alpine/musl node image path in PR #9 with a Debian/glibc builder
  and runtime, copied the architecture-specific wasmvm shared library, and ran
  `ldconfig` in the runtime image.
- Verified both GitHub Docker builds and all other PR #9 gates pass; converted
  PR #9 from draft to ready and requested the required independent review from
  `xxlfan72` and `ijuedt`.
- Audited every GH-11 denomination boundary and found a production validator
  tree stake that had been renamed to `upnyx` without base-unit scaling.
- Corrected the default stake from 0.1 PNYX to 100,000 PNYX and added a
  regression test across all seven generated nodes.
- Corrected Compose, node-init, environment, operator-documentation, and
  maintained-client gas prices so the migration preserves economic values.
- Normalized canonical bank metadata by removing conflicting legacy PNYX
  metadata while preserving unrelated asset metadata; expanded tests.
- Full Go build/race/vet, Rust test/Clippy/audit, maintained-client
  test/lint/build/audit, documentation consistency, and diff checks pass.
- The added validator-tree regression brings the recovery-verified total to
  573: 541 Go cases, 26 Rust tests, and six maintained-client tests.

## 2026-07-11 22:08 EEST - GH-14 bank-escrow audit remediation

- Rebased `fix/GH-14-bank-escrow` onto the fully green PR #15 head so PR #16
  remains an ordered, reviewable custody-only stack.
- Audited bank escrow, cache-context atomicity, signer claims, treasury payouts,
  validator stake settlement, slashing, CLI construction, and CosmWasm bindings.
- Found and fixed a high-severity contract regression: three CosmWasm encoders
  omitted the authenticated `Sender` required by hardened messages.
- Found and fixed a high-severity custody flaw: slashed stake was credited to an
  admin-withdrawable domain treasury. Penalties now burn exact module escrow.
- Added the minimum `burner` module permission plus exact-burn, burn-failure
  rollback, permission, missing-signer, and binding-validation regression tests.
- Updated the token/ledger audit to mark the GH-11 bank-genesis cap and GH-14
  custody slices remediated while retaining GH-13/GH-10/GH-12 blockers.
- Go build, vet, race, coverage, focused governance tests, and 557 full Go cases
  pass locally. Rust test/Clippy/audit, maintained-client install/lint/test/build/
  audit, documentation consistency, and diff checks also pass. The branch total
  is now 589 with 26 Rust and six maintained-client tests.
- Published PR #16 head `d3ae4cf`; both Go/Race and Docker runs, Rust, client,
  docs, DeepScan, and the manually triggered full CodeRabbit review completed.
- Accepted CodeRabbit's slash-atomicity hardening and three stale documentation
  findings. Rejected only its obsolete suggestion that refreshed PR #15 checks
  were pending because live GitHub evidence shows them green at `e0ff339`.

## 2026-07-11 23:02 EEST - GH-13 canonical issuance rebase and audit

- Rebasing retained only the GH-13 implementation commit on top of verified PR
  #16 head `fa693a8`; stale GH-13 documentation commits were intentionally
  skipped and reconstructed from current evidence.
- Audited canonical supply reads, minter/burner permissions, cap clipping,
  validator/domain allocation, interval snapshots, vote transfers, slashing,
  EndBlock ordering, and mint/burn failure paths against whitepaper equations
  4/5 and the 21,000,000 PNYX rule.
- Found and fixed a high-severity abstraction bypass: validator slashing now
  uses `token.IssuanceService.Burn` rather than calling the bank keeper directly.
- Found and fixed a high-severity partial-settlement boundary: staking and
  domain issuance now share one outer EndBlock cache, with a second-mint failure
  regression proving claims, timers, and snapshots roll back together.
- Added missing-bank, nil/negative supply, failed burn, and second-mint failure
  coverage. Canonical supply input is validated before cap calculations.
- Go build, vet, 567 cases, race, and coverage pass. Current coverage is root
  10.2%, token 93.5%, treasury 97.0%, DEX 34.2%, and governance 55.8%.
- Updated the structured audit, public DEX limitations, canonical reward docs,
  599-test status, bridge decisions, security notes, and PR #17 audit record.
- Hardened the node Dockerfile to derive `libwasmvm` from Docker's target
  architecture (`amd64`/`arm64`) instead of host `uname`, added runtime linkage
  validation and `libgcc-s1`, and injected the immutable GitHub commit as the
  binary version.
- Added `.dockerignore`; the excluded Rust target and installed client
  dependencies alone reduce the prior 1.6 GB local build context by more than
  1.5 GB and prevent local environment/key files from entering the context.
- Re-ran Rust Clippy and audit after all changes. Clippy passes; audit reports
  no vulnerability and the same six allowed transitive dev-tooling warnings.
  Maintained-client and documentation gates pass; GitHub must execute the
  actual Docker build because this workstation has no Docker engine.
- The first GitHub image execution proved wasmvm selection/linkage correct but
  exposed an older application startup panic: `legacytx.StdTx` was registered
  directly and then a second time by the auth codec. Removed the duplicate,
  unified application/CLI codec construction, exposed the injected build
  version through Cobra, and added the 567th Go regression test.
- Published `b738d70`. Both refreshed Docker builds now pass their linkage and
  CLI-start smoke check; both Go jobs (including race/coverage), docs, DeepScan,
  and the complete manually triggered security workflow are green.
- Thread-aware GitHub inspection reports zero unresolved review threads. The
  prior full CodeRabbit review completed without findings; the incremental
  startup-fix refresh was acknowledged but temporarily rate-limited.
- The accepted final 33-file review found five inline issues plus two additional
  findings. Made mock-bank mint/burn deltas cache-backed and extended the
  second-mint regression to prove supply/escrow rollback and parity.
- Initialized restored-domain payout snapshots at genesis and lazily backfilled
  pre-GH-13 state, preventing historical payout windfalls. Added genesis and
  upgrade-compatible lazy-backfill regressions.
- Added the container `--version` smoke, reconciled PNYX-cap decisions, DEX
  status, equation signatures, and branch-head handover wording. Full local Go
  build/vet, 569 cases, focused race, coverage, and docs consistency pass.
- Published `0e6cf38`, replied to every inline finding with implementation and
  test evidence, and recorded the two non-inline fixes in the PR conversation.
  Both Go jobs, both Docker builds, docs, DeepScan, CodeRabbit, and the manually
  triggered security workflow pass; all five review threads are resolved.
## 2026-07-12 03:34 EEST - GH-10 DEX custody audit and verification

- Rebased `fix/GH-10-dex-custody` onto final PR #17
  head `29fb228`, preserving both governance and DEX module burn permissions.
- Audited direct and two-hop DEX value paths: exact
  account/module transfers, reserve changes, intermediate PNYX routing, burn
  deltas, LP ownership, authority checks, slippage, and cached commit boundaries.
- Fixed a valid-denom LP key collision by replacing
  textual prefixes with length-prefixed keys; added `atom`/`atom:staked`
  conservation regression coverage.
- Added KV-backed failure regressions for create/add/
  remove/swap/burn paths, proving account, pool, LP, analytics, and supply
  rollback on every settlement boundary.
- Verified 578 Go cases plus race/vet/build/CLI smoke,
  26 Rust tests and audit, six maintained-client tests plus lint/build/audit,
  Go module integrity, docs consistency, and clean diff/secret checks. Docker
  remains delegated to GitHub because no local Docker CLI is installed.
- Published rebased PR #18 head `3234741`; updated Issue #10 acceptance state,
  recovery epic #4, PR metadata, BRIDGE, audit, README, status JSON, and GitHub
  Pages source.
- GitHub docs, DeepScan, Go build/vet/race/coverage, and the real Docker build
  pass. Manually dispatched Security Scan run `29156922464`; govulncheck, Rust
  audit, canonical npm audit, and both informational legacy audits all pass.
- Requested focused CodeRabbit review. The service reported its quota exhausted
  for 44 minutes, so substantive external review remains explicitly pending.

## 2026-07-12 04:32 EEST - GH-12 genesis and runtime-invariant audit

- Rebased GH-12's code commit onto final PR #18 and discarded three obsolete
  documentation commits for evidence-based regeneration.
- Preserved the CLI Amino panic fix, both issuance tests, cache-aware bank mocks,
  and GH-10's collision-free LP keys while adapting global LP export/orphan
  detection to the new length-prefixed format.
- Reproduced and fixed the prototype's critical hard-coded validator secret.
  Production defaults are empty; InitChain accepts actual CometBFT Ed25519
  public keys and creates exact cap-checked module stake, or rejects startup.
- Replaced silent InitGenesis skips with explicit failure and fail-closed JSON
  export behavior.
- Expanded full-app evidence from escrow-only to independent supply, escrow,
  reserve, and LP invariant halts. Added non-empty bank/treasury/stake/pool/LP
  export-import preservation plus over-cap, duplicate, negative, and unbacked
  rejection tests.
- Verified 615 Go cases, race, vet, build, coverage (root 66.1%, token 92.6%,
  treasury 97.0%, DEX 45.3%, governance 56.6%), 26 Rust tests/audit, six
  maintained-client tests/lint/build/audit, CLI smoke, and module integrity.
- Recorded the residual GH-21 blocker: `scripts/init-node.sh` still invokes the
  unavailable `x/staking` gentx flow and must not launch production nodes.
- Published rebased PR #19 head `9d521ce`; synchronized Issue #12, recovery
  epic #4, PR metadata, BRIDGE, audit, README, status JSON, and GitHub Pages
  source.
- GitHub Docs, DeepScan, Web, Mobile, Rust, Go build/vet/test, and both Docker
  builds pass. Manually dispatched Security Scan run `29158360390`; all five
  jobs pass. Requested CodeRabbit review; its independent result is pending.

## 2026-07-12 05:12 EEST - GH-20 ZKP/authentication audit

- Rebased GH-20's single code commit onto final PR #19 and discarded two stale
  documentation snapshots for evidence-based regeneration.
- Preserved chain/rating signal binding, rating-independent chain/proposal
  nullifiers, transaction-time setup removal, canonical commitment checks, and
  altered-rating/cross-chain regressions.
- Found and fixed missing active-nullifier export. Genesis now round-trips exact
  nullifier records/heights without resurrecting values cleared by Big Purge.
- Pinned genesis VK configuration by supported circuit ID, SHA-256, BN254 curve,
  four-public-input shape, canonical serialization, and no trailing bytes.
- Recomputed genesis identity Merkle roots and rejected malformed commitments,
  history, rating authentication fields, nullifiers, and proof public inputs.
- Found and disabled a legacy wallet path that generated random mock proof bytes
  and called `signAndBroadcast` while claiming anonymity. Both maintained and
  legacy web clients now fail closed and display non-submittable preview status.
- Verified 643 Go cases, race, vet, build, coverage (root 66.1%, token 92.6%,
  treasury 97.0%, DEX 45.3%, governance 58.9%), 26 Rust tests/audit, eight
  maintained-client tests/lint/build/audit, four focused legacy ZKP tests/build/
  audit, module integrity, and diff/secret/setup checks.
- Published rebased PR #22 head `6732276`; synchronized Issue #20, recovery
  epic #4, PR metadata, BRIDGE, audit, README, status JSON, and GitHub Pages.
- GitHub Docs, DeepScan, Web, Mobile, Rust, both Go/Docker runs, and manual
  Security Scan run `29159603247` pass. CodeRabbit is check-green but explicitly
  rate-limited and produced no substantive independent cryptographic review.
- Rebased the six audited GH-21 implementation commits without content drift
  onto final GH-20 bridge head `fac50a4`.
- Reproduced and fixed a release regression introduced by the lifecycle move:
  `truerepublicd --version` failed and `truerepublicd version` was blank because
  Cobra/Cosmos SDK application metadata was no longer wired.
- Verified 649 Go cases and coverage (root 64.3%, token 92.6%, treasury 97.0%,
  DEX 45.3%, governance 58.9%), a real native block/SIGINT/same-home restart/
  export flow under race, vet, CGO build, both version interfaces, entrypoint
  syntax, and diff checks.
- Prepared the local GH-21 audit/bridge/public-status evidence at 683 total
  recovery cases. Local Docker and ShellCheck are unavailable; refreshed
  GitHub Docker restart and security jobs remain mandatory before approval.
- Published audited GH-21 head `ec1ce17`; synchronized PR #23, Issue #21, and
  recovery epic #4 with the 649 Go / 683 total evidence and version regression.
- GitHub Go build/vet/race/coverage and Docker block/restart run `29170712626`,
  Docs, DeepScan, Web, Mobile, Rust, and all five manual Security Scan
  `29170832988` jobs pass. Independent multi-node operations review remains.
- Created isolated GH-8 checkout and rebased only its CI/docs commits onto
  final GH-21 `b59efa2`, discarding the obsolete link-only 636-case handoff.
- Combined Action-major upgrades with read-only, non-persisted checkout
  credentials, Node 22 for the maintained client, main-only push automation,
  PR/manual feature execution, and no duplicate branch/PR suites.
- Found the docs gate silently checked nonexistent `wiki-github/` paths while
  the real wiki claimed v0.1-alpha, 182 tests, Go 1.23, Testnet Ready, usable
  anonymity/mobile, and no high/critical issues. Replaced those claims with
  evidence-backed 683-case recovery status and explicit blockers.
- Corrected the recovery installation path to select the branch that actually
  provides GH-21 lifecycle commands; created real wiki current/testing pages
  and reconciled linked architecture/operator toolchain facts.
- Local workflow YAML, docs consistency/arithmetic, JSON, wiki target,
  stale-current-claim, and diff checks pass. GitHub Action execution is pending.
- Published rebased GH-8 PR #24 head `3964f4a`; synchronized PR metadata,
  Issue #8, and recovery epic #4 with the 683-case docs/wiki/CI audit findings.
- GitHub Go race/coverage + Docker restart `29171461365`, Rust `29171461357`,
  Web `29171461355`, Mobile `29171461342`, Docs `29171461348`, DeepScan,
  CodeRabbit, and all five Security Scan `29171476126` jobs pass.
- PR #25 remains draft/red against unrecovered main; no gate bypass is allowed.

## 2026-07-12 12:35 EEST - GH-12 review remediation and stack refresh

- Accepted both actionable PR #19 review findings. Because `CreateDomain` has
  no error return, the escrow-divergence test now reads the domain back and
  validates its treasury, so a failed fixture cannot masquerade as invariant
  coverage.
- Expanded the production-node limitation to require a real Ed25519 key bound
  to a positive-power CometBFT validator, sufficient exact bank-backed stake,
  and preservation of the 21,000,000 PNYX canonical supply cap.
- Focused registered-invariant regression, full Go tests, vet, and build pass.
  Published commit `eec91c7`, answered and resolved both PR #19 review threads,
  rebased/published PR #22 at `0c72ad0`, rebased/published PR #23 at `49938a3`,
  and published the propagated PR #24 stack head.
- PRs #19, #22, #23, and #24 are mergeable on exact consecutive bases with no
  unresolved review threads. Their refreshed standard checks pass, as do all
  five jobs in Security runs `29172007410`, `29172246257`, `29172246373`, and
  `29172246235` respectively.
- PR #9 remains technically green and mergeable but correctly requires one
  independent approval. PR #25 remains red against unrecovered `main`; neither
  gate is bypassed, and meaningful recovery work remains available.

## 2026-07-12 12:48 EEST - GH-26 operator init recovery

- Inventoried all open issues/PRs and found a remaining operator footgun:
  `scripts/init-node.sh` still invoked unavailable `x/staking` gentx commands,
  wrote a mnemonic capture file, and was recommended by public install docs.
- Opened Issue #26 and isolated `fix/GH-26-pod-init-script` from final PR #24.
- Replaced the wrapper with a single quoted `truerepublicd init` delegation;
  retained chain/moniker/home, minimum gas price, and Prometheus configuration.
- Added a regression proving the exact command boundary, absence of all legacy
  account/gentx actions and mnemonic files, and both configuration edits.
- Reconciled operator/public/security documentation and advanced the verified
  source of truth to 684 cases (650 Go + 26 Rust + 8 maintained client).
- Focused test, real compiled-daemon wrapper smoke, full Go suite, vet, shell
  syntax, documentation/JSON consistency, and diff checks pass. The real smoke
  proves one consensus validator, matching custom PoD identity, canonical
  `upnyx` bank supply, no mnemonic artifact, and configured gas/Prometheus.
- Publication and GitHub Go/Docker/security verification remain in progress.
- Published implementation/audit head `86ff1c8` and opened stacked draft PR
  #27 against final PR #24. Refreshed GitHub gates are running.
- GitHub Go race/coverage and Docker restart run `29172845624`, Docs
  `29172845627`, DeepScan, CodeRabbit, and all five Security Scan
  `29172846057` jobs pass. PR #27 is mergeable with zero review threads.
- Completion audit found GitHub Pages still built `main:/docs` at old head
  `d8545cf`. Changed only the Pages source to the fully green recovery branch
  `fix/GH-26-pod-init-script:/docs` and explicitly queued build `1090733247`.
- The build completed without error at `50b0d9a`. Live HTTP verification shows
  the recovery warning, non-production boundary, 21M maximum supply, and 684
  verified cases. PR #25 and branch protection were not bypassed.

## 2026-07-14 00:28 EEST - GH-32 multi-validator recovery implementation

- Reopened GH-29 after confirming PR #31 closed the roadmap handoff rather than
  the seven rollout phases; created child Issue #32 with explicit bounded scope.
- Extracted single-validator genesis binding into a reusable internal public-key
  set assembler while preserving `init` refusal to overwrite an existing set.
- Added four independently generated validator homes and keys, one identical
  exactly bank-backed PoD genesis, explicit persistent peers, common-height
  app-hash checks, one-validator failure, three-validator continued quorum,
  restart/catch-up, clean shutdown, export, and exported-ledger validation.
- The first run correctly failed because strict CometBFT address books reject
  loopback peers and every node inherited pprof port 6060. Restricted
  address-book/duplicate-IP relaxation to temporary localhost configs and
  disabled pprof only for harness processes; production defaults are unchanged.
- Targeted genesis/binder tests pass. The full normal Go suite passes with 651
  cases. The separately gated four-validator harness passes twice locally, most
  recently in 55.84 seconds.
- Added a dedicated `multi-validator-recovery` Go CI job, operator runbook,
  685-case public source-of-truth update, and exact remaining-scope warnings.
- Full Go race/coverage passes: root 64.9%, token 92.6%, treasury 97.0%, DEX
  45.3%, and governance 58.9%; build and vet also pass.

## 2026-07-14 02:05 EEST - GH-32 review remediation

- PR #33's initial GitHub matrix passed, including the dedicated four-validator
  job, Go/Docker, documentation, security, Node audit, and static-analysis gates.
- Accepted review findings for an ambiguous public roadmap bullet, missing
  subprocess/request cancellation, and a recovery assertion that only checked
  the last survivor-produced height.
- The harness now threads the test context through every child process and HTTP
  request, selects a target two blocks after the stopped process restarts, then
  requires all four nodes to reach it and agree on that post-restart app hash.
- Rejected the proposed `actions/checkout@v7` edit: the official action has no
  v7 release, and a runtime-major migration is separate from this bounded
  network-recovery ticket. The new job remains on the repository's v5 baseline.
- Genesis/binder regressions pass and the hardened real four-validator harness
  passes in 68.90 seconds, its third successful local run.

## 2026-07-14 04:03 EEST - GH-32 merged and rollout tracker advanced

- Published review-remediation commit `7dd0fbb`; all four CodeRabbit threads
  are answered or automatically recognized as addressed and are resolved.
- The first final-head attempt failed four jobs before checkout because GitHub
  could not resolve action download metadata and returned `Service Unavailable`.
  Job-level logs confirmed no project command ran, so no code was changed.
- Reran only failed jobs. Go run `29253316692` passes build/race/coverage,
  Docker restart, and multi-validator recovery; Security run `29253316707`
  passes Go vulnerability, Rust, maintained-client, and both legacy audits.
  Docs, DeepScan, and CodeRabbit also pass.
- Squash-merged PR #33 as `9d68a6f`; GitHub automatically closed Issue #32.
  GH-29 remains open and its first Phase 1 checkbox now links the evidence.

## 2026-07-14 04:16 EEST - GH-32 final main and Pages proof

- Squash-merged bridge closure PR #34 as `2851759`; no pull requests remain
  open, GH-32 is closed with all nine criteria checked, and GH-29 remains open
  with only the completed multi-validator item checked.
- Final `main` Security run `29261145077` passes all five jobs. The previously
  documented upstream-only govulncheck annotations remain non-blocking.
- GitHub Pages build `1093339877` completed from `main:/docs` at exact commit
  `2851759`. Live HTTP verification confirms the recovery/non-production
  warning, 21M cap, 685 verified cases, and four-validator recovery evidence.

## 2026-07-15 20:10 EEST - GH-37 Codex subagent role configuration

- Opened GH-37 to track the project-scoped Codex agent role split.
- Added `.codex/config.toml` with one-level subagent nesting and a six-thread
  cap for predictable delegation.
- Added `.codex/agents/spark-worker.toml` as a narrow
  `gpt-5.3-codex-spark` worker for small bounded patches, file search, and
  focused checks.
- Documented that the main Codex agent keeps architecture, risk, integration,
  final verification, GitHub updates, merges, and Bridge responsibility.
- Verified both `.codex` TOML files parse with Python `tomllib`; `git diff
  --check` and `bash scripts/check-consistency.sh` pass.
- PR #38 opened for GH-37. GitHub Docs Consistency, Security Scan, and DeepScan
  pass; CodeRabbit remained pending without comments during the final merge
  decision.

## 2026-07-22 23:04 EEST - GH-48 fast post-merge audit

- Confirmed local `main` and `origin/main` were identical at `6a308c5`, with no
  open pull requests and only GH-4, GH-7, and GH-29 open before this task.
- Confirmed the latest scheduled Security Scan at `6a308c5`, the latest Pages
  build, and the latest code-bearing Go CI at `63b76bf` are green.
- Opened GH-48 after finding live documentation that still described the
  already merged PR #9 through PR #27 recovery sequence as drafts/unmerged.
- Corrected the contributor guide, root audit, DEX guide, ZKP limitation,
  security queue, and live project state without altering historical logs.
- Recorded three residual warnings: remaining GH-29 rollout evidence, root Go
  wildcard discovery in installed frontend dependencies, and the 1,678.30 kB
  maintained-client main chunk.
- Maintained-client install/lint/8 tests/build/audit, Rust format/Clippy/26
  tests, isolated Go build/vet/655 tests, docs consistency, shell syntax, JSON,
  workflow YAML, and diff checks pass. `cargo audit` exits cleanly with the six
  already documented allowed transitive warnings. Docker and `govulncheck` are
  unavailable locally, so current green GitHub security/Docker evidence remains
  required on the final published head.

## 2026-07-22 23:14 EEST - GH-50 GO-2026-5970 remediation

- PR #49's current vulnerability database found reachable GO-2026-5970 through
  Unicode normalization used by the ZKP dependency path. The affected indirect
  module was `golang.org/x/text` v0.37.0; the scanner reports v0.39.0 as fixed.
- Opened GH-50 and updated `golang.org/x/text` to v0.39.0. Module resolution
  also advances `golang.org/x/sync` to v0.21.0 and records the already directly
  imported `cosmossdk.io/x/tx` in the direct require block.
- A current local `govulncheck` no longer reports GO-2026-5970. Four reachable
  upstream findings remain with `Fixed in: N/A`; the existing Security Scan
  gate continues to fail only when a reachable fixable version exists.
- Exact CI-filter reproduction passes: no reachable finding with an available
  fix remains. Go build, vet, and the full 655-case suite pass after the update;
  the root package completes in 64.499 seconds.
- Accepted CodeRabbit's audit-tally finding: the document contains 17 distinct
  `[PASS]` findings, so the root audit and live project state now report
  0 FAIL / 3 WARN / 17 PASS.
- PR #49 final head `2bd5efd` passes Go build/race/coverage (6m49s), combined
  multi-validator recovery (5m42s), Docker restart (3m09s), Go vulnerability,
  Rust audit, maintained and legacy Node audits, docs consistency, DeepScan,
  and CodeRabbit. The one valid review thread is resolved.
- Squash-merged PR #49 as `7dbde858d0d3d5410f22d16a1a3bac614325d925`;
  GitHub closed GH-48 and GH-50. At the merge point, only GH-4, GH-7, and
  GH-29 remained open.
- Opened GH-51 for the remaining medium package-selection finding and linked
  the complete audit/remediation evidence from GH-29 and GH-4.

## 2026-07-22 11:24 UTC - GH-51 root Go package isolation

- Reproduced the issue: root `go list ./...` discovers
  `truerepublic/client-web/node_modules/flatted/golang/pkg/flatted` after the
  maintained-client dependencies are installed.
- Added `scripts/go-packages.sh`, which derives explicit package directories
  only from Git-managed, non-ignored Go source and defensively excludes
  `node_modules` and `vendor` trees. Makefile verification, contributor/agent
  guidance, and GitHub Go CI now use that same selector.
- Added `scripts/test-go-packages.sh`; its ignored dependency-source probe
  leaves the selected five repository packages unchanged.
- Ran `npm ci` concurrently with selector verification, build, vet, and the
  complete race/coverage suite. All pass; root coverage is 65.9%, and npm
  reports zero vulnerabilities. The normal 655-case suite also passes.
- The combined consensus recovery, trusted state-sync, and sanitized
  backup/restore/export/import harnesses pass in 265.381s. Compose/Docker cannot
  run locally because the Docker executable is unavailable, so final-head
  GitHub Docker evidence remains required before merge.
- CodeRabbit reported four valid findings: normalize the Bridge timestamp,
  remove the resolved medium-priority duplicate, ignore missing indexed source,
  and return to the repository root in the wiki test sequence. All were fixed
  in `6a668ee`, answered, and resolved.
- PR #52 final head passes all 11 checks: Go build/vet/race/coverage (7m11s),
  multi-validator recovery (5m44s), Docker restart (3m20s), Go vulnerability,
  Rust audit, maintained and legacy Node audits, docs consistency, DeepScan,
  and CodeRabbit. PR #52 was squash-merged as `ae7105a`; GH-51 closed.
