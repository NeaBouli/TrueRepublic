# Incident Command and Rehearsal

Status: repository-qualified recovery/testnet procedure. This document and the
synthetic rehearsal contract do not configure a live on-call rota, paging
destination, production topology, or authority to operate a public network.

Use this runbook to coordinate the existing specialist procedures without
weakening their safety boundaries. It defines who decides, which evidence must
exist before recovery, when to stop, and which procedure is authoritative for
each incident class.

## Non-negotiable safety boundaries

- Never put a private key, mnemonic, signer state, keyring, credential, token,
  signature, raw transaction, private host, or real topology into the public
  contract, logs, tickets, chat, or rehearsal evidence.
- One consensus identity may have exactly one active signer. Prove the source
  stopped and every automatic restart path is disabled before another signer
  can start. Never restore older `priv_validator_state.json` after later
  signing may have occurred.
- A compatible binary rollback is allowed only when the candidate failed
  before opening or mutating state. Otherwise remain stopped and escalate to a
  coordinated recovery decision.
- A legacy authority migration uses a distinct target chain and immutable
  source artifacts. Source and target validators may never run together and
  their data or signer state may never be mixed.
- Suspected operator-authority compromise cannot be repaired by consensus-key
  rotation. Freeze authority actions, isolate signers, remain safely stopped,
  and require a separately approved manual governance recovery.
- No restart or closure is valid without trusted common-height app-hash,
  progress, validator-power, and ledger-invariant evidence, except the explicit
  `safe-stopped` outcome for compromised operator authority.

## Roles and decision authority

The contract uses durable logical roles and fixed synthetic seat aliases rather
than personal names. Before a controlled canary, each role needs an
independently reachable primary and secondary human in a private operator
register; that mapping remains outside this contract and repository.

| Role | Responsibility |
|---|---|
| `incident-commander` | Severity, scope, decision log, deadlines, stop/go, and closure |
| `security-owner` | Containment, compromise assessment, secret/evidence boundaries |
| `protocol-owner` | Consensus, app-hash, validator-power, invariant, and restart judgment |
| `validator-operator` | Approved node/signing action under the selected specialist runbook |
| `evidence-custodian` | Immutable references, access control, checksums, and chain of custody |
| `communications-owner` | Rehearsal notices and stakeholder handoff without sensitive content |
| `release-owner` | Binary identity, compatibility boundary, rollout freeze, and rollback window |

Repository role aliases prove schema completeness only. They are not live
assignments, identities, contacts, or proof that external paging works.

## Required state progression

Every scenario has exactly seven forward-only phases:

1. `detect` — classify the incident and record start, alert snapshot, and
   logical chain identity.
2. `contain` — freeze concurrent changes, isolate the affected process or
   authority, and disable unsafe restart paths.
3. `preserve` — capture secret-free log references, trusted height, trusted app
   hash, and scenario-specific evidence before recovery changes state.
4. `decide` — record the accountable decision, deadline, communication path,
   and required approvals. Critical incidents require incident, security, and
   protocol approval.
5. `recover` — follow exactly the specialist runbook selected for the scenario.
6. `validate` — require the declared terminal outcome and all integrity gates.
7. `close` — preserve a postmortem reference and obtain commander approval.

Do not skip, reorder, reopen, or silently retry a phase. A required abort
condition ends the attempted path safely; it is not a test failure to hide.

## Scenario routing

| Scenario | Authority | Authoritative procedure | Critical boundary |
|---|---|---|---|
| Chain halt | Coordinated protocol/release | This runbook plus [Monitoring](monitoring.md) | Stop enough signers to prevent ambiguous progress; restart only from agreed height/app hash |
| Validator failure | Operator | [Validator Identity Custody and Recovery](validator-identity-recovery.md) | Source stopped, signer state current, exactly one recovered signer |
| Validator slashing | Operator with critical review | [Validator Slashing and Recovery](validator-slashing.md) | Preserve offense/economic evidence; distinguish downtime repair from compromised-key rotation |
| Consensus-key compromise | Independent operator account | [Validator Consensus-Key Rotation](validator-key-rotation.md) | Old signer isolated; operator authority known safe; old key permanently revoked |
| Operator-authority compromise | Manual governance | This runbook | No automated rotation or replacement; terminal outcome is safely stopped |
| Backup/restore | Operator | [Backup & Recovery](backup-recovery.md) | Sanitized archive only; restore to a fresh full-node identity, not a validator signer |
| Compatible binary upgrade/rollback | Coordinated protocol/release | [Compatible Binary Upgrades and Rollback](upgrades.md) | Simple rollback only before candidate state opens or mutates |
| Legacy authority migration | Manual governance | [Legacy Validator-Authority Migration](legacy-authority-migration.md) | Distinct chain IDs, frozen export, no source/target concurrency or state mixing |

## Offline rehearsal contract

The maintained example is
`configs/incidents/rehearsal.example.json`. It contains only controlled enums,
logical aliases, action classes, evidence classes, approvals, and abort
conditions. It intentionally contains no commands, paths outside repository
runbooks, values of evidence, hosts, URLs, providers, contact data, or secrets.

Validate it without initializing or reading a node home:

```bash
truerepublicd incident-rehearsal validate \
  --file configs/incidents/rehearsal.example.json \
  --output json |
  jq -e '.valid == true and .scenario_count == 8 and
    (.violations | type == "array" and length == 0)'
```

The parser rejects unknown or duplicate fields, trailing data, excessive JSON
depth, oversized input, unsupported strings, missing roles or scenarios,
duplicate actions, phase reordering, missing evidence/approvals, mismatched
authority, unsafe outcomes, and incomplete abort boundaries. Validation is
read-only: it never resolves a host, reads a node home, invokes a runbook,
creates evidence, sends a notification, or changes infrastructure.

## Rehearsal procedure

1. Copy the synthetic contract into an isolated exercise workspace. Keep its
   schema, scenario kinds, and fixed synthetic seat aliases unchanged. Map
   participants to those seats only in the private exercise record.
2. Run the offline validator and preserve its JSON report with the reviewed
   commit and binary checksum.
3. Select one scenario. The facilitator injects synthetic alerts and evidence
   references; participants progress through every phase and explicitly test
   at least one listed abort condition.
4. Use temporary local or private-testnet assets only when exercising an
   existing process harness. Never substitute production keys or infrastructure.
5. Record phase timestamps, role acknowledgements, evidence-reference IDs,
   decisions, and validation results in a private exercise record. Store
   sensitive evidence separately; the record contains references only.
6. Require protocol and security review of integrity gates. For a restored
   outcome, verify common-height app hash, height progress, ledger invariants,
   validator power where applicable, and single-signer safety.
7. Close only after the incident commander records residual risk and a
   postmortem reference. A failed or aborted rehearsal remains valid evidence
   of a rollout gap and must not be rewritten as a pass.

## Evidence record

The private exercise record should identify:

- rehearsal version, repository commit, scenario ID, start/end time, and role
  acknowledgements;
- immutable references to alert/log snapshots, trusted height/app hash,
  binaries, sanitized archives, and decisions;
- each phase result, required approval, triggered abort condition, and exact
  terminal outcome;
- common-height/app-hash, progress, ledger, validator-power, signer-safety, and
  source/target-isolation results where applicable; and
- follow-up owner and deadline for every failed gate.

Do not place evidence values in the public example. Checksums of secret key or
signer-state material remain inside the approved private custody system and are
not suitable for public tickets or source control.

## What this gate does not prove

- live notification delivery, acknowledgement, or named on-call coverage;
- production host access, firewall/TLS/DNS state, capacity, or DDoS resistance;
- HSM, remote-signer, vault, or key-custody integration;
- a governance-controlled in-place state migration or Cosmos SDK `x/upgrade`;
- independent production rehearsal or rollout approval.

Those items remain open under the [Road to Rollout](../../ROLLOUT_ROADMAP.md)
and [GH-29](https://github.com/NeaBouli/TrueRepublic/issues/29).
