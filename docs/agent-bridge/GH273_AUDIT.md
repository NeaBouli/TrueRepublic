# GH-273 audit — cross-run exact-commit rebuild evidence

Status: implementation locally verified; protected PR/merge and hosted pair
pending.

## Scope and base

- Issue: [GH-273](https://github.com/NeaBouli/TrueRepublic/issues/273)
- Branch: `feat/GH-273-cross-run-evidence`
- Exact base: `f13c1148b3ed979b4e65c0fe36e559d03312bf2d`
- Rollout remains 35/59; phase work 35/51; Phase 6 6/7;
  `production_ready=false`.

GH-273 adds a strict repository-only comparison of bounded metadata from two
distinct `Reproducible Linux Daemon` workflow executions of the same exact
`main` commit. It extends GH-258 same-job OCI identity parity and GH-261
same-run candidate aggregation. It does not create a tag, push a ref, retain
binary/OCI payloads, sign, attest, publish, deploy, approve production, or claim
long-term hermeticity.

## Implementation and review

- Kimi K3 implemented the bounded contract, strict Go package/CLI, scripts,
  fixtures, Make target, semantic repository mutation tests and workflow core.
  Kimi changed no Bridge/public docs and performed no external action.
- Codex Sol reviewed every delegated file, bound both artifact queries to the
  exact name plus `total_count == 1`, expanded the retained comparison artifact
  to exactly seven evidence JSON members plus report, synchronized public docs
  and threat model, and fixed the only Staticcheck finding (`S1016`).
- Claude Code's small read-only helper audit produced no result because its
  configured local budget was exhausted; no Claude diff exists.
- Kimi's independent final read-only review returned **APPROVE**, with no P0,
  P1 or P2 finding. Its sole P3 observation was closed by adding the maintained
  Candidate/Cross-run fixtures to the Apache-2.0 REUSE annotation.

## Local evidence

- `make verify`: Go package selector, build, Vet and complete Race/Coverage pass.
- Standard enumeration: 1,969 Go cases, including root/application 231 and
  `crossrunevidence` 103. With 26 Rust and 319 maintained-client cases, the
  authoritative standard total is 2,314.
- `crossrunevidence` coverage: 86.1%.
- Client: `npm ci`, lint, 319 tests, production build/budget, and high-level
  audit pass with zero vulnerabilities.
- Rust: format, Clippy with warnings denied, 26 tests pass; Cargo audit reports
  only the five policy-allowed warnings.
- Security: pinned Staticcheck v0.7.0, Gitleaks v8.30.1, Govulncheck v1.6.0,
  vulnerability positive/negative fixtures and repository security contract
  pass.
- Deterministic daemon, repeated OCI, candidate evidence, cross-run evidence,
  install lifecycle, release compatibility, rollout genesis, coverage, quality/
  fuzz, docs consistency, license policy/negative fixtures and diff gates pass.

## Hosted evidence pending

The implementation must still pass protected exact-head review and merge. Only
then may two distinct `workflow_dispatch` executions run on unchanged exact
`main`: first without a baseline, second with the successful first run ID. The
exact commit, run IDs, artifact IDs/digests and pair report will be appended
here after both runs succeed. Until then no hosted pair parity is claimed.
