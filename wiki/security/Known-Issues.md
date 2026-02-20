# Known Issues

Current vulnerabilities, bugs, and limitations.

## Table of Contents

1. [Critical Issues](#critical-issues)
2. [High Severity Issues](#high-severity-issues)
3. [Medium Severity Issues](#medium-severity-issues)
4. [Low Severity Issues](#low-severity-issues)
5. [Limitations](#limitations)
6. [Planned Fixes](#planned-fixes)

---

## Critical Issues

**Status:** None currently identified

---

## High Severity Issues

**Status:** None currently identified

---

## Medium Severity Issues

### M-1: Potential State Bloat

**Component:** x/truedemocracy state storage

**Description:**
Domain state grows unbounded as proposals accumulate. Old RED-zone proposals not automatically deleted without explicit vote.

**Impact:**
- Storage requirements increase over time
- State sync becomes slower
- Query performance degrades

**Affected Versions:** All

**Workaround:**

```bash
# Manual cleanup via governance proposal
# Or encourage domains to vote-to-delete old proposals
```

**Planned Fix:**
- Automatic pruning of proposals >90 days in RED zone
- State archival mechanism
- Target: v0.2.0

**Status:** In Progress

---

### M-2: No Transaction Priority

**Component:** Mempool

**Description:**
Transactions processed first-come-first-served. No priority fee mechanism.

**Impact:**
- High-value transactions can't jump queue
- Network congestion affects all equally
- No spam deterrent beyond base fee

**Affected Versions:** All

**Workaround:**

```bash
# None - submit transaction early
# Or run own validator for guaranteed inclusion
```

**Planned Fix:**
- Implement EIP-1559 style priority fees
- Target: v0.3.0

**Status:** Planned

---

## Low Severity Issues

### L-1: Limited Proposal Search

**Component:** Frontend query performance

**Description:**
No full-text search on proposals. Can only filter by domain.

**Impact:**
- Users can't search proposal content
- Finding specific proposals difficult in large domains

**Affected Versions:** Frontend v0.1.0

**Workaround:**

```bash
# Use browser find (Ctrl+F) on loaded proposals
# Or use REST API with external search indexer
```

**Planned Fix:**
- Elasticsearch integration
- Frontend search UI
- Target: v0.2.0

**Status:** In Progress

---

### L-2: No Batch Transactions

**Component:** Transaction broadcasting

**Description:**
Each action requires separate transaction. Can't batch multiple actions.

**Impact:**
- Multiple transactions needed for complex workflows
- Higher total fees
- Poor UX

**Example:**

```bash
# Want to: join domain + submit proposal + rate 3 suggestions
# Requires: 5 separate transactions
```

**Workaround:**

```bash
# Send transactions sequentially
# Accept multiple tx hashes
```

**Planned Fix:**
- Multi-message transactions
- Atomic execution
- Target: v0.2.0

**Status:** Planned

---

### L-3: Mobile Wallet Not Published

**Component:** React Native app

**Description:**
Mobile wallet code complete but not on App Store / Play Store.

**Impact:**
- Users must build from source
- Limited mobile adoption

**Affected Versions:** Mobile v0.1.0

**Workaround:**

```bash
# Use web wallet on mobile browser
# Or build from source: cd mobile-wallet && npm run build
```

**Planned Fix:**
- Submit to App Store
- Submit to Play Store
- Target: v0.2.0

**Status:** Planned

---

## Limitations

### Design Limitations

**1. No Cross-Domain Proposals**

**Description:**
Proposals are domain-specific. Can't create proposal affecting multiple domains.

**Rationale:** By design -- domains are independent

**Impact:** Multi-domain coordination requires separate proposals

---

**2. Immutable Proposals**

**Description:**
Once submitted, proposals can't be edited. Only deletion allowed.

**Rationale:** Integrity -- prevent bait-and-switch after ratings

**Impact:** Typos require delete + resubmit + PayToPut again

**Mitigation:** Add proposal edit period (first 24 hours)?

---

**3. Single Stone per Member**

**Description:**
Each member has exactly 1 stone at any time.

**Rationale:** Equality + simplicity

**Impact:** Can only prioritize one thing at a time

---

**4. PayToPut Cost Fixed**

**Description:**
10,000 PNYX flat fee, not market-adjustable.

**Rationale:** Predictability

**Impact:** May be too high or too low as PNYX price changes

**Mitigation:** Governance can adjust in future upgrade

---

### Technical Limitations

**1. Maximum Block Size**

**Specification:** 2 MB

**Impact:** ~200 transactions per block maximum

**Throughput:** ~40 TPS (5s blocks, 200 tx/block)

---

**2. State Sync Trust Requirement**

**Description:** State sync requires trusting RPC servers

**Impact:** Can't fully trustlessly sync (though can verify after)

**Mitigation:** Use multiple trusted RPC servers

---

**3. No IBC Yet**

**Description:** Inter-Blockchain Communication not enabled

**Impact:** Can't connect to other Cosmos chains

**Planned:** v0.3.0

---

## Planned Fixes

### v0.2.0 (Q2 2026)

**High Priority:**
- [ ] Automatic proposal pruning
- [ ] Batch transactions
- [ ] Priority fees
- [ ] Proposal edit period (24h)

**Medium Priority:**
- [ ] Full-text search (Elasticsearch)
- [ ] Mobile wallet app store release
- [ ] Enhanced monitoring dashboards

**Low Priority:**
- [ ] UI improvements
- [ ] Documentation updates

---

### v0.3.0 (Q3 2026)

**Major Features:**
- [ ] IBC integration
- [ ] Cross-chain governance
- [ ] Advanced treasury features
- [ ] CosmWasm governance templates

---

### v1.0.0 (Q4 2026)

**Mainnet Launch:**
- [ ] Full security audit
- [ ] Production deployment
- [ ] Bug bounty program
- [ ] Comprehensive documentation

---

## Reporting Issues

**Found a bug?**

1. Check [GitHub Issues](https://github.com/NeaBouli/TrueRepublic/issues)
2. Search existing issues
3. If new, create issue with:
   - Steps to reproduce
   - Expected behavior
   - Actual behavior
   - Logs/screenshots
   - Environment (OS, version, etc.)

**Security vulnerability?**

**DO NOT** create public issue.

Email: security@truerepublic.network

---

## Issue Severity Guide

| Severity | Examples |
|----------|---------|
| **Critical** | Network halt, consensus failure, fund loss, private key exposure |
| **High** | Validator slashing bug, transaction failure, major DoS, smart contract exploit |
| **Medium** | Performance degradation, minor DoS, UI/UX issues, documentation gaps |
| **Low** | Cosmetic bugs, minor inconveniences, feature requests, documentation typos |

---

## Next Steps

- [Best Practices](Best-Practices) -- Security recommendations
- [Audit Reports](Audit-Reports) -- Security audits
- [Test Coverage](Test-Coverage) -- Testing status
