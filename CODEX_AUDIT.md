# TrueRepublic — Token Supply and Ledger Audit
> Scope: `treasury/keeper`, `x/truedemocracy`, `x/dex`, app wiring, genesis, and maintained-client denomination handling  ·  Date: 2026-07-12  ·  Result: 1 FAIL / 2 WARN / 9 PASS

## Summary

The recovery branches now define one six-decimal `upnyx` base denomination,
validate and enforce the 21,000,000 PNYX cap against canonical bank supply,
back governance/stake claims with bank escrow, mint validator/domain
inflation through one capped issuance service, and settle DEX reserves, LP
ownership, swaps, and burns against canonical bank state. Custom genesis
reconciliation and registered runtime invariants remain blocking. This
code must not handle real funds or be treated as a production token economy
until those findings are resolved.

> Remediation update: GH-11 and GH-14 pass their local/GitHub gates. GH-13 now
> routes validator/domain inflation and validator slash burns through canonical
> bank supply, clips aggregate minting at the cap, and commits both EndBlock
> inflation phases atomically. Anonymous rating rewards remain deferred because
> current proof/signature payloads do not bind a safe recipient. GH-10 now
> provides bank-backed DEX custody, provider-owned LP shares, authority checks,
> and canonical burns. Custom genesis, runtime invariants, and final stacked
> review remain open.

## Findings by domain

### Denomination and bank-genesis supply cap — PASS

- **[PASS] Chain, maintained client, contracts, and operator docs use one 21M denomination model** — `token/denom.go`, `client-web/src/config/chains.ts`, `docs/node-operators/configuration/genesis-params.md`
  - What: `upnyx` is the canonical base denom, `pnyx`/PNYX is the six-decimal display unit, and the cap is 21,000,000,000,000 base units.
  - Evidence: GH-11 boundary tests, full local polyglot gates, and all stacked PR #15 checks pass.
  - Fix: Preserve the centralized token package and reject reintroduction of display-denom balances.

- **[PASS] Genesis and every recovery reward mint use canonical capped bank supply** — `token/denom.go`, `token/issuance.go`, `app.go`, `x/truedemocracy/validator.go`
  - What: Bank genesis rejects supply above 21,000,000 PNYX; the issuance service reads `x/bank` supply and mints at most the remaining base units into authorized module escrow.
  - Evidence: Tests cover below/at/above-cap supply, aggregate final-unit allocation, invalid supply, bank failure, exact burn/remint behavior, and end-block treasury/stake parity.
  - Fix: Preserve GH-10's canonical DEX burn path and add the independent runtime invariant in GH-12.

### Governance, treasury, and staking accounting — PASS

- **[PASS] Authenticated treasury and validator claims are backed by exact module escrow** — `x/truedemocracy/escrow.go`, `x/truedemocracy/msg_server.go`, `x/truedemocracy/treasury_bridge.go`, `x/truedemocracy/escrow_test.go`
  - What: Domain creation, proposal fees, validator registration, deposits, withdrawals, stake settlement, and signer claims now use exact `upnyx` bank transfers and cached atomic state transitions.
  - Evidence: Tests reject zero-balance declarations, spoofed identities, duplicate credits/pubkeys, dust withdrawals, missing CosmWasm signers, and injected transfer/burn failures; parity is checked after creation, fees, stake, withdrawal, payouts, and slashing.
  - Fix: Preserve `CacheContext` boundaries and expand the same custody model to DEX and genesis.

- **[PASS] Validator/domain inflation is bank-minted and vote rewards settle from escrow** — `token/issuance.go`, `x/truedemocracy/validator.go`, `x/truedemocracy/module.go`, `x/truedemocracy/issuance_test.go`
  - What: Validator and active-domain rewards mint exact capped `upnyx` into module escrow before matching internal claims commit. Both inflation phases share an outer EndBlock cache; vote rewards transfer existing treasury escrow and unsafe anonymous recipients remain deferred.
  - Evidence: Tests cover final-cap aggregation, first/second mint rollback, canonical decay, interval-only payouts, vote-transfer neutrality, slash burns, and escrow parity.
  - Fix: Preserve the single issuance boundary; complete recipient-bound anonymous reward claims in GH-7.

### DEX custody and authorization — PASS

- **[PASS] DEX reserves, swaps, burns, and LP shares are bank-backed** — `x/dex/custody.go`, `x/dex/msg_server.go`, `app.go`
  - What: Exact module-bank custody backs every pool reserve; cached public transitions move real inputs/outputs, assign shares by provider, and burn PNYX through the canonical issuance service.
  - Evidence: Tests cover direct/two-hop settlement, constant-product monotonicity, supply burns, foreign-share withdrawal rejection, denom-prefix key isolation, reserve/share divergence, slippage, and injected transfer/burn rollback.
  - Fix: Preserve the custody wrapper as the only public transition path and register its parity checks as runtime invariants in GH-12.

- **[PASS] Asset registry mutation requires chain authority** — `x/dex/custody.go`, `x/dex/msg_server.go`
  - What: Register and status messages compare the authenticated signer with the configured chain authority before state mutation.
  - Evidence: Unauthorized registration/status changes fail; the configured authority path succeeds.
  - Fix: Keep authority wiring explicit and exercise proposal execution when the governance runtime is completed.

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

1. Validate custom genesis and reconcile all internal ledgers to bank/module balances.

Implemented on stacked recovery branches: #11 (denom/cap), #14
(treasury/stake escrow), #13 (rewards), and #10 (DEX custody). Remaining
implementation ticket: #12 (genesis/invariants).

### 🟠 HIGH

None identified in this audit slice.

### 🟡 MEDIUM

1. Add runtime supply, escrow, reserve, and LP conservation invariants.
2. Downgrade detailed public implementation claims until those invariants pass.

### 🟢 LOW

None identified in this audit slice.
