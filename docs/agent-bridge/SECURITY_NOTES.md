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
  GH-43/GH-45 prove bounded four-validator failure/restart/catch-up,
  partition-recovery, trusted state-sync, and sanitized backup/restore/export/
  import slices without shared private material. Upgrades, rollback,
  validator-key compromise response, network policy, IBC, load/topology, and
  independent operations review remain pending; IBC staking/upgrade and
  standard CosmWasm staking/distribution stay explicit stubs.
- The v0.4 client production bundle is 1.68 MB (309 kB gzip); route-level code
  splitting is recommended before treating low-bandwidth/mobile UX as ready.
- PR #25 targets the unrecovered old main and its Go/Rust security gates fail.
  Do not weaken or bypass those checks to publish default-branch status; merge
  the reviewed foundation first, then rebase or replace the visibility track.

## Resolved during recovery

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

## Legacy client blockers

- `web-wallet`: obsolete CosmJS crypto, legacy Create React App toolchain, and
  extensive source-map warnings. Current npm audit is clear, but the dependency
  architecture is not approved for production keys.
- `mobile-wallet`: 51 npm advisories (22 high, 3 critical), obsolete CosmJS
  crypto, Expo 51 / React Native 0.74, and no test files.
- Both clients are now labeled deprecated in public status. Do not use them for
  real keys or production funds until GH-8 completes a migration or removal plan.
