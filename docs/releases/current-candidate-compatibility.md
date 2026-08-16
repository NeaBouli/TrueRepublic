# Current recovery candidate compatibility statement

Contract: `truerepublic.release-compatibility/v1`. Candidate state: unreleased;
Production: `false`; Tagged: `false`; Published: `false`; Signed: `false`.

Supported daemon targets are `linux-amd64` and `linux-arm64` on the contracted
`linux-glibc` runtime. `client-web` is the maintained Beta. `v0.4.0` labels the
recovery line and client package, while the exact daemon source commit is the
candidate binary identity; the March `v0.4.0` tag is historical.

## Breaking and bounded changes

| ID | Classification | Required action |
|---|---|---|
| `COMPAT-001` | `breaking_chain_state` | Use only a separately reviewed fresh-chain or migration procedure. |
| `COMPAT-002` | `breaking_api` | Replace standard staking/distribution calls with documented PoD surfaces. |
| `COMPAT-003` | `breaking_client` | Use `client-web`; retired clients remain Git-history-only. |
| `COMPAT-004` | `breaking_api` | Use registered protobuf gRPC methods over the maintained transport. |
| `COMPAT-005` | `breaking_client` | Keep anonymous proof submission disabled. |
| `COMPAT-006` | `breaking_operations` | Use verified lifecycle install and mandatory pre-start. |
| `COMPAT-007` | `breaking_operations` | Verify source commit and use only the exact governed `v0.4.1` path. |
| `COMPAT-008` | `compatible` | Treat ICS-20 evidence as local and bounded, not external qualification. |

## Explicitly unsupported surfaces

- `x_staking`
- `x_distribution`
- `ibc_staking`
- `ibc_client_upgrade`
- `external_relayers`
- `anonymous_zkp_submission`
- `legacy_custom_abci_queries`

The machine-readable contract is authoritative for repository evidence and
operator/user actions. This statement is not an artifact, support warranty,
deployment authorization or approval to use real keys or funds.
