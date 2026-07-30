# Monitoring

## Built-in Monitoring Stack

TrueRepublic's Docker Compose includes Prometheus and Grafana pre-configured.
The repository pins Prometheus 3.13.1 and Grafana 13.1.1 by tag and multi-
architecture digest so local and GitHub validation use the same monitoring
engines. Reproducible project-owned release artifacts, signatures, SBOM, and
provenance remain a separate release-engineering gate.
Set `PROMETHEUS_ENABLED=true` in the local `.env`; the entrypoint keeps the
two node endpoints on loopback and Prometheus shares that network namespace.
The same switch enables CometBFT instrumentation and the Cosmos SDK Prometheus
sink. When it is false, CometBFT instrumentation is disabled and the
application `/metrics` route is not registered; the Cosmos SDK gateway
fallback returns `501 Not Implemented`.

### Accessing Dashboards

| Service | URL | Credentials |
|---------|-----|-------------|
| Grafana | `http://localhost:3000` | admin / `GRAFANA_PASSWORD` from .env |
| Prometheus | `http://localhost:9091` | None |

### Prometheus Targets

Prometheus scrapes two private targets:

| Job | Private endpoint | Evidence |
|-----|------------------|----------|
| `truerepublic-node` | `127.0.0.1:26660/metrics` | CometBFT consensus, block, peer, and mempool signals |
| `truerepublic-app` | `127.0.0.1:1317/metrics?format=prometheus` | Cosmos SDK, Go/process, and bounded TrueRepublic application signals |

Both local targets carry the bounded static `node="local"` label so
cross-source divergence queries pair the CometBFT and application view of the
same node. A future multi-node scrape configuration must assign one stable,
unique `node` value to each matching endpoint pair; never use a user,
transaction, address, peer, or request-derived value.

Check both targets at `http://localhost:9091/targets`. They must remain
same-host/private. Nginx returns 404 for `/api/metrics`, so the SDK endpoint
cannot leak through the query proxy.

## Liveness and Readiness

The daemon exposes two local, fail-closed CLI probes:

```bash
truerepublicd healthcheck live
truerepublicd healthcheck ready
```

- `live` calls the loopback CometBFT `/health` endpoint and succeeds when the
  local RPC process returns a valid JSON-RPC response. It deliberately ignores
  synchronization so an orchestrator does not restart a healthy syncing node.
- `ready` calls `/status` and succeeds only when the response is valid, the
  latest block height is positive, and `catching_up` is false. Use it to decide
  whether the node may receive query traffic.

Both probes accept `--timeout` from greater than zero through 10 seconds and a
plain-HTTP `--rpc-url` with a literal loopback address only. They reject
userinfo, redirects, proxies, queries, fragments, and non-root paths, and never
include response bodies or URLs in errors.

The container image uses `live` for its Docker `HEALTHCHECK`. For Kubernetes,
keep the RPC listener private and use separate exec probes:

```yaml
livenessProbe:
  exec:
    command: ["/usr/local/bin/truerepublicd", "healthcheck", "live", "--timeout", "2s"]
readinessProbe:
  exec:
    command: ["/usr/local/bin/truerepublicd", "healthcheck", "ready", "--timeout", "2s"]
```

These signals prove local process/synchronization state only. They do not prove
validator signing, peer diversity, application invariants, or production
rollout readiness.

## Verified Metrics Baseline

The repository's process and Compose gates require the stable categories and
custom metric families below. The upstream `truerepublic_server_info` signal is
informational near startup and is not a durable gate because it legitimately
expires with the SDK retention window. The custom `truerepublic_*` collectors
carry no user, address, transaction, request-path, peer, or deployment labels.

### Block Production

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `cometbft_consensus_height` | Current block height | Should increase every ~5s |
| `cometbft_consensus_rounds` | Consensus rounds per height | > 0 indicates issues |
| `cometbft_consensus_block_interval_seconds` | Time between blocks | > 10s is concerning |

### Network Health

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `cometbft_p2p_peers` | Connected peer count | < 3 is dangerous |
| `cometbft_mempool_size` | Pending transactions | > 1000 may need attention |
| `cometbft_mempool_tx_size_bytes` | Mempool size in bytes | Monitor for growth |

CometBFT creates the peer gauge only after the first peer connection, so a
standalone or fully disconnected node can omit `cometbft_p2p_peers` instead of
exporting zero. Use `cometbft_p2p_peers or vector(0)` for dashboards and
low-peer alerts. The single-validator Compose gate verifies
`cometbft_mempool_size` as its always-instantiated network-health family; peer
connectivity remains a production-network check.

### Validator Health

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `cometbft_consensus_missing_validators` | Validators not signing | Should be 0 |
| `cometbft_consensus_byzantine_validators` | Misbehaving validators | Should be 0 |
| `cometbft_consensus_validators` | Total active validators | Should match expected |

### Runtime and Application

| Metric | Source | Meaning |
|--------|--------|---------|
| `go_goroutines` | Go runtime | Current goroutine count |
| `process_resident_memory_bytes` | Process collector | Resident memory used by the daemon |
| `truerepublic_server_info` | Cosmos SDK | Recent SDK/build identity signal |
| `truerepublic_app_last_successful_block_height` | TrueRepublic | Latest height whose full application EndBlock returned successfully |
| `truerepublic_app_completed_blocks_total` | TrueRepublic | Successful application EndBlock cycles in this process |
| `truerepublic_app_last_successful_invariant_cycle_height` | TrueRepublic | Latest height at which every registered crisis invariant passed |
| `truerepublic_token_pnyx_supply_base_units` | TrueRepublic bank state | Canonical PNYX supply in `upnyx` |
| `truerepublic_token_pnyx_supply_headroom_base_units` | TrueRepublic bank state | Remaining distance to the fixed 21,000,000 PNYX cap |

The invariant metric is valid because the crisis module is last in the
EndBlock order and its check period is one. A broken invariant panics before
EndBlock can return success. The two application heights are EndBlock-success
signals, not a substitute for independently observing a durable CometBFT
commit; compare them with `cometbft_consensus_height`.

The SDK Prometheus sink uses a 60-second retention window to limit stale or
high-cardinality SDK series. The fixed TrueRepublic collectors and Go/process
collectors do not depend on that recent-activity window. Keep both endpoints
private even with retention enabled: telemetry is operational evidence, not an
authorization or data-loss-prevention boundary. Cosmos SDK query telemetry can
derive recent series names from query paths; the short retention, private
listener, and reviewed proxy/rate-limit boundary reduce but do not eliminate
cardinality pressure. Production exposure and capacity qualification therefore
remain blocked.

## Grafana Dashboard

The Compose profile file-provisions the immutable
`TrueRepublic Blockchain Operations` dashboard with UID
`truerepublic-blockchain`. It uses the provisioned
`truerepublic-prometheus` datasource and covers:

- CometBFT height, peers, mempool, missing validators, rounds, transactions,
  and block interval;
- successful application height, application progress, consensus/application
  divergence, and invariant-cycle lag;
- canonical PNYX supply and exact fixed-cap headroom in base units; and
- Go goroutines and resident memory.

The repository gate rejects the old Grafana HTTP API wrapper shape, requires
stable panel IDs, query references, metric families, and datasource UIDs, and
uses the Compose runtime to fetch the provisioned dashboard from Grafana. It
then verifies the provisioned datasource by UID, executes a live application
query through Grafana's datasource proxy, submits every panel expression to
Prometheus, and requires the core application/supply queries to return a live
series.

This proves repository provisioning and query compatibility in the bounded
single-node Compose environment. It does not prove a production topology,
capacity qualification, long-term retention, external notification delivery,
or that every threshold is calibrated for a real validator set.

## Manual Monitoring

### Check Node Status

```bash
# Node status
curl -s http://localhost:26657/status | jq '{
  catching_up: .result.sync_info.catching_up,
  latest_height: .result.sync_info.latest_block_height,
  latest_time: .result.sync_info.latest_block_time
}'

# Peer count
curl -s http://localhost:26657/net_info | jq '.result.n_peers'

# Validator set
curl -s http://localhost:26657/validators | jq '.result.total'
```

### Check Validator Status

```bash
truerepublicd query truedemocracy validator <operator-address>
truerepublicd query truedemocracy validators
```

## Alerting

### Prometheus Alert Rules

Prometheus loads `monitoring/prometheus-alerts.yml`. The rules cover both
private scrape targets, consensus and application progress, cross-source
height divergence, invariant lag, missing validators, low peers, sustained
mempool pressure, and the PNYX cap/headroom boundary. Every rule has an
explicit activation window, severity, role owner, actionable description, and
runbook URL.

`monitoring/prometheus-alerts.test.yml` supplies deterministic evaluation
cases for every rule, including inactive-to-firing and recovery transitions
across the availability, progress, and integrity groups, through
`promtool test rules`. CI also requires Prometheus to load every alert with
healthy rule evaluation. A missing peer series is intentionally
coalesced to zero because CometBFT does not create that gauge before its first
peer event. Other required series are not coalesced: a dedicated critical
contract alert detects a missing family even while its scrape target is up.

These are conservative recovery/testnet defaults. Peer, mempool, divergence,
and timing thresholds must be recalibrated using the intended topology and
sustained-load evidence before a production go/no-go decision.

### Initial Service Objectives

The initial measurement window is seven continuous days in a private
multi-validator testnet. Restart the formal qualification window after a
monitoring configuration change. Count an unplanned loss of telemetry as
unavailable; never omit an outage from the record.

| Objective | Prometheus measurement | Initial target |
|-----------|------------------------|----------------|
| Private target availability | `avg_over_time(up{job=~"truerepublic-(node|app)"}[7d])` per target | at least 99.5% |
| Consensus progress | `avg_over_time((changes(cometbft_consensus_height{job="truerepublic-node"}[5m]) > bool 0)[7d:5m])` | at least 99% of five-minute windows |
| Application progress | `avg_over_time((increase(truerepublic_app_completed_blocks_total{job="truerepublic-app"}[5m]) > bool 0)[7d:5m])` | at least 99% of five-minute windows |
| Invariant alignment | `avg_over_time((truerepublic_app_last_successful_invariant_cycle_height{job="truerepublic-app"} == bool truerepublic_app_last_successful_block_height{job="truerepublic-app"})[7d:1m])` | 100%; any lag is a critical event |
| PNYX fixed-cap integrity | `min_over_time(truerepublic_token_pnyx_supply_headroom_base_units{job="truerepublic-app"}[7d]) >= 0` and `max_over_time(truerepublic_token_pnyx_supply_base_units{job="truerepublic-app"}[7d]) <= 21000000000000` | 100%; no breach |

These objectives qualify the repository's recovery/testnet operating model
only. They are not a production SLO commitment and do not authorize public
traffic or real funds.

### Escalation Ownership

Rule labels assign durable roles, not personal names. Before any controlled
canary, the release owner must map every role to a primary and secondary human
and configure an independently tested notification route.

| Severity | Acknowledge target | Primary role | Required escalation |
|----------|--------------------|--------------|---------------------|
| `critical` | 5 minutes | rule's `owner` label | secondary operator immediately; protocol/security owner for invariant or PNYX events; release/governance owner if progress or integrity cannot be restored |
| `warning` | 30 minutes | rule's `owner` label | secondary operator when the condition survives one additional alert window or threatens an objective |

Prometheus evaluates and displays alerts, but the repository intentionally
does not configure email, chat, webhook, PagerDuty, or other Alertmanager
destinations. No alert should be called operationally active until a controlled
end-to-end paging drill proves delivery, acknowledgement, and handoff.

### Alert Response Runbook

#### Target unavailable

1. Confirm whether only one scrape target is down; do not infer node failure
   from telemetry loss alone.
2. Run the local `healthcheck live` and `healthcheck ready` probes and inspect
   structured logs without exposing secrets.
3. Restore the private listener or Prometheus path. Escalate instead of
   widening the listener or proxying `/metrics` publicly.

#### Required metric missing

1. Confirm the target is reachable and query its private raw metrics endpoint.
2. Compare the missing family with the repository contract, telemetry toggle,
   binary version, and last monitoring change.
3. Roll back the telemetry/configuration regression or escalate it. Do not
   suppress the contract alert or substitute a constant value.

#### Consensus or application progress stalled

1. Compare CometBFT height, application height, peers, missing validators,
   readiness, and the last successful invariant height.
2. Preserve logs and current state before restart. Do not restore stale
   validator signing state or run two copies of one consensus key.
3. Follow the validator-failure and recovery procedures; escalate any
   multi-validator halt to the protocol and release owners.

#### Application divergence or invariant lag

1. Stop rollout changes and record both heights plus the alert start time.
2. Confirm both private scrape targets are healthy and compare committed app
   hashes across trusted nodes.
3. Treat persistent divergence or any invariant lag as a protocol/security
   incident. Do not clear it by suppressing the rule or rewriting state.

#### Validator, peer, or mempool degradation

1. Compare the condition across independent sentry/validator views.
2. Check expected validator membership, peer diversity, network policy, disk,
   resource pressure, and transaction ingress.
3. Escalate sustained degradation; do not expose validator P2P/RPC/API
   listeners or weaken rate limits as a shortcut.

#### PNYX cap or headroom breach

1. Halt rollout activity and preserve the exact height, app hash, supply, and
   headroom evidence.
2. Compare canonical bank supply and module custody using the existing ledger
   validation paths.
3. Escalate immediately to protocol, security, and governance owners. Do not
   mint, burn, migrate, or edit genesis/state as an ad-hoc repair.

## Log Monitoring

The supported daemon-operation path uses structured JSON at every log level.
`scripts/start-node.sh`, the container image, and the Compose profile pass
`--log_format=json`; direct `truerepublicd start` also refuses an unstructured
format. Every supported node log line is JSON with stable `level`, `message`,
and component fields supplied by the SDK and CometBFT.

The central logger replaces sensitive values with `[REDACTED]`. This includes
credentials and authorization headers, cookies and session values, mnemonics,
private or validator keys and signer state, keyrings, passwords/passphrases,
raw transaction bytes or payloads, proofs, and signatures. Public hashes,
heights, durations, module names, peer counts, and public keys remain usable
for diagnosis.

Redaction is defense in depth, not permission to put secrets into log calls,
command lines, environment variables, tickets, or support bundles. Never
enable a bypassing logger, paste a mnemonic/private key into a field with an
invented name, or treat retained logs as public. Apply host access control,
rotation, retention, encryption, and incident handling separately; their
production values remain rollout gates.

`truerepublicd start` rejects `--trace-store` because raw SDK KV tracing bypasses
this logger. Go runtime panic/crash output is not a structured logger event and
must be handled as sensitive incident evidence rather than ingested as trusted
JSON.

Parse and filter logs structurally rather than grepping free-form text:

```bash
docker compose logs --no-log-prefix truerepublic-node |
  jq -c 'select(.level == "error") | {time, level, module, message, height}'
```

Container `json-file` rotation is bounded to three 50 MiB files in the local
Compose profile. That recovery/development setting is not a production
retention policy.

### Docker logs

```bash
# Follow node logs
docker compose logs -f truerepublic-node

# Last 100 lines
docker compose logs --tail=100 truerepublic-node
```

### systemd logs (native)

```bash
sudo journalctl -u truerepublicd -f
sudo journalctl -u truerepublicd --since "1 hour ago"
```

## Next Steps

- [Backup & Recovery](backup-recovery.md)
- [Security Hardening](security.md)
