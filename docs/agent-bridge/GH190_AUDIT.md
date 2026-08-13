# GH-190 Maintained-client IBC Transfer and Recovery Audit

Status: locally verified candidate; protected publication remains pending.

## Implemented boundary

- Canonical upstream ICS-20 `MsgTransfer` is explicitly registered and pinned
  by a golden wire vector. The sender always comes from the unlocked signer.
- Channel discovery distinguishes authoritative empty results from typed
  transport, timeout, protocol, and decode failures and signs only over open,
  unordered `transfer`/`ics20-1` channels with one connection hop.
- Amount, native denom, fresh source balance plus a 10,000 upnyx reserve,
  receiver syntax, client-state height, and bounded height/timestamp timeouts
  fail closed before signing. The prior native Send invalid-to-zero coercion is
  removed as part of the shared prerequisite.
- Wallet-and-chain-scoped recovery records persist only public transaction and
  packet metadata. Lock, switch, delete, and stale completion invalidate active
  session state; passwords, mnemonics, signers, clients, and keys are excluded.
- Broadcast is never delivery. Exact send_packet evidence yields only pending
  relay; manual transaction-hash recovery uses source-chain ACK/timeout events,
  prioritizes canonical `packet_ack_hex`, reduces malformed or contradictory
  evidence to unknown, and never resubmits.

## Verification

- Maintained client: lint; 10 Node policy/budget cases; 249 Vitest cases;
  TypeScript production build; 20 lazy route entries; 70.05 kB gzip entry,
  4.91 kB maximum lazy route, 353.53 kB total JavaScript; no high/critical npm
  advisory.
- Disposable local chain: three client delivery cases pass, including a real
  canonical MsgTransfer that reaches the IBC router and is rejected for an
  absent channel instead of being fabricated as success.
- Repository Go build, package selector, vet, Race/Coverage pass. The standard
  baseline is 1,746 cases: 1,461 Go + 26 Rust + 259 maintained-client.
- `make ibc-two-chain` passes transfer/ACK/timeout/replay/recovery,
  close/timeout/replacement, and compatible binary-restart scenarios.
- Security contract, pinned Staticcheck v0.7.0, exact-policy Govulncheck v1.6.0
  and fixtures, Gitleaks v8.30.1, custom-query/mobile/web retirement contracts,
  JSON/docs consistency, and diff whitespace pass.
- Rust workspace formatting, strict Clippy, 26 tests, and cargo audit pass with
  the existing allowed warning set.

## Delegation and review

Kimi K3 supplied a large bounded core implementation covering types, canonical
registry, strict network channel parsing, transfer construction, send_packet
extraction, lifecycle reduction, and extensive tests. Sol reviewed and changed
that diff, implemented persistence/UI/source-event reconciliation, ran every
integration and security gate, and retains publication and merge ownership.
Claude Code supplied only the requested small read-only route/bundle inventory.
A later Kimi review session did not produce a completed report; Sol's final
review added canonical `packet_ack_hex` decoding. No known P0/P1/P2 remains.

## Residual boundary

This is repository and disposable local-chain evidence. It does not qualify or
operate an external relayer, public counterparty, real RPC, real key/account/
fund, IBC client upgrade, production deployment, release, or rollout.
