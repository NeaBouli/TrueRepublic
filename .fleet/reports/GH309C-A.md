id: GH309C-A
status: ok
worker: claude (interactive stand-in; mapping only, no product code, nothing executed, PlantUML not installed so not rendered)
branch: docs/GH309C-A-identity-activation-map @ c2cc41d, based on docs/GH-327-architecture-execution-map 1580c8e; MAP.md section 9 + one appended diagram each in map.puml and main-path.puml; not pushed
hops (payload; status):
  - C1 identityStore/IdentityVault custody (#338-#340): {secret, commitment, nullifier, createdAt} hex, preview layout; built, holds preview identities only.
  - C2 ZKPService.generateIdentity -> previewMockIdentityHash: FNV commitment/"nullifier"; quarantined preview, unreachable from UI since #340.
  - C3 canonical generation -> zkpEncoding.ts::mimcBn254([secret]): gnark BN254 MiMC; primitive built and Go-vector tested (zkpEncoding.test.ts); canonical secret sampling (< BN254_SCALAR_MODULUS) and its caller missing.
  - C4 MembershipService.registerIdentity -> txRegistry MsgRegisterIdentity -> deliverMessages: {sender, domain_name, commitment 64 hex}; codec built, service fail-closed (#337).
  - C5 ValidateBasic -> msgServer.RegisterIdentity (msg_server.go:743) -> Keeper.RegisterIdentityCommitment (anonymity.go:273): member check, canonical field element, dedupe, IdentityCommits append, Merkle root + history; built; chain cannot verify the preimage.
  - C6 ZKPService.getMerkleProof -> Query/MerkleProof: built.
  - C7 generateProof -> zkpWasmProver (circuit truerepublic/membership-vote/v2-bn254-mimc-depth20; leaf must equal mimcBn254([secret])): adapter built; real Groth16 prover artifact absent -> throws; isSubmittable false.
  - C8 MsgRateWithProof -> msgServer.RateWithProof (msg_server.go:760) -> VerifyMembershipProofForSignal (zkp.go:144): verifier built; fails closed without genesis VK (EnsureVerifyingKey).
canonical_primitive: yes — mimcBn254([secret]) is the single canonical commitment primitive, already used by the prover adapter for the Merkle leaf; GH309C must activate it, not add a second hash, wrapper, flag or registration path. Per-vote nullifier = MiMC(secret, externalNullifier) (zkpcircuit/circuit.go:59-64); the stored per-identity preview "nullifier" has no canonical meaning.
compatibility_boundary: historical preview identities (FNV commitments; secrets may exceed the BN254 modulus) are never registered, proven or promoted; they stay preview-only in the vault and remain exportable. Blocking gap: vault/Identity records carry no identity kind, so a canonical identity in the same shape would be indistinguishable from a preview one.
stages (smallest safe, each separately reviewed):
  - S1 identity kind: versioned custody payload {kind: "preview-v0" | "canonical-v1"} with preview records read as preview-v0 and never upgraded; files identityVault.ts, identityMigration.ts (kind tagging only), identityStore.ts, types/zkp.ts + tests. No generation yet.
  - S2 canonical generation: sample a uniform field element (rejection sampling via crypto.getRandomValues), commitment = mimcBn254([secret]), no stored nullifier; files new services/canonicalIdentity.ts (+test with Go vector), identityStore create action guarded by S1 kind; UI creation stays disabled.
  - S3 registration enablement: MembershipService.registerIdentity accepts only canonical-v1 identities (kind + recompute mimcBn254 before signing), one explicit confirmation; files membership.ts, OnboardingFlow.tsx, their tests, browser contract.
  - S4 truthful wording: UI states for preview (not registrable), canonical-unregistered, registered, and "anonymous voting unavailable" while C7/C8 gates are missing; files IdentityCustodyNotice.tsx, OnboardingFlow.tsx, VotingPanel.tsx + visual gate.
tests (adversarial): Go/client MiMC vector parity; secret >= modulus rejected; preview-v0 record refused by register (kind and commitment recompute); swapped kind tag in storage rejected; commitment recomputation mismatch before signing; duplicate registration error surfaced; no broadcast for preview identities; RPC allowlist for MsgRegisterIdentity only after explicit confirm; visual states preview/canonical-unregistered/registered/voting-unavailable at 1440x1000, 1180x820, 820x1180, 390x844.
blockers: Groth16 prover artifact (C7) and VK ceremony output (C8) absent; independent cryptographic review required before any production anonymous-voting claim (CLAUDE.md ZKP boundary). S2/S3 can ship registration of canonical commitments without voting, but that exposes on-chain commitments without a usable proof path.
decisions:
  - Gio: whether canonical identity creation and on-chain registration (S2/S3) should ship before the prover/VK exist (commitments become public and permanent per domain), or wait until C7/C8 are ready.
  - Codex: S1 schema (vault record v2 vs identity payload kind field) and whether stored preview "nullifier" is kept only for preview-v0.
next: Codex review of the map and stages; then an S1 brief.
