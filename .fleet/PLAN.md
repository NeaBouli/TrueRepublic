# TrueRepublic Fleet Plan

## Active milestone

GH-327 — publish a bounded architecture map and remaining-work execution plan before the next implementation block.

## Current verified baseline

- Base: `origin/main` at `1283a4452d36784ea962f7fc8ba13f9ad45472cd`.
- Current product implementation remains on the preserved dirty GH-306 worktree and is out of scope here.
- Canonical rollout remains 36/59 and production readiness remains false.
- Hosted CI quota is expected to reset on 2026-10-01; this documentation block must not trigger avoidable hosted runs before then.

## Parallel briefs

1. `GH327A` — Claude Code owns the code-grounded architecture map and PlantUML sources.
2. `GH327B` — Kimi owns the comprehensive remaining-work plan, dependency ordering, agent allocation, and Cloudflare-derived audit timing decision.

The briefs have disjoint write ownership. Neither worker may edit product code, Bridge history, public status, workflows, or GitHub.

## Decision and small-task routing

- `jev` may cheaply support one bounded yes/no, choice, or score decision when the context is public and secret-free. It receives no source code, diff, security finding, internal host, credential, private log, or non-public document. Its numeric answer is advisory; Codex retains the decision and records the question/value when it materially affects routing.
- Grok is Worker C for small, explicit tasks with a narrow file set (normally no more than five files), no repository exploration, no architecture ownership, and no merge/deploy authority.
- JEV can trigger an additional review but cannot waive deterministic Security, migration, cryptography, payment, authentication, deployment, or release gates.

## Integration gate

- Worker report is valid and within the brief.
- Documentation is internally consistent and references real files/symbols or explicitly marks nodes open.
- PlantUML sources parse when a local renderer exists; Mermaid remains visible without it.
- No rollout credit, production claim, deployment, live probe, or feature implementation.
- Codex integrates, updates append-only coordination, and decides publication timing around the CI reset.

## Next milestone after GH-327

Complete GH-306 from its preserved candidate, then use the map as the mandatory boundary for the next audit-remediation node. Security priority and exact order are determined by GH327B and the audit register, not by adding new feature scope.

## Local result — 2026-09-27 EEST

GH327A and GH327B are integrated locally. Required cross-reviews found only bounded documentation corrections, now applied. Publication remains intentionally deferred until the 2026-10-01 hosted-CI reset. Next implementation milestone: reconcile and close GH-306 on exact main.
