# Network Configuration

## Connecting to a Network

### Mainnet

```bash
# In .env or config.toml
CHAIN_ID=truerepublic-1
SEEDS=<mainnet-seed-nodes>
PERSISTENT_PEERS=<mainnet-peers>
```

### Testnet

```bash
CHAIN_ID=truerepublic-testnet-1
SEEDS=<testnet-seed-nodes>
```

## Peer Discovery

### Seeds vs Persistent Peers

| Type | Purpose | Behavior |
|------|---------|----------|
| **Seeds** | Initial discovery | Connect, share peers, then disconnect |
| **Persistent Peers** | Always connected | Maintain connection permanently |

### Finding Peers

1. **Seed nodes** -- Provided by the network coordinator
2. **Peer exchange (PEX)** -- Enabled by default, discovers peers automatically
3. **Manual** -- Get node IDs from other operators

### Getting Your Node ID

```bash
truerepublicd tendermint show-node-id
# Returns: abc123def456...
```

Your full peer address: `<node-id>@<your-ip>:26656`

## Firewall Configuration

### UFW (Ubuntu/Debian)

```bash
# P2P - Required for all nodes
sudo ufw allow 26656/tcp

# RPC - Only if serving queries publicly
sudo ufw allow 26657/tcp

# Block internal services from public access
sudo ufw deny 1317/tcp    # REST/LCD
sudo ufw deny 9090/tcp    # gRPC
sudo ufw deny 26660/tcp   # Prometheus

sudo ufw enable
```

### iptables

```bash
# Allow P2P
iptables -A INPUT -p tcp --dport 26656 -j ACCEPT

# Allow RPC (optional)
iptables -A INPUT -p tcp --dport 26657 -j ACCEPT

# Block internal services
iptables -A INPUT -p tcp --dport 1317 -j DROP
iptables -A INPUT -p tcp --dport 9090 -j DROP
```

## Sentry Node Architecture

For validators, use sentry nodes to protect against DDoS:

```
                     Internet
                        │
              ┌─────────┼─────────┐
              │         │         │
          ┌───┴───┐ ┌───┴───┐ ┌───┴───┐
          │Sentry1│ │Sentry2│ │Sentry3│  Public nodes
          └───┬───┘ └───┬───┘ └───┬───┘
              │         │         │
              └─────────┼─────────┘
                        │
                  ┌─────┴─────┐
                  │ Validator │        Private (no public P2P)
                  └───────────┘
```

### Validator config.toml (private)

```toml
[p2p]
pex = false                              # Disable peer exchange
persistent_peers = "sentry1,sentry2,sentry3"
addr_book_strict = false
```

### Sentry config.toml

```toml
[p2p]
pex = true
persistent_peers = "validator-id@validator-private-ip:26656"
private_peer_ids = "validator-node-id"   # Don't share validator's address
```

## Reverse Proxy (nginx)

For public RPC access behind nginx:

```nginx
server {
    listen 443 ssl;
    server_name rpc.truerepublic.network;

    location / {
        proxy_pass http://127.0.0.1:26657;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /websocket {
        proxy_pass http://127.0.0.1:26657/websocket;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

## Next Steps

- [Genesis Parameters](genesis-params.md)
- [Security Hardening](../operations/security.md)
