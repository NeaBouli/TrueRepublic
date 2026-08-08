import type { ChainConfig } from '@/types/chain';
import type { Domain, Issue, Suggestion, RatingStats } from '@/types/governance';
import type {
  ChainCoin,
  ChainDomain,
  ChainIssue,
  ChainSuggestion,
} from '@/types/chainData';
import {
  expectChainDomain,
  expectQueryArray,
  ModuleQueryClient,
  QUERY_PATHS,
} from './moduleQuery';

function timestamp(seconds: number): string {
  if (seconds <= 0) return '';
  const date = new Date(seconds * 1_000);
  return Number.isNaN(date.getTime()) ? '' : date.toISOString();
}

function treasuryAmount(coins: ChainCoin[] | null, denom: string): string {
  return coins?.find((coin) => coin.denom === denom)?.amount ?? '0';
}

function mapDomain(domain: ChainDomain, denom: string): Domain {
  return {
    domainId: domain.name,
    name: domain.name,
    treasury: treasuryAmount(domain.treasury, denom),
    memberCount: domain.members?.length ?? 0,
    createdAt: '',
  };
}

function mapIssue(domainId: string, issue: ChainIssue): Issue {
  return {
    issueId: issue.name,
    domainId,
    title: issue.name,
    description: issue.external_link || '',
    createdAt: timestamp(issue.creation_date),
    status: 'active',
  };
}

function ratingStats(suggestion: ChainSuggestion): RatingStats {
  const ratings = suggestion.ratings ?? [];
  const distribution: Record<number, number> = {};
  let sum = 0;
  for (const rating of ratings) {
    sum += rating.value;
    distribution[rating.value] = (distribution[rating.value] ?? 0) + 1;
  }
  return {
    suggestionId: suggestion.name,
    avgRating: ratings.length === 0 ? 0 : sum / ratings.length,
    count: ratings.length,
    distribution,
  };
}

function mapSuggestion(
  domainId: string,
  issueId: string,
  suggestion: ChainSuggestion
): Suggestion {
  const stats = ratingStats(suggestion);
  const zone = ['green', 'yellow', 'red'].includes(suggestion.color)
    ? (suggestion.color as 'green' | 'yellow' | 'red')
    : 'unzoned';
  return {
    suggestionId: suggestion.name,
    issueId,
    domainId,
    title: suggestion.name,
    description: suggestion.external_link || '',
    creator: suggestion.creator,
    avgRating: stats.avgRating,
    ratingCount: stats.count,
    greenStones: zone === 'green' ? suggestion.stones : 0,
    yellowStones: zone === 'yellow' ? suggestion.stones : 0,
    redStones: zone === 'red' ? suggestion.stones : 0,
    zone,
    createdAt: timestamp(suggestion.creation_date),
  };
}

export class GovernanceService {
  private readonly queries: ModuleQueryClient;

  constructor(
    private readonly config: ChainConfig,
    queries = new ModuleQueryClient(config)
  ) {
    this.queries = queries;
  }

  async listDomains(): Promise<Domain[]> {
    const result = await this.queries.query<unknown>(
      QUERY_PATHS.truedemocracy.domains
    );
    const domains = expectQueryArray<unknown>(
      QUERY_PATHS.truedemocracy.domains,
      result
    ).map((value) =>
      this.validateDomain(value, QUERY_PATHS.truedemocracy.domains)
    );
    return domains.map((domain) =>
      mapDomain(domain, this.config.coinMinimalDenom)
    );
  }

  async getDomain(domainId: string): Promise<Domain> {
    const domain = await this.getChainDomain(domainId);
    return mapDomain(domain, this.config.coinMinimalDenom);
  }

  async listIssues(domainId: string): Promise<Issue[]> {
    const domain = await this.getChainDomain(domainId);
    return (domain.issues ?? []).map((issue) => mapIssue(domainId, issue));
  }

  async getIssue(domainId: string, issueId: string): Promise<Issue | null> {
    const domain = await this.getChainDomain(domainId);
    const issue = (domain.issues ?? []).find((entry) => entry.name === issueId);
    return issue ? mapIssue(domainId, issue) : null;
  }

  async listSuggestions(
    domainId: string,
    issueId: string
  ): Promise<Suggestion[]> {
    const domain = await this.getChainDomain(domainId);
    const issue = (domain.issues ?? []).find((entry) => entry.name === issueId);
    return (issue?.suggestions ?? []).map((suggestion) =>
      mapSuggestion(domainId, issueId, suggestion)
    );
  }

  async getSuggestion(
    domainId: string,
    suggestionId: string
  ): Promise<Suggestion | null> {
    const domain = await this.getChainDomain(domainId);
    for (const issue of domain.issues ?? []) {
      const suggestion = (issue.suggestions ?? []).find(
        (entry) => entry.name === suggestionId
      );
      if (suggestion) return mapSuggestion(domainId, issue.name, suggestion);
    }
    return null;
  }

  async getRatingStats(
    domainId: string,
    suggestionId: string
  ): Promise<RatingStats | null> {
    const domain = await this.getChainDomain(domainId);
    for (const issue of domain.issues ?? []) {
      const suggestion = (issue.suggestions ?? []).find(
        (entry) => entry.name === suggestionId
      );
      if (suggestion) return ratingStats(suggestion);
    }
    return null;
  }

  private getChainDomain(domainId: string): Promise<ChainDomain> {
    return this.queries
      .query<unknown>(QUERY_PATHS.truedemocracy.domain, [
        { number: 1, type: 'string', value: domainId },
      ])
      .then((value) => this.validateDomain(value));
  }

  private validateDomain(
    value: unknown,
    path: string = QUERY_PATHS.truedemocracy.domain
  ): ChainDomain {
    return expectChainDomain(path, value);
  }
}
