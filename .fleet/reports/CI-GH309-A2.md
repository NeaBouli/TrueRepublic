id: CI-GH309-A2
status: ok
worker: claude (interactive stand-in; read-only gh API/log reads, no rerun, no GitHub write, no worktree edit)
snapshot: 2026-10-01T10:14Z
main: 1283a4452d36 (unchanged)
pr332: open, ready, base main, head 84ea15a, MERGEABLE/BLOCKED, REVIEW_REQUIRED; 22 pass, 1 skip, 0 fail.
pr336: draft, base main, head af0608b, MERGEABLE/BLOCKED, REVIEW_REQUIRED; 20 pass, 1 skip, 2 fail (build, node-audit-client).
pr337: draft, base agent/claude/GH335-modulequery-fetch, head bbc067d, MERGEABLE/UNSTABLE; 12 pass (incl. browser-quality on bbc067d), 1 skip, 1 fail (build).
pr338: draft, base main, head 8982d6a, MERGEABLE/BLOCKED, REVIEW_REQUIRED; 19 pass, 1 skip, 3 fail (build, node-audit-client, check).
pr339: draft, base agent/claude/GH309-custody-core, head d7e9249, MERGEABLE/UNSTABLE; 12 pass, 1 skip, 1 fail (build).
failures:
  - node-audit-client (#336 job 110245822007, #338 job 110294307352) and Client Web CI / build step "Audit dependencies" (#336 110245821694, #337 110279439824, #338 110294307581, #339 110311365034): "Blocking high/critical npm advisories: brace-expansion, minimatch", exit 1 — the known advisory fixed by #332; all earlier build steps passed.
  - NEW, real branch defect: Docs Consistency Check / check on #338 (run 36839314755, job 110294307557, required context on main): "FAIL services (status 17, source 19)" — scripts/check-consistency.sh:368 counts non-test files in client-web/src/services; #338 adds custodyEnvelope.ts and identityVault.ts but docs/status.json web_client.services is still 17. Only this one inconsistency was reported.
  - Consequences: #339 adds identityMigration.ts and previewIdentityHash.ts -> services must be 21 once it targets main (check does not run on its non-main base yet). The local GH309B2D1 branch adds components/zkp/IdentityCustodyNotice.tsx -> components 45 vs status 44 (services 21 vs 17); stores stay 9, routes unchanged.
fix_proposal: one narrow docs-inventory change per PR head (docs/status.json web_client.services 19 on #338; 21 on #339; components 45 + services 21 on GH309B2D1), plus any README/wiki line the check derives from status.json if it reports one. No source change. Owner/branch per Codex (the #338/#339 heads are accepted and pushed by Codex).
order: unchanged — #332 first; then #336 -> #337 and #338 -> #339 -> GH309B2D1; each with the inventory fix before leaving draft.
owner_action: #332 still needs one independent approving review or Gio's documented admin decision.
disk: one check at 10:15Z — 9.1 GiB free (was 3.1 GiB), above the 5 GiB rule.
next: Codex decides the inventory-fix ownership and the Chromium headless shell reinstall for GH309B2D1.
