import type { ChainConfig } from '@/types/chain';
import type { Domain, Issue, Suggestion, RatingStats } from '@/types/governance';

export class GovernanceService {
  private config: ChainConfig;

  constructor(config: ChainConfig) {
    this.config = config;
  }

  /**
   * List all domains
   */
  async listDomains(): Promise<Domain[]> {
    try {
      const response = await fetch(
        `${this.config.rest}/truerepublic/truedemocracy/domains`
      );

      if (!response.ok) {
        throw new Error('Failed to fetch domains');
      }

      const data = await response.json();
      return data.domains || [];
    } catch {
      return [];
    }
  }

  /**
   * Get domain by ID
   */
  async getDomain(domainId: string): Promise<Domain | null> {
    try {
      const response = await fetch(
        `${this.config.rest}/truerepublic/truedemocracy/domain/${domainId}`
      );

      if (!response.ok) return null;

      const data = await response.json();
      return data.domain || null;
    } catch {
      return null;
    }
  }

  /**
   * List issues for domain
   */
  async listIssues(domainId: string): Promise<Issue[]> {
    try {
      const response = await fetch(
        `${this.config.rest}/truerepublic/truedemocracy/issues/${domainId}`
      );

      if (!response.ok) return [];

      const data = await response.json();
      return data.issues || [];
    } catch {
      return [];
    }
  }

  /**
   * Get issue by ID
   */
  async getIssue(domainId: string, issueId: string): Promise<Issue | null> {
    try {
      const response = await fetch(
        `${this.config.rest}/truerepublic/truedemocracy/issue/${domainId}/${issueId}`
      );

      if (!response.ok) return null;

      const data = await response.json();
      return data.issue || null;
    } catch {
      return null;
    }
  }

  /**
   * List suggestions for issue
   */
  async listSuggestions(
    domainId: string,
    issueId: string
  ): Promise<Suggestion[]> {
    try {
      const response = await fetch(
        `${this.config.rest}/truerepublic/truedemocracy/suggestions/${domainId}/${issueId}`
      );

      if (!response.ok) return [];

      const data = await response.json();
      return data.suggestions || [];
    } catch {
      return [];
    }
  }

  /**
   * Get suggestion by ID
   */
  async getSuggestion(
    domainId: string,
    suggestionId: string
  ): Promise<Suggestion | null> {
    try {
      const response = await fetch(
        `${this.config.rest}/truerepublic/truedemocracy/suggestion/${domainId}/${suggestionId}`
      );

      if (!response.ok) return null;

      const data = await response.json();
      return data.suggestion || null;
    } catch {
      return null;
    }
  }

  /**
   * Get rating stats for suggestion
   */
  async getRatingStats(
    domainId: string,
    suggestionId: string
  ): Promise<RatingStats | null> {
    try {
      const response = await fetch(
        `${this.config.rest}/truerepublic/truedemocracy/ratings/${domainId}/${suggestionId}`
      );

      if (!response.ok) return null;

      const data = await response.json();
      return data.stats || null;
    } catch {
      return null;
    }
  }
}
