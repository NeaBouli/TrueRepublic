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

### Smart Contracts

The CosmWasm contracts in `contracts/` are quarantined, non-production
prototypes (issue #308; see [`contracts/QUARANTINE.md`](../contracts/QUARANTINE.md)).
They are not deployable: their crates reject `wasm32` builds, and they must not
be stored, instantiated or migrated on any chain. Use the native modules
instead. Contract code is only built and tested on the host:

```bash
cd contracts && cargo test --workspace
```

## Run Tests

```bash
# All Go tests in the package-scoped standard suite (2,132; process gates run separately in CI)
./scripts/go-packages.sh go test -count=1 -timeout=600s

# Repository-only candidate evidence contract (no tag/release/deployment)
make candidate-evidence-contract-test

# Two-distinct-run exact-commit metadata comparison contract
make cross-run-evidence-contract-test

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
