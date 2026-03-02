import { create } from 'zustand';
import { DEXService } from '@/services/dex';
import type {
  Pool,
  RegisteredAsset,
  PoolStats,
  SwapEstimate,
} from '@/types/dex';
import { DEFAULT_CHAIN } from '@/config/chains';

interface DEXStore {
  // State
  pools: Pool[];
  assets: RegisteredAsset[];
  currentPool: Pool | null;
  currentPoolStats: PoolStats | null;
  swapEstimate: SwapEstimate | null;
  isLoading: boolean;
  error: string | null;

  // Actions
  loadPools: () => Promise<void>;
  loadAssets: () => Promise<void>;
  selectPool: (assetDenom: string) => Promise<void>;
  loadPoolStats: (assetDenom: string) => Promise<void>;
  estimateSwap: (
    inputDenom: string,
    inputAmount: string,
    outputDenom: string
  ) => Promise<void>;
  clearEstimate: () => void;
  clearSelection: () => void;
}

const dexService = new DEXService(DEFAULT_CHAIN);

export const useDEXStore = create<DEXStore>((set) => ({
  // State
  pools: [],
  assets: [],
  currentPool: null,
  currentPoolStats: null,
  swapEstimate: null,
  isLoading: false,
  error: null,

  // Actions
  loadPools: async () => {
    try {
      set({ isLoading: true, error: null });
      const pools = await dexService.listPools();
      set({ pools, isLoading: false });
    } catch (error: unknown) {
      const message =
        error instanceof Error ? error.message : 'Failed to load pools';
      set({ error: message, isLoading: false });
    }
  },

  loadAssets: async () => {
    try {
      set({ isLoading: true, error: null });
      const assets = await dexService.listAssets();
      set({ assets, isLoading: false });
    } catch (error: unknown) {
      const message =
        error instanceof Error ? error.message : 'Failed to load assets';
      set({ error: message, isLoading: false });
    }
  },

  selectPool: async (assetDenom: string) => {
    try {
      set({ isLoading: true, error: null });
      const pool = await dexService.getPool(assetDenom);

      if (!pool) {
        throw new Error('Pool not found');
      }

      set({ currentPool: pool, isLoading: false });

      // Auto-load pool stats
      const stats = await dexService.getPoolStats(assetDenom);
      set({ currentPoolStats: stats });
    } catch (error: unknown) {
      const message =
        error instanceof Error ? error.message : 'Failed to select pool';
      set({ error: message, isLoading: false });
    }
  },

  loadPoolStats: async (assetDenom: string) => {
    try {
      set({ isLoading: true, error: null });
      const stats = await dexService.getPoolStats(assetDenom);
      set({ currentPoolStats: stats, isLoading: false });
    } catch (error: unknown) {
      const message =
        error instanceof Error ? error.message : 'Failed to load pool stats';
      set({ error: message, isLoading: false });
    }
  },

  estimateSwap: async (
    inputDenom: string,
    inputAmount: string,
    outputDenom: string
  ) => {
    try {
      set({ isLoading: true, error: null });
      const estimate = await dexService.estimateSwap(
        inputDenom,
        inputAmount,
        outputDenom
      );
      set({ swapEstimate: estimate, isLoading: false });
    } catch (error: unknown) {
      const message =
        error instanceof Error ? error.message : 'Failed to estimate swap';
      set({ error: message, isLoading: false });
    }
  },

  clearEstimate: () => {
    set({ swapEstimate: null });
  },

  clearSelection: () => {
    set({
      currentPool: null,
      currentPoolStats: null,
      swapEstimate: null,
    });
  },
}));
