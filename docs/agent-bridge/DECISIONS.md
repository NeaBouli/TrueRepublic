# Decisions

## 2026-07-11 - Recovery baseline

- Current GitHub `main`, not the divergent old local checkout, is the source baseline.
- The old local checkout is preserved and selectively reconciled; it is not reset.
- Recovery work happens on `fix/GH-4-recovery-foundation` in an isolated worktree.

## 2026-07-11 - PNYX maximum supply

- Intended maximum supply is **21,000,000 whole PNYX**. Enforcement is pending
  the ordered GH-7 remediation and end-to-end verification.
- Public status defines six decimals; the intended base-unit cap is
  `21,000,000,000,000 upnyx`, pending the same enforcement verification.

## 2026-07-11 - Status publication

- Public project status is evidence-based: no feature, test count, security
  state, or release completeness claim may exceed verified code and CI results.

## 2026-07-11 - Validator slash custody

- Slashed validator PNYX is burned from the `truedemocracy` module escrow.
- It must not be credited to an admin-withdrawable domain treasury because the
  whitepaper removes the penalty from circulation and the treasury path would
  allow validator/admin collusion to recover it.
