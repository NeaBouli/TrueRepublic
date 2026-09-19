# TrueRepublic — ZKP, Cryptography & Contracts Deep Audit

- **Series:** Collateral Web3 Open Audits
- **Date:** 2026-09-15
- **Target:** NeaBouli/TrueRepublic @ `f5de5a150b9ff0edc51d68411b7829e78bcf78e6`
- **Scope:** x/truedemocracy ZKP stack (merkle/zkp/anonymity), Rust/CosmWasm contracts (core + 4 examples), client-web crypto surface (zkp.ts, identity/wallet stores)
- **Method:** contracts+client deep-recon agent + lead verification of every finding at file:line; read-only; no keys touched
- **Pre-acknowledged (not re-reported):** SRF-ZKP-001 — no production Groth16 prover/ceremony (critical, internal, open)
- **Register:** TRR-13 … TRR-22 (this report) — **2 Critical / 2 High / 4 Medium / 1 Low / 1 Info**

---

## Executive summary

The **verifier side** of the ZKP stack is exemplary: real gnark Groth16 verification, genesis-pinned VK (circuit ID + SHA-256 + BN254 + public-input count + canonical re-encoding + trailing-byte rejection, re-checked at InitGenesis), versioned v2 signals binding chain/domain/issue/suggestion/rating/recipient with unambiguous length-prefixing, nullifiers scoped to survive rating/recipient changes and export/import, setup never in the tx path, prover reachable only from tests. **The contract layer is the opposite:** the `core` treasury contract is an unbacked fiction ledger with an unauthenticated withdraw that ignores its recipient, and the `zkp-aggregator` example wraps "anonymous voting" language around a flow that verifies **nothing** — self-asserted nullifier strings grant unlimited votes while emitting real state-changing messages into the canonical Go modules. No contract shows evidence of instantiation, so these are pre-deployment Criticals — the exact moment to delete or rewrite them. On the client, the identity secret sits **unencrypted** in localStorage despite the type docs claiming otherwise, and mock FNV commitments can be registered on-chain via real signed transactions, permanently poisoning the identity tree.

## Severity table

| ID | Severity | Title |
|----|----------|-------|
| TRR-13 | Critical¹ | `core/treasury.rs`: unbacked fiction ledger — Withdraw unauthenticated, recipient ignored, no funds ever move |
| TRR-14 | Critical¹ | `examples/zkp-aggregator`: "anonymous" voting with self-asserted nullifiers — unlimited Sybil votes emitting real Go-state messages |
| TRR-15 | High¹ | `examples/governance-dao`: zero-vote execution, no voting-period enforcement, treasury-drain chain if made domain admin |
| TRR-16 | High¹ | `core/governance.rs`: unlimited repeat voting; only the instantiator can ever vote |
| TRR-17 | Medium¹ | `examples/dex-bot`: shadow order book — no escrow, no settlement, admin-fed prices |
| TRR-18 | Medium¹ | All contracts: monolithic single-key state + no migrate/cw2 (brick trajectory) |
| TRR-19 | Medium | client-web: ZKP identity secret persisted **unencrypted** in localStorage — type docs claim "stored encrypted" (false) |
| TRR-20 | Medium | client-web: mock FNV commitments registerable on-chain via real signed tx → permanent dead leaves + misleading privacy UI |
| TRR-21 | Low | Legacy `Keeper.RateProposal` (in-keeper privkey signing) lacks the −5..+5 rating check; signs v1 while consensus requires v2 |
| TRR-22 | Info | Big Purge re-vote semantics (intended?); testing-utils mock always answers `used:false` |

¹ conditional on instantiation/wiring — no evidence any contract is instantiated; code ships in-repo and is presented as usable.

---

## TRR-13 — Critical¹ — `core/treasury.rs`: fiction ledger

**Evidence (lead-verified):** `contracts/core/src/treasury.rs:51-77` — `execute()` ignores `_info` entirely. `Deposit{amount}` only `checked_add`s an internal `state.balance` — no `info.funds` inspection, no `cw-utils::must_pay`, no denom check (coins actually sent are silently trapped). `Withdraw{amount, recipient}` has **no authorization check** (any sender), **ignores `recipient`** (`recipient: _`, `:67`), and emits **no `BankMsg::Send`** — it merely decrements the shared internal ledger. The "balance" is unbacked; anyone can burn it down; nothing real ever moves. It contradicts the Go modules' canonical bank custody 1:1. Zero tests exist for this contract (F-10: `contracts/core/` has none).

**Recommendation:** delete or rewrite — authenticate via `info.sender` vs stored admin, assert funds with `cw-utils::one_coin`/`must_pay`, send via `BankMsg::Send`; or remove the contract if the Go modules own custody (they do).

## TRR-14 — Critical¹ — `examples/zkp-aggregator`: "anonymous" voting with no proof at all

**Evidence (lead-verified):** `contracts/examples/zkp-aggregator/src/contract.rs:77-150` — `SubmitVote` takes a caller-chosen `nullifier_hex` string. The on-chain custom query only returns whether that string is *used* (`x/truedemocracy/wasm_bindings.go:282-286`); **no Groth16 proof, signature, or any credential is verified** — the `used_nullifiers` table is only populated by `RateProposalWithZKP` on the Go side, and the aggregator never marks anything on-chain. The local dedup (`:96-102`) is keyed by `(round_id, caller-supplied string)` — the same voter invents a fresh string per vote; nullifiers are reusable across rounds. Each vote also emits real `PlaceStoneOnSuggestion` custom messages (`:133-142`) into canonical Go state (landing only if the contract is a domain member, `stones.go:40,89`) and aggregates scores. Bonus divergence: a parallel, never-purged nullifier list contradicts the Go Big-Purge lifecycle (TRR-22).

**Impact (if instantiated):** "anonymous voting" that is actually unlimited pseudonymous ballot stuffing — worse than TRR-03 because it wears the ZKP brand.

**Recommendation:** never deploy; ideally remove in favor of the native `MsgRateWithProof` path; if pursued, votes must carry proof material verified on the Go side.

## TRR-15 — High¹ — `examples/governance-dao`: zero-vote execution + treasury-drain chain

**Evidence (lead-verified):** `contracts/examples/governance-dao/src/contract.rs:131-142` — the quorum check is skipped when `domain.member_count == 0`, and the threshold check is skipped when `total_votes == 0` (each guarded by `> 0`); with `quorum_bps == 0` (accepted at instantiate without validation, `:39-49`), a proposal with **zero votes** passes. `ExecuteProposal` never requires `env.block.time > open_until` (`:112-142`) — execution seconds after creation. It is callable by anyone; stored `config.admin` gates nothing. **Impact amplifier (cross-layer, agent-verified):** the wasm encoder sets `Sender = contract` (`x/truedemocracy/wasm_bindings.go:396-407`) and the Go keeper requires `authorizer == domain.Admin` (`treasury_bridge.go:74-77`) — if this contract is ever made a domain admin, anyone can drain the entire domain treasury via a zero-vote `WithdrawFromDomain` proposal (`contract.rs:166-172`).

**Recommendation:** validate bps params at instantiate (1..10000); require voting-period end and `total_votes > 0`; make checks unconditional; document that a DAO contract must never be domain admin until fixed.

## TRR-16 — High¹ — `core/governance.rs`: unlimited repeat voting; only the instantiator can ever vote

**Evidence:** `contracts/core/src/governance.rs:100-124` — `Vote` appends to `proposal.votes` with no per-voter duplicate check; authorization is a caller-supplied `public_key` string compared against stored pairs, but the only pair ever created is `{owner: instantiator, public_key: "pk_placeholder"}` (`:54-61`) with no registration path — effectively only the instantiator can vote, forever, with a publicly known "key". `votes: Vec<i8>` and `proposals` grow unboundedly in a single state blob (see TRR-18). No tests.

## TRR-17 — Medium¹ — `examples/dex-bot`: shadow order book

**Evidence:** `contracts/examples/dex-bot/src/contract.rs:63-86` — orders take no funds (`info.funds` never read); `Filled` status (`:159-168`) never transfers anything; fills are driven solely by admin-pushed snapshots (`:104-130`) with **no staleness check**, `CheckAndExecute` permissionless; zero `input_amount`, identical in/out denoms and foreign pairs all produce nonsensical fills (`:143-158`). Its AMM formula matches Go `x/dex/keeper.go:101-115` **today**, but fee/burn constants are duplicated rather than queried — silent-divergence hazard behind a "Matches Go" comment.

## TRR-18 — Medium¹ — Monolithic single-key state; no migrate/cw2 anywhere

**Evidence:** all five contracts serialize their entire state (ever-growing `Vec`s of proposals/voters/orders/schedules/rounds/nullifiers) under one key on every execute (`core/src/{governance,treasury}.rs:7-8`, `examples/*/src/contract.rs:11-13`) — growth → O(n) gas per tx → eventual permanent bricking. No `cw-storage-plus` `Map`/`Item`, no `migrate` entry point, no `cw2` version storage, no `reply` handlers anywhere in the workspace (grep-verified). `cw-utils`/`cw2`/`cw-storage-plus` are entirely absent from Cargo.lock — the root cause of the TRR-13/TRR-18 patterns. Dependency posture otherwise clean: cosmwasm-std/vm 3.0.9, serde 1.0.229, thiserror 1.0.69, schemars 0.8.22 — no known advisories (agent OSV check).

## TRR-19 — Medium — Client: identity secret plaintext in localStorage, docs claim "encrypted"

**Evidence (lead-verified):** `client-web/src/stores/identityStore.ts:23-82` persists the full `identity` object — **including `secret`** — via zustand `persist` → plaintext `localStorage["identity-store"]`; `client-web/src/types/zkp.ts:13` states "Client-side ZKP identity (**stored encrypted in localStorage**)" — false. `exportIdentity()` (`:57-60`) copies the raw JSON (secret included) to the clipboard. Wallet mnemonics at least get AES-GCM/PBKDF2 (`services/wallet.ts:370-476`); the identity secret gets nothing. Becomes High the moment real proofs ship (secret = identity = vote unlinkability).

**Recommendation:** encrypt with the wallet password (reuse `WalletService.encrypt`), strip `secret` from persisted state, warn on export, fix the type doc.

## TRR-20 — Medium — Client: mock FNV commitments registerable on-chain via real signed tx

**Evidence (lead-verified):** `client-web/src/services/zkp.ts:77-88,205-216` — `generateIdentity()` builds commitment/nullifier with a 32-bit FNV variant repeated 8× (2³² distinct values, explicitly not MiMC). Yet `OnboardingFlow.tsx:59-93` → `membership.ts:143-182` submits `MsgRegisterIdentity` with that commitment as a **real signed transaction**; the chain accepts any canonical 32-byte hex (`anonymity.go:292-296`). Result: permanent dead leaves in the canonical identity tree (no valid proof can ever exist for them), the identity shown as "Active", and UI text "Your votes cannot be linked back to you" (`IdentityManager.tsx:108-113`) — a misleading security claim for a preview identity. A real MiMC implementation exists in-repo (`services/zkpEncoding.ts:61-75`) but is unused by identity creation.

**Recommendation:** block `registerIdentity` while `isSubmittable == false`, or compute the commitment with `mimcBn254` so leaves remain usable; fix the UI claim.

## TRR-21 — Low — Legacy `Keeper.RateProposal` gaps

**Evidence:** `x/truedemocracy/keeper.go:188-266` — unlike `RateProposalWithSignature` (`:275`) and `RateProposalWithZKP` (`:358`), this path never enforces −5..+5; consensus entry is safe (`msg_server.go:585` signature path, `msgs.go:413-430` ValidateBasic), so the gap is reachable only via internal callers (`tree.go:45` `PropagateAsync` — see TRR-09 — and tests/CLI). An out-of-range rating stored this way also violates genesis export validation (`genesis_validation.go:648-650`). This path still signs the **v1** payload while comments say consensus requires v2 (`merkle.go:269-273`).

## TRR-22 — Info — Big Purge re-vote semantics; mock blind spot

After a Big Purge, wiped commitments/nullifiers mean a re-registered identity can vote again on the same suggestion (`big_purge.go:62-87`) — matches whitepaper §4 intent; flagged so the project confirms it is *intended* (the wasm-side aggregator contradicts it by never purging, TRR-14). Separately, `contracts/packages/testing-utils/src/mock_querier.rs:100` hard-codes `NullifierResponse{used:false}` — the on-chain-used rejection path is untestable with the current mock.

---

## Verified strengths (evidence-checked)

- **Genesis ZKP pinning is exemplary:** circuit-ID exact match, VK SHA-256 fingerprint, BN254 + public-witness count, canonical re-serialization equality, lowercase-hex enforcement, trailing-byte rejection, re-validation at InitGenesis (`zkp.go:257-279`, `genesis_validation.go:438-453`, `module.go:231-240`).
- **Signal/nullifier protocol design:** v2 binds chain/domain/issue/suggestion/rating (BE int64)/reward-recipient with unambiguous length-prefixing (`merkle.go:283-302`); nullifier scope deliberately excludes rating and recipient so the one-vote rule is reward- and change-proof (`merkle.go:249-267`, `keeper.go:394-397`).
- **Verify path never runs setup in consensus** (`anonymity.go:257-265`); VK comes only from genesis; proof/VK deserialization rejects trailing bytes (`zkp.go:221,243`); public inputs must be 32-byte canonical BN254 elements with non-zero signal (`zkp.go:176-203`).
- **Prover isolation:** `SetupMembershipCircuit`/`GenerateMembershipProof` reachable only from `*_test.go`; the wasm prover command exits(2) on non-wasm builds.
- **Nullifier lifecycle:** per-domain store, genesis round-trip preserved and validated, purge-on-Big-Purge with announcement events, purged nullifiers cannot resurrect via genesis (`zkp_genesis_security_test.go`).
- **Client ZKP posture is fail-closed:** `ZKPService.initialize`/`generateProof` throw unconditionally without an injected test prover (`services/zkp.ts:54-63,172-178`); `isSubmittable` hardcoded `false`; the vote button is disabled with an explicit warning banner (`VotingPanel.tsx:238-261`); `MsgRateWithProof` is commented out and **absent from `txRegistry.ts`** (fail-closed by omission — and no `MsgWithdrawFromDomain` either, so the client cannot perform admin treasury withdrawals at all).
- **Client money math:** BigInt-only (`utils/format.ts:8-31`, `services/ibcTransfer.ts:86-100`), strict regex parsing, over-precision rejection; denom/decimals match Go canonical exactly (`config/chains.ts:17-19` vs `token/denom.go:12,18`).
- **Go treasury bridge:** admin-only withdrawal, upnyx-only, cache-context debit-before-send, balance checks (`treasury_bridge.go:64-110`); payout recipients validated canonically, blocked module accounts rejected (`escrow.go:272-307`).
- **wasm custom bindings** derive identity from the contract sender (`wasm_bindings.go:346-412`); staking/distribution surfaces fail closed (`wasm_stubs.go`).
