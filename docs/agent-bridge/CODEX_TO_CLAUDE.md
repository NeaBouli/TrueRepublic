# Codex to Claude Code

## Handover entry

Date: 2026-07-11 21:45 EEST
Agent: Codex
Mode: AUDIT / FIX / SECURITY / DEVOPS
Task: GH-4 recovery foundation and GH-11 canonical PNYX cap
Result: PR #9 head `acfc3d5` is fully green and awaits independent approval;
GH-11 final audit fixes are rebased onto it in PR #15. See `PROJECT_STATE.md`,
`ACTION_LOG.md`, and `PR15_AUDIT.md`.
Known risks: runtime issuance/custody/DEX invariants remain in the later stacked
issues; legacy wallets remain unsafe; Rust dev-tooling warnings remain.
Next recommended action: complete refreshed PR #15 GitHub gates, obtain the
required independent PR #9 approval, and do not bypass branch protection.
Do not touch: `/Users/gio/Desktop/repos/TrueRepublic` dirty legacy checkout.
