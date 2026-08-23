# Developer Documentation

Technical documentation for developers building on or contributing to TrueRepublic.

## Table of Contents

### Architecture
- [System Architecture](architecture/system-overview.md) -- Layers, modules, data flow
- [Module Reference](architecture/module-reference.md) -- Detailed module documentation
- [Optional Ballot Architecture](../GOVERNANCE_BALLOT_ARCHITECTURE.md) -- Deferred domain ballot, privacy, state and migration contract
- [Sovereign V4 Edge Architecture](../SOVEREIGN_V4_EDGE_ARCHITECTURE.md) -- Native edge-node, civic workflow and sandboxed domain-app target

### API Reference
- [CLI Commands](api-reference/cli-commands.md) -- All transaction and query commands
- [Module Queries](api-reference/abci-queries.md) -- Supported CLI and protobuf gRPC boundary
- [REST & RPC Endpoints](api-reference/rest-rpc.md) -- HTTP API reference

### Integration
- [Web Client Integration](integration-guide/web-wallet.md) -- canonical
  `client-web` boundary and legacy retirement record
- [Retired Mobile Client](integration-guide/mobile.md) -- GH-102 retirement and
  replacement requirements
- [CosmJS Examples](integration-guide/cosmjs-examples.md) -- Code examples for common operations
- [Dependency Risk Acceptance](DEPENDENCY_RISK_ACCEPTANCE.md) -- Time-boxed dependency exceptions and removal conditions

### Smart Contracts
- [CosmWasm Contracts](smart-contracts/cosmwasm.md) -- Governance and treasury contracts

## Tech Stack

| Layer | Technology | Version |
|-------|-----------|---------|
| Consensus | CometBFT | v0.38.21 |
| Application | Cosmos SDK | v0.50.13 |
| Language | Go | 1.23.5 |
| Smart Contracts | CosmWasm | cosmwasm-std 3 |
| Web Frontend | React | 18.2 |
| Blockchain Client | CosmJS | 0.32-0.38 |
| CSS Framework | Tailwind CSS | 3.4 |

## Project Structure

```
TrueRepublic/
├── app.go                  # Cosmos SDK application entry point
├── go.mod                  # Go module (SDK v0.50.13)
├── Makefile                # Build targets
├── Dockerfile              # Multi-stage Docker build
├── docker-compose.yml      # Full stack deployment
├── x/
│   ├── truedemocracy/      # Governance module (13 msg types)
│   └── dex/                # DEX module (4 msg types)
├── treasury/
│   └── keeper/
│       └── rewards.go      # Tokenomics equations 1-5
├── contracts/              # CosmWasm smart contracts (Rust)
├── client-web/             # Maintained React/TypeScript/Vite frontend
├── docs/                   # Documentation
└── .github/                # CI/CD workflows
```

## Getting Started

### Build the Blockchain

```bash
git clone https://github.com/NeaBouli/TrueRepublic.git
cd TrueRepublic
go mod tidy
make build
make test    # Run 182 tests
```

### Run the Maintained Web Client

```bash
cd client-web
npm ci
npm run dev    # Development server on port 3001
```

### Build Smart Contracts

```bash
cd contracts
cargo build --release --target wasm32-unknown-unknown
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Ensure all tests pass: `make test`
5. Ensure code passes lint: `make lint`
6. Submit a pull request

## Next Steps

- [System Architecture](architecture/system-overview.md) -- Understand the design
- [CLI Commands](api-reference/cli-commands.md) -- Explore the API
- [CosmJS Examples](integration-guide/cosmjs-examples.md) -- Start building
