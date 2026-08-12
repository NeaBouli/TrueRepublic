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
| create-pool | `truerepublicd tx dex create-pool [asset-denom] [upnyx-amount] [asset-amount]` | Create a PNYX/asset liquidity pool |
| swap | `truerepublicd tx dex swap [input-denom] [input-amount] [output-denom]` | Swap tokens via AMM (0.3% fee, 1% PNYX burn) |
| add-liquidity | `truerepublicd tx dex add-liquidity [asset-denom] [upnyx-amount] [asset-amount]` | Add liquidity and receive LP shares |
| remove-liquidity | `truerepublicd tx dex remove-liquidity [asset-denom] [shares]` | Remove liquidity by burning LP shares |

## CLI Query Commands

### truedemocracy module (9 commands)

| Command | Usage | gRPC method |
|---------|-------|-------------|
| domain | `truerepublicd query truedemocracy domain [name]` | `/truedemocracy.Query/Domain` |
| domains | `truerepublicd query truedemocracy domains` | `/truedemocracy.Query/Domains` |
| validator | `truerepublicd query truedemocracy validator [addr]` | `/truedemocracy.Query/Validator` |
| validators | `truerepublicd query truedemocracy validators` | `/truedemocracy.Query/Validators` |
| nullifier | `truerepublicd query truedemocracy nullifier [domain] [hash]` | `/truedemocracy.Query/Nullifier` |
| purge-schedule | `truerepublicd query truedemocracy purge-schedule [domain]` | `/truedemocracy.Query/PurgeSchedule` |
| zkp-state | `truerepublicd query truedemocracy zkp-state [domain]` | `/truedemocracy.Query/ZKPState` |
| merkle-proof | `truerepublicd query truedemocracy merkle-proof [domain] [commitment]` | `/truedemocracy.Query/MerkleProof` |
| pay-to-put | `truerepublicd query truedemocracy pay-to-put [domain]` | `/truedemocracy.Query/PayToPut` |

### dex module (9 commands)

| Command | Usage | gRPC method |
|---------|-------|-------------|
| pool | `truerepublicd query dex pool [asset-denom]` | `/dex.Query/Pool` |
| pools | `truerepublicd query dex pools` | `/dex.Query/Pools` |
| registered-assets | `truerepublicd query dex registered-assets` | `/dex.Query/RegisteredAssets` |
| asset | `truerepublicd query dex asset [denom-or-symbol]` | `/dex.Query/AssetByDenom` or `/dex.Query/AssetBySymbol` |
| estimate-swap | `truerepublicd query dex estimate-swap [input] [amount] [output]` | `/dex.Query/EstimateSwap` |
| pool-stats | `truerepublicd query dex pool-stats [asset]` | `/dex.Query/PoolStats` |
| spot-price | `truerepublicd query dex spot-price [input] [output]` | `/dex.Query/SpotPrice` |
| liquidity-depth | `truerepublicd query dex liquidity-depth [input] [output]` | `/dex.Query/LiquidityDepth` |
| lp-position | `truerepublicd query dex lp-position [asset] [shares]` | `/dex.Query/LPPosition` |

## Supported module query boundary

The supported interfaces for custom-module reads are the daemon CLI, the
protobuf gRPC services on the private operator gRPC listener, and the maintained
browser client's typed use of those registered methods through CometBFT
`abci_query` on the existing RPC proxy. The former
compatibility-only custom ABCI paths were retired under GH-116 after their last
clients were removed. They are not a supported API.

TrueRepublic does not currently register grpc-gateway HTTP routes for its two
custom modules. Port 1317 continues to expose the registered standard Cosmos
SDK routes only. See the [module query reference](developers/api-reference/abci-queries.md).

## REST/LCD Endpoints (port 1317)

Only REST routes for mounted modules are available. TrueRepublic does not
mount `x/staking` or `x/distribution`; their REST, gRPC, CLI and transaction
surfaces are intentionally absent and their required IBC/CosmWasm adapters
reject fail-closed.

| Endpoint | Description |
|----------|-------------|
| `/cosmos/bank/v1beta1/balances/{address}` | Account balances |
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
| JSON-RPC `abci_query` with `path` and protobuf `data` | SDK/gRPC query transport used by the maintained browser; only documented registered methods are supported |
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
