# HANDOVER-CLAUDE-2026-10-24 — TrueRepublic full takeover packet

Recorded 2026-10-23T21:10Z (2026-10-24 00:10 EEST) by Claude Code (Opus 5.5).
Purpose: if Claude Code drops out, Codex (or another agent) can take over TrueRepublic
without asking and without re-research. This file supersedes
`.fleet/HANDOVER-CLAUDE-2026-10-10.md` as the entry point; that file stays as history.

**Operating state: PAUSED by Gio** since 2026-10-10T19:18Z. Nothing was built for
TrueRepublic after that. No agent is working on TrueRepublic right now. Start a new
block only after Gio explicitly reactivates TrueRepublic.

All facts below were read fresh on 2026-10-23T21:07Z unless marked "historical".

---

## 1. Read order for the taker (10 minutes)

1. This file.
2. `/Users/gio/AGENTS.md`, `/Users/gio/.claude/CLAUDE.md` (Claude→Codex CLI task mode
   from 2026-10-13), `/Users/gio/BRIDGE.md` (search "TrueRepublic").
3. Codex's own packets (read-only, private coordinator, NEVER PUSH):
   - `/Users/gio/Documents/Codex/TrueRepublic-GH327/.fleet/HANDOVER-CODEX-2026-10-10-PAUSED.md`
     (full per-PR evidence table, retained qualifications, resume order)
   - `/Users/gio/Documents/Codex/TrueRepublic-GH327/.fleet/HANDOVER-CODEX-TO-CLAUDE-20261010.md` @ fca06f0
   - `/Users/gio/Documents/Codex/TrueRepublic-GH327/.fleet/CODEX_SECURITY_GATE.md`
   - `/Users/gio/Documents/Codex/TrueRepublic-GH327/docs/architecture/MAP.md`
4. Project `CLAUDE.md`, `BRIDGE.md` tail, `.fleet/PLAN.md`, `docs/agent-bridge/TODO.md`
   in this worktree.
5. Reports in `.fleet/reports/`: `GH365-INTEGRATION-EVIDENCE-CC.md`, `GH315D2-CC.md`,
   `GH315D2-REVIEW-KIMI.md`, `GH315D2-evidence/d2-assertions.js`.

## 2. Roles, authority, hard limits

- Owner/decider: Gio. Owner decisions: official Codex Security scan authorization (scope,
  destination, cost), any security-exception renewal on main, money, access, production,
  mainnet, publication to live Pages/Wiki.
- Main merge + release gate: Codex/Gio only. Claude never merged to main and never deployed.
- Claude: main developer (implementation, tests, evidence, docs) on `agent/claude/*` branches.
- Helpers available to the coordinating agent (Gio 2026-10-13/15):
  - **Codex CLI**: `/Users/gio/.local/bin/codex-worker <PROJECT_DIR>` with the task on stdin
    (fresh `codex exec`, gpt-6.1-sol xhigh, workspace-write, no merge/production authority,
    separate worktree for concurrent edits, no secrets in the task).
  - **OpenCode** `/Users/gio/.local/bin/opencode` v2.0.29, model `opencode/space-bunny-free`
    only, other providers denied in `~/.config/opencode/opencode.json` (never loosen, no paid
    models, no updates). Call:
    `opencode run --standalone --model opencode/space-bunny-free --format json "<task>"`.
    **Security finding (verified 2026-10-15 in the binary):** OpenCode auto-loads every
    `AGENTS.md` from the cwd up to `$HOME` plus `~/.config/opencode/AGENTS.md`. Run it only
    from an isolated folder outside `/Users/gio` (e.g. a scratch dir under `/private/tmp`)
    containing copies of the approved files only; copy results back after review.
    Otherwise `/Users/gio/AGENTS.md` (internal host data/rules) is sent to the service.
  - **Kimi 3**: `~/agent-fleet/bin/dispatch.py run --no-fallback kimi <brief>`; creates
    `claude-standin-20260930-wt/kimi-<id>`. Used for GH315D2 review (ok).
    Note memory 2026-10-21: Kimi weekly limit reached; fallback reviewer is Codex CLI.
- Codex desktop app: subscription cancelled 2026-10-11; no thread watcher, no `codex queue`.
- Guardrails: no main push, no force-push, no sudo, no `rm -rf` outside worktree/tmp, no
  `~/.ssh`/Keychain/wallet access, no secrets in briefs/logs/prompts, never push the private
  coordinator `/Users/gio/Documents/Codex/TrueRepublic-GH327`, never edit Codex source
  worktrees under `/Users/gio/Documents/Codex/TrueRepublic-*`, preserve the divergent original
  checkout `/Users/gio/Desktop/repos/TrueRepublic` (branch feature/GH-300-browser-zkp-prover @ f5de5a1).
- Claim: `~/agent-fleet/standin/claims/019f5012-64a0-7233-ad73-3b2b36a41ffa.md`
  (Codex thread 019f5012-64a0-7233-ad73-3b2b36a41ffa). Taker: update status line, do not delete.

## 3. Repository and project facts

- Repo `NeaBouli/TrueRepublic` (public). Cosmos SDK chain + `client-web` (React/Vite) +
  Rust/CosmWasm `contracts` + public Landing `docs/index.html` on GitHub Pages.
- Version label v0.4.0, recovery audit active, **not production-ready**.
- Main-backed public numbers: 2,486 standard-suite cases, rollout 36/59, production false.
  Do not change these without a non-overlapping recount on accepted main.
- Toolchain pins: Go 1.26.6, Cosmos SDK v0.50.15, CometBFT v0.38.26, ibc-go v8.7.0,
  wasmd v0.53.4 / **wasmvm v2.2.8 (consensus-critical, retained by Gio)**.
- Verification commands: see project `CLAUDE.md` "Required verification".
- Stacked Draft PRs do not auto-trigger docs CI: `gh workflow run docs-check.yml --ref <branch>`.
- Go gate: `scripts/check-go-vulnerabilities.sh` + `configs/security/gates.json`. An allowance
  is valid while `expires >= today (UTC)`; any reachable fixable advisory is rejected.

## 4. What exists (built, all unmerged unless stated)

### main
- `main` = **ff63e9d** (2026-10-04, "Merge pull request #357"). Unchanged since then.

### Claude's own branches (this stand-in)
| Branch | Head | PR | Content | State |
|---|---|---|---|---|
| agent/claude/truerepublic-standin-20260930 | see git log | none | coordinator: BRIDGE/PLAN/handover/reports only | pushed after this commit |
| agent/claude/GH315D2-landing-semantics | d3728f5 | #372 Draft, base agent/codex/GH315-discovery (#371 @ 5f99c25) | `docs/index.html` only: visible wrapping nav ≤1024px (no hidden nav), 44×44 nav targets, `:focus-visible`, `scroll-margin-top`, `<nav aria-label="Primary">`, `<main id="main">`, footer h4→h2, decorative icons `aria-hidden`, logo `alt=""` | docs-check run 38078994518 success; Playwright 9 viewports (320/390/768/769/820/1024/1025/1180/1440) candidate PASS 9/9, base FAIL 9/9; screenshots inspected (`TrueRepublic-wt/_evidence/GH315D2-d3728f5`); Kimi review ok. **Review owed to Codex: integrated visual review.** Not live. |
| agent/claude/GH309-* (5) | 337 825e6f2, 338 86fc26f, 339 30fd630, 340 5ccbbae, 341 4e13402 | #337→#341 stacked on #365 | identity-custody chain (TRR-19/20) | all green historically (see Codex table) |
| agent/claude/GH300-* (4) | 343 79d0e07, 344 3d178b3, 345 b2c0934, 346 3055af2 | #343→#346 stacked on #341 | test-only browser Groth16 prover path + VK admission gate | green historically; no production artifacts/ceremony |
| agent/claude/GH308-quarantine | b3a0122 | #334 | CosmWasm prototype quarantine | superseded by Codex #370 integration |
| agent/claude/GH324-325-port-harness | 994da7c | #330 | harness port fixes | superseded by Codex #369 lineage |

Other older Claude worktrees are archived (`TrueRepublic-wt/*.ARCHIVED.md`); restore with
`~/agent-fleet/bin/worktree-restore.sh <path>`. Do not recreate.

### Codex branches (integration candidates, all Draft)
#365 GH364 advisories @ 6d1c79d (base main) · #366 GH314 whitepaper @ 25f69c7 · #367 Landing
checkpoint @ 878a0a1 · #368 GH313I1 operator guidance @ 4e628ec · #369 GH306I2 + V042 upgrade
@ 0c29a9f · #370 GH308I1 contract quarantine @ 80b518d · #371 GH315 discovery @ 5f99c25 ·
#363 source-map @ 4900d58 · #361 cosmwasm-std 3.0.10 @ f931a53 · #358 GH356 mobile wallet @ 0edbed5.
All heads re-checked 2026-10-23: **identical** to Codex's 2026-10-10 table.

### Issues
Open tracker: #29 rollout epic, #303 audit register (TRR-01…42), tickets #300 #306–#316 #324
#325 #327 #328 #333 #355 #356 #362 #364 #373 #232. #373 = F1 from GH315D2 (License-section
secondary buttons unreadable, pre-existing, not in #372 scope).

## 5. Status quo — what changed while paused (fresh 2026-10-23)

1. **Main security is red.** Scheduled "Security Scan" on main failed 2026-10-12 and
   2026-10-19 (run 39224075570): go-vuln, node-audit-client, rust-audit FAIL
   (secret-scan, go-static, retirement checks success).
   - go-vuln: the four main allowances expired 2026-10-13 (GO-2026-5932, GO-2026-4740,
     GO-2023-1881, GO-2023-1821) **plus new reachable advisories** GO-2026-6599, -6600, -6603,
     -6605, -6607, -6608, -6609, -6610, -6611, -6612, -6613, -6617 (module/fixed versions not
     yet analysed — first job on resume).
   - rust-audit: rustls 0.23.45 → needs ≥0.23.46 (RUSTSEC-2026-0347/0348/0349);
     ruzstd 0.5.0 → ≥0.9.1 (RUSTSEC-2026-0346); paste 1.0.15 unmaintained warning (0344).
   - node-audit-client: 1 high severity (package not yet identified).
   Log copy: `gh run view 39224075570 -R NeaBouli/TrueRepublic --log-failed`.
2. **#365 is stale.** Its green CI (30 success/1 skip) is from 2026-10-10. Its renewed
   allowances (GO-2026-5932, GO-2023-1881, GO-2023-1821) expired **2026-10-20**, and the new
   advisories above are not covered. A rerun would fail. #365 needs a refresh commit
   (new triage + fixes/bounded renewals) before any acceptance — the Oct-10 evidence pack
   `.fleet/reports/GH365-INTEGRATION-EVIDENCE-CC.md` is historical.
3. Every PR based on main now fails go-vuln: #371 (go-vuln, node-audit-client), #374, #375,
   #377, #317. #360 and #329 show green only because their runs are old.
4. New Dependabot PRs (not touched by any agent):
   - #374 github-actions group (3 updates): FAIL build, release evidence contract, go-vuln, go-static, node-audit-client.
   - #375 **wasmvm 2.2.8 → 2.2.10** (go-maintenance): FAIL build, check, go-vuln,
     client-chain-integration, node-audit-client. wasmvm is consensus-critical and Gio
     retained 2.2.8 → HOLD; any change requires the consensus/recovery matrix and Gio's yes.
   - #377 client-maintenance group (11 npm updates): FAIL check, go-vuln, rust-audit.
5. Official Codex Security scan: still NOT RUN, not authorized. No lead exception granted.
6. Disk on the Mac: 3.8 GiB free on 2026-10-23 (was nearly full on 2026-10-10). Check
   `df -h /` before any Go/Docker build.
7. Unpushed coordinator commits (3260f4e and two before) are pushed together with this packet.
8. Kimi review worktree `claude-standin-20260930-wt/kimi-GH315D2-REVIEW-KIMI` has one
   untracked report already copied into `.fleet/reports/GH315D2-REVIEW-KIMI.md`; safe to archive.

## 6. Open work, in exact dependency order

0. **Owner decision (Gio):** how to restore main security. Options:
   (a) authorize a bounded official Codex Security scan (scope, destination, cost), or
   (b) record a bounded lead exception, and in both cases let an agent refresh #365.
   Without this, nothing can land on main.
1. **Refresh #365** (free engineering, no owner needed until merge):
   triage GO-2026-66xx (stdlib vs module; toolchain patch bump if stdlib), bump rustls ≥0.23.46
   and ruzstd ≥0.9.1 in `contracts` lockfile, fix node high advisory in `client-web`
   (`npm audit`, `npm ci`), re-evaluate the three no-fix allowances with fresh evidence and
   new dates (never silently renew), run full required verification + go-vuln/rust-audit/
   npm audit locally, push to a new agent branch or ask Codex to update #365.
   Then: independent review → main merge by Codex/Gio.
2. After #365 on main: rebase/compose GH309 #337→#338→#339→#340→#341 (exact order, controlled
   base reconciliation, no blind retarget, no duplicate custody build).
3. Then GH300 #343→#344→#345→#346 (test-only; no production artifacts, no toxic-waste keys).
4. Dependency-ordered Codex candidates: #366, #368, #369, #370 (on #361), Landing #367,
   #371 → #372 (GH315D2). Recount public numbers only on accepted main.
5. #373 License buttons (free bounded UI block; `docs/index.html`; visual gate required;
   candidate owner Kimi or Codex CLI). Base on #372 to avoid conflicts.
6. Dependabot: #374 and #377 after main is green (rebase, re-run); #375 HOLD (wasmvm);
   #360 cosmwasm-vm major HOLD; #329 thiserror patch; #317 conflicting docs status PR.
7. Ticket order constraints: #307 requires #306 merged; #312 requires #309 merged;
   #310 → #311. #333 CosmWasm instantiation policy open. T1 allocator owner choice HOLD.
8. Not started / long-term: production ceremony, independent crypto/privacy/consensus
   reviews, signed release, topology/soak/paging/canary, formal go/no-go (#29). V3/V4 deferred.

## 7. Review owed

- #372 GH315D2: integrated human visual review by Codex (UI change).
- Any refreshed #365: independent security review (security surface).

## 8. Exact first steps for the taker

1. `df -h /`; read claims; `git -C <this worktree> fetch` and confirm main = ff63e9d and PR heads
   as in section 4 (`gh pr list -R NeaBouli/TrueRepublic --state open`).
2. Confirm with Gio that TrueRepublic is reactivated and which option (section 6, item 0).
3. Append `## <UTC> [AUTOR: <agent>]` ACK to this worktree's `BRIDGE.md` and a pointer to
   `/Users/gio/BRIDGE.md`; update the claim status.
4. Start section 6 item 1 in a fresh worktree from `origin/agent/codex/GH364-go-advisories`
   (or main), one writer only.

## 9. Documentation discipline

After each step: project `BRIDGE.md` entry (UTC, author, heads, commands, real test results,
NOT RUN reasons), `.fleet/PLAN.md` line, `docs/agent-bridge/TODO.md`, ticket comment, central
`/Users/gio/BRIDGE.md` pointer. Landing/README numbers only from accepted-main evidence.
No secrets, tokens or host IPs in this public repository.
