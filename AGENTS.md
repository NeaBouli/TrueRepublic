# TrueRepublic Repository Instructions

This file supplements, but does not duplicate or weaken:

- this repository's `BRIDGE.md` and `docs/agent-bridge/`; and
- when present in the local operator environment, the optional external
  coordination files `/Users/gio/AGENTS.md` and
  `/Users/gio/WORKFLOW_INDEX.md`.

The project-local files are authoritative for TrueRepublic task state,
evidence, security boundaries, and handoffs.

## Stack and source of truth

- Go 1.26.5, Cosmos SDK, CometBFT, and Wasmd implement the node and
  consensus-critical modules.
- Rust/CosmWasm sources live under `contracts/`.
- The maintained TypeScript client lives under `client-web/`. The former
  `web-wallet/` and `mobile-wallet/` prototypes were retired and removed
  under GH-112 and GH-102 respectively; they exist only in Git history.
- GitHub `NeaBouli/TrueRepublic` and canonical `origin/main` are the repository
  baseline. The preserved legacy checkout is never a merge source.

## Required workflow

- Read `docs/agent-bridge/README.md` in its stated order before material work.
- One GitHub issue is one reviewable task. Preserve unrelated work and keep
  Bridge updates append-only.
- Codex Sol owns scope, architecture, security, integration, complete
  verification, external writes, and closure.
- Kimi K3 is the preferred secret-free implementation/review partner for
  larger bounded work. Sol reviews every Kimi diff and reruns the relevant
  gates. Delegated agents may not delegate further.
- Consensus, cryptography, wallets, token accounting, DEX, authentication,
  genesis migration, release, and production work are high risk and require a
  current, exactly bounded approval.

## Build and verification

Use repository-owned package selection so installed frontend dependencies
cannot change the Go package set:

```bash
make build
make verify
./scripts/check-consistency.sh
git diff --check
```

Run focused Go tests first, then the complete relevant gate. The gated
multi-validator matrix is:

```bash
TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test . \
  -run '^(TestMultiValidatorConsensusRecovery|TestMultiValidatorLegacyAuthorityMigrationRollback|TestMultiValidatorConsensusSlashing|TestMultiValidatorConsensusKeyRotation|TestMultiValidatorTrustedSnapshotStateSync|TestMultiValidatorBackupRestoreExportImport|TestMultiValidatorPersistedBinaryUpgradeRollback|TestValidatorIdentityColdFailover)$' \
  -count=1 -timeout=1500s -v
```

For client or contract changes, use the exact lockfile-driven commands and CI
versions documented in their package/workspace files. Use `npm ci`, never
`npm install`.

## Project-specific safety

- Consensus transitions must be synchronous, deterministic, and
  export/import/restart safe.
- Identity must come from authenticated signers or verified proofs, never from
  caller-supplied address strings.
- PNYX supply is capped at 21,000,000 whole PNYX and all claims must reconcile
  with canonical bank/module custody.
- Never copy, log, commit, or delegate private keys, mnemonics, signer state,
  credentials, or production artifacts.
- No deploy, public-network launch, mainnet action, force-push, destructive
  reset, or production mutation follows from a repository task.
