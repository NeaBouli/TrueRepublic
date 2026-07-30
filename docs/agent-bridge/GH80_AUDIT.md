# TrueRepublic GH-80 Metrics Baseline — Audit
> Scope: application collectors, EndBlock integration, telemetry configuration,
> private scrape topology, process/Compose evidence, and operator claims  ·
> Date: 2026-07-30  ·  Result: 0 FAIL / 2 WARN / 11 PASS

## Summary

GH-80 adds a bounded two-source baseline without widening the existing
loopback-only listener policy. CometBFT remains the source for consensus,
peer, and block signals; the Cosmos SDK API endpoint supplies recent SDK,
Go/process, and five fixed-cardinality TrueRepublic application/invariant
series. Successful local and exact-head process/container evidence covers
start, persistent restart, metric progression, exact PNYX cap arithmetic,
disabled-route 404 semantics at the nginx boundary plus disabled SDK-route
`501 Not Implemented` semantics. PR #81 passed the complete repository,
recovery, Docker/Compose, security, DeepScan, and review gates and merged as
`e629374`.

## Findings by domain

### Scope and network boundary — PASS

- **[LOW] No public listener or production topology change** —
  `monitoring/prometheus.yml`, `scripts/docker-entrypoint.sh`,
  `docs/node-operators/operations/monitoring.md`
  - What: Both scrape targets remain literal loopback endpoints in the node's
    shared network namespace.
  - Path: Prometheus reaches `127.0.0.1:26660` and `127.0.0.1:1317`; nginx does
    not publish raw metrics (`/api/metrics` and descendants return 404), and no
    host/server/firewall action exists.
  - Fix: None in GH-80; retain the role-policy and production deployment gate.

### Enable/disable semantics — PASS

- **[LOW] One validated switch controls both metrics sources** —
  `scripts/docker-entrypoint.sh`, `scripts/init-node.sh`,
  `init_node_script_test.go`
  - What: `PROMETHEUS_ENABLED` accepts only true/false/1/0, maps to both
    CometBFT and SDK telemetry, disables hostname/service labels, and uses a
    60-second SDK retention window.
  - Path: Invalid values stop before daemon execution. False disables
    CometBFT and the SDK sink; a real process reaches the unregistered SDK
    gateway fallback and returns `501 Not Implemented` from `/metrics`.
  - Fix: None.

### Metric cardinality and privacy — PASS

- **[LOW] Custom families are fixed and label-free** —
  `observability/app_metrics.go`, `observability/app_metrics_test.go`
  - What: Five `truerepublic_*` families carry no user, address, transaction,
    request, peer, hostname, service, or deployment labels.
  - Path: Registration tests require exactly one unlabeled series per family;
    user activity cannot create additional custom series.
  - Fix: None.

### SDK metric cardinality — WARN

- **[MEDIUM] Generic SDK query telemetry can derive recent series from query
  paths** — Cosmos SDK `baseapp/abci.go`,
  `docs/node-operators/operations/monitoring.md`
  - What: Enabling SDK telemetry also enables upstream query-path metrics,
    whose family names are not governed by the five custom collector rules.
  - Path: A same-host or reviewed-proxy caller can issue varied ABCI query
    paths and create recent SDK series until the 60-second retention expires.
  - Fix: Keep the endpoint loopback-only, retain proxy/rate limits, and require
    capacity/cardinality qualification before any production exposure.

### EndBlock and invariant correctness — PASS

- **[LOW] Success is recorded only after all modules and crisis invariants
  pass** — `app.go`, `app_metrics_test.go`
  - What: Crisis remains the last EndBlocker with check period one. A broken
    invariant panics before `app.mm.EndBlock` can return success.
  - Path: Error returns record nothing; successful returns update application
    and invariant heights without changing the ABCI response or state.
  - Fix: None. The coupling guard must fail if order or period changes.

### Commit interpretation — WARN

- **[LOW] Application height proves EndBlock success, not durable commit** —
  `observability/app_metrics.go`,
  `docs/node-operators/operations/monitoring.md`
  - What: Metrics update after successful EndBlock but before an independently
    observed durable CometBFT commit.
  - Path: A later commit/storage failure could leave the application gauge
    temporarily ahead of `cometbft_consensus_height`.
  - Fix: Operators must compare both sources; names and documentation do not
    claim committed durability.

### Supply and precision — PASS

- **[LOW] PNYX supply and fixed-cap headroom remain exact** —
  `observability/app_metrics.go`, `observability/app_metrics_test.go`
  - What: Canonical bank supply and `21,000,000 * 1,000,000` base-unit cap are
    below the exact-integer limit of float64.
  - Path: Boundary tests cover zero, one, cap-minus-one, and cap. Negative
    headroom is not clamped, so a hypothetical breach remains visible.
  - Fix: None.

### Registration and concurrency — PASS

- **[LOW] Metrics failures cannot become consensus failures** —
  `observability/app_metrics.go`, `observability/app_metrics_test.go`
  - What: Isolated registration is all-or-nothing; the process default is a
    mutex-protected singleton; collisions fall back to unregistered collectors.
  - Path: Repeated and concurrent app construction cannot panic or
    double-register. Counter/gauge updates are race-safe.
  - Fix: None.

### Process evidence — PASS

- **[LOW] Native start/restart/disable behavior is exercised** —
  `server_lifecycle_test.go`
  - What: The compiled daemon exposes SDK/runtime/custom families, preserves
    advancing app/invariant heights after a persistent restart, reconciles
    supply plus headroom to 21 trillion `upnyx`, and returns `501 Not
    Implemented` when disabled.
  - Path: The test also keeps structured logs, exported state, and ledger
    behavior in the existing lifecycle path. The SDK emits the one-shot
    `truerepublic_server_info` signal near startup, but neither the process nor
    durable Compose contract depends on it; both rely on non-expiring
    Go/process and TrueRepublic collectors.
  - Fix: None.

### Compose and exact container evidence — PASS

- **[LOW] Exact-head Docker/Compose proves the private runtime contract** —
  `.github/workflows/go-ci.yml`, `docker-compose.yml`
  - What: Final head `4b880b4` requires both scrape targets, exact emitted
    families, cap arithmetic, public-proxy denial, and post-restart
    progression.
  - Path: GitHub job `90743893853` passed the complete stack in 7m43s before
    PR #81 merged.
  - Fix: None. Keep this gate mandatory for affected paths.

### CI policy — PASS

- **[LOW] Metrics configuration changes trigger the runtime gate** —
  `.github/workflows/go-ci.yml`, `network_policy_repository_test.go`
  - What: `monitoring/**` and `scripts/init-node.sh` now trigger Go CI; static
    tests pin both jobs, private endpoints, telemetry flags, and runtime metric
    assertions. Runtime assertions exclude SDK series that legitimately expire
    after the configured 60-second recent-activity window.
  - Path: A config-only regression cannot silently skip the Compose gate.
  - Fix: None.

### Documentation truth — PASS

- **[LOW] Operator claims distinguish baseline from production monitoring** —
  `docs/node-operators/operations/monitoring.md`,
  `docs/node-operators/configuration/node-config.md`
  - What: Sources, semantics, retention, cardinality risk, private access,
    disabled behavior, and EndBlock-versus-commit meaning are explicit.
  - Path: Dashboard provisioning, panels, alert rules, objectives, paging,
    ownership, capacity, and production exposure remain open rather than
    implied complete.
  - Fix: None.

### Dashboard and alerting boundary — PASS

- **[LOW] GH-80 does not overclaim the next Phase-6 gate** —
  `docs/node-operators/operations/monitoring.md`,
  `docs/ROLLOUT_ROADMAP.md`
  - What: Grafana health is not treated as proof that the legacy dashboard
    definition provisions or renders correctly; alert examples remain
    illustrative.
  - Path: Dashboard runtime evidence, application panels, rules, objectives,
    escalation ownership, and paging remain separate work.
  - Fix: None in GH-80.

## Priority matrix

### 🔴 BLOCKING

None in the reviewed local implementation.

### 🟠 HIGH

None.

### 🟡 MEDIUM

1. Keep generic SDK query telemetry private and capacity-qualify its
   query-path cardinality before production use.

### 🟢 LOW

1. Always compare application EndBlock height with CometBFT committed height.
