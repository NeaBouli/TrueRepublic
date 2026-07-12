# Security Notes

## Open

- The token/ledger slice passes locally/GitHub through GH-12, but remains
  stacked and unmerged; independent review is still required.
  Canonical denomination and declared treasury/stake custody are remediated on
  stacked branches but remain unmerged. See `CODEX_AUDIT.md` and GH-7.
- Anonymous legacy rating signatures and Groth16 proofs do not bind a bank
  reward recipient. Direct payout to the transaction sender is front-runnable;
  production handlers therefore defer those rewards pending GH-7.
- GH-20's on-chain ZKP binding passes locally, but a real compatible prover,
  ceremony artifact, and independent circuit review do not yet exist. Both web
  clients fail closed and must remain non-submittable.
- Rust stable CosmWasm 3.0.4 dev-tooling pulls unmaintained/unsound transitive
  crates through Wasmer. No fixable cargo-audit vulnerability remains, but the
  warnings require monitoring or a stable upstream upgrade.
- GH-21 node lifecycle and final stacked/independent review remain pending.
- The legacy node initialization script invokes unavailable `x/staking` gentx
  commands. PR #19 deliberately refuses a hard-coded validator secret and
  requires real CometBFT keys; GH-21 must deliver the production PoD bootstrap.
- The v0.4 client production bundle is 1.68 MB (309 kB gzip); route-level code
  splitting is recommended before treating low-bandwidth/mobile UX as ready.

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

## Legacy client blockers

- `web-wallet`: obsolete CosmJS crypto, legacy Create React App toolchain, and
  extensive source-map warnings. Current npm audit is clear, but the dependency
  architecture is not approved for production keys.
- `mobile-wallet`: 51 npm advisories (22 high, 3 critical), obsolete CosmJS
  crypto, Expo 51 / React Native 0.74, and no test files.
- Both clients are now labeled deprecated in public status. Do not use them for
  real keys or production funds until GH-8 completes a migration or removal plan.
