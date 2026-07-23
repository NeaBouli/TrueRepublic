# TrueRepublic repository guide

This file is the working guide for Claude Code and other engineering agents.
Historical release narratives are available in Git history; they are not a
source of current security or production-readiness claims.

## Current status

- Version label: `v0.4.0`
- Status: recovery audit active; not production-ready
- Canonical repository: `NeaBouli/TrueRepublic`
- Recovery epic: GitHub issue #4
- Continuous handoff: `BRIDGE.md` and `docs/agent-bridge/`
- Authoritative machine-readable status: `docs/status.json`
- Verified recovery total: 689 cases (655 Go, 26 Rust, eight maintained-client)
- PNYX cap: 21,000,000 PNYX = 21,000,000,000,000 `upnyx`

The recovery foundation was reviewed and merged to `main` through this ordered
PR sequence:

1. PR #9 — recovery foundation
2. PR #15 — canonical PNYX denomination and cap
3. PR #16 — bank-backed governance/stake escrow
4. PR #17 — canonical capped issuance
5. PR #18 — bank-backed DEX custody
6. PR #19 — genesis reconciliation and runtime invariants
7. PR #22 — ZKP vote statement/VK safety
8. PR #23 — persistent Cosmos/Comet node lifecycle
9. PR #24 — final documentation, public-status and CI-runtime reconciliation
10. PR #27 — retire the broken legacy init wrapper (GH-26)

All ten PRs above are merged. Treat their order as historical implementation
provenance, not as an open merge queue or production approval.

## Repository shape

TrueRepublic is a multi-project repository rather than a single-package app:

- root Go module — Cosmos SDK application and custom modules
- `x/truedemocracy` — governance, PoD validators, anonymity and ZKP primitives
- `x/dex` — asset registry, pools, swaps and liquidity custody
- `token` — canonical denomination, cap and issuance boundary
- `treasury/keeper` — deterministic reward equations
- `contracts` — Rust/CosmWasm workspace
- `client-web` — maintained React/Vite client
- `web-wallet`, `mobile-wallet` — deprecated legacy clients
- `docs` — public and operator documentation

It is not a Go workspace with multiple Go modules. The root Go module contains
the blockchain, while Rust and JavaScript have their own workspace/package
boundaries.

## Toolchain

- Go toolchain 1.26.5
- Cosmos SDK 0.50.14
- CometBFT 0.38.22
- ibc-go 8.7.0
- wasmd 0.53.3 / wasmvm 2.2.2
- gnark 0.14.0
- React 18.2, TypeScript 5.9, Vite 8.1.4, CosmJS 0.39.0
- Rust 1.75+

## Required verification

Run from the repository root:

```bash
./scripts/go-packages.sh go build
./scripts/go-packages.sh go vet
./scripts/go-packages.sh go test -count=1
./scripts/go-packages.sh go test -race -count=1 -timeout=600s
./scripts/check-consistency.sh
git diff --check
```

`scripts/go-packages.sh` derives the root-module package set from Git-managed,
non-ignored Go sources and excludes dependency trees such as `node_modules` and
`vendor`. Use `./scripts/go-packages.sh --list` to inspect the selected package
directories; do not replace the wrapper with a repository-root `./...`
wildcard in automation.

Maintained client:

```bash
cd client-web
npm ci
npm run lint
npm test -- --run
npm run build
npm audit
```

Contracts:

```bash
cd contracts
cargo test --workspace
cargo clippy --workspace --all-targets -- -D warnings
cargo audit
```

The Go CI also builds the non-root Docker image, starts a node, waits for a
block, restarts the same container and requires the height to advance.

## Consensus and ledger rules

- `upnyx` is the only base denomination; PNYX has six decimals.
- Canonical `x/bank` supply must never exceed 21,000,000,000,000 `upnyx`.
- Governance treasury/stake claims must equal the complete
  `truedemocracy` module balance.
- DEX reserves must equal the complete `dex` module balances.
- LP provider positions must sum exactly to pool total shares.
- Mint/burn operations use `token.IssuanceService`.
- `x/crisis` checks supply, escrow, reserves and LP conservation every block
  during recovery.
- Custom genesis never self-funds internal claims; bank/custom state must
  reconcile before initialization.

## Node lifecycle

`truerepublicd` uses standard Cosmos server commands and a persistent configured
home. `init` binds the generated Comet private-validator public key to the PoD
bootstrap validator. Native and Docker restart smoke tests must stay green.

IBC core/transfer wiring starts in the node, but staking and upgrade keepers are
explicit stubs. Do not claim production IBC, relayer or upgrade support without
multi-node evidence and replacement/approval of those boundaries.

## ZKP and anonymity boundaries

The GH-20 recovery circuit binds membership proofs to a chain-scoped proposal
and exact rating while retaining a rating-independent one-vote nullifier.
Transaction execution never generates trusted setup; a missing configured VK
fails closed.

The maintained and legacy web clients reject mock proof generation and
submission. Anonymous voting is not production-approved until a compatible
real prover and independent cryptographic review exist. Anonymous rewards
remain deferred because current proof/signature submissions do not bind a safe
payout recipient.

## Client policy

- `client-web` is the only maintained client.
- `web-wallet` and `mobile-wallet` remain preserved for migration/reference.
- Never recommend legacy clients for real keys or funds.
- Never silently copy legacy checkout changes into recovery branches.

The divergent checkout at `/Users/gio/Desktop/repos/TrueRepublic` is preserved
unchanged. Recovery worktrees are based on current GitHub state.

## Engineering workflow

- Read `BRIDGE.md` before starting.
- Work against a GitHub issue and update the bridge/action log continuously.
- Use conventional commits and scoped, reviewable changes.
- Keep each new rollout task tied to a GitHub issue, focused branch, and
  reviewable PR.
- Preserve unrelated local changes.
- Never commit secrets, generated key material, databases or node homes.
- Update README, `docs/status.json`, limitations and landing-page totals only
  from verified evidence.
- A green local suite is not enough: GitHub Go, docs, static review and Security
  Scan gates must pass on the final published head.

## Production boundary

Recovery testnet functionality is not a mainnet approval. GH-53 proves only
compatible persisted-state binary replacement and fail-before-open rollback.
Remaining gates include consensus-breaking state migration and partially
applied migration recovery, validator-key compromise response, network policy,
IBC/load/topology evidence, monitoring/alerting, independent consensus/
cryptographic/operations review, real client proof generation, and a formal
release process.
