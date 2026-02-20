# Node Setup Guide

Complete guide to deploying a TrueRepublic full node.

## Table of Contents

1. [Requirements](#requirements)
2. [Docker Setup](#docker-setup)
3. [Native Setup](#native-setup)
4. [Configuration](#configuration)
5. [Starting the Node](#starting-the-node)
6. [Verification](#verification)
7. [Backup & Recovery](#backup--recovery)
8. [Upgrading](#upgrading)

---

## Requirements

### Minimum Hardware

- **CPU:** 4 cores
- **RAM:** 8 GB
- **Storage:** 500 GB SSD
- **Network:** 100 Mbps up/down
- **OS:** Ubuntu 20.04+ (or Docker)

### Recommended Hardware

- **CPU:** 8 cores
- **RAM:** 16 GB
- **Storage:** 1 TB NVMe SSD
- **Network:** 1 Gbps up/down
- **OS:** Ubuntu 22.04 LTS

### Why These Specs?

**CPU:** Consensus + state transitions are CPU-intensive
**RAM:** Block processing + mempool require memory
**Storage:** Blockchain grows ~50 GB/month
**Network:** P2P sync + block propagation

### Software Requirements

- Docker 20.10+ (for Docker setup)
- Go 1.23+ (for native setup)
- Git 2.30+

---

## Docker Setup

**Recommended for:** Most users, easy maintenance

### Step 1: Install Docker

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install dependencies
sudo apt install -y curl git

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add user to docker group
sudo usermod -aG docker $USER

# Install Docker Compose
sudo apt install -y docker-compose

# Verify
docker --version
docker-compose --version
```

**Logout and login** for group changes to take effect.

### Step 2: Clone Repository

```bash
cd ~
git clone https://github.com/NeaBouli/TrueRepublic.git
cd TrueRepublic
```

### Step 3: Configure Environment

```bash
# Copy example environment file
cp .env.example .env

# Edit configuration
nano .env
```

**Required settings:**
```bash
# Node configuration
MONIKER=my-node-name
EXTERNAL_IP=YOUR_SERVER_IP

# Chain configuration
CHAIN_ID=truerepublic-1

# Network
P2P_PORT=26656
RPC_PORT=26657
REST_PORT=1317
GRPC_PORT=9090

# Database
DB_BACKEND=goleveldb

# Pruning (saves disk space)
PRUNING=default
PRUNING_KEEP_RECENT=100
PRUNING_INTERVAL=10

# Monitoring
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=26660
```

**Optional settings:**
```bash
# Snapshot download (fast sync)
SNAPSHOT_URL=https://snapshots.truerepublic.network/latest.tar.gz

# State sync (ultra-fast sync)
STATE_SYNC_ENABLED=true
STATE_SYNC_RPC_SERVERS=rpc1.truerepublic.network:26657,rpc2.truerepublic.network:26657

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

### Step 4: Build Images

```bash
make docker-build
```

This builds:
- truerepublic-node (blockchain)
- prometheus (metrics)
- grafana (dashboards)
- nginx (reverse proxy)

**Time:** ~5 minutes

### Step 5: Start Node

```bash
# Start all services
make docker-up

# Or with docker-compose directly:
docker-compose up -d
```

**Services started:**
- truerepublic-node (blockchain node)
- prometheus (port 9090)
- grafana (port 3000)
- nginx (port 80)

### Step 6: Check Logs

```bash
# Follow node logs
docker-compose logs -f truerepublic-node

# Check all services
docker-compose ps

# Last 100 lines
docker-compose logs --tail=100 truerepublic-node
```

**What to look for:**
```
✅ "Executed block" - blocks being processed
✅ "Committed state" - state being saved
✅ "Indexed block" - block indexed for queries
❌ "Connection refused" - check network/firewall
❌ "Out of memory" - increase RAM or adjust cache
```

---

## Native Setup

**Recommended for:** Developers, advanced users, performance

### Step 1: Install Go

```bash
# Download Go 1.23
cd ~
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz

# Remove old Go (if exists)
sudo rm -rf /usr/local/go

# Extract
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz

# Add to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify
go version
# Should show: go version go1.23.0 linux/amd64
```

### Step 2: Clone and Build

```bash
# Clone repository
cd ~
git clone https://github.com/NeaBouli/TrueRepublic.git
cd TrueRepublic

# Install dependencies
go mod download

# Build binary
make build

# Install to system
sudo cp build/truerepublicd /usr/local/bin/

# Verify
truerepublicd version
```

### Step 3: Initialize Node

```bash
# Initialize node
truerepublicd init my-node-name --chain-id truerepublic-1

# This creates:
# ~/.truerepublic/config/config.toml
# ~/.truerepublic/config/app.toml
# ~/.truerepublic/config/genesis.json
# ~/.truerepublic/data/
```

### Step 4: Download Genesis

```bash
# Download genesis file
cd ~/.truerepublic/config
wget https://raw.githubusercontent.com/NeaBouli/TrueRepublic/main/genesis.json

# Verify checksum
sha256sum genesis.json
# Should match official checksum
```

### Step 5: Configure Node

Edit `config.toml`:

```bash
nano ~/.truerepublic/config/config.toml
```

Key settings:

```toml
[p2p]
# Your node's external address
external_address = "tcp://YOUR_IP:26656"

# Seed nodes (initial peers)
seeds = "seed1@seed1.truerepublic.network:26656,seed2@seed2.truerepublic.network:26656"

# Persistent peers (always connect)
persistent_peers = ""

# Maximum number of peers
max_num_inbound_peers = 40
max_num_outbound_peers = 10

[consensus]
# Block time
timeout_commit = "5s"

[mempool]
# Mempool size
size = 5000
cache_size = 10000

[rpc]
# Enable RPC
laddr = "tcp://0.0.0.0:26657"

[instrumentation]
# Enable Prometheus metrics
prometheus = true
prometheus_listen_addr = ":26660"
```

Edit `app.toml`:

```bash
nano ~/.truerepublic/config/app.toml
```

Key settings:

```toml
[api]
# Enable REST API
enable = true
address = "tcp://0.0.0.0:1317"

[grpc]
# Enable gRPC
enable = true
address = "0.0.0.0:9090"

[state-sync]
# Snapshots for state sync
snapshot-interval = 1000
snapshot-keep-recent = 2

[pruning]
# Pruning strategy: default, nothing, everything, custom
pruning = "default"
pruning-keep-recent = "100"
pruning-interval = "10"
```

### Step 6: Create Systemd Service

```bash
sudo nano /etc/systemd/system/truerepublicd.service
```

Service file:

```ini
[Unit]
Description=TrueRepublic Node
After=network-online.target

[Service]
User=YOUR_USERNAME
ExecStart=/usr/local/bin/truerepublicd start
Restart=always
RestartSec=3
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service
sudo systemctl enable truerepublicd

# Start service
sudo systemctl start truerepublicd

# Check status
sudo systemctl status truerepublicd

# View logs
sudo journalctl -u truerepublicd -f
```

---

## Configuration

### Pruning Strategies

**default:**
- Keep recent: 100 blocks
- Interval: 10 blocks
- Saves ~60% disk space

**nothing:**
- Keep all blocks
- Archive node
- Requires ~500 GB+

**everything:**
- Keep only latest state
- Minimal disk usage
- Can't serve historical queries

**custom:**
```toml
pruning = "custom"
pruning-keep-recent = "500"
pruning-interval = "50"
```

### State Sync (Fast Sync)

Enable state sync for ultra-fast sync:

```toml
[statesync]
enable = true

rpc_servers = "rpc1.truerepublic.network:26657,rpc2.truerepublic.network:26657"

# Get trust height and hash from RPC:
# curl -s http://rpc1.truerepublic.network:26657/block | jq -r '.result.block.header.height + "," + .result.block_id.hash'

trust_height = 1234567
trust_hash = "ABC123..."
```

**Advantages:**
- Sync in minutes (not hours)
- Downloads state snapshot
- Skips historical blocks

**Disadvantages:**
- Trust RPC servers
- Won't have full history

### Firewall Configuration

Open required ports:

```bash
# P2P
sudo ufw allow 26656/tcp

# RPC (if needed publicly)
sudo ufw allow 26657/tcp

# REST API (if needed publicly)
sudo ufw allow 1317/tcp

# gRPC (if needed publicly)
sudo ufw allow 9090/tcp

# Prometheus (local only)
sudo ufw allow from 127.0.0.1 to any port 26660

# Enable firewall
sudo ufw enable
```

---

## Starting the Node

### Docker

```bash
# Start
make docker-up

# Or
docker-compose up -d

# Restart
docker-compose restart truerepublic-node

# Stop
docker-compose stop

# Stop and remove
docker-compose down
```

### Native

```bash
# Using systemd
sudo systemctl start truerepublicd
sudo systemctl stop truerepublicd
sudo systemctl restart truerepublicd

# Or directly (for testing)
truerepublicd start

# With custom home directory
truerepublicd start --home /path/to/data
```

---

## Verification

### Check Sync Status

```bash
# Using curl
curl -s localhost:26657/status | jq .result.sync_info

# Output:
{
  "latest_block_hash": "ABC123...",
  "latest_block_height": "1234567",
  "latest_block_time": "2025-02-20T10:30:00Z",
  "catching_up": false
}
```

`catching_up` meanings:
- `true` -- Still syncing, be patient
- `false` -- Fully synced!

### Check Peers

```bash
curl -s localhost:26657/net_info | jq .result.n_peers

# Should show: 10-40 peers
```

If 0 peers:
1. Check firewall (port 26656)
2. Check seeds in config.toml
3. Check external_address setting
4. Wait a few minutes

### Check Block Production

```bash
# Watch logs for new blocks
docker-compose logs -f truerepublic-node | grep "Executed block"

# Or with systemd
sudo journalctl -u truerepublicd -f | grep "Executed block"

# Should see new blocks every ~5 seconds
```

### Query Chain Data

```bash
# Get latest block
truerepublicd query block

# Get node info
truerepublicd status

# Get account balance
truerepublicd query bank balances cosmos1abc...

# Query domains
truerepublicd query truedemocracy domains
```

---

## Backup & Recovery

### What to Backup

**Critical:**
- Node key: `~/.truerepublic/config/node_key.json`
- Validator key: `~/.truerepublic/config/priv_validator_key.json` (if validator)
- Configuration: `~/.truerepublic/config/`

**Optional:**
- Chain data: `~/.truerepublic/data/` (can resync)

### Backup Script

Create `backup.sh`:

```bash
#!/bin/bash

BACKUP_DIR=~/truerepublic-backups
DATE=$(date +%Y%m%d-%H%M%S)
BACKUP_PATH=$BACKUP_DIR/backup-$DATE

mkdir -p $BACKUP_PATH

# Backup keys and config
cp -r ~/.truerepublic/config $BACKUP_PATH/

# Backup chain data (optional, large)
# cp -r ~/.truerepublic/data $BACKUP_PATH/

# Compress
tar -czf $BACKUP_PATH.tar.gz -C $BACKUP_DIR backup-$DATE
rm -rf $BACKUP_PATH

# Keep only last 7 backups
ls -t $BACKUP_DIR/backup-*.tar.gz | tail -n +8 | xargs rm -f

echo "Backup complete: $BACKUP_PATH.tar.gz"
```

Schedule with cron:

```bash
# Edit crontab
crontab -e

# Add daily backup at 2 AM
0 2 * * * /home/YOUR_USERNAME/backup.sh
```

### Recovery

**Restore from backup:**

```bash
# Stop node
sudo systemctl stop truerepublicd

# Extract backup
cd ~
tar -xzf truerepublic-backups/backup-YYYYMMDD-HHMMSS.tar.gz

# Restore config
cp -r backup-YYYYMMDD-HHMMSS/config/* ~/.truerepublic/config/

# Start node
sudo systemctl start truerepublicd
```

**If data is corrupted:**

```bash
# Remove data
rm -rf ~/.truerepublic/data

# Reinitialize
truerepublicd init my-node-name --chain-id truerepublic-1

# Restore config
# (keys, genesis, config.toml, app.toml)

# Resync from scratch or use state sync
```

---

## Upgrading

### Docker Upgrade

```bash
# Pull latest code
cd ~/TrueRepublic
git pull origin main

# Rebuild images
make docker-build

# Stop old version
docker-compose down

# Start new version
docker-compose up -d

# Check logs
docker-compose logs -f truerepublic-node
```

### Native Upgrade

```bash
# Stop node
sudo systemctl stop truerepublicd

# Backup current binary
sudo cp /usr/local/bin/truerepublicd /usr/local/bin/truerepublicd.backup

# Pull latest code
cd ~/TrueRepublic
git pull origin main

# Build new version
make build

# Install
sudo cp build/truerepublicd /usr/local/bin/

# Verify version
truerepublicd version

# Start node
sudo systemctl start truerepublicd

# Check logs
sudo journalctl -u truerepublicd -f
```

### Rollback if Issues

```bash
# Stop node
sudo systemctl stop truerepublicd

# Restore old binary
sudo cp /usr/local/bin/truerepublicd.backup /usr/local/bin/truerepublicd

# Start node
sudo systemctl start truerepublicd
```

---

## Next Steps

- [Validator Guide](Validator-Guide) -- Become a validator
- [Monitoring](Monitoring) -- Set up monitoring
- [Troubleshooting](Troubleshooting) -- Common issues
