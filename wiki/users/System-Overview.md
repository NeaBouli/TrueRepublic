# System Overview

Understanding TrueRepublic: A blockchain for direct democracy.

## What is TrueRepublic?

TrueRepublic is a blockchain-based direct democracy platform that enables communities to govern themselves through transparent, fair, and efficient decision-making.

### Core Concept

Traditional democracy:
```
Representatives decide → Citizens vote for representatives → Citizens have indirect influence
```

TrueRepublic:
```
Citizens propose → Citizens rate (-5 to +5) → Best solutions emerge → Citizens execute
```

### Key Innovation: Systemic Consensing

Instead of "Yes/No" voting, members rate proposals on a scale from -5 (strong resistance) to +5 (strong support).

**Why this matters:**
- Captures intensity of feelings
- Finds solutions with least resistance
- Minority concerns visible
- Better outcomes than majority rule

### Real-World Example

**Traditional Voting:**
```
Proposal: Build new park
51% vote Yes → Park built
49% unhappy
```

**Systemic Consensing:**
```
Proposal A: Build new park → Average: +1.2
Proposal B: Renovate existing → Average: +3.8
Result: Renovate (broader support)
```

## How TrueRepublic Works

### 1. Domains (Communities)

**What:** Self-governing communities focused on specific topics

**Examples:**
- Tech Domain (technology governance)
- Climate Domain (environmental initiatives)
- Education Domain (school reform)

**Each domain has:**
- Members (anyone can join)
- Treasury (shared funds)
- Proposals (issues + solutions)
- Admin (elected by stones)

### 2. Proposals (Issue + Suggestion)

**Structure:**
```
ISSUE (Problem)
"Our UI is difficult to use"

SUGGESTION (Solution)
"Implement dark mode and better navigation"
```

**Lifecycle:**
- GREEN (0-7 days): New, high visibility
- YELLOW (7-30 days): Mature, needs decision
- RED (30+ days): Expiring, may be deleted

### 3. Voting Systems

**A) Systemic Consensing (Rating)**

Rate each suggestion: -5 to +5
```
-5: Strongly oppose
-3: Have concerns
 0: Neutral
+3: Support
+5: Strongly support
```

Average determines success.

**B) Stones Voting (Priority)**

Each member has 1 stone. Place on:
- Issue (this problem matters)
- Suggestion (good solution)
- Member (recognize expertise)

Stones affect ranking in lists.

**C) Vote to Delete**

2/3 majority to remove spam/outdated proposals.

### 4. Validators (Proof-of-Domain)

**What:** Nodes that secure the network

**Requirements:**
- Be member of a domain
- Stake 100,000+ PNYX
- Run validator node

**Why Proof-of-Domain?**
- Anti-whale: Can't accumulate unlimited power
- Community-aligned: Validators care about domains
- Transfer limits: Max 10% of domain payouts

### 5. DEX (Decentralized Exchange)

**What:** Built-in token exchange

**How it works:**
- Constant Product AMM (x*y=k)
- Liquidity pools (PNYX/ATOM, PNYX/USDC, etc.)
- 0.3% fee to liquidity providers
- 1% PNYX burn (deflationary)

**Why built-in?**
- No external exchange needed
- PNYX burn reduces supply
- Liquidity providers earn passive income

### 6. Tokenomics (PNYX)

**Token:** PNYX

**Uses:**
- Governance (voting, proposals)
- Staking (become validator)
- DEX (trading, liquidity)
- Fees (transactions)

**Distribution:**
- Node rewards: 10% APY
- Domain interest: 2.5% APY
- VoteToEarn: Active participation rewards

**Deflationary:**
- 1% burn on DEX swaps
- Gradual supply reduction
- Value accrual to holders

### 7. Anonymous Voting

**Problem:** Your votes in one domain visible in another

**Solution:** Dual-key system
```
Master Key (Keplr wallet)
    ↓
Domain A Key (unlinkable)
Domain B Key (unlinkable)
Domain C Key (unlinkable)
```

**Result:** Votes in Tech Domain can't be linked to votes in Climate Domain

## Who Uses TrueRepublic?

### End Users

**Activities:**
- Join domains
- Submit proposals (PayToPut: 10,000 PNYX)
- Rate suggestions (-5 to +5)
- Place stones
- Earn VoteToEarn rewards

**Tools:**
- Web wallet (browser)
- Mobile wallet (iOS/Android)

### Node Operators

**Activities:**
- Run full nodes
- Provide infrastructure
- Earn node rewards (10% APY)

**Tools:**
- Docker / Native binary
- Monitoring (Prometheus/Grafana)

### Validators

**Activities:**
- Secure network
- Produce blocks
- Participate in consensus
- Earn staking rewards

**Requirements:**
- 100,000 PNYX stake
- Domain membership
- 24/7 uptime

### Developers

**Activities:**
- Build dApps
- Create CosmWasm contracts
- Integrate with TrueRepublic
- Contribute to core

**Tools:**
- Cosmos SDK modules
- CosmJS library
- REST/gRPC APIs

## Technology Stack

**Blockchain:**
- Cosmos SDK (modular framework)
- CometBFT (Byzantine Fault Tolerant consensus)
- 5-second block time
- Instant finality

**Smart Contracts:**
- CosmWasm (Wasm-based)
- Written in Rust
- Secure sandboxing

**Frontend:**
- Web: React + Tailwind CSS
- Mobile: React Native
- Wallet: Keplr integration

**Infrastructure:**
- Docker (deployment)
- Prometheus/Grafana (monitoring)
- Nginx (reverse proxy)

## Key Benefits

### For Communities

- **True Democracy:** Every member has equal voice
- **Transparent:** All votes on-chain
- **Fair:** Systemic Consensing finds best solutions
- **Efficient:** Fast decisions (5-second blocks)

### For Validators

- **Rewards:** Earn from staking
- **Influence:** Shape network direction
- **Anti-Whale:** Proof-of-Domain prevents centralization

### For Developers

- **Modular:** Easy to extend
- **Well-Documented:** Comprehensive guides
- **Active:** Engaged community
- **Interoperable:** IBC-ready (future)

## Comparison with Other Systems

### vs Traditional Blockchain Governance

| Feature | Other Chains | TrueRepublic |
|---------|--------------|--------------|
| Voting | Yes/No | -5 to +5 rating |
| Whale Prevention | Token-weighted | Proof-of-Domain |
| Privacy | Public | Anonymous (dual-key) |
| Engagement | Low | VoteToEarn rewards |

### vs Traditional Democracy

| Feature | Traditional | TrueRepublic |
|---------|-------------|--------------|
| Speed | Months/Years | Days |
| Transparency | Limited | 100% on-chain |
| Participation | Vote once every 4 years | Vote daily |
| Consensus | Majority rule | Least resistance |

## Getting Started

**Next Steps:**

1. **Users:** [Installation Wizards](Installation-Wizards)
2. **Operators:** [Node Setup](../operations/Node-Setup)
3. **Developers:** [Development Setup](../develop/Development-Setup)

## Further Reading

- [How It Works](How-It-Works) -- Detailed mechanics
- [User Manuals](User-Manuals) -- Complete guides
- [Frontend Guide](Frontend-Guide) -- Web/mobile wallets
- [FAQ](FAQ) -- Common questions
