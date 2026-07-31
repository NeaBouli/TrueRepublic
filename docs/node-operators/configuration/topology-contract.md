# Multi-Node Topology Qualification Contract

Status: repository qualification for GH-89. This validator does not deploy a
network, inspect a server, change a firewall, authorize production, or prove
that a declared topology is running.

The GH-71 [role-based network policy](network-policy.md) validates one
initialized node home. The topology contract adds the missing cross-node
boundary: it verifies that all intended seed, sentry, validator, and RPC roles
form one explicit graph with reciprocal sentry protection, distinct failure
zones, deny-by-default flows, and bounded public query ingress.

## Committed example and private operator contract

The maintained example is
[`configs/topology/qualification.example.json`](../../../configs/topology/qualification.example.json).
It contains only synthetic 40-character node IDs and `.invalid` hostnames.
It must never be replaced with a real operator inventory.

A real contract can correlate validator P2P identities with sentries. Although
these are public node IDs rather than private keys, publishing that correlation
would defeat the sentry privacy boundary. Keep real contracts in an
operator-controlled evidence store outside the public repository. Never place
private validator keys, signer state, node keys, mnemonics, credentials,
tokens, provider identifiers, or complete node homes in a contract.

## Offline validation

```bash
truerepublicd topology-policy validate \
  --file configs/topology/qualification.example.json
```

Automation can request a deterministic, secret-free report:

```bash
truerepublicd topology-policy validate \
  --file topology.json \
  --output json
```

The command reads one strict JSON document of at most 256 KiB. Duplicate keys,
unknown fields, trailing values, excessive nesting, and unsupported versions
fail before policy evaluation. Values are never expanded from environment
variables and hostnames are never resolved.

## Version 1 contract

`version` must be `truerepublic.topology/v1`. The contract contains:

- a logical `chain_id`;
- explicit `deny` defaults for inbound and outbound traffic;
- nodes with canonical names, roles, zones, public node IDs, optional public
  P2P endpoints, dialed peers, and sentry protection relationships;
- explicit directed flows carrying only `p2p`, `rpc`, or `api`;
- separate RPC and API ingress policies.

The schema intentionally has no key, credential, environment, command,
template, or provider configuration field.

## Topology invariants

The validator requires at least one seed, sentry, validator, and RPC node.
Every logical name, public node ID, and public endpoint is unique.

- A validator has no public P2P endpoint, dials at least two declared sentries
  in distinct zones, and dials no seed, RPC node, or validator.
- Each validator sentry reciprocally declares that validator under
  `protects`. The sentry must not list the validator as a dialed peer.
- Seed and sentry nodes dial only declared seeds or sentries.
- RPC nodes dial only declared seeds or sentries.
- Public P2P is limited to seed, sentry, and RPC roles.
- Every peer and protection relationship has a matching explicit P2P flow.
- Every flow is backed by a declared relationship; implicit external peers
  and undeclared node-to-node RPC/API paths fail.
- Internet traffic reaches only a declared public P2P endpoint or TLS
  proxy-only RPC/API ingress. It never reaches a validator.

These checks correlate logical intent. Before a node starts, its effective
`config.toml` and `app.toml` must still pass the GH-71 per-home command:

```bash
truerepublicd network-policy validate \
  --role validator \
  --home "$HOME/.truerepublic"
```

## Abuse-control ceilings

Enabled public ingress must be TLS-only and proxy-only, with an explicit route
allowlist. The contract accepts stricter limits but rejects values weaker than:

| Surface | Request rate | Burst |
|---|---:|---:|
| CometBFT RPC | 10 requests/second | 20 |
| REST API | 30 requests/second | 50 |

Both surfaces also require:

- request bodies between 1 byte and 1 MiB;
- timeouts between 1 and 30 seconds;
- a positive concurrency bound no greater than 1,024;
- explicit HTTP method allowlists and exact non-root path prefixes without
  wildcards or parent traversal;
- WebSocket upgrades only on the explicitly allowlisted RPC `/websocket` path;
- metrics, admin, unsafe RPC, debug, and profiling surfaces disabled.

These ceilings are qualification policy, not DDoS protection or
authorization. The production proxy, provider controls, and load assumptions
remain separately reviewed rollout gates.

## Qualification and evidence procedure

1. Create the real contract only in an operator-private workspace. Use logical
   zone labels; do not add provider account identifiers.
2. Review all node IDs and relationships out of band with two operators.
3. Run `topology-policy validate` and preserve only its report, contract
   checksum, reviewer identities, and timestamp in the rollout evidence store.
4. On every initialized node, run the GH-71 role validator against the
   effective home. Do not upload the home or its configuration bundle.
5. Compare the contract's directed flow set with both provider and host
   firewall review output. This repository does not generate or apply rules.
6. Exercise the topology first in an authorized private testnet with telemetry,
   failure drills, and rollback ownership.
7. Treat any mismatch, missing evidence, weak ingress value, or unexpected
   listener as a failed rollout gate.

## Change, rollback, and isolation

Topology changes require a new contract version or checksum and a complete
revalidation; editing the report is not evidence.

On a failed change:

1. remove the new public route at the outer proxy or firewall;
2. stop the affected node cleanly;
3. restore only the reviewed topology and node configuration;
4. never roll validator signer state backward or start a second signer;
5. repeat topology and per-home validation before any separately authorized
   retry.

If a validator identity or forbidden surface becomes public, isolate it at the
outer boundary, preserve single-signer custody, and follow the
[validator identity](../operations/validator-identity-recovery.md) or
[consensus-key rotation](../operations/validator-key-rotation.md) runbook.

## What GH-89 does not prove

Passing the contract proves only repository-owned logical qualification. It
does not prove actual host placement, provider separation, DNS, TLS,
firewall state, sustained capacity, DDoS resistance, operational rehearsal,
or a production deployment. The corresponding GH-29 Phase 6 deployment
checkbox remains open until separately authorized live evidence exists.
