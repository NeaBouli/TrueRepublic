import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { Wallet, Balance } from '@/types/wallet';
import type {
  SendParams,
  SubmittedTransaction,
  TransactionHistoryFailure,
  TransactionResult,
} from '@/types/transaction';
import { WalletService } from '@/services/wallet';
import { DEFAULT_CHAIN } from '@/config/chains';

interface WalletStore {
  // State
  currentWallet: Wallet | null;
  wallets: Wallet[];
  balances: Balance[];
  isLocked: boolean;
  password: string | null;
  isLoading: boolean;
  error: string | null;

  // Submitted-transaction history state (GH-131)
  historyTransactions: SubmittedTransaction[];
  historyStatus: 'idle' | 'loading' | 'ready' | 'error';
  historyFailure: TransactionHistoryFailure | null;
  historyError: string | null;
  historyPage: number;
  historyPageSize: number;
  historyTotal: number;
  historyHasMore: boolean;
  historyAddress: string | null;
  historyGeneration: number;

  // Actions
  createWallet: (name: string, password: string) => Promise<Wallet>;
  importWallet: (name: string, mnemonic: string, password: string) => Promise<Wallet>;
  switchWallet: (address: string, password: string) => Promise<void>;
  deleteWallet: (address: string) => void;
  lock: () => void;
  unlock: (password: string) => Promise<void>;
  refreshBalance: () => Promise<void>;
  loadWallets: () => void;
  getWallet: (address: string, password: string) => Promise<Wallet>;
  sendTokens: (params: SendParams) => Promise<TransactionResult>;
  loadHistoryPage: (page: number) => Promise<void>;
  refreshHistory: () => Promise<void>;
  nextHistoryPage: () => Promise<void>;
  prevHistoryPage: () => Promise<void>;
  clearHistory: () => void;
}

let blockchainService: Promise<import('@/services/blockchain').BlockchainService> | null = null;
let transactionModule: Promise<typeof import('@/services/transaction')> | null = null;
let transactionService: Promise<import('@/services/transaction').TransactionService> | null = null;
let historyAbortController: AbortController | null = null;

function getBlockchainService() {
  if (!blockchainService) {
    blockchainService = import('@/services/blockchain').then(
      ({ BlockchainService }) => new BlockchainService(DEFAULT_CHAIN)
    );
  }
  return blockchainService;
}

function getTransactionModule() {
  if (!transactionModule) {
    transactionModule = import('@/services/transaction').catch((error) => {
      transactionModule = null;
      throw error;
    });
  }
  return transactionModule;
}

function getTransactionService() {
  if (!transactionService) {
    transactionService = getTransactionModule().then(
      ({ TransactionService }) => new TransactionService(DEFAULT_CHAIN)
    ).catch((error) => {
      transactionService = null;
      throw error;
    });
  }
  return transactionService;
}

function historyFailureFrom(error: unknown): TransactionHistoryFailure {
  if (
    error instanceof Error &&
    error.name === 'TransactionHistoryError' &&
    'failure' in error &&
    (error.failure === 'unavailable' ||
      error.failure === 'timeout' ||
      error.failure === 'protocol' ||
      error.failure === 'decode')
  ) {
    return error.failure;
  }
  return 'unavailable';
}

const emptyHistory = {
  historyTransactions: [] as SubmittedTransaction[],
  historyStatus: 'idle' as const,
  historyFailure: null,
  historyError: null,
  historyPage: 1,
  historyPageSize: 20,
  historyTotal: 0,
  historyHasMore: false,
  historyAddress: null,
};

export const useWalletStore = create<WalletStore>()(
  persist(
    (set, get) => ({
      // State
      currentWallet: null,
      wallets: [],
      balances: [],
      isLocked: true,
      password: null,
      isLoading: false,
      error: null,

      // Submitted-transaction history state
      ...emptyHistory,
      historyGeneration: 0,

      // Actions
      createWallet: async (name: string, password: string) => {
        set({ isLoading: true, error: null });

        try {
          const wallet = await WalletService.createWallet({ name, password });

          get().clearHistory();

          set((state) => ({
            wallets: [...state.wallets, { ...wallet, mnemonic: undefined }],
            currentWallet: { ...wallet, mnemonic: undefined },
            password,
            isLocked: false,
            isLoading: false,
          }));

          get().refreshBalance();
          return wallet;
        } catch (error: unknown) {
          const message = error instanceof Error ? error.message : 'Failed to create wallet';
          set({ error: message, isLoading: false });
          throw error;
        }
      },

      importWallet: async (name: string, mnemonic: string, password: string) => {
        set({ isLoading: true, error: null });

        try {
          const wallet = await WalletService.importWallet({ name, mnemonic, password });

          get().clearHistory();

          set((state) => ({
            wallets: [...state.wallets, { ...wallet, mnemonic: undefined }],
            currentWallet: { ...wallet, mnemonic: undefined },
            password,
            isLocked: false,
            isLoading: false,
          }));

          get().refreshBalance();
          return wallet;
        } catch (error: unknown) {
          const message = error instanceof Error ? error.message : 'Failed to import wallet';
          set({ error: message, isLoading: false });
          throw error;
        }
      },

      switchWallet: async (address: string, password: string) => {
        set({ isLoading: true, error: null });

        try {
          const wallet = await WalletService.getWallet(address, password);

          // A different unlocked wallet must never show the previous wallet's
          // history, even while the next page is still loading.
          get().clearHistory();

          set({
            currentWallet: { ...wallet, mnemonic: undefined },
            password,
            isLocked: false,
            isLoading: false,
          });

          get().refreshBalance();
        } catch (error: unknown) {
          const message = error instanceof Error ? error.message : 'Failed to switch wallet';
          set({ error: message, isLoading: false });
          throw error;
        }
      },

      deleteWallet: (address: string) => {
        WalletService.deleteWallet(address);
        const removingCurrent = get().currentWallet?.address === address;
        if (removingCurrent) get().clearHistory();
        set((state) => ({
          wallets: state.wallets.filter((w) => w.address !== address),
          currentWallet:
            state.currentWallet?.address === address ? null : state.currentWallet,
        }));
      },

      lock: () => {
        get().clearHistory();
        set({
          isLocked: true,
          password: null,
          currentWallet: null,
          balances: [],
        });
      },

      unlock: async (password: string) => {
        const { wallets } = get();
        if (wallets.length === 0) {
          throw new Error('No wallets found');
        }

        const firstWallet = wallets[0];
        await get().switchWallet(firstWallet.address, password);
      },

      refreshBalance: async () => {
        const { currentWallet } = get();
        if (!currentWallet) return;

        try {
          const service = await getBlockchainService();
          const balances = await service.getBalance(currentWallet.address);
          set({ balances });
        } catch {
          // Balance refresh is best-effort; node may be offline
        }
      },

      loadWallets: () => {
        const wallets = WalletService.loadWallets().map((w) => ({
          ...w,
          mnemonic: undefined,
        }));
        set({ wallets });
      },

      getWallet: async (address: string, password: string) => {
        return WalletService.getWallet(address, password);
      },

      sendTokens: async (params: SendParams) => {
        const { currentWallet, password } = get();
        if (!currentWallet || !password) {
          throw new Error('Wallet not unlocked');
        }

        set({ isLoading: true, error: null });

        try {
          const signingWallet = await WalletService.getWalletForSigning(
            currentWallet.address,
            password
          );

          const service = await getTransactionService();
          const result = await service.send(signingWallet, params);

          if (!result.success) {
            throw new Error(result.error || 'Transaction failed');
          }

          await get().refreshBalance();

          // A committed send must become visible in the submitted history;
          // failures above never reach this point, so no fake row can appear.
          if (get().currentWallet?.address === currentWallet.address) {
            // History is secondary evidence. A transient query/chunk failure
            // must never turn an already committed send into a reported send
            // failure.
            await get().loadHistoryPage(1).catch(() => undefined);
          }

          set({ isLoading: false });
          return result;
        } catch (error: unknown) {
          const message = error instanceof Error ? error.message : 'Transaction failed';
          set({ error: message, isLoading: false });
          throw error;
        }
      },

      loadHistoryPage: async (page: number) => {
        const { currentWallet, isLocked } = get();
        if (!currentWallet || isLocked) return;

        const address = currentWallet.address;
        const requestedPage = Number.isSafeInteger(page) && page >= 1 ? page : 1;
        historyAbortController?.abort();
        const controller = new AbortController();
        historyAbortController = controller;
        const generation = get().historyGeneration + 1;
        set({
          historyGeneration: generation,
          historyStatus: 'loading',
          historyFailure: null,
          historyError: null,
          historyTransactions: [],
          historyPage: requestedPage,
          historyTotal: 0,
          historyHasMore: false,
          historyAddress: address,
        });

        const isStale = () => {
          const state = get();
          return (
            state.historyGeneration !== generation ||
            state.isLocked ||
            state.currentWallet?.address !== address
          );
        };

        try {
          const service = await getTransactionService();
          const result = await service.getSubmittedTransactions(
            address,
            requestedPage,
            undefined,
            controller.signal
          );
          if (isStale()) return;

          set({
            historyTransactions: result.transactions,
            historyStatus: 'ready',
            historyFailure: null,
            historyError: null,
            historyPage: result.page,
            historyPageSize: result.pageSize,
            historyTotal: result.total,
            historyHasMore: result.hasMore,
            historyAddress: address,
          });
        } catch (error: unknown) {
          if (isStale()) return;

          const failure = historyFailureFrom(error);
          const message =
            error instanceof Error
              ? error.message
              : 'Failed to load submitted transactions';
          set({
            historyStatus: 'error',
            historyFailure: failure,
            historyError: message,
            historyAddress: address,
          });
        } finally {
          if (historyAbortController === controller) {
            historyAbortController = null;
          }
        }
      },

      refreshHistory: async () => {
        await get().loadHistoryPage(1);
      },

      nextHistoryPage: async () => {
        const { historyHasMore, historyPage, historyStatus } = get();
        if (!historyHasMore || historyStatus === 'loading') return;
        await get().loadHistoryPage(historyPage + 1);
      },

      prevHistoryPage: async () => {
        const { historyPage, historyStatus } = get();
        if (historyPage <= 1 || historyStatus === 'loading') return;
        await get().loadHistoryPage(historyPage - 1);
      },

      clearHistory: () => {
        // Bumping the generation invalidates every in-flight page load so a
        // stale response can never overwrite the cleared or reloaded state.
        historyAbortController?.abort();
        historyAbortController = null;
        set((state) => ({
          ...emptyHistory,
          historyGeneration: state.historyGeneration + 1,
        }));
      },
    }),
    {
      name: 'wallet-store',
      partialize: (state) => ({
        wallets: state.wallets,
      }),
    }
  )
);
