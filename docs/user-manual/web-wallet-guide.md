# Maintained Web Client Guide

`client-web` is the only maintained TrueRepublic browser client. The former
`web-wallet` prototype was retired and removed under
[GH-112](https://github.com/NeaBouli/TrueRepublic/issues/112); do not recover it
from Git history for real keys or funds.

## Current recovery boundary

The repository is still in recovery and is not production-ready. Use local or
explicitly approved test environments only. Anonymous-vote submission remains
disabled until a compatible real prover and independent cryptographic review
exist.

For local development:

```bash
git clone https://github.com/NeaBouli/TrueRepublic.git
cd TrueRepublic/client-web
npm ci
npm run dev
```

The development server listens on `http://localhost:3001`. The Docker Compose
profile builds the same maintained client and exposes it on loopback port 3001;
the reverse-proxy entry point is `http://localhost:8080`.

GH-193 qualifies the local maintained-client wallet boundary: bounded BIP-39
import, versioned encrypted storage with legacy re-encryption, exact account and
chain binding, and signer invalidation on lock/switch/delete/reload. This is not
approval for real keys or funds and does not protect a compromised browser or
same-origin script. Before release use, complete the remaining gates in
[ROLLOUT_ROADMAP.md](../ROLLOUT_ROADMAP.md), including real ZKP integration,
external review, production custody and staged-rollout evidence.
