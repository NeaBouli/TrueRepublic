# Validator Guide

## Requirements

- **Domain membership:** Must be a member of at least one domain
- **Minimum stake:** 100,000 PNYX (`StakeMin` constant)
- **Ed25519 key pair:** 32-byte public key (hex-encoded)
- **Hardware:** 2+ CPU, 4GB+ RAM, 100GB+ SSD, stable internet

## Becoming a Validator

### Step 1: Create or Join a Domain

```bash
# Create a new domain with initial treasury
truerepublicd tx truedemocracy create-domain my-domain 200000pnyx \
    --from mykey --chain-id truerepublic-1

# Or join an existing domain (domain admin must add you)
```

### Step 2: Register as Validator

```bash
truerepublicd tx truedemocracy register-validator \
    <pubkey-hex> \
    <stake-amount>pnyx \
    <domain-name> \
    --from mykey --chain-id truerepublic-1
```

**Voting power** is calculated as: `stake / StakeMin` (integer division).

### Step 3: Monitor Your Validator

```bash
# Query your validator status
truerepublicd query truedemocracy validator <your-operator-address>

# Check if jailed
truerepublicd query truedemocracy validators
```

## Proof of Domain (PoD)

TrueRepublic uses Proof of Domain instead of traditional Proof of Stake:

- Validators **must maintain domain membership** to remain active
- `EndBlock` enforces membership every block
- Validators removed from all domains are **automatically evicted**
- Domain participation is verified on-chain

### Transfer Limit (WP S7)

Stake withdrawals are capped at **10% of the domain's cumulative total payouts**:

```
max_withdrawal = domain_total_payouts * 0.10
```

This prevents validators from extracting more value than the domain generates.

## Staking Rewards

### Node Rewards (eq.5)

- **APY:** 10% (`ApyNode = 0.10`)
- **Distribution:** Every `RewardInterval` (3600 seconds / 1 hour)
- **Release decay:** Rewards decrease as total supply approaches 22M PNYX

```
reward = stake * ApyNode * timeInYears * (1 - totalReleased / 22000000)
```

### Domain Interest (eq.4)

Domain treasuries earn interest at **25% APY** (`ApyDom = 0.25`), also subject to release decay.

## Slashing

| Infraction | Slash | Jail Duration |
|------------|-------|---------------|
| Double-sign | 5% of stake | 100 minutes |
| Downtime (>50 missed in 100 blocks) | 1% of stake | 10 minutes |

### After Slashing

- If stake drops below `StakeMin` (100,000 PNYX), voting power is set to 0
- Jailed validators **do not earn rewards**
- Unjail after jail period expires (if still above minimum stake):

```bash
truerepublicd tx truedemocracy unjail \
    --from mykey --chain-id truerepublic-1
```

## Operational Best Practices

### Monitoring

- Enable Prometheus in `config.toml`: `prometheus = true`
- CometBFT exposes metrics on port **26660**
- Key metrics to watch:
  - `cometbft_consensus_height` — block height progression
  - `cometbft_consensus_missing_validators` — should be 0
  - `cometbft_p2p_peers` — peer connectivity
  - `cometbft_consensus_rounds` — should mostly be 0 (single round)

### Backups

```bash
# Daily automated backup
0 3 * * * /path/to/scripts/backup.sh

# Manual backup before upgrades
tar -czf pre-upgrade-backup.tar.gz ~/.truerepublic
```

### Security

- Run behind a firewall (UFW recommended)
- Only expose P2P (26656) and optionally RPC (26657) publicly
- Use a sentry node architecture for DDoS protection
- Keep the validator key (`priv_validator_key.json`) secure and backed up offline

### Firewall Rules

```bash
sudo ufw allow 26656/tcp   # P2P (required)
sudo ufw allow 26657/tcp   # RPC (optional, for queries)
sudo ufw deny 1317/tcp     # LCD (internal only)
sudo ufw deny 9090/tcp     # gRPC (internal only)
sudo ufw enable
```
