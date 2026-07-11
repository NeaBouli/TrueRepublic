# Recovery Queue

## P0 - security and reproducibility

- [ ] GH-4: keep the recovery epic and acceptance criteria current.
- [ ] GH-5: local Go 1.26.5 build/race/vet/govulncheck and Rust
  tests/Clippy/audit are green; confirm GitHub security CI.
- [ ] GH-6: local v0.4 client lint, tests, build, exact amount handling,
  maintained wallet crypto, and npm audit are green; confirm GitHub CI.
- [x] GH-8: reproduce legacy web wallet and mobile wallet CI/security state.

## P1 - consensus and wallet audit

- [ ] GH-7: audit PNYX 21M cap and six-decimal base-unit conversion end to end.
- [ ] GH-7: audit bank/treasury conservation, fee transfer, reward, burn, and replay rules.
- [ ] GH-7: audit ZKP/nullifier/domain-key authentication and client-side mock boundaries.
- [ ] GH-7: audit DEX rounding, slippage, pool accounting, and authorization.

## P2 - delivery

- [ ] GH-8: reconcile useful changes from the preserved legacy checkout.
- [ ] GH-8: align README, CLAUDE.md, status JSON, limitations, website, and test counts.
- [ ] GH-8: add bridge/docs consistency to CI.
- [ ] Commit atomic recovery blocks, push branch, open draft PR, and monitor checks.
