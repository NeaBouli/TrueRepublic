# TrueRepublic GH-219 — Audit

> Scope: community Apache-2.0 publication, provenance boundary, public
> documentation, machine policy and verification · Date: 2026-08-26 · Result:
> **0 FAIL / 3 WARN / 12 PASS**

## Outcome

GH-219 is ready for protected review. The exact community decision is bound to
the machine-readable state and implemented consistently across the canonical
root text, notice, component metadata, maintained documentation, landing page,
roadmap and wiki sources. No runtime, consensus, dependency, release,
deployment, production or rollout-credit change is present.

## Priority matrix

| Priority | Count | Outcome |
|---|---:|---|
| P0 — critical | 0 | None |
| P1 — high | 0 | None |
| P2 — material | 0 | None after remediation |
| P3 — follow-up | 3 | Explicitly bounded warnings below |

## PASS findings

1. Root `LICENSE` matches the canonical Apache-2.0 text byte-for-byte by its
   pinned SHA-256 digest.
2. `NOTICE` records contributor-retained copyright, collective attribution as
   “TrueRepublic contributors,” and no central corporate owner.
3. The exact GH-219 governance-comment URL is bound in
   `configs/legal/license-decision.json` and the deterministic policy.
4. Maintained source/documentation scope and all provenance exclusions are
   consistent across the manifest, `NOTICE`, `REUSE.toml`, README,
   CONTRIBUTING, client documentation and public surfaces.
5. npm and every maintained Cargo package declare the selected project
   identifier consistently; Go correctly relies on the root project grant.
6. Brand assets, artwork, historical PDFs, archived evidence, the retired UI
   prototype and third-party materials are not silently included.
7. README, landing page, roadmap, whitepapers, Alpha/V4 architecture, local
   wiki sources and status JSON expose the same current project state.
8. Public arithmetic remains 2,094 tests, rollout 35/59, phase work 35/51,
   Phase 6 at 6/7 and production false; GH-219 earns no rollout checkbox.
9. The policy is portable to macOS Bash 3.2 and passes 31 positive/adversarial
   fixtures, including wrong-decision-record and conflicting-identity cases.
10. Full Go build, vet and Race/Coverage pass across all maintained packages;
    full Rust format, Clippy-with-warnings-denied, workspace build and 26 tests
    pass.
11. The maintained client passes install, lint, 10 policy tests, 309 Vitest
    cases with 4 documented skips, production build, bundle budget and the
    high/critical npm audit gate. Chromium/Firefox browser quality passes
    locally.
12. Module verification, Go vulnerability policy and its fixtures, Staticcheck,
    repository secret scan and its fixtures, RustSec audit, documentation
    consistency, release/rollout contracts and retirement contracts pass.

## WARN findings

1. `REUSE.toml` is intentionally a maintained-scope declaration, not a claim
   of complete REUSE conformance. Excluded provenance-gated material must stay
   excluded until file-specific evidence exists.
2. Future binary or image distributions still require a release-specific
   third-party notices bundle and per-component review. GH-219 publishes the
   project grant but does not create or approve a release.
3. The local macOS 12 Playwright runtime provides a frozen WebKit build whose
   protocol is incompatible with the current driver. Chromium and Firefox
   passed 26 cases with one intended mobile-keyboard skip; the protected Linux
   browser workflow remains the authoritative WebKit gate before merge.

## Independent review record

Kimi K3's first read-only review identified under-coverage of maintained
configuration surfaces, accidental inclusion of a historical session summary,
and index/worktree divergence. Sol remediated all three findings by tightening
the REUSE boundary, removing the historical file from the maintained scope and
staging one exact final diff. Kimi's final read-only review returned APPROVE
with no P0–P2 finding. Its one completeness observation was also applied:
`docs/status.json` is now explicitly included as maintained machine-readable
documentation.

## Merge boundary

Merge is permitted only after the protected exact-head checks pass with no
unresolved review threads. After merge, GitHub license detection, the separate
GitHub wiki repository, exact `main`, Pages and both Bridge records must be
verified before GH-219 is closed.
