# TrueRepublic Architecture

## System Layers

```
┌─────────────────────────────────────────────────┐
│  Client Layer                                    │
│  Web Wallet (React 18 + Keplr + CosmJS)         │
│  Mobile Wallet (Expo / React Native + CosmJS)   │
├─────────────────────────────────────────────────┤
│  Smart Contract Layer (CosmWasm)                 │
│  governance.rs  │  treasury.rs                   │
├─────────────────────────────────────────────────┤
│  Application Layer (Cosmos SDK v0.50.13)         │
│  truedemocracy  │  dex  │  treasury              │
├─────────────────────────────────────────────────┤
│  Consensus Layer (CometBFT v0.38.21)             │
│  Proof of Domain (PoD) validator selection       │
└─────────────────────────────────────────────────┘
```

## Module Architecture

### truedemocracy module (`x/truedemocracy/`)

The core governance module implementing the whitepaper specification.

| Component | File | Description |
|-----------|------|-------------|
| Keeper | `keeper.go` | Domain CRUD, proposal submission, anonymous ratings |
| Validator | `validator.go` | PoD registration, membership enforcement, staking rewards (eq.5) |
| Slashing | `slashing.go` | Double-sign (5%), downtime (1%), jailing |
| Anonymity | `anonymity.go` | Permission register, domain key pairs (WP S4) |
| Stones | `stones.go` | VoteToEarn rewards, list sorting (WP S3.1) |
| Lifecycle | `lifecycle.go` | Green/yellow/red zones, auto-delete, fast-delete (WP S3.1.2) |
| Governance | `governance.go` | Admin election, member exclusion, inactivity cleanup (WP S3.6) |
| CLI | `cli.go` | 13 tx commands + 4 query commands |
| Querier | `querier.go` | ABCI query routes |
| Messages | `msgs.go` | 13 SDK message types |
| Types | `types.go` | Domain, Validator, Issue, Suggestion, Rating, VoteCommitment |
| Module | `module.go` | SDK wiring, InitGenesis, EndBlock |

**EndBlock processing order:**
1. Distribute staking rewards (every 3600 blocks)
2. Distribute domain interest
3. Enforce domain membership (evict validators without domains)
4. Process suggestion lifecycles (zone transitions, auto-delete)
5. Process governance (admin election, inactivity cleanup)
6. Return validator set updates

### dex module (`x/dex/`)

Automated Market Maker with constant-product formula (x * y = k).

| Component | File | Description |
|-----------|------|-------------|
| Keeper | `keeper.go` | CreatePool, Swap, AddLiquidity, RemoveLiquidity |
| CLI | `cli.go` | 4 tx commands + 2 query commands |
| Messages | `msgs.go` | 4 SDK message types |
| Types | `types.go` | Pool type, fee constants |

**Fee structure:**
- Swap fee: 0.3% (SwapFeeBps = 30)
- PNYX burn: 1% on swaps to PNYX (BurnBps = 100, WP S5)

**Swap formula:**
```
output = (outReserve * input * 9970) / (inReserve * 10000 + input * 9970)
```

### treasury module (`treasury/keeper/`)

Tokenomics equations from the whitepaper.

| Equation | Function | Description |
|----------|----------|-------------|
| eq.1 | `CalcDomainCost(fee)` | Domain creation cost: `fee * CDom * CEarn` |
| eq.2 | `CalcReward(treasure)` | Treasury reward: `treasure / CEarn` |
| eq.3 | `CalcPutPrice(treasure, nUser)` | Post price: `min(reward * CPut, reward * nUser)` |
| eq.4 | `CalcDomainInterest(...)` | Domain interest: APY 25%, release-decay adjusted |
| eq.5 | `CalcNodeReward(...)` | Staking reward: APY 10%, release-decay adjusted |

**Constants:**
- `CDom = 2`, `CPut = 15`, `CEarn = 1000`
- `StakeMin = 100,000 PNYX`
- `SupplyMax = 21,000,000 PNYX`
- `ApyDom = 0.25` (25%), `ApyNode = 0.10` (10%)

**Release decay:** `(1 - totalReleased / SupplyMax)` — inflation decreases as supply approaches max.

## CosmWasm Contracts (`contracts/src/`)

| Contract | Description |
|----------|-------------|
| `governance.rs` | On-chain proposals with systemic consensing (-5 to +5), domain key pair validation |
| `treasury.rs` | Deposit/withdraw treasury operations with balance tracking |

Built with cosmwasm-std 3, compiled to WASM target.

## Chain Configuration

| Parameter | Value |
|-----------|-------|
| Chain ID | `truerepublic-1` |
| Bech32 prefix | `truerepublic` |
| Token denom | `pnyx` |
| P2P port | 26656 |
| RPC port | 26657 |
| LCD/REST port | 1317 |
| gRPC port | 9090 |
| Prometheus port | 26660 |

## Data Flow

```
User → Wallet (Keplr/CosmJS) → Sign Tx
  → RPC (port 26657) → CometBFT Mempool
  → Block Proposal → ABCI DeliverTx
  → Cosmos SDK Router → Module Handler
  → Keeper → KV Store (commit)
  → EndBlock → Validator Updates
```

## Test Coverage

| Module | Tests | Lines |
|--------|-------|-------|
| truedemocracy | 116 | 2,077 |
| dex | 24 | 423 |
| treasury | 31 | 205 |
| **Total** | **182** | **2,705** |
