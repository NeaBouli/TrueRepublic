# Validator Consensus-Key Rotation

Status: protocol-supported after GH-56, but still a rollout-stage procedure.
Use temporary or approved testnet keys only until the rollout checklist grants
an explicit production go/no-go decision.

## Security Model

The validator operator account and the CometBFT Ed25519 consensus key are
separate authorities:

- the operator account authorizes an on-chain key change;
- the consensus key signs blocks only;
- the old consensus key is permanently revoked after rotation;
- stake, domains, escrow, voting power, jail state, and missed-block state do
  not move or reset.

Rotation is accepted only while the validator is active, not jailed, and has
positive application voting power. An inactive validator must use coordinated
recovery; submitting a replacement key cannot reactivate it.

`truerepublicd init` requires `--bootstrap-operator`. Supply an address whose
private key is held independently from `priv_validator_key.json`. Init records
the public account identity in auth genesis, but never generates, imports, or
stores its private key. Fund the operator account in the approved genesis or by
an authorized transfer before it must pay transaction fees.

Never use the account address derived from the consensus public key. Init
rejects that coupled configuration.

## Breaking Genesis Boundary

GH-56 changes validator identity at genesis. A home or exported state created
before GH-56 can retain a consensus-derived operator and does not become safely
rotatable merely by replacing the binary. TrueRepublic has no registered state
migration for rewriting that authority. Because the project is prelaunch, the
supported path is a reviewed fresh genesis with independent operator accounts;
do not reuse a legacy data volume or claim an in-place upgrade. A future
governance-controlled migration must authenticate every replacement operator
explicitly and is tracked as a separate rollout gate.

## Activation Boundary

If the rotation transaction is delivered at height `H`, the application emits
both updates in FinalizeBlock `H`:

- old key: power `0`;
- new key: the validator's unchanged power.

CometBFT v0.38 applies those updates at `H+2`. The event records this as
`scheduled_activation_height`; activation remains conditional on the validator
staying eligible through FinalizeBlock `H`. The old signer may therefore be
selected through `H+1`. TrueRepublic records the rotation as pending until the
end of `H+2`, keeps old-key evidence attributable to the same operator during
that interval, and rejects another rotation for that operator or either key.
The pending record and permanent revocation survive export/import.

## Planned Rotation

1. Confirm the network retains more than two-thirds active voting power after
   the old signer stops.
2. Confirm the application validator is active, not jailed, and has positive
   power. Stop and escalate if it is already inactive.
3. Generate the replacement consensus key in its final protected home. Do not
   copy it into tickets, chat, shell history, source control, or ordinary
   backups.
4. Configure the replacement as a full node with the approved genesis and
   peers, start it without validator power, and wait until it is synchronized.
5. Record the old and new public keys, operator address, expected validator
   power, chain ID, and current height in the private change record.
6. Prove the replacement private key corresponds to the reviewed new public
   key. The transaction authenticates the operator; operational custody must
   still prevent a typo or unavailable replacement key.
7. Gracefully stop the old signer. Disable container, systemd, supervisor, and
   host-level automatic restarts. Keep its signer-state file unchanged as
   evidence.
8. From the independently controlled operator account, broadcast:

   ```bash
   truerepublicd tx truedemocracy rotate-validator-key \
     EXPECTED_OLD_PUBKEY_HEX NEW_PUBKEY_HEX \
     --from OPERATOR_KEY \
     --chain-id CHAIN_ID \
     --node TRUSTED_RPC \
     --fees APPROVED_FEE \
     --keyring-backend APPROVED_BACKEND
   ```

9. Record the delivered height `H`; a broadcast response alone is not proof of
   successful delivery.
10. At block results for `H`, verify the old power-zero and new unchanged-power
   updates appear together. At validator set `H+1`, expect the old key to remain
   active. At `H+2`, require the old key absent and the new key active.
11. Require a commit signature from the new validator at `H+2` or later and a
    monotonically advancing replacement signer-state file. Keep the old process
    stopped and verify its signer state did not change.
12. Verify common height and app hash across the surviving validators and the
    replacement. Export state and confirm the validator's operator, stake,
    domains, and power are unchanged and the old key is listed as revoked.

Do not restart the old signer to test revocation. A transaction attempting to
rotate back to the old key must fail while the network remains live.

## Failure and Compromise Decisions

| Suspected compromise | Required response |
|---|---|
| Consensus key only; operator safe | Isolate the old signer, preserve evidence, rotate from the safe operator account, and treat the old key as active until `H+2`. |
| Operator authority | Halt the rotation procedure. Coordinate a chain halt and governance/manual recovery; the compromised operator can authorize arbitrary future keys. |
| Both authorities or custody uncertain | Halt and isolate. Do not automate replacement or start another signer until a coordinated recovery decision is approved. |
| Transaction rejected before delivery | Correct the request while the old signer stays isolated or deliberately restore service only under the approved incident plan. |
| Transaction delivered but replacement unavailable | Do not reuse the revoked old key. Preserve quorum where possible and use coordinated emergency recovery; a one-validator network can halt. |

## Export During the Pending Window

An export between `H` and the completion of `H+2` contains the new active key,
the old permanent revocation, and the pending activation record. Preserve all
three. Deleting the pending or revoked records can break attribution or allow a
retired key to return and is an invalid genesis edit.

## Automated Evidence

The gated process test uses temporary keys and proves the full transition:

```bash
TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 \
  go test . -run '^TestMultiValidatorConsensusKeyRotation$' \
  -count=1 -timeout=420s -v
```

This procedure does not replace HSM/remote-signer integration, governance
emergency recovery, custody review, or the final production rollout decision.
The application currently does not feed CometBFT ABCI++ misbehavior and
last-commit data into its economic slashing handlers. Revocation and historical
ownership are consensus state, but automatic equivocation/downtime penalties
remain a separate rollout blocker; do not claim operational slashing coverage.
