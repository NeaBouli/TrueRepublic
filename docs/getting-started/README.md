# Getting Started

Welcome to TrueRepublic! Choose your path:

## I want to...

### Use TrueRepublic (End User)
Participate in governance, vote on proposals, and manage PNYX tokens.

1. [Install Keplr Wallet](../user-manual/web-wallet-guide.md)
2. [Connect to TrueRepublic](../user-manual/web-wallet-guide.md#step-2-connect-to-truerepublic)
3. [Join a Domain](../user-manual/governance-tutorial.md#joining-a-domain)
4. [Start Voting](../user-manual/governance-tutorial.md#rating-suggestions-systemic-consensing)

**Estimated time:** 10 minutes

### Run a Node (Node Operator)
Run a TrueRepublic full node to support the network.

1. [Check Requirements](../node-operators/installation/requirements.md)
2. [Docker Setup](../node-operators/installation/docker-setup.md) (recommended)
   or [Native Build](../node-operators/installation/native-build.md)
3. [Configure Your Node](../node-operators/configuration/node-config.md)
4. [Set Up Monitoring](../node-operators/operations/monitoring.md)

**Estimated time:** 30 minutes (Docker) / 1 hour (native)

### Become a Validator
Secure the network and earn staking rewards.

1. [Set up a full node first](#run-a-node-node-operator)
2. [Read Validator Requirements](../validators/README.md#requirements)
3. [Register as Validator](../validators/README.md#becoming-a-validator)
4. [Set Up Monitoring](../node-operators/operations/monitoring.md)

**Estimated time:** 2 hours (including node setup)

### Build on TrueRepublic (Developer)
Integrate with or contribute to TrueRepublic.

1. [System Architecture](../developers/architecture/system-overview.md)
2. [API Reference](../developers/api-reference/cli-commands.md)
3. [CosmJS Examples](../developers/integration-guide/cosmjs-examples.md)
4. [Smart Contracts](../developers/smart-contracts/cosmwasm.md)

**Estimated time:** Variable

## Quick Install

```bash
# Clone
git clone https://github.com/NeaBouli/TrueRepublic.git
cd TrueRepublic

# Option A: Docker (everything at once)
cp .env.example .env
make docker-build && make docker-up

# Option B: Build blockchain only
make build
./build/truerepublicd start

# Option C: Web wallet development
cd web-wallet && npm install && npm start
```

## Documentation Map

```
docs/
├── getting-started/         ← You are here
├── user-manual/             End-user guides
│   ├── web-wallet-guide
│   ├── governance-tutorial
│   ├── systemic-consensing-explained
│   ├── stones-voting-guide
│   ├── dex-trading-guide
│   └── troubleshooting
├── node-operators/          Running nodes
│   ├── installation/
│   ├── configuration/
│   └── operations/
├── validators/              Validator guide
├── developers/              Developer docs
│   ├── architecture/
│   ├── api-reference/
│   ├── integration-guide/
│   └── smart-contracts/
├── FAQ.md                   Frequently asked questions
└── GLOSSARY.md              Term definitions
```
