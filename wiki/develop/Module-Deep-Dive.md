# Module Deep-Dive

Detailed documentation of all TrueRepublic modules, their message types, handlers, state, and query endpoints.

## Table of Contents

1. [x/truedemocracy Module](#xtruedemocracy-module)
   - [Data Types](#data-types)
   - [Domain Management](#domain-management-messages)
   - [Proposals & Ratings](#proposal--rating-messages)
   - [Stones Voting](#stones-voting-messages)
   - [Governance Actions](#governance-action-messages)
   - [Validator Operations](#validator-operation-messages)
   - [Anonymous Voting](#anonymous-voting-messages)
   - [Query Endpoints](#truedemocracy-query-endpoints)
   - [EndBlock Logic](#endblock-logic)
2. [x/dex Module](#xdex-module)
3. [treasury Module](#treasury-module)
4. [CosmWasm Contracts](#cosmwasm-contracts)

---

## x/truedemocracy Module

**Path:** `x/truedemocracy/`
**Purpose:** Core governance -- domains, proposals, voting, validators
**Messages:** 13 transaction types
**Queries:** 4 query endpoints
**Tests:** 116 unit tests across 6 test files

### Data Types

Defined in `x/truedemocracy/types.go`:

#### Domain

```go
type Domain struct {
    Name              string            // Unique identifier
    Admin             sdk.AccAddress    // Elected admin (by stones)
    Members           []string          // Member addresses
    Treasury          sdk.Coins         // Community funds
    Issues            []Issue           // Active issues
    Options           DomainOptions     // Configuration flags
    PermissionReg     []string          // Anonymous voting public keys (hex)
    TotalPayouts      int64             // Cumulative PNYX distributed
    TransferredStake  int64             // Cumulative validator withdrawals
}

type DomainOptions struct {
    AdminElectable    bool    // Whether admin is elected by stones
    AnyoneCanJoin     bool    // Open vs. invite-only
    OnlyAdminIssues   bool    // Whether only admin can create issues
    CoinBurnRequired  bool    // Whether proposals require PNYX fee
    ApprovalThreshold int64   // Rating threshold in basis points (default 500 = 5%)
    DefaultDwellTime  int64   // Seconds per lifecycle zone (default 86400 = 1 day)
}
```

#### Issue & Suggestion

```go
type Issue struct {
    Name           string        // Issue title
    Stones         int           // Stone count (for sorting)
    Suggestions    []Suggestion  // Proposed solutions
    CreationDate   int64         // Unix timestamp
    LastActivityAt int64         // Last interaction timestamp
    ExternalLink   string        // Optional URL reference
}

type Suggestion struct {
    Name            string    // Suggestion title
    Creator         string    // Creator address
    Stones          int       // Stone count
    Ratings         []Rating  // Systemic consensing ratings (-5 to +5)
    Color           string    // Lifecycle zone: "green", "yellow", "red"
    DwellTime       int64     // Seconds in current zone
    CreationDate    int64     // Unix timestamp
    ExternalLink    string    // Optional URL
    EnteredYellowAt int64     // Timestamp of green → yellow transition
    EnteredRedAt    int64     // Timestamp of yellow → red transition
    DeleteVotes     int       // Fast-delete vote counter
}

type Rating struct {
    DomainPubKeyHex string  // Hex-encoded domain Ed25519 public key
    Value           int     // Systemic consensing value: -5 to +5
}
```

#### Validator

```go
type Validator struct {
    OperatorAddr string     // Bech32 operator address
    PubKey       []byte     // Ed25519 public key (32 bytes)
    Stake        sdk.Coins  // Staked PNYX
    Domains      []string   // PoD domain memberships
    Power        int64      // Voting power = stake / StakeMin
    Jailed       bool       // Currently jailed?
    JailedUntil  int64      // Unix timestamp when unjailable
    MissedBlocks int64      // Missed blocks in current window
}
```

---

### Domain Management Messages

#### MsgCreateDomain

**Purpose:** Create a new governance domain with initial treasury.

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Unique domain name |
| `admin` | AccAddress | Creator becomes initial admin |
| `initial_coins` | Coins | Initial treasury funding |

**Handler logic** (`msg_server.go`):
1. Validate domain name doesn't already exist
2. Deduct domain creation cost from creator: `fee * CDom * CEarn` (eq.1)
3. Create domain with creator as admin and sole member
4. Set treasury to `initial_coins`
5. Store domain in KV store

**CLI:**
```bash
truerepublicd tx truedemocracy create-domain [name] [initial-coins] \
    --from mykey --chain-id truerepublic-1
```

---

### Proposal & Rating Messages

#### MsgSubmitProposal

**Purpose:** Submit an issue with a suggestion for community evaluation.

| Field | Type | Description |
|-------|------|-------------|
| `sender` | AccAddress | Submitter (must be domain member) |
| `domain_name` | string | Target domain |
| `issue_name` | string | Issue title (the problem) |
| `suggestion_name` | string | Suggestion title (proposed solution) |
| `creator` | AccAddress | Creator address |
| `fee` | Coins | PayToPut fee |
| `external_link` | string | Optional URL for detailed proposal |

**Handler logic:**
1. Verify sender is domain member
2. Deduct PayToPut fee: `min(reward * CPut, reward * nMembers)` where `reward = treasury / CEarn` (eq.2, eq.3)
3. Credit fee to domain treasury
4. Create Issue with the given name
5. Create Suggestion under the issue (color = "green", DwellTime = 0)
6. Set creation timestamps

**PayToPut calculation example:**
```
Treasury = 500,000 PNYX, Members = 10
reward = 500,000 / 1,000 = 500 PNYX
fee = min(500 * 15, 500 * 10) = min(7,500, 5,000) = 5,000 PNYX
```

**CLI:**
```bash
truerepublicd tx truedemocracy submit-proposal \
    [domain] [issue] [suggestion] [fee]pnyx [external-link] \
    --from mykey --chain-id truerepublic-1
```

---

### Stones Voting Messages

Implemented in `x/truedemocracy/stones.go`.

#### MsgPlaceStoneOnIssue

| Field | Type | Description |
|-------|------|-------------|
| `sender` | AccAddress | Voter |
| `domain_name` | string | Domain |
| `issue_name` | string | Target issue |
| `member_addr` | AccAddress | Voting member |

**Handler logic:**
1. Verify member is in domain
2. Remove member's previous stone on any issue (1 stone rule)
3. Increment issue's stone count
4. Re-sort issues by stone count (descending), then by date
5. Calculate VoteToEarn reward: `treasury / CEarn` (eq.2)
6. Transfer reward from domain treasury to voter
7. Update domain's TotalPayouts

#### MsgPlaceStoneOnSuggestion

Same pattern as issue, but targets a suggestion within an issue.

#### MsgPlaceStoneOnMember

Places a stone on another member for **admin election** (WP S3.6).

**Handler logic:**
1. Verify both voter and target are domain members
2. Remove voter's previous member stone
3. Increment target member's stone count
4. The member with the most stones becomes admin (checked in EndBlock)

---

### Governance Action Messages

Implemented in `x/truedemocracy/governance.go`.

#### MsgVoteToExclude

| Field | Type | Description |
|-------|------|-------------|
| `sender` | AccAddress | Voter |
| `domain_name` | string | Domain |
| `target_member` | AccAddress | Member to exclude |
| `voter_addr` | AccAddress | Voting member |

**Handler logic:**
1. Verify voter is domain member
2. Record vote
3. If votes >= 2/3 of members (6,667 basis points): remove target from domain
4. If target was admin, trigger admin re-election

#### MsgVoteToDelete

**Purpose:** Fast-delete a suggestion by 2/3 majority vote.

| Field | Type | Description |
|-------|------|-------------|
| `sender` | AccAddress | Voter |
| `domain_name` | string | Domain |
| `issue_name` | string | Issue containing suggestion |
| `suggestion_name` | string | Suggestion to delete |
| `member_addr` | AccAddress | Voting member |

**Handler logic:**
1. Verify voter is domain member
2. Increment suggestion's `DeleteVotes`
3. If `DeleteVotes >= 2/3 * len(domain.Members)`: delete suggestion immediately

---

### Validator Operation Messages

Implemented in `x/truedemocracy/validator.go` and `slashing.go`.

#### MsgRegisterValidator

| Field | Type | Description |
|-------|------|-------------|
| `sender` | AccAddress | Operator |
| `operator_addr` | AccAddress | Validator address |
| `pub_key` | string | Ed25519 public key (hex, 32 bytes) |
| `stake` | Coins | Stake amount (min 100,000 PNYX) |
| `domain_name` | string | Domain membership |

**Handler logic:**
1. Verify stake >= `StakeMin` (100,000 PNYX)
2. Verify sender is member of specified domain
3. Create validator with `Power = stake / StakeMin`
4. Deduct stake from sender's account
5. Return validator update to CometBFT

#### MsgWithdrawStake

**Transfer limit (WP S7):**
```
max_withdrawal = domain.TotalPayouts * 10%
already_withdrawn = domain.TransferredStake
remaining_limit = max_withdrawal - already_withdrawn
actual_withdrawal = min(requested_amount, remaining_limit)
```

#### MsgUnjail

Unjails a validator after the jail period has expired.

**Requirements:**
- Current time > `validator.JailedUntil`
- Stake still >= `StakeMin` (100,000 PNYX)

#### Slashing Parameters

| Infraction | Slash | Jail Duration | Detection |
|-----------|-------|---------------|-----------|
| Double-sign | 5% of stake | 100 minutes | CometBFT evidence |
| Downtime | 1% of stake | 10 minutes (600s) | >50 missed in 100-block window |

**Post-slash:** If stake drops below StakeMin, voting power is set to 0.

---

### Anonymous Voting Messages

Implemented in `x/truedemocracy/anonymity.go`.

#### MsgJoinPermissionRegister

| Field | Type | Description |
|-------|------|-------------|
| `sender` | AccAddress | Member |
| `domain_name` | string | Domain |
| `member_addr` | AccAddress | Member address |
| `domain_pub_key` | string | Domain Ed25519 public key (hex) |

**Handler logic:**
1. Verify sender is domain member
2. Verify public key is valid hex-encoded Ed25519 (32 bytes)
3. Add public key to domain's PermissionReg
4. Future ratings use this key (unlinkable to member identity)

#### MsgPurgePermissionRegister

| Field | Type | Description |
|-------|------|-------------|
| `caller` | AccAddress | Must be domain admin |
| `domain_name` | string | Domain |

**Handler logic:**
1. Verify caller is domain admin
2. Clear domain's PermissionReg (all keys removed)
3. Members must re-register new keys for future anonymous voting

---

### truedemocracy Query Endpoints

| Route | ABCI Path | Returns |
|-------|-----------|---------|
| Domain | `custom/truedemocracy/domain/{name}` | Single Domain JSON |
| Domains | `custom/truedemocracy/domains` | Array of all domains |
| Validator | `custom/truedemocracy/validator/{addr}` | Single Validator JSON |
| Validators | `custom/truedemocracy/validators` | Array of all validators |

**CLI:**
```bash
truerepublicd query truedemocracy domain [name]
truerepublicd query truedemocracy domains
truerepublicd query truedemocracy validator [addr]
truerepublicd query truedemocracy validators
```

---

### EndBlock Logic

Executed at the end of every block, defined in `module.go`:

```
EndBlock(ctx, req):
  1. Every RewardInterval (3600s):
     → Distribute staking rewards to validators (eq.5: stake * 10% APY * decay)
     → Distribute domain interest to treasuries (eq.4: treasury * 25% APY * decay)

  2. Enforce PoD membership:
     → For each validator: if len(validator.Domains) == 0 → evict

  3. Process suggestion lifecycles (lifecycle.go):
     → For each suggestion:
        - Increment DwellTime
        - Green → Yellow: when DwellTime > DefaultDwellTime
        - Yellow → Red: when DwellTime > DefaultDwellTime * 2
        - Red expired: auto-delete when DwellTime > DefaultDwellTime * 3

  4. Process governance (governance.go):
     → Update admin election (highest-stoned member)
     → Remove inactive members (>360 days since last activity)

  5. Return validator set updates to CometBFT
```

---

## x/dex Module

**Path:** `x/dex/`
**Purpose:** Decentralized exchange with AMM
**Formula:** x * y = k (constant product)
**Messages:** 4 transaction types
**Queries:** 2 query endpoints
**Tests:** 24 unit tests

### Data Types

```go
type Pool struct {
    PnyxReserve  math.Int  // PNYX reserve (x in x*y=k)
    AssetReserve math.Int  // Asset reserve (y)
    AssetDenom   string    // Asset denomination (e.g., "atom")
    TotalShares  math.Int  // LP share tokens outstanding
    TotalBurned  math.Int  // Cumulative PNYX burned
}

// Fee constants
const SwapFeeBps = 30    // 0.3% swap fee
const BurnBps    = 100   // 1% PNYX burn on PNYX output
```

### MsgCreatePool

Creates a new PNYX/asset liquidity pool.

| Field | Type | Description |
|-------|------|-------------|
| `sender` | AccAddress | Pool creator |
| `asset_denom` | string | Asset denomination |
| `pnyx_amt` | Int | PNYX to deposit |
| `asset_amt` | Int | Asset to deposit |

**Handler logic:**
1. Verify pool doesn't exist for this denom
2. Transfer both tokens from sender to module account
3. Calculate initial shares: `sqrt(pnyx_amt * asset_amt)` (integer sqrt)
4. Create pool with reserves and shares
5. Credit shares to creator

### MsgSwap

Swaps tokens via the AMM.

| Field | Type | Description |
|-------|------|-------------|
| `sender` | AccAddress | Trader |
| `input_denom` | string | Input token |
| `input_amt` | Int | Input amount |
| `output_denom` | string | Output token |

**Swap formula (with 0.3% fee):**
```
output = (outReserve * input * 9970) / (inReserve * 10000 + input * 9970)
```

**PNYX burn (1% when output is PNYX):**
```
burn = output * BurnBps / 10000
user_receives = output - burn
pool.TotalBurned += burn
```

**Example:** Swap 1,000 PNYX for ATOM in a pool with 100,000 PNYX / 50,000 ATOM:
```
output = (50000 * 1000 * 9970) / (100000 * 10000 + 1000 * 9970)
       = 498,500,000,000 / 1,009,970,000
       ≈ 493 ATOM
```

### MsgAddLiquidity

| Field | Type | Description |
|-------|------|-------------|
| `sender` | AccAddress | LP provider |
| `asset_denom` | string | Pool asset |
| `pnyx_amt` | Int | PNYX to add |
| `asset_amt` | Int | Asset to add |

**Shares calculation:**
```
shares = min(
    pnyx_added * total_shares / pnyx_reserve,
    asset_added * total_shares / asset_reserve
)
```

### MsgRemoveLiquidity

| Field | Type | Description |
|-------|------|-------------|
| `sender` | AccAddress | LP provider |
| `asset_denom` | string | Pool asset |
| `shares` | Int | Shares to burn |

**Returns:**
```
pnyx_out  = pnyx_reserve  * shares / total_shares
asset_out = asset_reserve * shares / total_shares
```

### DEX Query Endpoints

| Route | ABCI Path | Returns |
|-------|-----------|---------|
| Pool | `custom/dex/pool/{denom}` | Single Pool JSON |
| Pools | `custom/dex/pools` | Array of all pools |

---

## treasury Module

**Path:** `treasury/keeper/`
**Purpose:** Whitepaper tokenomics equations 1-5
**Tests:** 31 unit tests

### Constants

| Constant | Value | Description |
|----------|-------|-------------|
| `CDom` | 2 | Domain creation cost multiplier |
| `CPut` | 15 | PayToPut price cap |
| `CEarn` | 1000 | Reward divisor |
| `StakeMin` | 100,000 | Minimum validator stake (PNYX) |
| `SupplyMax` | 22,000,000 | Maximum PNYX supply |
| `ApyDom` | 0.25 | Domain interest: 25% APY |
| `ApyNode` | 0.10 | Staking reward: 10% APY |
| `SecondsPerYear` | 31,557,600 | 365.25 days |
| `RewardInterval` | 3,600 | Distribution frequency (seconds) |

### Equations

| # | Function | Formula | Description |
|---|----------|---------|-------------|
| 1 | `CalcDomainCost(fee)` | `fee * CDom * CEarn` | Domain creation cost |
| 2 | `CalcReward(treasure)` | `treasure / CEarn` | VoteToEarn reward |
| 3 | `CalcPutPrice(treasure, n)` | `min(reward * CPut, reward * n)` | Proposal submission fee |
| 4 | `CalcDomainInterest(treasure, T, released)` | `treasure * ApyDom * T * (1 - released/SupplyMax)` | Domain treasury interest |
| 5 | `CalcNodeReward(stake, T, released)` | `stake * ApyNode * T * (1 - released/SupplyMax)` | Validator staking reward |

**Release decay:** `factor = 1 - totalReleased / 22,000,000`

As more PNYX enters circulation, rewards decrease proportionally, preventing runaway inflation.

---

## CosmWasm Contracts

**Path:** `contracts/src/`
**Language:** Rust
**Framework:** cosmwasm-std 3.x

### governance.rs

On-chain governance with systemic consensing:
- Submit proposals with -5 to +5 ratings
- Domain key pair validation for anonymous voting
- Proposal queries by domain

### treasury.rs

Treasury management:
- Deposit PNYX to domain treasury
- Withdraw PNYX from treasury
- Balance queries

### Building

```bash
cd contracts
rustup target add wasm32-unknown-unknown
cargo build --release --target wasm32-unknown-unknown
```

### Deploying

```bash
truerepublicd tx wasm store governance.wasm \
    --from wallet --gas auto --fees 10000pnyx

truerepublicd tx wasm instantiate $CODE_ID '{}' \
    --from wallet --label "governance-v1" \
    --admin $(truerepublicd keys show wallet -a) \
    --gas auto --fees 10000pnyx
```

---

## Module Interaction Diagram

```
                    ┌───────────────────┐
                    │   CometBFT        │
                    │   (Consensus)     │
                    └────────┬──────────┘
                             │ ABCI
                    ┌────────┴──────────┐
                    │   BaseApp Router  │
                    └──┬──────┬─────┬───┘
                       │      │     │
          ┌────────────┘      │     └────────────┐
          │                   │                  │
┌─────────┴─────────┐ ┌──────┴──────┐ ┌─────────┴─────────┐
│  x/truedemocracy  │ │   x/dex     │ │  treasury/keeper  │
│  13 msg types     │ │  4 msg types│ │  5 equations      │
│  116 tests        │ │  24 tests   │ │  31 tests         │
│                   │ │             │ │                   │
│  Domains          │ │  Pools      │ │  Domain interest  │
│  Proposals        │ │  Swaps      │ │  Staking rewards  │
│  Stones voting    │ │  Liquidity  │ │  VoteToEarn       │
│  PoD validators   │ │  PNYX burn  │ │  Release decay    │
│  Anonymous voting │ │             │ │                   │
│  Admin election   │ │             │ │                   │
│  Lifecycle zones  │ │             │ │                   │
└───────────────────┘ └─────────────┘ └───────────────────┘
          │                   │                  │
          └───────────────────┴──────────────────┘
                              │
                    ┌─────────┴─────────┐
                    │   KV Store        │
                    │   (IAVL Trees)    │
                    └───────────────────┘
```

---

## Next Steps

- [API Reference](API-Reference) -- Complete endpoint documentation
- [Development Setup](Development-Setup) -- Set up your environment
- [Frontend Architecture](Frontend-Architecture) -- React architecture details
