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

For the ephemeral local client-to-chain delivery gate:

```bash
make -C .. build
TRUEREPUBLIC_CLIENT_CHAIN_INTEGRATION=1 npm run test:chain
```

Its chain metadata lives in `client-web/src/config/chains.ts`; query,
transaction, wallet, governance, DEX, membership, network, and ZKP boundaries
live in `client-web/src/services/`. New browser work must extend this client
and its tests rather than creating another signing-capable frontend.

## Canonical signing registry

All maintained signing clients are created through
`client-web/src/services/signingClient.ts`. Its registry contains CosmJS'
default types plus the exact reviewed custom boundary below:

- `/truedemocracy.MsgCreateDomain`
- `/truedemocracy.MsgSubmitProposal`
- `/truedemocracy.MsgPlaceStoneOnSuggestion`
- `/truedemocracy.MsgPlaceStoneOnIssue`
- `/truedemocracy.MsgApproveOnboarding`
- `/truedemocracy.MsgAddMember`
- `/truedemocracy.MsgOnboardToDomain`
- `/truedemocracy.MsgRegisterIdentity`
- `/dex.MsgAddLiquidity`
- `/dex.MsgRemoveLiquidity`
- `/dex.MsgSwapExact`

Do not instantiate an ad-hoc signing client or register legacy aliases. Unknown
type URLs fail before signing; amounts use decimal strings and protobuf-safe
integer conversion. The canonical account prefix is `truerepublic`.

GH-115's gated integration test builds a temporary local chain and uses only
generated disposable accounts. It proves standard bank send plus each supported
custom transaction family, an expected authorization failure, and rejection of
the disabled legacy DEX swap. It is repository evidence only and does not
authorize production RPCs, keys, accounts, broadcasts, or funds.

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
