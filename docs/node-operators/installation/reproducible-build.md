# Reproducible Linux daemon build

The maintained build contract is
`configs/build/deterministic-linux-daemon.json`. It fixes Go 1.26.6, CGO,
the Linux amd64 and arm64 targets, the main package, and all deterministic Go
and linker flags. The source ref is a full Git commit SHA and is injected into
`main.version`; verification requires `truerepublicd --version` to report that
exact ref.

Because the daemon links wasmvm through CGO, each architecture is built on a
native GitHub-hosted Linux runner. The workflow does not claim that a
cross-compiled CGO binary is equivalent. Each runner uses two independent Go
build caches, builds the same commit twice with `-trimpath`,
`-buildvcs=false`, an empty Go build ID, and a disabled ELF linker build ID,
then requires identical SHA-256 values.

For a clean checkout at the current commit, run:

```bash
make deterministic-build-contract-test
make deterministic-linux-daemon \
  DETERMINISTIC_TARGET=linux-amd64 \
  SOURCE_REF="$(git rev-parse HEAD)"
```

The second command must run natively on the selected Linux architecture and
requires an output directory that does not already exist. It emits one binary,
`CHECKSUMS.sha256`, and deterministic `build-metadata.json`
under `build/deterministic/<target>/`. Pull-request CI uploads only the
checksum and metadata files as a short-lived workflow artifact; the daemon
binary is not uploaded. This does not create a tag, GitHub Release, registry
image, signature, SBOM, provenance attestation, deployment, production
artifact, or rollout approval.

The repository-only Phase 7 foundation additionally pins the maintained
container bases and release/SBOM toolchains and provides a strict offline
two-target evidence verifier. See [Offline release evidence](release-evidence.md).

GH-258 adds the versioned `configs/build/reproducible-oci.json` contract and a
separate native amd64/arm64 CI matrix. For each platform it exports the daemon
and maintained-client OCI layout twice with `--no-cache`, `--pull`, disabled
provenance/SBOM attestations, and a commit-derived `SOURCE_DATE_EPOCH`. The
strict offline verifier compares OCI index, manifest, config, and ordered layer
digests rather than requiring incidental tar header order to match:

```bash
make reproducible-oci-contract-test
./scripts/verify-reproducible-oci.sh --evidence /path/to/oci-evidence
```

Docker/Buildx is required to generate real evidence; local contract fixtures do
not substitute for the protected native runners. CI uploads only the evidence
manifest and verifier report for 14 days, not the image archives. This proves
same-job identity parity, not a cross-time hermetic rebuild: the runner/BuildKit
identity is intentionally recorded as unpinned and Debian packages are resolved
from live repositories. No image is tagged, pushed, signed, attested, deployed,
or approved for rollout.
