# TrueRepublic / PNYX

[![Go CI](https://github.com/NeaBouli/TrueRepublic/actions/workflows/go-ci.yml/badge.svg)](https://github.com/NeaBouli/TrueRepublic/actions/workflows/go-ci.yml)
[![Rust CI](https://github.com/NeaBouli/TrueRepublic/actions/workflows/rust-ci.yml/badge.svg)](https://github.com/NeaBouli/TrueRepublic/actions/workflows/rust-ci.yml)
[![Web CI](https://github.com/NeaBouli/TrueRepublic/actions/workflows/react-ci.yml/badge.svg)](https://github.com/NeaBouli/TrueRepublic/actions/workflows/react-ci.yml)
[![Mobile CI](https://github.com/NeaBouli/TrueRepublic/actions/workflows/react-native-ci.yml/badge.svg)](https://github.com/NeaBouli/TrueRepublic/actions/workflows/react-native-ci.yml)

---

## 🌍 Vision
TrueRepublic is a platform for **direct democracy** and **digital self-determination**.  
The token **PNYX** enables governance, treasury mechanisms and a decentralized DEX.  

---

## 📂 Repository Structure & Status

TrueRepublic/
├── blockchain/ ✅ Cosmos SDK Chain (modules: truedemocracy, dex, treasury)
├── contracts/ ✅ CosmWasm Smart Contracts (governance)
├── web-wallet/ ✅ React Web Wallet (Keplr integration)
├── mobile-wallet/ 🔵 React Native Mobile Wallet (basic version, features pending)
├── docs/ ✅ White Paper, API, DEX, Install
├── scripts/ 🔵 DevOps & deployment (planned)
├── tests/ 🔴 Unit & E2E tests largely missing
└── .github/
├── ISSUE_TEMPLATE ✅ available
└── workflows/ 🔵 CI/CD added, security scans pending

yaml
Code kopieren

---

## 📑 Documentation

- [Structured White Paper (Markdown)](docs/WhitePaper_TR.md)  
- [TrueRepublic Native White Paper (PDF)](docs/WhitePaper_TR_eng.pdf)  
- [Security Policy](SECURITY.md)  
- [CI/CD Security Guide](TrueRepublic_CI_CD_Security.pdf)  
- API & DEX Docs:  
  - [API.md](docs/API.md) 🔵  
  - [DEX.md](docs/DEX.md) 🔵  

---

## 🛠️ Build & Development

### Blockchain
```bash
cd blockchain
go mod tidy
go build ./...
go test ./... -race -cover
Contracts (CosmWasm)
bash
Code kopieren
cd contracts
cargo fmt --all -- --check
cargo clippy --all-targets -- -D warnings
cargo test --all
Web Wallet
bash
Code kopieren
cd web-wallet
npm ci
npm test
npm run build
Mobile Wallet
bash
Code kopieren
cd mobile-wallet
npm ci
npm test
🚀 Improvements & To-dos
Blockchain: add more unit tests per module (happy & error paths)

Contracts: modularization, strict clippy enforcement

Wallets: more mock & E2E tests (Jest, Detox)

CI/CD: add security scans (Trivy/Grype), SBOM generation

Docs: complete API/DEX documentation

📌 Current Status
✅ Repo now has White Paper (Markdown + PDF) and structured README

🔵 CI/CD workflows prepared, integration pending

🔴 Tests & security checks still missing

