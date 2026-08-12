<p align="center">
  <img src="https://raw.githubusercontent.com/NeaBouli/TrueRepublic/main/assets/logo.png" alt="TrueRepublic Logo" width="200"/>
</p>

# TrueRepublic / PNYX Technical Wiki

> **Recovery audit active — not production-ready.** The verified recovery
> foundation is merged to `main`. Do not use the project for real keys, funds,
> anonymous voting, or a public network until the remaining independent
> security and operations reviews pass.

## Current recovery evidence

| Item | Verified state |
|---|---|
| Version label | v0.4.0 recovery |
| Tests | 1,610 standard-suite total: 1,443 Go, 26 Rust, 141 maintained-client; separate GH-175/GH-178/GH-181 IBC gate excluded |
| PNYX cap | 21,000,000 PNYX = 21,000,000,000,000 `upnyx` |
| Node | Restart, four-validator recovery, state sync, sanitized backup/restore, compatible binary rollback, cold identity failover, secret-safe JSON logs, private metrics, and the GH-85 dashboard/alert/objective baseline verified |
| ZKP client | Mock generation/submission disabled; real prover pending |
| Maintained client | `client-web` |
| Legacy clients | Web and mobile prototypes retired under GH-112/GH-102; Git history only |

Authoritative machine status: [`docs/status.json`](https://github.com/NeaBouli/TrueRepublic/blob/main/docs/status.json).
Recovery tracking: [Issue #4](https://github.com/NeaBouli/TrueRepublic/issues/4).

## Navigation

### Developers

- [Architecture Overview](develop/Architecture-Overview)
- [Code Structure](develop/Code-Structure)
- [Module Deep-Dive](develop/Module-Deep-Dive)

### Users

- [System Overview](users/System-Overview)
- [Installation Wizards](users/Installation-Wizards)
- [User Manuals](users/User-Manuals)
- [How It Works](users/How-It-Works)

### Node operators

- [Node Setup](operations/Node-Setup)
- [Validator Guide](operations/Validator-Guide)
- [Deployment Options](operations/Deployment-Options)
- [Monitoring](operations/Monitoring)
- [Troubleshooting](operations/Troubleshooting)

### Security and status

- [Current Status](status/Current-Status)
- [Testing Status](status/Testing-Status)
- [Audit Reports](security/Audit-Reports)
- [Known Issues](security/Known-Issues)
- [Security Architecture](security/Security-Architecture)
- [Cross-System Threat Model](https://github.com/NeaBouli/TrueRepublic/blob/main/docs/security/THREAT_MODEL.md)
- [Best Practices](security/Best-Practices)

## Technology baseline

| Layer | Recovery version |
|---|---|
| Go | 1.26.5 |
| Cosmos SDK | v0.50.15 |
| CometBFT | v0.38.25 |
| ibc-go | v8.7.0 |
| wasmd / wasmvm | v0.53.4 / v2.2.2 |
| Maintained web client | React 18.2, TypeScript 5.9, Vite 8.2, CosmJS 0.39 |

Historical milestone documents describe implemented surface area, not current
production approval. Use the status and audit pages above for current claims.
