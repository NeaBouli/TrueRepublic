# Validator Guide

Complete guide to running a TrueRepublic validator with Proof of Domain (PoD) consensus.

## What is a Validator?

Validators are nodes that participate in consensus -- proposing and signing blocks to secure the TrueRepublic blockchain. In return, validators earn staking rewards.

TrueRepublic uses **Proof of Domain (PoD)** instead of traditional Proof of Stake. This means validators must be **active members of at least one domain** to participate in consensus.

## Requirements

| Requirement | Details |
|-------------|---------|
| **Minimum Stake** | 100,000 PNYX |
| **Domain Membership** | At least one domain |
| **Ed25519 Key** | 32-byte public key (hex-encoded) |
| **Hardware** | 2+ CPU, 4GB+ RAM, 100GB+ SSD |
| **Uptime** | 99.9%+ recommended |
| **Network** | Stable, low-latency connection |

## Becoming a Validator

### Step 1: Set Up a Full Node

Follow the [Node Operators Guide](../node-operators/README.md) to set up and sync a full node first.

### Step 2: Create or Join a Domain

You must be a member of at least one domain:

```bash
# Option A: Create a new domain with initial treasury
truerepublicd tx truedemocracy create-domain my-domain 200000pnyx \
    --from mykey --chain-id truerepublic-1

# Option B: Join an existing domain
# (Domain admin or open-join domain required)
```

### Step 3: Register as Validator

```bash
truerepublicd tx truedemocracy register-validator \
    <pubkey-hex> \
    <stake-amount>pnyx \
    <domain-name> \
    --from mykey --chain-id truerepublic-1
```

**Parameters:**
- `pubkey-hex` -- Your Ed25519 public key in hex encoding (32 bytes)
- `stake-amount` -- Amount to stake (minimum 100,000 PNYX)
- `domain-name` -- Domain you're a member of

### Step 4: Verify Registration

```bash
truerepublicd query truedemocracy validator <your-operator-address>
```

## Proof of Domain (PoD)

### How PoD Works

Unlike traditional PoS where anyone with tokens can validate:

1. **Domain Membership Required** -- Validators must be active members of at least one domain
2. **Continuous Enforcement** -- `EndBlock` checks domain membership every block
3. **Automatic Eviction** -- Validators removed from all domains are evicted from the validator set
4. **Community Accountability** -- Domain members can vote to exclude bad actors

### Why PoD?

- Ensures validators are **invested in governance**, not just financially
- Creates **accountability** through community oversight
- Prevents **plutocracy** -- wealth alone doesn't guarantee validation rights
- Aligns **validator incentives** with the democratic mission

### Transfer Limit (WP S7)

Stake withdrawals are capped to prevent value extraction:

```
max_withdrawal = domain_total_payouts * 10%
```

This means validators cannot withdraw more than 10% of what the domain has paid out in total. This ensures validators don't extract more value than they help create.

## Voting Power

Voting power determines how much influence a validator has in consensus:

```
voting_power = stake / StakeMin = stake / 100,000
```

| Stake | Voting Power |
|-------|-------------|
| 100,000 PNYX | 1 |
| 200,000 PNYX | 2 |
| 500,000 PNYX | 5 |
| 1,000,000 PNYX | 10 |

## Staking Rewards

### Node Rewards (eq.5)

Validators earn staking rewards at **10% APY**, subject to release decay:

```
reward = stake * 0.10 * (time_in_years) * (1 - total_released / 22,000,000)
```

- Distributed every **3,600 seconds** (1 hour)
- Rewards decrease as total supply approaches 22M PNYX
- Jailed validators do **not** earn rewards

### Domain Interest (eq.4)

Domain treasuries earn interest at **25% APY** (also subject to release decay). This indirectly benefits validators who are domain members.

## Slashing

Validators are penalized for misbehavior:

| Infraction | Slash Amount | Jail Duration | Recovery |
|------------|-------------|---------------|----------|
| **Double-signing** | 5% of stake | 100 minutes | Unjail after period |
| **Downtime** | 1% of stake | 10 minutes | Unjail after period |

### Downtime Detection

- Tracked over a **100-block window**
- Must sign at least **50 blocks** (50%)
- Missing 51+ blocks triggers downtime slashing

### After Being Slashed

1. Your stake is reduced by the slash percentage
2. You are **jailed** (removed from active validator set)
3. You do **not earn rewards** while jailed
4. If stake drops below 100,000 PNYX, voting power becomes 0
5. After jail period expires, unjail yourself:

```bash
truerepublicd tx truedemocracy unjail \
    --from mykey --chain-id truerepublic-1
```

### Preventing Slashing

- Ensure **99.9%+ uptime**
- Use redundant infrastructure
- Monitor block signing with Prometheus metrics
- Set up alerts for `cometbft_consensus_missing_validators`
- **Never run two nodes with the same validator key** (double-sign risk)

## Validator Operations

### Checking Status

```bash
# Your validator info
truerepublicd query truedemocracy validator <your-operator-addr>

# All validators
truerepublicd query truedemocracy validators

# Node sync status
curl http://localhost:26657/status | jq .result.sync_info
```

### Withdrawing Stake

```bash
truerepublicd tx truedemocracy withdraw-stake <amount>pnyx \
    --from mykey --chain-id truerepublic-1
```

Remember: withdrawals are capped at 10% of domain total payouts.

### Removing Your Validator

```bash
truerepublicd tx truedemocracy remove-validator <your-operator-addr> \
    --from mykey --chain-id truerepublic-1
```

## Operational Checklist

- [ ] Full node synced and running
- [ ] Domain membership confirmed
- [ ] Minimum 100,000 PNYX staked
- [ ] Validator key backed up securely offline
- [ ] Monitoring and alerting configured
- [ ] Sentry node architecture for DDoS protection
- [ ] Automated backup schedule
- [ ] Firewall configured (only P2P port public)
- [ ] TLS for any public endpoints
- [ ] UPS or redundant power for the server

## Best Practices

1. **Never expose your validator node directly** -- Use sentry nodes
2. **Monitor continuously** -- Set up Prometheus + Grafana + alerts
3. **Back up your validator key offline** -- This is your identity
4. **Keep your node updated** -- Apply security patches promptly
5. **Stay active in your domain** -- PoD requires ongoing membership
6. **Don't run multiple nodes with the same key** -- This will cause double-signing
7. **Test upgrades on testnet first** -- Before applying to production

## Next Steps

- [Node Setup](../node-operators/README.md) -- Set up your node
- [Monitoring](../node-operators/operations/monitoring.md) -- Monitor your validator
- [Security](../node-operators/operations/security.md) -- Harden your setup
