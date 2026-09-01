# Optional Domain Ballot Architecture

**Design issue:** [GH-231](https://github.com/NeaBouli/TrueRepublic/issues/231)

**Deferred implementation epic:** [GH-232](https://github.com/NeaBouli/TrueRepublic/issues/232)

**Status:** implementation-ready proposal; no consensus or client behavior is
implemented by this document.

This design adds optional, domain-configurable formal ballots without replacing
TrueRepublic's systemic-consensing process. It separates proposal development
from binding decision rules and supports public, pseudonymous and private
participation profiles.

Architecture documentation earns no rollout checkbox. The published state
remains 2,319 recovery-verified tests, rollout 35/59, phase work 35/51, Phase 6
6/7 and `production_ready: false`.

## 1. Goals and non-goals

### Goals

- preserve systemic consensing as the default proposal-development mechanism;
- let each domain define defaults for future ballots;
- freeze an immutable policy, option set and electorate for every ballot;
- support systemic-consensing, yes/no/abstain, person-election and hybrid
  consensing-to-ratification ballots;
- make quorum, thresholds, abstentions, ties, runoffs and revoting explicit;
- keep all consensus transitions deterministic, restart-safe and export/import
  safe;
- separate public candidate identity from voter identity and ballot secrecy;
- reuse the existing membership-proof/nullifier foundation after its production
  qualification gates pass; and
- expose one fail-closed transaction/query contract to Beta and Alpha clients.

### Non-goals

- changing the current chain, module version, ZKP artifacts or clients now;
- representing blockchain auditability as legal validity;
- putting names, identity documents or external identity numbers on-chain;
- claiming receipt-freeness, coercion resistance or production ZKP readiness;
- activating a new mode retroactively for an existing domain or election; or
- making a ballot result execute arbitrary code or treasury actions.

## 2. Verified current behavior

This section describes code on `main`, not the proposed system.

### 2.1 Systemic consensing

- Suggestions accept ratings from -5 to +5. `scoring.go` sums ratings and ranks
  by score, then stones, then stable creation order.
- Domain-key ratings use the permission register. The Groth16 path proves
  membership and persists a nullifier. GH-209 binds chain, domain, issue,
  suggestion, rating and reward recipient into the `TrueRepublic/vote/v2`
  signal.
- The committed prover artifacts remain synthetic/test-only. Maintained-client
  proof submission is disabled pending the rollout Phase 2 ceremony,
  reproducibility and independent cryptographic/privacy gates.
- Ratings belong to the continuous suggestion lifecycle. They are not
  first-class, time-bounded ballots with a frozen electorate or final result.

### 2.2 Person-election storage

`types.go` defines simple majority, absolute majority and a systemic-consensing
mode plus approve/abstain choices. `election.go` stores one value under
`elecvote:{domain}:{issue}:{voter}`. A later vote overwrites the earlier value.

The current tally helper:

- uses the live domain member list rather than a frozen electorate;
- has no opening/closing height, quorum, finalized outcome or execution step;
- supports candidate approval or abstention, not yes/no/abstain;
- excludes abstentions for simple majority and uses all live members for
  absolute majority;
- uses a placeholder `bestVotes > 0` result for the systemic mode;
- iterates a Go map when choosing an equal-vote leader; and
- has no production EndBlock, message or query caller. It is exercised by Go
  tests but must not be described as a binding election engine.

Election votes are keyed by signer-derived voter address and are therefore
linkable. They are also absent from the current genesis export/import model.

### 2.3 Existing reusable patterns

- `upgrade_gov.go` already creates a sorted, deduplicated, immutable electorate
  snapshot and uses exact integer `votes * 3 >= eligible * 2` threshold math.
- Exclusion and fast-delete rules already use basis points.
- Signer resolution rejects caller-supplied identity claims that do not match
  the authenticated signer.
- The ZKP circuit, Merkle membership, accepted roots and nullifier persistence
  provide a future substrate for anonymous ballots after production
  qualification.
- EndBlock already owns synchronous governance lifecycle processing.

## 3. Domain defaults and immutable ballot policy

A domain may store a versioned `DomainBallotDefaults` record. Defaults only
pre-fill a new ballot. Creation copies and validates all effective values into
an immutable `BallotPolicy`; later domain changes affect only later ballots.

Every ballot freezes at creation:

- domain, ballot ID, creator and creation height;
- ballot mode and privacy profile;
- exact option/candidate IDs and content hashes;
- sorted, deduplicated eligible member commitments or addresses;
- opening and closing heights;
- quorum, participation and approval denominator rules;
- abstention, revote, tie, runoff, cancellation and finalization rules;
- parent/stage linkage for hybrid and runoff ballots; and
- the policy version and canonical policy hash.

No admin can amend an active ballot. A content or policy amendment creates a
new ballot ID and leaves the old record auditable.

## 4. Ballot lifecycle

```text
PENDING --opens_at_height--> OPEN --closes_at_height--> CLOSED --> FINALIZED
   |                          |                                      |  PASSED
   +---- authorized cancel --+--> CANCELLED                         |  REJECTED
                                  reason_code + height              |  NO_QUORUM
                                                                    |  NO_WINNER
                                                                    +  RUNOFF_CREATED
```

- `PENDING`: policy, electorate and options are frozen; votes are rejected.
- `OPEN`: eligible participants may cast a valid vote under the selected
  privacy/revote profile.
- `CLOSED`: votes are rejected; deterministic tally has not yet committed.
- `FINALIZED`: outcome and all tally inputs are immutable.
- `CANCELLED`: terminal; allowed only by the policy's explicit authority and
  only from `PENDING` or `OPEN`, optionally before an earlier policy cutoff.
  `CLOSED` and `FINALIZED` ballots cannot be cancelled. A non-empty canonical
  reason code and cancellation height are stored in both the ballot and terminal
  outcome record. Cancellation emits `ballot_cancelled`; it never produces
  tallies, a winner or a runoff.

EndBlock performs height-based transitions synchronously. It finalizes closed
ballots in a deterministic bounded batch ordered by `(domain, ballot_id)`.
Capacity bounds must prevent an attacker from making EndBlock unbounded; excess
eligible finalizations remain closed and are processed in the next block in the
same canonical order. A permissionless finalize message may be added only if it
calls the identical pure tally function and cannot choose timing-dependent
inputs.

## 5. Supported modes

### 5.1 Systemic-consensing ballot

Participants rate every eligible option from -5 to +5. Policy defines whether
all options must be rated and how missing ratings are handled. The first
implementation should require a complete rating vector per participant so
unequal option participation cannot bias raw sums.

The canonical score is an integer sum. Highest score means least resistance,
matching current `scoring.go`. Ties never use map iteration: the outcome is
`NO_WINNER` or a runoff according to policy. Stones may be displayed as
context but must not silently break a formal ballot tie.

### 5.2 Yes/no/abstain ballot

One immutable proposition accepts `YES`, `NO` or `ABSTAIN`. Abstention may count
toward participation but never silently change the approval denominator. The
ballot stores the exact denominator rule described in Section 6.

### 5.3 Person election

- A single-candidate confirmation uses yes/no/abstain.
- A multi-candidate election accepts one candidate, abstain and optionally
  `NONE_OF_THE_ABOVE`.
- Policy selects simple majority, absolute majority, qualified plurality or a
  deterministic runoff.
- Candidate names/profiles may be public while voter identity remains private.
  These are separate policy decisions.
- A runoff is a new linked ballot with the same original electorate snapshot,
  an exact candidate subset and a new voting window.

Ranked-choice/Condorcet voting is intentionally outside the first
implementation. It requires a separately reviewed tally specification.

### 5.4 Consensing-to-ratification hybrid

1. Stage A rates competing proposals by systemic consensing.
2. At close, the deterministic winner and exact proposal-content hash are
   frozen. A tie or failed quorum produces no ratification ballot.
3. Stage B creates a linked yes/no/abstain ballot over that exact hash.
4. Only Stage B can produce a passed formal decision.

The same creation-time electorate is recommended for both stages. If a domain
chooses a fresh Stage B electorate, that choice and snapshot rule must be fixed
in Stage A's policy before voting begins.

## 6. Exact counting semantics

All rates use integers in `[0, 10_000]` basis points. No floating-point math,
wall-clock time, randomized tie-breaking or unordered map iteration is allowed.

### Quorum

```text
quorum_met := participants * 10_000 >= electorate_size * quorum_bps
```

`participation_rule` is one of:

- `ALL_CAST`: yes/no/candidate/abstain all count;
- `VALID_NON_ABSTAIN`: abstentions do not count; or
- `COMPLETE_RATING_VECTOR`: one complete systemic-consensing vector counts.

### Approval denominator

- `VALID_DECISIVE`: `yes + no`, excluding abstentions;
- `ALL_CAST`: `yes + no + abstain`; or
- `ELECTORATE`: frozen eligible count, including non-participation.

```text
threshold_met := approvals * 10_000 >= denominator * approval_bps
```

Policy also stores `comparison = AT_LEAST | STRICTLY_GREATER`. Strict majority
therefore uses `approval_bps = 5_000` and `STRICTLY_GREATER`; exact two thirds
uses an explicit rational `(2,3)` comparison where required so `6_667` basis
points cannot introduce unintended rounding. Zero denominators fail closed.

### Qualified plurality

For a multi-candidate ballot, qualified plurality means that exactly one top
candidate must both outpoll every other candidate and meet the policy's
explicit `approval_bps` (or exact rational) against its selected approval
denominator:

```text
qualified := top_votes * 10_000 >= denominator * approval_bps
```

`comparison = STRICTLY_GREATER` changes `>=` to `>` exactly as above. The
denominator is one of `VALID_DECISIVE` (all valid candidate selections,
excluding abstentions), `ALL_CAST`, or `ELECTORATE`; it is immutable and cannot
be inferred from UI wording. A zero denominator, equal top count, or top count
below the threshold yields `NO_WINNER`, unless the frozen `tie_rule`/failure
rule requires a linked runoff. Meeting the threshold by one integer cross-
multiplication unit passes; missing it by one fails. A runoff repeats the same
formula under its own immutable policy and never promotes a candidate by
lexicographic order.

### Revoting, duplicates and ties

- `FINAL_VOTE_WINS`: signer-attributed profiles may overwrite until close;
- `IMMUTABLE_FIRST_VOTE`: later votes fail; required for the initial ZK
  nullifier design; or
- a future unlinkable revote protocol, which needs separate cryptographic
  review and is not implied here.

Duplicate transactions are idempotent only when their canonical vote bytes are
identical. Conflicting reuse fails. Equal top results produce `NO_WINNER` or a
linked runoff; lexicographic ordering is only for deterministic storage/output,
not a substantive winner rule.

## 7. Proposed Go/state contract

The following names are design-level and deliberately not present in code yet.

```go
type Ballot struct {
    DomainName        string
    ID                uint64
    Creator           string
    Status            BallotStatus
    Policy            BallotPolicy
    PolicyHash        []byte
    Electorate        []string // sorted addresses or commitments
    ElectorateRoot    []byte
    Options           []BallotOption
    CreatedAtHeight   int64
    OpensAtHeight     int64
    ClosesAtHeight    int64
    ParentBallotID    uint64
    Stage             uint32
    CancellationReasonCode string // non-empty only when CANCELLED
    CancelledAtHeight int64        // zero unless CANCELLED
}

type BallotPolicy struct {
    Version             uint32
    Mode                BallotMode
    Privacy             BallotPrivacy
    QuorumBps           uint32
    ParticipationRule   ParticipationRule
    ApprovalRule        ApprovalRule
    AbstentionAllowed   bool
    RevoteRule          RevoteRule
    TieRule             TieRule
    RunoffSize          uint32
    CancellationRule    CancellationRule
}

type BallotVote struct {
    BallotID       uint64
    Choice         BallotChoice
    Ratings        []OptionRating
    VoterKey       string // address/profile key; signer profiles only
    NullifierHash  []byte // ZK profiles only
    VoteDigest     []byte // hash of canonical choice/rating bytes
    CastAtHeight   int64
}

type BallotNullifierRecord struct {
    NullifierHash  []byte
    VoteDigest     []byte
    AcceptedAtHeight int64
}

type BallotOutcome struct {
    BallotID       uint64
    Status         BallotStatus // FINALIZED or CANCELLED
    Code           OutcomeCode // CANCELLED when Status is CANCELLED
    Tallies        []OptionTally
    Participants   uint64
    QuorumMet      bool
    WinnerOptionID string
    TerminatedAtHeight int64
    CancellationReasonCode string // non-empty only when CANCELLED
    ResultHash     []byte
}
```

### Store prefixes

```text
ballot:next:{domain}                         -> uint64
ballot:{domain}:{ballot_id}                  -> Ballot
ballot:vote:{domain}:{ballot_id}:{voter_key} -> BallotVote (signer profiles)
ballot:zk-vote:{domain}:{ballot_id}:{nullifier_hash} -> BallotVote (SECRET_ZK)
ballot:nullifier:{domain}:{ballot_id}:{hash} -> BallotNullifierRecord
ballot:outcome:{domain}:{ballot_id}          -> BallotOutcome
ballot:close:{height}:{domain}:{ballot_id}    -> empty index value
ballot:domain-policy:{domain}                -> DomainBallotDefaults
```

Keys use length-delimited binary encoding in the implementation; the readable
forms above only specify namespace and ordering. `SECRET_ZK` never uses an
empty `voter_key`: its ballot-scoped nullifier is the vote-record key. The
nullifier record persists the canonical vote digest, so replaying byte-identical
vote content is idempotent while reusing the nullifier for a different digest
fails. Signer-attributed profiles remain indexed by `VoterKey`.

### Messages

- `MsgSetDomainBallotDefaults` — governed update affecting future ballots only;
- `MsgCreateBallot` — validates/finalizes policy, options and snapshots;
- `MsgCastBallotVote` — signer-authenticated public/pseudonymous vote;
- `MsgCastBallotVoteWithProof` — proof, ballot nullifier and choice signal;
- `MsgCancelBallot` — exact authority, state and reason validation; and
- `MsgFinalizeBallot` — optional permissionless trigger for the pure tally.

Messages derive identity from authenticated signers or verified proofs. A
caller-supplied voter address never establishes authority.

### Queries and events

Queries: `Ballot`, `BallotsByDomain`, `BallotPolicy`, `BallotOutcome`,
`BallotTally`, `BallotEligibility` and profile-appropriate receipt/nullifier
status. A cancelled ballot returns its immutable reason code and termination
height through `Ballot` and `BallotOutcome`, with `Code = CANCELLED` and empty
tallies. `OutcomeCode` therefore contains `CANCELLED` in addition to the five
finalized-result codes shown in Section 4. Secret profiles never return a
voter-to-choice mapping.

Events: `ballot_created`, `ballot_opened`, `ballot_vote_accepted`,
`ballot_closed`, `ballot_finalized`, `ballot_cancelled`, and
`ballot_runoff_created`. Secret-profile events expose ballot ID and nullifier,
not signer or reward-recipient linkage unless the profile explicitly declares
that leakage.

## 8. Genesis, migration and compatibility

Implementation requires a `truedemocracy` consensus-version bump and an
explicit migration. The migration must be additive:

- initialize domain defaults as disabled;
- do not convert legacy `elecvote:` records into binding ballots;
- preserve legacy records only as explicitly historical/advisory state or
  remove them in a later separately approved migration after export evidence;
- export/import domain defaults, ballots, electorate snapshots, votes,
  nullifiers (including canonical vote digests), close indexes, cancellation
  metadata and terminal outcomes exactly;
- reject duplicate IDs/options/electorate entries, unsorted snapshots, invalid
  hashes, impossible heights/states, unknown enums, unsafe thresholds and
  outcome/tally mismatches; and
- re-import at the same logical height with identical app hash and future
  transitions.

The feature remains disabled until a governed activation parameter is approved.
Old clients must fail closed on unknown ballot messages rather than reinterpret
them as legacy election votes.

## 9. Identity and privacy profiles

### Candidate identity is not voter identity

A candidate may publish a real name, public number or profile. That does not
require publishing who voted for the candidate. Candidate metadata is separate
from the electorate credential and vote record.

### Profiles

| Profile | On-chain voter link | Intended use | Limitation |
|---|---|---|---|
| `PUBLIC` | signer address and choice | transparent councils/delegates | permanent voter-choice link |
| `PSEUDONYMOUS` | domain activity key and choice | low-sensitivity internal polls | transaction signer, timing and reward metadata may still identify the voter |
| `SECRET_ZK` | ballot nullifier and choice, no voter address in ballot state | secret member/person elections | requires production-qualified proof plus a submission path that does not expose the voter's signer; does not alone prevent coercion |
| `SEALED_SECRET` | encrypted/committed choice until tally | delayed-result secret ballots | requires a separately selected and audited threshold/decryption protocol |

Privacy is selected before opening and cannot be downgraded. `PUBLIC` must be
an explicit domain/ballot choice, never the silent default for political or
person elections.

Cosmos transactions expose their signer. A ZK proof inside a transaction signed
by the voter therefore does not create a secret ballot by itself. `SECRET_ZK`
requires a proof-authorized message that can be submitted by a non-voter
relayer/fee payer, plus analysis of timing, RPC and network metadata. The chain
must reject any attempt to treat the transport signer as the hidden voter.

### External identity numbers and credentials

Real-world onboarding may validate a person and issue an anonymous domain
credential. The chain stores only an unlinkable commitment or accepted
credential root. Names, government numbers and the mapping table stay off-chain
under a documented controller, retention and deletion policy. Hashing a stable
identity number without a secret salt is not anonymization and is forbidden.

If any party can reasonably reconnect a number to a person, it remains
pseudonymous personal data. The system must therefore support anonymous proof
of eligibility rather than depend on publishing the number.

### Ballot-scoped nullifiers

The future external-nullifier domain is equivalent to:

```text
H("TrueRepublic/ballot/v1", chain_id, domain, ballot_id, policy_hash)
```

The proof binds the exact choice or rating vector and policy hash. One identity
can vote once per ballot without producing a reusable cross-ballot identifier.
Nullifier inclusion in the result allows public duplicate-prevention auditing.

`SECRET_ZK` guarantees ballot-scoped voter/choice **unlinkability**, not that an
aggregate result can never reveal a person's choice. Its choices and aggregate
tallies are public: a singleton electorate, a unanimous small group or enough
votes known outside the protocol can make the remaining choice inferable. The
UI and policy review must warn about this residual leakage and may enforce a
domain-approved minimum electorate, but such a minimum cannot eliminate
auxiliary-information attacks. A domain requiring stronger result secrecy must
use a separately implemented and audited `SEALED_SECRET` protocol; it must not
label `SECRET_ZK` as providing that stronger guarantee. Tally queries and events
therefore publish only aggregates and nullifiers, never a voter mapping, while
making this limitation explicit.

### Rewards, receipts and coercion

GH-209 demonstrates atomic, signal-bound reward delivery for ratings, but a
direct payout address links reward and participation. Secret ballots should
default to no individual reward or a later unlinkable claim credential.

Public signatures and publishable proofs can act as receipts. ZK membership
privacy alone does not prevent a voter from revealing secrets, selling a vote
or being coerced. Commit-reveal also does not automatically solve these
problems, may reveal individual choices during the reveal phase and introduces
non-reveal behavior. A sealed profile therefore needs a separately selected
threshold-encryption or equivalent protocol. TrueRepublic must not claim
receipt-freeness or coercion resistance without a dedicated protocol and audit.

## 10. Client and Sovereign Alpha integration

`client-web` remains the Beta. After the chain engine exists, its registry gains
only the exact new message/query types and renders policy hashes, snapshot
height, privacy warnings, quorum/threshold math and final outcome. Unknown
profiles or policy versions block signing.

The Sovereign Alpha Go core owns ballot encoding, proof generation, signer
access and encrypted local receipts. Flutter renders the same canonical model;
messaging carries discussion only. The blockchain remains authoritative for
policy, eligibility commitment, vote acceptance and outcome.

Both clients must:

- show whether a surface is a discussion, advisory rating or formal ballot;
- show the immutable proposal hash before signing;
- distinguish candidate publicity from voter privacy;
- never imply that submission equals inclusion or passage;
- verify chain ID, domain, ballot ID and policy hash; and
- keep ZK submission disabled until the Phase 2 gates pass.

## 11. Security and legal boundary

On-chain finality can make policy, participation proofs and tallies replayable
and independently auditable. It cannot establish a person's civil identity,
organizational authority or the legal validity of an election.

Before any legally relied-upon use, the operator/domain needs a
jurisdiction-specific assessment of applicable statutes/bylaws, voter
eligibility, secrecy, accessibility, notice, recount/challenge, retention,
controller roles and lawful processing. Political opinions and linkable vote
records may be specially protected personal data. A data-protection impact
assessment is expected where risk is high.

No personally identifying data should be written to an immutable public chain
when a commitment, proof or off-chain reference can meet the requirement.

## 12. Staged implementation plan

GH-232 remains deferred until the existing rollout is stable enough for a
separately approved consensus change.

1. **B0 — approve specification:** ADR, legal/process matrix, enums, canonical
   encoding, integer math and migration plan.
2. **B1 — pure tally library:** no state changes; table, property and fuzz
   tests for every quorum/threshold/tie boundary and deterministic serialization.
3. **B2 — public Y/N/A engine:** store, lifecycle, snapshots, genesis,
   export/import/restart, signer messages/queries/events behind a disabled flag.
4. **B3 — person elections:** candidate validation, explicit no-winner results,
   runoff linkage and legacy-election retirement plan.
5. **B4 — systemic and hybrid ballots:** complete rating vectors, proposal
   hashes and two-stage state linkage.
6. **B5 — private profiles and clients:** ballot-scoped proof/nullifier circuit,
   ceremony artifacts, Beta/Alpha support and privacy-safe reward decision.
7. **B6 — qualification:** independent consensus/crypto/privacy review,
   multi-validator export/import/upgrade/recovery testing, private/public
   testnet and separately governed activation.

### Required verification matrix

- unit tables for every enum, state transition and exact arithmetic boundary;
- property tests for monotonic quorum/approval behavior and vote conservation;
- fuzz tests for messages, encodings, proofs, genesis and corrupted indexes;
- deterministic tests with reordered inputs and repeated EndBlock execution;
- signer/authority, replay, duplicate, revote, late-vote and cancellation tests;
- anonymous nullifier, wrong-chain/domain/ballot/policy/choice proof negatives;
- all-abstain, zero-vote, zero-denominator, tie and repeated-runoff cases;
- genesis export/import, same-home restart, upgrade/rollback and app-hash checks;
- multi-validator agreement and bounded-EndBlock load tests;
- Beta and Alpha golden transaction/query/error vectors;
- accessibility and privacy-warning UI checks; and
- independent security, cryptographic, privacy and applicable legal/process
  review before any binding deployment.

## 13. Normative legal/privacy references

These references inform the design boundary; they do not certify TrueRepublic
or replace jurisdiction-specific advice.

- [GDPR Article 4](https://eur-lex.europa.eu/eli/reg/2016/679/art_4/oj) —
  personal data, pseudonymisation and identifiability definitions.
- [GDPR Article 9](https://eur-lex.europa.eu/eli/reg/2016/679/art_9/oj) —
  special categories including data revealing political opinions.
- [EDPB Guidelines on processing of personal data through blockchain technologies (final version)](https://www.edpb.europa.eu/documents/guideline/guidelines-on-processing-of-personal-data-through-blockchain-technologies_en)
  — data protection by design, minimisation and immutable-ledger risks.
- [Council of Europe CM/Rec(2017)5](https://book.coe.int/en/legal-instruments/7609-standards-for-e-voting-recommendation-cmrec20175-guidelines-and-explanatory-memorandum.html)
  — legal, operational and technical standards for e-voting.
- [German Federal Constitutional Court, 2 BvC 3/07 and 2 BvC 4/07](https://www.bundesverfassungsgericht.de/SharedDocs/Entscheidungen/EN/2009/03/cs20090303_2bvc000307en.html)
  — public verifiability principles for state electronic elections.

## 14. Decisions still requiring bounded approval

- default quorum/threshold values and who may set domain defaults;
- whether Stage B reuses or refreshes the Stage A electorate;
- exact systemic missing-rating rule;
- public-ballot eligibility and warning policy;
- sealed-secret cryptographic protocol and recovery model;
- unlinkable participation rewards, if any;
- cancellation authority and optional pre-close cutoff within `PENDING`/`OPEN`;
  and
- migration/retirement treatment of historical `elecvote:` keys.

Until those decisions, implementation, audits and activation gates are
complete, systemic consensing and the current advisory election storage retain
their documented boundaries. This proposal creates no production ballot.
