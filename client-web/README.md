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
- **Vite 8.1** (build tooling)
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

### Stores (Zustand)
- `walletStore`: Current wallet, balances, lock state
- `governanceStore`: Domains, issues, suggestions
- `dexStore`: Pools, assets, swap estimates
- `identityStore`: Anonymous identity (persisted)
- `membershipStore`: Domain memberships
- `adminStore`: Admin status, members, stats
- `networkStore`: Network info, validators, blocks

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

- Wallet encryption: AES-GCM (256-bit)
- Key derivation: PBKDF2 (100k iterations, SHA-256)
- Mnemonic: BIP39 (24 words)
- Derivation path: m/44'/118'/0'/0/0 (Cosmos)
- Storage: localStorage (encrypted)

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

Unknown, legacy, or mismatched type URLs are rejected before signing. Integer
token amounts remain decimal strings until protobuf encoding. Anonymous proof
generation remains disabled as described above.

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
- Main JavaScript: 1,708.31 kB; 318.91 kB gzip
- CSS: 22.25 kB; 4.84 kB gzip
- Vite still warns that the main chunk exceeds 500 kB; bundle splitting and a performance budget remain open rollout work.

## License

See main repository LICENSE.
