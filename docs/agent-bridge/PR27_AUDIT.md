# PR #27 Audit — GH-26 Safe Operator Initialization

Date: 2026-07-12
Branch: `fix/GH-26-pod-init-script`
Base: `fix/GH-8-docs-final`
Issue: [GH-26](https://github.com/NeaBouli/TrueRepublic/issues/26)
Result: PASS for the wrapper scope; project remains non-production

## Scope and guarantee

The operator wrapper has one bootstrap boundary: `truerepublicd init`. That
daemon command generates the CometBFT validator key, binds it to a positive-
power PoD validator, creates only the exact cap-checked bank backing, validates
the resulting ledger genesis, and writes it atomically. The wrapper changes
only local gas-price and Prometheus settings after successful initialization.

## Finding fixed

### High — public wrapper invoked an unavailable and unsafe staking flow

The old script created a keyring mnemonic, captured command output in
`genesis-key.txt`, added another genesis balance, and invoked `genesis gentx`
plus `collect-gentxs`. TrueRepublic does not wire `x/staking`; the flow could
not create the intended PoD validator and contradicted the audited daemon init.

The replacement invokes the daemon exactly once and cannot create an account,
mnemonic artifact, staking gentx, or extra supply. A regression uses an isolated
fake daemon to assert the complete command and observable file boundary.

## Verification

- `sh -n scripts/init-node.sh`: PASS
- `go test . -run TestInitNodeScriptUsesOnlyPoDBootstrap -count=1`: PASS
- Real compiled-daemon wrapper smoke: PASS; one consensus/PoD validator, exact
  canonical bank supply, correct chain ID, gas price, metrics, no mnemonic file
- `go test ./... -count=1 -timeout=600s`: PASS, 650 Go cases
- `go vet ./...`: PASS
- Documentation consistency, JSON, shell syntax, and diff checks: PASS
- GitHub Go race/coverage and Docker restart `29172845624`: PASS
- GitHub Docs `29172845627`, DeepScan, and CodeRabbit: PASS
- Manual Security Scan `29172846057`: PASS, all five jobs
- GitHub mergeability: MERGEABLE; unresolved review threads: zero

## Explicitly out of scope

- Independent multi-node, backup/restore, IBC/upgrade operations review
- Real ZKP prover/ceremony and anonymous reward-recipient binding
- Ordered protected-branch merge and default-branch visibility
