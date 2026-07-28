# Node Configuration

## Configuration Files

All configuration is stored in `~/.truerepublic/config/`:

| File | Purpose |
|------|---------|
| `config.toml` | CometBFT consensus, P2P, RPC settings |
| `app.toml` | Application-level settings (gas, pruning, API) |
| `genesis.json` | Chain genesis state and parameters |

## Key config.toml Settings

### P2P Configuration

```toml
[p2p]
# Public seed/sentry/RPC roles bind one explicit interface and advertise it.
laddr = "tcp://YOUR_INTERFACE_IP:26656"
external_address = "tcp://YOUR_PUBLIC_ADDRESS:26656"

# Seeds for initial peer discovery
seeds = "node-id@seed1.truerepublic.network:26656"

# Persistent peers (always maintain connection)
persistent_peers = "node-id@peer1.truerepublic.network:26656"

# Maximum number of peers
max_num_inbound_peers = 40
max_num_outbound_peers = 10

# Peer exchange (share peers with other nodes)
pex = true
```

Validator and private roles instead use `laddr = "tcp://127.0.0.1:26656"`,
leave `external_address` empty, disable PEX, and dial only reviewed persistent
peers. Validate the complete role profile before startup.

### RPC Configuration

```toml
[rpc]
# RPC listen address
laddr = "tcp://127.0.0.1:26657"
unsafe = false
cors_allowed_origins = []
pprof_laddr = ""

# Maximum number of open connections
max_open_connections = 900
```

### Consensus Configuration

```toml
[consensus]
# Timeout for proposing a block
timeout_propose = "3s"

# Timeout for prevote/precommit
timeout_prevote = "1s"
timeout_precommit = "1s"

# Timeout for committing a block
timeout_commit = "5s"
```

### Prometheus Metrics

```toml
[instrumentation]
# Enable Prometheus metrics
prometheus = true

# Metrics listen address
prometheus_listen_addr = "127.0.0.1:26660"
```

## Key app.toml Settings

### Minimum Gas Price

```toml
# Minimum gas price for transactions
minimum-gas-prices = "1000upnyx"
```

### API Configuration

```toml
[api]
# Enable REST API
enable = true

# REST API address
address = "tcp://127.0.0.1:1317"
enabled-unsafe-cors = false

[grpc]
# Enable gRPC
enable = true

# gRPC address
address = "127.0.0.1:9090"
```

Client-facing listeners stay on loopback. A deliberately public query role
must use a separately reviewed TLS reverse proxy and pass the
[role-based network-policy validator](network-policy.md); never bind RPC, REST,
gRPC, gRPC-web, pprof, or metrics directly to a public/wildcard interface.

### Pruning

```toml
# Pruning strategy
# "default" - keep last 100 states
# "nothing" - keep all states (full node / archive)
# "everything" - keep only current state (minimal disk)
pruning = "default"
```

## Performance Tuning

### For Validators

```toml
# config.toml
[consensus]
timeout_commit = "5s"         # Standard block time
skip_timeout_commit = false   # Don't skip (ensures consistency)

[mempool]
size = 5000                   # Mempool transaction limit
cache_size = 10000            # Transaction cache

[p2p]
max_num_inbound_peers = 40
max_num_outbound_peers = 10
flush_throttle_timeout = "100ms"
```

### For Full Nodes (RPC)

```toml
# config.toml
[rpc]
max_open_connections = 900
max_subscription_clients = 100

# app.toml
[api]
enable = true
max-open-connections = 1000
```

### For Archive Nodes

```toml
# app.toml
pruning = "nothing"           # Keep all historical state
```

## Next Steps

- [Network Configuration](network-config.md)
- [Role-Based Network Policy](network-policy.md)
- [Genesis Parameters](genesis-params.md)
- [Monitoring](../operations/monitoring.md)
