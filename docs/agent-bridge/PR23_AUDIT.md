# PR #23 Audit — GH-21 Persistent PoD Node Lifecycle
> Scope: `app.go`, `server_lifecycle.go`, genesis integration, Docker/Compose, entrypoint, and Go CI  ·  Date: 2026-07-12  ·  Result: 0 FAIL / 2 WARN / 8 PASS

## Summary

GH-21 replaces the in-memory `select {}` node placeholder with the standard
Cosmos SDK/CometBFT server lifecycle and a persistent application database.
Initialization binds the node's generated CometBFT Ed25519 public key to one
exactly bank-backed PoD validator and refuses to overwrite an existing
consensus set. Native start, signal shutdown, same-home restart, height
advancement, invariant execution, and export pass locally. The stacked draft
is not production-approved: refreshed GitHub Docker evidence and independent
multi-node operations review remain required.

## Findings by domain

### Persistent application lifecycle — PASS

- **[PASS] Standard Cosmos server owns startup and shutdown** — `server_lifecycle.go:160`
  - What: `server.AddCommands` supplies the standard start/export/rollback/
    Comet command set and `newApp` receives the configured persistent database,
    home, and BaseApp options.
  - Path: The daemon no longer creates `NewMemDB()` or blocks forever; SIGINT
    reaches the SDK shutdown path and returns exit status zero.
  - Fix: Preserve the standard server creator/exporter boundary.

- **[PASS] Same-home restart preserves and advances committed state** — `server_lifecycle_test.go:60`
  - What: A real compiled daemon initializes a temporary home, commits blocks,
    stops, restarts from the same database, advances height, and exports state.
  - Path: Losing the DB/home, failing IAVL store reopening, or bypassing graceful
    shutdown fails the integration test.
  - Fix: Keep this subprocess regression in blocking Go CI.

### Validator identity and ledger safety — PASS

- **[PASS] Generated consensus identity is the PoD genesis identity** — `server_lifecycle.go:220`
  - What: `init` loads only the generated private-validator public key, creates
    matching CometBFT and custom-module validators, and validates exact module
    bank backing before writing genesis.
  - Path: Invalid key length, mismatched custom state, unbacked stake, or a cap
    overflow aborts initialization.
  - Fix: Never introduce a default private key or separate validator identity.

- **[PASS] Existing consensus validators are never silently replaced** — `server_lifecycle.go:260`
  - What: Binding refuses any pre-existing consensus validator set before
    modifying app state.
  - Path: Re-running the binder against an initialized/multi-validator genesis
    returns an error and byte-for-byte preserves the file.
  - Fix: Require an explicit audited multi-validator genesis workflow later.

- **[PASS] Genesis replacement is atomic and private** — `server_lifecycle.go:304`
  - What: The completed genesis is synced to a temporary file, closed, and
    atomically renamed with mode `0600`.
  - Path: Validation or write failure cannot leave a partially rewritten
    genesis; validator and account configuration are not world-readable.
  - Fix: Preserve atomic replacement semantics.

### CLI and configuration — PASS

- **[PASS] Operator configuration and binary version are honored** — `server_lifecycle.go:101`, `server_lifecycle.go:179`
  - What: Home, pruning, DB, RPC/P2P, API/gRPC, gas prices, and shutdown flags
    flow through SDK configuration. Both `version` and `--version` expose the
    linker-injected application version.
  - Path: The audit reproduced an unknown `--version` flag and blank `version`
    output, then added metadata wiring and a regression test.
  - Fix: Keep application and Cosmos SDK version metadata synchronized.

### Container and CI — PASS

- **[PASS] Container defaults are persistent and non-root** — `Dockerfile:35`, `scripts/docker-entrypoint.sh:1`
  - What: Debian/glibc loads wasmvm, a system user owns the node home, the
    entrypoint initializes only a missing genesis, and RPC/API/gRPC bind to
    container interfaces with a health check.
  - Path: Restarts reuse the same Docker volume rather than creating a new
    chain or ephemeral database.
  - Fix: Preserve the non-root home and glibc/wasmvm architecture mapping.

- **[PASS] GitHub CI exercises a real container restart** — `.github/workflows/go-ci.yml:48`
  - What: CI validates Compose, builds the image, waits for a committed block,
    restarts the same container, and requires height advancement.
  - Path: Linkage, entrypoint, binding, health/RPC exposure, persistence, or
    restart failures make the job fail.
  - Fix: Keep this job required for node-runtime changes.

### Operations verification — WARN

- **[MEDIUM] Local Docker execution is unavailable** — `Dockerfile`, `.github/workflows/go-ci.yml`
  - What: This host has no Docker CLI/daemon, so only shell syntax and native
    lifecycle execution were locally reproduced.
  - Path: A container-only defect is not excluded until the refreshed GitHub
    Docker restart job passes on the published head.
  - Fix: Treat the GitHub Docker job as mandatory before approving PR #23.

- **[HIGH] Evidence remains single-node and stub-bounded** — `ibc_stubs.go`, `wasm_stubs.go`, `docs/LIMITATIONS.md`
  - What: IBC staking/upgrade and standard CosmWasm staking/distribution remain
    explicit stubs; no multi-node peer, relayer, upgrade, backup/restore, or
    adversarial operations test is claimed.
  - Path: A single validator can restart successfully while multi-node or
    relayer/upgrade behavior remains incomplete.
  - Fix: Require separate IBC/upgrade and multi-node operations tickets plus an
    independent operations review before a public network.

## Verification

- `go test ./... -count=1 -timeout=600s`: PASS, 649 Go cases
- `go test ./... -json -cover -count=1 -timeout=600s`: PASS; root 64.3%, token
  92.6%, treasury 97.0%, DEX 45.3%, governance 58.9%
- Targeted root/binder/native restart race test: PASS
- `go vet ./...`, `CGO_ENABLED=1 go build ./...`: PASS
- Linker-injected `truerepublicd --version` and `truerepublicd version`: PASS
- `sh -n scripts/docker-entrypoint.sh`, `git diff --check`: PASS
- Local Docker and ShellCheck: unavailable; GitHub gates required
- Recovery total: 683 (649 Go + 26 Rust + 8 maintained-client)

## Priority matrix

### 🔴 BLOCKING

None in the locally executed single-node lifecycle slice.

### 🟠 HIGH

1. Multi-node, IBC/upgrade, backup/restore, and independent operations evidence
   is not part of GH-21 and remains required before a public network.

### 🟡 MEDIUM

1. Refreshed GitHub Docker restart evidence is pending for the rebased head.

### 🟢 LOW

None identified.
