# Installation Guide

## Quick Start

### Option A: Docker (Recommended)

Run the complete TrueRepublic stack (node + web wallet + monitoring) with Docker:

```bash
git clone https://github.com/NeaBouli/TrueRepublic.git
cd TrueRepublic
cp .env.example .env    # Edit: set MONIKER, GRAFANA_PASSWORD
make docker-build
make docker-up
```

**Verify:**
```bash
curl http://localhost:26657/status          # Node
open http://localhost:3001                   # Web Wallet
open http://localhost:3000                   # Grafana (admin / your-password)
```

**Stop:** `make docker-down`

### Option B: Build from Source

```bash
git clone https://github.com/NeaBouli/TrueRepublic.git
cd TrueRepublic

# Build blockchain
go mod tidy
make build                    # Binary: ./build/truerepublicd

# Start node
./build/truerepublicd start

# Build web wallet (separate terminal)
cd web-wallet
npm install
npm start                     # Development server: http://localhost:3000
```

### Option C: Web Wallet Only

```bash
cd web-wallet
npm install
npm start
```

## Prerequisites

| Component | Requirement | Check |
|-----------|-------------|-------|
| **Docker** (Option A) | Docker 24.0+, Compose v2.20+ | `docker --version` |
| **Go** (Option B) | Go 1.23.5+ | `go version` |
| **Node.js** (web wallet) | Node.js 20+ | `node --version` |
| **Rust** (smart contracts) | Rust 1.75+ | `rustc --version` |

## What Gets Installed

### Docker Setup

| Service | Port | URL |
|---------|------|-----|
| Blockchain Node | 26656 (P2P), 26657 (RPC) | `http://localhost:26657` |
| Web Wallet | 3001 | `http://localhost:3001` |
| Grafana | 3000 | `http://localhost:3000` |
| Prometheus | 9091 | `http://localhost:9091` |
| REST API | 1317 | `http://localhost:1317` |
| gRPC | 9090 | `localhost:9090` |

### Native Build

Binary: `./build/truerepublicd` (or `$GOPATH/bin/truerepublicd` with `make install`)

Data directory: `~/.truerepublic/`

## Build Commands

| Command | Description |
|---------|-------------|
| `make build` | Build blockchain binary |
| `make install` | Install to $GOPATH/bin |
| `make test` | Run 182 tests with race detector |
| `make lint` | Run vet and staticcheck |
| `make clean` | Remove build artifacts |
| `make docker-build` | Build Docker images |
| `make docker-up` | Start Docker Compose stack |
| `make docker-down` | Stop Docker Compose stack |

## Building Smart Contracts

```bash
cd contracts
rustup target add wasm32-unknown-unknown
cargo build --release --target wasm32-unknown-unknown
```

## Building Mobile Wallet

```bash
cd mobile-wallet
npm install
npm start           # Start Expo dev server
npm run android     # Build for Android
npm run ios         # Build for iOS
```

## Next Steps

- **End users:** [User Manual](docs/user-manual/README.md)
- **Node operators:** [Node Operators Guide](docs/node-operators/README.md)
- **Validators:** [Validator Guide](docs/validators/README.md)
- **Developers:** [Developer Docs](docs/developers/README.md)
- **FAQ:** [Frequently Asked Questions](docs/FAQ.md)
