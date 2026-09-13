# ZKP Protocol Freeze

- Status: frozen production-candidate specification
- Schema: `truerepublic/zkp-protocol-freeze/v1`
- Tracking: [GH-297](https://github.com/NeaBouli/TrueRepublic/issues/297)

This document records the immutable protocol boundary for the current
anonymous-rating circuit. The machine-readable authority is
[`configs/security/zkp-protocol-freeze.json`](../../configs/security/zkp-protocol-freeze.json).
The freeze earns the protocol-definition item in Phase 2; it does **not** make
the ZKP path production-ready and does not enable anonymous submission.

## Frozen profile

The candidate profile is `truerepublic/anonymous-rating/v2`, consensus version
2, using circuit
`truerepublic/membership-vote/v2-bn254-mimc-depth20`. Public inputs are ordered
exactly as follows:

1. `merkle_root`
2. `nullifier_hash`
3. `external_nullifier`
4. `signal_hash`

Each value is encoded as exactly one 32-byte, big-endian, canonical BN254
scalar-field element. Non-canonical, short, long, reordered or additional
values fail closed.

The one-vote nullifier scope remains `TrueRepublic/vote/v1` and binds the
chain ID, Domain, issue and suggestion. It deliberately excludes the rating
and reward recipient so changing either cannot create another vote. The
canonical `TrueRepublic/vote/v2` signal binds those four fields plus the exact
rating and canonical bech32 reward recipient. Legacy v1 signal payloads remain
synthetic-fixture compatibility material only and are not production signals.

## Version and activation rules

The frozen profile is never edited in place:

| Change | Required action |
|---|---|
| Constraint system, curve, hash construction, Merkle depth, identity commitment, nullifier hash or public-input order | Publish a new circuit ID. |
| Field encoding, vote-context domain/field order, rating encoding or recipient encoding | Publish a new protocol profile. |
| Nullifier domain, bound fields or hash construction | Allocate a new nullifier keyspace and document replay/migration behavior. |
| Any consensus-visible activation | Use fresh genesis or an explicit governed consensus upgrade with deterministic migration evidence. |
| External review requires a semantic change | Publish a new version and migration plan; never rewrite this profile. |

The manifest digest-binds the existing circuit specification. That source
specification and its CS/PK/VK fixtures remain classified
`TEST-ONLY SINGLE-PARTY TOXIC WASTE`; this freeze does not promote them.

## Gates that remain open

- a compatible real Groth16 prover in the maintained client;
- reproducible production constraint-system, proving-key and verification-key
  artifacts;
- ceremony provenance, participant assumptions and artifact rotation policy;
- real browser-to-chain production proof compatibility;
- independent cryptographic, privacy and trusted-setup review.

Until every separate gate passes, `submission_allowed=false`,
`production_ready=false`, and the maintained client's `isSubmittable` guard
must remain hard false. Direct reward payout also remains publicly linkable to
its chosen address; this protocol freeze does not claim shielded payout
privacy, coercion resistance or protection from vote buying.
