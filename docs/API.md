# API Reference

## CLI Transaction Commands

### truedemocracy module (13 commands)

| Command | Usage | Description |
|---------|-------|-------------|
| create-domain | `truerepublicd tx truedemocracy create-domain [name] [initial-coins]` | Create a new domain with initial treasury |
| submit-proposal | `truerepublicd tx truedemocracy submit-proposal [domain] [issue] [suggestion] [fee] [external-link]` | Submit a proposal (issue + suggestion) |
| register-validator | `truerepublicd tx truedemocracy register-validator [pubkey-hex] [stake] [domain]` | Register as a PoD validator |
| withdraw-stake | `truerepublicd tx truedemocracy withdraw-stake [amount]` | Withdraw staked PNYX (10% transfer limit) |
| remove-validator | `truerepublicd tx truedemocracy remove-validator [operator-addr]` | Remove a validator |
| unjail | `truerepublicd tx truedemocracy unjail` | Unjail validator after jail period expires |
| join-permission-register | `truerepublicd tx truedemocracy join-permission-register [domain] [domain-pubkey-hex]` | Register domain key for anonymous voting |
| purge-permission-register | `truerepublicd tx truedemocracy purge-permission-register [domain]` | Purge the permission register (admin only) |
| place-stone-issue | `truerepublicd tx truedemocracy place-stone-issue [domain] [issue]` | Place a stone on an issue |
| place-stone-suggestion | `truerepublicd tx truedemocracy place-stone-suggestion [domain] [issue] [suggestion]` | Place a stone on a suggestion |
| place-stone-member | `truerepublicd tx truedemocracy place-stone-member [domain] [target-member]` | Place a stone on a member (admin election) |
| vote-exclude | `truerepublicd tx truedemocracy vote-exclude [domain] [target-member]` | Vote to exclude a member (2/3 majority required) |
| vote-delete | `truerepublicd tx truedemocracy vote-delete [domain] [issue] [suggestion]` | Vote to fast-delete a suggestion (2/3 majority) |

### dex module (4 commands)

| Command | Usage | Description |
|---------|-------|-------------|
| create-pool | `truerepublicd tx dex create-pool [asset-denom] [pnyx-amount] [asset-amount]` | Create a PNYX/asset liquidity pool |
| swap | `truerepublicd tx dex swap [input-denom] [input-amount] [output-denom]` | Swap tokens via AMM (0.3% fee, 1% PNYX burn) |
| add-liquidity | `truerepublicd tx dex add-liquidity [asset-denom] [pnyx-amount] [asset-amount]` | Add liquidity and receive LP shares |
| remove-liquidity | `truerepublicd tx dex remove-liquidity [asset-denom] [shares]` | Remove liquidity by burning LP shares |

## CLI Query Commands

### truedemocracy module (4 commands)

| Command | Usage | ABCI Path |
|---------|-------|-----------|
| domain | `truerepublicd query truedemocracy domain [name]` | `custom/truedemocracy/domain/{name}` |
| domains | `truerepublicd query truedemocracy domains` | `custom/truedemocracy/domains` |
| validator | `truerepublicd query truedemocracy validator [addr]` | `custom/truedemocracy/validator/{addr}` |
| validators | `truerepublicd query truedemocracy validators` | `custom/truedemocracy/validators` |

### dex module (2 commands)

| Command | Usage | ABCI Path |
|---------|-------|-----------|
| pool | `truerepublicd query dex pool [asset-denom]` | `custom/dex/pool/{denom}` |
| pools | `truerepublicd query dex pools` | `custom/dex/pools` |

## ABCI Query Paths

All module queries use the legacy ABCI querier pattern:

```
/custom/{module}/{route}/{params...}
```

### truedemocracy routes

| Route | Parameters | Returns |
|-------|-----------|---------|
| `domain/{name}` | Domain name | Single Domain JSON |
| `domains` | None | Array of all domains |
| `validator/{addr}` | Operator address | Single Validator JSON |
| `validators` | None | Array of all validators |

### dex routes

| Route | Parameters | Returns |
|-------|-----------|---------|
| `pool/{denom}` | Asset denomination | Single Pool JSON |
| `pools` | None | Array of all pools |

## REST/LCD Endpoints (port 1317)

Standard Cosmos SDK REST endpoints are available:

| Endpoint | Description |
|----------|-------------|
| `/cosmos/bank/v1beta1/balances/{address}` | Account balances |
| `/cosmos/staking/v1beta1/validators` | Validator set |
| `/cosmos/tx/v1beta1/txs` | Transaction search |
| `/node_info` | Node information |
| `/syncing` | Sync status |

## RPC Endpoints (port 26657)

Standard CometBFT RPC:

| Endpoint | Description |
|----------|-------------|
| `/status` | Node status and sync info |
| `/block` | Latest block |
| `/block?height=N` | Block at height N |
| `/validators` | Current validator set |
| `/abci_query?path=...&data=...` | Custom ABCI query |
| `/broadcast_tx_sync` | Broadcast transaction |
| `/tx?hash=0x...` | Transaction by hash |

## Data Types

### Domain
```json
{
  "name": "string",
  "admin": "truerepublic1...",
  "members": ["truerepublic1..."],
  "treasury": "1000000",
  "issues": [{ "name": "string", "stones": 0, "suggestions": [...] }],
  "options": { "adminElection": true, "joinRule": "open" }
}
```

### Pool
```json
{
  "denom": "atom",
  "pnyx_reserve": "1000000",
  "asset_reserve": "500000",
  "total_shares": "707106",
  "cumulative_burned_pnyx": "1000"
}
```

### Validator
```json
{
  "operator_address": "truerepublic1...",
  "pub_key": "hex-encoded-ed25519",
  "stake": "100000",
  "domains": ["domain-name"],
  "power": 1,
  "jailed": false
}
```
