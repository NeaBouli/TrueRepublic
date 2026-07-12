# Audit Reports

## Status

TrueRepublic is in an internal recovery audit. There is no completed external
mainnet, cryptographic, consensus, or operations audit and no production
approval.

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

Reports live under [`docs/agent-bridge/`](https://github.com/NeaBouli/TrueRepublic/tree/fix/GH-8-docs-final/docs/agent-bridge).
They document local and GitHub evidence but do not substitute for independent
review.

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
