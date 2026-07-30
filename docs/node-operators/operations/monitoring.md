# Monitoring

## Built-in Monitoring Stack

TrueRepublic's Docker Compose includes Prometheus and Grafana pre-configured.
Set `PROMETHEUS_ENABLED=true` in the local `.env`; the entrypoint keeps the
node endpoint on loopback and Prometheus shares that network namespace.

### Accessing Dashboards

| Service | URL | Credentials |
|---------|-----|-------------|
| Grafana | `http://localhost:3000` | admin / `GRAFANA_PASSWORD` from .env |
| Prometheus | `http://localhost:9091` | None |

### Prometheus Targets

Prometheus scrapes CometBFT metrics from port 26660. Check target health at:
`http://localhost:9091/targets`

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

## Key Metrics

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

## Grafana Dashboard

The pre-configured dashboard (`monitoring/grafana/dashboards/`) shows:

- **Block height** over time
- **Connected peers** gauge
- **Mempool size** graph
- **Consensus rounds** histogram
- **Block interval** average
- **Missing validators** counter
- **Transactions per block** rate

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

Create alert rules in `monitoring/prometheus-alerts.yml`:

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
