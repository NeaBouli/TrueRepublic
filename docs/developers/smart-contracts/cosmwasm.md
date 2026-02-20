# CosmWasm Smart Contracts

TrueRepublic includes CosmWasm smart contracts for governance and treasury operations.

## Overview

| Contract | File | Purpose |
|----------|------|---------|
| Governance | `contracts/src/governance.rs` | On-chain proposals with systemic consensing (-5 to +5) |
| Treasury | `contracts/src/treasury.rs` | Deposit/withdraw treasury operations |

## Tech Stack

| Technology | Version |
|-----------|---------|
| Rust | 1.75+ |
| cosmwasm-std | 1.5 |
| Target | wasm32-unknown-unknown |

## Building Contracts

### Prerequisites

```bash
# Install Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# Add WASM target
rustup target add wasm32-unknown-unknown
```

### Build

```bash
cd contracts

# Debug build
cargo build

# Production build (optimized WASM)
cargo build --release --target wasm32-unknown-unknown

# Output: target/wasm32-unknown-unknown/release/*.wasm
```

### Optimize (for deployment)

```bash
# Using cosmwasm optimizer
docker run --rm -v "$(pwd)":/code \
    cosmwasm/workspace-optimizer:0.15.0
```

## Deploying Contracts

```bash
# Store contract on chain
truerepublicd tx wasm store governance.wasm \
    --from wallet --gas auto --fees 10000pnyx

# Get code ID from transaction result
CODE_ID=1

# Instantiate contract
truerepublicd tx wasm instantiate $CODE_ID '{}' \
    --from wallet --label "governance-v1" \
    --admin $(truerepublicd keys show wallet -a) \
    --gas auto --fees 10000pnyx
```

## Governance Contract

### Messages

```rust
// Submit a proposal with systemic consensing rating
ExecuteMsg::SubmitProposal {
    domain: String,
    issue: String,
    suggestion: String,
    rating: i8,         // -5 to +5
}

// Rate an existing proposal
ExecuteMsg::Rate {
    domain: String,
    issue: String,
    suggestion: String,
    rating: i8,         // -5 to +5
    domain_pub_key: String,  // For anonymous voting
}
```

### Queries

```rust
// Get all proposals for a domain
QueryMsg::GetProposals { domain: String }

// Get specific proposal
QueryMsg::GetProposal {
    domain: String,
    issue: String,
    suggestion: String,
}
```

## Treasury Contract

### Messages

```rust
// Deposit funds into treasury
ExecuteMsg::Deposit {
    domain: String,
    amount: Uint128,
}

// Withdraw funds from treasury
ExecuteMsg::Withdraw {
    domain: String,
    amount: Uint128,
    recipient: String,
}
```

### Queries

```rust
// Get treasury balance
QueryMsg::GetBalance { domain: String }
```

## Contract Architecture

The smart contracts complement the native Go modules:

```
Native Modules (Go)              CosmWasm Contracts (Rust)
┌──────────────────┐              ┌──────────────────┐
│ truedemocracy    │◄────────────►│ governance.rs    │
│ (keeper.go)      │  Interop     │ (proposals, SC)  │
├──────────────────┤              ├──────────────────┤
│ treasury         │◄────────────►│ treasury.rs      │
│ (rewards.go)     │  Interop     │ (deposit/withdraw)│
└──────────────────┘              └──────────────────┘
```

- Native modules handle core state and consensus logic
- Smart contracts provide programmable extensions
- Both can be used depending on the use case

## Testing

```bash
cd contracts
cargo test
```

## Next Steps

- [System Architecture](../architecture/system-overview.md)
- [Module Reference](../architecture/module-reference.md)
- [CosmJS Examples](../integration-guide/cosmjs-examples.md)
