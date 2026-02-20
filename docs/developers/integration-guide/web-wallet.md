# Web Wallet Integration

How the TrueRepublic web wallet is built and how to extend it.

## Tech Stack

| Technology | Purpose |
|-----------|---------|
| React 18 | UI framework |
| React Router 6 | Client-side routing |
| Tailwind CSS 3.4 | Styling |
| CosmJS 0.32-0.38 | Blockchain interaction |
| Keplr Wallet | User key management |
| react-app-rewired | Webpack config overrides |

## Project Structure

```
web-wallet/
├── src/
│   ├── index.js              # Entry point, router setup
│   ├── index.css             # Tailwind directives + globals
│   ├── App.js                # Main governance layout
│   ├── components/
│   │   ├── ThreeColumnLayout.js   # Responsive 3-column layout
│   │   ├── Header.js             # Navigation + wallet connect
│   │   ├── DomainList.js         # Left sidebar: domain list
│   │   ├── ProposalFeed.js       # Center: issues + suggestions
│   │   └── DomainInfo.js         # Right sidebar: stats + forms
│   ├── pages/
│   │   ├── Wallet.js             # Send/receive PNYX
│   │   ├── Dex.js                # Token swaps
│   │   └── Governance.js         # Redirect to /
│   ├── hooks/
│   │   └── useWallet.js          # Keplr wallet management
│   └── services/
│       └── api.js                # Blockchain API abstraction
├── public/
│   ├── index.html
│   └── logo.svg
├── config-overrides.js       # Webpack Node.js polyfills
├── tailwind.config.js        # Custom color theme
├── postcss.config.js
└── package.json
```

## Connecting to Keplr

The `useWallet` hook handles all wallet interactions:

```javascript
import useWallet from "./hooks/useWallet";

function MyComponent() {
    const wallet = useWallet();

    return (
        <div>
            {wallet.connected ? (
                <p>Connected: {wallet.address}</p>
            ) : (
                <button onClick={wallet.connect}>Connect</button>
            )}
        </div>
    );
}
```

### Hook API

| Property/Method | Type | Description |
|----------------|------|-------------|
| `address` | string | Connected wallet address |
| `balance` | object | `{ amount, denom }` |
| `connected` | boolean | Whether wallet is connected |
| `loading` | boolean | Connection in progress |
| `error` | string | Error message |
| `connect()` | function | Connect to Keplr |
| `disconnect()` | function | Disconnect wallet |
| `refreshBalance()` | function | Manually refresh balance |

## API Service

The `api.js` service centralizes all blockchain interactions:

```javascript
import { fetchDomains, submitProposal, castVote } from "./services/api";

// Query domains
const domains = await fetchDomains();

// Submit a proposal
const result = await submitProposal(
    senderAddress, "Climate", "Carbon Tax", "Implement 5% carbon levy"
);

// Cast a vote
const result = await castVote(
    senderAddress, "Climate", "Carbon Tax", "Implement 5% carbon levy", 3
);
```

### Available Functions

| Function | Parameters | Description |
|----------|-----------|-------------|
| `getQueryClient()` | none | Read-only Stargate client |
| `getSigningClient()` | none | Signing client (requires Keplr) |
| `fetchDomains()` | none | Get all domains |
| `submitProposal(sender, domain, issue, suggestion)` | 4 strings | Submit proposal |
| `castVote(sender, domain, issue, suggestion, stones)` | 4 strings + number | Cast vote |
| `getBalance(address)` | address string | Get PNYX balance |
| `sendTokens(sender, recipient, amount)` | 3 strings | Send PNYX |
| `fetchPools()` | none | Get all DEX pools |
| `swapTokens(sender, inputDenom, inputAmt, outputDenom)` | 4 args | Swap tokens |

## Webpack Configuration

The `config-overrides.js` provides Node.js polyfills required by CosmJS:

```javascript
// Polyfills enabled:
// - crypto-browserify (for @cosmjs/crypto)
// - stream-browserify (for @cosmjs/stargate)
// - buffer (global Buffer)

// Disabled (set to false):
// - vm, path, os, fs, http, https, zlib, url, assert
```

## Adding a New Feature

### 1. Add API Function

In `src/services/api.js`:

```javascript
export async function myNewFunction(sender, param1, param2) {
    const client = await getSigningClient();
    const msg = {
        typeUrl: "/truedemocracy.MsgMyNewMessage",
        value: { sender, param1, param2 },
    };
    return client.signAndBroadcast(sender, [msg], "auto");
}
```

### 2. Create Component

In `src/components/MyFeature.js`:

```javascript
import React from "react";

export default function MyFeature({ data, onAction, connected }) {
    return (
        <div className="bg-dark-800 border border-dark-700 rounded-xl p-4">
            {/* Your UI here */}
        </div>
    );
}
```

### 3. Integrate

In `App.js` or relevant page, import and use the component.

## Design System

### Colors

The Tailwind config defines two custom palettes:

- **`republic-*`** -- Blues for primary actions and accents (50-950)
- **`dark-*`** -- Slate grays for backgrounds and text (50-950)

### Common Patterns

```jsx
{/* Card */}
<div className="bg-dark-800 border border-dark-700 rounded-xl p-4">

{/* Primary button */}
<button className="px-4 py-2 bg-republic-600 text-white rounded-lg hover:bg-republic-700">

{/* Input field */}
<input className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-200" />

{/* Label */}
<label className="block text-xs font-medium text-dark-400 mb-1">

{/* Section header */}
<h2 className="text-sm font-semibold text-dark-400 uppercase tracking-wider mb-3">
```

## Building for Production

```bash
cd web-wallet
npm run build
# Output: build/ directory ready for static hosting
```

The Docker setup serves the build via nginx with proper SPA fallback routing.
