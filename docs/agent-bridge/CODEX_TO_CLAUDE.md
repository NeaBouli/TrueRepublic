# Codex to Claude Code

## Handover entry

Date: 2026-07-11 22:08 EEST
Agent: Codex
Mode: AUDIT / FIX / SECURITY / DEVOPS
Task: GH-4 foundation, GH-11 canonical PNYX cap, and GH-14 bank escrow
Result: PR #9 head `acfc3d5` is fully green and awaits independent approval;
GH-11 is green in PR #15; GH-14 is rebased onto it in PR #16 and its two audit
findings are fixed locally. See `PROJECT_STATE.md`, `ACTION_LOG.md`,
`PR15_AUDIT.md`, and `PR16_AUDIT.md`.
Known risks: runtime issuance/custody/DEX invariants remain in the later stacked
issues; legacy wallets remain unsafe; Rust dev-tooling warnings remain.
Next recommended action: finish PR #16 verification and GitHub publication,
then proceed to GH-13 while the required PR #9 approval remains pending. Do not
bypass branch protection.
Do not touch: `/Users/gio/Desktop/repos/TrueRepublic` dirty legacy checkout.
