# TrueRepublic Validator Identity Recovery — Audit
> Scope: `validator_identity_recovery_harness_test.go`, shared smoke helpers, backup/restore scripts, Go CI, validator/operator recovery and security docs · Date: 2026-07-23 · Result: 0 FAIL / 1 WARN / 5 PASS

## Summary

The GH-55 path is safe to ship as bounded cold-failover evidence after its
local and GitHub gates pass. The source signer is stopped before the coupled
consensus key and current signer state are captured; the replacement retains a
new P2P identity, and the old signer remains unreachable while the replacement
signs. Sanitized chain data is moved through the maintained backup/restore
scripts, while the consensus identity is transferred separately with owner-only
permissions. True protocol-level key rotation remains a high-severity rollout
gap and is explicitly tracked in GH-56 rather than being misrepresented as
implemented.

## Findings by domain

### Consensus single-signer safety — PASS

- **[BLOCKING] Source signer must stop before state capture** —
  `validator_identity_recovery_harness_test.go`
  - What: The harness gracefully stops the source and proves its RPC is
    unreachable before reading the consensus key and signer state.
  - Path: This prevents a final source signature after an older state snapshot
    from making the recovered signer regress and double-sign.
  - Fix: No fix required; keep this ordering mandatory.

### Identity separation — PASS

- **[HIGH] P2P and consensus identities remain distinct** —
  `validator_identity_recovery_harness_test.go`,
  `docs/node-operators/operations/validator-identity-recovery.md`
  - What: The recovery home retains a newly generated node key while receiving
    the unchanged registered consensus identity.
  - Path: A node-key change cannot be mistaken for consensus-key rotation, and
    the recovered peer cannot silently reuse the source P2P identity.
  - Fix: No fix required.

### Custody and secret handling — PASS

- **[BLOCKING] Routine archives exclude signer secrets** —
  `scripts/backup.sh`, `scripts/restore.sh`,
  `docs/node-operators/operations/backup-recovery.md`
  - What: Sanitized artifacts exclude node keys, validator keys, signer state,
    and keyrings. The project does not emit a plaintext identity archive.
  - Path: Ordinary remote or scheduled backups cannot silently become a copy of
    live consensus authority.
  - Fix: No fix required; an HSM/KMS path needs separate review.

### Recovery correctness — PASS

- **[HIGH] Recovered signer state is exact, then advances** —
  `validator_identity_recovery_harness_test.go`
  - What: The pre-start signer state is byte-identical to the post-stop source
    state and must advance strictly after recovery starts.
  - Path: Stale, reset, or inert signing state causes a deterministic test
    failure instead of being accepted as recovery.
  - Fix: No fix required.

### Chain and ledger continuity — PASS

- **[HIGH] Consensus and accounting survive failover** —
  `validator_identity_recovery_harness_test.go`
  - What: The real four-validator drill checks unchanged consensus power,
    common-height app hashes, valid export, exact bank backing, and re-import.
  - Path: A replacement that signs but diverges application state or ledger
    claims cannot pass the gate.
  - Fix: No fix required.

### Protocol key rotation — WARN

- **[HIGH] Authenticated atomic rotation is not implemented** —
  `x/truedemocracy/validator.go`, `server_lifecycle.go`
  - What: Existing removal withdraws stake subject to the domain transfer limit,
    and bootstrap operator authority is derived from the consensus key.
  - Path: Remove plus re-register can fail, alter economic state, and cannot
    safely recover a bootstrap validator or revoke a compromised old key.
  - Fix: Complete GH-56 with separate operator authority, atomic key
    replacement, deterministic power transition, permanent revocation,
    export/import preservation, and multi-validator failure tests.

## Priority matrix

### 🔴 BLOCKING

None inside the bounded GH-55 cold-failover scope.

### 🟠 HIGH

1. Complete GH-56 before claiming consensus-key rotation or compromise-key
   recovery support.

### 🟡 MEDIUM

None.

### 🟢 LOW

None.
