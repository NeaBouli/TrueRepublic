# Release-candidate evidence

GH-261 adds a repository-only evidence layer above the deterministic daemon
and repeated OCI checks. It proves that the metadata produced by those
protected CI jobs belongs to one exact 40-character Git commit and the
explicitly simulated, currently uncreated `v0.5.0` future tag string. It does
not create or push that Git tag and it grants no release authority.

The versioned contract is
`configs/release/candidate-evidence.json`. It pins the deterministic-daemon and
OCI contracts, both native daemon targets, both OCI platforms, both maintained
OCI targets per platform, and two OCI repetitions. The verifier requires these
claims to be explicitly present and false:

- `real_tag_created`
- `ref_pushed`
- `signed`
- `published`
- `deployed`
- `production`

## Local contract verification

Run the committed positive/adversarial fixtures and repository wiring tests:

```bash
make candidate-evidence-contract-test
```

Verify a complete evidence directory directly:

```bash
./scripts/verify-candidate-evidence.sh \
  --evidence /path/to/candidate-evidence \
  --output json
```

The directory must contain exactly the manifest-declared JSON and checksum
members. Unknown or duplicate JSON fields, malformed tag/commit identity,
contract/target/platform drift, missing halves, stale digests, path escape,
symlinks, undeclared members, and any promoted status claim are rejected.
Daemon binaries and OCI archives are neither required nor accepted.

## Protected CI aggregation

The `release candidate evidence` job in
`.github/workflows/reproducible-daemon.yml` waits for both native deterministic
daemon builds and both native repeated-OCI jobs. It downloads only their
short-lived metadata/checksum and JSON/report artifacts, assembles a normalized
manifest with `scripts/generate-candidate-evidence.sh`, verifies it offline,
and uploads only:

- `candidate-evidence.json`
- `candidate-evidence-report.json`

The source metadata stays available as the four separate artifacts in the same
workflow run for 14 days. The final two-file candidate artifact is therefore a
digest-bound result and verification report, not a portable release bundle.
Reproducing its verification requires the four source metadata artifacts from
that exact workflow run.

The workflow-scoped downloads provide the same-run guarantee in protected CI.
The offline JSON schemas themselves bind commit, contract, platform and target,
but carry no authenticated GitHub run identifier; manually mixing two otherwise
valid same-commit runs is therefore outside the standalone verifier's claim.

## Security and rollout boundary

This control detects stale, mixed-commit, missing-platform and promoted-claim
evidence before a real release ceremony. It does not authenticate the actor or
runner, make the floating runner/BuildKit/APT inputs hermetic, sign provenance,
publish a binary or image, deploy a node, or approve production. The Phase 7
tagged-binary/container checkbox remains open, rollout remains 35/59, and
`production_ready` remains false.
