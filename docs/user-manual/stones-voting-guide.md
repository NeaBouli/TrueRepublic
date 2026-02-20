# Stones Voting Guide

Stones are TrueRepublic's mechanism for highlighting importance, signaling support, and electing domain administrators.

## What are Stones?

A **stone** is a vote you place on an issue, suggestion, or member to signal that it matters to you. Unlike systemic consensing ratings (which measure resistance on a scale), stones are a simple signal: "I think this is important."

## How Stones Work

### Key Rules

1. **One stone at a time** -- You can only have one active stone per category
2. **Moving your stone** -- Placing a new stone automatically removes your previous one
3. **VoteToEarn** -- Placing stones earns you PNYX rewards from the domain treasury
4. **Sorting** -- Items with more stones appear higher in lists

### Where You Can Place Stones

| Target | CLI Command | Purpose |
|--------|-------------|---------|
| **Issue** | `place-stone-issue` | Highlights important problems |
| **Suggestion** | `place-stone-suggestion` | Supports good solutions |
| **Member** | `place-stone-member` | Admin election (WP S3.6) |

## Placing Stones via Web Wallet

### On an Issue

1. Navigate to Governance (home page)
2. Select a domain from the left sidebar
3. Expand an issue card in the center panel
4. The stone count is displayed next to the issue name
5. Use the voting controls to place your stone

### On a Suggestion

1. Expand an issue card
2. Each suggestion shows its stone count and rating bar
3. Select the suggestion from the dropdown
4. Enter stone count and click **"Vote"**

### On a Member (Admin Election)

Admin election happens through the CLI:

```bash
truerepublicd tx truedemocracy place-stone-member \
    [domain-name] [target-member-address] \
    --from mykey --chain-id truerepublic-1
```

The member with the most stones becomes the domain admin.

## VoteToEarn Rewards (WP S3.1)

Every time you place a stone, you earn a PNYX reward from the domain treasury:

```
reward = domain_treasury / CEarn
       = domain_treasury / 1000
```

**Example:** If a domain treasury holds 500,000 PNYX:
```
reward = 500,000 / 1,000 = 500 PNYX per stone placement
```

### Important Notes

- Rewards come from the **domain treasury**, not from token inflation
- As the treasury shrinks, rewards decrease proportionally
- This creates a self-regulating economy
- Active domains with larger treasuries offer better rewards

## List Sorting

Stones affect how items are ordered in the UI:

1. Items are sorted by **stone count** (descending)
2. Items with equal stones are sorted by **creation date** (newest first)
3. This ensures the most important issues and best suggestions rise to the top

### Example

| Issue | Stones | Position |
|-------|--------|----------|
| Infrastructure Funding | 12 | 1st |
| Education Reform | 8 | 2nd |
| Healthcare Access | 8 | 3rd (older) |
| Park Renovation | 3 | 4th |

## Stones for Admin Election (WP S3.6)

Domain admins are elected through stone voting on members:

1. Any member can place a stone on another member
2. The member with the **most stones** becomes admin
3. Admin election is continuous -- it updates whenever stones change
4. This creates a fluid, ongoing representation system

### Admin Powers

The elected admin can:
- Purge the permission register (reset anonymous voting keys)
- Configure domain options (if domain allows it)
- All regular member actions

### Example Election

| Member | Stones Received | Status |
|--------|----------------|--------|
| Alice | 5 | **Admin** (most stones) |
| Bob | 3 | Member |
| Charlie | 1 | Member |
| Diana | 0 | Member |

If two members place stones on Bob, bringing him to 5:
- Tie-breaking uses the member who reached the count first
- Or the domain may require a clear majority

## Strategy Tips

### When to Place Stones

- **On Issues:** When you believe a problem is being overlooked or is urgent
- **On Suggestions:** When you find a particularly good solution
- **On Members:** When you trust someone to represent the domain well

### Stone Economy

Since you only have one stone per category:
- **Choose carefully** -- your stone is your voice
- **Move strategically** -- if a new issue arises that matters more, move your stone
- **Stay engaged** -- check if your stones still reflect your priorities

### Common Patterns

| Pattern | Description |
|---------|-------------|
| **Bandwagon** | Stone the already-popular item | Reinforces consensus |
| **Underdog** | Stone overlooked but important items | Surfaces hidden value |
| **Tactical** | Move stone to break ties | Decisive action |
| **Loyal** | Keep stone on same item long-term | Consistent support |

## CLI Reference

```bash
# Place stone on issue
truerepublicd tx truedemocracy place-stone-issue \
    [domain] [issue-name] \
    --from mykey --chain-id truerepublic-1

# Place stone on suggestion
truerepublicd tx truedemocracy place-stone-suggestion \
    [domain] [issue-name] [suggestion-name] \
    --from mykey --chain-id truerepublic-1

# Place stone on member (admin election)
truerepublicd tx truedemocracy place-stone-member \
    [domain] [target-member-address] \
    --from mykey --chain-id truerepublic-1
```

## Next Steps

- [Systemic Consensing Explained](systemic-consensing-explained.md) -- The rating system
- [Governance Tutorial](governance-tutorial.md) -- Full governance guide
- [DEX Trading Guide](dex-trading-guide.md) -- Token trading
