# GH-74 Node Health Signals Audit

Status: locally verified; final-head GitHub runtime evidence pending  
Issue: [GH-74](https://github.com/NeaBouli/TrueRepublic/issues/74)  
Parent: [GH-29](https://github.com/NeaBouli/TrueRepublic/issues/29)

## Scope

GH-74 separates node liveness from readiness without changing consensus,
transactions, application state, or production infrastructure:

- `truerepublicd healthcheck live` checks that local CometBFT RPC answers
  `/health` with HTTP 200 and a JSON-RPC 2.0 result object.
- `truerepublicd healthcheck ready` checks `/status`, a positive committed
  height, and `catching_up=false`.
- Docker uses only liveness as its restart signal. Kubernetes guidance uses
  distinct exec probes so synchronization cannot cause a restart loop.
- GitHub container gates require both signals and post-restart height progress.

## Security and failure boundaries

- Probe targets must use plain HTTP and a literal IPv4/IPv6 loopback address.
  Userinfo, query strings, fragments, non-root paths, hostnames, public,
  unspecified, and non-loopback addresses fail before dialing.
- Environment proxies and HTTP redirects are disabled.
- Probe timeout must be positive and at most ten seconds; response reads are
  capped at 64 KiB.
- JSON-RPC version, result shape, sync state, and height are validated
  explicitly.
- Operator-facing failures are stable and do not include URLs, local paths,
  response bodies, userinfo, or credential material.
- The commands bypass node-home initialization and do not create or mutate
  configuration.

## Local evidence

- `go test ./healthcheck -race -count=2`: PASS.
- Focused root command/config independence tests: PASS.
- `go vet ./healthcheck` and `go vet .`: PASS.
- Authoritative package-scoped Go test run: 1,009 cases PASS:
  root 71, healthcheck 55, migration 82, network policy 126, token 12,
  treasury 36, DEX 116, and governance 511.
- Full `make verify`: PASS. Coverage: root 69.5%, healthcheck 97.2%,
  migration 84.6%, network policy 95.5%, token 92.6%, treasury 97.0%,
  DEX 45.3%, and governance 62.2%.
- Combined public total: 1,043 cases (1,009 Go, 26 Rust, and eight maintained
  client).

## Review

Kimi K3 implemented the bounded probe package, CLI, and focused tests. Sol
reviewed every resulting diff and additionally disabled redirects and
environment proxies, required JSON-RPC 2.0, removed an environment-sensitive
port assumption, and kept the new concurrent checks race-clean. Kimi's final
independent read-only review found no P0/P1/P2 issue. Sol remediated its one
actionable P3 by preventing an empty follow-up height read from poisoning the
GitHub restart baseline; two non-blocking maintainability/test observations
remain documented. Architecture, security, integration, documentation,
GitHub writes, and closure remain with Sol.

## Audit result

- FAIL: 0
- WARN: 1
- PASS: 12

The warning is evidence-related, not an identified implementation defect:
Docker is unavailable in the local environment, so the real image/Compose
liveness and readiness paths must pass on the exact final GitHub head before
merge.

CodeRabbit's final-head review opened six actionable threads. Sol retained the
compromised-key runbook rehearsal boundary, corrected the migration-focused
harness wording, made the deployment curl fail on HTTP errors, corrected the
Docker unhealthy-versus-restart semantics, used context-aware ephemeral
listening in the test, and bounded documentation total matches against adjacent
digits/commas. All six require focused local verification, refreshed exact-head
CI, and resolved GitHub threads before merge.

## Non-claims

This task does not deploy a node, configure a server, firewall, DNS, or real
topology, authorize a public network, prove production monitoring coverage, or
change consensus. Structured logging, broader metrics/alerting,
load/capacity/topology qualification, independent operations review, and all
other GH-29 exit gates remain separate work.
