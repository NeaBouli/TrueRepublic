# TrueRepublic - Installationsanleitung

## Voraussetzungen
- Go 1.23.0 (Blockchain)
- Rust + Wasm (Smart Contracts)
- Node.js & npm (Web/Mobile Wallets)
- React Native CLI (Mobile Wallet)

## Blockchain-Installation
```bash
git clone https://github.com/NeaBouli/TrueRepublic
cd TrueRepublic/blockchain
go mod tidy
go run app.go
cd ../contracts
cargo build --release --target wasm32-unknown-unknown
wasmd tx wasm store governance.wasm --from wallet --gas auto --fees 10000pnyx
wasmd tx wasm store treasury.wasm --from wallet --gas auto --fees 10000pnyx
cd ../web-wallet
npm install
npm start
cd ../mobile-wallet
npm install
npx react-native run-android  # Android
npx react-native run-ios      # iOS

##### `README.md`
```bash
cat << 'EOF' > README.md
# TrueRepublic Project

## Overview

TrueRepublic is dedicated to enhancing organizational decision-making processes by increasing member participation while safeguarding individual privacy.

## Concept

TrueRepublic organizes participants into **domains**, the primary structure where topics and suggestions are collected and evaluated. Key features include:

- **Privacy and Transparency:** Individual privacy is protected while group opinions are shared securely.
- **Fee and Reward Economy:** Simple economic principles incentivize participation, enhance content quality, prevent spam, and eliminate the need for moderation.
- **Proxy Parties:** [https://pmonien.medium.com/] TrueRepublic aims to enable political proxy parties directly controlled by their participants.

### Further Information
- **Whitepaper:** [https://www.dropbox.com/s/nvdythg6rh42zwc/WhitePaper_TR_eng.pdf?dl=0]
- **Contact:** [t.me/truerepublic](t.me/truerepublic)

---

## Implementation

The project builds on the **Cosmos SDK** (v0.50.13) and uses **CometBFT** (v0.38.21) as its consensus engine.

### Architecture

1. **Base Layer (CometBFT for Consensus):**
   - CometBFT's Byzantine Fault Tolerance (BFT) ensures network-wide consensus on blockchain state, maintaining consistency across nodes.
2. **Application Layer (Custom Logic):**
   - Custom modules in Cosmos SDK handle transactional data (e.g., domains, issues, suggestions) and economic logic (e.g., PNYX tokenomics).
   - Implemented in Go, with additional Rust-based CosmWasm smart contracts for governance and treasury.
3. **Inter-Node Communication:**
   - Nodes communicate using protocols supported by Cosmos SDK (e.g., gRPC), ensuring efficient data processing.
4. **Wallets:**
   - Web (React) and Mobile (React Native) wallets provide user interfaces for PNYX transactions, governance, and DEX operations.

### Additional Project Details
- **Modules:** `truedemocracy` (governance), `dex` (decentralized exchange), `ibc` (inter-blockchain communication), `treasury` (fund management).
- **Smart Contracts:** Governance (systemic consensing) and Treasury (deposit/withdraw) implemented in Rust with CosmWasm.
- **DEX:** PNYX/ATOM AMM pool (`x * y = k`) with 0.3% swap fee and 1% PNYX burn. Additional pairs (BTC, ETH, LUSD) planned via IBC in v0.3.
- **IBC:** Basic cross-chain functionality with potential for Cosmos Hub, Osmosis, and Juno integration.
- **Anonymity:** Prepared for Zero-Knowledge Proofs (ZKP) and key pair-based voting in governance.

---

## How You Can Support TrueRepublic

### 1. **Join the Development Team**
Developers can apply by emailing **[p.cypher@protonmail.com]** with:
- A brief description of their programming background.
- Interest in the project.
Selected contributors will be listed with their BTC addresses for direct funding.

### 2. **Form a Local Group**
Organize local groups to raise funds for developers through crowdfunding initiatives.

### 3. **Donate to Developers**
Directly donate to developers listed below to support ongoing work.

### List of Active Developers
- Team (BTC multi-sig): `bc1qyamf3twgcqckuqrvmwgwnhzupgshxs37eejdgl0ntcqve98qnvhqe6cjl9`

---

## Status (Update: 18/03/25)

### Version: v0.1-alpha
- **Implemented:** 100% of core functionality (TRChain, Domains, Systemic Consensing, Tokenomics, Nodes, DEX, Wallets, Anonymity, PoD).
- **Latest Commit:** Full project rebuilt and pushed on March 18, 2025, with updated dependencies (`cosmos-sdk v0.50.13`, `cometbft v0.38.21`).

### Features
- **TRChain:** Built with Cosmos SDK and Tendermint.
- **Domains:** Include Member-, Issue-, Suggestion-Lists, Treasury-Wallet, and Proof of Domain staking.
- **PNYX Token:** 22M supply with PayToPut, RateToEarn, VoteToEarn mechanics implemented in treasury.
- **Proof of Stake + Proof of Domain (PoD):** Combined staking mechanism with domain-specific staking.
- **Systemic Consensing:** Rating system (-5 to +5) for decision-making in both keeper and smart contracts.
- **DEX:** PNYX/ATOM swaps implemented. Multi-asset expansion (BTC, ETH, LUSD) planned for v0.3 via IBC.
- **Anonymity:** Key pair-based voting in governance smart contract, ZKP-ready.
- **Wallets:** Web (React) and Mobile (React Native) with Keplr integration, real-time balance updates.

### Installation
See `docs/INSTALL.md` for detailed instructions.

### Test Scenario
50 users create a 5-point party program:
- Run: `go run tests/main.go`
- Example: "Flat Tax" wins for "Steuerreform" with Score 15 and 17 Stones (simulated).

### Contributors
- NeaBouli

### Repository
- URL: [https://github.com/NeaBouli/TrueRepublic](https://github.com/NeaBouli/TrueRepublic)

### Roadmap
- **v0.2 (Q2 2025):** Full UI integration (mobile/web app enhancements), ZKP implementation for anonymity.
- **v0.3 (Q3 2025):** Network scalability tests (175+ nodes), multi-asset DEX expansion (ETH, LUSD).
- **v1.0 (Q4 2025):** Mainnet launch with full IBC support and proxy party functionality.

---
