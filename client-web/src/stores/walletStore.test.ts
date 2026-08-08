import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { toBech32 } from '@cosmjs/encoding';
import type { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';
import type { SubmittedTransactionsPage } from '@/types/transaction';
import type { Wallet } from '@/types/wallet';
import { WalletService } from '@/services/wallet';
import { TransactionHistoryError } from '@/services/transaction';
import { useWalletStore } from './walletStore';

const originalLoadHistoryPage = useWalletStore.getState().loadHistoryPage;

const serviceMocks = vi.hoisted(() => ({
  getSubmittedTransactions: vi.fn(),
  send: vi.fn(),
}));

vi.mock('@/services/blockchain', () => ({
  BlockchainService: vi.fn(() => ({
    getBalance: vi.fn(async () => []),
  })),
}));

vi.mock('@/services/transaction', () => ({
  TransactionService: class {
    getSubmittedTransactions = serviceMocks.getSubmittedTransactions;
    send = serviceMocks.send;
  },
  TransactionHistoryError: class TransactionHistoryError extends Error {
    constructor(
      public readonly failure: string,
      message: string
    ) {
      super(message);
      this.name = 'TransactionHistoryError';
    }
  },
  HISTORY_DEFAULT_PAGE_SIZE: 20,
  HISTORY_MAX_PAGE_SIZE: 50,
}));

const ADDRESS = toBech32(
  'truerepublic',
  Uint8Array.from({ length: 20 }, (_, index) => index + 1)
);
const OTHER_ADDRESS = toBech32('truerepublic', new Uint8Array(20).fill(7));

const wallet: Wallet = {
  address: ADDRESS,
  name: 'Primary',
  createdAt: 1,
};

function pageResult(
  overrides: Partial<SubmittedTransactionsPage> = {}
): SubmittedTransactionsPage {
  return {
    address: ADDRESS,
    page: 1,
    pageSize: 20,
    total: 0,
    hasMore: false,
    transactions: [],
    ...overrides,
  };
}

function resetHistoryState() {
  useWalletStore.setState({
    currentWallet: { ...wallet },
    isLocked: false,
    password: 'pw',
    isLoading: false,
    error: null,
    historyTransactions: [],
    historyStatus: 'idle',
    historyFailure: null,
    historyError: null,
    historyPage: 1,
    historyPageSize: 20,
    historyTotal: 0,
    historyHasMore: false,
    historyAddress: null,
    historyGeneration: 0,
    loadHistoryPage: originalLoadHistoryPage,
  });
}

describe('wallet store submitted-transaction history', () => {
  beforeEach(() => {
    localStorage.clear();
    serviceMocks.getSubmittedTransactions.mockReset();
    serviceMocks.send.mockReset();
    resetHistoryState();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('loads a page into dedicated history state', async () => {
    serviceMocks.getSubmittedTransactions.mockResolvedValue(
      pageResult({
        page: 1,
        total: 45,
        hasMore: true,
        transactions: [
          {
            hash: 'AB'.repeat(32),
            height: 42,
            timestamp: '2026-08-08T12:00:00Z',
            code: 0,
            status: 'success',
            error: null,
            memo: null,
            messages: [{ typeUrl: '/cosmos.bank.v1beta1.MsgSend' }],
            fee: [{ denom: 'upnyx', amount: '12500' }],
          },
        ],
      })
    );

    await useWalletStore.getState().loadHistoryPage(1);

    const state = useWalletStore.getState();
    expect(serviceMocks.getSubmittedTransactions).toHaveBeenCalledWith(ADDRESS, 1);
    expect(state.historyStatus).toBe('ready');
    expect(state.historyFailure).toBeNull();
    expect(state.historyAddress).toBe(ADDRESS);
    expect(state.historyPage).toBe(1);
    expect(state.historyTotal).toBe(45);
    expect(state.historyHasMore).toBe(true);
    expect(state.historyTransactions).toHaveLength(1);
  });

  it('does not query when the wallet is locked', async () => {
    useWalletStore.setState({ isLocked: true, currentWallet: null });

    await useWalletStore.getState().loadHistoryPage(1);

    expect(serviceMocks.getSubmittedTransactions).not.toHaveBeenCalled();
  });

  it('keeps typed failures distinct from an empty result', async () => {
    useWalletStore.setState({
      historyTransactions: [
        {
          hash: 'AB'.repeat(32),
          height: 42,
          timestamp: null,
          code: 0,
          status: 'success',
          error: null,
          memo: null,
          messages: [],
          fee: null,
        },
      ],
      historyPage: 1,
      historyTotal: 1,
    });
    serviceMocks.getSubmittedTransactions.mockRejectedValue(
      new TransactionHistoryError('timeout', 'request timed out')
    );

    await useWalletStore.getState().loadHistoryPage(2);

    const state = useWalletStore.getState();
    expect(state.historyStatus).toBe('error');
    expect(state.historyFailure).toBe('timeout');
    expect(state.historyError).toContain('timed out');
    expect(state.historyTransactions).toEqual([]);
    expect(state.historyPage).toBe(2);
    expect(state.historyTotal).toBe(0);
  });

  it('maps unexpected errors to unavailable', async () => {
    serviceMocks.getSubmittedTransactions.mockRejectedValue(
      new Error('socket hangup')
    );

    await useWalletStore.getState().loadHistoryPage(1);

    expect(useWalletStore.getState().historyFailure).toBe('unavailable');
    expect(useWalletStore.getState().historyStatus).toBe('error');
  });

  it('ignores a stale response superseded by a newer page load', async () => {
    let resolveFirst: (value: SubmittedTransactionsPage) => void = () => {};
    serviceMocks.getSubmittedTransactions
      .mockImplementationOnce(
        () =>
          new Promise<SubmittedTransactionsPage>((resolve) => {
            resolveFirst = resolve;
          })
      )
      .mockResolvedValueOnce(pageResult({ page: 2, total: 45 }));

    const first = useWalletStore.getState().loadHistoryPage(1);
    await useWalletStore.getState().loadHistoryPage(2);
    resolveFirst(pageResult({ page: 1, total: 45 }));
    await first;

    const state = useWalletStore.getState();
    expect(state.historyPage).toBe(2);
    expect(state.historyStatus).toBe('ready');
  });

  it('ignores a response that arrives after lock', async () => {
    let resolveLoad: (value: SubmittedTransactionsPage) => void = () => {};
    serviceMocks.getSubmittedTransactions.mockImplementation(
      () =>
        new Promise<SubmittedTransactionsPage>((resolve) => {
          resolveLoad = resolve;
        })
    );

    const pending = useWalletStore.getState().loadHistoryPage(1);
    // Let the in-flight request reach the mocked service before locking.
    await new Promise((resolve) => setTimeout(resolve, 0));
    useWalletStore.getState().lock();
    resolveLoad(pageResult({ total: 3 }));
    await pending;

    const state = useWalletStore.getState();
    expect(state.historyStatus).toBe('idle');
    expect(state.historyTransactions).toEqual([]);
    expect(state.historyAddress).toBeNull();
  });

  it('clears history on lock', async () => {
    serviceMocks.getSubmittedTransactions.mockResolvedValue(
      pageResult({ total: 2 })
    );
    await useWalletStore.getState().loadHistoryPage(1);
    expect(useWalletStore.getState().historyStatus).toBe('ready');

    useWalletStore.getState().lock();

    const state = useWalletStore.getState();
    expect(state.historyStatus).toBe('idle');
    expect(state.historyTransactions).toEqual([]);
    expect(state.historyAddress).toBeNull();
    expect(state.historyTotal).toBe(0);
  });

  it('clears history when switching wallets', async () => {
    serviceMocks.getSubmittedTransactions.mockResolvedValue(
      pageResult({ total: 2 })
    );
    await useWalletStore.getState().loadHistoryPage(1);

    vi.spyOn(WalletService, 'getWallet').mockResolvedValue({
      address: OTHER_ADDRESS,
      name: 'Other',
      createdAt: 2,
    });

    await useWalletStore.getState().switchWallet(OTHER_ADDRESS, 'pw');

    const state = useWalletStore.getState();
    expect(state.currentWallet?.address).toBe(OTHER_ADDRESS);
    expect(state.historyStatus).toBe('idle');
    expect(state.historyTransactions).toEqual([]);
    expect(state.historyAddress).toBeNull();
  });

  it('clears previous-wallet history when creating a replacement wallet', async () => {
    useWalletStore.setState({
      historyStatus: 'ready',
      historyAddress: ADDRESS,
      historyTotal: 1,
    });
    vi.spyOn(WalletService, 'createWallet').mockResolvedValue({
      address: OTHER_ADDRESS,
      name: 'Created',
      createdAt: 2,
    });

    await useWalletStore.getState().createWallet('Created', 'pw');

    const state = useWalletStore.getState();
    expect(state.currentWallet?.address).toBe(OTHER_ADDRESS);
    expect(state.historyStatus).toBe('idle');
    expect(state.historyAddress).toBeNull();
    expect(state.historyTotal).toBe(0);
  });

  it('navigates pages only within bounds', async () => {
    serviceMocks.getSubmittedTransactions.mockImplementation(
      async (_address: string, page: number) =>
        pageResult({ page, total: 45, hasMore: page < 3 })
    );

    await useWalletStore.getState().loadHistoryPage(1);
    expect(useWalletStore.getState().historyPage).toBe(1);

    // Already on the first page: previous is a no-op.
    await useWalletStore.getState().prevHistoryPage();
    expect(serviceMocks.getSubmittedTransactions).toHaveBeenCalledTimes(1);

    await useWalletStore.getState().nextHistoryPage();
    expect(useWalletStore.getState().historyPage).toBe(2);

    await useWalletStore.getState().nextHistoryPage();
    expect(useWalletStore.getState().historyPage).toBe(3);

    // Last page: next is a no-op.
    await useWalletStore.getState().nextHistoryPage();
    expect(serviceMocks.getSubmittedTransactions).toHaveBeenCalledTimes(3);

    await useWalletStore.getState().prevHistoryPage();
    expect(useWalletStore.getState().historyPage).toBe(2);
  });

  it('refreshes page 1 after a successful send', async () => {
    vi.spyOn(WalletService, 'getWalletForSigning').mockResolvedValue(
      {} as DirectSecp256k1HdWallet
    );
    serviceMocks.send.mockResolvedValue({
      hash: 'EF'.repeat(32),
      height: 50,
      success: true,
    });
    serviceMocks.getSubmittedTransactions.mockResolvedValue(pageResult());

    const result = await useWalletStore.getState().sendTokens({
      to: OTHER_ADDRESS,
      amount: '1000',
      denom: 'upnyx',
    });

    expect(result.success).toBe(true);
    expect(serviceMocks.getSubmittedTransactions).toHaveBeenCalledWith(ADDRESS, 1);
    expect(useWalletStore.getState().historyStatus).toBe('ready');
  });

  it('does not report a committed send as failed when history refresh rejects', async () => {
    vi.spyOn(WalletService, 'getWalletForSigning').mockResolvedValue(
      {} as DirectSecp256k1HdWallet
    );
    serviceMocks.send.mockResolvedValue({
      hash: 'EF'.repeat(32),
      height: 50,
      success: true,
    });
    const rejectingRefresh = vi.fn(async () => {
      throw new Error('history chunk unavailable');
    });
    useWalletStore.setState({ loadHistoryPage: rejectingRefresh });

    const result = await useWalletStore.getState().sendTokens({
      to: OTHER_ADDRESS,
      amount: '1000',
      denom: 'upnyx',
    });

    expect(result.success).toBe(true);
    expect(rejectingRefresh).toHaveBeenCalledWith(1);
    expect(useWalletStore.getState().error).toBeNull();
  });

  it('creates no history row when a send fails before commit', async () => {
    vi.spyOn(WalletService, 'getWalletForSigning').mockResolvedValue(
      {} as DirectSecp256k1HdWallet
    );
    serviceMocks.send.mockResolvedValue({
      hash: '',
      height: 0,
      success: false,
      error: 'insufficient funds',
    });

    await expect(
      useWalletStore.getState().sendTokens({
        to: OTHER_ADDRESS,
        amount: '1000',
        denom: 'upnyx',
      })
    ).rejects.toThrow('insufficient funds');

    expect(serviceMocks.getSubmittedTransactions).not.toHaveBeenCalled();
    const state = useWalletStore.getState();
    expect(state.historyTransactions).toEqual([]);
    expect(state.historyStatus).toBe('idle');
  });
});
