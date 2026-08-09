# GH-172 Concurrency, Replay, and Restart Audit

Date: 2026-08-10
Baseline: `origin/main` `348d622`
Scope: repository tests, CI evidence, and status documentation only

## Verdict

Approved for protected pull-request verification. Independent Kimi K3 review
found no P0, P1, or P2 issue. Its one actionable P3 was remediated: the
post-restart replay assertion now requires Cosmos code 32 and the explicit
`account sequence mismatch` reason. The remediated real harness passes.

## Evidence reviewed

- Two authenticated accounts concurrently contend for one canonical domain;
  exactly one commits, the other fails with the duplicate-domain error, and
  balances, treasury, custody, and object count prove one economic effect.
- Three concurrent broadcasts use identical signed bytes; the mutex-protected
  CometBFT v0.38.25 cache admits exactly one and explicitly rejects duplicates.
- After commit the bytes remain rejected by the duplicate cache. After a clean
  same-home restart, the fresh cache reaches application CheckTx and rejects
  the old sequence with code 32.
- Historical app hash, live state, custody, balances, ledger reconciliation,
  21M cap, export singularity, and reimport remain valid.
- Standard status arithmetic is 1,607 cases: 1,440 Go, 26 Rust, and 141 client.
  The opt-in process case is excluded, and rollout remains 23/59.

## Verification

- Focused repository contract: PASS.
- Full build, Vet, Race, and Coverage across maintained Go packages: PASS.
- Generative/fuzz quality and security repository contract: PASS.
- Staticcheck v0.7.0: PASS.
- Gitleaks v8.30.1: PASS, no leak found.
- Documentation consistency, workflow YAML, status JSON, formatting, and diff
  integrity: PASS.
- Real remediated process harness: PASS in 90.93s; post-restart rejection is
  code 32 with `account sequence mismatch`.

## Residual boundary

This does not test a hostile public mempool, Byzantine multi-node consensus,
production topology, wallet custody, real keys or funds, deployment, release,
or an independent external security audit. Protected exact-head CI and final
merged-main evidence remain required before Done; production readiness remains
false.
