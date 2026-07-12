# TrueRepublic — Token Supply and Ledger Audit
> Scope: `treasury/keeper`, `x/truedemocracy`, `x/dex`, app wiring, genesis, ZKP authentication, maintained-client safety, and native node lifecycle  ·  Date: 2026-07-12  ·  Result: 16 PASS

## Summary

The recovery branches now define one six-decimal `upnyx` base denomination,
validate and enforce the 21,000,000 PNYX cap against canonical bank supply,
back governance/stake claims with bank escrow, mint validator/domain
inflation through one capped issuance service, and settle DEX reserves, LP
ownership, swaps, and burns against canonical bank state. GH-12 now validates
and reconciles custom genesis before mutation, exports LP ownership, and checks
supply, escrow, reserves, and LP conservation every block. This ledger slice
passes on stacked, unmerged recovery branches; it is not an external audit or
production approval.

> Remediation update: GH-11 and GH-14 pass their local/GitHub gates. GH-13 now
> routes validator/domain inflation and validator slash burns through canonical
> bank supply, clips aggregate minting at the cap, and commits both EndBlock
> inflation phases atomically. Anonymous rating rewards remain deferred because
> current proof/signature payloads do not bind a safe recipient. GH-10 now
> provides bank-backed DEX custody, provider-owned LP shares, authority checks,
> and canonical burns. GH-12 now closes custom-genesis and runtime-invariant
> findings locally. GH-20 now binds anonymous votes to chain/proposal/rating,
> fails closed on pinned genesis VK state, preserves active nullifiers, and
> disables mock client submission. GH-21 now replaces the MemDB placeholder
> with persistent Cosmos/Comet lifecycle and generated-key, bank-backed PoD
> genesis. Recipient binding, a real prover/ceremony, multi-node operations
> evidence, and independent review remain open. GH-21 GitHub Go, Docker, docs,
> client, Rust, static, and security gates pass.

## Findings by domain

### Denomination and bank-genesis supply cap — PASS

- **[PASS] Chain, maintained client, contracts, and operator docs use one 21M denomination model** — `token/denom.go`, `client-web/src/config/chains.ts`, `docs/node-operators/configuration/genesis-params.md`
  - What: `upnyx` is the canonical base denom, `pnyx`/PNYX is the six-decimal display unit, and the cap is 21,000,000,000,000 base units.
  - Evidence: GH-11 boundary tests, full local polyglot gates, and all stacked PR #15 checks pass.
  - Fix: Preserve the centralized token package and reject reintroduction of display-denom balances.

- **[PASS] Genesis and every recovery reward mint use canonical capped bank supply** — `token/denom.go`, `token/issuance.go`, `app.go`, `x/truedemocracy/validator.go`
  - What: Bank genesis rejects supply above 21,000,000 PNYX; the issuance service reads `x/bank` supply and mints at most the remaining base units into authorized module escrow.
  - Evidence: Tests cover below/at/above-cap supply, aggregate final-unit allocation, invalid supply, bank failure, exact burn/remint behavior, and end-block treasury/stake parity.
  - Fix: Preserve GH-10's canonical DEX burn path and GH-12's independent runtime supply invariant.

### Governance, treasury, and staking accounting — PASS

- **[PASS] Authenticated treasury and validator claims are backed by exact module escrow** — `x/truedemocracy/escrow.go`, `x/truedemocracy/msg_server.go`, `x/truedemocracy/treasury_bridge.go`, `x/truedemocracy/escrow_test.go`
  - What: Domain creation, proposal fees, validator registration, deposits, withdrawals, stake settlement, and signer claims now use exact `upnyx` bank transfers and cached atomic state transitions.
  - Evidence: Tests reject zero-balance declarations, spoofed identities, duplicate credits/pubkeys, dust withdrawals, missing CosmWasm signers, and injected transfer/burn failures; parity is checked after creation, fees, stake, withdrawal, payouts, and slashing.
  - Fix: Preserve `CacheContext` boundaries and GH-12's exact genesis reconciliation.

- **[PASS] Validator/domain inflation is bank-minted and vote rewards settle from escrow** — `token/issuance.go`, `x/truedemocracy/validator.go`, `x/truedemocracy/module.go`, `x/truedemocracy/issuance_test.go`
  - What: Validator and active-domain rewards mint exact capped `upnyx` into module escrow before matching internal claims commit. Both inflation phases share an outer EndBlock cache; vote rewards transfer existing treasury escrow and unsafe anonymous recipients remain deferred.
  - Evidence: Tests cover final-cap aggregation, first/second mint rollback, canonical decay, interval-only payouts, vote-transfer neutrality, slash burns, and escrow parity.
  - Fix: Preserve the single issuance boundary; complete recipient-bound anonymous reward claims in GH-7.

### DEX custody and authorization — PASS

- **[PASS] DEX reserves, swaps, burns, and LP shares are bank-backed** — `x/dex/custody.go`, `x/dex/msg_server.go`, `app.go`
  - What: Exact module-bank custody backs every pool reserve; cached public transitions move real inputs/outputs, assign shares by provider, and burn PNYX through the canonical issuance service.
  - Evidence: Tests cover direct/two-hop settlement, constant-product monotonicity, supply burns, foreign-share withdrawal rejection, denom-prefix key isolation, reserve/share divergence, slippage, and injected transfer/burn rollback.
  - Fix: Preserve the custody wrapper as the only public transition path and register its parity checks as runtime invariants in GH-12.

- **[PASS] Asset registry mutation requires chain authority** — `x/dex/custody.go`, `x/dex/msg_server.go`
  - What: Register and status messages compare the authenticated signer with the configured chain authority before state mutation.
  - Evidence: Unauthorized registration/status changes fail; the configured authority path succeeds.
  - Fix: Keep authority wiring explicit and exercise proposal execution when the governance runtime is completed.

### Genesis and invariants — PASS

- **[PASS] Custom genesis is structurally valid and exactly bank-backed before mutation** — `genesis.go`, `x/truedemocracy/genesis_validation.go`, `x/dex/genesis.go`
  - What: Full-app initialization rejects malformed, duplicate, negative, unbacked, share-divergent, and over-cap state before any custom module writes. Export includes provider LP ownership and non-empty export/import preserves supply and custody.
  - Evidence: Full-app and module tests cover exact treasury/stake/DEX backing, duplicate pools/assets/providers, invalid validators/domains, non-empty round trips, and startup at the cap boundary.
  - Fix: Preserve pre-mutation validation and require real CometBFT validator public keys; never restore a hard-coded bootstrap private secret.

### Precision and deterministic math — PASS

- **[PASS] Tokenomics and AMM calculations use deterministic integer/legacy-decimal math** — `treasury/keeper/rewards.go:29`, `x/dex/keeper.go:79`
  - What: Consensus calculations avoid floating-point arithmetic and truncate deterministically.
  - Path: Identical integer inputs produce identical results on all validators.
  - Fix: Preserve this property when introducing base-unit constants and bank accounting.

- **[PASS] Release decay clamps rewards at and above its configured maximum** — `treasury/keeper/rewards.go:75`
  - What: Negative decay is clamped to zero.
  - Path: A release value above the configured denominator does not produce a negative reward.
  - Fix: Retain the clamp and canonical bank-supply input.

- **[PASS] Maintained-client PNYX formatting no longer loses integer precision** — `client-web/src/utils/format.ts:4`
  - What: Display/base-unit conversion uses validated strings and `BigInt`.
  - Path: Values above JavaScript's safe integer limit round-trip without `Number` precision loss.
  - Fix: Reuse the same helper everywhere after the canonical denom/decimal decision.

### Recovery guardrails — PASS

- **[PASS] Registered crisis invariants detect every ledger divergence** — `token/invariant.go`, `x/truedemocracy/module.go`, `x/dex/module.go`, `genesis_integration_test.go`
  - What: `x/crisis` checks canonical supply, governance escrow, DEX reserves, and provider LP conservation every block.
  - Evidence: Full-app tests deliberately corrupt each of the four boundaries and prove the registered invariant halts; exact state passes.
  - Fix: Keep the one-block invariant period until equivalent operational evidence justifies a measured alternative.

- **[PASS] Public ledger claims distinguish stacked recovery from production** — `README.md`, `docs/status.json`, `docs/index.html`
  - What: Public status identifies the exact locally verified branch and retains the non-production warning for unmerged review, ZKP, and node-lifecycle work.
  - Evidence: Documentation consistency checks use one 683-test source of truth.
  - Fix: Keep status evidence-based through final stack consolidation.

### ZKP authentication and replay resistance — PASS

- **[PASS] Anonymous proofs and legacy signatures bind chain and exact vote context** — `x/truedemocracy/merkle.go`, `x/truedemocracy/keeper.go`, `x/truedemocracy/zkp.go`
  - What: Versioned length-prefixed signals bind chain ID, domain, issue, suggestion, and rating. The one-vote nullifier includes chain/proposal identity but deliberately excludes rating.
  - Evidence: Altered-rating and cross-chain proof/signature regressions fail without consuming the valid nullifier; the correctly bound proof succeeds.
  - Fix: Preserve the versioned encoding and require circuit/prover migrations to be explicit consensus upgrades.

- **[PASS] Consensus setup and genesis ZKP state fail closed** — `x/truedemocracy/anonymity.go`, `x/truedemocracy/genesis_validation.go`, `x/truedemocracy/module.go`
  - What: Transactions never run randomized Groth16 setup. Genesis pins circuit ID, VK SHA-256, BN254 curve, public-input shape, and canonical bytes; identity roots are recomputed from canonical commitments.
  - Evidence: Tests reject missing/mismatched IDs and fingerprints, trailing VK bytes, malformed fields, mismatched trees, and absent VK at vote time.
  - Fix: Treat genesis ceremony artifacts as the trust anchor and require external review before enabling a real prover.

- **[PASS] Double-vote state and mock-client boundaries survive lifecycle changes** — `x/truedemocracy/module.go`, `client-web/src/services/zkp.ts`, `web-wallet/src/services/api.js`
  - What: Export/import preserves the exact active nullifier records and heights without resurrecting values cleared by Big Purge. Both web clients reject mock proof submission.
  - Evidence: Nullifier round-trip/purge regressions, 8 maintained-client tests, and 4 focused legacy-client tests pass; both clients build and audit cleanly.
  - Fix: Keep anonymous rewards deferred and submission disabled until a compatible real prover and recipient-binding design pass independent review.

### Persistent node lifecycle — PASS

- **[PASS] Generated-key PoD bootstrap and persistent restart replace the in-memory placeholder** — `server_lifecycle.go`, `server_lifecycle_test.go`, `genesis_integration_test.go`
  - What: Standard Cosmos server commands use the configured home and database.
    `init` binds the generated CometBFT Ed25519 public key to matching custom
    consensus state with exact cap-checked bank backing, refuses conflicting
    validator sets, and writes genesis atomically with mode `0600`.
  - Evidence: 649 Go cases pass. A real daemon produces blocks, stops cleanly on
    SIGINT, restarts from the same home, advances height, runs invariants, and
    exports state; targeted lifecycle race, vet, CGO build, and both CLI version
    interfaces pass.
  - Fix: Preserve the green GitHub Docker restart/security gates and require
    separate independent multi-node/IBC/upgrade operations evidence before a
    public network. See `docs/agent-bridge/PR23_AUDIT.md`.

## Priority matrix

### 🔴 BLOCKING

None inside the locally implemented ledger/ZKP/single-node lifecycle slice. A
real prover, external cryptographic review, recipient binding, independent
multi-node operations review, and ordered merge remain
project-level release blockers.

### 🟠 HIGH

None identified in this audit slice.

### 🟡 MEDIUM

None identified in this audit slice.

### 🟢 LOW

None identified in this audit slice.
