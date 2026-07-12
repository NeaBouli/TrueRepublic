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
  <img src="https://img.shields.io/badge/tests-683%20recovery--verified-orange" alt="Recovery-verified tests"/>
  <img src="https://img.shields.io/badge/version-v0.4.0-blue" alt="Version"/>
  <img src="https://img.shields.io/badge/recovery-active-orange" alt="Recovery active"/>
  <img src="https://img.shields.io/badge/Go-1.26.5-00ADD8?logo=go" alt="Go"/>
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
> **Recovery audit active:** v0.4.0 functionality exists, but production-readiness
> claims are being re-verified in [GitHub issue #4](https://github.com/NeaBouli/TrueRepublic/issues/4).
> `client-web` is the maintained web client. `web-wallet` and `mobile-wallet`
> are legacy clients with unresolved high/critical dependency advisories and
> must not be used for real keys or production funds.

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

# Option C: Maintained v0.4 web client
cd client-web && npm ci && npm run dev
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
| **DEX (stacked recovery)** | PR #18 adds custody/LP ownership/burns; PR #19 reconciles genesis and checks reserves/shares every block | [DEX Guide](docs/user-manual/dex-trading-guide.md) |
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
- Governance: Browse domains and create suggestions; anonymous submission remains disabled until a real prover exists
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
│   ├── truedemocracy/          Governance module (23 msg types, 418 test cases)
│   └── dex/                    DEX module (7 msg types, 116 test cases)
├── treasury/keeper/            Tokenomics equations 1-5 (36 test cases)
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

## Implemented Features

| Feature | Status | Location |
|---------|--------|----------|
| Domains & Governance | ✅ | `x/truedemocracy/keeper.go` |
| Systemic Consensing (-5..+5) | ✅ | `x/truedemocracy/keeper.go` |
| Proof of Domain (PoD) | ✅ | `x/truedemocracy/validator.go` |
| Validator Slashing | ✅ | `x/truedemocracy/slashing.go` |
| Tokenomics (eq.1-5) | ✅ | `treasury/keeper/rewards.go` |
| DEX / AMM (x*y=k) | 🟡 Recovery verified on PR #19 | Atomic custody/burns plus exact genesis and every-block reserve/LP invariants |
| Multi-Asset DEX (BTC/ETH/LUSD) | 🟡 Recovery verified on PR #19 | Provider LP ownership and chain-authorized registry; recovery implementation merged to `main` |
| Node Staking Rewards (10% APY) | ✅ | `treasury/keeper/rewards.go` (eq.5) |
| Domain Interest (25% APY) | ✅ | `treasury/keeper/rewards.go` (eq.4) |
| Release Decay | ✅ | `treasury/keeper/rewards.go` |
| Anonymous Voting (WP S4) | ✅ | `x/truedemocracy/anonymity.go` |
| Zero-Knowledge Proofs (Groth16) | 🟡 Recovery verified on PR #22 | Chain/rating binding and fail-closed VK; real client prover and external review pending |
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
| Maintained Web Client | 🟡 Recovery verified | `client-web/` |
| Legacy Web Wallet | 🔴 Deprecated / vulnerable dependencies | `web-wallet/` |
| Legacy Mobile Wallet | 🔴 Deprecated / vulnerable dependencies | `mobile-wallet/` |
| CI/CD Workflows | ✅ | `.github/workflows/` |

---

## Build & Test

```bash
# Blockchain
go mod tidy
go build ./...
go test ./... -race -cover -count=1 -timeout=600s    # 649 tests

# Smart contracts
cd contracts && cargo test --workspace       # 26 tests

# Maintained web client
cd client-web && npm ci && npm run lint && npm test -- --run && npm run build
```

---

## Tech Stack

| Component | Version | Status |
|-----------|---------|--------|
| Cosmos SDK | v0.50.14 | Production |
| CometBFT | v0.38.22 | Recovery verified |
| CosmWasm | v0.53.3 | Production |
| ibc-go | v8.7.0 | Transfer Active |
| gnark (ZKP) | v0.14.0 | On-chain recovery verified; client disabled |
| Go | 1.26.5 | Recovery verified |
| Rust | 1.75+ | Contracts |
| React | 18.2 | Maintained v0.4 client |
| React Native + Expo | 0.74 / 51.0 | Legacy; security migration required |
| Keplr + CosmJS | 0.39 | Maintained v0.4 client |

**Known Limitations:** IBC staking/upgrade remains stubbed (PoD is used instead), a real ZKP prover/ceremony review is pending, and PR #23 still needs independent multi-node operations evidence. See [LIMITATIONS.md](docs/LIMITATIONS.md).

---

## Current Status

**Version: v0.4.0 — recovery audit active; not production-ready**

The checklist below records implemented surface area, not a production security
approval. Current evidence, risks, and commands are maintained in
[`BRIDGE.md`](BRIDGE.md) and [GitHub issue #4](https://github.com/NeaBouli/TrueRepublic/issues/4).

- 🟡 683 tests recovery-verified locally (649 Go + 26 Rust + 8 maintained-client)
- ✅ Core blockchain compiles and runs
- 🟡 Tokenomics, exact custom genesis, and every-block ledger invariants are locally verified through stacked PR #19
- 🟡 Governance surface implemented; escrow/auth recovery is in stacked review
- 🟡 Groth16 voting backend tested; reward-recipient binding and real web proof generation remain open
- ✅ CosmWasm smart contract integration (wasmd v0.53.3)
- 🟡 Domain-Bank escrow recovery implemented on stacked PR #16
- ✅ IBC Transfer module (ibc-go v8.4.0, cross-chain PNYX transfers)
- 🟡 Multi-Asset DEX bank custody, provider LP ownership, authority checks, and
  canonical burns are locally verified on stacked PR #18
- 🟡 GH-12 genesis/runtime conservation is locally verified on stacked PR #19
- 🟡 PR #23 locally replaces the legacy `x/staking` gentx path with generated
  CometBFT-key, bank-backed PoD genesis and proves native restart/export;
  GitHub Docker restart passes and independent multi-node operations evidence
  remains open
- 🟡 ZKP UI is a clearly disabled preview until a compatible real Groth16 prover exists
- ✅ Developer Tooling: 4 CosmWasm example contracts, shared bindings, testing utils
- 🟡 DEX burns reduce canonical bank supply on stacked PR #18
- ✅ Canonical v0.4 web client with 3-column governance UI
- 🔴 Legacy mobile wallet is deprecated and security-blocked
- ✅ Comprehensive documentation (30+ guides)

### Roadmap

- ✅ **v0.1.x (Feb 2026):** Security fixes, documentation, elections
- ✅ **v0.2.x (Feb 2026):** Governance core — Systemic Consensing, Tokenomics, Elections
- 🟡 **v0.3.0 (Q1 2026): historical feature surface implemented; recovery verification incomplete**
  - ✅ Weeks 1-4: ZKP Anonymity Layer (Groth16, Merkle trees, nullifiers)
  - ✅ Week 5: CosmWasm Integration (wasmd v0.53.3, custom bindings)
  - ✅ Week 6: Domain-Bank Bridge (dual accounting, deposit/withdraw)
  - ✅ Week 7: IBC Integration (ICS-20 transfer, relayer support)
  - ✅ Week 8: Multi-Asset DEX (asset registry, trading validation, symbol resolution)
  - ✅ Week 9: Cross-Chain Liquidity (multi-hop swaps, analytics)
  - ✅ Week 10: UI Components (ZKP voting, DEX analytics)
  - ✅ Week 11: Developer Tooling (contract examples, testing utils)
  - ✅ Week 12: Complete Documentation (API, deployment, architecture)
- 🟡 **v0.4.0 (recovery audit active since July 2026): Web Client**
  - ✅ Wallet Foundation (create/import/encrypt/send)
  - ✅ Governance UI (domains, issues, suggestions, stones)
  - ✅ DEX Interface (swap, liquidity, LP positions)
  - 🟡 ZKP Anonymous Voting (on-chain binding verified; mock submission disabled; real prover pending)
  - ✅ Domain Membership & Onboarding
  - ✅ Admin Dashboard (member management, stats)
  - ✅ Network Explorer (validators, blocks, IBC)
- 📋 **v0.5.0 (Q3 2026):** Native Apps (iOS/Android)
- 🎯 **v1.0.0 (Q4 2026):** Production Release — External audit, mainnet launch

> Historical test count: 577. The authoritative recovery-verified total is 683
> (649 Go + 26 Rust + 8 maintained-client), reproduced locally on the current branch.

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
