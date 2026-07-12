# Current Status

**Version:** v0.4.0 recovery
**Release state:** recovery foundation merged to `main`
**Production-ready:** no

## Verified foundation

- Canonical `upnyx` base denomination with six decimals.
- Exact maximum supply of 21,000,000 PNYX / 21,000,000,000,000 `upnyx`.
- Bank-backed governance escrow, capped issuance, DEX custody, genesis
  reconciliation, and every-block supply/custody invariants on the ordered
  recovery stack.
- Chain/proposal/rating-bound ZKP statement and pinned genesis verification-key
  identity; both public web clients reject mock proof submission.
- Persistent Cosmos/Comet single-node lifecycle with generated-key,
  bank-backed PoD genesis and native plus Docker restart evidence.
- 683 recovery-verified tests: 649 Go, 26 Rust, and 8 maintained-client.

## Merged recovery sequence

PR #9 → #15 → #16 → #17 → #18 → #19 → #22 → #23 → #24.

The recovery foundation was reviewed, verified, and merged to `main` in this
order. PR #27 remains the deployment-initialization follow-up.

## Release blockers

- Release qualification and independent security review.
- Compatible real Groth16 client prover and external circuit/ceremony review.
- Privacy-preserving anonymous reward recipient binding.
- Independent multi-node, IBC/upgrade, backup/restore, monitoring, and
  operations evidence.
- Migration or removal of the deprecated legacy web/mobile clients.

See [Issue #4](https://github.com/NeaBouli/TrueRepublic/issues/4),
[`BRIDGE.md`](https://github.com/NeaBouli/TrueRepublic/blob/fix/GH-8-docs-final/BRIDGE.md),
and [`docs/status.json`](https://github.com/NeaBouli/TrueRepublic/blob/fix/GH-8-docs-final/docs/status.json).
