import { create } from 'zustand';
import { GovernanceService } from '@/services/governance';
import type { Domain, Issue, Suggestion } from '@/types/governance';
import { DEFAULT_CHAIN } from '@/config/chains';

interface GovernanceStore {
  // State
  domains: Domain[];
  currentDomain: Domain | null;
  issues: Issue[];
  currentIssue: Issue | null;
  suggestions: Suggestion[];
  currentSuggestion: Suggestion | null;
  isLoading: boolean;
  error: string | null;

  // Actions
  loadDomains: () => Promise<void>;
  selectDomain: (domainId: string) => Promise<void>;
  loadIssues: (domainId: string) => Promise<void>;
  selectIssue: (domainId: string, issueId: string) => Promise<void>;
  loadSuggestions: (domainId: string, issueId: string) => Promise<void>;
  selectSuggestion: (domainId: string, suggestionId: string) => Promise<void>;
  clearSelection: () => void;
}

const governanceService = new GovernanceService(DEFAULT_CHAIN);

export const useGovernanceStore = create<GovernanceStore>((set, get) => ({
  // State
  domains: [],
  currentDomain: null,
  issues: [],
  currentIssue: null,
  suggestions: [],
  currentSuggestion: null,
  isLoading: false,
  error: null,

  // Actions
  loadDomains: async () => {
    try {
      set({ isLoading: true, error: null });
      const domains = await governanceService.listDomains();
      set({ domains, isLoading: false });
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Failed to load domains';
      set({ error: message, isLoading: false });
    }
  },

  selectDomain: async (domainId: string) => {
    try {
      set({ isLoading: true, error: null });
      const domain = await governanceService.getDomain(domainId);

      if (!domain) {
        throw new Error('Domain not found');
      }

      set({ currentDomain: domain, isLoading: false });
      await get().loadIssues(domainId);
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Failed to select domain';
      set({ error: message, isLoading: false });
    }
  },

  loadIssues: async (domainId: string) => {
    try {
      set({ isLoading: true, error: null });
      const issues = await governanceService.listIssues(domainId);
      set({ issues, isLoading: false });
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Failed to load issues';
      set({ error: message, isLoading: false });
    }
  },

  selectIssue: async (domainId: string, issueId: string) => {
    try {
      set({ isLoading: true, error: null });
      const issue = await governanceService.getIssue(domainId, issueId);

      if (!issue) {
        throw new Error('Issue not found');
      }

      set({ currentIssue: issue, isLoading: false });
      await get().loadSuggestions(domainId, issueId);
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Failed to select issue';
      set({ error: message, isLoading: false });
    }
  },

  loadSuggestions: async (domainId: string, issueId: string) => {
    try {
      set({ isLoading: true, error: null });
      const suggestions = await governanceService.listSuggestions(domainId, issueId);
      set({ suggestions, isLoading: false });
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Failed to load suggestions';
      set({ error: message, isLoading: false });
    }
  },

  selectSuggestion: async (domainId: string, suggestionId: string) => {
    try {
      set({ isLoading: true, error: null });
      const suggestion = await governanceService.getSuggestion(domainId, suggestionId);

      if (!suggestion) {
        throw new Error('Suggestion not found');
      }

      set({ currentSuggestion: suggestion, isLoading: false });
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Failed to select suggestion';
      set({ error: message, isLoading: false });
    }
  },

  clearSelection: () => {
    set({
      currentDomain: null,
      currentIssue: null,
      currentSuggestion: null,
      issues: [],
      suggestions: [],
    });
  },
}));
