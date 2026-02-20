# System Requirements

## Minimum Requirements

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| CPU | 2 cores | 4+ cores |
| RAM | 4 GB | 8+ GB |
| Storage | 100 GB SSD | 250+ GB NVMe SSD |
| Network | 100 Mbps | 1 Gbps |
| OS | Ubuntu 22.04+ / Debian 12+ / macOS 13+ | Ubuntu 24.04 LTS |

## Software Requirements

### Docker Deployment
- Docker 24.0+
- Docker Compose v2.20+

### Native Build
- Go 1.23.5+
- Make
- Git

### Optional
- Rust 1.75+ (for CosmWasm contracts)
- Node.js 20+ (for web wallet development)

## Network Requirements

### Ports

| Port | Direction | Required | Purpose |
|------|-----------|----------|---------|
| 26656/tcp | Inbound + Outbound | Yes | P2P peer communication |
| 26657/tcp | Inbound | Optional | RPC endpoint for clients |
| 1317/tcp | Inbound | No | REST API (keep internal) |
| 9090/tcp | Inbound | No | gRPC (keep internal) |
| 26660/tcp | Inbound | No | Prometheus metrics (internal) |

### Bandwidth

- Initial sync: ~10-50 GB (depends on chain height)
- Ongoing: ~1-5 GB/day (depends on transaction volume)
- P2P requires stable connectivity for block propagation

## Storage Considerations

- **Block data** grows over time (~1-5 GB/month initially)
- **State database** depends on number of domains and pools
- **Pruning** can reduce storage (default pruning retains recent state)
- Use **SSD storage** -- HDDs will cause consensus timeouts

## Validator-Specific Requirements

If you plan to run a validator, additional requirements apply:

| Resource | Validator Requirement |
|----------|----------------------|
| Uptime | 99.9%+ recommended |
| Missed blocks | Max 50 in 100-block window before slashing |
| Stake | Minimum 100,000 PNYX |
| Domain membership | At least one domain |
| Key security | Hardware security module (HSM) recommended |

See the [Validator Guide](../../validators/README.md) for full details.
