# Community Audits — Collateral Web3 Open Audits

External audit series of the TrueRepublic project, commissioned by the repository
owner and published here with his explicit authorization (public series like
IFR/Ekklesia/Stealth/Prometheus). **Report-only**: no audited code was changed;
no node started; no chain actions; no keys/mnemonics/signer material touched
(repo AGENTS.md). The project's own review protocol was honored: the baseline
SHA was recorded and `go run ./cmd/security-review` **passed** on the frozen
checkout ("security review readiness contract verified; independent review and
production claims remain false"). The internal findings register
(`configs/security/review-findings.json`) was read and **not** re-reported;
findings below are new and map to threat-model IDs where applicable.

- **Audit date:** 2026-09-15
- **Baseline:** `main` @ `f5de5a150b9ff0edc51d68411b7829e78bcf78e6`
- **Register:** TRR-01 … TRR-42 — **3 Critical / 4 High / 15 Medium / 16 Low / 4 Informational**
- **Severity context:** TrueRepublic is a recovery project, not production-approved;
  no public network exists. TRR-01 is a live-code consensus defect reachable in
  normal usage on any chain running this code. TRR-13/TRR-14 are conditional on
  CosmWasm instantiation (no evidence any contract is instantiated).
- **Prior baselines:** CODEX_AUDIT.md (2026-07-23, 0F/2W/18P, superseded-banner)
  re-baselined — all supersession claims verified in repo evidence (TRR-37);
  docs/GH-241-V4-0-CODEX_AUDIT.md (V4-0, disjoint scope, no contradiction).
- **Finding tracker:** umbrella issue (see issue list) with one checkbox per finding.

## Reports

| # | Report | Register | Severity (C/H/M/L/I) | SHA-256 |
|---|--------|----------|----------------------|---------|
| 1 | [Full-scope security (Go chain)](trr-full-scope-audit-2026-09-15.md) | TRR-01…12 | 1/2/5/4/0 | `d9b3958857cc50c8499a489d8d48cd7fd601daf72bb863b21eb926228be64869` |
| 2 | [ZKP, cryptography & contracts](trr-zkp-crypto-contracts-audit-2026-09-15.md) | TRR-13…22 | 2/2/4/1/1 | `4e380d50eefa5883a80e36a862c6534e8b992a588e02a95283eeb9d352ee0e52` |
| 3 | [Integration & surfaces](trr-integration-surfaces-audit-2026-09-15.md) | TRR-23…28 | 0/0/3/2/1 | `99f5d30b10ecc1a139c1947bf02f67d43a2c4176070a1a36cf6448694ab971cc` |
| 4 | [Content & coherence](trr-content-coherence-audit-2026-09-15.md) | TRR-29…37 | 0/0/3/5/1 | `0a0b6c51bac84daf62c8140220a1463617e38971cd40003378f0ed0d24563ba9` |
| 5 | [AI-readiness & landing quality](trr-ai-readiness-landing-audit-2026-09-15.md) | TRR-38…42 | 0/0/0/4/1 | `f8154754fd7137fdd016a16e39ca2e7bc947d83009b0c470d8fa8e69f13e36db` |

## Headline findings

- **TRR-01 (Critical, live code):** `ElectAdmin` permanently corrupts every
  domain's admin field (raw-bytes vs bech32 confusion,
  `x/truedemocracy/governance.go:135,139`) — reachable in normal usage
  (default `AdminElectable: true`, runs every EndBlock), locks domain treasuries
  forever, and bricks export/import recovery. Tests mask it with ASCII
  pseudo-addresses. **Fix before any further recovery drill.**
- **TRR-13/TRR-14 (Critical, if instantiated):** `core/treasury.rs` is an
  unbacked fiction ledger (unauthenticated withdraw, recipient ignored, no
  funds move); `examples/zkp-aggregator` performs "anonymous voting" with
  self-asserted nullifier strings — unlimited Sybil votes emitting real
  Go-state messages.
- **TRR-02 (High):** validator stake effectively illiquid — the WP §7 10%
  cumulative payout limit gates *all* exits; bootstrap validators can never exit.
- **TRR-03 (High):** unlimited pseudonymous voting keys per member → ballot
  stuffing + N-fold RateToEarn payouts; admin-approval onboarding is dead code.
- **TRR-19/20 (Medium):** client identity secret plaintext in localStorage
  (docs claim "encrypted"); mock FNV commitments registerable on-chain via real
  signed transactions → permanent dead leaves + misleading privacy UI.
- **TRR-24/25 (Medium):** wiki operations pages teach broken setups (6 defects);
  fictional `truerepublic.network` (NXDOMAIN) used as security contact and seed
  host — squatting hazard.
- **TRR-29…31 (Medium):** whitepaper overclaims ("contracts are live", absolute
  anonymity, GG20 TSS boilerplate, price-increase promises).

## Verified strengths (selection)

Project review-contract verifier PASS on the frozen baseline; 21M cap single
issuance boundary + per-block crisis invariant; escrow parity incl. pending
exit holds; CacheContext discipline on every multi-step mutation; signer-vs-
identity enforced by generated GetSigners; exemplary genesis ZKP pinning
(circuit ID + VK SHA-256 + BN254 + canonical re-encoding + trailing-byte
rejection); real gnark Groth16 verifier with recipient-bound v2 signals;
client ZKP fail-closed (submission disabled, tx-registry omission); bundle
budget CI-gated (GH-128 closure real); zero third-party scripts on the landing;
headline numbers CI-pinned and independently re-verified (2,462 decomposition,
issue #29 arithmetic); wiki↔repo-wiki sync; license stack machine-enforced;
no secrets, private keys, or active infrastructure credentials found
(grep-verified).
