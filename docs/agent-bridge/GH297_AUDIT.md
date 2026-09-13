# TrueRepublic GH-297 — ZKP Protocol Freeze Audit

> Scope: frozen protocol manifest, read-only repository contract, security/CI
> wiring and public rollout state · Date: 2026-09-13 · Local result: APPROVE

## Outcome

GH-297 freezes the current v2 anonymous-rating protocol contract without
changing circuit, runtime, consensus or client submission behavior. The
manifest digest-binds the existing test-only circuit specification and pins
the exact circuit/profile versions, public-input order, canonical BN254
encoding, one-vote nullifier semantics, recipient-bound signal and mandatory
version/activation rules.

The existing CS/PK/VK remain forge-capable single-party toxic-waste fixtures.
Production prover integration, reproducible production artifacts, ceremony
provenance, browser-to-chain compatibility and independent cryptographic/
privacy review remain separate open gates. `production_ready=false` and
`submission_allowed=false` remain enforced.

## Verification

- PASS: source-spec SHA-256
  `27fb8e4eaeb365706535f107c0d6a39a282e423f01385c14495f3f06414c4ab7`.
- PASS: 62/62 focused contract events, including 48 semantic mutations,
  source-digest drift, unsafe/symlink/non-regular evidence and five strict-JSON
  failures.
- PASS: focused race/coverage, package and repository `go vet`, and exactly
  676/676 governance package plus 2,117/2,117 complete serial Go events.
- PASS: composed threat-model/security-review contracts, documentation
  consistency, Apache-2.0 policy, JSON validity and diff hygiene.
- PASS: Kimi K3 final read-only independent review — no blocker and no
  high/medium issue. Kimi authored only the bounded Go contract test; Sol
  reviewed and integrated every line.

Client and Rust code did not change. Their 319 and 26 established cases remain
in the published arithmetic; protected CI must rerun applicable workflows on
the exact PR head. Staticcheck, secret scanning and full repo-wide race remain
protected-CI gates because the host was below 1.1 GiB free and no foreign cache
or process was removed.

## Accounting

After protected merge, the canonical tracker moves to 36/59 overall and 36/51
phase work. Phase 2 becomes 2/7 canonical items; Phase 6 remains 6/7, Phase 7
remains 3/10 and production readiness remains false.
