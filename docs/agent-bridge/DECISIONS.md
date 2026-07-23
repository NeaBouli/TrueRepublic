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
- `truerepublicd init` is the only valid PoD bootstrap boundary. The GH-26
  wrapper delegates exclusively to it and cannot create staking gentxs,
  keyring mnemonics, extra genesis accounts, or additional token supply.

## 2026-07-12 - Persistent PoD node bootstrap

- Superseded by the 2026-07-23 GH-56 decision below. The generated CometBFT
  Ed25519 key remains the consensus identity, but it is no longer permitted to
  define the operator authority.
- Bootstrap stake is created only as exact cap-checked `x/bank` module backing
  for the matching custom validator. Conflicting existing consensus sets are
  rejected rather than silently replaced.
- Standard Cosmos server lifecycle, persistent database/home, signal shutdown,
  restart, and export are the supported node path. The old MemDB/`select {}` and
  `x/staking` gentx paths are retired.
- Single-node success does not prove multi-node, IBC upgrade, relayer, backup,
  or restore readiness; those require separate operations evidence.

## 2026-07-23 - Independent validator operator authority and key rotation

- Every genesis validator binds an explicit account operator independently
  from its CometBFT consensus key. Same-validator, cross-validator, active-key,
  revoked-key, and reserved module-account collisions fail closed.
- `truerepublicd init --bootstrap-operator` records only the public operator
  account and creates no private key, mnemonic, gentx, or liquid allocation.
- An authenticated rotation is conditional on an active, non-jailed,
  positive-power validator, binds the signed request to the expected old key,
  preserves claims, schedules the new key for H+2, and permanently revokes the
  old key. Reuse fails across export/import.
- Pre-GH-56 homes do not gain this separation through a binary replacement.
  The supported prelaunch transition is a reviewed fresh genesis; no in-place
  authority migration is claimed.

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

## 2026-07-12 - Documentation authority and CI runtimes

- `docs/status.json` is the machine source for the current version, recovery
  test totals/module split, technology versions, 21M cap, and feature limits.
- README, CLAUDE, landing page, and real wiki Home/current/testing pages must
  contain the current version and total; CI also proves suite/module sums and
  decimal-to-base-unit cap arithmetic.
- Historical milestone counts may remain only when explicitly labeled
  historical, never as current security or production evidence.
- GitHub workflows use current official Action majors with `contents: read` and
  non-persisted checkout credentials. Project Node/Go versions remain explicit
  and separate from the Actions embedded runtime.
- Feature branches run through pull requests or manual dispatch; routine push
  automation is main-only to avoid duplicate evidence.

## 2026-07-14 - Multi-validator recovery boundary

- Shared validator genesis contains public CometBFT identities only. Every
  private validator key remains in its independently generated node home.
- The single-node `truerepublicd init` command continues to refuse replacing an
  existing consensus set. Multi-validator assembly reuses an internal audited
  public-identity function without weakening that operator boundary.
- Loopback address-book relaxation, duplicate-IP permission, and disabled
  pprof apply only to temporary localhost harness configuration. Production
  CometBFT defaults remain strict.
- Four-validator failure/restart/catch-up and app-hash agreement close one
  Phase 1 checklist item, not multi-node or public-network readiness.

## 2026-07-15 - Codex subagent role split

- The primary Codex agent remains responsible for architecture, security/risk
  decisions, final verification, Bridge updates, GitHub issues, PRs, merges, and
  public status claims.
- Project-scoped `.codex` configuration defines `spark_worker` as a narrow
  `gpt-5.3-codex-spark` worker for small bounded patches, file search, and
  focused checks.
- Subagent recursion stays capped at one level (`agents.max_depth = 1`) and
  concurrency at six open threads (`agents.max_threads = 6`) to keep token use
  predictable while allowing targeted parallel help.
