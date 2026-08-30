# Feature Matrix

The machine-readable project state is
[`docs/status.json`](https://github.com/NeaBouli/TrueRepublic/blob/main/docs/status.json).
The human-readable delivery boundary is maintained in the
[rollout roadmap](https://github.com/NeaBouli/TrueRepublic/blob/main/docs/ROLLOUT_ROADMAP.md)
and [limitations](https://github.com/NeaBouli/TrueRepublic/blob/main/docs/LIMITATIONS.md).

Current headline state: 2,187 verified standard-suite cases, 35/59 rollout
items complete, Phase 6 at 6/7, and `production_ready=false`. Implemented or
architected surface area must not be interpreted as rollout approval.

| Release-engineering control | Verified boundary | Still open |
|---|---|---|
| GH-212 offline release evidence | Exact checksums, metadata, normalized SBOM and unsigned provenance contract | Signing, authenticated provenance and publication |
| GH-258 repeated OCI evidence | Native same-job daemon/client index, manifest, config and ordered-layer parity | Cross-time runner/BuildKit/APT hermeticity |
| GH-261 candidate evidence | Exact commit plus simulated-tag binding across both daemon records and four OCI identities; metadata/JSON only; all promotion claims false | Real tag, signed/published artifacts, deployment and go/no-go |
