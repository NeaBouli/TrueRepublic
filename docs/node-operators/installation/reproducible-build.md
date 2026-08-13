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
