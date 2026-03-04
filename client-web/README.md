# TrueRepublic Web Client

React-based web client for the TrueRepublic/PNYX blockchain.

## Quick Start

```bash
npm install
npm run dev      # Development server (http://localhost:5173)
npm run build    # Production build
```

## Tech Stack

- **React 18** + TypeScript 5.9
- **Vite 7.3** (build tooling)
- **CosmJS 0.32.4** (blockchain interaction)
- **Zustand 4.5** (state management)
- **TailwindCSS 3.4** (styling)
- **React Router v6** (routing)
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

**v0.4.0**: Mock implementation with SHA-256
- 2-second simulated proof generation
- Identity commitment creation/export/import
- Placeholder for real gnark-wasm

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

Default chain config points to local node:
- RPC: `http://localhost:26657`
- REST: `http://localhost:1317`
- Bech32: `true1...`
- Denom: `pnyx` (6 decimals)

## Build Output

Production build (~3 MB):
- Main bundle: ~2.9 MB (CosmJS + protobuf)
- CSS: ~22 kB (TailwindCSS, purged)
- Gzip: ~690 kB

## License

See main repository LICENSE.
