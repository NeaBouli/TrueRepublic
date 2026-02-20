# DEX Trading Guide

Guide to using TrueRepublic's built-in Decentralized Exchange (DEX) for swapping tokens and providing liquidity.

## Overview

The TrueRepublic DEX uses an **Automated Market Maker (AMM)** model based on the constant-product formula `x * y = k`. All pools are paired with PNYX as the base token.

### Key Features

| Feature | Detail |
|---------|--------|
| **Model** | Constant-product AMM (x * y = k) |
| **Swap Fee** | 0.3% per trade |
| **PNYX Burn** | 1% burned on PNYX output (WP S5) |
| **LP Shares** | Proportional ownership of pool reserves |
| **Supported Pairs** | PNYX/ATOM (more pairs planned) |

## Swapping Tokens

### Via Web Wallet

1. Navigate to the **DEX** page
2. Connect your wallet
3. Select **From** token (e.g., PNYX)
4. Select **To** token (e.g., ATOM)
5. Enter the amount to swap
6. Click **"Swap"**
7. Approve the transaction in Keplr

### Via CLI

```bash
truerepublicd tx dex swap [input-denom] [input-amount] [output-denom] \
    --from mykey --chain-id truerepublic-1

# Example: Swap 1000 PNYX for ATOM
truerepublicd tx dex swap pnyx 1000 atom \
    --from mykey --chain-id truerepublic-1
```

### How the Swap Price is Calculated

The AMM uses the constant-product formula with a 0.3% fee:

```
output = (outReserve * input * 9970) / (inReserve * 10000 + input * 9970)
```

**Example:** Pool has 100,000 PNYX and 50,000 ATOM. You swap 1,000 PNYX:
```
output = (50000 * 1000 * 9970) / (100000 * 10000 + 1000 * 9970)
       = 498,500,000,000 / 1,009,970,000
       = ~493 ATOM
```

### Price Impact

Larger trades have more **price impact** (slippage):
- Small trades (< 1% of pool): Minimal impact
- Medium trades (1-5% of pool): Noticeable slippage
- Large trades (> 5% of pool): Significant slippage -- consider splitting

### PNYX Burn Mechanism (WP S5)

When you swap **to PNYX** (buying PNYX), 1% of the output is **burned** (permanently destroyed):

```
burn_amount = pnyx_output * 1%
you_receive = pnyx_output - burn_amount
```

This creates deflationary pressure on PNYX supply over time, benefiting all PNYX holders.

## Fees

| Fee Type | Rate | Recipient |
|----------|------|-----------|
| Swap fee | 0.3% of input | Stays in pool (benefits LPs) |
| PNYX burn | 1% of PNYX output | Burned (removed from supply) |
| Gas fee | ~0.001 PNYX | Network validators |

## Viewing Pools

### Web Wallet

The DEX page shows all active liquidity pools with:
- Pool pair (e.g., PNYX / ATOM)
- Reserve amounts
- Total burned PNYX

### CLI

```bash
# View all pools
truerepublicd query dex pools

# View specific pool
truerepublicd query dex pool atom
```

### Pool Data Structure

```json
{
  "asset_denom": "atom",
  "pnyx_reserve": "1000000",
  "asset_reserve": "500000",
  "total_shares": "707106",
  "total_burned": "5000"
}
```

## Providing Liquidity

Liquidity providers (LPs) deposit both tokens into a pool and receive **LP shares** representing their proportional ownership.

### Creating a New Pool

```bash
truerepublicd tx dex create-pool [asset-denom] [pnyx-amount] [asset-amount] \
    --from mykey --chain-id truerepublic-1

# Example: Create PNYX/ATOM pool
truerepublicd tx dex create-pool atom 100000 50000 \
    --from mykey --chain-id truerepublic-1
```

Initial LP shares = sqrt(pnyx_amount * asset_amount)

### Adding Liquidity

```bash
truerepublicd tx dex add-liquidity [asset-denom] [pnyx-amount] [asset-amount] \
    --from mykey --chain-id truerepublic-1
```

You must add both tokens in the same ratio as the current pool reserves. LP shares received:

```
shares = min(
    pnyx_added * total_shares / pnyx_reserve,
    asset_added * total_shares / asset_reserve
)
```

### Removing Liquidity

```bash
truerepublicd tx dex remove-liquidity [asset-denom] [shares] \
    --from mykey --chain-id truerepublic-1
```

You receive both tokens proportional to your share:

```
pnyx_returned  = pnyx_reserve  * your_shares / total_shares
asset_returned = asset_reserve * your_shares / total_shares
```

### LP Economics

**Benefits of providing liquidity:**
- Earn a share of the 0.3% swap fee on every trade
- Fees accumulate in the pool, increasing the value of your shares

**Risks:**
- **Impermanent loss** -- If token prices diverge significantly, you may have been better off holding
- Pool reserves shift with every trade

## Impermanent Loss

When the price ratio of the two tokens changes after you deposit, you experience **impermanent loss**. The larger the price change, the larger the loss compared to simply holding.

| Price Change | Impermanent Loss |
|-------------|------------------|
| 1.25x | 0.6% |
| 1.5x | 2.0% |
| 2x | 5.7% |
| 3x | 13.4% |
| 5x | 25.5% |

This loss is "impermanent" because it reverses if the price returns to the original ratio.

## Next Steps

- [Web Wallet Guide](web-wallet-guide.md) -- Managing your PNYX
- [Governance Tutorial](governance-tutorial.md) -- Participating in governance
- [Troubleshooting](troubleshooting.md) -- Common issues
