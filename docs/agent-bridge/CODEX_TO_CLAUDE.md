# Codex to Claude Code

## Handover entry

Date: 2026-07-11 22:08 EEST
Agent: Codex
Mode: AUDIT / FIX / SECURITY / DEVOPS
Task: GH-4 foundation, GH-11 canonical PNYX cap, and GH-14 bank escrow
Result: PR #9 head `acfc3d5` is fully green and awaits independent approval;
PR #15 head `e0ff339` is fully green; PR #16 is mergeable and
its local plus GitHub CI matrix is green after two audit fixes. See `PROJECT_STATE.md`, `ACTION_LOG.md`,
`PR15_AUDIT.md`, and `PR16_AUDIT.md`.
Known risks: runtime issuance/custody/DEX invariants remain in the later stacked
issues; legacy wallets remain unsafe; Rust dev-tooling warnings remain.
Next recommended action: publish and verify the PR #16 review remediation, then
proceed to GH-13 while the required PR #9 approval remains pending. Do not
bypass branch protection.
Do not touch: `/Users/gio/Desktop/repos/TrueRepublic` dirty legacy checkout.
