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

## 📂 Repository structure & status

```text
TrueRepublic/
 ├── blockchain/        ✅  Cosmos SDK chain (modules: truedemocracy, dex, treasury)
 ├── contracts/         ✅  CosmWasm smart contracts (governance)
 ├── web-wallet/        ✅  React web wallet (Kepl / Keplr integration)
 ├── mobile-wallet/     🔵  React Native mobile wallet (basic version; features pending)
 ├── docs/              ✅  White Paper (MD + PDF), API, DEX, INSTALL
 ├── scripts/           🔵  DevOps & deployment scripts (planned)
 ├── tests/             🔴  Unit & E2E tests largely missing
 └── .github/
     ├── ISSUE_TEMPLATE ✅  available
     └── workflows/     🔵  CI/CD workflows added (security scans pending)
📑 Documentation (quick links)
Structured White Paper (Markdown): docs/WhitePaper_TR.md

TrueRepublic Native White Paper (PDF): docs/WhitePaper_TR_eng.pdf

Security Policy: SECURITY.md

CI/CD Security Guide: TrueRepublic_CI_CD_Security.pdf

API & DEX docs (skeletons):

docs/API.md 🔵

docs/DEX.md 🔵

🛠️ Build & development (commands)
Blockchain (Cosmos SDK)
bash
Code kopieren
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
Web wallet (React)
bash
Code kopieren
cd web-wallet
npm ci
npm test
npm run build
Mobile wallet (React Native)
bash
Code kopieren
cd mobile-wallet
npm ci
npm test
✅ Current status (short)
✅ White Paper (Markdown + PDF) added

✅ README cleaned and standardized (this file)

🔵 CI/CD workflows prepared but not all enabled / security scans pending

🔴 Tests & automated security scanning still required across stack

🚀 Immediate next priorities (recommended)
Add minimal unit tests for each blockchain module (truedemocracy, dex, treasury) — 1 happy / 1 error path each.

Enable CI workflows (Go / Rust / Web / Mobile) and add SAST / dependency scanning (Trivy / Grype).

Create placeholder files so folder structure is visible on GitHub (.keep) for scripts/ and tests/.

Add CONTRIBUTING.md and a short developer onboarding guide (docs/INSTALL.md) explaining how to bring up a local devnet.

🧭 If you want — next immediate actions I will perform for you:
Create and push .keep placeholders for scripts/ and tests/.

Add a minimal CONTRIBUTING.md and docs/INSTALL.md skeleton.

Prepare 3 minimal unit test stubs (Go) for the blockchain modules as EOF patches you can apply.

Tell me which of those three you want next and I'll produce the exact shell-blocks (paste-ready) — I'll do them one at a time so you can confirm after each push.
