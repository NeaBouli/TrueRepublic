# PR #22 Audit — GH-20 ZKP Authentication and Replay Resistance

Date: 2026-07-12
Branch: `fix/GH-20-zkp-binding`
Base: `fix/GH-12-genesis-invariants`
Issue: [GH-20](https://github.com/NeaBouli/TrueRepublic/issues/20)
Result: PASS for the fail-closed on-chain/client scope; real prover and external
cryptographic review remain release blockers

## Scope and guarantees

- Groth16 public inputs bind a versioned, length-prefixed chain ID, domain,
  issue, suggestion, and exact rating signal.
- The one-vote nullifier includes chain/proposal identity but excludes rating,
  so changing the score cannot create a fresh voting scope.
- Consensus handlers never compile a circuit or perform randomized trusted
  setup. Missing or mismatched genesis VK configuration fails closed.
- Genesis pins the expected circuit ID, VK SHA-256, BN254 curve, four-public-
  input shape, and canonical serialized bytes.
- Commitments, roots, nullifiers, and proof public inputs use exact canonical
  BN254 field encodings; genesis recomputes identity roots from commitments.
- Export/import preserves the exact active nullifier records and used heights,
  while values deliberately cleared by Big Purge remain cleared.
- Maintained and legacy web clients cannot generate or broadcast mock proofs.
- Anonymous rewards remain deferred because no safe recipient is proof-bound.

## Audit findings fixed

### Critical — active nullifiers were not exported

Nullifier consumption lived in dedicated KV records while module export only
serialized ratings. Re-import therefore lost the active one-vote set. Inferring
all historical rating nullifiers would also be wrong because Big Purge
deliberately clears that set. Genesis now exports/imports the exact active
records and heights. Regressions prove both persistence and purge semantics.

### Critical — a legacy wallet broadcast random mock proof bytes

The legacy web wallet labeled random bytes as a ready proof, offered a Submit
Anonymous Vote action, and called `signAndBroadcast`. The API now rejects every
anonymous submission until a compatible real prover exists; the UI is disabled
and explicitly labeled preview-only. The maintained client also rejects both
initialization and direct proof generation and never invokes its submit callback.

### High — verifying-key identity was only deserialization-checked

Any structurally valid BN254 VK could pass genesis validation. The trusted
genesis artifact now declares the exact supported circuit ID and VK SHA-256;
validation also checks BN254, four public inputs, canonical bytes, no trailing
data, and exact fingerprint equality. This pins the ceremony output selected by
the chain specification; it does not replace an external ceremony/circuit audit.

### High — ZKP genesis fields and identity roots were not reconciled

Genesis now rejects short, uppercase, out-of-field, duplicate, or malformed
commitments/root history/nullifiers/domain keys, recomputes the current MiMC
Merkle root, caps history length, and rejects roots without commitments.

### High — proof signal and legacy signatures lacked complete replay context

The circuit has an explicit rating signal public input. Both proof signals and
legacy Ed25519 payloads use the same versioned length-prefixed chain/proposal
context. Altered ratings and cross-chain replay fail in regression tests.

### High — transaction-time trusted setup was nondeterministic

`EnsureVerifyingKey` previously performed randomized Groth16 setup on first
consensus use. It now only loads genesis-configured bytes or returns an error.
Setup remains a test/tooling function and is not reachable from handlers.

## Verification

- `go build ./...`, `go vet ./...`: PASS
- `go test ./... -count=1`: PASS, 643 Go cases
- `go test -race ./... -count=1`: PASS
- Coverage: root 66.1%, token 92.6%, treasury 97.0%, DEX 45.3%, governance 58.9%
- `go mod verify`: PASS
- `cargo test --workspace`: PASS, 26 Rust cases
- `cargo audit --no-fetch`: no blocking vulnerability; six tracked transitive
  CosmWasm/Wasmer warnings
- Maintained client offline install/lint/8 tests/build/audit: PASS, zero
  vulnerabilities; bundle performance warning remains
- Legacy web focused ZKP 4 tests/build/audit: PASS; obsolete CosmJS/Create React
  App architecture and source-map warnings remain non-production blockers
- Diff, JSON, documentation, production-secret, and consensus-setup checks: PASS
- GitHub Docs, DeepScan, Web, Mobile, Rust, both Go/Docker runs, and manually
  dispatched Security Scan run `29159603247`: PASS

Authoritative recovery evidence: 677 tests (643 Go + 26 Rust + 8 maintained
client). Four focused legacy-web regressions are additional and excluded from
that total.

## Explicitly out of scope

- Compatible real browser/mobile Groth16 prover and proving-key distribution
- Independent trusted-setup ceremony and circuit/VK audit
- Privacy-preserving anonymous reward recipient binding
- GH-21 production PoD node initialization, persistence, restart, and ops
- Final independent stacked review and merge consolidation

CodeRabbit's status check is green, but its conversation explicitly reports the
review limit was reached and contains no substantive findings. It is not counted
as the required independent cryptographic review.

Do not enable anonymous submission or claim production anonymity until the
out-of-scope cryptographic deliverables are complete.
