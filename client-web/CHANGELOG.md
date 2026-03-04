# Changelog

## [0.4.0] - 2026-03-04

### Added

**Week 1: Wallet Foundation**
- Wallet creation with 24-word mnemonic
- Wallet import (12/24 words)
- AES-GCM encryption with PBKDF2
- Balance display (PNYX + IBC assets)
- Lock/unlock functionality

**Week 2: Transactions & Governance**
- Send PNYX with gas estimation
- Domain browser
- Issue list with status badges
- Suggestion display with ratings

**Week 3: DEX Swap**
- Asset selection (PNYX-paired pools)
- Multi-hop swap routing
- Slippage tolerance settings
- Real-time swap estimation

**Week 4: ZKP Anonymous Voting**
- Identity creation/import/export
- Mock proof generation (2s, ready for gnark-wasm)
- Voting panel with rating slider (-5 to +5)
- Nullifier-based double-vote prevention

**Week 5: Domain Membership**
- Invite link parsing (truerepublic://)
- 2-step onboarding flow
- Membership status badges

**Week 6: Advanced Governance**
- Create suggestions with PayToPut fee
- Place stones (endorsements)
- Stone count display
- Dual modals (vote + stones)

**Week 7: DEX Liquidity**
- Add liquidity (proportional deposits)
- Remove liquidity (share redemption)
- LP position value calculator
- Pool list with liquidity actions

**Week 8: Admin Dashboard**
- Domain statistics (members, issues, treasury)
- Member management (MsgAddMember)
- Invite link generation
- Admin-only access control

**Week 9: Network Explorer**
- Real-time network stats (CometBFT RPC)
- PoD validator list (stake, power, domains)
- Recent blocks with auto-refresh (10s)
- IBC channel status

**Week 10: Final Polish**
- Mobile navigation (floating menu)
- Error boundary (global error handling)
- Loading skeletons
- Toast notification system
- Slide-in animations

### Tech Stack
- React 18.2 + TypeScript 5.9 + Vite 7.3
- CosmJS 0.32.4
- Zustand 4.5 (state management)
- TailwindCSS 3.4
- React Router v6

### Known Limitations
- ZKP: Mock implementation (SHA-256, 2s proof gen)
  - Real gnark-wasm planned for v0.4.1
- Transaction history: Not yet implemented
- IBC: Query support only (no cross-chain transfers in UI)
