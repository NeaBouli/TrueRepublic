# Installation Wizards

Step-by-step guides to get started with TrueRepublic.

## Choose Your Path

### I want to use TrueRepublic
[End User Setup](#end-user-setup) (5 minutes)

### I want to run a node
[Node Operator Setup](#node-operator-setup) (30 minutes)

### I want to become a validator
[Validator Setup](#validator-setup) (1 hour)

### I want to develop on TrueRepublic
[Developer Setup](#developer-setup) (45 minutes)

---

## End User Setup

**Time:** 5 minutes
**Requirements:** Chrome/Firefox/Brave browser

### Step 1: Install Keplr Wallet (2 min)

1. Visit https://www.keplr.app/
2. Click "Install Keplr"
3. Add to browser
4. Click Keplr icon in toolbar
5. Choose "Create new wallet"
6. **CRITICAL:** Write down 24-word seed phrase
7. Store seed phrase securely (paper, safe)
8. Never share seed phrase with anyone

**Seed Phrase Example:**
```
word1 word2 word3 word4 word5 word6
word7 word8 word9 word10 word11 word12
word13 word14 word15 word16 word17 word18
word19 word20 word21 word22 word23 word24
```

### Step 2: Get PNYX Tokens (2 min)

**Testnet:**

1. Join TrueRepublic Discord
2. Go to #faucet channel
3. Type: `/faucet <your-address>`
4. Receive 10,000 PNYX

**Mainnet:**

1. Buy on exchange
2. Withdraw to your Keplr address

### Step 3: Connect to TrueRepublic (1 min)

1. Visit https://truerepublic.app
2. Click "Connect Keplr Wallet"
3. Keplr popup appears
4. Click "Approve"
5. See your balance in top-right

**Done! You're ready to participate.**

**Next Steps:**
- [User Manuals](User-Manuals) -- Learn how to use features
- [Governance Tutorial](/docs/user-manual/governance-tutorial.md) -- Join domains and vote
- [DEX Guide](/docs/user-manual/dex-trading-guide.md) -- Trade tokens

---

## Node Operator Setup

**Time:** 30 minutes
**Requirements:**
- Ubuntu 20.04+ or Docker
- 4 CPU cores
- 8 GB RAM
- 500 GB SSD
- 100 Mbps connection

### Option A: Docker Setup (Recommended)

#### Step 1: Install Docker (5 min)

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo apt install docker-compose -y

# Verify installation
docker --version
docker-compose --version
```

#### Step 2: Clone Repository (2 min)

```bash
cd ~
git clone https://github.com/NeaBouli/TrueRepublic.git
cd TrueRepublic
```

#### Step 3: Configure Environment (5 min)

```bash
# Copy example env file
cp .env.example .env

# Edit configuration
nano .env
```

**Set these values:**
```
MONIKER=my-node-name
EXTERNAL_IP=your.server.ip
CHAIN_ID=truerepublic-1
```

#### Step 4: Start Node (2 min)

```bash
# Build and start
make docker-build
make docker-up

# Check logs
docker-compose logs -f truerepublic-node
```

#### Step 5: Verify Node Running (1 min)

```bash
# Check sync status
curl localhost:26657/status | jq .result.sync_info

# Should show:
# "catching_up": false  (when fully synced)
# "latest_block_height": "<current height>"
```

#### Step 6: Access Monitoring (2 min)

Visit: `http://your-server-ip:3000`
- Username: `admin`
- Password: `admin` (change immediately)

**Done! Your node is running.**

### Option B: Native Setup

See [Node Setup Guide](../operations/Node-Setup) for native installation.

**Next Steps:**
- [Monitoring Setup](../operations/Monitoring) -- Configure alerts
- [Backup Strategy](../operations/Monitoring) -- Protect your data
- [Validator Guide](../operations/Validator-Guide) -- Upgrade to validator

---

## Validator Setup

**Time:** 1 hour
**Requirements:**
- Running full node (synced)
- 100,000 PNYX minimum
- Domain membership
- 24/7 uptime capability

### Prerequisites Check

Before starting, ensure:

```bash
# 1. Node is fully synced
curl localhost:26657/status | jq .result.sync_info.catching_up
# Should return: false

# 2. Have sufficient PNYX
truerepublicd query bank balances <your-address>
# Should show: 100000+ pnyx

# 3. Member of a domain
truerepublicd query truedemocracy domains
# Check you're in at least one domain
```

### Step 1: Join a Domain (5 min)

If not already a member:

```bash
truerepublicd tx truedemocracy join-domain <domain-name> \
    --from <your-key> \
    --chain-id truerepublic-1 \
    --gas auto \
    --gas-adjustment 1.3
```

### Step 2: Generate Validator Keys (5 min)

```bash
# Create validator key
truerepublicd keys add validator \
    --keyring-backend file

# CRITICAL: Backup the output!
# Save address, pubkey, and mnemonic securely
```

### Step 3: Fund Validator Address (5 min)

```bash
# Send PNYX to validator address
truerepublicd tx bank send <your-key> <validator-address> 100000000000upnyx \
    --from <your-key> \
    --chain-id truerepublic-1
```

### Step 4: Register as Validator (10 min)

```bash
truerepublicd tx truedemocracy register-validator \
    <domain-name> \
    100000000000upnyx \
    --from validator \
    --chain-id truerepublic-1 \
    --gas auto \
    --gas-adjustment 1.3
```

### Step 5: Verify Validator Status (5 min)

```bash
# Check validator info
truerepublicd query truedemocracy validator <validator-address>

# Should show:
# - status: active
# - jailed: false
# - domain: <your-domain>
```

### Step 6: Monitor Performance (10 min)

```bash
# Check signing status
truerepublicd query slashing signing-info <validator-consensus-address>

# Monitor uptime
# Uptime must stay above 95% to avoid slashing
```

### Step 7: Configure Monitoring Alerts (10 min)

Edit `monitoring/prometheus.yml`:

```yaml
alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - localhost:9093

rule_files:
  - "alerts.yml"
```

Create `monitoring/alerts.yml`:

```yaml
groups:
  - name: validator_alerts
    rules:
      - alert: ValidatorDown
        expr: up{job="validator"} == 0
        for: 5m
        annotations:
          summary: "Validator is down"

      - alert: MissedBlocks
        expr: increase(tendermint_consensus_validators_missed_blocks[1h]) > 10
        annotations:
          summary: "Validator missing blocks"
```

**Done! You're a validator.**

**Next Steps:**
- [Validator Guide](../operations/Validator-Guide) -- Daily maintenance
- [Troubleshooting](../operations/Troubleshooting) -- Recover from issues

---

## Developer Setup

**Time:** 45 minutes
**Requirements:**
- Go 1.23+
- Node.js 18+
- Git

### Step 1: Install Dependencies (15 min)

**Go:**

```bash
# Download Go
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz

# Extract
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz

# Add to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify
go version
```

**Node.js:**

```bash
# Install via nvm
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
source ~/.bashrc

# Install Node.js
nvm install 18
nvm use 18

# Verify
node --version
npm --version
```

### Step 2: Clone Repository (5 min)

```bash
cd ~
git clone https://github.com/NeaBouli/TrueRepublic.git
cd TrueRepublic
```

### Step 3: Build Backend (10 min)

```bash
# Install Go dependencies
go mod download

# Build blockchain binary
make build

# Verify
./build/truerepublicd version
```

### Step 4: Build Frontend (10 min)

```bash
# Web wallet
cd web-wallet
npm install
npm run build

# Mobile wallet (optional)
cd ../mobile-wallet
npm install
```

### Step 5: Run Local Testnet (5 min)

```bash
# Initialize local testnet
./scripts/init-node.sh

# Start node
./scripts/start-node.sh

# In another terminal, check status
curl localhost:26657/status
```

### Step 6: Run Tests

```bash
# Backend tests
make test

# Frontend tests
cd web-wallet
npm test

# Test coverage
go test ./... -cover
```

**Done! Development environment ready.**

**Next Steps:**
- [Architecture Overview](../develop/Architecture-Overview) -- Understand the system
- [Code Structure](../develop/Code-Structure) -- Navigate the codebase
- [Module Deep-Dive](../develop/Module-Deep-Dive) -- Detailed module docs
- [Contributing Guide](../develop/Contributing-Guide) -- How to contribute

---

## Troubleshooting

### Keplr won't connect

**Solution:**
1. Refresh page
2. Unlock Keplr
3. Clear browser cache
4. Try different browser

### Node won't sync

**Solution:**

```bash
# Check peers
curl localhost:26657/net_info | jq .result.n_peers

# If 0 peers, add seeds to config.toml:
nano ~/.truerepublic/config/config.toml

# Add under [p2p]:
seeds = "seed1@ip:port,seed2@ip:port"
```

### Validator jailed

**Solution:**

```bash
# Wait for jail period to expire
truerepublicd query truedemocracy validator <address>

# Check jail_until time
# When expired, unjail:
truerepublicd tx truedemocracy unjail \
    --from validator \
    --chain-id truerepublic-1
```

### Build fails

**Solution:**

```bash
# Clean and rebuild
make clean
go mod tidy
make build

# If still fails, check Go version:
go version
# Should be 1.23 or higher
```

## Getting Help

- Documentation: This wiki + `/docs` folder
- Issues: https://github.com/NeaBouli/TrueRepublic/issues
- Discussions: https://github.com/NeaBouli/TrueRepublic/discussions
- Telegram: https://t.me/truerepublic
