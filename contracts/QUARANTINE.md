# CosmWasm prototype quarantine (issue #308)

The contracts below are **non-production prototypes. They are not deployable.**
They stay in the workspace only as historical code for host-side unit tests.
They must never be built for `wasm32`, stored, instantiated or migrated on any
chain, with or without real funds.

Fail-closed boundary:

- Every affected crate sets `publish = false`.
- Every affected crate rejects `wasm32` compilation unconditionally through a
  `#[cfg(target_arch = "wasm32")] compile_error!` guard in its `lib.rs`. No
  feature or flag re-enables a deployable build.
- No CI workflow, container image, release contract or maintained guide
  packages or deploys them. `contract_quarantine_repository_test.go` enforces
  this boundary.

| Crate | Path | Contract | Audit findings | Disposition |
|---|---|---|---|---|
| `truerepublic-contracts` | `core/` | `treasury.rs` | TRR-13 (unbacked ledger, unauthenticated withdraw), TRR-18 | archive |
| `truerepublic-contracts` | `core/` | `governance.rs` | TRR-16 (repeat voting, placeholder key), TRR-18 | archive |
| `zkp-aggregator` | `examples/zkp-aggregator/` | `contract.rs` | TRR-14 (self-asserted nullifiers, no proof verification), TRR-18 | archive |
| `dex-bot` | `examples/dex-bot/` | `contract.rs` | TRR-17 (shadow order book, no escrow or settlement), TRR-18 | archive |
| `governance-dao` | `examples/governance-dao/` | `contract.rs` | TRR-15 (zero-vote execution, no voting period), TRR-18 | quarantine |
| `token-vesting` | `examples/token-vesting/` | `contract.rs` | TRR-18; schedules are unfunded and claims withdraw from the domain treasury | quarantine |

The two library crates stay as they are and are not part of this quarantine:
`truerepublic-bindings` (maintained bindings) and `truerepublic-testing-utils`
(test helpers).

What the dispositions mean:

- **archive** — historical test material only. The native modules already
  provide the maintained function: the domain treasury, `x/truedemocracy`
  governance and ZKP voting, and `x/dex`. No rewrite is planned.
- **quarantine** — remains non-deployable until a separately reviewed
  rewrite exists. Such a rewrite needs:
  - authenticated custody;
  - escrowed funding for vesting;
  - voting-period, quorum and threshold enforcement over a snapshotted
    electorate;
  - per-voter duplicate protection and blocked zero-vote execution;
  - bounded indexed state (`cw-storage-plus`);
  - `cw2` versioning with a `migrate` entry point;
  - contract-level adversarial tests;
  - an independent contract review.

None of this changes chain-level CosmWasm upload or instantiate permissions;
that policy is tracked separately in issue #333.
