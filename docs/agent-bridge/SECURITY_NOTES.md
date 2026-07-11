# Security Notes

## Open

- Token/ledger audit still blocks production on runtime supply enforcement,
  cap-backed rewards, DEX custody/authorization, and genesis reconciliation.
  Canonical denomination and declared treasury/stake custody are remediated on
  stacked branches but remain unmerged. See `CODEX_AUDIT.md` and GH-7.
- Anonymous legacy rating signatures and Groth16 proofs do not bind a bank
  reward recipient. Direct payout to the transaction sender is front-runnable;
  production handlers therefore defer those rewards pending GH-13/GH-7.
- `docs/status.json` says the ZKP web client is a SHA-256 mock; user-facing
  anonymity claims must clearly distinguish mock proof generation from real Groth16.
- Rust stable CosmWasm 3.0.4 dev-tooling pulls unmaintained/unsound transitive
  crates through Wasmer. No fixable cargo-audit vulnerability remains, but the
  warnings require monitoring or a stable upstream upgrade.
- Full consensus/token-conservation/authentication audit is pending.
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

## Legacy client blockers

- `web-wallet`: 68 npm advisories (26 high, 2 critical), obsolete CosmJS crypto,
  legacy Create React App toolchain, and extensive source-map warnings.
- `mobile-wallet`: 51 npm advisories (22 high, 3 critical), obsolete CosmJS
  crypto, Expo 51 / React Native 0.74, and no test files.
- Both clients are now labeled deprecated in public status. Do not use them for
  real keys or production funds until GH-8 completes a migration or removal plan.
