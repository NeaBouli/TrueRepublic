# Security Notes

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
- Historical pre-GH-115 measurement: the v0.4 client production bundle was
  1.70 MB (317.94 kB gzip). The current GH-115 measurement is 318.91 kB gzip;
  route-level code splitting remains recommended before treating
  low-bandwidth/mobile UX as ready.

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
