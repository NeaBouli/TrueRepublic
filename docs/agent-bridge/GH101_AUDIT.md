# TrueRepublic GH-101 — Audit
> Scope: `deploymentevidence/`, maintained synthetic fixture, root/CI wiring, repository tests, operator documentation, and GH-101 bridge state · Date: 2026-08-04 · Result: 0 FAIL / 0 WARN / 7 PASS

## Summary

The repository-only private-topology evidence envelope is ready to ship. It
strictly parses a bounded secret-free manifest, binds it to the exact raw bytes
of a separately supplied and already valid GH-89 topology contract, requires
the canonical eleven gates and logical two-seat approval separation, and emits
only a fixed secret-free report. Sol reviewed and hardened Kimi's bounded
implementation, the full Go Race/Coverage suite and polyglot checks pass, and
Kimi's independent final review reports no P0/P1/P2 finding. This verifier does
not inspect source artifacts, authenticate people, probe infrastructure, or
prove a deployment; Phase 6 therefore remains 6/7 and production readiness
remains false.

## Findings by domain

### Strict parsing and no-reflection — PASS

- **[PASS] Untrusted manifests fail closed without reflecting input** — `deploymentevidence/parse.go`, `deploymentevidence/deploymentevidence_test.go`
  - Documents are size/depth bounded and reject empty, malformed, duplicate,
    trailing, scalar, array, and unknown-field input. Parser/schema/flag/topology
    errors collapse to fixed categories, and planted secrets never appear in
    errors or reports.

### Topology binding — PASS

- **[PASS] Verification derives trusted facts from exact validated contract bytes** — `deploymentevidence/load.go`, `deploymentevidence/verify.go`
  - SHA-256, chain ID, node count, and role counts come from the supplied GH-89
    contract after its existing validation. Report counts use canonical policy
    and derived topology values rather than rejected manifest declarations.

### Gate and time policy — PASS

- **[PASS] Canonical gates, digests, order, freshness, and skew are enforced** — `deploymentevidence/types.go`, `deploymentevidence/verify.go`
  - Exactly eleven ordered passing gates with unique lowercase SHA-256 digests
    are required. Strict UTC timestamps enforce start/completion/preparation
    order, a 30-day maximum age, and five-minute future skew.

### Approval separation — PASS

- **[PASS] Logical preparation and approval seats are structurally separated** — `deploymentevidence/verify.go`
  - One preparer plus distinct operator and independent-reviewer seats are
    required; approvals cannot reuse the preparer seat and bind the same exact
    topology digest. Documentation correctly states that seats do not prove
    distinct humans.

### CLI and CI integration — PASS

- **[PASS] The command is offline, config-independent, and deterministic in CI** — `deploymentevidence/command.go`, `server_lifecycle.go`, `.github/workflows/go-ci.yml`
  - The command reads only two explicit local paths, never initializes a node
    home or performs network access, and CI pins the synthetic evaluation time
    while asserting exact valid/count/empty-violation semantics.

### Regression coverage — PASS

- **[PASS] Positive, negative, repository, command, and race tests cover the boundary** — `deploymentevidence/deploymentevidence_test.go`, `deployment_evidence_repository_test.go`
  - The package reaches 90.8% Race/Coverage. Repository tests pin fixture
    binding, exclude secret/inventory markers, and prevent silent CI or
    documentation-boundary removal.

### Documentation and rollout claims — PASS

- **[PASS] Guidance does not claim real infrastructure or rollout evidence** — `docs/node-operators/operations/deployment-evidence.md`, `docs/LIMITATIONS.md`
  - Private storage, evidence collection, two-person review, abort/rollback,
    retention, and synthetic-fixture boundaries are explicit. The verifier is
    described as envelope validation only; Phase 6 remains open.

## Priority matrix

### 🔴 BLOCKING

None inside GH-101.

### 🟠 HIGH

None inside GH-101.

### 🟡 MEDIUM

None inside GH-101.

### 🟢 LOW

None inside GH-101.

## Residual project risks

- Real private topology, source-artifact inspection, human identity checks,
  provider/firewall/DNS/TLS state, and deployment remain separately authorized
  rollout work.
- A repository-wide non-GH-101 check found current dependency advisories in the
  maintained web and mobile clients. GH-102 tracks their separate remediation
  because GH-101 changes no client file or client workflow.
