# IBC Relayer Setup Guide

TrueRepublic supports **Inter-Blockchain Communication (IBC)** via the ICS-20 Transfer module (ibc-go v8.4.0). This guide covers relayer configuration for local testing and testnet deployment.

---

## Overview

IBC enables cross-chain PNYX transfers between TrueRepublic and any IBC-enabled Cosmos chain (Cosmos Hub, Osmosis, Neutron, etc.). The setup requires:

1. Two running chains (source + destination)
2. An IBC relayer (Hermes or Go Relayer)
3. IBC client, connection, and channel established between the chains

---

## Supported Relayers

| Relayer | Language | Recommended | Link |
|---------|----------|-------------|------|
| **Hermes** | Rust | Yes | [hermes.informal.systems](https://hermes.informal.systems) |
| **Go Relayer** | Go | Alternative | [github.com/cosmos/relayer](https://github.com/cosmos/relayer) |

---

## Local Two-Chain Testing

### Prerequisites

- Go 1.24+
- `truerepublicd` built (`make build`)
- Hermes installed (`cargo install ibc-relayer-cli`)

### Chain A: TrueRepublic

```bash
# Initialize chain A
truerepublicd init test-node-a --chain-id truerepublic-test-1 --home ~/.truerepublic-a

# Add genesis account
truerepublicd keys add validator-a --keyring-backend test --home ~/.truerepublic-a
truerepublicd genesis add-genesis-account validator-a 10000000pnyx --keyring-backend test --home ~/.truerepublic-a

# Start chain A (default ports: RPC 26657, gRPC 9090)
truerepublicd start --home ~/.truerepublic-a
```

### Chain B: Second TrueRepublic Instance (or any Cosmos chain)

```bash
# Initialize chain B with different chain-id and ports
truerepublicd init test-node-b --chain-id truerepublic-test-2 --home ~/.truerepublic-b

# Add genesis account
truerepublicd keys add validator-b --keyring-backend test --home ~/.truerepublic-b
truerepublicd genesis add-genesis-account validator-b 10000000pnyx --keyring-backend test --home ~/.truerepublic-b

# Start chain B (offset ports to avoid conflicts)
truerepublicd start --home ~/.truerepublic-b \
  --rpc.laddr tcp://0.0.0.0:26658 \
  --grpc.address 0.0.0.0:9091 \
  --p2p.laddr tcp://0.0.0.0:26656
```

### Hermes Configuration

Create `~/.hermes/config.toml`:

```toml
[global]
log_level = 'info'

[mode]
[mode.clients]
enabled = true
refresh = true
misbehaviour = true

[mode.connections]
enabled = true

[mode.channels]
enabled = true

[mode.packets]
enabled = true
clear_interval = 100
clear_on_start = true
tx_confirmation = true

[[chains]]
id = 'truerepublic-test-1'
type = 'CosmosSdk'
rpc_addr = 'http://127.0.0.1:26657'
grpc_addr = 'http://127.0.0.1:9090'
websocket_addr = 'ws://127.0.0.1:26657/websocket'
rpc_timeout = '10s'
account_prefix = 'cosmos'
key_name = 'relayer-a'
store_prefix = 'ibc'
default_gas = 200000
max_gas = 1000000
gas_price = { price = 0.025, denom = 'pnyx' }
gas_multiplier = 1.2
clock_drift = '5s'
max_block_time = '30s'
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }

[[chains]]
id = 'truerepublic-test-2'
type = 'CosmosSdk'
rpc_addr = 'http://127.0.0.1:26658'
grpc_addr = 'http://127.0.0.1:9091'
websocket_addr = 'ws://127.0.0.1:26658/websocket'
rpc_timeout = '10s'
account_prefix = 'cosmos'
key_name = 'relayer-b'
store_prefix = 'ibc'
default_gas = 200000
max_gas = 1000000
gas_price = { price = 0.025, denom = 'pnyx' }
gas_multiplier = 1.2
clock_drift = '5s'
max_block_time = '30s'
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }
```

### Create IBC Connection

```bash
# Add relayer keys (use existing validator keys or create new ones)
hermes keys add --chain truerepublic-test-1 --mnemonic-file relayer-a-mnemonic.txt
hermes keys add --chain truerepublic-test-2 --mnemonic-file relayer-b-mnemonic.txt

# Create clients on both chains
hermes create client \
  --host-chain truerepublic-test-1 \
  --reference-chain truerepublic-test-2

hermes create client \
  --host-chain truerepublic-test-2 \
  --reference-chain truerepublic-test-1

# Create connection (uses the clients created above)
hermes create connection \
  --a-chain truerepublic-test-1 \
  --b-chain truerepublic-test-2

# Create transfer channel
hermes create channel \
  --a-chain truerepublic-test-1 \
  --a-connection connection-0 \
  --a-port transfer \
  --b-port transfer

# Start the relayer
hermes start
```

### Test IBC Transfer

```bash
# Send 1000 PNYX from chain A to chain B
truerepublicd tx ibc-transfer transfer \
  transfer \
  channel-0 \
  cosmos1<recipient-on-chain-b> \
  1000pnyx \
  --from validator-a \
  --keyring-backend test \
  --home ~/.truerepublic-a \
  --chain-id truerepublic-test-1 \
  --fees 10pnyx

# Verify on chain B (after relayer processes the packet)
truerepublicd query bank balances cosmos1<recipient-on-chain-b> \
  --home ~/.truerepublic-b \
  --chain-id truerepublic-test-2

# Expected: ibc/<hash> denomination with 1000 amount
# The IBC denom is: ibc/SHA256(transfer/channel-0/pnyx)
```

---

## Testnet Deployment

### 1. Deploy TrueRepublic Testnet

Follow the [Node Setup Guide](node-operators/README.md) to deploy a TrueRepublic testnet.

### 2. Choose Target Chain

Compatible IBC chains for testnet:
- **Cosmos Hub Testnet** (theta-testnet-001)
- **Osmosis Testnet** (osmo-test-5)
- **Neutron Testnet** (pion-1)

### 3. Run Hermes on VPS

```bash
# Install Hermes on a VPS with connectivity to both chains
cargo install ibc-relayer-cli --version 1.10.0

# Configure with testnet endpoints (update config.toml)
# Use publicly available RPC/gRPC endpoints for the target chain

# Create and start the relayer
hermes create client --host-chain truerepublic-testnet-1 --reference-chain theta-testnet-001
hermes create connection --a-chain truerepublic-testnet-1 --b-chain theta-testnet-001
hermes create channel --a-chain truerepublic-testnet-1 --a-connection connection-0 --a-port transfer --b-port transfer
hermes start
```

### 4. Fund Relayer Accounts

The relayer needs tokens on both chains to pay transaction fees:

```bash
# Fund relayer on TrueRepublic
truerepublicd tx bank send validator relayer-address 100000pnyx --chain-id truerepublic-testnet-1

# Fund relayer on target chain (use that chain's faucet or send tokens)
```

---

## CLI Reference

### IBC Transfer Commands

```bash
# Send tokens cross-chain
truerepublicd tx ibc-transfer transfer [src-port] [src-channel] [receiver] [amount] [flags]

# Example: Send 5000 PNYX to Cosmos Hub
truerepublicd tx ibc-transfer transfer transfer channel-0 cosmos1abc... 5000pnyx --from user
```

### IBC Query Commands

```bash
# List all channels
truerepublicd query ibc channel channels

# List all connections
truerepublicd query ibc connection connections

# List all clients
truerepublicd query ibc client states

# Query denom traces (see IBC token origins)
truerepublicd query ibc-transfer denom-traces

# Query escrow address for a channel
truerepublicd query ibc-transfer escrow-address transfer channel-0
```

---

## Monitoring

```bash
# Hermes health check
hermes health-check

# Query pending packets
hermes query packet pending --chain truerepublic-test-1 --port transfer --channel channel-0

# Query unreceived packets
hermes query packet unreceived-packets --chain truerepublic-test-2 --port transfer --channel channel-0

# Clear pending packets manually
hermes clear packets --chain truerepublic-test-1 --port transfer --channel channel-0
```

---

## Troubleshooting

| Issue | Cause | Solution |
|-------|-------|----------|
| Connection timeout | Chain not reachable | Check RPC endpoints and firewall rules |
| Client creation fails | Clock drift too large | Sync system clocks, increase `clock_drift` |
| Channel creation fails | Connection not established | Verify connection exists: `hermes query connection connections` |
| Packets not relaying | Relayer key out of funds | Fund relayer account on both chains |
| Denom not recognized | IBC denom hash mismatch | Use `query ibc-transfer denom-traces` to find correct denom |
| Timeout on receive | Block time mismatch | Increase timeout in transfer command (`--packet-timeout-timestamp`) |

---

## Architecture Notes

- **Transfer Port:** `transfer` (bound at genesis via ICS-20 module)
- **IBC Store Key:** `ibc` (IBC core state: clients, connections, channels)
- **Capability Store:** `capability` + `memory:capability` (port/channel binding)
- **Escrow Accounts:** Per-channel escrow addresses hold locked tokens during transfer
- **Denom Format:** Received tokens use `ibc/<SHA256-HASH>` denomination
- **Unbonding Period:** 3 weeks (used by IBC light client for trust verification)

---

**Related:**
- [Installation Guide](../INSTALLATION.md)
- [Validator Guide](validators/README.md)
- [DEX Trading Guide](user-manual/dex-trading-guide.md)
