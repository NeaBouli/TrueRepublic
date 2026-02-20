# ABCI Queries

All module queries use the legacy ABCI querier pattern via CometBFT RPC.

## Query Path Format

```
/custom/{module}/{route}/{params...}
```

## truedemocracy Queries

| Route | Parameters | Returns |
|-------|-----------|---------|
| `custom/truedemocracy/domain/{name}` | Domain name | Single Domain JSON |
| `custom/truedemocracy/domains` | None | Array of all domains |
| `custom/truedemocracy/validator/{addr}` | Operator address | Single Validator JSON |
| `custom/truedemocracy/validators` | None | Array of all validators |

### Domain Response

```json
{
  "name": "Climate",
  "admin": "truerepublic1abc...",
  "members": ["truerepublic1abc...", "truerepublic1def..."],
  "treasury": [{"denom": "pnyx", "amount": "500000"}],
  "issues": [
    {
      "name": "Carbon Reporting",
      "stones": 5,
      "creation_date": 1700000000,
      "last_activity_at": 1700100000,
      "external_link": "https://example.com",
      "suggestions": [
        {
          "name": "Quarterly Reports",
          "creator": "truerepublic1abc...",
          "stones": 3,
          "color": "green",
          "dwell_time": 86400,
          "creation_date": 1700000000,
          "ratings": [
            {"domain_pub_key_hex": "abcdef...", "value": 4},
            {"domain_pub_key_hex": "123456...", "value": -1}
          ],
          "delete_votes": 0
        }
      ]
    }
  ],
  "options": {
    "admin_electable": true,
    "anyone_can_join": true,
    "only_admin_issues": false,
    "coin_burn_required": false,
    "approval_threshold": 500,
    "default_dwell_time": 86400
  },
  "total_payouts": 50000,
  "transferred_stake": 5000
}
```

### Validator Response

```json
{
  "operator_address": "truerepublic1abc...",
  "pub_key": "abcdef1234567890...",
  "stake": [{"denom": "pnyx", "amount": "150000"}],
  "domains": ["Climate", "Tech"],
  "power": 1,
  "jailed": false,
  "jailed_until": 0,
  "missed_blocks": 0
}
```

## dex Queries

| Route | Parameters | Returns |
|-------|-----------|---------|
| `custom/dex/pool/{denom}` | Asset denomination | Single Pool JSON |
| `custom/dex/pools` | None | Array of all pools |

### Pool Response

```json
{
  "asset_denom": "atom",
  "pnyx_reserve": "1000000",
  "asset_reserve": "500000",
  "total_shares": "707106",
  "total_burned": "5000"
}
```

## Querying via RPC

### Using curl

```bash
# Query all domains
curl -s 'http://localhost:26657/abci_query?path="custom/truedemocracy/domains"' \
    | jq -r '.result.response.value' | base64 -d | jq

# Query specific domain
curl -s 'http://localhost:26657/abci_query?path="custom/truedemocracy/domain/Climate"' \
    | jq -r '.result.response.value' | base64 -d | jq

# Query all pools
curl -s 'http://localhost:26657/abci_query?path="custom/dex/pools"' \
    | jq -r '.result.response.value' | base64 -d | jq
```

### Using CosmJS

```javascript
import { SigningStargateClient } from "@cosmjs/stargate";

const client = await SigningStargateClient.connect("http://localhost:26657");

// Query domains
const result = await client.queryAbci(
    "custom/truedemocracy/domains",
    new Uint8Array()
);
const domains = JSON.parse(new TextDecoder().decode(result.value));
```

## REST/LCD Endpoints (Port 1317)

Standard Cosmos SDK REST endpoints:

| Endpoint | Description |
|----------|-------------|
| `GET /cosmos/bank/v1beta1/balances/{address}` | Account balances |
| `GET /cosmos/staking/v1beta1/validators` | Validator set |
| `POST /cosmos/tx/v1beta1/txs` | Broadcast transaction |
| `GET /cosmos/tx/v1beta1/txs/{hash}` | Transaction by hash |
| `GET /node_info` | Node information |
| `GET /syncing` | Sync status |

## CometBFT RPC Endpoints (Port 26657)

| Endpoint | Description |
|----------|-------------|
| `GET /status` | Node status and sync info |
| `GET /block` | Latest block |
| `GET /block?height=N` | Block at specific height |
| `GET /validators` | Current validator set |
| `GET /net_info` | Network information |
| `POST /broadcast_tx_sync` | Broadcast transaction (sync) |
| `POST /broadcast_tx_async` | Broadcast transaction (async) |
| `GET /tx?hash=0x...` | Transaction by hash |
| `GET /abci_query?path=...` | Custom ABCI query |
