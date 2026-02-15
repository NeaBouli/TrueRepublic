# TrueRepublic / PNYX

[![Go CI](https://github.com/NeaBouli/TrueRepublic/actions/workflows/go-ci.yml/badge.svg)](https://github.com/NeaBouli/TrueRepublic/actions/workflows/go-ci.yml)
[![Rust CI](https://github.com/NeaBouli/TrueRepublic/actions/workflows/rust-ci.yml/badge.svg)](https://github.com/NeaBouli/TrueRepublic/actions/workflows/rust-ci.yml)
[![Web CI](https://github.com/NeaBouli/TrueRepublic/actions/workflows/react-ci.yml/badge.svg)](https://github.com/NeaBouli/TrueRepublic/actions/workflows/react-ci.yml)
[![Mobile CI](https://github.com/NeaBouli/TrueRepublic/actions/workflows/react-native-ci.yml/badge.svg)](https://github.com/NeaBouli/TrueRepublic/actions/workflows/react-native-ci.yml)

---

## Vision

TrueRepublic is a platform for **direct democracy** and **digital self-determination**.
The token **PNYX** enables governance, treasury mechanisms and a decentralized DEX.

---

## Repository Structure & Status

```text
TrueRepublic/
├── app.go                      ✅  Cosmos SDK application entry point
├── go.mod / go.sum             ✅  Go module (SDK v0.50.13, CometBFT v0.38.17)
├── x/
│   ├── truedemocracy/          ✅  Governance module (domains, voting, PoD consensus)
│   │   ├── keeper.go               Domain CRUD, proposals, anonymous ratings (eq.2, eq.3)
│   │   ├── anonymity.go            Permission register, anonymous voting (WP §4)
│   │   ├── stones.go               Stone voting, VoteToEarn, list sorting (WP §3.1)
│   │   ├── lifecycle.go            Suggestion zones, auto-delete, fast delete (WP §3.1.2)
│   │   ├── governance.go           Member stones, admin election, exclusion (WP §3.6)
│   │   ├── validator.go            Proof of Domain validator lifecycle
│   │   ├── slashing.go             Double-sign & downtime penalties
│   │   ├── module.go               SDK module wiring, InitGenesis, EndBlock
│   │   ├── types.go                Domain, Validator, Issue, Rating, VoteCommitment
│   │   ├── tree.go                 Hierarchical node tree for vote propagation
│   │   ├── stones_test.go           20 stones / VoteToEarn tests
│   │   ├── lifecycle_test.go        22 lifecycle / zone tests
│   │   ├── governance_test.go      27 governance / election / exclusion tests
│   │   ├── anonymity_test.go       15 anonymity / permission register tests
│   │   ├── validator_test.go       18 validator / PoD tests
│   │   └── slashing_test.go        6 slashing tests
│   └── dex/                    ✅  DEX module (AMM constant-product swap)
│       ├── keeper.go               CreatePool, Swap (x*y=k), AddLiquidity, RemoveLiquidity
│       ├── module.go               SDK module wiring, InitGenesis
│       ├── types.go                Pool type, swap fee constant (0.3%)
│       └── keeper_test.go          20 DEX unit tests
├── treasury/
│   └── keeper/
│       ├── rewards.go          ✅  Whitepaper tokenomics equations 1-5
│       └── rewards_test.go         31 tokenomics tests
├── ui/                         🔵  C++ desktop UI (prototype)
├── contracts/                  🔵  CosmWasm smart contracts (skeletons)
├── docs/                       ✅  Whitepaper (PDF + Markdown), install guide
├── web-wallet/                 🔵  React web wallet (skeleton)
├── mobile-wallet/              🔵  React Native wallet (skeleton)
├── SECURITY.md                 ✅  Security policy
└── .github/                    🔵  CI/CD workflows
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
| Node Staking Rewards | ✅ | `treasury/keeper/rewards.go` (eq.5) |
| Domain Interest | ✅ | `treasury/keeper/rewards.go` (eq.4) |
| Release Decay | ✅ | `treasury/keeper/rewards.go` |
| Treasury Drainage | ✅ | `treasury/keeper/rewards.go` (eq.2) |
| Anonymous Voting (WP §4) | ✅ | `x/truedemocracy/anonymity.go` |
| Permission Register & Purge | ✅ | `x/truedemocracy/anonymity.go` |
| Domain Key Pairs (unlinkable) | ✅ | `x/truedemocracy/keeper.go` |
| Stones Voting (WP §3.1) | ✅ | `x/truedemocracy/stones.go` |
| VoteToEarn Rewards | ✅ | `x/truedemocracy/stones.go` |
| List Sorting (stones + date) | ✅ | `x/truedemocracy/stones.go` |
| Suggestion Lifecycle (WP §3.1.2) | ✅ | `x/truedemocracy/lifecycle.go` |
| Green/Yellow/Red Zones | ✅ | `x/truedemocracy/lifecycle.go` |
| Auto-Delete (red expiry) | ✅ | `x/truedemocracy/lifecycle.go` |
| Fast Delete (2/3 majority) | ✅ | `x/truedemocracy/lifecycle.go` |
| Member Ranking (stones) | ✅ | `x/truedemocracy/governance.go` |
| Admin Election (WP §3.6) | ✅ | `x/truedemocracy/governance.go` |
| Member Exclusion (2/3 vote) | ✅ | `x/truedemocracy/governance.go` |
| Inactivity Cleanup (360 days) | ✅ | `x/truedemocracy/governance.go` |
| External Links (issues/suggestions) | ✅ | `x/truedemocracy/types.go` |

---

## Build & Test

```bash
go mod tidy
go build ./...
go test ./... -race -cover
```

---

## Current Status

- ✅ Core blockchain compiles and runs (Cosmos SDK v0.50.13)
- ✅ 171 unit tests passing across 3 modules
- ✅ Whitepaper tokenomics fully implemented
- ✅ Proof of Domain consensus with validator management
- ✅ DEX with AMM swap, liquidity pools, 0.3% fees
- ✅ Anonymous voting with domain key pairs and permission register (WP §4)
- ✅ Stones voting with VoteToEarn rewards and list sorting (WP §3.1)
- ✅ Suggestion lifecycle with green/yellow/red zones and auto-delete (WP §3.1.2)
- ✅ Fast delete by 2/3 majority vote
- ✅ Member ranking, admin election, and member exclusion by 2/3 vote (WP §3.6)
- ✅ Inactivity cleanup (360-day timeout) and external links
- 🔵 CLI transaction commands and gRPC services not yet wired
- 🔵 Wallets and contracts are skeleton placeholders
- 🔵 CI/CD workflows prepared but not all enabled
