# GH365 — integration/authorization evidence (Claude, 2026-10-10T18:34Z)

Purpose: bounded decision pack for Codex/Gio. No gate was rerun; values read fresh from GitHub.

## Candidate
- PR #365 "fix: reconcile Go and client advisories with bounded exception evidence"
- head 6d1c79d9227643ef23972b7d5889563fe1684164, base main ff63e9d087f1bbf09c01b3182b896ce30ac9116c (main unchanged since morning)
- State: Draft, mergeable=MERGEABLE, mergeStateStatus=BLOCKED, reviewDecision=REVIEW_REQUIRED
- Check-runs on exact head: 29 success, 1 skipped (cross-run rebuild comparison, conditional), 0 failure; incl. go-vuln, node-audit-client, rust-audit, secret-scan, reproducible OCI amd64/arm64, release evidence, multi-validator-recovery, governed-upgrade.
- Diff vs main: 81 files, +459/-201.

## Time-sensitive fact (verified in source)
- main configs/security/gates.json: four go_vulnerability_exceptions (GO-2026-5932, GO-2026-4740, GO-2023-1881, GO-2023-1821) with expires 2026-10-13.
- scripts/check-go-vulnerabilities.sh accepts while expires >= today (UTC). So main's Go gate stays valid through 2026-10-13 and rejects from 2026-10-14 (UTC) unless #365 (renewed no-fix set to 2026-10-20, review 2026-10-16, GO-2026-4740 removed) or another policy change is on main.
- Additionally the gate rejects any reachable *fixable* advisory; new advisories since the last main run make fresh main-based PR runs red already (observed on #367/#371). #365 fixes the twelve fixable ones.

## What blocks integration (held, not Claude's to decide)
1. Lead acceptance/review of #365 (REVIEW_REQUIRED) and main merge — Codex/Gio.
2. Project-local Codex Security gate (.fleet/CODEX_SECURITY_GATE.md): scan NOT RUN, scan authorization NOT GRANTED. Integration needs either an authorized scoped scan of 6d1c79d (Gio: destination, transmitted paths, exclusions, cost ceiling) or an explicit bounded lead exception recorded for this candidate.
3. Renewal of the three no-fix exceptions is part of #365's content; it must not be renewed silently elsewhere.

## Recommendation (Claude)
Decide before 2026-10-13 23:59Z: either (a) authorize the scoped scan of 6d1c79d with exact bounds, or (b) record a bounded lead exception for #365 (dependency-only change, all 29 CI gates green), then mark ready + merge (Codex/Gio). Without either, main's Go gate rejects from 2026-10-14 and every later PR stays red.
