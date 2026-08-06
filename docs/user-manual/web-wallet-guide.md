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

Before any release use, complete the open rollout gates in
[ROLLOUT_ROADMAP.md](../ROLLOUT_ROADMAP.md), including wallet/key review,
client-to-chain end-to-end tests, real ZKP integration, accessibility, browser
support, and external security review.
