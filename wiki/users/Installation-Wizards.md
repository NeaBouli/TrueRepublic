# Installation Wizards

Step-by-step guides to get started with TrueRepublic.

> Recovery is active. There is no approved production/mainnet client or fund
> flow. Use only local or explicitly approved test environments.

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

### Step 1: Open the maintained client (2 min)

1. Follow the [canonical Docker setup](/docs/node-operators/installation/docker-setup.md)
   and run `make docker-build` followed by `make docker-up`.
2. Open the loopback-only maintained client at `http://localhost:3001`.
3. Create or import a test-only local wallet.
4. Never enter a real mnemonic while recovery status remains active.
5. Keep any test mnemonic isolated from real funds and production accounts.

**Seed Phrase Example:**
```
word1 word2 word3 word4 word5 word6
word7 word8 word9 word10 word11 word12
word13 word14 word15 word16 word17 word18
word19 word20 word21 word22 word23 word24
```

### Step 2: Get PNYX Tokens (2 min)

**Test networks:** the project operates no faucet. Test PNYX exist only in a
test genesis allocation that you or your test-network operator control.

There is no approved mainnet or real-funds flow during recovery.

### Step 3: Connect to TrueRepublic (1 min)

1. Confirm that the configured local/test node is reachable.
2. Open the local test wallet in the maintained client.
3. Confirm the test address and balance.

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

# The Docker install script includes the Compose v2 plugin
# (required: Docker 24.0+, Compose v2.20+)

# Verify installation
docker --version
docker compose version
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

**Set these values** (the keys `.env.example` defines; there is no
`EXTERNAL_IP` setting):
```
MONIKER=my-node-name
CHAIN_ID=truerepublic-1
BOOTSTRAP_OPERATOR=<independent operator address>
GRAFANA_PASSWORD=<strong password>   # required; compose refuses to start without it
```

#### Step 4: Start Node (2 min)

```bash
# Build and start
make docker-build
make docker-up

# Check logs
docker compose logs -f truerepublic-node
```

#### Step 5: Verify Node Running (1 min)

```bash
# Check sync status (RPC is reachable on the host only through nginx)
curl -s http://127.0.0.1:8080/rpc/status | jq .result.sync_info

# Should show:
# "catching_up": false  (when fully synced)
# "latest_block_height": "<current height>"
```

#### Step 6: Access Monitoring (2 min)

Grafana listens only on the host loopback: open `http://127.0.0.1:3000` on the
node host (or through an SSH tunnel). Log in as `admin` with the
`GRAFANA_PASSWORD` you set in `.env`; there is no default password.

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
# 1. Node is fully synced (Docker setup: through nginx on the host)
curl -s http://127.0.0.1:8080/rpc/status | jq .result.sync_info.catching_up
# Should return: false

# 2. Have sufficient PNYX: at least 100,000,000,000 upnyx (100,000 PNYX).
#    truerepublicd has no bank command; check with a qualified client.

# 3. Member of a domain
truerepublicd query truedemocracy domains
# Check you're in at least one domain
```

### Step 1: Join a Domain (5 min)

If not already a member, the domain admin adds you
(`truerepublicd tx truedemocracy add-member <domain-name> <your-address>`), or
you request onboarding with `truerepublicd tx truedemocracy onboard-to-domain
[domain] [domain-pubkey-hex] [global-pubkey-hex] [signature-hex]`, which the
admin approves with `approve-onboarding`. There is no `join-domain` command.

### Step 2: Generate Validator Keys (5 min)

```bash
# Create validator key
truerepublicd keys add validator \
    --keyring-backend file

# CRITICAL: Backup the output!
# Save address, pubkey, and mnemonic securely
```

### Step 3: Fund Validator Address (5 min)

The validator address must already hold the stake. `truerepublicd` has no
`bank send` command; fund it through the qualified genesis allocation or a
client you have independently qualified.

### Step 4: Register as Validator (10 min)

`register-validator` takes `[pubkey-hex] [stake] [domain]`; the public key is
this node's Ed25519 consensus key as hex, not the signing account's public key.
Query the running node's public status; do not read or copy validator private-key
files. Choose the Docker or native status command for the actual node:

```bash
set -euo pipefail
# Docker node: query inside its container, not an unrelated host home.
STATUS_JSON=$(docker compose exec -T truerepublic-node truerepublicd status --node tcp://127.0.0.1:26657 --output json)
# Native alternative (replace the command above):
# STATUS_JSON=$(truerepublicd status --node tcp://127.0.0.1:26657 --output json)
PUBKEY_HEX=$(printf '%s' "$STATUS_JSON" | jq -er '.validator_info.pub_key | select(.type == "tendermint/PubKeyEd25519") | .value' | base64 -d | xxd -p -c 64)
[[ "$PUBKEY_HEX" =~ ^[0-9a-f]{64}$ ]] || exit 1

truerepublicd tx truedemocracy register-validator \
    "$PUBKEY_HEX" \
    100000000000upnyx \
    <domain-name> \
    --from validator \
    --chain-id truerepublic-1
```

The signing account and its qualified client may live separately from the node.
If public status is unavailable, obtain an independently confirmed public
32-byte Ed25519 consensus key; do not substitute a key from another node or account.

### Step 5: Verify Validator Status (5 min)

```bash
# Check validator info
truerepublicd query truedemocracy validator <validator-address>

# Should show your stake, power > 0, jailed: false and the domain
# (fields: operator_addr, pub_key, stake, domains, power, jailed,
#  jailed_until, missed_blocks)
```

### Step 6: Monitor Performance (10 min)

```bash
# Signing health is part of the validator record (no slashing CLI)
truerepublicd query truedemocracy validator <validator-address>

# missed_blocks: more than 50 missed commits in a complete 100-block
# window jails the validator and burns 1% of its stake
```

### Step 7: Configure Monitoring Alerts (10 min)

Use the repository-owned `monitoring/prometheus-alerts.yml`; do not create a
second local rule file or copy legacy `tendermint_*` examples. The shipped
rules use the verified `cometbft_*` and `truerepublic_*` families and include
deterministic Promtool tests.

Review the [supported monitoring guide](../operations/Monitoring) and map each
role label to a primary and secondary operator. The repository intentionally
does not configure an external Alertmanager destination. External paging must
remain off until a separately approved end-to-end delivery and
acknowledgement drill passes.

**The local recovery/testnet monitoring profile is now configured.** This does
not by itself qualify a production validator or authorize real funds.

**Next Steps:**
- [Validator Guide](../operations/Validator-Guide) -- Daily maintenance
- [Troubleshooting](../operations/Troubleshooting) -- Recover from issues

---

## Developer Setup

**Time:** 45 minutes
**Requirements:**
- Go 1.26.9
- Node.js 22+ (required by `client-web/package.json`)
- Git

### Step 1: Install Dependencies (15 min)

**Go:**

```bash
# Download Go
wget https://go.dev/dl/go1.26.9.linux-amd64.tar.gz

# Extract
sudo tar -C /usr/local -xzf go1.26.9.linux-amd64.tar.gz

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
# Maintained web client
cd client-web
npm ci
npm run lint
npm test
npm run build

# No native mobile build target exists; the former prototype was retired under GH-102.
```

### Step 5: Run Local Testnet (5 min)

```bash
# Initialize local testnet
BINARY=./build/truerepublicd ./scripts/init-node.sh

# Start node
BINARY=./build/truerepublicd ./scripts/start-node.sh

# In another terminal, check status
curl localhost:26657/status
```

The initializer delegates only to the generated-key, bank-backed PoD bootstrap.
TrueRepublic does not use `x/staking` gentx or collect-gentxs commands.

### Step 6: Run Tests

```bash
# Backend tests
make test

# Frontend tests
cd client-web
npm ci
npm test -- --run
cd ..

# Test coverage
./scripts/go-packages.sh go test -cover
```

**Done! Development environment ready.**

**Next Steps:**
- [Architecture Overview](../develop/Architecture-Overview) -- Understand the system
- [Code Structure](../develop/Code-Structure) -- Navigate the codebase
- [Module Deep-Dive](../develop/Module-Deep-Dive) -- Detailed module docs
- [Contributing Guide](../develop/Contributing-Guide) -- How to contribute

---

## Troubleshooting

### Local test wallet will not open

**Solution:**
1. Confirm the test-wallet password.
2. Confirm that the browser profile still contains the encrypted wallet entry.
3. Re-import only the isolated test mnemonic when necessary.
4. Record the browser error before clearing storage.

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
# Clean and rebuild (do not run `go mod tidy`: dependencies are pinned and
# release builds use -mod=readonly; verify the module cache instead)
make clean
go mod verify
make build

# If still fails, check Go version:
go version
# Should be 1.26.9
```

## Getting Help

- Documentation: This wiki + `/docs` folder
- Issues: https://github.com/NeaBouli/TrueRepublic/issues
- Discussions: https://github.com/NeaBouli/TrueRepublic/discussions
- Telegram: https://t.me/truerepublic
