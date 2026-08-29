# GH-258 Reproducible OCI Evidence Audit

Date: 2026-08-29

## Decision boundary

GH-258 is a repository-only release-engineering control. It does not create a
tag or GitHub Release, upload an OCI archive, push to a registry, sign or attest
an artifact, deploy a node, approve production, or change the 35/59 rollout
count.

## Implemented control

- `configs/build/reproducible-oci.json` fixes the four maintained targets:
  daemon and `client-web` on native Linux amd64 and arm64. It binds exact
  Dockerfiles, contexts, digest-pinned bases, source-ref injection, two
  repetitions, no cache, pull, OCI output, disabled provenance/SBOM attestations,
  and commit-derived timestamp rewriting.
- `ocievidence` provides a standard-library-only offline verifier. It rejects
  malformed, duplicate-key, unknown-field, trailing, oversized, or over-deep
  JSON; unsafe evidence members; symlinks and special files; unsafe or duplicate
  tar paths; unbounded entries/content/buffered blobs/layers; invalid OCI media
  types, descriptors, sizes, digests, platform metadata, or target sets; and any
  repeated-build index/manifest/config/ordered-layer mismatch.
- `.github/workflows/reproducible-daemon.yml` runs the two targets twice on each
  native platform, verifies the result, and retains only the evidence manifest
  and digest report for 14 days. It has contents-read permission and no registry,
  signing, publishing, deployment, or production surface.
- The repository contract binds all direct final-image inputs to workflow
  triggers, including the root Docker ignore, daemon entrypoint, and all
  maintained-client inputs. The daemon is multi-stage; intermediate `/src`
  layers are not part of the exported final image.

## Independent review

- Claude Code inspected the pre-implementation surface read-only. It confirmed
  there was no existing OCI evidence path and identified the important
  distinction between same-job parity and cross-time hermetic rebuilding.
- Kimi K3 performed the full independent diff and BuildKit-source review. During
  review it identified canonical OCI directory entries, cumulative buffering,
  missing JSON evidence retention, build-argument/config-platform binding, and
  incomplete direct image-input triggers. Sol implemented and retested every
  applicable correction. Its final condition was correcting the README test
  badge from 2,028 to the reproduced 2,084 total; that correction is applied.
  The final independent read-only verdict is **APPROVE** with no P0/P1/P2
  finding. Merge remains conditional on hosted real-image success and the
  truthful limitation boundary documented here.

## Reproduced local evidence

- Focused OCI package, command, repository, shell syntax, vet, and consistency
  gates pass on the final local diff.
- The fresh package-scoped standard suite passes 1,739 Go cases with zero failed
  test events. Combined public arithmetic is 2,084 = 1,739 Go + 26 Rust + 319
  maintained-client cases.
- Docker/Buildx is unavailable on the local host. Native hosted amd64/arm64
  exports remain the authoritative real-image evidence and must pass before
  merge.

## Hosted diagnosis and correction

- The first PR #259 native run passed the fail-closed verifier and all repeated
  `client-web` identities, but rejected both daemon pairs because their OCI
  manifests/configs differed between repetitions.
- Hosted logs and independent Kimi review isolated wall-clock content written by
  APT/DPKG in the final Debian layer. BuildKit timestamp rewriting normalizes
  filesystem metadata, not timestamps embedded inside `/var/log/dpkg.log` and
  APT log content.
- The Dockerfile now removes package-manager logs in the same layer, normalizes
  the service-account shadow date, and aligns the in-image CGO build with the
  existing deterministic daemon flags (`-trimpath`, no VCS/build IDs, readonly
  modules and fixed upgrade/version bindings). The repository contract prevents
  silent removal of these controls. Focused/full root, deterministic-contract,
  Vet and docs gates pass; Kimi's final fix-diff verdict is **APPROVE** with no
  P0/P1/P2 finding. A fully green hosted rerun remains mandatory.
- The second exact-head run again rejected both daemon pairs while both client
  pairs remained identical. The remaining highest-confidence source is glibc's
  `/var/cache/ldconfig/aux-cache`, which records per-library stat data after the
  builder-created wasmvm library is copied; export-time timestamp rewriting does
  not normalize that embedded content. The next bounded candidate removes this
  rebuildable cache after `ldconfig` and APT binary caches in their creating
  layer. Invalid reports now retain only per-repetition identity/layer digests,
  never OCI archives, so any further mismatch identifies its exact layer index.
  Focused/full root, Vet and docs gates pass; Kimi's final review is **APPROVE**
  with no P0/P1/P2 finding.
- The third exact-head report fulfilled that diagnostic purpose: all daemon
  layers match except ordered layer index 1 (the final-stage package/account RUN).
  The binary, wasmvm copy, entrypoint and final `ldconfig` layer all match, so the
  CGO and aux-cache hypotheses are now excluded as the remaining cause. The next
  bounded candidate removes the complete APT cache/log directory surface and
  non-runtime account backup/login databases in their creating layer. These are
  unnecessary for the locked non-login service user and package-manager-free
  runtime, while the actual passwd/group/shadow state remains intact.
- Sol's focused OCI/evidence tests, full root suite, Vet, docs consistency and
  diff validation pass for that candidate. Kimi K3 independently returned
  **APPROVE** with no P0/P1/P2 finding after its own contract, Vet and full-root
  checks. Native repeated amd64/arm64 verification of the pushed exact head is
  the sole remaining merge gate.
- The fourth protected exact-head run `33275792489` kept that gate closed. Both
  client pairs, both repeated daemon binaries, docs, security and release
  evidence pass. On both architectures, only daemon layer index 1 still differs
  and all other ordered layers match. The next diagnostic therefore compares
  entries inside that one layer using paths, metadata and content digests only;
  it must not emit file contents or retain/publish OCI archives.
- The bounded per-entry diagnostic is implemented for tar and gzip layers. It
  streams content directly into SHA-256, normalizes paths, sorts all reported
  differences and metadata-map inputs, and rejects traversal, duplicates,
  unsupported types/compression, digest/size drift and every configured bound.
  It adds 10 standard-suite events, taking OCI evidence to 35 cases at 81.3%
  coverage and public arithmetic to 2,094 = 1,749 Go + 26 Rust + 319 client.
- **Audit result:** 0 FAIL / 3 LOW residual resource notes / all reviewed
  security boundaries PASS. Kimi K3 returned **APPROVE** with no P0/P1/P2
  finding after reproducing Vet, 35 tests, 81.3% coverage, docs consistency and
  diff validation. The low notes concern only theoretical maximum memory/report
  size and repeated archive scans on hostile bounded evidence; no raw content,
  image, archive, tag, signature or deployment is produced.
- Sol then ran `make reproducible-oci-contract-test`, Vet and the complete
  CGO-enabled package-scoped Go suite. Every package passed, including root,
  `ocievidence`, policy/evidence modules, Treasury, DEX and TrueDemocracy.
  Documentation consistency and diff validation pass as well.
- Protected exact-head run `33277444211` exercised the new diagnostic and
  isolated one path on both architectures: `var/cache/ldconfig/aux-cache` in
  layer index 1. Per architecture its metadata digest is identical while its
  content digest differs. Package post-install hooks create it in that layer;
  the existing later removal cannot rewrite an earlier OCI layer. The bounded
  correction therefore deletes it in both creating/recreating RUNs.
- The exact two-removal correction passes repository/evidence contracts, full
  root Go test (108.158s), Vet, docs consistency and diff validation. Kimi K3
  independently returned **APPROVE** with no P0/P1/P2 finding and verified that
  the inline negative mutation is detected only by the new two-occurrence
  invariant, leaving published test-event accounting unchanged.

## Residual risks

- The GitHub runner labels and runner-provided Buildx/BuildKit versions float.
- Debian packages in the daemon build are resolved from live repositories.
- Therefore GH-258 proves same-commit, same-job OCI identity parity, not a
  cross-time hermetic rebuild.
- Tagged builds, authenticated provenance, signing, publication, staged
  deployment, independent release approval, and remaining artifact classes stay
  open under GH-29 Phase 7.
