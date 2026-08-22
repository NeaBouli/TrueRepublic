# Governed Application Upgrades and Rollback

TrueRepublic wires Cosmos SDK `x/upgrade` and supports the exact fresh-genesis
`v0.4.1` governed migration path. A genesis-only `governance` Domain snapshots
its electorate on the first matching proposal vote. At least two thirds of
that immutable electorate must submit the same plan name, height, and info;
the height must retain at least ten blocks of lead time. The same snapshot and
threshold control cancellation.

The old binary has no `v0.4.1` handler and deterministically halts consensus at
the scheduled height. Only a binary built with the exact release identity
`main.upgradePlan=v0.4.1` registers the reviewed handler, runs module migrations, and
records completion atomically. Ordinary accounts cannot invoke the stock
upgrade authority because it is the non-signing `truedemocracy` module account.

This path is supported only for chains created with the upgrade store already
present. Introducing that store into an existing pre-GH-184 chain requires a
separate reviewed store-loader release and is not covered here. IBC client
upgrades, arbitrary plan names, public deployment, and real-funds operation
remain unsupported.

Coordinate every rehearsal, abort, or rollback decision through
[Incident Command and Rehearsal](incident-command.md).

## Schedule or cancel v0.4.1

Every governance member uses its own key and submits the identical command:

```bash
truerepublicd tx truedemocracy vote-software-upgrade \
  v0.4.1 <upgrade-height> '<reviewed-info>' \
  --from <member-key> --chain-id <chain-id>
```

For four eligible genesis members, three matching votes schedule the plan.
Membership changes after the first vote do not alter that proposal's
electorate; GH-184 additionally forbids adding or excluding members from this
reserved Domain after genesis. Duplicate votes, conflicting plans, late
members, insufficient lead time, and runtime creation fail closed. A pending
plan and its schedule/cancel votes are included in application genesis export
and restore the real plan on a validated import.

Before the scheduled height, the same original electorate may cancel. The
same two-thirds path also clears an unscheduled proposal whose lead-time window
expired, preventing a stale first vote from blocking future plans:

```bash
truerepublicd tx truedemocracy vote-cancel-software-upgrade \
  v0.4.1 --from <member-key> --chain-id <chain-id>
```

After successful execution, use the same two-thirds cancellation vote as an
authenticated cleanup step before proposing the next named release. `x/upgrade`
has already cleared the executed plan; this vote removes only the retained
governance record and makes the next exact proposal available.

Install the reviewed candidate at a separate immutable path before the halt,
but do not run it early. After all operators observe the expected halt and
preserve logs, stop the diagnostic/RPC processes that CometBFT may leave alive,
then start the exact candidate on the unchanged homes. Never use
`--unsafe-skip-upgrades` as a recovery shortcut.

## Safety rules

- Treat consensus-engine patch releases as coordinated network upgrades unless
  mixed-version compatibility is explicitly proven. In particular, CometBFT
  v0.38.26 increases the maximum signature and commit-size constants for its
  ML-DSA verification backport. A v0.38.25 node and a v0.38.26 node can
  therefore calculate different maximum block-data sizes. Do not use the
  compatible rolling-replacement procedure below across that version boundary;
  halt and restart every validator on the reviewed v0.38.26 artifact at the
  agreed height.
- Before scheduling that restart, record the effective `block.max_bytes` and
  `evidence.max_bytes` consensus parameters plus the largest validator-set size
  permitted during the upgrade window. Calculate the v0.38.26 worst-case data
  envelope as `max_bytes - 731 - (3358 * validators) - evidence_bytes`. The
  constants are the v0.38.26 maximum block/header/commit framing plus its
  3,309-byte maximum signature. The result must be non-negative. If it is
  negative, do not restart: first enact a separately reviewed coordinated
  consensus-parameter change or approve and enforce a smaller validator-set
  bound. Preserve the parameter height, inputs, result, and bound approval in
  the upgrade evidence; the current live count alone is not a sufficient bound.
- Rehearse the exact old and new artifacts on a private network first.
- Record commit IDs, versions, and SHA-256 checksums for both binaries.
- Keep the last known-good binary available until the recovery window closes.
- Replace one validator binary at a time while the remaining validators retain
  quorum. Stop if app hashes, validator power, or catch-up behavior diverge.
- Never copy a validator key between operators.
- Never restore an older `data/priv_validator_state.json` after that validator
  may have signed a newer height. Regressing signing state can enable a
  conflicting signature at an already signed height.
- Routine chain-data archives must remain sanitized. `scripts/backup.sh`
  intentionally excludes node keys, validator keys, signing state, and
  keyrings; manage identity keys through a separate offline procedure.

## Preflight

For the v0.38.25 to v0.38.26 boundary, obtain the effective parameters from at
least two trusted RPCs at the agreed pre-upgrade height. Set
`MAX_VALIDATORS_DURING_UPGRADE` to the independently reviewed governance or
operator bound, never merely the current validator count, and fail closed on a
negative result:

```bash
export PARAMS_HEIGHT="12345" # replace with the agreed pre-upgrade height
export MAX_VALIDATORS_DURING_UPGRADE="100" # replace with the reviewed bound

params="$(curl --fail --silent \
  "http://trusted-rpc-a:26657/consensus_params?height=$PARAMS_HEIGHT")"
max_bytes="$(printf '%s' "$params" | jq -er \
  '.result.consensus_params.block.max_bytes | tonumber')"
evidence_bytes="$(printf '%s' "$params" | jq -er \
  '.result.consensus_params.evidence.max_bytes | tonumber')"
max_data_bytes="$((max_bytes - 731 - \
  (3358 * MAX_VALIDATORS_DURING_UPGRADE) - evidence_bytes))"

printf 'height=%s max_bytes=%s evidence_bytes=%s validator_bound=%s max_data_bytes=%s\n' \
  "$PARAMS_HEIGHT" "$max_bytes" "$evidence_bytes" \
  "$MAX_VALIDATORS_DURING_UPGRADE" "$max_data_bytes"
test "$max_data_bytes" -ge 0
```

Repeat against the second trusted RPC and require identical parameter values.
Archive both outputs and the reviewed source of the validator-set bound. A
failed `test` is a hard stop, not permission to reduce evidence capacity or the
validator set ad hoc.

Choose immutable paths instead of overwriting the running binary in place:

```bash
export CHAIN_HOME="$HOME/.truerepublic"
export OLD_BINARY="/opt/truerepublic/v0.4.0/truerepublicd"
export NEW_BINARY="/opt/truerepublic/v0.4.1/truerepublicd"
export CHECKPOINT_HEIGHT="12345" # replace with the coordinated checkpoint
export TRUSTED_RPCS="http://trusted-rpc-a:26657 http://trusted-rpc-b:26657"

"$OLD_BINARY" version
"$NEW_BINARY" version
sha256sum "$OLD_BINARY" "$NEW_BINARY"

reference_app_hash=""
for rpc in $TRUSTED_RPCS; do
  latest_height="$(curl --fail --silent "$rpc/status" | jq -er \
    '.result.sync_info.latest_block_height | tonumber')"
  if [ "$latest_height" -lt "$CHECKPOINT_HEIGHT" ]; then
    echo "$rpc is below checkpoint height $CHECKPOINT_HEIGHT" >&2
    exit 1
  fi
  app_hash="$(curl --fail --silent \
    "$rpc/block?height=$CHECKPOINT_HEIGHT" | jq -er \
    '.result.block.header.app_hash')"
  printf '%s height=%s checkpoint_app_hash=%s\n' \
    "$rpc" "$latest_height" "$app_hash"
  if [ -z "$reference_app_hash" ]; then
    reference_app_hash="$app_hash"
  elif [ "$app_hash" != "$reference_app_hash" ]; then
    echo "trusted RPC app hashes disagree at height $CHECKPOINT_HEIGHT" >&2
    exit 1
  fi
done
```

Replace the example checkpoint height and RPC URLs before executing the block.
It fails if either trusted endpoint is below the checkpoint or their app hashes
disagree. Record the matching heights and hash, then stop the local service
cleanly and create a sanitized chain-data backup:

```bash
sudo systemctl stop truerepublicd
CHAIN_HOME="$CHAIN_HOME" ./scripts/backup.sh "$HOME/truerepublic-backups"
sha256sum \
  "$CHAIN_HOME/config/node_key.json" \
  "$CHAIN_HOME/config/priv_validator_key.json" \
  "$CHAIN_HOME/data/priv_validator_state.json"
```

The signing-state checksum is an audit record, not a file to restore later.
Once the validator signs again, its signing state must move forward.

## Compatible rolling replacement

Point the service at the separately installed candidate binary, then start the
single validator:

```bash
sudo systemctl edit truerepublicd
```

Enter this drop-in, clearing the original command before replacing it:

```ini
[Service]
ExecStart=
ExecStart=/opt/truerepublic/v0.4.1/truerepublicd start --home /home/<operator>/.truerepublic
```

Then reload and start the service:

```bash
sudo systemctl daemon-reload
sudo systemctl start truerepublicd
```

Before proceeding to another validator, verify:

```bash
curl --fail --silent http://127.0.0.1:26657/status | jq \
  '.result.sync_info | {latest_block_height, catching_up}'
curl --fail --silent 'http://127.0.0.1:26657/block?height=<checkpoint-height>' | jq -r \
  '.result.block.header.app_hash'
curl --fail --silent http://127.0.0.1:26657/validators | jq '.result.validators'
```

The checkpoint app hash must match the preflight record, `catching_up` must
eventually be false, new blocks must continue, and the expected validator set
and power must remain visible. Confirm that the node and validator key
checksums are unchanged. The signing-state height may advance but must never
decrease.

## Failed candidate and binary rollback

If the candidate exits before it opens state or fails its readiness checks:

1. Stop the service and preserve its logs.
2. Do **not** delete the validator home.
3. Do **not** restore an old validator signing-state file.
4. Repoint `ExecStart` to the recorded last known-good binary.
5. Start the service and require catch-up, checkpoint app-hash equality,
   validator-power visibility, and new block production.

```bash
sudo systemctl stop truerepublicd
sudo systemctl edit truerepublicd
```

Enter the last known-good command as the complete override:

```ini
[Service]
ExecStart=
ExecStart=/opt/truerepublic/v0.4.0/truerepublicd start --home /home/<operator>/.truerepublic
```

Then reload and start the service:

```bash
sudo systemctl daemon-reload
sudo systemctl start truerepublicd
```

If the candidate may have migrated or mutated the database, this simple binary
rollback is not safe. Keep the validator stopped and escalate to a coordinated
chain recovery decision. Do not improvise by combining historical chain data
with a regressed signer state.

## Chain-data recovery boundary

A sanitized archive may be restored into a freshly initialized **non-validator
full-node home**, which can then catch up from trusted peers. Follow
[Backup & Recovery](backup-recovery.md). For a validator, preserve the
operator's current identity and monotonic signing state and require an explicit
double-sign safety review before any data restoration.

## Automated evidence

`TestMultiValidatorPersistedBinaryUpgradeRollback` exercises this bounded
procedure with four real validator processes. It commits non-empty application
state, performs compatible rolling replacements, tests a deterministic failure
before state is opened, returns every validator to the baseline binary, and
checks historical and current app hashes, validator power, unchanged identity
keys, non-regressing signing state, exported ledger invariants, and re-import.

`TestGovernedUpgradeMultiValidatorHaltFailureRecovery` adds the governed path:
four real validators reach a three-of-four vote, the old artifact halts at the
same height, an intentional partial cached write returns an error without
committing marker or height, and the fixed `v0.4.1` artifact resumes with equal
app hashes and validator power. A second candidate restart proves the migration
marker and done height execute exactly once.

This local harness completes the bounded implementation gate. A private clean-
infrastructure rehearsal, independent review, pre-GH-184 store introduction,
IBC client upgrade, and production approval remain separate rollout gates.

The bounded pre-GH-56 operator-authority transition is documented separately
in [Legacy Validator-Authority Migration](legacy-authority-migration.md). It
uses a distinct-chain reviewed fresh genesis and must not be treated as an
in-place binary upgrade.

## Next steps

- [Incident Command and Rehearsal](incident-command.md)
- [Backup & Recovery](backup-recovery.md)
- [Security Hardening](security.md)
- [Road to Rollout](../../ROLLOUT_ROADMAP.md)
