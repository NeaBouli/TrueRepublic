# Monitoring

## Built-in Monitoring Stack

TrueRepublic's Docker Compose includes Prometheus and Grafana pre-configured.

### Accessing Dashboards

| Service | URL | Credentials |
|---------|-----|-------------|
| Grafana | `http://localhost:3000` | admin / `GRAFANA_PASSWORD` from .env |
| Prometheus | `http://localhost:9091` | None |

### Prometheus Targets

Prometheus scrapes CometBFT metrics from port 26660. Check target health at:
`http://localhost:9091/targets`

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
