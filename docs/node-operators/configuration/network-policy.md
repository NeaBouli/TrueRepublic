# Role-Based Network Policy

Status: recovery-verified repository policy. This document does not authorize a
public network, production deployment, firewall change, or use of real keys.

TrueRepublic nodes must be assigned one explicit network role before an
operator treats their configuration as rollout evidence:

| Role | Public P2P | PEX | Required peers | Direct RPC/API/gRPC |
|---|---:|---:|---|---|
| `seed` | Yes | On | Optional seed/persistent peers | Disabled or loopback only |
| `sentry` | Yes | On | Upstream sentry/seed peers; protected validator IDs | Disabled or loopback only |
| `validator` | No | Off | At least two sentry persistent peers | Disabled |
| `rpc` | Yes | On | Trusted persistent peers | Loopback behind a proxy |
| `private` | No | Off | Trusted persistent peers | Disabled or loopback only |

The role validator checks the files actually used by an initialized node. It
does not edit them:

```bash
truerepublicd network-policy validate \
  --role validator \
  --home "$HOME/.truerepublic"
```

Automation can request structured output:

```bash
truerepublicd network-policy validate \
  --role validator \
  --home "$HOME/.truerepublic" \
  --output json
```

A non-zero exit is a rollout failure. Fix every reported configuration path and
rerun the validator before starting the node. Loopback peers can be enabled only
through the Go library used by repository process tests; the operator CLI
deliberately exposes no bypass.

## Canonical peer endpoints

Every seed and persistent peer must use the canonical CometBFT form:

```text
<40-lowercase-hex-node-id>@<dns-name-or-ip>:<1-65535-port>
```

The validator rejects:

- uppercase, short, non-hex, missing, or duplicate node IDs;
- duplicate IDs mapped to different endpoints;
- missing, zero, non-numeric, or out-of-range ports;
- wildcard, unspecified, or loopback peer hosts;
- the local node's own ID in a remote peer list;
- malformed `private_peer_ids` or `unconditional_peer_ids`.

Do not put private keys, consensus keys, signer state, credentials, or complete
node homes into policy files, tickets, logs, or chat.

## Required listener boundary

CometBFT RPC, REST API, gRPC, gRPC-web, and pprof must never listen directly on
`0.0.0.0`, `::`, or a public interface. Bind client-facing services to
loopback and place any deliberately public query service behind a separately
reviewed TLS reverse proxy:

```toml
# config.toml
[rpc]
laddr = "tcp://127.0.0.1:26657"
unsafe = false
cors_allowed_origins = []
pprof_laddr = ""

# app.toml
[api]
enable = true
address = "tcp://127.0.0.1:1317"
enabled-unsafe-cors = false

[grpc]
enable = true
address = "127.0.0.1:9090"

[grpc-web]
enable = false
```

Seed, sentry, and validator roles do not need REST or gRPC-web and must leave
them disabled. A validator also uses `pex = false`, zero general inbound peers,
explicit sentry persistent peers, and matching unconditional sentry IDs. Each
sentry lists the protected validator in both `private_peer_ids` and
`unconditional_peer_ids` so its address is neither gossiped nor evicted by
general peer limits. The validator initiates its sentry connections; sentries
must not require the protected validator as a dial-out persistent peer.

Public P2P roles (`seed`, `sentry`, and `rpc`) bind port 26656 to one explicit
non-loopback IP interface and set an explicit routable `external_address`.
Wildcard, loopback, and DNS listener binds fail validation. `validator` and `private` roles
bind P2P to loopback and dial only their reviewed persistent peers. A validator
requires at least two sentries. Validator/private profiles also reject discovery
seeds and any self ID in peer-protection lists.

When Prometheus is enabled, metrics remain same-host only:

```toml
[instrumentation]
prometheus = true
prometheus_listen_addr = "127.0.0.1:26660"
```

## Firewall matrix

This matrix is the policy source. Provider firewalls and host firewalls must
both implement it; repository validation does not mutate either firewall.

| Role | Source | Destination port | Action |
|---|---|---:|---|
| Seed | Internet | 26656/TCP | Allow |
| Sentry | Internet | 26656/TCP | Allow |
| Validator | Named sentry addresses only | 26656/TCP | Allow |
| RPC node | Internet | 26656/TCP | Allow when participating in PEX |
| Public query proxy | Internet | 443/TCP | Allow |
| Any node | Internet | 26657, 1317, 9090, 9091, 26660/TCP | Deny |
| Private monitoring | Same-host collector only | 26660/TCP | Allow |
| Operator access | Named administration network only | 22/TCP | Allow |
| Any other inbound path | Any | Any | Deny |

Never expose a validator's RPC, REST, gRPC, metrics, profiling, or SSH service
to the public Internet. The validator must not be a seed or public query node.

## Reverse-proxy and rate-limit floor

A public query role terminates TLS at a reverse proxy on port 443 and forwards
only to loopback listeners. The minimum repository policy is:

- CometBFT RPC: 10 requests/second per client, burst 20;
- REST API: 30 requests/second per client, burst 50;
- explicit request-body and upstream timeouts;
- WebSocket upgrade only on the intended RPC path;
- no wildcard CORS, unsafe RPC, pprof, metrics, or admin endpoint;
- forward and retain a trusted client address only through an operator-reviewed
  proxy chain.

Rate limits are abuse containment, not authorization. Transaction validity,
fees, replay protection, and chain authorization remain consensus rules.

## Change and rollback procedure

1. Stop the node cleanly. Never copy or restore stale validator signer state.
2. Record hashes of `config.toml` and `app.toml`; do not archive
   `priv_validator_key.json`, `priv_validator_state.json`, `node_key.json`, or a
   complete home in the change artifact.
3. Apply the reviewed role-specific values to a temporary copy.
4. Run `network-policy validate` against the candidate home.
5. Review host and provider firewall diffs independently before any change.
6. Start one node, verify peers, height, app hash, local RPC, and absence of
   forbidden public listeners, then proceed only under a separate rollout
   authorization.
7. On failure, stop the process and restore only the reviewed configuration
   files. Do not roll validator signer state backward.

## Incident isolation

If a validator address, unexpected listener, or abusive query path becomes
public:

1. remove the public route at the outer firewall or proxy;
2. keep the validator signer single and do not start a replacement copy;
3. verify the current role policy offline;
4. check peer, proxy, and consensus evidence without copying private key files;
5. follow the validator identity or key-rotation runbook if key compromise is
   suspected;
6. require a new explicit rollout decision before restoring public service.

See also:

- [Network Configuration](network-config.md)
- [Node Configuration](node-config.md)
- [Security Hardening](../operations/security.md)
- [Validator Identity Custody and Recovery](../operations/validator-identity-recovery.md)
- [Authenticated Validator Key Rotation](../operations/validator-key-rotation.md)
