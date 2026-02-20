# CLI Commands Reference

Complete reference for all `truerepublicd` commands.

## Global Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--chain-id` | `truerepublic-1` | Chain identifier |
| `--from` | (required) | Signing key name |
| `--keyring-backend` | `test` | Keyring backend (os, file, test) |
| `--home` | `~/.truerepublic` | Node home directory |
| `--output` | `text` | Output format (text, json) |
| `--node` | `tcp://localhost:26657` | RPC endpoint |

## Transaction Commands

### truedemocracy (13 commands)

```bash
# Create a domain with initial treasury
truerepublicd tx truedemocracy create-domain [name] [initial-coins] \
    --from mykey --chain-id truerepublic-1

# Submit a proposal (issue + suggestion)
truerepublicd tx truedemocracy submit-proposal \
    [domain] [issue] [suggestion] [fee] [external-link] \
    --from mykey --chain-id truerepublic-1

# Register as PoD validator
truerepublicd tx truedemocracy register-validator \
    [pubkey-hex] [stake]pnyx [domain] \
    --from mykey --chain-id truerepublic-1

# Withdraw validator stake (10% transfer limit)
truerepublicd tx truedemocracy withdraw-stake [amount]pnyx \
    --from mykey --chain-id truerepublic-1

# Remove a validator
truerepublicd tx truedemocracy remove-validator [operator-addr] \
    --from mykey --chain-id truerepublic-1

# Unjail validator after jail period
truerepublicd tx truedemocracy unjail \
    --from mykey --chain-id truerepublic-1

# Register domain key for anonymous voting
truerepublicd tx truedemocracy join-permission-register \
    [domain] [domain-pubkey-hex] \
    --from mykey --chain-id truerepublic-1

# Purge permission register (admin only)
truerepublicd tx truedemocracy purge-permission-register [domain] \
    --from mykey --chain-id truerepublic-1

# Place stone on an issue
truerepublicd tx truedemocracy place-stone-issue [domain] [issue] \
    --from mykey --chain-id truerepublic-1

# Place stone on a suggestion
truerepublicd tx truedemocracy place-stone-suggestion \
    [domain] [issue] [suggestion] \
    --from mykey --chain-id truerepublic-1

# Place stone on a member (admin election)
truerepublicd tx truedemocracy place-stone-member \
    [domain] [target-member] \
    --from mykey --chain-id truerepublic-1

# Vote to exclude a member (2/3 majority required)
truerepublicd tx truedemocracy vote-exclude \
    [domain] [target-member] \
    --from mykey --chain-id truerepublic-1

# Vote to fast-delete a suggestion (2/3 majority)
truerepublicd tx truedemocracy vote-delete \
    [domain] [issue] [suggestion] \
    --from mykey --chain-id truerepublic-1
```

### dex (4 commands)

```bash
# Create a liquidity pool
truerepublicd tx dex create-pool [asset-denom] [pnyx-amt] [asset-amt] \
    --from mykey --chain-id truerepublic-1

# Swap tokens (0.3% fee, 1% PNYX burn)
truerepublicd tx dex swap [input-denom] [input-amt] [output-denom] \
    --from mykey --chain-id truerepublic-1

# Add liquidity to pool
truerepublicd tx dex add-liquidity [asset-denom] [pnyx-amt] [asset-amt] \
    --from mykey --chain-id truerepublic-1

# Remove liquidity (burn LP shares)
truerepublicd tx dex remove-liquidity [asset-denom] [shares] \
    --from mykey --chain-id truerepublic-1
```

## Query Commands

### truedemocracy (4 commands)

```bash
# Query a specific domain
truerepublicd query truedemocracy domain [name]

# List all domains
truerepublicd query truedemocracy domains

# Query a specific validator
truerepublicd query truedemocracy validator [operator-addr]

# List all validators
truerepublicd query truedemocracy validators
```

### dex (2 commands)

```bash
# Query a specific pool
truerepublicd query dex pool [asset-denom]

# List all pools
truerepublicd query dex pools
```

## Examples

### Create a Domain and Submit a Proposal

```bash
# Create domain with 200,000 PNYX treasury
truerepublicd tx truedemocracy create-domain "Climate" 200000pnyx \
    --from alice --chain-id truerepublic-1

# Submit a proposal
truerepublicd tx truedemocracy submit-proposal \
    "Climate" \
    "Carbon Emissions Reporting" \
    "Require quarterly carbon reports from all members" \
    10000pnyx \
    "https://example.com/proposal-details" \
    --from alice --chain-id truerepublic-1
```

### Become a Validator

```bash
# Create/join a domain first
truerepublicd tx truedemocracy create-domain "Validators" 100000pnyx \
    --from operator --chain-id truerepublic-1

# Register as validator
truerepublicd tx truedemocracy register-validator \
    $(cat pubkey.hex) 150000pnyx "Validators" \
    --from operator --chain-id truerepublic-1

# Check status
truerepublicd query truedemocracy validators
```

### Trade on DEX

```bash
# Create a PNYX/ATOM pool
truerepublicd tx dex create-pool atom 100000 50000 \
    --from lp-provider --chain-id truerepublic-1

# Swap 1000 PNYX for ATOM
truerepublicd tx dex swap pnyx 1000 atom \
    --from trader --chain-id truerepublic-1

# Check pool state
truerepublicd query dex pool atom
```
