# TrueRepublic Architecture

**Version:** v0.3.0

## System Overview

TrueRepublic is a Cosmos SDK blockchain with:
- **Zero-Knowledge Proof** anonymous voting
- **CosmWasm** smart contracts
- **IBC** cross-chain transfers
- **Multi-Asset DEX** with AMM

```
┌─────────────────────────────────────────────────────────────┐
│                     TrueRepublic Chain                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌───────────────┐  ┌──────────────┐  ┌────────────────┐  │
│  │ x/truedemocracy│  │    x/dex     │  │  x/treasury    │  │
│  │               │  │              │  │                │  │
│  │ • Domains     │  │ • Pools      │  │ • Tokenomics   │  │
│  │ • ZKP Voting  │  │ • Multi-Asset│  │ • Equations    │  │
│  │ • Governance  │  │ • AMM Swaps  │  │ • Rewards      │  │
│  └───────────────┘  └──────────────┘  └────────────────┘  │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              CosmWasm Integration                    │  │
│  │  • Custom Queries (7 types)                          │  │
│  │  • Custom Messages (5 types)                         │  │
│  │  • Domain↔Bank Bridge                                │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              IBC Transfer Module                     │  │
│  │  • Cross-chain PNYX                                  │  │
│  │  • Multi-asset support                               │  │
│  │  • Relayer compatibility                             │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
        │                    │                    │
        ▼                    ▼                    ▼
    Cosmos Hub           Osmosis            Other Chains
```

---

## Module Architecture

### x/truedemocracy

**Purpose:** Domain-based governance with ZKP anonymity

**Key Components:**
- **Domains:** Organizational units with treasuries
- **Issues:** Topics for collective decision-making
- **Suggestions:** Proposals within issues
- **ZKP Voting:** Anonymous rating (1-10) with Groth16 proofs
- **Stones:** Weighted votes (Green/Yellow/Red)
- **Elections:** Person election system

**State Storage:**
```
domains/{domain_id}                      → Domain
members/{domain_id}/{address}            → Member
issues/{domain_id}/{issue_id}            → Issue
suggestions/{domain_id}/{suggestion_id}  → Suggestion
ratings/{suggestion_id}/{nullifier}      → bool
merkle_roots/{domain_id}                 → MerkleRoot
verification_key                         → VerificationKey
```

**ZKP Circuit:**
- **Inputs:** Identity commitment, Merkle proof, rating
- **Outputs:** Nullifier (prevents double-voting)
- **Proof System:** Groth16 on BN254 curve
- **Library:** gnark

**EndBlock Processing Order:**
1. Distribute staking rewards (every 3600 blocks)
2. Distribute domain interest (eq.4)
3. Enforce domain membership (evict validators without domains)
4. Process suggestion lifecycles (zone transitions, auto-delete)
5. Process governance (admin election, inactivity cleanup)
6. Execute Big Purge if scheduled
7. Return validator set updates

---

### x/dex

**Purpose:** Multi-asset decentralized exchange

**Key Components:**
- **Pools:** Constant product AMM (x*y=k)
- **Asset Registry:** Whitelisted IBC assets
- **Multi-Hop Routing:** Automatic path finding
- **Liquidity Provision:** LP shares with fees
- **Analytics:** Volume, APY, slippage tracking

**State Storage:**
```
pools/{pool_id}                    → Pool
lp_shares/{pool_id}/{provider}     → Uint128
registered_assets/{ibc_denom}      → RegisteredAsset
asset_symbols/{symbol}             → ibc_denom
pool_stats/{pool_id}               → PoolStatistics
```

**AMM Formula:**
```
Output = (InputAmount * 9970 * ReserveOut) / (ReserveIn * 10000 + InputAmount * 9970)
Fee: 0.3% (30 bps)
PNYX Burn: 1% on PNYX output (100 bps)
```

**Multi-Hop:**
- BFS algorithm for shortest path
- Max 5 hops configurable
- Atomic execution (all-or-nothing)
- Total slippage protection

---

### treasury/keeper

**Purpose:** Economic equations and tokenomics

**Equations:**

| # | Function | Description |
|---|----------|-------------|
| 1 | `CalcDomainCost(fee)` | Domain creation cost: `fee * CDom * CEarn` |
| 2 | `CalcReward(treasure)` | VoteToEarn reward: `treasure / CEarn` |
| 3 | `CalcPutPrice(treasure, nUser)` | Suggestion cost: `min(reward * CPut, reward * nUser)` |
| 4 | `CalcDomainInterest(...)` | Domain treasury interest: 25% APY, release-decay adjusted |
| 5 | `CalcNodeReward(...)` | Validator staking reward: 10% APY, release-decay adjusted |

**Constants:**
- `CDom = 2`, `CPut = 15`, `CEarn = 1000`
- `StakeMin = 100,000 PNYX`
- `SupplyMax = 21,000,000 PNYX`
- `ApyDom = 0.25` (25%), `ApyNode = 0.10` (10%)
- Release decay: `(1 - totalReleased / SupplyMax)`

---

## CosmWasm Integration

### Architecture

```
┌────────────────────────────────────────────────┐
│           Smart Contract (Wasm)                │
│  • Rust code compiled to WebAssembly          │
│  • Sandboxed execution environment            │
└────────────────────────────────────────────────┘
                      │
                      │ Custom Bindings
                      ▼
┌────────────────────────────────────────────────┐
│         TrueRepublic Custom Bindings           │
│                                                │
│  Queries (7):                                  │
│  • Domain, DomainMembers, Issue, Suggestion    │
│  • PurgeSchedule, Nullifier, DomainTreasury    │
│                                                │
│  Messages (5):                                 │
│  • PlaceStoneOnIssue, PlaceStoneOnSuggestion   │
│  • CastElectionVote                            │
│  • DepositToDomain, WithdrawFromDomain         │
└────────────────────────────────────────────────┘
                      │
                      ▼
┌────────────────────────────────────────────────┐
│         Native Modules (Go)                    │
│  • x/truedemocracy                             │
│  • x/dex                                       │
│  • x/bank (via bridge)                         │
└────────────────────────────────────────────────┘
```

### Contract Workspace (7 crates)

| Crate | Path | Purpose |
|-------|------|---------|
| truerepublic-contracts | `core/` | Governance + treasury contracts |
| truerepublic-bindings | `packages/bindings/` | Shared query/msg types |
| truerepublic-testing-utils | `packages/testing-utils/` | Mock querier, AMM pool, fixtures |
| governance-dao | `examples/governance-dao/` | DAO proposal lifecycle |
| dex-bot | `examples/dex-bot/` | Limit orders, arbitrage detection |
| zkp-aggregator | `examples/zkp-aggregator/` | Anonymous vote aggregation |
| token-vesting | `examples/token-vesting/` | Linear vesting with cliff |

### Domain↔Bank Bridge

**Dual Accounting System:**
- Contracts use `Domain.Treasury` (internal ledger)
- Users use `x/bank` (Cosmos native)
- Bridge maintains 1:1 parity

**Operations:**
- `Deposit`: x/bank → Domain.Treasury
- `Withdraw`: Domain.Treasury → x/bank
- `Transfer`: Within Domain.Treasury (cheap)

---

## IBC Integration

### Modules Used

- **ibc-go v8.4.0:** Core IBC protocol
- **ICS-20 Transfer:** Token transfers
- **Capability Module:** Port/channel binding

### Channel Architecture

```
TrueRepublic                    Cosmos Hub
│                               │
│  ┌─────────────────────┐      │
├──│ Client (light client)│─────┤
│  └─────────────────────┘      │
│                               │
│  ┌─────────────────────┐      │
├──│ Connection          │─────┤
│  └─────────────────────┘      │
│                               │
│  ┌─────────────────────┐      │
├──│ Channel (transfer)  │─────┤
│  └─────────────────────┘      │
│                               │
▼                               ▼
IBC Transfers ←──────────────→ IBC Transfers
```

### Relayer

- **Hermes** or **Go Relayer** supported
- Monitors both chains for packets
- Submits proofs for verification
- Handles timeouts and acknowledgments
- Config guide: `docs/IBC_RELAYER_SETUP.md`

---

## Zero-Knowledge Proofs

### Circuit Design

**Public Inputs:**
- Merkle root (domain membership)
- Nullifier (uniqueness)
- Rating (1-10)

**Private Inputs:**
- Identity secret
- Merkle path (proof of membership)

**Verification:**
```
Verify(proof, publicInputs, verificationKey) → bool
```

### Big Purge Mechanism

- **Purpose:** Refresh anonymity set
- **Interval:** 90 days (configurable)
- **Process:**
  1. New Merkle tree built from current members
  2. Old ratings archived
  3. New verification key generated
  4. Nullifiers reset

---

## Data Flow Examples

### Anonymous Vote

```
User generates ZKP proof (client-side)
├─ Inputs: identity, merkle_path, rating
└─ Output: proof + nullifier

Submit MsgRateWithProof
├─ Nullifier checked (not used)
└─ Proof verified against VK

If valid:
├─ Store rating under suggestion
├─ Mark nullifier as used
└─ Emit event
```

### Multi-Asset Swap

```
User wants: PNYX → ETH
├─ No direct pool exists
└─ Route finder: PNYX → BTC → ETH

MsgSwapExact submitted
├─ Route: [pool-0, pool-1]
└─ Min output: 1000 ETH

Execute atomically:
├─ Swap PNYX → BTC in pool-0
├─ Swap BTC → ETH in pool-1
└─ Check: output ≥ min_output

If all succeed:
├─ Transfer tokens
└─ Emit multi_hop_swap event
```

### CosmWasm Contract Call

```
Contract queries domain members
├─ Custom query: DomainMembers
└─ Returns: [addr1, addr2, ...]

Contract executes stone placement
├─ Custom message: PlaceStoneOnSuggestion
└─ Calls x/truedemocracy

Verification:
├─ Contract authorized?
├─ Stone type valid?
└─ Suggestion exists?

If valid:
├─ Place stone
└─ Return success
```

---

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

---

## Security Considerations

**ZKP:**
- Groth16 trusted setup (ceremony required for mainnet)
- MiMC hash for Merkle tree (ZK-friendly)
- Nullifier prevents double-voting
- Big Purge limits traceability

**DEX:**
- Slippage protection on all swaps
- Asset whitelisting (admin controlled)
- Fee validation (0.3% enforced)
- Reserve overflow checks

**CosmWasm:**
- Gas metering prevents DoS
- Sandboxed execution
- Custom binding authorization
- Bridge balance checks

**IBC:**
- Light client verification
- Timeout handling
- Escrow accounts
- Channel upgrades supported

---

## Test Coverage

| Language | Tests | Key Areas |
|----------|-------|-----------|
| Go | 533 | Modules, IBC, treasury |
| Rust | 26 | Contract examples, testing utils |
| Frontend | 18 | ZKP + DEX components |
| **Total** | **577** | **All passing** |
