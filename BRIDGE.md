# TrueRepublic Agent Bridge

This file is the public recovery bridge on the default branch. Detailed task
history remains attached to the linked GitHub issues and stacked pull requests.

- Recovery epic: [GH-4](https://github.com/NeaBouli/TrueRepublic/issues/4)
- Public-status reconciliation: [GH-8](https://github.com/NeaBouli/TrueRepublic/issues/8)
- Final stacked recovery review: [PR #24](https://github.com/NeaBouli/TrueRepublic/pull/24)

## 2026-07-11 19:57 EEST GH-8 public recovery status → Approved

- **Branch:** `agent/public-recovery-status`
- **Issue:** [GH-8](https://github.com/NeaBouli/TrueRepublic/issues/8)
- **Changed:**
  - `README.md` — separates the preserved `main` baseline from the unmerged
    recovery evidence and links the ordered PR stack
  - `docs/index.html` — adds a production warning to the GitHub Pages landing page
  - `docs/status.json` — marks recovery active and records the verified 21M cap
  - `docs/LIMITATIONS.md` — adds the current security and client-use warning
- **Tests:** `./scripts/check-consistency.sh` → PASS; `jq` recovery/cap
  assertions → PASS; `git diff --check` → PASS; stale-claim scan → PASS
- **Risk:** Low — documentation-only; no runtime or consensus code
- **Ready for:** GitHub CI, then merge to `main`

### Lead Dev notes

This update intentionally contains no recovery implementation. Its only purpose
is to make the public default branch truthful while the high-risk stacked PRs
remain under review.

### Codex review feedback

Approved as a minimal status-only correction. The diff contains no runtime,
consensus, dependency, or wallet implementation changes and distinguishes
`main` baseline evidence from unmerged recovery evidence.

---
