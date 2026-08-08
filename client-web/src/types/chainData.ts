export interface ChainCoin {
  denom: string;
  amount: string;
}

export interface ChainRating {
  value: number;
}

export interface ChainSuggestion {
  name: string;
  creator: string;
  stones: number;
  ratings: ChainRating[] | null;
  color: string;
  creation_date: number;
  external_link: string;
}

export interface ChainIssue {
  name: string;
  stones: number;
  suggestions: ChainSuggestion[] | null;
  creation_date: number;
  external_link: string;
}

export interface ChainDomain {
  name: string;
  admin: string;
  members: string[] | null;
  treasury: ChainCoin[] | null;
  issues: ChainIssue[] | null;
  permission_reg: string[] | null;
  identity_commits: string[] | null;
  merkle_root: string;
}

export interface ChainMerkleProof {
  root: string;
  commitment: string;
  path_indices: number[];
  path_elements: string[];
}

export interface ChainPayToPut {
  base_cost: string;
  domain_multiplier: number;
  final_cost: string;
  formula: string;
}
