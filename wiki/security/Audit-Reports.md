# Audit Reports

## Status

TrueRepublic is in an internal recovery audit. There is no completed external
mainnet, cryptographic, consensus, or operations audit and no production
approval.

GH-294 adds a machine-verifiable
[independent-review preparation package](https://github.com/NeaBouli/TrueRepublic/blob/main/docs/security/INDEPENDENT_REVIEW_GUIDE.md)
with a versioned scope, evidence index and findings lifecycle. It is prepared
by the project and does not claim reviewer independence or complete the open
Phase-5 gate.

## Internal recovery artifacts

The repository contains scoped, evidence-backed reports for the ordered stack:

- PR #9 — foundation/toolchain/security
- PR #15 — canonical PNYX denomination and 21M cap
- PR #16 — governance/stake bank escrow
- PR #17 — cap-checked issuance
- PR #18 — DEX custody and LP ownership
- PR #19 — genesis reconciliation and runtime invariants
- PR #22 — ZKP authentication and replay resistance
- PR #23 — persistent PoD node lifecycle

Reports live under [`docs/agent-bridge/`](https://github.com/NeaBouli/TrueRepublic/tree/main/docs/agent-bridge).
They document local and GitHub evidence but do not substitute for independent
review.

The root `CODEX_AUDIT.md` is a retained July 2026 historical snapshot. Its
current-status claims are superseded by the GH-294 review guide, threat model,
and findings register; the original content remains available for provenance.

## Required independent work

- Review the ordered consensus/ledger stack before merge.
- Audit the Groth16 circuit, ceremony artifact, verifying key, prover, privacy,
  and recipient-binding design before anonymous voting is enabled.
- Exercise multi-node consensus, peer failure, IBC relaying/upgrades,
  backup/restore, monitoring, alerting, and incident rollback.
- Reassess dependencies and clients on the final merged commit.

## Reporting a security issue

Use the repository's private
[Security Advisory form](https://github.com/NeaBouli/TrueRepublic/security/advisories/new)
for sensitive vulnerabilities. Do not publish secrets, validator keys, wallet
material, or exploit details in a public issue.
