# TrueRepublic — Token Supply and Ledger Audit
> Scope: `treasury/keeper`, `x/truedemocracy`, `x/dex`, app wiring, genesis, and maintained-client denomination handling  ·  Date: 2026-07-11  ·  Result: 3 FAIL / 2 WARN / 7 PASS

## Summary

The recovery branches now define one six-decimal `upnyx` base denomination,
validate and enforce the 21,000,000 PNYX cap against canonical bank supply,
back governance/stake claims with bank escrow, and mint validator/domain
inflation through one capped issuance service. DEX custody/burn integration,
custom genesis reconciliation, and runtime invariants remain blocking. This
code must not handle real funds or be treated as a production token economy
until those findings are resolved.

> Remediation update: GH-11 and GH-14 pass their local/GitHub gates. GH-13 now
> routes validator/domain inflation and validator slash burns through canonical
> bank supply, clips aggregate minting at the cap, and commits both EndBlock
> inflation phases atomically. Anonymous rating rewards remain deferred because
> current proof/signature payloads do not bind a safe recipient. DEX integration,
> custom genesis, runtime invariants, and final review remain open.

## Findings by domain

### Denomination and bank-genesis supply cap — PASS

- **[PASS] Chain, maintained client, contracts, and operator docs use one 21M denomination model** — `token/denom.go`, `client-web/src/config/chains.ts`, `docs/node-operators/configuration/genesis-params.md`
  - What: `upnyx` is the canonical base denom, `pnyx`/PNYX is the six-decimal display unit, and the cap is 21,000,000,000,000 base units.
  - Evidence: GH-11 boundary tests, full local polyglot gates, and all stacked PR #15 checks pass.
  - Fix: Preserve the centralized token package and reject reintroduction of display-denom balances.

- **[PASS] Genesis and every recovery reward mint use canonical capped bank supply** — `token/denom.go`, `token/issuance.go`, `app.go`, `x/truedemocracy/validator.go`
  - What: Bank genesis rejects supply above 21,000,000 PNYX; the issuance service reads `x/bank` supply and mints at most the remaining base units into authorized module escrow.
  - Evidence: Tests cover below/at/above-cap supply, aggregate final-unit allocation, invalid supply, bank failure, exact burn/remint behavior, and end-block treasury/stake parity.
  - Fix: Route DEX burns through the same service in GH-10 and add the independent runtime invariant in GH-12.

### Governance, treasury, and staking accounting — PASS

- **[PASS] Authenticated treasury and validator claims are backed by exact module escrow** — `x/truedemocracy/escrow.go`, `x/truedemocracy/msg_server.go`, `x/truedemocracy/treasury_bridge.go`, `x/truedemocracy/escrow_test.go`
  - What: Domain creation, proposal fees, validator registration, deposits, withdrawals, stake settlement, and signer claims now use exact `upnyx` bank transfers and cached atomic state transitions.
  - Evidence: Tests reject zero-balance declarations, spoofed identities, duplicate credits/pubkeys, dust withdrawals, missing CosmWasm signers, and injected transfer/burn failures; parity is checked after creation, fees, stake, withdrawal, payouts, and slashing.
  - Fix: Preserve `CacheContext` boundaries and expand the same custody model to DEX and genesis.

- **[PASS] Validator/domain inflation is bank-minted and vote rewards settle from escrow** — `token/issuance.go`, `x/truedemocracy/validator.go`, `x/truedemocracy/module.go`, `x/truedemocracy/issuance_test.go`
  - What: Validator and active-domain rewards mint exact capped `upnyx` into module escrow before matching internal claims commit. Both inflation phases share an outer EndBlock cache; vote rewards transfer existing treasury escrow and unsafe anonymous recipients remain deferred.
  - Evidence: Tests cover final-cap aggregation, first/second mint rollback, canonical decay, interval-only payouts, vote-transfer neutrality, slash burns, and escrow parity.
  - Fix: Preserve the single issuance boundary; complete recipient-bound anonymous reward claims in GH-7.

### DEX custody and authorization — FAIL

- **[🔴 BLOCKING] DEX reserves, swaps, burns, and LP shares are not bank-backed** — `x/dex/keeper.go:14`, `x/dex/msg_server.go:117`, `x/dex/msg_server.go:136`, `x/dex/msg_server.go:157`, `x/dex/msg_server.go:174`, `x/dex/keeper.go:455`
  - What: The DEX keeper has no bank keeper or module account. Handlers mutate declared reserves and return values without taking input coins, paying output coins, burning bank PNYX, or recording per-provider LP ownership.
  - Path: Any signer can declare liquidity they do not own, swap without paying input, then call remove-liquidity with arbitrary shares up to global total shares. The pool state reports returned coins and burned PNYX, but no bank balances change.
  - Fix: Add DEX module custody, atomic bank transfers, real PNYX burn permission, and provider-indexed LP share balances. Every state transition must reconcile module balances, pool reserves, and total/provider shares.

- **[🟠 HIGH] Any signer can register assets or toggle trading status** — `x/dex/msg_server.go:192`, `x/dex/msg_server.go:214`, `x/dex/asset_registry.go:52`, `x/dex/asset_registry.go:116`
  - What: Registry messages require a signer but perform no authority check.
  - Path: An arbitrary account registers a spoofed denom/symbol or disables a legitimate market, affecting routing and pool availability chain-wide.
  - Fix: Require governance/module authority or an explicit allowlisted admin and test unauthorized callers.

### Genesis and invariants — FAIL

- **[🔴 BLOCKING] Genesis accepts unbacked custom module ledgers** — `x/truedemocracy/module.go:73`, `x/truedemocracy/module.go:106`, `x/dex/module.go:50`, `x/dex/module.go:83`
  - What: Both custom modules return success from `ValidateGenesis` and load declared treasuries, stakes, pools, and shares without reconciling corresponding module bank balances.
  - Path: Bank genesis can respect the 21M cap while unrelated internal claims add unbacked treasury/stake/DEX purchasing power that nodes still accept.
  - Fix: Validate denomination, non-negative values, uniqueness, module escrow parity, pool reserves, and LP share totals before chain start.

### Precision and deterministic math — PASS

- **[PASS] Tokenomics and AMM calculations use deterministic integer/legacy-decimal math** — `treasury/keeper/rewards.go:29`, `x/dex/keeper.go:79`
  - What: Consensus calculations avoid floating-point arithmetic and truncate deterministically.
  - Path: Identical integer inputs produce identical results on all validators.
  - Fix: Preserve this property when introducing base-unit constants and bank accounting.

- **[PASS] Release decay clamps rewards at and above its configured maximum** — `treasury/keeper/rewards.go:75`
  - What: Negative decay is clamped to zero.
  - Path: A release value above the configured denominator does not produce a negative reward.
  - Fix: Retain the clamp and canonical bank-supply input.

- **[PASS] Maintained-client PNYX formatting no longer loses integer precision** — `client-web/src/utils/format.ts:4`
  - What: Display/base-unit conversion uses validated strings and `BigInt`.
  - Path: Values above JavaScript's safe integer limit round-trip without `Number` precision loss.
  - Fix: Reuse the same helper everywhere after the canonical denom/decimal decision.

### Recovery guardrails — WARN

- **[🟡 MEDIUM] No crisis invariant detects ledger divergence at runtime** — `x/truedemocracy/module.go:97`, `x/dex/module.go:74`
  - What: Both modules register no invariants, so module escrow/reserve/claim divergence is silent.
  - Path: A faulty transition commits mismatched balances and subsequent blocks continue until a withdrawal or swap exposes the deficit.
  - Fix: Add deterministic invariants for bank supply cap, treasury escrow parity, DEX reserve parity, and LP share conservation; expose them to continuous tests even if `x/crisis` is later removed.

- **[🟡 MEDIUM] Public feature descriptions still imply functional token economics** — `README.md:196`, `README.md:197`, `docs/ARCHITECTURE.md:84`
  - What: The recovery warning is visible, but detailed tables still mark tokenomics and DEX implementation with check marks.
  - Path: A reader can interpret implemented arithmetic/state simulation as bank-backed functionality despite the recovery warning.
  - Fix: Mark these rows as simulation/recovery-blocked until the bank-backed invariants pass.

## Priority matrix

### 🔴 BLOCKING

1. Rebuild DEX custody, bank settlement, burn, and per-provider LP ownership.
2. Validate custom genesis and reconcile all internal ledgers to bank/module balances.

Implemented on stacked recovery branches: #11 (denom/cap), #14
(treasury/stake escrow), and #13 (rewards). Remaining implementation tickets:
#10 (DEX custody) and #12 (genesis/invariants).

### 🟠 HIGH

1. Restrict asset registry and trading-status changes to chain authority.

### 🟡 MEDIUM

1. Add runtime supply, escrow, reserve, and LP conservation invariants.
2. Downgrade detailed public implementation claims until those invariants pass.

### 🟢 LOW

None identified in this audit slice.
