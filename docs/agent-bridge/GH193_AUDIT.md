# GH-193 Maintained-client Wallet and Signing Safety Audit

Status: locally verified candidate; full repository and protected publication
gates pass. Protected PR #194 exact-head review and merge remain pending.

## Implemented boundary

- Normalized 12/24-word English BIP-39 imports validate wordlist and checksum
  with bounded errors. Service-layer name and password checks cannot be bypassed
  by calling below the form.
- New custody payloads are versioned AES-256-GCM with unique salts/IVs and
  PBKDF2-HMAC-SHA-256 at 600,000 iterations. Successfully authenticated legacy
  100,000-iteration payloads are transparently re-encrypted; wrong passwords,
  truncation, corruption and malformed storage fail closed.
- A decrypted mnemonic must re-derive the exact selected address. The canonical
  signing client rejects empty, malformed, wrong-length, foreign-prefix and RPC
  chain-ID mismatches before signing. Native send inputs are service-validated.
- Session-bound signer proxies check the active generation before and after
  account access and direct signing. Lock, switch, active-wallet deletion,
  reload, and stale asynchronous unlock completion invalidate the signer.
- Persisted Zustand state contains wallet metadata only. Password, current
  wallet, mnemonic, signer, private key, history and IBC records remain outside
  the encrypted secret payload or are explicitly non-secret/scoped.

## Verification so far

- 10 Node policy/budget cases and 298 Vitest cases pass.
- Client lint, TypeScript build, Vite production build and bundle budgets pass:
  20 lazy routes, 71.15 kB gzip entry, 4.91 kB maximum lazy route and 355.07 kB
  total JavaScript.
- Three disposable local-chain delivery/rejection cases pass with the canonical
  chain-ID gate active.
- Candidate standard baseline: 1,795 = 1,461 Go + 26 Rust + 308 client. Rollout
  candidate: 32/59 overall, 32/51 phase work, Phase 6 6/7, production false.

## Delegation and independent evidence

Kimi K3 implemented the large bounded wallet/service/test slice. Sol reviewed
and changed every accepted write, including the versioned KDF migration and
live signer-session boundary. Claude Code provided only a small read-only map of
signing paths and identified the chain-ID, derived-address, stale-component and
active-delete gaps. No delegated agent committed, pushed or wrote externally.

Kimi's initial P2 concurrency finding was resolved by moving the wallet-storage
read/check/merge/write section after asynchronous encryption and adding a
concurrent-save regression. Sol also closed the stale balance-refresh race and
bounded malformed-JSON failures before the final client rerun. No known
P0/P1/P2 remains.

## Standards and residual risk

The 600,000 PBKDF2-HMAC-SHA-256 work factor follows the current
[OWASP Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html),
and AES-GCM follows the W3C authenticated-encryption guidance. JavaScript cannot
guarantee immediate erasure of immutable strings, and a compromised same-origin
runtime can access an unlocked session. Hardware/extension custody, real keys,
accounts/funds, public RPC, release, deployment and production remain outside
this evidence. Locking also cannot recall transaction bytes if it lands after
`signDirect` has returned them in the final handoff window before broadcast.
