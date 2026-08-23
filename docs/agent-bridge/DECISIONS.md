# Decisions

## 2026-08-23 - V4 uses one native settlement chain and a hybrid edge layer

- TRChain remains the only L1, settlement, governance, PNYX and final-result
  authority. V4 evolves Sovereign Alpha above that boundary.
- General user-node, edge, small-app and asynchronous-messaging ideas may be
  adopted, but Minima code, dependencies, a second chain/token and bridge
  commitment are rejected.
- Mobile is a light/verifying participant, never falsely a full validator.
  Capable desktop/home profiles may run pruned non-validator full nodes.
- Ekklesia/pnyx bill lifecycles and civic metrics are workflow inspiration only;
  they do not replace consensus, identity or custody.
- `client-web` remains Beta. V4-0..V4-5 are reversible future slices and earn
  no rollout credit until separately implemented and qualified.

## 2026-08-22 - Optional domain ballots are post-rollout and opt-in

- Systemic consensing remains the proposal-development foundation. Formal
  ballots are optional per domain: systemic, yes/no/abstain, person-election or
  two-stage consensing-to-ratification.
- Domain defaults seed only future ballots. Each ballot freezes policy,
  options/proposal hashes and electorate; deterministic math, explicit
  abstention/denominator/tie/runoff rules and export/import safety are required.
- Candidate publicity and voter privacy are separate. Identity data stays
  off-chain. Public, pseudonymous, secret-ZK and future sealed-secret profiles
  are explicit and cannot silently downgrade.
- A ZK proof from the voter's visible Cosmos signer is not a secret ballot.
  Secret-ZK needs proof authorization, non-voter relayer/fee-payer submission
  and metadata analysis. No receipt-freeness, coercion-resistance, legal-validity
  or production-ZKP claim exists without separate qualification.
- GH-231 documents this contract and earns no rollout credit. GH-232 remains
  deferred behind rollout, consensus, cryptographic, privacy and legal gates.

## 2026-08-09 - Canonical versioned cross-system threat register

- Security risk state is canonical in `configs/security/threat-model.json`;
  `docs/security/THREAT_MODEL.md` explains it and must remain bidirectionally
  consistent. The initial schema is `truerepublic.threat-model/v1`.
- Severity describes inherent impact while `residual_severity` describes the
  remaining risk after verified controls. Mitigated/not-applicable threats may
  not retain high/critical residual risk. Every other high/critical residual
  maps to GH-7 (independent review) or GH-29 (rollout engineering).
- Only existing repository paths count as verified evidence. External/live
  claims remain deferred or blocked; the register is not an audit and cannot
  set `production_ready=true`.

## 2026-08-08 - Maintained browser module-query transport

- The maintained browser uses registered protobuf gRPC query paths through the
  already configured CometBFT JSON-RPC `abci_query` endpoint. It does not
  revive the removed `custom/...` shim, invent REST aliases, add a public
  listener, or advertise unregistered grpc-gateway routes.
- One client boundary encodes protobuf requests, decodes the module's typed
  `Result` envelope, validates runtime response structure, and classifies
  transport, remote, protocol, and decode failures explicitly. Failure is
  never converted into an authoritative empty list, false nullifier, or
  fabricated economic value.
- Merkle proof and Pay-to-Put are canonical registered read methods. Merkle
  responses bind the stored domain root and exact commitment and return the
  depth-20 MiMC path; Pay-to-Put calls the chain's existing treasury formula.
- This browser transport trusts the configured RPC endpoint. Browser-side
  light-client verification, production RPC deployment, public ingress, and
  rollout approval remain separate tasks.

## 2026-07-11 - Recovery baseline

- Current GitHub `main`, not the divergent old local checkout, is the source baseline.
- The old local checkout is preserved and selectively reconciled; it is not reset.
- Recovery work happens on `fix/GH-4-recovery-foundation` in an isolated worktree.

## 2026-07-11 - PNYX maximum supply

- Maximum supply is **21,000,000 whole PNYX**. GH-11 enforces the
  `21,000,000,000,000 upnyx` bank-genesis boundary and GH-13 enforces the same
  canonical bank-supply cap for recovered runtime reward issuance.
- GH-10 routes DEX burns through the canonical issuance service. The
  independent runtime supply/custody invariant remains pending in GH-12; no
  custom module creates a second supply source.

## 2026-07-11 - Status publication

- Public project status is evidence-based: no feature, test count, security
  state, or release completeness claim may exceed verified code and CI results.

## 2026-07-11 - Validator slash custody

- Slashed validator PNYX is burned from the `truedemocracy` module escrow.
- It must not be credited to an admin-withdrawable domain treasury because the
  whitepaper removes the penalty from circulation and the treasury path would
  allow validator/admin collusion to recover it.

## 2026-07-11 - Canonical reward issuance

- `x/bank` `upnyx` supply is the only release-decay and cap source of truth;
  `pod:total-release` is retired from consensus logic.
- `token.IssuanceService` is the governance module's only reward/slash supply
  boundary. Minting is clipped to remaining capacity in a cached context.
- Validator rewards have deterministic priority over domain interest when both
  compete for final cap capacity; allocation within each category follows
  deterministic store-key order.
- Domain interest uses payouts since the prior interval snapshot, not the same
  cumulative historical payouts repeatedly.

## 2026-07-12 - DEX custody and LP ownership

- The `dex` module account is the sole bank custodian for every pool reserve.
- Public create/add/remove/swap transitions use a cached all-or-nothing bank
  settlement and must pass reserve and LP conservation before commit.
- LP ownership is indexed by pool and authenticated provider; global pool
  shares are not transferable withdrawal authority.
- PNYX output burns reduce both pool reserves and canonical bank supply through
  `token.IssuanceService`.
- Asset registry/status mutation requires the configured chain authority.

## 2026-07-12 - Safe consensus genesis

- Production defaults contain no validator private secret or fixed validator
  identity.
- When custom PoD genesis is empty, InitChain accepts only real positive-power
  Ed25519 validators supplied by CometBFT and creates exact bank-backed minimum
  stake for those public keys within the 21M cap.
- Explicit custom treasury/stake/DEX claims must equal the complete module bank
  balances before any custom state mutation.
- Supply, governance escrow, DEX reserves, and provider LP totals are registered
  `x/crisis` routes checked every block.
- `truerepublicd init` is the only valid PoD bootstrap boundary. The GH-26
  wrapper delegates exclusively to it and cannot create staking gentxs,
  keyring mnemonics, extra genesis accounts, or additional token supply.

## 2026-07-12 - Persistent PoD node bootstrap

- Superseded by the 2026-07-23 GH-56 decision below. The generated CometBFT
  Ed25519 key remains the consensus identity, but it is no longer permitted to
  define the operator authority.
- Bootstrap stake is created only as exact cap-checked `x/bank` module backing
  for the matching custom validator. Conflicting existing consensus sets are
  rejected rather than silently replaced.
- Standard Cosmos server lifecycle, persistent database/home, signal shutdown,
  restart, and export are the supported node path. The old MemDB/`select {}` and
  `x/staking` gentx paths are retired.
- Single-node success does not prove multi-node, IBC upgrade, relayer, backup,
  or restore readiness; those require separate operations evidence.

## 2026-07-23 - Independent validator operator authority and key rotation

- Every genesis validator binds an explicit account operator independently
  from its CometBFT consensus key. Same-validator, cross-validator, active-key,
  revoked-key, and reserved module-account collisions fail closed.
- `truerepublicd init --bootstrap-operator` records only the public operator
  account and creates no private key, mnemonic, gentx, or liquid allocation.
- An authenticated rotation is conditional on an active, non-jailed,
  positive-power validator, binds the signed request to the expected old key,
  preserves claims, schedules the new key for H+2, and permanently revokes the
  old key. Reuse fails across export/import.
- Pre-GH-56 homes do not gain this separation through a binary replacement.
  The supported prelaunch transition is a reviewed fresh genesis; no in-place
  authority migration is claimed.

## 2026-07-12 - ZKP circuit and nullifier trust boundary

- Anonymous vote proofs bind a versioned length-prefixed chain ID, domain,
  issue, suggestion, and exact rating signal.
- The one-vote nullifier excludes rating but includes chain and proposal
  identity, so changing a rating cannot create another voting scope.
- Consensus never performs randomized Groth16 setup. Genesis is the ceremony
  trust anchor and pins the expected circuit ID, VK SHA-256, BN254 curve,
  four-public-input shape, and canonical serialized bytes.
- Genesis recomputes identity Merkle roots and exports the exact active
  nullifier records. Historical ratings are not used to resurrect nullifiers
  intentionally cleared by a Big Purge.
- Mock proof generation is not a degraded transaction mode. Both web clients
  fail closed until a compatible real prover is shipped and reviewed.
- Anonymous rewards remain deferred until the proof or a separate claim binds
  a safe recipient without destroying vote privacy.

## 2026-07-12 - Documentation authority and CI runtimes

- `docs/status.json` is the machine source for the current version, recovery
  test totals/module split, technology versions, 21M cap, and feature limits.
- README, CLAUDE, landing page, and real wiki Home/current/testing pages must
  contain the current version and total; CI also proves suite/module sums and
  decimal-to-base-unit cap arithmetic.
- Historical milestone counts may remain only when explicitly labeled
  historical, never as current security or production evidence.
- GitHub workflows use current official Action majors with `contents: read` and
  non-persisted checkout credentials. Project Node/Go versions remain explicit
  and separate from the Actions embedded runtime.
- Feature branches run through pull requests or manual dispatch; routine push
  automation is main-only to avoid duplicate evidence.

## 2026-07-14 - Multi-validator recovery boundary

- Shared validator genesis contains public CometBFT identities only. Every
  private validator key remains in its independently generated node home.
- The single-node `truerepublicd init` command continues to refuse replacing an
  existing consensus set. Multi-validator assembly reuses an internal audited
  public-identity function without weakening that operator boundary.
- Loopback address-book relaxation, duplicate-IP permission, and disabled
  pprof apply only to temporary localhost harness configuration. Production
  CometBFT defaults remain strict.
- Four-validator failure/restart/catch-up and app-hash agreement close one
  Phase 1 checklist item, not multi-node or public-network readiness.

## 2026-07-15 - Codex subagent role split

- The primary Codex agent remains responsible for architecture, security/risk
  decisions, final verification, Bridge updates, GitHub issues, PRs, merges, and
  public status claims.
- Project-scoped `.codex` configuration defines `spark_worker` as a narrow
  `gpt-5.3-codex-spark` worker for small bounded patches, file search, and
  focused checks.
- Subagent recursion stays capped at one level (`agents.max_depth = 1`) and
  concurrency at six open threads (`agents.max_threads = 6`) to keep token use
  predictable while allowing targeted parallel help.

## 2026-07-26 - Sol, Kimi K3, and helper role split

- Codex Sol remains accountable for scope, architecture, security, integration,
  all external writes, complete verification, GitHub coordination, and closure.
- Kimi K3 is the preferred senior partner for larger bounded implementations,
  difficult bugs, repo-wide analysis, architecture alternatives, and
  independent deep reviews. It receives only secret-free repository scope and
  cannot delegate.
- Kimi implementation output is never self-approving: Sol reviews the complete
  diff and reruns the relevant full test and integration chain.
- Claude Code and `spark_worker` remain limited to small, clearly bounded
  helper work. Delegation never grants production, deployment, IAM, payment,
  authentication, migration, publishing, deletion, or secret access.

## 2026-08-06 - Retire the legacy native mobile prototype

- GH-102 retires and removes the former `mobile-wallet` source and its cosmetic
  pass-with-no-tests CI. Git history remains the audit record; it is not a
  supported build or migration source and must not be used with real keys.
- The final baseline reproduced 51 advisories (7 low, 16 moderate, 24 high,
  4 critical), no tests, two Expo Doctor failures, and an Android bundle failure
  at CosmJS' Node `crypto` import. The code retained a mnemonic in ordinary UI
  state and contained a real signing/broadcast path, while governance/DEX used
  obsolete query paths and the swap action was a stub.
- Clearing the material findings would require coordinated major migrations
  across Expo, React Native, React, React Navigation, and CosmJS plus new secure
  key custody, chain queries, tests, and physical-device evidence. That is a new
  high-risk product, not a safe dependency upgrade.
- Any replacement native client requires a separate approved issue, maintained
  dependencies, platform-backed key custody, deterministic signing/chain tests,
  blocking audit/build gates, physical-device evidence, and independent wallet/
  cryptography review before keys, signing, funds, app-store, or rollout use.

## 2026-08-06 - Retire the duplicate legacy web client

- GH-112 removes `web-wallet` instead of migrating it into a second canonical
  browser client. Git history remains the audit record; `client-web` is the
  only maintained implementation and now owns the Compose/nginx frontend path.
- The clean baseline reproduced 70 advisories (18 low, 20 moderate, 29 high,
  3 critical). Its 18 component tests and production build passed, but no test
  exercised the real query/signing boundary: `queryAbci` was called on the
  wrong CosmJS client surface, custom messages were unregistered or nonexistent,
  the legacy DEX swap is rejected on-chain, and bank send could still broadcast.
- A safe migration would require a CRA/Vite and CosmJS major rewrite plus new
  query, registry, transaction, DEX, security, and integration tests, duplicating
  the maintained client. Retirement is therefore the smaller secure boundary.
- `scripts/check-web-wallet-retirement.sh` blocks source, runtime wiring,
  non-blocking legacy audit, status, and operational install regressions.
- Removing or time-bounding the now-unused chain-side legacy `custom/...`
  compatibility shim is tracked in GH-116. Maintained-client custom transaction
  registration and client-to-chain delivery are tracked explicitly in GH-115;
  neither is silently folded into GH-112.

## 2026-08-07 - Retire the consumerless custom ABCI query shim

- GH-116 removes the application-level `custom/truedemocracy/...` and
  `custom/dex/...` routing override and both module queriers. No maintained
  consumer uses this legacy protocol, and retired paths now fail closed.
- The supported custom-module query contract is the generated CLI and
  protobuf gRPC Query service: seven governance methods and ten DEX methods,
  all registered and exercised through the application query boundary.
- The modules intentionally register no grpc-gateway handlers. Documentation
  must not imply that `/truerepublic/...` HTTP aliases are supported.
- The maintained browser client still calls those unregistered aliases. That
  pre-existing fail-soft compatibility defect is isolated in GH-121 and must
  be resolved before any rollout claim; it does not justify retaining an
  unrelated legacy `custom/...` shim.

## 2026-08-16 - Sovereign Alpha architecture direction

- `client-web` remains the supported Beta. GH-215 is a parallel architecture
  track and earns no recovery-rollout checkbox.
- The Alpha target is a native-compiled Flutter UI for Android, iOS and desktop
  over a Go sovereign core exposed through a narrow generated/FFI boundary.
  The core exclusively owns keys, signing, group cryptography, envelope
  validation and encrypted local storage; Beta TypeScript behavior becomes
  golden compatibility evidence, not an Alpha runtime dependency.
- Waku is the conditional messaging transport behind `MessagingPort`, subject
  to an A0 real-device, lifecycle, resource, license/SBOM and reproducibility
  qualification. Status is a reference, not a fork. Telegram/TDLib are rejected
  because open clients do not provide a sovereign backend; Matrix remains only
  an explicit homeserver-trust fallback.
- RFC 9420 MLS is the target for domain group encryption. No custom group
  cryptography, private-domain confidentiality, forward secrecy or
  post-compromise-security claim exists before implementation qualification and
  independent cryptographic review.
- Exit sovereignty means no mandatory single TrueRepublic-operated service for
  custody, messaging, governance or normal operation. Bootstrap, RPC, store,
  distribution and OS-push infrastructure remains replaceable but risky.
- BIP-39 restores the chain account, not historical chat epochs. Chat recovery
  requires an encrypted user-held backup or transfer from an enrolled device.
- Third-party implementation is blocked until the owner publishes a root
  project license and exact per-component compatibility review passes. This
  architecture does not choose a license for the owner.
