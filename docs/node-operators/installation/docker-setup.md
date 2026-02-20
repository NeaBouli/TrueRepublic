# Docker Setup

The recommended way to run a TrueRepublic node. Uses multi-service Docker Compose with built-in monitoring.

## Prerequisites

- Docker 24.0+ and Docker Compose v2.20+
- Git

## Step 1: Clone and Configure

```bash
git clone https://github.com/NeaBouli/TrueRepublic.git
cd TrueRepublic
cp .env.example .env
```

Edit `.env` with your settings:

```bash
# Required
CHAIN_ID=truerepublic-1
MONIKER=my-truerepublic-node    # Your node's public name

# Network (leave empty for fresh chain, add for joining existing)
SEEDS=
PERSISTENT_PEERS=

# Ports (defaults shown)
P2P_PORT=26656
RPC_PORT=26657
LCD_PORT=1317
GRPC_PORT=9090

# Gas
MIN_GAS_PRICE=0.001pnyx

# Monitoring
PROMETHEUS_ENABLED=true
GRAFANA_PASSWORD=your-secure-password

# Web wallet
RPC_URL=http://truerepublic-node:26657
```

## Step 2: Build

```bash
make docker-build
```

This builds:
- **truerepublic-node** -- Multi-stage Go build (golang:1.23-alpine -> alpine:3.19)
- **web-wallet** -- React build served by nginx

## Step 3: Start

```bash
make docker-up
```

This starts all services:

| Service | Port | Description |
|---------|------|-------------|
| truerepublic-node | 26656, 26657, 1317, 9090 | Blockchain node |
| web-wallet | 3001 | React frontend |
| nginx | 80, 443 | Reverse proxy |
| prometheus | 9091 | Metrics collection |
| grafana | 3000 | Dashboards (admin / your-password) |

## Step 4: Verify

```bash
# Check node status
curl http://localhost:26657/status | jq .result.sync_info

# Check if node is syncing
curl http://localhost:26657/status | jq .result.sync_info.catching_up

# Check web wallet
curl -s http://localhost:3001 | head -5

# Check Grafana
open http://localhost:3000
```

## Step 5: Stop

```bash
make docker-down
```

## Docker Compose Services

### Node Service

```yaml
truerepublic-node:
  build: .
  ports:
    - "${P2P_PORT}:26656"
    - "${RPC_PORT}:26657"
    - "${LCD_PORT}:1317"
    - "${GRPC_PORT}:9090"
  volumes:
    - node-data:/root/.truerepublic
  environment:
    - MONIKER=${MONIKER}
    - CHAIN_ID=${CHAIN_ID}
    - MIN_GAS_PRICE=${MIN_GAS_PRICE}
```

### Data Persistence

Node data is stored in a Docker volume `node-data`. To inspect:

```bash
docker volume inspect truerepublic_node-data
```

To backup:

```bash
docker run --rm -v truerepublic_node-data:/data -v $(pwd):/backup \
    alpine tar czf /backup/node-backup.tar.gz /data
```

## Joining an Existing Network

To join an existing TrueRepublic network:

1. Get the **genesis file** from the network coordinator
2. Get **seed node** addresses
3. Update `.env`:

```bash
SEEDS=node-id@seed1.truerepublic.network:26656
PERSISTENT_PEERS=node-id@peer1.truerepublic.network:26656
```

4. Replace the genesis file:

```bash
# Copy genesis into the container
docker cp genesis.json truerepublic-node:/root/.truerepublic/config/genesis.json
docker restart truerepublic-node
```

## Troubleshooting

### Container won't start

```bash
# Check logs
docker compose logs truerepublic-node

# Common issues:
# - Port already in use: Change ports in .env
# - Genesis mismatch: Ensure correct genesis.json
# - Permission issues: Check volume permissions
```

### Node not syncing

```bash
# Check peer connections
curl http://localhost:26657/net_info | jq .result.n_peers

# Check if seeds are reachable
docker exec truerepublic-node ping seed1.truerepublic.network
```

### Reset node data

```bash
make docker-down
docker volume rm truerepublic_node-data
make docker-up
```

## Next Steps

- [Node Configuration](../configuration/node-config.md) -- Tune your node
- [Monitoring](../operations/monitoring.md) -- Set up alerts
- [Validator Guide](../../validators/README.md) -- Become a validator
