# GH-71 Role-Based Network Policy Audit

> Scope: `networkpolicy/`, daemon/startup integration, Docker/Compose/nginx/
> monitoring defaults, CI smoke evidence, operator documentation, and public
> status synchronization · Date: 2026-07-29 · Final result: 0 FAIL / 1 WARN /
> 13 PASS

## Boundary

This is repository-only recovery evidence. No real topology, server, firewall,
DNS, credential, key, deployment, production, mainnet, or public-network action
was performed or approved.

## Findings

### FAIL

None.

### WARN

1. **Compose is a recovery/development profile, not a rollout role.** Native
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
13. PR #72 exact final head passed the complete GitHub Compose smoke: RPC
    through nginx, wallet routing, healthy node metrics, and Grafana health.

## Local evidence

- `make verify`: PASS; build, vet, Race, and coverage all green.
- Final coverage: root 69.5%, migration 84.6%, network policy 95.5%, token
  92.6%, treasury 97.0%, DEX 45.3%, governance 62.2%.
- Process evidence: seven scenarios passed in the combined matrix; concurrent
  review/test load exhausted the global 1500-second budget during the eighth
  slashing scenario. The isolated slashing rerun then passed in 303.98 seconds.
  GitHub exact final-head eight-scenario matrix passed in one 14m29s job.
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

## Merge evidence

PR #72 exact head `83ffaad86274efdf3d2bb54dfff6ce8a625d5dee`
passed Go build/vet/Race/coverage, the eight-scenario process matrix, full
Docker/Compose runtime smoke, docs/static/security checks, and DeepScan. There
were zero review findings and zero unresolved threads. CodeRabbit remained
stuck in `Review in progress` for about ten hours without emitting a result;
the stale external status was documented on the PR and bypassed under owner
authorization after the independent review and every technical gate passed.
PR #72 squash-merged as
`9c369ac0d589f749e055af33a03f5f4981020101` and closed GH-71.
