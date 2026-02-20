# Security Architecture

Comprehensive security design of TrueRepublic blockchain.

## Table of Contents

1. [Security Principles](#security-principles)
2. [Authentication & Authorization](#authentication--authorization)
3. [Network Security](#network-security)
4. [Cryptographic Security](#cryptographic-security)
5. [Smart Contract Security](#smart-contract-security)
6. [Infrastructure Security](#infrastructure-security)
7. [Attack Vectors & Mitigations](#attack-vectors--mitigations)

---

## Security Principles

### Defense in Depth

Multiple layers of security:

```
Layer 1: Network (Firewall, DDoS protection)
Layer 2: Application (Input validation, rate limiting)
Layer 3: Consensus (BFT, slashing)
Layer 4: Cryptography (Signatures, encryption)
Layer 5: Infrastructure (Access control, monitoring)
```

### Zero Trust

**Assumptions:**
- All inputs are untrusted
- All networks are hostile
- All users must authenticate
- All transactions must be signed

**Verification:**
- Signature verification on all transactions
- State transition validation
- Consensus validation
- Input sanitization

### Principle of Least Privilege

**Access Control:**
- Domain membership required for proposals
- Validator status required for consensus
- Admin role required for domain config
- Module permissions enforced

---

## Authentication & Authorization

### Account System

**Cosmos SDK Accounts:**

```
secp256k1 ECDSA
├── Private Key (32 bytes)
├── Public Key (33 bytes compressed)
└── Address (20 bytes, Bech32 encoded)
```

**Address Format:**

```
cosmos1abc...xyz  (user address)
cosmosvaloper1... (validator operator)
cosmosvalcons1... (validator consensus)
```

### Transaction Signing

**Signature Process:**

```
1. Create transaction
2. Sign with private key (ECDSA)
3. Attach signature + public key
4. Broadcast to network
5. Node verifies signature
6. Execute if valid
```

**Signature Verification:**

```go
func VerifySignature(tx Tx) bool {
    pubKey := tx.GetSigners()[0]
    signature := tx.GetSignature()
    message := tx.GetSignBytes()

    return pubKey.VerifySignature(message, signature)
}
```

### Authorization Checks

**Domain Membership:**

```go
func (k Keeper) SubmitProposal(ctx sdk.Context, msg MsgSubmitProposal) error {
    domain := k.GetDomain(ctx, msg.Domain)

    // Authorization check
    if !domain.IsMember(msg.Creator) {
        return ErrNotAuthorized
    }

    // Proceed with proposal
}
```

**Validator Status:**

```go
func (k Keeper) RegisterValidator(ctx sdk.Context, msg MsgRegisterValidator) error {
    // Check domain membership
    if !k.IsDomainMember(ctx, msg.Address, msg.Domain) {
        return ErrNotDomainMember
    }

    // Check stake provenance (PoD)
    if err := k.ValidateStakeProvenance(ctx, msg); err != nil {
        return err
    }

    // Register validator
}
```

---

## Network Security

### P2P Security

**Peer Authentication:**

```
1. Peer connects
2. Handshake exchange
3. Node ID verification
4. Peer added to known peers
```

**Sybil Resistance:**
- Limited peer connections (40 inbound, 10 outbound)
- Peer reputation tracking
- Connection rate limiting
- Ban list for malicious peers

**DDoS Protection:**

```toml
[p2p]
max_num_inbound_peers = 40
max_num_outbound_peers = 10
send_rate = 5120000  # 5 MB/s
recv_rate = 5120000  # 5 MB/s

# Connection limits
max_packet_msg_payload_size = 1024
```

### Firewall Configuration

**Minimal Attack Surface:**

```bash
# Allow only necessary ports
sudo ufw allow 26656/tcp  # P2P
sudo ufw allow 26657/tcp  # RPC (optional, local only recommended)

# Block everything else
sudo ufw default deny incoming
sudo ufw default allow outgoing
```

**Recommended Setup:**

```
Validator Node (private)
    ↓ Private connection
Sentry Nodes (public)
    ↓ Public internet
External Peers
```

### Rate Limiting

**API Rate Limits:**

```nginx
# nginx.conf
limit_req_zone $binary_remote_addr zone=api:10m rate=30r/s;
limit_req_zone $binary_remote_addr zone=rpc:10m rate=10r/s;

location /api/ {
    limit_req zone=api burst=50;
}

location /rpc/ {
    limit_req zone=rpc burst=20;
}
```

---

## Cryptographic Security

### Signature Scheme

**Algorithm:** secp256k1 ECDSA (same as Bitcoin, Ethereum)

**Why secp256k1?**
- Well-studied (Bitcoin uses it)
- Efficient verification
- Widely supported
- Hardware wallet compatible

**Key Derivation:**

```
Mnemonic (24 words)
    ↓ BIP39
Seed (64 bytes)
    ↓ BIP44
Private Key (32 bytes)
    ↓ ECDSA
Public Key (33 bytes)
    ↓ Hash
Address (20 bytes)
```

### Anonymous Voting Keys

**Dual-Key System:**

```
Master Private Key
    ↓ Derive (Domain A)
Domain A Private Key (unlinkable to master)
    ↓ Derive (Domain B)
Domain B Private Key (unlinkable to A or master)
```

**Key Derivation Function:**

```go
func DeriveDomainKey(masterKey, domain string) PrivateKey {
    // Use HKDF with domain name as info
    hkdf := hkdf.New(sha256.New, masterKey, nil, []byte(domain))

    derivedKey := make([]byte, 32)
    hkdf.Read(derivedKey)

    return PrivateKey(derivedKey)
}
```

**Unlinkability:**
- No mathematical relationship between keys
- Cannot derive master from domain key
- Cannot link two domain keys
- Forward secrecy (old keys can be deleted)

### Hash Functions

**SHA-256:**
- Block hashing
- Transaction hashing
- Merkle tree construction

**Properties:**
- Collision resistant
- Pre-image resistant
- Avalanche effect

---

## Smart Contract Security

### CosmWasm Security Model

**Sandboxing:**

```
CosmWasm Contract
    ↓ Wasm Runtime
Limited Memory
No File System Access
No Network Access
Gas Metering
```

**Actor Model:**
- Contracts isolated from each other
- No shared state
- Message passing only
- Sequential execution

**Gas Metering:**

```rust
// Every operation costs gas
pub fn execute(deps: DepsMut, env: Env, info: MessageInfo) -> Result<Response> {
    // Gas is consumed for:
    // - Storage reads/writes
    // - Computation
    // - Message passing

    if msg.amount > limit {
        return Err(ContractError::InsufficientGas {});
    }

    Ok(Response::new())
}
```

### Contract Deployment

**Deployment Process:**

```
1. Developer writes contract
2. Compile to Wasm
3. Submit to governance
4. Governance votes
5. If approved, store code
6. Instantiate contract
```

**Code Verification:**
- Deterministic builds
- Source code available
- Audit reports required
- Governance approval

---

## Infrastructure Security

### Server Hardening

**SSH Security:**

```bash
# Disable password auth
PasswordAuthentication no
PubkeyAuthentication yes

# Disable root login
PermitRootLogin no

# Change default port
Port 2222
```

**Automatic Updates:**

```bash
# Enable unattended upgrades
sudo apt install unattended-upgrades
sudo dpkg-reconfigure --priority=low unattended-upgrades
```

**Fail2ban:**

```bash
# Install
sudo apt install fail2ban

# Configure
sudo nano /etc/fail2ban/jail.local

[sshd]
enabled = true
maxretry = 3
bantime = 3600
```

### Key Management

**Validator Key Protection:**

**Option 1: Hardware Security Module (HSM)**

```
Private Key stored in HSM
    ↓
Signing requests sent to HSM
    ↓
HSM signs, returns signature
    ↓
Signature broadcast to network
```

**Option 2: KMS (Key Management Service)**

```
AWS KMS / Google KMS
    ↓
Encrypted key storage
    ↓
API signing service
    ↓
Signature returned
```

**Option 3: Tmkms (Tendermint KMS)**

```
Validator Node
    ↓ Unix socket
Tmkms Process
    ↓ Hardware key
YubiHSM / Ledger
```

**Key Backup:**

```bash
# Backup critical keys
cp ~/.truerepublic/config/priv_validator_key.json ~/backup/
cp ~/.truerepublic/config/node_key.json ~/backup/

# Encrypt backup
gpg --encrypt --recipient your@email.com backup/

# Store offline (USB drive, safe, etc.)
```

---

## Attack Vectors & Mitigations

### 1. Double-Spend Attack

**Attack:** Submit same transaction twice

**Mitigation:**
```
1. Nonce/sequence number per account
2. Transaction hash stored in mempool
3. Duplicate detection
4. Consensus validation
```

### 2. Sybil Attack

**Attack:** Create many fake identities

**Mitigation:**
```
1. Proof-of-Domain (validators)
2. PayToPut (proposals)
3. Stake requirement (100K PNYX)
4. Peer connection limits
```

### 3. Eclipse Attack

**Attack:** Isolate node from network

**Mitigation:**
```
1. Multiple seed nodes
2. Diverse peer connections
3. Peer reputation
4. Monitor peer count (alert <5)
```

### 4. Long-Range Attack

**Attack:** Rewrite history from old state

**Mitigation:**
```
1. Instant finality (CometBFT)
2. No chain reorganizations
3. Checkpointing
4. Social consensus (fork choice)
```

### 5. Nothing-at-Stake

**Attack:** Validator votes on multiple forks

**Mitigation:**
```
1. Slashing (5% for double-sign)
2. Tombstoning (permanent ban)
3. Evidence submission
4. Consensus rules enforcement
```

### 6. DDoS Attack

**Attack:** Overwhelm node with requests

**Mitigation:**
```
1. Rate limiting (nginx)
2. Firewall rules
3. DDoS protection service (Cloudflare)
4. Sentry node architecture
5. Connection limits
```

### 7. Smart Contract Exploits

**Attack:** Malicious contract code

**Mitigation:**
```
1. Gas limits (prevent infinite loops)
2. Sandboxing (Wasm isolation)
3. Code audits required
4. Governance approval
5. Bug bounty program (future)
```

### 8. Private Key Compromise

**Attack:** Steal validator/user keys

**Mitigation:**
```
1. Hardware wallets
2. HSM/KMS
3. Multi-sig (future)
4. Key rotation
5. Monitoring for unauthorized txs
```

### 9. Governance Attack

**Attack:** Malicious proposal passed

**Mitigation:**
```
1. PayToPut (10K PNYX spam deterrent)
2. Systemic Consensing (resistance visible)
3. Vote to Delete (2/3 majority)
4. Domain isolation
5. Time delays (proposal lifecycle)
```

### 10. Stake Grinding

**Attack:** Manipulate validator selection

**Mitigation:**
```
1. Deterministic validator selection
2. Proof-of-Domain limits
3. Stake provenance tracking
4. Transfer limits (10%)
```

---

## Security Checklist

### Node Operators

- [ ] Firewall configured (only necessary ports)
- [ ] SSH hardened (key auth, no root, non-standard port)
- [ ] Fail2ban installed
- [ ] Automatic security updates enabled
- [ ] Monitoring + alerting configured
- [ ] Backups automated + tested
- [ ] Keys backed up offline
- [ ] DDoS protection enabled

### Validators

- [ ] HSM or KMS for validator key
- [ ] Sentry node architecture
- [ ] Redundant infrastructure
- [ ] 24/7 monitoring
- [ ] Incident response plan
- [ ] Key rotation schedule
- [ ] Regular security audits
- [ ] Disaster recovery tested

### Developers

- [ ] Input validation on all messages
- [ ] Access control checks
- [ ] Error handling (no sensitive info leaks)
- [ ] Rate limiting
- [ ] SQL injection prevention (if using DB)
- [ ] XSS prevention (frontend)
- [ ] CSRF protection (frontend)
- [ ] Dependency scanning (Dependabot)
- [ ] Security tests in CI

---

## Incident Response

### Detection

**Indicators of Compromise:**
- Unusual transaction patterns
- Unauthorized key usage
- High missed blocks
- Unexpected stake changes
- Abnormal resource usage

### Response Plan

**1. Identify (0-15 min)**
- Confirm incident
- Assess severity
- Activate response team

**2. Contain (15-60 min)**
- Isolate affected systems
- Revoke compromised keys
- Stop validators if necessary

**3. Eradicate (1-4 hours)**
- Remove malicious code
- Patch vulnerabilities
- Rotate keys

**4. Recover (4-24 hours)**
- Restore from backup
- Verify integrity
- Resume operations

**5. Lessons Learned (24-48 hours)**
- Post-mortem analysis
- Update procedures
- Improve defenses

---

## Security Contact

**Report vulnerabilities:**
- Email: security@truerepublic.network
- PGP Key: (see GitHub)
- Response time: 24 hours

**Bug Bounty Program:**
- Coming soon
- Rewards: 100 - 10,000 PNYX

---

## Next Steps

- [Audit Reports](Audit-Reports) -- Security audit findings
- [Test Coverage](Test-Coverage) -- Testing documentation
- [Known Issues](Known-Issues) -- Current vulnerabilities
- [Best Practices](Best-Practices) -- Security recommendations
