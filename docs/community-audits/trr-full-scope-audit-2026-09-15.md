# TrueRepublic — Full-Scope Security Audit (Go Chain)

- **Series:** Collateral Web3 Open Audits
- **Date:** 2026-09-15
- **Target:** NeaBouli/TrueRepublic @ `f5de5a150b9ff0edc51d68411b7829e78bcf78e6`
- **Scope:** Cosmos SDK chain: app wiring, genesis, token/issuance cap, treasury math, x/truedemocracy (governance, escrow, validators, anonymity, stones, upgrades), x/dex, lifecycle, policies, build/deploy tooling
- **Method:** 3 deep-recon agents + lead verification of every finding at exact file:line; read-only; no node started; no keys/mnemonics touched; project review protocol honored (baseline SHA recorded, `make security-review-contract-test` **PASS** on the frozen checkout)
- **Pre-acknowledged (not re-reported):** SRF-ZKP-001 (no real prover), SRF-CLI-001, SRF-OPS-001, SRF-REL-001, SRF-REV-001
- **Register:** TRR-01 … TRR-12 (this report) — **1 Critical / 2 High / 5 Medium / 4 Low**

---

## Executive summary

The recovery-foundation discipline is real: the 21M cap has a single issuance boundary with a per-block crisis invariant, escrow parity holds through pending exit holds, every multi-step mutation runs in CacheContext, signer-vs-identity is enforced by generated GetSigners contracts, and the project's own fail-closed review-contract verifier passes on my checkout. **However, this audit found a critical consensus bug the internal audits missed:** the admin-election routine corrupts every domain's admin field through a raw-bytes/bech32 confusion — reachable in completely normal usage, irreversible on-chain, locking every domain treasury and bricking export/import disaster recovery. The test suite masks it with ASCII pseudo-addresses. Two more high-severity economic/governance defects (stake illiquidity, unlimited pseudonymous voting keys) compound it. All three are fixable before any network launch — which is exactly the phase this recovery project is in.

## Severity table

| ID | Severity | Title | TM-map |
|----|----------|-------|--------|
| TRR-01 | Critical | ElectAdmin permanently corrupts every domain admin (raw-bytes vs bech32 confusion) — treasuries lock, export/import bricks | TM-GOV-001 |
| TRR-02 | High | Validator stake effectively illiquid: 10% cumulative payout limit gates *all* exits; bootstrap validators can never exit | TM-TOK-001 |
| TRR-03 | High | Unlimited pseudonymous voting keys per member → ballot stuffing + N-fold RateToEarn payouts; admin-approval onboarding is dead code | TM-GOV-002 |
| TRR-04 | Medium | Stone-move reward farming: every placement/move pays treasury/1000, no cooldown | TM-TOK-001 |
| TRR-05 | Medium | Member exclusion does not revoke anonymous voting power (up to 90 days) | TM-GOV-001 |
| TRR-06 | Medium | Reserved "governance" domain never anchored → entire upgrade-governance machinery unreachable on init-node chains | TM-OPS |
| TRR-07 | Medium | DEX registry authority is an unspendable module account — register/status messages dead; hardcoded "atom" default asset | TM-TOK-002 |
| TRR-08 | Medium | Unbounded EndBlock work: three full domain scans per block over whole-blob domain values | TM-CON-001 |
| TRR-09 | Low | `tree.go` compiled into production: async goroutine consensus-state writes + hardcoded fake nodes wired into keeper | TM-CON-002 |
| TRR-10 | Low | `CoinBurnRequired` proposal check unsatisfiable; "burn" is actually a treasury credit | TM-TOK |
| TRR-11 | Low | Onboarding signature not chain-bound (cross-fork replay) | TM-GOV |
| TRR-12 | Low | Cluster: VoteToExclude can exclude admin without stripping rights; IBC light-client governance recovery impossible; DEX dust burns shares for zero output; panic in GetAllLPPositions | TM-GOV/IBC/TOK |

---

## TRR-01 — Critical — ElectAdmin permanently corrupts every domain admin

**Evidence (lead-verified):** `x/truedemocracy/governance.go:135` — `bestAddr == string(domain.Admin)`: `domain.Admin` is a 20-byte `sdk.AccAddress` (`types.go:61`); `string()` on it yields raw binary, never the bech32 member string, so the "no change" guard **never fires**. `:139` — `domain.Admin = sdk.AccAddress(bestAddr)`: stores the ~45 ASCII bytes of the bech32 **text** as the address.

**Trigger path (normal usage):** `EndBlock → ProcessGovernance` (`module.go:321`) runs `ElectAdmin` for **every** domain (`governance.go:337-347`) — and `AdminElectable: true` is the default for every `MsgCreateDomain` domain (`keeper.go:79`) and the genesis Bootstrap domain (`genesis.go:122`). As soon as any member places **one** member-stone (`MsgPlaceStoneOnMember`, no reserved-domain guard, `governance.go:31-61`), the election runs and corrupts the admin field.

**Impact:** all admin-gated operations compare `caller.Equals(domain.Admin)` — a 20-byte real address can never equal 45 bytes of bech32 text, so they become **permanently impossible**: `AddMember` (`keeper.go:100`), `WithdrawFromDomain` (`treasury_bridge.go:75` — **domain treasury locked forever**), `PurgePermissionRegister` (`anonymity.go:64`), `ApproveOnboardingRequest` (`anonymity.go:193`). Additionally `validateGenesisDomain` requires `admin ∈ members` (`genesis_validation.go:606`), so a chain containing one corrupted domain produces exports that **fail import** (`app.go:513,516`) — the export/import disaster-recovery path bricks. There is no on-chain repair path once corrupted.

**Why the tests missed it:** fixtures use ASCII pseudo-addresses (`governance_test.go:11,144`: `sdk.AccAddress("admin1")`, members `"alice"…`, assertion `string(domain.Admin) != "dave"`), where the buggy conversion round-trips cleanly. Real bech32 members break it.

**Recommendation:** `bestAddr == domain.Admin.String()` for the guard; `sdk.AccAddressFromBech32(bestAddr)` (with error handling) for the assignment; regression-test with real bech32 fixtures; audit the same `string(AccAddress)` anti-pattern repo-wide. **Highest priority of this audit — fix before any further recovery drill, since any chain that has run an election with real addresses already carries corrupted state.**

## TRR-02 — High — Validator stake effectively illiquid

**Evidence (lead-verified):** `x/truedemocracy/validator.go:374` routes every stake withdrawal through `ValidateStakeTransfer` (`governance.go:315-333`): cumulative `TransferredStake + amount ≤ TotalPayouts × StakeTransferLimitBps / 10000` (10%), and `TotalPayouts == 0` → **hard error** ("no payouts yet — stake transfers not allowed", `:321`). The full-exit path `RemoveValidatorWithEscrow` (`escrow.go:180-211`) calls the same `WithdrawStake` — a full exit needs `TotalPayouts ≥ 10×stake ≥ 1M PNYX` at minimum stake. The Bootstrap domain starts with treasury 0 (`genesis.go:120`), so bootstrap/genesis validators can **never** exit. Under-staked-after-slash validators can neither unjail (`slashing.go:63` requires stake ≥ StakeMin) nor exit → permanent lock. Perverse escape hatch: stone-reward farming inflates `TotalPayouts` (`stones.go:222`), so validators can unlock only by draining domain treasuries (compounds TRR-04).

**Impact:** a documented WP §7 rule (10% cumulative transfer limit, anti-whale) applied to *all* exits makes the minimum-viable validator position irreversible — for a recovery project courting test validators, this is a fund-lock trap waiting for its first real stake.

**Recommendation:** exempt full exits from the cumulative limit or switch to time-based unbonding; at minimum document the illiquidity in the validator guide.

## TRR-03 — High — Unlimited pseudonymous voting keys per member

**Evidence (lead-verified):** `JoinPermissionRegister` (`anonymity.go:16-52`) checks only membership and key duplication — **no per-member key cap**. `MsgOnboardToDomain` (`msg_server.go:659-700`) registers keys **directly** (signature-verified), bypassing the two-step admin-approval flow: `SetOnboardingRequest` has no production creator (grep: only tests and approve/reject updates). `AnyoneCanJoin` (`types.go:77`) is never enforced at runtime. `RegisterIdentityCommitment` (`anonymity.go:273-328`) has the same shape — unlimited ZKP commitments/nullifiers per member.

**Impact:** double-vote protection is per-key (`HasDomainKeyVoted`, `anonymity.go:93-108`) and per-nullifier — so one member = N keys = **N votes per suggestion** and **N RateToEarn payouts per suggestion** (`keeper.go:341,452`), draining domain treasuries and capturing rating outcomes. The whitepaper's Systemic-Consensing integrity depends on one-member-one-voice; the code does not enforce it.

**Recommendation:** cap keys/commitments per member (e.g. one active key), wire request→approve as the only onboarding path, enforce or delete `AnyoneCanJoin`. Note the design tension with anonymity (key↔member linkage) — an owner decision, but the current state is the worst of both worlds: pseudonymity *and* Sybil.

## TRR-04 — Medium — Stone-move reward farming

**Evidence:** `stones.go:73,127` call `payStoneReward` on **every** placement including pure moves (`:53-70,107-124`). A member can alternate one stone between two issues forever; each tx pays `CalcReward` = treasury/1000 (`treasury/keeper/rewards.go:33-38`) — ~694 moves ≈ 50% of a treasury at ~0.001 PNYX fee/tx. Also inflates `TotalPayouts` (feeds TRR-02's limit and domain-interest eligibility `validator.go:614`).

**Recommendation:** pay only on first placement, or per-epoch/per-member reward caps.

## TRR-05 — Medium — Member exclusion does not revoke anonymous voting power

**Evidence:** `removeMember` (`governance.go:206-257`) cleans the member list and stones only; `PermissionReg` keys and ZKP identity commitments survive. Rating handlers check only key-in-register (`keeper.go:300`) or commitment-in-tree (`keeper.go:366-418`), never membership. An excluded member keeps voting (and earning) until the next Big Purge — up to 90 days (`anonymity.go:114`).

**Recommendation:** on exclusion, drop the member's domain keys/commitments, or re-check membership at vote time (note the anonymity tension — owner decision).

## TRR-06 — Medium — Reserved "governance" domain never anchored

**Evidence:** `ensureConsensusGenesis` builds only the "Bootstrap" domain (`genesis.go:115-126`); runtime creation of `ReservedGovernanceDomain` is rejected (`escrow.go:49-51`); `VoteSoftwareUpgrade` requires it (`upgrade_gov.go:368-371`). The default init boundary (`init-node.sh:32`, `docker-entrypoint.sh:30-33`) therefore yields chains that can never schedule or cancel an upgrade — the entire governed-upgrade machinery (`upgrade_gov.go`, `upgrade_handlers.go`, `make governed-upgrade`) is unreachable on init-node chains.

**Recommendation:** anchor the governance domain (with an explicit electorate) in `ensureConsensusGenesis`, or provide a governed creation path.

## TRR-07 — Medium — DEX registry authority is an unspendable module account

**Evidence:** `RequireAuthority` (`custody.go:37-42`) demands signer == `k.authority`, set to the `gov` module address (`app.go:219,326`) — but no x/gov is mounted (`app.go:85-97`), and module accounts cannot sign. `MsgRegisterAsset`/`MsgUpdateAssetStatus` (`msg_server.go:176-212`) can therefore never execute: assets enter only via genesis, and real IBC voucher denoms can never be listed at runtime. `DefaultGenesisState` hardcodes an "atom" asset (`types.go:73-80`).

**Recommendation:** route registry control through the truedemocracy governance adapter (like x/upgrade) or declare the DEX genesis-frozen.

## TRR-08 — Medium — Unbounded EndBlock work

**Evidence:** per block: `ProcessAllLifecycles` (`module.go:318`, `lifecycle.go:129-138`), `ProcessGovernance` (`:321`), `CheckAndExecuteBigPurges` (`:324`, `big_purge.go:12-25`), plus `DistributeDomainInterest` iteration with per-domain snapshot writes each interval (`validator.go:603-623`). Each domain is a single blob containing all members/issues/ratings/commitments (`keeper.go:82-83`) — full unmarshal per domain per scan per block, ungas-metered in EndBlock. Growth is cheap: domain creation needs only ≥1 upnyx escrow (`escrow.go:14-24,55`), and proposal text fields have no length caps (`keeper.go:118-182`, `msgs.go:59-70`).

**Impact:** consensus-liveness degradation at scale — block time inflates with domain count × blob size.

**Recommendation:** cap domains/items/string fields; index state per-key instead of per-blob; meter or schedule the sweeps.

## TRR-09 — Low — `tree.go` compiled into production

**Evidence:** `Node.PropagateAsync` (`x/truedemocracy/tree.go:42-55`) spawns a goroutine that mutates consensus state through an `sdk.Context` — instant nondeterminism/fork if ever wired to a handler (no callers today). `Keeper.RateProposal` accepting a raw `*ed25519.PrivKey` (`keeper.go:188-266`) survives as production code. `BuildTree()` creates 7 hardcoded "TestParty" nodes with 100k PNYX stake each (`tree.go:23-40`) and is wired into the app keeper (`app.go:325`), though the field is otherwise unused.

**Recommendation:** delete or move to `_test.go`; drop the keeper `nodes` field.

## TRR-10 — Low — `CoinBurnRequired` unsatisfiable; "burn" is a credit

**Evidence:** `keeper.go:130`: `fee < CalcDomainCost(fee)` where `CalcDomainCost(fee) = 2000×fee` (`treasury/keeper/rewards.go:69-74`) — true for every positive fee, so such domains reject all proposals; the fee is credited to the domain treasury (`keeper.go:147`), not burned, contradicting the option's name. Latent: no runtime path sets the option today.

## TRR-11 — Low — Onboarding signature not chain-bound

**Evidence:** `ConstructOnboardingMessage` = `"ONBOARD:{requester}:{domain}:{pubkey}"` (`crypto.go:13-16`), verified in `msg_server.go:673`. No chain ID → the same global key reused on a fork/another chain allows replay. The v2 vote payloads do bind the chain (`merkle.go:283-294`) — extend that pattern here.

## TRR-12 — Low — Cluster

- `VoteToExclude` can exclude the domain admin without stripping admin rights (`governance.go:151-203`); the excluded admin keeps all `Equals`-based powers and now points `domain.Admin` at a non-member — breaking export/import validation (`genesis_validation.go:606`, same failure class as TRR-01).
- IBC light-client governance recovery is impossible: `ibcKeeper` authority is the unspendable gov module address (`app.go:219,293-301`); expired/frozen clients can only be worked around via permissionless `MsgCreateClient` + new channels.
- DEX `RemoveLiquidity` dust: truncation (`keeper.go:488-489`) can yield a zero one-sided output while shares are still consumed (`custody.go:304`) — reject when either output is zero.
- `GetAllLPPositions` panics on malformed LP key (`custody.go:87`, reached via `ExportGenesis`) — module-written keys only, but an error return fits the fail-closed style better.

---

## Verified strengths (evidence-checked)

- **21M cap:** single issuance boundary (`token/issuance.go`), aggregate clamp shared across both EndBlock inflation phases in one cache (`module.go:265-275`), genesis cap validation pre-mutation, crisis invariant asserting supply ≤ cap **every block** (`token/invariant.go`, check-period 1 at `app.go:267`, crisis last in EndBlock order). No mint path outside `IssuanceService` (maccPerms `app.go:69-75`).
- **Escrow parity** including pending exit holds (`escrow.go:373-388`); every multi-step bank/state mutation in `CacheContext` with validation inside the cache; parity registered as an invariant.
- **Signer-vs-identity:** `requireSignerClaim` (`escrow.go:33-38`) on every identity-bearing message; generated GetSigners for both modules with startup panic on contract violation.
- **Reward/AMM math:** integer/LegacyDec only, no floats; decay clamped [0,1]; swap output rounds down in the pool's favor; exact-ratio liquidity adds kill donation/share-inflation attacks; drain protection; positive `minOutput` enforced on the only live swap path; legacy no-slippage swap disabled; length-prefixed LP keys.
- **Validator lifecycle:** rotation/tombstone/evidence-window machinery unusually thorough; genesis validation cross-checks the entire slashing state machine.
- **Upgrade governance:** bounded plan alphabet, min lead, snapshot electorate, exact 2/3 rule, atomic scheduler writes, halt-by-default binaries with ldflag plan binding.
- **Init/lifecycle:** generated-key PoD bootstrap, operator/consensus-key independence incl. derived-collision checks, 0600 atomic writes, **no hardcoded keys/mnemonics anywhere** (grep-verified), `migration_command.go` bounded reads + constant-time compares.
- **Policies/tooling:** networkpolicy/topologypolicy/incidentpolicy/capacitypolicy read-only and fail-closed; observability sanitizer deep with public-field allowlist; `go.mod` no replace directives; Dockerfile digest-pinned, non-root, loopback port binds.
- **Project review protocol honored:** `go run ./cmd/security-review` **PASS** on the frozen baseline ("security review readiness contract verified; independent review and production claims remain false").
