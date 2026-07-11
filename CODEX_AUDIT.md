# TrueRepublic — Token Supply and Ledger Audit
> Scope: `treasury/keeper`, `x/truedemocracy`, `x/dex`, app wiring, genesis, and maintained-client denomination handling  ·  Date: 2026-07-11  ·  Result: 7 FAIL / 2 WARN / 3 PASS

## Summary

The intended maximum supply is 21,000,000 whole PNYX, but the current chain
does not enforce that invariant against the bank supply. Governance, staking,
and DEX state use parallel internal ledgers that can be credited from declared
message amounts without moving real bank coins. Rewards and DEX burns likewise
change only internal state, so they neither mint nor burn bank supply and can
create unbacked withdrawal claims. This code must not handle real funds or be
treated as a production token economy until the blocking ledger and genesis
findings below are resolved.

> Remediation update: GH-11 now implements the canonical `upnyx` metadata,
> six-decimal base-unit conversion, and bank-genesis cap check on its stacked
> branch. The audit result remains open until that branch passes GitHub review;
> the custody, issuance, DEX, and custom-genesis findings are unaffected.

## Findings by domain

### Denomination and supply cap — FAIL

- **[🔴 BLOCKING] Six-decimal UI and zero-decimal chain disagree on the 21M cap** — `treasury/keeper/rewards.go:10`, `treasury/keeper/rewards.go:11`, `client-web/src/config/chains.ts:10`, `client-web/src/config/chains.ts:11`, `docs/node-operators/configuration/genesis-params.md:9`
  - What: Consensus code uses `pnyx` amounts and `SupplyMax = 21_000_000`, while the maintained client treats the same minimal denom as six-decimal PNYX and some balance paths search for `upnyx`. Operator documentation still declares zero decimals.
  - Path: A chain balance of `21_000_000pnyx` is the full cap in reward math, but the maintained client formats it as `21.000000 PNYX`; other screens fail to find it because they search only for `upnyx`. Conversely, a user entering 21M PNYX produces 21,000,000,000,000 minimal units, one million times the reward cap.
  - Fix: Make one consensus decision and centralize it. For six decimals, use `upnyx` as the base denom, `pnyx` as display metadata, and cap bank supply at `21_000_000_000_000upnyx`; migrate every module, genesis field, client, contract binding, and document together.

- **[🔴 BLOCKING] SupplyMax affects decay only; bank supply is never capped** — `treasury/keeper/rewards.go:72`, `x/truedemocracy/validator.go:256`, `x/truedemocracy/module.go:144`
  - What: `SupplyMax` is only a denominator in release-decay math. There is no invariant comparing total bank supply plus outstanding internal claims against the cap.
  - Path: Genesis can allocate arbitrary bank balances, while internal validator/domain rewards continue until the separately stored `pod:total-release` reaches the constant. The real bank supply can already exceed 21M without stopping the reward path.
  - Fix: Define a single supply source of truth from `x/bank`, validate genesis supply at or below the base-unit cap, and enforce remaining mint capacity atomically before every reward mint.

### Governance, treasury, and staking accounting — FAIL

- **[🔴 BLOCKING] User-declared amounts create unbacked treasury and stake balances** — `x/truedemocracy/msg_server.go:262`, `x/truedemocracy/keeper.go:46`, `x/truedemocracy/msg_server.go:276`, `x/truedemocracy/keeper.go:109`, `x/truedemocracy/msg_server.go:294`, `x/truedemocracy/validator.go:25`
  - What: Domain creation, proposal fees, and validator registration copy amounts from signed messages into internal state without sending coins from the signer to a module account.
  - Path: An account with zero PNYX can submit a signed create-domain or register-validator message declaring a large amount. The keeper stores that amount as treasury or stake and derives validator power from it even though no bank balance was debited.
  - Fix: Require an authenticated account address, transfer exact coins through `BankKeeper` into dedicated module escrow before state mutation, and reject any amount not actually received.

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

1. Decide and migrate the canonical six-decimal base denom and 21M base-unit cap.
2. Replace declared governance/staking values with authenticated bank escrow.
3. Back all reward issuance with cap-checked bank minting and canonical supply.
4. Rebuild DEX custody, bank settlement, burn, and per-provider LP ownership.
5. Validate genesis and reconcile all custom ledgers to bank/module balances.

Implementation tickets: #11 (denom/cap), #14 (treasury/stake escrow), #13
(rewards), #10 (DEX custody), and #12 (genesis/invariants).

### 🟠 HIGH

1. Restrict asset registry and trading-status changes to chain authority.

### 🟡 MEDIUM

1. Add runtime supply, escrow, reserve, and LP conservation invariants.
2. Downgrade detailed public implementation claims until those invariants pass.

### 🟢 LOW

None identified in this audit slice.
