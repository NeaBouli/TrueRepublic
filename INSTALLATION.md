# Installation Guide

## Quick Start

### Option A: Docker recovery profile

Run the complete TrueRepublic stack (node + maintained web client + monitoring)
with Docker:

```bash
git clone https://github.com/NeaBouli/TrueRepublic.git
cd TrueRepublic
cp .env.example .env    # Edit: set MONIKER, GRAFANA_PASSWORD
make docker-build
make docker-up
```

**Verify:**
```bash
docker compose exec -T truerepublic-node \
  truerepublicd healthcheck live             # Process liveness
docker compose exec -T truerepublic-node \
  truerepublicd healthcheck ready            # Synced traffic readiness
curl http://localhost:8080/rpc/status        # Node status through local proxy
open http://localhost:3001                   # Maintained web client
open http://localhost:3000                   # Grafana (admin / your-password)
```

**Stop:** `make docker-down`

### Option B: Build from Source

```bash
git clone https://github.com/NeaBouli/TrueRepublic.git
cd TrueRepublic

# Build blockchain without changing the module graph
make build                    # Binary: ./build/truerepublicd

# Start node
./build/truerepublicd start

# Build maintained web client (separate terminal)
cd client-web
npm ci
npm run dev                   # Development server: http://localhost:3001
```

### Option C: Maintained Web Client Only

```bash
cd client-web
npm ci
npm run dev
```

## Prerequisites

| Component | Requirement | Check |
|-----------|-------------|-------|
| **Docker** (Option A) | Docker 24.0+, Compose v2.20+ | `docker --version` |
| **Go** (Option B) | Go 1.26.6 | `go version` |
| **Node.js** (web client) | Node.js 22+ | `node --version` |
| **Rust** (smart contracts) | Rust 1.75+ | `rustc --version` |

## What Gets Installed

### Docker Setup

| Service | Port | URL |
|---------|------|-----|
| Local reverse proxy | 8080 | `http://localhost:8080` |
| Blockchain P2P (local development) | 26656 | `tcp://127.0.0.1:26656` |
| Maintained Web Client | 3001 | `http://localhost:3001` |
| Grafana | 3000 | `http://localhost:3000` |
| Prometheus | 9091 | `http://localhost:9091` |

RPC and REST stay inside the node container on loopback and are reached
through the local reverse proxy. gRPC is disabled by the default Compose
profile. Production role exposure must follow the
[network policy](docs/node-operators/configuration/network-policy.md).

### Native Build

Binary: `./build/truerepublicd`

Data directory: `~/.truerepublic/`

## Build Commands

| Command | Description |
|---------|-------------|
| `make build` | Build blockchain binary |
| `make test` | Run the Go test suite with race detector |
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

## Mobile Wallet

The former Expo prototype was retired and removed under GH-102. It was not safe
for real keys, had no meaningful tests, and could not produce a current Android
bundle. There is no supported mobile build target; use `client-web` until a new
mobile client is designed and independently reviewed.

## Retired Legacy Web Wallet

The former Create React App prototype was retired and removed under GH-112.
It carried 70 dependency advisories, broken custom-query calls, obsolete custom
transaction paths, and a real bank-send path. Git history preserves it for
audit only; use `client-web` for all current development.

## Next Steps

- **Artifact lifecycle:** [Install, upgrade, rollback, and uninstall](docs/node-operators/installation/lifecycle.md)
- **End users:** [User Manual](docs/user-manual/README.md)
- **Node operators:** [Node Operators Guide](docs/node-operators/README.md)
- **Validators:** [Validator Guide](docs/validators/README.md)
- **Developers:** [Developer Docs](docs/developers/README.md)
- **FAQ:** [Frequently Asked Questions](docs/FAQ.md)
