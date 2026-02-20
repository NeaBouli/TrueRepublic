# Mobile Wallet Integration

Guide to the TrueRepublic React Native mobile wallet.

## Tech Stack

| Technology | Version | Purpose |
|-----------|---------|---------|
| React Native | 0.74 | Mobile framework |
| Expo | 51.0 | Build toolchain |
| React | 18.2 | UI framework |
| React Navigation | 6.5 | Navigation (bottom tabs) |
| CosmJS | 0.32-0.38 | Blockchain interaction |

## Project Structure

```
mobile-wallet/
├── src/
│   └── screens/
│       ├── WalletScreen.js       # Balance + send PNYX
│       ├── GovernanceScreen.js   # Domains + voting
│       └── DexScreen.js          # Token swaps
├── App.js                        # Navigation setup
├── app.json                      # Expo configuration
└── package.json                  # Dependencies
```

## Running the App

```bash
cd mobile-wallet
npm install

# Start Expo dev server
npm start

# Platform-specific
npm run android
npm run ios
```

## Navigation

The app uses a bottom-tab navigator:

```
┌──────────────────────────────┐
│                              │
│        Screen Content        │
│                              │
├──────────────────────────────┤
│  Wallet  │ Governance │ DEX  │
└──────────────────────────────┘
```

## Screens

### WalletScreen
- Connect wallet
- Display balance
- Send PNYX transactions

### GovernanceScreen
- Browse domains
- View issues and suggestions
- Submit proposals and vote

### DexScreen
- View liquidity pools
- Swap tokens

## CosmJS Integration

The mobile wallet uses the same CosmJS libraries as the web wallet:

```javascript
import { SigningStargateClient } from "@cosmjs/stargate";

const RPC_ENDPOINT = "https://rpc.truerepublic.network";

// Connect to chain
const client = await SigningStargateClient.connect(RPC_ENDPOINT);

// Query balance
const balance = await client.getBalance(address, "pnyx");
```

## Building for Production

### Android

```bash
# Build APK
npx expo build:android

# Or EAS Build
npx eas build --platform android
```

### iOS

```bash
# Build IPA
npx expo build:ios

# Or EAS Build
npx eas build --platform ios
```

## Differences from Web Wallet

| Feature | Web | Mobile |
|---------|-----|--------|
| Wallet | Keplr browser extension | In-app key management |
| Layout | Three-column desktop | Single-column with tabs |
| Styling | Tailwind CSS | React Native StyleSheet |
| Navigation | React Router | React Navigation |
| Build | Webpack (react-scripts) | Metro (Expo) |
