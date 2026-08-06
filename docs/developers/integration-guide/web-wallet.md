# Web Client Integration

`client-web/` is the single maintained browser client. Install and verify it
with the repository lockfile:

```bash
cd client-web
npm ci
npm run lint
npm test
npm run build
```

Its chain metadata lives in `client-web/src/config/chains.ts`; query,
transaction, wallet, governance, DEX, membership, network, and ZKP boundaries
live in `client-web/src/services/`. New browser work must extend this client
and its tests rather than creating another signing-capable frontend.

## Legacy retirement

The previous `web-wallet/` Create React App prototype was removed under
[GH-112](https://github.com/NeaBouli/TrueRepublic/issues/112). The clean
baseline reproduced 70 advisories (including 29 high and 3 critical), broken
custom-query calls, unregistered custom transaction messages, an obsolete DEX
swap, and a functional bank-send route. Migration would have duplicated the
canonical client, so the repository retains the prototype only in Git history
for audit purposes.

The blocking `scripts/check-web-wallet-retirement.sh` contract prevents its
source, non-blocking dependency audit, runtime service, or operational install
instructions from returning.

No legacy client is approved for real keys, signing, broadcasts, or funds.
