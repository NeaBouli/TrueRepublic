package truedemocracy

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Person election voting uses KV keys:
//   "elecvote:{domainName}:{issueName}:{voterAddr}" → candidateName (or "ABSTAIN")
//
// This implements Whitepaper §3.7: voting modes for person elections.

const abstainSentinel = "ABSTAIN"

func electionVoteKey(domainName, issueName, voterAddr string) []byte {
	return []byte("elecvote:" + domainName + ":" + issueName + ":" + voterAddr)
}

// CastElectionVote records a vote (approve a candidate or abstain) in a person
// election. Each member can vote for exactly one candidate per issue, or abstain.
func (k Keeper) CastElectionVote(ctx sdk.Context, domainName, issueName, candidateName, voterAddr string, choice VoteChoice) error {
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "domain %s not found", domainName)
	}

	if !isMember(domain, voterAddr) {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "only domain members can vote")
	}

	issueIdx := findIssueIndex(domain, issueName)
	if issueIdx == -1 {
		return errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "issue not found")
	}

	store := ctx.KVStore(k.StoreKey)
	key := electionVoteKey(domainName, issueName, voterAddr)

	switch choice {
	case VoteChoiceAbstain:
		if !domain.Options.AbstentionAllowed {
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "abstention is not allowed in this domain")
		}
		store.Set(key, []byte(abstainSentinel))

	case VoteChoiceApprove:
		if candidateName == "" {
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "candidate name required for approve vote")
		}
		// Candidate must be a suggestion in the issue's suggestion list.
		if findSuggestionIndex(domain.Issues[issueIdx], candidateName) == -1 {
			return errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "candidate not found in suggestion list")
		}
		store.Set(key, []byte(candidateName))

	default:
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid vote choice")
	}

	// Update issue activity timestamp.
	domain.Issues[issueIdx].LastActivityAt = ctx.BlockTime().Unix()
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	store.Set([]byte("domain:"+domainName), bz)
	return nil
}

// ElectionResult holds the outcome of a person election tally.
type ElectionResult struct {
	Candidate string // winning candidate (empty if no winner)
	Votes     int    // votes received by winner
	Total     int    // total votes cast (excl. abstentions for simple majority)
	Abstained int    // number of explicit abstentions
	Elected   bool   // whether a winner meets the threshold
}

// TallyElection evaluates an election for the given issue according to the
// domain's VotingMode (WP §3.7). For VotingModeSystemicConsensing, the
// standard rating-based scoring in §3.2 applies and this function is not used.
func (k Keeper) TallyElection(ctx sdk.Context, domainName, issueName string) (ElectionResult, error) {
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return ElectionResult{}, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "domain %s not found", domainName)
	}

	issueIdx := findIssueIndex(domain, issueName)
	if issueIdx == -1 {
		return ElectionResult{}, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "issue not found")
	}

	store := ctx.KVStore(k.StoreKey)

	// Count votes per candidate and abstentions.
	candidateVotes := make(map[string]int)
	abstained := 0
	totalVoters := 0

	for _, member := range domain.Members {
		bz := store.Get(electionVoteKey(domainName, issueName, member))
		if bz == nil {
			continue // did not vote
		}
		totalVoters++
		vote := string(bz)
		if vote == abstainSentinel {
			abstained++
		} else {
			candidateVotes[vote]++
		}
	}

	// Find candidate with most votes.
	bestCandidate := ""
	bestVotes := 0
	for candidate, votes := range candidateVotes {
		if votes > bestVotes {
			bestVotes = votes
			bestCandidate = candidate
		}
	}

	result := ElectionResult{
		Candidate: bestCandidate,
		Votes:     bestVotes,
		Abstained: abstained,
	}

	switch domain.Options.VotingMode {
	case VotingModeSimpleMajority:
		// >50% of votes cast, excluding abstentions.
		effectiveVotes := totalVoters - abstained
		result.Total = effectiveVotes
		if effectiveVotes > 0 && bestVotes*2 > effectiveVotes {
			result.Elected = true
		}

	case VotingModeAbsoluteMajority:
		// >50% of all eligible members.
		totalMembers := len(domain.Members)
		result.Total = totalMembers
		if totalMembers > 0 && bestVotes*2 > totalMembers {
			result.Elected = true
		}

	case VotingModeSystemicConsensing:
		// Systemic consensing uses the rating mechanism (§3.2), not this tally.
		// Return the stone-based leader as a fallback.
		result.Total = totalVoters - abstained
		result.Elected = bestVotes > 0

	default:
		// Default to simple majority.
		effectiveVotes := totalVoters - abstained
		result.Total = effectiveVotes
		if effectiveVotes > 0 && bestVotes*2 > effectiveVotes {
			result.Elected = true
		}
	}

	return result, nil
}
