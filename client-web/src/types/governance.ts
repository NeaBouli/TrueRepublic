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
