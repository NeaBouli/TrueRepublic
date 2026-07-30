# Monitoring

## Built-in Monitoring Stack

TrueRepublic's Docker Compose includes Prometheus and Grafana pre-configured.
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

The repository's current dashboard definition
(`monitoring/grafana/dashboards/blockchain.json`) targets this CometBFT
baseline:

- **Block height** over time
- **Connected peers** gauge
- **Mempool size** graph
- **Consensus rounds** histogram
- **Block interval** average
- **Missing validators** counter
- **Transactions per block** rate

GH-80 proves Grafana process health, not successful dashboard provisioning or
panel-query rendering. Application panels, reviewed alert rules, objectives,
escalation ownership, paging integration, and dashboard runtime evidence are
not shipped. They remain the next Phase-6 gate; do not treat this baseline as
an active production monitoring program.

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

The following is an illustrative starting point only. The repository does not
currently ship `monitoring/prometheus-alerts.yml`, Alertmanager routing,
objectives, or an on-call owner:

```yaml
groups:
  - name: truerepublic
    rules:
      - alert: NodeDown
        expr: up{job="truerepublic"} == 0
        for: 1m
        labels:
          severity: critical

      - alert: BlockProductionStalled
        expr: rate(cometbft_consensus_height[5m]) == 0
        for: 2m
        labels:
          severity: critical

      - alert: LowPeerCount
        expr: cometbft_p2p_peers < 3
        for: 5m
        labels:
          severity: warning

      - alert: HighMempoolSize
        expr: cometbft_mempool_size > 1000
        for: 10m
        labels:
          severity: warning

      - alert: MissingValidators
        expr: cometbft_consensus_missing_validators > 0
        for: 5m
        labels:
          severity: warning
```

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
