import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { StargateClient } from '@cosmjs/stargate';
import { DEFAULT_CHAIN } from '@/config/chains';
import { WalletService } from '@/services/wallet';
import { BlockchainService } from '@/services/blockchain';
import { IbcChannelError, NetworkService } from '@/services/network';
import {
  NATIVE_DENOM,
  parseTransferAmount,
  reconcilePacketLifecycle,
  reducePacketLifecycle,
  sendIbcTransfer,
} from '@/services/ibcTransfer';
import type {
  IbcTransferParams,
  IbcTransferPhase,
  IbcTransferRecord,
  SendPacketEvidence,
  TransferChannel,
} from '@/types/ibc';
import { useWalletStore } from './walletStore';

export const IBC_TRANSFER_RESERVE_UPNYX = 10_000n;

type OperationStatus = 'idle' | 'loading' | 'ready' | 'error';

export type IbcTransferStoreFailure =
  | 'wallet_locked'
  | 'insufficient_balance'
  | 'stale_session'
  | 'record_not_found'
  | 'missing_packet_evidence'
  | 'channel_query_failed'
  | 'submission_failed'
  | 'reconciliation_failed';

export class IbcTransferStoreError extends Error {
  constructor(
    public readonly failure: IbcTransferStoreFailure,
    message: string,
    public readonly originalCause?: unknown
  ) {
    super(message);
    this.name = 'IbcTransferStoreError';
  }
}

interface IbcTransferStore {
  channels: TransferChannel[];
  channelStatus: OperationStatus;
  channelFailure: string | null;
  channelError: string | null;
  submissionStatus: OperationStatus;
  submissionPhase: IbcTransferPhase | null;
  submissionError: string | null;
  activeTxHash: string | null;
  reconcileStatus: OperationStatus;
  reconcileError: string | null;
  reconcileTxHash: string | null;
  recordsByScope: Record<string, IbcTransferRecord[]>;
  sessionGeneration: number;
  loadChannels: () => Promise<void>;
  submitTransfer: (params: IbcTransferParams) => Promise<IbcTransferRecord>;
  reconcileTransfer: (txHash: string) => Promise<IbcTransferPhase>;
  invalidateSession: () => void;
}

const networkService = new NetworkService(DEFAULT_CHAIN);
const blockchainService = new BlockchainService(DEFAULT_CHAIN);

export function ibcTransferScopeKey(chainId: string, address: string): string {
  return `${chainId}:${address}`;
}

function currentSession(): { address: string; password: string } {
  const { currentWallet, password, isLocked } = useWalletStore.getState();
  if (isLocked || !currentWallet || !password) {
    throw new IbcTransferStoreError('wallet_locked', 'Unlock a wallet first.');
  }
  return { address: currentWallet.address, password };
}

function message(error: unknown): string {
  return error instanceof Error ? error.message : 'Unknown IBC transfer error';
}

function persistedPacket(packet: SendPacketEvidence | null) {
  return packet === null ? null : { ...packet, sequence: packet.sequence.toString() };
}

function runtimePacket(record: IbcTransferRecord): SendPacketEvidence {
  if (record.packet === null || !/^[0-9]+$/.test(record.packet.sequence)) {
    throw new IbcTransferStoreError(
      'missing_packet_evidence',
      'This transfer has no valid send_packet evidence to reconcile.'
    );
  }
  return { ...record.packet, sequence: BigInt(record.packet.sequence) };
}

const emptySessionState = {
  channels: [] as TransferChannel[],
  channelStatus: 'idle' as OperationStatus,
  channelFailure: null,
  channelError: null,
  submissionStatus: 'idle' as OperationStatus,
  submissionPhase: null,
  submissionError: null,
  activeTxHash: null,
  reconcileStatus: 'idle' as OperationStatus,
  reconcileError: null,
  reconcileTxHash: null,
};

export const useIbcTransferStore = create<IbcTransferStore>()(
  persist(
    (set, get) => ({
      ...emptySessionState,
      recordsByScope: {},
      sessionGeneration: 0,

      loadChannels: async () => {
        const { address } = currentSession();
        const generation = get().sessionGeneration;
        set({ channelStatus: 'loading', channelFailure: null, channelError: null, channels: [] });
        try {
          const channels = await networkService.getTransferChannels();
          const wallet = useWalletStore.getState();
          if (
            get().sessionGeneration !== generation ||
            wallet.isLocked ||
            wallet.currentWallet?.address !== address
          ) return;
          set({ channels, channelStatus: 'ready' });
        } catch (error) {
          const wallet = useWalletStore.getState();
          if (
            get().sessionGeneration !== generation ||
            wallet.isLocked ||
            wallet.currentWallet?.address !== address
          ) return;
          set({
            channelStatus: 'error',
            channelFailure: error instanceof IbcChannelError ? error.failure : 'unknown',
            channelError: message(error),
          });
          throw new IbcTransferStoreError('channel_query_failed', message(error), error);
        }
      },

      submitTransfer: async (params) => {
        const { address, password } = currentSession();
        const generation = get().sessionGeneration;
        const amount = parseTransferAmount(params.amount, DEFAULT_CHAIN.coinDecimals);
        let freshBalances;
        try {
          freshBalances = await blockchainService.getBalance(address);
        } catch (error) {
          throw new IbcTransferStoreError(
            'submission_failed',
            'A fresh native balance could not be verified.',
            error
          );
        }
        const walletAfterBalance = useWalletStore.getState();
        if (
          get().sessionGeneration !== generation ||
          walletAfterBalance.isLocked ||
          walletAfterBalance.currentWallet?.address !== address
        ) {
          throw new IbcTransferStoreError(
            'stale_session',
            'Wallet session changed before signing.'
          );
        }
        const balanceText =
          freshBalances.find((balance) => balance.denom === NATIVE_DENOM)
            ?.amount ?? '0';
        if (!/^[0-9]+$/.test(balanceText)) {
          throw new IbcTransferStoreError('insufficient_balance', 'Native balance is invalid.');
        }
        if (BigInt(balanceText) < BigInt(amount) + IBC_TRANSFER_RESERVE_UPNYX) {
          throw new IbcTransferStoreError(
            'insufficient_balance',
            'Insufficient upnyx balance after the required fee reserve.'
          );
        }

        set({
          submissionStatus: 'loading',
          submissionPhase: 'validating',
          submissionError: null,
          activeTxHash: null,
        });
        try {
          const signer = await WalletService.getWalletForSigning(address, password);
          const submission = await sendIbcTransfer(DEFAULT_CHAIN, signer, params, {
            onPhase: (phase) => {
              const wallet = useWalletStore.getState();
              if (
                get().sessionGeneration === generation &&
                !wallet.isLocked &&
                wallet.currentWallet?.address === address
              ) set({ submissionPhase: phase });
            },
          });
          const record: IbcTransferRecord = {
            txHash: submission.txHash.toUpperCase(),
            chainId: DEFAULT_CHAIN.chainId,
            walletAddress: address,
            channel: params.channel,
            receiver: params.receiver,
            amount,
            memo: params.memo ?? '',
            height: submission.height,
            phase: submission.phase,
            packet: persistedPacket(submission.packet),
            submittedAt: Date.now(),
          };
          const scope = ibcTransferScopeKey(DEFAULT_CHAIN.chainId, address);
          set((state) => ({
            recordsByScope: {
              ...state.recordsByScope,
              [scope]: [
                record,
                ...(state.recordsByScope[scope] ?? []).filter(
                  (candidate) => candidate.txHash !== record.txHash
                ),
              ],
            },
          }));
          const wallet = useWalletStore.getState();
          if (
            get().sessionGeneration === generation &&
            !wallet.isLocked &&
            wallet.currentWallet?.address === address
          ) {
            set({
              submissionStatus: 'ready',
              submissionPhase: record.phase,
              activeTxHash: record.txHash,
            });
            await wallet.refreshBalance().catch(() => undefined);
          }
          return record;
        } catch (error) {
          const wallet = useWalletStore.getState();
          if (
            get().sessionGeneration === generation &&
            !wallet.isLocked &&
            wallet.currentWallet?.address === address
          ) set({ submissionStatus: 'error', submissionError: message(error) });
          throw error;
        }
      },

      reconcileTransfer: async (txHash) => {
        const { address } = currentSession();
        const generation = get().sessionGeneration;
        const normalizedHash = txHash.toUpperCase();
        const scope = ibcTransferScopeKey(DEFAULT_CHAIN.chainId, address);
        const record = (get().recordsByScope[scope] ?? []).find(
          (candidate) => candidate.txHash === normalizedHash
        );
        if (!record) {
          throw new IbcTransferStoreError('record_not_found', 'Transfer record not found.');
        }
        const packet = runtimePacket(record);
        set({ reconcileStatus: 'loading', reconcileError: null, reconcileTxHash: normalizedHash });
        let client: StargateClient | undefined;
        try {
          client = await StargateClient.connect(DEFAULT_CHAIN.rpc);
          const evidence = await reconcilePacketLifecycle(client, packet);
          const phase = reducePacketLifecycle(evidence);
          const wallet = useWalletStore.getState();
          if (
            get().sessionGeneration !== generation ||
            wallet.isLocked ||
            wallet.currentWallet?.address !== address
          ) {
            throw new IbcTransferStoreError('stale_session', 'Wallet session changed during recovery.');
          }
          set((state) => ({
            recordsByScope: {
              ...state.recordsByScope,
              [scope]: (state.recordsByScope[scope] ?? []).map((candidate) =>
                candidate.txHash === normalizedHash ? { ...candidate, phase } : candidate
              ),
            },
            reconcileStatus: 'ready',
          }));
          return phase;
        } catch (error) {
          if (!(error instanceof IbcTransferStoreError && error.failure === 'stale_session')) {
            set({ reconcileStatus: 'error', reconcileError: message(error) });
          }
          throw error;
        } finally {
          client?.disconnect();
        }
      },

      invalidateSession: () =>
        set((state) => ({ ...emptySessionState, sessionGeneration: state.sessionGeneration + 1 })),
    }),
    {
      name: 'ibc-transfer-store',
      partialize: (state) => ({ recordsByScope: state.recordsByScope }),
    }
  )
);

export function invalidateIbcTransferSession(): void {
  useIbcTransferStore.getState().invalidateSession();
}

export function recordsForWallet(address: string): IbcTransferRecord[] {
  return useIbcTransferStore.getState().recordsByScope[
    ibcTransferScopeKey(DEFAULT_CHAIN.chainId, address)
  ] ?? [];
}
