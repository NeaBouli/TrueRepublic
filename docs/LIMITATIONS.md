# Known Limitations (v0.4.0 recovery)

## Recovery status (2026-07-11)

The repository is undergoing a security and reproducibility recovery tracked in
[GitHub issue #4](https://github.com/NeaBouli/TrueRepublic/issues/4). It is not
approved for production or real funds during this audit.

- `client-web` is the maintained v0.4 web client and has passed the current
  dependency, lint, test, and production-build recovery checks.
- `web-wallet` is legacy and still uses obsolete CosmJS cryptography and Create
  React App. Its mock ZKP submission path is disabled and its focused tests,
  build, and current npm audit pass, but it is not approved for real keys.
- `mobile-wallet` is legacy and currently has unresolved high/critical Expo,
  React Native, Axios, protobuf, XML, and CosmJS dependency advisories.
- Client-side ZKP generation remains a mock, not real Groth16 proof generation.
  Both web clients now fail closed and cannot submit mock proofs.

## IBC Modules (Stubbed)

### IBC Staking
**Status:** Stubbed
**Reason:** TrueRepublic uses Proof of Democracy (PoD), not traditional PoS
**Impact:** Cannot delegate to validators via IBC
**Code:** `ibc_stubs.go - IBCStakingKeeper`

### IBC Upgrade
**Status:** Stubbed (No-op)
**Reason:** x/upgrade module not integrated
**Impact:** IBC upgrades must be handled manually
**Code:** `ibc_stubs.go - IBCUpgradeKeeper`

## CosmWasm Stubs

### Staking Module
**Status:** Stubbed
**Reason:** PoD consensus instead of PoS
**Impact:** Contracts cannot query validator info
**Code:** `wasm_stubs.go - WasmStakingKeeper`

### Distribution Module
**Status:** Stubbed
**Reason:** Custom reward system (VoteToEarn, NodeReward)
**Impact:** Contracts cannot query standard distribution
**Code:** `wasm_stubs.go - WasmDistributionKeeper`

## Production Node Bootstrap

**Status:** Recovery-blocked by GH-21
**Reason:** PR #19 removes the insecure hard-coded bootstrap validator secret
and accepts only real Ed25519 public keys supplied by CometBFT genesis. The
legacy `scripts/init-node.sh` still invokes `x/staking` gentx commands even
though TrueRepublic uses PoD and does not wire `x/staking`.
**Impact:** Do not use the legacy initialization script for production. A
PoD-aware command must bind the node's real Ed25519
`priv_validator_key.json` public key to an actual positive-power CometBFT
validator with sufficient, exact bank-backed custom stake before launch. The
resulting canonical supply must remain within the 21,000,000 PNYX cap.

## ZKP Client

**Status:** On-chain binding recovery-verified on stacked PR #22; client disabled
**Timeline:** v0.4.0
**Current:** Proofs bind chain/proposal/rating, nullifiers persist across export,
and the trusted genesis VK is pinned by circuit ID, SHA-256, curve, shape, and
canonical encoding. Anonymous rewards remain deferred.
**Future:** Compatible real prover, ceremony artifacts, and independent circuit review

## Workarounds

### For IBC Staking
Use TrueRepublic's PoD system instead of traditional staking.

### For Upgrades
Manual chain halt + restart with new binary.

### For ZKP
Do not submit anonymous votes from either web client. Use the reviewed
domain-key path without anonymous rewards, or wait for a compatible real prover.

## Reporting Issues

If you encounter limitations not listed here:
- Check: https://github.com/NeaBouli/TrueRepublic/issues
- Report: New issue with label `limitation`
