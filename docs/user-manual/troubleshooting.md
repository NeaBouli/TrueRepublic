# Troubleshooting

Common issues and solutions for TrueRepublic users.

## Wallet Connection Issues

### Local wallet cannot be opened

**Problem:** The maintained web client cannot open or unlock the local wallet.

**Solutions:**
1. Confirm that the wallet password is correct.
2. Confirm that this browser profile still contains the encrypted local wallet.
3. Use the original mnemonic to import the wallet again if the local entry was removed.
4. Do not enter a real mnemonic while recovery status remains active; use local or explicitly approved test credentials only.

### Wallet action does not continue

**Solutions:**
1. Re-open the wallet with its password.
2. Confirm that the configured local/test node is reachable.
3. Refresh the page and retry only in the approved test environment.
4. Record the browser console error when reporting a reproducible defect.

### Chain connection fails

**Problem:** The maintained client cannot reach the configured TrueRepublic chain.

**Solutions:**
1. Confirm that the local/test node is running and synchronized.
2. Re-check the configured RPC and REST endpoints in the
   [Maintained Web Client Guide](web-wallet-guide.md)

## Balance Issues

### Balance shows 0

**Possible causes:**
1. **Wrong network** -- Ensure you're connected to `truerepublic-1`
2. **Node not synced** -- The RPC node may be catching up
3. **New account** -- You haven't received any PNYX yet
4. **Display delay** -- Click Refresh or wait for auto-refresh (10 seconds)

### Balance doesn't update after transaction

**Solutions:**
1. Click the Refresh button
2. Wait 10 seconds for auto-refresh
3. Check the transaction hash on a block explorer
4. The RPC node may be temporarily behind -- wait a moment

## Transaction Issues

### Transaction failed

**Common causes and fixes:**

| Error | Cause | Fix |
|-------|-------|-----|
| "insufficient funds" | Not enough PNYX for amount + gas | Reduce amount or get more PNYX |
| "account sequence mismatch" | Pending transaction | Wait for pending tx to confirm, then retry |
| "out of gas" | Transaction needs more gas | Review the test transaction's gas limit |
| "signature verification failed" | Key mismatch | Re-open the correct local test wallet |
| "unauthorized" | Wrong sender | Check you're using the correct account |

### Transaction stuck / pending

**Solutions:**
1. Wait -- transactions should confirm within ~5 seconds
2. If stuck, the mempool may be full -- wait a few blocks
3. Check node status: is the RPC endpoint responding?
4. Try refreshing the page and resubmitting

### "Invalid address" error

**Solutions:**
1. TrueRepublic addresses start with `truerepublic1...`
2. Ensure the full address is copied (no truncation)
3. Check for extra spaces before/after the address
4. Verify the address hasn't been mistyped

## Governance Issues

### Can't submit proposal

**Possible causes:**
1. **Not a member** -- You must join the domain first
2. **Insufficient funds** -- PayToPut fee required (treasury / 1000 * min(15, members))
3. **Wallet not connected** -- Connect wallet first
4. **Domain restrictions** -- Domain may have "Only Admin Issues" enabled

### Can't see domains

**Solutions:**
1. Check if the RPC node is responding
2. Refresh the page
3. The domain list loads on page load -- wait for it to complete
4. Check browser console (F12) for error messages

### Vote not recording

**Solutions:**
1. Ensure the local test wallet was opened and the transaction was submitted
2. Check that you received a transaction hash
3. Refresh to see updated counts
4. You may be trying to vote on your own proposal (check domain rules)

## DEX Issues

### Swap failed

**Common causes:**
1. **Insufficient balance** -- Check you have enough of the input token
2. **Pool doesn't exist** -- The trading pair may not have a pool yet
3. **Same input/output** -- Can't swap a token for itself
4. **Amount too large** -- May exceed pool reserves

### High slippage

**Solutions:**
1. Reduce trade size (split into multiple smaller trades)
2. Wait for more liquidity in the pool
3. Check current pool reserves before trading

## Node/Network Issues

### RPC endpoint not responding

**Solutions:**
1. Check your internet connection
2. The public RPC may be experiencing high load -- wait and retry
3. Try an alternative RPC endpoint if available
4. Check [status page](https://status.truerepublic.network) for outages

### Slow block times

**Normal:** Blocks should arrive every ~5 seconds. If slower:
1. The network may have fewer active validators
2. A validator may be experiencing issues
3. This is usually temporary and self-correcting

## Browser Issues

### Page won't load

**Solutions:**
1. Clear browser cache and cookies for the site
2. Disable browser extensions that might interfere
3. Try incognito/private mode
4. Try a different browser

### Tailwind styles not rendering

**Solutions:**
1. Hard refresh: Ctrl+Shift+R (Cmd+Shift+R on Mac)
2. Clear browser cache
3. Check if CSS file loaded (F12 > Network tab)

## Getting Help

If your issue isn't listed here:

1. Check the [FAQ](../FAQ.md)
2. Join the community: [t.me/truerepublic](https://t.me/truerepublic)
3. Open an issue: [github.com/NeaBouli/TrueRepublic/issues](https://github.com/NeaBouli/TrueRepublic/issues)
