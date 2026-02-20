# Web Wallet Guide

Complete guide to using the TrueRepublic Web Wallet.

## Prerequisites

- A modern browser (Chrome, Firefox, Brave, or Edge)
- Keplr Wallet browser extension

## Step 1: Install Keplr Wallet

1. Visit [keplr.app](https://www.keplr.app/)
2. Click "Install Keplr" for your browser
3. Follow the installation wizard
4. Create a new account or import an existing one
5. **Save your seed phrase securely (24 words)**

> **Warning:** Never share your seed phrase with anyone. Anyone with your seed phrase can access all your funds. Write it on paper and store it safely -- never save it digitally.

## Step 2: Connect to TrueRepublic

1. Open the TrueRepublic web wallet
2. Click **"Connect Wallet"** in the top-right corner
3. Keplr will prompt you to add the TrueRepublic chain -- click **"Approve"**
4. Keplr will ask for permission to connect -- click **"Approve"**
5. Your address and balance appear in the header

The TrueRepublic chain is automatically configured with:

| Parameter | Value |
|-----------|-------|
| Chain ID | `truerepublic-1` |
| RPC | `https://rpc.truerepublic.network` |
| REST | `https://lcd.truerepublic.network` |
| Bech32 Prefix | `truerepublic` |
| Token | PNYX |

## Viewing Your Balance

Your PNYX balance is shown on the **Wallet** page. The balance auto-refreshes every 10 seconds. Click **Refresh** to update manually.

## Sending PNYX

1. Navigate to the **Wallet** page
2. Enter the recipient address (starts with `truerepublic1...`)
3. Enter the amount in PNYX
4. Click **"Send PNYX"**
5. Review the transaction in Keplr and click **"Approve"**
6. A confirmation alert shows the transaction hash

### Transaction Fees

| Type | Fee |
|------|-----|
| Standard transfer | ~0.001 PNYX (auto-calculated) |
| Governance proposal | PayToPut fee (varies by domain) |
| DEX swap | 0.3% swap fee + gas |

## Receiving PNYX

1. Copy your address from the Wallet page (displayed at the top)
2. Share it with the sender
3. Transactions confirm in ~5 seconds (one block)
4. Balance updates automatically

## Navigation

The web wallet has three main sections accessible from the header:

| Section | Description |
|---------|-------------|
| **Governance** | Browse domains, view proposals, vote, submit proposals |
| **Wallet** | View balance, send/receive PNYX |
| **DEX** | Swap tokens, view liquidity pools |

## Security Best Practices

**Do:**
- Keep your seed phrase written on paper in a secure location
- Double-check recipient addresses before sending
- Start with small test transactions
- Use a hardware wallet (Ledger) for large holdings
- Lock Keplr when not in use

**Don't:**
- Share your seed phrase with anyone (not even "support")
- Store your seed phrase in photos, cloud storage, or notes apps
- Click links claiming to be Keplr updates from untrusted sources
- Leave large amounts in browser extension wallets long-term

## Disconnecting

Click **"Disconnect"** in the top-right corner to disconnect your wallet from the web app. This does not affect your funds -- it only removes the session connection.

## Next Steps

- [Governance Tutorial](governance-tutorial.md) -- Join domains and participate
- [DEX Trading Guide](dex-trading-guide.md) -- Swap tokens
- [Troubleshooting](troubleshooting.md) -- Common issues
