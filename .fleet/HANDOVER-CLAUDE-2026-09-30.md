# Handover — Claude Code standing in for Codex (2026-09-30)

Audience: Codex on return. Thread: `codex://threads/019f5012-64a0-7233-ad73-3b2b36a41ffa`.
Claim: `~/agent-fleet/standin/claims/019f5012-64a0-7233-ad73-3b2b36a41ffa.md`.
Goal: continue exactly from the hand-back point without re-research.

## Starting point (last Codex state, 2026-09-27T20:35Z)
- Integration branch `docs/GH-327-architecture-execution-map` at `7266f7f`
  (~/Documents/Codex/TrueRepublic-GH327), 9 commits ahead of `origin/main`
  `1283a4452d36784ea962f7fc8ba13f9ad45472cd`, local only, no PR.
- GH-327 (map + execution plan) and GH-328 (landing visualization, 69/69 contract
  audit, Kimi review OK) integrated locally. Push/PR/merge wait for the hosted-CI
  reset on 2026-10-01; post-deploy readback of the two Pages links
  (`architecture/MAP.md`, `architecture/EXECUTION_PLAN.md`, Jekyll, no `.nojekyll`)
  is a mandatory gate.
- Next product block per `docs/architecture/EXECUTION_PLAN.md` §5 M0: complete and
  locally gate the preserved GH-306 candidate.

## Mandate
- Gio: "ok leg los" after confirming the codex-standin scope (2026-09-30).
- Stays with Codex: every push/PR/merge, the GH-328 publication, release gate,
  the integration of GH-306, and the security review of GH-306.

## Since then (Claude Code)
| UTC | Ticket | What | Result / evidence |
| --- | --- | --- | --- |
| 2026-09-30 | GH-306 | Exported the uncommitted Codex candidate (`~/Documents/Codex/TrueRepublic-GH306`, untouched) as a verbatim snapshot onto `origin/main` 1283a445 | `agent/claude/GH306-reconcile` `ee7a004`; patch sha256 `d5a78cd86c40e8f349dfd34105f34d8e3f9a76e04065d3971115fe5394e11cca` + 2 untracked files (`configs/security/zkp-protocol-freeze-cv3.json`, `x/truedemocracy/gh306_validator_exit_test.go`) |
| 2026-09-30 | GH-306 | Verified snapshot: build, vet, full `go test ./...`, race on `.` and `x/truedemocracy`, `scripts/check-consistency.sh`, `git diff --check` | all pass |
| 2026-09-30 | GH-306 | Finding F1 fixed (below) + regression test + count bump | `192dccf`; Go pass events 2148 (truedemocracy 707) via the jq `pass` event method; full suite + race + consistency pass |
| 2026-09-30 | GH-306 | Plan M0 gate `make verify` (race + cover, all packages) on `192dccf` | exit 0, no FAIL; truedemocracy coverage 65.3% |
| 2026-09-30 | GH-324/325 | Shared harness port fix on exact main `1283a445`: `agent/claude/GH324-325-port-harness` `b5c7a1a` (worktree `~/Desktop/repos/TrueRepublic-wt/claude-GH324-325`) | new `harness_port_allocator_test.go` + fail-fast in `multi_validator_harness_test.go`; old `freeTCPPort` removed from `server_lifecycle_test.go`; 4 new tests pass under `-race`; `make governed-upgrade` 3/3 pass (379s/544s/268s); `TestMultiValidatorLegacyAuthorityMigrationRollback` 3/3 pass (424s/199s/186s); full CI multi-validator gate, capacity, concurrency-replay **still running at hand-over time** |

## Current state
- Worktrees (own): `~/Desktop/repos/TrueRepublic-wt/claude-GH306` (`agent/claude/GH306-reconcile`,
  base `1283a44`), `~/Desktop/repos/TrueRepublic-wt/claude-standin-20260930`
  (`agent/claude/truerepublic-standin-20260930`, base `7266f7f`, coordination only).
- `~/Desktop/repos/TrueRepublic-wt/claude-GH324-325` (`agent/claude/GH324-325-port-harness`, base `1283a44`) — #324/#325 harness fix.
- Codex worktrees and the paused GH-300 main checkout were only read, never changed.
- Live changes: none. Pushes: none.
- Trial merge `agent/claude/GH306-reconcile` × `7266f7f`: only conflict is the append-only
  `docs/agent-bridge/ACTION_LOG.md`; `docs/index.html` merges cleanly.

## Findings in existing code
- **F1 (fixed, `192dccf`)** — `x/truedemocracy/genesis_validation.go` pending-removal block:
  with GH-306 an excluded validator (custody retained, `Domains` empty, e.g. the GH-60
  "excluded claim" imported from genesis) can now exit, but its budget-exempt hold failed
  `ValidateGenesisState` ("must reference exactly one accounting domain") → export/import
  broken. Budget-exempt holds use no domain accounting (`slashing.go`
  `reducePendingTransferAccounting` returns early), so they now only require listed
  domains to exist; legacy holds keep exactly-one + coverage. Test:
  `TestDomainlessBudgetExemptExitRoundTripsGenesis` (fails without the fix).
- **F2 (open, decision for Codex)** — `upgrade_handlers.go`: the only release handler is
  plan `v0.4.1`, which runs `RunMigrations` from version 1 and now lands on version 3
  (1→2 no-op, 2→3 stone baseline). A chain already at version 2 has no plan name that
  executes the new 2→3 migration. Docs say "registered governed 2→3 migration or fresh
  genesis". No live chain exists (rollout 36/59, production false), so this is a
  release-naming decision, not a bug today — decide whether GH-306 needs a new plan
  (e.g. `v0.4.2`) or whether the next release starts from fresh genesis.
- **Note** — the 2→3 baseline marks only stones present at the upgrade boundary. A member
  whose stone was cleaned up before the upgrade can earn the first-placement reward once
  more afterwards (bounded: at most once per member and scope). Acceptable in my view;
  record explicitly in the TRR-04 closure.
- Earlier map triage candidates from GH-327A (proposal-submit domain membership check,
  `CoinBurnRequired` probably unsatisfiable) remain untouched and still need their
  cross-review.

- **GH-324/325 design** (`harness_port_allocator_test.go`): root cause is TOCTOU on a
  kernel-ephemeral port (p2p dials take ephemeral source ports). Ports now come from
  20000–29999, skipping the host's configured Linux ephemeral range; process registry +
  per-port `O_EXCL` lock files in `$TMPDIR/truerepublic-harness-ports` (pid owner; stale
  owner reclaimed via signal 0; never on Windows) + bind probe; released by `t.Cleanup`
  (runs after the later-registered process-stop cleanups). Fail-fast: `smokeExit`
  closed by the Wait goroutine without consuming `done`; `waitForSmokeHeight`/
  `waitForSmokeRPC` call `failIfExited` (80-line log tail). Assertions/timeouts unchanged.
- **F3 (open, not fixed)** — `server_lifecycle_test.go` `waitForNodeHeight`: the
  `cmd.ProcessState != nil && Exited()` check never fires because nobody calls
  `cmd.Wait()` before it; the single-node smoke only fails at the 60s deadline. Left
  as-is (callers Wait later; fixing needs a small restructure) — candidate follow-up.
- **Counts for #324/325 not yet bumped**: the 4 new root tests change `root` 246 → 250
  and Go 2,132 → 2,136 on main (pass-event method) — verify with the jq count and update
  docs before the PR; note GH-306 also bumps counts, so rebase order matters.

## Review owed to Codex
- GH-306 is consensus/economics code (TRR-02/TRR-04). The candidate was authored by Codex
  and only reviewed by Claude here; F1 is Claude-authored. Independent review is still
  required by the issue acceptance criteria before the protected PR.

## Next step (exact)
1. Review `agent/claude/GH306-reconcile` (`ee7a004..192dccf`); decide F2.
2. Check the remaining #324/#325 gate results (logs:
   `/private/tmp/claude-501/-Users-gio/600e71b2-a2c2-4dbd-ae72-df4dc5d1d30d/scratchpad/gh324/`,
   `summary.txt`), bump counts, add BRIDGE evidence on the branch.
3. After the 2026-10-01 reset: #324/#325 flake PR first (plan M1), then GH-306 protected PR
   from this branch on exact main (resolve the ACTION_LOG append conflict if GH-327/328 lands first).
4. Publish GH-327/328 and do the Pages link readback.

## Open decisions for Gio
- Publication of the GH-328 landing change (GitHub Pages) — separate approval.
- Dependabot PRs #326/#329 — plan says reconcile, don't merge blind.
