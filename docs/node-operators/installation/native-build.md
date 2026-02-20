# Native Build

Build and run TrueRepublic directly on your system without Docker.

## Prerequisites

- Go 1.23.5+ ([download](https://go.dev/dl/))
- Make
- Git

Verify Go installation:
```bash
go version
# go version go1.23.5 linux/amd64
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

# Or install to $GOPATH/bin
make install
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

# Initialize (creates ~/.truerepublic/)
./scripts/init-node.sh
```

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
./scripts/start-node.sh

# Or directly:
./build/truerepublicd start
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
ExecStart=/usr/local/bin/truerepublicd start
Restart=on-failure
RestartSec=10
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

Enable and start:
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

Create `~/Library/LaunchAgents/com.truerepublic.node.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.truerepublic.node</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/truerepublicd</string>
        <string>start</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
```

```bash
launchctl load ~/Library/LaunchAgents/com.truerepublic.node.plist
```

## Build Targets

| Command | Description |
|---------|-------------|
| `make build` | Build binary to `./build/truerepublicd` |
| `make install` | Install to `$GOPATH/bin` |
| `make test` | Run all tests with race detector |
| `make lint` | Run vet and staticcheck |
| `make clean` | Remove build artifacts |

## Next Steps

- [Node Configuration](../configuration/node-config.md)
- [Monitoring](../operations/monitoring.md)
- [Validator Guide](../../validators/README.md)
