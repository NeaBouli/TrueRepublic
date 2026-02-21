# Genesis & Chain Parameters

## Chain Parameters

| Parameter | Value | Description |
|-----------|-------|-------------|
| Chain ID | `truerepublic-1` | Unique chain identifier |
| Bech32 Prefix | `truerepublic` | Address prefix |
| Token Denom | `pnyx` | Native token denomination |
| Coin Decimals | 0 | PNYX is indivisible (whole units) |
| Max Supply | 21,000,000 PNYX | Maximum token supply |
| BIP44 Coin Type | 118 | HD wallet derivation path |

## Consensus Parameters

| Parameter | Value | Description |
|-----------|-------|-------------|
| Block time | ~5 seconds | Target time between blocks |
| Signed blocks window | 100 blocks | Window for uptime tracking |
| Min signed per window | 50% (50 blocks) | Minimum blocks to sign |
| Downtime jail duration | 600 seconds (10 min) | Jail time for downtime |
| Double-sign slash | 5% of stake | Penalty for equivocation |
| Downtime slash | 1% of stake | Penalty for missed blocks |

## Governance Parameters

| Parameter | Value | Description |
|-----------|-------|-------------|
| Approval threshold | 500 basis points (5%) | Min average rating for approval |
| Delete/exclude majority | 6667 basis points (66.67%) | 2/3 majority requirement |
| Default dwell time | 86,400 seconds (1 day) | Time in each lifecycle zone |
| Inactivity timeout | 31,104,000 seconds (~360 days) | Auto-remove inactive members |

## Tokenomics Constants

| Constant | Value | Description |
|----------|-------|-------------|
| CDom | 2 | Domain creation cost factor |
| CPut | 15 | Put price cap |
| CEarn | 1000 | Reward divisor |
| StakeMin | 100,000 PNYX | Minimum validator stake |
| SupplyMax | 21,000,000 PNYX | Maximum total supply |
| ApyDom | 0.25 (25%) | Domain treasury interest APY |
| ApyNode | 0.10 (10%) | Validator staking reward APY |
| Reward interval | 3600 seconds (1 hour) | Reward distribution frequency |

## DEX Parameters

| Parameter | Value | Description |
|-----------|-------|-------------|
| Swap fee | 30 bps (0.3%) | Fee on each swap |
| PNYX burn rate | 100 bps (1%) | Burn on PNYX output |
| AMM model | x * y = k | Constant-product formula |

## Validator Parameters

| Parameter | Value | Description |
|-----------|-------|-------------|
| Min stake | 100,000 PNYX | Minimum to become validator |
| Voting power | stake / StakeMin | Integer division |
| Transfer limit | 10% of domain payouts | Max stake withdrawal (WP S7) |
| Domain requirement | >= 1 domain | Must be member of a domain |

## Genesis State

The default genesis includes:

```json
{
  "chain_id": "truerepublic-1",
  "app_state": {
    "truedemocracy": {
      "domains": [
        {
          "name": "TestParty",
          "treasury": "500000pnyx"
        }
      ],
      "validators": [
        {
          "stake": "100000pnyx"
        }
      ]
    },
    "dex": {
      "pools": []
    }
  }
}
```

## Modifying Genesis

To customize genesis for a new network:

```bash
# Generate default genesis
truerepublicd init my-node --chain-id my-chain

# Edit genesis
vi ~/.truerepublic/config/genesis.json

# Validate genesis
truerepublicd validate-genesis
```

## Next Steps

- [Node Configuration](node-config.md)
- [Validator Guide](../../validators/README.md)
