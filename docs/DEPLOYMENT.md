# Deployment Guide

**Version:** v0.4.0 recovery baseline

> This guide covers local development and recovery-testnet evidence only.
> Mainnet, public deployment, production keys, and real funds are not approved.

## Prerequisites

- Go 1.26.6 (repository-pinned native toolchain)
- Docker and Docker Compose (for containerized deployment)
- Rust toolchain (for CosmWasm contracts)
- Node.js 22+ (for the maintained web client)

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
# Process liveness and traffic readiness
docker compose exec -T truerepublic-node truerepublicd healthcheck live
docker compose exec -T truerepublic-node truerepublicd healthcheck ready

# Check node status through the loopback proxy
curl --fail --silent --show-error http://localhost:8080/rpc/status

# Check maintained web client
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
export BOOTSTRAP_OPERATOR=truerepublic1... # independently controlled account
./build/truerepublicd init "$MONIKER" --chain-id "$CHAIN_ID" \
  --bootstrap-operator "$BOOTSTRAP_OPERATOR"
```

### 3. Configure

```bash
# Set minimum gas price
sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "25000upnyx"/' \
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
./build/truerepublicd init node1 --chain-id truerepublic-testnet \
  --bootstrap-operator "$NODE1_OPERATOR"
# Note the node ID
./build/truerepublicd tendermint show-node-id
# e.g., abc123def456...
```

### Node 2+

```bash
./build/truerepublicd init node2 --chain-id truerepublic-testnet \
  --bootstrap-operator "$NODE2_OPERATOR"

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
EnvironmentFile=/etc/truerepublic/lifecycle.env
ExecStartPre=/usr/local/libexec/truerepublic-install-lifecycle --contract=${LIFECYCLE_CONTRACT} --prefix=${INSTALL_PREFIX} --operator-state=${OPERATOR_STATE} --sha256=${ARTIFACT_SHA256} --source-ref=${SOURCE_REF} --target=${TARGET} --runtime=${RUNTIME} pre-start
ExecStart=/opt/truerepublic/bin/truerepublicd start --home /home/truerepublic/.truerepublic
Restart=on-failure
RestartSec=10
LimitNOFILE=65535
Environment="CGO_ENABLED=1"

[Install]
WantedBy=multi-user.target
```

Build `./cmd/install-lifecycle` from the same reviewed source commit and install
it root-owned at `/usr/local/libexec/truerepublic-install-lifecycle`. Install
the reviewed contract root-owned, mode `0644`, in a root-owned non-writable
directory outside the managed prefix so the service user can read it. Then
create the
root-owned, mode `0600` `/etc/truerepublic/lifecycle.env` with fixed values for
`LIFECYCLE_CONTRACT`, `INSTALL_PREFIX`, `OPERATOR_STATE`, `ARTIFACT_SHA256`,
`SOURCE_REF`, `TARGET`, and `RUNTIME` from the accepted release evidence. Do
not place shell syntax or secrets in that file. The complete `ExecStartPre`
above fails closed and prevents service startup if any identity check fails.

After verifying that pre-start gate manually:

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
sudo ufw allow 26656/tcp  # Public seed/sentry/RPC roles only
sudo ufw allow 443/tcp    # Reviewed TLS query proxy
sudo ufw deny 26657/tcp   # RPC remains loopback-only
sudo ufw deny 1317/tcp    # REST remains loopback-only
sudo ufw deny 9090/tcp    # gRPC remains loopback-only
sudo ufw enable
```

Apply the role matrix and validator from
[`docs/node-operators/configuration/network-policy.md`](node-operators/configuration/network-policy.md)
before any rollout review.

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
| Maintained Web Client | 3001 | React/Vite frontend |
| Nginx | 80/443 | Reverse proxy |
| Prometheus | 9091 | Metrics collection |
| Grafana | 3000 | Dashboards |

---

## Monitoring

- **Prometheus** privately scrapes CometBFT at `127.0.0.1:26660` and
  SDK/application metrics at `127.0.0.1:1317`.
- Prometheus 3.13.1 and Grafana 13.1.1 are pinned by tag plus multi-architecture
  digest in the local Compose profile.
- **Grafana** health, native dashboard provisioning, datasource UID and proxy
  query execution, every panel expression, Prometheus rule loading, and
  synthetic alert behavior are CI-verified at `http://localhost:3000`.
- The dashboard covers CometBFT, application success/invariant, PNYX
  supply/headroom, and runtime signals. Eleven recovery/testnet rules,
  measurable initial objectives, role ownership, and first-response guidance
  are shipped.
- External paging, production topology, private-environment capacity/retention
  qualification, named on-call assignment, and private live operator rehearsal
  remain separate rollout gates. GH-93 verifies only the repository-owned
  synthetic rehearsal; GH-97 verifies only bounded loopback capacity evidence.
- Configuration: `monitoring/prometheus.yml`, `monitoring/grafana/`

---

## Backup & Recovery

### Automated Backup

```bash
# Run daily at 3 AM via cron
0 3 * * * /path/to/scripts/backup.sh
```

### Manual Chain Data Backup

```bash
CHAIN_HOME=~/.truerepublic ./scripts/backup.sh ~/truerepublic-backups
```

The backup script writes a sanitized chain-data archive. It excludes
`config/node_key.json`, `config/priv_validator_key.json`,
`data/priv_validator_state.json`, and keyring directories. Store validator keys
and signer state through a separate offline key-custody process.

### Restore Sanitized Chain Data

```bash
./build/truerepublicd init restored-node --chain-id truerepublic-1 \
  --home ~/.truerepublic-restored \
  --bootstrap-operator "$RESTORE_BOOTSTRAP_OPERATOR"
./scripts/restore.sh ~/truerepublic-backups/truerepublic_YYYY-MM-DD.tar.gz ~/.truerepublic-restored
./build/truerepublicd start --home ~/.truerepublic-restored
```

---

## Upgrade Procedures

Do not overwrite a running binary in place. Verify the immutable release
evidence, stage the candidate outside the managed prefix, stop the service,
apply the identity-bound lifecycle upgrade, and run the pre-start verification
before the service may restart. The repository-owned procedure and fail-closed
tool are documented in
[Artifact Lifecycle](node-operators/installation/lifecycle.md).

Only the governed `v0.4.1` migration path is currently implemented. Arbitrary
or breaking migrations, store-loader changes for pre-GH-184 chains, and
destructive state resets are unsupported. Follow
[Governed Application Upgrades and Rollback](node-operators/operations/upgrades.md)
and stop for coordinated recovery if a candidate may have mutated state.

---

## Troubleshooting

| Issue | Solution |
|-------|----------|
| `CGO_ENABLED` error | Set `CGO_ENABLED=1` and install build-essential |
| Node won't sync | Check seeds/persistent_peers in config.toml |
| Out of memory | Increase RAM or enable swap |
| Port already in use | Check for existing processes: `lsof -i :26657` |
| Suspected WAL or database corruption | Stop the node, preserve logs and state, then follow the reviewed backup/recovery and incident-command procedure; do not reset validator data |
| IBC timeout | Verify relayer is running and channels are open |
