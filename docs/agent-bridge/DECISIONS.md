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

## 2026-07-11 - Canonical reward issuance

- `x/bank` `upnyx` supply is the only release-decay and cap source of truth;
  `pod:total-release` is retired from consensus logic.
- `token.IssuanceService` is the governance module's only reward/slash supply
  boundary. Minting is clipped to remaining capacity in a cached context.
- Validator rewards have deterministic priority over domain interest when both
  compete for final cap capacity; allocation within each category follows
  deterministic store-key order.
- Domain interest uses payouts since the prior interval snapshot, not the same
  cumulative historical payouts repeatedly.
