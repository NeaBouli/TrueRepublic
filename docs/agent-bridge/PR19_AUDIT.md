# PR #19 Audit — GH-12 Genesis and Runtime Invariants

Date: 2026-07-12
Branch: `fix/GH-12-genesis-invariants`
Base: `fix/GH-10-dex-custody`
Issue: [GH-12](https://github.com/NeaBouli/TrueRepublic/issues/12)
Result: PASS for the GH-12 ledger/genesis scope; project remains non-production

## Scope and guarantees

- Full app genesis validates canonical PNYX supply and both custom modules before
  any custom store mutation.
- Governance treasury/stake claims and DEX reserve claims must equal the complete
  corresponding module bank balances, including all denoms.
- DEX genesis validates registered assets, pools, unique provider positions, and
  total provider shares; export persists provider LP ownership.
- Non-empty InitChain/block/commit/export/re-import preserves canonical supply,
  governance escrow, DEX reserves, and LP positions.
- `x/crisis` checks PNYX cap, governance escrow, DEX reserves, and LP
  conservation every block.

## Audit findings fixed

### Critical — default consensus private key was publicly derivable

The prototype generated its production bootstrap validator with
`GenPrivKeyFromSecret` and a source-code constant. Anyone could derive the same
private key and sign or double-sign as that validator. Production defaults are
now empty; InitChain bootstraps only from actual positive-power Ed25519 public
keys supplied by CometBFT and adds exact minimum stake within the canonical cap.
All deterministic private keys now exist only in tests.

### High — runtime invariant claim exceeded its integration evidence

Only escrow divergence had been exercised through registered `x/crisis` routes.
Full-app tests now deliberately corrupt canonical supply, governance escrow,
DEX reserves, and pool total shares independently and prove every route halts.

### High — non-empty custody was not round-trip tested

The prior test exported only the default/bootstrap ledger. A new full-app test
starts with bank-backed treasury/stake and DEX pool/LP claims, commits a block,
exports, validates, re-imports, compares supply/LP ownership, and asserts all
invariants.

### High — InitGenesis silently skipped unexpected failures

Malformed JSON, validator restore errors, and invalid verifying keys now panic
during the module's no-error-return InitGenesis boundary instead of committing
partial custom state.

### High — rebased LP parser needed collision-free key support

GH-12's global LP export/orphan detection originally parsed the earlier textual
key. It now decodes GH-10's length-prefixed denom format, preserving isolation
for valid names such as `atom` and `atom:staked`.

### Medium — invariant fixture did not verify domain creation

The escrow-parity corruption fixture invoked the no-error-return `CreateDomain`
helper without checking its effect. It now reads the domain back and validates
the expected treasury, preventing a fixture failure from being misread as
successful registered-invariant evidence.

The operator limitation also records the complete safe-bootstrap contract: a
real Ed25519 key must identify a positive-power CometBFT validator, the custom
stake must be sufficiently and exactly bank-backed, and canonical supply must
remain within the 21,000,000 PNYX cap.

## Verification

- `go build` and CLI `--help` / `--version`: PASS
- `go vet ./...`: PASS
- `go test ./... -count=1`: PASS, 615 Go cases
- `go test -race ./... -count=1`: PASS
- Coverage: root 66.1%, token 92.6%, treasury 97.0%, DEX 45.3%, governance 56.6%
- `cargo test --workspace`: PASS, 26 Rust cases
- `cargo audit --no-fetch`: no blocking vulnerability; six tracked transitive
  CosmWasm/Wasmer warnings
- Maintained client offline install/lint/six tests/build/audit: PASS, zero
  vulnerabilities
- `go mod verify`, docs/JSON/diff checks: PASS
- Review remediation: focused registered-invariant regression, full Go tests,
  `go vet ./...`, and `go build ./...`: PASS
- GitHub Docs, DeepScan, Web, Mobile, Rust, Go build/vet/test, and both Docker
  jobs: PASS
- Manually dispatched Security Scan run `29158360390`: PASS, all five jobs

Total recovery evidence: 647 tests (615 Go + 26 Rust + 6 maintained client).

## Explicitly out of scope

- GH-20: ZKP signal/verifying-key and recipient-binding audit
- GH-21: production PoD node initialization, persistent lifecycle, and restart
- Independent stacked review and final merge consolidation (CodeRabbit pending)

The current `scripts/init-node.sh` still assumes unavailable `x/staking` gentx
commands. Do not use it for production launch.
