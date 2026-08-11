# GH-178 IBC Channel-Recovery Audit

Date: 2026-08-12
Status: LOCAL PASS / PROTECTED PR GATES PENDING
Risk: High — IBC packet commitments, bank escrow, refunds, channel state, and
persistent proof verification

## Scope

- `ibc_two_chain_harness_test.go`
- `Makefile`
- `.github/workflows/go-ci.yml`
- IBC/security/rollout/status documentation and Bridge state

## Verified behavior

- [PASS] Two persistent TrueRepublic applications reuse their established
  Tendermint clients and connection for a distinct replacement ICS-20 channel.
- [PASS] ICS-20's real `MsgChannelCloseInit` path rejects user-initiated close;
  the committed CLOSED counterparty fixture matches ibc-go's own
  `TimeoutOnClose` test boundary and survives database reopen.
- [PASS] A real `MsgChannelCloseConfirm` verifies the recovered counterparty
  channel proof and closes the source end.
- [PASS] A real `MsgTimeoutOnClose` verifies both CLOSED-channel membership and
  unordered packet-receipt absence, refunds native escrow, and deletes the
  packet commitment exactly once.
- [PASS] Duplicate timeout-on-close is a Cosmos success/no-op with unchanged
  balance and no second refund. The closed old channel rejects a new transfer
  specifically with `channel is not OPEN`.
- [PASS] The replacement channel has distinct channel IDs, escrow address, and
  voucher denom; its transfer and acknowledgement complete while both old ends
  remain CLOSED and both applications pass crisis invariants.

## Independent review

Kimi K3 independently checked the ibc-go v8.7 close, proof, restart,
timeout-on-close, replay, and replacement APIs. It confirmed that ICS-20
intentionally rejects close-init and that ibc-go itself uses a committed CLOSED
fixture for timeout-on-close. No P0/P1/P2 finding surfaced. Its one actionable
P3 observation was that the old-channel negative case should assert the exact
failure reason; Sol tightened it to `channel is not OPEN` and reran the gate.

## Local validation

- `make ibc-two-chain` → PASS; both GH-175/GH-178 scenarios pass.
- `make build && make verify` → PASS; build, package tests, Vet, and Race/Coverage
  pass (root 71.9%, DEX 51.1%, governance 63.8%).
- `make quality-depth && make security-contract` → PASS; deterministic quality,
  live fuzz, and security repository contract pass.
- Pinned Staticcheck, Gitleaks, Govuln policy/fixtures → PASS; no secret found and
  the exact active no-fix vulnerability policy passes.
- Maintained client: `npm ci`, lint, 141 tests, production build/budget, and audit
  → PASS; 0 vulnerabilities.
- Contracts: format, strict Clippy, 26 tests, and audit → PASS with the five
  repository-accepted warnings and no vulnerability failure.
- `./scripts/check-consistency.sh` and `git diff --check` → PASS.
- Status arithmetic → 1,609 = 1,442 Go + 26 Rust + 141 client; rollout stays
  25/59 and `production_ready=false` because post-upgrade recovery, external
  relayer qualification, and other rollout gates remain open.

## Residual boundaries

- The initial CLOSED counterparty end is a committed deterministic fixture, not
  a production close authority or public-chain operation.
- External relayer/counterparty qualification, IBC application-upgrade
  recovery, governance upgrade support, staking/distribution/upgrade stubs,
  deployment, real keys/funds, and release approval are not proven here.
- Protected exact-head CI, review-thread closure, merge, final-main checks, and
  live Pages readback remain required before GH-178 is Done.

## Verdict

LOCAL PASS. The branch is ready for protected review; production remains false.
