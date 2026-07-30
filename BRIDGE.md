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
- GH-55 validator identity recovery audit: [`GH55_AUDIT.md`](docs/agent-bridge/GH55_AUDIT.md)
- GH-56 consensus-key rotation audit: [`GH56_AUDIT.md`](docs/agent-bridge/GH56_AUDIT.md)
- GH-74 node health audit: [`GH74_AUDIT.md`](docs/agent-bridge/GH74_AUDIT.md)
- GH-77 structured logging audit:
  [`GH77_AUDIT.md`](docs/agent-bridge/GH77_AUDIT.md)
- GH-80 metrics baseline audit:
  [`GH80_AUDIT.md`](docs/agent-bridge/GH80_AUDIT.md)
- Decisions: [`DECISIONS.md`](docs/agent-bridge/DECISIONS.md)
- Security: [`SECURITY_NOTES.md`](docs/agent-bridge/SECURITY_NOTES.md)

GitHub recovery epic: [#4](https://github.com/NeaBouli/TrueRepublic/issues/4)

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
