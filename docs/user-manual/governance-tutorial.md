# Governance Tutorial

Learn how to participate in TrueRepublic's direct democracy governance system.

## What is a Domain?

A **Domain** is a community focused on a specific topic. Think of it as a governance space where members collaborate on related issues and propose solutions.

Examples:
- **Climate Domain** -- Environmental initiatives and policy
- **Education Domain** -- Educational reform proposals
- **Tech Domain** -- Technology policies and development priorities

### Domain Structure

Each domain contains:

| Component | Description |
|-----------|-------------|
| **Members** | People who have joined the domain |
| **Admin** | Elected by member stones (highest-stoned member) |
| **Treasury** | Shared PNYX funds for initiatives |
| **Issues** | Problems identified by members |
| **Suggestions** | Proposed solutions to issues |
| **Permission Register** | Anonymous voting keys (WP S4) |

### Domain Options

Domains can be configured with:
- **Admin Electable** -- Whether admin is elected by stones
- **Anyone Can Join** -- Open vs. invite-only membership
- **Only Admin Issues** -- Whether only admin can create issues
- **Coin Burn Required** -- Whether proposals require a PNYX fee
- **Approval Threshold** -- Rating threshold for suggestion approval (default 5%)
- **Dwell Time** -- How long suggestions stay in each zone (default 1 day)

## Browsing Domains

1. Open the web wallet (Governance is the home page)
2. The **left sidebar** shows all available domains
3. Click a domain to select it
4. The **center panel** shows issues and suggestions
5. The **right panel** shows domain statistics and actions

## Understanding Proposals

Each proposal has two parts:

**Issue** -- The problem being addressed:
```
"Infrastructure funding is insufficient"
```

**Suggestion** -- A proposed solution:
```
"Allocate 20% of treasury to infrastructure projects"
```

### Suggestion Lifecycle (WP S3.1.2)

Suggestions move through three color-coded zones:

| Zone | Color | Duration | Meaning |
|------|-------|----------|---------|
| **New** | Green | Default: 1 day | Fresh suggestion, gathering ratings |
| **Mature** | Yellow | Default: 1 day | Enough time for evaluation |
| **Expiring** | Red | Default: 1 day | Will be auto-deleted if not supported |

- Suggestions in the **red zone** are automatically deleted when time expires
- Any suggestion can be **fast-deleted** by a 2/3 majority vote at any time
- Placing a stone on a suggestion resets its lifecycle

## Submitting a Proposal

1. Connect your wallet
2. Select a domain (you must be a member)
3. In the **right panel**, find "Submit Proposal"
4. Enter the **Issue Name** (the problem)
5. Enter the **Suggestion** (your proposed solution)
6. Click **"Submit Proposal"**
7. Approve the transaction in Keplr

### PayToPut Fee

Submitting a proposal costs PNYX. The fee is calculated as:

```
fee = treasury / 1000 * min(15, number_of_members)
```

This prevents spam while keeping costs proportional to domain size.

## Rating Suggestions (Systemic Consensing)

TrueRepublic uses **Systemic Consensing** instead of Yes/No voting. You rate each suggestion on a scale from **-5 to +5**.

See [Systemic Consensing Explained](systemic-consensing-explained.md) for the full guide.

### Quick Rating Guide

1. Expand an issue card (click to toggle)
2. Select a suggestion from the dropdown
3. Enter your stone count
4. Click **"Vote"**

### What the Ratings Mean

| Rating | Meaning |
|--------|---------|
| **-5** | Strong resistance -- "This would cause serious harm" |
| **0** | Neutral -- "No strong feelings either way" |
| **+5** | Strong support -- "This is essential" |

## Stones Voting

Stones are a separate mechanism for highlighting importance. See [Stones Voting Guide](stones-voting-guide.md) for details.

You can place stones on:
- **Issues** -- "This problem matters"
- **Suggestions** -- "This solution is good"
- **Members** -- For admin election

## Voting to Delete

Vote to delete a suggestion when it is spam, duplicated, already solved, or violates domain rules.

- Requires **2/3 majority** of domain members
- Triggers immediate removal (fast-delete)

## VoteToEarn Rewards

Active participation earns PNYX rewards:

```
reward = domain_treasury / 1000
```

You earn rewards by:
- Placing stones on issues
- Placing stones on suggestions
- Regular participation

Rewards come from the domain treasury and are distributed to active voters.

## Admin Election (WP S3.6)

The domain admin is the member with the **most stones** from other members.

- Any member can place a stone on another member
- The member with the highest stone count becomes admin
- Admin can:
  - Purge the permission register
  - Configure domain options (if enabled)
  - All regular member actions

## Member Exclusion

A member can be excluded by **2/3 majority vote**:

1. A member initiates a vote-exclude transaction
2. Other members cast their votes
3. If 2/3 of members vote to exclude, the member is removed
4. Excluded members lose access to the domain

## Inactivity Cleanup

Members inactive for **360 days** are automatically removed during EndBlock processing. Stay active by voting, placing stones, or submitting proposals.

## Next Steps

- [Systemic Consensing Explained](systemic-consensing-explained.md) -- Deep dive into the rating system
- [Stones Voting Guide](stones-voting-guide.md) -- Using stones effectively
- [DEX Trading Guide](dex-trading-guide.md) -- Trading on the DEX
