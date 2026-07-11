# Decisions

## 2026-07-11 - Recovery baseline

- Current GitHub `main`, not the divergent old local checkout, is the source baseline.
- The old local checkout is preserved and selectively reconciled; it is not reset.
- Recovery work happens on `fix/GH-4-recovery-foundation` in an isolated worktree.

## 2026-07-11 - PNYX maximum supply

- Maximum supply is fixed at **21,000,000 whole PNYX**.
- Public status defines six decimals; the expected base-unit cap is
  `21,000,000,000,000 upnyx`, pending end-to-end code verification.

## 2026-07-11 - Status publication

- Public project status is evidence-based: no feature, test count, security
  state, or release completeness claim may exceed verified code and CI results.
