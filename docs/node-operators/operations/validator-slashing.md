# Validator Slashing and Recovery

TrueRepublic consumes CometBFT ABCI++ misbehavior and decided-last-commit data
before transactions execute. Slashing, replay markers, liveness state,
validator power changes, and PNYX burns are committed atomically.

## Penalties

| Condition | Economic penalty | Consensus action | Recovery |
|---|---:|---|---|
| Duplicate vote or light-client attack | 5% of current validator stake, burned | Immediate jail and power-zero update; offending consensus key is tombstoned | Isolate the signer, rotate to a fresh consensus key, wait for the jail period, then submit `unjail` |
| More than 50 missed commits in the complete rolling 100-block window | 1% of current validator stake, burned | Jail for 600 seconds and power-zero update | Correct availability, wait for the jail period, then submit `unjail` |

A nil precommit counts as signed. Only an absent vote counts as missed. The
downtime threshold is not evaluated until a complete 100-block observation
window exists.

## Delayed evidence and validator exits

Consensus-key ownership is retained with exact activation and retirement
heights. Evidence for a rotated key is charged to the same operator when the
key was active at the offense height.

A full validator removal does not immediately return stake. The complete stake
claim remains in module escrow until both configured CometBFT evidence limits
have been strictly exceeded:

- block age since the key's final possible signing height; and
- elapsed time since consensus retirement was observed.

This hold also applies to a full-balance `withdraw-stake`, so exit cannot bypass
delayed slashing. Valid evidence delivered during the hold burns the penalty
from the pending claim before any payout.

Partial validator withdrawals are currently rejected. They will remain
disabled until a generalized slashable-unbonding record can retain every
reduced claim for the evidence window. Operators must use a full validator exit
instead; this is a deliberate fail-closed restriction.

## Incident procedure

1. Isolate the suspected signer. Never copy `priv_validator_state.json` to a
   second active signer.
2. Preserve node, peer, consensus, transaction, and host logs. Record the
   offense height, evidence hash, validator address, operator address, current
   power, stake, and app hash.
3. Query application and consensus state:

   ```bash
   truerepublicd query truedemocracy validator <operator-address>
   curl -s http://127.0.0.1:26657/validators
   curl -s http://127.0.0.1:26657/status
   ```

4. For equivocation, assume the old consensus key is compromised. Follow
   [Validator Consensus-Key Rotation](validator-key-rotation.md) with a freshly
   generated offline key. A tombstoned key cannot be unjailed.
5. For downtime, repair networking, disk, clock, process supervision, and
   sentry connectivity before requesting unjail:

   ```bash
   truerepublicd tx truedemocracy unjail \
     --from <operator-key> \
     --chain-id <chain-id>
   ```

6. Verify new blocks, current validator power, stake, supply, escrow parity,
   peer health, and common-height app-hash agreement before closing the
   incident.

## Replay and restart guarantees

Processed infractions, the canonical last-commit cursor, rolling liveness
bitmaps, historical consensus-key ownership, tombstones, jail state, and
pending exits are included in export/import state. Replaying an identical
commit or restarting from an exported state cannot burn stake twice.

## Automated process evidence

The gated four-process test accelerates consensus only inside temporary test
homes. It proves 100-block downtime ingestion, restart/catch-up as a full node,
real CometBFT duplicate-vote evidence broadcast, exact 1% and 5% burns, two
power-zero transitions, app-hash convergence, and ledger-valid export:

```bash
TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 \
go test . -run '^TestMultiValidatorConsensusSlashing$' \
  -count=1 -timeout=360s -v
```

This recovery evidence does not replace an independent security review or
production incident drills.
