<<<<<<< HEAD
=======
# TrueRepublic Project   

## Overview  

TrueRepublic is dedicated to enhance organizational decision-making processes by increasing member participation while safeguarding individual privacy. 

## Concept  

TrueRepublic organizes participants into **domains**, the primary structure where topics and suggestions are collected and evaluated. Key features include:  

- **Privacy and Transparency:** Individual privacy is protected while group opinions are shared securely.  
- **Fee and Reward Economy:** Simple economic principles incentivize participation, enhance content quality, prevent spam, and eliminate the need for moderation.  
- **Proxy Parties:** [https://pmonien.medium.com/] TrueRepublic aims to enable political proxy parties directly controlled by their participants.  

### Further Information  
- **Clip:** [<URL>]  
- **Whitepaper:** [https://www.dropbox.com/s/nvdythg6rh42zwc/WhitePaper_TR_eng.pdf?dl=0]  
- **Contact:** [t.me/truerepublic](t.me/truerepublic)  

---

## Implementation  

The project builds on the **Cosmos SDK** and uses **Tendermint** as its foundation.  

### Architecture  

1. **Base Layer (Tendermint for Consensus):**  
   - Tendermint's Byzantine Fault Tolerance (BFT) ensures network-wide consensus on blockchain state, maintaining consistency across nodes.  

2. **Application Layer (Custom Logic):**  
   - Custom modules in Cosmos SDK handle transactional and non-transactional data.  
   - Each node integrates an external SQL database for non-transactional data. This reduces storage requirements as the state history for this data does not need to be archived. Synchronization is achieved through Cosmos SDK's event system, ensuring identical data operations across nodes.  

3. **Inter-Node Communication:**  
   - Nodes communicate using gRPC or other protocols supported by Cosmos SDK, ensuring efficient data synchronization and processing.  

---

## How You Can Support TrueRepublic  

### 1. **Join the Development Team**  
Developers can apply to join by emailing **[p.cypher@protonmail.com]** with:  
- A brief description of their programming background.  
- Interest in the project.  

Selected contributors will be listed with their BTC addresses to receive direct funding.  

### 2. **Form a Local Group**  
Organize local groups to raise funds for developers through crowdfunding initiatives.  

### 3. **Donate to Developers**  
Directly donate to developers listed in this repository to support ongoing work. 

### List of active developers:
Team (btc multi. sig): bc1qyamf3twgcqckuqrvmwgwnhzupgshxs37eejdgl0ntcqve98qnvhqe6cjl9

==============================================================================


# TrueRepublic
Decentralized Democracy Tool based on Cosmos SDK

## Status
- Version: v0.1-alpha
- Implemented: 100% of core functionality (TRChain, Domains, Systemic Consensing, Tokenomics, Nodes, Anonymity, DEX, UI)

## Features
- TRChain with Cosmos SDK and Tendermint
- Domains with Member-, Issue-, Suggestion-Lists and Treasury-Wallet
- PNYX Token (21M supply) with PayToPut, RateToEarn, VoteToEarn
- Proof of Stake + Proof of Domain (PoD)
- Systemic Consensing (-5 to +5)
- Real-time feedback via binary tree
- DEX for PNYX-BTC/ETH/LUSD swaps
- Anonymity with global and domain-specific key pairs

## Installation
1. Install Go: `go install`
2. Run the blockchain: `go run app.go`
3. Compile UI: `g++ ui.cpp -o truerepublic_ui`

## Test Scenario
50 users create a 5-point party program:
- `go test -v main_test.go`
- Example: "Flat Tax" wins for "Steuerreform" with Score 15 and 17 Stones

## Contributors
- Developed with Platonas (xAI)

## Repository
- URL: https://github.com/NeaBouli/TrueRepublic

## Roadmap
- v0.2: Full mobile app UI
- v0.3: Network scalability tests (175+ nodes)
