# TrueRepublic GH-93 — Audit

> Scope: incident command, synthetic rehearsal contract, offline validator,
> CI/repository gates, and operator runbook integration · Date: 2026-08-01 ·
> Result: 0 open FAIL / 0 open WARN / 7 PASS

## Summary

GH-93 composes the existing recovery procedures into one repository-owned,
secret-free incident rehearsal system. The strict versioned contract covers
chain halt, validator failure and slashing, consensus-key and operator-authority
compromise, sanitized backup/restore, compatible binary upgrade/rollback, and
the separately bounded legacy-authority migration path.

The maintained fixture is synthetic and the validator is offline. Neither is
evidence of a live rehearsal, paging integration, production topology,
operator assignment, private evidence store, deployment, or rollout approval.

## Findings by domain

### Strict input and disclosure boundary — PASS

- **[🔴 BLOCKING, protected] Untrusted plans fail before rehearsal use.**
  - One JSON document is limited to 256 KiB and bounded nesting depth.
  - Duplicate and unknown fields, trailing data, unsupported values, unsafe
    actions, incomplete evidence, and ambiguous authority fail closed.
  - The schema has no field for hosts, endpoints, commands, credentials,
    identities, private keys, contacts, or production evidence values.
- **[🟠 HIGH, remediated] Role seats cannot hide secret-like free text.**
  - Review path: a merely pattern-safe alias such as `incident-secret-primary`
    could pass the first implementation.
  - Fix: each role now requires one exact synthetic seat alias. Real participant
    mapping remains in a private exercise record outside the repository.
- **[🟠 HIGH, remediated] Rejected values never reach public diagnostics.** —
  `incidentpolicy/parse.go`, `incidentpolicy/validate.go`,
  `incidentpolicy/command.go`
  - What: the pre-ship review found that duplicate JSON keys and unsupported
    enum values could be quoted in parser or validator errors.
  - Path: a secret pasted into an action, evidence, role, abort, JSON-key, or
    output-format field failed validation but could be echoed to stdout/stderr.
  - Fix: parse categories, rejected-value, positional-argument, flag, output,
    and command errors are generic. Report version, exercise ID, scenario IDs,
    and scenario count are fixed synthetic constants. A broad regression
    mutates every public field and proves the sentinel is absent.

### Incident authority and stop/go control — PASS

- Every scenario has exactly seven ordered phases: detect, contain, preserve,
  decide, recover, validate, and close.
- Severity, primary/secondary command, protocol, security, validator,
  evidence, communications, and release authority are explicit.
- Critical execution requires the declared approval set. Unsafe or incomplete
  paths end in `safe-stopped`, never in an optimistic success claim.

### Signer and validator safety — PASS

- **[🔴 BLOCKING, protected] Single-signer and stale-state recovery fail
  closed.**
  - All signer-relevant scenarios require a `second-signer-detected` abort.
  - Slashing, key compromise, validator failure, halt, upgrade, and legacy
    paths require their scenario-specific isolation and evidence boundaries.
- **[🟠 HIGH, remediated] Dual-signer coverage is complete.**
  - Review found the abort mandatory only for validator failure.
  - The validator, maintained fixture, and negative tests now cover every
    signer-relevant scenario.

### Upgrade and legacy rollback safety — PASS

- Compatible rollback is allowed only while candidate state remains unopened;
  install-before-rollback ordering is enforced and negatively tested.
- Legacy migration remains a separate fresh-genesis path. It requires source
  and target isolation, stops the target before rollback, and restarts only the
  untouched source chain. Concurrent source/target operation and state mixing
  are explicit abort conditions.

### CLI and CI integration — PASS

- `truerepublicd incident-rehearsal validate` is config-independent and does
  not initialize or read a node home.
- Machine-readable reports contain only version, exercise identifier, scenario
  count, validity, and generic violations; plan values are not reflected.
- The Go workflow triggers on `configs/incidents/**` and validates the real
  command's JSON result for eight scenarios and an empty violation array.

### Documentation consistency — PASS

- The canonical incident-command guide routes to the existing monitoring,
  slashing, identity recovery, key rotation, backup/restore, upgrade, and
  legacy migration procedures without weakening their boundaries.
- A stale key-rotation claim that contradicted the merged ABCI++ slashing
  implementation was corrected and pinned by a repository regression.
- Generic firewall/proxy snippets are explicitly marked incomplete and
  non-production; strict network and topology policy remain authoritative.

### Independent review and test isolation — PASS

- Kimi K3 was invoked for the bounded secret-free review but its provider quota
  returned HTTP 403 before producing a finding or diff.
- Terra's independent read-only reviews found the stale slashing statement,
  exact-seat gap, incomplete dual-signer coverage, and missing upgrade-order
  regression. All were fixed; final review reports 0 open P0/P1 findings.
- Spark contributed only focused package tests. Sol reviewed and hardened the
  delegated diff and owns the final integration.
- The first final full run exposed a pre-existing test-isolation weakness under
  simultaneous heavy local compilation: the lifecycle harness inherited the
  fixed pprof port 6060 and allowed only 35 seconds for first consensus. The
  focused test reproduced as healthy; its harness now disables pprof and uses
  a 60-second readiness budget. Focused race and the complete rerun pass.

## Verification boundary

- Focused incident-policy race/coverage: PASS, 90.7%.
- Root command, config independence, non-leak, repository contracts, JSON
  fixture, documentation consistency, and diff hygiene: PASS.
- Complete package selection, build, vet, race, and coverage: PASS. Coverage
  includes root/application 69.1%, incident policy 90.7%, network policy 95.5%,
  topology policy 85.8%, and governance 62.2%.
- Mandatory before merge: exact-head GitHub build/race/coverage, all eight
  recovery harnesses, Docker/Compose, docs/static/security, review, and
  unresolved-thread gates.
- Mandatory before any real rollout: separately authorized private operator
  mapping, real topology, live rehearsal, paging, credentials, evidence store,
  deployment, and production approval.

## Priority matrix

### 🔴 BLOCKING

- None open.

### 🟠 HIGH

- None open. Secret-like role-seat and rejected-value reflection paths were
  remediated and covered by negative tests.

### 🟡 MEDIUM

- None open. Dual-signer and upgrade-order review findings were remediated.

### 🟢 LOW

- None open.
