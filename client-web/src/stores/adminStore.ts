import { create } from 'zustand';
import { AdminService } from '@/services/admin';
import type { DomainMember, DomainStats } from '@/types/admin';
import { DEFAULT_CHAIN } from '@/config/chains';

interface AdminState {
  isAdmin: Record<string, boolean>;
  domainMembers: Record<string, DomainMember[]>;
  domainStats: Record<string, DomainStats>;
  isLoading: boolean;
  error: string | null;
}

interface AdminStore extends AdminState {
  checkAdmin: (domainName: string, address: string) => Promise<void>;
  loadDomainMembers: (domainName: string) => Promise<void>;
  loadDomainStats: (domainName: string) => Promise<void>;
}

const adminService = new AdminService(DEFAULT_CHAIN);

export const useAdminStore = create<AdminStore>((set) => ({
  isAdmin: {},
  domainMembers: {},
  domainStats: {},
  isLoading: false,
  error: null,

  checkAdmin: async (domainName: string, address: string) => {
    try {
      const result = await adminService.isAdmin(domainName, address);
      set((state) => ({
        isAdmin: { ...state.isAdmin, [domainName]: result },
      }));
    } catch (err: unknown) {
      set({
        error: err instanceof Error ? err.message : 'Failed to check admin',
      });
    }
  },

  loadDomainMembers: async (domainName: string) => {
    try {
      set({ isLoading: true, error: null });
      const members = await adminService.getDomainMembers(domainName);
      set((state) => ({
        domainMembers: { ...state.domainMembers, [domainName]: members },
        isLoading: false,
      }));
    } catch (err: unknown) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load members',
        isLoading: false,
      });
    }
  },

  loadDomainStats: async (domainName: string) => {
    try {
      set({ isLoading: true, error: null });
      const stats = await adminService.getDomainStats(domainName);
      if (stats) {
        set((state) => ({
          domainStats: { ...state.domainStats, [domainName]: stats },
          isLoading: false,
        }));
      } else {
        set({ isLoading: false });
      }
    } catch (err: unknown) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load stats',
        isLoading: false,
      });
    }
  },
}));
