<p align="center">
  <img src="https://raw.githubusercontent.com/NeaBouli/TrueRepublic/main/assets/logo.png" alt="TrueRepublic Logo" width="300"/>
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
  <img src="https://img.shields.io/badge/status-recovery%20active-orange" alt="Recovery active"/>
  <img src="https://img.shields.io/badge/main-577%20baseline%20tests-yellow" alt="Main baseline tests"/>
  <img src="https://img.shields.io/badge/PNYX%20cap-21%2C000%2C000-blue" alt="PNYX maximum supply"/>
  <img src="https://img.shields.io/badge/Go-1.24-00ADD8?logo=go" alt="Go"/>
  <img src="https://img.shields.io/badge/Cosmos%20SDK-v0.50.14-5C4EE5" alt="Cosmos SDK"/>
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

> [!WARNING]
> **Recovery audit active — not production-ready.** The default `main` branch is
> the preserved pre-recovery baseline. Security, token accounting, DEX custody,
> ZKP binding, genesis invariants, and the persistent node lifecycle are being
> recovered in the ordered draft-PR stack tracked by
> [issue #4](https://github.com/NeaBouli/TrueRepublic/issues/4). Do not use the
> current software, legacy wallets, real keys, or real funds in production.

> **Current recovery evidence:** 636 verified cases on the final stacked branch,
> including enforcement of the fixed **21,000,000 PNYX** maximum supply. The
> implementation remains unmerged pending ordered review. See
> [PR #24](https://github.com/NeaBouli/TrueRepublic/pull/24) and
> [`BRIDGE.md`](BRIDGE.md).

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

# Option C: Web wallet only (development baseline; never use real keys)
cd web-wallet && npm ci && npm start
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

## Web Client (v0.4.0)

React-based web client with full governance and DEX functionality:
```bash
cd client-web
npm ci
npm run dev
```

- Wallet: Create/import, encrypted storage, send PNYX
- Governance: Browse domains, vote anonymously, create suggestions
- DEX: Swap tokens, provide liquidity, manage LP positions
- Membership: Join domains, 2-step onboarding
- Admin: Domain management, member verification
- Explorer: Network stats, validators, blocks, IBC

See [`client-web/README.md`](client-web/README.md) for details.

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
├── client-web/                 React 18 + TypeScript + Vite + CosmJS (v0.4.0)
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

## Historical Feature Surface

This inventory describes code present in the pre-recovery baseline. A checkmark
does not imply production approval; current verification status is tracked below.

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

# Legacy web wallet (audit/build only; never use real keys)
cd web-wallet && npm ci && npm run build

# Legacy mobile wallet (audit only; never use real keys)
cd mobile-wallet && npm ci
```

---

## Tech Stack

| Component | Version | Status |
|-----------|---------|--------|
| Cosmos SDK | v0.50.14 | Baseline; recovery audit active |
| CometBFT | v0.38.21 | Baseline; recovery audit active |
| CosmWasm | v0.53.3 | Baseline; recovery audit active |
| ibc-go | v8.4.0 | Single-node recovery verified; relayer unverified |
| gnark (ZKP) | v0.9.x | Mock client; external review required |
| Go | 1.24 | Baseline toolchain |
| Rust | 1.75+ | Contracts |
| React | 18.2 | Web Wallet |
| React Native + Expo | 0.74 / 51.0 | Mobile |
| Keplr + CosmJS | 0.32-0.38 | Wallet |

**Known Limitations:** IBC staking/upgrade stubbed (PoD used instead), ZKP client integration v0.4.0. See [LIMITATIONS.md](docs/LIMITATIONS.md).

---

## Current Status

**Recovery audit active — the project is not production-ready.**

- `main`: preserved pre-recovery baseline with 577 historical test cases.
- Recovery stack: [#9](https://github.com/NeaBouli/TrueRepublic/pull/9) →
  [#15](https://github.com/NeaBouli/TrueRepublic/pull/15) →
  [#16](https://github.com/NeaBouli/TrueRepublic/pull/16) →
  [#17](https://github.com/NeaBouli/TrueRepublic/pull/17) →
  [#18](https://github.com/NeaBouli/TrueRepublic/pull/18) →
  [#19](https://github.com/NeaBouli/TrueRepublic/pull/19) →
  [#22](https://github.com/NeaBouli/TrueRepublic/pull/22) →
  [#23](https://github.com/NeaBouli/TrueRepublic/pull/23) →
  [#24](https://github.com/NeaBouli/TrueRepublic/pull/24).
- Recovery evidence: 636 verified cases (604 Go, 26 Rust, 6 maintained-client).
- Token invariant: fixed cap of **21,000,000 PNYX**
  (`21,000,000,000,000 upnyx`) is enforced and regression-tested on the stack.
- Remaining blockers: ordered review/merge, compatible real ZKP client prover,
  external cryptographic/consensus/operations review, and multi-node IBC/upgrade work.
- `client-web` is the maintained client. `web-wallet` and `mobile-wallet` are
  legacy clients and must not be used with real keys or funds.

Follow the live recovery record in
[`BRIDGE.md`](BRIDGE.md) and
[GitHub issue #4](https://github.com/NeaBouli/TrueRepublic/issues/4).

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
