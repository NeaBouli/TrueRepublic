# Retired Mobile Client

TrueRepublic currently has no supported native mobile client. The former Expo
51 / React Native 0.74 prototype was retired and removed under GH-102; its source
remains available only in Git history for audit purposes.

## Why it was retired

The final clean baseline reproduced all of the following:

- 51 dependency advisories: 7 low, 16 moderate, 24 high, and 4 critical;
- obsolete CosmJS 0.32 cryptography on the mnemonic/signing path;
- an Android bundle failure because Metro could not resolve Node `crypto`;
- two failed Expo Doctor checks for invalid icon configuration and incompatible
  package versions;
- no test files, while CI passed through `--passWithNoTests`;
- plaintext mnemonic entry retained in React component memory; and
- governance/DEX calls built on obsolete query paths, with swap remaining a UI
  stub.

A secure upgrade would have required coordinated major migrations across Expo,
React Native, React, React Navigation, and CosmJS plus a rewrite of key custody,
queries, tests, and chain integration. That is a new high-risk product, not a
dependency patch.

## Replacement requirements

Any future native mobile client needs its own reviewed issue and must, before it
is offered to users:

1. use maintained Expo/React Native and CosmJS versions with a blocking audit;
2. keep mnemonics and private keys out of ordinary UI/application memory and use
   platform-backed secure storage with a documented recovery model;
3. test derivation, signing bytes, fees, broadcast failure, chain ID, address
   prefix, and current gRPC/query compatibility against deterministic fixtures;
4. pass real unit, integration, Android bundle, iOS bundle, and physical-device
   checks without `--passWithNoTests` or advisory bypasses; and
5. receive independent wallet/cryptography review before any release, keys,
   signing, funds, app-store action, or rollout approval.

Until then, `client-web` is the only maintained client. Do not retrieve or run
the retired prototype with real keys or funds.
