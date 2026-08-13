# Known Limitations (v0.4.0 recovery)

## Recovery status (2026-07-11)

The repository is undergoing a security and reproducibility recovery tracked in
[GitHub issue #4](https://github.com/NeaBouli/TrueRepublic/issues/4). It is not
approved for production or real funds during this audit.

- `client-web` is the maintained v0.4 web client and has passed the current
  dependency, lint, test, and production-build recovery checks.
- The former `web-wallet` prototype was retired and removed under GH-112 after
  reproducing 70 advisories, broken query calls, obsolete/unregistered custom
  transaction paths, and a real bank-send path. Git history is audit-only.
- The former `mobile-wallet` prototype was retired and removed under GH-102. It
  had no meaningful tests, could not produce an Android bundle, handled a
  mnemonic in UI memory, and carried high/critical Expo, React Native, Axios,
  protobuf, XML, and obsolete CosmJS dependency exposure. There is currently no
  supported native mobile client.
- Client-side ZKP generation remains a preview, not real Groth16 proof
  generation. The maintained client fails closed and cannot submit mock proofs.
- GH-115 gives the maintained client one exact custom protobuf registry and a
  secret-free local simulate/sign/deliver proof for bank, governance,
  membership, identity, and slippage-bounded DEX flows. Unknown and legacy
  custom messages fail closed. This is local test evidence, not approval for
  real keys, public RPCs, production broadcasts, or funds.
- GH-193 verifies bounded English BIP-39 import errors, versioned AES-GCM with
  PBKDF2-HMAC-SHA-256 at 600,000 iterations, authenticated legacy-payload
  re-encryption, exact derived-address and endpoint chain-ID matching, and
  session-bound signer rejection after lock, switch, delete, or reload. Browser
  localStorage remains exposed to a compromised same-origin runtime; hardware
  custody, extension wallets, real keys/accounts/funds, and production wallet
  operation remain unqualified.

## IBC Compatibility Boundaries

The IBC core, Tendermint light-client types, capability module, and ICS-20
transfer application are real wired modules. GH-175 locally verifies two
isolated TrueRepublic states through client, connection, and unordered channel
handshakes; native PNYX escrow and voucher minting; acknowledgement and timeout
cleanup; replay safety; and database recovery with a pending acknowledgement.
GH-178 additionally proves a committed CLOSED counterparty end survives
database recovery, then drives real proof-verified close-confirm and
timeout-on-close messages, refunds exactly once, rejects the old channel, and
completes transfer/ACK on a distinct replacement channel. ICS-20 intentionally
rejects user-initiated close, so the initial CLOSED end is a deterministic
fixture rather than a claimed production close authority. GH-181 additionally
reopens the same application databases with a separately linked compatible
test binary, preserves the open IBC path plus pending packet/ACK and economic
state, completes the pending ACK, and proves a fresh ACK and timeout refund
exactly once. GH-184 separately wires `x/upgrade` and proves a governed
fresh-genesis `v0.4.1` halt, cached-write failure rollback, fixed-candidate
recovery, common app hashes, and exact-once completion across four validators.
GH-190 adds local maintained-client evidence for canonical native MsgTransfer
encoding, strict open-channel/amount/balance/timeout validation, wallet-scoped
non-secret recovery records, and manual source-chain acknowledgement/timeout
reconciliation. A broadcast or send_packet is never represented as delivery,
and the client never resubmits automatically.
These remain local harnesses; they do not qualify an external relayer release,
public counterparty, pre-GH-184 store introduction, IBC client upgrade, daemon
rollout, or arbitrary cross-version compatibility.

### IBC Staking Compatibility
**Status:** Explicitly unsupported and fail-closed
**Reason:** TrueRepublic uses Proof of Domain (PoD), not `x/staking` PoS
**Impact:** IBC cannot expose delegation or a fabricated unbonding period. The
keeper required by ibc-go returns the stable unsupported-surface error.
**Code:** `ibc_stubs.go - UnsupportedIBCStakingKeeper`

### IBC Upgrade
**Status:** Application upgrade supported for the exact fresh-genesis v0.4.1
path; IBC client upgrade unsupported
**Reason:** GH-184 replaces the IBC upgrade stub with the real `x/upgrade`
keeper and a `truedemocracy` two-thirds governance adapter
**Impact:** Pre-GH-184 chains still need a separate store-loader transition;
IBC client upgrades and arbitrary release plans remain exit-gated
**Code:** `app.go`, `upgrade_handlers.go`, `x/truedemocracy/upgrade_gov.go`

## CosmWasm Compatibility Boundary

### Staking Module
**Status:** Explicitly unsupported and fail-closed
**Reason:** PoD consensus replaces `x/staking`; no standard PoS state is
fabricated from custom validator records
**Impact:** Every standard staking query and message is rejected. Contracts
must use supported TrueRepublic custom interfaces instead of assuming Cosmos
staking semantics.
**Code:** `wasm_stubs.go - UnsupportedWasmStakingKeeper`

### Distribution Module
**Status:** Explicitly unsupported and fail-closed
**Reason:** Custom VoteToEarn and node rewards do not implement
`x/distribution` semantics
**Impact:** Every standard distribution query and message is rejected; empty
or zero-valued responses are never presented as real distribution state.
**Code:** `wasm_stubs.go - UnsupportedWasmDistributionKeeper`

GH-187 regression-tests that `x/staking` and `x/distribution` remain absent
from module basics, runtime modules, stores, genesis, gRPC services, message
routers, and root CLI query/transaction commands. The supported boundary is
auth/bank/consensus params/crisis, TrueRepublic governance and DEX, CosmWasm,
IBC core plus ICS-20 transfer, and the bounded governed `x/upgrade` path
described above. This local boundary evidence is not external-relayer,
counterparty, public-network, IBC client-upgrade, or production approval.

## Production Node Lifecycle

**Status:** Single-node lifecycle is merged; GH-32/GH-41/GH-43/GH-45/GH-53/GH-56 add
bounded four-validator failure/restart/catch-up, partition recovery, trusted
snapshot state-sync, sanitized backup/restore/export/import, compatible binary
replacement, fail-before-open rollback, and authenticated validator-key
rotation/revocation harnesses. GH-61 adds a bounded reviewed-fresh-genesis
legacy-authority transformer and real historical four-validator
halt/export/transform/start/rollback drill, merged through PR #69. Independent
operations and broader migration-security review remain pending.
**Current:** The standard `truerepublicd init --bootstrap-operator` command
binds an independently controlled account authority to the generated CometBFT
Ed25519 consensus key and matching bank-backed positive-power PoD validator.
Same-key, cross-validator, revoked-key, and reserved module-account authority
collisions are rejected. Initialization
rejects canonical supply above the 21,000,000 PNYX cap. A real native process
produces blocks, shuts down on SIGINT, restarts from the same home, advances
height, preserves invariants, and exports state. GH-77 makes the supported
native/container start path emit structured JSON through one defensively
redacting SDK/CometBFT logger boundary; raw KV trace stores are rejected.
GH-80 adds a private two-source CometBFT plus SDK/application metrics baseline
with bounded TrueRepublic collectors, persistent-restart progression, and
disabled-route evidence.
Redaction remains defense in depth rather than general DLP. The non-root
Debian/glibc container has a blocking restart gate. GH-53 additionally proves
compatible rolling replacement on the same homes, deterministic failure before
state is opened, full return to the baseline binary, unchanged identity keys,
non-regressing signer state, app-hash agreement, and ledger-valid export/import.
**Impact:** `scripts/init-node.sh` delegates exclusively to the supported daemon
init boundary and never creates staking gentxs or extra accounts. The Docker
restart job passes. The GH-32/GH-41/GH-43/GH-45 gates prove common-height
app-hash agreement, one-validator failure, continued quorum, restart/catch-up,
partition recovery, trusted snapshot state sync, sanitized backup/restore,
restored export/re-import, compatible binary replacement/rollback, and
single-signer validator-identity cold failover, plus authenticated atomic
rotation with permanent revocation. The GH-61 path is intentionally limited to
empty CosmWasm state and a new chain ID; it is not an in-place upgrade or a
generic governance migration. Do not claim public-network readiness until the
GH-184 path is reproduced on clean private infrastructure and independently
reviewed, and until external paging drills, production sizing and multi-day
soak evidence, real topology deployment, broader independent migration/ABCI++
slashing security review, and independent operations review pass. GH-85 supplies the
repository-owned dashboard, alert rules, recovery/testnet objectives, and role
ownership; it does not configure a production paging destination or claim a
production SLO. GH-89 supplies a strict synthetic multi-node qualification
contract and CI validator; it does not commit the private operator inventory or
deploy firewall, TLS, DNS, seed, sentry, validator, or RPC infrastructure.
GH-97 adds a bounded four-validator, 24-wave/96-transaction regression with
parallel signers, throughput/latency, disk/log growth, RSS, metrics,
snapshot-retention, restart, and ledger evidence plus static finite
log/telemetry configuration checks. It does not prove multi-day soak behavior,
actual Docker/Prometheus retention, production traffic, hardware sizing,
bandwidth, cost, or public topology.
GH-101 adds only an offline, digest-bound deployment-evidence envelope for a
private GH-89 contract. A passing report proves strict structure, contract
binding, freshness, complete gate names, and claimed two-seat approval
separation. It cannot prove that referenced artifacts exist, that observations
were honestly collected, that seats are distinct humans, or that any real
infrastructure is deployed. The Phase 6 production-topology gate remains open.

Partial validator stake withdrawals are disabled until generalized slashable
unbonding can retain the withdrawn claim through the CometBFT evidence window.
Full validator exits remain supported through the evidence-window escrow hold.

## ZKP Client

**Status:** Real synthetic Go/WASM compatibility verified; production ZKP submission disabled
**Timeline:** v0.4.0
**Current:** Proofs bind chain/proposal/rating, nullifiers persist across export,
and the trusted genesis VK is pinned by circuit ID, SHA-256, curve, shape, and
canonical encoding. GH-203 freezes the circuit and encoding contract in
`configs/security/zkp-circuit.json`: fail-closed repository tests cross-check
it against Go constants/behavior, the GH-198 fixture manifest and golden
vector, the active `go.mod` gnark toolchain, and the maintained-client
`zkpEncoding` contract, and prove deterministic byte/hash parity of the
recompiled constraint system only. Committed Groth16 proving/verifying keys are
randomized single-party toxic-waste test artifacts, not production or
reproducible ceremony artifacts; no ceremony is claimed. Anonymous rewards
remain deferred. GH-206's isolated Go/WASM command consumes only those exact
artifacts, generates a fresh proof through the maintained-client adapter, and
proves native Go verification. It is not shipped in the production bundle, has
no wallet/RPC/broadcast capability, and does not change the hard-false
`isSubmittable` guard.
**Future:** Production ceremony artifacts, audited prover/submission integration, and independent circuit review

## Workarounds

### For IBC Staking
Use TrueRepublic's PoD system instead of traditional staking. Do not issue
standard staking/distribution queries or transactions; they intentionally fail
closed.

### For Upgrades
Use only the bounded governed application-upgrade path documented for the
fresh-genesis v0.4.1 baseline. Pre-GH-184 store introduction and IBC client
upgrades remain unsupported.

### For ZKP
Do not submit anonymous votes from the web client. GH-206 is synthetic test-only
compatibility evidence, not a production prover. Use the reviewed domain-key
path without anonymous rewards until ceremony and independent review complete.

## Reporting Issues

If you encounter limitations not listed here:
- Check: https://github.com/NeaBouli/TrueRepublic/issues
- Report: New issue with label `limitation`
