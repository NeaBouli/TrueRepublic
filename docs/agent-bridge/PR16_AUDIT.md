# PR #16 Audit — GH-14 Bank Escrow

Date: 2026-07-11  
Branch: `fix/GH-14-bank-escrow`  
Base: `fix/GH-11-pnyx-cap`  
Issue: [GH-14](https://github.com/NeaBouli/TrueRepublic/issues/14)  
Result: PASS for the GH-14 custody scope; repository remains non-production

## Scope and guarantees

- Domain creation, proposal fees, deposits, validator registration, stake
  withdrawals, treasury withdrawals, and authenticated reward payouts settle
  exact canonical `upnyx` through the `truedemocracy` module account.
- State and bank movement use cached SDK contexts where a fallible transfer can
  otherwise leave one side committed.
- Claimed creator, member, voter, and operator identities must equal the
  authenticated signer; the same rule applies to CLI and CosmWasm messages.
- Module balance equals aggregate domain-treasury plus validator-stake claims in
  the tested lifecycle.
- Validator penalties reduce the stake claim and burn the exact matching module
  escrow, keeping parity while taking the penalty out of circulation.

## Audit findings fixed

### High — CosmWasm encoder omitted authenticated sender

The encoder populated legacy member/voter strings but not the new `Sender`
field. The hardened messages would therefore fail validation for every contract
stone or election vote. The encoder now preserves the contract sender and each
binding test asserts both signer equality and `ValidateBasic()` success.

### High — slash penalty was recyclable through domain treasury

The initial patch moved a validator penalty into the validator's primary domain
treasury. A domain admin could withdraw that balance, allowing an aligned admin
and validator to recover the penalty. The keeper now burns the exact penalty
from module escrow; the module has only the required burner permission. Tests
cover exact burn, parity, and rollback when burning fails.

## Negative-path evidence

- Unfunded and duplicate domain/stake claims do not commit state.
- Creator/operator/member/voter spoofing is rejected.
- Duplicate validator public keys are rejected.
- Mixed-denom, zero, negative, and dust operations are rejected.
- Failed account-to-module, module-to-account, reward-payout, and burn calls do
  not commit the paired custom-ledger mutation.
- Anonymous rating rewards remain deliberately deferred because the legacy
  signature/proof payload does not bind a safe recipient.

## Verification

- `go build ./...`: PASS
- `go vet ./...`: PASS
- `go test ./... -count=1`: PASS, 557 Go cases
- `go test ./... -race -count=1`: PASS
- `go test ./... -cover -count=1`: PASS; governance coverage 54.6%
- Rust workspace: 26 tests PASS; Clippy PASS; audit PASS with six allowed
  transitive dev-tooling warnings
- Maintained client: install, lint, six tests, production build, and zero-advisory
  audit PASS
- Documentation consistency and `git diff --check`: PASS
- Remaining publication gates: refreshed GitHub checks and automated review.

## Explicitly out of scope

- GH-13: cap-checked, bank-backed reward issuance
- GH-10: DEX custody, LP ownership, settlement, and burns
- GH-12: custom-genesis reconciliation and runtime conservation invariants
- GH-7: final ZKP/authentication and token-economy review

These open slices continue to block production use and real funds.
