# Codex to Claude Code

## Handover entry

Date: 2026-07-12 03:34 EEST
Agent: Codex
Mode: AUDIT / FIX / SECURITY / DEVOPS
Task: GH-4 foundation through GH-10 DEX custody
Result: PR #9 head `acfc3d5` is fully green and awaits independent approval;
PR #15 head `e0ff339` is fully green; PR #16 head `fa693a8` is mergeable and
fully green; PR #17 branch `fix/GH-13-cap-issuance` is rebased, its final
full-review remediation is locally/GitHub green, and all five threads are
resolved. PR #18 is rebased onto final PR #17 and locally green with bank-backed
DEX reserves, provider LP ownership, canonical burns, and authority checks; its
GitHub refresh is pending. Use the live PR head as source of truth. See
`PROJECT_STATE.md`, `ACTION_LOG.md`, `PR17_AUDIT.md`, and `PR18_AUDIT.md`.
Known risks: GH-12 custom-genesis/runtime invariants remain open; legacy wallets
remain unsafe; Rust dev-tooling warnings remain.
Next recommended action: complete PR #18 GitHub verification, then begin GH-12
without collapsing the ordered stack. The independent PR #9
approval remains required; do not bypass branch protection.
Do not touch: `/Users/gio/Desktop/repos/TrueRepublic` dirty legacy checkout.
