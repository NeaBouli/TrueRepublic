# Deployment Guide

**Version:** v0.3.0

## Prerequisites

- Go 1.24+ (for native build)
- Docker and Docker Compose (for containerized deployment)
- Rust toolchain (for CosmWasm contracts)
- Node.js 20+ (for web wallet)

## System Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| CPU | 2 cores | 4+ cores |
| RAM | 4 GB | 8+ GB |
| Storage | 100 GB SSD | 500 GB NVMe |
| Network | 100 Mbps | 1 Gbps |
| OS | Ubuntu 22.04 | Ubuntu 22.04/24.04 |

---

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

---

## Option B: Native Build

### 1. Build

```bash
CGO_ENABLED=1 make build
# Binary: ./build/truerepublicd
```

### 2. Initialize Node

```bash
export CHAIN_ID=truerepublic-1
export MONIKER=my-node
./build/truerepublicd init $MONIKER --chain-id $CHAIN_ID
```

### 3. Configure

```bash
# Set minimum gas price
sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0.025pnyx"/' \
  ~/.truerepublic/config/app.toml

# Enable Prometheus metrics
sed -i 's/prometheus = false/prometheus = true/' \
  ~/.truerepublic/config/config.toml
```

### 4. Start

```bash
./build/truerepublicd start
```

---

## Multi-Node Testnet

### Node 1 (Seed)

```bash
./build/truerepublicd init node1 --chain-id truerepublic-testnet
# Note the node ID
./build/truerepublicd tendermint show-node-id
# e.g., abc123def456...
```

### Node 2+

```bash
./build/truerepublicd init node2 --chain-id truerepublic-testnet

# Add seed node
sed -i 's/seeds = ""/seeds = "abc123def456@node1-ip:26656"/' \
  ~/.truerepublic/config/config.toml

# Copy genesis from node1
scp node1:/root/.truerepublic/config/genesis.json \
  ~/.truerepublic/config/genesis.json

./build/truerepublicd start
```

---

## Validator Setup

### Requirements

- Domain membership (must be a member of at least one domain)
- Minimum stake: 100,000 PNYX
- Reliable server (see system requirements)

### Register

```bash
# Create or join a domain first
./build/truerepublicd tx truedemocracy create-domain my-domain "My Domain" \
  --from validator-key

# Register as validator
./build/truerepublicd tx truedemocracy register-validator \
  <pubkey-hex> <stake-amount> <domain-name> \
  --from validator-key
```

See `docs/VALIDATOR_GUIDE.md` for detailed instructions.

---

## IBC Relayer Configuration (Hermes)

### Install Hermes

```bash
cargo install ibc-relayer-cli --version 1.10.0 --bin hermes
```

### Configure

See `docs/IBC_RELAYER_SETUP.md` for complete Hermes configuration including:
- Chain configuration for TrueRepublic + counterparty
- Key management
- Channel creation
- Monitoring

---

## Systemd Service

```ini
# /etc/systemd/system/truerepublicd.service
[Unit]
Description=TrueRepublic Node
After=network.target

[Service]
Type=simple
User=truerepublic
ExecStart=/usr/local/bin/truerepublicd start
Restart=on-failure
RestartSec=10
LimitNOFILE=65535
Environment="CGO_ENABLED=1"

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable truerepublicd
sudo systemctl start truerepublicd
sudo journalctl -u truerepublicd -f
```

---

## Security Hardening

### Firewall (UFW)

```bash
sudo ufw allow 26656/tcp  # P2P
sudo ufw allow 26657/tcp  # RPC (restrict to trusted IPs in production)
sudo ufw allow 443/tcp    # HTTPS
sudo ufw enable
```

### Non-Root User

```bash
sudo useradd -m -s /bin/bash truerepublic
sudo su - truerepublic
```

### Key Management

- Use hardware signing module (HSM) for validator keys in production
- Never expose `priv_validator_key.json` publicly
- Back up mnemonic phrases securely offline

---

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

---

## Monitoring

- **Prometheus** scrapes CometBFT metrics from port 26660
- **Grafana** dashboard at `http://localhost:3000` shows:
  - Block height, connected peers, mempool size
  - Consensus rounds, transactions per block
  - Block interval, missing validators
- Configuration: `monitoring/prometheus.yml`, `monitoring/grafana/`

---

## Backup & Recovery

### Automated Backup

```bash
# Run daily at 3 AM via cron
0 3 * * * /path/to/scripts/backup.sh
```

### Manual Backup

```bash
tar -czf truerepublic_backup.tar.gz ~/.truerepublic
```

### Restore

```bash
tar -xzf truerepublic_backup.tar.gz -C ~/
./build/truerepublicd start
```

---

## Upgrade Procedures

### Binary Upgrade (Non-Breaking)

```bash
# Stop node
sudo systemctl stop truerepublicd

# Replace binary
cp ./build/truerepublicd /usr/local/bin/

# Start node
sudo systemctl start truerepublicd
```

### State Migration (Breaking)

```bash
# Export state at upgrade height
./build/truerepublicd export --height <upgrade-height> > genesis_export.json

# Migrate genesis
./build/truerepublicd-new migrate genesis_export.json --chain-id truerepublic-2 > new_genesis.json

# Reset and restart with new genesis
./build/truerepublicd-new tendermint unsafe-reset-all
cp new_genesis.json ~/.truerepublic/config/genesis.json
sudo systemctl start truerepublicd
```

---

## Troubleshooting

| Issue | Solution |
|-------|----------|
| `CGO_ENABLED` error | Set `CGO_ENABLED=1` and install build-essential |
| Node won't sync | Check seeds/persistent_peers in config.toml |
| Out of memory | Increase RAM or enable swap |
| Port already in use | Check for existing processes: `lsof -i :26657` |
| WAL corruption | `truerepublicd tendermint unsafe-reset-all` (data loss) |
| IBC timeout | Verify relayer is running and channels are open |
