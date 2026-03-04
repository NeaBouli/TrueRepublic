import { create } from 'zustand';
import { NetworkService } from '@/services/network';
import type {
  NetworkInfo,
  Validator,
  Block,
  IBCChannel,
} from '@/types/network';
import { DEFAULT_CHAIN } from '@/config/chains';

interface NetworkState {
  networkInfo: NetworkInfo | null;
  validators: Validator[];
  recentBlocks: Block[];
  ibcChannels: IBCChannel[];
  isLoading: boolean;
  error: string | null;
}

interface NetworkStore extends NetworkState {
  loadNetworkInfo: () => Promise<void>;
  loadValidators: () => Promise<void>;
  loadRecentBlocks: () => Promise<void>;
  loadIBCChannels: () => Promise<void>;
  loadAll: () => Promise<void>;
}

const networkService = new NetworkService(DEFAULT_CHAIN);

export const useNetworkStore = create<NetworkStore>((set, get) => ({
  networkInfo: null,
  validators: [],
  recentBlocks: [],
  ibcChannels: [],
  isLoading: false,
  error: null,

  loadNetworkInfo: async () => {
    try {
      const info = await networkService.getNetworkInfo();
      set({ networkInfo: info });
    } catch (err: unknown) {
      set({
        error:
          err instanceof Error ? err.message : 'Failed to load network info',
      });
    }
  },

  loadValidators: async () => {
    try {
      const validators = await networkService.getValidators();
      set({ validators });
    } catch (err: unknown) {
      set({
        error:
          err instanceof Error ? err.message : 'Failed to load validators',
      });
    }
  },

  loadRecentBlocks: async () => {
    try {
      const blocks = await networkService.getRecentBlocks(10);
      set({ recentBlocks: blocks });
    } catch (err: unknown) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load blocks',
      });
    }
  },

  loadIBCChannels: async () => {
    try {
      const channels = await networkService.getIBCChannels();
      set({ ibcChannels: channels });
    } catch (err: unknown) {
      set({
        error:
          err instanceof Error ? err.message : 'Failed to load IBC channels',
      });
    }
  },

  loadAll: async () => {
    set({ isLoading: true, error: null });
    await Promise.all([
      get().loadNetworkInfo(),
      get().loadValidators(),
      get().loadRecentBlocks(),
      get().loadIBCChannels(),
    ]);
    set({ isLoading: false });
  },
}));
