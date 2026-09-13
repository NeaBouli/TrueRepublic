# TrueRepublic GH-294 — Audit
> Scope: `securityreview/`, `cmd/security-review/`, review manifests, repository contracts, CI wiring, and review documentation  ·  Date: 2026-09-13  ·  Result: 0 FAIL / 2 WARN / 5 PASS

## Summary

GH-294 provides a strict, versioned independent-review readiness package and
keeps both reviewer-independence and production-readiness claims explicitly
false. An independent read-only diff review initially found four lifecycle,
threat-binding, contract-composition, and filesystem-race gaps; all four were
addressed before this audit was recorded. The block is suitable for protected
CI and maintainer review, but it is not an independent security audit and
grants no rollout credit. The remaining warnings concern the immutable-checkout
operating precondition and a repo-wide race run that must complete in protected
CI because the local host had insufficient free disk.

## Findings by domain

### Schema and parser boundary — PASS

- **[PASS] Strict bounded input contract** — `securityreview/parse.go`
  - What: JSON documents are size-bounded, reject unknown fields and trailing
    values, and require explicit safety claims.
  - Path: malformed, oversized, extended, or claim-promoting documents fail
    before repository evidence is accepted.
  - Fix: none.

### Finding lifecycle — PASS

- **[PASS] External closure cannot be represented as internal-only evidence** — `securityreview/verify.go`
  - What: `verified_closed` requires remediation evidence plus separate
    `verification_source=external` evidence containing the finding ID.
  - Path: an internally originated finding may be remediated internally, but
    cannot reach verified closure without the external verification fields.
  - Fix: none.

### Threat and gate composition — PASS

- **[PASS] Findings remain bound to canonical threats and full contracts** — `securityreview/verify.go`, `Makefile`
  - What: every finding requires unique known threat IDs, and the readiness
    target runs the complete threat-model and security-gate repository tests.
  - Path: deleted/unknown threat links or deeper gate/register drift fail the
    composed Make target.
  - Fix: none.

### Repository path safety — WARN

- **[MEDIUM] Verification assumes a frozen checkout** — `securityreview/parse.go`, `docs/security/INDEPENDENT_REVIEW_GUIDE.md`
  - What: traversal, symlinks, non-regular files, oversize files and detectable
    check/open replacements are rejected, but a hostile concurrent mutation of
    ancestor directories is outside the portable verifier's trust boundary.
  - Path: a concurrent privileged writer could change repository topology while
    a review command is running.
  - Fix: run only on the documented clean, immutable exact-commit checkout and
    record its full SHA; use an isolated filesystem for an adversarial source.

### Verification environment — WARN

- **[MEDIUM] Local repo-wide race rerun is deferred to protected CI** — `.github/workflows/ci.yml`
  - What: focused race tests, full serial Go tests, static analysis, vulnerability,
    secret and license gates passed locally; the repo-wide race/link step was
    interrupted by local disk exhaustion rather than a test assertion.
  - Path: no project test failure was observed, but the complete race matrix must
    finish on the protected PR commit before merge.
  - Fix: require the protected GitHub workflow to complete successfully.

### Documentation boundary — PASS

- **[PASS] No independence or production claim** — `docs/security/INDEPENDENT_REVIEW_GUIDE.md`, `CODEX_AUDIT.md`
  - What: current documentation labels the package as repository-prepared input
    and the older audit as historical.
  - Path: readers are directed to the machine-checked current status without a
    certification claim.
  - Fix: none.

### CI and consistency wiring — PASS

- **[PASS] Contract runs in maintained security CI** — `.github/workflows/security-scan.yml`, `scripts/check-consistency.sh`
  - What: the review contract is part of the maintained security workflow and
    documentation consistency path.
  - Path: manifest, anchor, threat, gate or public-status drift blocks CI.
  - Fix: none.

## Priority matrix

### 🔴 BLOCKING

None.

### 🟠 HIGH

None.

### 🟡 MEDIUM

1. Preserve the clean, immutable exact-commit checkout precondition during any
   real independent review.
2. Require the repo-wide protected race workflow to finish before merge.

### 🟢 LOW

None.
