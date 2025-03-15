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

The project builds on the **Cosmos SDK** (v0.50.13) and uses **Tendermint** (v0.35.9) as its foundation.

### Architecture

1. **Base Layer (Tendermint for Consensus):**
   - Tendermint's Byzantine Fault Tolerance (BFT) ensures network-wide consensus on blockchain state, maintaining consistency across nodes.
2. **Application Layer (Custom Logic):**
   - Custom modules in Cosmos SDK handle transactional data (e.g., domains, issues, suggestions) and economic logic (e.g., PNYX tokenomics).
   - Currently implemented in Go, with plans for additional data synchronization features in future releases.
3. **Inter-Node Communication:**
   - Nodes communicate using protocols supported by Cosmos SDK (e.g., gRPC), ensuring efficient data processing.

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

## Status (Update: 15/03/25)

### Version: v0.1-alpha
- **Implemented:** 100% of core functionality (TRChain, Domains, Systemic Consensing, Tokenomics, Nodes, Anonymity, DEX).
- **Latest Commit:** Updated dependencies to `cosmos-sdk v0.50.13` and `tendermint v0.35.9` on March 15, 2025.

### Features
- **TRChain:** Built with Cosmos SDK and Tendermint.
- **Domains:** Include Member-, Issue-, Suggestion-Lists and Treasury-Wallet.
- **PNYX Token:** 21M supply with PayToPut, RateToEarn, VoteToEarn mechanics.
- **Proof of Stake + Proof of Domain (PoD):** Combined staking mechanism.
- **Systemic Consensing:** Rating system (-5 to +5) for decision-making.
- **DEX:** PNYX-BTC swaps implemented (ETH/LUSD planned).
- **Anonymity:** Global and domain-specific key pairs.

### Installation
1. Install Go: `go install`
2. Clone the repository: `git clone https://github.com/NeaBouli/TrueRepublic`
3. Install dependencies: `go mod tidy`
4. Run the blockchain: `go run app.go`
5. Run tests: `go run tests/main.go`

### Test Scenario
50 users create a 5-point party program:
- Run: `go run tests/main.go`
- Example: "Flat Tax" wins for "Steuerreform" with Score 15 and 17 Stones.

### Contributors
- NeaBluli GioMario.

### Repository
- URL: [https://github.com/NeaBouli/TrueRepublic](https://github.com/NeaBouli/TrueRepublic)

### Roadmap
- **v0.2:** Add UI (mobile/web app integration).
- **v0.3:** Network scalability tests (175+ nodes).
