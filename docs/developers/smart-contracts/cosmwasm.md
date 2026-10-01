# CosmWasm Smart Contracts

> **Quarantined, not deployable (issue #308).** The CosmWasm contracts in
> `contracts/` are non-production prototypes kept only for host-side tests.
> Their crates reject `wasm32` builds, and they must not be stored, instantiated
> or migrated on any chain. See
> [`contracts/QUARANTINE.md`](../../../contracts/QUARANTINE.md). The sections
> below describe the historical prototype interfaces only.

## Overview

| Contract | File | Purpose |
|----------|------|---------|
| Governance | `contracts/src/governance.rs` | On-chain proposals with systemic consensing (-5 to +5) |
| Treasury | `contracts/src/treasury.rs` | Deposit/withdraw treasury operations |

## Tech Stack

| Technology | Version |
|-----------|---------|
| Rust | 1.75+ |
| cosmwasm-std | 3 |
| Target | host only (`wasm32` is rejected by the quarantine guard) |

## Building Contracts

Host-side build and tests only:

```bash
cd contracts
cargo build --workspace
cargo test --workspace
```

## Deploying Contracts

Not supported. The prototypes are quarantined (issue #308): there is no
deployable artifact, and no store, instantiate or migrate recipe is maintained.

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
