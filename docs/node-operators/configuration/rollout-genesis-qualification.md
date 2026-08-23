# Rollout Genesis Qualification

GH-244 adds an offline, deterministic qualification contract for a future
rollout genesis. It does not create or freeze that genesis, approve a network,
or handle private validator material.

## What the verifier binds

The `truerepublic.rollout-genesis-manifest/v1` manifest binds one exact raw
genesis file to:

- the exact 40-character source commit used as the current unreleased daemon
  version;
- chain ID and SHA-256 of the unmodified genesis bytes;
- at most 64 exact Ed25519 consensus keys, independent Bech32 operator
  authorities, integer power and exactly divisible PNYX stake;
- every initial bank allocation, per-denomination supply and the
  21,000,000-PNYX cap;
- domain admins, membership and active PoD validator/domain relationships;
- the exact `truedemocracy` escrow and DEX reserve custody held by their
  canonical module accounts; and
- exact module-account addresses/permissions, validator operator accounts, and
  isolation of the fee collector, Wasm and transfer module accounts at initial
  state.

The manifest is strict: unknown or missing fields, duplicate keys, trailing
JSON, malformed addresses, duplicate identities, unsorted allocations/coins,
negative or non-canonical amounts, mismatched custody and ambiguous validator
sets fail closed. Ordinary indented Cosmos genesis JSON is accepted; its exact
raw bytes, including whitespace, are covered by the SHA-256 binding.

## Verify offline

The pinned Go toolchain and the complete module cache must be prepared first;
the verifier then runs fully offline from the exact reviewed repository
checkout:

```bash
GOTOOLCHAIN=local GOFLAGS=-mod=readonly GOPROXY=off GOSUMDB=off \
  go run ./cmd/genesis-evidence verify \
  --manifest /offline/review/rollout-genesis-manifest.json \
  --genesis /offline/review/genesis.json \
  --output json
```

The command emits `truerepublic.rollout-genesis-evidence/v1` with stable
manifest/genesis digests and ordered PASS/FAIL checks. It contains no timestamp
or environment-dependent field. Invalid evidence is printed before the command
returns a non-zero status.

Reviewers must still compare the manifest's exact source commit with the
reproducible daemon evidence and independently approve the real allocation,
authority and consensus-parameter choices. GH-244 deliberately ships no
candidate manifest or network genesis, so it completes no rollout checkbox.

## Safety boundary

Do not put private validator keys, signer state, mnemonics, credentials or
private inventory into either file. This verifier performs no tag, signing,
publication, deployment, testnet/mainnet mutation, release freeze or go/no-go
action. `production_ready` remains false until all GH-29 exit gates pass.
