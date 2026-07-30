# TrueRepublic GH-77 — Secret-Safe Structured Logging Audit
> Scope: `observability/`, daemon logger interception, native/container start
> paths, Docker log CI, operator guidance, and GH-77 coordination · Date:
> 2026-07-30 · Result: 0 FAIL / 1 WARN / 16 PASS

## Summary

The daemon's Cosmos SDK and CometBFT logger now crosses one defensive wrapper
that preserves JSON structure and safe operational metadata while minimizing
credential, key, signer, transaction, proof, and signature data. Native start,
restart, explicit/environment log-format overrides, raw trace-store refusal,
and the logger's adversarial value surface have focused regression evidence.
No consensus, application state, production collector, server, or public
network is changed. Exact head `fd8378c` passed all 11 GitHub checks and merged
through PR #78 as `133fb3b`; both review threads are resolved.

## Findings by domain

### Central logger boundary — PASS

- **[PASS] SDK and CometBFT share the wrapped server logger** —
  `server_lifecycle.go:188`, `server_lifecycle.go:210`
  - What: startup keeps the SDK interception path, obtains its configured
    server logger, and wraps it before command execution.
  - Path: server, BaseApp, CometBFT adapters, API, gRPC, and module child
    loggers derive through `With`; no repository or pinned runtime dependency
    consumes `Logger.Impl`.
  - Fix: none.

- **[PASS] Child context cannot unwrap the base logger** —
  `observability/logger.go:9`, `observability/logger.go:39`
  - What: all four levels and `With` sanitize; `Impl` returns the wrapper rather
    than the underlying zerolog implementation.
  - Path: a downstream `With` call retains the boundary and a direct
    `Impl().(*zerolog.Logger)` bypass is unavailable.
  - Fix: none.

### Secret and private-data minimization — PASS

- **[PASS] Sensitive structured keys are normalized and redacted** —
  `observability/sanitize.go:41`, `observability/sanitize.go:78`
  - What: case and separators cannot bypass key classification for credentials,
    authentication, passwords, mnemonic/private/signing material, signer state,
    key stores, raw transactions, payloads, proofs, signatures, cookies,
    sessions, and auth tokens.
  - Path: a sensitive key causes wholesale replacement before its value can
    invoke custom formatting.
  - Fix: none.

- **[PASS] Free text covers reviewed credential and transaction forms** —
  `observability/sanitize.go:19`, `observability/sanitize.go:98`
  - What: PEM private-key blocks, authorization/Bearer/Basic values, URL
    userinfo, multi-word mnemonics/passwords, raw transactions, proofs, and
    signatures are replaced by one stable marker.
  - Path: messages, errors, and Stringer output pass through the same compiled,
    linear-time patterns.
  - Fix: none.

- **[PASS] Nested values are copied and recursively minimized** —
  `observability/sanitize.go:215`, `observability/sanitize.go:230`
  - What: interfaces, pointers, maps, structs, arrays, and slices are walked
    without changing caller-owned data; unknown binary payloads fail closed.
  - Path: explicit public keys and reviewed public hash fields retain copied
    bytes; other binary material becomes `[REDACTED]`.
  - Fix: none.

- **[PASS] Malformed values cannot panic or echo panic text** —
  `observability/sanitize.go:141`, `observability/sanitize.go:399`
  - What: non-string keys, odd pairs, typed nils, cycles, panicking errors, and
    panicking Stringers are bounded and recovered.
  - Path: recovery occurs before the base logger can format the unsafe value.
  - Fix: none.

- **[PASS] Work and output are bounded** — `observability/sanitize.go:13`,
  `observability/sanitize.go:290`
  - What: text, recursion, and collection sizes have deterministic caps.
  - Path: an attacker-influenced error/value cannot force unbounded reflective
    traversal or a single unbounded logger event through this wrapper.
  - Fix: none.

- **[LOW] Heuristic redaction is not general DLP** —
  `observability/sanitize.go:95`, `docs/node-operators/operations/monitoring.md`
  - What: an unlabeled bare secret or a secret deliberately placed under an
    invented safe-looking name cannot be identified perfectly.
  - Path: a future developer violates the logging contract and emits secret text
    without a recognized label or structured sensitive key.
  - Fix: retain the explicit operator/developer prohibition, keep regression
    fixtures current, and treat logs as sensitive rather than public.

### Structure and configuration — PASS

- **[PASS] Node operation defaults to JSON and rejects bypasses** —
  `server_lifecycle.go:248`, `server_lifecycle.go:295`
  - What: native start marks JSON as an explicit Cobra value, rejects explicit
    plain flags and reviewed environment aliases, validates the effective SDK
    value, and rejects raw KV trace stores.
  - Path: Viper/config precedence cannot leave an effective plain logger; raw
    trace output cannot silently bypass the wrapper.
  - Fix: none.

- **[PASS] Config-independent commands remain independent** —
  `server_lifecycle.go:188`, `server_lifecycle.go:307`
  - What: health and network-policy checks still avoid node-home/config
    initialization and logging interception.
  - Path: probes retain their bounded, operator-safe standalone behavior.
  - Fix: none.

- **[PASS] Safe metadata and level filtering remain functional** —
  `observability/logger_test.go`
  - What: module, height, duration, peer count, reviewed public keys/hashes, and
    configured log levels survive the wrapper.
  - Path: operators can query structured failure evidence without private
    payloads.
  - Fix: none.

### Runtime and CI evidence — PASS

- **[PASS] Complete local integration and recovery gates pass** —
  `Makefile`, `.github/workflows/go-ci.yml`
  - What: the authoritative package selector, Go build/vet/race/coverage,
    repository consistency, shell/YAML parsing, Rust format/clippy/build/tests,
    active web-client lint/tests/build/high-severity audit, Cargo audit, and the
    complete eight-scenario multi-validator recovery matrix passed.
  - Path: the recovery matrix completed in 1161.044s and covered legacy
    authority migration rollback, consensus recovery, trusted snapshot state
    sync, backup/restore/export/import, persisted binary upgrade rollback,
    validator cold failover, consensus key rotation, and slashing.
  - Fix: none.

- **[PASS] Real native start/restart output is JSON** —
  `server_lifecycle_test.go:131`, `server_lifecycle_test.go:280`
  - What: a compiled daemon initializes, starts, commits, stops, restarts, and
    advances height while every captured non-empty stdout/stderr line parses as
    JSON with `level` and `message`.
  - Path: explicit plain, environment plain, and raw trace-store attempts fail
    before state opens.
  - Fix: none.

- **[PASS] Reviewed operator paths declare JSON explicitly** —
  `scripts/start-node.sh`, `Dockerfile`, `docker-compose.yml`
  - What: the script, image, and local Compose profile do not rely only on an
    implicit daemon default.
  - Path: repository policy tests fail if those defaults regress.
  - Fix: none.

- **[PASS] GitHub validates post-restart container logs** —
  `.github/workflows/go-ci.yml:130`
  - What: both container streams since the restart boundary must be non-empty
    JSON Lines with string `level` and `message`.
  - Path: malformed or unstructured normal-operation output fails the Docker
    job.
  - Fix: none.

- **[PASS] Exact container and Compose runtime evidence is green** —
  `.github/workflows/go-ci.yml`
  - What: exact head `fd8378c` built the image, started and restarted a
    persistent node, advanced height, validated both log streams as structured
    JSON Lines, and passed the full loopback Compose stack in 6m57s.
  - Path: the GitHub runner disproves image, entrypoint, restart, and
    Compose-only regressions that could not be tested without local Docker.
  - Fix: none.

### Operator boundary — PASS

- **[PASS] Documentation avoids production and DLP overclaim** —
  `docs/node-operators/operations/monitoring.md`
  - What: the guide distinguishes logger events from Go crash output, rejects
    raw KV tracing, treats redaction as defense in depth, and leaves retention,
    access control, encryption, collector deployment, and incident policy open.
  - Path: local Compose rotation cannot be mistaken for production retention or
    rollout approval.
  - Fix: none.

## Priority matrix

### 🔴 BLOCKING

None.

### 🟠 HIGH

None.

### 🟡 MEDIUM

None.

### 🟢 LOW

1. Redaction remains defensive minimization, not general DLP; continue to forbid
   secrets at the logging call site and treat retained logs as sensitive.
