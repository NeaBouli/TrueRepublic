# IBC Relayer Setup Guide

TrueRepublic wires **Inter-Blockchain Communication (IBC)** through the ICS-20
transfer module on ibc-go v8.7.0. This document is an operator recipe, not proof
that an external chain or relayer deployment has been qualified. The maintained
GH-175/GH-178/GH-181 gates use two isolated TrueRepublic application states and submit the
same client, connection, channel, receive, acknowledgement, and timeout
messages with real state proofs, without an external relayer process. GH-178
also proves close-confirm, timeout-on-close, exactly-once refund, persistent
closed state, and a replacement channel. Because ICS-20 rejects
user-initiated close, its initial committed CLOSED counterparty end is a
deterministic test fixture, matching ibc-go's own timeout-on-close tests.
GH-181 additionally runs the same package through a separately linked
compatible candidate test binary, reopens the existing LevelDB state without
`InitChain` or genesis export/import, and proves pending/fresh ACK plus timeout
recovery. It remains a test relay and compatible restart boundary, not daemon,
external-relayer, `x/upgrade`, migration, or public-network qualification.

**CLI boundary.** `truerepublicd` registers transaction and query commands only
for `truedemocracy` and `dex`. It has no `ibc`, `ibc-transfer`, `bank` or
`genesis` command, so IBC transfers, IBC state queries, balance checks and
genesis-account funding cannot be done with this binary's CLI. The maintained,
repository-owned evidence path is `make ibc-two-chain`, which builds funded test
genesis states and submits the IBC messages programmatically. Any manual flow
needs a separately qualified client or relayer tooling that this repository does
not provide.

---

## Overview

The wired protocol can transfer PNYX over a compatible ICS-20 counterparty. Each
specific external chain, relayer release, channel, trust period, and operational
deployment still requires separate qualification. A setup requires:

1. Two running chains (source + destination)
2. An IBC relayer (Hermes or Go Relayer)
3. IBC client, connection, and channel established between the chains

---

## Candidate Relayers (Not Yet Qualified)

| Relayer | Language | Qualification | Link |
|---------|----------|---------------|------|
| **Hermes** | Rust | Unqualified candidate | [hermes.informal.systems](https://hermes.informal.systems) |
| **Go Relayer** | Go | Unqualified candidate | [github.com/cosmos/relayer](https://github.com/cosmos/relayer) |

---

## Local Two-Chain Testing

### Prerequisites

- Go 1.26.9
- `truerepublicd` built (`make build`)
- Hermes installed (`cargo install ibc-relayer-cli`)

### Chain A: TrueRepublic

```bash
# Create the independently controlled operator account, then initialize chain A
truerepublicd keys add validator-a --keyring-backend test --home ~/.truerepublic-a
OPERATOR_A="$(truerepublicd keys show validator-a -a --keyring-backend test --home ~/.truerepublic-a)"
truerepublicd init test-node-a --chain-id truerepublic-test-1 \
  --home ~/.truerepublic-a --bootstrap-operator "$OPERATOR_A"

# Fund the genesis account
# No genesis CLI is registered: add the validator-a account and its upnyx
# balance to ~/.truerepublic-a/config/genesis.json with your own reviewed
# tooling (make ibc-two-chain does this programmatically for its test chains).

# Start chain A (default ports: RPC 26657, gRPC 9090)
truerepublicd start --home ~/.truerepublic-a
```

### Chain B: Second TrueRepublic Instance (or any Cosmos chain)

```bash
# Create a separate operator, then initialize chain B with different settings
truerepublicd keys add validator-b --keyring-backend test --home ~/.truerepublic-b
OPERATOR_B="$(truerepublicd keys show validator-b -a --keyring-backend test --home ~/.truerepublic-b)"
truerepublicd init test-node-b --chain-id truerepublic-test-2 \
  --home ~/.truerepublic-b --bootstrap-operator "$OPERATOR_B"

# Fund the genesis account
# No genesis CLI is registered: add the validator-b account and its upnyx
# balance to ~/.truerepublic-b/config/genesis.json with your own reviewed
# tooling (make ibc-two-chain does this programmatically for its test chains).

# Start chain B (offset ports to avoid conflicts)
truerepublicd start --home ~/.truerepublic-b \
  --rpc.laddr tcp://127.0.0.1:26658 \
  --grpc.address 127.0.0.1:9091 \
  --p2p.laddr tcp://127.0.0.1:26656
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
account_prefix = 'truerepublic'
key_name = 'relayer-a'
store_prefix = 'ibc'
default_gas = 200000
max_gas = 1000000
gas_price = { price = 0.025, denom = 'upnyx' }
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
account_prefix = 'truerepublic'
key_name = 'relayer-b'
store_prefix = 'ibc'
default_gas = 200000
max_gas = 1000000
gas_price = { price = 0.025, denom = 'upnyx' }
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
# truerepublicd has no ibc-transfer or bank command. Send the ICS-20 transfer
# (for example 1000upnyx over channel-0) with a separately qualified client,
# then verify the recipient balance on chain B with that client.
# Expected: ibc/<hash> denomination with 1000 base units
# The IBC denom is: ibc/SHA256(transfer/channel-0/upnyx)
```

---

## Testnet Deployment

### 1. Deploy TrueRepublic Testnet

Follow the [Node Setup Guide](node-operators/README.md) to deploy a TrueRepublic testnet.

### 2. Choose Target Chain

Potential counterparties that require their own compatibility and trust review:
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

Fund the relayer account on TrueRepublic through the qualified genesis
allocation or a separately qualified client; `truerepublicd` has no `bank send`
command, and the project operates no faucet. Fund the counterparty account with
that chain's own tooling.

---

## CLI Reference

Not available in `truerepublicd`: there are no `tx ibc-transfer`, `query ibc` or
`query ibc-transfer` commands. Inspect channels, connections, clients and denom
traces with your qualified relayer tooling or gRPC client, and keep transfer
amounts in `upnyx` (the only base denomination).

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
