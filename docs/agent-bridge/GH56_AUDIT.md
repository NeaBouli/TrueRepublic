# GH-56 Consensus-Key Rotation Audit

Date: 2026-07-23

Scope: authenticated validator consensus-key rotation, bootstrap authority
separation, delayed CometBFT activation, permanent revocation, process-level
recovery evidence, and operator documentation.

## Result

Local result: **0 FAIL / 3 WARN / 16 PASS**. No P0 finding remains. GH-56 is
ready for final-head GitHub review and CI, but this result is recovery evidence,
not production approval.

## Passed controls

1. The transaction signer must equal the registered operator authority.
2. The signed request binds both the expected old key and the proposed new key.
3. Rotation is limited to active, non-jailed, positive-power validators.
4. Pending, duplicate, malformed, active, removed, and revoked keys fail closed.
5. Operator addresses cannot be current, historical, or cross-validator
   consensus-key-derived authorities.
6. Bootstrap rejects reserved module addresses and consensus-derived authority
   coupling before mutation.
7. Missing public operator accounts are materialized as zero-balance base
   accounts without generating a mnemonic or private key.
8. Stake, domain, power, escrow, jail, and liveness claims remain on the same
   operator during the atomic key transition.
9. CometBFT receives one old-key power-zero update and activates the replacement
   at the documented H+2 boundary.
10. Same-block and next-block inactivation do not create absent-key removal or
    repeated-zero update failures.
11. The old key remains attributable during the delayed evidence window.
12. Revocation is permanent and round-trips through export/import.
13. A stopped old signer remains offline and its signer state does not advance.
14. A pre-synchronized replacement node signs after activation and advances its
    signer state.
15. All nodes converge on one app hash and exported state re-imports cleanly.
16. Operator and compromise-response procedures document the supported boundary
    and forbid concurrent old/new signing.

## Residual warnings

- **WARN / GH-59:** production ABCI++ misbehavior and last-commit data are not
  yet wired to the economic double-sign/downtime handlers. The delayed lookup
  is present, but automatic evidence delivery remains a rollout gate.
- **WARN / GH-60:** inactive, excluded, jailed, and under-staked validator state
  is not fully export/import round-trip safe. GH-56 proves the active rotation
  and revocation path only.
- **WARN / GH-61:** homes created before separate operator authority retain the
  legacy coupled identity. They require an explicit governance-controlled
  migration or a reviewed fresh genesis; a compatible binary replacement alone
  is not a migration.

## Reproduced evidence

- Focused authority, genesis, and transition regressions: PASS.
- Full normal Go package suite: PASS.
- `go vet` over repository-owned packages: PASS.
- Documentation consistency, shell syntax, JSON, and `git diff --check`: PASS.
- `TestMultiValidatorConsensusKeyRotation`: PASS in 168.12 seconds against four
  active validators and one pre-synchronized replacement node.
- CI-only operator-prefix remediation: the legacy consensus-recovery harness
  now passes from isolated test selection in 99.33 seconds.
- Independent final read-only security review: no P0 and no additional P2
  finding; the two P1 rollout boundaries are tracked by GH-59 and GH-60.

GitHub CI and review evidence must be appended after the published final head
passes and before GH-56 is closed.
