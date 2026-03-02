<p align="center">
  <img src="docs/assets/images/logo.png" alt="TrueRepublic Logo" width="200"/>
</p>

<h1 align="center">TrueRepublic</h1>

<p align="center">
  <strong>A Cosmos SDK blockchain with Zero-Knowledge Proof anonymity, CosmWasm smart contracts, and Multi-Asset DEX</strong>
</p>

<p align="center">
  <a href="#key-features">Features</a> &bull;
  <a href="#quick-start">Quick Start</a> &bull;
  <a href="#-documentation">Documentation</a> &bull;
  <a href="#current-status">Roadmap</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/tests-577%20passing-brightgreen" alt="Tests"/>
  <img src="https://img.shields.io/badge/version-v0.3.0-blue" alt="Version"/>
  <img src="https://img.shields.io/badge/Go-1.24-00ADD8?logo=go" alt="Go"/>
  <img src="https://img.shields.io/badge/Rust-1.75+-orange?logo=rust" alt="Rust"/>
</p>

<p align="center">
  <a href="https://github.com/NeaBouli/TrueRepublic/actions/workflows/go-ci.yml"><img src="https://github.com/NeaBouli/TrueRepublic/actions/workflows/go-ci.yml/badge.svg" alt="Go CI"/></a>
  <a href="https://github.com/NeaBouli/TrueRepublic/actions/workflows/rust-ci.yml"><img src="https://github.com/NeaBouli/TrueRepublic/actions/workflows/rust-ci.yml/badge.svg" alt="Rust CI"/></a>
  <a href="https://github.com/NeaBouli/TrueRepublic/actions/workflows/react-ci.yml"><img src="https://github.com/NeaBouli/TrueRepublic/actions/workflows/react-ci.yml/badge.svg" alt="Web CI"/></a>
  <a href="https://github.com/NeaBouli/TrueRepublic/actions/workflows/react-native-ci.yml"><img src="https://github.com/NeaBouli/TrueRepublic/actions/workflows/react-native-ci.yml/badge.svg" alt="Mobile CI"/></a>
</p>

<p align="center">
  <a href="https://neabouli.github.io/TrueRepublic/">Website</a> &bull;
  <a href="https://github.com/NeaBouli/TrueRepublic/wiki">Wiki</a> &bull;
  <a href="docs/WhitePaper_TR_eng.md">Whitepaper</a> &bull;
  <a href="https://t.me/truerepublic">Telegram</a>
</p>

---

## What is TrueRepublic?

TrueRepublic is a platform for **direct democracy** and **digital self-determination**. Instead of electing representatives, participants make decisions directly through community-governed **Domains** using **Systemic Consensing** (rating -5 to +5) and **Stones Voting**.

The native token **PNYX** -- named after the hill in Athens where citizens gathered to vote -- powers governance, treasury mechanisms, staking, and a built-in DEX.

---

## 📚 Documentation

### [📖 Complete Wiki](https://github.com/NeaBouli/TrueRepublic/wiki)

**For Developers:**
- [Architecture Overview](https://github.com/NeaBouli/TrueRepublic/wiki/develop-Architecture-Overview)
- [Code Structure](https://github.com/NeaBouli/TrueRepublic/wiki/develop-Code-Structure)
- [Module Deep-Dive](https://github.com/NeaBouli/TrueRepublic/wiki/develop-Module-Deep-Dive)

**For Users:**
- [System Overview](https://github.com/NeaBouli/TrueRepublic/wiki/users-System-Overview)
- [Installation Wizards](https://github.com/NeaBouli/TrueRepublic/wiki/users-Installation-Wizards)
- [User Manuals](https://github.com/NeaBouli/TrueRepublic/wiki/users-User-Manuals)
- [How It Works](https://github.com/NeaBouli/TrueRepublic/wiki/users-How-It-Works)

**For Node Operators:**
- [Node Setup](https://github.com/NeaBouli/TrueRepublic/wiki/operations-Node-Setup)
- [Validator Guide](https://github.com/NeaBouli/TrueRepublic/wiki/operations-Validator-Guide)
- [Monitoring](https://github.com/NeaBouli/TrueRepublic/wiki/operations-Monitoring)

**For Security:**
- [Security Architecture](https://github.com/NeaBouli/TrueRepublic/wiki/security-Security-Architecture)
- [Best Practices](https://github.com/NeaBouli/TrueRepublic/wiki/security-Best-Practices)

### Additional Docs

| Guide | Audience | Description |
|-------|----------|-------------|
| **[Getting Started](docs/getting-started/README.md)** | Everyone | Choose your path: user, operator, validator, or developer |
| **[Installation](INSTALLATION.md)** | Everyone | Quick install guide (Docker / native / web wallet) |
| **[User Manual](docs/user-manual/README.md)** | End Users | Wallet, governance, voting, DEX trading |
| **[Node Operators](docs/node-operators/README.md)** | Operators | Setup, configuration, monitoring, backup |
| **[Validator Guide](docs/validators/README.md)** | Validators | PoD consensus, staking, slashing, operations |
| **[Developer Docs](docs/developers/README.md)** | Developers | Architecture, API reference, CosmJS, smart contracts |
| **[FAQ](docs/FAQ.md)** | Everyone | Frequently asked questions |
| **[Glossary](docs/GLOSSARY.md)** | Everyone | Term definitions |
| **[Whitepaper](docs/WhitePaper_TR_eng.md)** | Everyone | Full whitepaper |

---

## Quick Start

```bash
# Clone
git clone https://github.com/NeaBouli/TrueRepublic.git
cd TrueRepublic

# Option A: Docker (full stack)
cp .env.example .env && make docker-build && make docker-up

# Option B: Build from source
make build && ./build/truerepublicd start

# Option C: Web wallet only
cd web-wallet && npm install && npm start
```

See [INSTALLATION.md](INSTALLATION.md) for detailed instructions.

---

## Key Features

| Feature | Description | Docs |
|---------|-------------|------|
| **Domains** | Community-governed spaces for specific topics | [Governance Tutorial](docs/user-manual/governance-tutorial.md) |
| **Systemic Consensing** | Rate proposals -5 to +5 instead of Yes/No | [SC Explained](docs/user-manual/systemic-consensing-explained.md) |
| **Stones Voting** | Highlight importance, elect admins, earn rewards | [Stones Guide](docs/user-manual/stones-voting-guide.md) |
| **Anonymous Voting** | Domain key pairs for unlinkable ratings (WP S4) | [Architecture](docs/developers/architecture/module-reference.md) |
| **Proof of Domain** | Validators must be active domain members | [Validator Guide](docs/validators/README.md) |
| **DEX (AMM)** | Token swaps with 0.3% fee + 1% PNYX burn | [DEX Guide](docs/user-manual/dex-trading-guide.md) |
| **VoteToEarn** | Earn PNYX rewards for active participation | [Stones Guide](docs/user-manual/stones-voting-guide.md) |
| **Suggestion Lifecycle** | Green/yellow/red zones with auto-delete | [Governance](docs/user-manual/governance-tutorial.md) |
| **IBC Transfers** | Cross-chain PNYX via ICS-20 (ibc-go v8) | [IBC Setup](docs/IBC_RELAYER_SETUP.md) |

---

## Repository Structure

```text
TrueRepublic/
├── app.go                      Cosmos SDK application entry point
├── go.mod / go.sum             Go module (SDK v0.50.14, CometBFT v0.38.21)
├── Makefile                    Build targets (build, test, lint, docker)
├── INSTALLATION.md             Quick install guide
├── x/
│   ├── truedemocracy/          Governance module (23 msg types, 419 tests)
│   └── dex/                    DEX module (6 msg types, 68 tests)
├── treasury/keeper/            Tokenomics equations 1-5 (31 tests)
├── contracts/                  CosmWasm workspace (7 crates, 26 tests)
│   ├── core/                   Governance + treasury contracts
│   ├── packages/bindings/      TrueRepublic custom query/msg types
│   ├── packages/testing-utils/ Mock querier, AMM pool, fixtures
│   └── examples/               governance-dao, dex-bot, zkp-aggregator, token-vesting
├── web-wallet/                 React 18 + Tailwind + Keplr + CosmJS
├── mobile-wallet/              React Native + Expo
├── docs/
│   ├── getting-started/        Quick start guides
│   ├── user-manual/            End-user documentation (7 guides)
│   ├── node-operators/         Node setup, config, operations (9 guides)
│   ├── validators/             Validator guide with PoD
│   ├── developers/             Architecture, API, integration (8 guides)
│   ├── FAQ.md                  Frequently asked questions
│   └── GLOSSARY.md             Term definitions (60+ terms)
└── .github/                    CI/CD workflows (Go, Rust, React, RN)
```

---

## Implemented Features

| Feature | Status | Location |
|---------|--------|----------|
| Domains & Governance | ✅ | `x/truedemocracy/keeper.go` |
| Systemic Consensing (-5..+5) | ✅ | `x/truedemocracy/keeper.go` |
| Proof of Domain (PoD) | ✅ | `x/truedemocracy/validator.go` |
| Validator Slashing | ✅ | `x/truedemocracy/slashing.go` |
| Tokenomics (eq.1-5) | ✅ | `treasury/keeper/rewards.go` |
| DEX / AMM (x*y=k) | ✅ | `x/dex/keeper.go` |
| Multi-Asset DEX (BTC/ETH/LUSD) | ✅ | `x/dex/keeper.go` (asset registry + trading validation) |
| Node Staking Rewards (10% APY) | ✅ | `treasury/keeper/rewards.go` (eq.5) |
| Domain Interest (25% APY) | ✅ | `treasury/keeper/rewards.go` (eq.4) |
| Release Decay | ✅ | `treasury/keeper/rewards.go` |
| Anonymous Voting (WP S4) | ✅ | `x/truedemocracy/anonymity.go` |
| Zero-Knowledge Proofs (Groth16) | ✅ | `x/truedemocracy/zkp.go` |
| CosmWasm Smart Contracts | ✅ | `x/truedemocracy/wasm_bindings.go` |
| Domain-Bank Bridge | ✅ | `x/truedemocracy/treasury_bridge.go` |
| IBC Transfer (ICS-20) | ✅ | `app.go` (ibc-go v8.4.0) |
| Stones Voting (WP S3.1) | ✅ | `x/truedemocracy/stones.go` |
| VoteToEarn Rewards | ✅ | `x/truedemocracy/stones.go` |
| Suggestion Lifecycle (WP S3.1.2) | ✅ | `x/truedemocracy/lifecycle.go` |
| Green/Yellow/Red Zones | ✅ | `x/truedemocracy/lifecycle.go` |
| Auto-Delete & Fast Delete (2/3) | ✅ | `x/truedemocracy/lifecycle.go` |
| Admin Election (WP S3.6) | ✅ | `x/truedemocracy/governance.go` |
| Member Exclusion (2/3 vote) | ✅ | `x/truedemocracy/governance.go` |
| PoD Transfer Limit (10%, WP S7) | ✅ | `x/truedemocracy/validator.go` |
| CLI Commands (24 tx + 7 query) | ✅ | `x/truedemocracy/cli.go` |
| DEX CLI (6 tx + 4 query) | ✅ | `x/dex/cli.go` |
| CosmWasm Contracts (7 crates) | ✅ | `contracts/` (workspace) |
| Web Wallet (React + Keplr) | ✅ | `web-wallet/` |
| Mobile Wallet (Expo + RN) | ✅ | `mobile-wallet/` |
| CI/CD Workflows | ✅ | `.github/workflows/` |

---

## Build & Test

```bash
# Blockchain
go mod tidy
go build ./...
go test ./... -race -cover -count=1 -timeout=600s    # 533 tests

# Smart contracts
cd contracts && cargo test --workspace       # 26 tests

# Web wallet
cd web-wallet && npm install && npm run build

# Mobile wallet
cd mobile-wallet && npm install
```

---

## Tech Stack

| Layer | Technology | Version |
|-------|-----------|---------|
| Consensus | CometBFT | v0.38.21 |
| Application | Cosmos SDK | v0.50.14 |
| Language | Go | 1.24 |
| IBC | ibc-go | v8.4.0 |
| Smart Contracts | CosmWasm (Rust) | cosmwasm-std 3 |
| Web Frontend | React + Tailwind CSS | 18.2 / 3.4 |
| Mobile | React Native + Expo | 0.74 / 51.0 |
| Wallet | Keplr + CosmJS | 0.32-0.38 |

---

## Current Status

**Version: v0.3.0 (100% Complete)**

- ✅ 577 tests (533 Go + 26 Rust + 18 Frontend)
- ✅ Core blockchain compiles and runs
- ✅ Whitepaper tokenomics fully implemented (equations 1-5)
- ✅ Complete governance system (domains, proposals, voting, lifecycle)
- ✅ Zero-Knowledge Proofs (Groth16 ZK-SNARKs for anonymous voting)
- ✅ CosmWasm smart contract integration (wasmd v0.53.3)
- ✅ Domain-Bank Bridge (dual accounting, deposit/withdraw)
- ✅ IBC Transfer module (ibc-go v8.4.0, cross-chain PNYX transfers)
- ✅ Multi-Asset DEX: asset registry, trading validation, symbol resolution
- ✅ Cross-Chain Liquidity: multi-hop swaps, pool analytics, slippage protection
- ✅ UI Components: ZKP voting panel, DEX analytics (8 React components)
- ✅ Developer Tooling: 4 CosmWasm example contracts, shared bindings, testing utils
- ✅ DEX with AMM, liquidity pools, swap fees, PNYX burn
- ✅ Web wallet with 3-column governance UI
- ✅ Mobile wallet with bottom-tab navigation
- ✅ Comprehensive documentation (30+ guides)

### Roadmap

- ✅ **v0.1.x (Feb 2026):** Security fixes, documentation, elections
- ✅ **v0.2.x (Feb 2026):** Governance core — Systemic Consensing, Tokenomics, Elections
- ✅ **v0.3.0 (Q1 2026): ZKP Anonymity, CosmWasm, IBC, Multi-Asset DEX (100% COMPLETE)**
  - ✅ Weeks 1-4: ZKP Anonymity Layer (Groth16, Merkle trees, nullifiers)
  - ✅ Week 5: CosmWasm Integration (wasmd v0.53.3, custom bindings)
  - ✅ Week 6: Domain-Bank Bridge (dual accounting, deposit/withdraw)
  - ✅ Week 7: IBC Integration (ICS-20 transfer, relayer support)
  - ✅ Week 8: Multi-Asset DEX (asset registry, trading validation, symbol resolution)
  - ✅ Week 9: Cross-Chain Liquidity (multi-hop swaps, analytics)
  - ✅ Week 10: UI Components (ZKP voting, DEX analytics)
  - ✅ Week 11: Developer Tooling (contract examples, testing utils)
  - ✅ Week 12: Complete Documentation (API, deployment, architecture)
- 📋 **v0.4.0 (Q2 2026):** Optional Indexer Stack — SQL analytics, Read-Only API, Explorer
- 📋 **v0.5.0 (Q3 2026):** DEX Expansion — BTC/ETH/LUSD via IBC
- 🎯 **v1.0.0 (Q4 2026):** Production Release — External audit, mainnet launch

> **v0.3.0 Milestone Achieved!** All 12 weeks of the roadmap completed.
> 577 tests (533 Go + 26 Rust + 18 Frontend), zero regressions.

---

## Developer Documentation

| Guide | Description |
|-------|-------------|
| [API Reference](docs/API_REFERENCE.md) | Complete API overview |
| [Deployment Guide](docs/DEPLOYMENT.md) | Production setup |
| [Architecture](docs/ARCHITECTURE.md) | System design |
| [Quick Start](docs/QUICKSTART.md) | 5-minute setup |
| [Contributing](CONTRIBUTING.md) | Development guide |

## Contributing

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Ensure all tests pass: `make test`
5. Submit a pull request

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

## Community

- Telegram: [t.me/truerepublic](https://t.me/truerepublic)
- Issues: [github.com/NeaBouli/TrueRepublic/issues](https://github.com/NeaBouli/TrueRepublic/issues)
- Email: p.cypher@protonmail.com

## Contributors

- NeaBouli

## Donations

Team (BTC multi-sig): `bc1qyamf3twgcqckuqrvmwgwnhzupgshxs37eejdgl0ntcqve98qnvhqe6cjl9`
