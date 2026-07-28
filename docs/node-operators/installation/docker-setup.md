# Docker Setup

This Compose stack is a loopback-bound recovery/development setup with built-in
monitoring. It is not a production or public-network topology.

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

# Network topology is edited in the initialized config.toml and then validated.

# Local-only P2P host port
P2P_PORT=26656

# Gas
MIN_GAS_PRICE=1000upnyx

# Monitoring
PROMETHEUS_ENABLED=true
GRAFANA_PASSWORD=your-secure-password

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
| truerepublic-node | 127.0.0.1:26656 and proxy port 8080 | Local blockchain node |
| web-wallet | 127.0.0.1:3001 | Deprecated local frontend |
| nginx | 127.0.0.1:8080 | Local HTTP reverse proxy; not a TLS rollout proxy |
| prometheus | 127.0.0.1:9091 | Same-namespace local metrics collection |
| grafana | 127.0.0.1:3000 | Local dashboards |

## Step 4: Verify

```bash
# Check node status
curl http://localhost:8080/rpc/status | jq .result.sync_info

# Check if node is syncing
curl http://localhost:8080/rpc/status | jq .result.sync_info.catching_up

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
    - "127.0.0.1:${P2P_PORT}:26656"
    - "127.0.0.1:8080:80"
  volumes:
    - node-data:/root/.truerepublic
  environment:
    - MONIKER=${MONIKER}
    - CHAIN_ID=${CHAIN_ID}
    - MIN_GAS_PRICE=${MIN_GAS_PRICE}
```

Do not change these mappings to wildcard/public bindings. A rollout candidate
must use an explicit node role, pass the
[network-policy validator](../configuration/network-policy.md), and place any
public query traffic behind the reviewed TLS/rate-limit boundary.

The native `scripts/start-node.sh` wrapper requires `NETWORK_ROLE` and rejects
`SEEDS`/`PERSISTENT_PEERS` environment substitution. This prevents shell data
from silently rewriting reviewed TOML.

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
3. Stop the node and edit the initialized
   `/home/truerepublic/.truerepublic/config/config.toml` in the named volume
   through the reviewed offline change procedure:

```toml
seeds = "<canonical-seed-endpoints>"
persistent_peers = "<canonical-persistent-peer-endpoints>"
```

4. Install the checksum-verified genesis artifact into that same stopped volume
   without changing the non-root ownership or permissions.
5. Validate the selected role with `truerepublicd network-policy validate`
   against the candidate home before restart.

Do not inject peer topology through `.env`, copy a genesis over a running
container, or use the obsolete `/root/.truerepublic` path.

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
