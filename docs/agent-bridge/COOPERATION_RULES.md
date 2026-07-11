# Cooperation Rules

## Roles

- Gio: product, governance, release, and risk decisions.
- Lead developer: implementation, tests, build fixes, and handover evidence.
- Codex: audit, focused implementation, security review, rechecks, and GitHub coordination.

## Safety boundaries

- Preserve unrelated and pre-existing local changes.
- No destructive reset, production deployment, release, force-push, or mainnet action.
- Consensus, cryptography, wallet, token, DEX, and authentication changes are high risk.
- Derive identity from verified signers/proofs; never trust caller-supplied identity strings.
- Move tokens through the bank/treasury accounting layer; never credit declared amounts.
- Consensus state transitions must be synchronous and deterministic.

## Workflow

1. One GitHub Issue per reviewable recovery unit.
2. Branch names include the issue ID.
3. Every change has tests or an explicit NOT RUN reason.
4. Every handover lists files, commands, real results, risks, and next action.
5. Pull requests remain draft until all required checks pass.
