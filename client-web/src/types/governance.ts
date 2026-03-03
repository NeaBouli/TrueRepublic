export interface Domain {
  domainId: string;
  name: string;
  treasury: string;
  memberCount: number;
  createdAt: string;
}

export interface Issue {
  issueId: string;
  domainId: string;
  title: string;
  description: string;
  createdAt: string;
  status: 'active' | 'closed';
}

export interface Suggestion {
  suggestionId: string;
  issueId: string;
  domainId: string;
  title: string;
  description: string;
  creator: string;
  avgRating: number;
  ratingCount: number;
  greenStones: number;
  yellowStones: number;
  redStones: number;
  zone: 'green' | 'yellow' | 'red' | 'unzoned';
  createdAt: string;
}

export interface RatingStats {
  suggestionId: string;
  avgRating: number;
  count: number;
  distribution: Record<number, number>;
}

/**
 * Parameters for MsgSubmitProposal (Go: submit_proposal).
 * Creates a suggestion under an issue (creates the issue too if it doesn't exist).
 */
export interface CreateSuggestionParams {
  domain_name: string;
  issue_name: string;
  suggestion_name: string;
  creator: string;
  fee: { denom: string; amount: string }[];
  external_link?: string;
}

/**
 * Parameters for MsgPlaceStoneOnSuggestion (Go: place_stone_suggestion).
 * Stones are support/endorsement counts — no color field in the Go message.
 */
export interface PlaceStoneOnSuggestionParams {
  domain_name: string;
  issue_name: string;
  suggestion_name: string;
  member_addr: string;
}

/**
 * Parameters for MsgPlaceStoneOnIssue (Go: place_stone_issue).
 */
export interface PlaceStoneOnIssueParams {
  domain_name: string;
  issue_name: string;
  member_addr: string;
}

export interface PayToPutCalculation {
  baseCost: string;
  domainMultiplier: number;
  finalCost: string;
  formula: string;
}

export interface Election {
  electionId: string;
  domainId: string;
  title: string;
  description: string;
  candidates: string[];
  startTime: string;
  endTime: string;
  status: 'upcoming' | 'active' | 'ended';
  totalVotes: number;
}
