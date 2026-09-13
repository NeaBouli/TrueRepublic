# Independent Security Review Guide

> Schema: `truerepublic.security-review-scope/v1` and
> `truerepublic.security-review-findings/v1`  
> Updated: 2026-09-13  
> Status: repository-prepared review input; **not an independent audit**  
> `external_independence_claim=false` · `production_ready=false`

This guide is the canonical entry point for a future independent security
review of the maintained TrueRepublic recovery branch. It binds the intended
surface, critical invariants, existing evidence and finding lifecycle to files
that the repository can verify. It does not select an auditor, certify the
reviewer's independence, close Phase 5, authorize a release or approve
production.

#### Canonical inputs

| Input | Purpose |
|---|---|
| [`review-scope.json`](../../configs/security/review-scope.json) | Versioned modules, invariants, threat IDs and evidence index |
| [`review-findings.json`](../../configs/security/review-findings.json) | Current internal limitations and future internal/external findings |
| [`threat-model.json`](../../configs/security/threat-model.json) | Canonical threat register and residual-risk state |
| [`gates.json`](../../configs/security/gates.json) | Exact maintained security tools, pins and bounded exceptions |
| [`SECURITY.md`](../../SECURITY.md) | Private reporting, scanner and exception policy |
| [`ROLLOUT_ROADMAP.md`](../ROLLOUT_ROADMAP.md) | Phase gates and public production boundary |

The review manifests reference these sources; they do not copy or replace
their policy. If a referenced path, threat ID, invariant anchor or evidence
anchor disappears, the repository verifier fails closed.

The verifier must run against a clean, immutable checkout of the exact commit
under review, with no concurrent process changing files. Its traversal and
open-time identity checks reject symlinks and detectable replacements, but it
is not a sandbox for a hostile, concurrently mutable filesystem. The reviewer
records the full commit SHA and establishes that frozen-checkout precondition
before relying on the result.

## Review surface

The manifest covers application wiring and genesis, PNYX issuance and supply,
governance and treasury custody, DEX custody and LP accounting, migrations,
network policy, observability, the maintained `client-web`, and CosmWasm
contracts. The minimum invariant set is:

1. the 21,000,000 PNYX canonical supply cap;
2. exact governance treasury/stake escrow parity;
3. DEX bank custody and LP-share conservation;
4. one-time proposal-scoped anonymous-vote nullifiers;
5. genesis-pinned Groth16 verification-key enforcement.

An independent reviewer may expand this scope. Narrowing it requires a
separately reviewed manifest change and must not be used to hide an existing
finding.

## Findings lifecycle

The register accepts `open`, `remediated`, `verified_closed`, and
`accepted_risk` states. Severity is one of `critical`, `high`, `medium`, or
`low`; source is `internal` or `external`. Every finding carries one or more
unique `threat_ids` that must resolve in the canonical threat model.

- `remediated` and `verified_closed` require repository-verifiable remediation
  evidence containing the finding ID.
- `verified_closed` additionally requires `verification_source=external` and
  separate repository-verifiable external-review evidence containing the
  finding ID. The finding's `source` records where it originated and may remain
  `internal`; it cannot substitute for independent closure verification.
- `accepted_risk` requires a named accountable owner and an expiry after the
  acceptance date but no more than 30 days later; expired acceptance fails the
  verifier.
- Accepted critical/high risk remains unresolved for the Phase-5 exit gate;
  risk acceptance cannot substitute for the required remediation.
- Non-accepted findings cannot carry acceptance dates.
- Both manifests must explicitly keep independence and production claims
  false. Repository contributors cannot turn this preparation package into an
  independent-review claim by editing a boolean.

The initial register has no external findings. Its internal open entries make
the already documented ZKP, client trust/custody, live operations, release
provenance and independent-review gaps visible in one machine-checked place.

## Reviewer workflow

1. Start from the exact protected commit under review and record its full Git
   SHA outside this repository.
2. Read the threat model, scope manifest, findings register, security policy,
   rollout roadmap and historical audit notice before inspecting code.
3. Verify the manifests and repository bindings. The Make target also composes
   the complete threat-model and security-gate repository contracts rather
   than relying only on their version fields:

   ```bash
   make security-review-contract-test
   go run ./cmd/security-review \
     --repo-root . \
     --scope configs/security/review-scope.json \
     --findings configs/security/review-findings.json
   ```

4. Run the complete exact-version gates documented in `SECURITY.md`, followed
   by `make verify`, client and contract checks applicable to the final review
   commit.
5. Report sensitive findings through the private process in `SECURITY.md`.
   Record only sanitized metadata in the public register.
6. Re-test remediation on the exact amended commit. Keep a finding open until
   the independent reviewer verifies the fix.

## Exit boundary

GH-294 is complete when this package and its fail-closed repository contract
are merged and green. The Phase-5 rollout item remains open until a real
independent review covers the agreed surface and every critical/high finding
is resolved with evidence. A future review also does not by itself authorize
tagging, signing, artifact publication, deployment, real keys/funds, mainnet or
production; those remain separate Phase-7 and go/no-go gates.
