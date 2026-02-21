# Architecture Overview

Complete architectural documentation of TrueRepublic/PNYX blockchain.

## Table of Contents

1. [High-Level Architecture](#high-level-architecture)
2. [Technology Stack](#technology-stack)
3. [Design Decisions](#design-decisions)
4. [Module Architecture](#module-architecture)
5. [Data Flow](#data-flow)
6. [Security Architecture](#security-architecture)
7. [Performance](#performance)

---

## High-Level Architecture

### System Layers

```
CLIENT LAYER
├── Web Wallet (React 18 + Keplr + CosmJS + Tailwind)
├── Mobile Wallet (React Native + Expo + CosmJS)
└── CLI (truerepublicd, Cobra-based)
    ↓
API LAYER
├── CometBFT RPC (port 26657) ← Primary query/broadcast interface
├── REST/LCD API (port 1317) ← Standard Cosmos REST
├── gRPC API (port 9090) ← Protobuf RPC
├── ABCI Queries ← Custom module queries via /custom/{module}/...
└── WebSocket (port 26657/websocket) ← Real-time events
    ↓
APPLICATION LAYER (Cosmos SDK v0.50.13)
├── x/truedemocracy ← Core governance (13 msg types, 116 tests)
│   ├── keeper.go ← Domain CRUD, proposals, ratings
│   ├── anonymity.go ← Permission register, domain key pairs (WP S4)
│   ├── stones.go ← VoteToEarn, stone voting, list sorting (WP S3.1)
│   ├── lifecycle.go ← Green/yellow/red zones, auto-delete (WP S3.1.2)
│   ├── governance.go ← Admin election, exclusion, cleanup (WP S3.6)
│   ├── validator.go ← Proof of Domain, staking, transfer limits
│   └── slashing.go ← Double-sign (5%), downtime (1%)
├── x/dex ← AMM exchange (4 msg types, 24 tests)
│   └── keeper.go ← CreatePool, Swap (x*y=k), Add/RemoveLiquidity
├── treasury/keeper ← Tokenomics equations 1-5 (31 tests)
│   └── rewards.go ← Domain interest, staking rewards, decay
├── CosmWasm ← Smart contracts (governance.rs, treasury.rs)
└── Standard modules (auth, bank, staking, etc.)
    ↓
CONSENSUS LAYER (CometBFT v0.38.21)
├── Byzantine Fault Tolerance (instant finality)
├── P2P Networking (port 26656)
├── Block Production (~5s blocks)
├── Proof of Domain validator selection
└── Prometheus Metrics (port 26660)
    ↓
STORAGE LAYER
└── IAVL Trees (KV Store per module)
    ├── domains/{name} → Domain state
    ├── validators/{addr} → Validator state
    └── pools/{denom} → DEX pool state
```

---

## Technology Stack

### Backend: Go 1.23+

| Aspect | Detail |
|--------|--------|
| **Why Go?** | Cosmos SDK requirement, excellent performance, strong concurrency |
| **Key Libraries** | Cosmos SDK, CometBFT, Cobra CLI, LevelDB |
| **Build** | `make build` produces `truerepublicd` binary |
| **Test** | `go test ./... -race -cover` (182 tests) |

### Framework: Cosmos SDK v0.50.13

| Aspect | Detail |
|--------|--------|
| **Why Cosmos SDK?** | Proven blockchain framework, modular, IBC-ready, battle-tested |
| **Custom Modules** | `x/truedemocracy`, `x/dex` |
| **Standard Modules** | auth, bank, staking, gov, distribution |
| **Codec** | Amino (legacy) + Protobuf (modern) |

### Consensus: CometBFT v0.38.21

| Aspect | Detail |
|--------|--------|
| **Why CometBFT?** | BFT consensus, instant finality, energy-efficient |
| **Block Time** | ~5 seconds |
| **Finality** | Instant (no reorgs, no confirmations needed) |
| **Fault Tolerance** | Tolerates < 1/3 Byzantine validators |

### Smart Contracts: Rust + CosmWasm v3

| Aspect | Detail |
|--------|--------|
| **Why Rust?** | Memory safety, no GC, prevents buffer overflows |
| **Why CosmWasm?** | Wasm sandboxing, gas metering, Cosmos-native |
| **Contracts** | `governance.rs` (proposals + SC voting), `treasury.rs` (deposit/withdraw) |

### Frontend: React 18

| Aspect | Web | Mobile |
|--------|-----|--------|
| **Framework** | React 18 | React Native 0.74 + Expo 51 |
| **Styling** | Tailwind CSS 3.4 | React Native StyleSheet |
| **Navigation** | React Router 6 | React Navigation 6.5 |
| **Wallet** | Keplr browser extension | In-app key management |
| **Blockchain** | CosmJS 0.32-0.38 | CosmJS 0.32-0.38 |

---

## Design Decisions

### 1. Why Cosmos SDK over Custom Chain?

**Decision:** Build on Cosmos SDK

**Alternatives Considered:**
- Substrate (Polkadot) -- Less familiar, different ecosystem
- Custom from scratch -- Too much effort, security risk
- Ethereum L2 -- EVM constraints, not governance-focused

**Reasoning:** Modularity, IBC interoperability, battle-tested security, large ecosystem, faster development velocity.

### 2. Why Systemic Consensing over Binary Voting?

**Decision:** -5 to +5 rating scale

**Alternatives Considered:**
- Yes/No voting -- Too simplistic, no nuance
- Quadratic voting -- Complex, harder to understand
- Ranked choice -- Complex counting, no intensity measure

**Reasoning:** Captures intensity of preferences, finds least-resistance solutions, makes minority concerns visible, prevents polarization. Whitepaper requirement (WP S3).

### 3. Why Proof-of-Domain over Standard PoS?

**Decision:** Validators must be domain members

**Reasoning:** Prevents plutocracy (wealth alone doesn't guarantee validation), creates community accountability, aligns validator incentives with governance mission. Transfer limit (10% of domain payouts, WP S7) prevents value extraction.

### 4. Why AMM DEX over Order Book?

**Decision:** Constant-product AMM (x*y=k)

**Reasoning:** Simplicity, fully on-chain (no off-chain matching), anyone can provide liquidity, proven model. 1% PNYX burn creates deflationary pressure (WP S5).

### 5. Why Domain Key Pairs over ZKP?

**Decision:** Ed25519 domain-specific keys for anonymous voting

**Reasoning:** Simpler than zero-knowledge proofs, no complex cryptographic overhead, domain key is unlinkable to master key, sufficient privacy for governance voting. ZKP can be added later.

### 6. Why a 3-Column Layout?

**Decision:** Left (domains) / Center (proposals) / Right (details + actions)

**Reasoning:** UI design patterns inspired by Telegram's open-source web client (3-panel messaging layout). This is design inspiration only -- TrueRepublic has no code dependencies on Telegram and shares no source code with it. The pattern maps naturally to governance flow (browse domains -> view proposals -> take action) and is responsive (collapses on mobile).

### 7. Why Flat Module Files over Subdirectories?

**Decision:** All module files in single directory (e.g., `x/truedemocracy/`)

**Reasoning:** Each file has a clear responsibility (stones.go, lifecycle.go, governance.go, etc.), easier to navigate for a single-team project, follows Cosmos SDK conventions for simpler modules.

---

## Module Architecture

### x/truedemocracy -- Core Governance

**13 message types, 116 tests, 6 source files + types/module/CLI**

```
x/truedemocracy/
├── keeper.go         ← Domain CRUD, proposal submission, ratings
├── anonymity.go      ← Permission register, domain key pairs
├── stones.go         ← VoteToEarn rewards, stone voting, list sorting
├── lifecycle.go      ← Green/yellow/red zones, auto-delete, fast-delete
├── governance.go     ← Admin election, member exclusion, inactivity cleanup
├── validator.go      ← PoD registration, staking, transfer limits
├── slashing.go       ← Double-sign (5%), downtime (1%), jailing
├── msgs.go           ← 13 SDK message types with validation
├── msg_server.go     ← gRPC message handlers
├── cli.go            ← 13 tx + 4 query CLI commands
├── querier.go        ← Legacy ABCI query routes
├── query_server.go   ← gRPC query handlers
├── types.go          ← Domain, Validator, Issue, Suggestion, Rating
├── tree.go           ← Hierarchical node tree for vote propagation
├── module.go         ← SDK module wiring, InitGenesis, EndBlock
└── *_test.go         ← 116 tests (stones, lifecycle, governance, anonymity, validator, slashing)
```

**EndBlock Processing Order:**
1. Distribute staking rewards (every 3,600s)
2. Distribute domain treasury interest (25% APY)
3. Enforce domain membership (evict validators without domains)
4. Process suggestion lifecycles (zone transitions, auto-delete)
5. Process governance (admin election, inactivity cleanup)
6. Return validator set updates to CometBFT

### x/dex -- Decentralized Exchange

**4 message types, 24 tests**

```
x/dex/
├── keeper.go         ← CreatePool, Swap (x*y=k), AddLiquidity, RemoveLiquidity
├── msgs.go           ← 4 message types
├── msg_server.go     ← gRPC message handlers
├── cli.go            ← 4 tx + 2 query CLI commands
├── querier.go        ← Legacy ABCI query routes
├── query_server.go   ← gRPC query handlers
├── types.go          ← Pool type, fee constants (SwapFeeBps=30, BurnBps=100)
├── module.go         ← SDK module wiring
└── keeper_test.go    ← 24 tests (swap, liquidity, fees, burn)
```

**Swap Formula:** `output = (outReserve * input * 9970) / (inReserve * 10000 + input * 9970)`

### treasury -- Tokenomics

**5 equations, 31 tests**

```
treasury/keeper/
├── rewards.go        ← All 5 whitepaper tokenomics equations
└── rewards_test.go   ← 31 equation validation tests
```

**Constants:** CDom=2, CPut=15, CEarn=1000, StakeMin=100,000, SupplyMax=21,000,000, ApyDom=25%, ApyNode=10%

---

## Data Flow

### Proposal Submission Flow

```
1. User fills form in Web Wallet (right panel)
2. Web Wallet builds message via services/api.js
3. Keplr signs transaction (Ed25519)
4. Transaction broadcast to CometBFT RPC (port 26657)
5. CometBFT adds to mempool
6. Validator includes in next block (~5s)
7. Cosmos SDK BaseApp routes to x/truedemocracy msg_server
8. MsgServer.SubmitProposal handler:
   a. Validate domain exists and sender is member
   b. Deduct PayToPut fee from sender
   c. Credit fee to domain treasury
   d. Create Issue + Suggestion (green zone)
   e. Set creation timestamps
   f. Store in KV store
9. State committed to IAVL tree
10. Block finalized (instant finality)
11. Web Wallet polls/refreshes to see new proposal
```

### DEX Swap Flow

```
1. User enters swap amount on DEX page
2. Web Wallet calls swapTokens() from services/api.js
3. Keplr signs MsgSwap transaction
4. Transaction broadcast and included in block
5. x/dex msg_server processes:
   a. Validate pool exists
   b. Calculate output: (outReserve * input * 9970) / (inReserve * 10000 + input * 9970)
   c. Transfer input tokens from user to pool
   d. If output is PNYX: burn 1% (BurnBps=100)
   e. Transfer remaining output to user
   f. Update pool reserves
   g. Record burned amount in pool.TotalBurned
6. User receives output tokens
7. Balance auto-refreshes (useWallet hook, 10s interval)
```

### Stones Voting Flow

```
1. User clicks vote on an issue/suggestion
2. Web Wallet builds MsgPlaceStoneOnIssue/Suggestion
3. Transaction processed by x/truedemocracy:
   a. Verify sender is domain member
   b. Remove sender's previous stone (if any)
   c. Increment target's stone count
   d. Re-sort lists by stone count (descending)
   e. Calculate VoteToEarn reward: treasury / CEarn
   f. Transfer reward from domain treasury to voter
4. UI updates to show new stone counts
```

---

## Security Architecture

### Authentication & Authorization

| Layer | Mechanism |
|-------|-----------|
| Transaction signing | secp256k1 / Ed25519 via Keplr |
| On-chain verification | Cosmos SDK `GetSigners()` per message |
| Domain membership | Keeper checks `domain.Members` |
| Admin actions | Keeper checks `domain.Admin` |
| Validator actions | Keeper checks `validator.OperatorAddr` |

### Anonymous Voting (WP S4)

```
Master Key (Keplr wallet)
    │
    ├── Domain A Key (Ed25519, unlinkable)
    ├── Domain B Key (Ed25519, unlinkable)
    └── Domain C Key (Ed25519, unlinkable)
```

1. Member generates domain-specific Ed25519 key pair
2. Public key registered via `MsgJoinPermissionRegister`
3. Ratings submitted with domain public key (not master key)
4. Votes are **unlinkable** between domains and to member identity
5. Admin can purge register (`MsgPurgePermissionRegister`) to reset

### Validator Security

| Infraction | Penalty | Recovery |
|-----------|---------|----------|
| Double-signing | 5% slash + jail 100 min | `MsgUnjail` after period |
| Downtime (>50/100 blocks) | 1% slash + jail 10 min | `MsgUnjail` after period |
| No domain membership | Eviction (EndBlock) | Re-register with domain |
| Exceeding transfer limit | Tx rejected | Wait for domain payouts |

### Smart Contract Security

- Wasm sandboxing (isolated execution)
- Gas metering (DOS protection)
- Actor model (no shared state)
- Code upload governance (controlled deployment)

---

## Performance

### Throughput

| Metric | Value |
|--------|-------|
| Block time | ~5 seconds |
| Finality | Instant |
| Target TPS | 1,000+ |
| State writes | IAVL tree (LevelDB/RocksDB) |

### Storage Growth

| Data Type | Size Per Unit |
|-----------|--------------|
| Domain | ~1-5 KB (depends on members/issues) |
| Issue + suggestions | ~2-5 KB |
| Rating | ~100 bytes |
| Validator | ~500 bytes |
| Pool | ~200 bytes |

### Pruning Options

| Strategy | Description |
|----------|-------------|
| `default` | Keep last 100 states |
| `nothing` | Keep all (archive node) |
| `everything` | Keep only current state (minimal disk) |

---

## Next Steps

- [Code Structure](Code-Structure) -- File organization and conventions
- [Module Deep-Dive](Module-Deep-Dive) -- Detailed message and handler documentation
- [Development Setup](Development-Setup) -- Set up your development environment
