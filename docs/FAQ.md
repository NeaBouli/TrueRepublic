# Frequently Asked Questions

## General

### What is TrueRepublic?
TrueRepublic is a platform for **direct democracy** and **digital self-determination** built on the Cosmos SDK blockchain. Instead of electing representatives, participants make decisions directly through domains, proposals, and voting.

### What is PNYX?
PNYX is the native token of the TrueRepublic blockchain, named after the Pnyx hill in Athens where ancient Greek citizens gathered for democratic assemblies. It is used for governance, treasury funding, staking, and DEX trading.

### What is Systemic Consensing?
A decision-making method that measures **resistance** on a scale from -5 to +5, instead of simple Yes/No voting. It finds solutions with the least overall resistance, leading to better outcomes for the community. See [Systemic Consensing Explained](user-manual/systemic-consensing-explained.md).

### What is Proof of Domain (PoD)?
TrueRepublic's consensus mechanism that requires validators to be **active members of at least one domain**. This ensures validators are invested in governance, not just financially. See [Validator Guide](validators/README.md).

### How is TrueRepublic different from other governance platforms?
- **On-chain governance** -- All decisions are recorded on the blockchain
- **Systemic Consensing** -- Better than Yes/No voting
- **Anonymous voting** -- Domain key pairs protect voter identity
- **Stones voting** -- Highlights importance, elects admins
- **VoteToEarn** -- Rewards active participation
- **Proof of Domain** -- Validators must be community members

## Tokens & Economics

### What is the total supply of PNYX?
Maximum supply is **22,000,000 PNYX**. New PNYX is minted through staking rewards and domain interest, subject to release decay as supply approaches the cap.

### How do I get PNYX?
- **Testnet:** From the faucet
- **Mainnet:** From the DEX, or by participating in governance (VoteToEarn rewards)
- **Staking:** Validators earn 10% APY

### What are the fees?
- **Transaction gas:** ~0.001 PNYX (auto-calculated)
- **PayToPut (proposals):** `treasury / 1000 * min(15, members)` -- varies by domain
- **DEX swap fee:** 0.3%
- **DEX PNYX burn:** 1% on PNYX output

### What is VoteToEarn?
When you place a stone (vote), you earn PNYX from the domain treasury. The reward is `treasury / 1000` per stone placement. See [Stones Voting Guide](user-manual/stones-voting-guide.md).

### What is PayToPut?
A fee charged when submitting proposals. It prevents spam and funds the domain treasury. The fee scales with domain size.

## Governance

### What is a Domain?
A community-governed space focused on a specific topic. Domains have members, a treasury, issues, suggestions, and an elected admin. See [Governance Tutorial](user-manual/governance-tutorial.md).

### How do I create a Domain?
```bash
truerepublicd tx truedemocracy create-domain [name] [initial-coins]pnyx \
    --from mykey --chain-id truerepublic-1
```

### How do I submit a proposal?
Via the web wallet (Governance page, right panel) or CLI:
```bash
truerepublicd tx truedemocracy submit-proposal [domain] [issue] [suggestion] [fee]pnyx "" \
    --from mykey --chain-id truerepublic-1
```

### How is the admin elected?
The domain admin is the member with the **most stones** from other members. Admin election is continuous -- it updates whenever stones change. See [Stones Voting Guide](user-manual/stones-voting-guide.md).

### Can a member be removed?
Yes, by **2/3 majority vote** of domain members. Any member can initiate a vote-exclude transaction.

### What are the suggestion lifecycle zones?
| Zone | Meaning |
|------|---------|
| Green | New/active suggestion |
| Yellow | Mature, has been evaluated |
| Red | Expiring, will be auto-deleted if not supported |

### Is voting anonymous?
Yes. Members register domain-specific key pairs via the permission register. Ratings are submitted with domain public keys, making them **unlinkable** to member identities.

## Wallet

### Which wallets are supported?
- **Keplr** (browser extension) -- Recommended for web
- **Mobile wallet** (React Native app) -- For iOS/Android

### How long do transactions take?
Approximately **5 seconds** (one block confirmation).

### What if my transaction fails?
Common causes: insufficient funds, wrong network, pending transaction. See [Troubleshooting](user-manual/troubleshooting.md).

## Node Operation

### What are the hardware requirements?
Minimum: 2 CPU, 4 GB RAM, 100 GB SSD. Recommended: 4+ CPU, 8+ GB RAM, 250+ GB NVMe SSD. See [Requirements](node-operators/installation/requirements.md).

### Docker or native build?
Docker is recommended for most operators. It includes the full stack (node, web wallet, monitoring). See [Docker Setup](node-operators/installation/docker-setup.md).

### What ports need to be open?
Only **26656/tcp** (P2P) is required. Optionally open 26657 (RPC) if serving public queries.

## Validators

### How do I become a validator?
1. Run a full node
2. Join or create a domain
3. Stake minimum 100,000 PNYX
4. Register as validator

See [Validator Guide](validators/README.md).

### What happens if my validator goes down?
If you miss more than 50 out of 100 blocks, you're slashed 1% and jailed for 10 minutes. See [Slashing](validators/README.md#slashing).

### What are staking rewards?
Validators earn **10% APY** on their stake, subject to release decay. Rewards are distributed every hour.

### What is the transfer limit?
Validator stake withdrawals are capped at **10% of the domain's total payouts**. This prevents value extraction.

## Development

### What is the tech stack?
Go 1.23 (blockchain), Cosmos SDK v0.50.13, CometBFT v0.38.17, React 18 (web), Expo/React Native (mobile), Rust/CosmWasm (smart contracts).

### How do I contribute?
Fork the repo, create a branch, write tests, and submit a PR. See [Developer Docs](developers/README.md).

### Where are the tests?
182 tests across 3 modules. Run with `make test` or `go test ./... -race -cover`.

### How do I integrate with CosmJS?
See [CosmJS Examples](developers/integration-guide/cosmjs-examples.md) for complete code samples.
