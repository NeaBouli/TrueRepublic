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
  <img src="https://img.shields.io/badge/tests-2319%20recovery--verified-orange" alt="Recovery-verified tests"/>
  <img src="https://img.shields.io/badge/release-unreleased-orange" alt="Unreleased recovery candidate"/>
  <img src="https://img.shields.io/badge/recovery-active-orange" alt="Recovery active"/>
  <img src="https://img.shields.io/badge/Go-1.26.6-00ADD8?logo=go" alt="Go"/>
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
> **Recovery foundation verified; rollout still active:** the completed recovery
> evidence is preserved in [GitHub issue #4](https://github.com/NeaBouli/TrueRepublic/issues/4),
> while every remaining production-readiness gate is tracked in
> [GitHub issue #29](https://github.com/NeaBouli/TrueRepublic/issues/29).
> `client-web` is the sole maintained web client. The legacy `web-wallet`
> and `mobile-wallet` prototypes were retired and removed under GH-112 and
> GH-102; Git history preserves them for audit only.

Delivery priority is the Basic TrueRepublic rollout foundation first.
Sovereign V4-1 runtime wiring and the optional GH-232 ballot engine remain
separate, deferred tracks until rollout stabilization or a later explicit
documented decision; V4-0 stays unwired and does not change rollout readiness.

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
| **[Sovereign Alpha Architecture](docs/SOVEREIGN_ALPHA_ARCHITECTURE.md)** | Product / Engineering | Future decentralized app, messaging, governance and PNYX wallet architecture |
| **[Sovereign V4 Edge Architecture](docs/SOVEREIGN_V4_EDGE_ARCHITECTURE.md)** | Product / Engineering | Native TRChain settlement with local-first edge nodes, domain apps and civic workflows |
| **[Optional Ballot Architecture](docs/GOVERNANCE_BALLOT_ARCHITECTURE.md)** | Product / Engineering | Future domain-configurable consensing, majority, person-election, hybrid and privacy modes |
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
| **Anonymous Suggestion Ratings** | Domain key pairs for unlinkable suggestion ratings (WP S4) | [Architecture](docs/developers/architecture/module-reference.md) |
| **Proof of Domain** | Validators must be active domain members | [Validator Guide](docs/validators/README.md) |
| **DEX (stacked recovery)** | PR #18 adds custody/LP ownership/burns; PR #19 reconciles genesis and checks reserves/shares every block | [DEX Guide](docs/user-manual/dex-trading-guide.md) |
| **VoteToEarn** | Earn PNYX rewards for active participation | [Stones Guide](docs/user-manual/stones-voting-guide.md) |
| **Suggestion Lifecycle** | Green/yellow/red zones with auto-delete | [Governance](docs/user-manual/governance-tutorial.md) |
| **IBC Transfers** | GH-175/GH-178/GH-181 locally verify packet and compatible-restart recovery; GH-184 adds governed fresh-genesis application-upgrade recovery. External relayer and IBC client-upgrade qualification remain open | [IBC Setup](docs/IBC_RELAYER_SETUP.md) |

---

## Current Beta Client (v0.4.0)

The maintained React web client remains the current Beta while the future
website-independent Sovereign Alpha is designed under
[GH-215](https://github.com/NeaBouli/TrueRepublic/issues/215). The Beta has
governance and DEX functionality, but it is not approved for production keys or
funds:

```bash
cd client-web
npm ci
npm run dev
```

- Wallet: Create/import, encrypted storage, send PNYX
- Governance: Browse domains and create suggestions; GH-266 extends the test-only Go/WASM proof gate through the real keeper payout boundary while anonymous submission remains disabled pending production ceremony and review
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
├── go.mod / go.sum             Go module (SDK v0.50.15, CometBFT v0.38.26)
├── Makefile                    Build targets (build, test, lint, docker)
├── INSTALLATION.md             Quick install guide
├── x/
│   ├── truedemocracy/          Governance module (26 tx message types, 614 standard-suite cases)
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
| Anonymous Suggestion Ratings (WP S4) | 🟡 Recovery backend verified | Permission-register and test-only Groth16 paths; formal secret ballots and production ZKP qualification remain future work |
| Zero-Knowledge Proofs (Groth16) | 🟡 Test-only keeper compatibility verified | Chain/rating/recipient binding, fail-closed VK, and fresh synthetic Go/WASM proof replay through atomic keeper payout; production ceremony, submission, and external review pending |
| CosmWasm Smart Contracts | ✅ | `x/truedemocracy/wasm_bindings.go` |
| Domain-Bank Bridge | ✅ | `x/truedemocracy/treasury_bridge.go` |
| IBC Transfer (ICS-20) | 🟡 Two-chain lifecycle, channel replacement, compatible restart, and governed fresh-genesis app-upgrade recovery verified locally through GH-184; external relayer and IBC client-upgrade evidence pending | `app.go` (ibc-go v8.7.0) |
| Stones Voting (WP S3.1) | ✅ | `x/truedemocracy/stones.go` |
| VoteToEarn Rewards | ✅ | `x/truedemocracy/stones.go` |
| Suggestion Lifecycle (WP S3.1.2) | ✅ | `x/truedemocracy/lifecycle.go` |
| Green/Yellow/Red Zones | ✅ | `x/truedemocracy/lifecycle.go` |
| Auto-Delete & Fast Delete (2/3) | ✅ | `x/truedemocracy/lifecycle.go` |
| Admin Election (WP S3.6) | ✅ | `x/truedemocracy/governance.go` |
| Member Exclusion (2/3 vote) | ✅ | `x/truedemocracy/governance.go` |
| PoD Transfer Limit (10%, WP S7) | ✅ | `x/truedemocracy/validator.go` |
| CLI Commands (26 tx + 9 query) | ✅ | `x/truedemocracy/cli.go` |
| DEX CLI (7 tx + 9 query) | ✅ | `x/dex/cli.go` |
| CosmWasm Contracts (7 crates) | ✅ | `contracts/` (workspace) |
| Maintained Web Client | 🟡 Recovery verified | `client-web/` |
| Legacy Web Wallet | ⚫ Retired and removed under GH-112 | Git history only |
| Legacy Mobile Wallet | ⚫ Retired and removed under GH-102 | Git history only |
| CI/CD Workflows | ✅ | `.github/workflows/` |

---

## Build & Test

```bash
# Blockchain (the committed module graph remains unchanged)
./scripts/go-packages.sh go build
CGO_ENABLED=1 ./scripts/go-packages.sh go test -race -cover -count=1 -timeout=600s    # 1,974 Go cases
make cross-run-evidence-contract-test                                  # metadata-only GH-273 contract
make ibc-two-chain                                                     # separate GH-175/GH-178/GH-181 proof gate

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
| CometBFT | v0.38.26 | Recovery verified |
| CosmWasm | v0.53.4 | Recovery verified |
| ibc-go | v8.7.0 | Two-chain packet lifecycle verified locally; external relayer/upgrade unqualified |
| gnark (ZKP) | v0.14.0 | On-chain recovery verified; client disabled |
| Go | 1.26.6 | Recovery verified |
| Rust | 1.75+ | Contracts |
| React | 18.2 | Maintained v0.4 client |
| Native mobile client | — | Retired under GH-102; replacement pending |
| Local encrypted test wallet + CosmJS | 0.39 | Maintained v0.4 client |

**Known Limitations:** TrueRepublic uses PoD and does not mount `x/staking` or `x/distribution`; GH-187 makes every required IBC/CosmWasm compatibility adapter reject those unsupported surfaces fail-closed. GH-184 wires real `x/upgrade` for the exact governed fresh-genesis v0.4.1 path, but pre-GH-184 store introduction and IBC client upgrades remain unsupported. GH-206 verifies only synthetic test compatibility, and GH-266 extends that evidence through the real keeper payout boundary without creating a production prover or submission path. GH-209 implements front-running-safe recipient-bound anonymous rewards while documenting direct-payout linkability; production ZKP qualification, ceremony, submission integration, and independent cryptographic/privacy review remain pending. GH-93 provides a strict synthetic incident-command and rehearsal contract, but a private live operator rehearsal remains pending. GH-85 provides dashboard/application runtime evidence, alert rules, recovery/testnet objectives, and role ownership. GH-89 adds a strict synthetic topology qualification contract. GH-97 adds bounded four-validator load, resource, retention, restart, and ledger evidence; it does not establish production sizing or multi-day soak behavior. GH-101 adds a digest-bound offline deployment-evidence envelope and verifier; it does not prove or perform a live deployment. GH-212 pins release tools, supported platforms and maintained container bases and adds strict unsigned offline release evidence. GH-258 adds native same-job double-build OCI identity evidence for daemon and `client-web`. GH-261 binds the exact commit and an explicitly simulated future tag to both native daemon metadata records and both-platform OCI digest reports in protected CI, retaining metadata/JSON only. Floating runner/BuildKit versions and live Debian package indexes still prevent a cross-time hermetic reproducibility claim. These controls create no real tag, publish or sign no artifact, authenticate no provenance, deploy nothing, and approve no rollout. Real seed/sentry/validator/RPC deployment, firewall/TLS/DNS evidence, external paging drills, private-environment capacity evidence, and independent live operations review remain open. See [LIMITATIONS.md](docs/LIMITATIONS.md).

---

## Current Status

**Recovery line: v0.4.0 — current daemon candidate untagged; binary identity is the source commit; not production-ready**

The checklist below records implemented surface area, not a production security
approval. Current evidence, risks, and commands are maintained in
[`BRIDGE.md`](BRIDGE.md) and the active
[rollout tracker #29](https://github.com/NeaBouli/TrueRepublic/issues/29).

- 🟡 2,319 tests recovery-verified locally (1,974 Go, including GH-273's strict cross-run evidence contract, GH-261's candidate-evidence contract, GH-258's OCI build/evidence contract, GH-244's rollout-genesis qualification contract, GH-225's release-compatibility contract, GH-222's verified install-lifecycle and repository contracts, GH-209's recipient-binding and atomic-payout adversarial coverage, + 26 Rust + 319 maintained-client, including its v2 encoding and canonical-recipient validation), plus the separately gated GH-266 fresh Go/WASM-to-keeper payout/replay proof, GH-206 native-verifier compatibility proof, GH-175/GH-178/GH-181 IBC proof and GH-184 governed-upgrade recovery proof, GH-172 shared-state contention/exact-replay/restart proof, GH-145 bounded live fuzz campaigns, GH-193 maintained-client wallet/signing-safety proof, GH-190 maintained-client IBC transfer/recovery proof, GH-131 real submitted-history pagination proof, GH-121 real browser-query boundary, GH-115 local client-chain delivery proof, GH-56 rotation, GH-59 slashing, GH-60 inactive-validator genesis, GH-61 legacy-authority migration, GH-93 incident rehearsal, and GH-97 sustained-load process harnesses; production rollout evidence remains required
- 🟡 Rollout accounting remains 35/59 overall and 35/51 phase work. Phase 6
  is 6/7 and Phase 7 is 3/10; release freeze and accountable go/no-go are two
  mandatory subchecks of one counted Phase-7 tracker item. Production remains
  false.
- 🟡 GH-212 verifies exact release tool/platform/container-base pins, repeated
  normalized SBOM parity and a strict unsigned two-target evidence bundle.
  GH-258 additionally requires two native no-cache OCI exports per daemon and
  maintained-client target to agree on index, manifest, config and ordered layer
  digests. GH-261 aggregates their digest reports with both deterministic daemon
  metadata records under one exact commit and simulated-tag manifest. GH-273
  adds a strict metadata-only comparison capability for two distinct protected
  executions of the unchanged exact commit. Hosted runs
  [33465480131](https://github.com/NeaBouli/TrueRepublic/actions/runs/33465480131)
  and [33466167289](https://github.com/NeaBouli/TrueRepublic/actions/runs/33466167289)
  verified both daemon identities and all four OCI identities on exact commit
  `3b0d1639bb40c7df6733dd13a86252e1c8c9efd3`. This proves only that recorded
  pair; long-term hermetic rebuilds, a real tag, signing, attestation,
  publication and staged rollout remain open.
- ✅ Maintained-client production build remains within budget at 355.17 kB
  total JavaScript gzip, with a 71.16 kB entry and 4.94 kB largest lazy route
- ✅ Core blockchain compiles and runs
- 🟡 Tokenomics, exact custom genesis, and every-block ledger invariants are recovery-verified and merged through PR #19
- 🟡 Governance escrow/auth recovery is verified and merged; independent release review remains open
- 🟡 Groth16 voting backend, fresh synthetic Go/WASM proof replay through the real keeper payout boundary, one-time nullifier enforcement, and recipient-bound atomic rewards tested; production ceremony, audited submission, and independent cryptographic/privacy review remain open
- ✅ CosmWasm smart contract integration (wasmd v0.53.4)
- 🟡 Domain-Bank escrow recovery implemented and merged via PR #16
- 🟡 IBC transfer on ibc-go v8.7.0 has local proof-driven two-chain transfer,
  acknowledgement, timeout refund, replay-safety, pending-ACK restart,
  timeout-on-close, replacement-channel, and compatible in-place binary-restart
  evidence on GH-175/GH-178/GH-181; external relayers and governed
  consensus-migration behavior remain open
- 🟡 GH-190 adds canonical native ICS-20 signing to the maintained client,
  strict fail-closed channel/amount/balance/timeout checks, wallet-scoped
  recovery records, and manual source-chain ACK/timeout reconciliation. A
  committed broadcast is never displayed as delivery or automatically retried.
- 🟡 GH-193 hardens the maintained browser wallet with bounded BIP-39
  errors, versioned AES-GCM/PBKDF2-600k storage and legacy re-encryption, exact
  derived-address/RPC-chain binding, and in-flight signer invalidation after
  lock, switch, delete, or reload. Same-origin XSS, hardware custody, real keys,
  real funds, and production wallet operation remain outside this evidence.
- 🟡 Multi-Asset DEX bank custody, provider LP ownership, authority checks, and
  canonical burns are recovery-verified and merged via PR #18
- 🟡 GH-12 genesis/runtime conservation is recovery-verified and merged via PR #19
- 🟡 PR #23 provides generated CometBFT-key, bank-backed PoD genesis and
  proves native restart/export; GH-26 makes `scripts/init-node.sh` delegate only
  to that supported path. GH-32/GH-41/GH-43/GH-45/GH-53 add four-validator
  failure/restart/catch-up, partition, state-sync, and sanitized backup/restore
  gates plus compatible binary upgrade/fail-before-open rollback evidence;
  consensus-breaking migrations and broader multi-node operations remain open
- 🟡 ZKP UI remains a clearly disabled preview; GH-266's keeper replay uses GH-206's isolated synthetic test-only prover and cannot submit
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
  - 🟡 ZKP Anonymous Voting (GH-266 keeper replay, GH-206 test-only proving compatibility, and GH-209 recipient-bound rewards verified; production qualification, ceremony, submission, and independent review pending)
  - ✅ Domain Membership & Onboarding
  - ✅ Admin Dashboard (member management, stats)
  - ✅ Network Explorer (validators, blocks, IBC)
- 📐 **Sovereign Alpha track (GH-215; implementation not started):** an
  installable, Telegram-like domain/discussion/governance client with an
  integrated non-custodial PNYX wallet and no mandatory hosted-website
  dependency. `client-web` remains the Beta during construction; whether an
  optional web interface survives is deferred until the Alpha exit gates pass.
- 📐 **Sovereign V4 edge track (GH-236/GH-241):** evolves the Alpha
  with TRChain as the sole settlement chain, user-operated pruned nodes, mobile
  verification, local-first civic workflows and sandboxed domain apps. It adds
  no Minima dependency, second chain, token, bridge commitment or rollout credit.
  V4-0 now provides an unwired canonical Go protocol package with strict signed
  certificates, envelopes, manifests and adversarial vectors; no edge runtime
  or client is active.
- 🎯 **Production release (no committed date):** external reviews, signed
  reproducible clients and chain artifacts, staged networks, and explicit
  go/no-go approval are still required.

> Historical test count: 577. The authoritative recovery-verified total is 2,319
> (1,974 Go + 26 Rust + 319 maintained-client), reproduced from fresh
> package-scoped output using the established passing-case method.

---

## Developer Documentation

| Guide | Description |
|-------|-------------|
| [API Reference](docs/API_REFERENCE.md) | Complete API overview |
| [Deployment Guide](docs/DEPLOYMENT.md) | Production setup |
| [Architecture](docs/ARCHITECTURE.md) | System design |
| [Sovereign Alpha Architecture](docs/SOVEREIGN_ALPHA_ARCHITECTURE.md) | Future decentralized client, messaging and wallet design |
| [Sovereign V4 Edge Architecture](docs/SOVEREIGN_V4_EDGE_ARCHITECTURE.md) | Native edge-node, civic-workflow and domain-app target design |
| [Optional Ballot Architecture](docs/GOVERNANCE_BALLOT_ARCHITECTURE.md) | Future domain-configurable formal ballot and privacy design |
| [Quick Start](docs/QUICKSTART.md) | 5-minute setup |
| [Contributing](CONTRIBUTING.md) | Development guide |

## Contributing

Contributions are welcome through issues and pull requests. Intentionally
submitted contributions follow the repository's Apache-2.0 inbound-equals-
outbound model; contributors retain copyright in their work. See
[CONTRIBUTING.md](CONTRIBUTING.md) for review, testing, provenance, and
scope requirements.

## License

TrueRepublic is a community-governed open-source project without a central
corporate owner. Maintained source code and maintained documentation are
Apache-2.0; see the root `LICENSE` and `NOTICE`. Copyright remains with the
individual contributors. The collective attribution is “TrueRepublic contributors.”

Brand assets, artwork, historical PDFs, archived historical evidence, and
third-party materials remain outside that project grant unless a file-specific
notice documents otherwise. The exact governance record is
[GH-219](https://github.com/NeaBouli/TrueRepublic/issues/219#issuecomment-5423337355);
the reviewed decision package and machine state are
[documented here](docs/legal/GH-219-LICENSE-DECISION.md).

## Community

- Telegram: [t.me/truerepublic](https://t.me/truerepublic)
- Issues: [github.com/NeaBouli/TrueRepublic/issues](https://github.com/NeaBouli/TrueRepublic/issues)
- Email: p.cypher@protonmail.com

## Contributors

- NeaBouli

## Donations

Team (BTC multi-sig): `bc1qyamf3twgcqckuqrvmwgwnhzupgshxs37eejdgl0ntcqve98qnvhqe6cjl9`

Developer support (BTC): `bc1p2kh7reqf7l5zmssk8kdx2fx0q6suzfmsgdrphl9wjsxarns56jkq4tqvjm`
