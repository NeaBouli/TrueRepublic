# Project Handover -- TrueRepublic / PNYX

## Current Status

**Version:** v0.3.0 (100% complete) -- 02.03.2026
**Phase:** v0.3.0 COMPLETE. All 12 weeks finished: ZKP + CosmWasm + Bank Bridge + IBC + Multi-Asset DEX + Cross-Chain Liquidity + UI Components + Developer Tooling + Documentation. Zero P0 issues.

### Repository State

| Repo | Branch | HEAD | Path |
|------|--------|------|------|
| **Main** | `main` | `500c251` (feat: logos & icons) | `/Users/gio/TrueRepublic/` |
| **Wiki** | `master` | `09276ab` (docs: update wiki for v0.3.0 100% completion) | `/Users/gio/TrueRepublic/wiki-github/` |

- Working tree: **clean**, up-to-date with `origin/main`
- `wiki-github/` is untracked in the main repo (it is a separate git clone of the GitHub Wiki repo -- this is expected and correct)
- Remote: `https://github.com/NeaBouli/TrueRepublic`

### Releases (16 total)

| Tag | Title | Date |
|-----|-------|------|
| v0.1.1 | Security Fixes | 21.02.2026 |
| v0.1.2 | Supply Correction & Version Consistency | 21.02.2026 |
| v0.1.3 | Whitepaper Markdown conversion | 22.02.2026 |
| v0.1.4 | Documentation completeness & consistency | 22.02.2026 |
| v0.1.5 | Logo integration | 22.02.2026 |
| v0.1.6 | Maintenance & Verification | 22.02.2026 |
| v0.1.7 | Pre-Production Baseline | 22.02.2026 |
| v0.1.8 | Person Election Voting Modes | 24.02.2026 |
| v0.2.0 | Feature Complete Governance | 24.02.2026 |
| v0.2.1 | Documentation Sync | 24.02.2026 |
| v0.2.2 | v0.3.0 Roadmap & Documentation | 24.02.2026 |
| v0.2.3 | Documentation Sync | 25.02.2026 |
| v0.2.4 | Documentation Sync | 25.02.2026 |
| v0.2.5 | Documentation Sync | 25.02.2026 |
| v0.3.0-dev | Anonymity Layer Foundation | 25.02.2026 |
| v0.3.0 | ZKP Anonymity & Multi-Asset DEX (Latest) | 02.03.2026 |

### Live Deployments

- **GitHub Pages:** https://neabouli.github.io/TrueRepublic/ (deployed from `/docs` on `main`)
- **Wiki:** https://github.com/NeaBouli/TrueRepublic/wiki (30 pages, synced from `wiki-github/`)
- **Releases:** https://github.com/NeaBouli/TrueRepublic/releases

### Key Metrics

- 577 tests across 3 languages (533 Go + 26 Rust + 18 Frontend)
- 29 transaction types (23 governance + 6 DEX)
- 12 query endpoints (7 governance + 5 DEX)
- 5 tokenomics equations fully implemented + domain interest in EndBlock
- CosmWasm: 7 custom queries, 5 custom messages for smart contracts
- IBC: ICS-20 Transfer module (ibc-go v8.4.0), cross-chain PNYX transfers
- ~14,000 lines of source code (Go + JS + Rust)
- 30+ wiki pages, 39 docs files

### Completed Work (v0.1.1 -- v0.2.0)

1. **v0.1.1 -- Security:** 5 CVEs patched (curve25519-dalek RUSTSEC-2024-0344, CometBFT GO-2026-4361, jose2go GO-2025-4123, crypto/tls GO-2026-4337, Go 1.24.13)
2. **v0.1.2 -- Supply Fix:** Critical 22M to 21M correction in source code and all documentation; GitHub Pages deployed
3. **v0.1.3 -- Whitepaper:** PDF converted to Markdown (`docs/WhitePaper_TR_eng.md`); appendix 8.1.1 corrected; old PDF with 22M error removed
4. **v0.1.4 -- Documentation:** 8 wiki stub pages filled (+1,222 lines); 3 P0 inconsistencies fixed (Tendermint naming x2, PNYX/USDC x2)
5. **v0.1.5 -- Logos:** 3 official logos integrated across GitHub Pages, README, Whitepaper, Wiki
6. **v0.1.6 -- Verification:** Frontend independence confirmed (no Telegram FOSS code); project integrity verified
7. **v0.1.7 -- Baseline:** Pre-production baseline release marking end of audit cycle
8. **v0.1.8 -- Elections:** Person election voting modes (WP §3.7): Simple Majority, Absolute Majority, Systemic Consensing, Abstention; `VotingMode`/`VoteChoice` types, `CastElectionVote`/`TallyElection` logic, 15 new tests (197 total)
9. **v0.2.0 -- Feature Complete Governance:** 4 milestones delivering ~95% whitepaper coverage:
   - **Systemic Consensing Score** (WP §3.2): `scoring.go` with `ComputeSuggestionScore`, `RankSuggestionsByScore`, `FindConsensusWinner`; 15 tests verifying whitepaper table
   - **MsgRateProposal**: on-chain anonymous rating with ed25519 domain key signature verification; `RateProposalWithSignature` keeper, msg_server handler, gRPC handler, CLI command
   - **MsgCastElectionVote**: on-chain person election voting (approve/abstain); msg_server handler, gRPC handler, CLI command
   - **Domain Interest EndBlock** (eq.4): `DistributeDomainInterest()` runs every RewardInterval, credits active domain treasuries with interest capped by payouts, decays with release; 5 tests
10. **v0.3.0 Week 7 -- IBC Integration:** 2 milestones delivering cross-chain PNYX transfers:
    - **Milestone 7.1 -- IBC Transfer Module** (`19f774a`): ParamsKeeper, CapabilityKeeper, IBCKeeper, TransferKeeper wired in app.go; IBCStakingKeeper stub (3-week unbonding) + IBCUpgradeKeeper stub (7 no-op methods) in `ibc_stubs.go`; IBC Router with transfer route; BeginBlocker for IBC client updates; refactored InitChainer to use module manager with default genesis filling; wasm keeper updated with real IBC keepers (5 stubs removed from `wasm_stubs.go`); 9 tests
    - **Milestone 7.2 -- Relayer Configuration** (`1f3935e`): `docs/IBC_RELAYER_SETUP.md` (~300 line Hermes guide); genesis default filling for IBC modules; transfer port binding at InitGenesis; 6 integration tests (denom trace, escrow, genesis, params, keys, port); README updated with IBC features
11. **v0.3.0 Week 8 -- Multi-Asset DEX:** 2 milestones + CI fix delivering multi-asset trading:
    - **Milestone 8.1 -- Asset Registry System** (`63a5405`): RegisteredAsset type, RegisterAsset/DeregisterAsset/GetAssetByDenom/GetAssetBySymbol/GetAllAssets/UpdateAssetTradingStatus keeper methods, ValidateBasic, MsgRegisterAsset + MsgUpdateAssetStatus with gRPC handlers, 5 query endpoints, CLI commands (register-asset, update-asset-status, registered-assets, asset), DefaultGenesisState with PNYX + ATOM; 19 new tests
    - **Milestone 8.2 -- DEX Integration** (`be004ff`): validateAssetForTrading in CreatePool + Swap, GetSymbolForDenom, Pool.AssetSymbol display field, CLI symbol resolution (resolveSymbolOrDenom), event enrichment (asset_symbol, input_symbol, output_symbol); 10 new tests
    - **CI Fix** (`402420f`): Cosmos SDK v0.50.14, wasmd v0.53.3, xz v0.5.15; test timeout 600s
12. **v0.3.0 Week 9 -- Cross-Chain Liquidity** (`1f0fdda`, `fe1da0d`): Multi-hop swap routing, pool analytics (volume, fees, liquidity depth), slippage protection, SpotPrice queries, LP position tracking; 38 new tests (481->519)
13. **v0.3.0 Week 10 -- UI Components** (`fac91f6`): 8 React components (3 ZKP + 5 DEX analytics), legacy querier extensions (+8 routes), 14 Go querier tests + 18 frontend component tests (519->551)
14. **v0.3.0 Week 11 -- Developer Tooling** (`815e897`): Cargo workspace with 7 crates: packages/bindings (TrueRepublic custom query/msg types), packages/testing-utils (mock querier, AMM pool, fixtures), 4 example contracts (governance-dao, dex-bot, zkp-aggregator, token-vesting); CI updated with --workspace flags; 26 Rust tests (551->577)
15. **v0.3.0 Week 12 -- Final Documentation** (`49d875e`): API_REFERENCE.md (complete API overview, all endpoints, CosmWasm bindings, error codes), DEPLOYMENT.md (production setup, systemd, security hardening, IBC relayer), ARCHITECTURE.md (module architecture, ZKP circuit, data flow, security), CONTRIBUTING.md (workflow, testing, code review), QUICKSTART.md (5-minute getting started); 5 new files, ~1,200 lines
16. **Logo & Brand Integration** (`500c251`): Official TrueRepublic owl logo + PNYX token icon (1024x1024 each); placed in assets/, docs/assets/images/, web-wallet/public/, web-wallet/src/assets/; README header with centered logo + version badges; docs/index.html favicon + header + hero + footer; web-wallet PWA manifest + favicon; PoolStats PNYX icon; build verified (788 kB)

---

## Architecture Decisions

### Blockchain Stack

| Layer | Choice | Version | Rationale |
|-------|--------|---------|-----------|
| Consensus | CometBFT | v0.38.21 | BFT consensus with fast finality; Cosmos ecosystem standard |
| Application Framework | Cosmos SDK | v0.50.14 | Module-based architecture, IBC support, mature tooling |
| Language | Go | 1.24 (CI: 1.24.13) | Performance, concurrency, Cosmos SDK native language |
| IBC | ibc-go | v8.4.0 | Cross-chain transfers via ICS-20 |
| Smart Contracts | CosmWasm (Rust) | cosmwasm-std 3 | Deterministic WASM execution, memory-safe |
| Web Frontend | React 18 + Tailwind 3.4 | 18.2 / 3.4.19 | Component-based UI, utility-first CSS |
| Mobile | React Native + Expo | 0.74 / 51.0 | Cross-platform mobile from shared JS codebase |
| Wallet Integration | Keplr + CosmJS | 0.32-0.38 | Standard Cosmos wallet; CosmJS for chain interaction |

### Module Architecture

Two custom Cosmos SDK modules plus a treasury package:

1. **x/truedemocracy** (~12,200 lines, 38 files) -- Governance: domains, proposals, systemic consensing scoring (-5 to +5), stones voting, suggestion lifecycle (green/yellow/red zones), validator PoD, slashing, anonymous voting (domain key signatures + ZKP membership proofs), admin elections, member exclusion, person election voting modes (simple/absolute majority, abstention), domain interest payout (eq.4), two-step onboarding (add member + domain key registration), Big Purge EndBlock execution, ZKP anonymous voting (MiMC Merkle tree, Groth16 membership proofs, identity commitments, nullifier store, MsgRateWithProof), CosmWasm custom bindings (7 queries + 5 messages), Domain-Bank bridge (deposit/withdraw with dual accounting)
4. **IBC** (ibc-go v8.4.0) -- ICS-20 Transfer module for cross-chain PNYX transfers; ParamsKeeper, CapabilityKeeper, IBCKeeper, TransferKeeper; IBCStakingKeeper/IBCUpgradeKeeper stubs in `ibc_stubs.go`; IBC Router; BeginBlocker for client updates
2. **x/dex** (~2,200 lines, 12 files) -- Multi-asset AMM DEX: constant-product (x*y=k), asset registry, trading validation, symbol resolution, 0.3% swap fee, 1% PNYX burn
3. **treasury/keeper** (371 lines, 2 files) -- Tokenomics equations 1-5: domain cost, rewards, put price, domain interest (25% APY), node staking (10% APY), release decay

### Critical Constants (DO NOT change without core dev approval)

```go
// treasury/keeper/rewards.go
SupplyMax int64 = 21_000_000    // Fixed max PNYX supply -- was incorrectly 22M before v0.1.2
CDom      int64 = 2             // Domain cost multiplier
CPut      int64 = 15            // Put price cap
CEarn     int64 = 1000          // Reward division factor
StakeMin  int64 = 100_000       // Min node stake in PNYX
ApyDom  = 0.25                  // 25% domain treasury APY
ApyNode = 0.10                  // 10% node staking APY
```

### Frontend Decision: 100% Custom (No Telegram FOSS)

Verified in v0.1.6: Both frontends are entirely custom-built. The "Telegram-inspired" label in commit `e71be84` refers only to the 3-column layout design pattern, not to any Telegram source code. Zero Telegram imports, dependencies, or code patterns exist anywhere.

### DEX Scope

- **Current:** Multi-asset DEX with asset registry. PNYX + ATOM in default genesis. BTC/ETH/LUSD via RegisterAsset.
- Asset validation: CreatePool + Swap require registered + trading-enabled assets
- Symbol resolution: CLI resolves symbols (BTC) -> IBC denoms via chain query
- The wiki `users-System-Overview.md` previously said "PNYX/USDC" -- corrected to "PNYX/ATOM; BTC, ETH, LUSD planned via IBC" in v0.1.4

### Naming: CometBFT (not Tendermint)

The consensus engine was rebranded from Tendermint to CometBFT. All documentation references have been updated. The only remaining "Tendermint" references are in Go dependency paths (`github.com/cometbft/cometbft`) and historical commit messages, which is correct.

---

## Project Structure

```
TrueRepublic/
├── app.go                          Cosmos SDK app entry point (~350 lines)
│                                   Wires truedemocracy + dex + IBC + CosmWasm modules, ABCI handlers
├── ibc_stubs.go                    IBCStakingKeeper (2 methods) + IBCUpgradeKeeper (7 methods)
├── ibc_test.go                     15 IBC tests (stubs, module config, integration)
├── go.mod / go.sum                 Go 1.23.5, toolchain go1.24.1
├── Makefile                        build, test, lint, docker-build, docker-up/down
├── Dockerfile                      Multi-stage: Go 1.23-alpine -> Alpine 3.19
├── docker-compose.yml              6 services: node, web-wallet, nginx, prometheus, grafana
├── .env.example                    Chain config template (CHAIN_ID, ports, etc.)
├── README.md                       Project overview (252 lines)
├── INSTALLATION.md                 Quick install guide
├── SECURITY.md                     Security documentation
├── CLAUDE.md                       THIS FILE -- project handover context
│
├── x/truedemocracy/                GOVERNANCE MODULE (35 files, ~10,800 lines)
│   ├── keeper.go                   Domain CRUD, proposal submission, fee validation,
│   │                               RateProposalWithSignature (anonymous rating)
│   ├── msg_server.go               Message handlers (23 tx types)
│   ├── query_server.go             gRPC query handlers (7 query types)
│   ├── cli.go                      Cobra CLI commands (24 tx + 7 query)
│   ├── module.go                   Module registration, codecs, EndBlock hooks
│   ├── msgs.go                     Message type definitions (23 types)
│   ├── types.go                    Domain, DomainOptions, VotingMode, VoteChoice,
│   │                               NullifierRecord structs
│   ├── scoring.go                  Systemic Consensing: ComputeSuggestionScore,
│   │                               RankSuggestionsByScore, FindConsensusWinner (WP §3.2)
│   ├── governance.go               Admin election, member exclusion
│   ├── election.go                 Person election voting: CastElectionVote, TallyElection (WP §3.7)
│   ├── validator.go                PoD consensus, registration, transfer limit (10%),
│   │                               DistributeDomainInterest (eq.4)
│   ├── slashing.go                 5% double-sign, 1% downtime penalties
│   ├── stones.go                   Stones voting + VoteToEarn rewards
│   ├── lifecycle.go                Suggestion lifecycle (green/yellow/red, auto-delete)
│   ├── anonymity.go                Domain key pairs, identity commitment registration,
│   │                               nullifier KV store for anonymous voting (WP S4)
│   ├── big_purge.go                EndBlock Big Purge: permission reg + identity commits
│   │                               + nullifiers cleanup, announcement tracking (WP S4)
│   ├── merkle.go                   MiMC Merkle tree (depth 20, BN254), commitment/nullifier
│   │                               computation, proof generation/verification
│   ├── zkp.go                      Groth16 membership proof circuit (BN254/MiMC),
│   │                               setup, prove, verify, serialization
│   ├── crypto.go                   Ed25519 dual-key derivation (global + domain keys)
│   ├── tree.go                     Tree data structures
│   ├── querier.go                  Legacy query interface
│   ├── wasm_bindings.go            CosmWasm custom query/message bindings (7 queries, 5 msgs)
│   ├── treasury_bridge.go          Domain-Bank bridge: DepositToDomain, WithdrawFromDomain
│   ├── *_test.go (19 files)        367 tests: governance, validator, stones, lifecycle,
│   │                               anonymity, slashing, elections, scoring, domain interest,
│   │                               crypto (dual-key onboarding), big purge EndBlock,
│   │                               onboarding (two-step flow), Merkle tree, ZKP circuit,
│   │                               identity commitments + nullifier store, ZKP voting
│   │                               (MsgRateWithProof, E2E flow, Big Purge cycle),
│   │                               ZKP queries (nullifier, purge schedule, ZKP state),
│   │                               Merkle root history, genesis round-trip,
│   │                               CosmWasm bindings (query + msg encoder),
│   │                               treasury bridge (deposit, withdraw, round-trip)
│
├── x/dex/                          DEX MODULE (12 files, ~2,200 lines)
│   ├── keeper.go                   AMM pool ops (x*y=k), asset validation, symbol resolution
│   ├── asset_registry.go           RegisterAsset, DeregisterAsset, GetAssetByDenom/Symbol
│   ├── msg_server.go               6 msg handlers (CreatePool, Swap, AddLiquidity,
│   │                               RemoveLiquidity, RegisterAsset, UpdateAssetStatus)
│   ├── query_server.go             5 query handlers (Pool, Pools, RegisteredAssets,
│   │                               AssetByDenom, AssetBySymbol)
│   ├── cli.go                      DEX CLI (6 tx + 4 query) with symbol resolution
│   ├── module.go                   DEX module setup
│   ├── msgs.go / types.go          Message + type definitions (RegisteredAsset, Pool)
│   ├── querier.go                  Legacy queries
│   ├── keeper_test.go              24 tests (pool, swap, liquidity, fees, burn)
│   ├── asset_registry_test.go      14 tests (register, deregister, validation, genesis)
│   ├── multi_asset_test.go         10 tests (trading validation, symbols, multi-pool)
│
├── treasury/keeper/                TOKENOMICS (2 files, 371 lines)
│   ├── rewards.go                  Equations 1-5: domain cost, rewards, put price,
│   │                               domain interest, node staking (with decay)
│   ├── rewards_test.go             31 tests
│
├── contracts/                      COSMWASM WORKSPACE (7 crates, 26 Rust tests)
│   ├── Cargo.toml                  Workspace root (resolver = "2")
│   ├── core/                       Governance + treasury contracts (from original src/)
│   │   ├── src/lib.rs              Module exports
│   │   ├── src/governance.rs       Governance contract (systemic consensing)
│   │   └── src/treasury.rs         Treasury contract (balance management)
│   ├── packages/bindings/          TrueRepublic custom query/msg types
│   │   └── src/                    7 queries + 5 messages mirroring wasm_bindings.go
│   ├── packages/testing-utils/     Mock querier, AMM pool, test fixtures (4 tests)
│   ├── examples/governance-dao/    DAO proposal lifecycle (5 tests)
│   ├── examples/dex-bot/           Limit orders, AMM simulation (6 tests)
│   ├── examples/zkp-aggregator/    Anonymous vote aggregation (5 tests)
│   └── examples/token-vesting/     Linear vesting with cliff (6 tests)
│
├── web-wallet/                     WEB FRONTEND (13 files, 1,053 lines)
│   ├── public/                     logo.png, logo192.png, logo512.png, manifest.json
│   │   └── assets/pnyx-icon.png    PNYX icon (public)
│   ├── src/App.js                  Root component, domain/proposal state
│   ├── src/assets/images/          pnyx-icon.png (React import)
│   ├── src/components/             ProposalFeed, DomainInfo, DomainList, Header,
│   │                               ThreeColumnLayout
│   ├── src/pages/                  Dex, Wallet, Governance
│   ├── src/services/api.js         CosmJS blockchain API integration
│   ├── src/hooks/useWallet.js      Keplr wallet hook
│   ├── package.json                React 18.2, Tailwind 3.4, CosmJS 0.32-0.38
│   ├── Dockerfile                  Node 20 builder -> nginx
│   ├── nginx.conf                  Frontend server config
│
├── mobile-wallet/                  MOBILE FRONTEND (4 files, 282 lines)
│   ├── src/App.js                  Bottom-tab navigation (Wallet/Governance/DEX)
│   ├── src/screens/                WalletScreen, GovernanceScreen, DexScreen
│   ├── package.json                React Native 0.74, Expo 51, CosmJS
│   ├── app.json                    Expo configuration
│
├── docs/                           DOCUMENTATION (39 files)
│   ├── index.html                  GitHub Pages landing site (454 lines)
│   ├── assets/                     logo.png, pnx_logo.png, pnx_ticker.png
│   │   └── images/                 logo.png (owl), pnyx-icon.png (for GitHub Pages)
│   ├── WhitePaper_TR_eng.md        English whitepaper (~430 lines, corrected 21M, §3.7 added)
│   ├── WhitePaper_TR.md            German whitepaper
│   ├── API.md                      REST/RPC/CLI reference
│   ├── FAQ.md                      30+ Q&A
│   ├── GLOSSARY.md                 60+ term definitions
│   ├── INSTALL.md / DEPLOYMENT.md  Setup and deployment guides
│   ├── ARCHITECTURE.md             System architecture overview
│   ├── IBC_RELAYER_SETUP.md        Hermes relayer configuration guide (~300 lines)
│   ├── SESSION_SUMMARY_2026-02-28.md  Session summary (Weeks 1-7)
│   ├── VALIDATOR_GUIDE.md          Validator operations
│   ├── getting-started/            Quick start guide
│   ├── user-manual/                7 end-user guides (governance, DEX, voting, etc.)
│   ├── node-operators/             9 guides (install, config, operations)
│   ├── developers/                 8 guides (architecture, API, integration, contracts)
│   ├── validators/                 Validator guide with PoD
│
├── assets/                         PROJECT LOGOS & BRAND ASSETS
│   ├── asset_icon_logo.png         Official owl logo (809K, 1024x1024)
│   ├── asset_icon_pnyx.png         PNYX token icon (1.6M, 1024x1024, transparent)
│   ├── logo.png                    TrueRepublic logo (938K)
│   ├── pnx_logo.png               PNYX coin logo (1.1M)
│   ├── pnx_ticker.png             PNYX ticker icon (1.1M)
│
├── wiki-github/                    GITHUB WIKI (separate git repo, 30 .md files)
│   ├── Home.md                     Wiki landing page with navigation
│   ├── _Sidebar.md                 Sidebar navigation
│   ├── develop-*.md (8)            Architecture, Code Structure, Module Deep-Dive,
│   │                               API Reference, Dev Setup, Contributing, Smart
│   │                               Contracts, Frontend Architecture
│   ├── users-*.md (6)              System Overview, Installation, Manuals, How It
│   │                               Works, Frontend Guide, FAQ
│   ├── operations-*.md (5)         Node Setup, Validator Guide, Deployment, Monitoring,
│   │                               Troubleshooting
│   ├── security-*.md (5)           Security Architecture, Audit Reports, Test Coverage,
│   │                               Known Issues, Best Practices
│   ├── status-*.md (5)             Current Status, Roadmap, Feature Matrix, Testing
│   │                               Status, Known Bugs (STUB)
│
├── monitoring/                     Prometheus + Grafana configs
├── nginx/                          Reverse proxy config (rate limiting, CORS)
├── scripts/                        Build and deployment scripts
├── ui/                             UI design assets
│
├── .github/workflows/              CI/CD (5 workflows)
│   ├── go-ci.yml                   Go build + vet + test (Go 1.24.13)
│   ├── rust-ci.yml                 cargo fmt + clippy + build + test
│   ├── react-ci.yml                npm install + build + test (Node 20)
│   ├── react-native-ci.yml         npm ci + jest (Node 20)
│   ├── security-scan.yml           Weekly: govulncheck, cargo audit, npm audit
│
├── .gitignore                      OS, Go, Rust, Node, env, IDE, chain data exclusions
└── LICENSE                         Apache 2.0
```

---

## Open TODOs

### Wiki & Docs

- All 30 wiki pages complete (zero stubs)
- Roadmap dates updated to 2026/2027
- GitHub Pages shows 577 tests, 100% complete
- Full v0.3.0 roadmap spec at `docs/V0.3.0_ROADMAP.md`

### v0.3.0: COMPLETE (all 12 weeks finished)

### v0.4.0 Scope (Q2 2026) — Optional Indexer Stack

**Full Specification:** `docs/V0.4.0_OPTIONAL_INDEXER_STACK.md` (to be created)
**Critical:** Entire stack is OPTIONAL — zero consensus impact, fail-safe architecture.

- **Phase 1 (P0):** Snapshot Indexer — separate Go service, Postgres, event consumption via RPC, idempotent upserts (+15-20 tests)
- **Phase 2 (P1):** Full-History Mode — history tables, rebuild/backfill CLI, deterministic replay (+10-15 tests)
- **Phase 3 (P0):** Read-Only API — REST endpoints, OpenAPI spec, rate limiting, read-only DB user (+8-10 tests)
- **Phase 4 (P1):** Minimal Explorer — Next.js dashboard, domain/proposal browser, NO wallet (+5 tests)
- **Phase 5 (P2):** Wallet Actions — requires external security review first, Keplr/Leap, TX signing (+20 tests)

Security: 10-point threat model required, DB/API down → chain unaffected, read-only by default.

### v0.5.0 Scope (Q3 2026) — DEX Expansion

- BTC/ETH/LUSD pools via IBC
- Cross-chain swaps
- Liquidity analytics dashboard

### v1.0.0 Scope (Q4 2026) — Production

- External security audit (full scope)
- Mainnet launch preparation
- Genesis ceremony and validator onboarding
- Production deployment with monitoring/alerting

### No LICENSE File Confirmed

The README references Apache 2.0, but the presence of an actual `LICENSE` file in the repo root should be verified. If missing, it should be added.

---

## Known Issues

### Zero P0 Issues

As of v0.2.0, there are no known critical (P0) issues. All previously identified P0 issues were resolved:
- 22M supply bug (fixed v0.1.2)
- Tendermint naming in INSTALL/DEPLOYMENT (fixed v0.1.4)
- PNYX/USDC in System-Overview (fixed v0.1.4)

### Low-Priority Items

1. **Go version:** `go.mod` says `go 1.24` (bumped from 1.23.5 in v0.3.0 Week 2 for gnark ZKP dependency). CI runs `1.24.13`.
2. **No explicit linting configs:** No `.eslintrc`, `.prettierrc`, or `golangci-lint.yml` exist. Go uses `go vet` + optional staticcheck. Rust enforces `cargo fmt` + `clippy` in CI. JavaScript has no enforced linting.
3. **Test count in docs/index.html:** Updated to 577 tests, 100% v0.3.0 complete (as of Week 12).
4. **No protobuf generation:** The Makefile has a `proto-gen` stub target but no actual protobuf schema files or generation pipeline. Messages are defined as Go structs directly.
5. **Web wallet polyfills:** The web wallet requires Node.js polyfills for `@cosmjs/crypto` (webpack 5 removed them). This is handled by `react-app-rewired` with `config-overrides.js`.

---

## Explicit Non-Goals

1. **No Telegram FOSS code:** The frontend is 100% custom. The "Telegram-inspired" label refers only to the 3-column layout design pattern. This was explicitly verified and confirmed in v0.1.6. Do not introduce Telegram dependencies.
2. **No multi-asset DEX in v0.2.x:** Only PNYX/ATOM is supported. BTC, ETH, LUSD pools require IBC and are planned for v0.3.
3. **No protobuf/gRPC schema files:** The project uses direct Go struct definitions for messages, not `.proto` files. This is a deliberate simplification for the alpha phase.
4. **No i18n/localization:** Documentation and UI are English-only (with a German whitepaper variant).
5. **No mainnet deployment:** The project is testnet-ready only. Mainnet launch is planned for v1.0.
6. **No external audit yet:** Security scanning is automated (govulncheck, cargo audit, npm audit) but no professional third-party audit has been conducted.
7. **Do not change the 21M supply constant:** The `SupplyMax` value was corrected from 22M to 21M in v0.1.2 after careful analysis. The whitepaper section 3.4.2 ("21 million") is the authoritative source; appendix 8.1.1 had a typo that has been corrected in the Markdown version.

---

## Code Conventions

### Go (Blockchain)

- **Module path:** `truerepublic` (import as `truerepublic/x/truedemocracy`, `truerepublic/x/dex`, `truerepublic/treasury/keeper`)
- **File naming:** Lowercase, descriptive: `keeper.go`, `msg_server.go`, `lifecycle.go`
- **Test files:** `*_test.go` in same package, table-driven with `t.Run()` subtests
- **Error handling:** `errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "description")`
- **Store keys:** Prefix-based byte keys: `[]byte("domain:" + name)`, `[]byte("pool:" + denom)`
- **Serialization:** Legacy Amino codec with `MustMarshalLengthPrefixed` / `MustUnmarshalLengthPrefixed`
- **Math:** `cosmossdk.io/math` for all integer/decimal operations, never native int for token amounts
- **CLI:** Cobra commands via `GetTxCmd()` and `GetQueryCmd()` per module
- **Build:** `make build` produces `./build/truerepublicd`, version injected via `-ldflags`
- **Testing:** `go test ./... -race -cover -count=1`

### Rust (Smart Contracts)

- **Framework:** cosmwasm-std 3 with `#[entry_point]` macros
- **Entry points:** `instantiate`, `execute`, `query` per contract
- **State:** JSON serialization with `serde`, manual `save_state`/`load_state` helpers
- **Derive macros:** `Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema`
- **CI enforced:** `cargo fmt --check` and `clippy -D warnings`

### JavaScript (Web Wallet)

- **Framework:** React 18.2 with functional components and hooks
- **State:** `useState` + `useCallback` + `useEffect` (no Redux or Context)
- **Styling:** Tailwind CSS 3.4 utility classes
- **API:** Service abstraction in `src/services/api.js` using CosmJS
- **Wallet:** Custom `useWallet` hook wrapping Keplr integration
- **Build:** `react-app-rewired` (CRA with webpack overrides for Node.js polyfills)

### JavaScript (Mobile Wallet)

- **Framework:** React Native 0.74 + Expo 51
- **Navigation:** `@react-navigation/bottom-tabs` (Wallet / Governance / DEX)
- **Theme:** Dark mode: background `#1F2937`, tabs `#374151`, active `#3B82F6`

### Git & Release Workflow

- **Main repo commits:** From `/Users/gio/TrueRepublic/`, push to `origin/main`
- **Wiki commits:** From `/Users/gio/TrueRepublic/wiki-github/`, push to `origin/master` (different branch name)
- **Wiki file naming:** Flat structure with `section-PageName.md` (e.g., `develop-API-Reference.md`, `users-FAQ.md`)
- **Local wiki mirror:** `/Users/gio/TrueRepublic/wiki/` has a nested folder structure (`develop/`, `users/`, etc.) -- this is a local reference copy, not directly pushed
- **Release creation:**
  ```bash
  git tag -a vX.Y.Z -m "vX.Y.Z"
  git push origin vX.Y.Z
  gh release create vX.Y.Z --title "vX.Y.Z -- Title" --notes "..."
  ```
- **Commit messages:** Conventional-ish prefixes: `fix:`, `docs:`, `security:`, `deps:`, `ci:`, `assets:`
- **Co-authorship:** Commits made with Claude include `Co-Authored-By: Claude <noreply@anthropic.com>` (used in some commits)

### Documentation

- **GitHub Pages:** Static HTML in `docs/index.html`, deployed automatically from `/docs` on `main`
- **Logos:** Originals in `assets/` (for README/wiki), owl icons in `docs/assets/images/` (for GitHub Pages), PWA icons in `web-wallet/public/`
- **Wiki links in Home.md:** Use relative paths without `.md` extension (e.g., `[Title](develop-API-Reference)`)
- **README wiki links:** Use full GitHub URLs (e.g., `https://github.com/NeaBouli/TrueRepublic/wiki/develop-Architecture-Overview`)

---

## Next Immediate Step

v0.3.0 is COMPLETE. All 12 weeks finished. 577 tests (533 Go + 26 Rust + 18 Frontend), all passing.

**Completed v0.3.0 work (all 12 weeks):**
- Weeks 1-4: ZKP Anonymity Layer (Groth16, Merkle trees, nullifiers, MsgRateWithProof)
- Week 5: CosmWasm Integration (wasmd v0.53.3, custom bindings)
- Week 6: Domain-Bank Bridge (dual accounting, deposit/withdraw)
- Week 7: IBC Integration (ibc-go v8.4.0, ICS-20 transfer, Hermes relayer config)
- Week 8: Multi-Asset DEX (asset registry, trading validation, symbol resolution, CI fix)
- Week 9: Cross-Chain Liquidity (multi-hop swaps, pool analytics, slippage protection)
- Week 10: UI Components (8 React components: 3 ZKP + 5 DEX analytics, 18 frontend tests)
- Week 11: Developer Tooling (Cargo workspace with 7 crates, 4 example contracts, 26 Rust tests)
- Week 12: Final Documentation (API reference, deployment, architecture, contributing, quickstart)

**Post-release updates:**
- v0.3.0 GitHub Release published (02.03.2026)
- Official brand assets integrated (owl logo + PNYX token icon)
- Web wallet build verified (788 kB main bundle)

**Next action:** v0.4.0 (Optional Indexer Stack) or v1.0.0 preparation.

Await core dev instruction on direction.
