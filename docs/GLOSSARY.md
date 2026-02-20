# Glossary

| Term | Definition |
|------|-----------|
| **ABCI** | Application Blockchain Interface -- the protocol between CometBFT and the application layer |
| **Admin** | The elected leader of a domain, determined by stone voting |
| **AMM** | Automated Market Maker -- the DEX model where prices are set by a formula (x*y=k) |
| **Anonymous Voting** | Voting with domain key pairs so ratings cannot be linked to member identity |
| **Approval Threshold** | Minimum average rating for a suggestion to be approved (default 5%) |
| **Bech32** | Address encoding format; TrueRepublic addresses start with `truerepublic1` |
| **BIP44** | HD wallet standard for key derivation; TrueRepublic uses coin type 118 |
| **Block** | A batch of transactions committed to the blockchain (~5 seconds per block) |
| **CEarn** | Tokenomics constant (1000) used as the divisor for VoteToEarn rewards |
| **CometBFT** | The Byzantine Fault Tolerant consensus engine (formerly Tendermint) |
| **Constant-Product** | AMM formula x*y=k that determines swap prices in liquidity pools |
| **CosmJS** | JavaScript library for interacting with Cosmos SDK blockchains |
| **Cosmos SDK** | The Go framework TrueRepublic is built on (v0.50.13) |
| **CosmWasm** | WebAssembly smart contract platform for Cosmos chains |
| **CPut** | Tokenomics constant (15) that caps the PayToPut price |
| **CDom** | Tokenomics constant (2) used as domain creation cost multiplier |
| **DEX** | Decentralized Exchange -- trade tokens without a central authority |
| **Domain** | A community governance space focused on a specific topic |
| **Domain Key Pair** | Ed25519 key pair used for anonymous voting within a domain |
| **Double-Signing** | A validator signing two different blocks at the same height (slashable offense) |
| **Downtime** | When a validator fails to sign blocks (slashable if >50 missed in 100) |
| **Dwell Time** | How long a suggestion stays in each lifecycle zone (default 1 day) |
| **EndBlock** | Application logic executed at the end of each block (rewards, governance, lifecycle) |
| **Fast-Delete** | Immediate deletion of a suggestion by 2/3 majority vote |
| **Full Node** | A node that maintains the full blockchain state but doesn't validate |
| **Gas** | Computational cost of a transaction; users pay gas fees |
| **Genesis** | The initial state of the blockchain (genesis.json) |
| **Green Zone** | First lifecycle zone for new suggestions |
| **gRPC** | Remote procedure call protocol used by Cosmos SDK (port 9090) |
| **HSM** | Hardware Security Module -- dedicated hardware for storing validator keys |
| **IBC** | Inter-Blockchain Communication -- protocol for cross-chain transfers |
| **Impermanent Loss** | Potential loss from providing liquidity when token prices diverge |
| **Issue** | A problem or topic raised within a domain |
| **Jailing** | Temporarily removing a validator from the active set after slashing |
| **Keplr** | Browser extension wallet for Cosmos chains |
| **KV Store** | Key-Value store used by each module for persistent state |
| **LCD** | Light Client Daemon -- REST API endpoint (port 1317) |
| **Lifecycle** | The progression of suggestions through green/yellow/red zones |
| **LP Shares** | Liquidity Provider shares representing ownership in a DEX pool |
| **Mempool** | Queue of pending transactions waiting to be included in a block |
| **Node** | A computer running the TrueRepublic blockchain software |
| **PayToPut** | Fee charged when submitting proposals (prevents spam) |
| **Permission Register** | List of domain public keys used for anonymous voting |
| **PNYX** | The native token of TrueRepublic (max supply: 22,000,000) |
| **PoD** | Proof of Domain -- consensus requiring validators to be domain members |
| **Pool** | A liquidity pool on the DEX containing two tokens |
| **Pruning** | Removing old blockchain state to save storage |
| **Purge** | Admin action to reset the permission register (enables new anonymous keys) |
| **Rating** | Systemic consensing score from -5 to +5 on a suggestion |
| **Red Zone** | Third lifecycle zone; suggestions here are auto-deleted when time expires |
| **Release Decay** | Mechanism that reduces new token issuance as supply approaches max |
| **RPC** | Remote Procedure Call -- primary API endpoint (CometBFT, port 26657) |
| **Seed Node** | A node used for initial peer discovery; disconnects after sharing peers |
| **Sentry Node** | A public-facing node that shields a validator from direct internet exposure |
| **Slashing** | Penalty (stake reduction) for validator misbehavior |
| **Stake** | PNYX locked by a validator to participate in consensus |
| **StakeMin** | Minimum validator stake: 100,000 PNYX |
| **Stone** | A vote token placed on issues, suggestions, or members |
| **Suggestion** | A proposed solution to an issue within a domain |
| **SupplyMax** | Maximum total PNYX supply: 22,000,000 |
| **Swap Fee** | DEX trading fee of 0.3% per swap |
| **Systemic Consensing** | Decision-making by measuring resistance (-5 to +5) instead of Yes/No |
| **Transfer Limit** | Validator withdrawal cap at 10% of domain total payouts |
| **Treasury** | PNYX funds held by a domain for community initiatives |
| **Validator** | A node that participates in consensus (proposing and signing blocks) |
| **Voting Power** | A validator's influence in consensus = stake / StakeMin |
| **VoteToEarn** | Rewards earned by placing stones (treasury / 1000 per placement) |
| **Yellow Zone** | Second lifecycle zone for mature suggestions |
