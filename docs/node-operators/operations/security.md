# Security Hardening

## System Security

### User Management

Run the node as a dedicated non-root user:

```bash
# Create dedicated user
sudo useradd -r -s /bin/false truerepublic

# Set ownership
sudo chown -R truerepublic:truerepublic /home/truerepublic/.truerepublic

# Run service as this user (see systemd config in native-build.md)
```

### SSH Hardening

```bash
# Disable root login
sudo sed -i 's/PermitRootLogin yes/PermitRootLogin no/' /etc/ssh/sshd_config

# Disable password auth (use keys only)
sudo sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config

# Restart SSH
sudo systemctl restart sshd
```

### Firewall

```bash
# Allow only necessary ports
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 26656/tcp    # P2P (required)
sudo ufw allow 26657/tcp    # RPC (only if public)
sudo ufw enable
```

### Automatic Updates

```bash
# Enable unattended security updates (Ubuntu)
sudo apt install unattended-upgrades
sudo dpkg-reconfigure -plow unattended-upgrades
```

## Key Security

### Validator Identity Protection

The consensus key (`priv_validator_key.json`) and latest signer state
(`priv_validator_state.json`) form one safety unit. The signer state prevents
height/round/step regression; a key-only recovery, stale state, reset state, or
two active copies can double-sign.

1. **Restrict permissions:**
```bash
CHAIN_HOME="${CHAIN_HOME:-$HOME/.truerepublic}"
chmod 600 \
  "$CHAIN_HOME/config/priv_validator_key.json" \
  "$CHAIN_HOME/data/priv_validator_state.json"
```

2. **Custody as a pair:** capture both files only after a clean validator stop,
using an approved encrypted offline vault or encrypted removable media.

3. **Never expose:** do not include either file in Docker images, full-volume or
configuration archives, cloud backups, tickets, logs, chat, or version control.

4. **Keep identities distinct:** `node_key.json` is only the P2P identity.
Changing it does not rotate or contain a compromised consensus key.

5. **Consider remote custody:** Production use requires a separately reviewed
HSM/KMS or remote-signer design; the repository does not yet prove one.

Follow [Validator Identity Custody and Recovery](validator-identity-recovery.md)
for the single-signer failover and compromise-containment contract.

### Key Management Best Practices

| Practice | Description |
|----------|-------------|
| Offline custody | Store the consensus key and current signer state together on encrypted offline media |
| Access control | Only the node operator should have access |
| Single signer | Prove the source signer stopped before starting a recovered signer |
| Key rotation | Use the authenticated [validator key-rotation runbook](validator-key-rotation.md) only for active positive-power validators with independent operator custody |
| Monitoring | Alert on unexpected validator behavior |

## Network Security

### DDoS Protection

For validators, use the [sentry node architecture](../configuration/network-config.md):
- Validator node has no public-facing ports
- Sentry nodes absorb DDoS traffic
- Multiple sentries across different providers

### Rate Limiting (nginx)

```nginx
http {
    limit_req_zone $binary_remote_addr zone=rpc:10m rate=10r/s;

    server {
        location / {
            limit_req zone=rpc burst=20 nodelay;
            proxy_pass http://127.0.0.1:26657;
        }
    }
}
```

### TLS/HTTPS

Always use TLS for public-facing endpoints:

```bash
# Install certbot
sudo apt install certbot python3-certbot-nginx

# Get certificate
sudo certbot --nginx -d rpc.truerepublic.network
```

## Monitoring for Security

### Watch for

- Unexpected validator key usage (double-signing)
- Unusual peer connections
- Failed authentication attempts (SSH)
- High CPU/memory usage (possible attack)
- Unexpected process activity

### Fail2ban

```bash
sudo apt install fail2ban
sudo systemctl enable fail2ban
```

## Checklist

- [ ] Dedicated non-root user for node
- [ ] SSH key-only authentication
- [ ] Firewall configured (UFW)
- [ ] Consensus key and current signer state held together in encrypted offline custody
- [ ] Planned failover drill proves exactly one signer and a fresh P2P identity
- [ ] File permissions restricted (600 for consensus key and signer state)
- [ ] Automatic security updates enabled
- [ ] Monitoring and alerting configured
- [ ] TLS for public endpoints
- [ ] Sentry node architecture (validators)
- [ ] Regular backup schedule

## Next Steps

- [Monitoring](monitoring.md)
- [Backup & Recovery](backup-recovery.md)
- [Validator Guide](../../validators/README.md)
