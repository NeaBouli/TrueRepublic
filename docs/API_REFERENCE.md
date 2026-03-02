# API Reference

**Version:** v0.3.0

## Overview

TrueRepublic exposes APIs via:
- **CLI:** `truerepublicd` binary (Cobra commands)
- **RPC:** CometBFT RPC on port 26657
- **REST:** LCD/API on port 1317
- **gRPC:** On port 9090
- **CosmWasm:** Custom query/message bindings

---

## x/truedemocracy

### Transaction Messages (23 types)

#### Domain Management

| Message | CLI Command | Description |
|---------|-------------|-------------|
| `MsgCreateDomain` | `tx truedemocracy create-domain` | Create a new domain |
| `MsgDeleteDomain` | `tx truedemocracy delete-domain` | Delete a domain (admin only) |
| `MsgJoinDomain` | `tx truedemocracy join-domain` | Join an existing domain |
| `MsgLeaveDomain` | `tx truedemocracy leave-domain` | Leave a domain |
| `MsgUpdateDomainOptions` | `tx truedemocracy update-domain-options` | Update domain settings |

#### Issues & Suggestions

| Message | CLI Command | Description |
|---------|-------------|-------------|
| `MsgCreateIssue` | `tx truedemocracy create-issue` | Create an issue in a domain |
| `MsgCreateSuggestion` | `tx truedemocracy create-suggestion` | Create a suggestion for an issue |
| `MsgDeleteSuggestion` | `tx truedemocracy delete-suggestion` | Delete a suggestion (2/3 vote) |

#### Voting

| Message | CLI Command | Description |
|---------|-------------|-------------|
| `MsgPlaceStone` | `tx truedemocracy place-stone` | Place a stone on a suggestion |
| `MsgRateProposal` | `tx truedemocracy rate-proposal` | Rate with domain key signature |
| `MsgRateWithProof` | `tx truedemocracy rate-with-proof` | Rate with ZKP (anonymous) |
| `MsgCastElectionVote` | `tx truedemocracy cast-election-vote` | Vote in person election |

#### Governance

| Message | CLI Command | Description |
|---------|-------------|-------------|
| `MsgExcludeMember` | `tx truedemocracy exclude-member` | Exclude a member (2/3 vote) |
| `MsgStartElection` | `tx truedemocracy start-election` | Start admin election |
| `MsgTallyElection` | `tx truedemocracy tally-election` | Tally election results |

#### Validator

| Message | CLI Command | Description |
|---------|-------------|-------------|
| `MsgRegisterValidator` | `tx truedemocracy register-validator` | Register as PoD validator |
| `MsgUnregisterValidator` | `tx truedemocracy unregister-validator` | Unregister validator |

#### ZKP

| Message | CLI Command | Description |
|---------|-------------|-------------|
| `MsgRegisterIdentity` | `tx truedemocracy register-identity` | Register identity commitment |
| `MsgRegisterDomainKey` | `tx truedemocracy register-domain-key` | Register domain key pair |

#### Treasury Bridge

| Message | CLI Command | Description |
|---------|-------------|-------------|
| `MsgDepositToDomain` | `tx truedemocracy deposit-to-domain` | Deposit PNYX to domain treasury |
| `MsgWithdrawFromDomain` | `tx truedemocracy withdraw-from-domain` | Withdraw from domain treasury |

### Query Endpoints (7 types)

| Query | CLI Command | Description |
|-------|-------------|-------------|
| `QueryDomain` | `query truedemocracy domain` | Get domain details |
| `QueryDomainMembers` | `query truedemocracy members` | List domain members |
| `QueryIssue` | `query truedemocracy issue` | Get issue details |
| `QuerySuggestion` | `query truedemocracy suggestion` | Get suggestion details |
| `QueryPurgeSchedule` | `query truedemocracy purge-schedule` | Get Big Purge schedule |
| `QueryNullifier` | `query truedemocracy nullifier` | Check nullifier status |
| `QueryZKPState` | `query truedemocracy zkp-state` | Get ZKP verification state |

---

## x/dex

### Transaction Messages (6 types)

| Message | CLI Command | Description |
|---------|-------------|-------------|
| `MsgCreatePool` | `tx dex create-pool` | Create liquidity pool |
| `MsgSwap` | `tx dex swap` | Swap tokens |
| `MsgAddLiquidity` | `tx dex add-liquidity` | Add liquidity to pool |
| `MsgRemoveLiquidity` | `tx dex remove-liquidity` | Remove liquidity from pool |
| `MsgRegisterAsset` | `tx dex register-asset` | Register IBC asset |
| `MsgUpdateAssetStatus` | `tx dex update-asset-status` | Enable/disable asset trading |

### Query Endpoints (5 types)

| Query | CLI Command | Description |
|-------|-------------|-------------|
| `QueryPool` | `query dex pool` | Get pool details |
| `QueryPools` | `query dex pools` | List all pools |
| `QueryRegisteredAssets` | `query dex registered-assets` | List registered assets |
| `QueryAssetByDenom` | `query dex asset` | Get asset by denom |
| `QueryAssetBySymbol` | `query dex asset-by-symbol` | Get asset by symbol |

### AMM Parameters

| Parameter | Value | Description |
|-----------|-------|-------------|
| SwapFeeBps | 30 | 0.3% swap fee |
| BurnBps | 100 | 1% PNYX burn on output |
| MinLiquidity | 1000 | Minimum pool liquidity |

---

## CLI Examples

### Domain Operations

```bash
# Create domain
truerepublicd tx truedemocracy create-domain \
  my-domain "My Domain Description" \
  --from alice --chain-id truerepublic-1

# Join domain
truerepublicd tx truedemocracy join-domain my-domain \
  --from bob --chain-id truerepublic-1

# Query domain
truerepublicd query truedemocracy domain my-domain
```

### DEX Operations

```bash
# Register asset
truerepublicd tx dex register-asset \
  ibc/BTC "BTC" "Bitcoin" 8 cosmoshub-4 channel-0 \
  --from admin

# Create pool
truerepublicd tx dex create-pool pnyx ibc/BTC 1000000 10000 \
  --from alice

# Swap (accepts symbol or denom)
truerepublicd tx dex swap pool-0 pnyx 1000 500 \
  --from alice

# Query pool
truerepublicd query dex pool pool-0
```

### ZKP Voting

```bash
# Register identity commitment
truerepublicd tx truedemocracy register-identity \
  my-domain <commitment-hex> \
  --from alice

# Submit anonymous vote with proof
truerepublicd tx truedemocracy rate-with-proof \
  my-domain issue-1 suggestion-1 \
  <proof-hex> <nullifier-hex> <rating> \
  --from alice
```

### Treasury Bridge

```bash
# Deposit to domain treasury
truerepublicd tx truedemocracy deposit-to-domain \
  my-domain 1000pnyx \
  --from alice

# Withdraw from domain treasury
truerepublicd tx truedemocracy withdraw-from-domain \
  my-domain 500pnyx \
  --from admin
```

---

## CosmWasm Custom Bindings

### Custom Queries (7 types)

Contracts can query chain state via `TrueRepublicQuery`:

```rust
use truerepublic_bindings::{TrueRepublicQuery, DomainResponse};

// Query domain info
let query = TrueRepublicQuery::Domain {
    name: "my-domain".to_string(),
};
let response: DomainResponse = deps.querier.query(&query.into())?;
```

| Query | Response Type | Fields |
|-------|---------------|--------|
| `Domain { name }` | `DomainResponse` | name, admin, member_count, treasury, issue_count, merkle_root, total_payouts, options |
| `DomainMembers { domain_name }` | `DomainMembersResponse` | domain_name, members |
| `Issue { domain_name, issue_name }` | `IssueResponse` | name, stones, suggestion_count, suggestions, creation_date, external_link |
| `Suggestion { domain_name, issue_name, suggestion_name }` | `SuggestionResponse` | name, creator, stones, color, rating_count, score, dwell_time, creation_date, external_link, delete_votes |
| `PurgeSchedule { domain_name }` | `PurgeScheduleResponse` | domain_name, next_purge_time, purge_interval, announcement_lead |
| `Nullifier { domain_name, nullifier_hex }` | `NullifierResponse` | used |
| `DomainTreasury { domain_name }` | `DomainTreasuryResponse` | domain_name, amount |

### Custom Messages (5 types)

Contracts can execute chain actions via `TrueRepublicMsg`:

```rust
use truerepublic_bindings::TrueRepublicMsg;

// Place stone on suggestion
let msg = TrueRepublicMsg::PlaceStoneOnSuggestion {
    domain_name: "my-domain".to_string(),
    issue_name: "issue-1".to_string(),
    suggestion_name: "suggestion-1".to_string(),
    stone_type: "green".to_string(),
};
let cosmos_msg: CosmosMsg<TrueRepublicMsg> = msg.into();
```

| Message | Fields |
|---------|--------|
| `PlaceStoneOnIssue` | domain_name, issue_name, stone_type |
| `PlaceStoneOnSuggestion` | domain_name, issue_name, suggestion_name, stone_type |
| `CastElectionVote` | domain_name, candidate, vote_type |
| `DepositToDomain` | domain_name, amount |
| `WithdrawFromDomain` | domain_name, amount |

---

## REST API Examples

```bash
# Query domain
curl http://localhost:1317/truerepublic/truedemocracy/domain/my-domain

# Query pool
curl http://localhost:1317/truerepublic/dex/pool/pool-0

# Query registered assets
curl http://localhost:1317/truerepublic/dex/registered-assets

# Node status
curl http://localhost:26657/status

# Latest block
curl http://localhost:26657/block
```

---

## Error Codes

| Code | Description |
|------|-------------|
| `ErrInvalidRequest` | Malformed request parameters |
| `ErrUnauthorized` | Sender not authorized for action |
| `ErrInsufficientFunds` | Not enough PNYX for operation |
| `ErrDomainNotFound` | Domain does not exist |
| `ErrAlreadyMember` | User already in domain |
| `ErrNotMember` | User not a member of domain |
| `ErrInvalidProof` | ZKP proof verification failed |
| `ErrNullifierUsed` | Nullifier already consumed |
| `ErrPoolNotFound` | DEX pool does not exist |
| `ErrInsufficientLiquidity` | Not enough liquidity for swap |
| `ErrSlippageExceeded` | Output below minimum amount |
| `ErrAssetNotRegistered` | Asset not in registry |
| `ErrTradingDisabled` | Asset trading is disabled |
