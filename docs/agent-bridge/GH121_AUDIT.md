# TrueRepublic GH-121 — Audit

> Scope: maintained-browser module queries, registered Go query surface,
> runtime validation, real disposable-chain evidence, dependencies, API and
> operator documentation · Date: 2026-08-08 · Result: 0 open FAIL / 2 open
> WARN / 6 PASS

## Summary

GH-121 replaces every maintained-browser custom-module REST alias with a
single registered protobuf gRPC-over-CometBFT-ABCI JSON-RPC transport. Missing
Merkle proof and Pay-to-Put reads are registered and tested. The browser now
fails closed for transport, remote, protocol, decode, and nested-shape errors
instead of treating them as empty or safe chain state.

No deployment, listener, grpc-gateway route, signing path, key material,
account, transaction broadcast, proof ceremony, or funds were changed. The
result is locally release-verifiable but is not production rollout approval.

## Findings by domain

### Query transport and protocol registration — PASS

- **[🟠 HIGH, remediated] Unregistered REST aliases could never reach the
  custom modules.** All 23 consumers across 18 unique aliases now use exact
  registered `truedemocracy.Query/*` or `dex.Query/*` method paths.
- The shared boundary encodes only known protobuf string/int64 fields, sends
  hex request bytes to `abci_query`, decodes the protobuf `Result` bytes, and
  classifies HTTP, RPC, ABCI, envelope, protobuf, JSON, and schema failures.
- A missing resource remains a chain error; a real empty list remains an empty
  list. The two states cannot collapse into each other.

### Governance, identity, and economics — PASS

- **[🔴 BLOCKING, remediated] Nullifier query failure previously became
  `false`.** It is now an explicit UI-visible failure and cannot authorize a
  duplicate-vote flow.
- **[🔴 BLOCKING, remediated] Pay-to-Put previously fabricated a default.** A
  registered chain query now calls the canonical treasury formula, validates
  unsigned numeric fields, and blocks suggestion creation when unavailable.
- Governance issue, suggestion, and rating views derive from the one canonical
  Domain response and validate its nested structure.

### Merkle proof boundary — PASS

- The registered query accepts exactly one canonical 32-byte commitment,
  rejects missing, duplicate, malformed, or inconsistent state, rebuilds the
  depth-20 MiMC tree, and refuses a root mismatch.
- Browser validation requires 64-character lowercase hex root/commitment and
  exactly 20 binary indices plus 20 sibling hashes before the proof reaches a
  consumer.

### DEX response integrity — PASS

- Pool, asset, statistics, swap estimate, spot price, and LP position payloads
  validate all required string, number, boolean, and array fields at runtime.
  TypeScript assertions alone are not trusted as a network boundary.

### Dependencies and network exposure — PASS

- **[🟠 HIGH, remediated] Live npm audit identified vulnerable `nanoid`.** The
  lock-only update to 3.3.18 clears the advisory without changing application
  dependencies or weakening the guarded audit policy.
- No new listener or route is exposed. The existing `/rpc/` reverse-proxy path
  reaches the configured CometBFT JSON-RPC service and the request retains the
  existing ingress controls.

### Independent review and verification — PASS

- Kimi K3 independently reproduced the route mismatch, counted the consumers,
  identified the nullifier and pricing fail-soft defects, and implemented the
  bounded Go query block. Sol reviewed every changed line and owns integration.
- Kimi's final read-only diff review found no P0/P1/P2. Its four substantive P3
  observations were remediated before publication: strict fixed-width protobuf
  bounds, per-item validator validation, and visible rejection handling in both
  direct LP-position consumers.
- Its post-CodeRabbit remediation review again found no P0/P1/P2. The only two
  P3 findings—response-body timeout classification and a stale privacy comment—
  were corrected and regression-tested before the final local rerun.
- Full Go, Rust, maintained-client, security, build, and real disposable-chain
  gates pass locally. Protected exact-head GitHub review and final-main checks
  remain mandatory before closure.

### Configured RPC trust — WARN (LOW)

- The browser receives authoritative query results from its configured RPC
  endpoint and does not verify headers or proofs through a browser light
  client. This is explicit and unchanged from the repository's trust model.
- Resolution belongs to a separately scoped light-client/trusted-RPC rollout
  decision; GH-121 must not imply cryptographic client-side verification.

### Client bundle scalability — WARN (MEDIUM)

- The production build is 322.63 kB gzip and still reports one minified chunk
  above 500 kB. Canonical Domain queries also return the complete nested domain
  until pagination-specific registered methods exist.
- Route-level splitting and large-state pagination remain separate rollout
  performance work. Neither warning changes GH-121 correctness, but both must
  be resolved or explicitly accepted before low-bandwidth production rollout.

## Priority matrix

| Priority | Open findings |
|---|---|
| BLOCKING | None |
| HIGH | None |
| MEDIUM | Client bundle and whole-domain scalability |
| LOW | Configured RPC trust without browser light-client verification |

## Verification boundary

- Full Go build, vet, race, coverage, and 1,352 cases: PASS.
- Rust fmt, clippy with warnings denied, build, 26 tests, and audit: PASS; five
  reviewed unmaintained/unsound transitive warnings remain allowed.
- Maintained client clean install, ESLint, 98 tests, production build, and live
  High audit: PASS.
- Real disposable-chain browser query delivery and explicit missing-pool
  failure: PASS.
- Merge evidence: independent final-head review found no open P0/P1/P2, all
  eight review threads were resolved, and every protected exact-head check
  passed before PR #126 was squash-merged as `239e6c3`.
- Mandatory before rollout: separate production infrastructure, ingress,
  trusted-RPC/light-client, performance, and operations evidence.
