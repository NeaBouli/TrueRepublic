# Legacy Validator-Authority Migration

Status: recovery-stage procedure for GH-61. This is not a production rollout
approval and is not a general-purpose `x/upgrade` or governance migration.

Use [Incident Command and Rehearsal](incident-command.md) for the accountable
halt, approval, evidence, target validation, and rollback decision record.

This runbook covers the one-time transition from pre-GH-56 state, where
validator operator account addresses were derived from CometBFT consensus keys,
to a reviewed fresh genesis with independently controlled operator accounts.
The canonical implementation is:

```text
truerepublicd migration legacy-authority
```

## Trust and authorization boundary

Pre-GH-56 state has no independent on-chain governance signer authorized to
rewrite validator identity. Legacy validator signatures cannot solve that
problem because the legacy operator accounts and consensus keys share the same
authority.

The supported transition is therefore an explicitly reviewed fresh-genesis
ceremony. The migration descriptor:

- binds the source and distinct target chain IDs;
- binds the exact halt height and trusted 32-byte source app hash;
- binds SHA-256 of the exact raw source-export bytes;
- identifies one deterministic transform;
- maps every consensus-key-derived operator to one fresh account key; and
- carries proof of possession from every fresh replacement key over the same
  canonical descriptor bytes.

Those proofs authenticate fresh-key possession only. They do not create
retroactive governance approval. The accountable reviewers and operators must
approve the final descriptor and transformed genesis out of band.

## Supported state boundary

The transformer reconciles:

- all active, inactive, historical, revoked, pending-rotation, and
  pending-removal validator references;
- domain administrators and members, suggestion creators, bootstrap
  operators, signing records, and infractions;
- Auth `BaseAccount` address/public key while preserving account number and
  sequence;
- Bank balance ownership without changing supply or module custody;
- DEX LP providers and asset-registration authorities; and
- a populated CometBFT genesis validator set against exact active application
  keys and power.

It rejects nonempty CosmWasm code or contract state. A contract-aware adapter
and separate review are required before such a chain can use this procedure.
Unknown modules are preserved, but any remaining literal legacy operator
address aborts the transform.

## Preconditions

Do not begin unless every item is true:

1. The exact source binary and all validator homes have offline, access-
   controlled recovery copies.
2. Every replacement operator uses a newly generated account key that is not
   derived from any active, historical, pending, or revoked consensus key.
3. The target chain ID differs from the source chain ID.
4. Every required descriptor mapping is prepared. Fresh-key proofs are created
   only after the exact halted export is frozen and its SHA-256 is placed in
   the descriptor.
5. The source has no CosmWasm code or contract state.
6. At least two trusted RPC observations agree on the halt-height block ID,
   app hash, validator set, and power.
7. A rollback owner, decision deadline, and communication channel are recorded.
8. Operators can prove that source and target signers cannot run
   simultaneously.

Never copy private keys, keyring contents, or signer state into the descriptor,
logs, issue comments, or shared review artifacts.

## Descriptor

The descriptor is strict JSON. Byte fields use standard JSON base64 encoding.
Mappings are sorted by decoded old-operator address bytes, not by their printed
bech32 strings.

```json
{
  "version": 1,
  "source_chain_id": "SOURCE_CHAIN_ID",
  "target_chain_id": "DISTINCT_TARGET_CHAIN_ID",
  "halt_height": 12345,
  "source_app_hash": "BASE64_32_BYTE_HASH",
  "source_genesis_sha256": "BASE64_SHA256_OF_EXACT_RAW_EXPORT",
  "transform_id": "reviewed-transform-id",
  "mappings": [
    {
      "old_operator": "truerepublic1...",
      "new_operator": "truerepublic1...",
      "pub_key_type": "secp256k1",
      "pub_key": "BASE64_FRESH_PUBLIC_KEY",
      "signature": "BASE64_FRESH_KEY_SIGNATURE"
    }
  ]
}
```

After the halted export is written, compute SHA-256 over its exact raw bytes
and place the 32-byte value in `source_genesis_sha256`. All replacement
operators then sign the bytes returned by `migration.SigningBytes` for the
complete descriptor. Signatures are excluded from those bytes, so every
participant signs one identical payload. Use reviewed offline signing tooling
or hardware; do not move a production private key into an ad-hoc script. The
final transform independently verifies every proof, the exact raw export
digest, and the complete state-derived consensus-key inventory.

## Halt and export

1. Stop transaction submission and announce the fixed halt height.
2. After the halt block is observed, stop and isolate enough source validators
   to prevent another commit. Do not submit validator-set changes; preserve the
   state and app hash observed at the migration boundary.
3. Require trusted RPC observers to agree on the exact height and app hash:

   ```bash
   curl --fail --silent \
     "http://TRUSTED_RPC_A:26657/block?height=${HALT_HEIGHT}" \
     | jq -r '.result.block.header.app_hash'

   curl --fail --silent \
     "http://TRUSTED_RPC_B:26657/block?height=${HALT_HEIGHT}" \
     | jq -r '.result.block.header.app_hash'
   ```

4. Stop and isolate every remaining source validator. Confirm that no source
   process can restart automatically.
5. Export from a halted source home:

   ```bash
   set -euo pipefail
   umask 077

   output=/offline/artifacts/source-export.json
   test ! -e "$output"
   tmp="$(mktemp "${output}.tmp.XXXXXX")"
   trap 'rm -f "$tmp"' EXIT

   truerepublicd export --home /isolated/source-home >"$tmp"
   mv -T --no-clobber "$tmp" "$output"
   test ! -e "$tmp"
   trap - EXIT
   ```

   The temporary file is created in the destination directory so the final
   no-clobber rename is atomic. If a destination appears concurrently, the
   remaining temporary path makes the step fail and the trap removes it. A
   failed export leaves no destination artifact, and the exact successful byte
   stream becomes the file hashed and signed below.

6. Verify that `initial_height` equals `HALT_HEIGHT + 1`, `chain_id` equals the
   descriptor source chain ID, and the independently observed app hash equals
   the descriptor hash. The export's embedded `app_hash` must be empty. A
   running-chain Cosmos SDK export may also omit `consensus.validators`; in that
   canonical form, the validated truedemocracy state remains the source of the
   target's InitChain validator updates.
7. Compute SHA-256 of the exact export file, place its raw 32-byte value in
   `source_genesis_sha256`, obtain every fresh-key signature over the completed
   descriptor, and independently verify those signatures. Do not reformat or
   edit the export after this point.
8. Record checksums of the source binary, export, completed descriptor, and
   preserved source-home recovery artifacts.

The source app hash cannot be recomputed from genesis JSON. The CLI deliberately
requires the independently observed trusted header hash. Separately, it hashes
the exact export bytes it reads and rejects any mismatch with the signed
`source_genesis_sha256` commitment.

## Transform

Run the current reviewed binary in an offline workspace:

```bash
truerepublicd migration legacy-authority \
  --descriptor /offline/artifacts/descriptor.json \
  --genesis /offline/artifacts/source-export.json \
  --output /offline/artifacts/target-genesis.json \
  --source-app-hash "$TRUSTED_SOURCE_APP_HASH"
```

The output path must not exist. The command creates one private `0600` file
atomically and does not print descriptor proof material or genesis contents.
Any validation or write error leaves no target output.

After success:

```bash
sha256sum /offline/artifacts/target-genesis.json
jq -r '.chain_id, .initial_height' \
  /offline/artifacts/target-genesis.json
```

Reviewers must compare the target checksum and descriptor transform ID. Do not
edit the generated file.

## Target start

1. Keep all source homes stopped and isolated.
2. Initialize clean target homes with fresh P2P identities and the target chain
   ID. Install each approved replacement operator through the transformed
   genesis.
3. Transfer existing consensus-key custody only through the validator identity
   procedure. The distinct target chain uses its own signer state beginning at
   the target initial height; preserve the source signer state unchanged for
   rollback, never merge state between chain IDs, and never run a source and
   target copy concurrently.
4. Install the identical transformed genesis on every target node.
5. Start the target validator set in the agreed window.
6. Require:

   - progress beyond the transformed `initial_height`;
   - one common app hash at the same height;
   - every expected consensus public key with positive power;
   - the target chain ID on every RPC;
   - no legacy operator address in exported application state; and
   - a ledger-valid target export that reimports into a clean node.

Keep the rollback window open until these checks and an independent review pass.

## Rollback

Rollback is a chain-level choice, not a merge of source and target state:

1. Stop every target validator and prove no automatic restart remains.
2. Preserve target homes for investigation; do not copy their data or signer
   state into source homes.
3. Restart the untouched source homes with the exact recorded source binary and
   source chain ID.
4. Require source height progress, validator power, and common-height app-hash
   agreement.
5. Either remain on the source chain or repeat the entire ceremony with a new
   target chain ID and descriptor. Never resume a partially trusted target
   artifact.

Source and target chain IDs differ, so account transactions and consensus
messages cannot be treated as belonging to the same chain. That separation does
not permit simultaneous signers; operational isolation remains mandatory.

## Fail-closed conditions

Abort before target start if any condition occurs:

- trusted RPC app hashes or halt heights disagree;
- the export height is not exactly the halt successor;
- the exact raw export SHA-256 differs from the signed descriptor;
- any descriptor proof or mapping is missing, duplicated, unsorted, stale, or
  derived from consensus authority;
- Auth account numbers collide or a mapped account is not a `BaseAccount`;
- target ownership already exists in Auth, Bank, or DEX;
- Bank supply, module custody, DEX state, or application ledger validation
  changes;
- CometBFT and application validators disagree;
- CosmWasm state is nonempty;
- an output path already exists; or
- source signer isolation cannot be proven.

## Automated evidence

The opt-in process drill builds the pinned pre-GH-56 revision, starts four real
legacy validators, halts at a common app hash, exports, transforms through the
current CLI, starts four target validators, verifies convergence and
export/reimport, stops all target signers, and restarts the untouched source
chain:

```bash
TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 \
go test . \
  -run '^TestMultiValidatorLegacyAuthorityMigrationRollback$' \
  -count=1 \
  -timeout=900s \
  -v
```

This evidence closes the bounded GH-61 fresh-genesis path only. Generic
governance-controlled upgrades, partially applied in-place migrations,
nonempty contract-state migration, independent security review, and production
rollout remain separate gates.
