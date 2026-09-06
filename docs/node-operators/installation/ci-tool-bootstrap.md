# Locked CI/release tool bootstrap

GH-278 replaces live-resolved CI/release tool installation with
repository-owned locked bootstraps plus deterministic fail-closed composition
evidence. Previously the Reproducible Linux Daemon and Security Scan workflows
resolved CycloneDX GoMod, `@cyclonedx/cyclonedx-npm`, govulncheck, staticcheck
and gitleaks through `go install ...@version` and `npm install --global`
without any repository-owned Go module or npm lockfile, so a drifted tool
dependency could affect both same-job SBOM outputs without detection.

## Repository-owned locks

- `tools/ci/go.mod` + `tools/ci/go.sum` form a nested tool-only Go module that
  locks CycloneDX GoMod, govulncheck, staticcheck and gitleaks including
  transitives. It builds no repository code and is excluded from the root
  module's package selection (`scripts/go-packages.sh`).
- `tools/ci/npm/package.json` + `tools/ci/npm/package-lock.json` pin
  `@cyclonedx/cyclonedx-npm` exactly. Workflows install it only through
  `npm ci`; no `npm install`, global install, or `npx` live resolution remains
  in the covered workflows.
- `configs/release/tool-platform.json` (`truerepublic.release-tool-platform/v1`)
  declares the bootstrap allowlist: for each tool its id, kind, build
  package/module, the `configs/security/gates.json` version key, and the
  artifact path whose digest the evidence binds. The security gate contract is
  the operational version source of truth; the same exact versions are
  necessarily repeated and cross-checked in the repository lockfiles.
  Dependabot covers the new npm lock without weakening the existing
  ecosystems.

`./scripts/build-ci-tool.sh --tool <id> --output-dir <new-dir>` builds one
allowlisted tool fail-closed: it runs `go mod verify`, requires the locked
module version to equal the configured gate version, builds with
`go build -mod=readonly -trimpath -buildvcs=false` and deterministic ldflags,
and checks the built binary's reported version. Relative output paths are
resolved once against their existing parent before the script changes its
working directory, and failed builds remove their partial output. The npm tool installs through
`npm ci --ignore-scripts` from the copied repository lock into a new output
directory. The script reads no secrets and mutates no project state.

## Composition evidence

The protected `tool bootstrap evidence` job builds every configured tool twice
and binds the composition in `truerepublic.tool-bootstrap-evidence/v1`:

- the exact configured tool versions from `configs/security/gates.json`;
- SHA-256 of the tool-platform contract and the security gate contract;
- SHA-256 of all four repository-owned lock files;
- SHA-256 of each twice-built tool artifact (both builds must match exactly);
- explicitly false `signed`, `published`, `deployed`, `production` and
  `long_term_hermetic` claims.

Verify the contract, fixtures and repository guards offline:

```bash
make ci-tool-bootstrap-contract-test
```

Verify a generated evidence bundle directly:

```bash
./scripts/verify-tool-bootstrap-evidence.sh \
  --evidence /path/to/tool-bootstrap-evidence \
  --artifacts /path/to/ci-tools-b \
  --output json
```

The offline Go verifier strictly parses bounded JSON and rejects unknown,
duplicate, trailing or missing fields, path or symlink escape, contract or
lock digest drift, version drift, artifact digest drift, incomplete or
duplicate tools, undeclared artifact members, and any true
signed/published/deployed/production/long-term-hermetic claim. The retained
Actions artifact contains only the bounded evidence and report JSON with
14-day retention; no tool binaries or build payloads are retained or
published.

## Security and rollout boundary

The evidence proves only that the two recorded same-job builds of the pinned
tools agreed and that the repository locks match the configured versions at
verification time. It does not authenticate the runner, does not make tool
resolution long-term hermetic (module and npm registry availability remain
external inputs), and creates no tag, signature, attestation, release,
registry push, deployment, production qualification, or rollout credit.
Rollout remains 35/59 and `production_ready` remains false.

See also [Reproducible build](reproducible-build.md),
[Offline release evidence](release-evidence.md), and
[Cross-run rebuild evidence](cross-run-evidence.md).
