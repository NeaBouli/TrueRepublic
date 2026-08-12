# TrueRepublic GH-181 — Audit
> Scope: compatible IBC binary-restart harness, gate wiring, and public evidence · Date: 2026-08-12 · Result: 0 FAIL / 1 WARN / 8 PASS

## Summary

GH-181 is locally ready for protected publication. A separately linked package
test binary reopens two closed LevelDB application states in place, retains the
real ibc-go client/connection/channel and packet/economic state, and completes
pending ACK, fresh ACK, and timeout/refund paths with explicit duplicate no-op
checks. The largest residual limitation is intentional: baseline and candidate
use the same source tree, so this qualifies only a compatible binary restart,
not `x/upgrade`, a state migration, or arbitrary cross-version compatibility.

## Findings by domain

### State persistence and recovery — PASS

- **[PASS] In-place reopen preserves height and avoids genesis reset** —
  `ibc_two_chain_harness_test.go:705`
  - What: Both candidate applications reopen the original LevelDB directories.
  - Path: The candidate constructs a new app around each closed database and
    asserts the exact prior height before any packet continuation.
  - Fix: None.

- **[PASS] Fail-before-open rollback is byte-protected** —
  `ibc_two_chain_harness_test.go:415`
  - What: The exact candidate exits with code 42 before reading the manifest or
    opening state.
  - Path: Sorted relative paths and complete file bytes from both databases are
    hashed before and after the failed candidate; any mutation fails the test.
  - Fix: None.

### IBC protocol and economics — PASS

- **[PASS] Core IBC state is preserved explicitly** —
  `ibc_two_chain_harness_test.go:489`
  - What: Packet commitment, receipt, ACK commitment, Tendermint clients, OPEN
    connections, unordered OPEN channel, send sequence, escrow and balances are
    asserted after candidate reopen.
  - Path: The manifest carries only deterministic test metadata; canonical
    state is reread from the reopened keepers before relay resumes.
  - Fix: None.

- **[PASS] Pending and fresh acknowledgements are exactly once** —
  `ibc_two_chain_harness_test.go:523`
  - What: The pending ACK is completed after reopen and a fresh transfer/ACK
    follows on the same channel.
  - Path: Commitments are removed, duplicate ACK relays succeed as ibc-go no-ops,
    and source/escrow/receiver balances remain unchanged after each duplicate.
  - Fix: None.

- **[PASS] Timeout refund is exactly once** —
  `ibc_two_chain_harness_test.go:546`
  - What: An unreceived post-restart packet times out against a fresh verified
    consensus state.
  - Path: The source and escrow return to their exact pre-send balances; a
    duplicate timeout cannot create another economic effect.
  - Fix: None.

### Provenance and subprocess safety — PASS

- **[PASS] Candidate provenance is asserted inside the process** —
  `ibc_two_chain_harness_test.go:408`
  - What: `go test -c` links a separate artifact with a fixed package version
    marker.
  - Path: The candidate phase refuses to open state unless its runtime marker is
    exactly `gh181-compatible-candidate`.
  - Fix: None.

- **[PASS] Child execution is bounded and shell-free** —
  `ibc_two_chain_harness_test.go:770`
  - What: Child phases use argument arrays, exact test selection, context
    cancellation, and a two-minute child timeout.
  - Path: No manifest path or phase string is interpreted by a shell; unexpected
    exit codes and child output fail the parent with diagnostics.
  - Fix: None.

### CI and documentation — PASS

- **[PASS] The bounded proof is protected by CI and honest public wording** —
  `Makefile:27`, `.github/workflows/go-ci.yml:102`, `docs/status.json`
  - What: The opt-in gate runs all three two-chain scenarios and public status
    records the compatible-restart evidence without broadening production claims.
  - Path: Documentation consistency enforces 1,610 = 1,443 Go + 26 Rust + 141
    client and rollout 26/59 while `production_ready=false` remains unchanged.
  - Fix: None.

### Compatibility boundary — WARN

- **[LOW] Same-source candidate does not qualify future migrations** —
  `ibc_two_chain_harness_test.go:389`
  - What: Only the link-time version marker differs between baseline and
    candidate; application source and stores are schema-compatible by design.
  - Path: A future consensus-breaking store migration could still invalidate
    clients or packets despite this green compatible-restart harness.
  - Fix: Keep `x/upgrade`, governed halt/resume, migration rollback, IBC client
    upgrades, external relayers and arbitrary version qualification as separate
    open rollout gates.

## Priority matrix

### 🔴 BLOCKING

None.

### 🟠 HIGH

None.

### 🟡 MEDIUM

None.

### 🟢 LOW

1. Preserve the same-source limitation until a separately scoped governed
   migration design and version-pair test matrix exist.

## Verification evidence

- `make ibc-two-chain` — PASS; GH-175, GH-178 and GH-181 scenarios pass.
- `make build && make verify` — PASS; build, package selection, build, Vet,
  Race and Coverage pass (root coverage 71.9%).
- `make quality-depth` and `make security-contract` — PASS.
- Pinned Staticcheck, Gitleaks and Govuln gates — PASS; no secrets and active
  no-fix vulnerability policy satisfied.
- Maintained client — PASS: clean install, lint, 141 tests, production build,
  bundle budget and high/critical audit.
- Rust workspace — PASS: format, strict Clippy, 26 tests and audit; five pinned
  allowed transitive warnings, no blocking vulnerability.
- Documentation consistency, JSON parse and `git diff --check` — PASS.
- Independent Kimi review — no P0/P1/P2; both P3 hardening notes remediated and
  the focused/full relevant gates rerun afterward.
