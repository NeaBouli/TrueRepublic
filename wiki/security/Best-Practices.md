# Best Practices

Security recommendations for all TrueRepublic stakeholders.

## Table of Contents

1. [For Users](#for-users)
2. [For Node Operators](#for-node-operators)
3. [For Validators](#for-validators)
4. [For Developers](#for-developers)
5. [General Security](#general-security)

---

## For Users

### Wallet Security

**DO:**

1. **Use Hardware Wallet**
   - Ledger or Trezor
   - Keys never touch computer
   - Sign transactions on device

2. **Backup Seed Phrase**
   - Write on paper (never digital)
   - Store in multiple secure locations
   - Consider fireproof/waterproof safe
   - Never take photos

3. **Verify Addresses**
   - Always double-check recipient
   - Use address book for frequent recipients
   - Be wary of address substitution malware

4. **Check Transaction Details**
   - Review amount before signing
   - Verify network (mainnet vs testnet)
   - Check gas fees

5. **Keep Software Updated**
   - Update Keplr wallet
   - Update browser
   - Update OS

**DON'T:**

1. **Share Seed Phrase**
   - Never with anyone
   - Not support staff
   - Not family
   - Not in messages

2. **Use Same Password**
   - Unique password for Keplr
   - Not reused from other sites

3. **Trust Blindly**
   - Verify URLs (https://truerepublic.app)
   - Check for phishing
   - Don't click suspicious links

4. **Use Public WiFi**
   - For wallet operations
   - Without VPN

5. **Store Keys Digitally**
   - No screenshots
   - No cloud storage
   - No email/messages

---

### Proposal Participation

**Before Rating:**

1. **Read Full Proposal**
   - Issue description
   - Suggested solution
   - External links (verify legitimacy)

2. **Check Proposal Age**
   - GREEN (0-7 days): New
   - YELLOW (7-30 days): Mature
   - RED (30+ days): Expiring

3. **Review Discussion**
   - Check domain chat/forum
   - See what others think
   - Consider all perspectives

4. **Verify External Links**
   - Check URL legitimacy
   - Don't click suspicious links
   - Verify on official channels

**Rating Guidelines:**

- **+5:** Strongly support, ready to implement now
- **+3:** Support with minor reservations
- **0:** Neutral, need more info
- **-3:** Concerns, needs changes
- **-5:** Strongly oppose, fundamentally flawed

**Don't:**
- Rate based on person, rate on merit
- Rate without reading
- Copy others' ratings
- Vote on every single proposal (causes noise)

---

### Domain Safety

**Joining Domains:**

1. **Research Domain**
   - Read domain description
   - Check member count
   - Review existing proposals
   - Verify it's legitimate

2. **Understand Rules**
   - Domain-specific guidelines
   - Admin policies
   - Proposal requirements

3. **Start Small**
   - Observe before participating
   - Learn community norms
   - Build reputation gradually

**Red Flags:**
- Promises of guaranteed returns
- Requests for seed phrases
- Pressure to vote certain way
- Suspicious external links
- Too-good-to-be-true proposals

---

## For Node Operators

### Server Security

**DO:**

1. **Use Dedicated Server**
   - Not shared with other services
   - Clean OS install
   - Minimal software

2. **Harden SSH**

   ```bash
   # Disable password auth
   PasswordAuthentication no

   # Use key auth only
   PubkeyAuthentication yes

   # Disable root
   PermitRootLogin no

   # Non-standard port
   Port 2222
   ```

3. **Enable Firewall**

   ```bash
   sudo ufw allow 26656/tcp  # P2P
   sudo ufw allow 2222/tcp   # SSH (custom port)
   sudo ufw default deny incoming
   sudo ufw enable
   ```

4. **Automatic Updates**

   ```bash
   sudo apt install unattended-upgrades
   sudo dpkg-reconfigure unattended-upgrades
   ```

5. **Install Fail2ban**

   ```bash
   sudo apt install fail2ban
   ```

6. **Monitor Resources**
   - CPU usage
   - RAM usage
   - Disk space
   - Network traffic

7. **Backup Regularly**
   - Node keys
   - Configuration
   - (Optional: chain data)

**DON'T:**

1. **Expose Unnecessary Ports**
   - RPC (26657) should be localhost only (unless needed)
   - gRPC (9090) should be localhost only
   - Only P2P (26656) needs to be public

2. **Use Default Credentials**
   - Change all default passwords
   - Use strong, unique passwords

3. **Run as Root**
   - Create dedicated user
   - Use sudo for admin tasks

4. **Ignore Alerts**
   - Set up monitoring
   - Respond to alerts promptly

5. **Skip Backups**
   - Automate backup process
   - Test restoration regularly

---

### Monitoring

**Essential Metrics:**

1. **Block Height** -- Should increase steadily, alert if stuck >5 minutes
2. **Peer Count** -- Maintain 10+ peers, alert if <5
3. **Disk Space** -- Monitor daily, alert at 80% full
4. **Missed Blocks** (validators) -- Should be near zero, alert at >10 per hour

**Monitoring Stack:**

```
Prometheus (metrics)
    ↓
Grafana (dashboards)
    ↓
Alertmanager (alerts)
    ↓
Telegram/Email/Slack
```

---

## For Validators

### Operational Security

**DO:**

1. **Use Sentry Nodes**

   ```
   Validator (private IP)
       ↓
   Sentry 1, Sentry 2 (public IPs)
       ↓
   Public network
   ```

2. **Secure Validator Key**
   - Use HSM (Yubico, Ledger)
   - Or KMS (AWS KMS, Google KMS)
   - Or Tmkms (Tendermint KMS)
   - Never store plaintext on server

3. **Redundant Infrastructure**
   - Primary + backup validator
   - Only one active at a time
   - Automated failover

4. **Monitor 24/7**
   - Uptime monitoring
   - Alert on downtime
   - Alert on missed blocks
   - On-call rotation (if team)

5. **Document Procedures**
   - Incident response plan
   - Failover procedure
   - Key rotation process
   - Upgrade checklist

6. **Regular Security Audits**
   - Quarterly infrastructure review
   - Penetration testing
   - Log analysis

**DON'T:**

1. **Run Two Validators with Same Key**
   - Causes double-signing
   - 5% slash + permanent jail
   - Use only one active validator

2. **Ignore Missed Blocks**
   - Indicates issues
   - Can lead to jail (1% slash)
   - Investigate immediately

3. **Delay Upgrades**
   - Risk network incompatibility
   - Schedule upgrade window
   - Test on testnet first

4. **Share Validator Access**
   - Limit who has access
   - Use separate accounts
   - Audit access logs

---

### High Availability

**Architecture:**

```
Load Balancer
    ↓
Sentry 1 (Region A)
Sentry 2 (Region B)
    ↓ Private VPN
Validator (Standby in Region C)
Validator (Active in Region A)
    ↓
HSM / KMS
```

**Failover Process:**
1. Monitoring detects validator down
2. Alert sent to on-call
3. Backup validator activated
4. Primary investigated
5. Primary fixed and returned to standby

**Testing:**
- Monthly failover drills
- Quarterly disaster recovery
- Annual full audit

---

## For Developers

### Secure Development

**DO:**

1. **Input Validation**

   ```go
   func ValidateProposal(msg MsgSubmitProposal) error {
       if len(msg.Issue) < 10 || len(msg.Issue) > 200 {
           return ErrInvalidLength
       }

       if !isAlphanumeric(msg.Issue) {
           return ErrInvalidCharacters
       }

       return nil
   }
   ```

2. **Use Prepared Statements**

   ```go
   // Good
   query := "SELECT * FROM domains WHERE name = ?"
   db.Query(query, userInput)

   // Bad (SQL injection)
   query := "SELECT * FROM domains WHERE name = '" + userInput + "'"
   ```

3. **Avoid Hardcoded Secrets**

   ```bash
   # Use environment variables
   export API_KEY="secret"

   # Or secret management
   vault kv get secret/api-key
   ```

4. **Dependency Scanning**

   ```yaml
   # .github/workflows/security.yml
   - name: Run Gosec
     uses: securego/gosec@v2

   - name: Run npm audit
     run: npm audit
   ```

5. **Code Reviews**
   - All PRs require approval
   - Security-focused checklist
   - Test coverage requirements

**DON'T:**

1. **Trust User Input**
   - Always validate
   - Sanitize for display
   - Use parameterized queries

2. **Expose Sensitive Info in Errors**

   ```go
   // Bad
   return fmt.Errorf("database password incorrect: %s", dbPassword)

   // Good
   return fmt.Errorf("authentication failed")
   ```

3. **Skip Tests**
   - Write tests for new code
   - Maintain coverage >80%
   - Include edge cases

4. **Commit Secrets**

   ```bash
   # Use .gitignore
   .env
   *.key
   secrets/
   ```

### Smart Contract Security

**Rust/CosmWasm:**

1. **Check Integer Overflow**

   ```rust
   // Use checked math
   let result = amount.checked_add(fee)?;

   // Not
   let result = amount + fee;  // Can overflow!
   ```

2. **Validate Inputs**

   ```rust
   if amount.is_zero() {
       return Err(ContractError::InvalidAmount {});
   }
   ```

3. **Use Correct Types**

   ```rust
   // Use Uint128 for token amounts
   pub amount: Uint128,

   // Not u64 (too small)
   pub amount: u64,
   ```

4. **Test Edge Cases**
   - Zero amounts
   - Maximum values
   - Reentrancy
   - Access control

---

## General Security

### Password Management

**DO:**
- Use password manager (1Password, Bitwarden)
- Unique password per service
- 20+ character passwords
- Enable 2FA everywhere

**DON'T:**
- Reuse passwords
- Use dictionary words
- Store in plain text
- Share with anyone

---

### Phishing Prevention

**Warning Signs:**
- Urgent requests
- Spelling/grammar errors
- Suspicious sender
- Too good to be true
- Requests for seed phrase

**Verify:**
- Official domain (truerepublic.app)
- HTTPS certificate
- Social media accounts
- Community confirmation

---

### Social Engineering

**Be Skeptical:**
- "Support" contacting you
- Unsolicited investment advice
- Pressure to act quickly
- Requests for remote access
- Offers requiring seed phrase

**Golden Rule:**

> No legitimate person will ever ask for your seed phrase.

---

## Security Checklist

### Daily
- [ ] Check monitoring alerts
- [ ] Review logs for anomalies
- [ ] Verify backup success

### Weekly
- [ ] Review security advisories
- [ ] Check for software updates
- [ ] Review access logs

### Monthly
- [ ] Test backup restoration
- [ ] Review user permissions
- [ ] Update documentation

### Quarterly
- [ ] Security audit
- [ ] Penetration testing
- [ ] Disaster recovery drill

### Annually
- [ ] Key rotation
- [ ] Full infrastructure review
- [ ] Team security training

---

## Learning Resources

**Cosmos SDK:**
- https://docs.cosmos.network/
- https://tutorials.cosmos.network/

**Blockchain Security:**
- https://consensys.github.io/smart-contract-best-practices/

**General Security:**
- https://www.owasp.org/ (OWASP Top 10)
- https://www.sans.org/security-resources/

---

## Next Steps

- [Security Architecture](Security-Architecture) -- Security design
- [Known Issues](Known-Issues) -- Current vulnerabilities
- [Audit Reports](Audit-Reports) -- Security audits
