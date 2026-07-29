# Network Configuration

## Connecting to a Network

### Candidate network

```bash
# Set CHAIN_ID in .env. Set peer topology only in config.toml.
CHAIN_ID=<reviewed-chain-id>
# config.toml:
seeds = "<canonical-seed-node-endpoints>"
persistent_peers = "<canonical-persistent-peer-endpoints>"
```

### Testnet

```bash
CHAIN_ID=truerepublic-testnet-1
# config.toml: seeds = "<testnet-seed-nodes>"
```

`scripts/start-node.sh` deliberately rejects `SEEDS` and `PERSISTENT_PEERS`
environment variables. Peer topology is reviewed persistent state, not an
environment-substitution boundary.

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
# P2P - public seed/sentry/RPC roles only
sudo ufw allow 26656/tcp
# Validator/private roles instead deny public P2P; permit only reviewed
# sentry paths if the host design requires inbound P2P.

# Public query traffic terminates at a reviewed TLS reverse proxy
sudo ufw allow 443/tcp

# Block direct client and internal services from public access
sudo ufw deny 26657/tcp   # CometBFT RPC
sudo ufw deny 1317/tcp    # REST/LCD
sudo ufw deny 9090/tcp    # gRPC
sudo ufw deny 9091/tcp    # gRPC-web
sudo ufw deny 26660/tcp   # Prometheus

sudo ufw enable
```

### iptables

```bash
# Allow P2P only on public seed/sentry/RPC roles
iptables -A INPUT -p tcp --dport 26656 -j ACCEPT

# Public query traffic terminates at the reviewed TLS reverse proxy
iptables -A INPUT -p tcp --dport 443 -j ACCEPT

# Block direct client and internal services
iptables -A INPUT -p tcp --dport 26657 -j DROP
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
persistent_peers = "<sentry-id>@<sentry-host>:26656,..."
unconditional_peer_ids = "<sentry-node-ids>"
max_num_inbound_peers = 0
addr_book_strict = true
```

### Sentry config.toml

```toml
[p2p]
pex = true
external_address = "<sentry-public-host>:26656"
persistent_peers = "<upstream-sentry-or-seed-id>@<host>:26656,..."
private_peer_ids = "<validator-node-id>"   # Do not gossip validator identity
unconditional_peer_ids = "<validator-node-id>"
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

- [Role-Based Network Policy](network-policy.md)
- [Genesis Parameters](genesis-params.md)
- [Security Hardening](../operations/security.md)
