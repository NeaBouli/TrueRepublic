# Cross-run rebuild evidence

GH-273 adds a repository-only comparison of two distinct
`Reproducible Linux Daemon` workflow executions for the same exact `main`
commit. It extends the same-job OCI parity from GH-258 and the same-run
candidate aggregation from GH-261. It retains JSON metadata and digests only;
it does not retain daemon binaries or OCI archives.

The versioned contract is `configs/release/cross-run-rebuild.json`. It pins the
repository, workflow path, `main` branch, `workflow_dispatch` event, decimal
run-ID grammar, 14-day evidence window, and the exact GH-261 candidate contract
digest.

## Local contract verification

Run the positive and adversarial fixtures plus the semantic workflow contract:

```bash
make cross-run-evidence-contract-test
```

Verify a complete seven-member comparison directory directly:

```bash
./scripts/verify-cross-run-evidence.sh \
  --evidence /path/to/cross-run-evidence \
  --repository NeaBouli/TrueRepublic \
  --workflow-path .github/workflows/reproducible-daemon.yml \
  --branch main \
  --commit <40-lowercase-hex-commit> \
  --baseline-run-id <baseline-run-id> \
  --current-run-id <current-run-id> \
  --output json
```

The verifier requires the comparison manifest, both run receipts, and both
candidate manifests/reports with exact basenames and SHA-256 bindings. It
rejects unknown or duplicate JSON fields, trailing/deep/oversized JSON,
symlinks, path escape, non-regular or undeclared members, same-run reuse,
identity or retention drift, malformed GitHub metadata, different commits,
binary digest differences, and any difference in the four ordered OCI index,
manifest, config, or layer vectors.

All of these claims must be explicitly present and false:

- `real_tag_created`
- `ref_pushed`
- `signed`
- `attested`
- `published`
- `deployed`
- `production`
- `long_term_hermetic`

## Protected two-run qualification

The first protected `workflow_dispatch` run is started without
`baseline_run_id`. After that run succeeds, a second dispatch of the unchanged
exact `main` commit supplies the first run ID. A separate comparison job has
only `contents: read` and `actions: read`. Before downloading anything it
checks the baseline through the GitHub API for the exact repository, workflow,
branch, event, commit, successful completion, 14-day window, and exactly one
non-expired exact-name candidate artifact. The pinned download action treats a
digest mismatch as an error and never checks out a caller-selected ref.

The retained comparison artifact contains only the seven bounded evidence JSON
members and the verifier report. A successful hosted pair proves equality only
for those two recorded runs and that exact commit.

## Security and rollout boundary

This capability does not make the builds long-term hermetic. GitHub runner
images, runner-provided BuildKit, and live Debian APT indexes remain floating
inputs. It authenticates neither the actor nor the runner independently, and it
does not create a real tag, push a ref, sign or attest provenance, publish a
release, deploy a node, approve production, or close the Phase 7 tagged and
signed artifact gate. Rollout remains 35/59 and `production_ready` remains
false.

See also [Release-candidate evidence](candidate-evidence.md),
[Reproducible build](reproducible-build.md), and
[Offline release evidence](release-evidence.md).
