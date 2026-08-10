# GH-175 IBC Two-Chain Packet and Recovery Audit

Date: 2026-08-10
Baseline: `origin/main` `10629f1`
Scope: local IBC application integration, deterministic proof relay, CI, and
status documentation

## Verdict

Local implementation is approved for protected pull-request verification. The
test exposed and fixed a production integration defect: the concrete
Tendermint light-client protobuf types were not registered in the application
transaction codec, so `MsgCreateClient` decoded as Cosmos code 2. A standard
regression test now pins that registration, and the real two-chain gate passes.
Kimi's independent review found one P2 evidence-quality defect: the original
duplicate receive stopped at a stale proof before the receipt replay check.
Sol remediated it by refreshing the client and proving ibc-go's real code-0
receive no-op without a second mint. No P0/P1 or unresolved P2 finding remains.

## Evidence reviewed

- Two distinct persistent TrueRepublic application databases use fixed,
  disposable validators, bank-backed PoD genesis, and supplies below the 21M
  PNYX cap.
- Real ibc-go Tendermint clients, connection handshakes, and an unordered
  ICS-20 channel open with committed IAVL state proofs.
- Native `upnyx` leaves the sender, enters the canonical channel escrow, and
  mints exactly one destination voucher with the expected denom trace.
- Receive writes one receipt and acknowledgement; acknowledgement relay clears
  the source commitment without refunding a successful transfer.
- With a fresh counterparty client proof, duplicate receive reaches ibc-go's
  receipt replay check and succeeds as a deliberate no-op without a second
  mint. Duplicate acknowledgement and timeout use the same idempotent contract
  and have no second economic effect.
- An expired unreceived packet refunds the sender exactly once and clears its
  commitment.
- The destination database is closed and reopened while an acknowledgement is
  pending, commits a fresh post-restart header, verifies that header through the
  counterparty client, and completes the acknowledgement without state drift.
- Both application crisis invariant sets pass after the packet scenarios.

## Verification

- Dedicated remediated `make ibc-two-chain`: PASS in 1.17s locally.
- Focused IBC standard tests: PASS; the opt-in harness skips outside its gate.
- Standard Go arithmetic: 1,441 passing cases; the opt-in harness is excluded.
- Documentation consistency, JSON, workflow YAML, formatting, and diff
  integrity: PASS.
- Full build, Vet, Race/Coverage (root 71.9%), Staticcheck, security contract,
  Gitleaks, and exact active-policy Go vulnerability gates: PASS locally.
- Protected exact-head workflow results remain required before merge and Done.

## Residual boundary

The harness is a deterministic proof-relay test, not an external Hermes/Go
Relayer process or public counterparty qualification. It does not prove channel
closure/replacement, timeout-on-close, IBC behavior across an application
upgrade, governance-controlled client upgrades, production topology, real
keys/funds, deployment, mainnet readiness, or independent external audit.
`IBCStakingKeeper`, `IBCUpgradeKeeper`, and the CosmWasm staking/distribution
compatibility surfaces remain explicitly limited. Production readiness stays
false.
