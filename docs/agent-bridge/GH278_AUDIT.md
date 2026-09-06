# TrueRepublic GH-278 — Audit
> Scope: repository-owned CI/release tool locks, bootstrap/evidence verifier, workflow integration, tests and synchronized public documentation  ·  Date: 2026-09-06  ·  Result: 0 FAIL / 4 WARN / 7 PASS

## Summary

GH-278 is locally ready for protected review. The release and security Go/npm
tools in scope are locked by repository-owned module and package locks, built
through fail-closed scripts, and bound into deterministic composition evidence
whose release and production claims remain false. Sol closed every material
finding from Kimi K3's implementation and two independent reviews, then reran
the complete relevant local verification chain on the stable final diff. No
runtime, consensus, genesis, wallet, token, deployment or production behavior
changes.

## Findings by domain

### Tool bootstrap and supply chain — PASS

- **[PASS] Exact repository-owned tool composition** — `tools/ci/go.mod`, `tools/ci/go.sum`, `tools/ci/npm/package.json`, `tools/ci/npm/package-lock.json`
  - What: CycloneDX GoMod, govulncheck, staticcheck, gitleaks and CycloneDX npm
    have exact direct pins plus repository-owned transitive locks.
  - Path: `go mod verify` and `go build -mod=readonly` cover Go; `npm ci`
    consumes only the committed npm lock.
  - Fix: none.

- **[PASS] Version source and artifact paths fail closed** — `scripts/build-ci-tool.sh`
  - What: only contract-allowlisted tools may build; pins are cross-checked
    against the security-gate contract and built version output.
  - Path: malformed tool IDs, unavailable/existing output directories,
    unclean artifact paths, lock drift and version drift terminate the build
    and clean incomplete output.
  - Fix: none.

### Evidence integrity — PASS

- **[PASS] Strict bounded offline verifier** — `toolbootstrapevidence/`
  - What: strict JSON rejects duplicate, unknown, trailing, deep and oversized
    input while recomputing contract, gate, lock and artifact digests.
  - Path: symlinked/extra/missing artifacts, incomplete or duplicate tool sets,
    digest drift and any non-false promotion claim are rejected.
  - Fix: none.

- **[PASS] Two-build composition binding** — `scripts/generate-tool-bootstrap-evidence.sh`
  - What: two distinct canonical artifact roots must have identical declared
    digests, and the generated evidence is verified against both roots.
  - Path: aliased roots, first- or second-build symlinks, missing members and
    digest drift fail before a complete evidence bundle remains.
  - Fix: none.

### CI and release integration — PASS

- **[PASS] Covered live installs removed** — `.github/workflows/reproducible-daemon.yml`, `.github/workflows/security-scan.yml`, `.github/workflows/docs-check.yml`
  - What: covered `go install ...@version`, global `npm install` and `npx`
    bootstrap paths are replaced or prohibited by semantic repository tests.
  - Path: protected CI builds the locked tools and the reproducible workflow
    builds all five twice before producing bounded JSON-only evidence.
  - Fix: none.

- **[PASS] Nested module isolation and maintenance policy** — `scripts/go-packages.sh`, `.github/dependabot.yml`, `.gitignore`, `REUSE.toml`
  - What: the tool-only Go module is excluded from root package selection;
    Dependabot covers the npm lock; only the two required npm fixture package
    manifests are commit-visible under synthetic `node_modules` trees.
  - Path: repository contract mutations prove missing coverage or accidental
    nested selection is rejected.
  - Fix: none.

### Compatibility and public truth — PASS

- **[PASS] Additive strict release-contract compatibility** — `releaseevidence/types.go`
  - What: the release tool contract explicitly binds the new bootstrap member
    while the shared parser continues to reject unknown JSON.
  - Path: the real additive contract and release evidence package tests pass;
    contract digest drift remains fail closed.
  - Fix: none.

- **[PASS] Test and rollout accounting** — `docs/status.json`, `scripts/check-consistency.sh`, public docs and Wiki sources
  - What: real enumeration is 2,353 = 2,008 Go + 26 Rust + 319 client, including
    root/application 244 and tool-bootstrap evidence 22.
  - Path: module totals equal the Go total and the consistency gate requires the
    new Wiki module row. Rollout remains 35/59 and production remains false.
  - Fix: none.

### Residual scope — WARN

- **[LOW] Generator output path is not normalized to absolute form** — `scripts/generate-tool-bootstrap-evidence.sh`
  - What: unlike the build script, the generator retains a caller-relative
    output path.
  - Path: the generator never changes working directory, rejects existing or
    symlinked output and safely removes only incomplete output, so no current
    correctness or traversal path exists.
  - Fix: normalize in a future cleanup if uniform path handling is desired.

- **[LOW] Missing bootstrap is enforced downstream** — `releaseevidence/types.go`, `toolbootstrapevidence/verify.go`
  - What: Go JSON decoding permits an omitted bootstrap member as its zero
    value in the general release-evidence type.
  - Path: the dedicated verifier rejects an empty tool set, repository tests
    require the exact member, and the complete contract digest is bound; there
    is no acceptance bypass.
  - Fix: optional future explicit presence validation in the general parser.

- **[LOW] Two pre-existing fetch paths remain outside GH-278** — `.github/workflows/security-scan.yml`, `.github/workflows/react-ci.yml`
  - What: locked cargo-audit installation and Playwright browser provisioning
    remain outside this ticket's Go/npm release-tool scope.
  - Path: neither supplies the five GH-278 tools or its composition evidence.
  - Fix: assess separately if the project later expands hermetic CI scope.

- **[LOW] REUSE CLI unavailable locally** — `REUSE.toml`
  - What: the host does not have `reuse lint` installed.
  - Path: all new maintained paths were inspected against the repository policy
    and the project's deterministic Apache-2.0 license gate passed.
  - Fix: rely on protected tooling or add a repository-pinned REUSE tool in a
    separately scoped ticket.

## Verification evidence

- Kimi K3 implemented the bounded secret-free core, then performed two
  independent read-only reviews. Final verdict: `APPROVE`, no P0/P1/P2.
- Protected first-head review identified a vulnerable indirect
  `github.com/go-git/go-git/v5` version and imprecise CLI/TODO failure-state
  handling. The tool lock now uses patched v5.16.5, missing option values exit
  with usage status 2, the queue separates completed local work from hosted
  closeout, and focused plus complete Race/Coverage verification passes.
- `make verify`: build, vet and full Race/Coverage passed; root 73.6%, release
  evidence 74.5%, tool-bootstrap evidence 79.6%.
- Full standard Go enumeration: 2,008 passing cases; root 244 and
  tool-bootstrap evidence 22.
- Client: lint, 10 Node tests, 309 Vitest tests, production build and high audit
  passed; four documented skips remain expected. npm audit reports zero known
  vulnerabilities for the new tool lock.
- Rust: fmt, Clippy with denied warnings, 26 tests, workspace build and audit
  policy passed.
- Real twice-built five-tool evidence, govulncheck, staticcheck, gitleaks and
  SBOM parity paths passed during implementation verification.
- Focused adversarial contracts, docs consistency, Apache-2.0 license policy,
  YAML/JSON parsing and `git diff --check` passed.

## Priority matrix

### 🔴 BLOCKING

None.

### 🟠 HIGH

None.

### 🟡 MEDIUM

None.

### 🟢 LOW

1. Optional generator output-path normalization.
2. Optional explicit general-parser bootstrap presence check.
3. Separately assess pre-existing cargo/Playwright provisioning if hermetic CI
   scope expands.
4. Optional repository-pinned REUSE CLI coverage.
