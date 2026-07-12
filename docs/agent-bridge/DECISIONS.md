# Decisions

## 2026-07-11 - Recovery baseline

- Current GitHub `main`, not the divergent old local checkout, is the source baseline.
- The old local checkout is preserved and selectively reconciled; it is not reset.
- Recovery work happens on `fix/GH-4-recovery-foundation` in an isolated worktree.

## 2026-07-11 - PNYX maximum supply

- Maximum supply is **21,000,000 whole PNYX**. GH-11 enforces the
  `21,000,000,000,000 upnyx` bank-genesis boundary and GH-13 enforces the same
  canonical bank-supply cap for recovered runtime reward issuance.
- GH-10 routes DEX burns through the canonical issuance service. The
  independent runtime supply/custody invariant remains pending in GH-12; no
  custom module creates a second supply source.

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

## 2026-07-12 - DEX custody and LP ownership

- The `dex` module account is the sole bank custodian for every pool reserve.
- Public create/add/remove/swap transitions use a cached all-or-nothing bank
  settlement and must pass reserve and LP conservation before commit.
- LP ownership is indexed by pool and authenticated provider; global pool
  shares are not transferable withdrawal authority.
- PNYX output burns reduce both pool reserves and canonical bank supply through
  `token.IssuanceService`.
- Asset registry/status mutation requires the configured chain authority.

## 2026-07-12 - Safe consensus genesis

- Production defaults contain no validator private secret or fixed validator
  identity.
- When custom PoD genesis is empty, InitChain accepts only real positive-power
  Ed25519 validators supplied by CometBFT and creates exact bank-backed minimum
  stake for those public keys within the 21M cap.
- Explicit custom treasury/stake/DEX claims must equal the complete module bank
  balances before any custom state mutation.
- Supply, governance escrow, DEX reserves, and provider LP totals are registered
  `x/crisis` routes checked every block.
- The legacy `x/staking` gentx initialization script is not a valid PoD launch
  path and remains blocked under GH-21.

## 2026-07-12 - ZKP circuit and nullifier trust boundary

- Anonymous vote proofs bind a versioned length-prefixed chain ID, domain,
  issue, suggestion, and exact rating signal.
- The one-vote nullifier excludes rating but includes chain and proposal
  identity, so changing a rating cannot create another voting scope.
- Consensus never performs randomized Groth16 setup. Genesis is the ceremony
  trust anchor and pins the expected circuit ID, VK SHA-256, BN254 curve,
  four-public-input shape, and canonical serialized bytes.
- Genesis recomputes identity Merkle roots and exports the exact active
  nullifier records. Historical ratings are not used to resurrect nullifiers
  intentionally cleared by a Big Purge.
- Mock proof generation is not a degraded transaction mode. Both web clients
  fail closed until a compatible real prover is shipped and reviewed.
- Anonymous rewards remain deferred until the proof or a separate claim binds
  a safe recipient without destroying vote privacy.
