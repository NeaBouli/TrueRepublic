# Validator Identity Custody and Recovery

Status: cold-failover procedure for moving the same coupled key and signer
state. For replacing a consensus key, use the separate
[validator key-rotation procedure](validator-key-rotation.md). Neither path
authorizes production keys or real funds before rollout approval.

## The Safety Unit

A TrueRepublic validator has three different local identities:

| Artifact | Purpose | Recovery treatment |
|----------|---------|--------------------|
| `config/priv_validator_key.json` | Ed25519 consensus signing identity | Custody offline with signer state |
| `data/priv_validator_state.json` | Last signed height/round/step and signature safety record | Capture only after a clean stop; never regress or reset |
| `config/node_key.json` | CometBFT P2P identity | Generate a new one on the recovery host |

The first two files are a coupled consensus-safety unit. File mode must be
`0600`, ownership must belong to the dedicated service account, and plaintext
copies must not enter ordinary archives, cloud storage, source control, tickets,
logs, chat, or shell history.

Use an organization-approved encrypted offline vault, removable encrypted
media, or remote signer/HSM custody process. TrueRepublic deliberately does not
ship a script that emits a plaintext validator-identity archive.

## Prepare a Cold Recovery Host

1. Provision and harden a fresh host with the approved TrueRepublic binary.
2. Initialize a new home with the correct chain ID and approved genesis.
3. Keep the newly generated `config/node_key.json`; confirm its node ID differs
   from the active validator.
4. Configure trusted peers, firewall rules, monitoring, and time
   synchronization.
5. Keep the replacement stopped until the transfer preconditions below pass.

## Planned Failover

1. Announce the maintenance window and verify the network retains quorum.
2. Gracefully stop the active validator and prove the process cannot restart
   through systemd, containers, supervisors, or an old host.
3. After the stop completes, capture the consensus key and the now-current
   signer state as one encrypted custody transaction. Record their checksums,
   chain ID, consensus public key, source host, and capture time in the private
   custody register. Never publish the checksums of secret material.
4. Transfer both files to the fresh home using an authenticated encrypted
   channel. Preserve the fresh host's node key.
5. Assign both files to the dedicated service account, set owner-only
   permissions, and verify ownership before startup:

   ```bash
   CHAIN_HOME="${CHAIN_HOME:-$HOME/.truerepublic}"
   sudo chown truerepublic:truerepublic \
     "$CHAIN_HOME/config/priv_validator_key.json" \
     "$CHAIN_HOME/data/priv_validator_state.json"
   chmod 600 \
     "$CHAIN_HOME/config/priv_validator_key.json" \
     "$CHAIN_HOME/data/priv_validator_state.json"
   stat -c '%U:%G %a %n' \
     "$CHAIN_HOME/config/priv_validator_key.json" \
     "$CHAIN_HOME/data/priv_validator_state.json"
   ```

   Both records must report `truerepublic:truerepublic 600`. Use the actual
   dedicated service account when the deployment uses a different name.

6. Before start, verify all of the following:
   - the old signer remains stopped and isolated;
   - the recovered consensus public key equals the registered validator key;
   - the signer height/round/step equals the custody record and is not older
     than any known signing activity;
   - the P2P node ID is new;
   - the chain ID and genesis checksum are approved.
7. Start exactly one recovered signer. Monitor missed blocks, double-sign
   evidence, consensus public key and power, peer identity, height, and app hash.
8. Keep the old host disabled. Do not use it as a hot standby.

If any check fails, do not start the validator. Re-establish provenance and
freshness or escalate to coordinated network recovery.

## Suspected or Confirmed Compromise

1. Stop and isolate the signer immediately. Disable every automatic restart.
2. Preserve volatile logs, host and access evidence, relevant heights, and
   custody audit records without copying secrets into the incident ticket.
3. Treat both the active key and every backup of it as unusable. Rotating only
   `node_key.json` does not change consensus authority.
4. Notify network operators and coordinate power-zero eviction or a chain halt
   before any recovery signer starts.
5. Do not start a second copy of the compromised consensus key. If the old
   identity cannot be safely evicted, remain stopped and use a coordinated
   manual recovery decision.

TrueRepublic provides a separate authenticated atomic rotation flow with
permanent old-key revocation. Do not approximate it with `remove-validator`:
that flow withdraws stake and is subject to the domain transfer limit. Cold
failover remains the correct procedure only when the consensus identity itself
must remain unchanged.

## Automated Evidence

The opt-in recovery harness proves the bounded planned-failover path with
temporary keys:

```bash
TRUEREPUBLIC_VALIDATOR_IDENTITY_RECOVERY=1 \
  go test . -run '^TestValidatorIdentityColdFailover$' \
  -count=1 -timeout=300s -v
```

It must prove the source signer stopped before transfer, the consensus public
key and power remained unchanged, the recovery P2P identity changed, signer
height/round/step advanced monotonically, validators converged on one app hash,
and exported ledger invariants remained valid.

## Explicitly Not Proven

- recovery from a compromised consensus key;
- simultaneous or automatic hot failover;
- HSM/KMS or remote-signer integration;
- governance emergency recovery after operator authority compromise.

Track these rollout boundaries in
[Issue #29](https://github.com/NeaBouli/TrueRepublic/issues/29) and the
[Road to Rollout](../../ROLLOUT_ROADMAP.md).
