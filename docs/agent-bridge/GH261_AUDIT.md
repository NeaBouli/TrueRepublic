# GH-261 Unified Simulated-Tag Candidate Evidence Audit

Date: 2026-08-29

## Decision boundary

GH-261 is a repository-only release-engineering control. It binds the exact
commit and an explicitly simulated future tag string to pre-existing native
daemon and repeated-OCI digest evidence. It creates or pushes no ref or tag,
signs and publishes no artifact, uploads no binary or OCI archive, deploys no
node, handles no real key or fund, and grants no rollout or production credit.

## Implemented control

- `configs/release/candidate-evidence.json` fixes the candidate schema, exact
  40-character lowercase commit grammar, semantic simulated-tag grammar, pinned
  daemon/OCI contract digests, two native daemon targets, and four OCI target
  identities across linux/amd64 and linux/arm64.
- `candidateevidence` and `cmd/candidate-evidence` provide a standard-library
  offline verifier and CLI. Strict JSON parsing rejects duplicate/unknown keys,
  trailing values, excessive depth/size, unsafe members, symlinks, path escapes,
  target/order/digest drift, incomplete or extra evidence, and any true tag,
  push, signing, publication, deployment, or production claim.
- `scripts/generate-candidate-evidence.sh` accepts four distinct canonical input
  directories, requires their exact metadata-only member sets, cross-checks all
  contract/commit/platform/target/digest fields, emits deterministic sorted JSON,
  and removes partial output on failure.
- `.github/workflows/reproducible-daemon.yml` waits for both native daemon and
  repeated OCI matrices, downloads only their same-run JSON/checksum metadata,
  aggregates the exact commit under simulated `v0.5.0`, and retains only the
  candidate manifest plus verifier report for 14 days.
- The OCI producer writes its evidence and report into one directory. Their
  common upload-artifact root therefore downloads as the exact two flat files
  required by `bind_oci`; the repository contract pins this layout.

## Independent review and remediation

- The bounded Claude Code documentation inventory was attempted with the local
  wrapper but exhausted its configured USD 0.25 budget. It returned no usable
  report and changed no file.
- Kimi K3 implemented/reviewed the bounded core and independently reviewed the
  integration. It identified corrected test arithmetic, collision with the
  historical `v0.4.0` tag, incomplete forbidden-operation probes, the standalone
  JSON same-run limitation, and finally a protected-CI OCI artifact layout
  mismatch.
- Sol corrected each finding: 2,187 arithmetic, simulated `v0.5.0`, expanded
  repository probes without added test events, explicit workflow-vs-JSON trust
  boundaries, and a common OCI artifact root. Kimi's final read-only verdict is
  **APPROVE** with no P0/P1/P2 finding.
- CodeRabbit's protected review found two minor but valid coverage/error-path
  gaps. The checksum mutations now refresh the enclosing member digest so the
  strict checksum parser is actually reached; the text CLI now fails closed on
  output write errors. The related fixture assertion and two dead-code nits were
  also corrected without changing test-event arithmetic. Focused, Vet,
  staticcheck, complete Go and Race/Coverage gates pass after remediation.

## Reproduced local evidence

- Go build, Vet, normal tests, and Race/Coverage pass across every maintained Go
  package. The authoritative total is 1,842 test events; root is 207 at 73.6%
  statement coverage and `candidateevidence` is 77 at 87.3%.
- Rust formatting, strict Clippy, all 26 tests, and `cargo audit` pass. Cargo
  audit reports five allowed transitive maintenance/unsoundness warnings and no
  blocking vulnerability.
- The maintained client passes install, lint, 319 tests, production build and
  high-severity npm audit with zero vulnerabilities. The measured build remains
  within its enforced bundle budget.
- Pinned `govulncheck`, staticcheck, and gitleaks gates pass; gitleaks reports no
  secret. The security, deterministic daemon, repeated OCI, release
  compatibility, rollout-genesis, install-lifecycle, candidate, documentation,
  JSON arithmetic, shell syntax, and diff-hygiene contracts all pass.
- Public arithmetic is 2,187 = 1,842 Go + 26 Rust + 319 maintained-client.
  Rollout remains 35/59, phase work 35/51, Phase 6 6/7, production false.

## Protected closeout evidence

- PR #262 final head `e3c758a5a24da84a94162378072d68a55245ad84`
  passed all 25 reported hosted contexts. Both CodeRabbit review threads were
  answered and resolved before the authorized squash merge.
- The implementation merged as exact main
  `7b45c7a9522b79601ba6d32b0d8a84609fcadff8`; GH-261 closed and the remote
  feature branch was deleted.
- Exact merged main passed Go CI `33286839624`, Reproducible Linux Daemon
  `33286839598`, Security `33286839603`, Docs `33286839602`, and Pages
  `33286839220`.
- The live GitHub Wiki is synchronized at
  `2305c63dbab6928979c423a09a01aa2cc7fd8ef3`. GH-29 issue comment
  `5466191666` records the boundary without changing rollout credit.

## Residual risks and release gate

- Protected CI executed real native amd64/arm64 double builds and OCI exports
  for the exact PR head and exact merged main. That same-run evidence is green;
  it is not a substitute for a separately authorized real tagged release.
- Runner/BuildKit identities float and daemon Debian packages come from live
  repositories, so cross-time hermetic reproducibility remains unproven.
- Standalone retained JSON has no authenticated workflow-run identity. The
  protected workflow supplies same-run selection; the manifest itself binds
  only commit, simulated tag, contracts, platforms, targets, and digests.
- Real tag creation, authenticated provenance, signing, publication, staged
  deployment, independent release approval, and go/no-go remain open.
