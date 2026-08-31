# GH-266 Maintained-Client WASM Proof Keeper Replay Audit

Date: 2026-08-31

## Decision boundary

GH-266 is a repository-only, test-only compatibility control. It extends the
existing GH-206 isolated Go/WASM prover gate to the real truedemocracy keeper
payout boundary. It changes no consensus, runtime, genesis, migration, wallet,
RPC, broadcast, or production path. The committed constraint/proving/verifying
artifacts remain forge-capable single-party toxic-waste fixtures. The maintained
client keeps `isSubmittable` hard false.

## Implemented control

- `client-web/src/services/zkpWasmProver.integration.test.ts` writes
  `truerepublic/zkp-wasm-handoff/v1` with the exact chain, domain, issue,
  suggestion, rating, canonical reward recipient, proof, Merkle root, nullifier,
  and four public signals used to generate the fresh proof.
- `x/truedemocracy/zkp_wasm_keeper_test.go` bounds the handoff at 16 KiB and
  rejects missing/unknown/duplicate/trailing JSON, noncanonical lowercase hex,
  noncanonical BN254 field elements, wrong proof length, invalid recipients,
  signal/root/nullifier inconsistency, and drift from the pinned GH-198,
  GH-203, and GH-209 vectors.
- The gate reconstructs an in-memory domain, member commitment, proposal,
  verifying key, treasury, and bank escrow, then invokes the actual
  `RateProposalWithZKPPayout` path with the fresh maintained-client proof.
- Success requires the exact calculated PNYX reward at only the signal-bound
  recipient, one rating, one consumed nullifier, exact treasury/module/payout
  accounting, and escrow parity. Exact replay must fail without double payment.
- Fresh keeper instances prove corrupted proof, recipient substitution, rating
  drift, foreign chain scope, unrecognized root, uppercase/noncanonical
  recipient, and blocked module recipient all leave rating, nullifier, treasury,
  payout, escrow, module balance, and watched account balance unchanged.
- `scripts/test-zkp-wasm-client.sh` fails if the handoff is empty and executes
  both the native verifier and keeper tests with the environment gate enabled,
  so protected CI cannot silently record skipped keeper evidence.

## Agent contributions and Sol remediation

- Claude Code completed a bounded read-only inventory of documentation surfaces
  and focused reproduction commands. It produced no file change.
- Kimi K3 independently analyzed the missing seam, implemented the three-file
  test harness, and ran the first focused verification set.
- Sol reviewed every changed line and corrected two integration risks before
  acceptance: required JSON fields are checked explicitly, and the SDK Bech32
  prefix is configured only after the opt-in environment gate is present. Thus
  normal skipped package tests cannot mutate global address configuration.
- Kimi's independent full-diff review found no P0/P1/P2 and one P3 stale wiki
  handler count. Sol corrected `all 13` to the verified 26 message handlers;
  Kimi's focused follow-up returned **APPROVE** for the GH-266 changeset. The
  follow-up also exposed pre-existing sibling count/version drift; Sol folded
  those exact documentation-only corrections into this closeout. A final
  stale-value review found and Sol removed the remaining adjacent DEX, treasury,
  query, CometBFT, Go, CosmJS, and standard-test count drift before publication,
  including the final system-overview module table exposed by the last focused
  review.

## Reproduced local evidence

- `./scripts/test-zkp-wasm-client.sh`: PASS. One maintained-client real-WASM
  proof test, the native Go verifier consumer, the keeper happy/replay path and
  all adversarial subtests pass.
- `go test ./x/truedemocracy -run '^TestWASMClient' -count=1 -v` without the
  environment variable: PASS with both opt-in tests skipped before any global
  prefix mutation.
- `make verify`: PASS for package selection, build, Vet, and complete
  Race/Coverage. Governance remains 64.3%, above the 63.7% critical threshold.
- Maintained client: ESLint, ten Node cases, 309 Vitest cases, TypeScript/Vite
  production build, exact bundle budgets, and high-severity npm audit PASS.
- Rust/CosmWasm: format, strict Clippy, workspace build, all 26 tests, and audit
  PASS. Audit reports only the five centrally allowed transitive maintenance/
  unsoundness warnings and no blocking vulnerability.
- Exact active no-fix Go vulnerability policy and negative fixtures, pinned
  staticcheck, gitleaks, critical coverage, bounded generative/fuzz campaigns,
  deterministic daemon, repeated OCI, candidate, install, release,
  rollout-genesis, license, retirement, documentation, JSON, shell, and diff
  contracts PASS.
- A fresh package-scoped JSON recount reports 614 passing governance events and
  three skips. The two GH-266 tests are separate-gate evidence, so the standard
  source-of-truth remains 2,187 = 1,842 Go + 26 Rust + 319 maintained-client.
- Kimi's final read-only review re-counted 36 treasury, 138 DEX, 614 governance
  and 1,842 complete Go standard-suite pass events, re-ran documentation and
  diff gates after the final edit, and returned **APPROVE** with zero unresolved
  P0/P1/P2/P3 finding.

## Residual risks and production gate

- The proof is cryptographically real but generated from publicly committed,
  synthetic toxic-waste test artifacts. Anyone with those artifacts can forge
  fixture proofs; they are prohibited from production.
- The test configures the production `truerepublic` Bech32 account prefix in a
  dedicated filtered Go test process. Future script changes must retain that
  isolation and must not run the opt-in tests after unrelated address encodings
  in the same process.
- Atomic direct payout remains publicly linkable to the selected address. Fresh
  addresses reduce reuse linkage but do not provide shielded payout privacy or
  coercion resistance.
- Production ceremony/provenance, reproducible keys, audited prover and
  submission integration, independent cryptographic/privacy review, real
  network evidence, and rollout approval remain open.
- GH-266 closes no Phase 2 checkbox and grants no rollout credit. State remains
  35/59 overall, 35/51 phase work, Phase 6 at 6/7, and production false.

## Protected closeout evidence

PR #267 final head `f736948` passed all 29 hosted contexts. All four CodeRabbit
threads were answered and resolved before the authorized squash merge. Exact
main `735f42af889a21643d7c37d42128b0fea11d9800` passed Docs
`33436474061`, Client `33436474049`, Go CI `33436474108`, Security
`33436474119`, Reproducible Linux Daemon `33436474216`, and Pages
`33436472010`; GH-266 is closed. The live Wiki is synchronized at
`b1cf839ac8daacfc2649e4304af15279de908d23`, the cache-busted Landing Page and
`status.json` expose the test-only keeper boundary, and GH-29 records the
evidence in comment `5484406635`. No rollout credit or production action
occurred.
