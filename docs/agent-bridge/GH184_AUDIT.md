# GH-184 Governed Upgrade Audit

## Scope

GH-184 implements the bounded fresh-genesis `v0.4.1` application-upgrade path.
It does not authorize production, public networks, real keys or funds,
pre-GH-184 store introduction, IBC client upgrades, or arbitrary release plans.

## Architecture and authority

- Cosmos SDK `x/upgrade` is mounted as a real module and IBC keeper dependency.
- Its authority is the non-signing `truedemocracy` module account.
- A reserved genesis-only `governance` Domain supplies an immutable sorted
  electorate snapshot. Matching schedule or cancellation votes require at
  least two thirds of that snapshot.
- The exact plan tuple is bound on the first vote. Conflicts, duplicates, late
  members, insufficient lead time, invalid metadata, and scheduler failures
  fail closed.
- The governance electorate cannot add or exclude members after genesis, and a
  pending plan plus schedule/cancel votes round-trip through genesis export and
  deterministically restore the real upgrade plan.
- Only an artifact with exact `main.upgradePlan=v0.4.1` registers the reviewed handler; `main.version` remains the source commit.

## Recovery evidence

`TestGovernedUpgradeMultiValidatorHaltFailureRecovery` builds three distinct
artifacts and runs four real validators. Three of four governance voters
schedule the plan. The baseline artifact halts at the due height. A test-only
handler performs a cached write and returns an error; no state is committed.
The fixed candidate then migrates, resumes common-height app hashes with the
same validator powers, and survives another restart without replaying the
migration.

Keeper tests additionally cover the governance-domain boundary, plan
validation, threshold, immutable electorate, duplicate and conflicting votes,
schedule rollback/retry, cancellation authorization, clear rollback/retry,
pending-plan export/import, stale proposal cleanup, post-execution cleanup,
and rejection of malformed upgrade genesis.

## Delegated review

Kimi K3 received secret-free bounded architecture and implementation work for
the `x/truedemocracy` adapter, then independently reviewed the final diff.
Sol reviewed every resulting write, remediated all P2/P3 observations, added
the adversarial tests and real process harness, retained all security and
integration decisions, and owns the complete local and GitHub verification
chain. Final review reports no P0/P1/P2 finding.

## Residual risk

- Chains created before the upgrade store existed need a separate reviewed
  store-loader transition.
- Clean private-infrastructure reproduction remains a rollout exit gate.
- External relayers, public counterparties, IBC client upgrades, remaining
  staking/distribution stubs, production deployment, and release approval stay
  open.

## Verification

- Governed four-validator halt/failure/recovery/exact-once harness: PASS in
  231.39 seconds.
- Fresh standard-suite arithmetic: PASS, 1,624 cases (1,457 Go, 26 Rust,
  141 maintained client); process gates are separate.
- Build/Vet/full Race/Coverage: PASS; root coverage 72.2%, governance 64.0%,
  DEX 51.1%, all above maintained floors.
- Property/fuzz, security contract, deterministic build contract, Staticcheck,
  exact-policy Govulncheck, Gitleaks positive/negative fixtures: PASS.
- Client fresh install/lint/test/build/bundle/audit and Rust fmt/clippy/build/
  26 tests/audit: PASS. Rust audit retains five allowed upstream warnings.
- Documentation consistency, JSON and diff integrity: PASS.
- Kimi final read-only review: no P0/P1/P2; all actionable P3 notes resolved.
- Protected PR #185 exact head `9c4f58c` passed all 21 reported contexts with
  zero review threads and merged as `bf32ad5`. Final-main Docs, Security,
  reproducible Linux, Go, and Pages pass; live Pages confirms 1,624 cases,
  rollout 28/59, Phase 6 6/7, production false, and the GH-184 status. GH-29
  is synchronized.
