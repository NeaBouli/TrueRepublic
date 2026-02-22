<p align="center">
  <img src="https://raw.githubusercontent.com/NeaBouli/TrueRepublic/main/assets/logo.png" alt="TrueRepublic Logo" width="200"/>
</p>

<h2 align="center">TrueRepublic/PNYX Technical Wiki</h2>

<p align="center">
  <img src="https://raw.githubusercontent.com/NeaBouli/TrueRepublic/main/assets/pnx_logo.png" alt="PNYX Coin" width="100"/>
</p>

Welcome to the TrueRepublic technical documentation wiki.

## Quick Navigation

### For Developers
- [Architecture Overview](develop/Architecture-Overview)
- [Code Structure](develop/Code-Structure)
- [Module Deep-Dive](develop/Module-Deep-Dive)
- [API Reference](develop/API-Reference)
- [Development Setup](develop/Development-Setup)
- [Contributing Guide](develop/Contributing-Guide)
- [Smart Contracts](develop/Smart-Contracts)
- [Frontend Architecture](develop/Frontend-Architecture)

### For Users
- [System Overview](users/System-Overview)
- [Installation Wizards](users/Installation-Wizards)
- [User Manuals](users/User-Manuals)
- [How It Works](users/How-It-Works)
- [Frontend Guide](users/Frontend-Guide)
- [FAQ](users/FAQ)

### For Node Operators
- [Node Setup Guide](operations/Node-Setup)
- [Validator Guide](operations/Validator-Guide)
- [Deployment Options](operations/Deployment-Options)
- [Monitoring & Maintenance](operations/Monitoring)
- [Troubleshooting](operations/Troubleshooting)

### Security & Audits
- [Security Architecture](security/Security-Architecture)
- [Audit Reports](security/Audit-Reports)
- [Test Coverage](security/Test-Coverage)
- [Known Issues](security/Known-Issues)
- [Best Practices](security/Best-Practices)

### Project Status
- [Current Status](status/Current-Status)
- [Roadmap](status/Roadmap)
- [Feature Matrix](status/Feature-Matrix)
- [Testing Status](status/Testing-Status)
- [Known Bugs](status/Known-Bugs)

## Project Information

| | |
|---|---|
| **Repository** | https://github.com/NeaBouli/TrueRepublic |
| **Version** | 0.1.0-alpha |
| **Status** | Pre-production (Testnet Ready) |
| **License** | Apache 2.0 |

## Technology Stack

| Layer | Technology | Version |
|-------|-----------|---------|
| Consensus | CometBFT | v0.38.21 |
| Application | Cosmos SDK | v0.50.13 |
| Language | Go | 1.23.5 |
| Smart Contracts | CosmWasm (Rust) | cosmwasm-std 3 |
| Web Frontend | React + Tailwind CSS | 18.2 / 3.4 |
| Mobile | React Native + Expo | 0.74 / 51.0 |
| Wallet | Keplr + CosmJS | 0.32-0.38 |
| Infrastructure | Docker, Prometheus, Grafana, Nginx | |

## Key Features

| Feature | Description |
|---------|-------------|
| Direct Democracy Governance | Domains, proposals, voting |
| Systemic Consensing | Rating scale -5 to +5 |
| Stones Voting | Dynamic ranking, VoteToEarn |
| Proof-of-Domain Consensus | Anti-whale, community-aligned |
| Integrated DEX | AMM with x*y=k, 0.3% fee, 1% PNYX burn |
| Anonymous Voting | Domain key pairs (WP S4) |
| VoteToEarn | Rewards for participation |
| CosmWasm Support | Wasm smart contracts |
| Suggestion Lifecycle | Green/yellow/red zones, auto-delete |
| Admin Election | Stone-based, continuous |

## Getting Started

| Audience | Start Here |
|----------|-----------|
| **Developers** | [Architecture Overview](develop/Architecture-Overview) |
| **Users** | [Installation Wizards](users/Installation-Wizards) |
| **Operators** | [Node Setup Guide](operations/Node-Setup) |

## Key Metrics

- **182 unit tests** across 3 modules (2,705 lines of test code)
- **17 transaction types** (13 governance + 4 DEX)
- **6 query endpoints** (4 governance + 2 DEX)
- **5 tokenomics equations** fully implemented
- **30+ documentation pages** in `/docs`

## Community

- Issues: https://github.com/NeaBouli/TrueRepublic/issues
- Telegram: https://t.me/truerepublic
- Email: p.cypher@protonmail.com

## License

Apache License 2.0 -- See LICENSE file
