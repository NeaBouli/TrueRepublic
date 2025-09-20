# TrueRepublic / PNYX

[![Go CI](https://github.com/NeaBouli/TrueRepublic/actions/workflows/go-ci.yml/badge.svg)](https://github.com/NeaBouli/TrueRepublic/actions/workflows/go-ci.yml)
[![Rust CI](https://github.com/NeaBouli/TrueRepublic/actions/workflows/rust-ci.yml/badge.svg)](https://github.com/NeaBouli/TrueRepublic/actions/workflows/rust-ci.yml)
[![Web CI](https://github.com/NeaBouli/TrueRepublic/actions/workflows/react-ci.yml/badge.svg)](https://github.com/NeaBouli/TrueRepublic/actions/workflows/react-ci.yml)
[![Mobile CI](https://github.com/NeaBouli/TrueRepublic/actions/workflows/react-native-ci.yml/badge.svg)](https://github.com/NeaBouli/TrueRepublic/actions/workflows/react-native-ci.yml)

---

## 🌍 Vision
TrueRepublic ist die Plattform für **direkte Demokratie** und **digitale Selbstbestimmung**.  
Der Token **PNYX** ermöglicht Governance, Treasury-Mechanismen und einen dezentralen DEX.  

---

## 📂 Ordnerstruktur & Status

TrueRepublic/
├── blockchain/ ✅ Cosmos SDK Chain (Module: truedemocracy, dex, treasury)
├── contracts/ ✅ CosmWasm Smart Contracts (Governance)
├── web-wallet/ ✅ React Web Wallet (Keplr-Integration)
├── mobile-wallet/ 🔵 React Native Mobile Wallet (Basis vorhanden, Features offen)
├── docs/ ✅ White Paper, API, DEX, Install
├── scripts/ 🔵 DevOps & Deployment (geplant)
├── tests/ 🔴 Unit- & E2E-Tests fehlen weitgehend
└── .github/
├── ISSUE_TEMPLATE ✅ vorhanden
└── workflows/ 🔵 CI/CD ergänzt, Security-Scans noch offen

yaml
Code kopieren

---

## 📑 Dokumentation

- [White Paper (Markdown)](docs/WhitePaper_TR.md)  
- [Security Policy](SECURITY.md)  
- [CI/CD Security Guide](TrueRepublic_CI_CD_Security.pdf)  
- API & DEX Docs:  
  - [API.md](docs/API.md) 🔵  
  - [DEX.md](docs/DEX.md) 🔵  

---

## 🛠️ Build & Entwicklung

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
🚀 Verbesserungen & To-dos
Blockchain: mehr Unit-Tests in jedem Modul (happy & error paths)

Contracts: Modularisierung, Clippy strikt enforced

Wallets: mehr Mock- & E2E-Tests (Jest, Detox)

CI/CD: Security-Scans (Trivy/Grype), SBOM-Generierung

Docs: API/DEX-Dokumentation vervollständigen

📌 Status-Quo
✅ Repo ist jetzt mit White Paper & strukturierter README ausgestattet

🔵 CI/CD Workflows vorbereitet, müssen ins Repo integriert werden

🔴 Tests & Security Checks fehlen größtenteils

