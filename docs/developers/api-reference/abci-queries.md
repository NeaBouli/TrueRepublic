# Module Queries (gRPC)

TrueRepublic supports custom-module reads through the `truerepublicd query`
commands and the protobuf gRPC services registered on the private operator gRPC
listener. The CLI uses the same gRPC query clients.

## Supported clients

- `truerepublicd query truedemocracy ...`
- `truerepublicd query dex ...`
- protobuf gRPC clients using the exact service methods below
- maintained `client-web` services using typed protobuf requests through the
  existing CometBFT JSON-RPC `abci_query` transport and `/rpc/` reverse proxy
- in-process Go integrations using `truedemocracy.NewQueryClient` or
  `dex.NewQueryClient`

The maintained browser client does not consume the retired legacy query surface
or unregistered custom-module HTTP aliases. Module-query failures remain errors;
they are not converted into empty lists, false authorization/nullifier state, or
fallback economic parameters.

## truedemocracy service

Service: `truedemocracy.Query`

| Method path | Request fields | Result |
|-------------|----------------|--------|
| `/truedemocracy.Query/Domain` | `name` | One domain as JSON bytes |
| `/truedemocracy.Query/Domains` | none | All domains as JSON bytes |
| `/truedemocracy.Query/Validator` | `operator_addr` | One validator as JSON bytes |
| `/truedemocracy.Query/Validators` | none | All validators as JSON bytes |
| `/truedemocracy.Query/Nullifier` | `domain_name`, `nullifier_hash` | Nullifier-use result as JSON bytes |
| `/truedemocracy.Query/PurgeSchedule` | `domain_name` | Purge schedule as JSON bytes |
| `/truedemocracy.Query/ZKPState` | `domain_name` | ZKP domain state as JSON bytes |
| `/truedemocracy.Query/MerkleProof` | `domain_name`, `commitment` | Verified Merkle membership path as JSON bytes |
| `/truedemocracy.Query/PayToPut` | `domain_name` | Canonical current proposal fee calculation as JSON bytes |

CLI examples:

```bash
truerepublicd query truedemocracy domains
truerepublicd query truedemocracy domain Climate
truerepublicd query truedemocracy validator truerepublic1...
truerepublicd query truedemocracy nullifier Climate <hash>
truerepublicd query truedemocracy merkle-proof Climate <commitment>
truerepublicd query truedemocracy pay-to-put Climate
```

## DEX service

Service: `dex.Query`

| Method path | Request fields | Result |
|-------------|----------------|--------|
| `/dex.Query/Pool` | `asset_denom` | One pool as JSON bytes |
| `/dex.Query/Pools` | none | All pools as JSON bytes |
| `/dex.Query/RegisteredAssets` | none | Registered assets as JSON bytes |
| `/dex.Query/AssetByDenom` | `ibc_denom` | One asset as JSON bytes |
| `/dex.Query/AssetBySymbol` | `symbol` | One asset as JSON bytes |
| `/dex.Query/EstimateSwap` | `input_denom`, `input_amt`, `output_denom` | Route and expected output as JSON bytes |
| `/dex.Query/PoolStats` | `asset_denom` | Pool statistics as JSON bytes |
| `/dex.Query/SpotPrice` | `input_denom`, `output_denom` | Price and route as JSON bytes |
| `/dex.Query/LiquidityDepth` | `input_denom`, `output_denom` | Slippage-depth levels as JSON bytes |
| `/dex.Query/LPPosition` | `asset_denom`, `shares` | Underlying LP values as JSON bytes |

CLI examples:

```bash
truerepublicd query dex pools
truerepublicd query dex pool atom
truerepublicd query dex registered-assets
truerepublicd query dex estimate-swap upnyx 1000000 atom
```

## HTTP and legacy compatibility boundary

grpc-gateway HTTP routes are not registered for custom modules. Port 1317
exposes the registered standard Cosmos SDK gateway routes, not HTTP aliases for
the services above.

The compatibility-only legacy `custom/...` ABCI surface was removed under
GH-116 after both prototype clients that referenced it were retired. It is not
versioned or supported, and calls fail closed through the normal Cosmos SDK
unknown-query response. Integrations must use the CLI, registered protobuf gRPC
methods, or the maintained browser's `ModuleQueryClient`, which sends those
same typed requests through the supported CometBFT JSON-RPC `abci_query`
transport.

The browser boundary sends protobuf request bytes for the registered methods
above through CometBFT JSON-RPC `abci_query` on the configured RPC endpoint. The
response `result.response.value` is the base64 protobuf `Query*Response`; its
field 1 contains the module's JSON result bytes. This reuses the existing
same-origin `/rpc/` proxy and does not expose the private gRPC listener or add a
custom grpc-gateway surface. A non-zero ABCI code, transport failure, or invalid
protobuf/JSON response is an unavailable/error result, never an authoritative
empty result. This is a trusted-RPC read boundary, not a browser light-client
proof: operators and users must configure an RPC endpoint they trust.
