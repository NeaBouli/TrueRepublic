import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { Identity } from '@/types/zkp';
import { ZKPService } from '@/services/zkp';
import { DEFAULT_CHAIN } from '@/config/chains';

interface IdentityStore {
  // State
  identity: Identity | null;
  hasIdentity: boolean;
  isInitialized: boolean;

  // Actions
  createIdentity: () => void;
  loadIdentity: () => void;
  clearIdentity: () => void;
  exportIdentity: () => string | null;
  importIdentity: (exported: string) => void;
}

const zkpService = new ZKPService(DEFAULT_CHAIN);

export const useIdentityStore = create<IdentityStore>()(
  persist(
    (set, get) => ({
      // State
      identity: null,
      hasIdentity: false,
      isInitialized: false,

      // Actions
      createIdentity: () => {
        const identity = zkpService.generateIdentity();
        set({
          identity,
          hasIdentity: true,
          isInitialized: true,
        });
      },

      loadIdentity: () => {
        const { identity } = get();
        set({
          hasIdentity: !!identity,
          isInitialized: true,
        });
      },

      clearIdentity: () => {
        set({
          identity: null,
          hasIdentity: false,
        });
      },

      exportIdentity: () => {
        const { identity } = get();
        if (!identity) return null;
        return JSON.stringify(identity);
      },

      importIdentity: (exported: string) => {
        const identity = JSON.parse(exported) as Identity;

        if (!identity.secret || !identity.commitment) {
          throw new Error('Invalid identity data');
        }

        set({
          identity,
          hasIdentity: true,
          isInitialized: true,
        });
      },
    }),
    {
      name: 'identity-store',
      partialize: (state) => ({
        identity: state.identity,
      }),
    }
  )
);
