# Backup & Recovery

TrueRepublic uses two deliberately separate recovery paths:

1. sanitized chain-data backup for ordinary node recovery; and
2. offline validator-identity custody for a validator failover.

Never put consensus keys, signer state, node keys, or account keyrings in a
routine archive or remote backup.

## Identity and Data Boundaries

| Item | Location | Recovery rule |
|------|----------|---------------|
| Consensus key | `config/priv_validator_key.json` | Custody together with the latest signer state |
| Consensus signer state | `data/priv_validator_state.json` | Must never regress in height/round/step |
| P2P node key | `config/node_key.json` | Generate independently for a replacement host |
| Account keyring | configured keyring backend | Back up using the keyring's own secure procedure |
| Genesis and configuration | `config/genesis.json`, `config/*.toml` | Reproduce from approved network configuration |
| Chain data | remaining `data/` contents | Sanitized backup or re-sync |

The consensus key and its latest signer state are one safety unit. A key-only
restore, stale signer state, `priv_validator_state.json` reset, or two active
copies of one consensus identity can cause double-signing. The P2P node key is
not a substitute for the consensus key and may be different on a recovery
host.

For validator custody and failover, follow
[Validator Identity Custody and Recovery](validator-identity-recovery.md).

## Sanitized Chain-Data Backup

Run the maintained script manually or from a scheduler:

```bash
./scripts/backup.sh /path/to/backup/dir

# Example: daily at 03:00
0 3 * * * CHAIN_HOME=/home/truerepublic/.truerepublic /path/to/TrueRepublic/scripts/backup.sh /path/to/backups
```

The script stops no process. Stop the service first when a crash-consistent
snapshot cannot be guaranteed by the underlying storage:

```bash
sudo systemctl stop truerepublicd
CHAIN_HOME="$HOME/.truerepublic" ./scripts/backup.sh "$HOME/truerepublic-backups"
sudo systemctl start truerepublicd
```

The archive intentionally excludes:

- `config/node_key.json`;
- `config/priv_validator_key.json`;
- `data/priv_validator_state.json`; and
- keyring directories.

Backups are retained for 30 days by default. Remote replication is acceptable
only for this sanitized artifact and must still use encrypted transport,
encrypted storage, restricted credentials, and restore testing.

## Restore Sanitized Chain Data

Initialize a fresh target home first. This creates new local node and validator
keys; the restore script preserves them while restoring only sanitized data.
This path therefore restores a full node, not an existing validator identity.

```bash
truerepublicd init restored-node \
  --chain-id truerepublic-1 \
  --home /path/to/restore-home

./scripts/restore.sh \
  /path/to/backups/truerepublic_YYYY-MM-DD.tar.gz \
  /path/to/restore-home

truerepublicd start --home /path/to/restore-home
```

Verify the genesis checksum, approved peers, chain ID, sync status, and current
app hash before using the recovered node for RPC or transaction submission.

## Docker Nodes

Do not archive the complete Docker volume: it can contain plaintext consensus
keys, signer state, and account keyrings. Stop the container and run the same
sanitized backup script against the mounted chain home. Restore into a newly
initialized volume using `scripts/restore.sh`.

## Pre-Upgrade Recovery Point

Do not create a tarball of the full chain home before an upgrade. Stop the node,
create a sanitized chain-data backup, record the running binary checksum and
height, and preserve validator identity only through the separate custody
procedure:

```bash
sudo systemctl stop truerepublicd
sha256sum /usr/local/bin/truerepublicd > truerepublic-binary.sha256
CHAIN_HOME="$HOME/.truerepublic" ./scripts/backup.sh "$HOME/truerepublic-backups"
sudo systemctl start truerepublicd
```

See [Upgrades](upgrades.md) for the tested binary-replacement boundary.
Consensus-breaking state migrations and rollback after a partially applied
migration remain unsupported.

## Fail-Closed Conditions

Do not start a recovered validator when any of these is true:

- the original signer is not proven stopped and isolated;
- the consensus key and signer state do not come from the same custody point;
- signer-state freshness or integrity cannot be established;
- the chain ID, genesis, or intended consensus public key is uncertain; or
- compromise of the consensus key is suspected.

For suspected compromise, follow the containment procedure in
[Validator Identity Custody and Recovery](validator-identity-recovery.md). A
backup of a compromised key is also compromised.

## Next Steps

- [Validator Identity Custody and Recovery](validator-identity-recovery.md)
- [Upgrades](upgrades.md)
- [Security Hardening](security.md)
