# GH-71 Role-Based Network Policy Audit

> Scope: `networkpolicy/`, daemon/startup integration, Docker/Compose/nginx/
> monitoring defaults, CI smoke evidence, operator documentation, and public
> status synchronization · Date: 2026-07-29 · Local result: 0 FAIL / 2 WARN /
> 12 PASS

## Boundary

This is repository-only recovery evidence. No real topology, server, firewall,
DNS, credential, key, deployment, production, mainnet, or public-network action
was performed or approved.

## Findings

### FAIL

None.

### WARN

1. **Final-head container evidence remains external.** Docker is unavailable in
   the local environment. The workflow now starts the complete loopback Compose
   stack and verifies RPC through nginx, wallet routing, a healthy same-
   namespace node-metrics target, and Grafana health. That exact gate must pass
   on GitHub before merge.
2. **Compose is a recovery/development profile, not a rollout role.** Native
   `scripts/start-node.sh` requires role validation before startup. The
   deliberately isolated single-node Compose profile uses safe loopback
   defaults but does not claim to satisfy a seed/sentry/validator/RPC/private
   production topology. A future Compose-based rollout candidate must add an
   explicit `NETWORK_ROLE` gate.

### PASS

1. Five explicit roles parse deterministically; unknown roles fail closed.
2. Missing/malformed node homes and node identity fail closed without path or
   file-content disclosure.
3. Peer endpoints require canonical lowercase CometBFT IDs, explicit ports,
   non-wildcard production hosts, uniqueness, and no self-reference.
4. Validators accept no discovery seeds or inbound peers, disable PEX, expose
   no external address, and dial at least two pinned sentries.
5. Sentries use public explicit-IP P2P, PEX, reviewed upstream peers, and
   matching private/unconditional protected-validator IDs without requiring a
   dial-out validator peer.
6. Seed/RPC public P2P and private/outbound-only profiles enforce their role
   listener and peer-capacity requirements.
7. RPC, REST, gRPC, gRPC-web, pprof, and Prometheus are loopback-only; unsafe
   RPC, wildcard CORS, and unsafe API CORS fail.
8. Startup rejects environment-substituted peer topology and prevents daemon
   start when validation fails.
9. Docker/Compose publish only loopback development endpoints; nginx removes
   wildcard CORS and enforces request, body, and timeout limits.
10. Prometheus shares the node namespace, the entrypoint forces metrics to
    loopback, and invalid monitoring toggles fail before daemon invocation.
11. The independent Kimi K3 review found no P0/P1 and changed no files; Sol
    remediated its remaining documentation/test evidence observations.
12. The authoritative current count reproduces as 952 Go cases (root 69,
    migration 82, network policy 126, token 12, treasury 36, DEX 116,
    governance 511), plus 26 Rust and eight maintained-client cases = 986.

## Local evidence

- `make verify`: PASS; build, vet, Race, and coverage all green.
- Final coverage: root 69.5%, migration 84.6%, network policy 95.5%, token
  92.6%, treasury 97.0%, DEX 45.3%, governance 62.2%.
- Process evidence: seven scenarios passed in the combined matrix; concurrent
  review/test load exhausted the global 1500-second budget during the eighth
  slashing scenario. The isolated slashing rerun then passed in 303.98 seconds.
  GitHub must still pass the exact final-head eight-scenario matrix in one job.
- Focused startup/container/policy tests, shell syntax, YAML parsing,
  documentation consistency, JSON validation, gofmt, and `git diff --check`:
  PASS.
- Local Docker/Compose execution: not available and not claimed.

## Independent review

Kimi K3 performed a bounded read-only security/integration review. It confirmed
validator-initiated sentry topology, fail-closed parsing/startup, listener and
peer controls, secret/path safety, and exact case arithmetic. Sol corrected the
review's remaining P3 documentation/test observations and added the complete
Compose CI smoke. No independent-review blocker remains.

## Merge gate

Do not merge until the exact PR head passes Go build/vet/Race/coverage, the
eight-scenario process matrix, full Docker/Compose runtime smoke, docs/static/
security checks, review, and zero unresolved actionable threads.
