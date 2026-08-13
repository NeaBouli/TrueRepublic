# Security Notes

## GH-184 governed application-upgrade boundary

- The real `x/upgrade` authority is the non-signing `truedemocracy` module
  account. Only the narrow keeper adapter can schedule or clear a plan after an
  immutable genesis-governance electorate reaches two thirds.
- Plan name, height, info, lead time, duplicate/conflicting votes, late-member
  admission, scheduling failure, cancellation failure, and the genesis-only
  reserved Domain all fail closed with cached atomicity.
- Only `main.upgradePlan=v0.4.1` registers the release handler while `main.version` remains the source commit. The prior binary halts;
  the failure fixture writes a marker then errors; the fixed candidate rejects
  any leaked marker, migrates, and records done height exactly once.
- The evidence is a local four-validator fresh-genesis harness. It does not
  qualify pre-GH-184 store introduction, arbitrary plans, IBC client upgrades,
  external relayers, public networks, production, real keys, or real funds.

## GH-181 compatible IBC binary-restart boundary

- The candidate is a separately linked package test binary built from the same
  source with an asserted version marker. It reopens the exact LevelDB state in
  place without `InitChain`, genesis export/import, height reset, public RPC, or
  a relayer process.
- Before candidate open, recursive hashes cover both closed application
  databases. The intentional exit-42 candidate path reads no manifest or state,
  and the hashes must remain byte-identical before rollback/continuation.
- Clients, OPEN connections, unordered ICS-20 channel ends, sequence state,
  escrow, packet commitment, receipt, and acknowledgement commitment are
  checked after reopen. Pending and fresh ACK plus timeout refund complete once;
  duplicate ACK/timeout relays are economic no-ops and crisis invariants pass.
- This evidence does not qualify `x/upgrade`, governance-controlled halt/resume,
  consensus-breaking migration, IBC client upgrades, arbitrary source changes,
  external relayers, daemon deployment, production, real keys, or real funds.

## GH-178 IBC channel-recovery boundary

- The initial counterparty CLOSED channel end is a deterministic committed test
  fixture because the real ICS-20 callback rejects user-initiated close. It is
  not a production close authority, governance mechanism, or external-relayer
  claim.
- Close-confirm and timeout-on-close are real messages verified against the
  reopened counterparty database's channel-membership and receipt-absence
  proofs. Native escrow refunds once, packet commitment deletion persists, and
  a duplicate timeout-on-close relay is an ibc-go success/no-op without another
  economic effect.
- A distinct channel over the maintained connection receives and acknowledges
  a transfer with separate escrow and voucher denomination while both old ends
  stay CLOSED. External relayers, cross-upgrade recovery, stubs, deployment,
  production, and real keys/funds remain unqualified.

## GH-175 IBC packet and recovery boundary

- Two isolated persistent TrueRepublic apps now prove Tendermint client,
  connection, unordered transfer-channel, native escrow, voucher mint,
  acknowledgement, timeout refund, replay handling, and pending-ack restart
  behavior with real ibc-go state proofs.
- The first proof run found a code-2 decode failure because the concrete
  Tendermint light-client types were absent from the application registry.
  `app.go` now registers them and a standard test prevents regression.
- A duplicate receive with a refreshed counterparty proof reaches ibc-go's
  receipt replay check and cannot mint twice. ibc-go deliberately treats that
  receive and already relayed acknowledgements/timeouts as successful no-ops;
  balance and commitment assertions prove no second economic effect.
- All keys, accounts, databases, and funds are generated disposable local test
  material. The evidence does not qualify an external relayer/counterparty,
  channel closure/replacement, cross-upgrade behavior, governance upgrades,
  deployment, production, or real funds. Production remains false.

## GH-172 concurrency, replay, and restart boundary

- The opt-in process harness uses only generated disposable localhost accounts
  and one ephemeral in-memory signing key. It never reads or emits a mnemonic,
  private key, production endpoint, external account, or real funds.
- Two distinct authenticated transactions contend for one canonical domain and
  prove one successful mutation, one explicit duplicate-domain failure, and
  one escrow effect. Three broadcasts reuse byte-for-byte identical signed
  transaction bytes; duplicate-cache rejection is backed by committed-state,
  balance, custody, object-count, and later account-sequence evidence.
- A clean same-home restart preserves the historical app hash and rejects the
  saved transaction again. Final export, ledger reconciliation, 21M cap check,
  and reimport prove persisted state remains bank-backed and singular.
- This is bounded repository evidence, not a claim about adversarial public
  mempools, Byzantine consensus, production topology, wallet custody, real
  keys/funds, deployment, release approval, or independent external audit.
  `production_ready` remains false and rollout arithmetic is unchanged.

## GH-169 cross-system threat-model boundary

- `configs/security/threat-model.json` is the canonical versioned register. It
  covers consensus/P2P, governance/identity, token/treasury/DEX, ZKP/privacy,
  IBC/upgrades, client/wallet/RPC, operations, dependencies/CI and release
  artifacts with stable IDs, controls, repository evidence, residual risk,
  owners and explicit GH-7/GH-29 gates.
- `production_ready` is structurally pinned false. High/critical residual risks
  cannot be marked closed and must map to GH-7 or GH-29. Evidence must be a
  normalized repository-relative path that exists; malformed/trailing JSON,
  schema drift, unsafe evidence, secret/host shapes and prose/register drift
  fail closed in the root test.
- The model is a repository-grounded self-assessment, not an independent audit.
  It does not claim a real prover/ceremony, wallet-custody review, two-chain IBC
  evidence, private topology, live paging/SLOs, signed release provenance,
  production/mainnet readiness or real-key/real-funds safety.

## GH-148 maintained security-gate boundary

- Every third-party GitHub Action is bound to a reviewed full commit SHA and
  every workflow job has a finite timeout. Go, Node, and Rust toolchains plus
  govulncheck, staticcheck, gitleaks, and cargo-audit have exact versions in one
  repository contract. GH-153 narrows weekly Dependabot version maintenance to
  structurally validated minor/patch groups; ordinary major updates are ignored
  and require separately scoped review. Dependabot security updates remain a
  deliberate separately reviewed exception to that version-maintenance rule.
- Go vulnerabilities, static analysis, maintained-tree secrets, Rust advisories,
  and high/critical maintained-client advisories fail closed. Four secret-scan
  exceptions match exact public synthetic test/documentation strings; broad
  path, commit, and stopword allowlists are forbidden and tested.
- Fresh local evidence finds no secret, no fixable Go vulnerability, no Rust
  vulnerability, and no high/critical npm advisory. Four reachable Go entries
  have no upstream fix and are allowed only by exact IDs with 30-day-bounded
  expiry; unknown, fixable, stale, expired, malformed, or scanner-error cases
  fail. Cargo-audit reports five monitored warning-only transitive advisories;
  neither class is represented as production approval.
- The unchanged maintained client rebuilds inside budget at 76.40 kB gzip
  initial entry, 5.03 kB maximum lazy route, and 353.56 kB total JavaScript
  gzip. Phase-7 image/base-digest pinning, artifact signing, SBOM, provenance,
  publishing, branch-policy administration, and production systems are outside
  GH-148.

## GH-139 critical-path coverage boundary

- Root/application, DEX, and governance statement coverage now have explicit
  fail-closed repository floors of 71.8%, 51.0%, and 63.7%; fresh local totals
  are 71.9%, 51.1%, and 63.8%.
- New regressions exercise atomic rollback/error joining, DEX registry
  authority and bank custody on failed/successful liquidity and swaps, legacy
  slippage-free swap rejection, governance escrow and signer/admin boundaries,
  VoteToEarn settlement, and covered withdrawals.
- This block changes tests, CI, and evidence only. It does not change or claim
  review of consensus, formulas, production configuration, wallet/key/signing,
  deployments, networks, or real funds; independent release-security review
  remains required.

## GH-141 maintained-client Docker build boundary

- The client Docker builder now includes the same maintained scripts used by
  the package build, preserving the enforced bundle budget inside container
  builds rather than bypassing it.
- This repairs build context only. It does not change runtime application code,
  nginx policy, wallet/key/signing behavior, deployment, networks, or funds.
- Local package verification is green; protected CI supplied the authoritative
  image/restart proof because Docker is unavailable in the local environment.

## GH-131 submitted-transaction history boundary

- The maintained client queries only transactions submitted by the validated
  unlocked `truerepublic` address using the Cosmos SDK v0.50 raw CometBFT
  `tx.acc_seq CONTAINS '<address>/'` contract. Incoming-only transfers are not
  indexed or implied by this surface.
- Server pages are newest-first and capped at 50. Envelope, total, transaction,
  message, fee, hash, height, code, and response-count shapes fail closed into
  typed unavailable, timeout, protocol, or decode states. Chain failure logs
  are stripped of control characters and capped at 200 characters.
- Wallet lock/switch/create/import aborts in-flight requests, invalidates their
  generations, and clears prior rows. A history refresh can never change an
  already committed send into a reported send failure. The configured REST
  provider remains trusted for
  completeness; a browser light client or participant index is outside scope.
- The combined production build remains within policy at 76.29 kB gzip initial
  entry, 5.03 kB maximum lazy route, and 353.44 kB total JavaScript gzip.

## GH-132 maintained-browser quality boundary

- The protected Ubuntu matrix pins Playwright 1.55.1 and exercises Chromium,
  Firefox, and WebKit desktop/mobile profiles on `/unlock`, `/create`, and
  `/import`. Serious/critical axe findings, desktop physical-keyboard focus,
  responsive overflow, delayed lazy loading, and third-party requests are
  fail-closed.
- Tests never submit create/import forms, unlock stored material, enter seed
  phrases, sign, broadcast, or mutate a chain. Authenticated wallet flows and
  production custody remain separate rollout gates.
- The production build remains within policy at 75.86 kB gzip initial entry,
  5.03 kB maximum lazy route, and 349.73 kB total JavaScript gzip. The secure
  pinned WebKit matrix is authoritative on protected Linux CI; an unsupported
  frozen local host does not justify a dependency downgrade or gate waiver.

## GH-121 registered browser query boundary

- The maintained browser no longer calls unregistered custom-module REST
  aliases. It sends protobuf requests only to registered gRPC method paths over
  the existing configured CometBFT JSON-RPC endpoint; no new listener or public
  gateway route is introduced.
- RPC/ABCI/protobuf/JSON and nested response-shape failures remain explicit.
  They cannot manufacture empty governance or DEX state, an unused nullifier,
  a default Pay-to-Put price, or a malformed Merkle path.
- Merkle proof queries require canonical 32-byte hex commitments, exact unique
  stored membership, depth-20 MiMC paths, and agreement with the stored root.
  Pay-to-Put uses the canonical treasury calculation and fails closed.
- The browser still trusts the configured RPC provider; a browser light client
  is outside GH-121. Production RPC exposure and ingress policy remain governed
  by the existing rollout gates.
- A new live High npm audit found `nanoid` below 3.3.17 in the lockfile. The
  compatible lock-only resolution is now 3.3.18 and the unchanged fail-closed
  audit policy passes with no live High advisory.
- GH-128 splits all 19 page routes, defers signing/protobuf dependencies, and
  enforces a deterministic build budget. Its exact GH-128 initial entry is 75.79 kB gzip,
  the largest direct lazy route is 5.03 kB gzip, and the complete deferred
  JavaScript set is 349.42 kB total JavaScript gzip. Budgets cover raw and gzip
  entry, route, individual chunk, and total sizes; chunk-import failure reaches
  the existing fail-closed application error boundary.

## Open

- The token/ledger recovery slice through GH-12 is merged to `main` after local
  and GitHub verification. Independent release-security review is still
  required. See `CODEX_AUDIT.md` and GH-7.
- Anonymous legacy rating signatures and Groth16 proofs do not bind a bank
  reward recipient. Direct payout to the transaction sender is front-runnable;
  production handlers therefore defer those rewards pending GH-7.
- GH-20's on-chain ZKP binding passes locally, but a real compatible prover,
  ceremony artifact, and independent circuit review do not yet exist. Both web
  clients fail closed and must remain non-submittable.
- Rust stable CosmWasm 3.0.4 dev-tooling pulls unmaintained/unsound transitive
  crates through Wasmer. No fixable cargo-audit vulnerability remains, but the
  warnings require monitoring or a stable upstream upgrade.
- GH-21 native and GitHub-container single-node lifecycle passes. GH-32/GH-41/
  GH-43/GH-45/GH-53/GH-55 prove bounded four-validator failure/restart/catch-up,
  partition-recovery, trusted state-sync, and sanitized backup/restore/export/
  import, compatible binary replacement/rollback, and single-signer identity
  failover slices. GH-56 proves authenticated atomic consensus-key rotation,
  permanent old-key revocation, separate operator authority, and a stopped old
  signer. GH-59 wires automatic ABCI++ evidence to slashing, and merged GH-60
  proves complete inactive-validator export/import locally and on its final
  GitHub head. GH-71 adds deterministic role-based network policy and safe
  local/container listener defaults. GH-74 adds separate bounded liveness and
  synchronization-aware readiness probes. GH-77 adds secret-minimized
  structured JSON logs through the central SDK/CometBFT logger boundary.
  GH-80 adds a private CometBFT plus SDK/application metrics baseline with
  bounded custom families and fail-closed public-proxy denial. GH-85 adds the
  immutable operations dashboard, eleven deterministic alert rules,
  recovery/testnet objectives, role ownership, and runtime query proof without
  claiming external paging or production SLOs. Consensus-breaking migration
  recovery, IBC, paging drills, load/capacity/topology, and independent
  operations review remain pending;
  IBC staking/upgrade and
  standard CosmWasm staking/distribution stay explicit stubs.
- Historical pre-GH-128 measurement: the v0.4 client shipped one 1.72 MB
  JavaScript entry (322.63 kB by Vite's reporter). GH-128 replaces that
  ambiguous one-file figure with pinned Node-zlib measurements and a 234.32 kB
  raw / 75.79 kB gzip initial entry. Broader authenticated low-bandwidth and browser
  qualification remain open.

## Resolved during recovery

- Updated `golang.org/x/text` from v0.37.0 to v0.39.0 after the 2026-07-22
  vulnerability database exposed reachable GO-2026-5970 through the ZKP
  dependency path. The four remaining reachable Go findings have no available
  fixed version and remain explicitly visible in Security Scan output.
- Updated Go dependencies for fixable `go-getter` and `x/net` advisories.
- Updated Go toolchain away from vulnerable Go 1.24.13.
- Updated v0.4 client dependencies, including CosmJS crypto, Vite, Vitest,
  happy-dom, React Router, Axios transitives, and protobufjs transitives.
- Updated `crossbeam-epoch` and `rustls-webpki` to fixed Rust versions.
- Go 1.26.5 `govulncheck`: no reachable finding with an available fix remains.
- Domain/proposal/stake claims now require authenticated exact bank escrow;
  injected transfer failures and duplicate/spoofed claims are regression-tested.
- CosmWasm stone/election encoders preserve the authenticated contract sender;
  validator slashing burns escrowed PNYX and cannot credit an admin withdrawal.
- Validator/domain inflation and validator slash burns use one canonical,
  cap-checked bank-supply service; both EndBlock inflation phases settle under
  one cache and treasury-funded vote rewards remain supply-neutral.
- DEX reserves are held by the module bank account; create/add/remove/swap
  settlement is cached and atomic, withdrawals require provider-owned LP
  shares, registry messages require chain authority, and swap burns reduce
  canonical PNYX supply. Length-prefixed LP keys prevent valid denom-prefix
  collisions from corrupting conservation totals.
- Custom genesis rejects malformed/unbacked state before mutation, non-empty
  export/import preserves canonical supply and custody, and every-block crisis
  routes halt on cap, escrow, reserve, or LP divergence.
- Removed the GH-12 prototype's publicly derivable bootstrap-validator private
  secret; production code contains no default validator secret.
- ZKP proofs and legacy domain-key signatures bind the chain and exact vote;
  the rating-independent nullifier binds chain/proposal identity. Consensus
  fails closed without a circuit/version/fingerprint-pinned genesis VK.
- Genesis validates canonical BN254 commitments, roots, nullifiers, and public
  inputs, recomputes identity roots, and preserves exact active nullifier state
  without undoing Big Purge semantics.
- Maintained and legacy web clients cannot generate or broadcast mock ZKP
  proofs; focused client regressions assert the fail-closed boundary.
- The MemDB/`select {}` node placeholder is removed. Native block production,
  graceful SIGINT shutdown, same-home restart, height advancement, invariant
  execution, and export pass with the generated validator key; genesis writes
  are atomic and mode `0600`.
- The operator init wrapper no longer creates keyring mnemonics/accounts or
  invokes unavailable `x/staking` gentx commands. It delegates exclusively to
  the generated-key, exact bank-backed PoD daemon init boundary.
- Modernized workflows use read-only permissions and do not persist checkout
  credentials. Maintained-client jobs stay on Node 22; legacy jobs are
  informational and do not convert vulnerable clients into approved targets.

## GH-60 inactive validator genesis boundary

- New exports explicitly distinguish active consensus validators from inactive
  retained custody claims while preserving complete domains and stored power.
- An active claim must be unjailed, carry exact positive stake-derived power,
  and belong to every listed domain. An inactive claim with positive stored
  power must be jailed; malformed flag, jail, power, or domain combinations
  fail before state mutation.
- Revoked consensus keys remain unusable by active and inactive claims, and
  existing rotation/history/signing/infraction/pending-exit relations remain
  mandatory.
- The real four-process recovery drill proves exact bank-backed export/import
  of downtime- and equivocation-disabled validators. This closes GH-60 locally,
  not the broader production rollout or independent operations review.
- PR #67's first Security Scan surfaced new High advisories in PostCSS and
  brace-expansion. Patched PostCSS `8.5.23`, brace-expansion `5.0.8`, and a
  compatible minimatch `10.2.5` transitive path pass clean install, lint, tests,
  build, and the unchanged High-severity npm audit gate. Two Moderate React
  Router advisories remain below that gate pending a dedicated breaking v7
  migration review.

## Legacy client blockers

- Historical GH-112 replacement check: maintained `client-web` passed lint,
  six audit-policy cases, 41 Vitest cases, production build, and guarded live
  audit. GH-115 now records 85 Vitest plus six audit-policy cases.
  The Vite 8.1 build reporter records the current branch bundle as 318.91 kB
  gzip; the existing oversized-bundle rollout item remains open and this task
  does not claim performance closure.

- Former `web-wallet`: retired and removed under GH-112. The final baseline
  reproduced 70 advisories (18 low, 20 moderate, 29 high, 3 critical), broken
  CosmJS custom queries, unregistered/nonexistent custom messages, a
  chain-rejected unsafe legacy swap, and a functional bank-send path.
- The former `mobile-wallet` final baseline reproduced 51 npm advisories
  (7 low, 16 moderate, 24 high, 4 critical), obsolete CosmJS crypto, Expo 51 /
  React Native 0.74, no tests, a broken Android bundle, unsafe mnemonic-in-UI
  handling, and obsolete governance/DEX queries. GH-102 retires and removes it;
  Git history is audit-only and must not be recovered for real keys or funds.
- Both former client prototypes are Git-history-only. A future native or
  alternate client is a separate high-risk implementation requiring secure
  custody, current chain compatibility, real tests/builds, and independent
  wallet/crypto review.

## GH-61 exact source-export binding

- **Resolved locally (GH61-SB-001):** every fresh replacement operator now
  signs one canonical descriptor containing the exact 32-byte SHA-256 of the
  raw source-export bytes. Both direct transform entry points and the offline
  CLI reject a missing, malformed, or mismatched commitment before parsing,
  rewriting, or creating output. A CLI regression proves that a post-signing
  byte mutation leaves no output artifact.
- The exact export digest complements, but does not replace, the independently
  observed source header app hash. No app-hash-to-export cryptographic proof or
  retroactive governance-authorization claim is made.
- Independent Kimi review found no P0/P1/P2 issue in the artifact binding. Sol
  corrected its one in-scope test-attribution observation by canonically
  sorting an added mapping before proving the signed mapping count changed.
- **Open side finding (GH61-SB-002):** the bounded offline transformer still
  amplifies memory and CPU for large valid artifacts. A measured 64 MiB padded
  artifact reached approximately 1.75 GiB maximum resident memory and a
  33.51-second focused transform path. The per-module/per-mapping stale-address
  sweep is also multiplicative. This requires a separate approved hardening
  task and was not changed by the exact-binding remediation.
- **P3 boundary:** stale-address detection in preserved unknown-module JSON is
  a literal canonical-string sweep; alternate textual encodings of the same
  address are outside that sweep. Typed Auth, Bank, DEX, and truedemocracy
  fields remain structurally decoded and reconciled.
- PR #69's first final-head scan surfaced newly published reachable
  GO-2026-6061 in gRPC `v1.79.3`. Upgrading to the scanner-named fixed
  `v1.82.1` removes that finding; current local `govulncheck` reports no
  reachable vulnerability with an available fix. The complete Race/Coverage
  gate and all eight real multi-validator process scenarios pass with the new
  transport dependency graph.
- PR #69 review remediation makes the offline ceremony fail closed around the
  operator procedure: validators are stopped/isolated without changing the
  source validator set, and the exact source export is written privately to a
  same-directory temporary file before an atomic no-overwrite publication.
- A proposed rejection of active application validators when the exported
  Comet validator list is empty was disproved by the real historical process
  drill: canonical Cosmos SDK running-chain exports omit that list and rely on
  validated application state for InitChain validator updates. The transform
  therefore accepts the empty canonical export form but continues exact
  key/power reconciliation whenever a Comet list is present.
- All other review findings are implemented, including independent supply
  snapshotting, complete legacy-address assertions, stderr-safe subprocess
  diagnostics, foreign-prefix and malformed-key regressions, one shared
  traversal of every consensus-key-bearing collection, and a decode-once
  canonical mapping sort. The corrected eight-scenario matrix passes in
  1169.685s; the prior 1200-second cap had only about 30 seconds of margin.
- PR #69 final head passed all 11 GitHub checks and merged as `264ab7c`; GH-61
  is closed with zero unresolved review threads. This closes the bounded
  recovery task, not the production boundary: the path still requires a new
  chain ID and empty CosmWasm state and does not provide an in-place migration,
  retroactive governance authorization, or public-network approval.

## 2026-07-29 - GH-71 role-based network boundary

- Seed, sentry, validator, RPC/API, and private roles now fail closed around
  canonical peers, PEX, discovery seeds, persistent/private/unconditional IDs,
  public P2P intent, and client/metrics listener exposure.
- Validators dial at least two sentries outbound with no public P2P listener,
  PEX, seeds, or inbound capacity. Sentries protect validator IDs through
  private and unconditional lists without requiring a dial-out validator peer.
- RPC, REST, gRPC, gRPC-web, pprof, and metrics remain loopback-only; public
  query traffic terminates at a separately reviewed TLS proxy. Unsafe RPC,
  wildcard CORS, public client listeners, and environment-injected topology
  are rejected.
- The repository does not mutate provider/host firewalls or authorize any real
  topology. The local Compose profile is loopback recovery/development only;
  PR #72 final-head container runtime evidence passed on GitHub. All production
  rollout decisions remain external gates.

## 2026-07-31 - GH-89 cross-node topology qualification

- The versioned contract composes, but does not replace, GH-71: effective node
  homes still require the per-role validator before start.
- Committed topology data is synthetic-only. A real contract correlating
  validator public P2P identity with sentries remains operator-private because
  public correlation would weaken identity hiding even though it is not a
  private key.
- Parser and report boundaries reject ambiguous structure and never reflect a
  trailing value, local contract path, or rejected value in command output.
- The external `internet` principal is reserved; public endpoints are unique
  after IP canonicalization and exclude local/private/special address classes.
- Every relationship and flow is explicit under inbound/outbound deny defaults.
  Encoded or ambiguous routes, unsafe/admin/metrics/debug/peer-dial surfaces,
  and non-empty disabled ingress fail closed.
- Passing this validator does not prove an applied firewall, proxy/TLS/DNS
  configuration, real provider separation, DDoS resistance, capacity, or
  production deployment. Those remain GH-29 rollout gates.

### Merged evidence

- PR #90 merged as `2f2acdd` after final head `cb14829` passed the complete
  Go, recovery, Docker/Compose, docs, dependency/security, static-analysis, and
  review gates with zero unresolved threads.
- Review additionally forced operator-safe contract-file close failures and
  exact API-versus-RPC evidence labels. The repository still contains only the
  synthetic `.invalid` inventory; no production address, node ID, credential,
  firewall rule, DNS name, or private validator correlation was published.

## 2026-08-07 - GH-116 query protocol boundary

- The removed custom ABCI queriers had no authentication or authorization
  role and no maintained consumer. Removing them reduces duplicate parsing and
  routing surface; unknown legacy paths fail closed at the application query
  boundary.
- The retained protobuf gRPC QueryServers are read-only and covered by route-
  registration plus live gRPC-over-ABCI execution tests. No grpc-gateway route
  is registered for either custom module.
- GH-121 records the maintained browser client's unsupported HTTP aliases and
  fail-soft empty-state behavior as an explicit rollout blocker. GH-116 makes
  no production, deployment, key, signing, broadcast, account, fund, or
  anonymous-proof change.

## 2026-08-12 - GH-187 fail-closed protocol compatibility boundary

- wasmd's former default query plugins could return empty or zero-valued
  successful staking/distribution responses through required compatibility
  keepers. GH-187 installs explicit unsupported-request handlers instead.
- Default staking/distribution message encoders are also replaced with stable
  rejection, preventing unsupported standard messages from reaching a later
  missing-handler failure. The ibc-go adapter no longer fabricates a PoS
  unbonding duration.
- Regression tests prove `x/staking` and `x/distribution` remain absent from
  module, store, genesis, gRPC, message-router, and root CLI surfaces. PoD,
  ICS-20, CosmWasm, and governed application-upgrade process gates remain
  green. Dependency upgrades must preserve and re-verify this boundary.

## 2026-08-13 - GH-190 maintained-client IBC transfer boundary

- Canonical MsgTransfer encoding, signer-derived sender identity, strict
  native-denom/channel/amount/fresh-balance/timeout validation, and a 10,000
  upnyx reserve all fail closed before signing.
- Persisted recovery records are chain-and-wallet scoped and contain no
  password, mnemonic, signer, key, or client. Wallet lock, switch, deletion,
  and stale async completion invalidate active session state.
- Broadcast is never equated with delivery. Only exact source-chain
  send_packet, acknowledgement, and timeout events advance status; malformed
  or contradictory evidence becomes unknown, and no path auto-resubmits.
- The verified production build remains inside policy at 70.05 kB gzip initial
  entry, 4.91 kB maximum lazy route, and 353.53 kB total JavaScript.
- This is repository/local-chain evidence only. External relayers,
  counterparties, public RPCs, real keys/accounts/funds, release, deployment,
  and production remain outside GH-190.

## 2026-08-13 - GH-193 maintained-client wallet and signing boundary

- Imports accept only normalized 12/24-word English BIP-39 phrases with bounded
  wordlist/checksum errors; service callers cannot bypass wallet-name or minimum
  encryption-password validation.
- New local custody payloads use versioned AES-256-GCM with unique 16-byte salts,
  12-byte IVs and PBKDF2-HMAC-SHA-256 at 600,000 iterations. Authenticated legacy
  100,000-iteration payloads remain readable and are immediately re-encrypted
  after successful unlock. This follows current OWASP PBKDF2 guidance and W3C's
  recommended authenticated AES-GCM boundary.
- Decrypted mnemonics must re-derive the exact selected 20-byte account. The
  canonical signing client rejects empty/malformed/foreign-prefix accounts and
  RPC chain-ID mismatches before signing. Native sends validate recipient,
  positive amount and denom before network work.
- Session-generation-bound signer proxies re-check before and after account reads
  and `signDirect`; lock, switch, active-wallet delete, reload, and stale unlock
  completion invalidate them. Password/current wallet/signer/mnemonic remain
  excluded from persisted Zustand, history and IBC records.
- PASS so far: 10 Node + 295 Vitest cases, lint, TypeScript production build,
  20 lazy routes, 71.13 kB gzip entry, 4.91 kB maximum route, and 355.08 kB total JavaScript;
  the three-case disposable local-chain gate also passes.
- Residual boundary: JavaScript cannot guarantee string-memory erasure, and an
  XSS/compromised same-origin runtime can access an unlocked session. Hardware
  custody, extension wallets, real keys/accounts/funds, public RPC, production,
  release and deployment remain unqualified.

### GH-193 final review closure

- Kimi's independent review found a concurrent encrypted-record persistence
  race. The final implementation performs the synchronous storage read/check/
  merge/write only after asynchronous encryption and proves two concurrent
  distinct saves are retained.
- Balance refreshes are generation/address/lock scoped, so a response started
  before lock or switch cannot restore stale balances afterward. Malformed JSON
  also fails with a bounded message without echoing stored content.
- Final maintained-client evidence is 10 Node + 298 Vitest cases, lint and
  production build/budgets. Final bundle: 355.07 kB total JavaScript. No known
  P0/P1/P2 remains.
