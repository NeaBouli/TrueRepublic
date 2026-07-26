# Cooperation Rules

## Roles

- Gio: product, governance, release, and risk decisions.
- Lead developer: implementation, tests, build fixes, and handover evidence.
- Codex: audit, focused implementation, security review, rechecks, and GitHub coordination.
- Codex Sol/main agent: architecture, task splitting, risk calls, integration,
  final verification, GitHub updates, PR merge decisions, and Bridge updates.
- Kimi K3: preferred senior implementation and deep-review partner for larger
  bounded changes, difficult bugs, repo-wide analysis, and long-context work.
  Sol reviews every Kimi write diff and reruns the complete relevant checks.
- `spark_worker`: small bounded patches, file search, UI/text fixes, and focused
  local checks only. It must return findings to the main agent for integration.

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
6. The main Codex agent may delegate a larger bounded secret-free implementation
   or deep review to Kimi K3 through the configured wrapper, and small
   implementation/search/check tasks to `spark_worker`; it keeps architecture,
   security judgment, merge, push, and final status responsibility.
7. Delegated agents may not start other agents and never receive secrets,
   production access, or external-write authority.
