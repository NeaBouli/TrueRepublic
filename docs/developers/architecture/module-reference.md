# Module Reference

Detailed documentation of all TrueRepublic blockchain modules.

## truedemocracy Module

### Message Types (13)

#### MsgCreateDomain
Creates a new governance domain with initial treasury.

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Unique domain name |
| `admin` | AccAddress | Initial admin address |
| `initial_coins` | Coins | Initial treasury funding |

**Cost:** `fee * CDom * CEarn = fee * 2000` (eq.1)

#### MsgSubmitProposal
Submits an issue with a suggestion for evaluation.

| Field | Type | Description |
|-------|------|-------------|
| `sender` | AccAddress | Submitter address |
| `domain_name` | string | Target domain |
| `issue_name` | string | Issue title |
| `suggestion_name` | string | Suggestion title |
| `creator` | AccAddress | Creator address |
| `fee` | Coins | PayToPut fee |
| `external_link` | string | Optional URL |

**Cost:** PayToPut fee = `min(reward * 15, reward * members)` (eq.3)

#### MsgRegisterValidator
Registers as a PoD validator.

| Field | Type | Description |
|-------|------|-------------|
| `sender` | AccAddress | Operator address |
| `operator_addr` | AccAddress | Validator operator |
| `pub_key` | string | Ed25519 public key (hex) |
| `stake` | Coins | Stake amount (min 100,000 PNYX) |
| `domain_name` | string | Domain membership |

#### MsgWithdrawStake
Withdraws staked PNYX (capped at 10% of domain payouts).

| Field | Type | Description |
|-------|------|-------------|
| `sender` | AccAddress | Operator address |
| `operator_addr` | AccAddress | Validator operator |
| `amount` | Coins | Amount to withdraw |

#### MsgRemoveValidator
Removes a validator from the set.

| Field | Type | Description |
|-------|------|-------------|
| `sender` | AccAddress | Operator address |
| `operator_addr` | AccAddress | Validator to remove |

#### MsgUnjail
Unjails a validator after the jail period expires.

| Field | Type | Description |
|-------|------|-------------|
| `sender` | AccAddress | Operator address |
| `operator_addr` | AccAddress | Validator to unjail |

#### MsgJoinPermissionRegister
Registers a domain key pair for anonymous voting (WP S4).

| Field | Type | Description |
|-------|------|-------------|
| `sender` | AccAddress | Member address |
| `domain_name` | string | Domain name |
| `member_addr` | AccAddress | Member address |
| `domain_pub_key` | string | Domain Ed25519 public key (hex) |

#### MsgPurgePermissionRegister
Purges the permission register (admin only).

| Field | Type | Description |
|-------|------|-------------|
| `caller` | AccAddress | Admin address |
| `domain_name` | string | Domain name |

#### MsgPlaceStoneOnIssue
Places a stone on an issue (earns VoteToEarn reward).

| Field | Type | Description |
|-------|------|-------------|
| `sender` | AccAddress | Voter address |
| `domain_name` | string | Domain name |
| `issue_name` | string | Issue name |
| `member_addr` | AccAddress | Member placing stone |

#### MsgPlaceStoneOnSuggestion
Places a stone on a suggestion with systemic consensing.

| Field | Type | Description |
|-------|------|-------------|
| `sender` | AccAddress | Voter address |
| `domain_name` | string | Domain name |
| `issue_name` | string | Issue name |
| `suggestion_name` | string | Suggestion name |
| `member_addr` | AccAddress | Member placing stone |

#### MsgPlaceStoneOnMember
Places a stone on a member for admin election.

| Field | Type | Description |
|-------|------|-------------|
| `sender` | AccAddress | Voter address |
| `domain_name` | string | Domain name |
| `target_member` | AccAddress | Target member |
| `voter_addr` | AccAddress | Voter address |

#### MsgVoteToExclude
Votes to exclude a member (requires 2/3 majority).

| Field | Type | Description |
|-------|------|-------------|
| `sender` | AccAddress | Voter address |
| `domain_name` | string | Domain name |
| `target_member` | AccAddress | Member to exclude |
| `voter_addr` | AccAddress | Voter address |

#### MsgVoteToDelete
Votes to fast-delete a suggestion (requires 2/3 majority).

| Field | Type | Description |
|-------|------|-------------|
| `sender` | AccAddress | Voter address |
| `domain_name` | string | Domain name |
| `issue_name` | string | Issue name |
| `suggestion_name` | string | Suggestion name |
| `member_addr` | AccAddress | Voter address |

---

## dex Module

### Message Types (4)

#### MsgCreatePool
Creates a new PNYX/asset liquidity pool.

| Field | Type | Description |
|-------|------|-------------|
| `sender` | AccAddress | Pool creator |
| `asset_denom` | string | Asset denomination |
| `pnyx_amt` | Int | PNYX amount |
| `asset_amt` | Int | Asset amount |

Initial shares = `sqrt(pnyx_amt * asset_amt)`

#### MsgSwap
Swaps tokens via the AMM (0.3% fee, 1% PNYX burn).

| Field | Type | Description |
|-------|------|-------------|
| `sender` | AccAddress | Trader address |
| `input_denom` | string | Input token denom |
| `input_amt` | Int | Input amount |
| `output_denom` | string | Output token denom |

Output formula: `(outReserve * input * 9970) / (inReserve * 10000 + input * 9970)`

#### MsgAddLiquidity
Adds liquidity and receives LP shares.

| Field | Type | Description |
|-------|------|-------------|
| `sender` | AccAddress | LP provider |
| `asset_denom` | string | Pool asset denom |
| `pnyx_amt` | Int | PNYX to add |
| `asset_amt` | Int | Asset to add |

#### MsgRemoveLiquidity
Removes liquidity by burning LP shares.

| Field | Type | Description |
|-------|------|-------------|
| `sender` | AccAddress | LP provider |
| `asset_denom` | string | Pool asset denom |
| `shares` | Int | Shares to burn |

---

## treasury Module

### Equations

| Equation | Function | Formula |
|----------|----------|---------|
| eq.1 | `CalcDomainCost(fee)` | `fee * CDom * CEarn` (= fee * 2000) |
| eq.2 | `CalcReward(treasure)` | `treasure / CEarn` (= treasure / 1000) |
| eq.3 | `CalcPutPrice(treasure, n)` | `min(reward * CPut, reward * n)` |
| eq.4 | `CalcDomainInterest(...)` | `treasure * 0.25 * T * decay` |
| eq.5 | `CalcNodeReward(...)` | `stake * 0.10 * T * decay` |

Where `decay = 1 - totalReleased / 21,000,000`

### Constants

| Constant | Value | Description |
|----------|-------|-------------|
| `CDom` | 2 | Domain creation multiplier |
| `CPut` | 15 | Put price cap |
| `CEarn` | 1000 | Reward divisor |
| `StakeMin` | 100,000 | Minimum validator stake |
| `SupplyMax` | 21,000,000 | Maximum PNYX supply |
| `ApyDom` | 0.25 | 25% domain interest APY |
| `ApyNode` | 0.10 | 10% staking reward APY |
| `SecondsPerYear` | 31,557,600 | 365.25 days |
