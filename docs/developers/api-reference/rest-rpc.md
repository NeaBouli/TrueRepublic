# REST & RPC Endpoints

## CometBFT RPC (Port 26657)

The primary interface for queries and transaction broadcasting.

### Node Information

```bash
# Node status
GET /status
curl http://localhost:26657/status

# Network info (peers)
GET /net_info
curl http://localhost:26657/net_info

# Health check
GET /health
curl http://localhost:26657/health
```

### Blocks

```bash
# Latest block
GET /block
curl http://localhost:26657/block

# Block at height
GET /block?height=100
curl http://localhost:26657/block?height=100

# Block results (events)
GET /block_results?height=100
curl http://localhost:26657/block_results?height=100
```

### Transactions

```bash
# Broadcast transaction (wait for check)
POST /broadcast_tx_sync
curl -X POST http://localhost:26657/broadcast_tx_sync \
    -d '{"tx": "<base64-encoded-tx>"}'

# Broadcast transaction (fire and forget)
POST /broadcast_tx_async

# Broadcast transaction (wait for delivery)
POST /broadcast_tx_commit

# Get transaction by hash
GET /tx?hash=0xABCDEF...
curl http://localhost:26657/tx?hash=0xABCDEF...

# Search transactions
GET /tx_search?query="tx.height=100"
```

### Validators

```bash
# Current validator set
GET /validators
curl http://localhost:26657/validators

# Validator set at height
GET /validators?height=100
```

### Module queries

Custom-module reads use the daemon CLI or the registered protobuf gRPC services
documented in [Module Queries](abci-queries.md). The maintained browser sends
those typed requests through CometBFT JSON-RPC `abci_query` on the existing RPC
proxy. The retired compatibility-only custom ABCI routes and unregistered HTTP
aliases are not supported.

## REST/LCD API (Port 1317)

Cosmos SDK's auto-generated REST API.

### Bank Module

```bash
# Account balances
GET /cosmos/bank/v1beta1/balances/{address}
curl http://localhost:1317/cosmos/bank/v1beta1/balances/truerepublic1abc...

# Total supply
GET /cosmos/bank/v1beta1/supply
curl http://localhost:1317/cosmos/bank/v1beta1/supply

# Supply of specific denom
GET /cosmos/bank/v1beta1/supply/by_denom?denom=upnyx
```

### Auth Module

```bash
# Account info (sequence, account number)
GET /cosmos/auth/v1beta1/accounts/{address}
curl http://localhost:1317/cosmos/auth/v1beta1/accounts/truerepublic1abc...
```

### Transaction

```bash
# Get transaction by hash
GET /cosmos/tx/v1beta1/txs/{hash}
curl http://localhost:1317/cosmos/tx/v1beta1/txs/ABCDEF...

# Broadcast transaction
POST /cosmos/tx/v1beta1/txs
curl -X POST http://localhost:1317/cosmos/tx/v1beta1/txs \
    -H "Content-Type: application/json" \
    -d '{"tx_bytes": "<base64>", "mode": "BROADCAST_MODE_SYNC"}'
```

### Node Status

```bash
# Node info
GET /cosmos/base/tendermint/v1beta1/node_info

# Syncing status
GET /cosmos/base/tendermint/v1beta1/syncing

# Latest block
GET /cosmos/base/tendermint/v1beta1/blocks/latest

# Block at height
GET /cosmos/base/tendermint/v1beta1/blocks/{height}
```

## gRPC (Port 9090)

Standard gRPC endpoints mirror their registered REST routes. TrueRepublic's
custom modules expose protobuf gRPC services without grpc-gateway HTTP aliases.
Use `grpcurl` for discovery and testing:

```bash
# List available services
grpcurl -plaintext localhost:9090 list

# Query bank balance
grpcurl -plaintext -d '{"address": "truerepublic1abc..."}' \
    localhost:9090 cosmos.bank.v1beta1.Query/Balance

# Discover the custom module services
grpcurl -plaintext localhost:9090 list truedemocracy.Query
grpcurl -plaintext localhost:9090 list dex.Query
```

## WebSocket (Port 26657)

Subscribe to real-time events via WebSocket:

```bash
# Connect to WebSocket
wscat -c ws://localhost:26657/websocket

# Subscribe to new blocks
{"jsonrpc": "2.0", "method": "subscribe", "id": 1,
 "params": {"query": "tm.event='NewBlock'"}}

# Subscribe to transactions
{"jsonrpc": "2.0", "method": "subscribe", "id": 2,
 "params": {"query": "tm.event='Tx'"}}
```

### JavaScript WebSocket Example

```javascript
const ws = new WebSocket("ws://localhost:26657/websocket");

ws.onopen = () => {
    ws.send(JSON.stringify({
        jsonrpc: "2.0",
        method: "subscribe",
        id: 1,
        params: { query: "tm.event='NewBlock'" }
    }));
};

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    console.log("New block:", data.result.data.value.block.header.height);
};
```
