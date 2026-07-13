# TrueRepublic GH-32 — Audit
> Scope: multi-validator genesis assembly, four-process consensus/recovery harness, Go CI, operator/public status, and bridge  ·  Date: 2026-07-14  ·  Result: 0 FAIL / 2 WARN / 8 PASS

## Summary

GH-32 is conditionally ready to publish for its bounded local four-validator
scope. It generates independent private-validator homes, shares public
identities only, creates matching CometBFT and exactly bank-backed PoD genesis,
and proves quorum continuation plus deterministic catch-up and app-hash
agreement after one-validator failure. Production CometBFT defaults are not
weakened. The largest remaining risk is scope: partitions, state sync,
validator-set changes, backup/restore, upgrades, IBC, load, topology, and
independent operations review are not proven by this ticket.

## Findings by domain

### Consensus identity and custody — PASS

- **[PASS] Shared genesis is public-identity-only** — `server_lifecycle.go:267`, `multi_validator_harness_test.go:98`
  - What: Four daemon `init` calls create independent CometBFT private-validator
    homes. The shared artifact receives copied Ed25519 public keys only.
  - Path: A private validator key cannot enter `genesis.json` through the
    identity type; the harness also rejects a shared artifact containing a
    `priv_key` field.
  - Fix: Preserve the public-only type and temporary-home boundary.

- **[PASS] CometBFT and PoD validator sets are exact peers** — `server_lifecycle.go:296`
  - What: Each public key creates one equal-power CometBFT validator and one
    ordered PoD genesis validator.
  - Path: Missing names, duplicate names, malformed keys, duplicate public keys,
    or mismatched generated sets fail before the genesis file is written.
  - Fix: Keep set construction in one function and retain the key-by-key check.

- **[PASS] Validator stake remains exactly bank-backed** — `server_lifecycle.go:326`
  - What: The existing cap-aware PoD bootstrap creates exact minimum stake for
    all four validators and validates module-bank parity.
  - Path: Over-cap, malformed, duplicate, or unbacked state fails closed before
    mutation; recovered export is validated again after catch-up.
  - Fix: Keep `ensureConsensusGenesis` and `validateLedgerGenesis` mandatory.

- **[PASS] Single-node init refusal is preserved** — `server_lifecycle.go:246`
  - What: `truerepublicd init` still supplies exactly one generated public key
    and rejects any pre-existing consensus validator set.
  - Path: The internal multi-set helper cannot make a repeated operator `init`
    silently replace consensus state; byte-preservation regressions remain
    green.
  - Fix: Do not expose replacement semantics through `init`.

### Failure and persistence behavior — PASS

- **[PASS] Four validators reach a common committed state** — `multi_validator_harness_test.go:160`
  - What: All nodes start with byte-identical genesis and explicit persistent
    peers, reach height two, and return one app hash at that height.
  - Path: Peer, genesis, ABCI, PoD, or application nondeterminism prevents the
    common-height gate from passing.
  - Fix: Retain common-height comparison rather than comparing moving latest
    heights.

- **[PASS] One-validator failure preserves quorum and catches up** — `multi_validator_harness_test.go:168`
  - What: One of four equal-power validators stops; the other three commit two
    additional blocks; the stopped home restarts and catches up.
  - Path: A false-positive restart cannot pass unless all four nodes reach the
    post-failure height and agree on its app hash.
  - Fix: Keep the explicit pre-/post-failure height and app-hash gates.

- **[PASS] Shutdown, logs, and export are deterministic** — `multi_validator_harness_test.go:149`, `multi_validator_harness_test.go:183`
  - What: SIGINT must return cleanly, forced shutdown is reported, all logs are
    emitted on failure, and recovered state exports after processes close.
  - Path: Hung or non-zero shutdown, lost state, truncated export, missing PoD
    validators, or ledger divergence fails the test.
  - Fix: Preserve bounded shutdown waits and failure-log output.

### CI and documentation — PASS

- **[PASS] The expensive operations test has an isolated blocking job** — `.github/workflows/go-ci.yml:48`
  - What: `multi-validator-recovery` runs the exact documented command under an
    eight-minute timeout; the normal race suite runs the fast genesis regression.
  - Path: A peer/process failure cannot be hidden inside coverage output or
    converted into an informational result.
  - Fix: Keep both Go jobs required for node/genesis changes.

### Remaining operations scope — WARN

- **[HIGH] The harness does not prove adverse network or lifecycle operations** — `docs/ROLLOUT_ROADMAP.md:29`
  - What: Validator join/leave/replacement, partitions, delayed peers, state
    sync, backup/restore, upgrades/rollback, IBC, load, and topology remain open.
  - Path: A four-node localhost restart can pass while public-network recovery
    or operator procedures still fail.
  - Fix: Keep the remaining Phase 1/3/6 gates open and require independent
    operations evidence before any public-network approval.

- **[MEDIUM] Multi-validator genesis assembly is not yet an operator CLI** — `server_lifecycle.go:279`
  - What: The audited assembler is internal and exercised by tests; there is no
    supported production ceremony or public command for collecting validators.
  - Path: Treating this harness as deployment tooling would require manual file
    handling outside the supported `init` boundary.
  - Fix: Design an explicit reviewed genesis ceremony/tool only with the later
    validator join/replacement and key-management workstream.

## Priority matrix

### 🔴 BLOCKING

None for the bounded GH-32 harness scope.

### 🟠 HIGH

1. Complete adverse network, state-sync, disaster-recovery, upgrade, IBC, load,
   topology, and independent operations gates before public-network approval.

### 🟡 MEDIUM

1. Deliver an operator-supported multi-validator genesis ceremony together
   with validator-set and key-management procedures.

### 🟢 LOW

None identified.
