# Capacity Qualification

Status: synthetic loopback regression evidence only. This qualification does
not establish production sizing, authorize a public network, or approve real
keys, identities, funds, hosts, or operator inventory.

## Maintained qualification

The repository-owned contract at
`configs/capacity/qualification.example.json` fixes one reproducible slice:

- exactly four temporary validators with equal bank-backed voting power;
- 96 deterministic transactions submitted as 24 consecutive waves of four
  parallel signers, committed without failure across at least 24 blocks;
- default pruning, snapshots every two blocks, and at most three retained
  snapshots;
- a private application metrics endpoint configured with a 60-second SDK
  retention window and a private CometBFT metrics endpoint;
- a statically repository-verified local Compose `json-file` configuration of
  50 MiB per file and three retained files; and
- explicit disk-growth, log-growth, RSS, goroutine, duration, and commit-
  latency envelopes.

The harness measures the temporary node homes and process logs before and
after the workload, samples each daemon's resident memory, verifies required
application and consensus metrics, restarts one validator, requires common
app hashes and validator powers, and validates an exported bank-backed ledger.
Its checked one-million-block disk projection is a regression calculation from
this short run, not a storage forecast or hardware recommendation. The
workload is bounded concurrent multi-block pressure with measured milli-TPS,
average block time, and commit latency. It is not a production traffic model.

## Run locally

Prerequisites are Go 1.26.6, the repository's CGO/wasmvm build prerequisites,
permission to bind temporary loopback ports, and enough temporary disk space
for four node homes.

```bash
evidence_dir="$(mktemp -d)"
evidence="$evidence_dir/capacity-evidence.json"
TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 \
TRUEREPUBLIC_CAPACITY_EVIDENCE_OUT="$evidence" \
  go test . -run '^TestMultiValidatorCapacitySustainedLoad$' \
  -count=1 -timeout=900s -v

go run . capacity-policy verify \
  --contract configs/capacity/qualification.example.json \
  --evidence "$evidence" --output json
```

`TRUEREPUBLIC_CAPACITY_EVIDENCE_OUT` is optional. When supplied, it must name a
new file; the test creates it exclusively with mode `0600` and refuses to
overwrite an existing file. The evidence contains measurements and booleans,
not node homes, process arguments, keys, addresses, transaction hashes, logs,
or endpoint URLs. Treat any exported file as private operational evidence.

The maintained plan requires an operator to abort on any transaction failure,
stalled consensus, app-hash or validator-power divergence, ledger-invariant
failure, resource-envelope breach, unbounded retention, arithmetic overflow,
or detected secret material. The verifier structurally limits its evidence to
fixed identifiers, measurements, and booleans; it is not a general log DLP
scanner. Do not waive a failed assertion or substitute a manual claim.

## CI gate

GitHub Actions validates the strict contract in the normal build job and runs
the real-process workload in the separate `capacity-qualification` job. The
job writes evidence only beneath the ephemeral runner directory and verifies
that exact evidence with `capacity-policy verify` before it can pass.

## Not proven

This bounded test does not prove multi-day soak behavior, peak or adversarial
traffic, state growth over real chain history, Docker's actual log-rotation
implementation, Prometheus TSDB retention or cardinality under production
queries, snapshots usable by an external recovery target, cloud volume IOPS,
network bandwidth, sentry/RPC sizing, geographic latency, autoscaling, cost,
or alert delivery. Those require a reviewed non-production environment and
operator-owned measurements before rollout. Production topology and live
deployment remain separate gates in
[Issue #29](https://github.com/NeaBouli/TrueRepublic/issues/29).

See also [Monitoring](monitoring.md) and the
[Road to Rollout](../../ROLLOUT_ROADMAP.md).
