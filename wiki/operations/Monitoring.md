# Monitoring & Maintenance

Complete monitoring setup for TrueRepublic nodes.

## Table of Contents

1. [Prometheus Setup](#prometheus-setup)
2. [Grafana Dashboards](#grafana-dashboards)
3. [Key Metrics](#key-metrics)
4. [Alerting](#alerting)
5. [Log Management](#log-management)
6. [Maintenance Tasks](#maintenance-tasks)

---

## Prometheus Setup

### Installation (Docker)

Already included in docker-compose.yml:

```yaml
prometheus:
  image: prom/prometheus:latest
  ports:
    - "9090:9090"
  volumes:
    - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
```

### Configuration

**File: monitoring/prometheus.yml**

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'truerepublic-node'
    static_configs:
      - targets: ['truerepublic-node:26660']
        labels:
          instance: 'node-1'
          type: 'validator'

  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['alertmanager:9093']

rule_files:
  - 'alerts.yml'
```

### Metrics Exposed

**CometBFT metrics (port 26660):**

```
# Consensus
tendermint_consensus_height
tendermint_consensus_validators
tendermint_consensus_validators_power
tendermint_consensus_missing_validators
tendermint_consensus_byzantine_validators

# P2P
tendermint_p2p_peers
tendermint_p2p_message_send_bytes_total
tendermint_p2p_message_receive_bytes_total

# Mempool
tendermint_mempool_size
tendermint_mempool_tx_size_bytes

# State
tendermint_state_block_processing_time
```

### Querying Metrics

```bash
# Current block height
curl localhost:9090/api/v1/query?query=tendermint_consensus_height

# Number of peers
curl localhost:9090/api/v1/query?query=tendermint_p2p_peers

# Missed blocks (last hour)
curl localhost:9090/api/v1/query?query=increase(tendermint_consensus_validators_missed_blocks[1h])
```

---

## Grafana Dashboards

### Access Grafana

```
URL: http://your-server-ip:3000
Username: admin
Password: admin (change on first login!)
```

### Add Prometheus Data Source

1. Settings -> Data Sources
2. Add data source -> Prometheus
3. URL: `http://prometheus:9090`
4. Save & Test

### Import Dashboard

**File: monitoring/grafana/dashboards/truerepublic.json**

```json
{
  "dashboard": {
    "title": "TrueRepublic Node",
    "panels": [
      {
        "title": "Block Height",
        "targets": [
          {
            "expr": "tendermint_consensus_height"
          }
        ]
      },
      {
        "title": "Peers",
        "targets": [
          {
            "expr": "tendermint_p2p_peers"
          }
        ]
      },
      {
        "title": "Missed Blocks",
        "targets": [
          {
            "expr": "increase(tendermint_consensus_validators_missed_blocks[1h])"
          }
        ]
      }
    ]
  }
}
```

**Import:**

1. Dashboards -> Import
2. Upload JSON file
3. Select Prometheus data source
4. Import

### Key Dashboard Panels

**1. Node Status**
- Block height (current)
- Sync status (catching up?)
- Uptime

**2. Network**
- Peer count
- Inbound/outbound connections
- Network traffic

**3. Consensus**
- Block time
- Validator set size
- Missed blocks
- Voting power

**4. Performance**
- Block processing time
- Transaction throughput
- Mempool size

**5. Resources**
- CPU usage
- RAM usage
- Disk usage
- Disk I/O

---

## Key Metrics

### Critical Metrics (Monitor 24/7)

**1. Missed Blocks**

```promql
increase(tendermint_consensus_validators_missed_blocks[5m]) > 10
```

- Alert when: >10 missed blocks in 5 minutes
- Action: Check node health immediately

**2. Peer Count**

```promql
tendermint_p2p_peers < 5
```

- Alert when: <5 peers
- Action: Check network, firewall, seeds

**3. Block Height**

```promql
increase(tendermint_consensus_height[5m]) == 0
```

- Alert when: No new blocks in 5 minutes
- Action: Node is stuck or offline

**4. Disk Space**

```promql
node_filesystem_free_bytes{mountpoint="/"} / node_filesystem_size_bytes < 0.1
```

- Alert when: <10% disk free
- Action: Clean up or add storage

### Important Metrics (Check Daily)

**Block Processing Time:**

```promql
tendermint_state_block_processing_time
```

**Mempool Size:**

```promql
tendermint_mempool_size
```

**Network Traffic:**

```promql
rate(tendermint_p2p_message_send_bytes_total[5m])
```

---

## Alerting

### Alert Rules

**File: monitoring/alerts.yml**

```yaml
groups:
  - name: validator_alerts
    interval: 30s
    rules:
      - alert: ValidatorDown
        expr: up{job="truerepublic-node"} == 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Validator is down"
          description: "Validator {{ $labels.instance }} has been down for 2 minutes"

      - alert: HighMissedBlocks
        expr: increase(tendermint_consensus_validators_missed_blocks[10m]) > 50
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High missed blocks"
          description: "Validator missed {{ $value }} blocks in last 10 minutes"

      - alert: LowPeerCount
        expr: tendermint_p2p_peers < 5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Low peer count"
          description: "Only {{ $value }} peers connected"

      - alert: NodeNotSyncing
        expr: increase(tendermint_consensus_height[5m]) == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Node stopped syncing"
          description: "No new blocks in 5 minutes"

      - alert: DiskSpaceLow
        expr: node_filesystem_free_bytes{mountpoint="/"} / node_filesystem_size_bytes < 0.15
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Disk space low"
          description: "Only {{ $value | humanizePercentage }} disk space remaining"
```

### Alertmanager Setup

**File: monitoring/alertmanager.yml**

```yaml
global:
  resolve_timeout: 5m

route:
  group_by: ['alertname', 'instance']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h
  receiver: 'telegram'

receivers:
  - name: 'telegram'
    telegram_configs:
      - bot_token: 'YOUR_BOT_TOKEN'
        chat_id: YOUR_CHAT_ID
        parse_mode: 'HTML'
        message: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'

  - name: 'email'
    email_configs:
      - to: 'alerts@example.com'
        from: 'alertmanager@truerepublic.network'
        smarthost: 'smtp.gmail.com:587'
        auth_username: 'your-email@gmail.com'
        auth_password: 'your-app-password'

  - name: 'slack'
    slack_configs:
      - api_url: 'YOUR_SLACK_WEBHOOK'
        channel: '#validator-alerts'
        title: 'TrueRepublic Alert'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'
```

---

## Log Management

### View Logs

**Docker:**

```bash
docker-compose logs -f truerepublic-node
docker-compose logs --tail=100 truerepublic-node
```

**Systemd:**

```bash
sudo journalctl -u truerepublicd -f
sudo journalctl -u truerepublicd --since "1 hour ago"
```

### Log Levels

Edit `app.toml`:

```toml
[telemetry]
# Options: trace, debug, info, warn, error
log_level = "info"
```

**Levels:**
- `error` -- Only errors (production)
- `warn` -- Warnings + errors (production)
- `info` -- General info (default)
- `debug` -- Detailed logs (troubleshooting)
- `trace` -- Everything (development only)

### Log Rotation

**Logrotate config:**

```bash
sudo nano /etc/logrotate.d/truerepublic
```

```
/var/log/truerepublic/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0640 truerepublic truerepublic
}
```

---

## Maintenance Tasks

### Daily

Automated monitoring checks:

```bash
#!/bin/bash
# Check missed blocks
MISSED=$(curl -s localhost:26657/status | jq '.result.validator_info.missed_blocks_counter')
if [ "$MISSED" -gt 10 ]; then
    echo "WARNING: $MISSED missed blocks"
fi

# Check disk space
DISK=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
if [ "$DISK" -gt 80 ]; then
    echo "WARNING: Disk at $DISK%"
fi

# Check peer count
PEERS=$(curl -s localhost:26657/net_info | jq '.result.n_peers')
if [ "$PEERS" -lt 5 ]; then
    echo "WARNING: Only $PEERS peers"
fi
```

### Weekly

1. Review Grafana dashboards
2. Check for software updates
3. Verify backups
4. Check alert history

### Monthly

1. Security audit
2. Performance review
3. Capacity planning
4. Key rotation (if applicable)

---

## Next Steps

- [Troubleshooting](Troubleshooting) -- Common issues
- [Validator Guide](Validator-Guide) -- Validator operations
- [Node Setup](Node-Setup) -- Advanced configuration
