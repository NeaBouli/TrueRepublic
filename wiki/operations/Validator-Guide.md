# Validator Guide

Complete guide to becoming and operating a TrueRepublic validator.

## Table of Contents

1. [Overview](#overview)
2. [Requirements](#requirements)
3. [Proof-of-Domain](#proof-of-domain)
4. [Registration](#registration)
5. [Operations](#operations)
6. [Rewards](#rewards)
7. [Slashing](#slashing)
8. [Unjailing](#unjailing)
9. [Best Practices](#best-practices)

---

## Overview

### What is a Validator?

Validators:
- Secure the network
- Produce blocks
- Participate in consensus
- Earn staking rewards

### Responsibilities

**Must:**
- Run 24/7 with high uptime (>95%)
- Sign blocks correctly
- Stay updated with upgrades
- Monitor performance

**Must NOT:**
- Double-sign blocks (5% slash)
- Go offline too long (1% slash)
- Run multiple instances with same key

---

## Requirements

### Technical Requirements

**Hardware:**
- CPU: 8 cores
- RAM: 16 GB
- Storage: 1 TB NVMe SSD
- Network: 1 Gbps up/down
- 24/7 uptime

**Software:**
- Running synced full node
- Monitoring setup
- Backup systems

### Token Requirements

**Minimum Stake:**
- 100,000 PNYX

**Recommended Stake:**
- 200,000+ PNYX (more voting power)

### Proof-of-Domain Requirement

**Must be member of a domain:**
- Join domain first
- Validator tied to domain
- Stake provenance validated

---

## Proof-of-Domain

### What is Proof-of-Domain?

Anti-whale mechanism that prevents single entity from controlling too many validators.

### Rules

**1. Domain Membership Required**

```bash
# Must be member before registering
truerepublicd query truedemocracy domain my-domain

# Should show your address in members list
```

**2. Stake Provenance Tracked**

System tracks where your PNYX came from:
- Self-generated: No limit
- Bought from exchange: No limit
- Received from domain: Max 10% of domain payouts

**3. Transfer Limit: 10%**

```
Domain pays out 1,000,000 PNYX/month
    ↓
Validator can receive max 100,000 PNYX/month from domain
    ↓
Limits validator concentration per domain
```

### Why PoD?

**Without PoD:**
```
Whale buys 10M PNYX
    ↓
Creates 50 validators
    ↓
Controls network
```

**With PoD:**
```
Whale buys 10M PNYX
    ↓
Can only create validators across multiple domains
    ↓
Must be member of each domain
    ↓
Transfer limits prevent concentration
    ↓
Network stays decentralized
```

### Checking Provenance

```bash
# Check domain payouts
truerepublicd query truedemocracy domain YOUR_DOMAIN

# System automatically validates provenance on registration
```

`truerepublicd` registers no `bank` command (only `truedemocracy` and `dex`
transactions and queries), so check account balances with the maintained client
or another tool you have independently qualified.

---

## Registration

### Prerequisites Check

```bash
# 1. Node fully synced (native node host; in the Docker setup use
#    curl -s http://127.0.0.1:8080/rpc/status)
curl -s http://127.0.0.1:26657/status | jq .result.sync_info.catching_up
# Should return: false

# 2. Sufficient balance: at least StakeMin = 100,000,000,000 upnyx
#    (100,000 PNYX); check it with a qualified client (no bank CLI command)

# 3. Domain membership
truerepublicd query truedemocracy domain YOUR_DOMAIN
# Should show you in members list
```

### Step 1: Create Validator Key

```bash
# Generate new key
truerepublicd keys add validator --keyring-backend file

# Output:
# - name: validator
# - address: cosmos1abc...
# - pubkey: cosmospub1...
# - mnemonic: word1 word2 ... word24

# CRITICAL: Backup mnemonic securely!
```

**Key Management:**
- Store mnemonic offline (paper + safe)
- Never share with anyone
- Keep multiple backups
- Consider hardware wallet (Ledger)

### Step 2: Fund Validator Address

The operator address must already hold the stake. `truerepublicd` registers no
`bank send` command; fund the address through the qualified genesis allocation
or with a client you have independently qualified. `100000000000upnyx` =
100,000 PNYX (`upnyx` is the only base denomination, 6 decimals).

### Step 3: Register Validator

`register-validator` takes exactly three arguments: the hex-encoded 32-byte
Ed25519 consensus public key of this node, the stake, and the domain:

```bash
# Consensus public key of this node as hex
PUBKEY_HEX=$(jq -r .pub_key.value ~/.truerepublic/config/priv_validator_key.json | base64 -d | xxd -p -c 64)

truerepublicd tx truedemocracy register-validator \
    "$PUBKEY_HEX" \
    100000000000upnyx \
    YOUR_DOMAIN \
    --from validator \
    --chain-id <chain-id> \
    --keyring-backend file
```

This command requires domain membership, validates stake provenance, escrows
the stake in the `truedemocracy` module and creates the validator record with
power `stake / StakeMin`.

### Step 4: Verify Registration

```bash
# Check validator status
truerepublicd query truedemocracy validator VALIDATOR_ADDRESS

# Fields: operator_addr, pub_key, stake, domains, power, jailed,
# jailed_until, missed_blocks
```

---

## Operations

### Monitoring Signing Status

```bash
# Signing health is part of the validator record (no slashing module/CLI)
truerepublicd query truedemocracy validator VALIDATOR_ADDRESS
```

**What to watch:**
- `missed_blocks` should stay low (jail after more than 50 missed commits in
  a complete 100-block window)
- `jailed` should be `false`

### Checking Validator Set

```bash
# View all validators
truerepublicd query truedemocracy validators

# Your validator info
truerepublicd query truedemocracy validator VALIDATOR_ADDRESS

# Consensus validator set (native node host)
curl -s http://127.0.0.1:26657/validators
```

### Updating Validator Info

**Commission:** not available. TrueRepublic uses Proof of Domain without
`x/staking`; there is no commission or `edit-validator` operation.

**Change domain:**

```bash
# Must unjail, withdraw, re-register with new domain
```

---

## Rewards

### Reward Sources

**1. Block Rewards**
- Fixed per block
- Distributed to active validators
- Proportional to voting power

**2. Transaction Fees**
- Collected from all transactions
- Distributed to block proposer

**3. Domain Alignment**
- Bonus for validating domain's interests
- (Future feature)

### Reward Calculation

```
R_validator = (YourStake / TotalStake) * BlockReward * BlocksProduced
```

**Example:**
```
Your stake:  100,000 PNYX
Total stake: 10,000,000 PNYX
Your share:  1%

Block reward:  10 PNYX
Blocks per day: 17,280 (5s block time)
Daily block rewards: 172,800 PNYX

Your daily reward: 172,800 * 0.01 = 1,728 PNYX
Annual: 630,720 PNYX

APY: 630,720 / 100,000 = ~6.3%
```

**Factors affecting rewards:**
- Total staked PNYX (more stake = lower APY)
- Your uptime (downtime = missed rewards)
- Slashing (reduces stake = less rewards)

### Claiming Rewards

**Automatic:**
- Rewards added to bonded pool
- Compound automatically

**Manual withdrawal:** not available. There is no `x/distribution` module and
no commission withdrawal command.

---

## Slashing

### Slashing Conditions

**1. Double-Signing (5% slash)**

**What:** Signing two different blocks at same height

**How it happens:**
- Running two validator instances with same key
- Key compromise
- Software bug

**Penalty:**
- 5% of stake slashed
- Permanent tombstoning
- Cannot unjail

**Prevention:**
- Never run duplicate instances
- Secure private keys
- Use HSM or signing service

**2. Downtime (1% slash)**

**What:** Missing too many blocks

**Threshold:**
- Miss >5% of blocks in signed_blocks_window
- Default: 5% of 10,000 blocks = 500 blocks
- ~42 minutes at 5s block time

**Penalty:**
- 1% of stake slashed
- Temporary jail
- Can unjail after period

**Prevention:**
- 99.5%+ uptime
- Redundant infrastructure
- Monitoring + alerts
- Fast response to issues

### Slashing Events

**When slashed:**
```
1. Stake reduced by penalty %
2. Validator jailed
3. Removed from active set
4. Stop earning rewards
5. Event emitted on-chain
```

**Check if slashed:**

```bash
truerepublicd query truedemocracy validator VALIDATOR_ADDRESS

# Look for:
# - jailed: true
# - jailed_until: <future unix timestamp>
```

---

## Unjailing

### When Can You Unjail?

**Downtime Slash:**
- Can unjail after jail period
- Must wait ~10 minutes
- Can rejoin validator set

**Double-Sign Slash:**
- The offending consensus key is tombstoned and cannot be unjailed
- Rotate to a freshly generated consensus key
  ([Validator Consensus-Key Rotation](../../docs/node-operators/operations/validator-key-rotation.md)),
  wait for the jail period, then unjail

### Unjail Process

**Step 1: Wait for Jail Period**

```bash
# Check when you can unjail
truerepublicd query truedemocracy validator VALIDATOR_ADDRESS

# Look at jailed_until (unix timestamp) and wait until it has passed
```

**Step 2: Ensure Node is Running**

```bash
# Check sync status (native node host)
curl -s http://127.0.0.1:26657/status | jq .result.sync_info.catching_up
# Should be: false

# Check node is signing
# Should see new blocks in logs
```

**Step 3: Send Unjail Transaction**

```bash
truerepublicd tx truedemocracy unjail \
    --from validator \
    --chain-id truerepublic-1 \
    --gas auto \
    --gas-adjustment 1.3
```

**Step 4: Verify Unjailed**

```bash
# Check validator status
truerepublicd query truedemocracy validator VALIDATOR_ADDRESS

# Should show:
# "jailed": false
# "status": "active"
```

**Step 5: Monitor Closely**

After unjailing:
- Watch for missed blocks
- Check uptime metrics
- Ensure infrastructure stable
- Monitor for 24+ hours

---

## Best Practices

### Infrastructure

**1. Redundancy**

```
Primary Node
    +
Backup Node (standby)
    +
Sentry Nodes (optional, protect validator)
```

**2. Monitoring**
- Prometheus + Grafana
- Alerting on missed blocks
- Uptime monitoring
- Disk space alerts

**3. Security**
- Firewall (only necessary ports)
- SSH key auth (no password)
- Fail2ban (brute force protection)
- DDoS protection
- HSM for validator key (optional)

### Operational Checklist

**Daily:**
- Check missed blocks (<10)
- Check disk space (>20% free)
- Check peer count (>10)
- Check logs for errors

**Weekly:**
- Review Grafana dashboards
- Check for software updates
- Verify backups work
- Test alert system

**Monthly:**
- Security audit
- Performance review
- Capacity planning
- Key rotation (if needed)

### Upgrade Procedure

**1. Preparation:**
- Monitor announcements
- Read upgrade docs
- Test on testnet first
- Schedule maintenance window

**2. Execution:**

```bash
# Stop validator
sudo systemctl stop truerepublicd

# Backup
./backup.sh

# Upgrade binary
cd ~/TrueRepublic
git pull origin main
make build
sudo cp build/truerepublicd /usr/local/bin/

# Verify version
truerepublicd version

# Start validator
sudo systemctl start truerepublicd

# Monitor closely
sudo journalctl -u truerepublicd -f
```

**3. Post-Upgrade:**
- Check signing immediately
- Monitor for 1 hour
- Verify no missed blocks
- Check peers reconnected

### Emergency Response

**Node Down:**
```
1. Check systemd status
2. Check logs
3. Check disk space
4. Check network
5. Restart if needed
6. Escalate if >5 minutes
```

**High Missed Blocks:**
```
1. Check sync status
2. Check peer count
3. Check resource usage (CPU/RAM/disk)
4. Restart if necessary
5. Unjail if slashed
```

---

## Next Steps

- [Monitoring](Monitoring) -- Set up comprehensive monitoring
- [Node Setup](Node-Setup) -- Advanced node configuration
- [Troubleshooting](Troubleshooting) -- Common validator issues
