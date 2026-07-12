# PR #24 Audit — GH-8 Documentation, Public Status, and CI Runtime
> Scope: agent/operator/public docs, repository wiki, docs consistency gate, and six GitHub workflows  ·  Date: 2026-07-12  ·  Result: 0 FAIL / 1 WARN / 7 PASS

## Summary

GH-8 now reconciles contributor, operator, landing-page, FAQ, and wiki status to
the 683-case recovery source of truth without claiming that the stacked code is
merged or production-ready. The documentation gate validates the real wiki
instead of silently skipping a nonexistent directory and also proves suite,
module, and 21M-cap arithmetic. Official GitHub Actions are modernized while
preserving read-only, non-persisted checkout credentials and avoiding duplicate
feature-branch runs. Every modernized Action and the manual security matrix
passes on GitHub. The separate default-branch visibility decision remains open.

## Findings by domain

### Recovery truth and safety claims — PASS

- **[PASS] Current counts and token cap have one machine source** — `docs/status.json`, `scripts/check-consistency.sh`
  - What: CI requires 649 Go + 26 Rust + 8 maintained-client = 683, requires
    Go module counts to sum to 649, and requires 21,000,000 × 10^6 to equal
    21,000,000,000,000 base units.
  - Path: A stale README, agent guide, landing page, or wiki status total now
    fails the documentation job.
  - Fix: Keep `docs/status.json` authoritative and change evidence atomically.

- **[PASS] False availability and readiness claims are removed** — `docs/FAQ.md`, `wiki/Home.md`, `wiki/security/Known-Issues.md`
  - What: The prior wiki/FAQ claimed anonymous voting and a mobile wallet were
    available, called the project Testnet Ready, and reported no high/critical
    issues. Current docs state fail-closed ZKP clients, deprecated legacy
    wallets, an unmerged stack, and explicit release blockers.
  - Path: Users are no longer directed toward unsafe clients or led to treat
    internal green checks as external production approval.
  - Fix: Preserve the non-production warning until every release gate closes.

### Installation and agent handoff — PASS

- **[PASS] Commands identify the branch that actually implements them** — `docs/INSTALL.md`, `CLAUDE.md`
  - What: The old instructions cloned `main` and immediately used GH-21-only
    lifecycle commands. Recovery setup now explicitly selects the published
    GH-21 branch and lists PR #24 in the ordered stack.
  - Path: A fresh operator no longer mistakes the old main placeholder for the
    tested persistent node.
  - Fix: Remove the temporary branch switch only after ordered merge to main.

### Wiki integrity — PASS

- **[PASS] The gate checks real, existing wiki status pages** — `wiki/status/Current-Status.md`, `wiki/status/Testing-Status.md`, `.github/workflows/docs-check.yml`
  - What: `wiki-github/` did not exist and every wiki check was silently
    skipped. CI now watches `wiki/**`, checks real Home/current/testing pages,
    and all Home navigation targets exist.
  - Path: Stale public wiki counts can no longer pass solely because the script
    looked in the wrong directory.
  - Fix: Keep missing required status pages fatal rather than optional.

### CI trust and runtime — PASS

- **[PASS] Action-major modernization preserves prior hardening** — `.github/workflows/*.yml`
  - What: checkout v5, setup-go v6, and setup-node v5 are combined with
    `contents: read`, `persist-credentials: false`, explicit project Node
    versions, and manual dispatch support.
  - Path: Updating the embedded Action runtime does not reintroduce a persisted
    GitHub token into later build steps.
  - Fix: Keep official Action major updates separate from project runtimes.

- **[PASS] Feature branches no longer create duplicate push and PR suites** — `.github/workflows/go-ci.yml`, `.github/workflows/react-ci.yml`, `.github/workflows/react-native-ci.yml`, `.github/workflows/rust-ci.yml`
  - What: Push automation is limited to main; pull requests and manual dispatch
    cover feature/recovery branches.
  - Path: One branch update produces one authoritative PR suite instead of
    competing duplicate contexts.
  - Fix: Preserve main-push + PR + manual dispatch semantics.

### GitHub execution — PASS

- **[PASS] Every modernized Action runs on the rebased head** — `.github/workflows/*.yml`
  - What: checkout v5, setup-go v6, and setup-node v5 execute across docs, Go,
    Docker, Rust, maintained web, legacy mobile, and the manual security matrix.
  - Path: GitHub proves runner/input compatibility in addition to local YAML
    structure.
  - Fix: Keep these workflows green on the final documentation head.

### Remaining delivery boundary — WARN

- **[HIGH] Default-branch visibility remains separate and currently red** — PR #25
  - What: The public-main docs PR targets the vulnerable old main and its Go/
    Rust security jobs fail. The remediations live in PR #9 and later stack.
  - Path: Weakening/bypassing PR #25 checks would publish recovery claims over
    an unrecovered runtime and make the landing page imply a safer main.
  - Fix: Keep PR #25 draft; merge/review the ordered foundation first, then
    rebase or replace the visibility PR against the safe main state.

## Verification

- `./scripts/check-consistency.sh`: PASS, including suite/module/cap arithmetic
  and README/CLAUDE/landing/wiki checks
- Ruby YAML parse of all `.github/workflows/*.yml`: PASS
- `jq empty docs/status.json`: PASS
- Stale current-claim scan across agent/install/FAQ/landing/wiki/workflows: PASS
- Wiki Home target existence check: PASS
- `git diff --check`: PASS
- Underlying GH-21 final head `b59efa2`: unchanged and GitHub green
- GitHub Go/Docker `29171461365`, Rust `29171461357`, Web `29171461355`,
  Mobile `29171461342`, Docs `29171461348`, DeepScan, and CodeRabbit: PASS
- Manual Security Scan `29171476126`: PASS, all five jobs

## Priority matrix

### 🔴 BLOCKING

None in the locally audited GH-8 docs/CI diff.

### 🟠 HIGH

1. Do not merge/bypass PR #25 over the vulnerable old main; preserve ordered
   foundation merge and rebase the visibility track later.

### 🟡 MEDIUM

None identified.

### 🟢 LOW

None identified.
