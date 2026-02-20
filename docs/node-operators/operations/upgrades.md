# Upgrades

## Upgrade Process

### Step 1: Back Up

Always back up before upgrading:

```bash
sudo systemctl stop truerepublicd
tar -czf truerepublic_pre-upgrade.tar.gz ~/.truerepublic
```

### Step 2: Build New Version

```bash
cd TrueRepublic
git fetch origin
git checkout <new-version-tag>
make build
```

### Step 3: Replace Binary

```bash
# Native
sudo cp ./build/truerepublicd /usr/local/bin/truerepublicd

# Docker
make docker-build
```

### Step 4: Start

```bash
# Native
sudo systemctl start truerepublicd

# Docker
make docker-up
```

### Step 5: Verify

```bash
curl http://localhost:26657/status | jq .result.node_info.version
```

## Docker Upgrades

```bash
cd TrueRepublic
git pull origin main
make docker-down
make docker-build
make docker-up
```

## Rollback

If an upgrade fails:

```bash
# Stop the node
sudo systemctl stop truerepublicd

# Restore from backup
rm -rf ~/.truerepublic
tar -xzf truerepublic_pre-upgrade.tar.gz -C ~/

# Restore old binary
git checkout <previous-version-tag>
make build
sudo cp ./build/truerepublicd /usr/local/bin/truerepublicd

# Start
sudo systemctl start truerepublicd
```

## Chain Upgrades (Breaking Changes)

For upgrades that change the state machine (consensus-breaking):

1. All validators must upgrade at the coordinated block height
2. The upgrade proposal will specify the exact height
3. Nodes that don't upgrade will halt at the upgrade height
4. After upgrading, the chain resumes automatically

## Next Steps

- [Backup & Recovery](backup-recovery.md)
- [Security Hardening](security.md)
