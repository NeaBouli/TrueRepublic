# TrueRepublic PR #9 — Foundation Merge Audit
> Scope: `origin/main...fix/GH-4-recovery-foundation` · Date: 2026-07-11 · Result: 0 FAIL / 3 WARN / 7 PASS

## Summary

PR #9 is suitable as a recovery foundation, not as a production release. It
changes toolchains, dependencies, CI, the maintained client, documentation, and
the agent bridge; it does not modify Go consensus, ledger, genesis, ZKP, or DEX
runtime code. Reproduced builds and tests pass, and no reachable Go or npm
vulnerability with an available fix remains. The unresolved economic and
consensus findings remain blocking for production and are tracked separately in
`CODEX_AUDIT.md` and the ordered recovery issues.

## Findings by domain

### Scope and consensus isolation — PASS

- **[PASS] No consensus or ledger implementation is mixed into the foundation** — `git diff --name-status origin/main...HEAD`
  - What: Executable changes are limited to client TypeScript, CI, toolchain and dependency metadata.
  - Path: The branch cannot silently alter validator state transitions because no Go runtime source file changes.
  - Fix: Preserve this boundary during merge.

### Go toolchain and dependencies — WARN

- **[🟡 MEDIUM] Four reachable upstream vulnerabilities have no available fix** — `go.mod`, `/tmp/truerepublic-gh9-govulncheck.txt`
  - What: `openpgp`, `shamaton/msgpack`, and two Cosmos SDK crisis findings remain `Fixed in: N/A`.
  - Path: The imports remain reachable even though the dependency update removes every reachable finding with an available fix.
  - Fix: Keep the project non-production, reduce affected import paths where feasible, and re-run govulncheck on every recovery PR.

- **[PASS] Fixable Go and standard-library findings are removed** — `go.mod`, `go.sum`, `.github/workflows/security-scan.yml`
  - What: Go 1.26.5, `golang.org/x/crypto` v0.52.0, OpenTelemetry v1.43.0,
    and the updated transitive graph clear the fixable gate.
  - Path: The exact CI filter finds no reachable `Fixed in` value other than `N/A`.
  - Fix: Retain the fail-closed fixable-vulnerability gate.

### Rust contracts and tooling — WARN

- **[🟡 MEDIUM] Six transitive Rust advisories remain in upstream/dev-tooling paths** — `contracts/Cargo.lock`, `contracts/core/Cargo.toml`
  - What: Unmaintained `paste`/proc-macro crates and unsound `memmap2`/`rkyv` versions remain through CosmWasm/Wasmer tooling; the Wasmer paths are dev dependencies.
  - Path: Contract tests and local VM tooling resolve these crates even though `cargo audit` exits successfully with allowed warnings.
  - Fix: Track upstream CosmWasm/Wasmer releases and remove warnings when compatible versions become available.

- **[PASS] Contract behavior remains green after lockfile repair** — `contracts/Cargo.lock`
  - What: 26 workspace tests and Clippy with warnings denied pass.
  - Path: The two security lockfile updates do not change contract APIs or source.
  - Fix: Preserve the lockfile and audit gate.

### Maintained client — PASS

- **[PASS] Exact PNYX formatting avoids JavaScript integer truncation** — `client-web/src/utils/format.ts`, `client-web/src/utils/format.test.ts`
  - What: String/BigInt conversion replaces floating-point amount conversion.
  - Path: The 21M cap and values above `Number.MAX_SAFE_INTEGER` round-trip in tests.
  - Fix: Keep chain amounts as strings at all service and component boundaries.

- **[PASS] Client dependency and lint recovery is reproducible** — `client-web/package.json`, `client-web/package-lock.json`
  - What: `npm ci`, lint, five tests, production build, and high-level audit pass with zero advisories.
  - Path: CI now targets `client-web`, not the deprecated wallet.
  - Fix: Retain `npm ci`; never reintroduce the legacy wallet as the canonical gate.

- **[🟢 LOW] Main client bundle remains large** — `client-web` production build
  - What: The main JavaScript bundle is 1.68 MB before gzip.
  - Path: Initial download and parse cost can degrade lower-end/mobile clients.
  - Fix: Add route-level code splitting after correctness recovery.

### CI and container reproducibility — PASS with follow-up warning

- **[PASS] Docker gate proves compatible glibc/wasmvm linkage** — `Dockerfile`, `.github/workflows/go-ci.yml`
  - What: The original Alpine image linked wasmvm's default glibc `.so` with musl and failed on unresolved `GLIBC_*` symbols.
  - Path: Debian Bookworm now builds the CGO binary, copies the architecture-specific shared library, and registers it with `ldconfig`.
  - Fix: Both push and pull-request Docker jobs pass; retain the blocking Docker gate.

- **[🟡 MEDIUM] Official actions still target deprecated Node 20 runtimes** — `.github/workflows/*.yml`
  - What: checkout v4 and setup-go v5 are forced onto Node 24 by current runners.
  - Path: Future runner enforcement can turn warnings into CI failures.
  - Fix: GH-8 PR #24 already upgrades the action majors; keep that ordered follow-up.

### Public status and recovery boundaries — PASS

- **[PASS] Production claims are explicitly withheld** — `README.md`, `docs/status.json`, `docs/LIMITATIONS.md`, `CODEX_AUDIT.md`
  - What: Public surfaces label recovery active, legacy wallets unsafe, and ledger findings blocking.
  - Path: Readers are directed to GH-4 and cannot reasonably treat PR #9 as mainnet approval.
  - Fix: Keep public status synchronized after every ordered merge.

## Priority matrix

### 🔴 BLOCKING

None inside the PR #9 foundation scope. The seven separate ledger/consensus blockers in `CODEX_AUDIT.md` still block production use.

### 🟠 HIGH

None introduced by PR #9.

### 🟡 MEDIUM

1. Track the four reachable Go findings without upstream fixes.
2. Track and remove the six transitive Rust tooling warnings when upstream permits.
3. Land the Node-24 action-major update in the ordered GH-8 documentation/CI PR.

### 🟢 LOW

1. Split the 1.68 MB client bundle after the recovery foundation is stable.
