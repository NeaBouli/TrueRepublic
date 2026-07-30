# Monitoring & Maintenance

> Recovery/testnet scope only. The canonical, tested operator guide is
> [`docs/node-operators/operations/monitoring.md`](../../docs/node-operators/operations/monitoring.md).
> This page summarizes that repository-owned contract. It does not authorize a
> production topology, public metrics, real-funds operation, or mainnet.

## Supported Stack

The local Compose profile pins:

- Prometheus 3.13.1, sharing the node network namespace;
- Grafana 13.1.1, exposed on host loopback only; and
- the immutable `TrueRepublic Blockchain Operations` dashboard with UID
  `truerepublic-blockchain`.

Set `PROMETHEUS_ENABLED=true` and provide a non-default `GRAFANA_PASSWORD`.
Prometheus remains on `http://127.0.0.1:9091`; Grafana remains on
`http://127.0.0.1:3000`.

## Health Signals

Use the separate local probes:

```bash
truerepublicd healthcheck live
truerepublicd healthcheck ready
```

`live` proves that loopback CometBFT RPC responds and intentionally stays
successful while the node synchronizes. `ready` additionally requires a
positive height and `catching_up=false`; use it before routing query traffic.
Neither probe proves validator signing, peer diversity, application
invariants, or rollout readiness.

## Private Metrics Targets

| Job | Endpoint | Signals |
|-----|----------|---------|
| `truerepublic-node` | `127.0.0.1:26660/metrics` | CometBFT consensus, height, peer, mempool, validator, and block timing |
| `truerepublic-app` | `127.0.0.1:1317/metrics?format=prometheus` | SDK/runtime plus bounded application, invariant, supply, and cap-headroom signals |

The SDK/API metrics route and every descendant remain blocked by the public
query proxy. Do not bind either metrics listener publicly.

## Dashboard

The provisioned dashboard covers:

- consensus height, rounds, transactions, and block interval;
- peer count, mempool pressure, and missing validators;
- successful application height/rate and consensus/application divergence;
- last successful invariant height and invariant lag;
- canonical PNYX supply and fixed-cap headroom in base units; and
- Go goroutines and resident memory.

Repository tests reject the former Grafana HTTP API wrapper shape and enforce
stable panel IDs, datasource UIDs, required metric families, and valid PromQL.
GitHub Compose CI must fetch the provisioned dashboard through Grafana and
submit every panel expression to Prometheus.

## Alert Rules

Prometheus loads `monitoring/prometheus-alerts.yml`. The eleven reviewed rules
cover:

- both scrape targets and missing required metric families;
- consensus and application progress stalls;
- consensus/application divergence and invariant lag;
- missing validators, low/absent peers, and mempool pressure; and
- exact 21,000,000 PNYX cap/headroom integrity.

Every alert includes a bounded activation window, severity, role owner,
description, and runbook URL. `monitoring/prometheus-alerts.test.yml` provides
deterministic Promtool cases for all eleven rules, including firing and
recovery behavior across the availability, progress, and integrity groups.

The low-peer fallback treats CometBFT's not-yet-created peer gauge as zero.
Other required metric families are never silently replaced with constants;
their absence raises a critical contract alert.

## Initial Recovery/Testnet Objectives

Measure seven continuous days on a private multi-validator testnet:

| Objective | Initial target |
|-----------|----------------|
| Private node/application scrape availability | at least 99.5% per target |
| Consensus progress | progress in at least 99% of five-minute windows |
| Application progress | progress in at least 99% of five-minute windows |
| Invariant alignment | 100%; any lag is critical |
| PNYX fixed-cap integrity | 100%; supply never exceeds 21,000,000,000,000 base units and headroom never becomes negative |

Restart the qualification window after a monitoring configuration change.
Count unplanned telemetry loss as unavailable. These targets are not a
production SLO commitment.

## Escalation Ownership

Rule labels assign roles rather than personal data:

| Severity | Acknowledge target | Escalation |
|----------|--------------------|------------|
| `critical` | 5 minutes | primary role owner, secondary operator, then protocol/security and release/governance owners as applicable |
| `warning` | 30 minutes | primary role owner, then secondary operator if the condition persists or threatens an objective |

Before any controlled canary, the release owner must map every role to a
primary and secondary human and test the complete delivery/acknowledgement
path. The repository intentionally ships no email, Telegram, Slack, webhook,
PagerDuty, or other Alertmanager destination.

## First Response

1. Confirm both private scrape targets and the separate liveness/readiness
   probes.
2. Compare consensus height, application height, invariant height, peers,
   missing validators, mempool, PNYX supply, and cap headroom.
3. Preserve structured logs and state evidence before a restart.
4. Never restore stale validator signing state, run two copies of one
   consensus key, expose private listeners, suppress an integrity alert, or
   rewrite chain state as an ad-hoc repair.
5. Follow the detailed response sections in the
   [canonical guide](../../docs/node-operators/operations/monitoring.md#alert-response-runbook).

## Routine Checks

Daily:

- verify both scrape targets and alert-rule health;
- review critical/warning states and structured error logs;
- confirm consensus/application/invariant height alignment; and
- confirm supply plus headroom equals exactly 21,000,000,000,000 base units.

Weekly:

- review objective measurements and alert history;
- verify backup evidence and operator-role assignments; and
- record threshold changes with a new qualification window.

Capacity, disk growth, retention, production topology, external paging, and
end-to-end incident rehearsal remain separate rollout gates.

## Next Steps

- [Canonical monitoring guide](../../docs/node-operators/operations/monitoring.md)
- [Troubleshooting](Troubleshooting)
- [Validator guide](Validator-Guide)
- [Backup and recovery](../../docs/node-operators/operations/backup-recovery.md)
