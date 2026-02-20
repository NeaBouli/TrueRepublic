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

### Validator Key Protection

The validator key (`priv_validator_key.json`) is the most sensitive file:

1. **Restrict permissions:**
```bash
chmod 600 ~/.truerepublic/config/priv_validator_key.json
```

2. **Back up offline:** Copy to an air-gapped device or encrypted USB

3. **Never expose:** Don't include in Docker images, backups uploaded to cloud, or version control

4. **Consider HSM:** For production validators, use a Hardware Security Module (e.g., YubiHSM) or Tendermint KMS

### Key Management Best Practices

| Practice | Description |
|----------|-------------|
| Offline backup | Store validator key on encrypted offline media |
| Access control | Only the node operator should have access |
| Key rotation | Plan for key rotation procedures |
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
- [ ] Validator key backed up offline
- [ ] File permissions restricted (600 for keys)
- [ ] Automatic security updates enabled
- [ ] Monitoring and alerting configured
- [ ] TLS for public endpoints
- [ ] Sentry node architecture (validators)
- [ ] Regular backup schedule

## Next Steps

- [Monitoring](monitoring.md)
- [Backup & Recovery](backup-recovery.md)
- [Validator Guide](../../validators/README.md)
