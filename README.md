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
