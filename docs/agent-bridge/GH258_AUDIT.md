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

## Residual risks

- The GitHub runner labels and runner-provided Buildx/BuildKit versions float.
- Debian packages in the daemon build are resolved from live repositories.
- Therefore GH-258 proves same-commit, same-job OCI identity parity, not a
  cross-time hermetic rebuild.
- Tagged builds, authenticated provenance, signing, publication, staged
  deployment, independent release approval, and remaining artifact classes stay
  open under GH-29 Phase 7.
