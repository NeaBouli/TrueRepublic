# Offline release evidence

TrueRepublic's repository-owned release-evidence contract is
`truerepublic.release-evidence/v1`. It binds an exact Git commit and the
deterministic daemon-build contract to both supported Linux artifacts, their
checksums and build metadata, two normalized CycloneDX SBOMs, the pinned
release tool/platform contract, and a separate unsigned provenance record.

## Supported repository targets

| Target | Native CI runner | Runner architecture |
|---|---|---|
| `linux-amd64` | `ubuntu-24.04` | `x86_64` |
| `linux-arm64` | `ubuntu-24.04-arm` | `aarch64` |

The exact contract is `configs/release/tool-platform.json`. It pins Go
1.26.6, Node 22.22.2, npm 10.9.7, `cyclonedx-gomod` v1.10.0,
`@cyclonedx/cyclonedx-npm` 6.0.1, and the manifest-list digests used by both
maintained Dockerfiles. The image manifests cover Linux amd64 and arm64.
Digest pinning does not yet prove byte-reproducible container images.

## Generate and verify a local candidate bundle

Build both daemon targets natively with
`scripts/build-deterministic-daemon.sh`. Generate each Go and maintained-client
SBOM twice with the pinned tools, then pass the four raw JSON files to:

```bash
./scripts/generate-release-evidence.sh \
  --source-ref "$(git rev-parse HEAD)" \
  --amd64-dir /path/to/linux-amd64 \
  --arm64-dir /path/to/linux-arm64 \
  --go-sbom-a /path/to/go-a.json \
  --go-sbom-b /path/to/go-b.json \
  --client-sbom-a /path/to/client-a.json \
  --client-sbom-b /path/to/client-b.json \
  --output-dir /new/path/release-evidence

./scripts/verify-release-evidence.sh \
  --bundle /new/path/release-evidence
```

The generator removes random SBOM serial numbers and timestamps, requires the
two normalized outputs for each component to be byte-identical, assembles the
bundle, and invokes the verifier before leaving an output directory behind.
The verifier rejects malformed, oversized, deeply nested, duplicate-key,
unknown-field, trailing-data, path-traversal, absolute-path, backslash and
symlink-escape inputs. It also rejects missing, extra, duplicated, reordered or
cross-unbound targets, checksums, metadata, SBOMs and provenance.

Verification sets `GOTOOLCHAIN=local`, `GOPROXY=off` and
`GOFLAGS=-mod=readonly`; the pinned Go toolchain and module cache must already
be present. It performs no network access.

## CI and trust boundary

Pull-request CI generates the Go and maintained-client SBOM twice with the
exact pinned tools, compares normalized parity, and exercises a synthetic
two-target bundle through positive and adversarial verification. CI does not
upload daemon binaries or publish a release bundle.

The provenance file is intentionally unsigned. Every bundle must explicitly
state `signed: false`, `published: false`, and `production: false`. The
verifier proves internal digest and contract consistency; it does not prove
who created the bundle, independently authenticate SBOM component truth, sign
an artifact, create a tag or GitHub Release, push a container, deploy a node,
or authorize rollout. Those remain Phase 7 release gates.
