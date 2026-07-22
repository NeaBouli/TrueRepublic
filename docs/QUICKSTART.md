# Developer Quickstart

Get started with TrueRepublic development in 5 minutes.

## Prerequisites

```bash
# Install Go 1.24+
# See https://go.dev/dl/

# Install build tools (Linux)
sudo apt-get install build-essential

# Install build tools (macOS)
xcode-select --install
```

## Quick Setup

```bash
# Clone
git clone https://github.com/NeaBouli/TrueRepublic.git
cd TrueRepublic

# Build (with CGO for wasmvm)
CGO_ENABLED=1 make build

# Verify
./build/truerepublicd version
# v0.4.0
```

## Start Local Chain

```bash
# Initialize generated-key, bank-backed PoD genesis
BINARY=./build/truerepublicd \
CHAIN_ID=truerepublic-dev \
MONIKER=dev \
CHAIN_HOME="$PWD/.truerepublic-dev" \
./scripts/init-node.sh

# Start chain
./build/truerepublicd start --home "$PWD/.truerepublic-dev"
```

Chain is now running on `localhost:26657` (RPC) and `localhost:1317` (REST).
TrueRepublic does not wire `x/staking`; do not add staking gentxs. Transaction
examples below require a separately funded account and are not a faucet flow.

## Quick Examples

### Create Domain

```bash
./build/truerepublicd tx truedemocracy create-domain \
  governance "Governance Domain" \
  --from alice \
  --chain-id truerepublic-dev
```

### Create DEX Pool

```bash
# Register BTC asset first (as admin)
./build/truerepublicd tx dex register-asset \
  ibc/BTC "BTC" "Bitcoin" 8 cosmoshub-4 channel-0 \
  --from alice

# Create PNYX/BTC pool
./build/truerepublicd tx dex create-pool upnyx ibc/BTC 1000000 10000 \
  --from alice
```

### Swap Tokens

```bash
./build/truerepublicd tx dex swap pool-0 upnyx 1000 0 \
  --from alice
```

### Deploy Smart Contract

```bash
# Build contract
cd contracts/examples/governance-dao
cargo wasm

# Store on chain
./build/truerepublicd tx wasm store \
  target/wasm32-unknown-unknown/release/governance_dao.wasm \
  --from alice \
  --gas 2000000

# Instantiate (use code ID from store tx logs)
./build/truerepublicd tx wasm instantiate 1 \
  '{"domain_name":"governance","quorum_bps":5100,"threshold_bps":6700,"voting_period":86400}' \
  --from alice \
  --label "gov-dao" \
  --no-admin
```

## Run Tests

```bash
# All Go tests (655)
./scripts/go-packages.sh go test -timeout=600s

# Specific module
go test ./x/dex/...

# All Rust tests (26)
cd contracts && cargo test --workspace

# Maintained frontend tests (8)
cd client-web && npm ci && npm test -- --run
```

## Next Steps

- Read [API_REFERENCE.md](API_REFERENCE.md) for complete API
- See [ARCHITECTURE.md](ARCHITECTURE.md) for system design
- Check [DEPLOYMENT.md](DEPLOYMENT.md) for production setup
- Review [CONTRIBUTING.md](../CONTRIBUTING.md) to contribute
