# Deployment Guide

## Prerequisites

- Go 1.23.5+ (for native build)
- Docker and Docker Compose (for containerized deployment)
- (Optional) Rust toolchain for CosmWasm contracts
- (Optional) Node.js 20+ for web wallet development

## Option A: Docker Deployment (Recommended)

### 1. Configure

```bash
cp .env.example .env
# Edit .env: set MONIKER, CHAIN_ID, GRAFANA_PASSWORD
```

### 2. Build and Start

```bash
make docker-build
make docker-up
```

### 3. Verify

```bash
# Check node status
curl http://localhost:26657/status

# Check web wallet
curl http://localhost:3001

# Check Grafana
open http://localhost:3000  # admin / <GRAFANA_PASSWORD>

# Check Prometheus targets
open http://localhost:9091/targets
```

### 4. Stop

```bash
make docker-down
```

## Option B: Native Build

### 1. Build

```bash
make build
# Binary: ./build/truerepublicd
```

### 2. Initialize Node

```bash
export CHAIN_ID=truerepublic-1
export MONIKER=my-node
./scripts/init-node.sh
```

### 3. Start

```bash
./scripts/start-node.sh
```

## Validator Setup

### Requirements

- Domain membership (must be a member of at least one domain)
- Minimum stake: 100,000 PNYX
- Reliable server (2+ CPU, 4GB+ RAM, 100GB+ SSD)

### Register

```bash
# Create or join a domain first
truerepublicd tx truedemocracy create-domain my-domain 200000pnyx

# Register as validator
truerepublicd tx truedemocracy register-validator \
    <pubkey-hex> <stake-amount> <domain-name>
```

See `docs/VALIDATOR_GUIDE.md` for detailed instructions.

## Services & Ports

| Service | Port | Description |
|---------|------|-------------|
| Node P2P | 26656 | Peer-to-peer networking |
| Node RPC | 26657 | CometBFT RPC |
| Node LCD | 1317 | REST API |
| Node gRPC | 9090 | gRPC endpoint |
| Node Metrics | 26660 | Prometheus metrics |
| Web Wallet | 3001 | React frontend |
| Nginx | 80/443 | Reverse proxy |
| Prometheus | 9091 | Metrics collection |
| Grafana | 3000 | Dashboards |

## Monitoring

- **Prometheus** scrapes CometBFT metrics from port 26660
- **Grafana** dashboard at `http://localhost:3000` shows:
  - Block height, connected peers, mempool size
  - Consensus rounds, transactions per block
  - Block interval, missing validators
- Configuration: `monitoring/prometheus.yml`, `monitoring/grafana/`

## Backup & Recovery

### Automated Backup

```bash
# Run daily at 3 AM via cron
0 3 * * * /path/to/scripts/backup.sh

# Or run manually
./scripts/backup.sh /path/to/backup/dir
```

Backups are retained for 30 days. Configure `rclone` in the script for remote storage.

### Manual Backup

```bash
tar -czf truerepublic_backup.tar.gz ~/.truerepublic
```

### Restore

```bash
tar -xzf truerepublic_backup.tar.gz -C ~/
./scripts/start-node.sh
```

## Firewall (UFW)

```bash
sudo ufw allow 26656/tcp  # P2P
sudo ufw allow 26657/tcp  # RPC
sudo ufw allow 443/tcp    # HTTPS
sudo ufw enable
```
