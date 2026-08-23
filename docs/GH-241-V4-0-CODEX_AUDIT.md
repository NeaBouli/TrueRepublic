# TrueRepublic V4-0 Canonical Edge Protocol — Audit

> Scope: `sovereignv4/protocol` and its documentation boundary  
> Date: 2026-08-23  
> Result: **0 FAIL / 0 WARN / 8 PASS**

## Summary

The V4-0 slice is an unwired, standard-library-only protocol foundation. It
does not activate a native client, peer-to-peer runtime, domain app, rollout
item, release or production path. Sol reviewed the complete diff and Kimi K3
performed an independent read-only security review. Every P1/P2 finding and
both non-blocking P3 observations from that review were remediated before this
audit was approved.

## Findings by domain

| Domain | Result | Evidence |
|---|---|---|
| Canonical codec | PASS | Fixed-order, domain-separated binary encodings reject wrong type/version, malformed lengths, non-canonical order and trailing bytes; golden, mutation, property and fuzz coverage is present. |
| Certificate identity | PASS | Account signatures bind chain, domain, member account, account key, messaging key, membership height/proof hash, epoch and validity window. External light-proof verification remains an explicit caller boundary. |
| Envelope authentication | PASS | Envelope signature and certificate hash bind author, chain, domain, topic, epoch, sequence, timestamp and exact payload hash. Valid decisions carry an unexported binding to the exact verified message ID. |
| Rotation and revocation | PASS | Stale epochs fail, future/uncommitted epochs remain provisional, exact current epochs are required, and stale/offline revocation state never upgrades authentication. |
| Replay and concurrency | PASS | Replay state is chain/domain scoped, stores canonical hashes rather than payloads, detects duplicates/equivocation, is mutex-protected and fails closed at its hard bound. |
| Domain-app authorization | PASS | Deny-by-default capability checks cross-bind verified publisher fact, domain, app ID, app version, publisher key, exact artifact hash, protocol range and signature. |
| Resource limits | PASS | All externally supplied identifiers, payloads, capabilities and replay entries have explicit bounds; invalid store sizes fail closed. |
| Tests and documentation | PASS | 15 deterministic tests plus 3 fuzz targets; focused Race/Coverage is 86.1%, Vet and pinned Staticcheck pass, and documentation consistently labels V4-0 unwired and non-production. |

## Priority matrix

No open P0, P1, P2 or P3 findings remain in this slice. Protected repository
checks, merge and exact-main/public closeout are delivery gates, not code-audit
findings.

## Residual system boundary

Membership and publisher registry proofs are deliberately caller-verified in
V4-0. Persistent replay checkpoints, transport, peer discovery, content
storage, wallet/key custody, domain-app execution and user interfaces belong
to later V4 slices and must not infer trust from this package alone.
