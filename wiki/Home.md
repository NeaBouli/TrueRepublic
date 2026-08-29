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
| Tests | 2,094 standard-suite total: 1,749 Go (including GH-258 repeated-OCI evidence, GH-244 rollout-genesis qualification, GH-225 release compatibility, GH-222 install-lifecycle, GH-209 recipient-binding and GH-206 pinned test-only prover coverage), 26 Rust, 319 maintained-client; separate GH-206/GH-209 Go/WASM compatibility, GH-175/GH-178/GH-181 IBC, and GH-184 governed-upgrade gates excluded |
| PNYX cap | 21,000,000 PNYX = 21,000,000,000,000 `upnyx` |
| Node | Restart, four-validator recovery, state sync, sanitized backup/restore, compatible binary rollback, cold identity failover, secret-safe JSON logs, private metrics, and the GH-85 dashboard/alert/objective baseline verified |
| ZKP client | Real synthetic Go/WASM compatibility verified on GH-206 and recipient-bound rewards on GH-209; production submission remains hard-disabled |
| Maintained client | `client-web` |
| Legacy clients | Web and mobile prototypes retired under GH-112/GH-102; Git history only |
| Project licensing | Maintained source and documentation are Apache-2.0; individual contributors retain copyright, collectively attributed as “TrueRepublic contributors”; brand/art assets, historical PDFs, archived evidence, and third-party material remain excluded unless an applicable file-specific notice exists or provenance and permission are documented |

Authoritative machine status: [`docs/status.json`](https://github.com/NeaBouli/TrueRepublic/blob/main/docs/status.json).
Completed recovery foundation:
[Issue #4](https://github.com/NeaBouli/TrueRepublic/issues/4). Active rollout
tracking: [Issue #29](https://github.com/NeaBouli/TrueRepublic/issues/29).
The exact community licensing decision is recorded in
[GH-219](https://github.com/NeaBouli/TrueRepublic/issues/219#issuecomment-5423337355).

## Navigation

### Developers

- [Architecture Overview](develop/Architecture-Overview)
- [Optional Ballot Architecture](https://github.com/NeaBouli/TrueRepublic/blob/main/docs/GOVERNANCE_BALLOT_ARCHITECTURE.md)
- [Sovereign V4 Edge Architecture](https://github.com/NeaBouli/TrueRepublic/blob/main/docs/SOVEREIGN_V4_EDGE_ARCHITECTURE.md)
- [Code Structure](develop/Code-Structure)
- [Module Deep-Dive](develop/Module-Deep-Dive)
- [API Reference](develop/API-Reference)
- [Development Setup](develop/Development-Setup)
- [Contributing Guide](develop/Contributing-Guide)

### Users

- [System Overview](users/System-Overview)
- [Installation Wizards](users/Installation-Wizards)
- [User Manuals](users/User-Manuals)
- [How It Works](users/How-It-Works)
- [FAQ](users/FAQ)

### Node operators

- [Node Setup](operations/Node-Setup)
- [Validator Guide](operations/Validator-Guide)
- [Deployment Options](operations/Deployment-Options)
- [Monitoring](operations/Monitoring)
- [Troubleshooting](operations/Troubleshooting)

### Security and status

- [Current Status](status/Current-Status)
- [Roadmap](status/Roadmap)
- [Feature Matrix](status/Feature-Matrix)
- [Known Bugs and Limitations](status/Known-Bugs)
- [Testing Status](status/Testing-Status)
- [Audit Reports](security/Audit-Reports)
- [Known Issues](security/Known-Issues)
- [Security Architecture](security/Security-Architecture)
- [Cross-System Threat Model](https://github.com/NeaBouli/TrueRepublic/blob/main/docs/security/THREAT_MODEL.md)
- [Best Practices](security/Best-Practices)

## Technology baseline

| Layer | Recovery version |
|---|---|
| Go | 1.26.6 |
| Cosmos SDK | v0.50.15 |
| CometBFT | v0.38.26 |
| ibc-go | v8.7.0 |
| wasmd / wasmvm | v0.53.4 / v2.2.8 |
| Maintained web client | React 18.2, TypeScript 5.9, Vite 8.2, CosmJS 0.39 |

Historical milestone documents describe implemented surface area, not current
production approval. Use the status and audit pages above for current claims.
