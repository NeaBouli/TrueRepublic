# Node Operators Guide

This guide covers recovery and pre-rollout node operation. It is not a
production or public-network approval.

## Table of Contents

### Installation
- [System Requirements](installation/requirements.md)
- [Docker Setup](installation/docker-setup.md) (Recommended)
- [Native Build](installation/native-build.md)
- [Reproducible Build](installation/reproducible-build.md)
- [Offline Release Evidence](installation/release-evidence.md)
- [Artifact Lifecycle](installation/lifecycle.md)

### Configuration
- [Node Configuration](configuration/node-config.md)
- [Network Configuration](configuration/network-config.md)
- [Role-Based Network Policy](configuration/network-policy.md)
- [Multi-Node Topology Qualification](configuration/topology-contract.md)
- [Private Deployment Evidence Gate](operations/deployment-evidence.md)
- [Genesis & Chain Parameters](configuration/genesis-params.md)

### Operations
- [Incident Command and Rehearsal](operations/incident-command.md)
- [Monitoring](operations/monitoring.md)
- [Backup & Recovery](operations/backup-recovery.md)
- [Validator Identity Custody and Recovery](operations/validator-identity-recovery.md)
- [Validator Consensus-Key Rotation](operations/validator-key-rotation.md)
- [Validator Slashing and Recovery](operations/validator-slashing.md)
- [Legacy Validator-Authority Migration](operations/legacy-authority-migration.md)
- [Multi-Validator Recovery Harness](operations/multi-validator-recovery.md)
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

Verify the local Compose proxy:
`curl http://localhost:8080/rpc/status`

Verify the distinct local operation signals:

```bash
docker compose exec -T truerepublic-node truerepublicd healthcheck live
docker compose exec -T truerepublic-node truerepublicd healthcheck ready
```

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
│ Client (maintained web / CLI)            │
├──────────────────────────────────────────┤
│ RPC (26657) │ REST (1317) │ gRPC (9090)  │
├──────────────────────────────────────────┤
│ Cosmos SDK Application Layer             │
│ truedemocracy │ dex │ treasury modules   │
├──────────────────────────────────────────┤
│ CometBFT Consensus (v0.38.25)            │
│ P2P (26656) │ Metrics (26660)            │
└──────────────────────────────────────────┘
```

## Ports Reference

| Port | Protocol | Service | Expose Publicly? |
|------|----------|---------|------------------|
| 26656 | TCP | P2P networking | Seed/sentry/RPC roles; named sentries only for validators |
| 26657 | TCP | CometBFT RPC | No; loopback behind a TLS proxy |
| 1317 | TCP | REST/LCD API | No (internal only) |
| 9090 | TCP | gRPC | No (internal only) |
| 26660 | TCP | Prometheus metrics | No (internal only) |

## Next Steps

- New operators: Start with [Docker Setup](installation/docker-setup.md)
- Want to validate: See the [Validator Guide](../validators/README.md)
- Need to monitor: See [Monitoring](operations/monitoring.md)
