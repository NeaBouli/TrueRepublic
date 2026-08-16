# TrueRepublic Web Client

React-based web client for the TrueRepublic/PNYX blockchain.

## Quick Start

```bash
npm ci
npm run dev      # Development server (http://localhost:3001)
npm run build    # Production build
```

## Tech Stack

- **React 18** + TypeScript 5.9
- **Vite 8.2** (build tooling)
- **CosmJS 0.39** (blockchain interaction)
- **Zustand 4.5** (state management)
- **TailwindCSS 3.4** (styling)
- **React Router v7** (declarative SPA routing)
- **Heroicons** (icons)

## Architecture

```
src/
├── components/
│   ├── auth/           Wallet create/import/unlock
│   ├── wallet/         Dashboard, balances, send
│   ├── governance/     Domains, issues, suggestions, stones
│   ├── dex/            Swap, liquidity, pools, positions
│   ├── zkp/            Identity, anonymous voting
│   ├── membership/     Invites, onboarding
│   ├── admin/          Domain management dashboard
│   ├── network/        Explorer, validators, blocks
│   ├── ibc/            Native transfer and evidence recovery
│   └── common/         Button, Card, Input, Toast, etc.
├── services/           Blockchain query & tx services
├── stores/             Zustand state stores
├── types/              TypeScript type definitions
├── config/             Chain configuration
└── utils/              Formatting, clipboard helpers
```

## Key Components

### Services
- `WalletService`: Create/import wallets, encryption
- `BlockchainService`: Balance queries, account info
- `TransactionService`: Send transactions
- `GovernanceService`: Domain/issue/suggestion queries
- `GovernanceTxService`: Create suggestions, place stones
- `DEXService`: Pool queries, swap estimates, LP positions
- `DEXTxService`: Swap, add/remove liquidity transactions
- `MembershipService`: Onboarding, domain membership
- `AdminService`: Domain management, member verification
- `NetworkService`: Chain statistics, validators, blocks
- `IbcTransferService`: Canonical native ICS-20 transfer and source evidence

### Stores (Zustand)
- `walletStore`: Current wallet, balances, lock state
- `governanceStore`: Domains, issues, suggestions
- `dexStore`: Pools, assets, swap estimates
- `identityStore`: Anonymous identity (persisted)
- `membershipStore`: Domain memberships
- `adminStore`: Admin status, members, stats
- `networkStore`: Network info, validators, blocks
- `ibcTransferStore`: Wallet/chain-scoped non-secret transfer recovery records

## ZKP Implementation

**v0.4.0**: Preview helpers with fail-closed proof submission
- Local identity commitment creation/export/import remains preview-only
- `initialize()` and `generateProof()` reject because no compatible real
  prover is installed
- Mock-derived values are never submitted as Groth16 proofs

**v0.4.1 (Planned)**: Real ZKP
- gnark-wasm compilation from Go circuit
- Groth16 proof generation
- MiMC hashing (matching on-chain circuit)

## Security

- Wallet encryption: versioned AES-GCM (256-bit) payloads
- Key derivation: PBKDF2-HMAC-SHA-256 (600k iterations); authenticated legacy
  100k payloads are re-encrypted after a successful unlock
- Mnemonic: bounded English BIP-39 validation (12 or 24 words)
- Derivation path: m/44'/118'/0'/0/0 (Cosmos)
- Storage: localStorage (encrypted)

The derived account must exactly match the selected stored address, every
signing endpoint must report the configured chain ID, and session-bound signers
reject after lock, switch, active-wallet deletion, or reload. This protects the
qualified local browser flow; it does not defend a compromised same-origin
runtime, replace hardware custody, or authorize real keys or funds.

## Chain Configuration

The development defaults point directly to a local node:
- RPC: `http://localhost:26657`
- REST: `http://localhost:1317`
- Bech32: `truerepublic1...`
- Base denom: `upnyx`; display denom: `PNYX` (6 decimals)

The production container sets `VITE_TRUEREPUBLIC_RPC=/rpc` and
`VITE_TRUEREPUBLIC_REST=/api`, resolved against the browser origin. Its nginx
proxy reaches the loopback-only node APIs through the shared Compose network
namespace, so ports 26657 and 1317 remain unexposed.

## Supported transaction boundary

Every signing path uses the single fail-closed registry in
`src/services/txRegistry.ts` and the simulate/sign/broadcast boundary in
`src/services/signingClient.ts`. In addition to CosmJS' standard messages, the
maintained client supports exactly these custom message identities:

- `truedemocracy`: `MsgCreateDomain`, `MsgSubmitProposal`,
  `MsgPlaceStoneOnSuggestion`, `MsgPlaceStoneOnIssue`,
  `MsgApproveOnboarding`, `MsgAddMember`, `MsgOnboardToDomain`, and
  `MsgRegisterIdentity`
- `dex`: `MsgAddLiquidity`, `MsgRemoveLiquidity`, and `MsgSwapExact`
- `ibc`: upstream canonical `MsgTransfer` for native `upnyx` only

Unknown, legacy, or mismatched type URLs are rejected before signing. Integer
token amounts remain decimal strings until protobuf encoding. Anonymous proof
generation remains disabled as described above.

The `/ibc/transfer` route accepts only schema-validated open unordered ICS-20
channels, verifies a fresh balance plus fee reserve, derives bounded timeouts
from the channel client state, and derives the sender from the unlocked signer.
Its persisted ledger contains no password, mnemonic, signer, or key. Broadcast
is not shown as delivery: users explicitly check source-chain acknowledgement
or timeout evidence by transaction hash, and the client never auto-resubmits.

Run the deterministic registry tests with `npm test -- --run`. The secret-free
local chain proof uses generated, disposable accounts in a temporary genesis:

```bash
make build
cd client-web
TRUEREPUBLIC_CLIENT_CHAIN_INTEGRATION=1 npm run test:chain
```

The canonical Bech32 prefix is `truerepublic`. Wallets created by an older
preview with the incorrect `true` prefix are migrated locally to the canonical
`truerepublic` address without changing its encrypted wallet payload; never
send funds to an address solely because an obsolete preview displayed it.

## Build Output

Current recovery build (`npm run build`):
- Initial JavaScript entry: 71.15 kB gzip using the pinned
  project measurement
- 20 lazy page-route entries; largest direct route entry: 4.91 kB gzip
- 355.07 kB total JavaScript (gzip across all deferred chunks)
- CSS: 22.38 kB; 4.86 kB gzip
- `npm run build` fails if the raw or gzip entry, route, chunk, or total budget
  regresses. Signing and protobuf dependencies remain deferred until needed.

## License

The repository currently has no root `LICENSE` file. `CONTRIBUTING.md` describes
Apache 2.0 intent, but GH-215 Alpha dependency/code adoption remains blocked
until maintainers publish the exact project license and compatibility review.
