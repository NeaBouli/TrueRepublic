# PR #15 Audit - Canonical PNYX Denomination and Genesis Cap

Updated: 2026-07-11 21:09 EEST

Scope: GitHub issue #11 and stacked PR #15.

## Result

PASS after remediation, for the GH-11 scope. The branch defines `upnyx` as the
base denomination, `pnyx` as the six-decimal display denomination, and enforces
the 21,000,000 PNYX maximum at bank-genesis initialization.

## Verified invariants

- `1 PNYX = 1,000,000 upnyx`.
- Maximum genesis supply is `21,000,000,000,000 upnyx`.
- Cap-minus-one and exact-cap genesis states pass; cap-plus-one fails.
- Positive legacy `pnyx` supply or balances fail genesis validation.
- Missing explicit bank supply is derived from balances before cap validation.
- Canonical metadata is idempotent, removes conflicting legacy PNYX metadata,
  and preserves unrelated asset metadata.
- The default 100,000 PNYX validator stake is represented as
  `100,000,000,000 upnyx`.
- Node and client gas prices retain their display-token economics after the
  base-denom migration (`0.001 PNYX = 1000 upnyx`,
  `0.025 PNYX = 25000 upnyx`).

## Findings fixed during final audit

- HIGH: the production validator tree used `100000upnyx`, only 0.1 PNYX.
- MEDIUM: Compose, node-init, environment examples, and operator docs still
  configured legacy `pnyx` gas fees.
- MEDIUM: the maintained client renamed its gas denom without scaling the
  amount, reducing the configured price by a factor of one million.
- LOW: legacy PNYX bank metadata could coexist with canonical metadata.

## Out of scope / next stack items

GH-11 does not claim complete runtime supply conservation. Cap-checked reward
issuance, bank-backed treasury/stake custody, DEX custody/burns, and full
genesis/runtime conservation invariants remain isolated in GH-13, GH-14,
GH-10, and GH-12 respectively. The repository remains non-production until
those stacked remediations are reviewed and merged.

## Merge condition

PR #15 must be rebased on the final merged PR #9 head before it is made ready.
The protected `main` branch requires an independent approval; no administrative
bypass is permitted.
