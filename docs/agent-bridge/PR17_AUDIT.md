# PR #17 Audit — GH-13 Canonical Reward Issuance

Date: 2026-07-11
Branch: `fix/GH-13-cap-issuance`
Base: `fix/GH-14-bank-escrow`
Issue: [GH-13](https://github.com/NeaBouli/TrueRepublic/issues/13)
Result: PASS for the GH-13 issuance scope; repository remains non-production

## Scope and guarantees

- `x/bank` `upnyx` supply is the release-decay and maximum-supply source of
  truth; the parallel `pod:total-release` counter is removed from consensus.
- `token.IssuanceService` is the only reward/slash supply-change boundary in
  the recovered governance module.
- Every mint reads canonical supply and clips the aggregate request to the
  remaining 21,000,000,000,000 `upnyx` capacity in the same cached context.
- Validator and active-domain inflation mints exact module escrow before the
  matching stake/treasury claims commit.
- Staking issuance and domain issuance share one outer EndBlock reward cache;
  failure in either phase discards both claim/timer phases.
- Treasury-funded vote rewards transfer existing escrow without changing
  supply. Anonymous rating rewards remain deferred until the proof/signature
  binds a safe recipient.
- Validator slashing burns through the same issuance service and reopens only
  the exact burned cap capacity.

## Audit findings fixed

### High — slash burns bypassed the canonical issuance boundary

The rebased GH-14 slash path called the bank keeper directly even though GH-13
declared the issuance service as the sole supply-change boundary. The slash path
now calls `IssuanceService.Burn`, preserving one audited mint/burn abstraction.

### High — two reward phases lacked a shared transaction boundary

Staking and domain issuance each used an internal cache, but a direct module
EndBlock call could publish the staking phase before a later domain mint error.
The module now nests both phases under one outer cache and commits only after
both succeed. A second-mint failure regression test proves validator claims,
domain claims, reward timers, and payout snapshots remain unchanged.

### Medium — canonical supply input was trusted without validation

The issuance service now rejects nil or negative supply values before cap math.
Tests also cover missing bank wiring and failed mint/burn operations.

### High — the node image built but every CLI/node invocation panicked

The new container smoke check exposed duplicate registration of
`legacytx.StdTx`: the app registered it directly and then registered it again
through the auth module. Both the CLI and application constructor now use one
shared codec factory, auth owns the legacy transaction registration, and a
regression test plus `truerepublicd --help`/`--version` smoke checks prove the
binary reaches Cobra without a panic.

### Major review remediation — rollback evidence used an in-memory bank mock

The second-mint failure test originally proved governance KV rollback but its
mock bank mutated Go maps outside `CacheContext`. Issuance supply and module
balance deltas now live in the mounted KV store, so the regression test proves
unchanged canonical supply, unchanged module escrow, claim/timer rollback, and
escrow parity after the second mint fails.

### Major review remediation — restored payout history lacked a baseline

Genesis now stores each restored domain's current cumulative payouts as its
interest snapshot. Pre-GH-13 state is lazily backfilled at the same baseline,
including when the interest timer is first initialized, so historical payouts
cannot earn a one-time interest windfall. New genesis and lazy-backfill tests
cover both paths.

## Boundary and negative-path evidence

- Aggregate minting stops exactly at the final cap unit.
- Supply already above cap is rejected even for a zero reward request.
- Nil, negative, and missing-bank inputs are rejected.
- First-mint and second-mint failures do not commit reward claims or timers.
- Domain interest rewards only payouts since the prior interval snapshot.
- Exact burns lower canonical supply and only that capacity can be reminted.
- Vote rewards are supply-neutral and preserve escrow parity.
- The module account has exactly the required minter/burner capabilities.
- The node image selects wasmvm from Docker's target architecture, carries the
  runtime compiler support library, and fails its build if dynamic linkage or
  the CLI entrypoint is broken.

## Verification

- `go build ./...`: PASS
- `go vet ./...`: PASS
- `go test ./... -count=1`: PASS, 569 Go cases
- `go test ./... -race -count=1`: PASS
- `go test ./... -cover -count=1`: PASS; token 93.5%, governance 55.8%
- `cargo test --workspace`: PASS, 26 Rust cases
- `cargo clippy --workspace --all-targets -- -D warnings`: PASS
- `cargo audit`: PASS with six allowed transitive dev-tooling warnings
- Maintained client install/lint/6 tests/build/audit: PASS
- Documentation consistency, workflow YAML, Dockerfile artifact mapping, and
  diff checks: PASS
- Both GitHub Docker builds: PASS, including wasmvm selection/linkage and CLI
  startup smoke check
- Refreshed GitHub Go/race/coverage, docs, DeepScan, and manual security
  workflow: PASS; zero unresolved review threads
- CodeRabbit's final 33-file review completed. Five inline findings and two
  outside-diff/nitpick findings were verified and remediated. Both Go jobs,
  both Docker builds, docs, DeepScan, CodeRabbit, and the complete manual
  security workflow pass on `0e6cf38`; all five threads are resolved.

## Explicitly out of scope

- GH-10: DEX custody, LP ownership, bank settlement, authority, and real burns
- GH-12: custom-genesis reconciliation and runtime conservation invariants
- GH-7: recipient-bound anonymous reward claims and final authentication audit

These open slices continue to block production use and real funds.
