id: CI-RESET-A1
status: ok
worker: claude
snapshot: 2026-10-01T01:11:16Z, one bounded read-only snapshot (gh api / gh pr / gh run; no fetch, checkout, build, rerun, comment or write)
FACTS
- main (live): 1283a4452d36784ea962f7fc8ba13f9ad45472cd (#302, 2026-09-19T20:40:49Z) — unchanged since the CI reset.
- Branch protection main: required checks check, go-static, go-vuln, secret-scan, node-audit-client, rust-audit, custom-query-retirement, mobile-retirement, web-wallet-retirement; strict=false; 1 approving review; dismiss_stale_reviews=true; enforce_admins=false; no bypass allowance.
- #332 fix(client): patch brace-expansion audit advisories — OPEN, ready, base main@1283a445, head 84ea15a1f423756c10eb0d3bdee316a7d527ce24 (= local agent/claude/GH331-client-audit), MERGEABLE/BLOCKED, reviewDecision REVIEW_REQUIRED (0 reviews), updated 2026-09-30T23:46:35Z. Checks: 22 pass, 1 skipping (cross-run rebuild comparison, https://github.com/NeaBouli/TrueRepublic/actions/runs/36792353053/job/110151050907), 0 fail. Files: BRIDGE.md, client-web/package.json, client-web/package-lock.json.
- #330 test: harden harness ports and child-exit handling — OPEN, DRAFT, base main@1283a445, head 994da7c46bc0b3e42f9750b4fb99c93a09dfd7b6 (= local agent/claude/GH324-325-port-harness), MERGEABLE/BLOCKED, REVIEW_REQUIRED (0 reviews), updated 2026-09-30T22:58:49Z. Checks: 25 pass, 1 skipping (cross-run rebuild comparison, run 36785225221), 1 fail: node-audit-client (Security Scan run 36785225148, head 994da7c, 2026-09-30T22:22:49Z, https://github.com/NeaBouli/TrueRepublic/actions/runs/36785225148/job/110124915397). Heavy exact-head jobs on 994da7c: multi-validator-recovery, capacity-qualification, concurrency-replay, governed-upgrade, docker-restart-smoke = pass.
- Failed-check classification: global/pre-existing defect, not a branch defect. Log: "Blocking high/critical npm advisories: brace-expansion, minimatch"; #330 changes no client-web file; the fix is #332.
- Other open PRs: #329 dependabot thiserror (contracts/Cargo.lock) — no overlap; #326 dependabot client-maintenance (client-web/package.json + lock) — overlaps #332; #317 docs/GH-316-comprehensive-status (BRIDGE.md, README.md, docs/agent-bridge/ACTION_LOG.md, …) — already CONFLICTING/DIRTY, overlaps #330 (BRIDGE.md, README.md).
- File overlap #332 ↔ #330: BRIDGE.md only (both append at the end -> textual conflict expected for whichever lands second).
- All open PRs authored by NeaBouli except dependabot; GitHub does not allow the author account to approve its own PR.
RECOMMENDATIONS (Codex-owned)
1. #332 is merge-ready on checks (all required pass on exact head 84ea15a). Remaining gate: the required approving review. Since the agents act as NeaBouli, approval needs another collaborator account or an explicit admin bypass decision (enforce_admins=false) — Gio's decision.
2. After #332 lands: refresh #330 onto the new main (rebase or merge main) — needed both for the BRIDGE.md append conflict and so node-audit-client sees the patched lock; strict=false alone does not fix the failing check because the check reads #330's own head. Then mark ready.
3. Exact-head gates for refreshed #330: all required checks green (expect node-audit-client pass), heavy jobs re-run on the new head (dismiss_stale_reviews=true -> review after refresh), then the approving review.
4. #326: expect a lock conflict after #332; let Dependabot recreate/rebase it, keep it separate per plan. #317 stays out of this chain.
5. Local evidence note: GH324C1's locally infra-blocked long gates are covered by these exact-head CI passes on 994da7c (pre-refresh head).
