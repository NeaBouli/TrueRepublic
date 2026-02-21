# System Architecture

## Layer Diagram

```
┌─────────────────────────────────────────────────────┐
│  Client Layer                                        │
│  Web Wallet (React 18 + Keplr + CosmJS)             │
│  Mobile Wallet (Expo / React Native + CosmJS)       │
│  CLI (truerepublicd tx/query)                       │
├─────────────────────────────────────────────────────┤
│  Smart Contract Layer (CosmWasm)                     │
│  governance.rs  │  treasury.rs                       │
├─────────────────────────────────────────────────────┤
│  Application Layer (Cosmos SDK v0.50.13)             │
│  truedemocracy  │  dex  │  treasury                  │
├─────────────────────────────────────────────────────┤
│  Consensus Layer (CometBFT v0.38.21)                 │
│  Proof of Domain (PoD) validator selection           │
├─────────────────────────────────────────────────────┤
│  Storage Layer                                       │
│  IAVL Trees (KV Store per module)                    │
└─────────────────────────────────────────────────────┘
```

## Data Flow

### Transaction Flow

```
User → Wallet (Keplr/CosmJS) → Sign Transaction
  → CometBFT RPC (port 26657) → Mempool
  → Block Proposal → ABCI DeliverTx
  → Cosmos SDK Router → Module Message Handler
  → Keeper → KV Store (commit)
  → EndBlock → Validator Set Updates
  → Block Committed → Event Emitted
```

### Query Flow

```
Client → ABCI Query (RPC port 26657)
  → /custom/{module}/{route}/{params}
  → Querier → Keeper → KV Store (read)
  → JSON Response
```

## Application Entry Point

The application is defined in `app.go`:

```go
type TrueRepublicApp struct {
    *baseapp.BaseApp
    mm        *module.Manager       // Module manager
    cdc       *codec.LegacyAmino    // Amino codec
    appCodec  codec.Codec           // Protobuf codec
    keys      map[string]*storetypes.KVStoreKey
    tdKeeper  truedemocracy.Keeper  // Governance keeper
    dexKeeper dex.Keeper            // DEX keeper
}
```

### Key Functions

| Function | Purpose |
|----------|---------|
| `NewTrueRepublicApp()` | Initialize app, register modules, set up genesis |
| `Query()` | Custom ABCI query interceptor for `custom/` paths |
| `InitChainer()` | Process genesis state, create initial validators |
| `EndBlocker()` | Run end-of-block logic for all modules |

### EndBlock Processing Order

1. Distribute staking rewards (every 3,600 blocks)
2. Distribute domain treasury interest
3. Enforce domain membership (evict validators without domains)
4. Process suggestion lifecycles (zone transitions, auto-delete)
5. Process governance (admin election, inactivity cleanup)
6. Return validator set updates to CometBFT

## Module Architecture

### truedemocracy Module

The core governance module implementing the whitepaper specification.

**Files:**

| File | Responsibility |
|------|---------------|
| `keeper.go` | Domain CRUD, proposal submission, anonymous ratings |
| `validator.go` | PoD validator registration, lifecycle, staking rewards |
| `slashing.go` | Double-sign (5%) and downtime (1%) penalties |
| `anonymity.go` | Permission register, domain key pairs for anonymous voting |
| `stones.go` | VoteToEarn rewards, stone voting, list sorting |
| `lifecycle.go` | Suggestion green/yellow/red zones, auto-delete, fast-delete |
| `governance.go` | Admin election, member exclusion (2/3 vote), inactivity cleanup |
| `msgs.go` | 13 SDK message types with validation |
| `cli.go` | CLI transaction and query commands |
| `querier.go` | ABCI query route handler |
| `types.go` | Data structures: Domain, Validator, Issue, Suggestion, Rating |
| `module.go` | SDK module wiring, InitGenesis, EndBlock |
| `tree.go` | Hierarchical node tree for vote propagation |

### dex Module

AMM-based decentralized exchange with constant-product formula.

**Files:**

| File | Responsibility |
|------|---------------|
| `keeper.go` | CreatePool, Swap, AddLiquidity, RemoveLiquidity |
| `msgs.go` | 4 SDK message types |
| `cli.go` | CLI commands |
| `types.go` | Pool type, fee constants (SwapFeeBps=30, BurnBps=100) |
| `module.go` | SDK module wiring |

### treasury Module

Tokenomics equations from the whitepaper.

**Files:**

| File | Responsibility |
|------|---------------|
| `rewards.go` | All 5 whitepaper equations for token economics |
| `rewards_test.go` | 31 equation validation tests |

## Data Models

### Domain

```go
type Domain struct {
    Name              string          // Unique identifier
    Admin             sdk.AccAddress  // Current admin (elected by stones)
    Members           []string        // Member addresses
    Treasury          sdk.Coins       // Domain funds
    Issues            []Issue         // Active issues
    Options           DomainOptions   // Configuration
    PermissionReg     []string        // Anonymous voting public keys
    TotalPayouts      int64           // Cumulative PNYX distributed
    TransferredStake  int64           // Cumulative validator withdrawals
}
```

### Issue & Suggestion

```go
type Issue struct {
    Name           string        // Issue title
    Stones         int           // Stone count for sorting
    Suggestions    []Suggestion  // Proposed solutions
    CreationDate   int64         // Unix timestamp
    LastActivityAt int64         // Last interaction timestamp
    ExternalLink   string        // Optional URL reference
}

type Suggestion struct {
    Name            string    // Suggestion title
    Creator         string    // Creator address
    Stones          int       // Stone count
    Ratings         []Rating  // Systemic consensing ratings
    Color           string    // Zone: green/yellow/red
    DwellTime       int64     // Time in current zone (seconds)
    CreationDate    int64     // Unix timestamp
    ExternalLink    string    // Optional URL
    EnteredYellowAt int64     // Timestamp of zone transition
    EnteredRedAt    int64     // Timestamp of zone transition
    DeleteVotes     int       // Fast-delete vote counter
}
```

### Validator

```go
type Validator struct {
    OperatorAddr string     // Bech32 operator address
    PubKey       []byte     // Ed25519 public key (32 bytes)
    Stake        sdk.Coins  // Staked PNYX
    Domains      []string   // PoD domain memberships
    Power        int64      // Voting power (stake / StakeMin)
    Jailed       bool       // Whether currently jailed
    JailedUntil  int64      // Unix timestamp when unjailable
    MissedBlocks int64      // Consecutive missed blocks
}
```

### Pool (DEX)

```go
type Pool struct {
    PnyxReserve  math.Int  // PNYX reserve (x in x*y=k)
    AssetReserve math.Int  // Asset reserve (y)
    AssetDenom   string    // Asset denomination
    TotalShares  math.Int  // LP share tokens
    TotalBurned  math.Int  // Cumulative PNYX burned
}
```

## Test Coverage

| Module | Tests | Test Lines |
|--------|-------|------------|
| truedemocracy | 116 | 2,077 |
| dex | 24 | 423 |
| treasury | 31 | 205 |
| **Total** | **182** | **2,705** |

## Next Steps

- [Module Reference](module-reference.md) -- Detailed module docs
- [CLI Commands](../api-reference/cli-commands.md) -- API reference
- [CosmJS Examples](../integration-guide/cosmjs-examples.md) -- Integration code
