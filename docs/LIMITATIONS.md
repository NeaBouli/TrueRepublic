# Known Limitations (v0.3.0)

> **Recovery warning (2026-07-11):** The project is under an active recovery
> audit and is not approved for production or real funds. The default `main`
> branch is the preserved pre-recovery baseline. Recovery evidence, including
> the enforced 21,000,000 PNYX cap, currently lives in the unmerged PR stack
> tracked by [GH-4](https://github.com/NeaBouli/TrueRepublic/issues/4).
> `web-wallet` and `mobile-wallet` are legacy clients with unresolved dependency
> risk and must not be used with real keys. Use `client-web` only for recovery
> testing until the stack and external reviews are complete.

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
