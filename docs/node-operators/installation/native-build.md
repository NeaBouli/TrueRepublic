# Native Build

Build and run TrueRepublic directly on your system without Docker.

## Prerequisites

- Go 1.26.6 ([download](https://go.dev/dl/))
- Make
- Git

Verify Go installation:
```bash
go version
# go version go1.26.6 linux/amd64
```

## Step 1: Clone Repository

```bash
git clone https://github.com/NeaBouli/TrueRepublic.git
cd TrueRepublic
```

## Step 2: Build

```bash
# Build binary to ./build/truerepublicd
make build

```

Verify:
```bash
./build/truerepublicd --help
```

## Step 3: Initialize Node

```bash
# Set environment variables
export CHAIN_ID=truerepublic-1
export MONIKER=my-node
export BINARY=./build/truerepublicd
export BOOTSTRAP_OPERATOR=truerepublic1... # independently controlled account

# Initialize generated-key, exact bank-backed PoD genesis
./scripts/init-node.sh
```

The wrapper calls only `truerepublicd init` and binds the public
`BOOTSTRAP_OPERATOR` identity separately from the generated consensus key. It
does not create a keyring mnemonic, private account key, staking gentx, or
additional token supply.

This creates:
```
~/.truerepublic/
├── config/
│   ├── app.toml          # Application configuration
│   ├── config.toml       # CometBFT configuration
│   ├── genesis.json      # Chain genesis state
│   ├── node_key.json     # Node identity key
│   └── priv_validator_key.json  # Validator signing key
└── data/                 # Blockchain state database
```

## Step 4: Start Node

```bash
BINARY=./build/truerepublicd ./scripts/start-node.sh

# Or directly:
./build/truerepublicd start --home "$HOME/.truerepublic"
```

The node starts with:
- P2P listening on port 26656
- RPC on port 26657
- REST/LCD on port 1317
- gRPC on port 9090

## Step 5: Verify

```bash
curl http://localhost:26657/status | jq .result.sync_info.latest_block_height
```

## Running as a System Service

### systemd (Linux)

Create `/etc/systemd/system/truerepublicd.service`:

```ini
[Unit]
Description=TrueRepublic Node
After=network.target

[Service]
Type=simple
User=truerepublic
EnvironmentFile=/etc/truerepublic/lifecycle.env
ExecStartPre=/usr/local/libexec/truerepublic-install-lifecycle --contract=${LIFECYCLE_CONTRACT} --prefix=${INSTALL_PREFIX} --operator-state=${OPERATOR_STATE} --sha256=${ARTIFACT_SHA256} --source-ref=${SOURCE_REF} --target=${TARGET} --runtime=${RUNTIME} pre-start
ExecStart=/opt/truerepublic/bin/truerepublicd start --home /home/truerepublic/.truerepublic
Restart=on-failure
RestartSec=10
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

Build `./cmd/install-lifecycle` from the same reviewed source commit and install
it root-owned at `/usr/local/libexec/truerepublic-install-lifecycle`. Install
the reviewed contract root-owned, mode `0644`, in a root-owned non-writable
directory outside the managed prefix so the service user can read it. Create
the
root-owned, mode `0600` `/etc/truerepublic/lifecycle.env` with fixed release
evidence values for `LIFECYCLE_CONTRACT`, `INSTALL_PREFIX`, `OPERATOR_STATE`,
`ARTIFACT_SHA256`, `SOURCE_REF`, `TARGET`, and `RUNTIME`; do not put shell
syntax or secrets in it. The complete `ExecStartPre` above prevents startup
when the identity-bound lifecycle gate fails.

After verifying that pre-start gate manually, enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable truerepublicd
sudo systemctl start truerepublicd

# Check status
sudo systemctl status truerepublicd

# View logs
sudo journalctl -u truerepublicd -f
```

### launchd (macOS)

The verified artifact lifecycle currently supports Linux only. macOS remains a
foreground local-development build target; no reviewed `launchd` installation
or unattended-service procedure is provided. Do not adapt the Linux `/opt`
layout to macOS without a separate contract, tests, and operator review.

## Build Targets

| Command | Description |
|---------|-------------|
| `make build` | Build binary to `./build/truerepublicd` |
| `make test` | Run all tests with race detector |
| `make lint` | Run vet and staticcheck |
| `make clean` | Remove build artifacts |

## Next Steps

- [Artifact lifecycle](lifecycle.md) -- checksum-bound installation, upgrade,
  rollback, and uninstall without touching operator state
- [Node Configuration](../configuration/node-config.md)
- [Monitoring](../operations/monitoring.md)
- [Validator Guide](../../validators/README.md)
