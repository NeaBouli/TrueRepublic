# TrueRepublic — Token Supply and Ledger Audit
> Scope: `treasury/keeper`, `x/truedemocracy`, `x/dex`, app wiring, genesis, and maintained-client denomination handling  ·  Date: 2026-07-11  ·  Result: 4 FAIL / 2 WARN / 6 PASS

## Summary

The recovery branches now define one six-decimal `upnyx` base denomination,
validate the 21,000,000 PNYX bank-genesis cap, and back governance treasury and
validator stake claims with bank escrow. Reward issuance, DEX custody/burns,
custom genesis reconciliation, and runtime invariants remain blocking. This
code must not handle real funds or be treated as a production token economy
until those findings are resolved.

> Remediation update: GH-11 implements canonical denomination metadata and a
> bank-genesis cap check; its local and GitHub gates pass. GH-14 now implements
> authenticated bank escrow, atomic treasury/stake settlement, signer-safe
> CosmWasm bindings, slash burns, and parity tests on its stacked branch.
> Anonymous rating rewards are deferred because their
> current proof/signature payloads do not bind a safe recipient. Issuance, DEX,
> custom genesis, runtime invariants, and final review remain open.

## Findings by domain

### Denomination and bank-genesis supply cap — PASS

- **[PASS] Chain, maintained client, contracts, and operator docs use one 21M denomination model** — `token/denom.go`, `client-web/src/config/chains.ts`, `docs/node-operators/configuration/genesis-params.md`
  - What: `upnyx` is the canonical base denom, `pnyx`/PNYX is the six-decimal display unit, and the cap is 21,000,000,000,000 base units.
  - Evidence: GH-11 boundary tests, full local polyglot gates, and all stacked PR #15 checks pass.
  - Fix: Preserve the centralized token package and reject reintroduction of display-denom balances.

- **[PASS] Bank genesis rejects supply above 21,000,000 PNYX** — `token/genesis.go`, `app.go`, `token/genesis_test.go`
  - What: Pre-init validation sums canonical bank supply, rejects legacy display-denom balances, and enforces the 21,000,000,000,000 `upnyx` boundary.
  - Evidence: Cap-minus-one, exact-cap, cap-plus-one, balance-fallback, and metadata regression tests pass locally and on stacked PR #15.
  - Fix: GH-13 must reuse the same canonical cap for every runtime mint; GH-12 must add the runtime invariant.

### Governance, treasury, and staking accounting — PARTIAL / FAIL

- **[PASS] Authenticated treasury and validator claims are backed by exact module escrow** — `x/truedemocracy/escrow.go`, `x/truedemocracy/msg_server.go`, `x/truedemocracy/treasury_bridge.go`, `x/truedemocracy/escrow_test.go`
  - What: Domain creation, proposal fees, validator registration, deposits, withdrawals, stake settlement, and signer claims now use exact `upnyx` bank transfers and cached atomic state transitions.
  - Evidence: Tests reject zero-balance declarations, spoofed identities, duplicate credits/pubkeys, dust withdrawals, missing CosmWasm signers, and injected transfer/burn failures; parity is checked after creation, fees, stake, withdrawal, payouts, and slashing.
  - Fix: Preserve `CacheContext` boundaries and expand the same custody model to rewards, DEX, and genesis.

- **[🔴 BLOCKING] Rewards create internal claims but neither mint nor cap bank coins** — `x/truedemocracy/validator.go:266`, `x/truedemocracy/validator.go:281`, `x/truedemocracy/validator.go:319`, `x/truedemocracy/validator.go:330`
  - What: Validator rewards increase internal stake and domain interest increases internal treasury. No `MintCoins` or bank transfer backs either credit; domain interest is not added to `pod:total-release` at all.
  - Path: After reward intervals, a domain can show a larger treasury than its module escrow. A later bank-backed withdrawal consumes existing depositor funds or fails because the module account lacks the claimed balance. Domain interest can grow without moving the release counter toward the cap.
  - Fix: Route issuance through one minter-controlled module, check remaining cap, mint exact base units, and either escrow them for internal claims or pay them directly. Track all issuance and burns from bank supply, not a parallel counter.

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

- **[🔴 BLOCKING] Genesis accepts unbacked module ledgers and resets release tracking** — `x/truedemocracy/module.go:73`, `x/truedemocracy/module.go:106`, `x/truedemocracy/module.go:148`, `x/dex/module.go:50`, `x/dex/module.go:83`
  - What: Both custom modules return success from `ValidateGenesis`, load declared treasuries/stakes/pools without reconciling bank balances, and initialize total release to zero regardless of genesis bank supply.
  - Path: A genesis file can declare bank supply above 21M plus unrelated internal domain, stake, and DEX balances. Nodes accept it, then calculate maximum early-stage rewards from a zero release counter.
  - Fix: Validate denomination, non-negative values, uniqueness, total bank supply, module escrow parity, pool reserves, LP share totals, and the 21M cap before chain start. Initialize release state from the canonical bank supply or remove the duplicate counter.

### Precision and deterministic math — PASS

- **[PASS] Tokenomics and AMM calculations use deterministic integer/legacy-decimal math** — `treasury/keeper/rewards.go:29`, `x/dex/keeper.go:79`
  - What: Consensus calculations avoid floating-point arithmetic and truncate deterministically.
  - Path: Identical integer inputs produce identical results on all validators.
  - Fix: Preserve this property when introducing base-unit constants and bank accounting.

- **[PASS] Release decay clamps rewards at and above its configured maximum** — `treasury/keeper/rewards.go:75`
  - What: Negative decay is clamped to zero.
  - Path: A release value above the configured denominator does not produce a negative reward.
  - Fix: Retain the clamp, but feed it canonical bank supply.

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

1. Back all reward issuance with cap-checked bank minting and canonical supply.
2. Rebuild DEX custody, bank settlement, burn, and per-provider LP ownership.
3. Validate custom genesis and reconcile all internal ledgers to bank/module balances.

Implementation tickets: #11 (denom/cap), #14 (treasury/stake escrow), #13
(rewards), #10 (DEX custody), and #12 (genesis/invariants).

### 🟠 HIGH

1. Restrict asset registry and trading-status changes to chain authority.

### 🟡 MEDIUM

1. Add runtime supply, escrow, reserve, and LP conservation invariants.
2. Downgrade detailed public implementation claims until those invariants pass.

### 🟢 LOW

None identified in this audit slice.
