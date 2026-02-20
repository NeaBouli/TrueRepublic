# Troubleshooting

Common issues and solutions for TrueRepublic nodes.

## Quick Diagnosis

```bash
# Check if node is running
curl localhost:26657/status

# Check sync status
curl localhost:26657/status | jq .result.sync_info.catching_up

# Check peers
curl localhost:26657/net_info | jq .result.n_peers

# Check logs
docker-compose logs --tail=100 truerepublic-node
# or
sudo journalctl -u truerepublicd --tail=100
```

---

## Node Won't Start

### Error: "address already in use"

**Cause:** Port conflict

**Solution:**

```bash
# Find process using port
sudo lsof -i :26656
sudo lsof -i :26657

# Kill process
sudo kill -9 <PID>

# Or change ports in config.toml
```

### Error: "failed to load genesis"

**Cause:** Missing or corrupt genesis.json

**Solution:**

```bash
# Download genesis
cd ~/.truerepublic/config
wget https://raw.githubusercontent.com/NeaBouli/TrueRepublic/main/genesis.json

# Verify checksum
sha256sum genesis.json
```

### Error: "out of memory"

**Cause:** Insufficient RAM

**Solution:**

```bash
# Reduce cache size in app.toml
[state-sync]
snapshot-interval = 500  # Lower value

# Or add swap
sudo fallocate -l 8G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
```

---

## Node Not Syncing

### Symptom: catching_up stays true

**Check 1: Peer count**

```bash
curl localhost:26657/net_info | jq .result.n_peers
```

**If 0 peers:**
1. Check firewall: `sudo ufw status`
2. Check seeds in config.toml
3. Check external_address setting

**Check 2: Block height not increasing**

```bash
# Run twice, 30 seconds apart
curl localhost:26657/status | jq .result.sync_info.latest_block_height
```

**If stuck:**

```bash
# Restart node
docker-compose restart truerepublic-node
# or
sudo systemctl restart truerepublicd
```

**Check 3: Disk full**

```bash
df -h
```

**If >90% full:**

```bash
# Enable pruning in app.toml
pruning = "default"

# Or clean old data
truerepublicd unsafe-reset-all
```

---

## High Missed Blocks

### Cause 1: Node offline

**Solution:**

```bash
# Check if running
sudo systemctl status truerepublicd

# Start if stopped
sudo systemctl start truerepublicd
```

### Cause 2: Network issues

**Solution:**

```bash
# Check peers
curl localhost:26657/net_info

# If low peers, add seeds
nano ~/.truerepublic/config/config.toml

# Under [p2p]:
seeds = "seed1@ip:port,seed2@ip:port"
```

### Cause 3: High load

**Solution:**

```bash
# Check CPU/RAM
top

# Upgrade server or optimize config
# Reduce cache, connections, etc.
```

---

## Validator Jailed

### Check jail status

```bash
truerepublicd query truedemocracy validator <your-address>
```

### Unjail process

```bash
# Wait for jail period to expire
# Then unjail:
truerepublicd tx truedemocracy unjail \
    --from validator \
    --chain-id truerepublic-1
```

See [Validator Guide - Unjailing](Validator-Guide#unjailing) for details.

---

## Cannot Query API

### Error: "connection refused"

**Cause:** API not enabled

**Solution:**

```bash
nano ~/.truerepublic/config/app.toml

# Enable API:
[api]
enable = true
address = "tcp://0.0.0.0:1317"

# Restart
sudo systemctl restart truerepublicd
```

### Error: "unauthorized"

**Cause:** API requires authentication

**Solution:**

```bash
# If behind nginx, check nginx config
# Or disable auth in app.toml
```

---

## Disk Space Issues

### Check usage

```bash
df -h
du -sh ~/.truerepublic/data
```

### Solutions

**1. Enable pruning:**

```bash
nano ~/.truerepublic/config/app.toml

pruning = "default"
pruning-keep-recent = "100"
pruning-interval = "10"
```

**2. Clean old data:**

```bash
# DANGER: Deletes all data, resync required
truerepublicd unsafe-reset-all
```

**3. Move to larger disk:**

```bash
# Stop node
sudo systemctl stop truerepublicd

# Move data
sudo mv ~/.truerepublic /mnt/large-disk/

# Symlink
ln -s /mnt/large-disk/.truerepublic ~/.truerepublic

# Start
sudo systemctl start truerepublicd
```

---

## Performance Issues

### Symptom: Slow block processing

**Check 1: System resources**

```bash
top
iostat
```

**Check 2: Peer quality**

```bash
# Check peer latency
curl localhost:26657/net_info | jq '.result.peers[] | {ip: .remote_ip, latency: .connection_status.duration}'
```

**Solutions:**
1. Upgrade hardware
2. Use faster SSD (NVMe)
3. Reduce mempool size
4. Limit peer connections

---

## Docker Issues

### Container won't start

```bash
# Check logs
docker-compose logs truerepublic-node

# Check if port conflict
docker ps
```

### Container keeps restarting

```bash
# Check exit code
docker ps -a

# View full logs
docker logs truerepublic-node
```

### Out of disk space

```bash
# Check Docker disk usage
docker system df

# Clean up
docker system prune -a
```

---

## Getting Help

**Before asking for help, gather:**

1. Node status: `curl localhost:26657/status`
2. Logs (last 100 lines)
3. System info: `uname -a`, `docker version`
4. Error messages

**Where to ask:**
- GitHub Issues: Bug reports
- GitHub Discussions: Questions
- Telegram: https://t.me/truerepublic

---

## Next Steps

- [Node Setup](Node-Setup) -- Setup guide
- [Validator Guide](Validator-Guide) -- Validator operations
- [Monitoring](Monitoring) -- Set up monitoring
