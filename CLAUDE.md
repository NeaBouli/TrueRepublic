# Project Handover -- TrueRepublic / PNYX

## Current Status

**Version:** v0.2.1 (Documentation Sync) -- tagged 24.02.2026
**Phase:** Pre-production, testnet-ready. ~95% whitepaper feature coverage (excluding ZKP/IBC). Zero P0 issues remain.

### Repository State

| Repo | Branch | HEAD | Path |
|------|--------|------|------|
| **Main** | `main` | `bc7d175` (docs: update CLAUDE.md HEAD refs) | `/Users/gio/TrueRepublic/` |
| **Wiki** | `master` | `eed3558` (docs: update wiki for v0.2.0) | `/Users/gio/TrueRepublic/wiki-github/` |

- Working tree: **clean**, up-to-date with `origin/main`
- `wiki-github/` is untracked in the main repo (it is a separate git clone of the GitHub Wiki repo -- this is expected and correct)
- Remote: `https://github.com/NeaBouli/TrueRepublic`

### Releases (10 total)

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
| v0.2.1 | Documentation Sync (Latest) | 24.02.2026 |

### Live Deployments

- **GitHub Pages:** https://neabouli.github.io/TrueRepublic/ (deployed from `/docs` on `main`)
- **Wiki:** https://github.com/NeaBouli/TrueRepublic/wiki (30 pages, synced from `wiki-github/`)
- **Releases:** https://github.com/NeaBouli/TrueRepublic/releases

### Key Metrics

- 225 unit tests across 3 modules (~3,800 lines of test code)
- 19 transaction types (15 governance + 4 DEX)
- 6 query endpoints (4 governance + 2 DEX)
- 5 tokenomics equations fully implemented + domain interest in EndBlock
- ~10,800 lines of source code (Go + JS + Rust)
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

---

## Architecture Decisions

### Blockchain Stack

| Layer | Choice | Version | Rationale |
|-------|--------|---------|-----------|
| Consensus | CometBFT | v0.38.21 | BFT consensus with fast finality; Cosmos ecosystem standard |
| Application Framework | Cosmos SDK | v0.50.13 | Module-based architecture, IBC support, mature tooling |
| Language | Go | 1.23.5 (CI: 1.24.13) | Performance, concurrency, Cosmos SDK native language |
| Smart Contracts | CosmWasm (Rust) | cosmwasm-std 3 | Deterministic WASM execution, memory-safe |
| Web Frontend | React 18 + Tailwind 3.4 | 18.2 / 3.4.19 | Component-based UI, utility-first CSS |
| Mobile | React Native + Expo | 0.74 / 51.0 | Cross-platform mobile from shared JS codebase |
| Wallet Integration | Keplr + CosmJS | 0.32-0.38 | Standard Cosmos wallet; CosmJS for chain interaction |

### Module Architecture

Two custom Cosmos SDK modules plus a treasury package:

1. **x/truedemocracy** (~7,300 lines, 25 files) -- Governance: domains, proposals, systemic consensing scoring (-5 to +5), stones voting, suggestion lifecycle (green/yellow/red zones), validator PoD, slashing, anonymous voting (domain key signatures), admin elections, member exclusion, person election voting modes (simple/absolute majority, abstention), domain interest payout (eq.4)
2. **x/dex** (1,637 lines, 9 files) -- AMM DEX: constant-product (x*y=k), PNYX/ATOM pool, 0.3% swap fee, 1% PNYX burn
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

- **Current:** PNYX/ATOM trading pair only
- **Planned (v0.3):** BTC, ETH, LUSD pools via IBC
- The wiki `users-System-Overview.md` previously said "PNYX/USDC" -- corrected to "PNYX/ATOM; BTC, ETH, LUSD planned via IBC" in v0.1.4

### Naming: CometBFT (not Tendermint)

The consensus engine was rebranded from Tendermint to CometBFT. All documentation references have been updated. The only remaining "Tendermint" references are in Go dependency paths (`github.com/cometbft/cometbft`) and historical commit messages, which is correct.

---

## Project Structure

```
TrueRepublic/
├── app.go                          Cosmos SDK app entry point (245 lines)
│                                   Wires truedemocracy + dex modules, ABCI handlers
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
├── x/truedemocracy/                GOVERNANCE MODULE (25 files, ~7,300 lines)
│   ├── keeper.go                   Domain CRUD, proposal submission, fee validation,
│   │                               RateProposalWithSignature (anonymous rating)
│   ├── msg_server.go               Message handlers (15 tx types)
│   ├── query_server.go             gRPC query handlers (4 query types)
│   ├── cli.go                      Cobra CLI commands (16 tx + 4 query)
│   ├── module.go                   Module registration, codecs, EndBlock hooks
│   ├── msgs.go                     Message type definitions (15 types)
│   ├── types.go                    Domain, DomainOptions, VotingMode, VoteChoice structs
│   ├── scoring.go                  Systemic Consensing: ComputeSuggestionScore,
│   │                               RankSuggestionsByScore, FindConsensusWinner (WP §3.2)
│   ├── governance.go               Admin election, member exclusion
│   ├── election.go                 Person election voting: CastElectionVote, TallyElection (WP §3.7)
│   ├── validator.go                PoD consensus, registration, transfer limit (10%),
│   │                               DistributeDomainInterest (eq.4)
│   ├── slashing.go                 5% double-sign, 1% downtime penalties
│   ├── stones.go                   Stones voting + VoteToEarn rewards
│   ├── lifecycle.go                Suggestion lifecycle (green/yellow/red, auto-delete)
│   ├── anonymity.go                Domain key pairs for anonymous voting (WP S4)
│   ├── tree.go                     Tree data structures
│   ├── querier.go                  Legacy query interface
│   ├── *_test.go (9 files)         170 tests: governance, validator, stones, lifecycle,
│   │                               anonymity, slashing, elections, scoring, domain interest
│
├── x/dex/                          DEX MODULE (9 files, 1,637 lines)
│   ├── keeper.go                   AMM pool operations (x*y=k)
│   ├── msg_server.go               CreatePool, Swap, AddLiquidity, RemoveLiquidity
│   ├── query_server.go             Pool queries
│   ├── cli.go                      DEX CLI (4 tx + 2 query)
│   ├── module.go                   DEX module setup
│   ├── msgs.go / types.go          Message + type definitions
│   ├── querier.go                  Legacy queries
│   ├── keeper_test.go              24 tests
│
├── treasury/keeper/                TOKENOMICS (2 files, 371 lines)
│   ├── rewards.go                  Equations 1-5: domain cost, rewards, put price,
│   │                               domain interest, node staking (with decay)
│   ├── rewards_test.go             31 tests
│
├── contracts/                      COSMWASM SMART CONTRACTS (Rust)
│   ├── src/lib.rs                  Module exports
│   ├── src/governance.rs           Governance contract (systemic consensing)
│   ├── src/treasury.rs             Treasury contract (balance management)
│   ├── Cargo.toml                  cosmwasm-std 3, serde, schemars
│
├── web-wallet/                     WEB FRONTEND (13 files, 1,053 lines)
│   ├── src/App.js                  Root component, domain/proposal state
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
│   ├── WhitePaper_TR_eng.md        English whitepaper (~430 lines, corrected 21M, §3.7 added)
│   ├── WhitePaper_TR.md            German whitepaper
│   ├── API.md                      REST/RPC/CLI reference
│   ├── FAQ.md                      30+ Q&A
│   ├── GLOSSARY.md                 60+ term definitions
│   ├── INSTALL.md / DEPLOYMENT.md  Setup and deployment guides
│   ├── ARCHITECTURE.md             System architecture overview
│   ├── VALIDATOR_GUIDE.md          Validator operations
│   ├── getting-started/            Quick start guide
│   ├── user-manual/                7 end-user guides (governance, DEX, voting, etc.)
│   ├── node-operators/             9 guides (install, config, operations)
│   ├── developers/                 8 guides (architecture, API, integration, contracts)
│   ├── validators/                 Validator guide with PoD
│
├── assets/                         PROJECT LOGOS
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

### Wiki Stubs

All 30 wiki pages are now complete. The last stub (`status-Known-Bugs.md`) was filled in v0.1.8 with 0 P0, 1 P2 (election CLI wiring), 4 P3 bugs, and 7 resolved entries.

### Roadmap Dates Are Stale

The roadmap dates in `README.md` (lines 221-226) and `wiki-github/status-Roadmap.md` reference Q2/Q3/Q4 **2025**, which is now past. If the core dev wants to update these to 2026 dates, the following files need editing:
- `README.md` (lines 223-225)
- `wiki-github/status-Roadmap.md`
- `wiki/status/Roadmap.md` (local wiki mirror)
- `wiki-github/status-Current-Status.md` (references dates)
- `docs/index.html` (if roadmap is mentioned)

### GitHub Pages Stats Section

`docs/index.html` shows 197 tests and v0.1.8 in the stats section. These should be updated to reflect 225 tests and v0.2.0.

### Roadmap Features Not Yet Implemented

Per the roadmap, the following are planned but not built:

**v0.3 scope:**
- Zero-Knowledge Proofs (ZKP) replacing domain key pairs for anonymous voting
- Full UI integration (enhanced web/mobile UX)
- WebSocket subscriptions for real-time updates
- Multi-asset DEX pools via IBC (BTC, ETH, LUSD)
- IBC channel setup (Cosmos Hub, Osmosis, Juno)
- Network scalability tests (175+ validator nodes)
- CosmWasm contract integration

**v1.0 scope:**
- Professional security audits (smart contracts + Go modules)
- Proxy party functionality
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

1. **Go version discrepancy:** `go.mod` says `go 1.23.5` but CI runs `1.24.13`. This is intentional (minimum vs CI target) but one wiki stub previously mentioned "1.24+" which was corrected.
2. **No explicit linting configs:** No `.eslintrc`, `.prettierrc`, or `golangci-lint.yml` exist. Go uses `go vet` + optional staticcheck. Rust enforces `cargo fmt` + `clippy` in CI. JavaScript has no enforced linting.
3. **Test count in docs/index.html:** The GitHub Pages site shows 197 tests and v0.1.8, which should be updated to 225 tests and v0.2.0.
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
- **Logos:** Available in both `assets/` (for README/wiki via raw.githubusercontent.com) and `docs/assets/` (for GitHub Pages)
- **Wiki links in Home.md:** Use relative paths without `.md` extension (e.g., `[Title](develop-API-Reference)`)
- **README wiki links:** Use full GitHub URLs (e.g., `https://github.com/NeaBouli/TrueRepublic/wiki/develop-Architecture-Overview`)

---

## Next Immediate Step

There are no blocked or in-progress tasks. The project is at v0.2.0 with ~95% whitepaper feature coverage. Possible next actions, in priority order:

1. **Update docs/index.html** -- Reflect 225 tests, v0.2.0
2. **Update wiki** -- status-Current-Status.md, status-Known-Bugs.md (P2 election CLI wiring is now resolved)
3. **Begin v0.3 development** -- ZKP for anonymous voting, CosmWasm integration, IBC/multi-asset DEX, UI integration

Await core dev instruction on which direction to proceed.
