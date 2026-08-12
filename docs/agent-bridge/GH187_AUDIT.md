# GH-187 Protocol Compatibility Boundary Audit

Status: local implementation, process, repository, and security gates verified;
protected publication pending. The deterministic Linux build is CI-owned
because its contract rejects the local macOS host.

## Scope

GH-187 retires success-like staking/distribution compatibility stubs without
adding `x/staking`, `x/distribution`, delegated staking, or fabricated PoS
state. It freezes the supported Cosmos/IBC surface while preserving PoD,
CosmWasm, ICS-20, and the governed `x/upgrade` path.

## Findings and remediation

- wasmd's default query plugins could turn the prior compatibility keepers
  into empty or zero-valued successful staking/distribution responses.
- Default message encoders could construct standard staking/distribution
  messages that failed only later because their module handlers were absent.
- The ibc-go staking compatibility keeper exposed a fabricated three-week
  unbonding duration even though TrueRepublic does not implement PoS staking.
- `docs/API.md` advertised a staking REST route that the application does not
  mount.

The application now installs explicit staking/distribution query and message
rejectors. Required IBC/CosmWasm keeper methods return one stable unsupported
surface error, and the IBC adapter no longer invents an unbonding period.
Regression tests prove `x/staking` and `x/distribution` are absent from module
basics, runtime modules, stores, default genesis, gRPC, message routers, and
root query/transaction CLI commands.

## Verification so far

- PASS: focused protocol/IBC/module boundary tests.
- PASS: fresh standard-suite count, 1,461 Go cases.
- PASS: `make ibc-two-chain`, including transfer/ACK/timeout/replay/recovery,
  channel close/replacement, and compatible binary restart (51.06s root).
- PASS: `make governed-upgrade`, four-validator halt/failure/rollback/recovery
  (236.47s test).
- PASS: documentation consistency, JSON parsing, security contract, gofmt and
  diff whitespace checks.
- PASS: full Go build/vet/Race/Coverage, critical coverage floors, generative
  property/fuzz checks, and shared-state contention/replay/restart.
- PASS: pinned Staticcheck, exact-policy Govulncheck plus fixtures, Gitleaks,
  maintained-client clean install/lint/141 tests/build/audit, and Rust
  format/Clippy/build/26 tests/advisory audit.
- EXPECTED LOCAL PLATFORM BLOCK: the clean-commit deterministic daemon command
  correctly rejects macOS with `linux-amd64 requires a native Linux x86_64
  runner`; the protected reproducible-Linux CI job must supply this evidence.

Kimi K3 performed a bounded, secret-free read-only architecture review and
identified the success-like query behavior, default message encoder exposure,
and fabricated IBC unbonding duration. Kimi produced no accepted write diff;
Codex Sol implemented and reviewed all changes and retains security,
integration, full-test, GitHub, merge, and closure ownership.

Kimi's final read-only diff review found one public test-attribution issue:
three pages incorrectly grouped GH-187's four standard-suite cases with
separately gated process evidence. Sol corrected all three; GH-187 is now
explicitly counted inside the 1,461 Go cases, while only GH-175/GH-178/GH-181
and GH-184 remain described as separate process gates.

## Residual boundary

Upstream ibc-go/wasmd interfaces still require compatibility adapters, so
dependency upgrades must re-run these regressions. External relayers,
counterparties, pre-GH-184 store introduction, IBC client upgrades, arbitrary
cross-version migrations, production deployment, real keys/funds, and release
approval remain explicitly outside this evidence.
