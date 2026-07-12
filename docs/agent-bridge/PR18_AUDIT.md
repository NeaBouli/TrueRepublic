# PR #18 Audit — GH-10 DEX Custody

Date: 2026-07-12
Branch: `fix/GH-10-dex-custody`
Base: `fix/GH-13-cap-issuance`
Issue: [GH-10](https://github.com/NeaBouli/TrueRepublic/issues/10)
Result: PASS for the GH-10 scope; repository remains non-production

## Scope and guarantees

- The `dex` module bank account holds the exact sum of all declared PNYX and
  paired-asset pool reserves.
- Create, proportional add, provider-owned remove, direct swap, and two-hop
  asset swap execute in `CacheContext` and commit only after bank settlement,
  reserve parity, and LP conservation succeed.
- Initial and newly minted LP shares are indexed by pool and authenticated
  provider. A caller cannot withdraw another provider's position.
- PNYX output burns call the canonical `token.IssuanceService`, reducing the
  DEX module balance and `x/bank` supply by the same exact amount.
- Registry and trading-status messages require configured chain authority.
- Legacy `MsgSwap` is disabled because it has no minimum-output protection.

## Audit findings fixed

### Critical — pool accounting did not move bank funds

All public DEX handlers previously mutated synthetic reserves without taking
inputs or paying outputs. They now use exact account-to-module and
module-to-account transfers inside the same cached transition as AMM state.

### Critical — global LP shares authorized arbitrary withdrawals

Provider-indexed balances now back every pool share. Removal checks the
authenticated sender's balance before calculating or transferring outputs.

### High — reported burns did not change canonical supply

The DEX module has burner permission and burns exact PNYX output fees through
the same issuance boundary used by recovered governance mint/burn paths.

### High — any signer could mutate the asset registry

Register/status messages now reject every signer except chain authority.

### High — textual LP prefixes allowed valid denom collisions

The original recovery patch encoded LP keys as `lp:<denom>:<provider>`. A
valid denom such as `atom:staked` therefore appeared under the `atom` iterator
prefix and inflated its provider-share total. Denoms are now length-prefixed;
a regression creates both pools and proves independent conservation.

### High — transfer rollback evidence was incomplete

KV-backed bank tests now inject failures for initial custody, add-liquidity
input, remove-liquidity output, swap output, and canonical burn. Every case
proves unchanged accounts, pools, LP ownership, analytics, and supply.

## Verification

- `go build`: PASS; CLI `--help` and `--version`: PASS
- `go vet ./...`: PASS
- `go test ./... -count=1`: PASS, 578 Go cases
- `go test -race ./... -count=1`: PASS
- `cargo test --workspace`: PASS, 26 Rust cases
- `cargo audit --no-fetch`: no blocking vulnerability; six tracked transitive
  warnings through CosmWasm/Wasmer dev tooling
- Maintained client offline install, lint, six tests, production build, and npm
  audit: PASS; zero vulnerabilities
- `go mod verify`, documentation consistency, secret/diff checks: PASS
- Local Docker rebuild unavailable because the workstation has no Docker CLI;
  PR #17 already proves the unchanged architecture-safe Dockerfile, and PR #18
  refreshed the GitHub Docker job successfully at head `3234741`.
- GitHub docs, DeepScan, Go build/vet/race/coverage, and Docker: PASS.
- Manual Security Scan
  [run 29156922464](https://github.com/NeaBouli/TrueRepublic/actions/runs/29156922464):
  PASS for govulncheck, Rust audit, canonical npm audit, and both informational
  legacy-client audit jobs.
- CodeRabbit accepted the focused review request but then reported its review
  quota exhausted for 44 minutes. No substantive external review is claimed;
  independent review remains pending.

## Explicitly out of scope

- GH-12: custom-genesis reconciliation and registered runtime invariants
- GH-7: recipient-bound anonymous rewards and ZKP authentication
- GH-21: remaining production node lifecycle/restart hardening

These slices continue to block production use and real funds.
