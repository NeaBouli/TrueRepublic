# Code Structure

Complete directory structure, file organization, and development conventions.

## Repository Layout

```
TrueRepublic/
│
├── app.go                          # Cosmos SDK application wiring (TrueRepublicApp)
├── go.mod / go.sum                 # Go module: SDK v0.50.15, CometBFT v0.38.26
├── Makefile                        # Build: build, install, test, lint, docker-*
├── Dockerfile                      # Multi-stage: Go 1.26.6 Bookworm → non-root Debian Bookworm slim
├── docker-compose.yml              # Full stack: node, client-web, nginx, prometheus, grafana
├── .env.example                    # Environment template
├── README.md                       # Project overview + feature matrix
├── INSTALLATION.md                 # Quick install guide
├── SECURITY.md                     # Security policy
│
├── x/                              # Custom Cosmos SDK modules
│   ├── truedemocracy/              # Core governance (26 transaction message types, 614 standard-suite cases)
│   │   ├── keeper.go               #   Domain CRUD, proposals, anonymous ratings
│   │   ├── anonymity.go            #   Permission register, domain key pairs (WP S4)
│   │   ├── stones.go               #   VoteToEarn, stone voting, list sorting (WP S3.1)
│   │   ├── lifecycle.go            #   Green/yellow/red zones, auto-delete (WP S3.1.2)
│   │   ├── governance.go           #   Admin election, exclusion, inactivity (WP S3.6)
│   │   ├── validator.go            #   PoD registration, staking, transfer limits
│   │   ├── slashing.go             #   Double-sign (5%), downtime (1%), jailing
│   │   ├── types.go                #   Domain, Validator, Issue, Suggestion, Rating types
│   │   ├── msgs.go                 #   26 SDK transaction message types with validation
│   │   ├── msg_server.go           #   gRPC message handlers (all 26)
│   │   ├── cli.go                  #   26 tx + 9 query CLI commands
│   │   ├── query_server.go         #   gRPC query handlers
│   │   ├── tree.go                 #   Hierarchical node tree for vote propagation
│   │   ├── module.go               #   SDK module wiring, InitGenesis, EndBlock
│   │   ├── stones_test.go          #   20 stone/VoteToEarn tests
│   │   ├── lifecycle_test.go       #   22 lifecycle/zone tests
│   │   ├── governance_test.go      #   27 governance/election/exclusion tests
│   │   ├── anonymity_test.go       #   15 anonymity/permission register tests
│   │   ├── validator_test.go       #   26 validator/PoD/transfer limit tests
│   │   └── slashing_test.go        #   6 slashing tests
│   │
│   └── dex/                        # DEX module (7 msg types, 138 recovery cases)
│       ├── keeper.go               #   CreatePool, Swap (x*y=k), Add/RemoveLiquidity
│       ├── types.go                #   Pool type, SwapFeeBps=30, BurnBps=100
│       ├── msgs.go                 #   4 SDK message types
│       ├── msg_server.go           #   gRPC message handlers
│       ├── cli.go                  #   7 tx + 9 query CLI commands
│       ├── query_server.go         #   gRPC query handlers
│       ├── module.go               #   SDK module wiring
│       └── keeper_test.go          #   24 DEX tests (swap, liquidity, fees, burn)
│
├── treasury/                       # Tokenomics
│   └── keeper/
│       ├── rewards.go              #   Whitepaper equations 1-5
│       └── rewards_test.go         #   36 equation validation tests
│
├── contracts/                      # CosmWasm smart contracts (Rust)
│   └── src/
│       ├── governance.rs           #   On-chain proposals, systemic consensing
│       └── treasury.rs             #   Deposit/withdraw treasury operations
│
├── client-web/                     # Maintained React/TypeScript/Vite client
│   ├── src/components/             # Wallet, governance, DEX, ZKP and network UI
│   ├── src/services/               # Typed chain query and transaction boundaries
│   ├── src/config/chains.ts        # Canonical chain metadata
│   ├── package.json                # Lockfile-driven lint/test/build scripts
│   ├── Dockerfile                 # Production build served by nginx
│   └── nginx.conf                  # SPA fallback and security headers
│
├── scripts/                        # Operational scripts
│   ├── init-node.sh                #   Initialize node (~/.truerepublic)
│   ├── start-node.sh               #   Start the node
│   └── backup.sh                   #   Backup with 30-day retention
│
├── monitoring/                     # Observability
│   ├── prometheus.yml              #   Prometheus scrape config
│   └── grafana/
│       ├── provisioning/           #   Auto-provisioning
│       └── dashboards/             #   Blockchain dashboard
│
├── nginx/
│   └── nginx.conf                  # Reverse proxy configuration
│
├── docs/                           # Documentation (30+ files)
│   ├── getting-started/
│   ├── user-manual/                #   7 end-user guides
│   ├── node-operators/             #   9 operator guides
│   ├── validators/                 #   Validator guide
│   ├── developers/                 #   8 developer guides
│   ├── FAQ.md
│   ├── GLOSSARY.md
│   ├── ARCHITECTURE.md
│   ├── API.md
│   ├── DEPLOYMENT.md
│   ├── VALIDATOR_GUIDE.md
│   └── WhitePaper_TR_eng.md
│
└── .github/
    └── workflows/
        ├── go-ci.yml               #   Go build + test
        ├── rust-ci.yml             #   Rust/CosmWasm build
        └── react-ci.yml            #   Maintained web-client checks
```

---

## Key Files by Purpose

### Where Do I Find...?

| Looking for... | File |
|----------------|------|
| Application entry point | `app.go` |
| Domain creation logic | `x/truedemocracy/keeper.go` → `CreateDomain()` |
| Proposal submission | `x/truedemocracy/msg_server.go` → `SubmitProposal()` |
| Systemic consensing ratings | `x/truedemocracy/keeper.go` → `RateAnonymous()` |
| Stones voting logic | `x/truedemocracy/stones.go` |
| VoteToEarn rewards | `x/truedemocracy/stones.go` → `PlaceStoneOnIssue()` |
| Suggestion lifecycle | `x/truedemocracy/lifecycle.go` |
| Admin election | `x/truedemocracy/governance.go` → `UpdateAdmin()` |
| Member exclusion | `x/truedemocracy/governance.go` → `VoteExclude()` |
| Anonymous voting keys | `x/truedemocracy/anonymity.go` |
| Validator registration | `x/truedemocracy/validator.go` → `RegisterValidator()` |
| Slashing logic | `x/truedemocracy/slashing.go` |
| Transfer limit (10%) | `x/truedemocracy/validator.go` → `WithdrawStake()` |
| DEX swap calculation | `x/dex/keeper.go` → `Swap()` |
| Pool creation | `x/dex/keeper.go` → `CreatePool()` |
| Liquidity provision | `x/dex/keeper.go` → `AddLiquidity()` / `RemoveLiquidity()` |
| Tokenomics equations | `treasury/keeper/rewards.go` |
| CLI commands (governance) | `x/truedemocracy/cli.go` |
| CLI commands (DEX) | `x/dex/cli.go` |
| Custom-module query services | `x/truedemocracy/query_server.go`, `x/dex/query_server.go` |
| Web-client services | `client-web/src/services/` |
| Wallet state | `client-web/src/stores/walletStore.ts` |
| Governance UI | `client-web/src/components/governance/` |
| DEX UI | `client-web/src/components/dex/` |
| Chain configuration | `client-web/src/config/chains.ts` |
| Docker configuration | `docker-compose.yml` |
| CI/CD pipelines | `.github/workflows/*.yml` |
| Monitoring config | `monitoring/prometheus.yml` |

---

## File Naming Conventions

### Go Files

```
keeper.go           # Core keeper logic
stones.go           # Feature: stones voting
lifecycle.go        # Feature: suggestion lifecycle
governance.go       # Feature: governance (election, exclusion)
anonymity.go        # Feature: anonymous voting
validator.go        # Feature: validator management
slashing.go         # Feature: validator slashing
types.go            # Data structures
msgs.go             # SDK message types
msg_server.go       # gRPC message handlers
cli.go              # CLI commands
query_server.go     # gRPC query handlers
module.go           # Module registration + EndBlock
*_test.go           # Tests (co-located with source)
```

Pattern: `{feature}.go` for source, `{feature}_test.go` for tests.

### React Files

```
ThreeColumnLayout.js  # Component (PascalCase)
ProposalFeed.js       # Component (PascalCase)
useWallet.js          # Hook (camelCase, "use" prefix)
api.js                # Service (camelCase)
index.js              # Entry point
index.css             # Global styles
```

Pattern: Components = PascalCase, Hooks = camelCase with "use", Services = camelCase.

### Import Order

**Go:**
```go
import (
    // Standard library
    "fmt"
    "time"

    // External (Cosmos SDK, CometBFT)
    sdk "github.com/cosmos/cosmos-sdk/types"
    abci "github.com/cometbft/cometbft/abci/types"

    // Internal (project modules)
    "truerepublic/treasury/keeper"
)
```

**JavaScript:**
```javascript
// React / external
import React, { useState, useEffect } from "react";
import { SigningStargateClient } from "@cosmjs/stargate";

// Internal components
import Header from "../components/Header";

// Hooks
import useWallet from "../hooks/useWallet";

// Services
import { fetchDomains, submitProposal } from "../services/api";
```

---

## Adding a New Feature

### Example: Add "EditProposal"

**Step 1: Backend message type**
```
→ x/truedemocracy/msgs.go        Add MsgEditProposal struct
→ x/truedemocracy/msg_server.go  Add EditProposal() handler
→ x/truedemocracy/keeper.go      Add EditProposal() keeper method
→ x/truedemocracy/cli.go         Add CLI command
→ x/truedemocracy/module.go      Register message in codec
```

**Step 2: Tests**
```
→ x/truedemocracy/*_test.go      Add test cases
→ Run: make test
```

**Step 3: Frontend**
```
→ client-web/src/services/        Add the typed transaction boundary
→ client-web/src/components/      Add UI with focused tests
→ Run: cd client-web && npm ci && npm run lint && npm test && npm run build
```

**Step 4: Documentation**
```
→ docs/developers/api-reference/  Update CLI commands
→ wiki/develop/Module-Deep-Dive   Add message documentation
```

### Development Workflow

```
1. git checkout -b feature/edit-proposal
2. Implement backend (msgs → handler → keeper → CLI)
3. Write tests (go test ./x/truedemocracy/... -v)
4. Implement frontend (typed service → component)
5. Build (make build && cd client-web && npm ci && npm run build)
6. Full test (make test)
7. git add . && git commit -m "Add edit proposal feature"
8. git push origin feature/edit-proposal
9. Open pull request
```

---

## Build System

### Makefile Targets

| Target | Command | Description |
|--------|---------|-------------|
| `build` | `go build -o build/truerepublicd .` | Build binary |
| `install` | `go install .` | Install to $GOPATH/bin |
| `test` | `./scripts/go-packages.sh go test -race -cover` | Run all tests |
| `lint` | `./scripts/go-packages.sh go vet && ./scripts/go-packages.sh staticcheck` | Static analysis |
| `clean` | `rm -rf build/` | Clean build artifacts |
| `docker-build` | `docker compose build` | Build Docker images |
| `docker-up` | `docker compose up -d` | Start full stack |
| `docker-down` | `docker compose down` | Stop full stack |

### Maintained Web Client Scripts

| Script | Command | Description |
|--------|---------|-------------|
| `dev` | `vite` | Dev server (port 3001) |
| `build` | `tsc -b && vite build` | Production build |
| `test` | audit policy tests + `vitest` | Run maintained-client tests |
| `lint` | `eslint .` | Static checks |

---

## Next Steps

- [Module Deep-Dive](Module-Deep-Dive) -- Detailed message and handler documentation
- [Development Setup](Development-Setup) -- Set up your environment
- [Contributing Guide](Contributing-Guide) -- How to contribute
