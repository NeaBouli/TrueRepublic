# Agent Bridge

This directory is the durable handover and recovery log for TrueRepublic.
GitHub Issues define externally visible tickets; these files preserve operational
context, evidence, decisions, risks, and the exact next action between sessions.

## Required reading order

1. `COOPERATION_RULES.md`
2. `PROJECT_STATE.md`
3. `TODO.md`
4. `SECURITY_NOTES.md`
5. The latest entries in `ACTION_LOG.md`

## Update discipline

- Update `ACTION_LOG.md` after every meaningful inspection, fix, or verification.
- Update `PROJECT_STATE.md` only when verified state changes.
- Keep `TODO.md` mapped to GitHub Issues and ordered by risk.
- Record product or architecture decisions in `DECISIONS.md`.
- Never write secrets, private keys, tokens, credentials, or `.env` values here.
- Public status (`docs/status.json`, landing page, README) may only claim results
  reproduced by the current branch or green GitHub CI.

