# Node Operators Guide

This guide covers everything you need to run a TrueRepublic node, from initial setup to production operations.

## Table of Contents

### Installation
- [System Requirements](installation/requirements.md)
- [Docker Setup](installation/docker-setup.md) (Recommended)
- [Native Build](installation/native-build.md)

### Configuration
- [Node Configuration](configuration/node-config.md)
- [Network Configuration](configuration/network-config.md)
- [Genesis & Chain Parameters](configuration/genesis-params.md)

### Operations
- [Monitoring](operations/monitoring.md)
- [Backup & Recovery](operations/backup-recovery.md)
- [Upgrades](operations/upgrades.md)
- [Security Hardening](operations/security.md)

## Quick Start

### Docker (Fastest)

```bash
git clone https://github.com/NeaBouli/TrueRepublic.git
cd TrueRepublic
cp .env.example .env    # Edit with your settings
make docker-build
make docker-up
```

Verify: `curl http://localhost:26657/status`

### Native Build

```bash
git clone https://github.com/NeaBouli/TrueRepublic.git
cd TrueRepublic
make build
./build/truerepublicd start
```

## Architecture Overview

```
┌──────────────────────────────────────────┐
│ Client (Web Wallet / CLI / Mobile)       │
├──────────────────────────────────────────┤
│ RPC (26657) │ REST (1317) │ gRPC (9090)  │
├──────────────────────────────────────────┤
│ Cosmos SDK Application Layer             │
│ truedemocracy │ dex │ treasury modules   │
├──────────────────────────────────────────┤
│ CometBFT Consensus (v0.38.21)            │
│ P2P (26656) │ Metrics (26660)            │
└──────────────────────────────────────────┘
```

## Ports Reference

| Port | Protocol | Service | Expose Publicly? |
|------|----------|---------|------------------|
| 26656 | TCP | P2P networking | Yes (required) |
| 26657 | TCP | CometBFT RPC | Optional (for queries) |
| 1317 | TCP | REST/LCD API | No (internal only) |
| 9090 | TCP | gRPC | No (internal only) |
| 26660 | TCP | Prometheus metrics | No (internal only) |

## Next Steps

- New operators: Start with [Docker Setup](installation/docker-setup.md)
- Want to validate: See the [Validator Guide](../validators/README.md)
- Need to monitor: See [Monitoring](operations/monitoring.md)
