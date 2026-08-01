# TrueRepublic GH-97 — Audit
> Scope: `capacitypolicy/`, `configs/capacity/`, capacity repository/process tests, lifecycle wiring, Go CI, operator/limitation docs, and GH-97 bridge state  ·  Date: 2026-08-01  ·  Result: 0 FAIL / 0 WARN / 7 PASS

## Summary
The repository-side synthetic capacity qualification is ready to ship: it is
strict, bounded, secret-free by construction, exercised against four real
temporary validators, and enforced in CI. The maintained run committed 96 of
96 authenticated transactions and verified consensus, restart, ledger,
metrics, resource, and snapshot-retention evidence. Independent review found
no remaining P0/P1/P2 issue after remediation, and the complete relevant local
test chain is green; one combined Race run initially exhausted the local disk,
after which every non-root package and the separately isolated root lifecycle
test passed with identical Race instrumentation. The largest remaining risk is
outside this change: a short loopback regression cannot establish multi-day
soak behavior, production traffic capacity, or real Docker/Prometheus runtime
retention, and the documentation keeps those rollout gates open.

## Findings by domain
### Architecture and scope — PASS
- **[PASS] Qualification is isolated from production and the node lifecycle** — `server_lifecycle.go:49`, `server_lifecycle.go:253`
  - What: The feature is a standalone policy command and opt-in test harness;
    normal node startup and unrelated commands remain unchanged.
  - Path: A normal daemon invocation registers `capacity-policy` but does not
    read a capacity fixture, create evidence, bind qualification ports, or
    touch a node home unless the explicit test path is selected.
  - Fix: None.

### Input validation and secrets — PASS
- **[PASS] Contract and evidence parsing fail closed** — `capacitypolicy/parse.go:38`
  - What: Inputs are size- and depth-bounded and reject duplicate keys,
    trailing values, unknown fields, empty documents, and reflective parser
    details; evidence contains only fixed identifiers, measurements, and
    booleans.
  - Path: A malformed, oversized, duplicated, or schema-expanded JSON document
    is rejected before policy evaluation, while repository tests reject secret,
    host, URL, and interpolation shapes in the maintained fixture.
  - Fix: None.

### Arithmetic and policy correctness — PASS
- **[PASS] Measured values and projections use checked arithmetic** — `capacitypolicy/validate.go:33`, `capacitypolicy/validate.go:204`
  - What: The verifier requires exact counters and fixed qualification
    metadata, checks throughput/block-time calculations, and rejects overflow,
    underflow, zero divisors, stalled progress, and inconsistent evidence.
  - Path: Manipulated evidence cannot turn an arithmetic wrap, divide-by-zero,
    missing transaction, or resource breach into a valid report; valid reports
    serialize an explicit empty violation array for stable CI matching.
  - Fix: None.

### Concurrency and lifecycle — PASS
- **[PASS] Load generation uses independent signers and bounded observation** — `capacity_sustained_load_harness_test.go:32`
  - What: Twenty-four waves submit four authenticated transactions concurrently
    through independent accounts, with per-submission timers and concurrent
    delivery checks; process, HTTP, filesystem, and restart waits are bounded.
  - Path: Signer sequence conflicts, transaction delivery failures, consensus
    stalls, divergent app hashes/power, restart failure, or a filesystem error
    fail the run instead of being hidden by shared state or unbounded waiting.
  - Fix: None.

### Runtime retention and resource evidence — PASS
- **[PASS] The harness observes the real node stores and processes** — `capacity_sustained_load_harness_test.go:663`
  - What: Disk/log growth, RSS, goroutines, required metrics, runtime snapshot
    heights, exported ledger state, and restart convergence are measured from
    the temporary nodes. Snapshot retention is observed only after new commits
    are halted so asynchronous create/prune work can settle.
  - Path: The checker counts only positive height directories in the actual
    local snapshot store, ignores only the known metadata database, and rejects
    unexpected entries; the maintained run observed exactly three heights.
  - Fix: None.

### CI and regression coverage — PASS
- **[PASS] Repository and real-process gates prevent silent drift** — `.github/workflows/go-ci.yml:119`, `capacity_policy_repository_test.go:13`
  - What: CI validates the fixed contract and runs the four-validator harness
    in a separate bounded job, then verifies the exact generated evidence.
    Repository tests structurally parse Compose logging configuration and pin
    required workflow/documentation behavior.
  - Path: Changing workload counts, removing the process gate, weakening log
    bounds, altering the fixture, or emitting invalid/null violation semantics
    makes a deterministic test or CI assertion fail.
  - Fix: None.

### Documentation and claim boundaries — PASS
- **[PASS] Operator guidance does not overstate rollout readiness** — `docs/node-operators/operations/capacity-qualification.md:3`, `docs/LIMITATIONS.md:97`
  - What: The guide supplies reproducible commands and evidence handling while
    explicitly excluding production sizing, multi-day soak, real traffic,
    Docker/Prometheus runtime retention, public topology, and rollout approval.
  - Path: A reader cannot reasonably treat the loopback projection or static
    Compose configuration as proof of hardware sizing or a live deployment
    gate without contradicting the explicit limitation sections.
  - Fix: None.

## Priority matrix
### 🔴 BLOCKING
None.

### 🟠 HIGH
None.

### 🟡 MEDIUM
None.

### 🟢 LOW
None.
