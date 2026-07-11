# Known Limitations (v0.3.0)

## Recovery status (2026-07-11)

The repository is undergoing a security and reproducibility recovery tracked in
[GitHub issue #4](https://github.com/NeaBouli/TrueRepublic/issues/4). It is not
approved for production or real funds during this audit.

- `client-web` is the maintained v0.4 web client and has passed the current
  dependency, lint, test, and production-build recovery checks.
- `web-wallet` is legacy and currently has unresolved high/critical dependency
  advisories, including obsolete CosmJS cryptography.
- `mobile-wallet` is legacy and currently has unresolved high/critical Expo,
  React Native, Axios, protobuf, XML, and CosmJS dependency advisories.
- Client-side ZKP generation remains a SHA-256 mock, not real Groth16 proof generation.

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

## ZKP Client

**Status:** Architecture complete, client integration pending
**Timeline:** v0.4.0
**Current:** Domain Key Pairs provide voting privacy
**Future:** gnark-wasm client-side proof generation

## Workarounds

### For IBC Staking
Use TrueRepublic's PoD system instead of traditional staking.

### For Upgrades
Manual chain halt + restart with new binary.

### For ZKP
Use Domain Keys (current) or wait for v0.4.0 client.

## Reporting Issues

If you encounter limitations not listed here:
- Check: https://github.com/NeaBouli/TrueRepublic/issues
- Report: New issue with label `limitation`
