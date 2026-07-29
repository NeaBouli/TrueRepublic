# Action Log

## 2026-07-23 19:40 EEST - GH-59 ABCI++ validator slashing

- Wired BaseApp's deterministic BeginBlock ABCI++ boundary to CometBFT
  misbehavior and decided-last-commit ingestion. Canonical commit cursors,
  normalized infraction IDs, processed markers, consensus-key active intervals,
  tombstones, and operator-scoped rolling liveness bitmaps now survive restart
  and export/import.
- Equivocation burns 5% exactly once per tombstoned consensus key, jails the
  operator, emits power zero, and remains attributable after key rotation.
  Downtime burns 1% only after more than 50 misses in a complete rolling
  100-block window, then jails and emits power zero.
- Closed the exit-before-evidence path: authenticated full removal and full
  withdrawal retain the complete stake in module escrow until both CometBFT
  evidence-age limits are strictly exceeded. Delayed evidence slashes the held
  claim before payout.
- Extended genesis validation/export/import for historical keys, replay state,
  liveness, jail state, processed infractions, and pending exits. A jailed
  below-minimum validator now round-trips without being resurrected or making
  the exported genesis invalid.
- Added the operator slashing/incident runbook and the gated real four-process
  harness. The harness proves 100-block downtime, exact 1% burn, restart and
  catch-up as a full node, real CometBFT duplicate-vote evidence broadcast,
  exact 5% burn, both power-zero transitions, common app hash, and ledger-valid
  export. The first diagnostic run used production's five-second commit
  interval and correctly exceeded a four-minute observation deadline; the
  test-only temporary homes now use bounded accelerated consensus timeouts.
  The complete real-process rerun passes in 200.95 seconds.
- Local normal `go test ./...` passes. Focused ABCI++, replay, rotation,
  liveness, exit-hold, and export/import regressions pass. Recovery arithmetic
  is 726 cases: 692 Go, 26 Rust, and eight maintained-client. The first
  adversarial review exposed an Unjail cursor halt, historical-key/hold reuse,
  partial-withdrawal custody, replay-binding, and genesis-relation gaps; all are
  remediated with focused regressions. The final review found 0 P0 / 0 P1 /
  0 P2. PR #65 passed all 11 final-head checks and merged as `934a042`; GH-59
  closed and GH-29 was synchronized.

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

## 2026-07-23 12:40 EEST - GH-56 local implementation and audit

- Implemented authenticated atomic validator consensus-key rotation with a
  signed expected-old-key precondition, permanent old-key revocation, and one
  deterministic H to H+2 CometBFT transition while preserving operator claims.
- Separated bootstrap operator authority from consensus identity, rejected
  self/cross/historical consensus-derived authorities and module accounts, and
  materialized missing public operator base accounts without generating secret
  material.
- Added delayed old-key evidence attribution, one-shot/deferred removal
  handling, genesis/export/import validation, CLI/API support, Docker/operator
  plumbing, and a dedicated rotation/compromise-response runbook.
- Added the real five-process rotation harness. Four active validators rotate
  to a pre-synchronized stopped-key replacement; the old signer stays offline,
  activation occurs at H+2, claims and app hash converge, old-key reuse fails,
  and export/re-import succeeds. Final local run PASS in 168.12 seconds.
- Full normal Go suite, focused authority/transition regressions, repository Go
  vet, docs consistency, shell syntax, and diff checks pass. Verified recovery
  arithmetic is now 717 cases: 683 Go, 26 Rust, and eight maintained-client.
- Final adversarial review found no P0. Residual rollout warnings are isolated
  in GH-59 (ABCI++ evidence/slashing), GH-60 (inactive validator round trip),
  and GH-61 (legacy authority migration). See `GH56_AUDIT.md`.
- PR #62's first combined process run exposed a CI-only harness setup defect:
  isolated test selection had not configured the Cosmos SDK global prefix, so
  legacy smoke helpers emitted `cosmos1...` operators that the hardened init
  correctly rejected. `smokeOperatorAddress` now encodes `truerepublic`
  explicitly, independent of test order. The previously first-failing
  four-validator recovery harness passes in isolation in 99.33 seconds.

## 2026-07-23 03:06 EEST - GH-56 GitHub closure

- PR #62 final head `239cc6f` passed all enforceable GitHub checks: Go
  build/vet/race/coverage in 6m44s, the combined multi-validator matrix in
  9m39s, Docker restart in 3m20s, Docs, Go/Rust/Node security, and DeepScan.
- CodeRabbit was rate-limited and produced no content review; the independent
  adversarial final review recorded in `GH56_AUDIT.md` found no P0 and no
  additional P2 finding.
- Squash-merged PR #62 as `80ab6741423049d6faa3bfc8d2671feebb314134`.
  GitHub closed GH-56; GH-59, GH-60, and GH-61 retain the explicit residual
  rollout boundaries.

## 2026-07-24 00:25 EEST - GH-59 final local merge gate

- Completed deterministic ABCI++ evidence and decided-last-commit ingestion,
  exact equivocation/downtime penalties, consensus-key history, tombstones,
  replay cursors, rolling liveness, and evidence-window exit custody.
- Added fail-closed genesis/export/import relations across validators, signing
  state, revocations, rotations, infractions, and pending exits. Mature payout
  now removes signing state atomically; partial withdrawal remains disabled
  until generalized slashable unbonding exists.
- A real four-process harness proves 100-block downtime, restart/catch-up,
  cryptographically valid duplicate-vote evidence, exact burns, validator
  power-zero transitions, app-hash convergence, and ledger-valid export.
- Fresh local checks pass: docs consistency at 726 tests and the 21M cap,
  package selection, all five Go packages (root/process package 158.108s),
  build, vet, diff checks, and the complete race/coverage gate. The latter
  reports no race and 68.5% root / 61.8% `x/truedemocracy` coverage.
- The final independent read-only merge-gate review reports 0 P0, 0 P1, and
  0 P2. Detailed invariants and evidence are recorded in `GH59_AUDIT.md`.
- Published commit `c8a56f8` in PR #65; the closure evidence follows below.

## 2026-07-24 00:50 EEST - GH-59 GitHub closure

- PR #65 final head `313b327` passed all 11 checks: Go build/vet/race/coverage
  in 7m05s, the combined multi-validator process matrix in 11m46s, Docker
  restart in 3m15s, docs, Go/Rust/Node security, DeepScan, and CodeRabbit.
- Zero unresolved review threads remained. CodeRabbit was rate-limited and
  produced no content review; the recorded independent final audit found
  0 P0 / 0 P1 / 0 P2.
- Squash-merged PR #65 as `934a042017e860852e945f1f82ca2da571fcff86`.
  GitHub closed GH-59. GH-60 and GH-61 remain the bounded consensus-state
  follow-ups under the still-open GH-29 rollout tracker.

## 2026-07-26 19:40 EEST - GH-60 inactive validator genesis start

- Reconstructed the task from clean `origin/main` at `996ae05`, GH-60, the
  global and local rules, and the complete current Bridge state. GH-59 remains
  closed and is not being repeated.
- GH-60 requires exact retained inactive validator state across export/import:
  custody, complete domain membership, jail, missed blocks, power semantics,
  rotations, revocations, signing/infraction state, pending exits, ledger
  backing, and cap parity. Resurrection and historical-key reuse must fail
  closed.
- Created `feature/GH-60-inactive-validator-genesis`; no unrelated local user
  changes were present.
- Adopted the Sol/Kimi role split. Kimi receives one bounded, secret-free local
  implementation slice; Sol retains architecture, security, integration,
  external writes, complete testing, review, PR/merge, and Bridge ownership.
- Verification has not run yet for GH-60. Next: inspect exact genesis paths,
  start the Kimi slice, review its diff, add remaining integration/process
  evidence, and run the complete relevant gate chain.

## 2026-07-26 20:50 EEST - GH-60 local merge gate

- Kimi K3 implemented the bounded genesis core: new exports explicitly carry
  active/inactive classification, all domains, and stored power while valid
  legacy records keep their previous single-domain, stake-derived semantics.
- Sol reviewed every changed line and added process-level recovery to the real
  four-validator slashing harness. The drill proves exact inactive export and
  re-import after downtime and double-sign penalties, with ledger parity and
  no stake/domain/jail/liveness/power drift. PASS in 441.52s.
- The first real drill exposed only an existing harness-timeout flake: the
  loaded host could not advance the 100-block liveness window within 120s. The
  liveness phase now has 300s headroom; the repeated full drill passed.
- Kimi's independent read-only review found one P2 fail-closed gap: an explicit
  inactive claim could be unjailed with positive power when its domains were
  empty. Sol now rejects every inactive unjailed positive-power record and
  added a dedicated regression. The refreshed review reports 0 P0 / 0 P1 /
  0 P2.
- Final exact local CI chain passes: package selector; all five Go packages;
  build; vet; race/coverage (68.5% root, 62.2% truedemocracy); docs consistency
  at 733 cases and 21,000,000 PNYX; and `git diff --check`.
- Added `GH60_AUDIT.md`; updated current Bridge/project/TODO state and public
  recovery totals from 726 to 733 (699 Go, 26 Rust, eight maintained-client).
- GH-60 remains open. Next: publish the branch, open/update the PR and GH-29,
  require final-head GitHub CI/review, remediate any finding, then merge and
  record closure. No production rollout approval is claimed.

## 2026-07-26 21:15 EEST - PR #67 maintained-client Security Scan remediation

- Published GH-60 commit `cfe065e` and opened PR #67. Updated GH-60 and GH-29
  with the local process/race/review evidence.
- GitHub's maintained-client npm audit failed twice after successful clean
  installs. The retiring quick endpoint returned HTTP 400; querying the
  official bulk advisory dataset exposed new High findings in
  `brace-expansion <=5.0.7` and `postcss <=8.5.17`.
- Updated direct PostCSS to `^8.5.23`; pinned `brace-expansion` to patched
  `5.0.8`; unified the old ESLint transitive minimatch path at `10.2.5` so the
  patched brace-expansion CommonJS contract remains compatible.
- Fresh `npm ci`, lint, eight tests, TypeScript/Vite production build, and
  `npm audit --audit-level=high` pass. Two Moderate React Router advisories
  remain visible and require a separately reviewed breaking v7 migration;
  they do not fail the existing High gate.
- Next: publish the lockfile remediation and require every final-head GitHub
  check and review to refresh before merge.

## 2026-07-26 21:17 EEST - PR #67 review remediation

- Triaged every CodeRabbit finding. Synchronized GH-60 local-complete versus
  publication-pending records, refreshed the rollout snapshot date, documented
  the uncached Go command and GH-60 recovery gate, pinned validation-test error
  reasons, and scoped the panic assertion only around `InitGenesis`.
- Softened an inaccurate inactive-state comment: stored power zero is itself
  an explicit non-consensus state and remains valid even for an unjailed,
  sufficiently staked member.
- Rejected the suggestion to derive export activity from domain/stake
  eligibility. The runtime consensus classification is exactly unjailed plus
  positive stored power; reclassifying malformed positive-power state as
  inactive would still fail import validation and would hide corruption rather
  than preserve it exactly.
- Focused governance tests, docs consistency at 733 cases / 21M PNYX, and diff
  checks pass. Next: run the complete final-diff Go gate, publish, answer and
  resolve review threads, then require refreshed GitHub checks before merge.

## 2026-07-26 21:25 EEST - PR #67 final review-diff gate

- All five uncached Go packages pass, followed by Vet, CGO build, complete race
  and coverage, docs consistency, and diff checks.
- Coverage is 68.5% root, 92.6% token, 97.0% treasury, 45.3% DEX, and 62.2%
  governance.
- No local blocker remains. Next: publish the review remediation, resolve its
  GitHub threads with evidence, and wait for the refreshed final-head gates.

## 2026-07-26 21:40 EEST - GH-60 closure

- Published final review remediation as `0cad6df`, answered and resolved all
  four CodeRabbit threads, and preserved fail-closed runtime activity
  classification rather than masking corrupt state during export.
- All 12 final-head GitHub checks passed: Go build/vet/race/coverage in 7m14s,
  combined real multi-validator recovery in 11m55s, Docker restart in 3m33s,
  maintained-client build, docs, Go/Rust/Node security, DeepScan, and
  CodeRabbit.
- The normal merge remained blocked only by the formal second-review branch
  rule. With explicit owner authorization and Kimi's independent 0 P0 / 0 P1 /
  0 P2 review recorded, Sol used the scoped admin squash merge. PR #67 merged
  as `c5a3d38`; GitHub closed GH-60.
- GH-29, the roadmap, Bridge, audit, security notes, TODO, and project state
  are synchronized in the closure branch. GH-61 is the next bounded
  consensus-state task; no deployment or production rollout is approved.

## 2026-07-26 21:56 EEST - GH-61 legacy authority migration start

- Reconstructed GH-61 from clean `origin/main` at `8c0c555`, the closed GH-60
  chain, GH-56's coupled-authority boundary, GH-29, and the 72-item rollout
  roadmap (11 complete, 61 open).
- The migration must be a deterministic halt/export/re-import ceremony, not a
  compatible binary replacement. It must authenticate every replacement
  operator independently from all consensus keys and preserve exact validator,
  consensus, domain, escrow, auth-account, revocation, liveness, infraction,
  pending-exit, and supply state.
- Created `feature/GH-61-legacy-authority-migration`; no unrelated user changes
  are present.
- Sol owns architecture, security, integration, external writes, complete
  verification, and closure. Kimi K3 receives one bounded secret-free
  architecture/development block and may not delegate.
- No GH-61 verification has run. Next: define the canonical plan/proof and
  atomic fail-closed rewrite, then add unit, genesis-integration, rollback, and
  real multi-validator migration evidence.

## 2026-07-26 22:09 EEST - GH-61 architecture security review

- Kimi's read-only review mapped the complete rewrite graph and confirmed that
  an offline atomic halt/export/transform/re-import boundary is smaller and
  more honest than claiming nonexistent `x/gov` or `x/upgrade` support.
- Sol rejected the review's proposed legacy-consent and >2/3 validator
  signatures because they were verified with the legacy consensus keys. That
  would reintroduce exactly the consensus-derived authority GH-61 must remove.
- No implementation was accepted or written. The remaining architecture
  question is whether pre-GH-56 state contains any independently controlled
  governance anchor. If not, GH-61 must introduce and authenticate one
  explicitly or keep the reviewed fresh-genesis path as the only supported
  transition.
- Next: run a bounded adversarial authorization review, record the decision,
  then delegate only the approved migration core.

## 2026-07-26 22:20 EEST - GH-61 authorization-anchor decision

- Kimi completed a bounded read-only adversarial audit of every plausible
  pre-GH-56 authorization anchor. No independently controlled signer with
  protocol-granted authority over validator identity exists.
- `x/gov` and `x/upgrade` are absent; module accounts cannot sign; the legacy
  validator operators and bootstrap domain admin are consensus-key-derived;
  domain permission/ZKP keys and arbitrary user accounts carry no chain
  migration authority.
- Sol therefore rejects retroactive “governance approval.” The only honest
  one-time recovery boundary is a reviewed, state-preserving fresh-genesis
  ceremony bound to the exact halt height and source app hash, with
  proof-of-possession from every fresh independent replacement operator.
- A new consensus-independent governance commitment may be introduced in the
  migrated genesis for future decisions, but cannot be presented as authority
  for the initial rewrite.
- No implementation or test run occurred during this review. Next: synchronize
  the finding to GH-61, narrow its acceptance boundary, and implement the
  canonical descriptor/proof verifier plus deterministic transformer and
  recovery drill.

## 2026-07-26 22:38 EEST - GH-61 canonical descriptor and fresh-key proofs

- Added `migration/legacy_authority.go` and its test suite. Descriptor v1 binds
  both chain IDs, halt height, source app hash, transform ID, and the complete
  mapping; each fresh ed25519/secp256k1 operator key signs the same
  domain-separated canonical bytes.
- Verification rejects malformed/mixed-prefix addresses, wrong key derivation
  or proof, descriptor mutation/replay, duplicates, unsorted mappings,
  old/new cross-coupling, malformed consensus-key input, and replacement
  addresses derived from any supplied legacy consensus key.
- Kimi implemented only this bounded package and its tests. Its first test run
  caught a stale-index sorting comparator; Kimi corrected it before handoff.
- Sol reviewed the complete diff, added stable JSON tags and exact account-size
  checks at canonical-encoding time, then independently reran the gate.
- `gofmt -l migration` and `git diff --check` are clean; `go vet ./migration`
  passes; uncached package tests pass in 0.946s; the race run passes in 2.578s.
- Remaining boundary: the future transformer must supply the complete exported
  consensus-key set and reconcile every typed reference. No CLI, transformer,
  integration drill, repository-wide gate, commit, push, or PR exists yet.

## 2026-07-26 22:59 EEST - GH-61 transformer threat-review refinement

- Enforced distinct source/target chain IDs in descriptor verification and
  added the negative regression, preventing legacy account-transaction replay
  on the reviewed fresh-genesis chain.
- Independent field review identified two additional fail-closed requirements:
  nonempty CosmWasm contract state needs a separately reviewed contract-aware
  migration adapter, and auth account numbers must remain globally unique so
  SDK import cannot silently renumber signer identities.
- The generic transformer must repack only mapped `BaseAccount` records with the
  fresh public key and unchanged account number/sequence, rejecting mapped
  module or unknown account types.
- A second Kimi block completed the detailed typed-transform design but was
  stopped before file writes after exceeding its bounded planning window. No
  partial transformer code was accepted.
- Uncached migration tests pass in 1.334s after the chain-ID change and
  `git diff --check` is clean. The transformer, CLI, integration/process drill,
  repo-wide verification, commit, push, PR, and merge remain open.

## 2026-07-27 00:09 EEST - GH-61 truedemocracy transform core

- Added the state-derived transform core and focused regressions. It inventories
  active, historical, revoked, pending-rotation, and pending-removal consensus
  keys directly from the export before verifying any descriptor proof.
- Exact mapping completeness, replacement collision, invalid proof, malformed
  export, atomic failure, input/descriptor immutability, post-transform genesis
  validation, and old-literal removal are covered.
- Spark's first draft was not accepted as written: Sol's gate found duplicate
  test symbols, incorrect secp256k1 address derivation, a short app hash, and an
  invalid fixture. Sol replaced the faulty tests and hardened SDK-prefix and
  domain-admin decoding before rerunning the complete package gate.
- `gofmt -l migration` and `git diff --check` are clean; `go vet ./migration`
  passes; uncached tests pass in 3.726s; race tests pass in 3.617s.
- Auth/Bank/DEX reconciliation, the explicit Wasm gate, outer genesis binding,
  CLI, process drills, repo-wide verification, and GitHub publication remain
  open.

## 2026-07-27 00:24 EEST - GH-61 full app-state reconciliation

- Added `migration/app_state.go` and focused application-genesis regressions.
  The atomic transformer now binds outer chain ID and halt successor height,
  repacks mapped BaseAccounts with fresh keys while preserving account
  identity counters, moves Bank and DEX ownership, and validates each typed
  module before returning output.
- Duplicate auth addresses/account numbers, unsupported mapped account types,
  pre-existing replacement ownership, invalid source/post state, stale legacy
  address literals, and descriptor proof failures all fail without output or
  input mutation.
- Generic migration now refuses nonempty CosmWasm code/contract state. Unknown
  modules remain preserved and are included in the stale-address gate.
- The earlier independent review supplied the critical Auth/Wasm and
  custom-module boundary requirements. Sol owned implementation, complete diff
  review, fixture corrections, and final verification for this slice.
- `gofmt -l migration` and `git diff --check` are clean; `go vet ./migration`
  passes; uncached migration tests pass in 2.177s; race tests pass in 6.398s.
- Remaining: trusted-header app-hash CLI verification, atomic file output,
  operator docs, export/transform/re-import and rollback drills, full repository
  gates, GitHub publication, commit, push, PR, review, and merge.

## 2026-07-27 00:54 EEST - GH-61 canonical offline CLI and re-import

- Added `truerepublicd migration legacy-authority` with strict descriptor
  decoding, explicit trusted source-app-hash comparison, bounded regular-file
  reads, input/output identity checks, no-overwrite private atomic output, and
  content-free success reporting.
- Added the missing App/Comet validator reconciliation: a populated consensus
  set must exactly match active application consensus keys and power; empty
  export validators remain valid for application `InitChain` reconstruction.
  Fresh target genesis rejects an embedded prior app hash.
- The CLI performs the existing exact cross-module ledger validation before
  creating output. The new end-to-end command regression transforms a valid
  legacy-coupled full genesis and reimports it into a fresh application.
- Kimi supplied a detailed bounded file-safety design but produced no write
  diff before Sol stopped the overlong planning run. Sol implemented, reviewed,
  hardened, and verified the complete slice.
- `gofmt -l` and `git diff --check` are clean; `go vet ./migration .` passes;
  uncached package/root tests pass in 3.951s/141.937s; focused race tests pass
  in 13.865s/13.296s; the canonical roundtrip passes again in 12.041s; daemon
  build/help smoke exposes the new command.
- Remaining: operator runbook, real pinned pre-GH-56 four-validator
  halt/export/transform/import and rollback drill, full repository gates,
  independent final security review, GitHub publication, commit, push, PR,
  review, and merge.

## 2026-07-27 01:34 EEST - GH-61 historical migration and rollback drill

- Added a gated four-validator harness that builds pinned pre-GH-56 revision
  `0e51a05b008f395e3f7391358e117f0817d4eb39`, runs a real coupled-authority
  source chain, establishes a stable no-quorum halt/app hash, exports and
  transforms through the current CLI, starts a distinct-ID target chain,
  verifies app-hash/key/power convergence and export/reimport, stops all target
  signers, and resumes the untouched historical source as rollback.
- The first target rehearsal discovered a real initial-height defect:
  genesis-restored consensus keys used runtime H+2 activation and rejected the
  first decided commit when initial height was greater than one. Split genesis
  activation (actual initial height) from runtime activation (H+2) and added a
  focused height 0/1/4 regression without weakening rotation semantics.
- Final live drill passes in 249.13s. The focused activation suite passes in
  3.588s; post-documentation truedemocracy/migration/root checks pass in
  2.537s/5.105s/3.688s; docs consistency, vet, formatting, and diff checks pass.
- Added the operator migration runbook and synchronized the operator index,
  upgrade boundary, rollout roadmap, and limitations. The procedure explicitly
  records that fresh-key proofs prove possession, not retroactive governance,
  and that source/target signers may never run concurrently.
- Remaining: full repository and security matrices, independent final
  adversarial review and remediation, commit/push/PR, final-head GitHub
  workflows/review, merge, and issue/roadmap closure.

## 2026-07-28 20:49 EEST - GH-61 exact export binding and local final gate

- Confirmed GH61-SB-001 in a disposable pre-fix CLI reproduction: a valid
  export could be changed after fresh operators signed the descriptor, and the
  changed governance state was transformed and re-imported because only the
  source header app hash was bound.
- Added `source_genesis_sha256` to descriptor v1 canonical signing bytes and
  require exactly 32 bytes in encoding and verification. The transform hashes
  the exact raw byte slice before JSON parsing and compares it in constant time,
  so direct package callers and the CLI share one fail-closed boundary.
- Updated descriptor, transform, full-application, CLI, and historical-process
  fixtures to sign the exact artifact. Raw-mutating negative fixtures recompute
  and re-sign their commitment so they still exercise the intended deeper
  Auth/Bank/DEX/Wasm/Comet and mapping controls.
- Added the project-local `AGENTS.md` required by the refreshed Bridge workflow
  and synchronized the operator runbook with exact-byte freeze/hash/sign
  ordering. Reformatting the export after signing is now explicitly forbidden.
- Kimi performed an independent secret-free read-only review and made no file
  changes. It found no P0/P1/P2 issue. Sol accepted and fixed one P3 test-
  attribution observation by sorting an added mapping before testing signature
  invalidation; the separate canonical-string sweep boundary remains recorded
  as P3 without expanding this task.
- Focused migration/root regressions pass (`migration` 3.318s, root 6.837s).
  `make verify` passes: repository selector, build, vet, and Race/Coverage for
  root 193.441s/69.4%, migration 18.935s/82.5%, token 5.397s/92.6%, treasury
  9.157s/97.0%, DEX 23.969s/45.3%, and truedemocracy 172.168s/62.2%.
  Documentation consistency, `gofmt -l`, and `git diff --check` pass.
- The exact real historical
  `TestMultiValidatorLegacyAuthorityMigrationRollback` drill passes again in
  174.82s (178.949s package): four-validator halt/export, exact-bound
  transform, distinct target convergence, export/re-import, complete target
  signer shutdown, and untouched source rollback all remain proven.
- GH61-SB-002 remains open but out of scope: a measured 64 MiB padded artifact
  reached about 1.75 GiB maximum resident memory and 33.51s in the focused
  transform path. No push, PR, merge, deployment, or follow-up remediation was
  authorized or performed in this task.

## 2026-07-29 02:06 EEST - GH-61 publication started

- Froze and staged exactly the 26 documented GH-61 files. The staged diff and
  secret-pattern scan were clean; no unrelated user changes were present.
- Created commit `432156fe9090b1f2711461d0a135715b572fd8bb` and pushed
  `feature/GH-61-legacy-authority-migration` to `origin` without force.
- Opened draft PR #69 against exact `main` base
  `8c0c55574389210162db59c2c1103bb2d99e2691`. The description closes GH-61
  only on merge and records the honest fresh-genesis boundary, real local
  evidence, unsupported production claims, and GH61-SB-002 residual risk.
- Initial GitHub checks started; DeepScan is green and CodeRabbit skipped the
  draft as configured. The PR remains draft until required final-head checks
  pass, after which review must be triggered and resolved before merge.

## 2026-07-29 02:09 EEST - PR #69 current-database gRPC finding

- Final-head Security Scan run 30361942572 failed only `go-vuln`: reachable
  GO-2026-6061 affects `google.golang.org/grpc@v1.79.3`, with `v1.82.1`
  reported as fixed. The four other reachable upstream findings remain `N/A`.
- This is a vulnerability-database/dependency gate change, not a GH-61
  transform regression. The bounded remediation is the scanner-named gRPC
  update plus unavoidable module metadata and full vulnerability, Race, and
  network-process verification before another push.

## 2026-07-29 02:29 EEST - GO-2026-6061 remediation locally verified

- Upgraded direct `google.golang.org/grpc` from `v1.79.3` to fixed `v1.82.1`.
  The required xDS, Genproto, Envoy validation, and GCP detector support graph
  followed; `go mod tidy` also classified already directly imported
  `cosmossdk.io/core` correctly. No application source changed.
- Current `govulncheck` removes GO-2026-6061 and shows zero reachable finding
  with an available fix. Four existing reachable `N/A` findings remain
  visible and tracked.
- `go mod verify`, docs consistency, diff check, package selector, build, vet,
  and full Race/Coverage PASS. Root/migration coverage remains 69.4%/82.5%.
- The complete eight-scenario real process matrix passes in 940.823s under the
  new transport graph, including GH-61 migration/rollback, recovery,
  state-sync, backup/restore, binary rollback, cold failover, key rotation,
  and slashing.

## 2026-07-29 03:08 EEST - PR #69 final-head CI green; review recovery

- Exact head `a28f5654390f58feecfd86afeeb4d7f74e204341` passed docs, Go,
  full multi-validator recovery, Docker restart, Go/Rust/Node security, and
  DeepScan checks; DeepScan reported zero new issues.
- CodeRabbit remained pending for about ten hours without a review or inline
  thread. Its explicit review command was acknowledged but did not replace the
  stale run; a draft/ready transition also left the same pending context.
- Recorded the external-review recovery state append-only. This
  documentation-only commit is intended to produce a fresh final head and bot
  event. No implementation/dependency behavior changes and no review bypass,
  merge, deployment, or production action were made.

## 2026-07-29 03:14 EEST - PR #69 review findings accepted for remediation

- CodeRabbit's full review of `a28f565` completed with six unresolved inline
  threads and seven additional review-body findings. The findings cover
  runbook atomicity/halt integrity, genesis consistency, deprecated SDK use,
  test diagnostics/coverage, mapping-sort efficiency, shared key traversal,
  and CI timeout margin.
- Accepted all findings as in-scope review remediation; the repository-wide
  docstring percentage remains informational.
- Assigned the bounded secret-free Go/test/CI block to Kimi K3. Sol retained
  documentation, architecture/security decisions, diff review, complete local
  and GitHub validation, thread resolution, external writes, and merge.

## 2026-07-29 03:33 EEST - Review remediation integration correction

- Kimi completed the bounded Go/test/CI patch; Sol reviewed the full diff.
  Focused tests/vet and `make verify` passed, raising migration coverage to
  85.0%.
- The first full process-matrix attempt failed in the historical GH-61 drill:
  a real halted running-chain export had an empty Comet validator list and four
  active truedemocracy validators. This proved CodeRabbit's proposed rejection
  incompatible with the supported Cosmos SDK export shape.
- Stopped the remaining matrix, restored canonical empty-consensus-export
  acceptance, added a direct regression plus code/runbook rationale, and kept
  strict reconciliation for non-empty consensus lists. A fresh complete matrix
  is required before publication.

## 2026-07-29 03:53 EEST - Corrected full process matrix passes

- Corrected migration tests and docs/diff gates passed.
- All eight real process scenarios passed in 1169.685s. The GH-61 historical
  migration/rollback slice passed in 196.40s; consensus recovery 119.13s; state
  sync 145.57s; backup/restore 90.99s; binary upgrade/rollback 223.52s; cold
  failover 87.03s; key rotation 88.48s; slashing 215.28s.
- The observed total leaves only about 30 seconds under the former 1200-second
  limit, validating the reviewed 1500-second test and 30-minute job budgets.
- The empty-consensus review suggestion is rejected with direct process
  evidence; every other review finding remains implemented pending final-head
  publication and thread resolution.

## 2026-07-29 03:58 EEST - Final review-remediation local gate passes

- Final `make verify` passed package selection, build, vet, and Race/Coverage:
  root 69.4%, migration 84.6%, token 92.6%, treasury 97.0%, DEX 45.3%, and
  truedemocracy 62.2%.
- Together with the corrected 8/8 process matrix and docs checks, the final
  local remediation diff is ready for staged review and publication.

## 2026-07-29 04:18 EEST - GH-61 merged and closed

- PR #69 final head `33895dca627139d541e74236afb63d2b9fb5bff2`
  passed all 11 GitHub checks. The complete process matrix passed in 13m38s;
  Go build/race/coverage, Docker restart, docs, Go/Rust/Node security,
  DeepScan, and CodeRabbit also passed.
- All original and final CodeRabbit threads were answered with evidence and
  resolved; zero unresolved threads remained before merge.
- PR #69 squash-merged to `main` as
  `264ab7c817414e3081149c61d8b5b6fb0ce5e368` and automatically closed
  GH-61.
- Started the documentation-only closure sync from that exact clean main commit
  to align Bridge, Project State, TODO, Road to Rollout, Limitations, Security
  Notes, and GH-29.

## 2026-07-29 04:24 EEST - Public recovery totals reconciled

- Reproduced 819 Go cases on merged main via package-scoped `go test -json`:
  root 62, migration 82, token 12, treasury 36, DEX 116, truedemocracy 511.
- Updated the authoritative public total from the stale GH-60 baseline
  733/699 to 853 total: 819 Go, 26 Rust, and eight maintained-client cases.
- Synchronized status JSON, README, agent guide, Pages source, wiki status,
  module arithmetic, Project State, and final root/migration/governance
  coverage. Separately gated process harnesses remain excluded from the count.

## 2026-07-29 04:35 EEST - PR #70 independent closure review remediated

- Kimi K3 performed a bounded read-only review and made no repository changes.
  It independently verified the PR #69 merge and issue state, all 11 final-head
  checks, zero unresolved review threads, exact 853/819 arithmetic and
  per-module Go counts, coverage evidence, documentation-only scope, and
  retained non-production boundaries.
- Review found no P0/P1, code, security, or production-claim issue. Corrected
  its sole material P2 (stale 733/699 roadmap baseline) and minor P3s (roadmap
  date and GH-61 omission from the wiki harness list).
- CodeRabbit did not rerun after its successful draft-skip result despite the
  ready-for-review/manual request; the independent Kimi review is retained as
  fallback evidence. Final merge still requires Sol diff review, green
  final-head checks, and zero unresolved threads.

## 2026-07-29 12:24 EEST - GH-71 locally verified and independently reviewed

- Implemented deterministic seed, sentry, validator, RPC/API, and private node
  policies with canonical peer validation, validator-initiated sentry
  isolation, listener/CORS/unsafe/metrics controls, fail-closed startup, safe
  Docker/nginx defaults, and an operator runbook.
- Kimi K3 implemented the bounded policy core and later completed an
  independent read-only review. Sol reviewed every diff, corrected topology,
  path-safety, P2P, metrics, Docker, documentation, and test gaps, and retained
  all external/GitHub authority.
- `make verify` passes on the final tree. Coverage is root 69.5%, migration
  84.6%, network policy 95.5%, token 92.6%, treasury 97.0%, DEX 45.3%, and
  governance 62.2%.
- Reproduced 952 Go cases: root 69, migration 82, network policy 126, token 12,
  treasury 36, DEX 116, governance 511. Together with 26 Rust and eight client
  cases, the public branch total is 986.
- Seven process scenarios passed in the combined run; concurrent load exhausted
  its 1500-second global limit during slashing. The isolated slashing rerun
  passed in 303.98 seconds. Exact final-head combined evidence remains a GitHub
  merge gate.
- Local Docker is unavailable. CI now runs the complete Compose stack and
  requires loopback proxy RPC, wallet routing, healthy node metrics, and
  Grafana health before merge.
- Structured audit: `GH71_AUDIT.md`. No P0/P1 or scope blocker remains.

## 2026-07-29 12:52 EEST - GH-71 merged and closed

- PR #72 exact final head `83ffaad86274efdf3d2bb54dfff6ce8a625d5dee`
  passed Go build/race/coverage (6m03s), the complete recovery matrix (14m29s),
  Docker/Compose restart and routing smoke (7m04s), docs, Go/Rust/Node security
  scans, and DeepScan with zero new findings.
- CodeRabbit remained in `Review in progress` for about ten hours without
  producing a review, finding, or thread. The stale external-bot condition was
  recorded on the PR; zero unresolved review threads existed. Independent Kimi
  review had already found no P0/P1, and every lower-severity observation was
  remediated before the exact final-head CI run.
- Under the repository owner's standing merge authorization, Sol used an admin
  squash merge only to bypass the stale bot status. PR #72 merged to `main` as
  `9c369ac0d589f749e055af33a03f5f4981020101`; GH-71 closed automatically.
- No server, firewall, DNS, production topology, deployment, key, credential,
  mainnet, or public-network action was performed. GH-29 remains the rollout
  tracker.

## 2026-07-29 17:45 EEST - GH-74 locally verified

- Added dependency-free `truerepublicd healthcheck live|ready` commands.
  Liveness checks only local RPC responsiveness; readiness additionally
  requires a positive committed height and `catching_up=false`.
- Hardened probes to literal-loopback plain HTTP, positive timeout at most ten
  seconds, 64 KiB responses, no userinfo/query/fragment/non-root path, no
  redirects or environment proxy, strict JSON-RPC 2.0, and stable
  operator-safe errors.
- Docker now uses liveness for its restart signal. GitHub container and
  Compose gates require liveness, readiness, and post-restart height progress.
  Operator documentation includes distinct Kubernetes exec probes.
- Kimi K3 implemented the bounded probe package/CLI/focused tests. Sol reviewed
  the complete write diff and corrected redirect/proxy handling, JSON-RPC
  version validation, an environment-sensitive port test, and race-safety.
- Local evidence: focused race tests PASS twice; focused root/config tests
  PASS; `go vet ./healthcheck` and `go vet .` PASS; `make verify` PASS.
  Package-scoped JSON evidence records 1,009 Go cases: root 71, healthcheck 55,
  migration 82, network policy 126, token 12, treasury 36, DEX 116, and
  governance 511. Public total: 1,043 with 26 Rust and eight maintained-client
  cases.
- Coverage: root 69.5%, healthcheck 97.2%, migration 84.6%, network policy
  95.5%, token 92.6%, treasury 97.0%, DEX 45.3%, and governance 62.2%.
- Docker is unavailable locally. Exact final-head GitHub image/Compose runtime
  evidence, independent review, and merge remain hard gates.
- Structured audit: `GH74_AUDIT.md`; current result 0 FAIL / 1 WARN / 12 PASS.
- Final independent Kimi review found no P0/P1/P2. Sol remediated its
  actionable P3 by validating the separately read block height before
  publishing or comparing the restart baseline. The duplicated timeout wording
  and standard released-port test idiom remain non-blocking P3 observations.
- No production, server, firewall, DNS, deployment, real topology, secret,
  credential, key, mainnet, or public-network action was performed.
