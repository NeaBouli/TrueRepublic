# GH-60 Audit — Inactive Validator Genesis Recovery

Date: 2026-07-26
Scope: GH-60 local merge gate on `feature/GH-60-inactive-validator-genesis`

## Result

- P0: 0
- P1: 0
- P2: 0
- Local merge gate: PASS

Kimi K3 implemented the bounded truedemocracy genesis core and then performed
an independent read-only review. That review found one P2 fail-closed gap:
an explicit inactive claim could be unjailed with positive stored power when
its domain list was empty. Sol tightened the invariant so every inactive claim
with positive power must be jailed and added the corresponding regression.
The refreshed review and complete final verification found no remaining P0,
P1, or P2 finding.

## Verified invariants

- New exports explicitly classify every validator active or inactive while
  preserving the complete domain list and stored power.
- Valid pre-GH-60 records remain compatible and retain their former single-
  domain, stake-derived-power semantics.
- Active claims require positive exact stake-derived power, no jail state, and
  membership in every listed domain.
- Inactive retained claims preserve positive stake custody even when jailed,
  excluded from membership, or below minimum stake; they emit no initial
  consensus power.
- Domain, stake, pubkey, jail, jail-until, missed-block, and power state import
  without drift.
- Revoked keys cannot be reused by active or inactive claims. Existing pending
  rotation, signing, infraction, consensus-history, and pending-exit relations
  remain fail closed.
- `GenesisEscrowClaims` continues to count every retained active/inactive stake,
  and the exported bank ledger remains exactly backed within the 21M cap.

## Fresh local evidence

- Focused active/inactive, legacy, round-trip, and malformed-resurrection tests:
  PASS.
- `TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test . -run
  '^TestMultiValidatorConsensusSlashing$' -count=1 -v`: PASS in 441.52s after
  increasing only the 100-block liveness phase timeout for loaded CI hosts.
  Four real validators prove downtime and double-sign penalties, export two
  inactive claims, validate ledger parity, and re-import both without drift.
- `./scripts/test-go-packages.sh`: PASS.
- `CGO_ENABLED=1 ./scripts/go-packages.sh go build`: PASS.
- `./scripts/go-packages.sh go vet`: PASS.
- `CGO_ENABLED=1 ./scripts/go-packages.sh go test -race -cover -count=1
  -timeout=600s`: PASS; no race finding, 68.5% root and 62.2%
  `x/truedemocracy` statement coverage.
- `./scripts/check-consistency.sh`: PASS at 733 cases and 21,000,000 PNYX.
- `git diff --check`: PASS.

## Remaining gate

This is a local merge gate only. GH-60 remains open until the branch is
published, final-head GitHub checks and review are green, the PR is merged, and
GH-29 plus the Bridge are synchronized. No production rollout approval is
claimed.

## PR #67 security-gate remediation

The first final-head Security Scan exposed newly published High advisories in
`brace-expansion <=5.0.7` and `postcss <=8.5.17`; npm's retiring quick endpoint
reported them as an HTTP 400 invalid-tree error. The maintained client now uses
PostCSS `^8.5.23`, patched brace-expansion `5.0.8`, and a compatible unified
minimatch `10.2.5` path. Fresh install, lint, eight tests, production build, and
the unchanged `npm audit --audit-level=high` gate pass locally. Two Moderate
React Router advisories remain explicit and require a separately reviewed
breaking v7 migration.
