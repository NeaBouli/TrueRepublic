# Multi-Validator Recovery Harness

Status: recovery verification only. This harness does not authorize a public
network, production keys, or real funds.

## Proven scope

The GH-32 harness creates four temporary validator homes and uses the daemon to
generate four independent CometBFT Ed25519 validator keys. It then:

1. assembles one identical genesis from the four public keys;
2. creates matching equal-power PoD validators with exact module-bank stake
   backing;
3. starts all validators on isolated RPC and P2P ports with explicit persistent
   peers;
4. proves common-height app-hash agreement;
5. stops one validator and proves the remaining three continue committing;
6. restarts the stopped validator and proves catch-up plus renewed app-hash
   agreement;
7. shuts every process down through SIGINT and exports the recovered state; and
8. validates that the exported PoD claims remain exactly bank-backed.

Private validator keys never enter the shared genesis or repository. They stay
inside temporary node homes. Loopback address-book relaxation and duplicate-IP
permission apply only to the temporary localhost configuration used by the
test; production defaults are unchanged.

## Run locally

Prerequisites:

- Go 1.26.5;
- the repository's CGO/wasmvm build prerequisites; and
- permission to bind temporary loopback RPC and P2P ports.

From the repository root:

```bash
TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 \
  go test . -run '^TestMultiValidatorConsensusRecovery$' \
  -count=1 -timeout=300s -v
```

The test uses temporary directories and ports, performs graceful cleanup, and
prints every validator log when it fails. A failure must not be converted into
a retry-only or non-blocking result.

## CI gate

GitHub Actions runs the same command in the required
`multi-validator-recovery` Go CI job. The normal race/coverage suite keeps the
long-running process harness disabled and exercises its genesis construction
logic as a fast unit regression.

## Not proven

GH-32/GH-41/GH-43/GH-45 are bounded four-validator failure/restart/catch-up,
partition-recovery, trusted state-sync, and sanitized backup/restore slices.
GH-55 separately proves a cold transfer of the coupled consensus key and
current signer state into a stopped recovery home with a new P2P identity.
The following Road to Rollout gates remain open:

- authenticated consensus-key rotation, old-key revocation, and bootstrap
  operator-authority separation;
- consensus-breaking state migration and partially applied migration recovery;
- IBC relayer and cross-chain failure recovery;
- sustained load, public topology, monitoring, and independent operations
  review.

Track the remaining gates in
[Issue #29](https://github.com/NeaBouli/TrueRepublic/issues/29) and the
[Road to Rollout](../../ROLLOUT_ROADMAP.md).
