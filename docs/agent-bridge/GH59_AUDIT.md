# GH-59 Audit — ABCI++ Validator Slashing

Date: 2026-07-24
Scope: GH-59 local merge gate on `feature/GH-59-abci-slashing`

## Result

- P0: 0
- P1: 0
- P2: 0
- Merge gate: PASS

The final independent read-only review found no remaining security or consensus
finding in the GH-59 scope. Earlier review findings were remediated before this
gate, including stale signing state after payout, incomplete pending-exit
genesis accounting, unjail cursor continuity, and incomplete consensus-key
history relations for revocations and pending rotations.

## Verified invariants

- ABCI++ evidence and decided-last-commit signals are deterministic,
  canonicalized, and replay-safe.
- Equivocation is burned once per tombstoned consensus key and delayed evidence
  remains attributable across key rotation.
- Downtime requires a complete rolling liveness window and cannot be reset by
  replay or stale signing state.
- Validator exit stake remains slashable through the CometBFT evidence window;
  release requires both height and time limits to be exceeded.
- Genesis import/export validates active, retired, revoked, rotated, jailed,
  pending-exit, signing, and processed-infraction cross-relations fail closed.
- Mature exit payout removes the corresponding signing state atomically.

## Fresh local evidence

- `git diff --check`: PASS
- `./scripts/check-consistency.sh`: PASS at 726 tests and 21,000,000 PNYX
- `./scripts/test-go-packages.sh`: PASS
- `CGO_ENABLED=1 ./scripts/go-packages.sh go test -count=1 -timeout=600s`:
  PASS for all five repository packages; root/process package 158.108s and
  `x/truedemocracy` 63.225s
- `CGO_ENABLED=1 ./scripts/go-packages.sh go build`: PASS
- `./scripts/go-packages.sh go vet`: PASS
- `CGO_ENABLED=1 ./scripts/go-packages.sh go test -race -cover -count=1
  -timeout=600s`: PASS; no race finding, with 68.5% root and 61.8%
  `x/truedemocracy` statement coverage

GitHub final-head CI remains the publication gate. This audit does not claim a
production rollout approval; the remaining rollout program continues in GH-29.
