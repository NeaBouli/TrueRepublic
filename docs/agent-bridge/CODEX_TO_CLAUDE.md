# Codex to Claude Code

## Handover entry

Date: 2026-07-12 04:32 EEST
Agent: Codex
Mode: AUDIT / FIX / SECURITY / DEVOPS
Task: GH-4 foundation through GH-12 genesis/runtime invariants
Result: PR #9 head `acfc3d5` is fully green and awaits independent approval;
PR #15 head `e0ff339` is fully green; PR #16 head `fa693a8` is mergeable and
fully green; PR #17 branch `fix/GH-13-cap-issuance` is rebased, its final
full-review remediation is locally/GitHub green, and all five threads are
resolved. PR #18 is rebased onto final PR #17 and locally/GitHub green with
bank-backed DEX reserves, provider LP ownership, canonical burns, and authority
checks; CodeRabbit is rate-limited and substantive external review is pending.
PR #19 is rebased onto final PR #18 and locally green after removing an unsafe
hard-coded validator secret, deriving bank-backed bootstrap state only from real
CometBFT keys, and proving all four crisis invariants plus non-empty round trips.
Use the live PR head as source of truth. See
`PROJECT_STATE.md`, `ACTION_LOG.md`, `PR18_AUDIT.md`, and `PR19_AUDIT.md`.
Known risks: GH-21 production node bootstrap, ZKP recipient binding, legacy
wallets, Rust dev-tooling warnings, and independent stacked reviews remain.
Next recommended action: publish/verify PR #19, retain the PR #18 review gate,
then audit GH-20/GH-21 without collapsing the ordered stack. The independent PR #9
approval remains required; do not bypass branch protection.
Do not touch: `/Users/gio/Desktop/repos/TrueRepublic` dirty legacy checkout.
