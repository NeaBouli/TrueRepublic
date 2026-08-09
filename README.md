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
  <img src="https://img.shields.io/badge/tests-1579%20recovery--verified-orange" alt="Recovery-verified tests"/>
  <img src="https://img.shields.io/badge/version-v0.4.0-blue" alt="Version"/>
  <img src="https://img.shields.io/badge/recovery-active-orange" alt="Recovery active"/>
  <img src="https://img.shields.io/badge/Go-1.26.5-00ADD8?logo=go" alt="Go"/>
  <img src="https://img.shields.io/badge/Cosmos%20SDK-v0.50.15-5C4EE5" alt="Cosmos SDK"/>
  <img src="https://img.shields.io/badge/Rust-1.75+-orange?logo=rust" alt="Rust"/>
</p>

<p align="center">
  <a href="https://github.com/NeaBouli/TrueRepublic/actions/workflows/go-ci.yml"><img src="https://github.com/NeaBouli/TrueRepublic/actions/workflows/go-ci.yml/badge.svg" alt="Go CI"/></a>
  <a href="https://github.com/NeaBouli/TrueRepublic/actions/workflows/rust-ci.yml"><img src="https://github.com/NeaBouli/TrueRepublic/actions/workflows/rust-ci.yml/badge.svg" alt="Rust CI"/></a>
  <a href="https://github.com/NeaBouli/TrueRepublic/actions/workflows/react-ci.yml"><img src="https://github.com/NeaBouli/TrueRepublic/actions/workflows/react-ci.yml/badge.svg" alt="Web CI"/></a>
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
> `client-web` is the sole maintained web client. The legacy `web-wallet`
> and `mobile-wallet` prototypes were retired and removed under GH-112 and
> GH-102; Git history preserves them for audit only.

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
- [Cross-System Threat Model](docs/security/THREAT_MODEL.md)
- [Best Practices](https://github.com/NeaBouli/TrueRepublic/wiki/security-Best-Practices)

### Additional Docs

| Guide | Audience | Description |
|-------|----------|-------------|
| **[Getting Started](docs/getting-started/README.md)** | Everyone | Choose your path: user, operator, validator, or developer |
| **[Installation](INSTALLATION.md)** | Everyone | Quick install guide (Docker / native / maintained web client) |
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
| **IBC Transfers** | ICS-20 transfer module wired; two-chain and relayer evidence pending | [IBC Setup](docs/IBC_RELAYER_SETUP.md) |

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
├── go.mod / go.sum             Go module (SDK v0.50.15, CometBFT v0.38.25)
├── Makefile                    Build targets (build, test, lint, docker)
├── INSTALLATION.md             Quick install guide
├── x/
│   ├── truedemocracy/          Governance module (23 msg types, 533 test cases)
│   └── dex/                    DEX module (7 msg types, 138 test cases)
├── treasury/keeper/            Tokenomics equations 1-5 (36 test cases)
├── contracts/                  CosmWasm workspace (7 crates, 26 tests)
│   ├── core/                   Governance + treasury contracts
│   ├── packages/bindings/      TrueRepublic custom query/msg types
│   ├── packages/testing-utils/ Mock querier, AMM pool, fixtures
│   └── examples/               governance-dao, dex-bot, zkp-aggregator, token-vesting
├── client-web/                 React 18 + TypeScript + Vite + CosmJS (v0.4.0)
├── docs/
│   ├── getting-started/        Quick start guides
│   ├── user-manual/            End-user documentation (7 guides)
│   ├── node-operators/         Node setup, configuration, and operations
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
| IBC Transfer (ICS-20) | 🟡 Module wired; two-chain/relayer evidence pending | `app.go` (ibc-go v8.7.0) |
| Stones Voting (WP S3.1) | ✅ | `x/truedemocracy/stones.go` |
| VoteToEarn Rewards | ✅ | `x/truedemocracy/stones.go` |
| Suggestion Lifecycle (WP S3.1.2) | ✅ | `x/truedemocracy/lifecycle.go` |
| Green/Yellow/Red Zones | ✅ | `x/truedemocracy/lifecycle.go` |
| Auto-Delete & Fast Delete (2/3) | ✅ | `x/truedemocracy/lifecycle.go` |
| Admin Election (WP S3.6) | ✅ | `x/truedemocracy/governance.go` |
| Member Exclusion (2/3 vote) | ✅ | `x/truedemocracy/governance.go` |
| PoD Transfer Limit (10%, WP S7) | ✅ | `x/truedemocracy/validator.go` |
| CLI Commands (24 tx + 7 query) | ✅ | `x/truedemocracy/cli.go` |
| DEX CLI (7 tx + 9 query) | ✅ | `x/dex/cli.go` |
| CosmWasm Contracts (7 crates) | ✅ | `contracts/` (workspace) |
| Maintained Web Client | 🟡 Recovery verified | `client-web/` |
| Legacy Web Wallet | ⚫ Retired and removed under GH-112 | Git history only |
| Legacy Mobile Wallet | ⚫ Retired and removed under GH-102 | Git history only |
| CI/CD Workflows | ✅ | `.github/workflows/` |

---

## Build & Test

```bash
# Blockchain
go mod tidy
./scripts/go-packages.sh go build
./scripts/go-packages.sh go test -race -cover -count=1 -timeout=600s    # 1,439 Go cases

# Smart contracts
cd contracts && cargo test --workspace       # 26 tests

# Maintained web client
cd client-web && npm ci && npm run lint && npm test -- --run && npm run build
```

---

## Tech Stack

| Component | Version | Status |
|-----------|---------|--------|
| Cosmos SDK | v0.50.15 | Recovery verified |
| CometBFT | v0.38.25 | Recovery verified |
| CosmWasm | v0.53.4 | Recovery verified |
| ibc-go | v8.7.0 | Transfer wired; two-chain/relayer unverified |
| gnark (ZKP) | v0.14.0 | On-chain recovery verified; client disabled |
| Go | 1.26.5 | Recovery verified |
| Rust | 1.75+ | Contracts |
| React | 18.2 | Maintained v0.4 client |
| Native mobile client | — | Retired under GH-102; replacement pending |
| Local encrypted test wallet + CosmJS | 0.39 | Maintained v0.4 client |

**Known Limitations:** IBC staking/upgrade remains stubbed (PoD is used instead), a real ZKP prover/ceremony review is pending, and the migration-focused harness proves compatible binary replacement and fail-before-open rollback—not consensus-breaking state migration. GH-93 provides a strict synthetic incident-command and rehearsal contract, but a private live operator rehearsal remains pending. GH-85 provides dashboard/application runtime evidence, alert rules, recovery/testnet objectives, and role ownership. GH-89 adds a strict synthetic topology qualification contract. GH-97 adds bounded four-validator load, resource, retention, restart, and ledger evidence; it does not establish production sizing or multi-day soak behavior. GH-101 adds a digest-bound offline deployment-evidence envelope and verifier; it does not prove or perform a live deployment. Real seed/sentry/validator/RPC deployment, firewall/TLS/DNS evidence, external paging drills, private-environment capacity evidence, and independent live operations review remain open. See [LIMITATIONS.md](docs/LIMITATIONS.md).

---

## Current Status

**Version: v0.4.0 — recovery audit active; not production-ready**

The checklist below records implemented surface area, not a production security
approval. Current evidence, risks, and commands are maintained in
[`BRIDGE.md`](BRIDGE.md) and [GitHub issue #4](https://github.com/NeaBouli/TrueRepublic/issues/4).

- 🟡 1,606 tests recovery-verified locally (1,439 Go + 26 Rust + 141 maintained-client), plus the separately gated GH-145 bounded live fuzz campaigns, GH-131 real submitted-history pagination proof, GH-121 real browser-query boundary, GH-115 local client-chain delivery proof, GH-56 rotation, GH-59 slashing, GH-60 inactive-validator genesis, GH-61 legacy-authority migration, GH-93 incident rehearsal, and GH-97 sustained-load process harnesses; protected publication and production rollout evidence remain required
- ✅ Core blockchain compiles and runs
- 🟡 Tokenomics, exact custom genesis, and every-block ledger invariants are recovery-verified and merged through PR #19
- 🟡 Governance escrow/auth recovery is verified and merged; independent release review remains open
- 🟡 Groth16 voting backend tested; reward-recipient binding and real web proof generation remain open
- ✅ CosmWasm smart contract integration (wasmd v0.53.4)
- 🟡 Domain-Bank escrow recovery implemented and merged via PR #16
- 🟡 IBC transfer module wired on ibc-go v8.7.0; two-chain relayer,
  acknowledgement, timeout, replay, interruption, and upgrade evidence pending
- 🟡 Multi-Asset DEX bank custody, provider LP ownership, authority checks, and
  canonical burns are recovery-verified and merged via PR #18
- 🟡 GH-12 genesis/runtime conservation is recovery-verified and merged via PR #19
- 🟡 PR #23 provides generated CometBFT-key, bank-backed PoD genesis and
  proves native restart/export; GH-26 makes `scripts/init-node.sh` delegate only
  to that supported path. GH-32/GH-41/GH-43/GH-45/GH-53 add four-validator
  failure/restart/catch-up, partition, state-sync, and sanitized backup/restore
  gates plus compatible binary upgrade/fail-before-open rollback evidence;
  consensus-breaking migrations and broader multi-node operations remain open
- 🟡 ZKP UI is a clearly disabled preview until a compatible real Groth16 prover exists
- ✅ Developer Tooling: 4 CosmWasm example contracts, shared bindings, testing utils
- 🟡 DEX burns reduce canonical bank supply via merged PR #18
- ✅ Canonical v0.4 web client with 3-column governance UI
- ⚫ Legacy mobile wallet retired and removed under GH-102
- ✅ Comprehensive documentation (30+ guides)

### Roadmap

- ✅ **v0.1.x (Feb 2026):** Security fixes, documentation, elections
- ✅ **v0.2.x (Feb 2026):** Governance core — Systemic Consensing, Tokenomics, Elections
- 🟡 **v0.3.0 (Q1 2026): historical feature surface implemented; recovery verification incomplete**
  - ✅ Weeks 1-4: ZKP Anonymity Layer (Groth16, Merkle trees, nullifiers)
  - ✅ Week 5: CosmWasm Integration (wasmd v0.53.4, custom bindings)
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

> Historical test count: 577. The authoritative recovery-verified total is 1,606
> (1,439 Go + 26 Rust + 141 maintained-client), reproduced from fresh
> package-scoped JSON output on GH-169 using the established passing-case method.

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

Developer support (BTC): `bc1p2kh7reqf7l5zmssk8kdx2fx0q6suzfmsgdrphl9wjsxarns56jkq4tqvjm`
