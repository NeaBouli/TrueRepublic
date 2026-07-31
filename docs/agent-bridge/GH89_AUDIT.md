# TrueRepublic GH-89 — Audit

> Scope: strict topology contract, cross-node graph policy, public-ingress
> abuse controls, offline CLI, synthetic fixture, CI/repository contracts, and
> operator qualification guidance · Date: 2026-07-31 · Result: 0 open FAIL /
> 0 open WARN / 7 PASS

## Summary

GH-89 adds a repository-owned qualification layer above the merged GH-71
per-home role policy. A versioned, secret-free contract now correlates seeds,
sentries, validators, and RPC nodes; proves reciprocal sentry protection and
failure-zone diversity; requires every allowed flow explicitly; and rejects
public-query controls weaker than the documented ceilings.

The maintained example contains only synthetic node IDs and `.invalid` hosts.
A real operator contract remains private because publishing validator P2P
identity correlations would defeat sentry privacy. This audit does not claim a
deployment, running firewall, DNS/TLS state, provider separation, sustained
capacity, DDoS resistance, production readiness, or rollout approval.

## Findings by domain

### Strict input boundary — PASS

- **[🟠 HIGH, protected] Ambiguous or oversized contracts fail before policy
  evaluation** — `topologypolicy/parse.go`
  - Protection: one JSON document of at most 256 KiB, maximum nesting depth,
    duplicate-key rejection, unknown-field rejection, numeric type safety, and
    trailing-data rejection.
  - No environment expansion, template execution, DNS lookup, node-home read,
    or secret-bearing schema field exists.
- **[🟢 LOW, remediated] Trailing values no longer appear in errors.**
  - Review path: an appended string token could be reflected by the original
    diagnostic.
  - Fix: the error is now generic; a regression proves rejected content never
    appears.

### Identity and topology isolation — PASS

- **[🟠 HIGH, protected] Validators remain behind reciprocal sentries.**
  - Every validator dials at least two declared sentries in distinct zones,
    exposes no public P2P endpoint, and dials no other role.
  - Each sentry reciprocally declares the validator as protected and may not
    dial that validator as a persistent peer.
- **[🟡 MEDIUM, remediated] External-principal and endpoint aliases are
  canonical.**
  - Review path: a node named `internet` could collide with the flow principal;
    equivalent IPv6 spellings could evade endpoint uniqueness.
  - Fix: `internet` is reserved, parsed IPs use canonical string keys, and
    regressions cover both bypasses.
- **[🟢 LOW, remediated] Non-public address classes fail closed.**
  - Loopback, private, unspecified, link-local, multicast, limited broadcast,
    and the `localhost` namespace cannot satisfy public P2P.

### Deny-by-default flow graph — PASS

- **[🔴 BLOCKING, protected] Every path is declared twice.**
  - Inbound and outbound defaults must both be `deny`.
  - Every peer/protection relationship requires a matching directed P2P flow,
    and every node-to-node flow must be backed by a relationship.
  - Internet reaches only a declared public P2P endpoint or enabled RPC/API
    ingress and can never reach a validator. Node-to-node RPC/API and unknown
    services fail.

### Public-ingress abuse controls — PASS

- **[🟠 HIGH, protected] Query exposure has explicit ceilings.**
  - Enabled ingress is TLS-only and proxy-only, with explicit method and route
    allowlists, request rate/burst, 1 MiB body, 30-second timeout, and bounded
    concurrency ceilings.
  - Metrics, admin, unsafe, debug, profiling, Comet peer-dial, and root/wildcard
    paths are forbidden. WebSocket upgrades are limited to the allowlisted RPC
    `/websocket` path.
- **[🟡 MEDIUM, remediated] Encoded and stale routes cannot bypass evidence.**
  - Review path: percent-encoded forbidden paths could be decoded downstream;
    disabled ingress retained active-looking limits and routes.
  - Fix: encoded, escaped, control, fragment, parameter, double-slash, parent,
    wildcard, and root forms fail; disabled ingress must be empty.

### CLI and CI integration — PASS

- **[🟡 MEDIUM, remediated] Offline validation bypasses daemon config
  initialization.**
  - The first root integration regression reproduced an empty-home Cosmos
    interceptor panic.
  - `topology-policy` is now explicitly config-independent; tests prove it
    creates no home and does not print the operator-private file path or
    rejected value.
- The Go workflow triggers on contract-only changes and executes the maintained
  contract through the real daemon command with exact JSON shape checks.

### Independent review — PASS

- Kimi K3 performed the initial architecture review and the final diff review.
  Its final run reached the provider quota only after reporting the reserved
  principal, trailing-value, and special-address findings; all were fixed.
- A separate Terra remediation review confirmed those three fixes and found
  the IPv6 alias, encoded-route, and disabled-ingress cases; all were fixed.
- Spark added the bounded positive/negative test matrix. Sol reviewed all
  delegated output, found and removed one vet-invalid test-helper statement,
  and owns the final integration and complete gates.

### Documentation and rollout claims — PASS

- Operator guidance separates the synthetic committed example from a private
  real inventory, links GH-71 effective-home validation, and documents
  qualification, evidence, rollback, and isolation.
- GH-29's actual production-topology deployment checkbox remains unchecked.
  GH-89 is definition-of-ready evidence, not the deployment itself.

## Verification boundary

- Focused topology package race/coverage: PASS, 86.2%.
- Root command registration, config independence, path/value secrecy, and
  repository contract: PASS.
- Workflow YAML, example CLI JSON contract, and diff hygiene: PASS.
- Complete repository package selection, build, vet, race, and coverage: PASS.
  Coverage includes root/application 69.1% and topology policy 86.2%.
- All eight maintained multi-validator process scenarios: PASS across two
  sequential runs. The first run's 25-minute global budget expired only after
  four PASS results while the heavily loaded linker was building the fifth;
  the remaining four then all passed without code changes.
- Mandatory before merge: all exact-head GitHub build/race/coverage, recovery,
  Docker/Compose, docs/static/security, review, and unresolved-thread gates.
- Mandatory before any real rollout: separately authorized private/live
  topology, firewall, TLS/DNS, capacity, drill, and operator evidence.
