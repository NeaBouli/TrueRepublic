import { create } from 'zustand';
import { MembershipService } from '@/services/membership';
import type { MembershipStatus, DomainInvite } from '@/types/membership';
import { DEFAULT_CHAIN } from '@/config/chains';

interface MembershipState {
  memberships: Record<string, MembershipStatus>;
  currentInvite: DomainInvite | null;
  isLoading: boolean;
  error: string | null;
}

interface MembershipStore extends MembershipState {
  parseInvite: (link: string) => boolean;
  loadMembership: (domainId: string, address: string) => Promise<void>;
  clearInvite: () => void;
}

const membershipService = new MembershipService(DEFAULT_CHAIN);

export const useMembershipStore = create<MembershipStore>((set) => ({
  memberships: {},
  currentInvite: null,
  isLoading: false,
  error: null,

  parseInvite: (link: string) => {
    const invite = membershipService.parseInviteLink(link);

    if (!invite) {
      set({ error: 'Invalid invite link' });
      return false;
    }

    set({ currentInvite: invite, error: null });
    return true;
  },

  loadMembership: async (domainId: string, address: string) => {
    try {
      set({ isLoading: true, error: null });

      const status = await membershipService.getMembershipStatus(
        domainId,
        address
      );

      if (!status) {
        set({ isLoading: false });
        return;
      }

      set((state) => ({
        memberships: {
          ...state.memberships,
          [domainId]: status,
        },
        isLoading: false,
      }));
    } catch (err: unknown) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load membership',
        isLoading: false,
      });
    }
  },

  clearInvite: () => {
    set({ currentInvite: null });
  },
}));
