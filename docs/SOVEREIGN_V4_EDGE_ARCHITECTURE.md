# Sovereign V4 Edge Architecture

**Issue:** [GH-236](https://github.com/NeaBouli/TrueRepublic/issues/236)  
**Status:** implementation-ready target architecture; no runtime implementation  
**Relationship:** evolves the [Sovereign Alpha](SOVEREIGN_ALPHA_ARCHITECTURE.md)
without replacing its Beta-to-Alpha safety gates.

> **Accounting boundary:** this document changes no code, dependency, protocol,
> token, release or deployment. Status remains 2,094 verified cases, rollout
> 35/59, phase work 35/51, Phase 6 at 6/7 and `production_ready: false`.

## 1. Decision

TrueRepublic remains the **only Layer 1, settlement and governance chain**.
V4 adds a local-first edge layer around TRChain: user-operated pruned nodes on
capable computers, mobile verifying participants, authenticated P2P discussion,
encrypted local storage, a non-custodial PNYX wallet and sandboxed domain apps.

V4 does not run on Minima and does not add a foreign chain. It adopts only
general edge-system ideas: software at the participant, asynchronous messaging,
small domain applications and user-operated infrastructure. The external
`NeaBouli/pnyx` project supplies civic-workflow inspiration such
as a real-bill lifecycle and representativeness/divergence views. It does not
replace TRChain consensus, identity or custody, and its centralized stack is
not imported. Its license and any reusable scope must be verified at DG-V4-1
before code is adopted.

“V4” names this future architecture generation. It is not the recovery release
`v0.4.0`.

### Fit and gap

| Option | Fit | Critical gaps | Decision |
|---|---|---|---|
| TRChain plus hosted web only | Reuses all canonical state | Website/RPC availability, no native messaging or device-owned application | Keep only as Beta |
| Separate Minima-style chain | Edge nodes and mini-app model | Second consensus, token and canonical truth; bridge/bootstrap/liquidity risk; governance rewrite | Reject |
| TRChain plus sovereign edge layer | One canonical settlement, local-first UX, user nodes, independent distribution | Requires transport, light verification, sandbox and device qualification | **Adopt** |

## 2. Principles and non-goals

### Principles

1. TRChain is canonical for membership, governance, PNYX and final outcomes.
2. Discussion and drafts are local-first and off-chain; only required hashes
   and authenticated state transitions settle on-chain.
3. The Go core owns secrets, signing, policy and protocol validation. UI and
   mini-apps receive narrow capabilities, never key material.
4. Every verification level is visible. RPC-derived data must not be presented
   as light-client verified.
5. Unknown versions, message types, capabilities and proof states fail closed.
6. `client-web` remains the maintained Beta until each V4 exit gate passes.

### Non-goals

- no Minima dependency, fork or second blockchain;
- no new token, second consensus or foreign settlement authority;
- no custodial wallet or mandatory TrueRepublic-operated website/service;
- no on-chain chat, attachment archive or personal identity database;
- no resurrection of retired `web-wallet` or `mobile-wallet` prototypes;
- no claim that mobile devices are validators or full nodes;
- no production ZKP, legal-validity, receipt-freeness, coercion-resistance or
  vote-buying-prevention claim;
- no committed bridge. Bridge readiness means protocol compatibility only.

## 3. Components and trust boundaries

| Component | Responsibility | Must not do | Failure boundary |
|---|---|---|---|
| TRChain | Domains, membership, ratings, ballots when implemented, PNYX, DEX, final settlement | Store private discussion or device secrets | Edge clients stop settlement if chain identity/finality is unverified |
| Pruned citizen node | User-run TRChain full verification on capable desktop/home hardware | Become a validator without the existing PoD process | Local app may fall back only to explicitly labelled verifying/RPC mode |
| Mobile verifying participant | Verify headers/proofs through the qualified light-client path and use Waku light protocols | Claim full-node/validator status | Unverified data is labelled and signing-sensitive actions fail closed |
| Sovereign Go core | Key separation, wallet, canonical encoding/signing, Waku adapter, sync, policy | Expose mnemonics/private keys or accept caller-declared identity | Locks and rejects all privileged calls |
| Flutter UI | Cross-platform presentation and accessibility | Sign, decrypt stores or interpret unknown protocol data independently | Can be replaced without changing key/protocol ownership |
| Waku transport | Authenticated domain discussion, discovery and asynchronous delivery | Decide membership, votes, balances or results | Withholding creates an explicit gap; chain state remains canonical |
| Encrypted local store | Drafts, message cache, queues, recovery metadata and preferences | Become authoritative governance state | Rebuild from chain and available peers; missing history stays visible |
| PNYX wallet | Non-custodial account and transaction approval | Custody funds remotely or auto-sign mini-app requests | User confirmation plus fail-closed transaction registry |
| Domain mini-app runtime | Domain-bound views and workflows through declared capabilities | Read keys, bypass policy, arbitrary network/filesystem access | Invalid manifest/capability/artifact is refused before execution |

### Mini-app manifest

A proposed canonical manifest contains:

```text
schema: truerepublic.domain-app/v1
app_id, version, artifact_hash
domain_id or reusable-domain policy
capabilities[]
minimum_core_protocol, maximum_core_protocol
publisher_key, signature
```

Capabilities are deny-by-default and mediated by the Go core. Initial PoC
capabilities are read-only discussion, local document storage and registered
chain queries. Transaction proposals require a separate capability, registry
validation and explicit wallet confirmation. No app receives raw keys.

## 4. Protocol and data flows

### Domain discussion

1. The core resolves the authenticated chain account and domain membership at
   a light-client-verified, pinned TRChain height.
2. It creates a domain-scoped messaging key separate from wallet and validator
   keys. The chain account signs a versioned authorization certificate binding
   the messaging public key, chain ID, domain ID, member account, membership
   height, key epoch, validity interval and membership-proof hash.
3. A versioned envelope binds that certificate and proof, chain ID, domain ID,
   topic/bill ID, message ID, parent ID, author key, key epoch, membership
   height, sequence, timestamp, payload hash and message-key signature.
4. Before displaying a message as authenticated, a peer verifies the TRChain
   light proof for membership and account key at the stated height, the account
   signature on the certificate, its domain/height/validity/epoch fields and
   the envelope signature. Missing or unverifiable proof is labelled
   unauthenticated and cannot enter governance or settlement inputs.
5. Rotation creates a higher-epoch certificate and never reuses a retired key.
   Compromise revocation is a committed TRChain revocation sequence; after a
   peer syncs that height it rejects envelopes from the revoked and older epochs.
   An offline peer marks post-expiry or not-yet-revocation-checked messages
   provisional instead of claiming current authorization.
6. Waku transports ciphertext; the local store persists validated envelopes.
7. Duplicates are idempotent. Sequence or parent gaps remain visible and cause
   an incomplete-history state, never a fabricated complete thread.

### Real-bill lifecycle

```text
registered source or local draft
  -> content-addressed version lineage
  -> domain discussion and signed amendments
  -> selected text hash
  -> systemic consensing / stones / ratings
  -> optional GH-232 formal ballot after its gates
  -> canonical TRChain transaction and result
  -> reproducible local archive
```

Every amendment references its parent hash. The settlement transaction binds
the final document hash and immutable policy snapshot. A later edit creates a
new version; it cannot rewrite the settled object.

### Consensing and civic metrics

Existing systemic consensing remains the proposal-development mechanism.
Clients may compute participation, rating-distribution, representativeness and
outcome-divergence views from canonical public/pseudonymous inputs. Formulas,
input height and missing-data state are displayed. These metrics are
informational and never silently become voting weights, eligibility or rewards.

### Optional ballots

V4 consumes the versioned contract in
[GOVERNANCE_BALLOT_ARCHITECTURE.md](GOVERNANCE_BALLOT_ARCHITECTURE.md) only
after GH-232's consensus, migration, privacy, ZKP and legal/process gates.
Systemic consensing remains available independently. A messaging identity or
visible Cosmos signer is never treated as a secret-ballot credential.

### Settlement

The Go core converts a user-approved action into an allowlisted registered
TRChain message, simulates it, displays exact effects, signs under the verified
chain/account context and broadcasts. P2P messages and local aggregates cannot
settle state. Final UI status comes from committed chain evidence, not a peer
acknowledgement.

### Offline and recovery

- outgoing envelopes and transactions have separate queues;
- chat retries are idempotent, but chain transactions are never automatically
  resubmitted after ambiguous broadcast;
- reconnect first verifies chain identity/height, then reconciles committed
  state, then requests bounded message gaps;
- mnemonic recovery restores the chain account only. Messaging/device keys and
  local history require an encrypted, explicitly user-controlled backup;
- revoked/lost device keys remain revocable without rotating wallet or
  validator keys.

## 5. Identity, privacy and Sybil boundary

- Governance Sybil resistance remains authenticated domain membership and the
  existing chain rules; validator admission remains Proof of Domain.
- Wallet, validator, device, messaging, domain and ballot keys are separated.
- Identity attributes and real-world identifiers stay off-chain. A credential
  may prove eligibility without publishing its source identifier only after an
  independently reviewed proof/issuer/revocation design exists.
- Encryption does not hide all peer, timing, topic, message-size or availability
  metadata. Store/relay correlation and traffic analysis remain residual risks.
- Current Groth16 support is synthetic/test-only. Ballot-scoped nullifiers,
  non-voter fee-payer/relayer submission and production ceremony/audit evidence
  are prerequisites for a secret formal ballot.
- Direct reward payouts remain linkable to the public recipient. Fresh
  addresses reduce reuse correlation but do not make payouts shielded.
- Anonymous ballots alone do not prevent coercion or vote buying. Receipt-free
  credentials and operational election controls require a separate design.

## 6. Threat and fail-closed matrix

| Threat | Required control | Fail-closed result |
|---|---|---|
| Malicious mini-app | Signed artifact, sandbox, narrow capabilities, resource quotas | Refuse load/terminate; no key or transaction access |
| Forged/cross-domain message | Domain-separated signature and canonical envelope | Drop before storage/display |
| Replay/equivocation | Stable message ID, author sequence, parent/hash rules | Idempotent duplicate or explicit conflict |
| Malicious/withholding Waku peer | Multiple peers, bounded sync, gap evidence | Show incomplete history; never invent messages |
| Malicious RPC | Pinned chain ID, light proof where qualified, source labelling | No verified-status or signing-sensitive continuation |
| Device/store compromise | OS keystore, encrypted store, lock timeout, key separation | Revoke device; wallet/validator keys remain separate |
| Ambiguous transaction delivery | Hash/sequence tracking and chain query | Submitted-unknown state; no automatic rebroadcast |
| Metadata correlation | Topic design, padding/batching decision gates, retention limits | Honest residual-risk warning |
| Foreign-chain/bridge confusion | TRChain-only settlement and explicit network identity | Reject foreign result as non-canonical |

## 7. Proposed device budgets

These are design ceilings, not measurements. Each must be qualified on real
supported devices before release.

| Profile | Role | Proposed ceiling | Qualification evidence |
|---|---|---|---|
| Mobile | Flutter + Go core + light verification + Waku light client | 512 MiB active RAM, 2 GiB bounded local data, 50 MiB/day typical sync, background/battery budget defined per OS | 24h foreground/background, loss/reconnect and low-storage tests |
| Citizen desktop | App plus pruned non-validator TRChain full node | 4 GiB RAM, 40 GiB bounded/pruned chain+app data | clean sync, pruning, restart, corruption/recovery and month-growth projection |
| Home node | Pruned full node plus optional Waku relay/store | 8 GiB RAM, 80 GiB bounded data, configurable bandwidth/retention | 7-day multi-peer soak, retention enforcement, upgrade/rollback and abuse tests |

Exceeding a ceiling blocks that profile; documentation may not relabel an
unmeasured device as supported.

## 8. Target repository and protocol layout

```text
sovereignv4/
  protocol/                   # V4-0 implemented: schemas, codecs and vectors
  core/                       # proposed; absent today
  app/                        # proposed; absent today
  domain-app-runtime/         # proposed; absent today
  profiles/                   # proposed; absent today
  test/e2e/                   # proposed; absent today
```

The implemented `sovereignv4/protocol` package is deliberately unwired. It uses
only the Go standard library and supplies strict version/type/length decoding,
canonical certificate/envelope/manifest bytes, Ed25519 signatures, caller-
verified membership and publisher facts, epoch/revocation decisions, bounded
authenticated replay detection, golden vectors and adversarial/property/fuzz
tests. It does not verify light proofs, trust registries or execute an app.

The wire schema starts at `truerepublic.edge-envelope/v1`. It binds `chain_id`,
domain/topic identity, canonical byte encoding and signature algorithm. Unknown
major versions are rejected. Additive readers require explicit conformance
vectors; there is no silent dual interpretation. Chain messages continue using
the existing registered type URLs and consensus versions.

Users own local plaintext and backup keys. Peer deletion removes a local copy,
not other peers' copies or immutable chain commitments. Retention is bounded and
configurable; legal or moderation removal is represented by a signed tombstone
without claiming global erasure.

## 9. Delivery slices and exit gates

| Slice | Deliverable | Exit gate | Rollback boundary |
|---|---|---|---|
| V4-0 Protocol | GH-241 certificate/envelope/manifest schemas, Go codec, golden/fuzz/property vectors | Local Race/Coverage, vet, staticcheck and adversarial tests pass; protected cross-platform evidence required before Done | No runtime adoption |
| V4-1 Discussion | Waku adapter, encrypted store, offline/gap/replay behavior | Two-device and restart/loss/reorder/adversarial-peer matrix | Disable transport; Beta unchanged |
| V4-2 Civic lifecycle | Versioned bills, amendments, consensing and local metrics on disposable chain | Hash lineage, settlement and metric reproducibility under missing/conflicting data | Feature flag off; no chain migration |
| V4-3 Device profiles | Mobile verifier and pruned citizen/home nodes | Real-device budgets, sync, pruning, recovery and upgrade tests | Unsupported profile removed |
| V4-4 Domain apps | Signed manifest, sandbox and capability broker | Escape, key-access, network, resource and transaction-confusion tests | Runtime disabled; core remains usable |
| V4-5 Hardening | Recovery, reproducible packaging, security/privacy review | No unresolved critical/high issue; signed supported artifacts and explicit go/no-go | Beta remains maintained |

Every slice requires unit, property/fuzz, integration, restart/recovery,
cross-platform and negative security tests appropriate to its boundary. V4-2
cannot implement formal ballots before GH-232. V4-5 cannot claim production
anonymous voting before the rollout Phase 2 exit gate.

## 10. Decision gates and risks

- **DG-V4-1:** repository Apache-2.0 publication is resolved by
  [GH-219](https://github.com/NeaBouli/TrueRepublic/issues/219#issuecomment-5423337355);
  every third-party component still requires compatibility, provenance, SBOM,
  and distribution review before adoption.
- **DG-V4-2:** qualify Waku licenses, bindings and real-device behavior.
- **DG-V4-3:** choose and prove the mobile light-verification protocol; until
  then RPC trust is explicit.
- **DG-V4-4:** select a sandbox technology only after key isolation and
  capability tests; no arbitrary native plugins.
- **DG-V4-5:** freeze bill/metric schemas and neutral terminology with civic,
  privacy and legal review before they influence formal processes.
- **DG-V4-6:** approve each supported device budget from measured evidence.

Primary risks are metadata correlation, transport availability, sandbox escape,
mobile resource cost, license incompatibility and accidental creation of a
second authority outside TRChain. Each is an exit-gate blocker, not deferred
production debt.

## 11. Bridge readiness

Bridge readiness means that edge envelopes pin network identity, settlement
uses stable registered TRChain messages and exported results are versioned and
verifiable. It permits a future separately reviewed interoperability adapter.
It does **not** create, select, fund, trust or promise a bridge, foreign chain,
relayer or wrapped asset. Existing IBC qualification boundaries remain intact.

## 12. Honest current status

GH-241 implements only the unwired `sovereignv4/protocol` V4-0 library and its
vectors/tests. No Flutter client, Go edge runtime, Waku binding, light client,
citizen-node profile, mini-app sandbox or production ZKP path exists in the
repository today. `client-web` remains the maintained Beta, TRChain remains
recovery-only and no rollout checkbox changes.
