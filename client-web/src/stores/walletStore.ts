import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { Wallet, Balance } from '@/types/wallet';
import type { Transaction, SendParams, TransactionResult } from '@/types/transaction';
import { WalletService } from '@/services/wallet';
import { BlockchainService } from '@/services/blockchain';
import { TransactionService } from '@/services/transaction';
import { DEFAULT_CHAIN } from '@/config/chains';

interface WalletStore {
  // State
  currentWallet: Wallet | null;
  wallets: Wallet[];
  balances: Balance[];
  transactions: Transaction[];
  isLocked: boolean;
  password: string | null;
  isLoading: boolean;
  error: string | null;

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
  loadTransactions: () => Promise<void>;
}

const blockchainService = new BlockchainService(DEFAULT_CHAIN);
const transactionService = new TransactionService(DEFAULT_CHAIN);

export const useWalletStore = create<WalletStore>()(
  persist(
    (set, get) => ({
      // State
      currentWallet: null,
      wallets: [],
      balances: [],
      transactions: [],
      isLocked: true,
      password: null,
      isLoading: false,
      error: null,

      // Actions
      createWallet: async (name: string, password: string) => {
        set({ isLoading: true, error: null });

        try {
          const wallet = await WalletService.createWallet({ name, password });

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
        set((state) => ({
          wallets: state.wallets.filter((w) => w.address !== address),
          currentWallet:
            state.currentWallet?.address === address ? null : state.currentWallet,
        }));
      },

      lock: () => {
        set({
          isLocked: true,
          password: null,
          currentWallet: null,
          balances: [],
          transactions: [],
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
          const balances = await blockchainService.getBalance(currentWallet.address);
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

          const result = await transactionService.send(signingWallet, params);

          if (!result.success) {
            throw new Error(result.error || 'Transaction failed');
          }

          await get().refreshBalance();

          set({ isLoading: false });
          return result;
        } catch (error: unknown) {
          const message = error instanceof Error ? error.message : 'Transaction failed';
          set({ error: message, isLoading: false });
          throw error;
        }
      },

      loadTransactions: async () => {
        // Transaction history indexing is a v0.4.0 Phase 1 feature
        set({ transactions: [] });
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
