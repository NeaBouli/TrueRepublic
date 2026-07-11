# TrueRepublic Agent Bridge

Canonical coordination lives in [`docs/agent-bridge/`](docs/agent-bridge/README.md).

- Current state: [`PROJECT_STATE.md`](docs/agent-bridge/PROJECT_STATE.md)
- Work queue: [`TODO.md`](docs/agent-bridge/TODO.md)
- Audit trail: [`ACTION_LOG.md`](docs/agent-bridge/ACTION_LOG.md)
- Decisions: [`DECISIONS.md`](docs/agent-bridge/DECISIONS.md)
- Security: [`SECURITY_NOTES.md`](docs/agent-bridge/SECURITY_NOTES.md)

GitHub recovery epic: [#4](https://github.com/NeaBouli/TrueRepublic/issues/4)

## 2026-07-11 20:09 EEST GH-4 foundation merge audit → Review

- **Branch:** `fix/GH-4-recovery-foundation`
- **Issue:** [GH-4](https://github.com/NeaBouli/TrueRepublic/issues/4)
- **PR:** [#9](https://github.com/NeaBouli/TrueRepublic/pull/9)
- **Changed:** audit-only follow-up adds Docker build coverage, records the
  merge review, and removes whitespace-only diff errors
- **Tests:** Go build/test/race/vet → PASS; govulncheck fixable gate → PASS;
  Rust 26 tests/Clippy/audit → PASS with six allowed transitive warnings;
  maintained client lint/5 tests/build/audit → PASS; docs consistency → PASS
- **Risk:** Medium — dependency/toolchain foundation; no consensus or ledger
  implementation changes
- **Ready for:** refreshed GitHub CI and an independent GitHub approval

### Codex review feedback

Conditional approval for the recovery-foundation scope. The seven ledger and
token-economy blockers in `CODEX_AUDIT.md` remain explicitly out of scope and
must stay non-production until the ordered implementation PRs land. Do not
bypass the required independent GitHub approval.

---

## 2026-07-11 20:54 EEST GH-4 wasmvm Docker linkage → Local verification

- **Branch:** `fix/GH-4-recovery-foundation`
- **Issue:** [GH-4](https://github.com/NeaBouli/TrueRepublic/issues/4)
- **PR:** [#9](https://github.com/NeaBouli/TrueRepublic/pull/9)
- **Changed:** replaced the Alpine/musl builder and runtime with Debian/glibc,
  copied the architecture-specific `libwasmvm` shared object into the runtime,
  and registered it with `ldconfig`
- **Tests:** reproduced GitHub musl/GLIBC linker failure twice; local Go build,
  docs consistency, workflow YAML, and diff checks → PASS; Docker rerun pending
  on GitHub because no local Docker runtime is installed
- **Risk:** Medium — node container build/runtime linkage
- **Ready for:** refreshed GitHub Docker build and full PR gate

### Codex review feedback

The patch matches the glibc/wasmvm linkage already proven by GH-21 while keeping
GH-4's existing entrypoint and root data path unchanged. Do not claim the image
fixed until the GitHub Docker job passes.

---
