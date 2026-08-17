# Current recovery candidate release notes

Contract: `truerepublic.release-compatibility/v1`

This is an **unreleased** repository candidate. Production: `false`. Tagged:
`false`. Published: `false`. Signed: `false`. The daemon version is the exact
source commit injected through `main.version`. `v0.4.0` is the recovery-line
label and maintained-client version; the same-named March tag is a separate
historical object and does not identify this candidate.

## Candidate changes

- `COMPAT-001` — pre-recovery state has no generic in-place migration.
- `COMPAT-002` — `x/staking` and `x/distribution` fail closed.
- `COMPAT-003` — retired clients are Git-history-only; `client-web` is Beta.
- `COMPAT-004` — legacy custom ABCI query aliases are removed.
- `COMPAT-005` — anonymous ZKP submission remains disabled.
- `COMPAT-006` — native artifacts use the verified install lifecycle.
- `COMPAT-007` — source-commit identity and exact `v0.4.1` upgrade boundary.
- `COMPAT-008` — bounded local ICS-20 compatibility evidence.

Exact actions, supported targets, unsupported surfaces and repository evidence
are in [the compatibility statement](current-candidate-compatibility.md) and
`configs/release/compatibility.json`. This document creates no tag, artifact,
signature, deployment, genesis freeze, production approval or go/no-go record.
