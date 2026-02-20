<p align="center">
  <img src="assets/logo-dark.svg" alt="TrueRepublic Logo" width="400"/>
</p>

<h1 align="center">TrueRepublic / PNYX</h1>

<p align="center">
  <a href="https://github.com/NeaBouli/TrueRepublic/actions/workflows/go-ci.yml"><img src="https://github.com/NeaBouli/TrueRepublic/actions/workflows/go-ci.yml/badge.svg" alt="Go CI"/></a>
  <a href="https://github.com/NeaBouli/TrueRepublic/actions/workflows/rust-ci.yml"><img src="https://github.com/NeaBouli/TrueRepublic/actions/workflows/rust-ci.yml/badge.svg" alt="Rust CI"/></a>
  <a href="https://github.com/NeaBouli/TrueRepublic/actions/workflows/react-ci.yml"><img src="https://github.com/NeaBouli/TrueRepublic/actions/workflows/react-ci.yml/badge.svg" alt="Web CI"/></a>
  <a href="https://github.com/NeaBouli/TrueRepublic/actions/workflows/react-native-ci.yml"><img src="https://github.com/NeaBouli/TrueRepublic/actions/workflows/react-native-ci.yml/badge.svg" alt="Mobile CI"/></a>
</p>

<p align="center">
  <strong>Direct democracy on the blockchain. Built with Cosmos SDK.</strong>
</p>

---

## What is TrueRepublic?

TrueRepublic is a platform for **direct democracy** and **digital self-determination**. Instead of electing representatives, participants make decisions directly through community-governed **Domains** using **Systemic Consensing** (rating -5 to +5) and **Stones Voting**.

The native token **PNYX** -- named after the hill in Athens where citizens gathered to vote -- powers governance, treasury mechanisms, staking, and a built-in DEX.

---

## Documentation

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
| **[Whitepaper](docs/WhitePaper_TR_eng.pdf)** | Everyone | Full whitepaper (PDF) |

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

---

## Repository Structure

```text
TrueRepublic/
├── app.go                      Cosmos SDK application entry point
├── go.mod / go.sum             Go module (SDK v0.50.13, CometBFT v0.38.17)
├── Makefile                    Build targets (build, test, lint, docker)
├── INSTALLATION.md             Quick install guide
├── x/
│   ├── truedemocracy/          Governance module (13 msg types, 116 tests)
│   └── dex/                    DEX module (4 msg types, 24 tests)
├── treasury/keeper/            Tokenomics equations 1-5 (31 tests)
├── contracts/                  CosmWasm smart contracts (Rust)
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
| Node Staking Rewards (10% APY) | ✅ | `treasury/keeper/rewards.go` (eq.5) |
| Domain Interest (25% APY) | ✅ | `treasury/keeper/rewards.go` (eq.4) |
| Release Decay | ✅ | `treasury/keeper/rewards.go` |
| Anonymous Voting (WP S4) | ✅ | `x/truedemocracy/anonymity.go` |
| Stones Voting (WP S3.1) | ✅ | `x/truedemocracy/stones.go` |
| VoteToEarn Rewards | ✅ | `x/truedemocracy/stones.go` |
| Suggestion Lifecycle (WP S3.1.2) | ✅ | `x/truedemocracy/lifecycle.go` |
| Green/Yellow/Red Zones | ✅ | `x/truedemocracy/lifecycle.go` |
| Auto-Delete & Fast Delete (2/3) | ✅ | `x/truedemocracy/lifecycle.go` |
| Admin Election (WP S3.6) | ✅ | `x/truedemocracy/governance.go` |
| Member Exclusion (2/3 vote) | ✅ | `x/truedemocracy/governance.go` |
| PoD Transfer Limit (10%, WP S7) | ✅ | `x/truedemocracy/validator.go` |
| CLI Commands (14 tx + 6 query) | ✅ | `x/truedemocracy/cli.go` |
| DEX CLI (4 tx + 2 query) | ✅ | `x/dex/cli.go` |
| CosmWasm Contracts | ✅ | `contracts/src/` |
| Web Wallet (React + Keplr) | ✅ | `web-wallet/` |
| Mobile Wallet (Expo + RN) | ✅ | `mobile-wallet/` |
| CI/CD Workflows | ✅ | `.github/workflows/` |

---

## Build & Test

```bash
# Blockchain
go mod tidy
go build ./...
go test ./... -race -cover    # 182 tests

# Smart contracts
cd contracts && cargo build

# Web wallet
cd web-wallet && npm install && npm run build

# Mobile wallet
cd mobile-wallet && npm install
```

---

## Tech Stack

| Layer | Technology | Version |
|-------|-----------|---------|
| Consensus | CometBFT | v0.38.17 |
| Application | Cosmos SDK | v0.50.13 |
| Language | Go | 1.23.5 |
| Smart Contracts | CosmWasm (Rust) | cosmwasm-std 1.5 |
| Web Frontend | React + Tailwind CSS | 18.2 / 3.4 |
| Mobile | React Native + Expo | 0.74 / 51.0 |
| Wallet | Keplr + CosmJS | 0.32-0.38 |

---

## Current Status

**Version: v0.1-alpha**

- ✅ 182 unit tests passing across 3 modules (2,705 lines of test code)
- ✅ Core blockchain compiles and runs
- ✅ Whitepaper tokenomics fully implemented (equations 1-5)
- ✅ Complete governance system (domains, proposals, voting, lifecycle)
- ✅ DEX with AMM, liquidity pools, swap fees, PNYX burn
- ✅ Web wallet with 3-column governance UI
- ✅ Mobile wallet with bottom-tab navigation
- ✅ Comprehensive documentation (30+ guides)

### Roadmap

- **v0.2 (Q2 2025):** Full UI integration, ZKP for anonymity
- **v0.3 (Q3 2025):** Network scalability tests (175+ nodes), multi-asset DEX
- **v1.0 (Q4 2025):** Mainnet launch with full IBC support

---

## Contributing

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Ensure all tests pass: `make test`
5. Submit a pull request

See [Developer Docs](docs/developers/README.md) for architecture details.

## Community

- Telegram: [t.me/truerepublic](https://t.me/truerepublic)
- Issues: [github.com/NeaBouli/TrueRepublic/issues](https://github.com/NeaBouli/TrueRepublic/issues)
- Email: p.cypher@protonmail.com

## Contributors

- NeaBouli

## Donations

Team (BTC multi-sig): `bc1qyamf3twgcqckuqrvmwgwnhzupgshxs37eejdgl0ntcqve98qnvhqe6cjl9`
