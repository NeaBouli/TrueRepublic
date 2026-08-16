# Recovery Queue

## P0 - security and reproducibility

- [x] GH-172: implement and locally verify bounded concurrent shared-state,
  exact signed-byte replay, clean same-home restart, app-hash, ledger, and
  export/reimport evidence without changing production behavior.
- [x] GH-172: complete independent review, protected exact-head CI, merge,
  GH-29/public-status synchronization, final-main and Pages verification.

- [x] GH-169: implement and locally verify a versioned cross-system threat
  register plus fail-closed repository contract across all maintained domains.
- [x] GH-169: publish through protected exact-head CI, merge, synchronize GH-29
  and public status, then verify final `main` and Pages before Done.
- [ ] GH-4: keep the recovery epic and acceptance criteria current.
- [ ] GH-29: complete the seven-phase production-readiness roadmap and attach
  evidence for every rollout exit gate before any public-network launch.
- [x] GH-148: locally implement immutable Action pins, exact security tool and
  toolchain versions, fail-closed Go vulnerability/static and maintained-tree
  secret gates, lock enforcement, bounded weekly dependency updates, and a
  repository-owned negative-fixture contract.
- [x] GH-148: complete independent final review, protected exact-head CI,
  merge, GH-29/public-status synchronization, and final-main/Pages verification.
- [x] GH-153: classify and close the three failing first-run Dependabot PRs;
  implement structurally tested minor/patch-only grouping with ordinary major
  updates excluded across Actions, Go, Cargo, and maintained npm.
- [x] GH-153: pass independent final review and complete local/protected CI,
  merge, issue/Bridge synchronization, and final-main verification.
- [x] GH-154: delete only exact merged-PR remote branches and clean exact-merged
  local worktrees; preserve and document every unique, dirty, archival, active,
  or ambiguous ref/worktree.
- [x] GH-154: pass independent final review with no unresolved P0-P3 finding.
- [x] GH-154: publish the audit through protected docs-only CI, merge, close the
  issue, and verify final-main Docs/Pages state.
- [x] GH-161: classify all three bounded generated updates; refresh and merge
  only the fully green Cargo lockfile patch.
- [x] GH-161: publish the typed policy refinement and bounded npm/Go replacement
  updates through independent review, protected CI, merge, final-main, Pages,
  issue, tracker, and Bridge closure.
- [x] GH-145: implement deterministic property, fuzz, invariant, replay,
  malformed-genesis, and focused race evidence for critical Go paths without
  changing production behavior.
- [x] GH-145: pass complete local and protected exact-head gates, merge,
  synchronize GH-29/public status/Bridge, and verify final `main` plus Pages.
- [x] GH-139: implement and locally verify bounded root/application, DEX, and
  governance critical-path coverage regression gates with meaningful rollback,
  authority, custody, slippage, escrow, reward, and withdrawal tests.
- [x] GH-139: publish through a protected exact-head PR, merge only after all
  relevant workflows pass, synchronize GH-29/public status/Bridge, and verify
  final `main` plus Pages before marking Done.
- [x] GH-141: copy the maintained-client build scripts into its Docker builder,
  pin copy-before-build plus bundle-budget invocation in a repository contract,
  and verify the direct package build locally.
- [x] GH-141: publish, pass protected Docker image/restart and exact-head
  evidence, merge as `11ef2f6`, close the issue, and refresh GH-140 on the
  repaired base.
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
- [x] GH-112: reproduce and explicitly retire the deprecated legacy web client;
  replace its Compose/nginx route with maintained `client-web` and add a
  blocking absence contract; published through merged PR #117.
- [x] GH-115: implement one exact maintained-client custom transaction
  registry, centralized fail-closed simulate/sign/deliver behavior, canonical
  address handling, chain-side DEX descriptor/signer compatibility, wire
  vectors, and ephemeral local client-chain evidence.
- [x] GH-115: complete independent final review and remediate its test-count
  and interrupted-setup cleanup observations with fresh evidence.
- [x] GH-115: complete protected exact-head CI, merge, issue closure, GH-29
  synchronization, Pages readback, and final-main verification before marking
  the ticket Done.
- [x] GH-116: inventory consumers, remove the consumerless legacy custom ABCI
  query shim, port DEX query coverage to the supported gRPC QueryServer, and
  document the exact CLI/protobuf gRPC boundary without advertising an
  unregistered HTTP gateway.
- [x] GH-116: complete protected exact-head CI, review, merge, issue closure,
  Bridge synchronization, and final-main verification.
- [x] GH-121: replace the maintained browser client's unregistered custom-
  module REST aliases with one explicitly supported query transport, reject
  unavailable data instead of silently returning empty state, and prove the
  boundary against a disposable chain.
- [x] GH-121: complete independent final review, protected exact-head CI,
  merge, issue closure, Bridge synchronization, and final-main verification.
- [x] GH-128: locally split all 19 maintained-client page routes, break the
  eager CosmJS entry dependency, add accessible loading/error behavior, and
  enforce deterministic raw/gzip entry, route, chunk, and total budgets.
- [x] GH-128: publish the reviewed branch, pass protected exact-head Client,
  Docs, and security checks, merge, close, synchronize GH-29, and verify final
  main plus Pages before marking the Bridge ticket Done.
- [x] GH-122: merge the isolated patched `js-yaml` lock resolution through PR
  #123 after final exact-head security and maintained-client checks pass;
  preserve the fail-closed live High audit policy.
- [x] GH-71: locally implement and verify deterministic role-based network
  policy, safe listener/container defaults, startup validation, regression
  coverage, and operator guidance without touching production infrastructure.
- [x] GH-71: publish the reviewed branch, obtain green final-head GitHub
  build/race/coverage, process, Docker, docs, static/security, and independent
  review evidence; merge through PR #72, close, and synchronize GH-29/Bridge.
- [x] GH-74: locally implement and verify separate bounded node liveness and
  synchronization-aware readiness commands, Docker/CI integration, repository
  policy coverage, and operator guidance.
- [x] GH-74: publish the reviewed branch, obtain green final-head GitHub
  build/race/coverage, process, Docker/Compose, docs, static/security, and
  independent review evidence; merge through PR #75, close, and synchronize
  GH-29/Bridge.
- [x] GH-77: implement and locally verify structured JSON daemon logging with
  central secret/private-transaction redaction, focused unit and real-process
  regressions, safe operator defaults, and documentation.
- [x] GH-77: publish the reviewed branch, obtain green exact-head GitHub
  build/race/coverage, recovery, Docker/Compose, docs/static/security, and
  independent review evidence; merge, close, and synchronize GH-29/Bridge.
- [x] GH-80: implement and locally verify a two-source metrics baseline for
  CometBFT consensus/peer/block signals plus Cosmos SDK/runtime and bounded
  TrueRepublic application/invariant-success signals.
- [x] GH-80: publish the reviewed branch, obtain green exact-head GitHub
  build/race/coverage, recovery, Docker/Compose metrics, docs/static/security,
  and independent review evidence; merge, close, and synchronize GH-29/Bridge.
- [x] GH-85: implement and locally verify a file-provisioned CometBFT plus
  application dashboard, actionable Prometheus rules, recovery/testnet
  objectives, role-based escalation ownership, and deterministic rule/query
  evidence without external paging or production deployment.
- [x] GH-85: publish the reviewed branch, obtain green exact-head GitHub
  build/race/coverage, recovery, Docker/Compose dashboard/rule/query,
  docs/static/security, and independent review evidence; merge, close, and
  synchronize GH-29, public status, and Bridge.
- [x] GH-89: define and locally verify a versioned, secret-free multi-node
  topology contract that composes GH-71 role policies into explicit
  seed/sentry/validator/RPC relationships, validator isolation, deny-by-default
  flows, and bounded public-ingress abuse protection without deploying
  production infrastructure.
- [x] GH-89: publish the reviewed branch, obtain green exact-head GitHub
  build/race/coverage, recovery, Docker/Compose, docs/static/security, and
  independent review evidence; merge and synchronize GH-29, public status,
  audit, and Bridge while leaving live production deployment open.
- [x] GH-97: implement, independently review, and verify a strict synthetic
  four-validator capacity qualification covering sustained multi-block
  transaction waves,
  disk/log/RSS growth, finite retention, metrics, restart, ledger invariants,
  and checked projection without making production-sizing claims.
- [x] GH-97: publish through a protected exact-head PR, close only after all
  required checks pass, then synchronize GH-29, public status, audit, and
  Bridge while leaving production topology and live deployment open.
- [x] GH-102: merge the maintained-client React Router migration through PR
  #107, preserve all 21 routes, remove both prior moderate advisories, and bind
  the temporary RSC-only exception to the exact package, architecture, owner,
  deadline, tests, and protected CI.
- [x] GH-102: reproduce and classify the legacy mobile dependency inventory,
  then migrate or explicitly retire that client without weakening wallet,
  signing, chain compatibility, or rollout boundaries.
- [x] GH-101: implement and independently review a strict secret-free private
  topology deployment-evidence manifest/verifier bound to the exact GH-89
  contract, with synthetic fixtures, repository/CLI/no-reflection tests, CI,
  and operator procedure; do not inspect or deploy real infrastructure.
- [x] GH-101: publish through a protected exact-head PR, close only after all
  relevant gates pass, and synchronize Bridge/GH-29 while leaving the live
  production-topology checkbox open for separately authorized evidence.

## P1 - consensus and wallet audit

- [x] GH-175: locally prove two-chain Tendermint clients, connection and
  unordered ICS-20 channel, native PNYX escrow/voucher/ACK, timeout refund,
  replay safety, and pending-ack persistent recovery; register the concrete
  Tendermint light-client codec types.
- [x] GH-175: publish the reviewed branch, pass protected exact-head CI, merge,
  synchronize GH-29/Pages/Bridge, and verify final `main` before Done.
- [ ] Phase 3 follow-up: qualify an external relayer/counterparty. Channel
  close/replacement and timeout-on-close are covered by GH-178; GH-181 covers
  only the supported compatible in-place binary restart. Governed
  consensus-breaking migrations and IBC client upgrades remain separate.

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
- [x] GH-209: bind both anonymous rating paths to one canonical, proof- or
  signature-covered reward recipient; preserve the recipient-independent
  nullifier, frozen circuit/setup artifacts, atomic escrow accounting and the
  hard-disabled maintained-client submission boundary.
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
- [x] GH-59: locally wire deterministic ABCI++ misbehavior and decided-last-
  commit ingestion into exact equivocation/downtime burns, replay-safe key
  history, rolling liveness, and evidence-window validator-exit custody.
- [x] GH-59: prove real four-process downtime, restart/catch-up, duplicate-vote
  evidence, power-zero transitions, app-hash convergence, and ledger-valid
  export.
- [x] GH-59: publish the audited branch, pass all 11 final-head GitHub checks,
  merge as `934a042`, close the issue, and synchronize GH-29 plus the Bridge.
- [x] GH-60: implement explicit active/inactive validator genesis state and
  preserve custody, all domains, jail, missed-block, power, signing,
  revocation/rotation/infraction, and pending-exit semantics exactly.
- [x] GH-60: add malformed-resurrection and revoked-key-reuse regressions plus
  real process-level export/import evidence; complete local verification and
  independent review.
- [x] GH-60: publish the verified implementation and close only after
  final-head GitHub CI and review pass.
- [x] GH-61: locally implement canonical fresh-operator proofs, exact
  truedemocracy/Auth/Bank/DEX/Comet reconciliation, empty-Wasm boundary,
  trusted-hash atomic CLI, and a real pinned pre-GH-56 four-validator
  migration/export/reimport/rollback drill with an operator runbook.
- [x] GH-61: complete full local/security gates and independent final review;
  remediate the in-scope review finding and record residual risks.
- [x] GH-61: publish the branch, remediate review/CI findings, pass all 11
  final-head checks with zero unresolved threads, merge through PR #69 as
  `264ab7c`, close the issue, and synchronize GH-29/Bridge.
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
## Completed - GH-178 IBC channel recovery

- [x] Prove counterparty channel closure plus `TimeoutOnClose` refunds one
  unreceived ICS-20 packet exactly once and removes its commitment.
- [x] Prove closed-channel state survives database recovery, replacement uses a
  distinct channel, and transfer/ACK succeeds without reopening the old path.
- [x] Add the bounded local/CI gate and synchronize repository docs and status.
- [x] Synchronize GH-29 and both Bridges as each remaining gate changes.
- [x] Record protected-PR evidence, final-main checks, and live Pages readback.

## Completed - GH-181 compatible IBC binary-restart recovery

- [x] Preserve real IBC client, connection, channel, sequence, escrow, packet
  commitment, receipt, and acknowledgement state across a separately compiled
  compatible test-binary restart without exporting or reinitializing genesis.
- [x] Complete a pre-restart pending acknowledgement after reopen, then prove a
  fresh transfer/ACK and timeout/refund exactly once on the same channel.
- [x] Prove a fail-before-open candidate leaves persistent state unchanged and
  that the baseline binary can resume safely.
- [x] Run focused and full local/security gates plus independent final review;
  publish only the bounded compatible-restart claim.
- [x] Synchronize GH-29, public status, both Bridges, protected PR/CI, final
  `main`, and live Pages evidence before marking Done.

## Completed - GH-184 governed application upgrade recovery

- [x] Wire real `x/upgrade` with non-signing module authority and remove the
  IBC upgrade stub.
- [x] Add immutable-electorate two-thirds schedule/cancel voting through the
  reserved genesis governance Domain, with conflicting/duplicate/late votes
  rejected and scheduler failures atomic.
- [x] Prove four-validator old-binary halt, partial cached-write rollback,
  fixed v0.4.1 recovery, common hashes, and exact-once restart.
- [x] Synchronize candidate runbook, limitations, threat register, rollout
  roadmap, public status, and Bridge evidence.
- [x] Complete every full local/security gate and independent final review.
- [x] Publish protected PR, merge green exact head, verify final `main`, update
  GH-29 and live Pages, and close both Bridges.

## Completed - GH-187 protocol compatibility boundary

- [x] Replace success-like IBC/CosmWasm staking and distribution stubs with
  stable, explicit fail-closed compatibility adapters.
- [x] Prove `x/staking` and `x/distribution` stay absent from module, store,
  genesis, gRPC, message-router, and CLI surfaces.
- [x] Preserve ICS-20 and governed-upgrade recovery through both real process
  gates; publish the exact supported Cosmos/IBC boundary in candidate docs.
- [x] Complete all repository, security, deterministic-build, client, Rust,
  and independent final-review gates.
- [x] Publish through protected PR, pass exact-head CI/review, merge, verify
  final `main`, update GH-29/live Pages/both Bridges, and close GH-187.

## Completed - GH-190 maintained-client IBC transfer UX

- [x] Register canonical ICS-20 MsgTransfer without weakening the fail-closed
  transaction registry; pin exact wire encoding.
- [x] Add strict open-channel, amount/denom/fee, recipient and bounded timeout
  validation; repair native Send's silent invalid-to-zero amount path.
- [x] Persist wallet/chain-scoped transfer records and render evidence-only
  committed, pending-relay, acknowledgement, timeout/refund and unknown states.
- [x] Add accessible lazy transfer/status routes and comprehensive focused,
  registry, store, component, integration-rejection and bundle regressions.
- [x] Publish protected PR, pass exact-head checks/review, merge, synchronize
  GH-29 and both Bridges, verify final `main` and live Pages, then close GH-190.

## In progress - GH-193 maintained-client wallet and signing safety

- [x] Validate 12/24-word English BIP-39 imports with bounded errors and enforce
  service-layer wallet-name/password bounds.
- [x] Version AES-GCM wallet payloads, raise PBKDF2-HMAC-SHA-256 to 600,000
  iterations, and transparently re-encrypt authenticated legacy payloads.
- [x] Reject stored-address/derived-account, signer prefix/length, recipient,
  and RPC chain-ID mismatches before signing or broadcast.
- [x] Invalidate session-bound signers and stale unlocks after lock, switch,
  active-wallet deletion, or reload; keep session secrets out of persistence.
- [x] Pass focused, full client, build/budget and disposable-chain gates; publish
  exact limitations and candidate arithmetic.
- [x] Complete full repository/security gates and independent review.
- [x] Publish protected PR #194, pass exact-head checks with zero review
  threads, merge as `10a9a83`, synchronize GH-29/live Pages/Bridges, and close
  GH-193.

## Done - GH-198 test-only ZKP compatibility foundation

- [x] Pin synthetic single-party CS/PK/VK/golden-proof artifacts with exact
  checksums, sizes, circuit ID, public-input order and toxic-waste boundary.
- [x] Replay the real proof through verifier and keeper; reject corruption,
  wrong root/chain/rating, manifest drift and trailing data.
- [x] Match maintained-client BN254 MiMC, field and vote-context encodings to the
  shared Go vector while keeping `isSubmittable` hard false.
- [x] Pass full local repository/client/security/docs gates and Kimi review.
- [x] Reconcile PR #197 through merged commit `fc5b418` and rebase the verified
  candidate without changing the approved GH-198 scope.
- [x] Complete PR #199 review, pass exact-head checks, merge, verify final
  main/Pages/Bridges and close GH-198.

## Done - GH-200 iteration-bounded generative-quality harness

- [x] Diagnose the final-main `quality-depth` failure as hosted-runner
  wall-clock expiry after successful fuzz executions, with no invariant failure.
- [x] Replace both 10-second fuzz deadlines with a fail-closed, configurable
  10,000-iteration default capped at 60,000; keep targets and race/property
  stages unchanged.
- [x] Pass local full quality, Go race, vet, docs and diff gates; obtain focused
  independent review with no remaining P0/P1/P2/P3.
- [x] Merge PR #201 as `b0e963f`, close GH-200, and verify final-main workflows
  and live Pages without changing rollout or production-readiness claims.

## Done - GH-203 ZKP circuit-identity freeze

- [x] Add the strict versioned machine-readable circuit/encoding specification
  with explicit test-only/non-production classification.
- [x] Add fail-closed Go cross-checks against Go constants/behavior, the GH-198
  fixture manifest/golden vector, the active go.mod gnark toolchain, and the
  maintained-client zkpEncoding contract with hard-false `isSubmittable`.
- [x] Prove deterministic in-memory constraint-system byte/hash parity against
  pinned `membership_v2.cs` without regenerating or comparing PK/VK/proof
  bytes.
- [x] Complete Sol line-by-line review, full local/security gates, candidate
  arithmetic recount, and independent re-review.
- [x] Publish a protected PR, merge a green exact head, verify final main and
  Pages, synchronize GH-29 and both Bridges, and close GH-203.

## In progress - GH-206 test-only maintained-client Groth16 prover

- [x] Select and document the smallest browser-compatible proving architecture
  that consumes only the exact GH-198/GH-203 synthetic fixtures.
- [x] Implement a strict versioned client/prover boundary and real proof engine
  without enabling anonymous submission or changing consensus.
- [x] Prove exact browser/client-to-Go verification compatibility and reject
  malformed inputs, drift, invalid witnesses, tampered output, and unavailable
  runtime states fail closed.
- [x] Complete Sol line-by-line/security review and full local
  repository/client/security/docs gates.
- [x] Complete the protected PR, merge, final-main, Pages, tracker, and
  both-Bridge synchronization before Done.
  Kimi deep review was attempted twice but provider quota returned HTTP 403;
  an independent read-only review was completed and all findings remediated.

## Completed - GH-212 repository-only release evidence

- [x] Define a bounded versioned manifest binding the exact source commit,
  deterministic build contract, per-platform checksums/metadata, normalized
  Go/client SBOMs, unsigned provenance, tool pins and supported platforms.
- [x] Add a network-free offline verifier with strict decoding, size/depth/path
  bounds, complete digest binding and adversarial failure fixtures.
- [x] Generate each maintained-source Go/client SBOM twice and require
  byte-identical output after normalization (which removes only
  `serialNumber` and `metadata.timestamp`); pin the exact reviewed SBOM tool
  versions, runtime toolchains and digest-pinned container base images
  consistently across `configs/release/tool-platform.json`,
  `configs/security/gates.json`, the Go verifier literals, the generator and
  the workflow, failing closed on any contract schema, tool, platform,
  base-image or Dockerfile FROM mismatch.
- [x] Extend protected CI to assemble and verify an ephemeral evidence bundle
  without publishing binaries/images, tags, releases, signatures or attestations.
- [x] Publish the exact supported-platform and verification boundary; complete
  full local/security gates, Kimi/Sol review and the Claude helper check.
- [x] Publish protected PR, merge a green exact head, update GH-29/Bridges and
  verify final main/Pages while production remains false.

## Local candidate - GH-212 repository-only release evidence

- [x] Define and implement the exact source/build/tool/two-target manifest,
  normalized SBOM and explicitly unsigned provenance bindings.
- [x] Add strict network-free offline verification and adversarial JSON,
  path/symlink, ordering, digest, source, metadata and status failures.
- [x] Pin reviewed tools, supported native runners and maintained multi-arch
  container manifests; cross-bind the Dockerfiles to the release contract.
- [x] Generate both real maintained-source SBOMs twice in CI and exercise a
  synthetic two-target bundle without uploading binaries or release evidence.
- [x] Complete Sol hardening and Kimi independent deep review; Claude remained
  unavailable due expired OAuth and made no changes.
- [ ] Complete all local build/verify/client/Rust/security gates, protected PR,
  exact-head review/CI, merge, final-main/Pages and Bridge/GH-29 closeout.

## Local PASS - GH-212 repository-only release evidence

- [x] Complete two fresh full Go build/verify passes after final pin hardening.
- [x] Pass maintained-client lint, 319 tests, production budget and high audit.
- [x] Pass 26 Rust tests, fmt, Clippy, build and cargo audit.
- [x] Pass deterministic/release/security contracts, staticcheck, gitleaks,
  govuln exact-policy fixtures, real repeated SBOM parity and docs/diff checks.
- [x] Correct moving Node 22 base to official exact Node 22.22.2/npm 10.9.7
  multi-arch image after final registry-config inspection.
- [ ] Commit/push, protected PR, review threads, exact-head CI, merge,
  final-main/Pages, GH-29/GH-212 and both-Bridge closeout.

## Completed - GH-215 sovereign decentralized Alpha architecture

- [x] Audit current maintained-client, wallet, governance, query and signing
  boundaries while preserving `client-web` as Beta.
- [x] Evaluate upstream Telegram open-source components/protocol boundaries,
  decentralization mismatch and licenses from primary sources.
- [x] Define the Alpha component model, trust boundaries, identity/device/key
  model, PNYX wallet, decentralized messaging/storage/sync, moderation,
  governance mappings and no-required-website distribution path.
- [x] Define threat/privacy model, phased Beta-to-Alpha migration, test plan,
  decision gates and explicitly deferred web-interface choice.
- [x] Complete Kimi deep contribution and independent review.
- [x] Complete Sol security/integration review and documentation consistency.
- [x] Attempt the small Claude helper check; local OAuth was expired and Claude
  changed no file.
- [x] Complete protected PR review/checks, merge and exact-main closeout.
- [x] Keep rollout at 34/59 and production false; architecture documentation
  earns no existing rollout checkbox.

## In progress - GH-218 GitHub repository hygiene

- [x] Correct misleading historical release state without moving historical
  tags or claiming a new release.
- [x] Close completed recovery trackers and keep GH-29 as the active rollout
  epic with current evidence.
- [x] Preserve archived history, remove superseded active branches and enable
  automatic deletion of future merged branches.
- [ ] Require only universal protected status contexts and prove them on the
  exact GH-218 PR head.
- [x] Open an explicit license-decision issue without choosing a license.
- [x] Reconcile the tracker arithmetic conservatively: GH-29 has exactly 33
  checked of 59 items (33/51 phase work), so correct the previously published
  off-by-one without inventing completion evidence.
- [x] Complete Kimi review and final audit.
- [ ] Complete Sol verification, protected merge, exact-main/Pages and both-
  Bridge closeout while rollout stays 33/59 and production false.
