# GH-273 audit — cross-run exact-commit rebuild evidence

Status: implementation merged and hosted exact-commit pair verified; public
closeout in progress.

## Scope and base

- Issue: [GH-273](https://github.com/NeaBouli/TrueRepublic/issues/273)
- Implementation PR: [#274](https://github.com/NeaBouli/TrueRepublic/pull/274)
- Final implementation head: `1e7635f4ae4e16cfe080a6c8229a30596b663f4c`
- Exact merge/main commit: `3b0d1639bb40c7df6733dd13a86252e1c8c9efd3`
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
- Protected review of PR #274 then identified the candidate-artifact layout,
  omitted-candidate-claims, timestamp-tolerance and premature Pages wording
  defects plus three test/inspection hardenings. Sol corrected all confirmed
  findings. The invalid-claims current receipt intentionally remains
  `in_progress` with an empty conclusion because that is the workflow-bound
  verifier state while the comparison run executes.
- Kimi independently re-reviewed the exact remediation and returned
  **APPROVE**, with no P0, P1 or P2 finding. Its focused package, repository,
  script, YAML and diff reruns all passed.

## Local evidence

- `make verify`: Go package selector, build, Vet and complete Race/Coverage pass.
- Final post-review enumeration: 1,974 Go cases, including root/application 232
  and `crossrunevidence` 107. With 26 Rust and 319 maintained-client cases, the
  authoritative standard total is 2,319.
- `crossrunevidence` coverage: 86.5%.
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

## Protected merge and exact-main evidence

- PR #274 passed its complete applicable exact-head matrix with all review
  threads resolved and merged as
  `3b0d1639bb40c7df6733dd13a86252e1c8c9efd3`.
- Exact-main Go `33464556692`, Reproducible Linux Daemon `33464556701`, Docs
  `33464556713`, Security `33464556731`, and Pages `33464555897` succeeded.

## Hosted exact-commit pair

- Baseline run
  [33465480131](https://github.com/NeaBouli/TrueRepublic/actions/runs/33465480131):
  successful protected `workflow_dispatch` on exact main, candidate artifact
  `9784898032`, digest
  `sha256:64a3a4dafa6ec1bacbd377d46855ee65f7e38dd21a11804d06017e6be4410514`.
- Comparison run
  [33466167289](https://github.com/NeaBouli/TrueRepublic/actions/runs/33466167289):
  successful protected `workflow_dispatch` on the same exact main, candidate
  artifact `9785126589`, digest
  `sha256:8a2f19714df68d160ec17ff7a0f088fbb328fb5f687326d9bc22509e4d737dfc`.
- Pair artifact `9785138134`, digest
  `sha256:cb646d73943223699705d01a904ea2b043deb37cd1aaddfe613185780ca24292`,
  contains exactly `cross-run-report.json` plus the seven declared evidence
  JSON members. No undeclared member or payload exists.
- Report schema `truerepublic.cross-run-report/v1` is valid with zero
  violations. Repository, workflow, branch, event and exact commit match; both
  daemon targets and all four OCI target identity vectors match.
- Baseline/current candidate manifests are byte-identical with SHA-256
  `9a1aa898fa078e3938daae7d15ee48ea529b44ae7bd8456444e2afd13e2c5bb9`;
  reports are byte-identical with SHA-256
  `27a133ca663cda17a180a40610e8b0b44e2a42f3dd7ee8281217726183e706b3`.
- All real-tag, ref-push, signing, attestation, publication, deployment,
  production and long-term-hermetic claims are explicitly false. The 14-day
  hosted expiry is one second before the nominal timestamp boundary in both
  runs, within the contract's documented 60-second GitHub tolerance.

The different candidate artifact ZIP digests reflect run-specific archive
packaging. GH-273 deliberately compares the verified inner candidate
identities. This pair does not prove long-term hermeticity or complete any
tagged/signed/published/deployed release gate; rollout remains unchanged.
