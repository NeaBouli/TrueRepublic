# Private Deployment Evidence Gate

Status: offline evidence-envelope preparation only. This procedure does not
deploy, probe, approve, or prove a production network. The GH-29 Phase 6
production-topology checkbox remains open until separately authorized private
infrastructure exists and its source artifacts are independently inspected.

## Purpose

GH-71 validates one effective node home, and GH-89 validates the intended
cross-node graph. The `deployment-evidence` command binds a compact manifest to
the exact raw bytes of that private GH-89 contract and checks that every
required rollout observation has a fresh, distinct SHA-256 evidence digest.
The public report contains only counts and fixed violation categories.

The verifier never reads the referenced evidence artifacts. It does not
resolve DNS, open a socket, inspect a host, contact a provider, mutate a
firewall, configure TLS, or start a node. A passing report establishes only a
well-formed and internally consistent evidence envelope.

## Private inputs

Keep both the real topology contract and deployment manifest in an
operator-controlled evidence store outside the public repository. The manifest
contains logical approval seats, timestamps, counts, and digests only. It must
not contain:

- hosts, IP addresses, URLs, node IDs, validator-to-sentry correlations, or
  provider/account identifiers;
- credentials, tokens, mnemonics, private keys, signer state, node keys, home
  directories, logs, commands, or environment values; or
- raw firewall, DNS, TLS, telemetry, capacity, incident, or inventory output.

Store each source artifact privately and place only its lowercase SHA-256 in
the matching gate. Different gates require different artifacts and digests.

## Required gates

The version-1 manifest requires these eleven entries in this order:

1. `role-policy`
2. `provider-separation`
3. `dns-tls`
4. `provider-firewall`
5. `host-firewall`
6. `listener-exposure`
7. `proxy-abuse-controls`
8. `telemetry`
9. `capacity`
10. `incident-rehearsal`
11. `two-person-review`

Every gate records `passed`, a bounded start/completion window, and a unique
artifact digest. The verifier rejects missing, duplicated, reordered, failed,
future, stale, malformed, or digest-reused entries.

## Verification

Use the exact private topology contract that was reviewed and the manifest
created for it:

```bash
truerepublicd deployment-evidence verify \
  --contract topology.private.json \
  --manifest deployment-evidence.private.json \
  --output json
```

For reproducible review, pin the evaluation instant explicitly:

```bash
truerepublicd deployment-evidence verify \
  --contract topology.private.json \
  --manifest deployment-evidence.private.json \
  --at 2026-08-01T12:00:00Z \
  --output json
```

The command hashes the exact raw contract bytes, re-runs the GH-89 validator,
and derives chain, node, and role counts itself. Any whitespace or byte change
to the contract changes its digest and invalidates the manifest. `--at` uses
strict UTC `YYYY-MM-DDTHH:MM:SSZ`; normal operator execution should omit it so
the current UTC time is used.

## Two-person review

One preparer and exactly two different approval seats are required. The
approval roles are `operator` and `independent-reviewer`; neither approval seat
may equal the preparer, and both approvals bind the same topology digest.
Logical seats prove separation of manifest claims, not distinct human beings.
The release owner must verify identities and artifact contents out of band.

## Abort and rollback

Abort the rollout gate on any failed command, stale or missing artifact,
unexpected listener, topology digest mismatch, weak approval separation,
firewall/proxy discrepancy, capacity breach, failed rehearsal, or evidence
whose source cannot be independently reproduced.

If a private rollout attempt has already begun, remove newly exposed routes at
the outer boundary, stop affected nodes cleanly, preserve the evidence store,
and return to the last reviewed topology. Never restore stale validator signer
state and never start a second signer. Re-run the GH-71 per-home validator, the
GH-89 topology validator, and this evidence verifier before any separately
authorized retry.

## Retention and publication

Retain the private topology contract, every source artifact, approval record,
manifest, command version, and final report under the operator's evidence
retention policy. Publishing the secret-free report or its aggregate counts is
not a substitute for independent inspection of the private source artifacts.
Do not copy the private manifest or contract into this repository.

The committed `configs/deployment/evidence.example.json` is synthetic and binds
only the `.invalid` GH-89 example. It is CI regression data, not rollout
evidence and not a template carrying real inventory.
