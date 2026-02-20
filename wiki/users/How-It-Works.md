# How It Works

Deep dive into TrueRepublic's mechanics.

## Table of Contents

1. [Systemic Consensing Mechanics](#systemic-consensing-mechanics)
2. [Stones Voting System](#stones-voting-system)
3. [Proof-of-Domain Consensus](#proof-of-domain-consensus)
4. [DEX AMM Mechanics](#dex-amm-mechanics)
5. [Anonymous Voting](#anonymous-voting)
6. [VoteToEarn Rewards](#votetoearn-rewards)
7. [Proposal Lifecycle](#proposal-lifecycle)

---

## Systemic Consensing Mechanics

### The Rating Scale

```
-5: Maximum Resistance
-4: Strong Resistance
-3: Moderate Resistance
-2: Slight Resistance
-1: Minor Concerns
 0: Neutral
+1: Slight Support
+2: Moderate Support
+3: Strong Support
+4: Very Strong Support
+5: Maximum Support
```

### How Averages Work

**Example Domain:** 5 members vote on "Add Dark Mode"

```
Alice:   +5 (loves it)
Bob:     +3 (supports)
Charlie:  0 (neutral)
Diana:   -2 (concerns)
Eve:     +4 (very positive)

Average: (+5 +3 +0 -2 +4) / 5 = +2.0
```

**Result:** Moderate support, likely to pass

### Comparison: Binary vs Systemic

**Binary Voting:**
```
Proposal: Increase fees 10%
Yes: 51% → Passes
No: 49% → Ignored

Result: 49% of community unhappy
```

**Systemic Consensing:**
```
Proposal A: Increase 10%
Ratings: +1, +2, -3, -4, -2, +1, -3
Average: -1.14 (resistance)

Proposal B: Increase 5%
Ratings: +2, +3, -1, 0, +1, +2, 0
Average: +1.0 (support)

Result: Choose B (less resistance)
```

### Why It Works

1. **Captures Intensity:** Not just yes/no, but how strongly
2. **Finds Consensus:** Lowest resistance = best choice
3. **Minority Voice:** Strong opposition visible in average
4. **Better Outcomes:** Research shows better decisions

---

## Stones Voting System

### The One Stone Rule

Each member has exactly 1 stone at any time.

```
Member A places stone on Issue 1
    ↓
Issue 1: stones = 1

Member A places stone on Issue 2
    ↓
Issue 1: stones = 0 (auto-removed)
Issue 2: stones = 1 (new placement)
```

### What Stones Do

**1. Ranking:**

Items with more stones appear higher in lists

```
Issues List:
1. Security Bug       (15 stones)
2. UI Improvement     (8 stones)
3. Documentation      (3 stones)
4. Logo Change        (1 stone)
```

**2. Admin Election:**

Member with most stones in domain = Admin

```
Tech Domain:
Alice:   12 stones → Admin
Bob:      8 stones
Charlie:  5 stones
```

**3. Attention Signal:**

"I think this is important right now"

### Strategic Use

**Good:**
- Stone critical issues
- Recognize expert members
- Highlight time-sensitive proposals

**Bad:**
- Stone friends (not merit-based)
- Stone own proposals (bias)
- Forget to move stone (wasted influence)

---

## Proof-of-Domain Consensus

### The Problem: Whale Dominance

**Traditional PoS:**
```
Whale buys 51% of tokens
    ↓
Controls 51% of validators
    ↓
Controls network
```

### The Solution: Domain Requirement

**TrueRepublic:**
```
Validator must be domain member
    +
Stake provenance tracked
    +
Transfer limit: 10% of domain payouts
    =
Can't accumulate unlimited validator power
```

### How It Works

**Example:**
```
Tech Domain pays out 100,000 PNYX/month
    ↓
Validator can receive max 10,000 PNYX/month from domain
    ↓
Validator needs other stake sources
    ↓
Prevents single-domain dominance
```

### Stake Provenance Tracking

```
Validator registers with 500,000 PNYX stake

System checks:
1. Where did stake come from?
   - 250,000 from validator's wallet    ✅
   - 150,000 from external exchange     ✅
   - 100,000 from Tech Domain           ✅ (within 10% limit)

2. Is provenance valid?
   - Tech Domain pays 1M/month
   - 10% = 100,000                      ✅
   - Stake approved                     ✅
```

### Anti-Whale Benefits

1. **Decentralization:** No single entity can dominate
2. **Community Alignment:** Validators care about domains
3. **Fair Distribution:** Power spread across domains
4. **Resistance to Attacks:** Harder to acquire 51%

---

## DEX AMM Mechanics

### Constant Product Formula

```
x * y = k

Where:
x = Reserve of Token A
y = Reserve of Token B
k = Constant
```

### Example: PNYX/ATOM Pool

**Initial:**
```
PNYX: 1,000,000 (x)
ATOM:    50,000 (y)
k = 1,000,000 * 50,000 = 50,000,000,000
```

**User swaps 1,000 PNYX for ATOM:**
```
New PNYX reserve: 1,001,000
k must stay constant: 50,000,000,000
New ATOM reserve: 50,000,000,000 / 1,001,000 = 49,950

ATOM out: 50,000 - 49,950 = 50 ATOM
```

### Fee Calculation

**0.3% Swap Fee (SwapFeeBps=30):**
```
Swap 1,000 PNYX
Fee: 1,000 * 0.003 = 3 PNYX (to liquidity providers)
Actual swap: 997 PNYX
```

**1% PNYX Burn (BurnBps=100):**
```
If output is PNYX:
1,000 output calculated
1% burn: 10 PNYX → permanently burned
User receives: 990 PNYX
```

**Actual Swap Formula:**
```
output = (outReserve * input * 9970) / (inReserve * 10000 + input * 9970)
```

### Impermanent Loss

**Initial:**
```
Provide: 1,000 PNYX + 50 ATOM
Value: $1,000
```

**After Price Change:**
```
PNYX doubles in price
Pool rebalances
Your share: 707 PNYX + 35 ATOM
Value: $1,061

vs Holding:
1,000 PNYX + 50 ATOM
Value: $1,100

Impermanent Loss: $39
```

**But you earned fees:** Potentially offsets loss

---

## Anonymous Voting

### The Privacy Problem

**Without anonymity:**
```
Alice votes in Tech Domain
    ↓
Alice votes in Climate Domain
    ↓
Everyone can link Alice's votes
    ↓
Privacy compromised
```

### Dual-Key Solution

**Setup:**
```
Master Key: cosmos1alice...
    ↓
Tech Domain Key: cosmos1techkey...
    ↓
Climate Domain Key: cosmos1climatekey...

Keys are cryptographically unlinkable
```

**Voting:**
```
Alice votes in Tech Domain
Uses: cosmos1techkey...

Alice votes in Climate Domain
Uses: cosmos1climatekey...

Result: Votes can't be linked
```

### Key Lifecycle

```
1. Join domain with master key
2. Generate domain-specific Ed25519 key pair
3. Register public key via MsgJoinPermissionRegister
4. Use domain key for all ratings in that domain
5. Admin can purge register (MsgPurgePermissionRegister) to reset
```

### Privacy Guarantees

- **Unlinkable:** Domain keys cannot be mathematically linked to master key
- **Domain-isolated:** Votes in Domain A invisible from Domain B
- **Admin-protected:** Only admin can purge permission register
- **Future-proof:** ZKP can be added later for stronger guarantees

---

## VoteToEarn Rewards

### How Rewards Are Earned

**Activities that qualify:**
- Placing a stone on an issue or suggestion
- Active participation in governance

**Calculation:**
```
Reward = DomainTreasury / CEarn

Where:
CEarn = 1000 (constant)
DomainTreasury = current domain treasury balance
```

### Example

```
Tech Domain:
- Treasury: 1,000,000 PNYX
- CEarn: 1000

Alice places a stone:
Reward = 1,000,000 / 1000 = 1,000 PNYX

Tech Domain treasury after: 999,000 PNYX
Alice balance increase: 1,000 PNYX
```

### Important Rules

1. **One stone per member:** Moving stone triggers reward once
2. **Treasury-limited:** Rewards decrease as treasury shrinks
3. **Incentive alignment:** Encourages active participation
4. **Self-balancing:** Popular domains have larger treasuries = bigger rewards

### Distribution Mechanism

```
PlaceStoneOnIssue/Suggestion handler:
1. Remove previous stone (if any)
2. Place new stone on target
3. Re-sort list by stone count
4. Calculate reward: treasury / CEarn
5. Transfer reward from domain treasury to voter
```

---

## Proposal Lifecycle

### Zones

```
GREEN (0-7 days)
    ↓ Time passes
YELLOW (7-30 days)
    ↓ Time passes
RED (30+ days)
    ↓ If 2/3 vote to delete
DELETED
```

### Zone Details

**GREEN Zone (0-7 days):**
- New proposal, high visibility
- Encourage ratings and stones
- VoteToEarn rewards active
- Cannot be deleted yet

**YELLOW Zone (7-30 days):**
- Mature proposal
- Decision time: implement if positive rating
- Can be voted for deletion
- Lower visibility in feeds

**RED Zone (30+ days):**
- Expiring proposal
- Auto-delete if 2/3 majority votes to delete
- Fast-delete: any suggestion in red with 2/3 delete votes removed immediately
- Prevents clutter in domain

### Lifecycle Management

**EndBlock Hook (every block):**
```
For each suggestion in each domain:
    age = current_time - created_at

    if age < 7 days:
        status = GREEN
    else if age < 30 days:
        status = YELLOW
    else:
        status = RED

        // Check for auto-deletion
        delete_votes = count votes to delete
        total_members = domain member count
        if delete_votes >= (total_members * 2/3):
            delete suggestion
```

### Approval Threshold

For a suggestion to be considered "approved":

```
Approval requires:
- Rating average > 0 (net positive)
- Approval threshold: 5% (500 basis points) of members must have rated
- Not in RED zone or voted for deletion
```

### Why Zones Matter

| Zone | Purpose | Action |
|------|---------|--------|
| GREEN | Gather initial feedback | Rate, discuss, stone |
| YELLOW | Make decisions | Implement or archive |
| RED | Clean up | Delete or force decision |

---

## Next Steps

- [User Manuals](User-Manuals) -- Practical guides
- [Frontend Guide](Frontend-Guide) -- Using the UI
- [FAQ](FAQ) -- Common questions
