id: CI-GH309-B2D1-WAIT
status: ok
worker: claude (interactive stand-in; read-only gh reads at 2026-10-01T10:14Z, no GitHub write, no fetch/checkout, no rerun; GH309B2D1 untouched)
main: 1283a4452d36
heads:
  - #332 agent/claude/GH331-client-audit @ 84ea15a -> main; ready; MERGEABLE/BLOCKED; REVIEW_REQUIRED; checks 22 pass / 1 skip / 0 fail.
  - #336 agent/claude/GH335-modulequery-fetch @ af0608b -> main; draft; MERGEABLE/BLOCKED; REVIEW_REQUIRED; 20 pass / 1 skip / 2 fail.
  - #337 agent/claude/GH309-client-containment @ bbc067d -> #336 branch (stacked); draft; MERGEABLE/UNSTABLE; no review decision (non-main base); 12 pass / 1 skip / 1 fail.
  - #338 agent/claude/GH309-custody-core @ 8982d6a -> main; draft; MERGEABLE/BLOCKED; REVIEW_REQUIRED; 19 pass / 1 skip / 3 fail.
  - #339 agent/claude/GH309-migration-core @ d7e9249 -> #338 branch (stacked); draft; MERGEABLE/UNSTABLE; no review decision; 12 pass / 1 skip / 1 fail.
failing_contexts:
  - #336 node-audit-client (job 110245822007) and build "Audit dependencies" (110245821694): "Blocking high/critical npm advisories: brace-expansion, minimatch" -> #332 advisory only.
  - #337 build "Audit dependencies" (110279439824): same advisory only; browser-quality on bbc067d passed (Chromium/Firefox/WebKit).
  - #338 node-audit-client (110294307352) and build "Audit dependencies" (110294307581): advisory only. Docs Consistency Check / check (110294307557, required on main): "FAIL services (status 17, source 19)" -> real branch defect, docs/status.json web_client.services must be 19 (Codex fixes #338/#339 separately).
  - #339 build "Audit dependencies" (110311365034): advisory only.
client_audit_executed: yes on #338 (node-audit-client + build audit step) and on #339 (build audit step); node-audit-client and the docs check are not triggered on #339's non-main base yet. After retargeting to main #339 needs services 21.
collisions: no source-file overlap between the #336/#337 stack (moduleQuery, membership, OnboardingFlow, browser spec) and the #338/#339 stack (custodyEnvelope, wallet, identityVault, identityMigration, previewIdentityHash, zkp); only the append-only docs/agent-bridge/ACTION_LOG.md and docs/status.json inventory. GH309B2D1 (local) merges both stacks and also edits OnboardingFlow and the #337 spec, so it must land after #337 and #339.
safe_order: #332 -> (#336 -> retarget #337 to main -> #337) and (#338 with services 19 -> retarget #339 to main with services 21 -> #339) -> GH309B2D1 rebased/retargeted on top with components 45 / services 21. Re-run hosted CI on each new base before its merge.
owner_action: #332 needs one independent approving review or Gio's documented admin merge decision; nothing else needs Gio.
