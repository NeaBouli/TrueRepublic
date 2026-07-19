# Backup & Recovery

## What to Back Up

| Item | Location | Priority |
|------|----------|----------|
| Validator key | `~/.truerepublic/config/priv_validator_key.json` | **Critical** |
| Node key | `~/.truerepublic/config/node_key.json` | High |
| Genesis file | `~/.truerepublic/config/genesis.json` | High |
| Configuration | `~/.truerepublic/config/*.toml` | Medium |
| Chain data | `~/.truerepublic/data/` | Low (can re-sync) |

> **Warning:** The validator key (`priv_validator_key.json`) is your most critical file. If lost, you lose your validator identity. If compromised, an attacker can double-sign and get your stake slashed. Store it securely.

## Automated Backup

### Using the Backup Script

```bash
# Run manually
./scripts/backup.sh /path/to/backup/dir

# Schedule daily at 3 AM via cron
crontab -e
0 3 * * * /path/to/TrueRepublic/scripts/backup.sh /path/to/backups
```

Backups are retained for 30 days. Old backups are automatically cleaned up.
The script creates a sanitized chain-data artifact: it intentionally excludes
`config/node_key.json`, `config/priv_validator_key.json`,
`data/priv_validator_state.json`, and keyring directories. Store validator and
node keys through the separate offline key-backup procedure, not in routine
chain-data archives.

### Restore a Sanitized Backup

Initialize the target home first, then restore the sanitized data over it. This
preserves the target's local node and validator keys.

```bash
truerepublicd init restored-node --chain-id truerepublic-1 --home /path/to/restore-home
./scripts/restore.sh /path/to/backups/truerepublic_YYYY-MM-DD.tar.gz /path/to/restore-home
truerepublicd start --home /path/to/restore-home
```

### Docker Volume Backup

```bash
# Stop the node first for consistent backup
docker compose stop truerepublic-node

# Backup volume
docker run --rm \
    -v truerepublic_node-data:/data:ro \
    -v $(pwd)/backups:/backup \
    alpine tar czf /backup/node-$(date +%Y%m%d).tar.gz /data

# Restart
docker compose start truerepublic-node
```

## Manual Backup

### Chain Data Backup

```bash
# Stop the node for consistency
sudo systemctl stop truerepublicd

# Create sanitized chain-data backup
CHAIN_HOME="$HOME/.truerepublic" ./scripts/backup.sh "$HOME/truerepublic-backups"

# Restart
sudo systemctl start truerepublicd
```

### Configuration Only (No Chain Data)

```bash
tar -czf truerepublic_config_$(date +%Y%m%d).tar.gz ~/.truerepublic/config
```

### Validator Key Only

```bash
cp ~/.truerepublic/config/priv_validator_key.json ~/validator_key_backup.json
# Store this file OFFLINE in a secure location
```

## Recovery

### From Sanitized Backup

```bash
# Stop the node
sudo systemctl stop truerepublicd

# Initialize a fresh target home first
truerepublicd init restored-node --chain-id truerepublic-1 --home ~/.truerepublic-restored

# Restore sanitized chain data while preserving local keys
./scripts/restore.sh ~/truerepublic-backups/truerepublic_YYYY-MM-DD.tar.gz ~/.truerepublic-restored

# Start
truerepublicd start --home ~/.truerepublic-restored
```

### From Docker Volume Backup

```bash
make docker-down
docker volume rm truerepublic_node-data
docker volume create truerepublic_node-data

docker run --rm \
    -v truerepublic_node-data:/data \
    -v $(pwd)/backups:/backup \
    alpine tar xzf /backup/node-YYYYMMDD.tar.gz -C /

make docker-up
```

### From Configuration Only (Re-sync Chain)

```bash
# Restore config
tar -xzf truerepublic_config_YYYYMMDD.tar.gz -C ~/

# Start - node will sync from peers
sudo systemctl start truerepublicd
```

### Validator Key Recovery

If you only have the validator key backup:

1. Initialize a fresh node
2. Copy the validator key:
```bash
cp validator_key_backup.json ~/.truerepublic/config/priv_validator_key.json
```
3. Ensure genesis.json matches the network
4. Start and let it sync

## Pre-Upgrade Backup

Always back up before chain upgrades:

```bash
# Tag the backup with the current version
sudo systemctl stop truerepublicd
tar -czf truerepublic_pre-upgrade_v$(cat VERSION).tar.gz ~/.truerepublic
sudo systemctl start truerepublicd
```

## Remote Backup

Configure `rclone` in `scripts/backup.sh` for remote storage:

```bash
# Install rclone
curl https://rclone.org/install.sh | sudo bash

# Configure a remote (S3, GCS, etc.)
rclone config

# Test upload
rclone copy backups/ remote:truerepublic-backups/
```

## Next Steps

- [Upgrades](upgrades.md)
- [Security Hardening](security.md)
